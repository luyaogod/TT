package fep

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"tt/internal/fep"
	"tt/internal/output"
)

// `tt fep do ... --field <colName>` 的接线（契约 C3 + A5.3 + A18）。
//
// 纪律：这一组**不发任何事件** —— 被测的是"解析 → 决定发不发"那一层（纯函数），
// 以及互斥判定（用法错）。真发包与否由 `fep.Do` 的既有发前校验说了算，
// 那边已有自己的测试，这里不越界。

// cliFieldAUI 是合成读侧夹具（不回放、不连真机）：一张表单上的两类字段，
// 外加一个"同 colName 两种 fieldId"的真歧义与一个只当明细列的字段。
const cliFieldAUI = `om 1 {{an 0 Window 4097 {{name "w_axmt500"}} {{` +
	`{an 4097 FormField 6000 {{name "formonly.b_xmda001"}{colName "b_xmda001"}{fieldId "18"}{text "品号"}} {}} ` +
	`{an 4097 FormField 6002 {{name "formonly.b_amb"}{colName "b_amb"}{fieldId "6"}{text "歧义甲"}} {}} ` +
	`{an 4097 FormField 6004 {{name "formonly.b_amb"}{colName "b_amb"}{fieldId "7"}{text "歧义乙"}} {}} ` +
	`{an 4097 TableColumn 6006 {{name "formonly.b_col"}{colName "b_col"}{fieldId "9"}{text "明细列"}} {}} ` +
	`}}}}`

func cliFieldModel(t *testing.T) *fep.Model {
	t.Helper()
	m := fep.NewModel()
	m.ApplyMessage(cliFieldAUI)
	if _, ok := m.Get(6000); !ok {
		t.Fatalf("合成夹具没进模型（ApplyMessage 的形态不对）：len=%d", m.Len())
	}
	return m
}

// exitOf 取一条命令错的退出码（`internal/cli/fep` 的码表）。
func exitOf(t *testing.T, err error) int {
	t.Helper()
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("错误必须是 *output.Error（带退出码），得到 %T: %v", err, err)
	}
	return oe.Exit
}

// TestFieldAndTargetAreMutuallyExclusive：**要目标的动作**里，`--field` 与位置参数
// **互斥** ⇒ 用法错（2）。kind 是必填参数（契约 A28）：判据只在 needsTarget 一处。
func TestFieldAndTargetAreMutuallyExclusive(t *testing.T) {
	if err := checkFieldExclusive("click", "品号", "b_xmda001"); err == nil {
		t.Fatal("同时给名字与 --field 必须报用法错，不能自己挑一个")
	} else if got := exitOf(t, err); got != exitUsage {
		t.Fatalf("互斥的退出码 = %d，想要 %d（码表见 fep.go:25-33）", got, exitUsage)
	}
	if err := checkFieldExclusive("input", "", "b_xmda001"); err != nil {
		t.Fatalf("只给 --field 是合法用法：%v", err)
	}
	if err := checkFieldExclusive("click", "品号", ""); err != nil {
		t.Fatalf("只给名字是合法用法：%v", err)
	}
	if err := checkFieldExclusive("click", "  ", "\t"); err == nil {
		t.Fatal("两个都没给也必须报用法错（别把空串当目标去点空气）")
	} else if got := exitOf(t, err); got != exitUsage {
		t.Fatalf("两个都没给的退出码 = %d，想要 %d", got, exitUsage)
	}
}

// TestFieldKindForVerb：`kind` 用动作类型对应值（`act` → click，`type`/`pick` → input）。
func TestFieldKindForVerb(t *testing.T) {
	cases := map[string]string{"click": "click", "input": "input", "pick": "input", "key": ""}
	for kind, want := range cases {
		if got := fieldKindFor(kind); got != want {
			t.Errorf("kind=%s → %q，想要 %q", kind, got, want)
		}
	}
}

// TestResolveFieldOrRefuse：纯函数层 —— 命中就给 `#id`，不命中就给四码。
func TestResolveFieldOrRefuse(t *testing.T) {
	m := cliFieldModel(t)

	tgt, err := resolveFieldOrRefuse(m, fep.Scope{}, "b_xmda001", "input")
	if err != nil {
		t.Fatalf("唯一命中却被拒：%v", err)
	}
	if tgt.Target != "#6000" || tgt.ID != 6000 {
		t.Fatalf("解析结果 = %+v，想要 #6000", tgt)
	}
	if tgt.ColName != "b_xmda001" || tgt.FieldID != 18 {
		t.Fatalf("身份 = (%s, %d)，想要 (b_xmda001, 18)", tgt.ColName, tgt.FieldID)
	}

	cases := []struct {
		name string
		col  string
		kind string
		code string
		exit int
	}{
		{"作用域里没这个名字", "b_nope", "input", fep.CodeFieldNotInScope, fep.ExitFieldNotInScope},
		{"同一名字两种编号", "b_amb", "input", fep.CodeFieldAmbiguous, fep.ExitFieldAmbiguous},
		{"明细列不接受点击", "b_col", "click", fep.CodeFieldNotActionable, fep.ExitFieldNotActionable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := resolveFieldOrRefuse(m, fep.Scope{}, c.col, c.kind)
			var ref *fep.Refusal
			if !errors.As(err, &ref) {
				t.Fatalf("该按四码拒绝（%s），得到 %v", c.code, err)
			}
			if ref.Code != c.code || ref.Exit != c.exit {
				t.Fatalf("码/退出码 = %s/%d，想要 %s/%d", ref.Code, ref.Exit, c.code, c.exit)
			}
			if ref.Hint == "" {
				t.Fatalf("拒绝必须给可操作 Hint（C4.3）：%+v", ref)
			}
		})
	}

	// 没有模型（这一屏取不到）→ FORM_UNKNOWN / 1，而不是"没这个字段"
	_, err = resolveFieldOrRefuse(nil, fep.Scope{}, "b_xmda001", "input")
	var ref *fep.Refusal
	if !errors.As(err, &ref) || ref.Code != fep.CodeFormUnknown || ref.Exit != fep.ExitFormUnknown {
		t.Fatalf("没有模型时 = %v，想要 FORM_UNKNOWN/1", err)
	}

	// 窗口收窄只**收窄**（A18）：窗口名不符 ⇒ 拒，而报错里要点出它到底在哪个窗口
	_, err = resolveFieldOrRefuse(m, fep.Scope{Window: "别的窗口"}, "b_xmda001", "input")
	if !errors.As(err, &ref) || ref.Code != fep.CodeFieldNotInScope {
		t.Fatalf("窗口收窄没生效：%v", err)
	}
}

// TestDoVerbArgsAllowFieldWithoutTarget：`--field` 时位置参数该少一个，
// 而且**照着码表**报用法错（2），不是 cobra 默认的 1。
func TestDoVerbArgsAllowFieldWithoutTarget(t *testing.T) {
	old := doField
	defer func() { doField = old }()

	doField = "b_xmda001"
	if err := targetOrFieldArgs(doActCmd, nil); err != nil {
		t.Fatalf("act --field 不带位置参数是合法用法：%v", err)
	}
	if err := targetOrFieldArgs(doActCmd, []string{"品号"}); err == nil {
		t.Fatal("act --field 再给位置参数必须报用法错")
	} else if got := exitOf(t, err); got != exitUsage {
		t.Fatalf("退出码 = %d，想要 %d", got, exitUsage)
	}
	if err := valueAfterFieldArgs(doTypeCmd, []string{"ABC"}); err != nil {
		t.Fatalf("type --field 之后只该有一个值参数：%v", err)
	}
	if err := valueAfterFieldArgs(doTypeCmd, nil); err == nil {
		t.Fatal("type --field 少了值参数必须报用法错")
	} else if got := exitOf(t, err); got != exitUsage {
		t.Fatalf("退出码 = %d，想要 %d", got, exitUsage)
	}

	doField = ""
	if err := targetOrFieldArgs(doActCmd, []string{"品号"}); err != nil {
		t.Fatalf("act 带名字是旧用法，不许变：%v", err)
	}
	if err := targetOrFieldArgs(doActCmd, nil); err == nil {
		t.Fatal("既没 --field 也没名字必须报用法错")
	} else if got := exitOf(t, err); got != exitUsage {
		t.Fatalf("退出码 = %d，想要 %d", got, exitUsage)
	}
	if err := valueAfterFieldArgs(doTypeCmd, []string{"品号", "ABC"}); err != nil {
		t.Fatalf("type 带名字+值是旧用法，不许变：%v", err)
	}
}

// TestFieldFlagIsRegisteredOnEveryDoVerb：`--field` 挂在三个动词上，但含义不同
// （契约 A28）：`act` / `type` 用它点名栏位；`key` 挂它**是为了明确拒绝**
// —— 不挂的话报错是 cobra 的 `unknown flag`，讲不清为什么不行。
func TestFieldFlagIsRegisteredOnEveryDoVerb(t *testing.T) {
	for _, c := range []*cobra.Command{doActCmd, doTypeCmd} {
		f := c.Flags().Lookup("field")
		if f == nil {
			t.Fatalf("%s 上没有 --field", c.Name())
		}
		if !strings.Contains(f.Usage, "colName") {
			t.Fatalf("%s 的 --field 说明没讲清它是 colName：%q", c.Name(), f.Usage)
		}
	}
	f := doKeyCmd.Flags().Lookup("field")
	if f == nil {
		t.Fatal("key 上也要有 --field：拒绝要讲清理由，而不是一句 unknown flag")
	}
	if !strings.Contains(f.Usage, "没有目标") {
		t.Fatalf("key 的 --field 说明要讲清「key 没有目标」：%q", f.Usage)
	}
	// 挂上 ≠ 能用：口径是**拒绝**（见 do_key_test.go 的 TestDoKeyRejectsFieldFlagExplicitly）
	if err := checkFieldExclusive("key", "", "b_xmda001"); err == nil {
		t.Fatal("key + --field 必须报用法错")
	}
}
