package fep

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var (
	uiRolesWindow string
	uiRolesKind   string
	uiRolesSource string
)

// uiRolesCmd 输出"稳定目录"：可操作控件 → 角色 / 类型 / 状态 / 支持的操作。
//
// 这是给自动化用的接口，因此：
//   - 寻址用「窗口/角色」（角色优先取协议 name，其次中文名），脚本里**不出现 id**
//   - 每个控件的 ops 由它的类型决定（见 internal/fep/roles.go 的类型表）
//   - 同名角色单列一段：今天实测"交易对象编号"就是被裸 id 建模型搞错的那个
var uiRolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "稳定目录：可操作控件 → 角色/类型/状态/可用操作（给自动化用，不含 id 寻址）",
	Long: `把 AUI 折成一份**角色寻址**的目录：每个可操作控件给出窗口、角色、类型、
当前状态（value/active/hidden）与它这个类型**支持的操作**。

寻址一律用「窗口/角色」或角色名，脚本里不要写 id —— 实测 FEP 的数字 id
不是全局身份（同一会话里"交易对象编号"会命中别的窗口的候选项），
按裸 id 建模型必然张冠李戴。

每个控件支持什么操作由**类型**决定（Edit=set、ButtonEdit=set+open、
ComboBox/Item=choose、CheckBox=toggle、Action/ToolBarItem/MenuAction=invoke…），
类型不支持的操作会被 do 直接拒绝，而不是发出去碰运气。`,
	Args: noArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		m, r, _, err := loadModel(uiFrom)
		if err != nil {
			return err
		}
		cat := fep.Catalog(m, r)
		// M5：带上源码骨架 —— 角色改用源码字段名（稳定），并挂上 required/opener。
		srcNote := ""
		if uiRolesSource != "" {
			prog := ""
			if reg, _, rerr := registry(); rerr == nil {
				if s, lerr := reg.Load(); lerr == nil {
					prog = s.Prog
				}
			}
			if prog == "" {
				srcNote = "没能从会话登记里拿到作业编号（--source 需要它）"
			} else if src, serr := fep.LoadSource(uiRolesSource, prog); serr != nil {
				srcNote = serr.Error()
			} else {
				n := fep.ApplySource(cat, src)
				srcNote = fmt.Sprintf("源码：%s（表单 %d 个 / 字段 %d 条 / 动作 %d 个）—— 角色已改用源码字段名，匹配上 %d 个控件",
					prog, len(src.Forms), len(src.Fields), len(src.Actions), n)
			}
		}
		if uiRolesKind != "" {
			cat = filterKind(cat, uiRolesKind)
		}
		if uiRolesWindow != "" {
			cat = filterWindow(cat, uiRolesWindow)
		}
		// 同名冲突：同一角色出现在多个窗口（今天踩过的坑，单独列出来）
		byRole := map[string]map[string]bool{}
		for _, c := range cat {
			if byRole[c.Role] == nil {
				byRole[c.Role] = map[string]bool{}
			}
			byRole[c.Role][c.Window] = true
		}
		var dup []string
		for role, wins := range byRole {
			if len(wins) > 1 {
				dup = append(dup, role)
			}
		}
		sort.Strings(dup)

		view := map[string]any{
			"protocol":       "tt-aui/1",
			"note":           "寻址用 窗口/角色；id 仅作落点提示，不要写进脚本",
			"controls":       cat,
			"ambiguousRoles": dup,
			"source":         srcNote,
		}
		return printJSONOrText(view, func() {
			fmt.Printf("协议 tt-aui/1 —— 可操作控件 %d 个（按窗口/角色寻址，id 仅供落点参考）\n", len(cat))
			if srcNote != "" {
				fmt.Printf("%s\n", srcNote)
			}
			cur := ""
			for _, c := range cat {
				if c.Window != cur {
					cur = c.Window
					fmt.Printf("\n窗口 %s\n", cur)
				}
				state := []string{}
				if c.Source != nil && c.Source.Required {
					state = append(state, "required")
				}
				if c.Source != nil && c.Source.Opener != "" {
					state = append(state, "opener="+c.Source.Opener)
				}
				if c.Value != "" {
					state = append(state, fmt.Sprintf("value=%q", c.Value))
				}
				if c.Active != "" {
					state = append(state, "active="+c.Active)
				}
				if c.ActionActive != "" {
					state = append(state, "actionActive="+c.ActionActive)
				}
				if c.ActionIDRef != "" && c.ActionIDRef != "0" {
					state = append(state, "actionIdRef="+c.ActionIDRef)
				}
				fmt.Printf("  %-22s %-14s ops=%-14s #%-6d %s\n",
					c.Role, c.Kind, strings.Join(c.Ops, ","), c.ID, strings.Join(state, " "))
			}
			if len(dup) > 0 {
				fmt.Printf("\n⚠ 同名冲突（同一角色出现在多个窗口，脚本里必须写「窗口/角色」）：\n")
				for _, role := range dup {
					wins := []string{}
					for w := range byRole[role] {
						wins = append(wins, w)
					}
					sort.Strings(wins)
					fmt.Printf("  %s → %s\n", role, strings.Join(wins, " | "))
				}
			}
		})
	},
}

func filterKind(cat []fep.Control, kind string) []fep.Control {
	var out []fep.Control
	for _, c := range cat {
		if strings.EqualFold(c.Kind, kind) {
			out = append(out, c)
		}
	}
	return out
}

func filterWindow(cat []fep.Control, want string) []fep.Control {
	want = strings.ToLower(want)
	var out []fep.Control
	for _, c := range cat {
		if strings.Contains(strings.ToLower(c.Window), want) {
			out = append(out, c)
		}
	}
	return out
}

func init() {
	uiRolesCmd.Flags().StringVar(&uiRolesWindow, "window", "", "只看某个窗口（名字片段）")
	uiRolesCmd.Flags().StringVar(&uiRolesKind, "kind", "", "只看某种控件类型（Edit/ButtonEdit/Action…）")
	uiRolesCmd.Flags().StringVar(&uiRolesSource, "source", "",
		"源码镜像根目录（如 D:/T100/4gl文档 的同级……用 tt dict mirror path 拿到的那个）：\n带上它，角色改用源码字段名并显示 required/opener")
}
