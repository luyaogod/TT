package fep

import (
	"fmt"
	"sort"
	"strings"
)

// —— 这一步之后，树到底发生了什么 ——
//
// 为什么要它（用户的原话）：基础设施"每次执行命令过后，应该同时返回
// Tree 发生了什么变化、现在焦点在哪里、发生了哪些自动带值、控件的哪些值发生了改变"。
//
// 这三个问题以前都答不上来：只报"运行期 +N 字节"，人只能自己猜发生了什么。
// 而它们恰恰是**操作手册**要写的东西 —— 自动带值（供应商编号 → 供应商名称）就是
// 用户眼里"系统自己填好了"的那一步，必须能看见。

// NodeBrief 是树差分里的一条对象摘要（够人读懂，不刷屏）。
type NodeBrief struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name,omitempty"`
	Comment string `json:"comment,omitempty"`
}

func (n NodeBrief) String() string {
	s := fmt.Sprintf("#%d %s", n.ID, n.Type)
	if n.Name != "" {
		s += " " + n.Name
	}
	if n.Comment != "" {
		s += " (" + n.Comment + ")"
	}
	return s
}

// TreeDiff 是两个 AUI 模型之间的差异。
type TreeDiff struct {
	AddedCount    int         `json:"addedCount"`
	RemovedCount  int         `json:"removedCount"`
	UpdatedCount  int         `json:"updatedCount"`
	Added         []NodeBrief `json:"added,omitempty"` // 只留"值得看"的（有名字/是窗口对话框动作），有条数上限
	Removed       []NodeBrief `json:"removed,omitempty"`
	OpenedWindows []string    `json:"openedWindows,omitempty"`
	ClosedWindows []string    `json:"closedWindows,omitempty"`
	GainedActions []string    `json:"gainedActions,omitempty"`
	LostActions   []string    `json:"lostActions,omitempty"`
}

// maxBrief 是差分摘要的条数上限：报告是给人看的，不是把树倒出来。
const maxBrief = 12

// Empty 表示"树一点没动"。
func (d *TreeDiff) Empty() bool {
	return d == nil || (d.AddedCount == 0 && d.RemovedCount == 0 && d.UpdatedCount == 0 &&
		len(d.OpenedWindows) == 0 && len(d.ClosedWindows) == 0 &&
		len(d.GainedActions) == 0 && len(d.LostActions) == 0)
}

// DiffModels 比两个模型，回答"树变了什么"。
//
// 口径：按**裸 id** 比（模型里 Get(id) 给的是"最近声明的那一份"，与事件寻址口径一致）；
// 只有**新增/删除**才进摘要列表（"改动"只报数量 —— 属性改动太碎，列出来会淹掉信号）。
func DiffModels(before, after *Model) *TreeDiff {
	d := &TreeDiff{}
	if before == nil || after == nil {
		return d
	}
	beforeObjs := map[int]*Object{}
	for _, o := range before.Objects() {
		if cur, ok := before.Get(o.ID); ok {
			beforeObjs[o.ID] = cur
		}
	}
	afterObjs := map[int]*Object{}
	for _, o := range after.Objects() {
		if cur, ok := after.Get(o.ID); ok {
			afterObjs[o.ID] = cur
		}
	}

	worth := func(o *Object) bool {
		return o.Attr("name") != "" || o.Attr("comment") != "" ||
			o.Type == "Window" || o.Type == "Dialog" || isActionLike(o.Type)
	}
	brief := func(o *Object) NodeBrief {
		return NodeBrief{ID: o.ID, Type: o.Type, Name: o.Attr("name"), Comment: o.Attr("comment")}
	}
	appendBrief := func(dst *[]NodeBrief, o *Object) {
		if len(*dst) >= maxBrief || !worth(o) {
			return
		}
		*dst = append(*dst, brief(o))
	}

	for id, o := range afterObjs {
		if _, had := beforeObjs[id]; !had {
			d.AddedCount++
			appendBrief(&d.Added, o)
			continue
		}
		if !sameAttrs(beforeObjs[id], o) {
			d.UpdatedCount++
		}
	}
	for id, o := range beforeObjs {
		if _, still := afterObjs[id]; !still {
			d.RemovedCount++
			appendBrief(&d.Removed, o)
		}
	}

	// 窗口开关
	beforeWin, afterWin := windowNameSet(before), windowNameSet(after)
	for n := range afterWin {
		if _, had := beforeWin[n]; !had && n != "" {
			d.OpenedWindows = append(d.OpenedWindows, n)
		}
	}
	for n := range beforeWin {
		if _, still := afterWin[n]; !still && n != "" {
			d.ClosedWindows = append(d.ClosedWindows, n)
		}
	}

	// 动作的增失（"这一步之后能做什么了"—— 驱动与手册都要看）
	beforeAct, afterAct := actionNameSet(before), actionNameSet(after)
	for n := range afterAct {
		if _, had := beforeAct[n]; !had {
			d.GainedActions = append(d.GainedActions, n)
		}
	}
	for n := range beforeAct {
		if _, still := afterAct[n]; !still {
			d.LostActions = append(d.LostActions, n)
		}
	}

	sort.Strings(d.OpenedWindows)
	sort.Strings(d.ClosedWindows)
	sort.Strings(d.GainedActions)
	sort.Strings(d.LostActions)
	sort.Slice(d.Added, func(i, j int) bool { return d.Added[i].ID < d.Added[j].ID })
	sort.Slice(d.Removed, func(i, j int) bool { return d.Removed[i].ID < d.Removed[j].ID })
	return d
}

// sameAttrs 判两个对象是不是"同一样子"。
//
// 为什么不用 Seq 比：Seq 是**同一个模型内**的声明/触碰序，两个独立建起来的模型
// 各自从 1 开始，拿它比会把"属性确实变了"错判成没变（测试就抓到了这一点）。
func sameAttrs(a, b *Object) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Type != b.Type || a.Parent != b.Parent || len(a.Attrs) != len(b.Attrs) {
		return false
	}
	for k, v := range a.Attrs {
		if b.Attrs[k] != v {
			return false
		}
	}
	return true
}

func windowNameSet(m *Model) map[string]bool {
	out := map[string]bool{}
	for _, w := range m.Windows() {
		if n := w.Attr("name"); n != "" {
			out[n] = true
		}
	}
	return out
}

func actionNameSet(m *Model) map[string]bool {
	out := map[string]bool{}
	for _, o := range m.Objects() {
		if !isActionLike(o.Type) {
			continue
		}
		if n := o.Attr("name"); n != "" {
			out[n] = true
		}
	}
	return out
}

// Summary 打成一行给人看（空则返回"树没动"）。
func (d *TreeDiff) Summary() string {
	if d == nil {
		return ""
	}
	if d.Empty() {
		return "树没动"
	}
	var parts []string
	if d.AddedCount > 0 {
		parts = append(parts, fmt.Sprintf("+%d 节点", d.AddedCount))
	}
	if d.RemovedCount > 0 {
		parts = append(parts, fmt.Sprintf("-%d 节点", d.RemovedCount))
	}
	if d.UpdatedCount > 0 {
		parts = append(parts, fmt.Sprintf("改动 %d 处", d.UpdatedCount))
	}
	if len(d.OpenedWindows) > 0 {
		parts = append(parts, "开窗："+strings.Join(d.OpenedWindows, ","))
	}
	if len(d.ClosedWindows) > 0 {
		parts = append(parts, "关窗："+strings.Join(d.ClosedWindows, ","))
	}
	if len(d.GainedActions) > 0 {
		parts = append(parts, "新增动作 "+strings.Join(head(d.GainedActions, 6), "/"))
	}
	if len(d.LostActions) > 0 {
		parts = append(parts, "消失动作 "+strings.Join(head(d.LostActions, 6), "/"))
	}
	return strings.Join(parts, "；")
}

func head(ss []string, n int) []string {
	if len(ss) <= n {
		return ss
	}
	return ss[:n]
}
