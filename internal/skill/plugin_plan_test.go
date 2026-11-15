package skill

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Origin 的字符串是**对外契约**：Web 与脚本按它分类，插件来源必须与官方/第三方区分开。
func TestOriginPluginString(t *testing.T) {
	if OriginPlugin != "plugin" {
		t.Fatalf("OriginPlugin = %q, want %q", OriginPlugin, "plugin")
	}
	seen := map[Origin]bool{}
	for _, o := range []Origin{OriginOfficial, OriginExternal, OriginPlugin} {
		if o == "" || seen[o] {
			t.Fatalf("来源字符串必须非空且互不相同: %q", o)
		}
		seen[o] = true
	}
}

// 顺序即优先级：官方 → 用户登记的第三方 → 插件。插件撞名只告警，计划照常产出。
func TestBuildPlanWithPluginSkillsPrecedence(t *testing.T) {
	official := t.TempDir()
	mkSkill(t, official, "alpha", goodBody("alpha"))

	ext := t.TempDir()
	beta := mkSkill(t, ext, "beta", goodBody("beta"))

	pA := t.TempDir()
	gamma := mkSkill(t, pA, "gamma", goodBody("gamma"))
	alphaDup := mkSkill(t, pA, "alpha", goodBody("alpha")) // 与官方撞名

	pB := t.TempDir()
	betaDup := mkSkill(t, pB, "beta", goodBody("beta")) // 与第三方撞名

	plan, err := BuildPlanWithPluginSkills(official, []string{beta}, []Source{
		{Origin: OriginPlugin, Plugin: "com.example.a", Dir: gamma},
		{Origin: OriginPlugin, Plugin: "com.example.a", Dir: alphaDup},
		{Origin: OriginPlugin, Plugin: "com.example.b", Dir: betaDup},
	})
	if err != nil {
		t.Fatalf("插件侧的问题只该告警: %v", err)
	}

	var got []string
	for _, s := range plan.Skills {
		got = append(got, string(s.Origin)+":"+s.Name)
	}
	want := []string{"official:alpha", "external:beta", "plugin:gamma"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("计划 = %v, want %v（插件必须排在官方与第三方之后）", got, want)
	}
	if plan.Skills[2].Plugin != "com.example.a" {
		t.Errorf("插件技能该记住来源插件: %+v", plan.Skills[2])
	}

	if len(plan.Warnings) != 2 {
		t.Fatalf("warnings = %+v, want 2 条撞名", plan.Warnings)
	}
	for _, w := range plan.Warnings {
		if w.Kind != StateConflict {
			t.Errorf("kind = %s, want conflict", w.Kind)
		}
		if w.Plugin == "" {
			t.Errorf("告警必须点名是哪个插件带的: %+v", w)
		}
		if !strings.Contains(w.Text(), "插件 "+w.Plugin) {
			t.Errorf("文案该点名插件: %q", w.Text())
		}
	}
	if !strings.Contains(plan.Warnings[0].Why, "官方技能") {
		t.Errorf("与官方撞名要说明官方优先: %q", plan.Warnings[0].Why)
	}
	if !strings.Contains(plan.Warnings[1].Why, "第三方") {
		t.Errorf("与第三方撞名要说明是谁占的: %q", plan.Warnings[1].Why)
	}
}

// 插件目录没了 / SKILL.md 坏了：只告警，官方与其它插件照装（与第三方同一策略）。
func TestBuildPlanPluginSkillMissingAndInvalid(t *testing.T) {
	official := t.TempDir()
	mkSkill(t, official, "alpha", goodBody("alpha"))

	good := t.TempDir()
	demo := mkSkill(t, good, "demo-skill", goodBody("demo-skill"))
	bad := t.TempDir()
	badSkill := mkSkill(t, bad, "bad-skill", "---\nname: 另一个名字\ndescription: x\n---\n")

	plan, err := BuildPlanWithPluginSkills(official, nil, []Source{
		{Origin: OriginPlugin, Plugin: "com.example.p", Dir: filepath.Join(t.TempDir(), "gone")},
		{Origin: OriginPlugin, Plugin: "com.example.p", Dir: badSkill},
		{Origin: OriginPlugin, Plugin: "com.example.p", Dir: demo},
	})
	if err != nil {
		t.Fatalf("插件侧问题不该让计划失败: %v", err)
	}
	if len(plan.Skills) != 2 || plan.Skills[1].Name != "demo-skill" {
		t.Fatalf("合规的插件技能该照装: %+v", plan.Skills)
	}
	kinds := []State{StateMissing, StateInvalid}
	if len(plan.Warnings) != len(kinds) {
		t.Fatalf("warnings = %+v", plan.Warnings)
	}
	for i, want := range kinds {
		if plan.Warnings[i].Kind != want {
			t.Errorf("第 %d 条 kind = %s, want %s（%s）", i, plan.Warnings[i].Kind, want, plan.Warnings[i].Why)
		}
		if plan.Warnings[i].Plugin != "com.example.p" {
			t.Errorf("第 %d 条没点名插件: %+v", i, plan.Warnings[i])
		}
	}
}

// P4-03 的兼容要求：没有插件来源时，两条入口必须给出**同一份计划**。
func TestBuildPlanUnchangedWithoutPluginSkills(t *testing.T) {
	official := t.TempDir()
	mkSkill(t, official, "alpha", goodBody("alpha"))
	if err := os.WriteFile(filepath.Join(official, "README.md"), []byte("索引\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ext := t.TempDir()
	ok := mkSkill(t, ext, "kb-faq", goodBody("kb-faq"))
	gone := filepath.Join(t.TempDir(), "gone")
	base, err := BuildPlan(official, []string{ok, gone})
	if err != nil {
		t.Fatal(err)
	}
	withPlugins, err := BuildPlanWithPluginSkills(official, []string{ok, gone}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(base, withPlugins) {
		t.Fatalf("BuildPlan 行为变了：\n base=%+v\n with=%+v", base, withPlugins)
	}
}

// P4-06：插件技能只进目标目录，**绝不写回 <exe>/skills**（那边是发版包）。
func TestPluginSkillsNeverCopiedIntoOfficialDir(t *testing.T) {
	official := t.TempDir()
	mkSkill(t, official, "alpha", goodBody("alpha"))
	if err := os.WriteFile(filepath.Join(official, "README.md"), []byte("索引\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := treeSnapshot(t, official)

	p := t.TempDir()
	demo := mkSkill(t, p, "demo-skill", goodBody("demo-skill"))
	plan, err := BuildPlanWithPluginSkills(official, nil, []Source{
		{Origin: OriginPlugin, Plugin: "com.example.p", Dir: demo},
	})
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "skills")
	rep, err := Apply(plan, dst, Options{Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, "demo-skill", "SKILL.md")); err != nil {
		t.Fatalf("插件技能没装到目标目录: %v（files=%v）", err, rep.Copied)
	}
	if after := treeSnapshot(t, official); !reflect.DeepEqual(before, after) {
		t.Fatalf("官方技能目录被改动了：\n before=%v\n after =%v", before, after)
	}
}

// treeSnapshot 记下一棵树的相对路径（含目录），用来断言"一个字节都没动"。
func treeSnapshot(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return rerr
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("遍历 %s: %v", root, err)
	}
	return out
}

//---------------------------------------------------------------------------
// Withdraw：只撤「清单记过 + 源确实撤走」的条目（P4-10）
//---------------------------------------------------------------------------

// 停用一个插件后：它的技能撤走，用户的目录与其它插件的技能必须留着。
func TestWithdrawOnlyRemovesGoneSources(t *testing.T) {
	official := t.TempDir()
	mkSkill(t, official, "alpha", goodBody("alpha"))

	p1 := t.TempDir()
	demo := mkSkill(t, p1, "demo-skill", goodBody("demo-skill"))
	p2 := t.TempDir()
	other := mkSkill(t, p2, "other-skill", goodBody("other-skill"))

	full, err := BuildPlanWithPluginSkills(official, nil, []Source{
		{Origin: OriginPlugin, Plugin: "com.example.p1", Dir: demo},
		{Origin: OriginPlugin, Plugin: "com.example.p2", Dir: other},
	})
	if err != nil {
		t.Fatal(err)
	}
	proj := t.TempDir()
	names := []string{"alpha", "demo-skill", "other-skill"}
	if _, err := Deploy(full, DeployRequest{Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: names}}}); err != nil {
		t.Fatal(err)
	}
	// 用户自己在平台目录里建的技能（清单里没有）。
	if err := os.MkdirAll(filepath.Join(proj, ".claude", "skills", "mine"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, ".claude", "skills", "mine", "SKILL.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}

	// p1 停用：新计划里没有 demo-skill；它属于"已启用→已停用"，可撤。
	after, err := BuildPlanWithPluginSkills(official, nil, []Source{
		{Origin: OriginPlugin, Plugin: "com.example.p2", Dir: other},
	})
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Withdraw(WithdrawRequest{
		Project:   proj,
		Targets:   []DeployTarget{{ID: "claude", Skills: []string{"alpha", "other-skill"}}},
		Removable: map[string]bool{"demo-skill": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Removed) != 1 || rep.Removed[0].Skill != "demo-skill" {
		t.Fatalf("removed = %+v, want 只撤 demo-skill", rep.Removed)
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", "demo-skill")); !os.IsNotExist(err) {
		t.Error("停用插件的技能该被撤走")
	}
	for _, keep := range []string{"alpha", "other-skill", "mine"} {
		if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", keep)); err != nil {
			t.Errorf("%s 不该被删: %v", keep, err)
		}
	}

	m, _, err := LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Skills("claude"); strings.Join(got, ",") != "alpha,other-skill" {
		t.Errorf("清单 = %v, want [alpha other-skill]", got)
	}
	_ = after // 计划只用于说明"源已经不在了"；撤走判据来自 Removable
}

// 不可撤的条目（第三方只是暂时失联）一律保留：那与"源被删"在盘上分不出来。
func TestWithdrawKeepsUncertainSources(t *testing.T) {
	official := t.TempDir()
	mkSkill(t, official, "alpha", goodBody("alpha"))
	proj := t.TempDir()
	if _, err := Deploy(officialPlan(t, "alpha", "kb-faq"), DeployRequest{
		Project: proj, Targets: []DeployTarget{{ID: "claude", Skills: []string{"alpha", "kb-faq"}}},
	}); err != nil {
		t.Fatal(err)
	}

	rep, err := Withdraw(WithdrawRequest{
		Project: proj,
		Targets: []DeployTarget{{ID: "claude", Skills: []string{"alpha"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Removed) != 0 || len(rep.Skipped) != 1 {
		t.Fatalf("removed=%+v skipped=%+v", rep.Removed, rep.Skipped)
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", "kb-faq")); err != nil {
		t.Error("状态不确定的条目该保留旧副本")
	}
	m, _, err := LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(m.Skills("claude"), ","); got != "alpha,kb-faq" {
		t.Errorf("清单该保留 kb-faq: %v", got)
	}
}
