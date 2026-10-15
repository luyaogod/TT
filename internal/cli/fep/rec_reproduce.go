package fep

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

// tt fep rec reproduce —— 把一份录制**在调试器下重放出来**，并报出两样证据：
// 指纹比对结果 + 程序停在哪儿。这是计划的终态目标那一步：
//
//	用户复现一次异常 → 得到录制 ID → AI 读代码、决定断点 → 这一条命令把它重放出来并捕获现场
//
// 为什么它不自己造一套回放：录制的重放在 `internal/fep` 的注入器上（`rec replay` 那套，
// 含指纹闸与 AUI 安全闸），调试器那一侧的命令面在 `tt debug`（子进程调，协议只一份）。
// 这条命令只负责**把两边按不能换的顺序串起来**，并把证据摆出来。
//
// 它与 `rec start` 的差别是"手不松开"：停在断点上之后**故意不收拾会话** ——
// AI 接着就要在那上面取证（`tt debug exec "print …"`）。收拾是人的事，命令会把怎么收说出来。

var (
	repBreak    []string
	repTimeout  time.Duration
	repAttached bool
	repEnv      string
	repProg     string
	repEnt      string
	repModule   string
	repConn     int
	repWindow   string
	repStopAt   int
	repSkip     []string
	repInterval time.Duration
	repWaitUI   time.Duration
)

var recReproduceCmd = &cobra.Command{
	Use:   "reproduce <录制 ID>",
	Short: "在调试器下重放这份录制，报出指纹比对结果与停站位置（AI 闭环那一步）",
	Long: `把一份录制在调试器下重放出来，并报出两样证据：**指纹比对结果**与**程序停在哪儿**。

它按顺序做四件事：
  1. 读录制身份（环境 / 作业 / ENT 从身份里来，不必手输）→ 起一轮调试线录制会话；
  2. 在放行**之前**下好 --break 给的断点（放行之后那一段已经跑过去了）；
  3. 过指纹闸 → 重放（与 rec replay 同一套：白名单 + AUI 安全闸）；
  4. 等停站，报出 file:line 与原因。

停站之后**不收拾会话** —— 接着用 tt debug exec "print …" 取证，收工是：
  tt debug quit ; tt fep session close

没停站（断点没命中）时退 1 并说清是超时还是程序已经结束 —— 那不是"发失败"，是"没捕获到现场"。

已经有一个会话在跑（例如你刚用 rec start 起过）就加 --attached，跳过第 1、2 步。`,
	Args: exactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// ① 录制那一面：身份决定怎么起会话
		t, err := resolveRecording(uiFrom, args[0])
		if err != nil {
			return err
		}
		// r.r 录的件不能在这里重放（azzp000 不带 -d，编号表与调试线不是同一张）；
		// r.d（T100 原生调试器）录的件**可以**：它那条作业连接也是 `fglrun -d` 起的，
		// 实测编号表与调试线逐条相同，闸门会按那条连接比（它里面另有一条 fgldeb 的连接，在这里没有对应）。
		if t.Path == fep.PathRunnerRR {
			return fail(exitState,
				"这份录制是 %s 录的：要在断点下重放，得用调试线录的那一份（tt fep rec start）。"+
					"r.r 录的表与调试线不是同一张（跨路径实测共同编号 0 条相同），重放会点错控件",
				t.Path.Label())
		}
		if !t.Path.Known() {
			fmt.Fprintln(cmd.ErrOrStderr(), "· 这份录制没有身份文件：不知道它是哪条路起的；下面的指纹闸会对一遍，不符就拒")
		}
		env, prog := firstNonEmptyStr(t.MetaEnv(), repEnv), firstNonEmptyStr(t.Prog, repProg)
		ent := t.MetaEnt()
		if ent == "" {
			ent = strings.TrimSpace(repEnt)
		}
		if env == "" || prog == "" {
			return fail(exitUsage,
				"这份录制没说它是哪个环境/哪个作业（没有身份文件）：用 --env 与 --prog 指明")
		}

		// ② 先把会话起起来（断点在放行前下好）
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		var started *fep.Runtime
		if !repAttached {
			rt, fgl, oerr := openDebugRecording(cmd, recSession{
				Ctx: ctx, Env: env, Prog: prog, Ent: ent, Module: repModule,
				Breaks: repBreak, Origin: fep.OriginUser,
			})
			if oerr != nil {
				return oerr
			}
			started = rt
			fmt.Fprintf(cmd.ErrOrStderr(), "· 会话已起（录制 ID %s，画面接线 %s）\n", rt.Session.Tag, fgl)
		}

		client, cerr := newClient()
		if cerr != nil {
			if started != nil {
				_ = started.Close()
			}
			return cerr
		}
		// 等界面出来再比闸门：作业刚放行时那一屏还在画
		if repWaitUI > 0 {
			if werr := waitLiveInterface(cmd, repWaitUI); werr != nil {
				return werr
			}
		}

		// ③ 指纹闸 + 重放（与 rec replay 同一套机制）
		rec := t.multi()
		live, ok, lerr := liveMulti(client)
		if lerr != nil {
			return fail(exitFailed, "%v", lerr)
		}
		if !ok {
			return fail(exitNoSession, "没有会话在跑（--attached 时请先起一个）")
		}
		v := fep.GateReplay(rec, live)
		fmt.Fprint(cmd.ErrOrStderr(), v.Report())
		if !v.OK {
			return fail(exitState, "指纹闸不过，没发任何事件：%s", v.Reason)
		}
		events, eerr := t.events()
		if eerr != nil {
			return eerr
		}
		plan := fep.PlanReplay(events, fep.ReplayOptions{StopAt: repStopAt, Skip: repSkip})
		fmt.Fprintf(cmd.ErrOrStderr(), "· 扫到 %d 条事件 → 计划回放 %d 条，跳过 %d 条\n",
			plan.Scanned, len(plan.Events), len(plan.Skips))
		if repInterval == 0 {
			repInterval = 300 * time.Millisecond
		}
		logf := func(format string, a ...any) {
			fmt.Fprintf(cmd.ErrOrStderr(), "· "+format+"\n", a...)
		}
		conn := repConn
		if conn == 0 && repWindow != "" {
			id, cerr := client.ConnForWindow(repWindow)
			if cerr != nil {
				return fail(exitTarget, "%v", cerr)
			}
			conn = id
		}
		res, rerr := client.RunReplay(plan, conn, repInterval, false, logf)
		if rerr != nil {
			return fail(exitFailed, "重放失败：%v", rerr)
		}

		// ④ 等停站，把现场报出来
		spot, timedOut, werr := waitStopped(cmd, repTimeout)
		if werr != nil {
			return fail(exitFailed, "%v", werr)
		}
		out := map[string]any{
			"recording": t.ID, "gate": v, "replay": res, "stop": spot, "timedOut": timedOut,
		}
		if started != nil {
			out["session"] = started.Session.Tag
		}
		if timedOut || spot.File == "" {
			_ = printJSONOrText(out, func() {
				fmt.Printf("重放：发 %d 条 / 跳 %d 条 / 失败 %d 条\n", res.Sent, res.Skipped, res.Failed)
				fmt.Printf("✗ 没捕获到断点现场（等了 %s）：断点没命中，或程序已经结束\n", repTimeout)
				fmt.Println("  下一步：tt debug status 看它现在在哪；tt debug wait --for stopped --timeout 300 接着等")
			})
			return fail(exitFailed, "没捕获到断点现场（断点没命中或程序已结束）")
		}
		return printJSONOrText(out, func() {
			fmt.Printf("重放：发 %d 条 / 跳 %d 条 / 失败 %d 条\n", res.Sent, res.Skipped, res.Failed)
			fmt.Printf("✓ 捕获现场：%s\n", spot)
			fmt.Println("  取证：tt debug exec \"print <变量>\" \"where\"")
			fmt.Println("  收工：tt debug quit ; tt fep session close")
		})
	},
}

// waitLiveInterface 等当前会话的界面出来（作业刚放行时那一屏还在画）。
//
// 等的是**录制件里的对象**，不是屏幕上有没有窗口：协议是唯一权威，
// 而闸门要比的正是这张编号表 —— 表还是空的就去比，只会得到一句"还没有界面"。
func waitLiveInterface(cmd *cobra.Command, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		client, cerr := newClient()
		if cerr != nil {
			return cerr
		}
		if live, ok, lerr := liveMulti(client); lerr == nil && ok && live.NonEmpty() {
			return nil
		}
		if time.Now().After(deadline) {
			return fail(exitState, "等了 %s 界面还没出来（作业没连上代理？看 tt fep session ps 与代理端口）", timeout)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "· 等界面出来…\n")
		time.Sleep(3 * time.Second)
	}
}

func firstNonEmptyStr(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// MetaEnv / MetaEnt 是 recTarget 上的两个小取值器（身份里没有就是空）。
func (t recTarget) MetaEnv() string {
	if t.Meta == nil {
		return ""
	}
	return t.Meta.Env
}

func (t recTarget) MetaEnt() string {
	if t.Meta == nil {
		return ""
	}
	return t.Meta.Ent
}

func init() {
	recReproduceCmd.Flags().StringArrayVar(&repBreak, "break", nil,
		"放行前先下的断点（`file:line` 或函数名，可重复）。例：--break asf_bsft001_wf.4gl:13640")
	recReproduceCmd.Flags().DurationVar(&repTimeout, "timeout", 120*time.Second, "重放完等停站多久")
	recReproduceCmd.Flags().DurationVar(&repWaitUI, "wait-ui", 90*time.Second, "等作业界面出来多久（界面没出来就没法比指纹闸）")
	recReproduceCmd.Flags().BoolVar(&repAttached, "attached", false, "用已经在跑的会话，不自己起（跳过起会话与下断点）")
	recReproduceCmd.Flags().StringVar(&repEnv, "env", "", "环境名（身份里没有时才需要）")
	recReproduceCmd.Flags().StringVar(&repProg, "prog", "", "作业程序名（身份里没有时才需要）")
	recReproduceCmd.Flags().StringVar(&repEnt, "ent", "", "企业编号（身份里没有时才需要）")
	recReproduceCmd.Flags().StringVarP(&repModule, "module", "m", "asf", "T100 模块目录名")
	recReproduceCmd.Flags().IntVar(&repConn, "conn", 0, "把事件发到哪条连接（0 = 自动）")
	recReproduceCmd.Flags().StringVar(&repWindow, "window", "", "按窗口名路由连接")
	recReproduceCmd.Flags().IntVar(&repStopAt, "stop-at", 0, "只回放到第 N 条（含）")
	recReproduceCmd.Flags().StringArrayVar(&repSkip, "skip", nil, "命中这段子串的事件就跳过（可重复）")
	recReproduceCmd.Flags().DurationVar(&repInterval, "interval", 300*time.Millisecond, "每条之间等多久")
	recCmd.AddCommand(recReproduceCmd)
}
