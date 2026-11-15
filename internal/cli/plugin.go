package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/output"
	"github.com/luyaogod/TT/internal/plugin"
)

// newPluginCmd 是 `tt plugin` 管理组。
//
// 管理面只碰**元数据与程序载荷**：注册表、版本目录、授权、启停状态。
// 它不启动第三方代码（那是 tt serve 里 supervisor 的事），因此即便某个插件的
// 程序坏了，这个组仍然能把它们列出来、禁用、卸载。
func newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "插件管理：安装 / 启停 / 授权 / 升级 / 回滚 / 日志 / 卸载 / 自检",
		Long: `插件管理：安装 / 启停 / 授权 / 升级 / 回滚 / 日志 / 卸载 / 自检。

插件包是一个 zip（或一个目录），根部要有 plugin.json。它最多贡献四类能力：
CLI 命名空间（tt <namespace> …）、常驻服务、前端页面、skills。
四类互相独立，可以只声明其中一类。

安装只解包与注册，**不执行**任何插件代码；启用（并给出授权）是另一步。
CLI 命名空间与 TT 内置命令/其它已启用插件冲突时安装即失败。

升级是**手动**的：tt plugin update 现读发布索引、下载、校验、装成未激活版本，
再停旧服务、原子切换、起新服务并健康检查；任何一步失败都退回原来的版本。
tt plugin rollback 切回上一个版本。TT 不做后台检查、不自动下载。

退出码：0 成功 / 2 包、参数或发布索引不合法 / 3 没有这个插件 / 1 其它失败。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(
		newPluginListCmd(),
		newPluginInfoCmd(),
		newPluginInstallCmd(),
		newPluginPackCmd(),
		newPluginInspectCmd(),
		newPluginEnableCmd(),
		newPluginDisableCmd(),
		newPluginUninstallCmd(),
		newPluginGrantCmd(),
		newPluginUpdateCmd(),
		newPluginRollbackCmd(),
		newPluginLogsCmd(),
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
		Short: "启用插件（安装并启用即信任）",
		Long: `启用插件。v2 起**不再逐项授权**：启用后的插件可以使用 Host API 的全部能力
与配置里的全部环境。

--allow / --env 仍然接受（兼容旧脚本），但会被忽略；权限边界由操作系统与你对
插件的信任决定，而不是清单字段。`,
		Example: `  tt plugin enable com.example.myplugin`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			_ = allow
			_ = envs
			rec, err := m.Enable(args[0], plugin.Grants{})
			if err != nil {
				return pluginErr(err)
			}
			return reportPluginState(rec, m, "已启用")
		},
	}
	cmd.Flags().StringSliceVar(&allow, "allow", nil, "已废弃：v2 启用即信任，此参数被忽略")
	cmd.Flags().StringSliceVar(&envs, "env", nil, "已废弃：v2 启用即信任，此参数被忽略")
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

//---------------------------------------------------------------------------
// update / rollback（手动升级，见 internal/plugin/upgrade.go）
//---------------------------------------------------------------------------

func newPluginUpdateCmd() *cobra.Command {
	var indexURL, id, to string
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "手动升级插件（读发布索引，切版本并做健康检查；失败自动回滚）",
		Long: `手动升级插件：现读发布索引、下载校验、装成未激活版本，再停旧服务、原子切换、
起新服务并健康检查、重算技能；任何一步失败都退回原来的版本。

**不自动**：不做后台检查、不自动下载。什么时候升、升哪一版都由这一次命令决定。

--index 显式给发布索引地址（https；也可以是本机路径或 file://，用于离线/共享盘）。
不给时用该插件上次安装时记下的索引地址 —— TT 没有内置的插件市场地址。
--id 只升一个（默认：全部已装插件）；--to 钉一个版本。
--dry-run 只打印计划，一个字节都不写（不下载、不切版本）。

升级成功后只保留"当前 + 上一个"两个程序版本，持久数据永远保留。
tt serve 正在跑某个插件的常驻服务时，命令行升级会被拒（先 tt serve --stop），
因为那时切换版本会留下「页面新、服务旧」的混用状态。`,
		Example: `  tt plugin update --index https://plugins.example.com/index.json --dry-run
  tt plugin update --id com.example.myplugin
  tt plugin update --id com.example.myplugin --to 0.2.0`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			results, err := m.Update(cmd.Context(), plugin.UpdateOptions{
				IndexURL: indexURL, ID: id, To: to, DryRun: dryRun,
			})
			if err != nil {
				if plugin.IsIndexError(err) {
					return output.Errorf("PLUGIN_INDEX_INVALID", 2, "%v", err)
				}
				return output.Errorf("PLUGIN_UPDATE_FAILED", 1, "%v", err)
			}
			if common.IsJSON() {
				if perr := common.PrintJSON(map[string]any{
					"ok": !anyUpdateFailed(results), "dryRun": dryRun,
					"count": len(results), "results": results,
				}); perr != nil {
					return perr
				}
			} else {
				for _, res := range results {
					printUpdateResult(res, "已升级")
				}
			}
			if anyUpdateFailed(results) {
				return output.Errorf("PLUGIN_UPDATE_FAILED", 1, "有插件没有升级成功（见上方每一条）")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&indexURL, "index", "", "发布索引地址（https 或本机路径；默认用该插件上次的索引）")
	cmd.Flags().StringVar(&id, "id", "", "只升这一个插件（默认全部已装插件）")
	cmd.Flags().StringVar(&to, "to", "", "钉一个版本（必须出现在索引里且与当前 Host API 兼容）")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只打印计划，不下载、不切版本")
	return cmd
}

// anyUpdateFailed 这一批里有没有失败项。
func anyUpdateFailed(results []plugin.UpdateResult) bool {
	for _, r := range results {
		if r.Error != "" {
			return true
		}
	}
	return false
}

// printUpdateResult 一条升级/回滚结果的人读形态。done 是"已升级"/"已回滚"这类动词。
func printUpdateResult(r plugin.UpdateResult, done string) {
	switch {
	case r.Error != "" && r.RolledBack:
		fmt.Printf("%s  %s → %s  已回滚\n  %s\n", r.ID, shortVersion(r.From), shortVersion(r.From), r.Error)
	case r.Error != "":
		fmt.Printf("%s  未升级（仍是 %s）\n  %s\n", r.ID, shortVersion(r.From), r.Error)
	case r.UpToDate:
		fmt.Printf("%s  %s  已是最新（%s）\n", r.ID, shortVersion(r.From), r.Reason)
	case r.DryRun:
		fmt.Printf("%s  %s → %s  （--dry-run，什么都没写）\n", r.ID, shortVersion(r.From), shortVersion(r.To))
	case r.Updated:
		fmt.Printf("%s  %s → %s  %s\n", r.ID, shortVersion(r.From), shortVersion(r.To), done)
	default:
		fmt.Printf("%s  %s\n", r.ID, shortVersion(r.From))
	}
	for _, n := range r.Notes {
		fmt.Printf("  · %s\n", n)
	}
}

func newPluginRollbackCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "rollback <id>",
		Short: "回滚到上一个版本（切回并重新起服务、重算技能）",
		Long: `回滚到上一个版本。

"上一个版本"是注册表记着的上次切换前那个版本（升级成功时写入）。回滚与升级走同一套
安全动作：校验目标版本、停当前服务、撤实例凭证、原子切换、起服务并健康检查、重算技能。
回滚本身失败时会把活跃版本切回原处并报清原因。

只保留当前与上一个版本是策略，所以更早的版本已经不在了；回滚不会删掉你的持久数据。`,
		Example: `  tt plugin rollback com.example.myplugin
  tt plugin rollback com.example.myplugin --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			res, err := m.Rollback(args[0], plugin.RollbackOptions{DryRun: dryRun})
			if err != nil {
				return pluginErr(err)
			}
			if common.IsJSON() {
				return common.PrintJSON(map[string]any{"ok": true, "result": res})
			}
			printUpdateResult(*res, "已回滚")
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只说明会切到哪个版本，不改任何东西")
	return cmd
}

//---------------------------------------------------------------------------
// logs
//---------------------------------------------------------------------------

func newPluginLogsCmd() *cobra.Command {
	var tail int
	cmd := &cobra.Command{
		Use:   "logs <id>",
		Short: "看插件最新一次运行的日志（默认最后 50 行）",
		Long: `看插件最新一次运行的日志。

只读 logs/<id>/ 里**最新**的那个文件：一个插件长期运行会攒很多个日志文件，把整个目录
拼起来会把十分钟前的崩溃混进刚才那行错误里。
插件没起来过就没有日志 —— 那不是错误，输出会明确说"还没有日志"。`,
		Example: `  tt plugin logs com.example.myplugin
  tt plugin logs com.example.myplugin --tail 200 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := pluginManager()
			if err != nil {
				return err
			}
			if _, ierr := m.Info(args[0]); ierr != nil {
				return pluginErr(ierr)
			}
			file, lines, lerr := m.Store.TailLog(args[0], tail)
			if lerr != nil {
				return output.Errorf("PLUGIN_LOGS_FAILED", 1, "%v", lerr)
			}
			if common.IsJSON() {
				return common.PrintJSON(map[string]any{
					"ok": true, "id": args[0], "file": file, "count": len(lines), "lines": lines,
				})
			}
			if file == "" {
				fmt.Printf("%s 还没有日志（服务没启动过，或日志目录被清掉了）\n", args[0])
				return nil
			}
			fmt.Printf("# %s\n", file)
			for _, ln := range lines {
				fmt.Println(ln)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&tail, "tail", plugin.DefaultLogTail, "只显示最后 N 行")
	return cmd
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
