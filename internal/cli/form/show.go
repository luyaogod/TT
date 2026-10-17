package form

// `tt form show`：把一张 `.42f` 读成字段表（契约 C5.1 / C5.2）。
//
// **输出契约冻结在 C5.2**（键名固定、可 grep、新增键要改契约）：
//
//	{"env":…,"module":…,"form":…,"variant":…,"build":…,
//	 "fields":[{"colName","fieldId","tag","widget","tabIndex","sqlType",
//	            "required","notNull","comment","textKey","geom":{"width","posX","posY"}}],
//	 "recordViews":[{"tabName","links":[…]}],
//	 "duplicates":[]}
//
// 由 golden 快照守着（testdata/golden/form_show_basic.json，牙齿 b）：改键名必红。

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/fep"
)

var (
	showModule  string
	showForm    string
	showVariant string
	showDir     string
	showNoCache bool
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "查看一张表单的字段表（工号 / 编号 / 控件 / 几何 / 必填）",
	Long: `取一张 .42f 并打字段表。默认给人读（表格）；--json 走契约 C5.2 的冻结键名。

字段表里的 colName（工号）才是**身份**：设计期 .4fd 的 fieldId 每次保存全量重排，
不能当标识用；这里的 fieldId 来自编译后的 .42f，在"表单 + 变体"内稳定。`,
	Args: noArgs,
	RunE: runShow,
}

func init() {
	showCmd.Flags().StringVar(&showModule, "module", "", "模块名（如 axm）")
	showCmd.Flags().StringVar(&showForm, "form", "", "表单名，不含扩展名（如 axmt500_wf）")
	showCmd.Flags().StringVar(&showVariant, "variant", "", "表单变体（默认按文件名后缀推：_tui→tui，其余 gui）")
	showCmd.Flags().StringVar(&showDir, "dir", "", "从本地目录/镜像取件（离线；也可用 TT_FORM_DIR）")
	showCmd.Flags().BoolVar(&showNoCache, "no-cache", false, "不读不写表单缓存（每次都重新取）")
}

func runShow(cmd *cobra.Command, _ []string) error {
	env, err := requireEnv()
	if err != nil {
		return err
	}
	module, form, err := requireModuleForm("show", showModule, showForm)
	if err != nil {
		return err
	}

	f, err := loadForm(cmd.Context(), env, module, form, showDir, showNoCache)
	if err != nil {
		return err
	}

	variant := showVariant
	if strings.TrimSpace(variant) == "" {
		variant = VariantOf(form)
	}
	doc := newShowDoc(env, module, form, variant, f)

	if wantJSON(cmd) {
		return common.PrintJSON(doc)
	}
	printShowTable(doc)
	return nil
}

// loadForm 是"取件 + 解析"的唯一入口（show 与 verify 共用）。
//
// 失败一律归 `FORM_UNKNOWN`（退出码 1，契约 C4.2 / C4.3）：取不到与"不是合法 .42f"
// 是同一档的两条病因，Hint 里给出可操作的下一步。
func loadForm(ctx context.Context, env, module, form, dir string, noCache bool) (*fep.F42Form, error) {
	dir = firstNonEmpty(dir, os.Getenv("TT_FORM_DIR"))
	src, err := sourceFor(env, sourceOpts{Dir: dir, NoCache: noCache})
	if err != nil {
		return nil, formUnknown(formHint(env, module, form, dir), "%v", err)
	}
	raw, err := src.Form(ctx, env, module, form)
	if err != nil {
		return nil, formUnknown(formHint(env, module, form, dir),
			"取不到表单 %s（环境 %s / 模块 %s）：%v", form, env, module, err)
	}
	f, perr := fep.ParseForm42F(form, string(raw))
	if perr != nil {
		return nil, formUnknown(formHint(env, module, form, dir),
			"%v（环境 %s / 模块 %s / %d 字节）", perr, env, module, len(raw))
	}
	return f, nil
}

// formHint 是 C4.3 要的"可操作"提示：把那条能跑通的命令原样写出来。
func formHint(env, module, form, dir string) string {
	flag := ""
	if dir != "" {
		flag = fmt.Sprintf(" --dir %q", dir)
	}
	return fmt.Sprintf("先看这张表单取到了什么：tt form show --env %s --module %s --form %s%s；"+
		"取件失败时可用 --dir <本地镜像/目录> 走离线", env, module, form, flag)
}

// ── 输出文档（C5.2 的键名是契约，字段顺序即 JSON 顺序）──────────────────────

type showDoc struct {
	Env         string          `json:"env"`
	Module      string          `json:"module"`
	Form        string          `json:"form"`
	Variant     string          `json:"variant"`
	Build       string          `json:"build"`
	Fields      []showField     `json:"fields"`
	RecordViews []showRecordVal `json:"recordViews"`
	Duplicates  []string        `json:"duplicates"`
}

type showField struct {
	ColName  string   `json:"colName"`
	FieldID  int      `json:"fieldId"`
	Tag      string   `json:"tag"`
	Widget   string   `json:"widget"`
	TabIndex int      `json:"tabIndex"`
	SQLType  string   `json:"sqlType"`
	Required bool     `json:"required"`
	NotNull  bool     `json:"notNull"`
	Comment  string   `json:"comment"`
	TextKey  string   `json:"textKey"`
	Geom     showGeom `json:"geom"`
}

// showGeom 只带 C5.2 冻结的三个键（width/posX/posY）。
// `.42f` 里还有 height/gridWidth，**刻意不输出**：C5.2 说"新增键要改本契约"。
type showGeom struct {
	Width int `json:"width"`
	PosX  int `json:"posX"`
	PosY  int `json:"posY"`
}

type showRecordVal struct {
	TabName string   `json:"tabName"`
	Links   []string `json:"links"`
}

func newShowDoc(env, module, form, variant string, f *fep.F42Form) showDoc {
	doc := showDoc{
		Env:         env,
		Module:      module,
		Form:        form,
		Variant:     variant,
		Build:       f.Build,
		Fields:      make([]showField, 0, len(f.Fields)),
		RecordViews: make([]showRecordVal, 0, len(f.RecordViews)),
		Duplicates:  append([]string{}, f.Duplicates...),
	}
	for _, fl := range f.Fields {
		doc.Fields = append(doc.Fields, showField{
			ColName:  string(fl.ColName),
			FieldID:  int(fl.FieldID),
			Tag:      fl.Tag,
			Widget:   fl.Widget,
			TabIndex: fl.TabIndex,
			SQLType:  fl.SQLType,
			Required: fl.Required,
			NotNull:  fl.NotNull,
			Comment:  fl.Comment,
			TextKey:  fl.TextKey,
			Geom:     showGeom{Width: fl.Geom.Width, PosX: fl.Geom.PosX, PosY: fl.Geom.PosY},
		})
	}
	for _, rv := range f.RecordViews {
		links := make([]string, 0, len(rv.Links))
		for _, l := range rv.Links {
			links = append(links, string(l.ColName))
		}
		doc.RecordViews = append(doc.RecordViews, showRecordVal{TabName: rv.TabName, Links: links})
	}
	return doc
}

// VariantOf 按表单名推变体（契约 C3.8 **暂定**：以文件名后缀为准；真机若发现更权威
// 的来源，以真机为准并回改该条）。认不出的后缀 → gui（C5.2 的样例 `axmt500_wf` 就是 gui）。
//
// **这是启发式**（契约 §4 **A15**）：它只写进报告的 `variant` 字段，**不参与任何**命中/拒绝
// 判定 —— 与 A7（`Scope.Form`/`Variant` 不作运行期闸门）、A9（`kind` 不淘汰候选）同一条
// 原则：只让可靠的量参与判定。守它的是 `TestVariantHeuristicOnlyDiagnoses`。
func VariantOf(form string) string {
	for _, v := range []string{"gui", "tui"} {
		if strings.HasSuffix(strings.ToLower(form), "_"+v) {
			return v
		}
	}
	return "gui"
}

// printShowTable 给人读的表格。别花哨：一列一个事实，行宽固定，便于肉眼看差异。
func printShowTable(doc showDoc) {
	writeOut("%s / %s / %s（变体 %s，build %s，%d 条字段）\n",
		doc.Env, doc.Module, doc.Form, doc.Variant, doc.Build, len(doc.Fields))
	writeOut("%-4s %-8s %-6s %-17s %-12s %4s %-10s %-4s %-4s %-18s %-18s %s\n",
		"#", "fieldId", "tag", "colName", "widget", "tab", "sqlType", "必填", "非空", "comment", "textKey", "geom(w,x,y)")
	for i, fl := range doc.Fields {
		writeOut("%-4d %-8d %-6s %-17s %-12s %4d %-10s %-4s %-4s %-18s %-18s %d,%d,%d\n",
			i, fl.FieldID, fl.Tag, fl.ColName, fl.Widget, fl.TabIndex, fl.SQLType,
			yesNo(fl.Required), yesNo(fl.NotNull), fl.Comment, fl.TextKey,
			fl.Geom.Width, fl.Geom.PosX, fl.Geom.PosY)
	}
	for _, rv := range doc.RecordViews {
		writeOut("RecordView %s → %s\n", rv.TabName, strings.Join(rv.Links, " "))
	}
	if len(doc.Duplicates) > 0 {
		writeOut("⚠ 同表单内重名的 colName（定位时必须拒）：%s\n", strings.Join(doc.Duplicates, " "))
	}
}

func yesNo(b bool) string {
	if b {
		return "是"
	}
	return ""
}
