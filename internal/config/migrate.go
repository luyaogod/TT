package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// 合并前的旧结构里，环境清单和调试设置挤在同一个 debug 节里：
//
//	{ "debug": { "sshs": [...], "listen": "...", "launchArgs": "...", ... } }        ← tdebug
//	{ "hosts": { "sshs": [...], "activeEnv": "..." }, "query": {...}, ... }          ← tdict
//
// 合并后拆成：环境清单统一进 hosts，debug 只留调试设置。
//
// 注意 TDictCli 原有的迁移规则是「debug 整节重命名为 hosts」，那条规则在这里
// 会把 TDebug 的 launchArgs/watchdogSeconds/termWidth 等设置一并吞进 hosts，
// 必须换成本文件的拆解规则。

// Source 一份待合并的旧配置。
type Source struct {
	Path string
	Root map[string]any
}

// Plan 一次迁移计划：合并哪些文件、结果是什么、写到哪。
type Plan struct {
	Dst     string
	Sources []Source
	Merged  map[string]any
}

// Migrate 在缺省落点上执行一次合并迁移；已是最新结构或没有可合并来源时返回 nil。
func Migrate() []string {
	dst := UserConfigPath()
	if dst == "" {
		return nil
	}
	p, err := PlanMigration(dst)
	if err != nil || p == nil || len(p.Sources) == 0 {
		return nil
	}
	if err := p.Apply(); err != nil {
		// 迁移没跑成就等于没迁移，返回 nil 让调用方继续走旧位置兜底。
		// 但要说一声：否则用户看到的是"我的环境全没了"，而没有任何线索。
		fmt.Fprintf(os.Stderr, "[tt] 迁移配置到 %s 失败: %v\n", dst, err)
		return nil
	}
	var paths []string
	for _, s := range p.Sources {
		paths = append(paths, s.Path)
	}
	return paths
}

// PlanMigration 收集所有可合并的旧配置，算出合并结果。
//
// 返回 nil 表示无需迁移：目标已是当前结构，或找不到任何可识别的配置来源。
func PlanMigration(dst string) (*Plan, error) {
	// 目标已存在且已是新结构 → 不再迁移
	if b, err := os.ReadFile(dst); err == nil {
		var root map[string]any
		if json.Unmarshal(b, &root) == nil && intOf(root["schemaVersion"]) >= SchemaVersion {
			return nil, nil
		}
	}

	sources := collectSources(dst)
	if len(sources) == 0 {
		return nil, nil
	}
	return &Plan{
		Dst:     dst,
		Sources: sources,
		Merged:  migrationResult(sources),
	}, nil
}

// migrationResult 决定这次迁移往新落点写什么：**已经是当前结构的来源原样搬过去，
// 只有全都是旧结构时才做合并**。
//
// 为什么不能一律走 MergeConfigs —— 这是改落点之后才出现的情形：
// %APPDATA%\T100\tt\config.json 是 0.2.0 的现役配置，它已经是 SchemaVersion 2。
// 它需要的不是"合并"，是"搬个地方"。而 MergeConfigs 是给合并前那两份不同结构写的，
// 它只认识 hosts/debug/listen/query/mirror/bdldoc/sync/tdev 这几节 —— **tzs 与任何
// 未知顶层键都会被丢掉**。tzs.workspace 没有缺省值，丢了它用户升级后 `tt dev tzs`
// 直接不能跑，还得自己回设置页重填一遍。
//
// 取**最后一个**当前结构的来源：collectSources 把旧落点排在 dst 那一格之前，
// 与"dst 自己排最后"是同一个位置，所以"谁是现役配置"由同一套顺序说了算。
func migrationResult(sources []Source) map[string]any {
	for i := len(sources) - 1; i >= 0; i-- {
		if intOf(sources[i].Root["schemaVersion"]) >= SchemaVersion {
			return sources[i].Root
		}
	}
	return MergeConfigs(sources)
}

// Apply 把合并结果写到目标位置，并给原有文件各留一份 .pre-merge.bak。
// 源文件本身不删除 —— 由用户确认后再自行清理。
func (p *Plan) Apply() error {
	if err := os.MkdirAll(filepath.Dir(p.Dst), 0o755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	for _, s := range p.Sources {
		if SamePath(s.Path, p.Dst) {
			continue
		}
		if b, err := os.ReadFile(s.Path); err == nil {
			_ = os.WriteFile(s.Path+".pre-merge.bak", b, 0o600)
		}
	}
	return Save(p.Dst, p.Merged)
}

// collectSources 按优先级收集可合并的配置来源。
//
// 顺序即优先级，但**具体哪一份胜出取决于各节的合并规则** —— MergeConfigs 里
// hosts.sshs / debug / 顶层 listen 是**靠后者胜出**，而 query/mirror/bdldoc/sync/tdev
// 那几节是**首个非空者胜出**。动这里的顺序之前，先去读 MergeConfigs 对应那一段。
//
//  1. 旧工具目录（<旧产品根>\tdebug、<旧产品根>\tdict）
//  2. exe 同目录 / 当前目录的 config.json（老式便携包或就地运行的遗留）
//  3. 本工具改落点之前的位置（<旧产品根>\tt\config.json）
//  4. 目标位置自身（已是旧结构的 tt 配置，例如上次迁移中断）
func collectSources(dst string) []Source {
	seen := map[string]bool{}
	var out []Source

	add := func(path string) bool {
		if path == "" {
			return false
		}
		abs, _ := filepath.Abs(path)
		if seen[abs] {
			return false
		}
		seen[abs] = true
		b, err := os.ReadFile(abs)
		if err != nil || !LooksLikeOwnConfig(b) {
			return false
		}
		var root map[string]any
		if json.Unmarshal(b, &root) != nil || root == nil {
			return false
		}
		out = append(out, Source{Path: abs, Root: root})
		return true
	}

	// 1. 旧工具目录。tdebug 在后 → 同名环境以它为准（见 MergeConfigs 的说明）。
	for _, p := range legacyToolPaths() {
		add(p)
	}
	// 2. exe 同目录、当前目录
	if dir := exeDir(); dir != "" {
		add(filepath.Join(dir, DefaultConfigName))
	}
	if cwd, err := os.Getwd(); err == nil {
		add(filepath.Join(cwd, DefaultConfigName))
	}
	// 3. 本工具改落点之前的位置（<旧产品根>\tt\config.json）。只收优先级最高的**那一份**：
	// 正常情况下只可能有一份，多个旧产品根同时有配置只可能是用户改过 TT_HOME ——
	// 那时低优先的那份是陈旧副本，并进来反而会让它把 query/mirror 那几节抢走
	// （MergeConfigs 里那几节是"首个非空者胜出"）。
	//
	// 放在这里（而不是前面）是有意的：它占的正是 dst 改落点之前在顺序里的那一格，
	// 所以"旧结构配置的合并优先级"与改落点之前逐条一致。顺序由
	// TestCollectSources_OldLocationSitsRightBeforeDst 钉着。
	for _, p := range oldToolConfigPaths() {
		if add(p) {
			break
		}
	}
	// 4. 目标自身（旧结构）
	add(dst)

	return out
}

// legacyToolPaths 合并前两个工具的配置路径，按 tdebug → tdict 排列。
// 顺序有意如此：两者都配了同一台环境时，tdict 那份是字典查询的现役配置，
// 保留它能让查询侧行为完全不变；tdebug 独有的环境仍然会被并进来。
//
// 每个旧产品根下都排一遍（根的顺序见 legacyRoots）；多根同时有配置是异常情形，
// 那时靠后的根胜出 —— 与"同名环境取靠后那份"是同一条规则。
func legacyToolPaths() []string {
	var out []string
	for _, root := range legacyRoots() {
		for _, tool := range legacyTools {
			out = append(out, filepath.Join(root, tool, DefaultConfigName))
		}
	}
	// 桌面版旧数据目录
	if appData := os.Getenv("APPDATA"); appData != "" {
		out = append(out, filepath.Join(appData, "TDebug", DefaultConfigName))
	}
	return out
}

// MergeConfigs 把若干份旧配置合并成一份新结构。
//
// 规则：
//   - hosts.sshs：先收 tdebug 的 debug.sshs，再用其余来源的 hosts.sshs 覆盖同名项；
//     不同名的环境全部并集保留。
//   - debug 节：只保留调试设置（listen/launchArgs/watchdogSeconds/fglserver/
//     termWidth/termHeight/printElements/persistBreakpoints），sshs 已被提走。
//     靠后的来源覆盖靠前的。
//   - 其余节（query/mirror/bdldoc/sync/tdev）：首个非空者胜出。
//   - listen：顶层 listen 优先，其次 debug.listen，最后缺省值。
//
// **tzs 与未知顶层键不在搬运范围内** —— 本函数只认识上面列出的那几节。所以"已经是
// 当前结构的配置"绝不能走它（会丢 tzs.workspace 这类没有缺省值的键），见 migrationResult。
func MergeConfigs(sources []Source) map[string]any {
	root := map[string]any{}
	root["schemaVersion"] = SchemaVersion

	debugSec := map[string]any{}
	hostsSec := map[string]any{"sshs": []any{}}
	order := []string{}        // 环境名出现顺序
	byName := map[string]any{} // 环境名 → 条目

	addEnv := func(entry any, override bool) {
		m, ok := entry.(map[string]any)
		if !ok {
			return
		}
		name, _ := m["name"].(string)
		if name == "" {
			return
		}
		if _, exists := byName[name]; !exists {
			order = append(order, name)
		} else if !override {
			return
		}
		byName[name] = m
	}

	takeEnvs := func(src map[string]any, override bool) {
		for _, sec := range []string{"debug", "hosts"} {
			s, _ := src[sec].(map[string]any)
			if s == nil {
				continue
			}
			list, _ := s["sshs"].([]any)
			for _, e := range list {
				addEnv(e, override)
			}
		}
	}
	// 先按“不覆盖”收一遍（tdebug 的环境占位），再按“覆盖”收一遍（tdict 的同名项胜出）。
	// 与 collectSources 的 tdebug → tdict 顺序配合，达到“同名保留 tdict”的效果。
	for i, s := range sources {
		takeEnvs(s.Root, i > 0)
	}

	sshs := make([]any, 0, len(order))
	for _, n := range order {
		sshs = append(sshs, byName[n])
	}
	hostsSec["sshs"] = sshs

	for _, s := range sources {
		src := s.Root

		// 调试设置：从 debug 节里抠掉 sshs/listen/activeEnv 之外一概保留
		if dbg, ok := src["debug"].(map[string]any); ok {
			for k, v := range dbg {
				switch k {
				case "sshs":
					continue // 已提到 hosts
				case "listen":
					if _, has := root["listen"]; !has {
						root["listen"] = v
					}
					continue
				}
				debugSec[k] = v
			}
		}
		// 顶层 listen 优先于 debug.listen
		if v, ok := src["listen"].(string); ok && v != "" {
			root["listen"] = v
		}
		// activeEnv 在旧 tdict 里属于 hosts 节
		if h, ok := src["hosts"].(map[string]any); ok {
			if v, ok := h["activeEnv"].(string); ok && v != "" {
				hostsSec["activeEnv"] = v
			}
		}
		// 其余节：首个非空者胜出
		for _, key := range []string{"query", "mirror", "bdldoc", "sync", "tdev"} {
			if _, has := root[key]; has {
				continue
			}
			if v, ok := src[key]; ok && v != nil {
				root[key] = v
			}
		}
	}

	// activeEnv 兜底取首条环境
	if _, has := hostsSec["activeEnv"]; !has && len(order) > 0 {
		hostsSec["activeEnv"] = order[0]
	}

	root["hosts"] = hostsSec
	if len(debugSec) > 0 {
		root["debug"] = debugSec
	}
	if _, has := root["listen"]; !has {
		root["listen"] = DefaultListen
	}
	return root
}
