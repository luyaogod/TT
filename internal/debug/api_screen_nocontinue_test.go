package debug

// 「挂监控」与「放行」是两件事（真机 2026-10-18 13:34 的现场）。
//
// debug 用途的会话一挂上监控，录制进程起完作业就自动 continue —— 入口停站被跳过，
// 断点还没下、调用栈是空的，人已经没法调试了。
//
// 判据：**用途决定放不放行** —— debug → 停在入口（NoContinue=true），record → 自动放行。
// 这一条钉在服务端：启动器拿到的 `NoContinue` 必须跟着用途走，而不是由前端猜。

import (
	"net/http"
	"testing"
)

func TestScreenStartNoContinueByPurpose(t *testing.T) {
	cases := []struct {
		purpose string
		want    bool
		why     string
	}{
		{PurposeDebug, true, "debug：作业停在入口，人下好断点再继续"},
		{PurposeRecord, false, "record：纯录制，起完就放行跑起来"},
	}
	for _, c := range cases {
		t.Run(c.purpose, func(t *testing.T) {
			hs, s := actionTestDebugServer(t)
			giveServerNamedEnv(s, "h1", "127.0.0.1", "36")
			sess := idleOn("s-1", s.cfg.CloneEnv("h1"))
			sess.setNameInTest("监控自检")
			sess.setPurpose(c.purpose)
			s.mgr.sessions[sess.ID] = sess

			var got *ScreenStartRequest
			s.SetScreenStarter(func(req ScreenStartRequest) (ScreenStartResult, error) {
				got = &req
				return ScreenStartResult{PID: 1234}, nil
			})

			resp := doReq(t, "POST", hs.URL+"/api/sessions/"+sess.ID+"/screen", "",
				`{"prog":"axmt500_wf"}`)
			if resp.StatusCode != http.StatusAccepted {
				t.Fatalf("挂监控该 202，实得 %d（%s）", resp.StatusCode, respError(t, resp))
			}
			if got == nil {
				t.Fatal("启动器没被调用")
			}
			if got.NoContinue != c.want {
				t.Fatalf("NoContinue=%v，用途 %s 该是 %v —— %s", got.NoContinue, c.purpose, c.want, c.why)
			}
			if got.Prog != "axmt500_wf" {
				t.Errorf("prog 该原样传下去，实得 %q", got.Prog)
			}
		})
	}
}
