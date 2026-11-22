package devdata

// brief / find（意图定位）的磁盘材料矩阵。
//
// 这两条命令是"第一眼"与"下一步"的兑现处：brief 给一页纸画像，find 把业务语言
// （"检核"、"刻号"）翻成具体加点。测试断言落在**数字是否来自同一份索引**、
// **孤儿有没有被单独列出来**（混进点位会被当成可改的落点）、以及**截断后总数还在**。

import (
	"strings"
	"testing"
)

// newBriefMaterials 造一份"各类都有"的材料：客制点行数不同、客制/标准孤儿、
// 脱版切片、无标记客制函数、两页页签。
func newBriefMaterials(t *testing.T) *matSet {
	t.Helper()
	return newMaterials(t, matOpts{
		Prog: "axmt500_wf",
		Pages: []matPage{
			{Seq: "1", ID: "1", Record: "s_browse", Columns: []string{"f1"}},
			{Seq: "12", ID: "73", Record: "s_detail73", Columns: []string{"xmdduc001"}},
		},
		Sections: []matSection{
			{ID: "p.input", Ver: "56", Src: "s", Slots: []matSlot{
				{Name: "input.a.page73.xmdduc001", Label: "刻号检核", Src: "c", Ver: "3",
					DeclBody: "IF x = '' THEN\n  MESSAGE 'x'\nEND IF\n",
					Body:     []string{"   IF x = '' THEN", "      MESSAGE 'x'", "   END IF"}},
				{Name: "input.b.page1.f1", Label: "客户预带", Src: "c", Ver: "1",
					DeclBody: "LET f1 = 1\n", Body: []string{"   LET f1 = 1"}},
			}},
			{ID: "p.construct", Ver: "12", Src: "s", Slots: []matSlot{
				{Name: "construct.c.page73.xmdduc001", Label: "ON ACTION", Src: "s", Ver: "9",
					DeclBody: "CALL open_win()\n", Body: []string{"   CALL open_win()"}},
			}},
		},
		Orphans: []matOrphan{
			{Name: "input.a.page73.xmdduc001_desc", Src: "c", Ver: "2", Body: "LET desc = 1\n"},
			{Name: "input.a.page9.std_residual", Src: "s", Ver: "1", Body: "LET std = 1\n"},
		},
		Funcs:  []matFunc{{Name: "p_feature", Body: []string{"  LET y = 2"}}},
		Slices: []matSlice{{Code: "a17", Used: 2, Current: 3, InBody: true}},
		TSDLabels: map[string]string{
			"f1": "客户", "xmdduc001": "刻号",
		},
		TSDPages: map[string]string{"page_1": "基本数据", "page_73": "示例页签"},
	})
}

// TestFixtureBrief_Matrix 钉住 brief 的五个问题：判型、页面、客制落点、风险、下一步。
// 所有数字必须与同一份索引一致（brief 不另算一份）。
func TestFixtureBrief_Matrix(t *testing.T) {
	set := newBriefMaterials(t)
	doc := set.mapIndex()
	lr := set.LinkRequest()

	b, err := Brief(BriefRequest{LinkRequest: lr})
	if err != nil {
		t.Fatalf("Brief: %v", err)
	}
	if b.Prog != doc.Prog || b.Env != doc.Env || b.Link != doc.Path {
		t.Errorf("身份/索引 = %+v", b)
	}
	if b.Mode.Type != "t01" || b.Mode.Identity != "c" || b.Stats.Points != doc.Stats.Points {
		t.Errorf("判型/统计没跟索引一致: %+v / %+v", b.Mode, b.Stats)
	}
	if len(b.Pages) != 2 || b.Pages[1].PageID != "73" || b.Pages[1].Title != "示例页签" {
		t.Errorf("页面 = %+v", b.Pages)
	}
	if len(b.Sections) != 2 {
		t.Errorf("section = %+v", b.Sections)
	}
	// 客制点按声明行数降序：IF 块（3 行）排最前；客制函数（1 行）与预带（1 行）随后。
	// 注意：无标记的客制函数也会被计入客制点（它确实是重产会保留的客户代码）。
	if b.CustomTotal != 3 || len(b.CustomPoints) != 3 {
		t.Fatalf("客制点 = %d / %+v", b.CustomTotal, b.CustomPoints)
	}
	if b.CustomPoints[0].Name != "input.a.page73.xmdduc001" || b.CustomPoints[0].DeclLines != 3 {
		t.Errorf("客制点排序 = %+v（应按行数降序）", b.CustomPoints)
	}
	if b.CustomPoints[1].Name != "function.p_feature" || !b.CustomPoints[1].Unmarked {
		t.Errorf("第二块 = %+v（1 行，按名字排；无标记客制函数也要算）", b.CustomPoints[1])
	}
	// 孤儿：客制 1 个（真的丢了客户代码），标准残留不算 OrphansCustom。
	if len(b.Orphans) != 2 || b.OrphansCustom != 1 {
		t.Errorf("孤儿 = %+v / custom=%d", b.Orphans, b.OrphansCustom)
	}
	// 脱版：产品用了旧模板 → Notes 必须给风险提示。
	if len(b.Slices.Drift) != 1 {
		t.Fatalf("切片脱版 = %+v", b.Slices.Drift)
	}
	foundDrift := false
	for _, n := range b.Notes {
		if strings.Contains(n, "切片脱版") {
			foundDrift = true
		}
	}
	if !foundDrift {
		t.Errorf("brief 没提示切片脱版: %v", b.Notes)
	}
	// 意图清单与下一步。
	if len(b.Intents) != len(Intents) {
		t.Errorf("意图清单 = %d，想要 %d", len(b.Intents), len(Intents))
	}
	hasNext := false
	for _, n := range b.Notes {
		if strings.Contains(n, "下一步") && strings.Contains(n, "tt dev") {
			hasNext = true
		}
	}
	if !hasNext {
		t.Errorf("brief 没给下一步命令: %v", b.Notes)
	}
	// 检查表原样带出（brief 不重新判定）。
	if len(b.Checks) != len(doc.Checks) {
		t.Errorf("检查表 = %d，索引里 %d", len(b.Checks), len(doc.Checks))
	}

	// TopPoints 截断：只列前 N 个，但总数保留、给调大的提示。
	trunc, err := Brief(BriefRequest{LinkRequest: lr, TopPoints: 1})
	if err != nil {
		t.Fatalf("Brief(top=1): %v", err)
	}
	if len(trunc.CustomPoints) != 1 || trunc.CustomTotal != 3 {
		t.Errorf("截断 = %d/%d", len(trunc.CustomPoints), trunc.CustomTotal)
	}
	foundTop := false
	for _, n := range trunc.Notes {
		if strings.Contains(n, "只列最大 1 个") {
			foundTop = true
		}
	}
	if !foundTop {
		t.Errorf("截断没提示: %v", trunc.Notes)
	}

	// 没有 .tab：brief 要明说"没有页面信息"，而不是静默给空页面。
	notab := newMaterials(t, matOpts{Prog: "axmt501", NoTab: true, NoTsd: true, No4fd: true})
	notab.mapIndex()
	nb, err := Brief(BriefRequest{LinkRequest: notab.LinkRequest()})
	if err != nil {
		t.Fatalf("Brief(noTab): %v", err)
	}
	foundNoPage := false
	for _, n := range nb.Notes {
		if strings.Contains(n, "没有页面信息") {
			foundNoPage = true
		}
	}
	if !foundNoPage {
		t.Errorf("无 .tab 时 brief 没提示: %v", nb.Notes)
	}
}

// TestFixtureFind_IntentMatrix 钉住"意图 → 落点"的筛选：
// 段名/槽位命中、字段（列名或中文名）收窄、页签收窄、limit 与 omittedPages、
// 孤儿与命中点分开列、未知意图的引导。
func TestFixtureFind_IntentMatrix(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt500_wf",
		Pages: []matPage{
			{Seq: "1", ID: "1", Record: "s_browse", Columns: []string{"xmdduc001"}},
			{Seq: "12", ID: "73", Record: "s_detail73", Columns: []string{"xmdduc001", "xmdduc002", "xmdduc003"}},
		},
		Sections: []matSection{
			{ID: "p.input", Ver: "1", Src: "s", Slots: []matSlot{
				{Name: "input.a.page73.xmdduc001", Label: "刻号检核", Src: "c", Ver: "3", DeclBody: "LET a = 1\n", Body: []string{"  LET a = 1"}},
				{Name: "input.b.page73.xmdduc002", Label: "料号检核", Src: "c", Ver: "2", DeclBody: "LET b = 1\n", Body: []string{"  LET b = 1"}},
			}},
			{ID: "p.construct", Ver: "2", Src: "s", Slots: []matSlot{
				{Name: "construct.c.page73.xmdduc001", Label: "开窗", Src: "s", Ver: "9", DeclBody: "CALL w()\n", Body: []string{"  CALL w()"}},
			}},
			{ID: "p.insert_b", Ver: "3", Src: "s", Slots: []matSlot{
				{Name: "insert_b.before_insert.xmdduc001", Label: "保存前", Src: "c", Ver: "1", DeclBody: "LET s = 1\n", Body: []string{"  LET s = 1"}},
			}},
			{ID: "p.show", Ver: "4", Src: "s", Slots: []matSlot{
				{Name: "show.display.page73.xmdduc003", Label: "显示", Src: "s", Ver: "1", DeclBody: "DISPLAY 1\n", Body: []string{"  DISPLAY 1"}},
			}},
		},
		Orphans: []matOrphan{
			{Name: "input.a.page73.xmdduc001_desc", Src: "c", Ver: "2", Body: "LET d = 1\n"},
		},
		Slices:    []matSlice{{Code: "a17", Used: 3, Current: 3, InBody: true}},
		TSDLabels: map[string]string{"xmdduc001": "刻号", "xmdduc002": "料号", "xmdduc003": "项次", "xmdduc001_desc": "品名"},
		TSDPages:  map[string]string{"page_73": "示例页签"},
	})
	set.mapIndex()
	lr := set.LinkRequest()

	// check 意图：命中 input 段两个点 + insert_b 的 before_insert 也算 save 不算 check。
	res, err := Find(FindRequest{LinkRequest: lr, Intent: "check"})
	if err != nil {
		t.Fatalf("Find(check): %v", err)
	}
	if len(res.Intents) != 1 || res.Intents[0].Key != "check" || res.Intents[0].Points != 2 {
		t.Fatalf("check 命中 = %+v", res.Intents)
	}
	if res.Total != 2 {
		t.Errorf("check 点位 = %+v", res.Points)
	}
	// 模板用量只列本程序真用到的（a17）。
	if len(res.Intents[0].Templates) != 1 || res.Intents[0].Templates[0].Code != "a17" {
		t.Errorf("模板用量 = %+v", res.Intents[0].Templates)
	}

	// save 意图：靠槽位段 before_insert 命中（head 是 insert_b，不在 Sections 里也能中）。
	save, err := Find(FindRequest{LinkRequest: lr, Intent: "保存前"})
	if err != nil {
		t.Fatalf("Find(save): %v", err)
	}
	if len(save.Intents) != 1 || save.Intents[0].Key != "save" || save.Total != 1 ||
		save.Points[0].Name != "insert_b.before_insert.xmdduc001" {
		t.Errorf("save 命中 = %+v / %+v", save.Intents, save.Points)
	}

	// show / query 等意图：只要段名对上即可，不要求是客制。
	show, err := Find(FindRequest{LinkRequest: lr, Intent: "display"})
	if err != nil || show.Total != 1 || show.Points[0].Name != "show.display.page73.xmdduc003" {
		t.Errorf("display 命中 = %+v / %v", show, err)
	}

	// 字段按画面中文名收窄 + 意图：刻号 → xmdduc001，只剩 input/construct 两个点里的刻号。
	byLabel, err := Find(FindRequest{LinkRequest: lr, Intent: "check", Field: "刻号"})
	if err != nil {
		t.Fatalf("Find(check,刻号): %v", err)
	}
	if byLabel.FieldBy != "label" || byLabel.Labels["xmdduc001"] != "刻号" {
		t.Errorf("字段解析 = %s / %v", byLabel.FieldBy, byLabel.Labels)
	}
	if byLabel.Total != 1 || byLabel.Points[0].Name != "input.a.page73.xmdduc001" {
		t.Errorf("check+刻号 = %+v", byLabel.Points)
	}

	// 只给字段（无意图）：这个栏位的逻辑散在哪些段都要列出来（含标准段）。
	onlyField, err := Find(FindRequest{LinkRequest: lr, Field: "刻号"})
	if err != nil {
		t.Fatalf("Find(刻号): %v", err)
	}
	if len(onlyField.Intents) != 0 {
		t.Errorf("只给字段不该套意图: %+v", onlyField.Intents)
	}
	names := map[string]bool{}
	for _, p := range onlyField.Points {
		names[p.Name] = true
	}
	if !names["input.a.page73.xmdduc001"] || !names["construct.c.page73.xmdduc001"] {
		t.Errorf("字段点位 = %+v", onlyField.Points)
	}
	// 同页孤儿要被单独带出来（不是混进 Points）。
	if len(onlyField.Orphans) != 1 || onlyField.Orphans[0].Name != "input.a.page73.xmdduc001_desc" {
		t.Errorf("孤儿 = %+v", onlyField.Orphans)
	}
	if !strings.Contains(onlyField.Note, "孤儿声明") {
		t.Errorf("孤儿提示 = %q", onlyField.Note)
	}

	// 页签收窄：page_id 与序号都接受。
	if res, _, _ := Fields(FieldsRequest{LinkRequest: lr, Page: "73"}); len(res) != 3 {
		t.Errorf("page 73 字段 = %d", len(res))
	}
	byPage, err := Find(FindRequest{LinkRequest: lr, Field: "刻号", Page: "12"})
	if err != nil {
		t.Fatalf("Find(page=12): %v", err)
	}
	// 刻号在 page73 上有两个点（input 客制 + construct 标准），页签收窄不该丢标准侧。
	if byPage.Total != 2 || !strings.Contains(byPage.Note, "page_id 73") {
		t.Errorf("序号收窄 = %+v / %q", byPage.Points, byPage.Note)
	}
	// 不存在的页：明确提示而不是静默 0 条。
	if res, _ := Find(FindRequest{LinkRequest: lr, Page: "99"}); !strings.Contains(res.Note, "没有页签") {
		t.Errorf("未知页提示 = %q", res.Note)
	}
	// 未知意图：引导到 brief 的意图清单。
	if res, _ := Find(FindRequest{LinkRequest: lr, Intent: "加个打印按钮"}); !strings.Contains(res.Note, "tt dev brief") {
		t.Errorf("未知意图提示 = %q", res.Note)
	}
	// 完全没有命中：给"能怎么查"的提示，而不是返回空数组就完事。
	if res, _ := Find(FindRequest{LinkRequest: lr, Field: "完全不存在的字段"}); res.Total != 0 ||
		!strings.Contains(res.Note, "画面中文名") {
		t.Errorf("无命中提示 = %+v / %q", res, res.Note)
	}

	// limit 截断：total 保留，且把被截掉的页签列出来（不然"没显示 page73"没人知道）。
	lim, err := Find(FindRequest{LinkRequest: lr, Field: "刻号", Limit: 1})
	if err != nil {
		t.Fatalf("Find(limit): %v", err)
	}
	if lim.Total != 2 || len(lim.Points) != 1 {
		t.Errorf("limit = total:%d shown:%d", lim.Total, len(lim.Points))
	}
	if !strings.Contains(lim.Note, "未显示的涉及") {
		t.Errorf("截断提示 = %q", lim.Note)
	}
}

// TestFixtureFind_OnlyOrphanMatches 钉住现场最常见的一种"没有命中"：
// 字段有声明但没有槽位（孤儿），find 必须说"有代码但不生效"，而不是让用户去改字段名。
func TestFixtureFind_OnlyOrphanMatches(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt600",
		Pages: []matPage{
			{Seq: "12", ID: "73", Record: "s_detail73", Columns: []string{"xmdducseq1"}},
		},
		Sections: []matSection{{
			ID: "p.input", Ver: "1", Src: "s", Slots: []matSlot{
				{Name: "input.a.page73.xmdducseq1", Label: "活点", Body: []string{"  LET a = 1"}},
			},
		}},
		Orphans: []matOrphan{
			{Name: "input.a.page73.xmdducseq", Src: "c", Ver: "2", Body: "LET x = 1\n"},
		},
	})
	set.mapIndex()
	res, err := Find(FindRequest{LinkRequest: set.LinkRequest(), Field: "xmdducseq"})
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	// 注意：xmdducseq 是 xmdducseq1 的后缀吗？不是（后缀是相反方向），所以只命中孤儿。
	if res.Total != 0 || len(res.Points) != 0 {
		t.Errorf("孤儿不该计入 Points: %+v", res.Points)
	}
	if !strings.Contains(res.Note, "没有槽位") || !strings.Contains(res.Note, "input.a.page73.xmdducseq") {
		t.Errorf("孤儿提示 = %q", res.Note)
	}
}
