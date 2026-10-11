package fep

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var doCellIn string

// doCellCmd：往网格格子里写值。
//
// 为什么单独一个命令（实测）：前端写格子时**同一条消息**里发两个事件 ——
//
//	ConfigureEvent{idRef <格子 Value>}{value "…"}
//	ConfigureEvent{idRef <该列 TableColumn>}{cursor "N"}{cursor2 "N"}
//
// 我先前拆成两条单独发，运行期不认（写了但作业没变）✗。所以这里照前端的形态发。
var doCellCmd = &cobra.Command{
	Use:   "cell <列角色|#列id> <值>",
	Short: "往网格格子写值（值与光标合成一条消息，照前端的形态发）",
	Long: `tt fep do cell 'pmdn_t.pmdn007' 1000     # 采购数量
tt fep do cell 'pmdn_t.pmdn010' 10       # 单价

列角色从 tt fep ui roles --kind TableColumn 拿（带源码时就是源码字段名）。
格子 = 该列 ValueList 下的第一个 Value；值与光标同一条消息发出，
与前端落键时的形态完全一致。`,
	Args: exactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, r, _, err := loadModel(uiFrom)
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		cat := fep.Catalog(m, r)
		col, cands, rerr := fep.ResolveRole(cat, args[0])
		if rerr != nil {
			for _, c := range cands {
				fmt.Fprintf(cmd.ErrOrStderr(), "  候选 %-44s %s #%d\n", c.RolePath(), c.Kind, c.ID)
			}
			return fail(exitUsage, "%v", rerr)
		}
		if col.Kind != "TableColumn" {
			return fail(exitUsage, "%s 是 %s，不是网格列；格子要写在某一列上", col.RolePath(), col.Kind)
		}
		cell, cerr := fep.FirstCellOf(m, col.ID)
		if cerr != nil {
			return fail(exitTarget, "%v", cerr)
		}
		ev := fep.ConfigureValueText(cell, args[1], nil) + "\n\t" +
			fep.ConfigureCursorText(col.ID, len([]rune(args[1])))
		// 去掉换行/缩进：事件体本身不能有空白分隔错位
		ev = "ConfigureEvent 0" + ev[len("ConfigureEvent 0"):]

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
		if doCellIn != "" {
			client.Window = doCellIn
		}
		res, derr := fep.Do(client, fep.DoConfig{
			Kind: "raw", RawEvent: ev, Target: fmt.Sprintf("#%d", col.ID),
			Conn: doConn, DryRun: doDryRun, Wait: doWait,
		})
		eff := map[string]any{
			"protocol": "tt-aui/1", "op": "cell", "column": col.RolePath(),
			"cell": cell, "value": args[1], "session": s.Prog,
		}
		if derr != nil {
			var ref *fep.Refusal
			if errors.As(derr, &ref) {
				return fail(ref.Exit, "%s", ref.Message)
			}
			return fail(exitFailed, "%v", derr)
		}
		if res.Changes == nil || (len(res.Changes.Values) == 0 && len(res.Changes.NewWindows) == 0) {
			eff["effect"] = "noEffect"
			if res.Changes != nil {
				eff["runtimeBytes"] = res.Changes.RuntimeBytes
			}
			eff["reason"] = "运行期校验后没有改变任何 AUI 节点"
		} else {
			eff["effect"] = "applied"
			eff["valueChanged"] = res.Changes.Values
		}
		_ = fep.AppendAudit(filepath.Join(recDirForAudit(s), "audit.log"), fep.AuditEntry{
			Verb: "cell", DryRun: doDryRun, Session: s.Prog, Result: eff["effect"].(string),
		})
		return printJSONOrText(eff, func() {
			fmt.Printf("cell %s ← %q（格子 #%d）\n", col.RolePath(), args[1], cell)
			switch eff["effect"] {
			case "applied":
				for _, vc := range res.Changes.Values {
					label := vc.Name
					if label == "" {
						label = fmt.Sprintf("#%d", vc.ID)
					}
					fmt.Printf("  效果：值变化 %s %q → %q\n", label, vc.From, vc.To)
				}
			default:
				fmt.Printf("  效果：无（%v）\n", eff["reason"])
			}
		})
	},
}

func init() {
	doCellCmd.Flags().StringVar(&doCellIn, "in", "", "限定窗口（路由用）")
	doCmd.AddCommand(doCellCmd)
}
