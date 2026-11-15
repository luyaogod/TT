package plugin

import (
	"os"
	"time"
)

// removeAllRetry 删一个目录，失败时短暂重试。
//
// 为什么需要重试（不是"多此一举"）：Windows 上进程**刚退出**时它的可执行文件句柄
// 可能还没释放（驱动/杀软扫描也会短暂持有），此时 RemoveAll 报
// "Access is denied"。用户的真实操作序列恰好是"停掉插件 → 立刻卸载"，
// 所以这条路径必须自己扛住这一小段窗口，而不是让人重试一次命令。
func removeAllRetry(path string) error {
	var err error
	for i := 0; i < 6; i++ {
		if err = os.RemoveAll(path); err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}
	return err
}
