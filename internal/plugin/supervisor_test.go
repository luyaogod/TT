package plugin

// P3-14 停止前实例身份核对：PID 会被复用，"活着"不能作为"可以杀"的依据。
//
// 这一组用例分三层：
//   - shouldKill / StaleReason 是**纯决策**：把"记录的身份 vs 现在的身份"摊成表，
//     不碰进程表，所以能把"路径不同 / 摘要不同 / 只有一边有证据"逐条钉住；
//   - ReconcileInstanceState 在真 Store 上验"陈旧状态文件会被清掉、仍然有效的不会被碰"；
//   - Stop 用**真进程**验最危险的那条：身份不符时一个字节都不许动那个进程。

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/hostrpc"
	"github.com/luyaogod/TT/internal/winproc"
)

// TestShouldKillIdentity 纯决策：什么情况下允许结束这个 PID。
func TestShouldKillIdentity(t *testing.T) {
	const dig = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	cases := []struct {
		name     string
		recorded exeIdentity
		current  exeIdentity
		want     bool
	}{
		{
			name:     "路径与摘要都对得上",
			recorded: exeIdentity{Path: `C:\plugins\p\p.exe`, Digest: dig},
			current:  exeIdentity{Path: `C:\plugins\p\p.exe`, Digest: dig},
			want:     true,
		},
		{
			name:     "路径大小写不同仍是同一个文件（Windows 路径不区分大小写）",
			recorded: exeIdentity{Path: `C:\Plugins\P\p.exe`, Digest: dig},
			current:  exeIdentity{Path: `c:\plugins\p\P.EXE`, Digest: strings.ToUpper(dig)},
			want:     true,
		},
		{
			name:     "路径不同 = PID 被复用",
			recorded: exeIdentity{Path: `C:\plugins\p\p.exe`, Digest: dig},
			current:  exeIdentity{Path: `C:\Windows\notepad.exe`, Digest: dig},
			want:     false,
		},
		{
			name:     "路径相同但文件已被换掉（摘要不同）",
			recorded: exeIdentity{Path: `C:\plugins\p\p.exe`, Digest: dig},
			current:  exeIdentity{Path: `C:\plugins\p\p.exe`, Digest: strings.Repeat("b", 64)},
			want:     false,
		},
		{
			name:     "旧状态文件没记路径：退回比摘要",
			recorded: exeIdentity{Digest: dig},
			current:  exeIdentity{Path: `C:\plugins\p\p.exe`, Digest: dig},
			want:     true,
		},
		{
			name:     "旧状态文件没记路径且摘要对不上：仍然不放行",
			recorded: exeIdentity{Digest: dig},
			current:  exeIdentity{Path: `C:\plugins\p\p.exe`, Digest: strings.Repeat("b", 64)},
			want:     false,
		},
		{
			name:     "拿不到现在的身份（非 Windows / 打不开进程）：无反证就放行",
			recorded: exeIdentity{Path: `C:\plugins\p\p.exe`, Digest: dig},
			current:  exeIdentity{},
			want:     true,
		},
		{
			name:     "记录里什么都没有：放行（旧版本留下的实例记录）",
			recorded: exeIdentity{},
			current:  exeIdentity{Path: `C:\Windows\notepad.exe`, Digest: dig},
			want:     true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := shouldKill(c.recorded, c.current); got != c.want {
				t.Fatalf("shouldKill(%+v, %+v) = %v，期望 %v", c.recorded, c.current, got, c.want)
			}
		})
	}
}

// TestInstanceStateStaleReason 状态文件的三种结局：仍有效 / 进程没了 / PID 被复用。
func TestInstanceStateStaleReason(t *testing.T) {
	const dig = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	st := &InstanceState{PluginID: "p", PID: 42, State: "running", ExePath: `C:\plugins\p\p.exe`, ExeDigest: dig}
	same := exeIdentity{Path: `c:\PLUGINS\p\P.EXE`, Digest: strings.ToUpper(dig)}

	if got := st.StaleReason(true, same); got != "" {
		t.Fatalf("进程与身份都对得上时该判有效，得 %q", got)
	}
	if got := st.StaleReason(false, same); !strings.Contains(got, "不在了") {
		t.Fatalf("进程不在该说清楚，得 %q", got)
	}
	if got := st.StaleReason(true, exeIdentity{Path: `C:\Windows\notepad.exe`}); !strings.Contains(got, "复用") {
		t.Fatalf("PID 指向别的程序该判复用，得 %q", got)
	}
	// 旧版本写的状态文件没有 exePath/exeDigest：进程活着且拿不到反证时仍然有效。
	old := &InstanceState{PID: 42, State: "running"}
	if got := old.StaleReason(true, exeIdentity{Path: `C:\Windows\notepad.exe`}); got != "" {
		t.Fatalf("没有身份证据的旧状态文件不该被误判为陈旧，得 %q", got)
	}
	// 空 PID 一定是坏状态。
	if got := (&InstanceState{}).StaleReason(true, same); !strings.Contains(got, "PID") {
		t.Fatalf("空 PID 该判坏状态，得 %q", got)
	}
}

// TestHelperHoldProcess 不是用例：被身份核对用例当成"另一个进程"起。
//
// 用测试二进制自身当容器（跨平台，不依赖 cmd/ping 之类的外部程序）；环境变量没设
// 就立刻返回，所以正常运行整套测试时它什么也不做。
func TestHelperHoldProcess(t *testing.T) {
	if os.Getenv("TT_TEST_HOLD_PROCESS") != "1" {
		return
	}
	time.Sleep(30 * time.Second)
}

// holdProcess 起一个活着但不做任何事的进程。
func holdProcess(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHoldProcess$")
	cmd.Env = append(os.Environ(), "TT_TEST_HOLD_PROCESS=1")
	if err := cmd.Start(); err != nil {
		t.Fatalf("起不了占位进程: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	return cmd
}

// finishedPID 返回一个"已经退掉"的进程 PID（跨平台：仍然用测试二进制自身）。
func finishedPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	if err := cmd.Run(); err != nil {
		t.Fatalf("起不了立刻退出的探针进程: %v", err)
	}
	return cmd.Process.Pid
}

// writeInstanceStateFile 手写一份 run/<id>/instance.json。
func writeInstanceStateFile(t *testing.T, s *Store, st *InstanceState) string {
	t.Helper()
	path := filepath.Join(s.RunDir(st.PluginID), "instance.json")
	b, err := json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, string(b))
	return path
}

// TestReconcileClearsStaleInstanceState 起服务前清掉陈旧状态：进程没了 / PID 被复用都要清。
func TestReconcileClearsStaleInstanceState(t *testing.T) {
	_, s := newManager(t)
	sup := NewSupervisor(s, hostrpc.NewTokenStore(), nil, "", "test")
	rec := &Record{ID: "com.example.demo", ActiveVersion: "0.1.0"}

	// 1) 进程已经不在了。
	dead := &InstanceState{
		PluginID: rec.ID, Version: "0.1.0", InstanceID: "svc-old", PID: finishedPID(t),
		State: "running", ExePath: `C:\plugins\old\p.exe`, ExeDigest: strings.Repeat("a", 64),
	}
	path := writeInstanceStateFile(t, s, dead)
	got, reason, err := sup.ReconcileInstanceState(rec)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.InstanceID != "svc-old" || !strings.Contains(reason, "不在了") {
		t.Fatalf("该报告并清掉死进程的状态，得 st=%+v reason=%q", got, reason)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("陈旧状态文件该被删掉，stat: %v", err)
	}
	// 已经清过一次：再核对就是"没有状态文件"。
	if got, reason, err := sup.ReconcileInstanceState(rec); err != nil || got != nil || reason != "" {
		t.Fatalf("清过之后该返回空，得 st=%+v reason=%q err=%v", got, reason, err)
	}

	// 2) PID 活着，但已经指向另一个可执行文件（PID 被复用）。
	hold := holdProcess(t)
	impostor := filepath.Join(t.TempDir(), "somebody-else.exe")
	sup.ExecutablePath = func(int) string { return impostor }
	reused := &InstanceState{
		PluginID: rec.ID, Version: "0.1.0", InstanceID: "svc-reused", PID: hold.Process.Pid,
		State: "running", ExePath: filepath.Join(t.TempDir(), "refplugin.exe"), ExeDigest: strings.Repeat("b", 64),
	}
	writeInstanceStateFile(t, s, reused)
	got, reason, err = sup.ReconcileInstanceState(rec)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || !strings.Contains(reason, "复用") {
		t.Fatalf("PID 被复用该报告并清掉，得 st=%+v reason=%q", got, reason)
	}
	if st, err := s.ReadInstanceState(rec.ID); err != nil {
		t.Fatal(err)
	} else if st != nil {
		t.Fatalf("复用状态文件该被删掉，得 %+v", st)
	}
	if !winproc.Alive(hold.Process.Pid) {
		t.Fatal("核对状态不该碰这个进程")
	}
}

// TestSupervisorStopLeavesReusedPIDAlone 身份不符时 Stop 绝不能杀那个 PID。
func TestSupervisorStopLeavesReusedPIDAlone(t *testing.T) {
	_, s := newManager(t)
	tokens := hostrpc.NewTokenStore()
	sup := NewSupervisor(s, tokens, nil, "", "test")
	sup.Logf = t.Logf
	sup.StopGrace = 200 * time.Millisecond
	sup.ExecutablePath = func(int) string { return filepath.Join(t.TempDir(), "somebody-else.exe") }

	hold := holdProcess(t)
	token := tokens.Issue(connsvc.Caller{Kind: connsvc.CallerServicePlugin, PluginID: "com.example.svc", InstanceID: "svc-x"})
	inst := &Instance{
		PluginID: "com.example.svc", Version: "0.1.0", InstanceID: "svc-x",
		PID: hold.Process.Pid, State: "running", StartedAt: time.Now(),
		ExePath: filepath.Join(t.TempDir(), "refplugin.exe"), token: token, cmd: hold,
	}
	sup.mu.Lock()
	sup.instances[inst.PluginID] = inst
	sup.mu.Unlock()

	if err := sup.Stop(inst.PluginID); err != nil {
		t.Fatalf("停止不该报错（只是不动它）: %v", err)
	}
	if !winproc.Alive(hold.Process.Pid) {
		t.Fatal("身份不符时把别人的进程杀了（PID 复用保护失效）")
	}
	if tokens.Count() != 0 {
		t.Errorf("撤销令牌与杀进程是两件事：令牌该撤销，剩 %d", tokens.Count())
	}
	st, err := s.ReadInstanceState(inst.PluginID)
	if err != nil {
		t.Fatal(err)
	}
	if st == nil || st.State != "stopped" {
		t.Fatalf("停止后状态文件不该还说 running，得 %+v", st)
	}
}
