package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeSkills 造一棵假的技能树，返回源目录。
func fakeSkills(t *testing.T, skills ...string) string {
	t.Helper()
	src := filepath.Join(t.TempDir(), "skills")
	for _, n := range skills {
		dir := filepath.Join(src, n)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		body := "---\nname: " + n + "\ndescription: 测试用技能\n---\n\n正文\n"
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return src
}

// TestInstallSkillsTree 覆盖：全新复制 / 冲突拒绝 / --force 刷新 / 源=目标拒绝。
func TestInstallSkillsTree(t *testing.T) {
	src := fakeSkills(t, "tdev")
	dst := filepath.Join(t.TempDir(), "skills")

	// ① 全新复制
	copied, err := installSkillsTree(src, dst, false)
	if err != nil {
		t.Fatalf("全新复制失败: %v", err)
	}
	if len(copied) != 1 || copied[0] != "tdev/SKILL.md" {
		t.Fatalf("复制结果 = %v，想要 [tdev/SKILL.md]", copied)
	}
	if _, err := os.Stat(filepath.Join(dst, "tdev", "SKILL.md")); err != nil {
		t.Fatalf("目标没写出来: %v", err)
	}

	// ② 已存在 → 默认拒绝，且提示 --force
	if _, err := installSkillsTree(src, dst, false); err == nil {
		t.Fatal("已存在同名技能目录时必须拒绝（除非 --force）")
	} else if !strings.Contains(err.Error(), "--force") {
		t.Errorf("拒绝信息里应提示 --force：%v", err)
	}

	// ③ --force：内容被刷新
	changed := "---\nname: tdev\ndescription: 改过的\n---\n\n新正文\n"
	if err := os.WriteFile(filepath.Join(dst, "tdev", "SKILL.md"), []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := installSkillsTree(src, dst, true); err != nil {
		t.Fatalf("--force 刷新失败: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "tdev", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "改过的") {
		t.Error("--force 应把旧内容覆盖掉")
	}

	// ④ 源 = 目标 → 拒绝
	if _, err := installSkillsTree(src, src, true); err == nil {
		t.Fatal("源与目标相同时必须拒绝")
	}
}

// TestInstallSkillsTreeValidation 源目录不合法时要明确报错（缺 SKILL.md / 空目录）。
func TestInstallSkillsTreeValidation(t *testing.T) {
	// 空源目录
	empty := filepath.Join(t.TempDir(), "skills")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := installSkillsTree(empty, filepath.Join(t.TempDir(), "out"), false); err == nil {
		t.Error("空源目录应报错")
	}
	// 有目录但没有 SKILL.md
	bad := filepath.Join(t.TempDir(), "skills", "tdev")
	if err := os.MkdirAll(bad, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := installSkillsTree(filepath.Dir(bad), filepath.Join(t.TempDir(), "out"), false); err == nil {
		t.Error("缺 SKILL.md 的技能目录应报错")
	} else if !strings.Contains(err.Error(), "SKILL.md") {
		t.Errorf("错误信息里应提到 SKILL.md：%v", err)
	}
}

// TestCmdInstallSkills 端到端：当前目录下生成 skills/tdev/SKILL.md（源用假技能树注入）。
func TestCmdInstallSkills(t *testing.T) {
	src := fakeSkills(t, "tdev")
	old := skillsSource
	skillsSource = func() (string, error) { return src, nil }
	t.Cleanup(func() { skillsSource = old })

	cwd := t.TempDir()
	t.Chdir(cwd)
	if code := silent(t, func() int { return cmdInstall([]string{"skills"}) }); code != 0 {
		t.Fatalf("install skills 退出码 %d", code)
	}
	// 目标就是「执行命令时的当前目录/skills」
	if _, err := os.Stat(filepath.Join(cwd, "skills", "tdev", "SKILL.md")); err != nil {
		t.Fatalf("没有装到 <cwd>/skills：%v", err)
	}
	// 再装一次 → 拒绝（非零退出）
	if code := silent(t, func() int { return cmdInstall([]string{"skills"}) }); code == 0 {
		t.Error("重复安装应被拒绝（除非 --force）")
	}
	// --force → 成功
	if code := silent(t, func() int { return cmdInstall([]string{"skills", "--force"}) }); code != 0 {
		t.Errorf("install skills --force 退出码 %d", code)
	}
	// --to 指定别处
	other := t.TempDir()
	if code := silent(t, func() int { return cmdInstall([]string{"skills", "--to", other}) }); code != 0 {
		t.Errorf("install skills --to 退出码 %d", code)
	}
	if _, err := os.Stat(filepath.Join(other, "tdev", "SKILL.md")); err != nil {
		t.Errorf("--to 目标没写出来：%v", err)
	}
	// 未知子命令 → 退出码 2
	if code := silent(t, func() int { return cmdInstall([]string{"nope"}) }); code != 2 {
		t.Errorf("未知子命令退出码 %d，想要 2", code)
	}
}

// TestMergeUserPath 钉住用户 PATH 的合并语义：追加、幂等、不改写原值。
func TestMergeUserPath(t *testing.T) {
	cases := []struct {
		old     string
		dir     string
		want    string
		added   bool
		comment string
	}{
		{"C:\\a;D:\\b", `E:\tdev`, `C:\a;D:\b;E:\tdev`, true, "普通追加"},
		{"C:\\a;", `E:\tdev`, `C:\a;E:\tdev`, true, "去掉尾部空段再追加"},
		{"", `E:\tdev`, `E:\tdev`, true, "原本为空"},
		{";;", `E:\tdev`, `E:\tdev`, true, "只有分隔符"},
		{`C:\a;E:\tdev`, `E:\tdev`, `C:\a;E:\tdev`, false, "已存在（幂等）"},
		{`C:\a;E:\TDEV\`, `e:\tdev`, `C:\a;E:\TDEV\`, false, "已存在（大小写 + 尾反斜杠等价）"},
		{`%USERPROFILE%\bin;C:\a`, `E:\tdev`, `%USERPROFILE%\bin;C:\a;E:\tdev`, true, "展开式变量原样保留"},
	}
	for _, c := range cases {
		got, added := mergeUserPath(c.old, c.dir)
		if got != c.want || added != c.added {
			t.Errorf("%s：mergeUserPath(%q, %q) = (%q, %v)，想要 (%q, %v)",
				c.comment, c.old, c.dir, got, added, c.want, c.added)
		}
	}
}
