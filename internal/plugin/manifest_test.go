package plugin

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writePluginDir 在临时目录里造一个最小可用插件包，返回目录。
//
// 这是所有安装/启用用例的共用夹具：清单 + CLI 入口 + 一个技能目录。
func writePluginDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ManifestFileName), `{
  "schemaVersion": 1,
  "id": "com.example.demo",
  "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "name": "演示插件",
  "contributions": {
    "cli": {"namespace": "demo", "entry": "bin/demo.bin", "summary": "演示"},
    "frontend": {"route": "demo", "title": "演示页面", "entry": "ui/index.html"},
    "skills": ["skills/demo-skill"]
  },
  "permissions": ["profiles.list", "db.connect"]
}`)
	writeFile(t, filepath.Join(dir, "bin", "demo.bin"), "#!/bin/sh\necho demo\n")
	writeFile(t, filepath.Join(dir, "ui", "index.html"), "<!doctype html><title>demo</title>")
	writeFile(t, filepath.Join(dir, "skills", "demo-skill", "SKILL.md"), "---\nname: demo-skill\ndescription: 演示技能\n---\n\n正文\n")
	return dir
}

// writeFile 建父目录并写文件。
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("建目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写文件失败: %v", err)
	}
}

// TestDecodeManifestIsStrict 清单解码不接受未知字段与多余内容。
//
// 静默忽略写错的键会让插件"装成功、跑起来才炸"，所以装在安装阶段就拒。
func TestDecodeManifestIsStrict(t *testing.T) {
	good := `{"schemaVersion":1,"id":"com.example.a","version":"0.1.0","hostApi":">=1.0.0 <2.0.0",
	          "contributions":{"cli":{"namespace":"a","entry":"bin/a.exe"}}}`
	if _, err := DecodeManifest([]byte(good)); err != nil {
		t.Fatalf("合法清单该通过: %v", err)
	}

	unknown := `{"schemaVersion":1,"id":"com.example.a","version":"0.1.0","hostApi":">=1.0.0",
	             "contributions":{"cli":{"namespace":"a","entry":"bin/a.exe","entryPath":"x"}}}`
	if _, err := DecodeManifest([]byte(unknown)); err == nil {
		t.Fatal("未知字段该被拒绝")
	} else if !strings.Contains(err.Error(), "未知字段") && !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("错误该指出未知字段，得 %v", err)
	}

	trailing := good + `{"extra":1}`
	if _, err := DecodeManifest([]byte(trailing)); err == nil {
		t.Fatal("多余内容该被拒绝")
	}

	if _, err := DecodeManifest([]byte(`{`)); err == nil {
		t.Fatal("坏的 JSON 该被拒绝")
	}
}

// TestNormalizeRelPath 包内路径必须被关在包根里。
func TestNormalizeRelPath(t *testing.T) {
	ok := map[string]string{
		"bin/a.exe":        "bin/a.exe",
		`bin\a.exe`:        "bin/a.exe",
		"./bin//a.exe":     "bin/a.exe",
		"bin/../bin/a.exe": "", // 会被拒（含 ..）
		"日本語/ファイル.txt":     "日本語/ファイル.txt",
	}
	for in, want := range ok {
		got, err := normalizeRelPath(in)
		if want == "" {
			if err == nil {
				t.Errorf("%q 该被拒绝", in)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q 该通过: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("%q 规范化 = %q, want %q", in, got, want)
		}
	}
	bad := []string{"", "/abs/x", "../x", "a/../../b", `C:\x`, `a\x/../../b`, "a\x00b"}
	for _, in := range bad {
		if _, err := normalizeRelPath(in); err == nil {
			t.Errorf("%q 该被拒绝", in)
		}
	}
}

// TestExpandEntrySubstitutesPlatform 入口支持 {os}/{arch} 占位。
func TestExpandEntrySubstitutesPlatform(t *testing.T) {
	got, err := expandEntry("bin/{os}-{arch}/tool")
	if err != nil {
		t.Fatal(err)
	}
	want := "bin/" + runtime.GOOS + "-" + runtime.GOARCH + "/tool"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestValidateShapes 三类名字的形状规则。
func TestValidateShapes(t *testing.T) {
	goodIDs := []string{"com.example.a", "io.github.user.plugin-1", "a.b"}
	for _, id := range goodIDs {
		if err := ValidateID(id); err != nil {
			t.Errorf("id %q 该通过: %v", id, err)
		}
	}
	badIDs := []string{"", "nodot", "Upper.Case", "a..b", ".a.b", "a.b.", "-a.b", "a/b"}
	for _, id := range badIDs {
		if err := ValidateID(id); err == nil {
			t.Errorf("id %q 该被拒绝", id)
		}
	}
	if err := ValidateNamespace("demo-1"); err != nil {
		t.Errorf("namespace 该通过: %v", err)
	}
	for _, ns := range []string{"", "1demo", "Demo", "de mo", "-demo", strings.Repeat("a", 33)} {
		if err := ValidateNamespace(ns); err == nil {
			t.Errorf("namespace %q 该被拒绝", ns)
		}
	}
	if err := ValidateRoute("demo-1"); err != nil {
		t.Errorf("route 该通过: %v", err)
	}
	for _, r := range []string{"", "-demo", "De mo", strings.Repeat("a", 33)} {
		if err := ValidateRoute(r); err == nil {
			t.Errorf("route %q 该被拒绝", r)
		}
	}
}

// TestSemverAndHostAPI 版本必须严格 SemVer；hostApi 是范围。
func TestSemverAndHostAPI(t *testing.T) {
	for _, v := range []string{"0.1.0", "1.2.3", "1.0.0-rc.1", "1.0.0+build.5"} {
		if err := ValidateSemver(v); err != nil {
			t.Errorf("%q 该通过: %v", v, err)
		}
	}
	for _, v := range []string{"", "1", "1.0", "v1.0.0", "1.0.0.0", "latest"} {
		if err := ValidateSemver(v); err == nil {
			t.Errorf("%q 该被拒绝", v)
		}
	}

	ok, err := HostAPISatisfies(">=1.0.0 <2.0.0", "1.0.0")
	if err != nil || !ok {
		t.Fatalf("1.0.0 该落在范围内: ok=%v err=%v", ok, err)
	}
	if ok, _ := HostAPISatisfies(">=2.0.0", "1.0.0"); ok {
		t.Error("主版本不兼容该判 false")
	}
	// 预发布版本不满足普通范围（SemVer 的常规语义），这条守住"别把 rc 当正式版"。
	if ok, _ := HostAPISatisfies(">=1.0.0 <2.0.0", "1.1.0-rc.1"); ok {
		t.Error("预发布版本不该自动满足普通范围")
	}
	if _, err := HostAPISatisfies("not-a-range", "1.0.0"); err == nil {
		t.Error("非法范围该报错")
	}
}

// TestValidateManifestCollectsAllProblems 校验要一次报全，而不是报一个改一个。
func TestValidateManifestCollectsAllProblems(t *testing.T) {
	dir := t.TempDir()
	m := &Manifest{
		SchemaVersion: 99,
		ID:            "Bad_ID",
		Version:       "1.0",
		HostAPI:       "garbage",
		Contributions: Contributions{
			CLI: &CLIContribution{Namespace: "serve", Entry: "../evil.exe"},
		},
		Permissions: []string{"nope", "db.connect", "db.connect"},
	}
	err := ValidateManifest(m, dir, ValidateOptions{
		HostAPIVersion:     "1.0.0",
		ReservedNamespaces: []string{"serve"},
	})
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("该是 ValidationError，得 %v", err)
	}
	joined := strings.Join(ve.Problems, "\n")
	for _, want := range []string{"schemaVersion", "id", "version", "hostApi", "serve", "未知权限", "重复声明"} {
		if !strings.Contains(joined, want) {
			t.Errorf("问题清单缺 %q:\n%s", want, joined)
		}
	}
}

// TestValidateManifestHappyPath 一个合规包该通过。
func TestValidateManifestHappyPath(t *testing.T) {
	dir := writePluginDir(t)
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateManifest(m, dir, ValidateOptions{
		HostAPIVersion:     "1.0.0",
		ReservedNamespaces: []string{"serve", "plugin"},
	}); err != nil {
		t.Fatalf("合规包该通过: %v", err)
	}
}

// TestValidateManifestRejectsMissingEntryAndSkills 声明的入口/技能目录必须真实存在。
func TestValidateManifestRejectsMissingEntryAndSkills(t *testing.T) {
	dir := writePluginDir(t)
	if err := os.Remove(filepath.Join(dir, "bin", "demo.bin")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(dir, "skills", "demo-skill")); err != nil {
		t.Fatal(err)
	}
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	verr := ValidateManifest(m, dir, ValidateOptions{HostAPIVersion: "1.0.0"})
	if verr == nil {
		t.Fatal("缺入口与技能目录该被拒绝")
	}
	msg := verr.Error()
	if !strings.Contains(msg, "cli.entry") || !strings.Contains(msg, "技能目录") {
		t.Fatalf("错误该指出缺什么，得 %v", msg)
	}
}

// TestValidateManifestRejectsNoContribution 四类贡献全无的包对 TT 没有意义。
func TestValidateManifestRejectsNoContribution(t *testing.T) {
	dir := t.TempDir()
	m := &Manifest{SchemaVersion: 1, ID: "com.example.empty", Version: "0.1.0", HostAPI: ">=1.0.0 <2.0.0"}
	err := ValidateManifest(m, dir, ValidateOptions{HostAPIVersion: "1.0.0"})
	if err == nil || !strings.Contains(err.Error(), "没有任何贡献") {
		t.Fatalf("该拒绝无贡献包，得 %v", err)
	}
}

// TestValidateManifestRejectsOtherPluginConflicts 与其他已启用插件的命名空间/路由冲突。
func TestValidateManifestRejectsOtherPluginConflicts(t *testing.T) {
	dir := writePluginDir(t)
	m, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	verr := ValidateManifest(m, dir, ValidateOptions{
		HostAPIVersion:  "1.0.0",
		OtherNamespaces: []string{"demo"},
		OtherRoutes:     []string{"demo"},
	})
	if verr == nil {
		t.Fatal("命名空间/路由冲突该被拒绝")
	}
	msg := verr.Error()
	if !strings.Contains(msg, "已被其他已启用插件占用") {
		t.Fatalf("错误该说明冲突，得 %v", msg)
	}
}
