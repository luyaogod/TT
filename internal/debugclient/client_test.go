package debugclient

// 这组测试盯的是"客户端把服务端的话**原样**传下去"这件事：
// 失败码、候选、hint、退出码一个都不能在客户端这一层被改写或吞掉 ——
// 从前的子进程实现只能靠退出码反推，正是这里要消灭的东西。

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tt/internal/target"
)

// TestFailureKeepsCodeCandidatesAndHint：失败信封的三个关键部分都要留住。
func TestFailureKeepsCodeCandidatesAndHint(t *testing.T) {
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"ok":false,"code":"SESSION_AMBIGUOUS",
			"error":"有 2 条 SSH 会话","hint":"照抄 ref 重发",
			"candidates":[{"id":"s1","ref":"主机109#1","name":"A","n":1},
			              {"id":"s2","ref":"主机109#2","name":"B","n":2}]}`))
	}))
	defer hs.Close()

	_, err := New(hs.URL).Do(http.MethodGet, "/sessions", nil)
	var f *Failure
	if !errors.As(err, &f) {
		t.Fatalf("该是 *Failure，实得 %T: %v", err, err)
	}
	if f.Code != target.CodeAmbiguous {
		t.Errorf("code 被改了：%q", f.Code)
	}
	if len(f.Candidates) != 2 || f.Candidates[0].Ref != "主机109#1" {
		t.Errorf("候选没留住：%+v", f.Candidates)
	}
	if f.Hint == "" {
		t.Error("hint 被丢了（它才是'下一步怎么做'）")
	}
	// 退出码要按码表算（歧义 = 3），不是 HTTP 409。
	if got := f.ExitCode(); got != 3 {
		t.Errorf("ExitCode = %d，要 3", got)
	}
}

// TestNonEnvelopeErrorDoesNotBecomeFailure：不是信封的失败不许伪装成带码失败。
func TestNonEnvelopeErrorDoesNotBecomeFailure(t *testing.T) {
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>bad gateway</html>"))
	}))
	defer hs.Close()

	_, err := New(hs.URL).Do(http.MethodGet, "/sessions", nil)
	var f *Failure
	if errors.As(err, &f) {
		t.Fatalf("没有信封时不该造一个 *Failure：%+v", f)
	}
	if err == nil {
		t.Fatal("非 2xx 必须报错")
	}
}

// TestResolveIsStrict：别名 / ref / 句柄都能解析，但**多条命中不自己挑**。
func TestResolveIsStrict(t *testing.T) {
	body := `{"ok":true,"sessions":[
		{"id":"s1","ref":"主机109#1","name":"A","env":"主机109","ent":"99","state":"idle","n":1},
		{"id":"s2","ref":"主机109#2","name":"B","env":"主机109","ent":"99","state":"idle","n":2}]}`
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer hs.Close()
	c := New(hs.URL)

	if b, err := c.Resolve("A"); err != nil || b.ID != "s1" {
		t.Fatalf("按别名解析失败：%v %+v", err, b)
	}
	if b, err := c.Resolve("主机109#2"); err != nil || b.ID != "s2" {
		t.Fatalf("按 ref 解析失败：%v %+v", err, b)
	}
	if b, err := c.Resolve("s1"); err != nil || b.ID != "s1" {
		t.Fatalf("按句柄解析失败：%v %+v", err, b)
	}
	if _, err := c.Resolve("没有这条"); err == nil {
		t.Fatal("找不到时必须报错")
	} else {
		var f *Failure
		if !errors.As(err, &f) || f.Code != target.CodeNotFound {
			t.Fatalf("该报 %q，实得 %v", target.CodeNotFound, err)
		}
	}
}

// TestWaitStoppedPollsUntilStopped：轮询到 stopped 就返回。
func TestWaitStoppedPollsUntilStopped(t *testing.T) {
	var calls int
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		state := "loading"
		if calls >= 2 {
			state = "stopped"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "state": state})
	}))
	defer hs.Close()

	snap, ok, err := New(hs.URL).WaitStopped("s1", 5*time.Second)
	if err != nil || !ok {
		t.Fatalf("应等到停站：ok=%v err=%v", ok, err)
	}
	if stateOf(snap) != "stopped" {
		t.Fatalf("快照状态 = %v", snap["state"])
	}
	if calls < 2 {
		t.Fatalf("应轮询至少两次，实得 %d", calls)
	}
}

// TestWaitStoppedTimeoutIsNotAnError：超时不是错误（调用方照常往下走并如实报超时）。
func TestWaitStoppedTimeoutIsNotAnError(t *testing.T) {
	hs := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"state":"running"}`))
	}))
	defer hs.Close()

	_, ok, err := New(hs.URL).WaitStopped("s1", 0)
	if err != nil {
		t.Fatalf("超时不该是错误：%v", err)
	}
	if ok {
		t.Fatal("还在运行不该报停站")
	}
}

// TestDiscoverUsesStateFile：状态文件里的地址 + 前缀优先。
func TestDiscoverUsesStateFile(t *testing.T) {
	dir := t.TempDir()
	st := `{"pid":1,"url":"http://127.0.0.1:29000","apiBase":"/debug/api"}`
	if err := os.WriteFile(filepath.Join(dir, ".tt-serve.json"), []byte(st), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := Discover(dir, ""), "http://127.0.0.1:29000/debug/api"; got != want {
		t.Fatalf("Discover = %q，要 %q", got, want)
	}
}

// TestDiscoverFallsBackToListen：没有状态文件 → 配置 listen；也没有 → 内置默认。
func TestDiscoverFallsBackToListen(t *testing.T) {
	dir := t.TempDir()
	if got, want := Discover(dir, "127.0.0.1:1234"), "http://127.0.0.1:1234"; got != want {
		t.Fatalf("Discover = %q，要 %q", got, want)
	}
	if got := Discover(dir, ""); got == "" || got == "http://" {
		t.Fatalf("没有状态文件也没有 listen 时必须给内置默认，实得 %q", got)
	}
}
