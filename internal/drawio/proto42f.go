package drawio

// `.42f` → 原型数据：把编译后的表单文件落成 drawio 图纸，**几何取真值，控件族不猜**。
//
// 契约面（唯一口径）：`.agents/plan/tt-genero-formdef-contract.md` 的 C1（`.42f` 解析）
// 与 A15/A18（拿不准就不猜、启发式不得参与判定）。本文件**不重解析 XML** —— 解析由
// `internal/form42f.ParseForm42F` 负责，这里只消费它的结果（`F42Field` / `RecordViews`）。
//
// 三条设计口径，改这里之前先读：
//
//  1. **几何是真值**。`posX/posY/width/height/gridWidth` 原样记进 `ProtoField`，
//     画到图上的坐标 = 网格值 × `CellPx`（缺省 10，一个可调的口径，不是判定）。
//     文件里**没给**坐标的字段（`posX=posY=0`，多为明细列）不假装有坐标 ——
//     它们落在图下方的「推断排布带」，位置是流式推出来的，宽度仍是文件里的真值。
//     两条带在结构上分开（`ProtoField.Geometry` + 图例），不混成一个"看起来对"的渲染。
//
//  2. **控件族只认同名**。`WidgetShapes42F` 里那十条是「`.42f` 的元素名与形状库的形状 id
//     逐字相同」，那不叫推测。其余（`Completer` / `Image` / `PhantomEdit` /
//     `DateTimeEdit` / `Aggregate` …）库里没有对应图形 ⇒ 判**未判定**，画通用虚线矩形
//     并在图例里点名（A15）。**不许**为了图好看把 `DateTimeEdit` 画成 `DateEdit`。
//
//  3. **必填/只读落在图形上**。`required` / `notNull` → 文字带 `*` 前缀；
//     `noEntry` → 文字带 `[只读]` 后缀。图例解释这两个记号。文字用 `colName`
//     （C1.8：不猜中文名；`cmt_*` / `lbl_*` 原样留在字段里给上层用）。
//
// 映射成功的字段走 `Compose`（形状库、拉伸、id 前缀、自查全复用），未判定字段与图例
// 是追加的单元格 —— 既有出图契约（`Compose` / catalog / mxlibrary）一字不动。

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tt/internal/form42f"
)

// DefaultProtoCellPx 是一个 `.42f` 网格单位画成多少 drawio 像素。
//
// 为什么是 10：形状库里的 `Edit` 宽 209 = 「20 格字段 + 约 1 格标签」在 10 px/格 下的
// 尺寸，两者对得上。它**只影响像素**，不影响任何判定 —— 要逐字节核对几何时用 1。
const DefaultProtoCellPx = 10.0

// 布局常量。原点与留白是画布上的排版参数，与 `.42f` 无关。
const (
	protoOriginX    = 40.0
	protoOriginY    = 40.0
	protoGap        = 8.0
	protoBandGap    = 60.0
	protoMargin     = 80.0
	protoPackWidth  = 1600.0
	protoLegendW    = 900.0
	protoLegendLine = 16.0
)

// 图例与未判定矩形用的样式串（内联写死：这两类图元不属于形状库，进库会让
// `TestSynthShapeCounts` 的 18 个形状变成 19 个，等于私自改既有契约）。
const (
	protoUnmappedStyle = "rounded=0;whiteSpace=wrap;html=1;fillColor=#f5f5f5;strokeColor=#999999;dashed=1;fontSize=9;fontColor=#555555;"
	protoLegendStyle   = "text;html=1;align=left;verticalAlign=middle;fontSize=10;fontColor=#333333;spacingLeft=2;"
	protoBandStyle     = "text;html=1;align=left;verticalAlign=middle;fontSize=10;fontColor=#1f6f8b;fontStyle=1;spacingLeft=2;"
)

// WidgetShapes42F 是 `.42f` 控件元素名 → 形状库形状 id。
//
// 只收**名字逐字相同**的十条。库里没有同名图形的一律未判定（契约 A15）：
// 拿不准就不猜，画通用矩形并在图例里点名，而不是挑一个"看着像"的图形糊上去。
var WidgetShapes42F = map[string]string{
	"Edit":       "Edit",
	"ButtonEdit": "ButtonEdit",
	"ComboBox":   "ComboBox",
	"CheckBox":   "CheckBox",
	"DateEdit":   "DateEdit",
	"SpinEdit":   "SpinEdit",
	"RadioGroup": "RadioGroup",
	"TextEdit":   "TextEdit",
	"Label":      "Label",
	"Button":     "Button",
}

// ProtoOptions 是展开原型时的可调项。零值即缺省。
type ProtoOptions struct {
	// CellPx 是一个网格单位画成多少像素；<=0 取 DefaultProtoCellPx。
	CellPx float64
	// Title 是图页名；空则用表单名。
	Title string
	// PackWidth 是「推断排布带」的宽度（像素）；<=0 取 1600。
	PackWidth float64
}

type protoOpts struct {
	cellPx    float64
	title     string
	packWidth float64
}

func normalizeProtoOptions(o *ProtoOptions) protoOpts {
	out := protoOpts{cellPx: DefaultProtoCellPx, packWidth: protoPackWidth}
	if o == nil {
		return out
	}
	if o.CellPx > 0 {
		out.cellPx = o.CellPx
	}
	if o.PackWidth > 0 {
		out.packWidth = o.PackWidth
	}
	out.title = strings.TrimSpace(o.Title)
	return out
}

// ProtoField 是原型图上的一块字段（`.42f` 的一条字段条目）。
//
// `PosX/PosY/Width/Height/GridWidth` 是**文件原文**；`X/Y/W/H` 是画到图上用的像素值。
// 两者分开是有意的：核对"与文件一致"看前者，看"画在哪"看后者。
type ProtoField struct {
	Index   int
	ColName string
	FieldID int
	Tag     string // FormField | TableColumn | PhantomColumn | PhantomFormField
	Widget  string // `.42f` 里的控件元素名（原样；空 = 条目下没有控件元素）
	Shape   string // 用哪个库形状；空 = 未判定
	Mapped  bool   // 形状库里有没有同名图形

	PosX, PosY               int
	Width, Height, GridWidth int

	X, Y, W, H float64
	// Geometry=true：坐标来自文件里的网格值（`posX/posY` 任一非 0）。
	// false：落在推断排布带上（文件没给坐标），位置是排出来的，宽度仍是真值。
	Geometry bool
	// WidthSource / HeightSource 记录尺寸从哪来：width / gridWidth / shapeDefault / min。
	WidthSource  string
	HeightSource string

	Required, NotNull, NoEntry bool
	Comment                    string
	TextKey                    string

	Label    string
	CellID   string   // 产物里的单元格 id（未判定的是 u{序号}）
	TabNames []string // 由 RecordView 反查出来的 tabName（空 = 未进任何 RecordView）
}

// ProtoTab 是一个 RecordView 的分组：tabName → 它链接到的 colName。
type ProtoTab struct {
	TabName  string
	ColNames []string
	// Missing 是链接里指向、但表单里没有的 colName（数据异常，只报不猜）。
	Missing []string
}

// Proto 是一次 `.42f` → 原型的全部产物。
type Proto struct {
	Form   string
	Title  string
	CellPx float64
	XML    string

	Fields   []ProtoField
	Tabs     []ProtoTab
	Unmapped []string // 未判定的控件名（去重、排序）
	Legend   string   // 图例全文（与画在图上的一致）

	Report      string // compose 的列宽/定位报告
	Cells       int    // 产物里的单元格数
	Count       int    // 进了原型的字段数
	Skipped     int    // `.42f` 里缺 colName/fieldId 被跳过的条目数（解析层计数）
	SizeMissing int    // 文件里连 width/gridWidth 都没有的字段数（已按最小单元画）
	Warnings    []string
}

// Field 按 colName 取一条字段；没有则 nil。
func (p *Proto) Field(colName string) *ProtoField {
	for i := range p.Fields {
		if p.Fields[i].ColName == colName {
			return &p.Fields[i]
		}
	}
	return nil
}

// ProtoFrom42FFile 读一个 `.42f` 文件并展开成原型数据。
//
// 表单名取文件名（去掉 `.42f`）：`.42f` 里的 `<Form name>` 实测会写成
// `axmt500_wf.tmp`（编译中间名），不能当 scope（见 `form42f.F42Form.DeclaredName`）。
func ProtoFrom42FFile(path string, opts *ProtoOptions) (*Proto, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读不到表单文件 %s：%w", path, err)
	}
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return ProtoFrom42F(name, string(b), opts)
}

// ProtoFrom42F 把 `.42f` 文本展开成原型数据。
//
// `form` 是调用方知道的表单名（scope 用）；解析失败原样透出 `fep` 的错误
// （那是 `FORM_UNKNOWN` 的来源，命令层据此退码）。
func ProtoFrom42F(form, text string, opts *ProtoOptions) (*Proto, error) {
	o := normalizeProtoOptions(opts)

	ff, err := form42f.ParseForm42F(form, text)
	if err != nil {
		return nil, err
	}

	src, err := Load()
	if err != nil {
		return nil, fmt.Errorf("形状源坏了：%w", err)
	}
	cat := src.Catalog("controls")
	if cat == nil {
		return nil, fmt.Errorf("形状源里没有 controls 库 —— 表单原型只能按控件库画")
	}

	title := o.title
	if title == "" {
		title = form
	}
	p := &Proto{
		Form: form, Title: title, CellPx: o.cellPx,
		Skipped:  ff.Skipped,
		Warnings: append([]string(nil), src.Warnings...),
	}

	// ① 逐条字段：形状、尺寸、文字
	for i := range ff.Fields {
		p.Fields = append(p.Fields, makeProtoField42F(i, &ff.Fields[i], cat, o, p))
	}
	p.Count = len(p.Fields)

	// ② 位置：真实几何带 + 推断排布带
	packTop, bottom, right := placeProtoFields42F(p.Fields, o)

	// ③ RecordView 分组（tabName → colName），并回填到字段上
	p.Tabs = protoTabs42F(ff, p.Fields)
	for ti := range p.Tabs {
		for _, col := range p.Tabs[ti].ColNames {
			if f := p.Field(col); f != nil {
				f.TabNames = append(f.TabNames, p.Tabs[ti].TabName)
			}
		}
	}
	p.Unmapped = unmappedKinds42F(p.Fields)

	// ④ 图例：先算字号行数，才能定画布尺寸
	legendX, legendY := right+60, protoOriginY
	lines := protoLegend42F(p, o)
	p.Legend = strings.Join(lines, "\n")
	legendBottom := legendY + float64(len(lines))*protoLegendLine + protoLegendLine

	// ⑤ 映射成功的字段走 compose —— 复用形状库、拉伸、id 前缀与自查
	spec := &Spec{Library: "controls", Title: title}
	itemOf := make([]int, len(p.Fields))
	for i := range itemOf {
		itemOf[i] = -1
	}
	for i := range p.Fields {
		f := &p.Fields[i]
		if !f.Mapped {
			continue
		}
		x, y, w, h := f.X, f.Y, f.W, f.H
		it := SpecItem{Shape: f.Shape, X: &x, Y: &y, W: &w, H: &h}
		it.Text = protoTextSlots42F(cat, f.Shape, f.Label)
		itemOf[i] = len(spec.Items)
		spec.Items = append(spec.Items, it)
	}
	pw := protoMaxF(right+protoMargin, legendX+protoLegendW+protoMargin)
	ph := protoMaxF(bottom+protoMargin, legendBottom+protoMargin)
	spec.Page = &Page{Width: &pw, Height: &ph}

	res, err := src.Compose(spec)
	if err != nil {
		return nil, err
	}
	p.XML, p.Report = res.XML, res.Report
	for i := range p.Fields {
		f := &p.Fields[i]
		if itemOf[i] >= 0 {
			f.CellID = fmt.Sprintf("i%d-%s", itemOf[i], outerBase42F(cat, f.Shape))
		}
	}

	// ⑥ 追加：未判定矩形 + 排布带标题 + 图例
	var extra strings.Builder
	for i := range p.Fields {
		f := &p.Fields[i]
		if f.Mapped {
			continue
		}
		extra.WriteString(cell42F(f.CellID, f.Label, protoUnmappedStyle, f.X, f.Y, f.W, f.H))
	}
	if packTop > 0 {
		extra.WriteString(cell42F("lg-band", protoBandTitle, protoBandStyle,
			protoOriginX, packTop-28, protoPackWidth, 20))
	}
	for k, line := range lines {
		extra.WriteString(cell42F(fmt.Sprintf("lg%d", k), line, protoLegendStyle,
			legendX, legendY+float64(k)*protoLegendLine, protoLegendW, protoLegendLine))
	}
	p.XML = insertBeforeRootClose(p.XML, extra.String())

	n, err := selfCheck(p.XML)
	if err != nil {
		return nil, err
	}
	p.Cells = n
	return p, nil
}

// makeProtoField42F 决定一条字段的形状、尺寸与文字。
func makeProtoField42F(i int, f *form42f.F42Field, cat *Catalog, o protoOpts, p *Proto) ProtoField {
	pf := ProtoField{
		Index: i, ColName: string(f.ColName), FieldID: int(f.FieldID),
		Tag: f.Tag, Widget: f.Widget,
		PosX: f.Geom.PosX, PosY: f.Geom.PosY,
		Width: f.Geom.Width, Height: f.Geom.Height, GridWidth: f.Geom.GridWidth,
		Required: f.Required, NotNull: f.NotNull, NoEntry: f.NoEntry,
		Comment: f.Comment, TextKey: f.TextKey,
	}

	// 控件族：只有「名字逐字相同」才算映射成功（A15）
	var cs *CatalogShape
	if id, ok := WidgetShapes42F[f.Widget]; ok {
		cs = cat.Find(id)
	}
	if cs != nil {
		pf.Mapped, pf.Shape = true, cs.ID
	} else {
		pf.CellID = fmt.Sprintf("u%d", i)
	}

	// 尺寸：width → gridWidth → 形状库默认 → 最小单元。来源逐条记下，不静默兜底。
	switch {
	case pf.Width > 0:
		pf.W, pf.WidthSource = float64(pf.Width)*o.cellPx, "width"
	case pf.GridWidth > 0:
		pf.W, pf.WidthSource = float64(pf.GridWidth)*o.cellPx, "gridWidth"
	case cs != nil:
		pf.W, pf.WidthSource = cs.W, "shapeDefault"
	default:
		pf.W, pf.WidthSource = o.cellPx, "min"
		p.SizeMissing++
	}
	switch {
	case pf.Height > 0:
		pf.H, pf.HeightSource = float64(pf.Height)*o.cellPx, "height"
	case cs != nil:
		pf.H, pf.HeightSource = cs.H, "shapeDefault"
	default:
		pf.H, pf.HeightSource = o.cellPx, "min"
	}

	// 文字：身份用 colName（不猜中文名）；必填 / 只读用记号区分；未判定带 ? 与说明
	label := pf.ColName
	if pf.Required || pf.NotNull {
		label = "*" + label
	}
	if pf.NoEntry {
		label += " [只读]"
	}
	if !pf.Mapped {
		label = "? " + label + " [未判定 " + widgetName42F(pf.Widget) + "]"
	}
	pf.Label = label
	return pf
}

// placeProtoFields42F 给每条字段定位置，返回（排布带顶 y，内容底，内容右）。
//
// 两条带**结构性分开**：
//   - 文件给了网格坐标（`posX/posY` 任一非 0）→ 真实坐标 × CellPx，`Geometry=true`；
//   - 其余（多为明细列：坐标由所在表格的列序决定，文件里没有）→ 下方流式排布，
//     `Geometry=false`。宽度是文件里的真值，只有位置是推断的。
func placeProtoFields42F(fields []ProtoField, o protoOpts) (packTop, bottom, right float64) {
	bottom, right = protoOriginY, protoOriginX
	hasGeom := false
	for i := range fields {
		f := &fields[i]
		if f.PosX == 0 && f.PosY == 0 {
			continue
		}
		f.Geometry = true
		f.X, f.Y = float64(f.PosX)*o.cellPx, float64(f.PosY)*o.cellPx
		hasGeom = true
		bottom = protoMaxF(bottom, f.Y+f.H)
		right = protoMaxF(right, f.X+f.W)
	}

	x, y := protoOriginX, protoOriginY
	if hasGeom {
		y = bottom + protoBandGap
	}
	packTop = y
	rowH := 0.0
	packed := false
	for i := range fields {
		f := &fields[i]
		if f.Geometry {
			continue
		}
		packed = true
		if x > protoOriginX && x+f.W > protoOriginX+o.packWidth {
			x, y, rowH = protoOriginX, y+rowH+protoGap, 0
		}
		f.X, f.Y = x, y
		rowH = protoMaxF(rowH, f.H)
		x += f.W + protoGap
		bottom = protoMaxF(bottom, y+f.H)
		right = protoMaxF(right, x-protoGap)
	}
	if !hasGeom || !packed {
		packTop = 0 // 没有单独的排布带，就不画带标题
	}
	return packTop, bottom, right
}

// protoTextSlots42F 选一个文字槽位把字段名放进去：优先 label（输入框的标签），
// 其次 value（单图元控件），再其次 field；都没有就按字典序取第一个（确定，不随机）。
func protoTextSlots42F(cat *Catalog, shape, label string) map[string]any {
	if label == "" || shape == "" {
		return nil
	}
	cs := cat.Find(shape)
	if cs == nil {
		return nil
	}
	for _, slot := range []string{"label", "value", "field"} {
		if _, ok := cs.Slots[slot]; ok {
			return map[string]any{slot: label}
		}
	}
	keys := make([]string, 0, len(cs.Slots))
	for k := range cs.Slots {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return nil
	}
	return map[string]any{keys[0]: label}
}

// outerBase42F 复述 `placeItem` 的三选一：外层格的 id 是 `t`（表格）、`g`（group）或 `2`。
//
// 这里必须与 `placeItem` 一致，否则 `CellID` 就是错的。它不是孤证：
// `TestProto42F_RealSample` 会拿每条 `CellID` 去产物里找，找不到就红。
func outerBase42F(cat *Catalog, shape string) string {
	cs := cat.Find(shape)
	if cs == nil {
		return "2"
	}
	switch {
	case strings.Contains(cs.XML, `id="t"`):
		return "t"
	case strings.Contains(cs.XML, `id="g"`):
		return "g"
	}
	return "2"
}

// protoTabs42F 把 RecordView 的链接翻成分组：tabName → 它链接到的 colName。
//
// 只报链接里写了的东西：链接里没有的字段一律不推测归属（A15）。
func protoTabs42F(ff *form42f.F42Form, fields []ProtoField) []ProtoTab {
	known := make(map[string]bool, len(fields))
	for i := range fields {
		known[fields[i].ColName] = true
	}
	out := make([]ProtoTab, 0, len(ff.RecordViews))
	for _, rv := range ff.RecordViews {
		t := ProtoTab{TabName: rv.TabName}
		seen := map[string]bool{}
		for _, l := range rv.Links {
			col := string(l.ColName)
			if seen[col] {
				continue
			}
			seen[col] = true
			if !known[col] {
				t.Missing = append(t.Missing, col)
				continue
			}
			t.ColNames = append(t.ColNames, col)
		}
		out = append(out, t)
	}
	return out
}

// unmappedKinds42F 列出未判定的控件名（去重、排序）。
func unmappedKinds42F(fields []ProtoField) []string {
	seen := map[string]bool{}
	var out []string
	for i := range fields {
		if fields[i].Mapped {
			continue
		}
		name := widgetName42F(fields[i].Widget)
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func widgetName42F(w string) string {
	if strings.TrimSpace(w) == "" {
		return "无控件元素"
	}
	return w
}

// protoBandTitle 是推断排布带的标题（画在带的上方）。
const protoBandTitle = "推断排布带：文件里没给网格坐标（posX/posY=0）的字段，按声明序排布 —— 宽度仍是文件里的真值"

// protoLegend42F 生成图例的全部行。与画在图上的文字**同一份**，不会漂。
func protoLegend42F(p *Proto, o protoOpts) []string {
	lines := []string{
		"图例 · 从 .42f 生成的表单原型",
		fmt.Sprintf("  表单 %s   字段 %d 条   跳过 %d 条   网格 1 单位 = %s px",
			p.Form, p.Count, p.Skipped, jsNum(o.cellPx)),
		"  形状按 .42f 控件名取形状库同名图形：" + mappedNames42F(),
		"  * 前缀 = 必填（required 或 notNull）；[只读] 后缀 = noEntry（不可输入）",
		"  ? 前缀 + 虚线灰框 = 控件族未判定（形状库里没有同名图形 —— A15：不猜，画通用矩形）",
		"  未判定控件：" + unmappedSummary42F(p),
		"  坐标：posX/posY 任一非 0 的字段用文件里的真实网格坐标；其余落在「推断排布带」，位置是推断的",
	}

	if len(p.Tabs) == 0 {
		lines = append(lines, "  RecordView：本表单没有 RecordView 分组")
	} else {
		lines = append(lines, "  RecordView（tabName → colName）：")
		for _, tab := range p.Tabs {
			line := fmt.Sprintf("    %s (%d): %s", tab.TabName, len(tab.ColNames), headList42F(tab.ColNames))
			if len(tab.Missing) > 0 {
				line += "   [链接指向了表单里没有的 colName: " + headList42F(tab.Missing) + "]"
			}
			lines = append(lines, line)
		}
	}

	orphan := 0
	for i := range p.Fields {
		if len(p.Fields[i].TabNames) == 0 {
			orphan++
		}
	}
	lines = append(lines, fmt.Sprintf("  未进任何 RecordView 的字段：%d 条（不推测归属）", orphan))
	if p.SizeMissing > 0 {
		lines = append(lines, fmt.Sprintf("  缺 width/gridWidth 的字段：%d 条（已按 1 个网格单位画，见字段的 WidthSource=min）", p.SizeMissing))
	}
	return lines
}

// mappedNames42F 把映射表按形状 id 排序拼成一行（图例用）。
func mappedNames42F() string {
	seen := map[string]bool{}
	names := make([]string, 0, len(WidgetShapes42F))
	for _, shape := range WidgetShapes42F {
		if seen[shape] {
			continue
		}
		seen[shape] = true
		names = append(names, shape)
	}
	sort.Strings(names)
	return strings.Join(names, " / ")
}

// unmappedSummary42F 报未判定控件的种类与条数 —— 只报数，不解释它们"应该"是什么。
func unmappedSummary42F(p *Proto) string {
	count := map[string]int{}
	for i := range p.Fields {
		if p.Fields[i].Mapped {
			continue
		}
		count[widgetName42F(p.Fields[i].Widget)]++
	}
	if len(count) == 0 {
		return "无"
	}
	names := make([]string, 0, len(count))
	for name := range count {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	total := 0
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%s×%d", name, count[name]))
		total += count[name]
	}
	return strings.Join(parts, "、") + fmt.Sprintf("（共 %d 条）", total)
}

// headList42F 截断长清单，免得一行图例被撑成一条竖线。
func headList42F(items []string) string {
	const max = 12
	if len(items) <= max {
		return strings.Join(items, ", ")
	}
	return strings.Join(items[:max], ", ") + fmt.Sprintf(" …共 %d 条", len(items))
}

// cell42F 拼一个单图元单元格（未判定矩形、图例文字都用它）。
func cell42F(id, value, style string, x, y, w, h float64) string {
	return fmt.Sprintf(`<mxCell id="%s" value="%s" style="%s" vertex="1" parent="1">`+
		`<mxGeometry x="%s" y="%s" width="%s" height="%s" as="geometry"/></mxCell>`,
		id, escapeAttr(value), escapeAttr(style),
		jsNum(x), jsNum(y), jsNum(w), jsNum(h))
}

// insertBeforeRootClose 把追加的单元格塞进 `</root>` 之前。
func insertBeforeRootClose(xmlStr, cells string) string {
	if cells == "" {
		return xmlStr
	}
	i := strings.LastIndex(xmlStr, "</root>")
	if i < 0 {
		return xmlStr
	}
	return xmlStr[:i] + cells + xmlStr[i:]
}

func protoMaxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
