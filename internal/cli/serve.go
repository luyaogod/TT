package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	debugcli "tt/internal/cli/debug"
	"tt/internal/config"
	"tt/internal/debug"
	dictserver "tt/internal/dict/server"
	"tt/internal/web"
)

// newServeCmd 是统一的本地 Web 服务。
//
// 合并前每个工具各有一个 serve：tdebug 起调试界面、tdict 起配置页，各占一个端口、
// 各读自己那份配置。现在是一个进程、一个端口、一份配置、一套页面：
//
//	/debug/  调试工作台（含统一设置页：站点管理 / 数据字典 / DEBUG / 应用设置）
//	/api/*   统一接口 —— 环境与数据库配置、配置派生状态，以及字典类动作
//	         （源码镜像拉取、字典同步、BDL 文档）
//
// 默认后台常驻（单实例）：把自身以 --foreground 重新拉起来后立即返回，终端不被占住；
// 前台运行（看实时日志）用 --foreground，停止用 --stop。运行状态写在与调试服务同一份
// 状态文件里（.tt-serve.json），所以 `tt debug status/start/exec` 对着哪个实例都能用。
func newServeCmd() *cobra.Command {
	var (
		listen           string
		stop             bool
		foreground       bool
		readyJSON        bool
		exitOnStdinClose bool
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "启动本地 Web 服务（工作站 + 统一设置页）",
		Long: `启动统一的本地 Web 服务。

一个进程提供工作站与统一设置页：

  /debug/          工作站：左边是「环境 → 会话」两级目录，右边是所选目标的面板
  /debug/#env/<环境名>/create       环境页：创建会话 / 接口日志 / 接口调试
  /debug/#session/<会话 ID>/debug   会话页：Debug
  /debug/#settings                  统一设置页：站点管理 / 数据字典 / DEBUG / 应用设置
                   —— 环境与数据库、查询数据源、源码镜像、字典同步、BDL 文档、
                      调试参数、明暗色，全在这里；三个工具共用同一份配置。

端口默认取配置里的 listen（缺省 127.0.0.1:28670），被占用时自动向后顺延。

默认后台常驻（单实例）：打印实际地址后立即返回，终端可以接着敲别的命令 ——
` + "`tt debug start/exec/status`" + ` 等控制命令会自动找到它。已在运行时打印它的地址后返回。

  --foreground  前台运行，日志直出终端（Ctrl+C 停止）
  --stop        停止后台实例

桌面壳（TT Desktop）用的两个开关，都是给进程间契约用的，命令行日常不需要：

  --ready-json           就绪时向 stdout 打**一行** JSON：{"ok":true,"attached":…,
                         "url":…,"pid":…,"config":…}。已有实例时输出 attached=true
                         并直接退出 0（不起第二个）。
  --exit-on-stdin-close  stdin 关闭即优雅退出。壳被强杀时管道会关，服务跟着收尾，
                         不留一个隐形进程占着端口与状态文件。

同一个配置目录下只能有一个实例：两条进程会争同一份状态文件（--stop 与自动寻址的依据）。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 两个桌面壳开关都只对前台运行有意义：后台模式由子进程去就绪，
			// 父进程既拿不到真实地址（端口会顺延），也不会替子进程收尾。
			if readyJSON && !foreground {
				return fmt.Errorf("--ready-json 只能与 --foreground 一起用：后台模式由子进程就绪，父进程拿到的不是真实地址")
			}
			if exitOnStdinClose && !foreground {
				return fmt.Errorf("--exit-on-stdin-close 只能与 --foreground 一起用")
			}

			path, err := common.ResolveConfig(true)
			if err != nil {
				return err
			}
			dir := filepath.Dir(path)

			if stop {
				return debugcli.StopBackground(dir)
			}

			// 首次运行落一份骨架：配置页要有文件可编辑，界面首屏也会检查
			// 数据目录里有没有建出 config.json。
			if err := config.EnsureExists(path); err != nil {
				return err
			}

			// 默认后台常驻：把自身以 --foreground 拉起来后立即返回，父进程不装配服务
			// （装配了也用不上，还要跟着子进程一起退）。
			if !foreground {
				args := []string{"serve"}
				if listen != "" {
					args = append(args, "--listen", listen)
				}
				_, err := debugcli.StartBackground(path, debugcli.ServeSpec{
					Prefix:   "TT",
					StopHint: "tt serve --stop",
					APIBase:  debugcli.APIBaseUnderDebug,
					Args:     args,
				})
				return err
			}

			// 桌面壳：先看同目录有没有已经在跑的实例，有就**附着**，不启动第二个。
			//
			// 判据是既有的那一份（状态文件 + HTTP 探活，见 debug.RunningServe），
			// 与 --stop、tt update 用的是同一套 —— 不另造一个"桌面版专用"的探活。
			//
			// 这一步必须在 RunForeground 之前：它会因"服务已在运行"直接报错退出，
			// 而桌面壳要的是"接上去"。
			if readyJSON {
				if st := debugcli.RunningServe(dir); st != nil {
					return debugcli.WriteServeReady(os.Stdout, st, path, true)
				}
			}

			// 允许空环境：首次运行还没有任何环境，服务要能先起来，
			// 用户再到配置页里添加。CLI 命令那条路径仍然是"没配环境就报错"。
			cfg, err := debug.LoadConfigAllowEmpty(path)
			if err != nil {
				return err
			}

			// 启动清理缓存:只删超过 CacheMaxAge 的,不动手上的东西 ——
			// spill 里那些文件的用途正是"刚才那条查询的完整结果",启动即全清会把
			// 上一条命令刚告诉用户的东西删掉(见 internal/config/cache.go)。
			// 清缓存是维护动作不是启动前提,失败只报一行,不让服务起不来。
			if removed, freed, cerr := config.CleanCache(dir, config.CacheMaxAge); cerr != nil {
				fmt.Fprintf(os.Stderr, "缓存清理未完成: %v\n", cerr)
			} else if removed > 0 {
				fmt.Printf("已清理过期缓存 %d 个文件(释放 %.1f MB)\n",
					removed, float64(freed)/(1<<20))
			}
			cfg.DataDir = dir

			// 调试服务内嵌 SPA（统一前端构建产物）。统一服务只把 /debug/api/ 转给它、
			// 页面由 internal/web 提供；这份 FS 是为了 `tt debug serve` 单独跑时也能出页面。
			debugSrv := debug.NewServer(cfg, common.WebFrontend(), path)

			// 库目录**不在这里注入**：dictserver 每次访问时按配置算（顶层 dict.dir，
			// 空则 <数据目录>/dict），这样设置页改完立即生效，不会"改了还写老地方"。
			// 统一层的配置状态仍用「单文件」口径；设置页已改为按环境显示库目录。
			syncDefault := config.LocationsAt(path).SyncTarget("", "")
			dictSrv := dictserver.New(path)

			err = debugcli.RunForeground(path, debugcli.ServeSpec{
				Prefix:           "TT",
				StopHint:         "tt serve --stop",
				APIBase:          debugcli.APIBaseUnderDebug,
				ReadyJSON:        readyJSON,
				ExitOnStdinClose: exitOnStdinClose,
			}, func(ctx context.Context, onReady func(addr string)) error {
				// 停止路径有两条（Ctrl+C 与 /api/shutdown），两条都可能先到，
				// 所以统一收口到 sync.Once —— 否则会重复 close 同一个 channel 而 panic。
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				var once sync.Once
				stopWeb := func() { once.Do(cancel) }

				svc := web.New(web.Options{
					DebugFS:    common.WebFrontend(),
					ConfigPath: path,
					Version:    versionString(),
					// 裸版本给更新机制用（与展示串分开，见 web.Options.BareVersion）：
					// 干净件下就是构建注入的那个值；本地 go build 时为空 → 只允许查到“版本号不可用”。
					BareVersion: Version,
					Debug:       debugSrv.Handler(),
					Dict:        dictSrv.Handler(),
					// 配置写入后让持有内存态的调试服务重新加载，否则它内存里还是旧环境，
					// 表现是"改了没生效"。
					Reloaders: []web.ConfigReloader{debugSrv},
					Shutdown:  stopWeb,
					// 统一层配置状态仍按旧的单文件口径显示；设置页已改用 /api/dbsync 的库目录。
					SyncDefaultTarget: syncDefault,
				})

				addr := listen
				if addr == "" {
					addr = cfg.Listen
				}
				used, err := svc.ListenAndServe(addr)
				if err != nil {
					return err
				}
				// 状态文件里的地址要给客户端点，所以通配监听(0.0.0.0/[::])换成本机地址。
				onReady(displayAddr(used))

				<-ctx.Done()
				return nil
			})

			// 探活与启动之间被另一条进程抢先：RunForeground 报的是"服务已在运行"。
			// 这时再探一次，探得到就按附着收场 —— 一个能用的服务不该被报成启动失败
			// （桌面壳看到非 0 退出会弹失败页）。
			if err != nil && readyJSON {
				if st := debugcli.RunningServe(dir); st != nil {
					fmt.Fprintf(os.Stderr, "%v（改为附着到已在运行的实例）\n", err)
					return debugcli.WriteServeReady(os.Stdout, st, path, true)
				}
			}
			return err
		},
	}
	cmd.Flags().StringVar(&listen, "listen", "", "监听地址（缺省取配置里的 listen）")
	cmd.Flags().BoolVar(&stop, "stop", false, "停止后台实例")
	cmd.Flags().BoolVar(&foreground, "foreground", false, "前台运行，日志直出终端（默认后台常驻）")
	cmd.Flags().BoolVar(&readyJSON, "ready-json", false,
		"就绪后向 stdout 打一行 JSON（桌面壳用；需与 --foreground 同用）")
	cmd.Flags().BoolVar(&exitOnStdinClose, "exit-on-stdin-close", false,
		"stdin 关闭即优雅退出（桌面壳收尾兜底；需与 --foreground 同用）")
	return cmd
}

// displayAddr 把 0.0.0.0/[::] 这类通配监听地址换成本机地址，方便直接点开。
func displayAddr(addr string) string {
	host, port, err := splitHostPort(addr)
	if err != nil {
		return addr
	}
	switch host {
	case "", "0.0.0.0", "::", "[::]":
		return "127.0.0.1:" + port
	}
	return addr
}

// splitHostPort 拆 host:port。用标准库会多一个 import 且对无括号 IPv6 更严格，
// 这里的输入只来自 net.Listen 的结果，够用。
func splitHostPort(addr string) (host, port string, err error) {
	i := strings.LastIndex(addr, ":")
	if i < 0 {
		return "", "", fmt.Errorf("地址缺少端口: %s", addr)
	}
	host = strings.Trim(addr[:i], "[]")
	return host, addr[i+1:], nil
}
