package web

// 回放进度（计划 C 的 D2/D7 的服务端）。
//
// 一件事：**回放跑在另一个进程里**（面板按钮走 `winproc.SpawnDetached`，CLI 也会起），
// 与本服务没有共享内存。它唯一留下的、双方都够得着的东西就是那份逐行 flush 的
// `<ID>-replay-<时刻>.jsonl`（见 internal/fep/replaylog.go）。所以进度是**读文件**读出来的：
//
//	GET  /api/recordings/{id}/replay-steps?since=N[&watch=1]  读进度（since = 已拿到多少条）
//	POST /api/recordings/{id}/replay-ctl {cmd}                暂停 / 继续 / 单步 / 停止
//
// 两条通道，各有各的位置：
//
//	**轮询 REST** 是权威 —— 断线、刷新、页面刚打开都靠它，`since` 保证不丢步也不重放。
//	**WS `rec.progress`** 只是"叫一声"，让界面立刻去读一次（省掉最多一个轮询周期的延迟）。
//
// `watch=1` 的那一次请求顺手把服务端的"盯盘"循环启动起来（见 recWatcher）：
// 它每 400ms 看一次日志，把新行通过 Options.RecProgress 推给页面。没有客户端再问、
// 或者这一轮已经收尾，它就自己退场 —— 不留常驻 goroutine。

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"tt/internal/fep"
)

// replayWatchInterval 是服务端盯盘的间隔。400ms：比人手快、比磁盘累得慢。
const replayWatchInterval = 400 * time.Millisecond

// replayWatchIdle 是"多久没有新行就收工"。给足收尾的时间（最后一步的回包最多等 8s）。
const replayWatchIdle = 30 * time.Second

// registerReplayProgress 挂上回放进度的两个端点。
func (s *Server) registerReplayProgress(m *http.ServeMux) {
	m.HandleFunc("GET /api/recordings/{id}/replay-steps", s.hReplaySteps)
	m.HandleFunc("POST /api/recordings/{id}/replay-ctl", s.hReplayCtl)
}

// hReplaySteps GET /api/recordings/{id}/replay-steps?since=N&watch=1
//
// `since` 是**已经拿到多少条**（步数游标，不是 `n`）：跳过与暂停都会产生额外的行，
// 用 `n` 当游标会丢步（见 internal/fep/replaylog.go 的纪律 3）。
func (s *Server) hReplaySteps(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("没给录制 ID"))
		return
	}
	since := atoiDefault(r.URL.Query().Get("since"), 0)
	if since < 0 {
		since = 0
	}
	if r.URL.Query().Get("watch") == "1" {
		s.watchReplay(id)
	}
	_, cacheDir, _, err := s.recDirs()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	lg, lerr := fep.LatestReplayLog(cacheDir, id)
	if lerr != nil {
		if errors.Is(lerr, fep.ErrNoReplayLog) {
			// 没回过放**不是错误**：界面据此把进度条整块收起来（而不是弹一个红条）。
			writeJSON(w, http.StatusOK, map[string]any{
				"ok": true, "id": id, "hasLog": false,
				"steps": []fep.ReplayStep{}, "since": since,
			})
			return
		}
		writeErr(w, http.StatusInternalServerError, lerr)
		return
	}
	running := replayRunning(s, id)
	out := map[string]any{
		"ok": true, "id": id, "hasLog": true,
		"path":       lg.Path,
		"running":    running,
		"total":      lg.Total,
		"since":      since,
		"from":       since,
		"count":      len(lg.Steps),
		"progress":   lg.Progress(),
		"startedAt":  msOrZero(lg.Started),
		"finishedAt": msOrZero(lg.Finished),
		"steps":      lg.From(since),
		"states":     lg.States(),
		"note":       lg.Note,
		"bad":        lg.Bad,
		"ctlPath":    fep.ReplayCtlPath(lg.Path),
	}
	if lg.Done && running {
		// 日志收了尾但登记还在：多半是"回放发完就退"留下的那份（--no-wait）。
		out["note"] = strings.TrimSpace(lg.Note + " 回放进程已经发完（登记与窗口可能还在）")
	}
	writeJSON(w, http.StatusOK, out)
}

// replayRunning 判断"这一份录制现在有没有一轮回放在跑"。
//
// 判据是**会话登记**（`.tt-fep.json`）＋ kind：日志的尾行在不在只能说明"写到哪了"，
// 而"是不是还在跑"只有登记知道（进程可能被杀在半路，日志永远等不到尾行）。
func replayRunning(s *Server, id string) bool {
	// 判据改成**回放日志本身**（进度条用的是同一个事实）：尾行没落 = 还在跑。
	//
	// 为什么不再看登记：回放进程登记的是**这一轮新录的那份**（tag = 新录制 ID），
	// 不是被回放的源录制 ID —— 拿 `sess.Tag == id` 判永远为假，面板上的
	// 暂停/单步/跳过就全被 409 挡在门外（真机 2026-10-18 18:16 现场）。
	dataDir, cacheDir, _, err := s.recDirs()
	if err != nil {
		return false
	}
	// ① 日志还在写（尾行没落）= 这一轮真的在跑。这是**主判据** —— 进度条用的也是它。
	if lg, lerr := fep.LatestReplayLog(cacheDir, id); lerr == nil && !lg.Done {
		return true
	}
	// ② 日志已收尾（或那一轮被杀在半路）：再看登记里有没有活着的 replay 会话。
	//    登记的 tag 是"这一轮新录的那份"，不是源录制 ID，所以这里只能按 kind 判 ——
	//    宁可松一点，也不能让面板上的暂停/跳过全被 409 挡掉（那正是真机撞到的样子）。
	sess, lerr := fep.NewRegistry(dataDir).Load()
	if lerr != nil || sess == nil {
		return false
	}
	return sess.Kind == fep.SessionKindReplay && !sess.Stale()
}

// hReplayCtl POST /api/recordings/{id}/replay-ctl {cmd: pause|resume|step|stop}
//
// 把人的命令写进 `<日志>.ctl`，由回放进程在**下一步之前**读一次。
//
// 为什么不是"直接给回放进程发信号"：它是另一个进程（还可能不是本服务起的），
// 而我们已经有"缓存目录 + 录制 ID"这个双方都知道的坐标。开端口要处理占用与鉴权，
// 一个一两百毫秒轮询一次的小文件语义清楚、进程死了也不留监听。
func (s *Server) hReplayCtl(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	var req struct {
		Cmd string `json:"cmd"`
	}
	if derr := decodeBody(r, &req); derr != nil {
		writeErr(w, http.StatusBadRequest, derr)
		return
	}
	cmd := fep.ReplayCtl(strings.ToLower(strings.TrimSpace(req.Cmd)))
	switch cmd {
	case fep.CtlPause, fep.CtlResume, fep.CtlStep, fep.CtlSkip, fep.CtlStop:
	default:
		writeErr(w, http.StatusBadRequest, fmt.Errorf("cmd 只能是 pause / resume / step / skip / stop（收到 %q）", req.Cmd))
		return
	}
	_, cacheDir, _, err := s.recDirs()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	lg, lerr := fep.LatestReplayLog(cacheDir, id)
	if lerr != nil {
		writeErr(w, http.StatusNotFound, fmt.Errorf("这一份录制没有回放日志（回放还没起过，或已经清过缓存）"))
		return
	}
	if !replayRunning(s, id) && cmd != fep.CtlStop {
		// 停止永远允许（收尾也要能点）；其余命令对着一轮已经结束的回放没有意义。
		writeErr(w, http.StatusConflict, fmt.Errorf("这一轮回放已经不在跑了（%s）：命令没有接收方", id))
		return
	}
	if werr := fep.WriteReplayCtl(lg.Path, cmd); werr != nil {
		writeErr(w, http.StatusInternalServerError, werr)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true, "id": id, "cmd": string(cmd), "ctl": fep.ReplayCtlPath(lg.Path),
		"note": ctlNote(cmd),
	})
}

// ctlNote 说清"这个命令什么时候生效"—— 暂停是在**下一步之前**生效，
// 而当前这一步可能正在等运行期回包（最长 8 秒）。不说清的话，用户会以为按钮没反应。
func ctlNote(cmd fep.ReplayCtl) string {
	if cmd == fep.CtlSkip {
		return "跳过：这一步**不发**，继续停在暂停态等下一步"
	}
	switch cmd {
	case fep.CtlPause:
		return "暂停会在当前这一步的回包之后生效（它可能正在等运行期，最长 8 秒）"
	case fep.CtlStep:
		return "单步：放行当前这一步，然后重新停下"
	case fep.CtlResume:
		return "继续：从停住的那一步接着发"
	default:
		return "停止：不再发后面的步骤，已经发出去的收不回来"
	}
}

// msOrZero 把 Unix 毫秒原样给出（0 = 没记）。
func msOrZero(v int64) int64 { return v }

// ---- 服务端的"盯盘"循环 ----

// recWatcher 盯着一份录制的回放日志，把新行推给页面（WS），并在没人看时退场。
type recWatcher struct {
	mu   sync.Mutex
	byID map[string]*watchHandle
}

// watchHandle 是一个盯盘循环的把手：停下它、并等它真的停下。
//
// 为什么要等（done 通道）：循环体里可能正拿着日志文件在读 —— 不等它退到安全点就返回，
// 调用方（测试的 TempDir 清理、或者停机清缓存）会撞上"文件被占用"（Windows 实测）。
type watchHandle struct {
	stop chan struct{}
	done chan struct{}
}

// watchReplay 确保这一份录制有一个盯盘循环在跑（幂等：已经在跑就什么都不做）。
//
// **由客户端请求启动**（`?watch=1`）而不是常驻扫描缓存目录：没人看的时候不需要推，
// 而"谁在看"这件事只有请求知道。没有新行 + 登记也没了 → 循环自己退出。
func (s *Server) watchReplay(id string) {
	if s.opt.RecProgress == nil {
		return // 没接广播（例如单独跑 web 的测试）：界面靠轮询，不启循环
	}
	if s.watch == nil {
		s.watch = &recWatcher{byID: map[string]*watchHandle{}}
	}
	s.watch.mu.Lock()
	if _, ok := s.watch.byID[id]; ok {
		s.watch.mu.Unlock()
		return
	}
	h := &watchHandle{stop: make(chan struct{}), done: make(chan struct{})}
	s.watch.byID[id] = h
	s.watch.mu.Unlock()
	go s.replayWatchLoop(id, h)
}

// stopWatches 停掉所有盯盘循环并**等它们退到安全点**（幂等）。
//
// 用途有二：测试（不然一个要跑 30 秒的 goroutine 会让 TempDir 删不掉）与服务停机收尾。
func (s *Server) stopWatches() {
	if s.watch == nil {
		return
	}
	s.watch.mu.Lock()
	handles := make([]*watchHandle, 0, len(s.watch.byID))
	for id, h := range s.watch.byID {
		handles = append(handles, h)
		delete(s.watch.byID, id)
	}
	s.watch.mu.Unlock()
	for _, h := range handles {
		close(h.stop)
	}
	for _, h := range handles {
		select {
		case <-h.done:
		case <-time.After(3 * time.Second): // 兜底：绝不让收尾把调用方挂住
		}
	}
}

// replayWatchLoop 是那个循环：每 400ms 看一次日志，把新行推出去。
func (s *Server) replayWatchLoop(id string, h *watchHandle) {
	defer close(h.done)
	cursor := 0
	curPath := ""
	lastSeen := time.Now()
	tick := time.NewTicker(replayWatchInterval)
	defer tick.Stop()
	for {
		select {
		case <-h.stop:
			return
		case <-tick.C:
		}
		_, cacheDir, _, err := s.recDirs()
		if err != nil {
			return
		}
		lg, lerr := fep.LatestReplayLog(cacheDir, id)
		if lerr != nil {
			// 还没有日志：等它出现（回放进程刚起来时会有一小段空窗）。
			if time.Since(lastSeen) > replayWatchIdle {
				return
			}
			continue
		}
		// **换了一份日志、或者同一份被截断重写** → 游标必须归零。
		//
		// 不归零的后果实测过：同一份录制回放第二轮时，`LatestReplayLog` 拿到的是新文件，
		// 而循环里的 cursor 还停在上一轮的行数上 —— 新日志的行数比它少，于是**一行都不推**，
		// 界面看着像"回放没在动"（这条路是拿合成日志连跑两遍才照出来的）。
		if lg.Path != curPath || len(lg.Steps) < cursor {
			curPath, cursor = lg.Path, 0
			lastSeen = time.Now()
		}
		if n := len(lg.Steps); n > cursor {
			for _, st := range lg.From(cursor) {
				s.opt.RecProgress(id, st.N, st.Total, st.Result)
			}
			cursor = n
			lastSeen = time.Now()
		}
		if lg.Done && !replayRunning(s, id) && time.Since(lastSeen) > 2*replayWatchInterval {
			return // 收尾了、登记也没了：再等一拍确认没有追加，就退场
		}
		if time.Since(lastSeen) > replayWatchIdle {
			return
		}
	}
}
