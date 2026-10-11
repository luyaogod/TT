package fep

import (
	"fmt"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

// tt fep ui focus —— **此刻焦点在哪**。
//
// 为什么必须有这条命令（用户点出来的要害）：
// 运行期只认"光标现在确实在这个栏位"时的动作。焦点不在那儿，动作名再对、
// actionActive 再是 1，点下去也是**空动作**（不报错、不生效，让人误判成"形状不对"）。
// 我们内部早就为**回放**这条线写下了这个教训（internal/fep/replay.go:21），
// 但驱动那条线一直没查它 —— 于是反复出现"点对了动作却没反应"。
//
// 焦点由运行期自己推（`un 0 {{focus "<id>"}}`），所以这是**读取**，不是推断。
var uiFocusCmd = &cobra.Command{
	Use:   "focus",
	Short: "现在焦点在哪（谁持有焦点、在哪个窗口/网格/第几行第几列、下一步能不能动手）",
	Long: `看焦点 —— 驱动每一步之前都该看的东西。

运行期会推 un 0 {{focus "<id>"}}（自己声明焦点在哪），本命令把它连同
"它挂在哪个网格、第几行第几列"一起打出来，并给一句判断：

  焦点在栏位/格子上  → 针对该栏位的动作（开窗、写值）才可能生效
  焦点在动作/菜单上  → 说明此刻在等一个"确认"（错误窗、提示窗），先把它按掉
  焦点在别处         → 想对某个栏位动手，得先把焦点移过去（否则是空动作）

退出码 0；没有会话 3；读不到运行期状态 1。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, r, _, err := loadModel(uiFrom)
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		rep := m.FocusNow()
		// 焦点对象的角色名/源码归属：能对到目录里就一起给出来
		role, kind, comment := "", "", ""
		if rep.FocusID != 0 {
			for _, c := range fep.Catalog(m, r) {
				if c.ID == rep.FocusID {
					role, kind, comment = c.Role, c.Kind, c.Comment
					break
				}
			}
		}
		verdict := ""
		switch {
		case rep.FocusID == 0 && rep.CursorID == 0:
			verdict = "读不到焦点（运行期没推 focus，也没有 cursor 属性）—— 判据不足，不猜"
		case kind == "MenuAction" || kind == "Action" || kind == "ToolBarItem":
			verdict = "焦点在一个**动作**上：此刻在等一次确认（错误窗/提示窗/菜单），先把它按掉再谈别的"
		default:
			verdict = "焦点在**栏位/格子**上：针对它的动作（写值、开窗）才可能生效；对别的栏位动手前要先移焦点"
		}
		out := map[string]any{
			"protocol": "tt-aui/1", "op": "focus",
			"focusId": rep.FocusID, "focusRole": role, "focusKind": kind, "focusComment": comment,
			"cursorId": rep.CursorID, "cursorName": rep.CursorNam,
			"window": m.WindowName(rep.WindowID), "windowId": rep.WindowID,
			"grid": rep.Grid, "row": rep.Row, "column": rep.Col,
			"verdict": verdict,
		}
		return printJSONOrText(out, func() {
			if rep.FocusID != 0 {
				line := fmt.Sprintf("焦点：#%d", rep.FocusID)
				if role != "" {
					line += "  " + role
				} else if rep.FocusName != "" {
					line += "  " + rep.FocusName
				}
				if kind != "" {
					line += "  (" + kind + ")"
				}
				fmt.Println(line)
			} else {
				fmt.Println("焦点：运行期没推（un 0 {{focus …}}）")
			}
			if rep.CursorID != 0 {
				fmt.Printf("光标栏位：#%d  %s\n", rep.CursorID, rep.CursorNam)
			}
			fmt.Printf("窗口：#%d %s\n", rep.WindowID, m.WindowName(rep.WindowID))
			if rep.Grid != 0 {
				fmt.Printf("网格：#%d  第 %d 行 第 %d 列\n", rep.Grid, rep.Row+1, rep.Col+1)
			}
			fmt.Println(verdict)
		})
	},
}

func init() {
	uiFocusCmd.Flags().StringVar(&uiFrom, "from", "", "读哪个录制件（缺省取当前会话）")
	uiCmd.AddCommand(uiFocusCmd)
}
