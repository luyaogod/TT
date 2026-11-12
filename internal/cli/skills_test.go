package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/cli/common"
	"tt/internal/skill"
	"tt/internal/testkit"
)

func withSkillSource(t *testing.T, names ...string) string {
	t.Helper()
	src := t.TempDir()
	mkSkills(t, src, names...)
	old := skillsSource
	skillsSource = func() (string, error) { return src, nil }
	t.Cleanup(func() { skillsSource = old })
	return src
}

func runSkillsMode(t *testing.T, human bool, args ...string) (int, string) {
	t.Helper()
	oldFormat, oldHuman, oldJSON := common.Format, common.Human, common.JSON
	common.Format, common.Human, common.JSON = "json", human, false
	t.Cleanup(func() { common.Format, common.Human, common.JSON = oldFormat, oldHuman, oldJSON })
	return testkit.CaptureStdout(t, func() int {
		cmd := newSkillsCmd()
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			return 1
		}
		return 0
	})
}

func runSkills(t *testing.T, args ...string) (int, string) {
	return runSkillsMode(t, false, args...)
}

func runSkillsHuman(t *testing.T, args ...string) (int, string) {
	return runSkillsMode(t, true, args...)
}

func TestSkillsDeployCmd(t *testing.T) {
	withSkillSource(t, "alpha", "beta")
	proj := t.TempDir()
	writeConfig(t, `{"schemaVersion":4}`)

	code, out := runSkillsHuman(t, "deploy", "--project", proj, "--agent", "claude")
	if code != 0 {
		t.Fatalf("deploy 退出码 = %d, out=%s", code, out)
	}
	if !strings.Contains(out, "已同步") {
		t.Errorf("人类输出该有结果: %s", out)
	}
	for _, n := range []string{"alpha", "beta"} {
		if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", n, "SKILL.md")); err != nil {
			t.Errorf("没装上 %s: %v", n, err)
		}
	}
	m, _, err := skill.LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Skills("claude"); len(got) != 2 {
		t.Errorf("清单 = %v", got)
	}
}

func TestSkillsDeployDryRun(t *testing.T) {
	withSkillSource(t, "alpha")
	proj := t.TempDir()
	writeConfig(t, `{"schemaVersion":4}`)

	code, out := runSkillsHuman(t, "deploy", "--project", proj, "--agent", "codex", "--dry-run")
	if code != 0 {
		t.Fatalf("dry-run 退出码 = %d: %s", code, out)
	}
	if !strings.Contains(out, "试算") {
		t.Errorf("该说明是试算: %s", out)
	}
	if _, err := os.Stat(filepath.Join(proj, ".codex")); !os.IsNotExist(err) {
		t.Error("dry-run 不该建目录")
	}
}

func TestSkillsDeployPicksDetected(t *testing.T) {
	withSkillSource(t, "alpha")
	proj := t.TempDir()
	if err := os.MkdirAll(filepath.Join(proj, ".dsh"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeConfig(t, `{"schemaVersion":4}`)

	code, out := runSkills(t, "deploy", "--project", proj)
	if code != 0 {
		t.Fatalf("退出码 = %d: %s", code, out)
	}
	if _, err := os.Stat(filepath.Join(proj, ".dsh", "skills", "alpha", "SKILL.md")); err != nil {
		t.Error("默认该落到已检测的 .dsh")
	}
	if _, err := os.Stat(filepath.Join(proj, ".agents", "skills", "alpha", "SKILL.md")); err != nil {
		t.Error("默认该落到 .agents")
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude")); !os.IsNotExist(err) {
		t.Error("没检测到的平台不该被创建")
	}
}

func TestSkillsRemoveNeedsYes(t *testing.T) {
	withSkillSource(t, "alpha")
	proj := t.TempDir()
	writeConfig(t, `{"schemaVersion":4}`)
	if code, out := runSkillsHuman(t, "deploy", "--project", proj, "--agent", "claude"); code != 0 {
		t.Fatalf("准备安装失败: %s", out)
	}
	code, _ := runSkills(t, "remove", "--project", proj, "--agent", "claude")
	if code == 0 {
		t.Fatal("没 --yes 的删除该失败")
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", "alpha")); err != nil {
		t.Error("失败时不该删掉任何东西")
	}

	if code, out := runSkills(t, "remove", "--project", proj, "--agent", "claude", "--yes"); code != 0 {
		t.Fatalf("带 --yes 的删除失败: %s", out)
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", "alpha")); !os.IsNotExist(err) {
		t.Error("删除后目录该没了")
	}
}

func TestSkillsListJSON(t *testing.T) {
	withSkillSource(t, "alpha")
	proj := t.TempDir()
	writeConfig(t, `{"schemaVersion":4}`)

	code, out := runSkills(t, "list", "--project", proj)
	if code != 0 {
		t.Fatalf("list 退出码 = %d: %s", code, out)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("list 该输出 JSON: %v\n%s", err, out)
	}
	targets, _ := got["targets"].([]any)
	if len(targets) != 5 {
		t.Fatalf("targets = %d, want 5", len(targets))
	}
	rows, _ := got["skills"].([]any)
	if len(rows) != 1 {
		t.Fatalf("skills = %v", got["skills"])
	}
}

func TestSkillsListBriefJSON(t *testing.T) {
	withSkillSource(t, "alpha")
	proj := t.TempDir()
	writeConfig(t, `{"schemaVersion":4}`)

	code, out := runSkills(t, "list", "--project", proj, "--brief")
	if code != 0 {
		t.Fatalf("list --brief 退出码 = %d: %s", code, out)
	}
	var got struct {
		Skills  []string `json:"skills"`
		Targets []any    `json:"targets"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("brief 该是 JSON: %v\n%s", err, out)
	}
	if len(got.Skills) != 1 || got.Skills[0] != "alpha" || len(got.Targets) != 5 {
		t.Errorf("brief 形状 = %+v", got)
	}
}

func TestSkillsGlobal(t *testing.T) {
	withSkillSource(t, "alpha")
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	old := homeDir
	homeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { homeDir = old })
	writeConfig(t, `{"schemaVersion":4}`)

	code, out := runSkills(t, "deploy", "--global", "--agent", "claude")
	if code != 0 {
		t.Fatalf("全局 deploy 退出码 = %d: %s", code, out)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "alpha", "SKILL.md")); err != nil {
		t.Error("全局落点该是 <home>/.claude/skills")
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", ".tt-manifest.json")); err != nil {
		t.Error("全局清单该在 <home>/.agents/.tt-manifest.json")
	}

	code, out = runSkills(t, "list", "--global")
	if code != 0 {
		t.Fatalf("全局 list 退出码 = %d: %s", code, out)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("list 该是 JSON: %v", err)
	}
	if got["scope"] != "global" || got["project"] != home {
		t.Errorf("全局 list = scope %v project %v", got["scope"], got["project"])
	}
}

func TestSkillsGlobalConflictsWithProject(t *testing.T) {
	withSkillSource(t, "alpha")
	proj := t.TempDir()
	writeConfig(t, `{"schemaVersion":4}`)
	if code, out := runSkills(t, "list", "--global", "--project", proj); code == 0 {
		t.Errorf("--global 与 --project 同给该失败: %s", out)
	}
}

func TestSkillsBadInput(t *testing.T) {
	withSkillSource(t, "alpha")
	proj := t.TempDir()
	writeConfig(t, `{"schemaVersion":4}`)

	cases := [][]string{
		{"deploy", "--project", proj, "--agent", "copilot"},
		{"deploy", "--project", proj, "--all", "--agent", "claude"},
		{"deploy", "--project", filepath.Join(proj, "gone"), "--agent", "claude"},
		{"deploy", "--project", proj, "--agent", "claude", "--skill", "nope"},
		{"list", "--project", filepath.Join(proj, "gone")},
	}
	for _, args := range cases {
		if code, out := runSkills(t, args...); code == 0 {
			t.Errorf("%v 该失败，输出：%s", args, out)
		}
	}
}
