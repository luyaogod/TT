package debug

import (
	"strings"
	"testing"
	"time"
)

// 收口的"执行者 / 等待者"协议：后来者必须**等这一次收口做完**，而不是立刻回 OK。
//
// 事故背景（用户 2026-10-16 报"点了没反应"）：前一条命令（另一个进程的
// `tt debug quit`）与另一条命令的 quit 撞在一起，后来者拿到"ok"就往下走，
// 紧跟着的 `debug topent` 撞上"仅会话空闲(idle)时可设置 TOPENT(当前 运行中)"，
// 整条操作静默失败 —— 界面上只看到"点了没反应"。
func TestBeginQuitSharesOneTeardown(t *testing.T) {
	s := &Session{state: StateLoading}

	done, mine := s.beginQuit()
	if !mine {
		t.Fatal("第一个调用者应当是执行者")
	}

	done2, mine2 := s.beginQuit()
	if mine2 {
		t.Fatal("第二个调用者不该成为执行者（否则两次收口会打架）")
	}
	if done2 != done {
		t.Fatal("等待者必须等**同一个**信号，否则等的是一个永远不会关闭的 chan")
	}

	// 执行者做完（这里直接摆成 idle 态）并关闭信号 → 等待者才返回，且返回后状态可干活。
	// 直接改 state 而不走 forceExit/enterIdle：那两个要真 PTY 与 emit 回调，这里只验协议本身。
	go func() {
		time.Sleep(150 * time.Millisecond)
		s.mu.Lock()
		s.state = StateIdle
		s.mu.Unlock()
		close(done)
	}()
	start := time.Now()
	if err := s.waitQuit(done2); err != nil {
		t.Fatalf("等待者该成功返回: %v", err)
	}
	if time.Since(start) < 100*time.Millisecond {
		t.Fatal("等待者没有真的等那次收口（立刻返回正是这个 bug 的形态）")
	}
	if st := s.State(); st == StateRunning || st == StateLoading {
		t.Fatalf("等待返回时状态还不可用: %s", st)
	}
}

// 收口做完但状态仍是 running/loading（收口没生效）时要说出来，不能当成功。
func TestWaitQuitReportsIneffectiveTeardown(t *testing.T) {
	s := &Session{state: StateLoading}
	done := make(chan struct{})
	close(done)
	err := s.waitQuit(done)
	if err == nil {
		t.Fatal("状态还是 loading 时不该报成功")
	}
	if !strings.Contains(err.Error(), "结束调试") {
		t.Fatalf("错误文案要说清是哪一步没生效：%v", err)
	}
}
