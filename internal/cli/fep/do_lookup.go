package fep

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
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
		// 开窗整条流程结束后要交一份"这一步之后"的报告（树/焦点/值），
		// 所以先把**开窗之前**的现场记下来 —— 整条开窗在用户眼里只是"一步"。
		beforeModel, _, _, _ := loadModel("")
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
		// 明细列的开窗按钮要求**光标先落在那一行那一列** —— 运行期才把它置为可用
		// （实测：先发 {网格}{currentRow}{currentColumn}，运行期回 un {currentRow …}
		// 并让相关动作 active=1）。所以顺序是：定位 → 重读 → 校验 → 点。
		// 先校验再定位会把一个"马上就能用"的按钮当成不可用（今天就这么卡了一次）。
		if grid := fep.GridOfColumn(m, id); grid != 0 {
			row := maxi(pl.Row, 1)
			step("先把光标放进网格 #%d 第 %d 行（第 2 列）再点开窗", grid, row)
			if err := send(fep.ConfigureGridCursorText(grid, row, 2)); err != nil {
				return fail(exitFailed, "定位光标：%v", err)
			}
			time.Sleep(1 * time.Second)
			if m2, rerr := reload(); rerr == nil {
				m = m2
			}
		}
		// 发前校验（R1）：不在当前交互指令里的动作、actionActive=0 的按钮，**发了也是空 ACK**。
		if cerr := fep.CheckClick(m, id); cerr != nil {
			return fail(exitState, "开窗按钮 #%d 现在不能点：%v", id, cerr)
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
		pl.WindowNm = rawWindowName(m, win)
		pl.Program = fep.OpenWinProgramFromWindow(pl.WindowNm)
		step("开窗已出现：窗口 #%d %s（开窗程序：%s）", win, displayName(m, win), orDash(pl.Program))

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
			// 窗口刚出来的那一刻运行期常还没推 dialogType（实测：窗口声明与 dialogType
			// 不在同一条消息里）—— 等一下再问，而不是拿旧值顶替。
			for i := 0; stage == "" && i < 8; i++ {
				time.Sleep(900 * time.Millisecond)
				m2, rerr := reload()
				if rerr != nil {
					return fail(exitFailed, "%v", rerr)
				}
				m = m2
				grid, stage = gridOfWindow(m, win)
			}
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
				// 顺序：先把光标放进目标行（INPUT ARRAY 的 ON CHANGE 只认当前行）→ 写勾选值
				// → **验证**（没被运行期确认就不往下走，R1）。为什么是"候选 + 验证"而不是算一个：
				// 运行期复用裸 id，同一窗口下同时挂着旧实例与本实例的格子，"哪一批是本页"
				// 只有运行期知道 —— 所以按声明顺序列一串、逐个写、逐个认（每个都是一次叶子写值，安全）。
				_ = send(fep.ConfigureGridCursorText(grid, pl.Row, 1))
				time.Sleep(900 * time.Millisecond)
				cands := checkCellCandidates(m, grid, pl.Row)
				if len(cands) == 0 {
					return fail(exitTarget, "网格 #%d 上找不到可写的勾选格（本实例）", grid)
				}
				val := lookupCheck
				if val == "" {
					val = "Y" // .4fd 的 valueChecked；源码里 q_imaf001_15 就是 Y
				}
				pl.CheckVal = val
				confirmed := false
				for _, c := range cands {
					before := runtimeBytes()
					_ = send(fep.ConfigureValueText(c, val, nil))
					time.Sleep(900 * time.Millisecond)
					m3, rerr := reload()
					if rerr != nil {
						// 运行期把程序打停了：**立刻停手**（再发就是白费），并说清楚是哪一格
						return fail(exitFailed,
							"写勾选格 #%d 把开窗程序打停了：%v\n  这一格看着像勾选格但不是"+
								"（-6327 = 值发给了容器）—— 回源码/回 ui find 确认勾选列再试", c, rerr)
					}
					if valueOf(m3, c) == val && runtimeBytes() != before {
						pl.Cell = c
						step("勾上第 %d 行：格子 #%d ← %q（运行期已确认）", pl.Row, c, val)
						confirmed = true
						break
					}
				}
				if !confirmed {
					return fail(exitState, "勾选没被运行期确认（试了 %d 个候选格）—— 不猜："+
						"先 tt fep ui find --value N --sort seq-desc 看当前页的勾选格", len(cands))
				}
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
			// 与 do 的命令一致：把"树变了什么 / 值变了什么 / 现在焦点在哪"一起交出来。
			if afterM, rr, _, err := loadModel(""); err == nil {
				imp := &fep.Changes{
					Tree:   fep.DiffModels(beforeModel, afterM),
					Values: fep.ValuesBetween(beforeModel, afterM, rr),
				}
				if f := afterM.FocusNow(); f.FocusID != 0 || f.CursorID != 0 {
					imp.Focus = &f
				}
				printImpact(imp, lookupWait.String())
			}
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

// rawWindowName 取窗口的**原始 name 属性**（如 w_qry_q_imaf001_15）。
//
// 注意别用 m.WindowName()：那是"给人看的显示名"（中文标题，实测
// "据点料件编号查询(排除料件类别为范本或虚拟品)(q_imaf001_15)"），推不出开窗程序。
func rawWindowName(m *fep.Model, win int) string {
	if o, ok := m.Get(win); ok {
		return o.Attr("name")
	}
	return ""
}

func displayName(m *fep.Model, win int) string {
	if o, ok := m.Get(win); ok {
		return o.Attr("text")
	}
	return ""
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
// 以及它的对话框阶段（运行期说的"这一步在哪个阶段"）。
//
// 阶段取自 `DialogTypeFor(win)` 而不是网格自己的 dialogType 属性：后者是**连接级**且**粘滞**的
// （实测运行期同一时刻把同一个值推给一大堆网格，且开窗关掉后不复位）——
// 按网格读会把上一个开窗的阶段当成现在这个网格的（今天就这么误判过）。
func gridOfWindow(m *fep.Model, win int) (int, string) {
	for _, g := range m.TablesOfWindow(win) {
		if _, ok := m.Get(g); ok {
			return g, m.DialogTypeFor(win)
		}
	}
	return 0, m.DialogTypeFor(win)
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

// checkCellCandidates 列出"目标行可能对应的勾选格"，按**声明顺序**（行序）排。
//
// 勾选格的判据来自 .4fd：`valueUnchecked="N"` —— 所以当前值是 N 的那些 Value 就是
// 本页未勾选的勾选格。为什么要列一串而不是算一个：运行期复用裸 id，同一个窗口下
// 同时挂着旧实例与本实例的格子（实测按 id 看是 #59/#105…，按声明顺序才是 #168/#3549…），
// 而"哪一批是本页"只有运行期知道 —— 调用方逐个写、逐个验（每个都是一次叶子写值）。
func checkCellCandidates(m *fep.Model, grid, row int) []int {
	type cand struct {
		id  int
		seq int
	}
	gridWin := m.WindowOf(grid)
	// 严格一点：先找"勾选列的 ValueList"（列名以 check 结尾的 TableColumn 的 ValueList 孩子），
	// 只收它下面的格子。
	//
	// 为什么要严：上一版按"值是 N 的 Value"筛，结果把一个**容器型 Value** 也收进来，
	// 写下去直接 **-6327** 把开窗程序打停（实测：qry_q_imaf001_15.4gl:345）。
	valueLists := map[int]bool{}
	for _, col := range m.Objects() {
		if col.Type != "TableColumn" || !strings.HasSuffix(strings.ToLower(col.Attr("name")), "check") {
			continue
		}
		if w := m.WindowOf(col.ID); w == 0 || (gridWin != 0 && w != gridWin) {
			continue
		}
		for _, vl := range m.Objects() {
			if vl.Type == "ValueList" && vl.Parent == col.ID {
				valueLists[vl.ID] = true
			}
		}
	}
	var all []cand
	for _, o := range m.Objects() {
		if o.Type != "Value" || !valueLists[o.Parent] {
			continue
		}
		all = append(all, cand{o.ID, o.Seq})
	}
	if len(all) == 0 {
		// 退化（找不到列定义时）：仍限定在"值就是未勾选字面值"的 Value 上
		for _, o := range m.Objects() {
			if o.Type != "Value" {
				continue
			}
			v := o.Attr("value")
			if v != "N" && v != "n" && v != "0" {
				continue
			}
			if w := m.WindowOf(o.ID); w == 0 || (gridWin != 0 && w != gridWin) {
				continue
			}
			all = append(all, cand{o.ID, o.Seq})
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].seq < all[j].seq })
	n := len(all)
	if n == 0 {
		return nil
	}
	const batchMargin = 14 // 当前页 + 余量（实测 pageSize=12）
	start := n - batchMargin
	if start < 0 {
		start = 0
	}
	var out []int
	for i := start; i < n; i++ {
		out = append(out, all[i].id)
	}
	if row > 0 && row <= len(out) {
		rot := append([]int(nil), out[row-1:]...)
		rot = append(rot, out[:row-1]...)
		return rot
	}
	return out
}

func valueOf(m *fep.Model, id int) string {
	if o, ok := m.Get(id); ok {
		return o.Attr("value")
	}
	return ""
}

// runtimeBytes 给"这次写有没有被运行期确认"一个可比较的量：
// 取所有活连接里累计接收字节的最大值（单调增），变了就说明运行期确实回了东西。
func runtimeBytes() int64 {
	c, err := newClient()
	if err != nil {
		return 0
	}
	conns, err := c.Status()
	if err != nil {
		return 0
	}
	var best int64
	for _, cc := range conns {
		if cc.BytesRT > best {
			best = cc.BytesRT
		}
	}
	return best
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
