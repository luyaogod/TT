package fep

import (
	"fmt"
	"strconv"
	"strings"
)

// Eligibility 回答一个具体问题：**此刻**这个控件能不能用、为什么不能。
//
// 判据只认**运行期广播的真实属性**（属性词表是从 94 个真实连接里挖出来的，
// 见 tools/gdc-uia/README.md）：
//
//	active        节点此刻是否可用（运行期发的**最频繁**的属性）
//	hidden        是否隐藏
//	noEntry       是否不允许进入/录入
//	actionActive  动作此刻是否可用
//	dialogType    当前处于哪种对话框（连接级，且会粘 → 只对"推它时那个窗口"有效）
//
// 原则与动作层防线一致：**判据不足不拦** —— 只有拿到"不可用"的**正证据**才拒绝。
// 为什么：拦错了会让本来能做的事做不成（而且驱动会以为自己在防错），
// 漏判最多是多一次被运行期拒绝的发送（有回显可查）。
type Eligibility struct {
	Usable bool   `json:"usable"`
	Reason string `json:"reason,omitempty"` // 不可用的人话原因
	By     string `json:"by,omitempty"`     // 依据：runtime / active / hidden / noEntry / actionActive / dialog / source
}

func (e Eligibility) String() string {
	if e.Usable {
		return "可用（依据 " + e.By + "）"
	}
	return "不可用：" + e.Reason + "（依据 " + e.By + "）"
}

// Eligible 判一个对象此刻能不能用（不看对话框归属）。
func Eligible(m *Model, o *Object) Eligibility {
	if o == nil {
		return Eligibility{Usable: false, Reason: "对象不存在", By: "model"}
	}
	return eligibleAttrs(o)
}

// EligibleIn 在 Eligible 之上再加一条：**必须属于指定的那个对话框**。
//
// 为什么必须收窄：实测动作 id 每次换状态都会被重新分配同名动作（#109 → #153 → #4538），
// 发进旧对话框的动作只会得到空 ACK。dialog 传 0 表示不做这项检查。
func EligibleIn(m *Model, o *Object, dialog int) Eligibility {
	if e := Eligible(m, o); !e.Usable {
		return e
	}
	if dialog == 0 {
		return Eligibility{Usable: true, By: "runtime"}
	}
	if d, ok := DialogOf(m, o.ID); !ok || d != dialog {
		got := 0
		if ok {
			got = d
		}
		return Eligibility{
			Usable: false,
			Reason: fmt.Sprintf("它不属于当前交互指令（属于 Dialog #%d，当前是 #%d）", got, dialog),
			By:     "dialog",
		}
	}
	return Eligibility{Usable: true, By: "runtime+dialog"}
}

// EligibleControl 判一个目录项（可能带源码信息）。
//
// 与 Eligible 的差别：源码里标了 noEntry（不可录入）或它是只读类型时，
// **即便运行期没广播 noEntry** 也拦下写值 —— 这是源码级的正证据。
func EligibleControl(m *Model, c Control) Eligibility {
	if c.Source != nil {
		if c.Source.NoEntry {
			return Eligibility{Usable: false, Reason: "源码标为不可录入（noEntry=true）：" + c.Source.Form, By: "source"}
		}
	}
	return eligibleAttrs(&Object{ID: c.ID, Type: c.Kind, Attrs: map[string]string{
		"active": c.Active, "hidden": c.Hidden, "noEntry": c.NoEntry,
		"actionActive": c.ActionActive,
	}})
}

func eligibleAttrs(o *Object) Eligibility {
	if o.Attr("hidden") == "1" {
		return Eligibility{Usable: false, Reason: "运行期标为隐藏（hidden=1）", By: "hidden"}
	}
	if o.Attr("active") == "0" {
		return Eligibility{Usable: false, Reason: "运行期标为不可用（active=0）", By: "active"}
	}
	if o.Attr("noEntry") == "1" {
		return Eligibility{Usable: false, Reason: "运行期标为不可录入（noEntry=1）", By: "noEntry"}
	}
	if isActionLike(o.Type) && o.Attr("actionActive") == "0" {
		return Eligibility{Usable: false, Reason: "动作此刻不可用（actionActive=0）", By: "actionActive"}
	}
	return Eligibility{Usable: true, By: "runtime"}
}

// DialogState 是"当前这一步处在什么交互指令里"的**可核对**描述。
//
// 它是驱动与人工共用的那一句话：现在是什么对话框、属于哪个窗口、提供哪些可用动作。
// dialogType 只对"推它时那个窗口"有效（运行期不复位、会粘），所以这里
// 既给窗口自己的结论，也给连接级的原始值，供交叉核对。
type DialogState struct {
	Window       string   `json:"window"`
	WindowID     int      `json:"windowId"`
	DialogType   string   `json:"dialogType,omitempty"`
	DialogID     int      `json:"dialogId,omitempty"`
	DialogActive bool     `json:"dialogActive"`
	Actions      []string `json:"actions,omitempty"` // 此刻可用的动作名（actionActive != "0"）
	Tables       []string `json:"tables,omitempty"`
}

// DialogStates 列出所有"活着的窗口"，各自处在哪种对话框、有哪些可用动作。
func (m *Model) DialogStates() []DialogState {
	var out []DialogState
	for _, w := range m.Windows() {
		ds := DialogState{Window: m.WindowName(w.ID), WindowID: w.ID}
		ds.DialogType = m.DialogTypeFor(w.ID)
		if d, ok := m.dialogOfWindow(w.ID); ok {
			ds.DialogID = d.ID
			ds.DialogActive = d.Attr("active") != "0"
		}
		for _, o := range m.Objects() {
			if !isActionLike(o.Type) || o.Attr("name") == "" {
				continue
			}
			if m.WindowOf(o.ID) != w.ID {
				continue
			}
			if o.Attr("actionActive") == "0" {
				continue
			}
			ds.Actions = appendUnique(ds.Actions, o.Attr("name"))
		}
		for _, t := range m.TablesOfWindow(w.ID) {
			ds.Tables = append(ds.Tables, fmt.Sprintf("#%d", t))
		}
		out = append(out, ds)
	}
	return out
}

// dialogOfWindow 找一个窗口实例下的 Dialog 对象（取 Seq 最大的那份 = 当前那个）。
func (m *Model) dialogOfWindow(win int) (*Object, bool) {
	var best *Object
	for _, o := range m.Objects() {
		if o.Type != "Dialog" || m.WindowOf(o.ID) != win {
			continue
		}
		if best == nil || o.Seq > best.Seq {
			best = o
		}
	}
	return best, best != nil
}

// ---------------------------------------------------------------------------
// 焦点（focus）：运行期认动作的**前置条件**
// ---------------------------------------------------------------------------

// Focus 返回运行期最近一次告知的焦点对象（`un 0 {{focus "N"}}`）。
//
// 为什么它必须是一等公民（我们自己踩过的教训，见内部/fep/replay.go 的注释）：
// 运行期**只在"光标现在确实在这个栏位"时才认后面的动作**。回放那条线早就修过了
// （把 ConfigureEvent{cursor…} 当导航噪声滤掉 → 回放跑不动）；但驱动那条线一直没管，
// 于是出现"动作名对、actionActive=1、点下去什么都没发生"——因为**焦点不在这儿**。
func (m *Model) Focus() (int, bool) {
	best, bestSeq := 0, -1
	for _, o := range m.Objects() {
		v := strings.TrimSpace(o.Attr("focus"))
		if v == "" {
			continue
		}
		id, err := strconv.Atoi(v)
		if err != nil || id == 0 {
			continue
		}
		if o.Seq > bestSeq {
			best, bestSeq = id, o.Seq
		}
	}
	if best == 0 {
		return 0, false
	}
	return best, true
}

// CursorField 返回某个窗口里"运行期说光标在这个栏位"的那个对象（cursor 属性最新的一份）。
//
// 它与 Focus 的区别：Focus 是运行期显式推的当前焦点；CursorField 是"光标落在哪个栏位"
// 的第二手证据（实测两者都有，Focus 可能缺，所以驱动两个都看）。
func (m *Model) CursorField(win int) (*Object, bool) {
	var best *Object
	for _, o := range m.Objects() {
		if o.Attr("cursor") == "" && o.Attr("cursor2") == "" {
			continue
		}
		if win != 0 && m.WindowOf(o.ID) != win {
			continue
		}
		if best == nil || o.Seq > best.Seq {
			best = o
		}
	}
	return best, best != nil
}

// FocusReport 是"此刻焦点在哪"的可核对描述 —— 驱动每一步之前都该看它。
type FocusReport struct {
	FocusID   int    `json:"focusId,omitempty"`
	FocusName string `json:"focusName,omitempty"`
	FocusKind string `json:"focusKind,omitempty"`
	WindowID  int    `json:"windowId,omitempty"`
	CursorID  int    `json:"cursorId,omitempty"`
	CursorNam string `json:"cursorName,omitempty"`
	Grid      int    `json:"gridId,omitempty"`
	Row       int    `json:"row"`
	Col       int    `json:"column"`
}

// FocusNow 汇总"现在焦点在哪"。
func (m *Model) FocusNow() FocusReport {
	rep := FocusReport{WindowID: m.CurrentWindow(), Row: -1, Col: -1}
	if id, ok := m.Focus(); ok {
		rep.FocusID = id
		if o, found := m.Get(id); found {
			rep.FocusKind = o.Type
			rep.FocusName = m.WindowName(o.ID)
			if n := o.Attr("name"); n != "" {
				rep.FocusName = n
			}
		}
	}
	if o, ok := m.CursorField(rep.WindowID); ok {
		rep.CursorID = o.ID
		rep.CursorNam = o.Attr("name")
	}
	if id, ok := m.Focus(); ok {
		if g := m.gridOfFocus(id); g != 0 {
			rep.Grid = g
			if go_, found := m.Get(g); found {
				rep.Row = atoiOr(go_.Attr("currentRow"), -1)
				rep.Col = atoiOr(go_.Attr("currentColumn"), -1)
			}
		}
	}
	return rep
}

// gridOfFocus 从焦点对象的父链上找最近的网格。
func (m *Model) gridOfFocus(id int) int {
	cur, ok := m.Get(id)
	for hop := 0; ok && cur != nil && hop < 24; hop++ {
		if cur.Type == "Grid" || cur.Type == "ScrollGrid" || cur.Type == "Table" {
			return cur.ID
		}
		if cur.Parent == 0 {
			break
		}
		cur, ok = m.Get(cur.Parent)
	}
	return 0
}

func atoiOr(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return n
	}
	return def
}

// FocusOnField 判"焦点是不是就在这个栏位（或它的格子上）"。
//
// 用途：驱动在按"字段级动作"（如 `s_detail1.pmdn001.controlp` 开窗）之前，
// 必须先确认这件事 —— 否则那一按就是**空动作**（实测：点了没反应、不报错，
// 让人误以为是形状不对）。判据不足（运行期没推 focus 也没推 cursor）时返回 unknown，
// 由调用方决定要不要照发（默认照发，保持"判据不足不拦"）。
func (m *Model) FocusOnField(fieldID int) (on bool, known bool) {
	if fieldID == 0 {
		return false, false
	}
	known = false
	if id, ok := m.Focus(); ok {
		known = true
		if id == fieldID {
			return true, true
		}
	}
	if o, ok := m.CursorField(0); ok {
		known = true
		if o.ID == fieldID {
			return true, true
		}
	}
	return false, known
}
