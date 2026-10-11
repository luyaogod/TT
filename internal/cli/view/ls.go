package view

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/view"
)

var (
	lsPID   uint32
	lsAll   bool
	lsLimit int
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "列窗口（默认只列 fep 会话里那个隔离实例的窗口）",
	Long: `列出候选窗口。默认 pid 取 fep 会话登记里的隔离实例 —— 抓图最怕的就是
"抓错了窗口"（用户自己的实例与本会话实例混在一起时，图看起来完全正常）。

给 --all 才列全部进程（排查用），给 --pid 指定某个进程。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		pid := lsPID
		if pid == 0 && !lsAll {
			if p, ok := sessionGDCPID(); ok {
				pid = p
			} else {
				return fail(exitUsage, "没有 fep 会话可以推断 pid：给 --pid <pid> 或 --all")
			}
		}
		wins, err := view.ListTimeout(pid, 5*time.Second)
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		picked := view.PickWindows(wins, pid)
		if lsLimit > 0 && len(picked) > lsLimit {
			picked = picked[:lsLimit]
		}
		return printJSONOrText(map[string]any{"pid": pid, "count": len(picked), "windows": picked}, func() {
			for _, w := range picked {
				fmt.Printf("#%-10x pid=%-6d %-14s %-30s %s\n", w.HWND, w.PID, w.Rect, truncate(w.Title, 30), w.Class)
			}
			fmt.Printf("（%d 个候选窗口）\n", len(picked))
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
	lsCmd.Flags().Uint32Var(&lsPID, "pid", 0, "只列这个进程的窗口（默认取 fep 会话的隔离实例）")
	lsCmd.Flags().BoolVar(&lsAll, "all", false, "列全部进程的窗口（排查用）")
	lsCmd.Flags().IntVar(&lsLimit, "limit", 0, "最多列几个（0=不限）")
}
