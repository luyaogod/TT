package cli

// 第三方 CLI 的接入层：把已启用插件的命名空间挂到命令树上，并把一次调用转发给插件进程。
//
// 两条硬约束（见 .agents/docs/plugin-system/05-cli-and-errors.md）：
//  1. 根帮助只读注册表里的静态元数据，**不启动**任何插件；
//  2. 插件进程的退出码原样返回，TT 不为它打印自己的错误信封。

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/output"
	"github.com/luyaogod/TT/internal/plugin"
)

// presetConfigFromArgs 在 cobra 解析之前把 `--config` 落到 TT_CONFIG。
//
// 为什么必须提前：插件命名空间要在解析**之前**挂到命令树上（这样 cobra 才能找到
// `tt <namespace>`），而"读哪份注册表"取决于配置位置。cobra 还没解析，所以只能自己看一眼
// os.Args。与 dev 的 takeRootFlags 同一个手法、同一个理由（解析只经 TT_CONFIG）。
//
// 只认 --config：其它全局开关不影响"命令树长什么样"。
func presetConfigFromArgs() {
	for i := 1; i < len(os.Args); i++ {
		a := os.Args[i]
		switch {
		case a == "--config":
			if i+1 < len(os.Args) {
				_ = os.Setenv("TT_CONFIG", os.Args[i+1])
			}
			return
		case strings.HasPrefix(a, "--config="):
			_ = os.Setenv("TT_CONFIG", strings.TrimPrefix(a, "--config="))
			return
		}
	}
}

// attachPluginNamespaces 把插件 CLI 命名空间挂到 rootCmd 上。
//
// 挂在 `Execute` 里、cobra 解析之前：这时 rootCmd 已由各包的 init 装配完毕，
// 而 `--at-list` 的预扫描还没跑。任何读取失败都**静默跳过** ——
// 插件基础设施不能成为 TT 启动的前置条件（注册表坏了也得能 tt --help、tt plugin doctor）。
//
// 未启用的插件也挂、但标记 Hidden：根帮助里看不到它们（契约要求），
// 而用户直接敲 `tt <那个命名空间>` 时会得到一句明确的"未启用"，而不是
// cobra 的 "unknown command"（那会让人以为装都没装上）。
func attachPluginNamespaces() {
	reg, err := pluginRegistry()
	if err != nil {
		return
	}
	for _, rec := range reg.Plugins {
		c := rec.Contributions.CLI
		if c == nil || c.Namespace == "" {
			continue
		}
		if findTopLevel(rootCmd, c.Namespace) != nil {
			// 与内置命令/别的插件同名：安装时已拦过，这里是第二道门（注册表可能被手改过）。
			fmt.Fprintf(os.Stderr, "跳过插件 %s 的命名空间 %q：与已有命令同名\n", rec.ID, c.Namespace)
			continue
		}
		cmd := newPluginNamespaceCmd(rec.ID, c)
		cmd.Hidden = !rec.Enabled
		rootCmd.AddCommand(cmd)
	}
}

// findTopLevel 在根部找同名命令（含别名）。
func findTopLevel(root *cobra.Command, name string) *cobra.Command {
	for _, c := range root.Commands() {
		if c.Name() == name {
			return c
		}
		for _, a := range c.Aliases {
			if a == name {
				return c
			}
		}
	}
	return nil
}

// pluginRegistry 读注册表（不建管理面，避免为一次帮助引入全套依赖）。
func pluginRegistry() (*plugin.Registry, error) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return nil, err
	}
	return plugin.NewStore(config.LocationsAt(cfgPath)).Load()
}

// newPluginNamespaceCmd 一个插件命名空间的 cobra 代理节点。
//
// DisableFlagParsing：业务参数与 flag 一律不预解析 —— TT 不解释插件的参数，
// 只把自己那几个全局开关摘出来（见 takePluginGlobalFlags）。
func newPluginNamespaceCmd(id string, c *plugin.CLIContribution) *cobra.Command {
	summary := strings.TrimSpace(c.Summary)
	if summary == "" {
		summary = "插件 " + id
	}
	cmd := &cobra.Command{
		Use:                c.Namespace + " [参数…]",
		Short:              summary,
		Long:               fmt.Sprintf("%s\n\n这是插件 %s 的命令命名空间：参数、标准流与退出码原样交给插件进程；\n具体用法请用 tt %s --help（由插件自己输出）。", summary, id, c.Namespace),
		DisableFlagParsing: true,
		Annotations:        map[string]string{pluginNamespaceAnnotation: "1"},
		RunE: func(cmd *cobra.Command, args []string) error {
			rest, envName, ent, err := takePluginGlobalFlags(args)
			if err != nil {
				return err
			}
			m, err := pluginManager()
			if err != nil {
				return err
			}
			rec, err := m.FindByNamespace(c.Namespace)
			if err != nil {
				return output.Errorf("PLUGIN_REGISTRY_BROKEN", 1, "读取插件注册表失败: %v", err)
			}
			if rec == nil {
				return output.Errorf("PLUGIN_DISABLED", 3,
					"插件 %s 未启用（tt plugin list --all 看状态；tt plugin enable %s 启用）", id, id)
			}
			d := &plugin.Dispatcher{
				Store:      m.Store,
				Authorizer: plugin.RegistryAuthorizer{Store: m.Store},
				TTVersion:  Version,
			}
			res, err := d.Run(cmd.Context(), rec, envName, ent, rest)
			if err != nil {
				return output.Errorf("PLUGIN_SPAWN_FAILED", 1, "%v", err)
			}
			if res.ExitCode != 0 {
				// 原样返回插件退出码，且**不**再打印信封（插件的输出已经是全部内容）。
				return pluginExitError{code: res.ExitCode}
			}
			return nil
		},
	}
	return cmd
}

// pluginExitError 把插件退出码原样送出去。
//
// AlreadyReported：插件自己已经把输出（含错误）打到 stdout/stderr 了，
// 根命令不能再补一个 {"ok":false,...} 信封 —— 那会让 stdout 上出现两个 JSON 对象。
type pluginExitError struct{ code int }

func (e pluginExitError) Error() string { return fmt.Sprintf("插件退出码 %d", e.code) }

// ExitCode 供 root 的 exitCodeOf 读取。
func (e pluginExitError) ExitCode() int { return e.code }

// AlreadyReported 表示"输出已经完整给出"。
func (e pluginExitError) AlreadyReported() bool { return true }

// takePluginGlobalFlags 从透传参数里摘出 TT 的全局开关，剩下的原样交给插件。
//
// 规则：**全局开关必须在业务参数之前**。从第一个非 TT 全局开关的参数起，后面全部
// 按顺序原样交给插件 —— 这样 `tt ns --env prod run --env x` 里只有第一个 --env 是 TT 的，
// 插件自己的同名 flag 不会被吃掉。无法区分"插件也恰好有 --env"这件事，所以用位置定界，
// 与用户敲 `tt ns <TT 开关> <插件参数…>` 的直觉一致。
//
// 摘出来的 --config 直接落到 TT_CONFIG（与 dev 的做法一致：解析只经 config.ResolvePath）。
func takePluginGlobalFlags(args []string) (rest []string, envName, ent string, err error) {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !isPluginGlobalFlag(a) {
			out = append(out, args[i:]...)
			return out, envName, ent, nil
		}
		switch {
		case a == "--config" || a == "--env" || a == "--at" || a == "--conn" || a == "--format":
			if i+1 >= len(args) {
				return nil, "", "", output.Errorf("USAGE", 2, "%s 需要参数", a)
			}
			val := args[i+1]
			i++
			if err := applyPluginGlobalFlag(a, val, &envName, &ent); err != nil {
				return nil, "", "", err
			}
		case strings.HasPrefix(a, "--config="), strings.HasPrefix(a, "--env="),
			strings.HasPrefix(a, "--at="), strings.HasPrefix(a, "--conn="), strings.HasPrefix(a, "--format="):
			kv := strings.SplitN(a, "=", 2)
			if err := applyPluginGlobalFlag(kv[0], kv[1], &envName, &ent); err != nil {
				return nil, "", "", err
			}
		default:
			// --json / --csv / --human / --verbose / -v：吃掉。
			// 插件输出不归 TT 重新排版，用户敲的全局开关也不该被插件当成业务参数。
		}
	}
	return out, envName, ent, nil
}

// isPluginGlobalFlag 判断是否 TT 的全局开关（用于"从第一个业务参数起停止摘取"）。
func isPluginGlobalFlag(a string) bool {
	switch a {
	case "--config", "--env", "--at", "--conn", "--format", "--json", "--csv", "--human", "--verbose", "-v":
		return true
	}
	return strings.HasPrefix(a, "--config=") || strings.HasPrefix(a, "--env=") ||
		strings.HasPrefix(a, "--at=") || strings.HasPrefix(a, "--conn=") || strings.HasPrefix(a, "--format=")
}

// applyPluginGlobalFlag 把一个全局开关的值落到对应的 TT 状态上。
func applyPluginGlobalFlag(name, val string, envName, ent *string) error {
	val = strings.TrimSpace(val)
	switch name {
	case "--config":
		_ = os.Setenv("TT_CONFIG", val)
	case "--env", "--conn":
		*envName = val
	case "--at":
		// <环境> 或 <环境>@<企业>（与 tt 的 --at 文法一致）。
		if i := strings.IndexByte(val, '@'); i >= 0 {
			*envName = strings.TrimSpace(val[:i])
			*ent = strings.TrimSpace(val[i+1:])
		} else {
			*envName = val
		}
	case "--format":
		// 只接受 TT 认识的取值；插件自己的输出排版与它无关。
		switch strings.ToLower(val) {
		case "json", "csv", "table", "text", "":
			common.Format = val
		default:
			return output.Errorf("USAGE", 2, "未知的输出形态 %q（可用: json / csv / table）", val)
		}
	}
	return nil
}
