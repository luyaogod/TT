package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTargetsTable(t *testing.T) {
	want := []struct{ id, dir string }{
		{"agents", ".agents/skills"},
		{"dsh", ".dsh/skills"},
		{"workbuddy", ".codebuddy/skills"},
		{"claude", ".claude/skills"},
		{"codex", ".codex/skills"},
	}
	if len(Targets) != len(want) {
		t.Fatalf("平台数 = %d, want %d", len(Targets), len(want))
	}
	for i, w := range want {
		if Targets[i].ID != w.id || Targets[i].Dir != w.dir {
			t.Errorf("第 %d 个 = %s/%s, want %s/%s", i, Targets[i].ID, Targets[i].Dir, w.id, w.dir)
		}
	}
}

func TestTargetByIDAliases(t *testing.T) {
	cases := map[string]string{
		"agents": "agents", "dsh": "dsh", "deepseek-harness": "dsh",
		"wb": "workbuddy", "claude-code": "claude", "CODEX": "codex",
	}
	for in, want := range cases {
		got, ok := TargetByID(in)
		if !ok || got.ID != want {
			t.Errorf("TargetByID(%q) = %q/%v, want %q", in, got.ID, ok, want)
		}
	}
	if _, ok := TargetByID("copilot"); ok {
		t.Error("copilot 不在 v1 的平台表里")
	}
}

func TestTargetDetected(t *testing.T) {
	proj := t.TempDir()
	if Targets[0].Detected(proj) {
		t.Error("空项目不该检测到 .agents")
	}
	if err := os.MkdirAll(filepath.Join(proj, ".dsh"), 0o755); err != nil {
		t.Fatal(err)
	}
	dsh, _ := TargetByID("dsh")
	if !dsh.Detected(proj) {
		t.Error(".dsh 存在时应检测到 DeepSeek Harness")
	}
	if got := dsh.AbsDir(proj); got != filepath.Join(proj, ".dsh", "skills") {
		t.Errorf("AbsDir = %s", got)
	}
	got := DetectedTargets(proj)
	if len(got) != 1 || got[0].ID != "dsh" {
		t.Errorf("DetectedTargets = %v", got)
	}
}

// 清单单独存在也应该把 .agents 认成“已存在”：后续同步就是往这里写。
func TestAgentsDetectedByManifestOnly(t *testing.T) {
	proj := t.TempDir()
	if err := os.MkdirAll(filepath.Join(proj, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	agents, _ := TargetByID("agents")
	if !agents.Detected(proj) {
		t.Error(".agents 存在（哪怕只有清单）应算已检测")
	}
}
