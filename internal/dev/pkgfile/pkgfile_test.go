package pkgfile

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//---------------------------------------------------------------------------
// 小工具
//---------------------------------------------------------------------------

// corpusRoot 返回真实语料根目录。缺失时返回 ""，相关测试自动跳过
// （保证在无语料的机器上 go test ./... 依然能过）。
func corpusRoot() string {
	if v := os.Getenv("TDEV_CORPUS"); v != "" {
		if st, err := os.Stat(v); err == nil && st.IsDir() {
			return v
		}
		return ""
	}
	def := `D:\t100_wrok_dir`
	if st, err := os.Stat(def); err == nil && st.IsDir() {
		return def
	}
	return ""
}

func corpusPackages(t *testing.T) []string {
	t.Helper()
	root := corpusRoot()
	if root == "" {
		t.Skip("没有真实语料（设置 TDEV_CORPUS 或准备 D:\\t100_wrok_dir）")
	}
	var out []string
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.EqualFold(filepath.Ext(p), ".tzc") {
			out = append(out, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("扫描语料失败: %v", err)
	}
	if len(out) == 0 {
		t.Skipf("%s 下没有 .tzc", root)
	}
	return out
}

func writeTemp(t *testing.T, b []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.tzc")
	if err != nil {
		t.Fatalf("建临时文件失败: %v", err)
	}
	if _, err := f.Write(b); err != nil {
		t.Fatalf("写临时文件失败: %v", err)
	}
	f.Close()
	return f.Name()
}

// rawEntries 直接读 zip，返回 名 -> 内容 sha256（保序）。
func rawEntries(t *testing.T, p string) ([]string, map[string]string) {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("读 %s 失败: %v", p, err)
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("%s zip 解析失败: %v", p, err)
	}
	var order []string
	sumb := map[string]string{}
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			t.Fatalf("%s 打开条目 %s 失败: %v", p, zf.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("%s 读条目 %s 失败: %v", p, zf.Name, err)
		}
		order = append(order, zf.Name)
		sumb[zf.Name] = Sha256Hex(data)
	}
	return order, sumb
}

//---------------------------------------------------------------------------
// S2：零改动 roundtrip —— 逐条目 sha256 100% 一致
//---------------------------------------------------------------------------

func TestRoundtripZeroChangeCorpus(t *testing.T) {
	pkgs := corpusPackages(t)
	bad := 0
	for _, p := range pkgs {
		p := p
		t.Run(filepath.Base(p), func(t *testing.T) {
			pkg, err := Open(p, OpenOptions{})
			if err != nil {
				t.Fatalf("Open 失败: %v", err)
			}
			newBytes, actions, err := pkg.Build(Rebuild{})
			if err != nil {
				t.Fatalf("Build 失败: %v", err)
			}
			for _, a := range actions {
				if a.Changed {
					t.Errorf("零改动却有变更条目: %s", a.Name)
				}
			}
			out := writeTemp(t, newBytes)

			wantOrder, wantSum := rawEntries(t, p)
			gotOrder, gotSum := rawEntries(t, out)

			if len(wantOrder) != len(gotOrder) {
				t.Fatalf("条目数变了: %d -> %d", len(wantOrder), len(gotOrder))
			}
			for i := range wantOrder {
				if wantOrder[i] != gotOrder[i] {
					t.Errorf("第 %d 个条目名/顺序变了: %q -> %q", i, wantOrder[i], gotOrder[i])
				}
				if wantSum[wantOrder[i]] != gotSum[gotOrder[i]] {
					t.Errorf("条目内容 sha256 不一致: %s", wantOrder[i])
				}
			}
		})
		if t.Failed() {
			bad++
		}
	}
	t.Logf("roundtrip 语料包数=%d 失败=%d", len(pkgs), bad)
}

//---------------------------------------------------------------------------
// 容器层单元测试（不依赖语料）
//---------------------------------------------------------------------------

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in      string
		major   int
		minor   int
		wantErr bool
	}{
		{"1.0\n", 1, 0, false},
		{"1.0.0.3\n", 1, 0, false},
		{"1.0", 1, 0, false},
		{"2.1.7", 2, 1, false},
		{"", 0, 0, true},
		{"abc", 0, 0, true},
		{"1", 0, 0, true},
	}
	for _, c := range cases {
		v, err := ParseVersion([]byte(c.in))
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseVersion(%q) 期待报错，实际 %+v", c.in, v)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseVersion(%q) 意外报错: %v", c.in, err)
			continue
		}
		if v.Major != c.major || v.Minor != c.minor {
			t.Errorf("ParseVersion(%q) = %d.%d，想要 %d.%d", c.in, v.Major, v.Minor, c.major, c.minor)
		}
	}
}

func TestKindFromExt(t *testing.T) {
	k, diff, ind, simple := kindFromExt(".tzc")
	if k != KindCode || diff || ind || simple {
		t.Errorf(".tzc 判定错误: %v %v %v %v", k, diff, ind, simple)
	}
	k, diff, _, _ = kindFromExt(".tzx")
	if k != KindCode || !diff {
		t.Errorf(".tzx 应为 Code+IsDiff")
	}
	_, _, ind, _ = kindFromExt(".tzf")
	if !ind {
		t.Errorf(".tzf 应为 isIndFun")
	}
	_, _, _, simple = kindFromExt(".tzv")
	if !simple {
		t.Errorf(".tzv 应为 IsSimpleForm")
	}
	if k, _, _, _ := kindFromExt(".zip"); k != KindNone {
		t.Errorf(".zip 应为 KindNone")
	}
}

func TestFormatErrorExitCode(t *testing.T) {
	var err error = &FormatError{Msg: "x"}
	if ec, ok := err.(ExitCodeError); !ok || ec.ExitCode() != 2 {
		t.Fatalf("FormatError 应映射退出码 2")
	}
	var ioErr error = &IOError{Msg: "y"}
	if ec, ok := ioErr.(ExitCodeError); !ok || ec.ExitCode() != 5 {
		t.Fatalf("IOError 应映射退出码 5")
	}
}

// Build 必须拒绝改写 .4gl（红线 R1）与新增条目（红线 R2）。
func TestBuildRedlines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "adzi999.tzc")
	mk := func() []byte {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		for _, n := range []string{"adzi999.4gl", "adzi999.tap", "adzi999.tgl", "ver"} {
			w, _ := zw.Create(n)
			if n == "ver" {
				_, _ = w.Write([]byte("1.0\n"))
				continue
			}
			_, _ = w.Write([]byte("x"))
		}
		zw.Close()
		return buf.Bytes()
	}
	if err := os.WriteFile(path, mk(), 0o644); err != nil {
		t.Fatal(err)
	}
	pkg, err := Open(path, OpenOptions{})
	if err != nil {
		t.Fatalf("Open 失败: %v", err)
	}
	if _, _, err := pkg.Build(Rebuild{Replace: map[string][]byte{"adzi999.4gl": []byte("y")}}); err == nil {
		t.Errorf("改写 .4gl 应被拒绝")
	}
	if _, _, err := pkg.Build(Rebuild{Replace: map[string][]byte{"new.tap": []byte("y")}}); err == nil {
		t.Errorf("新增条目应被拒绝")
	}
}
