package fep

import (
	"errors"
	"strings"
	"testing"
)

func testModel(t *testing.T) *Model {
	t.Helper()
	m := NewModel()
	// 合成夹具：一棵 Window 下的动作/字段/容器，形状与真实格式一致（无客户数据）
	m.ApplyMessage(`om 1 {{an 0 Window 4097 {{name "w"}} {{` +
		`{an 4097 ToolBarItem 5160 {{text "录入"}{actionActive "1"}} {}} ` +
		`{an 4097 Label 900 {{text "静态文本"}} {}} ` +
		`{an 4097 ButtonEdit 272 {{action "controlp"}} {}} ` +
		`{an 4097 ButtonEdit 273 {{name "x.y"}} {}} ` +
		`{an 4097 ButtonEdit 300 {{action "controlp"}} {{an 300 FormField 301 {{value ""}} {}}} {}} ` +
		`{an 4097 ToolBarItem 5156 {{text "确定"}{actionActive "0"}} {}} ` +
		`{an 4097 Value 5197 {{value ""}} {}}}}}`)
	if _, ok := m.Get(4097); !ok {
		t.Fatal("夹具没建起来")
	}
	return m
}

// 构造器必须与录制件里客户端自己发出的字节**逐字一致**。
func TestEventBuildersMatchRecordedBytes(t *testing.T) {
	cases := []struct{ got, want string }{
		{ActionEventText(5160), `ActionEvent 0{{idRef "5160"}}`},
		{KeyEventText("Tab"), `KeyEvent 0{{keyName "Tab"}}`},
		{ConfigureValueText(60, "LUYAO", nil), `ConfigureEvent 0{{idRef "60"}{ value "LUYAO"}}`},
		{ConfigureCursorText(6543, 5), `ConfigureEvent 0{{idRef "6543"}{ cursor "5"}{ cursor2 "5"}}`},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Fatalf("构造:\n got %s\nwant %s", c.got, c.want)
		}
	}
	caret := 7
	if got, want := ConfigureValueText(1, "AB", &caret), `ConfigureEvent 0{{idRef "1"}{ value "AB"}{ cursor "7"}{ cursor2 "7"}}`; got != want {
		t.Fatalf("带光标:\n got %s\nwant %s", got, want)
	}
	if got, want := strings.TrimRight(string(EncodeEvent(5, ActionEventText(5160))), "\n"), `event _om 5{}{{ActionEvent 0{{idRef "5160"}}}}`; got[9:] != want {
		t.Fatalf("包帧后的载荷 = %q", got[9:])
	}
}

// 计划 §3.8：6 类非法事件必须全被拦，且**不发包**（这里只要求"被拒"）。
func TestPreflightRefusesSixIllegalShapes(t *testing.T) {
	m := testModel(t)
	const seq = 9

	// 1 目标不存在
	if err := CheckClick(m, 12345, seq, seq); !refusedWith(err, 5) {
		t.Fatalf("1 目标不存在：%v", err)
	}
	// 2 类型不在白名单
	if err := CheckClick(m, 900, seq, seq); !refusedWith(err, 5) {
		t.Fatalf("2 类型不在白名单：%v", err)
	}
	// 3 ButtonEdit 没有 controlp（不是开窗按钮）
	if err := CheckClick(m, 273, seq, seq); !refusedWith(err, 5) {
		t.Fatalf("3 非开窗 ButtonEdit：%v", err)
	}
	// 4 actionActive 为假
	if err := CheckClick(m, 5156, seq, seq); !refusedWith(err, 4) {
		t.Fatalf("4 actionActive 为假：%v", err)
	}
	// 5 序号不是该连接的下一个
	if err := CheckClick(m, 5160, seq+1, seq); !refusedWith(err, 4) {
		t.Fatalf("5 序号不对：%v", err)
	}
	// 6 KeyEvent 带 idRef（回放时可能喂进来）
	if err := CheckRaw(`KeyEvent 0{{idRef "5160"}{keyName "Tab"}}`); !refusedWith(err, 2) {
		t.Fatalf("6 KeyEvent 带 idRef：%v", err)
	}

	// 合法事件必须放行（否则上面全是"永远拒绝"式的假通过）
	if err := CheckClick(m, 5160, seq, seq); err != nil {
		t.Fatalf("合法点击被拒：%v", err)
	}
	if err := CheckClick(m, 272, seq, seq); err != nil {
		t.Fatalf("合法开窗按钮被拒：%v", err)
	}
	if err := CheckClick(m, 5197, seq, seq); err != nil {
		t.Fatalf("合法 Value 点击被拒（录制件里它确实被点过）：%v", err)
	}
	if err := CheckKey("Tab"); err != nil {
		t.Fatalf("合法按键被拒：%v", err)
	}
}

// 打字该发给谁：优先字段的 FormField 子对象。
func TestResolveInputTargetPrefersFormField(t *testing.T) {
	m := testModel(t)
	target, via, err := ResolveInputTarget(m, 300)
	if err != nil || !via || target != 301 {
		t.Fatalf("FormField 优先：target=%d via=%v err=%v", target, via, err)
	}
	target, via, err = ResolveInputTarget(m, 272)
	if err != nil || via || target != 272 {
		t.Fatalf("没有 FormField 时用它自己并如实报告：target=%d via=%v err=%v", target, via, err)
	}
	if _, _, err := ResolveInputTarget(m, 99999); !refusedWith(err, 5) {
		t.Fatalf("不存在的字段：%v", err)
	}
}

func refusedWith(err error, exit int) bool {
	var r *Refusal
	return errors.As(err, &r) && r.Exit == exit
}
