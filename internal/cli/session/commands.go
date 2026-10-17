package session

// `tt session` 的子命令：建 / 列 / 看 / 改 ENT / 结束本轮 / 收掉 / 透传。
//
// 每一个都只做"参数 → 一次客户端调用 → 输出"，不含编排：需要编排的地方（录制那套）
// 已经在 `internal/cli/fep` 里收成计划了。命令层保持薄，是这套改造的一条底线。

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/debugclient"
	"tt/internal/output"
	"tt/internal/winproc"
)

var (
	// new
	nsEnv, nsEnt, nsName, nsProg, nsModule string
	nsTimeout                              int
	nsRecord                               bool // --record：建完会话顺手把录制挂上去
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

			if nsRecord {
				// 「建会话 + 起录制」一步到位：先建**空闲**会话（拿身份），再把录制挂上去。
				// 顺序不能反 —— 录制挂靠要一条已有会话，而它靠别名落回同一条。
				if prog == "" {
					return failUsage("--record 需要 --prog：录的是作业界面，没有作业就没什么可录")
				}
				if strings.TrimSpace(nsName) == "" {
					return failUsage("--record 需要 --name：录制靠别名落回同一条会话，没名字挂不回来")
				}
				id, serr := cl.SwitchToIdle(env, ent, strings.TrimSpace(nsName), time.Duration(nsTimeout)*time.Second)
				if serr != nil {
					return serr
				}
				cfgPath, cerr := common.ResolveConfig(true)
				if cerr != nil {
					return cerr
				}
				pid, logPath, rerr := startRecording(cfgPath, id, prog, ent)
				if rerr != nil {
					return fmt.Errorf("会话已建好（%s），但录制没起来：%w", id, rerr)
				}
				return emit(map[string]any{
					"ok": true, "session": id,
					"recording": map[string]any{"pid": pid, "log": logPath, "prog": prog},
					"next":      "tt recording ls 看这一轮的录制 ID；tt recording stop <ID> 收工",
				})
			}
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
			var out []map[string]any
			var rows [][]string
			for _, b := range list {
				if e := strings.TrimSpace(lsEnv); e != "" && b.Env != e {
					continue
				}
				if !lsAll && b.State == "exit" {
					continue
				}
				out = append(out, briefMap(b))
				rows = append(rows, []string{b.Ref, b.Name, b.Env, b.Ent, b.Prog, b.State, b.Epoch})
			}
			payload := map[string]any{"ok": true, "sessions": out}
			headers := []string{"REF", "名字", "环境", "ENT", "作业", "状态", "世代"}
			switch common.OutputFormat() {
			case output.FormatJSON:
				return emit(payload)
			case output.FormatCSV:
				// 自己写 CSV：`output.WriteCSV` 会带上 SQL 那一套表头注释（"只读事务 / 行数 0"），
				// 对会话列表是误导 —— 这里明明有行，注释却说 0 行。
				cw := csv.NewWriter(os.Stdout)
				_ = cw.Write(headers)
				for _, row := range rows {
					_ = cw.Write(row)
				}
				cw.Flush()
				return cw.Error()
			default:
				return output.WriteTable(os.Stdout, headers, rows)
			}
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
			// 附件：画面通道与录制（只读视图，来自接线层的查询）。
			// 取不到不算失败 —— 老服务没有这两个端点，而"看会话"这件事本身仍然成立。
			if scr, xerr := cl.Screen(b.ID); xerr == nil {
				snap["screen"] = scr
			}
			if rec, xerr := cl.Recording(b.ID); xerr == nil {
				snap["recording"] = rec
			}
			return emitFormatted(snap, func() {
				fmt.Printf("会话  : %s（%s）\n", b.Ref, b.Name)
				fmt.Printf("环境  : %s / ENT %s\n", b.Env, b.Ent)
				fmt.Printf("作业  : %s（%s）\n", b.Prog, b.State)
				if scr, ok := snap["screen"].(map[string]any); ok {
					if tag, ok := scr["tag"].(string); ok && tag != "" {
						fmt.Printf("画面  : %s（%v）\n", tag, scr["kind"])
					}
				}
			})
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
	newCmd.Flags().BoolVar(&nsRecord, "record", false, "建完会话顺手把录制挂上去（需要 --name 与 --prog）")

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

// startRecording 起一轮录制并**挂到指定会话**上（`session new --record` 的那一步）。
//
// 为什么是"起一个脱离进程"而不是在本进程里录：
//   - 录制要一直活着，直到用户按停止（或 `tt recording stop`）；命令本身应当立刻返回；
//   - 它持有的东西（GDC 实例、代理、登记）本来就设计成**脱离本进程**也能收尾
//     （见 internal/fep 的 register-first 与 `recording stop`）。
//
// 抽成变量是为了在测试里替换它：真起进程需要 GDC 与真实环境，而这里要钉的是
// "命令按什么参数起它、失败怎么报"。
var startRecording = func(cfgPath, sessionRef, prog, ent string) (int, string, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, "", err
	}
	dir, err := os.MkdirTemp("", "tt-rec-")
	if err != nil {
		return 0, "", err
	}
	logPath := filepath.Join(dir, "rec-start.log")
	args := []string{"recording", "start", "--session", sessionRef, "--prog", prog}
	if e := strings.TrimSpace(ent); e != "" {
		args = append(args, "--ent", e)
	}
	args = append(args, "--config", cfgPath)
	// TT_ACTOR=human 不传：这是 CLI 起的（AI 或人都有可能），沿用调用者自己的身份。
	pid, perr := winproc.SpawnDetached(exe, args, logPath)
	if perr != nil {
		return 0, logPath, perr
	}
	for waited := time.Duration(0); waited < 6*time.Second; waited += 200 * time.Millisecond {
		time.Sleep(200 * time.Millisecond)
		if !winproc.Alive(pid) {
			return pid, logPath, fmt.Errorf("录制进程起来就退了（pid %d）：详见 %s", pid, logPath)
		}
	}
	return pid, logPath, nil
}
