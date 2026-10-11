package fep

import (
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var doPickCmd = &cobra.Command{
	Use:   "pick <栏位名> <选项文本>",
	Short: "在下拉里选一项（候选项是协议对象，选择是语义动作）",
	Long: `按选项文字在一个栏位的下拉里选一项。

真机用 GDC 自带的 AUI Debug Tree 看得很清楚：FormField → ComboBox → Item（每个 Item 带 text）。
所以"弹出的那一下"虽然是前端本地渲染（协议看不见），但**候选项与选择结果都在协议里**：
本命令找到该栏位下的 ComboBox，挑出文字匹配的 Item，再对它发 ActionEvent —— 这就是"选择"这一步。

选项文本可以只写一部分（如 "L88" 或 "签收"）；命中多个会列出来让你写全。`,
	Args: exactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDo(cmd, fep.DoConfig{Kind: "pick", Target: args[0], Value: args[1], CommitKey: commitKeyFor("pick")})
	},
}

func init() {
	doPickCmd.Flags().IntVar(&doConn, "conn", 0, "连接号")
	doPickCmd.Flags().BoolVar(&doDryRun, "dry-run", false, "只校验与打印，一个字节都不发")
	doPickCmd.Flags().StringVar(&doWindow, "window", "", "只在某个窗口名里找目标")
	doPickCmd.Flags().DurationVar(&doWait, "wait", 10*time.Second, "发完等运行期回包多久")
	doCmd.AddCommand(doPickCmd)
}
