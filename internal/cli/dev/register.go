// Package dev 承载 tt dev 命令组（原 tdev 的设计器包管线）。
//
// tdev 原本用标准库 flag 做位置无关解析、并有明确的退出码契约
// （0 成功 / 2 包格式或用法错 / 3 校验失败 / 4 拒绝写入 / 5 IO 与环境失败）。
// 这两点都保留：本包只把它接进 cobra，实际派发仍交给 internal/dev/cli 的 Run()，
// 退出码经 exitCode 透传给根命令。
package dev

import "github.com/spf13/cobra"

// Group 是 tt dev 命令组。
var Group = &cobra.Command{
	Use:                "dev",
	Short:              "TDev - T100 设计器包工具",
	Long:               devLong,
	DisableFlagParsing: true, // 参数原样交给 tdev 自己的解析器
	RunE:               runDev,
}

// Register 把本命令组挂到根命令上。
func Register(root *cobra.Command) { root.AddCommand(Group) }
