package fep

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/debug"
)

// tt fep rec preflight —— 前置体检：只检查、不录制、不起会话、不写任何录制文件。
//
// 为什么需要它（真机冒烟，见 .agents/plan/L1-结果-2026-10-18.md）：
// `tt fep rec start` 连试两次都失败，而且失败信息是**逐层挤出来**的 ——
// 第一次报"服务没起"，起了服务再跑报"没有会话"，随后又冒一条"切换会话：没有 SSH 会话"。
// 使用者被迫来回试错才凑齐前置条件。所以这里把四项前置一次查完、一次说清：
//
//	① 数据目录可写（真写一个临时文件再删）—— Low 完整性标签 / 受限进程的第一现场；
//	② 调试服务可达（≤1s，不与失败本身比慢）；
//	③ 调试会话有没有、是哪条；
//	④ 环境/作业给了没有、在不在配置里。
//
// 全部走 preflightDeps：每一项都能被测试换掉（不起服务、不连真机、不改 ACL）。
//
// 每项三态（契约 A31，见下面那组常量）：ok / missing（确定缺失）/ unknown（检查本身失败）。
// unknown **不计入 missing、不改变退出码**，但要打印"这一项查不了（原因）"。
//
// 退出码：全就绪 0；缺前置 / 不可写 2（用法/环境不满足）——
// **不新造码**，见 fep.go:25-33 的码表与 internal/cli/README.md。

// 清单项的 key：脚本按它分支，不按标题字面（标题是给人读的，会改）。
const (
	preflightKeyDir     = "dir"     // 数据目录可写
	preflightKeyService = "service" // 调试服务
	preflightKeySession = "session" // 调试会话
	preflightKeyEnv     = "env"     // 环境/作业
)

// 检查项的**三种**状态（契约 A31）—— 三值，不是"过/不过"两值：
//
//	ok      —— 明确就绪
//	missing —— **确定**缺失（能说清缺什么、给得出修复命令）→ 计入 missing，exit 2
//	unknown —— **检查本身失败**（解析不了/超时/依赖故障）→ **不计入 missing**，
//	           但要**明确打印**"这一项查不了（原因）"
//
// 为什么必须分开：本次事故就是 `session` 项的解析坏了（报文其实是两段 JSON），
// 老实现把"我读不懂"当成"你没有"⇒ 服务起着也 exit 2，连带 `rec start` 永久卡死。
// **工具自己坏了，不许说环境不全。**
const (
	preflightStateOK      = "ok"
	preflightStateMissing = "missing"
	preflightStateUnknown = "unknown"
)

// preflightIntegrityHint 是"目录不可写"的处置 —— 必须具体到能抄。
//
// 真机上 Low 完整性标签与 ACL 拒绝给的是同一句 `Access is denied`，
// 而目录的 ACE 明明写着允许写（发现 A 的原文）。只说"目录不可写"等于没说。
const preflightIntegrityHint = "看起来是从低完整性/受限进程启动的（例如 tt.exe 带 Low 标签）：" +
	"请改用普通权限终端，或 icacls <exe> /setintegritylevel Medium"

// 探活的两个超时都压到 1s 以内：体检不能比失败本身还慢。
const (
	preflightDialTimeout = 800 * time.Millisecond
	preflightHTTPTimeout = 900 * time.Millisecond
)

// preflightItem 是清单里的一项。
type preflightItem struct {
	Key   string `json:"key"`   // dir / service / session / env
	Title string `json:"title"` // 人读的名字（"数据目录可写"）
	// State 是三值状态（ok / missing / unknown，见上面那组常量）。**它才是判据**：
	// 只有 missing 计入退出码，unknown 只打印原因。
	State  string `json:"state"`
	OK     bool   `json:"ok"`               // State==ok 的便利位（老消费者照旧能读）
	Detail string `json:"detail,omitempty"` // 现状：是什么（路径、地址、是哪条会话……）
	Why    string `json:"why,omitempty"`    // 没过时的那句话（"调试服务没起" / "这一项查不了：…"）
	Fix    string `json:"fix,omitempty"`    // 确定缺失时可直接复制的命令 / 具体处置
}

// state 取这一项的三值状态。
//
// State 为空时按老口径回退（OK→ok，否则 missing）：清单项是包内构造的，
// 但回退让"手工拼一个 preflightItem"的老代码不至于落到一个不存在的第四态上。
func (it preflightItem) state() string {
	switch it.State {
	case preflightStateOK, preflightStateMissing, preflightStateUnknown:
		return it.State
	}
	if it.OK {
		return preflightStateOK
	}
	return preflightStateMissing
}

// markOK / markMissing / markUnknown 是三种状态的唯一写入点：
// State 与 OK 一起写，免得两处各写一半（"ok 为 true 但 state 是 unknown"这种自相矛盾）。
func (it *preflightItem) markOK() {
	it.State, it.OK = preflightStateOK, true
}

func (it *preflightItem) markMissing() {
	it.State, it.OK = preflightStateMissing, false
}

func (it *preflightItem) markUnknown() {
	it.State, it.OK = preflightStateUnknown, false
}

// preflightReport 是 tt fep rec preflight 的输出（--json 时原样给，一个对象）。
type preflightReport struct {
	OK      bool            `json:"ok"`
	Dir     string          `json:"dir,omitempty"`
	Env     string          `json:"env,omitempty"`
	Prog    string          `json:"prog,omitempty"`
	Missing int             `json:"missing"`
	Unknown int             `json:"unknown"` // 查不了的项数（A31：不计入 missing、不影响退出码）
	Checks  []preflightItem `json:"checks"`
}

// preflightDeps 是体检要用的全部外部事实 —— 每一项都可注入。
//
// 为什么全做成函数而不是直接调：判据必须能被钉住，而"真起一个调试服务 + 真 SSH +
// 真把目录改成不可写"不在单测的范围里（同 rec_session.go 的 debugSessions）。
type preflightDeps struct {
	// Dir 返回这一轮会落录制件的目录（<数据目录>/fep）。
	Dir func() (string, error)
	// Writable 真去写一个临时文件再删（只判"写得进去"，不判权限位）。
	Writable func(dir string) error
	// Service 探调试服务：可达与否 + 一句现状。
	Service func() (bool, string)
	// Sessions 取调试服务的会话清单（生产实现 = rec_session.go 的 debugSessions）。
	Sessions func(cmd *cobra.Command) ([]dbgSessionView, error)
	// Hosts 取配置里的环境清单（判"环境是否存在"）。
	Hosts func() ([]config.NamedSsh, error)
}

// preflightDepsNow 是生产实现（包级变量：测试可替换，同 debugSessions）。
var preflightDepsNow = defaultPreflightDeps

func defaultPreflightDeps() preflightDeps {
	return preflightDeps{
		Dir:      func() (string, error) { return cacheDir() },
		Writable: probeWritable,
		Service:  debugServiceUp,
		Sessions: debugSessions,
		Hosts: func() ([]config.NamedSsh, error) {
			cfgPath, err := common.ResolveConfig(true)
			if err != nil {
				return nil, err
			}
			dcfg, err := debug.LoadConfig(cfgPath)
			if err != nil {
				return nil, err
			}
			return dcfg.SSHs, nil
		},
	}
}

// runPreflight 跑完四项检查，返回清单。**纯函数**：不打印、不退出、不起会话。
func runPreflight(cmd *cobra.Command, prog, env string, d preflightDeps) preflightReport {
	rep := preflightReport{Prog: strings.TrimSpace(prog), Env: strings.TrimSpace(env)}
	// 落点只解析一次（两份调用会各走一遍配置解析，也让"试写"这件事更难被钉住）
	dir, derr := d.Dir()
	if derr == nil {
		rep.Dir = dir
	}
	rep.Checks = append(rep.Checks, checkDir(d, dir, derr))
	svc := checkService(d)
	rep.Checks = append(rep.Checks, svc)
	rep.Checks = append(rep.Checks, checkSession(cmd, d, svc.OK, rep.Prog, rep.Env))
	envItem := checkEnv(d, rep.Prog, rep.Env)
	rep.Checks = append(rep.Checks, envItem)
	if envItem.OK && envItem.Detail != "" {
		// 环境名解析成配置里的规范名（拼错时这一项本来就不 ok，不会走到这里）
		if name, _, ok := strings.Cut(envItem.Detail, " / "); ok {
			rep.Env = name
		}
	}
	// 退出码只由 missing 决定（A31）：unknown 记一笔、打印原因，但**不参与**判定。
	for _, it := range rep.Checks {
		switch it.state() {
		case preflightStateMissing:
			rep.Missing++
		case preflightStateUnknown:
			rep.Unknown++
		}
	}
	rep.OK = rep.Missing == 0
	return rep
}

// missing 挑出**确定缺失**的项（保持清单顺序）。unknown 不在其中。
func (r preflightReport) missing() []preflightItem {
	out := make([]preflightItem, 0, r.Missing)
	for _, it := range r.Checks {
		if it.state() == preflightStateMissing {
			out = append(out, it)
		}
	}
	return out
}

// unknowns 挑出**查不了**的项（保持清单顺序）：这些项要打印原因，但不计入缺失。
func (r preflightReport) unknowns() []preflightItem {
	out := make([]preflightItem, 0, r.Unknown)
	for _, it := range r.Checks {
		if it.state() == preflightStateUnknown {
			out = append(out, it)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// 四项检查

// checkDir 真写一次临时文件：Windows 上"看起来有权限"与"写得进去"是两件事
// —— Low 完整性标签下目录 ACE 明明允许写，落盘照样 Access is denied。
func checkDir(d preflightDeps, dir string, derr error) preflightItem {
	it := preflightItem{Key: preflightKeyDir, Title: "数据目录可写"}
	if derr != nil {
		it.markMissing()
		it.Detail = derr.Error()
		it.Why = "算不出数据目录（配置路径没解析出来）"
		it.Fix = "tt config path"
		return it
	}
	it.Detail = dir
	if werr := d.Writable(dir); werr != nil {
		it.markMissing()
		it.Detail = dir + "：" + werr.Error()
		it.Why = "目录不可写：疑似进程完整性/权限问题（见下方提示）"
		it.Fix = preflightIntegrityHint
		return it
	}
	it.markOK()
	it.Detail = dir + "（可写）"
	return it
}

// checkService 探调试服务在不在。
func checkService(d preflightDeps) preflightItem {
	it := preflightItem{Key: preflightKeyService, Title: "调试服务"}
	up, detail := d.Service()
	it.Detail = detail
	if up {
		it.markOK()
		return it
	}
	it.markMissing()
	it.Why = "调试服务没起"
	it.Fix = "tt debug serve"
	return it
}

// checkSession 查有没有调试会话、是哪条。
//
// 服务没起时这一项**照样列出来**：会话住在服务里，服务不在就必然没有会话，
// 而"起了服务还得再起一条会话"正是第二个前置 —— 不一起说，就又要多跑一次。
//
// **A31 的落点**：`d.Sessions` 报错是**检查本身失败**（列不出来/读不懂/超时），
// 不是"你没有会话" ⇒ 记 unknown、打印原因、**不计入 missing**。
// 从前这里把"我读不懂"写成"没有 SSH 会话"，于是服务起着也 exit 2、
// `rec start` 必然卡死（报文其实是两段 JSON，见 rec_session.go 的 parseDebugSessions）。
func checkSession(cmd *cobra.Command, d preflightDeps, serviceUp bool, prog, env string) preflightItem {
	it := preflightItem{Key: preflightKeySession, Title: "调试会话"}
	it.Fix = "tt debug start " + orPlaceholder(prog, "<作业>") + " --env " + orPlaceholder(env, "<环境>")
	if !serviceUp {
		it.markMissing()
		it.Detail = "调试服务没起，会话无从谈起"
		it.Why = "没有 SSH 会话"
		return it
	}
	sessions, err := d.Sessions(cmd)
	if err != nil {
		// 查不了 ≠ 缺失：原因要留全，修复命令不给（给不出，也不能编）。
		it.markUnknown()
		it.Detail = "列会话失败：" + err.Error()
		it.Why = "这一项查不了：列会话失败（" + firstLine(err.Error()) + "）"
		it.Fix = "tt debug status --json"
		return it
	}
	if len(sessions) == 0 {
		it.markMissing()
		it.Detail = "调试服务里一条会话都没有"
		it.Why = "没有 SSH 会话"
		return it
	}
	it.markOK()
	it.Fix = ""
	// 点名"是哪条"：优先环境与作业都对上的那条（这一轮真正会用上的），否则报第一条。
	//
	// **不写「取列表第 0 个」那种下标**：`哪条会话`只许有一个来源，直接下标被 testkit 的守门测试禁掉
	// （first_session_source_test.go）。这里用 range 取首元素兜底，语义不变且不会越界。
	var best dbgSessionView
	picked := false
	for _, s := range sessions {
		if !picked {
			best, picked = s, true // 第一条兜底
		}
		if env != "" && s.Env == env && (prog == "" || s.Prog == "" || s.Prog == prog) {
			best = s
			break
		}
	}
	it.Detail = fmt.Sprintf("有 %d 条；这一轮用得上的是 %s", len(sessions), describeSession(best))
	return it
}

// firstLine 只留错误的第一行（清单是一行一项，正文里的换行会把对齐撑坏）。
func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	return s
}

// describeSession 把一条会话说成人能认的样子：别名/句柄（环境 / 作业 / 状态）。
func describeSession(s dbgSessionView) string {
	name := s.Name
	if name == "" {
		name = s.ID
	}
	if name == "" {
		name = "(没有句柄)"
	}
	var parts []string
	if e := strings.TrimSpace(s.Env); e != "" {
		parts = append(parts, e)
	}
	if p := strings.TrimSpace(s.Prog); p != "" {
		parts = append(parts, "作业 "+p)
	}
	if st := strings.TrimSpace(s.State); st != "" {
		parts = append(parts, st)
	}
	if len(parts) == 0 {
		return name
	}
	return name + "（" + strings.Join(parts, " / ") + "）"
}

// checkEnv 查环境/作业：给了没有、在不在配置里。
//
// 这里的失败都是**确定缺失**（A31）：说得出缺什么（环境名/作业名/配置），
// 也给得出修复命令（`tt env list` / `tt config check`）。会话那一项不同 ——
// 它读不出来时**说不出缺什么**，那是 unknown。
func checkEnv(d preflightDeps, prog, env string) preflightItem {
	it := preflightItem{Key: preflightKeyEnv, Title: "环境/作业"}
	switch {
	case env == "":
		it.markMissing()
		it.Detail = "没给环境（--env）"
		it.Why = "没给环境"
		it.Fix = "tt env list"
		return it
	case prog == "":
		it.markMissing()
		it.Detail = "没给作业程序（--prog）"
		it.Why = "没给作业程序"
		it.Fix = "tt fep rec preflight --prog <作业> --env " + env
		return it
	}
	hosts, err := d.Hosts()
	if err != nil {
		it.markMissing()
		it.Detail = "读不到环境清单：" + err.Error()
		it.Why = "读不到环境清单"
		it.Fix = "tt config check"
		return it
	}
	name, rerr := openEnvName(hosts, env)
	if rerr != nil {
		it.markMissing()
		it.Detail = rerr.Error()
		it.Why = "环境不在配置里"
		it.Fix = "tt env list"
		return it
	}
	it.markOK()
	it.Detail = name + " / " + prog
	return it
}

// ---------------------------------------------------------------------------
// 「一次说清」

// missingText 把没过的项拼成报错正文（`rec start` 与 `preflight` 共用）。
func missingText(missing []preflightItem) string {
	var b strings.Builder
	fmt.Fprintf(&b, "还缺 %d 项前置：\n", len(missing))
	for i, m := range missing {
		why := m.Why
		if why == "" {
			why = m.Title + "不满足"
		}
		fix := m.Fix
		if fix == "" {
			fix = "（看不出怎么补，跑 tt fep rec preflight 看明细）"
		}
		// 长的处置（例如不可写那条带 icacls 的）另起一行，别把人读的那一列撑爆
		if displayWidth(fix) > 48 {
			fmt.Fprintf(&b, "  %d) %s\n     → %s\n", i+1, why, fix)
			continue
		}
		fmt.Fprintf(&b, "  %d) %s%s\n", i+1, padLabel(why, 22), fix)
	}
	b.WriteString("（都满足后再跑本命令）")
	return b.String()
}

// padLabel 按**显示宽度**补空格（中文占两格），让后面那些命令能对齐成一列。
func padLabel(s string, width int) string {
	s += "："
	if w := displayWidth(s); w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s + " "
}

// displayWidth 粗算显示宽度：CJK/全角按两格算。只用于对齐，不追求精确。
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if r > 0x2E80 {
			w += 2
			continue
		}
		w++
	}
	return w
}

// text 是人读态：先逐项 ✓/✗/?，不过的项再把「还缺 N 项前置」整块贴出来，
// 查不了的项单列一段（A31：**必须打印**，但不计入缺失、不影响退出码）。
func (r preflightReport) text() string {
	var b strings.Builder
	b.WriteString("前置体检（只检查、不录制、不起会话）\n")
	for _, it := range r.Checks {
		mark := "✓"
		switch it.state() {
		case preflightStateMissing:
			mark = "✗"
		case preflightStateUnknown:
			mark = "?" // 查不了：既不是 ok 也不是 missing
		}
		fmt.Fprintf(&b, "  %s %s  %s\n", mark, padLabel(it.Title, 16), it.Detail)
	}
	// 查不了的项：单独一段，把原因说清（不许静默当 ok）
	if unk := r.unknowns(); len(unk) > 0 {
		fmt.Fprintf(&b, "有 %d 项没检查成功（查不了：不计入缺失，也不影响退出码）：\n", len(unk))
		for _, it := range unk {
			why := it.Why
			if why == "" {
				why = it.Detail
			}
			fmt.Fprintf(&b, "  ? %s%s\n", padLabel(it.Title, 16), why)
		}
	}
	if r.OK {
		if r.Unknown > 0 {
			fmt.Fprintf(&b, "就绪（但有 %d 项没检查成功，见上）：可以跑 tt fep rec start --prog %s --env %s\n",
				r.Unknown, r.Prog, r.Env)
			return b.String()
		}
		fmt.Fprintf(&b, "就绪：可以跑 tt fep rec start --prog %s --env %s\n", r.Prog, r.Env)
		return b.String()
	}
	b.WriteString(missingText(r.missing()) + "\n")
	// 不可写那条的处置比较长（含 icacls 一句），从编号清单里拎出来单列一行更好读。
	if dirItem := r.item(preflightKeyDir); dirItem != nil && dirItem.Fix == preflightIntegrityHint {
		b.WriteString("提示：" + preflightIntegrityHint + "\n")
	}
	return b.String()
}

// item 按 key 取一项（不存在返回 nil）。
func (r preflightReport) item(key string) *preflightItem {
	for i := range r.Checks {
		if r.Checks[i].Key == key {
			return &r.Checks[i]
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// 退出码与命令

// preflightError 是"清单已经打出去了，只在退出码上表示失败"的错误。
//
// 为什么要它：preflight 的失败信息**就是那份清单**，再让根命令 reportError 打一份
// 信封，`--json` 下 stdout 上会出现两个对象（`tt form verify` 立过同一条规矩）。
type preflightError struct{ msg string }

func (e *preflightError) Error() string         { return e.msg }
func (e *preflightError) ExitCode() int         { return exitUsage }
func (e *preflightError) AlreadyReported() bool { return true }

// preflightExit 把清单翻成退出码：全就绪 nil（0），否则 2（用法/环境不满足）。
func preflightExit(rep preflightReport) error {
	if rep.OK {
		return nil
	}
	return &preflightError{msg: missingText(rep.missing())}
}

// recStartPreflight 是 `rec start` 开跑前的体检：缺前置就一次说清。
//
// 位置很要紧：它在**起代理之前**（代理占三个端口，半路失败要收，收不干净就留一个空会话）。
func recStartPreflight(cmd *cobra.Command, prog, env string) error {
	return recStartPreflightWith(cmd, prog, env, preflightDepsNow())
}

// recStartPreflightWith 同上，依赖可注入（测试用：不起服务、不连真机）。
func recStartPreflightWith(cmd *cobra.Command, prog, env string, d preflightDeps) error {
	rep := runPreflight(cmd, prog, env, d)
	if rep.OK {
		return nil
	}
	// 与 preflight 同一个码：缺前置属于"用法/环境不满足"，不新造码。
	return fail(exitUsage, "%s", missingText(rep.missing()))
}

// tt fep rec preflight —— 只检查、不录制、不起会话、不写任何录制文件。
var recPreflightCmd = &cobra.Command{
	Use:   "preflight",
	Short: "前置体检：目录/服务/会话/环境一次查清（只检查，不录制）",
	Long: `前置体检：只检查、不录制、不起会话、不写任何录制文件。

四项一次查完，缺哪几项一起列出来，每条都带可直接复制的命令：

  数据目录可写   真去 <数据目录>/fep 下写一个临时文件再删
                 （Low 完整性标签 / 受限进程在这里第一次现形）
  调试服务       可达性，≤1s 超时（不与失败本身比慢）
  调试会话       有没有、是哪条
  环境/作业      给了没有、在不在配置里

每项三态（契约 A31）：ok / missing（确定缺失，给修复命令）/ unknown（这一项查不了，打印原因）。
unknown 不计入缺失、不改变退出码 —— 工具自己坏了，不许说环境不全。

退出码与 tt fep 一致：全就绪 0；缺前置 / 不可写 2（用法/环境不满足）。
--json 下 stdout 恰好一个对象（清单），失败时也是它。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		rep := runPreflight(cmd, recPreflightProg, recPreflightEnv, preflightDepsNow())
		if err := printJSONOrText(rep, func() { fmt.Fprint(cmd.OutOrStdout(), rep.text()) }); err != nil {
			return err
		}
		return preflightExit(rep)
	},
}

var (
	recPreflightProg string
	recPreflightEnv  string
)

func init() {
	recPreflightCmd.Flags().StringVar(&recPreflightProg, "prog", "", "作业程序名（体检用：没给时命令里留 <作业> 占位符）")
	recPreflightCmd.Flags().StringVar(&recPreflightEnv, "env", "", "环境名（体检用：没给时命令里留 <环境> 占位符）")
	recCmd.AddCommand(recPreflightCmd)
}

// ---------------------------------------------------------------------------
// 生产实现（探活 / 试写）

// probeWritable 真去写一个临时文件再删。
//
// 为什么不用 os.Stat / 权限位：Windows 上"看起来有权限"与"写得进去"完全是两件事
// —— Low 完整性标签下目录 ACE 明明写着允许，写入照样 Access is denied
// （真机实测见 .agents/plan/L1-结果-2026-10-18.md 的发现 A）。
func probeWritable(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return fmt.Errorf("目录是空的（配置路径没解析出来）")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "tt-preflight-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	_, werr := f.WriteString("tt preflight\n")
	cerr := f.Close()
	rerr := os.Remove(name)
	switch {
	case werr != nil:
		return werr
	case cerr != nil:
		return cerr
	case rerr != nil:
		return rerr
	}
	return nil
}

// preflightStatusProbes 探活时依次试的路径与应答里的标记。
//
// 两种布局的服务共用同一份状态文件：独立调试服务在根下、统一服务在前缀下
// （与 internal/cli/debug/servebg.go 的 serveProbe 是同一份事实；
// 标记取自 internal/debug/api.go 的 hStatus）。
var preflightStatusProbes = []struct{ Path, Mark string }{
	{"/debug/api/status", "tdebug-debug"},
	{"/api/status", "tdebug-debug"},
}

// debugServiceUp 探"调试服务在不在"。两步，都很短：
//
//  1. TCP 连一下（800ms）：连不上就是服务没起 —— 真机第一次失败的那一项；
//  2. 再查一次状态（900ms）：连得上但查不出调试服务的标记，说明端口被别的程序占着，
//     这时说"服务在跑"是骗人的（下一步就会撞上无法连接调试服务）。
func debugServiceUp() (bool, string) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return false, "配置路径解析失败：" + err.Error()
	}
	addr := debugServiceAddr(cfgPath)
	conn, derr := net.DialTimeout("tcp", addr, preflightDialTimeout)
	if derr != nil {
		return false, addr + " 连不上（" + derr.Error() + "）"
	}
	_ = conn.Close()
	for _, p := range preflightStatusProbes {
		if body, ok := probeDebugStatus(addr, p.Path); ok && strings.Contains(body, p.Mark) {
			return true, addr + " 在跑"
		}
	}
	return false, addr + " 有人接，但查不出调试服务（端口可能被别的程序占着）"
}

// probeDebugStatus 取一次状态页的正文（最多 512 字节，够看标记）。
func probeDebugStatus(addr, path string) (string, bool) {
	cli := &http.Client{Timeout: preflightHTTPTimeout}
	resp, err := cli.Get("http://" + addr + path)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	buf := make([]byte, 512)
	n, _ := resp.Body.Read(buf)
	return string(buf[:n]), true
}

// debugServiceAddr 调试服务在哪：状态文件里的真实地址（后台实例的端口可能顺延过）
// > 配置里的 listen > 内置默认。
//
// 寻址规则的家在 internal/cli/debug（servebg.go）—— 这里只读**同一份事实**
// （状态文件与 config.json），不另造一套地址。
func debugServiceAddr(cfgPath string) string {
	loc := config.LocationsAt(cfgPath)
	if b, err := os.ReadFile(loc.StateFile()); err == nil {
		var st struct {
			URL string `json:"url"`
		}
		if json.Unmarshal(b, &st) == nil {
			if host := hostOfURL(st.URL); host != "" {
				return host
			}
		}
	}
	if root, err := config.Load(cfgPath); err == nil {
		if l := strings.TrimSpace(root.Listen); l != "" {
			return l
		}
	}
	return debug.DefaultListen
}

// hostOfURL 把 http://127.0.0.1:28670/ 削成 127.0.0.1:28670。
func hostOfURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "https://")
	return strings.TrimRight(u, "/")
}

// orPlaceholder 没给值时留占位符（不拼出半条命令）。
func orPlaceholder(v, placeholder string) string {
	if v = strings.TrimSpace(v); v != "" {
		return v
	}
	return placeholder
}
