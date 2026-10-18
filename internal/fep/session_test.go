package fep

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRegistryRoundTripIsAtomic(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry(dir)
	s := &Session{Env: "主机109", Zone: "36", Prog: "axmt500_wf", GDCPID: 4242, GDCPort: 7400, ProxyPort: 7600, InjectPort: 7899, Tag: "fep", Started: time.Now()}
	if err := r.Save(s); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(r.Path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temp file left behind: the write is not atomic")
	}
	got, err := r.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Prog != s.Prog || got.GDCPID != s.GDCPID || got.InjectPort != s.InjectPort {
		t.Fatalf("round trip mismatch: %+v", got)
	}
	if err := r.Clear(); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, err := r.Load(); !errors.Is(err, ErrNoSession) {
		t.Fatalf("after clear: %v", err)
	}
}

func TestRegistryReportsCorruption(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry(dir)
	if err := os.WriteFile(r.Path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := r.Load()
	if err == nil || errors.Is(err, ErrNoSession) {
		t.Fatalf("corrupt registry must be reported, got %v", err)
	}
}

func TestSessionPIDsOnlyListsRegistered(t *testing.T) {
	s := &Session{GDCPID: 7}
	pids := s.PIDs()
	if len(pids) != 1 || pids[0] != 7 {
		t.Fatalf("PIDs = %v", pids)
	}
	// Cleanup must never touch anything but the registered process.
	if (&Session{}).PIDs() != nil {
		t.Fatal("empty session must own nothing")
	}
}

func TestRegistryStopClearsEvenWhenProcessIsGone(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry(dir)
	if err := r.Save(&Session{Tag: "fep-test", GDCPID: 0x7fffffff}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, SessionFileName)); !os.IsNotExist(err) {
		t.Fatal("registry not cleared")
	}
}

// ---------------------------------------------------------------------------
// 多会话（P1）：登记从"一份"变成"多份"，一条都不能丢
// ---------------------------------------------------------------------------

// TestRegistryListsMultipleSessions：两条登记互不覆盖。
func TestRegistryListsMultipleSessions(t *testing.T) {
	r := NewRegistry(t.TempDir())
	for _, tag := range []string{"fep-a", "fep-b"} {
		if err := r.Save(&Session{Tag: tag, Prog: tag, GDCPID: 0x7fffffff}); err != nil {
			t.Fatalf("save %s: %v", tag, err)
		}
	}

	list, err := r.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("两条登记应都在，实得 %d 条", len(list))
	}
	got, err := r.LoadTag("fep-b")
	if err != nil || got.Prog != "fep-b" {
		t.Fatalf("按 tag 取 fep-b 失败：%v %+v", err, got)
	}
}

// TestRegistrySaveUpsertsByTag：同一个 Tag 再存一次是**替换**，不是追加。
func TestRegistrySaveUpsertsByTag(t *testing.T) {
	r := NewRegistry(t.TempDir())
	if err := r.Save(&Session{Tag: "fep-a", Prog: "old"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Save(&Session{Tag: "fep-a", Prog: "new"}); err != nil {
		t.Fatal(err)
	}
	list, err := r.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Prog != "new" {
		t.Fatalf("同 Tag 应替换成一条，实得 %+v", list)
	}
}

// TestRegistrySavePrunesDeadEntries：Save 顺手清掉**已经死了**的登记。
//
// 不行的后果（真机 2026-10-18）：上一轮回放进程已退、登记还在，新一轮 Save 后
// 登记里两条 → 任何 `Load()` 都报"登记里有多条会话"，新一轮一个事件都发不出去。
func TestRegistrySavePrunesDeadEntries(t *testing.T) {
	r := NewRegistry(t.TempDir())
	dead := &Session{Tag: "fep-dead", OwnerPID: 999999} // 这个 pid 不该存在
	if err := r.Save(dead); err != nil {
		t.Fatal(err)
	}
	if err := r.Save(&Session{Tag: "fep-live", OwnerPID: os.Getpid()}); err != nil {
		t.Fatal(err)
	}
	got, err := r.Load()
	if err != nil {
		t.Fatalf("死登记该被清掉、只剩活的那条，实得 %v", err)
	}
	if got.Tag != "fep-live" {
		t.Fatalf("该剩 fep-live，实得 %q", got.Tag)
	}
}

// TestRegistryLoadIsAmbiguousWhenMultiple：多条时 `Load()` **不挑**，报歧义并列出 tag。
//
// 这是多会话之后最容易踩的一处：`tt fep do` 这类命令不带目标时会走 `Load()`，
// 若它悄悄返回第一条，动作就打在了别人的界面上。
func TestRegistryLoadIsAmbiguousWhenMultiple(t *testing.T) {
	r := NewRegistry(t.TempDir())
	for _, tag := range []string{"fep-a", "fep-b"} {
		if err := r.Save(&Session{Tag: tag}); err != nil {
			t.Fatal(err)
		}
	}
	_, err := r.Load()
	var amb *ErrAmbiguous
	if !errors.As(err, &amb) {
		t.Fatalf("多条时应报 ErrAmbiguous，实得 %T: %v", err, err)
	}
	if len(amb.Tags) != 2 {
		t.Fatalf("歧义要列出全部 tag，实得 %v", amb.Tags)
	}
}

// TestRegistryReadsLegacySingleObject：旧格式（单个对象）仍可读。
//
// 升级不该要求用户先手工删掉 .tt-fep.json —— 一份旧登记要在新版本里照常可用。
func TestRegistryReadsLegacySingleObject(t *testing.T) {
	r := NewRegistry(t.TempDir())
	legacy := `{"env":"主机109","zone":"36","prog":"axmt500_wf","gdcPid":4242,"tag":"old"}`
	if err := os.WriteFile(r.Path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := r.Load()
	if err != nil {
		t.Fatalf("旧格式应能读：%v", err)
	}
	if got.Tag != "old" || got.Prog != "axmt500_wf" || got.GDCPID != 4242 {
		t.Fatalf("旧格式读出来的东西不对：%+v", got)
	}
}

// TestRegistryStopTagOnlyRemovesThatOne：收一条不许动别人的登记。
func TestRegistryStopTagOnlyRemovesThatOne(t *testing.T) {
	r := NewRegistry(t.TempDir())
	for _, tag := range []string{"fep-a", "fep-b"} {
		if err := r.Save(&Session{Tag: tag, GDCPID: 0x7fffffff}); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := r.StopTag("fep-a"); err != nil {
		t.Fatalf("stop a: %v", err)
	}
	list, err := r.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Tag != "fep-b" {
		t.Fatalf("只该剩 fep-b，实得 %+v", list)
	}
}

// TestRegistrySaveRequiresTag：没有 Tag 的登记无法定位，存的时候就该拒。
func TestRegistrySaveRequiresTag(t *testing.T) {
	r := NewRegistry(t.TempDir())
	if err := r.Save(&Session{Prog: "无标签"}); err == nil {
		t.Fatal("缺 Tag 的登记应被拒（否则多条之后无法定位）")
	}
}

// TestRegistryPersistsSessionRef：登记要记住"这一轮落在调试线哪条会话上"（P3）。
//
// 两条线原本各记一套名字（这里 Tag、那边 ref），人得自己脑补对应关系；
// 收尾只能靠"用录制 ID 反推别名"去猜。Ref 进登记之后，两边指向同一个地址。
func TestRegistryPersistsSessionRef(t *testing.T) {
	r := NewRegistry(t.TempDir())
	if err := r.Save(&Session{Tag: "fep-r1", Prog: "bsft001_wf", Ref: "主机109#1"}); err != nil {
		t.Fatal(err)
	}
	got, err := r.LoadTag("fep-r1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Ref != "主机109#1" {
		t.Fatalf("ref 没留住：%q", got.Ref)
	}

	// 老登记（没有 ref）照旧能读：不许因为少一个键就当作坏文件。
	if err := r.Save(&Session{Tag: "fep-old", Prog: "p"}); err != nil {
		t.Fatal(err)
	}
	old, err := r.LoadTag("fep-old")
	if err != nil || old.Ref != "" {
		t.Fatalf("无 ref 的老登记该照常读：%v %+v", err, old)
	}
}

// TestScreenLookupMatchesByRef：按会话地址（ref）找画面通道；没 ref 的老登记不参与。
//
// 宁可不显示，也不能把别人的画面挂到这条会话上 —— 这个查询的返回会被 UI 当成"这条会话
// 现在挂着的画面"。
func TestScreenLookupMatchesByRef(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(dir)
	if err := reg.Save(&Session{Tag: "fep-a", Ref: "主机109#1", Prog: "axmt500_wf", GDCPID: 4242, ProxyPort: 7600, InjectPort: 7899}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Save(&Session{Tag: "fep-old", Prog: "old"}); err != nil { // 没有 ref 的老登记
		t.Fatal(err)
	}
	lookup := ScreenLookup(dir)

	got, ok := lookup("主机109#1")
	if !ok {
		t.Fatal("该按 ref 找到")
	}
	if got["tag"] != "fep-a" || got["proxyPort"] != 7600 {
		t.Fatalf("查出来的东西不对：%+v", got)
	}

	if _, ok := lookup("主机109#2"); ok {
		t.Fatal("不存在的地址不该命中")
	}
	if _, ok := lookup(""); ok {
		t.Fatal("空地址不该命中（那等于'谁都算'）")
	}
	if _, ok := lookup(""); ok {
		t.Fatal("没有 ref 的老登记不该被任何地址命中")
	}
}
