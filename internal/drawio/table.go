package drawio

import (
	"fmt"
	"strings"
)

// 表格的共享规则：合成期（040-table.json 的 table 段）与运行期（spec 的
// items[].table）必须用同一套，所以只此一份。
//
// 为什么非得用 drawio 原生的 shape=table + childLayout=tableLayout：选中表格时
// Format 面板 Arrange 标签里那几个「插入/删除行、插入/删除列」按钮，是靠
// isTable / isTableRow / isTableCell 三个谓词认出来的，而它们**只看父级链**。
// 换成"画一堆方框拼成表格"，结构性收益全没了。
//
// 坐标全部写死：TableLayout 只在编辑操作时执行，加载文件时不跑，不写死会全叠在 (0,0)。

// validateTable 校验一份表格数据。报错文案与原实现逐字一致 —— 它是给人看的。
func validateTable(t *Table, where string) error {
	if t == nil {
		return fmt.Errorf("%s: table 必须是对象", where)
	}
	if len(t.Columns) == 0 {
		return fmt.Errorf("%s: table.columns 必须是非空数组", where)
	}
	for i, c := range t.Columns {
		if !(c.W > 0) {
			return fmt.Errorf("%s: table.columns[%d].w 必须是正数", where, i)
		}
	}
	switch r := t.Rows.(type) {
	case nil:
		// 省略 = 只有表头
	case float64:
		// 数字简写：N = N 行全空格（列数跟 columns 走）
		if r != float64(int(r)) || r < 0 {
			return fmt.Errorf("%s: table.rows 数字形式必须是非负整数（= 空行数）", where)
		}
	case []any:
		for ri, row := range r {
			cells, ok := row.([]any)
			if !ok {
				return fmt.Errorf("%s: table.rows[%d] 必须是数组", where, ri)
			}
			if len(cells) != len(t.Columns) {
				return fmt.Errorf("%s: table.rows[%d] 有 %d 格，但 columns 有 %d 列",
					where, ri, len(cells), len(t.Columns))
			}
		}
	default:
		return fmt.Errorf("%s: table.rows 必须是数组或非负整数", where)
	}
	if t.HeaderHeight != nil && !(*t.HeaderHeight > 0) {
		return fmt.Errorf("%s: table.headerHeight 必须是正数", where)
	}
	if t.RowHeight != nil && !(*t.RowHeight > 0) {
		return fmt.Errorf("%s: table.rowHeight 必须是正数", where)
	}
	return nil
}

// rowsOf 把 rows 的数字简写统一展开成"空格数组"；其余原样返回。
func rowsOf(t *Table) [][]any {
	switch r := t.Rows.(type) {
	case float64:
		out := make([][]any, int(r))
		for i := range out {
			row := make([]any, len(t.Columns))
			for j := range row {
				row[j] = ""
			}
			out[i] = row
		}
		return out
	case []any:
		out := make([][]any, 0, len(r))
		for _, row := range r {
			if cells, ok := row.([]any); ok {
				out = append(out, cells)
				continue
			}
			out = append(out, nil)
		}
		return out
	}
	return nil
}

// tableSize 算出表格总尺寸：列宽之和 × (表头高 + 行数×行高)。
//
// 尺寸是算出来的、不用手写 —— 所以 040-table.json 里没有 w/h，由这里回填。
func tableSize(t *Table) (w, h float64) {
	rowH := 26.0
	if t.RowHeight != nil {
		rowH = *t.RowHeight
	}
	headH := rowH
	if t.HeaderHeight != nil {
		headH = *t.HeaderHeight
	}
	for _, c := range t.Columns {
		w += c.W
	}
	return w, headH + float64(len(rowsOf(t)))*rowH
}

// tableSlots 生成文字槽位：表头 h0..hN → 单元格 hcN；数据格 r行c列 → 同名。
// 槽位随列行数生成，所以 spec 里传了 table 之后可以继续用 text 覆盖其中几格。
func tableSlots(t *Table) map[string]string {
	slots := make(map[string]string)
	for i := range t.Columns {
		slots[fmt.Sprintf("h%d", i)] = fmt.Sprintf("hc%d", i)
	}
	for ri, row := range rowsOf(t) {
		for ci := range row {
			k := fmt.Sprintf("r%dc%d", ri, ci)
			slots[k] = k
		}
	}
	return slots
}

// tableXML 生成表格的三层嵌套：容器（shape=table）→ 行 → 列。
//
// 外层 geometry **只写 width/height、不带 x/y** —— compose 靠这一点定位外层单元格
// （它要用正则把 x/y 补进去）。少了这个约定，compose 会直接报"无法定位外层单元格"。
func tableXML(t *Table, presets Presets) string {
	cols := t.Columns
	rowH := 26.0
	if t.RowHeight != nil {
		rowH = *t.RowHeight
	}
	headH := rowH
	if t.HeaderHeight != nil {
		headH = *t.HeaderHeight
	}
	rows := rowsOf(t)
	W, H := tableSize(t)

	geom := func(x, y, w, h float64) string {
		return fmt.Sprintf(`<mxGeometry x="%s" y="%s" width="%s" height="%s" as="geometry"/>`,
			jsNum(x), jsNum(y), jsNum(w), jsNum(h))
	}

	// 外层**只写 width/height，不带 x/y** —— 这不是省事，是契约：compose 靠
	// 「外层格的 mxGeometry 以 width= 开头」认出它并补上目标坐标。这里补个 x="0"，
	// compose 就再也认不出来了（它会报"无法定位外层单元格"）。回归测试见 mxlib_test.go。
	cells := []string{
		fmt.Sprintf(`<mxCell id="t" value="" style="%s" vertex="1" parent="1">`+
			`<mxGeometry width="%s" height="%s" as="geometry"/></mxCell>`,
			escapeAttr(presets["tableBox"]), jsNum(W), jsNum(H)),
	}

	type rowSpec struct {
		id     string
		y, h   float64
		values []any
		kind   string
	}
	specs := []rowSpec{{id: "h", y: 0, h: headH, kind: "tableHeader"}}
	for _, c := range cols {
		specs[0].values = append(specs[0].values, c.Title)
	}
	for i, values := range rows {
		specs = append(specs, rowSpec{
			id: fmt.Sprintf("r%d", i), y: headH + float64(i)*rowH, h: rowH, values: values, kind: "tableCell",
		})
	}

	for _, rs := range specs {
		cells = append(cells, fmt.Sprintf(
			`<mxCell id="%s" value="" style="%s" vertex="1" parent="t">%s</mxCell>`,
			rs.id, escapeAttr(presets["tableRow"]), geom(0, rs.y, W, rs.h)))

		var x float64
		for ci, c := range cols {
			var v string
			if ci < len(rs.values) {
				v = jsString(rs.values[ci])
			}
			cells = append(cells, fmt.Sprintf(
				`<mxCell id="%sc%d" value="%s" style="%s" vertex="1" parent="%s">%s</mxCell>`,
				rs.id, ci, escapeAttr(v), escapeAttr(presets[rs.kind]), rs.id, geom(x, 0, c.W, rs.h)))
			x += c.W
		}
	}

	return `<mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/>` +
		strings.Join(cells, "") + `</root></mxGraphModel>`
}
