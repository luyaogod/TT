package debug

// annotateEnvFailure 第一版踩过的坑（2026-11-05）：把"rowid 跨库"误判成"环境没同步表"，
// 于是对**所有** QUERY_FAILED 无条件补一句"核对目标环境"。
//
// 这些用例锁住新口径：**先读错误在说什么，再决定提什么**；认不出来就不给建议。

import (
	"strings"
	"testing"

	"tt/internal/debugclient"
	"tt/internal/output"
)

func TestQueryFailureTipRowidCrossEnv(t *testing.T) {
	msg := "wsfa_t(wsfa_t) 查询出错: ORA-01410: invalid ROWID"
	tip := queryFailureTip("测试", msg)

	if !strings.Contains(tip, "只在") || !strings.Contains(tip, "产生它的环境") {
		t.Fatalf("rowid 跨库应提示 rowid 的环境归属，得到: %s", tip)
	}
	if !strings.Contains(tip, "测试") {
		t.Fatalf("应带上这次落到的环境，得到: %s", tip)
	}
	// 关键：这句话会把人引向"换环境再试"，而环境本来就是对的。
	if strings.Contains(tip, "核对目标环境") {
		t.Fatalf("rowid 跨库时不该说'核对目标环境'（这次环境是对的）: %s", tip)
	}
}

func TestQueryFailureTipTableMissing(t *testing.T) {
	tip := queryFailureTip("测试", "wsfa_t 查询出错: ORA-00942: table or view does not exist")
	if !strings.Contains(tip, "at-list") {
		t.Fatalf("表不在才是环境问题，应指 --at-list，得到: %s", tip)
	}
}

func TestQueryFailureTipUnknownGivesNothing(t *testing.T) {
	// 认不出来的失败不许乱给建议：宁缺毋滥。
	if tip := queryFailureTip("测试", "some unrelated network error"); tip != "" {
		t.Fatalf("未知失败不该编提示，得到: %s", tip)
	}
}

// 旧版服务只回 `ERROR at line 1:`（原因被吞）。这时不能装作认出来了，也不能什么都不说：
// 要给一条把两种可能都覆盖、且**顺序正确**的判据（rowid 跨界更常见也更便宜）。
func TestQueryFailureTipOldServeStillGuides(t *testing.T) {
	tip := queryFailureTip("测试", "wsfa_t 查询出错: ERROR at line 1:")
	if tip == "" {
		t.Fatal("旧服务下什么都不提示 = 又回到白跑一回合")
	}
	if !strings.Contains(tip, "只在产生它的环境有效") {
		t.Fatalf("应把 rowid 跨界列为第一嫌疑，得到: %s", tip)
	}
	// 不能一上来就"核对环境"——这次的环境是对的。
	if strings.Index(tip, "at-list") < strings.Index(tip, "产生它的环境") {
		t.Fatalf("应先把 rowid 归属排在环境之前，得到: %s", tip)
	}
	if !strings.Contains(tip, "测试") {
		t.Fatalf("应带上这次落到的环境，得到: %s", tip)
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
