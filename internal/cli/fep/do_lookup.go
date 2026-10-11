package fep

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"tt/internal/fep"

	"github.com/spf13/cobra"
)

// tt fep do lookup —— **开窗取数**：一条命令做完整套"点开窗 → 查 → 选中 → 回传"。
//
// 为什么非要把它做成一条命令（而不是让人手搓几条）：开窗的**每一步都会换掉对象 id**。
// 实测同一个开窗里："查询"按钮的 accept 是 #153 → 查询后变成 #4538；窗口实例从
// #4533 变成 #4552；勾选格从 #168 变成别的。手搓命令时你发的是**上一步读到的 id**，
// 运行期对旧 id 只回一个**空 ACK**（`om N {}`）—— 看起来像"这一步没生效"，
// 于是继续改参数继续试，每一步都在浪费（今天为此耗掉一整天）。
//
// 所以这个命令的骨架是：**每步都重新折叠一次 AUI、重新决定发什么**（read → decide → send → verify），
// 而不是先算好一串 id 再依次发出去。
//
// 语义不靠猜，靠两处权威：
//   - **源码**（tt fep src openwin <程序>）：accept 段读出来的"选中"语义 ——
//     `DISPLAY ARRAY` 那个对话框用 `li_ac`（当前行），`INPUT ARRAY` 那个用 `check`（勾选）；
//   - **运行期**：网格当前处在哪个阶段（`dialogType` = Construct / InputArray / DisplayArray）。
//
// 运行期说 InputArray 而源码说是勾选 → 就给目标行的勾选格写 `valueChecked`（来自 .4fd 的
// `valueChecked="Y"`），再把光标移到那一行，最后点确定。

var (
	lookupQuery  string
	lookupRow    int
	lookupWait   time.Duration
	lookupCheck  string
	lookupSource string
	lookupAssume bool
)

// lookupPlan 是这个命令每一步要做的事（先算出计划、再逐步执行、每步验证）。
type lookupPlan struct {
	Opener   string `json:"opener"` // 开窗按钮（角色或 #id）
	Window   int    `json:"window"` // 开窗实例的窗口 id
	WindowNm string `json:"windowName,omitempty"`
	Program  string `json:"openwinProgram,omitempty"` // 从窗口名推出的开窗程序
	Grid     int    `json:"grid"`                     // 它的结果网格
	Stage    string `json:"stage"`                    // 运行期当前阶段（dialogType）
	Select   string `json:"select"`                   // 选中语义（源码说的，或显式接受推断时用的）
	Why      string `json:"selectWhy,omitempty"`      // 这个结论的出处 / 或"判不出来"的原因
	Row      int    `json:"row"`                      // 目标行（1 起）
	CheckCol string `json:"checkCol,omitempty"`
	CheckVal string `json:"checkVal,omitempty"`
	Cell     int    `json:"cell,omitempty"` // 要写值的那一格
	Accept   int    `json:"accept"`         // 当前交互指令里的确定
	Result   string `json:"result,omitempty"`
}

var doLookupCmd = &cobra.Command{
	Use:   "lookup <开窗按钮：角色名或 #id>",
	Short: "开窗取数：点开窗 → 写条件/查询 → 选中目标行（按源码的语义）→ 确定回传",
	Long: `一条命令做完整套开窗取数。内部**每步现读 AUI**（id 每步都会变，缓存下来的必失效）。

选中语义从源码读，不猜：
  - 运行期 dialogType=DisplayArray → 该对话框的 accept 用行下标（当前行）→ 设 currentRow 再确定
  - 运行期 dialogType=InputArray  → 该对话框的 accept 用 check（勾选）→ 给那行的勾选格写 Y
  - 运行期 dialogType=Construct   → 查询阶段：写条件列 → 确定执行查询

开关：
  --query <值>   查询条件（写到构造列的第一列）
  --row N        目标行（1 起；0 = 不选行，只查询）
  --check <值>   勾选要写的值（缺省取 .4fd 的 valueChecked，通常是 Y）
  --source <根>  源码镜像根（读选中语义用；缺省用配置的 mirror.dir）
  --wait 12s     每一步等运行期回包多久`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, cerr := newClient()
		if cerr != nil {
			return cerr
		}
		if _, err := c.Status(); err != nil {
			return fail(exitNoSession, "%v", err)
		}
		// 注入口要求显式连接号（实测：传 0 会被拒 "no such connection: 0"）。
		// 选"最后活着的那条"：一个作业实例一条连接，开窗不新开连接。
		connID := 0
		if conns, cerr := c.Status(); cerr == nil {
			for _, cc := range conns {
				if cc.Alive && cc.ID > connID {
					connID = cc.ID
				}
			}
		}
		if connID == 0 {
			return fail(exitNoSession, "会话里没有活着的连接（作业还没连上来？）")
		}
		pl := &lookupPlan{Opener: args[0], Row: lookupRow}
		step := func(format string, a ...any) {
			fmt.Printf("  · "+format+"\n", a...)
		}
		send := func(ev string) error {
			if _, err := c.Send(connID, ev); err != nil {
				return err
			}
			time.Sleep(400 * time.Millisecond)
			return nil
		}
		reload := func() (*fep.Model, error) {
			m, _, _, err := loadModel("")
			if err != nil {
				return nil, err
			}
			if errs := fep.RuntimeErrors(m); len(errs) > 0 {
				for _, e := range errs {
					if e.ProgramStop() {
						return m, fmt.Errorf("作业程序停下了：%s", e.Where())
					}
				}
			}
			return m, nil
		}

		m, err := reload()
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		before := windowSet(m)

		// ① 点开窗。明细里的开窗按钮要求光标先落在那行那列（否则 actionActive=0）。
		// 参数可以是角色名，也可以直接是 #id（写死 id 不推荐，但调试时要能用）。
		id, name := 0, ""
		if strings.HasPrefix(pl.Opener, "#") {
			if _, err := fmt.Sscanf(pl.Opener, "#%d", &id); err != nil || id == 0 {
				return fail(exitUsage, "看不懂的目标 %q（要角色名或 #id）", pl.Opener)
			}
			o, ok := m.Get(id)
			if !ok {
				return fail(exitTarget, "AUI 里没有 #%d", id)
			}
			name = o.Attr("name")
		} else {
			ctrl, _, rerr := fep.ResolveRole(mustCatalog(m), pl.Opener)
			if rerr != nil {
				return fail(exitTarget, "找不到开窗按钮 %q：%v", pl.Opener, rerr)
			}
			id, name = ctrl.ID, ctrl.Role
		}
		pl.Opener = fmt.Sprintf("#%d(%s)", id, name)
		// 发前校验（R1）：不在当前交互指令里的动作、actionActive=0 的按钮，**发了也是空 ACK**。
		// 先在这里说清楚为什么不行，比发出去再看"没反应"好得多。
		if cerr := fep.CheckClick(m, id); cerr != nil {
			return fail(exitState, "开窗按钮 #%d 现在不能点：%v", id, cerr)
		}
		if grid := fep.GridOfColumn(m, id); grid != 0 {
			step("先把光标放进网格 #%d 第 %d 行再点开窗", grid, maxi(pl.Row, 1))
			_ = send(fep.ConfigureGridCursorText(grid, maxi(pl.Row, 1), 2))
		}
		if err := send(fep.ActionEventText(id)); err != nil {
			return fail(exitFailed, "点开窗：%v", err)
		}
		time.Sleep(lookupWait)

		// ② 等新窗口出现（判据 = AUI 多了一个 Window，**不是**连接数）
		m, err = reload()
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		win := newWindow(m, before)
		if win == 0 {
			return fail(exitTarget, "点开窗后没有新窗口：%v", strings.Join(windowNames(m), ", "))
		}
		pl.Window = win
		pl.WindowNm = m.WindowName(win)
		pl.Program = fep.OpenWinProgramFromWindow(pl.WindowNm)
		step("开窗已出现：窗口 #%d %s（开窗程序：%s）", win, pl.WindowNm, orDash(pl.Program))

		// ③ 循环推进：查询 → 选中 → 回传。每轮都**重新读状态**再决定。
		for round := 0; round < 6; round++ {
			m, err = reload()
			if err != nil {
				return fail(exitFailed, "%v", err)
			}
			if cur := m.CurrentWindow(); cur != win && !windowAlive(m, win) {
				pl.Result = "开窗已关闭（回传完成）"
				break
			}
			grid, stage := gridOfWindow(m, win)
			pl.Grid, pl.Stage = grid, stage
			// 选中语义：**先看源码，再看运行期阶段**，两边对不上就停手（R2）。
			// 没有源码时不用"InputArray 就是勾选"这种推断顶替 —— 除非调用方显式接受。
			mode, why := fep.ReconcileSelect(lookupOpenWin(pl.Program), stage)
			if mode == fep.SelectUnknown && lookupAssume {
				if alt := fep.StageDefaultSelect(stage); alt != fep.SelectUnknown {
					mode = alt
					why = "【推断·非源码】" + why + " → 按阶段默认取 " + string(alt)
				}
			}
			pl.Select, pl.Why = string(mode), why
			step("%s", why)
			if stage == "InputArray" || stage == "DisplayArray" {
				if mode == fep.SelectUnknown {
					return fail(exitState, "选中语义判不出来（%s）—— 不猜：给 --source 指到源码镜像，"+
						"或用 --assume 显式接受按阶段的默认推断", why)
				}
				if want := fep.StageDefaultSelect(stage); want != mode {
					step("⚠ 注意：这个阶段的常见语义是 %s，而源码说的是 %s —— 按**源码**走", want, mode)
				}
			}
			acc, ok := fep.ActionNameInDialog(m, "accept")
			pl.Accept = acc
			if !ok {
				return fail(exitState, "当前交互指令里没有 accept（窗口 #%d，阶段 %s）：宁缺不猜", win, stage)
			}
			switch stage {
			case "Construct":
				if lookupQuery != "" {
					col, cell, err := constructTarget(m, win)
					if err != nil {
						return fail(exitTarget, "查询条件列：%v", err)
					}
					step("写查询条件 %q → 列 #%d 的格子 #%d", lookupQuery, col, cell)
					_ = send(fep.ConfigureGridCursorText(grid, 1, 1))
					_ = send(fep.ConfigureValueText(cell, lookupQuery, nil))
				}
				step("执行查询（确定 #%d）", acc)
				if err := send(fep.ActionEventText(acc)); err != nil {
					return fail(exitFailed, "%v", err)
				}
				time.Sleep(lookupWait)
			case "InputArray":
				// 源码说这个对话框是"勾选式"：给目标行的勾选格写 valueChecked。
				if pl.Row <= 0 {
					pl.Result = "只要查询，不选行（--row 0）"
					break
				}
				cell, col, val, err := checkTarget(m, win, grid, pl.Row)
				if err != nil {
					return fail(exitTarget, "勾选格：%v", err)
				}
				pl.Cell, pl.CheckCol, pl.CheckVal = cell, fmt.Sprintf("#%d", col), val
				step("阶段=InputArray（源码：勾选式）→ 光标到第 %d 行、第 1 列，给勾选格 #%d 写 %q", pl.Row, cell, val)
				_ = send(fep.ConfigureGridCursorText(grid, pl.Row, 1))
				_ = send(fep.ConfigureValueText(cell, val, nil))
				time.Sleep(1 * time.Second)
				m2, err := reload()
				if err != nil {
					return fail(exitFailed, "%v", err)
				}
				acc2, _ := fep.ActionNameInDialog(m2, "accept")
				step("确定（#%d）", acc2)
				if err := send(fep.ActionEventText(acc2)); err != nil {
					return fail(exitFailed, "%v", err)
				}
				time.Sleep(lookupWait)
			case "DisplayArray":
				// 源码说这个对话框是"当前行式"：把光标移到目标行再确定。
				if pl.Row > 0 {
					step("阶段=DisplayArray（源码：当前行式）→ currentRow=%d", pl.Row)
					_ = send(fep.ConfigureGridRowText(grid, pl.Row))
				}
				acc2 := acc
				if m2, err := reload(); err == nil {
					if a, ok := fep.ActionNameInDialog(m2, "accept"); ok {
						acc2 = a
					}
				}
				step("确定（#%d）", acc2)
				if err := send(fep.ActionEventText(acc2)); err != nil {
					return fail(exitFailed, "%v", err)
				}
				time.Sleep(lookupWait)
			default:
				// 阶段读不出来：宁可停手（不发事件），也不要猜着发。
				return fail(exitTarget, "窗口 #%d 的网格阶段读不出来（dialogType=%q）：不猜，停手", win, stage)
			}
		}
		m, _ = reload()
		if cur := m.CurrentWindow(); cur != win {
			pl.Result = fmt.Sprintf("开窗已关闭，当前窗口回到 #%d %s", cur, m.WindowName(cur))
		}
		appendLookupAudit(pl)
		return printJSONOrText(pl, func() {
			fmt.Printf("开窗取数：%s\n", pl.Opener)
			fmt.Printf("  窗口 #%d %s  开窗程序 %s\n", pl.Window, pl.WindowNm, orDash(pl.Program))
			fmt.Printf("  网格 #%d  阶段 %s  选中语义 %s  行 %d\n", pl.Grid, pl.Stage, pl.Select, pl.Row)
			if pl.Why != "" {
				fmt.Printf("  依据：%s\n", pl.Why)
			}
			if pl.Cell != 0 {
				fmt.Printf("  勾选：%s ← %s（格子 #%d）\n", pl.CheckCol, pl.CheckVal, pl.Cell)
			}
			fmt.Printf("  结果：%s\n", pl.Result)
		})
	},
}

// ---- helpers ----

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func orDash(s string) string {
	if s == "" {
		return "（从窗口名推不出）"
	}
	return s
}

// lookupOpenWin 读开窗程序源码（读不到就回 nil —— 上层据此说"没有源码可对"）。
func lookupOpenWin(program string) *fep.OpenWinModel {
	if program == "" {
		return nil
	}
	root := lookupSource
	if root == "" {
		root = srcOpenWinSource()
	}
	if root == "" {
		return nil
	}
	ow, err := fep.LoadOpenWin(root, program)
	if err != nil {
		return nil
	}
	return ow
}

func mustCatalog(m *fep.Model) []fep.Control { return fep.Catalog(m, fep.NewResolver(m)) }

func windowSet(m *fep.Model) map[int]bool {
	out := map[int]bool{}
	for _, w := range m.Windows() {
		out[w.ID] = true
	}
	return out
}

func newWindow(m *fep.Model, before map[int]bool) int {
	best := 0
	for _, w := range m.Windows() {
		if before[w.ID] {
			continue
		}
		if w.ID > best {
			best = w.ID
		}
	}
	return best
}

func windowAlive(m *fep.Model, win int) bool {
	_, ok := m.Get(win)
	return ok
}

func windowNames(m *fep.Model) []string {
	var out []string
	for _, w := range m.Windows() {
		out = append(out, fmt.Sprintf("#%d %s", w.ID, w.Attr("name")))
	}
	return out
}

// gridOfWindow 返回窗口自己的那个结果网格（实测：网格与窗口在同一条消息里声明）
// 以及它的 dialogType（运行期说的"这一步在哪个阶段"）。
func gridOfWindow(m *fep.Model, win int) (int, string) {
	for _, g := range m.TablesOfWindow(win) {
		if o, ok := m.Get(g); ok {
			return g, o.Attr("dialogType")
		}
	}
	return 0, ""
}

// constructTarget 找 CONSTRUCT 阶段该往哪个格子写条件：该窗口网格里第一列真实列的第一个格子。
func constructTarget(m *fep.Model, win int) (col, cell int, err error) {
	grid, _ := gridOfWindow(m, win)
	if grid == 0 {
		return 0, 0, fmt.Errorf("窗口 #%d 没有网格", win)
	}
	for _, o := range m.Objects() {
		if o.Type != "TableColumn" || o.Parent != grid {
			continue
		}
		if strings.HasSuffix(o.Attr("name"), ".check") || o.Attr("name") == "formonly.check" {
			continue
		}
		if c, cerr := fep.FirstCellOfIn(m, o.ID, win); cerr == nil {
			return o.ID, c, nil
		}
	}
	return 0, 0, fmt.Errorf("窗口 #%d 的网格 #%d 下没找到可写条件列", win, grid)
}

// checkTarget 找"目标行的勾选格"。
//
// 两条实测规则：
//  1. **按声明顺序（Seq）**数行 —— 运行期复用裸 id，按 id 取到的往往是**旧实例**的格子
//     （实测 #59/#105 属于上一个实例；当前实例是 #168/#3549/#3548…）；
//  2. 勾选格的当前值是 `valueUnchecked`（.4fd 的 `valueUnchecked="N"`），
//     所以"值是 N 的格子"就是本页未勾选的那些。
func checkTarget(m *fep.Model, win, grid, row int) (cell, col int, val string, err error) {
	// 勾选列：本实例网格下名字以 check 结尾的那一列
	var checkCol *fep.Object
	for _, o := range m.Objects() {
		if o.Type == "TableColumn" && o.Parent == grid && strings.HasSuffix(o.Attr("name"), "check") {
			checkCol = o
		}
	}
	if checkCol == nil {
		return 0, 0, "", fmt.Errorf("网格 #%d 下没有 check 列", grid)
	}
	val = lookupCheck
	if val == "" {
		val = "Y" // .4fd 的 valueChecked；源码里 q_imaf001_15 就是 Y
	}
	// 该列 ValueList 下的格子，按 Seq（声明顺序）= 行序
	type cand struct {
		id  int
		seq int
	}
	var cands []cand
	for _, o := range m.Objects() {
		if o.Type != "Value" || o.Parent != 0 {
			continue
		}
		// 只收本实例的：父链指向 check 列的 ValueList，或值就是"未勾选"字面值
		if !inSubtree(m, o.ID, checkCol.ID) && !isUncheckedValue(o) {
			continue
		}
		cands = append(cands, cand{o.ID, o.Seq})
	}
	if len(cands) == 0 {
		return 0, 0, "", fmt.Errorf("本实例里没有勾选格")
	}
	for i := 1; i < len(cands); i++ {
		for j := i; j > 0 && cands[j-1].seq > cands[j].seq; j-- {
			cands[j-1], cands[j] = cands[j], cands[j-1]
		}
	}
	if row > len(cands) {
		return 0, 0, "", fmt.Errorf("只有 %d 个勾选格，要第 %d 行", len(cands), row)
	}
	return cands[row-1].id, checkCol.ID, val, nil
}

func isUncheckedValue(o *fep.Object) bool {
	v := o.Attr("value")
	return v == "N" || v == "n" || v == "0"
}

func inSubtree(m *fep.Model, id, root int) bool {
	for hop := 0; hop < 8; hop++ {
		o, ok := m.Get(id)
		if !ok || o.Parent == 0 {
			return false
		}
		if o.Parent == root {
			return true
		}
		id = o.Parent
	}
	return false
}

func appendLookupAudit(pl *lookupPlan) {
	b, _ := json.Marshal(pl)
	e := fep.AuditEntry{
		Time:   time.Now(),
		Verb:   "do.lookup",
		Target: pl.Opener,
		Result: pl.Result,
		Note:   string(b),
	}
	_ = fep.AppendAudit(filepath.Join(recDirForAudit(nil), "audit.log"), e)
}

func init() {
	doLookupCmd.Flags().StringVar(&lookupQuery, "query", "", "查询条件（写到构造阶段的第一列）")
	doLookupCmd.Flags().IntVar(&lookupRow, "row", 1, "目标行（1 起；0 = 只查询不选行）")
	doLookupCmd.Flags().StringVar(&lookupCheck, "check", "", "勾选写什么值（缺省取 .4fd 的 valueChecked，通常是 Y）")
	doLookupCmd.Flags().StringVar(&lookupSource, "source", "", "源码镜像根（读选中语义用；缺省取配置的 mirror.dir）")
	doLookupCmd.Flags().BoolVar(&lookupAssume, "assume", false,
		"允许在**没有源码**时按阶段的常见语义推断（InputArray→勾选 / DisplayArray→当前行）。"+
			"默认关：推断是观察不是规矩，输出里会标【推断·非源码】")
	doLookupCmd.Flags().DurationVar(&lookupWait, "wait", 8*time.Second, "每一步等运行期回包多久")
	doCmd.AddCommand(doLookupCmd)
}
