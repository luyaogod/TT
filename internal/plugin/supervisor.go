package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tt/internal/connsvc"
	"tt/internal/hostrpc"
	"tt/internal/winproc"
)

// 插件常驻服务的生命周期契约（TT 侧）：
//
//	启动：TT 注入 TT_PLUGIN_LISTEN=127.0.0.1:0；
//	      插件在 stdout 打一行 {"ready":true,"addr":"127.0.0.1:<port>"}；
//	      TT 再对 <addr><healthPath> 做健康检查（默认 /healthz）。
//	停止：TT 先请它自己退（POST /shutdown，失败则关 stdin）；
//	      超过宽限期就结束整棵进程树。
//
// 为什么用"一行 JSON"而不是等端口：插件拿到的可能是 :0（内核选端口），
// 只有它自己知道选到了哪个；而"起来了吗"必须能区分"进程活着"与"在监听"。
const (
	// DefaultStartTimeout 等 ready 行的上限。
	DefaultStartTimeout = 10 * time.Second
	// DefaultHealthTimeout 等健康检查通过的上限。
	DefaultHealthTimeout = 10 * time.Second
	// DefaultStopGrace 优雅停止的宽限期，超时则杀进程树。
	DefaultStopGrace = 3 * time.Second
	// DefaultMaxRestarts 崩溃后最多重启几次（退避后仍失败就停在 degraded）。
	DefaultMaxRestarts = 3
)

// readyLine 插件就绪行。
type readyLine struct {
	Ready bool   `json:"ready"`
	Addr  string `json:"addr"`
}

// Instance 一个运行中的插件服务实例。
type Instance struct {
	PluginID   string    `json:"pluginId"`
	Version    string    `json:"version"`
	InstanceID string    `json:"instanceId"`
	PID        int       `json:"pid"`
	Addr       string    `json:"addr"`
	HealthPath string    `json:"healthPath"`
	StartedAt  time.Time `json:"startedAt"`
	State      string    `json:"state"` // running | degraded | stopped
	LogPath    string    `json:"logPath"`
	// ExeDigest 可执行文件摘要：停止前核对身份用（PID 会被复用）。
	ExeDigest string `json:"exeDigest,omitempty"`
	Restarts  int    `json:"restarts"`

	// token 本实例的 Host API 令牌；停止/重启即撤销。
	token string
	cmd   *exec.Cmd
	// logFile 服务日志句柄：进程活着期间必须开着（pumpStdout 一直在往里写）。
	logFile *os.File
	// stdin 用于"优雅停止"的一条通道：关掉它，按约定退出的插件会自己收工。
	stdin io.Closer

	stopping bool
}

// Supervisor 管理插件的常驻服务进程。
type Supervisor struct {
	Store  *Store
	Tokens TokenStoreIface
	// ConnSvc 连接服务（Host API 实现）；停服务/撤权时要按插件关会话。
	ConnSvc *connsvc.Service
	// HostAPIURL 交给插件的 Host API 地址（tt serve 的 /hostrpc/）。
	HostAPIURL string
	TTVersion  string
	// Logf 落 TT 自己的日志（不是插件的日志文件）。
	Logf func(format string, a ...any)
	// Now/HTTP 测试缝。
	Now           func() time.Time
	HTTP          *http.Client
	StartTimeout  time.Duration
	HealthTimeout time.Duration
	StopGrace     time.Duration
	MaxRestarts   int
	// ExtraEnv 追加到服务进程环境之后（测试用来打开夹具的可控故障开关）。
	ExtraEnv []string

	mu        sync.Mutex
	instances map[string]*Instance
	closed    bool
}

// TokenStoreIface 只用到令牌表的三个方法（便于测试替身）。
type TokenStoreIface interface {
	Issue(caller connsvc.Caller) string
	Revoke(token string) bool
	RevokePlugin(pluginID string) int
}

// 编不过就用真实的 hostrpc.TokenStore —— 这里用一个别名避免调用方多写一层。
// NewSupervisor 构造 supervisor。
func NewSupervisor(store *Store, tokens *hostrpc.TokenStore, connSvc *connsvc.Service, hostAPIURL, ttVersion string) *Supervisor {
	return &Supervisor{
		Store: store, Tokens: tokens, ConnSvc: connSvc,
		HostAPIURL: hostAPIURL, TTVersion: ttVersion,
		instances: map[string]*Instance{},
		Now:       time.Now,
	}
}

func (s *Supervisor) logf(format string, a ...any) {
	if s.Logf != nil {
		s.Logf(format, a...)
	}
}

func (s *Supervisor) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Supervisor) startTimeout() time.Duration {
	if s.StartTimeout > 0 {
		return s.StartTimeout
	}
	return DefaultStartTimeout
}

func (s *Supervisor) healthTimeout() time.Duration {
	if s.HealthTimeout > 0 {
		return s.HealthTimeout
	}
	return DefaultHealthTimeout
}

func (s *Supervisor) stopGrace() time.Duration {
	if s.StopGrace > 0 {
		return s.StopGrace
	}
	return DefaultStopGrace
}

func (s *Supervisor) maxRestarts() int {
	if s.MaxRestarts > 0 {
		return s.MaxRestarts
	}
	return DefaultMaxRestarts
}

// Start 启动一个插件的常驻服务（已在运行则直接返回现有实例）。
func (s *Supervisor) Start(ctx context.Context, rec *Record) (*Instance, error) {
	if rec == nil {
		return nil, errors.New("没有插件记录")
	}
	if rec.Contributions.Service == nil {
		return nil, fmt.Errorf("插件 %s 没有 service 贡献", rec.ID)
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, errors.New("supervisor 已关闭")
	}
	if inst, ok := s.instances[rec.ID]; ok && inst.State == "running" {
		s.mu.Unlock()
		return inst, nil
	}
	s.mu.Unlock()

	inst, err := s.startOnce(ctx, rec, 0)
	if err != nil {
		s.recordError(rec.ID, err.Error())
		return nil, err
	}
	s.mu.Lock()
	s.instances[rec.ID] = inst
	s.mu.Unlock()
	s.writeInstanceState(inst)
	s.logf("插件服务已启动 id=%s version=%s pid=%d addr=%s", inst.PluginID, inst.Version, inst.PID, inst.Addr)

	go s.monitor(rec, inst)
	return inst, nil
}

// startOnce 起一次进程并等到就绪。
func (s *Supervisor) startOnce(ctx context.Context, rec *Record, attempt int) (*Instance, error) {
	dir := s.Store.ActiveDir(rec)
	if dir == "" {
		return nil, fmt.Errorf("插件 %s 没有活跃版本", rec.ID)
	}
	man, err := LoadManifest(dir)
	if err != nil {
		return nil, err
	}
	rel, err := man.ResolveServiceEntry()
	if err != nil {
		return nil, err
	}
	exe := filepath.Join(dir, filepath.FromSlash(rel))
	if st, serr := os.Stat(exe); serr != nil || !st.Mode().IsRegular() {
		return nil, fmt.Errorf("服务入口不存在或不是普通文件: %s", exe)
	}
	healthPath := man.Contributions.Service.HealthPathOrDefault()

	instanceID, err := newInstanceID("svc")
	if err != nil {
		return nil, err
	}
	caller := connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: rec.ID, InstanceID: instanceID}
	token := s.issueToken(caller)

	logDir := s.Store.LogsDir(rec.ID)
	if logDir == "" {
		s.revokeToken(token)
		return nil, errors.New("无法定位插件日志目录")
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		s.revokeToken(token)
		return nil, fmt.Errorf("创建插件日志目录失败: %w", err)
	}
	logPath := filepath.Join(logDir, fmt.Sprintf("service-%s.log", s.now().Format("20060102-150405")))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		s.revokeToken(token)
		return nil, fmt.Errorf("打开插件日志失败: %w", err)
	}

	cmd := exec.Command(exe, man.Contributions.Service.Args...)
	cmd.Dir = dir // 服务的工作目录 = 它自己的版本目录
	cmd.Env = append(os.Environ(),
		"TT_HOST_API_URL="+s.HostAPIURL,
		"TT_PLUGIN_TOKEN="+token,
		"TT_PLUGIN_ID="+rec.ID,
		"TT_PLUGIN_VERSION="+rec.ActiveVersion,
		"TT_PLUGIN_INSTANCE_ID="+instanceID,
		"TT_PLUGIN_KIND="+string(connsvc.CallerServicePlugin),
		"TT_PLUGIN_DATA_DIR="+s.Store.DataDir(rec.ID),
		"TT_PLUGIN_LOGS_DIR="+logDir,
		"TT_PLUGIN_LISTEN=127.0.0.1:0",
	)
	cmd.Env = append(cmd.Env, s.ExtraEnv...)
	winproc.HideConsole(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logFile.Close()
		s.revokeToken(token)
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logFile.Close()
		s.revokeToken(token)
		return nil, err
	}
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		logFile.Close()
		s.revokeToken(token)
		return nil, fmt.Errorf("启动插件服务失败: %w", err)
	}

	digest := ""
	if sum, derr := fileSHA256(exe); derr == nil {
		digest = sum
	}

	inst := &Instance{
		PluginID: rec.ID, Version: rec.ActiveVersion, InstanceID: instanceID,
		ExeDigest: digest,
		PID:       cmd.Process.Pid, HealthPath: healthPath, StartedAt: s.now(),
		State: "running", LogPath: logPath, token: token, cmd: cmd, stdin: stdin,
		logFile: logFile, Restarts: attempt,
	}

	readyCh := make(chan readyLine, 1)
	go s.pumpStdout(stdout, logFile, readyCh)

	select {
	case r := <-readyCh:
		if !r.Ready || strings.TrimSpace(r.Addr) == "" {
			s.stopProcess(inst)
			logFile.Close()
			return nil, fmt.Errorf("插件服务没有给出有效的就绪地址（got %q）", r.Addr)
		}
		inst.Addr = r.Addr
	case <-time.After(s.startTimeout()):
		s.stopProcess(inst)
		logFile.Close()
		return nil, fmt.Errorf("插件服务在 %s 内没有就绪（看日志 %s）", s.startTimeout(), logPath)
	case <-ctx.Done():
		s.stopProcess(inst)
		logFile.Close()
		return nil, ctx.Err()
	}

	if err := s.waitHealthy(inst); err != nil {
		s.stopProcess(inst)
		logFile.Close()
		return nil, err
	}
	return inst, nil
}

// pumpStdout 把服务 stdout 落日志，并抓第一行就绪 JSON。
func (s *Supervisor) pumpStdout(out io.Reader, logFile io.Writer, readyCh chan<- readyLine) {
	sc := bufio.NewScanner(out)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	sent := false
	for sc.Scan() {
		line := sc.Text()
		fmt.Fprintln(logFile, line)
		if sent {
			continue
		}
		var r readyLine
		if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &r); err == nil && r.Ready {
			readyCh <- r
			sent = true
		}
	}
	if !sent {
		// stdout 关了但没等到就绪行：给等待方一个明确的失败信号（零值 Ready=false）。
		select {
		case readyCh <- readyLine{}:
		default:
		}
	}
}

// waitHealthy 轮询健康检查，直到通过或超时。
func (s *Supervisor) waitHealthy(inst *Instance) error {
	client := s.HTTP
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	url := "http://" + inst.Addr + inst.HealthPath
	deadline := s.now().Add(s.healthTimeout())
	var lastErr error
	for {
		resp, err := client.Get(url)
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			lastErr = fmt.Errorf("健康检查返回 %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		if s.now().After(deadline) {
			return fmt.Errorf("插件服务健康检查未通过（%s）：%v（日志 %s）", url, lastErr, inst.LogPath)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// monitor 看着服务进程：非预期退出时按有限次数退避重启。
func (s *Supervisor) monitor(rec *Record, inst *Instance) {
	err := inst.cmd.Wait()
	if inst.logFile != nil {
		_ = inst.logFile.Close()
	}
	s.mu.Lock()
	stopping := inst.stopping || s.closed
	s.mu.Unlock()
	if stopping {
		inst.State = "stopped"
		s.writeInstanceState(inst)
		return
	}

	s.logf("插件服务退出 id=%s pid=%d err=%v", inst.PluginID, inst.PID, err)
	s.revokeToken(inst.token)
	if s.ConnSvc != nil {
		s.ConnSvc.CloseForInstance(inst.PluginID, inst.InstanceID)
	}

	if inst.Restarts >= s.maxRestarts() {
		inst.State = "degraded"
		s.writeInstanceState(inst)
		s.logf("插件服务重启次数用尽 id=%s（进入 degraded，可用 tt plugin disable/enable 重试）", inst.PluginID)
		s.recordError(inst.PluginID, fmt.Sprintf("服务反复退出，已停止重启（最后错误: %v）", err))
		return
	}
	backoff := time.Duration(1<<uint(inst.Restarts)) * time.Second
	if backoff > 30*time.Second {
		backoff = 30 * time.Second
	}
	s.logf("插件服务将在 %s 后重启 id=%s（第 %d 次）", backoff, inst.PluginID, inst.Restarts+1)
	time.Sleep(backoff)

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	delete(s.instances, inst.PluginID)
	s.mu.Unlock()

	next, serr := s.startOnce(context.Background(), rec, inst.Restarts+1)
	if serr != nil {
		s.recordError(inst.PluginID, "重启失败: "+serr.Error())
		s.logf("插件服务重启失败 id=%s: %v", inst.PluginID, serr)
		inst.State = "degraded"
		s.writeInstanceState(inst)
		return
	}
	s.mu.Lock()
	s.instances[rec.ID] = next
	s.mu.Unlock()
	s.writeInstanceState(next)
	go s.monitor(rec, next)
}

// Stop 停止一个插件的服务（不存在时什么也不做）。
func (s *Supervisor) Stop(pluginID string) error {
	s.mu.Lock()
	inst, ok := s.instances[pluginID]
	if ok {
		inst.stopping = true
		delete(s.instances, pluginID)
	}
	s.mu.Unlock()
	if !ok {
		return nil
	}
	s.stopProcess(inst)
	s.writeInstanceState(inst)
	s.logf("插件服务已停止 id=%s pid=%d", pluginID, inst.PID)
	return nil
}

// stopProcess 优雅收工 → 宽限期 → 结束进程树。
func (s *Supervisor) stopProcess(inst *Instance) {
	if inst == nil || inst.cmd == nil || inst.cmd.Process == nil {
		return
	}
	// 1) 请它自己退：关 stdin（约定：服务监听 stdin EOF）+ POST /shutdown。
	if inst.stdin != nil {
		_ = inst.stdin.Close()
	}
	s.requestShutdown(inst)

	// 2) 宽限期等它自己走。
	deadline := s.now().Add(s.stopGrace())
	for s.now().Before(deadline) {
		if !winproc.Alive(inst.PID) {
			s.revokeToken(inst.token)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 3) 还没走：结束整棵树（孙进程会继续占端口，只杀直接子进程不够）。
	if err := winproc.KillTree(inst.PID); err != nil {
		s.logf("结束插件服务进程树失败 id=%s pid=%d: %v", inst.PluginID, inst.PID, err)
	}
	s.revokeToken(inst.token)
}

// requestShutdown 尽力请求优雅停止；插件不支持就算了（宽限期后照样收）。
func (s *Supervisor) requestShutdown(inst *Instance) {
	if inst.Addr == "" {
		return
	}
	client := s.HTTP
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	req, err := http.NewRequest(http.MethodPost, "http://"+inst.Addr+"/shutdown", nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+inst.token)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// StopAll 停止全部服务（TT 退出/配置变更时）。
func (s *Supervisor) StopAll() {
	s.mu.Lock()
	s.closed = true
	ids := make([]string, 0, len(s.instances))
	for id := range s.instances {
		ids = append(ids, id)
	}
	s.mu.Unlock()
	for _, id := range ids {
		_ = s.Stop(id)
	}
}

// Instance 取一个插件的运行实例。
func (s *Supervisor) Instance(pluginID string) *Instance {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.instances[pluginID]
}

// Addr 取服务地址（插件私有 RPC 代理要用）。
func (s *Supervisor) Addr(pluginID string) (string, bool) {
	if inst := s.Instance(pluginID); inst != nil && inst.Addr != "" {
		return inst.Addr, true
	}
	return "", false
}

// Status 运行态的对外投影（设置页/诊断用）。
type Status struct {
	PluginID   string `json:"pluginId"`
	State      string `json:"state"`
	Version    string `json:"version"`
	InstanceID string `json:"instanceId,omitempty"`
	PID        int    `json:"pid,omitempty"`
	Addr       string `json:"addr,omitempty"`
	StartedAt  string `json:"startedAt,omitempty"`
	Restarts   int    `json:"restarts,omitempty"`
	LogPath    string `json:"logPath,omitempty"`
}

// Statuses 全部实例状态。
func (s *Supervisor) Statuses() []Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Status, 0, len(s.instances))
	for _, inst := range s.instances {
		out = append(out, Status{
			PluginID: inst.PluginID, State: inst.State, Version: inst.Version,
			InstanceID: inst.InstanceID, PID: inst.PID, Addr: inst.Addr,
			StartedAt: inst.StartedAt.Format(time.RFC3339), Restarts: inst.Restarts,
			LogPath: inst.LogPath,
		})
	}
	return out
}

// writeInstanceState 把实例状态写到 run/<id>/instance.json（重启后能看出上一轮是谁）。
func (s *Supervisor) writeInstanceState(inst *Instance) {
	if inst == nil {
		return
	}
	dir := s.Store.RunDir(inst.PluginID)
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	b, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		return
	}
	path := filepath.Join(dir, "instance.json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err == nil {
		_ = os.Rename(tmp, path)
	}
}

// issueToken 签发实例令牌。
func (s *Supervisor) issueToken(caller connsvc.Caller) string {
	if s.Tokens == nil {
		return ""
	}
	return s.Tokens.Issue(caller)
}

// revokeToken 撤销实例令牌。
func (s *Supervisor) revokeToken(token string) {
	if s.Tokens == nil || token == "" {
		return
	}
	s.Tokens.Revoke(token)
}

// recordError 把失败原因写进注册表（展示与诊断用，不改启用状态）。
func (s *Supervisor) recordError(pluginID, msg string) {
	if s.Store == nil {
		return
	}
	_ = s.Store.WithLock(func(reg *Registry) error {
		if r := reg.Find(pluginID); r != nil {
			r.LastError = msg
		}
		return nil
	})
}

// instancePort 取实例地址里的端口（诊断/展示用）。
func instancePort(addr string) int {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return 0
	}
	var p int
	fmt.Sscanf(port, "%d", &p)
	return p
}

// Endpoint 取某插件服务的本机地址与实例令牌（插件私有 RPC 代理用）。
//
// 令牌一起给：代理是**代表插件**去调它自己的服务，不是以宿主身份去调；
// 插件服务可以据此拒绝来路不明的请求。
func (s *Supervisor) Endpoint(pluginID string) (addr, token string, ok bool) {
	inst := s.Instance(pluginID)
	if inst == nil || inst.Addr == "" || inst.State != "running" {
		return "", "", false
	}
	return inst.Addr, inst.token, true
}
