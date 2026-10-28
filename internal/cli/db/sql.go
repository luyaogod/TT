package db

// tt db sql —— 只读 SQL（业务数据）。从 tt debug sql 抽出来，并且**不再依赖 serve**：
// 本机 → SSH → 服务器上的 sqlplus/ksql。
//
// 账号由配置的 ENT 清单决定（cfg.SSHs[].ents），所以同一句 --at 在 db / dict 两条线上
// 解析出同一个 schema；清单没同步就报“没同步”，不静默查 gzou_t 再猜一个答案。

import (
	"context"
	"errors"
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

// `tt db sql` 的三条执行路线（--route）。
const (
	// SQLRouteAuto 直连优先，建连失败回退服务器 sqlplus（默认）。
	SQLRouteAuto = "auto"
	// SQLRouteDirect 只走客户端直连；建连失败直接报错，不回退。
	SQLRouteDirect = "direct"
	// SQLRouteServer 只走服务器 sqlplus/ksql（旧行为）。
	SQLRouteServer = "server"
)

// sqlRouteValid 校验 --route 取值。
func sqlRouteValid(route string) bool {
	return route == SQLRouteAuto || route == SQLRouteDirect || route == SQLRouteServer
}

// normalizeSQLRoute 空路线一律按 server。
//
// 这条归一化是**兼容闸门**：`tt db sql` 自己的默认值是 auto，而隐藏的
// `tt debug sql` 转发不带路线 —— 它必须继续走服务器 sqlplus，否则一个只改默认值的
// 重构就会把旧命令的执行方式（以及取值渲染）悄悄换掉。
func normalizeSQLRoute(route string) string {
	if route == "" {
		return SQLRouteServer
	}
	return route
}

func newSQLCmd() *cobra.Command {
	var (
		file    string
		entFlag int
		timeout int
		route   string
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

# 执行路线（--route）

  auto    直连优先：从本机直连该企业的库；建不起来（网络/隧道/登录）才回退服务器
          sqlplus。默认。回退不是静默的：信封 route=server-sqlplus，notes 里写明原因。
  direct  只走客户端直连。建不起来就报 DB_UNREACHABLE，不回退（要确定"本机通不通"时用）。
  server  只走服务器 sqlplus/ksql（旧行为；每条 SQL 一次进程冷启动，实测约 +1.0s）。

两条路线查的是**同一个库、同一个账号**（企业编号→账号的解析只有一处），信封的
route 字段说明这次实际走的是哪条。**值的渲染有差别**：直连的 DATE/TIMESTAMP 输出为
2006-01-02 15:04:05，sqlplus 路线是服务器 NLS 格式（如 28-OCT-26）。需要与既有
报表口径逐字对齐时用 --route server。

--format csv 给"注释头 + 纯净 CSV"：注释行全以「# 」开头且全在头部，
grep -v '^# ' 切下去就是一个标准 CSV（逗号/引号/换行按 RFC4180 转义）。

列名从 SQL*Plus 的 CSV 元数据取，保留完整别名；旧版 sqlplus 不支持 CSV 模式时会明确报错，
不会把截断的文本表头当成列名返回。

限制（都是刻意的，两条路线一致）：
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
			return QueryOnce(sqlText, entOf(entFlag), timeout, route)
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "从本地文件读语句（与位置参数二选一）")
	cmd.Flags().IntVar(&entFlag, "ent", 0, "企业编号（也可用 --at <环境>@<企业>）")
	cmd.Flags().IntVar(&timeout, "timeout", 0, "查询超时秒数（默认 30，上限 120）")
	cmd.Flags().StringVar(&route, "route", SQLRouteAuto,
		"执行路线: auto(直连优先,建连失败回退) | direct(只直连) | server(服务器 sqlplus)")
	return cmd
}

// QueryOnce 执行一次只读 SQL 并渲染（tt db sql 与旧的 tt debug sql 转发共用）。
//
// route 为空按 server 处理：旧调用方（tt debug sql）走的还是服务器 sqlplus 那一条，
// 不能让一个隐藏的兼容命令悄悄换掉执行方式与取值格式。
func QueryOnce(sqlText string, ent, timeout int, route string) error {
	if ent <= 0 {
		return output.Errorf("USAGE", 2, "需要 --at <环境>@<企业>：企业编号决定用哪个账号连库（tt db sql --at-list 可看取值）")
	}
	route = normalizeSQLRoute(route)
	if !sqlRouteValid(route) {
		return output.Errorf("USAGE", 2, "--route 只接受 auto | direct | server（收到 %q）", route)
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
	if cfg.DB == nil {
		return output.Errorf("DB_FAILED", 4, "当前环境未挂数据库连接(设置-环境-SSH 页选择「数据库连接」)")
	}
	if !cfg.DB.ReadonlySQLEnabled() {
		return output.Errorf("DB_FAILED", 4, "该环境的只读 SQL 已在配置里关闭(debug.sshs[].db.readonlySql=false)")
	}

	fallbackNote := ""
	if route != SQLRouteServer {
		err := queryDirect(sqlText, ent, cfg, timeout, route)
		if err == nil {
			return nil // 已渲染
		}
		// 只有“直连建不起来”才允许回退：查询本身的失败回退没有意义
		// （同一条语句在另一条路上是同一个错误，而重查询还要再付一次超时）。
		var unreachable *debug.DirectUnreachableError
		if route == SQLRouteDirect || !errors.As(err, &unreachable) {
			code := "DB_FAILED"
			if errors.As(err, &unreachable) {
				code = "DB_UNREACHABLE"
			}
			return output.Errorf(code, 4, "%v", err)
		}
		fallbackNote = fmt.Sprintf("%v；已回退服务器 sqlplus 路线", err)
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
	res.Route = serverRoute(res.Dialect)
	if fallbackNote != "" {
		res.Notes = append(res.Notes, fallbackNote)
	}
	return printSQLResult(res)
}

// serverRoute 服务器侧路线按方言命名（sqlplus / ksql）。
func serverRoute(dialect string) string {
	if dialect == "kingbase" {
		return dbconfig.RouteServerKB
	}
	return dbconfig.RouteServerOracle
}

// queryDirect 走客户端直连并渲染；返回 nil 表示已经出结果。
//
// 回退与否由调用方判（见 QueryOnce）：这里把错误原样带出去，不吞。
func queryDirect(sqlText string, ent int, cfg *debug.Config, timeout int, route string) error {
	cc, err := dbConnFor(cfg, ent)
	if err != nil {
		return err // 身份错：不是“连不上”，交给调用方按 DB_FAILED 报，不触发回退
	}
	res, err := debug.RunDirectQuery(context.Background(), *cc, sqlText, timeout)
	if err != nil {
		return err
	}
	res.Env = cfg.EnvName()
	res.Zone = cfg.Zone
	res.Ent = ent
	res.Account = cc.Account
	res.AccountSource = dbconfig.AcctFromEntConfig
	res.DBTarget = cc.Address()
	res.Route = dbconfig.RouteClientDirect
	res.ViaTunnel = cc.ViaSSH != nil
	return printSQLResult(res)
}


func printSQLResult(res *debug.DBQueryResult) error {
	return writeSQLResult(os.Stdout, res)
}

// writeSQLResult 把一次只读查询写成三种形态之一（w 可换成 buffer 做测试）。
func writeSQLResult(w io.Writer, res *debug.DBQueryResult) error {
	route := res.Route
	if route == "" { // 老调用方（tt debug sql 的转发、旧测试）不带 Route：沿用按方言推
		route = serverRoute(res.Dialect)
	}
	src := res.AccountSource
	if src == "" {
		src = dbconfig.AcctFromEnt
	}
	meta := output.Meta{
		Source: "live", Env: res.Env, SSHHost: res.SSHHost, Zone: res.Zone,
		Ent: res.Ent, Account: res.Account, AccountSource: src,
		Target: res.DBTarget, Dialect: res.Dialect, Route: route, Readonly: true,
		ViaTunnel: res.ViaTunnel,
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
