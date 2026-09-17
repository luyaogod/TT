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
// 各读自己那份配置。现在是一个进程、一个端口、一份配置，两套页面用挂载前缀分开：
//
//	/debug/  调试工作台（原 TDebug 前端 + 其 REST/WS）
//	/dict/   字典/镜像/同步页（原 TDictCli 前端 + 其 API）
//	/api/*   两套页面共用的接口，其中 /api/hosts 是共享的环境管理端点
func newServeCmd() *cobra.Command {
	var (
		listen  string
		stop    bool
		desktop bool
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "启动本地 Web 服务（调试工作台 + 字典页）",
		Long: `启动统一的本地 Web 服务。

一个进程同时提供两套页面，导航栏里互相有跳转入口：

  /debug/  调试工作台：源码 / 断点 / 调用栈 / 变量 / 接口日志
  /dict/   数据字典：环境与数据库配置、字典查询、源码镜像、字典同步

两套页面共用同一份环境配置（config.json 的 hosts 节），在一处添加的 SSH 环境，
调试、字典查询、源码镜像立刻都能用。

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

			// 调试服务内嵌 SPA，用子 FS —— 前端构建在 web/dist/debug 下。
			// （统一服务自身只把 /debug/api/ 转给它，页面由 internal/web 提供；
			//   子 FS 是为了 `tt debug serve` 单独跑时也能出页面。）
			debugSrv := debug.NewServer(cfg, common.WebSub("debug"), path)
			// 这里不调 SetDBTarget：字典页的同步目标缺省取**当前目录**的 erp_data.db，
			// 与 `tt dict -d` 的解析顺序（cwd → exe 目录）第一条候选一致，所以
			// "服务在这里同步、命令行在这里读"天然对得上。
			// 桌面外壳启动时 cwd 就是数据目录（见 desktop/main.js 的 spawn cwd），
			// 于是桌面场景下它自然落在配置旁边。要换地方用 sync.target 或 -d。
			dictSrv := dictserver.New(path)

			// 停止路径有两条（Ctrl+C 与 /api/shutdown），两条都可能先到，
			// 所以统一收口到 sync.Once —— 否则会重复 close 同一个 channel 而 panic。
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
			defer cancel()
			var once sync.Once
			stop := func() { once.Do(cancel) }

			svc := web.New(web.Options{
				DebugFS:    common.WebSub("debug"),
				DictFS:     common.WebSub("dict"),
				ConfigPath: path,
				Version:    versionString(),
				Debug:      debugSrv.Handler(),
				Dict:       dictSrv.Handler(),
				// 两个页面写的是同一份 hosts；任一处保存后都让调试服务重新加载，
				// 否则它内存里还是旧环境，表现是"改了没生效"。
				Reloaders: []web.ConfigReloader{debugSrv},
				Shutdown:  stop,
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
				fmt.Printf("  调试工作台  %s/debug/\n", url)
				fmt.Printf("  数据字典    %s/dict/\n", url)
				fmt.Printf("  配置文件    %s\n", path)
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
