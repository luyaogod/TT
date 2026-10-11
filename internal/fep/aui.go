package fep

import (
	"sort"
	"strconv"
)

// Object is one merged AUI object.
type Object struct {
	ID     int
	Type   string
	Parent int
	// Scope 是它的**声明作用域**（声明时运行期告知的 currentWindow，
	// 拿不到时用父对象 id）。实测 FEP 的数字 id 不是全局身份，
	// 所以 "同一个 id 的两份不同对象" 要靠 Scope 区分。
	Scope int
	// Win 是它**最后一次被运行期触及时所处的当前窗口**（由 un 0 {{currentWindow "N"}} 得到）。
	// 为什么不用父链：实测开窗里的对象是运行期预载的，父链指向模板语境（
	// 一个 "交易对象编号" 列的父链会一路挂到别的窗口的 ComboBox 上），
	// 而运行期在开窗时会用 un 重新触及这些对象 —— 所以 "最后所在的窗口"才是
	// "哪个窗口在显示它"。这是 M1 的核心证据，见提交说明。
	//
	// **改写规则在 2026-10-11 收窄过一次**（原来无条件写）：只在
	//   (a) 本条消息**自己声明了**当前窗口（= 新实例构造，模板对象被重新触及），或
	//   (b) 它本来就属于那个窗口（Win/Scope 已经是它）
	// 时才改写。否则一条"在开窗的同时顺手更新主窗字段"的消息会把不相干的对象
	// 也记到新窗口名下 —— 实测症状：`ui roles --scope <料件窗口>` 里冒出供应商的列。
	Win int
	// DeclWindow 是**最近一次声明它的那条消息**所属的窗口。
	//
	// 它是归属的第一手证据：实测它对三种情形都准 ——
	//   · 主窗的明细网格 #644（声明于主窗的消息 → 178，即使它后来在同一条开窗消息里
	//     被顺手更新过）；
	//   · 开窗实例自己的网格（#122 → 119，#690/#3539/#3630 → 各自的实例）；
	//   · 两个开窗**共享**的模板列（formonly.oobxl003_2 → 声明它的那个开窗 687）。
	// 而父链对第三种会把它们指到别的实例（实测 #586 的链指向 155，而它属于 687）。
	//
	// 所以在**每次（重）声明**时更新："最近一次被运行期声明进哪个窗口"就是"它现在是谁的"。
	DeclWindow int
	// Seq 是这条对象"最后被声明或更新"的序号（模型内部单调自增）。
	//
	// 为什么需要它：实测运行期**会复用裸 id**，而且是在同一连接、同一窗口里：
	//
	//	{Action 4937 {{name "s_detail2.find"}}}   ← Dialog 的动作
	//	{an 376 Item 4937 {{name "7"}}}          ← 某下拉的第 7 项
	//
	// （同一个 4937，同一窗口，两种类型。）事件在线上只能给裸 id，所以驱动必须能回答
	// "#4937 **现在**是什么" —— 答案不是"先声明的那个"，而是"运行期**最后**声明/触碰的那个"，
	// 因为后来的声明就是运行期在用新对象覆盖这个 id。
	Seq   int
	Attrs map[string]string
}

// Attr returns the attribute value or "".
func (o *Object) Attr(key string) string {
	if o == nil || o.Attrs == nil {
		return ""
	}
	return o.Attrs[key]
}

// Model is the merged AUI of one or more connections.
//
// Merging rules are the ones validated against a live job (see the plan §2.10):
//
//   - a definition (`an <parent> <Type> <id>` or `<Type> <id>`) sets type and
//     parent; a parent seen as 0 may later be repaired to a real id, but a real
//     parent is never replaced by another one (a stray partial message must not
//     re-root an object into a different window);
//   - an update (`un` / `rn`) merges attributes only: it never changes type or
//     parent. This is what keeps field-name resolution stable;
//   - an update that arrives before its definition creates a stub that the
//     definition later completes.
type Model struct {
	objs      map[int]*Object
	order     []int
	conflicts int
	// locals 保存**同 id 但不同作用域**的声明（实测：同一个 110 既被声明成
	// Value 又被声明成 Dialog；#159 先 ActionDefault、后开窗第一列 TableColumn）。
	// 以前只能 "先声明者胜"，结果丢掉真对象（开窗第一列就是这样丢的）。
	locals     map[ObjKey]*Object
	localOrder []ObjKey
	// gone 记录被 rn 删掉的 id（诊断用；模型重建时会再声明回来）。
	gone map[int]bool
	// winTables 记"哪个窗口实例有哪些网格"：实测开窗的窗口与它的网格在同一条定义消息里声明
	// （an 0 Window 144 … {Table 145 …），而父链不可靠（会把开窗的列爬到主窗的网格上）。
	winTables map[int][]int
	// seq 是声明/更新的单调计数（配合 Object.Seq 决定"同一个裸 id 现在是谁"）。
	seq int
	// curWindow 是运行期最近一次告知的当前窗口（un 0 {{currentWindow "N"}}）。
	// 它后面的声明/更新都归到这个窗口下 —— 这是 M1 用来划分作用域的依据。
	curWindow int
	// dialogType 是**连接级**的"当前对话框阶段"（运行期最近一次推的值；见 DialogType）。
	// dialogTypeWin 是**推它的时候当前是哪个窗口** —— 这个值还会粘（实测开窗关掉后
	// 运行期不再复位它），所以它只能对"当时那个窗口"有效。
	dialogType    string
	dialogTypeWin int
}

// ObjKey 是对象的**作用域内身份**。
type ObjKey struct {
	Scope int
	ID    int
}

// NewModel returns an empty model.
func NewModel() *Model {
	return &Model{objs: make(map[int]*Object), locals: make(map[ObjKey]*Object), winTables: make(map[int][]int), gone: make(map[int]bool)}
}

// Len reports how many objects the model holds.
func (m *Model) Len() int { return len(m.objs) + len(m.locals) }

// Conflicts reports how often a definition tried to re-parent an existing
// object. Non-zero means the stream contained partial messages; the model keeps
// the first parent, so this is a diagnostic, not an error.
func (m *Model) Conflicts() int { return m.conflicts }

// nextSeq 推进声明/更新序号。
func (m *Model) nextSeq() int {
	m.seq++
	return m.seq
}

// Candidates 返回**所有**裸 id 等于 id 的对象（全局表 + 作用域表），按 Seq 降序（最新在前）。
//
// 存在的理由：运行期会复用裸 id（见 Object.Seq 的注释），所以"同一个 id"在模型里
// 可能有多份。想知道"现在是谁"用 Get；想知道"这个 id 被谁用过"用 Candidates（诊断/审计）。
func (m *Model) Candidates(id int) []*Object {
	var out []*Object
	if o, ok := m.objs[id]; ok {
		out = append(out, o)
	}
	for _, o := range m.locals {
		if o.ID == id {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Seq > out[j].Seq })
	return out
}

// newest 返回一批候选里 Seq 最大的那个（nil 表示空）。
func newest(objs []*Object) *Object {
	var best *Object
	for _, o := range objs {
		if best == nil || o.Seq > best.Seq {
			best = o
		}
	}
	return best
}

// Apply merges parsed nodes into the model and returns how many nodes were
// consumed.
//
// 三条规则都来自实测：
//  1. 对象身份 = **(作用域, id)**，不是裸 id：同一个 110 会同时是 Value 与 Dialog，
//     #159 先被声明为 ActionDefault、后又被声明为开窗第一列 TableColumn。
//     没有冲突时进全局表（事件就是按裸 id 寻址的），有冲突时按作用域进局部表。
//  2. 作用域优先用运行期推的 currentWindow；拿不到时用声明里的父对象 id。
//  3. 更新（un/rn）只合属性，不改类型/父 —— 这保证字段名解析稳定。
func (m *Model) Apply(nodes []Node) int {
	// 一条定义消息里若声明了顶层 Window，这批对象就属于**那个窗口实例**。
	// 实测：每个开窗实例拿新的 Window id（119 / 121 / 144），而模板对象（列/格子/动作）
	// 的 id 在实例间**复用** —— 所以作用域必须是"声明它的那个窗口 id"，
	// 否则几个实例会被并成一份（我因此一直把事件发到已关闭的旧实例上）。
	msgScope := 0
	for _, n := range nodes {
		if n.Type == "Window" && n.Parent == 0 && n.ID != 0 {
			msgScope = n.ID
		}
	}
	if msgScope != 0 {
		m.curWindow = msgScope
	}
	for _, n := range nodes {
		// 运行期告知的当前窗口（开窗/关窗时推）也能切作用域。
		if n.ID == 0 {
			if v := n.Attrs["currentWindow"]; v != "" {
				if w, err := strconv.Atoi(v); err == nil {
					m.curWindow = w
				}
			}
		}
		scope := msgScope
		if scope == 0 {
			scope = m.curWindow
		}
		if scope == 0 {
			scope = n.Parent
		}
		// 父=0 的声明（除顶层 Window 自己）是**连接级全局对象**：实测
		//
		//	{an 0 ActionDefaultList 93 {} {{ActionDefault 159 {{name "nextfield"} …}}}
		//	… {an 1986 Value 159 {{value ""}}}
		//
		// 同一个 159：一个是全局的 nextfield 动作默认值，一个是网格里第 159 号格子。
		// 以前给它们安上同一个 currentWindow 作用域，于是它们被当成"冲突" ——
		// 实际上它们从来不是一个东西。
		if n.Parent == 0 && n.Type != "Window" {
			scope = 0
		}
		switch n.Type {
		case "rn":
			// 官方 1522：AUI 同步只有三种操作，an = 加、un = 改、rn = **删**。
			// 实测流里就是"先删再加"：om 14 {{rn 4533} {an 119 Dialog 3533 …}} /
			// om 40 {{rn 3546} {an 121 Message 3551 …} {an 121 Dialog 3504 …}}。
			// 我以前把 rn 当 un 处理，被删的节点永远不死 —— 于是同一 id 残留多份，
			// 才有"#3555 既是 Item 又是 TableColumn"、accept 多个候选、focus 乱跳。
			m.remove(n.ID, scope)
		case "un":
			o := m.lookup(n.ID, scope)
			if o == nil {
				o = m.ensure(n.ID)
			}
			for k, v := range n.Attrs {
				o.Attrs[k] = v
			}
			// dialogType 是**连接级**状态（实测运行期同一时刻把它推给一大堆网格），
			// 所以还要在模型上记一份"当前阶段"，供驱动问"这一步在哪个阶段"。
			if v := n.Attrs["dialogType"]; v != "" {
				m.dialogType = v
				m.dialogTypeWin = m.curWindow
			}
			// 被运行期触及 = 它就是当前那份（见 Object.Seq）。
			o.Seq = m.nextSeq()
			// 归属只在两种情况下跟着走（见 Object.Win 的注释）：本条消息自己声明了
			// 当前窗口（新实例构造），或它本来就属于那个窗口。否则不动。
			if m.curWindow != 0 && n.ID != 0 && o.Type != "Window" {
				if msgScope != 0 || o.Win == m.curWindow || o.Scope == m.curWindow {
					o.Win = m.curWindow
				}
			}
		default:
			if n.ID == 0 {
				continue // 已当作全局状态处理（当前窗口）
			}
			// 顶层 Window 声明本身就该切换当前作用域：实测原始流里
			//     an 0 Window 144 {{name "w_qry_q_pmaa001_3"} {style "openwin"} …}
			// 紧接着就是属于它的 {Table 145 … —— 开窗的对象都在这条消息里声明。
			// 以前只靠运行期的 un 0 {{currentWindow "N"}} 切，结果开窗自己的对象
			// 被归到了旧作用域（实测 #145 的 scope 变成 1），于是我一直发到
			// 上一个已关闭实例的网格上（#156）。
			if n.Type == "Window" && n.Parent == 0 {
				m.curWindow = n.ID
				scope = n.ID
			}
			if msgScope != 0 && n.Type == "Table" {
				m.winTables[msgScope] = append(m.winTables[msgScope], n.ID)
			}
			o := m.objForDef(n, scope)
			if o.Scope == 0 {
				o.Scope = scope
			}
			if o.Type == "" || o.Type == "?" {
				o.Type = n.Type
			}
			// 声明顺序 = 身份判据（后声明的就是"现在这个"）。
			o.Seq = m.nextSeq()
			if m.curWindow != 0 && n.Type != "Window" {
				o.Win = m.curWindow
			}
			if m.curWindow != 0 {
				o.DeclWindow = m.curWindow
			}
			if n.Parent != 0 {
				switch {
				case o.Parent == 0:
					o.Parent = n.Parent
				case o.Parent != n.Parent:
					// 父链冲突：模板对象的父指向模板语境，不动它；只计数。
					m.conflicts++
				}
			}
			for k, v := range n.Attrs {
				o.Attrs[k] = v
			}
		}
	}
	return len(nodes)
}

// ApplyMessage parses and merges one message body.
func (m *Model) ApplyMessage(body string) int {
	return m.Apply(ParseMessage(body))
}

// ApplyFrame merges a decoded AUI frame. Non-AUI frames are ignored.
func (m *Model) ApplyFrame(f Frame) int {
	if f.Type != FrameTypeAUI {
		return 0
	}
	return m.ApplyMessage(string(f.Payload))
}

// remove 执行 AUI 同步的 `rn`（删节点）：**该裸 id 的所有候选**都清掉。
//
// 实测证据：运行期换 Dialog / 换工具栏 / 重建模板条目时都是"先 rn 再 an"。
//
// 为什么不管作用域：rn 说的是"运行期把这个 id 删了"。我算出来的作用域只是**我**的归属结论，
// 它可能是错的（模板对象的归属实测就经常错），而"这个 id 没了"是运行期说的硬事实。
// 早先只删作用域匹配的那一份，于是算错作用域时对象永远不死 —— 这正是
// "同一个 id 既是 Item 又是 Action""accept 三个候选""focus 乱跳"的来源。
func (m *Model) remove(id, scope int) {
	if id == 0 {
		return
	}
	_ = scope
	if _, ok := m.objs[id]; ok {
		delete(m.objs, id)
		m.gone[id] = true
	}
	for k := range m.locals {
		if k.ID == id {
			delete(m.locals, k)
			m.gone[id] = true
		}
	}
	// winTables 是"窗口↔它的网格"的账本；网格被删了账本也得跟着变，
	// 否则会拿着已删的网格 id 去发事件（实测踩过：发到已关闭实例的网格 #156 上）。
	for w, grids := range m.winTables {
		kept := grids[:0]
		for _, g := range grids {
			if g != id {
				kept = append(kept, g)
			}
		}
		m.winTables[w] = kept
	}
	m.conflicts++ // 删除是诊断信号（今天正是它被漏掉）
}

func (m *Model) ensure(id int) *Object {
	if o, ok := m.objs[id]; ok {
		return o
	}
	o := &Object{ID: id, Type: "?", Attrs: make(map[string]string)}
	m.objs[id] = o
	m.order = append(m.order, id)
	return o
}

// ensureLocal 取一个作用域内的局部对象（同 id 不同作用域的声明共存）。
func (m *Model) ensureLocal(key ObjKey, typ string) *Object {
	if o, ok := m.locals[key]; ok {
		return o
	}
	o := &Object{ID: key.ID, Scope: key.Scope, Type: typ, Attrs: make(map[string]string)}
	m.locals[key] = o
	m.localOrder = append(m.localOrder, key)
	return o
}

// lookup 按裸 id 找一个对象：先当前作用域那份，再当前窗口那份，最后全局/任意那份。
// （事件在线上只能给裸 id，所以更新要尽量落到"当前正在用的那一份"上。）
//
// 返回候选里**最新**的那一份（见 Object.Seq）：同一裸 id 被运行期复用时，
// 更新的应该是它此刻正在用的对象。
func (m *Model) lookup(id, scope int) *Object {
	var cand []*Object
	add := func(o *Object) {
		if o != nil {
			cand = append(cand, o)
		}
	}
	add(m.locals[ObjKey{scope, id}])
	if m.curWindow != 0 {
		for k, o := range m.locals {
			if k.ID == id && (k.Scope == m.curWindow || o.Win == m.curWindow) {
				add(o)
			}
		}
	}
	add(m.objs[id])
	for k, o := range m.locals {
		if k.ID == id {
			add(o)
		}
	}
	return newest(cand)
}

// objForDef 选一个可写的对象：
//
//   - 同名同作用域 → 合并到已有的；
//   - 其它情况（类型不同**或作用域不同**）→ 按 (作用域, id) 单独存一份。
//
// 为什么要看作用域而不只看类型：实测
//
//	… {Message 111 {{text "Verify OK!"}} {}} {Dialog 110 {{active "1"}} {{Action 109 {{name "accept"}}…}}
//	… {Value 107 {}} {Value 108 {}} {Value 109 {}} {Value 110 {{value ""}}} …
//
// 同一个 110 同时是"消息框的 Dialog"又在另一串 Value 里 —— 两棵树，不能并。
// 合并的后果实测过：Dialog 身上出现 value "电子零件类"，actionIdRef 还指到别的树的 accept。
func (m *Model) objForDef(n Node, scope int) *Object {
	if o, ok := m.objs[n.ID]; ok {
		sameType := o.Type == "" || o.Type == "?" || o.Type == n.Type
		sameScope := o.Scope == 0 || scope == 0 || o.Scope == scope
		if sameType && sameScope {
			return o
		}
		m.conflicts++
		return m.ensureLocal(ObjKey{scope, n.ID}, n.Type)
	}
	if o, ok := m.locals[ObjKey{scope, n.ID}]; ok {
		return o
	}
	return m.ensure(n.ID)
}

// Get returns an object by id：**最后声明/触碰的那一份**（见 Object.Seq）。
//
// 为什么不是"先声明者胜"：运行期会复用裸 id（同一个 4937 既做过下拉的 Item
// 又做过 Dialog 的 Action）。驱动问的是"#4937 现在是什么"，所以答案必须是
// 运行期最后一次给出的那个定义。
func (m *Model) Get(id int) (*Object, bool) {
	if o := newest(m.Candidates(id)); o != nil {
		return o, true
	}
	return nil, false
}

// Objects returns every object in first-seen order (global + scoped).
// 注意：`rn` 会删对象，order 里会留下已删 id —— 这里跳过它们（否则会 panic）。
func (m *Model) Objects() []*Object {
	out := make([]*Object, 0, m.Len())
	for _, id := range m.order {
		if o := m.objs[id]; o != nil {
			out = append(out, o)
		}
	}
	for _, k := range m.localOrder {
		if o := m.locals[k]; o != nil {
			out = append(out, o)
		}
	}
	return out
}

// ByType returns every object of one AUI type.
func (m *Model) ByType(t string) []*Object {
	var out []*Object
	for _, o := range m.Objects() {
		if o.Type == t {
			out = append(out, o)
		}
	}
	return out
}

// Chain walks from id towards the root. It reports whether the chain contains a
// cycle (which happens when partial messages are merged across windows) and the
// nearest Window id on the way, if any.
func (m *Model) Chain(id int) (chain []*Object, cycle bool, window int) {
	seen := make(map[int]bool)
	cur, ok := m.Get(id)
	for hops := 0; ok && hops < 64; hops++ {
		if seen[cur.ID] {
			return chain, true, window
		}
		seen[cur.ID] = true
		chain = append(chain, cur)
		if cur.Type == "Window" && window == 0 {
			window = cur.ID
		}
		if cur.Parent == 0 {
			return chain, false, window
		}
		cur, ok = m.Get(cur.Parent)
	}
	return chain, !ok && len(chain) > 0, window
}

// Window returns the window instance an object belongs to.
//
// 优先级：运行期标的"最后所在窗口"(Win) → 声明作用域(Scope，若它解析为 Window)
// → 父链上的 Window。第三种实测最不可靠（预载对象的父指向模板语境），所以只做兜底。
func (m *Model) Window(id int) (int, bool) {
	// Scope 是静态的"声明它的那个窗口实例"，最准；Win 是动态的"最后被触及时所在窗口"，
	// 会被跨实例的 un 更新带偏（实测 #94/#95 这种模板对象会漏到别的窗口下）。
	if o := m.lookup(id, m.curWindow); o != nil {
		if o.Scope != 0 {
			if s, ok := m.Get(o.Scope); ok && s.Type == "Window" {
				return s.ID, true
			}
		}
		if o.Win != 0 {
			return o.Win, true
		}
	}
	_, _, w := m.Chain(id)
	return w, w != 0
}

// WindowName returns the window title ("text" attribute) for a window id.
func (m *Model) WindowName(id int) string {
	if o, ok := m.objs[id]; ok {
		return o.Attr("text")
	}
	return ""
}

// Windows returns every Window object in first-seen order.
func (m *Model) Windows() []*Object { return m.ByType("Window") }

// Propagate 把运行期已标定的窗口归属**沿父子关系向下**传（父已定、子未定才传）。
//
// 为什么只向下：实测破的正是父链向上继承（预载对象的父指向模板语境），
// 所以不能拿父链去**推断**窗口；但"已确定属于某窗口的父对象，它的子对象也属于它"
// 是安全的 —— 实测这样开窗的列才会归到开窗名下。
func (m *Model) Propagate() int {
	total := 0
	for hop := 0; hop < 8; hop++ {
		n := 0
		for _, o := range m.Objects() {
			if o.Win != 0 || o.Parent == 0 || o.Type == "Window" {
				continue
			}
			p, ok := m.Get(o.Parent)
			if !ok || p.Win == 0 {
				continue
			}
			o.Win = p.Win
			n++
		}
		total += n
		if n == 0 {
			break
		}
	}
	return total
}

// CurrentWindow 返回运行期最近一次告知的当前窗口（un 0 {{currentWindow "N"}}）。
func (m *Model) CurrentWindow() int { return m.curWindow }

// TablesOfWindow 返回某个窗口实例声明过的网格（同一条定义消息里的 Table）。
// 用途：写格子前要像前端那样告诉运行期"这个网格现在显示着"，而网格 id 不能猜 ——
// 指错网格会把运行期带回错误的窗口语境（实测发成主窗的 #644 后，后面的动作全被丢）。
// WindowOf 回答"这个对象属于哪个窗口实例"—— 驱动按窗口收窄控件时用这一句。
//
// 三步，优先级**按实测数据**定（2026-10-11 在一份真会话上把三套证据逐条对过）：
//
//  1. `DeclWindow`（最近一次声明它的那条消息属于哪个窗口）—— 三种情形都准：
//     主窗的明细网格 #644 → 178（即使它后来在同一条开窗消息里被顺手更新过）、
//     开窗实例自己的网格 → 各自的实例、两个开窗共享的模板列 → 声明它的那个开窗；
//  2. 父链上最近的 `Window` 祖先（结构对象生来如此）；
//  3. `Win`（运行期最后一次把它当"某窗口的东西"触及）。
//
// 为什么不是父链优先：实测两个开窗共享的模板列（formonly.oobxl003_2）父链指向
// **另一个**开窗实例（#586 的链到 155，而它属于 687），父链优先会把它们算错。
// 三步都拿不到就回 0（不知道）—— 调用方照实降级，不许猜。
func (m *Model) WindowOf(id int) int {
	o, ok := m.Get(id)
	if !ok {
		return 0
	}
	if o.DeclWindow != 0 {
		// "还活着"必须是"那个 id 现在还是个 **Window**" —— 裸 id 会被复用
		// （实测 #60 被当成过 Item），只看"取得到"会把归属指向幽灵。
		if w, ok := m.Get(o.DeclWindow); ok && w.Type == "Window" {
			return o.DeclWindow
		}
	}
	cur := o
	for hop := 0; hop < 12; hop++ {
		if cur.Type == "Window" {
			return cur.ID
		}
		if cur.Parent == 0 {
			break
		}
		next, ok := m.Get(cur.Parent)
		if !ok {
			break
		}
		cur = next
	}
	if o.Win != 0 {
		if w, ok := m.Get(o.Win); ok && w.Type == "Window" {
			return o.Win
		}
	}
	return 0
}

func (m *Model) TablesOfWindow(w int) []int {
	var out []int
	seen := map[int]bool{}
	for _, o := range m.ByType("Table") {
		if m.WindowOf(o.ID) == w && !seen[o.ID] {
			seen[o.ID] = true
			out = append(out, o.ID)
		}
	}
	sort.Ints(out)
	return out
}

// DialogTypeFor 返回**窗口 win 此刻的对话框阶段**；不知道就回空串。
//
// 两条实测性质决定了这个形状：
//  1. **连接级**：运行期同一时刻把同一个值推给一大堆网格（un 164/un 166/un 803…
//     都是 {dialogType "DisplayArray"}），所以它是"当前阶段"，不是"某网格的属性"；
//  2. **会粘**：上一个对话框的阶段不会自动被清掉 —— 所以它只对"推它时那个窗口"有效。
//
// 但窗口**刚开出来**的那一刻它常常还没推（实测：窗口声明与 dialogType 不在同一条消息里）——
// 那种情况回空串（不知道），调用方应该**等一下再问**（tt fep do lookup 就是轮询），
// 而不是拿一个旧值顶替。
func (m *Model) DialogTypeFor(win int) string {
	if m.dialogType != "" && (m.dialogTypeWin == 0 || m.dialogTypeWin == win) {
		return m.dialogType
	}
	// 退化路径：直接看这个窗口的网格自己有没有被推过 dialogType
	for _, g := range m.TablesOfWindow(win) {
		if o, ok := m.Get(g); ok {
			if v := o.Attr("dialogType"); v != "" {
				return v
			}
		}
	}
	return ""
}
