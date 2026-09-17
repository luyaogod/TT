package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// 用本机真实存在的旧配置做一次迁移，检验合并没有丢东西。
//
// 这是唯一一类能证明"迁移规则贴合现实"的测试 —— 前面那些用的是手写夹具，
// 手写夹具只能验证我设想的旧结构；这个验证的是真的旧结构。
// 本机没有旧配置时（CI、别人的机器）自动跳过，不制造噪声。
func TestMigrate_RealLegacyConfigs(t *testing.T) {
	home := ToolsHome()
	if home == "" {
		t.Skip("定位不到统一用户目录")
	}

	var sources []Source
	for _, tool := range legacyTools {
		p := filepath.Join(home, tool, DefaultConfigName)
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if !LooksLikeOwnConfig(b) {
			t.Logf("跳过 %s：不像本工具的配置", p)
			continue
		}
		root, err := unmarshalRoot(b)
		if err != nil {
			t.Errorf("%s 解析失败: %v", p, err)
			continue
		}
		sources = append(sources, Source{Path: p, Root: root})
		t.Logf("读入旧配置 %s", p)
	}
	if len(sources) == 0 {
		t.Skip("本机没有合并前的旧配置，跳过")
	}

	merged := MergeConfigs(sources)

	// 1. 结果必须能被类型化视图读回
	dir := t.TempDir()
	out := filepath.Join(dir, DefaultConfigName)
	if err := Save(out, merged); err != nil {
		t.Fatalf("保存合并结果失败: %v", err)
	}
	r, err := Load(out)
	if err != nil {
		t.Fatalf("读回合并结果失败: %v", err)
	}

	// 2. 旧配置里的每个环境都必须还在
	for _, s := range sources {
		for _, key := range []string{"hosts", "debug"} {
			sec, _ := s.Root[key].(map[string]any)
			if sec == nil {
				continue
			}
			list, _ := sec["sshs"].([]any)
			for _, e := range list {
				m, _ := e.(map[string]any)
				name, _ := m["name"].(string)
				if name == "" {
					continue
				}
				if r.Hosts.ByName(name) == nil {
					t.Errorf("环境 %q（来自 %s）在合并结果里丢失", name, s.Path)
				}
			}
		}
	}

	// 3. 调试设置不能丢进 hosts
	for _, s := range sources {
		dbg, _ := s.Root["debug"].(map[string]any)
		if dbg == nil {
			continue
		}
		if v, ok := dbg["watchdogSeconds"]; ok && intOf(v) != 0 {
			if r.Debug.WatchdogSeconds != intOf(v) {
				t.Errorf("watchdogSeconds 丢失: 旧 %v, 合并后 %d", v, r.Debug.WatchdogSeconds)
			}
		}
		if v, _ := dbg["launchArgs"].(string); v != "" && r.Debug.LaunchArgs != v {
			t.Errorf("launchArgs 丢失: 旧 %q, 合并后 %q", v, r.Debug.LaunchArgs)
		}
	}

	// 4. 字典侧设置要带过来
	for _, s := range sources {
		for _, key := range []string{"query", "mirror", "bdldoc", "sync"} {
			sec, _ := s.Root[key].(map[string]any)
			if sec == nil {
				continue
			}
			for k, v := range sec {
				got, ok := lookupInRoot(r, key, k)
				if !ok {
					t.Errorf("%s.%s（来自 %s）在合并结果里丢失", key, k, s.Path)
					continue
				}
				if got != v {
					t.Errorf("%s.%s 变了: 旧 %v, 合并后 %v", key, k, v, got)
				}
			}
		}
	}

	// 5. hosts 节绝不能沾上调试设置 —— 这正是 TDictCli 旧规则会踩的坑
	var raw map[string]any
	if err := loadRaw(out, &raw); err != nil {
		t.Fatal(err)
	}
	hosts, _ := raw["hosts"].(map[string]any)
	for _, k := range []string{"watchdogSeconds", "launchArgs", "termWidth", "printElements", "fglserver"} {
		if _, has := hosts[k]; has {
			t.Errorf("调试设置 %q 被误并进了 hosts 节", k)
		}
	}

	t.Logf("合并后：%d 个环境，activeEnv=%q，listen=%q，schemaVersion=%d",
		len(r.Hosts.SSHs), r.Hosts.ActiveEnv, r.Listen, r.SchemaVersion)
}

func unmarshalRoot(b []byte) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func loadRaw(path string, v *map[string]any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// lookupInRoot 在类型化视图里按 节.键 取原始值（只覆盖测试用到的几个字符串/数字节）。
func lookupInRoot(r *Root, section, key string) (any, bool) {
	switch section {
	case "query":
		if key == "source" {
			return r.Query.Source, true
		}
	case "mirror":
		if key == "dir" {
			return r.Mirror.Dir, true
		}
	case "bdldoc":
		if key == "dir" {
			return r.Bdldoc.Dir, true
		}
	case "sync":
		if key == "target" {
			return r.Sync.Target, true
		}
	}
	return nil, false
}
