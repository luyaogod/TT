// Package recording 是 `tt recording` 命令组：**录制件是主角**的那套命令面。
//
// # 它和 `tt fep rec` 的关系
//
// 同一批动作（列 / 看 / 留 / 删 / 清 / 建索引）有两个入口：
//   - `tt fep rec …` 是旧入口，**暂留**：Web 的录制面板还在起 `tt recording start` 子进程
//     （见 internal/web/recordings.go），删掉它会把面板弄坏；
//   - `tt recording …` 是新入口，名字更直接（录制件不属于"前端协议"那一组）。
//
// 两个入口**共用同一份实现**：这里不复制 RunE，而是把旧命令的 RunE 与标志镜像过来。
// 复制实现是这类迁移最常见的坑 —— 改一处忘一处，两个入口慢慢长成两种行为。
//
// 等 Web 面板改成起 `tt recording start` 之后，旧入口整体删除（P5 的收尾）。
package recording

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"tt/internal/cli/fep"
)

// 新名字：`rec list` → `recording ls`（其余同名）。
//
// 为什么把 list 改成 ls：与 `tt session ls` 一致，"列"在这套命令里一律叫 ls。
var renames = map[string]string{
	"list":      "ls",
	"reproduce": "replay",
}

// Register 把这一组挂到根命令上。
func Register(root *cobra.Command) {
	group := &cobra.Command{
		Use:   "recording",
		Short: "录制：列 / 看 / 留 / 删 / 清 / 建索引",
		Long: `录制件（.raw 帧流 + 索引 + 元数据）的查看与维护。

  tt recording start …       起一轮录制（--launch debug 可调试 / standard 标准启动）
  tt recording stop <录制ID>  停止一轮录制：收掉代理/GDC、落身份、清登记
  tt recording replay <录制ID> 回放一份录制
  tt recording ls            列出录制件
  tt recording show <录制ID>  看一份录制件的详情
  tt recording keep <录制ID>  标记保留（清理时跳过）
  tt recording rm <录制ID>    删掉一份录制件
  tt recording prune          按保留期清理
  tt recording index [录制ID]  重建帧索引（写坏/缺失时用）

  tt recording active         列正在进行的录制（登记里还活着的）
  tt recording mark-stale <录制ID> 把一份录制标成已失效

每个动词都直连本地录制件或调试面；start/replay 会自己起会话，其余动作只碰本地文件。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	for _, src := range fep.RecordingCommands() {
		name := src.Name()
		if n, ok := renames[name]; ok {
			name = n
		}
		group.AddCommand(mirror(src, name))
	}
	root.AddCommand(group)
}

// mirror 复制一条命令的外观（名字/说明/参数校验/RunE/标志），但**不复制实现**。
//
// 标志用 `AddFlag` 逐个镜像：它们指向旧命令里那些包级变量，所以两个入口读写的是同一份状态。
// 这正是"一份实现"的意思 —— 不存在"新入口的标志改了、旧入口没改"的可能。
func mirror(src *cobra.Command, use string) *cobra.Command {
	c := &cobra.Command{
		Use:               use,
		Aliases:           src.Aliases,
		Short:             src.Short,
		Long:              src.Long,
		Example:           src.Example,
		Args:              src.Args,
		ValidArgsFunction: src.ValidArgsFunction,
		RunE:              src.RunE,
	}
	if c.Args == nil {
		c.Args = cobra.ArbitraryArgs
	}
	src.Flags().VisitAll(func(f *pflag.Flag) {
		c.Flags().AddFlag(f)
	})
	return c
}
