package drawio

// 形状源的数据面：library.json 的库配置、shapes/<库>/*.json 的组件定义，
// 以及由它们合成出来的 catalog（给 compose 用的机器可读清单）。
//
// 字段用指针的地方都是**区分"没写"与"写了零值"**：部件坐标写 0 是合法的
// （贴着左上角），文字写成空串也是合法的（一个空标签），所以不能拿零值当缺省。

// LibraryConfig 是 library.json 里一条库配置：目录名（= 库名）→ 展示名 / 版本 / 输出文件名。
type LibraryConfig struct {
	Name        string `json:"name"`
	Output      string `json:"output"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// Shape 是 shapes/<库名>/*.json 里的一份组件定义。
//
// 六种写法只能用一个（style / preset / parts / table / xml / xmlFile），
// 由 validate 拦下 —— 同时写两种在渲染时是静默取其一，那种 bug 最难查。
type Shape struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	// Label 是单图元控件真正显示的文字；不写就用 Title（原实现是 `shape.label ?? shape.title`）。
	Label      *string   `json:"label"`
	Tags       []string  `json:"tags"`
	Aspect     string    `json:"aspect"`
	W          *float64  `json:"w"`
	H          *float64  `json:"h"`
	WarnBounds *bool     `json:"warnBounds"`
	Style      string    `json:"style"`
	Preset     string    `json:"preset"`
	Parts      []Part    `json:"parts"`
	Table      *Table    `json:"table"`
	XML        string    `json:"xml"`
	XMLFile    string    `json:"xmlFile"`
	MultiTab   *MultiTab `json:"multiTab"`
}

// Key 是组件在图形面板里的 id；没写 id 时退回 title（原实现是 `shape.id ?? shape.title`）。
func (s *Shape) Key() string {
	if s.ID != "" {
		return s.ID
	}
	return s.Title
}

// Part 是组合控件里的一个部件 —— 一个矩形图元，坐标为**相对外层 group**。
type Part struct {
	Kind string   `json:"kind"`
	Style string  `json:"style"`
	X    *float64 `json:"x"`
	Y    *float64 `json:"y"`
	W    *float64 `json:"w"`
	H    *float64 `json:"h"`

	// Text 为 nil = 这个部件不参与文字替换，也不占用槽位；指向空串则是"占一个空槽位"。
	Text *string `json:"text"`
	// Slot 让槽位名与样式预设名解耦：输入框的预设可能叫 fieldRight / textArea，
	// 但对外一律叫 field，写 spec 的人不用记每个控件各自的叫法。缺省取 Kind。
	Slot string `json:"slot"`

	// FitText：宽度由文字算（Group 的图例这种"宽度跟着文字走"的部件），此时 w 可以不写。
	FitText  bool     `json:"fitText"`
	PadX     *float64 `json:"padX"`
	FontSize *float64 `json:"fontSize"`
	Rotation *float64 `json:"rotation"`
	Icon     string   `json:"icon"`

	// Fill / Deco 描述拉伸时这个子格怎么办（见 compose 的三类子格行为）：
	// fill = 框体填满外层（边框跟着动），deco = 装饰件保形（页签、图例不被拉变形）。
	Fill bool `json:"fill"`
	Deco bool `json:"deco"`
}

// Table 是表格数据。build 期的 040-table.json 与 spec 里的 items[].table 是同一形状，
// 生成规则共用 tableXML —— 两处各写一份必然漂。
type Table struct {
	HeaderHeight *float64      `json:"headerHeight"`
	RowHeight    *float64      `json:"rowHeight"`
	Columns      []TableColumn `json:"columns"`
	// Rows 是 []any（每行是 []any）或 float64（数字简写：N = N 行全空格，列数跟 Columns 走）。
	Rows any `json:"rows"`
}

// TableColumn 是一列：列头文字 + 列宽（像素）。
type TableColumn struct {
	Title string  `json:"title"`
	W     float64 `json:"w"`
}

// MultiTab 是 Folder 的多页签参数（只有 Folder 有）。
//
// Part 是**部件序号（1 基）**，指页签条那个部件；页签名由 spec 的 pages 给。
type MultiTab struct {
	Part  int     `json:"part"`
	W     float64 `json:"w"`
	H     float64 `json:"h"`
	XStep float64 `json:"xStep"`
}

// CatalogShape 是给 compose 用的机器可读条目。
//
// 与原实现一致：Slots 是「文字槽位 → 单元格 id」，XML 是**已经展开好的** mxGraphModel
// 片段（组合控件在合成期就摊平，compose 只负责换字、重编号、摆位置与拉伸）。
type CatalogShape struct {
	ID     string            `json:"id"`
	Title  string            `json:"title"`
	W      float64           `json:"w"`
	H      float64           `json:"h"`
	Tags   []string          `json:"tags"`
	Slots  map[string]string `json:"slots"`
	XML    string            `json:"xml"`
	Table  *Table            `json:"table,omitempty"`
	PFill  []int             `json:"pfill,omitempty"`
	PDeco  []int             `json:"pdeco,omitempty"`
	MultiTab *MultiTab       `json:"multiTab,omitempty"`

	// Aspect 只在合成 mxlibrary 时用得上，**不进 catalog JSON** —— 原实现的 catalog
	// 里没有这个字段，多写一个就不是等价移植了。
	Aspect string `json:"-"`
}

// Catalog 是一个库的全部条目 + 它的库配置。
type Catalog struct {
	// Name 是 library.json 里的键，也就是 shapes/ 下的目录名。
	// 只有渲染形状清单时报得出来"这是哪个库"，catalog JSON 里不出现。
	Name    string
	Library LibraryConfig
	Shapes  []CatalogShape
}

// Find 按 id、完整 title、或 title 里的英文名（"ButtonEdit 编辑开窗" → "ButtonEdit"）找一个形状。
//
// 三条都找不到时返回 nil —— 调用方负责给出可读的报错（含旧名对照）。
func (c *Catalog) Find(name string) *CatalogShape {
	norm := func(s string) string { return lowerTrim(s) }
	want := norm(name)
	for i := range c.Shapes {
		if norm(c.Shapes[i].ID) == want {
			return &c.Shapes[i]
		}
	}
	for i := range c.Shapes {
		if norm(c.Shapes[i].Title) == want {
			return &c.Shapes[i]
		}
	}
	for i := range c.Shapes {
		if firstWord(c.Shapes[i].Title) == want {
			return &c.Shapes[i]
		}
	}
	return nil
}

// MxEntry 是 mxlibrary 里的一条 —— drawio 只认这几个字段，多写的会被忽略。
//
// 字段顺序就是输出顺序（json 编码器按结构体字段序），与原实现的
// `{xml, w, h, title, aspect}` 再补 `tags` 一致。
type MxEntry struct {
	XML    string   `json:"xml"`
	W      float64  `json:"w"`
	H      float64  `json:"h"`
	Title  string   `json:"title"`
	Aspect string   `json:"aspect"`
	Tags   []string `json:"tags,omitempty"`
}
