package dev

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	devcli "tt/internal/dev/cli"
)

// devLong 是命令组的 --help 文本。参数与退出码契约沿用 tdev 自己的说明
// （internal/dev/cli.Usage），这里只补一句在 tt 里的调用方式。
const devLong = `T100 设计器包工具：.tzc 安全编辑 + .tzs 只读解压

在 tt 里，原来 tdev 的命令整体后移一级：
  tdev tzc export …   →  tt dev tzc export …
  tdev tzs export …   →  tt dev tzs export …
  tdev install path   →  tt dev install path（也可用统一的 tt install）

退出码与参数写法完全不变（0 成功 / 2 包格式或用法错 / 3 校验失败 /
4 拒绝写入 / 5 IO 与环境失败）。` + "\n\n" + devcli.Usage

// runDev 把参数原样交给 tdev 自己的解析器，并把它的退出码透传给根命令。
//
// tdev 的 flag 解析是位置无关的（-o/--json 可以出现在位置参数之后），
// 标准库 flag 做不到这一点，所以这里必须 DisableFlagParsing，
// 不能交给 cobra 解析。
func runDev(cmd *cobra.Command, args []string) error {
	args = takeConfigFlag(args)
	if code := devcli.Run(args); code != 0 {
		return exitCode(code)
	}
	return nil
}

// takeConfigFlag 从转发给 tdev 的参数里摘出 --config，改以 TT_CONFIG 环境变量告知。
//
// 为什么需要这一步：本命令组是 DisableFlagParsing，cobra 不会替它解析任何 flag，
// 所以 `tt --config X dev tzc export …` 和 `tt dev tzc export … --config X` 里的
// --config 都会原样落进 tdev 自己的解析器，而它不认识这个 flag，结果是
// 「未知子命令 "--config"」直接退出 2。摘掉它之后两种写法都能用。
//
// 用环境变量而不是直接把值传给 internal/dev：那边只经 config.ResolvePath 读配置，
// 而 ResolvePath 本就认 TT_CONFIG —— 不必为这一个 flag 给 TDev 侧开一条新接口。
func takeConfigFlag(args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--config" && i+1 < len(args):
			_ = os.Setenv("TT_CONFIG", args[i+1])
			i++
		case strings.HasPrefix(a, "--config="):
			_ = os.Setenv("TT_CONFIG", strings.TrimPrefix(a, "--config="))
		default:
			out = append(out, a)
		}
	}
	return out
}

// exitCode 是带退出码的错误，由根命令的 exitCodeOf 解出。
type exitCode int

func (e exitCode) Error() string { return fmt.Sprintf("退出码 %d", int(e)) }

// ExitCode 实现根命令期望的退出码接口。
func (e exitCode) ExitCode() int { return int(e) }
