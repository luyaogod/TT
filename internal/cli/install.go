// `tt install` —— 把 tt 装进使用者的环境。
//
//	tt install skills [--to <dir>|auto] [--agent <名字>] [--force]
//	                                           把 exe 同目录的 skills/ 复制到目标目录
//	tt install path   [--dry-run]              把 exe 所在目录追加到**用户** PATH(HKCU，不需要管理员)
//
// 合并前 TDebug / TDev / TDictCli 各有一份同形实现，注释里写着"命令形态与另外两个一致
// （改动请三边同步）"。合并后只有这一份。
//
// 设计取舍：
//   - skills **不进二进制**（不用 go:embed）：技能内容是文档，随包发布、可被人直接改；
//     每个技能是一个带 SKILL.md 的目录。合并后 skills/ 下同时放 tt-debug、tt-dev-tzc、
//     tt-dev-tzs、tt-dict、tt-drawio、tt-erp-read 六套技能，本命令整棵树一起装。
//   - PATH 只动 HKCU\Environment，绝不碰 HKLM/系统 PATH；实现在 internal/pathinstall，
//     与 Web 设置页的「加入 PATH」共用同一份。
package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/pathinstall"
	"tt/internal/update"
)

const skillsDirName = "skills"

// skillsSource 解析「exe 旁边的 skills/」；测试可替换。
var skillsSource = defaultSkillsSource

func defaultSkillsSource() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("取不到当前可执行文件路径: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), skillsDirName), nil
}

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "把 tt 装进你的环境（skills / path）",
		Long: `把 tt 装进你的环境：

  tt install skills [--to <dir>|auto] [--agent <名字>] [--force]
                                             把 exe 同目录的 skills/ 复制到目标目录
  tt install path   [--dry-run]              把 exe 所在目录追加到用户 PATH(HKCU)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("请指定子命令: skills 或 path\n\n%s", cmd.UsageString())
		},
	}
	cmd.AddCommand(newInstallSkillsCmd(), newInstallPathCmd())
	return cmd
}

//---------------------------------------------------------------------------
// tt install skills
//---------------------------------------------------------------------------

func newInstallSkillsCmd() *cobra.Command {
	var (
		to    string
		agent string
		force bool
	)
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "把 exe 同目录的 skills/ 复制到目标目录",
		Long: `把与 tt 可执行文件同目录的 skills/ 复制过去，默认落到 <当前目录>/skills。

skills/ 是普通可编辑的 markdown 文件（不内嵌二进制），每个技能是一个带 SKILL.md 的目录。
合并后这里同时装着六套技能：tt-debug（调试）、tt-dev-tzc（设计器代码包 .tzc）、
tt-dev-tzs（设计器表单包 .tzs）、tt-dict（数据字典）、tt-drawio（原型图）与
tt-erp-read（读 ERP 代码）。

要装到 agent 会去读的位置，用 --to auto 或点名一个 agent：

  tt install skills --to auto             # 已有 agent 目录就用它，否则建 .agents/skills
  tt install skills --agent claude-code   # 装到 Claude Code 读的位置
  tt install skills --agent agents        # 跨客户端公约数 .agents/skills

--agent 接受的名字：` + agentSkillNames() + `

--to auto 的探测顺序：.agents/skills 优先（读它的 agent 最多），其次 .claude/skills、
.github/skills、.cursor/skills；当前目录一个都没有就新建 .agents/skills。

目标已存在同名技能目录时默认拒绝（列出冲突）；确认要刷新再加 --force
（只删目标里名字来自源树的那些技能目录，不碰其它内容）。

安装前逐个校验技能：每个技能目录要有 SKILL.md，且里面 frontmatter 的 name 必须等于
目录名、description 非空且 ≤1024 字符。不合规的技能，失败都落在我们看不到的地方（description 缺失时
客户端直接跳过；name 不匹配各家宽容度不一），所以这里当场拒绝，而不是装一棵半成品树过去。`,
		Example: `  tt install skills --to auto
  tt install skills --agent claude-code
  tt install skills --to .agents/skills
  tt install skills --force`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := skillsSource()
			if err != nil {
				return err
			}
			if !isDir(src) {
				return fmt.Errorf("找不到 skills 目录: %s\n\nskills/ 与 tt.exe 同目录（免安装包里自带）；"+
					"从源码运行时请先构建到仓库根目录再执行", src)
			}
			dst, why, terr := resolveSkillsTarget(to, agent)
			if terr != nil {
				return terr
			}
			copied, err := installSkillsTree(src, dst, force)
			if err != nil {
				return err
			}
			// 记下"技能装到了哪、装它们的是哪个版本"。升级后靠它把 agent 目录里那份刷成
			// 与二进制同版 —— 不记的话它会永远停在装它的那个版本（见 update.SkillsStamp）。
			//
			// 没有版本号的构建（`make build` / `go test`）**不写**：戳的全部意义是"哪个
			// 版本装的"，记一个空版本只会让漂移提示说不清话，还会让测试在开发者的真实
			// 数据目录里留垃圾（它落的是 ResolveConfig 解析出的那个目录）。
			if Version != "" {
				if cfgPath, cerr := common.ResolveConfig(true); cerr == nil {
					_ = update.RecordSkillsTarget(filepath.Dir(cfgPath), dst, Version)
				}
			}
			if common.JSON {
				return common.PrintJSON(map[string]any{"ok": true, "source": src, "target": dst,
					"why": why, "files": copied})
			}
			fmt.Printf("技能安装完成: %d 个文件\n  来源: %s\n  目标: %s\n", len(copied), src, dst)
			if why != "" {
				fmt.Printf("  为什么是它: %s\n", why)
			}
			for _, f := range copied {
				fmt.Printf("    %s\n", f)
			}
			if to == "" && agent == "" {
				// 默认那个位置与 exe 旁边同形，是给人读的；会去扫它的 agent 很少
				fmt.Println("提示: 这个位置几乎没有 agent 会读；要让 agent 直接加载，用 --to auto")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "目标目录；auto = 自动挑一个 agent 会读的位置（默认 <当前目录>/skills）")
	cmd.Flags().StringVar(&agent, "agent", "", "按 agent 名落到它读的位置："+agentSkillNames())
	cmd.Flags().BoolVar(&force, "force", false, "覆盖已存在的同名技能目录")
	return cmd
}

//---------------------------------------------------------------------------
// 目标目录：--to / --agent / auto
//---------------------------------------------------------------------------

// autoSkillsTarget 是 --to 的特殊值：不是目录，而是「替我挑一个 agent 会读的位置」。
const autoSkillsTarget = "auto"

// agentSkillsTarget 是一个「agent 读技能的位置」。
type agentSkillsTarget struct {
	Name    string   // --agent 接受的名字
	Root    string   // 该 agent 在项目里的目录（= 检测用的路标，也是 skills/ 的父目录）
	Aliases []string // 顺手接受的名字
	Who     string   // 装完谁会读它（写进输出，让用户知道这个选择意味着什么）
}

// Dir 是目标目录：<agent 项目目录>/skills。
func (a agentSkillsTarget) Dir() string { return filepath.Join(a.Root, skillsDirName) }

// agentSkillsDirs 是「agent 各自读技能的位置」—— --agent 与 --to auto 的唯一依据。
//
// **表里的顺序就是 --to auto 的探测顺序**：.agents/skills 排第一，因为它是跨客户端
// 公约数（Agent Skills 规范点名的互操作路径，pi、Codex、Cursor、Gemini CLI、OpenCode、
// Copilot 都扫它），装一次伺候的人最多。
//
// 只收路径真的不同的那几个 —— 路径相同的（Codex、Cline、Zed…）没有单列的必要。
var agentSkillsDirs = []agentSkillsTarget{
	{Name: "agents", Root: ".agents", Aliases: []string{"universal"},
		Who: "跨客户端公约数：pi、Codex、Cursor、Gemini CLI、OpenCode、Copilot 都读它"},
	{Name: "claude-code", Root: ".claude", Aliases: []string{"claude"},
		Who: "Claude Code"},
	{Name: "copilot", Root: ".github", Aliases: []string{"github-copilot", "vscode"},
		Who: "GitHub Copilot / VS Code 的工作区技能"},
	{Name: "cursor", Root: ".cursor",
		Who: "Cursor 的专属目录（它也读 .agents/skills）"},
}

// agentSkillNames 是 --agent 的全部合法名字（含别名），用于帮助与报错。
func agentSkillNames() string {
	var out []string
	for _, a := range agentSkillsDirs {
		out = append(out, a.Name)
		out = append(out, a.Aliases...)
	}
	return strings.Join(out, " / ")
}

func findAgentSkillsDir(name string) (agentSkillsTarget, error) {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, a := range agentSkillsDirs {
		if a.Name == n {
			return a, nil
		}
		for _, al := range a.Aliases {
			if al == n {
				return a, nil
			}
		}
	}
	return agentSkillsTarget{}, fmt.Errorf("不认识的 agent: %s\n可选: %s（或直接用 --to <目录> 给路径）",
		name, agentSkillNames())
}

// resolveSkillsTarget 把 --to / --agent 解析成目标目录，并给出「为什么是它」。
//
// why 只在目标是自动挑出来的（--to auto / --agent）时非空：显式给了路径的人
// 不需要别人告诉他为什么要装这儿。
func resolveSkillsTarget(to, agent string) (dst, why string, err error) {
	if to != "" && agent != "" {
		return "", "", fmt.Errorf("--to 与 --agent 只能给一个：--to 直接指目录，--agent 按名字查目录")
	}
	if agent != "" {
		a, aerr := findAgentSkillsDir(agent)
		if aerr != nil {
			return "", "", aerr
		}
		cwd, cerr := os.Getwd()
		if cerr != nil {
			return "", "", fmt.Errorf("取不到当前目录: %w", cerr)
		}
		return filepath.Join(cwd, a.Dir()), a.Who + " 会读它", nil
	}
	if to != "" && to != autoSkillsTarget {
		return to, "", nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("取不到当前目录: %w", err)
	}
	if to == autoSkillsTarget {
		return resolveAutoTarget(cwd)
	}
	return filepath.Join(cwd, skillsDirName), "", nil
}

// resolveAutoTarget 挑一个落点：当前目录已经为某个 agent 初始化过就别另起一个，
// 否则建跨客户端公约数 .agents/skills。
//
// 探的是 agent 的项目目录（.claude/ 之类）而不是 skills/ 本身：那才是「这人用这个 agent」
// 的路标，而且第一次装的时候 skills/ 还不存在。
func resolveAutoTarget(cwd string) (dst, why string, err error) {
	for _, a := range agentSkillsDirs {
		if isDir(filepath.Join(cwd, a.Root)) {
			return filepath.Join(cwd, a.Dir()), "当前目录已有 " + a.Dir() + "（" + a.Who + "）", nil
		}
	}
	a, err := findAgentSkillsDir("agents")
	if err != nil {
		return "", "", err
	}
	return filepath.Join(cwd, a.Dir()), "当前目录还没有 agent 目录，新建 " + a.Dir() + "（" + a.Who + "）", nil
}

// listSkills 列出源目录下的技能目录（按名字排序），并逐个校验它是一份**合规**的技能：
// 有 SKILL.md，且里面 frontmatter 的 name 等于目录名、description 非空。
//
// 校验在这里而不是复制之后：装到一半失败会在目标目录留下一棵半成品树。
func listSkills(src string) ([]string, error) {
	ents, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("读不到 skills 源目录 %s: %w", src, err)
	}
	var names, probs []string
	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}
		b, rerr := os.ReadFile(filepath.Join(src, ent.Name(), "SKILL.md"))
		if rerr != nil {
			probs = append(probs, fmt.Sprintf("%s: 技能目录缺少 SKILL.md —— 规范要求每个技能目录里有它", ent.Name()))
			continue
		}
		meta, perr := parseSkillFrontmatter(b)
		if perr != nil {
			probs = append(probs, fmt.Sprintf("%s/SKILL.md: %v", ent.Name(), perr))
			continue
		}
		for _, p := range validateSkillMeta(ent.Name(), meta) {
			probs = append(probs, fmt.Sprintf("%s/SKILL.md: %s", ent.Name(), p))
		}
		names = append(names, ent.Name())
	}
	if len(probs) > 0 {
		return nil, fmt.Errorf("技能不合规，拒绝安装（%d 处）:\n  %s\n\n"+
			"Agent Skills 规范 → https://agentskills.io/specification（name 必须等于目录名；description 非空、≤1024 字符）",
			len(probs), strings.Join(probs, "\n  "))
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("skills 源目录里没有任何技能: %s", src)
	}
	sort.Strings(names)
	return names, nil
}

// skillMeta 是规范要求必填的两个字段 —— 也是别的 agent 拿来当索引的两个字段：
// name 决定技能身份，description 决定它何时被加载。
type skillMeta struct {
	Name        string
	Description string
}

// validateSkillMeta 按 Agent Skills 规范逐条检查，返回全部问题（一次说完，别让人改一处跑一次）。
//
// 为什么是硬失败：不合规的技能装出去**不会报错**，失败都在我们看不到的地方 ——
// description 缺失时客户端直接跳过该技能；name 不等于目录名是规范的硬约束，
// 但各家客户端的宽容度不一（有的警告后照装）。两种都不该由我们发出去。
func validateSkillMeta(dir string, m skillMeta) []string {
	var probs []string
	switch {
	case m.Name == "":
		probs = append(probs, "frontmatter 里没有 name")
	case m.Name != dir:
		probs = append(probs, fmt.Sprintf("name 是 %q，必须等于目录名 %q", m.Name, dir))
	}
	if m.Name != "" && !validSkillName(m.Name) {
		probs = append(probs, fmt.Sprintf("name %q 不合规：只允许小写字母、数字与连字符，"+
			"不能有连续或首尾连字符，最长 64 字符", m.Name))
	}
	switch n := utf8.RuneCountInString(m.Description); {
	case m.Description == "":
		probs = append(probs, "description 是空的：别的 agent 靠它决定何时加载这个技能")
	case n > skillDescriptionMax:
		probs = append(probs, fmt.Sprintf("description 有 %d 个字符，超过规范上限 %d", n, skillDescriptionMax))
	}
	return probs
}

const skillDescriptionMax = 1024

// validSkillName 是规范的 name 字符集：小写字母、数字、连字符，最长 64 字符，
// 不能以连字符开头或结尾，不能有连续连字符。
func validSkillName(s string) bool {
	if utf8.RuneCountInString(s) > 64 {
		return false
	}
	prevHyphen := true // 首字符不许是连字符
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			prevHyphen = false
		case r == '-':
			if prevHyphen {
				return false
			}
			prevHyphen = true
		default:
			return false
		}
	}
	return !prevHyphen // 末字符不许是连字符
}

// parseSkillFrontmatter 从 SKILL.md 里抠出 name 与 description。
//
// 不引 YAML 依赖：这里要判的是两个标量字符串，不是完整解析。只认**顶层**的 `键: 值` ——
// 缩进行属于 metadata 之类的嵌套块，一律跳过（否则 `metadata:` 底下的 name 会被当成技能的 name）。
// 值是 `|` / `>` 块标量时收拢后面的缩进行；两边成对的引号去掉。
func parseSkillFrontmatter(b []byte) (skillMeta, error) {
	s := strings.TrimPrefix(string(b), "\ufeff") // BOM
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return skillMeta{}, fmt.Errorf("开头没有 YAML frontmatter（规范要求第一行是 ---，里面写 name 与 description）")
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return skillMeta{}, fmt.Errorf("frontmatter 没有闭合的 ---")
	}

	body := lines[1:end]
	var m skillMeta
	for i := 0; i < len(body); i++ {
		line := body[i]
		if line == "" || line[0] == ' ' || line[0] == '\t' || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		val = strings.TrimSpace(val)
		switch strings.TrimSpace(key) {
		case "name":
			m.Name = unquoteYAML(val)
		case "description":
			if val == "|" || val == ">" || val == "|-" || val == ">-" {
				var parts []string
				for j := i + 1; j < len(body); j++ {
					l := body[j]
					if strings.TrimSpace(l) == "" {
						continue
					}
					if l[0] != ' ' && l[0] != '\t' {
						break
					}
					parts = append(parts, strings.TrimSpace(l))
					i = j
				}
				m.Description = strings.Join(parts, " ")
			} else {
				m.Description = unquoteYAML(val)
			}
		}
	}
	return m, nil
}

func unquoteYAML(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}

// topLevelEntries 返回源树顶层的全部条目名（技能目录与根级文件如 README.md 都算）。
func topLevelEntries(src string) ([]string, error) {
	ents, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("读不到 skills 源目录 %s: %w", src, err)
	}
	out := make([]string, 0, len(ents))
	for _, ent := range ents {
		out = append(out, ent.Name())
	}
	sort.Strings(out)
	return out, nil
}

// installSkillsTree 把 src 这棵技能树复制进 dst（合并式），返回复制到的文件路径（相对 dst，/ 分隔）。
// 已存在同名条目时：默认拒绝（列出冲突），--force 时只删该条目自己的那一层再复制。
func installSkillsTree(src, dst string, force bool) ([]string, error) {
	if _, err := listSkills(src); err != nil { // 只为校验；列名字用下面的 top
		return nil, err
	}
	// 顶层**全部**条目（不只技能目录）：根级文件（如 README.md）也是要复制的东西，
	// 而 os.CopyFS 用 O_EXCL —— 撞上一样失败，只是错误杧得多。
	top, err := topLevelEntries(src)
	if err != nil {
		return nil, err
	}
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return nil, fmt.Errorf("解析源目录失败 %s: %w", src, err)
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return nil, fmt.Errorf("解析目标目录失败 %s: %w", dst, err)
	}
	if strings.EqualFold(absSrc, absDst) {
		return nil, fmt.Errorf("目标目录与源目录相同: %s\nskills 已经在这里了；要装到别处请用 --to <dir>", absDst)
	}

	var conflicts []string
	for _, n := range top {
		if _, lerr := os.Lstat(filepath.Join(absDst, n)); lerr == nil {
			conflicts = append(conflicts, filepath.Join(dst, n))
		}
	}
	if len(conflicts) > 0 {
		if !force {
			shown := conflicts
			if len(shown) > 5 {
				shown = shown[:5]
			}
			return nil, fmt.Errorf("目标已存在 %d 个同名条目，拒绝覆盖: %s\n确认要刷新请加 --force",
				len(conflicts), strings.Join(shown, ", "))
		}
		// --force：只删「目标根目录下、名字来自源树」的那一层，绝不递归删别的东西
		for _, n := range top {
			p := filepath.Join(absDst, n)
			if filepath.Dir(p) != absDst {
				continue
			}
			if err := os.RemoveAll(p); err != nil {
				return nil, fmt.Errorf("删除旧条目失败 %s: %w", p, err)
			}
		}
	}

	if err := os.MkdirAll(absDst, 0o755); err != nil {
		return nil, fmt.Errorf("建目标目录失败 %s: %w", absDst, err)
	}
	// os.CopyFS：纯 stdlib 复制（目标文件用 O_EXCL，所以冲突必须在上一步处理掉）
	if err := os.CopyFS(absDst, os.DirFS(absSrc)); err != nil {
		return nil, fmt.Errorf("复制 skills 失败: %w", err)
	}
	var copied []string
	if err := fs.WalkDir(os.DirFS(absDst), ".", func(p string, d fs.DirEntry, werr error) error {
		if werr != nil || d.IsDir() {
			return werr
		}
		copied = append(copied, p)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("列复制结果失败: %w", err)
	}
	sort.Strings(copied)
	return copied, nil
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

//---------------------------------------------------------------------------
// tt install path
//---------------------------------------------------------------------------

func newInstallPathCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "path",
		Short: "把 exe 所在目录加入用户 PATH（HKCU，不需要管理员）",
		Long: `把 tt.exe 所在目录追加到**用户** PATH（HKCU\Environment\Path）。

只动当前用户的注册表项，不需要管理员，绝不碰系统 PATH。已有该项时幂等
（不重复追加、不重排顺序，原有的 %USERPROFILE% 之类可展开变量原样保留）。
与 Web 设置页「设置 → 加入 PATH」是同一份实现。`,
		Example: `  tt install path
  tt install path --dry-run`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			st, next, changed, err := pathinstall.Preview()
			if err != nil {
				return err
			}
			if !st.Supported {
				if common.JSON {
					return common.PrintJSON(st)
				}
				fmt.Printf("%s\n手动命令: %s\n", st.Note, st.Manual)
				return nil
			}

			if dryRun {
				if common.JSON {
					return common.PrintJSON(map[string]any{"ok": true, "dryRun": true, "exeDir": st.ExeDir,
						"changed": changed, "old": st.UserPath, "new": next})
				}
				fmt.Println("(--dry-run，未写注册表)")
				fmt.Printf("  exe 目录: %s\n", st.ExeDir)
				if changed {
					fmt.Printf("  现在(原值): %s\n", st.UserPath)
					fmt.Printf("  将要写入  : %s\n", next)
				} else {
					fmt.Println("  用户 PATH 里已经有该目录，无需改动")
				}
				return nil
			}

			if !changed {
				if common.JSON {
					return common.PrintJSON(map[string]any{"ok": true, "changed": false, "exeDir": st.ExeDir, "userPath": st.UserPath})
				}
				fmt.Printf("用户 PATH 里已经有 %s，无需改动\n", st.ExeDir)
				return nil
			}
			if _, err := pathinstall.Add(); err != nil {
				return err
			}
			if common.JSON {
				return common.PrintJSON(map[string]any{"ok": true, "changed": true, "exeDir": st.ExeDir, "userPath": next})
			}
			fmt.Printf("已把 %s 追加到用户 PATH（HKCU\\Environment\\Path，不需要管理员）\n", st.ExeDir)
			fmt.Printf("  新值: %s\n", next)
			fmt.Println("  新开的终端即可直接敲 `tt`（当前已开的终端需要重开）")
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只打印将要写入的内容，不碰注册表")
	return cmd
}
