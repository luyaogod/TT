# 迁移记录

记录这次合并**改了什么、为什么改**，以及迁移中发现的、与原项目注释不符的地方。
架构设计见 [ARCHITECTURE.md](ARCHITECTURE.md)。

来源：`D:\我的项目\TDebug`、`D:\我的项目\TDev`、`D:\我的项目\TDictCli`。
目标：`D:\我的项目\TT`（新建仓库；三个原仓库原地保留，作为只读参考，未做任何改动）。

## 一、合并动机：三处"请手工同步"

三个项目互相依赖，靠注释提醒人工保持同步：

| 位置 | 原文 |
|---|---|
| `TDebug/cli/root.go:53` | 存放规则(与 TDictCli 保持一致，改动请两边同步) |
| `TDictCli/cli/root.go:173` | 存放规则(与 TDebug 保持一致，改动请两边同步) |
| `TDebug/cli/install.go:6` | 命令形态与 TDictCli / TDev 一致(改动请三边同步) |

这三处现在各只有一份实现。另外两份 `cfgfile` 包、两份 `dbconfig`、两份 `host`、
两份 `pathinstall`、两份 `erpdb` 也合并了。

## 二、合并中发现的分叉（重要）

两个 `host` 包**不是**简单的复制关系，已经实质分叉。逐文件比对后，TDebug 版在
**每个有分叉的文件上都是更成熟、更安全的一版**：

### `host/ssh.go`

TDebug 版修掉了一处数据竞争：原实现把 `sess.CombinedOutput` 的结果写进外层变量，
超时分支返回后那个 goroutine 仍在写它（`go test -race` 可复现）。TDebug 版改用
带缓冲的 channel 传出结果。

同时新增 `OutputStdin`，为下面 `dbprobe` 的修复提供了基础。**采用 TDebug 版。**

### `host/dbprobe.go` —— 这次合并最重要的发现

TDictCli 版的 SQL 执行是**完整的命令注入面**：

```go
// 旧版（TDictCli）
func SqlplusRun(zone, sqlplusPath, connStr, sql string) string {
    q := strings.ReplaceAll(sql, "'", "'\\''")          // 只转义单引号
    return fmt.Sprintf(`bash -lc '%s; echo "%s" | %s -S %s'`, ChenvCmd(zone), q, p, connStr)
}
```

SQL 拼进命令串后再整体套一层 `bash -lc`，**两层解析**：数据里的 `$(...)`、反引号、
双引号都会在远端 shell 里展开，一个引号就能把后面的 `;` `|` 变成命令分隔符。
而 `tdict` 执行 SQL 的输入来自用户查询与 AI 生成的语句，不是常量。

TDebug 版把它换成了「返回不含 SQL 的命令行 + SQL 经 stdin 送入」：

```go
func SqlplusCmd(zone, sqlplusPath, connStr string, killAfterSec int) (string, error)
func KbCmd(ksqlPath, host, port, db, connStr string) (string, error)
// 调用方：conn.OutputStdin(cmd, []byte(sql), timeout)
```

并且补了 `reToolPath` / `reDBAcct` 白名单校验（工具路径、数据库账号），
拒绝了含换行的口令；`killAfterSec > 0` 时给远端命令套 `timeout -s TERM`——
Oracle 侧没有任何服务端超时机制，本地 SSH 层超时只是"不再等"、不杀进程，
这一步是必须的。

**采用 TDebug 版。TDictCli 那套不安全的 SQL 拼接没有保留。**
`tt dict` 的调用点相应改写（`internal/cli/dict`），这是本次合并里唯一需要
改动调用方行为的地方。

### `host/tenv.go`

TDictCli 版用变量名前缀（`TDICT_TOP`…）解析探针回显，但 PTY 只有纯文本，
登录脚本自己也会打印 `ZONE = t35prd`、`TOPENT = 99` 这类行，前缀法无法区分
"我方请求的值"与"服务器自己打印的内容"。TDebug 版用 `TDBG-BEGIN`/`TDBG-END`
定界符区间 + 状态机，超宽换行产生的碎片也落在区间之外。**采用 TDebug 版。**

### `host/mirror.go`

TDebug 没有这个文件（源码镜像只服务于字典查询）。**从 TDictCli 搬入。**

### `dbconfig`

TDictCli 版少了 `ReadonlySQL *bool` 与 `ReadonlySQLEnabled()`。**采用 TDebug 版（超集）。**

### `erpdb` / `pathinstall`

TDictCli 版是超集：`erpdb` 多了 `SelectAllSQL` / `ValidIdent` / `QuoteIdent` / `QuoteLit`；
`pathinstall` 已经是完整包（`Status`/`Get`/`Preview`/`Add`/`Remove` + 路径列表工具函数 + 测试），
而 TDebug 版是两个散落在 `cli/` 里的文件。**采用 TDictCli 版。**

### `cfgfile`

两份逐字节相同，只差一个节访问器（`Debug(root)` vs `Hosts(root)`）。
**重写为 `internal/config`。**

## 三、统一配置：结构变更

### 环境清单从两处收敛到一处

合并前的两份 schema **完全一样**，只是节名不同：

```jsonc
// tdebug: { "debug": { "sshs": [ {name, host, port, user, password, zone, topent, db{...}} ], "listen": ... } }
// tdict:  { "hosts": { "sshs": [ {name, host, port, user, password, zone, topent, db{...}} ], "activeEnv": ... } }
```

合并后只留 `hosts.sshs`。

### 一个必须换掉的迁移规则

TDictCli 原有的迁移规则是「顶层 `debug` 键整节重命名为 `hosts`」：

```go
// TDictCli/cfgfile/cfgfile.go（旧）
func Hosts(root map[string]any) (map[string]any, error) {
    if h, _ := root["hosts"].(map[string]any); h != nil { return h, nil }
    if legacy, _ := root["debug"].(map[string]any); legacy != nil {
        root["hosts"] = legacy
        delete(root, "debug")     // ← 合并后会把 TDebug 的调试设置一并吞进 hosts
        return legacy, nil
    }
    ...
}
```

在合并后的配置里 `debug` 是**真节**（放 `launchArgs` / `watchdogSeconds` / `termWidth` /
`printElements` / `persistBreakpoints` / `fglserver`），照搬这条规则会把这些设置
搬进 `hosts` 并从 `debug` 删掉。所以换成了 `internal/config/migrate.go` 的**拆解**规则：

- `debug.sshs` → `hosts.sshs`（提升）
- `debug.listen` → 顶层 `listen`（提升）
- `debug` 节其余键原样留在 `debug`

`internal/config/config_test.go` 里 `TestMergeConfigs_BothTools` 专门盯这一点：
断言合并后 `debug` 节仍存在、`debug.sshs` 不残留、`debug.listen` 已提升。

### 新增

- 顶层 `schemaVersion`（当前 2）。0/缺失 = 合并前旧结构，触发迁移。
- 顶层 `listen`。原来挂在 `debug.listen`，但它是整个服务的监听地址，不是调试专属。
  合并前 TDebug 与 TDictCli 的默认值就是同一个 `127.0.0.1:28670`。
- `tdev` 节。原 TDev **完全没有配置系统**，所有参数都是每次调用的 flag；
  新节只放跨调用稳定的默认值（`workspaceSuffix` / `defaultOut`），
  **flag 仍然优先**，且配置读取失败绝不让命令失败（见 `internal/dev/cli/settings.go`）。

### 迁移的合并顺序

环境清单取并集，**同名环境以 TDictCli 那份为准**。理由：它是字典查询的现役配置，
保留它能让查询侧行为完全不变；TDebug 独有的环境仍然会被并进来。
原文件不删除，各留一份 `.pre-merge.bak`。

## 四、配置位置

两份 `resolveConfigPath` / `toolsHome` / `isPortable` / `migrateLegacyConfig` /
`looksLikeOwnConfig` / `samePath` / `defaultConfigPath` 逐行相同，只有
`toolDirName`（`tdebug` / `tdict`）与环境变量名不同。合并为 `internal/config/paths.go` 一份：

| | 合并前 | 合并后 |
|---|---|---|
| 工具目录 | `%APPDATA%\T100\tdebug`、`%APPDATA%\T100\tdict` | `%APPDATA%\T100\tt` |
| 显式指定 | `TDEBUG_CONFIG` / `TDICT_CONFIG` | `TT_CONFIG`（旧的两个名仍识别） |
| 便携标记 | `.portable` | 不变 |
| 覆盖根目录 | `T100_HOME` | 不变 |

首次运行会自动发现并合并旧位置的两份配置。

`Open` 修了一处文档与实现不符：注释写"空文件视为空配置"，实现却对空文件
抛 `unexpected end of JSON input`。现在按文档行为处理（空文件/纯空白 → 空配置）。

`Save` 从「`path + ".tmp"` + rename」换成了 TDev 那套更完整的原子写
（同目录 `CreateTemp` + `Write` + `Sync` + 保留原权限位 + `Rename`）。

## 五、命令面

| 合并前 | 合并后 |
|---|---|
| `tdebug start` | `tt debug start` |
| `tdev tzc export` | `tt dev tzc export` |
| `tdict r.t` | `tt dict r.t` |
| `tdebug serve` / `tdict serve` | `tt serve`（一个服务，两套页面） |
| `tdebug install skills` / `tdict install skills` / `tdev install skills` | `tt install skills` |
| `tdebug env` + `tdebug topent` + `tdict env` | `tt env list/show/use/topent` |
| — | `tt config path/show/get/set/migrate/validate` |

旧命令名保留为别名（`tt tdebug …` = `tt debug …`），既有脚本与 AI skill 不改即可运行。

TDev 的两点被完整保留：位置无关的参数解析（`-o`/`--json` 可出现在位置参数之后，
标准库 `flag` 做不到），以及明确的退出码契约（0/2/3/4/5）。`tt dev` 用
`DisableFlagParsing` 把参数原样交给 TDev 自己的解析器，退出码经 `exitCode` 透传。

## 六、统一 Web 服务

一个进程、一个端口：

| 路径 | 内容 |
|---|---|
| `/debug/` | 调试工作台 SPA（前端构建产物 `web/dist/debug`） |
| `/debug/api/*` | 调试专属 API（后端内部仍按 `/api/…` 注册，靠 `StripPrefix` 挂在前缀下） |
| `/dict/` | 字典页 SPA（前端构建产物 `web/dist/dict`） |
| `/dict/api/*` | 字典专属 API（同上） |
| `/api/*` | 两套页面共用的端点 |

### 为什么必须按前缀分 API

合并前两个工具各自独立起服务，所以两边都注册了 `/api/status`、`/api/dbprobe`、
`/api/dbaccverify`、`/api/conntest` —— 放进同一个 mux 会直接冲突。

### 为什么子系统只挂 API，页面由统一服务提供

一开始的做法是把整个子系统 handler 用 `StripPrefix` 挂在 `/debug/` 下
（子系统内部自带 SPA 兜底 `/`）。这条路走不通：Go 的 `ServeMux` 一旦匹配到
`/debug/` 前缀就**不再向外回落**，子系统的 `/` 兜底会把该前缀下的所有请求吃掉，
页面反而出不来。

改成：`/debug/api/` 转给子系统，`/debug/` 由 `internal/web` 用
`SPAHandler(common.WebSub("debug"), …)` 从构建产物直接服务。`/dict/` 同理。
两个子系统因此只负责 API，页面归统一服务一家管。

配套的两个坑，都在 `internal/web/spa.go` 里注释着：

- **挂载点与目录不能交给 `http.FileServer`**：它对目录返回一次到 `/` 的 301，
  前缀被丢掉，用户会从 `/debug/` 被弹到根路径。现在目录一律直接给 `index.html`。
- **未命中的路径要回落 `index.html`**，否则刷新页面或粘深链接都是 404。

`//go:embed all:web/dist` 的 FS 根是 `web/dist`，而所有消费者都按"dist 根"理解路径，
所以 `main.go` 用 `fs.Sub` 剥掉一层；"前端放在哪"这件事只在 `common.WebSub` 里说一次。

### 共用的配置端点

`GET/PUT /api/hosts`：读回 `hosts`/`listen`/`debug`/`query`/`mirror`/`bdldoc`/`sync`/`tdev`，
PUT 只接受要改的节（缺省 = 不动），交给 `internal/config` 的原子写路径。两个页面都写它。

服务端在三处替前端兜底：

- **拒绝空环境列表、重名、端口越界、库类型非法、`activeEnv` 指向不存在的环境**，
  400 里给中文说明，前端原样展示 —— 校验归服务端，前端只管显示。
- **保存时保留表单不管理的 `db.viaSsh`**（SSH 端口转发隧道）。配置页没有它的输入控件，
  整节替换会把它抹掉 —— 那是"打开设置页点一下保存，手工配的隧道就没了"，
  而用户完全看不出发生过什么。按环境名对齐补齐；代价是通过页面删不掉 `viaSsh`，
  想删就手改 `config.json`。宁可选"删不掉"。
- **写入成功后调用子系统的 `ReloadConfig()`**（`internal/web.ConfigReloader`），
  让持有配置内存态的调试服务重新加载。否则从字典页改了环境，调试侧仍按旧环境连，
  表现是"改了没生效"。重新加载失败只记日志：配置已经落盘成功，不该把成功的写报成失败，
  但必须让用户看见。

前端保持**两套页面并存互链**（按用户选择，不做 UI 重建）：各自保留原有外观与交互，
导航栏互相有跳转入口，共享的只有环境配置数据与实现它的 `/api/hosts`。

## 六之二、合并过程中修掉的两个 bug

两处都是"整节替换"这个动作的副作用，都由测试抓出来：

1. **类型化 nil 指针装进 `interface{}` 不等于 nil。**
   `PUT /api/hosts` 原本用 `for key, v := range map[string]any{"query": req.Query, …}`
   加 `if v == nil { continue }` 判断"这一节没提交"。但那些字段是指针类型，
   nil 指针装进 `any` 后接口值非 nil，于是**本次没提交的节会被序列化成 `null`
   再当成空对象写回去，把用户原有的 `query`/`mirror`/`debug` 全清掉**。
   现在逐个类型化判空（`internal/web/server.go`）。

2. **`viaSsh` 被配置页静默丢弃**（见上）。

两个 bug 都有回归测试盯着：`internal/web/server_test.go` 的
`TestHostsPut_LeavesOtherSectionsAlone` 与 `TestHostsPut_PreservesViaSSH`。

## 七、实测与未做

### 已验证

- `internal/config`：合并规则、路径解析、端到端迁移、幂等性、未知键保留 —— 见
  `config_test.go` / `paths_test.go`（含"无关的 config.json 不被误迁"、
  "已是新结构则不再迁移"、"显式 `--config` 指向不存在文件时不得退回默认落点"）。
- **真实旧配置的迁移**：`realdata_test.go` 直接读本机 `%APPDATA%\T100\{tdebug,tdict}\config.json`，
  合并后断言每个环境都还在、调试设置没丢进 `hosts`、字典设置原样带过来。
  本机没有旧配置时自动跳过。实测：3 个环境并集、`activeEnv` 取 tdict 那份、无数据丢失。
- `internal/dev`：`tt dev tzc selftest` 31 项全过；全部 `_test.go` 通过。
- `internal/web`：`PUT /api/hosts` 的分节写入、校验、拒绝时不落盘、`viaSsh` 保留、
  SPA 静态服务的五个路径形态 —— 见 `server_test.go`。
- `internal/host`、`internal/pathinstall`、`internal/safesql`、`internal/dict`：原测试随包搬入并通过。
- **桌面链路**：`desktop/scripts/smoke.mjs` 无 GUI 跑通 —— 首启建配置骨架、打出 `TT_READY`、
  `/api/health` 可达、`/debug/` 与 `/dict/` 都返回各自界面、`/` 跳转到 `/debug/`、
  `POST /api/shutdown` 优雅退出（code=0）。

### 未做 / 已知取舍

- **`viaSsh` 无法通过配置页删除。** 为了不"保存一次就静默丢配置"，服务端保存时会把
  表单不管理的 `db.viaSsh` 从旧配置补回来（见 §六）。要删就手改 `config.json`。
- **`tt config set` 不做节内校验。** 深路径写入会绕过环境页的校验（环境名唯一、
  端口范围等）。改环境清单请用 `tt env` 或配置页。
- **`tt dev` 的 config 默认值只覆盖 `-o` 省略的场景。** TDev 原本没有任何配置系统，
  新增的 `tdev` 节刻意只放两个稳定的默认值，没有扩张。
- **`tt --config` 对 `tt dev` 需要额外一步**：该命令组是 `DisableFlagParsing`（为了保住
  TDev 位置无关的参数解析），cobra 不会替它解析全局 flag，所以 `--config` 会被摘出来
  转成 `TT_CONFIG` 环境变量再转发。三种写法（`tt --config X dev …`、`tt dev … --config X`、
  `TT_CONFIG=X tt dev …`）都已验证可用。
- **`internals/dev` 的 `store.ToolName` 仍写作 `"tdev tzc"`**：那是写进 `manifest.json`
  的持久化字段，没有任何代码读它；保持稳定意味着重命名二进制不会让已有工作区的
  manifest 变身份。
- 「当前环境」有两套语义，都保留：`tt env use <名称>` 改的是**配置默认**
  （`hosts.activeEnv`）；`tt debug env <名称>` 切的是**活动调试会话**（会重连）。
  `tt debug topent <值>` 是会话级覆盖，与 `tt env topent <名称> <编号>` 设的配置默认不同。
- 原三个仓库未做任何改动，仍在 `D:\我的项目\` 下；`TT` 是全新仓库，
  没有合并 git 历史（按用户选择）。
