package fep

import "fmt"

// TargetInfo 是"这条事件指向哪个控件"的**完整答案**（明细页的检查器照它渲染）。
//
// 为什么要有它、而不是让界面自己拼 Resolve()+Get()+WindowOf()：
// "**有没有**名字"与"**为什么**没有名字"是同一件事的两面 —— 分开做必然两处各说一套。
// 规则（R1–R4）、置信度、父链、断裂原因都住在 fields.go / aui.go 里，答案就该从这里出。
//
// 一条硬纪律：`Confidence == ConfNone` 时 `Name` 必须是空串。界面据此显示 `Type#id`，
// 而**不许**把 id 当名字、也不许拿父对象的 text 顶上（那是早先原型里"信誓旦旦的错名字"的来源）。
type TargetInfo struct {
	ID         int            `json:"id,omitempty"`
	Type       string         `json:"type,omitempty"`
	Name       string         `json:"name,omitempty"`
	Confidence NameConfidence `json:"confidence"`
	Source     string         `json:"source,omitempty"` // comment-key | name-suffix | own-text | inherit#<id>
	Window     int            `json:"window,omitempty"`
	WindowName string         `json:"windowName,omitempty"`
	// Parents 是从**父**到根的链（不含自己）。检查器的"名字来源"页签要它：
	// R2（列名后缀）与 R4（继承）都是沿着这条链找出来的，说得清才叫可复核。
	Parents []TargetParent `json:"parents,omitempty"`
	// Cycle 表示父链里有环（消息残缺/跨窗口合并的痕迹）。有环时名字一律降级为 low。
	Cycle bool `json:"cycle,omitempty"`
	// Why 是"名字定不下来时"的人话原因。**不许**留空：界面要显示出为什么，
	// 而不是一个空白格让人以为是界面坏了。
	Why string `json:"why,omitempty"`
}

// TargetParent 是父链上的一环（只带排障要看的几样）。
type TargetParent struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
	Text string `json:"text,omitempty"`
}

// DescribeTarget 回答"这个 id 现在是什么、叫什么、凭什么"。
//
// m 必须是**这份录制自己的**界面树（H1：id 不是身份，跨界面会撞）；
// r 是它的解析器；m 或 r 为 nil 时照实说"没有界面"，绝不猜。
func DescribeTarget(m *Model, r *Resolver, id int) TargetInfo {
	ti := TargetInfo{ID: id, Confidence: ConfNone}
	if id == 0 {
		ti.Why = "这条事件里没有 idRef：它不指向任何对象（例如纯键盘事件）"
		return ti
	}
	if m == nil {
		ti.Why = "这份录制里没有可用的界面树（AUI 一条都没读出来），对象类型与名字都无从得知"
		return ti
	}
	o, ok := m.Get(id)
	if !ok {
		// 事件指向的对象**没被录到**：最常见的原因是它出现在录制开始之前
		// （模板预载的对象在会话刚开始时就下去了）。
		ti.Why = fmt.Sprintf("这份录制的界面树里没有 #%d：它没被录到（多半在录制开始之前就下发过了）", id)
		return ti
	}
	ti.Type = o.Type
	if w := m.WindowOf(id); w != 0 {
		ti.Window = w
		ti.WindowName = m.WindowName(w)
	}
	if chain, cycle, _ := m.Chain(id); len(chain) > 0 {
		ti.Cycle = cycle
		for _, c := range chain[1:] {
			ti.Parents = append(ti.Parents, TargetParent{ID: c.ID, Type: c.Type, Text: c.Attr("text")})
		}
	}
	if r == nil {
		ti.Why = "没有解析器：类型照实给出，名字一栏空着"
		return ti
	}
	fn, ok := r.Resolve(id)
	if !ok {
		ti.Why = "R1–R4 都没命中：这个对象没有 comment（字典标签）、没有 col（列名）、自己也没有 text"
		return ti
	}
	ti.Name, ti.Source, ti.Confidence = fn.Text, fn.Source, fn.Confidence
	if fn.Confidence == ConfLow {
		ti.Why = "父链断裂或跨窗口（消息残缺）：名字只能排障用，不能当手册名引用"
	}
	return ti
}
