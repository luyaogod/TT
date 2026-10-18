package form

// `tt form verify`：把**离线解析出来的 `colName→fieldId`** 与基线逐条比对（契约 C5.1 / C7.4 / 变更记录 A5）。
//
// 基线有两处来源，优先级写死：
//  1. `--recording <目录>`：原始录制件（**优先**，A5 原文「若同时也给了录制目录，优先用录制」）；
//  2. 否则 `--map <文件>` / `TT_FORM_CORPUS_MAP`：归档语料表（A5 的等价基线，可离线跑）。
//
// 口径（照抄 `internal/form42f/form42f_test.go` 的 TestFormMatchesCorpus，避免两处漂移）：
//   - **可比** = 语料里有、这张表单里也有的 colName → **必须相等**；
//   - **仅在语料里** = 语料有、表单没有 → 允许（同一个窗口里还有别的表单/开窗/动态节点）；
//   - 语料里标了 `ambiguous` 的 → 表单里的编号必须**落在候选里**。
//
// 真数据走门控（`TestFormVerifyCorpusGated`），仓库内只放合成小语料。

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/form42f"
	"tt/internal/testkit"
)

// vformXML 是 verify 用例的合成夹具：四个可寻址字段（f1…f4）。
//
// 内联（而不是再落一个 .xml 文件）是因为这块 XML 只有这一组用例用，内联能一眼看出
// 「语料里写了 f3=9，表单里 f3=3」这种比对关系（B14 的既有先例）。
const vformXML = `<?xml version="1.0" encoding="ASCII"?>
<Form name="vform.tmp" build="9.9.9" spacing="normal" style="dialog4" width="20" height="5">
  <Grid width="20" height="5">
    <FormField name="formonly.f1" colName="f1" fieldId="1" tabIndex="1"><Edit width="4" posX="1" posY="0"/></FormField>
    <FormField name="formonly.f2" colName="f2" fieldId="2" tabIndex="2"><Edit width="4" posX="1" posY="1"/></FormField>
    <FormField name="formonly.f3" colName="f3" fieldId="3" tabIndex="3"><Edit width="4" posX="1" posY="2"/></FormField>
    <FormField name="formonly.f4" colName="f4" fieldId="4" tabIndex="4"><Edit width="4" posX="1" posY="3"/></FormField>
  </Grid>
</Form>`

func vformSource() *fakeSource {
	return &fakeSource{forms: map[string][]byte{"axm/vform": []byte(vformXML)}}
}

// ── 比对核心（纯函数）────────────────────────────────────────────────────────

func TestCompareCounts(t *testing.T) {
	f, err := form42f.ParseForm42F("vform", vformXML)
	if err != nil {
		t.Fatalf("解析夹具失败: %v", err)
	}
	b := &Baseline{
		Window:    "w_vdemo",
		Exact:     map[string]int{"f1": 1, "f2": 2, "f3": 9, "f9": 77},
		Ambiguous: map[string][]int{"f4": {4, 14}},
	}
	r := Compare(f, b)

	if r.Comparable != 3 { // f1 f2 f3
		t.Errorf("可比 = %d, want 3（f1/f2/f3 在两边都有）", r.Comparable)
	}
	if r.Consistent != 2 { // f1 f2
		t.Errorf("一致 = %d, want 2", r.Consistent)
	}
	if len(r.Mismatches) != 1 || r.Mismatches[0].ColName != "f3" {
		t.Errorf("不一致的清单 = %+v, want 只有 f3（语料 9 / 表单 3）", r.Mismatches)
	}
	if r.CorpusOnly != 1 { // f9
		t.Errorf("仅在语料里 = %d, want 1（f9）", r.CorpusOnly)
	}
	if r.AmbiguousHit != 1 || len(r.AmbiguousBad) != 0 {
		t.Errorf("歧义项：命中候选 %d / 没命中 %v，want 1/空（表单 f4=4 在候选 [4 14] 里）",
			r.AmbiguousHit, r.AmbiguousBad)
	}
	if r.OK() {
		t.Error("有不一致时 OK() 必须是假 —— 否则 verify 会退 0")
	}
	if !strings.Contains(r.Summary(), "可比 3 条 / 一致 2 条") {
		t.Errorf("摘要必须能一眼读出「可比 N 条 / 一致 N 条」: %q", r.Summary())
	}
}

func TestCompareAllConsistent(t *testing.T) {
	f, err := form42f.ParseForm42F("vform", vformXML)
	if err != nil {
		t.Fatalf("解析夹具失败: %v", err)
	}
	r := Compare(f, &Baseline{Window: "w_vdemo", Exact: map[string]int{"f1": 1, "f2": 2, "f3": 3}})
	if r.Comparable != 3 || r.Consistent != 3 {
		t.Errorf("可比/一致 = %d/%d, want 3/3", r.Comparable, r.Consistent)
	}
	if !r.OK() {
		t.Errorf("全一致时 OK() 必须为真: %s", r.Summary())
	}
}

func TestCompareAmbiguousNotInCandidates(t *testing.T) {
	f, err := form42f.ParseForm42F("vform", vformXML)
	if err != nil {
		t.Fatalf("解析夹具失败: %v", err)
	}
	// 表单里 f4=4，语料候选只给了 [14 24] → 必须报出来（这正是「两边不是同一张表单」的信号）
	r := Compare(f, &Baseline{Window: "w_vdemo", Ambiguous: map[string][]int{"f4": {14, 24}}})
	if len(r.AmbiguousBad) != 1 || r.AmbiguousBad[0].ColName != "f4" {
		t.Fatalf("歧义项没落在候选里必须报出来: %+v", r.AmbiguousBad)
	}
	if r.OK() {
		t.Error("歧义项都没命中候选时 OK() 必须为假")
	}
}

// ── 归档语料表的读取 ─────────────────────────────────────────────────────────

func TestLoadCorpusMap(t *testing.T) {
	c, err := LoadCorpusMap(filepath.Join("testdata", "verify", "corpus_small.json"))
	if err != nil {
		t.Fatalf("读合成语料失败: %v", err)
	}
	w, ok := c.ByWindow["w_vdemo"]
	if !ok {
		t.Fatalf("语料里的窗口没读出来: %v", c.Windows())
	}
	if w.Exact["f1"] != 1 || w.Exact["f3"] != 9 {
		t.Errorf("唯一观测没读对: %+v", w.Exact)
	}
	if got := w.Ambiguous["f4"]; len(got) != 2 || got[0] != 4 || got[1] != 14 {
		t.Errorf("歧义候选没读对: %+v", w.Ambiguous)
	}
	if !strings.Contains(c.Windows(), "w_vdemo") {
		t.Errorf("Windows() 该列出可用窗口: %q", c.Windows())
	}
}

func TestLoadCorpusMapRejectsBroken(t *testing.T) {
	p := filepath.Join(t.TempDir(), "broken.json")
	if err := os.WriteFile(p, []byte("{ 这不是 JSON"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCorpusMap(p); err == nil {
		t.Fatal("坏语料必须报错（不许当成「没有可比条目」静默通过）")
	}
	if _, err := LoadCorpusMap(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("语料文件不存在必须报错")
	}
}

// ── CLI：verify ─────────────────────────────────────────────────────────────

func TestFormVerifyCLIReportsCounts(t *testing.T) {
	tempDataDir(t)
	injectSource(t, vformSource())

	out, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vform",
		"--window", "w_vdemo", "--map", filepath.Join("testdata", "verify", "corpus_small.json"),
		"--format", "table")
	if err == nil {
		t.Fatalf("语料里有不一致（f3: 语料 9 / 表单 3）时必须非 0 退出:\n%s", out)
	}
	if got := exitOf(t, err); got != exitMismatch {
		t.Errorf("退出码 = %d, want %d（体检不过）", got, exitMismatch)
	}
	if !strings.Contains(out, "可比 3 条 / 一致 2 条") {
		t.Errorf("必须打印「可比 N 条 / 一致 N 条」:\n%s", out)
	}
	if !strings.Contains(out, "f3") {
		t.Errorf("不一致的那一条要点名:\n%s", out)
	}
}

func TestFormVerifyAllConsistentExitsZero(t *testing.T) {
	tempDataDir(t)
	injectSource(t, vformSource())

	out, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vform",
		"--window", "w_vdemo", "--map", filepath.Join("testdata", "verify", "corpus_ok.json"),
		"--json")
	if err != nil {
		t.Fatalf("全一致时必须退 0: %v\n%s", err, out)
	}
	doc := decodeShow(t, out)
	if doc["comparable"] != float64(3) || doc["consistent"] != float64(3) || doc["ok"] != true {
		t.Errorf("信封不对: %v", doc)
	}
}

func TestFormVerifyWithoutBaselineIsUsageError(t *testing.T) {
	tempDataDir(t)
	injectSource(t, vformSource())
	t.Setenv("TT_FORM_CORPUS_MAP", "")

	_, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vform")
	if err == nil {
		t.Fatal("没有基线必须报错（不许「没得比」读成成功）")
	}
	if got := exitOf(t, err); got != exitUsage {
		t.Errorf("退出码 = %d, want %d（用法错）", got, exitUsage)
	}
	if !strings.Contains(err.Error(), "--map") {
		t.Errorf("提示里该给出 --map 这条出路: %v", err)
	}
}

func TestFormVerifyWindowMissingFromCorpusIsUsageError(t *testing.T) {
	tempDataDir(t)
	injectSource(t, vformSource())

	_, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vform",
		"--window", "w_nope", "--map", filepath.Join("testdata", "verify", "corpus_small.json"))
	if err == nil {
		t.Fatal("语料里没有这个窗口必须报错，不能把零可比算成通过")
	}
	if got := exitOf(t, err); got != exitUsage {
		t.Errorf("退出码 = %d, want %d", got, exitUsage)
	}
	if !strings.Contains(err.Error(), "w_vdemo") {
		t.Errorf("错误里该列出语料里有哪些窗口: %v", err)
	}
}

func TestFormVerifyFetchFailureIsNonZero(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{err: context.DeadlineExceeded})

	_, err := runForm(t, "verify", "--env", "demoenv", "--module", "axm", "--form", "vform",
		"--window", "w_vdemo", "--map", filepath.Join("testdata", "verify", "corpus_small.json"))
	if err == nil {
		t.Fatal("取件失败必须非 0 退出")
	}
	if got := exitOf(t, err); got != exitFormUnknown {
		t.Errorf("退出码 = %d, want %d", got, exitFormUnknown)
	}
}

// 反证：把「比对」退化成「只数不比」这条测试必须红 —— 不一致一定要逐条留下，
// 且在 JSON 里看得见（agent 拿到的是 JSON，不是 Go 结构体）。
func TestCompareNeverSwallowsMismatch(t *testing.T) {
	f, err := form42f.ParseForm42F("vform", vformXML)
	if err != nil {
		t.Fatalf("解析夹具失败: %v", err)
	}
	r := Compare(f, &Baseline{Exact: map[string]int{"f1": 1, "f2": 22, "f3": 33}})
	if len(r.Mismatches) != 2 || r.Consistent != 1 {
		t.Fatalf("不一致必须逐条留下：%+v", r)
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"f2"`) || !strings.Contains(string(b), `"f3"`) {
		t.Errorf("JSON 里丢了不一致的条目: %s", b)
	}
}

// ── 门控：真样本 + 真归档语料（验收口径：w_axmt500 可比 509 条 → 必须 100%）────

// wantComparableAxmt500 是 M1 的验收数字（契约 §0 的 B2 / C7.4：本机 509/509）。
const wantComparableAxmt500 = 509

func TestFormVerifyCorpusGated(t *testing.T) {
	root := testkit.RepoRoot(t)
	samples := os.Getenv("TT_FORM_SAMPLE_DIR")
	if samples == "" {
		samples = filepath.Join(root, ".agents", "Genero Original Factory Compatible", "fixtures", "forms")
	}
	corpus := os.Getenv("TT_FORM_CORPUS_MAP")
	if corpus == "" {
		corpus = filepath.Join(root, ".agents", "Genero Original Factory Compatible",
			"out", "semantic", "colname-fieldid.json")
	}
	if _, err := os.Stat(filepath.Join(samples, "t100", "axmt500_wf.42f")); err != nil {
		t.Skipf("没有真样本（%v）—— 设 TT_FORM_SAMPLE_DIR 指向 fixtures/forms/", err)
	}
	if _, err := os.Stat(corpus); err != nil {
		t.Skipf("没有归档语料表（%v）—— 设 TT_FORM_CORPUS_MAP 指向 colname-fieldid.json", err)
	}

	tempDataDir(t)
	out, err := runForm(t, "verify",
		"--env", "t35prd", "--module", "axm", "--form", "axmt500_wf",
		"--window", "w_axmt500", "--map", corpus, "--dir", samples, "--json")
	if err != nil {
		t.Fatalf("门控 verify 失败: %v\n%s", err, out)
	}

	doc := decodeShow(t, out)
	comparable := int(doc["comparable"].(float64))
	consistent := int(doc["consistent"].(float64))
	if comparable != wantComparableAxmt500 {
		t.Errorf("可比 = %d, want %d（验收口径写死 509 —— 变化必须显式说明是哪一边变了）",
			comparable, wantComparableAxmt500)
	}
	if consistent != comparable || doc["ok"] != true {
		t.Errorf("必须 100%% 一致（不做百分比容忍）：一致 %d / 可比 %d\n%s", consistent, comparable, out)
	}
	t.Logf("门控：语料 %s + 样本 %s → 可比 %d 条 / 一致 %d 条", corpus, samples, comparable, consistent)
}