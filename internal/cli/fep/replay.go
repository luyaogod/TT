package fep

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var (
	replayConn     int
	replayWindow   string
	replayStopAt   int
	replaySkip     []string
	replayInterval time.Duration
	replayDryRun   bool
)

var recReplayCmd = &cobra.Command{
	Use:   "replay",
	Short: "回放一轮录制下来的动作（白名单 + AUI 安全闸 + 可停在第 N 步）",
	Long: `把录制件里真人做过的一轮动作重新发一遍。

与 rec steps 不是一回事：steps 给人看（丢导航、只要语义），replay 求保真（多带状态）。
白名单的边界是实测出来的 —— ConfigureEvent{currentRow/currentColumn}（前端回抱“当前行/列”）
**必须**回放（它就是“选中哪一行”），tabIndex/bufferSize/pageSize 才该丢。

三道保险：
  1. 白名单外的一律不发；
  2. AUI 安全闸：对象不在**当次界面**里就跳过（每一步都重新读界面，不用旧模型）；
  3. --dry-run 只出计划一个字节都不发；--stop-at 让人先跑一半看效果。

每一步都写审计。执行前会先把计划（要发什么、跳了什么、为什么）打出来 —— 先看得见再动手。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		events, err := loadEvents(uiFrom)
		if err != nil {
			return err
		}
		// 计划阶段全是离线的（读录制件），所以 --dry-run 不需要有会话 ——
		// “先看清楚要发什么”本来就应该在开会话之前做。
		client, cerr := newClient()
		if cerr != nil && !replayDryRun {
			return cerr
		}
		// 计划阶段**不做** AUI 判断：那会把“后面才出现的按钮”（确定/审核/是）提前剔掉，
		// 实测导致保存与审核那几步根本没发。AUI 安全闸只放在发之前那一刻（RunReplay 里逐条判）。
		opt := fep.ReplayOptions{StopAt: replayStopAt, Skip: replaySkip}
		plan := fep.PlanReplay(events, opt)
		fmt.Fprintf(cmd.ErrOrStderr(), "· 扫到 %d 条事件 → 计划回放 %d 条，跳过 %d 条\n",
			plan.Scanned, len(plan.Events), len(plan.Skips))
		for _, s := range plan.Skips {
			fmt.Fprintf(cmd.ErrOrStderr(), "· 跳过：%s（%s）\n", truncate(s.Raw, 60), s.Reason)
		}
		if replayDryRun {
			type out struct {
				DryRun  bool              `json:"dryRun"`
				Planned int               `json:"planned"`
				Skipped int               `json:"skipped"`
				Events  []fep.ReplayEvent `json:"events"`
			}
			res := out{DryRun: true, Planned: len(plan.Events), Skipped: len(plan.Skips), Events: plan.Events}
			return printJSONOrText(res, func() {
				for _, ev := range plan.Events {
					fmt.Printf("%3d. %s\n", ev.N, ev.Event)
				}
				fmt.Printf("（dry-run：以上 %d 条会被发出，一个字节都没发）\n", len(plan.Events))
			})
		}
		if client == nil {
			if client, cerr = newClient(); cerr != nil {
				return cerr
			}
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
		res, rerr := client.RunReplay(plan, conn, replayInterval, false, logf)
		entry := fep.AuditEntry{Verb: "replay", Result: "done"}
		if res != nil {
			entry.Conn = conn
			entry.Result = fmt.Sprintf("sent=%d skipped=%d failed=%d", res.Sent, res.Skipped, res.Failed)
		}
		if rerr != nil {
			entry.Result = "failed: " + rerr.Error()
		}
		_ = fep.AppendAudit(filepath.Join(recDirForAudit(nil), "audit.log"), entry)
		if rerr != nil {
			return fail(exitFailed, "%v", rerr)
		}
		return printJSONOrText(res, func() {
			fmt.Printf("回放：发 %d 条 / 跳 %d 条 / 失败 %d 条；运行期 +%d 字节\n",
				res.Sent, res.Skipped, res.Failed, res.RuntimeBytes)
		})
	},
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func init() {
	recReplayCmd.Flags().StringVar(&uiFrom, "from", "", "录制件：单个 .log 或存放它的目录（默认 fep 缓存目录）")
	recReplayCmd.Flags().IntVar(&replayConn, "conn", 0, "连接号（多窗口作业必给；给了 --window 就按窗口路由）")
	recReplayCmd.Flags().StringVar(&replayWindow, "window", "", "按窗口名路由连接")
	recReplayCmd.Flags().IntVar(&replayStopAt, "stop-at", 0, "只回放到第 N 条（含），用于先跑一半看效果")
	recReplayCmd.Flags().StringArrayVar(&replaySkip, "skip", nil, "命中这段子串的事件就跳过（可重复）")
	recReplayCmd.Flags().DurationVar(&replayInterval, "interval", 300*time.Millisecond, "每条之间等多久（太快会让上一屏还没建好）")
	recReplayCmd.Flags().BoolVar(&replayDryRun, "dry-run", false, "只出计划，一个字节都不发")
	recCmd.AddCommand(recReplayCmd)
}
