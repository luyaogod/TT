package db

// P6-10 的守卫：`tt db ping` 与第三方插件经 Host API 的 TestDB 是**同一份实现**。
//
// 为什么单列一条：这条链路以前经 internal/dict/live 自己建隧道、自己开连接，与设置页
// 和 Host API 的连接测试是两套编排。"两套"不会在功能测试里露出来 —— 只有把两条入口
// 指向同一个 Service 实例、同一份注入依赖，再比对结果与调用次数，才证明真的只有一条。

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"tt/internal/config"
	"tt/internal/connsvc"
	"tt/internal/dbconfig"
	"tt/internal/erpdb"
	"tt/internal/gen/hostv1"
	"tt/internal/gen/hostv1/hostv1connect"
	"tt/internal/hostrpc"
)

// pingSession erpdb.Session 的替身：只回答取版本。
type pingSession struct {
	typ     string
	version string
	opened  *int
	closed  bool
}

func (s *pingSession) Type() string { return s.typ }
func (s *pingSession) ServerVersion(context.Context) (string, error) {
	return s.version, nil
}
func (s *pingSession) Query(context.Context, string) ([]string, [][]string, error) {
	return nil, nil, nil
}
func (s *pingSession) Exec(context.Context, string) (int64, error) { return 0, nil }
func (s *pingSession) ReadOnly() bool                              { return true }
func (s *pingSession) Close()                                      { s.closed = true }

// pingAuthz 内置调用方全通；插件只被授权 prod 的 db.connect。
type pingAuthz struct{}

func (pingAuthz) Permitted(c connsvc.Caller, cap connsvc.Capability) bool {
	if c.IsBuiltin() {
		return true
	}
	return cap == connsvc.CapDBConnect
}

func (pingAuthz) AllowedEnvs(c connsvc.Caller, cap connsvc.Capability) ([]string, bool) {
	if c.IsBuiltin() {
		return nil, true
	}
	if cap == connsvc.CapDBConnect {
		return []string{"prod"}, false
	}
	return nil, false
}

// pingRT 给每个请求挂上实例令牌。
type pingRT struct {
	base  http.RoundTripper
	token string
}

func (t pingRT) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}

// writePingConfig 造一份只含一个环境的配置。
func writePingConfig(t *testing.T, withAccounts bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	conn := &dbconfig.Connection{
		Type: "oracle", Host: "db1", Port: 1521, Service: "t35prd",
		Accounts: []dbconfig.DBAcct{{Account: "ds", Password: "p"}},
	}
	if !withAccounts {
		conn.Accounts = nil
	}
	hosts := config.Hosts{SSHs: []config.NamedSsh{{Name: "prod", DB: conn}}}
	if err := config.SaveHosts(path, hosts); err != nil {
		t.Fatalf("写测试配置失败: %v", err)
	}
	return path
}

// TestPingAndRPCHitTheSameConnSvc 本地 ping 与 RPC TestDB 结果一致、且都进同一次实现。
func TestPingAndRPCHitTheSameConnSvc(t *testing.T) {
	cfgPath := writePingConfig(t, true)
	opened := 0
	svc := connsvc.New(connsvc.Options{
		ConfigPath: cfgPath,
		Authorizer: pingAuthz{},
		OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			opened++
			return &pingSession{typ: "oracle", version: "Oracle Database 19c", opened: &opened}, nil
		},
	})

	// 1) 第一方直连：tt db ping 走的就是这个函数。
	conn := dbconfig.Connection{
		Type: "oracle", Host: "db1", Port: 1521, Service: "t35prd",
		Accounts: []dbconfig.DBAcct{{Account: "ds", Password: "p"}},
	}
	local := runDBPing(context.Background(), svc, conn)
	if !local.OK {
		t.Fatalf("本地 ping 该成功: %s (%s)", local.Message, local.Stage)
	}

	// 2) 第三方入口：同一个 Service 实例走 ConnectRPC。
	tokens := hostrpc.NewTokenStore()
	h := hostrpc.NewHandler(hostrpc.Deps{Service: svc, Auth: hostrpc.NewTokenAuthenticator(tokens)})
	ts := httptest.NewServer(h)
	defer ts.Close()
	tok := tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p", InstanceID: "i"})
	hc := &http.Client{Transport: pingRT{base: ts.Client().Transport, token: tok}}
	client := hostv1connect.NewConnectionsServiceClient(hc, ts.URL+"/hostrpc/")
	rpc, err := client.TestDB(context.Background(), connect.NewRequest(&hostv1.TestDBRequest{Env: "prod"}))
	if err != nil {
		t.Fatalf("RPC TestDB 失败: %v", err)
	}

	if got, want := rpc.Msg.GetOk(), local.OK; got != want {
		t.Errorf("ok 不一致: rpc=%v local=%v", got, want)
	}
	if got, want := rpc.Msg.GetServerVersion(), local.ServerVersion; got != want {
		t.Errorf("版本不一致: rpc=%q local=%q", got, want)
	}
	if got, want := rpc.Msg.GetAddress(), local.Address; got != want {
		t.Errorf("地址不一致: rpc=%q local=%q", got, want)
	}
	if got, want := rpc.Msg.GetStage(), local.Stage; got != want {
		t.Errorf("阶段不一致: rpc=%q local=%q", got, want)
	}
	// 两条入口各开一次会话 —— 两者都进了 connsvc.TestDBConn，没有旁路。
	if opened != 2 {
		t.Errorf("期望两条入口各开一次会话（共 2 次），实际 %d 次：说明有一边没走 connsvc", opened)
	}
}

// TestPingFailureKeepsStageAndTarget 连接层失败要带"卡在哪一步"，且不吞目标地址。
func TestPingFailureKeepsStageAndTarget(t *testing.T) {
	cfgPath := writePingConfig(t, true)
	svc := connsvc.New(connsvc.Options{
		ConfigPath: cfgPath,
		Authorizer: pingAuthz{},
		OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			return nil, errPingDial
		},
	})
	conn := dbconfig.Connection{Type: "oracle", Host: "db1", Port: 1521, Service: "t35prd"}
	res := runDBPing(context.Background(), svc, conn)
	if res.OK {
		t.Fatal("连不上时不该报 OK")
	}
	if res.Stage != connsvc.StageConnect {
		t.Errorf("阶段该是 %s，得 %q", connsvc.StageConnect, res.Stage)
	}
	msg := pingFailure(res)
	if !strings.Contains(msg, "db1") || !strings.Contains(msg, "connect") {
		t.Errorf("失败描述该带阶段与目标地址，得 %q", msg)
	}
}

// errPingDial 模拟建连失败。
var errPingDial = &pingErr{"连接被拒绝"}

type pingErr struct{ s string }

func (e *pingErr) Error() string { return e.s }
