---
name: tt-pulgin-add
description: 为 TT 创建或改造插件时使用。触发词：写插件、新插件、plugin.json、CLI 命名空间、常驻服务、插件前端、消息桥、Host API、skills 贡献、插件打包、发布索引。
---

# 给 TT 写一个插件

TT 只做两件事：**统一管理数据库/SSH 的配置与连接**、**做 CLI 与 skills 的分发平台**。  
插件拿连接去做什么，TT 不关注——但插件只能走公开契约，不能碰 TT 内部。

写之前先读（都在 `.agents/docs/plugin-system/`）：

- `11-plugin-author-guide.md`：完整作者指南，本 skill 是它的执行摘要
- `03-trust-model.md`：进程信任边界（插件是本机可执行代码，TT 不是沙箱）
- `05-cli-and-errors.md`：CLI 转发与退出码
- `13-release-index.md`：发布索引格式

## 1. 一个插件包，四类贡献

插件包是 zip（或目录），根部必须有 `plugin.json`。四类贡献互相独立，至少声明一类：

| 贡献     | 清单字段                     | 是什么                               |
| ------ | ------------------------ | --------------------------------- |
| CLI    | `contributions.cli`      | 注册命名空间，`tt <namespace> …` 转发给插件进程 |
| 常驻服务   | `contributions.service`  | TT 拉起并监管的独立进程                     |
| 前端页面   | `contributions.frontend` | 沙箱 iframe 里的完整页面                  |
| Skills | `contributions.skills`   | 随插件安装/升级/卸载分发的技能目录                |

## 2. 包布局与硬规则

```text
myplugin.zip
├─ plugin.json                 # 必须；在包根或唯一顶层目录里
├─ bin/windows-amd64/tool.exe  # 多平台时用 {os}-{arch} 目录
├─ ui/index.html               # 前端入口
└─ skills/my-skill/SKILL.md    # 每个技能目录必须有 SKILL.md
```

- 入口一律是**包内相对路径**；绝对路径、盘符、UNC、`..`、符号链接全部拒绝。
- 入口文件必须存在、是普通文件、非空。
- 解包上限：5000 文件 / 单文件 64 MB / 总量 512 MB / 压缩比 200，超限整包拒绝。
- `entry` 支持 `{os}`、`{arch}` 占位（`runtime.GOOS`/`GOARCH`），展开后仍需是包内相对路径。

## 3. plugin.json 最小模板

```json
{
  "schemaVersion": 1,
  "id": "com.example.myplugin",
  "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "name": "我的插件",
  "description": "一句话说明",
  "contributions": {
    "cli": {"namespace": "myplugin", "entry": "bin/{os}-{arch}/myplugin", "summary": "一句话简介"},
    "service": {"entry": "bin/{os}-{arch}/myplugin", "args": ["serve"], "healthPath": "/healthz"},
    "frontend": {"route": "myplugin", "title": "我的插件", "entry": "ui/index.html"},
    "skills": ["skills/my-skill"]
  },
  "permissions": ["profiles.list", "db.connect", "ssh.connect"]
}
```

要点：

- `id` 反向 DNS，安装后不可变；`version` 严格 SemVer，同版本不可变。
- `schemaVersion` 必须是 `1`；`hostApi` 当前写 `>=1.0.0 <2.0.0`（TT 的 HostAPIVersion 是 1.0.0）。
- `cli.namespace`：小写字母/数字/连字符，≤32，撞内置命令或已启用插件则安装失败。
- `service.args`：同一二进制既当 CLI 又当服务时，用它区分启动模式（别靠隐式环境约定）。
- `permissions` 只是**诉求**；真正的授权来自用户 `tt plugin enable --allow … --env …`。

## 4. CLI 转发契约

`tt <namespace> <args…>` = 起一个一次性进程。TT 只做三件事：注入环境、转发标准流、原样返回退出码。

注入的环境变量（除此之外没有别的，**不注入口令/私钥/别的插件的令牌**）：

| 变量                                         | 含义                                              |
| ------------------------------------------ | ----------------------------------------------- |
| `TT_HOST_API_URL`                          | Host API 根地址（`http://127.0.0.1:<port>/hostrpc`） |
| `TT_PLUGIN_TOKEN`                          | 本实例令牌，调用结束即失效                                   |
| `TT_PLUGIN_ID` / `TT_PLUGIN_VERSION`       | 插件 ID / 活跃版本                                    |
| `TT_PLUGIN_INSTANCE_ID` / `TT_PLUGIN_KIND` | 实例 ID / `cli` 或 `service`                       |
| `TT_PLUGIN_DATA_DIR`                       | 持久数据目录（升级/卸载默认保留）                               |
| `TT_PLUGIN_LOGS_DIR`                       | 日志目录                                            |
| `TT_ENV` / `TT_ENT`                        | 本次调用点名的环境与企业                                    |

规则：业务参数原样转发（顺序、空格、中文、emoji 不动）；stdin/stdout/stderr 直通；工作目录是调用方当前目录；退出码原样返回；`tt <namespace> --help` 也转发给你，详细帮助由插件自己输出。

## 5. Host API v1（ConnectRPC）

地址 `TT_HOST_API_URL`，认证 `Authorization: Bearer $TT_PLUGIN_TOKEN`。proto 契约在  
`proto/tt/host/v1/`，用任何语言按 proto 生成自己的客户端（**不要 import TT 的 `internal/gen`**）。

| 方法                                      | 作用              | 需要能力                         |
| --------------------------------------- | --------------- | ---------------------------- |
| `HostService/GetCapabilities`           | 版本、可用方法、特征开关    | 无                            |
| `ProfilesService/List`                  | 已授权环境的脱敏元数据     | `profiles.list`              |
| `ConnectionsService/TestDB` / `TestSSH` | 测连接，不建会话        | `db.connect` / `ssh.connect` |
| `ConnectionsService/OpenDB` / `OpenSSH` | 建立有所有权的会话句柄     | 同上                           |
| `ConnectionsService/Close`              | 关闭自己的句柄         | 同上                           |
| `DBSessionService/Query` / `Exec`       | 查询/执行（行数、时间有上限） | `db.connect`                 |
| `SSHSessionService/Exec`                | 非交互命令           | `ssh.connect`                |

要点：

- **能力 ≠ 环境**：有 `db.connect` 但没被 `--env` 授权的环境，照样用不了。
- 只读会话上 `Exec` 返回 `DB_SESSION_READONLY`，不会静默降级。
- 错误 = Connect 错误 + `tt.host.v1.ErrorInfo`（稳定 `code` + 人读 `message`）；  
  连不上/认证失败是 `ok=false + stage` 的正常返回，不是 RPC 错误。
- 秘密永不出现：只有环境名、类型、地址、账号名。

## 6. 常驻服务协议

声明了 `service` 且插件已启用才启动。TT 注入 `TT_PLUGIN_LISTEN=127.0.0.1:0`（端口内核选）。

```text
启动：监听后在 stdout 打一行 JSON：{"ready":true,"addr":"127.0.0.1:54321"}
      TT 随后 GET http://<addr><healthPath>（默认 /healthz），2xx 通过
停止：TT 先 POST /shutdown（带 Bearer 令牌），再关 stdin；宽限期（默认 3s）后杀整棵进程树
崩溃：非预期退出 → 撤销令牌、关闭会话 → 退避重启最多 3 次；用尽后 degraded 并写 LastError
```

- stdout 第一行之后全是日志（落到 `logs/<id>/service-<时间戳>.log`）。
- 收到 stdin EOF 或 `/shutdown` 就自己退出，不留孙进程。

## 7. 前端页面与消息桥 v1

- 页面在 `sandbox="allow-scripts"` 的 iframe 里（**无** `allow-same-origin`）：拿不到宿主 DOM、store、preload。
- 静态资源由 TT 从活跃版本目录提供：`/plugins/<id>/ui/…`。
- 通信只走消息桥：父页面先发 `init`（带随机 `nonce`）；之后每条消息都带  
  `protocol:"tt.plugin.bridge"`、`v:1`、`id`、`session`、`nonce`、`ts`。  
  类型：`init` / `host.request`+`host.response` / `rpc.request`+`rpc.response` / `cancel` / `error`。  
  单条上限 1 MiB，响应行数上限 5000。
- 调 Host API：发 `host.request`（`method` 形如 `tt.host.v1.ProfilesService/List`），  
  由父页面代调——**页面永远拿不到令牌**。
- 调自己的服务：发 `rpc.request`（相对路径），TT 代理到 `/plugins/<id>/rpc/…`，只去本插件实例。
- 协议细节照抄 `web/app/src/pluginProtocol.ts`，不要自己发明字段。

## 8. Skills 贡献

- 清单里列**包内目录**，每个目录必须有带 frontmatter 的 `SKILL.md`。
- 安装/启用/升级/卸载都会重算期望态，走五平台分发管线，只动平台清单记过的条目。
- 同名冲突规则：官方优先、先登记优先；冲突会在 `tt plugin doctor` 和设置页显示。

## 9. 开发自测流程

```bash
# 参考样板：tools/refplugin（一个二进制兼 CLI 与服务）与 testdata/plugins/reference
go build -o refplugin.exe ./tools/refplugin   # 看样板怎么写

tt plugin install ./myplugin.zip        # 或目录；只解包注册，不执行代码
tt plugin enable com.example.myplugin --allow profiles.list --env <环境>
tt myplugin --help                       # 验证 CLI 转发
tt plugin doctor com.example.myplugin   # 清单/版本目录/授权/启用状态自检
tt plugin logs com.example.myplugin     # 看最近运行日志
tt plugin disable / uninstall           # 卸载默认保留持久数据
```

升级自测：新版本登记进发布索引 → `tt plugin update --index <url> --dry-run` 看计划 →  
去掉 `--dry-run` 真升 → 出事 `tt plugin rollback <id>`。TT 不做自动更新。

## 10. 写插件前必须接受的边界

- 插件是本机可执行代码，装了就等于信任；不要把 TT 当沙箱承诺给用户。
- 默认拿不到明文口令/私钥，要连接就走 Host API。
- 管理面（装/卸插件、授权、TT 更新、shutdown、别的插件的 RPC、任意文件）对插件关闭。
- 不要读 TT 的 `config.json`，不要 import TT 内部包，不要依赖 Host API 之外的任何行为。
