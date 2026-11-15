# TT 插件系统契约

本目录定义第三方插件与 TT 之间的边界。契约先于实现：Host API v1 已在 `proto/tt/host/v1/` 落地；插件运行时（manifest、registry、installer、supervisor、token 签发、CLI 分发、前端页面）尚未实现。

当前 `/hostrpc/` 已挂载，但没有任何有效插件 token，所有请求返回 401。本目录描述实现必须满足的规则，不是“已经可用”的承诺。

## TT 管什么，不管什么

一句话：TT 管“谁可以用哪个环境、连接何时开闭、会话归谁”，不管“插件用连接做什么”。TT 不是沙箱，也不替数据库或 SSH 服务端做权限判断。

| 维度 | TT 管 | TT 不管 |
|---|---|---|
| 身份 | 插件实例、token、能力、环境授权 | 插件进程内部行为 |
| 连接 | 何时开/关、会话归谁、失败如何报 | 插件拿连接做什么 |
| 资源 | 允许访问哪些环境 | OS 文件、网络、进程权限 |
| 审计 | action、caller、env、kind、ok/code、duration | SQL 文本、SSH 命令文本、密钥 |
| 前端 | 页面隔离与 Host API 边界 | 插件页面内部的业务逻辑 |

## 文档索引

| 文件 | 内容 |
|---|---|
| [01-permissions.md](01-permissions.md) | capability 矩阵、贡献≠权限、管理面拒绝、Host/Origin/token |
| [02-connection-boundary.md](02-connection-boundary.md) | 连接归属、只读分裂、会话失效、错误码 |
| [03-trust-model.md](03-trust-model.md) | 进程信任、页面隔离、Electron 约束、密钥处理 |
| [04-lifecycle.md](04-lifecycle.md) | 状态机、允许迁移、各阶段清理 |
| [05-cli-and-errors.md](05-cli-and-errors.md) | 命名空间、转发契约、退出码边界、错误信封 |
| [06-versioning.md](06-versioning.md) | 版本归属、proto/buf 规则、兼容策略 |
| [07-namespaces.md](07-namespaces.md) | 内置命令、旧别名、保留名、冲突规则 |
| [08-fixtures.md](08-fixtures.md) | 参考插件夹具布局与变体（设计） |
| [09-ssh-hostkey.md](09-ssh-hostkey.md) | 当前 SSH 主机密钥行为与兼容约束 |

## 现状

已实现：Host API proto 与生成物（`tt.host.v1`）、ConnectRPC `/hostrpc/`、Bearer 认证、能力探测、环境脱敏列表、连接测试/开关、DB 查询/执行、SSH 一次性执行、`ErrorInfo` 错误详情。

尚未实现：插件 manifest/registry/installer、CLI 分发、supervisor、插件实例 token、前端插件页面与 bridge、技能集成、插件升级/回滚。详见各篇的“尚未实现”。

v1 不含：流式、事务、参数绑定、PTY、SFTP。
