package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Handler 是 tt serve 的挂载点:用 StripPrefix("/dict", …) 挂在 /dict/ 下,
// 包内路径相对挂载点(/api/bdldoc → /dict/api/bdldoc)。
func TestHandlerMountedUnderPrefix(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"activeEnv":"a","sshs":[{"name":"a","host":"1.2.3.4"}]}}`)
	s := New(p)

	outer := http.NewServeMux()
	outer.Handle("/dict/", http.StripPrefix("/dict", s.Handler()))

	// 字典自己的端点
	rec := httptest.NewRecorder()
	outer.ServeHTTP(rec, httptest.NewRequest("GET", "/dict/api/bdldoc", nil))
	if rec.Code != 200 {
		t.Fatalf("GET /dict/api/bdldoc code=%d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"configPath"`) {
		t.Fatalf("应返回 bdldoc 状态: %s", rec.Body.String())
	}

	// 未命中 API 的路径落到兜底页(页面本身由 tt serve 提供)
	rec2 := httptest.NewRecorder()
	outer.ServeHTTP(rec2, httptest.NewRequest("GET", "/dict/pages/whatever", nil))
	if rec2.Code != 200 || !strings.Contains(rec2.Body.String(), "<!doctype html>") {
		t.Fatalf("兜底页应返回接口说明: code=%d body=%s", rec2.Code, rec2.Body.String())
	}
}

// 共享给 internal/web 的端点不在本包注册(同进程里两边都有会撞名):
// /api/dbprobe、/api/dbaccverify、/api/conntest、/api/status、/api/config
// 落到兜底页,而不是本包的 JSON 接口。
func TestSharedEndpointsNotServedHere(t *testing.T) {
	s := New(writeTempConfig(t, `{}`))
	h := s.Handler()

	for _, path := range []string{
		"/api/dbprobe", "/api/dbaccverify", "/api/conntest", "/api/status", "/api/config",
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("POST", path, nil))
		if strings.Contains(rec.Body.String(), `"configPath"`) {
			t.Errorf("%s 不该由本包提供: %s", path, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "<!doctype html>") {
			t.Errorf("%s 应落到兜底页, got %s", path, rec.Body.String())
		}
	}
}
