package fep

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var (
	uiFrom      string
	uiText      string
	uiType      string
	uiLimit     int
	uiOnlyNamed bool
	recAll      bool
)

// uiWindow / uiRow / uiStep 是命令的输出形状（JSON 信封里就是它们的字段名）。
type uiWindow struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
}

type uiRow struct {
	ID         int    `json:"id"`
	Type       string `json:"type"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	Confidence string `json:"confidence"`
	Window     string `json:"window"`
	Comment    string `json:"comment,omitempty"`
	Col        string `json:"col,omitempty"`
}

type uiStep struct {
	N      int    `json:"n"`
	Kind   string `json:"kind"`
	Target string `json:"target"`
	Name   string `json:"name,omitempty"`
	Conf   string `json:"confidence,omitempty"`
	Value  string `json:"value,omitempty"`
	Raw    string `json:"raw"`
}

func (s uiStep) targetLabel() string {
	if s.Name != "" {
		return s.Name + " (" + s.Target + ")"
	}
	return s.Target + " (名字未解析)"
}

func init() {
	for _, c := range []*cobra.Command{uiSnapshotCmd, uiFindCmd, recStepsCmd} {
		c.Flags().StringVar(&uiFrom, "from", "", "录制件：单个 .log 或存放它的目录（默认 fep 缓存目录）")
		c.Flags().BoolVar(&recAll, "all", false, "目录里有多份录制件时全部合并读（默认只读最新一份）")
	}
	uiFindCmd.Flags().StringVar(&uiText, "text", "", "按名字/文本筛（大小写不敏感）")
	uiFindCmd.Flags().StringVar(&uiType, "type", "", "只保留该类型的对象（如 Window / Table）")
	uiFindCmd.Flags().IntVar(&uiLimit, "limit", 0, "最多输出多少条（0 = 不限）")
	uiFindCmd.Flags().BoolVar(&uiOnlyNamed, "named", false, "只要解析出名字的")
}

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "界面：从录制件读 AUI（离线可用）",
	Long: `读一份录制件里的 AUI，回答"这个作业界面上有什么、叫什么名字"。

字段名一律来自协议（注释键 → 表.字段名后缀 → 自身文本 → 受限继承），
每条都带来源与置信度；低置信/未解析的一律照实标出来，不编名字。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return fail(exitUsage, "未知子命令 %q", args[0])
	},
}

var uiSnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "汇总界面结构：窗口/对象数/父链冲突/字段名覆盖率",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, r, frames, err := loadModel(uiFrom)
		if err != nil {
			return err
		}
		resolved, total := r.Coverage()
		var wins []uiWindow
		for _, w := range m.Windows() {
			wins = append(wins, uiWindow{ID: w.ID, Name: w.Attr("name"), Title: w.Attr("title")})
		}
		counts := map[string]int{}
		for _, o := range m.Objects() {
			counts[o.Type]++
		}
		view := map[string]any{
			"frames":     frames,
			"objects":    m.Len(),
			"conflicts":  m.Conflicts(),
			"windows":    wins,
			"types":      counts,
			"namesTotal": total,
			"namesOK":    resolved,
		}
		return printJSONOrText(view, func() {
			fmt.Printf("帧 %d / 对象 %d / 父链冲突 %d\n", frames, m.Len(), m.Conflicts())
			fmt.Printf("字段名覆盖：%d/%d", resolved, total)
			if total > 0 {
				fmt.Printf("（%.1f%%）", float64(resolved)*100/float64(total))
			}
			fmt.Println()
			fmt.Println("窗口：")
			for _, w := range m.Windows() {
				t := w.Attr("title")
				if t == "" {
					t = "-"
				}
				fmt.Printf("  #%d  %s  %s\n", w.ID, w.Attr("name"), t)
			}
			keys := make([]string, 0, len(counts))
			for k := range counts {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool {
				if counts[keys[i]] != counts[keys[j]] {
					return counts[keys[i]] > counts[keys[j]]
				}
				return keys[i] < keys[j]
			})
			fmt.Println("对象类型（按数量）：")
			for _, k := range keys {
				fmt.Printf("  %-24s %d\n", k, counts[k])
			}
		})
	},
}

var uiFindCmd = &cobra.Command{
	Use:   "find",
	Short: "找对象：按名字/文本筛，输出 id、类型、名字来源与置信度",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, r, _, err := loadModel(uiFrom)
		if err != nil {
			return err
		}
		objs := m.Objects()
		sort.Slice(objs, func(i, j int) bool { return objs[i].ID < objs[j].ID })
		var rows []uiRow
		for _, o := range objs {
			if uiType != "" && o.Type != uiType {
				continue
			}
			fn, ok := r.Resolve(o.ID)
			if uiOnlyNamed && (!ok || fn.Text == "") {
				continue
			}
			if uiText != "" {
				hay := strings.Join([]string{o.Attr("text"), o.Attr("label"), o.Attr("comment"),
					o.Attr("name"), o.Attr("col"), fn.Text}, "\x00")
				if !strings.Contains(strings.ToLower(hay), strings.ToLower(uiText)) {
					continue
				}
			}
			rows = append(rows, uiRow{ID: o.ID, Type: o.Type, Name: fn.Text, Source: fn.Source,
				Confidence: string(fn.Confidence), Window: m.WindowName(o.ID),
				Comment: o.Attr("comment"), Col: o.Attr("col")})
			if uiLimit > 0 && len(rows) >= uiLimit {
				break
			}
		}
		return printJSONOrText(map[string]any{"count": len(rows), "rows": rows}, func() {
			for _, x := range rows {
				name := x.Name
				if name == "" {
					name = "(未解析)"
				}
				fmt.Printf("#%-6d %-18s %-22s %-12s %s", x.ID, x.Type, name, x.Confidence, x.Source)
				if x.Comment != "" {
					fmt.Printf("  comment=%s", x.Comment)
				}
				if x.Col != "" {
					fmt.Printf("  col=%s", x.Col)
				}
				fmt.Println()
			}
			fmt.Printf("（%d 条）\n", len(rows))
		})
	},
}

var recCmd = &cobra.Command{
	Use:   "rec",
	Short: "录制件：把一轮原始事件抽成语义动作",
	Long: `把录制件里前端发来的原始事件，抽成"人做过的动作"。

只是路过的栏位（KeyEvent Tab）、光标位置、行号这一类导航与状态事件全部丢弃：
一步 = 点一个东西，或往一个栏位里输入一个值。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return fail(exitUsage, "未知子命令 %q", args[0])
	},
}

var recStepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "语义动作清单：点击/输入，附动作对象解析出的名字",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		events, err := loadEvents(uiFrom)
		if err != nil {
			return err
		}
		actions := fep.SemanticActions(events)
		// 名字从 AUI 里取；AUI 缺失不算失败（事件本身仍然是有效证据）
		m, r, _, merr := loadModel(uiFrom)
		steps := make([]uiStep, 0, len(actions))
		for i, a := range actions {
			s := uiStep{N: i + 1, Kind: string(a.Kind), Value: a.Value, Raw: a.Raw}
			if merr == nil {
				if o, ok := m.Get(a.ID); ok {
					s.Target = fmt.Sprintf("%s#%d", o.Type, o.ID)
				}
				if fn, ok := r.Resolve(a.ID); ok && fn.Text != "" && fn.Confidence == fep.ConfHigh {
					s.Name, s.Conf = fn.Text, string(fn.Confidence)
				}
			}
			if s.Target == "" {
				s.Target = fmt.Sprintf("#%d", a.ID)
			}
			steps = append(steps, s)
		}
		return printJSONOrText(map[string]any{"rawEvents": len(events), "steps": len(steps), "actions": steps}, func() {
			fmt.Printf("原始事件 %d → 语义动作 %d\n", len(events), len(steps))
			for _, s := range steps {
				if s.Kind == string(fep.ActionInput) {
					fmt.Printf("%2d. 输入  %-24s ← %q\n", s.N, s.targetLabel(), s.Value)
					continue
				}
				fmt.Printf("%2d. 点击  %-24s\n", s.N, s.targetLabel())
			}
		})
	},
}
