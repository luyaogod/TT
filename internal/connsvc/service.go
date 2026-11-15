package connsvc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/pkg/erpdb"
	"github.com/luyaogod/TT/internal/host"
	"github.com/luyaogod/TT/internal/sshtun"
)

// HostAPIVersion 首版 Host API 的语义化版本。插件清单的 hostApi 范围按它判。
const HostAPIVersion = "1.0.0"

// Capability 首版能力名（与插件清单 permissions 取值一致）。
type Capability string

const (
	// CapProfilesList 读已授权环境的脱敏元数据。
	CapProfilesList Capability = "profiles.list"
	// CapDBConnect 用已授权数据库配置建立会话。
	CapDBConnect Capability = "db.connect"
	// CapSSHConnect 用已授权 SSH 配置建立会话。
	CapSSHConnect Capability = "ssh.connect"
)

// AllCapabilities 首版全部能力（登记/校验用；顺序即对外展示顺序）。
var AllCapabilities = []Capability{CapProfilesList, CapDBConnect, CapSSHConnect}

// ValidCapability 能力名是否合法。
func ValidCapability(s string) bool {
	for _, c := range AllCapabilities {
		if string(c) == s {
			return true
		}
	}
	return false
}

// Feature 能力开关（Capabilities.Features 取值）。
const (
	FeatureDBQuery           = "db.query"
	FeatureDBExec            = "db.exec"
	FeatureSSHExec           = "ssh.exec"
	FeatureDBReadOnlySession = "db.readonly_session"
)

// CallerKind 调用方种类。内置调用方与第三方插件的权限来源不同，但走同一套会话规则。
type CallerKind string

const (
	// CallerBuiltin 第一方内置模块（进程内直接调用，不经 RPC）。
	CallerBuiltin CallerKind = "builtin"
	// CallerCLIPlugin 第三方插件的一次性 CLI 进程。
	CallerCLIPlugin CallerKind = "cli"
	// CallerServicePlugin 第三方插件的常驻服务进程。
	CallerServicePlugin CallerKind = "service"
)

// Caller 一次调用的身份。插件身份由 Core 按实例令牌解析，插件不能自报。
type Caller struct {
	Kind       CallerKind
	PluginID   string
	InstanceID string
}

// BuiltinCaller 内置调用方身份（第一方模块 / 设置页）。
func BuiltinCaller() Caller {
	return Caller{Kind: CallerBuiltin, PluginID: "tt", InstanceID: "builtin"}
}

// IsBuiltin 是否内置调用方。
func (c Caller) IsBuiltin() bool { return c.Kind == CallerBuiltin }

func (c Caller) String() string {
	if c.IsBuiltin() {
		return "builtin:tt"
	}
	return string(c.Kind) + ":" + c.PluginID + "/" + c.InstanceID
}

// Authorizer 授权来源。Phase 3 由插件注册表提供实现；在此之前只有内置调用方被放行。
type Authorizer interface {
	// Permitted 调用方是否持有某能力。
	Permitted(caller Caller, capability Capability) bool
	// AllowedEnvs 调用方在某能力下被授权的环境名；all=true 表示"配置里的全部环境"。
	AllowedEnvs(caller Caller, capability Capability) (envs []string, all bool)
}

// builtinOnlyAuthorizer 缺省授权：内置调用方全通，第三方一律拒绝。
//
// 这是**故意保守**的缺省值：没有装注册表实现的进程（如 tt 单体 CLI）不可能
// 意外放行一个第三方身份 —— 忘记注入授权器时得到的是拒绝，不是全开。
type builtinOnlyAuthorizer struct{}

func (builtinOnlyAuthorizer) Permitted(c Caller, _ Capability) bool { return c.IsBuiltin() }
func (builtinOnlyAuthorizer) AllowedEnvs(c Caller, _ Capability) ([]string, bool) {
	return nil, c.IsBuiltin()
}

// Profile 一个环境的脱敏元数据。**绝不含口令/私钥/完整配置 JSON**。
type Profile struct {
	Env           string
	Zone          string
	HasDB         bool
	DBType        string
	DBAddress     string
	DBAccounts    []string
	HasSSH        bool
	SSHHost       string
	SSHPort       int
	SSHUser       string
	DBViaSSH      bool
	DBReadonlySQL bool
}

// TestDBResult 一次数据库连通性测试的结果。
type TestDBResult struct {
	OK            bool
	ServerVersion string
	// Address 实际连到的地址 host:port/svc。经隧道时是 127.0.0.1:<本地端口>/svc。
	Address string
	// DBType "oracle" | "kingbase"。
	DBType string
	// ViaTunnel 本次是否经 SSH 隧道。
	ViaTunnel bool
	// Stage 失败阶段：resolve_config / tunnel / connect / ping；成功为空。
	Stage   string
	Message string
	// Duration 整个测试耗时。
	Duration time.Duration
}

// TestSSHResult 一次 SSH 链路测试的结果。
type TestSSHResult struct {
	OK       bool
	Banner   string
	Stage    string
	Message  string
	Duration time.Duration
}

// OpenDBResult 打开数据库会话的结果。
type OpenDBResult struct {
	ReadOnly bool
	DBType   string
}

// QueryResult 一次查询的结果（已按 MaxRows 截断）。
type QueryResult struct {
	Columns   []string
	Rows      [][]string
	Truncated bool
}

// ExecResult 一次 SSH 命令执行的结果。
type ExecResult struct {
	Output   string
	ExitCode int
	TimedOut bool
}

// Event 审计事件。**不记录 SQL 文本、SSH 命令内容与秘密**。
type Event struct {
	Action     string // db.test / db.open / db.query / db.exec / ssh.test / ssh.open / ssh.exec / close / reload
	Caller     Caller
	Env        string
	Kind       Kind
	OK         bool
	Code       string
	DurationMS int64
}

// Kind 连接类型。
type Kind string

const (
	// KindDB 数据库会话。
	KindDB Kind = "db"
	// KindSSH SSH 会话。
	KindSSH Kind = "ssh"
)

// Tunnel 一条 SSH 端口转发隧道（*sshtun.Tunnel 的窄接口，便于测试注入）。
type Tunnel interface {
	LocalPort() int
	Close() error
}

// SSHConn 一条 SSH 连接（*host.SSHConn 的窄接口，便于测试注入）。
type SSHConn interface {
	Output(cmd string, timeout time.Duration) (string, error)
	OutputContext(ctx context.Context, cmd string, timeout time.Duration) (string, error)
	Close() error
}

// 缺省上限。**有上限**本身就是一条契约：没有上限的能力等于让 agent 一次打爆上下文。
const (
	// DefaultMaxRows 单次查询缺省最多返回多少行。
	DefaultMaxRows = 1000
	// MaxMaxRows 调用方可以要求的硬上限。
	MaxMaxRows = 50000
	// DefaultQueryTimeout 单次查询缺省超时。
	DefaultQueryTimeout = 60 * time.Second
	// DefaultSSHTimeout 单次 SSH 命令缺省超时。
	DefaultSSHTimeout = 60 * time.Second
	// MaxSSHTimeout SSH 命令超时硬上限。
	MaxSSHTimeout = 10 * time.Minute
)

// Options 构造 Service 的依赖。零值都能跑（真实实现作为缺省），测试再逐项注入。
type Options struct {
	// ConfigPath config.json 路径（唯一配置来源；服务不自己拼第二套路径）。
	ConfigPath string
	// TTVersion tt 自身版本（信息性，进 Capabilities）。
	TTVersion string
	// Authorizer 授权来源；nil = 只放行内置调用方。
	Authorizer Authorizer
	// Audit 审计回调；nil = 不记录。
	Audit func(Event)
	// Now 时钟（测试用）；nil = time.Now。
	Now func() time.Time

	// 以下四个是测试缝；nil 时用真实实现。
	DialTunnel func(sshtun.Options) (Tunnel, error)
	OpenDB     func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error)
	DialSSH    func(config.SSHConfig) (SSHConn, error)
	TestSSH    func(config.SSHConfig) (*host.SSHTestOut, error)
}

// Service 进程内连接服务。并发安全。
type Service struct {
	opt Options

	mu       sync.Mutex
	sessions map[string]*session
	closed   bool
	// generation 配置代次：Reload 时 +1 并关闭全部会话。
	generation uint64
}

// session 一条会话的全部归属信息。
type session struct {
	handle     string
	kind       Kind
	env        string
	caller     Caller
	generation uint64
	openedAt   time.Time
	lastUsed   time.Time

	db  erpdb.Session
	tun Tunnel
	ssh SSHConn

	closed bool
}

// New 构造服务。
func New(opt Options) *Service {
	if opt.Now == nil {
		opt.Now = time.Now
	}
	if opt.Authorizer == nil {
		opt.Authorizer = builtinOnlyAuthorizer{}
	}
	if opt.DialTunnel == nil {
		opt.DialTunnel = defaultDialTunnel
	}
	if opt.OpenDB == nil {
		opt.OpenDB = func(ctx context.Context, c dbconfig.Connection, o erpdb.SessionOptions) (erpdb.Session, error) {
			return erpdb.OpenSession(ctx, c, o)
		}
	}
	if opt.DialSSH == nil {
		opt.DialSSH = func(cfg config.SSHConfig) (SSHConn, error) {
			c, err := host.Dial(cfg)
			if err != nil {
				return nil, err
			}
			return c, nil
		}
	}
	if opt.TestSSH == nil {
		opt.TestSSH = host.TestSSH
	}
	return &Service{opt: opt, sessions: map[string]*session{}}
}

// tunnelAdapter 把 *sshtun.Tunnel 的 LocalPort 字段包成方法。
type tunnelAdapter struct{ t *sshtun.Tunnel }

func (a *tunnelAdapter) LocalPort() int { return a.t.LocalPort }
func (a *tunnelAdapter) Close() error   { return a.t.Close() }

func defaultDialTunnel(o sshtun.Options) (Tunnel, error) {
	t, err := sshtun.Dial(o)
	if err != nil {
		return nil, err
	}
	return &tunnelAdapter{t: t}, nil
}

// ServiceCapability 一个已挂载服务的能力描述。
type ServiceCapability struct {
	Name    string
	Methods []string
}

// Capabilities Host API 能力清单。
type Capabilities struct {
	HostAPIVersion string
	TTVersion      string
	Features       []string
	Services       []ServiceCapability
}

// Capabilities 返回本服务实际交付的能力。**没实现的能力不列**。
func (s *Service) Capabilities() Capabilities {
	return Capabilities{
		HostAPIVersion: HostAPIVersion,
		TTVersion:      s.opt.TTVersion,
		Features:       []string{FeatureDBQuery, FeatureDBExec, FeatureSSHExec, FeatureDBReadOnlySession},
		Services: []ServiceCapability{
			{Name: "HostService", Methods: []string{"GetCapabilities"}},
			{Name: "ProfilesService", Methods: []string{"List"}},
			{Name: "ConnectionsService", Methods: []string{"TestDB", "TestSSH", "OpenDB", "OpenSSH", "Close"}},
			{Name: "DBSessionService", Methods: []string{"Query", "Exec"}},
			{Name: "SSHSessionService", Methods: []string{"Exec"}},
		},
	}
}

// loadConfig 读配置。文件不存在时 config.Load 返回空配置（首次运行不是错误）。
func (s *Service) loadConfig() (*config.Root, *Error) {
	if s.opt.ConfigPath == "" {
		return nil, errf(CodeConfigBroken, "", "连接服务没有配置路径（Options.ConfigPath 为空）").
			WithHint("这是 TT 自身装配问题，请报告")
	}
	root, err := config.Load(s.opt.ConfigPath)
	if err != nil {
		return nil, &Error{Code: CodeConfigBroken, Message: "读取配置失败: " + err.Error(), Cause: err}
	}
	return root, nil
}

// Profiles 返回全部环境的脱敏元数据（内置调用方 / 设置页用）。
//
// **不做默认环境回退**：这就是一份清单，不是"选中了哪个环境"。
func (s *Service) Profiles() ([]Profile, *Error) {
	root, cerr := s.loadConfig()
	if cerr != nil {
		return nil, cerr
	}
	out := make([]Profile, 0, len(root.Hosts.SSHs))
	for i := range root.Hosts.SSHs {
		out = append(out, profileOf(&root.Hosts.SSHs[i]))
	}
	return out, nil
}

// ProfilesFor 返回某调用方**获授权**的环境脱敏元数据。
//
// 未授权环境不出现在结果里（而不是出现但标红）：插件靠这个清单决定要不要显示菜单项，
// 多列一个就等于告诉它"这台机器上还有别的客户环境"。
func (s *Service) ProfilesFor(caller Caller) ([]Profile, *Error) {
	if !s.opt.Authorizer.Permitted(caller, CapProfilesList) {
		return nil, errf(CodePermissionDenied, "", "调用方 %s 没有 profiles.list 能力", caller).
			WithHint("让用户在 tt plugin enable 时授予 profiles.list")
	}
	all, err := s.Profiles()
	if err != nil {
		return nil, err
	}
	envs, allEnvs := s.opt.Authorizer.AllowedEnvs(caller, CapProfilesList)
	if allEnvs {
		return all, nil
	}
	allow := map[string]bool{}
	for _, e := range envs {
		allow[e] = true
	}
	out := make([]Profile, 0, len(allow))
	for _, p := range all {
		if allow[p.Env] {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Env < out[j].Env })
	return out, nil
}

// profileOf 把一个环境压成脱敏元数据。
func profileOf(e *config.NamedSsh) Profile {
	p := Profile{
		Env:     e.Name,
		Zone:    e.Zone,
		HasSSH:  e.Host != "",
		SSHHost: e.Host,
		SSHPort: e.Port,
		SSHUser: e.User,
	}
	if e.DB != nil {
		p.HasDB = true
		p.DBType = e.DB.Type
		p.DBAddress = e.DB.Address()
		p.DBViaSSH = e.DB.ViaSSH != nil
		p.DBReadonlySQL = e.DB.ReadonlySQLEnabled()
		for _, a := range e.DB.Accounts {
			if a.Account != "" {
				p.DBAccounts = append(p.DBAccounts, a.Account)
			}
		}
	}
	return p
}

// resolveEnv 解析环境并做授权检查。
//
// 顺序是**授权在前、存在性在后**：没授权的调用方连"这个环境存不存在"都不该问出来。
func (s *Service) resolveEnv(caller Caller, cap Capability, env string) (*config.NamedSsh, *Error) {
	env = normalizeEnvName(env)
	if !s.opt.Authorizer.Permitted(caller, cap) {
		return nil, errf(CodePermissionDenied, env, "调用方 %s 没有 %s 能力", caller, cap).
			WithHint("让用户在 tt plugin enable 时授予 %s", cap)
	}
	if env == "" {
		return nil, errf(CodeEnvUnknown, "", "必须显式指定环境（Host API 没有默认环境）").
			WithHint("先调 ProfilesService.List 看有哪些环境")
	}
	if !s.envAllowed(caller, cap, env) {
		return nil, errf(CodeEnvNotAuthorized, env, "调用方 %s 未获授权使用环境 %q", caller, env).
			WithHint("让用户在 tt plugin enable 时为该插件授权这个环境")
	}
	root, cerr := s.loadConfig()
	if cerr != nil {
		return nil, cerr
	}
	ref, err := root.Hosts.Resolve(env)
	if err != nil {
		return nil, &Error{Code: CodeEnvUnknown, Env: env, Message: err.Error(), Cause: err}
	}
	return ref.Env, nil
}

// envAllowed 判定某环境是否在授权范围内。
func (s *Service) envAllowed(caller Caller, cap Capability, env string) bool {
	envs, all := s.opt.Authorizer.AllowedEnvs(caller, cap)
	if all {
		return true
	}
	for _, e := range envs {
		if e == env {
			return true
		}
	}
	return false
}

// Reload 配置变更后调用：递增代次并关闭全部会话。
//
// 为什么直接关而不是"让旧会话继续跑"：配置里的库/服务器地址变了以后，旧会话连着的是
// **旧目标**，而调用方看到的仍是同一个环境名 —— 那正是"查错了地方却看起来正常"。
func (s *Service) Reload() {
	s.mu.Lock()
	s.generation++
	victims := make([]*session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		victims = append(victims, sess)
	}
	s.sessions = map[string]*session{}
	s.mu.Unlock()
	for _, sess := range victims {
		s.closeSessionLocked(sess)
	}
	s.emit(Event{Action: "reload", Caller: BuiltinCaller(), OK: true})
}

// CloseAll 关闭全部会话（TT 退出 / 授权撤销时用）。
func (s *Service) CloseAll() {
	s.mu.Lock()
	s.closed = true
	victims := make([]*session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		victims = append(victims, sess)
	}
	s.sessions = map[string]*session{}
	s.mu.Unlock()
	for _, sess := range victims {
		s.closeSessionLocked(sess)
	}
}

// CloseForPlugin 关闭某插件的全部会话（插件禁用/停止/授权撤销）。
func (s *Service) CloseForPlugin(pluginID string) int {
	s.mu.Lock()
	var victims []*session
	for id, sess := range s.sessions {
		if sess.caller.PluginID == pluginID && !sess.caller.IsBuiltin() {
			victims = append(victims, sess)
			delete(s.sessions, id)
		}
	}
	s.mu.Unlock()
	for _, sess := range victims {
		s.closeSessionLocked(sess)
	}
	return len(victims)
}

// CloseForInstance 关闭某插件**某一个实例**的会话。
//
// 一次性 CLI 退出时只该收拾自己开的会话：同一插件的常驻服务可能正在跑，
// 把它的连接一起关掉等于把另一个进程的工作打断。
func (s *Service) CloseForInstance(pluginID, instanceID string) int {
	if pluginID == "" || instanceID == "" {
		return 0
	}
	s.mu.Lock()
	var victims []*session
	for id, sess := range s.sessions {
		if sess.caller.PluginID == pluginID && sess.caller.InstanceID == instanceID && !sess.caller.IsBuiltin() {
			victims = append(victims, sess)
			delete(s.sessions, id)
		}
	}
	s.mu.Unlock()
	for _, sess := range victims {
		s.closeSessionLocked(sess)
	}
	return len(victims)
}

// SessionCount 当前存活会话数（诊断用）。
func (s *Service) SessionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions)
}

// emit 发一条审计事件（不含秘密、SQL 文本或命令内容）。
func (s *Service) emit(ev Event) {
	if s.opt.Audit == nil {
		return
	}
	s.opt.Audit(ev)
}

// newHandle 生成不透明随机句柄。带类型前缀只是为读日志方便，**不做任何权限判断**。
func (s *Service) newHandle(kind Kind) (string, *Error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", &Error{Code: CodeConfigBroken, Message: "生成会话句柄失败: " + err.Error(), Cause: err}
	}
	prefix := "db_"
	if kind == KindSSH {
		prefix = "ssh_"
	}
	return prefix + hex.EncodeToString(b), nil
}

// lookup 取一个属于 caller 的会话。
//
// 不属于 caller / 已关闭 / 不存在一律 SESSION_NOT_FOUND —— 三者不能区分，
// 否则句柄可以被拿来探测"别的插件开了几个会话"。
func (s *Service) lookup(caller Caller, handle string, want Kind) (*session, *Error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errf(CodeClosed, "", "连接服务已关闭")
	}
	sess, ok := s.sessions[handle]
	if !ok || sess.closed || sess.caller != caller {
		return nil, errf(CodeSessionNotFound, "", "会话不存在或不属于当前调用方").
			WithHint("句柄只在建立它的进程/插件实例内有效")
	}
	if want != "" && sess.kind != want {
		return nil, errf(CodeSessionKind, sess.env, "句柄类型是 %s，不能当 %s 用", sess.kind, want)
	}
	if sess.generation != s.generation {
		return nil, errf(CodeConfigChanged, sess.env, "配置已变更，该会话已失效").
			WithHint("重新 Open 一个会话")
	}
	sess.lastUsed = s.opt.Now()
	return sess, nil
}

// closeSessionLocked 释放会话持有的资源（调用方已把它从表里摘掉）。
func (s *Service) closeSessionLocked(sess *session) {
	if sess == nil || sess.closed {
		return
	}
	sess.closed = true
	if sess.db != nil {
		sess.db.Close()
	}
	if sess.tun != nil {
		_ = sess.tun.Close()
	}
	if sess.ssh != nil {
		_ = sess.ssh.Close()
	}
}

// accountFor 校验并选择账号：account 为空时沿用客户端直连惯例（账号清单首项）。
func accountFor(c *dbconfig.Connection, account string) *Error {
	account = normalizeEnvName(account)
	if account == "" {
		return nil
	}
	for _, a := range c.Accounts {
		if a.Account == account {
			c.Account = account
			return nil
		}
	}
	names := make([]string, 0, len(c.Accounts))
	for _, a := range c.Accounts {
		names = append(names, a.Account)
	}
	return errf(CodeAccountUnknown, "", "账号 %q 不在该环境的账号清单里（可用: %s）",
		account, joinOrNone(names))
}

// joinOrNone 把清单拼成可读字符串；空清单给一个明确的"没有"。
func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "（没有配置任何账号）"
	}
	return fmt.Sprint(items)
}
