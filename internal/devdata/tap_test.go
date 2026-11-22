package devdata

// 「代码在、跑起来没有」这条路：.tap 里的声明正文、孤儿点、中文标签。
//
// 这一组测试钉住三件试用中真的踩到的事：
//  1. at 打出来的行号必须以内容首行为准（少一个 ContentStart 就整段偏移）；
//  2. 只有 .tap 声明、没有槽位的点，Show 要能查到并说清"重产不保留"；
//  3. search 必须能搜到 .tap（产物里根本没有那段文本）。

import (
	"strings"
	"testing"
)

// tapOnlyFixture 造一个"有声明、没槽位"的点（孤儿），外加一个正常点。
func tapOnlyFixture(t *testing.T) string {
	t.Helper()
	cfg := writeConfig(t, `{"name":"e","host":"h","port":22,"user":"u"}`)
	progDir := progDirOf(cfg, "axmt500_wf")

	tapStr := tapDocText("",
		mkTapDecl("input.a.page1.gone", "", "c", "2", "LET x = 1\nMESSAGE '孤兒內容'\n")+
			mkTapDecl("input.a.page1.f1", "", "c", "2", "LET f1 = 1\n")+
			mkTapSection("p.input", "1", mkTapSlot("input.a.page1.f1", "過欄位")))
	g4 := []string{
		`{<section id="p.input" type="s" >}`,
		"   #add-point:過欄位 name=\"input.a.page1.f1\"",
		"   LET f1 = 1",
		"   #END add-point",
		`{</section>}`,
	}
	tabStr := tabDocText(`name="axmt500_wf" module="cxm" type="t01" page="1" cusver="0000"`,
		`<var id="body" page="1" page_id="1" record="s_detail1" value="f1"/>`)
	writeProgFiles(t, progDir, map[string]string{
		"dzx/axmt500_wf.tap": tapStr,
		"src/axmt500_wf.4gl": strings.Join(g4, "\n"),
		"dzx/axmt500_wf.tab": tabStr,
	})
	return cfg
}

func TestAt_ContentStartIsFirstContextLine(t *testing.T) {
	cfg, f1Line, _ := queryFixture(t)
	res, err := At(AtRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Line:        f1Line, Context: 2,
	})
	if err != nil {
		t.Fatalf("At: %v", err)
	}
	if res.ContentStart != f1Line-2 {
		t.Fatalf("ContentStart = %d，想要 %d（行号偏移就是从这里漏出去的）", res.ContentStart, f1Line-2)
	}
	if len(res.Content) < 3 {
		t.Fatalf("上下文太短: %v", res.Content)
	}
}

func TestShow_OrphanComesFromTapAndSaysNotKept(t *testing.T) {
	cfg := tapOnlyFixture(t)
	res, err := Show(ShowRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Query:       "input.a.page1.gone",
	})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if res.Match != "orphan" || len(res.Points) != 1 {
		t.Fatalf("match=%s points=%d", res.Match, len(res.Points))
	}
	p := res.Points[0]
	if !p.Orphan {
		t.Error("没标成孤儿")
	}
	if len(p.Content) != 2 || !strings.Contains(p.Content[0], "LET x = 1") {
		t.Errorf("正文应来自 .tap: %#v", p.Content)
	}
	if p.Line0 == 0 || p.Line1 < p.Line0 {
		t.Errorf("缺 .tap 行号: %d-%d", p.Line0, p.Line1)
	}
	if !strings.Contains(res.Note, "重产不会保留") {
		t.Errorf("提示没说清后果: %q", res.Note)
	}
}

func TestShow_OrphanBySuffix(t *testing.T) {
	cfg := tapOnlyFixture(t)
	res, err := Show(ShowRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Query:       "gone",
	})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if res.Match != "orphan" || res.Points[0].Name != "input.a.page1.gone" {
		t.Errorf("后缀命中孤儿失败: %+v", res.Points)
	}
}

func TestGrep_ScopeTapFindsOrphanBody(t *testing.T) {
	cfg := tapOnlyFixture(t)
	lr := LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"}

	// 默认（4gl）搜不到：孤儿正文只在 .tap 里
	res, err := Grep(GrepRequest{LinkRequest: lr, Pattern: "孤兒內容"})
	if err != nil {
		t.Fatalf("Grep: %v", err)
	}
	if res.Total != 0 || res.TapTotal != 0 {
		t.Errorf("默认范围不该搜 .tap: 4gl=%d tap=%d", res.Total, res.TapTotal)
	}

	res, err = Grep(GrepRequest{LinkRequest: lr, Pattern: "孤兒內容", Scope: "all"})
	if err != nil {
		t.Fatalf("Grep(all): %v", err)
	}
	if res.TapTotal != 1 || len(res.TapHits) != 1 {
		t.Fatalf("tap 命中 = %d / %+v", res.TapTotal, res.TapHits)
	}
	h := res.TapHits[0]
	if h.Point != "input.a.page1.gone" || h.HasSlot {
		t.Errorf("归属/槽位判定错: %+v", h)
	}
	if h.Src != "c" {
		t.Errorf("src = %q", h.Src)
	}
}

func TestResolveColumns_ByLabelAndFindByLabel(t *testing.T) {
	cfg := writeConfig(t, `{"name":"e","host":"h","port":22,"user":"u"}`)
	progDir := progDirOf(cfg, "axmt500_wf")
	tapStr := tapDocText("",
		mkTapDecl("input.a.page73.xmdduc003", "", "c", "2", "LET x = 1\n")+
			mkTapSection("p.input", "1", mkTapSlot("input.a.page73.xmdduc003", "過欄位")))
	g4 := []string{
		`{<section id="p.input" type="s" >}`,
		"   #add-point:過欄位 name=\"input.a.page73.xmdduc003\"",
		"   LET x = 1",
		"   #END add-point",
		`{</section>}`,
	}
	tabStr := tabDocText(`name="axmt500_wf" module="cxm" type="t01" page="1" cusver="0000"`,
		`<var id="body" page="12" page_id="73" record="s_detail73" value="xmdduc003"/>`)
	tsd := `<spec><strings><sfield name="lbl_xmdduc003" text="刻号" /></strings></spec>`
	writeProgFiles(t, progDir, map[string]string{
		"dzx/axmt500_wf.tap": tapStr,
		"src/axmt500_wf.4gl": strings.Join(g4, "\n"),
		"dzx/axmt500_wf.tab": tabStr,
		"dzx/axmt500_wf.tsd": tsd,
	})

	lr := LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"}
	res, err := Find(FindRequest{LinkRequest: lr, Field: "刻号"})
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if res.Total != 1 || res.Points[0].Name != "input.a.page73.xmdduc003" {
		t.Fatalf("按中文名没定位到: total=%d %+v", res.Total, res.Points)
	}
	if res.FieldBy != "label" {
		t.Errorf("FieldBy = %q，想要 label", res.FieldBy)
	}
	if res.Labels["xmdduc003"] != "刻号" {
		t.Errorf("Labels = %+v", res.Labels)
	}
	// 页签收窄：page_id 与序号两种写法都要能用
	res, err = Find(FindRequest{LinkRequest: lr, Field: "刻号", Page: "73"})
	if err != nil || res.Total != 1 {
		t.Fatalf("按 page_id 收窄失败: %v / %d", err, res.Total)
	}
	res, err = Find(FindRequest{LinkRequest: lr, Field: "刻号", Page: "12"})
	if err != nil || res.Total != 1 {
		t.Fatalf("按页签序号收窄失败: %v / %d", err, res.Total)
	}
	if !strings.Contains(res.Note, "page_id 73") {
		t.Errorf("序号换算没写进提示: %q", res.Note)
	}
}
