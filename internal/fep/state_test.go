package fep

import (
	"strings"
	"testing"
)

// 可用性判定（Eligibility）：**只认运行期广播的真属性**，且判据不足不拦。
//
// 依据：属性词表从 94 个真实连接挖出（tools/gdc-uia/README.md）——
// active 是运行期发的**最频繁**的属性（12.7 万次），另有 hidden / noEntry /
// actionActive / dialogType。这些正是"什么时候哪些控件可用"的判据，
// 以前没用它们，所以驱动老是在不可用的东西上发事件（空 ACK，严重时打停作业）。

func TestEligibilityFollowsRuntimeAttributes(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	m.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}} {{Action 201 {{name "accept"}}}}}}`)
	m.ApplyMessage(`om 2 {{un 201 {{actionActive "1"}}}}`)
	m.ApplyMessage(`om 3 {{an 178 Value 300 {{name "cmt_pmdl004"} {active "1"} {value ""}}}}`)
	m.ApplyMessage(`om 4 {{an 178 Value 301 {{name "cmt_pmdl005"} {active "0"} {value ""}}}}`)
	m.ApplyMessage(`om 5 {{an 178 Value 302 {{name "cmt_pmdl006"} {active "1"} {hidden "1"}}}}`)
	m.ApplyMessage(`om 6 {{an 178 Value 303 {{name "cmt_pmdl007"} {active "1"} {noEntry "1"}}}}`)
	m.ApplyMessage(`om 7 {{an 178 ToolBarItem 304 {{name "insert"} {actionActive "0"}}}}`)
	m.ApplyMessage(`om 8 {{an 178 ToolBarItem 305 {{name "modify"} {actionActive "1"}}}}`)

	get := func(id int) *Object { o, _ := m.Get(id); return o }

	for _, tc := range []struct {
		id     int
		want   bool
		by     string
		reason string
	}{
		{300, true, "runtime", ""},
		{301, false, "active", "不可用"},
		{302, false, "hidden", "隐藏"},
		{303, false, "noEntry", "不可录入"},
		{201, true, "runtime", ""},
		{304, false, "actionActive", "不可用"},
		{305, true, "runtime", ""},
	} {
		got := Eligible(m, get(tc.id))
		if got.Usable != tc.want {
			t.Errorf("#%d：可用性 = %v，想要 %v（%s）", tc.id, got.Usable, tc.want, got.Reason)
			continue
		}
		if got.By != tc.by {
			t.Errorf("#%d：依据 = %q，想要 %q", tc.id, got.By, tc.by)
		}
		if tc.reason != "" && !strings.Contains(got.Reason, tc.reason) {
			t.Errorf("#%d：原因 %q 没说明白（想要含 %q）", tc.id, got.Reason, tc.reason)
		}
	}
}

// TestEligibilityDoesNotBlockWithoutEvidence：判据不足**不拦**。
//
// 与动作层防线同一条原则：拦错了会让本来能做的事做不成，而漏判最多多一次
// 被运行期拒绝的发送（有回显可查）。这条测试防的是"以后有人把默认值改成拦"。
func TestEligibilityDoesNotBlockWithoutEvidence(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	m.ApplyMessage(`om 1 {{an 178 Value 400 {{name "cmt_xxx"}}}}`) // 一个属性都没有
	if got := Eligible(m, mustGet(t, m, 400)); !got.Usable {
		t.Errorf("没有任何属性时不该拦：%s", got)
	}
}

// TestEligibilityInRequiresCurrentDialog：动作必须属于**当前那一个** Dialog 实例。
func TestEligibilityInRequiresCurrentDialog(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	m.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}} {{Action 201 {{name "accept"}}}}}}`)
	m.ApplyMessage(`om 2 {{an 178 Dialog 300 {{active "0"}} {{Action 301 {{name "accept"}}}}}}`)

	if got := EligibleIn(m, mustGet(t, m, 201), 200); !got.Usable {
		t.Errorf("当前指令里的动作应可用：%s", got)
	}
	got := EligibleIn(m, mustGet(t, m, 301), 200)
	if got.Usable {
		t.Fatal("别的指令里的同名动作被放行了（发出去只会回空 ACK）")
	}
	if got.By != "dialog" {
		t.Errorf("依据 = %q，想要 dialog", got.By)
	}
}

// TestEligibilityOnRealCorpus：真语料上，**用户真按过的动作**必须判为可用。
//
// 这是"模型自证"的最小版本（用户要的根：不再走一步卡一步）：
// 出现反例 = 模型错（不是运行期错），要修模型。
func TestEligibilityOnRealCorpus(t *testing.T) {
	for _, name := range fixtureFiles(t) {
		m := foldFixture(t, name)
		if m == nil {
			continue
		}
		refused := 0
		for _, o := range m.Objects() {
			if !isActionLike(o.Type) {
				continue
			}
			if o.Attr("actionActive") == "" {
				continue // 运行期没广播 → 判据不足，不算反例
			}
			if o.Attr("actionActive") == "0" {
				refused++
				continue
			}
			if got := Eligible(m, o); !got.Usable {
				t.Errorf("%s：#%d(%s) 运行期标了 actionActive=%q 却判成不可用：%s",
					name, o.ID, o.Attr("name"), o.Attr("actionActive"), got.Reason)
			}
		}
		t.Logf("%s：动作 %d 个，其中运行期标为不可用 %d 个", name, len(m.Objects()), refused)
	}
}

func mustGet(t *testing.T, m *Model, id int) *Object {
	t.Helper()
	o, ok := m.Get(id)
	if !ok {
		t.Fatalf("模型里没有 #%d", id)
	}
	return o
}

// TestDialogStatesReportsWindows：对话框状态报告必须能问出"现在这个窗口有几个可用动作"。
func TestDialogStatesReportsWindows(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	m.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}} {{Action 201 {{name "accept"}}}}}}`)
	m.ApplyMessage(`om 2 {{an 178 ToolBarItem 305 {{name "modify"} {actionActive "1"}}}}`)
	m.ApplyMessage(`om 3 {{an 178 ToolBarItem 306 {{name "delete"} {actionActive "0"}}}}`)

	sts := m.DialogStates()
	if len(sts) == 0 {
		t.Fatal("没有窗口")
	}
	s := sts[0]
	if s.WindowID != 178 {
		t.Errorf("窗口 = %d，想要 178", s.WindowID)
	}
	joined := strings.Join(s.Actions, ",")
	if !strings.Contains(joined, "modify") || !strings.Contains(joined, "accept") {
		t.Errorf("可用动作少了：%v", s.Actions)
	}
	if strings.Contains(joined, "delete") {
		t.Errorf("actionActive=0 的动作不该出现在可用列表里：%v", s.Actions)
	}
}

// TestFocusIsReadFromRuntime：焦点是运行期推的（`un 0 {{focus "N"}}`），必须能读出来。
//
// 为什么这是地基（用户点出来的要害）：运行期只认"光标确实在这个栏位"时的动作。
// 焦点不在那儿，动作名再对、actionActive 再是 1，点下去也是空动作。以前驱动从不查它。
func TestFocusIsReadFromRuntime(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	m.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}} {{FormField 260 {{name "采购单号"}}}}}}`)
	m.ApplyMessage(`om 2 {{un 0 {{focus "260"}}}}`)
	m.ApplyMessage(`om 3 {{an 178 TableColumn 650 {{name "pmdn_t.pmdn001"}}}}`)

	id, ok := m.Focus()
	if !ok || id != 260 {
		t.Fatalf("Focus() = %d/%v，想要 260", id, ok)
	}
	rep := m.FocusNow()
	if rep.FocusID != 260 || rep.FocusKind != "FormField" {
		t.Errorf("FocusNow 焦点 = #%d/%s，想要 #260/FormField", rep.FocusID, rep.FocusKind)
	}
	if on, known := m.FocusOnField(260); !on || !known {
		t.Errorf("焦点就在 #260 上，却判成 on=%v known=%v", on, known)
	}
	if on, known := m.FocusOnField(650); on || !known {
		t.Errorf("焦点不在 #650 上，却判成 on=%v known=%v（焦点判据已知，不该说不知道）", on, known)
	}
}

// TestFocusUnknownIsNotABlock：运行期没推焦点时，判"不知道"，不能假装知道。
func TestFocusUnknownIsNotABlock(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	if _, ok := m.Focus(); ok {
		t.Fatal("没推焦点时不该说有焦点")
	}
	if _, known := m.FocusOnField(650); known {
		t.Fatal("没推焦点时不该说\"知道焦点不在那儿\"")
	}
}

// TestMenuActionIsClickable：**等确认**的窗口（错误窗/提示窗/菜单）靠 MenuAction 按掉。
//
// 实测依据：错误窗（cl_err_s01）弹出来时运行期把焦点指到 #690（MenuAction "OK"）——
// 那就是"运行期在等你去按它"。旧白名单漏了 MenuAction，于是只能靠瞎按键去猜（还猜错）。
func TestMenuActionIsClickable(t *testing.T) {
	if !ClickableTypes["MenuAction"] || !ClickableTypes["Menu"] {
		t.Fatal("MenuAction/Menu 不在可点击白名单里（错误窗就只能靠瞎按键了）")
	}
	m := NewModel()
	m.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	m.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}} {{MenuAction 690 {{name "OK"}}}}}}`)
	if err := CheckClick(m, 690); err != nil {
		t.Errorf("错误窗的 OK（MenuAction，属于当前 Dialog）应可点：%v", err)
	}
}
