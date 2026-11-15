//go:build windows

package winproc

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
)

// ExecutablePath：对**本进程**必须取到与 os.Executable() 同一个文件，
// 对已经不存在的 pid 必须老老实实返回空串（而不是一个猜的路径）。
//
// 为什么值得钉住：internal/plugin 停止服务前用它核对 pid 身份；如果它悄悄返回空串，
// 核对会退化成"无反证就放行"，PID 复用保护等于没有；如果它对不存在的 pid 返回别的
// 东西，又会把正常收工判成"被复用"而漏杀自己的子进程。
func TestExecutablePathMatchesCurrentImage(t *testing.T) {
	got := ExecutablePath(os.Getpid())
	if got == "" {
		t.Fatal("拿不到本进程的镜像路径（停止前的身份核对会因此形同虚设）")
	}
	want, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	gs, gerr := os.Stat(got)
	ws, werr := os.Stat(want)
	if gerr != nil || werr != nil {
		t.Fatalf("路径不可读: got=%q(%v) want=%q(%v)", got, gerr, want, werr)
	}
	if !os.SameFile(gs, ws) {
		t.Fatalf("ExecutablePath(%d) = %q，不是本进程的 %q", os.Getpid(), got, want)
	}

	// 起一个立刻退出、并已经 Wait 过的进程：它的 pid 一定不指向任何活进程。
	dead := exec.Command("cmd", "/c", "exit")
	if err := dead.Run(); err != nil {
		t.Fatalf("探针进程没跑成: %v", err)
	}
	if p := ExecutablePath(dead.Process.Pid); p != "" {
		t.Fatalf("已退出的 pid %d 该返回空串，得 %q", dead.Process.Pid, p)
	}
}

// HideConsole：父进程没有控制台时给子进程加 CREATE_NO_WINDOW；有控制台时一个字都不动。
//
// 为什么值得钉住（用户实测）：点一下面板的界面按钮会弹出一大堆终端窗口/标签页 ——
// `tt serve` 起的那条子进程链路里，`tt debug start` 是 DETACHED_PROCESS（没有控制台），
// 而它内部 runSelf 起的是一串控制台程序：父进程没有控制台时 Windows 会给**每一个**
// 子进程新分配一个控制台，于是每起一个子进程就多一个窗口。
//
// 两条断言：只在"没有控制台"时加这个位；而且**不覆盖**调用方已经设好的位
// （这条链路上别的调用方可能已经设了 DETACHED_PROCESS/CREATE_NEW_PROCESS_GROUP）。
func TestHideConsoleOnlyWhenNoConsole(t *testing.T) {
	const preset = 0x8 // 假装调用方已经设了 DETACHED_PROCESS
	c := exec.Command("cmd", "/c", "exit")
	c.SysProcAttr = &syscall.SysProcAttr{CreationFlags: preset}

	HideConsole(c)
	t.Logf("HasConsole=%v（这个值决定走哪条分支：go test 由谁起、有没有终端）", HasConsole())

	if c.SysProcAttr == nil {
		t.Fatal("HideConsole 不该把 SysProcAttr 抹成 nil")
	}
	got := c.SysProcAttr.CreationFlags
	if got&preset == 0 {
		t.Fatalf("不该覆盖调用方已设的标志位：%#x", got)
	}
	if HasConsole() {
		if got != preset {
			t.Fatalf("父进程有控制台时不该动 CreationFlags（子进程该照旧继承同一个终端）：%#x", got)
		}
		return
	}
	if got&createNoWindow == 0 {
		t.Fatalf("父进程没有控制台时该加 CREATE_NO_WINDOW（否则每个子进程都会弹一个窗口）：%#x", got)
	}
}
