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
//  4. %APPDATA%\TT\config.json                                    默认：数据目录
//  5. 旧位置兜底（首次运行自动合并迁移到 4）：
//     %APPDATA%\T100\tt\config.json（0.2.0 及更早的默认落点）、
//     <exe 目录>\config.json、<当前目录>\config.json、
//     %APPDATA%\T100\tdebug\config.json、%APPDATA%\T100\tdict\config.json、
//     %APPDATA%\TDebug\config.json
//
// 数据目录可用 TT_HOME 环境变量整体改写（如 TT_HOME=D:\tt）；旧名 T100_HOME 仍被识别，
// 语义相同。两者都不设时是 %APPDATA%\TT。
//
// 为什么从 %APPDATA%\T100\tt 挪出来：那一层 T100 是合并前两个工具
// （T100\tdebug、T100\tdict）的产品根，合并之后只剩一个工具，它底下没有别的东西了。

const (
	// ToolDirName 本工具在**旧**产品根（%APPDATA%\T100）下的子目录名。
	//
	// 它不再是新落点的一层 —— 新落点就是数据目录本身（%APPDATA%\TT，见 UserConfigDir）。
	// 留着这个常量，是因为 %APPDATA%\T100\tt\config.json 是 0.2.0 及更早的默认落点，
	// 现在得当迁移来源把它找回来；常量留着，值改了会让老用户配置找不着。
	ToolDirName = "tt"
	// DefaultConfigName 缺省配置文件名。
	DefaultConfigName = "config.json"
	// PortableMark 便携标记：exe 同目录存在该文件即视为便携包。
	PortableMark = ".portable"
)

// legacyTools 两个旧工具在旧产品根下的子目录名。用于发现待合并的旧配置。
var legacyTools = []string{"tdebug", "tdict"}

// configEnvVars 显式指定配置路径的环境变量，按优先级排列。
// TT_CONFIG 是本工具的；另两个是合并前各自的，保留以免既有脚本失效。
var configEnvVars = []string{"TT_CONFIG", "TDEBUG_CONFIG", "TDICT_CONFIG"}

// legacyRoots 可能放着本工具**旧**配置的"产品根"，按优先级排列。
//
// 每个根底下可能挂着：本工具的旧落点 tt\config.json，以及合并前两个工具的
// tdebug\ / tdict\。两个来源都是历史 ——
//
//   - T100_HOME：用户显式指定的目录。**改落点之前**它是产品根（配置在它下面的 tt\ 里），
//     现在它直接就是数据目录本身，所以它底下那份 tt\config.json 得当成旧配置找回来。
//   - %APPDATA%\T100：合并前那个产品根，tdebug\ / tdict\ / tt\ 都挂在它下面。
//
// **TT_HOME 不在其中**：它是这次新加的，语义就是数据目录本身，没有任何"旧配置"会落在
// 它底下 —— 把它当产品根，只会让我们去 <TT_HOME>\tt\ 找一个从不存在的文件。
//
// **不导出**：新代码只该问 UserConfigDir()/UserConfigPath()。它存在只为两件事 ——
// 迁移时找到旧配置，以及把"本机还有哪些旧配置"报给设置页。
func legacyRoots() []string {
	var out []string
	seen := map[string]bool{}
	add := func(p string) {
		if p == "" {
			return
		}
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		if seen[p] {
			return
		}
		seen[p] = true
		out = append(out, p)
	}
	add(os.Getenv("T100_HOME")) // 旧名，改落点前它是产品根
	if dir, err := os.UserConfigDir(); err == nil {
		add(filepath.Join(dir, "T100"))
	}
	return out
}

// dataDirOfEnv 把一个"数据目录"取值规范化成绝对路径；空串原样返回。
func dataDirOfEnv(env string) string {
	if env == "" {
		return ""
	}
	if abs, err := filepath.Abs(env); err == nil {
		return abs
	}
	return env
}

// UserConfigDir 数据目录（= 配置所在目录）；定位不到时返回空串。
//
// TT_HOME 优先，旧名 T100_HOME 次之。两者语义相同，都是**数据目录本身** ——
// 注意这与改落点之前不同：那时 T100_HOME=D:\t100 指的是 D:\t100\tt，现在指 D:\t100，
// 所以老用户设过它的机器会走一次迁移（来源是 <T100_HOME>\tt\config.json）。
//
// 缓存、快照、服务状态与日志都从这一个值派生（见 cache.go 的 CacheSubdirs），
// 多一个派生源就会多一处"清缓存没清到"或"配置被一起删了"。
func UserConfigDir() string {
	if dir := dataDirOfEnv(os.Getenv("TT_HOME")); dir != "" {
		return dir
	}
	if dir := dataDirOfEnv(os.Getenv("T100_HOME")); dir != "" {
		return dir
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "TT")
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

// LegacyToolConfigPaths 合并前两个旧工具留下的配置路径，只列出真正存在的。
//
// 与 legacyConfigPaths 一样从 legacyRoots 出发，但**用途不同**：这是给设置页展示
// "这台机器上还有哪些旧工具的配置可以并进来"，不含本工具自己的旧落点。
func LegacyToolConfigPaths() []string {
	var out []string
	for _, root := range legacyRoots() {
		for _, tool := range legacyTools {
			p := filepath.Join(root, tool, DefaultConfigName)
			if _, err := os.Stat(p); err == nil {
				out = append(out, p)
			}
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

// oldToolConfigPaths 本工具**改落点之前**可能放配置的位置（%APPDATA%\T100\tt\config.json
// 及设过 T100_HOME 时对应的 <T100_HOME>\tt\config.json），按 legacyRoots 顺序。
//
// "旧落点在哪"这件事只算这一次，两处用它 —— 用途不同但对象是同一份：
// legacyConfigPaths 拿它当读兜底的第一梯队，migrate.collectSources 拿它当合并来源。
func oldToolConfigPaths() []string {
	var out []string
	for _, root := range legacyRoots() {
		out = append(out, filepath.Join(root, ToolDirName, DefaultConfigName))
	}
	return out
}

// legacyConfigPaths 旧位置，按优先级排列。是本工具的配置（凭内容判断），
// 但不在默认落点上。
//
// 顺序 = **读兜底**优先级（第一个存在的胜出）。这与 collectSources 的**合并**顺序
// 不是一回事 —— 那边要的是"谁覆盖谁"。两份顺序别合成一套：合错了会改掉 query 那一节
// 的行为（MergeConfigs 里各节的胜出方向不一致，见那里的注释）。
//
// 旧 tt 落点排最前，因为它是所有旧位置里最权威的一份（0.2.0 及更早的现役配置）。
// 它只在迁移没跑成时才轮得到 —— 正常路径下新落点已在前一个候选里命中。
func legacyConfigPaths() []string {
	out := oldToolConfigPaths()
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
// 两边的键，tdev/tzs/listen/schemaVersion 是合并后新增的。
//
// 漏掉一个键不是形式问题：LooksLikeOwnConfig 靠它判断"当前目录里恰好存在的 config.json
// 是不是我们的"，漏了就会被当成陌生文件，于是旧位置检测与迁移都当它不存在。
var ownConfigKeys = []string{
	"hosts", "debug", "query", "mirror", "bdldoc", "sync", "tdev", "tzs", "listen", "schemaVersion",
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
			// 4. 数据目录；首次运行先把旧配置迁过来
			if src := Migrate(); len(src) > 0 {
				fmt.Fprintf(os.Stderr, "[tt] 已迁移旧配置到数据目录: %s -> %s\n",
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
		// 也会让 --config <临时目录> 的调用（测试与脚本）落在别处。
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
	if b, err := os.ReadFile(p); err == nil {
		if isCurrentSchema(b) {
			return p
		}
		// 文件在、但连 JSON 对象都不是：这不是"合并前的旧配置"，是一份坏文件或根本不是
		// 配置的东西（手改坏的、测试脚本写的临时文件）。**不能拿它当迁移目标** ——
		// 迁移会把统一用户目录里那份（含真实 SSH/数据库口令）整个写进来，等于随手把一个
		// 无关路径变成凭据副本。留给读路径去报"解析不了"。
		if !looksLikeConfig(b) {
			return p
		}
	}
	plan, err := PlanMigration(p)
	if err != nil || plan == nil || len(plan.Sources) == 0 {
		return p
	}
	if err := plan.Apply(); err != nil {
		fmt.Fprintf(os.Stderr, "[tt] 迁移旧配置到 %s 失败: %v\n", p, err)
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

// looksLikeConfig 只回答"这包内容看起来是不是一份配置"：是个 JSON 对象就算。
//
// 与 isCurrentSchema 分开，是因为两者的 false 含义完全不同 —— 前者是"不是配置"，
// 后者是"是配置但结构旧"。只有后者该被就地迁移改写；前者改写就等于覆盖一个陌生文件。
func looksLikeConfig(b []byte) bool {
	var root map[string]any
	return json.Unmarshal(b, &root) == nil
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
