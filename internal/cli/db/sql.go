package db

// tt db sql —— 只读 SQL（业务数据）。从 tt debug sql 抽出来，并且**不再依赖 serve**：
// 本机 → SSH → 服务器上的 sqlplus/ksql。
//
// 账号由配置的 ENT 清单决定（cfg.SSHs[].ents），所以同一句 --at 在 db / dict 两条线上
// 解析出同一个 schema；清单没同步就报“没同步”，不静默查 gzou_t 再猜一个答案。

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/dbconfig"
	"tt/internal/debug"
	"tt/internal/host"
	"tt/internal/output"
	"tt/internal/safesql"
)

func newSQLCmd() *cobra.Command {
	var (
		file    string
		entFlag int
		timeout int
	)
	cmd := &cobra.Command{
		Use:   `sql "<查询语句>"`,
		Short: "执行一条只读 SQL（单条 SELECT/WITH；只读事务；最多 200 行）",
		Long: `对指定环境的库执行一条**只读** SQL，用来回答"这条业务数据到底有没有/是什么"。

  tt db sql --at 主机109@99 "select bmaa001,bmaastus from bmaa_t where bmaa001='FCPU010100003'"
  tt db sql --at 主机109@99 --file q.sql          # 长语句从本地文件读
  tt db sql --at 主机109@99 --format table "…"    # 人读表格（默认 JSON 信封）

**默认输出是 JSON 信封**：环境/ENT/账号/库地址平铺在顶层，载荷在 data。
那几项必须看 —— 企业编号错了就是另一个 schema，拿到 0 行，而那看起来和
"数据不存在"一模一样。

--format csv 给"注释头 + 纯净 CSV"：注释行全以「# 」开头且全在头部，
grep -v '^# ' 切下去就是一个标准 CSV（逗号/引号/换行按 RFC4180 转义）。

列名从 SQL*Plus 的 CSV 元数据取，保留完整别名；旧版 sqlplus 不支持 CSV 模式时会明确报错，
不会把截断的文本表头当成列名返回。

限制（都是刻意的）：
  - 只允许**单条** SELECT / WITH；写操作、DDL、PL/SQL 块、多语句、sqlplus 命令一律拒绝；
  - 库会话是只读事务，常规写会被库自己挡回去；
  - **最多回 200 行**，totalRows 是下界（库侧故意多取一行判"还有更多"）；要更多请加 WHERE 收窄。

不需要 tt serve 在跑；环境/企业用 --at 点名（不知道填什么：tt db sql --at-list）。`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sqlText := ""
			if len(args) == 1 {
				sqlText = strings.TrimSpace(args[0])
			}
			if file != "" {
				if sqlText != "" {
					return fmt.Errorf("语句既给了位置参数又给了 --file，二选一")
				}
				b, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("读取语句文件失败: %s: %w", file, err)
				}
				sqlText = strings.TrimSpace(string(b))
			}
			if sqlText == "" {
				return fmt.Errorf(`需要一条查询语句（见 tt db sql --help），或用 --file 指定语句文件`)
			}
			return QueryOnce(sqlText, entOf(entFlag), timeout)
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "从本地文件读语句（与位置参数二选一）")
	cmd.Flags().IntVar(&entFlag, "ent", 0, "企业编号（也可用 --at <环境>@<企业>）")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "查询超时秒数（默认 30，上限 120）")
	return cmd
}

// QueryOnce 执行一次只读 SQL 并渲染（tt db sql 与旧的 tt debug sql 转发共用）。
func QueryOnce(sqlText string, ent, timeout int) error {
	if ent <= 0 {
		return output.Errorf("USAGE", 2, "需要 --at <环境>@<企业>：企业编号决定用哪个账号连库（tt db sql --at-list 可看取值）")
	}
	// 只读闸门先行：拒绝要退 3（“这句不让查”），与“连不上/查出错”（4）分开。
	if err := safesql.Check(sqlText); err != nil {
		return output.Errorf("SQL_REJECTED", 3, "%v", err)
	}
	cfg, _, err := loadCfg()
	if err != nil {
		return output.Errorf("TARGET", 2, "%v", err)
	}
	// 企业必须在配置的 ENT 清单里：这一步失败是“身份填错”（2），不是“库连不上”。
	if _, err := entAccount(cfg, ent); err != nil {
		return output.Errorf("TARGET", 2, "%v", err)
	}
	conn, err := host.Dial(cfg.SSH)
	if err != nil {
		return output.Errorf("SSH_FAILED", 4, "SSH 连接失败: %v", err)
	}
	defer conn.Close()

	m := debug.NewManager(cfg)
	res, err := m.RunReadonlyQuery(conn, debug.ReadonlyQueryReq{SQL: sqlText, Ent: ent, Timeout: timeout})
	if err != nil {
		return output.Errorf("DB_FAILED", 4, "%v", err)
	}
	return printSQLResult(res)
}

func printSQLResult(res *debug.DBQueryResult) error {
	return writeSQLResult(os.Stdout, res)
}

// writeSQLResult 把一次只读查询写成三种形态之一（w 可换成 buffer 做测试）。
func writeSQLResult(w io.Writer, res *debug.DBQueryResult) error {
	route := dbconfig.RouteServerOracle
	if res.Dialect == "kingbase" {
		route = dbconfig.RouteServerKB
	}
	src := res.AccountSource
	if src == "" {
		src = dbconfig.AcctFromEnt
	}
	meta := output.Meta{
		Source: "live", Env: res.Env, SSHHost: res.SSHHost, Zone: res.Zone,
		Ent: res.Ent, Account: res.Account, AccountSource: src,
		Target: res.DBTarget, Dialect: res.Dialect, Route: route, Readonly: true,
		TotalRows: res.TotalRows, Returned: len(res.Rows),
		Truncated: res.Truncated, ServerLimited: res.ServerLimited,
		Elapsed: res.Elapsed, Notes: res.Notes,
	}
	switch common.OutputFormat() {
	case output.FormatCSV:
		return output.WriteCSV(w, meta, res.Columns, res.Rows)
	case output.FormatTable:
		if err := output.WriteMeta(w, meta); err != nil {
			return err
		}
		return output.WriteTable(w, res.Columns, res.Rows)
	default:
		return output.Emit(w, output.Options{
			Format: output.FormatJSON, Meta: meta,
			Data: map[string]any{"columns": res.Columns, "rows": res.Rows},
		})
	}
}
