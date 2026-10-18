package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"tt/internal/fep"
)

// 回放进度（计划 C 的 D2/D7 的服务端）。
//
// 这里钉四件事，都是"错了会很难查"的那类：
//  1. **没回过放不是错误**（`hasLog:false`）：界面据此把进度条整块收起来，而不是弹红条；
//  2. `since` 按**行位置**分页，**不丢步**（跳过/暂停会产生额外的行）；
//  3. 控制命令的三种拒绝：非法词 400、没有日志 404、不在跑 409（但**停止永远允许**）；
//  4. 盯盘循环真的会把新行推出去（注入的 RecProgress 收得到）。

// seedReplayLog 在缓存目录里落一份**已收尾**的回放日志，返回它的路径。
//
// total 取步骤自己带的 `Total`（不是行数）：一行不等于一步（中间态、暂停都会多出行），
// 而"总共几步"是计划给的数。
func seedReplayLog(t *testing.T, dataDir, id string, steps []fep.ReplayStep, res fep.ReplayResult) string {
	t.Helper()
	cache := filepath.Join(dataDir, "fep")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	total := len(steps)
	if len(steps) > 0 && steps[0].Total > 0 {
		total = steps[0].Total
	}
	path := filepath.Join(cache, fep.ReplayLogName(id, time.Now()))
	w, err := fep.NewReplayLogWriter(path, id, 1, total)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range steps {
		w.Step(s)
	}
	if err := w.Close(res, "跑完"); err != nil {
		t.Fatal(err)
	}
	return path
}

// seedLiveReplay 让"这一份录制现在有一轮回放在跑"成立（会话登记 + 本进程的 pid）。
func seedLiveReplay(t *testing.T, dataDir, id string) {
	t.Helper()
	reg := fep.NewRegistry(dataDir)
	if err := reg.Save(&fep.Session{
		Env: "主机109", Zone: "36", Prog: "bsft001_wf", Ent: "99",
		Path: fep.PathDebug, Kind: fep.SessionKindReplay, Tag: id,
		OwnerPID: os.Getpid(), Started: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
}

func postJSON(t *testing.T, s *Server, target, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestReplayStepsWithoutLogIsNotAnError(t *testing.T) {
	s, _ := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	code, out := getJSON(t, s, "/api/recordings/fep-none/replay-steps")
	if code != http.StatusOK {
		t.Fatalf("没回过放不是错误：%d %v", code, out)
	}
	if out["hasLog"] != false {
		t.Fatalf("该明说 hasLog=false（界面据此把进度条收起来）：%v", out)
	}
	if steps, ok := out["steps"].([]any); !ok || len(steps) != 0 {
		t.Fatalf("没日志时 steps 该是空数组（不是 null，前端少一层判空）：%v", out["steps"])
	}
}

func TestReplayStepsPagingAndProgress(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	dataDir := filepath.Dir(cfgPath)
	at := func(ms int64) *int64 { return &ms }
	seedReplayLog(t, dataDir, "fep-rep", []fep.ReplayStep{
		{N: 1, Total: 3, Event: `ActionEvent 0{{idRef "42"}}`, Kind: "ActionEvent", Target: 42, SentAt: at(1000), Result: fep.ReplayWaiting},
		{N: 1, Total: 3, Event: `ActionEvent 0{{idRef "42"}}`, Kind: "ActionEvent", Target: 42, SentAt: at(1000), RespAt: at(1200), Result: fep.ReplaySent, RuntimeBytes: 300},
		{N: 2, Total: 3, Event: `ActionEvent 0{{idRef "77"}}`, Kind: "ActionEvent", Target: 77, SentAt: at(2000), Result: fep.ReplayNoResponse, Suspect: true},
		{N: 3, Total: 3, Event: `ConfigureEvent 0{{idRef "9"}}{{value "x"}}`, Kind: "ConfigureEvent", Result: fep.ReplaySkipped, Reason: "对象 #9 不在当前界面（AUI 安全闸）"},
	}, fep.ReplayResult{Sent: 2, Skipped: 1})

	code, out := getJSON(t, s, "/api/recordings/fep-rep/replay-steps")
	if code != http.StatusOK {
		t.Fatalf("读进度该成功：%d %v", code, out)
	}
	if out["hasLog"] != true || out["total"] != float64(3) || out["count"] != float64(4) {
		t.Fatalf("总数与行数要分开给（4 行 = 3 步，第 1 步两行）：%v", out)
	}
	// progress 是收敛过的（同一 n 取最后一行）：第 1 步算 sent，不算 waiting
	pr, _ := out["progress"].(map[string]any)
	for k, want := range map[string]float64{"total": 3, "done": 3, "sent": 1, "noResp": 1, "suspect": 1, "skipped": 1, "failed": 0} {
		if got, _ := pr[k].(float64); got != want {
			t.Errorf("progress.%s = %v，该是 %v（%v）", k, pr[k], want, pr)
		}
	}
	if pr["running"] != false {
		t.Errorf("有尾行就不是「还在跑」：%v", pr)
	}
	if st, _ := out["states"].([]any); len(st) != 3 {
		t.Fatalf("states 该是 3 步（不是 4 行）：%d", len(st))
	}
	// 分页：since 按行位置切，不丢步
	_, page2 := getJSON(t, s, "/api/recordings/fep-rep/replay-steps?since=3")
	steps2, _ := page2["steps"].([]any)
	if len(steps2) != 1 {
		t.Fatalf("since=3 只该剩最后一行：%d", len(steps2))
	}
	last, _ := steps2[0].(map[string]any)
	if int(last["n"].(float64)) != 3 || last["result"] != "skipped" {
		t.Fatalf("最后一行该是第 3 步被跳过：%v", last)
	}
	if last["sentAt"] != nil {
		t.Fatalf("跳过的步 sentAt 必须是 null：%v", last["sentAt"])
	}
	// respAt 也照实：没有回包就是 null（不是 0）
	_, all := getJSON(t, s, "/api/recordings/fep-rep/replay-steps?since=2")
	steps, _ := all["steps"].([]any)
	second, _ := steps[0].(map[string]any)
	if second["respAt"] != nil {
		t.Fatalf("没回包的步 respAt 该是 null：%v", second["respAt"])
	}
	if second["suspect"] != true {
		t.Fatalf("点击类事件没回包要标可疑（界面据此把「本来无应答」与「8s 无回包」分开）：%v", second)
	}
}

func TestReplayCtlRefusals(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	dataDir := filepath.Dir(cfgPath)

	if code, out := postJSON(t, s, "/api/recordings/fep-none/replay-ctl", `{"cmd":"pause"}`); code != http.StatusNotFound {
		t.Fatalf("没有回放日志该 404：%d %v", code, out)
	}
	path := seedReplayLog(t, dataDir, "fep-rep", []fep.ReplayStep{{N: 1, Total: 2, Result: fep.ReplayWaiting}}, fep.ReplayResult{})
	code, out := postJSON(t, s, "/api/recordings/fep-rep/replay-ctl", `{"cmd":"爆炸"}`)
	if code != http.StatusBadRequest {
		t.Fatalf("非法命令该 400：%d %v", code, out)
	}
	if msg, _ := out["error"].(string); !strings.Contains(msg, "pause") || !strings.Contains(msg, "stop") {
		t.Fatalf("拒绝理由要把允许的词列出来：%v", out)
	}
	// 不在跑（登记里没有这一轮）→ 409
	if code, out := postJSON(t, s, "/api/recordings/fep-rep/replay-ctl", `{"cmd":"pause"}`); code != http.StatusConflict {
		t.Fatalf("不在跑的时候暂停该 409：%d %v", code, out)
	}
	// 跑起来之后：暂停要说清"什么时候生效"（否则用户会以为按钮没反应）
	seedLiveReplay(t, dataDir, "fep-rep")
	code, out = postJSON(t, s, "/api/recordings/fep-rep/replay-ctl", `{"cmd":"pause"}`)
	if code != http.StatusOK {
		t.Fatalf("在跑的时候暂停该成功：%d %v", code, out)
	}
	if note, _ := out["note"].(string); !strings.Contains(note, "8 秒") {
		t.Fatalf("暂停要说清它在下一步之前生效：%v", out)
	}
	if got := fep.ReadReplayCtl(path); got != fep.CtlPause {
		t.Fatalf("暂停该落到 .ctl 文件里，读到 %q", got)
	}
	code, out = postJSON(t, s, "/api/recordings/fep-rep/replay-ctl", `{"cmd":"stop"}`)
	if code != http.StatusOK {
		t.Fatalf("停止该成功：%d %v", code, out)
	}
	if got := fep.ReadReplayCtl(path); got != fep.CtlStop {
		t.Fatalf("停止该覆盖掉上一条命令，读到 %q", got)
	}
	if note, _ := out["note"].(string); !strings.Contains(note, "收不回来") {
		t.Fatalf("停止要提醒「发出去的收不回来」：%v", out)
	}
	// 逐步调试的「跳过」（用户 2026-10-18）：停着的时候跳过当前这一步。
	code, out = postJSON(t, s, "/api/recordings/fep-rep/replay-ctl", `{"cmd":"skip"}`)
	if code != http.StatusOK {
		t.Fatalf("跳过该被接受：%d %v", code, out)
	}
	if got := fep.ReadReplayCtl(path); got != fep.CtlSkip {
		t.Fatalf("跳过该落到 .ctl 文件里，读到 %q", got)
	}
}

// TestReplayBreakTogglesAndPersists：左侧列表那个红点（**发之前停**）。
//
// 落在 `<回放日志>.breaks` 上：回放进程每一步现读，所以点了立刻生效、不用重启。
func TestReplayBreakTogglesAndPersists(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	path := seedReplayLog(t, filepath.Dir(cfgPath), "fep-bp", []fep.ReplayStep{{N: 1, Total: 3, Result: fep.ReplaySent}}, fep.ReplayResult{})
	if code, out := postJSON(t, s, "/api/recordings/fep-bp/replay-breaks", `{"n":2,"on":true}`); code != http.StatusOK {
		t.Fatalf("设断点该成功：%d %v", code, out)
	}
	if got := fep.ReadReplayBreaks(path); len(got) != 1 || got[0] != 2 {
		t.Fatalf("断点该落到 .breaks 里，读到 %v", got)
	}
	// 再加一个 + 取消第一个：集合语义（不是单值覆盖）
	if code, out := postJSON(t, s, "/api/recordings/fep-bp/replay-breaks", `{"n":5,"on":true}`); code != http.StatusOK {
		t.Fatalf("第二个断点该成功：%d %v", code, out)
	}
	if code, out := postJSON(t, s, "/api/recordings/fep-bp/replay-breaks", `{"n":2,"on":false}`); code != http.StatusOK {
		t.Fatalf("取消断点该成功：%d %v", code, out)
	}
	if got := fep.ReadReplayBreaks(path); len(got) != 1 || got[0] != 5 {
		t.Fatalf("取消之后只该剩第 5 步，读到 %v", got)
	}
	// 步号必须 ≥1
	if code, _ := postJSON(t, s, "/api/recordings/fep-bp/replay-breaks", `{"n":0,"on":true}`); code != http.StatusBadRequest {
		t.Fatalf("步号 0 该 400，实得 %d", code)
	}
}

// TestReplayPlanPrefillsTheLeftPane：左栏要能**在回放之前**铺满整份计划。
func TestReplayPlanPrefillsTheLeftPane(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	dataDir := filepath.Dir(cfgPath)
	seedLibRecording(t, dataDir, "主机109", "99", "fep-plan", "bsft001_wf")
	// 事件侧：seedLibRecording 只造界面侧，这里**在入库后的目录里**补一条事件侧
	ev := filepath.Join(dataDir, "records", "主机109", "99", "fep-plan", "fep-plan-1-gdc2rt.log")
	if err := os.WriteFile(ev, []byte("event _om 1{}{{ActionEvent 0{{idRef \"42\"}}}}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out := getJSON(t, s, "/api/recordings/fep-plan/replay-plan")
	if code != http.StatusOK {
		t.Fatalf("计划该读出来：%d %v", code, out)
	}
	steps, _ := out["steps"].([]any)
	if len(steps) != 1 {
		t.Fatalf("该有 1 步计划，实得 %d：%v", len(steps), out)
	}
	first, _ := steps[0].(map[string]any)
	if first["event"] == "" || first["kind"] != "ActionEvent" {
		t.Fatalf("这一步该带事件原文与类型：%v", first)
	}
	if first["degraded"] != true {
		t.Fatalf("没有 colName 的步该标降级：%v", first)
	}
}

// 盯盘循环：新行会被推出去（注入的 RecProgress 收得到）。
func TestReplayWatchPushesNewSteps(t *testing.T) {
	var mu sync.Mutex
	type ev struct {
		id     string
		n      int
		result string
	}
	var got []ev
	snapshot := func() []ev {
		mu.Lock()
		defer mu.Unlock()
		return append([]ev(nil), got...)
	}
	reset := func() {
		mu.Lock()
		defer mu.Unlock()
		got = nil
	}
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	// 注入一个收集器（生产路径上是调试子系统的 WS 广播）
	s.opt.RecProgress = func(id string, n, total int, result string) {
		mu.Lock()
		got = append(got, ev{id, n, result})
		mu.Unlock()
	}
	cache := filepath.Join(filepath.Dir(cfgPath), "fep")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cache, fep.ReplayLogName("fep-w", time.Now()))
	w, err := fep.NewReplayLogWriter(path, "fep-w", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	w.Step(fep.ReplayStep{N: 1, Total: 2, Result: fep.ReplayWaiting})
	w.Step(fep.ReplayStep{N: 1, Total: 2, Result: fep.ReplaySent})

	// 带上 watch=1 才会起循环（没人看不推）
	if code, _ := getJSON(t, s, "/api/recordings/fep-w/replay-steps?watch=1"); code != http.StatusOK {
		t.Fatalf("watch=1 该成功：%d", code)
	}
	// 循环是自己退场的（日志收尾 + 登记没了 + 空闲），但那是 30 秒后的事 ——
	// 测试要确定性收尾，不然 Windows 上 TempDir 会因为文件句柄还在读而删不掉（实测踩到）。
	defer s.stopWatches()
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if len(snapshot()) >= 2 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if g := snapshot(); len(g) != 2 || g[0].result != fep.ReplayWaiting || g[1].result != fep.ReplaySent {
		t.Fatalf("两条新行该各推一次（中间态也要推，界面靠它画「已等 Nms」）：%+v", g)
	}
	if g := snapshot(); g[0].id != "fep-w" {
		t.Fatalf("推出去的事件要带录制 ID（可能同时有好几轮）：%+v", g)
	}

	// 第二轮（同一份录制再回放一次）：日志是**新文件**，行数可能比上一轮少。
	// 游标必须跟着归零 —— 不归零就一行都不推，界面看着像"回放没在动"（实测照出来过）。
	reset()
	path2, err := fep.NewReplayLogWriter(filepath.Join(cache, fep.ReplayLogName("fep-w", time.Now().Add(time.Second))), "fep-w", 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	path2.Step(fep.ReplayStep{N: 1, Total: 3, Result: fep.ReplaySent})
	deadline = time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if len(snapshot()) >= 1 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if g := snapshot(); len(g) != 1 || g[0].id != "fep-w" {
		t.Fatalf("换了一份日志之后游标要归零（否则第二轮一行都不推）：%+v", g)
	}
	// 两个句柄都要放开（不写尾行的那种也放开）：Windows 上开着句柄的临时目录删不掉。
	_ = path2.Close(fep.ReplayResult{Sent: 1}, "第二轮")
	_ = w.Abandon()
}
