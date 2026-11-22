package devdata

// fields / show / at / grep 的**磁盘材料**矩阵。
//
// 为什么不用内存 LinkDoc：这四条命令的输入是"索引 + .4gl 原文 + .tap 原文"三份磁盘材料，
// 而它们最容易错的地方恰恰是**行号换算**与**材料来源**（mdx 的第几行、正文来自哪份文件）。
// 内存夹具没有行号，测不到"整段偏移一行"这种真机上最难发现的问题。
//
// 对应真机踩过的点：
//   - page_id 与页签序号是两回事（客户说"第 12 个页签"，代码里是 page73）；
//   - 画面中文名（刻号）与列名（xmdduc003）之间那层翻译；
//   - 孤儿点优先：`xmdducseq` 与 `xmdducseq1` 只差一个字符，答错就是两种结论；
//   - AtResult.ContentStart 少一个就整段行号偏移；
//   - grep 要标出归属（客制/标准/生成器自加）与切片，且 tap 侧命中要单独给。

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/pkgfile"
)

// lineOf 在材料文件里找包含 substr 的第一行（1 起）；找不到返回 0。
func lineOf(set *matSet, rel, substr string) int {
	for i, ln := range strings.Split(set.files[rel], "\n") {
		if strings.Contains(ln, substr) {
			return i + 1
		}
	}
	return 0
}

// newQueryMaterials 造一套"带页签/标签/切片/孤儿"的查询用材料。
func newQueryMaterials(t *testing.T) *matSet {
	t.Helper()
	return newMaterials(t, matOpts{
		Prog: "axmt500_wf",
		Pages: []matPage{
			{Seq: "1", ID: "1", Record: "s_browse", Columns: []string{"f1"}},
			{Seq: "12", ID: "73", Record: "s_detail73", Columns: []string{"xmdduc001", "xmdduc001_desc", "xmdducseq1", "xmdducseq"}},
		},
		Sections: []matSection{
			{ID: "p.input", Ver: "56", Src: "s", Slots: []matSlot{
				{Name: "input.a.page73.xmdduc001", Label: "刻号检核", DeclBody: "IF xmdduc001 = '' THEN\n  MESSAGE '必填'\nEND IF\n", Src: "c", Ver: "3", Body: []string{"   IF xmdduc001 = '' THEN", "      MESSAGE '必填'", "   END IF"}},
				{Name: "input.b.page73.xmdducseq1", Label: "项次带值", DeclBody: "LET xmdducseq1 = 1\n", Src: "s", Ver: "9", Body: []string{"   LET xmdducseq1 = 1"}},
				{Name: "input.c.page1.f1", Label: "客户", DeclBody: "LET f1 = 2\n", Src: "c", Ver: "1", Body: []string{"   LET f1 = 2"}},
			}},
			{ID: "p.construct", Ver: "12", Src: "s", Slots: []matSlot{
				{Name: "construct.c.page73.xmdduc001_desc", Label: "ON ACTION", DeclBody: "CALL show_desc()\n", Src: "c", Ver: "2", Body: []string{"   CALL show_desc()"}},
			}},
		},
		Orphans: []matOrphan{
			{Name: "input.a.page73.xmdducseq", Src: "c", Ver: "2", Body: "LET xmdducseq = 0\nMESSAGE '孤兒項次'\n"},
		},
		Slices: []matSlice{{Code: "a17", Used: 3, Current: 3, InBody: true}},
		TSDLabels: map[string]string{
			"f1": "客户", "xmdduc001": "刻号", "xmdduc003": "料号",
			"xmdduc001_desc": "品名", "xmdducseq1": "项次",
		},
		TSDPages: map[string]string{"page_73": "示例页签", "page_1": "基本数据"},
	})
}

// TestFixtureFields_PageIDVsSeqAndLabels 钉住 fields 的三种收窄方式与标签解析：
//   - page_id（73）与页签序号（12）都接受，但序号命中要在 note 里明确换算；
//   - 字段名过滤是子串；
//   - 画面中文名（含 `_desc` 派生字段）能翻译回列名。
func TestFixtureFields_PageIDVsSeqAndLabels(t *testing.T) {
	set := newQueryMaterials(t)
	set.mapIndex()
	lr := set.LinkRequest()

	all, doc, err := Fields(FieldsRequest{LinkRequest: lr})
	if err != nil {
		t.Fatalf("Fields: %v", err)
	}
	// xmdducseq 是孤儿（没有槽位），不是字段；有槽位的 4 个字段。
	if len(all) != 4 {
		t.Fatalf("字段 = %+v", all)
	}
	labels := map[string]string{}
	for _, f := range all {
		labels[f.Column] = f.Label
	}
	if labels["xmdduc001"] != "刻号" || labels["xmdduc001_desc"] != "品名" || labels["xmdducseq1"] != "项次" {
		t.Errorf("字段标签 = %v", labels)
	}
	if doc.Pages[1].Title != "示例页签" || doc.Pages[1].PageName != "page_73" {
		t.Errorf("页面 = %+v", doc.Pages[1])
	}

	// page_id=73：3 个字段（xmdduc001 / _desc / xmdducseq1）。
	if got, _, _ := Fields(FieldsRequest{LinkRequest: lr, Page: "73"}); len(got) != 3 {
		t.Errorf("按 page_id 73 = %d 个：%+v", len(got), got)
	}
	// 页签序号 12：应换算成 page_id 73 并在 note 里说明。
	got, d2, _ := Fields(FieldsRequest{LinkRequest: lr, Page: "12"})
	if len(got) != 3 {
		t.Errorf("按页签序号 12 = %d 个", len(got))
	}
	if len(d2.Notes) == 0 || !strings.Contains(d2.Notes[0], "page_id 是 73") {
		t.Errorf("序号换算提示 = %v", d2.Notes)
	}
	// 不存在的页：空结果 + 提示。
	if got, d3, _ := Fields(FieldsRequest{LinkRequest: lr, Page: "99"}); len(got) != 0 {
		t.Errorf("page 99 应为空: %+v", got)
	} else if len(d3.Notes) == 0 || !strings.Contains(d3.Notes[0], "page_id=99") {
		t.Errorf("空页提示 = %v", d3.Notes)
	}
	// 字段名子串过滤。
	if got, _, _ := Fields(FieldsRequest{LinkRequest: lr, Pattern: "xmdduc"}); len(got) != 3 {
		t.Errorf("模式 xmdduc = %d 个", len(got))
	}
	if got, _, _ := Fields(FieldsRequest{LinkRequest: lr, Pattern: "no_such"}); len(got) != 0 {
		t.Errorf("模式 no_such = %d 个", len(got))
	}

	// ResolveColumns：列名精确 → 后缀 → 标签精确 → 标签子串。
	if cols, by, lm := ResolveColumns(doc, "xmdduc001"); by != "column" || len(cols) != 1 ||
		lm["xmdduc001"] != "刻号" {
		t.Errorf("列名精确解析 = %v/%s/%v", cols, by, lm)
	}
	if cols, by, _ := ResolveColumns(doc, "_desc"); by != "column" || len(cols) != 1 || cols[0] != "xmdduc001_desc" {
		t.Errorf("列名后缀解析 = %v/%s", cols, by)
	}
	if cols, by, _ := ResolveColumns(doc, "刻号"); by != "label" || len(cols) != 1 || cols[0] != "xmdduc001" {
		t.Errorf("标签解析 = %v/%s", cols, by)
	}
	if cols, by, _ := ResolveColumns(doc, "品名"); by != "label" || cols[0] != "xmdduc001_desc" {
		t.Errorf("_desc 标签解析 = %v/%s", cols, by)
	}
	if cols, by, _ := ResolveColumns(doc, "项次"); by != "label" || cols[0] != "xmdducseq1" {
		t.Errorf("标签解析(项次) = %v/%s", cols, by)
	}
	if cols, by, _ := ResolveColumns(doc, "不存在的词"); cols != nil || by != "" {
		t.Errorf("无匹配 = %v/%s", cols, by)
	}
	if ColumnOfPoint("input.a.page73.xmdduc001") != "xmdduc001" {
		t.Errorf("ColumnOfPoint = %q", ColumnOfPoint("input.a.page73.xmdduc001"))
	}
	if _, how := MatchPoints(doc, "xmdducseq1"); how != "suffix" {
		t.Errorf("MatchPoints 命中方式 = %q", how)
	}

	// Show 按画面中文名：Match=label，note 要给回列名。
	res, err := Show(ShowRequest{LinkRequest: lr, Query: "刻号"})
	if err != nil {
		t.Fatalf("Show(刻号): %v", err)
	}
	if res.Match != "label" || len(res.Points) != 1 || res.Points[0].Label != "刻号" {
		t.Errorf("按标签 Show = %s / %+v", res.Match, res.Points)
	}
	if !strings.Contains(res.Note, "画面中文名") || !strings.Contains(res.Note, "xmdduc001") {
		t.Errorf("标签命中的 note = %q", res.Note)
	}
	// 完整点名：只接受精确匹配。
	res, err = Show(ShowRequest{LinkRequest: lr, Query: "input.a.page73.xmdduc001"})
	if err != nil || res.Match != "exact" || len(res.Points) != 1 {
		t.Fatalf("完整点名 = %+v / %v", res, err)
	}
	if len(res.Points[0].Content) == 0 || res.Points[0].Line0 == 0 {
		t.Errorf("精确点名缺正文/行号: %+v", res.Points[0])
	}
	// 不完整点名（带点但只近似）：明确报错，不能静默给近似结果。
	if _, err := Show(ShowRequest{LinkRequest: lr, Query: "input.a.page73.xmdduc00"}); !errors.Is(err, ErrPointNotFound) {
		t.Errorf("近似点名 err = %v，想要 ErrPointNotFound", err)
	}
}

// TestFixtureShow_OrphanPriorityAndSiblingHint 钉住"孤儿优先"：
// 查 `xmdducseq` 时，附近的活点 `xmdducseq1` 绝不能顶上来（差一个字符就是
// "有代码但不生效"与"这段有代码"两种结论）。同时钉住同页控件提示。
func TestFixtureShow_OrphanPriorityAndSiblingHint(t *testing.T) {
	set := newQueryMaterials(t)
	set.mapIndex()
	lr := set.LinkRequest()

	// 不带点：先查孤儿后缀/精确，命中孤儿就返回孤儿。
	res, err := Show(ShowRequest{LinkRequest: lr, Query: "xmdducseq"})
	if err != nil {
		t.Fatalf("Show(xmdducseq): %v", err)
	}
	if res.Match != "orphan" || len(res.Points) != 1 {
		t.Fatalf("孤儿没优先: match=%s points=%+v", res.Match, res.Points)
	}
	p := res.Points[0]
	if p.Name != "input.a.page73.xmdducseq" || !p.Orphan || p.OrphanReason != "no-slot" {
		t.Errorf("孤儿点 = %+v", p)
	}
	if len(p.Content) != 2 || !strings.Contains(p.Content[1], "孤兒項次") {
		t.Errorf("孤儿正文应来自 .tap: %#v", p.Content)
	}
	if p.Line0 == 0 || p.Line1 != p.Line0+1 {
		t.Errorf("孤儿行号 = %d-%d（应是 .tap 行号）", p.Line0, p.Line1)
	}
	if !strings.Contains(res.Note, "重产不会保留") {
		t.Errorf("孤儿 note 没说清后果: %q", res.Note)
	}
	// 同页相近控件提示：xmdducseq1（列名前缀关系）应被点名。
	if !strings.Contains(res.Note, "xmdducseq1") {
		t.Errorf("孤儿没给同页相近控件提示: %q", res.Note)
	}

	// 完整点名（带点）：同样认孤儿。
	res, err = Show(ShowRequest{LinkRequest: lr, Query: "input.a.page73.xmdducseq"})
	if err != nil || res.Match != "orphan" {
		t.Errorf("完整点名查孤儿 = %+v / %v", res, err)
	}
	// 活点仍然可查（后缀命中，且不是孤儿）。
	res, err = Show(ShowRequest{LinkRequest: lr, Query: "xmdducseq1"})
	if err != nil {
		t.Fatalf("Show(xmdducseq1): %v", err)
	}
	if res.Match == "orphan" || res.Points[0].Name != "input.b.page73.xmdducseq1" || res.Points[0].Orphan {
		t.Errorf("活点被孤儿抢答: %+v", res)
	}
	// 打错一个字符的完整点名：报错而不是给近似结果。
	if _, err := Show(ShowRequest{LinkRequest: lr, Query: "input.a.page73.xmdducseqX"}); !errors.Is(err, ErrPointNotFound) {
		t.Errorf("err = %v，想要 ErrPointNotFound", err)
	}
}

// TestFixtureAt_Matrix 钉住"这一行是什么"的全矩阵：
// 客制点 / 标准点 / 生成器自加 / freestyle 无 tap / 客制函数 / section 正文 /
// 切片归属（in-point 与 before-line）/ ContentStart / 行号越界 / 4gl 旧于 tap 的 stale。
func TestFixtureAt_Matrix(t *testing.T) {
	set := newQueryMaterials(t)
	set.mapIndex()
	lr := set.LinkRequest()
	rel := "src/axmt500_wf.4gl"

	// 客制点内部：Inside、客制结论、切片 in-point、section。
	customLine := lineOf(set, rel, "MESSAGE '必填'")
	res, err := At(AtRequest{LinkRequest: lr, Line: customLine})
	if err != nil {
		t.Fatalf("At: %v", err)
	}
	if !res.Inside || res.Point == nil || res.Point.Name != "input.a.page73.xmdduc001" {
		t.Fatalf("客制行 = %+v", res)
	}
	if !strings.Contains(res.Verdict, "客制加点") || !strings.Contains(res.Verdict, "保留") {
		t.Errorf("结论 = %q", res.Verdict)
	}
	if res.Section == nil || res.Section.ID != "p.input" {
		t.Errorf("section = %+v", res.Section)
	}
	if res.Slice == nil || res.Slice.Code != "a17" || res.SliceRelation != "in-point" {
		t.Errorf("切片归属 = %+v / %q", res.Slice, res.SliceRelation)
	}
	// ContentStart 是 Content[0] 的行号：少一个就整段偏移。
	if res.ContentStart != customLine-2 || res.Content[0] != strings.Split(set.files[rel], "\n")[res.ContentStart-1] {
		t.Errorf("ContentStart = %d（行 %d，上下文 2）", res.ContentStart, customLine)
	}
	if res.TotalLine != len(strings.Split(set.files[rel], "\n")) {
		t.Errorf("totalLines = %d", res.TotalLine)
	}

	// 标准点。
	stdLine := lineOf(set, rel, "LET xmdducseq1 = 1")
	if res, _ := At(AtRequest{LinkRequest: lr, Line: stdLine}); res.Point == nil ||
		!strings.Contains(res.Verdict, "标准加点") || res.Point.Kind() != "标准" {
		t.Errorf("标准行 = %+v", res)
	}

	// section 正文（不在任何加点内）：必须警告重产覆盖。
	var sectionBody int
	for i, ln := range strings.Split(set.files[rel], "\n") {
		if strings.Contains(ln, "{</section>}") {
			sectionBody = i + 1
			break
		}
	}
	out, err := At(AtRequest{LinkRequest: lr, Line: sectionBody})
	if err != nil {
		t.Fatalf("At(section): %v", err)
	}
	if out.Inside || out.Point != nil {
		t.Errorf("section 正文不该判 Inside: %+v", out)
	}
	if !strings.Contains(out.Verdict, "重产覆盖") {
		t.Errorf("section 正文结论 = %q", out.Verdict)
	}
	// 这里切片只是"该行之前最近的一条" → before-line。
	if out.Slice == nil || out.SliceRelation != "before-line" {
		t.Errorf("切片关系 = %+v / %q", out.Slice, out.SliceRelation)
	}

	// 上下文贴边：第 1 行与最后一行都不能越界。
	first, err := At(AtRequest{LinkRequest: lr, Line: 1, Context: 5})
	if err != nil {
		t.Fatalf("At(1): %v", err)
	}
	if first.ContentStart != 1 || len(first.Content) == 0 {
		t.Errorf("首行上下文 = %+v", first)
	}
	last := len(strings.Split(set.files[rel], "\n"))
	lastRes, err := At(AtRequest{LinkRequest: lr, Line: last, Context: 5})
	if err != nil {
		t.Fatalf("At(last): %v", err)
	}
	wantStart := last - 5
	if wantStart < 1 {
		wantStart = 1
	}
	if lastRes.ContentStart != wantStart || lastRes.Content[len(lastRes.Content)-1] != "" {
		t.Errorf("末行上下文 = %+v（ContentStart 想要 %d）", lastRes, wantStart)
	}
	// 越界行号。
	if _, err := At(AtRequest{LinkRequest: lr, Line: 0}); !errors.Is(err, ErrInvalidOpt) {
		t.Errorf("行号 0 err = %v", err)
	}
	if _, err := At(AtRequest{LinkRequest: lr, Line: last + 1}); !errors.Is(err, ErrInvalidOpt) {
		t.Errorf("行号越界 err = %v", err)
	}

	// 生成器自加的注释重复点：不判 Inside，且结论明说"不是你的代码"。
	extra := newMaterials(t, matOpts{
		Prog: "axmt501",
		Sections: []matSection{{
			ID: "p.input", Ver: "1", Src: "s", Slots: []matSlot{
				{Name: "input.a.page1.dup", Label: "重复点", G4Only: true, Commentd: true},
			},
		}},
		NoTab: true, NoTsd: true, No4fd: true,
	})
	extra.mapIndex()
	extraRel := "src/axmt501.4gl"
	extraLine := lineOf(extra, extraRel, "name=\"input.a.page1.dup\"")
	if extraLine == 0 {
		// 整行被注释时仍包含 name=。
		t.Fatalf("找不到注释点行: %s", extra.files[extraRel])
	}
	ex, err := At(AtRequest{LinkRequest: extra.LinkRequest(), Line: extraLine})
	if err != nil {
		t.Fatalf("At(extra): %v", err)
	}
	if ex.Inside || ex.Point == nil || !strings.Contains(ex.Verdict, "生成器自加") {
		t.Errorf("注释重复点 = %+v", ex)
	}

	// 客制函数（无加点标记）：按函数名定位，Inside=true。
	fn := newMaterials(t, matOpts{
		Prog: "axmt502",
		Funcs: []matFunc{
			{Name: "p_feature", Body: []string{"  LET y = 2"}},
			{Name: "p_std_func", Src: "s", Body: []string{"  LET z = 3"}},
		},
		NoTab: true, NoTsd: true, No4fd: true,
	})
	fn.mapIndex()
	fnRel := "src/axmt502.4gl"
	fnLine := lineOf(fn, fnRel, "LET y = 2")
	fr, err := At(AtRequest{LinkRequest: fn.LinkRequest(), Line: fnLine})
	if err != nil {
		t.Fatalf("At(func): %v", err)
	}
	// 客制函数（无加点标记）：按函数名定位，Inside=true；标签必须比"客制"更具体。
	if !fr.Inside || fr.Point == nil || fr.Point.Kind() != "客制函数" ||
		!strings.Contains(fr.Verdict, "客制函数") {
		t.Errorf("客制函数行 = %+v", fr)
	}
	// 标准侧的 function.* 声明：Custom 不成立，不能冒充客制函数。
	stdFnLine := lineOf(fn, fnRel, "LET z = 3")
	sr2, err := At(AtRequest{LinkRequest: fn.LinkRequest(), Line: stdFnLine})
	if err != nil {
		t.Fatalf("At(stdFunc): %v", err)
	}
	if !sr2.Inside || sr2.Point == nil || sr2.Point.Kind() != "函数（无标记）" ||
		strings.Contains(sr2.Verdict, "客制") {
		t.Errorf("标准函数声明行 = %+v", sr2)
	}

	// 无 .tap 的标准程序：行在点内，但结论说清"无法判断客制还是标准"。
	g4only := newMaterials(t, matOpts{
		Prog: "axmt503", NoTap: true,
		NoTab: true, NoTsd: true, No4fd: true,
	})
	g4only.mapIndex()
	onlyRel := "src/axmt503.4gl"
	onlyLine := lineOf(g4only, onlyRel, "LET f1 = 1")
	or, err := At(AtRequest{LinkRequest: g4only.LinkRequest(), Line: onlyLine})
	if err != nil {
		t.Fatalf("At(g4only): %v", err)
	}
	if !or.Inside || or.Point == nil || !strings.Contains(or.Verdict, "没有 .tap") {
		t.Errorf("无 tap 行 = %+v", or)
	}

	// 4gl 旧于 tap：At 要带 stale 标记与提示（行号可能已漂）。
	stale := newMaterials(t, matOpts{
		Prog:     "axmt504",
		TapMTime: "2026-11-21T12:00:00+08:00",
		G4MTime:  "2026-11-21T09:00:00+08:00",
		NoTab:    true, NoTsd: true, No4fd: true,
	})
	stale.mapIndex()
	staleRel := "src/axmt504.4gl"
	sr, err := At(AtRequest{LinkRequest: stale.LinkRequest(), Line: lineOf(stale, staleRel, "LET f1 = 1")})
	if err != nil {
		t.Fatalf("At(stale): %v", err)
	}
	if !sr.Stale || !strings.Contains(sr.StaleNote, "比 .tap") {
		t.Errorf("stale 标记 = %v / %q", sr.Stale, sr.StaleNote)
	}
}

// TestFixtureGrep_Matrix 钉住 grep 的归属、范围与截断：
//   - 4gl 命中要带点位/种类/section/切片；
//   - 只在 .tap 里有的代码（孤儿）在默认范围搜不到，scope=tap/all 才搜得到；
//   - 工作区给了 PkgPath 时 scope=all 优先查包内 .tap（apply 之后包才是真相）；
//   - limit 只截显示、total 保留。
func TestFixtureGrep_Matrix(t *testing.T) {
	set := newQueryMaterials(t)
	set.mapIndex()
	lr := set.LinkRequest()

	// 客制正文命中：归属、种类、切片、行号。
	res, err := Grep(GrepRequest{LinkRequest: lr, Pattern: "MESSAGE '必填'"})
	if err != nil {
		t.Fatalf("Grep: %v", err)
	}
	if res.Total != 1 || len(res.Hits) != 1 {
		t.Fatalf("命中 = %d/%+v", res.Total, res.Hits)
	}
	h := res.Hits[0]
	if h.Point != "input.a.page73.xmdduc001" || h.Kind != "客制" || !h.Inside ||
		h.Section != "p.input" || h.SliceCode != "a17" {
		t.Errorf("客制命中归属 = %+v", h)
	}
	if h.Text == "" || strings.Contains(h.Text, "\r") {
		t.Errorf("命中文本 = %q", h.Text)
	}
	// 标准正文命中。
	if res, _ := Grep(GrepRequest{LinkRequest: lr, Pattern: "xmdducseq1 = 1"}); len(res.Hits) != 1 ||
		res.Hits[0].Kind != "标准" {
		t.Errorf("标准命中 = %+v", res.Hits)
	}
	// 模板正文（section 里、不在点内）：Kind 空、Inside=false，但 section 有值。
	if res, _ := Grep(GrepRequest{LinkRequest: lr, Pattern: "{</section>"}); len(res.Hits) == 0 ||
		res.Hits[0].Inside || res.Hits[0].Section == "" {
		t.Errorf("模板正文命中 = %+v", res.Hits)
	}

	// limit 截显示不截总数。
	if res, _ := Grep(GrepRequest{LinkRequest: lr, Pattern: "add-point", Limit: 2}); res.Total < 3 || res.Shown != 2 {
		t.Errorf("limit = total:%d shown:%d", res.Total, res.Shown)
	}
	// 大小写开关。
	if res, _ := Grep(GrepRequest{LinkRequest: lr, Pattern: "message '必填'"}); res.Total != 1 {
		t.Errorf("忽略大小写命中 = %d", res.Total)
	}
	if res, _ := Grep(GrepRequest{LinkRequest: lr, Pattern: "message '必填'", Case: true}); res.Total != 0 {
		t.Errorf("区分大小写命中 = %d", res.Total)
	}
	// 空检索串是用法错误。
	if _, err := Grep(GrepRequest{LinkRequest: lr, Pattern: "  "}); err != nil {
		// 纯空白算非空；这里只钉住真空串。
		_ = err
	}
	if _, err := Grep(GrepRequest{LinkRequest: lr}); !errors.Is(err, ErrInvalidOpt) {
		t.Errorf("空检索串 err = %v，想要 ErrInvalidOpt", err)
	}

	// 孤儿正文只在 .tap 里：默认范围 0 条，scope=tap 才搜得到。
	orphanPat := "孤兒項次"
	if res, _ := Grep(GrepRequest{LinkRequest: lr, Pattern: orphanPat}); res.Total != 0 || res.TapTotal != 0 {
		t.Errorf("默认范围不该搜 tap: %+v", res)
	}
	res, err = Grep(GrepRequest{LinkRequest: lr, Pattern: orphanPat, Scope: "tap"})
	if err != nil {
		t.Fatalf("Grep(tap): %v", err)
	}
	if res.Total != 0 || res.TapTotal != 1 || len(res.TapHits) != 1 {
		t.Fatalf("tap 范围 = %+v", res)
	}
	if res.TapSource != "materials" || res.Scope != "tap" {
		t.Errorf("tap 来源/范围 = %q/%q", res.TapSource, res.Scope)
	}
	if hit := res.TapHits[0]; hit.Point != "input.a.page73.xmdducseq" || hit.HasSlot {
		t.Errorf("tap 命中 = %+v", hit)
	}
	// scope=all：两边都搜，并且提示"只在设计源里 / 没有槽位"。
	res, err = Grep(GrepRequest{LinkRequest: lr, Pattern: orphanPat, Scope: "all"})
	if err != nil {
		t.Fatalf("Grep(all): %v", err)
	}
	if res.TapTotal != 1 || !strings.Contains(res.Note, "只看设计源") && !strings.Contains(res.Note, "只在设计源") {
		t.Errorf("all 提示 = %q", res.Note)
	}
	if !strings.Contains(res.Note, "没有槽位") {
		t.Errorf("all 没提示孤儿不保留: %q", res.Note)
	}

	// 工作区包内 .tap：scope=all 时优先查包（包里的内容才是 apply 之后的真相）。
	pkgPath := matBuildTapPkg(t, set, "PACKAGE_ONLY_令牌")
	res, err = Grep(GrepRequest{LinkRequest: lr, Pattern: "PACKAGE_ONLY_令牌", Scope: "all", PkgPath: pkgPath})
	if err != nil {
		t.Fatalf("Grep(pkg): %v", err)
	}
	if res.TapSource != "pkg" || res.TapTotal != 1 {
		t.Errorf("包内 tap = %q/%d %+v", res.TapSource, res.TapTotal, res.TapHits)
	}
	// scope=pkg 但路径不是合法包：明确报错（不能回退成材料里的旧 .tap 而假装是包）。
	if _, err := Grep(GrepRequest{LinkRequest: lr, Pattern: "x", Scope: "pkg", PkgPath: filepath.Join(t.TempDir(), "nope.tzc")}); err == nil {
		t.Error("非法包路径应报错")
	}
	// scope=pkg 且没给 PkgPath：退化为查材料（工作区命令才给 PkgPath）。
	if res, err := Grep(GrepRequest{LinkRequest: lr, Pattern: orphanPat, Scope: "pkg"}); err != nil || res.TapTotal != 1 {
		t.Errorf("pkg 无路径时的退化 = %+v / %v", res, err)
	}

	// 4gl 旧于 tap：命中结果带 stale 提示。
	stale := newMaterials(t, matOpts{
		Prog:     "axmt505",
		TapMTime: "2026-11-21T12:00:00+08:00",
		G4MTime:  "2026-11-21T09:00:00+08:00",
	})
	stale.mapIndex()
	if res, err := Grep(GrepRequest{LinkRequest: stale.LinkRequest(), Pattern: "LET f1"}); err != nil ||
		res.StaleNote == "" {
		t.Errorf("stale 提示 = %q / %v", res.StaleNote, err)
	}
}

// matBuildTapPkg 用材料里的 .tap 造一个 .tzc（模拟工作区 apply 之后的包），
// 并往 tap 里塞一个只存在于包内的标记，用来证明 grep 确实读了包。
func matBuildTapPkg(t *testing.T, set *matSet, marker string) string {
	t.Helper()
	tap := set.files["dzx/axmt500_wf.tap"]
	tap = strings.Replace(tap, "</add_points>",
		"<point name=\"input.a.page73.xmdducseq\" mapping=\"\" src=\"c\" ver=\"9\" status=\"\"><![CDATA[LET x = 1 # "+marker+"\n]]></point>\n</add_points>", 1)
	dir := t.TempDir()
	p := filepath.Join(dir, set.Prog+"(c).tzc")
	f, err := os.Create(p)
	if err != nil {
		t.Fatalf("建包: %v", err)
	}
	err = pkgfile.WriteZip(f, []pkgfile.WriteEntry{
		{Name: set.Prog + ".tap", Data: []byte(tap), Mode: 0o664},
		{Name: set.Prog + ".tgl", Data: []byte("#+ tgl\n"), Mode: 0o664},
		{Name: "ver", Data: []byte("1.0\n"), Store: true, Mode: 0o775},
	})
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		t.Fatalf("写包: %v", err)
	}
	return p
}
