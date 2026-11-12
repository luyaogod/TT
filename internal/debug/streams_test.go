package debug

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 录制协议流回归：把真实 fgldb 的输出原样落盘成样本，表驱动重放。
//
// 目的有二：
//  1. 防止解析层再对某一版输出做"过强假设" —— 换 fgldb 版本、输出变形时，
//     这里先红，而不是等真机上把调试会话卡死；
//  2. 把"哪些行算停站"变成可审阅的事实，而不是散落在 onLine 的 if-else 里。
//
// 新增样本：把整段输出粘进 testdata/fgldb-streams/<名字>.txt，再在下表加一行。
func TestRecordedFgldbStreams(t *testing.T) {
	cases := []struct {
		stream    string
		cmd       string
		start     State // 命令发出后的状态（放行类命令为 running）
		wantState State
		wantLine  int  // 期望的停站行；0 表示不该产生停站
		wantStop  bool // 是否应回填停站信息
	}{
		// 本次事故：next 的停站行不带 `->` 箭头
		{"next-arrowless", "next", StateRunning, StateStopped, 1031, true},
		// 既有行为：next 的停站行带箭头
		{"next-arrow", "next", StateRunning, StateStopped, 812, true},
		// 既有行为：continue → Continuing. → 命中断点
		{"continue-to-breakpoint", "continue", StateStopped, StateStopped, 1023, true},
		// 既有行为：非放行类命令带出的源码行不算停站
		{"list-no-stop", "list", StateStopped, StateStopped, 0, false},
		// 既有行为：SIGINT 中断块
		{"interrupt-stop", "continue", StateRunning, StateStopped, 813, true},
	}

	for _, c := range cases {
		t.Run(c.stream, func(t *testing.T) {
			s := waitTestSession()
			s.started = true
			evs := collectEvents(s)
			p := &pendingCmd{cmd: c.cmd, mode: waitPrompt, res: make(chan *execResult, 1)}
			s.pending = p
			s.setState(c.start)

			for _, ln := range readStream(t, c.stream) {
				s.onLine(ln)
			}

			var got *execResult
			select {
			case got = <-p.res:
			default:
				t.Fatal("样本末行是提示符，命令必须被收口")
			}
			if (got.Stop != nil) != c.wantStop {
				t.Fatalf("停站回填=%v,期望=%v（stop=%+v）", got.Stop != nil, c.wantStop, got.Stop)
			}
			if c.wantStop && got.Stop != nil && got.Stop.Line != c.wantLine {
				t.Fatalf("停站行=%d,期望=%d", got.Stop.Line, c.wantLine)
			}
			if got := s.State(); got != c.wantState {
				t.Fatalf("终态=%s,期望=%s", got, c.wantState)
			}
			if c.wantStop != hasEvent(*evs, "stopped") {
				t.Fatalf("stopped 事件=%v,期望=%v", hasEvent(*evs, "stopped"), c.wantStop)
			}
		})
	}
}

// readStream 读一份录制样本，去掉末尾空行（保留行首空白：fgldb 的缩进是格式的一部分）。
func readStream(t *testing.T, name string) []string {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "fgldb-streams", name+".txt"))
	if err != nil {
		t.Fatalf("读样本失败: %v", err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, strings.TrimRight(sc.Text(), "\r"))
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("读样本失败: %v", err)
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
