package fep

// A31（契约 .agents/plan/tt-genero-formdef-contract.md §9）：
//
//	preflight 的每个检查项只能有三种状态：
//	  ok        —— 明确就绪
//	  missing   —— **确定**缺失（能说清缺什么、给得出修复命令）→ 计入 missing，exit 2
//	  unknown   —— **检查本身失败**（解析不了/超时/依赖故障）→ **不计入 missing**，
//	               但要**明确打印**"这一项查不了（原因）"
//	退出码只由 missing 决定；unknown 不许升级成失败，也不许静默当 ok。
//
// 事故现场就是这个反例：`tt debug status --json` 输出两段 JSON，老解析器只吃第一段 ⇒
// session 项报"列会话失败" ⇒ 被当成"没有会话"计入 missing ⇒ 服务起着也 exit 2，
// 连带 `rec start` 永久卡死。**工具自己坏了，却说环境不全。**
//
// 全部走假依赖：不起服务、不连真机、不碰服务器。

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// errAccessDenied 是真机上"目录不可写"的原话（Low 完整性标签与 ACL 拒绝给的是同一句）。
var errAccessDenied = errors.New(`open C:\Users\x\AppData\Roaming\tt\fep\tt-preflight-1.tmp: Access is denied.`)

// brokenSessionLister 是"列会话这一步自己坏了"的生产形状：真解析一段不是 JSON 的报文。
func brokenSessionLister(*cobra.Command) ([]dbgSessionView, error) {
	return parseDebugSessions([]byte("这根本不是 JSON"))
}

// ③ 解析失败的降级：unknown，不计入 missing、不改变退出码，但要打印原因。
func TestPreflightSessionParseFailureIsUnknown(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)
	d.Sessions = brokenSessionLister

	rep := runPreflight(new(cobra.Command), "axmt500_wf", "主机109", d)

	// unknown 不计入 missing
	if rep.Missing != 0 {
		t.Fatalf("检查本身失败不许计入 missing（A31）：Missing=%d %+v", rep.Missing, rep.Checks)
	}
	if rep.Unknown != 1 {
		t.Fatalf("unknown 要单独记一笔（否则人看不出这一项没查成功）：Unknown=%d %+v", rep.Unknown, rep.Checks)
	}
	// 退出码不变：全 ok + 一项查不了 ⇒ 仍然是 0
	if !rep.OK {
		t.Fatalf("unknown 不许把清单判成失败（退出码只由 missing 决定）：%+v", rep.Checks)
	}
	if err := preflightExit(rep); err != nil {
		t.Fatalf("unknown 不许改变退出码，实得错误：%v", err)
	}
	// `rec start` 开跑前过同一份体检：不许被"查不了"拦死（这就是回归的现场）
	if err := recStartPreflightWith(new(cobra.Command), "axmt500_wf", "主机109", d); err != nil {
		t.Fatalf("检查本身失败不该拦 rec start（真机回归：rec start 必然 exit 2）：%v", err)
	}

	// 状态是 unknown，不是 ok 也不是 missing
	it := rep.item(preflightKeySession)
	if it == nil {
		t.Fatal("清单里必须有 session 这一项（查不了也要列出来）")
	}
	if it.state() != preflightStateUnknown {
		t.Fatalf("session 项的状态该是 unknown，实得 %q（%+v）", it.state(), *it)
	}
	if it.OK {
		t.Fatalf("unknown 不许静默当 ok：%+v", *it)
	}
	// 原因要留下（"列会话失败 + 报文哪里坏了"），否则人只看到问号
	if !strings.Contains(it.Detail, "列会话失败") {
		t.Fatalf("Detail 要写清这一项为什么查不了：%q", it.Detail)
	}
	if !strings.Contains(it.Why, "查不了") {
		t.Fatalf("Why 要用「查不了」这个口径（与 missing 的「缺什么」分开）：%q", it.Why)
	}

	// 人读态里必须**明确打印**"这一项查不了（原因）"
	txt := rep.text()
	if !strings.Contains(txt, "查不了") {
		t.Fatalf("人读态里没有「查不了」这一句 —— 那就成了静默当 ok：\n%s", txt)
	}
	if !strings.Contains(txt, "这根本不是 JSON") {
		t.Fatalf("「查不了」后面要带原因（原文开头）：\n%s", txt)
	}
	if !strings.Contains(txt, "调试会话") {
		t.Fatalf("要看得出是哪一项查不了：\n%s", txt)
	}
	// 不许把 unknown 混进「还缺 N 项前置」那张清单
	if strings.Contains(txt, "还缺") {
		t.Fatalf("unknown 不是缺失，不许进「还缺 N 项前置」：\n%s", txt)
	}
}

// unknown 与 missing 并存时：退出码仍然只由 missing 决定（这里 missing=1 ⇒ exit 2），
// 而且两条消息各归各位 —— 不得把 unknown 说成"缺前置"。
func TestPreflightUnknownDoesNotHideRealMissing(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)
	d.Sessions = brokenSessionLister
	d.Writable = func(string) error { return errAccessDenied }

	rep := runPreflight(new(cobra.Command), "axmt500_wf", "主机109", d)
	if rep.Missing != 1 || rep.Unknown != 1 {
		t.Fatalf("该是 missing=1 / unknown=1，实得 %d/%d：%+v", rep.Missing, rep.Unknown, rep.Checks)
	}
	err := preflightExit(rep)
	if err == nil {
		t.Fatal("有确定缺失时必须 exit 2（unknown 只是不参与，不是豁免）")
	}
	if got := exitCodeOf(t, err); got != exitUsage {
		t.Fatalf("退出码 = %d，想要 %d", got, exitUsage)
	}
	if !strings.Contains(err.Error(), "还缺 1 项前置") {
		t.Fatalf("「还缺」只数 missing：\n%s", err.Error())
	}
	if strings.Contains(err.Error(), "查不了") {
		t.Fatalf("查不了的项不许进「还缺 N 项前置」那个报错正文：\n%s", err.Error())
	}
}

// 三种状态是**互斥**的三值：每一项都得落在 ok/missing/unknown 之一，
// 而且 JSON 里的 `state` 与 `ok` 不许互相矛盾（脚本按 key+state 分支）。
func TestPreflightItemStatesAreExclusive(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)
	d.Sessions = brokenSessionLister

	rep := runPreflight(new(cobra.Command), "axmt500_wf", "主机109", d)
	if len(rep.Checks) != 4 {
		t.Fatalf("清单该有 4 项，实得 %d", len(rep.Checks))
	}
	for _, it := range rep.Checks {
		switch it.state() {
		case preflightStateOK:
			if !it.OK {
				t.Fatalf("state=ok 时 ok 该为 true：%+v", it)
			}
		case preflightStateMissing, preflightStateUnknown:
			if it.OK {
				t.Fatalf("state=%s 时 ok 不许为 true：%+v", it.state(), it)
			}
		default:
			t.Fatalf("状态只能是 ok/missing/unknown，实得 %q：%+v", it.state(), it)
		}
	}
}
