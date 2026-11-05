package debug

// annotateEnvFailure 的措辞约束（2026-11-05 用户反馈「报错提示不应该那么绝对」）：
//
//  1. **只陈述事实与可选方向，不替调用方判定原因**，也不给可能性排序；
//  2. 环境名一律用**用户自己配置/传入的**那个（--at / --env 原值），
//     没传就如实说"未指定"，不许编造"当前环境"这种没信息量的说法
//     （没传时落点由服务端当前配置决定，CLI 本地算不出来）；
//  3. 认不出来的失败不给建议（宁缺毋滥）。

import (
	"strings"
	"testing"

	"tt/internal/debugclient"
	"tt/internal/output"
)

func TestQueryFailureTipRowidCrossEnv(t *testing.T) {
	tip := queryFailureTip("测试", "wsfa_t(wsfa_t) 查询出错: ORA-01410: invalid ROWID")

	if !strings.Contains(tip, "物理地址") {
		t.Fatalf("应说明 rowid 是物理地址，得到: %s", tip)
	}
	if !strings.Contains(tip, "测试") {
		t.Fatalf("环境名应取用户传入的那个，得到: %s", tip)
	}
	// 不许出现绝对化措辞：这些说法等于替用户宣布了结论。
	for _, abs := range []string{"必然失败", "按此顺序排", "核对目标环境"} {
		if strings.Contains(tip, abs) {
			t.Fatalf("措辞过于绝对（%q），得到: %s", abs, tip)
		}
	}
}

func TestQueryFailureTipTableMissing(t *testing.T) {
	tip := queryFailureTip("测试", "wsfa_t 查询出错: ORA-00942: table or view does not exist")
	if !strings.Contains(tip, "at-list") {
		t.Fatalf("应给可复制的下一步，得到: %s", tip)
	}
	if !strings.Contains(tip, "测试") {
		t.Fatalf("应带上用户传入的环境名，得到: %s", tip)
	}
}

func TestQueryFailureTipUnknownGivesNothing(t *testing.T) {
	if tip := queryFailureTip("测试", "some unrelated network error"); tip != "" {
		t.Fatalf("未知失败不该编提示，得到: %s", tip)
	}
}

// 没传 --at 时：不编造"当前环境"，而是如实说"未指定、落点由服务端决定"。
func TestQueryFailureTipNoEnvSaysUnspecified(t *testing.T) {
	tip := queryFailureTip("", "wsfa_t 查询出错: ERROR at line 1:")

	if tip == "" {
		t.Fatal("旧服务下什么都不提示 = 又回到白跑一回合")
	}
	if strings.Contains(tip, "当前环境") {
		t.Fatalf("不许编造“当前环境”（CLI 算不出服务端落点），得到: %s", tip)
	}
	if !strings.Contains(tip, "未指定环境") {
		t.Fatalf("应如实说明未指定环境，得到: %s", tip)
	}
	if !strings.Contains(tip, "服务端") {
		t.Fatalf("应说明落点由服务端决定，得到: %s", tip)
	}
}

// 旧服务场景：信息不完整时要说"原因没拿到"，不许下结论。
func TestQueryFailureTipOldServeStatesMissingCause(t *testing.T) {
	tip := queryFailureTip("正式", "wsfa_t 查询出错: ERROR at line 1:")

	if !strings.Contains(tip, "原因本轮没拿到") {
		t.Fatalf("应坦白原因未知，得到: %s", tip)
	}
	if !strings.Contains(tip, "ORA 原文") {
		t.Fatalf("应说明缺的是 ORA 原文，得到: %s", tip)
	}
	for _, abs := range []string{"按此顺序排", "上查不到这条 rowid", "核对目标环境"} {
		if strings.Contains(tip, abs) {
			t.Fatalf("不许下绝对结论（%q），得到: %s", abs, tip)
		}
	}
}

func TestAnnotateEnvFailureKeepsCodeAndExit(t *testing.T) {
	f := &debugclient.Failure{Code: output.CodeQueryFailed, Message: "ORA-01410: invalid ROWID"}
	got := annotateEnvFailure(f)
	if got != f {
		t.Fatal("应原地补 hint 并返回同一个 Failure")
	}
	if f.Code != output.CodeQueryFailed {
		t.Fatalf("不许改码，得到 %s", f.Code)
	}
	if f.Hint == "" {
		t.Fatal("rowid 跨库这种情况应补上 hint")
	}
}

func TestAnnotateEnvFailureIgnoresOtherCodes(t *testing.T) {
	f := &debugclient.Failure{Code: output.CodeNotFound, Message: "ORA-01410: invalid ROWID"}
	annotateEnvFailure(f)
	if f.Hint != "" {
		t.Fatalf("只认 QUERY_FAILED，其他码不该被加提示，得到: %s", f.Hint)
	}
}

// 环境名是**用户定义的、非固定的**：提示里出现的必须是用户传进来的那一个字符串，
// 且不许出现任何内置的具体环境名。用户取什么名字，就得原样显示什么名字。
func TestQueryFailureTipUsesExactlyUserEnvName(t *testing.T) {
	for _, env := range []string{"测试", "正式", "恒烁正式区", "prod-eu-1", "客户A_生产"} {
		tip := queryFailureTip(env, "wsfa_t 查询出错: ERROR at line 1:")
		if !strings.Contains(tip, env) {
			t.Fatalf("环境 %q 应原样出现在提示里，得到: %s", env, tip)
		}
	}
}

// 反向约束：源码里不许把某个具体环境名写进提示逻辑。
// 这里在所有失败形态上跑一遍，确保输出里除了"用户传进来的名字"之外没有别的环境名。
func TestQueryFailureTipHasNoBuiltinEnvNames(t *testing.T) {
	builtin := []string{"恒烁", "测试库", "正式库"}
	msgs := []string{
		"wsfa_t 查询出错: ERROR at line 1:",
		"ORA-01410: invalid ROWID",
		"ORA-00942: table or view does not exist",
	}
	for _, msg := range msgs {
		for _, env := range []string{"", "我的环境"} {
			tip := queryFailureTip(env, msg)
			for _, b := range builtin {
				if strings.Contains(tip, b) {
					t.Fatalf("提示里出现了内置环境名 %q（环境名应只来自用户配置）: %s", b, tip)
				}
			}
		}
	}
}
