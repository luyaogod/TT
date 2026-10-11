package fep

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

// uiStateCmd：一屏看清"作业此刻处在哪个交互指令、它能收到哪些动作、焦点在哪"。
//
// 为什么要它（今天踩了一整天的坑）：官方 1514 说得很清楚 ——
// AUI 用一组 Action 节点表示**当前交互指令**能收到的动作，这组动作由 Dialog 节点持有，
// 交互元素按 name 绑定到 Action。所以"能不能发某个动作"不是看图标（我先前的做法，错的），
// 而是看它在不在**当前 Dialog** 的动作集里。以前没有这个视图，我只能靠 image=accept 猜。
var uiStateCmd = &cobra.Command{
	Use:   "state",
	Short: "一屏看清：当前窗口 / 当前交互指令(Dialog) 及它持有的动作 / 焦点",
	Long: `把运行期这一路推来的状态汇成一屏：

  当前窗口（currentWindow 推送）
  当前交互指令：active=1 的 Dialog 节点，以及它直接持有的 AUI 动作（name/text/image）
  焦点：运行期推的 un 0 {{focus "<id>"}} 指向谁
  作业状态：runtimeStatus（interactive / …）

按官方 1514，能发给运行期的动作＝当前 Dialog 持有的那组；
所以这一步只看"当前 Dialog 有什么"，不再按图标猜。`,
	Args: noArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		m, r, _, err := loadModel(uiFrom)
		if err != nil {
			return err
		}
		type actRow struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Text   string `json:"text,omitempty"`
			Image  string `json:"image,omitempty"`
			Active string `json:"active,omitempty"`
		}
		type dlgRow struct {
			ID      int      `json:"id"`
			Window  string   `json:"window,omitempty"`
			Actions []actRow `json:"actions"`
		}
		focus := ""
		for _, o := range m.Objects() {
			if o.ID == 0 {
				if v := o.Attr("focus"); v != "" {
					focus = v
				}
			}
		}
		var dlgs []dlgRow
		for _, d := range m.ActiveDialogs() {
			row := dlgRow{ID: d.ID, Window: m.WindowName(d.ID)}
			for _, o := range m.Objects() {
				if o.Parent != d.ID {
					continue
				}
				switch o.Type {
				case "Action", "MenuAction", "IdleAction":
				default:
					continue
				}
				row.Actions = append(row.Actions, actRow{
					ID: o.ID, Name: o.Attr("name"), Text: o.Attr("text"),
					Image: o.Attr("image"), Active: o.Attr("active"),
				})
			}
			sort.Slice(row.Actions, func(i, j int) bool { return row.Actions[i].ID < row.Actions[j].ID })
			dlgs = append(dlgs, row)
		}
		focusName := ""
		if id, cerr := strconv.Atoi(strings.TrimSpace(focus)); cerr == nil && id != 0 {
			if o, ok := m.Get(id); ok {
				focusName = o.Type
				if fn, ok := r.Resolve(id); ok && fn.Text != "" {
					focusName = fn.Text + " (" + o.Type + ")"
				}
			}
		}
		view := map[string]any{
			"protocol":      "tt-aui/1",
			"currentWindow": m.CurrentWindow(),
			"focus":         focus,
			"focusName":     focusName,
			"activeDialogs": dlgs,
		}
		return printJSONOrText(view, func() {
			fmt.Printf("当前窗口：#%d %s\n", m.CurrentWindow(), m.WindowName(m.CurrentWindow()))
			if focus != "" {
				fmt.Printf("焦点：#%s  %s\n", focus, focusName)
			}
			if len(dlgs) == 0 {
				fmt.Println("当前交互指令：没有 active=1 的 Dialog（运行期可能没推，或作业没在等输入）")
				return
			}
			for _, d := range dlgs {
				fmt.Printf("当前交互指令：Dialog #%d %s —— 持有 %d 个动作\n", d.ID, d.Window, len(d.Actions))
				for _, a := range d.Actions {
					img := a.Image
					if img != "" {
						img = " image=" + img
					}
					fmt.Printf("    %-16s %-12s #%-6d%s\n", a.Name, a.Text, a.ID, img)
				}
			}
			if id, ok := fep.ActionNameInDialog(m, "accept"); ok {
				fmt.Printf("→ 当前指令里的 accept = #%d（要“确定”就发它）\n", id)
			} else {
				fmt.Println("→ 当前指令里没有名为 accept 的动作（那就别发 —— 宁缺不猜）")
			}
		})
	},
}
