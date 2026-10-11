package fep

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"tt/internal/host"
	"tt/internal/winproc"
)

// Defaults measured on a live job (GDC 3.20.25).
const (
	DefaultGDCPort    = 7400
	DefaultProxyPort  = 7600
	DefaultInjectPort = 7899

	// FGLSERVEROffset is the fixed offset between the port the job client dials
	// and the port named in FGLSERVER: `FGLSERVER=<ip>:1200` makes the client
	// connect to `<ip>:7600` (measured). So the proxy must listen at
	// FGLSERVER's number plus this offset.
	FGLSERVEROffset = 6400

	// gdcWarmup is extra grace after the GDC port accepts connections, before
	// anything is assumed about the window. Taking a frame mid-render produces a
	// half-drawn screen that looks like a real one.
	gdcWarmup = 700 * time.Millisecond
)

// Pre-flight failures that a caller must map to their own exit code:
// ErrPreflight -> environment check failed, ErrPortBusy -> session state.
var (
	ErrPreflight = errors.New("fep: 环境体检不过")
	ErrPortBusy  = errors.New("fep: 端口被占")
	// ErrUsage：参数组合不成立（退出码 2）。
	ErrUsage = errors.New("fep: 用法错")
)

// OpenConfig describes one isolated session.
type OpenConfig struct {
	Env  string // environment name (recorded, for the operator's benefit)
	Zone string
	Prog string // job program, for example axmt500_wf

	SSH host.SSHConfig

	GDCExe      string
	GDCPort     int
	ProxyPort   int
	InjectPort  int
	FGLServerIP string // empty = detect from the route to the SSH host

	DataDir   string // where the session registry lives
	RecordDir string // where the per-connection recordings land
	Tag       string // recording prefix
	GDCLog    string // the isolated instance's output

	WaitGDC time.Duration // how long to wait for the GDC port (default 20s)
	NoJob   bool          // start instance + proxy only (troubleshooting)
}

func (c *OpenConfig) fillDefaults() {
	if c.GDCExe == "" {
		c.GDCExe = `C:\Program Files\FourJs\gdc\bin\gdc.exe`
	}
	if c.GDCPort == 0 {
		c.GDCPort = DefaultGDCPort
	}
	if c.ProxyPort == 0 {
		c.ProxyPort = DefaultProxyPort
	}
	if c.InjectPort == 0 {
		c.InjectPort = DefaultInjectPort
	}
	if c.WaitGDC == 0 {
		c.WaitGDC = 20 * time.Second
	}
	if c.Tag == "" {
		c.Tag = "fep-" + strconv.FormatInt(time.Now().Unix(), 36)
	}
}

// Runtime is a live session. It lives in one resident process — the proxy must
// not die with a one-shot command — and everybody else addresses it through the
// registry and the control endpoint.
type Runtime struct {
	Session *Session

	cfg   OpenConfig
	reg   *Registry
	proxy *Proxy
	conn  *host.SSHConn
	pty   *host.PTYSession
	logf  func(string, ...any)

	jobDone chan struct{}
	closeMu sync.Mutex
	closed  bool
}

func nopLog(string, ...any) {}

// Open runs the whole start-up sequence, registering the session *before*
// anything is opened so an interrupted run can always be collected.
func Open(ctx context.Context, cfg OpenConfig, logf func(string, ...any)) (*Runtime, error) {
	cfg.fillDefaults()
	if logf == nil {
		logf = nopLog
	}
	if cfg.Prog == "" && !cfg.NoJob {
		return nil, errors.New("fep: 没给作业程序（--prog）")
	}
	if cfg.DataDir == "" {
		return nil, errors.New("fep: 没有数据目录（配置路径没解析出来）")
	}
	if err := preflightElevation(cfg.GDCExe); err != nil {
		return nil, err
	}
	for _, p := range []int{cfg.GDCPort, cfg.ProxyPort, cfg.InjectPort} {
		if err := portFree(p); err != nil {
			return nil, err
		}
	}

	r := &Runtime{cfg: cfg, reg: NewRegistry(cfg.DataDir), logf: logf, jobDone: make(chan struct{})}

	// 1. isolated instance (detached: it must outlive a crashed TT)
	pid, err := winproc.SpawnDetached(cfg.GDCExe,
		[]string{"-n", "-M", "-A", "0", "-p", strconv.Itoa(cfg.GDCPort)}, cfg.GDCLog)
	if err != nil {
		return nil, fmt.Errorf("fep: 起隔离实例失败: %w", err)
	}
	logf("GDC 已起：pid %d，端口 %d", pid, cfg.GDCPort)
	if err := waitPort(cfg.GDCPort, cfg.WaitGDC); err != nil {
		winproc.Kill(pid)
		return nil, err
	}
	time.Sleep(gdcWarmup)

	// 2. proxy (records every connection before anything is injected)
	p := NewProxy(ProxyConfig{
		Listen:  fmt.Sprintf("0.0.0.0:%d", cfg.ProxyPort),
		Target:  net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.GDCPort)),
		LogDir:  cfg.RecordDir,
		Tag:     cfg.Tag,
		Control: net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.InjectPort)),
	})
	if err := p.Start(); err != nil {
		winproc.Kill(pid)
		return nil, err
	}
	r.proxy = p

	// 3. register first, clean up later
	sess := &Session{
		Env: cfg.Env, Zone: cfg.Zone, Prog: cfg.Prog,
		GDCPID: pid, GDCPort: cfg.GDCPort,
		ProxyPort: cfg.ProxyPort, InjectPort: cfg.InjectPort,
		Tag: cfg.Tag, RecordDir: cfg.RecordDir, OwnerPID: os.Getpid(), Started: time.Now(),
	}
	if err := r.reg.Save(sess); err != nil {
		p.Close()
		winproc.Kill(pid)
		return nil, err
	}
	r.Session = sess
	logf("会话已登记：%s", r.reg.Path)

	// 4. open the job over SSH (the job client dials back into our proxy)
	if !cfg.NoJob {
		ip := cfg.FGLServerIP
		if ip == "" {
			ip, err = localAddrFor(cfg.SSH)
			if err != nil {
				r.Close()
				return nil, err
			}
			logf("FGLSERVER 用本机地址 %s（按到 %s 的路由探测）", ip, cfg.SSH.Host)
		}
		if err := r.openJob(ip); err != nil {
			r.Close()
			return nil, err
		}
		logf("作业已打开：%s，FGLSERVER=%s:%d", cfg.Prog, ip, cfg.ProxyPort-FGLSERVEROffset)
	}
	return r, nil
}

// Wait blocks until the session ends: the caller's context is cancelled, a
// control client sent `quit` (`tt fep session close`), or the remote job exited.
func (r *Runtime) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case <-r.proxy.Quit():
		return nil
	case <-r.jobDone:
		return nil
	}
}

// Close stops everything this session started, in the order that leaves nothing
// behind: remote job, proxy, isolated instance, registry.
func (r *Runtime) Close() error {
	r.closeMu.Lock()
	if r.closed {
		r.closeMu.Unlock()
		return nil
	}
	r.closed = true
	r.closeMu.Unlock()

	var errs []string
	if r.pty != nil {
		r.pty.Close()
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if r.proxy != nil {
		r.proxy.Close()
	}
	// Only ever the PIDs this session registered.
	for _, pid := range r.Session.PIDs() {
		if winproc.Alive(pid) {
			if err := winproc.Kill(pid); err != nil {
				errs = append(errs, err.Error())
			} else {
				r.logf("已收隔离实例 pid %d", pid)
			}
		}
	}
	if err := r.reg.Clear(); err != nil {
		errs = append(errs, err.Error())
	} else {
		r.logf("登记已清：%s", r.reg.Path)
	}
	if len(errs) > 0 {
		return fmt.Errorf("fep: 收尾有失败: %s", strings.Join(errs, "; "))
	}
	return nil
}

// openJob logs in, points FGLSERVER at our proxy and starts the job client.
func (r *Runtime) openJob(ip string) error {
	conn, err := host.Dial(r.cfg.SSH)
	if err != nil {
		return fmt.Errorf("fep: SSH 连接失败: %w", err)
	}
	r.conn = conn
	pty, err := conn.NewPTY(120, 40)
	if err != nil {
		return fmt.Errorf("fep: 开 PTY 失败: %w", err)
	}
	r.pty = pty

	if err := waitRegexp(pty, host.ReLoginMenu, 25*time.Second, "区域菜单"); err != nil {
		return err
	}
	if r.cfg.Zone == "" {
		return errors.New("fep: 这个环境没配 zone（区域菜单没法选）")
	}
	if err := pty.Write(r.cfg.Zone + "\r"); err != nil {
		return err
	}
	if err := waitRegexp(pty, host.ReShellPrompt, 25*time.Second, "shell 提示符"); err != nil {
		return err
	}

	port := r.cfg.ProxyPort - FGLSERVEROffset
	if port <= 0 || port > 65535 {
		return fmt.Errorf("%w: 代理端口 %d 减 %d 后不是合法端口", ErrUsage, r.cfg.ProxyPort, FGLSERVEROffset)
	}
	cmd := fmt.Sprintf("export FGLSERVER=%s:%d\r\nr.r %s\r\n", ip, port, r.cfg.Prog)
	if err := pty.Write(cmd); err != nil {
		return fmt.Errorf("fep: 下发作业启动命令失败: %w", err)
	}
	// The job client keeps running while the GUI is open; its connection ending
	// means the job is gone.
	go r.watchJob()
	return nil
}

// watchJob drains the PTY and reports when the remote job is over.
func (r *Runtime) watchJob() {
	defer close(r.jobDone)
	buf := make([]byte, 4096)
	for {
		n, err := r.pty.Read(buf)
		if n > 0 {
			line := strings.TrimSpace(string(buf[:n]))
			if line != "" {
				r.logf("远端: %s", firstLine(line))
			}
		}
		if err != nil {
			return
		}
	}
}

func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	if len(s) > 160 {
		s = s[:160] + "…"
	}
	return s
}

// waitRegexp waits for a pattern in the PTY stream.
func waitRegexp(pty *host.PTYSession, re *regexp.Regexp, timeout time.Duration, what string) error {
	type res struct {
		b   []byte
		err error
	}
	ch := make(chan res, 1)
	go func() {
		buf := make([]byte, 4096)
		var acc []byte
		for {
			n, err := pty.Read(buf)
			if n > 0 {
				acc = append(acc, buf[:n]...)
				if re.Match(acc) {
					ch <- res{}
					return
				}
				if len(acc) > 1<<20 {
					acc = acc[len(acc)-1<<16:]
				}
			}
			if err != nil {
				ch <- res{err: err}
				return
			}
		}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			return fmt.Errorf("fep: 等%s时连接断了: %w", what, r.err)
		}
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("fep: 等%s超时（%s）", what, timeout)
	}
	// 注意：超时后那个读 goroutine 仍在读 —— PTY 只能有一个读者，
	// 所以调用方在这一步失败时必须立刻 Close（Open 的失败路径都做了）。
}

// preflightElevation refuses to start when an already-running instance has
// different privileges than we do: that instance is the one the job would talk
// to, and it cannot be driven reliably from here.
func preflightElevation(gdcExe string) error {
	pids, err := winproc.ListProcesses(filepath.Base(gdcExe))
	if err != nil {
		return nil // 非 Windows：没有这层问题
	}
	for _, pid := range pids {
		e, err := CheckElevation(pid)
		if err != nil {
			continue
		}
		if !e.OK {
			return fmt.Errorf("%w（pid %d）: %s", ErrPreflight, pid, e.Advice)
		}
	}
	return nil
}

func portFree(port int) error {
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return fmt.Errorf("%w: %d 已被占用（%v）", ErrPortBusy, port, err)
	}
	ln.Close()
	return nil
}

func waitPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("%w: 隔离实例在 %s 内没在 %d 端口监听", ErrPortBusy, timeout, port)
}

// localAddrFor reports the local address the server would see for our SSH
// connection. A UDP "connection" sends nothing — it only asks the kernel which
// source address the route picks, which is what the job client on the server
// side will dial back to.
func localAddrFor(ssh host.SSHConfig) (string, error) {
	port := ssh.Port
	if port == 0 {
		port = 22
	}
	c, err := net.Dial("udp", net.JoinHostPort(ssh.Host, strconv.Itoa(port)))
	if err != nil {
		return "", fmt.Errorf("fep: 探测本机地址失败: %w", err)
	}
	defer c.Close()
	if la, ok := c.LocalAddr().(*net.UDPAddr); ok && la.IP != nil && !la.IP.IsUnspecified() {
		return la.IP.String(), nil
	}
	return "", errors.New("fep: 探测本机地址失败（路由没给出源地址）")
}

// RegistryPath reports the session registry file (for diagnostics).
func (r *Runtime) RegistryPath() string { return r.reg.Path }
