package dict

// tt dict fieldtab：**字段表**（M3 第一步的消费面）。
//
// 它回答"这张表单/作业有哪些字段，各自的中文名与类型是什么"：
// 数据源是本地目录里的 `.42f`（编译后的表单文件，明文 XML），解析直接调用
// `internal/form42f.ParseForm42F`（**不复制解析器**），中文名走字典既有的
// `dzeb_t`/`dzebl_t` 解析（`FieldNamer` → `QueryTable`）。
//
// 四条纪律：
//  1. **表单只读本地目录**：42f 从 `--root` / `TT_FORM_DIR` 指向的本地目录读；不连真机、不起会话。
//  2. **字典两种通道、缺省在线**（`--source online|local`，见 fieldtab_source.go）：
//     在线 = 直查服务器字典表（只发 SELECT），本地 = 本地 SQLite 镜像（只读）。
//     **不静默降级**：在线不可用就报错并给下一步，绝不偷用本地库（契约 A15）。
//  3. **不改 `tt form show` 的契约**：本命令在 `tt dict` 组下，输出键是**另一套**（见 README），
//     只新增 `dictSource`，既有键一个不改名、不删除。
//  4. **不猜**（契约 A15/A18/A31）：字典查不到就写"未解析"并列原因，绝不拿 colName 充中文名；
//     "本地库没同步"是确定缺失（missing），如实说明并指向 `tt dict db sync --env <环境>`。

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"

	"tt/internal/dict/formfields"
	"tt/internal/output"
)

var (
	// fieldtabRoot 42f 根目录（`--root` 或环境变量 `TT_FORM_DIR`）。
	fieldtabRoot string
	// fieldtabNoDict 只读表单、不接字典（整列未解析）；用于"字典库没同步"的机器。
	fieldtabNoDict bool
)

var fieldtabCmd = &cobra.Command{
	Use:   "fieldtab <模块>/<表单>",
	Short: "字段表:表单有哪些字段、中文名与类型(读本地 .42f;字典在线/本地两种通道)",
	Long: `把"编译后的表单文件"(.42f, 明文 XML)解析成字段表:每个字段的身份(colName)、
表单内编号(fieldId)、载体(FormField/TableColumn/…)、控件(widget)、SQL 类型、
必填,以及**中文名**。

中文名来自字典(dzeb_t/dzebl_t,与 tt dict r.t 同一套解析):把 .42f 里的
cmt_*/lbl_*/colName 键接到字典的字段说明上。**查不到就写「未解析」并列原因,
绝不猜** —— label 为空串、labelState=unresolved,原因分 no-dict / no-table /
not-in-dict / ambiguous。

表单来源是本地目录(参数 --root 或环境变量 TT_FORM_DIR),目录下按
<模块>/42f/<表单>.42f 或 <模块>/<表单>.42f 找;不连服务器、不起会话、只读。

字典来源**两种通道,缺省在线**(--source):
  --source online  直查服务器字典表(只发 SELECT);环境由 --env/--conn 或配置的
                   query.source 决定。连不上/没环境就**报错并给下一步**,绝不偷用本地库。
  --source local   查该环境的本地 SQLite 镜像（<数据目录>/dict/<环境>.db）：
                   用 --env <环境> 或 --env local:<环境> 点名；-d/--db 可显式覆盖。
                   库不存在就报错并指向 tt dict db sync --env <环境>。
用 --no-dict 可显式跳过字典(整列「未解析(no-dict)」)。
两种通道都是只读的;这一轮用的是哪种,JSON 里看 dictSource,人读输出里有一行说明。

输出键(JSON)见 internal/dict/README.md「tt dict fieldtab」;与 tt form show 的键互不影响。
本地没有 42f 目录时，可直接用 tt form show 取线上表单字段（不需要本地镜像）。`,
	Example: `  tt dict fieldtab t100/axmt500_wf --root "D:\tt-fixtures\forms"                     # 缺省:在线
  tt dict fieldtab t100/axmt500_wf --root "D:\tt-fixtures\forms" --source online --env 生产环境
  tt dict fieldtab t100/axmt500_wf --root "D:\tt-fixtures\forms" --source local      # 查本地镜像
  tt dict fieldtab fglfind --root "D:\tt-fixtures\forms" --no-dict                   # 不接字典`,
	Args: cobra.ExactArgs(1),
	// 覆盖命令组的数据源钩子:通道由本命令按 --source 自己开(见 fieldtab_source.go),
	// 这样"42f 目录没给""--source 值打错"这类用法错不会被组的钩子抢先变成"数据库没找到"。
	PersistentPreRunE: skipQuerySource,
	RunE: func(cmd *cobra.Command, args []string) error {
		module, form, err := splitFormArg(args[0])
		if err != nil {
			return err
		}
		root := fieldtabRoot
		if strings.TrimSpace(root) == "" {
			root = os.Getenv("TT_FORM_DIR")
		}
		if strings.TrimSpace(root) == "" {
			return fmt.Errorf("没有 42f 根目录:给 --root <目录> 或设 TT_FORM_DIR"+
				"(只读本地目录,不连真机)\n\n若只是要表单字段表，可改用：%s",
				fieldtabFormShowHint(module, form))
		}

		// 字典通道:**缺省在线**,显式 --source local 才走本地;都打不开就报错并给下一步。
		mode, err := fieldtabSourceMode()
		if err != nil {
			return err
		}
		d, err := openFieldtabDict(mode, openFieldtabOnline, openFieldtabLocal)
		if err != nil {
			return fmt.Errorf("%w\n\n若只是要表单字段表，可改用：%s", err, fieldtabFormShowHint(module, form))
		}
		defer d.Close()

		tab, err := formfields.LoadForm(root, module, form, d.Namer)
		if err != nil {
			return fmt.Errorf("%w\n\n若只是要表单字段表，可改用：%s", err, fieldtabFormShowHint(module, form))
		}

		rows := fieldtabRows(tab)
		if Format() != output.FormatTable {
			// 既有键一个不改:Table 的字段原样平铺,只多一个 dictSource。
			return emitDetail(fieldtabOut{Table: tab, DictSource: string(d.Mode)}, fieldtabCSVHeaders, rows)
		}
		printFieldtab(tab, d)
		return nil
	},
}

// splitFormArg 解析 `<模块>/<表单>`（也接受只有 `<表单>`；`/` 与 `\` 都当分隔符）。
func splitFormArg(arg string) (module, form string, err error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return "", "", fmt.Errorf("表单项为空:用 <模块>/<表单>，如 t100/axmt500_wf")
	}
	arg = strings.ReplaceAll(arg, "\\", "/")
	a := strings.Split(arg, "/")
	switch len(a) {
	case 1:
		if a[0] == "" {
			return "", "", fmt.Errorf("表单项为空:用 <模块>/<表单>，如 t100/axmt500_wf")
		}
		return "", a[0], nil
	case 2:
		if a[0] == "" || a[1] == "" {
			return "", "", fmt.Errorf("表单项 %q 不完整:用 <模块>/<表单>", arg)
		}
		return a[0], a[1], nil
	default:
		return "", "", fmt.Errorf("表单项 %q 层级太多:用 <模块>/<表单>", arg)
	}
}

// fieldtabOut 是 fieldtab 的 JSON 载荷：既有 `Table` 的键**一个不改名、不删除**（嵌入即平铺），
// 只**新增** `dictSource` —— 这一轮到底用的是在线还是本地，必须能从 JSON 里读出来，
// 否则"中文名全是未解析"到底该去 sync 还是去配 --env，只能靠猜。
type fieldtabOut struct {
	*formfields.Table
	DictSource string `json:"dictSource"` // online | local | none
}

// fieldtabCSVHeaders 是 CSV/表格形态的列（JSON 的键见 README）。
var fieldtabCSVHeaders = []string{"colName", "fieldId", "载体", "控件", "SQL类型", "SQL表",
	"必填", "中文名", "解析状态", "命中位置", "未解析原因"}

func fieldtabRows(tab *formfields.Table) [][]string {
	rows := make([][]string, 0, len(tab.Fields))
	for _, f := range tab.Fields {
		label := f.Label
		if f.LabelState != formfields.LabelResolved {
			// 「未解析」必须显式写出来:空单元格会被读成"这一列没有意义"。
			label = "未解析"
		}
		rows = append(rows, []string{
			f.ColName, itoaCLI(f.FieldID), f.Tag, f.Widget, f.SQLType, f.SQLTab,
			yesNo(f.Required), label, string(f.LabelState), f.LabelFrom, f.LabelReason,
		})
	}
	return rows
}

func itoaCLI(n int) string { return fmt.Sprintf("%d", n) }

func yesNo(b bool) string {
	if b {
		return "Y"
	}
	return ""
}

func printFieldtab(tab *formfields.Table, d *fieldtabDict) {
	name := tab.Form
	if tab.Module != "" {
		name = tab.Module + "/" + tab.Form
	}
	fmt.Printf("=== %s ===\n", name)
	fmt.Printf("源文件: %s\n", tab.Path)
	// 这一轮用的是哪种通道：只加一行，既有几行一个都不动。
	fmt.Printf("字典源: %s\n", d.ModeLine())
	if tab.LabelSource == formfields.LabelSourceDict {
		if d.Where != "" {
			fmt.Printf("字典: %s\n", d.Where)
		} else {
			fmt.Println("字典: 已接(来源未记录)")
		}
	} else {
		fmt.Println("字典: 无(中文名整列未解析;--no-dict 或本地库没同步)")
	}
	fmt.Printf("字段 %d 条;中文名 解析 %d / 未解析 %d\n\n",
		len(tab.Fields), tab.LabelResolved, tab.LabelUnresolved)

	output.PrintTable([]string{"colName", "fieldId", "载体", "控件", "SQL类型", "中文名", "状态"},
		fieldtabTextRows(tab))

	if tab.Skipped > 0 {
		fmt.Printf("\n注意: 有 %d 条缺属性(colName/fieldId)的条目被跳过(解析器跳过并计数)。\n", tab.Skipped)
	}
	if len(tab.Duplicates) > 0 {
		fmt.Printf("注意: 同表单内重名的 colName %d 个:%s\n",
			len(tab.Duplicates), strings.Join(tab.Duplicates, ", "))
	}
	for _, n := range tab.Notes {
		fmt.Printf("注意: %s\n", n)
	}
	if tab.LabelUnresolved > 0 {
		fmt.Println("\n未解析 = 字典里没有这个键(或没有可依的表/多个候选各说各话);" +
			"这一列不会回落成 colName。原因见 --json 的 labelReason。")
	}
}

func fieldtabTextRows(tab *formfields.Table) [][]string {
	rows := make([][]string, 0, len(tab.Fields))
	for _, f := range tab.Fields {
		label, st := f.Label, "已解析"
		if f.LabelState != formfields.LabelResolved {
			label, st = "未解析", f.LabelReason
		}
		rows = append(rows, []string{f.ColName, itoaCLI(f.FieldID), f.Tag, f.Widget, f.SQLType, label, st})
	}
	return rows
}

// fieldtabPathHint 之外的东西都留在 fieldtab 里：错误信息由 formfields.FindPath 给全
// （它会列出尝试过的路径），命令层不重复包装、不吞原始错误。

// fieldtabFormShowHint 当本机没有 42f 目录时，给出同一份字段表的另一条入口。
// form show 走线上 .42f，不需要本地镜像；不需要让用户在两个命令组之间自己猜。
func fieldtabFormShowHint(module, form string) string {
	env := strings.TrimSpace(common.Env)
	if strings.HasPrefix(strings.ToLower(env), "local:") {
		env = strings.TrimSpace(env[len("local:"):])
	}
	if env == "" {
		env = "<环境>"
	}
	return fmt.Sprintf("tt form show --env %s --module %s --form %s", env, module, form)
}

func init() { Group.AddCommand(fieldtabCmd) }

func init() {
	fieldtabCmd.Flags().StringVar(&fieldtabRoot, "root", "",
		"42f 根目录(只读本地目录;也可用环境变量 TT_FORM_DIR)")
	fieldtabCmd.Flags().StringVar(&fieldtabSourceName, "source", string(DefaultSourceMode),
		"字典通道:online(直查服务器字典表,缺省) | local(本地 SQLite 镜像)")
	fieldtabCmd.Flags().BoolVar(&fieldtabNoDict, "no-dict", false,
		"不接字典:中文名整列标「未解析」(no-dict),优先于 --source")
}
