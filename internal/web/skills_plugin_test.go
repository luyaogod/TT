package web

// 插件技能在设置页两条端点上都要看得见（P4-05）：
// GET /api/skills（技能库）与 GET /api/skills/overview（分发矩阵）。

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/plugin"
)

// writeWebPluginPackage 造一个只贡献技能的插件包。
func writeWebPluginPackage(t *testing.T, id, version string, skillNames ...string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "pkg")
	rels := make([]string, 0, len(skillNames))
	for _, n := range skillNames {
		writeTestFile(t, filepath.Join(dir, "skills", n, "SKILL.md"),
			"---\nname: "+n+"\ndescription: 插件技能\n---\n# "+n+"\n")
		rels = append(rels, `"skills/`+n+`"`)
	}
	writeTestFile(t, filepath.Join(dir, "plugin.json"),
		`{"schemaVersion":1,"id":"`+id+`","version":"`+version+`",`+
			`"hostApi":">=1.0.0 <2.0.0","name":"演示插件",`+
			`"contributions":{"skills":[`+strings.Join(rels, ",")+`]}}`)
	return dir
}

// pluginSkillsServer 造一个带插件管理面的服务（配置与注册表同一数据目录）。
func pluginSkillsServer(t *testing.T) (*Server, *plugin.Manager) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(cfgPath, []byte(`{"schemaVersion":4}`), 0o600); err != nil {
		t.Fatal(err)
	}
	m := plugin.NewManager(plugin.NewStore(config.LocationsAt(cfgPath)), nil)
	return New(Options{Version: "test", ConfigPath: cfgPath, PluginMgr: m}), m
}

// enablePluginSkills 装并启用一个插件包。
func enablePluginSkills(t *testing.T, m *plugin.Manager, pkg, id string) {
	t.Helper()
	if _, err := m.Install(plugin.InstallOptions{Source: pkg}); err != nil {
		t.Fatalf("安装插件: %v", err)
	}
	if _, err := m.Enable(id, plugin.Grants{}); err != nil {
		t.Fatalf("启用插件: %v", err)
	}
}

func TestSkillsGetSurfacesPluginSkills(t *testing.T) {
	withOfficialSkills(t, "alpha")
	s, m := pluginSkillsServer(t)
	// 一个合规的、一个与官方撞名的：前者进 plugins，后者只进 warnings。
	enablePluginSkills(t, m, writeWebPluginPackage(t, "com.example.p1", "0.1.0", "demo-skill", "alpha"), "com.example.p1")

	rec, out := doJSON(t, s, "GET", "/api/skills", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d: %s", rec.Code, rec.Body.String())
	}
	plugins, _ := out["plugins"].([]any)
	if len(plugins) != 2 {
		t.Fatalf("plugins = %v, want 2 行", out["plugins"])
	}
	var demo map[string]any
	for _, raw := range plugins {
		row, _ := raw.(map[string]any)
		if row["origin"] != "plugin" || row["plugin"] != "com.example.p1" {
			t.Errorf("插件行该标来源与插件 ID: %v", row)
		}
		if row["name"] == "demo-skill" {
			demo = row
		}
	}
	if demo == nil || demo["state"] != "ok" {
		t.Fatalf("合规的插件技能该列出来: %v", plugins)
	}

	warns, _ := out["warnings"].([]any)
	if len(warns) != 1 {
		t.Fatalf("warnings = %v, want 一条撞名", out["warnings"])
	}
	w, _ := warns[0].(map[string]any)
	if w["kind"] != "conflict" || w["plugin"] != "com.example.p1" {
		t.Errorf("告警该是 conflict 且点名插件: %v", w)
	}
}

func TestSkillsOverviewSurfacesPluginWarnings(t *testing.T) {
	withOfficialSkills(t, "alpha")
	s, m := pluginSkillsServer(t)
	enablePluginSkills(t, m, writeWebPluginPackage(t, "com.example.p1", "0.1.0", "demo-skill", "alpha"), "com.example.p1")

	proj := t.TempDir()
	rec, out := doJSON(t, s, "GET", "/api/skills/overview?project="+url.QueryEscape(proj), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d: %s", rec.Code, rec.Body.String())
	}
	rows, _ := out["skills"].([]any)
	// 官方 alpha + 插件的 demo-skill；撞名的那个不进矩阵（它在 warnings 里）。
	if len(rows) != 2 {
		t.Fatalf("矩阵行 = %v, want 2", out["skills"])
	}
	found := false
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		if row["name"] == "demo-skill" {
			found = true
			if row["origin"] != "plugin" || row["available"] != true {
				t.Errorf("插件技能行形态不对: %v", row)
			}
		}
	}
	if !found {
		t.Error("矩阵里没有插件技能行")
	}

	warns, _ := out["warnings"].([]any)
	if len(warns) != 1 {
		t.Fatalf("warnings = %v, want 一条撞名", out["warnings"])
	}
	// 逐条告警也要能单独看：概览端点必须把 plan.Warnings 带出来（P4-05）。
	if w, _ := warns[0].(map[string]any); w["plugin"] != "com.example.p1" {
		t.Errorf("告警该点名插件: %v", w)
	}
}

// 分发矩阵里出现的插件技能必须真的能装（走的就是计划与 Apply 那条既有路径）。
func TestSkillsDeployWritesPluginSkill(t *testing.T) {
	withOfficialSkills(t, "alpha")
	s, m := pluginSkillsServer(t)
	enablePluginSkills(t, m, writeWebPluginPackage(t, "com.example.p1", "0.1.0", "demo-skill"), "com.example.p1")

	proj := t.TempDir()
	rec, out := doJSON(t, s, "POST", "/api/skills/deploy", map[string]any{
		"project": proj,
		"targets": []map[string]any{{"id": "claude", "skills": []string{"alpha", "demo-skill"}}},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d: %s", rec.Code, rec.Body.String())
	}
	if written, _ := out["written"].([]any); len(written) != 2 {
		t.Fatalf("written = %v", out["written"])
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", "demo-skill", "SKILL.md")); err != nil {
		t.Errorf("插件技能没写入: %v", err)
	}
}
