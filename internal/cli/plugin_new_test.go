package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/plugin"
)

// TestScaffoldPluginWritesRunnableLayout 骨架：清单/入口/页面/构建脚本齐全。
func TestScaffoldPluginWritesRunnableLayout(t *testing.T) {
	dir := t.TempDir()
	err := writePluginScaffold(dir, scaffold{
		ID: "com.test.hello", Name: "Hello", Namespace: "hello", Route: "hello",
		CLI: true, Service: true, UI: true, Replace: "../tt",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"plugin.json", "go.mod", "client.js", "README.md", "scripts/build.ps1", "cmd/hello/main.go"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f))); err != nil {
			t.Fatalf("缺文件 %s: %v", f, err)
		}
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "plugin.json"))
	var man plugin.Manifest
	if err := json.Unmarshal(raw, &man); err != nil {
		t.Fatal(err)
	}
	if man.ID != "com.test.hello" || man.SchemaVersion != 2 {
		t.Fatalf("清单 = %+v", man)
	}
	if man.Contributions.CLI == nil || man.Contributions.CLI.Entry != "bin/windows-amd64/hello.exe" {
		t.Fatalf("cli 贡献 = %+v", man.Contributions.CLI)
	}
	if man.Contributions.Service == nil || len(man.Contributions.UI) != 1 || man.Contributions.UI[0].Entry != "client.js" {
		t.Fatalf("service/ui 贡献 = %+v / %+v", man.Contributions.Service, man.Contributions.UI)
	}
	gomod, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
	if !strings.Contains(string(gomod), "replace github.com/luyaogod/TT => ../tt") {
		t.Fatalf("go.mod = %s", gomod)
	}
	js, _ := os.ReadFile(filepath.Join(dir, "client.js"))
	if !strings.Contains(string(js), `ctx.slot('page'`) || !strings.Contains(string(js), `route: "hello"`) {
		t.Fatalf("client.js = %s", js)
	}
}

// TestScaffoldNoUI 关闭页面时不生成 client.js / ui 贡献。
func TestScaffoldNoUI(t *testing.T) {
	dir := t.TempDir()
	if err := writePluginScaffold(dir, scaffold{ID: "com.test.cli", Name: "CLI", Namespace: "cli", CLI: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "client.js")); err == nil {
		t.Fatal("不该生成 client.js")
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "plugin.json"))
	if strings.Contains(string(raw), `"ui"`) {
		t.Fatalf("不该有 ui 贡献: %s", raw)
	}
}
