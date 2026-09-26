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

本包没有自己的测试文件；行为由字典线的用例覆盖：

```bash
go test ./internal/dict/... -count=1
```

## 细节去哪

- 连接模型从哪来 → [../dbconfig/README.md](../dbconfig/README.md)
- 谁在用它、直查与本地镜像怎么选 → [../dict/live/README.md](../dict/live/README.md)
- 客户端到不了库时怎么走隧道 → [../sshtun/README.md](../sshtun/README.md)
