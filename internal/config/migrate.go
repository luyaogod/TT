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
		Merged:  MergeConfigs(sources),
	}, nil
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
// 顺序即优先级：靠后的在冲突时胜出。
//
//  1. 旧工具目录（%APPDATA%\T100\tdebug、%APPDATA%\T100\tdict）
//  2. exe 同目录 / 当前目录的 config.json（老式便携包或就地运行的遗留）
//  3. 目标位置自身（已是旧结构的 tt 配置，例如上次迁移中断）
func collectSources(dst string) []Source {
	seen := map[string]bool{}
	var out []Source

	add := func(path string) {
		if path == "" {
			return
		}
		abs, _ := filepath.Abs(path)
		if seen[abs] {
			return
		}
		seen[abs] = true
		b, err := os.ReadFile(abs)
		if err != nil || !LooksLikeOwnConfig(b) {
			return
		}
		var root map[string]any
		if json.Unmarshal(b, &root) != nil || root == nil {
			return
		}
		out = append(out, Source{Path: abs, Root: root})
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
	// 3. 目标自身（旧结构）
	add(dst)

	return out
}

// legacyToolPaths 合并前两个工具的配置路径，按 tdebug → tdict 排列。
// 顺序有意如此：两者都配了同一台环境时，tdict 那份是字典查询的现役配置，
// 保留它能让查询侧行为完全不变；tdebug 独有的环境仍然会被并进来。
func legacyToolPaths() []string {
	home := ToolsHome()
	if home == "" {
		return nil
	}
	var out []string
	for _, tool := range legacyTools {
		out = append(out, filepath.Join(home, tool, DefaultConfigName))
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
