package web

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 站点导出/导入的对内契约:文件名安全、内容可往返、坏文件不致命。
//
// 这两个端点读写的是本机目录,默认档测试就能覆盖全 —— 不需要真连一台服务器。

func TestSiteFileNameIsSafe(t *testing.T) {
	cases := []struct{ in, want string }{
		{"109tst", "109tst.json"},
		// 路径分隔符与 Windows 非法字符都要换掉:站点名是自由文本,直接拼进路径会写错位置。
		{"a/b:c*d?e\"f<g>h|i", "a_b_c_d_e_f_g_h_i.json"},
		{"  前后空格  ", "前后空格.json"},
		{"...hidden", "hidden.json"},
		// 全是分隔符时替换成下划线,仍是一个合法文件名(不必退化成 site.json)。
		{"///", "___.json"},
		{"", "site.json"},
	}
	for _, c := range cases {
		if got := siteFileName(c.in); got != c.want {
			t.Errorf("siteFileName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// 过长名字要截断(Windows 的 MAX_PATH,而目录本身可能已经很深)。
	long := siteFileName(strings.Repeat("测", 200))
	if n := len([]rune(long)); n > 70 {
		t.Errorf("long name not truncated: %d runes", n)
	}
}

func TestHostsExportThenImportRoundTrip(t *testing.T) {
	s, _ := newTestServer(t, seededConfig)
	dir := t.TempDir()

	rec, body := doJSON(t, s, "POST", "/api/hosts/export", map[string]any{"name": "开发环境", "dir": dir})
	if rec.Code != 200 {
		t.Fatalf("export status = %d (body %v)", rec.Code, body)
	}
	path, _ := body["path"].(string)
	if filepath.Dir(path) != dir || !strings.HasSuffix(path, "开发环境.json") {
		t.Fatalf("export path = %q", path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var f siteFile
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("导出的文件不是 JSON: %v", err)
	}
	if f.Kind != siteFileKind || f.SchemaVersion != siteFileVersion {
		t.Fatalf("kind/version = %q/%d", f.Kind, f.SchemaVersion)
	}
	// 数据库连接必须一起带走:否则"导出再导入"会静默丢掉库配置。
	if f.Site.DB == nil || f.Site.DB.Host != "10.0.0.2" {
		t.Fatalf("site.db 没有随文件导出: %+v", f.Site.DB)
	}

	rec, body = doJSON(t, s, "POST", "/api/hosts/import", map[string]any{"dir": dir})
	if rec.Code != 200 {
		t.Fatalf("import status = %d (body %v)", rec.Code, body)
	}
	sites, _ := body["sites"].([]any)
	if len(sites) != 1 {
		t.Fatalf("sites = %v", body["sites"])
	}
	got, _ := sites[0].(map[string]any)
	site, _ := got["site"].(map[string]any)
	if site["name"] != "开发环境" || site["host"] != "10.0.0.1" || site["port"] != float64(22) {
		t.Fatalf("imported site = %v", site)
	}
}

func TestHostsImportToleratesBadFilesAndBareSites(t *testing.T) {
	s, _ := newTestServer(t, seededConfig)
	dir := t.TempDir()

	// 1) 一个不是站点的 json
	if err := os.WriteFile(filepath.Join(dir, "noise.json"), []byte(`{"hello":"world"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	// 2) 手写的裸站点对象(不套 kind/site 外壳)—— 刻意留的口子
	bare := `{"name":"手写站点","host":"10.9.9.9","user":"u","password":"p"}`
	if err := os.WriteFile(filepath.Join(dir, "bare.json"), []byte(bare), 0o600); err != nil {
		t.Fatal(err)
	}
	// 3) 非 json 文件不该进清单
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	rec, body := doJSON(t, s, "POST", "/api/hosts/import", map[string]any{"dir": dir})
	if rec.Code != 200 {
		t.Fatalf("status = %d (body %v)", rec.Code, body)
	}
	sites, _ := body["sites"].([]any)
	if len(sites) != 2 {
		t.Fatalf("sites = %v, want 2 (*.json only)", body["sites"])
	}
	var okCount, errCount int
	for _, it := range sites {
		m, _ := it.(map[string]any)
		if m["error"] != nil {
			errCount++
			continue
		}
		okCount++
		site, _ := m["site"].(map[string]any)
		if site["name"] != "手写站点" {
			t.Errorf("bare site = %v", site)
		}
		// 端口缺省要补成 22,否则导进来的环境是个连不上的 "host:0"。
		if site["port"] != float64(22) {
			t.Errorf("port = %v, want 22", site["port"])
		}
	}
	if okCount != 1 || errCount != 1 {
		t.Fatalf("ok=%d err=%d, want 1/1", okCount, errCount)
	}
}

func TestHostsExportRejectsUnknownNameAndBadDir(t *testing.T) {
	s, _ := newTestServer(t, seededConfig)
	dir := t.TempDir()

	rec, body := doJSON(t, s, "POST", "/api/hosts/export", map[string]any{"name": "不存在", "dir": dir})
	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404 (body %v)", rec.Code, body)
	}
	rec, body = doJSON(t, s, "POST", "/api/hosts/export", map[string]any{"name": "开发环境", "dir": ""})
	if rec.Code != 400 {
		t.Fatalf("empty dir status = %d, want 400 (body %v)", rec.Code, body)
	}
	rec, body = doJSON(t, s, "POST", "/api/hosts/import", map[string]any{"dir": filepath.Join(dir, "nope")})
	if rec.Code != 400 {
		t.Fatalf("missing dir status = %d, want 400 (body %v)", rec.Code, body)
	}
}

