package fep

import (
	"strconv"
	"strings"
	"testing"
)

// `--window` 的判定必须与 `Scope.Window` **完全同口径**（契约 A12 精确化 + A18；台账 D1）。
//
// 口径只有一句，写死在这里，实现与测试都照它：
//
//	候选的窗口名 = m.WindowName(m.WindowOf(候选)) ；**精确相等**才算命中
//
// 三条不许（A18）：不接候选自身的 `text`（它不是窗口名）、不接启发式（`w_` + 剥
// `_wf/_tui/_gui`）推出来的值、不做子串/后缀/大小写之类的模糊匹配。
//
// 为什么必须钉死（D1 的实测后果）：`do.go` 原来写的是 `m.WindowName(o.ID) != window`，
// 也就是拿**候选自己的 text** 去比窗口名。`WindowName` 对非窗口对象回的就是它自己的
// `text`（`TableColumn` 的 text 是「单据别编号」），所以用 `--window` 选目标时可能
// 选错、也可能一个都选不出来。
//
// 夹具（两个窗口，标题形态与真语料一致：`标题(名字)`；见 identity_test.go 的 windowMsg）：
//
//	窗口 A(2100, name "w_axmt500",          title "料件主档作业(w_axmt500)")
//	  FormField 2400 b_xmda001 「品号」
//	窗口 B(2500, name "w_qry_q_imaf001_15", title "料件查询(q_imaf001_15)")
//	  FormField 2800 b_xmda001  「品号」                       ← 与 A 同名：窗口收窄得能分开
//	  FormField 2802 b_titlelook text = **窗口 A 的标题原样**   ← D1 的形态：text 冒充窗口名
const (
	doWinAID    = 2100
	doWinAName  = "w_axmt500"
	doWinATitle = "料件主档作业(w_axmt500)"
	doWinBID    = 2500
	doWinBName  = "w_qry_q_imaf001_15"
	doWinBTitle = "料件查询(q_imaf001_15)"

	doWinFieldA = 2400 // 窗口 A 里的「品号」
	doWinFieldB = 2800 // 窗口 B 里的「品号」（同名）
	doWinFake   = 2802 // 窗口 B 里 text 冒充窗口 A 标题的那个字段
)

// ownedField 造一个「colName + fieldId 成对、text 自定」的字段节点。
//
// 与 identity_test.go 的 formField 只差一处：colName 也叫参数给，好让"text 不是窗口名"
// 这个夹具不去污染 byName 索引（Resolver 的 R2 按 name 后缀建索引，同后缀会互相盖）。
func ownedField(id int, col string, fid int, text string) string {
	return node("FormField "+strconv.Itoa(id), [][2]string{
		{"name", "demo_t." + col}, {"colName", col},
		{"fieldId", strconv.Itoa(fid)}, {"text", text},
	})
}

// doWindowModel 装好整份合成夹具（两窗两字段 + 一个"text 冒充窗口名"的字段）。
func doWindowModel(t *testing.T) *Model {
	t.Helper()
	m := NewModel()
	m.ApplyMessage(windowMsg(doWinAID, doWinAName, doWinATitle, "axmt500_wf",
		ownedField(doWinFieldA, "b_xmda001", 18, "品号"),
	))
	m.ApplyMessage(windowMsg(doWinBID, doWinBName, doWinBTitle, "q_imaf001_15",
		ownedField(doWinFieldB, "b_xmda001", 18, "品号"),
		ownedField(doWinFake, "b_titlelook", 41, doWinATitle),
	))
	if m.Len() == 0 {
		t.Fatal("合成夹具没进模型：node/ApplyMessage 的形态不对")
	}
	// 夹具自检：窗口名与归属都算不出来，后面每一条断言都会变成"猜"。
	if got := m.WindowName(doWinAID); got != doWinATitle {
		t.Fatalf("夹具自检：窗口 A 的名字 = %q, want %q", got, doWinATitle)
	}
	if got := m.WindowName(doWinBID); got != doWinBTitle {
		t.Fatalf("夹具自检：窗口 B 的名字 = %q, want %q", got, doWinBTitle)
	}
	for id, want := range map[int]int{
		doWinFieldA: doWinAID, doWinFieldB: doWinBID, doWinFake: doWinBID,
	} {
		if got := m.WindowOf(id); got != want {
			t.Fatalf("夹具自检：#%d 的归属窗口 = %d, want %d", id, got, want)
		}
	}
	// 窗口的 `name` 属性（"w_axmt500"）**不是**窗口名 —— A18 说它只能进诊断。
	if o, ok := m.Get(doWinAID); !ok || o.Attr("name") != doWinAName {
		t.Fatalf("夹具自检：窗口 A 的 name 属性不对")
	}
	return m
}

// TestDoWindowMatchesTheSameNameAsScopeWindow：给对窗口名 ⇒ 选中该窗口里的目标；
// 并且与 `ByColName`（Scope.Window）**命中同一个节点** —— 这是"同口径"的机器证据。
func TestDoWindowMatchesTheSameNameAsScopeWindow(t *testing.T) {
	m := doWindowModel(t)
	r := NewResolver(m)

	// 不收窄：两个窗口各有一个「品号」⇒ 必须报歧义（先证明窗口收窄真的在起作用）
	if id, _, err := resolveTarget(m, r, "品号", "", "input"); err == nil {
		t.Fatalf("不给 --window 时两个窗口都有「品号」，本该报歧义，却选中了 #%d", id)
	}

	// 窗口 A 的原样标题 ⇒ 只选 A 里的那个
	idA, nameA, err := resolveTarget(m, r, "品号", doWinATitle, "input")
	if err != nil {
		t.Fatalf("--window %q 本该命中窗口 A 里的「品号」，却被拒：%v", doWinATitle, err)
	}
	if idA != doWinFieldA || nameA != "品号" {
		t.Fatalf("--window %q 选中了 #%d/%q，想要 #%d/「品号」", doWinATitle, idA, nameA, doWinFieldA)
	}

	// 同一个目标名、另一个窗口 ⇒ 选中 B 里那个
	idB, _, err := resolveTarget(m, r, "品号", doWinBTitle, "input")
	if err != nil {
		t.Fatalf("--window %q 本该命中窗口 B 里的「品号」，却被拒：%v", doWinBTitle, err)
	}
	if idB != doWinFieldB {
		t.Fatalf("--window %q 选中了 #%d，想要 #%d", doWinBTitle, idB, doWinFieldB)
	}

	// 同口径的机器证据：同一串窗口名交给 Scope.Window，命中的必须是同一个节点
	got, err := ByColName(m, Scope{Window: doWinATitle}, "b_xmda001", "input")
	if err != nil {
		t.Fatalf("ByColName(Scope.Window=%q) 本该命中：%v", doWinATitle, err)
	}
	if got.ID != idA {
		t.Fatalf("两条路的窗口口径不一致：--window 选中 #%d，Scope.Window 选中 #%d（A12 要求同一口径）",
			idA, got.ID)
	}
	if got.Scope.Window != doWinATitle {
		t.Fatalf("句柄自述的窗口名 = %q, want %q", got.Scope.Window, doWinATitle)
	}
}

// TestDoWindowDoesNotCompareTheCandidatesText 是 **D1 的回归测试**：
// 候选的 `text` 与窗口名一字不差，但它不属于那个窗口 ⇒ 不得因此命中。
func TestDoWindowDoesNotCompareTheCandidatesText(t *testing.T) {
	m := doWindowModel(t)
	r := NewResolver(m)

	// 夹具自检：这就是 D1 的形态 —— #2802 自己的 text 恰好等于窗口 A 的标题，
	// 而它实际属于窗口 B。旧实现（`m.WindowName(o.ID) != window`，比候选自身的 text）
	// 会把它当"窗口 A 的候选"而命中，所以下面那条断言在旧实现上必红。
	o, ok := m.Get(doWinFake)
	if !ok {
		t.Fatalf("夹具缺对象 #%d", doWinFake)
	}
	if got := o.Attr("text"); got != doWinATitle {
		t.Fatalf("夹具自检：#%d 的 text = %q, want %q（要的就是它与窗口 A 的名字同名）",
			doWinFake, got, doWinATitle)
	}
	if fn, ok := r.Resolve(doWinFake); !ok || fn.Text != doWinATitle || fn.Confidence != ConfHigh {
		t.Fatalf("夹具自检：#%d 的名字解析 = %+v/%v，本该是 %q 且高置信", doWinFake, fn, ok, doWinATitle)
	}

	// 目标名与窗口名给同一串：只有"候选的 text"这一路会把它算成命中（D1）。
	id, _, err := resolveTarget(m, r, doWinATitle, doWinATitle, "input")
	if err == nil {
		t.Fatalf("把候选的 text 当成了窗口名（D1 复发）：--window %q 竟然选中了 #%d（它属于窗口 B）",
			doWinATitle, id)
	}
	var ref *Refusal
	if !asRefusal(err, &ref) {
		t.Fatalf("失败必须是 *Refusal（C4.1），得到 %T: %v", err, err)
	}
	if ref.Code != "TARGET_MISSING" || ref.Exit != 5 {
		t.Fatalf("错误码/退出码 = %s/%d，想要 TARGET_MISSING/5（%s）", ref.Code, ref.Exit, ref.Message)
	}
	// 它确实躺在**别的窗口**里：错误里得点出那个窗口，人才知道是选错了窗口而不是选错了名字
	if !strings.Contains(ref.Message, doWinBTitle) {
		t.Fatalf("错误里没点出 #%d 真正所在的窗口 %q：%s", doWinFake, doWinBTitle, ref.Message)
	}
}

// TestDoWindowMissListsTheRealWindowNames：窗口名写错（含"启发式形状"与"子串"）⇒ 报错里
// 必须列出**真实存在**的窗口名，能直接复制回 `--window`。
func TestDoWindowMissListsTheRealWindowNames(t *testing.T) {
	m := doWindowModel(t)
	r := NewResolver(m)

	// 都是"看着像"但**不是**窗口名的写法：
	//   w_axmt500        —— 启发式形状（A18：只进诊断，不得参与判定）
	//   料件主档作业      —— 标题的**子串**
	//   AXMT500          —— 大小写
	for _, wrong := range []string{doWinAName, "料件主档作业", "AXMT500"} {
		_, _, err := resolveTarget(m, r, "品号", wrong, "input")
		if err == nil {
			t.Fatalf("--window %q 不是窗口名，却命中了：不许模糊匹配（A18）", wrong)
		}
		var ref *Refusal
		if !asRefusal(err, &ref) {
			t.Fatalf("--window %q：失败必须是 *Refusal，得到 %T: %v", wrong, err, err)
		}
		if ref.Code != "TARGET_MISSING" || ref.Exit != 5 {
			t.Fatalf("--window %q：错误码/退出码 = %s/%d，想要 TARGET_MISSING/5", wrong, ref.Code, ref.Exit)
		}
		// 真实窗口名必须**原样**出现在错误里（可复制），两个窗口都要有
		for _, want := range []string{doWinATitle, doWinBTitle} {
			if !strings.Contains(ref.Message, want) {
				t.Errorf("--window %q 的错误里没有真实窗口名 %q，人没法复制：%s", wrong, want, ref.Message)
			}
		}
	}

	// 窗口名对、但这个窗口里没有该目标 ⇒ 同样的报错形态（列真实窗口名，不猜）
	_, _, err := resolveTarget(m, r, "根本没有这个栏位", doWinATitle, "input")
	if err == nil {
		t.Fatal("目标名不存在时必须被拒")
	}
	var ref *Refusal
	if !asRefusal(err, &ref) || !strings.Contains(ref.Message, doWinATitle) {
		t.Fatalf("目标名不存在时，错误里也要有真实窗口名：%v", err)
	}
}
