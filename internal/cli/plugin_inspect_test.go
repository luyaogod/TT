package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeInspectFixture(t *testing.T, entry string, withEntry bool) string {
	t.Helper()
	dir := t.TempDir()
	man := map[string]any{
		"schemaVersion": 2,
		"id":            "com.test.inspect",
		"version":       "0.1.0",
		"hostApi":       ">=1.0.0 <2.0.0",
		"contributions": map[string]any{
			"cli": map[string]any{"namespace": "ins", "entry": entry},
		},
	}
	b, _ := json.Marshal(man)
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	if withEntry {
		full := filepath.Join(dir, filepath.FromSlash(entry))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("exe"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestInspectPackageDir 目录包：清单、文件、体积、入口齐全。
func TestInspectPackageDir(t *testing.T) {
	dir := writeInspectFixture(t, "bin/tool.exe", true)
	r, err := inspectPackage(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if r["ok"] != true {
		t.Fatalf("不该有问题: %v", r["problems"])
	}
	if r["kind"] != "package" || r["totalBytes"].(int64) <= 0 {
		t.Fatalf("report = %+v", r)
	}
	names := map[string]bool{}
	for _, f := range r["files"].([]map[string]any) {
		names[f["path"].(string)] = true
	}
	if !names["plugin.json"] || !names["bin/tool.exe"] {
		t.Fatalf("文件清单 = %v", names)
	}
}

// TestInspectPackageMissingEntry 入口缺失要报问题。
func TestInspectPackageMissingEntry(t *testing.T) {
	dir := writeInspectFixture(t, "bin/tool.exe", false)
	r, err := inspectPackage(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if r["ok"] != false {
		t.Fatal("缺入口该有问题")
	}
	if len(r["problems"].([]string)) == 0 {
		t.Fatal("problems 为空")
	}
}
