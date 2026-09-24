package dict

import (
	"fmt"
	"strings"

	"tt/internal/dict/dbsync"
	"tt/internal/output"

	"github.com/spf13/cobra"
)

var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看本地查询数据的覆盖情况(缺什么、各多少行)",
	Long: `查看本地 SQLite 里各数据族(表字典/校验带值/系统分类码/字段画面规格/可复用开窗/
系统消息/参数定义/程序与作业)是否已同步、各多少行、缺哪些表。

查询命令报「本地库尚未包含 XXX 数据」时用它定位;补齐用 tt dict db sync。
看的是 -d/--db 或 TDICT_DB 指向的库(与查询命令同一份);库不存在只报状态、不报错。
只读,不改任何文件;也不需要 config.json(还没配环境也能看)。`,
	Example: `  tt dict db status
  tt dict db status -d D:\data\erp_data.db
  tt dict db status --json`,
	Args: cobra.NoArgs,
	// status 不依赖 config.json(与 db 组其他子命令不同):自带空的 pre-run 覆盖父命令的
	PersistentPreRunE: skipQuerySource,
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := dbsync.InspectLocal(resolveSyncTarget())
		if err != nil {
			return err
		}
		if Format() != output.FormatTable {
			return emitOne(st)
		}
		printLocalStatus(st)
		return nil
	},
}

func printLocalStatus(st *dbsync.LocalStatus) {
	if !st.Exists {
		fmt.Printf("本地库不存在: %s\n", st.Path)
		fmt.Println("建立本地库: tt dict db sync(需能连 ERP);或查询时用 --env <环境名> 远程直查,不必同步。")
		return
	}
	fmt.Printf("本地库: %s  (%s, %s)\n\n", st.Path, humanSize(st.Size), st.Modified.Format("2006-01-02 15:04"))

	headers := []string{"状态", "数据族", "行数", "缺表", "依赖命令"}
	rows := make([][]string, 0, len(st.Families))
	missingAny := false
	for _, f := range st.Families {
		mark, missing, cnt := "✓", "-", fmt.Sprintf("%d", f.Rows)
		if !f.Complete {
			missingAny = true
			missing = strings.Join(f.Missing, ",")
			// 表不全时行数没有可解释的语义(命令会整体不可用),显示 — 而不是一个会误导人的数字
			cnt = "—"
			mark = "✗"
			if f.Present > 0 {
				mark = "✗ 部分"
			}
		}
		rows = append(rows, []string{mark, f.Name, cnt, missing, strings.Join(f.Commands, ",")})
	}
	output.PrintTable(headers, rows)

	if missingAny {
		fmt.Println("\n✗ = 该族的字典表不全,依赖它的命令会**整体**报错(不是部分可用);行数显示 — 表示该族不全、行数无意义。")
		fmt.Println("补齐: tt dict db sync(从 ERP 全量刷新,需能连 ERP);临时也可 `tt dict <命令> --env <环境名>` 远程直查。")
	}
}

// humanSize 把字节数写成人类可读(与 mirror 进度里的写法一致)。
func humanSize(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.2f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func init() {
	dbCmd.AddCommand(dbStatusCmd)
}
