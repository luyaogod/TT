// `tt install` —— 把 tt 装进使用者的环境。
//
//	tt install skills [--to <dir>] [--force]   把 exe 同目录的 skills/ 复制到 <当前目录>/skills
//	tt install path   [--dry-run]              把 exe 所在目录追加到**用户** PATH(HKCU，不需要管理员)
//
// 合并前 TDebug / TDev / TDictCli 各有一份同形实现，注释里写着"命令形态与另外两个一致
// （改动请三边同步）"。合并后只有这一份。
//
// 设计取舍：
//   - skills **不进二进制**（不用 go:embed）：技能内容是文档，随包发布、可被人直接改；
//     每个技能是一个带 SKILL.md 的目录。合并后 skills/ 下同时放 tdebug-debug、tdev、
//     tdict、erp-code-reader 四套技能，本命令整棵树一起装。
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

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/pathinstall"
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

  tt install skills [--to <dir>] [--force]   把 exe 同目录的 skills/ 复制到 <当前目录>/skills
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
		force bool
	)
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "把 exe 同目录的 skills/ 复制到目标目录",
		Long: `把与 tt 可执行文件同目录的 skills/ 复制到目标目录（默认 <当前目录>/skills）。

skills/ 是普通可编辑的 markdown 文件（不内嵌二进制），每个技能是一个带 SKILL.md 的目录。
合并后这里同时装着四套技能：tdebug-debug（调试）、tdev（设计器包）、tdict 与 erp-code-reader（数据字典）。

要让 Claude Code 直接加载，装到它读的位置：

  tt install skills --to .claude/skills

目标已存在同名技能目录时默认拒绝（列出冲突）；确认要刷新再加 --force
（只删目标里名字来自源树的那些技能目录，不碰其它内容）。`,
		Example: `  tt install skills
  tt install skills --to .claude/skills
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
			dst := to
			if dst == "" {
				cwd, cerr := os.Getwd()
				if cerr != nil {
					return fmt.Errorf("取不到当前目录: %w", cerr)
				}
				dst = filepath.Join(cwd, skillsDirName)
			}
			copied, err := installSkillsTree(src, dst, force)
			if err != nil {
				return err
			}
			if common.JSON {
				return common.PrintJSON(map[string]any{"ok": true, "source": src, "target": dst, "files": copied})
			}
			fmt.Printf("技能安装完成: %d 个文件\n  来源: %s\n  目标: %s\n", len(copied), src, dst)
			for _, f := range copied {
				fmt.Printf("    %s\n", f)
			}
			fmt.Println("提示: 要让 Claude Code 直接加载，用 --to .claude/skills")
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "目标目录（默认 <当前目录>/skills）")
	cmd.Flags().BoolVar(&force, "force", false, "覆盖已存在的同名技能目录")
	return cmd
}

// listSkills 列出源目录下的技能目录（按名字排序），并校验每个都带 SKILL.md。
func listSkills(src string) ([]string, error) {
	ents, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("读不到 skills 源目录 %s: %w", src, err)
	}
	var names []string
	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}
		if _, serr := os.Stat(filepath.Join(src, ent.Name(), "SKILL.md")); serr != nil {
			return nil, fmt.Errorf("技能目录缺少 SKILL.md: %s\n"+
				"Claude 技能规范要求每个技能目录里有 SKILL.md（frontmatter 的 name 必须等于目录名）", ent.Name())
		}
		names = append(names, ent.Name())
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("skills 源目录里没有任何技能: %s", src)
	}
	sort.Strings(names)
	return names, nil
}

// installSkillsTree 把 src 这棵技能树复制进 dst（合并式），返回复制到的文件路径（相对 dst，/ 分隔）。
// 已存在同名技能目录时：默认拒绝（列出冲突），--force 时只删该技能自己的目录再复制。
func installSkillsTree(src, dst string, force bool) ([]string, error) {
	names, err := listSkills(src)
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
			return nil, fmt.Errorf("目标已存在 %d 个同名技能目录，拒绝覆盖: %s\n确认要刷新请加 --force",
				len(conflicts), strings.Join(shown, ", "))
		}
		// --force：只删「目标根目录下、名字来自源树」的那一层，绝不递归删别的东西
		for _, n := range names {
			p := filepath.Join(absDst, n)
			if filepath.Dir(p) != absDst || !isDir(p) {
				continue
			}
			if err := os.RemoveAll(p); err != nil {
				return nil, fmt.Errorf("删除旧技能目录失败 %s: %w", p, err)
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
