# internal/erpdb — 客户端直连库

客户端驱动的封装：一个接口、两个实现、一个按库类型分派的工厂。

| 实现 | 驱动 | 用于 |
|---|---|---|
| 金仓 | `pgx`（PostgreSQL 协议） | `type: kingbase` |
| Oracle | `go-ora`（纯 Go，无 cgo） | `type: oracle` |

## 契约

- `Connector` 只暴露四件事：类型、服务端版本、执行一条查询、关闭。
- 查询返回"列名 + 行"，**方言差异在实现里消化**（不要漏到调用方）。
- **没有绑定参数**：所有值经标识符白名单与字面量引号处理后内联。所以拼进语句的每一处
  都必须过 `ValidIdent` / `QuoteIdent` / `QuoteLit` —— 这是本包唯一的注入面。
- 执行前过一遍只读语句检查；配合库侧只读事务构成两层防线。

## 入口

| 入口 | 作用 |
|---|---|
| `Open(ctx, conn)` | 按库类型分派出实现 |
| `Connector` | 接口（类型 / 版本 / 查询 / 关闭） |
| `SelectAllSQL(conn, table)` | 取全表的语句（表名过白名单） |
| `ValidIdent` / `QuoteIdent` / `QuoteLit` | 注入防线的三个原语 |

## 判据

```bash
go test ./internal/erpdb -count=1
```

**覆盖**的是**转义与闸门**那一层（安全关键，所以判据写得比"能跑"严）：
标识符白名单 `ValidIdent`（拒绝空格 / 分号 / 引号 / 点号 / 非 ASCII / 数字开头；
**不过滤保留字** —— 那是库的事）、`QuoteLit` 的单引号翻倍（含一组"想破墙"的载荷）、
`SelectAllSQL` 对用户名与表名的双重校验、`checkReadOnlySQL` 的**强度与边界**、
两个 DSN 的拼法（oracle 走 URL 形式要转义 `% : @ / ?`，金仓走 key=value）、
`scanString` 把驱动给的任意值文本化。

**故意不覆盖**：真连库（`Open` / `OpenOracle` / `OpenKingbase` / 各连接器的 `Query`）——
属 L4，要真 Oracle 或金仓。

一处**要读清的**：`checkReadOnlySQL` 是**前缀检查，不是解析器** —— 它挡的是"手写的
DML 直接贴进来"这种最常见的情况，**挡不住**前面加注释的写法。它不是唯一的闸门：
真正的文本闸门是 `internal/safesql`（会先剥注释），而 `internal/dict/live` 那条路径
自己构造 SQL（值走 `QuoteLit`、表名走 `ValidIdent`）。测试里把它的边界一并钉住了，
免得有人把它当"完善"来用。

## 细节去哪

- 连接模型从哪来 → [../dbconfig/README.md](../dbconfig/README.md)
- 谁在用它、直查与本地镜像怎么选 → [../dict/live/README.md](../dict/live/README.md)
- 客户端到不了库时怎么走隧道 → [../sshtun/README.md](../sshtun/README.md)
