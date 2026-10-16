package fep

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/debug"
	"tt/internal/fep"
)

// recSession 是"起一轮调试线录制会话"要的那些参数。
//
// 抽出来是因为有两个调用方：`rec start`（人开始录一轮）与 `rec reproduce`
// （把一份录制在断点下重放出来）。谁起的会话谁负责收 —— 这里只负责起。
type recSession struct {
	Ctx    context.Context
	Env    string
	Prog   string
	Ent    string
	Module string
	// Breaks 是**放行前**要下好的断点（`file:line`、`函数名`，见 tt debug exec 的 break）。
	// 顺序要紧：程序停在入口时下断点才有意义；放行之后再下，那一段已经跑过去了。
	Breaks []string
	// NoContinue 只下断点、不放行（给人自己下断点用：`rec start --no-continue`）。
	NoContinue bool
	Origin     fep.Origin
	// Kind 是这一轮会话的用途（`rec` / `replay`）：只进会话登记，让面板能分清
	// "人还在录"与"回放发完留下的窗口"（见 internal/fep/session.go 那一栏的说明）。
	Kind string
	// RecID 这一轮录制对应的**录制 ID**（= 代理落盘用的 tag）。它同时是**跨进程传目标的
	// 来源**：`recSessionAlias` 由它算出固定别名（T5.4 / 契约 C7.3），别名随 `--name`
	// 进 `debug start` 子进程，之后取证命令用同一个别名点回同一条会话。
	// 空 = 由本函数按本轮 tag 现算（`rec start` 那条路就是这样：tag 由它自己生成）。
	RecID string
}

// openDebugRecording 起隔离实例 + 代理，再把调试线的作业挂到这条代理上。
//
// 顺序不能换：代理先起来（画面得有地方去），再让调试线把作业挂上去（`--fglserver` 指到代理）。
// 中间任何一步失败都把已经起来的代理收掉 —— 不然会留下一个占着三个端口的空会话。
// 返回的 Runtime 归调用方：它决定什么时候 Wait/Close（`rec reproduce` 就是故意不收，
// 好让 AI 停在断点上取证）。
func openDebugRecording(cmd *cobra.Command, s recSession) (*fep.Runtime, string, error) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return nil, "", err
	}
	dcfg, err := debug.LoadConfig(cfgPath)
	if err != nil {
		return nil, "", fail(exitFailed, "%v", err)
	}
	name, rerr := openEnvName(dcfg.SSHs, s.Env)
	if rerr != nil {
		return nil, "", fail(exitUsage, "%v", rerr)
	}
	clone := dcfg.CloneEnv(name)
	if clone == nil {
		return nil, "", fail(exitUsage, "环境 %q 不存在（见 tt env list）", name)
	}
	dcfg = clone
	if dcfg.SSH.Host == "" {
		return nil, "", fail(exitPreflight, "环境 %q 没配 SSH host", name)
	}
	loc := config.LocationsAt(cfgPath)
	recDir := loc.CacheDir("fep")
	if recDir == "" {
		return nil, "", fail(exitPreflight, "缓存目录没解析出来（检查配置路径）")
	}
	if err := os.MkdirAll(recDir, 0o755); err != nil {
		return nil, "", fail(exitFailed, "建录制目录失败: %v", err)
	}
	// 上一轮可能还挂着一个调试会话（程序释放了、宿主会话还在）：那种状态下 TOPENT 设不上去
	// （服务端只允许空闲时设）。这一轮要的是**自己的**会话，所以先把它收掉。
	// 不收的后果实测过：`rec reproduce` 在 `debug topent` 那一步退 1，白起一个代理。
	logf := func(format string, a ...any) {
		fmt.Fprintf(cmd.ErrOrStderr(), "· "+format+"\n", a...)
	}
	if why, err := staleSessionToCollect(cmd, s, dcfg.EnvName()); err == nil && why != "" {
		logf("先收掉上一轮挂着的调试会话（它占着宿主会话，TOPENT 设不上）：%s", why)
		if qerr := runSelf(cmd, "debug", "quit"); qerr != nil {
			logf("收上一轮会话时失败（继续试）：%v", qerr)
		}
		time.Sleep(2 * time.Second)
	}
	alias := recSessionAlias(s.RecID)
	tag := "fep-" + time.Now().Format("20060102-150405")
	if alias == "" {
		// 调用方没给录制 ID（老调用方）：这一轮的 tag 就是它的录制 ID，
		// 别名按同一个来源现算 —— 两条路算出同一个名字（T5.4 / C7.3）。
		alias = recSessionAlias(tag)
	}
	rt, err := fep.Open(s.Ctx, fep.OpenConfig{
		Env: dcfg.EnvName(), Zone: dcfg.Zone,
		// 身份：作业是调试线起的（见 path.go）。写对了，回放前的路径比对才有双方的值。
		Prog: s.Prog, Path: fep.PathDebug, Ent: s.Ent, Origin: s.Origin, Kind: s.Kind,
		SSH: dcfg.SSH, DataDir: loc.DataDir(), RecordDir: recDir, Tag: tag,
		GDCLog: filepath.Join(recDir, tag+"-gdc.log"),
		NoJob:  true,
	}, logf)
	if err != nil {
		return nil, "", mapOpenErr(err)
	}
	fgl := rt.FGLServer()
	if fgl == "" {
		_ = rt.Close()
		return nil, "", fail(exitPreflight, "算不出这一轮的画面接线（FGLSERVER）：给环境配 SSH host，或用 --fgserver-ip")
	}
	logf("这一轮画面接线：FGLSERVER=%s（调试线按它把作业画到本代理上）", fgl)

	if serr := runRecStartSteps(cmd, recStartSteps(dcfg.EnvName(), s.Ent, s.Prog, s.Module, fgl, alias), s.Breaks, s.NoContinue); serr != nil {
		if cerr := rt.Close(); cerr != nil {
			logf("收代理时还有失败：%v", cerr)
		}
		return nil, "", serr
	}
	return rt, fgl, nil
}

// recSessionAlias 这一轮跨进程传目标用的**固定别名**（T5.4 / 契约 C7.3）。
//
// 别名由**录制身份**算出来，调用方不必给：同一条录制（同一个 tag）在 rec start 与
// rec reproduce 两条路上算出同一个名字。它随 `--name` 进 `debug start` 子进程，
// 之后（WP4c 把别名变成服务端事实之后）取证命令用 `--session <同一个别名>` 点回同一条会话。
//
// 为什么是"录制 ID"而不是"临时起的随机名"：Q23 的别名是**幂等键**——
// 重发同一条命令必须落回同一条会话，随机名每次都算新会话，等于没有幂等键。
//
// tag 为空（老调用方只填 Env/Prog）时返回空串：那种情况下不拼 `--name`，
// 既有行为一字不变（H2）。
func recSessionAlias(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ""
	}
	return "fep-" + tag
}

// staleSessionToCollect 判断"上一轮挂着的调试会话"里有没有**属于本次录制**的那一条。
//
// 返回 (人话说明, error)；没有该收的就返回 ("", nil)。
//
// **这是缺陷 D（T5.5）的落点**：从前这里只看"有没有会话在跑"——
//
//	if busy, _ := debugSessionOpen(cmd); busy { runSelf(cmd, "debug", "quit") }
//
// 多会话之后，B 开始录制会把正在跑断点的 A 的会话一枪 `quit` 掉，而 A 那边
// 看起来只是"会话莫名其妙没了"。所以判据先按**环境**收窄（T5.5 的第一半），
// 再按**录制身份**收窄（第二半：只能收自己那条）。
//
// 环境这一半今天就能做，而且是**定向**的那一半：`rec start --env X` 起会话时
// 服务端只会连到 X，所以上一轮挂着的、该收口的就只可能是 X 上的会话。
// 别人的会话（别的环境）**一条都不碰**。
func staleSessionToCollect(cmd *cobra.Command, s recSession, env string) (string, error) {
	sessions, err := debugSessions(cmd)
	if err != nil {
		return "", err
	}
	for _, sess := range sessions {
		if !recSessionOwned(sess, s, env) {
			continue
		}
		why := sess.Env + " 上的调试会话"
		if sess.Prog != "" {
			why += "（作业 " + sess.Prog + "）"
		}
		return why, nil
	}
	return "", nil
}

// recSessionOwned 判一条会话是不是"本次录制该收的那一条"。
//
// 判据（按**先在的先判**）：
//  1. **环境必须点名且相等** —— 环境不明的会话一律不认领（宁可多留一条，也不误杀
//     别人正在跑断点的会话，缺陷 D 的验收就是这条）；
//  2. **录制身份**（今天能拿到的部分 = 作业）：给了 `--prog` 而对方的作业不同，
//     那就不是我们起的那一条，跳过。
func recSessionOwned(sess dbgSessionView, s recSession, env string) bool {
	if env == "" || sess.Env == "" || sess.Env != env {
		return false
	}
	if prog := strings.TrimSpace(s.Prog); prog != "" && sess.Prog != "" && sess.Prog != prog {
		return false
	}
	return true
}

// dbgSessionView 是 `tt debug status --json` 里我们认得出的那几栏。
//
// 只读它给的 JSON，不碰调试服务的内部协议 —— 那个协议的家在 internal/cli/debug。
// `name`（别名）服务端今天还没吐（C5.3 的五个键是 WP4c 的活），所以这里也不依赖它；
// 等它有了，`recSessionOwned` 再加一条"别名相同"的判据即可（见契约 C7.3）。
type dbgSessionView struct {
	ID    string `json:"id"`
	Env   string `json:"env"`
	Prog  string `json:"prog"`
	State string `json:"state"`
	Name  string `json:"name"`
}

// debugSessions 取调试服务里的会话列表（`tt debug status --json`）。
//
// 它可以被测试替换（包级变量）:判据函数与"要不要收口"这一步必须能一起被钉住，
// 而真起一个调试服务 + 真 SSH 不在单测的范围里。
var debugSessions = debugSessionsReal

// debugSessionsReal 是生产实现:走子进程读 JSON。
func debugSessionsReal(cmd *cobra.Command) ([]dbgSessionView, error) {
	out, err := runSelfOut(cmd, "--json", "debug", "status")
	if err != nil {
		return nil, err
	}
	return parseDebugSessions(out)
}

// parseDebugSessions 读 `tt debug status --json` 的报文。
//
// 形状（实测）：一个 `{"sessions":[…]}` 文档。从前这里的 struct 只声明了 `state`，
// 于是"这条会话是谁的"这件事**在客户端根本没有材料**—— 那就是缺陷 D 能成立的原因之一。
func parseDebugSessions(out []byte) ([]dbgSessionView, error) {
	var st struct {
		Sessions []dbgSessionView `json:"sessions"`
	}
	if jerr := json.Unmarshal(out, &st); jerr != nil {
		return nil, jerr
	}
	return st.Sessions, nil
}

// stopSpot 是"程序停在哪儿"（调试服务给的现场）。
type stopSpot struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Func   string `json:"func"`
	Reason string `json:"reason"`
}

func (s stopSpot) String() string {
	if s.File == "" {
		if s.Reason != "" {
			return "（没说在哪个文件）原因：" + s.Reason
		}
		return "（没有停站信息）"
	}
	spot := fmt.Sprintf("%s:%d", s.File, s.Line)
	if s.Func != "" {
		spot += "（" + s.Func + "）"
	}
	if s.Reason != "" {
		spot += " 原因：" + s.Reason
	}
	return spot
}

// waitStopped 等调试会话停站，并把现场读出来。
//
// 走 `tt debug wait --for stopped --json`（子进程，输出落临时文件再读）：
//   - 不直接讲调试服务的 HTTP 协议 —— 那个协议只该有一份实现，它在 internal/cli/debug 里；
//   - 不用管道接输出：沙箱里管道会被拒，而临时文件哪都能用。
//
// 超时**不是错误**（与 `tt debug wait` 同一口径）：返回 timedOut=true，由调用方决定怎么报。
func waitStopped(cmd *cobra.Command, timeout time.Duration) (stopSpot, bool, error) {
	var spot stopSpot
	out, err := runSelfOut(cmd, "--json", "debug", "wait", "--for", "stopped,exit", "--timeout",
		fmt.Sprintf("%d", int(timeout.Seconds())))
	if err != nil {
		return spot, false, err
	}
	spot, timedOut, err := parseWaitResult(out)
	return spot, timedOut, err
}

// parseWaitResult 读 `tt debug wait --json` 的报文。
//
// 报文的形状（实测，别按直觉猜）：`event` 是**对象**不是字符串，而停站现场在
// 顶层的 `stop` 里也在 `event.stop` 里各有一份 —— 第一次写成 `Event string` 时，
// 断点明明命中了却报"读不懂等待结果"。
func parseWaitResult(out []byte) (stopSpot, bool, error) {
	var spot stopSpot
	var payload struct {
		Event struct {
			Type string    `json:"type"`
			Stop *stopSpot `json:"stop"`
		} `json:"event"`
		TimedOut bool      `json:"timedOut"`
		State    string    `json:"state"`
		Stop     *stopSpot `json:"stop"`
	}
	if jerr := json.Unmarshal(out, &payload); jerr != nil {
		return spot, false, fmt.Errorf("读不懂调试服务的等待结果: %w（原文 %s）", jerr, strings.TrimSpace(string(out)))
	}
	if payload.Stop != nil {
		spot = *payload.Stop
	} else if payload.Event.Stop != nil {
		spot = *payload.Event.Stop
	}
	// 事件到了但没有 stop 现场（退出/掉线）：也算"没捕获到断点现场"，照实报。
	return spot, payload.TimedOut || spot.File == "", nil
}
