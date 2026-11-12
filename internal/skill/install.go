package skill

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Options 是一次安装的行为开关。
type Options struct {
	Force  bool // 目标已存在同名条目时删掉重装（默认跳过，好让多个来源共用 .agents/skills）
	DryRun bool // 只算不写
}

// Report 是一次安装的结果。
//
// Copied 里是相对目标目录的文件路径（/ 分隔），Rewritten 是「被写入了静态文件指针」
// 的技能名 —— 那几个 SKILL.md 与源文件**故意不一致**，说出来比让人自己发现好。
// Ignored 是被刻意跳过的版本库元数据（同样相对目标目录、/ 分隔）：只记 `.git` 这个
// 条目本身，不逐个列它下面的对象文件 —— 那正是它要被跳过的原因。
type Report struct {
	Target    string    `json:"target"`
	DryRun    bool      `json:"dryRun,omitempty"`
	Copied    []string  `json:"files"`
	Ignored   []string  `json:"ignored,omitempty"`
	Skipped   []string  `json:"skipped,omitempty"`
	Rewritten []string  `json:"rewritten,omitempty"`
	Warnings  []Warning `json:"warnings,omitempty"`
}

// Apply 把计划装进 dst。
//
// 目标里已存在同名条目时默认**跳过**（追加模式：不覆盖用户已有的技能，方便多个来源
// 共用 .agents/skills）；Force 时删掉该条目再复制。只动「名字在计划里」的顶层条目，
// 目标里的其它内容一律保留。
func Apply(p Plan, dst string, opt Options) (Report, error) {
	rep := Report{DryRun: opt.DryRun, Warnings: p.Warnings}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return rep, fmt.Errorf("解析目标目录失败 %s: %w", dst, err)
	}
	rep.Target = absDst

	for _, s := range p.Skills {
		if SamePath(s.Dir, absDst) {
			return rep, fmt.Errorf("目标目录与来源目录相同: %s\nskills 已经在这里了；要装到别处请用 --to <dir>", absDst)
		}
	}
	if !opt.DryRun {
		if err := os.MkdirAll(absDst, 0o755); err != nil {
			return rep, fmt.Errorf("建目标目录失败 %s: %w", absDst, err)
		}
	}

	for _, s := range p.Skills {
		dstEntry := filepath.Join(absDst, s.Name)
		exists, err := pathExists(dstEntry)
		if err != nil {
			return rep, err
		}
		if exists {
			if !opt.Force {
				rep.Skipped = append(rep.Skipped, s.Name)
				continue
			}
			if !opt.DryRun {
				if err := os.RemoveAll(dstEntry); err != nil {
					return rep, fmt.Errorf("删除旧条目失败 %s: %w", dstEntry, err)
				}
			}
		}

		// 根级 assets/ 不复制：只留一行指针指回源目录（见 assets.go）。
		hasAssets := HasAssets(s.Dir)

		if opt.DryRun {
			files, ignored, perr := plannedFiles(s.Dir, s.Name)
			if perr != nil {
				return rep, perr
			}
			rep.Copied = append(rep.Copied, files...)
			rep.Ignored = append(rep.Ignored, ignored...)
			if hasAssets {
				rep.Rewritten = append(rep.Rewritten, s.Name)
			}
			continue
		}

		ignored, cerr := copyTree(s.Dir, dstEntry)
		if cerr != nil {
			return rep, cerr
		}
		for _, p := range ignored {
			rep.Ignored = append(rep.Ignored, s.Name+"/"+p)
		}
		if hasAssets {
			changed, ierr := rewriteAssetsPointer(dstEntry, AssetsDir(s.Dir))
			if ierr != nil {
				return rep, ierr
			}
			if changed {
				rep.Rewritten = append(rep.Rewritten, s.Name)
			}
		}
		files, ferr := filesUnder(absDst, s.Name)
		if ferr != nil {
			return rep, ferr
		}
		rep.Copied = append(rep.Copied, files...)
	}

	// 官方树顶层的普通文件（README.md 之类）照旧原样复制；第三方没有这一层。
	for _, src := range p.ExtraFiles {
		base := filepath.Base(src)
		dstEntry := filepath.Join(absDst, base)
		exists, err := pathExists(dstEntry)
		if err != nil {
			return rep, err
		}
		if exists {
			if !opt.Force {
				rep.Skipped = append(rep.Skipped, base)
				continue
			}
			if !opt.DryRun {
				if err := os.RemoveAll(dstEntry); err != nil {
					return rep, fmt.Errorf("删除旧条目失败 %s: %w", dstEntry, err)
				}
			}
		}
		if !opt.DryRun {
			if err := copyFile(src, dstEntry); err != nil {
				return rep, err
			}
		}
		rep.Copied = append(rep.Copied, base)
	}

	sort.Strings(rep.Copied)
	sort.Strings(rep.Ignored)
	sort.Strings(rep.Skipped)
	sort.Strings(rep.Rewritten)
	return rep, nil
}

// rewriteAssetsPointer 把复制过来的 SKILL.md 正文首行改写成指针行。
// 返回是否真的改了（内容相同就不写盘）。
func rewriteAssetsPointer(skillDstDir, assetsDir string) (bool, error) {
	p := filepath.Join(skillDstDir, "SKILL.md")
	st, err := os.Stat(p)
	if err != nil {
		return false, fmt.Errorf("读不到复制后的 SKILL.md %s: %w", p, err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return false, fmt.Errorf("读不到复制后的 SKILL.md %s: %w", p, err)
	}
	out, changed := InjectAssetsPointer(b, assetsDir)
	if !changed {
		return false, nil
	}
	if err := os.WriteFile(p, out, st.Mode().Perm()); err != nil {
		return false, fmt.Errorf("改写 SKILL.md 失败 %s: %w", p, err)
	}
	return true, nil
}

// copyTree 复制技能目录，**只跳过两件事**：根级 assets/（改为写指针，见 assets.go）
// 与**任意层级**的版本库元数据（见 isVCSMeta）。返回被跳过的版本库元数据相对路径。
func copyTree(src, dst string) ([]string, error) {
	var ignored []string
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		rel, rerr := filepath.Rel(src, p)
		if rerr != nil {
			return rerr
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if d.IsDir() && rel == assetsDirName {
			return fs.SkipDir
		}
		if isVCSMeta(d.Name()) {
			ignored = append(ignored, filepath.ToSlash(rel))
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil // worktree / submodule 的 .git 是文件，跳过即可
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(p, target)
	})
	return ignored, err
}

// isVCSMeta 判断一个条目是不是版本库元数据 —— 复制技能树时**任何层级**都不带它。
//
// 第三方技能常常就是一份 git clone 下来的工作副本。把 .git 一起装进 agent 目录有三个
// 坏处：目标凭空多出几百 MB 对象库；`tt install skills` 的 files 清单会按对象文件逐个
// 列出（一次几千行）；agent 扫技能树时还会撞见别人的仓库历史。技能要的是它的内容，
// 不是它的版本控制历史。
//
// 按**名字**判断，不看类型：worktree / submodule 里的 .git 是文件而不是目录。
func isVCSMeta(name string) bool {
	switch name {
	case ".git", ".hg", ".svn":
		return true
	}
	return false
}

// plannedFiles 是 dry-run 用的：算出会复制哪些文件（相对目标目录，/ 分隔），不写盘。
// ignored 与 copyTree 的规则必须一致，否则 --dry-run 与实际安装会给出两份不同的清单。
func plannedFiles(src, name string) (files, ignored []string, err error) {
	err = filepath.WalkDir(src, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		rel, rerr := filepath.Rel(src, p)
		if rerr != nil {
			return rerr
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == assetsDirName {
				return fs.SkipDir
			}
			if isVCSMeta(d.Name()) {
				ignored = append(ignored, name+"/"+rel)
				return fs.SkipDir
			}
			return nil
		}
		if isVCSMeta(d.Name()) {
			ignored = append(ignored, name+"/"+rel)
			return nil
		}
		files = append(files, name+"/"+rel)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("列计划文件失败 %s: %w", src, err)
	}
	return files, ignored, nil
}

func copyFile(src, dst string) error {
	st, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("读取源文件失败 %s: %w", src, err)
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("读取源文件失败 %s: %w", src, err)
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("建目标目录失败 %s: %w", filepath.Dir(dst), err)
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, st.Mode().Perm())
	if err != nil {
		return fmt.Errorf("建目标文件失败 %s: %w", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("复制 %s 失败: %w", src, err)
	}
	return out.Close()
}

// filesUnder 列出 root 下 entry 内的全部文件（相对 root，/ 分隔）。
func filesUnder(root, entry string) ([]string, error) {
	var out []string
	if err := fs.WalkDir(os.DirFS(root), entry, func(p string, d fs.DirEntry, werr error) error {
		if werr != nil || d.IsDir() {
			return werr
		}
		out = append(out, p)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("列复制结果失败: %w", err)
	}
	return out, nil
}

func pathExists(p string) (bool, error) {
	_, err := os.Lstat(p)
	switch {
	case err == nil:
		return true, nil
	case os.IsNotExist(err):
		return false, nil
	default:
		return false, fmt.Errorf("检查目标条目失败 %s: %w", p, err)
	}
}

// SamePath 判断两个路径是否指向同一处（Windows 上大小写不敏感）。
// 导出版本给设置页用：登记里删哪一条、加的是不是已有那条，都要用同一套判据。
func SamePath(a, b string) bool {
	if abs, err := filepath.Abs(a); err == nil {
		a = abs
	}
	if abs, err := filepath.Abs(b); err == nil {
		b = abs
	}
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
