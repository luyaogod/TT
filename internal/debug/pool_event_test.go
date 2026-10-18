package debug

// 池与界面会话的事件（W2）。
//
// 这一组钉三件事：
//  1. **登记/移除各发一条**，形状带得走（`added` 带整份 `brief`，`removed` 带 `reason`）；
//  2. **顺序**：`added` 一定在会话已登记之后、`removed` 一定在删除之后 —— 收到的人会立刻
//     按 id 去 GET，顺序反了就会出现"列表里没有、快照却拿得到"的窗口；
//  3. **不发**没发生的事：建会话失败（拨号挂了）不许发 `added`。

import (
	"fmt"
	"testing"
	"time"
)

// poolEventManager 造一个**不拨号**的 Manager：`newSession` 换成直接返回空闲宿主的假实现。
//
// 为什么要注入：真路径要拨 SSH（`host.NewSession` 里没有可注入的缝）。本组测的是
// "登记/移除之后发了什么事件"，不是拨号本身。
func poolEventManager(t *testing.T) (*Manager, <-chan Event) {
	t.Helper()
	m := NewManager(cfgFor("h1", "36"))
	var n int
	m.newSession = func(cfg *Config, _, _, _, _, _ string, _ func(Event)) (*Session, error) {
		n++
		return idleOn(fmt.Sprintf("s-fake-%d", n), cfg), nil
	}
	ch, cancel := m.SubscribeScoped("pool-event-test", "")
	t.Cleanup(cancel)
	return m, ch
}

// nextEvent 等一条指定类型的事件（默认 2 秒超时）。等不到就把已经收到的打出来。
func nextEvent(t *testing.T, ch <-chan Event, want string) Event {
	t.Helper()
	deadline := time.After(2 * time.Second)
	var seen []string
	for {
		select {
		case ev := <-ch:
			seen = append(seen, ev.Type)
			if ev.Type == want {
				return ev
			}
		case <-deadline:
			t.Fatalf("2 秒内没等到 %q 事件（收到过：%v）", want, seen)
		}
	}
}

// 建会话成功 → `session.added`，带得走整份 brief（客户端收到就能入索引）。
func TestSessionAddedCarriesBrief(t *testing.T) {
	m, ch := poolEventManager(t)

	sess, err := m.CreateSessionOn(cfgFor("h1", "36"), "录像A", PurposeRecord)
	if err != nil {
		t.Fatalf("建会话失败：%v", err)
	}

	ev := nextEvent(t, ch, EvSessionAdded)
	if ev.SessionID != sess.ID {
		t.Errorf("added 的 sessionId = %q，该是新建那条的句柄 %q", ev.SessionID, sess.ID)
	}
	if ev.Brief == nil {
		t.Fatal("added 必须带 brief —— 否则客户端还得再 GET 一次，多一个“列表还没刷到”的窗口")
	}
	if ev.Brief.ID != sess.ID || ev.Brief.Ref == "" {
		t.Errorf("brief 的形状不对：id=%q ref=%q", ev.Brief.ID, ev.Brief.Ref)
	}
	if ev.Brief.Purpose != PurposeRecord {
		t.Errorf("brief.purpose = %q，该是建会话时点的 %q（W1 + W2 要能对上）", ev.Brief.Purpose, PurposeRecord)
	}
	if ev.Seq == 0 {
		t.Error("事件必须带 seq（游标语义）")
	}
}

// 移除 → `session.removed`，带原因；同一条先 added 后 removed。
func TestSessionRemovedCarriesReasonAndFollowsAdded(t *testing.T) {
	m, ch := poolEventManager(t)

	sess, err := m.CreateSessionOn(cfgFor("h1", "36"), "", PurposeDebug)
	if err != nil {
		t.Fatalf("建会话失败：%v", err)
	}
	added := nextEvent(t, ch, EvSessionAdded)

	m.RemoveAs(sess.ID, "close")
	removed := nextEvent(t, ch, EvSessionRemoved)

	if removed.SessionID != sess.ID {
		t.Errorf("removed 的 sessionId = %q，该是 %q", removed.SessionID, sess.ID)
	}
	if removed.Reason != "close" {
		t.Errorf("removed 的 reason = %q，该是 close（W2 的取值：close|exit|restart|switch）", removed.Reason)
	}
	if removed.Seq <= added.Seq {
		t.Errorf("removed 的 seq(%d) 必须大于 added 的 seq(%d)（同一条流上的顺序）", removed.Seq, added.Seq)
	}
	// 移除之后池里确实没有它 —— 事件不能比事实早，也不能比事实晚。
	if m.Get(sess.ID) != nil {
		t.Error("发了 removed 之后池里还剩着这条会话")
	}
}

// 建会话失败（拨号挂了）→ **不发** added。发了的话前端会先加一条、下一轮 List 再删一条。
func TestNoSessionAddedWhenCreateFails(t *testing.T) {
	m := NewManager(cfgFor("h1", "36"))
	m.newSession = func(*Config, string, string, string, string, string, func(Event)) (*Session, error) {
		return nil, fmt.Errorf("测试：拨号失败")
	}
	ch, cancel := m.SubscribeScoped("pool-event-test", "")
	defer cancel()

	if _, err := m.CreateSessionOn(cfgFor("h1", "36"), "", PurposeDebug); err == nil {
		t.Fatal("假实现必然失败，这里不该成功")
	}
	select {
	case ev := <-ch:
		t.Fatalf("建会话失败时不该发任何事件，实得 %q", ev.Type)
	case <-time.After(150 * time.Millisecond):
		// 好：一条都没有
	}
}

// 移除一条**不在池里**的会话：不发事件（不能凭空造一条"少了一条"）。
func TestRemoveUnknownEmitsNothing(t *testing.T) {
	m, ch := poolEventManager(t)
	m.RemoveAs("s-没有这条", "close")
	select {
	case ev := <-ch:
		t.Fatalf("移除不存在的会话不该发事件，实得 %q", ev.Type)
	case <-time.After(150 * time.Millisecond):
	}
}

// 界面会话的挂接变化：attached=true → screen.attached；false → screen.detached + reason。
func TestPublishScreenChange(t *testing.T) {
	m, ch := poolEventManager(t)

	m.PublishScreenChange(true, ScreenState{
		Ref: "主机109#1", Tag: "fep-20261018-101401", Kind: "rec",
		GDCPid: 27300, ProxyPort: 7600, InjectPort: 7899, OwnerPID: 23308,
	})
	at := nextEvent(t, ch, EvScreenAttached)
	if at.Screen == nil || at.Screen.Tag != "fep-20261018-101401" || at.Screen.ProxyPort != 7600 {
		t.Fatalf("screen.attached 的 screen 块不对：%+v", at.Screen)
	}
	if at.Reason != "" {
		t.Errorf("attached 不该带 reason，实得 %q", at.Reason)
	}

	m.PublishScreenChange(false, ScreenState{Ref: "主机109#1", Tag: "fep-20261018-101401", Reason: "stop"})
	de := nextEvent(t, ch, EvScreenDetached)
	if de.Screen == nil || de.Screen.Reason != "stop" {
		t.Fatalf("screen.detached 该带 reason=stop：%+v", de.Screen)
	}
}

// ref 为空的挂接变化不上事件（没有目标可指向，发了只会让客户端猜）。
func TestPublishScreenChangeWithoutRefIsIgnored(t *testing.T) {
	m, ch := poolEventManager(t)
	m.PublishScreenChange(true, ScreenState{Tag: "x"})
	select {
	case ev := <-ch:
		t.Fatalf("ref 为空时不该发事件，实得 %q", ev.Type)
	case <-time.After(150 * time.Millisecond):
	}
}
