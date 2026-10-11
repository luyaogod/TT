package fep

import (
	"strings"
	"testing"
)

// W3：开窗语义抽取的回归。
//
// 这两段源码片段**不是编的** —— 它们是现场两个开窗的真实形态（把注释与无关行删掉、
// 保留结构），出处写在夹具注释里：
//
//	q_pmaa001_3.4gl:361-546（供应商）；q_imaf001_15.4gl:345-475（料件）
//
// 为什么值得为它写测试：这段阅读一旦回退，驱动就会回到"试参数"的老路 ——
// 而那条路的代价是"运行期只回空 ACK、看起来像没反应"，我为此浪费了一整天。

// supplierSnippet 是 q_pmaa001_3 的结构：两个对话框，一个勾选式、一个当前行式。
const supplierSnippet = `
   DIALOG ATTRIBUTES(UNBUFFERED)
      INPUT ARRAY gr_qry FROM s_qry.*
         ATTRIBUTES(COUNT=g_data_cnt, APPEND ROW=FALSE, INSERT ROW=FALSE, DELETE ROW=FALSE)
         BEFORE ROW
            LET li_ac = DIALOG.getCurrentRow("s_qry")
         ON CHANGE check
            CALL q_pmaa001_3_qry_check("", gr_qry[li_ac].check, li_ac, li_ac)
      END INPUT
      BEFORE DIALOG
         LET ldig_curr = ui.Dialog.getCurrent()
      ON ACTION accept
         CALL q_pmaa001_3_get_multiret()
         LET gs_action = "exit"
         EXIT DIALOG
      ON ACTION cancel
         LET g_qryparam.return1 = NULL
         EXIT DIALOG
   END DIALOG

   DIALOG ATTRIBUTES(UNBUFFERED)
      DISPLAY ARRAY gr_qry TO s_qry.*
         BEFORE ROW
            LET li_ac = DIALOG.getCurrentRow("s_qry")
      ON ACTION accept
         IF li_ac > 0 THEN
            LET g_qryparam.return1 = gr_qry[li_ac].pmaa001_1
            LET gs_action = "exit"
         END IF
         EXIT DIALOG
   END DIALOG

PRIVATE FUNCTION q_pmaa001_3_get_multiret()
   DEFINE li_i   LIKE type_t.num10
   IF g_qryparam.state = "c" THEN
      LET g_qryparam.return1 = ""
      FOR li_i = 1 TO gr_qry_sel.getLength()
         IF gr_qry_sel[li_i].check = "Y" THEN
            LET g_qryparam.return1 = gr_qry_sel[li_i].pmaa001_1
         END IF
      END FOR
   END IF
END FUNCTION
`

// itemSnippet 是 q_imaf001_15 的结构：CONSTRUCT 条件列 + 两个对话框。
const itemSnippet = `
   CONSTRUCT gi_cons_where ON imaf001, imaal003, imaal004
      BEFORE CONSTRUCT
         CALL FGL_DIALOG_GETBUFFER() RETURNING ls_result
   END CONSTRUCT

   DIALOG ATTRIBUTES(UNBUFFERED)
      INPUT ARRAY gr_qry FROM s_qry.*
         BEFORE ROW
            LET li_ac = DIALOG.getCurrentRow("s_qry")
         ON CHANGE check
            CALL q_imaf001_15_qry_check("", gr_qry[li_ac].check, li_ac, li_ac)
      END INPUT
      ON ACTION accept
         CALL q_imaf001_15_get_multiret()
         LET gs_action = "exit"
         EXIT DIALOG
   END DIALOG

PRIVATE FUNCTION q_imaf001_15_get_multiret()
   DEFINE li_i   LIKE type_t.num10
   IF g_qryparam.state = "c" THEN
      LET g_qryparam.return1 = ""
      FOR li_i = 1 TO gr_qry_sel.getLength()
         IF gr_qry_sel[li_i].check = "Y" THEN
            LET g_qryparam.return1 = gr_qry_sel[li_i].imaf001_1
         END IF
      END FOR
   END IF
END FUNCTION
`

func TestParseSupplierOpenWinHasBothDialogSemantics(t *testing.T) {
	ow := ParseOpenWin4gl("q_pmaa001_3", supplierSnippet, nil)
	if len(ow.Dialogs) != 2 {
		t.Fatalf("对话框数 = %d，想要 2（这个开窗一勾选、一当前行）\n%+v", len(ow.Dialogs), ow.Dialogs)
	}
	// 运行期说 InputArray 时 → 勾选式
	got, d := SelectModeFor(ow, "InputArray")
	if got != SelectChecked {
		t.Errorf("InputArray 的选中语义 = %q，想要 %q", got, SelectChecked)
	}
	if d == nil || d.CheckColumn != "check" || d.ValueChecked != "Y" {
		t.Errorf("勾选栏读错：%+v", d)
	}
	if d != nil && len(d.ReturnFields) == 0 || (d != nil && d.ReturnFields[0] != "pmaa001_1") {
		t.Errorf("回传字段读错：%+v", d)
	}
	// 运行期说 DisplayArray 时 → 当前行式（现场就是这么成的）
	got, d = SelectModeFor(ow, "DisplayArray")
	if got != SelectCurrentRow {
		t.Errorf("DisplayArray 的选中语义 = %q，想要 %q", got, SelectCurrentRow)
	}
	if d == nil || d.RowVar != "li_ac" || len(d.ReturnFields) == 0 || d.ReturnFields[0] != "pmaa001_1" {
		t.Errorf("当前行式读错：%+v", d)
	}
	// Construct 阶段没有"选中"
	if got, _ := SelectModeFor(ow, "Construct"); got != SelectUnknown {
		t.Errorf("Construct 不该给出选中语义，得到 %q", got)
	}
	// 运行期报了一个源码里没有的类型 → 不许猜
	if got, _ := SelectModeFor(ow, "Whatever"); got != SelectUnknown {
		t.Errorf("未知 dialogType 必须回 unknown，得到 %q", got)
	}
}

func TestParseItemOpenWinReadsConstructColumns(t *testing.T) {
	ow := ParseOpenWin4gl("q_imaf001_15", itemSnippet, nil)
	if got, want := strings.Join(ow.ConstructCols, ","), "imaf001,imaal003,imaal004"; got != want {
		t.Errorf("查询条件列 = %q，想要 %q", got, want)
	}
	got, d := SelectModeFor(ow, "InputArray")
	if got != SelectChecked {
		t.Fatalf("InputArray 的选中语义 = %q，想要 %q —— 而现场我正是按「当前行」去试、"+
			"结果运行期只回空 ACK（没勾选就没有回传）", got, SelectChecked)
	}
	if d == nil || d.ValueChecked != "Y" || d.ReturnFields[0] != "imaf001_1" {
		t.Errorf("料件开窗的回传/勾选读错：%+v", d)
	}
	// 勾选栏是**标志**不是回传值：不许混进 ReturnFields
	for _, f := range d.ReturnFields {
		if f == "check" {
			t.Error("勾选栏混进了回传字段（它是标志，不是要写回的值）")
		}
	}
}

// TestParseOpenWinSaysUnknownWhenItCannotTell：读不出判据时必须如实说 unknown，
// 而不是默默给一个默认值 —— 默认值就是"猜"，而猜错的表现是"运行期只回空 ACK"。
func TestParseOpenWinSaysUnknownWhenItCannotTell(t *testing.T) {
	ow := ParseOpenWin4gl("q_x", `
   DIALOG ATTRIBUTES(UNBUFFERED)
      DISPLAY ARRAY gr_qry TO s_qry.*
      ON ACTION accept
         EXIT DIALOG
   END DIALOG
`, nil)
	if got := ow.Select(); got != SelectUnknown {
		t.Fatalf("没有判据时给了 %q，想要 unknown", got)
	}
}

// TestApplyCheckBox4fdFillsTheLiteralValue：勾上到底写什么，只能从 .4fd 读（Y / N）。
func TestApplyCheckBox4fdFillsTheLiteralValue(t *testing.T) {
	ow := ParseOpenWin4gl("q_imaf001_15", itemSnippet, nil)
	fields := ParseForm4fd("q_imaf001_15", `<CheckBox fieldId="1" name="check" noEntry="false" notNull="true" `+
		`valueChecked="Y" valueUnchecked="N" widget="CheckBox"/>`)
	ow.ApplyCheckBox4fd("q_imaf001_15.4fd", fields)
	d := ow.SelectDialog()
	if d == nil || d.ValueChecked != "Y" || d.ValueUnchecked != "N" {
		t.Fatalf("勾选字面值没读到：%+v", d)
	}
	if len(d.Evidence) == 0 {
		t.Error("补全字面值也该留下出处")
	}
}

// TestOpenWinProgramFromWindow：**驱动不该需要人告诉它"我刚开的是哪个开窗"**——
// 运行期在窗口名里写着了（实测三个开窗都对得上）。
func TestOpenWinProgramFromWindow(t *testing.T) {
	cases := map[string]string{
		"w_qry_q_pmaa001_3":  "q_pmaa001_3",
		"w_qry_q_imaf001_15": "q_imaf001_15",
		"w_qry_q_ooba002_1":  "q_ooba002_1",
		"w_apmt500":          "", // 主窗：不是开窗
		"w_qry_":             "",
		"w_qry_notaq":        "", // 开窗程序一定以 q_ 开头
		"":                   "",
	}
	for in, want := range cases {
		if got := OpenWinProgramFromWindow(in); got != want {
			t.Errorf("OpenWinProgramFromWindow(%q) = %q，想要 %q", in, got, want)
		}
	}
}

// TestReconcileSelectRefusesToGuess 是 R2 的落点：
// 没有源码、或源码里对不上时，必须回 unknown —— **不许**拿"InputArray 就一定是勾选"顶替。
//
// 那个推断在今天这两个开窗上恰好成立，但它是观察不是规矩；一旦哪个开窗反着来，
// 猜错的代价是"运行期只回空 ACK"，看起来像"没生效"（今天为此耗掉一整天）。
func TestReconcileSelectRefusesToGuess(t *testing.T) {
	ow := ParseOpenWin4gl("q_pmaa001_3", supplierSnippet, nil)
	// 有源码：对得上 → 用源码的（带出处）
	if got, why := ReconcileSelect(ow, "InputArray"); got != SelectChecked || why == "" {
		t.Errorf("InputArray → %q（%s），想要 %q", got, why, SelectChecked)
	}
	if got, _ := ReconcileSelect(ow, "DisplayArray"); got != SelectCurrentRow {
		t.Errorf("DisplayArray → %q，想要 %q", got, SelectCurrentRow)
	}
	// 没有源码 → unknown（并且说明原因）
	if got, why := ReconcileSelect(nil, "InputArray"); got != SelectUnknown || !strings.Contains(why, "没有源码") {
		t.Errorf("没有源码时 = %q（%s），想要 unknown + 说明", got, why)
	}
	// 源码里没有这个对话框 → unknown
	only := ParseOpenWin4gl("q_x", `
   DIALOG ATTRIBUTES(UNBUFFERED)
      INPUT ARRAY gr_qry FROM s_qry.*
      ON ACTION accept
         CALL q_x_get()
         EXIT DIALOG
   END DIALOG
`, nil)
	if got, why := ReconcileSelect(only, "DisplayArray"); got != SelectUnknown || !strings.Contains(why, "没有") {
		t.Errorf("源码没有 DISPLAY ARRAY 时 = %q（%s），想要 unknown + 说明", got, why)
	}
	// Construct 阶段没有"选中"
	if got, _ := ReconcileSelect(ow, "Construct"); got != SelectUnknown {
		t.Errorf("Construct → %q，想要 unknown", got)
	}
	// 阶段不认识 → unknown
	if got, _ := ReconcileSelect(ow, "Whatever"); got != SelectUnknown {
		t.Errorf("未知阶段 → %q，想要 unknown", got)
	}
	// 显式接受的"按阶段推断"：能用，但要能看出它不是源码结论
	if got := StageDefaultSelect("InputArray"); got != SelectChecked {
		t.Errorf("StageDefaultSelect(InputArray) = %q", got)
	}
	if got := StageDefaultSelect("Construct"); got != SelectUnknown {
		t.Errorf("StageDefaultSelect(Construct) = %q，想要 unknown", got)
	}
}
