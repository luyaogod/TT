package debug

// `bp del` 的三种形态（P1）：编号 / 位置 / --all。
//
// 现场三条摩擦，全部来自"删完编号会重排"这件事：
//  1. `bp del 4` 与 `bp del 5` 各删错一条 —— 第二次用的编号在第一次删完就失效了；
//  2. 只记得位置（file:line），不记得编号；
//  3. 想清空，只能一条条数。删除后不重列，下一次还会用过期编号。
//
// 所以这里钉四件事：删对了、删错了不假装成功、按位置删、--all 倒序删、删完重列。

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"tt/internal/cli/common"
	"tt/internal/testkit"
)

// bpDelStub 是有状态的假服务：DELETE 真的会改自己的断点表，GET 回当前表。
// 无状态桩会让"删完重列"这条断言永远看到旧表 —— 那就测不出重排这件事。
type bpDelStub struct {
	mu      sync.Mutex
	bps     []map[string]any
	deleted []string // 收到的 DELETE 路径（按顺序）
	sources []string

	reject string // 非空 ⇒ 一律 400（模拟 running 拒绝）
}

func (s *bpDelStub) install(t *testing.T) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		// 会话解析不能被模拟的 running 拒绝影响到：那条 400 说的是**调试命令**，
		// 而 GET /api/sessions 不经过写闸门。
		if strings.HasSuffix(p, "/api/sessions") {
			_ = json.NewEncoder(w).Encode(map[string]any{"sessions": []map[string]any{
				{"id": "sOnly", "state": "stopped", "env": "主机109", "prog": "axmt500_wf"},
			}})
			return
		}
		if s.reject != "" {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "code": "USAGE", "error": s.reject})
			return
		}
		switch {
		case strings.HasSuffix(p, "/sources"):
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "sources": s.sources})
		case r.Method == http.MethodDelete && strings.Contains(p, "/breakpoints/"):
			num := p[strings.LastIndex(p, "/")+1:]
			s.deleted = append(s.deleted, num)
			kept := s.bps[:0]
			for _, b := range s.bps {
				if fmt.Sprint(b["num"]) != num {
					kept = append(kept, b)
				}
			}
			s.bps = kept
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		case strings.HasSuffix(p, "/breakpoints"):
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "breakpoints": s.bps})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		}
	}))
	t.Cleanup(ts.Close)
	oldURL := dbgAPIURL
	dbgAPIURL = ts.URL
	t.Cleanup(func() { dbgAPIURL = oldURL })
}

func newBPDelStub(t *testing.T, bps ...map[string]any) *bpDelStub {
	t.Helper()
	resetTargetFlags(t)
	s := &bpDelStub{
		bps:     bps,
		sources: []string{"axm_axmt500_wf.4gl", "csub_s_bsft001_wf.4gl"},
	}
	s.install(t)
	common.JSON = false
	t.Cleanup(func() { bpDelAllFlag = false })
	return s
}

func bpRow(num int, file string, line int) map[string]any {
	return map[string]any{"num": num, "kind": "break", "file": file, "line": line,
		"enabled": true, "hits": 0}
}

// ① 按编号删：删对了，并且删完自动重列（否则下一次用的还是过期编号）。
func TestBPDelByNumReListsRemaining(t *testing.T) {
	s := newBPDelStub(t, bpRow(1, "axm_axmt500_wf.4gl", 4130), bpRow(2, "axm_axmt500_wf.4gl", 4140))

	_, out := testkit.CaptureStdout(t, func() int {
		if err := debugBPDelCmd.RunE(debugBPDelCmd, []string{"2"}); err != nil {
			t.Errorf("按编号删失败：%v", err)
		}
		return 0
	})
	if len(s.deleted) != 1 || s.deleted[0] != "2" {
		t.Fatalf("DELETE 收到 %v，期望只删 2", s.deleted)
	}
	for _, want := range []string{"断点 #2 已删除", "#1", "axm_axmt500_wf.4gl:4130"} {
		if !strings.Contains(out, want) {
			t.Errorf("输出里少了 %q：\n%s", want, out)
		}
	}
	if strings.Contains(out, "axm_axmt500_wf.4gl:4140") {
		t.Errorf("删掉的 #2 还留在重列里：\n%s", out)
	}
}

// ② 编号不存在：必须报错且**不发 DELETE**（fgldb 对不存在的编号只回一句英文提示，
// 把它当成功会给人"删掉了"的假象，而下一次仍然删错）。
func TestBPDelUnknownNumFailsWithoutDelete(t *testing.T) {
	s := newBPDelStub(t, bpRow(1, "axm_axmt500_wf.4gl", 4130))

	err := debugBPDelCmd.RunE(debugBPDelCmd, []string{"9"})
	if err == nil {
		t.Fatal("删一个不存在的编号却成功了")
	}
	if !strings.Contains(err.Error(), "没有断点 #9") {
		t.Fatalf("错误说得不清楚：%v", err)
	}
	if !strings.Contains(err.Error(), "#1") {
		t.Fatalf("没告诉人当前有哪些编号：%v", err)
	}
	if len(s.deleted) != 0 {
		t.Fatalf("编号不存在却发了 DELETE：%v", s.deleted)
	}
}

// ③ 按位置删：人只记得 file:line（编号会重排，位置不会）。
func TestBPDelByLocation(t *testing.T) {
	s := newBPDelStub(t, bpRow(1, "axm_axmt500_wf.4gl", 4130), bpRow(2, "axm_axmt500_wf.4gl", 4140))

	_, out := testkit.CaptureStdout(t, func() int {
		if err := debugBPDelCmd.RunE(debugBPDelCmd, []string{"axm_axmt500_wf.4gl:4140"}); err != nil {
			t.Errorf("按位置删失败：%v", err)
		}
		return 0
	})
	if len(s.deleted) != 1 || s.deleted[0] != "2" {
		t.Fatalf("DELETE 收到 %v，期望删编号 2（4140 那一行）", s.deleted)
	}
	if !strings.Contains(out, "断点 #2 已删除") {
		t.Errorf("没说删了哪一条：\n%s", out)
	}
}

// ④ 给了文件不给行号、文件上有多条：不替人挑一条（挑错了看起来还像成功）。
func TestBPDelByLocationAmbiguousRefuses(t *testing.T) {
	s := newBPDelStub(t, bpRow(1, "axm_axmt500_wf.4gl", 4130), bpRow(2, "axm_axmt500_wf.4gl", 4140))

	err := debugBPDelCmd.RunE(debugBPDelCmd, []string{"axm_axmt500_wf.4gl"})
	if err == nil {
		t.Fatal("两条都匹配却删了一条")
	}
	if !strings.Contains(err.Error(), "匹配到 2 条") {
		t.Fatalf("没把歧义说清楚：%v", err)
	}
	if len(s.deleted) != 0 {
		t.Fatalf("有歧义却还是发了 DELETE：%v", s.deleted)
	}
}

// ⑤ --all 必须**倒序**删：fgldb 删完会重排剩余编号，正序会跳过一半。
func TestBPDelAllDeletesDescending(t *testing.T) {
	s := newBPDelStub(t,
		bpRow(1, "axm_axmt500_wf.4gl", 4130),
		bpRow(2, "axm_axmt500_wf.4gl", 4140),
		bpRow(3, "csub_s_bsft001_wf.4gl", 9385),
	)
	bpDelAllFlag = true

	_, out := testkit.CaptureStdout(t, func() int {
		if err := debugBPDelCmd.RunE(debugBPDelCmd, nil); err != nil {
			t.Errorf("--all 失败：%v", err)
		}
		return 0
	})
	want := []string{"3", "2", "1"}
	if strings.Join(s.deleted, ",") != strings.Join(want, ",") {
		t.Fatalf("删除顺序 = %v，期望 %v（正序删会跳过一半）", s.deleted, want)
	}
	if !strings.Contains(out, "断点 3 个已全部删除") {
		t.Errorf("没说清删了几条：\n%s", out)
	}
	if !strings.Contains(out, "断点 0 个") {
		t.Errorf("删完没有重列（下一次还会用过期编号）：\n%s", out)
	}
}

// ⑥ --all 带参数：矛盾要当场报，不要静默按 --all 执行。
func TestBPDelAllWithArgIsUsageError(t *testing.T) {
	newBPDelStub(t, bpRow(1, "axm_axmt500_wf.4gl", 4130))
	bpDelAllFlag = true
	if err := debugBPDelCmd.RunE(debugBPDelCmd, []string{"1"}); err == nil {
		t.Fatal("--all 与参数并存却执行了")
	}
}

// ⑦ bp files 是"位置该写成什么"的权威来源，要真的列出来。
func TestBPFilesListsLoadedSources(t *testing.T) {
	newBPDelStub(t)
	_, out := testkit.CaptureStdout(t, func() int {
		if err := debugBPFilesCmd.RunE(debugBPFilesCmd, nil); err != nil {
			t.Errorf("bp files 失败：%v", err)
		}
		return 0
	})
	for _, want := range []string{"已加载源文件 2 个", "axm_axmt500_wf.4gl", "csub_s_bsft001_wf.4gl"} {
		if !strings.Contains(out, want) {
			t.Errorf("输出里少了 %q：\n%s", want, out)
		}
	}
}

// ⑧ running 拒绝要带 interrupt 指引（与 bp add/list 同一套纪律）。
func TestBPDelRunningGivesInterruptHint(t *testing.T) {
	s := newBPDelStub(t, bpRow(1, "axm_axmt500_wf.4gl", 4130))
	s.reject = "程序运行中,禁止发送调试命令(仅允许中断)"
	err := debugBPDelCmd.RunE(debugBPDelCmd, []string{"1"})
	if err == nil {
		t.Fatal("running 时删断点却成功了")
	}
	if !strings.Contains(err.Error(), "tt debug interrupt") {
		t.Fatalf("没给 interrupt 指引：%v", err)
	}
}
