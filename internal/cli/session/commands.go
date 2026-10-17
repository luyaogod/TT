package session

// `tt session` 的子命令：建 / 列 / 看 / 改 ENT / 结束本轮 / 收掉 / 透传。
//
// 每一个都只做"参数 → 一次客户端调用 → 输出"，不含编排：需要编排的地方（录制那套）
// 已经在 `internal/cli/fep` 里收成计划了。命令层保持薄，是这套改造的一条底线。

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/debugclient"
)

var (
	// new
	nsEnv, nsEnt, nsName, nsProg, nsModule string
	nsTimeout                              int
	// 定位
	lsSession, lsEnv string
	lsAll            bool
	// 属性
	entValue string
	entClear bool
	// 透传

)

func init() {
	newCmd := &cobra.Command{
		Use:   "new",
		Short: "建会话（可同时起作业）",
		Long: `建一条会话。

不给 --prog 建出来的是**空闲会话**（已登录、无作业）：之后可以 tt session exec / tt debug 操作它，
也可以再用 tt debug start 在里面起作业。
给了 --prog 则同时起作业并等到入口停站。

ENT 必填：企业编号决定用哪个账号连库与据点校验，不再沿用登录 shell 的 $TOPENT。`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			env := strings.TrimSpace(nsEnv)
			ent := strings.TrimSpace(nsEnt)
			if env == "" {
				return failUsage("需要 --env <环境名>（tt env ls 可看全部）")
			}
			if ent == "" {
				return failUsage("需要 --ent <企业编号>（决定用哪个账号连库；tt debug ents 可查）")
			}
			cl, err := newClient()
			if err != nil {
				return err
			}
			prog := strings.TrimSpace(nsProg)
			if prog == "" {
				// 空闲会话：走"切到该环境并命名"，服务端会建一条 idle 的宿主。
				id, serr := cl.SwitchToIdle(env, ent, strings.TrimSpace(nsName), time.Duration(nsTimeout)*time.Second)
				if serr != nil {
					return serr
				}
				return emitSession(cl, id)
			}
			// 有作业：建会话 + 起作业 + 等入口停站。
			id, lerr := cl.Launch(launchReq(env, ent, prog))
			if lerr != nil {
				return lerr
			}
			if _, ok, werr := cl.WaitStopped(id, time.Duration(nsTimeout)*time.Second); werr != nil {
				return werr
			} else if !ok {
				return fmt.Errorf("起作业后没等到入口停站（%s）", prog)
			}
			return emitSession(cl, id)
		},
	}

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "列会话",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := newClient()
			if err != nil {
				return err
			}
			list, lerr := cl.Sessions()
			if lerr != nil {
				return lerr
			}
			out := make([]map[string]any, 0, len(list))
			for _, b := range list {
				if e := strings.TrimSpace(lsEnv); e != "" && b.Env != e {
					continue
				}
				if !lsAll && b.State == "exit" {
					continue
				}
				out = append(out, briefMap(b))
			}
			return emit(map[string]any{"ok": true, "sessions": out})
		},
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "看一条会话的详情（身份 / 作业 / 停站位置 / 附件）",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := newClient()
			if err != nil {
				return err
			}
			b, rerr := resolve(cl, lsSession)
			if rerr != nil {
				return rerr
			}
			snap, serr := cl.Snapshot(b.ID)
			if serr != nil {
				return serr
			}
			return emit(snap)
		},
	}

	entCmd := &cobra.Command{
		Use:   "ent <企业编号>",
		Short: "改会话的企业编号（仅空闲时可设）",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			value := strings.TrimSpace(entValue)
			if len(args) == 1 && value == "" {
				value = strings.TrimSpace(args[0])
			}
			if entClear {
				value = ""
			}
			cl, err := newClient()
			if err != nil {
				return err
			}
			b, rerr := resolve(cl, lsSession)
			if rerr != nil {
				return rerr
			}
			if terr := cl.Topent(b.ID, value); terr != nil {
				return terr
			}
			return emit(map[string]any{"ok": true, "session": b.Ref, "ent": value})
		},
	}

	quitCmd := &cobra.Command{
		Use:   "quit",
		Short: "结束本轮调试（会话保留，回空闲）",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := newClient()
			if err != nil {
				return err
			}
			b, rerr := resolve(cl, lsSession)
			if rerr != nil {
				return rerr
			}
			if qerr := cl.Quit(b.ID); qerr != nil {
				return qerr
			}
			return emit(map[string]any{"ok": true, "session": b.Ref, "state": "idle"})
		},
	}

	rmCmd := &cobra.Command{
		Use:   "rm",
		Short: "收掉会话（断连接、清资源）",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := newClient()
			if err != nil {
				return err
			}
			b, rerr := resolve(cl, lsSession)
			if rerr != nil {
				return rerr
			}
			if cerr := cl.Close(b.ID); cerr != nil {
				return cerr
			}
			return emit(map[string]any{"ok": true, "session": b.Ref, "removed": true})
		},
	}

	execCmd := &cobra.Command{
		Use:   "exec -- \"<fgldb 命令>\"",
		Short: "透传一条 fgldb 命令（等价 tt debug exec）",
		Long: `透传一条 fgldb 原生命令，原样返回输出。

  tt session exec --session A -- "break 812"
  tt session exec --session A -- "print g_qryparam.cond"

-- 之后整段原样交给调试器，不再由 tt 解析。`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			command := strings.TrimSpace(strings.Join(args, " "))
			if command == "" {
				return failUsage(`需要一条命令：tt session exec --session A -- "break 812"`)
			}
			cl, err := newClient()
			if err != nil {
				return err
			}
			b, rerr := resolve(cl, lsSession)
			if rerr != nil {
				return rerr
			}
			lines, xerr := cl.Exec(b.ID, command)
			if xerr != nil {
				return xerr
			}
			return emit(map[string]any{"ok": true, "session": b.Ref, "command": command, "lines": lines})
		},
	}

	newCmd.Flags().StringVar(&nsEnv, "env", "", "环境名（必填）")
	newCmd.Flags().StringVar(&nsEnt, "ent", "", "企业编号（必填）")
	newCmd.Flags().StringVar(&nsName, "name", "", "别名（跨轮追踪用它；1–64 字符，禁 / 与 #）")
	newCmd.Flags().StringVar(&nsProg, "prog", "", "作业程序名（给了就同时起作业）")
	newCmd.Flags().StringVarP(&nsModule, "module", "m", "", "T100 模块目录名（留空由作业名自动解析）")
	newCmd.Flags().IntVar(&nsTimeout, "timeout", 120, "等待空闲 / 入口停站的超时（秒）")

	lsCmd.Flags().StringVar(&lsEnv, "env", "", "只看这个环境")
	lsCmd.Flags().BoolVarP(&lsAll, "all", "a", false, "含已断开的会话")

	showCmd.Flags().StringVar(&lsSession, "session", "", "目标会话（ref / 别名 / 句柄）")

	entCmd.Flags().StringVar(&entValue, "value", "", "企业编号（也可用位置参数）")
	entCmd.Flags().BoolVar(&entClear, "clear", false, "清除会话级覆盖（回到登录默认）")
	entCmd.Flags().StringVar(&lsSession, "session", "", "目标会话（ref / 别名 / 句柄）")

	quitCmd.Flags().StringVar(&lsSession, "session", "", "目标会话（ref / 别名 / 句柄）")
	rmCmd.Flags().StringVar(&lsSession, "session", "", "目标会话（ref / 别名 / 句柄）")
	execCmd.Flags().StringVar(&lsSession, "session", "", "目标会话（ref / 别名 / 句柄）")

	group.AddCommand(newCmd, lsCmd, showCmd, entCmd, quitCmd, rmCmd, execCmd)
}

// launchReq 组装"建会话 + 起作业"的请求（ENT 必填，模块可留空）。
func launchReq(env, ent, prog string) debugclient.LaunchReq {
	return debugclient.LaunchReq{
		Env: env, Ent: ent, Prog: prog,
		Module: strings.TrimSpace(nsModule),
		Name:   strings.TrimSpace(nsName),
	}
}

// emitSession 按句柄取一次会话并输出（建完/改完之后用它回显"落在哪条"）。
func emitSession(cl *debugclient.Client, id string) error {
	list, err := cl.Sessions()
	if err != nil {
		return err
	}
	for _, b := range list {
		if b.ID == id {
			return emit(map[string]any{"ok": true, "session": briefMap(b)})
		}
	}
	// 建完立刻列不到（极少数竞态）：至少把句柄给出来，别让调用方以为没建成。
	return emit(map[string]any{"ok": true, "session": map[string]any{"id": id}})
}
