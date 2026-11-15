package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestPluginManagementRejectsNonLocalAndCrossSite 插件管理只接受本机请求、本机来源。
//
// 为什么要单列：这些端点会**安装并启用可执行代码**，一个恶意网页向 127.0.0.1 发
// POST 就能做到这件事，代价太高。挡的不是"本机进程"（那是文档写明的信任边界），
// 而是"别的网页替用户点了这个按钮"。
func TestPluginManagementRejectsNonLocalAndCrossSite(t *testing.T) {
	s, _ := pluginTestServer(t)

	cases := []struct {
		name   string
		remote string
		origin string
		want   int
	}{
		{"回环地址无 Origin：放行", "127.0.0.1:5000", "", http.StatusOK},
		{"回环地址本机 Origin：放行", "127.0.0.1:5000", "http://127.0.0.1:28670", http.StatusOK},
		{"非回环远端：拒绝", "192.168.1.9:5000", "", http.StatusForbidden},
		{"回环远端但外站 Origin：拒绝", "127.0.0.1:5000", "http://evil.example", http.StatusForbidden},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/plugins", nil)
			req.RemoteAddr = c.remote
			if c.origin != "" {
				req.Header.Set("Origin", c.origin)
			}
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)
			if rec.Code != c.want {
				t.Fatalf("status = %d, want %d（%s）", rec.Code, c.want, rec.Body.String())
			}
		})
	}
}

// TestPluginManagementGuardCoversWrites 写端点同样过这道门（不是只有列表）。
func TestPluginManagementGuardCoversWrites(t *testing.T) {
	s, _ := pluginTestServer(t)
	for _, path := range []string{
		"/api/plugins/install",
		"/api/plugins/com.example.demo/enable",
		"/api/plugins/com.example.demo/disable",
		"/api/plugins/com.example.demo/uninstall",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("POST", path, nil)
			req.RemoteAddr = "10.1.2.3:5000"
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("非本机来源该 403，得 %d", rec.Code)
			}
		})
	}
}
