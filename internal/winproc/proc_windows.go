//go:build windows

// Package winproc 只做三件事：起一个不随本进程死的子进程、判它活着、杀掉它。
//
// 单独成包是因为有两处要用：`tt serve` 把自己重新拉起来做后台常驻（internal/cli/debug），
// 和 `tt dev tzs` 托管 C# 引擎（internal/dev/tzs）。原来只有前者，实现就住在
// internal/cli/debug/servebg_windows.go 里；第二处出现时按「清掉重复」的纪律抽出来。
package winproc

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// SpawnDetached 启动一个脱离子进程（不等待），返回 pid。输出追加到 logPath，stdin 是空设备。
// extraEnv 追加在 os.Environ() 之后（同名后者胜），用来把工作区之类的设置传给子进程。
//
// 关于句柄继承 —— 这里**故意不做** C# 侧那种「扫全句柄表清 HANDLE_FLAG_INHERIT」的封印，
// 因为 Go 不需要：syscall.StartProcess 只要会继承句柄，就一定会设
// PROC_THREAD_ATTRIBUTE_HANDLE_LIST，把继承限制在 ProcAttr.Files 那三个上
// （exec_windows.go:420-427，注释原文 "Do not accidentally inherit more than these handles."）。
//
// 这一步在 C# 那边是必须的，而且踩得很深：Process.Start 无差别复制每一个可继承句柄，于是
// spawn 出来的守护进程拿到了调用方捕获 stdout 的那个管道写端，调用方（subprocess 的
// capture_output、CI runner、`| head`）永远等不到 EOF。只清 GetStdHandle(STD_*) 三个还不够——
// .NET 运行时在 Main 之前给自己复制了第二个可继承的 stdout 句柄，点不到它。
//
// 本包实测过这个差异（临时探针，两种模式都试了）：父进程以 capture_output 起一个 Go 进程，
// 后者 DETACHED_PROCESS 拉起真正的 tzs-server，父进程 0.38 s 拿到 EOF，不挂。
func SpawnDetached(exe string, args []string, logPath string, extraEnv ...string) (int, error) {
	logf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, err
	}
	defer logf.Close()

	cmd := exec.Command(exe, args...)
	cmd.Stdout = logf
	cmd.Stderr = logf
	cmd.Stdin = nil // os.DevNull：子进程不该等着谁的输入
	cmd.Env = append(append(os.Environ(), "TT_SERVE_LOG="+logPath), extraEnv...)

	// DETACHED_PROCESS(0x8) 脱离控制台；CREATE_NEW_PROCESS_GROUP(0x200) 独立进程组
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x8 | 0x200}
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	go cmd.Wait() // 回收句柄，不阻塞调用方
	return cmd.Process.Pid, nil
}

// Alive 判断 pid 是否存活。用于快速失败，**不用于身份判断**：pid 会被复用，
// 而且「进程活着」不等于「它在监听」。
func Alive(pid int) bool {
	if pid <= 0 {
		return false
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	return code == 259 // STILL_ACTIVE
}

// Kill 结束指定进程。
func Kill(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
