package fep

import (
	"fmt"
	"strings"

	"tt/internal/fep"
)

// printImpact 渲染"这一步之后发生了什么"。
//
// 用户对基础设施的要求（原话）：每次执行命令过后，应该同时返回
//
//	· Tree 发生了什么变化
//	· 现在焦点在哪里
//	· 发生了哪些自动带值
//	· 控件的哪些值发生了改变
//
// 这里把四件事按固定顺序打出来，一眼能核对；JSON 形态里它们分别是
// changes.tree / changes.focus / changes.values[].kind / changes.values[]。
func printImpact(ch *fep.Changes, wait string) {
	if ch == nil {
		return
	}
	switch {
	case ch.TimedOut:
		fmt.Printf("  运行期在 %s 内没有回包（事件发出去了，作业没显式变化）\n", wait)
	case ch.RuntimeBytes > 0:
		fmt.Printf("  运行期 +%d 字节 / 新消息 %d 条\n", ch.RuntimeBytes, ch.NewMessages)
	}

	// ① 树：变了什么
	if t := ch.Tree; t != nil {
		fmt.Printf("  树：%s\n", t.Summary())
		for _, n := range headBrief(t.Added, 6) {
			fmt.Printf("      + %s\n", n)
		}
		for _, n := range headBrief(t.Removed, 6) {
			fmt.Printf("      - %s\n", n)
		}
	}

	// ② 值：哪个是我写的、哪个是系统自动带出来的、哪个归属不明（不猜）
	var sent, auto, plain []fep.ValueChange
	for _, v := range ch.Values {
		switch v.Kind {
		case "sent":
			sent = append(sent, v)
		case "auto":
			auto = append(auto, v)
		default:
			plain = append(plain, v)
		}
	}
	printValueGroup("我写的", sent)
	printValueGroup("自动带值", auto)
	printValueGroup("值变化", plain)

	// ③ 焦点：现在在哪（下一步能不能动手的前置条件）
	if f := ch.Focus; f != nil {
		who := fmt.Sprintf("#%d", f.FocusID)
		if f.FocusKind != "" {
			who += " " + f.FocusKind
		}
		if f.FocusName != "" {
			who += " " + f.FocusName
		}
		line := "  焦点：" + who
		if f.Grid != 0 {
			line += fmt.Sprintf("（网格 #%d 第 %d 行 第 %d 列）", f.Grid, f.Row+1, f.Col+1)
		}
		fmt.Println(line)
	}

	if ch.RuntimeError != "" {
		fmt.Printf("  ⚠ 作业停下：%s\n", ch.RuntimeError)
	}
}

// maxValuePrint 是每类值变化最多打几行。
//
// 为什么要有上限：开窗一取数就是整屏格子从空变有值（实测一次 374 行），
// 全列出来会把真正要看的那两行（供应商编号 / 供应商名称）淹掉。
// 报告是给人看的"这一步发生了什么"，不是把树倒出来；要看全部走 --format json。
const maxValuePrint = 8

func printValueGroup(title string, vs []fep.ValueChange) {
	if len(vs) == 0 {
		return
	}
	for i, v := range vs {
		if i >= maxValuePrint {
			fmt.Printf("      · %s：…还有 %d 处（--format json 看全部）\n", title, len(vs)-maxValuePrint)
			break
		}
		switch v.Kind {
		case "sent":
			fmt.Printf("      · %s：%s %q → %q", title, valueLabel(v), v.From, v.To)
		case "auto":
			fmt.Printf("      · %s：%s → %q", title, valueLabel(v), v.To)
		default:
			fmt.Printf("      · %s：%s %q → %q", title, valueLabel(v), v.From, v.To)
		}
		if v.Comment != "" {
			fmt.Printf("  (%s)", v.Comment)
		}
		fmt.Println()
	}
}

// valueLabel 给人看的字段名（中文名优先，其次协议名，最后 #id）。
func valueLabel(v fep.ValueChange) string {
	switch {
	case v.Role != "" && v.Label != "" && v.Role != v.Label:
		return v.Label + "[" + v.Role + "]"
	case v.Label != "":
		return v.Label
	case v.Role != "":
		return v.Role
	}
	return fmt.Sprintf("#%d", v.ID)
}

func head(ss []string, n int) []string {
	if len(ss) <= n {
		return ss
	}
	return ss[:n]
}

// briefNodeSummary 是给"树没动但有动作增失"这类情况的兜底可读串（供 job 用）。
func briefNodeSummary(ch *fep.Changes) string {
	if ch == nil || ch.Tree == nil {
		return ""
	}
	var parts []string
	if len(ch.Tree.GainedActions) > 0 {
		parts = append(parts, "新增动作 "+strings.Join(head(ch.Tree.GainedActions, 6), "/"))
	}
	if len(ch.Tree.LostActions) > 0 {
		parts = append(parts, "消失动作 "+strings.Join(head(ch.Tree.LostActions, 6), "/"))
	}
	return strings.Join(parts, "；")
}

// headBrief 是 head 在节点摘要上的版本（Go 泛型要 1.18+，这里直接写两份更省事）。
func headBrief(ns []fep.NodeBrief, n int) []fep.NodeBrief {
	if len(ns) <= n {
		return ns
	}
	return ns[:n]
}
