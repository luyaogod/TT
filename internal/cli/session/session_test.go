package session

// 这组判据盯的是"命令 → 端点"的对应关系：新命令面每个动词打哪个端点、带什么参数。
//
// 为什么不做成"跑真服务"的集成测试：会话的动作都要真 SSH + 真 fgldb，单测跑不起；
// 而真正容易写错的是**接线**（打错端点、参数名写错、rm 用了 quit 的端点），
// 用一个假服务就能钉死。语义对不对由服务端自己的判据负责。

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"tt/internal/debugclient"
	"tt/internal/testkit"
)

type recFake struct {
	mu    sync.Mutex
	calls []string
	bod   []map[string]any
}

func (f *recFake) install(t *testing.T) {
	t.Helper()
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		f.calls = append(f.calls, r.Method+" "+r.URL.Path)
		f.bod = append(f.bod, body)
		f.mu.Unlock()

		switch {
		case r.URL.Path == "/api/sessions" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"ok":true,"sessions":[
				{"id":"s1","ref":"主机109#1","name":"A","env":"主机109","ent":"99","n":1,"epoch":"3","state":"idle","prog":"bsft001_wf"}]}`))
		case r.URL.Path == "/api/sessions" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"ok":true,"sessionId":"s9"}`))
		case r.URL.Path == "/api/sessions/switch":
			_, _ = w.Write([]byte(`{"ok":true,"sessionId":"s1","state":"idle"}`))
		case strings.HasSuffix(r.URL.Path, "/screen"):
			_, _ = w.Write([]byte(`{"ok":true,"session":"主机109#1","attached":true,"screen":{"tag":"fep-R1","kind":"rec"}}`))
		case strings.HasSuffix(r.URL.Path, "/record"):
			_, _ = w.Write([]byte(`{"ok":true,"recording":{"tag":"fep-R1","kind":"rec"}}`))
		case strings.HasPrefix(r.URL.Path, "/api/sessions/s1") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"ok":true,"id":"s1","ref":"主机109#1","state":"stopped","stop":{"file":"a.4gl","line":812}}`))
		case strings.HasPrefix(r.URL.Path, "/api/sessions/s9") && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"ok":true,"id":"s9","ref":"主机109#2","state":"stopped"}`))

		case strings.HasSuffix(r.URL.Path, "/raw"):
			_, _ = w.Write([]byte(`{"ok":true,"lines":["Breakpoint 1 at a.4gl:812"]}`))
		default:
			_, _ = w.Write([]byte(`{"ok":true}`))
		}
	}))
	t.Cleanup(hs.Close)

	old := newClient
	newClient = func() (*debugclient.Client, error) { return debugclient.New(hs.URL), nil }
	t.Cleanup(func() { newClient = old })
}

func (f *recFake) snapshot() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.calls))
	copy(out, f.calls)
	return out
}

func (f *recFake) bodies() []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]map[string]any, len(f.bod))
	copy(out, f.bod)
	return out
}

// resetFlags 把命令级标志清回默认（同一个进程里反复 Execute 时必做）。
func resetFlags() {
	nsEnv, nsEnt, nsName, nsProg, nsModule = "", "", "", "", ""
	lsEnv, lsSession, entValue = "", "", ""
	lsAll, entClear = false, false
}

// runCmd 跑一次子命令，返回 stdout。
func runCmd(t *testing.T, name string, args ...string) string {
	t.Helper()
	// 用组本身当入口（cobra 支持的用法）：直接 Execute 子命令会走另一条解析路径，
	// 而且会把 -- 之后的透传参数吃掉。
	//
	// **每个用例先清标志**：cobra 的 flag 值挂在命令对象上，同一个测试进程里再执行一次
	// 会继承上一次的值（`--env` 没给却还是旧的）—— 那会让"必填项校验"这条判据假绿。
	resetFlags()
	group.SetArgs(append([]string{name}, args...))
	_, out := testkit.CaptureStdout(t, func() int {
		if err := group.Execute(); err != nil {
			t.Fatalf("%s %v 报错：%v", name, args, err)
		}
		return 0
	})
	return out
}

// TestSessionNewIdle：不給 --prog → 建**空闲**会话（走 switch），并回显身份。
func TestSessionNewIdle(t *testing.T) {
	f := &recFake{}
	f.install(t)

	out := runCmd(t, "new", "--env", "主机109", "--ent", "99", "--name", "A")
	if !strings.Contains(out, `"ref": "主机109#1"`) {
		t.Fatalf("应回显落在哪条会话（ref）：%s", out)
	}
	got := f.snapshot()
	if len(got) < 1 || got[0] != "POST /api/sessions/switch" {
		t.Fatalf("空闲会话该走 switch，实得 %v", got)
	}
	if b := f.bodies()[0]; b["ent"] != "99" || b["name"] != "A" || b["env"] != "主机109" {
		t.Fatalf("switch 请求体缺字段：%+v", b)
	}
}

// TestSessionNewWithProg：给了 --prog → 走起会话 + 等入口停站。
func TestSessionNewWithProg(t *testing.T) {
	f := &recFake{}
	f.install(t)

	runCmd(t, "new", "--env", "主机109", "--ent", "99", "--prog", "bsft001_wf", "-m", "asf")
	got := f.snapshot()
	if len(got) < 2 || got[0] != "POST /api/sessions" {
		t.Fatalf("有作业时该先 POST /api/sessions，实得 %v", got)
	}
	if b := f.bodies()[0]; b["prog"] != "bsft001_wf" || b["ent"] != "99" || b["module"] != "asf" {
		t.Fatalf("起会话请求体缺字段：%+v", b)
	}
}

// TestSessionNewRequiresEnvAndEnt：两个必填项缺一不可（ENT 决定账号与据点校验）。
func TestSessionNewRequiresEnvAndEnt(t *testing.T) {
	f := &recFake{}
	f.install(t)

	for _, args := range [][]string{
		{"--ent", "99"},    // 缺 env
		{"--env", "主机109"}, // 缺 ent
	} {
		resetFlags() // 同一进程里反复 Execute：先清掉上一次的标志值，否则"缺 env"测不出来
		group.SetArgs(append([]string{"new"}, args...))
		if err := group.Execute(); err == nil {
			t.Fatalf("%v 应当报错（必填项缺失）", args)
		}
	}
}

// TestSessionLsAndShow：列表与详情各自打对端点。
func TestSessionLsAndShow(t *testing.T) {
	f := &recFake{}
	f.install(t)

	out := runCmd(t, "ls")
	if !strings.Contains(out, `"ref": "主机109#1"`) {
		t.Fatalf("ls 应列出会话：%s", out)
	}

	out = runCmd(t, "show", "--session", "A")
	if !strings.Contains(out, `"line": 812`) {
		t.Fatalf("show 应给出停站现场：%s", out)
	}
	// 附件也一并给出来（画面通道 / 录制）：不然人得再跑两条命令才知道这条会话在录什么。
	if !strings.Contains(out, `"tag": "fep-R1"`) {
		t.Fatalf("show 应带上画面/录制附件：%s", out)
	}
}

// TestSessionEntQuitRm：三条动作分别打 topent / quit(DELETE) / close(POST)。
//
// 特别钉住 rm：它必须走 `/close` 而不是 `DELETE` —— 后者是"结束本轮、会话保留"。
// 两个动作混用的后果是"以为收掉了，其实还留着一个开着的会话"。
func TestSessionEntQuitRm(t *testing.T) {
	f := &recFake{}
	f.install(t)

	runCmd(t, "ent", "99", "--session", "A")
	runCmd(t, "quit", "--session", "A")
	runCmd(t, "rm", "--session", "A")

	got := f.snapshot()
	joined := strings.Join(got, " | ")
	if !strings.Contains(joined, "POST /api/sessions/s1/topent") {
		t.Errorf("ent 应打 topent：%s", joined)
	}
	if !strings.Contains(joined, "DELETE /api/sessions/s1") {
		t.Errorf("quit 应打 DELETE /api/sessions/{id}（结束本轮、会话保留）：%s", joined)
	}
	if !strings.Contains(joined, "POST /api/sessions/s1/close") {
		t.Errorf("rm 应打 POST /api/sessions/{id}/close（收掉会话）：%s", joined)
	}
}

// TestSessionExecPassthrough：透传命令原样送出去，输出原样回。
func TestSessionExecPassthrough(t *testing.T) {
	f := &recFake{}
	f.install(t)

	out := runCmd(t, "exec", "--session", "A", "--", "break 812")
	if !strings.Contains(out, "Breakpoint 1 at a.4gl:812") {
		t.Fatalf("透传输出丢了：%s", out)
	}
	bodies := f.bodies()
	var rawBody map[string]any
	for i, c := range f.snapshot() {
		if strings.HasSuffix(c, "/raw") {
			rawBody = bodies[i]
		}
	}
	if rawBody == nil || rawBody["command"] != "break 812" {
		t.Fatalf("透传命令没送到 /raw：%+v", rawBody)
	}
}
