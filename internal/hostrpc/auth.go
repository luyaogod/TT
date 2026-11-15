// Package hostrpc 是 Host API 的**进程外**入口：ConnectRPC + Protobuf，挂载在 /hostrpc/。
//
// 与 internal/connsvc 的分工：connsvc 是唯一实现（第一方在进程内直接调它），
// 本包只做三件事 —— 认证（实例令牌 → 调用方身份）、协议映射（DTO ↔ 内部类型）、
// 错误映射（稳定码 ↔ Connect code + ErrorInfo）。**不复制任何连接逻辑**。
package hostrpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/pkg/hostv1"
)

// codeUnauthenticated / codeForbidden 是 Host API 自己的稳定码（不是 connsvc 的）。
const (
	codeUnauthenticated = "UNAUTHENTICATED"
	codeBadOrigin       = "BAD_ORIGIN"
	codeBadHost         = "BAD_HOST"
)

// TokenStore 实例令牌表：令牌 → 调用方身份。
//
// 令牌用 crypto/rand 生成（不引入 JWT：没有跨进程验签需求，也就不需要那套复杂度）。
// 令牌只在内存里：TT 重启后旧令牌全部失效，插件必须重新握手 —— 这正是想要的语义。
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]tokenEntry
	now    func() time.Time
}

type tokenEntry struct {
	caller   connsvc.Caller
	issuedAt time.Time
}

// NewTokenStore 构造空令牌表。
func NewTokenStore() *TokenStore {
	return &TokenStore{tokens: map[string]tokenEntry{}, now: time.Now}
}

// Issue 为一个调用方签发新令牌。
//
// 同一个插件实例多次 Issue 会得到多个有效令牌（CLI 与 service 各一个），
// RevokePlugin / RevokeInstance 会把它们一起清掉 —— 授权撤销不能漏掉任何一个。
func (s *TokenStore) Issue(caller connsvc.Caller) string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand 失败时**不能**退化成弱随机：那等于静默降低安全边界。
		panic("hostrpc: crypto/rand 不可用: " + err.Error())
	}
	tok := hex.EncodeToString(b)
	s.mu.Lock()
	s.tokens[tok] = tokenEntry{caller: caller, issuedAt: s.now()}
	s.mu.Unlock()
	return tok
}

// Revoke 撤销一个令牌（已知令牌串时用）。
func (s *TokenStore) Revoke(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tokens[token]; !ok {
		return false
	}
	delete(s.tokens, token)
	return true
}

// RevokePlugin 撤销某插件的全部令牌，返回撤销个数。
func (s *TokenStore) RevokePlugin(pluginID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for tok, e := range s.tokens {
		if e.caller.PluginID == pluginID && !e.caller.IsBuiltin() {
			delete(s.tokens, tok)
			n++
		}
	}
	return n
}

// RevokeInstance 撤销某插件实例的全部令牌。
func (s *TokenStore) RevokeInstance(pluginID, instanceID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for tok, e := range s.tokens {
		if e.caller.PluginID == pluginID && e.caller.InstanceID == instanceID && !e.caller.IsBuiltin() {
			delete(s.tokens, tok)
			n++
		}
	}
	return n
}

// Lookup 按令牌取调用方身份。
func (s *TokenStore) Lookup(token string) (connsvc.Caller, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.tokens[token]
	return e.caller, ok
}

// Count 当前有效令牌数（诊断用）。
func (s *TokenStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tokens)
}

// Authenticator 把一次 RPC 请求变成调用方身份。
type Authenticator interface {
	Authenticate(ctx context.Context, req connect.AnyRequest) (connsvc.Caller, error)
}

// TokenAuthenticator 用 Bearer 实例令牌认证。
type TokenAuthenticator struct {
	Store *TokenStore
}

// NewTokenAuthenticator 用给定令牌表构造认证器。
func NewTokenAuthenticator(store *TokenStore) *TokenAuthenticator {
	return &TokenAuthenticator{Store: store}
}

// Authenticate 读 Authorization: Bearer <token> 并解析身份。
func (a *TokenAuthenticator) Authenticate(_ context.Context, req connect.AnyRequest) (connsvc.Caller, error) {
	if a == nil || a.Store == nil {
		return connsvc.Caller{}, unauthenticated("Host API 未启用（没有令牌表）")
	}
	h := req.Header().Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return connsvc.Caller{}, unauthenticated("缺少实例令牌（Authorization: Bearer <token>）")
	}
	tok := strings.TrimSpace(strings.TrimPrefix(h, prefix))
	caller, ok := a.Store.Lookup(tok)
	if !ok {
		return connsvc.Caller{}, unauthenticated("实例令牌无效或已失效（插件被停用/重启/撤销授权后旧令牌即失效）")
	}
	return caller, nil
}

// callerKey context 中存放调用方身份的键。
type callerKey struct{}

// WithCaller 把调用方身份放进 ctx。
func WithCaller(ctx context.Context, c connsvc.Caller) context.Context {
	return context.WithValue(ctx, callerKey{}, c)
}

// CallerFrom 取实现侧的调用方身份（只可能是本包认证器放进去的，不是插件自报的）。
func CallerFrom(ctx context.Context) (connsvc.Caller, bool) {
	c, ok := ctx.Value(callerKey{}).(connsvc.Caller)
	return c, ok
}

// mustCaller 取调用方；缺失是**装配错误**（漏了认证拦截器），按内部错误返回。
func mustCaller(ctx context.Context) (connsvc.Caller, error) {
	c, ok := CallerFrom(ctx)
	if !ok {
		return connsvc.Caller{}, connect.NewError(connect.CodeInternal, errText("Host API 缺少调用方身份（认证拦截器未生效）"))
	}
	return c, nil
}

// errText 把字符串变成 error（避免 connect 包里再引入一个错误类型）。
type errText string

func (e errText) Error() string { return string(e) }

// hostv1ErrorInfo 稳定的错误详情。
func hostv1ErrorInfo(e *connsvc.Error) *hostv1.ErrorInfo {
	return &hostv1.ErrorInfo{
		Code:    e.Code,
		Message: e.Message,
		Hint:    e.Hint,
		Env:     e.Env,
	}
}

// rpcError 把 connsvc 稳定错误映射成 Connect 错误（带 ErrorInfo 详情）。
func rpcError(err error) *connect.Error {
	e, ok := connsvc.AsError(err)
	if !ok {
		return connect.NewError(connect.CodeInternal, err)
	}
	ce := connect.NewError(codeFor(e.Code), errText(e.Message))
	if d, derr := connect.NewErrorDetail(hostv1ErrorInfo(e)); derr == nil {
		ce.AddDetail(d)
	}
	return ce
}

// codeFor 稳定码 → Connect code 的唯一映射表。
//
// 插件应当只依赖 ErrorInfo.code；这里给的是"传输层大致属于哪类"，方便通用客户端重试策略。
func codeFor(code string) connect.Code {
	switch code {
	case connsvc.CodeEnvUnknown, connsvc.CodeSessionNotFound:
		return connect.CodeNotFound
	case connsvc.CodeEnvNotAuthorized, connsvc.CodePermissionDenied, connsvc.CodeSessionReadOnly:
		return connect.CodePermissionDenied
	case connsvc.CodeSessionKind, connsvc.CodeAccountUnknown:
		return connect.CodeInvalidArgument
	case connsvc.CodeDBNoConfig, connsvc.CodeConfigChanged, connsvc.CodeClosed:
		return connect.CodeFailedPrecondition
	case connsvc.CodeConfigBroken:
		return connect.CodeInternal
	case connsvc.CodeDBConnectFailed, connsvc.CodeSSHConnectFailed:
		return connect.CodeUnavailable
	case connsvc.CodeTimeout:
		return connect.CodeDeadlineExceeded
	case connsvc.CodeQueryFailed, connsvc.CodeExecFailed, connsvc.CodeSSHExecFailed:
		return connect.CodeUnknown
	default:
		return connect.CodeInternal
	}
}

// unauthenticated 造一个"没身份"的 Connect 错误（也带 ErrorInfo）。
func unauthenticated(msg string) *connect.Error {
	e := connect.NewError(connect.CodeUnauthenticated, errText(msg))
	if d, derr := connect.NewErrorDetail(&hostv1.ErrorInfo{Code: codeUnauthenticated, Message: msg}); derr == nil {
		e.AddDetail(d)
	}
	return e
}

// badOrigin 造一个来源/主机校验失败的 Connect 错误。
func badOrigin(msg string) *connect.Error {
	e := connect.NewError(connect.CodePermissionDenied, errText(msg))
	if d, derr := connect.NewErrorDetail(&hostv1.ErrorInfo{Code: codeBadOrigin, Message: msg}); derr == nil {
		e.AddDetail(d)
	}
	return e
}

// forbiddenHost 造一个 Host 校验失败的 Connect 错误。
func forbiddenHost(msg string) *connect.Error {
	e := connect.NewError(connect.CodePermissionDenied, errText(msg))
	if d, derr := connect.NewErrorDetail(&hostv1.ErrorInfo{Code: codeBadHost, Message: msg}); derr == nil {
		e.AddDetail(d)
	}
	return e
}
