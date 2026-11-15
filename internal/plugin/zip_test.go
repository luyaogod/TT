package plugin

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// zentry 描述一个待写进测试 zip 的条目。mode 非零才写进 FileHeader，
// 好让普通文件保持 Go 缺省的"无权限位"形态。
type zentry struct {
	name  string
	body  string
	mode  os.FileMode
	dir   bool
	store bool
}

// buildZip 在内存里造 zip 再落到临时文件：ExtractZip 走 zip.OpenReader，必须有真文件。
func buildZip(t *testing.T, entries ...zentry) string {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		name := e.name
		if e.dir && !strings.HasSuffix(name, "/") {
			name += "/"
		}
		h := &zip.FileHeader{Name: name, Method: zip.Deflate}
		if e.store {
			h.Method = zip.Store
		}
		if e.mode != 0 {
			h.SetMode(e.mode)
		}
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatalf("造条目 %q 失败: %v", name, err)
		}
		if strings.HasSuffix(name, "/") {
			continue // 目录条目没有内容，往里写 zip.Writer 会报错
		}
		if _, err := w.Write([]byte(e.body)); err != nil {
			t.Fatalf("写条目 %q 失败: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("收尾 zip 失败: %v", err)
	}
	p := filepath.Join(t.TempDir(), "pkg.zip")
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("落盘 zip 失败: %v", err)
	}
	return p
}

func mustNotExist(t *testing.T, p string) {
	t.Helper()
	if _, err := os.Stat(p); err == nil {
		t.Errorf("%s 不该存在（条目写出了目标目录）", p)
	}
}

// 正常包必须能完整解开：嵌套目录、中文名、exe、可执行位、反斜杠分隔名。
func TestExtractZipHappyPath(t *testing.T) {
	src := buildZip(t,
		zentry{name: "docs", dir: true},
		zentry{name: "docs/说明.txt", body: "你好，插件"},
		zentry{name: "bin/tool.exe", body: "MZ fake"},
		zentry{name: "bin/tool", body: "#!/bin/sh\n", mode: 0o755},
		zentry{name: `legacy\note.txt`, body: "反斜杠分隔"},
	)
	dst := filepath.Join(t.TempDir(), "out")
	if err := ExtractZip(src, dst, DefaultZipLimits); err != nil {
		t.Fatalf("正常包不该失败: %v", err)
	}

	for _, tc := range []struct{ rel, want string }{
		{"docs/说明.txt", "你好，插件"},
		{"bin/tool.exe", "MZ fake"},
		{"bin/tool", "#!/bin/sh\n"},
		{"legacy/note.txt", "反斜杠分隔"},
	} {
		got, err := os.ReadFile(filepath.Join(dst, filepath.FromSlash(tc.rel)))
		if err != nil {
			t.Errorf("读 %s 失败: %v", tc.rel, err)
			continue
		}
		if string(got) != tc.want {
			t.Errorf("%s 内容不对: %q", tc.rel, got)
		}
	}
	if fi, err := os.Stat(filepath.Join(dst, "docs")); err != nil || !fi.IsDir() {
		t.Errorf("目录条目不该被当成文件: %v %v", fi, err)
	}

	// +x 分两层验：算出来的权限两平台都验；盘上的位只在 unix 上验，
	// Windows 的 os.Stat 只回报 0444/0666，可执行位在那里根本落不了盘。
	if got := entryPerm(0o644); got&0o111 != 0 {
		t.Errorf("普通文件不该带 +x，得到 %v", got)
	}
	if got := entryPerm(0o755); got&0o111 == 0 {
		t.Errorf("可执行位没保留，得到 %v", got)
	}
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(filepath.Join(dst, "bin", "tool"))
		if err != nil {
			t.Fatal(err)
		}
		if fi.Mode().Perm()&0o111 == 0 {
			t.Errorf("盘上丢了 +x: %v", fi.Mode())
		}
	}
}

// 越界条目：既要判为 ErrZipUnsafe、点名条目，也不能在 dst 外面留下任何东西。
func TestExtractZipRejectsTraversal(t *testing.T) {
	cases := []struct{ label, name string }{
		{"上级目录", "../evil.txt"},
		{"多级回跳", "a/../../evil.txt"},
		{"绝对路径", "/etc/passwd"},
		{"盘符路径", `C:\Windows\evil.txt`},
		{"反斜杠回跳", `..\evil.txt`},
		{"盘符相对", `C:evil.txt`},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			root := t.TempDir()
			src := buildZip(t, zentry{name: tc.name, body: "x"})
			dst := filepath.Join(root, "out")
			err := ExtractZip(src, dst, DefaultZipLimits)
			if !errors.Is(err, ErrZipUnsafe) {
				t.Fatalf("条目 %q 该判为不安全，得到 %v", tc.name, err)
			}
			// 错误里必须点名条目，否则用户不知道该删包里哪一项。
			if !strings.Contains(err.Error(), strconv.Quote(tc.name)) {
				t.Errorf("错误里没点名条目 %q: %v", tc.name, err)
			}
			for _, p := range []string{
				filepath.Join(root, "evil.txt"),
				filepath.Join(root, "etc", "passwd"),
				filepath.Join(root, "Windows", "evil.txt"),
				filepath.Join(filepath.Dir(root), "evil.txt"),
			} {
				mustNotExist(t, p)
			}
			if ents, err := os.ReadDir(dst); err == nil && len(ents) > 0 {
				t.Errorf("坏条目不该在 dst 里留下东西: %v", ents)
			}
		})
	}
}

// 名字本身就不能用的条目：NUL、清空了、只剩点。
func TestExtractZipRejectsBadNames(t *testing.T) {
	cases := []struct{ label, name string }{
		{"含 NUL", "bad\x00name.txt"},
		{"清理后为空", "."},
		{"只有斜杠", "./"},
		{"空名", ""},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			src := buildZip(t, zentry{name: tc.name, body: "x"})
			err := ExtractZip(src, filepath.Join(t.TempDir(), "out"), DefaultZipLimits)
			if !errors.Is(err, ErrZipUnsafe) {
				t.Fatalf("条目 %q 该被判为不安全，得到 %v", tc.name, err)
			}
		})
	}
}

// Windows 上大小写不敏感，重复落点必须按小写键挡掉，否则后者会静默覆盖前者。
func TestExtractZipRejectsDuplicates(t *testing.T) {
	cases := []struct {
		label   string
		entries []zentry
	}{
		{"同名条目", []zentry{{name: "dir/x", body: "1"}, {name: "dir/x", body: "2"}}},
		{"只有大小写不同", []zentry{{name: "Dir/x", body: "1"}, {name: "dir/x", body: "2"}}},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			src := buildZip(t, tc.entries...)
			err := ExtractZip(src, filepath.Join(t.TempDir(), "out"), DefaultZipLimits)
			if !errors.Is(err, ErrZipUnsafe) {
				t.Fatalf("重复路径该被拒绝，得到 %v", err)
			}
			// 报错要落在第二个条目上，用户才知道毛病在哪一项。
			if !strings.Contains(err.Error(), strconv.Quote("dir/x")) && !strings.Contains(err.Error(), strconv.Quote("Dir/x")) {
				t.Errorf("错误里没点名重复条目: %v", err)
			}
		})
	}
}

// 符号链接、设备文件都不能落成盘上的条目。
func TestExtractZipRejectsSpecialEntries(t *testing.T) {
	cases := []struct {
		label string
		mode  os.FileMode
	}{
		{"符号链接", os.ModeSymlink | 0o777},
		{"设备文件", os.ModeDevice | 0o600},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			src := buildZip(t, zentry{name: "special", body: "/etc/passwd", mode: tc.mode})
			dst := filepath.Join(t.TempDir(), "out")
			err := ExtractZip(src, dst, DefaultZipLimits)
			if !errors.Is(err, ErrZipUnsafe) {
				t.Fatalf("特殊条目该被拒绝，得到 %v", err)
			}
			mustNotExist(t, filepath.Join(dst, "special"))
		})
	}
}

// 三个上限各验一遍：条目数、单文件大小、总量。
func TestExtractZipLimits(t *testing.T) {
	cases := []struct {
		label   string
		entries []zentry
		limits  ZipLimits
	}{
		{
			label:   "条目数超限",
			entries: []zentry{{name: "a.txt", body: "a"}, {name: "b.txt", body: "b"}, {name: "c.txt", body: "c"}},
			limits:  ZipLimits{MaxFiles: 2},
		},
		{
			label:   "单文件超限",
			entries: []zentry{{name: "big.txt", body: "0123456789abcdef"}},
			limits:  ZipLimits{MaxFileSize: 8},
		},
		{
			label:   "总量超限",
			entries: []zentry{{name: "a.txt", body: "01234567"}, {name: "b.txt", body: "01234567"}},
			limits:  ZipLimits{MaxTotalSize: 12},
		},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			src := buildZip(t, tc.entries...)
			err := ExtractZip(src, filepath.Join(t.TempDir(), "out"), tc.limits)
			if !errors.Is(err, ErrZipUnsafe) {
				t.Fatalf("超过上限该被拒绝，得到 %v", err)
			}
		})
	}
}

// 头部可以撒谎：声明 10 字节、实际 100 字节。这种包必须整体失败，不能解出一个
// 被悄悄截断的文件；archive/zip 自己按声明值截断（ErrFormat），复制闸门是第二道。
func TestExtractZipRejectsLyingHeader(t *testing.T) {
	src := buildZip(t, zentry{name: "big.bin", body: strings.Repeat("A", 100), store: true})
	patchDeclaredUncompressedSize(t, src, 10)
	err := ExtractZip(src, filepath.Join(t.TempDir(), "out"),
		ZipLimits{MaxFiles: 10, MaxFileSize: 50, MaxTotalSize: 1000, MaxRatio: 200})
	if err == nil {
		t.Fatal("撒谎的头不该解包成功")
	}
	if !strings.Contains(err.Error(), strconv.Quote("big.bin")) {
		t.Errorf("错误里没点名条目: %v", err)
	}
}

// 复制闸门单独验：头部先验之外，落盘最多只到预算，越界一律 errOverBudget。
func TestCopyLimited(t *testing.T) {
	cases := []struct {
		label    string
		size     int64
		budget   int64
		wantOver bool
	}{
		{"刚好用满", 100, 100, false},
		{"超过 1 字节", 101, 100, true},
		{"远超预算", 1 << 20, 10, true},
		{"空文件", 0, 0, false},
		{"零预算非空", 1, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := copyLimited(&buf, bytes.NewReader(make([]byte, tc.size)), tc.budget)
			if got := errors.Is(err, errOverBudget); got != tc.wantOver {
				t.Fatalf("copyLimited(size=%d,budget=%d) = %v，wantOver=%v", tc.size, tc.budget, err, tc.wantOver)
			}
			if int64(buf.Len()) > tc.budget {
				t.Errorf("落盘 %d 字节，超过预算 %d", buf.Len(), tc.budget)
			}
			if n < int64(buf.Len()) {
				t.Errorf("返回字节数 %d 小于落盘字节数 %d", n, buf.Len())
			}
			// 越界时多探了 1 字节（只用来判断，不写盘）。
			if tc.wantOver && n != tc.budget+1 {
				t.Errorf("越界该按预算+1 字节上报，得到 %d", n)
			}
		})
	}
}

// patchDeclaredUncompressedSize 把唯一一个条目的声明解压大小改小（局部头与中央目录两处），
// 真实字节数不动，只改"声称值"，用来模拟撒谎的头。
func patchDeclaredUncompressedSize(t *testing.T, path string, unc uint32) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(raw, []byte{'P', 'K', 3, 4}) {
		t.Fatal("不是预期的 zip 局部头")
	}
	binary.LittleEndian.PutUint32(raw[22:], unc) // 局部头：压缩后大小在 18，解压后大小在 22
	cd := bytes.Index(raw, []byte{'P', 'K', 1, 2})
	if cd < 0 {
		t.Fatal("找不到中央目录项")
	}
	binary.LittleEndian.PutUint32(raw[cd+24:], unc) // 中央目录：压缩后大小在 20，解压后大小在 24
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

// 高压缩比的大条目就是 zip 炸弹的基本形态。
func TestExtractZipRejectsBomb(t *testing.T) {
	bomb := strings.Repeat("0", 2<<20) // 2MiB 的零：deflate 后只剩几 KB
	src := buildZip(t, zentry{name: "bomb.bin", body: bomb})
	limits := DefaultZipLimits
	limits.MaxRatio = 5
	err := ExtractZip(src, filepath.Join(t.TempDir(), "out"), limits)
	if !errors.Is(err, ErrZipUnsafe) {
		t.Fatalf("高压缩比条目该被拒绝，得到 %v", err)
	}
}

// 反向对照：不压缩的大文件压缩比是 1，不该被当成炸弹误杀。
func TestExtractZipAcceptsIncompressibleLargeEntry(t *testing.T) {
	src := buildZip(t, zentry{name: "big.bin", body: strings.Repeat("x", 2<<20), store: true})
	limits := DefaultZipLimits
	limits.MaxRatio = 5
	dst := filepath.Join(t.TempDir(), "out")
	if err := ExtractZip(src, dst, limits); err != nil {
		t.Fatalf("不压缩的大文件不该被拒: %v", err)
	}
	if fi, err := os.Stat(filepath.Join(dst, "big.bin")); err != nil || fi.Size() != 2<<20 {
		t.Fatalf("大文件没完整落盘: %v %v", fi, err)
	}
}

// 压缩比判定的边界与"没有压缩数据"这种构造出来的形态。
func TestCheckRatio(t *testing.T) {
	cases := []struct {
		label     string
		comp, unc int64
		maxRatio  int64
		wantErr   bool
	}{
		{"守门线及以下不判", 100, ratioGuardSize, 5, false},
		{"压缩比等于上限", 1 << 20, 1 << 20, 1, false},
		{"压缩比超过上限", 1 << 20, 2 << 20, 1, true},
		{"有内容却没有压缩数据", 0, 2 << 20, 200, true},
		{"不压缩的大文件", 2 << 20, 2 << 20, 200, false},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			err := checkRatio("e.bin", tc.comp, tc.unc, tc.maxRatio)
			if (err != nil) != tc.wantErr {
				t.Fatalf("checkRatio(%d,%d,%d) = %v，wantErr=%v", tc.comp, tc.unc, tc.maxRatio, err, tc.wantErr)
			}
			if tc.wantErr && !errors.Is(err, ErrZipUnsafe) {
				t.Errorf("拒绝要包装 ErrZipUnsafe，得到 %v", err)
			}
		})
	}
}

// 空包也是合法包：只建 dst，不报错。
func TestExtractZipEmptyArchive(t *testing.T) {
	src := buildZip(t)
	dst := filepath.Join(t.TempDir(), "out")
	if err := ExtractZip(src, dst, DefaultZipLimits); err != nil {
		t.Fatalf("空包不该失败: %v", err)
	}
	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("dst 该被创建: %v", err)
	}
	if !fi.IsDir() {
		t.Fatalf("dst 不是目录: %v", fi.Mode())
	}
	ents, err := os.ReadDir(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 0 {
		t.Errorf("空包不该解出条目: %v", ents)
	}
}
