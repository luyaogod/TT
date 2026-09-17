package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
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

// readyMark 是桌面模式（--desktop）在 stdout 上打印的就绪行前缀。
// Electron 外壳读它拿到真实地址（端口可能顺延），再交给 BrowserWindow 加载。
const readyMark = "TT_READY "

// readyInfo 是就绪行的载荷。
type readyInfo struct {
	URL    string `json:"url"`
	PID    int    `json:"pid"`
	Config string `json:"config"`
	Log    string `json:"log,omitempty"`
}

// newServeCmd 是统一的本地 Web 服务。
//
// 合并前每个工具各有一个 serve：tdebug 起调试界面、tdict 起配置页，各占一个端口、
// 各读自己那份配置。现在是一个进程、一个端口、一份配置、一套页面：
//
//	/debug/  调试工作台（含统一设置页：站点管理 / 数据字典 / DEBUG / 应用设置）
//	/api/*   统一接口 —— 环境与数据库配置、配置派生状态，以及字典类动作
//	         （源码镜像拉取、字典同步、BDL 文档）
func newServeCmd() *cobra.Command {
	var (
		listen  string
		stop    bool
		desktop bool
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "启动本地 Web 服务（调试工作台 + 统一设置页）",
		Long: `启动统一的本地 Web 服务。

一个进程提供工作台与统一设置页：

  /debug/          调试工作台：源码 / 断点 / 调用栈 / 变量 / 接口日志
  /debug/#settings 统一设置页：站点管理 / 数据字典 / DEBUG / 应用设置
                   —— 环境与数据库、查询数据源、源码镜像、字典同步、BDL 文档、
                      调试参数、明暗色，全在这里；三个工具共用同一份配置。

端口默认取配置里的 listen（缺省 127.0.0.1:28670），被占用时自动向后顺延。

前台运行，Ctrl+C 停止；` + "`--stop`" + ` 用于停止后台调试服务（tt debug serve 起的那个）。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := common.ResolveConfig(true)
			if err != nil {
				return err
			}
			dir := filepath.Dir(path)

			if stop {
				return debugcli.StopBackground(dir)
			}

			// 首次运行落一份骨架：配置页要有文件可编辑，桌面外壳也会检查
			// 数据目录里有没有建出 config.json（原 tdebug desktop 就做这件事）。
			if err := config.EnsureExists(path); err != nil {
				return err
			}

			// 允许空环境：首次运行/桌面版还没有任何环境，服务要能先起来，
			// 用户再到配置页里添加。CLI 命令那条路径仍然是"没配环境就报错"。
			cfg, err := debug.LoadConfigAllowEmpty(path)
			if err != nil {
				return err
			}
			cfg.DataDir = dir

			// 调试服务内嵌 SPA（统一前端构建产物）。统一服务只把 /debug/api/ 转给它、
			// 页面由 internal/web 提供；这份 FS 是为了 `tt debug serve` 单独跑时也能出页面。
			debugSrv := debug.NewServer(cfg, common.WebFrontend(), path)
			// 字典同步的目标缺省取**当前目录**的 erp_data.db，与 `tt dict -d` 的解析顺序
			// （cwd → exe 目录）第一条候选一致，所以"服务在这里同步、命令行在这里读"天然对得上。
			// 桌面外壳启动时 cwd 就是数据目录（见 desktop/main.js 的 spawn cwd）。
			// 要换地方用设置页的 sync.target，或命令行的 -d。
			//
			// 把这个值**同时**交给统一层与字典子系统：设置页会显示「默认位置：X」，
			// 而同步动作真正往那儿写；两边各算一次的话，显示的那个路径可能不是实际写入的那个。
			syncDefault := config.DefaultSyncTarget()
			dictSrv := dictserver.New(path)
			dictSrv.SetDBTarget(syncDefault)

			// 停止路径有两条（Ctrl+C 与 /api/shutdown），两条都可能先到，
			// 所以统一收口到 sync.Once —— 否则会重复 close 同一个 channel 而 panic。
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()
			var once sync.Once
			stop := func() { once.Do(cancel) }

			svc := web.New(web.Options{
				DebugFS:    common.WebFrontend(),
				ConfigPath: path,
				Version:    versionString(),
				Debug:      debugSrv.Handler(),
				Dict:       dictSrv.Handler(),
				// 配置写入后让持有内存态的调试服务重新加载，否则它内存里还是旧环境，
				// 表现是"改了没生效"。
				Reloaders: []web.ConfigReloader{debugSrv},
				Shutdown:  stop,
				// 与字典子系统同一个值，见上面的 syncDefault
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
			url := "http://" + displayAddr(used)

			if desktop {
				// 桌面模式：只吐一行机器可读的就绪信息，人看的界面在 Electron 窗口里。
				out, _ := json.Marshal(readyInfo{
					URL: url, PID: os.Getpid(), Config: path,
					Log: os.Getenv("TT_SERVE_LOG"),
				})
				fmt.Println(readyMark + string(out))
			} else {
				fmt.Printf("TT 服务已启动\n")
				fmt.Printf("  工作台与设置  %s/debug/\n", url)
				fmt.Printf("  配置文件      %s\n", path)
				fmt.Println("  按 Ctrl+C 停止")
			}

			<-ctx.Done()
			if !desktop {
				fmt.Println("\n已停止")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&listen, "listen", "", "监听地址（缺省取配置里的 listen）")
	cmd.Flags().BoolVar(&stop, "stop", false, "停止后台调试服务（tt debug serve 起的那个）")
	cmd.Flags().BoolVar(&desktop, "desktop", false, "桌面模式：只输出一行 TT_READY {json} 就绪信息")
	_ = cmd.Flags().MarkHidden("desktop")
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
