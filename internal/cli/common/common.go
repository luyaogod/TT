// Package common 是三个工具命令组共享的 CLI 上下文：全局开关、内嵌前端、
// 配置路径解析、统一输出约定。
//
// 它是叶子包（不 import 任何 tt 内部命令包），所以 internal/cli 下的
// debug/dev/dict 三个命令组都能安全依赖它，不会与根命令形成循环。
package common

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	"tt/internal/config"
)

// 全局开关。由根命令的 persistent flags 绑定；各命令组只读。
var (
	// ConfigPath --config 显式指定的配置文件路径（空 = 按规则解析）
	ConfigPath string
	// JSON --json：机器可读输出
	JSON bool
	// CSV --csv：CSV 输出（字典类命令用）
	CSV bool
	// Verbose -v：显示解析细节（如实际使用的数据源路径）
	Verbose bool
	// Env --env 指定的环境名；--conn 是其别名（合并前 tdict 的写法）
	Env string
	// WebFS 内嵌的前端构建产物（web/dist），由 main 注入
	WebFS fs.FS
	// Version 版本号，由 build_portable.bat 经 -ldflags 注入
	Version string
)

// ResolveConfig 按统一规则解析配置文件路径。
//
//	allowMissing=false：找不到任何配置时报错（读命令用）
//	allowMissing=true ：返回缺省落点，允许首次保存创建文件（写命令用）
func ResolveConfig(allowMissing bool) (string, error) {
	return config.ResolvePath(ConfigPath, allowMissing)
}

// ConfigHint 给 --help 用的一行配置位置提示。
func ConfigHint() string { return config.DefaultConfigPathHint() }

// WebFrontend 返回内嵌前端的 FS（dist 根，即 web/app 的构建产物）。
//
// 合并前这里是 WebSub(name) —— 当时有两套 SPA（调试工作台与字典页）各占 dist 下一个
// 子目录；字典页已并入统一设置页，只剩一套，所以不需要再按名字取子目录。
// main.go 已把嵌入 FS 的根收敛到 dist，这里直接用。
//
// 取不到时返回 nil —— 调用方应据此回落到引导页，而不是 panic：
// 前端没构建（只有 .gitkeep 占位）是完全正常的状态。
func WebFrontend() fs.FS {
	return WebFS
}

// PrintJSON 以缩进 JSON 输出 v（各命令的 --json 通道统一走它，保证格式一致）。
func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// Fatal 打印错误到 stderr 并以退出码 1 结束。
func Fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
