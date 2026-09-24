package cli

// root_test.go —— 根命令帮助与 README 的漂移防线。
//
// **为什么这条测试住在这个包**：`internal/dev/cli` 里的 TestUsageTextHasNoStaleAdvice
// 覆盖不到 rootCmd.Long，而那个包不能 import 本包（`internal/cli/root.go` →
// `internal/cli/dev` → `internal/dev/cli`，反过来就成环）。rootCmd 是本包的包级变量，
// 同包测试才直接看得到它 —— 所以这条断言只能住在这里。
//
// 它要抓的是**同一个事实写在两处、只改了一处**：`tt --help` 与 README 的首屏都有一张
// "四块能力"的清单，`.tzs` 那行曾经两处都还写着已删除的 `call` 网关，而当时的漂移测试
// 只看 tzsUsage/Usage 两个常量 —— 改一处不影响另一处，也没有任何测试会响。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRootHelpAndReadmeHaveNoStaleAdvice 断言根帮助与 README 都不再教已删除的调用。
//
// 读包外的文件是有先例的（`internal/dev/tzs/corpus_test.go` 读 `../../../engine/corpus.manifest`），
// 而 README.md 属于仓库本体、不是外部语料 —— 所以它**不在**就判失败，不用 Skip 糊过去。
func TestRootHelpAndReadmeHaveNoStaleAdvice(t *testing.T) {
	readmePath := filepath.Join("..", "..", "README.md")
	b, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("读不到 %s：%v（README.md 是仓库本体的一部分，缺了就是仓库坏了）", readmePath, err)
	}

	stale := []struct{ bad, why string }{
		{"走 call", "call 网关已删除：动词就是函数名，参数用 JSON 给"},
		{"call <fn>", "call 网关已删除"},
		{"tzs call", "call 网关已删除"},
	}
	for name, text := range map[string]string{
		"rootCmd.Long": rootCmd.Long,
		"README.md":    string(b),
	} {
		for _, s := range stale {
			if strings.Contains(text, s.bad) {
				t.Errorf("%s 里还在教 %s（%s）", name, s.bad, s.why)
			}
		}
	}
}

// TestRootHelpPointsAtNamedVerbs 断言根帮助把 .tzs 的入口说成"具名动词"。
//
// 只禁掉旧写法（上一条）不够：把那一行整句删掉也能让它通过，而调用方从此不知道
// 表单包该怎么改 —— 那比教错更坏。所以要有一条**正向**断言。
func TestRootHelpPointsAtNamedVerbs(t *testing.T) {
	long := rootCmd.Long
	for _, want := range []string{"tt dev tzs", "tt dev tzc", "tt dict", "tt debug"} {
		if !strings.Contains(long, want) {
			t.Errorf("根帮助里该出现 %s（四块能力的入口要一眼看全）", want)
		}
	}
	if !strings.Contains(long, "动词") {
		t.Errorf("根帮助里该说清 .tzs 的入口是具名动词，而不是删掉的 call 网关")
	}
}
