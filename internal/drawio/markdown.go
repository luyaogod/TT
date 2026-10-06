package drawio

import (
	"fmt"
	"sort"
	"strings"
)

// 形状清单：给 AI 看的紧凑表格。
//
// 为什么单独渲染一份而不是让它去读 mxlibrary：库文件里每条都是一整段转义过的 XML，
// 读它既费上下文又容易看错；清单只列 id / 尺寸 / 槽位 / 标签，够它选控件。
// 刻意做紧凑 —— 这份是要进上下文的东西。

// CatalogMarkdown 渲染全部库的形状清单。
func CatalogMarkdown(catalogs []Catalog) string {
	var b strings.Builder
	b.WriteString("# T100 组件库 · 形状清单\n\n")
	b.WriteString("> 由 `tt drawio lib --catalog` 现打，**不要手抄、不要手改**。\n")
	b.WriteString("> 这是给 AI 用的：拼图时只用下表的 `id`，不要去读 mxlibrary 文件，也不要自己编 XML。\n\n")
	b.WriteString("**文字槽位**说明这个控件哪些文字能替换，写进 `text` 对象，例如 `{\"label\": \"供应商\"}`；")
	b.WriteString("`—` 表示不能改文字。\n\n")

	for _, c := range catalogs {
		fmt.Fprintf(&b, "## %s — %s\n\n", c.Name, c.Library.Name)
		if c.Library.Description != "" {
			b.WriteString(c.Library.Description + "\n\n")
		}
		b.WriteString("| id | 名称 | 尺寸 | 文字槽位 | 标签 |\n")
		b.WriteString("| --- | --- | --- | --- | --- |\n")
		for _, s := range c.Shapes {
			dims := fmt.Sprintf("%s×%s", jsNum(s.W), jsNum(s.H))
			if s.Table != nil {
				dims += "（列行可由 spec 的 `table` 参数自定义）"
			}
			if s.MultiTab != nil {
				dims += "（支持 spec 的 `pages` 多页签）"
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s |\n",
				s.ID, s.Title, dims, slotCell(s.Slots), strings.Join(s.Tags, ", "))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// slotCell 把槽位名排成 `a` `b` 的样子；一个都没有时给 `—`。
//
// 排序是刻意的：map 迭代序随机，不排的话同一份清单两次跑出来不一样，
// "产物可复现"这条就假了。
func slotCell(slots map[string]string) string {
	if len(slots) == 0 {
		return "—"
	}
	names := make([]string, 0, len(slots))
	for k := range slots {
		names = append(names, k)
	}
	sort.Strings(names)
	for i, n := range names {
		names[i] = "`" + n + "`"
	}
	return strings.Join(names, " ")
}
