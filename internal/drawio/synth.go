package drawio

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"path"
	"sort"
	"strings"
)

// 合成：形状源 → catalog。原实现是 build.js，这里逐条照搬它的规则与报错文案。

// Presets 是 presets.json：部件样式预设（kind → drawio 样式串）。
//
// 以 `_` 开头的键是文件里的说明条目（`_comment` / `_table` 之类），从不被 kind 引用，
// 加载时原样留着 —— 它们是可读的文档，不是需要过滤的噪音。
type Presets map[string]string

// Source 是形状源的一次加载结果：全部库 + 部件样式预设。
//
// 合成与展开（compose）都要用预设，所以两者一起加载、一起交出去 —— 分两次读
// 只会让调用方多写一行，还多一次"忘了加载预设"的机会。
type Source struct {
	Catalogs []Catalog
	Presets  Presets
	// Warnings 是形状源的告警（部件超出外层 group 的边界那类），不阻断合成。
	Warnings []string
}

// Load 读形状源、逐库校验并合成 catalog。
//
// 库按 library.json 的键排序（不依赖 map 迭代序，也不依赖目录扫描序 ——
// icons/ 与 raw/ 不是库，扫目录会把它们当成库）。
func Load() (*Source, error) {
	fsys := ShapesFS()

	presets, err := loadPresets(fsys)
	if err != nil {
		return nil, err
	}
	libs, err := loadLibraryConfigs(fsys)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(libs))
	for name := range libs {
		names = append(names, name)
	}
	sort.Strings(names)

	src := &Source{Presets: presets}
	for _, name := range names {
		shapes, warns, err := readLibrary(fsys, name, presets)
		if err != nil {
			return nil, err
		}
		src.Warnings = append(src.Warnings, warns...)
		src.Catalogs = append(src.Catalogs, Catalog{Name: name, Library: libs[name], Shapes: shapes})
	}
	return src, nil
}

// Catalog 按名字取一个库；没有则返回 nil。
func (s *Source) Catalog(name string) *Catalog {
	for i := range s.Catalogs {
		if s.Catalogs[i].Name == name {
			return &s.Catalogs[i]
		}
	}
	return nil
}

func loadPresets(fsys fs.FS) (Presets, error) {
	var p Presets
	if err := readJSONFile(fsys, "presets.json", &p); err != nil {
		return nil, err
	}
	return p, nil
}

func loadLibraryConfigs(fsys fs.FS) (map[string]LibraryConfig, error) {
	var m map[string]LibraryConfig
	if err := readJSONFile(fsys, "library.json", &m); err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("library.json 里一个库都没有")
	}
	return m, nil
}

// readLibrary 读 shapes/<库名>/ 下所有 .json（按文件名排序）并合成条目。
func readLibrary(fsys fs.FS, name string, presets Presets) ([]CatalogShape, []string, error) {
	dir := path.Join("", name)
	ents, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, nil, fmt.Errorf("读不到库目录 shapes/%s: %w", name, err)
	}

	files := make([]string, 0, len(ents))
	for _, e := range ents {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, nil, fmt.Errorf("库 shapes/%s 里没有任何 .json", name)
	}

	seen := map[string]string{} // id → 首次出现的位置（报重复 id 用）
	icons := newIconCache(fsys)
	var (
		out      []CatalogShape
		warnings []string
	)
	for _, file := range files {
		where := fmt.Sprintf("shapes/%s/%s", name, file)

		var shape Shape
		if err := readJSONFile(fsys, path.Join(name, file), &shape); err != nil {
			return nil, nil, fmt.Errorf("%s: JSON 解析失败 —— %v", where, err)
		}
		if err := validateShape(&shape, where, seen); err != nil {
			return nil, nil, err
		}

		xml, warns, err := resolveXML(fsys, icons, &shape, presets)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", where, err)
		}
		for _, w := range warns {
			warnings = append(warnings, where+": "+w)
		}

		out = append(out, catalogEntry(&shape, xml))
	}
	return out, warnings, nil
}

// validateShape 逐条校验一份组件定义。报错文案照抄原实现 —— 它是给人看的。
func validateShape(shape *Shape, where string, seen map[string]string) error {
	id := shape.Key()

	if shape.Title == "" {
		return fmt.Errorf("%s: 缺少 title（图形面板里显示的名字）", where)
	}
	for _, r := range id {
		if !(r == '.' || r == '-' || r == '_' || r >= '0' && r <= '9' ||
			r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return fmt.Errorf("%s: id %q 只能包含字母、数字、. - _", where, id)
		}
	}
	if prev, dup := seen[id]; dup {
		return fmt.Errorf("%s: id %q 与 %s 重复", where, id, prev)
	}
	seen[id] = where

	// 表格的尺寸由列宽与行数算出来，不用手写；列行数据的校验在 table.go（compose 共用）
	if shape.Table != nil {
		if err := validateTable(shape.Table, where); err != nil {
			return err
		}
		w, h := tableSize(shape.Table)
		shape.W, shape.H = &w, &h
	}

	for _, k := range []struct {
		name string
		v    *float64
	}{{"w", shape.W}, {"h", shape.H}} {
		if k.v == nil || !(*k.v > 0) {
			return fmt.Errorf("%s: %s 必须是正数，当前是 %v", where, k.name, deref(k.v))
		}
	}

	var modes []string
	if shape.Style != "" {
		modes = append(modes, "style")
	}
	if shape.Preset != "" {
		modes = append(modes, "preset")
	}
	if shape.Parts != nil {
		modes = append(modes, "parts")
	}
	if shape.Table != nil {
		modes = append(modes, "table")
	}
	if shape.XML != "" {
		modes = append(modes, "xml")
	}
	if shape.XMLFile != "" {
		modes = append(modes, "xmlFile")
	}
	if len(modes) > 1 {
		return fmt.Errorf("%s: style / preset / parts / table / xml / xmlFile 只能用一个，当前同时有 %s",
			where, strings.Join(modes, "、"))
	}

	for i, part := range shape.Parts {
		if part.Kind == "" && part.Style == "" {
			return fmt.Errorf("%s: parts[%d] 需要 kind 或 style", where, i)
		}
		// fitText 的部件宽度由文字算，不用写 w
		required := []struct {
			name string
			v    *float64
		}{{"x", part.X}, {"y", part.Y}, {"h", part.H}}
		if !part.FitText {
			required = append(required, struct {
				name string
				v    *float64
			}{"w", part.W})
		}
		for _, r := range required {
			if r.v == nil {
				return fmt.Errorf("%s: parts[%d] 缺少 %s", where, i, r.name)
			}
		}
	}

	// 槽位名撞车的话后者会静默盖掉前者，用的人还找不到原因，这里直接拦下
	seenSlots := map[string]int{}
	for i, part := range shape.Parts {
		if part.Text == nil {
			continue
		}
		slot := part.Slot
		if slot == "" {
			slot = part.Kind
		}
		if prev, dup := seenSlots[slot]; dup {
			return fmt.Errorf("%s: 文字槽位 %q 被 parts[%d] 和 parts[%d] 同时占用", where, slot, prev, i)
		}
		seenSlots[slot] = i
	}
	return nil
}

// resolveXML 把一份组件定义摊平成 mxGraphModel 片段。六种写法，优先级与原实现一致。
func resolveXML(fsys fs.FS, icons *iconCache, shape *Shape, presets Presets) (string, []string, error) {
	switch {
	case shape.Table != nil:
		return tableXML(shape.Table, presets), nil, nil
	case shape.Parts != nil:
		return modelFromParts(shape, presets, icons)
	case shape.XML != "":
		return shape.XML, nil, nil
	case shape.XMLFile != "":
		b, err := fs.ReadFile(fsys, path.Clean(shape.XMLFile))
		if err != nil {
			return "", nil, fmt.Errorf("读不到 %s: %w", shape.XMLFile, err)
		}
		return strings.TrimSpace(string(b)), nil, nil
	default:
		return modelFromStyle(shape, presets)
	}
}

// modelFromStyle 用一条样式串生成单节点的 mxGraphModel。
// 样式可以内联写，也可以 preset 引用 presets.json —— 后者让多个形状共用一份样式。
func modelFromStyle(shape *Shape, presets Presets) (string, []string, error) {
	base := shape.Style
	if base == "" {
		if shape.Preset != "" {
			v, ok := presets[shape.Preset]
			if !ok {
				return "", nil, fmt.Errorf("preset %q 在 presets.json 里不存在", shape.Preset)
			}
			base = v
		}
	}
	if base == "" {
		base = "rounded=1;whiteSpace=wrap;html=1;"
	}

	label := shape.Title
	if shape.Label != nil {
		label = *shape.Label
	}
	return `<mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/>` +
		fmt.Sprintf(`<mxCell id="2" value="%s" style="%s" vertex="1" parent="1">`,
			escapeAttr(label), escapeAttr(base)) +
		fmt.Sprintf(`<mxGeometry width="%s" height="%s" as="geometry"/></mxCell>`,
			jsNum(num(shape.W)), jsNum(num(shape.H))) +
		`</root></mxGraphModel>`, nil, nil
}

// modelFromParts 生成组合控件：外层一个 group 单元格 + 每个部件一个子单元格。
// 插入画布后整体拖动、整体缩放；要拆开右键 Ungroup。
func modelFromParts(shape *Shape, presets Presets, icons *iconCache) (string, []string, error) {
	cells := []string{
		fmt.Sprintf(`<mxCell id="g" value="" style="group" vertex="1" connectable="0" parent="1">`+
			`<mxGeometry width="%s" height="%s" as="geometry"/></mxCell>`,
			jsNum(num(shape.W)), jsNum(num(shape.H))),
	}

	var warnings []string

	for i, part := range shape.Parts {
		w, err := partWidth(icons, part)
		if err != nil {
			return "", nil, err
		}

		// 部件按未旋转的框算边界 —— 旋转部件（如图标的手柄）可能略微超出，只提示不报错
		if shape.WarnBounds == nil || *shape.WarnBounds {
			x, y, h := num(part.X), num(part.Y), num(part.H)
			if x < 0 || y < 0 || x+w > num(shape.W) || y+h > num(shape.H) {
				kind := part.Kind
				if kind == "" {
					kind = "style"
				}
				warnings = append(warnings, fmt.Sprintf(
					"parts[%d](%s) 超出 group %s×%s —— 该部件是 %s,%s %s×%s",
					i, kind, jsNum(num(shape.W)), jsNum(num(shape.H)),
					jsNum(x), jsNum(y), jsNum(w), jsNum(h)))
			}
		}

		value := ""
		if part.Text != nil {
			value = *part.Text
		}
		style, err := styleFor(icons, part, presets)
		if err != nil {
			return "", nil, err
		}

		cells = append(cells, fmt.Sprintf(
			`<mxCell id="p%d" value="%s" style="%s" vertex="1" parent="g">`+
				`<mxGeometry x="%s" y="%s" width="%s" height="%s" as="geometry"/></mxCell>`,
			i+1, escapeAttr(value), escapeAttr(style),
			jsNum(num(part.X)), jsNum(num(part.Y)), jsNum(w), jsNum(num(part.H))))
	}

	return `<mxGraphModel><root><mxCell id="0"/><mxCell id="1" parent="0"/>` +
		strings.Join(cells, "") + `</root></mxGraphModel>`, warnings, nil
}

// styleFor 拼一个部件的样式串：预设（或内联覆盖）+ 旋转 + 图标。
func styleFor(icons *iconCache, part Part, presets Presets) (string, error) {
	base := part.Style
	if base == "" {
		v, ok := presets[part.Kind]
		if !ok {
			return "", fmt.Errorf("未知的 kind %q —— 在 presets.json 里加一条，或在部件上直接写 style", part.Kind)
		}
		base = v
	}
	style := base
	if !strings.HasSuffix(style, ";") {
		style += ";"
	}
	if part.Rotation != nil && *part.Rotation != 0 {
		style += "rotation=" + jsNum(*part.Rotation) + ";"
	}
	if part.Icon != "" {
		uri, err := icons.uri(part.Icon)
		if err != nil {
			return "", err
		}
		style += "image=" + uri + ";imageAspect=0;"
	}
	return style, nil
}

// partWidth 算一个部件的宽度。声明 fitText 的部件宽度由文字决定（只用来给
// "宽度跟着文字走"的部件一个合理初值），其余用写死的 w。
func partWidth(icons *iconCache, part Part) (float64, error) {
	if !part.FitText {
		return num(part.W), nil
	}
	fontSize := 11.0
	if part.FontSize != nil {
		fontSize = *part.FontSize
	}
	padX := 10.0
	if part.PadX != nil {
		padX = *part.PadX
	}
	text := ""
	if part.Text != nil {
		text = *part.Text
	}
	return estimateTextWidth(text, fontSize) + padX, nil
}

// estimateTextWidth 粗估一行文字的宽度。CJK 按 fontSize+1，其余按 0.62×fontSize ——
// 不精确，但比写死宽度强。
func estimateTextWidth(text string, fontSize float64) float64 {
	var w float64
	for _, r := range text {
		if r > 0x2e80 {
			w += fontSize + 1
		} else {
			w += fontSize * 0.62
		}
	}
	return math.Ceil(w)
}

// catalogEntry 把一份定义 + 它展开好的 XML 变成 catalog 条目。
func catalogEntry(shape *Shape, xml string) CatalogShape {
	entry := CatalogShape{
		ID:    shape.Key(),
		Title: shape.Title,
		W:     num(shape.W),
		H:     num(shape.H),
		Tags:   shape.Tags,
		Slots:  textSlots(shape),
		XML:    xml,
		Aspect: shape.Aspect,
	}
	if entry.Tags == nil {
		entry.Tags = []string{}
	}
	if shape.Table != nil {
		entry.Table = shape.Table
	}
	// pfill / pdeco 是部件序号（1 基），描述拉伸时各子格的行为。
	for i, p := range shape.Parts {
		if p.Fill {
			entry.PFill = append(entry.PFill, i+1)
		}
		if p.Deco {
			entry.PDeco = append(entry.PDeco, i+1)
		}
	}
	entry.MultiTab = shape.MultiTab
	return entry
}

// textSlots 说明组件里哪些文字能被替换、对应哪个单元格 id。
//
// slot 让"槽位名"与"样式预设名"解耦：输入框的预设可能叫 fieldRight / textArea，
// 但对外一律叫 field，写 spec 的人不用记每个控件各自的叫法。
func textSlots(shape *Shape) map[string]string {
	slots := map[string]string{}
	switch {
	case shape.Table != nil:
		return tableSlots(shape.Table)
	case shape.Parts != nil:
		for i, part := range shape.Parts {
			if part.Text == nil {
				continue
			}
			slot := part.Slot
			if slot == "" {
				slot = part.Kind
			}
			slots[slot] = fmt.Sprintf("p%d", i+1)
		}
		return slots
	case shape.XML != "" || shape.XMLFile != "":
		// 原始 XML 的组件结构不固定，不做文字替换
		return slots
	default:
		return map[string]string{"value": "2"}
	}
}

// MxLibrary 把一批条目渲染成 drawio 能加载的 <mxlibrary>。
//
// **顺序不能颠倒**：先 JSON 序列化（把 " 写成 \"），再把整段 JSON 当 XML 文本节点转义
// （只转 & < >）。少了第一层，JSON 里的引号会提前结束 XML 属性；少了第二层，
// JSON 里的 < 会被 XML 解析器当成标签开头 —— 两种情况 drawio 都直接打不开。
func MxLibrary(entries []MxEntry) (string, error) {
	body, err := MarshalJSON(entries)
	if err != nil {
		return "", err
	}
	return "<mxlibrary>" + escapeText(body) + "</mxlibrary>", nil
}

// deref 只为报错文案服务：把 `*float64(nil)` 打成 `null`（与 JS 的 JSON.stringify 一致）。
func deref(v *float64) string {
	if v == nil {
		return "null"
	}
	return jsNum(*v)
}

func lowerTrim(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// firstWord 取 title 的第一个空格分词（"ButtonEdit 编辑开窗" → "ButtonEdit"）。
func firstWord(s string) string {
	if i := strings.IndexByte(s, ' '); i >= 0 {
		return strings.ToLower(strings.TrimSpace(s[:i]))
	}
	return lowerTrim(s)
}

// readJSONFile 读 fs 里的一个 JSON 文件。报错带上文件名 —— 形状源是数据，
// "哪个文件坏了"是唯一有用的信息。
func readJSONFile(fsys fs.FS, name string, v any) error {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("读不到 %s: %w", name, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: JSON 解析失败 —— %v", name, err)
	}
	return nil
}
