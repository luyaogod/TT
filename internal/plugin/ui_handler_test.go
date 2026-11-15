package plugin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/hostrpc"
)

// uiFixture 装一个带页面（可选带服务）的插件并启用，返回 handler。
func uiFixture(t *testing.T, withService bool) (http.Handler, *Store, *Record) {
	t.Helper()
	m, s := newManager(t)
	src := t.TempDir()
	manifest := `{
  "schemaVersion": 1, "id": "com.example.ui", "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "contributions": {
    "frontend": {"route": "ui", "title": "UI 插件", "entry": "ui/index.html"}%s
  }
}`
	svc := ""
	if withService {
		svc = `,
    "service": {"entry": "bin/svc.exe", "healthPath": "/healthz"}`
	}
	writeFile(t, filepath.Join(src, ManifestFileName), strings.Replace(manifest, "%s", svc, 1))
	writeFile(t, filepath.Join(src, "ui", "index.html"), "<!doctype html><title>p</title>插件页面")
	writeFile(t, filepath.Join(src, "ui", "assets", "app-DCXJThG6.js"), "console.log(1)")
	writeFile(t, filepath.Join(src, "ui", "note.txt"), "hello")
	if withService {
		writeFile(t, filepath.Join(src, "bin", "svc.exe"), "x")
	}
	rec, err := m.Install(InstallOptions{Source: src})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}
	h := NewWebHandler(s, nil)
	return h.Handler(), s, rec
}

// do 发一个本机请求。
func do(h http.Handler, method, target string, remote string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	req.RemoteAddr = remote
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// TestUIHandlerServesPageWithSecurityHeaders 页面从活跃版本目录读出，并带安全头。
func TestUIHandlerServesPageWithSecurityHeaders(t *testing.T) {
	h, _, _ := uiFixture(t, false)
	rec := do(h, "GET", "/plugins/com.example.ui/ui/", "127.0.0.1:5000")
	if rec.Code != 200 {
		t.Fatalf("status = %d（%s）", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "插件页面") {
		t.Fatalf("没读到页面内容: %s", rec.Body.String())
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("nosniff = %q", got)
	}
	csp := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "connect-src 'none'") || !strings.Contains(csp, "frame-ancestors 'self'") {
		t.Errorf("CSP 不对: %q", csp)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("index.html 该 no-cache，得 %q", got)
	}

	// 带哈希的静态资源可长缓存；普通文件不长缓存。
	hashed := do(h, "GET", "/plugins/com.example.ui/ui/assets/app-DCXJThG6.js", "127.0.0.1:5000")
	if hashed.Code != 200 || !strings.Contains(hashed.Header().Get("Cache-Control"), "immutable") {
		t.Errorf("带哈希资源该 immutable，得 %d %q", hashed.Code, hashed.Header().Get("Cache-Control"))
	}
	plain := do(h, "GET", "/plugins/com.example.ui/ui/note.txt", "127.0.0.1:5000")
	if plain.Code != 200 || plain.Header().Get("Cache-Control") != "" {
		t.Errorf("普通文件不该长缓存: %d %q", plain.Code, plain.Header().Get("Cache-Control"))
	}
}

// TestUIHandlerRejectsTraversal 资源路径不能越出插件目录。
func TestUIHandlerRejectsTraversal(t *testing.T) {
	h, s, rec := uiFixture(t, false)
	// 在插件目录外放一个"秘密"文件。
	secret := filepath.Join(filepath.Dir(s.VersionDir(rec.ID, rec.ActiveVersion)), "secret.txt")
	writeFile(t, secret, "TOP-SECRET")

	for _, target := range []string{
		"/plugins/com.example.ui/ui/../../secret.txt",
		"/plugins/com.example.ui/ui/%2e%2e/%2e%2e/secret.txt",
		"/plugins/com.example.ui/ui/..%2f..%2fsecret.txt",
	} {
		res := do(h, "GET", target, "127.0.0.1:5000")
		if res.Code == 200 && strings.Contains(res.Body.String(), "TOP-SECRET") {
			t.Fatalf("路径穿越读到了插件目录外的文件: %s", target)
		}
	}
}

// TestUIHandlerErrorsAreVisible 未启用/没有页面/资源缺失都要给可读的错误页。
func TestUIHandlerErrorsAreVisible(t *testing.T) {
	h, s, rec := uiFixture(t, false)

	// 未启用的插件。
	if _, err := NewManager(s, nil).Disable(rec.ID); err != nil {
		t.Fatal(err)
	}
	res := do(h, "GET", "/plugins/com.example.ui/ui/", "127.0.0.1:5000")
	if res.Code != 404 || !strings.Contains(res.Body.String(), "未启用") {
		t.Fatalf("未启用该给 404 + 说明，得 %d（%s）", res.Code, res.Body.String())
	}

	// 不存在的插件。
	res = do(h, "GET", "/plugins/com.example.nope/ui/", "127.0.0.1:5000")
	if res.Code != 404 || !strings.Contains(res.Body.String(), "没有这个插件") {
		t.Fatalf("不存在该 404，得 %d", res.Code)
	}
}

// TestUIHandlerRPCDisabledWithoutService 没有 service 贡献时明确拒绝，而不是把请求打到别处。
func TestUIHandlerRPCDisabledWithoutService(t *testing.T) {
	h, _, _ := uiFixture(t, false)
	res := do(h, "POST", "/plugins/com.example.ui/rpc/echo", "127.0.0.1:5000")
	if res.Code != http.StatusConflict {
		t.Fatalf("没有服务该 409，得 %d（%s）", res.Code, res.Body.String())
	}
}

// TestUIHandlerRPCProxiesToOwnInstance 只转发到该插件自己的实例，并带上实例令牌。
func TestUIHandlerRPCProxiesToOwnInstance(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer inst-token" {
			t.Errorf("该带上实例令牌，得 %q", got)
		}
		if got := r.Header.Get("Cookie"); got != "" {
			t.Errorf("不该把宿主 Cookie 转给插件，得 %q", got)
		}
		if r.URL.Path != "/echo" {
			t.Errorf("转发路径 = %q, want /echo", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer backend.Close()

	h, s, _ := uiFixture(t, true)
	u := strings.TrimPrefix(backend.URL, "http://")
	h = NewWebHandler(s, func(id string) (string, string, bool) { return u, "inst-token", true }).Handler()

	req := httptest.NewRequest("POST", "/plugins/com.example.ui/rpc/echo", nil)
	req.RemoteAddr = "127.0.0.1:5000"
	req.Header.Set("Cookie", "session=leak")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("代理失败: %d（%s）", rec.Code, rec.Body.String())
	}
}

// TestUIHandlerRPCNeverProxiesToRemote 非回环目标一律拒绝（不做通用本机代理）。
func TestUIHandlerRPCNeverProxiesToRemote(t *testing.T) {
	h, s, _ := uiFixture(t, true)
	h = NewWebHandler(s, func(id string) (string, string, bool) {
		return "10.1.2.3:8080", "", true
	}).Handler()
	res := do(h, "POST", "/plugins/com.example.ui/rpc/x", "127.0.0.1:5000")
	if res.Code != http.StatusForbidden {
		t.Fatalf("非回环目标该 403，得 %d", res.Code)
	}
}

// TestUIHandlerRPCRequiresLocalOrigin 跨站来源与外部远端都不能调代理。
func TestUIHandlerRPCRequiresLocalOrigin(t *testing.T) {
	h, _, _ := uiFixture(t, true)
	res := do(h, "POST", "/plugins/com.example.ui/rpc/x", "10.0.0.9:5000")
	if res.Code != http.StatusForbidden {
		t.Fatalf("非本机远端该 403，得 %d", res.Code)
	}

	req := httptest.NewRequest("POST", "/plugins/com.example.ui/rpc/x", nil)
	req.RemoteAddr = "127.0.0.1:5000"
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("外站 Origin 该 403，得 %d", rec.Code)
	}
}

// TestUIHandlerSwitchesWithActiveVersion 升级切换活跃版本后，页面来自新版本（不串版）。
func TestUIHandlerSwitchesWithActiveVersion(t *testing.T) {
	h, s, rec := uiFixture(t, false)

	// 手工造一个新版本目录并把它设为活跃版本（模拟升级后的状态）。
	oldDir := s.VersionDir(rec.ID, rec.ActiveVersion)
	newDir := s.VersionDir(rec.ID, "0.2.0")
	writeFile(t, filepath.Join(newDir, ManifestFileName), `{
  "schemaVersion": 1, "id": "com.example.ui", "version": "0.2.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "contributions": {"frontend": {"route": "ui", "title": "UI 插件", "entry": "ui/index.html"}}
}`)
	writeFile(t, filepath.Join(newDir, "ui", "index.html"), "新版本页面")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := s.WithLock(func(reg *Registry) error {
		r := reg.Find(rec.ID)
		r.ActiveVersion = "0.2.0"
		r.Versions = append(r.Versions, InstalledVersion{Version: "0.2.0"})
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	res := do(h, "GET", "/plugins/com.example.ui/ui/", "127.0.0.1:5000")
	if res.Code != 200 || !strings.Contains(res.Body.String(), "新版本页面") {
		t.Fatalf("该从新版本目录取页面，得 %d（%s）", res.Code, res.Body.String())
	}
}

// TestUIHandlerEndToEndHostAPIAndOwnService 参考插件「页面 + 常驻服务」的端到端链路（P5-17）：
//
//	安装 → 启用（授权 profiles.list + prod）→ supervisor 起服务 →
//	(a) 页面经桥 /plugins/{id}/hostrpc/tt.host.v1.ProfilesService/List 拿到**已授权环境**；
//	(b) 页面经 /plugins/{id}/rpc/echo 打到插件**自己的服务**。
//
// 边界：这里用 httptest 直接调 WebHandler，**没有真浏览器**。所以它验的是"页面那条
// HTTP 路径 + 服务端身份/授权"是真的，不验 iframe 的 sandbox/消息桥（那部分由
// web/app/scripts/plugins.test.mjs 的纯 Node 用例覆盖）。两者合起来才是完整链路。
func TestUIHandlerEndToEndHostAPIAndOwnService(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{Capabilities: []string{"profiles.list"}, Envs: []string{"prod"}}); err != nil {
		t.Fatal(err)
	}

	cfgPath := s.Locations().ConfigPath
	if err := config.SaveHosts(cfgPath, config.Hosts{SSHs: []config.NamedSsh{
		{Name: "prod", SSHConfig: config.SSHConfig{Host: "10.0.0.1", Port: 22, User: "u"}},
		{Name: "test", SSHConfig: config.SSHConfig{Host: "10.0.0.2", Port: 22, User: "u"}},
	}}); err != nil {
		t.Fatal(err)
	}
	tokens := hostrpc.NewTokenStore()
	conn := connsvc.New(connsvc.Options{ConfigPath: cfgPath, Authorizer: RegistryAuthorizer{Store: s}})
	sup := NewSupervisor(s, tokens, conn, "", "test")
	sup.Logf = t.Logf
	sup.StartTimeout = 10 * time.Second
	sup.HealthTimeout = 10 * time.Second
	sup.StopGrace = 2 * time.Second
	t.Cleanup(sup.StopAll)
	if _, err := sup.Start(context.Background(), rec); err != nil {
		t.Fatalf("起插件服务失败: %v", err)
	}

	// 真的 Host API（与 tt serve 同构：同一份 connsvc + 同一张令牌表）。
	api := httptest.NewServer(hostrpc.NewHandler(hostrpc.Deps{
		Service: conn, Auth: hostrpc.NewTokenAuthenticator(tokens),
	}))
	t.Cleanup(api.Close)

	h := NewWebHandler(s, sup.Endpoint)
	h.HostAPIURL = api.URL + hostrpc.Prefix
	h.Tokens = tokens
	handler := h.Handler()

	// 页面本身能取到（活跃版本目录里的参考插件页面）。
	page := do(handler, "GET", "/plugins/"+rec.ID+"/ui/", "127.0.0.1:5000")
	if page.Code != 200 || !strings.Contains(page.Body.String(), "参考插件页面") {
		t.Fatalf("插件页面没取到: %d（%s）", page.Code, page.Body.String())
	}

	// (a) 经桥调 Host API：只返回该插件**获授权**的环境。
	req := httptest.NewRequest("POST",
		"/plugins/"+rec.ID+"/hostrpc/tt.host.v1.ProfilesService/List", strings.NewReader("{}"))
	req.RemoteAddr = "127.0.0.1:5000"
	req.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	if rec2.Code != 200 {
		t.Fatalf("桥接调 Host API 失败: %d（%s）", rec2.Code, rec2.Body.String())
	}
	var list struct {
		Profiles []struct {
			Env string `json:"env"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &list); err != nil {
		t.Fatalf("Host API 响应不是预期 JSON: %v（%s）", err, rec2.Body.String())
	}
	var envs []string
	for _, p := range list.Profiles {
		envs = append(envs, p.Env)
	}
	if len(envs) != 1 || envs[0] != "prod" {
		t.Fatalf("插件页面只该看到已授权的 prod，得 %v", envs)
	}
	// 桥用完即撤销：令牌表里只剩服务实例自己那一个。
	if tokens.Count() != 1 {
		t.Errorf("桥接后该只剩服务实例令牌，得 %d", tokens.Count())
	}

	// (b) 经 /rpc/ 打到插件自己的服务（参考插件 /echo 回自己的标识）。
	rpc := do(handler, "POST", "/plugins/"+rec.ID+"/rpc/echo", "127.0.0.1:5000")
	if rpc.Code != 200 || !strings.Contains(rpc.Body.String(), `"plugin":"refplugin"`) {
		t.Fatalf("插件自身服务没有回答: %d（%s）", rpc.Code, rpc.Body.String())
	}
}

// TestUIHandlerHostBridgeMintsIdentityForThePlugin 页面经桥调 Host API：桥按插件身份
// 签发一次性令牌，页面自己拿不到令牌。
func TestUIHandlerHostBridgeMintsIdentityForThePlugin(t *testing.T) {
	m, s := newManager(t)
	src := t.TempDir()
	writeFile(t, filepath.Join(src, ManifestFileName), `{
  "schemaVersion": 1, "id": "com.example.ui", "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "contributions": {"frontend": {"route": "ui", "title": "UI", "entry": "ui/index.html"}},
  "permissions": ["profiles.list"]
}`)
	writeFile(t, filepath.Join(src, "ui", "index.html"), "page")
	rec, err := m.Install(InstallOptions{Source: src})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable(rec.ID, Grants{Capabilities: []string{"profiles.list"}, Envs: []string{"prod"}}); err != nil {
		t.Fatal(err)
	}

	// 一个真的 Host API（与 serve 里同构：同一份 connsvc + 令牌表）。
	tokens := hostrpc.NewTokenStore()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	if err := config.SaveHosts(cfgPath, config.Hosts{SSHs: []config.NamedSsh{
		{Name: "prod", SSHConfig: config.SSHConfig{Host: "10.0.0.1", Port: 22, User: "u"}},
	}}); err != nil {
		t.Fatal(err)
	}
	svc := connsvc.New(connsvc.Options{ConfigPath: cfgPath, Authorizer: RegistryAuthorizer{Store: s}})
	api := httptest.NewServer(hostrpc.NewHandler(hostrpc.Deps{Service: svc, Auth: hostrpc.NewTokenAuthenticator(tokens)}))
	defer api.Close()

	h := NewWebHandler(s, nil)
	h.HostAPIURL = api.URL + hostrpc.Prefix
	h.Tokens = tokens
	handler := h.Handler()

	req := httptest.NewRequest("POST", "/plugins/com.example.ui/hostrpc/tt.host.v1.ProfilesService/List", strings.NewReader("{}"))
	req.RemoteAddr = "127.0.0.1:5000"
	req.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	if rec2.Code != 200 {
		t.Fatalf("桥接调用失败: %d（%s）", rec2.Code, rec2.Body.String())
	}
	if !strings.Contains(rec2.Body.String(), "\"prod\"") {
		t.Fatalf("该拿到已授权环境: %s", rec2.Body.String())
	}
	// 一次性令牌用完即撤销。
	if tokens.Count() != 0 {
		t.Errorf("桥接后令牌该撤销，剩 %d", tokens.Count())
	}
}
