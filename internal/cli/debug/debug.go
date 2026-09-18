package debug

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/debug"
)

var (
	dbgModule string
	dbgProg   string
	dbgLine   int
	dbgListen string
	dbEnt     int
	dbRefresh bool
)

// debugProbeCmd M0 尖刺:全自动跑通 登录→启动→下断点→步进→求值→退出
var debugProbeCmd = &cobra.Command{
	Use:   "probe",
	Short: "M0 尖刺:验证协议驱动器(全自动,不依赖 GDC 交互)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := resolveConfigPath()
		if err != nil {
			return err
		}
		cfg, err := debug.LoadConfig(cfgPath)
		if err != nil {
			return err
		}
		fmt.Printf("[probe] 服务器 %s:%d 区域 %s\n", cfg.SSH.Host, cfg.SSH.Port, cfg.Zone)

		sess, err := debug.NewSession(cfg, dbgModule, dbgProg, "", "", "", func(ev debug.Event) {
			switch ev.Type {
			case "state":
				fmt.Printf("[state] %s\n", ev.State)
			case "watchdog":
				fmt.Printf("[watchdog] %s\n", ev.Text)
			case "output":
				if common.Verbose {
					fmt.Printf("[out] %q\n", ev.Text)
				}
			}
		})
		if err != nil {
			return err
		}
		defer sess.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() { // Ctrl-C 优雅退出
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
			<-sig
			cancel()
		}()

		fmt.Println("[probe] 启动调试会话(Launch)...")
		if err := sess.Launch(ctx); err != nil {
			return fmt.Errorf("Launch 失败: %w", err)
		}
		fmt.Printf("[probe] 已到 (fgldb) 入口停站 ✓\n\n")

		fmt.Println("[probe] break main(函数断点):")
		bp, err := sess.Break("main")
		if err != nil {
			return err
		}
		fmt.Printf("  Breakpoint %d at %s:%d ✓\n", bp.Num, bp.File, bp.Line)

		fmt.Println("\n[probe] run(等待命中断点)...")
		if _, err := sess.Run(); err != nil {
			return err
		}
		stop, err := sess.WaitForStop(60 * time.Second)
		if err != nil {
			return err
		}
		fmt.Printf("  命中: %s:%d (%s) ✓\n", stop.File, stop.Line, stop.Func)
		for _, sl := range stop.Source {
			mark := "   "
			if sl.IsCur {
				mark = "-> "
			}
			fmt.Printf("  %s%-6d %s\n", mark, sl.Num, sl.Text)
		}

		fmt.Println("\n[probe] where:")
		frames, err := sess.Where()
		if err != nil {
			return err
		}
		for _, f := range frames {
			fmt.Printf("  #%d %s at %s:%d\n", f.Idx, f.Func, f.File, f.Line)
		}

		fmt.Println("\n[probe] next:")
		stop2, err := sess.Step("next")
		if err != nil {
			return err
		}
		if stop2 != nil {
			fmt.Printf("  -> %s:%d (%s)\n", stop2.File, stop2.Line, stop2.Func)
			for _, sl := range stop2.Source {
				mark := "   "
				if sl.IsCur {
					mark = "-> "
				}
				fmt.Printf("  %s%-6d %s\n", mark, sl.Num, sl.Text)
			}
		}

		fmt.Println("\n[probe] print num_args():")
		v, err := sess.Print("num_args()")
		fmt.Printf("  %s (err=%v)\n", v, err)

		fmt.Println("\n[probe] 全流程验证完成,退出会话...")
		return nil
	},
}

// debugServeCmd M1+:启动本地调试服务(REST+WS+Web 前端)。
// 默认在后台常驻(单实例):打印实际地址后立即返回,终端不被占用;
// 端口被占用时自动顺延到下一个空闲端口,并把真实地址写入状态文件供 debugctl 自动发现。
var debugServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "后台常驻调试服务(tt debug start/exec/status 等控制命令自动发现的守护进程)",
	Long: `后台常驻调试服务 —— tt debug 各条控制命令(start/exec/status/quit/wait/…)自动发现的守护进程。

它与 tt serve 是同一个服务的两种装法,服务端能力完全相同:
  · tt debug serve  只服务调试面:工作台 API 直接在根的 /api/… 下;
  · tt serve        统一服务:调试工作台 /debug/ + 统一设置页,调试面在 /debug/api/… 下。
两者默认都后台常驻、都写同一份状态文件(.tt-serve.json),控制端命令据此自动寻址
(各自的前缀也记在里面),所以 tt debug start/exec/status 对着哪个跑着的实例都能用。
同一个配置目录下只能起一个 —— 两条件会争同一份状态文件。

默认后台运行(单实例):命令打印服务地址与 pid 后立即返回,当前会话可继续输入其它命令;
已有一个实例在跑时打印它的地址后直接返回。
停止用 tt debug serve --stop;前台运行(日志直出终端)用 --foreground。
监听地址取 config.json 顶层 listen(兼容旧结构的 debug.listen;默认 127.0.0.1:28670,不常用端口);
端口被占用时自动顺延到下一个空闲端口并打印真实地址。`,
	Example: `  tt debug serve                  # 后台启动(单实例),打印地址后返回
  tt debug serve --stop          # 停止后台实例
  tt debug serve --foreground    # 前台运行,日志直出终端
  tt debug serve --listen 127.0.0.1:9123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if dbgServeStop {
			return debugServeStop()
		}
		cfgPath, err := resolveConfigPath()
		if err != nil {
			return err
		}
		cfg, err := debug.LoadConfig(cfgPath)
		if err != nil {
			return err
		}
		cfg.ApplyDefaultEnv() // 默认环境的连接/参数合并到顶层
		if dbgListen != "" {
			cfg.Listen = dbgListen
		}
		// 数据目录 = config.json 所在目录(断点持久化等)
		cfg.DataDir = filepath.Dir(cfgPath)
		if dbgServeForeground {
			return debugServeForeground(cfg, cfgPath)
		}
		return debugServeBackground(cfgPath)
	},
}

// debugDBCmd 数据库连接探查:登录服务器 → 查 gzou_t 企业→账号映射 → 尝试用对应账号连库
var debugDBCmd = &cobra.Command{
	Use:   "db",
	Short: "数据库连接探查:按企业(TOPENT)查账号并尝试连接",
	Long: `登录服务器,读取 gzou_t 中企业编号与数据库账号(schema)的映射,
并尝试用对应账号连接数据库(密码优先取该连接的账号清单,未收录按"账号=密码"惯例)。

  tt debug db            列出全部企业→账号映射
  tt debug db --ent 99   验证企业 99 对应账号的连接(默认企业取该 SSH 环境的 topent)
  tt debug db --ent 99 --json  输出 JSON

这条命令是**连接体检**:它会真的连一次库验证账号可用。
只要清单(不验证连接)用 tt debug ents —— 那份带快照,不联网也能答。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := resolveConfigPath()
		if err != nil {
			return err
		}
		cfg, err := debug.LoadConfig(cfgPath)
		if err != nil {
			return err
		}
		cfg.ApplyDefaultEnv() // 默认环境的连接/参数合并到运行时字段
		// 数据目录 = 配置所在目录:企业目录快照与 serve 侧落在同一份(见 ents.go)
		cfg.DataDir = filepath.Dir(cfgPath)
		ent := dbEnt
		if ent <= 0 {
			ent = cfg.TopentInt()
		}
		rep, err := debug.ProbeDB(cfg, ent, debug.EntListOpt{Ent: ent, Refresh: dbRefresh})
		if err != nil {
			return err
		}
		if IsJSON() {
			return printJSON(rep)
		}
		// 人类可读输出
		if kdb := rep.Env["database"]; kdb != "" {
			fmt.Printf("区域: %s   数据库: 人大金仓 %s@%s:%s\n", rep.Zone, kdb, rep.Env["host"], rep.Env["port"])
			if ks := rep.Env["ksql"]; ks != "" {
				fmt.Printf("ksql: %s\n", ks)
			}
		} else {
			fmt.Printf("区域: %s   Oracle: %s:%s (service %s)\n",
				rep.Zone, rep.Env["host"], rep.Env["port"], rep.TNS)
			if sp := rep.Env["sqlplus"]; sp != "" {
				fmt.Printf("sqlplus: %s\n", sp)
			}
			if oh := rep.Env["oracleHome"]; oh != "" {
				fmt.Printf("ORACLE_HOME: %s\n", oh)
			}
		}
		fmt.Printf("\n企业(TOPENT) → 账号(schema),共 %d 个:\n", len(rep.Mappings))
		for _, m := range rep.Mappings {
			mark := " "
			if rep.Probe != nil && m.Ent == rep.Probe.Ent {
				mark = "*"
			}
			fmt.Printf("  %s %-6d → %s\n", mark, m.Ent, m.Account)
		}
		if rep.Probe != nil {
			p := rep.Probe
			fmt.Printf("\n企业 %d 连接验证: %s@%s\n", p.Ent, p.Account, rep.TNS)
			if p.Host != "" {
				fmt.Printf("  主机: %s:%s  服务名: %s\n", p.Host, p.Port, p.Service)
			}
			if p.Connect {
				fmt.Println("  结果: 连接成功 ✓")
			} else {
				fmt.Printf("  结果: 连接失败 ✗  %s\n", p.Error)
			}
			fmt.Println("  (密码:优先取该连接的账号清单(accounts/主账号),未收录按 账号=密码 惯例;若失败请在 设置-环境-DB 页的账号清单维护密码)")
		} else {
			fmt.Println("\n(未指定企业,仅列出映射。用 --ent <企业号> 验证连接)")
		}
		return nil
	},
}

// printJSON 以 JSON 输出结果(统一走 common.PrintJSON,格式与根命令一致)
func printJSON(v any) error { return common.PrintJSON(v) }

func init() {
	// --module 原为 debug 父命令的持久 flag,扁平化后只挂在真正读取它的命令上
	debugProbeCmd.Flags().StringVarP(&dbgModule, "module", "m", "asf", "T100 模块目录名(如 asf)")
	debugProbeCmd.Flags().StringVarP(&dbgProg, "prog", "p", "bsft001_wf", "作业名(如 bsft001_wf)")
	debugProbeCmd.Flags().IntVarP(&dbgLine, "line", "l", 4452, "探针断点行号")
	debugServeCmd.Flags().StringVar(&dbgListen, "listen", "", "覆盖监听地址(默认取配置)")
	debugServeCmd.Flags().BoolVar(&dbgServeForeground, "foreground", false, "前台运行,日志直出终端(默认后台常驻)")
	debugServeCmd.Flags().BoolVar(&dbgServeStop, "stop", false, "停止后台运行的调试服务(单实例)")
	debugDBCmd.Flags().IntVar(&dbEnt, "ent", 0, "企业编号(TOPENT),验证该企业账号连接;0=取该环境的 topent")
	debugDBCmd.Flags().BoolVar(&dbRefresh, "refresh", false, "跳过企业目录缓存与快照,强制现查 gzou_t")

	Group.AddCommand(debugProbeCmd, debugServeCmd, debugDBCmd)
}
