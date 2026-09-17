package dict

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"tt/internal/cli/common"
	"tt/internal/dict/dbsync"

	"github.com/spf13/cobra"
)

var (
	dbSyncTable string
)

var dbSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "从 ERP 刷新本地查询数据",
	Long: `从 ERP 刷新本地查询数据(表字典、校验、分类码、画面规格、开窗、消息、参数、
程序与作业等全部内容,约 86 万行),查询命令读的就是它。
写入 -d/--db 或 TDICT_DB 指向的数据库(默认 ./erp_data.db);原库自动备份为 .bak。
--table 可只刷部分;缺省数据源 = 默认环境(activeEnv,tt env list 查看 / tt env use 切换)的库,
可用 --env <环境名> 指定。
各数据族与本地覆盖情况见 tt dict db status。
也可在 tt serve 的「数据同步」页面里选环境执行(带进度)。`,
	Example: `  tt dict db sync
  tt dict db sync --table dzea_t,dzeal_t
  tt dict db sync --env 正式区`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, envName, err := resolveDbConn(common.Env)
		if err != nil {
			return err
		}
		tables := dbsync.DictTables
		if dbSyncTable != "" {
			tables = splitNames(dbSyncTable)
			if len(tables) == 0 {
				return fmt.Errorf("--table 未指定有效表名")
			}
		}

		// 目标 SQLite:优先已存在的库(与查询命令读的是同一个),不存在再按 TDICT_DB/exe 目录/当前目录创建
		target := resolveSyncTarget()

		fmt.Printf("正在从 ERP 拉取字典数据 (环境: %s) -> %s\n", envName, target)
		st, err := dbsync.Run(context.Background(), *conn, target, tables, func(p dbsync.Progress) {
			// 每张表完成时打印一行(开始回调 TableDur=0,跳过)
			if p.Phase == "table" && p.TableDur > 0 {
				fmt.Printf("  %-10s %8d 行 (%v)\n", p.Table, p.TableRows, p.TableDur.Round(time.Millisecond))
			}
		})
		if err != nil {
			return err
		}
		for _, w := range st.Warnings {
			fmt.Println(w)
		}
		fmt.Printf("同步完成: %d 张表, 共 %d 行\n", st.Tables, st.Rows)
		fmt.Printf("数据库已更新: %s\n", st.Target)
		if st.Backup != "" {
			fmt.Printf("原数据库已备份到: %s\n", st.Backup)
		}
		return nil
	},
}

// resolveSyncTarget 解析数据同步(写库)的目标路径。便携优先:
//  1. 已存在的库(TDICT_DB > -d 绝对 > exe 同目录 > 当前目录 中先找到的那个)——与查询命令读同一个;
//  2. -d 绝对路径(显式指定);
//  3. 都不存在时用 exe 同目录(便携版自带位置,分发到任何机器都成立);
//  4. 再退当前目录。
//
// 注意:不再无条件采用 TDICT_DB —— 否则宿主机上遗留的旧环境变量会把便携版的写入目标带偏。
// 需要固定/共享位置时,用 config.json 顶层 sync.target(设置页可改)或 -d 绝对路径。
func resolveSyncTarget() string {
	if p, err := resolveDBPath(dbPath); err == nil {
		return p
	}
	if filepath.IsAbs(dbPath) {
		return dbPath
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), dbPath)
	}
	if abs, err := filepath.Abs(dbPath); err == nil {
		return abs
	}
	return dbPath
}

func init() {
	dbSyncCmd.Flags().StringVar(&dbSyncTable, "table", "", "仅同步指定的表 (逗号分隔, 默认全部字典表)")
	dbCmd.AddCommand(dbSyncCmd)
}
