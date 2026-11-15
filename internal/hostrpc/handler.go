package hostrpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/gen/hostv1"
	"github.com/luyaogod/TT/internal/gen/hostv1/hostv1connect"
)

// Deps 挂载 Host API 所需的依赖。
type Deps struct {
	// Service 进程内连接服务（唯一实现；本包不复制任何连接逻辑）。
	Service *connsvc.Service
	// Auth 认证器；nil = 一律拒绝（宁可不给入口，也不开一个匿名可用的连接 API）。
	Auth Authenticator
}

// Prefix Host API 的固定挂载前缀。handler 自己认这个前缀，所以挂载方只需
// `m.Handle(hostrpc.Prefix+"/", hostrpc.NewHandler(...))`，**不要**再 StripPrefix。
const Prefix = "/hostrpc"

// NewHandler 构造 Host API 的 http.Handler。
func NewHandler(deps Deps) http.Handler {
	opts := []connect.HandlerOption{
		connect.WithInterceptors(
			authInterceptor(deps.Auth),
			recoverInterceptor(),
		),
	}
	// Connect 生成的 handler 认的是 `<包>.<服务>/<方法>` 绝对路径；前缀在下面统一加一次。
	inner := http.NewServeMux()
	inner.Handle(hostv1connect.NewHostServiceHandler(&hostServer{svc: deps.Service}, opts...))
	inner.Handle(hostv1connect.NewProfilesServiceHandler(&profilesServer{svc: deps.Service}, opts...))
	inner.Handle(hostv1connect.NewConnectionsServiceHandler(&connectionsServer{svc: deps.Service}, opts...))
	inner.Handle(hostv1connect.NewDBSessionServiceHandler(&dbSessionServer{svc: deps.Service}, opts...))
	inner.Handle(hostv1connect.NewSSHSessionServiceHandler(&sshSessionServer{svc: deps.Service}, opts...))

	mux := http.NewServeMux()
	mux.Handle(Prefix+"/", http.StripPrefix(Prefix, inner))
	return guard(mux)
}

// authInterceptor 把 Bearer 令牌解析成调用方身份并放进 ctx。
func authInterceptor(auth Authenticator) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if auth == nil {
				return nil, unauthenticated("Host API 未启用认证")
			}
			caller, err := auth.Authenticate(ctx, req)
			if err != nil {
				return nil, err
			}
			return next(WithCaller(ctx, caller), req)
		}
	}
}

// recoverInterceptor 把 handler panic 变成 Internal 错误。
//
// 理由很实际：Host API 与 Core 同进程，一次 panic 不能带走整个 tt serve。
func recoverInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			defer func() {
				if p := recover(); p != nil {
					resp, err = nil, connect.NewError(connect.CodeInternal,
						errText(fmt.Sprintf("Host API 内部错误: %v", p)))
				}
			}()
			return next(ctx, req)
		}
	}
}

// ---- 服务实现 ----

// hostServer HostService。
type hostServer struct {
	svc *connsvc.Service
}

// GetCapabilities 返回能力清单。
func (h *hostServer) GetCapabilities(_ context.Context, _ *connect.Request[hostv1.GetCapabilitiesRequest]) (*connect.Response[hostv1.GetCapabilitiesResponse], error) {
	c := h.svc.Capabilities()
	out := &hostv1.Capabilities{
		HostApiVersion: c.HostAPIVersion,
		TtVersion:      c.TTVersion,
	}
	for _, f := range c.Features {
		out.Features = append(out.Features, featureProto(f))
	}
	for _, s := range c.Services {
		out.Services = append(out.Services, &hostv1.ServiceCapability{Name: s.Name, Methods: s.Methods})
	}
	return connect.NewResponse(&hostv1.GetCapabilitiesResponse{Capabilities: out}), nil
}

// profilesServer ProfilesService。
type profilesServer struct {
	svc *connsvc.Service
}

// List 只返回调用方获授权的环境。
func (p *profilesServer) List(ctx context.Context, _ *connect.Request[hostv1.ListRequest]) (*connect.Response[hostv1.ListResponse], error) {
	caller, err := mustCaller(ctx)
	if err != nil {
		return nil, err
	}
	profs, cerr := p.svc.ProfilesFor(caller)
	if cerr != nil {
		return nil, rpcError(cerr)
	}
	out := &hostv1.ListResponse{}
	for _, pr := range profs {
		out.Profiles = append(out.Profiles, profileProto(pr))
	}
	return connect.NewResponse(out), nil
}

// connectionsServer ConnectionsService。
type connectionsServer struct {
	svc *connsvc.Service
}

// TestDB 测数据库连通性（连接层失败是 ok=false 的正常返回）。
func (c *connectionsServer) TestDB(ctx context.Context, req *connect.Request[hostv1.TestDBRequest]) (*connect.Response[hostv1.TestDBResponse], error) {
	caller, err := mustCaller(ctx)
	if err != nil {
		return nil, err
	}
	res, cerr := c.svc.TestDB(ctx, caller, req.Msg.GetEnv(), req.Msg.GetAccount())
	if cerr != nil {
		return nil, rpcError(cerr)
	}
	return connect.NewResponse(&hostv1.TestDBResponse{
		Ok:            res.OK,
		ServerVersion: res.ServerVersion,
		Address:       res.Address,
		Stage:         res.Stage,
		Message:       res.Message,
		DurationMs:    res.Duration.Milliseconds(),
	}), nil
}

// TestSSH 测 SSH 链路。
func (c *connectionsServer) TestSSH(ctx context.Context, req *connect.Request[hostv1.TestSSHRequest]) (*connect.Response[hostv1.TestSSHResponse], error) {
	caller, err := mustCaller(ctx)
	if err != nil {
		return nil, err
	}
	res, cerr := c.svc.TestSSH(ctx, caller, req.Msg.GetEnv())
	if cerr != nil {
		return nil, rpcError(cerr)
	}
	return connect.NewResponse(&hostv1.TestSSHResponse{
		Ok:         res.OK,
		Banner:     res.Banner,
		Stage:      res.Stage,
		Message:    res.Message,
		DurationMs: res.Duration.Milliseconds(),
	}), nil
}

// OpenDB 打开数据库会话。
func (c *connectionsServer) OpenDB(ctx context.Context, req *connect.Request[hostv1.OpenDBRequest]) (*connect.Response[hostv1.OpenDBResponse], error) {
	caller, err := mustCaller(ctx)
	if err != nil {
		return nil, err
	}
	handle, res, cerr := c.svc.OpenDB(ctx, caller, req.Msg.GetEnv(), req.Msg.GetAccount(), req.Msg.GetReadOnly())
	if cerr != nil {
		return nil, rpcError(cerr)
	}
	return connect.NewResponse(&hostv1.OpenDBResponse{Handle: handle, ReadOnly: res.ReadOnly, DbType: res.DBType}), nil
}

// OpenSSH 打开 SSH 会话。
func (c *connectionsServer) OpenSSH(ctx context.Context, req *connect.Request[hostv1.OpenSSHRequest]) (*connect.Response[hostv1.OpenSSHResponse], error) {
	caller, err := mustCaller(ctx)
	if err != nil {
		return nil, err
	}
	handle, cerr := c.svc.OpenSSH(ctx, caller, req.Msg.GetEnv())
	if cerr != nil {
		return nil, rpcError(cerr)
	}
	return connect.NewResponse(&hostv1.OpenSSHResponse{Handle: handle}), nil
}

// Close 关闭会话。
func (c *connectionsServer) Close(ctx context.Context, req *connect.Request[hostv1.CloseRequest]) (*connect.Response[hostv1.CloseResponse], error) {
	caller, err := mustCaller(ctx)
	if err != nil {
		return nil, err
	}
	if cerr := c.svc.Close(caller, req.Msg.GetHandle()); cerr != nil {
		return nil, rpcError(cerr)
	}
	return connect.NewResponse(&hostv1.CloseResponse{}), nil
}

// dbSessionServer DBSessionService。
type dbSessionServer struct {
	svc *connsvc.Service
}

// Query 执行查询。
func (d *dbSessionServer) Query(ctx context.Context, req *connect.Request[hostv1.QueryRequest]) (*connect.Response[hostv1.QueryResponse], error) {
	caller, err := mustCaller(ctx)
	if err != nil {
		return nil, err
	}
	res, cerr := d.svc.DBQuery(ctx, caller, req.Msg.GetHandle(), req.Msg.GetSql(), int(req.Msg.GetMaxRows()))
	if cerr != nil {
		return nil, rpcError(cerr)
	}
	out := &hostv1.QueryResponse{Columns: res.Columns, Truncated: res.Truncated, RowCount: int64(len(res.Rows))}
	for _, r := range res.Rows {
		out.Rows = append(out.Rows, &hostv1.Row{Values: r})
	}
	return connect.NewResponse(out), nil
}

// Exec 执行写语句。
func (d *dbSessionServer) Exec(ctx context.Context, req *connect.Request[hostv1.ExecRequest]) (*connect.Response[hostv1.ExecResponse], error) {
	caller, err := mustCaller(ctx)
	if err != nil {
		return nil, err
	}
	n, cerr := d.svc.DBExec(ctx, caller, req.Msg.GetHandle(), req.Msg.GetSql())
	if cerr != nil {
		return nil, rpcError(cerr)
	}
	return connect.NewResponse(&hostv1.ExecResponse{RowsAffected: n}), nil
}

// ---- 批量与事务 ----
//
// 契约已在 proto 冻结（P0），实现随 P2「资源服务」落地。这组方法故意不注册
// FEATURE_DB_BATCH，能力清单里没有它时插件本来就不该调用；真调了得到 Unimplemented
// 而不是静默的半成品。

func (d *dbSessionServer) Batch(_ context.Context, _ *connect.Request[hostv1.BatchRequest]) (*connect.Response[hostv1.BatchResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("批量执行尚未接入"))
}

func (d *dbSessionServer) BeginTx(_ context.Context, _ *connect.Request[hostv1.BeginTxRequest]) (*connect.Response[hostv1.BeginTxResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("显式事务尚未接入"))
}

func (d *dbSessionServer) TxExec(_ context.Context, _ *connect.Request[hostv1.TxExecRequest]) (*connect.Response[hostv1.TxExecResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("显式事务尚未接入"))
}

func (d *dbSessionServer) CommitTx(_ context.Context, _ *connect.Request[hostv1.CommitTxRequest]) (*connect.Response[hostv1.CommitTxResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("显式事务尚未接入"))
}

func (d *dbSessionServer) RollbackTx(_ context.Context, _ *connect.Request[hostv1.RollbackTxRequest]) (*connect.Response[hostv1.RollbackTxResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("显式事务尚未接入"))
}

// sshSessionServer SSHSessionService。
type sshSessionServer struct {
	svc *connsvc.Service
}

// Exec 执行一次性命令（非零退出码是正常结果，不报错）。
func (s *sshSessionServer) Exec(ctx context.Context, req *connect.Request[hostv1.SSHExecRequest]) (*connect.Response[hostv1.SSHExecResponse], error) {
	caller, err := mustCaller(ctx)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(req.Msg.GetTimeoutMs()) * time.Millisecond
	res, cerr := s.svc.SSHExec(ctx, caller, req.Msg.GetHandle(), req.Msg.GetCommand(), timeout)
	if cerr != nil {
		return nil, rpcError(cerr)
	}
	return connect.NewResponse(&hostv1.SSHExecResponse{
		Output:   res.Output,
		ExitCode: int32(res.ExitCode),
		TimedOut: res.TimedOut,
	}), nil
}

// ---- 映射 ----

// featureProto 内部 feature 串 → proto 枚举。
func featureProto(s string) hostv1.Feature {
	switch s {
	case connsvc.FeatureDBQuery:
		return hostv1.Feature_FEATURE_DB_QUERY
	case connsvc.FeatureDBExec:
		return hostv1.Feature_FEATURE_DB_EXEC
	case connsvc.FeatureSSHExec:
		return hostv1.Feature_FEATURE_SSH_EXEC
	case connsvc.FeatureDBReadOnlySession:
		return hostv1.Feature_FEATURE_DB_READONLY_SESSION
	default:
		return hostv1.Feature_FEATURE_UNSPECIFIED
	}
}

// profileProto 内部脱敏元数据 → proto。
func profileProto(p connsvc.Profile) *hostv1.Profile {
	return &hostv1.Profile{
		Env:           p.Env,
		Zone:          p.Zone,
		HasDb:         p.HasDB,
		DbType:        p.DBType,
		DbAddress:     p.DBAddress,
		DbAccounts:    p.DBAccounts,
		HasSsh:        p.HasSSH,
		SshHost:       p.SSHHost,
		SshPort:       int32(p.SSHPort),
		SshUser:       p.SSHUser,
		DbViaSsh:      p.DBViaSSH,
		DbReadonlySql: p.DBReadonlySQL,
	}
}

// ---- Host / Origin 守卫 ----

// guard 校验 Host 与 Origin 头。
//
// 为什么 loopback 不够：TT 的 Web 壳与插件页面都跑在本机浏览器里，一个恶意本机网页
// 也能向 127.0.0.1 发请求。真正的身份来自实例令牌（见 TokenAuthenticator）；
// 这里挡的是 DNS rebinding（Host 头不是本机名）与跨站直连（Origin 不是本机）。
//
// 沙箱 iframe（sandbox="allow-scripts" 且不带 allow-same-origin）发出的 Origin 是
// 字面量 "null"。放它过 Host/Origin 这一关是安全的：**令牌仍然必需**，
// 而令牌只在插件进程与该插件的桥会话里存在。
func guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hostIsLoopback(r.Host) {
			writeGuardError(w, badOrigin("Host 头不是本机地址（拒绝 DNS rebinding 风格请求）"))
			return
		}
		if o := strings.TrimSpace(r.Header.Get("Origin")); o != "" && o != "null" {
			u, err := url.Parse(o)
			if err != nil || !hostIsLoopback(u.Host) {
				writeGuardError(w, badOrigin("Origin 不是本机地址"))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// hostIsLoopback host[:port] 是否本机名或回环地址。
func hostIsLoopback(hostport string) bool {
	if hostport == "" {
		return false
	}
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	} else {
		// 没有端口（或 IPv6 字面量）：去掉方括号再判。
		host = strings.Trim(host, "[]")
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// writeGuardError 用 Connect 的错误 JSON 形态回一个守卫失败。
func writeGuardError(w http.ResponseWriter, e *connect.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"code":"` + e.Code().String() + `","message":"` + jsonEscape(e.Message()) + `"}`))
}

// jsonEscape 最小 JSON 字符串转义（守卫错误只用 ASCII 文案，这里够用）。
func jsonEscape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	return r.Replace(s)
}
