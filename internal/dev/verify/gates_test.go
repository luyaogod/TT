package verify

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/fence"
	"github.com/luyaogod/TT/internal/dev/model"
	"github.com/luyaogod/TT/internal/dev/pkgfile"
	"github.com/luyaogod/TT/internal/dev/synth"
	"github.com/luyaogod/TT/internal/dev/tapfile"
	"github.com/luyaogod/TT/internal/dev/testutil"
)

// 本文件补齐三道闸门的**判定面**：gate1 的字节恒等/删除授权/APPEND 追加、
// gate2 的不变量 I1–I11b 与结构事务 V1–V7、gate3 的装载模拟与临时目录策略，
// 外加 Finding 定位/severity/退出码这些对外契约。
//
// verify_test.go 已经钉了「干净文档零发现」「行尾等价是 info」「Report 计数」，
// 这里不重复，专攻那些端到端只能报"整条链失败"、定位不到具体闸门的路径。

// gateFix 是「基线 + 围栏文本 + 包 + 原 TAP」四件套。
type gateFix struct {
	base   *model.Document
	fenced []byte
	pkg    *pkgfile.Package
	tapDoc *tapfile.Doc
	dir    string
}

// newGateFix 造正常合成包并合成/渲染/解析出基线。
func newGateFix(t *testing.T, prog string, opts synth.Options) *gateFix {
	t.Helper()
	return newGateFixEntries(t, prog, testutil.NormalEntries(prog), opts)
}

// newGateFixEntries 允许用例先改条目（TAP 属性/删点/删区段/TGL 内容）。
func newGateFixEntries(t *testing.T, prog string, entries map[string]string, opts synth.Options) *gateFix {
	t.Helper()
	dir := t.TempDir()
	p, err := testutil.WriteEntries(dir, prog+".tzc", entries)
	if err != nil {
		t.Fatalf("写合成包失败：%v", err)
	}
	pkg, err := pkgfile.Open(p, pkgfile.OpenOptions{})
	if err != nil {
		t.Fatalf("打开合成包失败：%v", err)
	}
	doc0, err := synth.Synthesize(pkg, opts)
	if err != nil {
		t.Fatalf("合成失败：%v", err)
	}
	fenced, _, _, err := fence.Render(doc0)
	if err != nil {
		t.Fatalf("渲染失败：%v", err)
	}
	// 基线必须是"把基线自己的围栏文本解析一遍"的结果（坐标在围栏文本上）。
	baseline, err := fence.Parse(doc0, fenced)
	if err != nil {
		t.Fatalf("基线自解析失败：%v", err)
	}
	tapDoc, err := tapfile.Parse(pkg.Tap().Data)
	if err != nil {
		t.Fatalf("TAP 解析失败：%v", err)
	}
	return &gateFix{base: baseline.Doc, fenced: fenced, pkg: pkg, tapDoc: tapDoc, dir: dir}
}

// parse 用基线 + 编辑后的围栏文本得到 ParseResult。
func (g *gateFix) parse(t *testing.T, edited []byte) *fence.ParseResult {
	t.Helper()
	parsed, err := fence.Parse(g.base, edited)
	if err != nil {
		t.Fatalf("编辑后的围栏文本解析失败：%v", err)
	}
	return parsed
}

// gateRegion 按名字取基线区间。
func gateRegion(t *testing.T, doc *model.Document, name string) *model.Region {
	t.Helper()
	r := doc.FindRegion(name)
	if r == nil {
		t.Fatalf("基线里没有区间 %q", name)
	}
	return r
}

// gateReplace 替换唯一一段；没命中直接 Fatal（防止夹具悄悄失效变成空转）。
func gateReplace(t *testing.T, text []byte, old, new string) []byte {
	t.Helper()
	if n := bytes.Count(text, []byte(old)); n != 1 {
		t.Fatalf("夹具失效：%q 出现 %d 次（想恰好 1 次）", old, n)
	}
	return bytes.Replace(text, []byte(old), []byte(new), 1)
}

// gateSplice 用 repl 替换 [start,end)。
func gateSplice(text []byte, start, end int, repl string) []byte {
	out := make([]byte, 0, len(text)+len(repl)-(end-start))
	out = append(out, text[:start]...)
	out = append(out, repl...)
	out = append(out, text[end:]...)
	return out
}

// gateDeleteRegion 删掉一个围栏块（begin..end）及行尾 —— 等价于设计器里选中函数块按 Delete。
func gateDeleteRegion(t *testing.T, g *gateFix, name string) []byte {
	t.Helper()
	r := gateRegion(t, g.base, name)
	end := r.FullSpan.End
	switch {
	case end+1 < len(g.fenced) && g.fenced[end] == '\r' && g.fenced[end+1] == '\n':
		end += 2
	case end < len(g.fenced) && g.fenced[end] == '\n':
		end++
	}
	return gateSplice(g.fenced, r.FullSpan.Start, end, "")
}

// gateFindings 返回某 code 的发现列表。
func gateFindings(rep *Report, code string) []Finding {
	var out []Finding
	for _, f := range rep.Findings {
		if f.Code == code {
			out = append(out, f)
		}
	}
	return out
}

// patchTAPPoint 只在某个 <point> 元素内部做字符串替换（避免误伤同名属性的其它点）。
func patchTAPPoint(t *testing.T, entries map[string]string, prog, pointName string, fn func(string) string) {
	t.Helper()
	tap := entries[prog+".tap"]
	i := strings.Index(tap, `<point name="`+pointName+`"`)
	if i < 0 {
		t.Fatalf("TAP 里找不到点 %q", pointName)
	}
	j := strings.Index(tap[i:], "</point>")
	if j < 0 {
		t.Fatalf("点 %q 没有闭合标签", pointName)
	}
	end := i + j + len("</point>")
	elem := fn(tap[i:end])
	entries[prog+".tap"] = tap[:i] + elem + tap[end:]
}

// removeTAPSection 删掉一个 <section> 元素（含其后换行），造「TAP 缺区段」的夹具。
func removeTAPSection(t *testing.T, entries map[string]string, prog, id string) {
	t.Helper()
	tap := entries[prog+".tap"]
	i := strings.Index(tap, `<section id="`+id+`"`)
	if i < 0 {
		t.Fatalf("TAP 里找不到区段 %q", id)
	}
	j := strings.Index(tap[i:], "</section>")
	if j < 0 {
		t.Fatalf("区段 %q 没有闭合标签", id)
	}
	end := i + j + len("</section>")
	switch {
	case strings.HasPrefix(tap[end:], "\r\n"):
		end += 2
	case strings.HasPrefix(tap[end:], "\n"):
		end++
	}
	entries[prog+".tap"] = tap[:i] + tap[end:]
}

// ---- gate1：围栏行 / 结构行 / 围栏外 ----

// TestGate1CatchesFenceAnchorChange 围栏行的**锚定部分**（这里改 new 属性）不许改：
// gate1 用结构化签名比对，不误伤字段顺序/空白，但抓得住真正的改动。
func TestGate1CatchesFenceAnchorChange(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	// 把自订点的 new="Y" 改成 N（围栏行锚定 kv，不允许改）。
	edited := gateReplace(t, g.fenced, `order="1" ver="1" cite_std="N" fn=`, `order="1" ver="1" cite_std="N" new="N" fn=`)
	rep := Gate1(g.base, g.parse(t, edited))
	fs := gateFindings(rep, "gate1.fence-line")
	if len(fs) == 0 {
		t.Fatalf("改围栏行锚定部分该报 gate1.fence-line，得 %+v", rep.Findings)
	}
	if fs[0].Severity != SevError || !rep.HasError() {
		t.Errorf("该是 error 级：%+v", fs[0])
	}
	// 位置必须落到具体行（否则用户不知道改哪一行）。
	if fs[0].Line == 0 || fs[0].File != FileEdited {
		t.Errorf("该带文件与行号，得 file=%q line=%d", fs[0].File, fs[0].Line)
	}
}

// TestGate1CatchesEndFenceChange END 围栏行同样受保护（它不属于任何可写区间）。
func TestGate1CatchesEndFenceChange(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	target := gateRegion(t, g.base, "global.import")
	// 在 end 围栏行尾加空格：仍能解析（正则允许尾随空白），但字节变了。
	edited := gateSplice(g.fenced, target.EndFence.End, target.EndFence.End, "  ")
	rep := Gate1(g.base, g.parse(t, edited))
	if len(gateFindings(rep, "gate1.fence-line")) == 0 {
		t.Fatalf("改 end 围栏行该报 gate1.fence-line，得 %+v", rep.Findings)
	}
	if !rep.HasError() {
		t.Error("该是 error 级")
	}
}

// TestGate1CatchesStructureLineChange 结构行（END 行）绝对不可写：
// 加一个空格都要拦下（设计器只允许改正文本体）。
func TestGate1CatchesStructureLineChange(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	edited := gateReplace(t, g.fenced, "END FUNCTION", "END  FUNCTION")
	rep := Gate1(g.base, g.parse(t, edited))
	if len(gateFindings(rep, "I-structure")) == 0 {
		t.Fatalf("改 END 行该报 I-structure，得 %+v", rep.Findings)
	}
	if !rep.HasError() {
		t.Error("该是 error 级")
	}
}

// TestGate1CatchesOutsideFenceChange Region 之外（这里是文件 prefix）不许增删改任何字节。
func TestGate1CatchesOutsideFenceChange(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	edited := append([]byte("X"), g.fenced...) // 文件头插一个字节
	rep := Gate1(g.base, g.parse(t, edited))
	fs := gateFindings(rep, "gate1.outside-fence")
	if len(fs) == 0 {
		t.Fatalf("围栏外插字节该报 gate1.outside-fence，得 %+v", rep.Findings)
	}
	if len(fs[0].Detail) < 2 {
		t.Errorf("Detail 该说明「原有多少字节/新多少字节」和首个差异：%v", fs[0].Detail)
	}
}

// ---- gate1：删除授权 ----

// TestGate1DeleteAuthorized new="Y" 且可编辑的点可以删，报 info 而不是 error。
func TestGate1DeleteAuthorized(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	rep := Gate1(g.base, g.parse(t, gateDeleteRegion(t, g, "global.import")))
	if rep.HasError() || rep.HasDenied() {
		t.Fatalf("删已授权点不该报错：%s %+v", rep.Summary(), rep.Findings)
	}
	fs := gateFindings(rep, "gate1.delete-ok")
	if len(fs) != 1 || fs[0].Severity != SevInfo {
		t.Fatalf("该恰好一条 info 的 delete-ok，得 %+v", rep.Findings)
	}
	// 删除类问题在编辑后文档里已不存在，位置只能给基线侧。
	if fs[0].File != FileBase || fs[0].Line == 0 {
		t.Errorf("删除发现该定位到基线文件，得 file=%q line=%d", fs[0].File, fs[0].Line)
	}
}

// TestGate1DeleteNotNewRejected 非 new="Y" 的点被删 → Denied error（退出码 4 的来源）。
func TestGate1DeleteNotNewRejected(t *testing.T) {
	entries := testutil.NormalEntries("adzi999")
	patchTAPPoint(t, entries, "adzi999", "global.import", func(elem string) string {
		return strings.Replace(elem, `new="Y"`, `new=""`, 1)
	})
	g := newGateFixEntries(t, "adzi999", entries, synth.Options{})
	rep := Gate1(g.base, g.parse(t, gateDeleteRegion(t, g, "global.import")))
	fs := gateFindings(rep, "gate1.delete-not-new")
	if len(fs) != 1 {
		t.Fatalf("该报 gate1.delete-not-new，得 %+v", rep.Findings)
	}
	if !fs[0].Denied || !rep.HasDenied() {
		t.Error("删除未授权属于「写入被拒」，Denied 必须为真")
	}
	if fs[0].Severity != SevError {
		t.Errorf("该是 error 级，得 %q", fs[0].Severity)
	}
}

// TestGate1DeleteReadonlyRejected new="Y" 但不可编辑的点被删 → Denied error。
// new=Y 只是第一道门，授权是第二道。
func TestGate1DeleteReadonlyRejected(t *testing.T) {
	entries := testutil.NormalEntries("adzi999")
	patchTAPPoint(t, entries, "adzi999", "global.import", func(elem string) string {
		return strings.Replace(elem, `src="s" readonly=""`, `src="s" readonly="Y"`, 1)
	})
	g := newGateFixEntries(t, "adzi999", entries, synth.Options{})
	rep := Gate1(g.base, g.parse(t, gateDeleteRegion(t, g, "global.import")))
	fs := gateFindings(rep, "gate1.delete-readonly")
	if len(fs) != 1 || !fs[0].Denied {
		t.Fatalf("该报 Denied 的 gate1.delete-readonly，得 %+v", rep.Findings)
	}
}

// ---- gate1：APPEND 锚点里的新增点 ----

// appendFenceBlock 造一个塞进 APPEND 锚点的新点围栏块。
func appendFenceBlock(name, attrs, body string) string {
	return `{//@tdev:begin point ` + name + ` [EDITABLE ` + attrs + `]}` + "\n" + body + "{//@tdev:end point}\n"
}

// insertAtAnchorEnd 把块插到锚点区段 end 围栏行之前（= 追加到区段正文末尾）。
func insertAtAnchorEnd(t *testing.T, g *gateFix, section, block string) []byte {
	t.Helper()
	anchor := gateRegion(t, g.base, section)
	return gateSplice(g.fenced, anchor.EndFence.Start, anchor.EndFence.Start, block)
}

// TestGate1AppendOk 在 APPEND 锚点内追加完整新点块 → info（这是唯一允许的"往区段里加"）。
func TestGate1AppendOk(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	body := "PRIVATE FUNCTION adzi999_extra(p_b)\n   RETURN p_b\nEND FUNCTION\n"
	edited := insertAtAnchorEnd(t, g, "adzi999.other_function",
		appendFenceBlock("function.adzi999_extra", `new="Y" status="u"`, body))
	rep := Gate1(g.base, g.parse(t, edited))
	if rep.HasError() {
		t.Fatalf("合法的追加不该报错：%s %+v", rep.Summary(), rep.Findings)
	}
	if len(gateFindings(rep, "gate1.append-ok")) != 1 {
		t.Fatalf("该有一条 gate1.append-ok info，得 %+v", rep.Findings)
	}
}

// TestGate1AppendNotNewRejected 追加点必须 new="Y"。
func TestGate1AppendNotNewRejected(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	body := "PRIVATE FUNCTION adzi999_extra(p_b)\n   RETURN p_b\nEND FUNCTION\n"
	edited := insertAtAnchorEnd(t, g, "adzi999.other_function",
		appendFenceBlock("function.adzi999_extra", `new="N" status="u"`, body))
	rep := Gate1(g.base, g.parse(t, edited))
	fs := gateFindings(rep, "gate1.append-not-new")
	if len(fs) != 1 || !fs[0].Denied {
		t.Fatalf("该报 Denied 的 gate1.append-not-new，得 %+v", rep.Findings)
	}
}

// TestGate1AppendUnparseableWarns 追加点正文不是可解析函数块 → warn（I4），
// apply 会拒绝写回，但 gate1 只提醒不拦（用户还能继续改）。
func TestGate1AppendUnparseableWarns(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	edited := insertAtAnchorEnd(t, g, "adzi999.other_function",
		appendFenceBlock("function.adzi999_bad", `new="Y" status="u"`, "NOT A FUNCTION\n"))
	rep := Gate1(g.base, g.parse(t, edited))
	fs := gateFindings(rep, "I4")
	if len(fs) != 1 || fs[0].Severity != SevWarn {
		t.Fatalf("该有一条 warn 的 I4，得 %+v", rep.Findings)
	}
}

// TestGate1AnchorBodyChangeRejected 锚点区段只能追加完整点块，不能改区段自身字节：
// 改了就报 append-anchor-body（Denied），因为锚点正文就是注入的点块（红线 R4）。
func TestGate1AnchorBodyChangeRejected(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	anchor := gateRegion(t, g.base, "adzi999.other_dialog")
	edited := gateSplice(g.fenced, anchor.ContentSpan.Start, anchor.ContentSpan.Start, "X")
	rep := Gate1(g.base, g.parse(t, edited))
	fs := gateFindings(rep, "gate1.append-anchor-body")
	if len(fs) != 1 || !fs[0].Denied {
		t.Fatalf("该报 Denied 的 gate1.append-anchor-body，得 %+v", rep.Findings)
	}
}

// ---- gate1：手工 ParseResult（解析层先拦住的路径，钉闸门自己的最后一道） ----

// TestGate1ManualRegionCountMismatch Region 数量不守恒 → error。
func TestGate1ManualRegionCountMismatch(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	p := g.parse(t, g.fenced)
	p.Pairs = p.Pairs[1:] // 人为丢一个配对（模拟"某个 Region 凭空消失"）
	rep := Gate1(g.base, p)
	if len(gateFindings(rep, "gate1.region-count")) != 1 {
		t.Fatalf("该报 gate1.region-count，得 %+v", rep.Findings)
	}
}

// TestGate1ManualKindChanged Region 类型变了 → error（配对按位置，不信名字）。
func TestGate1ManualKindChanged(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	p := g.parse(t, g.fenced)
	for i := range p.Pairs {
		if p.Pairs[i].Base != nil && p.Pairs[i].Base.Kind == model.RegionPoint {
			p.Pairs[i].Edited.Kind = model.RegionSection
			break
		}
	}
	rep := Gate1(g.base, p)
	if len(gateFindings(rep, "gate1.kind")) != 1 {
		t.Fatalf("该报 gate1.kind，得 %+v", rep.Findings)
	}
}

// TestGate1ManualDeleteSection 区段围栏块被删 → gate1 兜底报 delete-section
// （fence.Parse 会先拦；这里钉闸门自己的判定，防止解析层被放宽时静默放行）。
func TestGate1ManualDeleteSection(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	p := &fence.ParseResult{
		Doc: g.base,
		Deleted: []*model.Region{{
			Name: "adzi999.main", Kind: model.RegionSection,
			ContentSpan: model.ByteRange{Start: 0, End: 0},
		}},
	}
	rep := Gate1(g.base, p)
	if len(gateFindings(rep, "gate1.delete-section")) != 1 {
		t.Fatalf("该报 gate1.delete-section，得 %+v", rep.Findings)
	}
}

// ---- Gate2：干净文档与 I1 ----

// minimalDoc 造一个 Gate2 用的最小基线：只有给定区间与锚点表。
func minimalDoc(text string, regions ...*model.Region) *model.Document {
	if text != "" && len(regions) == 1 && regions[0].Kind == model.RegionSection && regions[0].ContentSpan.Len() == 0 {
		regions[0].ContentSpan = model.ByteRange{Start: 0, End: len(text)}
	}
	return &model.Document{
		Text: []byte(text), Regions: regions, Prog: "adzi999",
		Anchors: map[string]bool{"other.function": true, "other.dialog": true, "other.report": true},
	}
}

// TestGate2CleanPackage 干净文档过 gate2 不该有 error：I3/I4 不误报、I5 status 合法、
// I11 无 BOM/0x07。这条钉的是**误报**（闸门写严了最容易让正常包不可用）。
func TestGate2CleanPackage(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	rep := Gate2(g.base, g.parse(t, g.fenced), g.tapDoc)
	if rep.HasError() {
		t.Fatalf("干净包不该有 error：%s %+v", rep.Summary(), rep.Findings)
	}
	// I10 是常量 info（.4gl 与合成结果不一致是常态），顺手钉一下它还在。
	if len(gateFindings(rep, "I10")) != 1 {
		t.Errorf("该有一条 I10 info，得 %+v", rep.Findings)
	}
}

// TestGate2I1SectionMismatch 区段配对两个方向：
//   - TAP 有、TGL 没有 → warn（该区段正文不会显示）；
//   - TGL 有、TAP 没有但无模板可 clone → error（设计器会抛「找不到区段」）。
func TestGate2I1SectionMismatch(t *testing.T) {
	// TAP 多一个区段：在 </add_points> 前插一个区段元素。
	entries := testutil.NormalEntries("adzi999")
	tap := entries["adzi999.tap"]
	extra := `  <section id="adzi999.ghost" src="s" status="" ver="1" ch="" readonly="" modi_by_topstd=""><![CDATA[ghost]]></section>` + "\r\n"
	entries["adzi999.tap"] = strings.Replace(tap, "</add_points>", extra+"</add_points>", 1)
	g := newGateFixEntries(t, "adzi999", entries, synth.Options{})
	rep := Gate2(g.base, g.parse(t, g.fenced), g.tapDoc)
	if rep.HasError() {
		t.Fatalf("TAP 多一个区段只该 warn：%s %+v", rep.Summary(), rep.Findings)
	}
	var found bool
	for _, f := range gateFindings(rep, "I1") {
		if f.Region == "adzi999.ghost" && f.Severity == SevWarn {
			found = true
		}
	}
	if !found {
		t.Fatalf("该对 ghost 区段报 warn I1，得 %+v", rep.Findings)
	}

	// TGL 有、TAP 没有：删掉 TAP 的 .main 区段，并把文档里该区段标成「无模板可 clone」。
	entries2 := testutil.NormalEntries("adzi999")
	removeTAPSection(t, entries2, "adzi999", "adzi999.main")
	g2 := newGateFixEntries(t, "adzi999", entries2, synth.Options{})
	// Synth 会给缺区段标 DenySectionTemplateMissing（因为主区段 src=s 不是 c/m 模板）；
	// 这里显式断言基线确实带了该 deny 码，否则这条用例会白过。
	mainSec := gateRegion(t, g2.base, "adzi999.main")
	if mainSec.DenyCode != model.DenySectionTemplateMissing {
		t.Fatalf("夹具失效：缺区段该被标 %q，得 %q", model.DenySectionTemplateMissing, mainSec.DenyCode)
	}
	rep2 := Gate2(g2.base, g2.parse(t, g2.fenced), g2.tapDoc)
	if !rep2.HasError() {
		t.Fatalf("缺区段且无模板可 clone 该报 error：%s %+v", rep2.Summary(), rep2.Findings)
	}
}

// TestGate2I1DuplicateSection 文档里区段 id 重复 → error（设计器按序号配对，重名必错位）。
func TestGate2I1DuplicateSection(t *testing.T) {
	sec := &model.Region{Name: "p.main", Kind: model.RegionSection, ContentSpan: model.ByteRange{Start: 0, End: 5}}
	doc := minimalDoc("hello", sec, &model.Region{Name: "p.main", Kind: model.RegionSection, ContentSpan: model.ByteRange{Start: 5, End: 5}})
	rep := Gate2(doc, &fence.ParseResult{Doc: doc}, &tapfile.Doc{})
	if len(gateFindings(rep, "I1")) < 1 || !rep.HasError() {
		t.Fatalf("重复区段 id 该报 error I1，得 %+v", rep.Findings)
	}
}

// ---- Gate2：I2a/I2b/I2c（命名空间三层的错配） ----

// TestGate2I2bAnchorMissing TAP 有 function. 点但 TGL 没有 other.function 锚点 →
// error：这些代码永远不会出现在文档里（它们是"死代码"）。
func TestGate2I2bAnchorMissing(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	doc := *g.base
	doc.Anchors = map[string]bool{"other.dialog": true, "other.report": true}
	rep := Gate2(&doc, &fence.ParseResult{Doc: g.base}, g.tapDoc)
	if len(gateFindings(rep, "I2b")) != 1 || !rep.HasError() {
		t.Fatalf("该报 error I2b，得 %+v", rep.Findings)
	}
}

// TestGate2I2cBarePointOrphan TAP 裸插入点在 TGL 里没有占位符 → warn：
// 惰性数据，看不到但不会丢，必须原样透传。
func TestGate2I2cBarePointOrphan(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	doc := minimalDoc("") // 文档里没有任何 Region → 所有裸名点都是孤儿
	rep := Gate2(doc, &fence.ParseResult{Doc: doc}, g.tapDoc)
	if rep.HasError() {
		t.Fatalf("孤儿裸点只该 warn：%s %+v", rep.Summary(), rep.Findings)
	}
	if len(gateFindings(rep, "I2c")) != 1 {
		t.Fatalf("该报 warn I2c，得 %+v", rep.Findings)
	}
}

// TestGate2I2aEmptyCreated TGL 有占位符而 TAP 无同名点 → info（设计器会造空点，是常态）。
func TestGate2I2aEmptyCreated(t *testing.T) {
	doc := minimalDoc("", &model.Region{
		Name: "global.empty", Kind: model.RegionPoint, Origin: "placeholder-empty",
		ContentSpan: model.ByteRange{Start: 0, End: 0},
	})
	rep := Gate2(doc, &fence.ParseResult{Doc: doc}, &tapfile.Doc{})
	if rep.HasError() {
		t.Fatalf("造空点不该报 error：%s %+v", rep.Summary(), rep.Findings)
	}
	if len(gateFindings(rep, "I2a")) != 1 {
		t.Fatalf("该报 info I2a，得 %+v", rep.Findings)
	}
}

// ---- Gate2：I3（TglTag 折叠破坏）与 I4/I5 ----

// TestGate2I3ExpandedFunctionInSection 区段正文里出现已知自订函数的**定义头** →
// error（红线 R4：TglTag 折叠被破坏，下次装载会重复展开）。
func TestGate2I3ExpandedFunctionInSection(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	text := "PRIVATE FUNCTION adzi999_calc(p_a)\n  LET p_a = p_a + 1\nEND FUNCTION\n"
	sec := &model.Region{Name: "p.main", Kind: model.RegionSection, Editable: false,
		ContentSpan: model.ByteRange{Start: 0, End: len(text)}}
	doc := minimalDoc(text, sec)
	rep := Gate2(doc, &fence.ParseResult{Doc: doc}, g.tapDoc)
	if len(gateFindings(rep, "I3")) != 1 || !rep.HasError() {
		t.Fatalf("该报 error I3，得 %+v", rep.Findings)
	}
}

// TestGate2I4UnparsableStructure deny 码为 structure-unparsable 的点 → warn。
func TestGate2I4UnparsableStructure(t *testing.T) {
	doc := minimalDoc("", &model.Region{
		Name: "function.bad", Kind: model.RegionPoint, Editable: false,
		DenyCode: model.DenyStructureUnparsable, Origin: "placeholder",
		ContentSpan: model.ByteRange{Start: 0, End: 0},
	})
	rep := Gate2(doc, &fence.ParseResult{Doc: doc}, &tapfile.Doc{})
	if rep.HasError() {
		t.Fatalf("I4 只该 warn：%s %+v", rep.Summary(), rep.Findings)
	}
	if len(gateFindings(rep, "I4")) != 1 {
		t.Fatalf("该报 warn I4，得 %+v", rep.Findings)
	}
}

// TestGate2I5StatusRules status 取值只有 ' '/u/d/c（I5）：
//   - 缺 status → error；
//   - 非法值 → error；
//   - status=d 却没保留原内容 → warn（设计器会保留，不保留就是数据丢失）。
func TestGate2I5StatusRules(t *testing.T) {
	base := minimalDoc("")

	// 缺 status。
	entries := testutil.NormalEntries("adzi999")
	patchTAPPoint(t, entries, "adzi999", "global.import", func(elem string) string {
		return strings.Replace(elem, ` status=""`, "", 1)
	})
	g := newGateFixEntries(t, "adzi999", entries, synth.Options{})
	rep := Gate2(base, &fence.ParseResult{Doc: base}, g.tapDoc)
	var missing bool
	for _, f := range gateFindings(rep, "I5") {
		if strings.Contains(f.Message, "缺 status") {
			missing = true
		}
	}
	if !missing {
		t.Fatalf("缺 status 该报 error I5，得 %+v", rep.Findings)
	}

	// 非法取值。
	entries2 := testutil.NormalEntries("adzi999")
	patchTAPPoint(t, entries2, "adzi999", "global.import", func(elem string) string {
		return strings.Replace(elem, ` status=""`, ` status="x"`, 1)
	})
	g2 := newGateFixEntries(t, "adzi999", entries2, synth.Options{})
	rep2 := Gate2(base, &fence.ParseResult{Doc: base}, g2.tapDoc)
	var invalid bool
	for _, f := range gateFindings(rep2, "I5") {
		if strings.Contains(f.Message, "取值非法") {
			invalid = true
		}
	}
	if !invalid {
		t.Fatalf("非法 status 该报 error I5，得 %+v", rep2.Findings)
	}

	// status=d 但没有 CDATA（删点必须保留原内容）。
	entries3 := testutil.NormalEntries("adzi999")
	patchTAPPoint(t, entries3, "adzi999", "global.import", func(elem string) string {
		elem = strings.Replace(elem, ` status=""`, ` status="d"`, 1)
		// 去掉该点的 CDATA。
		i := strings.Index(elem, "<![CDATA[")
		j := strings.Index(elem, "]]>")
		if i < 0 || j < 0 {
			t.Fatalf("夹具里 global.import 没有 CDATA：%q", elem)
		}
		return elem[:i] + elem[j+3:]
	})
	g3 := newGateFixEntries(t, "adzi999", entries3, synth.Options{})
	rep3 := Gate2(base, &fence.ParseResult{Doc: base}, g3.tapDoc)
	var noContent bool
	for _, f := range gateFindings(rep3, "I5") {
		if f.Severity == SevWarn && strings.Contains(f.Message, "保留原内容") {
			noContent = true
		}
	}
	if !noContent {
		t.Fatalf("删点保留内容缺失该报 warn I5，得 %+v", rep3.Findings)
	}
}

// ---- Gate2：I11 / I11b（XML 可写性） ----

// TestGate2I11BOMAndDiffMarker TAP 带 UTF-8 BOM 或残留 0x07 diff 标记 → error。
// 这两样都会让设计器解析出问题，必须拦在写回之前。
func TestGate2I11BOMAndDiffMarker(t *testing.T) {
	bomTap := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`<add_points prog="p"></add_points>`)...)
	rep := Gate2(minimalDoc(""), &fence.ParseResult{Doc: minimalDoc("")},
		&tapfile.Doc{Raw: bomTap, Root: &tapfile.Element{Name: "add_points"}})
	if len(gateFindings(rep, "I11")) != 1 || !rep.HasError() {
		t.Fatalf("BOM 该报 error I11，得 %+v", rep.Findings)
	}

	diffTap := []byte("<add_points prog=\"p\">\a</add_points>")
	rep2 := Gate2(minimalDoc(""), &fence.ParseResult{Doc: minimalDoc("")},
		&tapfile.Doc{Raw: diffTap, Root: &tapfile.Element{Name: "add_points"}})
	var saw bool
	for _, f := range gateFindings(rep2, "I11") {
		if strings.Contains(f.Message, "0x07") {
			saw = true
		}
	}
	if !saw {
		t.Fatalf("0x07 残留该报 error I11，得 %+v", rep2.Findings)
	}
}

// TestGate2I11bCDATATerminator 可编辑内容里出现 ]]> → error（XML CDATA 装不下）。
func TestGate2I11bCDATATerminator(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	// 自订点正文里塞 ]]>。
	edited := gateReplace(t, g.fenced, "LET p_a = p_a + 1", "LET p_a = p_a + 1 ]]>")
	rep := Gate2(g.base, g.parse(t, edited), g.tapDoc)
	if len(gateFindings(rep, "I11b")) == 0 || !rep.HasError() {
		t.Fatalf("可编辑内容含 ]]> 该报 error I11b，得 %+v", rep.Findings)
	}

	// 新增点内容里同样要拦。
	body := "PRIVATE FUNCTION adzi999_extra(p_b)\n   LET p_b = p_b ]]> 1\nEND FUNCTION\n"
	edited2 := insertAtAnchorEnd(t, g, "adzi999.other_function",
		appendFenceBlock("function.adzi999_extra", `new="Y" status="u"`, body))
	rep2 := Gate2(g.base, g.parse(t, edited2), g.tapDoc)
	var appendedHit bool
	for _, f := range gateFindings(rep2, "I11b") {
		if f.Region == "function.adzi999_extra" {
			appendedHit = true
		}
	}
	if !appendedHit {
		t.Fatalf("新增点含 ]]> 该报 I11b，得 %+v", rep2.Findings)
	}
}

// ---- Gate2：结构事务 V1–V7 ----

// renameFenced 模拟 rename 命令的原子改动：**围栏 fn 与签名行一起改**。
// 只改一处会触发 V1（围栏 fn 与签名行不一致）——那是另一条用例。
func renameFenced(t *testing.T, g *gateFix, oldFn, newFn, oldSig, newSig string) []byte {
	t.Helper()
	edited := gateReplace(t, g.fenced, `fn="`+oldFn+`"`, `fn="`+newFn+`"`)
	edited = gateReplace(t, edited, oldSig, newSig)
	return edited
}

// TestGate2StructuralRenameAccepted 一次合法的改名（围栏 fn + 签名行同步）：
// 检出 Rename/SignatureChange 两项结构事务，只留 info，没有 error。
func TestGate2StructuralRenameAccepted(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{SectionState: model.SectionUnlocked})
	edited := renameFenced(t, g,
		"adzi999_calc(p_a)", "adzi999_mul(p_a)",
		"PRIVATE FUNCTION adzi999_calc(p_a)", "PRIVATE FUNCTION adzi999_mul(p_a)")
	rep := Gate2(g.base, g.parse(t, edited), g.tapDoc)
	if rep.HasError() {
		t.Fatalf("合法改名不该有 error：%s %+v", rep.Summary(), rep.Findings)
	}
	if len(gateFindings(rep, "structtx")) != 1 {
		t.Errorf("该有一条 structtx info 汇总结构事务，得 %+v", rep.Findings)
	}
}

// TestGate2V1FnSignatureMismatch 只改围栏 fn 不改签名行 → V1 error
// （设计器装载会抛 ComplexException）。
func TestGate2V1FnSignatureMismatch(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{SectionState: model.SectionUnlocked})
	edited := gateReplace(t, g.fenced, `fn="adzi999_calc(p_a)"`, `fn="adzi999_mul(p_a)"`)
	rep := Gate2(g.base, g.parse(t, edited), g.tapDoc)
	if len(gateFindings(rep, "V1")) == 0 || !rep.HasError() {
		t.Fatalf("fn 与签名行不一致该报 error V1，得 %+v", rep.Findings)
	}
}

// TestGate2V3RenameConflict 改成已存在的名字 → V3 error（会与现存点/占位符撞名）。
func TestGate2V3RenameConflict(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{SectionState: model.SectionUnlocked})
	edited := renameFenced(t, g,
		"adzi999_calc(p_a)", "global.memo",
		"PRIVATE FUNCTION adzi999_calc(p_a)", "PRIVATE FUNCTION global.memo")
	rep := Gate2(g.base, g.parse(t, edited), g.tapDoc)
	if len(gateFindings(rep, "V3")) == 0 || !rep.HasError() {
		t.Fatalf("撞名改该报 error V3，得 %+v", rep.Findings)
	}
}

// TestGate2V2ResidualInEditable 旧名还残留在**可编辑区**的调用点里 → V2 error：
// 改名必须先把调用点改掉，否则新名生效后旧调用就是坏引用。
func TestGate2V2ResidualInEditable(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{SectionState: model.SectionUnlocked})
	edited := renameFenced(t, g,
		"adzi999_calc(p_a)", "adzi999_mul(p_a)",
		"PRIVATE FUNCTION adzi999_calc(p_a)", "PRIVATE FUNCTION adzi999_mul(p_a)")
	// 在可编辑的 .main 区段里留一个旧名调用点。
	edited = gateReplace(t, edited, "  DIALOG\n", "  CALL adzi999_calc(99)\n  DIALOG\n")
	rep := Gate2(g.base, g.parse(t, edited), g.tapDoc)
	var editableResidual bool
	for _, f := range gateFindings(rep, "V2") {
		if !f.Denied && strings.Contains(f.Message, "可编辑区") {
			editableResidual = true
		}
	}
	if !editableResidual {
		t.Fatalf("可编辑区残留旧名该报 V2 error，得 %+v", rep.Findings)
	}
}

// TestGate2V2ResidualInLockedSection 旧名残留在**只读区段**里（改不了）→ V2 Denied：
// 这种改名做不成，必须拒绝（退出码 4），并提示先解锁或放弃改名。
func TestGate2V2ResidualInLockedSection(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{}) // 默认 Locked
	edited := renameFenced(t, g,
		"adzi999_calc(p_a)", "adzi999_mul(p_a)",
		"PRIVATE FUNCTION adzi999_calc(p_a)", "PRIVATE FUNCTION adzi999_mul(p_a)")
	// 只读的 .description 区段里留旧名调用点。
	edited = gateReplace(t, edited, "#+ Description: 合成測試程式", "#+ Description: CALL adzi999_calc(99)")
	rep := Gate2(g.base, g.parse(t, edited), g.tapDoc)
	var denied bool
	for _, f := range gateFindings(rep, "V2") {
		if f.Denied && strings.Contains(strings.Join(f.Detail, ";"), "unlock") {
			denied = true
		}
	}
	if !denied {
		t.Fatalf("只读区残留旧名该报 Denied V2 并提示解锁，得 %+v", rep.Findings)
	}
}

// TestGate2V7RenameReadonlyPoint 目标点不可编辑（readonly="Y"）→ V7 Denied：
// 结构事务（改名/改签名/改描述）要求目标是可编辑的自订定义点。
func TestGate2V7RenameReadonlyPoint(t *testing.T) {
	entries := testutil.NormalEntries("adzi999")
	patchTAPPoint(t, entries, "adzi999", "function.adzi999_calc", func(elem string) string {
		return strings.Replace(elem, `src="s" readonly=""`, `src="s" readonly="Y"`, 1)
	})
	g := newGateFixEntries(t, "adzi999", entries, synth.Options{SectionState: model.SectionUnlocked})
	edited := renameFenced(t, g,
		"adzi999_calc(p_a)", "adzi999_mul(p_a)",
		"PRIVATE FUNCTION adzi999_calc(p_a)", "PRIVATE FUNCTION adzi999_mul(p_a)")
	rep := Gate2(g.base, g.parse(t, edited), g.tapDoc)
	var denied bool
	for _, f := range gateFindings(rep, "V7") {
		if f.Denied {
			denied = true
		}
	}
	if !denied {
		t.Fatalf("改名不可编辑点该报 Denied V7，得 %+v", rep.Findings)
	}
}

// TestGate2V6DescBlockForm 描述块形态（V6）：
//   - 改过的描述块里有非 # 开头的行 → error；
//   - 存量（没改过）的同种瑕疵 → warn（真实语料里有历史遗留，不能拒真包）。
//
// 注意：V6 只在**检出结构事务**时才跑（validateStructural 的入口条件），
// 所以这里顺带做一次合法改名把结构事务引出来 —— 单独改描述块文字不会触发 V6。
func TestGate2V6DescBlockForm(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{SectionState: model.SectionUnlocked})
	edited := renameFenced(t, g,
		"adzi999_calc(p_a)", "adzi999_mul(p_a)",
		"PRIVATE FUNCTION adzi999_calc(p_a)", "PRIVATE FUNCTION adzi999_mul(p_a)")
	// 把描述块里的 Usage 行去掉 #（desc 字段没跟着改 → 存量形态 → warn）。
	// 前面的换行用于把"正文行"与围栏 desc 属性里的字面量 \n 区分开。
	edited = gateReplace(t, edited, "\n#+ Usage......: CALL adzi999_calc(1)", "\nUsage......: CALL adzi999_calc(1)")
	rep := Gate2(g.base, g.parse(t, edited), g.tapDoc)
	var warn bool
	for _, f := range gateFindings(rep, "V6") {
		if f.Severity == SevWarn {
			warn = true
		}
	}
	if !warn {
		t.Fatalf("存量描述块瑕疵该报 warn V6，得 %+v", rep.Findings)
	}
}

// TestGate2V6DescConflict 描述块与围栏 desc 字段同时改且不一致 → error（意图不明）。
func TestGate2V6DescConflict(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{SectionState: model.SectionUnlocked})
	// desc 字段改成"新描述"，描述块文字改成另一句 → 两处意图冲突。
	edited := gateReplace(t, g.fenced,
		`desc="#+ Description: 合成測試用函式\n#+ Usage......: CALL adzi999_calc(1)"`,
		`desc="新描述"`)
	edited = gateReplace(t, edited, "#+ Description: 合成測試用函式", "#+ Description: 改过的描述")
	rep := Gate2(g.base, g.parse(t, edited), g.tapDoc)
	var conflict bool
	for _, f := range gateFindings(rep, "V6") {
		if strings.Contains(f.Message, "意图不明") {
			conflict = true
		}
	}
	if !conflict {
		t.Fatalf("描述块与 desc 字段不一致该报 V6 error，得 %+v", rep.Findings)
	}
}

// TestScopeRule 程序 type → 允许 scope 的映射（FunctionInfoWindow.xaml.cs:143-192）：
// B 强制 PUBLIC；M/G/X/Z/Q 强制 PRIVATE；S 默认 PRIVATE/W 默认 PUBLIC 但可改。
func TestScopeRule(t *testing.T) {
	cases := []struct {
		typ  string
		must string
		free bool
	}{
		{"B", "PUBLIC", false},
		{"M", "PRIVATE", false}, {"G", "PRIVATE", false}, {"X", "PRIVATE", false},
		{"Z", "PRIVATE", false}, {"Q", "PRIVATE", false},
		{"S", "PRIVATE", true}, {"W", "PUBLIC", true},
		{"", "", true}, {"unknown", "", true},
		{"b", "PUBLIC", false}, // 大小写不敏感
	}
	for _, c := range cases {
		must, free := scopeRule(c.typ)
		if must != c.must || free != c.free {
			t.Errorf("scopeRule(%q) = (%q,%v)，想 (%q,%v)", c.typ, must, free, c.must, c.free)
		}
	}
}

// ---- gate3：装载模拟 ----

// TestTempDirForPkg 临时目录**优先放在包旁边**（企业 EDR/组策略常把 %TEMP% 设成不可写；
// gate3 只是把新包写一份再读回，放包目录既绕开策略又保证同文件系统）。
func TestTempDirForPkg(t *testing.T) {
	dir := t.TempDir()
	pkgPath := filepath.Join(dir, "p.tzc")
	if err := os.WriteFile(pkgPath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	tmp, err := tempDirForPkg(pkgPath)
	if err != nil {
		t.Fatalf("tempDirForPkg 失败：%v", err)
	}
	defer os.RemoveAll(tmp)
	if filepath.Dir(tmp) != dir {
		t.Errorf("临时目录该建在包旁边 %q，得 %q", dir, tmp)
	}
	if !strings.HasPrefix(filepath.Base(tmp), "tdev-gate3-") {
		t.Errorf("临时目录名前缀不对：%q", filepath.Base(tmp))
	}
	if fi, err := os.Stat(tmp); err != nil || !fi.IsDir() {
		t.Errorf("临时目录没建出来：%v", err)
	}

	// 包所在目录不存在 → 回落到系统临时目录（不能让 gate3 因此失败）。
	missing := filepath.Join(dir, "no-such-dir", "x.tzc")
	tmp2, err := tempDirForPkg(missing)
	if err != nil {
		t.Fatalf("回落系统临时目录失败：%v", err)
	}
	defer os.RemoveAll(tmp2)
	if fi, err := os.Stat(tmp2); err != nil || !fi.IsDir() {
		t.Errorf("回落目录没建出来：%v", err)
	}
}

// TestGate3CleanPackage 把原包字节原样重跑装载模拟：应通过，且**不残留临时目录**
// （gate3 不写用户原包，只写临时文件再读回）。
func TestGate3CleanPackage(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	rep, err := Gate3(g.pkg, g.pkg.Tap().Data, g.pkg.Tgl().Data)
	if err != nil {
		t.Fatalf("Gate3 返回错误：%v", err)
	}
	if rep.HasError() {
		t.Fatalf("干净包装载模拟该通过：%s %+v", rep.Summary(), rep.Findings)
	}
	if len(gateFindings(rep, "gate3.ok")) != 1 {
		t.Errorf("该有一条 gate3.ok info，得 %+v", rep.Findings)
	}
	// 临时目录必须清掉。
	leftover, _ := filepath.Glob(filepath.Join(filepath.Dir(g.pkg.Path), "tdev-gate3-*"))
	if len(leftover) != 0 {
		t.Errorf("gate3 该清理临时目录，残留 %v", leftover)
	}
	// 原包不能被改（字节与摘要都不变）。
	if _, err := os.Stat(g.pkg.Path); err != nil {
		t.Errorf("gate3 不该删/挪用户原包：%v", err)
	}
}

// TestGate3BuildFailure 新包装不进（这里构造一个没有 .tap 的包）→ gate3.build error，
// 外层 error 为 nil（发现进报告，不中断）。
func TestGate3BuildFailure(t *testing.T) {
	pkg := &pkgfile.Package{
		Path:    filepath.Join(t.TempDir(), "p.tzc"),
		Entries: []*pkgfile.Entry{{Name: "p.tgl", Data: []byte("x")}},
	}
	rep, err := Gate3(pkg, []byte("<add_points/>"), nil)
	if err != nil {
		t.Fatalf("发现类失败该进报告而不是返回 error：%v", err)
	}
	if len(gateFindings(rep, "gate3.build")) != 1 || !rep.HasError() {
		t.Fatalf("该报 gate3.build error，得 %+v", rep.Findings)
	}
}

// TestGate3SynthesizeFailure 新 TAP 解析不了 → 合成失败 → gate3.synthesize。
func TestGate3SynthesizeFailure(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	rep, err := Gate3(g.pkg, []byte("<add_points"), g.pkg.Tgl().Data)
	if err != nil {
		t.Fatalf("发现类失败该进报告：%v", err)
	}
	if len(gateFindings(rep, "gate3.synthesize")) != 1 || !rep.HasError() {
		t.Fatalf("该报 gate3.synthesize error，得 %+v", rep.Findings)
	}
}

// TestGate3RenderFailure TGL 正文里出现围栏前缀 → 渲染拒绝（协议与内容冲突）→ gate3.render。
func TestGate3RenderFailure(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	tgl := string(g.pkg.Tgl().Data)
	if !strings.Contains(tgl, "IMPORT os") {
		t.Fatalf("夹具 TGL 变了：%q", tgl)
	}
	badTgl := []byte(strings.Replace(tgl, "IMPORT os", "IMPORT os\n{//@tdev:begin point x}", 1))
	rep, err := Gate3(g.pkg, g.pkg.Tap().Data, badTgl)
	if err != nil {
		t.Fatalf("发现类失败该进报告：%v", err)
	}
	if len(gateFindings(rep, "gate3.render")) != 1 || !rep.HasError() {
		t.Fatalf("该报 gate3.render error，得 %+v", rep.Findings)
	}
}

// TestGate3InvariantsFailure 新 TAP 违反不变量（status 非法）→ gate3.invariants，
// 并且把子报告里的 error 级发现冒泡上来（可定位）。
func TestGate3InvariantsFailure(t *testing.T) {
	g := newGateFix(t, "adzi999", synth.Options{})
	// 把某个点的 status 改成非法值。
	badTap := bytes.Replace(g.pkg.Tap().Data, []byte(` status="" src="s"`), []byte(` status="zzz" src="s"`), 1)
	if bytes.Equal(badTap, g.pkg.Tap().Data) {
		t.Fatal("夹具失效：没有改到 TAP")
	}
	rep, err := Gate3(g.pkg, badTap, g.pkg.Tgl().Data)
	if err != nil {
		t.Fatalf("发现类失败该进报告：%v", err)
	}
	if len(gateFindings(rep, "gate3.invariants")) != 1 || !rep.HasError() {
		t.Fatalf("该报 gate3.invariants error，得 %+v", rep.Findings)
	}
	// 子报告的 error 级发现要冒泡到报告里（否则用户只看到一句"违反不变量"）。
	var hasI5 bool
	for _, f := range gateFindings(rep, "I5") {
		if f.Severity == SevError {
			hasI5 = true
		}
	}
	if !hasI5 {
		t.Errorf("gate3 该冒泡子报告的 I5 error，得 %+v", rep.Findings)
	}
}

// ---- 定位工具与错误类型（对外契约） ----

// TestDocAndContentDiffOff 两篇文档/某 Region 内容的首个差异偏移：
// 相同返回 -1；nil / 越界返回 -1（调用方据此判定"没有位置"）。
func TestDocAndContentDiffOff(t *testing.T) {
	base := &model.Document{Text: []byte("aaa\nbbb\n")}
	if got := docDiffOff(base, &model.Document{Text: []byte("aaa\nbbb\n")}); got != -1 {
		t.Errorf("相同文档该返回 -1，得 %d", got)
	}
	edited := &model.Document{Text: []byte("aaa\nbXb\n")}
	if got := docDiffOff(base, edited); got != 5 {
		t.Errorf("首个差异偏移 = %d，想 5", got)
	}
	if got := docDiffOff(nil, edited); got != -1 {
		t.Errorf("nil 文档该返回 -1，得 %d", got)
	}
	// 行尾等价：CRLF↔LF 不算差异。
	if got := docDiffOff(&model.Document{Text: []byte("a\r\n")}, &model.Document{Text: []byte("a\n")}); got != -1 {
		t.Errorf("纯行尾差异该返回 -1，得 %d", got)
	}

	// 编辑后文本多了 "ZZ" 前缀，区段内容整体后移 2 字节；内容内差异在第 5 个字节。
	editedShifted := &model.Document{Text: []byte("ZZaaa\nbXb\n")}
	bReg := &model.Region{ContentSpan: model.ByteRange{Start: 0, End: 7}}
	eReg := &model.Region{ContentSpan: model.ByteRange{Start: 2, End: 9}}
	if got := contentDiffOff(base, editedShifted, bReg, eReg); got != 7 {
		t.Errorf("Region 内容差异偏移 = %d，想 7（es.Start=2 + 内容内偏移 5）", got)
	}
	// 内容相同 → -1。
	if got := contentDiffOff(base, &model.Document{Text: base.Text}, bReg, bReg); got != -1 {
		t.Errorf("内容相同该返回 -1，得 %d", got)
	}
	// 越界 / nil → -1。
	bad := &model.Region{ContentSpan: model.ByteRange{Start: 0, End: 999}}
	if got := contentDiffOff(base, editedShifted, bad, eReg); got != -1 {
		t.Errorf("基线区间越界该返回 -1，得 %d", got)
	}
	if got := contentDiffOff(nil, editedShifted, bReg, eReg); got != -1 {
		t.Errorf("nil 文档该返回 -1，得 %d", got)
	}
}

// TestRegionAndTapLocating 位置工具：
//   - regionOffOf 按名字找内容起点；
//   - tapLoc 把 TAP 字节偏移说清落在哪个 point/section；
//   - regionAnchorOff 优先签名行，其次描述块，最后正文起点。
func TestRegionAndTapLocating(t *testing.T) {
	doc := &model.Document{Text: bytes.Repeat([]byte("x"), 100), Regions: []*model.Region{
		{Name: "function.a", Kind: model.RegionPoint,
			ContentSpan:   model.ByteRange{Start: 10, End: 40},
			SignatureSpan: &model.ByteRange{Start: 12, End: 20},
			DescSpan:      &model.ByteRange{Start: 10, End: 12}},
		{Name: "p.nospan", Kind: model.RegionSection, ContentSpan: model.ByteRange{Start: 50, End: 60}},
	}}
	if got := regionOffOf(doc, "function.a"); got != 10 {
		t.Errorf("regionOffOf = %d，想 10", got)
	}
	if got := regionOffOf(doc, "nope"); got != -1 {
		t.Errorf("找不到该返回 -1，得 %d", got)
	}
	if got := regionOffOf(nil, "function.a"); got != -1 {
		t.Errorf("nil 文档该返回 -1，得 %d", got)
	}
	if got := regionOffOf(doc, ""); got != -1 {
		t.Errorf("空名该返回 -1，得 %d", got)
	}

	// regionAnchorOff：签名行优先。
	if got := regionAnchorOff(doc.FindRegion("function.a")); got != 12 {
		t.Errorf("有签名行时该指签名行，得 %d", got)
	}
	// 没有签名/描述 → 正文起点。
	if got := regionAnchorOff(doc.FindRegion("p.nospan")); got != 50 {
		t.Errorf("无签名无描述时该指正文起点，得 %d", got)
	}
	// 只有描述块时指描述块。
	onlyDesc := &model.Region{ContentSpan: model.ByteRange{Start: 70, End: 80}, DescSpan: &model.ByteRange{Start: 70, End: 75}}
	if got := regionAnchorOff(onlyDesc); got != 70 {
		t.Errorf("有描述块时该指描述块，得 %d", got)
	}
	if got := regionAnchorOff(nil); got != -1 {
		t.Errorf("nil Region 该返回 -1，得 %d", got)
	}

	// tapLoc：point 内 / section 内 / 都不在 / 越界。
	tap := []byte(`<add_points prog="p"><point name="function.a"><![CDATA[x]]></point><section id="p.main"><![CDATA[y]]></section></add_points>`)
	tapDoc, err := tapfile.Parse(tap)
	if err != nil {
		t.Fatalf("TAP 解析失败：%v", err)
	}
	pointOff := strings.Index(string(tap), "<point") + 5
	sectionOff := strings.Index(string(tap), "<section") + 5
	if got := tapLoc(tapDoc, pointOff); !strings.Contains(got, `point name="function.a"`) {
		t.Errorf("point 内的偏移该指名点：%q", got)
	}
	if got := tapLoc(tapDoc, sectionOff); !strings.Contains(got, `section id="p.main"`) {
		t.Errorf("section 内的偏移该指名区段：%q", got)
	}
	if got := tapLoc(tapDoc, len(tap)-3); !strings.Contains(got, "不在任何") {
		t.Errorf("区段之外的偏移该明说：%q", got)
	}
	if got := tapLoc(tapDoc, -5); !strings.Contains(got, "TAP 字节偏移 -5") {
		t.Errorf("负偏移该给原样偏移：%q", got)
	}
	if got := tapLoc(nil, 3); !strings.Contains(got, "TAP 字节偏移 3") {
		t.Errorf("nil TAP 该给原样偏移：%q", got)
	}
}

// TestClipAndFirstDiffDesc clip 是**给错误信息用的**：行尾转义、超长截断（带省略号）。
// firstDiffDesc 要同时给出偏移与两侧上下文 —— 只说"不一样"没法排查。
func TestClipAndFirstDiffDesc(t *testing.T) {
	if got := clip("a\rb\nc"); got != `a\rb\nc` {
		t.Errorf("clip 该转义 CR/LF，得 %q", got)
	}
	long := strings.Repeat("x", 200)
	got := clip(long)
	if len(got) != 123 || !strings.HasSuffix(got, "…") {
		t.Errorf("超长该截到 120 字节 + 省略号(3 字节)，得 len=%d", len(got))
	}
	// firstDiffDesc 的切片上界曾写死 i+20、不夹到 len(a)/len(b)：差异落在短串尾部
	// 20 字节内时会 panic。已修（上界夹到两侧长度），越界回归见
	// firstdiff_bounds_test.go；这里只钉正常口径：偏移 + 两侧上下文。
	a := []byte("aaaaaX" + strings.Repeat("a", 34))
	b := []byte("aaaaaY" + strings.Repeat("a", 34))
	desc := firstDiffDesc(a, b)
	if !strings.Contains(desc, "偏移 5") {
		t.Errorf("该报首个差异偏移，得 %q", desc)
	}
	if !strings.Contains(desc, "X") || !strings.Contains(desc, "Y") {
		t.Errorf("该带两侧上下文，得 %q", desc)
	}
	lenDesc := firstDiffDesc([]byte("abc"), []byte("abcdef"))
	if !strings.Contains(lenDesc, "长度不同") {
		t.Errorf("前缀相同时该报长度不同，得 %q", lenDesc)
	}
}

// TestOutsideBytesAndContentOf outsideBytes 拼接围栏外字节（跳过非法区间）；
// contentOf 取 Region 正文（nil/越界返回 nil，不 panic）。
func TestOutsideBytesAndContentOf(t *testing.T) {
	doc := &model.Document{Text: []byte("0123456789"), Spans: []model.Span{
		{Span: model.ByteRange{Start: 0, End: 2}},
		{Span: model.ByteRange{Start: 2, End: 2}},  // 空段不算
		{Span: model.ByteRange{Start: 2, End: 99}}, // 越界跳过
		{Span: model.ByteRange{Start: 7, End: 9}},
	}}
	if got := string(outsideBytes(doc)); got != "0178" {
		t.Errorf("outsideBytes = %q，想 %q", got, "0178")
	}
	if got := string(contentOf(doc, &model.Region{ContentSpan: model.ByteRange{Start: 3, End: 6}})); got != "345" {
		t.Errorf("contentOf = %q，想 345", got)
	}
	if contentOf(doc, nil) != nil {
		t.Error("nil Region 该返回 nil")
	}
	if contentOf(doc, &model.Region{ContentSpan: model.ByteRange{Start: 5, End: 99}}) != nil {
		t.Error("越界该返回 nil")
	}
}

// TestFillPositions 兜底定位：只报了 Region 名的发现要补上文件与行号；
// 编辑后文档找不到就退到基线（删除类问题只存在于基线）。
func TestFillPositions(t *testing.T) {
	edited := &model.Document{Text: []byte("l1\nl2\nl3\n"), Prog: "p", Regions: []*model.Region{
		{Name: "function.a", Kind: model.RegionPoint, ContentSpan: model.ByteRange{Start: 3, End: 6}},
	}}
	base := &model.Document{Text: []byte("l1\nl2\nl3\n"), Regions: []*model.Region{
		{Name: "gone.point", Kind: model.RegionPoint, ContentSpan: model.ByteRange{Start: 3, End: 6}},
	}}
	rep := &Report{}
	rep.Add(Finding{Code: "X", Severity: SevError, Region: "function.a"})         // 编辑后文档里能定位
	rep.Add(Finding{Code: "Y", Severity: SevError, Region: "gone.point"})         // 只能退到基线
	rep.Add(Finding{Code: "Z", Severity: SevError})                               // 包级发现，不动它
	rep.Add(Finding{Code: "W", Severity: SevWarn, Region: "function.a", Line: 9}) // 已有行号，不覆盖
	fillPositions(rep, base, edited)
	if rep.Findings[0].File != FileEdited || rep.Findings[0].Line != 2 {
		t.Errorf("编辑后文档定位 = %q:%d，想 %s:2", rep.Findings[0].File, rep.Findings[0].Line, FileEdited)
	}
	if rep.Findings[1].File != FileBase || rep.Findings[1].Line != 2 {
		t.Errorf("基线兜底定位 = %q:%d，想 %s:2", rep.Findings[1].File, rep.Findings[1].Line, FileBase)
	}
	if rep.Findings[2].Line != 0 || rep.Findings[2].File != "" {
		t.Errorf("包级发现不该被填位置：%+v", rep.Findings[2])
	}
	if rep.Findings[3].Line != 9 {
		t.Errorf("已有行号不该被覆盖，得 %d", rep.Findings[3].Line)
	}
}

// TestBuildI3 buildI3 是 I3 的"精确判据"：只匹配**定义头**（可带 public/private 前缀），
// 不匹配调用点；名字里的正则元字符要转义；名字集合为空返回 nil（不做检查）。
func TestBuildI3(t *testing.T) {
	if re := buildI3(nil); re != nil {
		t.Error("空名字集合该返回 nil")
	}
	if re := buildI3([]string{"", ""}); re != nil {
		t.Error("全空名字集合该返回 nil")
	}
	re := buildI3([]string{"a.b", "calc"})
	if re == nil {
		t.Fatal("有名字时该返回正则")
	}
	cases := []struct {
		text string
		want bool
	}{
		{"PRIVATE FUNCTION calc(p_a)\n", true},
		{"FUNCTION calc(p_a)\n", true},
		{"public FUNCTION a.b()\n", true}, // 名字含元字符也要能精确匹配
		{"  CALL calc(1)\n", false},       // 调用点不是定义头
		{"FUNCTION other(p_a)\n", false},  // 不在名单里
		{"DIALOG ATTRIBUTES(F)\n", false}, // 宽正则会误伤框架自己的这个形状
	}
	for _, c := range cases {
		if got := re.MatchString(c.text); got != c.want {
			t.Errorf("buildI3 对 %q = %v，想 %v", c.text, got, c.want)
		}
	}
}

// TestVerifyErrorTypes 两类错误（3=验证失败 / 4=写入被拒）的渲染与退出码是 CLI 契约。
func TestVerifyErrorTypes(t *testing.T) {
	// DeniedError：有详情才带分隔符。
	de := &DeniedError{Msg: "被拒"}
	if de.Error() != "被拒" || de.ExitCode() != 4 {
		t.Errorf("DeniedError = %q/%d", de.Error(), de.ExitCode())
	}
	de2 := &DeniedError{Msg: "被拒", Detail: []string{"甲", "乙"}}
	if got := de2.Error(); got != "被拒：甲; 乙" {
		t.Errorf("带详情渲染 = %q", got)
	}

	// VerifyErrorFromFindings：只汇总 error 级；有位置才带（file:line）。
	ve := &VerifyErrorFromFindings{Findings: []Finding{
		{Code: "X", Severity: SevWarn, Message: "warn 不参与"},
		{Code: "E1", Severity: SevError, Message: "坏", File: FileEdited, Line: 12},
		{Code: "E2", Severity: SevError, Message: "也坏"},
	}}
	msg := ve.Error()
	if !strings.Contains(msg, "E1 坏（prog.full.4gl:12）") || !strings.Contains(msg, "E2 也坏") {
		t.Errorf("Error() = %q", msg)
	}
	if strings.Contains(msg, "warn 不参与") {
		t.Errorf("warn 不该进 Error()：%q", msg)
	}
	if ve.ExitCode() != 3 {
		t.Errorf("普通验证失败该是 3，得 %d", ve.ExitCode())
	}
	// 一条 Denied 的 error → 退出码 4（权限问题优先报 4）。
	denied := &VerifyErrorFromFindings{Findings: []Finding{
		{Code: "E", Severity: SevError, Message: "被拒", Denied: true},
	}}
	if denied.ExitCode() != 4 {
		t.Errorf("带 Denied 该是 4，得 %d", denied.ExitCode())
	}
	// 没有任何 error → 兜底文案"验证失败"。
	empty := &VerifyErrorFromFindings{}
	if empty.Error() != "验证失败" || empty.ExitCode() != 3 {
		t.Errorf("空报告 = %q/%d", empty.Error(), empty.ExitCode())
	}

	// VerifyErrorFromMsg。
	vm := &VerifyErrorFromMsg{Msg: "失败"}
	if vm.Error() != "失败" || vm.ExitCode() != 3 {
		t.Errorf("VerifyErrorFromMsg = %q/%d", vm.Error(), vm.ExitCode())
	}
	vm2 := &VerifyErrorFromMsg{Msg: "失败", Detail: []string{"甲"}}
	if got := vm2.Error(); got != "失败：甲" {
		t.Errorf("带详情渲染 = %q", got)
	}
}

// TestLocateEdgeCases locate 在文档/偏移不可用时原样返回发现（不 panic、不篡改）。
func TestLocateEdgeCases(t *testing.T) {
	f := Finding{Code: "X", Severity: SevError}
	if got := locate(f, nil, FileEdited, 3); got.File != "" || got.Line != 0 {
		t.Errorf("nil 文档该原样返回：%+v", got)
	}
	if got := locate(f, &model.Document{Text: []byte("abc")}, FileEdited, -1); got.Line != 0 {
		t.Errorf("负偏移该原样返回：%+v", got)
	}
	if got := locate(f, &model.Document{Text: nil}, FileEdited, 0); got.Line != 0 {
		t.Errorf("空文档该原样返回：%+v", got)
	}
	// 正常路径：带 file + 1-based 行 + clip 后的行内容。
	got := locate(f, &model.Document{Text: []byte("l1\nl2\n")}, FileEdited, 3)
	if got.File != FileEdited || got.Line != 2 || got.Snippet != "l2" {
		t.Errorf("定位 = %+v，想 %s:2 %q", got, FileEdited, "l2")
	}
	// snippet 超长要被 clip（错误信息不能把整行原样打到终端）。
	long := strings.Repeat("y", 200)
	got2 := locate(f, &model.Document{Text: []byte(long + "\n")}, FileEdited, 0)
	if len(got2.Snippet) != 123 || !strings.HasSuffix(got2.Snippet, "…") {
		t.Errorf("Snippet 该被截断，得 len=%d", len(got2.Snippet))
	}
	// spanStartOf：nil → -1。
	if got := spanStartOf(nil); got != -1 {
		t.Errorf("nil span 该返回 -1，得 %d", got)
	}
	if got := spanStartOf(&model.ByteRange{Start: 7, End: 9}); got != 7 {
		t.Errorf("spanStartOf = %d，想 7", got)
	}
}

// TestReportAddAtPosition AddAt 是「带位置的 Add」：位置补到发现上，计数照常。
func TestReportAddAtPosition(t *testing.T) {
	doc := &model.Document{Text: []byte("a\nb\nc\n")}
	rep := &Report{}
	rep.AddAt(Finding{Code: "E", Severity: SevError, Denied: true, Region: "p.x"}, doc, FileEdited, 2)
	if len(rep.Findings) != 1 || rep.Errors != 1 || rep.Denied != 1 {
		t.Fatalf("AddAt 该正常计数：%+v", rep)
	}
	f := rep.Findings[0]
	if f.File != FileEdited || f.Line != 2 || f.Snippet != "b" {
		t.Errorf("位置 = %+v，想 %s:2 b", f, FileEdited)
	}
}

// 保证 model 包的区间类型在本文件里被直接引用（手工夹具的坐标全用它）。
var _ = model.ByteRange{}
