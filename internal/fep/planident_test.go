package fep

import (
	"encoding/json"
	"strings"
	"testing"
)

// 计划步骤的**身份**（契约 C3 + 计划 §"计划与录制元数据存 colName"）。
//
// 语义计划（`BuildPlan`）是**离线**从 `.4fd` 推出来的，所以它只能记 `colName`
// （`.4fd` 里就有、且表单内稳定）；`fieldId` **绝不从 `.4fd` 取** ——
// `.4fd` 的编号每次保存全量重排（engine/SPEC.md:257，契约 B5/H5）。那一栏留给
// 运行期证据填（回放侧见 ReplayStep.FieldID），没证据时就是 0 = 未记录。

const planSnippetWithFieldID = `
<Edit name="pmdl_t.pmdl002" colName="pmdl002" fieldId="18" sqlTabName="pmdl_t" widget="Edit" comment="cmt_pmdl002" tabIndex="1" posX="40" posY="20"/>
<ComboBox name="pmdl_t.pmdl005" colName="pmdl005" fieldId="21" sqlTabName="pmdl_t" widget="ComboBox" comment="cmt_pmdl005" tabIndex="2"/>
<CheckBox name="formonly.pmdl012" colName="pmdl012" fieldId="33" sqlTabName="pmdl_t" widget="CheckBox" comment="cmt_pmdl012" tabIndex="3" valueChecked="Y"/>
`

// TestPlanStepCarriesColNameFromFormDefinition：计划步骤要带 colName（身份的一等键）。
func TestPlanStepCarriesColNameFromFormDefinition(t *testing.T) {
	steps := BuildPlan("apmt500_wf", ParseForm4fd("apmt500_wf", planSnippetWithFieldID))
	if len(steps) != 3 {
		t.Fatalf("步骤数 = %d，想要 3：\n%s", len(steps), PlanText(steps))
	}
	want := []string{"pmdl002", "pmdl005", "pmdl012"}
	for i, w := range want {
		if steps[i].ColName != w {
			t.Errorf("第 %d 步的 colName = %q，想要 %q（没有身份就没法回放时重新定位）",
				i+1, steps[i].ColName, w)
		}
	}
}

// TestPlanStepNeverTakesFieldIDFrom4fd 是**反证**（契约 B5/H5 的红线）：
// 即使 `.4fd` 里写着 `fieldId="18"`，计划步骤的 fieldId 也必须是 0（未记录）。
// 谁要是"顺手从 .4fd 取编号"，这条立刻红 —— 那正是每次保存都会全量重排的那个号。
func TestPlanStepNeverTakesFieldIDFrom4fd(t *testing.T) {
	steps := BuildPlan("apmt500_wf", ParseForm4fd("apmt500_wf", planSnippetWithFieldID))
	for _, s := range steps {
		if s.FieldID != 0 {
			t.Fatalf("第 %d 步的 fieldId = %d —— `.4fd` 的编号不是稳定标识，绝不许当身份用",
				s.Seq, s.FieldID)
		}
	}
}

// TestPlanStepJSONOmitsUnrecordedIdentity：旧读者/旧 JSON 的兼容面 ——
// 没记身份时不冒空键；记了就要在。
func TestPlanStepJSONOmitsUnrecordedIdentity(t *testing.T) {
	b, err := json.Marshal(PlanStep{Seq: 1, Op: "set", Role: "cmt_pmdl002"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "colName") || strings.Contains(string(b), "fieldId") {
		t.Fatalf("没记身份的计划步骤不该写出空键：%s", b)
	}
	b, err = json.Marshal(PlanStep{Seq: 1, Op: "set", Role: "cmt_pmdl002", ColName: "pmdl002", FieldID: 18})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"colName":"pmdl002"`) || !strings.Contains(string(b), `"fieldId":18`) {
		t.Fatalf("身份没写进 JSON：%s", b)
	}
	// 旧输出（没有这两个键）照样读得进来：零值 = 未记录，不是错误
	var old PlanStep
	if err := json.Unmarshal([]byte(`{"seq":2,"op":"set","role":"cmt_pmdl005"}`), &old); err != nil {
		t.Fatalf("旧计划步骤读不动了：%v", err)
	}
	if old.ColName != "" || old.FieldID != 0 || old.Op != "set" {
		t.Fatalf("旧计划步骤读出来的形状不对：%+v", old)
	}
}
