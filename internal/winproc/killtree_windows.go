//go:build windows

package winproc

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
)

// KillTree 结束一个进程**及其派生进程**（Windows）。
//
// 为什么必须连树一起杀：插件是第三方程序，它可以再拉起自己的辅助进程。只杀直接子进程时，
// 孙进程会继续活着、继续占着端口/文件，下一次启动就撞上"端口被占用" —— 而 tt 这边看起来
// 是"已经停了"。
//
// 用 taskkill /T（这是一棵树的官方口径）而不是自己遍历父进程链：父链在进程退出后会断，
// 遍历本身有竞态；而且 Windows 没有 POSIX 那样的进程组语义。
// taskkill 不可用时退回单进程 Kill，并把两层错误都带上。
func KillTree(pid int) error {
	if pid <= 0 {
		return nil
	}
	if !Alive(pid) {
		return nil // 已经不在了：收尾要的是"它别再挡路"
	}
	tk := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid))
	tk.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow}
	if out, err := tk.CombinedOutput(); err == nil {
		return nil
	} else if kerr := Kill(pid); kerr == nil {
		// 树没杀成但单进程杀成了：如实说明"树没清干净"，别假装成功。
		return fmt.Errorf("结束进程树失败（taskkill: %v，输出: %s），已结束主进程 %d —— 它派生的进程可能还在",
			err, trimOutput(out), pid)
	}
	return fmt.Errorf("结束进程树失败（taskkill 与 Kill 都没成功）")
}

// trimOutput 把 taskkill 的输出裁到能贴进一行错误里的长度。
func trimOutput(b []byte) string {
	s := string(b)
	if len(s) > 300 {
		return s[:300] + "…"
	}
	return s
}
