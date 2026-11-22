package devdata

// 蓝图 ↔ 代码 的三份输入解析：.tab（蓝图）、.tap（骨架 + 加点声明）、.4gl（产物）。
//
// 这里只做"把文件读成结构"，不判断语义；判断在 link.go。
// 三个解析器都要能容忍真机上的畸形数据 —— 已知的几种（都实测于 axmt500_wf）：
//
//  1. `.tap` 里的 `#add-point:… name="…"` 可能被**截断**（值里带了换行），
//     此时正确的名字在那行下面的 `{<point name="…"/>}` 洞里 —— 所以洞才是权威。
//  2. `.4gl` 里有些槽位被**整行注释掉**（行首是 `#` + 空白）。
//  3. `.tap` 里存在的洞，`.4gl` 里可能**一个都没有**（4gl 比 tap 旧，或该段被生成器丢弃）。
//
// 所以：槽位的身份以"洞/标记行的名字 + 出现次序"为准，位置以行号为准，两者不互相推断。

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// ---------- 锚点正则（.4gl / 切片） ----------

var (
	reSectionOpen     = regexp.MustCompile(`\{<section id="([^"]+)"`)
	reSectionClose    = regexp.MustCompile(`\{</section>\}`)
	reAddPoint        = regexp.MustCompile(`#add-point:.*?name="([^"]*)`)
	reAddPointEnd     = regexp.MustCompile(`^\s*#+\s*(?i:end) add-point`)
	reCommentedMarker = regexp.MustCompile(`^\s*#\s+#add-point:`)
	reSliceMark       = regexp.MustCompile(`#應用\s+([A-Za-z]+\d+)\s+樣板自動產生\(Version:(\d+)\)`)
)

// ---------- .tap ----------

// tapSlot 是骨架里的一个注入点（一个洞）。
type tapSlot struct {
	Name      string `json:"name"`
	Label     string `json:"label,omitempty"`     // `#add-point:` 后面的人类可读标签
	Commentd  bool   `json:"commented,omitempty"` // 标记行被注释掉
	Line      int    `json:"line"`                // 洞在 .tap 里的行号
	MarkerRaw string `json:"-"`                   // 标记行原文（排查截断用）
}

// tapDecl 是 `<point name=…>内容</point>` 声明：槽位要填的东西。
type tapDecl struct {
	Name     string `json:"name"`
	Slot     string `json:"slot,omitempty"` // mapping：槽位名（客制点可挂到标准槽位上）
	Src      string `json:"src,omitempty"`  // s=标准  c=客制
	Ver      string `json:"ver,omitempty"`
	Status   string `json:"status,omitempty"`
	New      string `json:"new,omitempty"`
	CiteStd  string `json:"citeStd,omitempty"`
	Blank    bool   `json:"blank,omitempty"` // 声明了但内容只有空白
	Lines    int    `json:"lines"`           // 非空行数
	Bytes    int    `json:"bytes"`
	FirstUse string `json:"firstUse,omitempty"` // 内容第一行非空文本（人读摘要）
	// Line 是 `<point` 标签在 .tap 里的行号；Content 是声明正文（原样）。
	// 这两项是给"孤儿点"用的：它们的正文只存在于 .tap，.4gl 里一个字都没有。
	Line        int    `json:"line,omitempty"`
	ContentLine int    `json:"contentLine,omitempty"`
	Content     string `json:"-"`
}

// tapSection 是 `.tap` 里的一个 section 及其中按顺序出现的槽位。
type tapSection struct {
	ID     string    `json:"id"`
	Ver    string    `json:"ver,omitempty"`
	Src    string    `json:"src,omitempty"`
	Type   string    `json:"type,omitempty"`
	Readon string    `json:"readonly,omitempty"`
	Line   int       `json:"line"`
	Slots  []tapSlot `json:"slots"`
}

// tapDoc 是一整份 `.tap`。
type tapDoc struct {
	Attrs    map[string]string `json:"attrs"`
	Sections []tapSection      `json:"sections"`
	Decls    []tapDecl         `json:"decls"`
}

// parseTap 解析 `.tap`。
//
// 用流式 decoder 而不是正则整篇扫：`<point>` 的内容是 CDATA，正文里同样可能出现
// `#add-point:` 与 `{<point/>}` 字样，正则会把"内容里的字"当成"骨架里的洞"。
func parseTap(data []byte) (*tapDoc, error) {
	doc := &tapDoc{Attrs: map[string]string{}}
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false // 设计器产出的文件里偶有非严格字符；结构错误仍会报

	lineStarts := newLineIndex(data)
	var (
		inPoint     bool
		curAttrs    map[string]string
		curLine     int
		contentLine int
		pointText   strings.Builder
		inSection   bool
		curSec      tapSection
		body        strings.Builder
		bodyLine    int // section 正文第一个字符所在行（CDATA 紧跟标签时与标签同行）
		bodyBegin   bool
	)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("解析 .tap 失败（偏移 %d）: %w", dec.InputOffset(), err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "add_points":
				for _, a := range t.Attr {
					doc.Attrs[a.Name.Local] = a.Value
				}
			case "point":
				// 自闭合 = 洞（在 section 正文里，作为文本出现，不会走到这里）；
				// 带内容 = 声明。
				inPoint = true
				curLine = lineStarts.lineOf(dec.InputOffset())
				contentLine = 0
				curAttrs = make(map[string]string, len(t.Attr))
				for _, a := range t.Attr {
					curAttrs[a.Name.Local] = a.Value
				}
				pointText.Reset()
			case "section":
				inSection = true
				curSec = tapSection{Line: lineStarts.lineOf(dec.InputOffset())}
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "id":
						curSec.ID = a.Value
					case "ver":
						curSec.Ver = a.Value
					case "src":
						curSec.Src = a.Value
					case "type":
						curSec.Type = a.Value
					case "readonly":
						curSec.Readon = a.Value
					}
				}
				body.Reset()
				bodyBegin = false
				bodyLine = 0
			}
		case xml.CharData:
			switch {
			case inSection:
				if !bodyBegin {
					// 用"正文首字符的偏移"定行号，而不是 section 标签的行号：
					// 设计器的写法有两种（CDATA 紧跟 `>` 或换一行），差一行会让所有行号全偏。
					bodyLine = lineStarts.lineOf(dec.InputOffset() - int64(len(t)))
					bodyBegin = true
				}
				body.Write(t)
			case inPoint:
				if contentLine == 0 {
					contentLine = lineStarts.lineOf(dec.InputOffset() - int64(len(t)))
				}
				pointText.Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "point":
				if inPoint {
					txt := pointText.String()
					if strings.TrimSpace(txt) != "" {
						d := mkDecl(curAttrs, txt)
						d.Line, d.ContentLine, d.Content = curLine, contentLine, txt
						doc.Decls = append(doc.Decls, d)
					}
					inPoint = false
					curAttrs = nil
				}
			case "section":
				if inSection {
					curSec.Slots = parseTapSlots(body.String(), bodyLine)
					doc.Sections = append(doc.Sections, curSec)
					inSection = false
				}
			}
		}
	}
	return doc, nil
}

// mkDecl 把 <point> 的属性与内容整理成 tapDecl。
func mkDecl(attrs map[string]string, text string) tapDecl {
	d := tapDecl{
		Name:    attrs["name"],
		Slot:    attrs["mapping"],
		Src:     attrs["src"],
		Ver:     attrs["ver"],
		Status:  attrs["status"],
		New:     attrs["new"],
		CiteStd: attrs["cite_std"],
		Bytes:   len(text),
	}
	for _, ln := range strings.Split(text, "\n") {
		s := strings.TrimSpace(ln)
		if s == "" {
			continue
		}
		d.Lines++
		// FirstUse 只当"这段在干什么"的一行摘要，所以跳过注释行：
		// 客制函数开头常是一整片 `########` 横幅，取第一行等于没取。
		if d.FirstUse == "" && !strings.HasPrefix(s, "#") {
			d.FirstUse = s
		}
	}
	d.Blank = d.Lines == 0
	return d
}

// parseTapSlots 从 section 正文里按顺序取出槽位。
//
// 名字取洞（标记行的 name 可能被截断），标签与"是否被注释掉"取标记行。
// baseLine 是 `<section …>` 所在行号（1 起），正文第 0 行即它的下一行。
func parseTapSlots(body string, baseLine int) []tapSlot {
	lines := strings.Split(body, "\n")
	var (
		out         []tapSlot
		pendingLbl  string
		pendingCut  bool
		pendingLine int
	)
	reset := func() { pendingLbl, pendingCut, pendingLine = "", false, 0 }
	for i, ln := range lines {
		if m := reMarkerLabel.FindStringSubmatch(ln); m != nil {
			pendingLbl = strings.TrimSpace(m[1])
			pendingCut = reCommentedMarker.MatchString(ln)
			pendingLine = baseLine + i // 标记行（人 grep 的是它，不是洞）
		}
		if idx := strings.Index(ln, `{<point name="`); idx >= 0 {
			rest := ln[idx+len(`{<point name="`):]
			name := rest
			if q := strings.Index(name, `"`); q >= 0 {
				name = name[:q]
			}
			if name = strings.TrimSpace(name); name != "" {
				line := baseLine + i
				if pendingLine > 0 {
					line = pendingLine
				}
				out = append(out, tapSlot{
					Name: name, Label: pendingLbl, Commentd: pendingCut,
					Line: line, MarkerRaw: strings.TrimSpace(ln),
				})
			}
			reset()
		}
	}
	return out
}

// reMarkerLabel 取 `#add-point:<label> name="…"` 里的 label（容忍 name 被截断）。
var reMarkerLabel = regexp.MustCompile(`#add-point:(.*?)(?:\s+name="|$)`)

// ---------- .4gl ----------

// g4Slot 是产物里的一个槽位（`#add-point:` 行到 `#END add-point` 行）。
type g4Slot struct {
	Name      string `json:"name"`
	Line0     int    `json:"line0"`
	Line1     int    `json:"line1"`
	Commentd  bool   `json:"commented,omitempty"` // 整行被注释掉（生成器有意为之）
	HasBody   bool   `json:"hasBody,omitempty"`
	SectionID string `json:"-"`
}

// g4Section 是产物里的一个 section 及其槽位。
type g4Section struct {
	ID    string   `json:"id"`
	Line0 int      `json:"line0"`
	Line1 int      `json:"line1"`
	Slots []g4Slot `json:"slots"`
}

// g4Slice 是一条 `#應用 aXX 樣板自動產生(Version:N)`。
type g4Slice struct {
	Code string `json:"code"`
	Ver  int    `json:"ver"`
	Line int    `json:"line"`
}

// g4Doc 是一整份 `.4gl` 的锚点视图。
type g4Doc struct {
	TotalLines      int         `json:"totalLines"`
	Freestyle       bool        `json:"freestyle"`       // 非生成器产物（无 section 锚且没声明"非freestyle"）
	UnfoldedSection bool        `json:"unfoldedSection"` // 首行自述"未開Section"
	Sections        []g4Section `json:"sections"`
	Slices          []g4Slice   `json:"slices"`

	Lines []string `json:"-"` // 原文（按函数名定位客制函数时要用）
}

// parse4gl 扫描 `.4gl` 的锚点。逐行扫：要行号区间，且要区分"注释掉的槽位"。
func parse4gl(lines []string) *g4Doc {
	doc := &g4Doc{TotalLines: len(lines), Lines: lines}
	saysGenerated := false
	head := lines
	if len(head) > 5 {
		head = head[:5]
	}
	for _, ln := range head {
		if strings.Contains(ln, "未解開Section") {
			doc.UnfoldedSection = true
		}
		if strings.Contains(ln, "非freestyle程式") {
			saysGenerated = true
		}
	}

	var cur *g4Section
	// 同名 section 可能分几次出现（同一段的槽位被拆开写）；按 id 合并，保持首次出现的位置。
	idx := map[string]int{}
	closeCur := func(end int) {
		if cur == nil {
			return
		}
		cur.Line1 = end
		if i, ok := idx[cur.ID]; ok {
			doc.Sections[i].Slots = append(doc.Sections[i].Slots, cur.Slots...)
			if cur.Line1 > doc.Sections[i].Line1 {
				doc.Sections[i].Line1 = cur.Line1
			}
		} else {
			idx[cur.ID] = len(doc.Sections)
			doc.Sections = append(doc.Sections, *cur)
		}
		cur = nil
	}
	for i, ln := range lines {
		if m := reSectionOpen.FindStringSubmatch(ln); m != nil {
			closeCur(i) // 上一个没闭合：补到本行之前
			cur = &g4Section{ID: m[1], Line0: i + 1}
			continue
		}
		if reSectionClose.MatchString(ln) {
			closeCur(i + 1)
			continue
		}
		if cur == nil {
			continue
		}
		if m := reAddPoint.FindStringSubmatch(ln); m != nil {
			s := g4Slot{
				Name: m[1], Line0: i + 1, Line1: i + 1,
				Commentd: reCommentedMarker.MatchString(ln), SectionID: cur.ID,
			}
			for j := i + 1; j < len(lines) && j < i+4000; j++ {
				if reAddPointEnd.MatchString(lines[j]) {
					s.Line1 = j + 1
					break
				}
				if strings.TrimSpace(lines[j]) != "" {
					s.HasBody = true
				}
			}
			cur.Slots = append(cur.Slots, s)
		}
	}
	closeCur(len(lines))
	_ = idx

	for i, ln := range lines {
		if m := reSliceMark.FindStringSubmatch(ln); m != nil {
			v, _ := strconv.Atoi(m[2])
			doc.Slices = append(doc.Slices, g4Slice{Code: m[1], Ver: v, Line: i + 1})
		}
	}
	doc.Freestyle = len(doc.Sections) == 0 && !doc.UnfoldedSection && !saysGenerated
	return doc
}

// ---------- .tab ----------

// tabAssembly 是 `<assembly>` 的属性（程序级：版型、页数、标准/客制版次）。
type tabAssembly struct {
	Prog     string `json:"prog"`
	Module   string `json:"module"`
	Type     string `json:"type"`     // t01 / q01 / r01 / p01 …（主模板 = code_<type>.template）
	Industry string `json:"industry"` // std / _wf …
	Page     string `json:"page"`     // 页签数
	SDVer    string `json:"sdver"`
	CusVer   string `json:"cusver"` // 0000 = 从未客制
	TabVer   string `json:"tabver"`
	Desc     string `json:"description"`
}

// tabPage 是蓝图里的一个页面（body）：page 是第几个页签，pageID 是稳定 id（点名用它）。
type tabPage struct {
	Page    string   `json:"page"`
	PageID  string   `json:"pageId"`
	Record  string   `json:"record"`
	Columns []string `json:"columns"`
}

// tabDoc 是一整份 `.tab`。
type tabDoc struct {
	Assembly tabAssembly `json:"assembly"`
	Pages    []tabPage   `json:"pages"`
	Head     []string    `json:"head,omitempty"` // 单头字段（不带 page，无法直接点名）
}

// parseTab 解析 `.tab`：只需要 assembly 属性与每个 body 的字段清单。
//
// **不能用 encoding/xml**：真机上 .tab 里有 `<32wc …>` 这类以数字开头的标签名，
// 严格 XML 解析器会在那里报 invalid XML name。我们只要两类元素的属性，自己扫更稳。
func parseTab(data []byte) (*tabDoc, error) {
	doc := &tabDoc{}
	if asm := scanTagAttrs(data, "assembly"); len(asm) > 0 {
		a := asm[0]
		doc.Assembly = tabAssembly{
			Prog: a["name"], Module: a["module"], Type: a["type"], Industry: a["industry"],
			Page: a["page"], SDVer: a["sdver"], CusVer: a["cusver"], TabVer: a["tabver"],
			Desc: a["description"],
		}
	}
	for _, a := range scanTagAttrs(data, "var") {
		cols := splitTabColumns(a["value"])
		switch a["id"] {
		case "body":
			doc.Pages = append(doc.Pages, tabPage{
				Page: a["page"], PageID: a["page_id"], Record: a["record"], Columns: cols,
			})
		case "head":
			doc.Head = cols
		}
	}
	return doc, nil
}

// scanTagAttrs 扫出所有 `<tag …>` 的属性表（容忍标签名非法、容忍属性里有单/双引号）。
func scanTagAttrs(data []byte, tag string) []map[string]string {
	needle := []byte("<" + tag)
	var out []map[string]string
	for i := 0; i+len(needle) < len(data); {
		j := bytes.Index(data[i:], needle)
		if j < 0 {
			break
		}
		j += i
		k := j + len(needle)
		if k < len(data) {
			switch data[k] {
			case ' ', '\t', '\r', '\n', '>', '/':
			default:
				i = k
				continue
			}
		}
		// 扫到闭合的 '>'，跳过引号里的内容
		attrs, end := parseAttrsAt(data, k)
		if end < 0 {
			break
		}
		out = append(out, attrs)
		i = end
	}
	return out
}

// parseAttrsAt 从 k（标签名之后）扫到闭合 `>`，返回属性表与 `>` 之后的下标。
// 没有闭合 `>` 时返回 -1（截断的标签，调用方应当停在这里）。
func parseAttrsAt(data []byte, k int) (map[string]string, int) {
	m := k
	var quote byte
	for m < len(data) {
		c := data[m]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
		} else if c == '"' || c == '\'' {
			quote = c
		} else if c == '>' {
			break
		}
		m++
	}
	if m >= len(data) {
		return nil, -1
	}
	return parseAttrPairs(data[k:m]), m + 1
}

var reAttrPair = regexp.MustCompile(`([A-Za-z_:][-A-Za-z0-9_:.]*)\s*=\s*(?:"([^"]*)"|'([^']*)')`)

// parseAttrPairs 从一段标签正文里取 key=value。
func parseAttrPairs(b []byte) map[string]string {
	m := map[string]string{}
	for _, g := range reAttrPair.FindAllSubmatch(b, -1) {
		v := g[2]
		if len(v) == 0 {
			v = g[3]
		}
		m[string(g[1])] = string(v)
	}
	return m
}

// splitTabColumns 把 `<var value="a,b(imaal_t.imaal003),c(chr80)">` 拆成字段名。
// 括号里是展延来源或类型，不是字段名，要去掉。
func splitTabColumns(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(value, ",") {
		name := part
		if i := strings.IndexByte(name, '('); i >= 0 {
			name = name[:i]
		}
		if name = strings.TrimSpace(name); name != "" {
			out = append(out, name)
		}
	}
	return out
}

// ---------- .tsd / .4fd（画面侧：字段标签、页签标题） ----------

// tsdSpec 是从 .tsd 里取到的画面文案。
type tsdSpec struct {
	// Fields 字段名 → 中文名（`lbl_<字段>`）。
	Fields map[string]string
	// Pages 页签键 → 中文名（`page_13` → 示例页签）。
	Pages map[string]string
}

// parseTsd 从 .tsd 取「字段名 → 画面中文名」与「页签键 → 页签标题」。
//
// 依据：`.tsd` 的 `<strings>` 段里，标签叫 `lbl_<字段名>`、页签叫 `page_<n>`，例如
//
//	<sfield name="lbl_xmdduc001" text="料号" />
//	<sfield name="page_13" text="示例页签" />
//
// 只认这两种前缀：其余 sfield 是按钮/其它文案，按名字硬配会错配。
// 页签键本身（`page_13`）**不与 .tab 的页序对应** —— 它由 .4fd 的 `<Table name>` 桥接
// （见 parse4fd），不要拿序号去猜。
func parseTsd(data []byte) tsdSpec {
	spec := tsdSpec{Fields: map[string]string{}, Pages: map[string]string{}}
	for _, a := range scanTagAttrs(data, "sfield") {
		name, text := strings.TrimSpace(a["name"]), strings.TrimSpace(a["text"])
		if text == "" {
			continue
		}
		switch {
		case strings.HasPrefix(name, "lbl_"):
			field := strings.TrimPrefix(name, "lbl_")
			if _, dup := spec.Fields[field]; !dup {
				spec.Fields[field] = text
			}
		case strings.Contains(name, "page_"):
			// 键有四种形态：page_13 / bpage_1（browse 页）/ axmt500_wf.page_10 /
			// axmt500.page_7（带程序名前缀）。只认 `page_` 会漏掉 bpage_*，
			// 而那 6 个恰好是明细页（订单明细/交期明细/…）。
			if _, dup := spec.Pages[name]; !dup {
				spec.Pages[name] = text
			}
		}
	}
	return spec
}

// fdSpec 是从 .4fd（表单定义）里取到的页面骨架。
type fdSpec struct {
	// PageByRecord Screen Record（`s_detail73`）→ 页签键（`page_13`）。
	PageByRecord map[string]string
}

// parse4fd 解析 .4fd 的 `<Page>` / `<Table>` 结构。
//
// 为什么必须走 .4fd：`.tab` 只有 `page="12" page_id="73" record="s_detail73"`，
// `.tsd` 只有 `page_13 → 示例页签`，两边**没有共同的键**。`.4fd` 是唯一同时含
// `<Page>` 与 `<Table name="s_detail73">` 的地方。
//
// 页签键取 **text 属性**（不是 name）：实测 `name` 是控件 id，与文案表对不上 ——
// 备注页是 `name="page_1" text="page_12"`，取 name 会把它接成"基本数据"；主档页是
// `name="master_Folder_page" text="page_1"`。text 才是 tsd 里 `page_*` 的键。
//
// 归属规则：一个 `<Table>` 属于**最近打开且尚未关闭**的那个 `<Page>`（按 `</Page>`
// 回退）。这样嵌套页（模板里 page_info 套在 page_1 里）也能把 grid 归到内层。
func parse4fd(data []byte) *fdSpec {
	spec := &fdSpec{PageByRecord: map[string]string{}}
	var stack []string
	for i := 0; ; {
		pi := bytes.Index(data[i:], []byte("<Page"))
		if pi >= 0 {
			pi += i
		}
		ci := bytes.Index(data[i:], []byte("</Page"))
		if ci >= 0 {
			ci += i
		}
		ti := bytes.Index(data[i:], []byte("<Table"))
		if ti >= 0 {
			ti += i
		}
		next, kind := -1, ""
		for _, cand := range []struct {
			at   int
			kind string
		}{{pi, "page"}, {ci, "close"}, {ti, "table"}} {
			if cand.at < 0 {
				continue
			}
			if next < 0 || cand.at < next {
				next, kind = cand.at, cand.kind
			}
		}
		if next < 0 {
			break
		}
		switch kind {
		case "page", "table":
			open := "<Page"
			if kind == "table" {
				open = "<Table"
			}
			k := next + len(open)
			if k >= len(data) || !isTagNameEnd(data[k]) {
				i = k
				continue
			}
			attrs, end := parseAttrsAt(data, k)
			if end < 0 {
				return spec
			}
			if kind == "page" {
				if key := pageKeyOf(attrs); key != "" {
					stack = append(stack, key)
				}
			} else if rec := strings.TrimSpace(attrs["name"]); rec != "" && len(stack) > 0 {
				if _, dup := spec.PageByRecord[rec]; !dup {
					spec.PageByRecord[rec] = stack[len(stack)-1]
				}
			}
			i = end
		case "close":
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			k := next + len("</Page")
			if k < len(data) && data[k] == '>' {
				k++
			}
			i = k
		}
	}
	return spec
}

// pageKeyOf 取页签键：text 优先（真标题也在这里），没有 text 才退回 name。
func pageKeyOf(attrs map[string]string) string {
	if s := strings.TrimSpace(attrs["text"]); s != "" {
		return s
	}
	return strings.TrimSpace(attrs["name"])
}

// isTagNameEnd 判断标签名后是否就是属性区（区分 `<Page` 与 `<PageX`）。
func isTagNameEnd(c byte) bool {
	switch c {
	case ' ', '\t', '\r', '\n', '>', '/':
		return true
	}
	return false
}

// looksLikePageKey 判断一个 4fd 的 text 是"页签键"还是"真标题"。
//
// 实测本程序全是键（page_1 / bpage_1 / axmt500_wf.page_10）；若别处真把标题写进 text
// （如"订单明细"），它不含 page，就按标题用。
func looksLikePageKey(s string) bool {
	return strings.Contains(strings.ToLower(s), "page")
}

// applyPageTitles 给页面补 PageName/Title。
//
// 页签键只来自 .4fd（record→text）；标题取 .tsd 同名键的文案。两者都没有就留空 ——
// 宁可没有标题，也不按序号猜（.tab 的 page 4 对应的是 4fd 的 bpage_4，tsd 的 page_4
// 反而落在 .tab 的 page 9 上，按序号对齐会整页错位）。
func applyPageTitles(doc *LinkDoc, tsd tsdSpec, fd *fdSpec) {
	if fd == nil {
		return
	}
	for i := range doc.Pages {
		pg := &doc.Pages[i]
		key := fd.PageByRecord[pg.Record]
		if key == "" {
			continue
		}
		pg.PageName = key
		if t := tsd.Pages[key]; t != "" {
			pg.Title = t
		} else if !looksLikePageKey(key) {
			pg.Title = key // 真把标题写进 text 的表单
		}
	}
}

// applyFieldLabels 把标签贴到字段上（没有就不贴，绝不用字段名猜）。
func applyFieldLabels(fields []LinkField, labels map[string]string) {
	if len(labels) == 0 {
		return
	}
	for i := range fields {
		if l := labels[fields[i].Column]; l != "" {
			fields[i].Label = l
		}
	}
}

// ---------- 行号索引 ----------

type lineIndex struct{ starts []int64 }

func newLineIndex(b []byte) *lineIndex {
	li := &lineIndex{starts: []int64{0}}
	for i, c := range b {
		if c == '\n' {
			li.starts = append(li.starts, int64(i)+1)
		}
	}
	return li
}

// lineOf 把字节偏移换成 1 起行号（二分）。
func (li *lineIndex) lineOf(off int64) int {
	lo, hi := 0, len(li.starts)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if li.starts[mid] <= off {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo + 1
}
