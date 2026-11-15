package connsvc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/host"
	"github.com/luyaogod/TT/internal/sshtun"
	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/pkg/erpdb"
)

// TestDB 测一次数据库连通性：解析配置 → （按需）建隧道 → 建连 → 取版本 → 关闭。
//
// 返回约定：
//   - error（*Error）只在**契约层**失败时返回：没能力、没授权、环境不存在、没配库、账号不存在；
//   - 连不上/认证失败这类**连接层**结果是 ok=false + stage/message 的正常返回 ——
//     设置页与插件都要能显示"卡在哪一步"，而不是拿到一个笼统的 RPC 错误。
func (s *Service) TestDB(ctx context.Context, caller Caller, env, account string) (*TestDBResult, *Error) {
	start := s.opt.Now()
	named, cerr := s.resolveEnv(caller, CapDBConnect, env)
	if cerr != nil {
		return nil, cerr
	}
	if named.DB == nil {
		return nil, errf(CodeDBNoConfig, named.Name, "该环境没有配置数据库连接（hosts.sshs[].db）").
			WithHint("在设置页补该环境的数据库连接")
	}
	conn := *named.DB
	if e := accountFor(&conn, account); e != nil {
		e.Env = named.Name
		return nil, e
	}
	res := s.TestDBConn(ctx, conn)
	res.Duration = s.opt.Now().Sub(start)
	s.emit(Event{Action: "db.test", Caller: caller, Env: named.Name, Kind: KindDB, OK: res.OK, Code: res.Stage, DurationMS: res.Duration.Milliseconds()})
	return res, nil
}

// TestDBConn 测一份**显式连接配置**（设置页表单里的当前值，不必先保存）。
//
// 这是设置页与 Host API 共用的那条实现：REST 端点传表单值，插件端点先按环境名解析
// 出同样的 dbconfig.Connection 再进来 —— 不存在"测试走一条路、查询走另一条"。
func (s *Service) TestDBConn(ctx context.Context, conn dbconfig.Connection) *TestDBResult {
	start := s.opt.Now()
	res := &TestDBResult{Stage: StageResolveConfig}
	done := func() *TestDBResult {
		res.Duration = s.opt.Now().Sub(start)
		return res
	}

	tun, e := s.dialDBConn(&conn)
	res.Address = conn.Address()
	res.DBType = conn.Type
	res.ViaTunnel = conn.ViaSSH != nil
	if e != nil {
		res.Stage = e.Stage
		if res.Stage == "" {
			res.Stage = StageConnect
		}
		res.Message = e.Message
		return done()
	}
	defer closeTunnel(tun)

	res.Stage = StageConnect
	sess, oerr := s.opt.OpenDB(ctx, conn, erpdb.SessionOptions{ReadOnly: true})
	if oerr != nil {
		res.Message = oerr.Error()
		return done()
	}
	defer sess.Close()

	res.Stage = StagePing
	ver, verr := sess.ServerVersion(ctx)
	if verr != nil {
		res.Message = verr.Error()
		return done()
	}
	res.OK = true
	res.Stage = ""
	res.ServerVersion = strings.TrimSpace(ver)
	return done()
}

// TestSSH 测一次 SSH 链路：解析配置 → 登录 → 取主机名 → 断开。
func (s *Service) TestSSH(ctx context.Context, caller Caller, env string) (*TestSSHResult, *Error) {
	named, cerr := s.resolveEnv(caller, CapSSHConnect, env)
	if cerr != nil {
		return nil, cerr
	}
	if named.Host == "" {
		return nil, errf(CodeSSHConnectFailed, named.Name, "该环境没有配置 SSH 主机").
			WithHint("在设置页补 host/port/user")
	}
	res := s.TestSSHConn(named.SSHConfig)
	s.emit(Event{Action: "ssh.test", Caller: caller, Env: named.Name, Kind: KindSSH, OK: res.OK, Code: res.Stage, DurationMS: res.Duration.Milliseconds()})
	return res, nil
}

// TestSSHConn 测一份**显式 SSH 配置**（设置页表单里的当前值）。
func (s *Service) TestSSHConn(cfg config.SSHConfig) *TestSSHResult {
	start := s.opt.Now()
	res := &TestSSHResult{Stage: StageConnect}
	out, err := s.opt.TestSSH(cfg)
	if err != nil {
		res.Message = err.Error()
		res.Duration = s.opt.Now().Sub(start)
		return res
	}
	res.OK = true
	res.Stage = ""
	res.Banner = strings.TrimSpace(out.Hostname)
	if out.Note != "" {
		// 主机名读不到不算失败（部分镜像没有 hostname），但必须如实带出去。
		res.Banner = out.Note
	}
	res.Duration = s.opt.Now().Sub(start)
	return res
}

// OpenDB 打开一条数据库会话，返回不透明句柄。
//
// readOnly 由调用方显式选择（见 OpenDBRequest.read_only 的注释）：内置只读语义的那两层
// 防线不随本方法的存在而改变，插件也不会因为调用通用会话就"继承"它。
func (s *Service) OpenDB(ctx context.Context, caller Caller, env, account string, readOnly bool) (string, OpenDBResult, *Error) {
	return s.openDB(ctx, caller, env, account, readOnly, 0)
}

// OpenDBWithEnt 按企业编号解析账号后打开会话（ent>0 时 account 被忽略）。
//
// 企业目录来自配置里的 gzou_t 同步副本；没同步过就是"该企业未知"，
// 提示用户先 `tt env ents --refresh`。
func (s *Service) OpenDBWithEnt(ctx context.Context, caller Caller, env, account string, readOnly bool, ent int) (string, OpenDBResult, *Error) {
	return s.openDB(ctx, caller, env, account, readOnly, ent)
}

func (s *Service) openDB(ctx context.Context, caller Caller, env, account string, readOnly bool, ent int) (string, OpenDBResult, *Error) {
	var out OpenDBResult
	start := s.opt.Now()

	named, cerr := s.resolveEnv(caller, CapDBConnect, env)
	if cerr != nil {
		return "", out, cerr
	}
	if named.DB == nil {
		return "", out, errf(CodeDBNoConfig, named.Name, "该环境没有配置数据库连接（hosts.sshs[].db）")
	}
	if ent > 0 {
		acct, eerr := entAccount(named, ent)
		if eerr != nil {
			return "", out, eerr
		}
		account = acct
	}
	conn := *named.DB
	if e := accountFor(&conn, account); e != nil {
		e.Env = named.Name
		return "", out, e
	}

	tun, e := s.dialDBConn(&conn)
	if e != nil {
		s.emit(Event{Action: "db.open", Caller: caller, Env: named.Name, Kind: KindDB, OK: false, Code: e.Code, DurationMS: s.opt.Now().Sub(start).Milliseconds()})
		return "", out, e
	}
	sess, oerr := s.opt.OpenDB(ctx, conn, erpdb.SessionOptions{ReadOnly: readOnly})
	if oerr != nil {
		closeTunnel(tun)
		e := &Error{Code: CodeDBConnectFailed, Env: named.Name, Stage: StageConnect, Message: oerr.Error(), Cause: oerr}
		s.emit(Event{Action: "db.open", Caller: caller, Env: named.Name, Kind: KindDB, OK: false, Code: e.Code, DurationMS: s.opt.Now().Sub(start).Milliseconds()})
		return "", out, e
	}

	handle, herr := s.newHandle(KindDB)
	if herr != nil {
		sess.Close()
		closeTunnel(tun)
		return "", out, herr
	}
	now := s.opt.Now()
	sess2 := &session{
		handle:     handle,
		kind:       KindDB,
		env:        named.Name,
		caller:     caller,
		generation: s.currentGeneration(),
		openedAt:   now,
		lastUsed:   now,
		db:         sess,
		tun:        tun,
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		s.closeSessionLocked(sess2)
		return "", out, errf(CodeClosed, "", "连接服务已关闭")
	}
	s.sessions[handle] = sess2
	s.mu.Unlock()

	out.ReadOnly = sess.ReadOnly()
	out.DBType = conn.Type
	out.Account = conn.Account
	out.Ent = ent
	s.emit(Event{Action: "db.open", Caller: caller, Env: named.Name, Kind: KindDB, OK: true, DurationMS: s.opt.Now().Sub(start).Milliseconds()})
	return handle, out, nil
}

// entAccount 企业编号 → 账号；不在清单里 / 没配账号时给稳定错误码。
func entAccount(e *config.NamedSsh, ent int) (string, *Error) {
	for _, it := range e.Ents {
		if it.Ent != ent {
			continue
		}
		acct := strings.TrimSpace(it.Account)
		if acct == "" {
			return "", errf(CodeEntNoAccount, e.Name, "企业 %d 在 gzou_t 里没有配可用账号", ent).
				WithHint("检查该企业的账号配置，或用 tt env ents --refresh 重新同步企业目录")
		}
		return acct, nil
	}
	return "", errf(CodeEntUnknown, e.Name, "企业 %d 不在该环境的启用清单里", ent).
		WithHint("先用 tt env ents --refresh 同步企业目录（或确认企业编号）")
}

// OpenSSH 打开一条 SSH 会话，返回不透明句柄。
func (s *Service) OpenSSH(ctx context.Context, caller Caller, env string) (string, *Error) {
	start := s.opt.Now()
	named, cerr := s.resolveEnv(caller, CapSSHConnect, env)
	if cerr != nil {
		return "", cerr
	}
	if named.Host == "" {
		return "", errf(CodeSSHConnectFailed, named.Name, "该环境没有配置 SSH 主机").WithHint("在设置页补 host/port/user")
	}
	conn, derr := s.opt.DialSSH(named.SSHConfig)
	if derr != nil {
		e := &Error{Code: CodeSSHConnectFailed, Env: named.Name, Stage: StageConnect, Message: derr.Error(), Cause: derr}
		s.emit(Event{Action: "ssh.open", Caller: caller, Env: named.Name, Kind: KindSSH, OK: false, Code: e.Code, DurationMS: s.opt.Now().Sub(start).Milliseconds()})
		return "", e
	}
	handle, herr := s.newHandle(KindSSH)
	if herr != nil {
		_ = conn.Close()
		return "", herr
	}
	now := s.opt.Now()
	sess := &session{
		handle:     handle,
		kind:       KindSSH,
		env:        named.Name,
		caller:     caller,
		generation: s.currentGeneration(),
		openedAt:   now,
		lastUsed:   now,
		ssh:        conn,
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		s.closeSessionLocked(sess)
		return "", errf(CodeClosed, "", "连接服务已关闭")
	}
	s.sessions[handle] = sess
	s.mu.Unlock()
	s.emit(Event{Action: "ssh.open", Caller: caller, Env: named.Name, Kind: KindSSH, OK: true, DurationMS: s.opt.Now().Sub(start).Milliseconds()})
	return handle, nil
}

// Close 关闭一个属于 caller 的会话（种类不限）。
func (s *Service) Close(caller Caller, handle string) *Error {
	s.mu.Lock()
	sess, ok := s.sessions[handle]
	if !ok || sess.closed || sess.caller != caller {
		s.mu.Unlock()
		return errf(CodeSessionNotFound, "", "会话不存在或不属于当前调用方")
	}
	delete(s.sessions, handle)
	s.mu.Unlock()
	s.closeSessionLocked(sess)
	s.emit(Event{Action: "close", Caller: caller, Env: sess.env, Kind: sess.kind, OK: true})
	return nil
}

// argsSession 支持参数绑定与批量执行的会话。
//
// 真实连接器（Oracle/Kingbase）都实现；测试替身可以只实现 Session。
// 用可选接口而不是扩 Session：老替身与老调用方不用改，能力缺失时给出稳定错误。
type argsSession interface {
	QueryArgs(ctx context.Context, sql string, params []string) ([]string, [][]string, error)
	ExecArgs(ctx context.Context, sql string, params []string) (int64, error)
	ExecBatch(ctx context.Context, stmts []erpdb.Stmt, atomic bool) (int64, int, error)
}

// DBQuery 在数据库会话上执行一条查询，按 maxRows 截断后返回。
func (s *Service) DBQuery(ctx context.Context, caller Caller, handle, sql string, maxRows int) (*QueryResult, *Error) {
	return s.dbQuery(ctx, caller, handle, sql, nil, maxRows)
}

// DBQueryArgs 带参数查询（占位符 `?`，由驱动绑定）。
func (s *Service) DBQueryArgs(ctx context.Context, caller Caller, handle, sql string, params []string, maxRows int) (*QueryResult, *Error) {
	return s.dbQuery(ctx, caller, handle, sql, params, maxRows)
}

func (s *Service) dbQuery(ctx context.Context, caller Caller, handle, sql string, params []string, maxRows int) (*QueryResult, *Error) {
	sess, lerr := s.lookup(caller, handle, KindDB)
	if lerr != nil {
		return nil, lerr
	}
	if strings.TrimSpace(sql) == "" {
		return nil, errf(CodeQueryFailed, sess.env, "SQL 为空")
	}
	limit := maxRows
	if limit <= 0 {
		limit = DefaultMaxRows
	}
	if limit > MaxMaxRows {
		limit = MaxMaxRows
	}

	qctx, cancel := withDefaultTimeout(ctx, DefaultQueryTimeout)
	defer cancel()
	start := s.opt.Now()
	var cols []string
	var rows [][]string
	var err error
	if len(params) == 0 {
		cols, rows, err = sess.db.Query(qctx, sql)
	} else {
		adv, ok := sess.db.(argsSession)
		if !ok {
			return nil, errf(CodeUnsupported, sess.env, "该会话不支持参数绑定")
		}
		cols, rows, err = adv.QueryArgs(qctx, sql, params)
	}
	dur := s.opt.Now().Sub(start)
	if err != nil {
		e := queryError(sess.env, err)
		s.emit(Event{Action: "db.query", Caller: caller, Env: sess.env, Kind: KindDB, OK: false, Code: e.Code, DurationMS: dur.Milliseconds()})
		return nil, e
	}
	res := &QueryResult{Columns: cols, Rows: rows}
	if len(res.Rows) > limit {
		res.Rows = res.Rows[:limit]
		res.Truncated = true
	}
	s.emit(Event{Action: "db.query", Caller: caller, Env: sess.env, Kind: KindDB, OK: true, DurationMS: dur.Milliseconds()})
	return res, nil
}

// DBExec 在数据库会话上执行一条写语句。
//
// 只读会话上调用它一律报 DB_SESSION_READONLY —— 不静默降级成查询，也不偷偷改成可写。
func (s *Service) DBExec(ctx context.Context, caller Caller, handle, sql string) (int64, *Error) {
	return s.dbExec(ctx, caller, handle, sql, nil)
}

// DBExecArgs 带参数写语句。
func (s *Service) DBExecArgs(ctx context.Context, caller Caller, handle, sql string, params []string) (int64, *Error) {
	return s.dbExec(ctx, caller, handle, sql, params)
}

func (s *Service) dbExec(ctx context.Context, caller Caller, handle, sql string, params []string) (int64, *Error) {
	sess, lerr := s.lookup(caller, handle, KindDB)
	if lerr != nil {
		return 0, lerr
	}
	if strings.TrimSpace(sql) == "" {
		return 0, errf(CodeExecFailed, sess.env, "SQL 为空")
	}
	if sess.db.ReadOnly() {
		e := errf(CodeSessionReadOnly, sess.env, "该会话以只读建立，不能执行写语句").
			WithHint("需要写能力时用 OpenDB(read_only=false) 重开会话")
		s.emit(Event{Action: "db.exec", Caller: caller, Env: sess.env, Kind: KindDB, OK: false, Code: e.Code})
		return 0, e
	}
	ectx, cancel := withDefaultTimeout(ctx, DefaultQueryTimeout)
	defer cancel()
	start := s.opt.Now()
	var n int64
	var err error
	if len(params) == 0 {
		n, err = sess.db.Exec(ectx, sql)
	} else {
		adv, ok := sess.db.(argsSession)
		if !ok {
			return 0, errf(CodeUnsupported, sess.env, "该会话不支持参数绑定")
		}
		n, err = adv.ExecArgs(ectx, sql, params)
	}
	dur := s.opt.Now().Sub(start)
	if err != nil {
		code := CodeExecFailed
		if isTimeout(ectx, err) {
			code = CodeTimeout
		}
		e := &Error{Code: code, Env: sess.env, Stage: StageExec, Message: err.Error(), Cause: err}
		s.emit(Event{Action: "db.exec", Caller: caller, Env: sess.env, Kind: KindDB, OK: false, Code: code, DurationMS: dur.Milliseconds()})
		return 0, e
	}
	s.emit(Event{Action: "db.exec", Caller: caller, Env: sess.env, Kind: KindDB, OK: true, DurationMS: dur.Milliseconds()})
	return n, nil
}

// DBBatch 批量执行多条语句。atomic=true 时整批一个事务，失败全部回滚。
//
// 返回 failedIndex（0 基）；-1 表示全部成功。失败语句下标写进错误消息，
// 插件据此知道"卡在第几条"。
func (s *Service) DBBatch(ctx context.Context, caller Caller, handle string, stmts []erpdb.Stmt, atomic bool) (int64, int, *Error) {
	sess, lerr := s.lookup(caller, handle, KindDB)
	if lerr != nil {
		return 0, -1, lerr
	}
	if len(stmts) == 0 {
		return 0, -1, errf(CodeBatchFailed, sess.env, "批量语句为空")
	}
	if sess.db.ReadOnly() {
		e := errf(CodeSessionReadOnly, sess.env, "该会话以只读建立，不能执行写语句").
			WithHint("需要写能力时用 OpenDB(read_only=false) 重开会话")
		s.emit(Event{Action: "db.batch", Caller: caller, Env: sess.env, Kind: KindDB, OK: false, Code: e.Code})
		return 0, -1, e
	}
	adv, ok := sess.db.(argsSession)
	if !ok {
		return 0, -1, errf(CodeUnsupported, sess.env, "该会话不支持批量执行")
	}
	bctx, cancel := withDefaultTimeout(ctx, DefaultQueryTimeout)
	defer cancel()
	start := s.opt.Now()
	n, failed, err := adv.ExecBatch(bctx, stmts, atomic)
	dur := s.opt.Now().Sub(start)
	if err != nil {
		code := CodeBatchFailed
		if isTimeout(bctx, err) {
			code = CodeTimeout
		}
		msg := err.Error()
		if failed >= 0 {
			msg = fmt.Sprintf("第 %d 条失败: %s", failed+1, msg)
		}
		e := &Error{Code: code, Env: sess.env, Stage: StageExec, Message: msg, Cause: err}
		s.emit(Event{Action: "db.batch", Caller: caller, Env: sess.env, Kind: KindDB, OK: false, Code: code, DurationMS: dur.Milliseconds()})
		return n, failed, e
	}
	s.emit(Event{Action: "db.batch", Caller: caller, Env: sess.env, Kind: KindDB, OK: true, DurationMS: dur.Milliseconds()})
	return n, -1, nil
}

// SSHExec 在 SSH 会话上执行一条非交互命令。
//
// 非零退出码**不是错误**（那是一次正常的命令结果），只有会话/权限/超时才走 error。
func (s *Service) SSHExec(ctx context.Context, caller Caller, handle, cmd string, timeout time.Duration) (*ExecResult, *Error) {
	sess, lerr := s.lookup(caller, handle, KindSSH)
	if lerr != nil {
		return nil, lerr
	}
	if strings.TrimSpace(cmd) == "" {
		return nil, errf(CodeSSHExecFailed, sess.env, "命令为空")
	}
	if timeout <= 0 {
		timeout = DefaultSSHTimeout
	}
	if timeout > MaxSSHTimeout {
		timeout = MaxSSHTimeout
	}
	qctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := s.opt.Now()
	out, err := sess.ssh.OutputContext(qctx, cmd, timeout)
	dur := s.opt.Now().Sub(start)
	res := &ExecResult{Output: out}
	if err != nil {
		if qctx.Err() != nil || isTimeout(qctx, err) {
			res.TimedOut = true
			res.ExitCode = -1
			s.emit(Event{Action: "ssh.exec", Caller: caller, Env: sess.env, Kind: KindSSH, OK: false, Code: CodeTimeout, DurationMS: dur.Milliseconds()})
			return res, nil
		}
		res.ExitCode = host.ExitStatus(err)
		if res.ExitCode < 0 {
			e := &Error{Code: CodeSSHExecFailed, Env: sess.env, Stage: StageExec, Message: err.Error(), Cause: err}
			s.emit(Event{Action: "ssh.exec", Caller: caller, Env: sess.env, Kind: KindSSH, OK: false, Code: e.Code, DurationMS: dur.Milliseconds()})
			return nil, e
		}
		// 非零退出码：带着远端输出正常返回。
		s.emit(Event{Action: "ssh.exec", Caller: caller, Env: sess.env, Kind: KindSSH, OK: true, DurationMS: dur.Milliseconds()})
		return res, nil
	}
	s.emit(Event{Action: "ssh.exec", Caller: caller, Env: sess.env, Kind: KindSSH, OK: true, DurationMS: dur.Milliseconds()})
	return res, nil
}

// OpenBuiltin 打开一条**内置只读语义**的连接（文本层 + 库侧只读），含 viaSsh 隧道。
//
// 这是给第一方只读路径用的低层入口（字典直查、dbsync）：它把"起隧道 → 改写地址 →
// 连库 → 收口"这套编排收在一处，原来 internal/dict/live 与 internal/dict/dbsync
// 各写一遍的流程现在都走这里。返回的 release 必须 defer 调用（它同时关隧道）。
//
// 与 Service.OpenDB 的区别：这里不给句柄、不做插件授权 —— 调用方是 TT 自己的
// 只读链路，走的是既有两层防线（见 erpdb/session.go）。
func OpenBuiltin(ctx context.Context, conn dbconfig.Connection) (erpdb.Connector, func(), error) {
	tun, e := dialConn(&conn, defaultDialTunnel)
	if e != nil {
		return nil, nil, e
	}
	sess, err := erpdb.OpenSession(ctx, conn, erpdb.SessionOptions{ReadOnly: true})
	if err != nil {
		closeTunnel(tun)
		return nil, nil, err
	}
	return sess, func() {
		sess.Close()
		closeTunnel(tun)
	}, nil
}

// dialConn 为连接准备可达地址：配了 viaSsh 就先建隧道并把 host/port 改到本地端口。
//
// 这段编排以前在 internal/dict/live、internal/dict/dbsync 与 internal/web 各写一遍；
// 现在只有这一份，新增调用方不再复制。返回的 Tunnel 由调用方负责关闭。
// conn 会被**就地改写**（host/port 换成 127.0.0.1:<本地端口>），调用方拿到的
// `conn.Address()` 因此如实反映"实际连的是哪儿"。
func dialConn(conn *dbconfig.Connection, dial func(sshtun.Options) (Tunnel, error)) (Tunnel, *Error) {
	if conn.ViaSSH == nil {
		return nil, nil
	}
	remoteHost, remotePort := conn.ViaSSH.EffectiveRemote(conn.Host, conn.Port)
	tun, err := dial(sshtun.Options{
		Host:       conn.ViaSSH.Host,
		Port:       conn.ViaSSH.Port,
		User:       conn.ViaSSH.User,
		Password:   conn.ViaSSH.Password,
		RemoteHost: remoteHost,
		RemotePort: remotePort,
		LocalPort:  conn.ViaSSH.BindPort,
	})
	if err != nil {
		return nil, &Error{
			Code:    CodeDBConnectFailed,
			Stage:   StageTunnel,
			Message: fmt.Sprintf("建立 SSH 隧道失败（viaSsh → %s:%d）: %v", remoteHost, remotePort, err),
			Cause:   err,
		}
	}
	// 驱动改连本地转发端口；service/库名不变。
	conn.Host = "127.0.0.1"
	conn.Port = tun.LocalPort()
	return tun, nil
}

// dialDBConn 是 dialConn 的实例方法形态（用本服务注入的隧道实现）。
func (s *Service) dialDBConn(conn *dbconfig.Connection) (Tunnel, *Error) {
	return dialConn(conn, s.opt.DialTunnel)
}

// closeTunnel 关闭可选隧道。
func closeTunnel(t Tunnel) {
	if t != nil {
		_ = t.Close()
	}
}

// currentGeneration 读配置代次。
func (s *Service) currentGeneration() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.generation
}

// queryError 把查询错误归类成稳定码。
func queryError(env string, err error) *Error {
	code := CodeQueryFailed
	if errors.Is(err, context.DeadlineExceeded) {
		code = CodeTimeout
	}
	if errors.Is(err, erpdb.ErrReadOnlySession) {
		code = CodeSessionReadOnly
	}
	return &Error{Code: code, Env: env, Stage: StageQuery, Message: err.Error(), Cause: err}
}

// isTimeout 判断一次调用是否因超时收场（同时看 ctx 与错误链）。
func isTimeout(ctx context.Context, err error) bool {
	if ctx != nil && ctx.Err() != nil {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}

// withDefaultTimeout 给没有 deadline 的 ctx 套一个缺省超时。
//
// 有 deadline 的**不覆盖**：调用方给的更短，就该按它来。
func withDefaultTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, d)
}
