// Package dict 承载 tt dict 命令组（原 tdict 的 CLI）。
//
// 合并前的 tdict 是独立二进制、独立 cli 包；现在并入 tt 后以子命令组形式存在，
// 数据源实现在 internal/dict（db 本地镜像 / live 远程直查 / dbsync 同步）。
package dict

import (
	"github.com/spf13/cobra"

	"tt/internal/cli/common"
)

// Group 是 tt dict 命令组。各命令文件在自己的 init() 里往它挂子命令。
var Group = &cobra.Command{
	Use:   "dict",
	Short: "TDict - ERP 数据字典查询工具",
	Long: `查询 ERP 数据字典。

查询数据源（r.t/r.v/desc/scc/r.q）默认直查默认环境（config.json hosts.activeEnv）的远程库
（金仓/Oracle）；一个环境都没配时才用本地 SQLite 镜像（erp_data.db，由 tt dict db sync 同步）。
可用 config.json 顶层 query.source 固定数据源，或 --env <环境名|local> 覆盖本次调用。`,

	// RunE：**收到不认识的子命令必须非 0**（见 common.UnknownSubcommand）。
	// 没有这个 RunE，cobra 会打一份帮助就退 0 —— 而"我打错了命令"读成成功，
	// 是调用方最难自己发现的失败形态（实测 2026-09-25：`tt dict nope` 退 0，
	// 与它自己文档里写的「1 = 用法错」相矛盾）。
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return common.UnknownSubcommand(cmd, args)
	},
}

// Register 把本命令组挂到根命令上。
func Register(root *cobra.Command) { root.AddCommand(Group) }
