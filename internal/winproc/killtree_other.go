//go:build !windows

package winproc

// KillTree 在非 Windows 上等价于 Kill：没有 taskkill，也没有 Windows 那种
// "孙进程继续占着端口"的同等麻烦（POSIX 侧由调用方用进程组/信号处理）。
func KillTree(pid int) error { return Kill(pid) }
