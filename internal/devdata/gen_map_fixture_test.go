package devdata

// gen map 的**磁盘材料矩阵**：每条测试都从 fixtures_test.go 的构建器起一套真实材料，
// 而不是直接喂内存结构体。
//
// 为什么这一层必须存在：buildLink 的纯内存测试已经覆盖了配对算法本身，但它测不到
// "材料怎么读进来" —— manifest 的 srcMtime 决定 freshness、.tsd/.4fd 决定标签与页签、
// mdl 目录决定脱版。这些恰恰是真机上最容易错、也最容易"看起来对"的部分。
//
// 场景与真机证据：
//   - 多页/单页：.tab 的 page 是页签序号，page_id 才是点名用的；
//   - 报表/批次：没有 body 页面字段（assembly 仍然有）；
//   - freestyle：非生成器产物，没有 section 锚；
//   - 无 .tap：从未被设计器接管的标准程序；
//   - .4gl 比 .tap 旧：改了蓝图没跑 F7 的直接证据；
//   - 槽位不一致：只在 tap（4gl 旧）/ 只在 4gl（生成器自加的注释重复点）；
//   - 标记行名字截断、重名点、孤儿声明、cusver=0000、页签桥用 text、
//     切片标记与当前模板 version 不同（脱版）。

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// matCheckLevel 取一条检查的 level（没有这条检查返回 "missing"）。
func matCheckLevel(doc *LinkDoc, id string) string {
	for i := range doc.Checks {
		if doc.Checks[i].ID == id {
			return doc.Checks[i].Level
		}
	}
	return "missing"
}

// matCheckDetail 取一条检查的 detail。
func matCheckDetail(doc *LinkDoc, id string) string {
	for i := range doc.Checks {
		if doc.Checks[i].ID == id {
			return doc.Checks[i].Detail
		}
	}
	return ""
}

// TestFixtureMap_CompleteAllChecksGreen 是"一套完整材料"的基准形态：
// 两侧 section/slot 对齐、声明挂在点上、切片版本与模板一致、4gl 不旧于 tap。
// 它同时钉住 link.json 的落盘形态（Map 的最终兑现物）。
func TestFixtureMap_CompleteAllChecksGreen(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt500_wf",
		Pages: []matPage{
			{Seq: "1", ID: "1", Record: "s_browse", Columns: []string{"f1", "f2"}},
			{Seq: "12", ID: "73", Record: "s_detail73", Columns: []string{"xmdduc001"}},
		},
		Sections: []matSection{
			{ID: "p.input", Ver: "56", Src: "s", Slots: []matSlot{
				{Name: "input.a.page1.f1", Label: "過欄位檢查", DeclBody: "LET f1 = 1\n", Src: "c", Ver: "3", Body: []string{"   LET f1 = 1"}},
				// 客制点挂到标准槽位上：mapping 与 name 不同，两者都要留。
				{Name: "input.a.page73.xmdduc001", Label: "刻号检核", DeclBody: "IF x THEN MESSAGE 'x' END IF\n", Src: "c", Ver: "4", Mapping: "input.a.page73.xmdduc001", Body: []string{"   IF x THEN", "      MESSAGE 'x'", "   END IF"}},
			}},
			{ID: "p.construct", Ver: "12", Src: "s", Slots: []matSlot{
				{Name: "construct.c.page73.xmdduc001", Label: "ON ACTION", DeclBody: "LET std = 1\n", Src: "s", Ver: "9", Body: []string{"   LET std = 1"}},
			}},
		},
		Slices: []matSlice{
			{Code: "a17", Used: 3, Current: 3, InBody: true},
			{Code: "a05", Used: 2, Current: 2, InBody: true},
		},
		TSDLabels: map[string]string{"f1": "料号", "xmdduc001": "刻号"},
		TSDPages:  map[string]string{"page_1": "基本数据", "page_73": "示例页签"},
	})

	doc := set.mapIndex()

	// —— 索引身份与来源指纹 ——
	if doc.Schema != LinkSchema || doc.Prog != "axmt500_wf" || doc.Env != "e" {
		t.Fatalf("索引身份 = %q/%q/%q", doc.Schema, doc.Prog, doc.Env)
	}
	if doc.Sources.Tab == nil || doc.Sources.Tsd == nil || doc.Sources.Fd == nil {
		t.Fatalf("三份画面侧材料都该记来源: %+v", doc.Sources)
	}
	if doc.Sources.Tap.SHA256 == "" || doc.Sources.G4.SHA256 == "" {
		t.Errorf("产物/骨架缺 sha256: %+v", doc.Sources)
	}
	if doc.Sources.Tap.Rel != "dzx/axmt500_wf.tap" || doc.Sources.G4.Rel != "src/axmt500_wf.4gl" {
		t.Errorf("来源相对路径 = %q / %q", doc.Sources.Tap.Rel, doc.Sources.G4.Rel)
	}
	if doc.Sources.Tap.ModTime != matDefaultTapMTime {
		t.Errorf("srcMtime 应从 manifest 读（服务器时间），得到 %q", doc.Sources.Tap.ModTime)
	}
	// 4gl 比 tap 新 → freshness 必须 ok，不能靠 os.Stat 的"刚刚"骗过去。
	if lv := matCheckLevel(doc, "freshness"); lv != "ok" {
		t.Errorf("freshness = %q，想要 ok", lv)
	}

	// —— 判型/版次 ——
	if doc.Mode.Type != "t01" || doc.Mode.Industry != "std" || doc.Mode.Module != "cxm" {
		t.Errorf("判型 = %+v", doc.Mode)
	}
	if doc.Mode.Identity != "c" || doc.Mode.SectionFlag != "Y" || doc.Mode.CusVer != "0003" {
		t.Errorf("身份/版次 = %+v", doc.Mode)
	}
	if doc.Mode.Pages != 2 || doc.Mode.Freestyle || doc.Mode.UnfoldedSection {
		t.Errorf("页数/freestyle = %+v", doc.Mode)
	}

	// —— 点位与声明 ——
	if doc.Stats.Points != 3 || doc.Stats.SlotsTap != 3 || doc.Stats.SlotsG4 != 3 {
		t.Fatalf("统计 = %+v", doc.Stats)
	}
	if doc.Stats.SectionsTap != 2 || doc.Stats.SectionsG4 != 2 {
		t.Errorf("section 数 = %d/%d", doc.Stats.SectionsTap, doc.Stats.SectionsG4)
	}
	if doc.Stats.PointsCustom != 2 || doc.Stats.CustomLines != 2 {
		t.Errorf("客制统计 = %+v（两条客制声明各 1 行）", doc.Stats)
	}
	// 声明册：3 条声明，其中 2 条客制。
	if doc.Stats.DeclTotal != 3 || doc.Stats.DeclCustom != 2 || doc.Stats.DeclOrphans != 0 {
		t.Errorf("声明统计 = %+v", doc.Stats)
	}
	for _, p := range doc.Points {
		switch p.Name {
		case "input.a.page1.f1":
			if !p.Custom || p.Src != "c" || p.Preview == "" {
				t.Errorf("点 %s = %+v（应带客制身份与一行摘要）", p.Name, p)
			}
		case "input.a.page73.xmdduc001":
			// mapping 与 name 不同时，Slot 要留挂靠的标准槽位（客制点挂标准洞）。
			if !p.Custom || p.Slot != "input.a.page73.xmdduc001" {
				t.Errorf("点 %s = %+v（mapping 要留在 Slot 上）", p.Name, p)
			}
		case "construct.c.page73.xmdduc001":
			if p.Custom || p.Src != "s" || p.Kind() != "标准" {
				t.Errorf("标准点 %s = %+v / %s", p.Name, p, p.Kind())
			}
		}
	}

	// —— 每类检查都在（这是"索引可信"的兑现处）——
	wantChecks := map[string]string{
		"sections-pair": "ok",
		"slots-pair":    "ok",
		"slice-version": "ok",
		"freshness":     "ok",
		"blueprint":     "ok",
	}
	for id, want := range wantChecks {
		if got := matCheckLevel(doc, id); got != want {
			t.Errorf("检查 %s = %q，想要 %q（detail=%q）", id, got, want, matCheckDetail(doc, id))
		}
	}

	// —— 切片：用量与脱版 ——
	if doc.Slices.Total != 2 || doc.Slices.Distinct != 2 {
		t.Errorf("切片 = %+v", doc.Slices)
	}
	if len(doc.Slices.Drift) != 0 {
		t.Errorf("版本一致时不该报脱版: %+v", doc.Slices.Drift)
	}
	if len(doc.Slices.Markers) != 2 {
		t.Errorf("切片标记行 = %+v", doc.Slices.Markers)
	}

	// —— 页面/字段：.tab → .4fd(text) → .tsd 的整条桥 ——
	if len(doc.Pages) != 2 {
		t.Fatalf("页面 = %+v", doc.Pages)
	}
	if doc.Pages[0].PageID != "1" || doc.Pages[0].PageName != "page_1" || doc.Pages[0].Title != "基本数据" {
		t.Errorf("页 0 = %+v（页签键必须来自 4fd 的 text）", doc.Pages[0])
	}
	if doc.Pages[1].PageID != "73" || doc.Pages[1].PageName != "page_73" || doc.Pages[1].Title != "示例页签" {
		t.Errorf("页 1 = %+v", doc.Pages[1])
	}
	if doc.Pages[1].Fields != 1 || doc.Pages[1].Columns != 1 {
		t.Errorf("页 1 字段数 = %d/%d", doc.Pages[1].Fields, doc.Pages[1].Columns)
	}
	if len(doc.Fields) != 2 || doc.Stats.FieldsLinked != 2 {
		t.Fatalf("字段 = %+v", doc.Fields)
	}
	if doc.Fields[0].Label != "料号" || doc.Fields[1].Label != "刻号" {
		t.Errorf("字段标签 = %+v", doc.Fields)
	}
	if doc.Stats.ColumnsTotal != 3 {
		t.Errorf("蓝图字段总数 = %d，想要 3", doc.Stats.ColumnsTotal)
	}

	// —— link.json 落盘形态 ——
	if doc.Path != set.linkJSONPath() {
		t.Errorf("索引落点 = %q", doc.Path)
	}
	back := set.readLinkJSON()
	if back.Schema != LinkSchema || back.Prog != doc.Prog || len(back.Points) != len(doc.Points) {
		t.Errorf("落盘/回读不一致: %+v", back)
	}
	if back.Sources.Tsd == nil || back.Pages[1].Title != "示例页签" {
		t.Errorf("落盘 JSON 少了 tsd/标题: %+v", back.Sources)
	}
	if fi, err := filepath.Abs(set.linkJSONPath()); err != nil || !strings.HasSuffix(fi, "link.json") {
		t.Errorf("索引路径形态 = %q (%v)", fi, err)
	}
}

// TestFixtureMap_ReportWithoutPageFields 钉住"报表/批次版型"：assembly 有版型但
// 一个 body 页面都没有 → Pages 空、Fields 空、ColumnsTotal=0，而点位仍在。
// 同时钉住 cusver=0000（从未客制）在 blueprint 检查里的口径。
func TestFixtureMap_ReportWithoutPageFields(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog:     "axmr500",
		Type:     "r01",
		Industry: "_wf",
		CusVer:   "0000",
		NoPages:  true, // 报表没有画面页签字段
		No4fd:    true, // 报表也没有表单定义
		NoTsd:    true,
		Sections: []matSection{{
			ID: "p.report", Ver: "3", Src: "s", Slots: []matSlot{
				{Name: "report.prepare.header", Label: "報表前置", DeclBody: "LET h = 1\n", Src: "c", Ver: "2", Body: []string{"   LET h = 1"}},
			},
		}},
	})
	doc := set.mapIndex()

	if doc.Mode.Type != "r01" || doc.Mode.Industry != "_wf" {
		t.Errorf("报表判型 = %+v", doc.Mode)
	}
	if doc.Stats.Points != 1 || doc.Stats.PointsCustom != 1 {
		t.Errorf("报表点位 = %+v", doc.Stats)
	}
	if len(doc.Pages) != 0 || len(doc.Fields) != 0 {
		t.Errorf("报表不该有页面/字段: pages=%+v fields=%+v", doc.Pages, doc.Fields)
	}
	if doc.Stats.ColumnsTotal != 0 || doc.Stats.FieldsLinked != 0 {
		t.Errorf("报表字段统计 = %+v", doc.Stats)
	}
	// 报表点位名里没有 .pageN. 形态 → 不能被误当成字段。
	if got := checkByID(doc, "blueprint"); got == nil || got.Level != "ok" {
		t.Fatalf("blueprint = %+v", got)
	} else if !strings.Contains(got.Detail, "版型 r01") {
		t.Errorf("blueprint detail = %q", got.Detail)
	}
	if !strings.Contains(matCheckDetail(doc, "blueprint"), "cusver=0000：从未客制") {
		t.Errorf("blueprint 应说明从未客制: %q", matCheckDetail(doc, "blueprint"))
	}
}

// TestFixtureMap_FreestyleNoSection 钉住 freestyle：.4gl 不是生成器产物（没有
// `{<section>}` 锚）。两种情况都要能建索引、且 sections-pair 给出准确原因：
//  1. .tap 也没 section（程序不是分段落版型）；
//  2. .tap 有 section 而 .4gl 没有（产物不是同一版生成器出的）。
func TestFixtureMap_FreestyleNoSection(t *testing.T) {
	// 情况 1：两侧都没有 section 锚。
	bare := newMaterials(t, matOpts{
		Prog: "axmt100_fz", Freestyle: true, NoTap: true, NoTab: true,
	})
	d1 := bare.mapIndex()
	if !d1.Mode.Freestyle {
		t.Errorf("Mode.Freestyle = false（应判非生成器产物）: %+v", d1.Mode)
	}
	if d1.Stats.SectionsTap != 0 || d1.Stats.SectionsG4 != 0 || d1.Stats.Points != 0 {
		t.Errorf("freestyle 无锚时的统计 = %+v", d1.Stats)
	}
	if lv := matCheckLevel(d1, "sections-pair"); lv != "warn" {
		t.Errorf("sections-pair = %q，想要 warn", lv)
	}
	if !strings.Contains(matCheckDetail(d1, "sections-pair"), "没有 section") {
		t.Errorf("sections-pair detail = %q", matCheckDetail(d1, "sections-pair"))
	}

	// 情况 2：.tap 有 section（设计器那边是分段的），.4gl 却 freestyle。
	mixed := newMaterials(t, matOpts{
		Prog:      "axmt100_fz",
		Freestyle: true,
		Sections: []matSection{{
			ID: "p.main", Ver: "1", Src: "s",
			Slots: []matSlot{{Name: "main.run", Label: "主程式", Body: []string{"  LET x = 1"}}},
		}},
		NoTab: true,
	})
	d2 := mixed.mapIndex()
	if !d2.Mode.Freestyle || d2.Stats.SectionsG4 != 0 {
		t.Errorf("混合形态判型 = %+v / %+v", d2.Mode, d2.Stats)
	}
	// tap 有洞、4gl 里根本没有 section → 洞落地不了，只能记 missing。
	if d2.Stats.PointsMissing != 1 || len(d2.Orphans) != 0 {
		t.Errorf("missing/orphan = %d/%+v", d2.Stats.PointsMissing, d2.Orphans)
	}
	if lv := matCheckLevel(d2, "sections-pair"); lv != "warn" {
		t.Errorf("sections-pair = %q，想要 warn（4gl 没有 section 锚）", lv)
	}
	if !strings.Contains(matCheckDetail(d2, "sections-pair"), "4gl 里没有 section") &&
		!strings.Contains(matCheckDetail(d2, "sections-pair"), ".4gl 里没有 section") {
		t.Errorf("sections-pair detail = %q", matCheckDetail(d2, "sections-pair"))
	}
}

// TestFixtureMap_NoTapStandardProgram 钉住"从未被设计器接管的标准程序"：
// 服务器上端到端没有 .tap，gen 仍要能建索引（点位全部 g4Only、无法判客制/标准）。
func TestFixtureMap_NoTapStandardProgram(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog:  "axmt200",
		NoTap: true,
		Pages: []matPage{{Seq: "1", ID: "1", Record: "s_detail1", Columns: []string{"f1"}}},
		NoTsd: true, No4fd: true,
		Sections: []matSection{
			{ID: "p.main", Ver: "1", Src: "s", Slots: []matSlot{
				{Name: "main.run", Label: "主程式", Body: []string{"  LET x = 1"}},
			}},
			{ID: "p.input", Ver: "2", Src: "s", Slots: []matSlot{
				{Name: "input.a.page1.f1", Label: "過欄位", Body: []string{"  LET f1 = 1"}},
			}},
		},
	})
	doc := set.mapIndex()

	if doc.Sources.Tap.SHA256 != "" || doc.Sources.Tap.Rel != "" {
		t.Errorf("没有 .tap 却记了来源: %+v", doc.Sources.Tap)
	}
	if doc.Stats.SlotsTap != 0 || doc.Mode.Identity != "" {
		t.Errorf("没有 .tap 时槽位/身份 = %+v / %+v", doc.Stats, doc.Mode)
	}
	if doc.Stats.Points != 2 {
		t.Fatalf("点位 = %+v", doc.Points)
	}
	for _, p := range doc.Points {
		if !p.G4Only || p.Kind() != "未知（无 .tap）" {
			t.Errorf("点 %s 应判 g4Only: %+v / %s", p.Name, p, p.Kind())
		}
	}
	if lv := matCheckLevel(doc, "sections-pair"); lv != "warn" {
		t.Errorf("sections-pair = %q，想要 warn", lv)
	}
	if doc.Stats.FieldsLinked != 1 || len(doc.Fields) != 1 {
		t.Errorf("没有 .tap 时字段仍应从点名推出: %+v", doc.Fields)
	}
	if got := checkByID(doc, "blueprint"); got == nil || got.Level != "ok" {
		t.Errorf("blueprint = %+v", got)
	}
}

// TestFixtureMap_StaleG4OlderThanTap 钉住"改了 .tap 没跑 F7（产生程式）"：
// manifest 里的 srcMtime 是服务器时间，4gl 更旧 → freshness 必须报 warn，
// 且当槽位只在 tap 里时 freshness-detail 要给出"重产一次即可对齐"。
func TestFixtureMap_StaleG4OlderThanTap(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog:     "axmt300",
		TapMTime: "2026-11-21T12:00:00+08:00",
		G4MTime:  "2026-11-21T09:00:00+08:00", // 旧 3 小时
		Sections: []matSection{{
			ID: "p.input", Ver: "1", Src: "s",
			Slots: []matSlot{
				{Name: "input.a.page1.f1", Label: "過欄位", Body: []string{"  LET f1 = 1"}},
				{Name: "input.a.page1.f9", Label: "新加的洞", TapOnly: true}, // 只在 tap：4gl 旧了
			},
		}},
	})
	doc := set.mapIndex()

	if lv := matCheckLevel(doc, "freshness"); lv != "warn" {
		t.Fatalf("freshness = %q，想要 warn（4gl 比 tap 旧）", lv)
	}
	if !strings.Contains(matCheckDetail(doc, "freshness"), "F7") {
		t.Errorf("freshness detail 没说清原因: %q", matCheckDetail(doc, "freshness"))
	}
	if lv := matCheckLevel(doc, "freshness-detail"); lv != "warn" {
		t.Errorf("freshness-detail = %q，想要 warn", lv)
	}
	if doc.Stats.PointsMissing != 1 {
		t.Errorf("missing = %d，想要 1", doc.Stats.PointsMissing)
	}
	if doc.Sources.G4.ModTime != "2026-11-21T09:00:00+08:00" {
		t.Errorf("G4 ModTime = %q（应取 manifest srcMtime）", doc.Sources.G4.ModTime)
	}
}

// TestFixtureMap_SlotMismatchBothDirections 钉住两侧槽位不一致的两种方向：
// 只在 .tap（4gl 旧、没落地）与只在 .4gl（生成器自加的注释重复点）。
// 两个方向分在不同 section 里，避免 pairSlots 的"剩余按顺序配"掩盖结论。
func TestFixtureMap_SlotMismatchBothDirections(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt400",
		Sections: []matSection{
			{ID: "p.old", Ver: "1", Src: "s", Slots: []matSlot{
				{Name: "old.a.page1.gone", Label: "tap 有 4gl 无", TapOnly: true},
			}},
			{ID: "p.extra", Ver: "2", Src: "s", Slots: []matSlot{
				{Name: "extra.b.page1.dup", Label: "4gl 自加", G4Only: true, Commentd: true},
			}},
		},
	})
	doc := set.mapIndex()

	if doc.Stats.PointsMissing != 1 || doc.Stats.PointsExtra != 1 || doc.Stats.PointsCommented != 1 {
		t.Fatalf("统计 = %+v（想要 missing=1 extra=1 commented=1）", doc.Stats)
	}
	var gotMissing, gotExtra bool
	for _, p := range doc.Points {
		switch p.Name {
		case "old.a.page1.gone":
			gotMissing = p.Missing && p.Kind() == "未落地"
		case "extra.b.page1.dup":
			gotExtra = p.Extra && p.Commented && p.Kind() == "生成器自加"
		}
	}
	if !gotMissing || !gotExtra {
		t.Errorf("点位标记 = missing:%v extra:%v（%+v）", gotMissing, gotExtra, doc.Points)
	}
	if lv := matCheckLevel(doc, "slots-pair"); lv != "warn" {
		t.Errorf("slots-pair = %q，想要 warn", lv)
	}
	if lv := matCheckLevel(doc, "slots-commented"); lv != "ok" {
		t.Errorf("slots-commented = %q，想要 ok", lv)
	}
}

// TestFixtureMap_TruncatedMarkerAndDuplicatePoints 钉住两个真机形态：
//   - `#add-point:… name="chk_` 被截断：完整名字在洞里，配对按名字+次序；
//     两侧名字不同要留证（g4Name），不能假装一致；
//   - 重名点（实测 construct.c.page1.xmdc027 ×5）：键是"名字+出现次序"，
//     多出来的（被注释掉的生成器自加点）记 extra，不能算 missing。
func TestFixtureMap_TruncatedMarkerAndDuplicatePoints(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt500",
		Sections: []matSection{
			{ID: "p.chk", Ver: "2", Src: "s", Slots: []matSlot{
				{Name: "chk_data_exist.define", Label: "chk段define", TruncatedMarker: true,
					DeclBody: "DEFINE x\n", Src: "s", Ver: "2", Body: []string{"   DEFINE x"}},
			}},
			{ID: "p.construct", Ver: "3", Src: "s", Slots: []matSlot{
				{Name: "construct.c.page1.xmdc027", Label: "a", Body: []string{"   LET a = 1"}},
				{Name: "construct.c.page1.xmdc027", Label: "b", Body: []string{"   LET b = 2"}},
				{Name: "construct.c.page1.xmdc027", Label: "c", Body: []string{"   LET c = 3"}},
				{Name: "construct.c.page1.xmdc027", Label: "d", G4Only: true, Commentd: true},
				{Name: "construct.c.page1.xmdc027", Label: "e", G4Only: true, Commentd: true},
			}},
		},
	})
	doc := set.mapIndex()

	// 截断点：名字取洞（完整名），g4Name 记截断后的名字。
	var trunc *LinkPoint
	for i := range doc.Points {
		if doc.Points[i].Name == "chk_data_exist.define" {
			trunc = &doc.Points[i]
		}
	}
	if trunc == nil {
		t.Fatalf("截断点没配上: %+v", doc.Points)
	}
	if trunc.G4Name == "" || trunc.G4Name == trunc.Name {
		t.Errorf("应留证两侧名字差异: name=%q g4Name=%q", trunc.Name, trunc.G4Name)
	}
	// 注意：声明 src=s 时 Custom=false，这里显式钉住"标准声明的截断点"身份。
	if trunc.DeclLines != 1 || trunc.Custom {
		t.Errorf("截断点声明 = %+v", trunc)
	}

	// 重名点：5 个点位（3 有效 + 2 注释掉的 extra），按行号排序。
	if doc.Stats.Points != 6 { // 1 截断 + 5 重名
		t.Fatalf("点位 = %d，想要 6：%+v", doc.Stats.Points, set.pointNames(doc))
	}
	if doc.Stats.PointsExtra != 2 || doc.Stats.PointsMissing != 0 {
		t.Errorf("extra/missing = %d/%d，想要 2/0", doc.Stats.PointsExtra, doc.Stats.PointsMissing)
	}
	if doc.Stats.PointsCommented != 2 {
		t.Errorf("commented = %d，想要 2", doc.Stats.PointsCommented)
	}
	// 3 个有效重名点必须是 3 个独立点位（不是被去重成 1 个）。
	dups := pointsByName(doc, "construct.c.page1.xmdc027")
	if len(dups) != 5 {
		t.Errorf("重名点 = %d 个，想要 5", len(dups))
	}
}

// TestFixtureMap_OrphansAndUnmarkedFunctions 钉住声明侧的两种"落不了地"：
//   - 孤儿声明（骨架没洞、4gl 里也没有）→ 重产不会保留，必须进 Orphans 并分级
//     （客制孤儿才是"客户代码丢了"，标准残留只是噪音）；
//   - `function.*` 以函数形式落地但无加点标记 → 按函数名定位行号，不能当孤儿。
func TestFixtureMap_OrphansAndUnmarkedFunctions(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt600",
		Sections: []matSection{{
			ID: "p.input", Ver: "1", Src: "s", Slots: []matSlot{
				{Name: "input.a.page1.f1", Label: "正常点", DeclBody: "LET f1 = 1\n", Src: "c", Ver: "2", Body: []string{"   LET f1 = 1"}},
			},
		}},
		Orphans: []matOrphan{
			{Name: "input.a.page9.gone", Src: "c", Ver: "3", Body: "IF x THEN MESSAGE 'gone' END IF\n"},
			{Name: "input.a.page9.stdgone", Src: "s", Ver: "1", Body: "LET std = 1\n"},
		},
		Funcs: []matFunc{{Name: "p_feature", Body: []string{"  LET y = 2", "  RETURN y"}}},
	})
	doc := set.mapIndex()

	if doc.Stats.DeclTotal != 4 || doc.Stats.DeclOrphans != 2 || doc.Stats.DeclOrphansCustom != 1 {
		t.Fatalf("声明统计 = %+v", doc.Stats)
	}
	if len(doc.Orphans) != 2 {
		t.Fatalf("孤儿 = %+v", doc.Orphans)
	}
	for _, o := range doc.Orphans {
		if o.Reason != "no-slot" || o.Lines == 0 || o.Preview == "" {
			t.Errorf("孤儿缺原因/摘要: %+v", o)
		}
	}
	if lv := matCheckLevel(doc, "decls-orphan"); lv != "warn" {
		t.Errorf("decls-orphan = %q，想要 warn", lv)
	}
	detail := matCheckDetail(doc, "decls-orphan")
	if !strings.Contains(detail, "客制 1 个") || !strings.Contains(detail, "标准声明的残留") {
		t.Errorf("decls-orphan detail 没说清分级: %q", detail)
	}

	// unmarked：声明以函数形式落地，行号由函数名定位。
	if doc.Stats.DeclUnmarked != 1 {
		t.Fatalf("unmarked = %d，想要 1", doc.Stats.DeclUnmarked)
	}
	var fn *LinkPoint
	for i := range doc.Points {
		if doc.Points[i].Unmarked {
			fn = &doc.Points[i]
		}
	}
	if fn == nil || fn.Name != "function.p_feature" || fn.Line0 == 0 || fn.Line1 < fn.Line0 {
		t.Fatalf("客制函数定位 = %+v", fn)
	}
	// 客制函数是最具体的一类：unmarked + Custom 必须先判，否则"客制函数"标签不可达。
	if fn.Kind() != "客制函数" {
		t.Errorf("Kind = %q，想要 客制函数", fn.Kind())
	}
	if !fn.Custom || !fn.Unmarked {
		t.Errorf("unmarked 点 = %+v", fn)
	}
	if lv := matCheckLevel(doc, "decls-unmarked"); lv != "ok" {
		t.Errorf("decls-unmarked = %q，想要 ok", lv)
	}
	// 函数行号要真的指向 .4gl 里的 FUNCTION 块（不是 0）。
	if fn.Preview == "" {
		t.Errorf("函数点缺摘要: %+v", fn)
	}
}

// TestFixtureMap_PageBridgeUsesTextNotName 钉住 .4fd 的页签桥规则：
// 页签键取 **text**（不是 name 控件 id）；text 里没有 "page" 时按真标题用。
// 这条链一旦接错，客户说的"第 12 个页签"会整页错位。
func TestFixtureMap_PageBridgeUsesTextNotName(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt700",
		Pages: []matPage{
			{Seq: "1", ID: "1", Record: "s_browse", Columns: []string{"f1"}},
			{Seq: "12", ID: "73", Record: "s_detail73", Columns: []string{"f2"}},
			{Seq: "2", ID: "2", Record: "s_detail2", Columns: []string{"f3"}},
		},
		// 真机形态：主档页 name="master_Folder_page" text="page_1"；
		// 备注页 name="page_1" text="page_12"（取 name 会接成"基本数据"）。
		FDPages: []matFdPage{
			{Name: "master_Folder_page", Key: "page_1", Record: "s_browse"},
			{Name: "page_1", Key: "page_12", Record: "s_detail73"},
			{Name: "page_2", Key: "订单明细", Record: "s_detail2"},
		},
		Sections: []matSection{{
			ID: "p.input", Ver: "1", Src: "s", Slots: []matSlot{
				{Name: "input.a.page1.f1", Label: "f1", Body: []string{"  LET f1 = 1"}},
				{Name: "input.a.page73.f2", Label: "f2", Body: []string{"  LET f2 = 1"}},
				{Name: "input.a.page2.f3", Label: "f3", Body: []string{"  LET f3 = 1"}},
			},
		}},
		TSDPages: map[string]string{"page_1": "基本数据", "page_12": "备注信息"},
	})
	doc := set.mapIndex()

	byID := map[string]LinkPage{}
	for _, p := range doc.Pages {
		byID[p.PageID] = p
	}
	if got := byID["1"]; got.PageName != "page_1" || got.Title != "基本数据" {
		t.Errorf("主档页 = %+v（键应取 text=page_1）", got)
	}
	if got := byID["73"]; got.PageName != "page_12" || got.Title != "备注信息" {
		t.Errorf("备注页 = %+v（name 是 page_1，取 name 会错接成基本数据）", got)
	}
	if got := byID["2"]; got.PageName != "订单明细" || got.Title != "订单明细" {
		t.Errorf("标题型 text = %+v（不含 page 的 text 按真标题用）", got)
	}
}

// TestFixtureMap_SliceDriftAndMarkers 钉住切片脱版：产物里用 Version:2、
// 当前模板（mdl 目录里的 code_XXX.template 首行）是 Version:3 → slice-version warn。
// 同时钉住 markers（行号→切片）与 ByCode 聚合，供 At/Grep 反查。
func TestFixtureMap_SliceDriftAndMarkers(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt800",
		Slices: []matSlice{
			{Code: "a17", Used: 2, Current: 3, InBody: true}, // 脱版
			{Code: "a05", Used: 2, Current: 2},               // 一致，写在 section 外
		},
	})
	doc := set.mapIndex()

	if doc.Stats.SlicesTotal != 2 || doc.Stats.SlicesDistinct != 2 {
		t.Errorf("切片统计 = %+v", doc.Stats)
	}
	if len(doc.Slices.Drift) != 1 {
		t.Fatalf("脱版 = %+v（想要只有 a17）", doc.Slices.Drift)
	}
	d := doc.Slices.Drift[0]
	if d.Code != "a17" || d.Used != 2 || d.Current != 3 {
		t.Errorf("脱版项 = %+v", d)
	}
	if lv := matCheckLevel(doc, "slice-version"); lv != "warn" {
		t.Errorf("slice-version = %q，想要 warn", lv)
	}
	if !strings.Contains(matCheckDetail(doc, "slice-version"), "旧版模板") {
		t.Errorf("slice-version detail = %q", matCheckDetail(doc, "slice-version"))
	}
	// markers 是"行号 → 切片"的反查表：a17 写在槽位正文里、a05 写在 section 前，
	// 顺序按行号（a05 在前），不能假设与声明顺序一致。
	if len(doc.Slices.Markers) != 2 {
		t.Fatalf("切片标记 = %+v", doc.Slices.Markers)
	}
	mark := map[string]SliceMark{}
	for _, m := range doc.Slices.Markers {
		mark[m.Code] = m
	}
	if got := mark["a17"]; got.Ver != 2 || got.Line <= 1 {
		t.Errorf("a17 标记 = %+v（应带产物 Version 与行号）", got)
	}
	if got := mark["a05"]; got.Ver != 2 || got.Line != 1 {
		t.Errorf("a05 标记 = %+v（写在 section 前，第一行）", got)
	}
	if len(doc.Slices.ByCode) != 2 || doc.Slices.ByCode[0].Code != "a05" || doc.Slices.ByCode[0].Count != 1 {
		t.Errorf("按代号聚合 = %+v", doc.Slices.ByCode)
	}
}

// TestFixtureMap_CustomFunctionMissingFromG4IsOrphan 钉住一种隐蔽形态：
// .tap 里声明了 `function.xxx`，但 .4gl 里这个函数根本不存在（客制代码丢了）。
// 这时不能按 unmarked 算（那是"落地了但没标记"），必须进孤儿并说明原因。
func TestFixtureMap_CustomFunctionMissingFromG4IsOrphan(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt900",
		Orphans: []matOrphan{
			{Name: "function.p_gone", Src: "c", Ver: "2", Body: "PRIVATE FUNCTION p_gone()\nEND FUNCTION\n"},
		},
	})
	doc := set.mapIndex()
	if doc.Stats.DeclUnmarked != 0 {
		t.Errorf("函数不在 .4gl 里，不该算 unmarked: %+v", doc.Stats)
	}
	if doc.Stats.DeclOrphans != 1 || len(doc.Orphans) != 1 || doc.Orphans[0].Name != "function.p_gone" {
		t.Fatalf("孤儿 = %+v", doc.Orphans)
	}
	if lv := matCheckLevel(doc, "decls-unmarked"); lv != "missing" {
		t.Errorf("没有 unmarked 时不该有 decls-unmarked 检查，得到 %q", lv)
	}
}

// TestFixtureMap_MultiPageVsSinglePage 把"多页/单页"放在同一条测试里对照：
// 页数来自 .tab 的 body 个数与 Mode.Pages；每个字段必须落到正确的 page_id。
func TestFixtureMap_MultiPageVsSinglePage(t *testing.T) {
	single := newMaterials(t, matOpts{Prog: "axmt010", NoTsd: true, No4fd: true})
	d1 := single.mapIndex()
	if d1.Mode.Pages != 1 || len(d1.Pages) != 1 || d1.Stats.ColumnsTotal != 1 {
		t.Errorf("单页形态 = %+v / %+v（缺省单槽位只推导出 1 列）", d1.Mode, d1.Pages)
	}

	multi := newMaterials(t, matOpts{
		Prog:  "axmt011",
		NoTsd: true, No4fd: true,
		Sections: []matSection{{
			ID: "p.input", Ver: "1", Src: "s", Slots: []matSlot{
				{Name: "input.a.page12.f1", Label: "f1", Body: []string{"  LET f1 = 1"}},
				{Name: "input.a.page73.f2", Label: "f2", Body: []string{"  LET f2 = 1"}},
				{Name: "input.a.page73.f3", Label: "f3", Body: []string{"  LET f3 = 1"}},
			},
		}},
	})
	d2 := multi.mapIndex()
	if d2.Mode.Pages != 2 || len(d2.Pages) != 2 {
		t.Fatalf("多页形态 = %+v / %+v", d2.Mode, d2.Pages)
	}
	if d2.Pages[0].PageID != "12" || d2.Pages[0].Page != "1" || d2.Pages[0].Fields != 1 {
		t.Errorf("页 0 = %+v（序号与 page_id 是两回事）", d2.Pages[0])
	}
	if d2.Pages[1].PageID != "73" || d2.Pages[1].Fields != 2 || d2.Pages[1].Columns != 2 {
		t.Errorf("页 1 = %+v", d2.Pages[1])
	}
	if got := fmt.Sprint(d2.Pages[1].PageID); got != "73" {
		t.Errorf("page_id = %s", got)
	}
}
