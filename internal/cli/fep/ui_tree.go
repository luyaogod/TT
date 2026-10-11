package fep

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var (
	uiTreeWindow string
	uiTreeDepth  int
)

var uiTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "按窗口看它自己的那棵树（开窗是独立一棵树，它的按钮只在它自己树里）",
	Long: `按窗口（名字片段）打印它自己的子树：窗口 → 容器 → 工具栏/对话框 → 按钮。

为什么需要它：开窗、对话框都是**独立的窗口树**，它们的按钮/动作只挂在
那一棵树上。在整个 AUI 里按名字全局搜，会搜到别的窗口的同名对象
（主窗的「确定」、别的弹窗的「确定」），只能靠猜 —— 实测就是这么踩的。
所以要看某个开窗有什么按钮，就得先找到它的窗口，再看它的树。`,
	Args: noArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		m, r, _, err := loadModel(uiFrom)
		if err != nil {
			return err
		}
		kids := map[int][]*fep.Object{}
		for _, o := range m.Objects() {
			kids[o.Parent] = append(kids[o.Parent], o)
		}
		want := strings.ToLower(uiTreeWindow)
		// 窗口的 name 常是 w_qry_xxx，中文标题挂在它里面的 Form 上
		// （实测 #586 w_qry_q_pmaa001_3 / 里面的 Form 才叫「供应商编号查询」），
		// 所以匹配要连子对象的名字一起看，否则 --window 供应商 会找不到。
		match := func(w *fep.Object) bool {
			if want == "" {
				return true
			}
			names := []string{w.Attr("name"), w.Attr("title"), w.Attr("text")}
			for _, ch := range kids[w.ID] {
				names = append(names, ch.Attr("name"), ch.Attr("title"), ch.Attr("text"))
			}
			for _, n := range names {
				if strings.Contains(strings.ToLower(n), want) {
					return true
				}
			}
			return false
		}
		var wins []*fep.Object
		for _, w := range m.Windows() {
			if match(w) {
				wins = append(wins, w)
			}
		}
		if len(wins) == 0 {
			return fmt.Errorf("没有匹配的窗口：%q（--window 传名字片段；窗口清单看 ui snapshot）", uiTreeWindow)
		}
		depth := uiTreeDepth
		if depth <= 0 {
			depth = 6
		}
		for _, w := range wins {
			title := w.Attr("name")
			if title == "" {
				title = w.Attr("title")
			}
			fmt.Printf("窗口 #%d  %s\n", w.ID, title)
			printSubtree(kids, r, w.ID, 1, depth, map[int]bool{})
		}
		return nil
	},
}

// printSubtree 打印一棵子树。seen 防环（动态开窗的父链实测出现过来回指的情况）。
func printSubtree(kids map[int][]*fep.Object, r *fep.Resolver, id, level, maxDepth int, seen map[int]bool) {
	if level > maxDepth {
		return
	}
	objs := kids[id]
	sort.Slice(objs, func(i, j int) bool { return objs[i].ID < objs[j].ID })
	for _, o := range objs {
		if seen[o.ID] {
			continue
		}
		seen[o.ID] = true
		name := o.Attr("name")
		if fn, ok := r.Resolve(o.ID); ok && fn.Text != "" && fn.Text != name {
			// 名字解析会把 name 盖成中文；这里两个都留着 —— 按源码动作名找对象时靠后者。
			name = fn.Text + " (`" + o.Attr("name") + "`)"
		}
		extra := ""
		if cm := o.Attr("comment"); cm != "" {
			extra = "  comment=" + cm
		}
		if v := o.Attr("value"); v != "" {
			extra += fmt.Sprintf("  value=%q", v)
		}
		fmt.Printf("%s#%-6d %-16s %s%s\n", strings.Repeat("  ", level), o.ID, o.Type, name, extra)
		printSubtree(kids, r, o.ID, level+1, maxDepth, seen)
	}
}

func init() {
	uiTreeCmd.Flags().StringVar(&uiTreeWindow, "window", "", "窗口名字片段（如 供应商编号查询）；缺省列出所有窗口")
	uiTreeCmd.Flags().IntVar(&uiTreeDepth, "depth", 6, "展开到第几层")
}
