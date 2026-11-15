package cli

// tt plugin dev：不安装、直接在前台跑一个插件目录，并把临时 Host API 注入给它。
//
// 用途是开发循环：改代码 → 构建 → `tt plugin dev . -- hello`，日志直出终端；
// 服务模式用 `--service`（插件自己打印 ready 行）。

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/eventbus"
	"github.com/luyaogod/TT/internal/hostrpc"
	"github.com/luyaogod/TT/internal/plugin"
)

// devAuthorizer 开发态放行一切（插件没安装，注册表里没有它的记录）。
type devAuthorizer struct{}

func (devAuthorizer) Permitted(connsvc.Caller, connsvc.Capability) bool { return true }
func (devAuthorizer) AllowedEnvs(connsvc.Caller, connsvc.Capability) ([]string, bool) {
	return nil, true
}

func newPluginDevCmd() *cobra.Command {
	var service bool
	cmd := &cobra.Command{
		Use:   "dev <目录> [-- 参数…]",
		Short: "前台运行插件目录（注入临时 Host API，不安装）",
		Long: `不安装、不注册，直接在前台运行插件目录里的程序：

  tt plugin dev ./myplugin -- hello        # 跑 CLI 入口，参数是 hello
  tt plugin dev ./myplugin --service       # 跑常驻服务入口

它会启动一个临时的 Host API（只监听回环）并把 TT_PLUGIN_* 环境注入子进程。
Ctrl+C 与子进程同退出。`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			rest := args[1:]
			// 允许 `-- args` 形式：cobra 已把 -- 后的内容放进 args。
			if len(rest) > 0 && rest[0] == "--" {
				rest = rest[1:]
			}
			return runPluginDev(dir, rest, service)
		},
	}
	cmd.Flags().BoolVar(&service, "service", false, "运行常驻服务入口（清单里的 service.args）")
	return cmd
}

func runPluginDev(dir string, args []string, service bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	man, err := plugin.LoadManifest(abs)
	if err != nil {
		return pluginErr(err)
	}
	var entry string
	if service {
		if man.Contributions.Service == nil {
			return fmt.Errorf("清单没有 service 贡献")
		}
		entry, err = man.ResolveServiceEntry()
		if err == nil && len(args) == 0 {
			args = man.Contributions.Service.Args
		}
	} else {
		if man.Contributions.CLI == nil {
			return fmt.Errorf("清单没有 cli 贡献（用 --service 跑服务）")
		}
		entry, err = man.ResolveCLIEntry()
	}
	if err != nil {
		return pluginErr(err)
	}
	entryPath := filepath.Join(abs, filepath.FromSlash(entry))
	if st, serr := os.Stat(entryPath); serr != nil || st.IsDir() {
		return fmt.Errorf("入口不可用: %s", entryPath)
	}

	// 临时 Host API。
	tokens := hostrpc.NewTokenStore()
	events := eventbus.New(100)
	svc := connsvc.New(connsvc.Options{ConfigPath: common.ConfigPath, TTVersion: Version, Authorizer: devAuthorizer{}})
	handler := hostrpc.NewHandler(hostrpc.Deps{Service: svc, Auth: hostrpc.NewTokenAuthenticator(tokens), Events: events})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: handler}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()
	apiURL := "http://" + ln.Addr().String() + hostrpc.Prefix

	caller := connsvc.Caller{Kind: connsvc.CallerCLIPlugin, PluginID: man.ID, InstanceID: "dev"}
	token := tokens.Issue(caller)
	defer tokens.Revoke(token)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	child := exec.CommandContext(ctx, entryPath, args...)
	child.Dir = abs
	child.Stdin, child.Stdout, child.Stderr = os.Stdin, os.Stdout, os.Stderr
	child.Env = append(os.Environ(),
		"TT_HOST_API_URL="+apiURL,
		"TT_PLUGIN_TOKEN="+token,
		"TT_PLUGIN_ID="+man.ID,
		"TT_PLUGIN_VERSION="+man.Version,
		"TT_PLUGIN_DATA_DIR="+filepath.Join(abs, ".dev-data"),
		"TT_PLUGIN_LOGS_DIR="+filepath.Join(abs, ".dev-logs"),
	)
	if service {
		child.Env = append(child.Env, "TT_PLUGIN_LISTEN=127.0.0.1:0")
	}
	events.Publish("plugin.reload", map[string]any{"reason": "dev"})

	fmt.Fprintf(os.Stderr, "[dev] %s v%s → %s（Host API %s）\n", man.ID, man.Version, entry, apiURL)
	if err := child.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return pluginExitError{code: ee.ExitCode()}
		}
		return err
	}
	return nil
}
