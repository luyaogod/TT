package pkgfile

// 从零造包的形态断言。
//
// 钉住的是**设计器能不能打开**这件事，而不是"是不是合法 zip"：真机上 archive/zip.Writer
// 产出的包（置 bit 3 + 补数据描述符）设计器直接报
// 「Data descriptor signature not found」，而 tt 自己写出来的必须没有这个问题。

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteZipDesignerShape(t *testing.T) {
	entries := []WriteEntry{
		{Name: "axmt500_wf.4gl", Data: bytes.Repeat([]byte("let g_x = 1\n"), 200)},
		{Name: "axmt500_wf.tap", Data: bytes.Repeat([]byte("<point/>\n"), 300)},
		{Name: "axmt500_wf.tgl", Data: bytes.Repeat([]byte("tgl\n"), 50)},
		{Name: "ver", Data: []byte("1.0\n"), Store: true, Mode: 0o775},
	}
	var buf bytes.Buffer
	if err := WriteZip(&buf, entries); err != nil {
		t.Fatalf("WriteZip: %v", err)
	}

	rw, err := parseRawZip(buf.Bytes())
	if err != nil {
		t.Fatalf("parseRawZip: %v", err)
	}
	if len(rw.entries) != len(entries) {
		t.Fatalf("条目数 = %d，想要 %d", len(rw.entries), len(entries))
	}

	for i, e := range rw.entries {
		want := entries[i]
		if e.name != want.Name {
			t.Errorf("[%d] 条目名 = %q，想要 %q（顺序必须保持）", i, e.name, want.Name)
		}
		// 核心：不许置 bit 3（数据描述符），否则设计器打不开。
		if e.flags&zipFlagDataDescriptor != 0 {
			t.Errorf("[%d] %s flags=0x%04x 置了 bit 3（数据描述符）", i, e.name, e.flags)
		}
		if e.flags != 0 {
			t.Errorf("[%d] %s flags=0x%04x，想要 0", i, e.name, e.flags)
		}
		// CRC 与两个大小必须内联在局部头里。
		if e.crc32 == 0 || e.csize32 == 0 || e.usize32 == 0 {
			t.Errorf("[%d] %s 局部头里的 crc/csize/usize 有空值: %d/%d/%d",
				i, e.name, e.crc32, e.csize32, e.usize32)
		}
		if int(e.usize32) != len(want.Data) {
			t.Errorf("[%d] %s usize = %d，想要 %d", i, e.name, e.usize32, len(want.Data))
		}
		wantMethod := uint16(8)
		if want.Store {
			wantMethod = 0
		}
		if e.method != wantMethod {
			t.Errorf("[%d] %s method = %d，想要 %d", i, e.name, e.method, wantMethod)
		}
		// 局部头与中央目录都带 UT + ux 两个 extra（真包形态）。
		if !hasExtraTag(e.localHdr[30+int(e.nlen):30+int(e.nlen)+int(e.elen)], zipUTTag) {
			t.Errorf("[%d] %s 局部头缺 UT extra", i, e.name)
		}
		if !hasExtraTag(e.localHdr[30+int(e.nlen):30+int(e.nlen)+int(e.elen)], zipUxTag) {
			t.Errorf("[%d] %s 局部头缺 ux extra", i, e.name)
		}
	}

	// 局部头之间必须首尾相接：中间不能塞数据描述符。
	// 用 localHdr/compressed 的实际长度算，不用 nlen/elen —— 后者取自中央目录，
	// 而 UT extra 在局部头里比中央目录多 4 字节（atime）。
	for i := 1; i < len(rw.entries); i++ {
		prev := rw.entries[i-1]
		end := int(prev.offset32) + len(prev.localHdr) + len(prev.compressed)
		if int(rw.entries[i].offset32) != end {
			t.Errorf("%s 的局部头偏移 = %d，但上一条结束于 %d（凭空多出 %d 字节，疑似数据描述符）",
				rw.entries[i].name, rw.entries[i].offset32, end, int(rw.entries[i].offset32)-end)
		}
	}

	// 标准库能读、内容一致（CRC 由标准库校验）。
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("archive/zip 打不开产出物: %v", err)
	}
	got := map[string][]byte{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("打开 %s: %v", f.Name, err)
		}
		body := new(bytes.Buffer)
		if _, err := body.ReadFrom(rc); err != nil {
			rc.Close()
			t.Fatalf("读 %s: %v", f.Name, err)
		}
		rc.Close()
		got[f.Name] = body.Bytes()
	}
	for _, want := range entries {
		if !bytes.Equal(got[want.Name], want.Data) {
			t.Errorf("%s 内容不一致（标准库读回 %d B，原 %d B）", want.Name, len(got[want.Name]), len(want.Data))
		}
	}
}

func TestWriteZipOpensAsCodePackage(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "axmt500_wf(c).tzc")
	f, err := os.Create(p)
	if err != nil {
		t.Fatalf("创建: %v", err)
	}
	err = WriteZip(f, []WriteEntry{
		{Name: "axmt500_wf.4gl", Data: []byte("PROGRAM axmt500_wf\n")},
		{Name: "axmt500_wf.tap", Data: []byte(`<add_points prog="axmt500_wf" identity="c"></add_points>`)},
		{Name: "axmt500_wf.tgl", Data: []byte("TGL\n")},
		{Name: "ver", Data: []byte("1.0\n"), Store: true, Mode: 0o775},
	})
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		t.Fatalf("写包: %v", err)
	}

	pkg, err := Open(p, OpenOptions{AcceptVer: "1.0"})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if pkg.Kind != KindCode {
		t.Errorf("Kind = %s，想要 Code", pkg.Kind)
	}
	if pkg.Ver.Raw != "1.0" {
		t.Errorf("ver = %q，想要 1.0", pkg.Ver.Raw)
	}
	names := []string{}
	for _, e := range pkg.Entries {
		names = append(names, e.Name)
	}
	if want := "axmt500_wf.4gl,axmt500_wf.tap,axmt500_wf.tgl,ver"; strings.Join(names, ",") != want {
		t.Errorf("条目 = %v，想要 %s", names, want)
	}
}

func TestWriteZipRejectsBadNames(t *testing.T) {
	for _, name := range []string{"", "../evil.tap", "a/../../b.tap"} {
		var buf bytes.Buffer
		if err := WriteZip(&buf, []WriteEntry{{Name: name, Data: []byte("x")}}); err == nil {
			t.Errorf("条目名 %q 应该被拒绝", name)
		}
	}
}

func hasExtraTag(extra []byte, tag uint16) bool {
	for i := 0; i+4 <= len(extra); {
		got := uint16(extra[i]) | uint16(extra[i+1])<<8
		size := int(extra[i+2]) | int(extra[i+3])<<8
		if got == tag {
			return true
		}
		i += 4 + size
	}
	return false
}
