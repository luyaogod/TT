package fep

import "testing"

func TestSplitEvents(t *testing.T) {
	body := `event _om 7{}{{ActionEvent 0{{idRef "5160"}}}{KeyEvent 0{{keyName "Tab"}}}}`
	got := SplitEvents(body)
	want := []string{`ActionEvent 0{{idRef "5160"}}`, `KeyEvent 0{{keyName "Tab"}}`}
	if len(got) != len(want) {
		t.Fatalf("got %d events: %v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("event %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// The rule that fixed the manual: steps come from user intent, not from the
// navigation the user performed on the way.
func TestSemanticActionsDropNavigation(t *testing.T) {
	events := []string{
		`ActionEvent 0{{idRef "5160"}}`,                                         // 录入
		`KeyEvent 0{{keyName "Tab"}}`,                                           // 经过栏位
		`ConfigureEvent 0{{idRef "797"}{ cursor "0"}{ cursor2 "0"}}`,            // 光标
		`ConfigureEvent 0{{idRef "5195"}{ currentRow "4"}{ currentColumn "1"}}`, // 行列位置
		`ConfigureEvent 0{{idRef "3009"}{ value "1000"}}`,                       // 输入
		`ConfigureEvent 0{{idRef "797"}{ bufferSize "10"}{ pageSize "9"}}`,      // 尺寸
		`ActionEvent 0{{idRef "5156"}}`,                                         // 确定
	}
	got := SemanticActions(events)
	if len(got) != 3 {
		t.Fatalf("got %d actions, want 3: %+v", len(got), got)
	}
	if got[0].Kind != ActionClick || got[0].ID != 5160 {
		t.Errorf("action 0 = %+v", got[0])
	}
	if got[1].Kind != ActionInput || got[1].Value != "1000" || got[1].ID != 3009 {
		t.Errorf("action 1 = %+v", got[1])
	}
	if got[2].Kind != ActionClick || got[2].ID != 5156 {
		t.Errorf("action 2 = %+v", got[2])
	}
}

func TestAttrOfHandlesMissingKeys(t *testing.T) {
	if _, ok := attrOf(`ActionEvent 0{{idRef "12"}}`, "value"); ok {
		t.Error("value must not be found in an ActionEvent")
	}
	v, ok := attrOf(`ConfigureEvent 0{{idRef "12"}{ value "abc"}}`, "value")
	if !ok || v != "abc" {
		t.Fatalf("attrOf = %q ok=%v", v, ok)
	}
}
