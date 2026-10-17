// Package screenstart 起一块**画面通道**：一个脱离本进程的 `tt recording start --session …`。
//
// # 为什么单独一个包
//
// 三层的分工是这样的：
//
//	internal/fep        画面通道**本身**（GDC + 代理 + 录制），只管自己那块；
//	internal/debug      调试面：不认识 GDC/代理，只收"起/停"两个函数（见 SetScreenStarter）；
//	本包                "把 tt 起起来"这件事 —— 谁需要画面通道就调它。
//
// 合在一起会怎样：调试服务得 import fep（形成 debug ↔ fep 耦合），或者 fep 得学会
// "如何把自己起成脱离进程"（那是进程管理，不是画面通道）。
package screenstart

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tt/internal/winproc"
)

// Starter 起一块画面通道并立刻返回（进程活着就算起成功）。
//
// 参数与 `tt recording start` 的挂靠模式一一对应；`cfgPath` 是配置文件路径 ——
// 服务可能用 `--config` 指到别处，子进程必须跟着走，否则它会去读默认配置。
type Starter func(cfgPath, ref, env, ent, prog, tag string) (pid int, logPath string, err error)

// New 返回生产实现。
func New() Starter {
	return func(cfgPath, ref, env, ent, prog, tag string) (int, string, error) {
		exe, err := os.Executable()
		if err != nil {
			return 0, "", err
		}
		dir, err := os.MkdirTemp("", "tt-screen-")
		if err != nil {
			return 0, "", err
		}
		logPath := filepath.Join(dir, "screen.log")

		// 只走挂靠模式：`--session <ref>` 保证画面挂到**被点名的会话**上，而不是另起一条。
		// `--env` 也带上：它是幂等键的一部分（同环境同别名复用宿主），少一个就可能落到别处。
		args := []string{"recording", "start", "--session", ref, "--prog", prog}
		if e := strings.TrimSpace(ent); e != "" {
			args = append(args, "--ent", e)
		}
		if e := strings.TrimSpace(env); e != "" {
			args = append(args, "--env", e)
		}
		if c := strings.TrimSpace(cfgPath); c != "" {
			args = append(args, "--config", c)
		}
		// `tag` 目前 `rec start` 不支持显式指定（录制 ID 由进程生成）—— 调用方拿到的
		// 是进程号与日志；真正的录制 ID 用 `GET /api/sessions/{id}/screen` 查回来。
		_ = tag

		pid, perr := winproc.SpawnDetached(exe, args, logPath)
		if perr != nil {
			return 0, logPath, fmt.Errorf("起画面通道进程失败: %w", perr)
		}
		// 起来就退 = 启动失败（配置错、环境不可达、GDC 不在…），把日志路径带出去。
		for waited := time.Duration(0); waited < 6*time.Second; waited += 200 * time.Millisecond {
			time.Sleep(200 * time.Millisecond)
			if !winproc.Alive(pid) {
				return pid, logPath, fmt.Errorf("画面通道进程起来就退了（pid %d）：详见 %s", pid, logPath)
			}
		}
		return pid, logPath, nil
	}
}
