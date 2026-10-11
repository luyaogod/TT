package fep

import (
	"fmt"
	"strings"
	"testing"
)

// node renders one object in the wire format the client actually uses:
//
//	Type id {{attrs…}} {{ {child} {child} }}      children are each wrapped
//
// Building the fixture programmatically keeps the brace levels honest: a hand
// written string with one brace too few silently flattens the tree, which is
// exactly the bug this test exists to catch.
func node(header string, attrs [][2]string, children ...string) string {
	var b strings.Builder
	b.WriteString(header)
	b.WriteString(" {")
	for _, kv := range attrs {
		fmt.Fprintf(&b, "{%s \"%s\"}", kv[0], kv[1])
	}
	b.WriteString("}")
	if len(children) == 0 {
		b.WriteString("{}")
		return b.String()
	}
	b.WriteString("{{")
	for _, c := range children {
		b.WriteString("{")
		b.WriteString(c)
		b.WriteString("}")
	}
	b.WriteString("}}")
	return b.String()
}

// demoAUI is synthetic and contains no customer data; its shape mirrors a real
// job: a form with a labelled field, a table whose columns carry their own text,
// a value list under a column, and a bare control that must inherit.
func demoAUI() string {
	field := node("FormField 400", [][2]string{{"name", "demo_t.demo001"}, {"col", "demo001"}},
		node("ButtonEdit 401", [][2]string{{"comment", "cmt_demo001"}, {"width", "12"}}),
	)
	labelA := node("Label 402", [][2]string{{"comment", "cmt_demo001"}, {"text", "演示字段一"}, {"posX", "10"}, {"posY", "2"}})
	labelB := node("Label 403", [][2]string{{"text", "无键说明"}, {"posX", "20"}, {"posY", "2"}})
	hbox := node("HBox 300", [][2]string{{"text", "演示分组"}}, field, labelA, labelB)

	colQty := node("TableColumn 501", [][2]string{{"name", "demo_t.demo007"}, {"colName", "demo007"}, {"text", "演示数量"}},
		node("Edit 502", [][2]string{{"comment", "cmt_demo007"}, {"width", "9"}}),
		node("ValueList 503", nil,
			node("Value 504", [][2]string{{"value", ""}}),
			node("Value 507", [][2]string{{"value", ""}}),
		),
	)
	colBare := node("TableColumn 505", [][2]string{{"name", "demo_t.demo008"}, {"colName", "demo008"}},
		node("Edit 506", nil),
	)
	table := node("Table 500", [][2]string{{"name", "demo_t.demo_tab"}}, colQty, colBare)
	form := node("Form 200", [][2]string{{"name", "demo_t"}, {"text", "演示作业"}}, hbox, table)
	return node("an 0 Window 100", [][2]string{{"name", "w_demo"}, {"text", "演示作业(w_demo)"}}, form)
}

// demoUpdates merges attributes and then tries to re-parent 504 from a stray
// partial message.
const demoUpdates = `{un 401 {{value "X"}}} {un 504 {{value "42"}}} {an 900 Value 504 {{value "42"}}}`

func demoModel(t *testing.T) *Model {
	t.Helper()
	m := NewModel()
	if n := m.ApplyMessage(demoAUI()); n == 0 {
		t.Fatal("no nodes parsed from the demo AUI")
	}
	m.ApplyMessage(demoUpdates)
	return m
}

func TestParseNestsEveryChildGroup(t *testing.T) {
	m := demoModel(t)
	cases := []struct {
		id, parent int
	}{
		{100, 0}, {200, 100}, {300, 200}, {400, 300}, {401, 400}, {402, 300},
		{403, 300}, {500, 200}, {501, 500}, {502, 501}, {503, 501}, {504, 503},
		{507, 503}, {505, 500}, {506, 505},
	}
	for _, c := range cases {
		o, ok := m.Get(c.id)
		if !ok {
			t.Fatalf("object %d missing", c.id)
		}
		if o.Parent != c.parent {
			t.Errorf("object %d parent = %d, want %d", c.id, o.Parent, c.parent)
		}
	}
}

func TestUpdatesMergeAttributesOnly(t *testing.T) {
	m := demoModel(t)
	o, _ := m.Get(401)
	if o.Type != "ButtonEdit" {
		t.Errorf("update changed type to %q", o.Type)
	}
	if o.Parent != 400 {
		t.Errorf("update changed parent to %d", o.Parent)
	}
	if o.Attr("value") != "X" {
		t.Errorf("update did not merge value: %q", o.Attr("value"))
	}
	if o.Attr("comment") != "cmt_demo001" {
		t.Errorf("update dropped an existing attribute")
	}
}

func TestStrayDefinitionDoesNotReRoot(t *testing.T) {
	m := demoModel(t)
	o, _ := m.Get(504)
	if o.Parent != 503 {
		t.Fatalf("object 504 was re-rooted to %d", o.Parent)
	}
	if m.Conflicts() == 0 {
		t.Error("expected the model to report the attempted re-parent")
	}
}

func TestChainReachesWindowWithoutCycle(t *testing.T) {
	m := demoModel(t)
	chain, cycle, win := m.Chain(504)
	if cycle {
		t.Fatal("chain reported a cycle")
	}
	if win != 100 {
		t.Fatalf("window = %d, want 100", win)
	}
	if len(chain) < 5 {
		t.Fatalf("chain too short: %d", len(chain))
	}
	if m.WindowName(100) != "演示作业(w_demo)" {
		t.Errorf("window title = %q", m.WindowName(100))
	}
}

func TestReparentAttemptIsRefused(t *testing.T) {
	// On a live job, stray partial messages re-rooted objects into other windows
	// and even created parent cycles. The model refuses the second parent, which
	// is what keeps field names stable.
	m := NewModel()
	m.Apply([]Node{{Type: "Window", ID: 1}, {Type: "Edit", ID: 2, Parent: 1}})
	m.Apply([]Node{{Type: "Edit", ID: 2, Parent: 3}})
	if o, _ := m.Get(2); o.Parent != 1 {
		t.Fatalf("object 2 was re-parented to %d", o.Parent)
	}
	if m.Conflicts() == 0 {
		t.Fatal("expected the refusal to be reported as a conflict")
	}
	if _, cycle, _ := m.Chain(2); cycle {
		t.Fatal("refusal must not create a cycle")
	}
}
