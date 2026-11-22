package devdata

// 蓝图 ↔ 代码 的索引（`tt gen map`）。
//
// 输出一份 link.json：把三份材料对成一张表 ——
//
//	.tab 的页面/字段  →  .tap 的槽位（加点）  →  .4gl 的行号区间  →  切片模板 aXX(Version:N)
//
// 为什么要它：一个 69,254 行的 .4gl 里，真正"人写过"的只有 9,048 行（13%），
// 而某一个需求涉及的往往只有一两百行。没有索引时，AI 只能 grep 整份代码，既慢又不知道
// "这段是客制还是标准、重产会不会被覆盖"。有了索引，三件事变成一次查询：
//
//	1. 这个字段/页面的代码在哪（精确到行）；
//	2. 这行是客制内容还是模板正文（后者重产就没了）；
//	3. 这段是哪个切片模板（aXX）生成的第几版（可做脱版检测）。
//
// 身份规则（与生成器一致，实测定下来的）：
//   · 槽位名 = `<功能>.<位置>.<页 id>.<字段>`，其中页用的是 `.tab` 的 **page_id**（不是页签序号）；
//   · 客制点可以挂到标准槽位上（`mapping` 与 `name` 不同），两者都留；
//   · 重名允许（实测 `construct.c.page1.xmdc027` ×3），所以键是"名字 + 出现次序"。

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/dev/pkgfile"
)

// LinkSchema 是 link.json 的 schema 标识。
// LinkSchema 变了就作废旧索引。改解析/关联逻辑时必须升这个版本 ——
// 否则 LoadIndex 会继续复用旧 link.json，行号悄悄错下去。
//
// v4：加了 .tsd 的字段标签（LinkField.Label）与 Sources.Tsd —— 旧索引没有标签，
// 复用旧索引会静默少一列信息，所以必须升版本。
const LinkSchema = "tt.gen.link/v4"

// linkJSONName 索引文件名（材料目录里；工作区会把它放到自己的 index/ 下）。
const linkJSONName = "link.json"

// ErrLinkInput 输入不足以建索引（材料缺件/读不动）。
var ErrLinkInput = errors.New("建立索引的输入不足")

// LinkRequest 是建索引的请求。
//
// 三个落点字段是给**工作区**用的：工作区里材料、索引、模板树都不在 devdata/<环境>/
// 下面，而是跟着任务目录走。空串一律表示"用缺省"，所以老调用点（只给 CfgPath/Env/Prog）
// 行为完全不变。
//
//	Dir  材料目录（缺省 = ProgDir(CfgPath,Env,Prog)）
//	Out  link.json 落点（缺省 = <材料目录>/link.json）
//	Mdl  模板树根（缺省 = MdlDir(CfgPath,Env)）
type LinkRequest struct {
	CfgPath string
	Env     string
	Prog    string
	Dir     string
	Out     string
	Mdl     string
}

// matDir 材料目录。
func (r LinkRequest) matDir() (string, error) {
	if d := strings.TrimSpace(r.Dir); d != "" {
		return filepath.Abs(d)
	}
	return ProgDir(r.CfgPath, r.Env, r.Prog)
}

// indexPath 索引落点（dir 是材料目录）。
func (r LinkRequest) indexPath(dir string) string {
	if p := strings.TrimSpace(r.Out); p != "" {
		return p
	}
	return filepath.Join(dir, linkJSONName)
}

// mdlRoot 模板树根；取不到就返回空串（finishSlices 会把"没镜像模板树"记成一条 ok 检查）。
func (r LinkRequest) mdlRoot() string {
	if d := strings.TrimSpace(r.Mdl); d != "" {
		return d
	}
	if d, err := MdlDir(r.CfgPath, r.Env); err == nil {
		return d
	}
	return ""
}

// LinkDoc 是 link.json 的内容。
type LinkDoc struct {
	Schema      string        `json:"schema"`
	Prog        string        `json:"prog"`
	Env         string        `json:"env"`
	GeneratedAt string        `json:"generatedAt"`
	Sources     LinkSources   `json:"sources"`
	Mode        LinkMode      `json:"mode"`
	Stats       LinkStats     `json:"stats"`
	Checks      []LinkCheck   `json:"checks"`
	Sections    []LinkSection `json:"sections"`
	Points      []LinkPoint   `json:"points"`
	Slices      LinkSlices    `json:"slices"`
	Fields      []LinkField   `json:"fields"`
	Pages       []LinkPage    `json:"pages,omitempty"`
	Notes       []string      `json:"notes,omitempty"`
	Orphans     []LinkOrphan  `json:"orphans,omitempty"`

	Path string `json:"-"`
}

// LinkSources 记录索引来自哪几份文件（带指纹，便于判断索引是否过期）。
type LinkSources struct {
	Tab *SrcInfo `json:"tab,omitempty"`
	// Tsd 表单规格（取字段标签 `lbl_<字段>` 与页签标题 `page_<n>`）。
	Tsd *SrcInfo `json:"tsd,omitempty"`
	// Fd 表单定义（.4fd；把 screen record 接到页签键，标题最终靠它落地）。
	Fd  *SrcInfo `json:"4fd,omitempty"`
	Tap SrcInfo  `json:"tap"`
	G4  SrcInfo  `json:"4gl"`
	Mdl string   `json:"mdlDir,omitempty"`
}

// SrcInfo 是一份输入的指纹。
type SrcInfo struct {
	Rel     string `json:"rel"`
	Bytes   int64  `json:"bytes"`
	SHA256  string `json:"sha256"`
	ModTime string `json:"modTime"`
}

// LinkMode 是"这份代码是怎么生成出来的"（决定锚点可信度）。
type LinkMode struct {
	Type            string `json:"type,omitempty"` // 主模板：t01 / q01 / r01 / p01
	Industry        string `json:"industry,omitempty"`
	Module          string `json:"module,omitempty"`
	Identity        string `json:"identity,omitempty"` // c=客制 s=标准
	SectionFlag     string `json:"sectionFlag,omitempty"`
	UnfoldedSection bool   `json:"unfoldedSection"`
	Freestyle       bool   `json:"freestyle"`
	CusVer          string `json:"cusVer,omitempty"` // 0000 = 从未客制
	Pages           int    `json:"pages,omitempty"`
}

// LinkStats 是给摘要看的计数。
type LinkStats struct {
	G4Lines         int `json:"g4Lines"`
	SectionsTap     int `json:"sectionsTap"`
	SectionsG4      int `json:"sectionsG4"`
	SlotsTap        int `json:"slotsTap"`
	SlotsG4         int `json:"slotsG4"`
	Points          int `json:"points"`
	PointsCustom    int `json:"pointsCustom"` // 声明里 src=c 且内容非空的点
	PointsCommented int `json:"pointsCommented"`
	PointsMissing   int `json:"pointsMissing"` // tap 有、4gl 无
	PointsExtra     int `json:"pointsExtra"`   // 4gl 有、tap 无
	CustomLines     int `json:"customLines"`
	FieldsLinked    int `json:"fieldsLinked"`
	ColumnsTotal    int `json:"columnsTotal"` // .tab 里的字段总数（含没有逻辑的纯数据字段）
	SlicesTotal     int `json:"slicesTotal"`
	SlicesDistinct  int `json:"slicesDistinct"`
	// 声明（.tap 的 <point> 内容）侧：与"点位"侧分开算 —— 声明可能没有槽位（orphan），
	// 也可能以函数形式落地（unmarked）。只看点位会漏掉这两类。
	DeclTotal       int `json:"declTotal"`
	DeclCustom      int `json:"declCustom"`
	DeclCustomLines int `json:"declCustomLines"`
	DeclUnmarked    int `json:"declUnmarked"`
	DeclOrphans     int `json:"declOrphans"`
	// 孤儿里的客制声明数：这才是"客户写的东西丢了"，标准声明残留只是噪音。
	DeclOrphansCustom int `json:"declOrphansCustom"`
}

// LinkCheck 是一条一致性检查结果（level：ok/warn/error）。
type LinkCheck struct {
	ID     string `json:"id"`
	Level  string `json:"level"`
	Detail string `json:"detail"`
}

// LinkSection 是一个 section 的两侧对账结果。
type LinkSection struct {
	ID      string `json:"id"`
	Ver     string `json:"ver,omitempty"`
	Src     string `json:"src,omitempty"`
	TapLine int    `json:"tapLine,omitempty"`
	Line0   int    `json:"line0,omitempty"`
	Line1   int    `json:"line1,omitempty"`
	Slots   int    `json:"slots"`
	SlotsG4 int    `json:"slotsG4"`
}

// LinkPoint 是一个槽位（加点）在两侧的对应关系。
type LinkPoint struct {
	Name      string `json:"name"`
	Section   string `json:"section"`
	TapLine   int    `json:"tapLine,omitempty"`
	Line0     int    `json:"line0,omitempty"`
	Line1     int    `json:"line1,omitempty"`
	Src       string `json:"src,omitempty"` // s=标准内容 c=客制内容（无声明则空）
	Ver       string `json:"ver,omitempty"` // 声明的版次
	Status    string `json:"status,omitempty"`
	Slot      string `json:"slot,omitempty"`   // mapping：挂靠的标准槽位
	G4Name    string `json:"g4Name,omitempty"` // 与 name 不同时（截断/改名）
	DeclLines int    `json:"declLines,omitempty"`
	Custom    bool   `json:"custom,omitempty"`
	Commented bool   `json:"commented,omitempty"`
	Missing   bool   `json:"missing,omitempty"` // tap 有、4gl 无（4gl 可能旧）
	Extra     bool   `json:"extra,omitempty"`   // 4gl 有、tap 无（生成器自加）
	// Unmarked=true：内容在 .4gl 里以函数形式落地，但**没有对应的 #add-point 标记**
	// （`function.*` 这类客制函数就是这样，行号由函数名定位，不是由槽位定位）。
	Unmarked bool `json:"unmarked,omitempty"`
	// G4Only=true：没有 .tap 可比（标准程序），点位只来自 .4gl，客制/标准未知。
	G4Only  bool   `json:"g4Only,omitempty"`
	Preview string `json:"preview,omitempty"`
}

// LinkOrphan 是一条"声明了但落不了地"的加点。
//
// 这是最容易骗人的一类：`.tap` 里明明写了客制内容（src=c），但骨架里没有对应的洞，
// 于是重产之后这段代码**根本不会出现在 .4gl 里**。实测例子：把某页的"项次"列删掉后，
// `input.a.page73.xmdducdocno` 的声明还在，槽位已经没了。
type LinkOrphan struct {
	Name    string `json:"name"`
	Src     string `json:"src,omitempty"`
	Ver     string `json:"ver,omitempty"`
	Lines   int    `json:"lines"`
	Reason  string `json:"reason"` // no-slot：骨架里没有这个洞；no-code：函数也没找到
	Preview string `json:"preview,omitempty"`
}

// LinkSlices 是切片（aXX 模板）的使用情况。
type LinkSlices struct {
	Total    int          `json:"total"`
	Distinct int          `json:"distinct"`
	ByCode   []SliceUse   `json:"byCode"`
	Drift    []SliceDrift `json:"drift,omitempty"`
	Markers  []SliceMark  `json:"markers,omitempty"`
}

// SliceUse 是一个切片的用量。
type SliceUse struct {
	Code  string `json:"code"`
	Ver   int    `json:"ver"`
	Count int    `json:"count"`
}

// SliceDrift 是"产物里的模板版次 ≠ 当前模板版次"。
type SliceDrift struct {
	Code    string `json:"code"`
	Used    int    `json:"used"`
	Current int    `json:"current"`
}

// SliceMark 是一条切片标记（行号 → 切片），供按行反查。
type SliceMark struct {
	Line int    `json:"l"`
	Code string `json:"c"`
	Ver  int    `json:"v"`
}

// LinkPage 是蓝图页面（page 是页签序号，pageId 才是点名用的 id）。
type LinkPage struct {
	Page    string `json:"page"`
	PageID  string `json:"pageId"`
	Record  string `json:"record,omitempty"`
	Columns int    `json:"columns"`
	Fields  int    `json:"fields"` // 有加点关联的字段数
	// PageName 是 .4fd 里的页签键（`page_13`）：它跟 .tsd 的标题、跟设计器上的页签
	// 都对得上，而 `page`（.tab 的序号）只对得上设计器的显示顺序。
	PageName string `json:"pageName,omitempty"`
	// Title 页签标题（.4fd 自带则用它的，否则用 .tsd 的 `page_<n>` 文案）。
	Title string `json:"title,omitempty"`
}

// Kind 把这个点位归成一句人读的来源标签（CLI 与 JSON 共用同一套判断）。
//
// 判断顺序是"越具体越先"：unmarked 比 Custom 更具体（`function.*` 这类点是按函数名
// 定位的、没有加点标记），所以客制函数必须先判。否则 Custom 先命中，整个
// "客制函数" 标签永远不可达（实测 18 个客制函数在 at/grep 里都被报成普通客制加点）。
func (p LinkPoint) Kind() string {
	switch {
	case p.Unmarked && p.Custom:
		return "客制函数"
	case p.Custom:
		return "客制"
	case p.Unmarked:
		// 标准侧也有以函数形式落地、没有加点标记的点：来源不是客制，别冒充
		return "函数（无标记）"
	case p.Extra:
		return "生成器自加"
	case p.Missing:
		return "未落地"
	case p.G4Only:
		return "未知（无 .tap）"
	case p.Src == "s":
		return "标准"
	default:
		return "模板"
	}
}

// LinkField 是蓝图字段到槽位的映射（.tab 的 column ↔ 点名里的 column）。
type LinkField struct {
	Page   string `json:"page,omitempty"`
	PageID string `json:"pageId"`
	Record string `json:"record,omitempty"`
	Column string `json:"column"`
	// Label 画面上的中文名（.tsd 的 `lbl_<字段>`）。缺失时为空 —— 不要用字段名去猜，
	// 猜出来的中文名会让人拿着错的词去搜代码。
	Label  string   `json:"label,omitempty"`
	Points []string `json:"points"`
}

// rePointPage 从点名里取页 id 与字段：`input.a.page73.xmdduc001` → (73, xmdduc001)。
var rePointPage = regexp.MustCompile(`\.page(\d+)\.(.+)$`)

// Map 建立索引并写 link.json。
func Map(req LinkRequest) (*LinkDoc, error) {
	if err := ValidProgName(req.Prog); err != nil {
		return nil, err
	}
	progDir, err := req.matDir()
	if err != nil {
		return nil, err
	}
	mf, err := readPackManifest(progDir, req.Prog)
	if err != nil {
		return nil, err
	}

	read := func(role string) ([]byte, SrcInfo, error) {
		rel, fe := relOfRole(mf, role)
		if rel == "" {
			return nil, SrcInfo{}, nil
		}
		b, rerr := os.ReadFile(filepath.Join(progDir, rel))
		if rerr != nil {
			return nil, SrcInfo{}, fmt.Errorf("读材料失败 (%s): %w", rel, rerr)
		}
		info := SrcInfo{Rel: rel, Bytes: int64(len(b)), SHA256: pkgfile.Sha256Hex(b)}
		// 用**服务器上的 mtime**（manifest 里记着），而不是下载到本地的时间 ——
		// 后者全是"刚刚"，会让"4gl 比 tap 旧"这条检查永远判 ok。
		if fe != nil && fe.SrcMtime != "" {
			info.ModTime = fe.SrcMtime
		} else if fi, serr := os.Stat(filepath.Join(progDir, rel)); serr == nil {
			info.ModTime = fi.ModTime().Format(time.RFC3339)
		}
		return b, info, nil
	}

	tapBytes, tapInfo, err := read("tap")
	if err != nil {
		return nil, err
	}
	g4Bytes, g4Info, err := read("4gl")
	if err != nil {
		return nil, err
	}
	tabBytes, tabInfo, err := read("tab")
	if err != nil {
		return nil, err
	}
	// .tsd 只用来取字段标签/页签标题，读不到都不该让整次建索引失败。
	tsdBytes, tsdInfo, err := read("tsd")
	if err != nil {
		return nil, err
	}
	// .4fd 用来把 screen record 接到页签键（.tab 与 .tsd 之间唯一的桥）。
	fdBytes, fdInfo, err := read("4fd")
	if err != nil {
		return nil, err
	}
	if len(g4Bytes) == 0 {
		return nil, fmt.Errorf("%w: %s 缺 .4gl（先 tt gen fetch %s）",
			ErrMaterialIncomplete, req.Prog, req.Prog)
	}

	var tap *tapDoc
	if len(tapBytes) > 0 {
		if tap, err = parseTap(tapBytes); err != nil {
			return nil, err
		}
	} else {
		tap = &tapDoc{Attrs: map[string]string{}}
	}
	g4 := parse4gl(splitLines(g4Bytes))
	var tab *tabDoc
	if len(tabBytes) > 0 {
		if tab, err = parseTab(tabBytes); err != nil {
			return nil, err
		}
	}

	doc := buildLink(req, tap, g4, tab)
	doc.Sources = LinkSources{Tap: tapInfo, G4: g4Info}
	if tab != nil {
		doc.Sources.Tab = &tabInfo
	}
	var tsd tsdSpec
	if len(tsdBytes) > 0 {
		doc.Sources.Tsd = &tsdInfo
		tsd = parseTsd(tsdBytes)
		applyFieldLabels(doc.Fields, tsd.Fields)
	}
	var fd *fdSpec
	if len(fdBytes) > 0 {
		doc.Sources.Fd = &fdInfo
		fd = parse4fd(fdBytes)
	}
	applyPageTitles(doc, tsd, fd)
	doc.Sources.Mdl = req.mdlRoot()

	// 切片脱版检测：产物里用的 Version 与当前模板比。
	doc.Slices, doc.Checks = finishSlices(doc.Slices, doc.Sources.Mdl, doc.Checks)
	// 新鲜度：.4gl 是否比 .tap 旧（"改了 tap 没跑 F7"的直证）。要等 Sources 填好才算。
	doc.Checks = append(doc.Checks, freshnessCheck(doc.Sources, doc.Stats)...)

	doc.Path = req.indexPath(progDir)
	if err := writeJSONAtomic(doc.Path, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// buildLink 做纯内存的关联（便于单测直接喂结构体）。
func buildLink(req LinkRequest, tap *tapDoc, g4 *g4Doc, tab *tabDoc) *LinkDoc {
	doc := &LinkDoc{
		Schema: LinkSchema, Prog: req.Prog, Env: req.Env,
		GeneratedAt: nowStamp(),
	}
	doc.Mode = LinkMode{
		Module:          tap.Attrs["module"],
		Identity:        tap.Attrs["identity"],
		SectionFlag:     tap.Attrs["section_flag"],
		UnfoldedSection: g4.UnfoldedSection,
		Freestyle:       g4.Freestyle,
	}
	if tab != nil {
		doc.Mode.Type = tab.Assembly.Type
		doc.Mode.Industry = tab.Assembly.Industry
		doc.Mode.CusVer = tab.Assembly.CusVer
		doc.Mode.Pages = len(tab.Pages)
	}

	declByName := map[string]tapDecl{}
	for _, d := range tap.Decls {
		if _, ok := declByName[d.Name]; !ok || declByName[d.Name].Blank {
			declByName[d.Name] = d
		}
	}

	g4ByID := make(map[string]*g4Section, len(g4.Sections))
	for i := range g4.Sections {
		g4ByID[g4.Sections[i].ID] = &g4.Sections[i]
	}

	doc.Stats.G4Lines = g4.TotalLines
	doc.Stats.SectionsTap = len(tap.Sections)
	doc.Stats.SectionsG4 = len(g4.Sections)

	// 按 tap 的 section 顺序逐个配；tap 里没有的 4gl section 另算。
	// 没有 .tap（从未被设计器接管过的标准程序就是这种）时，退化为"只用 4gl"：
	// 仍有 section / 槽位 / 切片 / 字段，只是没有"客制还是标准"这个维度。
	if len(tap.Sections) == 0 {
		for _, gs := range g4.Sections {
			doc.Sections = append(doc.Sections, LinkSection{
				ID: gs.ID, Line0: gs.Line0, Line1: gs.Line1, SlotsG4: len(gs.Slots),
			})
			for _, s := range gs.Slots {
				p := LinkPoint{Name: s.Name, Section: gs.ID, Line0: s.Line0, Line1: s.Line1,
					Commented: s.Commentd, G4Only: true}
				attachDecl(&p, declByName[s.Name])
				doc.Points = append(doc.Points, p)
			}
		}
	}
	for _, ts := range tap.Sections {
		gs := g4ByID[ts.ID]
		ls := LinkSection{ID: ts.ID, Ver: ts.Ver, Src: ts.Src, TapLine: ts.Line,
			Slots: len(ts.Slots)}
		if gs != nil {
			ls.Line0, ls.Line1, ls.SlotsG4 = gs.Line0, gs.Line1, len(gs.Slots)
		}
		doc.Sections = append(doc.Sections, ls)

		pts := pairSlots(ts, gs, declByName)
		doc.Points = append(doc.Points, pts...)
	}
	// 4gl 里多出来的 section（tap 没有）
	if len(tap.Sections) > 0 {
		for _, gs := range g4.Sections {
			if _, ok := sectionOf(tap, gs.ID); ok {
				continue
			}
			doc.Sections = append(doc.Sections, LinkSection{ID: gs.ID, Line0: gs.Line0, Line1: gs.Line1, SlotsG4: len(gs.Slots)})
			for _, s := range gs.Slots {
				doc.Points = append(doc.Points, LinkPoint{
					Name: s.Name, Section: gs.ID, Line0: s.Line0, Line1: s.Line1,
					Commented: s.Commentd, Extra: true,
				})
			}
			doc.Checks = append(doc.Checks, LinkCheck{
				ID: "section-extra", Level: "warn",
				Detail: fmt.Sprintf(".4gl 的 section %q 在 .tap 里没有（生成器自加或 tap 旧）", gs.ID),
			})
		}
	}

	// 声明侧收尾：把"没配到槽位"的声明要么按函数名定位，要么记成 orphan。
	// 这一步不能省：只看点位会把 `function.*`（18 个、几千行）与"槽位被删掉的残留声明"全漏掉。
	matched := make(map[string]bool, len(doc.Points))
	for _, p := range doc.Points {
		matched[p.Name] = true
	}
	for _, d := range tap.Decls {
		doc.Stats.DeclTotal++
		if d.Src == "c" && !d.Blank {
			doc.Stats.DeclCustom++
			doc.Stats.DeclCustomLines += d.Lines
		}
		if matched[d.Name] {
			continue
		}
		if fname, ok := strings.CutPrefix(d.Name, "function."); ok {
			if l0, l1, found := findFunctionSpan(g4.Lines, fname); found {
				doc.Points = append(doc.Points, LinkPoint{
					Name: d.Name, Section: "other_function", Line0: l0, Line1: l1,
					Src: d.Src, Ver: d.Ver, Slot: d.Slot, DeclLines: d.Lines,
					Custom: d.Src == "c", Unmarked: true, Preview: truncate(d.FirstUse, 100),
				})
				doc.Stats.DeclUnmarked++
				continue
			}
		}
		doc.Orphans = append(doc.Orphans, LinkOrphan{
			Name: d.Name, Src: d.Src, Ver: d.Ver, Lines: d.Lines,
			Reason: "no-slot", Preview: truncate(d.FirstUse, 100),
		})
		doc.Stats.DeclOrphans++
		if d.Src == "c" {
			doc.Stats.DeclOrphansCustom++
		}
	}

	// 统计与逐点收尾
	for i := range doc.Points {
		p := &doc.Points[i]
		if p.Custom {
			doc.Stats.PointsCustom++
			doc.Stats.CustomLines += p.DeclLines
		}
		if p.Commented {
			doc.Stats.PointsCommented++
		}
		if p.Missing {
			doc.Stats.PointsMissing++
		}
		if p.Extra {
			doc.Stats.PointsExtra++
		}
	}
	doc.Stats.Points = len(doc.Points)
	doc.Stats.SlotsTap = countTapSlots(tap)
	doc.Stats.SlotsG4 = countG4Slots(g4)

	// 蓝图字段 ↔ 点位
	doc.Fields = buildFields(doc.Points, tab)
	doc.Stats.FieldsLinked = len(doc.Fields)
	if tab != nil {
		byPage := map[string]int{}
		for _, f := range doc.Fields {
			byPage[f.PageID]++
		}
		for _, pg := range tab.Pages {
			doc.Stats.ColumnsTotal += len(pg.Columns)
			doc.Pages = append(doc.Pages, LinkPage{
				Page: pg.Page, PageID: pg.PageID, Record: pg.Record,
				Columns: len(pg.Columns), Fields: byPage[pg.PageID],
			})
		}
	}

	// 切片
	for _, s := range g4.Slices {
		doc.Slices.Total++
		doc.Slices.Markers = append(doc.Slices.Markers, SliceMark{Line: s.Line, Code: s.Code, Ver: s.Ver})
	}
	doc.Slices.ByCode, doc.Slices.Distinct = aggregateSlices(g4.Slices)
	doc.Stats.SlicesTotal = doc.Slices.Total
	doc.Stats.SlicesDistinct = doc.Slices.Distinct

	doc.Checks = append(doc.Checks, sectionChecks(tap, g4)...)
	doc.Checks = append(doc.Checks, slotChecks(doc.Sections, doc.Stats)...)
	doc.Checks = append(doc.Checks, declChecks(doc.Stats)...)
	doc.Checks = append(doc.Checks, blueprintCheck(tab, doc.Stats)...)
	return doc
}

// pairSlots 把一个 section 两侧的槽位配上。
//
// 三步（顺序很重要）：
//  1. **按名字 + 出现次序**精确配 —— 这是主路径，绝大多数点走这里；
//  2. 剩下的**按行号顺序**配 —— 处理名字被截断的情况（实测 `chk_data_exist.define`
//     在 .tap 的标记行里被写成 `name="chk_`，只有洞是完整的）；
//  3. 还剩下的：tap 侧 → missing（4gl 可能旧），4gl 侧 → extra（生成器自加）。
func pairSlots(ts tapSection, gs *g4Section, decls map[string]tapDecl) []LinkPoint {
	type pending struct{ tap *tapSlot }
	var out []LinkPoint

	g4Slots := []g4Slot{}
	if gs != nil {
		g4Slots = gs.Slots
	}
	used := make([]bool, len(g4Slots))

	// 名字 → 未使用的 4gl 下标队列（保持顺序）
	byName := map[string][]int{}
	for i, s := range g4Slots {
		byName[s.Name] = append(byName[s.Name], i)
	}

	pair := func(t *tapSlot, j int, nameMismatch bool) LinkPoint {
		used[j] = true
		s := g4Slots[j]
		p := LinkPoint{
			Name: t.Name, Section: ts.ID, TapLine: t.Line,
			Line0: s.Line0, Line1: s.Line1, Commented: s.Commentd,
		}
		if nameMismatch {
			p.G4Name = s.Name
		}
		attachDecl(&p, decls[t.Name])
		return p
	}

	// ① 名字 + 次序
	var restTap []*tapSlot
	for i := range ts.Slots {
		t := &ts.Slots[i]
		q := byName[t.Name]
		picked := -1
		for _, j := range q {
			if !used[j] {
				picked = j
				break
			}
		}
		if picked >= 0 {
			out = append(out, pair(t, picked, false))
		} else {
			restTap = append(restTap, t)
		}
	}

	// ② 剩余按顺序配
	var restG4 []int
	for j := range g4Slots {
		if !used[j] {
			restG4 = append(restG4, j)
		}
	}
	n := len(restTap)
	if len(restG4) < n {
		n = len(restG4)
	}
	for k := 0; k < n; k++ {
		out = append(out, pair(restTap[k], restG4[k], true))
	}

	// ③ 余下：tap 侧 missing
	for k := n; k < len(restTap); k++ {
		t := restTap[k]
		p := LinkPoint{Name: t.Name, Section: ts.ID, TapLine: t.Line, Missing: true, Commented: t.Commentd}
		attachDecl(&p, decls[t.Name])
		out = append(out, p)
	}
	// ④ 余下：4gl 侧 extra
	for k := n; k < len(restG4); k++ {
		s := g4Slots[restG4[k]]
		out = append(out, LinkPoint{
			Name: s.Name, Section: ts.ID, Line0: s.Line0, Line1: s.Line1,
			Commented: s.Commentd, Extra: true,
		})
	}

	// 按 4gl 行号排序（没有行号的排后面），让输出可读、可按行二分
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].Line0, out[j].Line0
		if a == 0 {
			return false
		}
		if b == 0 {
			return true
		}
		return a < b
	})
	return out
}

// attachDecl 把 .tap 的声明属性挂到点上。
func attachDecl(p *LinkPoint, d tapDecl) {
	if d.Name == "" {
		return
	}
	p.Src = d.Src
	p.Ver = d.Ver
	p.Status = d.Status
	p.Slot = d.Slot
	p.DeclLines = d.Lines
	p.Custom = d.Src == "c" && !d.Blank
	if d.FirstUse != "" {
		p.Preview = truncate(strings.TrimSpace(d.FirstUse), 100)
	}
}

// buildFields 把点位按"页 id + 字段"归组，并与 .tab 的页面/字段对账。
func buildFields(points []LinkPoint, tab *tabDoc) []LinkField {
	type key struct{ pageID, column string }
	seen := map[key][]string{}
	for _, p := range points {
		m := rePointPage.FindStringSubmatch(p.Name)
		if m == nil {
			continue
		}
		k := key{m[1], m[2]}
		if _, ok := seen[k]; !ok {
			seen[k] = nil
		}
		seen[k] = append(seen[k], p.Name)
	}
	// 蓝图侧的信息（page / record）
	meta := map[key]tabPage{}
	if tab != nil {
		for _, pg := range tab.Pages {
			for _, c := range pg.Columns {
				meta[key{pg.PageID, c}] = pg
			}
		}
	}
	keys := make([]key, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].pageID != keys[j].pageID {
			return keys[i].pageID < keys[j].pageID
		}
		return keys[i].column < keys[j].column
	})
	out := make([]LinkField, 0, len(keys))
	for _, k := range keys {
		if tab != nil {
			// 有蓝图就以蓝图为准：只认 .tab 里真实存在的 column。
			// 否则 `ui_dialog.page1.action` 这类"函数式点位"会被当成字段（实测 321 个里有几十个是假的）。
			if _, ok := meta[k]; !ok {
				continue
			}
		}
		f := LinkField{PageID: k.pageID, Column: k.column, Points: seen[k]}
		if pg, ok := meta[k]; ok {
			f.Page, f.Record = pg.Page, pg.Record
		}
		out = append(out, f)
	}
	return out
}

// aggregateSlices 按切片代号聚合用量。
func aggregateSlices(sl []g4Slice) ([]SliceUse, int) {
	type agg struct {
		ver   int
		count int
	}
	m := map[string]*agg{}
	for _, s := range sl {
		a := m[s.Code]
		if a == nil {
			a = &agg{ver: s.Ver}
			m[s.Code] = a
		}
		a.count++
	}
	out := make([]SliceUse, 0, len(m))
	for code, a := range m {
		out = append(out, SliceUse{Code: code, Ver: a.ver, Count: a.count})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, len(out)
}

// finishSlices 做脱版检测：产物里的 Version 与当前模板文件首行比。
func finishSlices(sl LinkSlices, mdlDir string, checks []LinkCheck) (LinkSlices, []LinkCheck) {
	if mdlDir == "" {
		return sl, checks
	}
	tplDir := filepath.Join(mdlDir, "slice")
	if _, err := os.Stat(tplDir); err != nil {
		return sl, append(checks, LinkCheck{ID: "slice-version", Level: "ok",
			Detail: "没镜像模板树（tt gen mirror --only mdl 后可做脱版检测）"})
	}
	drifted := 0
	for i := range sl.ByCode {
		u := &sl.ByCode[i]
		b, err := os.ReadFile(filepath.Join(tplDir, "code_"+u.Code+".template"))
		if err != nil {
			continue
		}
		cur := sliceVersionOf(b)
		if cur > 0 && cur != u.Ver {
			sl.Drift = append(sl.Drift, SliceDrift{Code: u.Code, Used: u.Ver, Current: cur})
			drifted++
		}
	}
	level := "ok"
	detail := fmt.Sprintf("%d 种切片，版本与当前模板一致", len(sl.ByCode))
	if drifted > 0 {
		level = "warn"
		detail = fmt.Sprintf("%d 种切片用了旧版模板（产物 Version 小于模板首行 Version）", drifted)
	}
	return sl, append(checks, LinkCheck{ID: "slice-version", Level: level, Detail: detail})
}

// sliceVersionOf 取模板首行里的 Version（复用 manifest.go 里的解析）。
func sliceVersionOf(b []byte) int {
	if v, ok := ParseTemplateVersion(b); ok {
		return v
	}
	return 0
}

// ---------- 检查 ----------

func sectionChecks(tap *tapDoc, g4 *g4Doc) []LinkCheck {
	var out []LinkCheck
	if len(tap.Sections) == 0 {
		return append(out, LinkCheck{ID: "sections-pair", Level: "warn",
			Detail: ".tap 里没有 section（该程序可能不是分段落版型）"})
	}
	if len(g4.Sections) == 0 {
		return append(out, LinkCheck{ID: "sections-pair", Level: "warn",
			Detail: ".4gl 里没有 section 锚（非生成器产物或 freestyle）"})
	}
	same := len(tap.Sections) == len(g4.Sections)
	if same {
		for i := range tap.Sections {
			if tap.Sections[i].ID != g4.Sections[i].ID {
				same = false
				break
			}
		}
	}
	if same {
		out = append(out, LinkCheck{ID: "sections-pair", Level: "ok",
			Detail: fmt.Sprintf("%d 个 section 两侧同名同序", len(tap.Sections))})
	} else {
		out = append(out, LinkCheck{ID: "sections-pair", Level: "error",
			Detail: fmt.Sprintf("section 两侧对不上：.tap %d 个 / .4gl %d 个（4gl 可能不是同一版）",
				len(tap.Sections), len(g4.Sections))})
	}
	return out
}

func slotChecks(sections []LinkSection, st LinkStats) []LinkCheck {
	var out []LinkCheck
	switch {
	case st.PointsMissing == 0 && st.PointsExtra == 0:
		out = append(out, LinkCheck{ID: "slots-pair", Level: "ok",
			Detail: fmt.Sprintf("%d 个槽位两侧一一对应", st.Points)})
	default:
		lvl := "warn"
		out = append(out, LinkCheck{ID: "slots-pair", Level: lvl,
			Detail: fmt.Sprintf("槽位 .tap %d / .4gl %d：%d 个只在 tap（4gl 可能旧）、%d 个只在 4gl（生成器自加，多为被注释掉的重复点）",
				st.SlotsTap, st.SlotsG4, st.PointsMissing, st.PointsExtra)})
	}
	if st.PointsCommented > 0 {
		out = append(out, LinkCheck{ID: "slots-commented", Level: "ok",
			Detail: fmt.Sprintf("%d 个槽位在 .4gl 里被注释掉（生成器有意屏蔽，不是丢失）", st.PointsCommented)})
	}
	return out
}

// declChecks 检查"声明了但落不了地"的加点。
func declChecks(st LinkStats) []LinkCheck {
	var out []LinkCheck
	if st.DeclUnmarked > 0 {
		out = append(out, LinkCheck{ID: "decls-unmarked", Level: "ok",
			Detail: fmt.Sprintf("%d 个声明以函数形式落地（function.*），无 #add-point 标记，按函数名定位", st.DeclUnmarked)})
	}
	if st.DeclOrphans > 0 {
		detail := fmt.Sprintf("%d 个声明在骨架里没有槽位、在 .4gl 里也找不到 = 写了但不会生效（多因字段/页被删）",
			st.DeclOrphans)
		if st.DeclOrphansCustom > 0 {
			detail += fmt.Sprintf("；其中**客制 %d 个**（这些是真的丢了客户代码，其余是标准声明的残留）",
				st.DeclOrphansCustom)
		} else {
			detail += "；全是标准声明的残留（不影响客制）"
		}
		out = append(out, LinkCheck{ID: "decls-orphan", Level: "warn", Detail: detail})
	}
	return out
}

func blueprintCheck(tab *tabDoc, st LinkStats) []LinkCheck {
	if tab == nil {
		return []LinkCheck{{ID: "blueprint", Level: "warn",
			Detail: "没有 .tab 蓝图：只能按点位检索，无法与页面/字段对账（tt gen fetch 会一起拉）"}}
	}
	lvl := "ok"
	detail := fmt.Sprintf("版型 %s，%d 个页面，%d 个字段能对上槽位", tab.Assembly.Type, len(tab.Pages), st.FieldsLinked)
	if tab.Assembly.CusVer == "0000" {
		detail += "（cusver=0000：从未客制）"
	}
	return []LinkCheck{{ID: "blueprint", Level: lvl, Detail: detail}}
}

// freshnessCheck 用源文件的 mtime 判断 .4gl 是不是比 .tap 旧。
//
// 这是"改了 tap 但没跑 F7（产生程式）"的直接证据：4gl 落后于 tap 时，
// 索引里的行号仍然有效，但"内容是否已生效"要打问号。
func freshnessCheck(src LinkSources, st LinkStats) []LinkCheck {
	var out []LinkCheck
	tap, err1 := time.Parse(time.RFC3339, src.Tap.ModTime)
	g4, err2 := time.Parse(time.RFC3339, src.G4.ModTime)
	if err1 != nil || err2 != nil {
		return out
	}
	if g4.Before(tap) {
		out = append(out, LinkCheck{ID: "freshness", Level: "warn",
			Detail: fmt.Sprintf(".4gl（%s）比 .tap（%s）旧：可能没跑 F7（产生程式），行号有效但内容未必已生效",
				src.G4.ModTime, src.Tap.ModTime)})
		if st.PointsMissing > 0 {
			out = append(out, LinkCheck{ID: "freshness-detail", Level: "warn",
				Detail: fmt.Sprintf("%d 个槽位只存在于 .tap —— 与上一条同因：重产一次即可对齐", st.PointsMissing)})
		}
	} else {
		out = append(out, LinkCheck{ID: "freshness", Level: "ok", Detail: ".4gl 不旧于 .tap"})
	}
	return out
}

// ---------- 小工具 ----------

var (
	reFuncDef = regexp.MustCompile(`(?i)^\s*(?:PRIVATE|PUBLIC)?\s*FUNCTION\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	reFuncEnd = regexp.MustCompile(`(?i)^\s*END FUNCTION`)
)

// findFunctionSpan 在 .4gl 里按函数名定位一段代码（1 起行号，闭区间）。
//
// `function.<名字>` 这类客制函数在 .4gl 里**没有 #add-point 标记**（实测 18 个），
// 只能靠函数名定位；找不到就说明这段客制根本不在产物里。
func findFunctionSpan(lines []string, fname string) (int, int, bool) {
	for i, ln := range lines {
		m := reFuncDef.FindStringSubmatch(ln)
		if m == nil || m[1] != fname {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if reFuncEnd.MatchString(lines[j]) {
				return i + 1, j + 1, true
			}
		}
		return i + 1, len(lines), true
	}
	return 0, 0, false
}

func sectionOf(tap *tapDoc, id string) (tapSection, bool) {
	for _, s := range tap.Sections {
		if s.ID == id {
			return s, true
		}
	}
	return tapSection{}, false
}

func countTapSlots(tap *tapDoc) int {
	n := 0
	for _, s := range tap.Sections {
		n += len(s.Slots)
	}
	return n
}

func countG4Slots(g4 *g4Doc) int {
	n := 0
	for _, s := range g4.Sections {
		n += len(s.Slots)
	}
	return n
}

// hashOf 算内容摘要（索引复用判断用）。
func hashOf(b []byte) string { return pkgfile.Sha256Hex(b) }

// splitLines 按 \n 切行（保留空行，行号即下标+1）。
func splitLines(b []byte) []string {
	s := string(b)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n")
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// writeJSONAtomic 先写临时文件再改名（同目录 rename 才有原子性）。
func writeJSONAtomic(path string, v any) error {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败 (%s): %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+tmpSuffixPrefix+"*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	name := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(name)
		return fmt.Errorf("写临时文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}
	_ = os.Remove(path)
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
		return fmt.Errorf("改名失败 (%s): %w", path, err)
	}
	return nil
}
