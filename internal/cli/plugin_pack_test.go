package cli

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestPackPluginWritesTTFloor 打包时把 tt 下限写成 SDK 版本，保留上限，排除 dist。
func TestPackPluginWritesTTFloor(t *testing.T) {
	root := t.TempDir()
	ttDir := filepath.Join(root, "tt")
	if err := os.MkdirAll(ttDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ttDir, "VERSION"), []byte("0.7.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plug := filepath.Join(root, "plug")
	if err := os.MkdirAll(filepath.Join(plug, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{
		"schemaVersion": 2,
		"id":            "com.test.pack",
		"version":       "0.1.0",
		"hostApi":       ">=1.0.0 <2.0.0",
		"tt":            ">=0.5.0 <0.8.0",
		"contributions": map[string]any{
			"cli": map[string]any{"namespace": "pk", "entry": "bin/pk.exe"},
		},
	}
	raw, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(plug, "plugin.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plug, "bin", "pk.exe"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plug, "go.mod"), []byte("module x\n\ngo 1.22\n\nrequire github.com/luyaogod/TT v0.0.0\n\nreplace github.com/luyaogod/TT => ../tt\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(plug, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plug, "dist", "old.zip"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, rng, id, ver, err := packPlugin(plug, "")
	if err != nil {
		t.Fatal(err)
	}
	if id != "com.test.pack" || ver != "0.1.0" {
		t.Fatalf("id/version = %s/%s", id, ver)
	}
	if rng != ">=0.7.3 <0.8.0" {
		t.Fatalf("tt 范围 = %q", rng)
	}
	if out != filepath.Join(plug, "dist", "com.test.pack-0.1.0.zip") {
		t.Fatalf("输出路径 = %s", out)
	}

	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	names := map[string]bool{}
	var mf []byte
	for _, f := range zr.File {
		names[f.Name] = true
		if f.Name == "plugin.json" {
			rc, _ := f.Open()
			buf := make([]byte, f.UncompressedSize64)
			_, _ = rc.Read(buf)
			rc.Close()
			mf = buf
		}
	}
	if !names["bin/pk.exe"] || !names["plugin.json"] {
		t.Fatalf("包内容 = %v", names)
	}
	if names["dist/old.zip"] {
		t.Fatal("dist 没有被排除")
	}
	var got map[string]any
	if err := json.Unmarshal(mf, &got); err != nil {
		t.Fatal(err)
	}
	if got["tt"] != ">=0.7.3 <0.8.0" {
		t.Fatalf("包内 tt = %v", got["tt"])
	}
}
