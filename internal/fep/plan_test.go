package fep

import (
	"strings"
	"testing"
)

// 从**表单定义（.4fd）**推导操作计划。
//
// 目标（用户定调）：步骤只由"输入/选择/点按钮"定义，经过栏位（Tab）不算步骤。
// 所以这里的输入是一张表单的控件清单，输出是一串可执行、可核对、带出处的语义步骤。

// 真实结构片段（照 apmt500_wf.4fd 的元素形态裁剪，只留与本测试有关的属性）。
const formSnippet = `
<Edit name="pmdl_t.pmdl002" colName="pmdl002" sqlTabName="pmdl_t" widget="Edit" comment="cmt_pmdl002" tabIndex="1" posX="40" posY="20" required="true"/>
<ButtonEdit name="pmdl_t.pmdl004" colName="pmdl004" sqlTabName="pmdl_t" widget="ButtonEdit" comment="cmt_pmdl004" tabIndex="2" posX="300" posY="20" image="16/openwindow.png"/>
<CheckBox name="formonly.pmdl012" sqlTabName="pmdl_t" widget="CheckBox" comment="cmt_pmdl012" tabIndex="3" valueChecked="Y" valueUnchecked="N"/>
<ComboBox name="pmdl_t.pmdl005" colName="pmdl005" sqlTabName="pmdl_t" widget="ComboBox" comment="cmt_pmdl005" tabIndex="4"/>
<Edit name="pmdn_t.pmdn001" colName="pmdn001" sqlTabName="pmdn_t" widget="Edit" comment="cmt_pmdn001" tabIndex="5" noEntry="true"/>
<Edit name="pmdn_t.pmdn002" colName="pmdn002" sqlTabName="pmdn_t" widget="Edit" comment="cmt_pmdn002" tabIndex="6" required="true"/>
<Label name="lbl_1" widget="Label" title="单价含税否" tabIndex="7"/>
`

func TestPlanComesFromTheFormDefinition(t *testing.T) {
	fields := ParseForm4fd("apmt500_wf", formSnippet)
	steps := BuildPlan("apmt500_wf", fields)

	// 只应产出"能做的事"：Label 与 noEntry 的栏位不进计划。
	if len(steps) != 5 {
		t.Fatalf("步骤数 = %d，想要 5（Label 与 noEntry 的栏位不该进）：\n%s", len(steps), PlanText(steps))
	}

	want := []struct {
		op, role, value string
	}{
		{"set", "cmt_pmdl002", ""},
		{"openwin", "cmt_pmdl004", ""},
		{"toggle", "cmt_pmdl012", "Y"}, // 勾选字面值来自 .4fd，不靠猜
		{"choose", "cmt_pmdl005", ""},
		{"set", "cmt_pmdn002", ""},
	}
	for i, w := range want {
		if steps[i].Op != w.op || steps[i].Role != w.role {
			t.Errorf("第 %d 步 = %s/%s，想要 %s/%s", i+1, steps[i].Op, steps[i].Role, w.op, w.role)
		}
		if w.value != "" && steps[i].Value != w.value {
			t.Errorf("第 %d 步的字面值 = %q，想要 %q", i+1, steps[i].Value, w.value)
		}
		if steps[i].Source == "" {
			t.Errorf("第 %d 步没有出处（每步都必须能追溯）", i+1)
		}
	}

	// 顺序必须按 tabIndex（运行期 DIALOG 用 FIELD ORDER FORM 时就是这个序）。
	// 这里刻意让 noEntry/单身的栏位夹在中间，验证的是"过滤后仍保持原序"。
	if steps[4].Role != "cmt_pmdn002" {
		t.Errorf("末步 = %s，想要 cmt_pmdn002（单身栏位）", steps[4].Role)
	}
	if steps[0].Table != "pmdl_t" || steps[4].Table != "pmdn_t" {
		t.Errorf("分组（sqlTabName）没带出来：%q / %q", steps[0].Table, steps[4].Table)
	}
}

// TestPlanSaysOpenWinSelectSemanticsFromSource：开窗那一步的"怎么选中"取自开窗程序的源码。
func TestPlanSaysOpenWinSelectSemanticsFromSource(t *testing.T) {
	fields := ParseForm4fd("apmt500_wf", formSnippet)
	steps := BuildPlan("apmt500_wf", fields)

	ow := &OpenWinModel{
		Program: "q_pmaa001_3",
		Dialogs: []OpenWinDialog{{
			Dialog: "DISPLAY ARRAY", Select: SelectCurrentRow,
			Evidence: []string{"q_pmaa001_3.4gl:484 ON ACTION accept"},
		}},
	}
	steps = PlanStepsFromOpenWin(steps, ow)

	var found bool
	for _, s := range steps {
		if s.Op != "openwin" {
			continue
		}
		found = true
		// 开窗程序名要与计划里的 opener 对上才注入语义；这里 opener 为空，
		// 所以只验证"注入函数不会误改别的步骤"。
		if s.Value != "" {
			t.Errorf("opener 对不上时不该注入选中语义，却得到 %q", s.Value)
		}
	}
	if !found {
		t.Fatal("计划里没有开窗步骤")
	}

	// 把 opener 补上 → 应该注入 current-row 并写出出处
	for i := range steps {
		if steps[i].Op == "openwin" {
			steps[i].Opener = "q_pmaa001_3"
		}
	}
	steps = PlanStepsFromOpenWin(steps, ow)
	for _, s := range steps {
		if s.Op != "openwin" {
			continue
		}
		if s.Value != "current-row" {
			t.Errorf("选中语义 = %q，想要 current-row", s.Value)
		}
		if !strings.Contains(s.Why, "q_pmaa001_3.4gl:484") {
			t.Errorf("没有把源码出处写进步骤：%q", s.Why)
		}
	}
}

// TestOpForWidgetIsATypeTable：操作**由控件类型决定**（用户原话：
// "这么多控件的情况永远测不完"）—— 所以这里是"按类型分派"，不是"按控件试探"。
func TestOpForWidgetIsATypeTable(t *testing.T) {
	cases := map[string]string{
		"Edit": "set", "DateEdit": "set", "SpinEdit": "set",
		"ComboBox": "choose", "RadioGroup": "choose",
		"CheckBox":   "toggle",
		"ButtonEdit": "openwin",
	}
	for w, want := range cases {
		got, ok := OpForWidget(w)
		if !ok || got != want {
			t.Errorf("widget=%s → %q/%v，想要 %q", w, got, ok, want)
		}
	}
	for _, w := range []string{"Label", "FFLabel", "Image"} {
		if _, ok := OpForWidget(w); ok {
			t.Errorf("widget=%s 不该有操作（它不是可操作控件）", w)
		}
	}
}
