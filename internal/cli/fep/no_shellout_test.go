package fep

// 静态判据：**`tt fep` 不许再起自己的子进程去调调试线**。
//
// # 钉的是什么
//
// 从前录制那套动作是靠 `runSelf(cmd, "debug", …)` 做的：把 `tt debug` 当 API 用。
// 代价是每次一个进程、参数只能用 argv 表达、失败原因只能从退出码反推 ——
// 于是"编排"永远沉不到内核里（P2 要消灭的正是这个）。
//
// 现在两条线共用 `internal/debugclient`。这条判据防止有人图省事再写回去：
// 那种改动**能编译、能跑**，只有多会话或错误码不一致时才露出来 —— 属于最贵的一类回归。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFepDoesNotShellOutToDebug(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("读包目录失败：%v", err)
	}
	forbidden := []struct{ needle, why string }{
		{"runSelf(", "起自己的子命令 = 又回到子进程调用"},
		{"selfArgs(", "同上（runSelf 的参数拼装）"},
		{`"os/exec"`, "本包不该再需要起子进程"},
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		b, rerr := os.ReadFile(filepath.Clean(name))
		if rerr != nil {
			t.Fatalf("读 %s 失败：%v", name, rerr)
		}
		// 只看**代码行**：注释里提到 runSelf（解释历史）不算违规。
		var code strings.Builder
		for _, line := range strings.Split(string(b), "\n") {
			if t := strings.TrimSpace(line); strings.HasPrefix(t, "//") {
				continue
			}
			code.WriteString(line)
			code.WriteByte('\n')
		}
		src := code.String()
		for _, f := range forbidden {
			if strings.Contains(src, f.needle) {
				t.Errorf("%s 里出现了 %q —— %s；调试面的调用请走 internal/debugclient", name, f.needle, f.why)
			}
		}
	}
}
