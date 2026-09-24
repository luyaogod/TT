package dict

import (
	"fmt"
	"strings"

	"tt/internal/dict/db"
	"tt/internal/output"

	"github.com/spf13/cobra"
)

var (
	progKW    string
	progLang  string
	progLimit int
	progSub   bool
)

var progCmd = &cobra.Command{
	Use:     "prog [程序编号]",
	Aliases: []string{"program", "job"},
	Short:   "查询程序/作业字典 (程序做什么、被哪些作业使用)",
	Long: `查询 T100 的程序与作业登记(azzi900 程式基本資料 / azzi910 作業基本資料):
程序编号对应的中文作业名、程序类别、归属模块、是否客制、引用主程序,
以及**哪些作业用了这个程序**——作业通过 gzzz_t.gzzz002 挂到程序上,
一个程序可以被多个作业使用(作业名称取它所挂程序的名称,与 azzi910 一致)。

读代码时的第一个问题"这个程序做什么"就用它;也可以拿作业编号去查它挂的是哪个程序。
主程序查不到时会继续查**子程序/元件登记**(gzde_t,参考作业 azzi901):子程序(aapq110_01)、
应用元件/库(cl_abi、s_xxx)都能给出它自己的说明与规格类别——两类登记互不重叠。
子程序/元件的列表与搜索用 --sub。
无参数时列出全部程序(--kw 按程序编号/中文名称搜索,从业务词找程序)。
程序名称有简体(--lang zh_CN,默认)与繁体(--lang zh_TW)两份,搜索用字要对应;
一个程序被很多作业使用时(如共用维护程序可达数百个),作业表默认只列前 --limit 行。
数据族为"程序与作业/程序与表格/子程序与元件"(--help 末尾会显示本地是否已同步)。`,
	Example: `  tt dict prog aapi011          # 程序做什么 + 被哪些作业使用 + 用了哪些表
  tt dict prog aapq110_01       # 子程序:自己的说明 + 它的主程序
  tt dict prog cl_abi           # 库/元件:说明 + 用表清单
  tt dict prog --kw 对帐        # 按中文作业名找程序
  tt dict prog --sub --kw ABI   # 在子程序/元件里按业务词搜
  tt dict prog aapi011 --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if progSub {
				return runSubProgList()
			}
			return runProgList()
		}
		return runProgDetail(args[0])
	},
}

// runProgDetail 打印一个程序编号的详情:程序信息 + 使用它的作业。
// 编号只登记为作业时,跟进它挂的程序再打印一遍(一眼看到"谁在用它")。
func runProgDetail(code string) error {
	info, err := GetDB().QueryProgInfo(code, progLang)
	if err != nil {
		if db.IsMissingTable(err) {
			fmt.Println(missingHint("程序与作业 (gzza_t/gzzz_t 等表)"))
			return printProgTables(code, true) // 用表索引是另一族,可能已同步
		}
		return err
	}
	if info == nil {
		// 主程序登记(gzza_t)里没有 → 查子程序/元件/库登记(gzde_t,参考作业 azzi901)。
		// 两张表互不重叠:实测 gzde_t 4,036 个 / gzza_t 4,147 个、交集 0。
		sub, serr := GetDB().QuerySubProgInfo(code, progLang)
		if serr != nil && !db.IsMissingTable(serr) {
			return serr
		}
		if sub != nil {
			printSubProgInfo(sub)
			if base := progBaseCode(code); base != "" {
				if bi, err := GetDB().QueryProgInfo(base, progLang); err == nil && bi != nil && bi.IsProg {
					fmt.Printf("它的主程序: %s(%s);子程序自己做的事以文件头 `#+ Description:` 为准。\n", base, bi.Name)
				}
			}
			return printProgTables(code, true)
		}
		fmt.Printf("未找到程序/作业/元件 '%s' 的登记。\n", code)
		if base := progBaseCode(code); base != "" {
			// 子程序/子元件形态,但两张登记表都没有:只提示主程序
			if bi, err := GetDB().QueryProgInfo(base, progLang); err == nil && bi != nil && bi.IsProg {
				fmt.Printf("提示: '%s' 是子程序/子元件形态,主程序为 '%s'(%s)。\n", code, base, bi.Name)
			} else {
				fmt.Printf("提示: '%s' 是子程序/子元件形态,主程序代码应是 '%s'(工具里没登记它)。\n", code, base)
			}
			fmt.Println("      子程序自己做什么,看它文件头的 `#+ Description:`;别把主程序的用途当成它的。")
		} else {
			fmt.Println("提示: 用业务词搜索试试 —— tt dict prog --kw <关键字>")
			fmt.Println("      库/元件/开窗(cl_abi、s_xxx、q_xxx)不登记为程序与作业,但下面的用表索引里可能有它。")
		}
		// 没登记为程序≠查不到:gzdg_t(用表索引)覆盖库/元件/开窗等(实测 cl_abi 22 条、s_apcp300 8 条)
		return printProgTables(code, true)
	}

	if Format() != output.FormatTable {
		jobs, jerr := GetDB().QueryProgJobs(code, progLang)
		if jerr != nil && !db.IsMissingTable(jerr) {
			return jerr
		}
		tables, terr := GetDB().QueryProgTables(code, progLang)
		if terr != nil && !db.IsMissingTable(terr) {
			return terr
		}
		return emitOne(map[string]any{"程序": info, "作业": jobs, "表格": tables})
	}

	if info.IsProg {
		printProgInfo(info)
	} else {
		fmt.Printf("=== %s ===\n该编号不是程序登记,是作业:挂的程序 = %s\n", code, orDash(info.JobProg))
		if info.JobProg != "" && info.JobProg != code {
			// 跟到它挂的程序:程序信息 + 作业清单(本作业会出现在其中)
			pi, err := GetDB().QueryProgInfo(info.JobProg, progLang)
			if err != nil && !db.IsMissingTable(err) {
				return err
			}
			if pi != nil && pi.IsProg {
				fmt.Println()
				printProgInfo(pi)
			}
		}
		return nil
	}

	jobs, err := GetDB().QueryProgJobs(code, progLang)
	if err != nil {
		if db.IsMissingTable(err) {
			fmt.Println(missingHint("程序与作业 (gzzz_t 等表)"))
			return printProgTables(code, false) // 程序与表格可能是另一族,继续试
		}
		return err
	}
	fmt.Printf("\n使用它的作业 (%d):\n", len(jobs))
	if len(jobs) == 0 {
		fmt.Println("  (无;该程序未被作业登记,或是被作其他用途引用)")
		return printProgTables(code, false)
	}
	headers := []string{"作业编号", "作业名称", "归属模块", "应用参数组", "参数组说明", "默认单据性质"}
	shown := jobs
	if progLimit > 0 && len(jobs) > progLimit {
		shown = jobs[:progLimit]
	}
	var rows [][]string
	for _, j := range shown {
		rows = append(rows, []string{j.JobCode, j.JobName, j.Module, j.ParamGrp, j.ParamDesc, j.DocType})
	}
	output.PrintTable(headers, rows)
	if len(shown) < len(jobs) {
		fmt.Printf("\n… 还有 %d 个未显示(共 %d 个);--limit 0 显示全部,或 --json 导出\n",
			len(jobs)-len(shown), len(jobs))
	}
	return printProgTables(code, false)
}

// printSubProgInfo 打印子程序/元件/库的登记(gzde_t + gzdel_t,参考作业 azzi901)。
func printSubProgInfo(p *db.SubProgInfo) {
	fmt.Printf("=== %s ===\n", p.Code)
	if p.Name != "" {
		fmt.Printf("说明: %s\n", p.Name)
	}
	var parts []string
	if p.Category != "" {
		parts = append(parts, "规格类别: "+p.Category+db.SubProgCategoryLabel(p.Category))
	}
	if p.Module != "" {
		parts = append(parts, "归属模块: "+p.Module)
	}
	if p.ProgCat != "" {
		parts = append(parts, "程序类别: "+p.ProgCat+categoryLabel(p.ProgCat))
	}
	if p.Cust != "" {
		parts = append(parts, "客制: "+p.Cust)
	}
	if p.Status != "" {
		parts = append(parts, "状态码: "+p.Status)
	}
	for _, x := range parts {
		fmt.Println(x)
	}
	if p.Industry != "" {
		fmt.Printf("归属行业别: %s\n", p.Industry)
	}
	fmt.Println("(登记在「子程序及应用元件基本数据表」gzde_t,不在程序/作业登记里)")
}

// runSubProgList 列出/搜索子程序与元件(--sub,--kw 按编号或说明过滤)。
func runSubProgList() error {
	list, err := GetDB().QuerySubProgList(progLang, progKW)
	if err != nil {
		if db.IsMissingTable(err) {
			return missingTableErr("子程序与元件 (gzde_t/gzdel_t)")
		}
		return err
	}
	if len(list) == 0 {
		if progKW != "" {
			fmt.Printf("没有匹配 '%s' 的子程序/元件。\n提示: 说明可能只有繁体或其它语言,换 --lang 或更短的词再试。\n", progKW)
		} else {
			fmt.Println(emptyHint("子程序与元件"))
		}
		return nil
	}
	headers := []string{"规格编号", "说明", "规格类别", "归属模块", "客制"}
	var rows [][]string
	for _, p := range list {
		rows = append(rows, []string{p.Code, p.Name, p.Category, p.Module, p.Cust})
	}
	if Format() != output.FormatTable {
		return emit(list, headers, rows)
	}
	output.PrintTable(headers, rows)
	fmt.Printf("\n共 %d 个", len(list))
	if progKW == "" {
		fmt.Print(";用 --kw <业务词> 过滤")
	}
	fmt.Println()
	return nil
}

// printProgTables 打印"这个编号用了哪些表"(gzdg_t,由 T100 自己维护;参考作业 azzq902)。
// 该数据族缺失时只提示、不影响上面的程序与作业信息。
// fromNotFound=true 表示调用方刚说过"未登记为程序/作业":索引里没记录就静默,
// 有记录则说明它是库/元件/开窗之类——照样给出来(实测 cl_abi 22 条、s_apcp300 8 条)。
func printProgTables(code string, fromNotFound bool) error {
	tables, err := GetDB().QueryProgTables(code, progLang)
	if err != nil {
		if db.IsMissingTable(err) {
			fmt.Println()
			fmt.Println(missingHint("程序与表格 (gzdg_t)"))
			return nil
		}
		return err
	}
	if len(tables) == 0 {
		if fromNotFound {
			return nil
		}
		fmt.Println("\n使用的表格 (0):\n  (无登记;该程序未在 gzdg_t 登记用表,或只用了动态 SQL)")
		return nil
	}
	fmt.Printf("\n使用的表格 (%d):\n", len(tables))
	shown := tables
	if progLimit > 0 && len(tables) > progLimit {
		shown = tables[:progLimit]
	}
	var rows [][]string
	for _, t := range shown {
		rows = append(rows, []string{t.Table, t.TableDesc, t.Ops})
	}
	output.PrintTable([]string{"表格编号", "表说明", "操作"}, rows)
	if len(shown) < len(tables) {
		fmt.Printf("\n… 还有 %d 张未显示(共 %d 张);--limit 0 显示全部\n", len(tables)-len(shown), len(tables))
	}
	fmt.Println("操作: S=SELECT 查询 / I=INSERT 新增 / U=UPDATE 修改 / D=DELETE 删除")
	return nil
}

// printProgInfo 打印程序登记信息。
func printProgInfo(p *db.ProgInfo) {
	fmt.Printf("=== %s ===\n", p.Code)
	if p.Name != "" {
		fmt.Printf("程序名称: %s\n", p.Name)
	}
	if p.ShortName != "" {
		fmt.Printf("程序简称: %s\n", p.ShortName)
	}
	parts := make([]string, 0, 5)
	if p.Category != "" {
		parts = append(parts, "程序类别: "+p.Category+categoryLabel(p.Category))
	}
	if p.Module != "" {
		parts = append(parts, "归属模块: "+p.Module)
	}
	if p.Cust != "" {
		parts = append(parts, "客制: "+p.Cust)
	}
	if p.Status != "" {
		parts = append(parts, "状态码: "+p.Status)
	}
	for _, s := range parts {
		fmt.Println(s)
	}
	if p.RefMain != "" {
		fmt.Printf("引用主程序: %s\n", p.RefMain)
	}
	if p.RunCmd != "" {
		fmt.Printf("系统运行指令: %s\n", p.RunCmd)
	}
}

// runProgList 列表模式:列出/搜索程序(附各自的作业数)。
func runProgList() error {
	list, err := GetDB().QueryProgList(progLang, progKW)
	if err != nil {
		if db.IsMissingTable(err) {
			fmt.Println(missingHint("程序与作业 (gzza_t 等表)"))
			return nil
		}
		return err
	}
	if len(list) == 0 {
		if progKW != "" {
			fmt.Printf("没有匹配 '%s' 的程序。\n提示: 程序名称分简体(--lang zh_CN,默认)与繁体(--lang zh_TW)两份,用字可能是异体(如 对账/对帐/對帳);换个写法、换 --lang 或用更短的词再试。\n", progKW)
		} else {
			fmt.Println(emptyHint("程序与作业"))
		}
		return nil
	}

	headers := []string{"程序编号", "程序名称", "程序类别", "归属模块", "客制", "作业数"}
	var rows [][]string
	for _, p := range list {
		rows = append(rows, []string{p.Code, p.Name, p.Category, p.Module, p.Cust, fmt.Sprintf("%d", p.JobCount)})
	}
	if Format() != output.FormatTable {
		return emit(list, headers, rows)
	}
	output.PrintTable(headers, rows)
	fmt.Printf("\n共 %d 个程序", len(list))
	if progKW == "" {
		fmt.Print(";用 --kw <业务词> 按程序编号/中文名称过滤,如 tt dict prog --kw 对帐")
	}
	fmt.Println()
	return nil
}

// categoryLabel 程序类别码的中文补充(取 T100 程序编号第 4 码的类别约定)。
// 注意 ERP 里存的是大写(gzza002='I'),这里统一转小写再比——否则注解永远不出现。
func categoryLabel(code string) string {
	switch strings.ToLower(code) {
	case "i":
		return "(基本资料维护)"
	case "m":
		return "(主档维护)"
	case "t":
		return "(交易处理)"
	case "s":
		return "(参数设定)"
	case "p":
		return "(批次处理)"
	case "q":
		return "(查询)"
	case "r":
		return "(报表)"
	default:
		return ""
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// progBaseCode 把子程序/子元件编号还原成主程序编号:aapq110_01 → aapq110、
// axmi125_wf → axmi125、aapt110_01_rep → aapt110。不是这些形态时返回空串
// (普通程序编号如 aapi011、开窗码如 q_adzi052 都不动)。
func progBaseCode(code string) string {
	base := code
	for {
		i := strings.LastIndex(base, "_")
		if i <= 0 || !isSubSuffix(base[i+1:]) {
			break
		}
		base = base[:i]
	}
	if base == code {
		return ""
	}
	return base
}

// isSubSuffix 判断是否为子程序/子元件后缀:_01/_02…、_x01、_g01、_k01、_s01、_wf、_rep。
func isSubSuffix(s string) bool {
	switch s {
	case "wf", "rep":
		return true
	}
	if s == "" {
		return false
	}
	if isDigits(s) {
		return true
	}
	switch s[0] {
	case 'x', 'g', 'k', 's':
		return len(s) > 1 && isDigits(s[1:])
	}
	return false
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func init() {
	progCmd.Flags().StringVar(&progKW, "kw", "", "按程序编号/中文名称搜索 (无编号时的列表模式)")
	progCmd.Flags().StringVar(&progLang, "lang", "zh_CN", "程序名称语言别 (gzzal002, 默认 zh_CN)")
	progCmd.Flags().IntVar(&progLimit, "limit", 20, "作业最多显示几行 (0 = 全部;大程序可被数百个作业使用)")
	progCmd.Flags().BoolVar(&progSub, "sub", false, "无编号时:列出/搜索子程序与元件(gzde_t)而不是主程序")
	Group.AddCommand(progCmd)
}
