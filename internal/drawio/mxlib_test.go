package drawio

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"tt/internal/testkit"
)

// 移植等价：把 Go 合成出来的库与**原 Node 实现**产出的库逐条比。
//
// 为什么不比文件字节：缩进与字符串转义的细节允许有差异（见 jsonx.go 的那条已知差异），
// 但**解码之后的值**必须一模一样 —— drawio 拿到手的就是解码后的值。比字节会让测试
// 对着噪音红，比语义才是对着契约红。

// baseline 读一份由原 Node 实现产出的库文件。文件在 testdata/drawio/ 下，
// 是搬进 TT 之前用 `npm run build` 生成的，此后**不再重新生成** —— 它钉的就是
// "Go 重写没有改变交给 drawio 的东西"。
func baseline(t *testing.T, name string) []map[string]any {
	t.Helper()
	p := filepath.Join(testkit.RepoRoot(t), "testdata", "drawio", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("读不到基线 %s：%v", p, err)
	}
	return decodeMxLibrary(t, string(b))
}

// decodeMxLibrary 按 drawio 的加载方式解一份库：先当 XML 解析，再取文本节点当 JSON。
//
// 这两步的顺序不能省，也不该被"简化"成一次 json.Unmarshal —— 漏一层转义时
// drawio 报的是 `Unexpected token 'T', "This page" ... is not valid JSON`，
// 这里用同样的两步就当场红。
func decodeMxLibrary(t *testing.T, payload string) []map[string]any {
	t.Helper()
	var root struct {
		Inner string `xml:",chardata"`
	}
	if err := xml.Unmarshal([]byte(payload), &root); err != nil {
		t.Fatalf("不是合法 XML：%v", err)
	}
	var entries []map[string]any
	if err := json.Unmarshal([]byte(root.Inner), &entries); err != nil {
		t.Fatalf("文本节点不是合法 JSON（drawio 会拒收）：%v", err)
	}
	return entries
}

// asPlain 把结构体过一遍 JSON，转成与基线同形的 map —— 这样 reflect.DeepEqual
// 比的是"解码后的值"，而不是两种语言各自的结构体。
func asPlain(t *testing.T, v any) []map[string]any {
	t.Helper()
	s, err := MarshalJSON(v)
	if err != nil {
		t.Fatalf("序列化失败：%v", err)
	}
	var out []map[string]any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		t.Fatalf("回读失败：%v", err)
	}
	return out
}

// TestPortEquivalence 是 P1 的核心判据：两个库逐条与 Node 版基线相同。
func TestPortEquivalence(t *testing.T) {
	src, err := Load()
	if err != nil {
		t.Fatalf("合成失败：%v", err)
	}

	for _, c := range src.Catalogs {
		t.Run(c.Name, func(t *testing.T) {
			want := baseline(t, fmt.Sprintf("%s-v%s.xml", c.Library.Output, c.Library.Version))

			payload, err := MxLibraryOf(c)
			if err != nil {
				t.Fatalf("合成库文件失败：%v", err)
			}
			got := decodeMxLibrary(t, payload)

			if len(got) != len(want) {
				t.Fatalf("条目数不一致：Go %d，Node %d", len(got), len(want))
			}
			for i := range want {
				if !reflect.DeepEqual(got[i], want[i]) {
					t.Errorf("第 %d 条不一致\n  Go  : %#v\n  Node: %#v", i, got[i], want[i])
				}
			}
		})
	}
}

// TestSynthShapeCounts 钉住形状源的规模。
//
// 这不是"数字好看"，是防静默丢件：形状源是 embed 的，少一个文件不会让任何东西报错，
// 只会让库变小 —— 那种事只有数数才发现得了。改形状源时这个数要跟着改。
func TestSynthShapeCounts(t *testing.T) {
	src, err := Load()
	if err != nil {
		t.Fatalf("合成失败：%v", err)
	}
	want := map[string]int{"controls": 18, "business": 9}
	if len(src.Catalogs) != len(want) {
		t.Fatalf("库数 %d，期望 %d", len(src.Catalogs), len(want))
	}
	for _, c := range src.Catalogs {
		if got := len(c.Shapes); got != want[c.Name] {
			t.Errorf("库 %s 有 %d 个形状，期望 %d", c.Name, got, want[c.Name])
		}
	}
}

// TestLibraryEntriesCarryWhatDrawioNeeds 断言每条都带了 drawio 认的那几个字段。
//
// 与等价测试不重复：那条比的是"和 Node 一样"，这条说的是"和 Node 一样**并且**
// 本身是完整的" —— 万一基线自己就是坏的（比如当初生成时就漏了 text 槽位），
// 只有这条会响。
func TestLibraryEntriesCarryWhatDrawioNeeds(t *testing.T) {
	src, err := Load()
	if err != nil {
		t.Fatalf("合成失败：%v", err)
	}
	for _, c := range src.Catalogs {
		entries := asPlain(t, MxEntries(c))
		for i, e := range entries {
			for _, k := range []string{"xml", "w", "h", "title", "aspect"} {
				if _, ok := e[k]; !ok {
					t.Errorf("%s 第 %d 条缺 %q", c.Name, i, k)
				}
			}
			xmlstr, _ := e["xml"].(string)
			if xmlstr == "" {
				t.Errorf("%s 第 %d 条（%v）的 xml 是空的", c.Name, i, e["title"])
			}
			if w, _ := e["w"].(float64); !(w > 0) {
				t.Errorf("%s 第 %d 条（%v）的 w 不是正数：%v", c.Name, i, e["title"], e["w"])
			}
		}
	}
}
