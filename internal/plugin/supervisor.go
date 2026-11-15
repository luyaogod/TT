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

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/hostrpc"
	"github.com/luyaogod/TT/internal/winproc"
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
	// ExePath 启动时那个服务入口的**绝对路径**：停止前核对身份的第一判据
	// （PID 会被复用；路径比摘要便宜，也能在摘要算不出来时仍然可判）。
	ExePath string `json:"exePath,omitempty"`
	// ExeDigest 可执行文件摘要：与 ExePath 一起核对身份（路径可能被同名的另一份覆盖）。
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
	// ExecutablePath 取某个 PID 现在的镜像路径（测试缝；nil = winproc.ExecutablePath）。
	//
	// 为什么要留缝：身份核对是"路径/摘要对不上就不杀"，而"PID 被复用"没法在测试里
	// 真造出来（得让系统刚好把同一个 PID 分给另一个程序）。把取路径这一步换成可控的，
	// 决策本身仍然测真的 shouldKill。
	ExecutablePath func(pid int) string

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

	// 起新实例之前先清掉**陈旧**的运行状态（P3-14）：上一轮 tt serve 可能崩溃了，
	// 留下的 instance.json 会让诊断（以及命令行升级的前置检查）以为服务还在跑。
	// 只清"进程不在 / PID 已被复用"的那些 —— 仍然有效的状态一个字节都不动。
	if old, reason, rerr := s.ReconcileInstanceState(rec); rerr != nil {
		s.logf("核对插件运行状态失败 id=%s: %v", rec.ID, rerr)
	} else if old != nil && reason != "" {
		s.logf("已清掉陈旧的插件运行状态 id=%s pid=%d：%s", rec.ID, old.PID, reason)
	}

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
		ExePath: exe, ExeDigest: digest,
		PID: cmd.Process.Pid, HealthPath: healthPath, StartedAt: s.now(),
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
	// 不论 stopProcess 走哪条分支（自己退了 / 被结束 / **身份不符而没动**），
	// 这条实例记录都不再代表"正在跑"——写 stopped，别让下一轮诊断把它当成活实例。
	inst.State = "stopped"
	s.writeInstanceState(inst)
	s.logf("插件服务已停止 id=%s pid=%d（身份不符时只撤销令牌，不动该进程）", pluginID, inst.PID)
	return nil
}

// stopProcess 身份核对 → 优雅收工 → 宽限期 → 结束进程树。
//
// 为什么第一步不是关 stdin/POST /shutdown，而是核对身份（P3-14）：PID 会被复用。
// 宽限期里我们起的那个进程可能已经退了，系统把同一个 PID 分给了别的程序；这时
// 一次 POST /shutdown 或 taskkill 就落到**别人的进程**上（端口、令牌都不是那个人的）。
// 所以任何停止动作之前都先确认"这个 PID 现在还是我们记录的可执行文件"。
func (s *Supervisor) stopProcess(inst *Instance) {
	if inst == nil || inst.cmd == nil || inst.cmd.Process == nil {
		return
	}
	if !s.sameInstance(inst) {
		s.logf("插件服务 PID 已被复用，按不杀处理 id=%s pid=%d（记录 %s，现在 %s）",
			inst.PluginID, inst.PID, inst.ExePath, s.executablePathOf(inst.PID))
		s.revokeToken(inst.token)
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
		// 宽限期内我们那个退了、PID 被别人接走：立刻停手，别等宽限期过完再误杀。
		if !s.sameInstance(inst) {
			s.logf("插件服务在停止过程中 PID 被复用，按不杀处理 id=%s pid=%d", inst.PluginID, inst.PID)
			s.revokeToken(inst.token)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 3) 最后一刻再核一次：宽限期与这里之间同样可能发生"它退了 + PID 被复用"。
	if !s.sameInstance(inst) {
		s.logf("插件服务结束前 PID 被复用，按不杀处理 id=%s pid=%d", inst.PluginID, inst.PID)
		s.revokeToken(inst.token)
		return
	}
	// 4) 还没走：结束整棵树（孙进程会继续占端口，只杀直接子进程不够）。
	if err := winproc.KillTree(inst.PID); err != nil {
		s.logf("结束插件服务进程树失败 id=%s pid=%d: %v", inst.PluginID, inst.PID, err)
	}
	s.revokeToken(inst.token)
}

// exeIdentity 一个进程的可执行身份：镜像路径 + 文件摘要。
//
// 两个都要，是因为它们能挡住不同的复用方式：路径挡住"PID 被另一个程序拿走"；
// 摘要挡住"路径还在、文件已经被换成另一份"（升级/手工替换）。
type exeIdentity struct {
	Path   string
	Digest string
}

// shouldKill 停止/结束前的**纯决策**：这个 PID 现在还是我们记录的那个实例吗？
//
// 返回 false = 什么都不要动（PID 被复用）。判据：
//   - 两边都有路径时必须一致（Windows 路径不区分大小写，用 EqualFold）；
//   - 两边都有摘要时必须一致；
//   - 只有一边有证据时按有证据的那边判：对不上就算复用；
//   - 什么证据都没有（非 Windows / 打不开进程）→ 放行。这里不能"因噎废食"：
//     调用方已经确认进程活着，拿不到身份不是"有罪"，否则会漏杀自己刚起的子进程。
func shouldKill(recorded, current exeIdentity) bool {
	if recorded.Path != "" && current.Path != "" && !sameExePath(recorded.Path, current.Path) {
		return false
	}
	if recorded.Digest != "" && current.Digest != "" && !strings.EqualFold(recorded.Digest, current.Digest) {
		return false
	}
	return true
}

// sameExePath 两个镜像路径是否指同一个文件（大小写不敏感，清理过分隔符）。
func sameExePath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

// identity 实例记录的可执行身份。
func (i *Instance) identity() exeIdentity {
	return exeIdentity{Path: i.ExePath, Digest: i.ExeDigest}
}

// identityAtPath 由镜像路径算身份；空路径 = 空身份（摘要算不出来就只留路径）。
func identityAtPath(path string) exeIdentity {
	if path == "" {
		return exeIdentity{}
	}
	id := exeIdentity{Path: path}
	if sum, err := fileSHA256(path); err == nil {
		id.Digest = sum
	}
	return id
}

// processIdentity 取 PID 现在的可执行身份（生产路径）。
func processIdentity(pid int) exeIdentity {
	return identityAtPath(winproc.ExecutablePath(pid))
}

// executablePathOf 取 PID 现在的镜像路径（日志/诊断用）。
func (s *Supervisor) executablePathOf(pid int) string {
	if s.ExecutablePath != nil {
		return s.ExecutablePath(pid)
	}
	return winproc.ExecutablePath(pid)
}

// currentIdentity 取 PID 现在的可执行身份（走测试缝）。
func (s *Supervisor) currentIdentity(pid int) exeIdentity {
	return identityAtPath(s.executablePathOf(pid))
}

// sameInstance 记录的实例与 PID 现在的身份是否对得上。
func (s *Supervisor) sameInstance(inst *Instance) bool {
	return shouldKill(inst.identity(), s.currentIdentity(inst.PID))
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

// InstanceState run/<id>/instance.json 的内容（supervisor 写，诊断/升级读）。
//
// 单独一个结构体而不是复用 Instance：Instance 带 cmd/token 这些内存态，
// 而这份文件是对外可读的事实（"上一轮是谁在跑"），字段要显式。
type InstanceState struct {
	PluginID   string    `json:"pluginId"`
	Version    string    `json:"version"`
	InstanceID string    `json:"instanceId"`
	PID        int       `json:"pid"`
	Addr       string    `json:"addr"`
	HealthPath string    `json:"healthPath,omitempty"`
	State      string    `json:"state"`
	StartedAt  time.Time `json:"startedAt"`
	Restarts   int       `json:"restarts,omitempty"`
	LogPath    string    `json:"logPath,omitempty"`
	ExePath    string    `json:"exePath,omitempty"`
	ExeDigest  string    `json:"exeDigest,omitempty"`
}

// Alive 状态文件说的那个进程还在不在。
//
// 只作**快速判据**：PID 会被复用，"活着"不等于"还是那一轮那个程序"。
// 要判"这份状态还算不算数"请用 StaleReason（它同时看可执行身份）。
func (i *InstanceState) Alive() bool {
	if i == nil || i.PID <= 0 {
		return false
	}
	return winproc.Alive(i.PID)
}

// StaleReason 这份运行状态是否已经不能代表"现在还在跑的那个实例"；空串 = 仍然成立。
//
// alive 由调用方给（判进程存活要碰系统，不属于纯决策）；current 是该 PID **现在**的
// 可执行身份。判据与停止前的 shouldKill 同源：对不上就是"PID 被复用"。
//
// 为什么需要它：tt serve 崩溃后留下的 instance.json 会让诊断（以及命令行升级的前置
// 检查）以为服务还在跑；只有"进程在 + 可执行身份对得上"才算数。
func (i *InstanceState) StaleReason(alive bool, current exeIdentity) string {
	if i == nil || i.PID <= 0 {
		return "状态文件没有有效的 PID"
	}
	if !alive {
		return fmt.Sprintf("PID %d 已经不在了", i.PID)
	}
	recorded := exeIdentity{Path: i.ExePath, Digest: i.ExeDigest}
	if !shouldKill(recorded, current) {
		return fmt.Sprintf("PID %d 现在指向另一个可执行文件（PID 被复用了）", i.PID)
	}
	return ""
}

// ClearInstanceState 删掉 run/<id>/instance.json（陈旧状态的清理）；不存在不算错。
func (s *Store) ClearInstanceState(id string) error {
	dir := s.RunDir(id)
	if dir == "" {
		return nil
	}
	err := os.Remove(filepath.Join(dir, "instance.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// ReconcileInstanceState 核对并清理 run/<id>/instance.json（P3-14）。
//
// 返回：
//   - (nil, "", nil)：没有状态文件，或读不出来（读失败由 err 报告）；
//   - (st, "", nil)：状态仍然有效，一个字节都没动；
//   - (st, reason, nil)：状态是陈旧的，**已经删掉**，reason 说明为什么。
//
// 只在 supervisor 要起新实例之前调用：此刻"上一轮是谁在跑"已经无关紧要，
// 留着陈旧状态只会误导诊断与升级前置检查。
func (s *Supervisor) ReconcileInstanceState(rec *Record) (*InstanceState, string, error) {
	if s.Store == nil || rec == nil {
		return nil, "", nil
	}
	st, err := s.Store.ReadInstanceState(rec.ID)
	if err != nil {
		return nil, "", err
	}
	if st == nil {
		return nil, "", nil
	}
	reason := st.StaleReason(winproc.Alive(st.PID), s.currentIdentity(st.PID))
	if reason == "" {
		return st, "", nil
	}
	if cerr := s.Store.ClearInstanceState(rec.ID); cerr != nil {
		return st, reason, fmt.Errorf("删除陈旧运行状态失败: %w", cerr)
	}
	return st, reason, nil
}

// ReadInstanceState 读 run/<id>/instance.json；文件不存在返回 (nil, nil)。
func (s *Store) ReadInstanceState(id string) (*InstanceState, error) {
	dir := s.RunDir(id)
	if dir == "" {
		return nil, nil
	}
	b, err := os.ReadFile(filepath.Join(dir, "instance.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var st InstanceState
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, fmt.Errorf("运行状态文件不是合法 JSON: %w", err)
	}
	return &st, nil
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
