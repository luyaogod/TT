package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// mkSkills 在 dir 下建一个技能源树：每个技能一个目录 + SKILL.md。
func mkSkills(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		p := filepath.Join(dir, n)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p, "SKILL.md"),
			[]byte("---\nname: "+n+"\n---\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// 合并前 TDebug 与 TDictCli 各有一份等价测试。三份 install 命令合成一个
// `tt install` 后，那三份测试随旧命令一起被丢弃 —— 这里把语义补回来，
// 否则合并后的安装命令一个测试都没有。
func TestListSkills(t *testing.T) {
	src := t.TempDir()
	mkSkills(t, src, "beta", "alpha")

	got, err := listSkills(src)
	if err != nil {
		t.Fatalf("listSkills: %v", err)
	}
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Fatalf("listSkills = %v, want [alpha beta]（按名字排序）", got)
	}

	// 技能目录缺 SKILL.md 必须报错 —— Claude 技能规范要求每个技能目录里有它，
	// 静默装过去只会得到一个不会被加载的目录。
	if err := os.MkdirAll(filepath.Join(src, "broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := listSkills(src); err == nil {
		t.Fatal("技能目录缺 SKILL.md 时应报错")
	}

	// 空源目录必须报错
	if _, err := listSkills(t.TempDir()); err == nil {
		t.Fatal("skills 源目录为空时应报错")
	}
}

func TestInstallSkillsTreeCopies(t *testing.T) {
	src := t.TempDir()
	mkSkills(t, src, "alpha", "beta")
	dst := filepath.Join(t.TempDir(), "skills")

	copied, err := installSkillsTree(src, dst, false)
	if err != nil {
		t.Fatalf("installSkillsTree: %v", err)
	}
	if len(copied) != 2 {
		t.Fatalf("copied = %v, want 2 项", copied)
	}
	// 装完必须是「技能名/SKILL.md」形态，不是扁平的一堆 .md
	for _, n := range []string{"alpha", "beta"} {
		if _, err := os.Stat(filepath.Join(dst, n, "SKILL.md")); err != nil {
			t.Fatalf("%s/SKILL.md 不存在: %v", n, err)
		}
	}
}

// 目标已有同名技能目录时默认拒绝，--force 才覆盖。
// 这条是"手滑把别人的技能目录冲掉"的唯一防线。
func TestInstallSkillsTreeConflict(t *testing.T) {
	src := t.TempDir()
	mkSkills(t, src, "alpha")
	dst := t.TempDir()
	mkSkills(t, dst, "alpha")

	if _, err := installSkillsTree(src, dst, false); err == nil {
		t.Fatal("目标已存在同名技能目录时，不加 --force 应拒绝")
	}
	if _, err := installSkillsTree(src, dst, true); err != nil {
		t.Fatalf("--force 应覆盖: %v", err)
	}
}

// --force 只该删「名字来自源树」的那一层，目标目录里的其它内容必须留着。
func TestInstallSkillsTreeForceKeepsOthers(t *testing.T) {
	src := t.TempDir()
	mkSkills(t, src, "alpha")
	dst := t.TempDir()
	mkSkills(t, dst, "alpha")
	mkSkills(t, dst, "别人的技能")

	if _, err := installSkillsTree(src, dst, true); err != nil {
		t.Fatalf("--force: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "别人的技能", "SKILL.md")); err != nil {
		t.Errorf("--force 误删了不在源树里的技能目录: %v", err)
	}
}

func TestInstallSkillsTreeRejectsSameDir(t *testing.T) {
	src := t.TempDir()
	mkSkills(t, src, "alpha")
	if _, err := installSkillsTree(src, src, true); err == nil {
		t.Fatal("目标与源相同应报错（而不是自我覆盖）")
	}
}

// 源目录不是技能树（比如后端开发态 exe 旁边还没有 skills/）时要给出可操作的提示，
// 而不是一句"文件不存在"。
func TestListSkills_NotADirectory(t *testing.T) {
	if _, err := listSkills(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("源目录不存在时应报错")
	}
}
