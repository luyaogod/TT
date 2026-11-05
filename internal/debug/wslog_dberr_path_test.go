package debug

// 这一组专门覆盖评审指出的 P1：dbQueryError 曾经被传入**单行** ln，
// 而它内部按 \n 再切，于是"原因在下一行"永远拿不到 —— 中心目标在真实路径上没生效。
// 原单测直接喂多行文本给 collectDBErrorLines，绕过了这条循环，所以掩盖了 bug。
//
// 这里必须**走循环本身**：用 extractDBError(out) 这个与调用点同一逻辑的入口。

import (
	"strings"
	"testing"
)

// sqlplus 真实输出形态：行号行 + 原因行 + 后续噪声。
const realOracleBadOut = `set heading off
ERROR at line 1:
ORA-00942: table or view does not exist

Connected.
SQL> `

// 真实路径：从整段输出里识别报错，并把原因行带进错误消息。
func TestDaemonErrorFromWholeOutputKeepsORA(t *testing.T) {
	err := extractDBError("ds", realOracleBadOut)
	if err == nil {
		t.Fatal("整段输出里有 ORA 报错，应识别出来")
	}
	msg := err.Error()
	if !strings.Contains(msg, "ORA-00942") {
		t.Fatalf("P1 回归：原因行又丢了，得到: %s", msg)
	}
	if !strings.Contains(msg, "ERROR at line 1") {
		t.Fatalf("行号行应保留: %s", msg)
	}
}

// 反过来：如果把单行喂进去（就是**旧的错误调用方式**），原因必然拿不到。
// 这条测试固定住"必须传整段"这个约定，防止有人改回单行。
func TestExtractDBErrorSingleLineLosesCause(t *testing.T) {
	single := "ERROR at line 1:"
	msg := extractDBError("ds", single).Error()
	if strings.Contains(msg, "ORA-00942") {
		t.Fatal("单行输入不可能拿到原因行；这条用例的语义是固定'必须传整段'")
	}
	full := extractDBError("ds", realOracleBadOut).Error()
	if !strings.Contains(full, "ORA-00942") {
		t.Fatalf("整段输入应拿到原因行，得到: %s", full)
	}
}

// 没有报错行的输出不该误判。
func TestExtractDBErrorCleanOutput(t *testing.T) {
	if err := extractDBError("ds", "2026-11-05 11:46:07  100  0.47s  作业=wssp00120_wf"); err != nil {
		t.Fatalf("干净输出不该报错，得到: %v", err)
	}
}

// P3：表名不许写两遍。
func TestDBQueryErrorMessageHasNoDuplicateTableName(t *testing.T) {
	msg := dbQueryError("ds", realOracleBadOut).Error()
	if strings.Contains(msg, "wsfa_t(wsfa_t)") {
		t.Fatalf("表名重复了: %s", msg)
	}
	if !strings.Contains(msg, "wsfa_t") {
		t.Fatalf("应说明是哪张表: %s", msg)
	}
}
