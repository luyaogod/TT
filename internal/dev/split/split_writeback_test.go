package split

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/fence"
	"github.com/luyaogod/TT/internal/dev/model"
	"github.com/luyaogod/TT/internal/dev/pkgfile"
	"github.com/luyaogod/TT/internal/dev/synth"
	"github.com/luyaogod/TT/internal/dev/tapfile"
	"github.com/luyaogod/TT/internal/dev/testutil"
)

// 本文件补的是 Split / foldSection / newPointAttrs 这条**写回主链**。
//
// 为什么必须在这里测：README 说它"故意不覆盖 Split，交给端到端" —— 但端到端
// 只会告诉你「整条链最后不对」，定位不到是授权判定、order 分配、还是 TglTag 折叠。
// 写回层的红线（只读区不可改、非 new="Y" 不可删、status 写 u 不写 c、
// 区段正文必须折回占位符原文）每一条都值得一个能指名道姓的用例。
//
// 夹具走**合成包**（不依赖真实语料）：testutil 造 .tzc → pkgfile 打开 →
// synth 合成 → fence 渲染 → 对渲染文本"编辑"→ fence 解析 → Split。
// 这正是生产链 apply 的输入形状（cli/dev/tzc 的 cli.go 同款组装）。

// splitFixture 是「基线 + 围栏文本 + 包」三件套。
type splitFixture struct {
	base   *model.Document
	fenced []byte
	pkg    *pkgfile.Package
}

// newSplitFixture 造正常合成包并合成/渲染/解析出基线。
func newSplitFixture(t *testing.T, prog string, opts synth.Options) *splitFixture {
	t.Helper()
	return newSplitFixtureFromEntries(t, prog, testutil.NormalEntries(prog), opts)
}

// newSplitFixtureFromEntries 允许用例改条目（改 TAP 根属性 / 点的属性 / 删点）。
func newSplitFixtureFromEntries(t *testing.T, prog string, entries map[string]string, opts synth.Options) *splitFixture {
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
	// 基线必须是"把基线自己的围栏文本解析一遍"的结果（坐标在围栏文本上），
	// 不能直接用 synth 产物（那是未加围栏的坐标）。
	baseline, err := fence.Parse(doc0, fenced)
	if err != nil {
		t.Fatalf("基线自解析失败：%v", err)
	}
	return &splitFixture{base: baseline.Doc, fenced: fenced, pkg: pkg}
}

// reparse 用基线 + 编辑后的围栏文本得到 ParseResult（Split 的第二个入参）。
func (f *splitFixture) reparse(t *testing.T, edited []byte) *fence.ParseResult {
	t.Helper()
	parsed, err := fence.Parse(f.base, edited)
	if err != nil {
		t.Fatalf("编辑后的围栏文本解析失败：%v", err)
	}
	return parsed
}

// replaceOnce 在围栏文本里替换唯一一段内容；没命中直接 Fatal（防止夹具悄悄失效）。
func replaceOnce(t *testing.T, text []byte, old, new string) []byte {
	t.Helper()
	if n := bytes.Count(text, []byte(old)); n != 1 {
		t.Fatalf("夹具失效：%q 在围栏文本里出现 %d 次（想恰好 1 次）", old, n)
	}
	return bytes.Replace(text, []byte(old), []byte(new), 1)
}

// regionOf 按名字取基线区间（含嵌套）。
func regionOf(t *testing.T, doc *model.Document, name string) *model.Region {
	t.Helper()
	r := doc.FindRegion(name)
	if r == nil {
		t.Fatalf("基线里没有区间 %q", name)
	}
	return r
}

// splice 用 repl 替换 [start,end) 并把文本返回（编辑后文本独立分配）。
func splice(text []byte, start, end int, repl string) []byte {
	out := make([]byte, 0, len(text)+len(repl)-(end-start))
	out = append(out, text[:start]...)
	out = append(out, repl...)
	out = append(out, text[end:]...)
	return out
}

// deleteRegion 删掉一个围栏块的整行（begin..end）及其行尾 —— 这就是设计器里
// "选中函数块按 Delete" 的等价操作。
func deleteRegion(t *testing.T, f *splitFixture, name string) []byte {
	t.Helper()
	r := regionOf(t, f.base, name)
	end := r.FullSpan.End
	switch {
	case end+1 < len(f.fenced) && f.fenced[end] == '\r' && f.fenced[end+1] == '\n':
		end += 2
	case end < len(f.fenced) && f.fenced[end] == '\n':
		end++
	}
	return splice(f.fenced, r.FullSpan.Start, end, "")
}

// opSetPointAttr 在计划里找 SetPointAttr，返回最后设置的值。
func opSetPointAttr(plan *Plan, name, key string) (string, bool) {
	v, found := "", false
	for _, op := range plan.Ops {
		if a, ok := op.(tapfile.SetPointAttr); ok && a.Name == name && a.Key == key {
			v, found = a.Value, true
		}
	}
	return v, found
}

// opSetRootAttr 同理。
func opSetRootAttr(plan *Plan, key string) (string, bool) {
	v, found := "", false
	for _, op := range plan.Ops {
		if a, ok := op.(tapfile.SetRootAttr); ok && a.Key == key {
			v, found = a.Value, true
		}
	}
	return v, found
}

// findAddPoint 找针对某点名的新增操作。
func findAddPoint(plan *Plan, name string) (tapfile.AddPoint, bool) {
	for _, op := range plan.Ops {
		if a, ok := op.(tapfile.AddPoint); ok && a.Attrs["name"] == name {
			return a, true
		}
	}
	return tapfile.AddPoint{}, false
}

// applyAndParse 把计划真的打进 TAP 再解析回来 —— 断言"落盘形态"而不只是"计划里有这个 op"。
func applyAndParse(t *testing.T, f *splitFixture, plan *Plan) *tapfile.Doc {
	t.Helper()
	newTap, err := tapfile.Rewrite(f.pkg.Tap().Data, plan.Ops...)
	if err != nil {
		t.Fatalf("把计划打进 TAP 失败：%v", err)
	}
	doc, err := tapfile.Parse(newTap)
	if err != nil {
		t.Fatalf("改写后的 TAP 解析失败：%v", err)
	}
	return doc
}

// ---- 改动自订定义点 ----

// TestSplitEditsPointBodyWritesStatusU 改自订定义点正文：
//   - 产出 SetPointCDATA（新正文）+ SetPointAttr status="u"；
//   - **status 必须是 "u" 不是 "c"**（红线 D-1：写 c 的点会被设计器下次保存静默丢弃，
//     因为 AddPointModel.ToXML 对 CREATE 直接 return null）；
//   - 真打进 TAP 后重新解析，新内容与 status 都要在。
func TestSplitEditsPointBodyWritesStatusU(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	const name = "function.adzi999_calc"
	edited := replaceOnce(t, f.fenced, "LET p_a = p_a + 1", "LET p_a = p_a + 2")
	plan, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err != nil {
		t.Fatalf("改自订点正文该成功，得 %v", err)
	}

	if len(plan.Changed) != 1 || plan.Changed[0] != name {
		t.Errorf("Changed = %v，想恰好 [%s]", plan.Changed, name)
	}
	if len(plan.Added) != 0 || len(plan.Deleted) != 0 || len(plan.Renamed) != 0 {
		t.Errorf("只改正文不该有增删改名：%+v", plan)
	}
	// 正文补丁内容：包含新字节、不含旧字节。
	var cdata []byte
	for _, op := range plan.Ops {
		if s, ok := op.(tapfile.SetPointCDATA); ok && s.Name == name {
			cdata = s.Content
		}
	}
	if cdata == nil {
		t.Fatal("没有产出 SetPointCDATA")
	}
	if !bytes.Contains(cdata, []byte("p_a + 2")) || bytes.Contains(cdata, []byte("p_a + 1")) {
		t.Errorf("CDATA 不是编辑后的正文：%q", cdata)
	}
	// 红线：status 写 "u"，绝不能写 "c"。
	got, ok := opSetPointAttr(plan, name, "status")
	if !ok {
		t.Fatal("没有产出 status 属性写入")
	}
	if got != "u" {
		t.Errorf("status = %q，想 \"u\"（写 \"c\" 会被设计器下次保存静默丢弃，红线 D-1）", got)
	}
	if got == "c" {
		t.Error("status 被写成 c —— 这条是红线")
	}

	// 落盘形态：真打进 TAP 再解析，内容与 status 都在。
	tapDoc := applyAndParse(t, f, plan)
	el := tapDoc.Point(name)
	if el == nil {
		t.Fatalf("改写后的 TAP 里找不到 %q", name)
	}
	if v, _ := el.Attr("status"); v != "u" {
		t.Errorf("落盘后 status = %q，想 u", v)
	}
	if !bytes.Contains(el.Content(tapDoc.Raw), []byte("p_a + 2")) {
		t.Errorf("落盘后的 CDATA 不是编辑后正文：%q", el.Content(tapDoc.Raw))
	}
}

// TestSplitPointContentNotParseable 把正文改成解析不出来的形状（END FUNCTION→END DIALOG）：
// 自订定义点的正文过不了解析就必须拒绝写回 —— 否则设计器装载会抛 FormatException。
func TestSplitPointContentNotParseable(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	edited := replaceOnce(t, f.fenced, "END FUNCTION", "END DIALOG")
	_, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err == nil {
		t.Fatal("正文不是可解析函数块时该拒绝写回")
	}
	var ve *VerifyError
	if !errors.As(err, &ve) || ve.ExitCode() != 3 {
		t.Fatalf("该是退出码 3 的 VerifyError，得 %T %v", err, err)
	}
	if !strings.Contains(ve.Msg, "不是可解析的函数块") {
		t.Errorf("错误消息没说明原因：%q", ve.Msg)
	}
}

// ---- 授权：只读点/只读区段 ----

// TestSplitDeniesReadonlyPointEdit 只读插入点的内容被改 → DeniedError（退出码 4），
// 而不是笼统的校验失败（3）。退出码是 CLI 契约：4 表示"权限不够，别改"。
func TestSplitDeniesReadonlyPointEdit(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	edited := replaceOnce(t, f.fenced, "   # 客制註解", "   # 客制註解改")
	_, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err == nil {
		t.Fatal("改只读点该被拒")
	}
	var de *DeniedError
	if !errors.As(err, &de) || de.ExitCode() != 4 {
		t.Fatalf("该是退出码 4 的 DeniedError，得 %T %v", err, err)
	}
	if !strings.Contains(de.Error(), "不可编辑") {
		t.Errorf("拒绝原因没说是「不可编辑」：%q", de.Error())
	}
	// Detail 要带上机器码 deny=，否则用户不知道是哪种权限不够。
	if !strings.Contains(de.Error(), "deny=env-edit-env") {
		t.Errorf("Detail 该带 deny 码：%q", de.Error())
	}
}

// TestSplitDeniesLockedSectionEdit 未解锁的区段正文被改 → DeniedError，
// 且理由里要给出下一步（`tt dev tzc unlock`）—— 只说"被拒"没法行动。
func TestSplitDeniesLockedSectionEdit(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{}) // 默认 Locked
	edited := replaceOnce(t, f.fenced, "INPUT BY NAME g_x", "INPUT BY NAME g_y")
	_, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err == nil {
		t.Fatal("改未解锁区段该被拒")
	}
	var de *DeniedError
	if !errors.As(err, &de) || de.ExitCode() != 4 {
		t.Fatalf("该是 DeniedError(4)，得 %T %v", err, err)
	}
	if !strings.Contains(de.Error(), "deny=section-locked") {
		t.Errorf("Detail 该带 deny=section-locked：%q", de.Error())
	}
	if !strings.Contains(de.Error(), "unlock") {
		t.Errorf("Detail 该提示先解锁：%q", de.Error())
	}
}

// TestSplitDeniesAnchorSectionEdit 锚点区段（other_dialog）正文被改 → DeniedError，
// deny=section-anchor（红线 R4：锚点正文就是注入的点块，写了等于把展开函数写进框架）。
func TestSplitDeniesAnchorSectionEdit(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{SectionState: model.SectionUnlocked})
	anchor := regionOf(t, f.base, "adzi999.other_dialog")
	// 锚点区段自身字节只有换行，改成 "\nX\n"。
	edited := splice(f.fenced, anchor.ContentSpan.Start, anchor.ContentSpan.End, "\nX\n")
	_, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err == nil {
		t.Fatal("改锚点区段该被拒")
	}
	var de *DeniedError
	if !errors.As(err, &de) || de.ExitCode() != 4 {
		t.Fatalf("该是 DeniedError(4)，得 %T %v", err, err)
	}
	if !strings.Contains(de.Error(), "deny=section-anchor") {
		t.Errorf("Detail 该带 deny=section-anchor：%q", de.Error())
	}
}

// ---- 区段正文：折叠 + 双写 + section_flag ----

// TestSplitSectionEditFoldsTglTag 钉住写回层最要命的一条（红线 R4）：
//   - 改可编辑区段正文 → 同时产出 .tap 的 SetSectionCDATA **和** .tgl 补丁（写两处），
//     否则包自相矛盾；
//   - 区段正文里内嵌的点块必须**折回原始 TglTag 占位符原文**，绝不能把展开后的
//     函数块写进区段（下次装载会重复展开）；
//   - 改过区段 → 根 section_flag 置 "Y"（设计器 SaveADPContent 的副作用）。
func TestSplitSectionEditFoldsTglTag(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{SectionState: model.SectionUnlocked})
	const sec = "adzi999.main"
	edited := replaceOnce(t, f.fenced, "END MAIN", "END  MAIN")
	plan, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err != nil {
		t.Fatalf("改可编辑区段该成功，得 %v", err)
	}

	if len(plan.Sections) != 1 || plan.Sections[0] != sec {
		t.Errorf("Sections = %v，想 [%s]", plan.Sections, sec)
	}
	if !plan.SetSectionFlag {
		t.Error("改过区段必须置 section_flag")
	}
	if v, ok := opSetRootAttr(plan, "section_flag"); !ok || v != "Y" {
		t.Errorf("根 section_flag 写入 = %q/%v，想 Y/true", v, ok)
	}
	if len(plan.TglPatches) != 1 || plan.TglPatches[0].ID != sec {
		t.Fatalf("TglPatches = %+v，想恰好一条 %s", plan.TglPatches, sec)
	}
	body := plan.TglPatches[0].Body

	// 内嵌点必须折回占位符原文。
	placeholder := `{<point name="main.define_customerization" edit="c"/>}`
	if !bytes.Contains(body, []byte(placeholder)) {
		t.Errorf("区段正文没有折回 TglTag 占位符 %q：\n%s", placeholder, body)
	}
	// 展开后的函数正文绝不能出现在区段里（红线 R4）。
	if bytes.Contains(body, []byte("CALL ccl_init")) {
		t.Errorf("区段正文里出现了展开后的点正文（红线 R4 被破坏）：\n%s", body)
	}
	// 用户改的那处要保留。
	if !bytes.Contains(body, []byte("END  MAIN")) {
		t.Errorf("区段正文丢了用户的改动：\n%s", body)
	}
	// .tap 的区段 CDATA 与 .tgl 补丁必须逐字节一致（写两处，不能一处一个样）。
	var tapBody []byte
	for _, op := range plan.Ops {
		if s, ok := op.(tapfile.SetSectionCDATA); ok && s.ID == sec {
			tapBody = s.Content
		}
	}
	if tapBody == nil {
		t.Fatal("没有产出 SetSectionCDATA")
	}
	if !bytes.Equal(tapBody, body) {
		t.Errorf(".tap 区段 CDATA 与 .tgl 补丁不一致：\n.tap=%q\n.tgl=%q", tapBody, body)
	}
	// 区段 status 写 u。
	gotStatus := ""
	for _, op := range plan.Ops {
		if a, ok := op.(tapfile.SetSectionAttr); ok && a.ID == sec && a.Key == "status" {
			gotStatus = a.Value
		}
	}
	if gotStatus != "u" {
		t.Errorf("区段 status = %q，想 u", gotStatus)
	}
}

// ---- 删除授权 ----

// TestSplitDeletesNewPoint 删 new="Y" 的自订点（这里用的是裸名插入点）→ MarkDeleted，
// 设计器语义是 status="d" 且**保留原内容**（I5），不是从 XML 里抹掉。
func TestSplitDeletesNewPoint(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	const name = "global.import"
	plan, err := Split(f.base, f.reparse(t, deleteRegion(t, f, name)), f.pkg)
	if err != nil {
		t.Fatalf("删 new=Y 的点该成功，得 %v", err)
	}
	if len(plan.Deleted) != 1 || plan.Deleted[0] != name {
		t.Errorf("Deleted = %v，想 [%s]", plan.Deleted, name)
	}
	var mark bool
	for _, op := range plan.Ops {
		if m, ok := op.(tapfile.MarkDeleted); ok && m.Name == name {
			mark = true
		}
	}
	if !mark {
		t.Error("没有产出 MarkDeleted")
	}
	tapDoc := applyAndParse(t, f, plan)
	el := tapDoc.PointExact(name)
	if el == nil {
		t.Fatal("删点后 TAP 里该还有墓碑（保留原内容）")
	}
	if v, _ := el.Attr("status"); v != "d" {
		t.Errorf("落盘后 status = %q，想 d", v)
	}
	if !el.HasCDATA() || len(el.Content(tapDoc.Raw)) == 0 {
		t.Error("墓碑必须保留原内容（I5）")
	}
}

// TestSplitDeniesDeleteNotNew 只有 new="Y" 的自订点能删（与设计器
// CodeEditorMainWindow.CanDeleteFunction 一致）；不是自订点的一律 DeniedError。
func TestSplitDeniesDeleteNotNew(t *testing.T) {
	entries := testutil.NormalEntries("adzi999")
	tap := entries["adzi999.tap"]
	old := `name="global.import" order="" ver="" cite_std="N" new="Y"`
	if !strings.Contains(tap, old) {
		t.Fatalf("夹具里没有预期的点行：%q", old)
	}
	entries["adzi999.tap"] = strings.Replace(tap, old,
		`name="global.import" order="" ver="" cite_std="N" new=""`, 1)
	f := newSplitFixtureFromEntries(t, "adzi999", entries, synth.Options{})

	_, err := Split(f.base, f.reparse(t, deleteRegion(t, f, "global.import")), f.pkg)
	if err == nil {
		t.Fatal("删 new!=Y 的点该被拒")
	}
	var de *DeniedError
	if !errors.As(err, &de) || de.ExitCode() != 4 {
		t.Fatalf("该是 DeniedError(4)，得 %T %v", err, err)
	}
	if !strings.Contains(de.Error(), "new=") {
		t.Errorf("Detail 该带 new= 实际值：%q", de.Error())
	}
}

// TestSplitDeniesDeleteReadonly 点虽然 new="Y" 但不可编辑（readonly="Y"）→ 删除仍被拒。
// new=Y 只是"可以删"的第一道门，授权是第二道。
func TestSplitDeniesDeleteReadonly(t *testing.T) {
	entries := testutil.NormalEntries("adzi999")
	tap := entries["adzi999.tap"]
	// 只改 global.import 的那一处：用它的名字定位后半段，避免误伤其它点。
	idx := strings.Index(tap, `name="global.import"`)
	if idx < 0 {
		t.Fatal("夹具里没有 global.import")
	}
	head := tap[:idx]
	rest := tap[idx:]
	rest = strings.Replace(rest, `src="s" readonly=""`, `src="s" readonly="Y"`, 1)
	entries["adzi999.tap"] = head + rest
	f := newSplitFixtureFromEntries(t, "adzi999", entries, synth.Options{})

	_, err := Split(f.base, f.reparse(t, deleteRegion(t, f, "global.import")), f.pkg)
	if err == nil {
		t.Fatal("删不可编辑的点该被拒")
	}
	var de *DeniedError
	if !errors.As(err, &de) {
		t.Fatalf("该是 DeniedError，得 %T %v", err, err)
	}
	if !strings.Contains(de.Error(), "deny=") {
		t.Errorf("Detail 该带 deny 码：%q", de.Error())
	}
}

// ---- 新增点（APPEND 锚点内）----

// addFuncBlock 是往其他锚点区段里追加一个自订函数点的围栏块。
func addFuncBlock(name, fn string) string {
	return `{//@tdev:begin point ` + name + ` [EDITABLE new="Y" status="u"]}` + "\n" +
		"#+ Description: 新增測試函式\n" +
		"PRIVATE FUNCTION " + fn + `(p_b)` + "\n" +
		"   DEFINE p_b INTEGER\n" +
		"   RETURN p_b\n" +
		"END FUNCTION\n" +
		"{//@tdev:end point}\n"
}

// TestSplitAddsAppendedPoint order 必须 = max(现存自订点 order)+1（ProgramInformation.Add），
// status 写 "u"（不是 "c"），new="Y"；落盘后能解析出这个新点。
func TestSplitAddsAppendedPoint(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	anchor := regionOf(t, f.base, "adzi999.other_function")
	const name = "function.adzi999_add"
	edited := splice(f.fenced, anchor.EndFence.Start, anchor.EndFence.Start, addFuncBlock(name, "adzi999_add"))
	plan, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err != nil {
		t.Fatalf("在 APPEND 锚点里新增点该成功，得 %v", err)
	}
	if len(plan.Added) != 1 || plan.Added[0] != name {
		t.Errorf("Added = %v，想 [%s]", plan.Added, name)
	}
	add, ok := findAddPoint(plan, name)
	if !ok {
		t.Fatal("没有产出 AddPoint")
	}
	// 现存自订点只有 function.adzi999_calc order="1" → max+1 = 2。
	if got := add.Attrs["order"]; got != "2" {
		t.Errorf("order = %q，想 2（max(现存 order)+1）", got)
	}
	if got := add.Attrs["status"]; got != "u" {
		t.Errorf("新点 status = %q，想 u（写 c 会被设计器丢弃，红线 D-1）", got)
	}
	if got := add.Attrs["new"]; got != "Y" {
		t.Errorf("新点 new = %q，想 Y", got)
	}
	if !bytes.Contains(add.Content, []byte("FUNCTION adzi999_add")) {
		t.Errorf("新点正文不对：%q", add.Content)
	}
	// 落盘形态：新点真的在 TAP 里，status=u。
	tapDoc := applyAndParse(t, f, plan)
	el := tapDoc.Point(name)
	if el == nil {
		t.Fatalf("落盘后 TAP 里没有 %q", name)
	}
	if v, _ := el.Attr("status"); v != "u" {
		t.Errorf("落盘后 status = %q，想 u", v)
	}
	if v, _ := el.Attr("order"); v != "2" {
		t.Errorf("落盘后 order = %q，想 2", v)
	}
}

// TestSplitAppendedNonParseableRejected 追加的正文过不了解析 → 拒绝写回（VerifyError）。
func TestSplitAppendedNonParseableRejected(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	anchor := regionOf(t, f.base, "adzi999.other_function")
	block := `{//@tdev:begin point function.adzi999_bad [EDITABLE new="Y" status="u"]}` + "\n" +
		"NOT A FUNCTION AT ALL\n" +
		"{//@tdev:end point}\n"
	edited := splice(f.fenced, anchor.EndFence.Start, anchor.EndFence.Start, block)
	_, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err == nil {
		t.Fatal("新增点正文不是可解析函数块时该拒绝写回")
	}
	var ve *VerifyError
	if !errors.As(err, &ve) || ve.ExitCode() != 3 {
		t.Fatalf("该是 VerifyError(3)，得 %T %v", err, err)
	}
	if !strings.Contains(ve.Msg, "不是可解析的函数块") {
		t.Errorf("错误消息没说明原因：%q", ve.Msg)
	}
}

// TestSplitAddsPointForPlaceholderOnlyTGL TGL 有占位符而 TAP 没有点时，设计器会造空点；
// 用户把内容填上后写回 = 在 TAP 里新增一个 <point>（不是 SetPointCDATA）。
func TestSplitAddsPointForPlaceholderOnlyTGL(t *testing.T) {
	entries := testutil.NormalEntries("adzi999")
	tap := entries["adzi999.tap"]
	if err := checkRemovable(tap, `name="global.import" order=""`); err != nil {
		t.Fatal(err)
	}
	entries["adzi999.tap"] = removeTAPPoint(t, tap, `name="global.import" order=""`)
	f := newSplitFixtureFromEntries(t, "adzi999", entries, synth.Options{})
	const name = "global.import"
	base := regionOf(t, f.base, name)
	if len(base.TglTag) == 0 || base.ContentSpan.Len() != 0 {
		t.Fatalf("夹具失效：期望 TGL 有占位符、TAP 无内容（TglTag=%q contentLen=%d）",
			base.TglTag, base.ContentSpan.Len())
	}
	edited := splice(f.fenced, base.ContentSpan.Start, base.ContentSpan.End, "IMPORT zzz\n")
	plan, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err != nil {
		t.Fatalf("填空点该成功，得 %v", err)
	}
	if len(plan.Added) != 1 || plan.Added[0] != name {
		t.Errorf("Added = %v，想 [%s]", plan.Added, name)
	}
	add, ok := findAddPoint(plan, name)
	if !ok {
		t.Fatal("没有产出 AddPoint")
	}
	if !bytes.Contains(add.Content, []byte("IMPORT zzz")) {
		t.Errorf("新增点正文不对：%q", add.Content)
	}
}

func checkRemovable(tap, marker string) error {
	if strings.Count(tap, marker) != 1 {
		return errors.New("夹具里标记不唯一：" + marker)
	}
	return nil
}

// removeTAPPoint 从 TAP 文本里删掉一个 <point ...>…</point> 元素（含其后换行）。
func removeTAPPoint(t *testing.T, tap, marker string) string {
	t.Helper()
	i := strings.Index(tap, "<point "+marker)
	if i < 0 {
		t.Fatalf("TAP 里找不到 %q", marker)
	}
	j := strings.Index(tap[i:], "</point>")
	if j < 0 {
		t.Fatalf("%q 没有闭合标签", marker)
	}
	end := i + j + len("</point>")
	// 吃掉元素后的第一个行尾（CRLF 或 LF）。
	switch {
	case strings.HasPrefix(tap[end:], "\r\n"):
		end += 2
	case strings.HasPrefix(tap[end:], "\n"):
		end++
	}
	return tap[:i] + tap[end:]
}

// ---- 改名事务 ----

// TestSplitRenameProducesTombstonePair 改名不是"改个属性"而是一次事务：
// 旧名留 status="d" 墓碑（保留原正文逐字节），新名 status="u" 且签名行已换名。
// 两者成对出现，设计器下次装载才认得。
func TestSplitRenameProducesTombstonePair(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	const oldName = "function.adzi999_calc"
	const newName = "function.adzi999_mul"
	edited := replaceOnce(t, f.fenced, `fn="adzi999_calc(p_a)"`, `fn="adzi999_mul(p_a)"`)
	plan, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err != nil {
		t.Fatalf("改名该成功，得 %v", err)
	}
	wantRename := oldName + " → " + newName
	if len(plan.Renamed) != 1 || plan.Renamed[0] != wantRename {
		t.Errorf("Renamed = %v，想 [%q]", plan.Renamed, wantRename)
	}

	// 操作序列：清旧墓碑 → 活点改名 → 整块重写（签名行换新）→ status=u → 补墓碑。
	var sawRemove, sawRename, sawCDATA bool
	for _, op := range plan.Ops {
		switch o := op.(type) {
		case tapfile.RemoveTombstones:
			sawRemove = o.Name == oldName
		case tapfile.RenamePoint:
			sawRename = o.From == oldName && o.To == newName
		case tapfile.SetPointCDATA:
			sawCDATA = o.Name == newName && bytes.Contains(o.Content, []byte("FUNCTION adzi999_mul"))
		}
	}
	if !sawRemove || !sawRename || !sawCDATA {
		t.Errorf("改名事务缺步：remove=%v rename=%v cdata=%v", sawRemove, sawRename, sawCDATA)
	}
	if v, ok := opSetPointAttr(plan, newName, "status"); !ok || v != "u" {
		t.Errorf("活点 status = %q/%v，想 u", v, ok)
	}
	tomb, ok := findAddPoint(plan, oldName)
	if !ok {
		t.Fatal("没有产出旧名墓碑")
	}
	if tomb.Attrs["status"] != "d" {
		t.Errorf("墓碑 status = %q，想 d", tomb.Attrs["status"])
	}
	if tomb.Attrs["new"] != "Y" {
		t.Errorf("墓碑 new = %q，想 Y", tomb.Attrs["new"])
	}
	// 墓碑正文必须是**基线原字节**（逐字节，不是折叠后/展开后的）。
	basePoint := regionOf(t, f.base, oldName)
	wantOld := f.base.Text[basePoint.ContentSpan.Start:basePoint.ContentSpan.End]
	if !bytes.Equal(tomb.Content, wantOld) {
		t.Errorf("墓碑正文不是基线原字节：\n得 %q\n想 %q", tomb.Content, wantOld)
	}

	// 落盘：一对点（旧 d + 新 u），新点签名行已换名。
	tapDoc := applyAndParse(t, f, plan)
	tombEl := tapDoc.PointExact(oldName)
	if tombEl == nil {
		t.Fatal("落盘后找不到旧名墓碑")
	}
	if v, ok := tombEl.Attr("status"); !ok || v != "d" {
		t.Errorf("墓碑落盘 status = %q", v)
	}
	ne := tapDoc.Point(newName)
	if ne == nil {
		t.Fatal("落盘后找不到新点")
	}
	if !bytes.Contains(ne.Content(tapDoc.Raw), []byte("FUNCTION adzi999_mul")) {
		t.Errorf("新点正文签名没换：%q", ne.Content(tapDoc.Raw))
	}
}

// TestSplitRenameEmptyTarget 围栏 fn 被清空 → VerifyError（改名目标不能为空）。
func TestSplitRenameEmptyTarget(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	edited := replaceOnce(t, f.fenced, `fn="adzi999_calc(p_a)"`, `fn=""`)
	_, err := Split(f.base, f.reparse(t, edited), f.pkg)
	if err == nil {
		t.Fatal("改名目标为空该被拒")
	}
	var ve *VerifyError
	if !errors.As(err, &ve) || ve.ExitCode() != 3 {
		t.Fatalf("该是 VerifyError(3)，得 %T %v", err, err)
	}
	if !strings.Contains(ve.Msg, "改名目标为空") {
		t.Errorf("错误消息不对：%q", ve.Msg)
	}
}

// ---- 手工 ParseResult：真实围栏解析层拦住的路径，直接构造输入 ----

// TestSplitDeleteSectionDenied 区段不可删除（fence.Parse 已经先拦，这里钉 Split 的最后一道）。
func TestSplitDeleteSectionDenied(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	parsed := &fence.ParseResult{
		Doc:     f.base,
		Deleted: []*model.Region{{Name: "adzi999.main", Kind: model.RegionSection}},
	}
	_, err := Split(f.base, parsed, f.pkg)
	if err == nil {
		t.Fatal("删区段该被拒")
	}
	var ve *VerifyError
	if !errors.As(err, &ve) || ve.ExitCode() != 3 {
		t.Fatalf("该是 VerifyError(3)，得 %T %v", err, err)
	}
	if !strings.Contains(ve.Msg, "区段不可删除") {
		t.Errorf("错误消息不对：%q", ve.Msg)
	}
}

// TestSplitDeleteMissingInTAP 要删的点在 TAP 里不存在 → VerifyError（没有内容可标记删除）。
func TestSplitDeleteMissingInTAP(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	parsed := &fence.ParseResult{
		Doc: f.base,
		Deleted: []*model.Region{{
			Name: "global.nope", Kind: model.RegionPoint, Editable: true,
			Meta: map[string]string{"new": "Y"},
		}},
	}
	_, err := Split(f.base, parsed, f.pkg)
	if err == nil {
		t.Fatal("删 TAP 里不存在的点该被拒")
	}
	var ve *VerifyError
	if !errors.As(err, &ve) {
		t.Fatalf("该是 VerifyError，得 %T %v", err, err)
	}
	if !strings.Contains(ve.Msg, "不存在") {
		t.Errorf("错误消息不对：%q", ve.Msg)
	}
}

// TestSplitTAPParseFailure TAP 本身解析不了 → VerifyError（退出码 3），并带底层原因。
func TestSplitTAPParseFailure(t *testing.T) {
	f := newSplitFixture(t, "adzi999", synth.Options{})
	bad := &pkgfile.Package{
		Path:    f.pkg.Path,
		Entries: []*pkgfile.Entry{{Name: "adzi999.tap", Data: []byte("<add_points")}},
	}
	_, err := Split(f.base, &fence.ParseResult{Doc: f.base}, bad)
	if err == nil {
		t.Fatal("TAP 解析失败该报错")
	}
	var ve *VerifyError
	if !errors.As(err, &ve) || ve.ExitCode() != 3 {
		t.Fatalf("该是 VerifyError(3)，得 %T %v", err, err)
	}
	if len(ve.Detail) == 0 {
		t.Error("该带上底层解析错误")
	}
}

// ---- newPointAttrs：属性集合与顺序（设计器 ctor 口径）----

// TestNewPointAttrs 钉住新增 <point> 的属性集：
//   - status 恒 "u"（红线 D-1，绝不写 "c"）；
//   - new="Y"、cite_std="N"、ind_extra="N"；
//   - src 取环境 env；topstd 模式写 "s"；
//   - 只有自订定义点有 order，裸名插入点 order 为空（设计器 SortIndex setter 拒绝赋值）。
func TestNewPointAttrs(t *testing.T) {
	env := model.EnvContext{Env: "c", Topind: "sd"}
	attrs := newPointAttrs("function.x", 7, env, map[string]string{"ignored": "1"})
	want := map[string]string{
		"name":      "function.x",
		"order":     "7",
		"ver":       "",
		"cite_std":  "N",
		"new":       "Y",
		"src":       "c",
		"status":    "u",
		"ch":        "",
		"ind_fun":   "sd",
		"ind_extra": "N",
	}
	if len(attrs) != len(want) {
		t.Errorf("属性集 = %v（%d 项），想 %d 项", attrs, len(attrs), len(want))
	}
	for k, v := range want {
		if got, ok := attrs[k]; !ok || got != v {
			t.Errorf("属性 %q = %q/%v，想 %q", k, got, ok, v)
		}
	}
	if attrs["status"] == "c" {
		t.Error("status 被写成 c —— 红线 D-1")
	}

	// 裸名插入点：order 必须为空串（不是 "0"、不是缺失）。
	bare := newPointAttrs("global.memo", 9, env, nil)
	if got, ok := bare["order"]; !ok || got != "" {
		t.Errorf("裸名点 order = %q/%v，想存在且为空串", got, ok)
	}

	// topstd 模式：src 写 "s"（标准来源语义），环境 env 不再直出。
	top := newPointAttrs("function.y", 1, model.EnvContext{Env: "c", Topind: "sd", IsTopstdMode: true}, nil)
	if got := top["src"]; got != "s" {
		t.Errorf("topstd 模式 src = %q，想 s", got)
	}

	// 设计器 ctor 的**属性顺序**（AddPointModel.cs:64-75）——map 无序，但集合相同即可；
	// 这里额外钉住"不多不少"：出现陌生属性说明有人往这里塞了非设计器字段。
	allowed := map[string]bool{"name": true, "order": true, "ver": true, "cite_std": true,
		"new": true, "src": true, "status": true, "ch": true, "ind_fun": true, "ind_extra": true}
	for k := range attrs {
		if !allowed[k] {
			t.Errorf("出现设计器 ctor 里没有的属性 %q", k)
		}
	}
}

// ---- foldSection：TglTag 折叠（红线 R4）----

// foldDoc 造一个可用于 foldSection 的手工文档。
//
// 编辑后文本：text；区段 ContentSpan 覆盖整段；子点 FullSpan 是块的位置。
// 基线区段只提供 TglTag（占位符原文）。
func foldDoc(text string, content ByteRange, childSpans []ByteRange) *model.Document {
	sec := &model.Region{Name: "p.s", Kind: model.RegionSection, ContentSpan: content}
	for _, sp := range childSpans {
		sec.Children = append(sec.Children, &model.Region{Kind: model.RegionPoint, FullSpan: sp})
	}
	return &model.Document{Text: []byte(text), Regions: []*model.Region{sec}}
}

// TestFoldSectionRestoresTglTag 把编辑后的区段正文折回占位符版：
// 子点位置换成基线 TglTag 原文，其余字节逐字节保留（含区间外的前缀/后缀不参与）。
func TestFoldSectionRestoresTglTag(t *testing.T) {
	const text = "PRE|AB|MID|CD|POST"
	doc := foldDoc(text, ByteRange{Start: 0, End: len(text)}, []ByteRange{{Start: 4, End: 6}, {Start: 11, End: 13}})
	baseSec := &model.Region{Name: "p.s", Children: []*model.Region{
		{Name: "function.a", Kind: model.RegionPoint, TglTag: []byte(`{<point name="function.a"/>}`)},
		{Name: "function.b", Kind: model.RegionPoint, TglTag: []byte(`{<point name="function.b"/>}`)},
	}}
	editedSec := doc.Regions[0]
	got, err := foldSection(doc, baseSec, editedSec)
	if err != nil {
		t.Fatalf("折叠不该失败：%v", err)
	}
	want := "PRE|" + `{<point name="function.a"/>}` + "|MID|" + `{<point name="function.b"/>}` + "|POST"
	if string(got) != want {
		t.Errorf("折叠结果 = %q\n想 %q", got, want)
	}

	// 没有子点的区段：原样返回正文。
	doc2 := foldDoc("SIMPLE", ByteRange{Start: 0, End: 6}, nil)
	got2, err := foldSection(doc2, &model.Region{Name: "p.s"}, doc2.Regions[0])
	if err != nil {
		t.Fatalf("无子点区段不该失败：%v", err)
	}
	if string(got2) != "SIMPLE" {
		t.Errorf("无子点区段该原样返回，得 %q", got2)
	}
}

// TestFoldSectionWithoutTglTagIsError 找不到基线 TglTag 的点块一律报错 ——
// 绝不把展开后的函数写进区段（红线 R4）。这是写回层最不能妥协的一条。
func TestFoldSectionWithoutTglTagIsError(t *testing.T) {
	doc := foldDoc("A|BLOCK|B", ByteRange{Start: 0, End: 9}, []ByteRange{{Start: 2, End: 7}})
	doc.Regions[0].Children[0].Name = "function.a" // 报错信息要能说出是哪个点
	baseSec := &model.Region{Name: "p.s", Children: []*model.Region{
		{Name: "function.a", Kind: model.RegionPoint}, // TglTag 为 nil
	}}
	_, err := foldSection(doc, baseSec, doc.Regions[0])
	if err == nil {
		t.Fatal("没有原始 TglTag 时必须报错")
	}
	var ve *VerifyError
	if !errors.As(err, &ve) || ve.ExitCode() != 3 {
		t.Fatalf("该是 VerifyError(3)，得 %T %v", err, err)
	}
	if !strings.Contains(ve.Msg, "R4") && !strings.Contains(ve.Msg, "TglTag") {
		t.Errorf("错误消息该点明 TglTag/红线 R4：%q", ve.Msg)
	}
	// Detail 要能指出是哪个区段、哪个点。
	joined := strings.Join(ve.Detail, ";")
	if !strings.Contains(joined, "p.s") || !strings.Contains(joined, "function.a") {
		t.Errorf("Detail 该带区段与点名：%v", ve.Detail)
	}
}

// TestFoldSectionFewerBaseChildren 编辑后多出子点（基线 TglTag 不够用）→ 报错，
// 而不是拿空占位符糊过去。
func TestFoldSectionFewerBaseChildren(t *testing.T) {
	doc := foldDoc("|A||B|", ByteRange{Start: 0, End: 7}, []ByteRange{{Start: 1, End: 2}, {Start: 5, End: 6}})
	baseSec := &model.Region{Name: "p.s", Children: []*model.Region{
		{Name: "function.a", Kind: model.RegionPoint, TglTag: []byte("{A}")},
	}}
	_, err := foldSection(doc, baseSec, doc.Regions[0])
	if err == nil {
		t.Fatal("基线 TglTag 不够用时必须报错")
	}
}

// ---- renameBlock 的防御分支 ----

// TestRenameBlockDefensive 改名拼块的防御分支：签名行缺失、签名区间相对正文越界。
//
// SignatureSpan 的**绝对**区间越界曾会先 `doc.Text[Start:End]` 再检查、直接 panic
// （生产链里由 fence.Parse 重算所以不可达，但 hand-made Region / 坏 regions.json 会踩到）。
// 已修为返回 VerifyError(3)，越界回归见 rename_bounds_test.go。
func TestRenameBlockDefensive(t *testing.T) {
	doc := &model.Document{Text: []byte("PRIVATE FUNCTION a()\n  LET x = 1\nEND FUNCTION\n")}
	edited := &model.Region{
		Name: "function.a", Kind: model.RegionPoint,
		ContentSpan: model.ByteRange{Start: 0, End: len(doc.Text)},
	}
	// 没有签名行 → 无法改名。
	if _, err := renameBlock(doc, edited, edited, "b", fence.Pair{}); err == nil {
		t.Error("签名行缺失该报错")
	}
	// 签名区间落在正文**之前**（相对起点为负）→ 内部错误，不许拼出半截签名。
	edited.SignatureSpan = &model.ByteRange{Start: 0, End: 5}
	edited.ContentSpan = model.ByteRange{Start: 10, End: len(doc.Text)}
	if _, err := renameBlock(doc, edited, edited, "b", fence.Pair{}); err == nil {
		t.Error("签名区间在正文之前该报错")
	}
	// 签名区间越过正文**末尾** → 内部错误。
	edited.SignatureSpan = &model.ByteRange{Start: len(doc.Text) - 3, End: len(doc.Text)}
	edited.ContentSpan = model.ByteRange{Start: 0, End: len(doc.Text) - 5}
	if _, err := renameBlock(doc, edited, edited, "b", fence.Pair{}); err == nil {
		t.Error("签名区间越过正文末尾该报错")
	}
	// 正常路径：签名行换成新名，正文其余字节逐字节保留。
	edited.ContentSpan = model.ByteRange{Start: 0, End: len(doc.Text)}
	edited.SignatureSpan = &model.ByteRange{Start: 0, End: len("PRIVATE FUNCTION a()")}
	got, err := renameBlock(doc, edited, edited, "b", fence.Pair{})
	if err != nil {
		t.Fatalf("正常改名不该失败：%v", err)
	}
	if !bytes.Contains(got, []byte("PRIVATE FUNCTION b")) {
		t.Errorf("新签名不对：%q", got)
	}
	if !bytes.Contains(got, []byte("  LET x = 1")) || !bytes.Contains(got, []byte("END FUNCTION")) {
		t.Errorf("正文其余部分被弄丢了：%q", got)
	}
}

// ByteRange 是 model 的别名（本文件只用来构造手工区间，避免 import 噪音）。
type ByteRange = model.ByteRange
