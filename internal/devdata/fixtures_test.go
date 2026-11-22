package devdata

// 磁盘材料夹具构建器（gen 系列测试的公共底座）。
//
// 为什么需要它：本包所有真实结论都建立在"一整套磁盘材料"之上 —— .tab（蓝图）、.tsd（文案）、
// .4fd（页签桥）、.tap（骨架+声明）、.tgl/.tgx/.tgx2（排序骨架）、.4gl（产物）、inc 与
// manifest。link/query/brief/intent 的测试如果各写各的零散字符串，会有两个后果：
//  1. 每加一个场景就再造一次材料，很快与真机形态漂移（比如忘了 manifest 里记 srcMtime，
//     freshness 检查就永远走"ok"分支，测试等于没测）；
//  2. 解析器与消费者各用一套夹具，"解析对了但索引接不上"这类问题测不出来。
//
// 所以这里只放**一个**构建器：给定程序名/版型/页面/各种开关，在 t.TempDir() 下写出
// 与 FetchRoles 完全一致的目录形态（dzx/ src/ inc/ + manifest.json），并允许按需关掉
// 其中任何一件材料（标准程序没有 .tap、老程序没有 .tsd/.4fd …）。
//
// 所有内容形态都对齐 linkparse.go 的解析器（不是"理想 XML"）：
//   - .tab 用 `<assembly>`+`<structure>`+`<var id="body" page=… page_id=… record=…>`，
//     且故意夹一个 `<32wc …>` 非法标签名（真机上有，严格 XML 会报错）；
//   - .tap 用 `<add_points>`+`<section><![CDATA[…]]></section>`+`<point>` 声明，
//     洞与标记行按真机写法（标记行 name 可被截断，洞才是权威）；
//   - .4gl 用 `{<section id="…" type="s" >}` / `#add-point:… name="…"` / `#END add-point`，
//     另支持 freestyle（无 section 锚）与"未解開Section"自述；
//   - .4fd 的页签键取 **text** 而不是 name（真机的坑：name 是控件 id）；
//   - .tgl 带 SD/PR 版次头（ParseTglHeader 的形态）；
//   - manifest 的每条 FileEntry 都带 sha256 与 srcMtime（后者驱动 freshness 检查）。

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ---------- 描述材料的类型 ----------

// matPage 是 .tab 里的一个页面（page 是页签序号，pageID 是点名用的稳定 id）。
type matPage struct {
	Seq     string   // 页签序号（.tab 的 page 属性）
	ID      string   // page_id（点名里的 page<id>）
	Record  string   // Screen Record（.4fd 用它桥接到页签键）
	Columns []string // 蓝图字段（不带类型/展延后缀；构建器会自己加）
}

// matSlot 是一个加点（槽位）。
type matSlot struct {
	Name  string // 点名（`input.a.page1.f1`）
	Label string // `#add-point:` 后的中文标签（人 grep 的是它）
	// DeclBody 非空 → 在 .tap 里写一个 `<point>` 声明（正文就是它）。
	DeclBody string
	// Src 声明身份：c=客制（默认）/ s=标准。
	Src string
	// Ver 声明版次。
	Ver string
	// Mapping 声明挂靠的标准槽位（客制点可挂到标准槽位上）。
	Mapping string
	// Body 是 .4gl 槽位里的正文（写在 #add-point 与 #END 之间）。
	Body []string
	// TapOnly=true 只在 .tap 里有洞（4gl 旧了/该段没生成）→ missing。
	TapOnly bool
	// G4Only=true 只在 .4gl 里有标记（生成器自加/被注释掉的重复点）→ extra。
	G4Only bool
	// Commentd=true 4gl 里这个槽位整行被注释掉（生成器有意屏蔽）。
	Commentd bool
	// TruncatedMarker=true 标记行里的 name 被截断（完整名字在洞里）。
	TruncatedMarker bool
}

// matSection 是一个 section（两侧同名同序时 sections-pair 才是 ok）。
type matSection struct {
	ID       string
	Ver      string
	Src      string
	Type     string // `type="s"`：设计器的段类型（可空）
	Readonly string // `readonly="Y"`：只读段（可空，真机上有）
	Slots    []matSlot
	// G4Only=true 只有 .4gl 侧有（.tap 没有 → section-extra 警告）。
	G4Only bool
}

// matOrphan 是"写了但骨架没有槽位"的声明（孤儿）。
type matOrphan struct {
	Name string
	Src  string
	Ver  string
	Body string
}

// matFunc 是 `function.<名字>` 形式的客制函数：.tap 里有声明、.4gl 里以函数落地，
// 没有 #add-point 标记（行号只能按函数名定位）。
type matFunc struct {
	Name string
	Body []string
	// Src 声明身份（空 = c）。标准侧也可能有 function.* 声明，用来区分两种落地口径。
	Src string
}

// matSlice 是一条切片标记（`#應用 aXX 樣板自動產生(Version:N)`）。
type matSlice struct {
	Code string
	// Used 产物里的 Version；Current 当前模板首行的 Version（>0 时构建器会写 mdl 模板）。
	Used    int
	Current int
	// InBody=true 时把标记写在第一个槽位正文里（At 的 in-point 切片归属要用）。
	InBody bool
}

// matFdPage 是 .4fd 里的 `<Page text=…><Table name=…>` 桥。
type matFdPage struct {
	// Name 是 `<Page name=…>`（控件 id，故意与 Key 不同以钉住"取 text 不取 name"）。
	Name string
	Key  string // text（.tsd 里 page_* 的键）
	// Record 是 `<Table name=…>`（.tab 的 screen record）。
	Record string
}

// matOpts 是一套材料的全部开关。
//
// 设计原则：**缺省形态就是"一套完整、两侧对齐、索引全绿"的材料**，测试只打开要钉的
// 那个变化（NoTap / Freestyle / 旧 4gl / 槽位不一致 …），这样每条测试的"为什么"都写在
// 开关上，而不是埋在几十行字符串里。
type matOpts struct {
	Prog string
	Env  string

	// 版型：Type=t01/q01/r01/p01…，Industry=std/_wf，Module=cxm…
	Type        string
	Industry    string
	Module      string
	SDVer       string
	TabVer      string
	CusVer      string
	Identity    string // .tap 根属性 identity（c=客制 / s=标准）
	SectionFlag string

	Pages []matPage // nil = 从槽位名自动推导（只推导 page<id>.<col> 形态）

	Sections           []matSection
	ExtraG4Sections    []matSection // 只有 .4gl 有的 section
	Orphans            []matOrphan
	Funcs              []matFunc
	Slices             []matSlice
	HeadColumns        []string          // .tab 的单头字段（不带 page，无法点名）
	TSDLabels          map[string]string // 列名 → 画面中文名（写 lbl_<列>）
	TSDPages           map[string]string // 页签键 → 标题（写 page_*/bpage_*）
	FDPages            []matFdPage       // nil = 从 Pages 自动造（Key=page_<id>）
	ManifestWarnings   []string
	ManifestMissing    []string
	ExtraFiles         map[string]string // 附加文件（相对程序目录）；manifest 里 role 按扩展名推
	TapMTime, G4MTime  string            // RFC3339；空 = 缺省（4gl 比 tap 新）
	TabMTime, TglMTime string

	NoTab  bool // 不写 .tab（蓝图缺件）
	NoTap  bool // 不写 .tap（从未被设计器接管的标准程序）
	NoTsd  bool
	No4fd  bool
	NoTgl  bool
	NoMask bool // 不写 inc/<prog>_mask.4gl
	// NoPages=true 表示"报表/批次这种没有画面页面的版型"：.tab 里一个 body 页都没有。
	// （不设时 Pages 为 nil 会从槽位名自动推导页面。）
	NoPages bool

	// Freestyle=true → .4gl 没有 section 锚（非生成器产物）。
	Freestyle bool
	// Unfolded=true → .4gl 头 5 行自述"未解開Section"（非 freestyle，只是没展开）。
	Unfolded bool
	// NonFreestyle=true → .4gl 头 5 行自述"非freestyle程式"（无 section 但也不判 freestyle）。
	NonFreestyle bool
}

// matSet 是构建结果：磁盘上的一套材料 + 依赖它的环境配置。
type matSet struct {
	t       *testing.T
	CfgPath string // config.json（Map/LoadIndex 走 CfgPath+Env+Prog 的路径解析）
	Env     string
	Prog    string
	Dir     string // 程序材料目录（<环境>/prog/<程序>）
	Mdl     string // 模板树根（<环境>/mdl；有切片时构建器会写 slice/code_*.template）

	files map[string]string // 相对程序目录的路径 → 内容（构建器实际写下的文件）
	opts  matOpts
}

// 缺省的源 mtime：4gl 比 tap 新（freshness=ok 的基准形态）。
const (
	matDefaultTapMTime = "2026-11-21T10:00:00+08:00"
	matDefaultG4MTime  = "2026-11-21T10:05:00+08:00"
	matDefaultTabMTime = "2026-11-21T10:06:00+08:00"
	matDefaultTglMTime = "2026-11-21T10:07:00+08:00"
)

var (
	reMatPagePoint = regexp.MustCompile(`\.page(\d+)\.(.+)$`)
	reMatPageID    = regexp.MustCompile(`\.page(\d+)\.`)
)

// newMaterials 在 t.TempDir() 下造一整套材料。
//
// 返回的 matSet 可以直接喂给 Map/LoadIndex/Show/At/Fields/Grep/Brief/Find：
//
//	set := newMaterials(t, matOpts{Prog: "axmt500_wf"})
//	doc := set.mapIndex(t)
//
// 需要改材料时用 set.writeFile / set.rewriteManifest，不要自己去 os.WriteFile。
func newMaterials(t *testing.T, opts matOpts) *matSet {
	t.Helper()
	opts = matNormalize(opts)
	set := &matSet{t: t, Env: opts.Env, Prog: opts.Prog, files: map[string]string{}, opts: opts}

	set.CfgPath = writeConfig(t, fmt.Sprintf(`{"name":%q,"host":"127.0.0.1","port":22,"user":"u"}`, opts.Env))
	dir, err := ProgDir(set.CfgPath, opts.Env, opts.Prog)
	if err != nil {
		t.Fatalf("ProgDir: %v", err)
	}
	set.Dir = dir
	mdl, err := MdlDir(set.CfgPath, opts.Env)
	if err != nil {
		t.Fatalf("MdlDir: %v", err)
	}
	set.Mdl = mdl

	pages := opts.Pages
	if opts.NoPages {
		pages = nil
	} else if len(pages) == 0 {
		pages = matAutoPages(opts.Sections)
		opts.Pages = pages
	}

	if !opts.NoTap {
		set.files["dzx/"+opts.Prog+".tap"] = matTapText(opts, pages)
	}
	if !opts.NoTab {
		set.files["dzx/"+opts.Prog+".tab"] = matTabText(opts, pages)
	}
	if !opts.NoTsd {
		set.files["dzx/"+opts.Prog+".tsd"] = matTsdText(opts)
	}
	if !opts.No4fd {
		set.files["dzx/"+opts.Prog+".4fd"] = mat4fdText(opts, pages)
	}
	if !opts.NoTgl {
		set.files["dzx/"+opts.Prog+".tgl"] = matTglText(opts)
	}
	set.files["dzx/"+opts.Prog+".tgx"] = "#+ tgx（功能键/工具条骨架，gen 侧只核对存在）\n"
	set.files["dzx/"+opts.Prog+".tgx2"] = "#+ tgx2（第二份骨架）\n"
	set.files["src/"+opts.Prog+".4gl"] = mat4glText(opts)
	if !opts.NoMask {
		set.files["inc/"+opts.Prog+"_mask.4gl"] = "#+ mask：画面掩码/检核片段（inc 角色）\n"
	}
	for rel, body := range opts.ExtraFiles {
		set.files[rel] = body
	}

	matWriteFiles(t, dir, set.files)
	matWriteMdl(t, mdl, opts.Slices)
	matWriteManifest(t, dir, opts, pages, set.files)
	return set
}

// newMinimalMaterials 是单页单槽位的最小成套材料（大多数测试的起点）。
func newMinimalMaterials(t *testing.T, prog string) *matSet {
	t.Helper()
	return newMaterials(t, matOpts{
		Prog: prog,
		Sections: []matSection{{
			ID: "p.input", Ver: "56", Src: "s",
			Slots: []matSlot{{
				Name: "input.a.page1.f1", Label: "過欄位檢查",
				DeclBody: "LET f1 = 1\n", Src: "c", Ver: "3",
				Body: []string{"   LET f1 = 1"},
			}},
		}},
		TSDLabels: map[string]string{"f1": "料号"},
	})
}

// matNormalize 补缺省值（让每条测试只写它关心的字段）。
func matNormalize(o matOpts) matOpts {
	if strings.TrimSpace(o.Prog) == "" {
		o.Prog = "axmt500_wf"
	}
	if strings.TrimSpace(o.Env) == "" {
		o.Env = "e"
	}
	if o.Type == "" {
		o.Type = "t01"
	}
	if o.Industry == "" {
		o.Industry = "std"
	}
	if o.Module == "" {
		o.Module = "cxm"
	}
	if o.SDVer == "" {
		o.SDVer = "0015"
	}
	if o.TabVer == "" {
		o.TabVer = "0001"
	}
	if o.CusVer == "" {
		o.CusVer = "0003"
	}
	if o.Identity == "" {
		o.Identity = "c"
	}
	if o.SectionFlag == "" {
		o.SectionFlag = "Y"
	}
	if o.TapMTime == "" {
		o.TapMTime = matDefaultTapMTime
	}
	if o.G4MTime == "" {
		o.G4MTime = matDefaultG4MTime
	}
	if o.TabMTime == "" {
		o.TabMTime = matDefaultTabMTime
	}
	if o.TglMTime == "" {
		o.TglMTime = matDefaultTglMTime
	}
	if len(o.Sections) == 0 && !o.Freestyle && !o.NonFreestyle {
		// 没给 section 时给一个最小的：单页单字段，保证"缺省形态全绿"。
		o.Sections = []matSection{{
			ID: "p.input", Ver: "56", Src: "s",
			Slots: []matSlot{{
				Name: "input.a.page1.f1", Label: "過欄位檢查",
				DeclBody: "LET f1 = 1\n", Src: "c", Ver: "3",
				Body: []string{"   LET f1 = 1"},
			}},
		}}
	}
	if o.TSDLabels == nil && o.TSDPages == nil {
		o.TSDLabels = map[string]string{"f1": "料号"}
	}
	return o
}

// matAutoPages 从槽位名推导页面（`input.a.page73.xmdduc001` → page_id 73 + 列 xmdduc001）：
// 这样"多页"只体现在点名的 page 号上，测试不用手写一遍页面清单。
func matAutoPages(sections []matSection) []matPage {
	byID := map[string][]string{}
	var order []string
	for _, sec := range sections {
		for _, sl := range sec.Slots {
			m := reMatPagePoint.FindStringSubmatch(sl.Name)
			if m == nil {
				continue
			}
			id, col := m[1], m[2]
			if _, ok := byID[id]; !ok {
				order = append(order, id)
			}
			if !matContains(byID[id], col) {
				byID[id] = append(byID[id], col)
			}
		}
	}
	if len(order) == 0 {
		return []matPage{{Seq: "1", ID: "1", Record: "s_detail1", Columns: []string{"f1", "f2"}}}
	}
	// 按 id 数值排序，页签序号从 1 起。
	sort.SliceStable(order, func(i, j int) bool {
		a, _ := strconv.Atoi(order[i])
		b, _ := strconv.Atoi(order[j])
		return a < b
	})
	out := make([]matPage, 0, len(order))
	for i, id := range order {
		out = append(out, matPage{
			Seq: strconv.Itoa(i + 1), ID: id,
			Record: "s_detail" + id, Columns: byID[id],
		})
	}
	return out
}

func matContains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// ---------- 各文件的文本构造（形态对齐 linkparse.go） ----------

// matTapText 造 `.tap`：根 add_points + 每个 section 的洞 + `<point>` 声明。
//
// CDATA 里刻意只放"人读骨架"：声明正文放在 <point> 里，因为 parseTap 用流式 decoder，
// 正文里的 `#add-point:` 字样不会被误当成洞（这条规则由 link_test 钉住）。
func matTapText(o matOpts, pages []matPage) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, `<add_points prog="%s" std_prog="%s" module="%s" identity="%s" section_flag="%s">`+"\n",
		o.Prog, o.Prog, o.Module, o.Identity, o.SectionFlag)

	// 声明（<point>）先写：与真机一致（设计器把它们都排在根下）。
	seenDecl := map[string]bool{}
	decl := func(name, mapping, src, ver, body string) {
		if name == "" || seenDecl[name] || strings.TrimSpace(body) == "" {
			return
		}
		seenDecl[name] = true
		fmt.Fprintf(&b, `<point name="%s" mapping="%s" src="%s" ver="%s" status=""><![CDATA[%s]]></point>`+"\n",
			name, mapping, src, ver, body)
	}
	for _, sec := range o.Sections {
		for _, sl := range sec.Slots {
			src, ver := sl.Src, sl.Ver
			if src == "" {
				src = "c"
			}
			decl(sl.Name, sl.Mapping, src, ver, sl.DeclBody)
		}
	}
	for _, or := range o.Orphans {
		decl(or.Name, or.Name, or.Src, or.Ver, or.Body)
	}
	for _, f := range o.Funcs {
		body := strings.Join(f.Body, "\n")
		if body == "" {
			body = "  LET x = 1"
		}
		src := f.Src
		if src == "" {
			src = "c"
		}
		decl("function."+f.Name, "function."+f.Name, src, "2", body+"\n")
	}

	for _, sec := range o.Sections {
		fmt.Fprintf(&b, `<section id="%s" ver="%s" src="%s"`, sec.ID, sec.Ver, sec.Src)
		if sec.Type != "" {
			fmt.Fprintf(&b, ` type="%s"`, sec.Type)
		}
		if sec.Readonly != "" {
			fmt.Fprintf(&b, ` readonly="%s"`, sec.Readonly)
		}
		b.WriteString("><![CDATA[\n")
		for _, sl := range sec.Slots {
			if sl.G4Only {
				continue
			}
			matWriteTapSlot(&b, sl)
		}
		b.WriteString("]]></section>\n")
	}
	b.WriteString("</add_points>\n")
	return b.String()
}

// matWriteTapSlot 写一个洞：标记行 + `{<point name="…"/>}`。
func matWriteTapSlot(b *strings.Builder, sl matSlot) {
	markerName := sl.Name
	if sl.TruncatedMarker {
		markerName = matTruncatedName(sl.Name)
	}
	label := sl.Label
	if label == "" {
		label = sl.Name
	}
	line := fmt.Sprintf(`#add-point:%s name="%s"`, label, markerName)
	hole := fmt.Sprintf(`   {<point name="%s"/>}`, sl.Name)
	end := "   #end add-point"
	if sl.Commentd {
		line = "#            " + line
		hole = "#            " + hole
		end = "#            #end add-point"
	}
	b.WriteString(line + "\n" + hole + "\n" + end + "\n")
}

// matTruncatedName 复刻真机的截断形态（`chk_data_exist.define` → `chk_`）：
// 标记行的 name 值只写到第 4 个字符，完整名字只在洞里。
func matTruncatedName(name string) string {
	n := 4
	if i := strings.IndexByte(name, '.'); i > 0 && i < n {
		n = i
	}
	if n > len(name) {
		n = len(name)
	}
	return name[:n]
}

// matTabText 造 `.tab`：assembly + structure + 每个 body 的字段清单。
//
// 故意带一个 `<32wc …>`（以数字开头的标签名）：真机上存在，parseTab 必须容忍，
// 严格 XML 解析器会整份失败。
func matTabText(o matOpts, pages []matPage) string {
	var b strings.Builder
	b.WriteString(`<?xml version='1.0' encoding='UTF-8'?>` + "\n")
	fmt.Fprintf(&b, `<assembly name="%s" module="%s" type="%s" industry="%s" page="%d" sdver="%s" cusver="%s" tabver="%s" description="%s">`+"\n",
		o.Prog, o.Module, o.Type, o.Industry, len(pages), o.SDVer, o.CusVer, o.TabVer, o.Prog+" 測試作業")
	b.WriteString("  <structure>\n")
	if len(o.HeadColumns) > 0 {
		fmt.Fprintf(&b, `    <var id="head" value="%s"/>`+"\n", strings.Join(o.HeadColumns, ","))
	}
	for _, pg := range pages {
		cols := make([]string, 0, len(pg.Columns))
		for i, c := range pg.Columns {
			// 第 1 列带展延来源、第 2 列带类型：钉住 splitTabColumns 去括号的规则。
			switch i {
			case 0:
				cols = append(cols, c+"(chr500)")
			case 1:
				cols = append(cols, c+"(inayl_t.inayl003)")
			default:
				cols = append(cols, c)
			}
		}
		fmt.Fprintf(&b, `    <var id="body" page="%s" page_id="%s" record="%s" value="%s"/>`+"\n",
			pg.Seq, pg.ID, pg.Record, strings.Join(cols, ","))
	}
	b.WriteString(`    <32wc id="dummy" value="1"/>` + "\n")
	b.WriteString("  </structure>\n</assembly>\n")
	return b.String()
}

// matTsdText 造 `.tsd`：`lbl_<列>` 字段文案 + `page_*` 页签标题。
func matTsdText(o matOpts) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n<spec><strings>\n")
	cols := make([]string, 0, len(o.TSDLabels))
	for c := range o.TSDLabels {
		cols = append(cols, c)
	}
	sort.Strings(cols)
	for _, c := range cols {
		fmt.Fprintf(&b, `  <sfield name="lbl_%s" text="%s" />`+"\n", c, o.TSDLabels[c])
	}
	keys := make([]string, 0, len(o.TSDPages))
	for k := range o.TSDPages {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, `  <sfield name="%s" text="%s" />`+"\n", k, o.TSDPages[k])
	}
	b.WriteString(`  <sfield name="btn_ok" text="确认" />` + "\n")
	b.WriteString("</strings></spec>\n")
	return b.String()
}

// mat4fdText 造 `.4fd`：`<Page name=… text=…><Table name=…/></Page>`。
//
// 关键形态：text 才是页签键，name 是控件 id（真机备注页 name="page_1" text="page_12"，
// 取错 name 会把页签接成"基本数据"）。缺省 Key=page_<page_id>。
func mat4fdText(o matOpts, pages []matPage) string {
	fd := o.FDPages
	if fd == nil {
		fd = make([]matFdPage, 0, len(pages))
		for _, pg := range pages {
			fd = append(fd, matFdPage{Name: "page_" + pg.Seq, Key: "page_" + pg.ID, Record: pg.Record})
		}
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n<ManagedForm name=\"ManagedForm\">\n")
	for _, p := range fd {
		name := p.Name
		if name == "" {
			name = "ctrl_" + p.Record
		}
		fmt.Fprintf(&b, `  <Page name="%s" text="%s">`+"\n", name, p.Key)
		fmt.Fprintf(&b, `    <Table name="%s"/>`+"\n", p.Record)
		b.WriteString("  </Page>\n")
	}
	b.WriteString("</ManagedForm>\n")
	return b.String()
}

// matTglText 造 `.tgl`：头部带 SD/PR 版次与 Build/Description（ParseTglHeader 的形态）。
func matTglText(o matOpts) string {
	return "#+" + " Standard Version.....: SD版次:" + o.SDVer + "(2025-11-13 17:33:55), PR版次:\n" +
		"#+ Customerized Version.: SD版次:0000(1900-01-01 00:00:00), PR版次:\n" +
		"#+ Build......: 000037\n" +
		"#+ Description: " + o.Prog + " 測試作業\n" +
		"#+ section 排序骨架（測試夾具）\n"
}

// mat4glText 造 `.4gl`：section 锚 + `#add-point` 块 + 切片标记（+ 客制函数）。
func mat4glText(o matOpts) string {
	var lines []string
	switch {
	case o.Unfolded:
		lines = append(lines, "#該程式未解開Section, 採用最新樣板產出!")
		lines = append(lines, "#該程式非freestyle程式!")
	case o.NonFreestyle:
		lines = append(lines, "#該程式非freestyle程式!")
	}

	// 切片标记：默认写在第一个可承载的槽位正文里（这样 At 能报 in-point）；
	// InBody=false 的写在顶层（用于钉住 before-line 关系）。
	var bodySlices, topSlices []matSlice
	for _, s := range o.Slices {
		if s.InBody {
			bodySlices = append(bodySlices, s)
		} else {
			topSlices = append(topSlices, s)
		}
	}
	for _, s := range topSlices {
		lines = append(lines, fmt.Sprintf("#應用 %s 樣板自動產生(Version:%d)", s.Code, s.Used))
	}

	if o.Freestyle {
		// 非生成器产物：没有 section 锚，也就没有任何加点可归。
		lines = append(lines,
			"PROGRAM "+o.Prog,
			"MAIN",
			"  DISPLAY 'freestyle'",
			"END MAIN")
		lines = append(lines, matFuncLines(o.Funcs)...)
		return strings.Join(lines, "\n") + "\n"
	}

	firstBodyUsed := false
	for _, sec := range o.Sections {
		lines = append(lines, fmt.Sprintf(`{<section id="%s" type="s" >}`, sec.ID))
		for _, sl := range sec.Slots {
			if sl.TapOnly {
				continue
			}
			body := append([]string(nil), sl.Body...)
			if !firstBodyUsed && len(bodySlices) > 0 {
				var marks []string
				for _, s := range bodySlices {
					marks = append(marks, fmt.Sprintf("#應用 %s 樣板自動產生(Version:%d)", s.Code, s.Used))
				}
				body = append(marks, body...)
				firstBodyUsed = true
			}
			label := sl.Label
			if label == "" {
				label = sl.Name
			}
			g4Name := sl.Name
			if sl.TruncatedMarker {
				g4Name = matTruncatedName(sl.Name)
			}
			slotLines := []string{fmt.Sprintf(`   #add-point:%s name="%s"`, label, g4Name)}
			slotLines = append(slotLines, body...)
			slotLines = append(slotLines, "   #END add-point")
			if sl.Commentd {
				for i := range slotLines {
					slotLines[i] = "#            " + slotLines[i]
				}
			}
			lines = append(lines, slotLines...)
		}
		lines = append(lines, `{</section>}`)
	}
	for _, sec := range o.ExtraG4Sections {
		lines = append(lines, fmt.Sprintf(`{<section id="%s" type="s" >}`, sec.ID))
		for _, sl := range sec.Slots {
			lines = append(lines, fmt.Sprintf(`   #add-point:%s name="%s"`, sl.Label, sl.Name))
			lines = append(lines, sl.Body...)
			lines = append(lines, "   #END add-point")
		}
		lines = append(lines, `{</section>}`)
	}
	lines = append(lines, matFuncLines(o.Funcs)...)
	if len(lines) == 0 {
		lines = append(lines, "PROGRAM "+o.Prog, "")
	}
	return strings.Join(lines, "\n") + "\n"
}

// matFuncLines 把客制函数写成 `.4gl` 里真实存在的 FUNCTION … END FUNCTION。
func matFuncLines(fs []matFunc) []string {
	var out []string
	for _, f := range fs {
		out = append(out, "", "PRIVATE FUNCTION "+f.Name+"()")
		out = append(out, f.Body...)
		out = append(out, "END FUNCTION")
	}
	return out
}

// ---------- 落盘 ----------

// matWriteFiles 写所有文件（自动建目录）。
func matWriteFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	rels := make([]string, 0, len(files))
	for rel := range files {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	for _, rel := range rels {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("建目录 %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(files[rel]), 0o644); err != nil {
			t.Fatalf("写材料 %s: %v", rel, err)
		}
	}
}

// matWriteMdl 把切片当前版本写成模板树里的 `slice/code_<code>.template`
// （Map 的脱版检测读它；没有 Current 的切片不写）。
func matWriteMdl(t *testing.T, mdl string, slices []matSlice) {
	t.Helper()
	for _, s := range slices {
		if s.Current <= 0 {
			continue
		}
		p := filepath.Join(mdl, "slice", "code_"+s.Code+".template")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("建模板目录: %v", err)
		}
		body := fmt.Sprintf("#應用 %s 樣板自動產生(Version:%d)\n# 模板正文（測試夾具）\n", s.Code, s.Current)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("写模板: %v", err)
		}
	}
}

// matRoleOf 由本地相对路径推 manifest role（与 FetchRoles 的落点一一对应）。
func matRoleOf(rel string) string {
	base := pathBase(rel)
	switch {
	case strings.HasSuffix(base, ".tab"):
		return "tab"
	case strings.HasSuffix(base, ".tsd"):
		return "tsd"
	case strings.HasSuffix(base, ".4fd"):
		return "4fd"
	case strings.HasSuffix(base, ".tap"):
		return "tap"
	case strings.HasSuffix(base, ".tgl"):
		return "tgl"
	case strings.HasSuffix(base, ".tgx2"):
		return "tgx2"
	case strings.HasSuffix(base, ".tgx"):
		return "tgx"
	case strings.HasSuffix(base, ".4gl") && strings.Contains(base, "_mask"):
		return "inc"
	case strings.HasSuffix(base, ".4gl"):
		return "4gl"
	}
	return "other"
}

func pathBase(p string) string {
	if i := strings.LastIndexAny(p, `/\`); i >= 0 {
		return p[i+1:]
	}
	return p
}

// matSrcOf 按 FetchRoles 的源路径模板算出 manifest 里的 src（服务器侧相对 $TOP）。
// 用它而不是 `erp/<mod>/<rel>`：status/check 的 probe 是拿 src 去服务器 Stat 的，
// 路径形态错了 "服务器上已没有" 会整片误报。
func matSrcOf(module, prog, role, rel string) string {
	mod := strings.ToLower(module)
	switch role {
	case "tab":
		return fmt.Sprintf("erp/%s/dzx/tab/%s.tab", mod, prog)
	case "tsd":
		return fmt.Sprintf("erp/%s/dzx/tsd/%s.tsd", mod, prog)
	case "4fd":
		return fmt.Sprintf("erp/%s/4fd/%s.4fd", mod, prog)
	case "tap":
		return fmt.Sprintf("erp/%s/dzx/tap/%s.tap", mod, prog)
	case "tgl":
		return fmt.Sprintf("erp/%s/dzx/tgl/%s.tgl", mod, prog)
	case "tgx":
		return fmt.Sprintf("erp/%s/dzx/tgl/%s.tgx", mod, prog)
	case "tgx2":
		return fmt.Sprintf("erp/%s/dzx/tgl/%s.tgx2", mod, prog)
	case "4gl":
		return fmt.Sprintf("erp/%s/4gl/%s.4gl", mod, prog)
	case "inc":
		return fmt.Sprintf("com/inc/erp/%s/%s_mask.4gl", mod, prog)
	}
	return "erp/" + mod + "/" + rel
}

// matWriteManifest 写 manifest.json（sha256 + srcMtime 都由构建器算，保证与文件一致）。
func matWriteManifest(t *testing.T, dir string, o matOpts, pages []matPage, files map[string]string) {
	t.Helper()
	mf := NewManifest("fetch", o.Env)
	mf.Prog = o.Prog
	mf.Module = o.Module
	mf.FetchedAt = "2026-11-21T10:10:00+08:00"
	mf.Source = Source{TopDir: "/u1/t35tst", Zone: "tst", Ent: 1}
	for _, s := range o.Slices {
		if s.Current > 0 {
			mf.Versions.TemplateVersions[s.Code] = s.Current
		}
	}
	mtimeOf := func(rel, role string) string {
		switch role {
		case "tab":
			return o.TabMTime
		case "tgl", "tgx", "tgx2":
			return o.TglMTime
		case "4gl":
			return o.G4MTime
		case "inc":
			return o.TglMTime
		default:
			return o.TapMTime
		}
	}
	rels := make([]string, 0, len(files))
	for rel := range files {
		rels = append(rels, rel)
	}
	sort.Strings(rels)
	for _, rel := range rels {
		role := matRoleOf(rel)
		mf.Files = append(mf.Files, FileEntry{
			Role:     role,
			Src:      matSrcOf(o.Module, o.Prog, role, rel),
			Rel:      rel,
			SHA256:   shaOfString(files[rel]),
			Size:     int64(len(files[rel])),
			SrcMtime: mtimeOf(rel, role),
		})
	}
	mf.Missing = append(mf.Missing, o.ManifestMissing...)
	mf.Warnings = append(mf.Warnings, o.ManifestWarnings...)
	if err := WriteManifest(dir, mf); err != nil {
		t.Fatalf("写 manifest: %v", err)
	}
	_ = pages
}

// ---------- matSet 的便捷操作 ----------

// LinkRequest 返回按 CfgPath+Env+Prog 解析的请求（走真实的路径规则）。
func (m *matSet) LinkRequest() LinkRequest {
	return LinkRequest{CfgPath: m.CfgPath, Env: m.Env, Prog: m.Prog}
}

// mapIndex 建索引（link.json 落到材料目录里）。
func (m *matSet) mapIndex() *LinkDoc {
	m.t.Helper()
	doc, err := Map(m.LinkRequest())
	if err != nil {
		m.t.Fatalf("Map: %v", err)
	}
	return doc
}

// loadIndex 复用/重建索引。
func (m *matSet) loadIndex() *LinkDoc {
	m.t.Helper()
	doc, err := LoadIndex(m.LinkRequest())
	if err != nil {
		m.t.Fatalf("LoadIndex: %v", err)
	}
	return doc
}

// filePath 返回材料目录里一个相对路径的绝对路径。
func (m *matSet) filePath(rel string) string {
	return filepath.Join(m.Dir, filepath.FromSlash(rel))
}

// writeFile 改一个材料文件并同步刷新 manifest 里的 sha/size（否则 Map 的 Sources
// 与磁盘不一致，freshness/索引复用判定的前提就假了）。
func (m *matSet) writeFile(rel, body string) {
	m.t.Helper()
	m.files[rel] = body
	p := m.filePath(rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		m.t.Fatalf("建目录: %v", err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		m.t.Fatalf("写文件 %s: %v", rel, err)
	}
	m.rewriteManifest(func(mf *Manifest) {})
}

// removeFile 删一个材料文件并从 manifest 摘掉它（模拟"这套材料本来就没有"）。
func (m *matSet) removeFile(rel string) {
	m.t.Helper()
	delete(m.files, rel)
	if err := os.Remove(m.filePath(rel)); err != nil && !os.IsNotExist(err) {
		m.t.Fatalf("删文件 %s: %v", rel, err)
	}
	m.rewriteManifest(func(mf *Manifest) {})
}

// rewriteManifest 读回 manifest、让 mutate 改完后重写（重新对齐 sha/size/srcMtime）。
func (m *matSet) rewriteManifest(mutate func(*Manifest)) {
	m.t.Helper()
	p := filepath.Join(m.Dir, ManifestFileName)
	b, err := os.ReadFile(p)
	if err != nil {
		m.t.Fatalf("读 manifest: %v", err)
	}
	var mf Manifest
	if err := json.Unmarshal(b, &mf); err != nil {
		m.t.Fatalf("manifest 不是 JSON: %v", err)
	}
	var kept []FileEntry
	for _, f := range mf.Files {
		body, ok := m.files[f.Rel]
		if !ok {
			if _, serr := os.Stat(m.filePath(f.Rel)); serr != nil {
				continue
			}
			raw, rerr := os.ReadFile(m.filePath(f.Rel))
			if rerr != nil {
				continue
			}
			body = string(raw)
			m.files[f.Rel] = body
		}
		f.SHA256 = shaOfString(body)
		f.Size = int64(len(body))
		kept = append(kept, f)
	}
	mf.Files = kept
	if mutate != nil {
		mutate(&mf)
	}
	if err := WriteManifest(m.Dir, &mf); err != nil {
		m.t.Fatalf("重写 manifest: %v", err)
	}
}

// readManifest 读回 manifest 用于断言。
func (m *matSet) readManifest() *Manifest {
	m.t.Helper()
	mf, err := ReadManifest(m.Dir)
	if err != nil || mf == nil {
		m.t.Fatalf("ReadManifest: %v / %+v", err, mf)
	}
	return mf
}

// readLinkJSON 读回落盘的 link.json。
func (m *matSet) readLinkJSON() *LinkDoc {
	m.t.Helper()
	p := filepath.Join(m.Dir, linkJSONName)
	b, err := os.ReadFile(p)
	if err != nil {
		m.t.Fatalf("读 link.json: %v", err)
	}
	var d LinkDoc
	if err := json.Unmarshal(b, &d); err != nil {
		m.t.Fatalf("link.json 不是合法 JSON: %v", err)
	}
	d.Path = p
	return &d
}

// linkJSONPath 索引落点。
func (m *matSet) linkJSONPath() string { return filepath.Join(m.Dir, linkJSONName) }

// checkByID 取一条检查结果（没有返回 nil）。
func (m *matSet) checkByID(doc *LinkDoc, id string) *LinkCheck {
	m.t.Helper()
	return checkByID(doc, id)
}

// pointNames 把点位名按出现顺序列出来（断言用）。
func (m *matSet) pointNames(doc *LinkDoc) []string {
	out := make([]string, 0, len(doc.Points))
	for _, p := range doc.Points {
		out = append(out, p.Name)
	}
	return out
}

// matIndexOf 在字符串列表里找一个值的下标（没有返回 -1）。
func matIndexOf(list []string, v string) int {
	for i, s := range list {
		if s == v {
			return i
		}
	}
	return -1
}

// matMustTime 解析 RFC3339 或直接 fail（测试里用它给 mtime 断言用）。
func matMustTime(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("时间 %q 不是 RFC3339: %v", s, err)
	}
	return tm
}
