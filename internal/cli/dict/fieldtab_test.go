package dict

// `tt dict fieldtab` 的钉子：**人读输出里必须写出「未解析」**，而且不许回落成 colName。
//
// 这不是可选的装饰：JSON 里 label 是空串 + labelState=unresolved，但文本表格里的空单元格
// 会被读成"这一列没意义" —— 所以文本形态必须把"未解析"三个字打出来（契约 A15/A18）。

import (
	"strings"
	"testing"

	"tt/internal/dict/formfields"
)

func sampleTable() *formfields.Table {
	return &formfields.Table{
		Module:          "t100",
		Form:            "demofm",
		Path:            `D:\fixtures\forms\t100\demofm.42f`,
		LabelSource:     formfields.LabelSourceDict,
		LabelResolved:   1,
		LabelUnresolved: 1,
		Fields: []formfields.Field{
			{
				ColName: "b_demo001", FieldID: 7, Tag: "TableColumn", Widget: "Edit",
				SQLType: "SMALLINT", Required: true,
				Comment: "cmt_demo001", TextKey: "lbl_demo001",
				Label: "演示栏位", LabelState: formfields.LabelResolved, LabelFrom: "demo_t.demo001",
			},
			{
				ColName: "b_demo004", FieldID: 18, Tag: "TableColumn", Widget: "Edit",
				SQLType: "VARCHAR(20)", Comment: "cmt_demo004",
				Label: "", LabelState: formfields.LabelUnresolved, LabelReason: formfields.ReasonNotInDict,
			},
		},
	}
}

func TestFieldtabRows_MarksUnresolved(t *testing.T) {
	tab := sampleTable()
	rows := fieldtabRows(tab)
	if len(rows) != 2 {
		t.Fatalf("行数 = %d, want 2", len(rows))
	}
	if len(fieldtabCSVHeaders) != len(rows[0]) {
		t.Fatalf("表头 %d 列、数据 %d 列 —— 两者必须对齐", len(fieldtabCSVHeaders), len(rows[0]))
	}
	// rows[0] 的中文名列（下标 7）
	if rows[0][7] != "演示栏位" {
		t.Errorf("已解析行的中文名 = %q, want 演示栏位", rows[0][7])
	}
	if rows[1][7] != "未解析" {
		t.Errorf("未解析行的中文名 = %q, want 未解析（空单元格会被读成「没有意义」）", rows[1][7])
	}
	if rows[1][7] == tab.Fields[1].ColName {
		t.Errorf("未解析行列落成了 colName —— 违反 A15/A18")
	}
	if rows[1][10] != formfields.ReasonNotInDict {
		t.Errorf("未解析原因列 = %q, want %q", rows[1][10], formfields.ReasonNotInDict)
	}
}

func TestFieldtabTextRows_ShowsState(t *testing.T) {
	rows := fieldtabTextRows(sampleTable())
	if len(rows) != 2 {
		t.Fatalf("行数 = %d, want 2", len(rows))
	}
	if rows[0][5] != "演示栏位" || rows[0][6] != "已解析" {
		t.Errorf("已解析行 = 中文名%q 状态%q，想要 演示栏位/已解析", rows[0][5], rows[0][6])
	}
	if rows[1][5] != "未解析" || rows[1][6] != formfields.ReasonNotInDict {
		t.Errorf("未解析行 = 中文名%q 状态%q，想要 未解析/%s",
			rows[1][5], rows[1][6], formfields.ReasonNotInDict)
	}
}

func TestSplitFormArg(t *testing.T) {
	cases := []struct{ in, module, form string }{
		{"t100/axmt500_wf", "t100", "axmt500_wf"},
		{`t100\axmt500_wf`, "t100", "axmt500_wf"},
		{"fglfind", "", "fglfind"},
	}
	for _, c := range cases {
		m, f, err := splitFormArg(c.in)
		if err != nil {
			t.Errorf("%q 解析失败: %v", c.in, err)
			continue
		}
		if m != c.module || f != c.form {
			t.Errorf("%q → %q/%q, want %q/%q", c.in, m, f, c.module, c.form)
		}
	}
	for _, bad := range []string{"", "/", "a/b/c", "/axmt500_wf"} {
		if _, _, err := splitFormArg(bad); err == nil {
			t.Errorf("%q 应当报错（不猜）", bad)
		} else if !strings.Contains(err.Error(), "模块") && !strings.Contains(err.Error(), "表单") {
			t.Errorf("%q 的错误信息要告诉人怎么打: %v", bad, err)
		}
	}
}
