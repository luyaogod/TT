package debug

// 数据库报错必须把**原因行**带出来（2026-11-05 实测的现场）。
//
// 从前 `wsfa_t 查询出错: ERROR at line 1:` 把紧随其后的 `ORA-00942: 表或视图不存在`
// 整行丢掉；调用方唯一的动作是"换个环境再试"，白跑一整个回合。

import (
	"strings"
	"testing"
)

func TestCollectDBErrorLinesKeepsCause(t *testing.T) {
	raw := "ERROR at line 1:\nORA-00942: table or view does not exist\n\n"
	got := collectDBErrorLines(raw, 6)
	if !strings.Contains(got, "ERROR at line 1") {
		t.Fatalf("丢了报错行: %q", got)
	}
	if !strings.Contains(got, "ORA-00942") {
		t.Fatalf("丢了原因行(ORA-00942 才是根因): %q", got)
	}
}

func TestCollectDBErrorLinesStopsAfterCause(t *testing.T) {
	raw := "ERROR at line 1:\nORA-00942: table or view does not exist\nsome log noise\n"
	got := collectDBErrorLines(raw, 6)
	if strings.Contains(got, "log noise") {
		t.Fatalf("原因行之后的日志噪声不该收进来: %q", got)
	}
}

// 真机现场：报错行只有一个行号，原因在下一行。错误里必须有 ORA 原文与读法提示。
func TestDBQueryErrorCarriesORAAndNextStep(t *testing.T) {
	err := dbQueryError("ds", "ERROR at line 1:\nORA-00942: table or view does not exist\n")
	msg := err.Error()

	if !strings.Contains(msg, "ORA-00942") {
		t.Fatalf("错误里没有 ORA 原文: %s", msg)
	}
	if !strings.Contains(msg, "ds") {
		t.Fatalf("错误里没说查的是哪个账号: %s", msg)
	}
	// 不许再是"只给一句行号"。
	if strings.HasSuffix(strings.TrimSpace(msg), "ERROR at line 1:") {
		t.Fatalf("这又变回了一句话的无效错误: %s", msg)
	}
}

// 报错行里**没有** ORA 号时，要明说"原因这一行没给"，别让调用方以为这就是全部。
func TestDBQueryErrorExplainsMissingORA(t *testing.T) {
	msg := dbQueryError("", "ERROR at line 1:").Error()
	if !strings.Contains(msg, "没带 ORA") {
		t.Fatalf("没有 ORA 号时应明确说明，得到: %s", msg)
	}
}

// 表名只有一处声明：探针与错误文案不许各写一份。
func TestWSLogTableConstant(t *testing.T) {
	if wsfaTable != "wsfa_t" {
		t.Fatalf("wsfaTable 常量被改了: %q", wsfaTable)
	}
}
