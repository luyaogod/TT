package plugin

// 诊断面（P6-07）：doctor 的新检查项 + 日志读取。
//
// 这些检查项的价值在"能指向下一步"：光说"插件坏了"没用，要说清坏在清单兼容、页面入口、
// 技能来源、还是运行态与注册表对不上。

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// codes 把问题清单压成"码 → 条目"，便于按码断言。
func codes(issues []Issue) map[string]Issue {
	out := map[string]Issue{}
	for _, is := range issues {
		out[is.Code] = is
	}
	return out
}

// writePluginWithSkill 造一个只有 cli + 一个技能的插件（用于同名冲突用例）。
func writePluginWithSkill(t *testing.T, id, namespace, skillDir string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ManifestFileName), fmt.Sprintf(`{
  "schemaVersion": 1, "id": %q, "version": "0.1.0", "hostApi": ">=1.0.0 <2.0.0",
  "contributions": {"cli": {"namespace": %q, "entry": "bin/p.bin"}, "skills": [%q]}
}`, id, namespace, skillDir))
	writeFile(t, filepath.Join(dir, "bin", "p.bin"), "#!/bin/sh\necho p\n")
	writeFile(t, filepath.Join(dir, skillDir, "SKILL.md"), "---\nname: x\ndescription: y\n---\n\n正文\n")
	return dir
}

// TestDoctorHostAPIIncompatible TT 的 Host API 升了大版本之后，老插件要被点名。
func TestDoctorHostAPIIncompatible(t *testing.T) {
	m, _ := newManager(t)
	if _, err := m.Install(InstallOptions{Source: writePluginDir(t)}); err != nil {
		t.Fatal(err)
	}
	// 模拟 TT 升到 Host API 2.x：清单里 >=1.0.0 <2.0.0 装不下了。
	m.HostAPIVersion = "2.0.0"
	issues, err := m.Doctor("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	byCode := codes(issues)
	is, ok := byCode["HOSTAPI_INCOMPATIBLE"]
	if !ok {
		t.Fatalf("该报 hostApi 不兼容，得 %+v", issues)
	}
	if is.Severity != "error" || !strings.Contains(is.Hint, "update") {
		t.Errorf("这条该是 error 且指向升级，得 %+v", is)
	}
	if _, dup := byCode["MANIFEST_INVALID"]; dup {
		t.Error("hostApi 已经单列一条，不该再重复报 MANIFEST_INVALID")
	}
}

// TestDoctorFindsMissingUIEntry 版本目录里的页面入口被删了要大写报告。
func TestDoctorFindsMissingUIEntry(t *testing.T) {
	m, s := newManager(t)
	if _, err := m.Install(InstallOptions{Source: writePluginDir(t)}); err != nil {
		t.Fatal(err)
	}
	ui := filepath.Join(s.VersionDir("com.example.demo", "0.1.0"), "ui", "index.html")
	if err := os.Remove(ui); err != nil {
		t.Fatal(err)
	}
	issues, err := m.Doctor("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := codes(issues)["UI_ENTRY_MISSING"]; !ok {
		t.Fatalf("该报页面入口缺失，得 %+v", issues)
	}
}

// TestDoctorServiceStateMatchesInstanceFile 运行态与 run/<id>/instance.json 对不上要能说清。
func TestDoctorServiceStateMatchesInstanceFile(t *testing.T) {
	runWith := func(t *testing.T, inst string) map[string]Issue {
		t.Helper()
		m, s := newManager(t)
		if _, err := m.Install(InstallOptions{Source: writeServicePluginDirV(t, "0.1.0")}); err != nil {
			t.Fatal(err)
		}
		if _, err := m.Enable("com.example.demo", Grants{}); err != nil {
			t.Fatal(err)
		}
		if inst != "" {
			writeFile(t, filepath.Join(s.RunDir("com.example.demo"), "instance.json"), inst)
		}
		issues, err := m.Doctor("com.example.demo")
		if err != nil {
			t.Fatal(err)
		}
		return codes(issues)
	}

	// 1) 声明了服务却没在跑。
	if _, ok := runWith(t, "")["SERVICE_NOT_RUNNING"]; !ok {
		t.Error("没有实例状态时该报 SERVICE_NOT_RUNNING")
	}
	// 2) 状态文件说在跑，但进程没了。
	dead := `{"pluginId":"com.example.demo","version":"0.1.0","pid":999999,"state":"running"}`
	if _, ok := runWith(t, dead)["SERVICE_STALE_STATE"]; !ok {
		t.Error("进程不在时该报 SERVICE_STALE_STATE")
	}
	// 3) 在跑的实例是旧版本（升级没收拾干净 = 页面新、服务旧）。
	stale := fmt.Sprintf(`{"pluginId":"com.example.demo","version":"0.0.9","pid":%d,"state":"running"}`, os.Getpid())
	byCode := runWith(t, stale)
	is, ok := byCode["SERVICE_VERSION_STALE"]
	if !ok {
		t.Fatal("运行中的实例版本与活跃版本不一致时该报 SERVICE_VERSION_STALE")
	}
	if is.Severity != "error" {
		t.Errorf("串版该是 error，得 %+v", is)
	}
	// 4) 正常：同一版本在跑，没有任何服务相关的问题。
	good := fmt.Sprintf(`{"pluginId":"com.example.demo","version":"0.1.0","pid":%d,"state":"running"}`, os.Getpid())
	for _, code := range []string{"SERVICE_NOT_RUNNING", "SERVICE_STALE_STATE", "SERVICE_VERSION_STALE"} {
		if _, bad := runWith(t, good)[code]; bad {
			t.Errorf("正常状态不该报 %s", code)
		}
	}
}

// TestDoctorSkillsSourceAndConflict 技能来源缺失、以及插件之间同名都要看得出来。
func TestDoctorSkillsSourceAndConflict(t *testing.T) {
	m, s := newManager(t)
	if _, err := m.Install(InstallOptions{Source: writePluginDir(t)}); err != nil {
		t.Fatal(err)
	}
	// 1) 来源缺失。
	skillFile := filepath.Join(s.VersionDir("com.example.demo", "0.1.0"), "skills", "demo-skill", "SKILL.md")
	if err := os.Remove(skillFile); err != nil {
		t.Fatal(err)
	}
	issues, err := m.Doctor("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := codes(issues)["SKILLS_SOURCE_MISSING"]; !ok {
		t.Fatalf("该报技能来源缺失，得 %+v", issues)
	}

	// 2) 同名冲突：另一个插件也声明 skills/demo-skill。
	if _, err := m.Install(InstallOptions{Source: writePluginWithSkill(t, "com.example.other", "other", "skills/demo-skill")}); err != nil {
		t.Fatal(err)
	}
	issues, err = m.Doctor("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := codes(issues)["SKILLS_CONFLICT"]; !ok {
		t.Fatalf("同名技能该被指出，得 %+v", issues)
	}
}

// TestDoctorStaleVersion 多出来的旧版本要被指出（清理策略只保留当前与上一个）。
func TestDoctorStaleVersion(t *testing.T) {
	m, _ := newManager(t)
	if _, err := m.Install(InstallOptions{Source: writePluginDirV(t, "0.1.0")}); err != nil {
		t.Fatal(err)
	}
	// 再装一个新版本：注册表里就有两个，但"上一个"没记录（不是切换出来的）。
	if _, err := m.Install(InstallOptions{Source: writePluginDirV(t, "0.2.0")}); err != nil {
		t.Fatal(err)
	}
	issues, err := m.Doctor("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := codes(issues)["STALE_VERSION"]; !ok {
		t.Fatalf("该指出多出来的旧版本，得 %+v", issues)
	}
}

//---- 日志 ----

// TestNewestLogFileAndTail 最新日志的判据与"最后 N 行"的语义。
func TestNewestLogFileAndTail(t *testing.T) {
	m, s := newManager(t)
	if _, err := m.Install(InstallOptions{Source: writePluginDir(t)}); err != nil {
		t.Fatal(err)
	}
	dir := s.LogsDir("com.example.demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(dir, "service-20260101-000000.log")
	newer := filepath.Join(dir, "service-20260102-000000.log")
	writeFile(t, old, "旧的一行\n")
	writeFile(t, newer, "a\nb\nc\nd\n")
	// 文件系统的修改时间精度可能只有一秒：显式拉开。
	oldTime := time.Now().Add(-time.Hour)
	if err := os.Chtimes(old, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	got, err := s.NewestLogFile("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	if got != newer {
		t.Fatalf("该取最新的那个日志，得 %s", got)
	}
	file, lines, err := s.TailLog("com.example.demo", 2)
	if err != nil {
		t.Fatal(err)
	}
	if file != newer {
		t.Errorf("TailLog 该读最新文件，得 %s", file)
	}
	if strings.Join(lines, ",") != "c,d" {
		t.Errorf("该是最后两行，得 %v", lines)
	}
	// 没有日志不是错误。
	if _, none, err := s.TailLog("com.example.other", 10); err != nil || len(none) != 0 {
		t.Errorf("没有日志时该返回空且不报错: %v %v", none, err)
	}
}

// TestReadTailFileAcrossChunks 文件大于一个读块时也要取对最后 N 行。
func TestReadTailFileAcrossChunks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "big.log")
	var b strings.Builder
	for i := 0; i < 20000; i++ {
		fmt.Fprintf(&b, "line-%05d %s\n", i, strings.Repeat("x", 40))
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err := ReadTailFile(path, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 || !strings.HasPrefix(lines[2], "line-19999") {
		t.Fatalf("该是最后 3 行且末行是 line-19999，得 %v", lines)
	}
}
