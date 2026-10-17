package fep

import (
	"fmt"
	"sort"
	"strings"
)

// PlanStep 是一个**语义步骤**：输入、选择、开窗、点动作。
//
// 用户定的规矩：经过栏位（Tab）不算步骤，**"我输入和选择的才是核心"**。
// 所以这里只产出可填入/可选/可按的东西，且每一步都带**出处**（哪个文件哪一行/哪个表单定义），
// 让"说明"和"画面"能互相验证。
type PlanStep struct {
	Seq      int    `json:"seq"`
	Op       string `json:"op"` // set / choose / openwin / toggle / act
	Role     string `json:"role"`
	Label    string `json:"label,omitempty"` // 中文名（.4fd 的 title；空的由字典补）
	Comment  string `json:"comment,omitempty"`
	Widget   string `json:"widget,omitempty"`
	Group    string `json:"group,omitempty"` // 分组：按 .4fd 的 sqlTabName（单头/单身）
	Table    string `json:"table,omitempty"`
	Required bool   `json:"required,omitempty"`
	Opener   string `json:"opener,omitempty"` // 开窗程序名（可选）
	Value    string `json:"value,omitempty"`  // 建议值（来自表单定义，如勾选字面值）
	Why      string `json:"why,omitempty"`
	Source   string `json:"source,omitempty"`

	// ── 身份（契约 C3 + A18；JSON 键名 `colName`/`fieldId`） ──────────────────
	//
	// 计划步骤要能"回放时重新找到这个控件"，所以除了给人看的中文名，还得记下**字段
	// 身份**。两个字段都**向后兼容**：旧计划里没有这两个键 ⇒ 反序列化得零值
	// （空 / 0）= "未记录"，**不是错误**，读侧照旧按名字走。
	//
	// ColName 来自 `.4fd` 的 `colName`（表单内稳定，可跨会话复用）。
	ColName string `json:"colName,omitempty"`
	// FieldID 是 `.4fd`/运行期的编号。**本层永远不填它**：`.4fd` 的 `fieldId` 每次
	// 保存全量重排（契约 B5/H5，engine/SPEC.md:257），拿它当身份就是"看起来对、其实错"。
	// 0 = 未记录；要填只许填**运行期证据**（回放侧 ReplayStep.FieldID 就是这么来的）。
	FieldID int `json:"fieldId,omitempty"`
}

// OpForWidget 把 **.4fd 的控件类型**映射成**可做的操作**。
//
// 这是"控件类型 → 支持的操作"表的源码侧那一半：类型决定能做什么，
// 而不是靠逐控件试探（用户原话："这么多控件的情况永远测不完"）。
//
// 出处：.4fd 的 widget 属性取值（实测 apmt500_wf.4fd：Edit 284 / ComboBox 21 /
// CheckBox 14 / ButtonEdit …），与运行期 AUI 里同名 style 一一对应。
func OpForWidget(widget string) (op string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(widget)) {
	case "edit", "textedit", "lineedit", "spinedit", "dateedit", "datetimeedit":
		return "set", true
	case "combo****":
		return "choose", true
	case "combobox":
		return "choose", true
	case "checkbox":
		return "toggle", true
	case "buttondit", "buttonedit":
		return "openwin", true
	case "radiogroup":
		return "choose", true
	case "image", "label", "fflabel", "statictext":
		return "", false
	}
	return "", false
}

// BuildPlan 从**表单定义**推导出操作计划。
//
// 依据：
//   - 顺序 = TabIndex（.4fd 的 tabIndex；运行期 DIALOG 用 FIELD ORDER FORM 时
//     这个序就是运行期走栏位的序 —— 实测 apmt500_wf.4gl:2171
//     `DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)`）；
//   - 操作 = 控件类型（OpForWidget）；有开窗程序的 ButtonEdit → openwin；
//   - 分组 = sqlTabName（T100 里单头是 *_t、单身明细是另一张表）；
//   - 只收"能做的事"：不可录入（noEntry）与纯展示类型不进计划。
//
// 它刻意**不含业务分支**（5 万行 4gl 里的前置条件不做静态分析）——
// 分支靠运行期状态兜：见 Eligible()/EligibleIn()，以及"模型说可用而运行期拒绝 → 记反例"。
func BuildPlan(form string, fields []SrcField) []PlanStep {
	type item struct {
		f     SrcField
		op    string
		order int
	}
	var items []item
	for _, f := range fields {
		if f.NoEntry {
			continue
		}
		// 只收**业务字段**：T100 里真实栏位都带 comment（cmt_xxx），而表单级元数据
		// （name / appname 这类 NON_DATABASE 项）没有 —— 实测不筛会在计划头上冒出两条噪声。
		if f.Comment == "" && f.FieldType != "TABLE_COLUMN" {
			continue
		}
		op, ok := OpForWidget(f.Widget)
		if !ok {
			continue
		}
		if f.Opener != "" {
			op = "openwin"
		}
		order := f.TabIndex
		if order == 0 {
			order = 100000 + f.PosY*1000 + f.PosX
		}
		items = append(items, item{f: f, op: op, order: order})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].order < items[j].order })

	out := make([]PlanStep, 0, len(items))
	for i, it := range items {
		f := it.f
		role := f.Name
		if f.Comment != "" {
			role = f.Comment // cmt_* 是运行期对象上也带的键 → 驱动用它最稳
		}
		st := PlanStep{
			Seq: i + 1, Op: it.op, Role: role, Label: f.Title, Comment: f.Comment,
			Widget: f.Widget, Group: f.FieldType, Table: f.Table,
			Required: f.Required, Opener: f.Opener,
			Source: form,
			// 身份只带 `.4fd` 里那个**稳定**的键（colName）；fieldId 见 PlanStep 的注释：
			// 从 `.4fd` 取编号是红线上的一条（H5），所以这里**故意**不填。
			ColName: f.ColName,
		}
		switch it.op {
		case "openwin":
			st.Why = "开窗选一项回传（.4fd 带 image=\"16/openwindow.png\"）"
		case "choose":
			st.Why = "从选项里选一项（.4fd widget=" + f.Widget + "）"
		case "toggle":
			st.Why = "勾选/取消勾选（字面值来自 .4fd）"
			st.Value = f.ValueChecked
		default:
			st.Why = "输入值（.4fd widget=" + f.Widget + "）"
		}
		out = append(out, st)
	}
	return out
}

// PlanStepsFromOpenWin 把开窗的语义接进计划：开窗那一步要知道"选哪一行/勾哪一格"。
//
// 出处：开窗程序的 .4gl DIALOG 块（INPUT ARRAY → 勾选；DISPLAY ARRAY → 当前行），
// 见 internal/fep/openwin.go 的 ParseOpenWin4gl。
func PlanStepsFromOpenWin(steps []PlanStep, ow *OpenWinModel) []PlanStep {
	if ow == nil {
		return steps
	}
	for i := range steps {
		if steps[i].Op != "openwin" || steps[i].Opener == "" {
			continue
		}
		if steps[i].Opener != ow.Program {
			continue
		}
		switch ow.Select() {
		case SelectChecked:
			steps[i].Value = "checked"
			steps[i].Why += "；开窗是勾选式（INPUT ARRAY，" + ow.SelectSource() + "）"
		case SelectCurrentRow:
			steps[i].Value = "current-row"
			steps[i].Why += "；开窗是当前行式（DISPLAY ARRAY，" + ow.SelectSource() + "）"
		default:
			steps[i].Value = "unknown"
			steps[i].Why += "；开窗的回传语义**未判定**（不猜）"
		}
	}
	return steps
}

// PlanText 把计划打成给人看的一行行（供 CLI 与文档复用）。
func PlanText(steps []PlanStep) string {
	var b strings.Builder
	for _, s := range steps {
		req := ""
		if s.Required {
			req = " [必填]"
		}
		grp := ""
		if s.Table != "" {
			grp = " {" + s.Table + "}"
		}
		fmt.Fprintf(&b, "%2d. %-7s %-28s%s%s", s.Seq, s.Op, s.Role, req, grp)
		if s.Label != "" {
			fmt.Fprintf(&b, "  %s", s.Label)
		}
		if s.Opener != "" {
			fmt.Fprintf(&b, "  开窗 %s", s.Opener)
		}
		b.WriteString("\n")
	}
	return b.String()
}
