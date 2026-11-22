package devdata

// 画面文案（.tsd 的字段标签/页签标题、.4fd 的 record→页签）与索引的融合。
//
// 这三份文件里，只有 .4fd 同时含 `<Page name="page_13">` 与 `<Table name="s_detail73">`，
// 所以"客户说的第 12 个页签"能落到 page_id 的**唯一**依据是它。这里测试的就是这条链：
// .tab 的 record → .4fd 的 page → .tsd 的标题。

import (
	"strings"
	"testing"
)

func TestParseTsd_FieldLabelsAndPageTitles(t *testing.T) {
	tsd := `<spec prog="p" env="c">
  <strings>
    <sfield name="lbl_xmdduc001" text="料号" />
    <sfield name="lbl_xmdduc001_desc" text="品名" />
    <sfield name="lbl_xmdduc001" text="重复的应被忽略" />
    <sfield name="page_13" text="示例页签" />
    <sfield name="bpage_1" text="订单明细" />
    <sfield name="axmt500_wf.page_10" text="订单状况明细表" />
    <sfield name="btn_ok" text="确认" />
  </strings>
</spec>`
	spec := parseTsd([]byte(tsd))
	if got := spec.Fields["xmdduc001"]; got != "料号" {
		t.Errorf("lbl_xmdduc001 = %q，想要 料号（重复项取第一个）", got)
	}
	if got := spec.Fields["xmdduc001_desc"]; got != "品名" {
		t.Errorf("lbl_xmdduc001_desc = %q", got)
	}
	if _, ok := spec.Fields["btn_ok"]; ok {
		t.Error("不带 lbl_ 前缀的 sfield 不该进字段表")
	}
	if got := spec.Pages["page_13"]; got != "示例页签" {
		t.Errorf("page_13 = %q", got)
	}
	// bpage_* 与带程序名前缀的键都是页签：只认 page_ 前缀会漏掉 6 个明细页。
	if got := spec.Pages["bpage_1"]; got != "订单明细" {
		t.Errorf("bpage_1 = %q", got)
	}
	if got := spec.Pages["axmt500_wf.page_10"]; got != "订单状况明细表" {
		t.Errorf("axmt500_wf.page_10 = %q", got)
	}
	if _, ok := spec.Pages["page_1"]; ok {
		t.Error("不存在的页签不该凭空出现")
	}
}

func TestParse4fd_RecordToPageUsesTextNotName(t *testing.T) {
	// 真机的两个坑：备注页是 name="page_1" text="page_12"；主档页是
	// name="master_Folder_page" text="page_1"。键必须是 text。
	fd := `<ManagedForm name="ManagedForm">
  <Page name="master_Folder_page" text="page_1">
    <Table name="s_browse"/>
  </Page>
  <Page name="page_1" text="page_12">
    <Table name="s_detail12"/>
  </Page>
  <Page name="page_13" text="page_13">
    <Table name="s_detail73"/>
  </Page>
  <Page name="page_info" text="page_info">
    <Page name="inner" text="page_inner">
      <Table name="s_detail1"/>
    </Page>
  </Page>
</ManagedForm>`
	spec := parse4fd([]byte(fd))
	if got := spec.PageByRecord["s_browse"]; got != "page_1" {
		t.Errorf("主档页 → %q，想要 page_1（取 text 而不是控件 id）", got)
	}
	if got := spec.PageByRecord["s_detail12"]; got != "page_12" {
		t.Errorf("备注页 → %q，想要 page_12（name 是 page_1，取 name 会接成基本数据）", got)
	}
	if got := spec.PageByRecord["s_detail73"]; got != "page_13" {
		t.Errorf("EPI 页 → %q", got)
	}
	if got := spec.PageByRecord["s_detail1"]; got != "page_inner" {
		t.Errorf("嵌套 grid 应归到内层页: %q", got)
	}
}

func TestMap_FieldLabelAndPageTitleComeFromTsdAnd4fd(t *testing.T) {
	cfg := writeConfig(t, `{"name":"e","host":"h","port":22,"user":"u"}`)
	progDir := progDirOf(cfg, "axmt500_wf")

	tapStr := tapDocText("identity=\"c\"",
		mkTapDecl("input.a.page73.f1", "", "c", "2", "LET f1 = 1\n")+
			mkTapSection("p.input", "1", mkTapSlot("input.a.page73.f1", "過欄位")))
	g4 := []string{
		`{<section id="p.input" type="s" >}`,
		"   #add-point:過欄位 name=\"input.a.page73.f1\"",
		"   LET f1 = 1",
		"   #END add-point",
		`{</section>}`,
	}
	tabStr := tabDocText(`name="axmt500_wf" module="cxm" type="t01" page="1" cusver="0000"`,
		`<var id="body" page="1" page_id="73" record="s_detail73" value="f1,f2"/>`)
	tsd := `<spec><strings>
  <sfield name="lbl_f1" text="料号" />
  <sfield name="page_13" text="示例页签" />
</strings></spec>`
	fd := `<ManagedForm><Page name="page_13" text="page_13">
  <Table name="s_detail73"/>
</Page></ManagedForm>`

	writeProgFiles(t, progDir, map[string]string{
		"dzx/axmt500_wf.tap": tapStr,
		"src/axmt500_wf.4gl": strings.Join(g4, "\n"),
		"dzx/axmt500_wf.tab": tabStr,
		"dzx/axmt500_wf.tsd": tsd,
		"dzx/axmt500_wf.4fd": fd,
	})

	doc, err := Map(LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"})
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if doc.Sources.Tsd == nil || doc.Sources.Fd == nil {
		t.Fatalf("Sources 没记 tsd/4fd: %+v", doc.Sources)
	}
	if len(doc.Fields) != 1 || doc.Fields[0].Label != "料号" {
		t.Errorf("字段标签 = %+v", doc.Fields)
	}
	if len(doc.Pages) != 1 {
		t.Fatalf("页面 = %+v", doc.Pages)
	}
	pg := doc.Pages[0]
	if pg.PageName != "page_13" {
		t.Errorf("PageName = %q，想要 page_13（按 record 桥接，不能按页序猜）", pg.PageName)
	}
	if pg.Title != "示例页签" {
		t.Errorf("Title = %q，想要 示例页签（4fd 占位，真标题在 tsd）", pg.Title)
	}
}

func TestMap_WithoutTsdAnd4fdStillBuilds(t *testing.T) {
	cfg, _, _ := queryFixture(t)
	doc, err := Map(LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"})
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if doc.Sources.Tsd != nil || doc.Sources.Fd != nil {
		t.Errorf("没给 tsd/4fd 却记了来源: %+v", doc.Sources)
	}
	if len(doc.Fields) != 1 || doc.Fields[0].Label != "" {
		t.Errorf("没有标签时不该编一个: %+v", doc.Fields)
	}
	if doc.Pages[0].Title != "" || doc.Pages[0].PageName != "" {
		t.Errorf("没有 4fd 时不该有页签键/标题: %+v", doc.Pages[0])
	}
}
