package devdata

// CLI 层的检查项：错误码映射与两种输出形态的内容一致性。
//
// 领域逻辑的测试在 internal/devdata；这里只钉"用户看到的东西"：
//   - 领域错误 → 退出码（包装脚本与 AI 都按退出码分支）；
//   - --human 不丢内容（表格与 JSON 同源）。
//
// 输出走 output.Emit 到 os.Stdout，测试改为直接调 emit* 的构造逻辑不现实，
// 所以这里只测失败路径与格式换算（成功路径的真机验收在 P5，见 tasks 文件）。

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/devdata"
	"github.com/luyaogod/TT/internal/output"
)

// TestFailErrMapping 领域错误 → 稳定错误码 + 退出码。
func TestFailErrMapping(t *testing.T) {
	cases := []struct {
		err  error
		code string
		exit int
	}{
		{devdata.ErrEnvNotFound, "USAGE", exitUsage},
		{devdata.ErrInvalidEnv, "USAGE", exitUsage},
		{devdata.ErrInvalidProg, "USAGE", exitUsage},
		{devdata.ErrInvalidOpt, "USAGE", exitUsage},
		{devdata.ErrModuleUnknown, "USAGE", exitUsage},
		{devdata.ErrProgNotFound, "NOT_FOUND", exitNotFound},
		{devdata.ErrExists, "EXISTS", exitUsage},
		{context.Canceled, "CANCELLED", exitFailed},
		{errors.New("boom"), "FAILED", exitFailed},
	}
	for _, c := range cases {
		err := failErr(c.err)
		var oe *output.Error
		if !errors.As(err, &oe) {
			t.Errorf("%v 应映射成 output.Error，得到 %T", c.err, err)
			continue
		}
		if oe.Code != c.code || oe.Exit != c.exit {
			t.Errorf("%v → code=%s exit=%d，应为 %s/%d", c.err, oe.Code, oe.Exit, c.code, c.exit)
		}
	}
}

// TestUsageOrFailMapping status/ls 的只读性质："没拉过"不是错误。
func TestUsageOrFailMapping(t *testing.T) {
	var oe *output.Error
	if err := usageOrFail(devdata.ErrEnvNotFound); !errors.As(err, &oe) || oe.Exit != exitUsage {
		t.Errorf("环境不存在应退 %d，得到 %v", exitUsage, err)
	}
	if err := usageOrFail(devdata.ErrProgNotFound); !errors.As(err, &oe) || oe.Exit != exitFailed {
		// status/ls 不会因为"程序没拉过"失败；真走到这里说明状态查询本身出错了
		t.Errorf("status 的意外错误应退 %d，得到 %v", exitFailed, err)
	}
}

// TestEmitFormatsAgree --human 与 JSON 的**内容**一致（表格只换形态，不丢字段）。
func TestEmitFormatsAgree(t *testing.T) {
	res := &devdata.MirrorResult{
		Env: "测试环境", Dir: "/tmp/devdata/测试环境",
		TableCount: 51, RowTotal: 1234, MdlFiles: 281, MdlBytes: 3221007,
		Elapsed: "1m2s",
	}
	// JSON：字段都在
	jsonOut := captureEmit(t, output.FormatJSON, func() error { return emitMirror(res) })
	for _, want := range []string{"测试环境", "51", "1234", "281", "3221007"} {
		if !strings.Contains(jsonOut, want) {
			t.Errorf("JSON 输出缺 %q: %s", want, jsonOut)
		}
	}
	// 表格：同样是这四个数（人是看这张表判断"拉了多少"的）
	tableOut := captureEmit(t, output.FormatTable, func() error { return emitMirror(res) })
	for _, want := range []string{"51", "1234", "281"} {
		if !strings.Contains(tableOut, want) {
			t.Errorf("表格输出缺 %q:\n%s", want, tableOut)
		}
	}
}

// TestEmitCleanPlanDryRun dry-run 必须明确写在输出里（删材料不可撤销）。
func TestEmitCleanPlanDryRun(t *testing.T) {
	plan := &devdata.CleanPlan{
		Items: []devdata.CleanItem{{Kind: "prog", Env: "e", Prog: "axmt500_wf", Path: "/x", Files: 7, Bytes: 4096}},
		Files: 7, Bytes: 4096,
	}
	out := captureEmit(t, output.FormatJSON, func() error { return emitCleanPlan(plan, true) })
	if !strings.Contains(out, `"dryRun":true`) {
		t.Errorf("dry-run 应显式标出来: %s", out)
	}
	out = captureEmit(t, output.FormatJSON, func() error { return emitCleanPlan(plan, false) })
	if !strings.Contains(out, `"dryRun":false`) {
		t.Errorf("执行后 dryRun 应为 false: %s", out)
	}
}

// captureEmit 把 stdout 与全局输出开关临时换掉，跑完还原。
//
// 换的是 common 的全局开关（各命令组都读它），不另开一个"测试专用"参数：
// 那样测到的就不是生产那条路径了。
func captureEmit(t *testing.T, f output.Format, fn func() error) string {
	t.Helper()
	oldStdout := os.Stdout
	oldHuman, oldFormat, oldJSON, oldCSV := common.Human, common.Format, common.JSON, common.CSV
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	common.Human, common.JSON, common.CSV = false, false, false
	common.Format = "json"
	if f == output.FormatTable {
		common.Human = true
	}
	runErr := fn()
	_ = w.Close()
	os.Stdout = oldStdout
	common.Human, common.Format, common.JSON, common.CSV = oldHuman, oldFormat, oldJSON, oldCSV
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	if runErr != nil {
		t.Fatalf("输出失败: %v", runErr)
	}
	return buf.String()
}
