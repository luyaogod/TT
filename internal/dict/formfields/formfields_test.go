package formfields

// M3 第一步（dict 侧复用字段表）的**测试先行**。
//
// 契约口径（唯一口径文件 `.agents/plan/tt-genero-formdef-contract.md`）：
//   - C1.8 / A15 / A18：**启发式不得参与判定**，取不到就写"未解析"，绝不拿 colName 充当中文名；
//   - C7.2：仓库内只放合成夹具，真样本走 `TT_FORM_SAMPLE_DIR`，缺失就 `Skip` 并写明原因；
//   - C5.2：`tt form show` 的 JSON 键冻结 —— 本包**另起**键名，不动那边一个键。
//
// 反证钉子：`TestMerge_UnresolvedNeverFallsBackToColName` —— 把"未解析"改成"回落 colName"必红。

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"tt/internal/dict/db"
	"tt/internal/form42f"
)

// ── 假字典（不连真机、不读 SQLite） ──────────────────────────────────────────

// fakeNamer 是字典侧"字段中文名"的假件：表 → 字段 → 中文名。
// 真件是 internal/dict/db 的 *DB（本地镜像）或 live.Live（远程），两者都满足 FieldNamer。
type fakeNamer struct {
	tables map[string]map[string]string // 表名 → 字段名 → 中文名
	fails  map[string]error             // 表名 → 强制返回的错误
	calls  []string                     // 问过哪些表（断言"没乱查"）
}

func (f *fakeNamer) QueryTable(tableName string) ([]db.TableInfo, error) {
	f.calls = append(f.calls, tableName)
	if err, ok := f.fails[tableName]; ok {
		return nil, err
	}
	fields, ok := f.tables[tableName]
	if !ok {
		return nil, nil
	}
	names := make([]string, 0, len(fields))
	for n := range fields {
		names = append(names, n)
	}
	sort.Strings(names)
	rows := make([]db.TableInfo, 0, len(names))
	for i, n := range names {
		rows = append(rows, db.TableInfo{
			Seq:       strconv.Itoa(i + 1),
			FieldName: n,
			FieldDesc: fields[n],
		})
	}
	return rows, nil
}

// ── 夹具 ────────────────────────────────────────────────────────────────────

// synthForm 合成夹具：一条 TableColumn（comment/text 键齐全）+ 一个 RecordView 声明它属于 demo_t。
const synthForm = `<?xml version="1.0" encoding="UTF-8"?>
<Form name="demofm.tmp" build="3.21.03">
  <TableColumn name="formonly.b_demo001" colName="b_demo001" sqlType="SMALLINT" fieldId="7" sqlTabName="formonly" tabIndex="1" text="lbl_demo001">
    <Edit width="8" comment="cmt_demo001"><LStr comment="cmt_demo001"/></Edit>
  </TableColumn>
  <RecordView tabName="demo_t">
    <Link colName="b_demo001" fieldIdRef="7"/>
  </RecordView>
</Form>`

// synthFormNoTable 同一张表单，但**没有任何表可依**（既无 RecordView，也无 sqlTabName）。
const synthFormNoTable = `<?xml version="1.0" encoding="UTF-8"?>
<Form name="demofm.tmp" build="3.21.03">
  <FormField name="formonly.b_demo001" colName="b_demo001" sqlType="SMALLINT" fieldId="7" tabIndex="1" text="lbl_demo001">
    <Edit width="8" comment="cmt_demo001"/>
  </FormField>
</Form>`

// synthFormTwoTables 同一 colName 被两个 RecordView 声明 → 两张表都可能，字典若各说各话就不许挑。
const synthFormTwoTables = `<?xml version="1.0" encoding="UTF-8"?>
<Form name="demofm.tmp" build="3.21.03">
  <TableColumn name="formonly.b_demo001" colName="b_demo001" sqlType="SMALLINT" fieldId="7" sqlTabName="formonly" tabIndex="1" text="lbl_demo001">
    <Edit width="8" comment="cmt_demo001"/>
  </TableColumn>
  <RecordView tabName="demo_t">
    <Link colName="b_demo001" fieldIdRef="7"/>
  </RecordView>
  <RecordView tabName="demo2_t">
    <Link colName="b_demo001" fieldIdRef="7"/>
  </RecordView>
</Form>`

// parseSynth 解析合成夹具（复用 form42f 的解析器，绝不另写一份解析）。
func parseSynth(t *testing.T, text string) *form42f.F42Form {
	t.Helper()
	f, err := form42f.ParseForm42F("demofm", text)
	if err != nil {
		t.Fatalf("解析合成夹具失败: %v", err)
	}
	return f
}

// sampleDir 真样本目录（仓库外）：`fixtures/forms/`。缺失就 Skip 并写明原因（C7.3）。
func sampleDir(t *testing.T) string {
	t.Helper()
	if d := os.Getenv("TT_FORM_SAMPLE_DIR"); d != "" {
		if st, err := os.Stat(d); err != nil || !st.IsDir() {
			t.Skipf("TT_FORM_SAMPLE_DIR=%s 不是目录（%v）——跳过真样本测试", d, err)
		}
		return d
	}
	d := filepath.Join(repoRoot(t), ".agents", "Genero Original Factory Compatible", "fixtures", "forms")
	if st, err := os.Stat(d); err != nil || !st.IsDir() {
		t.Skipf("真样本目录不存在（%s）——跳过：仓库内只放合成夹具（契约 C7.2/C7.3）", d)
	}
	return d
}

// repoRoot 从包目录上溯到 go.mod 所在目录。
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("取工作目录失败: %v", err)
	}
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("上溯不到仓库根（go.mod）")
	return ""
}

func findField(t *testing.T, tab *Table, colName string) Field {
	t.Helper()
	for _, f := range tab.Fields {
		if f.ColName == colName {
			return f
		}
	}
	t.Fatalf("字段表里没有 colName=%s（共 %d 条）", colName, len(tab.Fields))
	return Field{}
}

// ── 1. 合成夹具：命中字典就填中文名 ──────────────────────────────────────────

func TestMerge_ResolvesLabelFromDictTable(t *testing.T) {
	n := &fakeNamer{tables: map[string]map[string]string{
		"demo_t": {"demo001": "演示栏位"},
	}}
	tab := Merge(parseSynth(t, synthForm), "demomod", n)
	if len(tab.Fields) != 1 {
		t.Fatalf("字段条目 = %d, want 1", len(tab.Fields))
	}
	f := tab.Fields[0]
	if f.LabelState != LabelResolved || f.Label != "演示栏位" {
		t.Errorf("label=%q state=%s，想要 resolved/演示栏位", f.Label, f.LabelState)
	}
	if f.LabelFrom != "demo_t.demo001" {
		t.Errorf("labelFrom = %q，想要 demo_t.demo001（表.字段，可回溯）", f.LabelFrom)
	}
	if f.Comment != "cmt_demo001" || f.TextKey != "lbl_demo001" {
		t.Errorf("comment/textKey 原样带出失败: %q / %q", f.Comment, f.TextKey)
	}
	if tab.LabelSource != LabelSourceDict {
		t.Errorf("labelSource = %q, want %q", tab.LabelSource, LabelSourceDict)
	}
}

// ── 2. 反证钉子：字典里没有这个键 → 必须"未解析"，且**绝不**回落成 colName ────

func TestMerge_UnresolvedNeverFallsBackToColName(t *testing.T) {
	// 字典里有 demo_t，但**没有** demo001 这个字段（只有 demo999）。
	n := &fakeNamer{tables: map[string]map[string]string{
		"demo_t": {"demo999": "别的栏位"},
	}}
	tab := Merge(parseSynth(t, synthForm), "demomod", n)
	f := tab.Fields[0]

	if f.Label != "" {
		t.Errorf("label = %q，字典里查不到这个键就必须留空（不许猜）", f.Label)
	}
	if f.LabelState != LabelUnresolved {
		t.Errorf("labelState = %q，想要 %q（取不到就标未解析）", f.LabelState, LabelUnresolved)
	}
	if f.LabelReason != ReasonNotInDict {
		t.Errorf("labelReason = %q，想要 %q", f.LabelReason, ReasonNotInDict)
	}
	// 这一组是反证的核心：任何"回落成身份"的做法都会在这里红。
	for _, forbidden := range []string{"b_demo001", "demo001", "cmt_demo001", "lbl_demo001", "demo_t.demo001"} {
		if f.Label == forbidden {
			t.Errorf("label 回落成了 %q —— 中文名列只允许放字典给出的名字", forbidden)
		}
	}
}

// ── 3. 两张表各说各话 → 不挑，标"未解析（多个候选）" ────────────────────────

func TestMerge_AmbiguousDictStaysUnresolved(t *testing.T) {
	n := &fakeNamer{tables: map[string]map[string]string{
		"demo_t":  {"demo001": "甲"},
		"demo2_t": {"demo001": "乙"},
	}}
	tab := Merge(parseSynth(t, synthFormTwoTables), "demomod", n)
	f := tab.Fields[0]
	if f.Label != "" || f.LabelState != LabelUnresolved || f.LabelReason != ReasonAmbiguous {
		t.Errorf("label=%q state=%s reason=%s，想要空/unresolved/%s（多个候选不许挑一个）",
			f.Label, f.LabelState, f.LabelReason, ReasonAmbiguous)
	}
}

// ── 4. 没有字典数据源 / 没有可依的表 → 未解析，两种原因要分得开 ───────────────

func TestMerge_NoNamerMarksUnresolved(t *testing.T) {
	tab := Merge(parseSynth(t, synthForm), "demomod", nil)
	f := tab.Fields[0]
	if f.Label != "" || f.LabelState != LabelUnresolved || f.LabelReason != ReasonNoDict {
		t.Errorf("label=%q state=%s reason=%s，想要空/unresolved/%s",
			f.Label, f.LabelState, f.LabelReason, ReasonNoDict)
	}
	if tab.LabelSource != LabelSourceNone {
		t.Errorf("labelSource = %q, want %q", tab.LabelSource, LabelSourceNone)
	}
}

func TestMerge_NoTableIsUnresolved(t *testing.T) {
	n := &fakeNamer{tables: map[string]map[string]string{"demo_t": {"demo001": "演示栏位"}}}
	tab := Merge(parseSynth(t, synthFormNoTable), "demomod", n)
	f := tab.Fields[0]
	if f.Label != "" || f.LabelState != LabelUnresolved || f.LabelReason != ReasonNoTable {
		t.Errorf("label=%q state=%s reason=%s，想要空/unresolved/%s",
			f.Label, f.LabelState, f.LabelReason, ReasonNoTable)
	}
	if len(n.calls) != 0 {
		t.Errorf("没有表可依时不该去查字典（问过 %v）", n.calls)
	}
}

// ── 5. 字典查表报错 → 未解析 + 记 notes，不冒充命中 ─────────────────────────

func TestMerge_DictErrorIsNotAName(t *testing.T) {
	n := &fakeNamer{fails: map[string]error{"demo_t": errBrokenDictStub{}}}
	tab := Merge(parseSynth(t, synthForm), "demomod", n)
	f := tab.Fields[0]
	if f.Label != "" || f.LabelState != LabelUnresolved {
		t.Errorf("label=%q state=%s，查表失败必须未解析", f.Label, f.LabelState)
	}
	if len(tab.Notes) == 0 {
		t.Errorf("查表失败要留 notes（否则「为什么没中文名」只能靠猜）")
	}
}

// TestMerge_MissingDictTableStaysQuiet 字典里没这张表是**正常**情形（明细列的
// sqlTabName 常写 formonly/undefined）：仍然未解析，但**不刷 notes** ——
// 563 行表单刷 500 条噪声 note，真正的错误就被埋了。
func TestMerge_MissingDictTableStaysQuiet(t *testing.T) {
	n := &fakeNamer{fails: map[string]error{
		"demo_t":   errMissingTableStub{},
		"formonly": errMissingTableStub{},
	}}
	tab := Merge(parseSynth(t, synthForm), "demomod", n)
	f := tab.Fields[0]
	if f.Label != "" || f.LabelState != LabelUnresolved {
		t.Errorf("label=%q state=%s，查表失败必须未解析", f.Label, f.LabelState)
	}
	if len(tab.Notes) != 0 {
		t.Errorf("缺表不该刷 notes，实际: %v", tab.Notes)
	}
	if len(n.calls) == 0 {
		t.Errorf("有表可依就该真去查（证明上面是「查了但没名字」，不是「没查」）")
	}
}

// errMissingTableStub 模拟"字典里没这张表"的底层错误（db.IsMissingTable 认得）。
type errMissingTableStub struct{}

func (errMissingTableStub) Error() string { return "no such table: dzeb_t" }

// errBrokenDictStub 模拟真正的故障（连接断、权限、SQL 错）—— 这类必须进 notes。
type errBrokenDictStub struct{}

func (errBrokenDictStub) Error() string { return "connection reset by peer" }

// ── 6. 路径候选：`<模块>/42f/<表单>.42f` 与 `<模块>/<表单>.42f` 都要认 ───────

func TestFindPath_AcceptsBothLayouts(t *testing.T) {
	root := t.TempDir()

	with42f := filepath.Join(root, "mod42f", "42f")
	if err := os.MkdirAll(with42f, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(with42f, "form_a.42f"), []byte(synthForm), 0o644); err != nil {
		t.Fatal(err)
	}
	p, err := FindPath(root, "mod42f", "form_a")
	if err != nil {
		t.Fatalf("42f 子目录布局没找到: %v", err)
	}
	if filepath.Base(filepath.Dir(p)) != "42f" {
		t.Errorf("找到的路径 = %s，想要 .../mod42f/42f/form_a.42f", p)
	}

	flat := filepath.Join(root, "modflat")
	if err := os.MkdirAll(flat, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(flat, "form_b.42f"), []byte(synthForm), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := FindPath(root, "modflat", "form_b"); err != nil {
		t.Errorf("真样本的扁平布局（<模块>/<表单>.42f）没找到: %v", err)
	}

	_, err = FindPath(root, "modflat", "form_missing")
	if err == nil {
		t.Fatalf("缺文件必须报错")
	}
	if !strings.Contains(err.Error(), "form_missing.42f") || !strings.Contains(err.Error(), "42f") {
		t.Errorf("错误里要列出尝试过的路径，实际: %v", err)
	}
}

// ── 7. 真样本门控：axmt500_wf.42f 的 b_xmda001 ──────────────────────────────

func TestLoadFormAxmt500FieldTable(t *testing.T) {
	dir := sampleDir(t)

	// 假字典：只给 xmda_t 的两个字段 —— 一个能解析、一个查不到（两条路都要断言）。
	n := &fakeNamer{tables: map[string]map[string]string{
		"xmda_t": {"xmda001": "料号", "b_xmda001": "不该命中这行"},
	}}
	tab, err := LoadForm(dir, "t100", "axmt500_wf", n)
	if err != nil {
		t.Fatalf("读真样本失败: %v", err)
	}
	if tab.Form != "axmt500_wf" || tab.Module != "t100" {
		t.Errorf("module/form = %q/%q, want t100/axmt500_wf", tab.Module, tab.Form)
	}
	if len(tab.Fields) != 563 {
		t.Errorf("字段条目 = %d, want 563（与 tt form show 同一个解析器，计数必须一致）", len(tab.Fields))
	}

	f := findField(t, tab, "b_xmda001")
	if f.FieldID != 18 {
		t.Errorf("b_xmda001 fieldId = %d, want 18", f.FieldID)
	}
	if f.Tag != "TableColumn" {
		t.Errorf("b_xmda001 tag = %q, want TableColumn", f.Tag)
	}
	if f.Widget != "Edit" {
		t.Errorf("b_xmda001 widget = %q, want Edit（明细列下第一个非字段元素）", f.Widget)
	}
	if f.SQLType != "SMALLINT" {
		t.Errorf("b_xmda001 sqlType = %q, want SMALLINT", f.SQLType)
	}
	if f.Comment != "cmt_xmda001" || f.TextKey != "lbl_xmda001" {
		t.Errorf("comment/textKey = %q/%q, want cmt_xmda001/lbl_xmda001", f.Comment, f.TextKey)
	}
	if f.LabelState != LabelResolved || f.Label != "料号" {
		t.Errorf("b_xmda001 label = %q/%s，想要 料号/resolved（cmt_xmda001 → xmda_t.xmda001）",
			f.Label, f.LabelState)
	}
	if f.LabelFrom != "xmda_t.xmda001" {
		t.Errorf("b_xmda001 labelFrom = %q, want xmda_t.xmda001", f.LabelFrom)
	}

	// 字典里查不到的键：必须未解析，且**不许**回落成 colName（真样本上的同一条钉子）。
	u := findField(t, tab, "b_xmda004")
	if u.Label != "" || u.LabelState != LabelUnresolved {
		t.Errorf("b_xmda004 label = %q/%s，字典没有这条就必须未解析", u.Label, u.LabelState)
	}
	if u.Label == u.ColName {
		t.Errorf("b_xmda004 的中文名列回落成了 colName —— 违反 A15/A18")
	}

	if tab.LabelResolved+tab.LabelUnresolved != len(tab.Fields) {
		t.Errorf("resolved(%d)+unresolved(%d) != 字段数(%d) —— 每个字段都必须有标注",
			tab.LabelResolved, tab.LabelUnresolved, len(tab.Fields))
	}
	if tab.LabelResolved == 0 || tab.LabelUnresolved == 0 {
		t.Errorf("真样本上应当两类都有（resolved=%d unresolved=%d）",
			tab.LabelResolved, tab.LabelUnresolved)
	}
}

// ── 8. 真样本门控：4 个 Genero 自带样本也能读（缺件 Skip） ──────────────────

func TestLoadFormGeneroFixtures(t *testing.T) {
	dir := sampleDir(t)
	for _, form := range []string{"fglfind", "fglfind_tui", "fglhelp", "fglreport"} {
		if _, err := FindPath(dir, "", form); err != nil {
			t.Skipf("样本 %s 不在（%v）——换份成果目录就跳", form, err)
		}
		tab, err := LoadForm(dir, "", form, nil)
		if err != nil {
			t.Fatalf("%s 读失败: %v", form, err)
		}
		if len(tab.Fields) == 0 {
			t.Errorf("%s 解析出 0 条字段 —— 解析器接线有问题", form)
		}
		for _, f := range tab.Fields {
			if f.LabelState != LabelUnresolved || f.Label != "" {
				t.Errorf("%s/%s 没字典时 label = %q/%s，想要空/%s",
					form, f.ColName, f.Label, f.LabelState, LabelUnresolved)
				break
			}
		}
	}
}

// ── 9. 接的是**真的**字典实现：db.DB.QueryTable（dzeb_t + dzebl_t zh_CN） ─────
//
// 上面几条用的是假件；这一条把 internal/dict/db 的真实现接进来（临时的空 SQLite，
// 只建字段档两张表，**不连任何服务器**），钉住"中文名走字典既有解析"这句话是真的：
// 键 cmt_demo001 → demo001 → dzeb_t(demo_t) JOIN dzebl_t(zh_CN) → "演示栏位"。

func TestMerge_WithRealDictDB(t *testing.T) {
	d, err := db.Open(filepath.Join(t.TempDir(), "dict.db"))
	if err != nil {
		t.Fatalf("打开临时字典库失败: %v", err)
	}
	defer d.Close()

	// 字段档：表 / 字段 / 中文兜底 / 主键 / 必填 / 备注 / 类型 / 长度 / 序号
	cols := []string{"dzeb001", "dzeb002", "dzeb003", "dzeb004", "dzeb005",
		"dzeb024", "dzeb007", "dzeb008", "dzeb021"}
	rows := [][]string{
		{"demo_t", "demo001", "兜底说明（不该被选）", "N", "N", "", "SMALLINT", "8", "1"},
		{"demo_t", "demo999", "别的栏位", "N", "N", "", "CHAR", "4", "2"},
	}
	if _, err := d.RebuildTable("dzeb_t", cols, rows); err != nil {
		t.Fatalf("建 dzeb_t 失败: %v", err)
	}
	// 多语言说明档：zh_CN 的那一行才是中文名（QueryTable 的 COALESCE 优先取它）
	if _, err := d.RebuildTable("dzebl_t",
		[]string{"dzebl001", "dzebl002", "dzebl003"},
		[][]string{{"demo001", "zh_CN", "演示栏位"}}); err != nil {
		t.Fatalf("建 dzebl_t 失败: %v", err)
	}

	tab := Merge(parseSynth(t, synthForm), "demomod", d)
	f := tab.Fields[0]
	if f.LabelState != LabelResolved || f.Label != "演示栏位" {
		t.Errorf("label = %q/%s，想要 演示栏位/resolved（真字典实现给的 zh_CN 说明）",
			f.Label, f.LabelState)
	}
	if f.LabelFrom != "demo_t.demo001" {
		t.Errorf("labelFrom = %q，想要 demo_t.demo001", f.LabelFrom)
	}
}