package devdata

// show / at / fields 三条查询的测试。
//
// 这三条是"索引有没有用"的真正兑现处：AI 问的是"这一行/这个字段是什么"，
// 不是"给我一份 800 KB 的 JSON"。所以断言落在**回答是否准确、是否有安全性结论**上。

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMatchPoints_ThreeLevels(t *testing.T) {
	doc := &LinkDoc{Points: []LinkPoint{
		{Name: "input.a.page1.f1", Line0: 10},
		{Name: "construct.c.page1.f1", Line0: 5},
		{Name: "input.a.page73.xmdduc001", Line0: 40},
		{Name: "other.function", Line0: 0},
	}}

	// 精确
	got, how := matchPoints(doc, "input.a.page1.f1")
	if how != "exact" || len(got) != 1 || got[0].Name != "input.a.page1.f1" {
		t.Fatalf("精确匹配 = %v %v", got, how)
	}
	// 后缀（只给字段名）→ 按行号排序（0 行的排最后）
	got, how = matchPoints(doc, "f1")
	if how != "suffix" || len(got) != 2 {
		t.Fatalf("后缀匹配 = %v %v", got, how)
	}
	if got[0].Line0 != 5 || got[1].Line0 != 10 {
		t.Errorf("后缀结果应按行号排序: %+v", got)
	}
	// 子串
	got, how = matchPoints(doc, "xmdduc")
	if how != "substring" || len(got) != 1 {
		t.Fatalf("子串匹配 = %v %v", got, how)
	}
	// 空 / 无匹配
	if got, _ := matchPoints(doc, ""); got != nil {
		t.Errorf("空查询应返回 nil: %v", got)
	}
	if got, _ := matchPoints(doc, "没这个东西"); len(got) != 0 {
		t.Errorf("无匹配应为空: %v", got)
	}
}

// queryFixture 造一套最小材料：一个客制加点 + 一段模板正文 + 一个字段。
func queryFixture(t *testing.T) (cfg string, f1Line, plainLine int) {
	t.Helper()
	cfg = writeConfig(t, `{"name":"e","host":"h","port":22,"user":"u"}`)
	progDir := progDirOf(cfg, "axmt500_wf")

	tapStr := tapDocText("",
		mkTapDecl("input.a.page1.f1", "input.a.page1.f1", "c", "3", "LET f1 = 1\nLET g = 2\n")+
			mkTapSection("p.input", "56", mkTapSlot("input.a.page1.f1", "過欄位")))
	g4 := []string{
		`{<section id="p.input" type="s" >}`,
		"#應用 a05 樣板自動產生(Version:3)",
		"   #add-point:過欄位 name=\"input.a.page1.f1\"",
		"   LET f1 = 1",
		"   LET g = 2",
		"   #END add-point",
		"   DISPLAY f1",
		`{</section>}`,
	}
	tabStr := tabDocText(`name="axmt500_wf" module="cxm" type="t01" page="1" cusver="0003"`,
		`<var id="body" page="1" page_id="1" record="s_detail1" value="f1,f2"/>`)

	writeProgFiles(t, progDir, map[string]string{
		"dzx/axmt500_wf.tap": tapStr,
		"src/axmt500_wf.4gl": strings.Join(g4, "\n"),
		"dzx/axmt500_wf.tab": tabStr,
	})
	// 4gl 第 4 行是加点里的第一句；第 7 行是 section 正文（不在加点内）
	return cfg, 4, 7
}

func progDirOf(cfg, prog string) string {
	return filepath.Join(filepath.Dir(cfg), "devdata", "e", "prog", prog)
}

func TestShow_ByFieldReturnsContent(t *testing.T) {
	cfg, f1Line, _ := queryFixture(t)
	res, err := Show(ShowRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Query:       "f1",
	})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if res.Match != "suffix" || len(res.Points) != 1 {
		t.Fatalf("匹配 = %s / %d 点", res.Match, len(res.Points))
	}
	p := res.Points[0]
	if p.Name != "input.a.page1.f1" || !p.Custom {
		t.Errorf("点位 = %+v", p.LinkPoint)
	}
	if len(p.Content) == 0 {
		t.Fatal("没有内容")
	}
	// 内容应是加点那几行（含标记行）
	if !strings.Contains(p.Content[0], "#add-point:") {
		t.Errorf("内容首行 = %q", p.Content[0])
	}
	if len(p.Content) != p.Line1-p.Line0+1 {
		t.Errorf("内容行数 = %d，区间 = %d-%d", len(p.Content), p.Line0, p.Line1)
	}
	if p.Line0 != f1Line-1 { // 标记行在 body 第一行的上一行
		t.Errorf("Line0 = %d，想要 %d", p.Line0, f1Line-1)
	}
	if len(p.SliceCodes) != 1 || p.SliceCodes[0] != "a05" {
		t.Errorf("段内切片 = %v", p.SliceCodes)
	}
	// 截断
	res, err = Show(ShowRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Query:       "f1", MaxLines: 2,
	})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if !res.Points[0].Truncated || len(res.Points[0].Content) != 2 {
		t.Errorf("截断 = %v / %d 行", res.Points[0].Truncated, len(res.Points[0].Content))
	}
}

func TestShow_NoMatchIsNotFound(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	_, err := Show(ShowRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Query:       "nope_xxx",
	})
	if !errors.Is(err, ErrPointNotFound) {
		t.Fatalf("err = %v，想要 ErrPointNotFound", err)
	}
}

func TestAt_CustomPointVsTemplateBody(t *testing.T) {
	cfg, f1Line, plainLine := queryFixture(t)

	in, err := At(AtRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Line:        f1Line,
	})
	if err != nil {
		t.Fatalf("At: %v", err)
	}
	if !in.Inside || in.Point == nil || in.Point.Name != "input.a.page1.f1" {
		t.Fatalf("客制行 = %+v", in)
	}
	if !strings.Contains(in.Verdict, "客制") || !strings.Contains(in.Verdict, "保留") {
		t.Errorf("结论 = %q（应说明重产会保留）", in.Verdict)
	}
	if in.Section == nil || in.Section.ID != "p.input" {
		t.Errorf("section = %+v", in.Section)
	}
	if in.Slice == nil || in.Slice.Code != "a05" {
		t.Errorf("切片 = %+v", in.Slice)
	}
	if len(in.Content) == 0 {
		t.Error("没有上下文")
	}

	// section 正文（不在加点内）→ 必须明确提示"手改会被重产覆盖"
	out, err := At(AtRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Line:        plainLine,
	})
	if err != nil {
		t.Fatalf("At: %v", err)
	}
	if out.Inside {
		t.Errorf("正文行不该判 Inside: %+v", out)
	}
	if !strings.Contains(out.Verdict, "重产覆盖") {
		t.Errorf("结论 = %q（应警告会覆盖）", out.Verdict)
	}
}

func TestAt_LineOutOfRange(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	_, err := At(AtRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Line:        99999,
	})
	if !errors.Is(err, ErrInvalidOpt) {
		t.Fatalf("err = %v，想要 ErrInvalidOpt", err)
	}
}

func TestFields_Filter(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	fs, doc, err := Fields(FieldsRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
	})
	if err != nil {
		t.Fatalf("Fields: %v", err)
	}
	if len(fs) != 1 || fs[0].Column != "f1" || fs[0].PageID != "1" {
		t.Fatalf("字段 = %+v（f2 没有点位，不该出现）", fs)
	}
	if fs[0].Page != "1" || fs[0].Record != "s_detail1" {
		t.Errorf("页信息 = %+v", fs[0])
	}
	if doc.Stats.ColumnsTotal != 2 {
		t.Errorf("columnsTotal = %d，想要 2（f1+f2）", doc.Stats.ColumnsTotal)
	}

	if got, _, _ := Fields(FieldsRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"}, Page: "2",
	}); len(got) != 0 {
		t.Errorf("page=2 应空: %+v", got)
	}
	if got, _, _ := Fields(FieldsRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"}, Pattern: "f1",
	}); len(got) != 1 {
		t.Errorf("模式 f1 应命中 1: %+v", got)
	}
	if got, _, _ := Fields(FieldsRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"}, Pattern: "zzz",
	}); len(got) != 0 {
		t.Errorf("模式 zzz 应空: %+v", got)
	}
}

// ---------- grep ----------

func TestGrep_AttributesHits(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	res, err := Grep(GrepRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Pattern:     "let f1",
	})
	if err != nil {
		t.Fatalf("Grep: %v", err)
	}
	if res.Total != 1 || len(res.Hits) != 1 {
		t.Fatalf("命中 = %d/%d", res.Total, len(res.Hits))
	}
	h := res.Hits[0]
	if !h.Inside || h.Point != "input.a.page1.f1" || h.Kind != "客制" {
		t.Errorf("命中归属 = %+v", h)
	}
	if h.Line != 4 {
		t.Errorf("行号 = %d，想要 4", h.Line)
	}
	res2, _ := Grep(GrepRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Pattern:     "LET F1",
	})
	if res2.Total != 1 {
		t.Errorf("忽略大小写时命中 = %d", res2.Total)
	}
	res3, _ := Grep(GrepRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Pattern:     "LET F1", Case: true,
	})
	if res3.Total != 0 {
		t.Errorf("区分大小写时命中 = %d", res3.Total)
	}
}

func TestGrep_LimitKeepsTotal(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	res, err := Grep(GrepRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Pattern:     "LET", Limit: 1,
	})
	if err != nil {
		t.Fatalf("Grep: %v", err)
	}
	if res.Total < 2 || res.Shown != 1 {
		t.Errorf("total=%d shown=%d（总数不能被 limit 截掉）", res.Total, res.Shown)
	}
}

// ---------- fields：序号 vs page_id ----------

func TestFields_AcceptsPageSeq(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	fs, doc, err := Fields(FieldsRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Page:        "1",
	})
	if err != nil {
		t.Fatalf("Fields: %v", err)
	}
	if len(fs) != 1 {
		t.Fatalf("字段 = %+v", fs)
	}
	if len(doc.Pages) != 1 || doc.Pages[0].PageID != "1" {
		t.Errorf("pages = %+v", doc.Pages)
	}
	if got, d2, _ := Fields(FieldsRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"}, Page: "99",
	}); len(got) != 0 {
		t.Errorf("page 99 应为空: %+v", got)
	} else if len(d2.Notes) == 0 {
		t.Error("空结果应带提示 note")
	}
}

// ---------- show：带点的查询必须精确 ----------

func TestShow_DottedQueryRequiresExact(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	_, err := Show(ShowRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Query:       "input.a.page1.f",
	})
	if !errors.Is(err, ErrPointNotFound) {
		t.Fatalf("err = %v，想要 ErrPointNotFound", err)
	}
	res, err := Show(ShowRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Query:       "f1",
	})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if res.Match != "suffix" || res.Note == "" {
		t.Errorf("match=%q note=%q", res.Match, res.Note)
	}
}

// ---------- 索引复用 ----------

func TestLoadIndex_InvalidatesOnSourceChange(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	first, err := LoadIndex(LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"})
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	p := filepath.Join(progDirOf(cfg, "axmt500_wf"), "src", "axmt500_wf.4gl")
	b, _ := os.ReadFile(p)
	if err := os.WriteFile(p, append(b, []byte("\nDISPLAY 1\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := LoadIndex(LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"})
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	if second.Sources.G4.SHA256 == first.Sources.G4.SHA256 {
		t.Error("源文件变了，索引里的 sha 也应变（否则就是复用旧索引）")
	}
	if second.Stats.G4Lines <= first.Stats.G4Lines {
		t.Errorf("行数没变：%d → %d", first.Stats.G4Lines, second.Stats.G4Lines)
	}
}
