package plugin

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// InstallOptions 安装参数。
type InstallOptions struct {
	// Source 本地目录或 zip 包（远程 URL 由阶段 6 的更新器负责下载后落到本地再进来）。
	Source string
	// Replace 同 ID 同版本已存在时重装（默认拒绝：那会掩盖"你以为换了个包"的事实）。
	Replace bool
}

// Install 安装一个插件包：解包到缓存 → 严格校验 → 原子移入版本目录 → 注册为**未启用**。
//
// 顺序是有意的：
//   - 先校验后移动：校验失败的半成品只在缓存里，不会出现在版本目录；
//   - 移动是 rename（同一卷内原子）：要么版本目录完整存在，要么不存在；
//   - 注册最后做：注册失败会回滚刚移入的版本目录。
//
// 权限确认与启用是**独立步骤**（见 Manager.Enable）：解包完不执行任何第三方代码。
func (m *Manager) Install(opt InstallOptions) (*Record, error) {
	src := strings.TrimSpace(opt.Source)
	if src == "" {
		return nil, errors.New("必须给出来源（本地插件目录或 zip 包路径）")
	}
	abs, err := filepath.Abs(src)
	if err != nil {
		return nil, fmt.Errorf("解析来源路径失败: %w", err)
	}
	st, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("来源不存在: %w", err)
	}
	if err := m.Store.EnsureDirs(); err != nil {
		return nil, err
	}

	stage, err := m.newStageDir()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(stage) // 移入版本目录后这里是空操作（stage 已被 rename 走）

	var source Source
	if st.IsDir() {
		source = Source{Kind: SourceDir, Ref: abs}
		if err := copyTree(abs, stage, DefaultZipLimits); err != nil {
			return nil, fmt.Errorf("复制插件目录失败: %w", err)
		}
	} else {
		sum, err := fileSHA256(abs)
		if err != nil {
			return nil, err
		}
		source = Source{Kind: SourceFile, Ref: abs, SHA256: sum}
		if err := ExtractZip(abs, stage, DefaultZipLimits); err != nil {
			return nil, fmt.Errorf("解包插件包失败: %w", err)
		}
	}

	// 常见 zip 布局是"一个顶层目录 + 里面的 plugin.json"：允许，并下探一层。
	root, err := manifestRoot(stage)
	if err != nil {
		return nil, err
	}
	man, err := LoadManifest(root)
	if err != nil {
		return nil, err
	}

	reg, err := m.Store.Load()
	if err != nil {
		return nil, err
	}
	vopt := ValidateOptions{
		HostAPIVersion:     m.hostAPIVersion(),
		ReservedNamespaces: m.ReservedNamespaces,
		OtherNamespaces:    reg.NamespacesInUse(man.ID),
		OtherRoutes:        reg.RoutesInUse(man.ID),
	}
	if err := ValidateManifest(man, root, vopt); err != nil {
		return nil, err
	}

	// 同 ID 已装记录：版本目录、启用状态的兼容性判断。
	existing := reg.Find(man.ID)
	if existing != nil && existing.Enabled {
		return nil, fmt.Errorf("插件 %s 正在启用中：先 tt plugin disable %s 再安装新版本（热升级在后续阶段提供）", man.ID, man.ID)
	}
	if existing != nil && existing.HasVersion(man.Version) && !opt.Replace {
		return nil, fmt.Errorf("插件 %s 的版本 %s 已安装：确要重装加 --replace", man.ID, man.Version)
	}

	target := m.Store.VersionDir(man.ID, man.Version)
	if target == "" {
		return nil, fmt.Errorf("插件 ID/版本不能用作目录名: %q / %q", man.ID, man.Version)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return nil, fmt.Errorf("创建版本父目录失败: %w", err)
	}
	if _, err := os.Stat(target); err == nil {
		// 只有 Replace 才走到这里；旧目录先挪开，避免"删一半"的中间态。
		aside := target + fmt.Sprintf(".old-%s", m.Store.now().Format("20060102-150405"))
		if err := os.Rename(target, aside); err != nil {
			return nil, fmt.Errorf("移走旧版本目录失败: %w", err)
		}
		defer os.RemoveAll(aside)
	}
	if err := os.Rename(root, target); err != nil {
		return nil, fmt.Errorf("移入版本目录失败: %w", err)
	}

	rec, err := m.registerInstalled(man, target, source, opt.Replace)
	if err != nil {
		// 注册失败 → 把载荷撤掉，不留"目录里有、注册表里没有"的孤儿。
		_ = os.RemoveAll(target)
		return nil, err
	}
	return rec, nil
}

// registerInstalled 把已落盘的版本写进注册表。
func (m *Manager) registerInstalled(man *Manifest, target string, source Source, replace bool) (*Record, error) {
	var out *Record
	err := m.Store.WithLock(func(reg *Registry) error {
		now := m.Store.Stamp()
		rec := reg.Find(man.ID)
		if rec == nil {
			rec = &Record{ID: man.ID, InstalledAt: now}
			reg.Put(rec)
		}
		if replace {
			rec.Versions = removeVersion(rec.Versions, man.Version)
		}
		rec.Versions = append(rec.Versions, InstalledVersion{
			Version: man.Version, InstalledAt: now, Source: source,
		})
		rec.ActiveVersion = man.Version
		rec.Name = man.Name
		rec.Source = source
		rec.Contributions = man.Contributions
		rec.Permissions = append([]string(nil), man.Permissions...)
		rec.UpdatedAt = now
		rec.LastError = ""
		// 新装的插件一律**未启用**；已装的保持它原来的启用状态（上面已拒绝"启用中"的情况，
		// 所以这里实际是 false）。
		out = rec
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// removeVersion 从版本清单里去掉某版本（重装时用）。
func removeVersion(in []InstalledVersion, v string) []InstalledVersion {
	out := in[:0]
	for _, iv := range in {
		if iv.Version != v {
			out = append(out, iv)
		}
	}
	return out
}

// newStageDir 在插件缓存目录里建一个唯一暂存目录。
func (m *Manager) newStageDir() (string, error) {
	cache := m.Store.CacheDir()
	if cache == "" {
		return "", errors.New("数据目录未确定，无法建立暂存目录")
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", fmt.Errorf("创建插件缓存目录失败: %w", err)
	}
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成暂存目录名失败: %w", err)
	}
	dir := filepath.Join(cache, "stage-"+hex.EncodeToString(b))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建暂存目录失败: %w", err)
	}
	return dir, nil
}

// manifestRoot 找到清单所在目录：根本身，或唯一的顶层子目录。
//
// zip 的习惯布局是 `<name>/plugin.json`，硬要求"清单必须在根"会让所有按习惯打包的
// 插件都装不上；而下探只允许**唯一**子目录，避免两个包混在一起时猜错。
func manifestRoot(stage string) (string, error) {
	if _, err := os.Stat(filepath.Join(stage, ManifestFileName)); err == nil {
		return stage, nil
	}
	entries, err := os.ReadDir(stage)
	if err != nil {
		return "", fmt.Errorf("读取包内容失败: %w", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) == 1 {
		cand := filepath.Join(stage, dirs[0])
		if _, err := os.Stat(filepath.Join(cand, ManifestFileName)); err == nil {
			return cand, nil
		}
	}
	if len(dirs) == 0 {
		return "", fmt.Errorf("包内没有 %s", ManifestFileName)
	}
	return "", fmt.Errorf("包根没有 %s，且顶层目录不止一个（无法确定哪个是插件根）", ManifestFileName)
}

// fileSHA256 算文件摘要（发布清单与"这个包是不是同一个"的判据）。
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开包文件失败: %w", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("读取包文件失败: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// copyTree 把目录复制到 dst（目录来源用）。
//
// 与 zip 解包同一套限制：**不跟随符号链接**（那等于复制了指向包外的洞），
// 仍然受文件数/单文件/总量上限约束。
func copyTree(src, dst string, limits ZipLimits) error {
	total := int64(0)
	files := 0
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			return fmt.Errorf("%w: 包内含符号链接 %s（不允许）", ErrZipUnsafe, rel)
		case d.IsDir():
			return os.MkdirAll(target, 0o755)
		case info.Mode().IsRegular():
			files++
			if limits.MaxFiles > 0 && files > limits.MaxFiles {
				return fmt.Errorf("%w: 文件数超过上限 %d", ErrZipUnsafe, limits.MaxFiles)
			}
			if limits.MaxFileSize > 0 && info.Size() > limits.MaxFileSize {
				return fmt.Errorf("%w: 文件 %s 超过单文件上限", ErrZipUnsafe, rel)
			}
			total += info.Size()
			if limits.MaxTotalSize > 0 && total > limits.MaxTotalSize {
				return fmt.Errorf("%w: 解包总量超过上限", ErrZipUnsafe)
			}
			return copyFile(path, target, info.Mode().Perm())
		default:
			return fmt.Errorf("%w: 包内含非普通文件 %s", ErrZipUnsafe, rel)
		}
	})
}

// copyFile 复制一个文件并保留权限位（可执行位要留下：CLI 插件靠它启动）。
func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm|0o200)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// RemoveVersionsExcept 清理除 keep 之外的版本目录（升级成功后调用；持久数据不动）。
func (m *Manager) RemoveVersionsExcept(rec *Record, keep string) []string {
	if rec == nil || rec.ActiveVersion == "" {
		return nil
	}
	parent := filepath.Dir(m.Store.VersionDir(rec.ID, rec.ActiveVersion)) // <installed>/<id>
	if parent == "" || parent == "." {
		return nil
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil
	}
	var removed []string
	for _, e := range entries {
		if !e.IsDir() || e.Name() == keep {
			continue
		}
		if err := os.RemoveAll(filepath.Join(parent, e.Name())); err == nil {
			removed = append(removed, e.Name())
		}
	}
	return removed
}

// nowString 给需要直接写时间戳的调用方。
func nowString() string { return time.Now().Format(time.RFC3339) }
