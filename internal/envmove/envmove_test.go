package envmove

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"tt/internal/config"
	"tt/internal/entdir"
)

func writeConfig(t *testing.T, dir, source string) string {
	t.Helper()
	cfg := map[string]any{
		"schemaVersion": config.SchemaVersion,
		"listen":        "127.0.0.1:1",
		"hosts": map[string]any{"sshs": []any{
			map[string]any{"name": "甲", "host": "10.0.0.1", "port": 22, "user": "u", "zone": "1"},
		}},
		"query": map[string]any{"source": source},
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func touch(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// 改名的三份落点都要跟着走，配置里 hosts 与 query.source 同时更新。
func TestRenameMigratesAllPaths(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, "local:甲")

	dictOld := config.DictDBPath(dir, "甲")
	entsOld := entdir.Path(dir, entdir.EnvSeg("甲", "10.0.0.1", "1"))
	mirrorOld := filepath.Join(dir, "srccache", entdir.EnvSeg("甲", "10.0.0.1", "1"))
	touch(t, dictOld, "dict")
	touch(t, entsOld, `{"version":1,"env":"甲","fingerprint":"x","mappings":[{"ent":99,"account":"a","placeholder":false}]}`)
	touch(t, filepath.Join(mirrorOld, "u1", "com", "a.4gl"), "src")

	moves, err := Plan(dir, "甲", "乙", "10.0.0.1", "1")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(moves) != 3 {
		t.Fatalf("该搬 3 份，实得 %d：%+v", len(moves), moves)
	}
	if err := Execute(moves, "乙"); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	rewritten, err := config.RenameHost(cfgPath, "甲", "乙")
	if err != nil {
		t.Fatalf("RenameHost: %v", err)
	}
	if !rewritten {
		t.Error("query.source 指向旧名，改名时该一起改写")
	}

	for _, old := range []string{dictOld, entsOld, mirrorOld} {
		if _, err := os.Stat(old); !os.IsNotExist(err) {
			t.Errorf("旧落点还在：%s", old)
		}
	}
	dictNew := config.DictDBPath(dir, "乙")
	entsNew := entdir.Path(dir, entdir.EnvSeg("乙", "10.0.0.1", "1"))
	for _, p := range []string{dictNew, entsNew, filepath.Join(dir, "srccache", "乙")} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("新落点不存在：%s（%v）", p, err)
		}
	}

	root, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if root.Hosts.SSHs[0].Name != "乙" {
		t.Errorf("环境名 = %q，期望 乙", root.Hosts.SSHs[0].Name)
	}
	if root.Query.Source != "local:乙" {
		t.Errorf("query.source = %q，期望 local:乙", root.Query.Source)
	}

	b, err := os.ReadFile(entsNew)
	if err != nil {
		t.Fatal(err)
	}
	var snap map[string]any
	if err := json.Unmarshal(b, &snap); err != nil {
		t.Fatal(err)
	}
	if snap["env"] != "乙" {
		t.Errorf("企业快照里的 env = %v，期望 乙", snap["env"])
	}
}

// 目标名已被占：整次拒绝，一个都不搬。
func TestPlanRefusesExistingTarget(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "dict", "甲.db"), "a")
	touch(t, filepath.Join(dir, "dict", "乙.db"), "b")

	if _, err := Plan(dir, "甲", "乙", "h", "1"); err == nil {
		t.Fatal("目标已存在时必须报错，不能覆盖")
	}
	if _, err := os.Stat(filepath.Join(dir, "dict", "甲.db")); err != nil {
		t.Errorf("报错后源文件不该动：%v", err)
	}
}

// 什么都没同步过/调试过：计划为空，但改名本身仍要成立。
func TestPlanWithoutPaths(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, "auto")
	moves, err := Plan(dir, "甲", "乙", "h", "1")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if len(moves) != 0 {
		t.Fatalf("不该有落点要搬：%+v", moves)
	}
	if rewritten, err := config.RenameHost(cfgPath, "甲", "乙"); err != nil || rewritten {
		t.Fatalf("RenameHost: rewritten=%v err=%v", rewritten, err)
	}
}

// 保存时校验：指向不存在的环境必须报错，指向存在的放行。
func TestValidateQuerySource(t *testing.T) {
	root := func(src string) map[string]any {
		return map[string]any{
			"hosts": map[string]any{"sshs": []any{map[string]any{"name": "甲"}}},
			"query": map[string]any{"source": src},
		}
	}
	for _, bad := range []string{"local:乙", "乙", "local:"} {
		if err := config.ValidateQuerySource(root(bad)); err == nil {
			t.Errorf("query.source = %q 该报错", bad)
		}
	}
	for _, ok := range []string{"", "auto", "local", "local:甲", "甲"} {
		if err := config.ValidateQuerySource(root(ok)); err != nil {
			t.Errorf("query.source = %q 不该报错：%v", ok, err)
		}
	}
}
