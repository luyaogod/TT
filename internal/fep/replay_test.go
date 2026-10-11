package fep

import (
	"net"
	"strings"
	"testing"
	"time"
)

// 白名单的边界是实测出来的：currentRow/currentColumn 必须回放（它就是"选中哪一行"），
// tabIndex/bufferSize/pageSize 必须丢掉（纯状态噪声）。
func TestReplayWhitelistBoundary(t *testing.T) {
	allow := []string{
		`ActionEvent 0{{idRef "5160"}}`,
		`KeyEvent 0{{keyName "Tab"}}`,
		`ConfigureEvent 0{{idRef "260"}{ value "TEST01"}}`,
		`ConfigureEvent 0{{idRef "1065"}{ currentRow "3"}}`,
		`ConfigureEvent 0{{idRef "1065"}{ currentColumn "2"}}`,
	}
	for _, ev := range allow {
		if !AllowedForReplay(ev) {
			t.Fatalf("应回放却没有: %s", ev)
		}
	}
	deny := []string{
		`ConfigureEvent 0{{idRef "1065"}{ tabIndex "260"}}`,
		`ConfigureEvent 0{{idRef "797"}{ bufferSize "10"}{ pageSize "9"}}`,
		`ConfigureEvent 0{{idRef "260"}{ cursor "0"}}`,
		"",
	}
	for _, ev := range deny {
		if AllowedForReplay(ev) {
			t.Fatalf("不该回放: %s", ev)
		}
	}
}

func TestPlanReplaySkipsAndStops(t *testing.T) {
	events := []string{
		`ConfigureEvent 0{{idRef "1065"}{ tabIndex "1"}}`, // 噪声
		`ActionEvent 0{{idRef "5160"}}`,
		`ActionEvent 0{{idRef "5170"}}`,
		`ActionEvent 0{{idRef "999999"}}`, // 界面里没有
		`KeyEvent 0{{keyName "Tab"}}`,
	}
	m := NewModel()
	m.ApplyMessage(`om 1 {{an 0 Window 4097 {{name "w"}} {{an 4097 ToolBarItem 5160 {{text "录入"}} {}} {an 4097 ToolBarItem 5170 {{text "查询"}} {}}}}}`)
	plan := PlanReplay(events, ReplayOptions{Model: m, Skip: []string{"5170"}})
	if len(plan.Events) != 2 {
		t.Fatalf("应剩 2 条（5160 + Tab）：%+v", plan.Events)
	}
	if len(plan.Skips) != 3 {
		t.Fatalf("应跳 3 条：%+v", plan.Skips)
	}
	if plan.Skips[1].Reason == "" || !strings.Contains(plan.Skips[2].Reason, "AUI 安全闸") {
		t.Fatalf("跳过原因要写清楚：%+v", plan.Skips)
	}
	if got := PlanReplay(events, ReplayOptions{Model: m, StopAt: 1}); len(got.Events) != 1 {
		t.Fatalf("StopAt 没生效: %+v", got.Events)
	}
}

func TestRunReplayEndToEnd(t *testing.T) {
	c, _, rt, _, _ := sessionFixture(t)
	events := []string{
		`ActionEvent 0{{idRef "5160"}}`,
		`ActionEvent 0{{idRef "999999"}}`, // 界面里没有 → 运行时安全闸应跳过
		`KeyEvent 0{{keyName "Tab"}}`,
	}
	plan := PlanReplay(events, ReplayOptions{})
	if len(plan.Events) != 3 {
		t.Fatalf("计划条数 = %d", len(plan.Events))
	}
	res, err := c.RunReplay(plan, 0, 10*time.Millisecond, false, nil)
	if err != nil {
		t.Fatalf("RunReplay: %v", err)
	}
	if res.Sent != 2 || res.Skipped != 1 || res.Failed != 0 {
		t.Fatalf("结果 = %+v", res)
	}
	// 真发出去的必须是那两条
	buf := make([]byte, 4096)
	rt.SetReadDeadline(time.Now().Add(2 * time.Second))
	seen := 0
	for seen < 2 {
		n, rerr := rt.Read(buf)
		if n > 0 {
			seen += strings.Count(string(buf[:n]), "event _om")
		}
		if rerr != nil {
			if ne, ok := rerr.(net.Error); ok && ne.Timeout() {
				break
			}
			t.Fatal(rerr)
		}
	}
	if seen != 2 {
		t.Fatalf("运行期只收到 %d 条", seen)
	}
	// dry-run 一个字节都不发
	before := rt
	_ = before
	res2, err := c.RunReplay(plan, 0, 0, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Sent != 0 {
		t.Fatalf("dry-run 不该发: %+v", res2)
	}
}
