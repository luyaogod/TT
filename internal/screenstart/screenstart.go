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

	"tt/internal/debug"
	"tt/internal/winproc"
)

// Request 起一块画面通道要的全部信息。
type Request struct {
	// CfgPath 是配置文件路径 —— 服务可能用 `--config` 指到别处，子进程必须跟着走，
	// 否则它会去读默认配置（挂到另一份配置的环境上）。
	CfgPath string
	// Ref 是挂到哪条会话（规范的 `环境#序号`）：少了它就变成"另起一条会话"。
	Ref string
	// Env 是冗余带上的环境名：它是幂等键的一部分（同环境同别名复用宿主）。
	Env string
	Ent string
	// Prog 是录哪个作业（必填：没有作业就没什么可录）。
	Prog string
	// Tag 是期望的录制 ID（目前 `recording start` 不支持显式指定，由进程自己生成）。
	Tag string
	// Actor 是"这次动作算谁的"（`X-Actor` 同一个口径，注入子进程的 `TT_ACTOR`）。
	//
	// 为什么必须传：这个子进程要用调试线的接口在**这条会话**上写（收掉上一轮、点 ENT、
	// 起作业），而调试线有模式闸门 —— 纯人工模式下 AI 不可写、协作模式下人不可写。
	// 人在界面上点「创建会话」时发起方是 human，子进程就得声明 human；AI 调 REST 时是 ai。
	// 传错的表现是一句莫名其妙的"纯人工模式:AI 不可操作"（实测 2026-10-18），
	// 而真正的原因是"谁在按这个按钮"没传下去。
	Actor string
}

// Starter 起一块画面通道并立刻返回（进程活着就算起成功）。
type Starter func(Request) (pid int, logPath string, err error)

// 几个可替换点：测试要能钉住"按什么参数起它"，而不真起一个跑 GDC 的进程。
var (
	executable     = os.Executable
	spawnDetached  = winproc.SpawnDetached
	processAlive   = winproc.Alive
	alivePollPause = 200 * time.Millisecond
	aliveDeadline  = 6 * time.Second
)

// New 返回生产实现。
func New() Starter {
	return func(req Request) (int, string, error) {
		exe, err := executable()
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
		args := []string{"recording", "start", "--session", req.Ref, "--prog", req.Prog}
		if e := strings.TrimSpace(req.Ent); e != "" {
			args = append(args, "--ent", e)
		}
		if e := strings.TrimSpace(req.Env); e != "" {
			args = append(args, "--env", e)
		}
		if c := strings.TrimSpace(req.CfgPath); c != "" {
			args = append(args, "--config", c)
		}
		// 空的 Tag 不再传：`recording start` 的录制 ID 由进程生成，调用方事后用
		// `GET /api/sessions/{id}/screen` 查回来。
		_ = req.Tag

		// 只有明确给了身份才注入：空 = 不设这个变量，子进程按 CLI 的缺省算 ai（既有行为不变）。
		var extraEnv []string
		if a := strings.TrimSpace(req.Actor); a != "" {
			extraEnv = append(extraEnv, debug.ActorEnv+"="+a)
		}

		pid, perr := spawnDetached(exe, args, logPath, extraEnv...)
		if perr != nil {
			return 0, logPath, fmt.Errorf("起画面通道进程失败: %w", perr)
		}
		// 起来就退 = 启动失败（配置错、环境不可达、GDC 不在…），把日志路径带出去。
		for waited := time.Duration(0); waited < aliveDeadline; waited += alivePollPause {
			time.Sleep(alivePollPause)
			if !processAlive(pid) {
				return pid, logPath, fmt.Errorf("画面通道进程起来就退了（pid %d）：详见 %s", pid, logPath)
			}
		}
		return pid, logPath, nil
	}
}
