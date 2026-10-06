package drawio

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// 把一份排版规格展开成 .drawio。原实现是 compose.js，这里逐条照搬它的规则。
//
// 整条链的分工：**写 spec 的人（多半是 AI）只决定「用哪个控件、放哪、写什么字」**，
// 查库、算栅格、平移坐标、重新编号、拼 XML 全在这里。这样它不碰 XML 就不会写坏转义，
// 也不用知道库的内部结构 —— 库改了不用重新教它。

// Result 是一次 compose 的产物。
type Result struct {
	XML     string // 完整的 .drawio 内容
	Library string // 用的是哪个库的展示名（报给用户看）
	Count   int    // 摆了几个控件
	Cells   int    // 产物里的单元格数（自查时数出来的）
	Report  string // 列宽报告（人读）：用了栅格就报每列的 x 与宽，全绝对定位就说明没有
}

// SelfCheckError 是**产物自查未通过**：id 重复、parent 悬空、几何里出现 NaN。
//
// 与"你的 spec 写错了"分开是有意的 —— 这一类是工具自己的 bug，重跑一次也不会好，
// 命令层据此退 3（见 internal/cli/drawio 的退出码表）。
type SelfCheckError struct{ Msg string }

func (e *SelfCheckError) Error() string { return e.Msg }

// 复合控件在库里是「外层 group/表格容器 + 若干部件」的结构，插入画布后要整体移动与缩放。
// 这几个正则就是那套重编号、拉伸、定位的落点。
var (
	innerRe = regexp.MustCompile(`(?s)<mxGraphModel><root>(.*)</root></mxGraphModel>`)
	// anyIDRe 找任意 id="…"，用于给单元格片段统一加前缀 —— 那一遍跑在**纯单元格**上，
	// 敞开匹配没有副作用。
	anyIDRe = regexp.MustCompile(`id="([^"]+)"`)
	// cellIDRe 只在自检时用，**必须带 `<mxCell`**：整份文档里 `grid="1"` 这种属性
	// 含有子串 `id="1"`，裸 `id="([^"]+)"` 会把它当成单元格 id，报出假的"id 重复"。
	cellIDRe = regexp.MustCompile(`<mxCell\s+id="([^"]+)"`)
	parentRe = regexp.MustCompile(`parent="([^"]+)"`)
	resizeRe  = regexp.MustCompile(`(<mxCell id="([^"]+)"[^>]*><mxGeometry) x="(-?[\d.]+)" y="(-?[\d.]+)" width="([\d.]+)" height="([\d.]+)"`)
	partNumRe = regexp.MustCompile(`^p(\d+)$`)
	// decoChildRe 认多页签生成的那些页签格（`p2_0`）—— 它们是装饰件，不参与拉伸。
	decoChildRe = regexp.MustCompile(`^p\d+_`)
)

// Compose 把一份 spec 展开成 .drawio。
func (s *Source) Compose(spec *Spec) (*Result, error) {
	cat := s.Catalog(spec.Library)
	if cat == nil {
		names := make([]string, 0, len(s.Catalogs))
		for _, c := range s.Catalogs {
			names = append(names, c.Name)
		}
		return nil, fmt.Errorf("没有名为 %q 的库（可选: %s）", spec.Library, strings.Join(names, ", "))
	}

	items, err := resolveItems(cat, spec)
	if err != nil {
		return nil, err
	}
	placed, colW := layoutSpec(spec, items)

	var cells strings.Builder
	for i := range placed {
		cell, err := placeItem(placed[i].catalog, s.Presets, &placed[i])
		if err != nil {
			return nil, err
		}
		cells.WriteString(cell)
	}

	out := buildDiagram(cells.String(), spec)
	n, err := selfCheck(out)
	if err != nil {
		return nil, err
	}
	return &Result{
		XML:     out,
		Library: cat.Library.Name,
		Count:   len(placed),
		Cells:   n,
		Report:  layoutReport(placed, colW),
	}, nil
}

// resolvedItem 是一个已经把控件查出来的 item。
type resolvedItem struct {
	SpecItem
	catalog *CatalogShape
	index   int
	x, y    float64 // 决定后的坐标（栅格算出来的，或 spec 直接给的）
	hasXY   bool
}

// colOf / rowOf 把"没写 col"当作第 0 列 —— 原实现是 `it.col ?? 0`。
//
// 这一条不只是取坐标：**没写 col 的控件也参与第 0 列的列宽计算**，所以不能跳过。
func colOf(it *resolvedItem) int {
	if it.Col == nil {
		return 0
	}
	return *it.Col
}

func rowOf(it *resolvedItem) int {
	if it.Row == nil {
		return 0
	}
	return *it.Row
}

// resolveItems 把 spec 里的每一项查成交互对象。查不到就报错，并列出可用 id ——
// 报错里给人下一步要的东西，比只喊一句"找不到"有用。
func resolveItems(cat *Catalog, spec *Spec) ([]resolvedItem, error) {
	out := make([]resolvedItem, 0, len(spec.Items))
	for i, it := range spec.Items {
		found := cat.Find(it.Shape)
		if found == nil {
			return nil, fmt.Errorf("items[%d] 找不到控件 %q（可用 id 见 `tt drawio lib --catalog`，或写 title 里的英文名）", i, it.Shape)
		}
		out = append(out, resolvedItem{SpecItem: it, catalog: found, index: i})
	}
	return out, nil
}

// layoutSpec 算栅格坐标。列宽取该列最宽控件，行高取该行最高控件；`grid.colWidths` 可覆盖列宽。
//
// spec 里显式给了 x/y 的项无视栅格（整屏/分区布局就靠这个）。
func layoutSpec(spec *Spec, items []resolvedItem) ([]resolvedItem, []float64) {
	originX, originY, colGap, rowGap := 40.0, 40.0, 16.0, 10.0
	var colWidths []float64
	if g := spec.Grid; g != nil {
		if g.OriginX != nil {
			originX = *g.OriginX
		}
		if g.OriginY != nil {
			originY = *g.OriginY
		}
		if g.ColGap != nil {
			colGap = *g.ColGap
		}
		if g.RowGap != nil {
			rowGap = *g.RowGap
		}
		colWidths = g.ColWidths
	}

	cols, rows := 1, 1
	for i := range items {
		if c := colOf(&items[i]) + 1; c > cols {
			cols = c
		}
		if r := rowOf(&items[i]) + 1; r > rows {
			rows = r
		}
	}

	colW := make([]float64, cols)
	rowH := make([]float64, rows)
	for i := range items {
		if w := items[i].catalog.W; w > colW[colOf(&items[i])] {
			colW[colOf(&items[i])] = w
		}
		if h := items[i].catalog.H; h > rowH[rowOf(&items[i])] {
			rowH[rowOf(&items[i])] = h
		}
	}
	for i, w := range colWidths {
		if i < cols {
			colW[i] = w
		}
	}

	colX := make([]float64, cols)
	for c, x := 0, originX; c < cols; c++ {
		colX[c] = x
		x += colW[c] + colGap
	}
	rowY := make([]float64, rows)
	for r, y := 0, originY; r < rows; r++ {
		rowY[r] = y
		y += rowH[r] + rowGap
	}

	for i := range items {
		it := &items[i]
		if it.X != nil {
			it.x = *it.X
		} else {
			it.x = colX[colOf(it)]
		}
		if it.Y != nil {
			it.y = *it.Y
		} else {
			it.y = rowY[rowOf(it)]
		}
	}
	return items, colW
}

// layoutReport 报每列的 x 与宽 —— 列宽是隐式的，宽控件（TextEdit 是 314）会把整列撑开，
// 表单看着被拉稀但不报任何错。把算出来的数打出来，跑完瞄一眼就知道有没有被撑开。
func layoutReport(items []resolvedItem, colW []float64) string {
	usedGrid := false
	for i := range items {
		if items[i].Col != nil || items[i].Row != nil {
			usedGrid = true
			break
		}
	}
	if !usedGrid {
		return "（本图全部使用 x/y 绝对定位，未使用 col/row 栅格，无列宽报告）"
	}

	colX := map[int]float64{}
	for i := range items {
		colX[colOf(&items[i])] = items[i].x
	}
	parts := make([]string, 0, len(colW))
	for c, w := range colW {
		xs := "-"
		if v, ok := colX[c]; ok {
			xs = jsNum(v)
		}
		parts = append(parts, fmt.Sprintf("col%d x=%s w=%s", c, xs, jsNum(w)))
	}
	return strings.Join(parts, "  ") +
		"\n      （某列被单个宽控件撑开时，用 grid.colWidths 显式指定列宽）"
}

// placeItem 把一个控件展开成可以塞进目标图的单元格片段。
//
// 五步，顺序都不能换：
//  1. 多页签（Folder）—— 页签条按 pages 画出全部页名，当前页白底
//  2. 换文字 —— **必须在加 id 前缀之前**：槽位里存的是原始单元格 id
//  3. 剥掉库条目自带的 0/1 图层根格，再给其余 id 统一加 `i{序号}-` 前缀防冲突
//  4. 拉伸 —— 三类子格行为，见下面
//  5. 定位外层格
func placeItem(cat *CatalogShape, presets Presets, it *resolvedItem) (string, error) {
	baseW, baseH := cat.W, cat.H
	slots := cat.Slots
	xmlStr := ""

	if it.Table != nil {
		if cat.Table == nil {
			return "", fmt.Errorf("items[%d]: 控件 %q 不支持 table 参数 —— 只有 Table 支持（清单里标了「列行可由 spec 的 table 参数自定义」的才支持）", it.index, cat.ID)
		}
		merged, err := mergeTable(it.Table, cat.Table, it.index)
		if err != nil {
			return "", err
		}
		if err := validateTable(merged, fmt.Sprintf("items[%d].table", it.index)); err != nil {
			return "", err
		}
		xmlStr = tableXML(merged, presets)
		baseW, baseH = tableSize(merged)
		slots = tableSlots(merged)
	} else {
		if cat.XML == "" {
			return "", fmt.Errorf("控件 %q 的 XML 是空的（形状源里的条目不对）", cat.ID)
		}
		xmlStr = cat.XML
	}

	effW, effH := baseW, baseH
	if it.W != nil {
		effW = *it.W
	}
	if it.H != nil {
		effH = *it.H
	}

	// ① 多页签
	if it.Pages != nil {
		next, err := applyPages(cat, presets, it, xmlStr)
		if err != nil {
			return "", err
		}
		xmlStr = next
	}

	// ② 换文字（在加前缀之前 —— 槽位里存的是原始 id）
	if err := applyText(slots, cat.ID, it, &xmlStr); err != nil {
		return "", err
	}

	// ③ 取内层、剥根格、加前缀
	m := innerRe.FindStringSubmatch(xmlStr)
	if m == nil {
		return "", fmt.Errorf("控件 %q 的 XML 结构异常", cat.ID)
	}
	inner := m[1]

	prefix := fmt.Sprintf("i%d-", it.index)
	outerBase := "2"
	if strings.Contains(inner, `id="t"`) {
		outerBase = "t"
	} else if strings.Contains(inner, `id="g"`) {
		outerBase = "g"
	}
	outerID := prefix + outerBase

	cells := strings.Replace(inner, `<mxCell id="0"/>`, "", 1)
	cells = strings.Replace(cells, `<mxCell id="1" parent="0"/>`, "", 1)
	cells = anyIDRe.ReplaceAllStringFunc(cells, func(s string) string {
		return `id="` + prefix + anyIDRe.FindStringSubmatch(s)[1] + `"`
	})
	cells = parentRe.ReplaceAllStringFunc(cells, func(s string) string {
		pid := parentRe.FindStringSubmatch(s)[1]
		// 父级是 0/1 的说明它挂在图层上、是顶层，前缀会把它变成悬空引用
		if pid == "0" || pid == "1" {
			return s
		}
		return `parent="` + prefix + pid + `"`
	})

	// ④ 拉伸：三类子格行为
	if sx, sy := effW/baseW, effH/baseH; math.Abs(sx-1) > 1e-9 || math.Abs(sy-1) > 1e-9 {
		cells = resizeCells(cells, cat, baseW, baseH, effW, effH, sx, sy)
	}

	// ⑤ 定位外层
	posRe := regexp.MustCompile(`(id="` + regexp.QuoteMeta(outerID) + `"[^>]*><mxGeometry) width="[^"]*" height="[^"]*"`)
	if !posRe.MatchString(cells) {
		return "", fmt.Errorf("无法定位 %q 的外层单元格 —— 形状条目里外层格的 mxGeometry 应当是「width=… height=…」开头，不能带 x/y（位置由 compose 填）", cat.ID)
	}
	cells = replaceFirstSubmatch(posRe, cells, func(m []string) string {
		return fmt.Sprintf(`%s x="%s" y="%s" width="%s" height="%s"`,
			m[1], jsNum(it.x), jsNum(it.y), jsNum(effW), jsNum(effH))
	})
	return cells, nil
}

// applyPages 处理 Folder 的多页签。
//
// 页签不可点击是**刻意的**（drawio 桌面版的自定义链接有已知 bug）：标准画法是
// 「页签条画全 + 内容框只画当前页，其余页并排再摆一个 Folder」。
func applyPages(cat *CatalogShape, presets Presets, it *resolvedItem, xmlStr string) (string, error) {
	if cat.MultiTab == nil {
		return "", fmt.Errorf("items[%d]: 控件 %q 不支持 pages 参数 —— 只有 Folder 支持", it.index, cat.ID)
	}
	if len(it.Pages) == 0 {
		return "", fmt.Errorf("items[%d].pages 必须是非空字符串数组", it.index)
	}
	act := 0
	if it.Active != nil {
		act = *it.Active
	}
	if act < 0 || act >= len(it.Pages) {
		return "", fmt.Errorf("items[%d].active 必须是 0..%d 的整数，当前是 %d", it.index, len(it.Pages)-1, act)
	}

	mt := cat.MultiTab
	bx := func(i int) float64 { return float64(i) * mt.XStep }
	// 其余页签是**兄弟格**（`p{part}_{i}`）；选中的那个**保持原来的 id**（`p{part}`）——
	// 库里那个部件本来就叫 p{part}，拿它当"当前页"是原地替换，不是新增。
	tab := func(i int, style string) string {
		return fmt.Sprintf(`<mxCell id="p%d_%d" value="%s" style="%s" vertex="1" parent="g">`+
			`<mxGeometry x="%s" y="0" width="%s" height="%s" as="geometry"/></mxCell>`,
			mt.Part, i, escapeAttr(it.Pages[i]), escapeAttr(style), jsNum(bx(i)), jsNum(mt.W), jsNum(mt.H))
	}
	activeTab := func(i int, style string) string {
		return fmt.Sprintf(`<mxCell id="p%d" value="%s" style="%s" vertex="1" parent="g">`+
			`<mxGeometry x="%s" y="0" width="%s" height="%s" as="geometry"/></mxCell>`,
			mt.Part, escapeAttr(it.Pages[i]), escapeAttr(style), jsNum(bx(i)), jsNum(mt.W), jsNum(mt.H))
	}

	// 选中的那个保持原来的 id（p{part}），其余页签是它的兄弟（p{part}_{i}）
	tabRe := regexp.MustCompile(fmt.Sprintf(
		`<mxCell id="p%d" value="[^"]*" style="[^"]*" vertex="1" parent="g">`+
			`<mxGeometry x="-?[\d.]+" y="-?[\d.]+" width="[\d.]+" height="[\d.]+" as="geometry"/></mxCell>`, mt.Part))
	if !tabRe.MatchString(xmlStr) {
		return "", fmt.Errorf("内部错误：控件 %q 的页签部件 p%d 找不到", cat.ID, mt.Part)
	}

	var rest strings.Builder
	for i := range it.Pages {
		if i == act {
			continue
		}
		rest.WriteString(tab(i, presets["tab"]))
	}
	return replaceFirst(tabRe, xmlStr, func(string) string {
		return activeTab(act, presets["tabActive"]) + rest.String()
	}), nil
}

// applyText 按文字槽位替换单元格里的文字。
func applyText(slots map[string]string, id string, it *resolvedItem, xmlStr *string) error {
	if len(it.Text) == 0 {
		return nil
	}
	// 按槽位名排序 —— Go 的 map 迭代序是随机的，不排的话同一份 spec 两次跑出的
	// 报错顺序都不一样，测试也没法钉。
	keys := make([]string, 0, len(it.Text))
	for k := range it.Text {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, slot := range keys {
		cellID, ok := slots[slot]
		if !ok {
			avail := make([]string, 0, len(slots))
			for k := range slots {
				avail = append(avail, k)
			}
			sort.Strings(avail)
			hint := "无，此控件不支持改文字"
			if len(avail) > 0 {
				hint = strings.Join(avail, ", ")
			}
			return fmt.Errorf("items[%d]: 控件 %q 没有文字槽位 %q（可用: %s）", it.index, id, slot, hint)
		}
		re := regexp.MustCompile(`(<mxCell id="` + regexp.QuoteMeta(cellID) + `" value=")[^"]*(")`)
		if !re.MatchString(*xmlStr) {
			return fmt.Errorf("内部错误：找不到单元格 %s", cellID)
		}
		value := jsString(it.Text[slot])
		*xmlStr = replaceFirstSubmatch(re, *xmlStr, func(m []string) string {
			return m[1] + escapeAttr(value) + m[2]
		})
	}
	return nil
}

// resizeCells 按三类子格行为重写子单元格的几何。
//
//	fill：容器框体，锚定 x/y 拉伸到填满外层（减去固定的另三边内缩）—— **边框跟着动**
//	deco：装饰件（页签、图例、树节点、多页签的兄弟页签）保持原尺寸锚在原位 —— **不被拉变形**
//	其余（含表格行列）：等比缩放 —— 字段行、表格内容随尺寸走
//
// 三类由形状源里的 `pfill` / `pdeco` 标出，值是**1 基的部件序号**。没有标的一律等比。
func resizeCells(cells string, cat *CatalogShape, baseW, baseH, effW, effH, sx, sy float64) string {
	pfill := map[int]bool{}
	for _, n := range cat.PFill {
		pfill[n] = true
	}
	pdeco := map[int]bool{}
	for _, n := range cat.PDeco {
		pdeco[n] = true
	}

	return resizeRe.ReplaceAllStringFunc(cells, func(s string) string {
		m := resizeRe.FindStringSubmatch(s)
		head, id := m[1], m[2]
		gx := atof(m[3])
		gy := atof(m[4])
		gw := atof(m[5])
		gh := atof(m[6])

		local := id
		if i := strings.IndexByte(id, '-'); i >= 0 {
			local = id[i+1:]
		}
		if pm := partNumRe.FindStringSubmatch(local); pm != nil {
			n := atoi(pm[1])
			if pfill[n] {
				ri := baseW - gx - gw
				bi := baseH - gy - gh
				return fmt.Sprintf(`%s x="%s" y="%s" width="%s" height="%s"`,
					head, jsNum(gx), jsNum(gy), jsNum(r2(effW-gx-ri)), jsNum(r2(effH-gy-bi)))
			}
			if pdeco[n] {
				return s
			}
		}
		if decoChildRe.MatchString(local) {
			return s
		}
		return fmt.Sprintf(`%s x="%s" y="%s" width="%s" height="%s"`,
			head, jsNum(r2(gx*sx)), jsNum(r2(gy*sy)), jsNum(r2(gw*sx)), jsNum(r2(gh*sy)))
	})
}

// mergeTable 把 spec 给的 table 与形状条目里的默认值合并成一份完整表格。
//
// 列行数据以 spec 为准；表头高与行高缺省时继承形状条目的，再缺省就是 26。
func mergeTable(t, def *Table, index int) (*Table, error) {
	merged := &Table{Columns: t.Columns, Rows: t.Rows}

	hh := 26.0
	if def.HeaderHeight != nil {
		hh = *def.HeaderHeight
	}
	if t.HeaderHeight != nil {
		hh = *t.HeaderHeight
	}
	rh := 26.0
	if def.RowHeight != nil {
		rh = *def.RowHeight
	}
	if t.RowHeight != nil {
		rh = *t.RowHeight
	}
	merged.HeaderHeight, merged.RowHeight = &hh, &rh

	// rows 的数字简写（N = N 行全空格）在这里展开成实打实的行 —— 下游统一按行数算尺寸
	if n, ok := t.Rows.(float64); ok {
		if n != math.Trunc(n) || n < 0 {
			return nil, fmt.Errorf("items[%d].table.rows 数字形式必须是非负整数（= 空行数）", index)
		}
		rows := make([]any, int(n))
		for i := range rows {
			row := make([]any, len(merged.Columns))
			for j := range row {
				row[j] = ""
			}
			rows[i] = row
		}
		merged.Rows = rows
	}
	return merged, nil
}

// buildDiagram 把单元格拼成一份完整的 .drawio。
func buildDiagram(cells string, spec *Spec) string {
	pw, ph := 850.0, 1100.0
	if spec.Page != nil {
		if spec.Page.Width != nil {
			pw = *spec.Page.Width
		}
		if spec.Page.Height != nil {
			ph = *spec.Page.Height
		}
	}
	title := spec.Title
	if title == "" {
		title = "Page-1"
	}
	return "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		`<mxfile host="app.diagrams.net" type="device">` +
		`<diagram id="t100" name="` + escapeAttr(title) + `">` +
		`<mxGraphModel dx="900" dy="700" grid="1" gridSize="10" guides="1" tooltips="1" connect="1"` +
		` arrows="1" fold="1" page="1" pageScale="1" pageWidth="` + jsNum(pw) + `" pageHeight="` + jsNum(ph) + `"` +
		` math="0" shadow="0">` +
		`<root><mxCell id="0"/><mxCell id="1" parent="0"/>` + cells + `</root>` +
		`</mxGraphModel></diagram></mxfile>` + "\n"
}

// selfCheck 自查产物：id 必须唯一、parent 必须能解析、几何里不能有 NaN。
//
// compose 靠 `i{序号}-` 前缀保证不撞，但那是"应该对"，不是"验过" —— 原实现这句话
// 说得对，这里照做，并**多加一条**：用真 XML 解析器走一遍 token，验它良构
// （原实现只查 NaN/undefined 与 id/parent，没验 XML 本身）。
func selfCheck(out string) (int, error) {
	if strings.Contains(out, "NaN") || strings.Contains(out, "undefined") {
		return 0, &SelfCheckError{
			Msg: "产物里出现 NaN/undefined 几何值 —— 通常是缩放或坐标计算出了 bug，请勿交付",
		}
	}

	ids := cellIDRe.FindAllStringSubmatch(out, -1)
	known := make(map[string]bool, len(ids))
	var dup []string
	for _, m := range ids {
		if known[m[1]] {
			dup = append(dup, m[1])
			continue
		}
		known[m[1]] = true
	}
	if len(dup) > 0 {
		return 0, &SelfCheckError{Msg: "产物 id 重复: " + strings.Join(uniq(dup), ", ")}
	}

	var dangling []string
	for _, m := range parentRe.FindAllStringSubmatch(out, -1) {
		if !known[m[1]] {
			dangling = append(dangling, m[1])
		}
	}
	if len(dangling) > 0 {
		return 0, &SelfCheckError{Msg: "产物里有悬空 parent: " + strings.Join(uniq(dangling), ", ")}
	}

	if err := wellFormed(out); err != nil {
		return 0, &SelfCheckError{Msg: fmt.Sprintf("产物不是良构 XML：%v", err)}
	}
	return len(ids), nil
}

// wellFormed 用真解析器走一遍 token。
func wellFormed(s string) error {
	dec := xml.NewDecoder(strings.NewReader(s))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// replaceFirst 只替换第一处 —— JS 里 String.replace 配非 /g 正则就是第一处。
func replaceFirst(re *regexp.Regexp, s string, repl func(string) string) string {
	loc := re.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + repl(s[loc[0]:loc[1]]) + s[loc[1]:]
}

// replaceFirstSubmatch 同上，但回调拿得到捕获组。
func replaceFirstSubmatch(re *regexp.Regexp, s string, repl func([]string) string) string {
	loc := re.FindStringSubmatchIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + repl(re.FindStringSubmatch(s)) + s[loc[1]:]
}

func r2(v float64) float64 { return math.Round(v*100) / 100 }

func atof(s string) float64 { f, _ := strconv.ParseFloat(s, 64); return f }

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }

func uniq(in []string) []string {
	seen := map[string]bool{}
	out := in[:0]
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
