package devdata

// 意图表与 brief：这两块的正确性不在"解析对不对"（那是 link_test 的事），
// 而在"给 AI 的下一步是不是真的能落地"。所以测试断言的是：
//   - 每条意图都能按 key 与中文别名取到；
//   - find 真的把点筛到该意图的段上，且字段后缀过滤不会漏/多；
//   - brief 的数字与同一份 LinkDoc 一致（不是另算一份）。

import (
	"strings"
	"testing"
)

func TestIntentByKey_KeyAndChineseAlias(t *testing.T) {
	for _, q := range []string{"check", "CHECK", "检核", "校验"} {
		got := IntentByKey(q)
		if len(got) != 1 || got[0].Key != "check" {
			t.Errorf("IntentByKey(%q) = %+v，想要 check", q, got)
		}
	}
	if got := IntentByKey("保存前"); len(got) != 1 || got[0].Key != "save" {
		t.Errorf("保存前 → %+v", got)
	}
	if got := IntentByKey("不存在的意图"); len(got) != 0 {
		t.Errorf("未知意图应为空: %+v", got)
	}
	if got := IntentByKey(""); len(got) != 0 {
		t.Errorf("空查询应为空: %+v", got)
	}
}

func TestIntentTable_KeysUniqueAndHintsPresent(t *testing.T) {
	seen := map[string]bool{}
	for _, in := range Intents {
		if in.Key == "" || in.Hint == "" {
			t.Errorf("意图缺 key 或 hint: %+v", in)
		}
		if seen[in.Key] {
			t.Errorf("意图 key 重复: %s", in.Key)
		}
		seen[in.Key] = true
		if len(in.Sections) == 0 && len(in.Slots) == 0 {
			t.Errorf("意图 %s 没有任何落点（find 永远命中不了）", in.Key)
		}
	}
}

func TestIntentMatches_HeadSectionAndSlots(t *testing.T) {
	check := IntentByKey("check")[0]
	if !intentMatches(check, LinkPoint{Name: "input.a.page73.xmdduc001"}) {
		t.Error("input.a.* 应落在 check")
	}
	if intentMatches(check, LinkPoint{Name: "ui_dialog.page1.action"}) {
		t.Error("ui_dialog 不该落在 check")
	}
	save := IntentByKey("save")[0]
	if !intentMatches(save, LinkPoint{Name: "insert_b.before_insert.x"}) {
		t.Error("insert_b.* 应落在 save")
	}
	if !intentMatches(save, LinkPoint{Name: "update_b.m_update10.before_update"}) {
		t.Error("before_update 槽位应落在 save")
	}
}

func TestFind_FiltersByIntentAndField(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	res, err := Find(FindRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Intent:      "检核", Field: "f1",
	})
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if len(res.Intents) != 1 || res.Intents[0].Key != "check" {
		t.Fatalf("意图 = %+v", res.Intents)
	}
	if res.Total != 1 || res.Points[0].Name != "input.a.page1.f1" {
		t.Fatalf("命中 = %+v", res.Points)
	}
	if res.Intents[0].Points != 1 {
		t.Errorf("意图内计数 = %d，想要 1", res.Intents[0].Points)
	}
	// 模板用量：夹具里 a05 出现 1 次，应出现在 hit.Templates 里
	if len(res.Intents[0].Templates) != 1 || res.Intents[0].Templates[0].Code != "a05" {
		t.Errorf("模板用量 = %+v", res.Intents[0].Templates)
	}
}

func TestFind_FieldOnlyDoesNotNeedIntent(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	res, err := Find(FindRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Field:       "f1",
	})
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if res.Total != 1 || len(res.Intents) != 0 {
		t.Errorf("只给字段：total=%d intents=%d", res.Total, len(res.Intents))
	}
	if !strings.Contains(res.Points[0].Name, ".f1") {
		t.Errorf("点位 = %s", res.Points[0].Name)
	}
}

func TestFind_UnknownIntentExplainsItself(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	res, err := Find(FindRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Intent:      "把报表打印出来",
	})
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if len(res.Intents) != 0 || res.Total != 0 {
		t.Errorf("未知意图不该命中: %+v", res)
	}
	if !strings.Contains(res.Note, "tt dev brief") {
		t.Errorf("提示没指向 brief: %q", res.Note)
	}
}

func TestFind_LimitKeepsTotal(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	res, err := Find(FindRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		Field:       "f1", Limit: 1,
	})
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if res.Total != 1 || len(res.Points) != 1 {
		t.Errorf("total=%d shown=%d", res.Total, len(res.Points))
	}
}

func TestBrief_ReportsPagesCustomAndIntentList(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	b, err := Brief(BriefRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
	})
	if err != nil {
		t.Fatalf("Brief: %v", err)
	}
	if b.Prog != "axmt500_wf" || b.Env == "" {
		t.Errorf("身份 = %q / %q", b.Prog, b.Env)
	}
	if b.Mode.Type != "t01" || !strings.EqualFold(b.Mode.Module, "cxm") {
		t.Errorf("判型 = %+v", b.Mode)
	}
	if len(b.Pages) != 1 || b.Pages[0].PageID != "1" || b.Pages[0].Fields != 1 {
		t.Errorf("页面 = %+v", b.Pages)
	}
	if b.CustomTotal != 1 || len(b.CustomPoints) != 1 {
		t.Errorf("客制点 = %d / %+v", b.CustomTotal, b.CustomPoints)
	}
	if b.CustomPoints[0].Name != "input.a.page1.f1" {
		t.Errorf("客制点 = %s", b.CustomPoints[0].Name)
	}
	if len(b.Intents) != len(Intents) {
		t.Errorf("意图清单 = %d，想要 %d", len(b.Intents), len(Intents))
	}
	if b.Stats.Points != 1 {
		t.Errorf("统计点位数 = %d，想要 1", b.Stats.Points)
	}
	if b.Link == "" {
		t.Error("brief 没带索引路径")
	}
}

func TestBrief_TopPointsTruncatesButKeepsTotal(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	b, err := Brief(BriefRequest{
		LinkRequest: LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"},
		TopPoints:   0, // 默认 15，夹具只有 1 个
	})
	if err != nil {
		t.Fatalf("Brief: %v", err)
	}
	if b.CustomTotal != 1 || len(b.CustomPoints) != 1 {
		t.Errorf("默认不该截断: %d/%d", b.CustomTotal, len(b.CustomPoints))
	}
}
