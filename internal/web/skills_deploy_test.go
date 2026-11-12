package web

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"tt/internal/skill"
)

func TestSkillsOverviewMatrix(t *testing.T) {
	withOfficialSkills(t, "tt-erp-read", "tt-dev-tzc")
	kb := mkExternal(t, "kb-faq")
	proj := t.TempDir()
	if err := os.MkdirAll(filepath.Join(proj, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	s, _ := newTestServer(t, `{"schemaVersion":4,"skills":{"externals":[{"path":`+jsonQuote(kb)+`,"label":"知识库"}]}}`)

	rec, out := doJSON(t, s, "GET", "/api/skills/overview?project="+url.QueryEscape(proj), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, body = %s", rec.Code, rec.Body.String())
	}
	rows, _ := out["skills"].([]any)
	if len(rows) != 3 {
		t.Fatalf("技能行 = %d, want 3（2 官方 + 1 第三方）: %v", len(rows), rows)
	}
	targets, _ := out["targets"].([]any)
	if len(targets) != 5 {
		t.Fatalf("平台列 = %d, want 5", len(targets))
	}
	var claude map[string]any
	for _, raw := range targets {
		m, _ := raw.(map[string]any)
		if m["id"] == "claude" {
			claude = m
		}
	}
	if claude == nil || claude["detected"] != true {
		t.Fatalf("claude 该是已检测: %v", claude)
	}
	if out["manifest"] != false {
		t.Error("没有清单文件时 manifest 该是 false")
	}
}

func TestSkillsDeployWritesAndOverwrites(t *testing.T) {
	withOfficialSkills(t, "tt-erp-read", "tt-dev-tzc")
	proj := t.TempDir()
	s, _ := newTestServer(t, `{"schemaVersion":4}`)

	body := map[string]any{
		"project": proj,
		"targets": []map[string]any{{"id": "claude", "skills": []string{"tt-erp-read", "tt-dev-tzc"}}},
	}
	rec, out := doJSON(t, s, "POST", "/api/skills/deploy", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 = %d, body = %s", rec.Code, rec.Body.String())
	}
	if written, _ := out["written"].([]any); len(written) != 2 {
		t.Fatalf("written = %v", out["written"])
	}
	target := filepath.Join(proj, ".claude", "skills", "tt-erp-read", "SKILL.md")
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("该写入目标: %v", err)
	}
	m, exists, err := skill.LoadManifest(proj)
	if err != nil || !exists {
		t.Fatalf("清单该写出来: %v %v", exists, err)
	}
	if got := m.Skills("claude"); len(got) != 2 {
		t.Errorf("清单 = %v", got)
	}

	// 改副本，再同步：覆盖（没有版本判断）。
	if err := os.WriteFile(target, []byte("手改的"), 0o644); err != nil {
		t.Fatal(err)
	}
	if rec, _ := doJSON(t, s, "POST", "/api/skills/deploy", body); rec.Code != http.StatusOK {
		t.Fatalf("第二次同步失败: %s", rec.Body.String())
	}
	if b, _ := os.ReadFile(target); string(b) == "手改的" {
		t.Error("同步该覆盖手改的副本")
	}
}

func TestSkillsDeployRemovesUnselected(t *testing.T) {
	withOfficialSkills(t, "tt-erp-read")
	proj := t.TempDir()
	s, _ := newTestServer(t, `{"schemaVersion":4}`)

	if rec, _ := doJSON(t, s, "POST", "/api/skills/deploy", map[string]any{
		"project": proj,
		"targets": []map[string]any{{"id": "dsh", "skills": []string{"tt-erp-read"}}},
	}); rec.Code != http.StatusOK {
		t.Fatalf("首次同步失败: %s", rec.Body.String())
	}
	rec, out := doJSON(t, s, "POST", "/api/skills/deploy", map[string]any{
		"project": proj,
		"targets": []map[string]any{{"id": "dsh", "skills": []string{}}},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("清空同步失败: %s", rec.Body.String())
	}
	if removed, _ := out["removed"].([]any); len(removed) != 1 {
		t.Fatalf("removed = %v", out["removed"])
	}
	if _, err := os.Stat(filepath.Join(proj, ".dsh", "skills", "tt-erp-read")); !os.IsNotExist(err) {
		t.Error("取消勾选后目录该被删")
	}
}

func TestSkillsDeployBadRequest(t *testing.T) {
	withOfficialSkills(t, "tt-erp-read")
	proj := t.TempDir()
	s, _ := newTestServer(t, `{"schemaVersion":4}`)

	cases := []map[string]any{
		{"project": proj, "targets": []map[string]any{{"id": "copilot"}}},
		{"project": proj, "targets": []map[string]any{{"id": "claude", "skills": []string{"nope"}}}},
		{"project": filepath.Join(proj, "gone"), "targets": []map[string]any{{"id": "claude"}}},
	}
	for _, body := range cases {
		rec, _ := doJSON(t, s, "POST", "/api/skills/deploy", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("该 400: %v -> %d %s", body, rec.Code, rec.Body.String())
		}
	}
}

func TestSkillsOverviewBadProject(t *testing.T) {
	s, _ := newTestServer(t, `{"schemaVersion":4}`)
	rec, _ := doJSON(t, s, "GET", "/api/skills/overview?project="+url.QueryEscape(filepath.Join(t.TempDir(), "gone")), nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("不存在的项目该 400: %d", rec.Code)
	}
}
