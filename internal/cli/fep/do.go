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

哪些子命令**要目标**（契约 A28：无目标的动作不受"目标 / --field 互斥"检查）：
  act / type   要 —— 名字 / #id，或 --field <colName>，二选一（同时给是用法错 2）
  key          不要 —— KeyEvent 不带 idRef，没有可指的对象；
               给了 --field 就是用法错 2（明确拒绝，不静默忽略）

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
	Use:   "act [名字|#id]",
	Short: "点一个动作对象（工具栏项 / 按钮 / 开窗按钮 / 单元格）",
	Long: `点一个动作对象。目标二选一：
  位置参数  中文名或 #id（照旧）
  --field   字段工号 colName（契约 C3：运行期用 (colName, fieldId) 双键定位，
            每步现解析、用完即弃；解析不过就报四码并不发包）

两者**互斥**：同时给是用法错（退出码 2）。`,
	Args: targetOrFieldArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dc := fep.DoConfig{Kind: "click"}
		if len(args) > 0 {
			dc.Target = args[0]
		}
		return runDo(cmd, dc)
	},
}

var doKeyCmd = &cobra.Command{
	Use:   "key <键名>",
	Short: "按一个键（Tab / Return 等；KeyEvent 不带 idRef）",
	Long: `按一个键（Tab / Return / 功能键等）。

**这条命令没有目标对象**（契约 A28）：KeyEvent 不带 idRef，没有可按身份点名的东西，
所以它不写名字/#id，也不收 --field —— 给了 --field 是**用法错（退出码 2）**，
明确拒绝而不是静默忽略（默默忽略会让人以为"按身份点名成功"了）。

它也是"写完栏位主动离开"那一步的底层动作：type 写完栏位默认就发本命令
（KeyEvent Tab），AFTER FIELD / 自动带值只在光标离开那一刻才跑。`,
	Args: exactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDo(cmd, fep.DoConfig{Kind: "key", KeyName: args[0]})
	},
}

var doTypeCmd = &cobra.Command{
	Use:   "type [名字|#id] <值>",
	Short: "往一个栏位输入值（优先发给该字段的 FormField）",
	Long: `往一个栏位输入值。目标二选一：
  位置参数           中文名或 #id（照旧）
  --field <colName>  字段工号（本命令下位置参数只剩**值**一个：
                     tt fep do type --field pmdl002 "A50"）

两者**互斥**：同时给名字与 --field 是用法错（退出码 2）。

写完默认**主动离开这个栏位**（发 KeyEvent Tab）：AFTER FIELD（校验 / 自动带值 /
下一栏位的联动）只在光标离开那一刻才跑，不离开就看不见系统自己做了什么。
--no-commit 关掉它（只要值、不要离开）。`,
	Args: valueAfterFieldArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dc := fep.DoConfig{Kind: "input", CommitKey: commitKeyFor("input")}
		if len(args) == 2 {
			dc.Target, dc.Value = args[0], args[1]
		} else if len(args) == 1 {
			dc.Value = args[0]
		}
		return runDo(cmd, dc)
	},
}

// runDo 是三个动词的共同通路：读登记 → 建 Client → 校验并（或）发送 → 写审计。
func runDo(cmd *cobra.Command, dc fep.DoConfig) error {
	dc.Conn, dc.DryRun = doConn, doDryRun
	// 互斥判定放在最前面：参数不成立是**用法错**（2），与"有没有会话"无关 ——
	// 否则没会话时会把一个敲错的命令行读成"没有会话"，排障方向整个跑偏。
	//
	// 但**只有要目标的动作**才适用它（契约 A28）：`key`（KeyEvent，不带 idRef）
	// 没有目标对象，交给 checkFieldExclusive 按 kind 判 —— 这里不许自己加条件，
	// 判据只有一处（needsTarget）。
	if err := checkFieldExclusive(dc.Kind, dc.Target, doField); err != nil {
		return err
	}
	reg, _, err := registry()
	if err != nil {
		return err
	}
	s, err := reg.Load()
	if errors.Is(err, fep.ErrNoSession) {
		return fail(exitNoSession, "没有登记中的会话（先跑 tt recording start --launch standard）")
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
	// --field：在这里落成 `#<id>`（解析不过就带四码 + Hint 退出，一个字节都不发）。
	var fieldCol fep.ColName
	if strings.TrimSpace(doField) != "" {
		if fieldCol, err = applyFieldFlag(cmd, &dc, client); err != nil {
			return err
		}
	}
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
	// 审计里要留下"这一步是按哪把工号点的"：事后复盘时 `#6000` 本身说明不了身份。
	if fieldCol != "" {
		entry.Note = strings.TrimSpace(entry.Note + " --field " + string(fieldCol))
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
	// --field 挂在三个动词上：`act` / `type` 用它按身份点名；
	// `key` 挂上是**为了明确拒绝**（契约 A28）—— 不挂的话 cobra 只会回一句
	// `unknown flag: --field`，讲不清"为什么不行"，用户会以为是自己敲错了。
	for _, c := range []*cobra.Command{doActCmd, doTypeCmd} {
		c.Flags().StringVar(&doField, "field", "",
			"按字段工号 colName 点名目标（与位置参数互斥；解析不过就报四码并不发包）")
	}
	doKeyCmd.Flags().StringVar(&doField, "field", "",
		"key 没有目标对象：给了 --field 就是用法错（显式拒绝，不静默忽略）")
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
