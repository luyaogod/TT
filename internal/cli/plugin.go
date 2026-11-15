package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/output"
	"tt/internal/plugin"
)

// newPluginCmd 是 `tt plugin` 管理组。
//
// 管理面只碰**元数据与程序载荷**：注册表、版本目录、授权、启停状态。
// 它不启动第三方代码（那是 tt serve 里 supervisor 的事），因此即便某个插件的
// 程序坏了，这个组仍然能把它们列出来、禁用、卸载。
func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "插件管理：安装 / 启停 / 授权 / 卸载 / 自检",
		Long: `插件管理：安装 / 启停 / 授权 / 卸载 / 自检。

插件包是一个 zip（或一个目录），根部要有 plugin.json。它最多贡献四类能力：
CLI 命名空间（tt <namespace> …）、常驻服务、前端页面、skills。
四类互相独立，可以只声明其中一类。

安装只解包与注册，**不执行**任何插件代码；启用（并给出授权）是另一步。
CLI 命名空间与 TT 内置命令/其它已启用插件冲突时安装即失败。

退出码：0 成功 / 2 包或参数不合法（含冲突）/ 3 没有这个插件 / 1 其它失败。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(
		newPluginListCmd(),
		newPluginInfoCmd(),
		newPluginInstallCmd(),
		newPluginEnableCmd(),
		newPluginDisableCmd(),
		newPluginUninstallCmd(),
		newPluginGrantCmd(),
		newPluginDoctorCmd(),
	)
	return cmd
}

// pluginManager 构造管理面。
//
// 保留命名空间从命令树现算（并排除插件命名空间自身），所以"装了 plugin 插件叫 plugin
// 的命名空间"这种冲突一定会在安装时被抓住。
//
// SkillsSync 接的是命令层的共享实现（skills_sync.go）：插件管理面自己不碰技能管线，
// 但每次启用/停用/升级/卸载之后都要重算一次"已分发技能"的期望态。
func pluginManager() (*plugin.Manager, error) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return nil, err
	}
	loc := config.LocationsAt(cfgPath)
	m := plugin.NewManager(plugin.NewStore(loc), reservedNamespaces())
	m.ConfigPath = cfgPath
	m.SkillsSync = syncPluginSkills
	return m, nil
}

// pluginNamespaceAnnotation 标在"由插件注册出来的"命令上，便于与内置命令区分。
const pluginNamespaceAnnotation = "tt.plugin.namespace"

// reservedNamespaces 内置命令 + 旧别名 + TT 保留名。
//
// 每次现算而不是维护一份手写清单：手写清单一定会与命令树漂移，而这条规则的作用
// 恰恰是"别撞上任何真实存在的命令"。
func reservedNamespaces() []string {
	out := map[string]bool{
		// TT 自己的管理组与 cobra 自动生成的命令。
		"plugin": true, "help": true, "completion": true,
		// 历史别名（Execute 里会改写 os.Args[1]，但对插件来说它们是"看起来像命令"的名字）。
		"tdebug": true, "tdev": true, "tdict": true,
	}
	for _, c := range rootCmd.Commands() {
		if c.Annotations[pluginNamespaceAnnotation] == "1" {
			continue // 插件注册出来的命名空间不占保留名
		}
		out[c.Name()] = true
		for _, a := range c.Aliases {
			out[a] = true
		}
	}
	names := make([]string, 0, len(out))
	for n := range out {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// pluginErr 把插件包的校验/冲突错误映射成带稳定码的错误。
//
// 分类只看错误的"种类"（是不是 ValidationError / ErrZipUnsafe），不看文案 ——
// 文案会改，退出码不该跟着变。
func pluginErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "没有插件"):
		return output.Errorf("PLUGIN_NOT_INSTALLED", 3, "%s", msg)
	case isValidationErr(err), isZipUnsafe(err):
		return output.Errorf("PLUGIN_INVALID", 2, "%s", msg)
	case strings.Contains(msg, "冲突"), strings.Contains(msg, "占用"):
		return output.Errorf("PLUGIN_CONFLICT", 2, "%s", msg)
	default:
		return err
	}
}

// isValidationErr 是否是清单/授权校验失败。
func isValidationErr(err error) bool {
	var ve *plugin.ValidationError
	return errors.As(err, &ve)
}

// isZipUnsafe 是否是包内容不安全。
func isZipUnsafe(err error) bool {
	return errors.Is(err, plugin.ErrZipUnsafe)
}

//---------------------------------------------------------------------------
// list
//---------------------------------------------------------------------------

func newPluginListCmd() *cobra.Command {
	var all, brief bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出已安装插件与启停状态",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			list, err := m.List()
			if err != nil {
				return pluginErr(err)
			}
			if !all {
				filtered := make([]plugin.Summary, 0, len(list))
				for _, s := range list {
					if s.Enabled {
						filtered = append(filtered, s)
					}
				}
				list = filtered
			}
			if common.IsJSON() {
				return common.PrintJSON(map[string]any{"ok": true, "count": len(list), "plugins": list})
			}
			if len(list) == 0 {
				fmt.Println("没有已安装的插件。装一个：tt plugin install <插件包路径>")
				return nil
			}
			if brief {
				for _, s := range list {
					state := "off"
					if s.Enabled {
						state = "on"
					}
					fmt.Printf("%s\t%s\n", s.ID, state)
				}
				return nil
			}
			fmt.Printf("插件 %d 个\n", len(list))
			for _, s := range list {
				state := "未启用"
				if s.Enabled {
					state = "启用中"
				}
				fmt.Printf("  %-32s %-6s %-10s %s\n", s.ID, shortVersion(s.Version), state, contributionLine(s))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "连未启用的插件一起列出")
	cmd.Flags().BoolVar(&brief, "brief", false, "只列 id 与启停状态（制表符分隔）")
	return cmd
}

// shortVersion 版本展示（空版本给占位，避免打印成空白列）。
func shortVersion(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

// contributionLine 一句话描述插件的贡献面。
func contributionLine(s plugin.Summary) string {
	var parts []string
	if s.Namespace != "" {
		parts = append(parts, "cli:"+s.Namespace)
	}
	if s.HasService {
		parts = append(parts, "service")
	}
	if s.Route != "" {
		parts = append(parts, "ui:"+s.Route)
	}
	if len(s.SkillDirs) > 0 {
		parts = append(parts, fmt.Sprintf("skills×%d", len(s.SkillDirs)))
	}
	if len(parts) == 0 {
		return "（无贡献）"
	}
	return strings.Join(parts, " ")
}

//---------------------------------------------------------------------------
// info
//---------------------------------------------------------------------------

func newPluginInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <id>",
		Short: "看一个插件的清单、版本、授权与落点",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			info, err := m.Info(args[0])
			if err != nil {
				return pluginErr(err)
			}
			if common.IsJSON() {
				return common.PrintJSON(map[string]any{"ok": true, "plugin": info})
			}
			fmt.Printf("%s  %s\n", info.ID, info.Name)
			state := "未启用"
			if info.Enabled {
				state = "启用中"
			}
			fmt.Printf("  版本     %s（已装 %d 个版本）\n", shortVersion(info.Version), len(info.Versions))
			fmt.Printf("  状态     %s\n", state)
			fmt.Printf("  贡献     %s\n", contributionLine(info.Summary))
			if len(info.Permissions) > 0 {
				fmt.Printf("  声明权限 %s\n", strings.Join(info.Permissions, "、"))
			}
			if len(info.Capabilities) > 0 || len(info.Envs) > 0 {
				fmt.Printf("  已授权   能力 %s；环境 %s\n",
					orNone(strings.Join(info.Capabilities, "、")), orNone(strings.Join(info.Envs, "、")))
			}
			fmt.Printf("  程序目录 %s\n", info.Paths.VersionDir)
			fmt.Printf("  数据目录 %s\n", info.Paths.DataDir)
			if info.LastError != "" {
				fmt.Printf("  最近错误 %s\n", info.LastError)
			}
			if info.ManifestErr != "" {
				fmt.Printf("  清单读取 %s\n", info.ManifestErr)
			}
			return nil
		},
	}
}

// orNone 空串给一个明确的占位。
func orNone(s string) string {
	if s == "" {
		return "（无）"
	}
	return s
}

//---------------------------------------------------------------------------
// install
//---------------------------------------------------------------------------

func newPluginInstallCmd() *cobra.Command {
	var replace bool
	cmd := &cobra.Command{
		Use:   "install <目录|zip>",
		Short: "安装插件包（只解包注册，不启用、不执行插件代码）",
		Long: `安装插件包（只解包注册，不启用、不执行插件代码）。

包内根部要有 plugin.json。来源可以是本地目录（开发态）或本地 zip 包。
安装完成后用 tt plugin enable <id> --allow … --env … 启用并授权。

同 ID 同版本已存在时默认拒绝；确要重装加 --replace。
插件正在启用中时不允许装新版本（先 disable）：热升级在后续阶段提供。`,
		Example: `  tt plugin install ./myplugin.zip
  tt plugin install ./plugin-src --replace
  tt plugin enable com.example.myplugin --allow profiles.list,db.connect --env prod`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			rec, err := m.Install(plugin.InstallOptions{Source: args[0], Replace: replace})
			if err != nil {
				return pluginErr(err)
			}
			if common.IsJSON() {
				return common.PrintJSON(map[string]any{
					"ok": true, "id": rec.ID, "version": rec.ActiveVersion,
					"enabled": rec.Enabled, "contributions": rec.Contributions,
					"versionDir": m.Store.VersionDir(rec.ID, rec.ActiveVersion),
				})
			}
			fmt.Printf("已安装 %s %s（未启用）\n", rec.ID, rec.ActiveVersion)
			fmt.Printf("  程序目录 %s\n", m.Store.VersionDir(rec.ID, rec.ActiveVersion))
			fmt.Printf("  下一步   tt plugin enable %s --allow <能力> [--env <环境>]\n", rec.ID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&replace, "replace", false, "同 ID 同版本已存在时重装")
	return cmd
}

//---------------------------------------------------------------------------
// enable / disable
//---------------------------------------------------------------------------

func newPluginEnableCmd() *cobra.Command {
	var allow, envs []string
	cmd := &cobra.Command{
		Use:   "enable <id>",
		Short: "启用插件并授权（能力 + 环境）",
		Long: `启用插件并授权（能力 + 环境）。

能力（--allow，逗号或重复给）：profiles.list / db.connect / ssh.connect。
环境（--env，逗号或重复给）：只对这些环境开放连接；环境必须已在配置里存在。

清单里声明的权限只是"诉求"，不授予任何东西；这里给的才是授权。
声明了连接能力却不给 --env 时插件拿不到任何环境 —— 这是有意的默认：
环境授权是逐环境的决定，不该因为"装了个插件"就自动放开。`,
		Example: `  tt plugin enable com.example.myplugin --allow profiles.list
  tt plugin enable com.example.myplugin --allow db.connect,ssh.connect --env prod --env test`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			grants, err := parseGrants(allow, envs)
			if err != nil {
				return err
			}
			rec, err := m.Enable(args[0], grants)
			if err != nil {
				return pluginErr(err)
			}
			return reportPluginState(rec, m, "已启用")
		},
	}
	cmd.Flags().StringSliceVar(&allow, "allow", nil, "授予的能力，逗号分隔或重复给（profiles.list/db.connect/ssh.connect）")
	cmd.Flags().StringSliceVar(&envs, "env", nil, "授权的环境名，逗号分隔或重复给")
	return cmd
}

func newPluginGrantCmd() *cobra.Command {
	var allow, envs []string
	cmd := &cobra.Command{
		Use:   "grant <id>",
		Short: "改授权（不改变启用状态）",
		Long: `改授权（不改变启用状态）。

--allow / --env 是**整体替换**，不是追加：传什么就是什么。
要撤销某项能力，把它从 --allow 里去掉（或传空的 --allow ""）。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			grants, err := parseGrants(allow, envs)
			if err != nil {
				return err
			}
			rec, err := m.Grant(args[0], grants)
			if err != nil {
				return pluginErr(err)
			}
			return reportPluginState(rec, m, "授权已更新")
		},
	}
	cmd.Flags().StringSliceVar(&allow, "allow", nil, "授予的能力（整体替换）")
	cmd.Flags().StringSliceVar(&envs, "env", nil, "授权的环境名（整体替换）")
	return cmd
}

// parseGrants 解析 --allow/--env（逗号分隔 + 重复给都能用）。
func parseGrants(allow, envs []string) (plugin.Grants, error) {
	split := func(in []string) []string {
		var out []string
		for _, item := range in {
			for _, part := range strings.Split(item, ",") {
				if p := strings.TrimSpace(part); p != "" {
					out = append(out, p)
				}
			}
		}
		return out
	}
	g := plugin.Grants{Capabilities: split(allow), Envs: split(envs)}
	// 能力名在这里先挡一道：拼错的能力名不该被"授权成功"糊过去。
	for _, c := range g.Capabilities {
		if !validPluginCapability(c) {
			return plugin.Grants{}, output.Errorf("USAGE", 2,
				"未知能力 %q（可用: profiles.list / db.connect / ssh.connect）", c)
		}
	}
	return g, nil
}

// validPluginCapability 能力名白名单（与 Host API 一致）。
func validPluginCapability(name string) bool {
	switch name {
	case "profiles.list", "db.connect", "ssh.connect":
		return true
	}
	return false
}

// reportPluginState 统一的"状态已变更"输出。
func reportPluginState(rec *plugin.Record, m *plugin.Manager, done string) error {
	if rec == nil {
		return nil
	}
	if common.IsJSON() {
		return common.PrintJSON(map[string]any{
			"ok": true, "id": rec.ID, "enabled": rec.Enabled,
			"capabilities": rec.Grants.Capabilities, "envs": rec.Grants.Envs,
		})
	}
	fmt.Printf("%s %s\n", done, rec.ID)
	if len(rec.Grants.Capabilities) > 0 {
		fmt.Printf("  能力 %s\n", strings.Join(rec.Grants.Capabilities, "、"))
	}
	if len(rec.Grants.Envs) > 0 {
		fmt.Printf("  环境 %s\n", strings.Join(rec.Grants.Envs, "、"))
	}
	return nil
}

func newPluginDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <id>",
		Short: "停用插件（保留程序、授权与数据）",
		Long: `停用插件（保留程序、授权与数据）。

停用会停止它的常驻服务、撤销实例令牌、关闭它建立的连接，并把 CLI/页面/skills
从生效面移走。程序目录、授权与持久数据都保留，重新 enable 即恢复。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			rec, err := m.Disable(args[0])
			if err != nil {
				return pluginErr(err)
			}
			return reportPluginState(rec, m, "已停用")
		},
	}
}

//---------------------------------------------------------------------------
// uninstall
//---------------------------------------------------------------------------

func newPluginUninstallCmd() *cobra.Command {
	var purgeData, yes bool
	cmd := &cobra.Command{
		Use:   "uninstall <id>",
		Short: "卸载插件（默认保留持久数据）",
		Long: `卸载插件（默认保留持久数据）。

卸载先要求插件已停用，然后删除程序版本目录与注册项。**默认保留**插件持久数据
（<数据目录>/plugins/data/<id>）—— 数据是用户的，不是程序的一部分。
确要一起删，显式加 --purge-data --yes。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			if purgeData && !yes {
				return output.Errorf("USAGE", 2, "--purge-data 会删除插件持久数据，必须再加 --yes")
			}
			if err := m.Uninstall(args[0], plugin.UninstallOptions{PurgeData: purgeData}); err != nil {
				return pluginErr(err)
			}
			if common.IsJSON() {
				return common.PrintJSON(map[string]any{"ok": true, "id": args[0], "purgedData": purgeData})
			}
			fmt.Printf("已卸载 %s", args[0])
			if purgeData {
				fmt.Print("（含持久数据）")
			}
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().BoolVar(&purgeData, "purge-data", false, "连插件持久数据一起删除（需配合 --yes）")
	cmd.Flags().BoolVar(&yes, "yes", false, "确认危险动作")
	return cmd
}

//---------------------------------------------------------------------------
// doctor
//---------------------------------------------------------------------------

func newPluginDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor [id]",
		Short: "自检插件（清单、版本目录、授权、启用状态）",
		Long: `自检插件（清单、版本目录、授权、启用状态）。

只看本地事实：不启动插件、不连网。给定 id 只查一个，不给查全部。
发现问题时退出码为 1（问题清单照样输出，便于脚本收集）。`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			id := ""
			if len(args) == 1 {
				id = args[0]
			}
			issues, err := m.Doctor(id)
			if err != nil {
				return pluginErr(err)
			}
			if common.IsJSON() {
				if err := common.PrintJSON(map[string]any{
					"ok": len(issues) == 0, "count": len(issues), "issues": issues,
				}); err != nil {
					return err
				}
			} else {
				if len(issues) == 0 {
					fmt.Println("没有发现问题。")
				}
				for _, is := range issues {
					fmt.Printf("[%s] %s: %s\n", is.Severity, is.Code, is.Message)
					if is.Hint != "" {
						fmt.Printf("  → %s\n", is.Hint)
					}
				}
			}
			if len(issues) > 0 {
				return output.Errorf("PLUGIN_UNHEALTHY", 1, "发现 %d 个插件问题（见上方清单）", len(issues))
			}
			return nil
		},
	}
}

// pluginLogDir 插件日志目录（logs 命令与排障共用）。
func pluginLogDir(id string) string {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return ""
	}
	return config.LocationsAt(cfgPath).PluginLogsDir(id)
}

// ensureLogDir 建好插件日志目录（plugin 日志写入方用；失败返回空串）。
func ensureLogDir(id string) string {
	dir := pluginLogDir(id)
	if dir == "" {
		return ""
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}
	return filepath.Clean(dir)
}
