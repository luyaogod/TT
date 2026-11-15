# SSH 主机密钥

## 当前行为

`internal/host.Dial` 与 `internal/sshtun.Dial` 都使用 `ssh.InsecureIgnoreHostKey()`：

- [internal/host/ssh.go](../../../internal/host/ssh.go#L29)
- [internal/sshtun/sshtun.go](../../../internal/sshtun/sshtun.go#L45)

这意味着服务器提供什么主机密钥就接受什么，**不做服务器身份验证**。这是内部工具既有的兼容行为，与手工登录的信任级别一致；不能把它描述成强安全边界。

## 风险

- 无法发现主机被替换或中间人攻击；在不可信网络上尤其明显。
- SSH 会话的“连上了”不等于“连对了服务器”。
- 插件通过 Host API 使用 SSH 时沿用同一行为，不能声称插件链路提供了额外的主机真实性保证。

## 兼容约束

- 插件项目不得为自身需要悄悄改掉内置路径的主机密钥行为；这会改变内置工具既有行为，且让“插件引入的变更”难以审计。
- 如要改为 known-hosts 或显式信任，必须作为独立的安全加固项，明确兼容策略、迁移方式与用户可见变化，不能夹带在插件阶段。
- 在加固完成前，文档、帮助与错误提示都不得暗示已做主机身份校验。

## 尚未实现

known-hosts 校验、显式信任配置均未实现。这是后续的独立 hardening 项，不属于插件 v1。
