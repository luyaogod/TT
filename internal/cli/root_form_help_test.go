package cli

// 切片 B 返工 · 架构师授权的那一条测试：`rootCmd.Long` 与 `internal/cli/README.md` 里
// **同一张退出码表**必须一致，而 `tt form` 那一行曾只加了 README 一处。
//
// 为什么单独一个文件（而不是挂进 root_test.go）：本切片只被授权"加一条测试"，
// 新增文件不动别人任何一行 —— 越界面保持在"既有文件只动 root.go 的 Long 文本"。
//
// 为什么用"从 README 抠出来再比"而不是再抄一份字面量：抄一份就是第三处副本，
// 三处一起漂的时候没人会红。README 自己写着"这张表是唯一的权威（根帮助里也是它）"，
// 所以这里把 README 当基准、根帮助当被测对象。

import (
	"os"
	"strings"
	"testing"
)

func TestRootHelpAndReadmeAgreeOnFormExitCodes(t *testing.T) {
	b, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("读不到 README.md：%v（它与根帮助是同一张表的两个落点）", err)
	}
	want := formExitTableLines(string(b))
	got := formExitTableLines(rootCmd.Long)

	if len(want) == 0 {
		t.Fatal("README.md 里没有 `tt form` 的码表行 —— 表被删了，或者行的写法变了（必须仍以 `tt form` 开头）")
	}
	if len(got) == 0 {
		t.Fatalf("根帮助（rootCmd.Long）里没有 `tt form` 的码表行。README 说「这张表是唯一的权威（根帮助里也是它）」，"+
			"两处必须都有；包装脚本按码分支，缺一行就少一条判据。\n现有码表：\n%s", rootCmd.Long)
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("两处码表文案不一致（必须逐行相同，只允许共同缩进不同）\n"+
			"--- rootCmd.Long ---\n%s\n--- README.md ---\n%s",
			strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// formExitTableLines 从一段文本里抠出 `tt form` 那一段码表：以 `tt form` 开头的行，
// 加上紧跟其后、**以空白缩进开头且不是新的一行命令**的续行；每行去掉首尾空白。
// 找不到返回 nil。
//
// 缩进要比对两处（根帮助整体缩进 2 格、README 的代码块不缩进），所以这里把首尾空白都去掉，
// 只比"文案本身"—— 这也正是要求里的"逐字一致"。
func formExitTableLines(text string) []string {
	lines := strings.Split(text, "\n")
	for i, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		// 只认**码表行**：`tt form` 后面直接跟数字（用法块里是 `tt form ls --env …`，要跳过）
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "tt form"))
		if !strings.HasPrefix(trimmed, "tt form") || rest == "" || rest[0] < '0' || rest[0] > '9' {
			continue
		}
		out := []string{trimmed}
		for j := i + 1; j < len(lines); j++ {
			body := lines[j]
			trimmed := strings.TrimSpace(body)
			// 续行 = **明显缩进**（≥4 格：根帮助的表格行只有 2 格，那是"下一行"不是续行）、
			// 非空、且不是另起一条命令（`tt …`）
			if trimmed == "" || !strings.HasPrefix(body, "    ") || strings.HasPrefix(trimmed, "tt ") {
				break
			}
			out = append(out, trimmed)
		}
		return out
	}
	return nil
}
