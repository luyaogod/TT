package dict

import (
	"fmt"
	"strings"

	"tt/internal/dict/db"
	"tt/internal/output"

	"github.com/spf13/cobra"
)

var (
	tableKW    string
	tableLang  string
	tableBrief bool
	tableWho   bool
	tableLimit int
)

var tableCmd = &cobra.Command{
	Use:     "r.t [table_name]",
	Aliases: []string{"rt", "table"},
	Short:   "查询数据表字典 (r.t)",
	Long: `查询一张或多张数据表的字典:这张表在系统里做什么(表说明/所属模块/表类型),
以及字段(中文含义/数据类型/长度/主键/必填)、键值与索引。
读代码、看 SQL、查界面字段含义时用它。
支持逗号分隔多个表名;输出简体中文。
无参数时列出全部表(--kw 按表名/中文表说明搜索,从业务词找表);
指定表名显示该表完整字典,--brief 则只给表级信息(表名/说明/模块/类型/字段数),
多表同查时用它避免打出上千行字段明细。
--who 反查"哪些程序在用这张表"(数据来自 gzdg_t 程序与应用表格功能分析表,由 T100
自己维护;参考作业 azzq902「程式編號對應表格查詢」),并给出各程序对它的操作类别
(S=查询/I=新增/U=修改/D=删除)——改表前的影响分析用它。
命令名对齐 T100 原生工具 r.t(旧名 rt/table 仍可用)。`,
	Example: `  tt dict r.t dzea_t
  tt dict r.t "dzea_t,dzeb_t,dzed_t"
  tt dict r.t --kw 应收        # 按中文说明找表(列表模式)
  tt dict r.t --brief "apca_t,apcb_t,glab_t"   # 多表只解语义,不打字段明细
  tt dict r.t glab_t --who                     # 反查:哪些程序在用 glab_t(含操作类别)
  tt dict r.t dzea_t --json
  tt dict r.t dzea_t --env 正式区   # 切到某环境的远程库直查(--env local 回本地)`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runTableList()
		}
		tables := splitNames(args[0])
		if len(tables) == 0 {
			return fmt.Errorf("未指定有效的表名")
		}

		// --who:反查"哪些程序在用这些表"(gzdg_t,参考作业 azzq902),与看字典是两回事
		if tableWho {
			return runTableWho(tables)
		}

		var dicts []*db.TableDict
		for _, t := range tables {
			d, err := queryTableDict(t)
			if err != nil {
				return err
			}
			dicts = append(dicts, d)
		}

		if Format() != output.FormatTable {
			return emit(dicts, tableDetailColumns, tableDetailRows(dicts))
		}

		// --brief:只要表级信息。多表同查时默认会打每张表的全部字段(7 张表 ≈ 700 行),
		// 只想确认"这几张表分别是什么"时用 --brief,和列表模式同样的表。
		if tableBrief {
			headers := []string{"表名", "表说明", "模块", "类型", "字段数"}
			var rows [][]string
			for _, d := range dicts {
				rows = append(rows, []string{d.TableName, d.TableDesc, d.Module, d.TableType,
					fmt.Sprintf("%d", len(d.Fields))})
			}
			output.PrintTable(headers, rows)
			return nil
		}

		for _, d := range dicts {
			printTableDict(d)
		}
		return nil
	},
}

// runTableList 列表模式:列出/搜索表(3,882 张),按中文表说明找表时用它。
func runTableList() error {
	list, err := GetDB().QueryTableList(tableLang, tableKW)
	if err != nil {
		if db.IsMissingTable(err) {
			return missingTableErr("表字典 (dzea_t/dzeb_t 等表)")
		}
		return err
	}
	if len(list) == 0 {
		if tableKW != "" {
			fmt.Printf("没有匹配 '%s' 的表。\n提示: 表说明有简繁两种写法(zh_CN 与原始繁体档),用字可能是异体(如 对账/对帐/對帳);换个写法或更短的词再试。\n", tableKW)
		} else {
			fmt.Println(emptyHint("表字典"))
		}
		return nil
	}

	headers := []string{"表名", "表说明", "模块", "类型", "字段数"}
	var rows [][]string
	for _, t := range list {
		rows = append(rows, []string{t.TableName, t.TableDesc, t.Module, t.TableType, fmt.Sprintf("%d", t.FieldCnt)})
	}
	if Format() != output.FormatTable {
		return emit(list, headers, rows)
	}
	output.PrintTable(headers, rows)
	fmt.Printf("\n共 %d 张表", len(list))
	if tableKW == "" {
		fmt.Print(";用 --kw <业务词> 按表名/中文说明过滤,如 tt dict r.t --kw 应收")
	}
	fmt.Println()
	return nil
}

// queryTableDict assembles a table's complete dictionary (meta + fields + keys + indexes).
func queryTableDict(name string) (*db.TableDict, error) {
	d := &db.TableDict{TableName: name}

	meta, err := GetDB().QueryTableMeta(name)
	if err != nil {
		return nil, err
	}
	if meta != nil {
		d.TableName = meta.TableName
		d.TableDesc = meta.TableDesc
		d.Module = meta.Module
		d.TableType = meta.TableType
	}

	fields, err := GetDB().QueryTable(name)
	if err != nil {
		return nil, err
	}
	d.Fields = fields

	// 键值/索引依赖 dzed_t/dzec_t，部分库可能未含这两张表
	keys, err := GetDB().QueryKeys(name)
	if err != nil {
		if db.IsMissingTable(err) {
			keys = nil
		} else {
			return nil, err
		}
	}
	d.Keys = keys

	indexes, err := GetDB().QueryIndexes(name)
	if err != nil {
		if db.IsMissingTable(err) {
			indexes = nil
		} else {
			return nil, err
		}
	}
	d.Indexes = indexes

	return d, nil
}

// printTableDict prints a table's dictionary as plain text sections.
func printTableDict(d *db.TableDict) {
	if d.TableDesc == "" && d.Module == "" && d.TableType == "" && len(d.Fields) == 0 {
		fmt.Printf("未找到表 '%s' 或该表无字段定义。\n\n", d.TableName)
		return
	}
	fmt.Printf("=== %s ===\n", d.TableName)
	if d.TableDesc != "" {
		fmt.Printf("表说明: %s\n", d.TableDesc)
	}
	var metaParts []string
	if d.Module != "" {
		metaParts = append(metaParts, "模块: "+d.Module)
	}
	if d.TableType != "" {
		metaParts = append(metaParts, "类型: "+d.TableType)
	}
	if len(metaParts) > 0 {
		fmt.Printf("%s\n", strings.Join(metaParts, "    "))
	}

	fmt.Printf("\n字段 (%d):\n", len(d.Fields))
	if len(d.Fields) == 0 {
		fmt.Println("  (无)")
	} else {
		headers := []string{"序号", "字段名", "字段说明", "数据类型", "长度", "主键", "必填", "备注"}
		var rows [][]string
		for _, f := range d.Fields {
			rows = append(rows, []string{f.Seq, f.FieldName, f.FieldDesc,
				f.DataType, f.Length, f.IsPK, f.Required, f.Remark})
		}
		output.PrintTable(headers, rows)
	}

	fmt.Printf("\n键值 (%d):\n", len(d.Keys))
	if len(d.Keys) == 0 {
		fmt.Println("  (无)")
	} else {
		headers := []string{"键名", "类型", "键值字段", "外键表", "外键字段"}
		var rows [][]string
		for _, k := range d.Keys {
			rows = append(rows, []string{k.KeyName, keyTypeLabel(k.KeyType), k.KeyField, k.RefTable, k.RefField})
		}
		output.PrintTable(headers, rows)
	}

	fmt.Printf("\n索引 (%d):\n", len(d.Indexes))
	if len(d.Indexes) == 0 {
		fmt.Println("  (无)")
	} else {
		headers := []string{"索引名", "类型", "索引字段"}
		var rows [][]string
		for _, ix := range d.Indexes {
			rows = append(rows, []string{ix.IndexName, ix.IndexType, ix.IndexField})
		}
		output.PrintTable(headers, rows)
	}
	fmt.Println()
}

// printTableCSV flattens the field definitions across tables into CSV rows.
// tableDetailColumns / tableDetailRows 表字典详情的列定义与摊平。
// 列定义只有一处:CSV 表头与 JSON 的键同源于 db.TableDict 这一份 DTO。
var tableDetailColumns = []string{"表名", "序号", "字段名", "字段说明", "数据类型", "长度", "主键", "必填", "备注"}

func tableDetailRows(dicts []*db.TableDict) [][]string {
	var rows [][]string
	for _, d := range dicts {
		for _, f := range d.Fields {
			rows = append(rows, []string{d.TableName, f.Seq, f.FieldName, f.FieldDesc,
				f.DataType, f.Length, f.IsPK, f.Required, f.Remark})
		}
	}
	return rows
}

// runTableWho 反查"哪些程序在用这些表"(gzdg_t 程序与应用表格功能分析表,由 T100 自己维护;
// 参考作业 azzq902 程式編號對應表格查詢)。改表前的影响分析用它。
func runTableWho(tables []string) error {
	var jsonOut []map[string]any
	var outRows [][]string
	structured := Format() != output.FormatTable
	for _, t := range tables {
		meta, err := GetDB().QueryTableMeta(t)
		if err != nil && !db.IsMissingTable(err) {
			return err
		}
		desc := ""
		if meta != nil {
			desc = meta.TableDesc
		}

		progs, err := GetDB().QueryTablePrograms(t, tableLang)
		if err != nil {
			if db.IsMissingTable(err) {
				return missingTableErr("程序与表格 (gzdg_t)")
			}
			return err
		}
		if structured {
			// 机器可读:不打标题行,只收集
			jsonOut = append(jsonOut, map[string]any{"表格编号": t, "表说明": desc, "使用程序": progs})
			for _, p := range progs {
				outRows = append(outRows, []string{t, desc, p.Prog, p.ProgName, p.Ops})
			}
			continue
		}

		title := t
		if desc != "" {
			title = fmt.Sprintf("%s (%s)", t, desc)
		}
		fmt.Printf("=== %s ===\n", title)
		if len(progs) == 0 {
			fmt.Println("没有程序登记使用这张表(或该表只被动态 SQL 访问)。")
			fmt.Println()
			continue
		}
		fmt.Printf("使用它的程序 (%d):\n", len(progs))
		shown := progs
		if tableLimit > 0 && len(progs) > tableLimit {
			shown = progs[:tableLimit]
		}
		var rows [][]string
		for _, p := range shown {
			rows = append(rows, []string{p.Prog, p.ProgName, p.Ops})
		}
		output.PrintTable([]string{"程序编号", "程序名称", "操作"}, rows)
		if len(shown) < len(progs) {
			fmt.Printf("\n… 还有 %d 个未显示(共 %d 个);--limit 0 显示全部,或 --json 导出\n",
				len(progs)-len(shown), len(progs))
		}
		fmt.Println("操作: S=SELECT 查询 / I=INSERT 新增 / U=UPDATE 修改 / D=DELETE 删除")
		fmt.Println()
	}
	if structured {
		return emit(jsonOut, []string{"表格编号", "表说明", "程序编号", "程序名称", "操作"}, outRows)
	}
	return nil
}

// keyTypeLabel maps dzed_t key type codes to Chinese labels.
func keyTypeLabel(code string) string {
	switch code {
	case "P":
		return "主键"
	case "F":
		return "外键"
	case "U":
		return "唯一"
	default:
		return code
	}
}

func init() {
	tableCmd.Flags().StringVar(&tableKW, "kw", "", "按表名/中文表说明搜索 (无表名时的列表模式)")
	tableCmd.Flags().StringVar(&tableLang, "lang", "zh_CN", "表说明语言别 (dzeal002, 默认 zh_CN)")
	tableCmd.Flags().BoolVar(&tableBrief, "brief", false, "只给表级信息(表名/说明/模块/类型/字段数),不打字段明细")
	tableCmd.Flags().BoolVar(&tableWho, "who", false, "反查哪些程序在用这些表(含操作类别 S/I/U/D)")
	tableCmd.Flags().IntVar(&tableLimit, "limit", 20, "--who 最多显示几个程序 (0 = 全部)")
	Group.AddCommand(tableCmd)
}

func splitNames(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
