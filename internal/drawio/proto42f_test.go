package drawio

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"tt/internal/testkit"
)

// `.42f` → 原型数据（proto42f.go）的判据。
//
// 这一层要证明的事只有一条：**图上的几何来自 `.42f` 的真值，不是估的**。
// 所以最有价值的不是"跑通了"，而是第 2 条测试（`TestProto42F_ZeroedGeometryIsRed`）：
// 把夹具里的坐标全改成 0，第 1 条的那些断言必须红 —— 否则它们只是在描述自己。

// basic42F 读合成夹具 `internal/fep/testdata/form42f/basic.xml`。
//
// 就地复用 fep 的夹具，不复制第二份：一个事实只有一个家，`.42f` 的解析口径由
// internal/fep 守，这里只消费它的产物。
func basic42F(t *testing.T) (string, string) {
	t.Helper()
	p := filepath.Join(testkit.RepoRoot(t), "internal", "fep", "testdata", "form42f", "basic.xml")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("读不到夹具 %s：%v", p, err)
	}
	return "basic", string(b)
}

// cellGeomRe 在产物里找「带 x/y/width/height 的单元格」—— 位置断言全靠它。
var cellGeomRe = regexp.MustCompile(
	`<mxCell id="([^"]+)"[^>]*>\s*<mxGeometry x="(-?[\d.]+)" y="(-?[\d.]+)" width="(-?[\d.]+)" height="(-?[\d.]+)"`)

// cellGeomIn 取某个单元格的几何；找不到返回 ok=false（不用 t.Fatalf ——
// 反证测试要靠返回值变红，不能靠 fatal）。
func cellGeomIn(xmlStr, id string) (x, y, w, h float64, ok bool) {
	for _, m := range cellGeomRe.FindAllStringSubmatch(xmlStr, -1) {
		if m[1] != id {
			continue
		}
		return atof(m[2]), atof(m[3]), atof(m[4]), atof(m[5]), true
	}
	return 0, 0, 0, 0, false
}

// assertBasicRealGeometry 是第 1 条与反证测试**共用**的断言。
//
// 抽成纯函数（返回问题列表而不是 t.Fatalf）是反证的关键：同一段判据喂给
// "真几何" 与 "全 0 占位"，一个必须空、一个必须非空。
func assertBasicRealGeometry(p *Proto) []string {
	var probs []string
	add := func(format string, a ...any) { probs = append(probs, fmt.Sprintf(format, a...)) }

	byName := map[string]*ProtoField{}
	for i := range p.Fields {
		byName[p.Fields[i].ColName] = &p.Fields[i]
	}
	if len(p.Fields) != 3 {
		add("夹具该有 3 条字段（2 FormField + 1 TableColumn），得到 %d", len(p.Fields))
	}

	cases := []struct {
		col                      string
		posX, posY, width, gridW int
		shape                    string
	}{
		{"demo001", 6, 0, 20, 20, "Edit"},
		{"demo002", 6, 1, 14, 14, "CheckBox"},
		{"demo003", 7, 3, 8, 0, "Edit"},
	}
	for _, c := range cases {
		f := byName[c.col]
		if f == nil {
			add("产物里缺字段 %s", c.col)
			continue
		}
		if f.PosX != c.posX || f.PosY != c.posY {
			add("%s 的 posX/posY 该是 %d/%d，得到 %d/%d", c.col, c.posX, c.posY, f.PosX, f.PosY)
		}
		if f.Width != c.width {
			add("%s 的 width 该是 %d，得到 %d", c.col, c.width, f.Width)
		}
		if f.GridWidth != c.gridW {
			add("%s 的 gridWidth 该是 %d，得到 %d", c.col, c.gridW, f.GridWidth)
		}
		if f.Shape != c.shape {
			add("%s 该画成 %s，得到 %q", c.col, c.shape, f.Shape)
		}
		if !f.Geometry {
			add("%s 在文件里有网格坐标，却没被当成真实几何", c.col)
		}

		wantX := float64(c.posX) * p.CellPx
		wantY := float64(c.posY) * p.CellPx
		wantW := float64(c.width) * p.CellPx
		if f.X != wantX || f.Y != wantY {
			add("%s 画在 (%s,%s)，该是 (%s,%s)", c.col, jsNum(f.X), jsNum(f.Y), jsNum(wantX), jsNum(wantY))
		}
		if f.W != wantW {
			add("%s 画成宽 %s，该是 %s（width %d 网格 × CellPx）", c.col, jsNum(f.W), jsNum(wantW), c.width)
		}
		// 真实几何既不是 0，也不是形状库的默认尺寸 —— 那两种都是"占位"。
		if gx, gy, gw, _, ok := cellGeomIn(p.XML, f.CellID); !ok {
			add("%s 的单元格 %q 在产物里找不到", c.col, f.CellID)
		} else {
			if gx != f.X || gy != f.Y || gw != f.W {
				add("%s 的单元格几何 (%s,%s,%s) 与字段记录 (%s,%s,%s) 不一致",
					c.col, jsNum(gx), jsNum(gy), jsNum(gw), jsNum(f.X), jsNum(f.Y), jsNum(f.W))
			}
			if c.posX != 0 && gx == 0 {
				add("%s 的 posX=%d，却画在 x=0 —— 真实几何被当成占位了", c.col, c.posX)
			}
		}
	}

	// RecordView → tabName 必须能看见（本层的第三条交付）。
	if len(p.Tabs) != 1 {
		add("夹具里有 1 个 RecordView，产物里是 %d 个", len(p.Tabs))
	} else {
		if p.Tabs[0].TabName != "demo_t" {
			add("tabName 该是 demo_t，得到 %q", p.Tabs[0].TabName)
		}
		if got := strings.Join(p.Tabs[0].ColNames, ","); got != "demo003" {
			add("RecordView demo_t 的 colName 该是 demo003，得到 %q", got)
		}
	}
	if !strings.Contains(p.Legend, "demo_t") || !strings.Contains(p.Legend, "demo003") {
		add("图例里没写出 tabName → colName 的关系：\n%s", p.Legend)
	}
	return probs
}

// TestProto42F_RealGeometryFrom42F 钉住「形状带的是 `.42f` 的真几何」。
func TestProto42F_RealGeometryFrom42F(t *testing.T) {
	name, text := basic42F(t)
	p, err := ProtoFrom42F(name, text, nil)
	if err != nil {
		t.Fatalf("生成原型失败：%v", err)
	}
	if probs := assertBasicRealGeometry(p); len(probs) > 0 {
		t.Fatalf("真实几何断言没过（%d 条）：\n  %s", len(probs), strings.Join(probs, "\n  "))
	}
	if err := wellFormed(p.XML); err != nil {
		t.Errorf("产物不是良构 XML：%v", err)
	}
	if _, err := selfCheck(p.XML); err != nil {
		t.Errorf("产物自查没过：%v", err)
	}
	if p.Form != "basic" {
		t.Errorf("表单名该是 basic，得到 %q", p.Form)
	}
}

// TestProto42F_ZeroedGeometryIsRed 是反证：**把真实几何换成全 0 占位，第 1 条必须红**。
//
// 做法不是"手搓一份假 Proto"（那证明不了第 1 条走的真是解析结果），而是把夹具文本里的
// posX/posY/width/gridWidth 全改成 0，再跑同一条流水线、喂同一个断言函数。
// 断言仍然通过 ⇒ 第 1 条测的是空气。
func TestProto42F_ZeroedGeometryIsRed(t *testing.T) {
	name, text := basic42F(t)
	zeroed := zeroGeometry42F(text)
	if zeroed == text {
		t.Fatal("反证夹具没生效：一字未改")
	}

	p, err := ProtoFrom42F(name, zeroed, nil)
	if err != nil {
		t.Fatalf("零几何也必须能出图（落到推断排布带），不该报错：%v", err)
	}
	probs := assertBasicRealGeometry(p)
	if len(probs) == 0 {
		t.Fatal("把真实几何换成全 0 之后断言仍然全过 —— 第 1 条测试是假的")
	}
	t.Logf("反证按预期变红（%d 条），首条：%s", len(probs), probs[0])
}

// zeroGeometry42F 把 .42f 文本里的几何属性全部归零。
func zeroGeometry42F(text string) string {
	for _, attr := range []string{"posX", "posY", "width", "height", "gridWidth"} {
		re := regexp.MustCompile(`\b` + attr + `="[^"]*"`)
		text = re.ReplaceAllString(text, attr+`="0"`)
	}
	return text
}

// TestProto42F_RealSample 是真样本门控：`axmt500_wf.42f` 的 563 条字段必须全部进原型，
// 且若干已知字段的几何与文件一致。
//
// 样本不在就跳过并写清原因 —— 跳过不是"通过"，日志里能看见。
func TestProto42F_RealSample(t *testing.T) {
	dir := os.Getenv("TT_FORM_SAMPLE_DIR")
	if dir == "" {
		dir = filepath.Join(testkit.RepoRoot(t), ".agents",
			"Genero Original Factory Compatible", "fixtures", "forms")
	}
	path := filepath.Join(dir, "t100", "axmt500_wf.42f")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("真样本不在 %s（用 TT_FORM_SAMPLE_DIR 覆盖）；跳过：%v", path, err)
	}

	p, err := ProtoFrom42FFile(path, nil)
	if err != nil {
		t.Fatalf("解析真样本失败：%v", err)
	}
	if len(p.Fields) != 563 {
		t.Fatalf("该有 563 条字段（TableColumn 397 + FormField 132 + PhantomFormField 22 + PhantomColumn 12），得到 %d", len(p.Fields))
	}
	if p.Count != 563 {
		t.Errorf("进原型的字段数该是 563，得到 %d", p.Count)
	}

	// 每条字段都得在产物里有一个带几何的单元格 —— 掉一条就是"看起来对、其实少"。
	ids := map[string]bool{}
	for _, m := range cellIDRe.FindAllStringSubmatch(p.XML, -1) {
		ids[m[1]] = true
	}
	for i := range p.Fields {
		f := &p.Fields[i]
		if f.CellID == "" {
			t.Fatalf("字段 %s 没有单元格 id", f.ColName)
		}
		if !ids[f.CellID] {
			t.Fatalf("字段 %s 的单元格 %q 不在产物里", f.ColName, f.CellID)
		}
	}

	byName := map[string]*ProtoField{}
	for i := range p.Fields {
		byName[p.Fields[i].ColName] = &p.Fields[i]
	}

	// 已知字段：几何来自文件原文（见 axmt500_wf.42f 的 <TextEdit …/> 一行）。
	help := byName["help"]
	if help == nil {
		t.Fatal("缺 help 字段")
	}
	if help.Widget != "TextEdit" || help.Shape != "TextEdit" {
		t.Errorf("help 该是 TextEdit，得到 widget=%q shape=%q", help.Widget, help.Shape)
	}
	if help.PosX != 1 || help.PosY != 0 || help.Width != 22 || help.Height != 5 || help.GridWidth != 22 {
		t.Errorf("help 的几何与文件不一致：posX=%d posY=%d width=%d height=%d gridWidth=%d"+
			"（文件是 posX=1 posY=0 width=22 height=5 gridWidth=22）",
			help.PosX, help.PosY, help.Width, help.Height, help.GridWidth)
	}
	if gx, _, gw, _, ok := cellGeomIn(p.XML, help.CellID); !ok || gx != 10 || gw != 220 {
		t.Errorf("help 该画在 x=10 宽 220（1/22 网格 × %s），得到 x=%s 宽=%s ok=%v",
			jsNum(p.CellPx), jsNum(gx), jsNum(gw), ok)
	}

	// 必填/只读在真样本里都能被认出来。
	docNo := byName["b_xmdadocno"]
	if docNo == nil {
		t.Fatal("缺 b_xmdadocno 字段")
	}
	if !docNo.Required || !docNo.NotNull {
		t.Errorf("b_xmdadocno 在文件里 required=1 notNull=1，字段上是 required=%v notNull=%v",
			docNo.Required, docNo.NotNull)
	}
	if docNo.Width != 20 || docNo.Shape != "Edit" {
		t.Errorf("b_xmdadocno 该是 width=20 的 Edit，得到 width=%d shape=%q", docNo.Width, docNo.Shape)
	}
	pic := byName["b_statepic"]
	if pic == nil {
		t.Fatal("缺 b_statepic 字段")
	}
	if !pic.NoEntry {
		t.Error("b_statepic 在文件里 noEntry=1，字段上却是 false")
	}
	if pic.Mapped {
		t.Errorf("b_statepic 的控件是 <Image>，形状库里没有同名图形 —— 该判未判定，而不是画成 %q", pic.Shape)
	}

	// 明细列（TableColumn）也在里面。
	tableCols := 0
	for i := range p.Fields {
		if p.Fields[i].Tag == "TableColumn" {
			tableCols++
		}
	}
	if tableCols != 397 {
		t.Errorf("TableColumn 该有 397 条，得到 %d", tableCols)
	}

	// 未判定清单与图例。
	if len(p.Unmapped) == 0 {
		t.Error("真样本里有 Image / Completer / PhantomEdit / DateTimeEdit 这些库里没有的控件，未判定清单不该是空的")
	}
	if !strings.Contains(p.Legend, "未判定") {
		t.Errorf("图例里该点名「未判定」：\n%s", p.Legend)
	}
	seenTab := false
	for _, tab := range p.Tabs {
		if tab.TabName == "" {
			continue
		}
		seenTab = true
		if !strings.Contains(p.Legend, tab.TabName) {
			t.Errorf("图例里缺 tabName %q", tab.TabName)
		}
	}
	if !seenTab {
		t.Error("真样本里 8 个 RecordView 都该有 tabName")
	}
}

// TestProto42F_UnmappedWidgetGetsGenericRect 钉住 A15：控件族认不出来时**不猜**，
// 画通用矩形 + 标「未判定」，而且**不许失败**。
func TestProto42F_UnmappedWidgetGetsGenericRect(t *testing.T) {
	const form = `<?xml version="1.0" encoding="ASCII"?>
<Form name="wx" width="40" height="10">
  <Grid width="40" height="10">
    <FormField name="formonly.w001" colName="w001" fieldId="0">
      <Wacky width="10" posX="3" posY="2"/>
    </FormField>
    <FormField name="formonly.w002" colName="w002" fieldId="1">
      <Edit width="10" posX="14" posY="2"/>
    </FormField>
    <FormField name="formonly.w003" colName="w003" fieldId="2"/>
  </Grid>
</Form>`

	p, err := ProtoFrom42F("wx", form, nil)
	if err != nil {
		t.Fatalf("认不出的控件不许让整份原型失败：%v", err)
	}
	byName := map[string]*ProtoField{}
	for i := range p.Fields {
		byName[p.Fields[i].ColName] = &p.Fields[i]
	}
	if f := byName["w001"]; f == nil || f.Mapped || f.Shape != "" {
		t.Fatalf("w001 的控件 <Wacky> 形状库里没有，该判未判定（mapped=false，shape 空），得到 %+v", f)
	}
	if f := byName["w003"]; f == nil || f.Mapped {
		t.Fatalf("w003 连控件元素都没有，也该判未判定，得到 %+v", f)
	}
	if f := byName["w002"]; f == nil || !f.Mapped || f.Shape != "Edit" {
		t.Fatalf("w002 是 Edit，该照常映射，得到 %+v", f)
	}

	if !strings.Contains(p.XML, "未判定") {
		t.Error("产物里该有标着「未判定」的通用矩形")
	}
	if !strings.Contains(p.XML, `dashed=1`) {
		t.Error("通用矩形该用虚线样式（与映射成功的形状区分开）")
	}
	if f := byName["w001"]; f != nil {
		if _, _, gw, _, ok := cellGeomIn(p.XML, f.CellID); !ok || gw != 100 {
			t.Errorf("w001 的通用矩形该按 width=10 网格画成宽 100，cell=%q ok=%v", f.CellID, ok)
		}
	}
	for _, w := range p.Unmapped {
		if w == "Wacky" {
			return
		}
	}
	t.Errorf("未判定清单里该有 Wacky，得到 %v", p.Unmapped)
}

// TestProto42F_RequiredAndNoEntryVisible 钉住第二条交付：必填与只读在图上可区分，
// 且图例解释了这些记号。
func TestProto42F_RequiredAndNoEntryVisible(t *testing.T) {
	const form = `<?xml version="1.0" encoding="ASCII"?>
<Form name="mk" width="40" height="20">
  <Grid width="40" height="20">
    <FormField name="formonly.req" colName="req001" fieldId="0" required="1">
      <Edit width="10" posX="1" posY="0"/>
    </FormField>
    <FormField name="formonly.nn" colName="nn001" fieldId="1" notNull="1">
      <Edit width="10" posX="1" posY="1"/>
    </FormField>
    <FormField name="formonly.ro" colName="ro001" fieldId="2" noEntry="1">
      <Edit width="10" posX="1" posY="2"/>
    </FormField>
    <FormField name="formonly.pl" colName="plain" fieldId="3">
      <Edit width="10" posX="1" posY="3"/>
    </FormField>
  </Grid>
</Form>`

	p, err := ProtoFrom42F("mk", form, nil)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	byName := map[string]*ProtoField{}
	for i := range p.Fields {
		byName[p.Fields[i].ColName] = &p.Fields[i]
	}
	for _, col := range []string{"req001", "nn001", "ro001", "plain"} {
		if byName[col] == nil {
			t.Fatalf("缺字段 %s", col)
		}
	}
	// 必填：required 与 notNull 都得标（两者都能让用户卡住）。
	for _, col := range []string{"req001", "nn001"} {
		if !strings.HasPrefix(byName[col].Label, "*") {
			t.Errorf("%s 是必填，图上该带 * 前缀，得到 %q", col, byName[col].Label)
		}
	}
	if !strings.Contains(byName["ro001"].Label, "只读") {
		t.Errorf("ro001 是 noEntry，图上该标只读，得到 %q", byName["ro001"].Label)
	}
	if strings.HasPrefix(byName["plain"].Label, "*") || strings.Contains(byName["plain"].Label, "只读") {
		t.Errorf("plain 既不必填也不只读，不该被标，得到 %q", byName["plain"].Label)
	}
	// 差别必须落在产物里（不是只存在于结构体上）。
	for _, want := range []string{
		escapeAttr(byName["req001"].Label),
		escapeAttr(byName["ro001"].Label),
		escapeAttr(byName["plain"].Label),
	} {
		if !strings.Contains(p.XML, `value="`+want+`"`) {
			t.Errorf("产物里找不到字段文字 %q", want)
		}
	}
	if byName["req001"].Label == byName["plain"].Label {
		t.Error("必填与非必填画出来一模一样 —— 不可区分")
	}
	// 图例解释这两个记号。
	for _, want := range []string{"必填", "只读"} {
		if !strings.Contains(p.Legend, want) {
			t.Errorf("图例里该解释 %q：\n%s", want, p.Legend)
		}
	}
}

// TestProto42F_DataTextIsNotGeometry 钉住一条被真样本打出来的坑：
// 自检原来拿 `strings.Contains(out,"NaN"|"undefined")` 扫**整份产物**，而 `.42f` 里
// 存在 `<RecordView tabName="undefined">`（实测 axmt500_wf.42f 第 8 个 RecordView），
// 于是"数据里出现这个词"被当成"几何坏了"。自检现在只看几何属性 —— 这条测试守它。
func TestProto42F_DataTextIsNotGeometry(t *testing.T) {
	const form = `<?xml version="1.0" encoding="ASCII"?>
<Form name="odd" width="40" height="10">
  <Grid width="40" height="10">
    <FormField name="formonly.nan" colName="NaN" fieldId="0" comment="undefined">
      <Edit width="10" posX="1" posY="0"/>
    </FormField>
  </Grid>
  <RecordView tabName="undefined">
    <Link colName="NaN" fieldIdRef="0"/>
  </RecordView>
</Form>`

	p, err := ProtoFrom42F("odd", form, nil)
	if err != nil {
		t.Fatalf("数据里出现 NaN/undefined 这两个词不该被当成几何 bug：%v", err)
	}
	if !strings.Contains(p.Legend, "undefined") {
		t.Errorf("tabName 原文该原样进图例（不修饰、不吞掉）：\n%s", p.Legend)
	}
	if err := wellFormed(p.XML); err != nil {
		t.Errorf("产物不是良构 XML：%v", err)
	}
}
