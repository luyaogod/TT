package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/testkit"
)

// mkSkills 在 dir 下建一个技能源树：每个技能一个目录 + SKILL.md。
// 写的是**合规**的 frontmatter（name 等于目录名、description 非空），
// 因为 listSkills 现在会校验这两项。
func mkSkills(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		p := filepath.Join(dir, n)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		writeSkill(t, dir, n, "---\nname: "+n+"\ndescription: 测试技能。\n---\n")
	}
}

// writeSkill 往 <dir>/<name>/SKILL.md 写任意内容（建目录），用于造不合规的样子。
func writeSkill(t *testing.T, dir, name, body string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
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

	// 技能目录缺 SKILL.md 必须报错 —— Agent Skills 规范要求每个技能目录里有它，
	// 装过去只会得到一个不会被加载的目录。
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

// 再装一次必须还是好的：根级文件（README.md）也要算冲突。
//
// 为什么值得一条：os.CopyFS 用 O_EXCL，而 --force 从前只删「技能目录」那一层 ——
// 于是第二次 --force 会撞在根级的 README.md 上，报一句 "The file exists"。
func TestInstallSkillsTreeForceIsIdempotent(t *testing.T) {
	src := t.TempDir()
	mkSkills(t, src, "alpha")
	readme := filepath.Join(src, "README.md")
	if err := os.WriteFile(readme, []byte("索引\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "skills")

	for i := 1; i <= 2; i++ {
		if _, err := installSkillsTree(src, dst, true); err != nil {
			t.Fatalf("第 %d 次 --force 安装失败: %v", i, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dst, "README.md")); err != nil {
		t.Errorf("根级 README.md 没复制过去: %v", err)
	}

	// 不加 --force 的第二次必须被闸门拦住（而不是靠 O_EXCL 报一个裸 OS 错误），
	// 且拒绝时不能动目标里的东西。
	if err := os.WriteFile(filepath.Join(dst, "README.md"), []byte("用户改过的\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := installSkillsTree(src, dst, false); err == nil {
		t.Error("目标已有同名条目时，不加 --force 应拒绝")
	}
	got, err := os.ReadFile(filepath.Join(dst, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "用户改过的\n" {
		t.Errorf("拒绝时不该动目标文件，实得 %q", got)
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

// 不合规的 frontmatter 必须当场拒绝，而不是装出去。
//
// 为什么值得一条：name 与 description 是别的 agent 拿来当索引的两个字段（name 决定技能
// 身份、description 决定何时加载）。装过去不会报错 —— 失败都在我们看不到的地方：
// description 缺失时客户端直接跳过，name 不匹配则各家宽容度不一。
func TestListSkillsRejectsBadFrontmatter(t *testing.T) {
	long := strings.Repeat("长", skillDescriptionMax+1)
	cases := []struct {
		name string // 技能目录名
		body string // SKILL.md 内容
		want string // 报错里必须出现的片段
	}{
		{"alpha", "---\nname: beta\ndescription: x\n---\n", "必须等于目录名"},
		{"alpha", "---\nname: Alpha\ndescription: x\n---\n", "只允许小写字母"},
		{"_alpha", "---\nname: _alpha\ndescription: x\n---\n", "只允许小写字母"},
		{"alpha", "---\nname: alpha-\ndescription: x\n---\n", "只允许小写字母"},
		{"alpha", "---\ndescription: x\n---\n", "没有 name"},
		{"alpha", "---\nname: alpha\n---\n", "description 是空的"},
		{"alpha", "---\nname: alpha\ndescription: \"\"\n---\n", "description 是空的"},
		{"alpha", "---\nname: alpha\ndescription: " + long + "\n---\n", "超过规范上限"},
		{"alpha", "name: alpha\ndescription: x\n", "没有 YAML frontmatter"},
		{"alpha", "---\nname: alpha\ndescription: x\n", "没有闭合"},
	}
	for _, c := range cases {
		src := t.TempDir()
		writeSkill(t, src, c.name, c.body)
		_, err := listSkills(src)
		if err == nil {
			t.Errorf("%s: 不合规的 frontmatter 应被拒绝", c.body)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: 报错里应有 %q，实得:\n%v", c.body, c.want, err)
		}
	}
}

// frontmatter 的几种合法写法都要认，尤其是：缩进行属于 metadata 嵌套块，
// 不能把它底下的 name/description 当成技能自己的。
func TestListSkillsAcceptsFrontmatterShapes(t *testing.T) {
	src := t.TempDir()
	writeSkill(t, src, "quoted", "---\nname: \"quoted\"\ndescription: '带引号的说明'\n---\n")
	writeSkill(t, src, "block", "---\nname: block\ndescription: |\n  第一行\n  第二行\n---\n")
	writeSkill(t, src, "nested", "---\nname: nested\nmetadata:\n  name: 不算\n  description: 也不算\ndescription: 正经说明\n---\n")

	got, err := listSkills(src)
	if err != nil {
		t.Fatalf("这几种写法都合规，不该报错: %v", err)
	}
	if len(got) != 3 || got[0] != "block" || got[1] != "nested" || got[2] != "quoted" {
		t.Fatalf("listSkills = %v, want [block nested quoted]", got)
	}

	// 嵌套块里的 name/description 不能顶替顶层那两个
	b, err := os.ReadFile(filepath.Join(src, "nested", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := parseSkillFrontmatter(b)
	if err != nil {
		t.Fatalf("parseSkillFrontmatter: %v", err)
	}
	if m.Name != "nested" || m.Description != "正经说明" {
		t.Errorf("解析出 %+v，want {nested 正经说明}", m)
	}
}

// --to auto：当前目录已经为某个 agent 初始化过就别另起一个。
func TestResolveAutoTarget(t *testing.T) {
	cases := []struct {
		name string
		dirs []string // 预先创建的 agent 项目目录
		want string   // 期望落点（相对 cwd）
	}{
		{"都没有 → 跨客户端公约数", nil, filepath.Join(".agents", "skills")},
		{"只有 .claude → 用它", []string{".claude"}, filepath.Join(".claude", "skills")},
		{"只有 .cursor → 用它", []string{".cursor"}, filepath.Join(".cursor", "skills")},
		{"两个都有 → .agents 优先（读它的 agent 最多）", []string{".claude", ".agents"}, filepath.Join(".agents", "skills")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cwd := t.TempDir()
			for _, d := range c.dirs {
				if err := os.MkdirAll(filepath.Join(cwd, d), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			got, why, err := resolveAutoTarget(cwd)
			if err != nil {
				t.Fatalf("resolveAutoTarget: %v", err)
			}
			if want := filepath.Join(cwd, c.want); got != want {
				t.Errorf("落点 = %s, want %s", got, want)
			}
			if why == "" {
				t.Error("自动挑的落点必须给出理由 —— 用户得知道为什么装这儿")
			}
		})
	}
}

// --agent 的名字表与 --to/--agent 的互斥。
func TestResolveSkillsTargetByAgent(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, name := range []string{"agents", "universal", "claude-code", "claude", "copilot", "vscode", "cursor"} {
		dst, why, err := resolveSkillsTarget("", name)
		if err != nil {
			t.Fatalf("--agent %s: %v", name, err)
		}
		if !filepath.IsAbs(dst) {
			t.Errorf("--agent %s 的落点不是绝对路径: %s", name, dst)
		}
		if why == "" {
			t.Errorf("--agent %s 没给出「谁会读它」", name)
		}
	}
	if _, _, err := resolveSkillsTarget("", "nope"); err == nil {
		t.Error("不认识的 agent 名应报错（而不是静默装到某个默认位置）")
	}
	if _, _, err := resolveSkillsTarget(".x", "cursor"); err == nil {
		t.Error("--to 与 --agent 同时给应报错")
	}
	// 不给就是原来的默认：<当前目录>/skills
	dst, why, err := resolveSkillsTarget("", "")
	if err != nil {
		t.Fatalf("默认落点: %v", err)
	}
	cwd, _ := os.Getwd()
	if dst != filepath.Join(cwd, skillsDirName) || why != "" {
		t.Errorf("默认落点 = %q（why=%q）, want %q（无解释）", dst, why, filepath.Join(cwd, skillsDirName))
	}
}

// 仓库自带的 skills/ 树必须永远是合规的。
//
// 它是发行物（便携包与安装包都带着整棵），而这条断言是它唯一的门：不合规的技能
// 装出去不会报错，失败全在我们看不到的地方。
func TestRepoTreSkillsAreValid(t *testing.T) {
	src := filepath.Join(testkit.RepoRoot(t), "skills")
	names, err := listSkills(src)
	if err != nil {
		t.Fatalf("仓库自带的 skills/ 不合规：%v", err)
	}
	have := map[string]bool{}
	for _, n := range names {
		have[n] = true
	}
	for _, want := range []string{"tt-debug", "tt-dev-tzc", "tt-dev-tzs", "tt-dict", "tt-drawio", "tt-erp-read"} {
		if !have[want] {
			t.Errorf("skills/ 里没有 %s —— 六套技能是发行契约，缺一套就是少了对外面", want)
		}
	}
}

// 命令面的接线：--to auto 要真的落到探测出来的位置，而不只是 resolveAutoTarget 单测过。
func TestInstallSkillsCmdToAuto(t *testing.T) {
	src := t.TempDir()
	mkSkills(t, src, "alpha")
	cwd := t.TempDir()
	if err := os.MkdirAll(filepath.Join(cwd, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(cwd)

	old := skillsSource
	skillsSource = func() (string, error) { return src, nil }
	defer func() { skillsSource = old }()

	code := testkit.Silent(t, func() int {
		cmd := newInstallSkillsCmd()
		cmd.SetArgs([]string{"--to", "auto"})
		if err := cmd.Execute(); err != nil {
			t.Errorf("--to auto: %v", err)
			return 1
		}
		return 0
	})
	if code != 0 {
		t.Fatalf("退出码 = %d, want 0", code)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".claude", "skills", "alpha", "SKILL.md")); err != nil {
		t.Fatalf("--to auto 没有落到已存在的 .claude/skills: %v", err)
	}
}
