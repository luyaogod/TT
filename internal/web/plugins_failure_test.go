package web

// P5-19：插件页面/RPC 坏掉（甚至 panic）不能扩散到宿主路由。
//
// 为什么值得单独一条：插件子系统与本服务在**同一个进程、同一个端口**上。
// 一个坏掉的插件清单、一个 panic 的插件 handler，最坏的后果不是"插件页打不开"，
// 而是把 /api/health、设置页、SPA 一起带走 —— 那时用户连"去关掉那个插件"的界面都进不去。
//
// 这一组用例把"插件那一侧坏到什么程度"逐层加深：
//   1. 记录缺失 / 清单坏掉 / 服务没起：插件路径给可读错误，宿主路由全 200；
//   2. 插件 handler 直接 panic：宿主仍然健康（panic 只断掉那一条连接）。

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/plugin"
)

// brokenPluginServer 造一个"装了、启用了、但版本目录已经被弄坏"的插件，
// 并把插件子资源 handler 挂进统一服务。
//
// 坏法刻意选最常见的两种现场：清单 JSON 被改坏（升级/手工编辑失败）、页面入口被删。
func brokenPluginServer(t *testing.T) (*Server, *plugin.Manager, string) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	if err := config.SaveHosts(cfgPath, config.Hosts{SSHs: []config.NamedSsh{
		{Name: "prod", SSHConfig: config.SSHConfig{Host: "10.0.0.1", Port: 22, User: "u"}},
	}}); err != nil {
		t.Fatal(err)
	}
	m := plugin.NewManager(plugin.NewStore(config.LocationsAt(cfgPath)), []string{"serve", "plugin"})
	m.ConfigPath = cfgPath

	src := filepath.Join(t.TempDir(), "src")
	writeTestFile(t, filepath.Join(src, "plugin.json"), `{
  "schemaVersion": 1,
  "id": "com.example.broken",
  "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "name": "坏掉的插件",
  "contributions": {
    "frontend": {"route": "broken", "title": "坏插件", "entry": "ui/index.html"},
    "service": {"entry": "bin/svc.exe", "args": ["service"]}
  },
  "permissions": ["profiles.list"]
}`)
	writeTestFile(t, filepath.Join(src, "ui", "index.html"), "<!doctype html><title>坏插件</title>页面")
	writeTestFile(t, filepath.Join(src, "bin", "svc.exe"), "x")

	rec, err := m.Install(plugin.InstallOptions{Source: src})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable(rec.ID, plugin.Grants{Capabilities: []string{"profiles.list"}, Envs: []string{"prod"}}); err != nil {
		t.Fatal(err)
	}

	// 弄坏它：清单写成非法 JSON，页面入口删掉。
	dir := m.Store.VersionDir(rec.ID, rec.ActiveVersion)
	writeTestFile(t, filepath.Join(dir, "plugin.json"), "{ 这不是 JSON")
	if err := os.Remove(filepath.Join(dir, "ui", "index.html")); err != nil {
		t.Fatal(err)
	}

	ui := plugin.NewWebHandler(m.Store, nil)
	s := New(Options{
		Version:    "test",
		ConfigPath: cfgPath,
		PluginMgr:  m,
		PluginUI:   ui.Handler(),
		// 一个很小的 SPA：用来证明 /debug/ 的深链接回退仍然可用。
		DebugFS: fstest.MapFS{
			"index.html":      &fstest.MapFile{Data: []byte("<html>HOST-APP</html>")},
			"assets/app-1.js": &fstest.MapFile{Data: []byte("console.log(1)")},
		},
	})
	return s, m, rec.ID
}

// getHost 发一个本机 GET 请求（插件管理端点会校验 RemoteAddr）。
func getHost(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.RemoteAddr = "127.0.0.1:5000"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// assertHostRoutesHealthy 宿主自己的四条主路必须全部还能用。
func assertHostRoutesHealthy(t *testing.T, s *Server) {
	t.Helper()
	if rec := getHost(t, s.Handler(), "/api/health"); rec.Code != http.StatusOK ||
		!strings.Contains(rec.Body.String(), `"server":"tt-unified"`) {
		t.Fatalf("/api/health 被插件故障带坏了: %d（%s）", rec.Code, rec.Body.String())
	}
	// SPA 深链接：设置页的分区（插件管理就在这里）刷新时必须回入口 HTML。
	if rec := getHost(t, s.Handler(), "/debug/settings/plugins"); rec.Code != http.StatusOK ||
		!strings.Contains(rec.Body.String(), "HOST-APP") {
		t.Fatalf("/debug/ SPA 回退被插件故障带坏了: %d（%s）", rec.Code, rec.Body.String())
	}
	// 管理面自己也要能列出插件 —— 否则用户没法进设置页关掉那个坏插件。
	rec := getHost(t, s.Handler(), "/api/plugins")
	if rec.Code != http.StatusOK {
		t.Fatalf("/api/plugins 被插件故障带坏了: %d（%s）", rec.Code, rec.Body.String())
	}
	var list struct {
		OK      bool `json:"ok"`
		Plugins []struct {
			ID string `json:"id"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || !list.OK || len(list.Plugins) != 1 {
		t.Fatalf("列表形状不对（%v）: %s", err, rec.Body.String())
	}
}

// TestPluginFailureDoesNotLeakToHostRoutes 插件清单坏掉时：插件路径给可读 404，
// 宿主路由（健康检查 / SPA / 插件列表）全部照常 200。
func TestPluginFailureDoesNotLeakToHostRoutes(t *testing.T) {
	s, _, id := brokenPluginServer(t)
	assertHostRoutesHealthy(t, s)

	// 已装但清单坏掉的插件：可见错误页，不是空白 404。
	res := getHost(t, s.Handler(), "/plugins/"+id+"/ui/")
	if res.Code != http.StatusNotFound {
		t.Fatalf("清单坏掉该 404，得 %d（%s）", res.Code, res.Body.String())
	}
	body := res.Body.String()
	if !strings.Contains(body, "插件页面不可用") || !strings.Contains(body, "清单读不出来") {
		t.Fatalf("该给可读错误页，得 %s", body)
	}
	if ct := res.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("错误页该是 HTML（给人看），得 %q", ct)
	}

	// 记录根本不存在。
	res = getHost(t, s.Handler(), "/plugins/com.example.ghost/ui/")
	if res.Code != http.StatusNotFound || !strings.Contains(res.Body.String(), "没有这个插件") {
		t.Fatalf("未知插件该 404 + 说明，得 %d（%s）", res.Code, res.Body.String())
	}

	// 服务没起（本进程没有 supervisor）：受控 503，而不是把请求打到别处。
	req := httptest.NewRequest(http.MethodPost, "/plugins/"+id+"/rpc/echo", nil)
	req.RemoteAddr = "127.0.0.1:5000"
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	// 清单坏了不影响 serveRPC 先按注册表判断"有 service 但没运行" → 503。
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("服务没起该 503，得 %d（%s）", rec.Code, rec.Body.String())
	}

	// 一轮插件路径的失败之后，宿主仍然健康（顺序上也要证明"没被带坏"）。
	assertHostRoutesHealthy(t, s)
}

// TestPluginUIPanicStaysOnPluginPath 镜像情形：插件 handler 直接 panic。
//
// 这里必须用**真 http.Server**（httptest）而不是 httptest.NewRecorder：recorder 里
// panic 会直接抛给测试进程，而真实服务里 net/http 的 conn.serve 会 recover 掉它、
// 只关闭出问题的那一条连接。我们断言两件事：
//   - 那次请求拿不到正常响应（连接被服务端断开）；
//   - 紧接着宿主三条主路全部 200 —— panic 没有扩散，进程也没死。
//
// 为什么不在本层加 recover 中间件：插件 loader 已经在 iframe/sandbox 与错误页两层
// 挡住了大多数坏法；这里要证明的是"就算最坏的一层漏了，宿主也还在"。真在插件处理器
// 里 catch 住 panic 会把"插件代码有 bug"变成一个静默的 500，反而更难查。
func TestPluginUIPanicStaysOnPluginPath(t *testing.T) {
	s := New(Options{
		Version:    "test",
		ConfigPath: filepath.Join(t.TempDir(), "config.json"),
		PluginMgr:  plugin.NewManager(plugin.NewStore(config.LocationsAt(filepath.Join(t.TempDir(), "config.json"))), nil),
		PluginUI: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("模拟：插件 UI handler 崩溃")
		}),
		DebugFS: fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte("<html>HOST-APP</html>")}},
	})
	srv := httptest.NewUnstartedServer(s.Handler())
	// 默认的 ErrorLog 会把整段 panic 栈打到测试输出里（几十行噪音，且容易被误读成失败）。
	srv.Config.ErrorLog = log.New(io.Discard, "", 0)
	srv.Start()
	t.Cleanup(srv.Close)

	// 插件路径：连接被断开（客户端拿不到响应）。
	res, err := srv.Client().Get(srv.URL + "/plugins/com.example.demo/ui/")
	if res != nil {
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}
	if err == nil && res != nil && res.StatusCode < 500 {
		t.Fatalf("插件 handler panic 后不该正常返回: %d", res.StatusCode)
	}

	// 宿主三条主路：各开一条新连接，必须全部 200。
	for _, path := range []string{"/api/health", "/debug/settings/plugins", "/api/plugins"} {
		resp, gerr := srv.Client().Get(srv.URL + path)
		if gerr != nil {
			t.Fatalf("%s 在插件 panic 后取不到响应: %v", path, gerr)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s 被插件 panic 带坏了: %d（%s）", path, resp.StatusCode, body)
		}
	}
}
