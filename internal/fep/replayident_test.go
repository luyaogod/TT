package fep

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 回放步骤的**身份**（契约 C3 + A7 + A18；H1/H2）。
//
// 这一组测试守三件事：
//
//  1. **录制时写入**：步骤里出现 `ColName`/`FieldID`（从当步目标节点取）；
//  2. **旧数据照旧**：没有这两个字段的旧计划/旧日志仍能读、能跑，行为一字不变；
//  3. **回放优先用身份**：步骤带 `ColName` 时每步现解析（`ByColName`）、用完即弃；
//     解析不过 → 按四码拒绝并**一个字节都不发**。
//
// 反证（架构师会亲手复现的那种）：把"优先用身份"改回"仍用旧 id"，
// TestReplaySendsToTheIdentityResolvedTarget 与 TestResolveStepTargetPrefersIdentity 必红。

// replayIdentityModel 造一份"带身份"的读侧夹具：同一个 (colName, fieldId) 有两帧、
// 两个不同的裸 id —— 正是 H1/H2 的情形（裸 id 会被回收复用，身份不会变）。
//
//	窗口 A(100, "w_axmt500")
//	  FormField 6000 b_target#5   ← 先声明的那一帧（录制时写下的 idRef 指向它）
//	  FormField 6008 b_target#5   ← 后声明的那一帧（C3.5：取最后声明的那一份）
//	  FormField 6002 b_amb#6 / 6004 b_amb#7   ← 真歧义（A7）
//	  TableColumn 6006 b_col#9                ← 只接受不了点击（A11）
func replayIdentityModel(t *testing.T) *Model {
	t.Helper()
	m := NewModel()
	m.ApplyMessage(windowMsg(testWinAID, testWinAName, testWinATitle, "demo_t",
		formField(6000, "b_target", 5, "目标字段"),
		formField(6002, "b_amb", 6, "歧义甲"),
		formField(6004, "b_amb", 7, "歧义乙"),
		node("TableColumn 6006", [][2]string{
			{"name", "demo_t.b_col"}, {"colName", "b_col"}, {"fieldId", "9"}, {"text", "明细列"},
		}),
	))
	// 第二帧单独下：它必须是"后声明"，才轮得到 C3.5 的取舍规则。
	m.ApplyMessage(windowMsg(testWinAID, testWinAName, testWinATitle, "demo_t",
		formField(6008, "b_target", 5, "目标字段（重建后的一帧）"),
	))
	if m.Len() == 0 {
		t.Fatal("合成夹具没进模型")
	}
	for _, id := range []int{6000, 6002, 6004, 6006, 6008} {
		if _, ok := m.Get(id); !ok {
			t.Fatalf("夹具缺对象 #%d", id)
		}
	}
	return m
}

// identityStep 造一步"带身份"的回放步骤（事件体照真语料：值 + idRef 同一条）。
func identityStep(col string, fieldID int) ReplayEvent {
	return ReplayEvent{N: 1, Event: `ConfigureEvent 0{{idRef "6000"}{ value "X"}}`,
		ColName: col, FieldID: fieldID}
}

// TestPlanRecordsIdentityFromModel：计划阶段（= 录制侧）要把身份写进步骤。
func TestPlanRecordsIdentityFromModel(t *testing.T) {
	m := replayIdentityModel(t)
	plan := PlanReplay([]string{`ConfigureEvent 0{{idRef "6000"}{ value "X"}}`}, ReplayOptions{Model: m})
	if len(plan.Events) != 1 {
		t.Fatalf("计划条数 = %d，想要 1", len(plan.Events))
	}
	ev := plan.Events[0]
	if ev.ColName != "b_target" || ev.FieldID != 5 {
		t.Fatalf("步骤身份 = (%q, %d)，想要 (b_target, 5) —— 录制时必须从目标节点取 colName/fieldId：%+v",
			ev.ColName, ev.FieldID, ev)
	}
	// 没有身份的（这里给一个模型里没有的裸 id）留空即可，绝不是错误
	lonely := PlanReplay([]string{`ActionEvent 0{{idRef "999999"}}`}, ReplayOptions{Model: m})
	if lonely.Events[0].ColName != "" {
		t.Fatalf("模型里没有的对象不该编出身份：%+v", lonely.Events[0])
	}
}

// TestPlanStageModelNeverDropsEvents 是**反证**：给了模型也只能"补记身份"，
// 不能拿它做取舍 —— 那会把"后面才出现的按钮"提前剔掉（replay.go 的既有结论）。
func TestPlanStageModelNeverDropsEvents(t *testing.T) {
	m := replayIdentityModel(t)
	events := []string{
		`ActionEvent 0{{idRef "999999"}}`, // 模型里没有
		`KeyEvent 0{{keyName "Tab"}}`,     // 没有 idRef
		`ConfigureEvent 0{{idRef "6000"}{ value "X"}}`,
	}
	withModel := PlanReplay(events, ReplayOptions{Model: m})
	without := PlanReplay(events, ReplayOptions{})
	if len(withModel.Events) != len(without.Events) {
		t.Fatalf("给了模型之后计划条数变了：%d → %d —— 模型只许补记身份，不许参与取舍",
			len(without.Events), len(withModel.Events))
	}
	for i := range withModel.Events {
		if withModel.Events[i].Event != without.Events[i].Event {
			t.Fatalf("第 %d 条事件被改了：%q → %q", i+1, without.Events[i].Event, withModel.Events[i].Event)
		}
	}
}

// TestResolveStepTargetPrefersIdentity：带身份 → 现解析出来的 id 说了算（H1/H2）。
func TestResolveStepTargetPrefersIdentity(t *testing.T) {
	m := replayIdentityModel(t)
	tgt, err := ResolveStepTarget(m, identityStep("b_target", 5), Scope{})
	if err != nil {
		t.Fatalf("身份定位本该命中，却被拒：%v", err)
	}
	if !tgt.FromIdentity {
		t.Fatalf("没有走身份解析：%+v", tgt)
	}
	// C3.5：同一个 (colName, fieldId) 有多帧 ⇒ 取**最后声明**那一份
	if tgt.ID != 6008 {
		t.Fatalf("解析出的 id = %d，想要 6008（最后声明的那一帧）", tgt.ID)
	}
	if !strings.Contains(tgt.Event, `idRef "6008"`) {
		t.Fatalf("事件体没换成解析到的 id：%q（还是旧 id 的话，H1 就白写了）", tgt.Event)
	}
	if strings.Contains(tgt.Event, `idRef "6000"`) {
		t.Fatalf("事件体里还留着录制时的旧 id：%q", tgt.Event)
	}
	// 值的部分必须原样保留：只换目标，不动载荷
	if !strings.Contains(tgt.Event, `value "X"`) {
		t.Fatalf("校验载荷被改坏了：%q", tgt.Event)
	}
}

// TestResolveStepTargetKeepsLegacyBehaviour：没有 `ColName` 的步骤**一字不变**。
func TestResolveStepTargetKeepsLegacyBehaviour(t *testing.T) {
	m := replayIdentityModel(t)
	raw := `ActionEvent 0{{idRef "6000"}}`
	tgt, err := ResolveStepTarget(m, ReplayEvent{N: 1, Event: raw}, Scope{})
	if err != nil {
		t.Fatalf("旧步骤（无身份）不该被拒：%v", err)
	}
	if tgt.FromIdentity {
		t.Fatalf("旧步骤不该走身份解析：%+v", tgt)
	}
	if tgt.Event != raw {
		t.Fatalf("旧步骤的事件体被改了：%q → %q", raw, tgt.Event)
	}
	if tgt.ID != 6000 {
		t.Fatalf("旧步骤的目标 = %d，想要 6000（照旧读 idRef）", tgt.ID)
	}
}

// TestResolveStepTargetRefusals：解析失败必须按四码拒绝，且**不改事件体**
// （纯函数层就能证明"不发包"：调用方拿到 error 就不会调 Send）。
func TestResolveStepTargetRefusals(t *testing.T) {
	m := replayIdentityModel(t)
	cases := []struct {
		name string
		ev   ReplayEvent
		want string
		exit int
	}{
		{"作用域里没这个 colName", identityStep("b_nope", 5), CodeFieldNotInScope, ExitFieldNotInScope},
		{"colName 命中但 fieldId 不符（C3.4 另一张表单）", identityStep("b_target", 9), CodeFieldNotInScope, ExitFieldNotInScope},
		{"同一 colName 两种 fieldId", identityStep("b_amb", 0), CodeFieldAmbiguous, ExitFieldAmbiguous},
		{"身份在明细列上，点击发不了", ReplayEvent{N: 1, Event: `ActionEvent 0{{idRef "6006"}}`, ColName: "b_col"},
			CodeFieldNotActionable, ExitFieldNotActionable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tgt, err := ResolveStepTarget(m, c.ev, Scope{})
			if err == nil {
				t.Fatalf("本该被拒（%s），却命中了 #%d", c.want, tgt.ID)
			}
			var ref *Refusal
			if !errors.As(err, &ref) {
				t.Fatalf("失败必须是 *Refusal（C4.1），得到 %T: %v", err, err)
			}
			if ref.Code != c.want || ref.Exit != c.exit {
				t.Fatalf("码/退出码 = %s/%d，想要 %s/%d（%s）", ref.Code, ref.Exit, c.want, c.exit, ref.Message)
			}
			if ref.Message == "" || ref.Hint == "" {
				t.Fatalf("拒绝必须带 Message 与可操作的 Hint（C4.3）：%+v", ref)
			}
			if tgt.Event != c.ev.Event || tgt.FromIdentity {
				t.Fatalf("被拒时不许改事件体、也不许自称解析成功：%+v", tgt)
			}
		})
	}

	// 没有模型（这一屏还取不到）→ FORM_UNKNOWN（1）
	tgt, err := ResolveStepTarget(nil, identityStep("b_target", 5), Scope{})
	var ref *Refusal
	if !errors.As(err, &ref) || ref.Code != CodeFormUnknown || ref.Exit != ExitFormUnknown {
		t.Fatalf("没有模型时该给 FORM_UNKNOWN/1，得到 %v", err)
	}
	if tgt.FromIdentity {
		t.Fatalf("没有模型却说自己解析成功：%+v", tgt)
	}
}

// TestOldReplayPlanJSONStillReads：**旧计划**（没有 colName/fieldId 两键）必须照旧读、
// 照旧跑（零值 = 未记录，不是错误）。
func TestOldReplayPlanJSONStillReads(t *testing.T) {
	const oldPlan = `{"events":[{"n":1,"event":"ActionEvent 0{{idRef \"6000\"}}"},
		{"n":2,"event":"KeyEvent 0{{keyName \"Tab\"}}"}],"skips":[{"raw":"meta","reason":"白名单外"}],"scanned":3}`
	var plan ReplayPlan
	if err := json.Unmarshal([]byte(oldPlan), &plan); err != nil {
		t.Fatalf("旧计划读不动了：%v", err)
	}
	if len(plan.Events) != 2 || len(plan.Skips) != 1 || plan.Scanned != 3 {
		t.Fatalf("旧计划的规模变了：%+v", plan)
	}
	if plan.Events[0].ColName != "" || plan.Events[0].FieldID != 0 {
		t.Fatalf("旧计划不该凭空有身份：%+v", plan.Events[0])
	}
	m := replayIdentityModel(t)
	tgt, err := ResolveStepTarget(m, plan.Events[0], Scope{})
	if err != nil {
		t.Fatalf("旧计划里的一步被拒了：%v", err)
	}
	if tgt.Event != plan.Events[0].Event {
		t.Fatalf("旧计划的事件体被改了：%q → %q", plan.Events[0].Event, tgt.Event)
	}
	// 新计划写出去时，没记身份的那些步不该冒出两个空键（旧读者看得懂，但没必要多两个键）
	b, err := json.Marshal(ReplayEvent{N: 1, Event: `KeyEvent 0{{keyName "Tab"}}`})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "colName") || strings.Contains(string(b), "fieldId") {
		t.Fatalf("没记身份的步骤不该写出空键：%s", b)
	}
}

// TestReplayLogCarriesIdentity：回放日志（= 回放这一轮的录制）要带身份；
// 而且**旧日志**（没有这两个键）照样读得动、进度照算。
func TestReplayLogCarriesIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fep-x-20260101-000000"+ReplayLogSuffix)
	w, err := NewReplayLogWriter(path, "fep-x", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	w.Step(ReplayStep{N: 1, Total: 2, Event: `ConfigureEvent 0{{idRef "6008"}{ value "X"}}`,
		Kind: "ConfigureEvent", Target: 6008, Result: ReplaySent,
		ColName: "b_target", FieldID: 5})
	w.Step(ReplayStep{N: 2, Total: 2, Event: `ActionEvent 0{{idRef "6006"}}`, Kind: "ActionEvent",
		Target: 6006, Result: ReplayRefused, Code: CodeFieldNotActionable, Reason: "明细列不接受点击"})
	if err := w.Close(ReplayResult{Sent: 1, Refused: 1}, "test"); err != nil {
		t.Fatal(err)
	}
	log, err := ReadReplayLog(path)
	if err != nil {
		t.Fatalf("自己写的日志读不动：%v", err)
	}
	steps := log.States()
	if len(steps) != 2 {
		t.Fatalf("步数 = %d，想要 2", len(steps))
	}
	if steps[0].ColName != "b_target" || steps[0].FieldID != 5 {
		t.Fatalf("日志里的身份丢了：%+v", steps[0])
	}
	if steps[1].Result != ReplayRefused || steps[1].Code != CodeFieldNotActionable {
		t.Fatalf("拒绝这一步要看得出码：%+v", steps[1])
	}
	if p := log.Progress(); p.Refused != 1 || p.Done != 2 || p.Sent != 1 {
		t.Fatalf("进度汇总 = %+v", p)
	}

	// 旧日志：一行一步、没有身份、没有 refused —— 照样读、照样算
	old := filepath.Join(dir, "fep-y-20260101-000000"+ReplayLogSuffix)
	lines := `{"v":1,"id":"fep-y","conn":1,"started":1,"total":1}
{"n":1,"total":1,"event":"ActionEvent 0{{idRef \"5160\"}}","kind":"ActionEvent","target":5160,"sentAt":null,"respAt":null,"result":"sent"}
{"finished":2,"sent":1,"skipped":0,"failed":0,"total":1}
`
	if err := os.WriteFile(old, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	olog, err := ReadReplayLog(old)
	if err != nil {
		t.Fatalf("旧日志读不动了：%v", err)
	}
	if olog.Bad != 0 || len(olog.Steps) != 1 {
		t.Fatalf("旧日志解析 = bad:%d steps:%d", olog.Bad, len(olog.Steps))
	}
	if olog.Steps[0].ColName != "" || olog.Steps[0].FieldID != 0 {
		t.Fatalf("旧日志不该凭空有身份：%+v", olog.Steps[0])
	}
	if p := olog.Progress(); p.Sent != 1 || p.Done != 1 || p.Refused != 0 {
		t.Fatalf("旧日志的进度 = %+v", p)
	}
}

// TestReplayStopsWithoutSendingWhenIdentityFails：解析不过 → 拒绝 + **一个字节都不发**。
func TestReplayStopsWithoutSendingWhenIdentityFails(t *testing.T) {
	c, _, rt, _, _ := sessionFixture(t)
	plan := ReplayPlan{Events: []ReplayEvent{identityStep("b_nope", 5)}}
	var seen []ReplayStep
	res, err := c.RunReplayWith(plan, ReplayRunOptions{
		Settle: 10 * time.Millisecond,
		OnStep: func(s ReplayStep) { seen = append(seen, s) },
	})
	var ref *Refusal
	if !errors.As(err, &ref) {
		t.Fatalf("身份解析失败必须以 *Refusal 收场，得到 %v", err)
	}
	if ref.Code != CodeFieldNotInScope {
		t.Fatalf("码 = %s，想要 %s", ref.Code, CodeFieldNotInScope)
	}
	if res == nil || res.Sent != 0 || res.Refused != 1 {
		t.Fatalf("结果 = %+v，想要 sent=0 refused=1", res)
	}
	if len(seen) == 0 || seen[len(seen)-1].Result != ReplayRefused || seen[len(seen)-1].Code != CodeFieldNotInScope {
		t.Fatalf("这一步没能如实落成 refused + 码：%+v", seen)
	}
	// 真的没发包：运行期那条连接上不该有任何东西
	buf := make([]byte, 4096)
	rt.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	if n, rerr := rt.Read(buf); n > 0 {
		t.Fatalf("身份解析失败却发了 %d 字节：%q", n, string(buf[:n]))
	} else if ne, ok := rerr.(net.Error); !ok || !ne.Timeout() {
		t.Fatalf("读运行期连接出错：%v", rerr)
	}
}

// TestReplaySendsToTheIdentityResolvedTarget 是"优先用身份"的**端到端反证**：
// 录制时那一帧的 id 是 6000，重建后同一身份（colName=b_target#5）落到 6008 ——
// 发出去的必须是 6008。改回"仍用旧 id ⇒ 6000"，这条必红。
func TestReplaySendsToTheIdentityResolvedTarget(t *testing.T) {
	c, _, rt, _, dir := sessionFixture(t)
	// 再喂两帧：先 6000，后 6008（同一 colName+fieldId）—— 合成读侧数据，不连真机
	for _, aui := range []string{
		`om 2 {{an 0 Window 4097 {{name "w_axmt500"}} {{` +
			`{an 4097 FormField 6000 {{name "formonly.b_target"}{colName "b_target"}{fieldId "5"}{text "目标字段"}} {}} ` +
			`}}}}`,
		`om 3 {{an 0 Window 4097 {{name "w_axmt500"}} {{` +
			`{an 4097 FormField 6008 {{name "formonly.b_target"}{colName "b_target"}{fieldId "5"}{text "目标字段（重建后）"}} {}} ` +
			`}}}}`,
	} {
		if _, err := rt.Write(Encode(Frame{Type: FrameTypeAUI, Payload: []byte(aui)})); err != nil {
			t.Fatalf("喂 AUI: %v", err)
		}
	}
	m := waitModelHas(t, dir, "t", 6008)
	plan := PlanReplay([]string{`ConfigureEvent 0{{idRef "6000"}{ value "X"}}`}, ReplayOptions{Model: m})
	if plan.Events[0].ColName != "b_target" {
		t.Fatalf("录制阶段没把身份写进步骤：%+v", plan.Events[0])
	}
	res, err := c.RunReplay(plan, 0, 10*time.Millisecond, false, nil)
	if err != nil {
		t.Fatalf("RunReplay: %v", err)
	}
	if res.Sent != 1 {
		t.Fatalf("结果 = %+v", res)
	}
	expectOnConn(t, rt, `idRef "6008"`)
}

// ── A23（契约 §7，用户决策 D-F1）：**实际发包**的回放强制身份 ────────────────
//
// 口径写死在这里，落在 RequireStepIdentity / RequirePlanIdentity / RunReplayIdentity：
//
//	步骤要发给某个对象（事件里有 idRef）却没有 `colName`  → 拒（明确报文 + "需重录"）
//	步骤有 `colName`                                → 走 ByColName 现解析（四码）
//	不发包的路径（DryRun / 纯计划展示）               → 放行，但每步标 NoIdentityNote
//
// A23 原文说"步骤没有 ColName → 拒"；**没有目标的步骤**（KeyEvent）不在此列 ——
// 它没有对象可指，重录也修不好（真录制里每条都有 Tab/Enter），拒了等于"永远不能回放"。
// 这一条由 TestRequirePlanIdentityRefusesStepsWithoutIdentity 的最后一段钉住。

// TestRequirePlanIdentityRefusesStepsWithoutIdentity：计划级判据（**发包之前**的闸门）。
func TestRequirePlanIdentityRefusesStepsWithoutIdentity(t *testing.T) {
	plan := ReplayPlan{Events: []ReplayEvent{
		{N: 1, Event: `ActionEvent 0{{idRef "5160"}}`, ColName: "b_target", FieldID: 5},
		{N: 2, Event: `ConfigureEvent 0{{idRef "6000"}{ value "X"}}`},
	}}
	err := RequirePlanIdentity(plan)
	var ref *Refusal
	if !errors.As(err, &ref) {
		t.Fatalf("无身份的一步必须被拒（A23），得到 %v", err)
	}
	if ref.Code != CodeFieldNotInScope || ref.Exit != ExitFieldNotInScope {
		t.Fatalf("码/退出码 = %s/%d，想要 %s/%d（契约 §7 A23 说按四码拒）",
			ref.Code, ref.Exit, CodeFieldNotInScope, ExitFieldNotInScope)
	}
	if !strings.Contains(ref.Message, "需重录") {
		t.Fatalf("报文必须写清后果（A23 原文：需重录）：%q", ref.Message)
	}
	if !strings.Contains(ref.Message, "第 2 步") {
		t.Fatalf("报文要点名是哪一步：%q", ref.Message)
	}
	if !strings.Contains(ref.Hint, "mark-stale") {
		t.Fatalf("Hint 要给旧件的那两条路之一（A24：重录 / mark-stale）：%q", ref.Hint)
	}
	if got := NoIdentitySteps(plan); len(got) != 1 || got[0] != 2 {
		t.Fatalf("无身份的步骤 = %v，想要 [2]", got)
	}

	// 每一步都带身份 ⇒ 放行
	full := ReplayPlan{Events: []ReplayEvent{
		{N: 1, Event: `ActionEvent 0{{idRef "5160"}}`, ColName: "b_target", FieldID: 5},
	}}
	if err := RequirePlanIdentity(full); err != nil {
		t.Fatalf("每一步都有身份却仍被拒：%v", err)
	}
	if got := NoIdentitySteps(full); len(got) != 0 {
		t.Fatalf("有身份的计划不该被标注：%v", got)
	}
	// 空计划：没有要发的东西 ⇒ 放行
	if err := RequirePlanIdentity(ReplayPlan{}); err != nil {
		t.Fatalf("空计划不该被拒：%v", err)
	}

	// 没有目标的步骤（KeyEvent）不在"A23 强制身份"的范围内：
	// 它没有对象可指，装不上身份，重录也一样 —— 拒了它等于每份真录制都不能回放
	// （实测语料里每条录制都有 Tab/Enter）。
	key := ReplayPlan{Events: []ReplayEvent{{N: 1, Event: `KeyEvent 0{{keyName "Tab"}}`}}}
	if err := RequirePlanIdentity(key); err != nil {
		t.Fatalf("无目标的按键步骤不该被拒（它没有对象可指）：%v", err)
	}
	if got := NoIdentitySteps(key); len(got) != 0 {
		t.Fatalf("没有目标的步骤也不该进「无身份」标注：%v", got)
	}
}

// TestRunReplayIdentityRefusesBeforeAnyByte 是 A23 的**硬要求**：
// 拒绝必须发生在**任何发包之前** —— 连"有身份、本该能发"的第 1 步也不许先发出去。
func TestRunReplayIdentityRefusesBeforeAnyByte(t *testing.T) {
	c, _, rt, _, _ := sessionFixture(t)
	plan := ReplayPlan{Events: []ReplayEvent{
		{N: 1, Event: `ActionEvent 0{{idRef "5160"}}`, ColName: "b_target", FieldID: 5},
		{N: 2, Event: `ConfigureEvent 0{{idRef "6000"}{ value "X"}}`},
	}}
	res, err := c.RunReplayIdentity(plan, ReplayRunOptions{Settle: 10 * time.Millisecond})
	if res != nil {
		t.Fatalf("被拒时不许给出「跑完了」的结果：%+v", res)
	}
	var ref *Refusal
	if !errors.As(err, &ref) || !strings.Contains(ref.Message, "需重录") {
		t.Fatalf("无身份的一步必须以「需重录」的拒绝收场，得到 %v", err)
	}
	// 运行期那条连接上**一个字节都没有**（连第 1 步都没发）
	buf := make([]byte, 4096)
	rt.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	if n, rerr := rt.Read(buf); n > 0 {
		t.Fatalf("拒绝之前发了 %d 字节：%q", n, string(buf[:n]))
	} else if ne, ok := rerr.(net.Error); !ok || !ne.Timeout() {
		t.Fatalf("读运行期连接出错：%v", rerr)
	}
}

// TestRunReplayIdentityAllowsDryRunWithoutIdentity：**例外**（A23 第二段）——
// 不发包的路径放行，而且照样一个字节都不发。
func TestRunReplayIdentityAllowsDryRunWithoutIdentity(t *testing.T) {
	c, _, rt, _, _ := sessionFixture(t)
	plan := ReplayPlan{Events: []ReplayEvent{
		{N: 1, Event: `ConfigureEvent 0{{idRef "6000"}{ value "X"}}`},
	}}
	res, err := c.RunReplayIdentity(plan, ReplayRunOptions{DryRun: true, Settle: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("dry-run 只是展示，不该被拒（A23 例外）：%v", err)
	}
	if res == nil || res.Sent != 0 || res.Skipped != 1 {
		t.Fatalf("dry-run 结果 = %+v，想要 sent=0 skipped=1", res)
	}
	buf := make([]byte, 4096)
	rt.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	if n, rerr := rt.Read(buf); n > 0 {
		t.Fatalf("dry-run 发了 %d 字节：%q", n, string(buf[:n]))
	} else if ne, ok := rerr.(net.Error); !ok || !ne.Timeout() {
		t.Fatalf("读运行期连接出错：%v", rerr)
	}
}

// waitModelHas 等录制件里的某一帧进模型（代理是另一个 goroutine 在写文件）。
func waitModelHas(t *testing.T, dir, tag string, id int) *Model {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if m, _, _, err := LoadModelFrom(dir, tag); err == nil {
			if _, ok := m.Get(id); ok {
				return m
			}
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatalf("模型里始终没有 #%d", id)
	return nil
}

// TestConfigureEventNeedsIdentityOnlyWhenWritingValue：`ConfigureEvent` 的身份判据按**是否写值**分。
//
// 网格的 bufferSize/pageSize、栏位的 tabIndex、光标位置都是运行期**定位与排版**，
// 这类对象天然没有 colName；拿"点控件"的纪律去要求它们，结果是每份真录制都回放不了
// （真机现场：一份刚录的 axmt500_wf 在第 14 步被拒，那一行是网格的 pageSize/bufferSize）。
func TestConfigureEventNeedsIdentityOnlyWhenWritingValue(t *testing.T) {
	allowed := []string{
		`ConfigureEvent 0{{idRef "766"}{ pageSize "9"}{ bufferSize "10"}}`,
		`ConfigureEvent 0{{idRef "1065"}{ tabIndex "1"}}`,
		`ConfigureEvent 0{{idRef "1034"}{ cursor "0"}}`,
	}
	for _, ev := range allowed {
		if err := RequireStepIdentity(ReplayEvent{N: 1, Event: ev}); err != nil {
			t.Errorf("定位/排版类 ConfigureEvent 不该被拒：%s → %v", ev, err)
		}
	}
	// 写值仍要身份：不知道写到哪个字段，发出去就是"改错字段"。
	write := ReplayEvent{N: 2, Event: `ConfigureEvent 0{{idRef "6000"}{ value "X"}}`}
	var ref *Refusal
	if err := RequireStepIdentity(write); err == nil {
		t.Fatal("写值的 ConfigureEvent 没有 colName 时必须拒（A23）")
	} else if !errors.As(err, &ref) || ref.Code != CodeFieldNotInScope {
		t.Fatalf("该报 %s，实得 %v", CodeFieldNotInScope, err)
	}
}

// TestKindForReplayEvent：事件 → `ByColName` 的动作类型。
//
// 这条判据决定"要不要挑一个能接受该动作的节点"。`ConfigureEvent` 里只有**写值**是输入；
// 定位/排版（tabIndex/bufferSize/pageSize）不限类型 —— 否则落在网格列上的那几条会被
// `FIELD_NOT_ACTIONABLE` 一并拒掉（真机现场：回放卡在第 1 步，imaa001@TableColumn）。
func TestKindForReplayEvent(t *testing.T) {
	cases := []struct{ ev, want string }{
		{`ActionEvent 0{{idRef "1"}}`, "click"},
		{`ConfigureEvent 0{{idRef "1"}{ value "X"}}`, "input"},
		{`ConfigureEvent 0{{idRef "766"}{ pageSize "9"}{ bufferSize "10"}}`, ""},
		{`ConfigureEvent 0{{idRef "1065"}{ tabIndex "1"}}`, ""},
		{`KeyEvent 0{{keyName "Tab"}}`, ""},
	}
	for _, c := range cases {
		if got := kindForReplayEvent(c.ev); got != c.want {
			t.Errorf("kindForReplayEvent(%s) = %q，要 %q", c.ev, got, c.want)
		}
	}
}
