package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManifestLoadMissing(t *testing.T) {
	proj := t.TempDir()
	m, exists, err := LoadManifest(proj)
	if err != nil || exists {
		t.Fatalf("缺失清单该是 exists=false、无错误: exists=%v err=%v", exists, err)
	}
	if len(m.Skills("claude")) != 0 {
		t.Error("空清单不该有技能")
	}
}

func TestManifestRoundTrip(t *testing.T) {
	proj := t.TempDir()
	var m Manifest
	if !m.Set("claude", []string{"tt-erp-read", "tt-dev-tzc", "tt-dev-tzc"}) {
		t.Fatal("Set 该报告变化")
	}
	if m.Set("claude", []string{"tt-dev-tzc", "tt-erp-read"}) {
		t.Error("同样的集合不该报告变化")
	}
	if err := m.Save(proj); err != nil {
		t.Fatal(err)
	}
	got, exists, err := LoadManifest(proj)
	if err != nil || !exists {
		t.Fatalf("回读失败: exists=%v err=%v", exists, err)
	}
	want := []string{"tt-dev-tzc", "tt-erp-read"} // 排序
	if list := got.Skills("claude"); len(list) != 2 || list[0] != want[0] || list[1] != want[1] {
		t.Errorf("回读 = %v, want %v", list, want)
	}
	if !got.Set("claude", nil) {
		t.Error("清空该报告变化")
	}
	if err := got.Save(proj); err != nil {
		t.Fatal(err)
	}
	again, _, err := LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := again.Targets["claude"]; ok {
		t.Errorf("清空后不该留下空数组: %v", again.Targets)
	}
}

func TestManifestBadJSON(t *testing.T) {
	proj := t.TempDir()
	if err := os.MkdirAll(filepath.Join(proj, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ManifestPath(proj), []byte("{坏"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadManifest(proj); err == nil {
		t.Fatal("坏清单该报错")
	}
}

func TestManifestDropsInvalidNames(t *testing.T) {
	proj := t.TempDir()
	var m Manifest
	m.Set("codex", []string{"ok-name", "../escape", "has space", "", "ok-name"})
	if err := m.Save(proj); err != nil {
		t.Fatal(err)
	}
	got, _, err := LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	list := got.Skills("codex")
	if len(list) != 1 || list[0] != "ok-name" {
		t.Errorf("应只留合规名字: %v", list)
	}
}
