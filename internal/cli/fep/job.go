package fep

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

// tt fep job —— 把"一个作业怎么操作"从**源码规范**读出来，并按它闭环执行。
//
// 与 `tt fep do` 的分工：
//
//	do    一次一个动作，人（或上层脚本）决定下一步；
//	job   一次一条**语义步骤**序列：每步现读运行期状态 → 判可用性 → 发 → 验 → 记证据，
//	      任一步"作业停下"或"没有效果"就立刻停手并指名是哪一步。
//
// 为什么要有 job 这一层（用户原话："我们走一步卡一步"）：每次只试一个动作时，
// 前面 19 步的状态要靠人重建；而"作业怎么操作"本身**写在源码里**（.4fd 表单定义 +
// .4gl 的 DIALOG/ON ACTION），所以步骤应该是**推导**出来的、一次跑完的，
// 不是逐步试探出来的。

var (
	jobFormFlag   string
	jobSourceFlag string
	jobWindowFlag string
	jobDryRun     bool
	jobStopOnNoOp bool
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "作业级：从源码推导操作计划、按计划闭环执行（每步现读状态、判可用、验效果）",
	Long: `作业级操作 —— 面向"把一件单据做完"，不是面向"发一条事件"。

  tt fep job plan <作业程序>     从 .4fd/.4gl 推导操作计划（输入/选择/开窗/动作）
  tt fep job run <步骤文件>      按步骤闭环执行：现读状态 → 判可用 → 发 → 验 → 记证据

计划与执行的依据都来自**规范**（源码镜像），不是试探：
  - 顺序：.4fd 的 tabIndex（运行期 DIALOG 用 FIELD ORDER FORM 时就是这个序）
  - 操作：控件类型决定（Edit→输入 / ComboBox→选 / CheckBox→勾 / ButtonEdit→开窗）
  - 开窗语义：开窗程序 .4gl 的 ON ACTION accept（当前行 or 勾选）
  - 可用性：运行期广播的 active / hidden / noEntry / actionActive`,
}

var jobPlanCmd = &cobra.Command{
	Use:   "plan <作业程序，如 apmt500_wf>",
	Short: "从源码推导操作计划：每个可操作字段一步（含类型、必填、开窗、勾选字面值）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := jobSourceRoot(jobSourceFlag)
		if root == "" {
			return fail(exitUsage, "没给源码镜像根：加 --source <镜像根>，或配 tt dict mirror dir，或设 TT_SOURCE_ROOT")
		}
		src, err := fep.LoadSource(root, args[0])
		if err != nil {
			return fail(exitTarget, "%v", err)
		}
		byForm := map[string][]fep.SrcField{}
		for _, f := range src.Fields {
			byForm[f.Form] = append(byForm[f.Form], f)
		}
		forms := make([]string, 0, len(byForm))
		for k := range byForm {
			forms = append(forms, k)
		}
		sort.Strings(forms)

		type formPlan struct {
			Form  string         `json:"form"`
			Steps []fep.PlanStep `json:"steps"`
		}
		var out []formPlan
		for _, form := range forms {
			if jobFormFlag != "" && form != jobFormFlag {
				continue
			}
			steps := fep.BuildPlan(form, byForm[form])
			// 开窗那一步的选中语义从**开窗程序**读（不猜）
			for i := range steps {
				if steps[i].Opener == "" {
					continue
				}
				if ow := lookupOpenWinSoft(root, steps[i].Opener); ow != nil {
					steps = fep.PlanStepsFromOpenWin(steps, ow)
					break
				}
			}
			if len(steps) > 0 {
				out = append(out, formPlan{Form: form, Steps: steps})
			}
		}
		if len(out) == 0 {
			return fail(exitTarget, "%s：源码里没推导出可操作步骤（表单：%v）", args[0], src.Forms)
		}
		return printJSONOrText(out, func() {
			fmt.Printf("%s 的操作计划（源码：%s）\n", args[0], strings.Join(src.Forms, ", "))
			for _, fp := range out {
				fmt.Printf("\n【表单 %s】%d 步\n", fp.Form, len(fp.Steps))
				fmt.Print(fep.PlanText(fp.Steps))
				for _, s := range fp.Steps {
					if s.Op == "openwin" && s.Value != "" {
						fmt.Printf("     ↳ 开窗选中语义：%s（%s）\n", s.Value, s.Why)
					}
				}
			}
		})
	},
}

// lookupOpenWinSoft 读开窗源码；读不到就返回 nil（计划照出，只是少一条语义注解）。
func lookupOpenWinSoft(root, program string) *fep.OpenWinModel {
	ow, err := fep.LoadOpenWin(root, program)
	if err != nil {
		return nil
	}
	return ow
}

// ---------------------------------------------------------------------------
// job run：闭环执行
// ---------------------------------------------------------------------------

var jobRunCmd = &cobra.Command{
	Use:   "run <步骤文件>",
	Short: "闭环执行一串语义步骤（每步现读状态、判可用性、发、验效果，出错就停手）",
	Long: `步骤文件每行一步（# 开头是注释）：

  act  <角色>                 点一个动作（工具栏项 / 按钮 / 开窗按钮）
  set  <角色> = <值>          往栏位写值
  pick <角色> = <选项文本>    从选项里选一项
  key  <键名>                按一个键（Tab / Return …）
  cell <列角色> = <值>        往网格"当前行"的该列格子写值（值+光标同一条消息）
  check <列角色>              勾选"当前行"该列的勾选格（写该列 .4fd 的 valueChecked）
  expect <角色> = <值>        只核对，不发事件

每一行都会：**重新读运行期状态**（不缓存 id 与角色）→ 按 active/hidden/noEntry/actionActive
判这一步能不能做（不能做就停，并说明依据）→ 发 → 等回包 → 检查"值变了/窗口多了/作业是否停下"。
任一步：作业停下、或发送被预检拒绝、或（--stop-on-noop）没有任何效果 → **立刻停手**并指名是哪一步。

退出码：0 全部完成 / 4 某步不可用被拦 / 5 某步无效果 / 6 作业停下。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, err := os.ReadFile(args[0])
		if err != nil {
			return fail(exitUsage, "读不了步骤文件 %s：%v", args[0], err)
		}
		steps := parseJobScript(string(raw))
		if len(steps) == 0 {
			return fail(exitUsage, "步骤文件里没有有效步骤：%s", args[0])
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		if jobWindowFlag != "" {
			client.Window = jobWindowFlag
		}

		var done int
		for i, st := range steps {
			res, derr := runJobStep(client, st)
			if res != nil {
				printJobStep(i+1, len(steps), st, res)
			}
			if derr != nil {
				var ref *fep.Refusal
				if errors.As(derr, &ref) {
					return fail(ref.Exit, "第 %d 步（%s）被拦：%s", i+1, st.String(), ref.Message)
				}
				return fail(exitFailed, "第 %d 步（%s）失败：%v", i+1, st.String(), derr)
			}
			done++
			if res != nil && res.Changes != nil {
				if res.Changes.RuntimeError != "" {
					return fail(exitState, "第 %d 步（%s）之后**作业停下了**：%s",
						i+1, st.String(), res.Changes.RuntimeError)
				}
				if jobStopOnNoOp && res.Sent && len(res.Changes.Values) == 0 &&
					len(res.Changes.NewWindows) == 0 {
					return fail(exitFailed, "第 %d 步（%s）发出去了但运行期没有任何变化（不继续猜）",
						i+1, st.String())
				}
			}
		}
		fmt.Printf("\n完成 %d/%d 步。\n", done, len(steps))
		return nil
	},
}

// jobStep 是一行步骤。
type jobStep struct {
	Op     string
	Role   string
	Value  string
	Expect bool
	Line   int
	Raw    string
}

func (s jobStep) String() string {
	if s.Value == "" {
		return s.Op + " " + s.Role
	}
	return s.Op + " " + s.Role + " = " + s.Value
}

// parseJobScript 解析步骤文件：宽容但**不猜** —— 认不出的行直接报错行号。
func parseJobScript(text string) []jobStep {
	var out []jobStep
	for i, ln := range strings.Split(text, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		// 去掉行尾注释（要求 # 前有空白，避免吃掉值里的 #）
		if k := strings.Index(ln, " #"); k > 0 {
			ln = strings.TrimSpace(ln[:k])
		}
		fields := strings.SplitN(ln, " ", 2)
		op := strings.ToLower(strings.TrimSpace(fields[0]))
		rest := ""
		if len(fields) > 1 {
			rest = strings.TrimSpace(fields[1])
		}
		st := jobStep{Op: op, Line: i + 1, Raw: ln}
		if k := strings.Index(rest, "="); k >= 0 {
			st.Role = strings.TrimSpace(rest[:k])
			st.Value = strings.TrimSpace(rest[k+1:])
		} else if op == "key" {
			st.Value = rest
		} else {
			parts := strings.SplitN(rest, " ", 2)
			st.Role = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				st.Value = strings.TrimSpace(parts[1])
			}
		}
		out = append(out, st)
	}
	return out
}

// runJobStep 执行一步：现读状态 → 判可用 → 发 → 交给 fep.Do 等回包并汇总变化。
func runJobStep(client *fep.Client, st jobStep) (*fep.DoResult, error) {
	m, r, _, err := loadModel(uiFrom)
	if err != nil {
		return nil, fmt.Errorf("读运行期状态失败：%w", err)
	}
	switch st.Op {
	case "expect":
		return jobExpect(m, r, st)
	case "act":
		if err := jobGate(m, r, st, "click"); err != nil {
			return nil, err
		}
		return fep.Do(client, fep.DoConfig{Kind: "click", Target: st.Role, Conn: doConn, DryRun: jobDryRun, Wait: doWait})
	case "set", "type":
		if err := jobGate(m, r, st, "input"); err != nil {
			return nil, err
		}
		return fep.Do(client, fep.DoConfig{Kind: "input", Target: st.Role, Value: st.Value, Conn: doConn, DryRun: jobDryRun, Wait: doWait})
	case "pick":
		if err := jobGate(m, r, st, "input"); err != nil {
			return nil, err
		}
		return fep.Do(client, fep.DoConfig{Kind: "pick", Target: st.Role, Value: st.Value, Conn: doConn, DryRun: jobDryRun, Wait: doWait})
	case "key":
		return fep.Do(client, fep.DoConfig{Kind: "key", KeyName: st.Value, Conn: doConn, DryRun: jobDryRun, Wait: doWait})
	case "cell":
		ev, target, gerr := jobCellEvent(m, r, st)
		if gerr != nil {
			return nil, gerr
		}
		return fep.Do(client, fep.DoConfig{Kind: "raw", RawEvent: ev, Target: target, Conn: doConn, DryRun: jobDryRun, Wait: doWait})
	case "check":
		ev, target, gerr := jobCheckEvent(m, r, st)
		if gerr != nil {
			return nil, gerr
		}
		return fep.Do(client, fep.DoConfig{Kind: "raw", RawEvent: ev, Target: target, Conn: doConn, DryRun: jobDryRun, Wait: doWait})
	}
	return nil, &fep.Refusal{Exit: exitUsage, Message: fmt.Sprintf("第 %d 行：认不出的操作 %q（可用：act/set/pick/key/cell/check/expect）", st.Line, st.Op)}
}

// jobGate 是**发送前**的可用性闸门。
//
// 宁可不判不猜：只有在运行期**明确标了不可用**、或源码明确标了不可录入时才拦。
// 拦下时把依据一起报出来（用户要知道"为什么现在不能动它"）。
func jobGate(m *fep.Model, r *fep.Resolver, st jobStep, mode string) error {
	cat := fep.Catalog(m, r)
	c, cands, err := fep.ResolveRole(cat, st.Role)
	if err != nil {
		hint := ""
		if len(cands) > 0 {
			names := make([]string, 0, len(cands))
			for _, x := range cands {
				names = append(names, fmt.Sprintf("#%d(%s)", x.ID, x.Kind))
			}
			hint = "候选：" + strings.Join(names, ", ")
		}
		return &fep.Refusal{Exit: exitTarget, Message: fmt.Sprintf("角色 %q 解析不出来：%v %s", st.Role, err, hint)}
	}
	// 类型先判：不支持的组合直接拒（"控件类型 → 支持的操作"）
	ok := false
	for _, op := range c.Ops {
		if op == mode || (mode == "click" && op == "invoke") || (mode == "input" && op == "set") {
			ok = true
			break
		}
	}
	if !ok {
		return &fep.Refusal{Exit: exitTarget, Message: fmt.Sprintf(
			"%s（#%d %s）不支持「%s」操作，它支持：%s", c.Role, c.ID, c.Kind, mode, strings.Join(c.Ops, "/"))}
	}
	el := fep.EligibleControl(m, c)
	if !el.Usable {
		return &fep.Refusal{Exit: exitState, Message: fmt.Sprintf(
			"%s（#%d %s）此刻不可用：%s", c.Role, c.ID, c.Kind, el)}
	}
	// 焦点闸门：**字段级动作**（形如 s_detail1.pmdn001.controlp）只在
	// "光标确实在那个栏位"时才有效 —— 焦点在别处时点下去是**空动作**
	// （不报错、不生效），这正是反复出现"点对了动作却没反应"的真因。
	// 依据：运行期自己推 un 0 {{focus "…"}}；内部早就为回放这条线写下过这个教训
	// （internal/fep/replay.go:21），驱动这条线现在补上。
	if mode == "click" {
		if err := jobFocusGate(m, c); err != nil {
			return err
		}
	}
	return nil
}

// jobFocusGate 拦"焦点不在目标栏位上"的字段级动作。
//
// 只对**带 .controlp 的动作**生效（那是"对本栏位开窗"）：这类动作的语义就是
// "把当前光标所在栏位拿去开窗"，焦点不在那儿就必然没意义。
// 工具栏动作（录入/确定）不在此列 —— 它们不依赖光标在哪个栏位。
//
// 判据不足（运行期没推 focus、也没 cursor）时不拦（与全局原则一致）。
func jobFocusGate(m *fep.Model, c fep.Control) error {
	part := fieldPartOfControlp(c.Role)
	if part == "" {
		return nil
	}
	rep := m.FocusNow()
	focusID := rep.CursorID
	if focusID == 0 {
		focusID = rep.FocusID
	}
	if focusID == 0 {
		return nil // 判据不足
	}
	if focusMatches(m, focusID, part) {
		return nil
	}
	who := fmt.Sprintf("#%d", focusID)
	if o, ok := m.Get(focusID); ok {
		if n := o.Attr("name"); n != "" {
			who += " " + n
		}
		if cm := o.Attr("comment"); cm != "" {
			who += " (" + cm + ")"
		}
	}
	return &fep.Refusal{Exit: exitState, Message: fmt.Sprintf(
		"焦点不在 %s 上，现在在 %s —— 这是字段级开窗动作（%s），焦点不在目标栏位时点下去是空动作。"+
			"先把光标移到 %s 那一格再发。", part, who, c.Role, part)}
}

// fieldPartOfControlp 从 "s_detail1.pmdn001.controlp" 里取出 "pmdn001"。
func fieldPartOfControlp(role string) string {
	if !strings.HasSuffix(role, ".controlp") {
		return ""
	}
	base := strings.TrimSuffix(role, ".controlp")
	if i := strings.LastIndex(base, "."); i >= 0 {
		return base[i+1:]
	}
	return ""
}

// focusMatches 判"持有焦点的这个对象是不是就是那个栏位（或它的子孙）"。
func focusMatches(m *fep.Model, focusID int, fieldName string) bool {
	if fieldName == "" || focusID == 0 {
		return false
	}
	cur, ok := m.Get(focusID)
	for hop := 0; ok && cur != nil && hop < 8; hop++ {
		n := cur.Attr("name")
		if n != "" && (strings.HasSuffix(n, "."+fieldName) || n == fieldName ||
			strings.HasSuffix(n, "."+fieldName+"_1")) {
			return true
		}
		if strings.HasSuffix(cur.Attr("comment"), fieldName) {
			return true
		}
		if cur.Parent == 0 {
			break
		}
		cur, ok = m.Get(cur.Parent)
	}
	return false
}

// jobExpect 只核对当前值（不发事件）—— 用来在链路中途确认"上一步真的生效了"。
func jobExpect(m *fep.Model, r *fep.Resolver, st jobStep) (*fep.DoResult, error) {
	cat := fep.Catalog(m, r)
	c, _, err := fep.ResolveRole(cat, st.Role)
	if err != nil {
		return nil, &fep.Refusal{Exit: exitTarget, Message: fmt.Sprintf("角色 %q 解析不出来：%v", st.Role, err)}
	}
	got := c.Value
	res := &fep.DoResult{Target: fmt.Sprintf("#%d", c.ID), Name: c.Role, Sent: false,
		Changes: &fep.Changes{Values: []fep.ValueChange{{ID: c.ID, Name: c.Role, To: got}}}}
	if !strings.EqualFold(strings.TrimSpace(got), strings.TrimSpace(st.Value)) {
		return res, &fep.Refusal{Exit: exitFailed, Message: fmt.Sprintf(
			"%s 当前值 = %q，期望 %q", c.Role, got, st.Value)}
	}
	return res, nil
}

// jobCellEvent 往网格"当前行"的某列格子写值。
//
// 形态照前端（实测）：**值与光标在同一条消息**里发，拆开运行期不认。
// 目标格子取**最近被触碰**那一格（Seq 最大）—— 同一列下会残留旧实例的格子，
// 按 id 最小去挑会挑到死格子（空 ACK，严重时把作业打停）。
func jobCellEvent(m *fep.Model, r *fep.Resolver, st jobStep) (string, string, error) {
	col, cerr := jobColumn(m, r, st)
	if cerr != nil {
		return "", "", cerr
	}
	cells := fep.CellsIn(m, col.ID, 0)
	if len(cells) == 0 {
		return "", "", &fep.Refusal{Exit: exitTarget, Message: fmt.Sprintf(
			"列 %s（#%d）下没有格子（这一行可能还不存在）", col.Role, col.ID)}
	}
	cell := cells[len(cells)-1].ID // Seq 最大 = 当前行
	ev := fep.ConfigureValueText(cell, st.Value, nil) + "}{" + fep.ConfigureCursorText(col.ID, len([]rune(st.Value)))
	return ev, fmt.Sprintf("#%d", col.ID), nil
}

// jobCheckEvent 勾选网格"当前行"的某列勾选格。
//
// 这是**按真机语料定的形状**（不是猜）：
//
//	ConfigureEvent{idRef <网格>}{currentRow "r"}{currentColumn "c"}   ← 先把光标移到那一行那一列
//	ConfigureEvent{idRef <该行该列的格子>}{ value "<valueChecked>" } ← 再写勾选字面值
//
// 出处：tools/gdc-uia/README.md「真机语料里"勾选"的原始字节」。
// 字面值取自该列 .4fd 的 CheckBox valueChecked；取不到就**拒发**（不许猜）。
func jobCheckEvent(m *fep.Model, r *fep.Resolver, st jobStep) (string, string, error) {
	col, cerr := jobColumn(m, r, st)
	if cerr != nil {
		return "", "", cerr
	}
	literal := ""
	if col.Source != nil {
		literal = col.Source.ValueChecked
	}
	if literal == "" {
		// 兜底：从该列的 Value 格子当前值反推"哪个是勾上的值"是猜 —— 拒发。
		return "", "", &fep.Refusal{Exit: exitTarget, Message: fmt.Sprintf(
			"列 %s 的源码里没有 valueChecked（勾选字面值）——按规矩不猜，先确认这列是不是勾选列", col.Role)}
	}
	cells := fep.CellsIn(m, col.ID, 0)
	if len(cells) == 0 {
		return "", "", &fep.Refusal{Exit: exitTarget, Message: fmt.Sprintf(
			"列 %s（#%d）下没有格子（这一行可能还不存在）", col.Role, col.ID)}
	}
	cell := cells[len(cells)-1]
	grid := fep.GridOfColumn(m, col.ID)
	row := -1
	if g, ok := m.Get(grid); ok {
		if v, err := strconv.Atoi(g.Attr("currentRow")); err == nil {
			row = v
		}
	}
	// 行号取运行期当前行（它自己知道）；拿不到就不发 currentRow，只写格子。
	prefix := ""
	if row >= 0 {
		prefix = fmt.Sprintf("{ConfigureEvent 0{{idRef \"%d\"}{ currentRow \"%d\"}}}", grid, row)
	}
	ev := prefix + "{" + fep.ConfigureValueText(cell.ID, literal, nil) + "}"
	return ev, fmt.Sprintf("#%d", cell.ID), nil
}

// jobColumn 把角色解析成**网格列**（写格子/勾选都只认 TableColumn）。
func jobColumn(m *fep.Model, r *fep.Resolver, st jobStep) (fep.Control, error) {
	cat := fep.Catalog(m, r)
	if strings.HasPrefix(st.Role, "#") {
		id, err := strconv.Atoi(strings.TrimPrefix(st.Role, "#"))
		if err != nil {
			return fep.Control{}, &fep.Refusal{Exit: exitUsage, Message: "列 id 写法应为 #数字：" + st.Role}
		}
		for _, c := range cat {
			if c.ID == id && c.Kind == "TableColumn" {
				return c, nil
			}
		}
		return fep.Control{}, &fep.Refusal{Exit: exitTarget, Message: fmt.Sprintf("#%d 不是网格列", id)}
	}
	c, cands, err := fep.ResolveRole(fep.FilterKind(cat, "TableColumn"), st.Role)
	if err != nil {
		return fep.Control{}, &fep.Refusal{Exit: exitTarget, Message: fmt.Sprintf(
			"列 %q 解析不出来：%v（候选 %d 个，给更完整的「窗口/角色」或 #id）", st.Role, err, len(cands))}
	}
	return c, nil
}

func printJobStep(i, total int, st jobStep, res *fep.DoResult) {
	mark := "·"
	if res.Sent {
		mark = "→"
	}
	fmt.Printf("[%d/%d] %s %s", i, total, mark, st.String())
	if res.Target != "" {
		fmt.Printf("  （%s", res.Target)
		if res.Name != "" {
			fmt.Printf(" %s", res.Name)
		}
		fmt.Printf("）")
	}
	if res.Refused != "" {
		fmt.Printf("  被拒：%s", res.Refused)
	}
	if res.Changes != nil {
		if res.Changes.RuntimeError != "" {
			fmt.Printf("  ⚠ 作业停下：%s", res.Changes.RuntimeError)
		} else {
			fmt.Printf("  运行期 +%d 字节 / %d 条", res.Changes.RuntimeBytes, res.Changes.NewMessages)
			for _, v := range res.Changes.Values {
				fmt.Printf(" / %s=%q", orDash(v.Name), v.To)
			}
			if len(res.Changes.NewWindows) > 0 {
				fmt.Printf(" / 新窗口 %s", strings.Join(res.Changes.NewWindows, ","))
			}
		}
	}
	fmt.Println()
}

func init() {
	jobPlanCmd.Flags().StringVar(&jobFormFlag, "form", "", "只看某个表单（如 apmt500_wf）")
	jobPlanCmd.Flags().StringVar(&jobSourceFlag, "source", "", "源码镜像根")
	jobRunCmd.Flags().StringVar(&jobWindowFlag, "window", "", "把事件路由到某个窗口（默认当前窗口）")
	jobRunCmd.Flags().BoolVar(&jobDryRun, "dry-run", false, "只校验与打印，一个字节都不发")
	jobRunCmd.Flags().BoolVar(&jobStopOnNoOp, "stop-on-noop", true, "某步发出去了但运行期毫无变化就停手")
	jobRunCmd.Flags().DurationVar(&doWait, "wait", 0, "每步等运行期回包多久（0 = 默认）")
	jobRunCmd.Flags().IntVar(&doConn, "conn", 0, "指定连接（默认自动）")
	jobCmd.AddCommand(jobPlanCmd, jobRunCmd)
}

// jobSourceRoot 解析源码镜像根：--source > 环境变量 > 配置的 mirror.dir。
// 与 src 命令共用同一套优先级，避免两个命令对"源码在哪"给出不同答案。
func jobSourceRoot(flag string) string {
	if flag != "" {
		return flag
	}
	return srcOpenWinSource()
}
