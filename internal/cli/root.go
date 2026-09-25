// Package cli 装配 tt 的根命令。
//
// 三个工具以子命令组并存，各自实现在独立子包里，互不 import：
//
//	tt debug …   原 tdebug（AI 人机协同调试）
//	tt dev …     原 tdev（设计器包安全编辑）
//	tt dict …    原 tdict（ERP 数据字典查询）
//
// 共享的 CLI 上下文在 internal/cli/common（叶子包，避免循环依赖）。
package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/cli/debug"
	"tt/internal/cli/dev"
	"tt/internal/cli/dict"
	"tt/internal/output"
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "tt",
	Short: "TT - T100 工具集（调试 / 设计器包 / 数据字典）",
	Long: `TT 是 TDebug、TDev、TDictCli 三个工具合并后的统一入口。

同一个二进制、同一份配置（config.json）、同一个本地 Web 服务：

  tt debug …   作业调试器：SSH 驱动 fglrun -d 的 (fgldb) 文本调试协议，
               提供本地 Web 调试界面（源码/断点/调用栈/变量/接口日志）与命令行控制端
  tt dev tzc … T100 设计器代码包（.tzc/.tzf/.tzx）：export 渲染带围栏的 4GL 工作区，
               apply 走三道闸门写回
  tt dev tzs … T100 设计器表单包（.tzs/.tzv）：export 纯解压只读；读写表单走具名动词
               （tt dev tzs <动词> --args '<JSON 对象>'），由设计器自己的引擎算，不是拼 XML
  tt dict …    ERP 数据字典查询：r.t / desc / scc / r.q / prog 等，支持本地镜像与远程直查

  tt env …     环境管理：列出/查看/切换 SSH 环境（三个工具共用同一份环境清单）
  tt config …  配置管理：位置/查看/修改/迁移/校验
  tt serve     启动本地 Web 配置服务（调试工作台，含覆盖所有命令组的统一设置页）
  tt install   安装 AI skills 到当前目录，或把 tt 加进用户 PATH

所有输出使用简体中文 (zh_CN)。

退出码 —— **同一个数字在各家含义不同，读哪条线就以那条线为准**：

  tt debug      0 成功 / 1 用法或运行失败（这条线未定义更细的码）
  tt dev tzc    0 成功 / 2 包格式或用法错 / 3 校验失败 / 4 拒绝写入 / 5 IO·环境失败
  tt dev tzs    0 成功 / 1 引擎内部错 / 2 参数或环境错 / 4 设计器拒绝 / 5 传输或环境失败
                （**没有 3**；2 也包括"manifest 拉不到"）
  tt dict       0 成功（含"查无结果"）/ 1 用法错 / 2 数据源错 / 3 缺表
  顶层          0 成功 / 1 用法错（未知命令、未知开关）

3 在 tzc 是"校验失败"、在 dict 是"缺表"、在 tzs 根本不用；4 只在 tzs 有（设计器拒绝）。
包装脚本要按子命令辨认，别只看码。

合并前的命令名仍可直接当子命令组用：tt tdebug … = tt debug …，
tt tdev … = tt dev …，tt tdict … = tt dict …。`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// 旧命令名别名：合并前各自的二进制名，让既有脚本与 AI skill 不改即可用。
var legacyAliases = map[string]string{
	"tdebug": "debug",
	"tdev":   "dev",
	"tdict":  "dict",
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVar(&common.ConfigPath, "config", "", "配置文件路径 (JSON;缺省取统一用户目录 "+common.ConfigHint()+")")
	pf.BoolVar(&common.JSON, "json", false, "以 JSON 输出（--format json 的语法糖，也是默认）")
	pf.BoolVar(&common.CSV, "csv", false, "以 CSV 输出（--format csv 的语法糖）")
	pf.StringVar(&common.Format, "format", "json",
		"输出形态: json(默认,带环境信息的信封) | csv(带 # 环境头) | table(人读表格)")
	pf.BoolVarP(&common.Verbose, "verbose", "v", false, "显示解析细节（如实际使用的数据源路径）")
	pf.StringVar(&common.Env, "env", "", "指定环境名（config.json hosts.sshs 中的 name）")
	// --conn 是合并前 tdict 的叫法，保留为别名，但不出现在 --help 里
	pf.StringVar(&common.Env, "conn", "", "指定环境名（--env 的别名）")
	_ = pf.MarkHidden("conn")

	debug.Register(rootCmd)
	dev.Register(rootCmd)
	dict.Register(rootCmd)

	rootCmd.AddCommand(newEnvCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newCacheCmd())
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newInstallCmd())
	rootCmd.AddCommand(newVersionCmd())
}

// Execute 运行根命令。
// web 承载内嵌的前端构建产物（web/dist）；未构建时可能为 nil 或空目录。
func Execute(web fs.FS) {
	common.WebFS = web
	common.Version = Version

	// 旧命令名兼容：tt tdebug … 等价 tt debug …
	if len(os.Args) > 1 {
		if target, ok := legacyAliases[os.Args[1]]; ok {
			os.Args[1] = target
		}
	}

	if err := rootCmd.Execute(); err != nil {
		// cobra 的 SilenceErrors 让错误只由这里打一次（各子命令只管 return err）。
		// 少了这一行，命令失败就**一声不响**地退出 1 —— 用户看不到原因，
		// 而这恰恰是最需要说清楚的时候（"为什么服务没起来/为什么没输出"）。
		// 带退出码的错误（既有 TDev 的 0/2/3/4/5 契约，以及 output.Error 的
		// 2=数据源 / 3=缺表）按码退出。
		//
		// **例外**：自己已经把失败打出去了的命令组（见 alreadyReported）跳过这一步 ——
		// 它们各有自己完整的输出契约，再抄一份会让 stdout 上出现两个 JSON 对象。
		var reported alreadyReported
		if !errors.As(err, &reported) || !reported.AlreadyReported() {
			reportError(err)
		}
		if code := exitCodeOf(err); code != 0 {
			os.Exit(code)
		}
		os.Exit(1)
	}
}

// alreadyReported 由"自己已经把失败打出去了"的命令组实现 —— 根命令据此跳过 reportError。
//
// 为什么需要它：`tt dev` 是完整的一套 CLI（自己的 Usage、自己的退出码、自己把失败打成
// stdout 的信封或帧），而 reportError 在 JSON 模式下也往 stdout 写信封。两份叠在一起的
// 后果是 **stdout 上恰好两个 JSON 对象**：`jq .error.code` 打出两行，其中一行是 null，
// 而"`--json` 下 stdout 恰好一个对象"正是 .tzs 那条线刚立下的规则。
//
// 用结构化接口而不是共享一个类型：记号由命令组自己声明（`exitCode.AlreadyReported`），
// 根命令只问"你报过了吗" —— 不必为了一个记号把 internal/cli/dev 的类型导出。
type alreadyReported interface{ AlreadyReported() bool }

// reportError 打一次错误。
//
// JSON 模式下错误也走 **stdout 的 JSON 信封**（与 tt dev 既有的 {"ok":false,...}
// 同一条路子），让 agent 靠 Code 分支而不是靠正则啃中文；其余模式保持 stderr 散文，
// 并多打一行「提示:」把修复建议单独拎出来。
//
// ⚠️ 这会改变"stderr 有内容 = 失败"这个既有判据：JSON 模式下失败信息在 stdout。
//
// **`tt dev` 不经过这里**：它自己是一套完整的 CLI（自己的 Usage、自己的退出码、
// 自己把失败打成 stdout 的信封或帧），所以它返回的错误实现了 alreadyReported ——
// 否则同一个失败会在 stdout 上出现两份信封，而 `jq .error.code` 会打出两行（一行是 null）。
func reportError(err error) {
	format := common.OutputFormat()
	var oe *output.Error
	if errors.As(err, &oe) {
		// 错误自己带了环境信息就用它的,否则补上当前数据源的 ——
		// 连不上库时,"连的是哪个环境"比错误文本本身更难猜。
		if oe.Meta.Env == "" && oe.Meta.Source == "" {
			oe.Meta = common.CurrentMeta()
		}
	} else {
		oe = &output.Error{Code: output.CodeUsage, Exit: exitCodeOf(err),
			Message: err.Error(), Meta: common.CurrentMeta()}
	}
	if oe.Exit == 0 {
		oe.Exit = 1
	}
	if format == output.FormatJSON {
		if werr := output.WriteError(os.Stdout, oe); werr == nil {
			return
		}
	}
	fmt.Fprintf(os.Stderr, "错误: %v\n", oe)
	if oe.Hint != "" {
		fmt.Fprintf(os.Stderr, "提示: %s\n", oe.Hint)
	}
}
