package fep

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 回放遥测：`<录制 ID>-replay-<时刻>.jsonl`，一行一步。
//
// 为什么要有它：回放**只在跳过/失败时写日志**，成功是静默的 —— 所以"回放到第几步了"
// 这个问题在当前实现里根本无从回答（用户的原话是"播放过账中我可以观察走到哪了"）。
// 回放是**另一个进程**干的（面板按钮走 SpawnDetached，CLI 也会起），所以进度只能落成
// 一个两边都读得到的东西：文件。落一次，三处都受益：
//
//	界面      轮询它 / 被 WS 叫一声之后读它（`/api/recordings/{id}/replay-steps`）
//	命令行    卡住时看最后一行就知道停在哪一步
//	事后复盘  回放是**一次性**的，日志是唯一还能回看的东西
//
// 三条纪律：
//
//  1. **写一行 flush 一行**：它要被**边跑边读**。攒在 bufio 里等 64 KB 的话，
//     60 步的回放直到结束才落盘，"实时进度"就成了假的。
//  2. **`sentAt`/`respAt` 用 null 表示"不知道"**，不用 0：0 会被读成"1970 年"或
//     "瞬回"，而"没有回包"和"0 毫秒回包"是两件完全不同的事（见 replay.go 的注释）。
//  3. **每一步可能有不止一行**（`waiting` 之后才是最终结果；暂停时还有 `paused`）。
//     所以读侧按 `n` 取**最后一行**（`ReplayLog.States()`），而不是"一行一步"。
//     这样"更细的中间状态"以后还能加，不必改格式版本。

// ReplayLogSuffix 是回放日志的扩展名（与帧索引同住缓存目录）。
const ReplayLogSuffix = ".replay.jsonl"

// 回放日志的格式版本（与帧索引同理：读侧靠它决定要不要重建/拒绝解析）。
const replayLogVersion = 1

// 一步的结果。**前四个是终态**，后两个是**中间态**（同一 n 之后还会有终态行）：
//
//	waiting     已发出、正在等运行期回包（界面画"进行中"，**不画时长**）
//	paused      用户按了暂停（或单步），这一步还没发出去
//	sent        发出去了，而且运行期回了包
//	no-response 发出去了，运行期在超时内没回包（**不等于 0 毫秒**）
//	skipped     没发（AUI 安全闸说对象不在当前界面 / dry-run / 用户 --skip）
//	refused     没发：这一步带字段身份，但身份**解析不过**（四码之一）——
//	            与 skipped 分开记：一个是"该跳"，一个是"找错了，别再往下跑"
//	failed      发的时候出错
const (
	ReplayWaiting    = "waiting"
	ReplayPaused     = "paused"
	ReplaySent       = "sent"
	ReplayNoResponse = "no-response"
	ReplaySkipped    = "skipped"
	ReplayRefused    = "refused"
	ReplayFailed     = "failed"
)

// ReplayStep 是回放日志里的一行（也逐字给到界面）。
type ReplayStep struct {
	N      int    `json:"n"`
	Total  int    `json:"total"`
	Event  string `json:"event"`
	Kind   string `json:"kind,omitempty"`   // ActionEvent | KeyEvent | ConfigureEvent
	Target int    `json:"target,omitempty"` // 载荷里的 idRef（取不到就没有）
	// SentAt 是**发出去的那一刻**（Unix 毫秒）。跳过/失败/暂停时是 null。
	SentAt *int64 `json:"sentAt"`
	// RespAt 是**运行期回包的那一刻**。null = 没有回包（不是 0 毫秒）。
	RespAt *int64 `json:"respAt"`
	Result string `json:"result"`
	// Suspect 是"这一步**本该**有回包却没有"——点击类事件（ActionEvent）通常一定会有反应，
	// 而光标/键盘那类本来就没应答。界面据此把"本来无应答"与"8s 无回包（可疑）"分开：
	// 两者的下一步完全不同（一个是正常，一个要去看运行期/作业状态）。
	Suspect bool `json:"suspect,omitempty"`
	// RuntimeBytes 是这一步之后运行期推回来的字节数（"被确认"的量度）。
	RuntimeBytes int64  `json:"runtimeBytes,omitempty"`
	Reason       string `json:"reason,omitempty"`
	// ── 身份（契约 C3 + A18）────────────────────────────────────────────────
	//
	// 录制时写入（从当步目标节点取），两个键都 `omitempty`：旧日志没有它们 ⇒ 读侧得
	// 零值 = 未记录，**不是错误**（`ReadReplayLog` 不做任何键集校验）。
	ColName string `json:"colName,omitempty"`
	// Resolved 说明这一步**最终怎么找到目标**的：`field`（按字段身份现解析）|
	// `bare`（按录制时的临时编号 —— 降级，跨会话可能点错）。空 = 没走到那一步。
	Resolved string `json:"resolved,omitempty"`
	// Degraded 是 Resolved=="bare" 的布尔形态：界面/汇总按它聚合，不必再解字符串。
	Degraded bool `json:"degraded,omitempty"`
	FieldID int    `json:"fieldId,omitempty"`
	// Code 是这一步被拒时的四码（`FIELD_NOT_IN_SCOPE` 等）；只有 `refused` 会带。
	Code string `json:"code,omitempty"`
}

// replayLine 是文件里一行的完整形状：步骤 + 头行/尾行的元数据（与帧索引同一套做法）。
type replayLine struct {
	ReplayStep
	V        int    `json:"v,omitempty"`
	ID       string `json:"id,omitempty"`
	Conn     int    `json:"conn,omitempty"`
	Started  int64  `json:"started,omitempty"`
	Finished int64  `json:"finished,omitempty"` // >0 = 尾行（这一轮结束了）
	Sent     int    `json:"sent,omitempty"`
	Skipped  int    `json:"skipped,omitempty"`
	Refused  int    `json:"refused,omitempty"`
	Failed   int    `json:"failed,omitempty"`
	Note     string `json:"note,omitempty"`
}

// ReplayLogWriter 逐行写回放日志。
type ReplayLogWriter struct {
	f     *os.File
	path  string
	total int // 计划步数（尾行报的是它，不是"落了多少行"——一行不等于一步）
	n     int // 落了多少行（不含头尾）
	err   error
}

// NewReplayLogWriter 创建日志并写头行（ID / 连接 / 计划步数）。
func NewReplayLogWriter(path, id string, conn, total int) (*ReplayLogWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := &ReplayLogWriter{f: f, path: path, total: total}
	head, _ := json.Marshal(replayLine{
		V: replayLogVersion, ID: id, Conn: conn, Started: time.Now().UnixMilli(),
		ReplayStep: ReplayStep{Total: total},
	})
	if _, err := f.Write(append(head, '\n')); err != nil {
		f.Close()
		return nil, err
	}
	return w, nil
}

// Path 返回日志落点（首页就要用它：界面按它轮询）。
func (w *ReplayLogWriter) Path() string {
	if w == nil {
		return ""
	}
	return w.path
}

// Step 落一步并**立刻 flush**（见文件头纪律 1）。
func (w *ReplayLogWriter) Step(s ReplayStep) {
	if w == nil || w.err != nil {
		return
	}
	b, err := json.Marshal(replayLine{ReplayStep: s})
	if err != nil {
		w.err = err
		return
	}
	if _, err := w.f.Write(append(b, '\n')); err != nil {
		w.err = err
		return
	}
	w.n++
}

// Abandon 关掉文件句柄但**不写尾行**。
//
// 两种真实情形：回放进程被杀（OS 替它关了句柄）、或者被中断时不想把半份进度标成"完成"
// （尾行的意思是"这一轮结束了"，而它并没有结束）。测试要复现"没有尾行"那个形状也走它。
func (w *ReplayLogWriter) Abandon() error {
	if w == nil {
		return nil
	}
	return w.f.Close()
}

// Close 写尾行（结果汇总）。尾行在 = "这一轮回放结束了"，界面据此把进度条收成"完成"。
func (w *ReplayLogWriter) Close(res ReplayResult, note string) error {
	if w == nil {
		return nil
	}
	if w.err == nil {
		tail, _ := json.Marshal(replayLine{
			Finished: time.Now().UnixMilli(), Sent: res.Sent, Skipped: res.Skipped,
			Refused: res.Refused, Failed: res.Failed, Note: note,
			ReplayStep: ReplayStep{Total: w.total, RuntimeBytes: res.RuntimeBytes},
		})
		if _, err := w.f.Write(append(tail, '\n')); err != nil {
			w.err = err
		}
	}
	cerr := w.f.Close()
	if w.err != nil {
		return w.err
	}
	return cerr
}

// ErrNoReplayLog 表示这一份录制还没有回放日志（没回过放 —— 那是正常状态，不是错误）。
var ErrNoReplayLog = errors.New("fep: 这一份录制还没有回放日志")

// ReplayLog 是一份读进来的回放日志。
type ReplayLog struct {
	Path     string
	ID       string
	Conn     int
	Total    int
	Started  int64
	Finished int64
	Done     bool // 尾行在（这一轮结束了）
	Steps    []ReplayStep
	Summary  ReplayResult
	Note     string
	Bad      int // 读不动的行数（>0 = 日志有损，界面照实标）
}

// ReadReplayLog 读一份回放日志。坏行只计数，不报错（它是排障文件，不该让页面打不开）。
func ReadReplayLog(path string) (*ReplayLog, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoReplayLog
		}
		return nil, err
	}
	defer f.Close()

	out := &ReplayLog{Path: path, ID: strings.TrimSuffix(filepath.Base(path), ReplayLogSuffix)}
	// 一行可能很长（事件文本 + 原因），给到 1 MB。
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var ln replayLine
		if err := json.Unmarshal(line, &ln); err != nil {
			out.Bad++
			continue
		}
		switch {
		case ln.V > 0:
			out.ID, out.Conn, out.Total, out.Started = ln.ID, ln.Conn, ln.Total, ln.Started
		case ln.Finished > 0:
			out.Finished, out.Done, out.Note = ln.Finished, true, ln.Note
			out.Summary = ReplayResult{Sent: ln.Sent, Skipped: ln.Skipped, Refused: ln.Refused, Failed: ln.Failed, RuntimeBytes: ln.RuntimeBytes}
		default:
			out.Steps = append(out.Steps, ln.ReplayStep)
		}
	}
	if out.Total == 0 {
		out.Total = len(out.Steps)
	}
	// 头行丢了（被杀在半路）也不影响：ID 从文件名切，步数按实际行数。
	if out.ID == "" {
		out.ID = strings.TrimSuffix(filepath.Base(path), ReplayLogSuffix)
	}
	return out, sc.Err()
}

// From 返回第 since 条之后的步骤（`since` = **已经拿到多少条**，不是 n ——
// 跳过与暂停会产生额外的行，用 n 当游标会丢步）。
func (l *ReplayLog) From(since int) []ReplayStep {
	if l == nil || since >= len(l.Steps) {
		return nil
	}
	if since < 0 {
		since = 0
	}
	return l.Steps[since:]
}

// States 按 n 收敛出**每一步的最新状态**（同一 n 的多行取最后一行）。
//
// 为什么需要它：中间态（waiting / paused）之后还有终态行，界面要画的是"这一步现在怎么样"，
// 而不是"这个 n 出现过几次"。
func (l *ReplayLog) States() []ReplayStep {
	if l == nil {
		return nil
	}
	idx := map[int]int{}
	var out []ReplayStep
	for _, s := range l.Steps {
		if i, ok := idx[s.N]; ok {
			out[i] = s
			continue
		}
		idx[s.N] = len(out)
		out = append(out, s)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].N < out[j].N })
	return out
}

// ReplayProgress 是给界面的一行汇总（进度条与计数徽标都照它渲染）。
type ReplayProgress struct {
	Total   int `json:"total"`
	Done    int `json:"done"`    // 已经有终态的步数
	Sent    int `json:"sent"`    // 发出且运行期回了包
	NoResp  int `json:"noResp"`  // 发出但没回包（其中 suspect 的是"可疑"）
	Suspect int `json:"suspect"` // 上面那些里"本该有回包却没有"的
	Skipped int `json:"skipped"`
	// Refused 是"带身份、解析不过 ⇒ 一个字节都没发"的步数（四码之一）。
	Refused int  `json:"refused,omitempty"`
	// Degraded 是"没有正式身份、按录制时的临时编号发出去"的步数（降级发送）。
	Degraded int  `json:"degraded,omitempty"`
	Failed   int  `json:"failed"`
	Pending bool `json:"pending"` // 有一步正在等回包（waiting）
	Paused  bool `json:"paused"`  // 停在第 N 步（等用户按继续/单步）
	Running bool `json:"running"` // 尾行还没落 = 这一轮还在跑
}

// Progress 汇总当前进度。
func (l *ReplayLog) Progress() ReplayProgress {
	p := ReplayProgress{Running: l != nil && !l.Done}
	if l == nil {
		return p
	}
	p.Total = l.Total
	if p.Total == 0 {
		p.Total = len(l.Steps)
	}
	for _, s := range l.States() {
		if s.Degraded {
			p.Degraded++
		}
		switch s.Result {
		case ReplayWaiting:
			p.Pending = true
		case ReplayPaused:
			p.Paused = true
		case ReplaySent:
			p.Done++
			p.Sent++
		case ReplayNoResponse:
			p.Done++
			p.NoResp++
			if s.Suspect {
				p.Suspect++
			}
		case ReplaySkipped:
			p.Done++
			p.Skipped++
		case ReplayRefused:
			p.Done++
			p.Refused++
		case ReplayFailed:
			p.Done++
			p.Failed++
		}
	}
	return p
}

// ReplayLogName 是回放日志的文件名（与录制件同一个前缀，便于按 ID 认领）。
//
// 落**缓存目录**而不是长期库：回放是**事件**不是录制件。要留证据就 `tt recording keep` 那一份
// 录制 —— 别把"事件的记录"混进"证据的库"（见计划 C 的决策 ⑤）。
func ReplayLogName(id string, at time.Time) string {
	return fmt.Sprintf("%s-%s%s", id, at.Format("20060102-150405"), ReplayLogSuffix)
}

// FindReplayLogs 列出某一份录制的回放日志（**新的在前**）。
func FindReplayLogs(dir, id string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	prefix := id + "-"
	var out []string
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasSuffix(n, ReplayLogSuffix) || !strings.HasPrefix(n, prefix) {
			continue
		}
		// 前缀能撞上（`fep-1` 与 `fep-10`）：再核一次切出来的 ID。
		if got := strings.TrimSuffix(n, ReplayLogSuffix); !strings.HasPrefix(got, prefix) {
			continue
		}
		out = append(out, filepath.Join(dir, n))
	}
	// 文件名里带 `20060102-150405`，字典序就是时间序；倒过来 = 新的在前。
	sort.Sort(sort.Reverse(sort.StringSlice(out)))
	return out, nil
}

// LatestReplayLog 取最新的一份（没有就 ErrNoReplayLog）。
func LatestReplayLog(dir, id string) (*ReplayLog, error) {
	files, err := FindReplayLogs(dir, id)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, ErrNoReplayLog
	}
	return ReadReplayLog(files[0])
}

// ReplayCtl 是"人对回放下的命令"落在文件里的那一个词。
//
// 为什么用一个文件而不是新开一个控制端口：回放跑在**另一个进程**里，而我们已经有
// "缓存目录 + 录制 ID"这个双方都知道的坐标。开端口要处理占用、鉴权与防火墙；
// 一个一两百毫秒轮询一次的小文件，语义清楚、进程死了也不留监听。
type ReplayCtl string

const (
	CtlPause  ReplayCtl = "pause"
	CtlResume ReplayCtl = "resume"
	CtlStep   ReplayCtl = "step" // 单步：继续一步，然后停下
	CtlStop   ReplayCtl = "stop"
	// CtlSkip：**跳过当前这一步**（不发），并**仍然停在暂停态**等下一步。
	// 与 CtlStep 的区别：step 是"发这一步然后停"，skip 是"不发这一步、仍停着"。
	CtlSkip ReplayCtl = "skip"
)

// ReplayCtlPath 是控制文件的落点（与日志同一个干，扩展名不同）。
func ReplayCtlPath(logPath string) string {
	return strings.TrimSuffix(logPath, ReplayLogSuffix) + ".ctl"
}

// WriteReplayCtl 写控制命令（覆盖式：只有最新一条有意义）。
func WriteReplayCtl(logPath string, c ReplayCtl) error {
	return os.WriteFile(ReplayCtlPath(logPath), []byte(c+"\n"), 0o644)
}

// ReplayBreaksPath 是"这一步之前停住"的断点集合（与 .ctl 同目录、同前缀）。
//
// 为什么单独一个文件而不是塞进 .ctl：.ctl 一次只有**一个**当前命令（覆盖式），
// 而断点是**一组**步号，要能随时增删、跨进程共享。
func ReplayBreaksPath(logPath string) string {
	return strings.TrimSuffix(logPath, ReplayLogSuffix) + ".breaks"
}

// ReadReplayBreaks 读断点集合（步号）。文件不在 = 没有断点（不是错误）。
func ReadReplayBreaks(logPath string) []int {
	b, err := os.ReadFile(ReplayBreaksPath(logPath))
	if err != nil {
		return nil
	}
	var out []int
	for _, ln := range strings.Split(string(b), "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		if n, cerr := strconv.Atoi(ln); cerr == nil && n > 0 {
			out = append(out, n)
		}
	}
	return out
}

// WriteReplayBreaks 覆盖写断点集合（排序去重，一行一个）。
func WriteReplayBreaks(logPath string, ns []int) error {
	seen := map[int]bool{}
	var clean []int
	for _, n := range ns {
		if n <= 0 || seen[n] {
			continue
		}
		seen[n] = true
		clean = append(clean, n)
	}
	sort.Ints(clean)
	var b strings.Builder
	for _, n := range clean {
		b.WriteString(strconv.Itoa(n))
		b.WriteString("\n")
	}
	return os.WriteFile(ReplayBreaksPath(logPath), []byte(b.String()), 0o644)
}

// HasReplayBreak 判某一步是否设了断点（每次现读：人可能刚点上）。
func HasReplayBreak(logPath string, n int) bool {
	for _, x := range ReadReplayBreaks(logPath) {
		if x == n {
			return true
		}
	}
	return false
}

// ReadReplayCtl 读控制命令（文件不在或读不动时返回空 = 什么都不做）。
//
// **一次性命令读完就删**（step / skip）：这俩是"动作"，不是"状态"。
// 不删的后果实测（2026-10-18）：回放每步都会再读到文件里那条 `skip`，
// 于是"跳过这一步"变成"从这里开始全跳过"；`step` 同理会把逐步变成一路跑到底。
// pause / resume / stop 是状态，留着（下一次读还是它，这正是它们该有的语义）。
func ReadReplayCtl(logPath string) ReplayCtl {
	p := ReplayCtlPath(logPath)
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	c := ReplayCtl(strings.TrimSpace(string(b)))
	if c == CtlStep || c == CtlSkip {
		_ = os.Remove(p)
	}
	return c
}
