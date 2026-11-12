package debug

// `--no-restore-bp`：本轮**不恢复**上次的断点（P1）。
//
// 存在理由：断点存档按 module+prog 复用，而"同一作业换了模块/客制副本"时上一轮的位置
// 可能已经不被加载了（下断前会核 info sources，不加载的直接拒）。想从干净状态开始、
// 又不想逐条 `bp del --all`，就需要这个开关。
//
// 两条纪律：① 开关只对**这一轮**生效（不写回 config，不污染下一轮）；
//          ② 关掉恢复时**一条 break 都不许发**（发了就等于"恢复了"）。

import (
	"strings"
	"testing"
	"time"
)

func TestRestoreBreakpointsSkippedByNoRestoreBP(t *testing.T) {
	dir := t.TempDir()
	store := []StoredBP{
		{File: "axm_axmt500_wf.4gl", Line: 4130, LineText: `LET g_x = 1`, Enabled: true},
		{File: "axm_axmt500_wf.4gl", Line: 4140, LineText: `LET g_y = 2`, Enabled: true},
	}
	if err := saveBPs(dir, "axm", "axmt500_wf", store); err != nil {
		t.Fatalf("准备存档失败：%v", err)
	}

	var sent []string
	s := &Session{
		Module: "axm", Prog: "axmt500_wf",
		bps:      map[int]*Breakpoint{},
		srcCache: map[string]srcCacheEntry{},
		cfg:      &Config{DataDir: dir, NoRestoreBPs: true},
		state:    StateStopped,
		emit:     func(Event) {},
	}
	s.execFn = func(kind, cmd string, mode int, timeout time.Duration) (*execResult, error) {
		sent = append(sent, cmd)
		return &execResult{Cmd: cmd}, nil
	}

	s.restoreBreakpoints()

	for _, c := range sent {
		if strings.HasPrefix(c, "break ") {
			t.Fatalf("--no-restore-bp 下仍然发了下断命令：%v", sent)
		}
	}
}

// 恢复时的"文件这次没加载"必须是**跳过**，不是故障（P1 第三件）：
// 存档按 module+prog 复用，换过客制副本/模块前缀时上一轮的位置本来就不该再挂。
// 它不该拦住同一份存档里其它还成立的断点。
func TestRestoreBreakpointsSkipsUnloadedFile(t *testing.T) {
	dir := t.TempDir()
	store := []StoredBP{
		{File: "axm_axmt500_wf.4gl", Line: 4130, Enabled: true},
		{File: "sub_s_bsft001_wf.4gl", Line: 2648, Enabled: true}, // 从没被加载的那份
	}
	if err := saveBPs(dir, "axm", "axmt500_wf", store); err != nil {
		t.Fatalf("准备存档失败：%v", err)
	}

	var sent []string
	var logs []string
	s := &Session{
		Module: "axm", Prog: "axmt500_wf",
		bps:      map[int]*Breakpoint{},
		srcCache: map[string]srcCacheEntry{},
		cfg:      &Config{DataDir: dir},
		state:    StateStopped,
		emit: func(e Event) {
			if e.Type == "log" {
				logs = append(logs, e.Text)
			}
		},
	}
	s.execFn = func(kind, cmd string, mode int, timeout time.Duration) (*execResult, error) {
		sent = append(sent, cmd)
		switch {
		case strings.HasPrefix(cmd, "break "):
			return &execResult{Cmd: cmd, Lines: []string{cmd,
				"Breakpoint 1 at 0x00000000: file axm_axmt500_wf.4gl, line 4130."}}, nil
		case strings.Contains(cmd, "info sources"):
			// 只有 axm 那份被加载
			return &execResult{Cmd: cmd, Lines: []string{cmd, "axm_axmt500_wf.4gl"}}, nil
		}
		return &execResult{Cmd: cmd}, nil
	}

	s.restoreBreakpoints()

	var breaks []string
	for _, c := range sent {
		if strings.HasPrefix(c, "break ") {
			breaks = append(breaks, c)
		}
	}
	if len(breaks) != 1 {
		t.Fatalf("下断命令 = %v，期望只有 axm 那条（未加载的必须跳过、且不拦住别人）", breaks)
	}
	if strings.Contains(breaks[0], "sub_s_bsft001_wf") {
		t.Fatalf("未加载的文件被恢复了：%v", breaks)
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "断点恢复跳过") || !strings.Contains(joined, "sub_s_bsft001_wf.4gl") {
		t.Fatalf("没把跳过的事实说清楚：\n%s", joined)
	}
}

// 反向：不给开关时要照常恢复（不然这个开关就成了"断点持久化坏了"的替罪羊）。
func TestRestoreBreakpointsStillRunsWithoutFlag(t *testing.T) {
	dir := t.TempDir()
	store := []StoredBP{{File: "axm_axmt500_wf.4gl", Line: 4130, Enabled: true}}
	if err := saveBPs(dir, "axm", "axmt500_wf", store); err != nil {
		t.Fatalf("准备存档失败：%v", err)
	}

	var sent []string
	s := &Session{
		Module: "axm", Prog: "axmt500_wf",
		bps:      map[int]*Breakpoint{},
		srcCache: map[string]srcCacheEntry{},
		cfg:      &Config{DataDir: dir},
		state:    StateStopped,
		emit:     func(Event) {},
	}
	s.execFn = func(kind, cmd string, mode int, timeout time.Duration) (*execResult, error) {
		sent = append(sent, cmd)
		// 让下断真的成功：回 fgldb 的真实报文形状。
		if strings.HasPrefix(cmd, "break ") {
			return &execResult{Cmd: cmd, Lines: []string{cmd,
				"Breakpoint 1 at 0x00000000: file axm_axmt500_wf.4gl, line 4130."}}, nil
		}
		if strings.Contains(cmd, "info sources") {
			return &execResult{Cmd: cmd, Lines: []string{cmd, "axm_axmt500_wf.4gl"}}, nil
		}
		return &execResult{Cmd: cmd}, nil
	}

	s.restoreBreakpoints()

	found := false
	for _, c := range sent {
		if strings.HasPrefix(c, "break ") {
			found = true
		}
	}
	if !found {
		t.Fatalf("没给 --no-restore-bp 却没恢复断点（命令：%v）", sent)
	}
}
