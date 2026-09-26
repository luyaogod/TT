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
const devLong = `T100 设计器包工具：.tzc 代码包 + .tzs 表单包

在 tt 里，原来 tdev 的命令整体后移一级：
  tdev tzc export …   →  tt dev tzc export …
  tdev tzs export …   →  tt dev tzs export …
  tdev install …      →  tt install …（安装不再住在这条线上）

.tzc 的退出码与参数写法完全不变（0 成功 / 2 包格式或用法错 / 3 校验失败 /
4 拒绝写入 / 5 IO 与环境失败）。.tzs 那条线**没有 3**，另有一个 1（引擎内部错）。` + "\n\n" + devcli.Usage

// runDev 把参数原样交给 tdev 自己的解析器，并把它的退出码透传给根命令。
//
// tdev 的 flag 解析是位置无关的（-o/--json 可以出现在位置参数之后），
// 标准库 flag 做不到这一点，所以这里必须 DisableFlagParsing，
// 不能交给 cobra 解析。
func runDev(cmd *cobra.Command, args []string) error {
	var err error
	args, err = takeRootFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitCode(2)
	}
	if code := devcli.Run(args); code != 0 {
		return exitCode(code)
	}
	return nil
}

// takeRootFlags 把根命令的常驻开关从转发给 tdev 的参数里摘出来，翻译成这条线认识的东西。
//
// 为什么需要这一步：本命令组是 DisableFlagParsing，cobra 不会替它解析任何 flag，
// 所以 `tt --config X dev tzc export …` 和 `tt dev tzc export … --config X` 里的
// --config 都会原样落进 tdev 自己的解析器，而它不认识这个 flag，结果是
// 「未知子命令 "--config"」直接退出 2。摘掉它之后两种写法都能用。
//
// `--json` / `--csv` / `--format` 是同一回事（实测 2026-09-25）：根命令的 --help 把它们
// 写作全局开关，而 `tt dev tzc selftest --format csv` 与 `tt --format csv dev tzc selftest`
// **两种位置**都只得到一句「未知子命令 "--format"」—— 位置不同、结果一样糟。
//
// `--config` 走环境变量而不是直接把值传给 internal/dev：那边只经 config.ResolvePath 读配置，
// 而 ResolvePath 本就认 TT_CONFIG —— 不必为这一个 flag 给 TDev 侧开一条新接口。
//
// 翻译表见 devFormat：json → --json（这条线只有这一个机器可读开关）；table/text → 丢掉
// （缺省就是人读文本）；csv → **明确拒绝**（这条线没有 CSV，而"要 CSV 拿到 JSON"正是最该
// 避免的静默走偏）。`--env/--conn` 同理拒绝并点名它属于哪条线。
func takeRootFlags(args []string) ([]string, error) {
	out := make([]string, 0, len(args))
	// 翻译出来的开关**攒着挂到最后**，不能就地放进 out：`tt --json dev tzc selftest` 里
	// 那个 --json 本来就在最前面，就地保留的话 tdev 会把 `args[0]` 当成子命令名，报
	// 「未知子命令 "--json"」—— 实测（2026-09-25）。tdev 的解析器位置无关，但"第一个
	// 位置参数是子命令"这条它认。
	extra := make([]string, 0, 2)
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--config" && i+1 < len(args):
			_ = os.Setenv("TT_CONFIG", args[i+1])
			i++
		case strings.HasPrefix(a, "--config="):
			_ = os.Setenv("TT_CONFIG", strings.TrimPrefix(a, "--config="))
		case a == "--json":
			extra = append(extra, "--json")
		case a == "--csv":
			return nil, devFormatErr("csv")
		case a == "--env" || a == "--conn" || strings.HasPrefix(a, "--env=") || strings.HasPrefix(a, "--conn="):
			return nil, fmt.Errorf("%s 是 tt dict / tt debug 那条线的开关（选哪个库/哪个环境）；"+
				"`tt dev` 这条线用 --workspace 或配置里的 tzs.workspace 定位", strings.SplitN(a, "=", 2)[0])
		case a == "--format" && i+1 < len(args):
			v, err := devFormat(args[i+1])
			if err != nil {
				return nil, err
			}
			i++
			if v != "" {
				extra = append(extra, v)
			}
		case strings.HasPrefix(a, "--format="):
			v, err := devFormat(strings.TrimPrefix(a, "--format="))
			if err != nil {
				return nil, err
			}
			if v != "" {
				extra = append(extra, v)
			}
		default:
			out = append(out, a)
		}
	}
	return append(out, extra...), nil
}

// devFormat 把根命令的 --format 取值翻译成这条线认识的东西。返回空串 = 这个取值在本线
// 就是"没有开关"（人读文本是缺省）。
func devFormat(v string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "json":
		return "--json", nil
	case "table", "text":
		return "", nil
	case "csv":
		return "", devFormatErr("csv")
	default:
		return "", fmt.Errorf("--format 不认识 %q：可用 json / csv / table（`tt dev` 这条线只支持 json 与 table）", v)
	}
}

func devFormatErr(v string) error {
	return fmt.Errorf("`tt dev` 这条线没有 %s 输出：要机器可读加 --json，要看人读文本去掉这个开关", strings.ToUpper(v))
}

// exitCode 是带退出码的错误，由根命令的 exitCodeOf 解出。
type exitCode int

func (e exitCode) Error() string { return fmt.Sprintf("退出码 %d", int(e)) }

// ExitCode 实现根命令期望的退出码接口。
func (e exitCode) ExitCode() int { return int(e) }

// AlreadyReported 告诉根命令「这次失败我已经打出去了」，让它别再抄一份。
//
// 为什么：`tt dev` 是一套完整的 CLI —— 自己的 Usage、自己的退出码、自己把失败打成
// stdout 的信封（或 .tzs 那条线的帧）。失败路径返回前**都**打过一份（`fail()` 或各自的
// 文案），根命令的 reportError 再打一次的后果是 **stdout 上出现两个 JSON 对象**：
// `jq .error.code` 会打出两行，其中一行还是 null。
//
// 记号由命令组自己声明，根命令只问"你报过了吗"（见 internal/cli/root.go 的
// alreadyReported），所以两边不必共享一个类型。
func (e exitCode) AlreadyReported() bool { return true }
