package fep

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"tt/internal/output"
)

// 事件构造与发前校验（计划 §3.8 / §6.3）。
//
// 三种事件的形状都是逐字节对着**真实录制件**里客户端自己发出的事件写的，不是照
// 记忆或推断：
//
//	ActionEvent 0{{idRef "5160"}}
//	KeyEvent 0{{keyName "Tab"}}
//	ConfigureEvent 0{{idRef "60"}{ value "LUYAO"}}
//	ConfigureEvent 0{{idRef "6543"}{ cursor "5"}{ cursor2 "5"}}
//
// 注意两个细节：
//   - 客户端的事件里，第一个属性紧贴 `{{`（`{{idRef "…"}`），后续属性前面**有一个
//     空格**（`{ value "…"}`）。这里照抄，宁可跟客户端逐字节一致。
//   - 「输入必须带 cursor/cursor2」是早期探针的结论，**与录制件不符**：真人打字
//     产生的是 value-only 事件，光标事件是另一条（客户端自己发）。所以构造器默认
//     发 value-only，要补光标事件时由调用方显式给 caret。
//
// 发前校验只做判断，**不发包** —— 拒绝时返回 *Refusal，调用方拿它映射退出码。

// ClickableTypes 是实测能接受 ActionEvent 的对象类型（白名单，计划 §6.3 第 2 条）。
// 列表来自真机录制里被点击过的对象类型，不是推断出来的全集。
var ClickableTypes = map[string]bool{
	"ToolBarItem":    true,
	"Action":         true,
	"ActionDefault":  true,
	"TopMenuCommand": true,
	"ButtonEdit":     true,
	"Item":           true,
	"Value":          true,
	"FormField":      true,
	// MenuAction：菜单里的可执行项。为什么它必须可点（实测依据）：
	// 运行期在**等一次确认**时会自己把焦点指到它身上 —— 实测错误窗（cl_err_s01）
	// 弹出来时运行期推 `un 0 {{focus "690"}}`，而 #690 就是 MenuAction{name "OK"}。
	// 也就是说"运行期正在等你去按它"。旧白名单漏了它，导致错误窗只能靠瞎按键去猜（还猜错）。
	"MenuAction": true,
	"Menu":       true,
}

// writableTypes 是"能把值发过去"的对象类型：字段值本身与它的编辑控件。
// 与 ClickableTypes 一样，来自真机录制里被输入过的对象，不是推断出来的全集。
var writableTypes = map[string]bool{
	"FormField":  true,
	"Value":      true,
	"Edit":       true,
	"ButtonEdit": true,
	"ComboBox":   true,
	"DateEdit":   true,
}

// Refusal is a pre-flight refusal: nothing was sent.
type Refusal struct {
	Code    string
	Exit    int
	Message string
	Hint    string
}

func (r *Refusal) Error() string { return r.Message }

func refuse(code string, exit int, format string, a ...any) *Refusal {
	return &Refusal{Code: code, Exit: exit, Message: fmt.Sprintf(format, a...)}
}

// ActionEventText renders a click on idRef.
func ActionEventText(idRef int) string {
	return fmt.Sprintf(`ActionEvent 0{{idRef %q}}`, strconv.Itoa(idRef))
}

// KeyEventText renders a key press. Keys carry no idRef (plan §3.8).
func KeyEventText(keyName string) string {
	return fmt.Sprintf(`KeyEvent 0{{keyName %q}}`, keyName)
}

// ConfigureValueText renders typing into a field, in the shape the client itself
// emits. When caret is non-nil the cursor pair is appended: the client sends that
// pair as its own event when the caret moves, so some fields want it.
func ConfigureValueText(id int, value string, caret *int) string {
	s := fmt.Sprintf(`ConfigureEvent 0{{idRef %q}{ value %q}`, strconv.Itoa(id), value)
	if caret != nil {
		s += fmt.Sprintf(`{ cursor %q}{ cursor2 %q}`, strconv.Itoa(*caret), strconv.Itoa(*caret))
	}
	return s + "}"
}

// WrapEvent wraps one event the way the client does: `event _om <seq>{}{{<event>}}`.
func WrapEvent(seq int, ev string) string {
	return fmt.Sprintf("event _om %d{}{{%s}}", seq, strings.TrimSpace(ev))
}

// EncodeEvent renders a wrapped event as wire bytes.
func EncodeEvent(seq int, ev string) []byte {
	return Encode(Frame{Type: FrameTypeAUI, Payload: []byte(WrapEvent(seq, ev))})
}

// CheckClick runs the pre-flight rules for a click and sends nothing.
//
// The sequence rule (plan §3.8 rule 4) is enforced by construction in the injector:
// it always uses the connection's next number and never accepts one from a caller.
// Replay computes its own numbers, so it calls CheckSeq separately.
func CheckClick(m *Model, idRef int) error {
	o, ok := m.Get(idRef)
	if !ok {
		return refuse(output.CodeTargetMissing, 5, "AUI 里没有 #%d：目标不存在（不发包）", idRef)
	}
	if !ClickableTypes[o.Type] {
		return refuse(output.CodeTargetMissing, 5, "类型 %s 不在可点击白名单里（第 2 条）", o.Type)
	}
	if o.Type == "ButtonEdit" && o.Attr("action") != "controlp" {
		return refuse(output.CodeTargetMissing, 5,
			"ButtonEdit#%d 没有 controlp（不是开窗按钮）：不点（第 2 条）", idRef)
	}
	if v := strings.ToLower(o.Attr("actionActive")); v == "0" || v == "false" {
		return refuse(output.CodeSessionState, 4, "目标 #%d 的 actionActive 为假：当前不可用（第 3 条）", idRef)
	}
	// 动作必须属于**当前交互指令**（官方 1514 + 实测：运行期每次状态变化都重分配动作 id）。
	// 这一条在发前就拦：发出去只会回空 ACK（看起来像"没反应"，实际是发错了对象）。
	if err := CheckActionInCurrentDialog(m, idRef); err != nil {
		return err
	}
	return nil
}

// CheckSeq enforces plan §3.8 rule 4 for callers that compute their own numbers:
// reusing a number resets the connection.
func CheckSeq(want, actual int) error {
	if want != actual {
		return refuse(output.CodeSessionState, 4,
			"_om 序号不对：这条连接的下一个是 %d，你要发 %d（第 4 条，重号会打断连接）", actual, want)
	}
	return nil
}

// CheckKey runs the pre-flight rules for a key press.
func CheckKey(keyName string) error {
	if strings.TrimSpace(keyName) == "" {
		return refuse(output.CodeUsage, 2, "没给按键名（如 Tab / Enter / Return）")
	}
	return CheckRaw(KeyEventText(keyName))
}

// CheckRaw checks a raw event text, the form replay feeds in: a KeyEvent must not
// carry an idRef (an event that does is not something the client ever sends).
func CheckRaw(ev string) error {
	if strings.Contains(ev, "KeyEvent") && strings.Contains(ev, "idRef") {
		return refuse(output.CodeUsage, 2, "KeyEvent 不能带 idRef（第 2 条）：%s", ev)
	}
	if strings.Contains(ev, "ConfigureEvent") && !strings.Contains(ev, "idRef") {
		return refuse(output.CodeUsage, 2, "ConfigureEvent 必须带 idRef（发给谁）：%s", ev)
	}
	// 空 idRef 是**能打断连接**的：实测手滑把格子 id 取成空串（`{idRef ""}{value "Y"}`）之后，
	// 那条连接直接消失（`conns: []`），整个作业要重开。它的形状看起来"无害" ——
	// 正因为看起来无害，才会有人（我）发出去。所以在这里挡掉。
	if strings.Contains(ev, `idRef ""`) {
		return refuse(output.CodeUsage, 2,
			"事件里的 idRef 是空的：这会打断连接（实测把作业整条连接打死）—— %s", ev)
	}
	return nil
}

// ResolveInputTarget answers "who should the typing go to".
//
// **只发给 FormField**（或它自己就是可写的 Value）。
//
// 这条规则的代价是实测出来的：真机上把 `ConfigureEvent{value}` 发给一个 `Edit`（字段
// 容器本身）会让**作业进程停下来**：
//
//	FORMS statement error number -6327
//	Internal error in the run time library file ../ui/indialog.c(10523):0
//	Program stopped at 'axm_axmt500_wf.4gl', line number 19839
//
// 所以这里宁可不做也不做错：找不到 FormField 就**拒发**，让调用方去查清对象关系。
func ResolveInputTarget(m *Model, id int) (target int, viaFormField bool, err error) {
	o, ok := m.Get(id)
	if !ok {
		return 0, false, refuse(output.CodeTargetMissing, 5, "AUI 里没有 #%d：目标不存在（不发包）", id)
	}
	if o.Type == "FormField" {
		return id, true, nil
	}
	if o.Type == "Value" {
		// 实测里字段当前值就写作 Value，它自己就是可写的目标。
		return id, false, nil
	}
	if p, ok := m.Get(o.Parent); ok && p.Type == "FormField" {
		// 真机布局里 FormField 常常是容器的**父对象**（FormField 271 包着 ButtonEdit 272）。
		return p.ID, true, nil
	}
	for _, c := range m.Objects() {
		if c.Parent == id && c.Type == "FormField" {
			return c.ID, true, nil
		}
	}
	return 0, false, refuse(output.CodeTargetMissing, 5,
		"#%d 是 %s，不是可写的 FormField：发给容器本身会让作业停在 -6327（实测）；"+
			"先 tt fep ui find --text <栏位名> 找到它的 FormField 再发", id, o.Type)
}

// FindType returns objects of one type sorted by id (small helper for commands
// that take a type and an index).
func FindType(m *Model, t string) []*Object {
	objs := m.ByType(t)
	for i := 1; i < len(objs); i++ {
		for j := i; j > 0 && objs[j-1].ID > objs[j].ID; j-- {
			objs[j-1], objs[j] = objs[j], objs[j-1]
		}
	}
	return objs
}

// ConfigureCursorText renders the caret-only event the client emits when the
// caret moves inside a field (`ConfigureEvent 0{{idRef "6543"}{ cursor "5"}{ cursor2 "5"}}`).
func ConfigureCursorText(id, caret int) string {
	return fmt.Sprintf(`ConfigureEvent 0{{idRef %q}{ cursor %q}{ cursor2 %q}}`,
		strconv.Itoa(id), strconv.Itoa(caret), strconv.Itoa(caret))
}

// IsWritableType reports whether a type can take a value (the same list the
// pre-flight uses). 标注该画在**栏位**上而不是它旁边的标签上，所以要能问这一句。
func IsWritableType(t string) bool { return writableTypes[t] }

// ConfigureGridCursorText 渲染"把光标放进网格第 row 行第 col 列"的事件。
//
// 实测形态（真人录制里就有）：`ConfigureEvent 0{{idRef "644"}{ currentRow "1"}{ currentColumn "2"}}`
// —— 明细列的开窗按钮要求光标先落在那一行那一列，否则 actionActive=0（点了没反应）；
// 而 INPUT ARRAY 的 `ON CHANGE` 也只认**当前行**的改动，所以写勾选前必须先把光标放进去。
func ConfigureGridCursorText(grid int, row, col int) string {
	return fmt.Sprintf(`ConfigureEvent 0{{idRef %q}{ currentRow %q}{ currentColumn %q}}`,
		strconv.Itoa(grid), strconv.Itoa(row), strconv.Itoa(col))
}

// ConfigureGridRowText 渲染"把当前行移到第 row 行"的事件（不碰列）。
//
// 实测形态：`ConfigureEvent 0{{idRef "122"}{ currentRow "1"}}`
// —— 运行期会回 `un 122 {{currentRow "1"}}`，这是 DISPLAY ARRAY 的"当前行式"选中要的前置。
func ConfigureGridRowText(grid int, row int) string {
	return fmt.Sprintf(`ConfigureEvent 0{{idRef %q}{ currentRow %q}}`,
		strconv.Itoa(grid), strconv.Itoa(row))
}

// （空 idRef 的检查已并入 CheckRaw —— 那条铁律在那里。）

// firstConfigureTarget 从一条事件体里抠出第一个 ConfigureEvent 的 idRef 与 value。
//
// 用途：`raw` 形态的步骤（实测：写格子 = 值与光标**同一条消息**，拆开发运行期不认）
// 发出去的是一串事件，报告里要能说出"这一串里我写的是哪个对象、什么值"，
// 才能把变化分成"我写的"与"运行期自动带出来的"。
//
// 抠不到就返回 0/""（调用方按"不知道"处理，不猜）。
func firstConfigureTarget(ev string) (int, string) {
	idx := reConfigureID.FindAllStringSubmatchIndex(ev, -1)
	if len(idx) == 0 {
		return 0, ""
	}
	id, err := strconv.Atoi(ev[idx[0][2]:idx[0][3]])
	if err != nil {
		return 0, ""
	}
	// value 只认**紧跟在这个 id 之后**的那一个（同一串里后面可能还有别的对象的值）
	tail := ev[idx[0][1]:]
	if vm := reConfigureValue.FindStringSubmatch(tail); vm != nil {
		return id, vm[1]
	}
	return id, ""
}

var (
	reConfigureID    = regexp.MustCompile(`ConfigureEvent 0\{\{idRef "([0-9]+)"`)
	reConfigureValue = regexp.MustCompile(`\{ value "([^"]*)"\}`)
)
