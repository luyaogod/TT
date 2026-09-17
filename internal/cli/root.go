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
	"io/fs"
	"os"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/cli/debug"
	"tt/internal/cli/dev"
	"tt/internal/cli/dict"
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "tt",
	Short: "TT - T100 工具集（调试 / 设计器包 / 数据字典）",
	Long: `TT 是 TDebug、TDev、TDictCli 三个工具合并后的统一入口。

同一个二进制、同一份配置（config.json）、同一个本地 Web 服务：

  tt debug …   作业调试器：SSH 驱动 fglrun -d 的 (fgldb) 文本调试协议，
               提供本地 Web 调试界面（源码/断点/调用栈/变量/接口日志）与命令行控制端
  tt dev …     T100 设计器包工具：.tzc 安全编辑（export/status/verify/apply）与 .tzs 只读解压
  tt dict …    ERP 数据字典查询：r.t / desc / scc / r.q / prog 等，支持本地镜像与远程直查

  tt env …     环境管理：列出/查看/切换 SSH 环境（三个工具共用同一份环境清单）
  tt config …  配置管理：位置/查看/修改/迁移/校验
  tt serve     启动本地 Web 配置服务（调试工作台，含覆盖三个工具的统一设置页）
  tt install   安装 AI skills 到当前目录，或把 tt 加进用户 PATH

所有输出使用简体中文 (zh_CN)。

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
	pf.BoolVar(&common.JSON, "json", false, "以 JSON 输出")
	pf.BoolVar(&common.CSV, "csv", false, "以 CSV 输出")
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
		// cobra 的 SilenceErrors 让错误只由这里打一次；
		// 带退出码的错误（TDev 的 0/2/3/4/5 契约）按码退出。
		if code := exitCodeOf(err); code != 0 {
			os.Exit(code)
		}
		os.Exit(1)
	}
}
