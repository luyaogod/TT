// Package drawio 承载 tt drawio 命令组：用 T100 组件库画原型图。
//
// 能力实现在 internal/drawio（形状库的合成、spec → .drawio、.tzs → 原型图）；
// 本包只做命令面 —— 开关、输出、落点、退出码。形状源 embed 在能力层里，
// 本包不读盘、不碰网络。
package drawio

import (
	"github.com/spf13/cobra"

	"tt/internal/cli/common"
)

// Group 是 tt drawio 命令组。
var Group = newGroup()

// newGroup 造一棵**全新的**命令树。
//
// 为什么是工厂而不是 `var Group` + 各文件 init() 挂子命令（dict/debug 那种写法）：
// 路由测试要每次拿一棵干净的树 —— cobra 与 pflag 的命令实例上挂着解析后的状态，
// 同一个实例跑两条用例会串味。internal/cli/dev 的 newDevCmd() 就是为这件事存在的。
func newGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drawio",
		Short: "用 T100 组件库画原型图（.drawio）",
		Long: `画 T100 的原型图。

需求调研完、要跟客户确认方案时用。控件尺寸、标签在左控件在右、字段行高、容器分区，
一律照 T100 设计器的惯例 —— 目的是让这些原型图长得一样，而不是各画各的。

  tt drawio lib …        把形状库导出成 drawio 能加载的文件
  tt drawio compose …    把「排版规格 spec」展开成 .drawio

形状库是**内置**的：不需要先跑一次构建，也没有任何外部文件依赖。`,
		// 收到不认识的子命令必须非 0（见 common.UnknownSubcommand）——
		// 没有这个 RunE，cobra 会打一份帮助就退 0，调用方会把"我打错了命令"读成成功。
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return common.UnknownSubcommand(cmd, args)
		},
	}
	cmd.AddCommand(newLibCmd())
	cmd.AddCommand(newComposeCmd())
	return cmd
}

// Register 把本命令组挂到根命令上。
func Register(root *cobra.Command) { root.AddCommand(Group) }
