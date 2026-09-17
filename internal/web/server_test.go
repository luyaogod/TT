package web

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"tt/internal/config"
)

// newTestServer 造一个只带共享 API 的服务（不接子系统），指向临时配置。
func newTestServer(t *testing.T, seed string) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if seed != "" {
		if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return New(Options{ConfigPath: path, Version: "test"}), path
}

func doJSON(t *testing.T, s *Server, method, target string, body any) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, target, &buf)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec, out
}

const seededConfig = `{
  "schemaVersion": 2,
  "listen": "127.0.0.1:28670",
  "hosts": {
    "activeEnv": "开发环境",
    "sshs": [{
      "name": "开发环境", "host": "10.0.0.1", "port": 22, "user": "u", "password": "p",
      "db": {
        "type": "oracle", "host": "10.0.0.2", "port": 1521, "service": "t100dev",
        "viaSsh": {"host": "10.0.0.1", "port": 22, "user": "u", "password": "p", "bindPort": 11521},
        "accounts": [{"account": "ds", "password": "x"}]
      }
    }]
  },
  "query": {"source": "local"},
  "debug": {"termWidth": 240},
  "未来的键": {"keep": true}
}`

// 配置页保存不能把表单不管理的 viaSsh 抹掉 —— 否则用户配好的 SSH 隧道
// 会在"打开设置页点一下保存"后静默消失。
func TestHostsPut_PreservesViaSSH(t *testing.T) {
	s, path := newTestServer(t, seededConfig)

	// 模拟配置页提交：它有 type/host/port/service/accounts，但没有 viaSsh
	body := map[string]any{
		"hosts": map[string]any{
			"activeEnv": "开发环境",
			"sshs": []any{map[string]any{
				"name": "开发环境", "host": "10.0.0.1", "port": 22, "user": "u", "password": "p",
				"db": map[string]any{
					"type": "oracle", "host": "10.0.0.9", "port": 1521, "service": "t100dev",
					"accounts": []any{map[string]any{"account": "ds", "password": "x"}},
				},
			}},
		},
	}
	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT 失败 HTTP %d: %s", rec.Code, rec.Body.String())
	}

	root, err := config.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	hosts, _ := root["hosts"].(map[string]any)
	sshs, _ := hosts["sshs"].([]any)
	if len(sshs) != 1 {
		t.Fatalf("环境数 = %d", len(sshs))
	}
	db, _ := sshs[0].(map[string]any)["db"].(map[string]any)

	// 表单改的字段要生效
	if db["host"] != "10.0.0.9" {
		t.Errorf("表单提交的 db.host 未生效: %v", db["host"])
	}
	// 表单不管的字段要保留
	via, _ := db["viaSsh"].(map[string]any)
	if via == nil {
		t.Fatal("viaSsh 被抹掉了 —— 用户配的 SSH 隧道静默丢失")
	}
	if via["bindPort"] != float64(11521) {
		t.Errorf("viaSsh 内容不完整: %v", via)
	}
}

// PUT /api/hosts 只该动自己带的节，其他顶层键（含未知键）原样保留。
func TestHostsPut_LeavesOtherSectionsAlone(t *testing.T) {
	s, path := newTestServer(t, seededConfig)

	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
		"debug": map[string]any{"termWidth": 300},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT 失败 HTTP %d: %s", rec.Code, rec.Body.String())
	}

	root, _ := config.Open(path)
	if _, has := root["未来的键"]; !has {
		t.Error("未知顶层键被抹掉了")
	}
	q, _ := root["query"].(map[string]any)
	if q["source"] != "local" {
		t.Error("本次没提交的 query 节被改动了")
	}
	d, _ := root["debug"].(map[string]any)
	if d["termWidth"] != float64(300) {
		t.Errorf("debug.termWidth 未写入: %v", d["termWidth"])
	}
	l, _ := root["listen"].(string)
	if l != "127.0.0.1:28670" {
		t.Errorf("listen 被误改: %q", l)
	}
}

// 校验拦在保存之前，且错误信息是给人看的中文。
func TestHostsPut_Validation(t *testing.T) {
	cases := []struct {
		name   string
		sshs   []any
		active string
		want   string
	}{
		{"空列表", []any{}, "", "至少要配置一个 SSH 环境"},
		{"缺名称", []any{map[string]any{"host": "h", "user": "u"}}, "", "缺少名称"},
		{"重名", []any{
			map[string]any{"name": "a", "host": "h", "user": "u"},
			map[string]any{"name": "a", "host": "h2", "user": "u"},
		}, "", "重复"},
		{"缺主机", []any{map[string]any{"name": "a", "user": "u"}}, "", "缺少服务器地址"},
		{"端口越界", []any{map[string]any{"name": "a", "host": "h", "user": "u", "port": 70000}}, "", "超出范围"},
		{"缺账号", []any{map[string]any{"name": "a", "host": "h"}}, "", "缺少登录账号"},
		{"库类型非法", []any{map[string]any{
			"name": "a", "host": "h", "user": "u",
			"db": map[string]any{"type": "mysql", "host": "d"},
		}}, "", "不支持"},
		{"默认环境不存在", []any{map[string]any{"name": "a", "host": "h", "user": "u"}}, "b", "不在环境列表里"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, _ := newTestServer(t, seededConfig)
			rec, out := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
				"hosts": map[string]any{"activeEnv": c.active, "sshs": c.sshs},
			})
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("应当被拒，却得到 HTTP %d: %s", rec.Code, rec.Body.String())
			}
			msg, _ := out["error"].(string)
			if !contains(msg, c.want) {
				t.Errorf("错误信息 %q 里没有 %q", msg, c.want)
			}
		})
	}
}

// 校验失败不能落盘 —— 一次失败的保存不该把配置改坏。
func TestHostsPut_RejectedWriteDoesNotTouchFile(t *testing.T) {
	s, path := newTestServer(t, seededConfig)
	before, _ := os.ReadFile(path)

	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
		"hosts": map[string]any{"activeEnv": "", "sshs": []any{}},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("应当被拒, 得到 HTTP %d", rec.Code)
	}

	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Error("被拒绝的写入改动了配置文件")
	}
}

// GET /api/hosts 返回设置页需要的那几个节。
func TestHostsGet(t *testing.T) {
	s, _ := newTestServer(t, seededConfig)
	rec, out := doJSON(t, s, http.MethodGet, "/api/hosts", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP %d", rec.Code)
	}
	for _, k := range []string{"config", "activeEnv", "sshs", "listen", "debug", "query", "mirror", "bdldoc", "sync", "tdev"} {
		if _, has := out[k]; !has {
			t.Errorf("响应缺少 %q", k)
		}
	}
	if out["activeEnv"] != "开发环境" {
		t.Errorf("activeEnv = %v", out["activeEnv"])
	}
}

// 没有前端构建产物时不能把请求吞掉，而是给出引导页。
func TestRouting_NoSubsystems(t *testing.T) {
	s, _ := newTestServer(t, seededConfig)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/debug/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("/debug/ → HTTP %d, 期望 200（引导页）", rec.Code)
	}
	body := rec.Body.String()
	if !contains(body, "前端尚未构建") {
		t.Errorf("/debug/ 没有给出引导页")
	}
	// 标题也要在：这条兜底路径上有过一版把 title 丢了，只剩一个光秃秃的「TT」——
	// 而那正是全新克隆第一次访问看到的页面。
	if !contains(body, "界面未构建") {
		t.Errorf("/debug/ 引导页丢了标题: %.120s", body)
	}
}

// 字典页那套 SPA 已经并进统一设置页，/dict/ 不再存在。
func TestRouting_DictPageRemoved(t *testing.T) {
	s, _ := newTestServer(t, seededConfig)
	for _, p := range []string{"/dict/", "/dict/api/mirror"} {
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s → HTTP %d, 期望 404（字典页已移除）", p, rec.Code)
		}
	}
}

// SPA 的静态服务：挂载点、深链接、真实资源各走各的路，都不能被重定向掉前缀。
func TestSPAHandler(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":      {Data: []byte("<html>APP</html>")},
		"assets/index.js": {Data: []byte("console.log(1)")},
		"favicon.ico":     {Data: []byte("ico")},
	}
	h := SPAHandler(fsys, "/debug/", "未构建")

	cases := []struct {
		path     string
		wantCode int
		wantBody string
	}{
		// 挂载点本身要给入口 HTML，不能 301 到 "/"（那会丢掉 /debug 前缀）
		{"/debug/", http.StatusOK, "APP"},
		// 深链接（前端路由）回落入口 HTML，而不是 404
		{"/debug/some/deep/link", http.StatusOK, "APP"},
		// 目录同样给入口 HTML
		{"/debug/assets", http.StatusOK, "APP"},
		// 真实文件原样返回
		{"/debug/assets/index.js", http.StatusOK, "console.log"},
		{"/debug/favicon.ico", http.StatusOK, "ico"},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, c.path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.wantCode {
			t.Errorf("%s → HTTP %d, 期望 %d", c.path, rec.Code, c.wantCode)
			continue
		}
		if !contains(rec.Body.String(), c.wantBody) {
			t.Errorf("%s 响应体不含 %q: %.60s", c.path, c.wantBody, rec.Body.String())
		}
	}
}

// 前端没构建（只有占位文件）时给引导页，而不是 404 或 panic。
func TestSPAHandler_NotBuilt(t *testing.T) {
	for _, fsys := range []fs.FS{nil, fstest.MapFS{".gitkeep": {Data: []byte("")}}} {
		h := SPAHandler(fsys, "/debug/", "调试工作台未构建")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/debug/", nil))
		if rec.Code != http.StatusOK {
			t.Errorf("HTTP %d, 期望 200 引导页", rec.Code)
		}
		if !contains(rec.Body.String(), "前端尚未构建") {
			t.Errorf("没有给出引导页: %.80s", rec.Body.String())
		}
	}
}

// 与 viaSsh 同源的另一处:每环境的调试覆盖项(launchArgs / watchdogSeconds)表单里
// 没有输入控件,保存站点会把它们丢掉。服务端按环境名补回。
func TestHostsPut_PreservesEnvDebugOverrides(t *testing.T) {
	seed := `{
  "schemaVersion": 2,
  "hosts": {"activeEnv": "开发环境", "sshs": [{
    "name": "开发环境", "host": "10.0.0.1", "port": 22, "user": "u", "password": "p",
    "zone": "36", "topent": "10001",
    "launchArgs": "BBDL512840855a 9 9 'Y' {prog}",
    "watchdogSeconds": 300
  }]}
}`
	s, path := newTestServer(t, seed)

	// 配置页提交:它只有 SSH/DB 那些字段,没有这两个覆盖项
	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
		"hosts": map[string]any{
			"activeEnv": "开发环境",
			"sshs": []any{map[string]any{
				"name": "开发环境", "host": "10.0.0.9", "port": 22, "user": "u", "password": "p",
				"zone": "36", "topent": "10001",
			}},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT 失败 HTTP %d: %s", rec.Code, rec.Body.String())
	}

	root, _ := config.Open(path)
	hosts, _ := root["hosts"].(map[string]any)
	sshs, _ := hosts["sshs"].([]any)
	e, _ := sshs[0].(map[string]any)

	if e["host"] != "10.0.0.9" {
		t.Errorf("表单改的 host 未生效: %v", e["host"])
	}
	if e["launchArgs"] != "BBDL512840855a 9 9 'Y' {prog}" {
		t.Errorf("每环境的 launchArgs 被抹掉了: %v", e["launchArgs"])
	}
	if e["watchdogSeconds"] != float64(300) {
		t.Errorf("每环境的 watchdogSeconds 被抹掉了: %v", e["watchdogSeconds"])
	}
}

// 配置文件不存在时也要能答,设置页在首次运行下就得渲染。
func TestConfigStatus_NoConfig(t *testing.T) {
	s, _ := newTestServer(t, "")
	rec, out := doJSON(t, s, http.MethodGet, "/api/config/status", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP %d", rec.Code)
	}
	if out["ok"] != true {
		t.Error("ok 应为 true")
	}
	for _, k := range []string{"config", "mirror", "bdldoc", "sync", "install"} {
		if _, has := out[k]; !has {
			t.Errorf("响应缺少 %q", k)
		}
	}
}

// 派生状态要真的去查磁盘:目录/文件存不存在只有服务端算得出来。
func TestConfigStatus_DerivedFromDisk(t *testing.T) {
	dir := t.TempDir()
	realDir := filepath.Join(dir, "mirror")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatal(err)
	}
	missingDir := filepath.Join(dir, "nope")
	syncFile := filepath.Join(dir, "erp_data.db")
	if err := os.WriteFile(syncFile, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	seed := `{"schemaVersion":2,"hosts":{"sshs":[]},
	  "mirror":{"dir":` + jsonStr(realDir) + `},
	  "bdldoc":{"dir":` + jsonStr(missingDir) + `},
	  "sync":{"target":` + jsonStr(syncFile) + `}}`
	s, _ := newTestServer(t, seed)

	rec, out := doJSON(t, s, http.MethodGet, "/api/config/status", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP %d", rec.Code)
	}
	mirror, _ := out["mirror"].(map[string]any)
	if mirror["exists"] != true {
		t.Errorf("存在的镜像目录应报 exists=true: %v", mirror)
	}
	bdldoc, _ := out["bdldoc"].(map[string]any)
	if bdldoc["exists"] != false {
		t.Errorf("不存在的目录应报 exists=false: %v", bdldoc)
	}
	sync, _ := out["sync"].(map[string]any)
	if sync["exists"] != true {
		t.Errorf("存在的同步目标应报 exists=true: %v", sync)
	}
	if sync["configured"] == "" || sync["defaultTarget"] == "" {
		t.Errorf("应同时给出显式配置值与缺省位置: %v", sync)
	}
}

func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && bytes.Contains([]byte(s), []byte(sub)))
}
