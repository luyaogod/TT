package cli

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
)

// newEnvCmd 是统一的环境管理命令组。
//
// 合并前 TDictCli 有 `tdict env [<环境名>]`、TDebug 有 `tdebug env` 与 `tdebug topent`，
// 三者操作的是同一份数据（一个环境清单），只是各自实现了一遍。合并后只有这一份。
func newEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "管理 SSH 环境（三个工具共用同一份环境清单）",
		Long: `列出/查看/切换 config.json 里 hosts.sshs 定义的 SSH 环境。

环境清单是三个工具共用的：调试、字典查询、源码镜像都从同一份 hosts 里取连接，
所以在这里加一台机器，tt debug 和 tt dict 都能立刻用。`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runEnvList(cmd)
			}
			return runEnvShow(cmd, args[0])
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "列出全部环境（同不带参数）",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return runEnvList(cmd) },
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show <环境名>",
		Short: "显示某个环境的连接详情（口令打码）",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return runEnvShow(cmd, args[0]) },
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "use <环境名>",
		Short: "把某环境设为默认（写 hosts.activeEnv）",
		Long: `把某环境设为默认环境。三个工具在没有显式指定环境时都用它。

原来的 tdict 用 hosts.activeEnv、tdebug 用 debug 节里的「当前环境」，
合并后统一走 hosts.activeEnv —— 在哪个工具里切换，另外两个也跟着切。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := common.ResolveConfig(true)
			if err != nil {
				return err
			}
			name := args[0]
			err = config.EditHosts(path, func(h *config.Hosts) error {
				if h.ByName(name) == nil {
					return fmt.Errorf("没有名为 %q 的环境（用 tt env list 看全部）", name)
				}
				h.ActiveEnv = name
				return nil
			})
			if err != nil {
				return err
			}
			cmd.Printf("默认环境已设为 %s\n", name)
			return nil
		},
	})

	cmd.AddCommand(newEnvSetTopentCmd())
	return cmd
}

// newEnvSetTopentCmd 设置某环境的默认企业编号（TOPENT）。
// 对应合并前 tdebug 的 `topent` 命令。
func newEnvSetTopentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "topent <环境名> <企业编号>",
		Short: "设置某环境的默认企业编号（TOPENT）",
		Long: `设置登录该环境时下发的默认企业编号。

企业编号可以是数字（如 10001）也可以是文本；留空字符串表示清除。`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := common.ResolveConfig(true)
			if err != nil {
				return err
			}
			name, topent := args[0], args[1]
			err = config.EditHosts(path, func(h *config.Hosts) error {
				e := h.ByName(name)
				if e == nil {
					return fmt.Errorf("没有名为 %q 的环境（用 tt env list 看全部）", name)
				}
				e.Topent = config.EntValue(topent)
				return nil
			})
			if err != nil {
				return err
			}
			if topent == "" {
				cmd.Printf("已清除 %s 的默认企业编号\n", name)
			} else {
				cmd.Printf("%s 的默认企业编号已设为 %s\n", name, topent)
			}
			return nil
		},
	}
}

// envView 是环境在 --json 输出里的形状。口令一律不出现在输出里。
type envView struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Zone      string `json:"zone,omitempty"`
	Topent    string `json:"topent,omitempty"`
	DBType    string `json:"dbType,omitempty"`
	DBAddress string `json:"dbAddress,omitempty"`
	Active    bool   `json:"active"`
}

func envViews(h *config.Hosts) []envView {
	out := make([]envView, 0, len(h.SSHs))
	for i := range h.SSHs {
		e := &h.SSHs[i]
		v := envView{
			Name:   e.Name,
			Host:   e.Host,
			Port:   e.Port,
			User:   e.User,
			Zone:   e.Zone,
			Topent: string(e.Topent),
			Active: e.Name == h.ActiveEnv,
		}
		if e.DB != nil {
			v.DBType = e.DB.Type
			v.DBAddress = e.DB.Address()
		}
		out = append(out, v)
	}
	return out
}

func runEnvList(cmd *cobra.Command) error {
	path, err := common.ResolveConfig(false)
	if err != nil {
		return err
	}
	root, err := config.Load(path)
	if err != nil {
		return err
	}
	h := root.Hosts

	if common.JSON {
		return common.PrintJSON(map[string]any{
			"activeEnv": h.ActiveEnv,
			"config":    path,
			"sshs":      envViews(&h),
		})
	}

	if len(h.SSHs) == 0 {
		cmd.Printf("尚未配置任何环境。\n配置文件：%s\n运行 tt serve 打开配置页添加，或编辑其中的 hosts.sshs。\n", path)
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, " \t名称\tSSH\t区域\t企业\t数据库")
	for _, v := range envViews(&h) {
		mark := " "
		if v.Active {
			mark = "*"
		}
		ssh := fmt.Sprintf("%s@%s:%d", v.User, v.Host, v.Port)
		db := v.DBAddress
		if db == "" {
			db = "-"
		} else if v.DBType != "" {
			db = v.DBType + " " + db
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			mark, v.Name, ssh, orDash(v.Zone), orDash(v.Topent), db)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	cmd.Printf("\n* = 默认环境（%s）\n配置文件：%s\n", h.ActiveEnv, path)
	return nil
}

func runEnvShow(cmd *cobra.Command, name string) error {
	path, err := common.ResolveConfig(false)
	if err != nil {
		return err
	}
	root, err := config.Load(path)
	if err != nil {
		return err
	}
	h := root.Hosts
	e := h.ByName(name)
	if e == nil {
		return fmt.Errorf("没有名为 %q 的环境（用 tt env list 看全部）", name)
	}

	if common.JSON {
		v := envViews(&h)
		for i := range v {
			if v[i].Name == e.Name {
				return common.PrintJSON(v[i])
			}
		}
	}

	cmd.Printf("名称      %s\n", e.Name)
	cmd.Printf("SSH       %s@%s:%d\n", e.User, e.Host, e.Port)
	cmd.Printf("口令      %s\n", mask(e.Password))
	cmd.Printf("区域      %s\n", orDash(e.Zone))
	cmd.Printf("企业      %s\n", orDash(string(e.Topent)))
	if e.LaunchArgs != "" {
		cmd.Printf("启动参数  %s\n", e.LaunchArgs)
	}
	if e.WatchdogSeconds > 0 {
		cmd.Printf("停站超时  %d 秒\n", e.WatchdogSeconds)
	}
	if e.DB == nil {
		cmd.Printf("数据库    未配置\n")
		return nil
	}
	cmd.Printf("数据库    %s %s\n", e.DB.Type, e.DB.Address())
	cmd.Printf("只读 SQL  %s\n", onOff(e.DB.ReadonlySQLEnabled()))
	if e.DB.ViaSSH != nil {
		cmd.Printf("隧道      %s@%s:%d\n", e.DB.ViaSSH.User, e.DB.ViaSSH.Host, e.DB.ViaSSH.Port)
	}
	if len(e.DB.Accounts) > 0 {
		names := make([]string, 0, len(e.DB.Accounts))
		for _, a := range e.DB.Accounts {
			names = append(names, a.Account)
		}
		sort.Strings(names)
		cmd.Printf("账号      %s\n", strings.Join(names, ", "))
	}
	return nil
}

// mask 把口令打码，只保留长度信息 —— 终端输出常被贴进 issue 或聊天记录。
func mask(s string) string {
	if s == "" {
		return "（未设置）"
	}
	return strings.Repeat("*", len([]rune(s)))
}

func onOff(b bool) string {
	if b {
		return "开启"
	}
	return "关闭"
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
