package fep

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/debug"
	"tt/internal/fep"
)

var (
	openProg     string
	openEnv      string
	openGDCExe   string
	openGDCPort  int
	openProxyPr  int
	openInjectPr int
	openFGLIP    string
	openNoJob    bool
	openWaitGDC  time.Duration
)

var sessionOpenCmd = &cobra.Command{
	Use:   "open",
	Short: "起一个隔离会话（隔离实例 + 代理 + 打开作业），前台常驻到收工",
	Long: `起一个隔离的作业会话，全程不打扰用户：不抢焦点、不碰键鼠、只在本机回环上说话。

它做四件事，顺序不能换：
  1. 环境体检：已经在跑的 gdc.exe 若与 TT 权限不一致，直接失败（那种情况下窗口驱动不可靠）；
  2. 端口体检：GDC / 代理 / 注入口三个端口都必须空着；
  3. 起隔离实例（gdc.exe -n -M -A 0 -p <port>）与本地代理（逐连接落原始字节）；
  4. 写会话登记（.tt-fep.json）——**先登记再开作业**，这样中途被打断也能收干净。

然后经 SSH 登录、选区域、下发 FGLSERVER=<本机>:<代理端口-6400> 并启动作业客户端。
作业一开着，本命令就前台常驻；Ctrl-C 或另开一个终端跑 tt fep session close 都能收工，
收工顺序是：远端作业 → 代理 → 隔离实例 → 清登记（只杀登记里的 pid）。

注意：作业窗口是**普通窗口**，用户看得见、也能自己操作；TT 不碰键鼠。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 用法错要在读配置、起任何东西之前就报出来
		if openProg == "" && !openNoJob {
			return fail(exitUsage, "没给作业程序（--prog axmt500_wf）")
		}
		cfgPath, err := common.ResolveConfig(true)
		if err != nil {
			return err
		}
		dcfg, err := debug.LoadConfig(cfgPath)
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		if openEnv != "" {
			c := dcfg.CloneEnv(openEnv)
			if c == nil {
				return fail(exitUsage, "环境 %q 不存在（见 tt env ls）", openEnv)
			}
			dcfg = c
		}
		loc := config.LocationsAt(cfgPath)
		recDir := loc.CacheDir("fep")
		if recDir == "" {
			return fail(exitPreflight, "缓存目录没解析出来（检查配置路径）")
		}
		if err := os.MkdirAll(recDir, 0o755); err != nil {
			return fail(exitFailed, "建录制目录失败: %v", err)
		}
		tag := "fep-" + time.Now().Format("20060102-150405")

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		logf := func(format string, a ...any) {
			fmt.Fprintf(os.Stderr, "· "+format+"\n", a...)
		}
		rt, err := fep.Open(ctx, fep.OpenConfig{
			Env:         dcfg.EnvName(),
			Zone:        dcfg.Zone,
			Prog:        openProg,
			SSH:         dcfg.SSH,
			GDCExe:      openGDCExe,
			GDCPort:     openGDCPort,
			ProxyPort:   openProxyPr,
			InjectPort:  openInjectPr,
			FGLServerIP: openFGLIP,
			DataDir:     loc.DataDir(),
			RecordDir:   recDir,
			Tag:         tag,
			GDCLog:      filepath.Join(recDir, tag+"-gdc.log"),
			WaitGDC:     openWaitGDC,
			NoJob:       openNoJob,
		}, logf)
		if err != nil {
			return mapOpenErr(err)
		}

		s := rt.Session
		if err := printJSONOrText(map[string]any{
			"prog": s.Prog, "env": s.Env, "zone": s.Zone,
			"gdcPid": s.GDCPID, "gdcPort": s.GDCPort,
			"proxyPort": s.ProxyPort, "injectPort": s.InjectPort,
			"tag": s.Tag, "registry": rt.RegistryPath(),
		}, func() {
			fmt.Printf("会话已起：%s（环境 %s / 区域 %s）\n", s.Prog, s.Env, s.Zone)
			fmt.Printf("  隔离实例 pid %d，端口 %d\n", s.GDCPID, s.GDCPort)
			fmt.Printf("  代理监听 %d（注入口 127.0.0.1:%d）\n", s.ProxyPort, s.InjectPort)
			fmt.Printf("  录制件 %s/%s-*-*.log\n", recDir, s.Tag)
			fmt.Println("收工：Ctrl-C，或另开终端跑 tt fep session close")
		}); err != nil {
			return err
		}

		waitErr := rt.Wait(ctx)
		closeErr := rt.Close()
		if waitErr != nil {
			return fail(exitFailed, "%v", waitErr)
		}
		if closeErr != nil {
			return fail(exitFailed, "%v", closeErr)
		}
		fmt.Fprintln(os.Stderr, "· 已收工，无残留")
		return nil
	},
}

// mapOpenErr 把内核的错误映射成这条线的退出码。
func mapOpenErr(err error) error {
	switch {
	case errors.Is(err, fep.ErrUsage):
		return fail(exitUsage, "%v", err)
	case errors.Is(err, fep.ErrPreflight):
		return fail(exitPreflight, "%v", err)
	case errors.Is(err, fep.ErrPortBusy):
		return fail(exitState, "%v", err)
	default:
		return fail(exitFailed, "%v", err)
	}
}

func init() {
	sessionOpenCmd.Flags().StringVar(&openProg, "prog", "", "作业程序名（必填，如 axmt500_wf）")
	sessionOpenCmd.Flags().StringVar(&openEnv, "env", "", "环境名（config.json hosts.sshs；默认取 activeEnv）")
	sessionOpenCmd.Flags().StringVar(&openGDCExe, "gdc-exe", "", "GDC 可执行文件（默认 C:\\Program Files\\FourJs\\gdc\\bin\\gdc.exe）")
	sessionOpenCmd.Flags().IntVar(&openGDCPort, "gdc-port", 0, "隔离实例监听端口（默认 7400）")
	sessionOpenCmd.Flags().IntVar(&openProxyPr, "proxy-port", 0, "代理监听端口（默认 7600；FGLSERVER 数值 = 它 - 6400）")
	sessionOpenCmd.Flags().IntVar(&openInjectPr, "inject-port", 0, "注入口端口（默认 7899）")
	sessionOpenCmd.Flags().StringVar(&openFGLIP, "fgserver-ip", "", "覆盖 FGLSERVER 里的本机地址（默认按路由探测）")
	sessionOpenCmd.Flags().BoolVar(&openNoJob, "no-job", false, "只起隔离实例与代理，不打开作业（排障用）")
	sessionOpenCmd.Flags().DurationVar(&openWaitGDC, "wait-gdc", 0, "等隔离实例监听的最长时间（默认 20s）")
	sessionCmd.AddCommand(sessionOpenCmd)
}
