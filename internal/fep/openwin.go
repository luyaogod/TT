package fep

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// W3：**从源码读开窗的交互语义** —— "选中"到底是什么意思，源码写着，不用猜。
//
// 2026-10-11 的实证（两件事都付了学费）：
//
//  1. 我在料件开窗 `q_imaf001_15` 上试了十几次"选中当前行 + 确定"，运行期每次都只回一个
//     空 ACK、界面什么都不返回 —— 看起来像"形态不对"。翻源码才看到：它的 accept 走
//     `q_imaf001_15_get_multiret()`，而那个函数只认 `check = "Y"`（勾选），
//     **没有勾选就没有回传**。所以"选中"= 勾选，不是"当前行"。
//
//     q_imaf001_15.4gl:363  ON ACTION accept
//     q_imaf001_15.4gl:364     CALL q_imaf001_15_get_multiret()
//     q_imaf001_15.4gl:829     IF gr_qry_sel[li_i].check = "Y" THEN
//     q_imaf001_15.4gl:841        LET g_qryparam.return1 = gr_qry_sel[li_i].imaf001_1
//
//  2. 而供应商开窗 `q_pmaa001_3` **有两个对话框**：一个 `INPUT ARRAY`（勾选式），
//     一个 `DISPLAY ARRAY`（当前行式）—— 同一个开窗里两种语义并存：
//
//     q_pmaa001_3.4gl:362  INPUT ARRAY gr_qry FROM s_qry.*      ← 对话框一
//     q_pmaa001_3.4gl:379  ON ACTION accept → get_multiret()（勾选）
//     q_pmaa001_3.4gl:474  DISPLAY ARRAY gr_qry TO s_qry.*      ← 对话框二
//     q_pmaa001_3.4gl:485  IF li_ac > 0 THEN                    ← 当前行
//     q_pmaa001_3.4gl:486     LET g_qryparam.return1 = gr_qry[li_ac].pmaa001_1
//
// 结论：**"选中一行"必须按开窗、按对话框分别读源码**；靠试是试不出来的（试的结果是
// "没反应"，而真实原因是"你没勾选"）。本文件把这段阅读自动化，每条结论都带出处。

// SelectMode 说明一个对话框里"选中一行"意味着什么。
type SelectMode string

const (
	// SelectCurrentRow：回传**当前行**（accept 用行下标 li_ac / getCurrentRow）。
	// 驱动要做：把光标移到那一行（网格 currentRow + 该列 cursor），再点确定。
	SelectCurrentRow SelectMode = "current-row"
	// SelectChecked：回传**勾选的行**（accept 读 check = valueChecked）。
	// 驱动要做：把光标移到那一行、给它勾选栏写 valueChecked，再点确定。
	SelectChecked SelectMode = "checked"
	// SelectUnknown：读不出判据。**不许猜**（输出里如实标 unknown）。
	SelectUnknown SelectMode = "unknown"
)

// OpenWinDialog 是开窗里一个结果对话框的语义。
type OpenWinDialog struct {
	Dialog       string     `json:"dialog"`      // INPUT ARRAY / DISPLAY ARRAY
	Array        string     `json:"array"`       // gr_qry
	ScreenArray  string     `json:"screenArray"` // s_qry（运行期网格对应的数组名）
	AcceptCalls  []string   `json:"acceptCalls,omitempty"`
	Select       SelectMode `json:"select"`
	ReturnFields []string   `json:"returnFields,omitempty"`
	// 勾选式：勾选栏名 + 勾上/取消的字面值（.4fd 的 CheckBox valueChecked/valueUnchecked）。
	CheckColumn    string `json:"checkColumn,omitempty"`
	ValueChecked   string `json:"valueChecked,omitempty"`
	ValueUnchecked string `json:"valueUnchecked,omitempty"`
	// 当前行式：源码里放行下标的那一行（如 li_ac），驱动要靠它把光标移到目标行。
	RowVar   string   `json:"rowVar,omitempty"`
	Evidence []string `json:"evidence,omitempty"`
}

// OpenWinModel 是一个开窗程序的完整交互语义（每一条都带出处）。
type OpenWinModel struct {
	Program string   `json:"program"`
	Files   []string `json:"files,omitempty"`
	Forms   []string `json:"forms,omitempty"`

	// 查询条件：`CONSTRUCT <var> ON a, b, c` 里的 a,b,c（前端往这些列写条件）。
	ConstructCols []string `json:"constructCols,omitempty"`
	// 结果对话框（**可能多个**；实测 q_pmaa001_3 就有两个，语义不同）。
	Dialogs []OpenWinDialog `json:"dialogs,omitempty"`
	// State 是 `g_qryparam.state` 在本程序里被比较过的取值（调用方决定用哪种模式）。
	States []string `json:"states,omitempty"`

	Evidence []string `json:"evidence,omitempty"`
}

// Select 汇总"这个开窗怎么选中"：只有一个对话框就直接用它；多个对话框时取
// **第一个非 unknown**（并保留 Dialogs 让人看清全貌）。
func (ow *OpenWinModel) Select() SelectMode {
	for _, d := range ow.Dialogs {
		if d.Select != SelectUnknown {
			return d.Select
		}
	}
	return SelectUnknown
}

// SelectDialog 返回承载选中语义的那个对话框（多个时取第一个非 unknown）。
func (ow *OpenWinModel) SelectDialog() *OpenWinDialog {
	for i := range ow.Dialogs {
		if ow.Dialogs[i].Select != SelectUnknown {
			return &ow.Dialogs[i]
		}
	}
	if len(ow.Dialogs) > 0 {
		return &ow.Dialogs[0]
	}
	return nil
}

var (
	reConstruct = regexp.MustCompile(`(?m)^[ \t]*CONSTRUCT\s+[A-Za-z_][A-Za-z0-9_]*\s+ON\s+([^\n]+)`)
	reDialogBeg = regexp.MustCompile(`(?m)^[ \t]*DIALOG\b`)
	reDialogEnd = regexp.MustCompile(`(?m)^[ \t]*END\s+DIALOG`)
	reArrayBeg  = regexp.MustCompile(`(?m)^[ \t]*(INPUT|DISPLAY)\s+ARRAY\s+([A-Za-z_][A-Za-z0-9_]*)\s+(?:FROM|TO)\s+([A-Za-z_][A-Za-z0-9_]*)\s*\.`)
	reActionAt  = regexp.MustCompile(`(?m)^[ \t]*ON\s+ACTION\s+([A-Za-z_][A-Za-z0-9_]*)`)
	reCallAny   = regexp.MustCompile(`CALL\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	reFuncStart = regexp.MustCompile(`(?m)^[ \t]*(?:PRIVATE\s+|PUBLIC\s+)?FUNCTION\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	reRowRef    = regexp.MustCompile(`\[\s*([A-Za-z_][A-Za-z0-9_]*)\s*\]\s*\.\s*([A-Za-z_][A-Za-z0-9_]*)`)
	reCheckEq   = regexp.MustCompile(`\.\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*"([^"]*)"`)
	reStateEq   = regexp.MustCompile(`g_qryparam\.state\s*=\s*"([^"]*)"`)
	reRetAssign = regexp.MustCompile(`g_qryparam\.return1\s*=\s*([^\n]+)`)
	reFieldRef  = regexp.MustCompile(`\.\s*([A-Za-z_][A-Za-z0-9_]*)`)
)

func lineOf(text string, off int) int {
	if off > len(text) {
		return 0
	}
	return strings.Count(text[:off], "\n") + 1
}

// lineOfSnippet 在**原文件**里找这段原文的行号（找不到回 0）。
//
// 为什么要单独一个：判定选中语义时我扫的是"accept 段 + 被调函数体"拼起来的文本，
// 那里的行号在文件里没有意义 —— 出处必须指回原文件（否则就是伪出处）。
func lineOfSnippet(text, snippet string) int {
	snippet = strings.TrimSpace(snippet)
	if snippet == "" {
		return 0
	}
	off := strings.Index(text, snippet)
	if off < 0 {
		return 0
	}
	return lineOf(text, off)
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// funcBody 返回某个 FUNCTION 的函数体（到下一条 FUNCTION 或文件末）。
// 简单的文本层就够：4GL 的函数不嵌套。
func funcBody(text, name string) string {
	idx := reFuncStart.FindAllStringSubmatchIndex(text, -1)
	for i, m := range idx {
		if text[m[2]:m[3]] != name {
			continue
		}
		end := len(text)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		return text[m[0]:end]
	}
	return ""
}

// acceptSegs 返回某个对话框里所有 `ON ACTION accept` 段（各自到下一个 ON ACTION 为止）。
func acceptSegs(block string) []struct {
	seg  string
	line int
} {
	var out []struct {
		seg  string
		line int
	}
	idx := reActionAt.FindAllStringSubmatchIndex(block, -1)
	for i, m := range idx {
		if block[m[2]:m[3]] != "accept" {
			continue
		}
		end := len(block)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		out = append(out, struct {
			seg  string
			line int
		}{block[m[0]:end], lineOf(block, m[0])})
	}
	return out
}

// ParseOpenWin4gl 从开窗程序源码里读出交互语义。
//
// files 用来给出处（文件 + 行）；传 nil 也能解析，只是出处里只有行号。
func ParseOpenWin4gl(program, text string, files map[string]string) *OpenWinModel {
	ow := &OpenWinModel{Program: program}
	src := ""
	for f := range files {
		src = f
		break
	}
	ev := func(line int, snippet string) {
		where := ""
		if src != "" {
			where = fmt.Sprintf("%s:%d ", src, line)
		}
		ow.Evidence = append(ow.Evidence, where+strings.TrimSpace(snippet))
	}

	if m := reConstruct.FindStringSubmatchIndex(text); m != nil {
		for _, c := range strings.Split(text[m[2]:m[3]], ",") {
			if c = strings.TrimSpace(c); c != "" {
				ow.ConstructCols = append(ow.ConstructCols, c)
			}
		}
		ev(lineOf(text, m[0]), strings.TrimSpace(text[m[0]:m[1]]))
	}
	for _, m := range reStateEq.FindAllStringSubmatch(text, -1) {
		if !contains(ow.States, m[1]) {
			ow.States = append(ow.States, m[1])
		}
	}

	// 逐个 DIALOG 块解析（实测一个开窗可能有多个对话框，语义不同）
	begs := reDialogBeg.FindAllStringIndex(text, -1)
	for _, b := range begs {
		ends := reDialogEnd.FindStringIndex(text[b[0]:])
		block := text[b[0]:]
		if ends != nil {
			block = text[b[0] : b[0]+ends[1]]
		}
		am := reArrayBeg.FindStringSubmatchIndex(block)
		segs := acceptSegs(block)
		if am == nil && len(segs) == 0 {
			continue
		}
		d := OpenWinDialog{Dialog: "DIALOG", Select: SelectUnknown}
		if am != nil {
			d.Dialog = block[am[2]:am[3]] + " ARRAY"
			d.Array = block[am[4]:am[5]]
			d.ScreenArray = block[am[6]:am[7]]
		}
		// accept 段 + 它调用的函数体 = 判定选中语义的**全部**依据
		scope := ""
		for _, s := range segs {
			d.Evidence = append(d.Evidence, fmt.Sprintf("%s:%d ON ACTION accept",
				src, lineOf(text, b[0])+s.line-1))
			scope += "\n" + s.seg
			for _, c := range reCallAny.FindAllStringSubmatch(s.seg, -1) {
				if !contains(d.AcceptCalls, c[1]) {
					d.AcceptCalls = append(d.AcceptCalls, c[1])
				}
				if body := funcBody(text, c[1]); body != "" {
					d.Evidence = append(d.Evidence, fmt.Sprintf("%s:%d FUNCTION %s（accept 调用）",
						src, lineOf(text, strings.Index(text, strings.TrimSpace(firstLine(body)))), c[1]))
					scope += "\n" + body
				}
			}
		}
		if scope == "" {
			scope = block
		}

		// 行下标变量：`[li_ac].字段` 里那个 li_ac（当前行的依据）
		rowVars := map[string]bool{}
		for _, m := range reRowRef.FindAllStringSubmatch(scope, -1) {
			rowVars[m[1]] = true
		}
		for v := range rowVars {
			lv := strings.ToLower(v)
			if v == "li_ac" || strings.Contains(lv, "curr") {
				d.RowVar = v
				break
			}
		}
		// 勾选式：`.<check> = "Y"`
		for _, m := range reCheckEq.FindAllStringSubmatch(scope, -1) {
			name, val := m[1], m[2]
			if strings.EqualFold(name, "check") || strings.HasSuffix(strings.ToLower(name), "check") {
				d.CheckColumn = name
				d.ValueChecked = val
				d.Evidence = append(d.Evidence, fmt.Sprintf("%s:%d %s",
					src, lineOfSnippet(text, m[0]), strings.TrimSpace(m[0])))
				break
			}
		}
		switch {
		case d.RowVar != "":
			d.Select = SelectCurrentRow
			d.Evidence = append(d.Evidence, fmt.Sprintf("%s:%d 行下标 %s（当前行）",
				src, lineOfSnippet(text, "["+d.RowVar+"]"), d.RowVar))
		case d.CheckColumn != "":
			d.Select = SelectChecked
		}
		// 回传字段：`…[li_ac].pmaa001_1`；勾选栏是**标志**不是值，不收
		for _, m := range reRowRef.FindAllStringSubmatch(scope, -1) {
			if !rowVars[m[1]] || (d.CheckColumn != "" && m[2] == d.CheckColumn) {
				continue
			}
			if !contains(d.ReturnFields, m[2]) {
				d.ReturnFields = append(d.ReturnFields, m[2])
			}
		}
		if len(d.ReturnFields) == 0 {
			for _, m := range reRetAssign.FindAllStringSubmatch(scope, -1) {
				fields := reFieldRef.FindAllStringSubmatch(m[1], -1)
				if len(fields) == 0 {
					continue
				}
				f := fields[len(fields)-1][1]
				if f != "check" && !contains(d.ReturnFields, f) {
					d.ReturnFields = append(d.ReturnFields, f)
				}
			}
		}
		ow.Dialogs = append(ow.Dialogs, d)
	}
	return ow
}

// ApplyCheckBox4fd 用开窗表单补全勾选栏的**字面值**（"勾上"到底写什么）。
//
// 依据（q_imaf001_15.4fd:59 实测原样）：
//
//	<CheckBox … name="check" valueChecked="Y" valueUnchecked="N" widget="CheckBox"/>
//
// 没有它，驱动只能猜"勾是用 Y 还是 1 还是 true"——而猜错就是空 ACK。
func (ow *OpenWinModel) ApplyCheckBox4fd(form string, fields []SrcField) {
	for i := range ow.Dialogs {
		d := &ow.Dialogs[i]
		if d.CheckColumn == "" {
			continue
		}
		for _, f := range fields {
			if f.Name != d.CheckColumn && !strings.HasSuffix(f.Name, "."+d.CheckColumn) {
				continue
			}
			if f.ValueChecked != "" {
				d.ValueChecked = f.ValueChecked
			}
			if f.ValueUnchecked != "" {
				d.ValueUnchecked = f.ValueUnchecked
			}
			d.Evidence = append(d.Evidence, fmt.Sprintf("%s：CheckBox %s valueChecked=%q valueUnchecked=%q",
				form, f.Name, f.ValueChecked, f.ValueUnchecked))
			break
		}
	}
}

// LoadOpenWin 从源码镜像读一个开窗程序的交互语义（找不到文件就如实报错，不猜）。
//
// 开窗程序住在 `com/qry/{4gl,4fd}/`（实测 q_imaf001_15 / q_pmaa001_3 / q_ooba002_1 都在），
// 也兼容 erp/<模块>/4gl 与 com/sub/4gl 下的私有开窗。
func LoadOpenWin(root, program string) (*OpenWinModel, error) {
	if root == "" {
		return nil, fmt.Errorf("fep: 没给源码镜像根（--source 或配置里的 mirror.dir）")
	}
	type cand struct{ gl, fd string }
	var cands []cand
	add := func(dir ...string) {
		p := filepath.Join(append([]string{root}, dir...)...)
		cands = append(cands, cand{
			gl: p + string(filepath.Separator) + program + ".4gl",
			fd: p + string(filepath.Separator) + program + ".4fd",
		})
	}
	add("com", "qry", "4gl")
	add("com", "sub", "4gl")
	if len(program) >= 3 {
		add("erp", strings.ToLower(program[:3]), "4gl")
	}
	// _wf 变体（同一目录）
	base := make([]cand, 0, len(cands))
	for _, c := range cands {
		base = append(base, c)
		base = append(base, cand{
			gl: strings.TrimSuffix(c.gl, ".4gl") + "_wf.4gl",
			fd: strings.TrimSuffix(c.fd, ".4fd") + "_wf.4fd",
		})
	}
	var glPath, fdPath string
	for _, c := range base {
		if glPath == "" {
			if _, err := os.Stat(c.gl); err == nil {
				glPath = c.gl
			}
		}
		if fdPath == "" {
			if _, err := os.Stat(c.fd); err == nil {
				fdPath = c.fd
			}
		}
	}
	// .4fd 与 .4gl 不一定同目录：开窗表单可能只在 .4gl 所在目录或 com/qry/4fd
	if fdPath == "" {
		if p := filepath.Join(root, "com", "qry", "4fd", program+".4fd"); fileExists(p) {
			fdPath = p
		}
	}
	if glPath == "" {
		return nil, fmt.Errorf("fep: 源码镜像里没有开窗程序 %s.4gl（镜像根 %s）", program, root)
	}
	b, err := os.ReadFile(glPath)
	if err != nil {
		return nil, fmt.Errorf("fep: 读 %s: %w", glPath, err)
	}
	rel := func(p string) string {
		if r, err := filepath.Rel(root, p); err == nil {
			return filepath.ToSlash(r)
		}
		return p
	}
	ow := ParseOpenWin4gl(program, string(b), map[string]string{rel(glPath): string(b)})
	ow.Files = []string{rel(glPath)}
	if fdPath != "" {
		if fb, err := os.ReadFile(fdPath); err == nil {
			ow.Forms = []string{rel(fdPath)}
			ow.ApplyCheckBox4fd(rel(fdPath), ParseForm4fd(program, string(fb)))
		}
	}
	return ow, nil
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// SelectModeFor 把"运行期当前对话框类型"映射到"源码里的哪个对话框"，
// 从而得到这一步的**选中语义** —— 这是 W3 的落点：驱动不再猜。
//
// 映射依据（两侧都是实测）：
//   - 运行期 `dialogType` 的取值分布（现场语料 94 个连接统计）：
//     `DisplayArray` 21817 次 / `InputArray` 11328 次 / `Input` 1697 次 / `Construct` 965 次；
//   - 源码里两种结果对话框：`INPUT ARRAY <arr> FROM <屏幕数组>.*`（可输入，实测配 `check` 勾选）
//     与 `DISPLAY ARRAY <arr> TO <屏幕数组>.*`（只看，实测配 `li_ac` 当前行）。
//
// 对不上（例如运行期报的类型源码里没有）时返回 unknown —— **宁可不判，不猜**。
func SelectModeFor(ow *OpenWinModel, dialogType string) (SelectMode, *OpenWinDialog) {
	if ow == nil {
		return SelectUnknown, nil
	}
	switch strings.ToLower(strings.TrimSpace(dialogType)) {
	case "inputarray":
		for i := range ow.Dialogs {
			if ow.Dialogs[i].Dialog == "INPUT ARRAY" {
				return ow.Dialogs[i].Select, &ow.Dialogs[i]
			}
		}
	case "displayarray":
		for i := range ow.Dialogs {
			if ow.Dialogs[i].Dialog == "DISPLAY ARRAY" {
				return ow.Dialogs[i].Select, &ow.Dialogs[i]
			}
		}
	case "construct":
		// 查询阶段：没有"选中"这回事，只有"写条件列 + 确定执行查询"。
		return SelectUnknown, nil
	}
	return SelectUnknown, nil
}

// OpenWinProgramFromWindow 从窗口名推出开窗程序名。
//
// 依据（实测，三种开窗都对得上）：
//
//	窗口 name              → 开窗程序
//	w_qry_q_pmaa001_3      → q_pmaa001_3
//	w_qry_q_imaf001_15     → q_imaf001_15
//	w_qry_q_ooba002_1      → q_ooba002_1
//
// 这样驱动就不需要人告诉它"我刚开的是哪个开窗"：运行期在 AUI 里写着了。
// 对不上（比如主窗 w_apmt500）就回空串 —— 调用方照实降级。
func OpenWinProgramFromWindow(name string) string {
	name = strings.TrimSpace(name)
	if !strings.HasPrefix(name, "w_qry_") {
		return ""
	}
	prog := strings.TrimPrefix(name, "w_qry_")
	if prog == "" {
		return ""
	}
	// 再保险一层：开窗程序一定以 q_ 开头（实测 T100 的命名约定）
	if !strings.HasPrefix(prog, "q_") {
		return ""
	}
	return prog
}

// ReconcileSelect 把"运行期现在在哪个阶段"与"源码里那个对话框怎么选中"对起来。
//
// 返回值：判定出的选中语义、以及**为什么**（给人/给日志看的出处或分歧说明）。
//
// 规则（宁可不判，不猜）：
//   - 运行期说 InputArray/DisplayArray，源码里有对应对话框且读出了语义 → 用源码的（带出处）；
//   - 源码里没有对应对话框，或那份源码读不出语义 → unknown（如实说"判不出来"）；
//   - 没有源码（镜像不在）→ unknown（**不**用"InputArray 就一定是勾选"这种推断顶替 ——
//     这个推断在 2026-10-11 的两个开窗上恰好成立，但它是巧合，不是规矩）。
//
// 只有调用方显式说"我可以接受按阶段的默认语义"时，才用 StageDefaultSelect。
func ReconcileSelect(ow *OpenWinModel, dialogType string) (SelectMode, string) {
	stage := strings.ToLower(strings.TrimSpace(dialogType))
	switch stage {
	case "construct":
		return SelectUnknown, "阶段 Construct：这一步是写查询条件 → 确定执行查询，没有'选中'这回事"
	case "inputarray", "displayarray":
	default:
		return SelectUnknown, fmt.Sprintf("阶段 %q 不认识（只认 Construct / InputArray / DisplayArray）", dialogType)
	}
	if ow == nil {
		return SelectUnknown, fmt.Sprintf("阶段 %s：没有源码可对（给 --source 或在配置里配 mirror.dir）", stage)
	}
	want := "INPUT ARRAY"
	if stage == "displayarray" {
		want = "DISPLAY ARRAY"
	}
	for i := range ow.Dialogs {
		d := &ow.Dialogs[i]
		if d.Dialog != want {
			continue
		}
		if d.Select == SelectUnknown {
			return SelectUnknown, fmt.Sprintf("阶段 %s：源码里的 %s 对话框读不出选中判据（看 %s 的 ON ACTION accept 段）",
				stage, want, ow.Program)
		}
		ev := ""
		if len(d.Evidence) > 0 {
			ev = "；出处 " + d.Evidence[0]
		}
		return d.Select, fmt.Sprintf("阶段 %s ↔ 源码 %s 对话框 → %s（%s）%s",
			stage, want, d.Select, ow.Program, ev)
	}
	return SelectUnknown, fmt.Sprintf("阶段 %s：源码 %s 里没有 %s 对话框 —— 两边对不上，不猜",
		stage, ow.Program, want)
}

// StageDefaultSelect 是"没有源码时按阶段猜"的默认映射。
//
// **明确标注这是推断**：它来自 2026-10-11 对四个 T100 开窗的观察（INPUT ARRAY 用 check 勾选、
// DISPLAY ARRAY 用 li_ac 当前行），但那是观察，不是规矩 —— 所以只在调用方显式接受时才用。
func StageDefaultSelect(dialogType string) SelectMode {
	switch strings.ToLower(strings.TrimSpace(dialogType)) {
	case "inputarray":
		return SelectChecked
	case "displayarray":
		return SelectCurrentRow
	}
	return SelectUnknown
}
