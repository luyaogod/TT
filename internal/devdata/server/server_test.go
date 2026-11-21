package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/devdata"
)

// fixture 一个只带环境的配置 + 可观测的假依赖。
type fixture struct {
	cfgPath string
	srv     *Server
	calls   map[string]int
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	body := `{"hosts":{"sshs":[{"name":"福联测试tst","host":"10.0.0.1","user":"u","zone":"2"}]},"schemaVersion":4}`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	f := &fixture{cfgPath: cfgPath, calls: map[string]int{}}
	deps := Deps{
		Status: func(_ context.Context, req devdata.StatusRequest) (*devdata.EnvStatus, error) {
			f.calls["status"]++
			return &devdata.EnvStatus{Env: req.Env, Progs: []devdata.ProgStatus{}, Jobs: req.Jobs,
				Dict: &devdata.DictStatus{}, Mdl: &devdata.MdlStatus{}}, nil
		},
		Mirror: func(ctx context.Context, req devdata.MirrorRequest, _ devdata.ProgressFunc) (*devdata.MirrorResult, error) {
			f.calls["mirror"]++
			<-ctx.Done() // 一直跑到被取消（用来验 409 / stop）
			return nil, ctx.Err()
		},
		Fetch: func(ctx context.Context, req devdata.FetchRequest, _ devdata.ProgressFunc) (*devdata.FetchResult, error) {
			f.calls["fetch"]++
			<-ctx.Done()
			return nil, ctx.Err()
		},
		Check: func(_ context.Context, req devdata.StatusRequest) (*devdata.ProgStatus, error) {
			f.calls["check"]++
			return &devdata.ProgStatus{Prog: req.Prog, Exists: true, Stale: true}, nil
		},
		PlanClean: func(_, env, prog string) (*devdata.CleanPlan, error) {
			f.calls["planClean"]++
			if prog != "axmt500_wf" {
				return nil, devdata.ErrProgNotFound
			}
			return &devdata.CleanPlan{Items: []devdata.CleanItem{{Kind: "prog", Env: env, Prog: prog}}}, nil
		},
		ExecuteClean: func(*devdata.CleanPlan) error { f.calls["executeClean"]++; return nil },
	}
	f.srv = New(Options{ConfigPath: cfgPath, Version: "test", Deps: deps})
	return f
}

// do 发一个请求，返回状态码与解析后的 JSON。
func (f *fixture) do(t *testing.T, method, target, body string) (int, map[string]any) {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}
	w := httptest.NewRecorder()
	f.srv.Handler().ServeHTTP(w, r)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

// TestEnvValidation env 缺失 → 400；不在清单 → 404；两者都不该调领域层。
func TestEnvValidation(t *testing.T) {
	f := newFixture(t)
	for _, tc := range []struct {
		target string
		code   int
		ecode  string
	}{
		{"/api/devdata/status", 400, CodeUsage},
		{"/api/devdata/status?env=", 400, CodeUsage},
		{"/api/devdata/status?env=没有这个环境", 404, CodeNotFound},
		{"/api/devdata/check?env=没有这个环境&prog=axmt500_wf", 404, CodeNotFound},
	} {
		code, out := f.do(t, http.MethodGet, tc.target, "")
		if code != tc.code {
			t.Errorf("%s → %d，应为 %d（%v）", tc.target, code, tc.code, out)
		}
		if out["code"] != tc.ecode {
			t.Errorf("%s 的 code = %v，应为 %s", tc.target, out["code"], tc.ecode)
		}
		if out["ok"] != false {
			t.Errorf("%s 应返回 ok=false: %v", tc.target, out)
		}
	}
	if f.calls["status"] != 0 || f.calls["check"] != 0 {
		t.Errorf("校验失败时不该调领域层: %v", f.calls)
	}
	// 也不该建出 devdata 目录（校验先于落盘）
	if _, err := os.Stat(filepath.Join(filepath.Dir(f.cfgPath), "devdata")); err == nil {
		t.Error("校验失败时不该创建 devdata 目录")
	}
}

// TestProgValidation 程序名非法 → 400，且不调领域层。
func TestProgValidation(t *testing.T) {
	f := newFixture(t)
	for _, prog := range []string{"../x", "Axmt500", "a", "axmt500-wf"} {
		code, out := f.do(t, http.MethodPost, "/api/devdata/fetch?env=福联测试tst",
			`{"env":"福联测试tst","prog":`+quote(prog)+`}`)
		if code != http.StatusBadRequest {
			t.Errorf("prog %q → %d，应为 400（%v）", prog, code, out)
		}
		if out["code"] != CodeUsage {
			t.Errorf("prog %q 的 code = %v", prog, out["code"])
		}
	}
	if f.calls["fetch"] != 0 {
		t.Errorf("非法程序名不该调领域层: %v", f.calls)
	}
}

// TestFetchStartBusyAndStop 启动 → 再启动 409 → stop → 结束。
func TestFetchStartBusyAndStop(t *testing.T) {
	f := newFixture(t)
	code, out := f.do(t, http.MethodPost, "/api/devdata/fetch", `{"env":"福联测试tst","prog":"axmt500_wf"}`)
	if code != http.StatusOK || out["ok"] != true {
		t.Fatalf("首次启动 = %d %v", code, out)
	}
	// 等任务真的跑起来（Manager 的 Start 是同步建状态，这里不用等）
	code, out = f.do(t, http.MethodPost, "/api/devdata/fetch", `{"env":"福联测试tst","prog":"axmt500_wf"}`)
	if code != http.StatusConflict || out["code"] != CodeConflict {
		t.Fatalf("重复启动应 409: %d %v", code, out)
	}
	// status 里应能看到在跑的任务
	code, out = f.do(t, http.MethodGet, "/api/devdata/status?env=福联测试tst", "")
	if code != http.StatusOK {
		t.Fatalf("status = %d %v", code, out)
	}
	jobs, _ := out["jobs"].([]any)
	if len(jobs) != 1 {
		t.Fatalf("status 应带 1 个在跑任务: %v", out["jobs"])
	}
	// 停止
	code, out = f.do(t, http.MethodPost, "/api/devdata/stop", `{"env":"福联测试tst","kind":"fetch","prog":"axmt500_wf"}`)
	if code != http.StatusOK || out["ok"] != true {
		t.Fatalf("stop = %d %v", code, out)
	}
	// 再停 → 409（没有任务可停）
	deadline := time.Now().Add(2 * time.Second)
	for {
		code, out = f.do(t, http.MethodPost, "/api/devdata/stop", `{"env":"福联测试tst","kind":"fetch","prog":"axmt500_wf"}`)
		if code == http.StatusConflict {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("任务没在取消后收尾: %d %v", code, out)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestDeleteRequiresYes 删除必须 ?yes=1；没拉过的程序 → 404。
func TestDeleteRequiresYes(t *testing.T) {
	f := newFixture(t)
	code, out := f.do(t, http.MethodDelete, "/api/devdata/prog?env=福联测试tst&prog=axmt500_wf", "")
	if code != http.StatusBadRequest || out["code"] != CodeUsage {
		t.Fatalf("无 yes 应 400: %d %v", code, out)
	}
	if f.calls["executeClean"] != 0 {
		t.Error("没有 yes 时不该真的删")
	}
	code, out = f.do(t, http.MethodDelete, "/api/devdata/prog?env=福联测试tst&prog=notthere&yes=1", "")
	if code != http.StatusNotFound || out["code"] != CodeNotFound {
		t.Fatalf("没拉过的程序应 404: %d %v", code, out)
	}
	code, out = f.do(t, http.MethodDelete, "/api/devdata/prog?env=福联测试tst&prog=axmt500_wf&yes=1", "")
	if code != http.StatusOK || out["ok"] != true {
		t.Fatalf("删除应成功: %d %v", code, out)
	}
	if f.calls["executeClean"] != 1 {
		t.Errorf("应执行一次删除，实际 %v", f.calls)
	}
}

// TestMirrorOnlyValidation only 只认 dict/mdl。
func TestMirrorOnlyValidation(t *testing.T) {
	f := newFixture(t)
	code, out := f.do(t, http.MethodPost, "/api/devdata/mirror", `{"env":"福联测试tst","only":"all"}`)
	if code != http.StatusBadRequest || out["code"] != CodeUsage {
		t.Fatalf("非法 only 应 400: %d %v", code, out)
	}
	code, out = f.do(t, http.MethodPost, "/api/devdata/mirror", `{"env":"福联测试tst","only":"dict"}`)
	if code != http.StatusOK {
		t.Fatalf("启动 mirror 应成功: %d %v", code, out)
	}
}

// TestUnknownEndpoint 未知 /api/devdata/* → 404 JSON（不能回 HTML）。
func TestUnknownEndpoint(t *testing.T) {
	f := newFixture(t)
	code, out := f.do(t, http.MethodGet, "/api/devdata/nope", "")
	if code != http.StatusNotFound || out["code"] != CodeNotFound {
		t.Fatalf("未知接口应 404 JSON: %d %v", code, out)
	}
}

// TestBodyValidation 请求体非法 → 400。
func TestBodyValidation(t *testing.T) {
	f := newFixture(t)
	code, out := f.do(t, http.MethodPost, "/api/devdata/fetch", `{not json`)
	if code != http.StatusBadRequest || out["code"] != CodeUsage {
		t.Fatalf("非法 body 应 400: %d %v", code, out)
	}
}

// TestCheckNotFetched check 走领域层，未拉取时 404。
func TestCheckNotFetched(t *testing.T) {
	f := newFixture(t)
	f.srv.deps.Check = func(context.Context, devdata.StatusRequest) (*devdata.ProgStatus, error) {
		return nil, devdata.ErrProgNotFound
	}
	code, out := f.do(t, http.MethodGet, "/api/devdata/check?env=福联测试tst&prog=axmt500_wf", "")
	if code != http.StatusNotFound || out["code"] != CodeNotFound {
		t.Fatalf("未拉取应 404: %d %v", code, out)
	}
}

// TestDomainErrorMapping 领域错误到 HTTP 码的映射（只按哨兵错误，不看文案）。
func TestDomainErrorMapping(t *testing.T) {
	cases := []struct {
		err  error
		code int
		ec   string
	}{
		{devdata.ErrEnvNotFound, 404, CodeNotFound},
		{devdata.ErrProgNotFound, 404, CodeNotFound},
		{devdata.ErrInvalidProg, 400, CodeUsage},
		{devdata.ErrInvalidEnv, 400, CodeUsage},
		{devdata.ErrInvalidOpt, 400, CodeUsage},
		{devdata.ErrModuleUnknown, 400, CodeUsage},
		{devdata.ErrBusy, 409, CodeConflict},
		{devdata.ErrNoJob, 409, CodeConflict},
		{devdata.ErrExists, 409, CodeConflict},
		{errors.New("别的问题"), 500, CodeInternal},
	}
	for _, c := range cases {
		w := httptest.NewRecorder()
		(&Server{}).failDomain(w, c.err)
		if w.Code != c.code {
			t.Errorf("%v → %d，应为 %d", c.err, w.Code, c.code)
		}
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		if out["code"] != c.ec {
			t.Errorf("%v 的 code = %v，应为 %s", c.err, out["code"], c.ec)
		}
	}
}

// quote 把字符串包成 JSON 字符串。
func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
