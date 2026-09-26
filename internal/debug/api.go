package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/coder/websocket"

	"tt/internal/host"
)

// Server 本地调试服务:REST + WebSocket + Web 前端
type Server struct {
	cfg     *Config
	cfgPath string // config.json 路径(重读配置用,可为空=只读)
	web     fs.FS
	webSub  fs.FS  // web/dist 子文件系统(未构建前端时为 nil)
	addr    string // 实际监听地址(端口顺延时与 cfg.Listen 不同;不写回配置)
	mgr     *Manager

	// dial 是拨 SSH 的入口。抽成**字段**是为了让默认测试不拨真机
	// （与 internal/dev/tzs 的 Dialer 接口、internal/web.Options 的函数字段同一个理由）。
	// 五个 handler 直接调它，注入一个必然失败的实现就能测它们的 400/500 分支 ——
	// 那批分支里有一处**顺序**性质（入参先校验、再收口会话，见 hWSLogDebug 的注释），
	// 改动它没有任何东西会响。
	//
	// NewServer 里默认成 host.Dial，所以**生产路径的行为与从前一字不差**；
	// 只有 _test.go 会替换它。
	dial func(host.SSHConfig) (*host.SSHConn, error)

	// mu 同时保护两类运行态:
	//   stop —— Serve 期间指向该次运行的 cancel(POST /api/shutdown 用,优雅停止);
	//   *cfg 的替换 —— 配置热替换必须整个换掉,不能让别人看到半个新配置。
	// 合并后 hosts 节由统一服务的 /api/hosts 统一读写,调试服务要能安全地重读并热替换,
	// 所以 cfg 的替换也收进这把锁(原来是直接 *s.cfg = nc,没有任何同步)。
	mu   sync.Mutex
	stop func()
}

// NewServer 创建服务实例；cfgPath 为 config.json 路径(重读配置用,可为空=只读)。
//
// web 是**本应用 SPA 的根**（即 web/dist 的内容），不是模块根的嵌入 FS ——
// 合并前两套前端各占 dist 下的一个子目录，剥离那一层得由调用方做；合并后只剩一套 SPA、
// 直接产出在 dist 根，调用方（tt serve / tt debug serve）传 common.WebFrontend() 即可。
// 这里只确认 index.html 在；不在就当没前端，回落成纯 API + 引导页。
func NewServer(cfg *Config, web fs.FS, cfgPath string) *Server {
	// 先把默认环境(hosts.activeEnv,可被 debug.activeEnv 覆盖)的连接与启动参数
	// 合并到运行时字段,再交给 Manager。统一服务 tt serve 只调 LoadConfigAllowEmpty
	// 就直接 NewServer,不会自己补这一步 —— 少了它 s.cfg.SSH/Zone/Topent/DB 全是空的,
	// /api/status 报不出主机,从 /api/sessions 起的会话也会拿空地址去连。
	// 重复调用无副作用(CLI 的 serve 路径本来就会先调一次)。
	cfg.ApplyDefaultEnv()
	s := &Server{cfg: cfg, cfgPath: cfgPath, web: web, mgr: NewManager(cfg), dial: host.Dial}
	if web != nil {
		if f, err := web.Open("index.html"); err == nil {
			f.Close()
			s.webSub = web
		}
	}
	return s
}

// Run 启动 HTTP 服务(阻塞到 ctx 取消)。
// 先按 cfg.Listen 建监听;端口被占用时自动顺延到下一个空闲端口(最多尝试 maxPortTries 个),
// 实际监听地址记在 s.addr(不改 cfg.Listen,避免顺延后的端口被设置页保存写回配置)。
func (s *Server) Run(ctx context.Context) error {
	ln, addr, err := s.Listen()
	if err != nil {
		return err
	}
	fmt.Printf("[tt debug] 服务已启动  前端+API: http://%s\n", addr)
	return s.Serve(ctx, ln)
}

// maxPortTries 端口被占用时最多顺延尝试的次数。
const maxPortTries = 50

// Listen 建监听;cfg.Listen 被占用时尝试后续端口,成功后将实际地址记入 s.addr。
// 返回 listener 与实际监听地址(host:port);cfg.Listen 保持配置原值不变。
func (s *Server) Listen() (net.Listener, string, error) {
	host, portStr, err := net.SplitHostPort(s.cfg.Listen)
	if err != nil {
		return nil, "", fmt.Errorf("非法监听地址 %q: %w", s.cfg.Listen, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		return nil, "", fmt.Errorf("非法监听端口 %q: %w", portStr, err)
	}
	var lastErr error
	for i := 0; i < maxPortTries; i++ {
		addr := net.JoinHostPort(host, strconv.Itoa(port+i))
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			// 报出真实监听地址:--listen 端口写 0 时端口由系统分配,只能回读才知道;
			// 端口顺延时同理(原来的 requested 端口只用来重试,不作为结果上报)。
			// host 为空 = 监听全部网卡,那种地址浏览器/壳用不了,上报回环地址。
			real := addr
			if tcp, ok := ln.Addr().(*net.TCPAddr); ok {
				h := host
				if h == "" {
					h = "127.0.0.1"
				}
				real = net.JoinHostPort(h, strconv.Itoa(tcp.Port))
			}
			s.addr = real
			return ln, real, nil
		}
		lastErr = err
	}
	return nil, "", fmt.Errorf("监听 %s 失败(尝试 %d 个端口均不可用): %w",
		s.cfg.Listen, maxPortTries, lastErr)
}

// Serve 用既有 listener 提供 HTTP 服务(阻塞到 ctx 取消或连接关闭)。
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	srv := &http.Server{Handler: s.Handler()}

	// 记下本次运行的 cancel,供 POST /api/shutdown 优雅停止
	runCtx, cancelRun := context.WithCancel(ctx)
	s.setStop(cancelRun)
	defer s.setStop(nil)

	go func() {
		<-runCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	err := srv.Serve(ln)
	if err == http.ErrServerClosed {
		s.mgr.CloseAll()
		return nil
	}
	return err
}

// setStop 记录/清除本次运行的停止钩子。
func (s *Server) setStop(f func()) {
	s.mu.Lock()
	s.stop = f
	s.mu.Unlock()
}

// ListenAddr 返回实际监听地址(端口顺延时与配置值不同;未启动时返回配置值)。
func (s *Server) ListenAddr() string {
	if s.addr != "" {
		return s.addr
	}
	return s.cfg.Listen
}

// Handler 返回调试工作台的全部 HTTP 面(REST + WebSocket + 内嵌 SPA 静态资源)。
//
// 面内的路径都是**相对挂载点**的,与合并前独立起服务时一模一样:
//
//	/api/status /api/shutdown /api/events /api/source-file /api/jobinfo
//	/api/sessions…(会话/断点/控制/求值/源码/接口日志)/api/wstest /api/wslogs…
//	/api/dbsql /api/source-preview /api/ws
//	/api/hosts(只读,见 hHostsGet)
//	/(SPA 兜底,hStatic)
//
// 统一服务(tt serve)用 http.StripPrefix("/debug", srv.Handler()) 把它挂在 /debug/ 下,
// 于是面内的 /api/sessions/… 对外是 /debug/api/sessions/…,内嵌 SPA 的 / 对外是 /debug/。
// 这样一个进程里,调试面内的 /api/status、/api/ws 等路径都落在 /debug/api/ 下,
// 与共享层 /api/* 的同名路径不会互相覆盖 —— 合并前两个工具各起各的服务,压根不存在
// 冲突;合并后靠挂载前缀区分。
//
// 注意:统一的 /api/health、/api/hosts、/api/dbprobe、/api/dbaccverify、/api/conntest、
// /api/config/meta 由 tt/internal/web 在顶层持有。本面挂在 /debug/ 前缀下,与它们不在
// 同一个 mux 上,故不会重复注册;hHostsGet 是给**独立运行**(tt debug serve)时的
// tt debug env 用的,挂到前缀下时即 /debug/api/hosts,不参与统一服务的路由。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.routes(mux)
	return mux
}

func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/status", s.hStatus)
	mux.HandleFunc("POST /api/shutdown", s.hShutdown)
	mux.HandleFunc("GET /api/events", s.hEvents)
	mux.HandleFunc("GET /api/source-file", s.hSourceFile)
	mux.HandleFunc("GET /api/jobinfo", s.hJobInfo)
	mux.HandleFunc("POST /api/sessions", s.hLaunch)
	mux.HandleFunc("GET /api/sessions", s.hList)
	mux.HandleFunc("GET /api/sessions/{id}", s.hSnapshot)
	mux.HandleFunc("DELETE /api/sessions/{id}", s.hQuit)
	mux.HandleFunc("POST /api/sessions/{id}/restart", s.hSessionRestart)
	mux.HandleFunc("POST /api/sessions/{id}/close", s.hSessionClose)
	mux.HandleFunc("POST /api/sessions/{id}/topent", s.hTopent)
	mux.HandleFunc("POST /api/sessions/{id}/mode", s.hMode)
	mux.HandleFunc("POST /api/sessions/switch", s.hSessionSwitch)
	mux.HandleFunc("GET /api/sessions/{id}/breakpoints", s.hBPList)
	mux.HandleFunc("POST /api/sessions/{id}/breakpoints", s.hBPAdd)
	mux.HandleFunc("DELETE /api/sessions/{id}/breakpoints/{num}", s.hBPDel)
	mux.HandleFunc("POST /api/sessions/{id}/control", s.hControl)
	mux.HandleFunc("POST /api/sessions/{id}/print", s.hPrint)
	mux.HandleFunc("POST /api/sessions/{id}/where", s.hWhere)
	mux.HandleFunc("POST /api/sessions/{id}/raw", s.hRaw)
	mux.HandleFunc("POST /api/sessions/{id}/why", s.hWhy)
	mux.HandleFunc("GET /api/sessions/{id}/wait", s.hWait)
	mux.HandleFunc("GET /api/sessions/{id}/locals", s.hLocals)
	mux.HandleFunc("GET /api/sessions/{id}/globals", s.hGlobals)
	mux.HandleFunc("GET /api/sessions/{id}/sources", s.hSources)
	mux.HandleFunc("GET /api/sessions/{id}/functions", s.hFunctions)
	mux.HandleFunc("GET /api/sessions/{id}/autovars", s.hAutovars)
	mux.HandleFunc("POST /api/sessions/{id}/autovars", s.hAutovarsSet)
	mux.HandleFunc("POST /api/sessions/{id}/frame", s.hFrame)
	mux.HandleFunc("POST /api/sessions/{id}/locate", s.hLocate)
	mux.HandleFunc("POST /api/sessions/{id}/calibrate", s.hCalibrate)
	mux.HandleFunc("POST /api/sessions/{id}/breakpoints/{num}/enabled", s.hBPEnabled)
	mux.HandleFunc("GET /api/sessions/{id}/source", s.hSource)
	mux.HandleFunc("GET /api/source-preview", s.hSourcePreview)
	mux.HandleFunc("POST /api/wstest", s.hWSTest)
	mux.HandleFunc("GET /api/wslogs", s.hWSLogs)
	mux.HandleFunc("POST /api/dbsql", s.hDBSQL)
	mux.HandleFunc("GET /api/wslogs/content", s.hWSLogContent)
	mux.HandleFunc("POST /api/wslogs/debug", s.hWSLogDebug)
	// 环境清单(只读):独立运行时 tt debug env 读它。统一服务下环境页改的是
	// 顶层 /api/hosts(tt/internal/web),这里那份只在 /debug/ 前缀下可达,互不干扰。
	mux.HandleFunc("GET /api/hosts", s.hHostsGet)
	mux.HandleFunc("GET /api/ws", s.hWS)
	mux.HandleFunc("/", s.hStatic)
}

// ---------- 工具 ----------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]any{"ok": false, "error": err.Error()})
}

func readBody[T any](w http.ResponseWriter, r *http.Request, out *T) bool {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		fail(w, 400, fmt.Errorf("请求体解析失败: %v", err))
		return false
	}
	return true
}

func (s *Server) sessOf(w http.ResponseWriter, r *http.Request) *Session {
	id := r.PathValue("id")
	sess := s.mgr.Get(id)
	if sess == nil {
		fail(w, 404, fmt.Errorf("会话不存在: %s", id))
	}
	return sess
}

// ---------- 归因与模式闸门 ----------

// actorHeader 调用方声明自己是谁。人与 AI 打的是同一批 REST 端点,服务端分辨不出来,
// 只能靠声明:CLI 统一带 X-Actor: ai,浏览器带 human(不带也按 human 算)。
const actorHeader = "X-Actor"

func actorOf(r *http.Request) string {
	if strings.EqualFold(strings.TrimSpace(r.Header.Get(actorHeader)), "ai") {
		return "ai"
	}
	return "human"
}

// launchModeFor 新会话的初始模式,由**发起方**决定:
// AI 从 CLI 发起的调试(start / wsdebug)默认协作(人只读),人从界面发起的保持纯人工。
// 两条启动路径(hLaunch / hWSLogDebug)必须都过这里 —— 漏一条就会出现
// "AI 起出来的会话却写着纯人工、AI 自己被自己的闸门拦下"的怪状。
func launchModeFor(r *http.Request) string {
	if actorOf(r) == "ai" {
		return ModeCollab
	}
	return ModeSolo
}

// MayWrite 该身份此刻能否对这个会话执行写操作。
// 闸门是**会话级的静态标志**,不做逐命令协商 —— 原先设想的 holder/TTL/续期那一套
// 最难的地方是服务端看不见 AI 的"一轮"(CLI 每次调用都是独立进程),降级成模式之后
// 这一整块就不存在了。
func (s *Session) MayWrite(actor string) bool {
	if s.Mode() == ModeCollab {
		return actor == "ai"
	}
	// solo:AI 只能读。否则"纯人工模式下 AI 只能读不能写"就是摆设。
	return actor != "ai"
}

// writeDeniedMsg 拒绝文案要能直接照做,而不是只说"没权限"。
func writeDeniedMsg(mode, actor string) string {
	if mode == ModeCollab {
		return "协作模式:本次调试由 AI 主导,请在对话中把操作委托给 AI(例如「在 4452 行下个断点」);" +
			"需要自己动手请点界面上的「接管」。"
	}
	if actor == "ai" {
		return "纯人工模式:AI 不可操作。请让用户在界面上点「交给 AI」,或由用户自己操作。"
	}
	return "当前模式不允许该操作"
}

// hMode 切换会话模式(纯人工 / 协作)。
// body: {"mode": "solo"|"collab"}
//
// 规则收窄了一步,理由是字面对称会开一个洞:**人可任意方向切**(人始终有最终控制权);
// **AI 只能请求,不能自升权限** —— 若 AI 能自己把 solo 切成 collab,那"纯人工模式下
// AI 只能读不能写"就是摆设,它一秒钟就能把自己放出来。
//
// 切换**不中断在飞命令**,只改变"之后谁能写",所以本身没有竞态。
func (s *Server) hMode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mode string `json:"mode"`
	}
	if !readBody(w, r, &req) {
		return
	}
	m := strings.ToLower(strings.TrimSpace(req.Mode))
	if m != ModeSolo && m != ModeCollab {
		fail(w, 400, fmt.Errorf("mode 只能是 %s 或 %s", ModeSolo, ModeCollab))
		return
	}
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	actor := actorOf(r)
	if actor == "ai" && m == ModeCollab && sess.Mode() != ModeCollab {
		fail(w, 403, fmt.Errorf("请让用户在界面上点「交给 AI」:AI 不能自行解除纯人工模式"))
		return
	}
	text := map[string]string{ModeSolo: "已接管:纯人工模式", ModeCollab: "已交给 AI:协作模式"}[m]
	sess.SetMode(m)
	ev := Event{Type: "log", SessionID: sess.ID, Time: time.Now(), Actor: actor, Action: "session.mode", Text: text}
	if actor == "ai" {
		ev.Type = "ai_action" // 让时间线上明确标出是 AI 做的
	}
	s.mgr.emit(ev)
	writeJSON(w, 200, map[string]any{"ok": true, "mode": m})
}

// sessOfWriteGlobal 用于**不按 {id} 定位**的写操作(切环境、按日志重放):
// 以"当前会话"的模式为准。没有当前会话时放行 —— 没有东西可保护。
// 归因留给调用方在操作成功之后自己调 emitAction(此时才有新会话 id)。
//
// 空会话同样放行:前端一连上来就会建一个还没挂程序的宿主会话,而它是 human 身份
// 建的(→ 纯人工)。若把它也算作"要保护的一轮",AI 就被一个什么都没在跑的会话
// 永远关在门外 —— 连自己发起调试都做不到。没有程序,就没有可被抢走的运行。
func (s *Server) sessOfWriteGlobal(w http.ResponseWriter, r *http.Request) bool {
	cur := s.mgr.Current()
	if cur == nil || cur.Bare() {
		return true
	}
	if actor := actorOf(r); !cur.MayWrite(actor) {
		fail(w, 403, fmt.Errorf("%s", writeDeniedMsg(cur.Mode(), actor)))
		return false
	}
	return true
}

// emitAction 把一次 AI 写操作记进操作时间线(前端按 origin:'ai' 渲染成蓝色行)。
//
// 在**操作成功之后**调用,不在放行时调用:被拒(模式闸门/上一条命令还在跑/参数错)
// 的命令不该在时间线上留痕 —— 真机验证时,一次忙等把 60 多次失败的 next 全记了进去。
// 长命令的"正在进行"由快照里的 inflight 承担(状态栏实时显示"⟳ next(已 12s)"),
// 不需要时间线抢答。
func (s *Server) emitAction(r *http.Request, sess *Session, action, text string) {
	if sess == nil || actorOf(r) != "ai" {
		return
	}
	s.mgr.emit(Event{
		Type: "ai_action", SessionID: sess.ID, Time: time.Now(),
		Actor: "ai", Action: action, Text: text,
	})
}

// sessOfWrite 写操作的会话入口 = sessOf + 模式闸门。被拒时已写好 403 并返回 nil。
//
// 闸门是会话级的静态标志,不做逐命令协商:协作模式下只有 AI 能写,
// 纯人工模式下只有人能写。切换靠界面上的按钮,不进这条路径。
func (s *Server) sessOfWrite(w http.ResponseWriter, r *http.Request) *Session {
	sess := s.sessOf(w, r)
	if sess == nil {
		return nil
	}
	if actor := actorOf(r); !sess.MayWrite(actor) {
		fail(w, 403, fmt.Errorf("%s", writeDeniedMsg(sess.Mode(), actor)))
		return nil
	}
	return sess
}

// ---------- 会话 ----------

func (s *Server) hStatus(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{
		"server":   "tdebug-debug",
		"ssh":      s.cfg.SSH.Host,
		"zone":     s.cfg.Zone,
		"zoneName": host.ZoneTNSName(s.cfg.Zone), // 区域代码 → T100 服务别名(如 36→t35prd;服务测试默认地址要用)
		"listen":   s.ListenAddr(),
		"watchdog": s.cfg.WatchdogSeconds,
	}
	// topDir 只在登录动态获取后才有意义(未登录无静态值);无会话/未探测时省略
	if top := s.cfg.TopDirActual(); top != "" {
		m["topDir"] = top
	}
	writeJSON(w, 200, m)
}

// hEvents 最近会话事件(tail):GET /api/events?tail=N,供 AI/CLI 观察"发生了什么"。
func (s *Server) hEvents(w http.ResponseWriter, r *http.Request) {
	tail, _ := strconv.Atoi(r.URL.Query().Get("tail"))
	writeJSON(w, 200, map[string]any{"ok": true, "events": s.mgr.Events(tail)})
}

// hShutdown 优雅停止服务:POST /api/shutdown {}。
// 效果等同 Ctrl+C —— ctx 取消 → http 收口 + mgr.CloseAll 收掉会话与 SSH 连接。
// 停服务的主要路径是 `tt serve --stop` / `tt debug serve --stop`(按 pid 杀进程);
// 这个端点留给"说得上话但拿不到 pid"的调用方(曾经是桌面外壳的关窗动作)。
// 要求 JSON 请求体(与全站一致):浏览器跨站发不出这个 Content-Type(会先被预检挡下),
// 所以只有本机同源调用方能用,不额外加鉴权。
func (s *Server) hShutdown(w http.ResponseWriter, r *http.Request) {
	var req struct{}
	if !readBody(w, r, &req) {
		return
	}
	s.mu.Lock()
	stop := s.stop
	s.mu.Unlock()
	if stop == nil {
		fail(w, 409, fmt.Errorf("服务尚未就绪"))
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
	go stop() // 先回响应再停,避免调用方拿到连接重置
}

// hSourceFile 会话外白名单源码读取:GET /api/source-file?module=&file=&path=&from=&to=
// (AI 信息通道;独立短连接,不经会话,路径限制在登录区源码目录)
func (s *Server) hSourceFile(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from, _ := strconv.Atoi(q.Get("from"))
	to, _ := strconv.Atoi(q.Get("to"))
	res, err := s.mgr.ReadSourceStandalone(q.Get("module"), q.Get("file"), q.Get("path"), from, to)
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "source": res})
}

// hJobInfo 作业→实体程序/模块解析:GET /api/jobinfo?prog=&module= (不启动会话)
func (s *Server) hJobInfo(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	prog := strings.TrimSpace(q.Get("prog"))
	if prog == "" {
		fail(w, 400, fmt.Errorf("需要 prog 参数"))
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "job": s.mgr.ResolveJob(strings.TrimSpace(q.Get("module")), prog)})
}

func (s *Server) hLaunch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Module string `json:"module"`
		Prog   string `json:"prog"`
		SSH    string `json:"ssh"`  // 多 SSH 配置名(设置页维护);空 = 默认连接
		Zone   string `json:"zone"` // 区域覆盖(31/35/36/39/t);空 = 默认区域
	}
	if !readBody(w, r, &req) {
		return
	}
	if req.Prog == "" {
		fail(w, 400, fmt.Errorf("prog 必填(module 可留空,自动按作业名解析)"))
		return
	}
	// 按名换 SSH / 覆盖区域:克隆配置,默认链路零影响
	cfg := s.cfg
	if req.SSH != "" || req.Zone != "" {
		c2 := *s.cfg
		if req.SSH != "" {
			c2.SSH = s.cfg.SSHByName(req.SSH)
		}
		c2 = *c2.CloneWithZone(req.Zone)
		cfg = &c2
	}
	// 启动会结束当前这次调试,所以它同样受模式闸门约束(无会话或空会话放行)
	if !s.sessOfWriteGlobal(w, r) {
		return
	}
	sess, err := s.mgr.LaunchWith(cfg, req.Module, req.Prog)
	if err != nil {
		fail(w, 409, err)
		return
	}
	// 模式由发起方决定:AI 从 CLI 发起 → 协作(人只读);人从界面发起 → 纯人工(与旧行为一致)
	sess.SetMode(launchModeFor(r))
	if actorOf(r) == "ai" {
		s.mgr.emit(Event{
			Type: "ai_action", SessionID: sess.ID, Time: time.Now(),
			Actor: "ai", Action: "session.launch",
			Text: fmt.Sprintf("启动调试 %s/%s", req.Module, req.Prog),
		})
	}
	go func() {
		if err := sess.Launch(r.Context()); err != nil {
			log.Printf("[debug] 会话 %s(%s) 启动失败: %v", sess.ID, sess.Prog, err)
			s.mgr.emit(Event{Type: "log", SessionID: sess.ID, Text: "启动失败: " + err.Error()})
			if sess.Booted() {
				// 复用宿主上启动失败:Launch 已把会话收回空闲(可换作业重试/重启会话);
				// 仅在异常退回 exit 时移除,避免卡在 loading
				if sess.State() == StateExit {
					s.mgr.Remove(sess.ID)
				}
			} else {
				// 首次登录失败:宿主不可用,终结并移除,避免卡在 loading 挡住下一次启动
				sess.ForceExit()
				s.mgr.Remove(sess.ID)
			}
		}
	}()
	// runProg: gzzz_t 解析出的实体程序(前端源码命名/预取要用它,而不是作业编号)
	writeJSON(w, 200, map[string]any{"ok": true, "sessionId": sess.ID,
		"module": sess.Module, "prog": sess.Prog, "runProg": sess.RunProg})
}

func (s *Server) hList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"ok": true, "sessions": s.mgr.Snapshot()})
}

func (s *Server) hSnapshot(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	cur := sess.Cur()
	resp := map[string]any{
		"ok": true,
		"id": sess.ID, "module": sess.Module, "prog": sess.Prog, "runProg": sess.RunProg,
		"env":   sess.EnvName(),
		"state": string(sess.State()), "stop": cur,
		"started":         sess.Started(),
		"breakpoints":     sess.Breakpoints(),
		"holdingSeconds":  sess.HoldingSeconds(),
		"watchdogSeconds": s.cfg.WatchdogSeconds,
		"topent":          sess.TopentOverride(),
		"topentCfg":       sess.TopentCfg(),
		"topentShell":     sess.TopentShell(),
		// 本次重放跑完后读回来的响应原文(非重放会话为空)。
		// 以前想看它只能"停在收尾断点再 continue 到 exit"从 stdout 抓,很绕。
		"replayResponse": sess.ReplayResponse(),
		// 运行态:静默多久 / 已跑多久。配合 why 判断"是在等用户还是在空转"。
		"silentSeconds": sess.SilentSeconds(),
		// 谁在驾驶 + 正在执行哪条命令(界面上的"AI 正在执行 continue(已 12s)")
		"mode": sess.Mode(),
	}
	if fl := sess.Inflight(); fl != nil {
		resp["inflight"] = fl
	}
	if since := sess.RunningSince(); !since.IsZero() {
		resp["runningSince"] = since
	}
	if cur.WaitingForUser {
		resp["waitingForUser"] = true
		resp["waitingKind"] = cur.WaitingKind
	}
	writeJSON(w, 200, resp)
}

func (s *Server) hQuit(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	// 「结束调试」只结束本轮运行,宿主会话保留(idle);再次启动免重新登录。
	// 卡死降级整体断开时(StateExit)才移除会话。
	if err := sess.EndRun(); err != nil {
		fail(w, 500, err)
		return
	}
	s.emitAction(r, sess, "session.quit", "结束本轮调试")
	kept := sess.State() != StateExit
	if !kept {
		s.mgr.Remove(sess.ID)
	}
	writeJSON(w, 200, map[string]any{"ok": true, "kept": kept, "state": "idle"})
}

// ---------- 会话管理(单一常驻会话的切换/重启/结束) ----------

// envCfgFor 取指定环境的生效配置(CloneEnv);若与当前生效目标同名(或未配置 envs 列表)则用当前配置快照
func (s *Server) envCfgFor(name string) *Config {
	if c := s.cfg.CloneEnv(name); c != nil {
		return c
	}
	if name != "" && s.cfg.EnvName() == name {
		cc := *s.cfg
		return &cc
	}
	return nil
}

// selectEnv 把当前环境切到 cfg 所属环境并热生效(仅内存,不写 config.json)。
// 当前环境是运行时概念:由会话切换/CLI env 命令确定,服务重启后回到 sshs 首条。
func (s *Server) selectEnv(cfg *Config) {
	if cfg != nil {
		*s.cfg = *cfg
	}
}

// bootToIdle 异步登录新会话到 idle(切换/重启用),失败终结并移除
func (s *Server) bootToIdle(ns *Session, ctx context.Context) {
	if err := ns.BootIdle(ctx); err != nil {
		log.Printf("[debug] 会话 %s(%s) 登录失败: %v", ns.ID, ns.EnvName(), err)
		s.mgr.emit(Event{Type: "log", SessionID: ns.ID, Text: "会话连接失败: " + err.Error()})
		ns.ForceExit()
		s.mgr.Remove(ns.ID)
	}
}

// hSessionClose 「结束会话」:结束当前 debug 并彻底断开连接(与设置默认解耦)
func (s *Server) hSessionClose(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	if err := sess.Close(); err != nil {
		fail(w, 500, err)
		return
	}
	s.emitAction(r, sess, "session.close", "关闭会话")
	s.mgr.Remove(sess.ID)
	writeJSON(w, 200, map[string]any{"ok": true, "state": "exit"})
}

// hSessionRestart 「重启会话」:断开并重连同环境,回到 idle 待启动
func (s *Server) hSessionRestart(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	env := sess.EnvName()
	id := sess.ID
	// 先收口当前会话(运行中会先中断/quit),再重建连接
	if err := sess.Close(); err != nil {
		fail(w, 500, err)
		return
	}
	s.mgr.Remove(id)
	cfg := s.envCfgFor(env)
	if cfg == nil {
		cfg = s.envCfgFor(s.cfg.EnvName()) // 原环境已被删除:退回当前默认
	}
	if cfg == nil {
		fail(w, 500, fmt.Errorf("找不到可用的连接配置(环境 %q 已被删除)", env))
		return
	}
	ns, err := s.mgr.CreateSessionOn(cfg)
	if err != nil {
		fail(w, 502, err)
		return
	}
	s.mgr.emit(Event{Type: "log", Text: fmt.Sprintf("重启会话:重新连接 %s …", cfg.EnvName())})
	s.emitAction(r, ns, "session.restart", "重新开始调试")
	go s.bootToIdle(ns, r.Context())
	writeJSON(w, 200, map[string]any{"ok": true, "sessionId": ns.ID, "env": cfg.EnvName(), "state": "loading"})
}

// hTopent 空闲态重新设置 TOPENT(会话内,立即下发到 shell)。
// 仅允许 idle(宿主 shell 就绪、无调试运行);值不限数字/文本(导出为环境变量),
// 服务端剔除两侧空白;留空 = 清除手动覆盖并重新下发配置默认。
func (s *Server) hTopent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Value string `json:"value"`
	}
	if !readBody(w, r, &req) {
		return
	}
	req.Value = trimTopent(req.Value)
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	if st := sess.State(); st != StateIdle {
		fail(w, 409, fmt.Errorf("仅会话空闲时可设置 TOPENT(当前 %s),请先结束当前调试", st))
		return
	}
	if err := sess.SetTopent(req.Value); err != nil {
		fail(w, 500, err)
		return
	}
	txt := "设置 TOPENT"
	if req.Value == "" {
		txt += "(清除)"
	} else {
		txt += "=" + req.Value
	}
	s.emitAction(r, sess, "session.topent", txt)
	writeJSON(w, 200, map[string]any{"ok": true, "topent": req.Value})
}

// trimTopent TOPENT 归一:剔除两侧空白(值不限数字/文本);全空白归一为空(清除)
func trimTopent(v string) string {
	return strings.TrimSpace(v)
}

// hSessionSwitch 「切换会话」:把当前环境切到目标环境(仅内存)并按该环境重连到 idle。
// 若已是目标环境会话则幂等返回(不改动正在进行的调试)。
func (s *Server) hSessionSwitch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Env string `json:"env"`
	}
	if !readBody(w, r, &req) {
		return
	}
	cfg := s.envCfgFor(req.Env)
	if cfg == nil {
		fail(w, 404, fmt.Errorf("环境 %q 不存在(请在设置中添加)", req.Env))
		return
	}
	if !s.sessOfWriteGlobal(w, r) {
		return
	}
	if s.cfg.EnvName() != req.Env {
		s.selectEnv(cfg)
	}
	cur := s.mgr.Current()
	if cur != nil {
		if cur.sameTarget(cfg) {
			writeJSON(w, 200, map[string]any{"ok": true, "sessionId": cur.ID, "env": req.Env, "state": string(cur.State())})
			return
		}
		s.mgr.emit(Event{Type: "log", SessionID: cur.ID, Text: fmt.Sprintf("切换会话到 %s,断开当前会话", req.Env)})
		_ = cur.Close()
		s.mgr.Remove(cur.ID)
	}
	ns, err := s.mgr.CreateSessionOn(cfg)
	if err != nil {
		fail(w, 502, err)
		return
	}
	s.mgr.emit(Event{Type: "log", Text: fmt.Sprintf("切换会话:正在连接 %s …", req.Env)})
	s.emitAction(r, ns, "session.switch", "切换环境到 "+req.Env)
	go s.bootToIdle(ns, r.Context())
	writeJSON(w, 200, map[string]any{"ok": true, "sessionId": ns.ID, "env": req.Env, "state": "loading"})
}

// ---------- 断点 ----------

func (s *Server) hBPList(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "breakpoints": sess.Breakpoints()})
}

func (s *Server) hBPAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Location string `json:"location"` // 行号 / 函数名 / file:line
	}
	if !readBody(w, r, &req) {
		return
	}
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	log.Println("[bp-add] session=" + sess.ID + " loc=" + req.Location)
	bp, err := sess.Break(req.Location)
	if err != nil {
		fail(w, 400, err)
		return
	}
	log.Printf("[bp-add] result #%d %s:%d", bp.Num, bp.File, bp.Line)
	// 用 fgldb 落定后的实际位置(它会把非可执行行上的断点下移到下一条语句)
	s.emitAction(r, sess, "bp.add", fmt.Sprintf("下断点 %s:%d", bp.File, bp.Line))
	writeJSON(w, 200, map[string]any{"ok": true, "breakpoint": bp})
}

func (s *Server) hBPDel(w http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(r.PathValue("num"))
	if err != nil {
		fail(w, 400, fmt.Errorf("断点编号无效"))
		return
	}
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	log.Println("[bp-del] session=" + sess.ID + " num=" + strconv.Itoa(num))
	if err := sess.DeleteBreakpoint(num); err != nil {
		fail(w, 400, err)
		return
	}
	s.emitAction(r, sess, "bp.del", fmt.Sprintf("删除断点 %d", num))
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ---------- 控制 / 求值 ----------

func (s *Server) hControl(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string `json:"action"` // continue|run|next|step|finish|until|interrupt
		Arg    string `json:"arg,omitempty"`
		// Wait 软等待秒数(仅步进类):到点若程序仍在跑就返回 softTimeout。
		// 步进可能撞上交互语句而长时间等用户,需要这条退路;continue/run 本来就立即返回。
		Wait int `json:"wait"`
	}
	if !readBody(w, r, &req) {
		return
	}
	// 闸门放在读 body 之后:这样归因文案能带上动作细节(continue / step 12 …)
	txt := req.Action
	if req.Arg != "" {
		txt += " " + req.Arg
	}
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	softWait := time.Duration(req.Wait) * time.Second
	if softWait < 0 {
		softWait = 0
	}
	var stop *StopInfo
	var soft bool
	var err error
	switch req.Action {
	case "continue":
		_, err = sess.Continue()
	case "run":
		_, err = sess.Run()
	case "next", "step", "finish":
		stop, soft, err = sess.StepSoft(req.Action, softWait)
	case "until":
		loc := req.Arg
		if loc == "" {
			stop, soft, err = sess.StepSoft("until", softWait)
		} else {
			stop, soft, err = sess.StepSoft("until "+loc, softWait)
		}
	case "interrupt":
		err = sess.Interrupt()
	default:
		fail(w, 400, fmt.Errorf("未知 action: %s", req.Action))
		return
	}
	if err != nil {
		fail(w, 409, err)
		return
	}
	s.emitAction(r, sess, "control."+req.Action, txt)
	resp := map[string]any{"ok": true, "state": string(sess.State())}
	if stop != nil {
		resp["stop"] = stop
	}
	if soft {
		resp["softTimeout"] = true
		resp["silentSeconds"] = sess.SilentSeconds()
		resp["hint"] = "步进到点仍未停站:程序仍在运行(大概率停在了交互界面上等用户)。" +
			"用 why 判断,或 wait 等它停下来"
	}
	writeJSON(w, 200, resp)
}

func (s *Server) hPrint(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	var req struct {
		Expr string `json:"expr"`
	}
	if !readBody(w, r, &req) {
		return
	}
	v, err := sess.Print(req.Expr)
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "value": v})
}

func (s *Server) hWhere(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	frames, err := sess.Where()
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "frames": frames})
}

func (s *Server) hRaw(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Command string `json:"command"`
		Timeout int    `json:"timeout"` // 等待命令完成的秒数(0=默认30);continue/until 等长命令可调大
		// Wait 软等待秒数:到点若程序仍在跑就直接返回(softTimeout),**不发 SIGINT、不取消命令**。
		// 与 timeout 的区别是硬/软:timeout 到点会发 \x03 探测并报错,wait 到点只是"不再等"。
		Wait int `json:"wait"`
	}
	if !readBody(w, r, &req) {
		return
	}
	cmd := strings.TrimSpace(req.Command)
	if cmd == "" || strings.ContainsAny(cmd, "\r\n") {
		fail(w, 400, fmt.Errorf("命令不能为空或含换行"))
		return
	}
	lower := strings.ToLower(cmd)
	if lower == "quit" || strings.HasPrefix(lower, "quit ") {
		fail(w, 400, fmt.Errorf("请使用会话控制接口执行 quit"))
		return
	}
	// 闸门放在入参与合法性校验之后:被拒的命令不该在时间线上留痕
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	// run / 入口停站时的 continue 有会话级语义,不能当裸命令透传:
	// fgldb 在 run 之前不接受 continue(会回 "The program is not being run"),
	// 而会话层的 Run()/Continue() 对这个状态有兜底(Continue 等效成 Run)。
	// 不分流的话**协作模式会死锁**:人按不了「继续」(模式闸门),AI 又发不出 run,
	// 谁都启动不了程序。Web 界面走的是 /control,两边落到同一实现。
	if lower == "run" || (lower == "continue" && !sess.Started()) {
		var er error
		if lower == "run" {
			_, er = sess.Run()
		} else {
			_, er = sess.Continue()
		}
		if er != nil {
			fail(w, 400, er)
			return
		}
		writeJSON(w, 200, map[string]any{
			"ok": true, "lines": []string{cmd},
			"state": string(sess.State()), "started": sess.Started(),
			"hint": "程序已放行,用 tt debug wait --for stopped 等它停下来",
		})
		return
	}
	// 入口态(程序还没 run)的 next/step/until/finish:fgldb 在 run 之前**不受理**这些命令,
	// 只回一句 "The program is not being run."。以前这里直接把它们当裸命令透传,后果是
	// 会话被卡死:execOpt 见"放行类命令"就把状态乐观翻成 running 且无回滚,之后
	// interrupt 因"未启动"被拒、exec 因"运行中"被拒、EndRun 的中断同样失败 ——
	// 只能整体断开会话重来。真机踩过。
	//
	// 会话语义上这几条本来是有办法的:stepFromEntry(tbreak main + run)。
	// /control 一直走的就是它,只有 /raw 漏了 —— 这里补齐,两条路行为对齐。
	if !sess.Started() && lower != "run" && lower != "continue" && IsResumeCmd(cmd) {
		soft := time.Duration(req.Wait) * time.Second
		if soft < 0 {
			soft = 0
		}
		st, completed, er := sess.StepSoft(cmd, soft)
		if er != nil {
			fail(w, 400, er)
			return
		}
		resp := map[string]any{
			"ok": true, "lines": []string{cmd},
			"state": string(sess.State()), "started": sess.Started(),
			"hint": "入口停站没有调用栈,step/next 会先 tbreak main 把程序放起来;" +
				"用 tt debug wait --for stopped 等它停到 MAIN",
		}
		if !completed && st == nil {
			resp["softTimeout"] = true
			resp["hint"] = "程序已放行但还没停站;用 tt debug wait --for stopped 等它停下来"
		} else if st != nil {
			resp["stop"] = st
		}
		writeJSON(w, 200, resp)
		return
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30
	}
	if timeout > 600 {
		timeout = 600
	}
	soft := time.Duration(req.Wait) * time.Second
	if soft < 0 {
		soft = 0
	}
	// 软等待必须短于硬超时,否则永远轮不到它
	if soft > 0 && soft >= time.Duration(timeout)*time.Second {
		soft = time.Duration(timeout-1) * time.Second
	}
	res, err := sess.RawSoft(cmd, time.Duration(timeout)*time.Second, soft)
	if err != nil {
		fail(w, 400, err)
		return
	}
	// 断点类裸命令:命令本身已经执行完了,但会话的断点缓存还没跟上 —— 见 IsBreakpointCmd
	// 的注释。/raw 走的 kind 是 "other",onLine 里那条按 kind 分派的解析不会进,
	// 所以必须在这里用 info breakpoints 回灌一次,否则这个断点在快照/持久化里都不存在。
	act := "raw"
	if IsBreakpointCmd(cmd) {
		sess.syncBreakpoints()
		act = "bp." + strings.ToLower(strings.Fields(cmd)[0])
	}
	s.emitAction(r, sess, act, cmd)
	resp := map[string]any{"ok": true, "lines": res.Lines}
	// 会话层就截断过(行数/单值超上限)要如实说,否则调用方会把残缺内容当成全部
	if res.Truncated {
		resp["truncated"] = true
		resp["truncReason"] = res.TruncReason
	}
	// 输出够大就落一份**完整**的本地副本并回报路径。
	// 理由:正文有上限(超了只给头部),但完整内容必须留得住 —— 否则想多看一点就得重跑命令,
	// 而重跑会改变现场。落本地后可以随时 grep/整读,零往返。
	// 阈值是为了别给每条小命令都留一个文件。
	if n := len(res.Lines); n > execLogMinLines || linesBytes(res.Lines) > execLogMinBytes {
		env := mirrorEnvSeg(s.cfg.EnvName(), s.cfg.SSH.Host, s.cfg.Zone)
		if p := writeExecLog(s.cfg.DataDir, env, execLogName(cmd), []byte(strings.Join(res.Lines, "\n"))); p != "" {
			resp["localPath"] = p
			resp["totalLines"] = n
		}
	}
	if res.SoftTimeout {
		// 不是错误:命令仍在飞,程序仍在跑
		resp["softTimeout"] = true
		resp["state"] = string(sess.State())
		resp["silentSeconds"] = sess.SilentSeconds()
		resp["hint"] = "程序仍在运行:命令没有被取消,也没有发 SIGINT。" +
			"用 tt debug wait 等停站(或 POST /api/sessions/{id}/wait)," +
			"用 tt debug why 判断它在等用户还是在空转"
	}
	writeJSON(w, 200, resp)
}

// ---------- 上下文查询 ----------

// hWhy 探测「程序此刻到底在干什么」——在等用户操作,还是在空转。
// 把 SKILL.md 里教 AI 人肉做的三步(interrupt → where → 看停在哪一行)收成一次调用。
// 已停站时只做分类,不发任何命令(零副作用);运行中才发 SIGINT。
// body: {"resume": true} —— 默认 true(探测完自动放回运行,免得把等用户的程序挂住)。
//
// 闸门上是**条件写**:只有运行中(会发 SIGINT)才受模式闸门约束;已停站时纯粹是
// 只读分类,放行给任一方 —— 否则人在协作模式下连"现在什么情况"都问不了。
func (s *Server) hWhy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Resume *bool `json:"resume"`
	}
	if !readBody(w, r, &req) {
		return
	}
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	if sess.State() == StateRunning {
		if sess = s.sessOfWrite(w, r); sess == nil {
			return
		}
	}
	resume := true
	if req.Resume != nil {
		resume = *req.Resume
	}
	res, err := sess.Why(resume)
	if err != nil {
		fail(w, 409, err)
		return
	}
	s.emitAction(r, sess, "why", "探测程序在等用户还是空转")
	writeJSON(w, 200, map[string]any{"ok": true, "why": res})
}

// hWait 长轮询:阻塞到会话出现指定事件或超时。
// GET /api/sessions/{id}/wait?for=stopped,exit,dead,watchdog&timeout=300
//
// 手法是「先订阅、后判现状」:订阅动作与读快照之间的窗口由快照本身补齐 ——
// 谓词只看权威现状(State()/Cur()),不看事件历史,所以历史丢失无关紧要,
// **hWait 因此不依赖单调序号**。(Event.Seq 是为 WS 开场补发与前端去重的
// **游标**语义引入的,与此处的**栅栏**判断互不冲突。)
// **超时不是错误**:返回 200 + timedOut:true + 当前快照。
func (s *Server) hWait(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	want := map[string]bool{}
	for _, t := range strings.Split(r.URL.Query().Get("for"), ",") {
		if t = strings.TrimSpace(t); t != "" {
			want[t] = true
		}
	}
	if len(want) == 0 {
		for _, t := range []string{"stopped", "exit", "dead", "watchdog"} {
			want[t] = true
		}
	}
	to := 300 * time.Second
	if v := r.URL.Query().Get("timeout"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			to = time.Duration(n) * time.Second
		}
	}
	if to > time.Hour {
		to = time.Hour
	}

	// 先订阅,后判现状:反向顺序会漏掉"订阅建立前刚发生"的事件
	ch, cancel := s.mgr.Subscribe("wait")
	defer cancel() // 必须 defer,否则 m.subs 泄漏

	if ev, ok := waitNow(sess, want); ok {
		writeJSON(w, 200, waitResp(sess, ev, false))
		return
	}

	timer := time.NewTimer(to)
	defer timer.Stop()
	for {
		select {
		case <-r.Context().Done():
			return // 客户端断开:不写响应
		case ev := <-ch:
			if ev.SessionID != "" && ev.SessionID != sess.ID {
				continue // 别条会话的事件
			}
			if waitMatch(ev, want) {
				writeJSON(w, 200, waitResp(sess, &ev, false))
				return
			}
		case <-timer.C:
			writeJSON(w, 200, waitResp(sess, nil, true))
			return
		}
	}
}

// waitMatch 事件是否命中等待条件。
// **注意**:stopped / exit / idle 在事件流里也可能是 `state` 事件的取值,不是独立类型 ——
// 尤其**入口停站只发 `state(stopped)`,不发 `stopped` 事件**(见 startRun),
// 不做这个映射就会漏掉它,tt debug start 会一直等到超时。
func waitMatch(ev Event, want map[string]bool) bool {
	if want[ev.Type] {
		return true
	}
	if ev.Type == "state" && ev.State != "" {
		if want[ev.State] {
			return true
		}
		if (ev.State == string(StateExit) || ev.State == string(StateIdle)) && (want["exit"] || want["dead"]) {
			return true
		}
	}
	return false
}

// waitNow 把"权威现状"映射成一条合成事件;不满足等待条件时返回 false。
func waitNow(sess *Session, want map[string]bool) (*Event, bool) {
	switch sess.State() {
	case StateStopped:
		if want["stopped"] {
			st := sess.Cur()
			return &Event{Type: "stopped", SessionID: sess.ID, Stop: &st}, true
		}
	case StateExit, StateIdle:
		if want["exit"] || want["dead"] {
			return &Event{Type: "exit", SessionID: sess.ID}, true
		}
	}
	return nil, false
}

func waitResp(sess *Session, ev *Event, timedOut bool) map[string]any {
	resp := map[string]any{
		"ok":            true,
		"state":         string(sess.State()),
		"silentSeconds": sess.SilentSeconds(),
	}
	if timedOut {
		resp["timedOut"] = true
	}
	if ev != nil {
		resp["event"] = ev
		if ev.Stop != nil {
			resp["stop"] = ev.Stop
		}
	}
	return resp
}

func (s *Server) hLocals(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	vars, err := sess.Locals()
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "vars": vars})
}

func (s *Server) hGlobals(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	vars, total, err := sess.Globals(limit)
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "vars": vars, "total": total})
}

// hLocate POST /api/sessions/{id}/locate {word}:定位函数到源文件与行号。
// 用 fgldb info line [module.]function(BDL 文档语法),仅停站状态可用;Ctrl+点击跳函数用
func (s *Server) hLocate(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	var req struct {
		Word string `json:"word"`
	}
	if !readBody(w, r, &req) {
		return
	}
	word := strings.TrimSpace(req.Word)
	if !reIdent.MatchString(word) {
		fail(w, 400, fmt.Errorf("函数名非法: %q", word))
		return
	}
	file, line, err := sess.InfoLine(word)
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "file": file, "line": line})
}

// hCalibrate 行号校准:停站后检测 fgldb(DVM)行号与磁盘源码的偏移量,
// 供前端把 Monaco 行号对齐到协议流。POST /api/sessions/{id}/calibrate
func (s *Server) hCalibrate(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	offset, err := sess.CalibrateOffset()
	if err != nil {
		fail(w, 400, err)
		return
	}
	log.Printf("[calibrate] session=%s offset=%d", sess.ID, offset)
	s.emitAction(r, sess, "session.calibrate", fmt.Sprintf("行号校准(偏移 %d)", offset))
	writeJSON(w, 200, map[string]any{"ok": true, "offset": offset})
}

func (s *Server) hSources(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	sources, err := sess.Sources()
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "sources": sources})
}

func (s *Server) hFunctions(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	fns, total, err := sess.Functions(limit)
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "functions": fns, "total": total})
}

func (s *Server) hAutovars(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "vars": sess.Autovars()})
}

// hAutovarsSet POST /api/sessions/{id}/autovars {auto}:开关停站后自动求值自动变量
// (默认关;开启且已停站时立即补一次求值,结果照常以 autovars 事件推送)
func (s *Server) hAutovarsSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Auto *bool `json:"auto"`
	}
	if !readBody(w, r, &req) || req.Auto == nil {
		return
	}
	verb := "关闭"
	if *req.Auto {
		verb = "开启"
	}
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	sess.SetAutovarsOn(*req.Auto)
	s.emitAction(r, sess, "session.autovars", verb+"自动变量求值")
	writeJSON(w, 200, map[string]any{"ok": true, "auto": *req.Auto})
}

func (s *Server) hFrame(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Num int `json:"num"`
	}
	if !readBody(w, r, &req) {
		return
	}
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	if err := sess.Frame(req.Num); err != nil {
		fail(w, 400, err)
		return
	}
	s.emitAction(r, sess, "session.frame", fmt.Sprintf("选帧 #%d", sess.CurFrame()))
	writeJSON(w, 200, map[string]any{"ok": true, "frame": sess.CurFrame()})
}

func (s *Server) hBPEnabled(w http.ResponseWriter, r *http.Request) {
	num, err := strconv.Atoi(r.PathValue("num"))
	if err != nil {
		fail(w, 400, fmt.Errorf("断点编号无效"))
		return
	}
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if !readBody(w, r, &req) || req.Enabled == nil {
		return
	}
	verb := "停用"
	if *req.Enabled {
		verb = "启用"
	}
	sess := s.sessOfWrite(w, r)
	if sess == nil {
		return
	}
	if err := sess.SetBPEnabled(num, *req.Enabled); err != nil {
		fail(w, 400, err)
		return
	}
	s.emitAction(r, sess, "bp.enable", fmt.Sprintf("%s断点 %d", verb, num))
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ---------- 源码 ----------

// hSourcePreview 会话建立前预取源码(消除启动期空白):GET /api/source-preview?module=&prog=
func (s *Server) hSourcePreview(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	module, prog := q.Get("module"), q.Get("prog")
	if module == "" || prog == "" {
		fail(w, 400, fmt.Errorf("需要 module 与 prog 参数"))
		return
	}
	sf, err := s.mgr.SourcePreview(module, prog)
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "source": sf})
}

func (s *Server) hSource(w http.ResponseWriter, r *http.Request) {
	sess := s.sessOf(w, r)
	if sess == nil {
		return
	}
	q := r.URL.Query()
	var sf *SourceFile
	var err error
	switch {
	case q.Get("path") != "":
		sf, err = sess.ReadPath(q.Get("path"))
	case q.Get("file") != "":
		sf, err = sess.ResolveSource(q.Get("file"), q.Get("module"))
	default:
		err = fmt.Errorf("需要 file 或 path 参数")
	}
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "source": sf})
}

// ---------- 接口服务测试(复刻 awsq990 集成服务测试) ----------

// hWSTest POST /api/wstest {mode,url,body,soap} → 经服务器 curl 调用接口
func (s *Server) hWSTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mode string `json:"mode"`
		URL  string `json:"url"`
		Body string `json:"body"`
		Soap bool   `json:"soap"`
	}
	if !readBody(w, r, &req) {
		return
	}
	if req.URL == "" {
		req.URL = WSDefaultURLFor(s.cfg, req.Mode)
	}
	conn, err := s.dial(s.cfg.SSH)
	if err != nil {
		fail(w, 500, fmt.Errorf("SSH 连接失败: %w", err))
		return
	}
	defer conn.Close()
	res, err := WSTest(conn, req.URL, req.Body, req.Soap, 60)
	if err != nil {
		fail(w, 400, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "result": res})
}

// ---------- 接口日志(wsfa_t) ----------

// wsZone 返回查询用 zone(空配置默认 36)
func (s *Server) wsZone() string {
	if s.cfg.Zone == "" {
		return "36"
	}
	return s.cfg.Zone
}

// hWSLogs GET /api/wslogs?service=&result=&origin=&server=&onlyFail=1&page=&pageSize=&startFrom=&endTo=
// 时间窗对齐原生 awsq990:startFrom 是 wsfa003(起始时间)下界,endTo 是 wsfa004(结束时间)上界
func (s *Server) hWSLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	conn, err := s.dial(s.cfg.SSH)
	if err != nil {
		fail(w, 500, fmt.Errorf("SSH 连接失败: %w", err))
		return
	}
	defer conn.Close()
	dbc, err := resolveDBRun(conn, s.cfg)
	if err != nil {
		fail(w, 500, err)
		return
	}
	items, hasMore, err := listWSLogs(conn, dbc, WSLogFilter{
		Service:   q.Get("service"),
		Job:       q.Get("job"),
		Result:    q.Get("result"),
		Origin:    q.Get("origin"),
		Server:    q.Get("server"),
		PID:       q.Get("pid"),
		OnlyFail:  q.Get("onlyFail") == "1",
		StartFrom: q.Get("startFrom"),
		EndTo:     q.Get("endTo"),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		fail(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "items": items, "hasMore": hasMore,
		"acct": wslogAcctFor(s.wslogTopent())})
}

// hWSLogContent GET /api/wslogs/content?rowid=
func (s *Server) hWSLogContent(w http.ResponseWriter, r *http.Request) {
	rowid := r.URL.Query().Get("rowid")
	conn, err := s.dial(s.cfg.SSH)
	if err != nil {
		fail(w, 500, fmt.Errorf("SSH 连接失败: %w", err))
		return
	}
	defer conn.Close()
	dbc, err := resolveDBRun(conn, s.cfg)
	if err != nil {
		fail(w, 500, err)
		return
	}
	item, content, err := WSLogDetail(conn, dbc, rowid)
	if err != nil {
		fail(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "item": item, "content": content,
		"acct": wslogAcctFor(s.wslogTopent())})
}

// hWSLogDebug POST /api/wslogs/debug {rowid, request?} → 用该日志的报文重放调试。
// request 非空 = 界面里改过的入参(落服务器临时文件后作为入参文件),空 = 用日志原报文。
func (s *Server) hWSLogDebug(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RowID   string `json:"rowid"`
		Request string `json:"request"`
	}
	if !readBody(w, r, &req) {
		return
	}
	// 入参先校验:非法请求不该先把当前会话收口(旧顺序是收口后才在 LaunchReplay 里失败)
	if err := checkReplayOverride(req.Request); err != nil {
		fail(w, 400, err)
		return
	}
	if !s.sessOfWriteGlobal(w, r) {
		return
	}
	conn, err := s.dial(s.cfg.SSH)
	if err != nil {
		fail(w, 500, fmt.Errorf("SSH 连接失败: %w", err))
		return
	}
	dbc, err := resolveDBRun(conn, s.cfg)
	if err != nil {
		conn.Close()
		fail(w, 500, err)
		return
	}
	item, content, err := WSLogDetail(conn, dbc, req.RowID)
	conn.Close()
	if err != nil {
		fail(w, 500, err)
		return
	}
	// 已有会话先收口:同目标(可复用空闲宿主)只结束本轮运行;不同目标直接断开。
	// 重放调试独占调试通道,旧会话不再有用。
	if cur := s.mgr.Current(); cur != nil {
		if cur.sameTarget(s.cfg) {
			s.mgr.emit(Event{Type: "log", SessionID: cur.ID, Text: "重放调试启动,结束当前调试(会话保留)"})
			_ = cur.EndRun()
			// EndRun 卡死时降级为 Close(StateExit),残留会话由 prepareSession 清扫
		} else {
			s.mgr.emit(Event{Type: "log", SessionID: cur.ID, Text: "重放调试启动,断开当前会话"})
			_ = cur.Close()
			s.mgr.Remove(cur.ID)
		}
	}
	sess, replayWarn, err := s.mgr.LaunchReplay(item, content, req.Request)
	if err != nil {
		fail(w, 409, err)
		return
	}
	// 模式与 hLaunch 同规则:谁发起谁驾驶。重放同样是 AI 从 CLI 发起的,
	// 漏掉这一句会让 AI 连自己刚启动的会话都写不了(quit 都会被 403 拦下)。
	sess.SetMode(launchModeFor(r))
	go func() {
		if err := sess.Launch(r.Context()); err != nil {
			s.mgr.emit(Event{Type: "log", SessionID: sess.ID, Text: "启动失败: " + err.Error()})
			if sess.Booted() {
				if sess.State() == StateExit {
					s.mgr.Remove(sess.ID)
				}
			} else {
				sess.ForceExit()
				s.mgr.Remove(sess.ID)
			}
		}
	}()
	s.emitAction(r, sess, "session.replay", "按接口日志重放调试")
	// warn 非空 = 这次重放用的是入库的截断文本,结果可能不可信(见 WriteReplayFiles)。
	// 随回包一起给出去,让 CLI 能当场说清,而不是等人自己去翻事件日志。
	writeJSON(w, 200, map[string]any{"ok": true, "sessionId": sess.ID,
		"module": sess.Module, "prog": sess.Prog, "runProg": sess.RunProg, "warn": replayWarn,
		"acct": wslogAcctFor(s.wslogTopent())})
}

// wslogTopent 取"此刻的 TOPENT 是什么"。
//
// **只用于说明** —— 查 wsfa_t 恒用系统账号 ds(见 wslogAcctFor),这个值不影响用哪个账号。
// 优先会话里的生效值(会话级覆盖 > 环境配置),没有会话就用环境配置。
func (s *Server) wslogTopent() string {
	if cur := s.mgr.Current(); cur != nil {
		if v := cur.TopentOverride(); v != "" {
			return v
		}
	}
	return string(s.cfg.Topent)
}

// hDBSQL 只读 SQL:POST /api/dbsql {sql, ent, timeout}
//
// **不过模式闸门** —— 与 /api/wslogs 同理:模式闸门管的是"谁能写调试会话",
// 而这条不碰会话、不移动程序位置,只是读库。人与 AI 都放行。
// 它有自己的配置闸门(db.readonlySql,默认开),那才是这个能力的开关。
//
// 归因用 Type=log(不是 ai_action —— 那个的语义是"AI 做了写操作",只读不该占用它),
// 且记在**执行成功之后**,与 emitAction 的时机一致。
func (s *Server) hDBSQL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SQL     string `json:"sql"`
		Ent     int    `json:"ent"`     // 0 = 用当前会话的 TOPENT,否则配置默认
		Timeout int    `json:"timeout"` // 秒
	}
	if !readBody(w, r, &req) {
		return
	}
	// 企业编号:优先当前会话的 override(tt debug topent 设的那个),再退到配置默认。
	// 不做成"必须挂会话"是为了让它独立可用;但只要会话在,就一定跟会话走 ——
	// 否则会出现"调试的是这个企业、查库查的是另一个"这种最难查的错。
	ent := req.Ent
	if ent <= 0 {
		if cur := s.mgr.Current(); cur != nil {
			ent = cur.TopentIntForDB()
		}
	}
	conn, err := s.dial(s.cfg.SSH)
	if err != nil {
		fail(w, 500, fmt.Errorf("SSH 连接失败: %w", err))
		return
	}
	defer conn.Close()
	res, err := s.mgr.RunReadonlyQuery(conn, ReadonlyQueryReq{SQL: req.SQL, Ent: ent, Timeout: req.Timeout})
	if err != nil {
		fail(w, 400, err)
		return
	}
	if actorOf(r) == "ai" {
		text := "只读 SQL(企业 " + fmt.Sprint(res.Ent) + " → " + res.Account + "): " + firstLine(req.SQL, 120)
		s.mgr.emit(Event{Type: "log", Actor: "ai", Action: "sql", Text: text})
	}
	writeJSON(w, 200, map[string]any{"ok": true, "result": res})
}

// firstLine 取首行并截断(时间线与日志里只留个线索,不留全文)
func firstLine(s string, max int) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	if len(s) > max {
		s = s[:max] + "…"
	}
	return s
}

// 命令输出落本地的阈值:太小会给每条命令都留一个文件,没必要
const (
	execLogMinLines = 200
	execLogMinBytes = 16 << 10
)

func linesBytes(ls []string) int {
	n := 0
	for _, s := range ls {
		n += len(s) + 1
	}
	return n
}

// execLogName 落盘文件名:时间戳 + 命令摘要,便于人肉对照这是哪个命令的输出。
// 摘要按 rune 走、用字节数封顶,别把中文劈成半个。
func execLogName(cmd string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(cmd) {
		switch {
		case r == ' ' || r == '\t':
			b.WriteRune('_')
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
		if b.Len() >= 60 {
			break
		}
	}
	return time.Now().Format("20060102-150405.000") + "-" + pathSafeSeg(b.String()) + ".txt"
}

// ---------- 设置 / 环境清单 ----------

// hHostsGet 返回环境清单。独立运行(tt debug serve)时给 tt debug env 读,
// 字段与统一服务的顶层 GET /api/hosts 一致(那边由 tt/internal/web 实现)。
//
// 合并前这条读的是 /api/settings,它吐的是整个 debug 配置节 —— 那时节既装
// 环境清单(debug.sshs)又装调试设置。拆节之后环境清单归 hosts、设置归 debug,
// 页面统一走顶层 /api/hosts;这里只为独立运行保留一份只读视图。
func (s *Server) hHostsGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{
		"config":    s.cfgPath,
		"activeEnv": s.cfg.envName, // 当前生效环境;空 = 按 hosts.activeEnv 取首条
		"sshs":      s.cfg.SSHs,
		"listen":    s.cfg.Listen,
		"debug":     debugSettingsOf(s.cfg),
	})
}

// ReloadConfig 重新从磁盘读取配置并热替换内存态。
//
// 统一服务(tt serve)的共享配置端点 /api/hosts 在写入后会调用它:统一设置页
// 改的就是同一份 hosts 节,不重读的话本服务还按旧环境连,表现就是"改了没生效"。
//
// 三条约定:
//   - 并发安全:整份配置的替换在 s.mu 里做,不会让读者看到半个新配置;
//   - 幂等:重复调用只是又读一遍同样的文件;
//   - 失败不动现场:文件缺失/格式错误时直接返回错误,内存里的 *s.cfg 原样保留 ——
//     别处一次坏写不该把正在跑的调试会话的配置清空。
func (s *Server) ReloadConfig() error {
	if s.cfgPath == "" {
		return fmt.Errorf("服务未挂接配置文件路径,无法重新读取配置")
	}
	cfg, err := LoadConfigAllowEmpty(s.cfgPath)
	if err != nil {
		return err
	}
	// 运行时字段不随文件走:数据目录由服务注入,当前环境由会话选定 ——
	// 保留它,设置页改一次 hosts 就不会把正在调试的会话悄悄切到默认环境。
	cfg.DataDir = s.cfg.DataDir
	if cur := s.cfg.envName; cur != "" {
		cfg.envName = cur
	}
	cfg.ApplyDefaultEnv() // 当前环境的连接/参数合并到运行时字段
	cfg.fillDefaults()
	s.swapConfig(cfg)
	return nil
}

// swapConfig 整份替换内存配置。配置写入路径与 ReloadConfig 共用同一把锁。
func (s *Server) swapConfig(cfg *Config) {
	s.mu.Lock()
	*s.cfg = *cfg
	s.mu.Unlock()
}

// ---------- WebSocket ----------

// wsReplayLimit WS 建连时补发多少条历史(对齐前端时间线的容量)
const wsReplayLimit = 500

// replayTypes 只补发"时间线类"事件:state/dead 由页面自己的快照拉取与状态同步负责,
// 混进补发批次反而可能让前端拿旧状态覆盖新状态。
var replayTypes = map[string]bool{"stopped": true, "ai_action": true, "log": true, "watchdog": true}

func (s *Server) hWS(w http.ResponseWriter, r *http.Request) {
	ch, replay, cancel := s.mgr.SubscribeWithReplay("ws", wsReplayLimit)
	defer cancel()
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	// ctx 与连接同生命周期:下面那个丢弃式 reader 一旦读到错误就取消它。
	ctx, cancelConn := context.WithCancel(r.Context())
	defer cancelConn()

	// 必须有人读:否则库处理不了对端的 close/ping 帧 —— 客户端关闭时服务端要干等到
	// 超时才发现,代理上的 ping 也得不到回应(这条通道是纯推送,收到的都丢掉)。
	go func() {
		defer cancelConn()
		for {
			if _, _, err := conn.Read(ctx); err != nil {
				return
			}
		}
	}()

	writeFrame := func(v any) error {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		wctx, wcancel := context.WithTimeout(ctx, 5*time.Second)
		defer wcancel()
		return conn.Write(wctx, websocket.MessageText, data)
	}

	// 开场补发:页面刷新后时间线不该是空的(而"看着 AI 干活时刷新一下"是很自然的动作)。
	// 前端收到 replay 帧**只追加时间线、不触发任何副作用**,页面状态另有 refreshSnapshot 负责。
	kept := replay[:0]
	for _, ev := range replay {
		if replayTypes[ev.Type] {
			kept = append(kept, ev)
		}
	}
	if err := writeFrame(map[string]any{"type": "replay", "epoch": s.mgr.Epoch, "events": kept}); err != nil {
		return
	}
	// hello 哨兵:带 epoch,前端据此判断服务端是否重启过(重启则序号归零,要重置去重游标)
	if err := writeFrame(map[string]any{"type": "hello", "epoch": s.mgr.Epoch}); err != nil {
		return
	}

	for {
		select {
		case ev := <-ch:
			if err := writeFrame(ev); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) hStatic(w http.ResponseWriter, r *http.Request) {
	if s.webSub == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html><html lang="zh"><meta charset="utf-8">
<title>tt debug debug</title><body style="font-family:system-ui;padding:40px;line-height:1.8">
<h2>tt debug server 已运行</h2>
<p>前端尚未构建。请在 web/ 目录执行 <code>npm install &amp;&amp; npm run build</code> 后重启服务。</p>
<p>API 已可用:<code>/api/status</code>、<code>/api/ws</code>、<code>/api/sessions</code> …</p></body>`)
		return
	}
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	f, err := s.webSub.Open(strings.TrimPrefix(path, "/"))
	if err != nil {
		// SPA 兜底:未命中路径回 index.html
		f2, err2 := s.webSub.Open("index.html")
		if err2 != nil {
			http.NotFound(w, r)
			return
		}
		f = f2
		path = "/index.html"
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if st.IsDir() {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, path, st.ModTime(), f.(readSeeker))
}

type readSeeker interface {
	Read(p []byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
}
