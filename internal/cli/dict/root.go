package dict

// tt dict 命令组的根:数据源钩子(打开/关闭查询数据源)与本地 SQLite 路径解析。
//
// 全局开关(--config/--json/--csv/-v/--env)由 tt 根命令的 persistent flags 提供
// (见 internal/cli/common),这里只保留字典独有的部分:--db 与数据源钩子。
// 合并前本文件还带着一份与 TDebug 逐行相同的配置路径解析(toolsHome/userConfigDir/
// isPortable/legacyConfigPaths/resolveConfigPath…),现在统一到 internal/config,
// 那一整块已删除。

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/dict/db"
)

// dbPath -d/--db:本地 SQLite 镜像路径(查询命令读它,db sync 写它)。
var dbPath string

func init() {
	// 字典独有的 persistent flag:--db 只对 tt dict 有意义
	Group.PersistentFlags().StringVarP(&dbPath, "db", "d", "erp_data.db",
		"本地 SQLite 数据库路径(本地查询数据源;也可用 TDICT_DB)")

	// 打开/关闭查询数据源:本地 SQLite(--db/TDICT_DB)或远程库(--env/query.source)。
	// 不依赖本地字典库的子命令(env/mirror/db/install/serve/bdldoc)各自覆盖本钩子。
	Group.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := runRootPreRun(args); err != nil {
			return err
		}
		return openQuerySource()
	}
	Group.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		closeQuerySource()
	}

	attachDataHint() // --help 末尾附一行版本 + 本地数据覆盖状态(见 helpdata.go)
}

// runRootPreRun 显式补跑祖先命令(tt 根)的 PersistentPreRunE。
//
// cobra 只运行「从叶命令往上找到的第一个」PersistentPreRunE —— 子命令一旦定义,
// 父命令的就被整个跳过。今天 tt 根命令没有钩子,但将来若有(全局初始化之类),
// 这里必须把它补上,否则会被静默漏掉。从 Group.Parent() 起找,即刻意跳过 Group
// 自己的钩子(调用方本就是在替代它)。
func runRootPreRun(args []string) error {
	for p := Group.Parent(); p != nil; p = p.Parent() {
		if p.PersistentPreRunE != nil {
			return p.PersistentPreRunE(p, args)
		}
		if p.PersistentPreRun != nil {
			p.PersistentPreRun(p, args)
			return nil
		}
	}
	return nil
}

// skipQuerySource 供「不依赖本地字典库」的子命令覆盖 Group 的数据源钩子,
// 同时补跑祖先命令的钩子。
func skipQuerySource(cmd *cobra.Command, args []string) error { return runRootPreRun(args) }

// resolveDBPath 解析本地 SQLite 路径,优先级:
//  1. TDICT_DB 环境变量
//  2. -d 原样(绝对路径)
//  3. -d 相对 exe 同目录
//  4. -d 相对当前目录
//
// 全部不存在时报错并列出尝试过的路径。
func resolveDBPath(flagPath string) (string, error) {
	var candidates []string

	// 1. TDICT_DB 环境变量(最高优先级)
	if env := os.Getenv("TDICT_DB"); env != "" {
		candidates = append(candidates, env)
	}

	// 2. -d 原样
	candidates = append(candidates, flagPath)

	// 3. -d 相对 exe 同目录
	if !filepath.IsAbs(flagPath) {
		if execPath, err := os.Executable(); err == nil {
			candidates = append(candidates, filepath.Join(filepath.Dir(execPath), flagPath))
		}
	}

	// 4. -d 相对当前目录
	if !filepath.IsAbs(flagPath) {
		if cwd, err := os.Getwd(); err == nil {
			candidates = append(candidates, filepath.Join(cwd, flagPath))
		}
	}

	var tried []string
	for _, p := range candidates {
		abs, _ := filepath.Abs(p)
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
		tried = append(tried, abs)
	}

	return "", fmt.Errorf(
		"本地数据库文件未找到。\n\n尝试了以下路径:\n%s\n\n设置 TDICT_DB 环境变量或用 -d 指定正确路径:\n  setx TDICT_DB \"D:\\path\\to\\erp_data.db\"\n  tt dict -d \"D:\\path\\to\\erp_data.db\" r.t dzea_t",
		config.FormatTriedPaths(tried),
	)
}

// GetDB 返回当前查询数据源(本地 SQLite 或某环境的远程库),
// 由 Group 的 PersistentPreRunE 打开(见 source.go)。
func GetDB() db.Source { return dataSrc }

// IsJSON 是否要求 JSON 输出(--json)。
func IsJSON() bool { return common.JSON }

// IsCSV 是否要求 CSV 输出(--csv)。
func IsCSV() bool { return common.CSV }
