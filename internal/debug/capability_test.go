package debug

// 三类能力 × 用途 × 模式 × 角色（W7）。
//
// 为什么要有这一组：`record` 用途要"页面不给编辑器、命令一律不发"，但**录制进程自己**
// 还要靠「起作业 + 起界面会话」才跑得起来（真机第一次挂画面就死在"一刀切只读"上）。
// 所以判据拆成三类，真值表在这里钉死；端点行为用两条请求钉住 —— 光有真值表不够，
// 得证明闸门真的接在端点上。

import (
	"net/http"
	"strings"
	"testing"
)

// capSession 造一条夹具会话：指定用途与模式。
func capSession(id, purpose, mode string) *Session {
	s := idleOn(id, envCfg("恒烁正式区"))
	// 夹具要**非 Bare**（Bare = prog/module 都空，是"还没有作业的空宿主"，模式闸门对它放行）。
	// 真值表判的是"有作业的会话"，所以给它一个作业名。
	s.Prog = "apmt520"
	s.setPurpose(purpose)
	s.SetMode(mode)
	return s
}

// 真值表：purpose × mode × actor × capability。
//
// 规则（W7）：
//   - CapOwn 任何用途都放行；
//   - CapInteractive 在 record 下一律拒；
//   - 其余按模式：solo 只有人可写、collab 只有 ai 可写。
func TestMayDoTruthTable(t *testing.T) {
	cases := []struct {
		purpose, mode, actor string
		cap                  Capability
		want                 bool
		why                  string
	}{
		// CapOwn：谁都收得掉（收不掉比收错了更糟，H5）
		{PurposeDebug, ModeSolo, "human", CapOwn, true, "solo 下人可以收尾"},
		{PurposeDebug, ModeSolo, "ai", CapOwn, true, "solo 下 AI 也可以收尾（收尾不受模式闸门）"},
		{PurposeRecord, ModeCollab, "human", CapOwn, true, "record 的收尾必须允许"},
		{PurposeRecord, ModeCollab, "ai", CapOwn, true, "同上"},

		// CapSetup：按模式（录制进程靠它起作业）
		{PurposeDebug, ModeSolo, "human", CapSetup, true, "solo 下人可以起作业"},
		{PurposeDebug, ModeSolo, "ai", CapSetup, false, "solo 下 AI 不能"},
		{PurposeDebug, ModeCollab, "ai", CapSetup, true, "collab 下 AI 可以"},
		{PurposeRecord, ModeSolo, "human", CapSetup, true, "record 下编排仍允许（录制进程/发起者）"},
		{PurposeRecord, ModeCollab, "ai", CapSetup, true, "同上"},

		// CapInteractive：record 一律拒
		{PurposeDebug, ModeSolo, "human", CapInteractive, true, "debug 会话照旧"},
		{PurposeDebug, ModeCollab, "ai", CapInteractive, true, "collab 下 AI 可写"},
		{PurposeDebug, ModeCollab, "human", CapInteractive, false, "collab 下人只读"},
		{PurposeReplay, ModeSolo, "human", CapInteractive, true, "回放：driver（solo=人）可写"},
		{PurposeReplay, ModeCollab, "ai", CapInteractive, true, "回放：driver（collab=AI）可写"},
		{PurposeRecord, ModeSolo, "human", CapInteractive, false, "record：人也不许发调试命令"},
		{PurposeRecord, ModeCollab, "ai", CapInteractive, false, "record：AI 也不许（防误操作）"},
	}
	for _, c := range cases {
		s := capSession(c.purpose+"-"+c.mode+"-"+c.actor, c.purpose, c.mode)
		if got := s.MayDo(c.cap, c.actor); got != c.want {
			t.Errorf("MayDo(purpose=%s mode=%s actor=%s cap=%d) = %v，该是 %v —— %s",
				c.purpose, c.mode, c.actor, c.cap, got, c.want, c.why)
		}
	}
}

// 端点行为：record 会话的 /control 回 403 + SESSION_READONLY；/job 不被用途拦。
func TestRecordSessionBlocksInteractiveButAllowsSetup(t *testing.T) {
	hs, s := actionTestDebugServer(t)
	giveServerNamedEnv(s, "h1", "127.0.0.1", "36")
	sess := idleOn("s-rec", s.cfg.CloneEnv("h1"))
	sess.setNameInTest("录像A")
	sess.setPurpose(PurposeRecord)
	s.mgr.sessions[sess.ID] = sess

	// 交互式：拒，且码是 SESSION_READONLY（Agent 靠码分支）
	resp := doReq(t, "POST", hs.URL+"/api/sessions/"+sess.ID+"/control", "", `{"action":"continue"}`)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("/control 对 record 会话该退 403，实得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	ok, code, msg := envelopeOf(t, resp)
	if ok || code != "SESSION_READONLY" {
		t.Fatalf("该回 ok=false + code=SESSION_READONLY，实得 ok=%v code=%q", ok, code)
	}
	for _, want := range []string{"纯录制", "另建一条 debug", "停止录制"} {
		if !strings.Contains(msg, want) {
			t.Errorf("文案缺 %q（人要靠它照做）：%q", want, msg)
		}
	}

	// 编排：不该被用途拦（会走到更后面因测试环境失败，但不能是 403/READONLY）
	resp2 := doReq(t, "POST", hs.URL+"/api/sessions/"+sess.ID+"/job", "", `{"prog":"axmt500_wf"}`)
	if resp2.StatusCode == http.StatusForbidden {
		t.Fatalf("/job 在 record 会话上不该被用途拒（录制进程就是靠它起作业）：%s", respError(t, resp2))
	}
	if _, code2, _ := envelopeOf(t, resp2); code2 == "SESSION_READONLY" {
		t.Fatal("/job 不该回 SESSION_READONLY")
	}
}

// debug 会话的交互式闸门保持原样（H2）：solo + 人 → 放行。
func TestDebugSessionKeepsInteractiveGate(t *testing.T) {
	hs, s := actionTestDebugServer(t)
	giveServerNamedEnv(s, "h1", "127.0.0.1", "36")
	sess := idleOn("s-dbg", s.cfg.CloneEnv("h1"))
	sess.setPurpose(PurposeDebug)
	s.mgr.sessions[sess.ID] = sess

	resp := doReq(t, "POST", hs.URL+"/api/sessions/"+sess.ID+"/control", "", `{"action":"continue"}`)
	if resp.StatusCode == http.StatusForbidden {
		t.Fatalf("debug 会话 + solo + 人不该被拒：%s", respError(t, resp))
	}
}


