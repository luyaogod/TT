package debug

import "testing"

// ---------- 放行类命令后的会话状态同步 ----------
//
// 真实事故（2026-11-11 福联 T100 排障）：fgldb 在 step/next 停下时，
// 某些发行版**只打源码窗、不打 `->` 箭头**（如 `1031        CALL xxx(...)`）。
// 旧实现把这种行当噪音吞掉 → stopCollect 从不创建 → 裸提示符收口时只
// completePending、不回写状态 → 会话永久停在 running：
//   之后每条 exec 都被 400「程序运行中,禁止发送调试命令」拒绝，
//   interrupt 又因为"进程其实停在提示符上"而毫无效果，只能 quit 重开。
//
// 下面两条用例把这两个入口都钉住；它们同时是"修复不破坏既有行为"的真值表。
// 带 `->` 箭头的既有行为由 waiting_test.go 的 TestStepStopWithoutHeaderIsRecognized 守护。

// 不带箭头的源码行 + 放行类命令在飞 = 一次停站。
func TestStopOnArrowlessSourceLine(t *testing.T) {
	s := waitTestSession()
	s.started = true
	evs := collectEvents(s)

	p := &pendingCmd{cmd: "next", mode: waitPrompt, res: make(chan *execResult, 1)}
	s.pending = p
	s.setState(StateRunning) // 与 execOpt 发出放行类命令后的状态一致

	s.onLine("1031        CALL s_bsft001_wf_gen_apmt500(l_sfaadocno)")
	s.onLine("(fgldb) ")

	select {
	case r := <-p.res:
		if r.Stop == nil {
			t.Fatal("停站应回填进命令响应")
		}
		if r.Stop.Line != 1031 {
			t.Fatalf("停站行应为 1031,实际 %d", r.Stop.Line)
		}
	default:
		t.Fatal("不带箭头的源码行也应当成停站并收口命令")
	}
	if !hasEvent(*evs, "stopped") {
		t.Fatal("必须发出 stopped 事件 —— 否则界面上的代码画面不会跟随")
	}
	if got := s.Cur().Line; got != 1031 {
		t.Fatalf("s.cur 应更新到 1031,实际 %d", got)
	}
	if s.State() != StateStopped {
		t.Fatalf("停站后状态应为 stopped,实际 %s", s.State())
	}
}

// 兜底：放行类命令在飞，但停站行一个字都没解析出来（连源码行都没有），
// 只要看到裸提示符，就必须回写 stopped —— 提示符是"程序停在调试器上"的铁证。
// 位置沿用上一次停站，不让界面因为拿不到行号而乱跳。
func TestPromptFallbackStopsRunningSession(t *testing.T) {
	s := waitTestSession()
	s.started = true
	evs := collectEvents(s)
	s.cur = StopInfo{Reason: "step", File: "csub_s_bsft001_wf.4gl", Line: 1023}

	p := &pendingCmd{cmd: "next", mode: waitPrompt, res: make(chan *execResult, 1)}
	s.pending = p
	s.setState(StateRunning)

	s.onLine("(fgldb) ")

	select {
	case r := <-p.res:
		if r.Stop == nil {
			t.Fatal("兜底收口也应回填一个停站位置")
		}
		if r.Stop.Line != 1023 {
			t.Fatalf("兜底位置应沿用 1023,实际 %d", r.Stop.Line)
		}
	default:
		t.Fatal("裸提示符应收口命令")
	}
	if !hasEvent(*evs, "stopped") {
		t.Fatal("兜底收口必须发出 stopped 事件")
	}
	if s.State() != StateStopped {
		t.Fatalf("兜底收口后状态应为 stopped,实际 %s", s.State())
	}
}

// 非放行类命令（list/where/info…）带出的源码行仍然不算停站 —— 不得误判。
func TestArrowlessSourceLineForNonResumeIsNotStop(t *testing.T) {
	s := waitTestSession()
	evs := collectEvents(s)
	p := &pendingCmd{cmd: "list", mode: waitPrompt, res: make(chan *execResult, 1)}
	s.pending = p

	s.onLine("812      DEFER INTERRUPT")
	s.onLine("(fgldb) ")

	if hasEvent(*evs, "stopped") {
		t.Fatal("非放行类命令的源码行不该被当成停站")
	}
	if got := s.Cur().Line; got != 0 {
		t.Fatalf("不该更新停站位置,实际 %d", got)
	}
	if s.State() != StateStopped {
		t.Fatalf("状态应保持 stopped,实际 %s", s.State())
	}
}

// continue 打出 Continuing. 之后必须如实翻 running —— 修复不能把这条也改坏。
func TestContinueStillMarksRunning(t *testing.T) {
	s := waitTestSession()
	s.started = true
	collectEvents(s)

	p := &pendingCmd{cmd: "continue", mode: waitPrompt, res: make(chan *execResult, 1)}
	s.pending = p
	s.setState(StateStopped)

	s.onLine("Continuing.")

	if s.State() != StateRunning {
		t.Fatalf("Continuing. 之后状态应为 running,实际 %s", s.State())
	}
}
