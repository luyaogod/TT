package drawio

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"tt/internal/testkit"
)

func loadSource(t *testing.T) *Source {
	t.Helper()
	src, err := Load()
	if err != nil {
		t.Fatalf("加载形状源失败：%v", err)
	}
	return src
}

// composeJSON 用一段 spec 直接跑一次 compose。
func composeJSON(t *testing.T, specJSON string) *Result {
	t.Helper()
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("spec 解析失败：%v", err)
	}
	res, err := loadSource(t).Compose(spec)
	if err != nil {
		t.Fatalf("compose 失败：%v", err)
	}
	return res
}

// cellByID 取一个单元格的**开标签连同它的几何** —— 属性在开标签里，坐标尺寸在紧随
// 其后的 <mxGeometry/> 里，少了后半截断言就没东西可看。
func cellByID(t *testing.T, xmlStr, id string) string {
	t.Helper()
	re := regexp.MustCompile(`<mxCell id="` + regexp.QuoteMeta(id) + `"[^>]*>(?:<mxGeometry[^>]*/>)?`)
	m := re.FindString(xmlStr)
	if m == "" {
		t.Fatalf("产物里找不到单元格 %s", id)
	}
	return m
}

// TestComposeMatchesNodeGolden 是 P2 的核心判据：拿一份**真实的**全屏规格
// （67 个控件 / 881 个单元格，容器、页签、明细表、绝对定位全都用上）与
// **原 Node 实现**产出的 .drawio 逐字节比。
//
// 为什么是逐字节：这条链没有时钟也没有随机源，同一份 spec 必须产出同一份字节。
// 逐字节比语义比更严 —— 少一个字符、多一个引号都会被抓住，而那正是 drawio
// 打不开的典型形态。
//
// 判据值钱的地方在于前面那两条真被抓住过：`grid="1"` 里含子串 `id="1"`（自检误报
// id 重复）、选中的页签要保留原 id（我给它加了兄弟后缀）。两条都是这条测试发现的。
func TestComposeMatchesNodeGolden(t *testing.T) {
	root := testkit.RepoRoot(t)
	dir := filepath.Join(root, "testdata", "drawio")

	specB, err := os.ReadFile(filepath.Join(dir, "ainq120-spec.json"))
	if err != nil {
		t.Fatalf("读不到规格：%v", err)
	}
	spec, err := ParseSpec(specB)
	if err != nil {
		t.Fatalf("spec 解析失败：%v", err)
	}
	res, err := loadSource(t).Compose(spec)
	if err != nil {
		t.Fatalf("compose 失败：%v", err)
	}

	wantB, err := os.ReadFile(filepath.Join(dir, "ainq120.drawio"))
	if err != nil {
		t.Fatalf("读不到基线：%v", err)
	}
	want := string(wantB)
	if res.XML != want {
		t.Errorf("与 Node 版基线不一致（第一处不同在字符 %d）：\n  Go  : …%s…\n  Node: …%s…",
			firstDiff(res.XML, want),
			snippet(res.XML, firstDiff(res.XML, want)),
			snippet(want, firstDiff(res.XML, want)))
	}
	if res.Count != 67 || res.Cells != 881 {
		t.Errorf("规模对不上：%d 个控件 / %d 个单元格，基线是 67 / 881", res.Count, res.Cells)
	}
}

// TestComposeSimpleGrid 用文档里那份样例（两列栅格）钉住最常用的一条路。
func TestComposeSimpleGrid(t *testing.T) {
	root := testkit.RepoRoot(t)
	b, err := os.ReadFile(filepath.Join(root, "testdata", "drawio", "purchase-order-form.json"))
	if err != nil {
		t.Fatalf("读不到样例：%v", err)
	}
	spec, err := ParseSpec(b)
	if err != nil {
		t.Fatalf("spec 解析失败：%v", err)
	}
	res, err := loadSource(t).Compose(spec)
	if err != nil {
		t.Fatalf("compose 失败：%v", err)
	}

	if res.Count != 9 {
		t.Errorf("样例该摆 9 个控件，得到 %d", res.Count)
	}
	// originX=60、colGap=24、列宽 209 → 第二列起点 60+209+24=293
	for _, want := range []string{"col0 x=60 w=209", "col1 x=293 w=209"} {
		if !strings.Contains(res.Report, want) {
			t.Errorf("列宽报告里缺 %q：\n%s", want, res.Report)
		}
	}
	if err := wellFormed(res.XML); err != nil {
		t.Errorf("产物不是良构 XML：%v", err)
	}
	if _, err := selfCheck(res.XML); err != nil {
		t.Errorf("自查没过：%v", err)
	}
}

// TestComposeStretchClasses 钉住三类子格行为 —— 这一条最容易出错，因为它全靠
// 形状源里那两个 1 基的部件序号（pfill / pdeco）。
//
// 拿 Folder 当样本（400×244，两个部件）：拉成两倍后，
//   p1 是 containerBox（fill）→ 框体跟着撑开，但要**减去另三边的内缩**（这里是 0，因为
//      它本来就从 y=24 贴到 y=244）
//   p2 是 tab（deco）→ 页签是装饰件，**一个像素都不许动**
func TestComposeStretchClasses(t *testing.T) {
	res := composeJSON(t, `{
	  "title": "拉伸",
	  "items": [ { "shape": "Folder", "x": 0, "y": 0, "w": 800, "h": 488 } ]
	}`)

	outer := cellByID(t, res.XML, "i0-g")
	if !strings.Contains(outer, `width="800" height="488"`) {
		t.Errorf("外层没被拉成 800×488：%s", outer)
	}
	box := cellByID(t, res.XML, "i0-p1")
	if !strings.Contains(box, `width="800" height="464"`) {
		t.Errorf("fill 的框体该填满外层（800×464）：%s", box)
	}
	tab := cellByID(t, res.XML, "i0-p2")
	if !strings.Contains(tab, `width="100" height="24"`) {
		t.Errorf("deco 的页签该保持原尺寸 100×24：%s", tab)
	}
}

// TestComposeDynamicTable 钉住「在 spec 里直接定义任意列行」。
func TestComposeDynamicTable(t *testing.T) {
	res := composeJSON(t, `{
	  "title": "明细",
	  "items": [ { "shape": "Table", "x": 10, "y": 20,
	    "table": {
	      "columns": [ { "title": "项次", "w": 50 }, { "title": "品号", "w": 120 } ],
	      "rows": 2
	    } } ]
	}`)

	// 尺寸 = 列宽和 (50+120) × (表头 26 + 2 行 × 26) = 170 × 78
	outer := cellByID(t, res.XML, "i0-t")
	if !strings.Contains(outer, `width="170" height="78"`) {
		t.Errorf("表格尺寸该由列宽与行数算出来（170×78）：%s", outer)
	}
	// 表头文字来自 columns[].title；数据行是空行
	for id, want := range map[string]string{
		"i0-hc0": `value="项次"`, "i0-hc1": `value="品号"`,
		"i0-r0c0": `value=""`, "i0-r1c1": `value=""`,
	} {
		if got := cellByID(t, res.XML, id); !strings.Contains(got, want) {
			t.Errorf("%s 该是 %s：%s", id, want, got)
		}
	}
}

// TestComposeMultiTab 钉住多页签画法，重点是**当前页那个格保留原来的 id**。
//
// 库里那个部件本来就叫 `p{part}`，当前页是把它**原地换掉**；其余页是新增的兄弟格
// `p{part}_{i}`。给当前页也加后缀的话，图上会同时出现两套页签格，而画布上看不出
// 差别 —— 只有逐字节对照才抓得住。
func TestComposeMultiTab(t *testing.T) {
	res := composeJSON(t, `{
	  "title": "多页签",
	  "items": [ { "shape": "Folder", "x": 0, "y": 0, "w": 600, "h": 244,
	               "pages": ["甲", "乙", "丙"], "active": 1 } ]
	}`)

	active := cellByID(t, res.XML, "i0-p2")
	if !strings.Contains(active, `value="乙"`) {
		t.Errorf("当前页该是「乙」：%s", active)
	}
	if !strings.Contains(active, "fillColor=#ffffff") {
		t.Errorf("当前页该用高亮白底样式：%s", active)
	}
	for i, want := range map[int]string{0: "甲", 2: "丙"} {
		got := cellByID(t, res.XML, "i0-p2_"+strconv.Itoa(i))
		if !strings.Contains(got, `value="`+want+`"`) {
			t.Errorf("兄弟页签 p2_%d 该是「%s」：%s", i, want, got)
		}
		if !strings.Contains(got, "fillColor=#ececec") {
			t.Errorf("兄弟页签该用未选中的灰底样式：%s", got)
		}
	}
	if _, err := selfCheck(res.XML); err != nil {
		t.Errorf("自查没过：%v", err)
	}
}

// TestComposeErrors 钉住几条"输入错"的报错 —— 报错里要带上调用方下一步需要的
// 东西（可用 id / 可用槽位），而不是只喊一句找不到。
func TestComposeErrors(t *testing.T) {
	cases := []struct {
		name, spec, want string
	}{
		{"未知控件", `{"items":[{"shape":"Nope","col":0,"row":0}]}`, "找不到控件"},
		{"未知槽位", `{"items":[{"shape":"Label","col":0,"row":0,"text":{"nope":"x"}}]}`, "没有文字槽位"},
		{"非 Folder 给 pages", `{"items":[{"shape":"Label","col":0,"row":0,"pages":["a"]}]}`, "不支持 pages 参数"},
		{"非 Table 给 table", `{"items":[{"shape":"Label","col":0,"row":0,"table":{"columns":[{"title":"a","w":10}]}}]}`, "不支持 table 参数"},
		{"active 越界", `{"items":[{"shape":"Folder","col":0,"row":0,"pages":["a"],"active":3}]}`, "active 必须是"},
		{"空 items", `{"items":[]}`, "必须是非空数组"},
		{"负数 col", `{"items":[{"shape":"Label","col":-1}]}`, "不能是负数"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			spec, err := ParseSpec([]byte(c.spec))
			if err == nil {
				_, err = loadSource(t).Compose(spec)
			}
			if err == nil {
				t.Fatalf("该报错，却成功了")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("报错该含 %q，得到：%v", c.want, err)
			}
		})
	}
}

// TestSelfCheckCatches 直接喂坏产物给自查 —— 三条判据各来一次。
//
// 这几条**不是**可达的输入路径（compose 自己不会产出这种 XML）；测的是"自查真的会响"，
// 否则整条防线就是一句空话。没有这条，上面那些"自查没过就退 3"全都没被验过。
func TestSelfCheckCatches(t *testing.T) {
	base := `<mxfile><diagram id="t"><mxGraphModel><root>` +
		`<mxCell id="0"/><mxCell id="1" parent="0"/>%s</root></mxGraphModel></diagram></mxfile>`

	cases := []struct {
		name, cells, want string
	}{
		{"id 重复", `<mxCell id="a"/><mxCell id="a"/>`, "id 重复"},
		{"parent 悬空", `<mxCell id="a" parent="nope"/>`, "悬空 parent"},
		{"几何里有 NaN", `<mxCell id="a"><mxGeometry width="NaN"/></mxCell>`, "NaN"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := selfCheck(strings.Replace(base, "%s", c.cells, 1))
			if err == nil {
				t.Fatalf("该被自查拦下")
			}
			var sc *SelfCheckError
			if !asSelfCheck(err, &sc) {
				t.Errorf("应当是 *SelfCheckError（命令层靠它退 3），得到 %T", err)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("报错该含 %q，得到：%v", c.want, err)
			}
		})
	}

	// 反面：`grid="1"` 里的 `id="1"` 是子串，不许被当成单元格 id（踩过一次）
	ok := `<mxfile><diagram id="t"><mxGraphModel grid="1" page="1"><root>` +
		`<mxCell id="0"/><mxCell id="1" parent="0"/><mxCell id="a" parent="1"/></root></mxGraphModel></diagram></mxfile>`
	if n, err := selfCheck(ok); err != nil {
		t.Errorf("这份是好的，不该报错（%v）", err)
	} else if n != 3 {
		t.Errorf("该数出 3 个单元格，得到 %d —— 属性里的 id= 被误算了", n)
	}
}

func asSelfCheck(err error, out **SelfCheckError) bool {
	if sc, ok := err.(*SelfCheckError); ok {
		*out = sc
		return true
	}
	return false
}

// firstDiff 返回第一处不同的字符下标（全等时返回 -1）。
func firstDiff(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	if len(a) != len(b) {
		return n
	}
	return -1
}

func snippet(s string, at int) string {
	if at < 0 {
		return "（两份相同？）"
	}
	lo, hi := at-60, at+60
	if lo < 0 {
		lo = 0
	}
	if hi > len(s) {
		hi = len(s)
	}
	return s[lo:hi]
}
