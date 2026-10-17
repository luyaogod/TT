package fep

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 录制件的身份（R0.2）。
//
// 落成 `<录制ID>.meta.json`，与录制件同目录、同前缀。为什么必须落盘而不是只记在心里：
// 录制件是**跨进程、跨时间**用的东西 —— 用户在前端面板录一轮，几小时后 AI 拿 ID 去
// 回放/调试，中间没有任何一个进程活着，所以身份只能跟着文件走。
//
// 为什么身份里要有"启动路径"和`.42r` 标识：指纹能发现"界面结构变了"，但发现不了
// "程序只是重编译了一次、界面一模一样"（那种情况下编号表逐条相同）。两者合起来才是
// 完整的"这份录制还算不算数"。

// MetaSuffix 是身份文件的扩展名（与录制件的 `<ID>-<连接>-rt2gdc.log` 同一前缀）。
const MetaSuffix = ".meta.json"

// ErrNoMeta 报告这个录制 ID 没有身份文件。
//
// 这不是"读失败"：早期录的、或者进程被硬杀时录的件就是没有身份 —— 调用方要能
// 把"身份不详"与"读不到文件"分开，前者照样能算指纹（指纹是从录制件的字节算出来的）。
var ErrNoMeta = errors.New("fep: 这个录制件没有身份文件（.meta.json）")

// MetaConn 是一条被录到的连接。
//
// 三个文件字段存的是**文件名**（不是绝对路径）：身份文件就住在录制件旁边，
// 而一份录制会被整体搬进长期库、也可能被整个目录拷给别人 —— 存绝对路径的话，
// 搬完之后身份里记的就是三个不存在的路径。
type MetaConn struct {
	ID       int    `json:"id"`
	Remote   string `json:"remote,omitempty"`
	RT2GDC   string `json:"rt2gdc,omitempty"` // 运行期 → 前端（AUI；界面状态就在这一份里）
	GDC2RT   string `json:"gdc2rt,omitempty"` // 前端 → 运行期（用户/注入的事件）
	Timeline string `json:"timeline,omitempty"`
	BytesRT  int64  `json:"bytesRT,omitempty"`
	BytesFE  int64  `json:"bytesFE,omitempty"`
}

// Path 把文件名还原成绝对路径（dir = 这份录制**现在**所在的目录）。
//
// 一律只取文件名再拼当前目录：身份里可能存着搬到长期库**之前**的那个绝对路径
// （早先的写法就是存绝对路径），照抄它会指到一个已经不存在的文件。
func (c MetaConn) Path(dir, name string) string {
	if name == "" {
		return ""
	}
	return filepath.Join(dir, filepath.Base(name))
}

// BinaryID 是作业编译件的标识（`.42r` 的大小 + 修改时间）。
//
// 它是**只读探测**来的，探测失败就留空 —— 不是所有启动路径都能知道这个路径：
// 走 `r.r` 时程序由 T100 启动器按登录 shell 的 `$AZZ` 决定，我们只给了作业名。
// 留空时身份文件里就照实空着，闸门也只说"编译件无从比对"，绝不猜。
type BinaryID struct {
	Ref     string `json:"ref,omitempty"`     // 服务器上的路径
	Size    int64  `json:"size,omitempty"`    // 字节数
	ModTime string `json:"modTime,omitempty"` // RFC3339
}

// Empty 报告这一项有没有内容。
func (b BinaryID) Empty() bool { return b.Ref == "" && b.Size == 0 && b.ModTime == "" }

// FingerprintString 给人读的形态。
func (b BinaryID) FingerprintString() string {
	if b.Empty() {
		return "（没探到）"
	}
	parts := []string{b.Ref}
	if b.Size > 0 {
		parts = append(parts, fmt.Sprintf("%d 字节", b.Size))
	}
	if b.ModTime != "" {
		parts = append(parts, b.ModTime)
	}
	return strings.Join(parts, " / ")
}

// RecordingMeta 是一份录制件的身份。
type RecordingMeta struct {
	ID   string     `json:"id"`
	Path LaunchPath `json:"path"`
	Prog string     `json:"prog,omitempty"`
	// Origin / Retention 是"这份录制是谁录的、留多久"（策略见 retention.go）。
	// 它们也解释"为什么这一份会在缓存目录里被自动清掉"。
	Origin    Origin    `json:"origin,omitempty"`
	Retention Retention `json:"retention,omitempty"`
	// 生效身份：作业落在哪个环境/区域/企业。ENT 空 = 沿用登录 shell 自己的 $TOPENT。
	Env  string `json:"env,omitempty"`
	Zone string `json:"zone,omitempty"`
	Ent  string `json:"ent,omitempty"`
	Host string `json:"host,omitempty"` // SSH host（不含账号与端口）
	// FGLServer 是这一次画面接线的地址（`<本机IP>:<代理端口-6400>`）。
	FGLServer string `json:"fglserver,omitempty"`

	Started  time.Time `json:"started"`
	Finished time.Time `json:"finished,omitempty"`

	Binary BinaryID `json:"binary,omitempty"`

	Conns []MetaConn `json:"conns,omitempty"`

	// Fingerprint 是收工那一刻算出来的界面指纹（编号表摘要）。
	// Total=0 表示这一轮根本没录到界面（作业没连上来 / 刚起来就断了）。
	Fingerprint Fingerprint `json:"fingerprint"`

	Note string `json:"note,omitempty"` // 一句话说明（例如"没有身份文件时补写的"）

	// ── A24（契约 §7 / 用户决策 D-F2）：旧件处置 = **只打标记，不加逃生门** ──
	//
	// 这三个键只回答一件事："这份录制**还算不算数**"。旧件（无身份 / v2 指纹）反正回放
	// 会被拒，而录制是**不可再生**的（再报一次 bug、再操作一遍才有），所以既不许删、
	// 也不许悄悄留着让人以为能用 —— 打上标记，列表与明细把它写在脸上。
	//
	// 三个键**全部 omitempty**：老身份文件里没有它们 ⇒ 读进来就是 false/零值，
	// 不需要迁移、不许因此报错（TestOldMetaWithoutStaleFieldsReadsAsLive）。
	Stale       bool      `json:"stale,omitempty"`   // true = 不可回放（需重录）
	StaleReason string    `json:"reason,omitempty"`  // 为什么标它（人写的文字，可以空）
	StaleAt     time.Time `json:"staleAt,omitempty"` // 标记的时刻
	// 注：`staleAt` 的 `omitempty` 只对**读**有意义 —— `encoding/json` 不认识 time.Time
	// 的零值，所以没标记的件写出去仍会带一个 `"staleAt":"0001-01-01T00:00:00Z"`
	// （既有 `finished` 一直是这个行为，不为此引入指针/自定义 Marshaler）。
	// **判据永远只看 `stale`**；读取方不许拿 staleAt 是否为零来判断"标记过没有"。
}

// StaleNote 是"这份录制不可回放"的**唯一文案**：列表、明细、命令回执都引它，
// 免得同一件事在三处写成三种说法。
const StaleNote = "不可回放（需重录）"

// IsStale 报告这份录制是不是被标记为不可回放。nil 安全（没有身份文件 = 没标记）。
//
// 为什么用方法而不是让调用方读字段：判据将来可能变（例如"指纹是 v2 也算不可回放"），
// 变化时要改的地方只有这一处。
func (m *RecordingMeta) IsStale() bool { return m != nil && m.Stale }

// MarkStale 给一份录制**打标记**：`stale=true` + 原因 + 时间。
//
// 它做的**只有**重写身份文件（`<ID>.meta.json`）：
//
//   - **绝不删任何文件**（录制不可再生，删了就造不回来）；
//   - **绝不改录制内容**（`-rt2gdc.log` / `-gdc2rt.log` / `-timeline.log` 一个字节不动）；
//   - 其它身份字段原样保留（这不是"重写身份"，只是加三件事）。
//
// 没有身份文件的件（收工被硬杀、极老的件）也能标记：补写一份**只有 ID 与标记**的
// 身份文件，并在 `Note` 里照实说这是补写的 —— 那种件恰恰最需要"别拿它回放"这句话。
//
// 空原因不写空键（omitempty）；时间由调用方给，便于测试与"按时间排"的展示。
func MarkStale(dir, id, reason string, at time.Time) (*RecordingMeta, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("fep: 标记要录制 ID（录制 ID 就是录制件的前缀）")
	}
	m, err := ReadMeta(dir, id)
	if err != nil {
		if !errors.Is(err, ErrNoMeta) {
			return nil, err
		}
		m = &RecordingMeta{ID: id, Note: "标记时补写的（这份录制原本没有身份文件）"}
	}
	m.ID = id
	m.Stale = true
	m.StaleReason = strings.TrimSpace(reason)
	m.StaleAt = at
	if _, werr := WriteMeta(dir, *m); werr != nil {
		return nil, werr
	}
	return m, nil
}

// Duration 报告这一轮录了多久（没结束时返回 0）。
func (m RecordingMeta) Duration() time.Duration {
	if m.Finished.IsZero() || m.Started.IsZero() {
		return 0
	}
	return m.Finished.Sub(m.Started)
}

// MetaPath 返回某个录制 ID 的身份文件落点。
func MetaPath(dir, id string) string { return filepath.Join(dir, id+MetaSuffix) }

// WriteMeta 落一份身份文件（临时文件 + 改名：写一半的身份比没有更坏，
// 因为它会被当成"这份录制的路径/指纹就是这些"）。
func WriteMeta(dir string, m RecordingMeta) (string, error) {
	if m.ID == "" {
		return "", errors.New("fep: 录制身份缺 ID（录制 ID 就是录制件的前缀）")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return "", err
	}
	path := MetaPath(dir, m.ID)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", err
	}
	return path, nil
}

// ReadMeta 读一份身份文件；没有它返回 ErrNoMeta。
func ReadMeta(dir, id string) (*RecordingMeta, error) {
	b, err := os.ReadFile(MetaPath(dir, id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoMeta
		}
		return nil, err
	}
	var m RecordingMeta
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("fep: 身份文件 %s: %w", MetaPath(dir, id), err)
	}
	if m.ID == "" {
		m.ID = id
	}
	return &m, nil
}

// ListIDs 列出目录里所有录制件的 ID（有身份文件的用身份文件的 ID，没有的从文件名推）。
//
// 一趟只认 **rt2gdc**（运行期→前端）那一侧：它才是界面状态的来源，也是指纹的来源；
// 只有 gdc2rt 的连接（前端发出去、界面一条都没回来）不算一份可用的录制。
func ListIDs(dir string) ([]string, error) {
	files, err := RecordingFiles(dir, "-rt2gdc.log")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var ids []string
	for _, f := range files {
		id, ok := RecordingIDFromFile(filepath.Base(f))
		if !ok || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

// RecordingIDFromFile 从 `<ID>-<连接号>-rt2gdc.log` 里切出录制 ID。
//
// 录制 ID 就是代理落盘用的 tag（`session open` 里形如 fep-20261015-170031）。
// 切法是"去掉末尾的 -<数字>-rt2gdc.log"：连接号一定是数字，而 ID 里允许有连字符，
// 所以不能按第一个连字符切。没有连接号那一段（单文件、夹具）就整段当 ID。
func RecordingIDFromFile(name string) (string, bool) {
	if !strings.HasSuffix(name, "-rt2gdc.log") {
		return "", false
	}
	return idFromStem(strings.TrimSuffix(name, "-rt2gdc.log"))
}

// RecordingIDOfAny 同上，但所有"属于一份录制"的文件后缀都认：三个流量文件 + 帧索引。
//
// 索引（`.frames.jsonl`）必须在这里认出来，否则 `library.Promote` 只搬 raw 不搬索引 ——
// 搬到长期库的那份看起来是完整的，其实明细页会退化成"块级时间"，而且**没人会知道为什么**
// （索引还留在缓存目录里，等着被清缓存删掉）。
func RecordingIDOfAny(name string) (string, bool) {
	if raw, ok := RawOfIndex(name); ok {
		name = raw // 索引与它的 raw 同一个干，先换回 raw 名再切 ID
	}
	for _, s := range []string{"-rt2gdc.log", "-gdc2rt.log", "-timeline.log"} {
		if strings.HasSuffix(name, s) {
			return idFromStem(strings.TrimSuffix(name, s))
		}
	}
	return "", false
}

func idFromStem(base string) (string, bool) {
	if base == "" {
		return "", false
	}
	i := strings.LastIndex(base, "-")
	if i <= 0 {
		return base, true
	}
	conn := base[i+1:]
	for _, c := range conn {
		if c < '0' || c > '9' {
			return base, true
		}
	}
	return base[:i], true
}
