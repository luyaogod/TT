package fep

import (
	"strings"
	"testing"
)

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
