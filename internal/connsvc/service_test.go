package connsvc

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"testing"
	"time"

	"tt/internal/config"
	"tt/internal/dbconfig"
	"tt/internal/erpdb"
	"tt/internal/host"
	"tt/internal/sshtun"
)

// ---- 测试夹具 ----

// testConfig 写一份含两个环境的配置，返回路径。
func testConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	hosts := config.Hosts{SSHs: []config.NamedSsh{
		{
			Name:      "prod",
			SSHConfig: config.SSHConfig{Host: "10.0.0.1", Port: 22, User: "u", Password: "ssh-secret-1"},
			Zone:      "36",
			DB: &dbconfig.Connection{
				Type: "oracle", Host: "db1", Port: 1521, Service: "t35prd",
				Accounts: []dbconfig.DBAcct{{Account: "ds", Password: "db-secret-1"}},
			},
		},
		{
			Name:      "test",
			SSHConfig: config.SSHConfig{Host: "10.0.0.2", Port: 2222, User: "u2", Password: "ssh-secret-2"},
			DB: &dbconfig.Connection{
				Type: "kingbase", Host: "db2", Port: 54321, Database: "topprd",
				Accounts: []dbconfig.DBAcct{{Account: "ds", Password: "db-secret-2"}},
				ViaSSH:   &dbconfig.ViaSSH{Host: "10.0.0.2", Port: 2222, User: "u2", Password: "ssh-secret-2"},
			},
		},
	}}
	if err := config.SaveHosts(path, hosts); err != nil {
		t.Fatalf("写测试配置失败: %v", err)
	}
	return path
}

// fakeAuth 可控的授权来源。
type fakeAuth struct {
	caps map[Capability]bool
	// envs 各能力下授权的环境；all=true 表示全部。
	envs map[Capability][]string
	all  bool
}

func (f fakeAuth) Permitted(_ Caller, c Capability) bool { return f.caps[c] }
func (f fakeAuth) AllowedEnvs(_ Caller, c Capability) ([]string, bool) {
	if f.all {
		return nil, true
	}
	return f.envs[c], false
}

// fakeSession erpdb.Session 的替身。
type fakeSession struct {
	typ      string
	readOnly bool
	version  string
	versionE error
	queryErr error
	execErr  error
	cols     []string
	rows     [][]string
	execN    int64
	block    bool // Query 阻塞到 ctx 结束（测超时）
	closed   bool
	execSQL  []string
}

func (f *fakeSession) Type() string { return f.typ }
func (f *fakeSession) ServerVersion(ctx context.Context) (string, error) {
	return f.version, f.versionE
}
func (f *fakeSession) Query(ctx context.Context, _ string) ([]string, [][]string, error) {
	if f.block {
		<-ctx.Done()
		return nil, nil, ctx.Err()
	}
	if f.queryErr != nil {
		return nil, nil, f.queryErr
	}
	return f.cols, f.rows, nil
}
func (f *fakeSession) Exec(_ context.Context, sql string) (int64, error) {
	if f.readOnly {
		return 0, erpdb.ErrReadOnlySession
	}
	f.execSQL = append(f.execSQL, sql)
	return f.execN, f.execErr
}
func (f *fakeSession) ReadOnly() bool { return f.readOnly }
func (f *fakeSession) Close()         { f.closed = true }

// fakeTunnel connsvc.Tunnel 的替身。
type fakeTunnel struct {
	port   int
	closed bool
}

func (f *fakeTunnel) LocalPort() int { return f.port }
func (f *fakeTunnel) Close() error   { f.closed = true; return nil }

// fakeSSH SSHConn 的替身。
type fakeSSH struct {
	output   string
	err      error
	block    bool
	closed   bool
	lastCmd  string
	lastWait time.Duration
}

func (f *fakeSSH) Output(cmd string, timeout time.Duration) (string, error) {
	f.lastCmd, f.lastWait = cmd, timeout
	if f.block {
		return "", context.DeadlineExceeded
	}
	return f.output, f.err
}
func (f *fakeSSH) OutputContext(ctx context.Context, cmd string, timeout time.Duration) (string, error) {
	f.lastCmd, f.lastWait = cmd, timeout
	if f.block {
		<-ctx.Done()
		return "", ctx.Err()
	}
	return f.output, f.err
}
func (f *fakeSSH) Close() error { f.closed = true; return nil }

// newTestService 造一个带假依赖的服务。
func newTestService(t *testing.T, opt Options) *Service {
	t.Helper()
	if opt.ConfigPath == "" {
		opt.ConfigPath = testConfig(t)
	}
	if opt.Now == nil {
		opt.Now = time.Now
	}
	return New(opt)
}

// ---- 能力与脱敏 ----

// TestCapabilitiesOnlyListWhatIsDelivered 未交付的能力不能出现在清单里。
func TestCapabilitiesOnlyListWhatIsDelivered(t *testing.T) {
	s := newTestService(t, Options{})
	c := s.Capabilities()
	if c.HostAPIVersion != HostAPIVersion {
		t.Fatalf("HostAPI 版本 = %q", c.HostAPIVersion)
	}
	want := map[string]bool{FeatureDBQuery: true, FeatureDBExec: true, FeatureSSHExec: true, FeatureDBReadOnlySession: true}
	for _, f := range c.Features {
		if !want[f] {
			t.Errorf("声明了不存在的能力 %q", f)
		}
		delete(want, f)
	}
	for f := range want {
		t.Errorf("漏了已实现的能力 %q", f)
	}
}

// TestProfileNeverCarriesSecrets 脱敏元数据里不能出现任何口令。
//
// 这条测的是"结构上就拿不到"：Profile 没有口令字段，而这里再确认账号清单只给名字。
func TestProfileNeverCarriesSecrets(t *testing.T) {
	s := newTestService(t, Options{})
	profs, err := s.Profiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(profs) != 2 {
		t.Fatalf("环境数 = %d, want 2", len(profs))
	}
	dump := fmt.Sprintf("%v", profs)
	for _, secret := range []string{"db-secret-1", "db-secret-2", "ssh-secret-1", "ssh-secret-2"} {
		if strings.Contains(dump, secret) {
			t.Fatalf("脱敏元数据里出现了口令 %q: %s", secret, dump)
		}
	}
	prod := profs[0]
	if prod.Env != "prod" || !prod.HasDB || prod.DBType != "oracle" {
		t.Fatalf("prod 元数据不对: %+v", prod)
	}
	if len(prod.DBAccounts) != 1 || prod.DBAccounts[0] != "ds" {
		t.Fatalf("账号清单该只有名字: %+v", prod.DBAccounts)
	}
	if prod.DBViaSSH {
		t.Error("prod 没配 viaSsh，不该标成经隧道")
	}
	if !profs[1].DBViaSSH {
		t.Error("test 配了 viaSsh，该标成经隧道")
	}
}

// ---- 授权与错误码 ----

// TestBuiltinOnlyAuthorizerRejectsPlugins 缺省授权器只放行内置调用方。
//
// 这是"忘记注入授权器时得到拒绝而不是全开"的守卫。
func TestBuiltinOnlyAuthorizerRejectsPlugins(t *testing.T) {
	s := newTestService(t, Options{})
	plugin := Caller{Kind: CallerServicePlugin, PluginID: "com.example.x", InstanceID: "i1"}

	if _, err := s.ProfilesFor(plugin); !IsCode(err, CodePermissionDenied) {
		t.Fatalf("插件该被拒，得 %v", err)
	}
	if _, _, err := s.OpenDB(context.Background(), plugin, "prod", "", true); !IsCode(err, CodePermissionDenied) {
		t.Fatalf("插件不能开库，得 %v", err)
	}
	// 内置调用方不受影响。
	if _, err := s.Profiles(); err != nil {
		t.Fatalf("内置调用方该能读元数据: %v", err)
	}
}

// TestProfilesForFiltersUnauthorizedEnvs 列表只含获授权环境。
func TestProfilesForFiltersUnauthorizedEnvs(t *testing.T) {
	s := newTestService(t, Options{Authorizer: fakeAuth{
		caps: map[Capability]bool{CapProfilesList: true},
		envs: map[Capability][]string{CapProfilesList: {"test"}},
	}})
	plugin := Caller{Kind: CallerServicePlugin, PluginID: "p", InstanceID: "i"}
	profs, err := s.ProfilesFor(plugin)
	if err != nil {
		t.Fatal(err)
	}
	if len(profs) != 1 || profs[0].Env != "test" {
		t.Fatalf("该只返回已授权环境，得 %+v", profs)
	}

	// 空授权 → 空清单（不是错误，也不是全部）。
	s2 := newTestService(t, Options{Authorizer: fakeAuth{caps: map[Capability]bool{CapProfilesList: true}}})
	if profs, err := s2.ProfilesFor(plugin); err != nil || len(profs) != 0 {
		t.Fatalf("空授权该得空清单，得 %+v / %v", profs, err)
	}
}

// TestUnauthorizedEnvDiffersFromUnknownEnv 未授权与不存在的区别只对已授权方可见。
//
// 未授权的调用方**不该**问出环境存不存在：两种情况都回 ENV_NOT_AUTHORIZED。
func TestUnauthorizedEnvDiffersFromUnknownEnv(t *testing.T) {
	plugin := Caller{Kind: CallerServicePlugin, PluginID: "p", InstanceID: "i"}
	s := newTestService(t, Options{Authorizer: fakeAuth{
		caps: map[Capability]bool{CapDBConnect: true},
		envs: map[Capability][]string{CapDBConnect: {"prod"}},
	}})

	// 存在但没授权。
	if _, _, err := s.OpenDB(context.Background(), plugin, "test", "", true); !IsCode(err, CodeEnvNotAuthorized) {
		t.Fatalf("未授权环境该报 ENV_NOT_AUTHORIZED，得 %v", err)
	}
	// 不存在也没授权 → 同样的码（不泄露存在性）。
	if _, _, err := s.OpenDB(context.Background(), plugin, "ghost", "", true); !IsCode(err, CodeEnvNotAuthorized) {
		t.Fatalf("不该泄露环境是否存在，得 %v", err)
	}
	// 已授权方才看得到"不存在"：换一个只放行内置调用方的服务。
	s2 := newTestService(t, Options{})
	if _, _, err := s2.OpenDB(context.Background(), BuiltinCaller(), "ghost", "", true); !IsCode(err, CodeEnvUnknown) {
		t.Fatalf("已授权方该得 ENV_UNKNOWN，得 %v", err)
	}
}

// TestNoDefaultEnv 空环境名一律报错，没有默认环境回退。
func TestNoDefaultEnv(t *testing.T) {
	s := newTestService(t, Options{})
	if _, _, err := s.OpenDB(context.Background(), BuiltinCaller(), "", "", true); !IsCode(err, CodeEnvUnknown) {
		t.Fatalf("空环境该报错，得 %v", err)
	}
	if _, err := s.TestDB(context.Background(), BuiltinCaller(), "  ", ""); !IsCode(err, CodeEnvUnknown) {
		t.Fatalf("全空白环境该报错，得 %v", err)
	}
}

// TestAccountMustBeInList 账号只能在清单内选。
func TestAccountMustBeInList(t *testing.T) {
	s := newTestService(t, Options{})
	if _, _, err := s.OpenDB(context.Background(), BuiltinCaller(), "prod", "nobody", true); !IsCode(err, CodeAccountUnknown) {
		t.Fatalf("清单外账号该报错，得 %v", err)
	}
}

// ---- 会话所有权 ----

// TestSessionOwnershipIsEnforced 拿别人的句柄一律 SESSION_NOT_FOUND。
func TestSessionOwnershipIsEnforced(t *testing.T) {
	fs := &fakeSession{typ: "oracle", readOnly: true, version: "v"}
	auth := fakeAuth{caps: map[Capability]bool{CapDBConnect: true}, all: true}
	s := newTestService(t, Options{
		Authorizer: auth,
		OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			return fs, nil
		},
	})
	owner := Caller{Kind: CallerCLIPlugin, PluginID: "p1", InstanceID: "i1"}
	other := Caller{Kind: CallerCLIPlugin, PluginID: "p2", InstanceID: "i2"}

	handle, _, err := s.OpenDB(context.Background(), owner, "prod", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DBQuery(context.Background(), owner, handle, "SELECT 1", 10); err != nil {
		t.Fatalf("owner 该能用: %v", err)
	}
	// 别的插件：同一个码，不泄露"这个句柄存在"。
	if _, err := s.DBQuery(context.Background(), other, handle, "SELECT 1", 10); !IsCode(err, CodeSessionNotFound) {
		t.Fatalf("跨插件该被拒，得 %v", err)
	}
	if err := s.Close(other, handle); !IsCode(err, CodeSessionNotFound) {
		t.Fatalf("跨插件关闭该被拒，得 %v", err)
	}
	// 伪造句柄同码。
	if _, err := s.DBQuery(context.Background(), owner, "db_deadbeef", "SELECT 1", 10); !IsCode(err, CodeSessionNotFound) {
		t.Fatalf("伪造句柄该被拒，得 %v", err)
	}
}

// TestHandleKindMismatch 拿 SSH 句柄查库要被明确拒绝。
func TestHandleKindMismatch(t *testing.T) {
	auth := fakeAuth{caps: map[Capability]bool{CapDBConnect: true, CapSSHConnect: true}, all: true}
	fssh := &fakeSSH{output: "ok"}
	s := newTestService(t, Options{
		Authorizer: auth,
		DialSSH:    func(config.SSHConfig) (SSHConn, error) { return fssh, nil },
	})
	handle, err := s.OpenSSH(context.Background(), BuiltinCaller(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DBQuery(context.Background(), BuiltinCaller(), handle, "SELECT 1", 10); !IsCode(err, CodeSessionKind) {
		t.Fatalf("类型不符该报 SESSION_KIND_MISMATCH，得 %v", err)
	}
}

// TestCloseReleasesResources 关闭会话要释放库连接与隧道，且只能关一次。
func TestCloseReleasesResources(t *testing.T) {
	fs := &fakeSession{typ: "oracle", readOnly: true, version: "v"}
	tun := &fakeTunnel{port: 4567}
	s := newTestService(t, Options{
		DialTunnel: func(sshtunOptions) (Tunnel, error) { return tun, nil },
		OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			return fs, nil
		},
	})
	handle, _, err := s.OpenDB(context.Background(), BuiltinCaller(), "test", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if tun.closed {
		t.Fatal("打开期间隧道不该关")
	}
	if err := s.Close(BuiltinCaller(), handle); err != nil {
		t.Fatal(err)
	}
	if !fs.closed {
		t.Error("关闭会话必须关掉库连接")
	}
	if !tun.closed {
		t.Error("关闭会话必须关掉隧道")
	}
	if err := s.Close(BuiltinCaller(), handle); !IsCode(err, CodeSessionNotFound) {
		t.Fatalf("重复关闭该报会话不存在，得 %v", err)
	}
	if s.SessionCount() != 0 {
		t.Errorf("会话表该空了，得 %d", s.SessionCount())
	}
}

// TestCloseForPluginAndReload 插件停止 / 配置变更都会作废会话。
func TestCloseForPluginAndReload(t *testing.T) {
	auth := fakeAuth{caps: map[Capability]bool{CapDBConnect: true, CapSSHConnect: true}, all: true}
	newSvc := func() (*Service, *fakeSSH) {
		fssh := &fakeSSH{output: "ok"}
		s := newTestService(t, Options{
			Authorizer: auth,
			DialSSH:    func(config.SSHConfig) (SSHConn, error) { return fssh, nil },
		})
		return s, fssh
	}
	plugin := Caller{Kind: CallerCLIPlugin, PluginID: "p1", InstanceID: "i1"}

	t.Run("按插件关闭", func(t *testing.T) {
		s, fssh := newSvc()
		h1, err := s.OpenSSH(context.Background(), plugin, "prod")
		if err != nil {
			t.Fatal(err)
		}
		h2, err := s.OpenSSH(context.Background(), BuiltinCaller(), "prod")
		if err != nil {
			t.Fatal(err)
		}
		if n := s.CloseForPlugin("p1"); n != 1 {
			t.Fatalf("该关 1 个，得 %d", n)
		}
		if _, err := s.SSHExec(context.Background(), plugin, h1, "hostname", time.Second); !IsCode(err, CodeSessionNotFound) {
			t.Fatalf("插件会话该没了，得 %v", err)
		}
		if _, err := s.SSHExec(context.Background(), BuiltinCaller(), h2, "hostname", time.Second); err != nil {
			t.Fatalf("内置会话不该被牵连: %v", err)
		}
		_ = fssh
	})

	t.Run("配置变更作废全部", func(t *testing.T) {
		s, fssh := newSvc()
		h, err := s.OpenSSH(context.Background(), BuiltinCaller(), "prod")
		if err != nil {
			t.Fatal(err)
		}
		s.Reload()
		if !fssh.closed {
			t.Error("Reload 该关掉底层 SSH 连接")
		}
		if _, err := s.SSHExec(context.Background(), BuiltinCaller(), h, "hostname", time.Second); !IsCode(err, CodeSessionNotFound) {
			t.Fatalf("配置变更后旧句柄该失效，得 %v", err)
		}
	})
}

// TestCloseAllRejectsNewSessions CloseAll 之后不再接受新会话。
func TestCloseAllRejectsNewSessions(t *testing.T) {
	auth := fakeAuth{caps: map[Capability]bool{CapSSHConnect: true}, all: true}
	s := newTestService(t, Options{
		Authorizer: auth,
		DialSSH:    func(config.SSHConfig) (SSHConn, error) { return &fakeSSH{}, nil },
	})
	s.CloseAll()
	if _, err := s.OpenSSH(context.Background(), BuiltinCaller(), "prod"); !IsCode(err, CodeClosed) {
		t.Fatalf("服务关闭后该拒新会话，得 %v", err)
	}
}

// ---- 数据面 ----

// TestQueryTruncates 超过 maxRows 即截断并标记，不静默丢行。
func TestQueryTruncates(t *testing.T) {
	rows := make([][]string, 5)
	for i := range rows {
		rows[i] = []string{fmt.Sprint(i)}
	}
	fs := &fakeSession{typ: "oracle", readOnly: true, version: "v", cols: []string{"n"}, rows: rows}
	s := newTestService(t, Options{
		OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			return fs, nil
		},
	})
	h, _, err := s.OpenDB(context.Background(), BuiltinCaller(), "prod", "", true)
	if err != nil {
		t.Fatal(err)
	}
	res, qerr := s.DBQuery(context.Background(), BuiltinCaller(), h, "SELECT n", 3)
	if qerr != nil {
		t.Fatal(qerr)
	}
	if len(res.Rows) != 3 || !res.Truncated {
		t.Fatalf("该截断到 3 行并标记，得 %d 行 truncated=%v", len(res.Rows), res.Truncated)
	}
	// 0 = 缺省上限；行数少于上限时不标记截断。
	res2, qerr := s.DBQuery(context.Background(), BuiltinCaller(), h, "SELECT n", 0)
	if qerr != nil {
		t.Fatal(qerr)
	}
	if len(res2.Rows) != 5 || res2.Truncated {
		t.Fatalf("缺省上限下 5 行该全给，得 %d 行 truncated=%v", len(res2.Rows), res2.Truncated)
	}
}

// TestExecOnReadOnlySessionIsRejected 只读会话上不允许写，且报专门的码。
func TestExecOnReadOnlySessionIsRejected(t *testing.T) {
	ro := &fakeSession{typ: "oracle", readOnly: true, version: "v"}
	open := func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
		return ro, nil
	}
	s := newTestService(t, Options{OpenDB: open})

	h, _, err := s.OpenDB(context.Background(), BuiltinCaller(), "prod", "", true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DBExec(context.Background(), BuiltinCaller(), h, "DELETE FROM x"); !IsCode(err, CodeSessionReadOnly) {
		t.Fatalf("只读会话上写该报 DB_SESSION_READONLY，得 %v", err)
	}
}

// TestExecOnWritableSessionWorks 普通会话可以写 —— 边界是数据库账号权限，不是 TT。
func TestExecOnWritableSessionWorks(t *testing.T) {
	rw := &fakeSession{typ: "oracle", readOnly: false, version: "v", execN: 7}
	var gotOpt erpdb.SessionOptions
	s := newTestService(t, Options{
		OpenDB: func(_ context.Context, _ dbconfig.Connection, o erpdb.SessionOptions) (erpdb.Session, error) {
			gotOpt = o
			return rw, nil
		},
	})
	h, res, err := s.OpenDB(context.Background(), BuiltinCaller(), "prod", "", false)
	if err != nil {
		t.Fatal(err)
	}
	if gotOpt.ReadOnly {
		t.Error("请求可写会话时不该传只读标志")
	}
	if res.ReadOnly {
		t.Error("结果该说这是可写会话")
	}
	n, xerr := s.DBExec(context.Background(), BuiltinCaller(), h, "UPDATE x SET y=1")
	if xerr != nil {
		t.Fatal(xerr)
	}
	if n != 7 {
		t.Fatalf("受影响行数 = %d, want 7", n)
	}
	if len(rw.execSQL) != 1 {
		t.Fatalf("该真的执行一次，得 %d", len(rw.execSQL))
	}
}

// TestQueryTimeoutIsTimeoutCode 超时归到 TIMEOUT，不混进普通查询失败。
func TestQueryTimeoutIsTimeoutCode(t *testing.T) {
	fs := &fakeSession{typ: "oracle", readOnly: true, version: "v", block: true}
	s := newTestService(t, Options{
		OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			return fs, nil
		},
	})
	h, _, err := s.OpenDB(context.Background(), BuiltinCaller(), "prod", "", true)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, qerr := s.DBQuery(ctx, BuiltinCaller(), h, "SELECT slow", 10); !IsCode(qerr, CodeTimeout) {
		t.Fatalf("超时该报 TIMEOUT，得 %v", qerr)
	}
}

// TestSSHExecReturnsExitCode 非零退出码是结果，不是错误。
func TestSSHExecReturnsExitCode(t *testing.T) {
	fssh := &fakeSSH{output: "boom", err: &fakeExitErr{code: 3}}
	s := newTestService(t, Options{
		DialSSH: func(config.SSHConfig) (SSHConn, error) { return fssh, nil },
	})
	h, err := s.OpenSSH(context.Background(), BuiltinCaller(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	res, xerr := s.SSHExec(context.Background(), BuiltinCaller(), h, "false", 5*time.Second)
	if xerr != nil {
		t.Fatalf("非零退出码不该是错误: %v", xerr)
	}
	if res.ExitCode != 3 || res.Output != "boom" {
		t.Fatalf("该带回 exit=3 与输出，得 %+v", res)
	}
}

// fakeExitErr 模拟"命令跑完但非零退出"——host.ExitStatus 认的是 ssh.ExitError，
// 这里用同样的可识别形状（见下：ExitStatus 由 fakeSSH 的 err 类型决定不了，
// 所以这条用例直接验证 connsvc 对 -1 的处理分支）。
type fakeExitErr struct{ code int }

func (e *fakeExitErr) Error() string { return fmt.Sprintf("exit status %d", e.code) }

// ExitStatus 让 host.ExitStatus 认得它（判据是接口，见 host.ExitStatus 的注释）。
func (e *fakeExitErr) ExitStatus() int { return e.code }

// TestSSHExecTimeoutIsMarked 超时被标记为 timed_out，而不是抛一个不可读的错误。
func TestSSHExecTimeoutIsMarked(t *testing.T) {
	fssh := &fakeSSH{block: true}
	s := newTestService(t, Options{
		DialSSH: func(config.SSHConfig) (SSHConn, error) { return fssh, nil },
	})
	h, err := s.OpenSSH(context.Background(), BuiltinCaller(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	res, xerr := s.SSHExec(context.Background(), BuiltinCaller(), h, "sleep 100", 30*time.Millisecond)
	if xerr != nil {
		t.Fatalf("超时该是结果: %v", xerr)
	}
	if !res.TimedOut || res.ExitCode != -1 {
		t.Fatalf("该标 timed_out，得 %+v", res)
	}
}

// ---- 隧道编排 ----

// TestViaSshRewritesConnectionAndClosesTunnel 配了 viaSsh 时必须经隧道，且地址如实回报。
func TestViaSshRewritesConnectionAndClosesTunnel(t *testing.T) {
	tun := &fakeTunnel{port: 45678}
	var gotConn dbconfig.Connection
	var gotTunOpt sshtunOptions
	s := newTestService(t, Options{
		DialTunnel: func(o sshtunOptions) (Tunnel, error) {
			gotTunOpt = o
			return tun, nil
		},
		OpenDB: func(_ context.Context, c dbconfig.Connection, _ erpdb.SessionOptions) (erpdb.Session, error) {
			gotConn = c
			return &fakeSession{typ: "kingbase", readOnly: true, version: "KingbaseES V8"}, nil
		},
	})
	res := s.TestDBConn(context.Background(), *mustConn(t, "test"))
	if !res.OK {
		t.Fatalf("该成功，得 %+v", res)
	}
	if gotTunOpt.Host != "10.0.0.2" || gotTunOpt.RemoteHost != "db2" || gotTunOpt.RemotePort != 54321 {
		t.Fatalf("隧道参数不对: %+v", gotTunOpt)
	}
	if gotConn.Host != "127.0.0.1" || gotConn.Port != 45678 {
		t.Fatalf("连接该改写到本地转发端口，得 %s:%d", gotConn.Host, gotConn.Port)
	}
	if res.Address != "127.0.0.1:45678/topprd" {
		t.Errorf("address 该是隧道本地地址，得 %q", res.Address)
	}
	if !res.ViaTunnel {
		t.Error("该标成经隧道")
	}
	if !tun.closed {
		t.Error("测试结束必须关掉隧道")
	}
}

// TestTunnelFailureIsStageTunnel 隧道建不起来时，阶段要如实标 tunnel。
func TestTunnelFailureIsStageTunnel(t *testing.T) {
	s := newTestService(t, Options{
		DialTunnel: func(sshtunOptions) (Tunnel, error) { return nil, errors.New("ssh refused") },
	})
	res := s.TestDBConn(context.Background(), *mustConn(t, "test"))
	if res.OK || res.Stage != StageTunnel {
		t.Fatalf("该是 tunnel 阶段失败，得 %+v", res)
	}
	if !strings.Contains(res.Message, "ssh refused") {
		t.Errorf("message 该带上底层原因，得 %q", res.Message)
	}
}

// TestTestDBConnNeverLeaksSecretsIntoMessage 失败信息里不该带口令。
func TestTestDBConnNeverLeaksSecretsIntoMessage(t *testing.T) {
	s := newTestService(t, Options{
		DialTunnel: func(sshtunOptions) (Tunnel, error) {
			return nil, errors.New("auth failed")
		},
	})
	res := s.TestDBConn(context.Background(), *mustConn(t, "test"))
	for _, secret := range []string{"db-secret-2", "ssh-secret-2"} {
		if strings.Contains(res.Message, secret) {
			t.Fatalf("失败信息泄露了口令: %q", res.Message)
		}
	}
}

// mustConn 取某环境的数据库连接（测试辅助）。
//
// 每次现写一份同内容的临时配置：夹具内容确定，取到的连接与该测试自身用的那份等价。
func mustConn(t *testing.T, env string) *dbconfig.Connection {
	t.Helper()
	root, err := config.Load(testConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	ref, err := root.Hosts.Resolve(env)
	if err != nil {
		t.Fatal(err)
	}
	return ref.Env.DB
}

// sshtunOptions 是本包对 sshtun.Options 的别名（测试里少写一个包名）。
type sshtunOptions = sshtun.Options

// TestSSHTestUsesInjectedTestSSH 内置 TestSSH 也走连接服务，宿主注入缝生效。
func TestSSHTestUsesInjectedTestSSH(t *testing.T) {
	var got config.SSHConfig
	s := newTestService(t, Options{
		TestSSH: func(cfg config.SSHConfig) (*host.SSHTestOut, error) {
			got = cfg
			return &host.SSHTestOut{Hostname: "t100dev"}, nil
		},
	})
	res, err := s.TestSSH(context.Background(), BuiltinCaller(), "prod")
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Banner != "t100dev" {
		t.Fatalf("该成功并带回主机名，得 %+v", res)
	}
	if got.Host != "10.0.0.1" || got.Port != 22 {
		t.Fatalf("该用该环境的 SSH 配置，得 %+v", got)
	}
	if s.SessionCount() != 0 {
		t.Error("测试不该留下会话")
	}
}
