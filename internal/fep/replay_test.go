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
		"",
		`om 3 {{an 0 Window 1 {}}}`, // 运行期方向的消息不是前端事件
		`meta Connection {{encoding "UTF-8"}}`,
	}
	for _, ev := range deny {
		// 只挡“不是事件”的东西；ConfigureEvent 的各类属性一律放行
		//（回放求保真：前端当时发过的，就按当时的样子再发一遍）。
		if AllowedForReplay(ev) {
			t.Fatalf("不该回放: %s", ev)
		}
	}
}

func TestPlanReplaySkipsAndStops(t *testing.T) {
	events := []string{
		`ConfigureEvent 0{{idRef "1065"}{ tabIndex "1"}}`,
		`ActionEvent 0{{idRef "5160"}}`,
		`ActionEvent 0{{idRef "5170"}}`,
		`ActionEvent 0{{idRef "999999"}}`, // 界面里没有 —— 现在由“发之前那一刻”拦，计划阶段不管
		`KeyEvent 0{{keyName "Tab"}}`,
	}
	// 计划阶段只做白名单与 --skip／--stop-at；AUI 判断在 RunReplay 里逐条做。
	// （早前在计划阶段就拦，会把“后面才出现的按钮”提前剔掉 —— 实测保存/审核 4 步没发出去。）
	plan := PlanReplay(events, ReplayOptions{Skip: []string{"5170"}})
	if len(plan.Events) != 4 {
		t.Fatalf("应剩 4 条（除掉被 --skip 的那条）：%+v", plan.Events)
	}
	if len(plan.Skips) != 1 || !strings.Contains(plan.Skips[0].Reason, "跳过") {
		t.Fatalf("只该跳 1 条且写清原因：%+v", plan.Skips)
	}
	if got := PlanReplay(events, ReplayOptions{StopAt: 1}); len(got.Events) != 1 {
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
