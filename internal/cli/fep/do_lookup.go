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
	doLookupQuery string
	doLookupRow   int
)

// doLookupCmd：一条命令完成"从开窗里取一个值"。
//
// 为什么需要它（今天手搓失败的地方）：开窗的每一步都会让运行期**换掉 Dialog 与动作 id**，
// 也会换当前窗口；用手敲命令时每步之间状态已经变了，于是取到旧 id、发出去被丢。
// 这个命令把**实测验证过的序列**固化下来，并且**每一步都重新读一次状态**：
//
//  1. 点开窗按钮（<开窗按钮角色|#id>）：一般是字段的 ButtonEdit
//  2. 探测并建立"实例 Dialog"：点实例自己的 reconstruct/accept，直到运行期给这个窗口建
//     `an <窗口> Dialog <id> {{Action <n> accept} …}`（实测运行期只在真实交互后才建）
//  3. 写查询条件：{网格}{pageSize}{bufferSize} + {格子}{value} + {列}{cursor}
//  4. 点刚建立那个 Dialog 的 accept → 查询执行
//  5. 设 {网格}{currentRow N}，再点（重新读到的）accept → 选中行回传
func doLookupCmd2() *cobra.Command {
	c := &cobra.Command{
		Use:   "lookup <开窗按钮角色|#id> [--query 条件] [--row N]",
		Short: "从开窗里取一个值（开窗→建实例→写条件→查询→选行→回传；每步重读状态）",
		Long: `tt fep do lookup '#651' --query 1 --row 1        # 明细的料件编号开窗
tt fep do lookup 'b_pmdl004' --query 0 --row 1   # 供应商编号开窗

每一步之间运行期都会换当前窗口/Dialog/动作 id，所以本命令每一步都重新读一次状态 ——
这是手敲命令时最容易踩的坑（实测：取到上一个窗口的 accept，发出去只回 18 字节空 ACK）。`,
		Args: exactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			send := func(kind, target string, raw string) (*fep.DoResult, error) {
				dc := fep.DoConfig{Kind: kind, Target: target, RawEvent: raw,
					Conn: doConn, DryRun: doDryRun, Wait: doWait}
				return fep.Do(client, dc)
			}
			logf := func(format string, a ...any) {
				fmt.Fprintf(cmd.ErrOrStderr(), "· "+format+"\n", a...)
			}

			// 1) 点开窗按钮
			before := liveWindows(client)
			if _, err := send("click", args[0], ""); err != nil {
				var ref *fep.Refusal
				if errors.As(err, &ref) {
					return fail(ref.Exit, "%s", ref.Message)
				}
				return fail(exitFailed, "%v", err)
			}
			newWin := waitNewWindow(client, before)
			if newWin == 0 {
				return fail(exitTarget, "点了 %s 之后没有新窗口（可能不是开窗按钮，或它此刻不可用）", args[0])
			}
			logf("开窗：窗口 #%d", newWin)

			// 2) 建立实例 Dialog：试探实例自己的 reconstruct / accept
			var acts []int
			for hop := 0; hop < 6; hop++ {
				m, _, _, lerr := loadModel(uiFrom)
				if lerr != nil {
					return fail(exitFailed, "%v", lerr)
				}
				if id, ok := instanceAccept(m, newWin); ok {
					acts = append(acts, id)
					break
				}
				cand := probeCandidates(m, newWin)
				if len(cand) == 0 {
					return fail(exitTarget, "窗口 #%d 里找不到可试探的动作（reconstruct/accept/refresh）", newWin)
				}
				id := cand[0]
				logf("试探 #%d（让运行期给这个窗口建实例 Dialog）", id)
				_, _ = send("click", fmt.Sprintf("#%d", id), "")
				acts = append(acts, id)
			}
			m, _, _, err := loadModel(uiFrom)
			if err != nil {
				return fail(exitFailed, "%v", err)
			}
			acc, ok := instanceAccept(m, newWin)
			if !ok {
				return fail(exitTarget, "运行期始终没给窗口 #%d 建实例 Dialog（accept 取不到）", newWin)
			}
			logf("实例 Dialog 的 accept = #%d", acc)

			// 3) 写查询条件：第一列（跳过多选）为条件列
			col, cell, grid, oerr := conditionTargets(m, newWin)
			if oerr != nil {
				return fail(exitTarget, "%v", oerr)
			}
			ev := ""
			if grid != 0 {
				ev += fmt.Sprintf("ConfigureEvent 0{{idRef \"%d\"}{ pageSize \"9\"}{ bufferSize \"10\"}}}{", grid)
			}
			q := doLookupQuery
			ev += fep.ConfigureValueText(cell, q, nil) + "}{" + fep.ConfigureCursorText(col, len([]rune(q)))
			logf("写条件 %q（列 #%d 格子 #%d 网格 #%d）", q, col, cell, grid)
			if _, err := send("raw", fmt.Sprintf("#%d", col), ev); err != nil {
				return fail(exitFailed, "%v", err)
			}

			// 4) 查询
			if _, err := send("click", fmt.Sprintf("#%d", acc), ""); err != nil {
				return fail(exitFailed, "%v", err)
			}
			logf("已触发查询")

			// 5) 选行 + 回传（accept 要重读：运行期换过 Dialog）
			row := doLookupRow
			if row <= 0 {
				row = 1
			}
			m2, _, _, _ := loadModel(uiFrom)
			acc2, _ := instanceAccept(m2, newWin)
			if acc2 == 0 {
				acc2 = acc
			}
			ev2 := ""
			if grid != 0 {
				ev2 = fmt.Sprintf("ConfigureEvent 0{{idRef \"%d\"}{ currentRow \"%d\"}}}{", grid, row)
			}
			if ev2 != "" {
				ev2 = strings.TrimSuffix(ev2, "}{")
			}
			if _, err := send("raw", fmt.Sprintf("#%d", grid), ev2); err != nil {
				return fail(exitFailed, "%v", err)
			}
			res, derr := send("click", fmt.Sprintf("#%d", acc2), "")
			if derr != nil {
				return fail(exitFailed, "%v", derr)
			}
			eff := map[string]any{"protocol": "tt-aui/1", "op": "lookup",
				"window": newWin, "query": q, "row": row, "accept": acc2, "session": s.Prog}
			if res.Changes == nil || (len(res.Changes.Values) == 0 && len(res.Changes.NewWindows) == 0) {
				eff["effect"] = "noEffect"
			} else {
				eff["effect"] = "applied"
				eff["valueChanged"] = res.Changes.Values
			}
			_ = fep.AppendAudit(filepath.Join(recDirForAudit(s), "audit.log"), fep.AuditEntry{
				Verb: "lookup", DryRun: doDryRun, Session: s.Prog, Result: eff["effect"].(string),
			})
			return printJSONOrText(eff, func() {
				fmt.Printf("lookup 窗口 #%d ← 条件 %q，选第 %d 行\n", newWin, q, row)
				if eff["effect"] == "applied" {
					for _, vc := range res.Changes.Values {
						fmt.Printf("  效果：值变化 #%d %q → %q\n", vc.ID, vc.From, vc.To)
					}
				} else {
					fmt.Println("  效果：无（运行期校验后没改动任何节点）")
				}
			})
		},
	}
	c.Flags().StringVar(&doLookupQuery, "query", "", "查询条件（写在开窗的第一列）")
	c.Flags().IntVar(&doLookupRow, "row", 1, "选中第几行回传（1 起）")
	return c
}

// liveWindows 返回当前 AUI 里的窗口 id 集合。
func liveWindows(c *fep.Client) map[int]bool {
	m, _, _, err := loadModel(uiFrom)
	out := map[int]bool{}
	if err != nil {
		return out
	}
	for _, w := range m.Windows() {
		out[w.ID] = true
	}
	return out
}

// waitNewWindow 等到出现一个不在 before 里的窗口，返回它的 id。
func waitNewWindow(c *fep.Client, before map[int]bool) int {
	for i := 0; i < 20; i++ {
		m, _, _, err := loadModel(uiFrom)
		if err == nil {
			for _, w := range m.Windows() {
				if !before[w.ID] && w.ID != 0 && w.ID != 92 {
					return w.ID
				}
			}
		}
		sleepMS(300)
	}
	return 0
}

// instanceAccept 返回"属于这个窗口实例的 Dialog"里的 accept。
func instanceAccept(m *fep.Model, win int) (int, bool) {
	for _, d := range m.ActiveDialogs() {
		if d.Parent != win && d.Scope != win {
			continue
		}
		for _, o := range m.Objects() {
			if o.Parent == d.ID && o.Type == "Action" && o.Attr("name") == "accept" {
				return o.ID, true
			}
		}
	}
	return 0, false
}

// probeCandidates 给出这个窗口里值得试探的动作（实例自己的 reconstruct/accept/refresh）。
func probeCandidates(m *fep.Model, win int) []int {
	var out []int
	want := map[string]bool{"reconstruct": true, "accept": true, "refresh": true}
	for _, o := range m.Objects() {
		if o.Type != "ToolBarItem" && o.Type != "Button" && o.Type != "Action" {
			continue
		}
		if !want[o.Attr("name")] {
			continue
		}
		if w, ok := m.Window(o.ID); !ok || w != win {
			continue
		}
		out = append(out, o.ID)
	}
	return out
}

// conditionTargets 找开窗里"写条件用的列/格子/网格"：第一列真实列（跳过 多选/check）。
func conditionTargets(m *fep.Model, win int) (col, cell, grid int, err error) {
	type cand struct {
		id, order int
		role      string
	}
	var cols []cand
	for _, o := range m.Objects() {
		if o.Type != "TableColumn" {
			continue
		}
		if w, ok := m.Window(o.ID); !ok || w != win {
			continue
		}
		cols = append(cols, cand{id: o.ID, order: o.ID, role: o.Attr("name")})
	}
	for _, c := range cols {
		if strings.Contains(c.role, "check") {
			continue
		}
		cellID, cerr := fep.FirstCellOfIn(m, c.id, win)
		if cerr != nil {
			continue
		}
		return c.id, cellID, fep.GridOfColumn(m, c.id), nil
	}
	return 0, 0, 0, fmt.Errorf("窗口 #%d 里找不到可写条件的列", win)
}

func init() { doCmd.AddCommand(doLookupCmd2()) }

// sleepMS 小睡一会儿（等运行期把新窗口/AUI 推过来）。
func sleepMS(ms int) { time.Sleep(time.Duration(ms) * time.Millisecond) }
