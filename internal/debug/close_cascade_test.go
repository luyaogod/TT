package debug

// 结束会话的级联收工（W3 / D5）。
//
// 这一组钉四件事：
//  1. 挂着界面会话时，**先调停止器、再断会话**（顺序错了会留下"会话没了但 GDC 还在"）；
//  2. 停止器失败**不阻止**断会话，但必须出现在 `warnings` 里（不假装成功）；
//  3. 没挂界面会话时**不调**停止器（不去碰别人的画面）；
//  4. 没接停止器（单独跑 `tt debug serve` 的老装配）时，如实告诉人手动收。

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// closeTestServer 起一个带一条空闲会话的服务，并按需注入画面查询与停止器。
func closeTestServer(t *testing.T, attached bool, stopper ScreenStopper) (*Server, *httptest.Server, string) {
	t.Helper()
	s, id := snapshotServerWith(t, "恒烁正式区", "apmt520", "本地会话")
	if attached {
		s.SetScreenLookup(func(ref string) (map[string]any, bool) {
			return map[string]any{"tag": "fep-close-1", "ref": ref, "kind": "rec"}, true
		})
	} else {
		s.SetScreenLookup(func(string) (map[string]any, bool) { return nil, false })
	}
	if stopper != nil {
		s.SetScreenStopper(stopper)
	}
	mux := http.NewServeMux()
	s.routes(mux)
	hs := newTestHTTPServer(t, mux)
	return s, hs, id
}

// closeResp 解出 close 的应答。
func closeResp(t *testing.T, resp *http.Response) (code int, body map[string]any) {
	t.Helper()
	body = map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("应答不是合法 JSON：%v", err)
	}
	return resp.StatusCode, body
}

// 核心：先停界面会话、再断会话，最后从池里移除；应答带 stoppedScreen。
func TestCloseStopsScreenBeforeRemovingSession(t *testing.T) {
	type call struct{ ref, tag string }
	var calls []call
	var inPoolWhenStopped bool

	// 三个变量先声明再用：停止器里要读 s 与 id，而它们由 closeTestServer 一起返回。
	var (
		s  *Server
		hs *httptest.Server
		id string
	)
	s, hs, id = closeTestServer(t, true, func(ref, tag string) error {
		calls = append(calls, call{ref, tag})
		// 停止器运行时，会话必须**还在池里** —— 否则就是"先删后收"，
		// 中间那一瞬别的调用方会拿到 404 而画面还活着。
		inPoolWhenStopped = s.mgr.Get(id) != nil
		return nil
	})

	resp := doReq(t, "POST", hs.URL+"/api/sessions/"+id+"/close", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("close 该退 200，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	code, body := closeResp(t, resp)
	if code != http.StatusOK {
		t.Fatalf("状态码 %d", code)
	}
	if len(calls) != 1 {
		t.Fatalf("停止器该被调一次，实得 %d 次：%v", len(calls), calls)
	}
	if calls[0].tag != "fep-close-1" || calls[0].ref == "" {
		t.Errorf("停止器拿到的参数不对：%+v", calls[0])
	}
	if !inPoolWhenStopped {
		t.Error("停止器运行时会话已经不在池里了 —— 顺序反了（该先收画面、再断会话）")
	}
	stopped, _ := body["stoppedScreen"].(map[string]any)
	if stopped == nil || stopped["tag"] != "fep-close-1" {
		t.Errorf("应答该带 stoppedScreen.tag=fep-close-1，实得 %v", body)
	}
	if _, bad := body["warnings"]; bad {
		t.Errorf("一切顺利时不该有 warnings：%v", body["warnings"])
	}
	if s.mgr.Get(id) != nil {
		t.Error("close 之后会话该从池里消失")
	}
}

// 停止器失败：照样断会话并移除，但原因进 warnings（不假装成功）。
func TestCloseWarnsWhenStopperFails(t *testing.T) {
	s, hs, id := closeTestServer(t, true, func(string, string) error {
		return http.ErrHandlerTimeout // 任意非 nil 错误
	})

	resp := doReq(t, "POST", hs.URL+"/api/sessions/"+id+"/close", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("收尾失败也要退 200（会话已经收了），得 %d", resp.StatusCode)
	}
	_, body := closeResp(t, resp)
	warns, _ := body["warnings"].([]any)
	if len(warns) == 0 {
		t.Fatalf("该在 warnings 里报出收界面会话失败：%v", body)
	}
	if msg, _ := warns[0].(string); !strings.Contains(msg, "fep-close-1") {
		t.Errorf("警告要带录制 ID（人要靠它去查）：%q", msg)
	}
	if s.mgr.Get(id) != nil {
		t.Error("会话该已被移除（收尾失败不阻止断会话）")
	}
}

// 没挂界面会话：不调停止器，也不去碰别人的画面。
func TestCloseSkipsStopperWhenNothingAttached(t *testing.T) {
	called := false
	_, hs, id := closeTestServer(t, false, func(string, string) error {
		called = true
		return nil
	})
	resp := doReq(t, "POST", hs.URL+"/api/sessions/"+id+"/close", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("close 该退 200，得 %d", resp.StatusCode)
	}
	if called {
		t.Error("没挂界面会话时不该调停止器")
	}
}

// 没接停止器（老装配）：仍然断会话，但如实告诉人手动收。
func TestCloseWithoutStopperWarnsToStopManually(t *testing.T) {
	_, hs, id := closeTestServer(t, true, nil)

	resp := doReq(t, "POST", hs.URL+"/api/sessions/"+id+"/close", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("close 该退 200，得 %d", resp.StatusCode)
	}
	_, body := closeResp(t, resp)
	warns, _ := body["warnings"].([]any)
	if len(warns) == 0 {
		t.Fatalf("没有停止器时必须报警告（否则人以为已经收干净了）：%v", body)
	}
	msg, _ := warns[0].(string)
	if !strings.Contains(msg, "tt recording stop") {
		t.Errorf("警告要给出可照做的命令：%q", msg)
	}
}
