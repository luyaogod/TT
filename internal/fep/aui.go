package fep

import "strconv"

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
	Win   int
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
	// curWindow 是运行期最近一次告知的当前窗口（un 0 {{currentWindow "N"}}）。
	// 它后面的声明/更新都归到这个窗口下 —— 这是 M1 用来划分作用域的依据。
	curWindow int
}

// ObjKey 是对象的**作用域内身份**。
type ObjKey struct {
	Scope int
	ID    int
}

// NewModel returns an empty model.
func NewModel() *Model {
	return &Model{objs: make(map[int]*Object), locals: make(map[ObjKey]*Object)}
}

// Len reports how many objects the model holds.
func (m *Model) Len() int { return len(m.objs) + len(m.locals) }

// Conflicts reports how often a definition tried to re-parent an existing
// object. Non-zero means the stream contained partial messages; the model keeps
// the first parent, so this is a diagnostic, not an error.
func (m *Model) Conflicts() int { return m.conflicts }

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
	for _, n := range nodes {
		// 运行期告知的当前窗口：从此之后的声明/更新都归到它名下。
		if n.ID == 0 {
			if v := n.Attrs["currentWindow"]; v != "" {
				if w, err := strconv.Atoi(v); err == nil {
					m.curWindow = w
				}
			}
		}
		scope := m.curWindow
		if scope == 0 {
			scope = n.Parent
		}
		switch n.Type {
		case "un", "rn":
			o := m.lookup(n.ID, scope)
			if o == nil {
				o = m.ensure(n.ID)
			}
			for k, v := range n.Attrs {
				o.Attrs[k] = v
			}
			if m.curWindow != 0 && n.ID != 0 && o.Type != "Window" {
				o.Win = m.curWindow
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
			o := m.objForDef(n, scope)
			if o.Scope == 0 {
				o.Scope = scope
			}
			if o.Type == "" || o.Type == "?" {
				o.Type = n.Type
			}
			if m.curWindow != 0 && n.Type != "Window" {
				o.Win = m.curWindow
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

// lookup 按裸 id 找一个对象：先当前作用域那份，再次当前窗口那份，最后全局那份。
// （事件在线上只能给裸 id，所以更新要尽量落到"当前正在用的那一份"上。）
func (m *Model) lookup(id, scope int) *Object {
	if o, ok := m.locals[ObjKey{scope, id}]; ok {
		return o
	}
	if m.curWindow != 0 {
		for k, o := range m.locals {
			if k.ID == id && (k.Scope == m.curWindow || o.Win == m.curWindow) {
				return o
			}
		}
	}
	if o, ok := m.objs[id]; ok {
		return o
	}
	for k, o := range m.locals {
		if k.ID == id {
			return o
		}
	}
	return nil
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

// Get returns an object by id (global table first, then any scoped one).
func (m *Model) Get(id int) (*Object, bool) {
	if o, ok := m.objs[id]; ok {
		return o, true
	}
	if o := m.lookup(id, m.curWindow); o != nil {
		return o, true
	}
	return nil, false
}

// Objects returns every object in first-seen order (global + scoped).
func (m *Model) Objects() []*Object {
	out := make([]*Object, 0, m.Len())
	for _, id := range m.order {
		out = append(out, m.objs[id])
	}
	for _, k := range m.localOrder {
		out = append(out, m.locals[k])
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

// Window returns the nearest enclosing Window id.
// 优先用运行期标的"当前窗口"（Win）；没有时才回退到父链。
func (m *Model) Window(id int) (int, bool) {
	if o, ok := m.objs[id]; ok && o.Win != 0 {
		return o.Win, true
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
