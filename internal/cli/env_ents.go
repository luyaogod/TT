package cli

// tt env ents —— 一个环境“支持哪些企业（ENT）、各用哪个数据库账号”。
//
// 它是 --at 取值的**盘点入口**：--at-list 给的是“这条命令能填什么”，这条给的是
// “这台机器上总共有哪些 (环境, 企业) 组合”。两者读同一份数据（config.json + entdir 快照），
// 所以不会一个说有一个说没有。
//
// 环境一律点名：不给环境名 = 逐个环境都列（不是挑一个默认）。某个环境读不到只影响它自己，
// 其余照常输出，最后用退出码 3 表示“至少有一个没取到”。

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/dbconfig"
	"tt/internal/debug"
	"tt/internal/output"
)

type envEntsView struct {
	Env        string             `json:"env"`
	Zone       string             `json:"zone,omitempty"`
	Dialect    string             `json:"dialect,omitempty"`
	Target     string             `json:"target,omitempty"`
	Account    string             `json:"account,omitempty"` // 查 gzou_t 用的系统账号（恒 ds）
	Source     string             `json:"source,omitempty"`
	Stale      bool               `json:"stale,omitempty"`
	FetchedAt  string             `json:"fetchedAt,omitempty"`
	AgeSeconds int                `json:"ageSeconds,omitempty"`
	TTLSeconds int                `json:"ttlSeconds,omitempty"`
	Count      int                `json:"count"`
	Ents       []debug.EntMapping `json:"ents,omitempty"`
	Error      string             `json:"error,omitempty"`
}

func newEnvEntsCmd() *cobra.Command {
	var (
		refresh bool
		cached  bool
		onlyEnt int
	)
	cmd := &cobra.Command{
		Use:   "ents [<环境名>]",
		Short: "列出环境支持的企业编号（ENT）与账号（--at 的取值来源）",
		Long: `列出环境在 gzou_t 里启用的企业编号（gzoustus='Y' 的 gzou001），以及各自的
数据库账号（gzou003）。

  tt env ents                   # 逐个环境列（不是挑一个默认）
  tt env ents 主机109           # 只看这个环境
  tt env ents --ent 99          # 只留含企业 99 的环境
  tt env ents --cached          # 只用快照，完全不连网
  tt env ents --refresh         # 跳过缓存与快照，强制现查
  tt env ents --json            # 结构化

清单固定用系统账号 ds 读 gzou_t；结果按环境落一份快照（数据目录 ents/ 下，与
tt debug ents / tt dict 同一份），10 分钟内再问秒回。环境一律点名：不给环境名
就是逐个环境都列 —— 某个环境读不到只影响它自己（退出码 3）。`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if refresh && cached {
				return output.Errorf("USAGE", 2, "--cached 与 --refresh 只能选一个")
			}
			name := ""
			if len(args) == 1 {
				name = strings.TrimSpace(args[0])
			} else if strings.TrimSpace(common.Env) != "" {
				name = strings.TrimSpace(common.Env)
			}
			return runEnvEnts(name, refresh, cached, onlyEnt)
		},
	}
	cmd.Flags().BoolVar(&refresh, "refresh", false, "跳过缓存与快照，强制现查 gzou_t")
	cmd.Flags().BoolVar(&cached, "cached", false, "只用快照，完全不连网（没有快照的环境会标出来）")
	cmd.Flags().IntVar(&onlyEnt, "ent", 0, "只留含这个企业编号的环境（0 = 全部）")
	return cmd
}

func runEnvEnts(name string, refresh, cached bool, onlyEnt int) error {
	path, err := common.ResolveConfig(false)
	if err != nil {
		return err
	}
	root, err := config.Load(path)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(root.Hosts.SSHs))
	if name != "" {
		if root.Hosts.ByName(name) == nil {
			return output.Errorf("ENV_UNKNOWN", 2, "没有名为 %q 的环境（用 tt env list 看全部）", name)
		}
		names = append(names, name)
	} else {
		for _, e := range root.Hosts.SSHs {
			names = append(names, e.Name)
		}
	}
	if len(names) == 0 {
		return output.Errorf("ENV_NONE", 2, "配置里还没有任何环境（%s）；先 tt serve 到设置页添加", path)
	}

	cfg, err := debug.LoadConfig(path)
	if err != nil {
		return err
	}
	cfg.DataDir = filepath.Dir(path)

	views := make([]envEntsView, 0, len(names))
	failed := 0
	for _, n := range names {
		v := envEntsView{Env: n, Account: dbconfig.SystemAccount}
		sub := cfg.CloneEnv(n)
		if sub == nil {
			v.Error = "环境不存在"
			failed++
			views = append(views, v)
			continue
		}
		sub.DataDir = cfg.DataDir
		l, err := debug.LookupEnts(sub, debug.EntListOpt{Refresh: refresh, Cached: cached})
		if err != nil {
			v.Error = err.Error()
			failed++
			views = append(views, v)
			continue
		}
		v.Zone, v.Dialect, v.Target = l.Zone, l.Dialect, l.Target
		v.Source, v.Stale = l.Source, l.Stale
		v.FetchedAt = l.FetchedAt.Format(time.RFC3339)
		v.AgeSeconds, v.TTLSeconds = l.AgeSeconds, l.TTLSeconds
		for _, m := range l.Ents {
			if onlyEnt > 0 && m.Ent != onlyEnt {
				continue
			}
			v.Ents = append(v.Ents, m)
		}
		v.Count = len(v.Ents)
		if onlyEnt > 0 && v.Count == 0 {
			continue // 这个环境没有那个企业：不列（--ent 的语义是“筛”）
		}
		views = append(views, v)
	}

	if common.JSON {
		env := map[string]any{
			"ok": failed == 0, "count": len(views), "failed": failed, "envs": views,
		}
		if err := common.PrintJSON(env); err != nil {
			return err
		}
	} else {
		writeEnvEnts(os.Stdout, views)
	}

	if failed > 0 {
		return partialEnvErr{failed: failed}
	}
	return nil
}

func writeEnvEnts(w *os.File, views []envEntsView) {
	for i, v := range views {
		if i > 0 {
			fmt.Fprintln(w)
		}
		if v.Error != "" {
			fmt.Fprintf(w, "!! 环境 %s 读取失败：%s\n", v.Env, v.Error)
			continue
		}
		fresh := v.Source
		if v.Stale {
			fresh += "（已过期）"
		} else if v.AgeSeconds > 0 {
			fresh += fmt.Sprintf("（%d 分钟前）", v.AgeSeconds/60)
		}
		meta := []string{"区域 " + orDash(v.Zone)}
		if v.Dialect != "" {
			meta = append(meta, v.Dialect)
		}
		if v.Target != "" {
			meta = append(meta, v.Target)
		}
		fmt.Fprintf(w, "环境 %s   %s   %s\n", v.Env, strings.Join(meta, " · "), fresh)
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  ENT\t账号")
		for _, m := range v.Ents {
			acct := m.Account
			if acct == "" || acct == "-" {
				acct = "-（未配账号）"
			}
			fmt.Fprintf(tw, "  %d\t%s\n", m.Ent, acct)
		}
		_ = tw.Flush()
		fmt.Fprintf(w, "  共 %d 个企业；--at 的值：%s@<上表中的企业>\n", v.Count, v.Env)
	}
}

// partialEnvErr 表示“已经逐个环境报过了，只是有失败”。AlreadyReported 让根命令
// 不再重复打印一条错误（否则 stdout 上会出现两份说法）。
type partialEnvErr struct{ failed int }

func (e partialEnvErr) Error() string {
	return fmt.Sprintf("%d 个环境没取到企业目录", e.failed)
}
func (e partialEnvErr) ExitCode() int         { return 3 }
func (e partialEnvErr) AlreadyReported() bool { return true }
