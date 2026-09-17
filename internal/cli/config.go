package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
)

// newConfigCmd 是统一的配置管理命令组。
//
// 合并前两个工具各有自己的位置规则与迁移逻辑，注释里写着"与对方保持一致，
// 改动请两边同步"。现在两者共用 internal/config 的实现，这里只做一层 CLI 门面。
func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "配置管理：位置 / 查看 / 读写 / 迁移 / 校验",
		Long: `管理唯一的 config.json。

三个工具共用这一份配置：顶层的 hosts 节是共用环境清单，
debug / query / mirror / bdldoc / sync / tdev 是各工具自己的设置。

配置位置按以下顺序确定（第一个存在的胜出）：
  1. TT_CONFIG 环境变量（兼容旧名 TDEBUG_CONFIG / TDICT_CONFIG）
  2. --config <路径>
  3. <exe 目录>\.portable 存在 → 便携包，配置留在包内
  4. ` + config.DefaultConfigPathHint() + `
统一位置可用 T100_HOME 环境变量整体改写。`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "打印配置文件的解析结果",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if common.JSON {
				path, err := common.ResolveConfig(true)
				if err != nil {
					return err
				}
				return common.PrintJSON(map[string]any{
					"path":     path,
					"default":  config.DefaultConfigPath(),
					"portable": config.IsPortable(),
					"exists":   fileExists(path),
				})
			}
			path, err := common.ResolveConfig(false)
			if err != nil {
				// 找不到时也把缺省落点和尝试过的路径打出来 —— 这正是用户要的信息
				cmd.Println("配置文件尚未创建。")
				cmd.Printf("缺省落点：%s\n", config.DefaultConfigPath())
				return nil
			}
			cmd.Println(path)
			return nil
		},
	})

	cmd.AddCommand(newConfigShowCmd())

	cmd.AddCommand(&cobra.Command{
		Use:   "get <键路径>",
		Short: "按点分路径读一个值（如 hosts.activeEnv、debug.termWidth）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := common.ResolveConfig(false)
			if err != nil {
				return err
			}
			root, err := config.Open(path)
			if err != nil {
				return err
			}
			v, ok := lookupPath(root, args[0])
			if !ok {
				return fmt.Errorf("配置里没有 %q", args[0])
			}
			if common.JSON {
				return common.PrintJSON(v)
			}
			return printScalarOrJSON(cmd, v)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <键路径> <值>",
		Short: "按点分路径写一个值（值按 JSON 解析，解析不了当字符串）",
		Long: `按点分路径写一个值。

值优先按 JSON 解析，所以 200 是数字、true 是布尔、'["a","b"]' 是数组；
解析失败则当作纯字符串（"10.0.0.1" 这种不必加引号）。

只用点分路径读写已存在的路径；若要新增环境或改环境清单，用 tt env 或配置页 ——
那里有校验（环境名唯一、端口范围、账号非空等），直接 set 会绕过它们。

写操作走与配置页完全相同的原子写路径，失败不会留下半截文件。`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := common.ResolveConfig(true)
			if err != nil {
				return err
			}
			keyPath, raw := args[0], args[1]
			value := parseValue(raw)
			err = config.Edit(path, nil, func(root map[string]any) error {
				return setPath(root, keyPath, value)
			})
			if err != nil {
				return err
			}
			cmd.Printf("%s = %s\n", keyPath, compactJSON(value))
			return nil
		},
	})

	cmd.AddCommand(newConfigMigrateCmd())
	cmd.AddCommand(newConfigValidateCmd())
	return cmd
}

// newConfigMigrateCmd 预览/执行旧配置合并。
func newConfigMigrateCmd() *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "把合并前 tdebug / tdict 的配置合并到统一位置",
		Long: `把 TDebug 与 TDictCli 各自留下的配置合并成一份统一配置。

合并规则：
  - 环境清单取并集，同名环境以 TDictCli 那份为准（它是字典查询的现役配置）；
  - TDebug 的 debug 节被拆开：sshs 提升为 hosts.sshs，其余键留在 debug；
  - query / mirror / bdldoc / sync 原样带过来；
  - 原文件不删除，各留一份 .pre-merge.bak。

tt 首次运行会自动做这件事，一般不需要手动执行 —— 这个命令用于预览结果，
或在上次迁移中断后重跑。

--dry-run 只打印合并结果，不写任何文件。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun {
				return runConfigMigrateDryRun(cmd)
			}
			path := config.UserConfigPath()
			if path == "" {
				return fmt.Errorf("定位不到统一用户目录（T100_HOME / %%APPDATA%%）")
			}
			srcs := config.Migrate()
			if len(srcs) == 0 {
				cmd.Println("无需迁移：目标已是当前结构，或找不到可合并的旧配置。")
				return nil
			}
			cmd.Printf("已合并 %d 份旧配置到 %s\n", len(srcs), path)
			for _, s := range srcs {
				cmd.Printf("  ← %s\n", s)
			}
			cmd.Println("源文件未删除，各留了一份 .pre-merge.bak。")
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只打印合并结果，不写文件")
	return cmd
}

func runConfigMigrateDryRun(cmd *cobra.Command) error {
	path := config.UserConfigPath()
	if path == "" {
		return fmt.Errorf("定位不到统一用户目录（T100_HOME / %%APPDATA%%）")
	}
	plan, err := config.PlanMigration(path)
	if err != nil {
		return err
	}
	if plan == nil || len(plan.Sources) == 0 {
		cmd.Println("无需迁移：目标已是当前结构，或找不到可合并的旧配置。")
		return nil
	}
	cmd.Printf("将写入：%s\n", plan.Dst)
	cmd.Println("将合并：")
	for _, s := range plan.Sources {
		cmd.Printf("  ← %s\n", s.Path)
	}
	out, err := json.MarshalIndent(plan.Merged, "", "  ")
	if err != nil {
		return err
	}
	cmd.Println("\n合并结果：")
	cmd.Println(string(out))
	return nil
}

// newConfigValidateCmd 校验配置能否被正常解析，并报出可疑之处。
func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "校验配置：能否解析、环境是否完整、缺省值是否被补全",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := common.ResolveConfig(false)
			if err != nil {
				return err
			}
			problems := validateConfig(path)
			if common.JSON {
				return common.PrintJSON(map[string]any{
					"config":   path,
					"problems": problems,
					"ok":       len(problems) == 0,
				})
			}
			if len(problems) == 0 {
				cmd.Printf("%s：未发现问题。\n", path)
				return nil
			}
			cmd.Printf("%s：发现 %d 个问题\n", path, len(problems))
			for _, p := range problems {
				cmd.Printf("  - %s\n", p)
			}
			// 有问题不代表不能用，只提示；不返回错误以免脚本里误判为致命
			return nil
		},
	}
}

// validateConfig 收集配置的可疑之处。返回空切片表示没问题。
func validateConfig(path string) []string {
	var problems []string

	root, err := config.Open(path)
	if err != nil {
		return []string{err.Error()}
	}
	if v := intOrZero(root["schemaVersion"]); v < config.SchemaVersion {
		problems = append(problems, fmt.Sprintf(
			"schemaVersion = %d，当前结构为 %d；这是合并前的旧配置，运行 tt config migrate 升级", v, config.SchemaVersion))
	}

	r, err := config.Load(path)
	if err != nil {
		return append(problems, err.Error())
	}
	if len(r.Hosts.SSHs) == 0 {
		problems = append(problems, "hosts.sshs 为空：三个工具都需要至少一个 SSH 环境")
	}

	seen := map[string]int{}
	for i := range r.Hosts.SSHs {
		e := &r.Hosts.SSHs[i]
		where := fmt.Sprintf("hosts.sshs[%d]", i)
		if e.Name == "" {
			problems = append(problems, where+" 缺少 name")
			where = "hosts.sshs[无名]"
		} else {
			seen[e.Name]++
			if seen[e.Name] > 1 {
				problems = append(problems, fmt.Sprintf("环境名 %q 重复出现", e.Name))
			}
		}
		if e.Host == "" {
			problems = append(problems, where+" 缺少 host")
		}
		if e.User == "" {
			problems = append(problems, where+" 缺少 user")
		}
		if e.DB == nil {
			problems = append(problems, where+" 未挂数据库连接（字典查询与 db 类命令会不可用）")
			continue
		}
		switch e.DB.Type {
		case "oracle", "kingbase":
		case "":
			problems = append(problems, where+".db 缺少 type（应为 oracle 或 kingbase）")
		default:
			problems = append(problems, fmt.Sprintf("%s.db.type = %q，应为 oracle 或 kingbase", where, e.DB.Type))
		}
		if e.DB.Host == "" {
			problems = append(problems, where+".db 缺少 host")
		}
		if len(e.DB.Accounts) == 0 {
			problems = append(problems, where+".db.accounts 为空（客户端直连取首项，会是空账号）")
		}
	}

	if r.Hosts.ActiveEnv != "" && r.Hosts.ByName(r.Hosts.ActiveEnv) == nil {
		problems = append(problems, fmt.Sprintf(
			"hosts.activeEnv = %q 指向不存在的环境", r.Hosts.ActiveEnv))
	}
	if r.Debug.ActiveEnv != "" && r.Hosts.ByName(r.Debug.ActiveEnv) == nil {
		problems = append(problems, fmt.Sprintf(
			"debug.activeEnv = %q 指向不存在的环境", r.Debug.ActiveEnv))
	}
	return problems
}

func newConfigShowCmd() *cobra.Command {
	var raw bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "打印配置内容（口令默认打码，--raw 原样输出）",
		Long: `打印 config.json 的内容。

默认把口令字段打码 —— 终端输出常被贴进 issue 或聊天记录里。
需要看原样内容时加 --raw。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(cmd, raw)
		},
	}
	cmd.Flags().BoolVar(&raw, "raw", false, "原样输出（含明文口令）")
	return cmd
}

func runConfigShow(cmd *cobra.Command, raw bool) error {
	path, err := common.ResolveConfig(false)
	if err != nil {
		return err
	}
	root, err := config.Open(path)
	if err != nil {
		return err
	}
	// redactSecrets 递归处理整棵树，顶层仍是 map；断言一下让类型收敛回 map[string]any。
	shown := any(root)
	if !raw {
		shown = redactSecrets(root)
	}
	out, err := json.MarshalIndent(shown, "", "  ")
	if err != nil {
		return err
	}
	cmd.Println(string(out))
	if !raw {
		cmd.Println("\n（口令已打码；加 --raw 查看原样内容）")
	}
	return nil
}

// redactSecrets 深度复制一份配置，把口令字段替换成占位符。
//
// 识别到的键名：password / passwd / pwd，以及 db.accounts[].password。
// 用键名判断而不是靠结构体标签，是因为未知节的未知键也要一起打码。
func redactSecrets(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			if isSecretKey(k) {
				if s, ok := val.(string); ok && s != "" {
					out[k] = "***"
					continue
				}
			}
			out[k] = redactSecrets(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = redactSecrets(e)
		}
		return out
	default:
		return v
	}
}

func isSecretKey(k string) bool {
	switch strings.ToLower(k) {
	case "password", "passwd", "pwd":
		return true
	}
	return false
}

// ---------- 点分路径读写 ----------

// lookupPath 按点分路径取值。路径段为对象键或数组下标。
func lookupPath(root any, path string) (any, bool) {
	cur := root
	for _, seg := range strings.Split(path, ".") {
		switch t := cur.(type) {
		case map[string]any:
			v, ok := t[seg]
			if !ok {
				return nil, false
			}
			cur = v
		case []any:
			i, err := strconv.Atoi(seg)
			if err != nil || i < 0 || i >= len(t) {
				return nil, false
			}
			cur = t[i]
		default:
			return nil, false
		}
	}
	return cur, true
}

// setPath 按点分路径写值。中间层缺失时自动建对象，数组下标必须已存在。
func setPath(root map[string]any, path string, value any) error {
	segs := strings.Split(path, ".")
	if len(segs) == 0 || segs[0] == "" {
		return fmt.Errorf("键路径不能为空")
	}

	var cur any = root
	for i, seg := range segs {
		last := i == len(segs)-1
		switch t := cur.(type) {
		case map[string]any:
			if last {
				t[seg] = value
				return nil
			}
			next, ok := t[seg]
			if !ok {
				m := map[string]any{}
				t[seg] = m
				cur = m
				continue
			}
			cur = next
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(t) {
				return fmt.Errorf("数组下标 %q 越界或无对应元素", seg)
			}
			if last {
				t[idx] = value
				return nil
			}
			cur = t[idx]
		default:
			return fmt.Errorf("%q 的上一级不是对象或数组，无法继续", seg)
		}
	}
	return nil
}

// parseValue 把命令行里的值按 JSON 解析；解析不了就当纯字符串。
// 这样 `tt config set debug.termWidth 200` 写进去的是数字而不是字符串 "200"。
func parseValue(raw string) any {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err == nil {
		return v
	}
	return raw
}

func printScalarOrJSON(cmd *cobra.Command, v any) error {
	switch t := v.(type) {
	case string:
		cmd.Println(t)
		return nil
	case float64:
		cmd.Println(strconv.FormatFloat(t, 'f', -1, 64))
		return nil
	case bool:
		cmd.Println(strconv.FormatBool(t))
		return nil
	case nil:
		cmd.Println("null")
		return nil
	default:
		out, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		cmd.Println(string(out))
		return nil
	}
}

func compactJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func intOrZero(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		i, _ := strconv.Atoi(n)
		return i
	}
	return 0
}

// fileExists 报告路径是否存在一个普通文件。
func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}
