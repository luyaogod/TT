package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/pkg/erpdb"
	"github.com/luyaogod/TT/internal/host"
	"github.com/luyaogod/TT/internal/sshtun"
)

// 三个"动作"端点：POST /api/dbprobe、POST /api/dbaccverify、POST /api/conntest。
//
// 它们都要连服务器/连库，所以默认档的测试进不去 —— 而这三条路径上的**错误分支**
// （HTTP 码、信封形状、stage 取值）恰恰是对外契约，改了没有任何东西会响。
//
// conntest 现在走 connsvc（与第三方插件的 ConnectRPC 是同一份实现），
// 所以这里注入的是连接服务最底层的那次"打开会话"，而不是整个 Connector：
// 隧道改写、阶段归类、关闭收口这些都在被测范围内。

// fakeSession 是 erpdb.Session 的假实现。只实现测试要的那几条，
// 其余返回零值 —— 这里不假装能验"库怎么答"，只验我们的信封。
type fakeSession struct {
	typ      string
	ver      string
	verErr   error
	readOnly bool
	closed   bool
}

func (f *fakeSession) Type() string { return f.typ }
func (f *fakeSession) ServerVersion(context.Context) (string, error) {
	return f.ver, f.verErr
}
func (f *fakeSession) Query(context.Context, string) ([]string, [][]string, error) {
	return nil, nil, nil
}
func (f *fakeSession) Exec(context.Context, string) (int64, error) { return 0, nil }
func (f *fakeSession) ReadOnly() bool                              { return f.readOnly }
func (f *fakeSession) Close()                                      { f.closed = true }

// actionServer 造一个注入了假外部依赖的服务。
func actionServer(t *testing.T, opt Options) *Server {
	t.Helper()
	return New(opt)
}

func decodeAll(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	out := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("响应不是合法 JSON：%v\n%s", err, rec.Body.String())
	}
	return out
}

// TestActionEndpointsRejectMalformedBody 请求体不是 JSON → 400。
// 三个端点的这一条路径都不该因为"体坏了"去连服务器/连库。
func TestActionEndpointsRejectMalformedBody(t *testing.T) {
	called := false
	s := actionServer(t, Options{
		ProbeDB:   func(host.DBProbeReq) (*host.DBProbeOut, error) { called = true; return nil, nil },
		VerifyAcc: func(host.DBAccVerifyReq) error { called = true; return nil },
		ConnSvc: testConnSvc(t, connsvc.Options{
			OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
				called = true
				return nil, nil
			},
		}),
	})
	for _, path := range []string{"/api/dbprobe", "/api/dbaccverify", "/api/conntest"} {
		t.Run(path, func(t *testing.T) {
			called = false
			req := httptest.NewRequest("POST", path, bytes.NewReader([]byte("{ 这不是 json")))
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)
			if rec.Code != 400 {
				t.Fatalf("该退 400，得 %d（%s）", rec.Code, rec.Body.String())
			}
			if called {
				t.Error("请求体都没解析成功，不该去连服务器/连库")
			}
		})
	}
}

// TestDBProbeErrorsAreBadGateway 探测失败 → **502**（不是 500，也不是 200）。
func TestDBProbeErrorsAreBadGateway(t *testing.T) {
	boom := errors.New("探测不了：测试用")
	s := actionServer(t, Options{
		ProbeDB: func(host.DBProbeReq) (*host.DBProbeOut, error) { return nil, boom },
	})
	req := httptest.NewRequest("POST", "/api/dbprobe", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != 502 {
		t.Fatalf("该退 502，得 %d（%s）", rec.Code, rec.Body.String())
	}
	got := decodeAll(t, rec)
	if got["error"] != boom.Error() {
		t.Errorf("错误文案该原样透出，得 %v", got["error"])
	}
}

func TestDBProbeSuccessIsOK(t *testing.T) {
	want := &host.DBProbeOut{}
	s := actionServer(t, Options{
		ProbeDB: func(host.DBProbeReq) (*host.DBProbeOut, error) { return want, nil },
	})
	req := httptest.NewRequest("POST", "/api/dbprobe", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("成功该退 200，得 %d（%s）", rec.Code, rec.Body.String())
	}
}

// TestDBAccVerifyErrorsAreBadGateway 账号校验失败 → 502；成功 → 200 + {"ok":true}。
func TestDBAccVerifyErrorsAreBadGateway(t *testing.T) {
	boom := errors.New("账号不对：测试用")
	s := actionServer(t, Options{VerifyAcc: func(host.DBAccVerifyReq) error { return boom }})

	req := httptest.NewRequest("POST", "/api/dbaccverify", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 502 {
		t.Fatalf("失败该退 502，得 %d（%s）", rec.Code, rec.Body.String())
	}
	if got := decodeAll(t, rec); got["error"] != boom.Error() {
		t.Errorf("错误文案该原样透出，得 %v", got["error"])
	}
}

func TestDBAccVerifySuccessIsOK(t *testing.T) {
	s := actionServer(t, Options{VerifyAcc: func(host.DBAccVerifyReq) error { return nil }})
	req := httptest.NewRequest("POST", "/api/dbaccverify", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("成功该退 200，得 %d", rec.Code)
	}
	if got := decodeAll(t, rec); got["ok"] != true {
		t.Errorf("该报 ok:true，得 %v", got)
	}
}

// TestConnTestFailuresAreResultsNotHTTPErrors 连通性测试的失败是**结果**，不是 HTTP 错误：
// 一律 200 + {"ok":false,"stage":…}。
//
// 这条设计是刻意的 —— "连不上"正是这个端点要回答的问题，用 502 表达会让前端把
// "测试结果是连不上"当成"请求本身失败了"，两件事在界面上长得完全不一样。
// 所以 stage 的取值也是契约：调用方靠它区分卡在连接还是卡在版本查询。
func TestConnTestFailuresAreResultsNotHTTPErrors(t *testing.T) {
	t.Run("连不上：stage=connect", func(t *testing.T) {
		boom := errors.New("连不上：测试用")
		s := actionServer(t, Options{
			ConnSvc: testConnSvc(t, connsvc.Options{
				OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
					return nil, boom
				},
			}),
		})
		rec, got := doJSON(t, s, "POST", "/api/conntest", map[string]any{})
		if rec.Code != 200 {
			t.Fatalf("该是 200（失败是结果不是 HTTP 错误），得 %d", rec.Code)
		}
		if got["ok"] != false || got["stage"] != "connect" {
			t.Errorf("该是 ok:false stage:connect，得 %v", got)
		}
		if got["error"] != boom.Error() {
			t.Errorf("错误文案该原样透出，得 %v", got["error"])
		}
	})

	t.Run("连上了但版本查不出：stage=version", func(t *testing.T) {
		boom := errors.New("版本查不出：测试用")
		fc := &fakeSession{typ: "oracle", verErr: boom}
		s := actionServer(t, Options{
			ConnSvc: testConnSvc(t, connsvc.Options{
				OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
					return fc, nil
				},
			}),
		})
		rec, got := doJSON(t, s, "POST", "/api/conntest", map[string]any{})
		if rec.Code != 200 {
			t.Fatalf("该是 200，得 %d", rec.Code)
		}
		if got["ok"] != false || got["stage"] != "version" {
			t.Errorf("该是 ok:false stage:version，得 %v", got)
		}
		// 连上了就该收口 —— 失败路径上漏掉 Close 会攒着连接不放。
		if !fc.closed {
			t.Error("版本查询失败之后也该把连接关掉")
		}
	})
}

// TestConnTestSuccessIsTheOnlyHappyPathInThisFile 这三个端点里唯一能测全的一条：
// 假 Session 走完 Open → ServerVersion → Close。
func TestConnTestSuccessIsTheOnlyHappyPathInThisFile(t *testing.T) {
	fc := &fakeSession{typ: "oracle", ver: "Oracle Database 11g"}
	s := actionServer(t, Options{
		ConnSvc: testConnSvc(t, connsvc.Options{
			OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
				return fc, nil
			},
		}),
	})
	rec, got := doJSON(t, s, "POST", "/api/conntest", map[string]any{
		"connection": map[string]any{"type": "oracle", "host": "db1", "service": "t35prd"},
	})
	if rec.Code != 200 {
		t.Fatalf("成功该退 200，得 %d（%s）", rec.Code, rec.Body.String())
	}
	if got["ok"] != true {
		t.Fatalf("该报成功，得 %v", got)
	}
	if got["type"] != "oracle" || got["serverVersion"] != "Oracle Database 11g" {
		t.Errorf("该把类型与版本带回来，得 %v", got)
	}
	// address 由**连接服务算**（隧道改写之后的实际地址），所以它能如实反映"连的是哪儿"。
	if got["address"] != "db1:1521/t35prd" {
		t.Errorf("address 该反映实际连接地址，得 %v", got["address"])
	}
	if !fc.closed {
		t.Error("成功路径必须把连接关掉（否则一次页面点一下就漏一条连接）")
	}
}

// TestConnTestUsesTunnelWhenConfigured 配了 viaSsh 时，测试必须经隧道同一条链路。
//
// 这是"测试走的链路 = 实际查询走的链路"那条要求的守卫：隧道建立之后驱动连的是
// 127.0.0.1:<本地端口>，而 address 也要如实反映这一点。
func TestConnTestUsesTunnelWhenConfigured(t *testing.T) {
	var gotConn dbconfig.Connection
	s := actionServer(t, Options{
		ConnSvc: testConnSvc(t, connsvc.Options{
			DialTunnel: func(o sshtun.Options) (connsvc.Tunnel, error) {
				return &fakeTunnel{port: 45678}, nil
			},
			OpenDB: func(_ context.Context, c dbconfig.Connection, _ erpdb.SessionOptions) (erpdb.Session, error) {
				gotConn = c
				return &fakeSession{typ: "oracle", ver: "v"}, nil
			},
		}),
	})
	rec, got := doJSON(t, s, "POST", "/api/conntest", map[string]any{
		"connection": map[string]any{
			"type": "oracle", "host": "db1", "port": 1521, "service": "t35prd",
			"viaSsh": map[string]any{"host": "jump", "port": 22, "user": "u", "password": "p"},
		},
	})
	if rec.Code != 200 || got["ok"] != true {
		t.Fatalf("该成功，得 %d %v", rec.Code, got)
	}
	if gotConn.Host != "127.0.0.1" || gotConn.Port != 45678 {
		t.Fatalf("连接该被改写到隧道本地端口，得 %s:%d", gotConn.Host, gotConn.Port)
	}
	if got["via"] != "tunnel" {
		t.Errorf("via 该是 tunnel，得 %v", got["via"])
	}
	if got["address"] != "127.0.0.1:45678/t35prd" {
		t.Errorf("address 该是隧道本地地址，得 %v", got["address"])
	}
}

// TestActionEndpointsUseTheRealImplWhenNotInjected 三个字段留空 = 用真实现。
//
// 这条不打真机：只确认**没注入时不会 panic**，且确实走到了真实现（那一步的失败
// 是"连不上"，正是真实现会报的错）。它守的是"nil 回落"这个约定本身 ——
// 少了回落，所有没注入的调用点会直接空指针崩。
func TestActionEndpointsUseTheRealImplWhenNotInjected(t *testing.T) {
	s := actionServer(t, Options{}) // 全 nil

	// conntest 用空连接：真 erpdb 会因没配账号而失败，落到 stage=connect。
	rec, got := doJSON(t, s, "POST", "/api/conntest", map[string]any{})
	if rec.Code != 200 {
		t.Fatalf("该是 200（失败是结果），得 %d。若这里是 502/500，说明 nil 回落坏了",
			rec.Code)
	}
	if got["ok"] != false {
		t.Errorf("空连接不该连上，得 %v", got)
	}
}
