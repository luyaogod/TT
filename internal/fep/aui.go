package fep

// Object is one merged AUI object.
type Object struct {
	ID     int
	Type   string
	Parent int
	Attrs  map[string]string
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
func (m *Model) Apply(nodes []Node) int {
	for _, n := range nodes {
		switch n.Type {
		case "un", "rn":
			o := m.ensure(n.ID)
			for k, v := range n.Attrs {
				o.Attrs[k] = v
			}
		default:
			o := m.ensure(n.ID)
			o.Type = n.Type
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
func (m *Model) Window(id int) (int, bool) {
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
