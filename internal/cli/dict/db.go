package dict

import (
	"fmt"
	"os"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/dbconfig"
	"tt/internal/host"

	"github.com/spf13/cobra"
)

var (
	dbCfg *host.Hosts
)

// dbCmd 管理 ERP 数据库连接(每个 SSH 环境一对一挂载的 db)与同步/远程直查。
//
// It defines its own PersistentPreRunE, which overrides the group's query-source
// hook (cobra runs the first persistent pre-run found walking leaf→root),
// so `tt dict db *` does not depend on the local erp_data.db.
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "管理 ERP 数据库连接与数据同步",
	Long: `从 config.json 读取配置 (--config / TT_CONFIG / TDICT_CONFIG)。数据库连接按环境一对一挂载:
hosts.sshs[].db 即该环境的库(显式 host/port/service|库名 + 账号列表)。
子命令: sync (从 ERP 拉取字典数据写入 SQLite) / list (列各环境的库) /
ping (验证连接可达) / discover (SSH 自动发现连接要素并写入环境 db)。
支持连接类型: kingbase (金仓, PostgreSQL 协议)、oracle (go-ora)。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := skipQuerySource(cmd, args); err != nil {
			return err
		}
		path, err := common.ResolveConfig(false)
		if err != nil {
			return err
		}
		if common.Verbose {
			fmt.Fprintf(os.Stderr, "[dict] 配置文件: %s\n", path)
		}
		cfg, err := config.LoadHosts(path)
		if err != nil {
			return err
		}
		dbCfg = cfg
		return nil
	},
}

func init() {
	Group.AddCommand(dbCmd)
}

// resolveDbConn 按环境名解析其 db(--env <环境名>);name 为空取活跃环境(activeEnv,
// 未设置自动取首条)。返回该连接的深拷贝与所属环境名。
func resolveDbConn(name string) (*dbconfig.Connection, string, error) {
	if dbCfg == nil {
		return nil, "", fmt.Errorf("配置未加载")
	}
	target := name
	if target == "" {
		target = dbCfg.ActiveEnv
	}
	if target == "" && len(dbCfg.SSHs) > 0 {
		target = dbCfg.SSHs[0].Name
	}
	if target == "" {
		return nil, "", fmt.Errorf("尚未配置 SSH 环境(运行 tt serve 添加,或编辑 config.json hosts.sshs)")
	}
	for i := range dbCfg.SSHs {
		e := &dbCfg.SSHs[i]
		if e.Name != target {
			continue
		}
		if e.DB == nil {
			return nil, target, fmt.Errorf("环境 %q 未配置数据库(设置-环境-数据库页添加)", target)
		}
		cc := *e.DB
		cc.Accounts = append([]dbconfig.DBAcct(nil), e.DB.Accounts...)
		return &cc, target, nil
	}
	return nil, name, fmt.Errorf("未找到环境 %q(可用: tt env list 查看)", name)
}
