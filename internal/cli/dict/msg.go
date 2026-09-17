package dict

// tt dict msg:查询系统消息档(gzze_t,由作业 azzi920 维护;运行时 cl_err/cl_getmsg
// 按编号+语言取用)。AI 在代码里遇到消息编号(如 std-00006 / azz-00041 / -263)
// 时用它查文本、建议处理方式与技术细节。
//
// 数据源与 r.t/r.v 等一致(本地 SQLite 镜像 / --env <环境> 远程直查);
// 语言策略:默认只显示 --lang(缺省 zh_CN)那一行,该编号无此语言行时列出
// 可用语言引导重查(与源系统一致:精确 (code,lang) 匹配,无自动回退)。

import (
	"fmt"
	"strings"

	"tt/internal/dict/db"
	"tt/internal/output"

	"github.com/spf13/cobra"
)

var msgLang string

var msgCmd = &cobra.Command{
	Use:   "msg <编号>[,<编号>...]",
	Short: "查询系统消息",
	Long: `查询系统消息:所有提示/报错消息的记录。程序报错/日志里出现编号(如
std-00006)时,查它的完整文本、建议处理方式与技术细节。
消息编号形如 std-00001 / azz-00041 / lib-xxxxx(类型-流水);负整数为
SQLCODE(如 -263,查询时需用 -- 分隔: tt dict msg -- -263)。
返回: 文本、建议处理、建议作业(附名称)、技术细节、类型(0=警告 1=错误
2=资讯)、状态。
消息按语言各一条:默认只显示 --lang(缺省 zh_CN)那一行;该编号无此语言时
列出可用语言。支持逗号分隔多编号。`,
	Example: `  tt dict msg std-00006
  tt dict msg azz-00041 --lang zh_TW
  tt dict msg "std-00006,azz-00041" --json
  tt dict msg -- -100                     # 负整数编号(SQLCODE)需用 -- 分隔
  tt dict msg aoo-00120 --env 正式区   # 远程直查该环境`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		codes := splitNames(args[0])
		if len(codes) == 0 {
			return fmt.Errorf("未指定有效的消息编号")
		}

		var picked []db.MsgRow
		for _, code := range codes {
			rows, err := GetDB().QueryMsg(code)
			if err != nil {
				if db.IsMissingTable(err) {
					fmt.Println(missingHint("消息档 (gzze_t/gzzal_t)"))
					return nil
				}
				return err
			}
			if len(rows) == 0 {
				if IsJSON() {
					continue
				}
				fmt.Printf("未找到消息编号 '%s'。\n", code)
				continue
			}
			// 语言策略:只取 --lang 行;无该语言行时提示可用语言
			match, avail := pickLangRow(rows, msgLang, func(r db.MsgRow) string { return r.Lang })
			if match == nil {
				if IsJSON() {
					continue
				}
				fmt.Printf("%s\n", langMissMsg(code, msgLang, avail))
				continue
			}
			picked = append(picked, *match)
			if !IsJSON() {
				printMsg(match)
			}
		}

		if IsJSON() {
			return output.PrintJSON(picked)
		}
		return nil
	},
}

// printMsg 文本模式输出单条消息。
func printMsg(m *db.MsgRow) {
	fmt.Printf("=== %s (%s) ===\n", m.Code, m.Lang)
	fmt.Printf("类型:     %s\n", typeWithCode(msgTypeLabel(m.TypeCode), m.TypeCode))
	fmt.Printf("状态:     %s\n", statusLabel(m.Status))
	fmt.Printf("文本:     %s\n", m.Text)
	if m.Action != "" {
		fmt.Printf("建议处理: %s\n", m.Action)
	}
	if m.ExecProg != "" && m.ExecProg != ":EXEPROG" {
		line := "建议作业: " + m.ExecProg
		if m.ProgName != "" {
			line += " (" + m.ProgName + ")"
		}
		fmt.Println(line)
	}
	if m.Detail != "" {
		fmt.Println("技术细节 (gzze006):")
		for _, line := range strings.Split(m.Detail, "\n") {
			fmt.Println("  " + line)
		}
	}
	fmt.Println()
}

// msgTypeLabel 讯息类型 (SCC 106): 0警告/1错误/2资讯。
func msgTypeLabel(code string) string {
	switch code {
	case "0":
		return "警告"
	case "1":
		return "错误"
	case "2":
		return "资讯"
	default:
		return code
	}
}

// statusLabel 状态 Y/N → 启用/停用。
func statusLabel(code string) string {
	switch code {
	case "Y":
		return "启用"
	case "N":
		return "停用"
	default:
		return code
	}
}

func init() {
	msgCmd.Flags().StringVar(&msgLang, "lang", "zh_CN", "显示语言别 (gzze002;缺省 zh_CN)")
	// 负整数编号(SQLCODE,如 -263)会被 flag 解析吃掉,报 "unknown shorthand flag: '2' in -263",
	// 看不出该怎么办。这里翻成可操作的提示(唯一需要 -- 分隔的参数位置)。
	msgCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		if strings.Contains(err.Error(), "unknown shorthand flag") {
			return fmt.Errorf("%w\n提示: 负整数编号(SQLCODE)要写成 `tt dict msg -- -263`(`--` 之后才当参数解析)", err)
		}
		return err
	})
	Group.AddCommand(msgCmd)
}
