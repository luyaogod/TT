package fep

import "strconv"

// Object is one merged AUI object.
type Object struct {
	ID     int
	Type   string
	Parent int
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
	// curWindow 是运行期最近一次告知的当前窗口（un 0 {{currentWindow "N"}}）。
	// 它后面的声明/更新都归到这个窗口下 —— 这是 M1 用来划分作用域的依据。
	curWindow int
}

// NewModel returns an empty model.
func NewModel() *Model {
	return &Model{objs: make(map[int]*Object)}
}

// Len reports how many objects the model holds.
func (m *Model) Len() int { return len(m.objs) }

// Conflicts reports how often a definition tried to re-parent an existing
// object. Non-zero means the stream contained partial messages; the model keeps
// the first parent, so this is a diagnostic, not an error.
func (m *Model) Conflicts() int { return m.conflicts }

// Apply merges parsed nodes into the model and returns how many nodes were
// consumed.
//
// 实测得来的两条规则：
//  1. 顶层**无 an 前缀**的声明（Value 110 / Dialog 110 同时出现）是局部节点，
//     id 不是全局身份 —— 不进全局模型，连它的子节点一并跳过（否则父链会被揉到一起，
//     实测就是这么把一个组合框候选项当成了"交易对象编号"列）。
//  2. 只认带 an 的声明与 un/rn 更新；同一个 id 被不同父重复声明时保留第一个，
//     并计一次 conflicts（作为诊断保留）。
func (m *Model) Apply(nodes []Node) int {
	skipped := map[int]bool{}
	for _, n := range nodes {
		if skipped[n.Parent] {
			skipped[n.ID] = true
			continue
		}
		if n.Bare && n.Parent == 0 {
			// 无 an 的顶层声明是**模板子树**：它们是真对象（实测开窗第一列就是它，
			// 前端写格子时也写向它的子节点），只是 id 不唯一（同一个 110 会同时是
			// Value 与 Dialog）。这里的取舍：**先声明者胜**（保留第一个），后来的
			// 同名声明只当冲突计数 —— 冲突数是诊断。
			//
			// 真正的解法是 (作用域, id) 双键（待重构，见提交说明）：现在靠
			// 运行期的 currentWindow 归属 + 角色的协议 name 把影响限制住。
		}
		// 运行期告知的当前窗口：从此之后的声明/更新都归到它名下。
		if n.ID == 0 {
			if v := n.Attrs["currentWindow"]; v != "" {
				if w, err := strconv.Atoi(v); err == nil {
					m.curWindow = w
				}
			}
		}
		switch n.Type {
		case "un", "rn":
			o := m.ensure(n.ID)
			for k, v := range n.Attrs {
				o.Attrs[k] = v
			}
			if m.curWindow != 0 && n.ID != 0 && o.Type != "Window" {
				o.Win = m.curWindow
			}
		default:
			o := m.ensure(n.ID)
			if o.Type != "" && o.Type != "?" && o.Type != n.Type {
				m.conflicts++ // 同一 id 的异型声明：先声明者胜（诊断保留）
			}
			if o.Type == "" || o.Type == "?" {
				o.Type = n.Type
			}
			if m.curWindow != 0 && n.ID != 0 && n.Type != "Window" {
				o.Win = m.curWindow
			}
			if n.ID == 0 {
				continue // 已当作全局状态处理（当前窗口）
			}
			if n.Parent != 0 {
				switch {
				case o.Parent == 0:
					o.Parent = n.Parent
				case o.Parent != n.Parent:
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

// Get returns an object by id.
func (m *Model) Get(id int) (*Object, bool) {
	o, ok := m.objs[id]
	return o, ok
}

// Objects returns every object in first-seen order.
func (m *Model) Objects() []*Object {
	out := make([]*Object, 0, len(m.order))
	for _, id := range m.order {
		out = append(out, m.objs[id])
	}
	return out
}

// ByType returns every object of one AUI type.
func (m *Model) ByType(t string) []*Object {
	var out []*Object
	for _, id := range m.order {
		if m.objs[id].Type == t {
			out = append(out, m.objs[id])
		}
	}
	return out
}

// Chain walks from id towards the root. It reports whether the chain contains a
// cycle (which happens when partial messages are merged across windows) and the
// nearest Window id on the way, if any.
func (m *Model) Chain(id int) (chain []*Object, cycle bool, window int) {
	seen := make(map[int]bool)
	cur, ok := m.objs[id]
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
		cur, ok = m.objs[cur.Parent]
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
			p, ok := m.objs[o.Parent]
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
