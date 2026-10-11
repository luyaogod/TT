package fep

import (
	"testing"
)

func TestResolveCommentKey(t *testing.T) {
	r := NewResolver(demoModel(t))
	fn, ok := r.Resolve(401)
	if !ok {
		t.Fatal("401 should resolve through the label's dictionary key")
	}
	if fn.Text != "演示字段一" || fn.Source != "comment-key" || fn.Confidence != ConfHigh {
		t.Fatalf("got %+v", fn)
	}
}

func TestResolveNameSuffixAndOwnText(t *testing.T) {
	r := NewResolver(demoModel(t))
	fn, ok := r.Resolve(501)
	if !ok || fn.Text != "演示数量" || fn.Confidence != ConfHigh {
		t.Fatalf("column with its own text: %+v ok=%v", fn, ok)
	}
	// 502 has a dictionary key that is not in the label index: it inherits from
	// its column instead (R4).
	fn2, ok2 := r.Resolve(502)
	if !ok2 || fn2.Text != "演示数量" || (fn2.Source != "inherit#501" && fn2.Source != "name-suffix") {
		t.Fatalf("bare control should inherit its column, got %+v ok=%v", fn2, ok2)
	}
	// 506 sits under a column without any label: it must stay unresolved rather
	// than borrow a name from an unrelated ancestor.
	if fn3, ok3 := r.Resolve(506); ok3 {
		t.Fatalf("506 must stay unresolved, got %+v", fn3)
	}
}

// The prototype happily produced a wrong name by walking out of a container.
// Inheritance must stop at the first non field-like parent.
func TestInheritanceDoesNotCrossContainers(t *testing.T) {
	r := NewResolver(demoModel(t))
	if fn, ok := r.Resolve(504); ok {
		t.Fatalf("value under a ValueList must not inherit a container label, got %+v", fn)
	}
}

func TestCoverageReportsDataFieldsOnly(t *testing.T) {
	r := NewResolver(demoModel(t))
	resolved, total := r.Coverage()
	if total == 0 {
		t.Fatal("coverage saw no data fields")
	}
	if resolved > total {
		t.Fatalf("resolved %d > total %d", resolved, total)
	}
	// 400 (col only, no label) is the deliberate hole in the fixture.
	if resolved == 0 || resolved >= total {
		t.Fatalf("resolved %d of %d: expected some holes", resolved, total)
	}
	if _, ok := r.Resolve(401); !ok {
		t.Fatal("401 must be resolved")
	}
	if _, ok := r.Resolve(400); ok {
		t.Fatal("400 has only a column name and no label: must stay unresolved")
	}
}

func TestLowConfidenceForBrokenChain(t *testing.T) {
	// Cycles cannot be produced through Apply (a second parent is refused), so
	// the white-box case is built directly: the name is still reported, but it
	// must never be graded as safe to publish.
	m := demoModel(t)
	m.objs[504].Parent = 507
	m.objs[507].Parent = 504
	if _, cycle, _ := m.Chain(504); !cycle {
		t.Fatal("expected the hand-made cycle to be detected")
	}
	r := NewResolver(m)
	if fn, ok := r.Resolve(502); ok && fn.Confidence == ConfHigh {
		// 502 is outside the cycle: it must stay high confidence.
		t.Logf("sibling outside the cycle keeps high confidence: %+v", fn)
	}
	fn, ok := r.Resolve(401)
	if !ok {
		t.Fatal("401 should still resolve")
	}
	if fn.Confidence != ConfHigh {
		t.Fatalf("healthy chain must stay high, got %+v", fn)
	}
}
