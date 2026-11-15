# 插件作者指南

面向"要写一个 TT 插件"的人。契约的**理由**在 [01](01-permissions.md)–[09](09-ssh-hostkey.md)，
本文只讲"怎么写才能装上、跑起来、升级不炸"。

> **v2 提示**：新插件用 `schemaVersion: 2`；同页 UI 走 `contributions.ui` 的槽位
> （`nav / page / header.actions / settings.section / statusbar / overlay`），不再需要
> `frontend` 与 iframe。`tt plugin new` 生成骨架、`tt plugin pack` 打包并写回 `tt` 下限、
> `tt plugin inspect` 查看包内容；v2 不再有 `permissions` 能力授权（启用即信任）。
> 下文的 `schemaVersion: 1` / `frontend` 示例是 legacy 参考。

先读三篇再动手：[03-trust-model.md](03-trust-model.md)（进程信任边界）、
[05-cli-and-errors.md](05-cli-and-errors.md)（转发与退出码）、
[01-permissions.md](01-permissions.md)（能力与环境）。

## 1. 最小可构建示例：参考插件

仓库里有两份可直接用的示例，**不要**把它们装到用户环境（`testdata/` 只进测试）：

| 位置 | 是什么 | 怎么用 |
|---|---|---|
| `tools/refplugin/main.go` | 一个 Go 程序，同一个二进制既是 CLI 又是常驻服务；带 `args/exit/crash/sleep/echo-stdin/hostrpc/service` 等模式 | `go build -o refplugin.exe ./tools/refplugin` |
| `testdata/plugins/reference/` | 上面的二进制 + `plugin.json` + `ui/index.html` + 一个技能目录 | 构建后把 exe 放进 `bin/`，`tt plugin install testdata/plugins/reference` |

参考插件的清单就是"四类贡献全都要"的样板：

```json
{
  "schemaVersion": 1,
  "id": "com.example.refplugin",
  "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "name": "参考插件",
  "description": "测试夹具：同一个二进制既当 CLI 又当常驻服务",
  "contributions": {
    "cli": {"namespace": "refplugin", "entry": "bin/refplugin.exe", "summary": "参考插件（测试夹具）"},
    "service": {"entry": "bin/refplugin.exe", "args": ["service"], "healthPath": "/healthz"},
    "frontend": {"route": "refplugin", "title": "参考插件", "entry": "ui/index.html"},
    "skills": ["skills/ref-skill"]
  },
  "permissions": ["profiles.list", "db.connect", "ssh.connect"]
}
```

`testdata/plugins/variants/` 里还有三个"故意装不上"的变体（非法清单、命名空间冲突、
零贡献），可以拿它们验证你的安装/诊断流程。

## 2. 包布局

```text
myplugin.zip
├─ plugin.json                 # 必须；包根或唯一的顶层目录里
├─ bin/
│  ├─ windows-amd64/tool.exe   # 按平台分目录时用 {os}-{arch}
│  ├─ linux-amd64/tool
│  └─ tool-service.exe
├─ ui/index.html               # 前端入口（可以引用同目录的 js/css）
└─ skills/my-skill/SKILL.md    # 技能目录（每个都要有 SKILL.md）
```

硬规则：

- 入口都是**包内相对路径**；绝对路径、盘符、UNC、`..`、符号链接一律拒绝。
- 入口文件必须存在、是普通文件、非空。
- zip 解包有上限（5000 个文件 / 单文件 64 MB / 总量 512 MB / 压缩比 200），
  超出就是整包拒绝，不是"跳过几个"。

## 3. `plugin.json` 字段

| 字段 | 必填 | 规则 |
|---|---|---|
| `schemaVersion` | 是 | 必须是 `1`。 |
| `id` | 是 | 反向 DNS（`com.example.myplugin`），安装后不可变。 |
| `version` | 是 | **严格 SemVer**（`0.2.0`）。版本目录用它命名，同一版本不可变。 |
| `hostApi` | 是 | 兼容范围，如 `>=1.0.0 <2.0.0`。当前 TT 的 `connsvc.HostAPIVersion` 是 `1.0.0`。 |
| `name` / `description` | 否 | 展示用。 |
| `contributions` | 是 | 至少一类：`cli` / `service` / `frontend` / `skills`。 |
| `contributions.cli` | 否 | `namespace`（小写字母/数字/连字符，≤32，不能撞内置命令）+ `entry` + `summary`。 |
| `contributions.service` | 否 | `entry` + 可选 `args`（启动参数）+ 可选 `healthPath`（默认 `/healthz`）。 |
| `contributions.frontend` | 否 | `route`（`#plugin/<route>`）+ `title` + `entry`。 |
| `contributions.skills` | 否 | 包内技能目录数组，每个都要有 `SKILL.md`。 |
| `permissions` | 否 | 能力名数组，**只是诉求**；用户 `tt plugin enable --allow …` 给的才是授权。 |

### 3.1 `entry` 里的 `{os}` / `{arch}` 占位

```json
{"cli": {"namespace": "tool", "entry": "bin/{os}-{arch}/tool"}}
```

- `{os}` → `runtime.GOOS`（`windows` / `linux` / `darwin`）
- `{arch}` → `runtime.GOARCH`（`amd64` / `arm64`）

只支持这两个占位；展开后仍然必须是包内相对路径。这样同一个包里可以放两个平台的二进制，
而清单只有一份。

### 3.2 `service.args`

同一个可执行文件既能当 CLI 又能当常驻服务是常见写法。`args` 就是启动服务时传的参数：

```json
{"service": {"entry": "bin/tool.exe", "args": ["serve", "--port", "0"]}}
```

没有它就只能靠环境变量约定，而那是隐式契约。

## 4. Host API v1

TT 在 `http://127.0.0.1:<port>/hostrpc/` 暴露 ConnectRPC 服务，认证是
`Authorization: Bearer <实例令牌>`。令牌由 TT 发放（每个服务实例一个），插件不能自报身份。

| 服务/方法 | 作用 | 需要的能力 |
|---|---|---|
| `HostService/GetCapabilities` | Host API 版本、可用服务与方法、特征开关 | 无（认证即可） |
| `ProfilesService/List` | 已授权环境的**脱敏**元数据（无口令/私钥） | `profiles.list` |
| `ConnectionsService/TestDB` / `TestSSH` | 测一条连接（不建立持久会话） | `db.connect` / `ssh.connect` |
| `ConnectionsService/OpenDB` / `OpenSSH` | 建立有所有权的会话，返回不透明句柄 | `db.connect` / `ssh.connect` |
| `ConnectionsService/Close` | 关闭自己的句柄 | 同上（只能关自己的） |
| `DBSessionService/Query` / `Exec` | 在句柄上查询/执行（有行数与时间上限） | `db.connect` |
| `SSHSessionService/Exec` | 在 SSH 会话上跑一条非交互命令 | `ssh.connect` |

要点：

- **能力 ≠ 环境**。只给 `db.connect` 不给 `--env` 时，任何环境都用不了；
  每个环境是单独授权的一次决定。
- 只读会话上 `Exec` 会得到 `DB_SESSION_READONLY`，不会静默降级成查询。
- `Query` 有行数上限（默认 1000，最多 50000）与超时（默认 60s），超出会带 `truncated`。
- 错误是 Connect 错误 + `tt.host.v1.ErrorInfo` 详情（稳定 `code` + 人读 `message`）；
  连接层失败（连不上/认证失败）是 `ok=false + stage` 的正常返回，不是 RPC 错误。
- 秘密永远不出现：只有环境名、类型、地址、账号名。

proto 契约在 `proto/tt/host/v1/`，生成物在 `internal/gen/`（插件不 import 它们，
用任何语言按 proto 生成自己的客户端）。

## 5. CLI 转发契约

`tt <namespace> …` 转成"起一个一次性进程"，TT 只做三件事：注入环境、转发标准流、
原样返回退出码。

注入的环境变量：

| 变量 | 含义 |
|---|---|
| `TT_HOST_API_URL` | Host API 根地址（`http://127.0.0.1:<port>/hostrpc`） |
| `TT_PLUGIN_TOKEN` | 本实例令牌（随调用结束失效） |
| `TT_PLUGIN_ID` / `TT_PLUGIN_VERSION` | 插件 ID / 活跃版本 |
| `TT_PLUGIN_INSTANCE_ID` / `TT_PLUGIN_KIND` | 实例 ID / 调用方种类（`cli` 或 `service`） |
| `TT_PLUGIN_DATA_DIR` | 持久数据目录（升级/卸载默认不动它） |
| `TT_PLUGIN_LOGS_DIR` | 日志目录 |
| `TT_ENV` / `TT_ENT` | 用户在这次调用里点名的环境与企业（`--at` / `--env`） |

不注入：完整配置、口令、私钥、别的插件的令牌、TT 的内部开关。

规则：

- 业务参数**原样**转发（顺序、空格、中文、emoji 都不动）；TT 只在第一个业务参数之前
  摘自己的全局开关（`--config/--env/--at/--conn/--format/--json/…`）。
- stdin/stdout/stderr 直通，工作目录是调用方的当前目录。
- 退出码原样返回；TT 自己的失败（插件未启用、清单坏了、进程起不来）用 TT 的错误信封
  与退出码，**不伪装**成插件业务退出码。
- `tt <namespace> --help` 也转发给你：详细帮助由你输出，TT 不生成。

## 6. 常驻服务协议

只有声明了 `service` 且插件已启用时才启动；监听地址由 TT 注入
`TT_PLUGIN_LISTEN=127.0.0.1:0`（端口由内核选，避免撞车）。

```text
启动：进程起来后，在 stdout 打一行 JSON：
        {"ready":true,"addr":"127.0.0.1:54321"}
      TT 随后对 http://<addr><healthPath> 做健康检查（默认 /healthz，2xx 算通过）
停止：TT 先 POST /shutdown（带 Bearer 令牌），再关它的 stdin；
      宽限期（默认 3s）后结束整棵进程树
崩溃：非预期退出 → 撤销实例令牌、关闭它的会话 → 退避重启，
      最多 3 次；用尽后停在 degraded 并写 LastError
```

为什么是"一行 JSON"而不是等端口：插件拿到的可能是 `:0`，只有它自己知道选到了哪个；
而"起来了吗"必须能区分"进程活着"与"在监听"。

工程要求：

- `stdout` 第一行之后的内容都当日志（TT 会落 `logs/<id>/service-<时间戳>.log`）。
- 优雅停止：收到 stdin EOF 或 `/shutdown` 就自己退出；不要留下孙进程。
- 服务崩了不能带走 TT：你的进程是独立进程，TT 会重启你而不是跟着挂。

## 7. 前端页面与消息桥 v1

- 页面在 `sandbox="allow-scripts"` 的 iframe 里（**没有** `allow-same-origin`），
  拿不到宿主 DOM、Zustand store、Desktop preload。
- 静态资源由 TT 从**活跃版本目录**提供：`/plugins/<id>/ui/…`（每次请求现解析版本，
  升级不会串版）。
- 与宿主通信只走消息桥：父页面先发 `init`（带随机 `nonce` 与会话事实），
  之后每条消息都要带 `protocol: "tt.plugin.bridge"`、`v: 1`、`id`、`session`、`nonce`、`ts`。
  消息类型：`init` / `host.request`+`host.response` / `rpc.request`+`rpc.response` /
  `cancel` / `error`；单条上限 1 MiB，响应行数上限 5000。
- 调 Host API：发 `host.request`（`method` 形如
  `tt.host.v1.ProfilesService/List`），由父页面转成 `/hostrpc/` 的调用 ——
  **页面自己永远拿不到令牌**。
- 调自己的服务：发 `rpc.request`（相对路径），由 TT 代理到
  `/plugins/<id>/rpc/…`。代理只去"这个插件自己的实例"，不是通用本机端口代理。
- 父页面是权限执行点：它不信任 iframe 自报的插件 ID / 环境 / 权限。

细节与错误码见 `web/app/src/pluginProtocol.ts` 与 [03-trust-model.md](03-trust-model.md)。

## 8. Skills

- 清单里列的是**包内目录**，每个目录必须有 `SKILL.md`（含 frontmatter）。
- 安装/启用/升级/卸载都会**重算期望态**，走现有 `internal/skill` 的五平台分发管线：
  只动平台清单（`.agents/.tt-manifest.json`）记过的条目，不整目录删。
- 升级切版本后 assets 指针会重新指向新版本目录；这是"必须重算"的原因。
- 同名冲突的规则是"官方优先、先登记优先"，冲突会在 `tt plugin doctor` 与设置页里显示。

## 9. 权限边界（写插件前必须接受）

- **进程信任**：插件是本机可执行代码，装了就等于信任。TT 不是沙箱。
- 默认拿不到明文口令/私钥；要连接就走 Host API。
- 管理面（插件安装/卸载、授权、TT 更新、shutdown、其他插件 RPC、任意文件）对你关闭。
- 能用哪些环境由用户逐环境授权；授权撤销后旧令牌立即失效。
- 不要试图直接读 `config.json`：那是 TT 的，且里面的秘密不属于你的插件。

## 10. 发布、安装、升级工作流

1. **打包**：`bin/` 按平台放好，`plugin.json` 的 `version` 就是这次发布的版本。
   zip 里不要带 `.git`、测试数据或你的本机配置。
2. **发布索引**：把 zip 上传到 https 地址，在索引里登记
   `{version, hostApi, assets:{<os>-<arch>:{url, sha256, size}}}`。
   索引格式与选版本规则见 [13-release-index.md](13-release-index.md)。
3. **本机自测**：

   ```bash
   tt plugin install ./myplugin.zip
   tt plugin enable com.example.myplugin --allow profiles.list --env prod
   mynamespace --help
   tt plugin doctor com.example.myplugin
   tt plugin logs com.example.myplugin
   ```

4. **升级自测**：把新版本登记进索引，`tt plugin update --index <url> --dry-run` 看计划，
   去掉 `--dry-run` 真升；崩了用 `tt plugin rollback com.example.myplugin`。
5. **不要**期望 TT 自动更新：没有后台检查，也没有自动下载。

用户视角的命令与信任提示见 [12-user-guide.md](12-user-guide.md)。
