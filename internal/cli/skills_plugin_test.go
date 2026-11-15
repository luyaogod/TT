package cli

// 插件技能进入分发管线（P4-03/05/06/07/08/09/10/11 的命令层用例）。
//
// 这里用真实插件包走一遍注册表：安装、启用、停用、换版本都由 plugin.Manager 完成，
// 技能来源则由 buildSkillsPlan / syncPluginSkills 从注册表投影出来 —— 与命令行实际
// 走的是同一条路。

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/plugin"
	"tt/internal/skill"
	"tt/internal/testkit"
)

// writePluginPackage 造一个只贡献技能的插件包：每个技能目录一份合规 SKILL.md。
func writePluginPackage(t *testing.T, id, version string, skillNames ...string) string {
	t.Helper()
	return writePluginPackageAssets(t, id, version, nil, skillNames...)
}

// writePluginPackageAssets 同上，但给 withAssets 里的技能各加一个 assets/ 文件。
func writePluginPackageAssets(t *testing.T, id, version string, withAssets map[string]bool, skillNames ...string) string {
	t.Helper()
	dir := t.TempDir()
	rels := make([]string, 0, len(skillNames))
	for _, n := range skillNames {
		writeSkill(t, filepath.Join(dir, "skills"), n, "---\nname: "+n+"\ndescription: 插件技能。\n---\n# "+n+"\n")
		if withAssets[n] {
			if err := os.MkdirAll(filepath.Join(dir, "skills", n, "assets"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "skills", n, "assets", "note.txt"), []byte("资源"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		rels = append(rels, `"skills/`+n+`"`)
	}
	manifest := `{"schemaVersion":1,"id":"` + id + `","version":"` + version + `",` +
		`"hostApi":">=1.0.0 <2.0.0","name":"演示插件",` +
		`"contributions":{"skills":[` + strings.Join(rels, ",") + `]}}`
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// pluginTestManager 用当前 TT_CONFIG 指向的数据目录造一个管理面。
// SkillsSync 故意不接：这些用例要显式控制"什么时候重算"。
func pluginTestManager(t *testing.T) *plugin.Manager {
	t.Helper()
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		t.Fatal(err)
	}
	return plugin.NewManager(plugin.NewStore(config.LocationsAt(cfgPath)), nil)
}

// installAndEnable 安装并启用一个插件包。
func installAndEnable(t *testing.T, m *plugin.Manager, pkg, id string) {
	t.Helper()
	if _, err := m.Install(plugin.InstallOptions{Source: pkg}); err != nil {
		t.Fatalf("安装插件: %v", err)
	}
	if _, err := m.Enable(id, plugin.Grants{}); err != nil {
		t.Fatalf("启用插件: %v", err)
	}
}

// withSyncRoot 把"当前项目"换成 proj、把"用户目录"换成一个空目录：
// syncPluginSkills 只看这两个根，测不到真实用户目录。
func withSyncRoot(t *testing.T, proj string) {
	t.Helper()
	t.Chdir(proj)
	home := t.TempDir()
	old := homeDir
	homeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { homeDir = old })
}

// dirTree 记下一棵树的相对路径（断言"一个字节都没动"用）。
func dirTree(t *testing.T, root string) []string {
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
// 计划来源
//---------------------------------------------------------------------------

// 启用的插件技能进计划；停用后从计划里消失（P4-03 / P4-11 的"停用即消失"）。
func TestBuildSkillsPlanIncludesPluginSkills(t *testing.T) {
	official := withSkillSource(t, "alpha")
	writeConfig(t, `{"schemaVersion":4}`)
	m := pluginTestManager(t)
	installAndEnable(t, m, writePluginPackage(t, "com.example.p1", "0.1.0", "demo-skill"), "com.example.p1")

	plan, src, err := buildSkillsPlan()
	if err != nil {
		t.Fatal(err)
	}
	if src != official {
		t.Errorf("计划该报出官方目录 %q，得 %q", official, src)
	}
	var got []string
	for _, s := range plan.Skills {
		got = append(got, string(s.Origin)+":"+s.Name)
	}
	if strings.Join(got, ",") != "official:alpha,plugin:demo-skill" {
		t.Fatalf("计划 = %v（插件技能必须排在官方之后）", got)
	}
	if plan.Skills[1].Plugin != "com.example.p1" {
		t.Errorf("该记住来源插件: %+v", plan.Skills[1])
	}

	if _, err := m.Disable("com.example.p1"); err != nil {
		t.Fatal(err)
	}
	plan, _, err = buildSkillsPlan()
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Skills) != 1 || plan.Skills[0].Origin != skill.OriginOfficial {
		t.Fatalf("停用后插件技能该从计划里消失: %+v", plan.Skills)
	}
}

// 两条既有安装路径都要装上插件技能（P4-09）。
func TestPluginSkillsFlowThroughBothInstallPaths(t *testing.T) {
	withSkillSource(t, "alpha")
	writeConfig(t, `{"schemaVersion":4}`)
	m := pluginTestManager(t)
	installAndEnable(t, m, writePluginPackage(t, "com.example.p1", "0.1.0", "demo-skill"), "com.example.p1")

	// 1) tt skills deploy
	proj := t.TempDir()
	code, out := runSkills(t, "deploy", "--project", proj, "--agent", "claude")
	if code != 0 {
		t.Fatalf("skills deploy 失败: %s", out)
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", "demo-skill", "SKILL.md")); err != nil {
		t.Fatalf("tt skills deploy 没装上插件技能: %v", err)
	}

	// 2) tt install skills
	dst := filepath.Join(t.TempDir(), "skills")
	if code, out := runInstallSkills(t, "--to", dst); code != 0 {
		t.Fatalf("install skills 失败: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dst, "demo-skill", "SKILL.md")); err != nil {
		t.Fatalf("tt install skills 没装上插件技能: %v", err)
	}
}

// P4-06：插件技能只进目标目录，官方 <exe>/skills 一个字节都不动。
func TestPluginSkillsNeverTouchOfficialDir(t *testing.T) {
	official := withSkillSource(t, "alpha")
	writeConfig(t, `{"schemaVersion":4}`)
	m := pluginTestManager(t)
	installAndEnable(t, m, writePluginPackage(t, "com.example.p1", "0.1.0", "demo-skill"), "com.example.p1")
	before := dirTree(t, official)

	dst := filepath.Join(t.TempDir(), "skills")
	if code, out := runInstallSkills(t, "--to", dst, "--force"); code != 0 {
		t.Fatalf("install skills 失败: %s", out)
	}
	proj := t.TempDir()
	if code, out := runSkills(t, "deploy", "--project", proj, "--all"); code != 0 {
		t.Fatalf("skills deploy 失败: %s", out)
	}

	if after := dirTree(t, official); strings.Join(after, "|") != strings.Join(before, "|") {
		t.Fatalf("官方技能目录被改动了：\n before=%v\n after =%v", before, after)
	}
}

// 插件技能与官方撞名：只告警，且命令行两个入口都看得见（P4-05）。
func TestPluginSkillConflictVisibleInCLI(t *testing.T) {
	withSkillSource(t, "alpha")
	writeConfig(t, `{"schemaVersion":4}`)
	m := pluginTestManager(t)
	// 插件的技能名与官方 alpha 相同 → 计划里跳过它，但插件本身照常启用。
	installAndEnable(t, m, writePluginPackage(t, "com.example.p1", "0.1.0", "alpha"), "com.example.p1")

	proj := t.TempDir()
	code, out := runSkills(t, "list", "--project", proj)
	if code != 0 {
		t.Fatalf("skills list 失败: %s", out)
	}
	var listed struct {
		Skills   []map[string]any `json:"skills"`
		Warnings []map[string]any `json:"warnings"`
	}
	if err := json.Unmarshal([]byte(out), &listed); err != nil {
		t.Fatalf("list 该是 JSON: %v\n%s", err, out)
	}
	if len(listed.Skills) != 2 {
		t.Fatalf("skills = %v（一条可用 + 一条冲突行）", listed.Skills)
	}
	avail := 0
	for _, row := range listed.Skills {
		if row["available"] == true {
			avail++
			continue
		}
		if row["origin"] != "plugin" || row["plugin"] != "com.example.p1" {
			t.Errorf("冲突行该标成插件来源: %v", row)
		}
	}
	if avail != 1 {
		t.Fatalf("冲突的插件技能不该进计划: %v", listed.Skills)
	}
	if len(listed.Warnings) != 1 {
		t.Fatalf("warnings = %v, want 一条撞名告警", listed.Warnings)
	}
	w := listed.Warnings[0]
	if w["kind"] != "conflict" || w["plugin"] != "com.example.p1" {
		t.Errorf("告警该是 conflict 且点名插件: %v", w)
	}
	if !strings.Contains(w["why"].(string), "官方技能") {
		t.Errorf("该说明输给谁了: %v", w["why"])
	}

	// tt install skills --dry-run 的 JSON 里也要有这一条。
	code, out = runInstallSkills(t, "--to", filepath.Join(t.TempDir(), "skills"), "--dry-run")
	if code != 0 {
		t.Fatalf("install skills --dry-run 失败: %s", out)
	}
	var install struct {
		Warnings []map[string]any `json:"warnings"`
	}
	if err := json.Unmarshal([]byte(out), &install); err != nil {
		t.Fatalf("dry-run 该是 JSON: %v\n%s", err, out)
	}
	if len(install.Warnings) != 1 {
		t.Fatalf("dry-run warnings = %v", install.Warnings)
	}
	if text, _ := install.Warnings[0]["text"].(string); !strings.Contains(text, "com.example.p1") {
		t.Errorf("告警文案该点名插件: %q", text)
	}
}

//---------------------------------------------------------------------------
// 期望态重算
//---------------------------------------------------------------------------

// P4-08：换版本后重算，已装副本的 assets 指针指向**新**版本目录，且没有残留旧版本引用。
func TestSyncRefreshesPluginSkillPointersAfterUpgrade(t *testing.T) {
	withSkillSource(t, "alpha")
	writeConfig(t, `{"schemaVersion":4}`)
	m := pluginTestManager(t)
	withAssets := map[string]bool{"demo-skill": true}
	installAndEnable(t, m, writePluginPackageAssets(t, "com.example.p1", "0.1.0", withAssets, "demo-skill"), "com.example.p1")

	proj := t.TempDir()
	if code, out := runSkills(t, "deploy", "--project", proj, "--agent", "claude"); code != 0 {
		t.Fatalf("首次分发失败: %s", out)
	}
	copyMD := filepath.Join(proj, ".claude", "skills", "demo-skill", "SKILL.md")
	oldAssets := filepath.Join(m.Store.VersionDir("com.example.p1", "0.1.0"), "skills", "demo-skill", "assets")
	if !strings.Contains(readFile(t, copyMD), oldAssets) {
		t.Fatalf("首次分发该指向旧版本目录 %s：\n%s", oldAssets, readFile(t, copyMD))
	}

	// 升级：停用 → 装 0.2.0 → 启用。真实流程里每一步都会重算；这里显式调一次，
	// 断言"重算"本身把已装副本刷新成新版本目录。
	if _, err := m.Disable("com.example.p1"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Install(plugin.InstallOptions{Source: writePluginPackageAssets(t, "com.example.p1", "0.2.0", withAssets, "demo-skill")}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable("com.example.p1", plugin.Grants{}); err != nil {
		t.Fatal(err)
	}

	withSyncRoot(t, proj)
	if err := syncPluginSkills("test upgrade"); err != nil {
		t.Fatalf("重算失败: %v", err)
	}

	got := readFile(t, copyMD)
	newAssets := filepath.Join(m.Store.VersionDir("com.example.p1", "0.2.0"), "skills", "demo-skill", "assets")
	if !strings.Contains(got, newAssets) {
		t.Errorf("指针没刷成新版本目录 %s：\n%s", newAssets, got)
	}
	if strings.Contains(got, oldAssets) {
		t.Errorf("还残留旧版本引用 %s：\n%s", oldAssets, got)
	}

	// 回滚（活跃版本切回 0.1.0）同样只靠这一次重算：指针必须跟着回去。
	if err := m.Store.WithLock(func(reg *plugin.Registry) error {
		rec := reg.Find("com.example.p1")
		if rec == nil {
			t.Fatal("注册表里没有 com.example.p1")
		}
		rec.ActiveVersion = "0.1.0"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := syncPluginSkills("test rollback"); err != nil {
		t.Fatalf("回滚重算失败: %v", err)
	}
	got = readFile(t, copyMD)
	if !strings.Contains(got, oldAssets) {
		t.Errorf("回滚后指针该指回 0.1.0 的目录 %s：\n%s", oldAssets, got)
	}
	if strings.Contains(got, newAssets) {
		t.Errorf("回滚后还残留 0.2.0 的引用 %s：\n%s", newAssets, got)
	}
}

// P4-10：停用一个插件后只撤它记过的技能；用户自建目录与其它插件的技能必须留着。
func TestSyncWithdrawsOnlyGonePluginSkills(t *testing.T) {
	withSkillSource(t, "alpha")
	writeConfig(t, `{"schemaVersion":4}`)
	m := pluginTestManager(t)
	installAndEnable(t, m, writePluginPackage(t, "com.example.p1", "0.1.0", "demo-skill"), "com.example.p1")
	installAndEnable(t, m, writePluginPackage(t, "com.example.p2", "0.1.0", "other-skill"), "com.example.p2")

	proj := t.TempDir()
	if code, out := runSkills(t, "deploy", "--project", proj, "--agent", "claude"); code != 0 {
		t.Fatalf("首次分发失败: %s", out)
	}
	// 用户自己在平台目录里建的技能（清单里没有）。
	mine := filepath.Join(proj, ".claude", "skills", "mine")
	if err := os.MkdirAll(mine, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mine, "SKILL.md"), []byte("我自己放的"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := m.Disable("com.example.p1"); err != nil {
		t.Fatal(err)
	}
	withSyncRoot(t, proj)
	if err := syncPluginSkills("test disable"); err != nil {
		t.Fatalf("重算失败: %v", err)
	}

	skillsDir := filepath.Join(proj, ".claude", "skills")
	for _, gone := range []string{"demo-skill"} {
		if _, err := os.Stat(filepath.Join(skillsDir, gone)); !os.IsNotExist(err) {
			t.Errorf("停用插件的技能 %s 该被撤走（err=%v）", gone, err)
		}
	}
	for _, keep := range []string{"alpha", "other-skill", "mine"} {
		if _, err := os.Stat(filepath.Join(skillsDir, keep, "SKILL.md")); err != nil {
			t.Errorf("%s 不该被删: %v", keep, err)
		}
	}
	man, _, err := skill.LoadManifest(proj)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(man.Skills("claude"), ","); got != "alpha,other-skill" {
		t.Errorf("清单 = %v, want [alpha other-skill]", got)
	}
}

// P4-11：一个平台写不进去只报告它，其它平台照常刷新。
func TestSyncReportsPartialPlatformFailure(t *testing.T) {
	official := withSkillSource(t, "alpha")
	writeConfig(t, `{"schemaVersion":4}`)
	pluginTestManager(t) // 没有插件也照样要重算（官方/第三方来源）

	proj := t.TempDir()
	if code, out := runSkills(t, "deploy", "--project", proj, "--agent", "claude", "--agent", "dsh"); code != 0 {
		t.Fatalf("首次分发失败: %s", out)
	}

	// 把 .dsh/skills 换成同名文件：这个平台从此写不进去（确定性失败，不靠权限）。
	dsh := filepath.Join(proj, ".dsh", "skills")
	if err := os.RemoveAll(dsh); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dsh, []byte("占位"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 改掉来源内容：刷新过的副本一眼看得出来。
	if err := os.WriteFile(filepath.Join(official, "alpha", "SKILL.md"),
		[]byte("---\nname: alpha\ndescription: 改过了\n---\n# 改过了\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	withSyncRoot(t, proj)
	err := syncPluginSkills("test partial")
	if err == nil {
		t.Fatal("有一个平台写不进去时必须报出来")
	}
	if !strings.Contains(err.Error(), "dsh") {
		t.Errorf("错误该点名坏掉的平台: %v", err)
	}
	got := readFile(t, filepath.Join(proj, ".claude", "skills", "alpha", "SKILL.md"))
	if !strings.Contains(got, "改过了") {
		t.Errorf("其它平台该照常刷新：\n%s", got)
	}
}

// syncPluginSkills 不外溢到没部署过技能的范围（不新建平台目录）。
func TestSyncNeverCreatesPlatformDirs(t *testing.T) {
	withSkillSource(t, "alpha")
	writeConfig(t, `{"schemaVersion":4}`)
	proj := t.TempDir()
	home := t.TempDir()
	old := homeDir
	homeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { homeDir = old })
	t.Chdir(proj)

	if err := syncPluginSkills("test empty"); err != nil {
		t.Fatalf("没有部署过时该是空操作: %v", err)
	}
	for _, dir := range []string{".agents", ".dsh", ".claude", ".codex"} {
		if _, err := os.Stat(filepath.Join(proj, dir)); !os.IsNotExist(err) {
			t.Errorf("不该新建 %s（err=%v）", dir, err)
		}
	}
	if _, err := os.Stat(filepath.Join(proj, ".agents", skill.ManifestRel[len(".agents/"):])); !os.IsNotExist(err) {
		t.Error("不该写平台清单")
	}
}

// 真正的命令层接线：`tt plugin enable/disable` 里的回调就是 skills_sync.go 的那一份
// （pluginManager() 装配），启用后技能出现在已部署平台，停用后撤走（P4-07）。
func TestPluginCommandsResyncSkills(t *testing.T) {
	withSkillSource(t, "alpha")
	writeConfig(t, `{"schemaVersion":4}`)
	m := pluginTestManager(t)
	if _, err := m.Install(plugin.InstallOptions{Source: writePluginPackage(t, "com.example.p1", "0.1.0", "demo-skill")}); err != nil {
		t.Fatalf("安装插件: %v", err)
	}
	proj := t.TempDir()
	if code, out := runSkills(t, "deploy", "--project", proj, "--agent", "claude"); code != 0 {
		t.Fatalf("首次分发失败: %s", out)
	}
	withSyncRoot(t, proj)
	installed := filepath.Join(proj, ".claude", "skills", "demo-skill")

	// 启用：插件的技能应当补进"已经部署过技能"的平台。
	if code := testkit.Silent(t, func() int {
		cmd := newPluginEnableCmd()
		cmd.SetArgs([]string{"com.example.p1"})
		if err := cmd.Execute(); err != nil {
			t.Errorf("enable: %v", err)
			return 1
		}
		return 0
	}); code != 0 {
		t.Fatalf("tt plugin enable 退出码 = %d", code)
	}
	if _, err := os.Stat(filepath.Join(installed, "SKILL.md")); err != nil {
		t.Fatalf("启用后插件技能没补进平台目录: %v", err)
	}

	// 停用：它的技能撤走，官方技能留着。
	if code := testkit.Silent(t, func() int {
		cmd := newPluginDisableCmd()
		cmd.SetArgs([]string{"com.example.p1"})
		if err := cmd.Execute(); err != nil {
			t.Errorf("disable: %v", err)
			return 1
		}
		return 0
	}); code != 0 {
		t.Fatalf("tt plugin disable 退出码 = %d", code)
	}
	if _, err := os.Stat(installed); !os.IsNotExist(err) {
		t.Errorf("停用后插件技能该被撤走（err=%v）", err)
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude", "skills", "alpha", "SKILL.md")); err != nil {
		t.Errorf("官方技能不该受影响: %v", err)
	}
}

// runInstallSkills 跑 `tt install skills` 并捕获 stdout（JSON 形态）。
func runInstallSkills(t *testing.T, args ...string) (int, string) {
	t.Helper()
	oldFormat, oldHuman, oldJSON := common.Format, common.Human, common.JSON
	common.Format, common.Human, common.JSON = "json", false, false
	t.Cleanup(func() { common.Format, common.Human, common.JSON = oldFormat, oldHuman, oldJSON })
	return testkit.CaptureStdout(t, func() int {
		cmd := newInstallSkillsCmd()
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			return 1
		}
		return 0
	})
}

// readFile 读一个文本文件（失败即终止用例）。
func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("读 %s: %v", p, err)
	}
	return string(b)
}
