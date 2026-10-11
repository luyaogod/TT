package fep

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var (
	doConn   int
	doDryRun bool
	doWindow string
	doWait   time.Duration
)

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "动作：点一下、按个键、往栏位里输入（发前校验，不发非法事件）",
	Long: `把"我要点这个/我要在这里输入"变成一条协议事件发出去。

发前校验全部在本地做完，**校验不过就一个字节都不发**：目标必须存在、类型必须能接受
这个动作、actionActive 必须为真、_om 序号必须由注入方自己推进（重号会打断连接）。

目标可以给名字（只认**高置信度**的名字，来自协议解析）或 #id。一个名字在一屏里往往
命中多个对象（Label / ButtonEdit / FormField 都叫「客户编号」），这时按动作类型收敛：
点击只认可点击的、输入只认可写的；仍不唯一就报歧义让你用 #id 指定，绝不自己挑。

每一步都写审计（动作 / 连接 / 目标 / 事件 / 是否真发），不管成没成。

--dry-run 只校验与打印，一个字节都不发（照着它检查完再真跑）。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return fail(exitUsage, "未知子命令 %q", args[0])
	},
}

var doActCmd = &cobra.Command{
	Use:   "act <名字|#id>",
	Short: "点一个动作对象（工具栏项 / 按钮 / 开窗按钮 / 单元格）",
	Args:  exactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDo(cmd, fep.DoConfig{Kind: "click", Target: args[0]})
	},
}

var doKeyCmd = &cobra.Command{
	Use:   "key <键名>",
	Short: "按一个键（Tab / Return 等；KeyEvent 不带 idRef）",
	Args:  exactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDo(cmd, fep.DoConfig{Kind: "key", KeyName: args[0]})
	},
}

var doTypeCmd = &cobra.Command{
	Use:   "type <名字|#id> <值>",
	Short: "往一个栏位输入值（优先发给该字段的 FormField）",
	Args:  exactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDo(cmd, fep.DoConfig{Kind: "input", Target: args[0], Value: args[1], CommitKey: commitKeyFor("input")})
	},
}

// runDo 是三个动词的共同通路：读登记 → 建 Client → 校验并（或）发送 → 写审计。
func runDo(cmd *cobra.Command, dc fep.DoConfig) error {
	dc.Conn, dc.DryRun = doConn, doDryRun
	reg, _, err := registry()
	if err != nil {
		return err
	}
	s, err := reg.Load()
	if errors.Is(err, fep.ErrNoSession) {
		return fail(exitNoSession, "没有登记中的会话（先跑 tt fep session open）")
	}
	if err != nil {
		return fail(exitFailed, "%v", err)
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	client.Window = doWindow
	dc.Wait = doWait
	res, derr := fep.Do(client, dc)

	entry := fep.AuditEntry{Verb: dc.Kind, DryRun: doDryRun, Session: s.Prog}
	if res != nil {
		entry.Conn, entry.Target, entry.Name, entry.Event, entry.Sent = res.Conn, res.Target, res.Name, res.Event, res.Sent
	}
	switch {
	case derr == nil && res != nil && res.Sent:
		entry.Result = "sent"
	case derr == nil:
		entry.Result = "dry-run"
	default:
		entry.Result = "refused: " + derr.Error()
	}
	if res != nil && res.Suggest != "" {
		entry.Note = res.Suggest
	}
	if res != nil && len(res.Warnings) > 0 {
		entry.Note = strings.Join(res.Warnings, "；")
	}
	auditErr := fep.AppendAudit(filepath.Join(recDirForAudit(s), "audit.log"), entry)

	if derr != nil {
		var ref *fep.Refusal
		if errors.As(derr, &ref) {
			return fail(ref.Exit, "%s", ref.Message)
		}
		return fail(exitFailed, "%v", derr)
	}
	if auditErr != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "· 审计写不进去（不影响这一步）：%v\n", auditErr)
	}
	return printJSONOrText(res, func() {
		what := "已发"
		if !res.Sent {
			what = "未发（dry-run）"
		}
		target := res.Target
		if res.Name != "" {
			target = res.Name + " (" + res.Target + ")"
		}
		fmt.Printf("%s #%d：%s\n", what, res.Conn, res.Event)
		fmt.Printf("  目标 %s\n", target)
		if res.NotSent != "" {
			fmt.Printf("  %s\n", res.NotSent)
		}
		if res.Suggest != "" {
			fmt.Printf("  提示：%s\n", res.Suggest)
		}
		for _, w := range res.Warnings {
			fmt.Printf("  ⚠ %s\n", w)
		}
		if ch := res.Changes; ch != nil {
			printImpact(ch, doWait.String())
		}
	})
}

// recDirForAudit 返回审计日志的落点（与录制件同目录）。
func recDirForAudit(s *fep.Session) string {
	if s != nil && s.RecordDir != "" {
		return s.RecordDir
	}
	if d, err := cacheDir(); err == nil {
		return d
	}
	return "."
}

func init() {
	doCmd.AddCommand(doActCmd, doKeyCmd, doTypeCmd, doRoleCmd)
	for _, c := range []*cobra.Command{doActCmd, doKeyCmd, doTypeCmd} {
		c.Flags().IntVar(&doConn, "conn", 0, "连接号（多窗口作业必给；给了 --window 就按窗口路由）")
		c.Flags().BoolVar(&doDryRun, "dry-run", false, "只校验与打印，一个字节都不发")
		c.Flags().StringVar(&doWindow, "window", "", "只在某个窗口名里找目标（并按它路由连接）")
		c.Flags().DurationVar(&doWait, "wait", 10*time.Second, "发完等运行期回包多久（负数=不等）")
	}
	Group.AddCommand(doCmd)
}

// doCommitOff 关掉"写完栏位后主动离开"（默认开）。
var doCommitOff bool

// commitKeyFor 返回"写完这个栏位后用什么键离开它"。
//
// 默认 Tab —— 依据是真机语料：键盘离开栏位时前端发的就是 `KeyEvent "Tab"`（大量出现）。
// 不离开的后果：AFTER FIELD 不跑，自动带值（如供应商编号 → 供应商名称）看不见。
// 用户明确要求："填写一个字段后，应该主动离开这个字段，不然怎么触发 AFTER FIELD"。
func commitKeyFor(kind string) string {
	if doCommitOff {
		return ""
	}
	switch kind {
	case "input", "pick":
		return "Tab"
	}
	return ""
}

func init() {
	doCmd.PersistentFlags().BoolVar(&doCommitOff, "no-commit", false,
		"写完栏位后**不**主动离开（默认会发 KeyEvent Tab 离开：只有离开才触发 AFTER FIELD 与自动带值）")
}
