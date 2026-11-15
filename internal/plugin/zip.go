// Package plugin 是插件包的安装面：只管把插件 zip 安全落到暂存目录，不参与命令装配。
//
// 解包是"外来字节直接变成可执行文件"的唯一入口，所以下面每条限制都不是防御性编程的
// 习惯动作，而是插件分发的前提：包来自用户或第三方，里面既有 CLI 二进制，也可能有
// 伪装成正常包的东西。
package plugin

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ratioGuardSize 压缩比只对大于它的条目生效：正常的小文本本来就能压到上千倍，
// 一体适用会误杀清单、许可证这类文件。
const ratioGuardSize = 1 << 20

// ErrZipUnsafe 是内容级拒绝的共同标记。调用方据此把"包坏了"和"包不安分"分开处理：
// 前者可以重下，后者要换来源。
var ErrZipUnsafe = errors.New("zip 内容不安全")

// ZipLimits 解包上限。
type ZipLimits struct {
	MaxFiles     int   // 条目总数上限
	MaxFileSize  int64 // 单个文件解压后上限
	MaxTotalSize int64 // 全部文件解压后总量上限
	MaxRatio     int64 // 单个文件的压缩比上限（仅对大于 ratioGuardSize 的文件生效）
}

// DefaultZipLimits 缺省上限。
var DefaultZipLimits = ZipLimits{MaxFiles: 5000, MaxFileSize: 64 << 20, MaxTotalSize: 512 << 20, MaxRatio: 200}

// withDefaults 把没填的字段补成缺省值：零值在这里只表示"调用方没说"，不表示"不限制"。
// 漏填就放开上限，等于给 zip 炸弹开绿灯。
func (l ZipLimits) withDefaults() ZipLimits {
	d := DefaultZipLimits
	if l.MaxFiles <= 0 {
		l.MaxFiles = d.MaxFiles
	}
	if l.MaxFileSize <= 0 {
		l.MaxFileSize = d.MaxFileSize
	}
	if l.MaxTotalSize <= 0 {
		l.MaxTotalSize = d.MaxTotalSize
	}
	if l.MaxRatio <= 0 {
		l.MaxRatio = d.MaxRatio
	}
	return l
}

// ExtractZip 把 src 处的一个 zip 安全解到 dst 目录（dst 不存在则创建）。
//
// 每个条目都要先过名字、类型、上限三关才落盘；任何一关不过就整体失败，不做"跳过坏的
// 继续解好的"——半装的插件比不装更难查。出错不清理 dst：调用方在暂存目录里解，验完
// 再整体换上去，失败时自己删暂存目录。
func ExtractZip(src, dst string, limits ZipLimits) error {
	limits = limits.withDefaults()

	zr, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("打开 %s 失败: %w", src, err)
	}
	defer zr.Close()

	// 条目数先按中央目录判：len(zr.File) 由包自己决定，不能拿它当预分配的依据，
	// 一个声称百万条目的头也不该先把内存吃掉。
	if len(zr.File) > limits.MaxFiles {
		return fmt.Errorf("压缩包有 %d 个条目，超过上限 %d: %w", len(zr.File), limits.MaxFiles, ErrZipUnsafe)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("建目标目录 %s 失败: %w", dst, err)
	}

	seen := make(map[string]string, len(zr.File)) // 上限已经判过，这个容量有界
	var total int64
	for _, f := range zr.File {
		if err := extractEntry(dst, f, limits, seen, &total); err != nil {
			return err
		}
	}
	return nil
}

// extractEntry 解一个条目。所有拒绝都点名条目：用户要知道"哪个条目"，而不是"解包失败"。
func extractEntry(dst string, f *zip.File, limits ZipLimits, seen map[string]string, total *int64) error {
	rel, err := safeEntryName(f.Name)
	if err != nil {
		return err
	}
	// Windows 文件系统大小写不敏感：A/x 与 a/x 落点是同一个文件，后者会静默覆盖前者，
	// "包里有什么"与"盘上有什么"就对不上了。按小写键去重。
	key := strings.ToLower(rel)
	if prev, dup := seen[key]; dup {
		return fmt.Errorf("条目 %q 与 %q 路径重复: %w", f.Name, prev, ErrZipUnsafe)
	}
	seen[key] = f.Name
	target := filepath.Join(dst, filepath.FromSlash(rel))

	mode := f.Mode()
	// 解包只该产出普通文件和目录：符号链接能指到包外，设备/管道/socket 在解包场景里
	// 没有正当用途，一律拒绝。
	if mode&os.ModeSymlink != 0 {
		return fmt.Errorf("条目 %q 是符号链接: %w", f.Name, ErrZipUnsafe)
	}
	if mode.IsDir() {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return fmt.Errorf("建目录 %s 失败: %w", target, err)
		}
		return nil
	}
	if !mode.IsRegular() {
		return fmt.Errorf("条目 %q 不是普通文件（%s）: %w", f.Name, mode.Type(), ErrZipUnsafe)
	}

	// 先按头部声明判一遍：头部会撒谎，但这一遍能挡掉绝大多数炸弹，让明显超标的包在
	// 读数据之前就失败。
	if err := checkSizes(f, limits, *total); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("建目录 %s 失败: %w", filepath.Dir(target), err)
	}
	return writeEntry(f, target, limits, total)
}

// checkSizes 按条目头声明的大小判上限与压缩比。
func checkSizes(f *zip.File, limits ZipLimits, total int64) error {
	unc := int64(f.UncompressedSize64)
	if unc > limits.MaxFileSize {
		return fmt.Errorf("条目 %q 解压后 %d 字节，超过单文件上限 %d: %w", f.Name, unc, limits.MaxFileSize, ErrZipUnsafe)
	}
	if total+unc > limits.MaxTotalSize {
		return fmt.Errorf("条目 %q 会把解压总量撑到 %d 字节，超过上限 %d: %w", f.Name, total+unc, limits.MaxTotalSize, ErrZipUnsafe)
	}
	return checkRatio(f.Name, int64(f.CompressedSize64), unc, limits.MaxRatio)
}

// checkRatio 判压缩比。ratioGuardSize 及以下不判：正常的小文本压缩比本来就能上千倍。
func checkRatio(name string, comp, unc, maxRatio int64) error {
	if unc <= ratioGuardSize {
		return nil
	}
	if comp == 0 {
		// 有解压后内容却没有压缩数据：合法 zip 里不会出现，只能是构造出来的。
		return fmt.Errorf("条目 %q 声称解压后 %d 字节却没有压缩数据: %w", name, unc, ErrZipUnsafe)
	}
	if unc/comp > maxRatio {
		return fmt.Errorf("条目 %q 压缩比 %d 超过上限 %d: %w", name, unc/comp, maxRatio, ErrZipUnsafe)
	}
	return nil
}

// writeEntry 落到盘上，并按实际复制的字节数记账。
func writeEntry(f *zip.File, target string, limits ZipLimits, total *int64) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("读条目 %q 失败: %w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, entryPerm(f.Mode()))
	if err != nil {
		return fmt.Errorf("建文件 %s 失败: %w", target, err)
	}
	defer out.Close() // 出错时兜底；成功后下面显式关，以拿到刷盘错误

	// 头部会撒谎，真正的上限必须落在复制这一侧：预算取"单文件余量"与"总量余量"的
	// 较小者，越界按实际写出的字节数判。
	budget := limits.MaxFileSize
	if left := limits.MaxTotalSize - *total; left < budget {
		budget = left
	}
	if budget < 0 {
		budget = 0
	}
	n, err := copyLimited(out, rc, budget)
	if errors.Is(err, errOverBudget) {
		return fmt.Errorf("条目 %q 解压后至少 %d 字节，超过上限 %d: %w", f.Name, n, budget, ErrZipUnsafe)
	}
	if err != nil {
		return fmt.Errorf("写条目 %q 到 %s 失败: %w", f.Name, target, err)
	}
	*total += n
	if err := out.Close(); err != nil {
		return fmt.Errorf("收尾 %s 失败: %w", target, err)
	}
	return nil
}

// errOverBudget 表示复制阶段越界，由调用方补上条目名后再归到 ErrZipUnsafe。
var errOverBudget = errors.New("解压后字节数超过上限")

// copyLimited 按 budget 上限复制，返回实际写出的字节数。
//
// 头部声明的大小只是先验，真正的闸门在这里：先写满预算，再多探 1 字节判断有没有
// 溢出；探到的这 1 字节不写盘，所以磁盘最多只到上限，撒谎的头也撑不爆。
func copyLimited(dst io.Writer, src io.Reader, budget int64) (int64, error) {
	lr := &io.LimitedReader{R: src, N: budget}
	n, err := io.Copy(dst, lr)
	if err != nil {
		return n, err
	}
	if lr.N > 0 {
		return n, nil
	}
	var one [1]byte
	k, rerr := src.Read(one[:])
	if k > 0 {
		return n + int64(k), errOverBudget
	}
	if rerr != nil && rerr != io.EOF {
		return n, rerr
	}
	return n, nil
}

// entryPerm 文件权限：缺省 0o644，zip 里带可执行位时补 0o111。
// // 插件可能随包带 CLI 二进制，解包时丢掉 +x，在 unix 上装完就"找不到命令"。
func entryPerm(mode fs.FileMode) os.FileMode {
	perm := os.FileMode(0o644)
	if mode&0o111 != 0 {
		perm |= 0o111
	}
	return perm
}

// safeEntryName 把条目名归一化成以 / 分段的相对路径；不安全就返回 ErrZipUnsafe。
//
// 先换反斜杠：Windows 打包工具写出来的条目名就是反斜杠，只按 / 判的话 `..\evil.txt`
// 会绕过检查，再被 filepath.FromSlash 拼成真的回跳。
func safeEntryName(name string) (string, error) {
	if strings.ContainsRune(name, 0) {
		return "", fmt.Errorf("条目 %q 含 NUL 字节: %w", name, ErrZipUnsafe)
	}
	p := strings.ReplaceAll(name, `\`, "/")
	// 绝对路径与盘符前缀都按字符串判：Windows 上 `C:/x` 算绝对路径，类 unix 上
	// filepath.IsAbs 不认它；统一判法才能两平台行为一致。
	if strings.HasPrefix(p, "/") || hasDrivePrefix(p) {
		return "", fmt.Errorf("条目 %q 是绝对路径（带盘符或根）: %w", name, ErrZipUnsafe)
	}
	for _, seg := range strings.Split(p, "/") {
		// 归档里的正常相对路径不需要 ..；一律拒绝，免得结论依赖 path.Clean 的具体行为。
		if seg == ".." {
			return "", fmt.Errorf("条目 %q 含 .. 段: %w", name, ErrZipUnsafe)
		}
		// Windows 上 `a:b` 是 NTFS 交换数据流：写不进目录列表，重复检测也看不见它。
		if strings.Contains(seg, ":") {
			return "", fmt.Errorf("条目 %q 含冒号: %w", name, ErrZipUnsafe)
		}
	}
	clean := path.Clean(p)
	if clean == "." || clean == "" {
		return "", fmt.Errorf("条目 %q 清理后为空: %w", name, ErrZipUnsafe)
	}
	// 兜底：万一上面的 .. 段检查被绕过，清理后的第一段是 .. 同样算逃逸。
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("条目 %q 逃出了目标目录: %w", name, ErrZipUnsafe)
	}
	return clean, nil
}

// hasDrivePrefix 认 C: 这类盘符前缀（C:/x、C:\x、C:x）。
func hasDrivePrefix(p string) bool {
	if len(p) < 2 || p[1] != ':' {
		return false
	}
	c := p[0]
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}
