//go:build !windows

package winproc

// 非 Windows 的后台子进程：以 Setsid 脱离会话，输出重定向到日志文件，父进程退出后继续运行。
//
// C# 引擎本身只有 Windows 版（它 LoadFrom 的是 .NET Framework/WPF 的程序集），所以这条路径
// 实际只服务 `tt serve` 的后台常驻。留着是为了让 internal/cli/debug 不必再分平台维护一份。

import (
	"os"
	"os/exec"
	"syscall"
)

// SpawnDetached 启动一个脱离子进程（不等待），返回 pid。语义与 Windows 版一致。
func SpawnDetached(exe string, args []string, logPath string, extraEnv ...string) (int, error) {
	logf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, err
	}
	defer logf.Close()

	cmd := exec.Command(exe, args...)
	cmd.Stdout = logf
	cmd.Stderr = logf
	cmd.Stdin = nil
	cmd.Env = append(append(os.Environ(), "TT_SERVE_LOG="+logPath), extraEnv...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	go cmd.Wait()
	return cmd.Process.Pid, nil
}

// Alive 判断 pid 是否存活。
func Alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

// Kill 结束指定进程。
func Kill(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
