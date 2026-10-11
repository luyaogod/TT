package fep

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var (
	doCellIn     string
	doCellAccept bool
	doCellRow    int
)

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
		// 允许直接给活实例里的列 id（形如 #3555）：跨实例同名角色太多，
		// 实测按角色会写到别的实例的格子上（#96）；用 id 最确定。
		if strings.HasPrefix(args[0], "#") {
			id, cerr := strconv.Atoi(strings.TrimPrefix(args[0], "#"))
			if cerr != nil {
				return fail(exitUsage, "列 id 写法应为 #数字，得到 %q", args[0])
			}
			if o, ok := m.Get(id); ok {
				// 同名冲突的教训：m.Get 可能拿到全局那份（实测 #3555 是 Item），
				// 而目录里那份才是当前实例的 TableColumn。所以优先在目录里按 id 找。
				kind, role := o.Type, args[0]
				for _, c := range cat {
					if c.ID == id {
						kind, role = c.Kind, c.Role
						break
					}
				}
				cat = []fep.Control{{ID: id, Kind: kind, Role: role}}
			} else {
				return fail(exitTarget, "模型里没有 #%d", id)
			}
		}
		if !strings.HasPrefix(args[0], "#") {
			// 格子操作只在 TableColumn 里选：带源码时角色是按字段名改写的，同名字段
			// 会同时落到"列"和某个工具栏项上（实测 formonly.pmaa001_1 就是），必须按类型收窄。
			cat = fep.FilterKind(cat, "TableColumn")
		}
		var col fep.Control
		var cands []fep.Control
		var rerr error
		if strings.HasPrefix(args[0], "#") {
			col = cat[0] // 已直接给 id，不再按名字解析
		} else {
			col, cands, rerr = fep.ResolveRole(cat, args[0])
		}
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
		// 先按前端的做法告诉运行期“这个网格现在显示着”（且必须是**当前实例**的网格）。
		// 实测前端在写条件之前就发这条；少了它，后面的 reconstruct 会被当成空动作。
		grid := fep.GridOfColumn(m, col.ID)
		prefix := ""
		if grid != 0 && grid != col.ID {
			prefix = fmt.Sprintf("ConfigureEvent 0{{idRef \"%d\"}{ pageSize \"9\"}{ bufferSize \"10\"}}}{", grid)
		}
		// --row：先设网格当前行（源码里 `IF li_ac > 0 THEN` 才回传选中行）。
		if doCellRow > 0 && grid != 0 {
			prefix += fmt.Sprintf("ConfigureEvent 0{{idRef \"%d\"}{ currentRow \"%d\"}}}{", grid, doCellRow)
		}
		// 多个事件在同一条消息里用 }{ 分隔（与前端形态一致）。
		ev := prefix + fep.ConfigureValueText(cell, args[1], nil) + "}{" +
			fep.ConfigureCursorText(col.ID, len([]rune(args[1])))
		// --accept：把当前可用的 ✓ 一起放进**同一条消息**。
		// 实测：条件、光标、✓ 三个事件同一条消息发，开窗才“带着条件去查”；
		// 分开发只会选中已有行（网格空时什么都不发生）。
		if doCellAccept {
			// 用**当前交互指令持有的 accept**（官方 1514：动作按 Dialog 组织），不按图标猜。
			acc, ok := fep.ActionNameInDialog(m, "accept")
			var aerr error
			if !ok {
				aerr = fmt.Errorf("当前交互指令里没有名为 accept 的动作（不要猜着发）")
			}
			if aerr != nil {
				return fail(exitTarget, "%v", aerr)
			}
			ev += "}{" + fep.ActionEventText(acc)
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
	doCellCmd.Flags().BoolVar(&doCellAccept, "accept", false,
		"把当前可用的 ✓ 一起放进同一条消息（写条件 + 回车查的那种形态）")
	doCellCmd.Flags().IntVar(&doCellRow, "row", 0, "先把网格当前行设成第 N 行（1 起；选行回传时用）")
	doCmd.AddCommand(doCellCmd)
}

// currentAccept 找"此刻可用的 ✓"：带 image=tb/accept.png（或 name=accept）且 actionActive=1。
// 实测这个 id 每个开窗都会变（上一轮 #3550、这一轮 #121），所以必须现查。
func currentAccept(m *fep.Model) (int, error) {
	for _, o := range m.Objects() {
		switch o.Type {
		case "ToolBarItem", "Action", "Button", "MenuAction":
		default:
			continue
		}
		if !strings.Contains(o.Attr("image"), "accept") && !strings.Contains(o.Attr("name"), "accept") {
			continue
		}
		if o.Attr("actionActive") == "1" {
			return o.ID, nil
		}
	}
	return 0, fmt.Errorf("此刻没有可用的 ✓（开窗可能没开或已关闭）")
}
