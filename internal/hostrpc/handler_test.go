package hostrpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/pkg/erpdb"
	"github.com/luyaogod/TT/pkg/hostv1"
	"github.com/luyaogod/TT/pkg/hostv1/hostv1connect"
)

// ---- 夹具 ----

func testConfigPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	hosts := config.Hosts{SSHs: []config.NamedSsh{
		{
			Name:      "prod",
			SSHConfig: config.SSHConfig{Host: "10.0.0.1", Port: 22, User: "u", Password: "ssh-secret"},
			DB: &dbconfig.Connection{
				Type: "oracle", Host: "db1", Port: 1521, Service: "t35prd",
				Accounts: []dbconfig.DBAcct{{Account: "ds", Password: "db-secret"}},
			},
		},
		{
			Name:      "test",
			SSHConfig: config.SSHConfig{Host: "10.0.0.2", Port: 22, User: "u", Password: "p2"},
			DB: &dbconfig.Connection{
				Type: "kingbase", Host: "db2", Port: 54321, Database: "topprd",
				Accounts: []dbconfig.DBAcct{{Account: "ds", Password: "p3"}},
			},
		},
	}}
	if err := config.SaveHosts(path, hosts); err != nil {
		t.Fatalf("写测试配置失败: %v", err)
	}
	return path
}

// authz 可控授权：内置调用方全通；插件按 caps/envs。
type authz struct {
	caps map[connsvc.Capability]bool
	envs map[connsvc.Capability][]string
}

func (a authz) Permitted(c connsvc.Caller, cap connsvc.Capability) bool {
	if c.IsBuiltin() {
		return true
	}
	return a.caps[cap]
}
func (a authz) AllowedEnvs(c connsvc.Caller, cap connsvc.Capability) ([]string, bool) {
	if c.IsBuiltin() {
		return nil, true
	}
	return a.envs[cap], false
}

// fakeSession erpdb.Session 的替身。
type fakeSession struct {
	typ     string
	version string
	cols    []string
	rows    [][]string
	closed  bool
}

func (f *fakeSession) Type() string { return f.typ }
func (f *fakeSession) ServerVersion(context.Context) (string, error) {
	return f.version, nil
}
func (f *fakeSession) Query(context.Context, string) ([]string, [][]string, error) {
	return f.cols, f.rows, nil
}
func (f *fakeSession) Exec(context.Context, string) (int64, error) { return 1, nil }
func (f *fakeSession) ReadOnly() bool                              { return true }
func (f *fakeSession) Close()                                      { f.closed = true }

// server 一套跑起来的 Host API + 客户端 + 令牌表。
type server struct {
	svc      *connsvc.Service
	tokens   *TokenStore
	http     *httptest.Server
	profiles hostv1connect.ProfilesServiceClient
	conns    hostv1connect.ConnectionsServiceClient
	db       hostv1connect.DBSessionServiceClient
	hosts    hostv1connect.HostServiceClient
}

func newServer(t *testing.T, authz connsvc.Authorizer, open func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error)) *server {
	t.Helper()
	opt := connsvc.Options{ConfigPath: testConfigPath(t), Authorizer: authz, OpenDB: open}
	if opt.OpenDB == nil {
		opt.OpenDB = func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			return &fakeSession{typ: "oracle", version: "Oracle Database 11g"}, nil
		}
	}
	svc := connsvc.New(opt)
	tokens := NewTokenStore()
	h := NewHandler(Deps{Service: svc, Auth: NewTokenAuthenticator(tokens)})
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	hc := &http.Client{Transport: tokenTransport{base: ts.Client().Transport}}
	base := ts.URL + "/hostrpc/"
	return &server{
		svc:      svc,
		tokens:   tokens,
		http:     ts,
		profiles: hostv1connect.NewProfilesServiceClient(hc, base),
		conns:    hostv1connect.NewConnectionsServiceClient(hc, base),
		db:       hostv1connect.NewDBSessionServiceClient(hc, base),
		hosts:    hostv1connect.NewHostServiceClient(hc, base),
	}
}

// authed 返回带该令牌的调用上下文。
//
// 令牌走 context + 自定义 RoundTripper 注入（见 tokenTransport）：这样每个用例只拿
// 一个 ctx，不必为每个令牌各造一套 client。
func authed(token string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if token != "" {
		ctx = context.WithValue(ctx, ctxTokenKey{}, token)
	}
	return ctx, cancel
}

// ctxTokenKey context 中存放待注入令牌的键。
type ctxTokenKey struct{}

// tokenTransport 把 context 里的令牌塞进 Authorization 头。
type tokenTransport struct{ base http.RoundTripper }

func (t tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if tok, ok := req.Context().Value(ctxTokenKey{}).(string); ok && tok != "" {
		req = req.Clone(req.Context())
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	return t.base.RoundTrip(req)
}

// errInfo 从 Connect 错误里取 ErrorInfo 详情。
func errInfo(t *testing.T, err error) *hostv1.ErrorInfo {
	t.Helper()
	var ce *connect.Error
	if !errors.As(err, &ce) {
		t.Fatalf("不是 Connect 错误: %v", err)
	}
	for _, d := range ce.Details() {
		msg, derr := d.Value()
		if derr != nil {
			continue
		}
		if info, ok := msg.(*hostv1.ErrorInfo); ok {
			return info
		}
	}
	t.Fatalf("错误没有 ErrorInfo 详情: %v", ce)
	return nil
}

// ---- 认证 ----

// TestRequestsWithoutTokenAreUnauthenticated 没有令牌一律 401，且不泄露任何环境信息。
func TestRequestsWithoutTokenAreUnauthenticated(t *testing.T) {
	s := newServer(t, authz{}, nil)
	ctx, cancel := authed("")
	defer cancel()

	_, err := s.profiles.List(ctx, connect.NewRequest(&hostv1.ListRequest{}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("该是 Unauthenticated，得 %v", err)
	}
	if info := errInfo(t, err); info.GetCode() != codeUnauthenticated {
		t.Fatalf("稳定码该是 %s，得 %q", codeUnauthenticated, info.GetCode())
	}

	_, err = s.hosts.GetCapabilities(ctx, connect.NewRequest(&hostv1.GetCapabilitiesRequest{}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("能力探测也要令牌，得 %v", err)
	}
}

// TestBadTokenIsRejected 伪造令牌进不来。
func TestBadTokenIsRejected(t *testing.T) {
	s := newServer(t, authz{}, nil)
	ctx, cancel := authed("deadbeef")
	defer cancel()
	_, err := s.profiles.List(ctx, connect.NewRequest(&hostv1.ListRequest{}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("伪造令牌该被拒，得 %v", err)
	}
}

// TestRevokedTokenStopsWorking 撤销后旧令牌立刻失效（重启/禁用/撤权都走这条路）。
func TestRevokedTokenStopsWorking(t *testing.T) {
	s := newServer(t, authz{caps: map[connsvc.Capability]bool{connsvc.CapProfilesList: true}}, nil)
	caller := connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p1", InstanceID: "i1"}
	tok := s.tokens.Issue(caller)

	ctx, cancel := authed(tok)
	defer cancel()
	if _, err := s.profiles.List(ctx, connect.NewRequest(&hostv1.ListRequest{})); err != nil {
		t.Fatalf("有效令牌该能用: %v", err)
	}
	if n := s.tokens.RevokePlugin("p1"); n != 1 {
		t.Fatalf("该撤销 1 个令牌，得 %d", n)
	}
	if _, err := s.profiles.List(ctx, connect.NewRequest(&hostv1.ListRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("撤销后该失效，得 %v", err)
	}
}

// ---- 授权 ----

// TestCapabilitiesArePublicToAuthenticatedCallers 能力探测无需额外权限。
func TestCapabilitiesArePublicToAuthenticatedCallers(t *testing.T) {
	s := newServer(t, authz{}, nil)
	tok := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p", InstanceID: "i"})
	ctx, cancel := authed(tok)
	defer cancel()

	resp, err := s.hosts.GetCapabilities(ctx, connect.NewRequest(&hostv1.GetCapabilitiesRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	c := resp.Msg.GetCapabilities()
	if c.GetHostApiVersion() != connsvc.HostAPIVersion {
		t.Fatalf("HostAPI 版本 = %q", c.GetHostApiVersion())
	}
	// 能力清单要能指导插件：数据库查询/执行与 SSH 执行都在。
	feats := map[hostv1.Feature]bool{}
	for _, f := range c.GetFeatures() {
		feats[f] = true
	}
	for _, want := range []hostv1.Feature{
		hostv1.Feature_FEATURE_DB_QUERY, hostv1.Feature_FEATURE_DB_EXEC, hostv1.Feature_FEATURE_SSH_EXEC,
	} {
		if !feats[want] {
			t.Errorf("能力清单缺 %v", want)
		}
	}
	if len(c.GetServices()) == 0 {
		t.Error("服务清单不该为空")
	}
}

// TestMissingPermissionIsDeniedWithStableCode 没授权 → PermissionDenied + 稳定码。
func TestMissingPermissionIsDeniedWithStableCode(t *testing.T) {
	s := newServer(t, authz{}, nil) // 插件没有任何能力
	tok := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p", InstanceID: "i"})
	ctx, cancel := authed(tok)
	defer cancel()

	_, err := s.conns.OpenDB(ctx, connect.NewRequest(&hostv1.OpenDBRequest{Env: "prod", ReadOnly: true}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("该是 PermissionDenied，得 %v", err)
	}
	if info := errInfo(t, err); info.GetCode() != connsvc.CodePermissionDenied {
		t.Fatalf("稳定码该是 %s，得 %q", connsvc.CodePermissionDenied, info.GetCode())
	}
}

// TestUnauthorizedEnvIsNotLeaked 未授权环境与不存在的环境对未授权方不可区分。
func TestUnauthorizedEnvIsNotLeaked(t *testing.T) {
	s := newServer(t, authz{caps: map[connsvc.Capability]bool{connsvc.CapProfilesList: true}}, nil)
	tok := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p", InstanceID: "i"})
	ctx, cancel := authed(tok)
	defer cancel()

	resp, err := s.profiles.List(ctx, connect.NewRequest(&hostv1.ListRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if n := len(resp.Msg.GetProfiles()); n != 0 {
		t.Fatalf("空授权该得空清单，得 %d 条", n)
	}
}

// TestProfilesOnlyContainAuthorizedEnvs 只返回获授权环境。
func TestProfilesOnlyContainAuthorizedEnvs(t *testing.T) {
	s := newServer(t, authz{
		caps: map[connsvc.Capability]bool{connsvc.CapProfilesList: true},
		envs: map[connsvc.Capability][]string{connsvc.CapProfilesList: {"test"}},
	}, nil)
	tok := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p", InstanceID: "i"})
	ctx, cancel := authed(tok)
	defer cancel()

	resp, err := s.profiles.List(ctx, connect.NewRequest(&hostv1.ListRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Msg.GetProfiles()) != 1 || resp.Msg.GetProfiles()[0].GetEnv() != "test" {
		t.Fatalf("该只返回 test，得 %v", resp.Msg.GetProfiles())
	}
	// 脱敏：proto 里就没有口令字段，这里确认账号只给名字。
	p := resp.Msg.GetProfiles()[0]
	if got := p.GetDbAccounts(); len(got) != 1 || got[0] != "ds" {
		t.Fatalf("账号清单只该有名字，得 %v", got)
	}
}

// ---- 会话与数据面 ----

// TestSessionLifecycleOverRPC 打开 → 查询 → 关闭。
func TestSessionLifecycleOverRPC(t *testing.T) {
	fs := &fakeSession{typ: "oracle", version: "v", cols: []string{"n"}, rows: [][]string{{"1"}, {"2"}}}
	s := newServer(t, authz{caps: map[connsvc.Capability]bool{connsvc.CapDBConnect: true}, envs: map[connsvc.Capability][]string{connsvc.CapDBConnect: {"prod"}}},
		func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			return fs, nil
		})
	tok := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p", InstanceID: "i"})
	ctx, cancel := authed(tok)
	defer cancel()

	open, err := s.conns.OpenDB(ctx, connect.NewRequest(&hostv1.OpenDBRequest{Env: "prod", ReadOnly: true}))
	if err != nil {
		t.Fatal(err)
	}
	handle := open.Msg.GetHandle()
	if handle == "" || !open.Msg.GetReadOnly() {
		t.Fatalf("句柄/只读标志不对: %v", open.Msg)
	}

	q, err := s.db.Query(ctx, connect.NewRequest(&hostv1.QueryRequest{Handle: handle, Sql: "SELECT n FROM t", MaxRows: 10}))
	if err != nil {
		t.Fatal(err)
	}
	if q.Msg.GetRowCount() != 2 || len(q.Msg.GetRows()) != 2 || q.Msg.GetRows()[1].GetValues()[0] != "2" {
		t.Fatalf("查询结果不对: %v", q.Msg)
	}

	if _, err := s.conns.Close(ctx, connect.NewRequest(&hostv1.CloseRequest{Handle: handle})); err != nil {
		t.Fatal(err)
	}
	if !fs.closed {
		t.Error("关闭后底层会话该关掉")
	}
	_, err = s.db.Query(ctx, connect.NewRequest(&hostv1.QueryRequest{Handle: handle, Sql: "SELECT 1"}))
	if connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("关闭后的句柄该 NotFound，得 %v", err)
	}
}

// TestCrossPluginHandleIsRejected 别的插件拿同一个句柄一律 NotFound。
func TestCrossPluginHandleIsRejected(t *testing.T) {
	s := newServer(t, authz{
		caps: map[connsvc.Capability]bool{connsvc.CapDBConnect: true},
		envs: map[connsvc.Capability][]string{connsvc.CapDBConnect: {"prod"}},
	}, nil)
	t1 := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: "p1", InstanceID: "i1"})
	t2 := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: "p2", InstanceID: "i2"})

	ctx1, c1 := authed(t1)
	defer c1()
	open, err := s.conns.OpenDB(ctx1, connect.NewRequest(&hostv1.OpenDBRequest{Env: "prod", ReadOnly: true}))
	if err != nil {
		t.Fatal(err)
	}

	ctx2, c2 := authed(t2)
	defer c2()
	if _, err := s.db.Query(ctx2, connect.NewRequest(&hostv1.QueryRequest{Handle: open.Msg.GetHandle(), Sql: "SELECT 1"})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("跨插件该 NotFound，得 %v", err)
	}
}

// TestReadOnlySessionRejectsWrite 只读会话上的写被拒，稳定码可分辨。
func TestReadOnlySessionRejectsWrite(t *testing.T) {
	s := newServer(t, authz{caps: map[connsvc.Capability]bool{connsvc.CapDBConnect: true}, envs: map[connsvc.Capability][]string{connsvc.CapDBConnect: {"prod"}}}, nil)
	tok := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p", InstanceID: "i"})
	ctx, cancel := authed(tok)
	defer cancel()

	open, err := s.conns.OpenDB(ctx, connect.NewRequest(&hostv1.OpenDBRequest{Env: "prod", ReadOnly: true}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec(ctx, connect.NewRequest(&hostv1.ExecRequest{Handle: open.Msg.GetHandle(), Sql: "DELETE FROM t"}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("只读会话写该 PermissionDenied，得 %v", err)
	}
	if info := errInfo(t, err); info.GetCode() != connsvc.CodeSessionReadOnly {
		t.Fatalf("稳定码该是 %s，得 %q", connsvc.CodeSessionReadOnly, info.GetCode())
	}
}

// ---- 一致性 ----

// TestLocalAndRPCProduceTheSameResult 同一份实现、两条入口，结果必须一致。
//
// 这是"同服务、双接入"最直接的那条守卫：第一方直接调用与插件经 RPC 调用，
// 拿到的 ok/阶段/版本/地址必须一模一样 —— 否则就是有两套连接逻辑。
func TestLocalAndRPCProduceTheSameResult(t *testing.T) {
	fs := &fakeSession{typ: "oracle", version: "Oracle Database 19c"}
	open := func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
		return fs, nil
	}
	s := newServer(t, authz{caps: map[connsvc.Capability]bool{connsvc.CapDBConnect: true}, envs: map[connsvc.Capability][]string{connsvc.CapDBConnect: {"prod"}}}, open)

	localRes, localErr := s.svc.TestDB(context.Background(), connsvc.BuiltinCaller(), "prod", "")
	if localErr != nil {
		t.Fatal(localErr)
	}

	tok := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p", InstanceID: "i"})
	ctx, cancel := authed(tok)
	defer cancel()
	rpcRes, err := s.conns.TestDB(ctx, connect.NewRequest(&hostv1.TestDBRequest{Env: "prod"}))
	if err != nil {
		t.Fatal(err)
	}

	if rpcRes.Msg.GetOk() != localRes.OK {
		t.Errorf("ok 不一致: rpc=%v local=%v", rpcRes.Msg.GetOk(), localRes.OK)
	}
	if rpcRes.Msg.GetServerVersion() != localRes.ServerVersion {
		t.Errorf("版本不一致: rpc=%q local=%q", rpcRes.Msg.GetServerVersion(), localRes.ServerVersion)
	}
	if rpcRes.Msg.GetAddress() != localRes.Address {
		t.Errorf("地址不一致: rpc=%q local=%q", rpcRes.Msg.GetAddress(), localRes.Address)
	}
	if rpcRes.Msg.GetStage() != localRes.Stage {
		t.Errorf("阶段不一致: rpc=%q local=%q", rpcRes.Msg.GetStage(), localRes.Stage)
	}
}

// ---- 崩溃隔离 ----

// TestHandlerPanicBecomesInternalError handler 里的 panic 不能带走进程。
func TestHandlerPanicBecomesInternalError(t *testing.T) {
	s := newServer(t, authz{caps: map[connsvc.Capability]bool{connsvc.CapDBConnect: true}, envs: map[connsvc.Capability][]string{connsvc.CapDBConnect: {"prod"}}},
		func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			panic("插件后端模拟崩溃")
		})
	tok := s.tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "p", InstanceID: "i"})
	ctx, cancel := authed(tok)
	defer cancel()

	_, err := s.conns.OpenDB(ctx, connect.NewRequest(&hostv1.OpenDBRequest{Env: "prod"}))
	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("panic 该变成 Internal，得 %v", err)
	}
	// 服务还活着：能力探测仍能回答。
	if _, err := s.hosts.GetCapabilities(ctx, connect.NewRequest(&hostv1.GetCapabilitiesRequest{})); err != nil {
		t.Fatalf("panic 之后服务该继续可用: %v", err)
	}
}

// ---- Host / Origin 守卫 ----

// TestOriginAndHostGuards 非本机 Origin / Host 一律 403（DNS rebinding 与跨站直连）。
func TestOriginAndHostGuards(t *testing.T) {
	s := newServer(t, authz{}, nil)
	path := Prefix + hostv1connect.HostServiceGetCapabilitiesProcedure

	cases := []struct {
		name   string
		origin string
		host   string
		want   int
	}{
		{"本机 Origin 放行（令牌仍必需）", "http://127.0.0.1:28670", "", http.StatusUnauthorized},
		{"沙箱 iframe 的 null Origin 放行（令牌仍必需）", "null", "", http.StatusUnauthorized},
		{"外站 Origin 拒绝", "http://evil.example", "", http.StatusForbidden},
		{"非本机 Host 拒绝", "", "evil.example", http.StatusForbidden},
		{"局域网 IP Host 拒绝", "", "192.168.1.10:28670", http.StatusForbidden},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			// httptest.NewRequest 的缺省 Host 是 example.com；本机请求要显式给。
			req.Host = "127.0.0.1:28670"
			if c.origin != "" {
				req.Header.Set("Origin", c.origin)
			}
			if c.host != "" {
				req.Host = c.host
			}
			rec := httptest.NewRecorder()
			s.http.Config.Handler.ServeHTTP(rec, req)
			if rec.Code != c.want {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, c.want, rec.Body.String())
			}
		})
	}
}

// ---- 令牌表 ----

// TestTokenStoreInstanceRevocation 按实例撤销只影响该实例。
func TestTokenStoreInstanceRevocation(t *testing.T) {
	store := NewTokenStore()
	a := store.Issue(connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: "p", InstanceID: "i1"})
	b := store.Issue(connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: "p", InstanceID: "i2"})
	if store.Count() != 2 {
		t.Fatalf("令牌数 = %d", store.Count())
	}
	if n := store.RevokeInstance("p", "i1"); n != 1 {
		t.Fatalf("该撤销 1 个，得 %d", n)
	}
	if _, ok := store.Lookup(a); ok {
		t.Error("i1 的令牌该失效")
	}
	if _, ok := store.Lookup(b); !ok {
		t.Error("i2 的令牌不该被牵连")
	}
}

// TestTokenIsRandom 令牌必须够长且互不相同（不引入 JWT，靠足够熵）。
func TestTokenIsRandom(t *testing.T) {
	store := NewTokenStore()
	seen := map[string]bool{}
	for i := 0; i < 64; i++ {
		tok := store.Issue(connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: "p", InstanceID: fmt.Sprint(i)})
		if len(tok) < 32 {
			t.Fatalf("令牌太短: %q", tok)
		}
		if seen[tok] {
			t.Fatalf("令牌重复: %q", tok)
		}
		seen[tok] = true
	}
}
