package form

// 牙齿 b：`tt form show --json` 的输出**逐字节**钉在 golden 文件上（契约 §3 / C5.2）。
//
// 为什么要这一条：C5.2 说"键名固定、可 grep；新增键要改本契约"。文档挡不住改名 ——
// 而改名对消费方的表现是 `jq .fields[].colName` 突然打出 null，看起来像"这张表没字段"。
// golden 一改就红，改的人必须显式动这份文件（并同步契约 C5.2），这就是它的全部价值。

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormShowJSONGolden(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{forms: map[string][]byte{
		"demomod/demoform": []byte(fixtureBasic(t)),
	}})

	out, err := runForm(t, "show", "--env", "demoenv", "--module", "demomod",
		"--form", "demoform", "--json")
	if err != nil {
		t.Fatalf("show --json 失败: %v", err)
	}

	gp := filepath.Join("testdata", "golden", "form_show_basic.json")
	want, rerr := os.ReadFile(gp)
	if rerr != nil {
		t.Fatalf("读 golden 失败：%v\n（牙 b 的 golden 必须入库 —— 删了它这条测试就失去意义）", rerr)
	}

	if out != string(want) {
		t.Errorf("--json 输出与 golden 不一致。\n"+
			"若是**有意**改了输出形状：先改契约 C5.2，再显式编辑 %s（本测试只比对，不自动重生成）。\n"+
			"--- got ---\n%s\n--- want ---\n%s", gp, out, want)
	}

	// 键集单列一遍：字节不同时上面那句只说"不一样"，这里直接点名多了/少了哪个键。
	if got, wantKeys := keySet(t, out), keySet(t, string(want)); got != wantKeys {
		t.Errorf("顶层键集漂移：\n got = %v\nwant = %v", got, wantKeys)
	}
}

// decodeShow 把 --json 的信封解成 map（用例里按契约键名取值）。
func decodeShow(t *testing.T, out string) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("--json 输出不是 JSON: %v\n%s", err, out)
	}
	return doc
}

// keySet 是顶层键的集合（排序后拼接，便于比较）。
func keySet(t *testing.T, s string) string {
	t.Helper()
	var doc map[string]json.RawMessage
	if err := json.Unmarshal([]byte(s), &doc); err != nil {
		t.Fatalf("解 JSON 失败: %v", err)
	}
	keys := make([]string, 0, len(doc))
	for k := range doc {
		keys = append(keys, k)
	}
	// 键数很少（7 个），手写插入排序比引 sort 更省事也不易错
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return strings.Join(keys, ",")
}
