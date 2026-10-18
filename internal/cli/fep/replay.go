package fep

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/debug"
	"tt/internal/fep"
)

var (
	replayConn     int
	replayWindow   string
	replayStopAt   int
	replaySkip     []string
	replayInterval time.Duration
	replayDryRun   bool
	replayNoGate   bool
)

var recReplayCmd = &cobra.Command{
	Use:   "replay",
	Short: "回放一轮录制下来的动作（指纹闸 + 白名单 + AUI 安全闸 + 可停在第 N 步）",
	Long: `把录制件里真人做过的一轮动作重新发一遍。

与 rec steps 不是一回事：steps 给人看（丢导航、只要语义），replay 求保真（多带状态）。
白名单的边界是实测出来的 —— ConfigureEvent{currentRow/currentColumn}（前端回抱“当前行/列”）
**必须**回放（它就是“选中哪一行”），tabIndex/bufferSize/pageSize 才该丢。

四道保险：
  1. 指纹闸：录制件的编号表必须和**当前界面**对得上（实测跨启动路径共同编号 0 条相同，
     而旧的安全闸拦不住它 —— 6/6 会安静地点到别的控件上）；
  2. 白名单外的一律不发；
  3. AUI 安全闸：对象不在**当次界面**里就跳过（每一步都重新读界面，不用旧模型）；
  4. --dry-run 只出计划一个字节都不发；--stop-at 让人先跑一半看效果。

每一步都写审计。执行前会先把闸门判决与计划（要发什么、跳了什么、为什么）打出来 —— 先看得见再动手。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 回放是**平台**在发事件（既不是人也不是 AI 的自由操作）：不声明就是把 CLI 默认的
		// `ai` 带进模式/用途闸门里。实测（2026-10-18 16:08）：录制会话的用途是 record，
		// 放行那一步被"纯录制会话不提供编辑器"整条拒掉 —— 回放一个字节都没发出去。
		// 与 `tt recording start` 同一条口径：平台驱动的动作声明成 system。
		if strings.TrimSpace(os.Getenv(debug.ActorEnv)) == "" {
			_ = os.Setenv(debug.ActorEnv, debug.ActorSystem)
		}		// ① 录制件那一面（离线，读文件）
		t, terr := resolveRecording(uiFrom, recID)
		if terr != nil {
			return terr
		}
		rec := t.multi()
		events, err := t.events()
		if err != nil {
			return err
		}
		// 计划阶段全是离线的（读录制件），所以 --dry-run 不需要有会话 ——
		// “先看清楚要发什么”本来就应该在开会话之前做。
		client, cerr := newClient()
		if cerr != nil && !replayDryRun {
			return cerr
		}
		// ② 指纹闸：先比"这份录制配不配得上现在这个界面"，再谈发什么
		if !replayNoGate {
			live, ok, lerr := liveMulti(client)
			if lerr != nil {
				return fail(exitFailed, "%v", lerr)
			}
			if !ok {
				if replayDryRun {
					fmt.Fprintln(cmd.ErrOrStderr(), "· 没有会话在跑：闸门跳过（--dry-run 只出计划）")
				} else {
					return fail(exitNoSession, "没有登记中的会话，指纹闸无从比对（先跑 tt recording start --launch standard）")
				}
			} else {
				v := fep.GateReplayInitial(rec, live)
				fmt.Fprint(cmd.ErrOrStderr(), v.Report())
				if !v.OK {
					if replayDryRun {
						fmt.Fprintln(cmd.ErrOrStderr(), "· --dry-run：只出计划，不拦")
					} else {
						code := exitState
						if v.Code == fep.GateNoLive {
							code = exitNoSession
						}
						return fail(code, "指纹闸不过：%s", v.Reason)
					}
				}
			}
		}
		// 计划阶段**不做** AUI 判断：那会把“后面才出现的按钮”（确定/审核/是）提前剔掉，
		// 实测导致保存与审核那几步根本没发。AUI 安全闸只放在发之前那一刻（RunReplay 里逐条判）。
		//
		// 但模型还是要**递进去**：它只用来**记身份**（`colName`/`fieldId`，契约 C3/A18）——
		// 有了它，回放时每一步都能按身份重新定位，而不是把录制时的裸 id 再发一遍。
		// 读不到模型 ⇒ 传 nil ⇒ 旧行为一字不变（见 planReplayFor）。
		plan := planReplayFor(t, events, replayStopAt, replaySkip)
		fmt.Fprintf(cmd.ErrOrStderr(), "· 扫到 %d 条事件 → 计划回放 %d 条，跳过 %d 条\n",
			plan.Scanned, len(plan.Events), len(plan.Skips))
		for _, s := range plan.Skips {
			fmt.Fprintf(cmd.ErrOrStderr(), "· 跳过：%s（%s）\n", truncate(s.Raw, 60), s.Reason)
		}
		if replayDryRun {
			// A23 的**例外**：不发包的路径放行，但每条无身份的步骤必须标出来 ——
			// 展示得出来 ≠ 跑得动，这句话要写在输出里（文本与 json 都要，默认是 json）。
			return printJSONOrText(replayDryRunReport(plan), func() {
				fmt.Print(replayDryRunText(plan))
			})
		}
		if client == nil {
			if client, cerr = newClient(); cerr != nil {
				return cerr
			}
		}

		// 降级规模先报出来（用户 2026-10-18 定：没有正式身份不再整份拒绝，而是照发 + 标注）：
		// 真实操作必含工具栏/表格点击，它们天生没有字段身份；这里只**说清楚**有多少步是降级。
		if n := fep.CountUnidentified(plan); n > 0 {
			fmt.Fprintf(cmd.ErrOrStderr(),
				"· 其中 %d 步没有正式身份：将按**录制时的临时编号**发送（日志逐条标 degraded）—— "+
					"现场与录制时不同的话，这些步可能点到别的控件\n", n)
		}

		conn := replayConn
		if conn == 0 && replayWindow != "" {
			id, cerr := client.ConnForWindow(replayWindow)
			if cerr != nil {
				return fail(exitTarget, "%v", cerr)
			}
			conn = id
		}
		if replayInterval == 0 {
			replayInterval = 300 * time.Millisecond
		}
		logf := func(format string, a ...any) {
			fmt.Fprintf(cmd.ErrOrStderr(), "· "+format+"\n", a...)
		}
		res, rerr := client.RunReplayIdentity(plan, fep.ReplayRunOptions{
			Conn: conn, Settle: replayInterval, Logf: logf,
		})
		entry := fep.AuditEntry{Verb: "replay", Result: "done"}
		if res != nil {
			entry.Conn = conn
			entry.Result = fmt.Sprintf("sent=%d skipped=%d failed=%d degraded=%d", res.Sent, res.Skipped, res.Failed, res.Degraded)
		}
		if rerr != nil {
			entry.Result = "failed: " + rerr.Error()
		}
		_ = fep.AppendAudit(filepath.Join(recDirForAudit(nil), "audit.log"), entry)
		if rerr != nil {
			return replayGateErr(cmd, rerr)
		}
		return printJSONOrText(res, func() {
			fmt.Printf("回放：发 %d 条 / 跳 %d 条 / 失败 %d 条；运行期 +%d 字节\n",
				res.Sent, res.Skipped, res.Failed, res.RuntimeBytes)
			if res.Degraded > 0 {
				fmt.Printf("  其中 %d 条是**降级发送**（没有正式身份，按录制时的临时编号）：现场与录制时不同的话，这些步可能点到别的控件\n", res.Degraded)
			}
		})
	},
}

// replayGateErr 把一次 `*fep.Refusal`（身份解析不过/有歧义）翻成 CLI 的失败信封：
// 这类拒绝仍然**逐条生效**（那才是"会点错"的场合），所以入口没变。
func replayGateErr(cmd *cobra.Command, err error) error {
	var ref *fep.Refusal
	if errors.As(err, &ref) {
		if ref.Hint != "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "· 怎么办：%s\n", ref.Hint)
		}
		return fail(ref.Exit, "%s（%s）", ref.Message, ref.Code)
	}
	return fail(exitFailed, "%v", err)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// ── 身份接线（契约 C3/A18）：把录制的界面模型递给计划 ────────────────────────
//
// 两件事分开写，是为了各自可测：`replayOptionsFor` 是纯接缝（给不给模型），
// `replayIdentityModel` 是读侧（能不能读出模型）。命令行为不因这次接线而改变。

// replayOptionsFor 把命令行开关 + 录制模型折成 `fep.ReplayOptions`。
//
// 接缝上只有一条要钉死的判据：**给了模型 ⇒ `Model` 非 nil；没给 ⇒ nil**，
// 且 `StopAt`/`Skip` 原样带过去（TestReplayOptionsForWiresTheModel）。
func replayOptionsFor(stopAt int, skip []string, m *fep.Model) fep.ReplayOptions {
	return fep.ReplayOptions{StopAt: stopAt, Skip: skip, Model: m}
}

// planReplayFor 是**两个回放入口（`rec replay` 与 `rec reproduce`）共用的一条**组装路径：
// 命令行开关 + **这份录制自己的**界面模型 → 回放计划。
//
// 为什么抽成函数而不是在两处各写一行：接线最容易被漏在"另一个入口"上（这次就是这么漏的）。
// 只留一条路径，测试对着它断言，两个命令就都不会掉队
// （TestReplayPlanKeepsLegacyShapeWithoutModel / TestReplayCommandPathSendsToResolvedIdentity）。
func planReplayFor(t recTarget, events []string, stopAt int, skip []string) fep.ReplayPlan {
	return fep.PlanReplay(events, replayOptionsFor(stopAt, skip, replayIdentityModel(t)))
}

// replayIdentityModel 把一份录制的**界面侧**读成一个模型，给 `PlanReplay` 记身份用。
//
// 为什么必须是**这份录制自己的**模型：计划里那些 `idRef` 是录制当时发出去的裸 id，
// 只有录制自己的 AUI 才能把它们翻成 `(colName, fieldId)`。拿"当前界面"的模型去翻，
// 翻的是另一张编号表（跨启动路径实测共同编号 0 条相同 —— 见 internal/fep/README 的指纹闸那节）。
//
// `recTarget.Files` 已经是**界面侧**（`-rt2gdc.log`）；这里仍逐个核后缀：事件侧的载荷
// 也能被 `ReadRecording` 解成文本，混进模型里就成了"凭空多出来的对象"。
//
// 读不到（没有界面侧 / 文件读不动 / 一条 AUI 都没有）⇒ 返回 **nil**：
// 身份链自动不参与，回放照旧按 `idRef` 走 —— **旧行为一字不变，绝不因此报错**
// （TestReplayIdentityModelIsNilWithoutAUISide / TestReplayPlanKeepsLegacyShapeWithoutModel）。
func replayIdentityModel(t recTarget) *fep.Model {
	var files []string
	for _, f := range t.Files {
		if strings.HasSuffix(f, "-rt2gdc.log") {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return nil
	}
	m := fep.NewModel()
	msgs := 0
	for _, f := range files {
		bodies, err := fep.ReadRecording(f)
		if err != nil {
			return nil
		}
		for _, body := range bodies {
			m.ApplyMessage(body)
			msgs++
		}
	}
	if msgs == 0 {
		return nil
	}
	// 窗口归属只向下传播一层（与 `loadModel` 同一口径）；身份判定不靠它，但补上不吃亏。
	m.Propagate()
	return m
}

// ── A23 的两个出口（契约 §7）：拒绝要带码，展示要带标注 ───────────────────────

// replayGateErr 把身份闸/回放里的 `*fep.Refusal` 翻成带退出码的失败。
//
// 为什么不能直接 `fail(exitFailed, …)`：契约 C4.2/A11 的四码各有退出码（5/5/1/4），
// 一律折成 1 就等于把"没这个字段"与"网络坏了"说成同一件事（自动化排障要能分开）。
// Hint 走 stderr：它是给人看的下一步，不该混进 json 信封的 message 里。
// dryRunStep 是干跑输出里的一步：`ReplayEvent` 原样平铺，只**多**一个 note。
//
// 为什么用嵌入而不是改 `fep.ReplayEvent`：有身份的步骤不多出任何键 ——
// 既有读者（脚本、面板）一个字都不用改；没有身份的那些多一句 `note`。
type dryRunStep struct {
	fep.ReplayEvent
	Note string `json:"note,omitempty"`
}

// replayDryRunReport 是 `--dry-run` 的 json 形态（默认输出就是 json，所以标注必须在）。
func replayDryRunReport(plan fep.ReplayPlan) any {
	noIdent := map[int]bool{}
	for _, n := range fep.NoIdentitySteps(plan) {
		noIdent[n] = true
	}
	steps := make([]dryRunStep, 0, len(plan.Events))
	for _, ev := range plan.Events {
		s := dryRunStep{ReplayEvent: ev}
		if noIdent[ev.N] {
			s.Note = fep.NoIdentityNote
		}
		steps = append(steps, s)
	}
	return struct {
		DryRun  bool         `json:"dryRun"`
		Planned int          `json:"planned"`
		Skipped int          `json:"skipped"`
		Events  []dryRunStep `json:"events"`
	}{DryRun: true, Planned: len(plan.Events), Skipped: len(plan.Skips), Events: steps}
}

// replayDryRunText 是 `--dry-run` 的人读形态：每条无身份的步骤后面跟一句标注。
//
// 抽成函数只为可测（命令直接写 os.Stdout）：标注是 A23 的硬要求，
// 而"要求写在输出里"这件事只有断言输出才守得住。
func replayDryRunText(plan fep.ReplayPlan) string {
	noIdent := map[int]bool{}
	for _, n := range fep.NoIdentitySteps(plan) {
		noIdent[n] = true
	}
	var b strings.Builder
	for _, ev := range plan.Events {
		fmt.Fprintf(&b, "%3d. %s", ev.N, ev.Event)
		if noIdent[ev.N] {
			fmt.Fprintf(&b, "   ← %s", fep.NoIdentityNote)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "（dry-run：以上 %d 条会被发出，一个字节都没发）\n", len(plan.Events))
	return b.String()
}

func init() {
	recReplayCmd.Flags().StringVar(&uiFrom, "from", "", "录制件：单个 .log 或存放它的目录（默认 fep 缓存目录）")
	recReplayCmd.Flags().StringVar(&recID, "id", "", "录制 ID：去默认落点找这一份（与 --from 二选一）")
	recReplayCmd.Flags().IntVar(&replayConn, "conn", 0, "连接号（多窗口作业必给；给了 --window 就按窗口路由）")
	recReplayCmd.Flags().StringVar(&replayWindow, "window", "", "按窗口名路由连接")
	recReplayCmd.Flags().IntVar(&replayStopAt, "stop-at", 0, "只回放到第 N 条（含），用于先跑一半看效果")
	recReplayCmd.Flags().StringArrayVar(&replaySkip, "skip", nil, "命中这段子串的事件就跳过（可重复）")
	recReplayCmd.Flags().DurationVar(&replayInterval, "interval", 300*time.Millisecond, "每条之间等多久（太快会让上一屏还没建好）")
	recReplayCmd.Flags().BoolVar(&replayDryRun, "dry-run", false, "只出计划，一个字节都不发")
	recReplayCmd.Flags().BoolVar(&replayNoGate, "no-gate", false,
		"跳过指纹闸（只给「我知道自己在做什么」的排障用：跨启动路径回放会安静地点到别的控件上）")
	recCmd.AddCommand(recReplayCmd)
}
