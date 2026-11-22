package task

// render.go 的逐函数用例。
//
// 为什么直接喂结构体而不是走命令：渲染层只换展示形态、不换口径（见 render.go 顶部），
// 它的分支密度远高于命令层（"有/没有页签""已 apply/未 apply""截断/不截断"），
// 走命令要造十几种工作区状态才能踩到每条分支，而直接喂结果对象能一条用例覆盖一片。
// 真正需要"命令 → 领域 → 渲染"串起来验的部分已由 cli_test.go / verbs_test.go 覆盖。
//
// 断言用中文关键词而不是整份快照：这些文案是给人读的契约（说清"是不是重产会保留"
// "有没有被截断"），钉住关键词比钉整段排版更稳。

import (
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/store"
	"github.com/luyaogod/TT/internal/dev/taskws"
	"github.com/luyaogod/TT/internal/devdata"
)

// assertContains 是渲染用例的统一断言：一条输出里全部关键词都要在。
func assertContains(t *testing.T, out string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("输出缺 %q：\n%s", w, out)
		}
	}
}

// assertNotContains 断某个分支**没有**被误踩（例如"没页签"不该照样打页签表头）。
func assertNotContains(t *testing.T, out string, nots ...string) {
	t.Helper()
	for _, w := range nots {
		if strings.Contains(out, w) {
			t.Errorf("输出不该含 %q：\n%s", w, out)
		}
	}
}

// TestPrintOpen 覆盖 open 摘要的两个分支：索引有/没有（索引是可选步骤的产物）。
func TestPrintOpen(t *testing.T) {
	t.Run("全量结果", func(t *testing.T) {
		res := &taskws.CreateResult{
			Dir: "D:/work/axmt500_wf-task",
			Task: &taskws.Task{
				Prog: "axmt500_wf", Task: "task1", Env: "测试区", Module: "CXM", Identity: "c",
			},
			Fetch: &devdata.FetchResult{
				Files:   []devdata.FileEntry{{Size: 1024}, {Size: 2048}},
				Missing: []string{"form"},
			},
			Index: &taskws.LinkSummary{Points: 1234, Custom: 56, Sections: 4},
			Pack:  &devdata.PackResult{Path: "pkg/axmt500_wf(c).tzc", Bytes: 4096},
			Export: &taskws.ExportResult{
				Dir: "ws", Regions: 9, Stats: store.Stats{Editable: 7},
			},
			Warnings: []string{"git 提交失败（工作区仍可用）"},
		}
		out := renderOutput(t, func() { printOpen(res) })
		assertContains(t, out,
			"已建工作区：D:/work/axmt500_wf-task",
			"axmt500_wf · task1（测试区）", "模块 CXM", "身份 c",
			"材料 2 个文件 / 3.0 KB", "可选素材缺 1 项", "form",
			"1,234 点", "客制 56 点", "4 个 section",
			"pkg/axmt500_wf(c).tzc（4.0 KB）",
			"9 个区域，可编辑 7",
			"git 提交失败",
			"下一步：cd D:/work/axmt500_wf-task",
			"tt dev brief", "tt dev find --intent check",
		)
	})

	t.Run("没有索引与任务名", func(t *testing.T) {
		res := &taskws.CreateResult{
			Dir:   "D:/w",
			Task:  &taskws.Task{Prog: "p", Env: "e", Module: "m", Identity: "s"},
			Fetch: &devdata.FetchResult{},
			Pack:  &devdata.PackResult{Path: "p.tzc", Bytes: 10},
			Export: &taskws.ExportResult{
				Dir: "ws",
			},
		}
		out := renderOutput(t, func() { printOpen(res) })
		assertContains(t, out, "已建工作区：D:/w", "p（e）", "材料 0 个文件 / 0 B", "p.tzc（10 B）", "0 个区域，可编辑 0")
		assertNotContains(t, out, "索引")
	})
}

// TestPrintUpdate 覆盖 update 结果里几种"发生了什么"的组合。
func TestPrintUpdate(t *testing.T) {
	status := &taskws.Status{
		Prog: "p", Env: "e", Dir: "D:/ws",
		Index: taskws.IndexStatus{Points: 12, Custom: 3},
		Pkg:   taskws.PkgStatus{Path: "pkg/p(c).tzc", Present: true, Bytes: 1024, Modified: true},
	}

	t.Run("检查过期加真拉料加已apply", func(t *testing.T) {
		res := &taskws.UpdateResult{
			Checked:      true,
			Stale:        true,
			StaleReasons: []devdata.StaleReason{{Code: "remote-newer", Detail: "服务器文件大小变了"}},
			Fetch:        &devdata.FetchResult{Files: []devdata.FileEntry{{Size: 1}, {Size: 2}}},
			IndexRebuilt: true,
			Status:       status,
			Warnings:     []string{"取的是客制模块"},
			Notes:        []string{"索引重建：12 个点 / 客制 3 个"},
		}
		out := renderOutput(t, func() { printUpdate(res) })
		assertContains(t, out,
			"材料已过期（1 条原因）", "remote-newer", "服务器文件大小变了",
			"材料已重拉：2 个文件",
			"索引已重建：12 点 / 客制 3 点",
			"包里有已 apply 的成果",
			"取的是客制模块", "索引重建",
		)
	})

	t.Run("材料一致且重建包与视图", func(t *testing.T) {
		res := &taskws.UpdateResult{
			Checked:    true,
			Stale:      false,
			Repacked:   true,
			Reexported: true,
		}
		out := renderOutput(t, func() { printUpdate(res) })
		assertContains(t, out, "材料与服务器一致：没有过期项", "包与视图已按新材料重建（视图干净，没有丢失改动）")
		assertNotContains(t, out, "视图有未落盘改动")
	})

	t.Run("有未落盘改动仅刷材料", func(t *testing.T) {
		res := &taskws.UpdateResult{
			Fetch: &devdata.FetchResult{Files: []devdata.FileEntry{{Size: 1}}},
			Status: &taskws.Status{
				Prog: "p", Env: "e",
				Ws:  taskws.WsStatus{Present: true, Dirty: true},
				Pkg: taskws.PkgStatus{Present: true},
			},
		}
		out := renderOutput(t, func() { printUpdate(res) })
		assertContains(t, out, "材料已重拉：1 个文件", "视图有未落盘改动：保留视图，只更新了材料与索引（--force）")
	})

	t.Run("索引重建但材料未变", func(t *testing.T) {
		res := &taskws.UpdateResult{IndexRebuilt: true}
		out := renderOutput(t, func() { printUpdate(res) })
		assertContains(t, out, "材料未变")
		assertNotContains(t, out, "材料已重拉", "索引已重建")
	})
}

// TestPrintStatus 覆盖 status 的"最新/需重建/缺失/未 apply/不一致"各状态词。
func TestPrintStatus(t *testing.T) {
	t.Run("状态齐全且已apply", func(t *testing.T) {
		st := &taskws.Status{
			Prog: "p", Task: "task1", Env: "e", Ent: 99, Dir: "D:/ws",
			Materials: taskws.MaterialStatus{
				FetchedAt: "2026-11-21T20:00:00+08:00", Module: "CXM",
				Files: 12, Bytes: 2048, Missing: []string{"form"}, DZAA002: 7,
			},
			Index: taskws.IndexStatus{Present: true, Fresh: true, Points: 1234, Custom: 56},
			Ws: taskws.WsStatus{
				Present: true, Dirty: true, SectionState: "Unlocked",
				Regions: 10, Editable: 4, PkgSha: "abc",
			},
			Pkg: taskws.PkgStatus{
				Path: "pkg/p(c).tzc", Present: true, Bytes: 1024, Sha256: "x",
				Modified: true, MatchesBase: true,
			},
			Mdl:   taskws.MdlStatus{Present: true, Templates: 3},
			Notes: []string{"视图有未落盘的改动：先 tt dev apply"},
		}
		out := renderOutput(t, func() { printStatus(st) })
		assertContains(t, out,
			"p · task1（e@99）",
			"2026-11-21T20:00:00+08:00", "CXM", "12 个文件 / 2.0 KB", "可选素材缺 1 项",
			"规格版次 dzaa002=7",
			"1,234 点 / 客制 56 点", "最新",
			"有未落盘改动", "Unlocked", "10 区域（可编辑 4）",
			"p(c).tzc", "已 apply（与材料不同，可交付）",
			"3 个切片",
			"提示   视图有未落盘的改动",
		)
	})

	t.Run("不一致与各缺失态", func(t *testing.T) {
		st := &taskws.Status{
			Prog:      "p",
			Env:       "e",
			Materials: taskws.MaterialStatus{},
			Index:     taskws.IndexStatus{Present: true, Fresh: false},
			Ws:        taskws.WsStatus{Present: true, PkgSha: "old"},
			Pkg: taskws.PkgStatus{
				Path: "pkg/p(c).tzc", Present: true, Sha256: "new",
				MatchesBase: false,
			},
		}
		out := renderOutput(t, func() { printStatus(st) })
		assertContains(t, out,
			"材料   -", // 拉取时间/模块为空 → orDash
			"需重建",
			"干净",
			"与视图基线不一致（不可交付）",
			"模板   未复制（无切片脱版检测）",
		)
		assertNotContains(t, out, "规格版次", "个切片")
	})

	t.Run("索引与视图缺失", func(t *testing.T) {
		st := &taskws.Status{
			Prog: "p", Env: "e",
			Materials: taskws.MaterialStatus{FetchedAt: "2026-01-01T00:00:00+08:00", Module: "m"},
			Pkg:       taskws.PkgStatus{Path: ""},
		}
		out := renderOutput(t, func() { printStatus(st) })
		assertContains(t, out, "索引   0 点 / 客制 0 点", "·  缺失", "视图   缺失", "包     -", "未 apply", "0 B")
	})
}

// TestPrintBrief 补 brief 渲染里命令驱动踩不到的分支（空标题页签、非客制孤儿、检查排序）。
func TestPrintBrief(t *testing.T) {
	b := &devdata.BriefResult{
		Prog: "p", Env: "e",
		Mode: devdata.LinkMode{
			Type: "t01", Module: "CXM", Identity: "c", CusVer: "0003",
			Industry: "sd", UnfoldedSection: true, SectionFlag: "N",
		},
		Stats: devdata.LinkStats{
			G4Lines: 1234, SlotsTap: 56, PointsCustom: 2, CustomLines: 30,
			DeclUnmarked: 1, DeclOrphans: 2,
		},
		Sections: []devdata.LinkSection{{ID: "s1"}, {ID: "s2"}},
		Pages: []devdata.LinkPage{
			{Page: "1", PageID: "73", Record: "x", Columns: 2, Fields: 1, Title: "", PageName: "page_73"},
			{Page: "2", PageID: "74", Record: "", Columns: 1, Fields: 0, Title: "页签B"},
		},
		CustomPoints: []devdata.LinkPoint{
			{Name: "input.a.page73.xmdduc001", Line0: 17, DeclLines: 3, Preview: "IF x = 1"},
		},
		CustomTotal: 5,
		Orphans: []devdata.LinkOrphan{
			{Name: "input.a.page73.gone001", Src: "c", Lines: 2, Reason: "no-slot", Preview: "LET x = 1"},
			{Name: "global.gone", Src: "s", Lines: 1, Reason: "no-code"},
		},
		Slices: devdata.LinkSlices{
			Drift: []devdata.SliceDrift{{Code: "a05", Used: 3, Current: 5}},
		},
		Checks: []devdata.LinkCheck{
			{ID: "z-warn", Level: "warn", Detail: "warn 提示"},
			{ID: "a-error", Level: "error", Detail: "error 提示"},
			{ID: "ok-check", Level: "ok", Detail: "ok 提示"},
		},
		Intents: []devdata.Intent{{Key: "check", Hint: "栏位检核改 input.a"}},
		Notes:   []string{"下一步：tt dev find"},
	}
	out := renderOutput(t, func() { printBrief(b) })
	assertContains(t, out,
		"p（e）", "主模板 t01", "模块 CXM", "身份 c", "客制版次 0003", "行业 sd",
		"未开 Section", "section_flag=N",
		"代码 1,234 行 / 2 段 / 56 槽位", "客制 2 点（30 行）",
		"页签", "page_id=73", "?", "page_73", "页签B",
		"客制逻辑（按行数，共 5 个，列前 1）",
		"客制孤儿", "原因：.tap 里有声明，骨架里没有对应的洞",
		"其余 1 个孤儿是标准声明的残留",
		"风险", "error 提示", "warn 提示",
		"切片脱版", "a05", "产物 v3 / 当前 v5",
		"意图 → 落点", "check", "下一步：tt dev find",
	)
	// 排序：error 必须先于 warn 打印（先看会挂的）。
	if strings.Index(out, "error 提示") > strings.Index(out, "warn 提示") {
		t.Errorf("风险应按 error 在前排序：\n%s", out)
	}
	assertNotContains(t, out, "ok 提示")
}

// TestPrintBriefMinimal 覆盖"什么都没有"时的最小画像：不能因为空数据就崩/乱打。
func TestPrintBriefMinimal(t *testing.T) {
	b := &devdata.BriefResult{Prog: "p", Env: "e"}
	out := renderOutput(t, func() { printBrief(b) })
	assertContains(t, out, "p（e）", "代码 0 行", "蓝图 0 页")
	assertNotContains(t, out, "页签", "客制逻辑", "风险", "切片脱版")
}

// TestPrintFind 覆盖命中/未命中的两种排版。
func TestPrintFind(t *testing.T) {
	t.Run("命中与孤儿", func(t *testing.T) {
		f := &devdata.FindResult{
			Intent: "check", Field: "刻号",
			Intents: []devdata.IntentHit{{
				Intent:    devdata.Intent{Key: "check", Hint: "栏位检核改 input.a"},
				Points:    2,
				Templates: []devdata.SliceUse{{Code: "a05", Count: 1}},
			}},
			Points: []devdata.LinkPoint{{
				Name: "input.a.page73.xmdduc001", Line0: 17, Line1: 19, DeclLines: 3,
				Custom: true, Preview: "IF xmdduc001 = \"\"",
			}},
			Labels:  map[string]string{"xmdduc001": "刻号"},
			Orphans: []devdata.LinkOrphan{{Name: "input.a.page73.gone001", Lines: 2, Preview: "LET x"}},
			Total:   1,
			Note:    "另有 1 个孤儿声明",
		}
		out := renderOutput(t, func() { printFind(f) })
		assertContains(t, out,
			"意图定位：check  字段：刻号",
			"check", "2 个点", "参考切片 a05×1", "栏位检核改 input.a",
			"命中 1 个（列 1）",
			"（刻号）", "客制", "3 行",
			"孤儿声明", "input.a.page73.gone001",
			"另有 1 个孤儿声明",
		)
	})

	t.Run("没有命中且给意图清单", func(t *testing.T) {
		f := &devdata.FindResult{Note: "没有命中：字段可以给列名"}
		out := renderOutput(t, func() { printFind(f) })
		assertContains(t, out, "意图定位", "没有命中。", "可用意图：", "check", "lookup")
	})

	t.Run("没有命中但给了字段就不列清单", func(t *testing.T) {
		f := &devdata.FindResult{Field: "nope", Note: "没有命中"}
		out := renderOutput(t, func() { printFind(f) })
		assertContains(t, out, "字段：nope", "没有命中")
		assertNotContains(t, out, "可用意图")
	})
}

// TestPrintAt 覆盖行号反查的三条结论分支（点内 / section 内 / 都不在）。
func TestPrintAt(t *testing.T) {
	t.Run("在点内", func(t *testing.T) {
		a := &devdata.AtResult{
			Line: 17, TotalLine: 30, ContentStart: 15,
			Point:     &devdata.LinkPoint{Name: "input.a.page73.xmdduc001", Line0: 17, Line1: 19, Custom: true},
			Section:   &devdata.LinkSection{ID: "axmt500_wf.input"},
			Content:   []string{"before", "target", "after"},
			StaleNote: "注意：.4gl 比 .tap 旧",
			Verdict:   "在客制加点内",
		}
		out := renderOutput(t, func() { printAt(a) })
		assertContains(t, out,
			"L17 → input.a.page73.xmdduc001（客制 · L17-19）",
			"section axmt500_wf.input",
			"! 注意：.4gl 比 .tap 旧",
			"15: before", "17: after",
		)
	})

	t.Run("不在任何区域", func(t *testing.T) {
		a := &devdata.AtResult{Line: 2, TotalLine: 10, Verdict: "不在任何 section 内"}
		out := renderOutput(t, func() { printAt(a) })
		assertContains(t, out, "L2 → 不在任何 section 内", "! 这一行不在任何加点/section 内", "请改用 tt dev at <点名>")
	})

	t.Run("只在section且无内容起点", func(t *testing.T) {
		a := &devdata.AtResult{
			Line: 5, TotalLine: 10, Section: &devdata.LinkSection{ID: "s"},
			Content: []string{"x"},
		}
		out := renderOutput(t, func() { printAt(a) })
		assertContains(t, out, "L5 → ", "section s", "5: x")
	})
}

// TestPrintShow 覆盖"没有匹配点"与"点名正文"（含截断/孤儿/切片）。
func TestPrintShow(t *testing.T) {
	t.Run("没有匹配", func(t *testing.T) {
		s := &devdata.ShowResult{Query: "zzz", Note: "试试 tt dev fields"}
		out := renderOutput(t, func() { printShow(s) })
		assertContains(t, out, "没有匹配的点位：zzz", "试试 tt dev fields")
	})

	t.Run("有点名正文", func(t *testing.T) {
		s := &devdata.ShowResult{
			Query: "xmdduc001", Match: "suffix", Note: "按字段名后缀命中",
			Points: []devdata.ShownPoint{
				{
					LinkPoint: devdata.LinkPoint{
						Name: "input.a.page73.xmdduc001", Line0: 17, Line1: 18,
						Custom: true, Section: "axmt500_wf.input",
					},
					Label: "刻号", Content: []string{"l1", "l2"}, Truncated: true,
					SliceCodes: []string{"a05"},
				},
				{
					LinkPoint: devdata.LinkPoint{
						Name: "function.p_calc", Src: "c", Line0: 1, Line1: 1,
					},
					Orphan: true, Content: []string{"body"},
				},
			},
		}
		out := renderOutput(t, func() { printShow(s) })
		assertContains(t, out,
			"xmdduc001（匹配方式：suffix）", "按字段名后缀命中",
			"（刻号）", "客制", "切片 a05", "17: l1", "18: l2",
			"只打前 2 行，--max 调大",
			"孤儿（无槽位，重产不保留）", "正文在 .tap（.4gl 里没有）",
		)
	})
}

// TestPrintGrep 覆盖 search 的两段式输出与截断提示。
func TestPrintGrep(t *testing.T) {
	t.Run("两段都有命中", func(t *testing.T) {
		g := &devdata.GrepResult{
			Scope: "all", Pattern: "SET(", Total: 5, Shown: 2,
			Hits: []devdata.GrepHit{
				{Line: 3, Text: "  SET(x)", Kind: "客制", Section: "s1", SliceCode: "a05"},
				{Line: 4, Text: "SET(y)", Point: "input.a.page73.x1", Kind: "标准"},
			},
			TapTotal: 2, TapSource: "pkg",
			TapHits: []devdata.TapHit{
				{Line: 9, Text: "no slot", HasSlot: false},
				{Line: 10, Text: "slot", Point: "global.memo", Src: "c", HasSlot: true},
			},
			StaleNote: "产物比 tap 旧", Note: "命中落在声明里",
		}
		out := renderOutput(t, func() { printGrep(g) })
		assertContains(t, out,
			`检索 "SET("：产物命中 5 条（列 2）`, "；.tap 命中 2 条（列 2）",
			"! 产物比 tap 旧", "· 命中落在声明里",
			"（不在加点内）", "客制 · s1 · a05", "标准",
			"来源：包内 .tap（工作区最新真相）",
			"（不在任何声明里）", "**无槽位，重产不保留**", "有槽位",
			"只列前 2 条/范围：--limit 调大看全部",
		)
	})

	t.Run("只看tap范围", func(t *testing.T) {
		g := &devdata.GrepResult{
			Scope: "tap", Pattern: "x", TapTotal: 1,
			TapHits: []devdata.TapHit{{Line: 1, Text: "x", Src: "c", HasSlot: true}},
		}
		out := renderOutput(t, func() { printGrep(g) })
		assertContains(t, out, `.tap 命中 1 条（列 1）`, "来源：材料里的 .tap")
		assertNotContains(t, out, "产物命中")
	})

	t.Run("截断看总数", func(t *testing.T) {
		g := &devdata.GrepResult{
			Scope: "all", Pattern: "x", Total: 9, Shown: 1,
			Hits:     []devdata.GrepHit{{Line: 1, Text: "x"}},
			TapTotal: 3, TapHits: []devdata.TapHit{{Line: 2, Text: "x", Src: "s", HasSlot: true}},
		}
		out := renderOutput(t, func() { printGrep(g) })
		assertContains(t, out, "产物命中 9 条（列 1）", ".tap 命中 3 条（列 1）", "只列前 1 条/范围")
	})
}

// TestPrintPkgPoint 覆盖"包内 .tap 的这个点"两种身份与缺槽位提醒。
func TestPrintPkgPoint(t *testing.T) {
	t.Run("客制有槽位", func(t *testing.T) {
		v := &devdata.TapPointView{
			Name: "input.a.page73.xmdduc001", Src: "c",
			Line0: 10, Line1: 12, HasSlot: true, Content: []string{"L1", "L2"},
		}
		out := renderOutput(t, func() { printPkgPoint(v, `D:\ws\pkg\p(c).tzc`) })
		assertContains(t, out, "input.a.page73.xmdduc001", "客制", "包内 .tap L10-12", "p(c).tzc", "10: L1", "11: L2")
		assertNotContains(t, out, "重产不会保留")
	})

	t.Run("标准无槽位", func(t *testing.T) {
		v := &devdata.TapPointView{
			Name: "global.memo", Src: "s", Line0: 3, Line1: 3, Content: []string{"memo"},
		}
		out := renderOutput(t, func() { printPkgPoint(v, "p.tzc") })
		assertContains(t, out, "global.memo", "标准", "! 骨架里没有这个点的槽位：重产不会保留它")
	})
}

// TestOpenPayload 钉住 open 的 JSON 载荷口径：不吐 Manifest 这类大对象。
func TestOpenPayload(t *testing.T) {
	res := &taskws.CreateResult{
		Dir:  "D:/ws",
		Task: &taskws.Task{Prog: "p", Env: "e", Module: "CXM", Identity: "c", Task: "task1"},
		Fetch: &devdata.FetchResult{Files: []devdata.FileEntry{
			{Rel: "p.tab", Size: 10}, {Rel: "p.4gl", Size: 20}, {Rel: "p.tap", Size: 30},
		}},
		Index:    &taskws.LinkSummary{Points: 7, Custom: 3, Sections: 2},
		Pack:     &devdata.PackResult{Path: "pkg/p(c).tzc", Bytes: 100},
		Export:   &taskws.ExportResult{Dir: "ws", Regions: 5, Stats: store.Stats{Editable: 4}},
		Warnings: []string{"w"},
	}
	p := openPayload(res)
	if p.Dir != "D:/ws" || p.Prog != "p" || p.Env != "e" || p.Task != "task1" {
		t.Errorf("身份字段不对：%+v", p)
	}
	if p.Materials != 3 || p.Points != 7 || p.Custom != 3 || p.Sections != 2 {
		t.Errorf("计数不对：%+v", p)
	}
	if p.Pkg != "pkg/p(c).tzc" || p.PkgBytes != 100 || p.Ws != "ws" || p.Regions != 5 || p.Editable != 4 {
		t.Errorf("落点/区域不对：%+v", p)
	}
	if len(p.Warnings) != 1 {
		t.Errorf("warnings 应是 1 条：%+v", p.Warnings)
	}

	// 没有索引（--no-map 类退化路径）时点位数应保持 0，而不是 panic。
	empty := openPayload(&taskws.CreateResult{
		Dir: "D:/x", Task: &taskws.Task{Prog: "p", Env: "e"},
		Fetch: &devdata.FetchResult{}, Pack: &devdata.PackResult{}, Export: &taskws.ExportResult{},
	})
	if empty.Points != 0 || empty.Custom != 0 || empty.Sections != 0 {
		t.Errorf("没有索引时计数应为 0：%+v", empty)
	}
}

// TestRenderHelpers 是渲染层小工具的逐条用例：它们错了不会崩，只会**悄悄少一列信息**
// （体积、千分位、截断、缺省横线），而人读输出正是靠这些细节判断"结果可不可信"。
func TestRenderHelpers(t *testing.T) {
	// fmtBytes 的全部量级
	byteCases := map[int64]string{
		0:                  "0 B",
		512:                "512 B",
		1024:               "1.0 KB",
		1536:               "1.5 KB",
		1024 * 1024:        "1.0 MB",
		1024 * 1024 * 1024: "1.0 GB",
	}
	for in, want := range byteCases {
		if got := fmtBytes(in); got != want {
			t.Errorf("fmtBytes(%d) = %q，期望 %q", in, got, want)
		}
	}

	// fmtCount 千分位
	if got := fmtCount(999); got != "999" {
		t.Errorf("fmtCount(999) = %q", got)
	}
	if got := fmtCount(1000); got != "1,000" {
		t.Errorf("fmtCount(1000) = %q", got)
	}
	if got := fmtCount(1234567); got != "1,234,567" {
		t.Errorf("fmtCount(1234567) = %q", got)
	}

	// clip：按 rune 截断（中文不能被切坏成半个字节）
	if got := clip("  短  ", 10); got != "短" {
		t.Errorf("clip 短串应只去空白：%q", got)
	}
	if got := clip("一二三四五", 2); got != "一二…" {
		t.Errorf("clip 长串 = %q，期望 一二…", got)
	}

	// orDash / missingText
	if orDash("") != "-" || orDash("  ") != "-" || orDash("x") != "x" {
		t.Errorf("orDash 口径不对：%q %q %q", orDash(""), orDash("  "), orDash("x"))
	}
	if got := missingText(nil); got != "" {
		t.Errorf("没有缺失时不该有文案：%q", got)
	}
	if got := missingText([]string{"form", "design"}); !strings.Contains(got, "缺 2 项") || !strings.Contains(got, "form, design") {
		t.Errorf("missingText = %q", got)
	}

	// filepathBase：Windows/Unix 两种分隔符 + 空 + 纯文件名
	if filepathBase("") != "-" || filepathBase("a/b/c.tzc") != "c.tzc" || filepathBase(`a\b\c.tzc`) != "c.tzc" || filepathBase("plain") != "plain" {
		t.Errorf("filepathBase 口径不对")
	}

	// labelFor：按列名后缀命中；空列与不命中要返回空
	labels := map[string]string{"xmdduc001": "刻号", "": "空列"}
	if got := labelFor(labels, "input.a.page73.xmdduc001"); got != "刻号" {
		t.Errorf("labelFor 后缀命中 = %q", got)
	}
	if got := labelFor(labels, "input.a.page73.other"); got != "" {
		t.Errorf("labelFor 不该命中别的列：%q", got)
	}

	// pointSpan：无行号 / 区间 / 单行
	if pointSpan(&devdata.LinkPoint{}) != "无行号" {
		t.Error("Line0=0 应报无行号")
	}
	if got := pointSpan(&devdata.LinkPoint{Line0: 10, Line1: 20}); got != "L10-20" {
		t.Errorf("pointSpan 区间 = %q", got)
	}
	if got := pointSpan(&devdata.LinkPoint{Line0: 10, Line1: 10}); got != "L10" {
		t.Errorf("pointSpan 单行 = %q", got)
	}

	// sectionOrDash
	if sectionOrDash("") != "?" || sectionOrDash("s1") != "section s1" {
		t.Error("sectionOrDash 口径不对")
	}

	// warnChecks：ok 不占篇幅；error 排在 warn 前，同级按 ID
	warns := warnChecks([]devdata.LinkCheck{
		{ID: "z", Level: "warn", Detail: "w"},
		{ID: "b", Level: "ok", Detail: "o"},
		{ID: "b2", Level: "error", Detail: "e2"},
		{ID: "a", Level: "error", Detail: "e1"},
	})
	if len(warns) != 3 {
		t.Fatalf("warnChecks 应过滤 ok：%+v", warns)
	}
	if warns[0].ID != "a" || warns[1].ID != "b2" || warns[2].ID != "z" {
		t.Errorf("warnChecks 排序不对：%+v", warns)
	}
	if got := warnChecks(nil); len(got) != 0 {
		t.Errorf("空输入应返回空：%+v", got)
	}

	// orphanWhy 三种 reason
	if !strings.Contains(orphanWhy("no-slot"), "骨架里没有对应的洞") {
		t.Error("no-slot 的翻译不对")
	}
	if !strings.Contains(orphanWhy("no-code"), ".4gl 里也找不到") {
		t.Error("no-code 的翻译不对")
	}
	if got := orphanWhy("其它"); got != "" {
		t.Errorf("未知 reason 应返回空：%q", got)
	}

	// containsString 忽略大小写
	if !containsString([]string{"XmddUc001"}, "xmdduc001") {
		t.Error("containsString 应忽略大小写")
	}
	if containsString([]string{"a"}, "b") {
		t.Error("containsString 不该误命中")
	}
}
