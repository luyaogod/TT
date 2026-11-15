package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SkillsSync 是插件状态变更的注入点：启用、停用、安装新版本、卸载都必须触发一次重算。
//
// 顺序也在这条用例里钉住：卸载的那次重算必须在删版本目录**之前** —— 技能名只能从
// 版本目录里的清单读出来，删完就再也算不出该撤走哪些副本。
func TestSkillsSyncCallbackWiring(t *testing.T) {
	m, s := newManager(t)
	src := writePluginDir(t)

	var calls []string
	payloadAliveOnUninstall := false
	m.SkillsSync = func(reason string) error {
		calls = append(calls, reason)
		if strings.HasPrefix(reason, "uninstall ") {
			if _, err := os.Stat(s.VersionDir("com.example.demo", "0.1.0")); err == nil {
				payloadAliveOnUninstall = true
			}
		}
		return nil
	}

	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable("com.example.demo", Grants{}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Disable("com.example.demo"); err != nil {
		t.Fatal(err)
	}
	if err := m.Uninstall("com.example.demo", UninstallOptions{}); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"install com.example.demo 0.1.0",
		"enable com.example.demo",
		"disable com.example.demo",
		"uninstall com.example.demo",
	}
	if strings.Join(calls, "|") != strings.Join(want, "|") {
		t.Fatalf("回调 = %v, want %v", calls, want)
	}
	if !payloadAliveOnUninstall {
		t.Error("卸载的重算必须在删版本目录之前：否则算不出该撤走哪些技能")
	}
}

// 技能同步失败不回退启用状态：插件本身已经可用，少的是平台目录里的副本。
func TestEnableKeepsStateWhenSkillsSyncFails(t *testing.T) {
	m, _ := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	m.SkillsSync = func(string) error { return errHook }

	if _, err := m.Enable("com.example.demo", Grants{}); err == nil {
		t.Fatal("技能同步失败该让命令报错（可重试）")
	}
	reg, _ := m.Store.Load()
	rec := reg.Find("com.example.demo")
	if rec == nil || !rec.Enabled {
		t.Fatalf("插件该保持启用（回退会把能用的插件按掉）: %+v", rec)
	}
	if !strings.Contains(rec.LastError, "技能同步失败") {
		t.Errorf("原因该记在 LastError，得 %q", rec.LastError)
	}
}

// 卸载前的重算失败要**中止**卸载：此刻一个字节都还没删。
func TestUninstallAbortsWhenSkillsSyncFails(t *testing.T) {
	m, s := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	m.SkillsSync = func(string) error { return errHook }

	if err := m.Uninstall("com.example.demo", UninstallOptions{}); err == nil {
		t.Fatal("重算失败时卸载该中止")
	}
	if _, err := os.Stat(s.VersionDir("com.example.demo", "0.1.0")); err != nil {
		t.Errorf("版本目录不该被删: %v", err)
	}
	reg, _ := m.Store.Load()
	if reg.Find("com.example.demo") == nil {
		t.Error("注册项不该被移除")
	}
}

// 技能目录取活跃版本目录：换版本后同一个插件指向新目录（P4-08 的来源）。
func TestEnabledSkillSourcesFollowsActiveVersion(t *testing.T) {
	m, s := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable("com.example.demo", Grants{}); err != nil {
		t.Fatal(err)
	}

	srcs, err := s.EnabledSkillSources()
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 || srcs[0].Plugin != "com.example.demo" {
		t.Fatalf("来源 = %+v", srcs)
	}
	old := filepath.Join(s.VersionDir("com.example.demo", "0.1.0"), "skills", "demo-skill")
	if srcs[0].Dir != old {
		t.Fatalf("Dir = %q, want %q", srcs[0].Dir, old)
	}

	// 造一个 0.2.0 版本目录并把活跃版本切过去。
	cur := s.VersionDir("com.example.demo", "0.1.0")
	next := s.VersionDir("com.example.demo", "0.2.0")
	if err := copyTree(cur, next, DefaultZipLimits); err != nil {
		t.Fatalf("准备新版本目录: %v", err)
	}
	if err := s.WithLock(func(reg *Registry) error {
		rec := reg.Find("com.example.demo")
		rec.ActiveVersion = "0.2.0"
		rec.Versions = append(rec.Versions, InstalledVersion{Version: "0.2.0"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	srcs, err = s.EnabledSkillSources()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(next, "skills", "demo-skill")
	if len(srcs) != 1 || srcs[0].Dir != want {
		t.Fatalf("换版本后来源 = %+v, want %q", srcs, want)
	}
}

// 停用后它的技能名进「可撤名单」；启用中的插件不进（它们还在计划里）。
func TestDisabledSkillNames(t *testing.T) {
	m, s := newManager(t)
	if _, err := m.Install(InstallOptions{Source: writePluginDir(t)}); err != nil {
		t.Fatal(err)
	}

	got, err := s.DisabledSkillNames()
	if err != nil {
		t.Fatal(err)
	}
	if !got["demo-skill"] {
		t.Fatalf("已装未启用的插件技能该在可撤名单里: %v", got)
	}

	if _, err := m.Enable("com.example.demo", Grants{}); err != nil {
		t.Fatal(err)
	}
	got, err = s.DisabledSkillNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("启用中插件的技能不该在可撤名单里: %v", got)
	}

	if _, err := m.Disable("com.example.demo"); err != nil {
		t.Fatal(err)
	}
	got, _ = s.DisabledSkillNames()
	if !got["demo-skill"] {
		t.Error("停用后该重新进可撤名单")
	}
}
