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
	"syscall"
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

	// Path 是这一次作业**怎么起的**（见 path.go）。它进录制件的身份，回放前的闸门
	// 拿它跟录制件比对。空 = 按"起没起作业"推默认（起了就是 r.r，没起就是 none）。
	Path LaunchPath

	// Kind 是这一轮会话的用途（SessionKindRec / SessionKindReplay，见 session.go）。
	// 空 = 录（老调用方的默认）。它只进会话登记，给"面板怎么称呼这一轮"用。
	Kind string

	// Ent 本次要用的企业编号（ENT）。空 = 不覆盖，沿用登录 shell 自己的 $TOPENT。
	//
	// 为什么必须由调用方显式给：环境配置里已经**没有**企业编号了
	// （config.NamedSsh 刻意没有 topent 字段），而作业落在哪个企业只由登录
	// shell 的 $TOPENT 决定。不给就无从指定；给错了的症状是"查不到数据"，
	// 与"这个企业本来就没有数据"完全一样（见 tt debug 的「生效身份」）。
	Ent string

	// BinaryRef 是作业编译件（.42r）在服务器上的绝对路径，给了就**只读**地探一次它的
	// 标识（大小 + 修改时间）记进录制身份。空 = 不探（走 r.r 时程序路径由登录环境决定，
	// 我们只给了作业名，所以那条路探不到 —— 这不是失败，是照实空着）。
	BinaryRef string

	// NoKeep 收工时**不**把这一轮的录制搬进长期库（`<数据目录>/records/<环境>/<ENT>/<ID>/`）。
	//
	// 默认按**归属**定（见 retention.go）：用户主动录的进库、系统自动录的留缓存。
	// 这个开关是给"用户主动起的一轮试验"用的：明知不会再用，别占库里的位置。
	NoKeep bool

	// Origin 是这一轮录制是谁发起的（`user` / `auto`）。空 = user。
	// 它决定默认留存（见 retention.go），也写进身份文件 —— 以后看一份录像为什么
	// 会在缓存目录里被清掉，答案就在这一栏。
	Origin Origin
	// Retention 是**显式**指定留在哪（空 = 按归属推默认）。`--no-keep` 走它。
	Retention Retention

	SSH host.SSHConfig

	GDCExe      string
	GDCPort     int
	ProxyPort   int
	InjectPort  int
	FGLServerIP string // empty = detect from the route to the SSH host

	DataDir   string // where the session registry lives
	RecordDir string // where the per-connection recordings land
	Tag       string // recording prefix（也是录制 ID）
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
	if c.Origin == "" {
		c.Origin = OriginUser
	}
	if c.NoKeep {
		c.Retention = RetainCache
	}
	c.Retention = retentionFor(c.Origin, c.Retention)
	if c.Path == "" {
		// 默认按"起没起作业"推：起了就是标准启动器（r.r），没起就是 none。
		// 显式给值优先 —— 调试线把作业挂到我们的代理上时，要自己声明成 debug。
		if c.NoJob {
			c.Path = PathNone
		} else {
			c.Path = PathRunnerRR
		}
	}
}

// Runtime is a live session. It lives in one resident process — the proxy must
// not die with a one-shot command — and everybody else addresses it through the
// registry and the control endpoint.
type Runtime struct {
	Session *Session

	cfg   OpenConfig
	reg   *Registry
	// screen 是这一轮的画面通道（隔离 GDC + 录制代理）。P4 起它是一等对象：
	// 起停、端口、注入都归 Screen，Runtime 只做组合。
	screen *Screen
	proxy  *Proxy // 便捷引用（= screen.Proxy()）：收尾与连接清单仍从它走
	// job 是这一轮的作业连接（可能为 nil：只起画面、不起作业）。
	job  *Job
	logf func(string, ...any)

	// ready 表示 Open 走完了全程。收工时只有 ready 的会话才落身份文件：
	// 起了一半就失败的会话没有"这一轮录了什么"可言，落一份进去只会让人以为它录到了。
	ready bool
	// fglServer 是这一次画面接线的地址（`<本机IP>:<代理端口-6400>`），进录制身份。
	fglServer string
	closeMu sync.Mutex
	closed  bool
}

func nopLog(string, ...any) {}

// Open runs the whole start-up sequence, registering the session *before*
// anything is opened so an interrupted run can always be collected.
func Open(ctx context.Context, cfg OpenConfig, logf func(string, ...any)) (*Runtime, error) {
	// "端口是不是调用方点名要的"必须在 fillDefaults **之前**看（它会把 0 填成默认值）
	portsPinned := cfg.GDCPort != 0 || cfg.ProxyPort != 0 || cfg.InjectPort != 0
	cfg.fillDefaults()
	if logf == nil {
		logf = nopLog
	}
	// 参数白名单在**起任何东西之前**：这些值会被拼进远端 shell（见 validate.go）
	if err := validateLaunch(cfg); err != nil {
		return nil, err
	}
	if !cfg.NoJob && cfg.SSH.Host == "" {
		return nil, fmt.Errorf("%w: 没给 SSH host（环境没应用或没配）", ErrPreflight)
	}
	if cfg.DataDir == "" {
		return nil, errors.New("fep: 没有数据目录（配置路径没解析出来）")
	}
	if err := preflightElevation(cfg.GDCExe); err != nil {
		return nil, err
	}
	if err := pickPorts(&cfg, portsPinned, logf); err != nil {
		return nil, err
	}

	r := &Runtime{cfg: cfg, reg: NewRegistry(cfg.DataDir), logf: logf, job: newJob(cfg, logf)}

	// 1–2. 画面通道（隔离 GDC + 录制代理）：交给 Screen 一处起停，Runtime 只做组合。
	screen, err := StartScreen(ScreenConfig{
		GDCExe: cfg.GDCExe, GDCLog: cfg.GDCLog,
		GDCPort: cfg.GDCPort, ProxyPort: cfg.ProxyPort, InjectPort: cfg.InjectPort,
		RecordDir: cfg.RecordDir, Tag: cfg.Tag, WaitGDC: cfg.WaitGDC,
	}, logf)
	if err != nil {
		return nil, err
	}
	r.screen = screen
	r.proxy = screen.Proxy()

	// 3. register first, clean up later
	sess := &Session{
		Env: cfg.Env, Zone: cfg.Zone, Prog: cfg.Prog, Ent: cfg.Ent, Path: cfg.Path,
		Kind:   cfg.Kind,
		GDCPID: screen.GDCPID(), GDCPort: cfg.GDCPort,
		ProxyPort: cfg.ProxyPort, InjectPort: cfg.InjectPort,
		Tag: cfg.Tag, RecordDir: cfg.RecordDir, OwnerPID: os.Getpid(), Started: time.Now(),
	}
	if err := r.reg.Save(sess); err != nil {
		_ = screen.Close()
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
		if err := r.job.Open(ip, r.setFGLServer); err != nil {
			r.Close()
			return nil, err
		}
		// 「用哪个企业」必须**说出来**：它决定了作业连哪个 schema，而错了的症状
		// 是"查不到数据"，与"数据不存在"无从区分。不给就如实说"沿用登录 shell 的"。
		if cfg.Ent != "" {
			logf("企业编号已下发：TOPENT=%s", cfg.Ent)
		} else {
			logf("没给企业编号（--ent）：沿用登录 shell 自己的 $TOPENT")
		}
		logf("作业已打开：%s，FGLSERVER=%s:%d", cfg.Prog, ip, cfg.ProxyPort-FGLSERVEROffset)
		// 编译件标识（只读，探不到不影响本轮）—— 录制身份里要用它发现"程序重编译过"。
		r.job.ProbeBinary()
		if b := r.job.Binary(); !b.Empty() {
			logf("编译件：%s", b.FingerprintString())
		}
	} else {
		// 只起代理时也要把画面接线算出来：`tt recording start` 就是"先起代理、再把调试线的
		// 作业挂到它上面"，而它得知道这个地址才知道该给调试线传什么 FGLSERVER。
		// 探测失败不算错（裸代理可以没有环境）：照实留空，由调用方决定怎么办。
		ip := cfg.FGLServerIP
		if ip == "" && cfg.SSH.Host != "" {
			if got, aerr := localAddrFor(cfg.SSH); aerr == nil {
				ip = got
				logf("FGLSERVER 用本机地址 %s（按到 %s 的路由探测）", ip, cfg.SSH.Host)
			}
		}
		if ip != "" {
			if err := r.setFGLServer(ip); err != nil {
				r.Close()
				return nil, err
			}
		}
	}
	r.ready = true
	return r, nil
}

// setFGLServer 记下"画面接到哪"：`<本机IP>:<代理端口 - FGLSERVEROffset>`。
//
// 端口要自己算并校验：FGLSERVER 里写的数值与代理真正监听的端口差一个固定偏移，
// 算错的表现是**安静**的 —— 隔离实例起来了、命令也发出去了，作业那边一片空白。
func (r *Runtime) setFGLServer(ip string) error {
	port := r.cfg.ProxyPort - FGLSERVEROffset
	if port <= 0 || port > 65535 {
		return fmt.Errorf("%w: 代理端口 %d 减 %d 后不是合法端口", ErrUsage, r.cfg.ProxyPort, FGLSERVEROffset)
	}
	r.fglServer = fmt.Sprintf("%s:%d", ip, port)
	return nil
}

// FGLServer 报告这一轮的画面接线（`<本机IP>:<代理端口-6400>`；算不出来时是空串）。
//
// 调用方拿它去告诉"谁来起作业"（调试线的 `--fglserver`）—— 实测值可以直接用，
// 但它只在这一轮有效：换一轮就是另一个端口与另一个本机地址。
func (r *Runtime) FGLServer() string { return r.fglServer }

// Wait blocks until the session ends: the caller's context is cancelled, a
// control client sent `quit` (`tt recording stop <录制ID>`), or the remote job exited.
func (r *Runtime) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case <-r.proxy.Quit():
		return nil
	case <-jobDone(r.job):
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
	// 作业连接整块收（PTY + SSH 都归 Job 管）。
	if r.job != nil {
		if err := r.job.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	// 连接清单要在**停代理之前**取：代理一关，运行期连进来的那些就从表里删了，
	// 而身份文件要如实写下"这一轮录到了几条连接、各多少字节"。
	var conns []ConnInfo
	if r.proxy != nil {
		conns = r.proxy.Conns()
	}
	// 画面通道整块收（代理 + GDC 都归它管）。P4 之前这里是"关代理"再靠 Session.PIDs()
	// 顺手杀 GDC —— 两条收尾路径分开写，漏一条就留下一个孤儿进程。
	if r.screen != nil {
		if err := r.screen.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	} else if r.proxy != nil {
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
	// 收工落身份（见 meta.go 的为什么）。它失败要**说出来**：没有身份的录制件
	// 以后只能靠"算指纹"用，路径与编译件都无从得知，而人会以为那份录像什么都没缺。
	if r.ready {
		path, err := r.writeMeta(conns)
		if err != nil {
			errs = append(errs, fmt.Sprintf("落录制身份失败: %v", err))
		} else if path != "" {
			r.logf("录制身份已落：%s", path)
			// 入长期库（R3）：缓存目录会被清，而录制是"以后还要用"的东西。
			// 入库失败**不**算收尾失败 —— 录制件与身份都还在缓存目录里，能用，只是会被清掉。
			if r.cfg.Retention == RetainCache {
				r.logf("按归属（%s）留在缓存目录：%s —— 它会随统一清理框架按年龄删", r.cfg.Origin.Label(), r.cfg.RecordDir)
			} else if item, perr := NewLibrary(r.cfg.DataDir).Promote(r.cfg.RecordDir, r.cfg.Tag); perr != nil {
				r.logf("没能入长期库（录制件留在缓存目录，清缓存会删它）：%v", perr)
				// D28：这一轮**压根没生成流量文件**时，刚写的身份描述的是一份**不存在的录像**。
				// 留着它会让人以为"有一份录制可以回放"（真机冒烟里正是这样被误导的：
				// 日志说"录制身份已落"，可缓存目录里根本没有对应的录制件）。
				// 只删**刚写的这一份**（path 来自 writeMeta），别的一概不碰。
				if dropMetaIfNoRecording(r.cfg.RecordDir, r.cfg.Tag) {
					if rerr := os.Remove(path); rerr == nil {
						r.logf("这一轮没有录制件，已删掉空身份文件：%s", filepath.Base(path))
					}
				}
			} else {
				r.logf("已入长期库：%s", item.Dir)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("fep: 收尾有失败: %s", strings.Join(errs, "; "))
	}
	return nil
}

// dropMetaIfNoRecording 判「这一份身份该不该删」：**没有流量文件**就该删。
//
// 为什么只认流量文件：L1 真机冒烟里出现过「日志说身份已落、缓存里却没有录制件」的情形
// （录制在别处失败）。那份身份描述的是**一份不存在的录像**，留着会让人以为能回放。
// 判据只按**这一份的 tag** 找文件 —— 别人的录制件不算（否则会误判成"有录制"而漏删）。
func dropMetaIfNoRecording(recordDir, tag string) bool {
	files, err := recordingLogFiles(recordDir, tag)
	return err == nil && len(files) == 0
}

// writeMeta 把这一轮的录制身份落成 `<录制ID>.meta.json`（与录制件同目录）。
//
// 指纹在这里算而不是只记路径：收工这一刻的连接与字节都还在手上，读完这一次之后
// 谁也不必再拼一遍"这一轮到底录到了什么界面"。
func (r *Runtime) writeMeta(conns []ConnInfo) (string, error) {
	mt := RecordingMeta{
		ID: r.cfg.Tag, Path: r.cfg.Path, Prog: r.cfg.Prog,
		Origin: r.cfg.Origin, Retention: r.cfg.Retention,
		Env: r.cfg.Env, Zone: r.cfg.Zone, Ent: r.cfg.Ent,
		Host: r.cfg.SSH.Host, FGLServer: r.fglServer,
		Started: r.Session.Started, Finished: time.Now(),
		Binary: jobBinary(r.job),
	}
	for _, c := range conns {
		mt.Conns = append(mt.Conns, MetaConn{
			ID: c.ID, Remote: c.Remote,
			RT2GDC:   fmt.Sprintf("%s-%d-rt2gdc.log", r.cfg.Tag, c.ID),
			GDC2RT:   fmt.Sprintf("%s-%d-gdc2rt.log", r.cfg.Tag, c.ID),
			Timeline: fmt.Sprintf("%s-%d-timeline.log", r.cfg.Tag, c.ID),
			BytesRT:  c.BytesRT, BytesFE: c.BytesFE,
		})
	}
	// 指纹从**录制下来的字节**算（不是从内存里的模型）：这一份是以后要审的东西，
	// 手上这个模型只代表"收工那一刻"。
	if m, _, _, err := LoadModelFrom(r.cfg.RecordDir, r.cfg.Tag); err == nil {
		mt.Fingerprint = FingerprintOf(EntriesOf(m))
	} else {
		mt.Note = "这一轮没读到 AUI 消息：" + err.Error()
	}
	return WriteMeta(r.cfg.RecordDir, mt)
}

// jobLaunchCmd 拼出启动那一段 shell 命令：先把企业编号下发到登录 shell，
// 再把画面指向我们的代理（FGLSERVER），最后启动作业。
//
// 顺序不能换：FGLSERVER 与 TOPENT 都必须在启动器（`r.r` / `r.d`）**之前**生效 ——
// 反了的话作业会带着上一轮的值跑起来，而 ENT 给错的表现是"查不到数据"，与"这个企业
// 本来就没数据"完全一样，没人能从结果里看出来。
//
// 返回值可能带错：`ip`、`prog` 是**裸插**进这条命令的，路径还要是已知的那几条
// （见 validate.go）。调用方不许忽略这个错。
func jobLaunchCmd(path LaunchPath, ip string, port int, prog, ent string) (string, error) {
	if err := ValidateProg(prog); err != nil {
		return "", err
	}
	if err := ValidateFGLServerIP(ip); err != nil {
		return "", err
	}
	launch, err := launchCommand(path, prog)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if ent != "" {
		b.WriteString(topentExportLine(ent))
	}
	fmt.Fprintf(&b, "export FGLSERVER=%s:%d\r\n%s\r\n", ip, port, launch)
	return b.String(), nil
}

// topentExportLine 拼出把企业编号下发到登录 shell 的那一行。
//
// 值与 tt debug 同一套规则（见 internal/debug 的 topentShellCmd）：单引号包裹
// **并剔除内嵌单引号**。值不限数字/文本，剔除是为了防注入/拆词 ——
// 少一个引号就能把整条命令拆开。两处规则必须一致，改一处就要改另一处。
func topentExportLine(ent string) string {
	return "export TOPENT='" + strings.ReplaceAll(ent, "'", "") + "'\r\n"
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

// waitAny waits until any of the patterns appears, returning its index.
// One reader only: the goroutine stops as soon as one matches.
func waitAny(pty *host.PTYSession, timeout time.Duration, res ...*regexp.Regexp) (int, error) {
	type outcome struct {
		idx int
		err error
	}
	ch := make(chan outcome, 1)
	go func() {
		buf := make([]byte, 4096)
		var acc []byte
		for {
			n, err := pty.Read(buf)
			if n > 0 {
				acc = append(acc, buf[:n]...)
				for i, re := range res {
					if re.Match(acc) {
						ch <- outcome{idx: i}
						return
					}
				}
				if len(acc) > 1<<20 {
					acc = acc[len(acc)-1<<16:]
				}
			}
			if err != nil {
				ch <- outcome{err: err}
				return
			}
		}
	}()
	select {
	case o := <-ch:
		if o.err != nil {
			return -1, fmt.Errorf("fep: 等登录提示时连接断了: %w", o.err)
		}
		return o.idx, nil
	case <-time.After(timeout):
		return -1, fmt.Errorf("fep: 等登录提示超时（%s）", timeout)
	}
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
		// "被占"有两种成因，报错要说清是哪一种 —— Windows 上很常见的一种是**根本没人监听**，
		// 而是这个端口落在系统/Hyper-V/WSL 保留的端口段里（bind 直接 WSAEACCES），
		// 这时说"已被占用"会把查的人指向错的方向（去看谁在听，当然什么也看不到）。
		if errors.Is(err, syscall.EACCES) || strings.Contains(err.Error(), "forbidden") {
			return fmt.Errorf("%w: %d 用不了（被系统保留，不是有人占用）："+
				"Windows 上常见于 Hyper-V/WSL 保留的端口段 —— "+
				"`netsh interface ipv4 show excludedportrange protocol=tcp` 看一眼，换个端口或重启 winnat 服务",
				ErrPortBusy, port)
		}
		return fmt.Errorf("%w: %d 已被占用（%v）", ErrPortBusy, port, err)
	}
	ln.Close()
	return nil
}

// fepPortOffsets 是一组端口里三个端口的相对关系（gdc / proxy / inject）。
//
// 必须成套：画面接线里那个数值恒等于**代理端口 - 6400**（见 open.go 里算 FGLSERVER 的那段），
// 所以换端口时三个一起换、偏移保持不变，服务端才能连回本机的代理。
var fepPortOffsets = [3]int{0, 200, 499}

// maxPortTries 是自动换端口时最多试几组（每组错开 1000）。
const maxPortTries = 8

// pickPorts 决定这一轮用哪三个端口。
//
// 为什么不是"检查一遍默认端口、不通就报错"：端口是 OS 层面的稀缺资源，而 Windows 会把
// **成片**的端口划给 Hyper-V/WSL 保留（`excludedportrange`）—— 落在里面的端口 bind 直接
// WSAEACCES，没有进程在监听也一样失败（实测 7899 落在 7809–7908 里，`Get-NetTCPConnection`
// 看是"空闲"）。这种时候默认那一组就整组不可用，而它跟我们的代码没半点关系。
//
// 所以：**调用方点名要的端口**（非 0）照旧尊重，不通就报错；否则默认那一组不通时，
// 自动往前挪 1000 再试（最多 maxPortTries 组），并把换端口这件事写进日志。
func pickPorts(cfg *OpenConfig, pinned bool, logf func(string, ...any)) error {
	ports := func(base int) [3]int {
		return [3]int{base + fepPortOffsets[0], base + fepPortOffsets[1], base + fepPortOffsets[2]}
	}
	want := ports(cfg.GDCPort)
	if pinned {
		// 点名要的端口就检查**调用方给的那三个**（也包括"只点了其中一个、其余走默认"这种混合）。
		// 从 GDC 端口按 +200/+499 推另外两个只对"默认那一组"成立：按偏移推的话，
		// 调用方给的 ProxyPort 被占**查不出来**（要等隔离实例起完、代理那一步才失败，
		// 白起一个 gdc），而 GDCPort+200 恰好被占时又会误报。这条是实测出来的（open_test.go）。
		for _, p := range []int{cfg.GDCPort, cfg.ProxyPort, cfg.InjectPort} {
			if err := portFree(p); err != nil {
				return err
			}
		}
		return nil
	}
	var firstErr error
	for i := 0; i < maxPortTries; i++ {
		try := ports(cfg.GDCPort + i*1000)
		busy := -1
		for _, p := range try {
			if err := portFree(p); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				busy = p
				break
			}
		}
		if busy >= 0 {
			continue
		}
		if i > 0 {
			logf("默认端口 %d/%d/%d 用不了：这一轮改用 %d/%d/%d（画面接线端口 = 代理端口 - 6400）",
				want[0], want[1], want[2], try[0], try[1], try[2])
		}
		cfg.GDCPort, cfg.ProxyPort, cfg.InjectPort = try[0], try[1], try[2]
		return nil
	}
	return fmt.Errorf("%w: 从 %d 起连着 %d 组端口都用不了：%v",
		ErrPortBusy, want[0], maxPortTries, firstErr)
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
	if ssh.Host == "" {
		return "", fmt.Errorf("%w: 环境没配 SSH host，本机地址无从探测", ErrPreflight)
	}
	port := ssh.Port
	if port == 0 {
		port = 22
	}
	c, err := net.Dial("udp", net.JoinHostPort(ssh.Host, strconv.Itoa(port)))
	if err != nil {
		return "", fmt.Errorf("fep: 探测本机地址失败: %w", err)
	}
	defer c.Close()
	la, ok := c.LocalAddr().(*net.UDPAddr)
	if !ok || la.IP == nil || la.IP.IsUnspecified() {
		return "", errors.New("fep: 探测本机地址失败（路由没给出源地址）")
	}
	ip := la.IP.String()
	// 回环地址在这里一定是错的：服务器不可能连回本机的 loopback（实测到过：
	// SSH host 为空时探测出 127.0.0.1，会话看起来起来了，作业却永远连不回来）。
	if la.IP.IsLoopback() {
		return "", fmt.Errorf("%w: 探测到的是回环地址 %s，服务器连不回来（用 --fgserver-ip 明确指定可达地址）", ErrPreflight, ip)
	}
	return ip, nil
}

// RegistryPath reports the session registry file (for diagnostics).
func (r *Runtime) RegistryPath() string { return r.reg.Path }

// SetRef 记下"这条录制落在调试线的哪条会话上"（C2.0 的 `环境#序号`）。
//
// 为什么由调用方在**会话建立之后**回填：登记的时机是"起代理时"（register first, clean up later），
// 那时调试会话还不存在。等它起来了再把地址补进来，两个世界就有了同一个坐标。
//
// 只改这一条：`Registry.Save` 是按 Tag upsert 的，不会动别的登记。
func (r *Runtime) SetRef(ref string) error {
	if r == nil || r.Session == nil {
		return nil
	}
	ref = strings.TrimSpace(ref)
	if ref == "" || r.Session.Ref == ref {
		return nil
	}
	prev := r.Session.Ref
	r.Session.Ref = ref
	if err := r.reg.Save(r.Session); err != nil {
		r.Session.Ref = prev
		return err
	}
	return nil
}
