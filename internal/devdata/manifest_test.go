package devdata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestManifestRoundTrip schema 读写往返（含空数组要序列化成 []，不是 null）。
func TestManifestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest(KindFetch, "福联测试tst")
	m.Prog = "axmt500_wf"
	m.Module = "axm"
	m.FetchedAt = "2026-11-21T14:30:00+08:00"
	m.TTVersion = "0.3.1"
	m.Source = Source{TopDir: "/u1/t35tst", Zone: "2", Ent: 1108}
	m.Versions.DZAA002 = 3
	m.Versions.TglHeader = &TglHeader{SDVer: "0025(2025-11-13 17:33:55)"}
	m.Versions.TemplateVersions["a05"] = 3
	m.AddFile(FileEntry{Role: "tab", Src: "erp/axm/dzx/tab/axmt500_wf.tab", Rel: "dzx/axmt500_wf.tab",
		SHA256: "abc", Size: 93468, SrcMtime: "2026-11-21T12:27:32+08:00"})
	m.AddMissing("dzx/axmt500_wf.tgx2")
	m.AddWarning("tab 比 tgl 新")

	if err := WriteManifest(dir, m); err != nil {
		t.Fatalf("写 manifest: %v", err)
	}
	got, err := ReadManifest(dir)
	if err != nil {
		t.Fatalf("读 manifest: %v", err)
	}
	if got.Schema != ManifestSchema || got.Kind != KindFetch || got.Prog != "axmt500_wf" {
		t.Fatalf("基本字段丢了: %+v", got)
	}
	if got.Versions.DZAA002 != 3 || got.Versions.TemplateVersions["a05"] != 3 {
		t.Fatalf("版本锚点丢了: %+v", got.Versions)
	}
	if got.Versions.TglHeader == nil || got.Versions.TglHeader.SDVer != "0025(2025-11-13 17:33:55)" {
		t.Fatalf("tglHeader 丢了: %+v", got.Versions.TglHeader)
	}
	if len(got.Files) != 1 || got.Files[0].SHA256 != "abc" {
		t.Fatalf("files 丢了: %+v", got.Files)
	}
	if len(got.Missing) != 1 || len(got.Warnings) != 1 {
		t.Fatalf("missing/warnings 丢了: %+v / %+v", got.Missing, got.Warnings)
	}

	// 文件不存在 → (nil,nil)，不是错误（"没拉过"是正常状态）
	if m2, err := ReadManifest(t.TempDir()); err != nil || m2 != nil {
		t.Fatalf("没有 manifest 时应返回 nil,nil，得到 %+v, %v", m2, err)
	}
}

// TestLocalStale 本地两种触发：文件丢了、内容被改。
func TestLocalStale(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "dzx"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte("tab-body")
	if err := os.WriteFile(filepath.Join(dir, "dzx", "a.tab"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	m := NewManifest(KindFetch, "e")
	m.AddFile(FileEntry{Role: "tab", Rel: "dzx/a.tab", SHA256: BytesSHA256(body), Size: int64(len(body))})

	if got := LocalStale(dir, m); len(got) != 0 {
		t.Fatalf("一致时不该报 stale: %+v", got)
	}

	// ① 内容被改（长度不变也会被抓到 —— 那正是 sha 的意义）
	if err := os.WriteFile(filepath.Join(dir, "dzx", "a.tab"), []byte("tab-BODY"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := LocalStale(dir, m)
	if len(got) != 1 || got[0].Code != "sha-mismatch" {
		t.Fatalf("改内容应报 sha-mismatch: %+v", got)
	}

	// ② 文件丢了
	if err := os.Remove(filepath.Join(dir, "dzx", "a.tab")); err != nil {
		t.Fatal(err)
	}
	got = LocalStale(dir, m)
	if len(got) != 1 || got[0].Code != "missing-local" {
		t.Fatalf("丢文件应报 missing-local: %+v", got)
	}

	// ③ 没有 manifest
	if got := LocalStale(dir, nil); len(got) != 1 || got[0].Code != "missing-local" {
		t.Fatalf("没有 manifest 应报 missing-local: %+v", got)
	}
}

// TestRemoteStale 服务器侧的三种触发：文件变新、大小变化、文件消失、版本变化。
func TestRemoteStale(t *testing.T) {
	base := time.Date(2026, 11, 13, 12, 27, 32, 0, time.UTC)
	m := NewManifest(KindFetch, "e")
	m.AddFile(FileEntry{Role: "tab", Src: "erp/axm/dzx/tab/a.tab", Rel: "dzx/a.tab",
		Size: 10, SrcMtime: base.Format(time.RFC3339)})
	m.AddFile(FileEntry{Role: "tgl", Src: "erp/axm/dzx/tgl/a.tgl", Rel: "dzx/a.tgl",
		Size: 20, SrcMtime: base.Format(time.RFC3339)})
	m.Versions.TemplateVersions["a05"] = 3
	m.Versions.TglHeader = &TglHeader{SDVer: "0025"}

	// 一致（+1 秒在容差内）
	remote := map[string]RemoteStamp{
		"erp/axm/dzx/tab/a.tab": {Src: "erp/axm/dzx/tab/a.tab", Size: 10, ModTime: base.Add(time.Second), Exists: true},
		"erp/axm/dzx/tgl/a.tgl": {Src: "erp/axm/dzx/tgl/a.tgl", Size: 20, ModTime: base, Exists: true},
	}
	if got := RemoteStale(m, remote, map[string]int{"a05": 3}, &TglHeader{SDVer: "0025"}); len(got) != 0 {
		t.Fatalf("一致时不该报 stale: %+v", got)
	}

	// ① mtime 变新
	remote["erp/axm/dzx/tab/a.tab"] = RemoteStamp{Src: "erp/axm/dzx/tab/a.tab", Size: 10,
		ModTime: base.Add(time.Hour), Exists: true}
	got := RemoteStale(m, remote, nil, nil)
	if len(got) != 1 || got[0].Code != "remote-newer" {
		t.Fatalf("mtime 变新应报 remote-newer: %+v", got)
	}

	// ② 大小变化（mtime 没变也要报 —— 覆盖"同秒内改大小"）
	remote["erp/axm/dzx/tab/a.tab"] = RemoteStamp{Src: "erp/axm/dzx/tab/a.tab", Size: 11,
		ModTime: base, Exists: true}
	got = RemoteStale(m, remote, nil, nil)
	if len(got) != 1 || got[0].Code != "remote-newer" {
		t.Fatalf("大小变化应报 remote-newer: %+v", got)
	}

	// ③ 文件消失
	remote["erp/axm/dzx/tab/a.tab"] = RemoteStamp{Src: "erp/axm/dzx/tab/a.tab", Exists: false}
	got = RemoteStale(m, remote, nil, nil)
	if len(got) != 1 || got[0].Code != "remote-gone" {
		t.Fatalf("文件消失应报 remote-gone: %+v", got)
	}

	// ④ 版本变化：模板版本 + tgl 头部
	remote["erp/axm/dzx/tab/a.tab"] = RemoteStamp{Src: "erp/axm/dzx/tab/a.tab", Size: 10, ModTime: base, Exists: true}
	got = RemoteStale(m, remote, map[string]int{"a05": 4}, &TglHeader{SDVer: "0026"})
	if len(got) != 2 {
		t.Fatalf("版本变化应报两条: %+v", got)
	}
	for _, r := range got {
		if r.Code != "version-changed" {
			t.Fatalf("应为 version-changed: %+v", got)
		}
	}
}

// TestConsistencyWarnings tab 比 tgl 新 → 产物可能滞后。
func TestConsistencyWarnings(t *testing.T) {
	m := NewManifest(KindFetch, "e")
	m.AddFile(FileEntry{Role: "tab", Rel: "dzx/a.tab", SrcMtime: "2026-11-21T12:27:32+08:00"})
	m.AddFile(FileEntry{Role: "tgl", Rel: "dzx/a.tgl", SrcMtime: "2026-11-21T10:00:00+08:00"})
	ws := ConsistencyWarnings(m)
	if len(ws) != 1 || !strings.Contains(ws[0], "产物可能滞后") {
		t.Fatalf("tab 新于 tgl 应给滞后警告: %+v", ws)
	}

	// 反过来（tgl 新）不报警：那是正常重产流程
	m2 := NewManifest(KindFetch, "e")
	m2.AddFile(FileEntry{Role: "tab", Rel: "dzx/a.tab", SrcMtime: "2026-11-21T10:00:00+08:00"})
	m2.AddFile(FileEntry{Role: "tgl", Rel: "dzx/a.tgl", SrcMtime: "2026-11-21T12:27:32+08:00"})
	if ws := ConsistencyWarnings(m2); len(ws) != 0 {
		t.Fatalf("tgl 新于 tab 不该报警: %+v", ws)
	}

	// 缺 tab / 缺 tgl 必须开口
	m3 := NewManifest(KindFetch, "e")
	if ws := ConsistencyWarnings(m3); len(ws) != 2 {
		t.Fatalf("两个都缺应给两条警告: %+v", ws)
	}
}

// TestParseTemplateVersion 版本标记在文件头部。
func TestParseTemplateVersion(t *testing.T) {
	cases := []struct {
		head string
		v    int
		ok   bool
	}{
		{"            #應用 a05 樣板自動產生(Version:3)\n", 3, true},
		{"#應用 a06 樣板自動產生(Version:8)", 8, true},
		{"no version here", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		v, ok := ParseTemplateVersion([]byte(c.head))
		if ok != c.ok || (ok && v != c.v) {
			t.Errorf("ParseTemplateVersion(%q) = %d,%v; want %d,%v", c.head, v, ok, c.v, c.ok)
		}
	}
	// 只看头 4 KB：版本标记在很后面时不该被误读
	long := strings.Repeat("#\n", 3000) + "(Version:99)"
	if v, ok := ParseTemplateVersion([]byte(long)); ok {
		t.Errorf("超出头部范围不该解析出版本，得到 %d", v)
	}
}

// TestParseTglHeader .tgl 头部形态（实测 axmt500_wf.tgl）。
func TestParseTglHeader(t *testing.T) {
	head := "#該程式未解開Section, 採用最新樣板產出!\n" +
		"#該程式非freestyle程式!\n" +
		"{<section id=\"axmt500_wf.description\" type=\"s\" >}\n" +
		"#應用 a00 樣板自動產生(Version:3)\n" +
		"#+ Standard Version.....: SD版次:0025(2025-11-13 17:33:55), PR版次:\n" +
		"#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:\n" +
		"#+ Build......: 000037\n" +
		"#+ Filename...: axmt500_wf\n" +
		"#+ Description: 訂單維護作業\n"
	h := ParseTglHeader([]byte(head))
	if h == nil {
		t.Fatal("应解析出 tglHeader")
	}
	if h.SDVer != "0025(2025-11-13 17:33:55)" {
		t.Errorf("SDVer = %q", h.SDVer)
	}
	if h.Build != "000037" {
		t.Errorf("Build = %q", h.Build)
	}
	if h.Desc != "訂單維護作業" {
		t.Errorf("Desc = %q", h.Desc)
	}
	if ParseTglHeader([]byte("nothing")) != nil {
		t.Error("没有标记时应返回 nil")
	}
}
