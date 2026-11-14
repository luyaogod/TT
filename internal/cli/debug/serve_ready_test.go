package debug

// 这组测试盯的是桌面版（TT Desktop）赖以启动的两条契约：
//
//  1. --ready-json 的那一行 JSON —— 壳按它开窗，键名/行数是跨进程契约；
//  2. --exit-on-stdin-close —— 壳被强杀时服务跟着收尾，不留隐形进程。
//
// 两条都不报错、只表现为"桌面版打不开/关了还在跑"，所以必须有测试钉住。

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// readyLine 的输出形状是契约：键只增不改（改了壳会读到空地址），且恰好一行。
func TestWriteServeReadyShape(t *testing.T) {
	const cfgPath = `C:\Users\u\AppData\Roaming\tt\config.json`
	var buf bytes.Buffer
	st := &ServeInfo{PID: 4242, URL: "http://127.0.0.1:28671", APIBase: APIBaseUnderDebug}
	if err := WriteServeReady(&buf, st, cfgPath, false); err != nil {
		t.Fatalf("WriteServeReady: %v", err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("契约要求恰好一行，得到 %d 行: %q", len(lines), buf.String())
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("不是合法 JSON: %v\n%s", err, lines[0])
	}
	wantKeys := map[string]bool{"ok": true, "attached": true, "url": true, "pid": true, "config": true}
	for k := range got {
		if !wantKeys[k] {
			t.Errorf("多了一个键 %q（壳不认识它；要加键先确认壳那边已兼容）", k)
		}
	}
	for k := range wantKeys {
		if _, has := got[k]; !has {
			t.Errorf("缺键 %q", k)
		}
	}
	if got["ok"] != true {
		t.Errorf("ok = %v，想要 true", got["ok"])
	}
	if got["url"] != st.URL {
		t.Errorf("url = %v，想要 %s", got["url"], st.URL)
	}
	if got["config"] != cfgPath {
		t.Errorf("config = %v，想要 %s", got["config"], cfgPath)
	}
	if pid, _ := got["pid"].(float64); int(pid) != st.PID {
		t.Errorf("pid = %v，想要 %d", got["pid"], st.PID)
	}
	if attached, _ := got["attached"].(bool); attached {
		t.Error("attached = true，但这次是自持（false）")
	}
}

// 附着与自持必须能区分：壳靠它决定退出时要不要收掉服务。
func TestWriteServeReadyAttached(t *testing.T) {
	var buf bytes.Buffer
	st := &ServeInfo{PID: 7, URL: "http://127.0.0.1:28670"}
	if err := WriteServeReady(&buf, st, "c.json", true); err != nil {
		t.Fatalf("WriteServeReady: %v", err)
	}
	var got struct {
		Attached bool `json:"attached"`
		PID      int  `json:"pid"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("不是合法 JSON: %v", err)
	}
	if !got.Attached || got.PID != 7 {
		t.Errorf("attached/pid 没传出去: %+v", got)
	}
}

// RunningServe 是"附着"的唯一判据（状态文件 + HTTP 探活）。
// 这里钉住两件事：真在跑能认出来，进程没了就认不出来 —— 后者错了会让桌面版
// 附着到一个死地址上，开窗即白屏。
func TestRunningServeAttachProbe(t *testing.T) {
	dir := t.TempDir()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"server":"tt-unified"}`))
	}))

	st := &ServeInfo{PID: 4321, URL: srv.URL, APIBase: APIBaseUnderDebug}
	if err := writeServeInfo(st, dir); err != nil {
		t.Fatalf("writeServeInfo: %v", err)
	}
	if got := RunningServe(dir); got == nil || got.URL != srv.URL {
		t.Fatalf("在跑的实例没认出来: %+v", got)
	}

	srv.Close()
	if got := RunningServe(dir); got != nil {
		t.Errorf("服务已经关了，RunningServe 仍返回 %+v（桌面版会附着到死地址）", got)
	}
}

// stdin 关闭 = 优雅收尾。桌面壳被强杀时靠这条把子进程带走。
func TestRunForeground_ExitOnStdinClose(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	var once sync.Once
	ready := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- RunForeground(cfgPath, ServeSpec{
			Prefix:           "[test]",
			StopHint:         "test --stop",
			ExitOnStdinClose: true,
		}, func(ctx context.Context, onReady func(addr string)) error {
			onReady("127.0.0.1:1")
			once.Do(func() { close(ready) })
			<-ctx.Done()
			return nil
		})
	}()

	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		t.Fatal("服务没在 5 秒内就绪")
	}
	// 管道关闭 —— 等价于桌面壳进程消失
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunForeground 返回 %v，想要 nil（EOF 应当按正常收尾处理）", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("stdin 关了 5 秒还没退出：桌面壳被强杀后会留下隐形的 tt.exe")
	}

	// 收尾要把状态文件带走，否则下次启动会以为还有一个实例在跑
	if _, err := os.Stat(filepath.Join(dir, ".tt-serve.json")); !os.IsNotExist(err) {
		t.Errorf("收尾后状态文件还在: %v", err)
	}
}
