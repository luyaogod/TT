package fep

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

// `tt fep do ... --field <colName>`：按**字段身份**点名目标（契约 C3 + A7 + A18）。
//
// 这一层只回答一个问题：**这一步该发给谁，还是干脆别发**。
//
//	--field <colName> → fep.ByColName → `#<id>` → 交给 fep.Do 的既有发前校验
//	解析不过           → 四码 + Hint + 非 0 退出码，**一个字节都不发**
//
// 界线很清楚：**解析**在这一层（可单测的纯函数 `resolveFieldOrRefuse`），**校验与发送**
// 仍在 `fep.Do`（那边的 CheckClick / ResolveInputTarget / actionActive 既有规则一条不改）。
// 所以本文件没有任何 socket 代码，测试也不需要会话。
//
// 三个不许：
//   - 不许把 `--field` 与位置参数混着用（互斥，用法错 → 退出码 2，见 fep.go 的码表）；
//   - 不许用启发式补窗口名（A15：`w_` + 剥 `_wf/_tui/_gui` 只可进 Hint）；
//   - 不许把"读不到模型"说成"没这个字段"（前者 FORM_UNKNOWN/1，后者 FIELD_NOT_IN_SCOPE/5，
//     两种排障方向完全相反）。

// doField 是 `--field` 的值（字段工号 colName）。
var doField string

// fieldTarget 是 `resolveFieldOrRefuse` 的出口：**目标怎么写、以及它的身份**。
type fieldTarget struct {
	// Target 是交给 `fep.Do` 的目标写法（`#<id>`）—— 这样既有解析/校验路径全都照旧。
	Target string
	ID     int
	// ColName / FieldID 是这次实际命中的身份（写审计用；FieldID 是节点自己的编号，A10）。
	ColName fep.ColName
	FieldID fep.FieldID
}

// fieldKindFor 把动作类型映射成 `ByColName` 的 kind。
//
// 取值与 `fep.ByColName` 的两张白名单同名（`click`/`input`）。`key` 没有目标对象，
// 所以 `--field` 不挂在 `key` 上（挂了就是骗人）。
func fieldKindFor(kind string) string {
	switch kind {
	case "click":
		return "click"
	case "input", "pick":
		return "input"
	}
	return ""
}

// needsTarget 回答一个问题：**这个动作需不需要一个目标对象**（契约 A28）。
//
// 判据只认动作类型，不看调用方心情：`KeyEvent` 不带 `idRef`（A25 的判据同此），
// 所以 `key` 没有可指的对象；`act`（click）/ `type`（input）与它们的变体（pick）
// 都要点名。它与 fieldKindFor 共用一张表 —— 一个动作若不能按身份点名，它也就没有目标。
func needsTarget(kind string) bool {
	return fieldKindFor(kind) != ""
}

// checkFieldExclusive：**要目标的动作**里，`--field` 与位置参数（中文名 / `#id`）互斥。
//
// kind 是**必填参数**而不是调用点自己判断的条件：切片 E/J 的回归就是"在一个共同通路
// 上无差别套检查"，把 kind 摆进签名，写调用点时就必须说明这个动作是哪种 —— 漏判不再是
// 一次沉默，而是编不过（A28）。
//
// 为什么不做"两个都给时按某个优先"：那等于替人挑目标，而挑错的那一次会安静地
// 发到另一个控件上（契约 A9/A11 的同一条原则：拿不准就拒，并且说清怎么改）。
//
// 无目标动作（`key`）：**不检查目标**（它本来就没有），但 `--field` 明确**拒绝** ——
// 挂了这个开关却静默忽略，等于让人以为"按身份点名成功了"。
func checkFieldExclusive(kind, target, field string) error {
	if !needsTarget(kind) {
		if strings.TrimSpace(field) != "" {
			return fail(exitUsage,
				"这个动作（%s）没有目标对象，不能用 --field %q：KeyEvent 不带 idRef，没有可按身份点名的对象。"+
					"要么去掉 --field（`tt fep do %s <键名>`），要么改用 act / type 点名栏位",
				kind, field, kind)
		}
		return nil
	}
	hasTarget := strings.TrimSpace(target) != ""
	hasField := strings.TrimSpace(field) != ""
	switch {
	case hasTarget && hasField:
		return fail(exitUsage,
			"目标「%s」与 --field %q 互斥：要么给名字/#id，要么给 --field <colName>，二选一"+
				"（给了 --field 就别再写位置参数）", target, field)
	case !hasTarget && !hasField:
		return fail(exitUsage, "没给目标：写名字/#id，或给 --field <colName>")
	}
	return nil
}

// resolveFieldOrRefuse 是"解析 → 决定发不发"的**全部**判定（纯函数：不读会话、不读文件、
// 不发包；给它一个模型和一次请求，它只回"发给谁"或"为什么别发"）。
func resolveFieldOrRefuse(m *fep.Model, sc fep.Scope, field, kind string) (fieldTarget, error) {
	field = strings.TrimSpace(field)
	if field == "" {
		return fieldTarget{}, fail(exitUsage, "没给 --field：要按身份找人，就得说清是哪个 colName")
	}
	k := fieldKindFor(kind)
	if k == "" {
		return fieldTarget{}, fail(exitUsage,
			"这个动作（%s）不接受 --field：只有带目标的动作才能按身份点名", kind)
	}
	r, err := fep.ByColName(m, sc, fep.ColName(field), k)
	if err != nil {
		return fieldTarget{}, err
	}
	return fieldTarget{
		Target:  fmt.Sprintf("#%d", r.ID),
		ID:      r.ID,
		ColName: r.ColName,
		FieldID: r.FieldID,
	}, nil
}

// resolveDoField 有会话时的那一半：取当前界面模型，再交给纯函数。
//
// 模型从**录制件**重建（运行期不会回答"当前界面是什么"，它只会推 `om`）——
// 这与 `fep.Do` 自己是同一个来源、同一份文件，所以两处看到的是同一屏。
func resolveDoField(client *fep.Client, kind string) (fieldTarget, error) {
	if fieldKindFor(kind) == "" {
		return fieldTarget{}, fail(exitUsage,
			"这个动作（%s）不接受 --field：只有带目标的动作才能按身份点名", kind)
	}
	m, _, _, err := fep.LoadModelFrom(client.RecordDir, client.Tag)
	if err != nil || m == nil || m.Len() == 0 {
		// 读不到这一屏 = FORM_UNKNOWN（C4.2），**不是**"没这个字段"：
		// 前者要去看会话/录制件，后者要去核名字 —— 两条完全不同的排障路。
		return fieldTarget{}, &fep.Refusal{
			Code: fep.CodeFormUnknown, Exit: fep.ExitFormUnknown,
			Message: fmt.Sprintf("读不到当前界面模型（%v）：--field 要靠它把 colName 落成对象", err),
			Hint: "先确认会话在跑、录制件已落盘：`tt recording active`；" +
				"再看这一屏的树：`tt fep ui tree --window <窗口名片段>`（窗口名照 `tt fep ui snapshot` 出的原样值）",
		}
	}
	sc := fep.Scope{}
	// `--window` 只**收窄**（A18：它必须是 `m.WindowName` 的原样值，不是推出来的）。
	if w := strings.TrimSpace(doWindow); w != "" {
		sc.Window = w
	}
	return resolveFieldOrRefuse(m, sc, doField, kind)
}

// applyFieldFlag 把 `--field` 落到 `DoConfig.Target` 上；失败时打 Hint 并给非 0 退出码。
//
// 返回值是这次命中的 colName（写审计用；没给 `--field` 时为空）。
func applyFieldFlag(cmd *cobra.Command, dc *fep.DoConfig, client *fep.Client) (fep.ColName, error) {
	tgt, err := resolveDoField(client, dc.Kind)
	if err != nil {
		var ref *fep.Refusal
		if errors.As(err, &ref) {
			fmt.Fprintf(cmd.ErrOrStderr(), "· %s（退出码 %d）\n", ref.Code, ref.Exit)
			if ref.Hint != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "· 怎么办：%s\n", ref.Hint)
			}
			return "", fail(ref.Exit, "%s", ref.Message)
		}
		return "", err
	}
	dc.Target = tgt.Target
	fmt.Fprintf(cmd.ErrOrStderr(), "· --field %s → %s（colName=%s fieldId=%d；每步现解析，用完即弃）\n",
		strings.TrimSpace(doField), tgt.Target, tgt.ColName, tgt.FieldID)
	return tgt.ColName, nil
}

// targetOrFieldArgs 是 `do act` 的参数校验：**要么**一个位置参数，**要么** `--field`。
//
// 不用 cobra 自带的原因：它给的退出码是 1，而这条线的表里 2 = 用法错（fep.go:25-33）。
func targetOrFieldArgs(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(doField) != "" {
		if len(args) != 0 {
			return fail(exitUsage, "%s：--field 与位置参数互斥（收到了 %d 个位置参数：%s）",
				cmd.CommandPath(), len(args), strings.Join(args, " "))
		}
		return nil
	}
	if len(args) != 1 {
		return fail(exitUsage, "%s 需要 1 个目标参数（名字或 #id），或用 --field <colName>（收到 %d 个）",
			cmd.CommandPath(), len(args))
	}
	return nil
}

// valueAfterFieldArgs 是 `do type` 的参数校验：给了 `--field` 就只该有一个**值**参数。
func valueAfterFieldArgs(cmd *cobra.Command, args []string) error {
	if strings.TrimSpace(doField) != "" {
		if len(args) != 1 {
			return fail(exitUsage, "%s --field：除 --field 之外还要且只要 1 个值参数（收到 %d 个）",
				cmd.CommandPath(), len(args))
		}
		return nil
	}
	if len(args) != 2 {
		return fail(exitUsage, "%s 需要 2 个参数（目标 + 值），或 --field <colName> <值>（收到 %d 个）",
			cmd.CommandPath(), len(args))
	}
	return nil
}
