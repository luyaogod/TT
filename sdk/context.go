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
	"sync"
	"sync/atomic"
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
	jobs    *jobService
	events  *eventService
	cmds    *commandService
}

func newContext(ctx context.Context, baseURL, token, id, version, dataDir string) *contextImpl {
	c := &contextImpl{ctx: ctx, cli: newHostClient(baseURL, token), id: id, version: version, dataDir: dataDir}
	c.jobs = &jobService{c: c, runs: map[string]*jobRun{}}
	c.events = &eventService{c: c, handlers: map[string][]func(json.RawMessage){}}
	c.cmds = &commandService{c: c, defs: map[string]cmdDef{}}
	return c
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

// Jobs 是进程内实现（任务属于插件自己的业务，不需要过 Host API）。
func (c *contextImpl) Jobs() JobService { return c.jobs }

// Events 轮询 Host API 的事件环（配置变更/插件重载）。
func (c *contextImpl) Events() EventService { return c.events }

// Commands 进程内子命令注册表（宿主不动态改 CLI 帮助；插件在 main 里用 Dispatch 分发）。
func (c *contextImpl) Commands() CommandService { return c.cmds }

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

// ---- 任务 ----

type jobRun struct {
	mu      sync.Mutex
	percent int
	message string
	err     error
	done    bool
}

func (j *jobRun) set(p int, msg string) {
	j.mu.Lock()
	j.percent, j.message = p, msg
	j.mu.Unlock()
}

func (j *jobRun) finish(err error) {
	j.mu.Lock()
	j.err, j.done = err, true
	j.mu.Unlock()
}

// jobService 进程内任务：进度与日志直接走插件日志，取消靠 context。
type jobService struct {
	c    *contextImpl
	mu   sync.Mutex
	seq  int64
	runs map[string]*jobRun
}

// Run 起一个后台任务，立即返回任务 id；fn 在独立 goroutine 里执行。
func (s *jobService) Run(parent context.Context, name string, fn func(context.Context, Reporter) error) (string, error) {
	if fn == nil {
		return "", &Error{Code: CodeInvalidArgument, Message: "任务函数为空"}
	}
	if parent == nil {
		parent = s.c.ctx
	}
	id := fmt.Sprintf("job-%d", atomic.AddInt64(&s.seq, 1))
	jr := &jobRun{}
	s.mu.Lock()
	s.runs[id] = jr
	s.mu.Unlock()

	ctx, cancel := context.WithCancel(parent)
	s.c.Log().Info("job %s 开始: %s", id, name)
	go func() {
		defer cancel()
		err := fn(ctx, &jobReporter{c: s.c, id: id, jr: jr})
		jr.finish(err)
		if err != nil {
			s.c.Log().Error("job %s 失败: %v", id, err)
			return
		}
		s.c.Log().Info("job %s 完成", id)
	}()
	return id, nil
}

type jobReporter struct {
	c  *contextImpl
	id string
	jr *jobRun
}

func (r *jobReporter) Progress(percent int, message string) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	r.jr.set(percent, message)
	r.c.Log().Info("job %s %d%% %s", r.id, percent, message)
}

func (r *jobReporter) Log(message string) { r.c.Log().Info("job %s %s", r.id, message) }

// ---- 事件 ----

// eventService 首次订阅时启动一个轮询 goroutine（2s 间隔，失败不退出）。
type eventService struct {
	c        *contextImpl
	mu       sync.Mutex
	handlers map[string][]func(json.RawMessage)
	started  bool
}

func (s *eventService) On(name string, handler func(payload json.RawMessage)) {
	if handler == nil || name == "" {
		return
	}
	s.mu.Lock()
	s.handlers[name] = append(s.handlers[name], handler)
	first := !s.started
	s.started = true
	s.mu.Unlock()
	if first {
		go s.loop()
	}
}

func (s *eventService) loop() {
	var after int64 // 0 = 从当前最新开始
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.c.ctx.Done():
			return
		case <-ticker.C:
		}
		resp, err := s.c.cli.events.Poll(s.c.ctx, connect.NewRequest(&hostv1.PollRequest{AfterId: after}))
		if err != nil {
			continue
		}
		for _, ev := range resp.Msg.GetEvents() {
			after = ev.GetId()
			s.dispatch(ev.GetName(), ev.GetPayload())
		}
	}
}

func (s *eventService) dispatch(name string, payload json.RawMessage) {
	s.mu.Lock()
	handlers := append([]func(json.RawMessage){}, s.handlers[name]...)
	handlers = append(handlers, s.handlers["*"]...)
	s.mu.Unlock()
	for _, h := range handlers {
		h(payload)
	}
}

// ---- 子命令 ----

type cmdDef struct {
	summary string
	handler func(context.Context, []string) error
}

type commandService struct {
	c    *contextImpl
	mu   sync.Mutex
	defs map[string]cmdDef
}

func (s *commandService) Register(name, summary string, handler func(context.Context, []string) error) {
	if name == "" || handler == nil {
		return
	}
	s.mu.Lock()
	s.defs[name] = cmdDef{summary: summary, handler: handler}
	s.mu.Unlock()
}

// Dispatch 按名字执行一个已注册子命令；args 是命令名之后的参数。
//
// 它给"CLI 转发进来后想自己再分子命令"的插件用：插件 main 里
// `sdk.Dispatch(ctx, os.Args[1:])` 即可，不必自己写 switch。
func Dispatch(ctx Context, args []string) error {
	if len(args) == 0 {
		return &Error{Code: CodeInvalidArgument, Message: "缺少子命令"}
	}
	cs, ok := ctx.Commands().(*commandService)
	if !ok || cs == nil {
		return &Error{Code: CodeUnimplemented, Message: "子命令服务不可用"}
	}
	cs.mu.Lock()
	def, ok := cs.defs[args[0]]
	cs.mu.Unlock()
	if !ok {
		return &Error{Code: CodeInvalidArgument, Message: "未知子命令: " + args[0]}
	}
	return def.handler(ctx.Context(), args[1:])
}
