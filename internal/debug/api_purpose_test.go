package debug

// 用途（`purpose`，W1）的判据。
//
// 三件事各钉一条：
//   1. 取值域只有 debug / record / replay，空 = debug（`ValidPurpose` + `Session.Purpose`）；
//   2. 同名 SSH 会话**用途不同**时不能悄悄返回既有那条 —— 名字是幂等键，键撞了但含义不同，
//      悄悄成功会让人以为"我要的录制会话建好了"，拿到的却是调试会话；
//   3. 恒定键：`SessionBrief` / 快照 / `tt session ls` 都要带它（键集由 api_snapshot_test 守）。

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestValidPurposeDomain(t *testing.T) {
	for _, ok := range []string{"", PurposeDebug, PurposeRecord, PurposeReplay} {
		if !ValidPurpose(ok) {
			t.Errorf("%q 该是合法用途（空 = debug）", ok)
		}
	}
	for _, bad := range []string{"Debug", "rec", "replay ", "录制", " ssh"} {
		if ValidPurpose(bad) {
			t.Errorf("%q 不该是合法用途：取值域只有 debug / record / replay", bad)
		}
	}
}

// 空用途的会话读出来必须是 debug（界面靠它决定藏不藏编辑器，空值不能让页面猜）。
func TestPurposeDefaultsToDebug(t *testing.T) {
	sess := idleOn("s1", envCfg("恒烁正式区"))
	if got := sess.Purpose(); got != PurposeDebug {
		t.Fatalf("新建会话的用途该是 %q，实得 %q", PurposeDebug, got)
	}
	sess.setPurpose("")
	if got := sess.Purpose(); got != PurposeDebug {
		t.Fatalf("setPurpose(\"\") 之后该回到 %q，实得 %q", PurposeDebug, got)
	}
	sess.setPurpose(PurposeRecord)
	if got := sess.Purpose(); got != PurposeRecord {
		t.Fatalf("setPurpose(record) 之后该是 record，实得 %q", got)
	}
}

// 用途取值非法 → 400 + USAGE，且**不建会话**（校验在拨号之前）。
func TestSessionNewRejectsUnknownPurpose(t *testing.T) {
	hs, s := actionTestDebugServer(t)
	giveServerNamedEnv(s, "h1", "127.0.0.1", "36")

	resp := doReq(t, "POST", hs.URL+"/api/sessions/new", "",
		`{"env":"h1","ent":"99","name":"录像A","purpose":"recording"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("非法用途该退 400，实得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	ok, code, msg := envelopeOf(t, resp)
	if ok || code != "USAGE" {
		t.Fatalf("该回 ok=false + code=USAGE，实得 ok=%v code=%q", ok, code)
	}
	if !strings.Contains(msg, "debug") || !strings.Contains(msg, "record") || !strings.Contains(msg, "replay") {
		t.Errorf("文案要把合法取值列出来（人要靠它照做），实得 %q", msg)
	}
	if n := len(s.mgr.Snapshot()); n != 0 {
		t.Errorf("被拒的请求不该留下会话，实得 %d 条", n)
	}
}

// **本文件的重点**：同名但用途不同 → 409 + SESSION_PURPOSE_MISMATCH，池子一个字不动。
func TestSessionNewRejectsPurposeMismatch(t *testing.T) {
	hs, s := actionTestDebugServer(t)
	giveServerNamedEnv(s, "h1", "127.0.0.1", "36")
	live := idleOn("s-live", s.cfg.CloneEnv("h1"))
	live.setNameInTest("录像A")
	live.setPurpose(PurposeRecord)
	s.mgr.sessions[live.ID] = live

	// 名字相同，但这次要的是 debug（缺省）。
	resp := doReq(t, "POST", hs.URL+"/api/sessions/new", "",
		`{"env":"h1","ent":"99","name":"录像A"}`)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("用途不同该退 409，实得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	ok, code, msg := envelopeOf(t, resp)
	if ok || code != "SESSION_PURPOSE_MISMATCH" {
		t.Fatalf("该回 ok=false + code=SESSION_PURPOSE_MISMATCH，实得 ok=%v code=%q", ok, code)
	}
	// 文案要能照做：说清既有那条是什么用途、这次要的是什么、下一步怎么办。
	for _, want := range []string{"record", "debug", "换一个名字"} {
		if !strings.Contains(msg, want) {
			t.Errorf("文案缺 %q（人要靠它照做）：%q", want, msg)
		}
	}
	if got := s.mgr.Get(live.ID); got == nil {
		t.Fatal("被拒时那条同名会话必须原样留着")
	}
	if n := len(s.mgr.Snapshot()); n != 1 {
		t.Fatalf("不该多出一条会话，实得 %d 条", n)
	}
}

// 同名 + 同用途 → 幂等（返回既有那条），并且回显的 purpose 是既有那条的。
func TestSessionNewSamePurposeReusedEchoesPurpose(t *testing.T) {
	hs, s := actionTestDebugServer(t)
	giveServerNamedEnv(s, "h1", "127.0.0.1", "36")
	live := idleOn("s-live", s.cfg.CloneEnv("h1"))
	live.setNameInTest("录像A")
	live.setPurpose(PurposeRecord)
	s.mgr.sessions[live.ID] = live

	resp := doReq(t, "POST", hs.URL+"/api/sessions/new", "",
		`{"env":"h1","ent":"99","name":"录像A","purpose":"record"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("同名同用途应幂等，实得 %d（%s）", resp.StatusCode, respError(t, resp))
	}
	var out struct {
		SessionID string `json:"sessionId"`
		Purpose   string `json:"purpose"`
		Reused    bool   `json:"reused"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("应答解不开: %v", err)
	}
	if out.SessionID != live.ID || !out.Reused {
		t.Fatalf("该幂等返回既有那条（%s, reused=true），实得 %q reused=%v", live.ID, out.SessionID, out.Reused)
	}
	if out.Purpose != PurposeRecord {
		t.Errorf("回显的 purpose 该是 %q，实得 %q", PurposeRecord, out.Purpose)
	}
}
