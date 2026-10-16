package debug

import (
	"fmt"
	"io"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"tt/internal/host"
	"tt/internal/target"

	"github.com/pkg/sftp"
)

// Manager 调试会话管理器:持有会话、向订阅者(WS/MCP)广播事件、保留最近事件供查询
type Manager struct {
	cfg *Config

	mu       sync.Mutex
	sessions map[string]*Session

	subMu sync.Mutex
	subs  map[chan Event]string

	// 最近事件环形缓冲(供 AI/日志 tail 查询;只存结构化事件,output/autovars 不入)
	evMu sync.Mutex
	evs  []Event

	// seq 事件单调序号。用原子而非在 evMu 里自增:output 事件也要有序号
	// (前端统一按序号去重),而 output 不入缓冲、根本不进 evMu。
	// 序号是进程级的、重启归零,所以配一个 Epoch 让前端知道该重置游标。
	seq   atomic.Uint64
	Epoch string

	// T100 动态环境缓存(登录区域 → 路径),TTL 5 分钟;探针失败不缓存
	envMu    sync.Mutex
	envCache map[string]cachedEnv

	// 会话外源码读取共用的连接(见 srcmirror.go):
	// 每次 tt debug source/grep 重新握手太贵,按 SSH 目标缓存一条长连接。
	//
	// 按 key 分开存（缺陷 Z / T5.3）:从前只有一条 + 一把全局锁,
	// 换目标即换连接 ⇒ 两个环境不可能同时在读源码。
	srcMu sync.Mutex
	src   map[string]*srcConn
	// srcKeyLocks 每个连接 key 一把锁:不同目标真正并行,同一目标排队（见 withSrcConn）。
	srcLocksMu  sync.Mutex
	srcKeyLocks map[string]*sync.Mutex

	// 企业目录(ENT→账号):只读 SQL / 接口日志都要先问一次 gzou_t,
	// 同一个环境(主机+账号+区域+库)下缓存 10 分钟,并落一份快照供离线使用。
	// 见 ents.go —— 取值顺序、降级规则、新鲜度口径都在那里,**只有那一份实现**。
	ents *EntCatalog
}

// maxBufferedEvents 环形缓冲上限。
const maxBufferedEvents = 4000

type cachedEnv struct {
	env *host.RuntimeEnv
	at  time.Time
}

func NewManager(cfg *Config) *Manager {
	return &Manager{
		cfg:      cfg,
		sessions: map[string]*Session{},
		subs:     map[chan Event]string{},
		envCache: map[string]cachedEnv{},
		src:      map[string]*srcConn{},
		// cfg.DataDir 可能为空(单测里的 NewManager(&Config{})):那一档只走内存不落盘
		ents:  NewEntCatalog(cfg.DataDir),
		Epoch: strconv.FormatInt(time.Now().UnixNano(), 36),
	}
}

// getRuntimeEnv 取 T100 动态环境:缓存命中直接返回;未命中(或过期)探针一次并缓存。
// 失败返回错误(不缓存),调用方须报错,无静态配置可回退。key 按 服务器+账号+区域 区分
// (用连接的真实 host,覆盖/多环境时不受 active 环境影响),换环境互不污染。
//
// **它只返回,不写任何配置**（缺陷 Y / T5.2 的修法）:动态路径是"这一次连接"算出来的事实,
// 由调用方放进**自己那一份**配置里（本轮请求克隆出来的、或会话自己持有的那一份）;
// 从前 `ensureRuntimeEnv` 把它写回 `m.cfg.Runtime` —— 而 `m.cfg` 是**共享**的,
// 于是"哪条会话先启动"决定了另一条看到的 TOP/ERP,两条会话的动态路径互相污染。
func (m *Manager) getRuntimeEnv(conn *host.SSHConn, zone string) (*host.RuntimeEnv, error) {
	cc := conn.Cfg()
	key := cc.Host + "|" + cc.User + "|" + zone
	m.envMu.Lock()
	if c, ok := m.envCache[key]; ok && time.Since(c.at) < 5*time.Minute && c.env.Valid() {
		m.envMu.Unlock()
		return c.env, nil
	}
	m.envMu.Unlock()
	env, err := host.ProbeTEnv(conn, zone)
	if err != nil {
		return nil, err
	}
	m.envMu.Lock()
	m.envCache[key] = cachedEnv{env: env, at: time.Now()}
	m.envMu.Unlock()
	return env, nil
}

// zoneOf 生效登录区域:空 = 默认 36（与既有行为一致）。
func zoneOf(cfg *Config) string {
	if cfg == nil || cfg.Zone == "" {
		return "36"
	}
	return cfg.Zone
}

// ensureRuntimeEnvOn 确保**这一份** cfg 有动态路径(探针+缓存),写进 cfg 自己而不是共享配置。
//
// 它是 `ensureRuntimeEnv` 的替代:后者写的是 `m.cfg.Runtime`（共享字段,缺陷 Y）。
// cfg 必须是调用方自己那一份（本轮的克隆 / 会话持有的快照）—— 传 `m.cfg` 进来
// 就是把这个缺陷原样搬个地方。
func (m *Manager) ensureRuntimeEnvOn(conn *host.SSHConn, cfg *Config) (*host.RuntimeEnv, error) {
	if cfg.Runtime != nil && cfg.Runtime.Valid() {
		return cfg.Runtime, nil
	}
	zone := zoneOf(cfg)
	env, err := m.getRuntimeEnv(conn, zone)
	if err != nil {
		return nil, fmt.Errorf("无法获取环境 %s(zone %s)的 T100 路径: %w", cfg.EnvName(), zone, err)
	}
	cfg.Runtime = env
	return env, nil
}

// Subscribe 订阅事件流(WS 用),返回取消函数
func (m *Manager) Subscribe(tag string) (<-chan Event, func()) {
	ch := make(chan Event, 4096)
	m.subMu.Lock()
	m.subs[ch] = tag
	m.subMu.Unlock()
	return ch, func() {
		m.subMu.Lock()
		delete(m.subs, ch)
		m.subMu.Unlock()
	}
}

// SubscribeWithReplay 订阅事件流,同时取回最近 n 条历史(供 WS 建连时开场补发)。
//
// 订阅与取快照放进**同一个 evMu→subMu 临界区**(锁序与 emit 一致):
// 这样保证**不遗漏** —— emit 的 append 先于 broadcast,窗口内产生的事件
// 要么已经进了快照、要么会进 channel。但**仍可能重复**(emit 的 append 与
// broadcast 之间不持锁),所以调用方必须用 Seq 去重。
func (m *Manager) SubscribeWithReplay(tag string, n int) (<-chan Event, []Event, func()) {
	ch := make(chan Event, 4096)
	m.evMu.Lock()
	m.subMu.Lock()
	m.subs[ch] = tag
	m.subMu.Unlock()
	var replay []Event
	if n > 0 && len(m.evs) > 0 {
		start := len(m.evs) - n
		if start < 0 {
			start = 0
		}
		replay = make([]Event, len(m.evs)-start)
		copy(replay, m.evs[start:])
	}
	m.evMu.Unlock()
	cancel := func() {
		m.subMu.Lock()
		delete(m.subs, ch)
		m.subMu.Unlock()
	}
	return ch, replay, cancel
}

// emit 会话事件回调:入环形缓冲并广播给所有订阅者(非阻塞,满则丢弃)
func (m *Manager) emit(ev Event) {
	ev.Seq = m.seq.Add(1) // 第一件事:序号必须先于任何入缓冲/广播动作
	// 只保留对"发生了什么"有信息量的结构化事件(输出流/自动变量列表不入缓冲,
	// 会占用大量内存且 AI 用 tail 查询时噪声过大)
	switch ev.Type {
	case "state", "stopped", "log", "dead", "watchdog", "ai_action":
		m.evMu.Lock()
		m.evs = append(m.evs, ev)
		if len(m.evs) > maxBufferedEvents {
			m.evs = m.evs[len(m.evs)-maxBufferedEvents:]
		}
		m.evMu.Unlock()
	}
	m.subMu.Lock()
	defer m.subMu.Unlock()
	for ch := range m.subs {
		select {
		case ch <- ev:
		default: // 订阅者消费太慢,丢弃(前端靠 state/stopped 事件对齐,丢原始行无碍)
		}
	}
}

// Events 返回最近 n 条事件(按时间升序;n<=0 返回全部缓冲)。
func (m *Manager) Events(n int) []Event {
	m.evMu.Lock()
	defer m.evMu.Unlock()
	if n <= 0 || n >= len(m.evs) {
		out := make([]Event, len(m.evs))
		copy(out, m.evs)
		return out
	}
	out := make([]Event, n)
	copy(out, m.evs[len(m.evs)-n:])
	return out
}

// Launch 创建并启动一个调试会话(同一时间仅允许一个会话,保证生产安全)。
// 启动前清扫已结束的残留会话(程序退出/后端死亡),重启无需先 DELETE。
func (m *Manager) Launch(module, prog string) (*Session, error) {
	return m.launchWith(m.cfg, module, prog, "")
}

// LaunchWith 按指定配置启动(多 SSH/多区域:调用方克隆好 cfg 再传入),并给这条会话起别名。
//
// name 是 `--name` 的取值（Q17）：空 = 未命名。别名**就是幂等键**（Q23 / C1.6）——
// 三种结局（空闲复用 / 忙碌报 SESSION_TARGET_BUSY / 名字不存在起新的）由调用方先判一次
// （`beginNamedLaunch`），这里负责把名字**登记成服务端的事实**（C5.3 的 `name` 键），
// 并在登记的瞬间兜住竞态（F18：`ErrNameTaken`）。
func (m *Manager) LaunchWith(cfg *Config, module, prog, name string) (*Session, error) {
	return m.launchWith(cfg, module, prog, name)
}

func (m *Manager) launchWith(cfg *Config, module, prog, name string) (*Session, error) {
	// 对齐 T100 gendbg:启动前连库把作业编号解析成实体程序+模块+启动引用(gzzz_t JOIN gzza_t),
	// 源码与 42r 都跟实体程序走;未配置 db/查询失败/未命中 → 会话内按名称文件搜索兜底
	runProg, launchRef, extra := "", "", ""
	if mod2, prog2, ref2, extra2, rerr := m.resolveJobWith(cfg, module, prog); rerr != nil {
		return nil, rerr
	} else if prog2 != "" {
		runProg = prog2
		launchRef, extra = ref2, extra2
		if module == "" && mod2 != "" {
			module = mod2
		}
	}
	sess, err := m.prepareSession(cfg, module, prog, runProg, launchRef, extra, "", name)
	if err != nil {
		return nil, err
	}
	// 新的一轮运行:清空上一轮攒下的源码副本(见 srcmirror.go)。
	// 放在 prepareSession 成功之后 —— 它可能因"已有活跃运行"而拒绝,
	// 那种失败不该顺手毁掉正在跑的那一轮的副本。
	// 两件事必须成对:清了目录还要清会话的"已镜像"记录,否则复用的宿主会话
	// 会以为文件早就镜像过,新一轮再也落不下副本。
	//
	// 清的是**这条会话自己的**那棵子树（缺陷 X / T5.1）:从前的 `clearMirror(cfg.DataDir)`
	// 删的是全局 `srccache/`,并发下 A 的启动会销毁 B 刚拉下来的镜像。
	// 复用的宿主会话要连"已镜像"那份记录一起继承（见 inheritMirrorState）。
	sess.inheritMirrorState()
	clearMirrorForSession(cfg.DataDir, sess.mirrorScope())
	sess.resetMirrored()
	if runProg != "" && runProg != prog {
		sess.emitEvent(Event{Type: "log", Text: fmt.Sprintf("作业编号 %s → 实体程序 %s(gzzz_t)", prog, runProg)})
	}
	return sess, nil
}

// prepareSession 启动一轮运行前的会话准备(单一常驻会话):
//   - 已有空闲(idle)且目标环境相同**且别名不冲突**的会话 → 复用宿主,仅刷新本轮运行参数(免重新登录);
//   - 已有空闲会话但目标环境不同(切环境/点名另一台服务器)→ 断开旧的再新建 ——
//     "同机同区域"不等于"同一环境":同库不同企业的两个环境在这四样上完全一样,
//     复用会让新一轮继续用旧的企业编号跑(见 Session.sameTarget);
//   - 已有活跃运行(运行中/已停站)→ 报错(请先结束当前调试);
//   - 无会话 → 新建。
//
// **name（`--name` 的取值）参与复用判据的第二半**（Q23 / C2.5b，见 nameCompatible）：
// 指定了 X 而那条空闲会话叫 Y 时**不复用**它，而是按"换目标"那条路收掉旧的、开一条叫 X 的。
// 别名只在创建时给一次，所以"要一条叫 X 的会话"绝不能落在一条已经叫 Y 的会话上。
//
// 返回已登记的会话,由调用方驱动 Launch 登录/启动。
func (m *Manager) prepareSession(cfg *Config, module, prog, runProg, launchRef, extra, argsOverride, name string) (*Session, error) {
	m.mu.Lock()
	for id, s := range m.sessions {
		if s.State() == StateExit {
			delete(m.sessions, id) // 已断开(close/切换/EOF)的残留,资源已释放
		}
	}
	var live *Session
	for _, s := range m.sessions {
		live = s
		break
	}
	m.mu.Unlock()

	if live != nil {
		switch decideHost(live, cfg, name) {
		case hostReuse:
			// 复用宿主跑新一轮：SetRun 会把世代换掉（F15：旧句柄当场失效），
			// 序号与别名**不动**（它们标识的是"这条会话"，跨轮不变：C2.0 / C2.5b）。
			live.SetRun(cfg, module, prog, runProg, launchRef, extra)
			live.ArgsOverride = argsOverride
			// 这条宿主此前没有名字，而这次点名要 X：别名登记给它（Q23 的"起新的"落在
			// 复用宿主上的形态）。它此前无名，所以不是改名（C2.5b 禁的是改名）。
			if err := m.nameSession(live, name); err != nil {
				return nil, err
			}
			return live, nil
		case hostBusy:
			// **运行中 / 已停站：不许悄悄收掉它** —— 这与今天那句"请先结束当前调试"
			// 是同一条取向（H2 保护的就是这句）。别名对不上也不改变这一条：
			// 要抢名字先停掉那条会话（F20 的 hint 是这么说的）。
			return nil, fmt.Errorf("已存在调试会话(%s 运行中/已停站),请先结束当前调试", live.Prog)
		default: // hostCollect
			// 收口旧会话再按新目标/新名字重建：空转的宿主没什么可丢的，
			// 而且它从池里移除的那一刻，它的别名与序号**当场释放**（C3.7b / Q19）。
			log.Printf("[debug] 目标切换为 %s(旧会话 %s 在 %s),断开旧会话后重建",
				cfg.EnvName(), live.ID, live.EnvNameNow())
			_ = live.Close()
			m.Remove(live.ID)
		}
	}

	// 新会话：别名先查（F18）再拨号 —— 注定要被拒的请求不该先花一次 SSH 握手。
	if err := m.checkNameFree(name); err != nil {
		return nil, err
	}
	sess, err := NewSession(cfg, module, prog, runProg, launchRef, extra, m.emit)
	if err != nil {
		return nil, err
	}
	if argsOverride != "" {
		sess.ArgsOverride = argsOverride
	}
	m.mu.Lock()
	if err := m.assignIdentityLocked(sess, name); err != nil {
		// 两次判断之间被人抢了名字（F18 的竞态那一支）：这条会话不能留在池外活着。
		m.mu.Unlock()
		sess.ForceExit()
		return nil, err
	}
	m.sessions[sess.ID] = sess
	m.mu.Unlock()
	return sess, nil
}

// resolveJob 连库按 gendbg 语义解析作业编号(gzzz_t:gzzz001 → gzzz002 实体程序 + gzzz005 模块)。
// 返回 (模块, 实体程序, err):err 仅在 SSH 连接/动态路径探测失败时非 nil;
// 未配置 db/查询失败/未命中 → ("","","","", nil),由调用方回退文件搜索。
func (m *Manager) resolveJob(module, job string) (mod, prog, launchRef, extra string, err error) {
	// 传一份**副本**（缺陷 Y / T5.2）:`resolveJobWith` 会把探到的动态路径写进它拿到的
	// 那一份 cfg,而这一条路拿到的正是共享的 `m.cfg` —— 一次作业解析就会把
	// "当前生效环境"的 Runtime 改掉,另一条会话随后读到的是别人的路径。
	cc := *m.cfg
	return m.resolveJobWith(&cc, module, job)
}

// resolveJobWith 同 resolveJob,但按传入 cfg(多 SSH/多区域克隆)解析。
func (m *Manager) resolveJobWith(cfg *Config, module, job string) (mod, prog, launchRef, extra string, err error) {
	if cfg.DB == nil || !reProgName.MatchString(job) {
		return "", "", "", "", nil
	}
	conn, err := host.Dial(cfg.SSH)
	if err != nil {
		return "", "", "", "", fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer conn.Close()
	// 动态路径(登录区域 → 环境脚本):按本次启动的区域探针+缓存(覆盖/多环境时不用
	// 全局 active 区域),结果写进**调用方自己那一份** cfg（缺陷 Y / T5.2:
	// 不再落到共享的 `m.cfg.Runtime` 上）——预会话模块解析/42r 校验直接用真实路径。
	// 失败即报错:T100 路径只来自登录动态获取,无静态配置可回退
	if _, err := m.ensureRuntimeEnvOn(conn, cfg); err != nil {
		return "", "", "", "", err
	}
	// 显式连接(SSH 页 dbConn 引用)连库解析作业;失败静默,由调用方回退文件搜索
	d, err := resolveDBRun(conn, cfg)
	if err != nil {
		return "", "", "", "", nil
	}
	jr, err := dbResolveJob(conn, d, job)
	if err != nil || jr.Prog == "" {
		return "", "", "", "", nil
	}
	p := jr.Prog
	// gendbg 原版语义:gzza004 的 $变量 由选区后 shell 展开为权威 42r 路径(标准/客制都覆盖)
	if !strings.HasPrefix(jr.LaunchRef, "$") {
		jr.LaunchRef, jr.Extra = "", ""
	}
	launchRef, extra = jr.LaunchRef, jr.Extra
	if module != "" {
		return module, p, launchRef, extra, nil // 用户显式指定模块:尊重指定,仅采纳实体程序
	}
	mod = strings.ToLower(jr.Module)
	if jr.Module == "" || jr.Module == "-" || !modHas42r(conn, cfg.TopDirActual(), mod, p) {
		// 模块码缺失或目录对不上:全模块搜索实体程序的 42r;0/多命中留给会话内兜底报错
		if cands := searchModule42r(conn, cfg.ModuleRootsActual(), p); len(cands) == 1 {
			mod = cands[0]
		} else {
			mod = ""
		}
	}
	return mod, p, launchRef, extra, nil
}

// LaunchReplay 接口日志重放调试:等价 T100 日志内嵌的 `r.dg <作业> '<req>' '<rsp>'`。
// 作业取 wsfa012(gzja_t 服务→程序的解析结果),再走 gzzz_t 解析实体程序+模块;
// 报文文件被清理时用 CLOB 内容落到服务器临时文件再重放。
// 会话复用规则与普通启动一致:同目标空闲宿主直接复用。
// LaunchReplay 用日志报文重放调试。reqOverride 非空 = 界面里改过的入参(落临时文件后作为 reqPath)。
// fglserver 非 nil = 本轮 FGLSERVER 覆盖(语义见 Server.launchCfg 与 ScreenTarget)。
// 返回的 warn 非空表示这次重放用的是入库的截断文本,结果可能不可信(见 WriteReplayFiles)。
func (m *Manager) LaunchReplay(item *WSLogItem, content *WSLogContent, reqOverride string, fglserver *string) (*Session, string, error) {
	// 本轮启动用的配置快照:只有"单轮 FGLSERVER 覆盖"会改它。
	// 直接改 m.cfg 会把一次重放的接线留在内存配置里 —— 下一轮(哪怕只是普通启动)
	// 都还带着它,而画面去哪恰恰是静默的(见 ScreenTarget)。
	cfg := m.cfg
	if fglserver != nil {
		cc := *m.cfg
		cc.FGLServer = strings.TrimSpace(*fglserver)
		cfg = &cc
	}
	job := strings.TrimSpace(item.Job)
	if job == "" {
		return nil, "", fmt.Errorf("该日志没有关联作业编号(wsfa012 为空),无法重放")
	}
	module, runProg, launchRef, extra := "", "", "", ""
	if mod2, prog2, ref2, extra2, rerr := m.resolveJobWith(cfg, "", job); rerr != nil {
		return nil, "", rerr
	} else if prog2 != "" {
		module, runProg, launchRef, extra = mod2, prog2, ref2, extra2
	}
	conn, err := host.Dial(cfg.SSH)
	if err != nil {
		return nil, "", fmt.Errorf("SSH 连接失败: %w", err)
	}
	defer conn.Close()
	reqPath, rspPath, replayWarn, err := WriteReplayFiles(conn, item, content, reqOverride)
	if err != nil {
		return nil, "", err
	}
	// 路径进 shell 单引号,剔除单引号防注入。
	// 对齐 r.dg:报文文件追加在标准参数(BBDL 会话标记等)之后,
	// cl_wss_init 依 BBDL 标记识别调试模式并重建 g_argv;只传报文会导致 g_argv 为空 → -1
	q := func(p string) string { return "'" + strings.ReplaceAll(p, "'", "") + "'" }
	args := strings.ReplaceAll(cfg.LaunchArgs, "{prog}", job)
	args += " " + q(reqPath)
	if rspPath != "" {
		args += " " + q(rspPath)
	}
	sess, err := m.prepareSession(cfg, module, job, runProg, launchRef, extra, args, "")
	if err != nil {
		return nil, "", err
	}
	// 新的一轮运行:清空上一轮攒下的源码副本(见 srcmirror.go);
	// "已镜像"记录必须跟着一起清(复用的宿主会话会把它带进下一轮)。
	// 同 launchWith:只清**这条会话自己的**那棵子树（缺陷 X / T5.1）。
	sess.inheritMirrorState()
	clearMirrorForSession(cfg.DataDir, sess.mirrorScope())
	sess.resetMirrored()
	// 用了入库的截断文本重放:说清楚,别让人把"程序一路退出"当成 bug 去查
	if replayWarn != "" {
		sess.emitEvent(Event{Type: "log", Text: "⚠ " + replayWarn})
	}
	if runProg != "" && runProg != job {
		sess.emitEvent(Event{Type: "log", Text: fmt.Sprintf("重放调试:作业 %s → 实体程序 %s(gzzz_t)", job, runProg)})
	}
	if reqOverride != "" {
		sess.emitEvent(Event{Type: "log", Text: fmt.Sprintf("重放使用界面修改后的入参(%d 字节),已写入 %s", len(reqOverride), reqPath)})
	}
	// 记下响应会落到哪:程序退出后要把它读回来回显(见 loadReplayResponse)。
	// 响应文件是程序退出时才写的,所以只能等跑完再读 —— 停在入口/断点时它还不存在。
	sess.mu.Lock()
	sess.replayRspPath = rspPath
	sess.mu.Unlock()
	sess.emitEvent(Event{Type: "log", Text: fmt.Sprintf("重放 %s:fglrun -d %s %s", item.Service, runProgOr(job, runProg), args)})
	return sess, replayWarn, nil
}

func runProgOr(job, runProg string) string {
	if runProg != "" {
		return runProg
	}
	return job
}

// Get 取会话
func (m *Manager) Get(id string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[id]
}

// Current 返回当前唯一会话(无则 nil)
func (m *Manager) Current() *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		return s
	}
	return nil
}

// Snapshot 会话列表快照
//
// 键集是契约 C5.1 / C5.2 / C5.3 定死的那一份（`session` 块的形状）：
// **十一个恒定键**（`id`/`ref`/`display`/`name`/`env`/`ent`/`n`/`kind`/`prog`/`state`/`epoch`）
// 加一批按 `omitempty` 出现的快照键。恒定键**取不到就给空串/零值，不省键** ——
// 前端与 CLI 按这份形状解包，少一个键它们就得写第二套兜底。
//
// 五个新键的语义（Q22：纯新增、不改旧键）：
//   - `ref`    规范 id `环境#序号`，**生命周期内恒定不变**（C2.0 / C5.4）——身份，不是快照；
//   - `display` 人读形态 `环境/作业#序号`，作业段是快照，会随 `SetRun` 漂（C5.4）；
//   - `name`   别名，空串 = 未命名（C2.5b）；
//   - `ent`    **当前**生效的企业编号，会随 `tt debug topent` 变（C2.5c）——快照，不是地址；
//   - `n`      序号：同一环境下的第 N 条（C2.2 / Q24）——身份，建立时分配、之后不变；
//   - `epoch`  世代号（C5.3 / F15）——见 `Session.Epoch` 的语义。
type SessionBrief struct {
	ID string `json:"id"`
	// Ref 是**规范 id**：`环境#序号`（C2.1.2 / C5.4）。只由不可变部分拼 ——
	// 企业编号与作业段都不在里面（C2.0：可变字段不进地址）。
	Ref string `json:"ref"`
	// Display 是人读形态：`环境/作业#序号`（C5.4）。比 ref 多一段**当前作业快照**，
	// 所以作业中途变了它就会变 —— 要恒定用 ref，要好读用 display，两者不混用。
	Display string `json:"display"`
	// Name 是别名（Q17 / C2.5b）。空串 = 未命名（**不省键**，C5.1 明写取不到给空串）。
	Name string `json:"name"`
	// Ent 是该会话**当前**生效的企业编号（C5.3 / C2.5c）：shell 回读值优先，
	// 取不到退到本会话显式给的值；空 = 未取到。**它是快照**，`SetTopent` 会改它。
	Ent string `json:"ent"`
	// N 是序号（同一环境下的第 N 条，从 1 起；C2.2 / Q24）。建立时分配，之后不变。
	N int `json:"n"`
	// Kind 是会话类型（C5.5）：本契约只寻址 SSH 会话，所以恒为 `ssh`
	// （界面会话不是第二个目标，它是这条会话的属性，走信封的 `ui` 块）。
	Kind     string  `json:"kind"`
	Module   string  `json:"module"`
	Prog     string  `json:"prog"`
	RunProg  string  `json:"runProg,omitempty"`
	State    string  `json:"state"`
	Env      string  `json:"env,omitempty"` // 会话所属环境名(设置页 envs)
	Started  bool    `json:"started"`
	File     string  `json:"file,omitempty"`
	Line     int     `json:"line,omitempty"`
	Func     string  `json:"func,omitempty"`
	Reason   string  `json:"reason,omitempty"`
	Holding  float64 `json:"holdingSeconds"`
	Breaks   int     `json:"breakpoints"`
	Watchdog int     `json:"watchdogSeconds"`
	// 停站现场就是交互语句(INPUT/MENU/DISPLAY ARRAY…):程序把控制权交给了界面。
	// 零成本判定(只看停站那一行的源码文本),供 AI 直接分支,不必再读源码做预测。
	WaitingForUser bool   `json:"waitingForUser,omitempty"`
	WaitingKind    string `json:"waitingKind,omitempty"`
	// SilentSeconds:运行态下距最近一次协议输出的秒数 —— 判断"跑了很久却毫无动静"
	// (可能就是停在界面上等用户,也可能是慢查询)。停站态为 0。
	SilentSeconds float64 `json:"silentSeconds,omitempty"`
	// Mode 谁在驾驶:solo(纯人工)| collab(协作,AI 主导)
	Mode string `json:"mode"`
	// Epoch 是世代号（C5.3 / F15）。空 = 没有代（不该出现：登记时必发一个）。
	Epoch string `json:"epoch"`
	// Inflight 正在执行的那条命令(空闲为 nil)—— 界面上"AI 正在执行 continue(已 12s)"
	Inflight *InflightInfo `json:"inflight,omitempty"`
}

// refOf / displayOf 拼身份的两个字符串（C5.4 的三分工）。
//
// 序号还没分配（0：只可能出现在手工造的会话上）时给**空串**，绝不编一个 `#0` ——
// `ref` 是恒定不变的地址，一个错的地址比一个空字段坏得多（C2.0 / C5.4。
// `internal/target` 的 `CanonicalRef` 在服务端没给 ref 时会退回原始句柄，那是它的兜底，
// 不是让服务端去编）。
func refOf(env string, n int) string {
	if strings.TrimSpace(env) == "" || n <= 0 {
		return ""
	}
	return target.RefOf(env, n)
}

func displayOf(env, prog string, n int) string {
	if refOf(env, n) == "" {
		return ""
	}
	return target.Session{Env: env, Prog: prog, N: n}.DisplayForm()
}

func (m *Manager) Snapshot() []SessionBrief {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]SessionBrief, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, m.briefOf(s))
	}
	return out
}

// briefOf 把一条会话投影成它的信封形状（C5.1 的 `session` 块）。
//
// `Snapshot` 与**候选清单**（C5.12 的九键）共用这一份投影 —— 候选是 `session` 的
// 真子集（C5.4b），投影写成两处必然漂，而漂的那一处恰好是"照着候选重发一枪"要用的字段。
//
// 调用方**不必**持有 `m.mu`：`briefOf` 只读会话自己的锁与不可变的 `m.cfg`。
func (m *Manager) briefOf(s *Session) SessionBrief {
	cur := s.Cur()
	env, n := s.EnvName(), s.Seq()
	return SessionBrief{
		ID: s.ID,
		// 身份（C5.1 末段：只有 id / ref / n 是身份，判定身份只许用这三个）
		Ref: refOf(env, n), Display: displayOf(env, s.Prog, n),
		N: n, Kind: target.KindSSH, Epoch: s.Epoch(),
		// 快照（会变：name 只在创建时给，ent 随 topent 改）
		Name: s.Name(), Ent: s.CurrentEnt(),
		Module: s.Module, Prog: s.Prog, RunProg: s.RunProg,
		State: string(s.State()), Env: s.EnvName(), Started: s.Started(), File: cur.File, Line: cur.Line,
		Func: cur.Func, Reason: cur.Reason,
		Holding:        s.HoldingSeconds(),
		Breaks:         len(s.Breakpoints()),
		Watchdog:       m.cfg.WatchdogSeconds,
		WaitingForUser: cur.WaitingForUser,
		WaitingKind:    cur.WaitingKind,
		SilentSeconds:  s.SilentSeconds(),
		Mode:           s.Mode(),
		Inflight:       s.Inflight(),
	}
}

// candidatesOf 把若干条会话投影成 C5.12 的候选九键（**唯一出处**是 `internal/target`）。
//
// 顺序按 `id` 排：候选块是给人照着重发用的，同一个失败在两次请求里必须列出同一种顺序
// （`m.sessions` 是 map，"顺序随机"会让报错信息没法当判据用）。
func (m *Manager) candidatesOf(sessions []*Session) []target.Candidate {
	out := make([]target.Candidate, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, sessionView(m.briefOf(s)).Candidate())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ---------- 会话身份：序号 / 别名 / 世代（契约 C5.3 / C2.2 / Q19 / Q23） ----------

// ErrNameTaken 是 F18：要登记的别名已经被池子里另一条**活着的**会话占着。
//
// 它判在**注册那一刻**：Q23 的三种结局（空闲复用 / 忙碌报错 / 起新的）在起会话那条路上
// 先判一次（见 `internal/debug/api.go` 的 `beginNamedLaunch`），这里兜住两次判断之间的
// 竞态 —— 两个同名请求同时在飞时，晚的那一个必须报错，**不许**把名字从别人手里抢过来。
type ErrNameTaken struct {
	Name     string
	Occupant string // 占用者的服务端原始句柄（调用方用它取候选行，C5.12）
}

func (e *ErrNameTaken) Error() string {
	return fmt.Sprintf("别名 %q 已被会话 %s 占用；换一个名字，或先结束那条会话", e.Name, e.Occupant)
}

// SweepExited 把已断开（`StateExit`）的会话从池子里清掉。
//
// 为什么本批需要它（Q19 / C3.7b）：**别名与序号都挂在会话对象上** —— 会话从池子里消失，
// 别名立刻可以被别人用。而"断开"到"被清掉"之间原本有一段窗口（只有 `prepareSession` /
// `CreateSessionOn` 顺手扫），起会话时若不先扫一遍，一条早就断了的会话会继续占着别名，
// 于是"重起同名被拒"，报的还是"别名已被 <ref> 占用"而 `<ref>` 指向一条早没了的会话 ——
// C3.7b 点名的正是这个极难查的故障。
//
// 判据锚在"**池子里还有没有它**"，与"人敲没敲 quit"无关：SSH 掉线 / 进程被杀 / 断电
// 都会走 `forceExit` → `StateExit`，于是同一套判据覆盖异常结束。
func (m *Manager) SweepExited() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweepExitedLocked()
}

func (m *Manager) sweepExitedLocked() {
	for id, s := range m.sessions {
		if s.State() == StateExit {
			delete(m.sessions, id) // 已断开的残留：资源已释放，身份（含别名）随之释放
		}
	}
}

// allocSeqLocked 在同一环境下取"**尚未被占用的最小正整数**"（C2.2 / C2.3 / Q16 / Q24），从 1 起。
//
// 企业编号**不参与**：按（环境, 企业编号）自增会让两条会话拿到同一个 `环境#序号`
// （企业 99 的 #1 与企业 88 的 #1 都叫 `恒烁正式区#1`），地址当场撞车 —— 2.1.3 写死了理由。
//
// 序号**不单独记账**：它按池子里活着的会话现算，所以"会话消失 → 序号释放"是天然的，
// 与别名同一条性质（Q19 的同源口径）。已断开（`StateExit`）的残留不算占用者。
func (m *Manager) allocSeqLocked(env string) int {
	used := map[int]bool{}
	for _, s := range m.sessions {
		if s.State() == StateExit || s.EnvName() != env {
			continue
		}
		if n := s.Seq(); n > 0 {
			used[n] = true
		}
	}
	for n := 1; ; n++ {
		if !used[n] {
			return n
		}
	}
}

// identityOwnerLocked 找别名的占用者：**整个会话池、不区分大小写**（C2.5 的查重口径）。
//
// 与 `internal/target` 的 `ResolveName` 同一个范围与同一条查重规矩（`SameAlias`）：
// 起会话时的"空闲复用 / 忙碌报错"用它，登记的撞车判定也用它，两处必须同源 ——
// 否则会出现"起的时候说没撞、找的时候找到两条"。
func (m *Manager) identityOwnerLocked(name string) *Session {
	if strings.TrimSpace(name) == "" {
		return nil
	}
	for _, s := range m.sessions {
		if s.State() == StateExit {
			continue
		}
		if s.Name() != "" && target.SameAlias(s.Name(), name) {
			return s
		}
	}
	return nil
}

// checkNameFree 在**建连接之前**问一句"这个名字能不能给"（F18）。
//
// 顺序是刻意的：注定要被拒的请求不该先拨一次 SSH 再拒。竞态由
// `assignIdentityLocked` 里的同一次判断兜住（两处都在 m.mu 下，判据只有一份）。
func (m *Manager) checkNameFree(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweepExitedLocked()
	if owner := m.identityOwnerLocked(name); owner != nil {
		return &ErrNameTaken{Name: name, Occupant: owner.ID}
	}
	return nil
}

// assignNameLocked 把别名登记给一条会话（空名字 = 什么都不做）。
func (m *Manager) assignNameLocked(sess *Session, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	if owner := m.identityOwnerLocked(name); owner != nil && owner != sess {
		return &ErrNameTaken{Name: name, Occupant: owner.ID}
	}
	sess.mu.Lock()
	sess.name = name
	sess.mu.Unlock()
	return nil
}

// nameSession 给一条**已经存在、还没名字**的会话补一个别名。
//
// 用途只有一处：Q23 的"名字不存在 → 起新的"在**复用宿主**形态下的落点 ——
// 池子里没有叫 X 的会话，而恰好有一条空闲的同目标宿主，于是这一轮就落在这条宿主上，
// X 就是"这条会话叫什么"。它此前没有名字，所以这不是改名（C2.5b 禁的是改名）。
func (m *Manager) nameSession(sess *Session, name string) error {
	if strings.TrimSpace(name) == "" {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sweepExitedLocked()
	return m.assignNameLocked(sess, name)
}

// assignIdentityLocked 给一条**刚建立**的会话分配身份（C5.3 的四键：n / name / epoch，ref 由 n 拼）。
//
// 顺序是刻意的：先查别名（撞车就整条不成立，调用方按 F18 报错），再发序号与世代。
// 别名与序号都只在这里发一次，之后**不再重算** —— 序号恒定的依据（C2.0）就在这一句上。
// 调用方必须持有 `m.mu`。
func (m *Manager) assignIdentityLocked(sess *Session, name string) error {
	if err := m.assignNameLocked(sess, name); err != nil {
		return err
	}
	seq := m.allocSeqLocked(sess.EnvName()) // 按**环境**自增（Q24），企业编号不参与
	sess.mu.Lock()
	sess.seq = seq
	if sess.epoch == "" { // NewSession 已经发过第一代；手工造的会话在这里补一个
		sess.epoch = nextSessionEpoch()
	}
	sess.mu.Unlock()
	return nil
}

// nameCompatible 判"这条既有会话能不能充当这次要的那条"（复用判据的第二半）。
//
// `want` 为空 = 这次没点名要别名 → 既有会话叫什么都能用（沿用今天的复用语义，H2）。
// `want` 非空 = 只许落在"还没名字"或"正好叫这个名字"的会话上：别名**只在创建时给一次**
// （C2.5b），把 X 挂到一条已经叫 Y 的会话上就是静默改名 —— 而 Y 那个句柄还在别人手上。
// 名字对不上时不复用：按"换目标"那条路收掉旧的、按新名字开一条（Q23 的"起新的"）。
func nameCompatible(stored, want string) bool {
	if strings.TrimSpace(want) == "" {
		return true
	}
	if strings.TrimSpace(stored) == "" {
		return true
	}
	return target.SameAlias(stored, want)
}

// hostDecision 是 `prepareSession` 对"池里那条既有会话"的三种处置。
type hostDecision int

const (
	// hostReuse：同目标 + 空闲 + 别名不冲突 → 复用宿主（免重新登录）。
	hostReuse hostDecision = iota
	// hostCollect：**空闲**的宿主，但换了目标（环境）或换了名字 → 收口它再按新的重建。
	// 空转的宿主没什么可丢的；它从池里消失的那一刻，别名与序号当场释放（C3.7b / Q19）。
	hostCollect
	// hostBusy：同目标而那条会话**正在跑 / 已停站** → 报错，**不许悄悄收掉它**。
	hostBusy
)

// decideHost 是"既有会话怎么处置"的**唯一判据**（单一常驻会话的复用规则 + Q23 的别名那一半）。
//
// 写成纯函数是为了让它可测：这三种处置各自的分支都要真拨 SSH 才走得到（除了复用），
// 而"什么时候敢收掉别人的会话"恰恰是最不该靠碰运气验的一条。
//
// 三条规矩，按顺序（**`wantName` 为空时逐字等于今天的行为**，这是 H2 的落点）：
//
//  1. **换了目标**（环境/连接不同）→ 收口重建。今天的原话：即使连接可复用也不复用 ——
//     否则会"换了环境却还连着旧企业/旧账号"。这一条今天就不看状态（收口会先中断）。
//  2. **同目标但不在空闲**（运行中 / 已停站）→ **报错**。今天那句"已存在调试会话,请先结束当前调试"
//     就是它；别名对不上也不改变它（要抢名字先停掉那条会话，F20 的 hint 同一条口径）。
//  3. **同目标 + 空闲** → 别名不冲突就复用；**别名对不上**（要 X 而它叫 Y）则收口重建 ——
//     它是空闲的，没有在跑的活要保护，而别名只能在创建时给一次（C2.5b）。
func decideHost(prev *Session, cfg *Config, wantName string) hostDecision {
	if prev == nil {
		return hostCollect
	}
	if !prev.sameTarget(cfg) {
		return hostCollect
	}
	if prev.State() != StateIdle {
		return hostBusy
	}
	if !nameCompatible(prev.Name(), wantName) {
		return hostCollect
	}
	return hostReuse
}

// Remove 移除已结束的会话
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()
}

// CreateSessionOn 新建并登记一个纯宿主会话(module/prog 为空,登录后回到 idle),
// 供「会话重启/切换/连接」在指定目标环境上重连;调用方随后驱动 BootIdle。
//
// name 是这条新会话要不要继承的别名：
//   - **会话重启**（`/api/sessions/{id}/restart`）传旧会话的名字 —— 重启是"同一条逻辑会话
//     重新登录"（F15 的前提就是"环境与序号都对、只是换了代"）：旧会话已从池里移除、
//     别名当场释放，这里就地把同一个名字重新登记给新会话，于是调用方手上的别名/ref 继续有效；
//   - **连到该环境**（`/api/sessions/switch`）传空 —— 那是新建一条会话，别名随旧会话消失
//     即释放（C3.7b / Q19），要名字就再给一次。
func (m *Manager) CreateSessionOn(cfg *Config, name string) (*Session, error) {
	m.mu.Lock()
	m.sweepExitedLocked()
	m.mu.Unlock()
	if err := m.checkNameFree(name); err != nil {
		return nil, err
	}
	sess, err := NewSession(cfg, "", "", "", "", "", m.emit)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	if err := m.assignIdentityLocked(sess, name); err != nil {
		m.mu.Unlock()
		sess.ForceExit()
		return nil, err
	}
	m.sessions[sess.ID] = sess
	m.mu.Unlock()
	return sess, nil
}

// SourcePreview 会话建立前预取源码:读母版消除前端启动期空白,顺带落本地镜像。
// module 为空(留待会话内自动解析)时直接跳过。
//
// 它跑在会话之前,所以作用域只能是**目标**（环境段,`srcConnScopeSeg`）——
// 与从前的落点逐字相同;会话建起来之后由该会话自己的作用域接着落（见 mirrorScope）。
func (m *Manager) SourcePreview(module, prog string) (*SourceFile, error) {
	if module == "" {
		return nil, fmt.Errorf("module 为空,跳过源码预取")
	}
	dvmFile := module + "_" + prog + ".4gl"
	// 一次调用只认一份配置快照:下面几处都用它,免得中途被别的请求改掉的共享态混进来。
	cfg := m.sourceCfg()
	scope := srcConnScopeSeg(cfg)
	var out *SourceFile
	err := m.withSrcConn(cfg, func(_ *host.SSHConn, cl *sftp.Client) error {
		var lastErr error
		for _, p := range sourceCandidatePaths(cfg.ModuleRootsActual(), module, dvmFile) {
			f, err := cl.Open(p)
			if err != nil {
				lastErr = err
				continue
			}
			var mt time.Time
			if fi, err := f.Stat(); err == nil {
				mt = fi.ModTime()
			}
			data, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				lastErr = err
				continue
			}
			out = &SourceFile{
				DVMFile: dvmFile, Path: p, Content: string(data), ModTime: mt,
				LocalPath: writeMirrorScoped(cfg.DataDir, scope, p, data),
			}
			return nil
		}
		return fmt.Errorf("源码未找到(%s): %w", dvmFile, lastErr)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// sourceCfg 取"会话外读取"要用的那份配置快照（缺陷 Y / T5.2）。
//
// 为什么是副本而不是 `m.cfg` 本身:探针会把动态路径写进拿到的那一份 cfg,
// 而 `m.cfg` 是**共享**的 —— 一条会话的启动或一次源码读取都会改掉另一条看到的路径。
func (m *Manager) sourceCfg() *Config {
	cc := *m.cfg
	return &cc
}

// ---------- 会话外只读能力(AI/CLI 信息通道;不经会话,独立短连接,白名单限登录区源码目录) ----------

// SourceResult 会话外源码读取结果(带行段裁剪,避免大文件整读灌满 AI 上下文)。
type SourceResult struct {
	SourceFile
	From int `json:"from"` // 1-based 返回行段起点
	To   int `json:"to"`   // 1-based 返回行段终点(含)
	All  int `json:"all"`  // 文件总行数
}

// maxSourceFileBytes 会话外读取的文件大小上限(源码一般远小于此)。
const maxSourceFileBytes = 4 << 20

// ReadSourceStandalone 读取登录区源码目录(动态 ERP/COM)内的源码文件。
//   - path 非空:直接按路径读(白名单校验源码目录前缀);
//   - path 为空:file(如 asf_bsft001_wf.4gl)按 sourceCandidatePaths 候选解析,module 可空(仅查公共目录);
//   - from/to(1-based):返回行段(from<=0 从头,to<=0 到底),行文本超长截断防注入超长行。
//
// 走会话外复用连接(srcmirror.go),顺带把整份源码落本地镜像并在结果里带出 LocalPath。
// 返回裁剪后的内容与文件总行数。
func (m *Manager) ReadSourceStandalone(module, file, path string, from, to int) (*SourceResult, error) {
	if path == "" && file == "" {
		return nil, fmt.Errorf("需要 file 或 path 参数")
	}
	var (
		data []byte
		mt   time.Time
	)
	// 一次调用只认一份配置快照（缺陷 Y / T5.2）:探到的动态路径写在这一份上,
	// 共享的 `m.cfg` 一字不动。
	cfg := m.sourceCfg()
	scope := srcConnScopeSeg(cfg)
	err := m.withSrcConn(cfg, func(_ *host.SSHConn, cl *sftp.Client) error {
		roots := cfg.ModuleRootsActual()
		if len(roots) == 0 {
			return fmt.Errorf("登录区源码目录未知(环境路径解析失败)")
		}
		readOne := func(p string) ([]byte, time.Time, error) {
			f, err := cl.Open(p)
			if err != nil {
				return nil, time.Time{}, err
			}
			defer f.Close()
			var t time.Time
			if fi, err := f.Stat(); err == nil {
				t = fi.ModTime()
			}
			d, err := io.ReadAll(f)
			return d, t, err
		}
		// 路径白名单:必须在登录区源码目录之内
		inRoots := func(p string) bool {
			for _, root := range roots {
				if strings.HasPrefix(p, root+"/") {
					return true
				}
			}
			return false
		}
		if path != "" {
			if !inRoots(path) {
				return fmt.Errorf("路径不在登录区源码目录内: %s", path)
			}
			d, t, err := readOne(path)
			if err != nil {
				return err
			}
			data, mt = d, t
			return nil
		}
		var lastErr error
		for _, p := range sourceCandidatePaths(roots, module, file) {
			if !inRoots(p) {
				continue
			}
			d, t, e := readOne(p)
			if e == nil {
				data, mt, path = d, t, p
				return nil
			}
			lastErr = e
		}
		return fmt.Errorf("源码未找到(%s): %v", file, lastErr)
	})
	if err != nil {
		return nil, err
	}
	// 镜像落的是**整份**源码(在下面的行段裁剪与大小截断之前),本地副本要能整读
	mirrorPath := writeMirrorScoped(cfg.DataDir, scope, path, data)
	if len(data) > maxSourceFileBytes {
		data = data[:maxSourceFileBytes]
	}
	lines := strings.Split(string(data), "\n")
	all := len(lines)
	// 去尾部空行(文件常见尾换行)
	for all > 0 && strings.TrimSpace(lines[all-1]) == "" {
		all--
	}
	if from <= 0 {
		from = 1
	}
	if to <= 0 || to > all {
		to = all
	}
	if from > to {
		from, to = 1, all
	}
	slice := lines[from-1 : to]
	// 单行超长截断(记录行/极长字符串防爆)
	for i, ln := range slice {
		if len(ln) > 4096 {
			slice[i] = ln[:4096] + "…(截断)"
		}
	}
	res := &SourceResult{
		SourceFile: SourceFile{DVMFile: file, Path: path, ModTime: mt, LocalPath: mirrorPath},
		From:       from, To: to, All: all,
	}
	if from <= all {
		res.Content = strings.Join(slice, "\n")
	}
	return res, nil
}

// JobInfo 作业编号 → 实体程序/模块解析结果(信息通道,不启动会话)。
type JobInfo struct {
	Prog      string `json:"prog"`                // 用户作业编号
	Module    string `json:"module"`              // 解析出的模块目录(空=未解析)
	RunProg   string `json:"runProg"`             // 实体程序(gzzz_t 解析;空=未解析)
	LaunchRef string `json:"launchRef,omitempty"` //
	Extra     string `json:"extra,omitempty"`     //
	Note      string `json:"note,omitempty"`      // 未解析/降级时的说明
}

// ResolveJob 尽力解析作业:未配置 db 或连库失败/未命中时返回 note,不报错(启动时会再次解析/兜底)。
func (m *Manager) ResolveJob(module, prog string) *JobInfo {
	out := &JobInfo{Prog: prog, Module: module}
	if m.cfg.DB == nil {
		out.Note = "config.json 未配置 debug.db,无法按 gzzz_t 解析实体程序;将按作业名直接启动"
		return out
	}
	if !reProgName.MatchString(prog) {
		out.Note = "作业名非标准标识符,跳过库解析"
		return out
	}
	// resolveJobWith 内部含 连库→gzzz_t→42r 校验;动态路径/连接失败(err)或解析失败(空串)都提示。
	// 传副本（缺陷 Y / T5.2）:它会把探到的动态路径写进拿到的那一份 cfg。
	cfg := m.sourceCfg()
	mod, run, ref, extra, rerr := m.resolveJobWith(cfg, module, prog)
	if rerr != nil {
		out.Note = "无法获取该环境 T100 路径: " + rerr.Error()
		return out
	}
	if run == "" {
		out.Note = "未能按 gzzz_t 解析(连库失败或未命中);将按作业名直接启动"
		return out
	}
	if mod != "" {
		out.Module = mod
	}
	out.RunProg, out.LaunchRef, out.Extra = run, ref, extra
	return out
}

// ---------- 会话生命周期 ----------

// CloseAll 断开全部会话(服务退出时调用)
func (m *Manager) CloseAll() {
	m.mu.Lock()
	ss := make([]*Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		ss = append(ss, s)
	}
	m.sessions = map[string]*Session{}
	m.mu.Unlock()
	for _, s := range ss {
		_ = s.Close()
	}
}
