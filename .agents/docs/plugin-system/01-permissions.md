# 权限与授权

权限的唯一单位是 **capability**。`cli`、`service`、`frontend`、`skills` 是贡献（contribution），不是权限；声明贡献不会自动获得任何能力或环境。

## Capability 矩阵

| capability | 授予 | 说明 |
|---|---|---|
| `profiles.list` | 读取已授权环境的脱敏元数据 | 只返回授权范围内的环境 |
| `db.connect` | `TestDB` / `OpenDB`，从而建立 DB 会话 | 同时覆盖只读会话与可写会话 |
| `ssh.connect` | `TestSSH` / `OpenSSH`，从而建立 SSH 会话 | |

`HostService.GetCapabilities` 只要求有效 token，不要求 capability；它返回 Host API 版本、tt 版本、`features` 与服务/方法清单。`features` 的取值：

| feature | 含义 |
|---|---|
| `FEATURE_DB_QUERY` | `DBSession.Query` 可用 |
| `FEATURE_DB_EXEC` | `DBSession.Exec` 可用；与内置只读防线无关 |
| `FEATURE_SSH_EXEC` | `SSHSession.Exec` 可用 |
| `FEATURE_DB_READONLY_SESSION` | 支持 `OpenDB(read_only=true)` |

## 贡献不是权限

`cli`、`service`、`frontend`、`skills` 只决定插件注册哪类贡献。若贡献即可授权，一个只画页面的插件就能顺带开库会话；因此贡献与 capability 必须分开授予。

- 声明 `service` 不等于 `db.connect` 或 `ssh.connect`。
- 声明 `frontend` 不等于可以调用管理面。
- 声明 `cli` 不等于获得额外的环境。

## 环境授权

环境授权按插件单独记录在 Core registry 中，粒度到环境。`ProfilesService.List` 只返回调用者获授权的环境；未授权环境不出现在结果里。

授权检查发生在环境存在性检查**之前**。这样未授权调用者无法用 `ENV_UNKNOWN` 与 `ENV_NOT_AUTHORIZED` 的差异枚举环境。插件也不能自行扩大授权：没有任何 Plugin API 可以修改自己的 environment grant。

## 管理面默认拒绝

插件 token 对下列操作一律拒绝，即使请求来自 loopback、即使 token 有效：

| 操作 | 默认结果 |
|---|---|
| 写 `config.json` | 拒绝 |
| 安装 / 卸载 / 启用 / 禁用插件 | 拒绝 |
| 修改授权（capability 或环境） | 拒绝 |
| TT 自更新 | 拒绝 |
| `/api/shutdown` | 拒绝 |
| 其他插件的私有 RPC | 拒绝 |
| 任意文件系统管理 | 拒绝 |
| 桌面 shell | 拒绝 |

## Host / Origin / token

`/hostrpc/` 使用 `Authorization: Bearer <token>`。没有 cookie 或隐式身份；每个请求都必须带有效实例 token。

| 检查 | 规则 | 失败码 |
|---|---|---|
| token | 缺失、过期、与实例不匹配 | `UNAUTHENTICATED` |
| `Host` | 必须是 loopback 名称或 IP | `BAD_HOST` |
| `Origin` | 存在时必须是 loopback 或沙箱 iframe 的字面量 `null` | `BAD_ORIGIN` |

Loopback 不是信任。本机任何进程都能向 `/hostrpc/` 发请求，因此 token 始终必需；Host/Origin 校验挡的是浏览器侧伪造来源的请求，不承担进程隔离。同 OS 用户能读用户文件这一事实不因这些检查改变，见 [03-trust-model.md](03-trust-model.md)。

## 尚未实现

插件实例 token 的签发、capability 授予与环境授权的管理入口尚未实现，因此当前 `/hostrpc/` 对一切请求返回 401。上述拒绝表是 token 生效后的默认策略，不是当前的运行时行为。
