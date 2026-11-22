package devdata

// gen map 的测试矩阵。
//
// 覆盖的是"真机上会遇到的形态"，不是"理想的 XML/4gl"：
// 名字被截断、槽位被注释掉、重名、只有一侧存在、没有 .tap、没有 .tab、
// .tab 里有非法标签名、报表/批次没有页面字段、freestyle 没有 section。
//
// 每个用例都对着一条现实证据写（括号里是实测出处），改实现时先看这些用例为什么存在。

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/pkgfile"
)

// shaOfString 是 manifest 里记的 sha256。
func shaOfString(s string) string { return pkgfile.Sha256Hex([]byte(s)) }

// ---------- 夹具 ----------

func tapDocText(attrs, body string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<add_points prog="p" std_prog="p" module="CXM" identity="c" section_flag="N" ` + attrs + `>` + "\n" +
		body + "\n</add_points>\n"
}

func mkTapSection(id, ver string, inner ...string) string {
	return `<section id="` + id + `" ver="` + ver + `" src="s"><![CDATA[` + strings.Join(inner, "") + `]]></section>` + "\n"
}

func mkTapSlot(name, label string) string {
	return "#add-point:" + label + ` name="` + name + `"` + "\n" +
		`   {<point name="` + name + `"/>}` + "\n" +
		"   #end add-point\n"
}

// mkTapSlotTruncated 复刻实测的截断标记：标记行 name 只写到 chk_，完整的名字在洞里。
func mkTapSlotTruncated(label, holeName string) string {
	return "#add-point:" + label + ` name="chk_` + "\n" +
		`   {<point name="` + holeName + `"/>}` + "\n" +
		"   #end add-point\n"
}

func mkTapDecl(name, mapping, src, ver, body string) string {
	return `<point name="` + name + `" mapping="` + mapping + `" src="` + src + `" ver="` + ver + `" status="">` +
		"<![CDATA[" + body + "]]></point>\n"
}

func mkG4Section(id string, inner ...string) []string {
	out := []string{`{<section id="` + id + `" type="s" >}`}
	out = append(out, inner...)
	return append(out, `{</section>}`, "")
}

func mkG4Slot(label, name string, body ...string) []string {
	out := []string{"   #add-point:" + label + ` name="` + name + `"`}
	out = append(out, body...)
	return append(out, "   #END add-point", "")
}

func mkG4SlotCommented(label, name string) []string {
	return []string{"#            #add-point:" + label + ` name="` + name + `"`,
		"#            {<point name=\"" + name + "\"/>}", "#            #end add-point"}
}

func tabDocText(assembly, body string) string {
	return `<?xml version='1.0' encoding='UTF-8'?>` + "\n" +
		`<assembly ` + assembly + `>` + "\n  <structure>\n" + body + "\n  </structure>\n</assembly>"
}

// parseAll 是三个解析器的便捷组合。
func parseAll(t *testing.T, tapStr, g4Str, tabStr string) (*tapDoc, *g4Doc, *tabDoc) {
	t.Helper()
	var tap *tapDoc
	if tapStr != "" {
		var err error
		if tap, err = parseTap([]byte(tapStr)); err != nil {
			t.Fatalf("parseTap: %v", err)
		}
	} else {
		tap = &tapDoc{Attrs: map[string]string{}}
	}
	g4 := parse4gl(splitLines([]byte(g4Str)))
	var tab *tabDoc
	if tabStr != "" {
		var err error
		if tab, err = parseTab([]byte(tabStr)); err != nil {
			t.Fatalf("parseTab: %v", err)
		}
	}
	return tap, g4, tab
}

func linkOf(t *testing.T, tapStr, g4Str, tabStr string) *LinkDoc {
	t.Helper()
	tap, g4, tab := parseAll(t, tapStr, g4Str, tabStr)
	return buildLink(LinkRequest{Env: "e", Prog: "p"}, tap, g4, tab)
}

func pointsByName(doc *LinkDoc, name string) []LinkPoint {
	var out []LinkPoint
	for _, p := range doc.Points {
		if p.Name == name {
			out = append(out, p)
		}
	}
	return out
}

func checkByID(doc *LinkDoc, id string) *LinkCheck {
	for i := range doc.Checks {
		if doc.Checks[i].ID == id {
			return &doc.Checks[i]
		}
	}
	return nil
}

// ---------- .tap 解析 ----------

func TestParseTap_SlotsDeclsAndSections(t *testing.T) {
	tapStr := tapDocText("", mkTapDecl("s.a", "", "c", "3", "LET a = 1\nLET b = 2\n")+
		mkTapSection("p.input", "56", mkTapSlot("input.a.page1.f1", "過欄位檢查")+mkTapSlot("input.b.page1.f1", "BEFORE FIELD")))
	tap, err := parseTap([]byte(tapStr))
	if err != nil {
		t.Fatalf("parseTap: %v", err)
	}
	if tap.Attrs["module"] != "CXM" || tap.Attrs["identity"] != "c" {
		t.Errorf("根属性没读到: %v", tap.Attrs)
	}
	if len(tap.Sections) != 1 || tap.Sections[0].ID != "p.input" {
		t.Fatalf("sections = %+v", tap.Sections)
	}
	sl := tap.Sections[0].Slots
	if len(sl) != 2 || sl[0].Name != "input.a.page1.f1" || sl[1].Name != "input.b.page1.f1" {
		t.Fatalf("槽位 = %+v", sl)
	}
	if sl[0].Label != "過欄位檢查" {
		t.Errorf("标签 = %q", sl[0].Label)
	}
	if sl[0].Line <= 1 {
		t.Errorf("行号没算出来: %d", sl[0].Line)
	}
	if len(tap.Decls) != 1 {
		t.Fatalf("声明数 = %d", len(tap.Decls))
	}
	d := tap.Decls[0]
	if d.Src != "c" || d.Ver != "3" || d.Lines != 2 || d.FirstUse != "LET a = 1" {
		t.Errorf("声明 = %+v", d)
	}
}

// 实测：.tap 里 `#add-point:… name="chk_` 被截断，完整名字只在洞里。
func TestParseTap_TruncatedMarkerUsesHoleName(t *testing.T) {
	tapStr := tapDocText("", mkTapSection("p.s", "1",
		mkTapSlotTruncated("chk_data_exist段define (請盡量不要修改)", "chk_data_exist.define")))
	tap, err := parseTap([]byte(tapStr))
	if err != nil {
		t.Fatalf("parseTap: %v", err)
	}
	sl := tap.Sections[0].Slots
	if len(sl) != 1 {
		t.Fatalf("槽位 = %+v", sl)
	}
	if sl[0].Name != "chk_data_exist.define" {
		t.Errorf("名字应取洞里的完整名，得到 %q", sl[0].Name)
	}
	if !strings.Contains(sl[0].Label, "chk_data_exist段define") {
		t.Errorf("标签应取标记行，得到 %q", sl[0].Label)
	}
}

// 声明里可能出现 `#add-point:` 字样（内容本身是代码/注释），不能被当成槽位。
func TestParseTap_IgnoresMarkerTextInsideContent(t *testing.T) {
	body := "   #add-point:看起來像槽位 name=\"bogus\"\n   LET x = 1\n"
	tapStr := tapDocText("", mkTapDecl("input.a.page1.f1", "", "s", "1", body)+
		mkTapSection("p.s", "1", mkTapSlot("input.a.page1.f1", "真槽位")))
	tap, err := parseTap([]byte(tapStr))
	if err != nil {
		t.Fatalf("parseTap: %v", err)
	}
	sl := tap.Sections[0].Slots
	if len(sl) != 1 || sl[0].Name != "input.a.page1.f1" {
		t.Fatalf("槽位 = %+v", sl)
	}
}

// ---------- .4gl 解析 ----------

func TestParse4gl_SectionsSlotsCommentedAndSlices(t *testing.T) {
	lines := []string{
		"#該程式未解開Section, 採用最新樣板產出!",
		"#該程式非freestyle程式!",
	}
	lines = append(lines, mkG4Section("p.main",
		"#應用 a00 樣板自動產生(Version:3)",
		"   #add-point:主程式 name=\"main.run\"",
		"   LET x = 1",
		"   #END add-point")...)
	inner := append([]string{"#應用 a05 樣板自動產生(Version:2)"},
		mkG4SlotCommented("註解掉的", "construct.c.page1.x")...)
	lines = append(lines, mkG4Section("p.input", inner...)...)
	lines = append(lines, mkG4Section("p.input",
		mkG4Slot("小寫end", "input.a.page1.f1", "   LET y = 2")...)...)

	d := parse4gl(lines)
	if !d.UnfoldedSection || d.Freestyle {
		t.Errorf("头部自述解析错: %+v", d)
	}
	if len(d.Sections) != 2 {
		t.Fatalf("sections = %d，想要 2（同名 section 合并）", len(d.Sections))
	}
	if d.Sections[0].ID != "p.main" || len(d.Sections[0].Slots) != 1 {
		t.Fatalf("首段 = %+v", d.Sections[0])
	}
	s := d.Sections[0].Slots[0]
	if s.Line0 >= s.Line1 {
		t.Errorf("行区间 = %d-%d，应包含 #add-point 与 #END", s.Line0, s.Line1)
	}
	if !s.HasBody {
		t.Error("槽位里有代码，HasBody 应为 true")
	}
	c := d.Sections[1].Slots[0]
	if !c.Commentd {
		t.Error("注释掉的槽位应标记 commented")
	}
	if len(d.Slices) != 2 {
		t.Fatalf("切片 = %+v", d.Slices)
	}
	if d.Slices[0].Code != "a00" || d.Slices[0].Ver != 3 {
		t.Errorf("切片[0] = %+v", d.Slices[0])
	}
}

func TestParse4gl_NoSectionsIsFreestyle(t *testing.T) {
	d := parse4gl(splitLines([]byte("PROGRAM p\nMAIN\n  DISPLAY 1\nEND MAIN\n")))
	if !d.Freestyle {
		t.Error("没有 section 锚且没自述 → 应判 freestyle")
	}
	if len(d.Sections) != 0 {
		t.Errorf("sections = %d", len(d.Sections))
	}
}

// ---------- .tab 解析 ----------

// 实测：真机 .tab 里有 `<32wc …>` 这种以数字开头的标签名，严格 XML 会报错。
func TestParseTab_TolerantTagNameAndPageID(t *testing.T) {
	tabStr := tabDocText(`name="p" module="cxm" type="t01" industry="std" page="2" sdver="0015" cusver="0003" description="訂單"`,
		`<var id="head" value="a1,a2(ooag_t.ooag011),a3(chr80)"/>`+"\n"+
			`<var id="body" page="12" page_id="73" record="s_detail73" value="x1(chr500),x2(inayl_t.inayl003),x3"/>`+"\n"+
			`<var id="body" page="1" page_id="1" record="s_detail1" value="y1"/>`+"\n"+
			`<32wc id="x" value="1"/>`)
	tab, err := parseTab([]byte(tabStr))
	if err != nil {
		t.Fatalf("parseTab: %v", err)
	}
	if tab.Assembly.Type != "t01" || tab.Assembly.CusVer != "0003" {
		t.Errorf("assembly = %+v", tab.Assembly)
	}
	if len(tab.Pages) != 2 {
		t.Fatalf("pages = %+v", tab.Pages)
	}
	if tab.Pages[0].Page != "12" || tab.Pages[0].PageID != "73" || tab.Pages[0].Record != "s_detail73" {
		t.Errorf("页 0 = %+v（page 是页签序号，page_id 才是点名用的）", tab.Pages[0])
	}
	if got, want := strings.Join(tab.Pages[0].Columns, ","), "x1,x2,x3"; got != want {
		t.Errorf("字段 = %s，想要 %s（括号里的类型/展延要去掉）", got, want)
	}
	if got, want := strings.Join(tab.Head, ","), "a1,a2,a3"; got != want {
		t.Errorf("单头字段 = %s，想要 %s", got, want)
	}
}

// ---------- 配对 ----------

func TestBuildLink_PairsByNameAndAttachesDecl(t *testing.T) {
	tapStr := tapDocText("", mkTapDecl("input.a.page1.f1", "", "c", "3", "LET f1 = 1\n")+
		mkTapDecl("input.b.page1.f1", "", "s", "9", "LET std = 1\n")+
		mkTapSection("p.input", "56",
			mkTapSlot("input.a.page1.f1", "過欄位")+
				mkTapSlot("input.b.page1.f1", "BEFORE FIELD")+
				mkTapSlot("input.c.page1.f2", "未声明")))
	var inner []string
	inner = append(inner, mkG4Slot("過欄位", "input.a.page1.f1", "  LET f1 = 1")...)
	inner = append(inner, mkG4Slot("BEFORE FIELD", "input.b.page1.f1")...)
	inner = append(inner, mkG4Slot("未声明", "input.c.page1.f2")...)
	g4 := mkG4Section("p.input", inner...)

	doc := linkOf(t, tapStr, strings.Join(g4, "\n"), "")
	if doc.Stats.Points != 3 {
		t.Fatalf("点位 = %d，想要 3：%+v", doc.Stats.Points, doc.Points)
	}
	p := pointsByName(doc, "input.a.page1.f1")
	if len(p) != 1 {
		t.Fatalf("按名字找到 %d 个", len(p))
	}
	if !p[0].Custom || p[0].Src != "c" || p[0].DeclLines != 1 || p[0].Line0 == 0 {
		t.Errorf("客制点属性 = %+v", p[0])
	}
	if doc.Stats.PointsCustom != 1 || doc.Stats.CustomLines != 1 {
		t.Errorf("统计 = %+v", doc.Stats)
	}
	if c := checkByID(doc, "slots-pair"); c == nil || c.Level != "ok" {
		t.Errorf("slots-pair = %+v", c)
	}
	if c := checkByID(doc, "sections-pair"); c == nil || c.Level != "ok" {
		t.Errorf("sections-pair = %+v", c)
	}
}

// 实测：`construct.c.page1.xmdc027` 在 .4gl 里出现 5 次（3 个有效 + 2 个被注释），
// 而 .tap 只有 3 个洞 —— 多出来的必须记成 extra，不能算成 missing。
func TestBuildLink_DuplicateNamesAndCommentedExtras(t *testing.T) {
	tapStr := tapDocText("", mkTapSection("p.construct", "",
		mkTapSlot("construct.c.page1.x", "")+mkTapSlot("construct.c.page1.x", "")+mkTapSlot("construct.c.page1.x", "")))
	var g4 []string
	g4 = append(g4, `{<section id="p.construct" type="s" >}`)
	g4 = append(g4, mkG4Slot("a", "construct.c.page1.x")...)
	g4 = append(g4, mkG4SlotCommented("b", "construct.c.page1.x")...)
	g4 = append(g4, mkG4Slot("c", "construct.c.page1.x")...)
	g4 = append(g4, mkG4SlotCommented("d", "construct.c.page1.x")...)
	g4 = append(g4, mkG4Slot("e", "construct.c.page1.x")...)
	g4 = append(g4, `{</section>}`)

	doc := linkOf(t, tapStr, strings.Join(g4, "\n"), "")
	if doc.Stats.Points != 5 {
		t.Fatalf("点位 = %d，想要 5", doc.Stats.Points)
	}
	if doc.Stats.PointsExtra != 2 || doc.Stats.PointsMissing != 0 {
		t.Errorf("统计 extra=%d missing=%d，想要 2/0", doc.Stats.PointsExtra, doc.Stats.PointsMissing)
	}
	if doc.Stats.PointsCommented != 2 {
		t.Errorf("注释点 = %d，想要 2", doc.Stats.PointsCommented)
	}
}

func TestBuildLink_MissingSlotWhenOnlyInTap(t *testing.T) {
	tapStr := tapDocText("", mkTapSection("p.input", "",
		mkTapSlot("input.a.page1.f1", "")+mkTapSlot("other.function", "")))
	g4 := strings.Join(mkG4Section("p.input", mkG4Slot("f1", "input.a.page1.f1")...), "\n")
	doc := linkOf(t, tapStr, g4, "")
	if doc.Stats.PointsMissing != 1 {
		t.Fatalf("missing = %d，想要 1：%+v", doc.Stats.PointsMissing, doc.Points)
	}
	var m *LinkPoint
	for i := range doc.Points {
		if doc.Points[i].Missing {
			m = &doc.Points[i]
		}
	}
	if m == nil || m.Name != "other.function" || m.Line0 != 0 {
		t.Errorf("missing 点 = %+v", m)
	}
}

// 实测：`chk_data_exist.define` 的标记行 name 被截断成 `chk_`（两侧都截断），
// 只能按顺序配；名字不同要留证（g4Name）。
func TestBuildLink_TruncatedNamePairsByOrder(t *testing.T) {
	tapStr := tapDocText("", mkTapDecl("chk_data_exist.define", "chk_data_exist.define", "s", "2", "DEFINE x\n")+
		mkTapSection("p.chk", "", mkTapSlotTruncated("chk段define", "chk_data_exist.define")))
	g4 := strings.Join(mkG4Section("p.chk",
		"   #add-point:chk段define name=\"chk_\"", "   DEFINE x", "   #end add-point"), "\n")

	doc := linkOf(t, tapStr, g4, "")
	if doc.Stats.Points != 1 {
		t.Fatalf("点位 = %+v", doc.Points)
	}
	p := doc.Points[0]
	if p.Name != "chk_data_exist.define" || p.G4Name != "chk_" {
		t.Errorf("应保留两侧名字差异：name=%q g4Name=%q", p.Name, p.G4Name)
	}
	if p.DeclLines != 1 {
		t.Errorf("声明行数 = %d", p.DeclLines)
	}
	if doc.Stats.PointsMissing != 0 || doc.Stats.PointsExtra != 0 {
		t.Errorf("不该有 missing/extra：%+v", doc.Stats)
	}
}

// ---------- 声明侧：无标记的函数、真正的孤儿 ----------

func TestBuildLink_UnmarkedFunctionsAndOrphans(t *testing.T) {
	tapStr := tapDocText("",
		mkTapDecl("function.p_feature", "function.p_feature", "c", "2", "PRIVATE FUNCTION p_feature()\n  LET x = 1\nEND FUNCTION\n")+
			mkTapDecl("input.a.page9.gone", "input.a.page9.gone", "c", "3", "IF x THEN y END IF\n")+
			mkTapSection("p.other_function", "", mkTapSlot("other.function", "")))
	g4 := []string{
		`{<section id="p.other_function" type="s" >}`,
		"#部門檢查",
		"PRIVATE FUNCTION p_feature()",
		"  LET x = 1",
		"END FUNCTION",
		`{</section>}`,
	}
	doc := linkOf(t, tapStr, strings.Join(g4, "\n"), "")
	if doc.Stats.DeclUnmarked != 1 {
		t.Fatalf("unmarked = %d：%+v", doc.Stats.DeclUnmarked, doc.Points)
	}
	var un *LinkPoint
	for i := range doc.Points {
		if doc.Points[i].Unmarked {
			un = &doc.Points[i]
		}
	}
	if un == nil || un.Line0 != 3 || un.Line1 != 5 {
		t.Errorf("函数定位 = %+v（应在第 3-5 行）", un)
	}
	if doc.Stats.DeclOrphans != 1 || len(doc.Orphans) != 1 {
		t.Fatalf("孤儿 = %d %+v", doc.Stats.DeclOrphans, doc.Orphans)
	}
	if doc.Orphans[0].Name != "input.a.page9.gone" || doc.Orphans[0].Reason != "no-slot" {
		t.Errorf("孤儿 = %+v", doc.Orphans[0])
	}
	if c := checkByID(doc, "decls-orphan"); c == nil || c.Level != "warn" {
		t.Errorf("decls-orphan = %+v", c)
	}
}

// ---------- 缺一侧材料 ----------

func TestBuildLink_NoTapDegradesToG4Only(t *testing.T) {
	var inner []string
	inner = append(inner, mkG4Slot("run", "main.run", "  LET x = 1")...)
	g4 := strings.Join(mkG4Section("p.main", inner...), "\n")
	g4 += "\n" + strings.Join(mkG4Section("p.input", mkG4Slot("f1", "input.a.page1.f1")...), "\n")
	tabStr := tabDocText(`name="p" module="cxm" type="t01" page="1" cusver="0000"`,
		`<var id="body" page="1" page_id="1" record="s_detail1" value="f1,f2"/>`)
	doc := linkOf(t, "", g4, tabStr)
	if doc.Stats.SlotsTap != 0 || doc.Stats.Points != 2 {
		t.Fatalf("统计 = %+v", doc.Stats)
	}
	for _, p := range doc.Points {
		if !p.G4Only {
			t.Errorf("没有 tap 时点位应标 g4Only：%+v", p)
		}
	}
	if doc.Stats.FieldsLinked != 1 {
		t.Errorf("字段仍应从 .tab 与点名推出来：%+v", doc.Fields)
	}
}

func TestBuildLink_NoTabStillWorks(t *testing.T) {
	tapStr := tapDocText("", mkTapSection("p.input", "", mkTapSlot("input.a.page1.f1", "")))
	g4 := strings.Join(mkG4Section("p.input", mkG4Slot("f1", "input.a.page1.f1")...), "\n")
	doc := linkOf(t, tapStr, g4, "")
	if c := checkByID(doc, "blueprint"); c == nil || c.Level != "warn" {
		t.Errorf("blueprint 检查 = %+v", c)
	}
	if len(doc.Fields) != 1 || doc.Fields[0].PageID != "1" || doc.Fields[0].Column != "f1" {
		t.Errorf("字段 = %+v", doc.Fields)
	}
}

func TestBuildLink_SectionMismatchIsError(t *testing.T) {
	tapStr := tapDocText("", mkTapSection("p.a", "")+mkTapSection("p.b", ""))
	g4 := strings.Join(mkG4Section("p.a"), "\n")
	doc := linkOf(t, tapStr, g4, "")
	c := checkByID(doc, "sections-pair")
	if c == nil || c.Level != "error" {
		t.Fatalf("sections-pair = %+v", c)
	}
}

func TestBuildLink_FreestyleDoesNotPanic(t *testing.T) {
	tapStr := tapDocText("", mkTapSection("p.main", "", mkTapSlot("main.run", "")))
	doc := linkOf(t, tapStr, "PROGRAM p\nMAIN\nEND MAIN\n", "")
	if doc.Stats.SectionsG4 != 0 || doc.Stats.Points != 1 {
		t.Errorf("统计 = %+v", doc.Stats)
	}
	if c := checkByID(doc, "sections-pair"); c == nil || c.Level != "warn" {
		t.Errorf("sections-pair = %+v", c)
	}
}

// ---------- 切片脱版 ----------

func TestFinishSlices_DetectsDrift(t *testing.T) {
	mdl := t.TempDir()
	if err := os.MkdirAll(filepath.Join(mdl, "slice"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mdl, "slice", "code_a05.template"),
		[]byte("#應用 a05 樣板自動產生(Version:3)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sl := LinkSlices{Total: 2, Distinct: 1, ByCode: []SliceUse{{Code: "a05", Ver: 2, Count: 2}}}
	got, checks := finishSlices(sl, mdl, nil)
	if len(got.Drift) != 1 || got.Drift[0].Code != "a05" || got.Drift[0].Current != 3 {
		t.Fatalf("脱版 = %+v", got.Drift)
	}
	if len(checks) != 1 || checks[0].Level != "warn" {
		t.Errorf("检查 = %+v", checks)
	}
}

// ---------- 端到端：写文件 ----------

func TestMap_WritesLinkJSON(t *testing.T) {
	cfg := writeConfig(t, `{"name":"e","host":"h","port":22,"user":"u"}`)
	progDir := filepath.Join(filepath.Dir(cfg), "devdata", "e", "prog", "axmt500_wf")

	tapStr := tapDocText("", mkTapDecl("input.a.page1.f1", "", "c", "3", "LET f1 = 1\n")+
		mkTapSection("p.input", "", mkTapSlot("input.a.page1.f1", "過欄位")))
	g4Str := strings.Join(mkG4Section("p.input", mkG4Slot("過欄位", "input.a.page1.f1", "  LET f1 = 1")...), "\n")
	tabStr := tabDocText(`name="axmt500_wf" module="cxm" type="t01" page="1" cusver="0003"`,
		`<var id="body" page="1" page_id="1" record="s_detail1" value="f1,f2"/>`)

	writeProgFiles(t, progDir, map[string]string{
		"dzx/axmt500_wf.tap": tapStr,
		"dzx/axmt500_wf.tab": tabStr,
		"src/axmt500_wf.4gl": g4Str,
	})

	res, err := Map(LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"})
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if res.Path != filepath.Join(progDir, "link.json") {
		t.Errorf("索引路径 = %s", res.Path)
	}
	b, err := os.ReadFile(res.Path)
	if err != nil {
		t.Fatalf("读 link.json: %v", err)
	}
	var back LinkDoc
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("link.json 不是合法 JSON: %v", err)
	}
	if back.Schema != LinkSchema || back.Prog != "axmt500_wf" {
		t.Errorf("schema/prog = %q/%q", back.Schema, back.Prog)
	}
	if back.Sources.Tap.SHA256 == "" || back.Sources.G4.SHA256 == "" {
		t.Errorf("sources 缺指纹: %+v", back.Sources)
	}
	if len(back.Points) != 1 || !back.Points[0].Custom {
		t.Errorf("点位 = %+v", back.Points)
	}
	if len(back.Fields) != 1 || back.Fields[0].Column != "f1" {
		t.Errorf("字段 = %+v", back.Fields)
	}
}

// writeProgFiles 写一套程序材料 + manifest（sha256 由内容算）。
func writeProgFiles(t *testing.T, progDir string, files map[string]string) {
	t.Helper()
	var entries []FileEntry
	for rel, body := range files {
		p := filepath.Join(progDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("建目录: %v", err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("写文件: %v", err)
		}
		role := "4gl"
		switch {
		case strings.HasSuffix(rel, ".tap"):
			role = "tap"
		case strings.HasSuffix(rel, ".tgl"):
			role = "tgl"
		case strings.HasSuffix(rel, ".tab"):
			role = "tab"
		case strings.HasSuffix(rel, ".tsd"):
			role = "tsd"
		case strings.HasSuffix(rel, ".4fd"):
			role = "4fd"
		}
		entries = append(entries, FileEntry{
			Role: role, Src: "erp/cxm/" + rel, Rel: rel,
			SHA256: shaOfString(body), Size: int64(len(body)),
		})
	}
	mf := Manifest{Schema: "tt.devdata/v1", Kind: "fetch", Env: "e",
		Prog: filepath.Base(progDir), Module: "cxm", Files: entries}
	b, _ := json.MarshalIndent(mf, "", "  ")
	if err := os.WriteFile(filepath.Join(progDir, "manifest.json"), b, 0o644); err != nil {
		t.Fatalf("写 manifest: %v", err)
	}
}
