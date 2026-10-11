package fep

import (
	"fmt"
	"sort"
	"strings"
)

// 本文件是"稳定层"的模型：**角色寻址**，而不是裸 id。
//
// 为什么必须这样（实证）：FEP 里的数字 id 不是全局身份 —— 实测同一会话里
// "交易对象编号"只解析出一个 TableColumn #159，而它的祖先链是
//   #159 → #93 ValueList → #110 Value → #687 Value「单据别编号数据查询(q_ooba002_1)」
//        → #686 ComboBox BINCODE → #685 TableColumn BINCODE → #644 Table(采购明细)
// 也就是它其实是明细里 BINCODE 组合框的一个候选项。按裸 id 建模型必然张冠李戴。
//
// 官方依据（本地语料 D:/T100/4gl文档/BDL-Markdown/11_user-interface/）：
//   - 1522-the-front-end-protocol.md ：FEP 是**内部协议**，用于同步 AUI 树
//     （七步：前端把用户操作作为事件发回 → 运行期校验 → 回发 AUI 修改命令）
//   - 1514-actions-in-the-abstract-user-interface-tree.md ：交互元素**按 name 绑定**
//     到 Action 节点；当前可收到的动作由 Dialog 节点持有；触发时发 ActionEvent
//   - 1527-debugging-the-front-end-protocol.md ：协议输出格式仅供调试，**随版本可能变**
//
// 所以稳定层只依赖"官方语义 + 控件类型"，线形态单独成层、可替换。

// kindOps 是"控件类型 → 该类型支持的操作"。
//
// 操作名是**能力名**，不是动词流水：能不能做由类型决定，做了会怎样由运行期回报。
// 未列出的类型一律**不进目录**（宁缺不猜）。
var kindOps = map[string][]string{
	"Edit":           {"set"},
	"ButtonEdit":     {"set", "open"}, // open = 源码里的 ON ACTION controlp INFIELD <字段>
	"DateEdit":       {"set"},
	"ComboBox":       {"choose"},
	"Item":           {"choose"},
	"CheckBox":       {"toggle"},
	"RadioGroup":     {"choose"},
	"Value":          {"set"},  // 网格格子 / 值列表里的一个值
	"TableColumn":    {"cell"}, // 定位光标（配合该列下的 Value 写值）
	"Action":         {"invoke"},
	"Button":         {"invoke"},
	"ToolBarItem":    {"invoke"},
	"MenuAction":     {"invoke"},
	"TopMenuCommand": {"invoke"},
	"Menu":           {"invoke"},
	"TreeItem":       {"choose"},
}

// OpsOf 返回某控件类型支持的操作；第二返回值表示这个类型认不认识。
func OpsOf(kind string) ([]string, bool) {
	ops, ok := kindOps[kind]
	return ops, ok
}

// opsFor 在类型表基础上做**上下文限定**。
//
// 实测（正是本 demo 当场抓到的）：同为 Value，在值列表/组合框下是**候选项**
// （实测 #131\"请购采购单\"、#132\"一般采购单\" = 单据别开窗的选项），操作是 choose；
// 而在表格列下才是**格子**（写值 = set）。所以操作由类型 + 上下文共同决定。
func opsFor(m *Model, o *Object) ([]string, bool) {
	base, known := kindOps[o.Type]
	if !known {
		return nil, false
	}
	if o.Type != "Value" {
		return base, true
	}
	cur := o
	for hop := 0; hop < 6; hop++ {
		if cur.Parent == 0 {
			break
		}
		parent, ok := m.Get(cur.Parent)
		if !ok {
			break
		}
		switch parent.Type {
		case "ValueList", "ComboBox", "Item":
			return []string{"choose"}, true
		case "TableColumn":
			return []string{"set"}, true
		}
		cur = parent
	}
	return base, true
}

// Control 是目录里的一条：一个可操作的控件，按角色寻址。
type Control struct {
	ID       int      `json:"id"`   // 仅作落点提示；脚本里不要用
	Kind     string   `json:"kind"` // 官方 AUI 节点类型
	Role     string   `json:"role"` // 稳定寻址名（优先协议 name，其次中文名）
	RoleName string   `json:"roleName,omitempty"`
	Window   string   `json:"window"` // 所属窗口（当前由父链推出，可能不准，照实输出）
	Comment  string   `json:"comment,omitempty"`
	Value    string   `json:"value,omitempty"`
	Active   string   `json:"active,omitempty"`
	Hidden   string   `json:"hidden,omitempty"`
	NoEntry  string   `json:"noEntry,omitempty"` // 运行期广播的"不允许进入/录入"
	Ops      []string `json:"ops"`
	// 下面三个用于"能不能按"：运行期自己标的可用状态与真正的动作对象
	ActionActive string `json:"actionActive,omitempty"`
	ActionIDRef  string `json:"actionIdRef,omitempty"`
	Scope        int    `json:"scope,omitempty"`
	// Seq 是它"最后被声明/触碰"的序号 —— 同名多份（模板 vs 活实例）时按它决胜：
	// 最新的那份才是运行期在用的那个（与 Get(id) 同一口径）。
	Seq int `json:"seq"`
	// WindowID 是它所属的**窗口实例**（列/格子的 id 跨实例复用，驱动前要按它收窄）
	WindowID int `json:"windowId,omitempty"`
	// Source 是源码里对应的字段（.4fd），有它角色才稳定
	Source *SrcField `json:"source,omitempty"`
}

// RolePath 是脚本里该用的寻址串：窗口/角色。
func (c Control) RolePath() string {
	if c.Window == "" {
		return c.Role
	}
	return c.Window + "/" + c.Role
}

// Catalog 扫出可操作控件（按窗口 + 角色排序，输出稳定）。
//
// 过滤规则来自己实测：AUI 的根部有一大棵 StyleList 子树（#1 StyleList →
// Value → Value …，里面是 windowType/position/actionPanelPosition 这类呈现属性），
// 它们不是可操作控件；因此**样式子树一律不进目录**。
func Catalog(m *Model, r *Resolver) []Control {
	var out []Control
	for _, o := range m.Objects() {
		ops, known := opsFor(m, o)
		if !known {
			continue
		}
		if inStyleTree(m, o) {
			continue
		}
		// 不可见的控件不进目录（照实依据 hidden 属性）
		if h := o.Attr("hidden"); h == "1" {
			continue
		}
		role := strings.TrimSpace(o.Attr("name")) // 协议名：源码动作名/控件名
		display := role
		if fn, ok := r.Resolve(o.ID); ok && fn.Text != "" {
			display = fn.Text
			if role == "" {
				role = fn.Text
			}
		}
		// 列（TableColumn）是"定位用"的，不是值持有者：用它的协议名（如 pmdm_t.pmdm002）
		// 当角色，避免和同名的字段抢中文名（实测 付款条件 的列和字段撞在一起）。
		if o.Type == "TableColumn" {
			if n := o.Attr("name"); n != "" {
				role = n
			} else {
				role = "列:" + display
			}
		}
		winID := m.WindowOf(o.ID)
		if role == "" {
			role = o.Type
		}
		out = append(out, Control{
			ID: o.ID, Kind: o.Type, Role: role, RoleName: display,
			Window: m.WindowName(o.ID), Comment: o.Attr("comment"),
			Value: o.Attr("value"), Active: o.Attr("active"), Hidden: o.Attr("hidden"),
			NoEntry:      o.Attr("noEntry"),
			Scope:        o.Scope,
			Seq:          o.Seq,
			WindowID:     winID,
			Ops:          append([]string(nil), ops...),
			ActionActive: o.Attr("actionActive"), ActionIDRef: o.Attr("actionIdRef"),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Window != out[j].Window {
			return out[i].Window < out[j].Window
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// ResolveRole 把一个角色寻址串解析成**唯一**控件。
//
// 两轮匹配：先精确比 Role / RolePath；只有在**没有任何精确命中**时才用中文名（RoleName）兜底。
// 为什么分两轮：实测 付款条件 的字段（ButtonEdit）和列（TableColumn）中文名相同，
// 一轮混着比就会变成"多个命中"，而调用方其实说的是那个字段。
//
// 命中多个时**不猜**：把候选原样返回，由调用方限定（写全「窗口/角色」或用 --in）。
func ResolveRole(cat []Control, want string) (Control, []Control, error) {
	want = strings.TrimSpace(want)
	var exact, byName []Control
	for _, c := range cat {
		switch {
		case c.RolePath() == want, c.Window == "" && c.Role == want, c.Role == want:
			exact = append(exact, c)
		case c.RoleName == want:
			byName = append(byName, c)
		}
	}
	hits := exact
	if len(hits) == 0 {
		hits = byName
	}
	switch len(hits) {
	case 1:
		return hits[0], nil, nil
	case 0:
		return Control{}, nil, &roleError{msg: "找不到这个角色：" + want}
	default:
		// 同名多份的常见情形：**模板对象与活实例**（实测同一列/同一栏位会有两份，
		// 模板那份先声明、活实例后声明）。按与 Get(id) 同一条规则决胜：
		// **最后声明/触碰的那份才是运行期在用的那个**。
		//
		// 这样就不需要调用方每次写全「窗口/角色」或查 id 了 —— 而写 id恰恰是脆弱做法
		// （id 每次换状态都会重新分配）。如果"最新"这份不唯一，仍然回到"不猜"。
		if best, ok := newestOne(hits); ok {
			return best, nil, nil
		}
		return Control{}, hits, &roleError{msg: "这个角色有多个命中，不猜 —— 请用「窗口/角色」写全，或加 --in 限定"}
	}
}

// newestOne 在一组同名候选里挑出**最新声明**的那一个。
//
// 只有在"最新那份比其余都新"时才认（否则不猜）。窗口实例内的对象与模板对象
// 的窗口归属不同，所以不拿窗口比，只拿声明序号比 —— 这也是 Get(id) 的口径。
func newestOne(hits []Control) (Control, bool) {
	if len(hits) < 2 {
		return Control{}, false
	}
	best := hits[0]
	same := 0
	for _, c := range hits[1:] {
		if c.Seq > best.Seq {
			best = c
		}
	}
	for _, c := range hits {
		if c.Seq == best.Seq {
			same++
		}
	}
	// 最新的份必须唯一（两份同 Seq 说明拿不到先后，交给调用方）
	if same != 1 {
		return Control{}, false
	}
	// 而且必须**确实比第二新的更靠后**（都相同则无从判断）
	second := -1
	for _, c := range hits {
		if c.Seq != best.Seq && c.Seq > second {
			second = c.Seq
		}
	}
	if second < 0 || best.Seq <= second {
		return Control{}, false
	}
	return best, true
}

type roleError struct{ msg string }

// inStyleTree 判断一个对象是否在样式子树里（Style* 祖先，或 Value 套 Value 的嵌套）。
// 防环：AUI 的父链实测出现过来回指，所以限制跳数。
func inStyleTree(m *Model, o *Object) bool {
	cur := o
	for hop := 0; hop < 12 && cur != nil; hop++ {
		if strings.HasPrefix(cur.Type, "Style") {
			return true
		}
		if cur.Parent == 0 {
			return false
		}
		parent, ok := m.Get(cur.Parent)
		if !ok {
			return false
		}
		// Value 里再套 Value：呈现属性的写法，不是格子
		if cur.Type == "Value" && parent.Type == "Value" {
			return true
		}
		cur = parent
	}
	return false
}

func (e *roleError) Error() string { return e.msg }

// FirstCellOf 返回某一列（TableColumn）当前可用于写值的**格子**。
//
// 实测结构：列下面挂一个 ValueList，它的子 Value 才是格子
// （开窗网格：TableColumn #172 → ValueList #127 → Value 126；
//
//	明细网格：TableColumn → ValueList #1980 → Value 1981…）。
//
// 前端写格子时会**同时**发两条事件：值发给格子、光标落在列上（见 do cell 命令）。
func FirstCellOf(m *Model, columnID int) (int, error) {
	return FirstCellOfIn(m, columnID, 0)
}

// FirstCellOfIn 同上，但只认某个作用域里的孩子。
//
// 为什么需要 scope：同一批 id 在不同树里会被重复使用（实测 #3555 既是料件开窗的
// TableColumn，又在另一棵树里是 Item），不过滤就会把值写到别的树的格子上（实测 #96）。
func FirstCellOfIn(m *Model, columnID, scope int) (int, error) {
	inScope := func(o *Object) bool { return scope == 0 || o.Scope == 0 || o.Scope == scope }
	var listID int
	for _, o := range m.Objects() {
		if o.Parent == columnID && o.Type == "ValueList" && inScope(o) {
			listID = o.ID
			break
		}
	}
	if listID == 0 {
		return 0, &roleError{msg: fmt.Sprintf("列 #%d 下没有 ValueList（这列可能不是可写网格列）", columnID)}
	}
	// 取**最近被触碰**的那一格（Seq 最大），不是 id 最小那一格。
	//
	// 依据是真机语料（不是推断）：`{ value "Y"}` 这类写格子的事件，其目标节点
	// 总是**紧跟在**运行期 `currentRow` 变更之后被 `un` 触碰的那个 Value；
	// 而同一个裸 id 会被复用（旧实例的格子还在模型里），按 id 最小去挑必然挑到
	// 旧实例 —— 实测症状就是空 ACK，严重时 -6327 把作业打停。
	// 见 tools/gdc-uia/README.md「真机语料里"勾选"的原始字节」。
	best, bestSeq := 0, -1
	for _, o := range m.Objects() {
		if o.Parent == listID && o.Type == "Value" && o.ID > 0 && inScope(o) {
			if o.Seq > bestSeq {
				best, bestSeq = o.ID, o.Seq
			}
		}
	}
	if best == 0 {
		return 0, &roleError{msg: fmt.Sprintf("列 #%d 的 ValueList(#%d) 里没有格子（这一行可能还不存在）", columnID, listID)}
	}
	// 取**最近被触碰**的那一格（Seq 最大），不是 id 最小那一格，也不是"第 N 个"。
	//
	// 两种错法的实测教训：
	//   · 按 id 最小：挑到模板/旧实例（实测 #34）；
	//   · 按 currentRow 取第 N 个：列表里混着旧实例的格子（实测同一列下有 #2096 seq=805 与
	//     #2095 seq=6316 两份），按下标会挑到旧的那份 —— 写下去运行期"效果：无"。
	//
	// 真正对齐的是：**运行期刚碰过的那一份就是当前行的活格子**
	//（换行/改值都会 un 它，所以它的 Seq 最新）。
	return best, nil
}

// currentRowOfGrid 读网格的运行期当前行（运行期推的属性；读不到返回 0）。
func (m *Model) currentRowOfGrid(grid int) int {
	if grid == 0 {
		return 0
	}
	o, ok := m.Get(grid)
	if !ok {
		return 0
	}
	return atoiOr(o.Attr("currentRow"), -1)
}

// CellsIn 返回某个网格列下**按 Seq 升序**排列的格子（= 行序）。
//
// 行序只能按"声明/触碰顺序"，不能按 id：实测料件开窗的勾选格按 id 是
// #59/#105/#106…（旧实例残留），按 Seq 才是当前页的第 1/2/3 行。
func CellsIn(m *Model, columnID, scope int) []*Object {
	inScope := func(o *Object) bool { return scope == 0 || o.Scope == 0 || o.Scope == scope }
	listIDs := map[int]bool{}
	for _, o := range m.Objects() {
		if o.Parent == columnID && o.Type == "ValueList" && inScope(o) {
			listIDs[o.ID] = true
		}
	}
	var out []*Object
	for _, o := range m.Objects() {
		if o.Type == "Value" && o.ID > 0 && listIDs[o.Parent] && inScope(o) {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Seq != out[j].Seq {
			return out[i].Seq < out[j].Seq
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// NearestAncestor 往上找最近的某个类型的祖先（防环）。
func NearestAncestor(m *Model, id int, types ...string) (*Object, bool) {
	cur, ok := m.Get(id)
	for hop := 0; ok && cur != nil && hop < 24; hop++ {
		for _, t := range types {
			if cur.Type == t {
				return cur, true
			}
		}
		if cur.Parent == 0 {
			break
		}
		cur, ok = m.Get(cur.Parent)
	}
	return nil, false
}

// DialogOf 返回某个控件的最近 Dialog 祖先（= 它属于哪个交互指令）。
func DialogOf(m *Model, id int) (int, bool) {
	d, ok := NearestAncestor(m, id, "Dialog")
	if !ok {
		return 0, false
	}
	return d.ID, true
}

// ActiveDialogs 返回运行期标了 active=1 的 Dialog（当前交互指令的候选）。
//
// 官方依据（1514 Actions in the abstract user interface tree）：AUI 用一组 Action 节点
// 表示"当前交互指令"能收到的动作，这组动作由 Dialog 节点持有；交互元素按 name 绑定到
// Action 节点。所以"能不能发某个动作"要看它在不在当前 Dialog 的动作集里。
func (m *Model) ActiveDialogs() []*Object {
	var out []*Object
	for _, o := range m.Objects() {
		if o.Type == "Dialog" && o.Attr("active") == "1" {
			out = append(out, o)
		}
	}
	return out
}

// DialogActionIDs 返回某个 Dialog 直接持有的 Action/MenuAction/IdleAction 的 id 集合。
func DialogActionIDs(m *Model, dialogID int) map[int]bool {
	out := map[int]bool{}
	cur, ok := m.Get(dialogID)
	if !ok {
		return out
	}
	// 直接子对象（用父链，Dialog 的动作在同一棵树内声明，实测可靠）
	for _, o := range m.Objects() {
		if o.Parent != cur.ID {
			continue
		}
		switch o.Type {
		case "Action", "MenuAction", "IdleAction":
			out[o.ID] = true
		}
	}
	return out
}

// ActionNameInDialog 在当前（active）Dialog 的动作集里按 name 找一个动作。
//
// 这就是"按 name 绑定"的落地：不去猜 id，而是问"当前交互指令提供名字为 accept 的动作吗"。
// 找不到就返回 false —— 调用方据此拒绝发送（宁缺不猜）。
//
// 关键（实测）：运行期会**为每个开窗实例新建自己的 Dialog**，并自带一套动作：
//
//	om 40 {an 121 Dialog 3504 {active "1"}} {{Action 3505 {name "accept"} {text "确定"} …}}
//	om 44 {an 121 Dialog 3546 {active "1"}} {{Action 3535 {name "accept"} …}}
//
// 而模板里那套旧的（如 #109）虽然也叫 accept，却**不属于当前指令**，发了会被丢（只回 18 字节）。
// 所以优先选"属于当前窗口实例"的那个 Dialog（父或作用域 = curWindow）。
func ActionNameInDialog(m *Model, name string) (int, bool) {
	cur := m.CurrentWindow()
	pick := func(matchWindow bool) (int, bool) {
		for _, d := range m.ActiveDialogs() {
			inWin := d.Parent == cur || d.Scope == cur
			if matchWindow != inWin {
				continue
			}
			for _, o := range m.Objects() {
				if o.Parent != d.ID {
					continue
				}
				switch o.Type {
				case "Action", "MenuAction", "IdleAction":
				default:
					continue
				}
				if o.Attr("name") == name {
					return o.ID, true
				}
			}
		}
		return 0, false
	}
	if id, ok := pick(true); ok {
		return id, true
	}
	// 实战教训：运行期**每换一次状态就替换 Dialog 并重新分配动作 id**
	// （实测同一个"确定"：模板 #109 → 实例 #3535 → 查询后 #4533）。
	// 所以当多个候选同名时，取**最新分配的那个（id 最大）**：
	// 发送旧的那个会被运行期当"不属于当前指令"丢掉（只回 18 字节空 ACK）。
	best, found := 0, false
	for _, d := range m.ActiveDialogs() {
		for _, o := range m.Objects() {
			if o.Parent != d.ID || o.Type != "Action" && o.Type != "MenuAction" && o.Type != "IdleAction" {
				continue
			}
			if o.Attr("name") != name {
				continue
			}
			if !found || o.ID > best {
				best, found = o.ID, true
			}
		}
	}
	return best, found
}

// FilterKind 只保留某种类型的控件（给"格子操作只在列里选"这类场景用）。
func FilterKind(cat []Control, kind string) []Control {
	var out []Control
	for _, c := range cat {
		if c.Kind == kind {
			out = append(out, c)
		}
	}
	return out
}

// GridOfColumn 找到某一列所属的网格（Table）——用于在写格子前按前端那样告诉运行期
// "这个网格现在显示着"（实测前端会发 {网格}{pageSize}{bufferSize}，且在**当前实例**的网格上）。
//
// 先沿父链找同作用域的 Table；不行就在同一作用域里找第一个 Table。
func GridOfColumn(m *Model, columnID int) int {
	col, ok := m.Get(columnID)
	if !ok {
		return 0
	}
	scope := col.Scope
	cur := col
	for hop := 0; hop < 12; hop++ {
		if cur.Type == "Table" && (scope == 0 || cur.Scope == scope) {
			return cur.ID
		}
		if cur.Parent == 0 {
			break
		}
		next, ok := m.Get(cur.Parent)
		if !ok {
			break
		}
		// 父链实测会把开窗的列爬到主窗的网格上（#644）；一旦要跨作用域就别再跟。
		if scope != 0 && next.Scope != 0 && next.Scope != scope {
			break
		}
		cur = next
	}
	// 兜底：同一作用域里的第一个 Table（实测开窗的窗口与它的网格在同一条消息里声明）。
	if scope != 0 {
		for _, o := range m.Objects() {
			if o.Type == "Table" && o.Scope == scope {
				return o.ID
			}
		}
	}
	// 拿不准就不发 —— 宁可少发一条，也不能把运行期带回错误的窗口语境。
	return 0
}

// FilterScope 只保留属于某个窗口实例的控件。
//
// 判据是 WindowOf（见它的注释）：**不再**拿 `Scope` 兑底 —— 实测模板对象的 Scope
// 停在旧实例上（新实例只是重新触及它们），拿 Scope 或它就会把旧实例的列当成当前实例的，
// 症状是 `ui roles --scope <料件窗口>` 里冒出供应商的列。
func FilterScope(cat []Control, window int) []Control {
	var out []Control
	for _, c := range cat {
		if c.WindowID == window {
			out = append(out, c)
		}
	}
	return out
}

// ColumnIndexOf 返回某一列在它所属网格里的**列下标**（0 起，按声明顺序 = 显示顺序）。
//
// 为什么必须有它：真语料里前端写/勾一个格子之前，一定先发
// `{ConfigureEvent{idRef <网格>}{ currentColumn "c"}{ currentRow "r" }}` ——
// 运行期靠这两个属性知道"现在动的是哪一格"。少了 currentColumn，
// 值发过去运行期不改任何节点（实测：写明细格子与写开窗勾选格都"无效果"）。
func ColumnIndexOf(m *Model, columnID int) int {
	grid := GridOfColumn(m, columnID)
	if grid == 0 {
		return -1
	}
	var cols []*Object
	for _, o := range m.Objects() {
		if o.Parent == grid && o.Type == "TableColumn" {
			cols = append(cols, o)
		}
	}
	sort.Slice(cols, func(i, j int) bool {
		if cols[i].Seq != cols[j].Seq {
			return cols[i].Seq < cols[j].Seq
		}
		return cols[i].ID < cols[j].ID
	})
	for i, c := range cols {
		if c.ID == columnID {
			return i
		}
	}
	return -1
}

// AtoiOr 是 atoiOr 的导出版本（CLI 侧读运行期属性时要用同一个口径）。
func AtoiOr(s string, def int) int { return atoiOr(s, def) }

// GridColumnCount 返回网格下 TableColumn 的个数（算"Tab 几次能走到某列"要用）。
func GridColumnCount(m *Model, grid int) int {
	n := 0
	for _, o := range m.Objects() {
		if o.Parent == grid && o.Type == "TableColumn" {
			n++
		}
	}
	return n
}
