package web

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/plugin"
)

// 插件管理 API：/api/plugins*。
//
// 这些端点只碰元数据与载荷，所以能真装真启用（不需要真插件程序）——
// 用临时数据目录里的一份最小插件包当夹具。

// writePluginFixture 造一个最小插件包目录。
func writePluginFixture(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "src")
	writeTestFile(t, filepath.Join(dir, "plugin.json"), `{
  "schemaVersion": 1,
  "id": "com.example.demo",
  "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "name": "演示插件",
  "contributions": {"cli": {"namespace": "demo", "entry": "bin/demo.exe", "summary": "演示"}},
  "permissions": ["profiles.list", "db.connect"]
}`)
	writeTestFile(t, filepath.Join(dir, "bin", "demo.exe"), "x")
	return dir
}

// writeTestFile 建父目录并写文件。
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// pluginTestServer 造一个带插件管理面的服务。
func pluginTestServer(t *testing.T) (*Server, *plugin.Manager) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	if err := config.SaveHosts(cfgPath, config.Hosts{SSHs: []config.NamedSsh{
		{Name: "prod", SSHConfig: config.SSHConfig{Host: "10.0.0.1", Port: 22, User: "u"}},
	}}); err != nil {
		t.Fatal(err)
	}
	m := plugin.NewManager(plugin.NewStore(config.LocationsAt(cfgPath)), []string{"serve", "plugin"})
	m.ConfigPath = cfgPath
	return New(Options{Version: "test", ConfigPath: cfgPath, PluginMgr: m}), m
}

// TestPluginAPILifecycle 安装 → 列表 → 启用 → 停用 → 卸载 的 HTTP 契约。
func TestPluginAPILifecycle(t *testing.T) {
	s, _ := pluginTestServer(t)
	src := writePluginFixture(t)

	rec, body := doJSON(t, s, "POST", "/api/plugins/install", map[string]any{"source": src})
	if rec.Code != 200 {
		t.Fatalf("安装该成功，得 %d（%v）", rec.Code, body)
	}
	if body["ok"] != true || body["id"] != "com.example.demo" || body["enabled"] != false {
		t.Fatalf("安装响应不对: %v", body)
	}

	rec, body = doJSON(t, s, "GET", "/api/plugins", nil)
	if rec.Code != 200 {
		t.Fatalf("列表该成功，得 %d", rec.Code)
	}
	if body["count"] != float64(1) {
		t.Fatalf("该有 1 个插件，得 %v", body)
	}

	rec, body = doJSON(t, s, "POST", "/api/plugins/com.example.demo/enable", map[string]any{
		"capabilities": []string{"profiles.list"},
	})
	if rec.Code != 200 || body["enabled"] != true {
		t.Fatalf("启用该成功，得 %d（%v）", rec.Code, body)
	}

	// 带未知能力 → 400（不是 500，也不是静默接受）。
	rec, _ = doJSON(t, s, "POST", "/api/plugins/com.example.demo/enable", map[string]any{
		"capabilities": []string{"root.all"},
	})
	if rec.Code != 400 {
		t.Fatalf("未知能力该 400，得 %d", rec.Code)
	}

	rec, body = doJSON(t, s, "GET", "/api/plugins/com.example.demo", nil)
	if rec.Code != 200 {
		t.Fatalf("详情该成功，得 %d", rec.Code)
	}
	if body["plugin"] == nil {
		t.Fatalf("详情该带 plugin，得 %v", body)
	}

	rec, body = doJSON(t, s, "POST", "/api/plugins/com.example.demo/disable", nil)
	if rec.Code != 200 || body["enabled"] != false {
		t.Fatalf("停用该成功，得 %d（%v）", rec.Code, body)
	}

	rec, body = doJSON(t, s, "POST", "/api/plugins/com.example.demo/uninstall", map[string]any{})
	if rec.Code != 200 || body["ok"] != true {
		t.Fatalf("卸载该成功，得 %d（%v）", rec.Code, body)
	}
	rec, body = doJSON(t, s, "GET", "/api/plugins/com.example.demo", nil)
	if rec.Code != 404 {
		t.Fatalf("卸载后详情该 404，得 %d（%v）", rec.Code, body)
	}
}

// TestPluginAPIRejectsBadRequests 请求体坏/来源不存在要给出明确状态码。
func TestPluginAPIRejectsBadRequests(t *testing.T) {
	s, _ := pluginTestServer(t)

	if rec, _ := doJSON(t, s, "POST", "/api/plugins/install", map[string]any{}); rec.Code != 400 {
		t.Fatalf("空 source 该 400，得 %d", rec.Code)
	}
	if rec, _ := doJSON(t, s, "POST", "/api/plugins/install", map[string]any{"source": "Z:/nope"}); rec.Code != 500 && rec.Code != 400 {
		t.Fatalf("来源不存在该 4xx/5xx，得 %d", rec.Code)
	}
	if rec, _ := doJSON(t, s, "POST", "/api/plugins/nope/enable", map[string]any{}); rec.Code != 404 {
		t.Fatalf("不存在的插件该 404，得 %d", rec.Code)
	}
}

// TestPluginAPIDoctor 自检端点返回问题清单且形状稳定。
func TestPluginAPIDoctor(t *testing.T) {
	s, m := pluginTestServer(t)
	src := writePluginFixture(t)
	if _, err := m.Install(plugin.InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	rec, body := doJSON(t, s, "GET", "/api/plugins/com.example.demo/doctor", nil)
	if rec.Code != 200 {
		t.Fatalf("自检该 200，得 %d", rec.Code)
	}
	if body["ok"] != true {
		t.Fatalf("健康插件该 ok:true，得 %v", body)
	}

	// 把版本目录删掉 → 自检该报问题（ok:false）。
	if err := os.RemoveAll(m.Store.VersionDir("com.example.demo", "0.1.0")); err != nil {
		t.Fatal(err)
	}
	_, body = doJSON(t, s, "GET", "/api/plugins/com.example.demo/doctor", nil)
	if body["ok"] != false {
		t.Fatalf("缺版本目录该 ok:false，得 %v", body)
	}
	issues, _ := body["issues"].([]any)
	if len(issues) == 0 {
		t.Fatalf("该有问题清单，得 %v", body)
	}
	first, _ := issues[0].(map[string]any)
	if first["code"] != "VERSION_DIR_MISSING" {
		t.Fatalf("问题码不对: %v", first)
	}
}

// TestPluginAPIUnavailableWhenNotWired 未装配管理面时是 503（让人知道"没接上"）。
func TestPluginAPIUnavailableWhenNotWired(t *testing.T) {
	s := New(Options{Version: "test"})
	req, _ := http.NewRequest("GET", "/api/plugins", nil)
	rec := newRecorder()
	s.Handler().ServeHTTP(rec, req)
	// 未装配时**没有注册**这些路由，落到 /api/ 未命中路径（无 Dict 时是 404）。
	if c := rec.Code(); c != http.StatusNotFound && c != http.StatusServiceUnavailable {
		t.Fatalf("未装配时该 404/503，得 %d", c)
	}
}

// newRecorder 用一个最小 ResponseRecorder（避免为此 import httptest）。
func newRecorder() *recorder { return &recorder{header: http.Header{}} }

type recorder struct {
	header http.Header
	code   int
	body   []byte
}

func (r *recorder) Header() http.Header { return r.header }
func (r *recorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}
func (r *recorder) WriteHeader(code int) { r.code = code }
func (r *recorder) Code() int {
	if r.code == 0 {
		return 200
	}
	return r.code
}

// TestPluginJSONShapeIsStable 列表里的字段名是对外契约（前端按它渲染）。
func TestPluginJSONShapeIsStable(t *testing.T) {
	s, m := pluginTestServer(t)
	src := writePluginFixture(t)
	if _, err := m.Install(plugin.InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	rec, _ := doJSON(t, s, "GET", "/api/plugins", nil)
	var payload struct {
		OK      bool              `json:"ok"`
		Count   int               `json:"count"`
		Plugins []json.RawMessage `json:"plugins"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.OK || payload.Count != 1 || len(payload.Plugins) != 1 {
		t.Fatalf("形状不对: %s", rec.Body.String())
	}
	var one map[string]any
	if err := json.Unmarshal(payload.Plugins[0], &one); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "version", "enabled", "namespace"} {
		if _, ok := one[key]; !ok {
			t.Errorf("列表项缺字段 %q: %s", key, rec.Body.String())
		}
	}
	if !strings.Contains(rec.Body.String(), `"namespace":"demo"`) {
		t.Errorf("该带上 CLI 命名空间: %s", rec.Body.String())
	}
}
