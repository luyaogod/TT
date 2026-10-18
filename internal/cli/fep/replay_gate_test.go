package fep

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tt/internal/fep"
)

// ── A23 的产品面：两条**实际发包**的入口都必须先过身份闸 ────────────────────
//
// 两个入口（`tt fep replay` 与 `tt recording replay`）共用一条组装路径
// `planReplayFor`，发包也只走一条：`Client.RunReplayIdentity`（= 身份闸 + 原来那条）。
// 这里对着那条路断言两件事：
//
//	1. 计划里有没有身份，是**录制件自己**决定的（没有界面侧 ⇒ 没有身份 ⇒ 早于身份上线）；
//	2. 一旦有这么一步，拒绝发生在**任何发包之前** —— 注入口一条都收不到；
//	3. `--dry-run`：能展示，而且每步都标「无身份（真跑按临时编号发送）」。

// legacyRecFixture 造一份"早于身份上线"的录制：只有事件侧能读出事件，
// 界面侧里那些对象**没有 colName** ⇒ 计划里没有身份（正是老件的形态）。
func legacyRecFixture(t *testing.T) (recTarget, string) {
	t.Helper()
	dir := t.TempDir()
	id := "t-old"
	aui := filepath.Join(dir, id+"-1-rt2gdc.log")
	ev := filepath.Join(dir, id+"-1-gdc2rt.log")
	// 界面侧有对象，但没有 colName/fieldId（A23 之前录的件就是这个样子）
	writeFrames(t, aui,
		`om 1 {{an 0 Window 4097 {{name "w_axmt500"}} {{`+
			`{an 4097 FormField 6000 {{name "formonly.b_target"}{text "目标字段"}} {}} `+
			`}}}}`,
	)
	writeFrames(t, ev, `event _om 1{}{{ConfigureEvent 0{{idRef "6000"}{ value "X"}}}}`)
	return recTarget{ID: id, Dir: dir, Files: []string{aui}}, dir
}

// TestReplaySendsIdentitylessPlanByBareID：**2026-10-18 的决策**（替换原来的"整份拒绝"）。
//
// 老件（没有 colName）不再被一票否决：照录制时的临时编号发出去，并且**每一步都标出来**。
// 这一条用产品那条组装路径（`planReplayFor`）+ 注入传输来断言"确实发了、也确实标了"。
func TestReplaySendsIdentitylessPlanByBareID(t *testing.T) {
	tgt, dir := legacyRecFixture(t)
	events, err := tgt.events()
	if err != nil {
		t.Fatalf("读事件侧: %v", err)
	}
	plan := planReplayFor(tgt, events, replayStopAt, replaySkip)
	if len(plan.Events) != 1 {
		t.Fatalf("计划条数 = %d，想要 1", len(plan.Events))
	}
	if plan.Events[0].ColName != "" {
		t.Fatalf("这份夹具本该没有身份：%+v", plan.Events[0])
	}
	if n := fep.CountUnidentified(plan); n != 1 {
		t.Fatalf("该报 1 步降级，实得 %d", n)
	}

	client, sent := fakeInject(t, dir)
	res, rerr := client.RunReplayIdentity(plan, fep.ReplayRunOptions{Conn: 1, Settle: 10 * time.Millisecond})
	if rerr != nil {
		t.Fatalf("没有身份不再整份拒绝，实得 %v", rerr)
	}
	if res == nil || res.Sent != 1 || res.Degraded != 1 {
		t.Fatalf("结果 = %+v，想要 sent=1 degraded=1", res)
	}
	if got := sent(); len(got) != 1 {
		t.Fatalf("该发 1 条（按临时编号），实得 %d 条：%v", len(got), got)
	}
}

// TestDryRunTextMarksIdentitylessSteps：**例外**必须写在脸上（A23 第二段）。
func TestDryRunTextMarksIdentitylessSteps(t *testing.T) {
	plan := fep.ReplayPlan{Events: []fep.ReplayEvent{
		{N: 1, Event: `ConfigureEvent 0{{idRef "6000"}{ value "X"}}`}, // 无身份
		{N: 2, Event: `ActionEvent 0{{idRef "6008"}}`, ColName: "b_target", FieldID: 5},
		{N: 3, Event: `KeyEvent 0{{keyName "Tab"}}`}, // 没有目标：装不上身份，也不标
	}}
	text := replayDryRunText(plan)
	if !strings.Contains(text, fep.NoIdentityNote) {
		t.Fatalf("dry-run 输出里没有「%s」：\n%s", fep.NoIdentityNote, text)
	}
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) < 3 {
		t.Fatalf("每条步骤都要有一行：\n%s", text)
	}
	// 只有第 1 步带标注：有身份的那步、没有目标的那步都不许标
	if !strings.Contains(lines[0], fep.NoIdentityNote) {
		t.Fatalf("第 1 步（无身份）没被标注：%q", lines[0])
	}
	if strings.Contains(lines[1], fep.NoIdentityNote) {
		t.Fatalf("第 2 步有身份却被标注了：%q", lines[1])
	}
	if strings.Contains(lines[2], fep.NoIdentityNote) {
		t.Fatalf("第 3 步没有目标对象，不该按「无身份」标注：%q", lines[2])
	}
	// 还要说清"这些不会发出去"
	if !strings.Contains(text, "一个字节都没发") {
		t.Fatalf("dry-run 的收尾要写清没发包：\n%s", text)
	}
	// 空计划也别崩
	if s := replayDryRunText(fep.ReplayPlan{}); !strings.Contains(s, "一个字节都没发") {
		t.Fatalf("空计划也该给出收尾行：%q", s)
	}
}

// TestDryRunJSONCarriesIdentitylessSteps：默认输出是 json，所以标注也必须在 json 里。
func TestDryRunJSONCarriesIdentitylessSteps(t *testing.T) {
	plan := fep.ReplayPlan{Events: []fep.ReplayEvent{
		{N: 1, Event: `ConfigureEvent 0{{idRef "6000"}{ value "X"}}`},
		{N: 2, Event: `ActionEvent 0{{idRef "6008"}}`, ColName: "b_target", FieldID: 5},
	}}
	b, err := json.Marshal(replayDryRunReport(plan))
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, fep.NoIdentityNote) {
		t.Fatalf("json 里没有「%s」：%s", fep.NoIdentityNote, s)
	}
	if !strings.Contains(s, `"n":1`) || !strings.Contains(s, `"note"`) {
		t.Fatalf("标注要挂在具体那一步上：%s", s)
	}
	if strings.Count(s, fep.NoIdentityNote) != 1 {
		t.Fatalf("只有一步无身份，标注不该出现两次：%s", s)
	}
}
