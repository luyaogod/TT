package dict

import (
	"fmt"

	"tt/internal/dict/db"
	"tt/internal/output"

	"github.com/spf13/cobra"
)

var (
	sccKW   string
	sccLang string
)

// sccListItem is the display form of one SCC in list mode.
type sccListItem struct {
	ID     string `json:"分类码"`
	Group  string `json:"群组"`
	Status string `json:"状态"`
	Name   string `json:"名称"`
	ValCnt int    `json:"值数"`
}

// sccHeaderView is the display form of the SCC header.
type sccHeaderView struct {
	Group  string `json:"群组"`
	Status string `json:"状态"`
	Aux2   string `json:"gzca002"`
	Aux3   string `json:"gzca003"`
	Name   string `json:"名称"`
	Desc   string `json:"说明"`
	Desc2  string `json:"说明2"`
}

// sccValueView is the display form of one classification value.
type sccValueView struct {
	Val  string `json:"值"`
	Desc string `json:"说明"`
	Sort string `json:"排序"`
	Cust string `json:"标准/客制"`
	B3   string `json:"gzcb003"`
	B4   string `json:"gzcb004"`
	B5   string `json:"gzcb005"`
	B6   string `json:"gzcb006"`
	B7   string `json:"gzcb007"`
	B8   string `json:"gzcb008"`
	B9   string `json:"gzcb009"`
	B10  string `json:"gzcb010"`
	B11  string `json:"gzcb011"`
	B14  string `json:"gzcb014"`
	B15  string `json:"gzcb015"`
}

// sccDetail is the display form of a complete SCC definition.
type sccDetail struct {
	ID     string         `json:"分类码"`
	Header sccHeaderView  `json:"单头"`
	Values []sccValueView `json:"分类值"`
}

var sccCmd = &cobra.Command{
	Use:   "scc [分类码]",
	Short: "查询系统分类码 (SCC)",
	Long: `查询系统分类码 (SCC):系统里下拉/选项的"选项字典",一个分类码 = 一组
「值 → 说明」,画面上的下拉选项就来自这里。
无参数时列出全部分类码(--kw 按分类码/名称过滤);指定分类码显示详情:单头与分类值列表。
分类码多为数字(如 4);部分分类码的值带扩展数据列,含义按各分类码自己定义。`,
	Example: `  tt dict scc
  tt dict scc --kw 币别
  tt dict scc 4
  tt dict scc 4 --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			return runSccDetail(args[0])
		}
		return runSccList()
	},
}

func runSccList() error {
	rows, err := GetDB().QuerySccList(sccLang, sccKW)
	if err != nil {
		if db.IsMissingTable(err) {
			return missingTableErr("系统分类码 (gzca_t 等表)")
		}
		return err
	}
	if len(rows) == 0 {
		fmt.Println(emptyHint("系统分类码"))
		return nil
	}

	items := make([]sccListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, sccListItem{
			ID:     r.ID,
			Group:  r.Group,
			Status: typeWithCode(sccStatusLabel(r.Status), r.Status),
			Name:   r.Name,
			ValCnt: r.ValCnt,
		})
	}

	headers := []string{"分类码", "群组", "状态", "名称", "值数"}
	var tableRows [][]string
	for _, it := range items {
		tableRows = append(tableRows, []string{it.ID, it.Group, it.Status, it.Name, fmt.Sprint(it.ValCnt)})
	}
	if Format() != output.FormatTable {
		return emit(items, headers, tableRows)
	}
	output.PrintTable(headers, tableRows)
	return nil
}

func runSccDetail(id string) error {
	if IsCSV() {
		return fmt.Errorf("详情模式不支持 --csv，请使用列表模式或 --json")
	}

	header, err := GetDB().QuerySccHeader(id, sccLang)
	if err != nil {
		if db.IsMissingTable(err) {
			return missingTableErr("系统分类码 (gzca_t 等表)")
		}
		return err
	}
	if header == nil {
		fmt.Printf("未找到分类码 '%s'。可执行 tt dict scc 查看全部分类码。\n", id)
		return nil
	}

	values, err := GetDB().QuerySccValues(id, sccLang)
	if err != nil && !db.IsMissingTable(err) {
		return err
	}

	detail := sccDetail{
		ID: header.ID,
		Header: sccHeaderView{
			Group:  header.Group,
			Status: typeWithCode(sccStatusLabel(header.Status), header.Status),
			Aux2:   header.Aux2,
			Aux3:   header.Aux3,
			Name:   header.Name,
			Desc:   header.Desc,
			Desc2:  header.Desc2,
		},
	}
	for _, v := range values {
		detail.Values = append(detail.Values, sccValueView{
			Val:  v.Val,
			Desc: v.Desc,
			Sort: v.Sort,
			Cust: custLabel(v.Cust),
			B3:   v.B3, B4: v.B4, B5: v.B5, B6: v.B6, B7: v.B7, B8: v.B8,
			B9: v.B9, B10: v.B10, B11: v.B11, B14: v.B14, B15: v.B15,
		})
	}

	if Format() == output.FormatTable {
		printSccDetail(&detail)
		return nil
	}
	return emitOne(detail)
}

func printSccDetail(d *sccDetail) {
	fmt.Printf("=== %s ===\n", d.ID)
	fmt.Printf("群组:   %s\n", d.Header.Group)
	fmt.Printf("状态:   %s\n", d.Header.Status)
	fmt.Printf("gzca002: %s\n", d.Header.Aux2)
	fmt.Printf("gzca003: %s\n", d.Header.Aux3)
	fmt.Printf("名称:   %s\n", d.Header.Name)
	fmt.Printf("说明:   %s\n", d.Header.Desc)
	fmt.Printf("说明2:  %s\n", d.Header.Desc2)

	if len(d.Values) == 0 {
		fmt.Println("\n分类值: (无)")
	} else {
		fmt.Printf("\n分类值 (%d):\n", len(d.Values))
		headers := []string{"值", "说明", "排序", "标准/客制"}
		var rows [][]string
		for _, v := range d.Values {
			rows = append(rows, []string{v.Val, v.Desc, v.Sort, v.Cust})
		}
		output.PrintTable(headers, rows)

		// 扩展数据列: 仅展示非空值, 避免超宽表格
		var extHeaders = []string{"值", "字段", "值"}
		var extRows [][]string
		for _, v := range d.Values {
			for _, c := range sccExtColumns(v) {
				extRows = append(extRows, []string{v.Val, c[0], c[1]})
			}
		}
		if len(extRows) > 0 {
			fmt.Printf("\n扩展数据列 (gzcb003~gzcb015, 仅非空):\n")
			output.PrintTable(extHeaders, extRows)
		}
	}
	fmt.Println(sccExtNote)
}

// sccExtColumns lists the non-empty generic data columns of a value row.
func sccExtColumns(v sccValueView) [][2]string {
	var cols [][2]string
	for _, c := range [][2]string{
		{"gzcb003", v.B3}, {"gzcb004", v.B4}, {"gzcb005", v.B5}, {"gzcb006", v.B6},
		{"gzcb007", v.B7}, {"gzcb008", v.B8}, {"gzcb009", v.B9}, {"gzcb010", v.B10},
		{"gzcb011", v.B11}, {"gzcb014", v.B14}, {"gzcb015", v.B15},
	} {
		if c[1] != "" {
			cols = append(cols, c)
		}
	}
	return cols
}

// sccStatusLabel maps gzcastus codes to Chinese labels.
func sccStatusLabel(code string) string {
	switch code {
	case "Y":
		return "启用"
	case "N":
		return "停用"
	default:
		return code
	}
}

const sccExtNote = `
注: gzcb003~gzcb011/gzcb014/gzcb015 为按分类码复用的通用数据列, 含义依各 SCC 而定
(如 SCC 4 的 gzcb004/gzcb005 存群组编号范围, SCC 241 的 gzcb003 存格式)。`

func init() {
	sccCmd.Flags().StringVar(&sccKW, "kw", "", "按分类码/名称过滤 (仅列表模式)")
	sccCmd.Flags().StringVar(&sccLang, "lang", "zh_CN", "说明语言别 (gzcal002/gzcbl003, 默认 zh_CN)")
	Group.AddCommand(sccCmd)
}
