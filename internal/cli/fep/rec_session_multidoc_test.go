package fep

// A31 的第一半（契约 .agents/plan/tt-genero-formdef-contract.md §9 A30/A31，
// 2026-10-18 真机）：`tt debug status --json` 实测输出**两段 JSON** ——
//
//	第一段 = GET /api/status 的原样应答（internal/debug/api.go 的 hStatus）
//	第二段 = GET /api/sessions 的原样应答 {"ok":true,"sessions":[…]}
//
// 两句 fmt.Println 就是它的全部实现（internal/cli/debug/debugctl.go 的 debugStatusCmd），
// 所以报文里必然是「一段 JSON + 换行 + 一段 JSON」。
//
// 老 `parseDebugSessions` 只做**一次** json.Unmarshal ⇒
// `invalid character '{' after top-level value` ⇒ preflight 的 session 项报
// "列会话失败" ⇒ 服务明明起着，`tt recording start` 的前置体检 也 exit 2；
// 而 `rec start` 开跑前过同一份体检 ⇒ **`rec start` 必然 exit 2**，path=debug 的件录不下来。
//
// 本文件只喂字符串给纯函数（不连真机、不起会话、不碰服务器）。

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// twoDocStatusJSON 是实测报文的形状：服务状态 doc + 会话 doc（两段，换行分隔）。
const twoDocStatusJSON = `{"server":"tdebug-debug","ssh":"主机109","zone":"t","listen":"127.0.0.1:28670","watchdog":120}
{"ok":true,"sessions":[
  {"id":"sA","env":"恒烁正式区","prog":"apmt520","state":"stopped","name":"fep-20261018-101500"}
]}`

// ① 两段 JSON：必须能取出 sessions（回归的第一现场）。
func TestParseDebugSessionsReadsTwoDocuments(t *testing.T) {
	got, err := parseDebugSessions([]byte(twoDocStatusJSON))
	if err != nil {
		t.Fatalf("两段报文必须读得懂（真机回归：只 unmarshal 一次 = "+
			"invalid character '{' after top-level value）：%v", err)
	}
	if len(got) != 1 {
		t.Fatalf("该读出 1 条会话，实得 %d：%+v", len(got), got)
	}
	if got[0].ID != "sA" || got[0].Env != "恒烁正式区" || got[0].Name != "fep-20261018-101500" {
		t.Fatalf("会话身份没读全（判据要靠它认领自己那条）：%+v", got[0])
	}
}

// ② 三段 + 空行 + CRLF 也不许炸：报文的段数不是契约，别把段数焊死。
//
// 中间那段落故意带一个**标量** `"sessions":0`：要找的是"带 sessions **数组**"的那一段，
// 不是"带 sessions 这个键"的那一段（服务端哪天在状态里放个计数，解析器不该当场躺下）。
func TestParseDebugSessionsToleratesMoreDocumentsAndBlankLines(t *testing.T) {
	raw := strings.Join([]string{
		`{"server":"tdebug-debug","listen":"127.0.0.1:28670"}`,
		"",
		`  {"ok":true,"sessions":0}  `,
		"",
		`{"ok":true,"sessions":[{"id":"s1","env":"测试1012","prog":"bsft001_wf","state":"idle"}]}`,
		"",
	}, "\r\n")
	got, err := parseDebugSessions([]byte(raw))
	if err != nil {
		t.Fatalf("三段/带空行/CRLF 的报文必须读得懂：%v", err)
	}
	if len(got) != 1 || got[0].ID != "s1" {
		t.Fatalf("该从那一段数组里取出 s1，实得 %+v", got)
	}
}

// 两段都对、但第一段里也带 sessions 数组时：认**第一个**数组（服务端的会话 doc 在第二段，
// 这里只钉"从所有文档里找"这条不变量，不钉优先级细节）。
func TestParseDebugSessionsTakesTheArrayDocument(t *testing.T) {
	raw := `{"sessions":[{"id":"first"}]}
{"ok":true,"sessions":[{"id":"second"}]}`
	got, err := parseDebugSessions([]byte(raw))
	if err != nil {
		t.Fatalf("两段都带数组时不该报错：%v", err)
	}
	if len(got) != 1 || got[0].ID != "first" {
		t.Fatalf("该认第一个数组，实得 %+v", got)
	}
}

// 空会话清单（`{"sessions":[]}`）是**读得懂**的"没有会话"，与"读不懂"必须分开：
// 前者 preflight 该报 missing（确定缺失），后者只能报 unknown（A31）。
func TestParseDebugSessionsReadsEmptyList(t *testing.T) {
	got, err := parseDebugSessions([]byte(`{"server":"tdebug-debug"}
{"ok":true,"sessions":[]}`))
	if err != nil {
		t.Fatalf("空清单不是解析失败：%v", err)
	}
	if len(got) != 0 {
		t.Fatalf("空清单该是 0 条，实得 %+v", got)
	}
}

// 完全不是 JSON：报错，且**错误里带原文开头**（现场要能看出收到的是什么）。
func TestParseDebugSessionsGarbageKeepsTheOriginal(t *testing.T) {
	got, err := parseDebugSessions([]byte("这根本不是 JSON"))
	if err == nil {
		t.Fatalf("不是 JSON 必须报错，得 %+v", got)
	}
	if !strings.Contains(err.Error(), "这根本不是 JSON") {
		t.Fatalf("错误里要留下原文开头（否则现场无从下手）：%v", err)
	}
}

// 一段 JSON 但没有 sessions 数组：也算读不出来（宁可报 unknown，也不假装"没有会话"）。
func TestParseDebugSessionsWithoutSessionsDoc(t *testing.T) {
	if got, err := parseDebugSessions([]byte(`{"server":"tdebug-debug"}`)); err == nil {
		t.Fatalf("没有 sessions 那一段时必须报错（不许当成「0 条会话」）：%+v", got)
	}
}

// 回归的最终现场：preflight 走**真解析**（两段报文）时，session 项必须 ok。
//
// 修前这里报 "列会话失败：invalid character '{' after top-level value"，于是
// `tt recording start` 的前置体检 exit 2 —— 服务明明起着，工具自己坏了却说环境不全。
func TestPreflightSessionOKOnTwoDocStatusJSON(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)
	d.Sessions = func(*cobra.Command) ([]dbgSessionView, error) {
		return parseDebugSessions([]byte(twoDocStatusJSON))
	}

	rep := runPreflight(new(cobra.Command), "axmt500_wf", "主机109", d)
	if !rep.OK || rep.Missing != 0 {
		t.Fatalf("服务起着、会话也在，体检必须就绪：%+v", rep.Checks)
	}
	it := rep.item(preflightKeySession)
	if it == nil || it.state() != preflightStateOK {
		t.Fatalf("session 项该 ok：%+v", it)
	}
	if !strings.Contains(it.Detail, "fep-20261018-101500") {
		t.Fatalf("session 项该点名这一轮用得上的那条（别名）：%q", it.Detail)
	}
	if err := preflightExit(rep); err != nil {
		t.Fatalf("全就绪必须退 0：%v", err)
	}
	if err := recStartPreflightWith(new(cobra.Command), "axmt500_wf", "主机109", d); err != nil {
		t.Fatalf("`rec start` 开跑前的同一份体检不许拦（真机回归：它必然 exit 2）：%v", err)
	}
}
