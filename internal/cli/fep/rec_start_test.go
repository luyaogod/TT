package fep

// 录制启动那套动作的判据。
//
// # 为什么从"比对 argv"改成"比对调用序列"
//
// 从前这套动作是靠**起子进程** `tt debug …` 做的，于是判据只能比字符串：
// `debug env 主机109 --name X` 里那几个词对不对。那种判据有两个毛病：
//   - 它把"我们自己的命令行语法"钉死了 —— 改一个 flag 名，测试就红，哪怕语义没变；
//   - 它测不到真正要命的事：**每个动作打到哪个端点、带什么参数、顺序对不对**。
//
// 现在动作直连调试面（`internal/debugclient`），判据就可以直接断言"调了哪些端点、什么顺序"。
// 顺序仍然是这套东西的灵魂：少一步或顺序反了，症状全是**安静**的（界面不出来、作业跑在
// 别的企业上、断点没下上），真机才发现代价很大。

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"tt/internal/debugclient"
)

// recFake 是一个最小的调试面假服务：记录每次调用，并按端点回固定应答。
type recFake struct {
	mu    sync.Mutex
	calls []string // "METHOD /path"

	// 预置：会话列表（供 Resolve 收口用）
	sessions []map[string]any
}

func (f *recFake) record(method, path string) {
	f.mu.Lock()
	f.calls = append(f.calls, method+" "+path)
	f.mu.Unlock()
}

func (f *recFake) snapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	copy(out, f.calls)
	return out
}

func (f *recFake) server(t *testing.T) *debugclient.Client {
	t.Helper()
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.record(r.Method, r.URL.Path)
		switch {
		case r.URL.Path == "/api/sessions" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "sessions": f.sessions})
		case r.URL.Path == "/api/sessions" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"ok":true,"sessionId":"s-job"}`))
		case r.URL.Path == "/api/sessions/switch":
			_, _ = w.Write([]byte(`{"ok":true,"sessionId":"s-host","state":"idle"}`))
		case strings.HasPrefix(r.URL.Path, "/api/sessions/s-host") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"ok":true,"state":"stopped","stop":{"file":"a.4gl","line":812}}`))
		case strings.HasPrefix(r.URL.Path, "/api/sessions/s-job") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"ok":true,"state":"stopped"}`))
		case strings.HasSuffix(r.URL.Path, "/raw"):
			_, _ = w.Write([]byte(`{"ok":true,"lines":[]}`))
		default:
			_, _ = w.Write([]byte(`{"ok":true}`))
		}
	}))
	t.Cleanup(hs.Close)
	return debugclient.New(hs.URL)
}

func quietLog(string, ...any) {}

// TestRecRunPlanOrder：六个动作必须按这个顺序打到这些端点。
func TestRecRunPlanOrder(t *testing.T) {
	f := &recFake{}
	cl := f.server(t)

	_, err := runRecPlan(cl, recRunPlan{
		Alias: "fep-R1", Env: "主机109", Ent: "99",
		Prog: "bsft001_wf", Module: "asf", FGL: "10.0.0.1:1200",
		Breaks: []string{"812"},
	}, quietLog)
	if err != nil {
		t.Fatalf("runRecPlan: %v", err)
	}

	got := f.snapshot()
	want := []string{
		"GET /api/sessions",                // ① 收口前先找自己那条（没有就跳过）
		"POST /api/sessions/switch",        // ② 点名环境并命名
		"POST /api/sessions/s-host/topent", // ③ 定 ENT
		"POST /api/sessions",               // ④ 起作业
		"GET /api/sessions/s-job",          // ⑤ 等入口停站
		"POST /api/sessions/s-job/raw",     // ⑥ 下断点
		"GET /api/sessions/s-job",          // ⑦ 回填 ref（P3：两条线认同一个地址）
		"POST /api/sessions/s-job/control", // ⑧ 放行
	}
	if len(got) != len(want) {
		t.Fatalf("调了 %d 次，要 %d 次：\n实得 %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 次是 %q，要 %q（顺序是这套动作的灵魂）\n全部：%v", i, got[i], want[i], got)
		}
	}
}

// TestRecRunPlanQuitsItsOwnSessionFirst：上一轮自己那条还在 → 先收掉它再往下走。
func TestRecRunPlanQuitsItsOwnSessionFirst(t *testing.T) {
	f := &recFake{sessions: []map[string]any{{"id": "s-old", "ref": "主机109#1", "name": "fep-R1", "state": "stopped"}}}
	cl := f.server(t)

	if _, err := runRecPlan(cl, recRunPlan{Alias: "fep-R1", Env: "主机109", Ent: "99", Prog: "p", FGL: "1:2"}, quietLog); err != nil {
		t.Fatalf("runRecPlan: %v", err)
	}
	got := f.snapshot()
	if len(got) < 2 || got[0] != "GET /api/sessions" || got[1] != "DELETE /api/sessions/s-old" {
		t.Fatalf("应先把上一轮那条收掉，实得 %v", got)
	}
}

// TestRecRunPlanSkipsTopentWithoutEnt：没给 ENT 就不该打 topent。
//
// （ENT 现在在起作业那一步是必填；这里测的是"计划里没 ENT 时不多打一次空请求"。）
func TestRecRunPlanSkipsTopentWithoutEnt(t *testing.T) {
	f := &recFake{}
	cl := f.server(t)

	if _, err := runRecPlan(cl, recRunPlan{Alias: "fep-R1", Env: "主机109", Prog: "p", FGL: "1:2"}, quietLog); err != nil {
		t.Fatalf("runRecPlan: %v", err)
	}
	for _, c := range f.snapshot() {
		if strings.Contains(c, "/topent") {
			t.Fatalf("没 ENT 时不该打 topent：%v", f.snapshot())
		}
	}
}

// TestRecRunPlanNoContinueSkipsRelease：--no-continue 时不放行（断点下完就停在那儿）。
func TestRecRunPlanNoContinueSkipsRelease(t *testing.T) {
	f := &recFake{}
	cl := f.server(t)

	if _, err := runRecPlan(cl, recRunPlan{
		Alias: "fep-R1", Env: "主机109", Ent: "99", Prog: "p", FGL: "1:2",
		Breaks: []string{"812"}, NoContinue: true,
	}, quietLog); err != nil {
		t.Fatalf("runRecPlan: %v", err)
	}
	for _, c := range f.snapshot() {
		if strings.Contains(c, "/control") {
			t.Fatalf("--no-continue 时不该放行：%v", f.snapshot())
		}
	}
}

// TestRecSessionAliasIsStable：别名由录制身份算出，同一个 ID 永远同一个名字。
func TestRecSessionAliasIsStable(t *testing.T) {
	if recSessionAlias("fep-20261016-120000") != recSessionAlias("fep-20261016-120000") {
		t.Fatal("同一个录制 ID 必须算出同一个别名（别名是幂等键，Q23）")
	}
	if recSessionAlias("") != "" {
		t.Fatal("没有录制 ID 时不该编一个别名出来")
	}
	if a := recSessionAlias("fep-20261016-120000"); strings.ContainsAny(a, `/#`) {
		t.Fatalf("别名不许含 / 与 #（契约 C2.5 / Q17）：%q", a)
	}
}
