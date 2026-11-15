# 用户指南：安装、授权、升级、回滚、卸载

面向"用 TT 装一个别人写的插件"的人。插件作者请读 [11-plugin-author-guide.md](11-plugin-author-guide.md)。

## 0. 一句话的信任提示

**装插件 = 在这台机器上装并运行别人的程序。** TT 管的是"它能用哪些环境、什么时候能用"，
不提供沙箱：插件是普通本机进程，能读你能读的文件、能连你能连的网络。只装你信任的来源；
不确定就先在测试环境试。

页面隔离（iframe + 消息桥）保护的是**页面**，不是进程。v2 起同页插件页面用插件自己的 RPC/服务，宿主不再替 iframe 代调 Host API。

## 1. 安装

```bash
tt plugin install ./myplugin.zip          # 本地 zip
tt plugin install ./myplugin-src          # 本地目录（开发态）
tt plugin install ./myplugin.zip --replace  # 同 ID 同版本已存在时重装
```

安装做四件事：解包到缓存 → 严格校验清单 → 原子移进 `plugins/installed/<id>/<version>/` →
登记为**未启用**。它**不执行**任何插件代码。

同 ID 同版本已装会拒绝（`--replace` 才重装）；插件正在启用中不允许覆盖安装。
命名空间撞内置命令或别的已启用插件时，安装即失败。

## 2. 启用（v2：启用即信任）

```bash
tt plugin enable com.example.myplugin
```

- v2 起**没有逐项能力/环境授权**：启用后的插件可以使用 Host API 的全部能力与配置里的
  全部环境。`--allow` / `--env` 仍被接受但忽略（兼容旧脚本）。
- 信任边界是"安装并启用"：只装你信任的来源；不确定就先在测试环境试。
- 清单里的 `permissions` 字段已废弃（v2 清单声明它会被拒绝）。
- `tt plugin enable` 之后：CLI 命名空间出现在 `tt --help`，页面出现在菜单，技能进入分发
  期望态；有常驻服务的插件，服务由 `tt serve` 拉起（命令行自己不起长命进程）。
- 同样的动作也可以在 **设置 → 插件管理** 里做：启用、停用、自检、看日志；
  两侧写的是同一份注册表，不另存一套状态。

## 3. 停用与卸载

```bash
tt plugin disable   com.example.myplugin     # 停服务、撤令牌、关会话；程序/授权/数据都留着
tt plugin uninstall com.example.myplugin     # 先要求已停用，再删程序与注册项
tt plugin uninstall com.example.myplugin --purge-data --yes   # 连持久数据一起删
```

- 停用**不删**程序与授权：重新 `enable` 就恢复。
- 卸载默认**保留**持久数据（`plugins/data/<id>/`）与日志。删数据是单独的一次明确动作。
- 卸载只清该插件自己的目录与它登记过的技能条目，不碰别的东西。

## 4. 升级（手动）

```bash
tt plugin update --index https://plugins.example.com/index.json --dry-run   # 先看计划
tt plugin update --index https://plugins.example.com/index.json             # 升全部已装
tt plugin update --id com.example.myplugin                                   # 用上次的索引地址
tt plugin update --id com.example.myplugin --to 0.2.0                        # 钉版本
```

**TT 不自动更新插件**：没有后台检查，不自动下载，也不在 `tt update` 里顺带升插件。
什么时候升、升到哪一版，由这一次命令决定。

### 升级做了什么

顺序是有意的，每一步都为"失败不破坏现在能用的版本"：

| 步 | 动作 | 失败时 |
|---|---|---|
| 1 | 读发布索引（https；也支持本机文件/共享盘） | 什么都没变 |
| 2 | 选版本：比当前新 + 与本机 Host API 兼容 + 有本平台资产 | 报"已是最新"或点明原因 |
| 3 | 检查没有别的 tt 进程正跑着它的服务 | 报错并让你先 `tt serve --stop` |
| 4 | 下载到 `plugins/cache/`，校验大小 + SHA-256 | 删掉半成品，旧版本照跑 |
| 5 | 解包到暂存目录，校验清单（ID/版本必须与索引一致） | 旧版本照跑 |
| 6 | 装成**未激活**版本（登记进注册表，不改活跃版本） | 撤掉刚移进来的载荷 |
| 7 | 停旧服务 → 撤旧实例令牌、关它的会话 | 旧服务重新拉起，活跃版本不变 |
| 8 | **原子切换**活跃版本（一次带锁写入） | 同上 |
| 9 | 起新服务 + 健康检查 → 检查 UI 入口 → 重算技能 | **自动回滚**（见下） |
| 10 | 清理旧程序版本，只留"当前 + 上一个" | 清理失败不影响已升级的版本 |

持久数据（`plugins/data/<id>/`）全程不动。

### 回滚

- 第 9 步任何失败（服务起不来、健康检查不过、页面入口丢了、技能刷不进平台目录）都会：
  停掉失败的新版本服务 → 撤销实例令牌 → 把活跃版本切回旧的 → 重新起旧服务并健康检查 →
  重算技能 → 报清"升级到 X 失败，已回滚到 Y：原因"。原因也会写在注册表的 `lastError`。
- 自动回滚也失败时，错误里会给出下一步：`tt plugin disable` 再 `tt plugin rollback`。

```bash
tt plugin rollback com.example.myplugin           # 切回上一个版本
tt plugin rollback com.example.myplugin --dry-run # 只看会切到哪
```

"上一个版本"是升级时记下的那个。回滚会走同一套安全动作（校验、停服务、撤凭证、
原子切换、起服务、健康检查、重算技能）；回滚本身失败时会把活跃版本切回原处并报清原因。
只有更早的版本被清理过，再往回就没有了（"只留当前与上一个"是策略）。

### 为什么命令行升级会被"tt serve 正跑着服务"挡住

`tt plugin update` 是一次性进程：它改得动注册表与程序目录，但停不掉 `tt serve` 手里的
子进程。若强行切换，会出现「页面/CLI 是新版、常驻服务还是旧版」，旧令牌与旧会话都还活着 ——
那正是要避免的新旧混用。所以这种情况会直接拒绝，并告诉你怎么做：

```bash
tt serve --stop
tt plugin update --id com.example.myplugin --index <url>
tt serve
```

没有常驻服务的插件（只有 CLI / 页面 / 技能）不受影响，可以直接升级。

## 5. 看日志

```bash
tt plugin logs com.example.myplugin             # 最新一个日志文件的最后 50 行
tt plugin logs com.example.myplugin --tail 200
tt plugin logs com.example.myplugin --json
```

只读 `plugins/logs/<id>/` 里**最新**的那个文件（一个插件会攒很多个日志文件，把目录拼起来
会把十分钟前的东西混进刚才那行错误）。没有日志不是错误：那说明服务没起来过。

设置页同一份数据：`GET /api/plugins/{id}/logs?tail=N`（只读，本机请求）。

## 6. 自检

```bash
tt plugin doctor                       # 全部插件
tt plugin doctor com.example.myplugin  # 一个
```

只看本地事实：不启动插件、不连网。有问题时退出码 1（问题清单照样打印）。

| 代码 | 严重度 | 意思 | 下一步 |
|---|---|---|---|
| `NO_ACTIVE_VERSION` / `VERSION_DIR_MISSING` | error | 注册表与磁盘对不上 | 重新安装 |
| `MANIFEST_BROKEN` / `MANIFEST_INVALID` | error | 清单坏了或被改过 | 重新安装 |
| `HOSTAPI_INCOMPATIBLE` | error | 插件要的 Host API 与当前 TT 不兼容 | `tt plugin update` 或等作者发兼容版 |
| `UI_ENTRY_MISSING` | error | 页面入口文件不在了 | 重新安装 |
| `SERVICE_VERSION_STALE` | error | 跑着的服务是旧版本（串版） | `tt serve --stop` 后 `tt plugin update` |
| `SERVICE_NOT_RUNNING` / `SERVICE_STALE_STATE` | warning | 声明了服务却没在跑 / 状态文件是残留 | 重启 `tt serve` |
| `SKILLS_SOURCE_MISSING` / `SKILLS_CONFLICT` | error / warning | 技能来源缺失 / 与别的插件同名 | 重新安装 / 改名或只留一个 |
| `GRANTS_INVALID` / `PERMISSION_NOT_GRANTED` | warning | 授权不再有效 / 声明了但没授予 | `tt plugin grant …` |
| `STALE_VERSION` / `PREVIOUS_VERSION_MISSING` | warning | 多出来的旧版本 / 记着的上一个版本不在了 | 下次 `update` 会清理；回滚会因此失败 |
| `LAST_ERROR` | warning | 最近一次启用/升级失败的原因 | 看 `tt plugin logs` |

## 7. 看状态

```bash
tt plugin list            # 已启用的
tt plugin list --all      # 连未启用的
tt plugin info <id>       # 清单、版本（含上一个）、授权、落点
```

设置页的插件列表在同样字段之外还带运行态（`state` / `addr` / `pid` / `restarts`），
那是增量字段，命令行不看它。

## 8. 常见问题

| 现象 | 原因 | 处理 |
|---|---|---|
| `tt: unknown command "x"` | 插件没启用 | `tt plugin list --all` 看状态，`tt plugin enable` |
| 调用插件得到 `PLUGIN_DISABLED` | 它被停用了 | 重新 `enable` |
| 插件里拿不到环境 | 只授了能力没授环境 | `tt plugin grant --env <环境>` |
| `DB_UNREACHABLE` | 网络/VPN/口令/隧道问题（不是授权问题） | 先用 `tt db ping --at <环境>` 定位 |
| 升级报"服务正由另一个 tt 进程运行" | `tt serve` 正跑着它 | `tt serve --stop` 后再升 |
| 页面白屏 | 页面入口缺失或版本被删改 | `tt plugin doctor` → 重装 |
| 升级后行为还是旧的 | 服务没重启（命令行升级不了别人的子进程） | 重启 `tt serve`；doctor 会报 `SERVICE_VERSION_STALE` |

## 9. 相关文档

- 契约与边界：[README.md](README.md)、[01](01-permissions.md)、[02](02-connection-boundary.md)、[03](03-trust-model.md)
- 生命周期：[04-lifecycle.md](04-lifecycle.md)
- 发布索引（作者侧）：[13-release-index.md](13-release-index.md)
- 插件作者指南：[11-plugin-author-guide.md](11-plugin-author-guide.md)
