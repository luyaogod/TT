package fep

// `rec stop` —— 结束一轮录制（新命令面 `tt recording stop` 的实现）。
//
// 为什么需要它：录制不是"跑完就没了"，它占着三个端口、一个 GDC 实例、一份登记。
// 结束的方式决定留下什么 ——
//   - 只杀进程 → 录制件没有身份文件（以后只能靠算指纹），登记还挂着；
//   - 走这一步 → 先请常驻进程自己收（写完帧、落身份）、只杀登记的 pid、最后摘登记。
//
// 与 `tt recording stop <录制ID>` 是**同一份实现**（`closeRegisteredSession`）：两个入口、
// 一条收尾路径。复制一份的后果是"两个入口对同一轮录制留下不同的现场"。

import (
	"strings"

	"github.com/spf13/cobra"
)

// recStopTag 是 `--tag` 的取值（也可用位置参数给录制 ID）。
var recStopTag string

var recStopCmd = &cobra.Command{
	Use:   "stop [录制ID]",
	Short: "停止一轮录制：收掉代理与 GDC、落录制身份、清登记",
	Long: `停止一轮录制。

  tt recording stop fep-20261018-101500
  tt recording stop --tag fep-20261018-101500

收尾顺序（顺序不对会留下东西）：
  1. 请常驻进程自己退（它要把帧写完、把录制身份落盘）；
  2. 只杀登记里记的那些 pid（绝不按名字扫）；
  3. 调试线起的会话，调试侧按 ref 收口（结束本轮，宿主保留）；
  4. 最后摘掉这一条登记（别的登记不动）。`,
	// atMostOneArg：cobra 自带的参数错会让根命令退 1，而这条线的表里 2 = 用法错。
	Args: atMostOneArg,
	RunE: func(cmd *cobra.Command, args []string) error {
		tag := strings.TrimSpace(recStopTag)
		if len(args) == 1 {
			if tag != "" {
				return fail(exitUsage, "录制 ID 给了两次（位置参数与 --tag），二选一")
			}
			tag = strings.TrimSpace(args[0])
		}
		if tag == "" {
			return fail(exitNoSession, "需要录制 ID：tt recording ls 可查")
		}
		return closeRegisteredSession(cmd, tag)
	},
}

func init() {
	recStopCmd.Flags().StringVar(&recStopTag, "tag", "", "录制 ID（也可用位置参数）")
}
