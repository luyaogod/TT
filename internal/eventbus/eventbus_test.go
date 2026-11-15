package eventbus

import "testing"

// TestBusPublishSince 增量读取与容量覆盖。
func TestBusPublishSince(t *testing.T) {
	b := New(2)
	if b.Latest() != 0 {
		t.Fatalf("初始 latest = %d", b.Latest())
	}
	b.Publish("a", nil)
	b.Publish("b", map[string]string{"k": "v"})
	if got := b.Since(0); len(got) != 2 || got[0].Name != "a" || got[1].Name != "b" {
		t.Fatalf("since0 = %+v", got)
	}
	if got := b.Since(1); len(got) != 1 || got[0].ID != 2 {
		t.Fatalf("since1 = %+v", got)
	}
	if string(b.Since(1)[0].Payload) != `{"k":"v"}` {
		t.Fatalf("payload 丢了")
	}
	b.Publish("c", nil)
	if got := b.Since(0); len(got) != 2 || got[0].ID != 2 || got[1].ID != 3 {
		t.Fatalf("环覆盖后 = %+v", got)
	}
	if b.Latest() != 3 {
		t.Fatalf("latest = %d", b.Latest())
	}
}
