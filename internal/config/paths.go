// Package config 是 tt 的统一配置层：位置解析、读写、校验、迁移。
//
// 合并自 TDebug/cli/root.go 与 TDictCli/cli/root.go 里两份逐行相同、各自标注
// "与对方保持一致，改动请两边同步" 的路径解析，以及两份近乎相同的 cfgfile 包。
// 现在只有这一份实现。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ---------- 配置文件位置 ----------
//
// 存放规则：
//  1. TT_CONFIG 环境变量（兼容旧名 TDEBUG_CONFIG / TDICT_CONFIG）  显式指定
//  2. --config <路径>                                             显式指定
//  3. <exe 目录>\.portable 存在                                    便携包：配置留在包内
//  4. %APPDATA%\T100\tt\config.json                               默认：固定用户的统一位置
//  5. 旧位置兜底（首次运行自动合并迁移到 4）：
//     <exe 目录>\config.json、<当前目录>\config.json、
//     %APPDATA%\T100\tdebug\config.json、%APPDATA%\T100\tdict\config.json、
//     %APPDATA%\TDebug\config.json
//
// 统一位置可用 T100_HOME 环境变量整体改写（如 T100_HOME=D:\t100）。

const (
	// ToolDirName 统一用户目录下本工具的子目录。数据目录 = 配置所在目录，
	// 所以 srccache / debug-bps / logs / .tt-serve.json 都跟着落在同一个目录里。
	ToolDirName = "tt"
	// DefaultConfigName 缺省配置文件名。
	DefaultConfigName = "config.json"
	// PortableMark 便携标记：exe 同目录存在该文件即视为便携包。
	PortableMark = ".portable"
)

// legacyTools 旧工具在统一用户目录下的子目录名。用于发现待合并的旧配置。
var legacyTools = []string{"tdebug", "tdict"}

// configEnvVars 显式指定配置路径的环境变量，按优先级排列。
// TT_CONFIG 是本工具的；另两个是合并前各自的，保留以免既有脚本失效。
var configEnvVars = []string{"TT_CONFIG", "TDEBUG_CONFIG", "TDICT_CONFIG"}

// ToolsHome 固定用户的统一工具目录：T100_HOME 优先，否则 %APPDATA%\T100
// （os.UserConfigDir() 在 Windows 上即 %AppData%）。
func ToolsHome() string {
	if env := os.Getenv("T100_HOME"); env != "" {
		if abs, err := filepath.Abs(env); err == nil {
			return abs
		}
		return env
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "T100")
	}
	return ""
}

// UserConfigDir 统一用户目录下的本工具目录；定位不到时返回空串。
func UserConfigDir() string {
	if home := ToolsHome(); home != "" {
		return filepath.Join(home, ToolDirName)
	}
	return ""
}

// UserConfigPath 统一用户目录下的配置路径；定位不到时返回空串。
func UserConfigPath() string {
	if dir := UserConfigDir(); dir != "" {
		return filepath.Join(dir, DefaultConfigName)
	}
	return ""
}

// LegacyToolConfigPaths 合并前的两个工具在统一用户目录下的配置路径。
// 只列出真正存在的。
func LegacyToolConfigPaths() []string {
	home := ToolsHome()
	if home == "" {
		return nil
	}
	var out []string
	for _, tool := range legacyTools {
		p := filepath.Join(home, tool, DefaultConfigName)
		if _, err := os.Stat(p); err == nil {
			out = append(out, p)
		}
	}
	return out
}

// IsPortable 是否便携包（决定配置留在包内还是进统一用户目录）。
func IsPortable() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(filepath.Dir(exe), PortableMark))
	return err == nil
}

// exeDir 当前可执行文件所在目录；取不到返回空串。
func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

// legacyConfigPaths 旧位置，按优先级排列。是本工具的配置（凭内容判断），
// 但不在默认落点上。
func legacyConfigPaths() []string {
	var out []string
	if dir := exeDir(); dir != "" {
		out = append(out, filepath.Join(dir, DefaultConfigName))
	}
	if cwd, err := os.Getwd(); err == nil {
		out = append(out, filepath.Join(cwd, DefaultConfigName))
	}
	// 桌面版旧数据目录（安装版曾用 %APPDATA%\TDebug）
	if appData := os.Getenv("APPDATA"); appData != "" {
		out = append(out, filepath.Join(appData, "TDebug", DefaultConfigName))
	}
	return out
}

// LooksLikeOwnConfig 判断一段 JSON 是否像本工具的配置 —— 只认自己那几节的顶层键，
// 避免把别的项目（当前目录下恰好存在的）config.json 误迁移过来。
func LooksLikeOwnConfig(b []byte) bool {
	var root map[string]json.RawMessage
	if json.Unmarshal(b, &root) != nil {
		return false
	}
	for _, k := range ownConfigKeys {
		if _, ok := root[k]; ok {
			return true
		}
	}
	return false
}

// ownConfigKeys tt 配置的顶层键。hosts/debug/query/mirror/bdldoc/sync 是合并前
// 两边的键，tdev/listen/schemaVersion 是合并后新增的。
var ownConfigKeys = []string{
	"hosts", "debug", "query", "mirror", "bdldoc", "sync", "tdev", "listen", "schemaVersion",
}

// DefaultConfigPath 未经显式指定时的默认落点：
// 便携包内，否则统一用户目录。
func DefaultConfigPath() string {
	if IsPortable() {
		if dir := exeDir(); dir != "" {
			return filepath.Join(dir, DefaultConfigName)
		}
	}
	if p := UserConfigPath(); p != "" {
		return p
	}
	if abs, err := filepath.Abs(DefaultConfigName); err == nil {
		return abs
	}
	return DefaultConfigName
}

// DefaultConfigPathHint 给 --help 用的一行提示（取不到用户目录时退回文件名）。
func DefaultConfigPathHint() string {
	if IsPortable() {
		return "<exe 目录>\\" + DefaultConfigName
	}
	if dir := UserConfigDir(); dir != "" {
		return filepath.Join(dir, DefaultConfigName)
	}
	return DefaultConfigName
}

// ResolvePath 解析配置文件路径，优先级见包注释。
// allowMissing 为 true 时，若所有候选都不存在则返回默认落点而非报错
// （写配置的命令需要它：首次保存要能创建文件）。
func ResolvePath(flagPath string, allowMissing bool) (string, error) {
	var candidates []string
	explicit := false // 是否由用户显式指定（环境变量或 --config）

	// 1. 环境变量（最高优先级）
	for _, env := range configEnvVars {
		if v := os.Getenv(env); v != "" {
			candidates = append(candidates, v)
			explicit = true
			break
		}
	}

	if flagPath != "" {
		// 2. --config 显式指定：原样 / 相对 exe 同目录 / 相对当前目录
		candidates = append(candidates, flagPath)
		explicit = true
		if !filepath.IsAbs(flagPath) {
			if dir := exeDir(); dir != "" {
				candidates = append(candidates, filepath.Join(dir, flagPath))
			}
			if cwd, err := os.Getwd(); err == nil {
				candidates = append(candidates, filepath.Join(cwd, flagPath))
			}
		}
	} else if !explicit {
		if IsPortable() {
			// 3. 便携包：配置留在包内，不参与统一用户目录与迁移
			if dir := exeDir(); dir != "" {
				candidates = append(candidates, filepath.Join(dir, DefaultConfigName))
			}
		} else if p := UserConfigPath(); p != "" {
			// 4. 统一用户目录；首次运行先把旧配置合并过来
			if src := Migrate(); len(src) > 0 {
				fmt.Fprintf(os.Stderr, "[tt] 已合并旧配置到统一位置: %s -> %s\n",
					joinPaths(src), p)
			}
			candidates = append(candidates, p)
		}
		// 5. 旧位置兜底（用户目录不可用或迁移失败时仍能用）
		candidates = append(candidates, legacyConfigPaths()...)
	}

	var tried []string
	for _, p := range candidates {
		abs, _ := filepath.Abs(p)
		if _, err := os.Stat(abs); err == nil {
			return migrateInPlace(abs), nil
		}
		tried = append(tried, abs)
	}

	if allowMissing {
		// 写路径。用户显式指定了路径时，那就是答案 —— 文件还不存在正是要新建它，
		// 不能悄悄改用默认落点：那会让便携版把配置写到用户目录里去，
		// 也会让 --config <临时目录> 的调用（桌面外壳、冒烟测试）落在别处。
		if explicit && len(candidates) > 0 {
			if abs, err := filepath.Abs(candidates[0]); err == nil {
				return migrateInPlace(abs), nil
			}
			return candidates[0], nil
		}
		if p := DefaultConfigPath(); p != "" {
			return p, nil
		}
	}

	return "", fmt.Errorf(
		"配置文件未找到。\n\n尝试了以下路径:\n%s\n\n缺省位置: %s\n设置 TT_CONFIG 环境变量或使用 --config 指定正确路径:\n  setx TT_CONFIG \"D:\\path\\to\\config.json\"\n  tt --config \"D:\\path\\to\\config.json\" debug status",
		FormatTriedPaths(tried), DefaultConfigPath(),
	)
}

// migrateInPlace 确保 p 处的配置是当前结构：若它还是合并前的旧结构，就地合并迁移。
//
// 为什么不能只依赖缺省落点上的那次迁移：
//   - 便携包。`<exe 目录>\config.json` 在便携模式下优先于统一用户目录，所以
//     迁移分支根本不会走；而便携用户最自然的做法就是把上一版的 config.json
//     直接拷进新包 —— 不迁移的话，旧结构里的 `debug.sshs` 读不出来，环境清单
//     会是空的，用户看到的是"我的环境全没了"。
//   - `--config <路径>` 显式指定时同理。
//
// 文件不存在且没有可合并的旧配置时什么都不做（首次运行由调用方的 EnsureExists
// 落骨架）。迁移失败不报错中断 —— 按未迁移的内容继续，读得出来多少算多少。
func migrateInPlace(p string) string {
	if b, err := os.ReadFile(p); err == nil && isCurrentSchema(b) {
		return p
	}
	plan, err := PlanMigration(p)
	if err != nil || plan == nil || len(plan.Sources) == 0 {
		return p
	}
	if err := plan.Apply(); err != nil {
		fmt.Fprintf(os.Stderr, "[tt] 合并旧配置到 %s 失败: %v\n", p, err)
	}
	return p
}

// isCurrentSchema 判断一段配置内容是否已是合并后的结构。
func isCurrentSchema(b []byte) bool {
	var root map[string]any
	if json.Unmarshal(b, &root) != nil {
		return false // 解析不了就当需要处理，交给 PlanMigration 去报
	}
	return intOf(root["schemaVersion"]) >= SchemaVersion
}

// FormatTriedPaths 把尝试过的路径渲染成 --help/报错里的清单。
func FormatTriedPaths(paths []string) string {
	var s string
	for _, p := range paths {
		exists := ""
		if _, err := os.Stat(p); err == nil {
			exists = " (found)"
		}
		s += fmt.Sprintf("  - %s%s\n", p, exists)
	}
	return s
}

// SamePath 两个路径是否指向同一个文件（比较绝对路径，失败则退回字面比较）。
func SamePath(a, b string) bool {
	aa, err1 := filepath.Abs(a)
	bb, err2 := filepath.Abs(b)
	if err1 != nil || err2 != nil {
		return a == b
	}
	return aa == bb
}

// joinPaths 把来源清单拼成一行，供日志使用。
func joinPaths(paths []string) string {
	if len(paths) == 1 {
		return paths[0]
	}
	s := paths[0]
	for _, p := range paths[1:] {
		s += " + " + p
	}
	return s
}
