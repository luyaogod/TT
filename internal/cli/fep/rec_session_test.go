package fep

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// 两段判据同住这个文件：
//   - `parseWaitResult`：真机抓下来的等待报文（形状反直觉，见下）；
//   - WP5 / T5.5（缺陷 **D**）：录制起会话时不许误杀别人的会话（见文件下半）。

// 这段报文是**真机抓下来的**（2026-10-15 18:52，`tt debug wait --for stopped,exit --json`
// 在断点命中那一刻的原文，只删了中间几行 source）。留它是因为报文的形状反直觉：
// `event` 是**对象**不是字符串，而停站现场在顶层 `stop` 与 `event.stop` 里各有一份 ——
// 第一次按直觉写成 `Event string` 时，断点明明命中了，命令却报"读不懂等待结果"。
const realWaitPayload = `{"event":{"type":"stopped","sessionId":"s1792061476981",
"time":"0001-01-01T00:00:00Z","stop":{"reason":"breakpoint","bpNum":1,
"func":"asf_bsft001_wf.bsft001_wf_input()","file":"asf_bsft001_wf.4gl","line":13645,
"source":[{"num":13642,"text":"#add-point:AFTER FIELD sfajwfdocno name=\"input.a.sfajwfdocno\"","isCur":false},
{"num":13645,"text":"IF  NOT cl_null(g_sfajwf_m.sfajwfdocno) THEN","isCur":true}]}},
"ok":true,"silentSeconds":0,"state":"stopped",
"stop":{"reason":"breakpoint","bpNum":1,"func":"asf_bsft001_wf.bsft001_wf_input()",
"file":"asf_bsft001_wf.4gl","line":13645}}`

func TestParseWaitResultReadsBreakpointStop(t *testing.T) {
	spot, timedOut, err := parseWaitResult([]byte(realWaitPayload))
	if err != nil {
		t.Fatalf("真机报文必须读得懂：%v", err)
	}
	if timedOut {
		t.Fatal("命中断点不是超时")
	}
	if spot.File != "asf_bsft001_wf.4gl" || spot.Line != 13645 {
		t.Fatalf("停站位置读错：%+v", spot)
	}
	if !strings.Contains(spot.Func, "bsft001_wf_input") || spot.Reason != "breakpoint" {
		t.Fatalf("函数/原因读错：%+v", spot)
	}
	if !strings.Contains(spot.String(), "asf_bsft001_wf.4gl:13645") {
		t.Fatalf("给人读的那行要带 file:line：%s", spot.String())
	}
}

// 只有 event 里带 stop（顶层没有）时也要认 —— 报文有两份现场，不能只认一份。
func TestParseWaitResultFallsBackToEventStop(t *testing.T) {
	payload := `{"event":{"type":"stopped","stop":{"file":"x.4gl","line":7,"reason":"breakpoint"}},"ok":true}`
	spot, timedOut, err := parseWaitResult([]byte(payload))
	if err != nil || timedOut {
		t.Fatalf("该读得懂且不算超时：%v timedOut=%v", err, timedOut)
	}
	if spot.File != "x.4gl" || spot.Line != 7 {
		t.Fatalf("该认 event.stop：%+v", spot)
	}
}

// 超时 / 退出（没有 stop 现场）：算"没捕获到现场"，但要与"读不懂"分开。
func TestParseWaitResultTimeoutAndExit(t *testing.T) {
	spot, timedOut, err := parseWaitResult([]byte(`{"event":{"type":"timeout"},"ok":true,"timedOut":true,"state":"running"}`))
	if err != nil {
		t.Fatalf("超时是正常收场，不该报错：%v", err)
	}
	if !timedOut || spot.File != "" {
		t.Fatalf("超时该报 timedOut 且没有现场：%+v timedOut=%v", spot, timedOut)
	}
	if _, timedOut, err := parseWaitResult([]byte(`{"event":{"type":"exit"},"ok":true,"state":"exit"}`)); err != nil || !timedOut {
		t.Fatalf("程序结束也算没捕获到现场：timedOut=%v err=%v", timedOut, err)
	}
	if _, _, err := parseWaitResult([]byte(`不是 JSON`)); err == nil {
		t.Fatal("读不懂要报错（与「没捕获到现场」分开：一个是服务坏了，一个是没有断点）")
	}
}

// ---------------------------------------------------------------------------
// WP5 / T5.5（缺陷 D）：录制起会话时**不许误杀别人的会话**
// ---------------------------------------------------------------------------
//
// 病根在 `rec_session.go` 的老写法：
//
//	if busy, _ := debugSessionOpen(cmd); busy { runSelf(cmd, "debug", "quit") }
//
// 判据只有"有没有会话在跑"，**不看那条会话是谁的**。多会话之后，
// B 开始录制会把正在跑断点的 A 的会话一枪 `quit` 掉，而 A 那边看起来只是
// "会话莫名其妙没了"。
//
// 这里钉的是**收口前的判据**（`staleSessionToCollect` / `recSessionOwned`）：
// 它必须在"该收的那条"上认领，在"别人的会话"上放过。
// 真起调试服务 + 真 SSH 的那一半不在这里（见 `internal/debug/README.md` 的"故意不覆盖"）。

// withFakeDebugSessions 把"列会话"这一步换成假的（真实现要起 tt 子进程连调试服务）。
func withFakeDebugSessions(t *testing.T, raw string) {
	t.Helper()
	old := debugSessions
	debugSessions = func(*cobra.Command) ([]dbgSessionView, error) {
		return parseDebugSessions([]byte(raw))
	}
	t.Cleanup(func() { debugSessions = old })
}

// 报文形状取自 `tt debug status --json` 的 `sessions` 数组（服务端 `SessionBrief`）。
const (
	// A：正在跑断点的会话，环境"恒烁正式区"
	fakeSessionsAOnly = `{"sessions":[
	  {"id":"sA","env":"恒烁正式区","prog":"apmt520","state":"stopped"}
	]}`
	// A + 自己上一轮留下的那条（同环境）
	fakeSessionsSameEnv = `{"sessions":[
	  {"id":"sA","env":"恒烁正式区","prog":"apmt520","state":"running"},
	  {"id":"sOld","env":"恒烁正式区","prog":"bsft001_wf","state":"idle"}
	]}`
)

// **D 的验收**：A 停在断点上时，B 另起一轮录制/回放 → A 的会话不受影响。
//
// B 录的是**另一个环境**（"测试1012"）。修前这里会返回"有会话在跑"，
// 于是紧接着那一枪 `debug quit` 会打在被服务端唯一认出的 A 身上 ——
// 多会话下就是"A 的会话莫名其妙没了"。
func TestStaleSessionToCollectIgnoresOtherEnvironments(t *testing.T) {
	withFakeDebugSessions(t, fakeSessionsAOnly)

	s := recSession{Env: "测试1012", Prog: "bsft001_wf"}
	why, err := staleSessionToCollect(new(cobra.Command), s, "测试1012")
	if err != nil {
		t.Fatalf("列会话失败: %v", err)
	}
	if why != "" {
		t.Fatalf("B 在另一个环境起录制时收掉了别人的会话（%q）—— 那就是缺陷 D："+
			"A 停在断点上，B 一录制就把它的会话 quit 掉，A 那边看起来只是「会话莫名其妙没了」", why)
	}
}

// 反向判据：**自己上一轮**留下的会话照旧要收（"有就收"这条便利对同环境仍然成立，
// 否则 TOPENT 设不上去，`rec reproduce` 会在那一步退 1 —— 老注释里写着实测）。
func TestStaleSessionToCollectTakesItsOwnEnvironment(t *testing.T) {
	withFakeDebugSessions(t, fakeSessionsSameEnv)

	s := recSession{Env: "恒烁正式区", Prog: "bsft001_wf"}
	why, err := staleSessionToCollect(new(cobra.Command), s, "恒烁正式区")
	if err != nil {
		t.Fatalf("列会话失败: %v", err)
	}
	if why == "" {
		t.Fatal("自己环境里挂着的会话必须照旧收掉（不收：TOPENT 设不上去，录制第二步被拒）")
	}
	if !strings.Contains(why, "恒烁正式区") {
		t.Fatalf("收口说明该点名是哪条会话：%q", why)
	}
}

// 录制身份那一半：同环境但**作业不同**的、正在跑的会话不是我们起的那一条。
func TestRecSessionOwnedMatchesRecordingIdentity(t *testing.T) {
	mine := dbgSessionView{ID: "s1", Env: "恒烁正式区", Prog: "bsft001_wf", State: "idle"}
	cases := []struct {
		name string
		sess dbgSessionView
		s    recSession
		env  string
		want bool
	}{
		{"同环境同作业 → 是自己那条", mine, recSession{Prog: "bsft001_wf"}, "恒烁正式区", true},
		{"别的环境 → 不是", mine, recSession{Prog: "bsft001_wf"}, "测试1012", false},
		{"环境不明 → 不认领（宁可多留一条，也不误杀）", dbgSessionView{ID: "s2", Env: "", Prog: "bsft001_wf"}, recSession{Prog: "bsft001_wf"}, "", false},
		{"对方环境不明 → 不认领", dbgSessionView{ID: "s3", Env: "", Prog: "bsft001_wf"}, recSession{Prog: "bsft001_wf"}, "恒烁正式区", false},
		{"同环境但作业不同 → 不是（那是别人正在跑的那条）", dbgSessionView{ID: "s4", Env: "恒烁正式区", Prog: "apmt520", State: "running"}, recSession{Prog: "bsft001_wf"}, "恒烁正式区", false},
		{"没给 --prog 时不靠作业判（环境相同即认）", dbgSessionView{ID: "s5", Env: "恒烁正式区", Prog: "apmt520"}, recSession{}, "恒烁正式区", true},
		{"对方作业未知时不靠作业判", dbgSessionView{ID: "s6", Env: "恒烁正式区"}, recSession{Prog: "bsft001_wf"}, "恒烁正式区", true},
	}
	for _, c := range cases {
		if got := recSessionOwned(c.sess, c.s, c.env); got != c.want {
			t.Errorf("%s：recSessionOwned = %v，期望 %v", c.name, got, c.want)
		}
	}
}

// `parseDebugSessions` 必须真的把"这条会话是谁的"读出来 —— 老实现只声明了 `state`，
// 于是判据在客户端**根本没有材料**（缺陷 D 能成立的另一半原因）。
func TestParseDebugSessionsReadsIdentity(t *testing.T) {
	got, err := parseDebugSessions([]byte(fakeSessionsSameEnv))
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("该读出 2 条会话，实得 %d", len(got))
	}
	if got[0].ID != "sA" || got[0].Env != "恒烁正式区" || got[0].Prog != "apmt520" || got[0].State != "running" {
		t.Fatalf("会话身份没读全:%+v", got[0])
	}
}
