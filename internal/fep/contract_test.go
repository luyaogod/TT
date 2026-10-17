package fep

import (
	"strings"
	"testing"

	"tt/internal/config"
)

// 牙 d：**一致性断言**（契约 §3："文档只记为什么；是什么由编译器 + 测试保证"）。
//
// 这个文件只做一种事：**有人改了口径而别处没跟上时，当场红**。
// 所以它不写业务行为（那些在 identity_test.go 里），只钉"唯一真源"：
//
//	1) 四码的字面值 = 契约 C4.2 的原文（含 A11 新增的第 4 码），退出码 = 5 / 5 / 1 / 4，
//	   且**码表就是真源**（拒绝是从码表取退出码的，不是各调用点各写一个数字）；
//	2) `forms` 必须在 `config.CacheSubdirs` 里 —— 表单缓存落点是
//	   `<数据目录>/forms`（契约 C2.3），而 B7 说清单外的名字拿空串 ⇒ 漏登记就是
//	   "取件那头静默写不进去"，只有这条断言能提前发现。

// TestRefusalCodesAndExits 钉四码与退出码（契约 C4.2 + A11）。
func TestRefusalCodesAndExits(t *testing.T) {
	// 1) 字面值 = 契约 C4.2 的原文（改名或改字面值，这里立刻红）。
	literals := []struct {
		name, got, want string
	}{
		{"CodeFieldNotInScope", CodeFieldNotInScope, "FIELD_NOT_IN_SCOPE"},
		{"CodeFieldAmbiguous", CodeFieldAmbiguous, "FIELD_AMBIGUOUS"},
		{"CodeFormUnknown", CodeFormUnknown, "FORM_UNKNOWN"},
		{"CodeFieldNotActionable", CodeFieldNotActionable, "FIELD_NOT_ACTIONABLE"},
	}
	for _, l := range literals {
		if l.got != l.want {
			t.Errorf("%s = %q, want %q（契约 C4.2 冻结的字面值）", l.name, l.got, l.want)
		}
	}

	// 2) 退出码 5 / 5 / 1 / 4（契约 C4.2 + A11；数字与 internal/cli/fep 的码表同档，见 B10）。
	exits := []struct {
		name string
		got  int
		want int
	}{
		{"ExitFieldNotInScope", ExitFieldNotInScope, 5},
		{"ExitFieldAmbiguous", ExitFieldAmbiguous, 5},
		{"ExitFormUnknown", ExitFormUnknown, 1},
		{"ExitFieldNotActionable", ExitFieldNotActionable, 4},
	}
	for _, e := range exits {
		if e.got != e.want {
			t.Errorf("%s = %d, want %d（契约 C4.2）", e.name, e.got, e.want)
		}
	}

	// 3) 拒绝里的退出码必须来自**那张码表**（免得同一个码在两个调用点给出两个数字）。
	for code, want := range map[string]int{
		CodeFieldNotInScope:    5,
		CodeFieldAmbiguous:     5,
		CodeFormUnknown:        1,
		CodeFieldNotActionable: 4,
	} {
		if got, ok := refusalExitCodes[code]; !ok || got != want {
			t.Errorf("码表 refusalExitCodes[%q] = %d/%v, want %d", code, got, ok, want)
		}
	}
	if len(refusalExitCodes) != 4 {
		t.Errorf("码表有 %d 条，want 4（C4.2 的四码）—— 加码要同步契约与这条断言", len(refusalExitCodes))
	}

	// 4) 行为面：四条路径上真的回这四个码 + 这四个退出码（常量不是摆设）。
	//    FORM_UNKNOWN：没有可用的 AUI 模型（这一屏的表单还取不到）。
	//    FIELD_NOT_IN_SCOPE：空模型里当然没有这个字段。
	//    FIELD_AMBIGUOUS：同一个 colName 挂了两种 fieldId。
	//    FIELD_NOT_ACTIONABLE：身份找到了，但那个节点不接受这个动作（A11）。
	m := identityModel(t)
	cases := []struct {
		name     string
		run      func() (Resolved, error)
		wantCode string
		wantExit int
	}{
		{"没有模型", func() (Resolved, error) {
			var nilModel *Model
			return ByColName(nilModel, Scope{}, "b_xmda001", "input")
		}, CodeFormUnknown, 1},
		{"空模型", func() (Resolved, error) {
			return ByColName(NewModel(), Scope{}, "b_xmda001", "input")
		}, CodeFieldNotInScope, 5},
		{"真歧义", func() (Resolved, error) {
			return ByColName(m, Scope{Window: testWinATitle}, "b_index", "input")
		}, CodeFieldAmbiguous, 5},
		{"明细列不接受该动作", func() (Resolved, error) {
			return ByColName(m, Scope{Window: testWinATitle}, "b_onlycol", "input")
		}, CodeFieldNotActionable, 4},
	}
	for _, c := range cases {
		_, err := c.run()
		if err == nil {
			t.Errorf("%s：本该被拒", c.name)
			continue
		}
		r, ok := err.(*Refusal)
		if !ok {
			t.Errorf("%s：失败必须是 *Refusal（C4.1），得到 %T", c.name, err)
			continue
		}
		if r.Code != c.wantCode || r.Exit != c.wantExit {
			t.Errorf("%s：得到 %s/%d, want %s/%d（%s）", c.name, r.Code, r.Exit, c.wantCode, c.wantExit, r.Message)
		}
		if r.Hint == "" {
			t.Errorf("%s：拒绝没有 Hint（C4.3 要求可操作）", c.name)
		}
	}
}

// TestFormsDirInCacheSubdirs 钉 C2.3：`forms` 必须在 config.CacheSubdirs 里。
//
// **这条红色是预期**（切片划分的故意设计）：清单里的那一行由**切片 B** 加
// （它拥有 internal/config/cache.go），切片 A 只负责断言存在。
// 这里**不去改 config**（越界）：红就是红，如实报给架构师。
func TestFormsDirInCacheSubdirs(t *testing.T) {
	found := false
	for _, n := range config.CacheSubdirs {
		if n == "forms" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("`forms` 不在 config.CacheSubdirs 里（现在是 %v）。\n"+
			"    契约 C2.3：表单缓存的落点是 <数据目录>/forms，必须登记进这份清单；\n"+
			"    B7 说清单外的名字 config.CacheDir 回空串 —— 漏登记就是取件那头静默写不进去。\n"+
			"    这一行由切片 B 加（internal/config/cache.go，恰好 +1 行）；本条红色是设计预期。",
			config.CacheSubdirs)
		return
	}

	// 登记了才有意义：这个名字必须真的能被 CacheDir 认下来（否则只是往清单里写了串字符）。
	if got := config.CacheDir("X:/tt-data", "forms"); got == "" || !strings.HasSuffix(got, "forms") {
		t.Errorf(`config.CacheDir("X:/tt-data", "forms") = %q：登记了却拼不出路径，取件那头还是写不进去`, got)
	}
}
