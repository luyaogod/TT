// `tt dev install` —— 把 tt 装进使用者的环境。
//
//	tt dev install skills   把 **exe 旁边的** skills/ 复制到 <当前目录>/skills
//	tt dev install path     把 exe 所在目录追加到**用户** PATH（HKCU，不需要管理员）
//
// 设计取舍：
//   - skills **不进二进制**（不用 go:embed）：技能内容是文档，随包发布、可被人直接改；
//     exe 只负责复制 exe 旁边的那个目录。免安装包因此是「exe + README + skills/」。
//   - PATH 只动 HKCU\Environment，绝不碰 HKLM/系统 PATH；也不用 setx
//     （setx 会把 %VAR% 展开并把长 PATH 截断到 1024 字符）。
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tt/internal/dev/store"
)

const installUsage = `tt dev install —— 把 tt 装进你的环境

用法：
  tt dev install skills [--to <dir>] [--force] [--json]
        把 exe 旁边的 skills/ 复制到 <当前目录>/skills（--to 换目标；已存在则拒绝，--force 覆盖）
  tt dev install path [--dry-run] [--json]
        把 tt.exe 所在目录追加到**用户** PATH（HKCU\Environment\Path，不需要管理员）

说明：skills 不内嵌在 exe 里，而是与 exe 同目录的 skills/（免安装包里自带）。
`

// skillsSource 解析「exe 旁边的 skills/」；测试可替换。
var skillsSource = defaultSkillsSource

func defaultSkillsSource() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", &store.IOError{Msg: "取不到当前可执行文件路径", Err: err}
	}
	return filepath.Join(filepath.Dir(exe), "skills"), nil
}

func cmdInstall(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, installUsage)
		return 2
	}
	switch args[0] {
	case "skills":
		return cmdInstallSkills(args[1:])
	case "path":
		return cmdInstallPath(args[1:])
	case "-h", "--help", "help":
		fmt.Print(installUsage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "未知子命令 %q\n\n%s", args[0], installUsage)
		return 2
	}
}

//---------------------------------------------------------------------------
// install skills
//---------------------------------------------------------------------------

func cmdInstallSkills(args []string) int {
	fs := flag.NewFlagSet("install skills", flag.ContinueOnError)
	to := fs.String("to", "", "目标目录（默认 <当前目录>/skills）")
	force := fs.Bool("force", false, "覆盖已存在的同名技能目录")
	asJSON := fs.Bool("json", false, "输出 JSON")
	if err := parseArgs(fs, args, "to"); err != nil {
		return 2
	}
	w := os.Stdout

	src, err := skillsSource()
	if err != nil {
		return fail(err, *asJSON)
	}
	if st, serr := os.Stat(src); serr != nil || !st.IsDir() {
		return fail(&store.IOError{
			Msg: "找不到 skills 目录：" + src,
			Err: errors.New("skills 与 tt.exe 同目录（免安装包里自带）；从源码运行时请用仓库根目录里的 tt.exe"),
		}, *asJSON)
	}
	dst := *to
	if dst == "" {
		cwd, cerr := os.Getwd()
		if cerr != nil {
			return fail(&store.IOError{Msg: "取不到当前目录", Err: cerr}, *asJSON)
		}
		dst = filepath.Join(cwd, "skills")
	}
	copied, err := installSkillsTree(src, dst, *force)
	if err != nil {
		return fail(err, *asJSON)
	}
	if *asJSON {
		emitJSON(w, map[string]any{"ok": true, "source": src, "target": dst, "files": copied})
		return 0
	}
	line(w, "已安装 skills（只复制文件，不改任何配置）")
	line(w, "  来源：%s", src)
	line(w, "  目标：%s", dst)
	for _, f := range copied {
		line(w, "    %s", f)
	}
	line(w, "下一步：让 AI 客户端把该目录当作技能目录加载（Claude Code 项目级可用 --to .claude/skills）")
	return 0
}

// listSkills 列出源目录下的技能目录（按名字排序），并校验每个都带 SKILL.md。
func listSkills(src string) ([]string, error) {
	ents, err := os.ReadDir(src)
	if err != nil {
		return nil, &store.IOError{Msg: "读不到 skills 源目录 " + src, Err: err}
	}
	var names []string
	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}
		if _, serr := os.Stat(filepath.Join(src, ent.Name(), "SKILL.md")); serr != nil {
			return nil, &store.IOError{
				Msg: "技能目录缺少 SKILL.md：" + ent.Name(),
				Err: errors.New("Claude 技能规范要求每个技能目录里有 SKILL.md（frontmatter 的 name 必须等于目录名）"),
			}
		}
		names = append(names, ent.Name())
	}
	if len(names) == 0 {
		return nil, &store.IOError{Msg: "skills 源目录里没有任何技能：" + src}
	}
	sort.Strings(names)
	return names, nil
}

// installSkillsTree 把 src 这棵技能树复制进 dst（合并式）。
//
// 返回复制到的文件路径（相对 dst，用 / 分隔），供报告用。
// 已存在同名技能目录时：默认拒绝（列出冲突），--force 时只删该技能自己的目录再复制。
func installSkillsTree(src, dst string, force bool) ([]string, error) {
	names, err := listSkills(src)
	if err != nil {
		return nil, err
	}
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return nil, &store.IOError{Msg: "解析源目录失败 " + src, Err: err}
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return nil, &store.IOError{Msg: "解析目标目录失败 " + dst, Err: err}
	}
	if absSrc == absDst {
		return nil, &store.IOError{
			Msg: "目标目录与源目录相同：" + absDst,
			Err: errors.New("skills 已经在这里了；要装到别处请用 --to <dir>"),
		}
	}

	var conflicts []string
	for _, n := range names {
		if p := filepath.Join(absDst, n); isDir(p) {
			conflicts = append(conflicts, filepath.Join(dst, n))
		}
	}
	if len(conflicts) > 0 {
		if !force {
			shown := conflicts
			if len(shown) > 5 {
				shown = shown[:5]
			}
			return nil, &store.IOError{
				Msg: fmt.Sprintf("目标已存在 %d 个同名技能目录，拒绝覆盖", len(conflicts)),
				Err: errors.New(strings.Join(shown, ", ") + "；确认要刷新请加 --force"),
			}
		}
		// --force：只删「目标根目录下、名字来自源树」的那一层，绝不递归删别的东西
		for _, n := range names {
			p := filepath.Join(absDst, n)
			if filepath.Dir(p) != absDst || !isDir(p) {
				continue
			}
			if err := os.RemoveAll(p); err != nil {
				return nil, &store.IOError{Msg: "删除旧技能目录失败 " + p, Err: err}
			}
		}
	}

	if err := os.MkdirAll(absDst, 0o755); err != nil {
		return nil, &store.IOError{Msg: "建目标目录失败 " + absDst, Err: err}
	}
	// os.CopyFS：纯 stdlib 复制（目标文件用 O_EXCL，所以冲突必须在上一步处理掉）
	if err := os.CopyFS(absDst, os.DirFS(absSrc)); err != nil {
		return nil, &store.IOError{Msg: "复制 skills 失败", Err: err}
	}
	var copied []string
	if err := fs.WalkDir(os.DirFS(absDst), ".", func(p string, d fs.DirEntry, werr error) error {
		if werr != nil || d.IsDir() {
			return werr
		}
		copied = append(copied, p)
		return nil
	}); err != nil {
		return nil, &store.IOError{Msg: "列复制结果失败", Err: err}
	}
	sort.Strings(copied)
	return copied, nil
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

//---------------------------------------------------------------------------
// install path
//---------------------------------------------------------------------------

// normalizePathEntry 归一化 PATH 里的一段，用于「已存在」判定：
// 去空白、去结尾的反斜杠/斜杠、大小写不敏感（Windows 路径）。
func normalizePathEntry(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimRight(p, `\/`)
	return strings.ToLower(p)
}

// mergeUserPath 把 dir 追加到用户 PATH 原值后面；已在其中则原样返回（added=false）。
//
// 不展开、不改写原有内容（`%USERPROFILE%` 之类的展开式变量必须原样保留），
// 也不重排顺序 —— 只做「去尾分号 + 追加」。
func mergeUserPath(old, dir string) (string, bool) {
	want := normalizePathEntry(dir)
	if want != "" {
		for _, p := range strings.Split(old, ";") {
			if normalizePathEntry(p) == want {
				return old, false
			}
		}
	}
	base := strings.TrimRight(old, ";")
	if strings.TrimSpace(base) == "" {
		return dir, true
	}
	return base + ";" + dir, true
}

func cmdInstallPath(args []string) int {
	fs := flag.NewFlagSet("install path", flag.ContinueOnError)
	dry := fs.Bool("dry-run", false, "只打印将要写入的内容，不碰注册表")
	asJSON := fs.Bool("json", false, "输出 JSON")
	if err := parseArgs(fs, args); err != nil {
		return 2
	}
	w := os.Stdout

	exe, err := os.Executable()
	if err != nil {
		return fail(&store.IOError{Msg: "取不到当前可执行文件路径", Err: err}, *asJSON)
	}
	dir := filepath.Dir(exe)

	cur, typ, err := userPath()
	if err != nil {
		return fail(err, *asJSON)
	}
	next, added := mergeUserPath(cur, dir)

	if *dry {
		if *asJSON {
			emitJSON(w, map[string]any{"ok": true, "dry_run": true, "exe_dir": dir,
				"added": added, "old": cur, "new": next, "reg_type": typ})
			return 0
		}
		line(w, "（--dry-run，未写注册表）")
		line(w, "  exe 目录：%s", dir)
		line(w, "  注册表值：HKCU\\Environment\\Path（%s）", typ)
		if added {
			line(w, "  现在（原值）：%s", cur)
			line(w, "  将要写入　　：%s", next)
		} else {
			line(w, "  用户 PATH 里已经有该目录，无需改动")
		}
		warnPathRisks(w, dir, next)
		return 0
	}

	if !added {
		if *asJSON {
			emitJSON(w, map[string]any{"ok": true, "changed": false, "exe_dir": dir, "path": cur})
			return 0
		}
		line(w, "用户 PATH 里已经有 %s，无需改动", dir)
		return 0
	}
	warnPathRisks(w, dir, next)
	if err := setUserPath(next, typ); err != nil {
		return fail(err, *asJSON)
	}
	envWarn := ""
	if berr := broadcastEnvChange(); berr != nil {
		envWarn = berr.Error()
	}
	if *asJSON {
		emitJSON(w, map[string]any{"ok": true, "changed": true, "exe_dir": dir,
			"path": next, "broadcast_error": envWarn})
		return 0
	}
	line(w, "已把 %s 追加到用户 PATH（HKCU\\Environment\\Path，不需要管理员）", dir)
	line(w, "  新值：%s", next)
	if envWarn != "" {
		line(w, "  提醒：环境变更广播失败（%s）；新开的终端仍会读到新值", envWarn)
	} else {
		line(w, "  新开的终端即可直接敲 `tt`（当前已开的终端需要重开）")
	}
	return 0
}

// warnPathRisks 打印两类「装完也会失效」的提醒（不阻断）。
func warnPathRisks(w io.Writer, dir, next string) {
	low := strings.ToLower(dir)
	tmp := strings.ToLower(os.TempDir())
	if tmp != "" && strings.HasPrefix(low, tmp) {
		line(w, "  警告：exe 在临时目录 %s 下，清理临时文件后这个 PATH 项会失效", dir)
	}
	if len(next) > 2048 {
		line(w, "  警告：新 PATH 长 %d 字符，偏长；建议清理后再装", len(next))
	}
}
