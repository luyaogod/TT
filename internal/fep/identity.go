package fep

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ── 字段寻址（契约 C3 + A7） ────────────────────────────────────────────────
//
// 这一层回答的是驱动每走一步都要问的那个问题：**"这一屏上我说的到底是哪个控件？"**
// 判据只能是"手上拿得到的硬事实"，不能是猜：
//
//  1. `colName` + `fieldId` 是编译期分配、**表单内稳定**的字段身份。实测它们总是成对
//     出现（`testdata/fep-recordings/*.txt` 三份夹具里 47 处 `colName` 与 47 处
//     `fieldId` 一一配对，形态：`{TableColumn 3631 {{name "formonly.ooba002_1"}
//     {colName "ooba002_1"} {fieldId "1"} {text "单据别编号"}} …}`）。
//  2. 会话 id（`Resolved.ID`）是**单步句柄**：运行期会复用裸 id（同一个 4937 既做过
//     下拉 Item 又做过 Dialog 的 Action，见 aui.go 里 `Object.Seq` 的实测注释），
//     所以它用完即弃（H1）；`rn` 之后更是全量失效（H2）。
//
// **A7（架构师 2026-10-17 定）**：运行期节点**不自带表单文件名**，所以
// `Scope.Form` / `Scope.Variant` **不作运行期闸门**，只用于离线查表与诊断 ——
// 照 C3.7 原文"取不到就拒"会把每一步都拒掉。**拒与不拒只由候选数决定**，
// 收窄顺序写死在 `ByColName` 的注释里。

// ── 冻结的四码与退出码（契约 C4.2 + A11） ──────────────────────────────────
//
// 数字与 `internal/cli/fep` 的码表同档（契约 B10：5 = 目标不存在，4 = 状态不符，
// 1 = 运行失败），但**不要**与 `output.Code*` 混用：fep 线有自己的码表
// （C4.1 明确"不走 target.ExitFor"）。
const (
	// CodeFieldNotInScope：作用域里没有这个 colName，或 colName 命中但 fieldId 不符
	// （后者按 C3.4 判为"另一张表单"）。
	CodeFieldNotInScope = "FIELD_NOT_IN_SCOPE"
	// CodeFieldAmbiguous：同一 colName 命中**多种** fieldId —— 真歧义，绝不自己挑一个。
	CodeFieldAmbiguous = "FIELD_AMBIGUOUS"
	// CodeFormUnknown：表单取不到（网络/权限），或不是合法 `.42f`，或根本没有模型可用。
	// 取件/解析那一族的出口（C1.7、C2），本文件只钉它的字面值与退出码，让两个切片同一个口径。
	CodeFormUnknown = "FORM_UNKNOWN"
	// CodeFieldNotActionable（A11）：身份**找到了**，但这个节点不接受该动作。
	// 典型：明细列 `TableColumn` 不接受直接输入/点击，应发到它的子 `Edit`。
	// 与 CodeFieldNotInScope 分开：一个是"没这个字段"，一个是"有这个字段、现在发不了"——
	// 自动化排障要看得出区别（前者多半是名字/窗口找错了，后者是该换节点发）。
	CodeFieldNotActionable = "FIELD_NOT_ACTIONABLE"

	// 退出码（契约 C4.2 + A11 冻结：5 / 5 / 1 / 4）。
	ExitFieldNotInScope    = 5
	ExitFieldAmbiguous     = 5
	ExitFormUnknown        = 1
	ExitFieldNotActionable = 4
)

// refusalExitCodes 是"码 → 退出码"的**唯一一张表**：拒绝一律从它取退出码，
// 免得同一个码在两个调用点给出两个数字（牙 d 的 TestRefusalCodesAndExits 钉它）。
var refusalExitCodes = map[string]int{
	CodeFieldNotInScope:    ExitFieldNotInScope,
	CodeFieldAmbiguous:     ExitFieldAmbiguous,
	CodeFormUnknown:        ExitFormUnknown,
	CodeFieldNotActionable: ExitFieldNotActionable,
}

// refusal 按冻结的码表构造拒绝（C4.1：失败一律 `*Refusal`，不走 `internal/target`）。
func refusal(code, message, hint string) *Refusal {
	exit, ok := refusalExitCodes[code]
	if !ok {
		// 码表就在本文件里，这里走不到；留判据是为了以后加码时不会静默给 0（0 = 成功）。
		exit = ExitFormUnknown
	}
	return &Refusal{Code: code, Exit: exit, Message: message, Hint: hint}
}

// Scope 是定位的作用域（契约 C3.1；按 A7/A10/A12 增 `WantFieldID` 并钉死 `Window` 语义）。
type Scope struct {
	// Window = **AUI 窗口名**，即既有 `m.WindowName(id)` 的口径（A12），**不接**窗口的
	// `name` 属性（"w_axmt500" 那种写法）也不接"给人看的标题"。非空时用来**收窄**：
	// 只认属于这个窗口实例的候选（归属用 `WindowOf` 算，见 windowNamed）。
	Window string
	// Form / Variant **只做离线校验与诊断，不作运行期闸门**（A7）：运行期节点不自带
	// 表单名，拿它当闸门就是"全拒"。
	Form    FormName
	Variant Variant
	// WantFieldID：调用方已知期望编号时用它消歧（A7）。
	//
	// **nil = 没给**；`&0` 是**明确的编号 0**（A10）。为什么必须是指针：`fieldId` 可以
	// 合法为 0（实测 `{colName "b_index"} {fieldId "0"}` 就在真语料里），用 0 当"没给"
	// 会把那一份字段吞掉 —— 候选里有 {0, 12} 两种编号时将永远无法指名 0 那一份。
	WantFieldID *FieldID
	// 注意：`Resolved.FieldID` 是**节点自己的编号**，不是这个请求值（A10）。
}

// Resolved 是一次定位的结果 —— **单步句柄**（契约 C3.2 / C3.6 / H1）。
//
// ID 只在该次动作内有效，**不要跨步缓存**（运行期会复用裸 id）。
// 想判断"手上这个句柄还指着当初那个对象吗"，比 `Seq`：`rn` 之后或裸 id 被复用之后，
// 同一个 ID 的 Seq 一定不同（H2；判据与用例见 identity_test.go 的
// TestResolvedHandleDiesAfterRemove）。
type Resolved struct {
	ID      int
	ColName ColName
	FieldID FieldID
	Scope   Scope
	Seq     uint32
}

// ByColName 定位一个字段，返回**能直接拿去发事件的**会话句柄。
//
// 收窄顺序（**A7 + A9 + A10 + A11 合成，写死**）：
//
//	候选 = 窗口内（Scope.Window 非空则收窄）且 colName 相符的节点（**不限类型**，A9）
//	  0 个 → FIELD_NOT_IN_SCOPE (5)
//	  1 个 → 该节点不接受该动作 → FIELD_NOT_ACTIONABLE (4)；否则命中
//	  >1 个 → 先按"能否接受该动作"分两组：
//	            可接受组非空 → 只在可接受组里做 A7 的 fieldId 分组判定
//	                           （只有一种 fieldId ⇒ 取 Seq 最后声明；多种 ⇒ FIELD_AMBIGUOUS (5)）
//	            可接受组为空 → FIELD_NOT_ACTIONABLE (4)，候选里列 colName#fieldId + 节点类型
//	  另：Scope.WantFieldID 非 nil ⇒ 先按它过滤（唯一命中就直接用它；一个都不符 ⇒
//	      FIELD_NOT_IN_SCOPE，C3.4：判为另一张表单）
//
// 三处与 A7 原文的**有意**差异，都记在这里：
//
//   - **A9**：`kind` **不淘汰候选**。两张白名单（`acceptableFor`）回答的是"能不能把事件
//     发给它"，不是"它是不是那个字段"——混在一起会把整类明细列字段永久排除
//     （实测真语料里 `colName` 同时挂在 `FormField` 与 `TableColumn` 上）。所以这里先按
//     身份收齐候选，再让"可接受"只影响**多候选时的取舍**，取舍不掉的就如实报 A11 的码。
//   - `sc.WantFieldID` 在 A7 里写在"分组之后"的括号里。本函数把它当**过滤器**放在分组之前：
//     C3.4 说"colName 命中但 fieldId 不符 ⇒ 判为另一张表单 ⇒ 拒"，放在后面会让
//     "给了错的 fieldId"被分组规则救回来（照 A7 括号后的顺序就是"从不拒"），与 C3.4 冲突。
//     位置提前之后，括号那句"唯一命中就直接用它"照样成立（命中的候选都同号）。
//   - 节点**读不出 fieldId 就整条跳过**（`nodeFieldID`）：双键缺一半时不可寻址，
//     与 C1.6"缺属性条目跳过并计数、绝不猜"同一个口径。若它是唯一的命中，结果就是拒。
//
// **没有任何副作用**：只读模型、不发包（允许的例外只是 `Hint` 里那份诊断列表）。
func ByColName(m *Model, sc Scope, colName ColName, kind string) (Resolved, error) {
	if m == nil {
		return Resolved{}, refusal(CodeFormUnknown,
			"没有可用的 AUI 模型：这一屏的表单还取不到（会话没起，或还没收到任何 om 帧）",
			formShowHint(sc))
	}
	colName = ColName(strings.TrimSpace(string(colName)))
	if colName == "" {
		return Resolved{}, refusal(CodeFieldNotInScope,
			"没给栏位名（colName 为空）：定位必须靠字段身份，不能靠位置或顺序猜",
			formShowHint(sc))
	}

	// 1) 窗口 → colName（**A9：`kind` 不在这里淘汰候选**）。
	var cands []*Object
	for _, o := range m.Objects() {
		if sc.Window != "" && !windowNamed(m, o, sc.Window) {
			continue
		}
		if ColName(o.Attr("colName")) != colName {
			continue
		}
		if _, ok := nodeFieldID(o); !ok {
			continue // 没有可读的 fieldId：不可寻址（见函数注释第二条）
		}
		cands = append(cands, o)
	}

	// 2) sc.WantFieldID 只做过滤（C3.4；位置提前的理由见函数注释）。
	if sc.WantFieldID != nil {
		var kept []*Object
		for _, o := range cands {
			if id, _ := nodeFieldID(o); id == *sc.WantFieldID {
				kept = append(kept, o)
			}
		}
		cands = kept
	}

	switch {
	case len(cands) == 0:
		return Resolved{}, notInScope(m, sc, colName, kind)
	case len(cands) == 1:
		// 唯一候选：它不接受该动作 ⇒ 身份找到了但发不了（A11）。
		if !acceptableFor(kind, cands[0].Type) {
			return Resolved{}, notActionable(m, sc, colName, kind, cands)
		}
		return resolveOne(m, sc, cands[0]), nil
	}

	// 3) >1 个：**先按"能否接受该动作"分两组**（A9/A11）。
	var actionable []*Object
	for _, o := range cands {
		if acceptableFor(kind, o.Type) {
			actionable = append(actionable, o)
		}
	}
	if len(actionable) == 0 {
		// 一个都发不了：**不是**"没这个字段"，而是"该换节点发"（A11，退 4）。
		return Resolved{}, notActionable(m, sc, colName, kind, cands)
	}
	// 可接受组非空 ⇒ 只在**可接受组**里做 A7 的 fieldId 分组判定（不可接受的那些不参与，
	// 否则"后声明的明细列"会把好端端的可写字段顶掉）。
	groups := make(map[FieldID][]*Object, 2)
	for _, o := range actionable {
		id, _ := nodeFieldID(o)
		groups[id] = append(groups[id], o)
	}
	if len(groups) > 1 {
		return Resolved{}, ambiguous(m, sc, colName, kind, groups)
	}
	// 只有一种 fieldId ⇒ 同一字段的多帧副本 ⇒ 取 Seq 最后声明的那一份（C3.5）。
	return resolveOne(m, sc, newest(actionable)), nil
}

// ByColName 是契约 C3.3 的**方法形式**：`fep` 与 `view` 共用一个 `Resolver` 时从它进来
// （字段名解析也挂在这个类型上，见 fields.go）。实现就是上面那个自由函数 ——
// 收窄顺序只有一处，不做第二份。`r` 只是调用点既有的承载；调用方没给模型时用 `r.m`。
func (r *Resolver) ByColName(m *Model, sc Scope, colName ColName, kind string) (Resolved, error) {
	if m == nil && r != nil {
		m = r.m
	}
	return ByColName(m, sc, colName, kind)
}

// resolveOne 把命中的节点包成句柄（C3.2/C3.6）。
func resolveOne(m *Model, sc Scope, o *Object) Resolved {
	fid, _ := nodeFieldID(o)
	out := Resolved{
		ID:      o.ID,
		ColName: ColName(o.Attr("colName")),
		FieldID: fid,
		Scope:   sc,
	}
	// 句柄自述：调用方**没给**窗口名时补上实际命中的那个（调用方给了的原样保留）。
	// `WantFieldID` 是**调用方的请求**，句柄不许伪造它 —— 所以这里不碰它（A10）。
	if out.Scope.Window == "" {
		if win := m.WindowOf(o.ID); win != 0 {
			out.Scope.Window = m.WindowName(win)
		}
	}
	if o.Seq > 0 {
		out.Seq = uint32(o.Seq)
	}
	return out
}

// nodeFieldID 读节点上的表单内编号（与 colName 成对出现，见本文件顶部）。
//
// 缺一个或读不出数字 ⇒ 不可寻址：双键定位缺一半时**宁可拒，也不猜**（C1.6 的口径）。
func nodeFieldID(o *Object) (FieldID, bool) {
	if o == nil {
		return 0, false
	}
	t := strings.TrimSpace(o.Attr("fieldId"))
	if t == "" {
		return 0, false
	}
	n, err := strconv.Atoi(t)
	if err != nil {
		return 0, false
	}
	return FieldID(n), true
}

// windowNamed 判断节点 o 是不是属于"窗口名 = want"的那个窗口实例（A12）。
//
// **口径 = `m.WindowName(id)`**（A12 明确：就是既有那个 API 的口径），**不接**窗口的
// `name` 属性（"w_axmt500" 那种）也不接别的拼法 —— 认法只有一种，才不会有"同一个窗口
// 有两种写法、写成另一种就静默拒"的坑。
//
// 归属用 `WindowOf` 算（aui.go：回答"这个对象属于哪个窗口实例"，注释里写明了"驱动按窗口
// 收窄控件时用这一句"），**不是** `WindowName(o.ID)`：后者对非窗口对象回的是**它自己的
// `text`**（实测 `TableColumn` 的 text 是"单据别编号"，不是窗口名），拿它比窗口名会把所有
// 候选一次筛掉 —— 那正是"看起来对、其实全拒"的形态（`do.go:487` 的 `--window` 过滤就是
// 这个写法，值得单独修，但它不是本切片的文件）。
func windowNamed(m *Model, o *Object, want string) bool {
	win := m.WindowOf(o.ID)
	if win == 0 {
		return false // 归属不明：既然要求"在某个窗口内"，就不算它在里面
	}
	return m.WindowName(win) == want
}

// windowLabel 给一个窗口实例写个人读的名字（**只用于提示文本**，不参与判定）。
func windowLabel(m *Model, o *Object) string {
	if o == nil {
		return "?"
	}
	return windowLabelByID(m, m.WindowOf(o.ID))
}

// windowLabelByID 同上，直接按窗口 id。
//
// 显示用：先取 `WindowName`（A12 的口径），空则退回 `name` 属性，再空就写编号 ——
// 这一处只影响 Hint 里那串人读文字，不影响任何收窄判定。
func windowLabelByID(m *Model, win int) string {
	if win == 0 {
		return "（窗口未知）"
	}
	if t := m.WindowName(win); t != "" {
		return t
	}
	if w, ok := m.Get(win); ok && w.Attr("name") != "" {
		return w.Attr("name")
	}
	return fmt.Sprintf("窗口#%d", win)
}

// windowsWithColName 列出"还有哪些窗口里有这个 colName"（只读诊断）。
//
// 用途只有一个：把`Hint` 从"没找到"变成"你找错窗口了，它在那个窗口里" ——
// C4.3 要求 Hint 可操作，列候选是最省事的那种可操作。
func windowsWithColName(m *Model, colName ColName, kind string) []string {
	var out []string
	seen := map[int]bool{}
	for _, o := range m.Objects() {
		if ColName(o.Attr("colName")) != colName {
			continue
		}
		if _, ok := nodeFieldID(o); !ok {
			continue
		}
		if !acceptableFor(kind, o.Type) {
			continue
		}
		win := m.WindowOf(o.ID)
		if win == 0 || seen[win] {
			continue
		}
		seen[win] = true
		out = append(out, windowLabelByID(m, win))
	}
	sort.Strings(out)
	return out
}

// notActionable 报"身份找到了，但这个节点发不了"（A11：FIELD_NOT_ACTIONABLE / Exit 4）。
//
// 为什么不复用 FIELD_NOT_IN_SCOPE：两者的**下一步动作完全不同**——
// "没这个字段"要去查名字/窗口（多半是找错了），"发不了"要**换节点发**（明细列底下的子
// 控件）。自动化排障看得懂这个区别，才不会一直改参数重试（实测踩过：对着不接受事件的对象连发）。
func notActionable(m *Model, sc Scope, colName ColName, kind string, cands []*Object) error {
	// 候选必须写清**身份 + 节点类型**（A11）：`colName#fieldId@类型`。
	parts := make([]string, 0, len(cands))
	for _, o := range cands {
		id, _ := nodeFieldID(o)
		parts = append(parts, fmt.Sprintf("%s#%d@%s", colName, id, o.Type))
	}
	return refusal(CodeFieldNotActionable,
		fmt.Sprintf("字段 %q 的身份找到了%s，但**没有**一个候选能接受这个动作（%s）—— "+
			"这不是「没这个字段」，是「该把事件发到别的节点上」：%s",
			colName, scopeDesc(sc), kindDesc(kind), strings.Join(parts, "、")),
		"典型情形：这个身份落在**明细列**（`TableColumn`）上，而列本身不接受输入/点击 —— "+
			"要发到它底下的子控件（`Edit`/`CheckBox`）。先看子节点："+
			"`tt fep ui find --parent <列id>`（列 id = 上面候选里的那个会话 id）；"+
			"或 `tt fep ui tree --window <窗口名片段> --depth 3` 看这一屏的树；"+formShowHint(sc))
}

// notInScope 报"作用域里没有这个字段"（C4.2：FIELD_NOT_IN_SCOPE / Exit 5）。
func notInScope(m *Model, sc Scope, colName ColName, kind string) error {
	extra := ""
	if sc.Window != "" {
		if others := windowsWithColName(m, colName, kind); len(others) > 0 {
			extra = fmt.Sprintf("；这个 colName 在**别的窗口**里有：%s（去掉窗口收窄，或改用它）",
				strings.Join(others, "、"))
		}
	}
	return refusal(CodeFieldNotInScope,
		fmt.Sprintf("作用域里没有字段 %q：%s（动作 %s）。两种可能：这张表单没这个 colName；"+
			"或它的 fieldId 与你要的不是同一张表单（C3.4：colName 命中而 fieldId 不符也算不在作用域）",
			colName, scopeDesc(sc), kindDesc(kind)),
		formShowHint(sc)+extra)
}

// ambiguous 报"同一 colName 挂了多种 fieldId"（C4.2：FIELD_AMBIGUOUS / Exit 5）。
//
// 消息里**必须**列出候选（`colName#fieldId` + 窗口名）：A7 与 C4.3 都要求"列候选供复制"，
// 而且绝不自己挑一个 —— 挑错的那一次会发到另一个字段上，看起来像"程序没反应"。
func ambiguous(m *Model, sc Scope, colName ColName, kind string, groups map[FieldID][]*Object) error {
	ids := make([]int, 0, len(groups))
	for id := range groups {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)

	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, fmt.Sprintf("%s#%d@%s", colName, id, windowLabel(m, groups[FieldID(id)][0])))
	}
	return refusal(CodeFieldAmbiguous,
		fmt.Sprintf("字段 %q 在作用域里命中 %d 种 fieldId%s（动作 %s）—— 真歧义，不替你挑：%s",
			colName, len(ids), scopeDesc(sc), kindDesc(kind), strings.Join(parts, "、")),
		fmt.Sprintf("把编号显式给上再调一次（Scope.WantFieldID = 指向 %d 的指针，候选里的第一个）；%s",
			ids[0], formShowHint(sc)))
}

// scopeDesc 把作用域写成人读的一小段（诊断用）。
func scopeDesc(sc Scope) string {
	var parts []string
	if sc.Window != "" {
		parts = append(parts, fmt.Sprintf("窗口 %q", sc.Window))
	}
	if sc.Form != "" {
		parts = append(parts, fmt.Sprintf("表单 %q", sc.Form))
	}
	if sc.Variant != "" {
		parts = append(parts, fmt.Sprintf("变体 %q", sc.Variant))
	}
	if sc.WantFieldID != nil {
		parts = append(parts, fmt.Sprintf("编号 %d", *sc.WantFieldID))
	}
	if len(parts) == 0 {
		return "（没有收窄：全模型）"
	}
	return "（" + strings.Join(parts, "，") + "）"
}

// kindDesc 把 kind 写成动作名（诊断用）。
func kindDesc(kind string) string {
	switch kind {
	case "click":
		return "点击"
	case "input":
		return "输入"
	case "":
		return "不限"
	}
	return kind
}

// formShowHint 给出"自己核一遍"的具体命令（C4.3：Hint 必须可操作）。
//
// `Scope` 里**没有** env/module（A7 划定的作用域就是窗口/表单/变体），所以那两个位置照
// 契约写成占位符：命令本身是完整的，把环境与模块填上就能跑。
// 只写契约点名过的旗标（C5.2 的 `--json` 与 C5.1 的三件 `--env/--module/--form`）——
// 别的旗标不存在，写进提示就是骗人。
func formShowHint(sc Scope) string {
	form := string(sc.Form)
	if form == "" {
		form = "<表单>"
	}
	return fmt.Sprintf("先照字段清单核一遍：tt form show --env <环境> --module <模块> --form %s --json"+
		"（键名固定，可 grep colName/fieldId；环境与模块照 `tt env ls` 与会话登记填）", form)
}
