package fep

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// 指纹 v3 的守护测试。判决口径 = 契约 **A21（非对称覆盖，取代 A19 的双向 Jaccard）**：
//
//	通过 ⟺ ① |live| ≥ max(10, 0.5×|rec|)
//	      且 ② live→rec 身份覆盖 ≥ 0.98
//	      且 ③ live→rec 边覆盖   ≥ 0.98
//	      且 ④ live 多余身份 ≤ 2%
//	      且 ⑤ 类型直方图差异 ≤ 2%（按 live 的身份集合收窄 rec 后比）
//
// 三件套：身份多重集（colName#fieldId）、引用边集合（父身份→子身份）、类型直方图；
// `idDigest` **只存不判**。差异清单**分两栏**：意外（live 有 rec 无 ⇒ 拒的理由）／
// 尚未加载（rec 有 live 无 ⇒ 只作信息）。
//
// 全部用**合成 Model**（不连真机、不起会话）：节点形态照真语料
// （`{TableColumn 3631 {{name …} {colName "ooba002_1"} {fieldId "1"}}}`）。

// v3Cols 是一张表单里的 12 个字段身份（够过 ① 的规模下限：15 条 ≥ max(10, 7.5)）。
var v3Cols = [][2]interface{}{
	{"b_xmda001", 18}, {"b_index", 25}, {"b_name", 26}, {"b_xmda004", 23},
	{"b_unit", 9}, {"b_qty", 12}, {"b_price", 13}, {"b_amt", 14},
	{"b_date", 15}, {"b_supp", 16}, {"b_rem", 17}, {"b_flag", 19},
}

// identField 造一个**带表单身份**的节点（colName + fieldId 成对）。
func identField(typ string, id int, col string, fid int) string {
	return node(fmt.Sprintf("%s %d", typ, id), [][2]string{
		{"name", "demo_t." + col}, {"colName", col}, {"fieldId", strconv.Itoa(fid)},
	})
}

// v3FieldList 造前 n 个字段的节点串（编号全部 +shift）。
func v3FieldList(shift, n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		col, fid := v3Cols[i][0].(string), v3Cols[i][1].(int)
		out = append(out, identField("FormField", 400+shift+i*2, col, fid))
	}
	return out
}

// v3Form 造"一个窗口 + 一张表单 + 前 n 个字段"的界面；`shift` 把编号整体平移
// （= 同一界面、换了一条启动路径：号码口径在这里判不过，身份口径必须判过）。
func v3Form(shift, n int) *Model {
	m := NewModel()
	m.ApplyMessage(windowMsg(100+shift, "w_axmt500", "料件主档作业(w_axmt500)", "demo_t",
		v3FieldList(shift, n)...))
	if m.Len() == 0 {
		panic("v3 合成夹具没进模型")
	}
	return m
}

// v3FormExtra 造"12 个公共字段 + 1 个录制件里没有的字段"（④ 意外身份那一侧）。
func v3FormExtra(shift int) *Model {
	fields := append(v3FieldList(shift, len(v3Cols)),
		identField("FormField", 900+shift, "b_extra", 99))
	m := NewModel()
	m.ApplyMessage(windowMsg(100+shift, "w_axmt500", "料件主档作业(w_axmt500)", "demo_t", fields...))
	return m
}

// v3Grouped 造"身份相同、但字段挂在不同的分组下"的界面（③ 边覆盖那一侧）：
// 前 first 个字段在左分组，其余在右分组。rec 用 first=8，live 用 first=6。
func v3Grouped(first int) *Model {
	left := v3FieldList(0, first)
	var right []string
	for i := first; i < len(v3Cols); i++ {
		col, fid := v3Cols[i][0].(string), v3Cols[i][1].(int)
		right = append(right, identField("FormField", 400+i*2, col, fid))
	}
	m := NewModel()
	m.ApplyMessage(node("an 0 Window 100", [][2]string{{"name", "w_g"}, {"text", "分组作业"}},
		node("Form 200", [][2]string{{"name", "demo_t"}},
			node("HBox 300", nil, left...),
			node("HBox 320", nil, right...))))
	return m
}

func fpOf(m *Model) Fingerprint { return FingerprintOf(EntriesOf(m)) }
func fpSnap(label string, m *Model) Snapshot {
	es := EntriesOf(m)
	return Snapshot{Label: label, Entries: es, FP: FingerprintOf(es)}
}

// fpSnapEntries 是"直接给条目"的快照（不经 Model）。
func fpSnapEntries(label string, es []FPEntry) Snapshot {
	return Snapshot{Label: label, Entries: es, FP: FingerprintOf(es)}
}

// ---------------------------------------------------------------------------
// 1) 号码全变、身份不变 ⇒ 通过（这是本切片的全部意义）。
// ---------------------------------------------------------------------------
func TestV3SameInterfaceDespiteAllIDsChanged(t *testing.T) {
	rec, live := EntriesOf(v3Form(0, 12)), EntriesOf(v3Form(1000, 12))

	// 前提自检：编号**一条都不重合**（否则这条测试证明不了"号码全变也能过"）。
	liveIDs := map[int]bool{}
	for _, e := range live {
		liveIDs[e.ID] = true
	}
	for _, e := range rec {
		if liveIDs[e.ID] {
			t.Fatalf("夹具不成立：编号 #%d 两边都出现（这条测试要的是号码全变）", e.ID)
		}
	}
	// 前提自检 2：idDigest（旧哈希）确实不同 —— 旧判据在这里是"判不过"的。
	if fpOf(v3Form(0, 12)).Hash == fpOf(v3Form(1000, 12)).Hash {
		t.Fatal("夹具不成立：两份的 idDigest 碰巧相同")
	}

	d, same := JudgeSameInterface(rec, live)
	if !same {
		t.Fatalf("同一界面、号码全变 ⇒ v3 必须判同一界面：%+v（意外 %v / 尚未加载 %v / 类型变 %v）",
			d, d.IdentUnexpected, d.IdentPending, d.TypeChanged)
	}
	if d.LiveCoverage != 1 || d.EdgeCoverage != 1 || d.ExtraShare != 0 || d.TypeDiff != 0 || !d.EvidenceOK {
		t.Fatalf("身份集合完全相同的一份该是 1/1/0/0：%+v", d)
	}

	v := gate(fpSnap("件", v3Form(0, 12)), fpSnap("会话", v3Form(1000, 12)))
	if !v.OK {
		t.Fatalf("同一界面、号码全变 ⇒ 闸门必须放行（这就是 v3 要消灭的误拒）：%s", v.Report())
	}
}

// 1b) 全同 ⇒ 通过；两个方向都通过（A21 单向覆盖在全等时不产生差异）。
func TestV3SameInterfacePassesBothWays(t *testing.T) {
	a, b := v3Form(0, 12), v3Form(1000, 12)
	if d, ok := JudgeSameInterface(EntriesOf(a), EntriesOf(a)); !ok {
		t.Fatalf("自己跟自己必须过：%+v", d)
	}
	if d, ok := JudgeSameInterface(EntriesOf(a), EntriesOf(b)); !ok {
		t.Fatalf("rec→live 方向必须过：%+v", d)
	}
	if d, ok := JudgeSameInterface(EntriesOf(b), EntriesOf(a)); !ok {
		t.Fatalf("live→rec 方向必须过：%+v", d)
	}
	// 三件套全等（哈希口径也认同一界面）
	if !SameInterface(fpOf(a), fpOf(b)) {
		t.Fatal("身份/边/类型全同，SameInterface 必须判同一界面")
	}
}

// ---------------------------------------------------------------------------
//  2. 半屏：live ⊂ rec（录制件是整会话、当前只开半屏）⇒ **通过**，
//     缺的那些列在「尚未加载」栏，**不构成拒绝**。这是 A21 存在的理由。
//
// ---------------------------------------------------------------------------
func TestV3AcceptsHalfScreenWhenRecordingIsSuperset(t *testing.T) {
	rec := EntriesOf(v3Form(0, 12))   // 15 条 / 12 身份（整会话）
	live := EntriesOf(v3Form(900, 9)) // 12 条 / 9 身份（半屏，≥ max(10, 7.5)）

	d, ok := JudgeSameInterface(rec, live)
	if !ok {
		t.Fatalf("半屏（live ⊂ rec）必须通过：%+v", d)
	}
	if !d.EvidenceOK || d.LiveCoverage != 1 || d.EdgeCoverage != 1 || d.ExtraShare != 0 {
		t.Fatalf("半屏的覆盖应当是 100%%/0 意外：%+v", d)
	}
	if len(d.IdentPending) == 0 || len(d.EdgePending) == 0 {
		t.Fatalf("半屏要在「尚未加载」栏列出缺的身份与边：%+v", d)
	}
	if len(d.IdentUnexpected) != 0 || len(d.EdgeUnexpected) != 0 {
		t.Fatalf("半屏不该有「意外」（那些都在录制件里）：%+v", d)
	}

	v := gate(fpSnap("件", v3Form(0, 12)), fpSnap("会话", v3Form(900, 9)))
	if !v.OK {
		t.Fatalf("半屏必须放行：%s", v.Report())
	}
	rep := v.Report()
	// 两栏都要看得见：尚未加载点名具体身份，但判决是通过的。
	for _, want := range []string{"［尚未加载］身份", "［意外］身份", "b_flag#19"} {
		if !strings.Contains(rep, want) {
			t.Fatalf("半屏的报文要能读懂（缺 %q）：\n%s", want, rep)
		}
	}
}

// ---------------------------------------------------------------------------
// 3) 证据不足：live 太小 ⇒ 拒，理由是"证据不足"，哪怕覆盖是 100%。
// ---------------------------------------------------------------------------
func TestV3RefusesWhenEvidenceTooSmall(t *testing.T) {
	rec := EntriesOf(v3Form(0, 12))   // 15 条
	live := EntriesOf(v3Form(900, 2)) // 5 条 < max(10, 7.5)

	d, ok := JudgeSameInterface(rec, live)
	if ok {
		t.Fatalf("证据量不足必须拒（否则 5 条也能说 100%% 覆盖）：%+v", d)
	}
	if d.EvidenceOK {
		t.Fatalf("① 应当判「证据不足」：%+v", d)
	}
	if d.LiveCoverage < FPIdentJaccardMin || d.EdgeCoverage < FPEdgeIsoMin {
		t.Fatalf("夹具不成立：覆盖本来就够，这条测试只让 ① 起作用：%+v", d)
	}
	v := gate(fpSnap("件", v3Form(0, 12)), fpSnap("会话", v3Form(900, 2)))
	if v.OK || v.Code != GateFPNoMatch {
		t.Fatalf("证据不足必须拒：%+v", v)
	}
	if !strings.Contains(v.Reason, "证据") {
		t.Fatalf("拒绝理由要写明「证据不足」：%s", v.Reason)
	}
}

// ---------------------------------------------------------------------------
// 4) 意外身份：live 多出录制件里没有的身份 > 2% ⇒ 拒，且列在**「意外」**栏。
// ---------------------------------------------------------------------------
func TestV3RefusesWhenLiveHasExtraIdentities(t *testing.T) {
	rec, live := EntriesOf(v3Form(0, 12)), EntriesOf(v3FormExtra(900))

	d, ok := JudgeSameInterface(rec, live)
	if ok {
		t.Fatalf("当前界面多出录制件没有的身份，必须拒：%+v", d)
	}
	if d.ExtraShare <= FPTypeHistMaxDiff {
		t.Fatalf("④ 多余身份占比要如实超线：%v", d.ExtraShare)
	}
	if !containsStr(d.IdentUnexpected, "b_extra#99") {
		t.Fatalf("多出来的身份要列在「意外」栏：%v", d.IdentUnexpected)
	}
	if containsStr(d.IdentPending, "b_extra#99") {
		t.Fatalf("「意外」的东西不许混进「尚未加载」栏：%v", d.IdentPending)
	}
	v := gate(fpSnap("件", v3Form(0, 12)), fpSnap("会话", v3FormExtra(900)))
	if v.OK {
		t.Fatalf("多出身份必须拒：%s", v.Report())
	}
	if !strings.Contains(v.Report(), "［意外］身份") || !strings.Contains(v.Report(), "b_extra#99") {
		t.Fatalf("拒绝报文要在「意外」栏点名它：\n%s", v.Report())
	}
}

// ---------------------------------------------------------------------------
//  5. 边覆盖：身份集合完全相同、但当前界面的父子关系在录制件里找不到 ⇒ 拒。
//     （证明 ③ 不是摆设：没有它，"字段挂错组"会被身份集合放过）
//
// ---------------------------------------------------------------------------
func TestV3RefusesWhenLiveEdgeNotInRecording(t *testing.T) {
	rec := EntriesOf(v3Grouped(8))  // 左 8 右 4
	live := EntriesOf(v3Grouped(6)) // 左 6 右 6（身份相同、分组不同）

	d, ok := JudgeSameInterface(rec, live)
	if ok {
		t.Fatalf("父子关系不同必须拒：%+v", d)
	}
	if d.LiveCoverage != 1 || d.ExtraShare != 0 || d.TypeDiff != 0 {
		t.Fatalf("夹具不成立：这条测试只让 ③ 不同（身份 %v / 类型差 %v）", d.LiveCoverage, d.TypeDiff)
	}
	if d.EdgeCoverage >= FPEdgeIsoMin {
		t.Fatalf("③ 边覆盖要如实低于门槛 %.2f：%v", FPEdgeIsoMin, d.EdgeCoverage)
	}
	if len(d.EdgeUnexpected) == 0 {
		t.Fatalf("当前有、录制件没有的边要列在「意外」栏：%+v", d)
	}
	if !strings.Contains(strings.Join(d.EdgeUnexpected, " "), "b_price#13") {
		t.Fatalf("「意外」栏要点名搬了家的那条边（挂到右边的 b_price#13）：%v", d.EdgeUnexpected)
	}
	if v := gate(fpSnap("件", v3Grouped(8)), fpSnap("会话", v3Grouped(6))); v.OK {
		t.Fatalf("父子关系不同必须拒：%s", v.Report())
	}
}

// ---------------------------------------------------------------------------
// 6) 类型直方图起作用：身份同、边同、某类节点数变了 > 2% ⇒ 拒。
// ---------------------------------------------------------------------------
func TestV3RefusesWhenTypeHistogramShifts(t *testing.T) {
	// 直接构造表格（不经 Model）：身份同、没有父链（所以边两边都为空）、只有类型变。
	build := func(types []string) []FPEntry {
		out := make([]FPEntry, 0, len(types))
		for i, typ := range types {
			out = append(out, FPEntry{
				Kind: typ, Scope: 0, ID: i + 1, Role: "demo_t.f",
				ColName: fmt.Sprintf("f%03d", i+1), FieldID: strconv.Itoa(i + 1),
			})
		}
		return out
	}
	recTypes := make([]string, 0, 100)
	for i := 0; i < 50; i++ {
		recTypes = append(recTypes, "Label")
	}
	for i := 0; i < 50; i++ {
		recTypes = append(recTypes, "Edit")
	}
	liveTypes := append([]string(nil), recTypes...)
	for i := 0; i < 3; i++ { // 3/100 = 3% > 2%
		liveTypes[i] = "Button"
	}

	rec, live := build(recTypes), build(liveTypes)
	d, same := JudgeSameInterface(rec, live)
	if same {
		t.Fatalf("有 3%% 的节点类型变了却判同一界面：%+v", d)
	}
	if d.LiveCoverage != 1 || d.EdgeCoverage != 1 || d.ExtraShare != 0 {
		t.Fatalf("夹具不成立：这条测试只让 ⑤ 不同（身份 %v / 边 %v）", d.LiveCoverage, d.EdgeCoverage)
	}
	if d.TypeDiff <= FPTypeHistMaxDiff {
		t.Fatalf("类型直方图差异要如实大于门槛 %.2f：%v", FPTypeHistMaxDiff, d.TypeDiff)
	}
	if !containsStr(d.TypeChanged, "Label 50 → 47") || !containsStr(d.TypeChanged, "Button 0 → 3") {
		t.Fatalf("差异清单要写清哪类节点数变了：%v", d.TypeChanged)
	}
	if v := gate(fpSnapEntries("件", rec), fpSnapEntries("会话", live)); v.OK {
		t.Fatalf("类型直方图变了必须拒：%s", v.Report())
	}

	// 2% 以内（100 条里变 2 条 = 2%）⇒ 仍在门内，别把容差写死成 0。
	live2 := append([]string(nil), recTypes...)
	live2[0], live2[1] = "Button", "Button"
	if d2, ok := JudgeSameInterface(rec, build(live2)); !ok {
		t.Fatalf("2%% 的容差内该放行：%+v", d2)
	}
}

// ---------------------------------------------------------------------------
// 7) idDigest 不参与：身份不同但 id 摘要碰巧相同 ⇒ 必须拒。
// ---------------------------------------------------------------------------
func TestV3NeverPassesOnIDDigestAlone(t *testing.T) {
	// (a) 直接构造指纹结构体：idDigest 相同，三件套里身份不同。
	base := Fingerprint{Total: 3, Hash: "deadbeefdeadbeef", Version: FPVersion,
		Ident: "aaaa", IdentN: 3, IdentUniq: 3, Edges: "bbbb", EdgeN: 2, Types: "Edit=3"}
	other := base
	other.Ident = "cccc"
	if SameInterface(base, other) {
		t.Fatal("身份不同、只有 idDigest 相同 ⇒ 绝不许判同一界面（旧判据借尸还魂）")
	}
	if !SameInterface(base, base) {
		t.Fatal("三件套全同必须判同一界面")
	}
	// 旧指纹（没有 v3 三件套）判不了：不许因为两边都空就放行。
	legacyA := Fingerprint{Total: 3, Hash: "deadbeefdeadbeef"}
	if SameInterface(legacyA, legacyA) {
		t.Fatal("v2 指纹没有身份证据，不能拿来判同一界面")
	}

	// (b) 条目层面：种类/作用域/编号/角色全同，只有一个 colName 不同 ⇒ idDigest 碰巧相同。
	build := func(tail string) []FPEntry {
		es := make([]FPEntry, 0, 12)
		for i := 0; i < 12; i++ {
			col := fmt.Sprintf("b_%02d", i)
			if i == 11 {
				col = tail
			}
			es = append(es, FPEntry{Kind: "FormField", Scope: 0, ID: 400 + i, Role: "demo_t.col",
				ColName: col, FieldID: strconv.Itoa(i)})
		}
		return es
	}
	rec, live := build("b_b"), build("b_z")
	if FingerprintOf(rec).Hash != FingerprintOf(live).Hash {
		t.Fatal("夹具不成立：两份的 idDigest 必须碰巧相同")
	}
	d, same := JudgeSameInterface(rec, live)
	if same {
		t.Fatalf("身份不同 ⇒ 即便 idDigest 相同也必须拒：%+v", d)
	}
	if !containsStr(d.IdentUnexpected, "b_z#11") {
		t.Fatalf("「意外」栏要点名它：%v", d.IdentUnexpected)
	}
	v := gate(fpSnapEntries("件", rec), fpSnapEntries("会话", live))
	if v.OK {
		t.Fatalf("idDigest 相同不许放行：%s", v.Report())
	}
	if !strings.Contains(v.Report(), "b_z#11") {
		t.Fatalf("拒绝报文要点名具体多出来的身份：\n%s", v.Report())
	}
}

// ---------------------------------------------------------------------------
// 8) 旧 v2 指纹字段：读到就**忽略并记诊断**，不许崩、不许因此拒/放。
// ---------------------------------------------------------------------------
func TestV3IgnoresLegacyV2FingerprintField(t *testing.T) {
	dir := t.TempDir()
	// 一份 v2 时代的身份文件：指纹只有 total/hash（没有 version/ident/edges/types）。
	legacyJSON := `{"id":"fep-1","path":"r.r","started":"2026-10-15T17:00:31Z",` +
		`"fingerprint":{"total":3976,"hash":"493b95ef0b0df2af"}}`
	if err := os.WriteFile(filepath.Join(dir, "fep-1"+MetaSuffix), []byte(legacyJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	mt, err := ReadMeta(dir, "fep-1")
	if err != nil {
		t.Fatalf("旧 v2 身份文件必须还能读（不许崩）：%v", err)
	}
	if mt.Fingerprint.Version != 0 || mt.Fingerprint.IsV3() {
		t.Fatalf("v2 指纹不该被当成 v3：%+v", mt.Fingerprint)
	}
	if notice := LegacyFingerprintNotice(mt.Fingerprint); notice == "" {
		t.Fatal("读到旧 v2 指纹必须留一条诊断（说清它被忽略）")
	}

	// 判决不看它：条目相同、旧指纹的 hash 与现算的不同 ⇒ 仍然放行，并把诊断带上。
	rec, live := EntriesOf(v3Form(0, 12)), EntriesOf(v3Form(1000, 12))
	v := GateReplay(
		OfSnapshot(Snapshot{Label: "件", Entries: rec, FP: mt.Fingerprint}),
		OfSnapshot(Snapshot{Label: "会话", Entries: live, FP: FingerprintOf(live)}),
	)
	if !v.OK {
		t.Fatalf("旧指纹字段只当诊断，不许影响判决：%s", v.Report())
	}
	if fpOf(v3Form(0, 12)).Hash == mt.Fingerprint.Hash {
		t.Fatal("夹具不成立：旧指纹 hash 与现算的恰好相同，说明不了「忽略」")
	}
	if !containsSub(v.Diagnostics, "v2") {
		t.Fatalf("判决里要带「旧指纹已忽略」的诊断：%v", v.Diagnostics)
	}
}

// 9) 阈值与规模下限是冻结的契约值（A21）：任何一格被改（含改成 0）这条必红。
func TestV3ThresholdsAreFrozenContractValues(t *testing.T) {
	if FPVersion != 3 {
		t.Fatalf("指纹口径版本必须是 3，got %d", FPVersion)
	}
	if FPIdentJaccardMin != 0.98 {
		t.Fatalf("身份覆盖门槛必须是 0.98（契约冻结，不读配置），got %v", FPIdentJaccardMin)
	}
	if FPEdgeIsoMin != 0.98 {
		t.Fatalf("边覆盖门槛必须是 0.98，got %v", FPEdgeIsoMin)
	}
	if FPTypeHistMaxDiff != 0.02 {
		t.Fatalf("多余身份/类型差异上限必须是 0.02，got %v", FPTypeHistMaxDiff)
	}
	if FPEvidenceMinNodes != 10 {
		t.Fatalf("证据量绝对下限必须是 10 条节点，got %v", FPEvidenceMinNodes)
	}
	if FPEvidenceRecShare != 0.5 {
		t.Fatalf("证据量相对下限必须是 0.5×|rec|，got %v", FPEvidenceRecShare)
	}
}

// 10) ② 的两侧（0.98 落在哪）：100 条里换 1 个身份仍过，换 3 个就不过。
func TestV3CoverageThresholdSitsBetweenOneAndThreeExchangedIdentities(t *testing.T) {
	build := func(exchange int) []FPEntry {
		es := make([]FPEntry, 0, 100)
		for i := 0; i < 100; i++ {
			col := fmt.Sprintf("f%03d", i+1)
			if i < exchange {
				col = fmt.Sprintf("x%03d", i+1) // 换成录制件里没有的身份
			}
			es = append(es, FPEntry{Kind: "FormField", Scope: 0, ID: i + 1, Role: "demo_t.f",
				ColName: col, FieldID: strconv.Itoa(i + 1)})
		}
		return es
	}
	// 换 1 条：99/100 = 0.99 ≥ 0.98，多余 1% ≤ 2% ⇒ 过
	if d, ok := JudgeSameInterface(build(0), build(1)); !ok {
		t.Fatalf("0.98 这一侧（99/100）该放行：%+v", d)
	}
	// 换 3 条：97/100 = 0.97 < 0.98，多余 3% > 2% ⇒ 拒
	if d, ok := JudgeSameInterface(build(0), build(3)); ok {
		t.Fatalf("0.98 那一侧（97/100）该拒：%+v", d)
	}
}

// 11) ① 的绝对下限与相对下限都要能单独触发（10 条 vs 0.5×|rec|）。
func TestV3EvidenceFloorUsesBothAbsoluteAndRelativeBounds(t *testing.T) {
	rec := make([]FPEntry, 0, 40)
	for i := 0; i < 40; i++ {
		rec = append(rec, FPEntry{Kind: "FormField", Scope: 0, ID: i + 1, Role: "demo_t.f",
			ColName: fmt.Sprintf("f%02d", i), FieldID: strconv.Itoa(i)})
	}
	// 相对下限：rec 40 条 ⇒ 要求 live ≥ 20 条（> 绝对下限 10）
	if d, ok := JudgeSameInterface(rec, rec[:20]); !ok || !d.EvidenceOK {
		t.Fatalf("恰好一半（20/40）该算证据够：%+v", d)
	}
	// 差一条就低于相对下限
	if d, ok := JudgeSameInterface(rec, rec[:19]); ok || d.EvidenceOK {
		t.Fatalf("19/40 低于相对下限，该判证据不足：%+v", d)
	}
	// 绝对下限：rec 很小（6 条）时要求 live ≥ 10，相对下限会给出 3，绝对下限兜住
	if d, ok := JudgeSameInterface(rec[:6], rec[:3]); ok || d.EvidenceOK {
		t.Fatalf("rec 很小时绝对下限（10 条）必须兜住：%+v", d)
	}
}

func containsStr(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func containsSub(list []string, sub string) bool {
	for _, s := range list {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
