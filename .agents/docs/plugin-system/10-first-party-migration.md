# 第一方迁移清单（下一批候选）

本文件回答一个问题：**除了已经迁掉的那两条链路，还有哪些第一方功能该迁到
`connsvc`，以及为什么现在不动手。**

判定标准只有一条（不是"代码看起来重复"）：

> 这段代码是否**自己**在建连接或隧道（`erpdb.OpenSession`、`sshtun.Dial`、`host.Dial`、
> 自己拼 ssh 命令），而不是通过 `internal/connsvc`？

自己建连接的每一处，都是"同一件事的第二份实现"：改一次超时、改一次只读策略、改一次
隧道收口，都得记得它。

## 1. 已迁移

| 链路 | 位置 | 迁到什么 | 验证 |
|---|---|---|---|
| 设置页「测试连接」（DB/SSH） | `internal/web/{conntest,sshtest}.go` → `connsvc.Service.TestDBConn / TestSSHConn` | 进程内 `connsvc` | `internal/hostrpc/handler_test.go` `TestLocalAndRPCProduceTheSameResult`（本地直调 vs RPC，同一实现、同一结果） |
| `tt db ping`（客户端直连） | `internal/cli/db/ping.go` → `connsvc.Service.TestDBConn` | 同上 | `internal/cli/db/ping_test.go` `TestPingAndRPCHitTheSameConnSvc`（同一条 Service 实例上，`runDBPing` 与 `/hostrpc/` 的 `TestDB` 结果一致，且各自都真的开了一次会话） |
| 内置只读数据面（字典直查 / dbsync） | `internal/dict/live`、`internal/dict/dbsync` → `connsvc.OpenBuiltin` | 收口了"起隧道 → 改地址 → 连库 → 收口"这段编排 | `internal/connsvc` 用例；`internal/dict/...` 回归 |

`tt db ping` 的迁移边界值得记一笔：**输出形状与退出码一个都没改**（`DB_UNREACHABLE` / 4、
JSON 字段 `env/ent/account/dialect/target/version`）。迁移是内部换实现，不是换接口。

## 2. 下一批候选（按建议顺序）

### 2.1 `dict` 远程直查 —— 第一候选

| 项 | 内容 |
|---|---|
| 现状 | `internal/dict/live` 已经走 `connsvc.OpenBuiltin`（连接与隧道已收口），但**查询编排**仍是自己的一份：取列/取行、截断、错误包装、只读判定。 |
| 要迁什么 | 把"在一条只读会话上跑一条 SQL、按上限截断、错误映射"的编排收敛到 `connsvc`（`DBQuery` 已经是这个形状），让 `tt dict r.t` 与 `tt db sql`、插件 `DBSession.Query` 共用同一段。 |
| 风险 | 中。字典查询是**用得最多**的只读路径，任何行为差异（列名大小写、NULL 表示、截断提示）都会被立刻看见。 |
| 阻塞 | `tt dict` 的返回信封与 `tt db sql` 不同（字典有自己的 `rows/cols` 约定）；先决定"是一份实现两个投影"还是"合并信封"，否则迁移会顺手改掉对外契约。 |
| 建议做法 | 只迁**执行**一段（连接、只读事务、超时、错误码），保留字典自己的投影层；加一条 `dict` 往返用例证明与 `tt db sql` 同库同结果。 |

### 2.2 `dbsync`

| 项 | 内容 |
|---|---|
| 现状 | `internal/dict/dbsync` 用 `connsvc.OpenBuiltin` 拿连接，但同步自己的批次、重试与进度上报各写一份。 |
| 要迁什么 | 批次/重试/取消的语义如果能用 `DBQuery` + `context` 表达，就迁；不能表达的部分（大批量拉取、写本地 SQLite）留在 dbsync。 |
| 风险 | 中高。dbsync 是**长事务**场景，与 `DefaultQueryTimeout`（60s）冲突；先要确认"长任务"是否需要独立的超时契约。 |
| 阻塞 | 会话超时策略：插件通用会话有 60s 默认超时与 5 万行硬上限，同步一趟可能远超。需要先设计"长任务"这条路径（不是简单放开上限）。 |

### 2.3 `debug` 会话（**不在本项目**）

| 项 | 内容 |
|---|---|
| 现状 | `internal/debug` 自己有 SSH 连接、远端进程、断点存档、源码镜像。 |
| 为什么不迁 | 它不是"一条 SQL/一条命令"的连接使用者：会话状态（调试目标、作业、企业快照）与连接生命周期**深度耦合**，还牵涉服务器侧进程与行锁。把它塞进 `connsvc` 会把"插件连接服务"变成一个调试框架。 |
| 结论 | **明确不在本插件项目范围内**。若将来要迁，应作为独立计划：先定义 `DebugSession` 契约，再谈复用连接。 |

### 2.4 `dev`（.tzc/.tzs 设计器管线）（**不在本项目**）

| 项 | 内容 |
|---|---|
| 现状 | `internal/dev`、`engine/`、`internal/cli/dev`：纯本地文件管线 + 一个 C# 引擎进程，几乎不碰数据库/SSH 会话。 |
| 为什么不迁 | 它与"连接授权"没有交集。硬迁只会给它加一层没有收益的间接。 |
| 结论 | **明确不在本插件项目范围内**。 |

## 3. 迁移的验收模板

任何一条链路迁完之后，至少要能回答：

1. **同一实现**：一份用例里，本地直调与 `/hostrpc/` 的 RPC 调用都打在同一个
   `connsvc.Service` 实例上，结果字段一致（参考 `TestLocalAndRPCProduceTheSameConnSvc`）。
2. **输出不变**：CLI 的 JSON 字段、退出码、`--human` 文案保持原样；改变对外形状要单独声明。
3. **失败可判**：连接层失败（连不上/认证失败）与契约层失败（没授权/环境不存在）在输出上
   仍能区分，与迁移前一致。
4. **没有第二份隧道**：迁移后该链路不再直接出现 `sshtun.Dial` / `host.Dial`。

## 4. 参考

- 连接边界：[02-connection-boundary.md](02-connection-boundary.md)
- 权限矩阵：[01-permissions.md](01-permissions.md)
- 升级/回滚与运行态：[12-user-guide.md](12-user-guide.md)
