package fep

import "testing"

// TestVerbPrefixIsNotAnObject guards a silent failure mode: the wire carries a
// verb in front of the message (`om 3 {{…}}`, `meta Connection {{…}}`). If the
// verb were parsed as data, a whole screen would collapse into one object of
// type "om" with no name and no window.
func TestVerbPrefixIsNotAnObject(t *testing.T) {
	body := `om 3 {{an 0 Window 4097 {{name "axmt500_wf"}} {{Entry 5157 {{text "录入"}} {} {Value 260 {{col "xmda001"}} {}}}}}`
	nodes := ParseMessage(body)
	if len(nodes) == 0 {
		t.Fatal("no nodes parsed")
	}
	if nodes[0].Type != "Window" {
		t.Fatalf("root type = %q, want Window (the verb leaked into the parse)", nodes[0].Type)
	}
	m := NewModel()
	m.ApplyMessage(body)
	for _, id := range []int{4097, 5157, 260} {
		if _, ok := m.Get(id); !ok {
			t.Fatalf("object %d missing", id)
		}
	}
	chain, cycle, win := m.Chain(260)
	if cycle {
		t.Fatal("chain reports a cycle")
	}
	if win != 4097 || len(chain) < 3 {
		t.Fatalf("Value 260 must hang under Window 4097: window=%d chain=%d", win, len(chain))
	}
}

// TestBareObjectHeaderIsNotAVerb makes sure the verb skip does not eat a body
// that opens with an object header: cutting at the first brace would throw the
// type and the id away.
func TestBareObjectHeaderIsNotAVerb(t *testing.T) {
	body := `Table 501 {{col "xmdc007"}} {{TableColumn 502 {{name "xmdc_t.xmdc007"}{text "订购量"}} {}}}`
	nodes := ParseMessage(body)
	if len(nodes) < 2 {
		t.Fatalf("nodes = %d", len(nodes))
	}
	if nodes[0].Type != "Table" || nodes[0].ID != 501 {
		t.Fatalf("root = %s#%d, want Table#501", nodes[0].Type, nodes[0].ID)
	}
}

// TestTrimNoiseHandlesPrototypeRecordings covers the recordings written by the
// earlier prototype: part of the frame header stayed in front of the message.
func TestTrimNoiseHandlesPrototypeRecordings(t *testing.T) {
	noisy := "\x00\x00\x00+   +om 0 {{an 0 UserInterface 0 {{name \"axmt500_wf\"}}}}"
	got, ok := TrimNoise(noisy)
	if !ok || got[:2] != "om" {
		t.Fatalf("TrimNoise = %q,%v", got, ok)
	}
	if _, ok := TrimNoise("\x00\x00\x18\x05\x06\xff!!__cached__!!"); ok {
		t.Fatal("a proxy marker must not pass as a message")
	}
	if _, ok := TrimNoise(`meta Connection {{encoding "UTF-8"}}`); !ok {
		t.Fatal("meta must be recognized")
	}
}
