package web

// /api/recordings —— 「录制」面板的服务端。
//
// 三件事：列（按环境 + ENT）、看一份、起一轮/停一轮。
//
// 为什么"起一轮"是**另起一个进程**（`winproc.SpawnDetached`）而不是在 serve 里起：
// 一轮录制 = 一个隔离 GDC + 一条代理 + 一个常驻 Runtime，它天然属于**一个前台进程**
// （`tt recording start` 就是这么活的）。让 serve 兼着管它会牵出两件麻烦事：
// serve 一重启录制就断，且它登记的是自己那个 pid（`tt recording stop <录制ID>` 会去杀 serve）。
// 这与 `tt update` 那条按钮的做法一致（api/update.go 的 hUpdateApply）。
//
// 会话怎么被发现：子进程起来就往 `.tt-fep.json` 写登记 —— 面板据此拿到**录制 ID**与端口，
// 命令行那一侧看到的是同一份登记（`tt recording active` / `session close` 都能用它收工）。

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"tt/internal/config"
	"tt/internal/debug"
	"tt/internal/fep"
	"tt/internal/winproc"
)

// registerRecordings 挂上录制面板的接口。
func (s *Server) registerRecordings(m *http.ServeMux) {
	m.HandleFunc("GET /api/recordings", s.hRecordingsList)
	m.HandleFunc("GET /api/recordings/current", s.hRecordingsCurrent)
	m.HandleFunc("GET /api/recordings/{id}", s.hRecordingShow)
	m.HandleFunc("POST /api/recordings", s.hRecordingStart)
	m.HandleFunc("POST /api/recordings/stop", s.hRecordingStop)
	m.HandleFunc("POST /api/recordings/{id}/replay", s.hRecordingReplay)
	m.HandleFunc("DELETE /api/recordings/{id}", s.hRecordingDelete)
	m.HandleFunc("POST /api/recordings/delete", s.hRecordingsBulkDelete)
	// 明细页的四个读侧端点（D1）：概览条带、全文流、三层模型、原始帧。
	s.registerRecordingDetail(m)
	// 回放进度（D2）：读进度 + 暂停/单步/停止（读的是回放进程落的 jsonl）。
	s.registerReplayProgress(m)
}

// recRow 是列表里的一行（前端直接照着渲染）。
type recRow struct {
	ID          string          `json:"id"`
	Where       string          `json:"where"` // 长期库 / 缓存
	Dir         string          `json:"dir"`
	Path        fep.LaunchPath  `json:"path,omitempty"`
	Prog        string          `json:"prog,omitempty"`
	Env         string          `json:"env,omitempty"`
	Zone        string          `json:"zone,omitempty"`
	Ent         string          `json:"ent,omitempty"`
	Origin      fep.Origin      `json:"origin,omitempty"`
	Retention   fep.Retention   `json:"retention,omitempty"`
	Started     string          `json:"started,omitempty"`
	Bytes       int64           `json:"bytes"`
	Conns       int             `json:"conns"`
	Fingerprint fep.Fingerprint `json:"fingerprint"`
	Note        string          `json:"note,omitempty"`
}

// recDirs 是录制件可能住的两处：长期库与缓存（见 internal/fep/retention.go）。
func (s *Server) recDirs() (dataDir, cacheDir string, lib *fep.Library, err error) {
	cfgPath, err := s.configPath(true)
	if err != nil {
		return "", "", nil, err
	}
	dataDir = filepath.Dir(cfgPath)
	cacheDir = config.CacheDir(dataDir, "fep")
	if cacheDir == "" {
		return "", "", nil, fmt.Errorf("缓存目录没解析出来（配置路径 %s）", cfgPath)
	}
	return dataDir, cacheDir, fep.NewLibrary(dataDir), nil
}

// hRecordingsList GET /api/recordings?env=&ent=
//
// 两处都列（长期库 + 缓存），按环境/ENT 过滤 —— 面板的第一屏就是它。
// 过滤按**身份里的值**（不是目录名）：库里那些确实按环境分层，但缓存里那些还没入库，
// 只能从身份里读；两边用同一套判据，前端不必知道"它住哪"。
func (s *Server) hRecordingsList(w http.ResponseWriter, r *http.Request) {
	_, cacheDir, lib, err := s.recDirs()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	env := strings.TrimSpace(r.URL.Query().Get("env"))
	ent := strings.TrimSpace(r.URL.Query().Get("ent"))

	var rows []recRow
	if items, lerr := lib.List(); lerr == nil {
		for _, it := range items {
			if it.Meta == nil {
				continue
			}
			row := rowFromMeta("长期库", it.Dir, it.Meta)
			if rowMatches(row, env, ent) {
				rows = append(rows, row)
			}
		}
	}
	if ids, cerr := fep.ListIDs(cacheDir); cerr == nil {
		for _, id := range ids {
			mt, merr := fep.ReadMeta(cacheDir, id)
			if merr != nil {
				continue // 还在录（身份是收工时落的）：不列半份
			}
			row := rowFromMeta("缓存", cacheDir, mt)
			if rowMatches(row, env, ent) {
				rows = append(rows, row)
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Started > rows[j].Started })
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "env": env, "ent": ent, "recordings": rows})
}

// hRecordingShow GET /api/recordings/{id}
func (s *Server) hRecordingShow(w http.ResponseWriter, r *http.Request) {
	_, cacheDir, lib, err := s.recDirs()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("没给录制 ID"))
		return
	}
	if it, found, lerr := lib.Find(id); lerr == nil && found && it.Meta != nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "recording": rowFromMeta("长期库", it.Dir, it.Meta), "meta": it.Meta})
		return
	}
	if mt, merr := fep.ReadMeta(cacheDir, id); merr == nil {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "recording": rowFromMeta("缓存", cacheDir, mt), "meta": mt})
		return
	}
	writeErr(w, http.StatusNotFound, fmt.Errorf("找不到录制 %s（长期库与缓存都没有）", id))
}

// hRecordingsCurrent GET /api/recordings/current —— 现在有没有在录。
//
// 判据是**会话登记**（`.tt-fep.json`）：它是"这一轮还活着"的唯一权威，
// 而录制件本身要等收工才落身份，看文件只会在录到一半时给出"没有"。
//
// 但登记会**比进程活得久**（`rec reproduce --no-wait` 发完就退，留下登记 + 一个还开着的
// gdc 窗口），所以除了 id 之外还要把两样事实说清楚：
//   - kind：这一轮是录还是回放（登记里那一栏，见 internal/fep/session.go）；
//   - stale：起它的那个进程已经不在（= 只剩窗口，面板该说"收工/关掉"而不是"正在录"）。
func (s *Server) hRecordingsCurrent(w http.ResponseWriter, r *http.Request) {
	dataDir, _, _, err := s.recDirs()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	// 多条登记时这一格给**最新的一条活登记**（面板那一栏问的是"现在在录什么"）；
	// 并发录制的完整清单在 `/api/recordings` 列表里，按 env/ent 过滤。
	list, lerr := fep.NewRegistry(dataDir).List()
	if lerr != nil || len(list) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "recording": nil})
		return
	}
	var sess *fep.Session
	for _, s := range list {
		if s == nil || s.Stale() {
			continue
		}
		if sess == nil || s.Started.After(sess.Started) {
			sess = s
		}
	}
	if sess == nil { // 全是过期登记：给最后一条，至少面板能说出这是谁留下的
		sess = list[len(list)-1]
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "recording": map[string]any{
		"id": sess.Tag, "env": sess.Env, "zone": sess.Zone, "ent": sess.Ent,
		"prog": sess.Prog, "path": sess.Path, "started": sess.Started,
		"kind": sess.Kind, "stale": sess.Stale(),
	}})
}

// recStartReq 是"开始录一轮"的请求体。
type recStartReq struct {
	Env       string `json:"env"`
	Prog      string `json:"prog"`
	Ent       string `json:"ent"`
	Module    string `json:"module"`
	BinaryRef string `json:"binaryRef"`
}

// busySession 读会话登记，并回答"这个环境里是不是**真的还占着**"。
//
//   - (sess, true)：起它的进程还活着 —— 它占着三个端口与调试会话，不能起新的；
//   - (nil, false)：没有该环境的活登记（含"只剩过期登记"的那一种）。
//
// 为什么要按环境过滤（多会话）：一条录制占的是**一个环境**的一组端口与一条调试会话。
// 从前登记只有一条，所以"有人占着"必然就是"这个环境被占着"；多会话之后这两件事分开了 ——
// 全局判断会把另一个环境的录制也算成占用，于是"Agent A 在 109 录、Agent B 想在同一台机器
// 的另一个环境录"被无故挡下。
//
// env 为空 = 不按环境过滤（保持旧的调用语义：那些调用问的是"有没有任何一轮在跑"）。
func busySession(dataDir, env string) (*fep.Session, bool) {
	list, err := fep.NewRegistry(dataDir).List()
	if err != nil {
		return nil, false
	}
	want := strings.TrimSpace(env)
	for _, s := range list {
		if s == nil || s.Stale() {
			continue
		}
		if want != "" && s.Env != want {
			continue
		}
		return s, true
	}
	return nil, false
}

// hRecordingStart POST /api/recordings
//
// 起一个脱离的 `tt recording start`（见文件头注释），等它把登记写出来，返回录制 ID。
// 用户此时**不会**被带到任何调试器面前：那一侧走的是我们自己的 fgldb（`rec start` 里那段）。
func (s *Server) hRecordingStart(w http.ResponseWriter, r *http.Request) {
	cfgPath, err := s.configPath(true)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	var req recStartReq
	if derr := decodeBody(r, &req); derr != nil {
		writeErr(w, http.StatusBadRequest, derr)
		return
	}
	if strings.TrimSpace(req.Env) == "" || strings.TrimSpace(req.Prog) == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("环境与作业都要给（作业落在哪个企业只由 ENT 决定，作业名不能空）"))
		return
	}
	dataDir := filepath.Dir(cfgPath)
	// 已经在录了就不要再起一个：两个会话会抢同一组端口（7400/7600/7899），
	// 而且用户看到的是"点了一下没反应"。**过期登记不算在录**（见 busySession）。
	reg := fep.NewRegistry(dataDir)
	prevTag := ""
	if sess, busy := busySession(dataDir, req.Env); busy {
		writeErr(w, http.StatusConflict,
			fmt.Errorf("已经在录了（%s，%s）：先停掉它再起新的", sess.Tag, sess.Prog))
		return
	} else if sess != nil {
		// 只留旧登记不算数：下面"等它登记"会立刻读到它，把**旧 ID** 当成新一轮报给用户。
		// 所以记住它是谁，等一个不一样的那一份（旧登记本身留着 —— 它是那个窗口的收工凭据）。
		prevTag = sess.Tag
	}
	exe, eerr := os.Executable()
	if eerr != nil {
		writeErr(w, http.StatusInternalServerError, eerr)
		return
	}
	// 日志落缓存目录，SpawnDetached 只 OpenFile 不建目录：全新装好、还没录过一轮时
	// 这个目录**不存在**，不建就会以"系统找不到指定的路径"起不来。
	cacheDir := config.CacheDir(dataDir, "fep")
	if merr := os.MkdirAll(cacheDir, 0o755); merr != nil {
		writeErr(w, http.StatusInternalServerError, merr)
		return
	}
	logPath := filepath.Join(cacheDir, "rec-start-"+time.Now().Format("20060102-150405")+".log")
	// 起的是 `tt recording start`（与命令面同一份实现，见 internal/cli/recording）。
	// 面板与命令行走同一个入口，不存在"面板这条、命令行那条"两套行为。
	args := []string{"recording", "start", "--env", req.Env, "--prog", req.Prog, "--config", cfgPath}
	if e := strings.TrimSpace(req.Ent); e != "" {
		args = append(args, "--ent", e)
	}
	if m := strings.TrimSpace(req.Module); m != "" {
		args = append(args, "--module", m)
	}
	if b := strings.TrimSpace(req.BinaryRef); b != "" {
		args = append(args, "--binary-ref", b)
	}
	spawnedAt := time.Now()
	if _, perr := spawnAndCheck(exe, args, logPath); perr != nil {
		writeErr(w, http.StatusInternalServerError, fmt.Errorf("起录制失败: %w", perr))
		return
	}
	// 等它登记（登记是它能干活之前的第一件事，所以这就是"就绪"的信号）。
	// 判据是"读到一份**不是上一轮**的登记"：上一轮那份可能还在（过期但没清，见 prevTag）。
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if sess, lerr := reg.Load(); lerr == nil && sess != nil &&
			(sess.Tag != prevTag || sess.Started.After(spawnedAt)) {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "recording": map[string]any{
				"id": sess.Tag, "env": sess.Env, "prog": sess.Prog, "ent": sess.Ent, "path": sess.Path,
			}, "log": logPath})
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
	// 超时要**把真因带回去**：真正的原因（端口被占、隔离实例起不来、TOPENT 设不上、
	// 被模式闸门拦下…）都写在那条日志里，只说一句"看日志"等于把查的人赶去别处
	// （实测：真因是"端口被占：隔离实例在 20s 内没在 8400 端口监听"，而文案却说"进程起来了"，
	// 与实际相反）。优先挑日志里**第一处** error（最早的才是真因），挑不到才退回尾巴。
	if msg := firstLogError(logPath); msg != "" {
		writeErr(w, http.StatusGatewayTimeout,
			fmt.Errorf("录制进程 30 秒内没登记（它多半已经退了）：%s（完整日志：%s）", msg, logPath))
		return
	}
	writeErr(w, http.StatusGatewayTimeout,
		fmt.Errorf("录制进程 30 秒内没登记（它多半已经退了）：%s", tailOf(logPath, 400)))
}

// hRecordingStop POST /api/recordings/stop
//
// 收工 = 请常驻进程自己退（它会落身份 + 搬进长期库）+ 收掉调试会话。
// 与命令行的 `tt recording stop <录制ID>` 同一件事，只是由面板触发。
func (s *Server) hRecordingStop(w http.ResponseWriter, r *http.Request) {
	dataDir, _, _, err := s.recDirs()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	reg := fep.NewRegistry(dataDir)
	sess, lerr := reg.Load()
	if lerr != nil || sess == nil {
		writeErr(w, http.StatusConflict, fmt.Errorf("现在没有在录的会话"))
		return
	}
	stopped, serr := reg.Stop()
	if serr != nil {
		writeErr(w, http.StatusInternalServerError, serr)
		return
	}
	// 调试线那一侧也要收（这一轮是 path=debug 起的）：只收代理会让它在服务里留成一个
	// 开着的会话，下一轮 `debug topent` 就设不上去。
	if stopped.Path == fep.PathDebug {
		// 用**真的那份**配置文件（`--config` 可能指向别处，不猜 dataDir/config.json）；
		// 日志目录同起录那条路，缺了先建。
		if cfgPath, cerr := s.configPath(true); cerr == nil {
			if exe, eerr := os.Executable(); eerr == nil {
				if cacheDir := config.CacheDir(filepath.Dir(cfgPath), "fep"); cacheDir != "" {
					if merr := os.MkdirAll(cacheDir, 0o755); merr == nil {
						logPath := filepath.Join(cacheDir, "rec-stop-"+time.Now().Format("20060102-150405")+".log")
						_, _ = winproc.SpawnDetached(exe, []string{"debug", "quit", "--config", cfgPath}, logPath, debug.ActorEnv+"=human")
					}
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "stopped": map[string]any{
		"id": stopped.Tag, "env": stopped.Env, "prog": stopped.Prog,
	}})
}

// spawnAndCheck 起一个脱离的子进程，并给它一个"起手就退"的检查窗口。
//
// 为什么需要：这两个按钮的动作都是**另起进程**，而子进程死在第一件事上（端口用不了、
// 配置读不到、被模式闸门拦下）时，界面这边什么都看不出来 —— 调用方拿到的是"已开始…"，
// 然后永远没有下文（实测 2026-10-16：回放点了三次都"没反应"，真因只在日志里）。
//
// 窗口从 1.2 秒放宽到 6 秒：真正会失败的那几步（起隔离实例 → 切环境 → 设 TOPENT → 起作业）
// 大多在 2~5 秒内出结果，1.2 秒太短、失败被当成"已开始"返回。进程活着就立刻返回，
// 正常路径不白等；死了就带着**日志里第一处真因**回去。
func spawnAndCheck(exe string, args []string, logPath string) (int, error) {
	// TT_ACTOR=human：这两个按钮都是**人在界面上点的**，而链路是"serve 起 tt 子进程 →
	// 子进程调调试线 REST"，CLI 默认声明自己是 ai —— 在纯人工模式下会被模式闸门拦下
	// （实测 2026-10-16：点「开始录制」直接报"纯人工模式:AI 不可操作"）。
	pid, err := winproc.SpawnDetached(exe, args, logPath, debug.ActorEnv+"=human")
	if err != nil {
		return 0, err
	}
	for waited := time.Duration(0); waited < 6*time.Second; waited += 200 * time.Millisecond {
		time.Sleep(200 * time.Millisecond)
		if !winproc.Alive(pid) {
			return pid, spawnFailErr(pid, logPath)
		}
	}
	return pid, nil
}

// spawnFailErr 子进程起来就退时的错误文案：优先日志里**第一处** error（最早的才是真因）。
func spawnFailErr(pid int, logPath string) error {
	if msg := firstLogError(logPath); msg != "" {
		return fmt.Errorf("进程起来就退了（pid %d）：%s（完整日志：%s）", pid, msg, logPath)
	}
	return fmt.Errorf("进程起来就退了（pid %d）：%s", pid, tailOf(logPath, 300))
}

// firstLogError 从那条日志里挑出**第一处真正的错误**（`"error": "..."`）给错误信息用。
//
// 为什么不直接甩日志尾巴：那条日志里通常有**两个** JSON（内层某一步的失败 + 最外层的汇总），
// 400 字的尾巴把两个都塞进弹窗，读起来是两段互相打架的话 —— 实测 2026-10-16 用户看到的是
// "没能入长期库…" 打头、真因（"纯人工模式:AI 不可操作"）埋在中间。挑第一处，
// 它就是最早、也就是真正的原因。
func firstLogError(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, ln := range strings.Split(string(b), "\n") {
		s := strings.TrimSpace(ln)
		if !strings.HasPrefix(s, `"error":`) {
			continue
		}
		v := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(s, `"error":`)), ","))
		if u, uerr := strconv.Unquote(v); uerr == nil {
			v = u
		}
		if v != "" {
			return v
		}
	}
	return ""
}

// tailOf 取日志末尾若干字符（只给错误信息用；读不到就照实说）。
func tailOf(path string, n int) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return "（读不到日志 " + path + "）"
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return "（日志是空的）"
	}
	if r := []rune(s); len(r) > n {
		s = "…" + string(r[len(r)-n:])
	}
	return s
}

// hRecordingReplay POST /api/recordings/{id}/replay
//
// 「回放」按钮：**另起一个脱离的进程**跑 `tt recording replay <ID> --no-wait`
// （与起录同一条先例）。为什么是另起进程而不是在 serve 里做：回放自己要走一遍
// "起调试线会话 + 过指纹闸 + 把事件发出去"，那是一轮**前台**的活，和录制一样天然属于一个
// 独立进程；serve 只负责把它拉起来并回报日志落点。
//
// `--no-wait`：面板不带断点（断点归 AI 在会话上接着下），等停站只会等到超时、看着像失败，
// 而那时事件其实已经原样发出去了。会话照样留着给后面取证。
//
// 两种情形直接拒，不浪费一次起进程：
//   - r.r 录的件：那条路不带 -d，编号表与调试线不是同一张（跨路径实测共同编号 0 条相同），
//     重放会点错控件；
//   - 已经有一轮录制在跑：两边要占同一组端口（隔离实例/代理）与同一个调试会话。
//     过期登记（回放发完留下的那份）不算"在跑"，见 busySession。
func (s *Server) hRecordingReplay(w http.ResponseWriter, r *http.Request) {
	cfgPath, err := s.configPath(true)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("没给录制 ID"))
		return
	}
	dataDir := filepath.Dir(cfgPath)
	_, cacheDir, lib, derr := s.recDirs()
	if derr != nil {
		writeErr(w, http.StatusInternalServerError, derr)
		return
	}
	mt, found := findRecordingMeta(lib, cacheDir, id)
	if !found {
		writeErr(w, http.StatusNotFound, fmt.Errorf("找不到录制 %s（长期库与缓存都没有）", id))
		return
	}
	if mt.Path == fep.PathRunnerRR {
		writeErr(w, http.StatusConflict,
			fmt.Errorf("这份录制是 r.r 录的：编号表与调试线不是同一张，重放会点错控件。"+
				"要在断点下重放，请用面板「开始录制」录的那一份（走 debug 线）"))
		return
	}
	if sess, busy := busySession(dataDir, mt.Env); busy {
		writeErr(w, http.StatusConflict,
			fmt.Errorf("现在有一轮录制在跑（%s，%s）：先停掉它再回放（两边要占同一组端口与调试会话）",
				sess.Tag, sess.Prog))
		return
	}
	exe, eerr := os.Executable()
	if eerr != nil {
		writeErr(w, http.StatusInternalServerError, eerr)
		return
	}
	// 日志落缓存目录；SpawnDetached 只 OpenFile 不建目录（全新装好时它还不存在）
	if merr := os.MkdirAll(cacheDir, 0o755); merr != nil {
		writeErr(w, http.StatusInternalServerError, merr)
		return
	}
	logPath := filepath.Join(cacheDir, "rec-replay-"+time.Now().Format("20060102-150405")+".log")
	pid, perr := spawnAndCheck(exe,
		[]string{"recording", "replay", id, "--no-wait", "--config", cfgPath}, logPath)
	if perr != nil {
		writeErr(w, http.StatusInternalServerError, fmt.Errorf("起回放失败: %w", perr))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "id": id, "pid": pid, "log": logPath,
		"note": "回放已交给后台进程：它按身份起调试线会话、过指纹闸、把事件原样发出去（不等停站）；" +
			"进度看这个日志，会话留着给 AI 取证",
	})
}

// findRecordingMeta 按 ID 找一份录制的身份：长期库优先，再缓存。
func findRecordingMeta(lib *fep.Library, cacheDir, id string) (*fep.RecordingMeta, bool) {
	if it, found, err := lib.Find(id); err == nil && found && it.Meta != nil {
		return it.Meta, true
	}
	if mt, err := fep.ReadMeta(cacheDir, id); err == nil {
		return mt, true
	}
	return nil, false
}

// hRecordingDelete DELETE /api/recordings/{id}?confirm=1
//
// 删**长期库**里的那一份（前端"手动清除"）。必须显式 confirm：录制不可再生
// （用户再报一次 bug、再操作一遍才有这一份），与命令行那条 `rec rm --yes` 同一口径。
func (s *Server) hRecordingDelete(w http.ResponseWriter, r *http.Request) {
	_, cacheDir, lib, err := s.recDirs()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if r.URL.Query().Get("confirm") != "1" {
		writeErr(w, http.StatusBadRequest,
			fmt.Errorf("删录制要二次确认（?confirm=1）：这份录制不可再生"))
		return
	}
	item, found, lerr := lib.Find(id)
	if lerr != nil {
		writeErr(w, http.StatusInternalServerError, lerr)
		return
	}
	if !found {
		// 在缓存里？如实说它归统一清理框架管
		if files, ferr := fep.RecordingFilesFor(cacheDir, id); ferr == nil && len(files) > 0 {
			writeErr(w, http.StatusConflict,
				fmt.Errorf("%s 在缓存目录，归统一清理框架管（tt cache clear / 启动清理按年龄删）", id))
			return
		}
		writeErr(w, http.StatusNotFound, fmt.Errorf("长期库里没有 %s", id))
		return
	}
	files, bytes, rerr := lib.Remove(item)
	if rerr != nil {
		writeErr(w, http.StatusInternalServerError, rerr)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id, "files": files, "bytes": bytes})
}

// hRecordingsBulkDelete POST /api/recordings/delete
//
// 批量彻底删除（用户 2026-10-18 定）：**长期库与缓存都删**，一次把勾中的清干净。
//
// 与单份那条（`DELETE /api/recordings/{id}`）的分工：单份是"在明细页里清掉这一份"，
// 只碰长期库（缓存归统一清理框架）；批量是"管理模式下勾一批，彻底清掉"，两边都碰。
//
// 必须显式 confirm：录制不可再生，删了就再也造不出来。
func (s *Server) hRecordingsBulkDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs     []string `json:"ids"`
		Confirm bool     `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("请求体不是合法 JSON: %w", err))
		return
	}
	if !body.Confirm {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("批量删除要二次确认（confirm:true）：这些录制不可再生"))
		return
	}
	_, cacheDir, lib, err := s.recDirs()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	type delOne struct {
		ID     string `json:"id"`
		Where  string `json:"where,omitempty"` // 长期库 / 缓存 / 两处
		Files  int    `json:"files"`
		Bytes  int64  `json:"bytes"`
		Failed string `json:"failed,omitempty"`
	}
	seen := map[string]bool{}
	out := make([]delOne, 0, len(body.IDs))
	for _, raw := range body.IDs {
		id := strings.TrimSpace(raw)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		// id 是**目录/文件名**的一部分，必须挡住路径穿越与通配：只收录制 ID 的形状。
		if strings.ContainsAny(id, `\\/:*?"<>|`) || strings.Contains(id, "..") {
			out = append(out, delOne{ID: id, Failed: "不是合法的录制 ID"})
			continue
		}
		one := delOne{ID: id}
		var where []string
		if item, found, lerr := lib.Find(id); lerr == nil && found {
			if n, b, rerr := lib.Remove(item); rerr != nil {
				one.Failed = "删长期库失败: " + rerr.Error()
			} else {
				one.Files += n
				one.Bytes += b
				where = append(where, "长期库")
			}
		}
		// 缓存：同一份录制在缓存里的文件是 `<id>-*.log` / `<id>.meta.json`（见 fep 的落盘命名）。
		if n, b, cerr := removeCacheRecording(cacheDir, id); cerr != nil {
			one.Failed = strings.TrimSpace(one.Failed + "；删缓存失败: " + cerr.Error())
		} else if n > 0 {
			one.Files += n
			one.Bytes += b
			where = append(where, "缓存")
		}
		if one.Failed == "" && len(where) == 0 {
			out = append(out, delOne{ID: id, Failed: "长期库与缓存里都没有这一份"})
			continue
		}
		one.Where = strings.Join(where, " + ")
		out = append(out, one)
	}
	okN, failN := 0, 0
	for _, d := range out {
		if d.Failed != "" {
			failN++
		} else {
			okN++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": failN == 0, "deleted": okN, "failed": failN, "items": out,
	})
}

// removeCacheRecording 删掉缓存目录里属于这一份录制的文件（严格按 id 前缀匹配）。
//
// 只认两种命名：`<id>-…`（各连接的双向日志/时间线/gdc 输出）与 `<id>.meta.json`。
// 不认"id 是别人前缀"的情形（`fep-1` 不会吃掉 `fep-10-…`）：判据在分隔符上。
func removeCacheRecording(dir, id string) (files int, bytes int64, err error) {
	ents, rerr := os.ReadDir(dir)
	if rerr != nil {
		return 0, 0, rerr
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if name != id+".meta.json" && !strings.HasPrefix(name, id+"-") {
			continue
		}
		p := filepath.Join(dir, name)
		info, ierr := os.Stat(p)
		if ierr == nil {
			bytes += info.Size()
		}
		if rmErr := os.Remove(p); rmErr != nil {
			return files, bytes, rmErr
		}
		files++
	}
	return files, bytes, nil
}

// rowFromMeta 把身份摊成列表里的一行（体积现算：它就在手边，而"这份多大"是用户会看的）。
func rowFromMeta(where, dir string, mt *fep.RecordingMeta) recRow {
	row := recRow{
		ID: mt.ID, Where: where, Dir: dir,
		Path: mt.Path, Prog: mt.Prog, Env: mt.Env, Zone: mt.Zone, Ent: mt.Ent,
		Origin: mt.Origin, Retention: mt.Retention, Fingerprint: mt.Fingerprint,
		Conns: len(mt.Conns), Note: mt.Note,
	}
	if !mt.Started.IsZero() {
		row.Started = mt.Started.Format(time.RFC3339)
	}
	for _, c := range mt.Conns {
		row.Bytes += c.BytesRT + c.BytesFE
	}
	return row
}

// rowMatches 按环境/ENT 过滤（空条件 = 不过滤）。大小写不敏感：环境名是人打的。
func rowMatches(row recRow, env, ent string) bool {
	if env != "" && !strings.EqualFold(strings.TrimSpace(row.Env), env) {
		return false
	}
	if ent != "" && strings.TrimSpace(row.Ent) != ent {
		return false
	}
	return true
}
