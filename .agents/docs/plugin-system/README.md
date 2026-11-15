# TT 插件系统契约

本目录定义第三方插件与 TT 之间的边界。契约先于实现：Host API v1 在 `proto/tt/host/v1/`，
运行时（manifest、registry、installer、supervisor、token 签发、CLI 分发、前端页面、
skills 一体化、升级回滚）均已落地，入口是 `tt plugin …`。

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
| [10-first-party-migration.md](10-first-party-migration.md) | 第一方迁移清单（已迁 / 下一批候选 / 明确不迁） |
| [11-plugin-author-guide.md](11-plugin-author-guide.md) | **插件作者指南**：清单、Host API、CLI、服务协议、桥、skills、发布 |
| [12-user-guide.md](12-user-guide.md) | **用户指南**：安装、授权、启停、升级、回滚、卸载、日志、自检、信任提示 |
| [13-release-index.md](13-release-index.md) | 插件发布索引 v1：形状、资产键、选版本规则、https 策略 |

## 已实现

- Host API proto 与生成物（`tt.host.v1`）、ConnectRPC `/hostrpc/`、Bearer 认证、能力探测、
  环境脱敏列表、连接测试/开关、DB 查询/执行、SSH 一次性执行、`ErrorInfo` 错误详情。
- 清单（严格 JSON）与校验、注册表（原子写 + 备份恢复 + 锁）、安全安装、启停、授权、卸载。
- 一次性 CLI 转发（参数/标准流/退出码原样）、常驻服务 supervisor（就绪行 + `/healthz` +
  有界重启 + 进程树终止）、每实例令牌、Host API 授权。
- 插件页面（`/plugins/{id}/ui/`）、消息桥（iframe + nonce）、插件私有 RPC 代理、
  Host API 桥。
- 技能一体化：来源、冲突、五平台分发、升级/卸载重算、assets 指针刷新。
- 手动升级与回滚（发布索引 + 下载校验 + 未激活安装 + 原子切换 + 健康检查 + 自动回滚 +
  保留当前与上一个版本）、`tt plugin doctor/logs` 与只读 Web 日志端点。

## 尚未实现 / 不在本项目

- 自动更新、后台检查、插件市场：计划明确不做（`tt plugin update` 是手动命令）。
- 插件间依赖解析、通用事件总线、任意 UI 注入点。
- `dict/dbsync` 之外的更多第一方迁移：见 [10-first-party-migration.md](10-first-party-migration.md)
  （`debug` / `dev` 明确不在本项目范围内）。
- v1 不含：流式、事务、参数绑定、PTY、SFTP。
