package fep

import (
	"regexp"
	"fmt"
	"strconv"
	"strings"
)

// 回放步骤的**身份**：录制时记下 `(colName, fieldId)`，回放时**每步现解析**。
//
// 为什么非这样不可（契约 H1/H2）：`idRef` 是**会话句柄** —— 运行期会回收复用裸 id
// （同一个 4937 既做过下拉 Item 又做过 Dialog 的 Action，见 aui.go 里 `Object.Seq`
// 的实测注释），`rn` 之后更是全量失效。所以"把录制时的 id 再发一遍"在跨会话回放里
// 是**安静地点到别的控件上**（指纹闸拦得住编号表整体漂移，拦不住单点复用）。
// 稳定的东西只有编译期分配的 `(colName, fieldId)`，所以：
//
//	录制/计划期  → IdentityOf 把身份写进步骤（ReplayEvent.ColName/FieldID）
//	回放期       → ResolveStepTarget 按身份重新定位（ByColName），用完即弃
//
// 两条纪律写在签名与返回值里：
//
//  1. **旧数据零改动**：步骤没带 `ColName` ⇒ `ResolveStepTarget` 把事件原样返回
//     （`FromIdentity=false`），回放行为一字不变；
//  2. **判据只有硬事实**（A15）：窗口收窄只由调用方给的 `Scope.Window` 决定，
//     本层**不推**窗口名（`w_` + 剥 `_wf/_tui/_gui` 那种启发式只许进 Hint）；
//     录制侧也不按窗口过滤（A18③）。
//
// 这一层**没有副作用**：不读文件、不碰 socket、不发包；拒绝时事件体一个字都不改，
// 所以"调用方拿到 error 就不会 Send"这件事在纯函数层就能断言。

// ReplayIdentity 是"这一步身份"的两个键（与 ReplayEvent / ReplayStep 的同名字段一致）。
//
// 双键缺一不可寻址，所以 `ColName == ""` 就整体表示"未记录"；`FieldID == 0`
// 既是"未记录"，也可能是**合法的编号 0**（A10：实测 `{colName "b_index"} {fieldId "0"}`
// 就在真语料里）。这一处的取舍：JSON 用 `omitempty`（旧读者不许多看见两个空键），
// 于是编号 0 存不下来 —— 代价只是"回放时不再按编号收窄"，最坏结果是
// `FIELD_AMBIGUOUS`（如实拒），**不会**指错控件。
type ReplayIdentity struct {
	ColName ColName
	FieldID FieldID
}

// IdentityOf 从模型里读某个对象的身份（**纯读**，不发包）。
//
// 判据与 `ByColName` 同一口径：`colName` 与 `fieldId` 成对出现才算可寻址
// （`nodeFieldID`：读不出数字就当没有，绝不猜）。
func IdentityOf(m *Model, id int) ReplayIdentity {
	if m == nil || id == 0 {
		return ReplayIdentity{}
	}
	o, ok := m.Get(id)
	if !ok {
		return ReplayIdentity{}
	}
	col := strings.TrimSpace(o.Attr("colName"))
	if col == "" {
		return ReplayIdentity{}
	}
	out := ReplayIdentity{ColName: ColName(col)}
	if fid, ok := nodeFieldID(o); ok {
		out.FieldID = fid
	}
	return out
}

// StepTarget 是"这一步到底发给谁"的结论（`ResolveStepTarget` 的出口）。
type StepTarget struct {
	// Event 是要发的**事件文本**：走身份解析时其中的 `idRef` 已换成新解析到的 id。
	Event string
	// ID 是这一步指向的对象 id（`KeyEvent` 这类不带 idRef 的步骤为 0）。
	ID int
	// ColName / FieldID 是这一步**实际用的**身份（解析成功时是模型里那一份）。
	ColName ColName
	FieldID FieldID
	// FromIdentity = 这一步走了 `ByColName` 重新解析（H1/H2）。false = 旧路径。
	FromIdentity bool
}

// ResolveStepTarget 决定这一步发给谁：**有身份就按身份现解析，没有就照旧**。
//
// 返回 error 时（一律 `*Refusal`，四码之一）调用方**不得**发送：返回的 Event 是
// 原样的录制事件，一步都没换 —— 这正是"解析不过就不发包"可断言的地方。
//
// `sc` 只用来**收窄**（A18：`Scope.Window` 是 `m.WindowName` 的原样值）。回放默认
// 不给窗口（空 Scope）：录制基线不得按窗口过滤（A18③），跨窗口串号要靠身份本身去判，
// 不靠调用方猜出来的窗口名。
func ResolveStepTarget(m *Model, ev ReplayEvent, sc Scope) (StepTarget, error) {
	out := StepTarget{
		Event:   strings.TrimSpace(ev.Event),
		ColName: ColName(strings.TrimSpace(string(ev.ColName))),
		FieldID: FieldID(ev.FieldID),
	}
	if id, ok := idRefOf(out.Event); ok {
		out.ID = id
	}
	// 旧数据（没有身份）：一字不变 —— 既有回放语义不许被动到。
	if out.ColName == "" {
		return out, nil
	}
	// 有身份但事件不带 idRef（例如 KeyEvent）：身份无处可发，别把它当"没关系"。
	if out.ID == 0 {
		return out, refusal(CodeFieldNotInScope,
			fmt.Sprintf("这一步记着字段身份 %q，但事件里没有 idRef（%s）—— 没有目标可换",
				out.ColName, eventKind(out.Event)),
			"回放日志里这一步的 `colName` 与事件体不是一回事：把身份记到**带 idRef 的事件**上"+
				"（ActionEvent / ConfigureEvent）；按键这类步骤本来就没有目标对象")
	}

	// 期望编号只做**过滤**（C3.4：colName 命中而 fieldId 不符 ⇒ 判为另一张表单 ⇒ 拒）。
	use := sc
	if out.FieldID != 0 {
		fid := out.FieldID
		use.WantFieldID = &fid
	}
	r, err := ByColName(m, use, out.ColName, kindForReplayEvent(out.Event))
	if err != nil {
		return out, err
	}
	out.FromIdentity = true
	out.ID = r.ID
	out.ColName, out.FieldID = r.ColName, r.FieldID
	out.Event = setIDRef(out.Event, r.ID)
	return out, nil
}

// kindForReplayEvent 把事件文本映射成 `ByColName` 的动作类型（A9：它只影响可选性，
// 不淘汰候选）。KeyEvent 没有目标，给空 = 不限。
func kindForReplayEvent(event string) string {
	switch strings.TrimSpace(eventKind(event)) {
	case "ActionEvent":
		return "click"
	case "ConfigureEvent":
		// 只有"写值"才是输入动作；定位/排版（tabIndex/bufferSize/pageSize/cursor…）不限动作类型 ——
		// 它们本来就不是"点某个控件"，按输入去要求可选性会把网格列一并拒掉（真机现场：
		// imaa001 落在 TableColumn 上，FIELD_NOT_ACTIONABLE）。
		if configureWritesValue(event) {
			return "input"
		}
		return ""
	}
	return ""
}

// setIDRef 把事件文本里**第一处** `idRef "<n>"` 换成新 id（其余字节原样保留，包括
// 值、空格与属性顺序 —— 回放求保真，只许换目标）。
func setIDRef(event string, id int) string {
	const needle = `idRef "`
	i := strings.Index(event, needle)
	if i < 0 {
		return event
	}
	rest := event[i+len(needle):]
	j := strings.IndexByte(rest, '"')
	if j < 0 {
		return event
	}
	if _, _, ok := numAt(rest, 0); !ok {
		return event
	}
	return event[:i+len(needle)] + strconv.Itoa(id) + rest[j:]
}

// refusalReason 取一条拒绝的正文（回放日志里的 `reason`）。
//
// 只写 Message，不带 Hint：日志一行要能被界面轮询地读，长提示留在命令行的 stderr。
func refusalReason(err error) string {
	if ref, ok := err.(*Refusal); ok {
		return ref.Code + "：" + ref.Message
	}
	return fmt.Sprint(err)
}

// ── A23（契约 §7，用户决策 D-F1）：**实际发包**的回放强制身份 ────────────────
//
// 口径（写死在这里，两条发包入口都从这条缝过）：
//
//	步骤要发给某个对象（事件里有 idRef）却没有 `colName`
//	    → **拒**（`FIELD_NOT_IN_SCOPE` / 5，报文写明"需重录"），且发生在**任何发包之前**
//	步骤有 `colName`
//	    → 走 `ByColName` 现解析，失败按四码拒（5/5/1/4）
//	不发包的路径（`--dry-run` / 纯计划展示）
//	    → **放行**，但每条无身份的步骤必须标 `NoIdentityNote`
//
// 为什么必须拒（而不是"照旧按 idRef 发"）：裸 id 是**会话句柄**，跨会话会安静地指到
// 别的控件上（H1/H2 的实测现场）。身份闸是"这一轮的点错风险"唯一的硬判据。
//
// 为什么留 dry-run 这个例外：不发包的路径没有"点错对象"的风险；但它必须把后果写在
// 脸上（`NoIdentityNote`），否则人会以为"能展示就能跑"。
//
// **没有目标的步骤不在这一条之内**：`KeyEvent` 这类事件不带 idRef（`KeyEventText`），
// 它根本没有对象可指，装不上身份，重录也一样 —— 拒了它等于**每份真录制都不能回放**
// （语料里每条录制都有 Tab/Enter）。所以判据是"**要发给某个对象**却不知道是哪个"。

// NoIdentityNote 是"这一步没有身份"在不发包路径上的**唯一标注**（列表/命令行共用）。
//
// 括号里的后果是重点：展示得出来 ≠ 跑得动，这句话就是把差别写在脸上。
const NoIdentityNote = "无身份（真跑会被拒）"

// stepNeedsIdentity 报告这一步在**实际发包**时是不是必须能说出"发给哪个字段"。
//
// 判据只有一条硬事实：事件里有没有 `idRef`（要发给某个对象）。不看类型名、不猜。
func stepNeedsIdentity(ev ReplayEvent) bool {
	// `ConfigureEvent` 里**只有"写值"需要身份**：`{ value "X" }` 是真的在改某个字段的值；
	// cursor / currentRow / currentColumn / tabIndex / bufferSize / pageSize 是运行期定位与排版
	// （网格这类对象天然没有 colName）。把它们一起要求身份，结果是**每份真录制都回放不了**，
	// 而它们出错的表现是"某个控件排版/分页不对"，不是"点到了别人的按钮上"。
	// （网格的 bufferSize/pageSize 就是这样，这类对象天然没有 colName）。
	// 拿动作那一套身份纪律去要求它，结果是每份真录制都回放不了 —— 而它出错的表现是
	// "某个控件的排版/分页不对"，不是"点到了别人的按钮上"。
	if strings.HasPrefix(strings.TrimSpace(ev.Event), "ConfigureEvent") {
		return configureWritesValue(ev.Event)
	}
	_, ok := idRefOf(strings.TrimSpace(ev.Event))
	return ok
}

// RequireStepIdentity 判一步能不能在实际发包时走（A23）。
//
// 返回 *Refusal（四码之一，`FIELD_NOT_IN_SCOPE` / 5）时，调用方**不得发送**：
// 这一步没记下身份，发出去只能靠录制时的裸 id，而那个 id 在新会话里可能已经是别的控件。
func RequireStepIdentity(ev ReplayEvent) error {
	if !stepNeedsIdentity(ev) {
		// 没有目标对象（KeyEvent 之类）：身份无处可指，重录也修不好 —— 不拦。
		return nil
	}
	if strings.TrimSpace(string(ev.ColName)) != "" {
		return nil
	}
	id, _ := idRefOf(strings.TrimSpace(ev.Event))
	return refusal(CodeFieldNotInScope,
		fmt.Sprintf("第 %d 步没有字段身份（colName）：它要发给 #%d，而录制里没记下这到底是哪个字段"+
			" —— 这个对象不是表单字段（网格/按钮之类），录制时拿不到 colName（裸 id 跨会话会点到别的控件上）；"+
				"需重录一段只含字段操作的录制，或用 --skip 跳掉这一步", ev.N, id),
		"重录一份（tt recording start 录一轮新的，计划里就会带上 colName/fieldId）；"+
			"只想留个标记、不删任何文件：tt recording mark-stale <录制 ID> --reason \"…\"；"+
			"只是想看看要发什么（不发包）：tt recording replay-log --id <录制 ID> --dry-run")
}

// RequirePlanIdentity 是**发包之前**的那道闸（A23）：计划的每一条都过一遍。
//
// 返回第一处问题即止 —— 后面的步骤依赖前面的步骤，接着发只会连错；而且"哪一步不行"
// 本身就是排障要找的那一句。
//
// 空计划放行：没有要发的东西。
func RequirePlanIdentity(plan ReplayPlan) error {
	for _, ev := range plan.Events {
		if err := RequireStepIdentity(ev); err != nil {
			return err
		}
	}
	return nil
}

// NoIdentitySteps 列出计划里"真跑会被拒"的步骤号（A23 的展示侧要逐条标注）。
//
// 与 RequirePlanIdentity 用**同一个**判据（stepNeedsIdentity + ColName 空），
// 免得"标注说没事、真跑却被拒"。
func NoIdentitySteps(plan ReplayPlan) []int {
	var out []int
	for _, ev := range plan.Events {
		if !stepNeedsIdentity(ev) {
			continue
		}
		if strings.TrimSpace(string(ev.ColName)) == "" {
			out = append(out, ev.N)
		}
	}
	return out
}

// RunReplayIdentity 是**实际发包**的那条路：先过身份闸（A23），再走原来的回放。
//
// 两个真跑入口（`tt recording replay` 与 `tt recording replay`）都只调它 —— 留一条缝，
// 就不会出现"另一个入口忘了接"（E2 那次就是这么漏的）。
//
// `opt.DryRun` 时**不判**身份（A23 的例外：不发包就没有点错对象的风险），但仍然
// 一个字节都不发（那是 RunReplayWith 自己的 dry-run 分支）。
//
// 被拒时返回 (nil, *Refusal)：`res` 为 nil 是刻意的 —— "拒了"不是"跑完了"，
// 调用方不该拿到一份看起来正常的统计。
func (c *Client) RunReplayIdentity(plan ReplayPlan, opt ReplayRunOptions) (*ReplayResult, error) {
	if !opt.DryRun {
		if err := RequirePlanIdentity(plan); err != nil {
			return nil, err
		}
	}
	return c.RunReplayWith(plan, opt)
}

// configureWritesValue 判一条 ConfigureEvent 是不是在**写字段值**。
//
// 判据是事件里有没有 `{ value "…" }` 这一项：有就是"把某个字段设成这个值"（要身份），
// 没有就是定位/排版（cursor、currentRow、tabIndex、bufferSize…，重录也拿不到 colName）。
func configureWritesValue(ev string) bool {
	return configureValueAttr.MatchString(ev)
}

var configureValueAttr = regexp.MustCompile(`\{\s*value\s*"`)
