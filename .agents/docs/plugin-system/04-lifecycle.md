# 生命周期

## 状态

`downloading` → `installed_disabled` → `enabling` → `running` | `degraded` → `disabling` → `disabled`；另有 `upgrading`、`rolling_back`、`uninstalling`、`corrupt`。

## 允许的迁移

| 当前 | 允许转入 | 条件 / 说明 |
|---|---|---|
| `downloading` | `installed_disabled` | 下载、校验、解包成功；失败不写 registry，清理临时物 |
| `installed_disabled` | `enabling`、`upgrading`、`uninstalling`、`corrupt` | 安装完成但未启用 |
| `enabling` | `running`、`degraded`、`installed_disabled` | 就绪则 `running`；进程存活但未就绪则 `degraded`；启用失败回到 `installed_disabled` |
| `running` | `degraded`、`disabling`、`corrupt` | 健康检查失败或进程崩溃进入 `degraded` |
| `degraded` | `running`、`disabling`、`corrupt` | 重启并恢复健康后回 `running` |
| `disabling` | `disabled` | 确认进程树退出后才算 `disabled`；停止失败留在 `disabling` 并报错 |
| `disabled` | `enabling`、`upgrading`、`uninstalling`、`corrupt` | 可重新启用 |
| `upgrading` | `installed_disabled`、`rolling_back`、`corrupt` | 只从非运行态进入；运行中先 `disabling` |
| `rolling_back` | `installed_disabled`、`corrupt` | 旧版本恢复成功则回到未启用；回滚失败为 `corrupt` |
| `uninstalling` | 移除 registry 条目（终态）、`corrupt` | 默认保留插件数据；清理失败标 `corrupt`，不假装已卸载 |
| `corrupt` | `uninstalling` | 不允许启用；修复以重新安装完成 |

## 清理与规则

| 时点 | 必须发生 |
|---|---|
| 安装 | 只落到 `installed_disabled`。安装、启用、授权是三个独立步骤；解包不得留下半激活插件 |
| 启用 | 从 registry 的期望状态登记 CLI/UI/skills；启动 service；失败时回滚登记，进程不残留 |
| 运行 / 降级 | service 崩溃按有界退避重启，绝不热循环；连续失败进入 `degraded` |
| 禁用 / 停止 | 撤销实例 token；关闭该实例的全部会话；从期望状态注销 CLI/UI/skills |
| 升级 | 程序版本不可变：新版本另装，旧版本保留到切换完成；插件数据默认保留 |
| 卸载 | 注销所有贡献；插件数据默认保留，除非用户显式要求删除 |
| 损坏 | 只允许卸载或重新安装；不猜测破损的 manifest，也不带着破损状态启用 |

## 注销方式

CLI 命令、UI 入口、skills 注册都必须按 registry 的**期望状态**重算，不能靠“删掉插件目录”顺带完成。删目录只清理文件，不负责注销；两者混在一起会让 registry 与磁盘长期不一致，出现“已卸载但仍能调用”的幽灵命令。

## 停止与崩溃

停止前先验证实例身份：token、实例 ID 与启动信息共同确认，**PID 单独不是身份**（PID 会被复用）。Windows 上必须杀掉整棵进程树，否则 supervisor 退出后子进程仍在运行，下一次启动会出现两个实例。

## 尚未实现

manifest/registry/installer、supervisor、token 签发与撤销、升级/回滚均未实现。上述状态机是后续阶段的实现契约。
