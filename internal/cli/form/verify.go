package form

// `tt form verify`：把离线字段表与**基线**逐条比对（契约 C5.1 / C7.4 / 变更记录 A5）。
//
// 基线的两个来源，优先级写死（A5）：
//
//  1. `--recording <目录>`（或 TT_FEP_RECORD_DIR）：原始录制件，**优先**；
//  2. 否则 `--map <colname-fieldid.json>`（或 TT_FORM_CORPUS_MAP）：归档语料表 ——
//     它是同一批录制派生的**等价基线**，且可离线跑（Q4b 已把原始录制删除）。
//
// 比对口径与 `internal/fep/form42f_test.go` 的 TestFormMatchesRecording **同一套**
// （可比 / 一致 / 仅在语料里 / 歧义项落候选），否则两处迟早各数各的。
//
// 验收（M1）：窗口 `w_axmt500` 可比 **509** 条 → 必须 100% 一致，不做百分比容忍。

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/fep"
)

var (
	verifyModule    string
	verifyForm      string
	verifyWindow    string
	verifyMap       string
	verifyRecording string
	verifyDir       string
	verifyNoCache   bool
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "与语料核对：离线字段表 vs 归档语料表（逐条，必须 100%）",
	Long: `把这张表单解析出来的 colName→fieldId 与基线逐条比：

  可比    语料里有、这张表单里也有的 colName —— **必须相等**
  仅在语料里  语料有、表单没有 —— 允许（同一窗口里还有别的表单/开窗/动态节点）
  歧义项  语料标了 ambiguous 的 —— 表单里的编号必须落在候选里

基线：--recording <目录>（原始录制件，优先）或 --map <colname-fieldid.json>
（归档语料表，也可用 TT_FORM_CORPUS_MAP）。

退出码：0 全部一致 / 6 有不一致（体检不过）/ 1 取件或解析失败 / 2 用法错（没有基线）。`,
	Args: noArgs,
	RunE: runVerify,
}

func init() {
	verifyCmd.Flags().StringVar(&verifyModule, "module", "", "模块名（如 axm）")
	verifyCmd.Flags().StringVar(&verifyForm, "form", "", "表单名，不含扩展名（如 axmt500_wf）")
	verifyCmd.Flags().StringVar(&verifyWindow, "window", "", "语料里的窗口名（默认按表单名推，如 axmt500_wf → w_axmt500）")
	verifyCmd.Flags().StringVar(&verifyMap, "map", "", "归档语料表 colname-fieldid.json（也可用 TT_FORM_CORPUS_MAP）")
	verifyCmd.Flags().StringVar(&verifyRecording, "recording", "", "原始录制件目录（给了就优先用它当基线）")
	verifyCmd.Flags().StringVar(&verifyDir, "dir", "", "从本地目录/镜像取件（离线；也可用 TT_FORM_DIR）")
	verifyCmd.Flags().BoolVar(&verifyNoCache, "no-cache", false, "不读不写表单缓存（每次都重新取）")
}

func runVerify(cmd *cobra.Command, _ []string) error {
	env, err := requireEnv()
	if err != nil {
		return err
	}
	module, form, err := requireModuleForm("verify", verifyModule, verifyForm)
	if err != nil {
		return err
	}
	// 窗口名：显式给了就用它；没给就按表单名推 —— 而**推出**来的那个是启发式（A15），
	// 所以：① 只用于"去哪一节找基线"和报告里的一行诊断；② 推出来在语料里不存在时，
	// 拒绝的理由写成"你没点名窗口"（用法错 + 列出语料里真实的窗口），不是数据判定。
	window := strings.TrimSpace(verifyWindow)
	derived := window == ""
	if derived {
		window = WindowOf(form)
	}

	base, err := loadBaseline(window, derived)
	if err != nil {
		return err
	}
	f, err := loadForm(cmd.Context(), env, module, form, verifyDir, verifyNoCache)
	if err != nil {
		return err
	}

	r := Compare(f, base)
	r.Env, r.Module, r.Window = env, module, window
	r.WindowDerived = derived
	// 录制件那一路没有"窗口分节"这回事：窗口名只是诊断，照原样带出来。
	if base.Window == "" {
		r.WindowsSeen = base.WindowsSeen
	}

	if wantJSON(cmd) {
		if perr := common.PrintJSON(r); perr != nil {
			return perr
		}
	} else {
		printVerifyTable(r)
	}
	if !r.OK() {
		// 报告已经打出去了（stdout 上恰好一个对象），所以这里只带退出码、
		// 并声明"已报告"，免得根命令再叠一份信封或一句散文。
		return &mismatchError{msg: r.Summary()}
	}
	return nil
}

// mismatchError 是"报告已打印，只在退出码上表示失败"的错误。
//
// 为什么要它：`tt form verify` 的失败信息**就是那份报告**（含逐条不一致）。
// 若再让根命令 reportError 打一次，JSON 模式下 stdout 上会出现两个对象 ——
// 而"stdout 恰好一个对象"是这条线刚立下的规则（见 internal/cli/root.go 的 alreadyReported）。
type mismatchError struct{ msg string }

func (e *mismatchError) Error() string         { return e.msg }
func (e *mismatchError) ExitCode() int         { return exitMismatch }
func (e *mismatchError) AlreadyReported() bool { return true }

// ── 基线 ────────────────────────────────────────────────────────────────────

// Baseline 是"被观测到的 colName→fieldId"。
type Baseline struct {
	// Window 是它来自哪个窗口 —— **归档语料表那一路**的事实（那个文件按窗口分节）。
	// 录制件那一路这里是空的：窗口名是启发式结果，不参与判定（A15）。
	Window string
	// Exact：唯一观测（只观测到一个编号）。
	Exact map[string]int
	// Ambiguous：观测到多个编号的 colName → 候选集合（**不猜**，也不静默取最后一个）。
	Ambiguous map[string][]int
	// WindowsSeen：**诊断用**（排序去重）—— "为什么一条都读不出来"的答案在这里。
	// 它**不参与**任何命中/拒绝判定（A15）。
	WindowsSeen []string
}

// loadBaseline 按 A5 的优先级造基线：录制件优先，其次归档语料表。
//
// window 只用于**归档语料表**那一路（那个文件的顶层键就是窗口名，属事实而非启发式）；
// 录制件那一路不按窗口过滤（见 LoadRecordingBaseline）。derived 只影响**报错措辞**：
// 窗口名是推出来的话，要说清"你没点名窗口"—— 而不是让启发式去当拒绝理由（A15）。
func loadBaseline(window string, derived bool) (*Baseline, error) {
	if dir := firstNonEmpty(verifyRecording, os.Getenv("TT_FEP_RECORD_DIR")); dir != "" {
		b, err := LoadRecordingBaseline(dir)
		if err != nil {
			return nil, formUnknown(
				"确认 --recording 指向的是录制件目录（里头应该有 *-rt2gdc.log）；"+
					"只想要归档语料表那条路就先别给 --recording",
				"从录制件建基线失败（目录 %s）：%v", dir, err)
		}
		return b, nil
	}

	mp := firstNonEmpty(verifyMap, os.Getenv("TT_FORM_CORPUS_MAP"))
	if mp == "" {
		return nil, failUsage("verify 需要一份基线：给 --map <colname-fieldid.json>（归档语料表）、" +
			"--recording <录制件目录>，或设 TT_FORM_CORPUS_MAP")
	}
	cm, err := LoadCorpusMap(mp)
	if err != nil {
		return nil, failUsage("读不了归档语料表 %s：%v（它应该在 .agents/Genero Original Factory Compatible/out/semantic/ 下）",
			mp, err)
	}
	b := cm.ByWindow[window]
	if b == nil {
		why := fmt.Sprintf("归档语料表里没有窗口 %q", window)
		if derived {
			why = fmt.Sprintf("没有 --window，按表单名推出来的窗口是 %q（**启发式，只作诊断**）；"+
				"归档语料表里没有这一节", window)
		}
		return nil, failUsage("%s（有：%s）；用 --window 指到其中一个", why, cm.Windows())
	}
	return b, nil
}

// CorpusMap 是一份归档语料表（`out/semantic/colname-fieldid.json`）。
type CorpusMap struct {
	// source 是语料自报的出处（只作诊断，不进输出契约）。
	source   string
	ByWindow map[string]*Baseline
}

// Windows 列出语料里有哪些窗口（排序、逗号分隔）—— "窗口名不对"时这是唯一有用的提示。
func (c *CorpusMap) Windows() string {
	names := make([]string, 0, len(c.ByWindow))
	for w := range c.ByWindow {
		names = append(names, w)
	}
	sort.Strings(names)
	return strings.Join(names, " / ")
}

// LoadCorpusMap 读归档语料表。
//
// 形状与 `internal/fep/form42f_test.go` 读到的一致：
//
//	{"map": {"w_axmt500": {"b_xmda001": 18, "ooff013": {"ambiguous":[{"fieldId":654,…}]}}}}
//
// **认不出的条目形状当场报错**：静默跳过会让"可比数"变小，而小了的可比数看起来
// 只是"语料覆盖少"（这正是最坏的一种假绿）。
func LoadCorpusMap(path string) (*CorpusMap, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Source string                                `json:"source"`
		Map    map[string]map[string]json.RawMessage `json:"map"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("不是合法 JSON: %w", err)
	}
	if len(doc.Map) == 0 {
		return nil, fmt.Errorf("里面没有 map 段（换一份语料表？）")
	}
	out := &CorpusMap{source: doc.Source, ByWindow: make(map[string]*Baseline, len(doc.Map))}
	for win, cols := range doc.Map {
		b := &Baseline{
			Window:    win,
			Exact:     make(map[string]int, len(cols)),
			Ambiguous: map[string][]int{},
		}
		for col, v := range cols {
			var id int
			if err := json.Unmarshal(v, &id); err == nil {
				b.Exact[col] = id
				continue
			}
			var amb struct {
				Ambiguous []struct {
					FieldID int `json:"fieldId"`
				} `json:"ambiguous"`
			}
			if err := json.Unmarshal(v, &amb); err == nil && len(amb.Ambiguous) > 0 {
				ids := make([]int, 0, len(amb.Ambiguous))
				for _, a := range amb.Ambiguous {
					ids = append(ids, a.FieldID)
				}
				b.Ambiguous[col] = ids
				continue
			}
			return nil, fmt.Errorf("窗口 %s 的 %s 条目形状不认识: %s", win, col, string(v))
		}
		out.ByWindow[win] = b
	}
	return out, nil
}

// LoadRecordingBaseline 从**原始录制件目录**重建基线（A5 的优先那条）。
//
// 三条口径：
//
//  1. 字段编号取自 AUI 对象上的 `fieldId` 属性；**没有这个属性就不猜**（拿会话 id 顶替
//     会得到一个"看起来有数、其实全错"的基线）；
//  2. 同一个 colName 被观测到**多个不同编号**（同一窗口里另有一张表单/另一个节点）
//     → 进 `Ambiguous` 候选集合，**绝不静默取最后一个**（A7：多种 fieldId = 真歧义）；
//  3. **不按窗口过滤**（A15）。实测仓库自带的三份夹具里，`m.WindowName` 给出来的是窗口
//     **标题**（"单据别编号数据查询(q_ooba002_1)" / "重新整理"）或空串，而按表单名推出来的
//     `w_<表单>` 一个都对不上 —— 拿它当过滤器会把 13–18 条证据全滤掉，再报"没有字段对象"，
//     那就是一次**由启发式决定的静默拒绝**。窗口名只进 `WindowsSeen` 供诊断。
//
// 录制件的命名沿既有约定：目录里的 `*-rt2gdc.log`（见 fep.ReadModelFrom 的说明）。
func LoadRecordingBaseline(dir string) (*Baseline, error) {
	m, _, _, err := fep.LoadModelFrom(dir, "")
	if err != nil {
		return nil, err
	}
	b := &Baseline{Exact: map[string]int{}, Ambiguous: map[string][]int{}}
	ids := map[string]map[int]int{} // colName → {fieldId: 出现次数}
	wins := map[string]int{}
	for _, o := range m.Objects() {
		if !recordingFieldTypes[o.Type] {
			continue
		}
		col := o.Attr("colName")
		if col == "" {
			col = o.Attr("col")
		}
		idText := strings.TrimSpace(o.Attr("fieldId"))
		if col == "" || idText == "" {
			continue
		}
		id, cerr := strconv.Atoi(idText)
		if cerr != nil {
			continue
		}
		if ids[col] == nil {
			ids[col] = map[int]int{}
		}
		ids[col][id]++
		// 诊断：这个对象挂在哪个窗口下（拿不到就记成"（无窗口）"—— 绝大多数节点如此）
		name := "（无窗口）"
		if win, ok := m.Window(o.ID); ok {
			if n := strings.TrimSpace(m.WindowName(win)); n != "" {
				name = n
			}
		}
		wins[name]++
	}
	for col, byID := range ids {
		list := make([]int, 0, len(byID))
		for id := range byID {
			list = append(list, id)
		}
		sort.Ints(list)
		if len(list) == 1 {
			b.Exact[col] = list[0]
			continue
		}
		b.Ambiguous[col] = list
	}
	b.WindowsSeen = sortedCountKeys(wins)
	if len(b.Exact)+len(b.Ambiguous) == 0 {
		return nil, fmt.Errorf("%s 里没有带 colName/fieldId 的字段对象（%s；观察到的窗口：%s）",
			dir, rtRecordingHint(), strings.Join(b.WindowsSeen, " / "))
	}
	return b, nil
}

// rtRecordingHint 在"录制件读不出东西"时给出下一步（可操作，C4.3）。
func rtRecordingHint() string {
	return "录制件是 *-rt2gdc.log 这种命名"
}

// sortedCountKeys 把计数 map 的键排序后返回（输出要可 diff，不能随 map 顺序抖）。
func sortedCountKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// recordingFieldTypes 是录制件里"可寻址字段"的对象类型（与 .42f 的四类载体对应）。
var recordingFieldTypes = map[string]bool{
	"FormField":        true,
	"TableColumn":      true,
	"PhantomColumn":    true,
	"PhantomFormField": true,
}

// ── 比对 ────────────────────────────────────────────────────────────────────

// Mismatch 是一条对不上的记录（两边各写出自己的编号，便于直接对照）。
//
// Corpus 用 omitempty：`ambiguousBad` 那一路是"表单编号没落在语料候选里"，没有"语料值"
// 可写 —— 输出 `"corpus": 0` 会被读成"语料里它等于 0"。
type Mismatch struct {
	ColName string `json:"colName"`
	Corpus  int    `json:"corpus,omitempty"`
	Form    int    `json:"form"`
}

// VerifyResult 是一次比对的结局（`--json` 下的信封）。
//
// 只有 `show` 的信封被契约 C5.2 冻结了键名；这里的键是**新增**的，口径见 A13。
// 末尾两个键（windowDerived / windowsSeen）是**诊断**：它们记录"窗口名是哪来的"，
// 明确不参与判定（A15）。
type VerifyResult struct {
	Env          string     `json:"env,omitempty"`
	Module       string     `json:"module,omitempty"`
	Form         string     `json:"form"`
	Variant      string     `json:"variant"`
	Window       string     `json:"window"`
	Comparable   int        `json:"comparable"`
	Consistent   int        `json:"consistent"`
	Mismatched   int        `json:"mismatched"`
	CorpusOnly   int        `json:"corpusOnly"`
	FormOnly     int        `json:"formOnly"`
	AmbiguousHit int        `json:"ambiguousHit"`
	Mismatches   []Mismatch `json:"mismatches"`
	AmbiguousBad []Mismatch `json:"ambiguousBad"`
	Pass         bool       `json:"ok"`
	// WindowDerived：窗口名没给、是按表单名推出来的（启发式 ⇒ 只作诊断）。
	WindowDerived bool `json:"windowDerived,omitempty"`
	// WindowsSeen：录制件那一路观察到的窗口名（诊断用，不参与判定）。
	WindowsSeen []string `json:"windowsSeen,omitempty"`
}

// OK 是"这次核对过了"：既要不一致为 0，也要没有"歧义项都没落候选"。
func (r *VerifyResult) OK() bool {
	return len(r.Mismatches) == 0 && len(r.AmbiguousBad) == 0
}

// Summary 是一行结论，人读与错误消息共用同一句话。
func (r *VerifyResult) Summary() string {
	pct := "n/a"
	if r.Comparable > 0 {
		pct = fmt.Sprintf("%.1f%%", 100*float64(r.Consistent)/float64(r.Comparable))
	}
	return fmt.Sprintf("窗口 %s × 表单 %s（变体 %s）：可比 %d 条 / 一致 %d 条（%s）；不一致 %d；"+
		"仅在语料里 %d；仅在表单里 %d；歧义项命中候选 %d",
		r.Window, r.Form, r.Variant, r.Comparable, r.Consistent, pct,
		len(r.Mismatches), r.CorpusOnly, r.FormOnly, r.AmbiguousHit)
}

// Compare 做逐条比对（口径见包内注释与 verify_test.go）。
//
// 结果里的两个清单**排序后返回**：map 的遍历顺序是随机的，而这份输出会被人和 agent
// 拿去逐行看差异 —— 每次跑出来的顺序都不同就等于不可 diff。
func Compare(f *fep.F42Form, b *Baseline) *VerifyResult {
	r := &VerifyResult{
		Form:         string(f.Name),
		Variant:      VariantOf(string(f.Name)),
		Window:       b.Window,
		Mismatches:   []Mismatch{},
		AmbiguousBad: []Mismatch{},
	}
	for col, want := range b.Exact {
		got, ok := f.ByColName[fep.ColName(col)]
		if !ok {
			r.CorpusOnly++
			continue
		}
		r.Comparable++
		if int(got) != want {
			r.Mismatches = append(r.Mismatches, Mismatch{ColName: col, Corpus: want, Form: int(got)})
			continue
		}
		r.Consistent++
	}
	for col, cands := range b.Ambiguous {
		got, ok := f.ByColName[fep.ColName(col)]
		if !ok {
			continue
		}
		hit := false
		for _, c := range cands {
			if c == int(got) {
				hit = true
				break
			}
		}
		if hit {
			r.AmbiguousHit++
			continue
		}
		r.AmbiguousBad = append(r.AmbiguousBad, Mismatch{ColName: col, Form: int(got)})
	}
	// 仅在表单里：诊断用。整张表都不在语料里，通常说明"窗口 ↔ 表单配错了" ——
	// 那时可比数会很小甚至为 0，若没有这个计数，配错会被读成"通过"。
	for col := range f.ByColName {
		s := string(col)
		if _, ok := b.Exact[s]; ok {
			continue
		}
		if _, ok := b.Ambiguous[s]; ok {
			continue
		}
		r.FormOnly++
	}
	r.Mismatched = len(r.Mismatches)
	sortMismatches(r.Mismatches)
	sortMismatches(r.AmbiguousBad)
	r.Pass = r.OK()
	return r
}

func sortMismatches(ms []Mismatch) {
	sort.Slice(ms, func(i, j int) bool { return ms[i].ColName < ms[j].ColName })
}

// WindowOf 按表单名推语料里的窗口名（`axmt500_wf` → `w_axmt500`）。
//
// **这是启发式**（A15）：只用于"没给 --window 时去哪一节找基线"与报告里的一行诊断。
// 它推出来的值在语料里不存在时，拒绝的理由是"你没点名窗口"（用法错，并列出语料里
// 真实存在的窗口）—— 绝不用它去判"命中/不命中"。
//
// 变体后缀（_wf/_tui/_gui）先剥掉：窗口名里没有变体。
func WindowOf(form string) string {
	base := strings.TrimSpace(form)
	low := strings.ToLower(base)
	for _, suf := range []string{"_wf", "_tui", "_gui"} {
		if strings.HasSuffix(low, suf) {
			base = base[:len(base)-len(suf)]
			break
		}
	}
	return "w_" + base
}

// printVerifyTable 给人读的报告：先一行结论，再逐条不一致，最后是诊断。
func printVerifyTable(r *VerifyResult) {
	writeOut("%s\n", r.Summary())
	if r.WindowDerived {
		writeOut("（窗口 %s 是按表单名推出来的启发式，仅作诊断；要指别的用 --window）\n", r.Window)
	}
	if len(r.WindowsSeen) > 0 {
		writeOut("（录制件里观察到的窗口：%s —— 诊断用，不参与判定）\n", strings.Join(r.WindowsSeen, " / "))
	}
	for _, m := range r.Mismatches {
		writeOut("  不一致：%s 语料=%d 表单=%d\n", m.ColName, m.Corpus, m.Form)
	}
	for _, m := range r.AmbiguousBad {
		writeOut("  歧义项没落候选：%s 表单=%d\n", m.ColName, m.Form)
	}
	if !r.OK() {
		writeOut("提示：先 tt form show --env %s --module %s --form %s 看这张表单实际长什么样；"+
			"两边不一致说明「窗口 ↔ 表单」或版本对不上，不要直接改语料\n",
			r.Env, r.Module, r.Form)
	}
}
