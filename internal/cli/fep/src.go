package fep

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/fep"

	"github.com/spf13/cobra"
)

// tt fep src openwin —— 从**源码**读出开窗的交互语义（"选中"到底是什么意思）。
//
// 为什么它是"不用人教"的关键一环：同一个界面语言、同一个"确定"按钮，两个开窗的
// "选中"语义可以完全不同 ——
//
//	q_pmaa001_3（供应商）：accept 里 `IF li_ac > 0 THEN return1 = gr_qry[li_ac].pmaa001_1`
//	                      → 选中 = **当前行**（光标移到那一行再确定）
//	q_imaf001_15（料件）：accept 里 `CALL ..._get_multiret()`，函数体读 `check = "Y"`
//	                      → 选中 = **勾选**（给那一行的勾选格写 Y 再确定）
//
// 后者靠试是试不出来的：不勾选直接确定，运行期**只回一个空 ACK**、界面什么都不返回，
// 看起来像"形态不对"，其实是"你没勾选"。所以这条命令的意义是：**别试，去读**。

var (
	srcOpenWinProgram    string
	srcOpenWinSourceFlag string
)

// srcCmd 是"读源码"这一组：从源码镜像里读出语义（不是从运行期猜）。
var srcCmd = &cobra.Command{
	Use:   "src",
	Short: "读源码：开窗的交互语义、字段类型与必填（T100 的 .4gl/.4fd）",
	Long: `从**源码镜像**读语义 —— 这是"不用人教"的那一半：

  - 开窗"选中"到底是当前行还是勾选？→ ON ACTION accept 段 + 它调的函数体
  - 查询条件写哪几列？→ CONSTRUCT … ON <列,列,列>
  - 勾选栏的"勾上"写什么值？→ .4fd 里 CheckBox 的 valueChecked

镜像根取自配置的 mirror.dir（tt dict mirror dir 设置），可用 --source 覆盖。`,
}

// srcOpenWinSource 解析源码镜像根：--source > 环境变量 > 配置的 mirror.dir。
func srcOpenWinSource() string {
	if srcOpenWinSourceFlag != "" {
		return srcOpenWinSourceFlag
	}
	if v := os.Getenv("TT_SOURCE_ROOT"); v != "" {
		return v
	}
	path, err := common.ResolveConfig(true)
	if err != nil {
		return ""
	}
	rt, err := config.Load(path)
	if err != nil || rt.Mirror.Dir == "" {
		return ""
	}
	if abs, err := filepath.Abs(rt.Mirror.Dir); err == nil {
		return abs
	}
	return rt.Mirror.Dir
}

var srcOpenWinCmd = &cobra.Command{
	Use:   "openwin <开窗程序名，如 q_imaf001_15>",
	Short: "读开窗源码：查询条件列 / 结果对话框 / 选中语义（当前行 or 勾选）/ 回传字段",
	Long: `读一个开窗程序（com/qry/4gl/<程序>.4gl + 同名 .4fd），输出它的交互语义：

  - **查询条件列**：CONSTRUCT … ON <列,列,列>（前端要往这些列写条件）
  - **结果对话框**：INPUT ARRAY / DISPLAY ARRAY + 数组名
  - **选中语义**（重点）：accept 段与它调用的函数体里读出来的结论
      current-row : 用 li_ac 这类行下标 → 驱动要把光标移到那一行再确定
      checked     : 读 check = valueChecked → 驱动要给那一行的勾选格写 Y 再确定
      unknown     : 读不出判据 —— **不许猜**（打印出来，让人或源码再确认）
  - **回传字段**：写回调用方字段的那个值（如 imaf001_1）
  - **勾选栏字面值**：.4fd 里 CheckBox 的 valueChecked/valueUnchecked（"勾上"写什么）

每条结论都带出处（文件:行 + 原文），没有出处的结论不会出现在输出里。

源码镜像根用 --source；不给就用配置里的镜像根（tt dict mirror path 能看当前值）。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := srcOpenWinSource()
		if root == "" {
			return fail(exitUsage, "没给源码镜像根：加 --source <镜像根>，"+
				"或用 tt dict mirror dir <目录> 配好，或设 TT_SOURCE_ROOT")
		}
		ow, err := fep.LoadOpenWin(root, args[0])
		if err != nil {
			return fail(exitTarget, "%v", err)
		}
		return printJSONOrText(ow, func() {
			fmt.Printf("开窗 %s   （源码：%s）\n", ow.Program, strings.Join(ow.Files, ", "))
			if len(ow.Forms) > 0 {
				fmt.Printf("  表单：%s\n", strings.Join(ow.Forms, ", "))
			}
			if len(ow.ConstructCols) > 0 {
				fmt.Printf("  查询条件列：%s\n", strings.Join(ow.ConstructCols, ", "))
			}
			if len(ow.States) > 0 {
				fmt.Printf("  g_qryparam.state 出现过的取值：%s（调用方决定用哪种模式）\n",
					strings.Join(ow.States, ", "))
			}
			if len(ow.Dialogs) == 0 {
				fmt.Println("  没读到结果对话框（DIALOG 段）：回源码看")
			}
			for i, d := range ow.Dialogs {
				head := fmt.Sprintf("【对话框 %d】%s", i+1, d.Dialog)
				if d.Array != "" {
					head += fmt.Sprintf("（数组 %s", d.Array)
					if d.ScreenArray != "" {
						head += " ← 屏幕数组 " + d.ScreenArray
					}
					head += "）"
				}
				fmt.Println(head)
				if len(d.AcceptCalls) > 0 {
					fmt.Printf("    确定（accept）调用：%s\n", strings.Join(d.AcceptCalls, ", "))
				}
				switch d.Select {
				case fep.SelectCurrentRow:
					fmt.Printf("    选中语义：**当前行**（行下标 %s）—— 把光标移到目标行再点确定\n", d.RowVar)
				case fep.SelectChecked:
					fmt.Printf("    选中语义：**勾选** —— 给目标行的勾选栏 %s 写 %q 再点确定（取消值 %q）\n",
						d.CheckColumn, d.ValueChecked, d.ValueUnchecked)
				default:
					fmt.Println("    选中语义：**读不出来** —— 不许猜：回源码看 ON ACTION accept 段")
				}
				if len(d.ReturnFields) > 0 {
					fmt.Printf("    回传字段：%s\n", strings.Join(d.ReturnFields, ", "))
				}
				for _, e := range d.Evidence {
					fmt.Printf("      - %s\n", e)
				}
			}
			if len(ow.Dialogs) > 1 {
				fmt.Println("  注意：这个开窗有多个对话框、语义不同 —— 驱动要按当前那个对话框的结论走")
			}
		})
	},
}

func init() {
	srcOpenWinCmd.Flags().StringVar(&srcOpenWinSourceFlag, "source", "",
		"源码镜像根目录（如 D:\\T100\\源代码\\主机109）；不给就用配置的 mirror.dir")
	srcCmd.AddCommand(srcOpenWinCmd)
}
