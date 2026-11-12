package debug

// 下断前核"这个文件这次运行到底加载了没有"（P1）。
//
// 现场（2026-11-11，福联 1108）：`bp add s_bsft001_wf_...` 被 fgldb 落成
// `Breakpoint 4 at 0x00000000: file sub_s_bsft001_wf.4gl, line 2648` —— 编号、位置、
// 落盘一切正常，然后 0 命中。真正被加载的是 `csub_s_bsft001_wf.4gl`（客制副本），
// `info sources` 里根本没有 `sub_` 那份。
//
// 判据只能是**内核说这次运行加载了哪些文件**（info sources）：文件在磁盘上、ResolveSource
// 也读得到，fgldb 就照单全收 —— 所以"下断成功"完全不代表"这一刀会命中"。

import (
	"strings"
	"testing"
	"time"
)

func TestBPLocFileLine(t *testing.T) {
	cases := []struct {
		loc, wantFile, wantLine string
	}{
		{"csub_s_bsft001_wf.4gl:9385", "csub_s_bsft001_wf.4gl", "9385"},
		{"csub_s_bsft001_wf.4gl", "csub_s_bsft001_wf.4gl", ""},
		{"9385", "", ""},                        // 裸行号：没有文件可核
		{"s_bsft001_wf_asft300_sub_po", "", ""}, // 函数名
		{"a.4gl:foo", "", ""},                   // 冒号后不是行号，不是文件+行位置
		{"", "", ""},                            // 空
		{"  asf_bsft001_wf.4gl:804  ", "asf_bsft001_wf.4gl", "804"}, // 前后空白
	}
	for _, c := range cases {
		f, l := bplocFileLine(c.loc)
		if f != c.wantFile || l != c.wantLine {
			t.Errorf("bplocFileLine(%q) = (%q,%q)，期望 (%q,%q)", c.loc, f, l, c.wantFile, c.wantLine)
		}
	}
}

// TestBPLocKey 钉"同一个文件的两种命名"必须归到同一把键上 —— 这正是现场两回的形态：
// 模块前缀写错（asf_/axm_）、标准与客制副本（sub_/csub_）。
func TestBPLocKey(t *testing.T) {
	cases := map[string]string{
		"asf_axmt500_wf.4gl":    "axmt500_wf",
		"axm_axmt500_wf.4gl":    "axmt500_wf",
		"sub_s_bsft001_wf.4gl":  "s_bsft001_wf",
		"csub_s_bsft001_wf.4gl": "s_bsft001_wf",
		"plain.4gl":             "plain",
	}
	for in, want := range cases {
		if got := bplocKey(in); got != want {
			t.Errorf("bplocKey(%q) = %q，期望 %q", in, got, want)
		}
	}
}

// TestBPLocSuggestPutsRealFileFirst 是现场的还原：人给 sub_，真正加载的是 csub_。
// 建议里 csub_ 必须排在第一位（第二位允许是"同作业的标准副本"这种弱候选）。
func TestBPLocSuggestPutsRealFileFirst(t *testing.T) {
	loaded := []string{"asf_bsft001_wf.4gl", "csub_s_bsft001_wf.4gl", "lib_cl_ap_code.4gl"}
	got := bplocSuggest("sub_s_bsft001_wf.4gl", loaded)
	if len(got) == 0 {
		t.Fatal("没给出任何建议：现场那条断点就会静默 0 命中")
	}
	if got[0] != "csub_s_bsft001_wf.4gl" {
		t.Fatalf("第一条建议是 %q，期望 csub_s_bsft001_wf.4gl（同键优先）；全部=%v", got[0], got)
	}
	for _, g := range got {
		if g == "lib_cl_ap_code.4gl" {
			t.Fatalf("无关文件被当成建议：%v", got)
		}
	}
}

// TestBPLocNotLoadedMessage 钉拒绝理由**可照做**：说清"永远 0 命中"、给出相近文件、
// 并直接拼好可以粘贴的 `bp add` 命令（只报错不给出路，人会去猜文件名）。
func TestBPLocNotLoadedMessage(t *testing.T) {
	err := &BPLocNotLoadedError{
		File: "sub_s_bsft001_wf.4gl", Line: "2648",
		Loaded: []string{"csub_s_bsft001_wf.4gl"}, Total: 41,
	}
	msg := err.Error()
	for _, want := range []string{
		"sub_s_bsft001_wf.4gl:2648",
		"永远命中 0 次",
		"csub_s_bsft001_wf.4gl",
		"tt debug bp add csub_s_bsft001_wf.4gl:2648",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("拒绝理由里少了 %q：\n%s", want, msg)
		}
	}
}

// bpSourceSession 造一条"停站 + info sources 有答案"的会话（PTY 由 execFn 顶掉）。
func bpSourceSession(sources []string, calls *int) *Session {
	s := &Session{
		bps:      map[int]*Breakpoint{},
		srcCache: map[string]srcCacheEntry{},
		cfg:      &Config{},
		state:    StateStopped,
	}
	s.execFn = func(kind, cmd string, mode int, timeout time.Duration) (*execResult, error) {
		if calls != nil {
			*calls++
		}
		if strings.Contains(cmd, "info sources") {
			return &execResult{Cmd: cmd, Lines: append([]string{cmd}, sources...)}, nil
		}
		return &execResult{Cmd: cmd}, nil
	}
	return s
}

func TestVerifyBPLoadedRejectsFileNotInSources(t *testing.T) {
	calls := 0
	s := bpSourceSession([]string{"asf_bsft001_wf.4gl, csub_s_bsft001_wf.4gl"}, &calls)
	err := s.VerifyBPLoaded(BreakpointOpts{Location: "sub_s_bsft001_wf.4gl:2648"})
	if err == nil {
		t.Fatal("位置在 info sources 里不存在，却放行了 —— 这正是那条 0 命中的断点")
	}
	var nl *BPLocNotLoadedError
	if !asNotLoaded(err, &nl) {
		t.Fatalf("该是 *BPLocNotLoadedError，实得 %T: %v", err, err)
	}
	if len(nl.Loaded) == 0 || nl.Loaded[0] != "csub_s_bsft001_wf.4gl" {
		t.Fatalf("没把真正加载的 csub_ 文件给出来：%v", nl.Loaded)
	}
	if nl.Total != 2 {
		t.Fatalf("Total = %d，期望 2（说明判据来自 info sources 而不是空表）", nl.Total)
	}
}

func TestVerifyBPLoadedAcceptsLoadedFile(t *testing.T) {
	s := bpSourceSession([]string{"asf_bsft001_wf.4gl, csub_s_bsft001_wf.4gl"}, nil)
	if err := s.VerifyBPLoaded(BreakpointOpts{Location: "csub_s_bsft001_wf.4gl:9385"}); err != nil {
		t.Fatalf("文件明明加载了却被拒：%v", err)
	}
	// 只差大小写也算加载（Linux 上不会发生，但不该因此拒掉）。
	if err := s.VerifyBPLoaded(BreakpointOpts{Location: "CSUB_S_BSFT001_WF.4GL:9385"}); err != nil {
		t.Fatalf("大小写不同被当成没加载：%v", err)
	}
}

// TestVerifyBPLoadedSkipsNonFileLocations 钉"不多花那趟往返"：
// 裸行号 / 函数名 / 监视点都没有文件可核，一个字节都不该发。
func TestVerifyBPLoadedSkipsNonFileLocations(t *testing.T) {
	for _, o := range []BreakpointOpts{
		{Location: "2648"},
		{Location: "s_bsft001_wf_asft300_sub_po"},
		{Location: "g_current_row", Kind: BPKindWatch},
	} {
		calls := 0
		s := bpSourceSession([]string{"asf_bsft001_wf.4gl"}, &calls)
		if err := s.VerifyBPLoaded(o); err != nil {
			t.Fatalf("%q 不该被拒：%v", o.Location, err)
		}
		if calls != 0 {
			t.Fatalf("%q 没有文件可核，却发了 %d 条命令（白多一趟往返）", o.Location, calls)
		}
	}
}

// TestVerifyBPLoadedFailSoftWhenSourcesUnknown 是最重要的一条反向纪律：
// **读不到 info sources 时必须放行**。把"不知道"当成"一个文件都没加载"，会让
// PTY 忙/未停站时每一条下断都被拒 —— 那比漏判一次严重得多。
func TestVerifyBPLoadedFailSoftWhenSourcesUnknown(t *testing.T) {
	s := bpSourceSession(nil, nil) // info sources 回空
	if err := s.VerifyBPLoaded(BreakpointOpts{Location: "whatever_unknown.4gl:10"}); err != nil {
		t.Fatalf("info sources 读不回来时应当放行，却拒了：%v", err)
	}

	failing := bpSourceSession([]string{"x.4gl"}, nil)
	failing.execFn = func(kind, cmd string, mode int, timeout time.Duration) (*execResult, error) {
		if strings.Contains(cmd, "info sources") {
			return nil, ErrNotStopped // 程序刚跑起来
		}
		return &execResult{Cmd: cmd}, nil
	}
	if err := failing.VerifyBPLoaded(BreakpointOpts{Location: "whatever_unknown.4gl:10"}); err != nil {
		t.Fatalf("读 info sources 失败时应当放行，却拒了：%v", err)
	}
}

// TestVerifyBPLoadedCachesSources 钉缓存：恢复断点会连着发十几条 break，
// 每条都读一次 info sources 就白多十几趟往返（真机上每趟几十毫秒）。
func TestVerifyBPLoadedCachesSources(t *testing.T) {
	calls := 0
	s := bpSourceSession([]string{"asf_bsft001_wf.4gl"}, &calls)
	for i := 0; i < 5; i++ {
		if err := s.VerifyBPLoaded(BreakpointOpts{Location: "asf_bsft001_wf.4gl:804"}); err != nil {
			t.Fatalf("第 %d 次被拒：%v", i+1, err)
		}
	}
	if calls != 1 {
		t.Fatalf("info sources 发了 %d 次，期望 1 次（10 秒内应命中缓存）", calls)
	}
}

// TestBreakWithRejectsUnloadedFileBeforeSendingBreak 钉"拒绝发生在发命令之前"：
// 拒绝时内核**不能**收到 break 命令 —— 收到了就等于已经把断点挂上去了，再报错也晚了。
func TestBreakWithRejectsUnloadedFileBeforeSendingBreak(t *testing.T) {
	var got []string
	s := bpSourceSession([]string{"csub_s_bsft001_wf.4gl"}, nil)
	s.execFn = func(kind, cmd string, mode int, timeout time.Duration) (*execResult, error) {
		got = append(got, cmd)
		if strings.Contains(cmd, "info sources") {
			return &execResult{Cmd: cmd, Lines: []string{cmd, "csub_s_bsft001_wf.4gl"}}, nil
		}
		return &execResult{Cmd: cmd}, nil
	}
	if _, err := s.BreakWith(BreakpointOpts{Location: "sub_s_bsft001_wf.4gl:2648"}); err == nil {
		t.Fatal("没加载的文件却下断成功")
	}
	for _, c := range got {
		if strings.HasPrefix(c, "break ") {
			t.Fatalf("拒绝之前已经把 break 发出去了（内核命令：%v）", got)
		}
	}
}

// asNotLoaded 是 errors.As 的短封装（本文件只关心这一种错误）。
func asNotLoaded(err error, target **BPLocNotLoadedError) bool {
	for e := err; e != nil; {
		if v, ok := e.(*BPLocNotLoadedError); ok {
			*target = v
			return true
		}
		u, ok := e.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}
