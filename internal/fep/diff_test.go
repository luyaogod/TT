package fep

import (
	"strings"
	"testing"
)

// 这一步之后"树变了什么" —— 差分。
//
// 用户对基础设施的要求：每次执行命令过后，要能同时回答
// 树的变化 / 焦点在哪 / 哪些是自动带值 / 哪些值变了。
// 本文件测前两件与"我写的 vs 自动带值"的判定依据（值变化本身在 do 的等待循环里收集）。

func TestDiffReportsTreeChanges(t *testing.T) {
	before := NewModel()
	before.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	before.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}} {{ToolBarItem 201 {{name "insert"}}}}}}`)
	before.ApplyMessage(`om 2 {{an 178 TableColumn 650 {{name "pmdn_t.pmdn001"}}}}`)

	after := NewModel()
	after.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	after.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}} {{ToolBarItem 201 {{name "insert"}}}}}}`)
	after.ApplyMessage(`om 2 {{an 178 TableColumn 650 {{name "pmdn_t.pmdn001"}}}}`)
	// 开窗：新窗口 + 它自己的对话框与动作
	after.ApplyMessage(`om 3 {{an 0 Window 586 {{name "w_qry_q_pmaa001_3"}}}}`)
	after.ApplyMessage(`om 4 {{an 586 Dialog 600 {{active "1"}} {{Action 601 {{name "accept"}}}}}}`)
	// 关掉一个旧对象
	removed := NewModel()
	removed.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	removed.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}}}}`)
	removed.ApplyMessage(`om 2 {{an 178 ToolBarItem 999 {{name "delete"}}}}`)
	_ = removed

	before2 := NewModel()
	before2.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	before2.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}}}}`)
	before2.ApplyMessage(`om 2 {{an 178 ToolBarItem 999 {{name "delete"}}}}`)
	after2 := NewModel()
	after2.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	after2.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}}}}`)

	d := DiffModels(before, after)
	if d.AddedCount == 0 {
		t.Fatal("新增了窗口与对话框，却没报出来")
	}
	if !contains(d.OpenedWindows, "w_qry_q_pmaa001_3") {
		t.Errorf("开窗没报出来：%v", d.OpenedWindows)
	}
	if !contains(d.GainedActions, "accept") {
		t.Errorf("新动作没报出来：%v", d.GainedActions)
	}
	if d.Empty() {
		t.Error("有变化却报了 Empty")
	}

	d2 := DiffModels(before2, after2)
	if d2.RemovedCount == 0 {
		t.Fatal("删了对象却没报出来")
	}
	if !contains(d2.LostActions, "delete") {
		t.Errorf("消失的动作没报出来：%v", d2.LostActions)
	}
	if !strings.Contains(d2.Summary(), "消失动作") {
		t.Errorf("摘要没提消失的动作：%s", d2.Summary())
	}
}

func TestDiffEmptyWhenNothingChanges(t *testing.T) {
	build := func() *Model {
		m := NewModel()
		m.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
		m.ApplyMessage(`om 1 {{an 178 Dialog 200 {{active "1"}}}}`)
		return m
	}
	if d := DiffModels(build(), build()); !d.Empty() {
		t.Errorf("同一个模型比两次应该没变化，却报了：%s", d.Summary())
	}
	if got := DiffModels(nil, build()); !got.Empty() {
		t.Error("before 为 nil 时应该什么都不报（判据不足）")
	}
}

// TestDiffSeesUpdatesButKeepsSummaryShort：属性改动只报数量（列出来会淹掉信号）。
func TestDiffSeesUpdatesButKeepsSummaryShort(t *testing.T) {
	before := NewModel()
	before.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	before.ApplyMessage(`om 1 {{an 178 ToolBarItem 201 {{name "insert"} {actionActive "0"}}}}`)
	after := NewModel()
	after.ApplyMessage(`om 0 {{an 0 Window 178 {{name "w_main"}}}}`)
	after.ApplyMessage(`om 1 {{an 178 ToolBarItem 201 {{name "insert"} {actionActive "1"}}}}`)

	d := DiffModels(before, after)
	if d.UpdatedCount == 0 {
		t.Fatal("actionActive 变了（0→1），却没报改动")
	}
	if d.AddedCount != 0 || d.RemovedCount != 0 {
		t.Errorf("只是属性改动，不该报增删：+%d -%d", d.AddedCount, d.RemovedCount)
	}
}

// TestFirstConfigureTargetTellsWhatIWritten：raw 形态（写格子 = 值与光标同一条消息）
// 必须能认出"我写的是哪个对象、什么值"，否则报告里会把"我写的"错当成"自动带值"。
func TestFirstConfigureTargetTellsWhatIWritten(t *testing.T) {
	ev := ConfigureValueText(6587, "Y", nil) + "}{" + ConfigureCursorText(650, 1)
	id, val := firstConfigureTarget(ev)
	if id != 6587 || val != "Y" {
		t.Fatalf("抠出来的目标 = #%d %q，想要 #6587 \"Y\"", id, val)
	}
	// 只关心 idRef 时（没有 value）也要给出 id
	if id2, _ := firstConfigureTarget(ConfigureCursorText(650, 1)); id2 != 650 {
		t.Errorf("只带 cursor 的事件也该认出目标 id，得到 #%d", id2)
	}
	if id3, _ := firstConfigureTarget("ActionEvent 0{{idRef \"9\"}}"); id3 != 0 {
		t.Errorf("ActionEvent 不是 ConfigureEvent，不该被当值写入：#%d", id3)
	}
}

