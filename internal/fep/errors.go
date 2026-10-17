package fep

import (
	"fmt"
	"regexp"
	"strings"

	"tt/internal/output"
)

// 运行期报错帧 —— **"作业程序停下了"**。
//
// 形态（2026-10-11 实测原样摘录；那天我自己把料件开窗搞停了一次）：
//
//	om 22 {{an 4811 Menu 39 {{text "ERROR"} {posY "0"} {style "winmsg"}
//	        {comment "Program stopped at 'qry_q_pmaa001_3.4gl', line number 278.\n
//	                  FORMS statement error number -6327.\n
//	                  Internal error in the run time library file ../ui/ndialog.c(10523):0.\n"}}}
//
// 为什么必须**专门**识别它：它不是"这一步没生效"，而是**这个窗口此后只会回空 ACK**。
// 那天我没认出来，于是对着一个已经停掉的窗口连发十几条事件、还当是"形态不对"，
// 接着改参数接着试 —— 每一步都在浪费，而且看不出哪儿错了。
//
// 今天的触发条件只有一类：**把值/事件发给容器**（`Edit`、`ValueList`、`ToolBar` 本身），
// 或发给一个不属于当前交互指令的动作。这正是 `writableTypes` / `ClickableTypes`
// 白名单与"绝不试探"那条纪律存在的理由。
//
// 夹具：`testdata/fep-recordings/program-stop.txt`（现场 20261011-170312 剪出来的），
// 单元测试就靠它，保证这条识别在别的机器上也真的跑。

// RuntimeError 是运行期推上来的一条报错界面消息。
type RuntimeError struct {
	ID      int    `json:"id"`
	Text    string `json:"text,omitempty"`    // 标题，实测 "ERROR"
	Comment string `json:"comment,omitempty"` // 正文，含 Program stopped at …
}

func (e RuntimeError) Error() string {
	if e.Text == "" {
		return e.Comment
	}
	return e.Text + ": " + e.Comment
}

// ProgramStop 判断这是不是"代码停下来"（而不只是提示）。
func (e RuntimeError) ProgramStop() bool {
	return strings.Contains(e.Comment, "Program stopped at")
}

// reStopWhere 从正文里抠出"哪个程序、第几行、什么错号"。
// 实测正文：`Program stopped at 'qry_q_pmaa001_3.4gl', line number 278.\nFORMS statement error number -6327.\n…`
var reStopWhere = regexp.MustCompile(`'([^']+)',\s*line number (\d+)`)

// reStopReason 从正文里抠出错误号（-6327 这类）。
var reStopReason = regexp.MustCompile(`error number (-?\d+)`)

// Where 给出人能直接用的一行定位（抠不出来就回空串，不猜）。
func (e RuntimeError) Where() string {
	m := reStopWhere.FindStringSubmatch(e.Comment)
	if m == nil {
		return ""
	}
	out := fmt.Sprintf("%s:%s", m[1], m[2])
	if r := reStopReason.FindStringSubmatch(e.Comment); r != nil {
		out += " (" + r[1] + ")"
	}
	return out
}

// RuntimeErrors 收集模型里现存的运行期报错帧。
//
// 判据两条（都来自实测：一条来自 170312 那次停摆，一条来自校验提示）：
//   - `Menu` 且 `text` = "ERROR"；
//   - 或任何对象的 `comment` 里写着 "Program stopped at"。
//
// 用 Candidates 而不是 Get：报错框的裸 id 可能被别的对象复用，而"报错过"这件事
// 不该因为 id 复用就看不见。
func RuntimeErrors(m *Model) []RuntimeError {
	var out []RuntimeError
	for _, o := range m.Objects() {
		if o == nil {
			continue
		}
		text, comment := o.Attr("text"), o.Attr("comment")
		if !(text == "ERROR" || strings.Contains(comment, "Program stopped at")) {
			continue
		}
		out = append(out, RuntimeError{ID: o.ID, Text: text, Comment: comment})
	}
	return out
}

// CheckRuntimeHealthy 在动作之后检查"作业还在不在"。
//
// 返回 *Refusal 时调用方应按"运行失败"退出（退出码 1）——因为**这一屏之后的操作
// 全都没意义了**：继续发只会得到一串空 ACK，看起来像"没生效"，实际是程序已经停下。
func CheckRuntimeHealthy(m *Model) error {
	for _, e := range RuntimeErrors(m) {
		if !e.ProgramStop() {
			continue
		}
		where := e.Where()
		if where == "" {
			where = "（运行期没给位置）"
		}
		return &Refusal{
			Code: output.CodeQueryFailed,
			Exit: 1,
			Message: fmt.Sprintf("作业程序停下了：%s —— 这一步之后这个窗口只会回空 ACK，"+
				"继续发事件都是白费（实测：发给容器会触发 FORMS -6327）", where),
			Hint: "重开会话（tt recording stop <录制ID> && tt recording start --launch standard --prog <作业>）再继续；" +
				"并在重放/计划里去掉那一步（它把值发给了容器，或发给了一个不属于当前交互指令的动作）",
		}
	}
	return nil
}

// isActionLike 判断类型是不是"点击类动作"（要受"必须属于当前交互指令"约束的那些）。
//
// 不列 `Button`：实测按钮是程序自己画的控件（如 #296 状态切换），它由程序管可用性，
// 不在 Dialog 的动作集里。`TopMenuCommand` 也不列（顶层菜单不属于任何 Dialog）。
func isActionLike(t string) bool {
	switch t {
	case "Action", "ActionDefault", "ToolBarItem", "MenuAction":
		return true
	}
	return false
}

// CheckActionInCurrentDialog 校验"这个动作是不是**当前交互指令**提供的"。
//
// 官方依据（1514）：动作由当前 Dialog 节点持有，交互元素按 name 绑定到 Action 节点 ——
// 所以"能不能发这个动作"= 它是否**属于当前那个 Dialog 实例**。实测：运行期每次状态变化
// 都重新分配动作 id（同一个"确定"：模板 #109 → 实例 #3535 → 查询后 #4533），
// 发到旧实例的动作只会回 18 字节空 ACK —— 看起来像"没反应"，其实是发错了对象。
//
// **判据不足时不拦**（模型里一个 active Dialog 都没有 = 这份录制没抓到 Dialog 帧，
// 此时任何判断都是猜）：只在"确实知道当前有哪些指令"的前提下才拒绝。
func CheckActionInCurrentDialog(m *Model, id int) error {
	acts := m.ActiveDialogs()
	if len(acts) == 0 {
		return nil // 判据不足：不猜，也不拦
	}
	d, ok := m.Get(id)
	if !ok {
		return nil // 由 CheckClick 去报"目标不存在"
	}
	if !isActionLike(d.Type) {
		return nil
	}
	owner, ok := NearestAncestor(m, id, "Dialog")
	if !ok {
		return nil // 顶层菜单这类不属于任何 Dialog，不受此约束
	}
	for _, a := range acts {
		if a.ID == owner.ID && a.Scope == owner.Scope {
			return nil
		}
	}
	return &Refusal{
		Code: output.CodeSessionState,
		Exit: 4,
		Message: fmt.Sprintf("动作 #%d（%s）不属于当前交互指令：它挂在 Dialog #%d(scope=%d) 上，"+
			"而当前的是 %s —— 发出去只会回空 ACK", id, d.Type, owner.ID, owner.Scope, activeDesc(acts)),
		Hint: "每步都**现读**状态再发（tt fep ui state 会给当前指令里的动作 id）；" +
			"运行期每次状态变化都会重新分配动作 id，缓存下来的 id 下一次就失效",
	}
}

// activeDesc 把当前交互指令列成一小段可读文本（诊断用）。
func activeDesc(acts []*Object) string {
	var parts []string
	for i, a := range acts {
		if i >= 3 {
			parts = append(parts, "…")
			break
		}
		parts = append(parts, fmt.Sprintf("#%d(scope=%d)", a.ID, a.Scope))
	}
	return "[" + strings.Join(parts, " ") + "]"
}
