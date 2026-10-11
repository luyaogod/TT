package fep

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
	"tt/internal/winproc"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "会话：作业实例 + 代理 + 登记",
	Long: `一个会话 = 一个隔离的 GDC 实例 + 一条本地代理 + 一份登记（数据目录下的 .tt-fep.json）。

登记先写、清理后做：任何一步被打断，都能靠登记把这一轮起的东西收回，
而且只收登记的 pid（绝不扫名字杀进程）。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return fail(exitUsage, "未知子命令 %q", args[0])
	},
}

var sessionPsCmd = &cobra.Command{
	Use:   "ps",
	Short: "列出已登记的会话（含 GDC 是否还活着、权限是否对得上）",
	Args:  noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, _, err := registry()
		if err != nil {
			return err
		}
		s, err := reg.Load()
		if errors.Is(err, fep.ErrNoSession) {
			return fail(exitNoSession, "没有登记中的会话（%s 不存在）", reg.Path)
		}
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		alive := winproc.Alive(s.GDCPID)
		view := map[string]any{"session": s, "gdcAlive": alive, "registry": reg.Path}
		if e, eerr := fep.CheckElevation(s.GDCPID); eerr == nil {
			view["elevation"] = e
			if !e.OK {
				view["warning"] = e.Advice
			}
		}
		return printJSONOrText(view, func() {
			fmt.Printf("会话    : %s (环境 %s / 区域 %s)\n", s.Prog, s.Env, s.Zone)
			fmt.Printf("GDC     : pid %d 端口 %d  %s\n", s.GDCPID, s.GDCPort, aliveText(alive))
			fmt.Printf("代理    : 监听 %d → 127.0.0.1:%d（注入口 %d）\n", s.ProxyPort, s.GDCPort, s.InjectPort)
			fmt.Printf("录制件  : %s\n", s.Tag)
			fmt.Printf("起始    : %s（启动者 pid %d）\n", s.Started.Format(time.RFC3339), s.OwnerPID)
			if e, eerr := fep.CheckElevation(s.GDCPID); eerr == nil && !e.OK {
				fmt.Printf("⚠ %s\n", e.Advice)
			}
		})
	},
}

var sessionCloseCmd = &cobra.Command{
	Use:   "close",
	Short: "收掉已登记的会话（只杀登记的 pid），并清掉登记",
	Args:  noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, _, err := registry()
		if err != nil {
			return err
		}
		s, err := reg.Load()
		if errors.Is(err, fep.ErrNoSession) {
			return fail(exitNoSession, "没有登记中的会话（%s 不存在）", reg.Path)
		}
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		alive := winproc.Alive(s.GDCPID)
		stopped, serr := reg.Stop()
		if serr != nil {
			return fail(exitFailed, "%v", serr)
		}
		return printJSONOrText(map[string]any{"stopped": stopped.Prog, "gdcPid": stopped.GDCPID, "wasAlive": alive, "registryCleared": true}, func() {
			if alive {
				fmt.Printf("已收会话 %s：GDC pid %d 已结束，登记已清\n", stopped.Prog, stopped.GDCPID)
			} else {
				fmt.Printf("已收会话 %s：GDC pid %d 早已不在，登记已清\n", stopped.Prog, stopped.GDCPID)
			}
		})
	},
}

func aliveText(alive bool) string {
	if alive {
		return "运行中"
	}
	return "已不在（登记过期）"
}
