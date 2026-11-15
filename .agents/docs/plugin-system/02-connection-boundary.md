# 连接边界

## 归属

| TT 管 | TT 不管 |
|---|---|
| 哪个调用者可以使用哪个环境 | 插件拿连接做什么业务 |
| 连接何时建立、何时关闭 | 插件 SQL/命令的业务正确性 |
| 会话归哪个插件实例所有 | 数据库账号自身的库权限 |
| 审计元数据：action、caller、env、kind、ok/code、duration | SQL 文本、SSH 命令文本、密钥 |
| 连接失败如何映射为稳定错误码 | SSH 服务端的权限与策略 |

数据库账号权限和 SSH 服务端权限才是最终边界。TT 不解析、不审查插件 SQL 或命令的业务意图；只读会话的文本层检查只用于拒绝明显非只读的语句。

## 会话与句柄

句柄是不透明随机串，绑定 **插件 ID + 实例 ID + 环境 + 连接类型**。借用或猜测他人句柄一律返回 `SESSION_NOT_FOUND`，不区分“不存在”和“不属于你”，避免枚举。

以下事件关闭会话：

- 插件 stop / disable。
- 环境授权被撤销。
- TT shutdown。
- 任意 config 写入（`connsvc.Reload`）。

因 config 变更关闭的会话必须重新 `OpenDB` / `OpenSSH`。TT 不会把旧会话透明改指到新环境或新账号；配置变了就重开，避免“看起来还在原会话”的误判。

## 只读分裂

这是最容易出错的地方：内置只读路径与插件通用会话是两条独立路径，策略不能互相继承或冒充。

| 路径 | 文本层 | 库侧只读 | 说明 |
|---|---|---|---|
| 内置只读：`tt db sql`、dict live、dbsync | `internal/safesql` | Kingbase `default_transaction_read_only`；Oracle `SET TRANSACTION READ ONLY` | 原有**两层**都保留 |
| 插件 `OpenDB(read_only=true)` | `Query` 仍过文本层检查 | 建立库侧只读会话 | 独立路径，不复用内置路径的策略对象 |
| 插件 `OpenDB(read_only=false)` | 无 | 无 | 普通会话，两层都不加 |

`read_only=true` 时对会话执行 `Exec` 返回 `DB_SESSION_READONLY`。插件不得把 `read_only=false` 描述成“内置只读”；内置只读也不得因插件需要被放宽。

## RPC 形态

v1 全部是 unary：没有 streaming、事务、参数绑定。`DBSession.Query` 默认最多 1000 行，硬上限 50000 行，超限置 `truncated`。`SSHSession.Exec` 是一次性非交互命令；非零退出码是**正常结果**，不是 RPC 错误。v1 没有 PTY、SFTP。

## 错误码

所有错误以 Connect error detail 携带 `ErrorInfo{code, message, hint, env}`。code 是稳定契约，已发布的不重命名；调用方按 `code` 分支，禁止解析 `message`（它允许变化）。

| code | 含义 |
|---|---|
| `ENV_UNKNOWN` | 环境不存在 |
| `ENV_NOT_AUTHORIZED` | 当前插件未获该环境授权（存在性检查之前判定） |
| `PERMISSION_DENIED` | 缺少对应 capability |
| `CONFIG_BROKEN` | 环境配置损坏，无法解析 |
| `CONFIG_CHANGED` | 配置已变化，原会话失效；需要重新打开 |
| `DB_NOT_CONFIGURED` | 该环境没有数据库配置 |
| `DB_CONNECT_FAILED` | 数据库连接失败 |
| `SSH_CONNECT_FAILED` | SSH 连接失败 |
| `SESSION_NOT_FOUND` | 句柄不存在、已过期或不属于调用者 |
| `SESSION_KIND_MISMATCH` | 句柄类型与操作不匹配（如拿 SSH 句柄查库） |
| `DB_SESSION_READONLY` | 在只读会话上执行写操作 |
| `DB_QUERY_FAILED` | 查询失败 |
| `DB_EXEC_FAILED` | 执行失败 |
| `SSH_EXEC_FAILED` | SSH 命令执行失败 |
| `TIMEOUT` | 连接、查询或命令超时 |
| `DB_ACCOUNT_UNKNOWN` | 指定账号不在该环境的账号清单内 |
| `SERVICE_CLOSED` | Host 服务已关闭或正在关闭 |
| `UNAUTHENTICATED` | 传输层：token 缺失或无效 |
| `BAD_ORIGIN` | 传输层：`Origin` 不是 loopback 或字面量 `null` |
| `BAD_HOST` | 传输层：`Host` 不是 loopback 名称/IP |

## 尚未实现

插件实例 token 与 supervisor 尚未接入，外部插件当前无法通过这些服务建立会话；Host 侧连接实现已存在。
