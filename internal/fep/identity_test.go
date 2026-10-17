package fep

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// 字段寻址（契约 C3 + A7）的守护测试。
//
// 全部用**合成夹具**：两个窗口 + 一张表单里的若干字段，数字与名字都是为了把每一条
// 收窄规则单独钉住而挑的。节点形态取自真语料的原样（见 testdata/fep-recordings/*.txt）：
//
//	{an 0 Window 687 {{name "w_qry_q_ooba002_1"} {text "单据别编号数据查询(q_ooba002_1)"}} …}
//	{TableColumn 3631 {{name "formonly.ooba002_1"} {colName "ooba002_1"} {fieldId "1"} {text "单据别编号"}} …}
//
// 夹具（窗口 A = 主档作业，窗口 B = 另一个窗口，两边**故意有同名 colName**）：
//
//	窗口 A(100, name "w_axmt500",  title "料件主档作业(w_axmt500)")
//	  FormField 400 b_xmda001#18  ← 第一帧
//	  FormField 402 b_xmda001#18  ← 同一字段的第二帧（C3.5：取 Seq 最后声明的那一份）
//	  FormField 404 b_index#25
//	  FormField 406 b_index#34    ← 两种 fieldId = 真歧义（C3.4/A7）
//	  FormField 408 b_unique#7
//	  FormField 410 b_noid（只有 colName，没有 fieldId）
//	  FormField 412 b_col#3       ← 同一 colName 也出现在 TableColumn 414 上
//	  TableColumn 414 b_col#3
//	  Item 416 b_two#11 / Edit 418 b_two#11   ← 动作类型（click/input）能收窄
//	  TableColumn 420 b_onlycol#15            ← **唯一候选是明细列**（A11）
//	  TableColumn 422 b_twocols#16 / 424 b_twocols#17  ← 多候选但**全不可接受**（A11）
//	  FormField 426 b_mixed#**0** / FormField 428 b_mixed#12  ← 编号 0 是**合法值**（A10）
//	  FormField 430 b_late#20 / TableColumn 432 b_late#20     ← 明细列**后声明**（A9 优先可接受）
//	窗口 B(500, name "w_qry_q_imaf001_15", title "料件查询(w_qry_q_imaf001_15)")
//	  FormField 800 b_xmda001#18  ← 与窗口 A 同名
//	  FormField 802 b_other#5

const (
	testWinAID    = 100
	testWinAName  = "w_axmt500"
	testWinATitle = "料件主档作业(w_axmt500)"
	testWinBID    = 500
	testWinBName  = "w_qry_q_imaf001_15"
	testWinBTitle = "料件查询(w_qry_q_imaf001_15)"
)

// formField 造一个带字段身份的节点（colName + fieldId 成对，与真语料一致）。
func formField(id int, col string, fid int, text string) string {
	return node(fmt.Sprintf("FormField %d", id), [][2]string{
		{"name", "demo_t." + col}, {"colName", col},
		{"fieldId", strconv.Itoa(fid)}, {"text", text},
	})
}

// windowMsg 造一条"一个窗口一棵树"的消息。
//
// **一条消息一个窗口**是有意的：声明那条消息所属的窗口就是 Object.DeclWindow，
// 而 WindowOf 的第一手证据正是它（aui.go 的注释：实测"两个开窗共享的模板列"只有它算得对）。
func windowMsg(winID int, name, title, form string, fields ...string) string {
	return node(fmt.Sprintf("an 0 Window %d", winID),
		[][2]string{{"name", name}, {"text", title}},
		node(fmt.Sprintf("Form %d", winID+100), [][2]string{{"name", form}, {"text", title}},
			node(fmt.Sprintf("HBox %d", winID+200), nil, fields...)))
}

// identityModel 装好整份合成夹具。
func identityModel(t *testing.T) *Model {
	t.Helper()
	m := NewModel()
	m.ApplyMessage(windowMsg(testWinAID, testWinAName, testWinATitle, "demo_t",
		formField(400, "b_xmda001", 18, "品号"),
		formField(402, "b_xmda001", 18, "品号"),
		formField(404, "b_index", 25, "索引"),
		formField(406, "b_index", 34, "索引（另一张表单的编号）"),
		formField(408, "b_unique", 7, "唯一字段"),
		node("FormField 410", [][2]string{{"name", "demo_t.b_noid"}, {"colName", "b_noid"}, {"text", "没编号的字段"}}),
		formField(412, "b_col", 3, "网格列（表单字段形态）"),
		node("TableColumn 414", [][2]string{{"name", "demo_t.b_col"}, {"colName", "b_col"}, {"fieldId", "3"}, {"text", "网格列"}}),
		node("Item 416", [][2]string{{"colName", "b_two"}, {"fieldId", "11"}, {"text", "只可点"}}),
		node("Edit 418", [][2]string{{"colName", "b_two"}, {"fieldId", "11"}, {"text", "只可写"}}),
		// A11：唯一候选是明细列（TableColumn 既不可写也不可点）
		node("TableColumn 420", [][2]string{{"name", "demo_t.b_onlycol"}, {"colName", "b_onlycol"}, {"fieldId", "15"}, {"text", "只在明细列上"}}),
		// A11：多候选但全不可接受（两种编号的明细列，只为了"列候选 + 节点类型"）
		node("TableColumn 422", [][2]string{{"name", "demo_t.b_twocols"}, {"colName", "b_twocols"}, {"fieldId", "16"}, {"text", "两个明细列"}}),
		node("TableColumn 424", [][2]string{{"name", "demo_t.b_twocols"}, {"colName", "b_twocols"}, {"fieldId", "17"}, {"text", "两个明细列（另一列）"}}),
		// A10：fieldId 可以合法为 0 —— 以前"0 = 没给"会把它吞掉
		formField(426, "b_mixed", 0, "编号 0 的字段"),
		formField(428, "b_mixed", 12, "编号 12 的字段"),
		// A9：同一个 #20 上，明细列**后声明** —— 谁在"可接受组"里才算数，不是谁最后声明
		formField(430, "b_late", 20, "先声明的可写字段"),
		node("TableColumn 432", [][2]string{{"name", "demo_t.b_late"}, {"colName", "b_late"}, {"fieldId", "20"}, {"text", "后声明的明细列"}}),
	))
	m.ApplyMessage(windowMsg(testWinBID, testWinBName, testWinBTitle, "qry_t",
		formField(800, "b_xmda001", 18, "品号"),
		formField(802, "b_other", 5, "别的字段"),
	))
	if m.Len() == 0 {
		t.Fatal("合成夹具没进模型：node/ApplyMessage 的形态不对")
	}
	// 夹具自检：归属与窗口名算不出来，后面每一条断言都会变成"猜"。
	for _, id := range []int{400, 402, 414, 800, 802} {
		if _, ok := m.Get(id); !ok {
			t.Fatalf("夹具缺对象 #%d", id)
		}
	}
	if got := m.WindowOf(402); got != testWinAID {
		t.Fatalf("夹具自检：#402 的归属窗口 = %d, want %d", got, testWinAID)
	}
	if got := m.WindowOf(800); got != testWinBID {
		t.Fatalf("夹具自检：#800 的归属窗口 = %d, want %d", got, testWinBID)
	}
	if got := m.WindowName(testWinAID); got != testWinATitle {
		t.Fatalf("夹具自检：窗口 A 的名字 = %q, want %q", got, testWinATitle)
	}
	return m
}

// mustResolve 定位并断言成功。
func mustResolve(t *testing.T, m *Model, sc Scope, col ColName, kind string) Resolved {
	t.Helper()
	got, err := ByColName(m, sc, col, kind)
	if err != nil {
		t.Fatalf("定位 %q（窗口 %q, kind=%s）本该命中，却被拒：%v", col, sc.Window, kind, err)
	}
	return got
}

// mustRefuse 定位并断言被拒，返回那条拒绝（C4.1：失败一律 *Refusal）。
func mustRefuse(t *testing.T, m *Model, sc Scope, col ColName, kind, wantCode string) *Refusal {
	t.Helper()
	got, err := ByColName(m, sc, col, kind)
	if err == nil {
		t.Fatalf("定位 %q（窗口 %q, kind=%s）本该被拒（%s），却命中了 #%d",
			col, sc.Window, kind, wantCode, got.ID)
	}
	r, ok := err.(*Refusal)
	if !ok {
		t.Fatalf("失败必须是 *Refusal（C4.1），得到 %T: %v", err, err)
	}
	if r.Code != wantCode {
		t.Fatalf("错误码 = %q, want %q（%s）", r.Code, wantCode, r.Message)
	}
	if r.Message == "" {
		t.Errorf("拒绝没有 Message：%+v", r)
	}
	if r.Hint == "" {
		t.Errorf("拒绝没有 Hint —— C4.3 要求可操作：%+v", r)
	}
	return r
}

// fingerprint 折叠模型的**全部可观察状态**，用来断言"拒绝路径没有任何副作用"。
//
// 只看"返回了错误"是不够的：一个会顺手在模型里留下痕迹的实现（改属性、把候选删掉、
// 把 Scope 写回对象）同样会返回错误，而那种实现会让**下一步**的行为变样。
func fingerprint(m *Model) string {
	var b strings.Builder
	fmt.Fprintf(&b, "len=%d conflicts=%d | ", m.Len(), m.Conflicts())
	for _, o := range m.Objects() {
		fmt.Fprintf(&b, "#%d{scope=%d win=%d decl=%d seq=%d type=%s parent=%d attrs=%v} ",
			o.ID, o.Scope, o.Win, o.DeclWindow, o.Seq, o.Type, o.Parent, o.Attrs)
	}
	return b.String()
}

// handleIsStale 判断旧句柄是不是还指着当初那个对象（H2 的判据）：
// 同一个裸 id 的 Seq 变了（或这个 id 已经不在了）就说明句柄失效。
func handleIsStale(m *Model, h Resolved) bool {
	o, ok := m.Get(h.ID)
	return !ok || uint32(o.Seq) != h.Seq
}

// wantID 造一个 `WantFieldID` 指针（A10：nil = 没给；`wantID(0)` 是**明确的编号 0**）。
func wantID(id FieldID) *FieldID { return &id }

// scoped 在已有作用域上补一个"我要的编号"，让测试读起来短一点。
func scoped(sc Scope, id *FieldID) Scope {
	sc.WantFieldID = id
	return sc
}

// ── C3.4 / A7：唯一命中 ───────────────────────────────────────────────────

func TestByColName_Unique(t *testing.T) {
	m := identityModel(t)

	got := mustResolve(t, m, Scope{Window: testWinATitle}, "b_unique", "input")
	if got.ID != 408 {
		t.Errorf("ID = %d, want 408", got.ID)
	}
	if got.FieldID != 7 {
		t.Errorf("FieldID = %d, want 7", got.FieldID)
	}
	if got.ColName != "b_unique" {
		t.Errorf("ColName = %q, want b_unique", got.ColName)
	}
	if got.Seq == 0 {
		t.Error("Seq = 0：句柄必须带上它在模型里的声明序号（C3.2）")
	}
	// 句柄自述：调用方没给窗口时，把**实际命中的**窗口补进 Scope（诊断字段，A7/A12）。
	if got.Scope.Window != testWinATitle {
		t.Errorf("Scope.Window = %q, want %q（命中的那个窗口）", got.Scope.Window, testWinATitle)
	}
	// A10：`WantFieldID` 是**调用方的请求**，句柄不许伪造它（nil = 调用方没给）。
	if got.Scope.WantFieldID != nil {
		t.Errorf("Scope.WantFieldID = %v，不许由句柄编出来（A10）", *got.Scope.WantFieldID)
	}
	// 节点自身的编号仍然照实带出（A10：Resolved.FieldID 是它自己的编号）。
	if got.FieldID != 7 {
		t.Errorf("Resolved.FieldID = %d, want 7", got.FieldID)
	}
	// 引用它的是 HBox 320（窗 A 的 Form 200 → HBox 300）：句柄必须是能直接发事件的会话 id。
	if o, ok := m.Get(got.ID); !ok || o.Attr("colName") != "b_unique" {
		t.Errorf("#%d 在模型里不是那个字段：%+v", got.ID, o)
	}

	// 不给窗口也照样唯一（全模型只有一份）。
	second := mustResolve(t, m, Scope{}, "b_other", "input")
	if second.ID != 802 {
		t.Errorf("不给窗口时 b_other 的 ID = %d, want 802", second.ID)
	}
}

// ── C3.5：同一字段的多帧副本 → 取 Seq 最后声明的那一份 ─────────────────────

func TestByColName_TakesLastDeclaredFrameOfSameField(t *testing.T) {
	m := identityModel(t)
	sc := Scope{Window: testWinATitle}

	got := mustResolve(t, m, sc, "b_xmda001", "input")
	if got.ID != 402 {
		t.Fatalf("ID = %d, want 402（同一 fieldId 的两帧要取**最后声明**的那一份，C3.5）", got.ID)
	}
	first, ok := m.Get(400)
	if !ok {
		t.Fatal("夹具缺 #400")
	}
	if got.Seq <= uint32(first.Seq) {
		t.Errorf("Seq = %d 不大于第一帧的 %d：说明取的不是最后声明的那一份", got.Seq, first.Seq)
	}
	if got.FieldID != 18 {
		t.Errorf("FieldID = %d, want 18", got.FieldID)
	}
	// 窗口 B 的那一份（#800）不该被牵进来 —— 收窄是窗口先做的。
	if got.ID == 800 {
		t.Error("命中了别的窗口里的副本：窗口收窄没生效")
	}
}

// ── A7：窗口收窄 ──────────────────────────────────────────────────────────

func TestByColName_WindowNarrows(t *testing.T) {
	m := identityModel(t)

	a := mustResolve(t, m, Scope{Window: testWinATitle}, "b_xmda001", "input")
	if a.ID != 402 {
		t.Errorf("窗口 A 里 = #%d, want 402", a.ID)
	}
	b := mustResolve(t, m, Scope{Window: testWinBTitle}, "b_xmda001", "input")
	if b.ID != 800 {
		t.Errorf("窗口 B 里 = #%d, want 800", b.ID)
	}

	// 不给窗口 ⇒ 不收窄 ⇒ 按 A7"取最后声明的那一份"。窗口 B 是后声明的，所以是 #800。
	// （这条**故意**钉住"没有窗口时不给任何偏好"，免得有人偷偷加"优先主窗"这种猜测。）
	none := mustResolve(t, m, Scope{}, "b_xmda001", "input")
	if none.ID != 800 {
		t.Errorf("不给窗口时 = #%d, want 800（全模型里最后声明的那一份）", none.ID)
	}

	// **A12**：`Scope.Window` 的语义就是 `m.WindowName` 的口径，**不接**窗口的 `name`
	// 属性（"w_axmt500" 这种）。给 name ⇒ 一个都不匹配 ⇒ 照实拒（而不是"我帮你猜一个窗口"）。
	mustRefuse(t, m, Scope{Window: testWinAName}, "b_xmda001", "input", CodeFieldNotInScope)
}

// ── C3.4 / A7：真歧义（多种 fieldId）→ 拒，且列候选、无副作用 ─────────────

func TestByColName_AmbiguousOnMultipleFieldIDs(t *testing.T) {
	m := identityModel(t)
	before := fingerprint(m)

	r := mustRefuse(t, m, Scope{Window: testWinATitle}, "b_index", "input", CodeFieldAmbiguous)

	if r.Exit != 5 {
		t.Errorf("Exit = %d, want 5（C4.2）", r.Exit)
	}
	// 候选必须列出来：colName#fieldId + 窗口名（A7/C4.3），而且要两个都在。
	for _, want := range []string{"b_index#25", "b_index#34", testWinATitle} {
		if !strings.Contains(r.Message, want) {
			t.Errorf("Message 里没列出候选 %q：%s", want, r.Message)
		}
	}
	// Hint 要**可操作**：能照着敲（把编号显式给上 + 离线字段表命令）。
	for _, needle := range []string{"WantFieldID", "--env", "--module", "--form"} {
		if !strings.Contains(r.Hint, needle) {
			t.Errorf("Hint 里没有 %q，不可照着敲：%s", needle, r.Hint)
		}
	}

	// **没有任何副作用**：被拒之后模型必须一字不差（不是"只断言返回了错误"）。
	if after := fingerprint(m); after != before {
		t.Errorf("歧义路径改动了模型：\n前 %s\n后 %s", before, after)
	}
	// 也不许"顺手挑一个"：真歧义时 #404/#406 谁都不该被返回 —— 上面 mustRefuse 已经
	// 断言了 err != nil，这里再钉一次句柄是空的。
	got, err := ByColName(m, Scope{Window: testWinATitle}, "b_index", "input")
	if err == nil || got.ID != 0 {
		t.Errorf("真歧义必须回空句柄 + 拒绝，得到 #%d / %v", got.ID, err)
	}
}

// ── C4.2 / A7：0 个候选 → FIELD_NOT_IN_SCOPE ──────────────────────────────

func TestByColName_NotInScope(t *testing.T) {
	m := identityModel(t)

	cases := []struct {
		name string
		sc   Scope
		col  ColName
		kind string
	}{
		{"这张表单里没这个 colName", Scope{Window: testWinATitle}, "b_nope", "input"},
		{"colName 命中但 fieldId 不符（C3.4：判为另一张表单）",
			Scope{Window: testWinATitle, WantFieldID: wantID(34)}, "b_xmda001", "input"},
		{"节点只有 colName、没有 fieldId：双键缺一半 ⇒ 不可寻址",
			Scope{Window: testWinATitle}, "b_noid", "input"},
		{"没给 colName", Scope{Window: testWinATitle}, "", "input"},
	}
	for _, c := range cases {
		r := mustRefuse(t, m, c.sc, c.col, c.kind, CodeFieldNotInScope)
		if r.Exit != 5 {
			t.Errorf("%s: Exit = %d, want 5（C4.2）", c.name, r.Exit)
		}
		if !strings.Contains(r.Hint, "tt form show") {
			t.Errorf("%s: Hint 没给出 `tt form show` 这条可操作命令：%s", c.name, r.Hint)
		}
	}
}

// 找错了窗口时，Hint 要把"这个 colName 在哪些窗口里"列出来（C4.3：可操作）。
func TestByColName_NotInScopePointsAtOtherWindows(t *testing.T) {
	m := identityModel(t)
	r := mustRefuse(t, m, Scope{Window: testWinATitle}, "b_other", "input", CodeFieldNotInScope)
	if !strings.Contains(r.Hint, testWinBTitle) {
		t.Errorf("Hint 该指出 b_other 在窗口 %q 里：%s", testWinBTitle, r.Hint)
	}
}

// ── A7/A10：调用方给了编号（`WantFieldID`）→ 唯一命中就直接用它（消歧） ─────

func TestByColName_WantFieldIDDisambiguates(t *testing.T) {
	m := identityModel(t)

	got25 := mustResolve(t, m, Scope{Window: testWinATitle, WantFieldID: wantID(25)}, "b_index", "input")
	if got25.ID != 404 || got25.FieldID != 25 {
		t.Errorf("WantFieldID=25 → #%d/%d, want #404/25", got25.ID, got25.FieldID)
	}
	got34 := mustResolve(t, m, Scope{Window: testWinATitle, WantFieldID: wantID(34)}, "b_index", "input")
	if got34.ID != 406 || got34.FieldID != 34 {
		t.Errorf("WantFieldID=34 → #%d/%d, want #406/34", got34.ID, got34.FieldID)
	}
	// 调用方给的编号原样保留（不把它的写法换掉），窗口照实带出。
	if got34.Scope.WantFieldID == nil || *got34.Scope.WantFieldID != 34 || got34.Scope.Window != testWinATitle {
		t.Errorf("Scope 回带 = %+v, want 原样保留 34 / %q", got34.Scope, testWinATitle)
	}
	// 给了编号但一个都不符 → 拒（C3.4），不是"退回按 colName 挑一个"。
	r := mustRefuse(t, m, Scope{Window: testWinATitle, WantFieldID: wantID(99)}, "b_index", "input", CodeFieldNotInScope)
	if r.Exit != 5 {
		t.Errorf("Exit = %d, want 5", r.Exit)
	}
}

// **A10 的立命之处**：`fieldId` 可以合法为 0，"0 = 没给"会把它吞掉。
// 用指针表达"给了"之后，编号 0 的那一份**能被指名**。
func TestByColName_WantFieldIDZeroIsExpressible(t *testing.T) {
	m := identityModel(t)
	sc := Scope{Window: testWinATitle}

	// b_mixed 有两份：#426 编号 0、#428 编号 12 ⇒ 不给编号时是真歧义。
	mustRefuse(t, m, sc, "b_mixed", "input", CodeFieldAmbiguous)

	// 指名编号 0（老写法 `FieldID: 0` 在这里等于"没给"，只能拿到歧义）。
	zero := mustResolve(t, m, scoped(sc, wantID(0)), "b_mixed", "input")
	if zero.ID != 426 || zero.FieldID != 0 {
		t.Errorf("WantFieldID=*0 → #%d/%d, want #426/0", zero.ID, zero.FieldID)
	}
	twelve := mustResolve(t, m, scoped(sc, wantID(12)), "b_mixed", "input")
	if twelve.ID != 428 || twelve.FieldID != 12 {
		t.Errorf("WantFieldID=12 → #%d/%d, want #428/12", twelve.ID, twelve.FieldID)
	}
}

// ── A9/A11：`kind` 不淘汰候选，只在多候选时**优先可接受的** ────────────────

// 同一个 colName#fieldId 上挂了两个动作面不同的控件：可点的（Item）与可写的（Edit）。
// A9 之后语义是"优先可接受"，所以 input 落到 Edit、click 落到 Item。
func TestByColName_KindSelectsByAction(t *testing.T) {
	m := identityModel(t)
	sc := Scope{Window: testWinATitle}

	in := mustResolve(t, m, sc, "b_two", "input")
	if in.ID != 418 {
		t.Errorf("input 时 = #%d, want 418（Edit 才是可写的）", in.ID)
	}
	clk := mustResolve(t, m, sc, "b_two", "click")
	if clk.ID != 416 {
		t.Errorf("click 时 = #%d, want 416（Item 才是可点的）", clk.ID)
	}
}

func TestByColName_KindPrefersActionable(t *testing.T) {
	m := identityModel(t)
	sc := Scope{Window: testWinATitle}

	// b_late#20 上：可写的 FormField **先**声明，#表后 声明的是一张明细列（TableColumn）。
	// 若只看"最后声明"就会挑到明细列（发不出去）；A9 要求先在"可接受组"里判定 ⇒ 取 #430。
	got := mustResolve(t, m, sc, "b_late", "input")
	if got.ID != 430 {
		t.Errorf("ID = #%d, want 430（明细列不在可接受组里，优先可接受的 #430）", got.ID)
	}
	if got.FieldID != 20 {
		t.Errorf("FieldID = %d, want 20", got.FieldID)
	}
	// 同一个 colName 在 FormField 与 TableColumn 上各一份 ⇒ 只在可接受组里判定，命中 FormField。
	col := mustResolve(t, m, sc, "b_col", "input")
	if col.ID != 412 {
		t.Errorf("b_col/input = #%d, want 412（#414 是明细列，不在可接受组）", col.ID)
	}
}

// 唯一候选是明细列 + kind="input" ⇒ **身份找到了，但发不了**（A11：退 4，不是 5）。
func TestByColName_NotActionable(t *testing.T) {
	m := identityModel(t)
	before := fingerprint(m)

	r := mustRefuse(t, m, Scope{Window: testWinATitle}, "b_onlycol", "input", CodeFieldNotActionable)
	if r.Exit != 4 {
		t.Errorf("Exit = %d, want 4（A11：状态/能力不符，与「没这个字段」分开）", r.Exit)
	}
	for _, needle := range []string{"b_onlycol#15", "TableColumn"} {
		if !strings.Contains(r.Message, needle) {
			t.Errorf("Message 里该说清哪个身份、什么类型（%q）：%s", needle, r.Message)
		}
	}
	// Hint 要给出"那么该发到哪"：明细列底下的子 Edit。
	for _, needle := range []string{"Edit", "tt fep ui"} {
		if !strings.Contains(r.Hint, needle) {
			t.Errorf("Hint 里没有 %q（该说明发到子控件、并给命令）：%s", needle, r.Hint)
		}
	}
	if after := fingerprint(m); after != before {
		t.Errorf("不可发路径改动了模型：\n前 %s\n后 %s", before, after)
	}

	// 明细列在**两张**白名单里都没有：换 click 也是同一个码（不是"没这个字段"）。
	r2 := mustRefuse(t, m, Scope{Window: testWinATitle}, "b_onlycol", "click", CodeFieldNotActionable)
	if r2.Exit != 4 {
		t.Errorf("click 时 Exit = %d, want 4", r2.Exit)
	}
}

// 多候选、但**全不可接受** ⇒ 同一个码（A11），候选里要列 `colName#fieldId` + **节点类型**。
func TestByColName_AllCandidatesNotActionable(t *testing.T) {
	m := identityModel(t)
	r := mustRefuse(t, m, Scope{Window: testWinATitle}, "b_twocols", "input", CodeFieldNotActionable)
	if r.Exit != 4 {
		t.Errorf("Exit = %d, want 4（A11）", r.Exit)
	}
	for _, needle := range []string{"b_twocols#16", "b_twocols#17", "TableColumn"} {
		if !strings.Contains(r.Message, needle) {
			t.Errorf("Message 里该列出候选 %q：%s", needle, r.Message)
		}
	}
	if !strings.Contains(r.Hint, "tt fep ui") {
		t.Errorf("Hint 该给出看子节点的命令：%s", r.Hint)
	}
}

// ── A7：Scope.Form / Variant 只做校验与诊断，**不作运行期闸门** ───────────

func TestByColName_FormDoesNotGate(t *testing.T) {
	m := identityModel(t)
	// 刻意给一个**对不上**的表单名与变体：A7 说这不构成拒绝理由
	// （离线表可能来自另一个变体/版本），拒绝只能由候选数决定。
	sc := Scope{Window: testWinATitle, Form: "another_form", Variant: "tui"}
	got := mustResolve(t, m, sc, "b_unique", "input")
	if got.ID != 408 {
		t.Errorf("ID = %d, want 408：Form/Variant 不该当闸门", got.ID)
	}
	if got.Scope.Form != "another_form" || got.Scope.Variant != "tui" {
		t.Errorf("Scope 回带把调用方的表单名/变体弄丢了：%+v", got.Scope)
	}
}

// ── H2：`rn` 之后旧句柄失效 ───────────────────────────────────────────────

func TestResolvedHandleDiesAfterRemove(t *testing.T) {
	m := identityModel(t)
	sc := Scope{Window: testWinATitle}

	h := mustResolve(t, m, sc, "b_unique", "input")
	if h.ID != 408 {
		t.Fatalf("ID = %d, want 408", h.ID)
	}
	if handleIsStale(m, h) {
		t.Fatal("刚拿到的句柄就被判成失效了：判据不对")
	}

	// 1) 运行期删掉它（`rn`，AUI 同步的三种操作之一：an 加 / un 改 / rn 删）。
	m.ApplyMessage("{rn 408}")
	if _, ok := m.Get(408); ok {
		t.Fatal("`rn 408` 之后 #408 还活着")
	}
	if !handleIsStale(m, h) {
		t.Error("rn 之后旧句柄仍被判成有效（H2）")
	}
	// 真发事件前的预检也要挡住它 —— 这是"旧句柄拿去用会怎样"的可执行答案。
	if err := CheckClick(m, h.ID); err == nil {
		t.Error("rn 之后拿旧句柄过预检居然通过了（会发到空气上）")
	} else if r, ok := err.(*Refusal); !ok || r.Code != "TARGET_MISSING" {
		t.Errorf("预检该报 TARGET_MISSING，得到 %T: %v", err, err)
	}

	// 2) 更阴的一档：运行期**复用**这个裸 id（实测形态是"先 rn 再 an"，见 aui.go 的注释）。
	m.ApplyMessage(node("an 300 FormField 408", [][2]string{
		{"name", "demo_t.b_other"}, {"colName", "b_other"}, {"fieldId", "5"}, {"text", "换人了"},
	}))
	now, ok := m.Get(408)
	if !ok {
		t.Fatal("复用后 #408 应该在模型里")
	}
	if now.Attr("colName") != "b_other" {
		t.Fatalf("#408 复用后 = %q, want b_other", now.Attr("colName"))
	}
	if !handleIsStale(m, h) {
		t.Errorf("同一个裸 id 换了对象（seq %d → %d），旧句柄仍被判成有效：%v", h.Seq, now.Seq, h)
	}
	if uint32(now.Seq) == h.Seq {
		t.Error("复用前后的 Seq 相同：分不出\"现在是谁\"")
	}
}

// ── C3.3：方法形式与自由函数是**同一处**收窄顺序 ───────────────────────────

// 这条是**实现之后补写**的（不是红→绿的那批）：它守的是一个 3 行薄包装，
// 目的事后钉住"`fep` 与 `view` 共用同一个 Resolver 时走的是同一份顺序"，
// 免得日后有人在方法里再写第二份收窄逻辑（两份迟早会分叉）。
// 它真的会咬人：见"红→绿"记录 §3 的第二条反证（把包装改成回零值 → 这条红）。
func TestResolverByColNameSharesOneNarrowingOrder(t *testing.T) {
	m := identityModel(t)
	sc := Scope{Window: testWinATitle}
	r := NewResolver(m)

	direct := mustResolve(t, m, sc, "b_unique", "input")
	viaMethod, err := r.ByColName(m, sc, "b_unique", "input")
	if err != nil {
		t.Fatalf("方法形式（C3.3）被拒：%v", err)
	}
	if viaMethod != direct {
		t.Errorf("方法形式 = %+v，自由函数 = %+v：两条路必须给同一个结果", viaMethod, direct)
	}

	// 调用方没给模型时用 Resolver 自带的那份（NewResolver 已经在手上）。
	viaResolverModel, err := r.ByColName(nil, sc, "b_unique", "input")
	if err != nil {
		t.Fatalf("Resolver 自带模型时被拒：%v", err)
	}
	if viaResolverModel != direct {
		t.Errorf("用 Resolver 自带模型 = %+v，想要 %+v", viaResolverModel, direct)
	}

	// 拒绝路径也照旧（不是只有命中那条路被包了一层）。
	if _, err := r.ByColName(nil, sc, "b_index", "input"); err == nil {
		t.Error("方法形式该把真歧义照原样拒掉")
	} else if ref, ok := err.(*Refusal); !ok || ref.Code != CodeFieldAmbiguous {
		t.Errorf("方法形式的拒绝形状不对：%T %v", err, err)
	}
}
