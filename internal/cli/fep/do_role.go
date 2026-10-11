package fep

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var (
	doRoleSet    string
	doRoleChoose string
	doRoleOpen   bool
	doRoleInvoke bool
	doRoleIn     string
)

// doRoleCmd：按**角色**操作一个控件。
//
// 与 do act/type 的区别是"不猜"：
//   - 目标是角色（窗口/角色），不是名字或 id —— 命中多个就报候选，不替调用方选
//   - 操作由控件**类型**决定（见 internal/fep/roles.go），类型不支持的直接拒绝
//   - 结果报**结构化效果**，并且把"运行期校验后什么都没变"（noEffect）当成一等结果 ——
//     实测过一次 accept 只回 18 字节的假成功，不报出来就会被当成做成了
var doRoleCmd = &cobra.Command{
	Use:   "role <窗口/角色>",
	Short: "按角色操作控件（操作集由控件类型决定；不支持就拒绝；结果报结构化效果）",
	Long: `tt fep do role '采购单维护作业/供应商编号' --open
tt fep do role '采购单维护作业/采购日期' --set 2026/10/11

角色从 tt fep ui roles 拿。每个控件的支持操作由它的类型决定：
Edit=set、ButtonEdit=set+open、ComboBox/Item=choose、CheckBox=toggle、
Action/ToolBarItem/MenuAction=invoke、TableColumn=cell。`,
	Args: exactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		op, _, err := pickRoleOp()
		if err != nil {
			return fail(exitUsage, "%v", err)
		}
		m, r, _, err := loadModel(uiFrom)
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		cat := fep.Catalog(m, r)
		ctl, cands, rerr := fep.ResolveRole(cat, args[0])
		if rerr != nil {
			if len(cands) > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "候选（共 %d 个）：\n", len(cands))
				for _, c := range cands {
					fmt.Fprintf(cmd.ErrOrStderr(), "  %-40s %s #%d\n", c.RolePath(), c.Kind, c.ID)
				}
			}
			return fail(exitUsage, "%v", rerr)
		}
		if !hasOp(ctl.Ops, op) {
			return fail(exitUsage, "%s 是 %s，不支持 %s；它支持：%s",
				ctl.RolePath(), ctl.Kind, op, strings.Join(ctl.Ops, ","))
		}

		dc := fep.DoConfig{Kind: "click", Target: fmt.Sprintf("#%d", ctl.ID)}
		switch op {
		case "set":
			dc.Kind, dc.Value = "input", doRoleSet
		case "choose":
			itemID, itemName, ferr := findChoice(m, r, ctl.ID, doRoleChoose)
			if ferr != nil {
				return fail(exitTarget, "%v", ferr)
			}
			dc.Target = fmt.Sprintf("#%d", itemID)
			ctl.Value = itemName
		}

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
		if doRoleIn != "" {
			client.Window = doRoleIn
		}
		dc.Conn, dc.DryRun, dc.Wait = doConn, doDryRun, doWait
		res, derr := fep.Do(client, dc)

		eff := map[string]any{
			"protocol": "tt-aui/1",
			"op":       op,
			"target":   ctl.RolePath(),
			"kind":     ctl.Kind,
			"session":  s.Prog,
		}
		if derr != nil {
			var ref *fep.Refusal
			if errors.As(derr, &ref) {
				eff["effect"] = "refused"
				eff["reason"] = ref.Message
				_ = printJSONOrText(eff, func() { fmt.Printf("拒绝：%s\n", ref.Message) })
				return fail(ref.Exit, "%s", ref.Message)
			}
			return fail(exitFailed, "%v", derr)
		}
		if res.Changes == nil {
			eff["effect"] = "noEffect"
			eff["reason"] = "运行期没有回包"
		} else {
			ch := res.Changes
			if len(ch.NewWindows) > 0 {
				eff["windowsOpened"] = ch.NewWindows
			}
			if len(ch.Values) > 0 {
				eff["valueChanged"] = ch.Values
			}
			if len(ch.NewWindows) == 0 && len(ch.Values) == 0 {
				eff["effect"] = "noEffect"
				eff["runtimeBytes"] = ch.RuntimeBytes
				eff["reason"] = "运行期校验后没有改变任何 AUI 节点（事件被忽略）"
			} else {
				eff["effect"] = "applied"
			}
		}
		_ = fep.AppendAudit(filepath.Join(recDirForAudit(s), "audit.log"), fep.AuditEntry{
			Verb: "role:" + op, DryRun: doDryRun, Session: s.Prog, Result: eff["effect"].(string),
		})
		return printJSONOrText(eff, func() {
			fmt.Printf("%s %s（%s #%d）\n", op, ctl.RolePath(), ctl.Kind, ctl.ID)
			switch eff["effect"] {
			case "applied":
				if w, ok := eff["windowsOpened"]; ok {
					fmt.Printf("  效果：新窗口 %v\n", w)
				}
				if v, ok := eff["valueChanged"]; ok {
					for _, vc := range v.([]fep.ValueChange) {
						label := vc.Name
						if label == "" {
							label = fmt.Sprintf("#%d", vc.ID)
						}
						fmt.Printf("  效果：值变化 %s %q → %q\n", label, vc.From, vc.To)
					}
				}
			case "noEffect":
				fmt.Printf("  效果：无（%v）—— 事件发出去了，运行期校验后什么都没变\n", eff["reason"])
			}
		})
	},
}

func pickRoleOp() (string, string, error) {
	n := 0
	op, arg := "", ""
	if doRoleSet != "" {
		op, arg, n = "set", doRoleSet, n+1
	}
	if doRoleOpen {
		op, n = "open", n+1
	}
	if doRoleInvoke {
		op, n = "invoke", n+1
	}
	if doRoleChoose != "" {
		op, arg, n = "choose", doRoleChoose, n+1
	}
	if n != 1 {
		return "", "", errors.New("要且只要一个操作：--set <值> / --open / --invoke / --choose <候选项>")
	}
	return op, arg, nil
}

func hasOp(ops []string, want string) bool {
	for _, o := range ops {
		if o == want {
			return true
		}
	}
	return false
}

// findChoice 在某个控件的子对象里按文字找候选项（Item/Value/TreeItem）。
func findChoice(m *fep.Model, r *fep.Resolver, parent int, text string) (int, string, error) {
	want := strings.ToLower(strings.TrimSpace(text))
	var names []string
	for _, o := range m.Objects() {
		if o.Parent != parent {
			continue
		}
		switch o.Type {
		case "Item", "Value", "TreeItem":
		default:
			continue
		}
		label := o.Attr("text")
		if fn, ok := r.Resolve(o.ID); ok && fn.Text != "" {
			label = fn.Text
		}
		names = append(names, label)
		if strings.Contains(strings.ToLower(label), want) {
			return o.ID, label, nil
		}
	}
	if len(names) == 0 {
		return 0, "", fmt.Errorf("这个控件下面没有候选项（可能要先查询才有值）")
	}
	return 0, "", fmt.Errorf("候选项里没有 %q；现有：%s", text, strings.Join(names, " | "))
}

func init() {
	doRoleCmd.Flags().StringVar(&doRoleSet, "set", "", "写入值（Edit/ButtonEdit/Value/DateEdit）")
	doRoleCmd.Flags().BoolVar(&doRoleOpen, "open", false, "触发开窗（ButtonEdit 的 controlp）")
	doRoleCmd.Flags().BoolVar(&doRoleInvoke, "invoke", false, "调用动作（Action/ToolBarItem/MenuAction）")
	doRoleCmd.Flags().StringVar(&doRoleChoose, "choose", "", "选候选项（ComboBox/Item/TreeItem）")
	doRoleCmd.Flags().StringVar(&doRoleIn, "in", "", "限定窗口（传窗口名字片段，路由用）")
}
