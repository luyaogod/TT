package form

// 返工第一项：`--recording`（原始录制件）那条路必须**真测**（架构师复核 §2 的校准：
// "本机无录制件"不成立 —— 仓库自带 `testdata/fep-recordings/` 三份）。
//
// 两条路都要覆盖：
//  ① 能读出对象并计数（"可比 N 条 / 一致 N 条"）；
//  ② 录制件缺失 / 不可解析 → 非 0 退出 + 明确错误。
//
// 用的是**真录制件**（不是手写流 —— 手写的流不算真流），两条证据链：
//
//	A. 真录制 × 合成表单：`id-reuse.txt` 里的 8 条 (colName, fieldId) 原样声明进一个小 .42f
//	   ⇒ 可比 8 / 一致 8；`openwin-with-dialogs.txt` 的 2 条真歧义（index{3,6} / count{4,7}）
//	   用来钉"歧义项落候选"。
//	B. 真录制 × 真表单（门控）：`id-reuse.txt` × `fixtures/forms/t100/q_pmaa001_3.42f` = **8/8**。
//
// 另外钉住 A15：**窗口名不得参与判定** —— 实测这三份夹具里 `m.WindowName` 给的是窗口**标题**
// （"单据别编号数据查询(q_ooba002_1)" / "重新整理"）或空串，按表单名推出来的 `w_<表单>` 一条都不
// 命中；若拿它当过滤器，上面两条证据会被滤成 0 条，然后报"没有字段对象" —— 那正是"由启发式决定的
// 静默拒绝"，A15 明令禁止。

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"tt/internal/fep"
	"tt/internal/output"
)

// recordingFixtureNames 是仓库自带的真录制夹具（`testdata/fep-recordings/`，随仓库走）。
const (
	recIDReuse  = "id-reuse.txt"
	recOpenWin  = "openwin-with-dialogs.txt"
	recStopFile = "program-stop.txt"
)

// recordingDir 把一份仓库夹具按 `fep` 的命名约定放进临时目录。
//
// 为什么要改名字：`fep.RecordingFiles` 只认 `*-rt2gdc.log`（现场录制件的既有命名），
// 而夹具为了能 diff/review 存成了 `.txt`。改的是文件名，内容一字不动。
func recordingDir(t *testing.T, fixture string) string {
	t.Helper()
	src := filepath.Join("..", "..", "..", "testdata", "fep-recordings", fixture)
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("读仓库自带录制夹具失败: %v\n（它是仓库本体的一部分，缺了就是仓库坏了 —— 这条**不许**跳过）", err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "fep-20261011-142206-2-rt2gdc.log"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// recFormXML 用一组 colName→fieldId 造一张小 .42f（合成件：录制件是真的，表单是不变量）。
//
// 键排序后输出，保证 XML 稳定（同一个 map 每次跑出来字节一致）。
func recFormXML(form string, fields map[string]int) string {
	cols := make([]string, 0, len(fields))
	for c := range fields {
		cols = append(cols, c)
	}
	sort.Strings(cols)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="ASCII"?>` + "\n")
	b.WriteString(`<Form name="` + form + `" build="1.0" spacing="normal" style="dialog4" width="40" height="5">` + "\n  <Grid width=\"40\" height=\"5\">\n")
	for _, c := range cols {
		b.WriteString(`    <TableColumn name="formonly.` + c + `" colName="` + c + `" fieldId="` +
			itoaSmall(fields[c]) + `"><Edit width="4" posX="1" posY="0"/></TableColumn>` + "\n")
	}
	b.WriteString("  </Grid>\n</Form>\n")
	return b.String()
}

func itoaSmall(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

// recFields 是 `id-reuse.txt` 里观测到的 8 条（真证据，见本文件头）：
// check=0 pmaa001_1=1 pmaal004_2=2 pmaal003_3=3 pmaa080_4=4 oocql004_5=5 index=6 count=7
func recFields() map[string]int {
	return map[string]int{
		"check": 0, "pmaa001_1": 1, "pmaal004_2": 2, "pmaal003_3": 3,
		"pmaa080_4": 4, "oocql004_5": 5, "index": 6, "count": 7,
	}
}

// ── ① 基线构建（真录制件）────────────────────────────────────────────────────

func TestLoadRecordingBaselineIDReuseFixture(t *testing.T) {
	b, err := LoadRecordingBaseline(recordingDir(t, recIDReuse))
	if err != nil {
		t.Fatalf("从真录制件建基线失败: %v", err)
	}
	want := recFields()
	if len(b.Exact) != len(want) {
		t.Fatalf("唯一观测 = %d 条, want %d（%v）", len(b.Exact), len(want), b.Exact)
	}
	for col, id := range want {
		if got, ok := b.Exact[col]; !ok || got != id {
			t.Errorf("%s = %d（在不在=%v）, want %d", col, got, ok, id)
		}
	}
	if len(b.Ambiguous) != 0 {
		t.Errorf("这份录制里没有同名异号，Ambiguous 该是空: %v", b.Ambiguous)
	}
}

func TestLoadRecordingBaselineReportsRealAmbiguity(t *testing.T) {
	b, err := LoadRecordingBaseline(recordingDir(t, recOpenWin))
	if err != nil {
		t.Fatalf("从真录制件建基线失败: %v", err)
	}
	// 实测：同一个 colName 在这份录制里被观测到两个不同编号（同一窗口里另有一张表单）。
	// 这**不能**静默取最后一个（"看起来对、其实错"），必须进候选集合。
	for col, want := range map[string][]int{"index": {3, 6}, "count": {4, 7}} {
		got := b.Ambiguous[col]
		if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
			t.Errorf("%s 候选 = %v, want %v", col, got, want)
		}
		if _, dup := b.Exact[col]; dup {
			t.Errorf("%s 既进了候选又进了唯一观测 —— 重名必须只进候选", col)
		}
	}
	if len(b.Exact) != 8 {
		t.Errorf("唯一观测 = %d 条, want 8（%v）", len(b.Exact), b.Exact)
	}
}

func TestRecordingBaselineWindowIsDiagnosticOnly(t *testing.T) {
	dir := recordingDir(t, recOpenWin)
	b, err := LoadRecordingBaseline(dir)
	if err != nil {
		t.Fatalf("建基线失败: %v", err)
	}
	if len(b.Exact)+len(b.Ambiguous) == 0 {
		t.Fatal("基线是空的 —— 多半是又拿窗口名过滤了（A15 禁止）")
	}
	// 这份录制属于 apmt500_wf，而它的窗口名是标题（单据别编号数据查询(q_ooba002_1)）；
	// 启发式给出来的 w_apmt500 **不在**观察到的窗口里。
	heuristic := WindowOf("apmt500_wf")
	for _, w := range b.WindowsSeen {
		if w == heuristic {
			t.Fatalf("观察到的窗口里竟然有启发式窗口 %q —— 这个反证不成立了，换一份夹具或换个断言", heuristic)
		}
	}
	if len(b.WindowsSeen) == 0 {
		t.Error("WindowsSeen 该留下诊断用的窗口名（不参与判定，但要能解释「为什么滤不出东西」）")
	}
}

// ── ① CLI：真录制件 → 可比 / 一致 ───────────────────────────────────────────

func TestFormVerifyRecordingRealFixtureCounts(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{forms: map[string][]byte{
		"axm/vrec": []byte(recFormXML("vrec", recFields())),
	}})

	out, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vrec",
		"--recording", recordingDir(t, recIDReuse), "--json")
	if err != nil {
		t.Fatalf("真录制件那条路必须跑通（8 条全部一致）: %v\n%s", err, out)
	}
	doc := decodeShow(t, out)
	if doc["comparable"] != float64(8) || doc["consistent"] != float64(8) || doc["ok"] != true {
		t.Errorf("可比/一致/ok = %v/%v/%v, want 8/8/true", doc["comparable"], doc["consistent"], doc["ok"])
	}
	if doc["windowDerived"] != true {
		t.Errorf("窗口名是按表单名推出来的，该在输出里标出来（只作诊断）: %v", doc["windowDerived"])
	}
}

func TestFormVerifyRecordingRealAmbiguityHitsCandidates(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{forms: map[string][]byte{
		"axm/vrec": []byte(recFormXML("vrec", recFields())),
	}})

	out, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vrec",
		"--recording", recordingDir(t, recOpenWin), "--json")
	// openwin 这份里没有 ooba002_1/oobxl003_2（表单没有 → 算"仅在语料里"），
	// 但 index=6 / count=7 都落在真歧义候选里 ⇒ 仍然过。
	if err != nil {
		t.Fatalf("歧义项都落候选时该过: %v\n%s", err, out)
	}
	doc := decodeShow(t, out)
	if doc["ambiguousHit"] != float64(2) {
		t.Errorf("歧义项命中候选 = %v, want 2（index / count 各一条）", doc["ambiguousHit"])
	}
	if doc["comparable"] != float64(6) || doc["consistent"] != float64(6) {
		t.Errorf("可比/一致 = %v/%v, want 6/6", doc["comparable"], doc["consistent"])
	}
}

func TestFormVerifyRecordingAmbiguousNotInCandidatesExitsSix(t *testing.T) {
	tempDataDir(t)
	bad := recFields()
	bad["index"] = 4 // 真候选是 {3,6}
	injectSource(t, &fakeSource{forms: map[string][]byte{"axm/vrec": []byte(recFormXML("vrec", bad))}})

	out, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vrec",
		"--recording", recordingDir(t, recOpenWin), "--json")
	if err == nil {
		t.Fatalf("表单编号没落在真候选里时必须非 0 退出:\n%s", out)
	}
	if got := exitOf(t, err); got != exitMismatch {
		t.Errorf("退出码 = %d, want %d", got, exitMismatch)
	}
	doc := decodeShow(t, out)
	if len(doc["ambiguousBad"].([]any)) != 1 {
		t.Errorf("该报出 1 条歧义没落候选: %v", doc["ambiguousBad"])
	}
}

func TestFormVerifyRecordingMismatchExitsSix(t *testing.T) {
	tempDataDir(t)
	bad := recFields()
	bad["check"] = 9
	injectSource(t, &fakeSource{forms: map[string][]byte{"axm/vrec": []byte(recFormXML("vrec", bad))}})

	out, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vrec",
		"--recording", recordingDir(t, recIDReuse), "--format", "table")
	if err == nil {
		t.Fatalf("不一致时必须非 0 退出:\n%s", out)
	}
	if got := exitOf(t, err); got != exitMismatch {
		t.Errorf("退出码 = %d, want %d", got, exitMismatch)
	}
	if !strings.Contains(out, "可比 8 条 / 一致 7 条") || !strings.Contains(out, "check") {
		t.Errorf("报告该打出「可比 8 条 / 一致 7 条」并点名 check:\n%s", out)
	}
}

// ── ② 缺失 / 不可解析 ───────────────────────────────────────────────────────

func TestFormVerifyRecordingMissingIsNonZero(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{forms: map[string][]byte{"axm/vrec": []byte(recFormXML("vrec", recFields()))}})

	_, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vrec",
		"--recording", filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Fatal("录制目录不存在必须非 0 退出")
	}
	if got := exitOf(t, err); got != exitFormUnknown {
		t.Errorf("退出码 = %d, want %d（FORM_UNKNOWN）", got, exitFormUnknown)
	}
	// 可操作的下一步在 **Hint** 里（C4.3），不是散文正文
	var oe *output.Error
	if !errors.As(err, &oe) || !strings.Contains(oe.Hint, "--recording") {
		t.Errorf("Hint 里该给出可操作的下一步（--recording / --map）: %+v", err)
	}
}

func TestFormVerifyRecordingUnparsableIsNonZero(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{forms: map[string][]byte{"axm/vrec": []byte(recFormXML("vrec", recFields()))}})

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "fep-broken-2-rt2gdc.log"), []byte("这不是录制件\n啥也不是\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vrec",
		"--recording", dir)
	if err == nil {
		t.Fatal("录制件读不出消息必须非 0 退出（不许当成零可比）")
	}
	if got := exitOf(t, err); got != exitFormUnknown {
		t.Errorf("退出码 = %d, want %d", got, exitFormUnknown)
	}
	if !strings.Contains(err.Error(), "录制件") {
		t.Errorf("错误该说清是录制件那一侧的问题: %v", err)
	}
}

// ── B. 真录制 × 真表单（门控：真 .42f 在成果目录里，不在仓库里）──────────────

func TestFormVerifyRecordingAndRealFormGated(t *testing.T) {
	forms := filepath.Join("..", "..", "..", ".agents", "Genero Original Factory Compatible",
		"fixtures", "forms", "t100")
	if _, err := os.Stat(filepath.Join(forms, "q_pmaa001_3.42f")); err != nil {
		t.Skipf("没有真表单样本（%v）—— 设 TT_FORM_SAMPLE_DIR 指向 fixtures/forms/", err)
	}
	tempDataDir(t)
	out, err := runForm(t, "verify", "--env", "t35prd", "--module", "t100", "--form", "q_pmaa001_3",
		"--recording", recordingDir(t, recIDReuse), "--dir", filepath.Dir(forms), "--json")
	if err != nil {
		t.Fatalf("真录制件 × 真表单必须 100%% 一致: %v\n%s", err, out)
	}
	doc := decodeShow(t, out)
	if doc["comparable"] != float64(8) || doc["consistent"] != float64(8) {
		t.Errorf("可比/一致 = %v/%v, want 8/8（实测 id-reuse.txt × q_pmaa001_3.42f）",
			doc["comparable"], doc["consistent"])
	}
	t.Logf("真录制 × 真表单：可比 %v / 一致 %v", doc["comparable"], doc["consistent"])
}

// ── A15：启发式只作诊断，不参与命中/拒绝 ────────────────────────────────────

func TestVariantHeuristicOnlyDiagnoses(t *testing.T) {
	f, err := fep.ParseForm42F("vform_tui", recFormXML("vform_tui", map[string]int{"f1": 1, "f2": 2}))
	if err != nil {
		t.Fatalf("解析夹具失败: %v", err)
	}
	r := Compare(f, &Baseline{Exact: map[string]int{"f1": 1, "f2": 2}})
	if !r.OK() || r.Comparable != 2 {
		t.Fatalf("变体后缀是启发式，不该影响判定: %+v", r)
	}
	if r.Variant != "tui" {
		t.Errorf("Variant = %q, want tui（只作诊断用的那一半要留着）", r.Variant)
	}
}

func TestWindowHeuristicNeverDecides(t *testing.T) {
	tempDataDir(t)
	injectSource(t, vformSource())

	// ① 推到语料里没有的窗口：拒绝的理由必须是"你点名的窗口不在语料里"（用法错 + 列出候选），
	//    而不是"表单名不匹配"之类的数据判定。
	_, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vform",
		"--map", filepath.Join("testdata", "verify", "corpus_ok.json"))
	if err == nil {
		t.Fatal("推到语料里没有的窗口时该报用法错")
	}
	if got := exitOf(t, err); got != exitUsage {
		t.Errorf("退出码 = %d, want %d", got, exitUsage)
	}
	if !strings.Contains(err.Error(), "w_vdemo") {
		t.Errorf("错误里该列出语料里真实存在的窗口: %v", err)
	}

	// ② 显式给出正确窗口：照常全绿 —— 判定只由数据（colName→fieldId）决定。
	out, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vform",
		"--window", "w_vdemo", "--map", filepath.Join("testdata", "verify", "corpus_ok.json"), "--json")
	if err != nil {
		t.Fatalf("显式给对窗口时必须照常过: %v\n%s", err, out)
	}
	doc := decodeShow(t, out)
	if doc["windowDerived"] != nil {
		t.Errorf("显式给了 --window 就不该标 windowDerived: %v", doc["windowDerived"])
	}
}
