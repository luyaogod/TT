// Context 的运行时实现：把 sdk 的服务接口接到 Host API 上。
//
// 原则：凭据只在 TT 手里；插件拿到的是句柄。这里的每个方法只做"请求-映射-错误翻译"。

package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"connectrpc.com/connect"

	hostv1 "github.com/luyaogod/TT/pkg/hostv1"
)

const queryMaxRows = 50000

// contextImpl 一次插件运行期的上下文。
type contextImpl struct {
	ctx     context.Context
	cli     *hostClient
	id      string
	version string
	dataDir string
}

func newContext(ctx context.Context, baseURL, token, id, version, dataDir string) *contextImpl {
	return &contextImpl{ctx: ctx, cli: newHostClient(baseURL, token), id: id, version: version, dataDir: dataDir}
}

func (c *contextImpl) Context() context.Context { return c.ctx }
func (c *contextImpl) PluginID() string         { return c.id }
func (c *contextImpl) Version() string          { return c.version }
func (c *contextImpl) DataDir() string          { return c.dataDir }

func (c *contextImpl) Config() ConfigService { return &configService{c: c} }
func (c *contextImpl) DB() DBService         { return &dbService{c: c} }
func (c *contextImpl) SSH() SSHService       { return &sshService{c: c} }
func (c *contextImpl) FS() FSService         { return &fsService{c: c} }
func (c *contextImpl) Log() Logger           { return &logger{id: c.id} }

// Jobs / Commands / Events 还没接入：返回 nil，插件必须判空后再用。
func (c *contextImpl) Jobs() JobService         { return nil }
func (c *contextImpl) Commands() CommandService { return nil }
func (c *contextImpl) Events() EventService     { return nil }

func errUnimplemented(what string) error {
	return &Error{Code: CodeUnimplemented, Message: what + "尚未接入"}
}

// ---- 配置 ----

type configService struct{ c *contextImpl }

func (s *configService) Envs() []Env {
	resp, err := s.c.cli.profiles.List(s.c.ctx, connect.NewRequest(&hostv1.ListRequest{}))
	if err != nil {
		s.c.Log().Error("读取环境清单失败: %v", mapErr(err))
		return nil
	}
	out := make([]Env, 0, len(resp.Msg.GetProfiles()))
	for _, p := range resp.Msg.GetProfiles() {
		out = append(out, Env{Name: p.GetEnv(), DBType: p.GetDbType(), DBAddr: p.GetDbAddress()})
	}
	return out
}

func (s *configService) Ents(env string) []Ent {
	resp, err := s.c.cli.profiles.List(s.c.ctx, connect.NewRequest(&hostv1.ListRequest{}))
	if err != nil {
		s.c.Log().Error("读取企业清单失败: %v", mapErr(err))
		return nil
	}
	for _, p := range resp.Msg.GetProfiles() {
		if p.GetEnv() != env {
			continue
		}
		out := make([]Ent, 0, len(p.GetEnts()))
		for _, e := range p.GetEnts() {
			out = append(out, Ent{Ent: int(e.GetEnt()), Account: e.GetAccount(), Placeholder: e.GetPlaceholder()})
		}
		return out
	}
	return nil
}

// Values 插件自己的配置（当前由插件自己从文件读；Config 段注入随 P4 落地）。
func (s *configService) Values() json.RawMessage { return nil }

// ---- 数据库 ----

type dbService struct{ c *contextImpl }

func (d *dbService) Open(ctx context.Context, env string, opt OpenOptions) (DBSession, error) {
	resp, err := d.c.cli.conns.OpenDB(ctx, connect.NewRequest(&hostv1.OpenDBRequest{
		Env: env, ReadOnly: opt.ReadOnly, Account: opt.Account, Ent: int32(opt.Ent),
	}))
	if err != nil {
		return nil, mapErr(err)
	}
	return &dbSession{cli: d.c.cli, handle: resp.Msg.GetHandle()}, nil
}

type dbSession struct {
	cli    *hostClient
	handle string
}

func (s *dbSession) Query(ctx context.Context, sql string, params ...string) (Rows, error) {
	resp, err := s.cli.db.Query(ctx, connect.NewRequest(&hostv1.QueryRequest{
		Handle: s.handle, Sql: sql, Params: params, MaxRows: queryMaxRows,
	}))
	if err != nil {
		return Rows{}, mapErr(err)
	}
	out := Rows{Columns: resp.Msg.GetColumns(), Truncated: resp.Msg.GetTruncated()}
	for _, r := range resp.Msg.GetRows() {
		out.Values = append(out.Values, r.GetValues())
	}
	return out, nil
}

func (s *dbSession) Exec(ctx context.Context, sql string, params ...string) (int64, error) {
	resp, err := s.cli.db.Exec(ctx, connect.NewRequest(&hostv1.ExecRequest{
		Handle: s.handle, Sql: sql, Params: params,
	}))
	if err != nil {
		return 0, mapErr(err)
	}
	return resp.Msg.GetRowsAffected(), nil
}

func (s *dbSession) Batch(ctx context.Context, stmts []Statement, atomic bool) (BatchResult, error) {
	req := &hostv1.BatchRequest{Handle: s.handle, Atomic: atomic}
	for _, st := range stmts {
		req.Statements = append(req.Statements, &hostv1.Statement{Sql: st.SQL, Params: st.Params})
	}
	resp, err := s.cli.db.Batch(ctx, connect.NewRequest(req))
	if err != nil {
		return BatchResult{FailedIndex: -1}, mapErr(err)
	}
	return BatchResult{RowsAffected: resp.Msg.GetRowsAffected(), FailedIndex: int(resp.Msg.GetFailedIndex())}, nil
}

func (s *dbSession) Begin(context.Context) (Tx, error) {
	return nil, errUnimplemented("显式事务")
}

func (s *dbSession) Close() error {
	_, err := s.cli.conns.Close(context.Background(), connect.NewRequest(&hostv1.CloseRequest{Handle: s.handle}))
	return mapErr(err)
}

// ---- SSH ----

type sshService struct{ c *contextImpl }

func (s *sshService) Open(ctx context.Context, env string) (SSHSession, error) {
	resp, err := s.c.cli.conns.OpenSSH(ctx, connect.NewRequest(&hostv1.OpenSSHRequest{Env: env}))
	if err != nil {
		return nil, mapErr(err)
	}
	return &sshSession{cli: s.c.cli, handle: resp.Msg.GetHandle()}, nil
}

type sshSession struct {
	cli    *hostClient
	handle string
}

func (s *sshSession) Exec(ctx context.Context, command string, timeout time.Duration) (SSHOutput, error) {
	resp, err := s.cli.ssh.Exec(ctx, connect.NewRequest(&hostv1.SSHExecRequest{
		Handle: s.handle, Command: command, TimeoutMs: int32(timeout.Milliseconds()),
	}))
	if err != nil {
		return SSHOutput{}, mapErr(err)
	}
	return SSHOutput{
		Output:   resp.Msg.GetOutput(),
		ExitCode: int(resp.Msg.GetExitCode()),
		TimedOut: resp.Msg.GetTimedOut(),
	}, nil
}

func (s *sshSession) Put(ctx context.Context, localPath, remotePath string) error {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return &Error{Code: CodeInvalidArgument, Message: "读取本地文件失败: " + err.Error()}
	}
	_, err = s.cli.ssh.PutFile(ctx, connect.NewRequest(&hostv1.PutFileRequest{
		Handle: s.handle, Path: remotePath, Data: data,
	}))
	return mapErr(err)
}

func (s *sshSession) Get(ctx context.Context, remotePath, localPath string) error {
	resp, err := s.cli.ssh.GetFile(ctx, connect.NewRequest(&hostv1.GetFileRequest{
		Handle: s.handle, Path: remotePath,
	}))
	if err != nil {
		return mapErr(err)
	}
	if dir := filepath.Dir(localPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return &Error{Code: CodeInternal, Message: "创建本地目录失败: " + err.Error()}
		}
	}
	if err := os.WriteFile(localPath, resp.Msg.GetData(), 0o644); err != nil {
		return &Error{Code: CodeInternal, Message: "写入本地文件失败: " + err.Error()}
	}
	return nil
}

func (s *sshSession) Close() error {
	_, err := s.cli.conns.Close(context.Background(), connect.NewRequest(&hostv1.CloseRequest{Handle: s.handle}))
	return mapErr(err)
}

// ---- 文件 ----

type fsService struct{ c *contextImpl }

func (s *fsService) DataDir() string { return s.c.dataDir }

func (s *fsService) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(s.resolve(path))
}

func (s *fsService) WriteFile(path string, data []byte) error {
	full := s.resolve(path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

// resolve 相对路径按插件数据目录解析；绝对路径原样用（插件自己的选择）。
func (s *fsService) resolve(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(s.c.dataDir, path)
}

// ---- 日志 ----

type logger struct{ id string }

func (l *logger) Info(format string, args ...any)  { l.log("INFO", format, args...) }
func (l *logger) Warn(format string, args ...any)  { l.log("WARN", format, args...) }
func (l *logger) Error(format string, args ...any) { l.log("ERROR", format, args...) }

func (l *logger) log(level, format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[%s] %s %s\n", l.id, level, fmt.Sprintf(format, args...))
}
