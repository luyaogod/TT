package fep

import "testing"

// TestPlanReplayUnwrapsWireEnvelope 钉住「录制件存线上形态、白名单认事件体」这条缝。
//
// 真机 2026-10-18 实测的 bug：录制件里每条事件都是 `event _om <n>{}{{<事件体>}}`
// （`WrapEvent` 的产物），而 `AllowedForReplay` 认的是**事件体**前缀。
// 少了剥信封这一步，一份含 23 条 `ActionEvent` 的录制**进计划 0 条**，
// 报告还写着"不在回放白名单"——看起来像"这一轮本来就没动作"，静默误判。
func TestPlanReplayUnwrapsWireEnvelope(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string // 进计划的那条事件体；空 = 应当被拒
	}{
		{
			name: "ActionEvent（线上形态）",
			raw:  `event _om 0{}{{ActionEvent 0{{idRef "4408"}}}}`,
			want: `ActionEvent 0{{idRef "4408"}}`,
		},
		{
			name: "KeyEvent（线上形态）",
			raw:  `event _om 12{}{{KeyEvent 0{{idRef "5348"}}{{keyName "Tab"}}}}`,
			want: `KeyEvent 0{{idRef "5348"}}{{keyName "Tab"}}`,
		},
		{
			name: "已经是事件体（不许动）",
			raw:  `ActionEvent 0{{idRef "4408"}}`,
			want: `ActionEvent 0{{idRef "4408"}}`,
		},
		{
			name: "ConfigureEvent（线上形态，白名单内）",
			raw:  `event _om 3{}{{ConfigureEvent 0{{idRef "960"}{ hidden "2"}}}}`,
			want: `ConfigureEvent 0{{idRef "960"}{ hidden "2"}}`,
		},
		{
			name: "信封残缺（不认识就别猜，照原样交给白名单去拒）",
			raw:  `event _om 3{} ActionEvent 0{{idRef "1"}}`,
			want: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plan := PlanReplay([]string{c.raw}, ReplayOptions{})
			if c.want == "" {
				if len(plan.Events) != 0 {
					t.Fatalf("该被拒，却进了计划：%+v", plan.Events)
				}
				return
			}
			if len(plan.Events) != 1 {
				t.Fatalf("该进计划 1 条，实得 %d（skips=%v）", len(plan.Events), plan.Skips)
			}
			if got := plan.Events[0].Event; got != c.want {
				t.Fatalf("事件体没还原对：\n  得到 %q\n  期望 %q", got, c.want)
			}
			// 发送端 Client.Send 收的就是这个形态（它自己拼 `@<conn> `），
			// 所以这里**不许**残留 `event _om` 信封。
			if len(plan.Events[0].Event) >= 9 && plan.Events[0].Event[:9] == "event _om" {
				t.Fatalf("事件体里还带着信封：%q", plan.Events[0].Event)
			}
		})
	}
}
