package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func officialPlan(t *testing.T, names ...string) Plan {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		mkSkill(t, dir, n, goodBody(n))
	}
	p, err := BuildPlan(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDeployWritesAndOverwrites(t *testing.T) {
	plan := officialPlan(t, "alpha", "beta")
	proj := t.TempDir()
	req := DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: []string{"beta", "alpha"}}}}
	rep, err := Deploy(plan, req)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Written) != 2 {
		t.Fatalf("written = %+v", rep.Written)
	}
	alpha := filepath.Join(proj, ".claude", "skills", "alpha", "SKILL.md")
	if got := readFile(t, alpha); got != goodBody("alpha") {
		t.Errorf("首次写入内容不对: %q", got)
	}
	m, exists, err := LoadManifest(proj)
	if err != nil || !exists {
		t.Fatalf("清单该写出来: exists=%v err=%v", exists, err)
	}
	if got := m.Skills("claude"); len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Errorf("清单 = %v", got)
	}

	// 改掉目标里的副本，再同步：必须被覆盖（不做版本判断）。
	if err := os.WriteFile(alpha, []byte("手改过的"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Deploy(plan, req); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, alpha); got != goodBody("alpha") {
		t.Errorf("同步没覆盖手改的副本: %q", got)
	}
}

func TestDeployRemovesOnlyManifestEntries(t *testing.T) {
	plan := officialPlan(t, "alpha")
	proj := t.TempDir()
	if _, err := Deploy(plan, DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: []string{"alpha"}}}}); err != nil {
		t.Fatal(err)
	}
	// 用户自己的同名目录之外的东西：不该被删。
	mine := filepath.Join(proj, ".claude", "skills", "mine")
	if err := os.MkdirAll(mine, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mine, "SKILL.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}

	rep, err := Deploy(plan, DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: nil}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Removed) != 1 || rep.Removed[0].Skill != "alpha" {
		t.Fatalf("removed = %+v", rep.Removed)
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", "alpha")); !os.IsNotExist(err) {
		t.Error("alpha 该被删掉")
	}
	if _, err := os.Stat(mine); err != nil {
		t.Error("清单外的 mine 不该被删")
	}
	m, _, err := LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Skills("claude")) != 0 {
		t.Errorf("清单该清空: %v", m.Skills("claude"))
	}
}

func TestDeployDryRunWritesNothing(t *testing.T) {
	plan := officialPlan(t, "alpha", "beta")
	proj := t.TempDir()
	rep, err := Deploy(plan, DeployRequest{
		Project: proj, DryRun: true,
		Targets: []DeployTarget{{ID: "dsh", Skills: []string{"alpha", "beta"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Written) != 2 || !rep.DryRun {
		t.Fatalf("试算该算出 2 条写入: %+v", rep)
	}
	if _, err := os.Stat(filepath.Join(proj, ".dsh")); !os.IsNotExist(err) {
		t.Error("试算不该建目录")
	}
	if _, err := os.Stat(ManifestPath(proj)); !os.IsNotExist(err) {
		t.Error("试算不该写清单")
	}
}

// 旧命令/手拷留下的同名目录：没有清单也照样覆盖，并纳入清单。
func TestDeployAdoptsExistingDir(t *testing.T) {
	plan := officialPlan(t, "alpha")
	proj := t.TempDir()
	old := filepath.Join(proj, ".codex", "skills", "alpha")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "SKILL.md"), []byte("旧副本"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Deploy(plan, DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "codex", Skills: []string{"alpha"}}}}); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(old, "SKILL.md")); got != goodBody("alpha") {
		t.Errorf("旧副本该被覆盖: %q", got)
	}
	m, _, err := LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Skills("codex"); len(got) != 1 || got[0] != "alpha" {
		t.Errorf("清单 = %v", got)
	}
}

func TestDeployRejectsBadInput(t *testing.T) {
	plan := officialPlan(t, "alpha")
	proj := t.TempDir()
	cases := []struct {
		name string
		req  DeployRequest
	}{
		{"没有平台", DeployRequest{Project: proj}},
		{"未知平台", DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "copilot"}}}},
		{"未知技能", DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: []string{"nope"}}}}},
		{"技能名不合规", DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: []string{"../x"}}}}},
		{"项目不存在", DeployRequest{Project: filepath.Join(proj, "nope"), Targets: []DeployTarget{{ID: "claude"}}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Deploy(plan, c.req); err == nil {
				t.Fatal("该报错")
			}
		})
	}
}

// 某个平台有第三方失联不影响其它平台：告警带出去，照常写。
func TestDeployCarriesWarnings(t *testing.T) {
	official := t.TempDir()
	mkSkill(t, official, "alpha", goodBody("alpha"))
	plan, err := BuildPlan(official, []string{filepath.Join(t.TempDir(), "gone")})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Warnings) != 1 {
		t.Fatalf("该有一条失联告警: %+v", plan.Warnings)
	}
	proj := t.TempDir()
	rep, err := Deploy(plan, DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: []string{"alpha"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Warnings) != 1 || len(rep.Written) != 1 {
		t.Errorf("告警该原样带出、写入照常: warnings=%v written=%v", rep.Warnings, rep.Written)
	}
}

// 清单里的旧名字被清掉时不影响另一个平台的记录。
func TestDeployManifestPerTarget(t *testing.T) {
	plan := officialPlan(t, "alpha")
	proj := t.TempDir()
	if _, err := Deploy(plan, DeployRequest{Project: proj, Targets: []DeployTarget{
		{ID: "claude", Skills: []string{"alpha"}},
		{ID: "dsh", Skills: []string{"alpha"}},
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Deploy(plan, DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: nil}}}); err != nil {
		t.Fatal(err)
	}
	m, _, err := LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Skills("claude")) != 0 {
		t.Errorf("claude 该清空: %v", m.Skills("claude"))
	}
	if got := m.Skills("dsh"); len(got) != 1 {
		t.Errorf("dsh 不该受影响: %v", got)
	}
	if _, err := os.Stat(filepath.Join(proj, ".dsh", "skills", "alpha")); err != nil {
		t.Error("dsh 的副本不该被删")
	}
}

// 技能源消失（移动硬盘没插、第三方被删）时，已装好的副本不删，清单也保留。
func TestDeployKeepsCopyWhenSourceGone(t *testing.T) {
	plan := officialPlan(t, "alpha", "beta")
	proj := t.TempDir()
	if _, err := Deploy(plan, DeployRequest{Project: proj, Targets: []DeployTarget{
		{ID: "claude", Skills: []string{"alpha", "beta"}},
	}}); err != nil {
		t.Fatal(err)
	}

	// 新计划里没有 alpha 了（模拟来源消失）。
	next := officialPlan(t, "beta")
	rep, err := Deploy(next, DeployRequest{Project: proj, Targets: []DeployTarget{
		{ID: "claude", Skills: []string{"beta"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Removed) != 0 {
		t.Errorf("来源消失的条目不删: %+v", rep.Removed)
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", "alpha")); err != nil {
		t.Error("alpha 的旧副本该保留")
	}
	m, _, err := LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Skills("claude"); len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Errorf("清单该保留 alpha: %v", got)
	}
	if len(rep.Skipped) != 1 {
		t.Errorf("该留一条 skipped: %+v", rep.Skipped)
	}
}

// 写入路径必须落在目标目录里（清单被手改成怪名字时也不能越界）。
func TestDeploySkipsUnsafeManifestName(t *testing.T) {
	plan := officialPlan(t, "alpha")
	proj := t.TempDir()
	if err := os.MkdirAll(filepath.Join(proj, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"schema":1,"targets":{"claude":["..\\evil","alpha"]}}`
	if err := os.WriteFile(ManifestPath(proj), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err := Deploy(plan, DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: []string{"alpha"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Skipped) != 1 {
		t.Errorf("越界名字该进 skipped: %+v", rep.Skipped)
	}
	if rep.Skipped[0].Why == "" {
		t.Error("skip 该带原因")
	}
}
