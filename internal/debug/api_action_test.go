package debug

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"tt/internal/host"
)

// 五个"动作型"端点的**错误半边**：POST /api/wstest、POST /api/wslogs/debug、
// POST /api/dbsql、GET /api/wslogs、GET /api/wslogs/content。
//
// 为什么只测错误半边：它们的**快乐路径**要 `*host.SSHConn` —— 那个结构体的 `cli` 是私有
// 字段（host/ssh.go:18），要从外部伪造得实现 8 个方法、其中三个返回不可构造的类型
// （`*PTYSession` / `*sftp.Client`）。为一个测试付那个代价不划算，所以 Server 上开了个
// `dial` 字段（生产默认 host.Dial，行为一字不差），这里注入一个**必然失败**的实现。
//
// 买到的四类东西，都是端到端测不到的：
//   1. 请求体解析失败 → 400（而不是拿半个请求去拨号）
//   2. 拨号失败 → 500 且信封形状一致（"SSH 连接失败: %w"）
//   3. 未注册的 API 路径 → 404 + JSON 信封（**不是** 405，也不是 200 + HTML ——
//      路由里有一条 catch-all，Go 的 mux 就不会自己产生 405 了。见 hStatic 的注释）
//   4. **入参校验在拨号之前** —— hWSLogDebug 里那处注释专门说过的顺序性质
//      （"非法请求不该先把当前会话收口"）。今天没有任何东西守着它。

// errTestNoDial 是注入进去的拨号失败。文案刻意与真实失败不同，
// 这样测试断言里出现它就说明**确实走进了被测的那条分支**，而不是碰巧撞上别的错。
var errTestNoDial = errors.New("测试：不拨号")

// actionTestServer 起一个不拨真机的调试服务。
// 复用 mode_test.go 的 doReq / respError 发请求收错误文案。
func actionTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	s := NewServer(&Config{Listen: "127.0.0.1:0"}, nil, "")
	s.dial = func(host.SSHConfig) (*host.SSHConn, error) { return nil, errTestNoDial }
	hs := httptest.NewServer(s.Handler())
	t.Cleanup(hs.Close)
	return hs
}

// TestActionEndpointsRejectMalformedBody 请求体不是 JSON → 400，
// 且不能是"拨号失败"那种 500 —— 半个请求根本不该走到拨号那一步。
func TestActionEndpointsRejectMalformedBody(t *testing.T) {
	hs := actionTestServer(t)
	for _, path := range []string{"/api/wstest", "/api/wslogs/debug", "/api/dbsql"} {
		t.Run(path, func(t *testing.T) {
			resp := doReq(t, "POST", hs.URL+path, "", "{ 这不是 json")
			if resp.StatusCode != 400 {
				t.Fatalf("该退 400，得 %d（%s）", resp.StatusCode, respError(t, resp))
			}
			if msg := respError(t, resp); !strings.Contains(msg, "请求体解析失败") {
				t.Errorf("错误文案该说清是请求体的问题，得 %q", msg)
			}
		})
	}
}

// TestActionEndpointsReportDialFailure 拨号失败 → 500，信封是 {"ok":false,"error":…}。
//
// 这条钉的是信封形状：前端与 CLI 都按它解析，改了形状它们会静默拿不到错误原因。
func TestActionEndpointsReportDialFailure(t *testing.T) {
	hs := actionTestServer(t)
	cases := []struct{ method, path, body string }{
		{"POST", "/api/wstest", "{}"},
		{"POST", "/api/wslogs/debug", "{}"},
		{"POST", "/api/dbsql", "{}"},
		{"GET", "/api/wslogs", ""},
		{"GET", "/api/wslogs/content?rowid=1", ""},
	}
	for _, c := range cases {
		t.Run(c.method+" "+c.path, func(t *testing.T) {
			resp := doReq(t, c.method, hs.URL+c.path, "", c.body)
			if resp.StatusCode != 500 {
				t.Fatalf("拨号失败该退 500，得 %d（%s）", resp.StatusCode, respError(t, resp))
			}
			msg := respError(t, resp)
			if !strings.Contains(msg, "SSH 连接失败") {
				t.Errorf("文案该说清是 SSH 连不上，得 %q", msg)
			}
			// 注入的那句话必须出现在里面 —— 证明走的是我们替换的那个 dial。
			if !strings.Contains(msg, errTestNoDial.Error()) {
				t.Errorf("该是由注入的 dial 报的错（否则这条测试没测到它），得 %q", msg)
			}
		})
	}
}

// TestUnregisteredAPIPathsReturnJSON404 钉住"打错的 API 路径不是 200 + HTML"。
//
// 这条断言本来长得不一样：它原先写的是"用 GET 打 POST 端点应当 405"，而那是**错的**
// —— 实测 200，因为路由里那条 `mux.HandleFunc("/", s.hStatic)` 的 SPA 兜底把任何
// 未注册路径都吞掉，未构建前端时回"前端尚未构建"那页、构建过就回 index.html。
//
// 当时撤掉了断言而不是把 200 钉下来（把已知不对的行为写成契约更坏），并把它记成一条
// 待定项。现在兜底收窄了，断言以它本来的意图回来，只把期望值从 405 改成 404 ——
// 405 要手写一张注册表才拿得到，而 404 与 405 对调用方是同一件事。理由见 hStatic 注释。
//
// 两种前端状态都测，因为**构建过的那半更坏**：它回的是一张真的 index.html，
// 调用方看不出任何异样；未构建那半至少文案里写着"前端尚未构建"。
func TestUnregisteredAPIPathsReturnJSON404(t *testing.T) {
	built := func() *httptest.Server {
		web := fstest.MapFS{"index.html": {Data: []byte("<html>APP</html>")}}
		hs := httptest.NewServer(NewServer(&Config{Listen: "127.0.0.1:0"}, web, "").Handler())
		t.Cleanup(hs.Close)
		return hs
	}()
	servers := map[string]*httptest.Server{"前端未构建": actionTestServer(t), "前端已构建": built}

	cases := []struct{ method, path string }{
		{"GET", "/api/wstest"}, // 路径注册了，方法不对
		{"GET", "/api/根本不存在"},  // 路径压根没有
		{"POST", "/api/根本不存在"},
		{"GET", "/api"}, // 前缀本身
		{"GET", "/api/"},
		{"DELETE", "/api/sessions/xyz/nope"}, // 注册前缀之下的未知子路径
	}
	for name, hs := range servers {
		for _, c := range cases {
			t.Run(name+" "+c.method+" "+c.path, func(t *testing.T) {
				resp := doReq(t, c.method, hs.URL+c.path, "", "")
				if resp.StatusCode != http.StatusNotFound {
					t.Fatalf("未注册的 API 路径该退 404，得 %d（%s）", resp.StatusCode, respError(t, resp))
				}
				if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
					t.Errorf("该回 JSON 信封，Content-Type 得 %q", ct)
				}
				if msg := respError(t, resp); msg == "" {
					t.Errorf("信封里该有错误原因（前端与 CLI 按它显示）,否则等于没说")
				}
			})
		}
	}
}

// TestStaticFallbackStillServesSPA 是上一条的**反向判据**：收窄兜底不能顺手把 SPA 打掉。
//
// `/apifoo` 与 `/api2/x` 是刻意的边界 —— 判据是 `/api` 或 `/api/` 前缀，不是
// `strings.HasPrefix(p, "/api")` 那种松写法，否则这两个正常的客户端路由会被误判成 API。
func TestStaticFallbackStillServesSPA(t *testing.T) {
	web := fstest.MapFS{"index.html": {Data: []byte("<html>APP</html>")}}
	hs := httptest.NewServer(NewServer(&Config{Listen: "127.0.0.1:0"}, web, "").Handler())
	t.Cleanup(hs.Close)

	for _, p := range []string{"/", "/sessions/abc", "/apifoo", "/api2/x"} {
		resp := doReq(t, "GET", hs.URL+p, "", "")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s 该走 SPA 兜底得 200，得 %d", p, resp.StatusCode)
		}
	}
}

// TestWSLogDebugValidatesInputBeforeDialing 钉住 hWSLogDebug 的**判定顺序**。
//
// 那个 handler 里有一句注释专门说了这件事：
//
//	入参先校验:非法请求不该先把当前会话收口(旧顺序是收口后才在 LaunchReplay 里失败)
//
// 顺序是这样的：读体 → checkReplayOverride（入参上限）→ 会话说闸门 → 拨号 → 查库 → 收口旧会话。
// 所以给一个**超限的 request**，应当拿到 **400**（入参被拒），而**不是** 500（拨号失败）。
// 两者都是"失败"，只有状态码能把它们区分开 —— 一旦有人把校验挪到拨号之后，
// 这条会从 400 变成 500 而红。
func TestWSLogDebugValidatesInputBeforeDialing(t *testing.T) {
	hs := actionTestServer(t)

	// maxReplayPayload 是 256 KB，给一个超过它的入参。
	huge := strings.Repeat("x", maxReplayPayload+1)
	body := `{"rowid":"1","request":"` + huge + `"}`

	resp := doReq(t, "POST", hs.URL+"/api/wslogs/debug", "", body)
	if resp.StatusCode == 500 {
		t.Fatalf("超限入参该在拨号**之前**被拒（400），却走到了拨号（500）：%s", respError(t, resp))
	}
	if resp.StatusCode != 400 {
		t.Fatalf("该退 400，得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	if msg := respError(t, resp); !strings.Contains(msg, "上限") {
		t.Errorf("文案该说清是入参超限，得 %q", msg)
	}
}
