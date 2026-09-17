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
| `tdebug serve` / `tdict serve` | `tt serve`（一个服务，一套页面） |
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
| `/api/*` | 统一接口：配置读写、派生状态、PATH 安装，以及字典类动作 |
| ~~`/dict/`~~ | 字典页 SPA 已删除（见第 4 期），其两个动作并入设置页 |

### 为什么必须按前缀分 API

合并前两个工具各自独立起服务，所以两边都注册了 `/api/status`、`/api/dbprobe`、
`/api/dbaccverify`、`/api/conntest` —— 放进同一个 mux 会直接冲突。

### 为什么子系统只挂 API，页面由统一服务提供

一开始的做法是把整个子系统 handler 用 `StripPrefix` 挂在 `/debug/` 下
（子系统内部自带 SPA 兜底 `/`）。这条路走不通：Go 的 `ServeMux` 一旦匹配到
`/debug/` 前缀就**不再向外回落**，子系统的 `/` 兜底会把该前缀下的所有请求吃掉，
页面反而出不来。

改成：`/debug/api/` 转给子系统，`/debug/` 由 `internal/web` 用
`SPAHandler(common.WebFrontend(), …)` 从构建产物直接服务。
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
  让持有配置内存态的调试服务重新加载。否则从别处改了环境，调试侧仍按旧环境连，
  表现是"改了没生效"。重新加载失败只记日志：配置已经落盘成功，不该把成功的写报成失败，
  但必须让用户看见。

前端最初按"两套页面并存互链"落地（各自保留原有外观与交互，导航栏互相有跳转入口）。
第 3 期把字典页接入共享主题层之后，第 4 期进一步取消了那套 SPA —— 见下。

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
  `/api/health` 可达、`/debug/` 返回界面、`/` 跳转到 `/debug/`、
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

## 八、前端统一设置页（合并完成后的一次重构）

三工具合并本身完成后，前端仍是两套 SPA（`web/debug` 与 `web/dict`），各带一个设置页，
两边的环境编辑器近乎重复。用户要求：**前端只保留一个设置页签**，用多级管理组织成
站点管理 / 数据字典 / DEBUG / 应用设置四块。这条要求最终把前端收敛成了一套页面。

### 第 0 期：抽出共享层 `web/shared/`

`docs/ARCHITECTURE.md` 本来就把 `web/shared/` 列在规划里，这一期把它落地：
主题变量表（`tokens.css`）、主题机制（`theme.ts`）、`cn`、UI 基元。
UI 基元按依赖拆成两个文件 —— `ui.tsx`（按钮/输入框/表格，不依赖 Radix）与
`ui-radix.tsx`（下拉/弹层/日历/确认弹窗/手风琴）—— 字典页不该被拖着引 Radix 全家桶。

顺带消掉一处重复：`Panels.tsx` 原来自己包了一层 `Accordion`/`AccordionItem`/`AccordionContent`。
现在共享件只做**中性的壳**（布局类留给调用方），面板侧用 `PanelItem`/`PanelContent`
加它需要的"撑满高度 + 条目间分割线"—— 设置页那套手风琴是轻量树形（不撑高、无分割线），
需求正好相反，把任一套的样式写进共享件都会让另一套到处覆盖。

### 第 1 期：设置页外壳 + 深链接，并修掉一处数据丢失 bug

设置页从「外观/环境/高级」三个平级分类重做成两级导航（分区 → 卡片）。
左树点分区展开它的卡片，点卡片滚到对应位置。

深链接 `#settings/<分区>`：字典页当时还在，它的「设置」入口要能精确落到数据字典分区。
`App.tsx` 的首屏逻辑改为**先认哈希再判空环境** —— 否则 `#settings/data-dict` 会被
"没有环境就跳设置"的兜底覆盖成站点管理。

**修掉的数据丢失 bug**（这是本期的主要动机之一）：`PUT /api/hosts` 是整节替换，
而 `DebugSettings` 每个字段都带 `omitempty` —— 省略即删除。老代码在「高级」里只发
5 个键，于是每次保存都会把 `debug.fglserver` / `persistBreakpoints` / `activeEnv`
从 config.json 抹掉。现在保存一律按**读到的整节 + 编辑项**构造，从结构上堵死这一类。

顺带把三个从没在界面上暴露过的字段补进 DEBUG（`fglserver`、`persistBreakpoints`
—— 它是 `*bool` 三态，用下拉不用复选框；以及 `activeEnv`）。

### 第 2 期：设置页只跟共享层说话

新增 `GET /api/config/status`：路径型取值的**派生状态**（配置值 + 在本机是否真的存在）。
单独一个端点而不并进 `/api/hosts`，因为它要 `os.Stat`，而后者是每次进设置都读的热路径。

`/api/install`（用户 PATH）从字典子系统移到统一层 —— 把 tt 加进 PATH 是应用级动作，
与字典查询无关。路径解析（`AbsPath` / `DirStatusOf` / `FileStatusOf` / `DefaultSyncTarget`）
上移到 `internal/config/pathutil.go`：原先散在字典子系统的三个文件里，而设置页也要同一份判断，
两边各算一遍会让同一个 `sync.target` 在两个页面上解析成不同结果。

**补掉同源的第二个丢失洞**：`hosts.sshs[].launchArgs` / `watchdogSeconds` 表单里没有
这两个控件，保存站点会把它们静默丢掉。服务端按环境名补回（把第 1 期的
`preserveUngovernedDBFields` 扩展成 `preserveUngovernedFields`，同时管环境级与 db 级键）。

### 第 3 期：字典页接入共享主题层

字典侧原先**完全没有主题机制**：它的 `dark:` 类走 Tailwind v4 默认的
`prefers-color-scheme`，跟随系统且不可切换。引入共享 `tokens.css` 之后 `.dark` 变体
变成 `:is(.dark *)`，从此跟随 `html.dark`。

这一期必须**原子提交**：一旦引入 `tokens.css` 而 `index.html` 的防闪色脚本或
`main.tsx` 的 `applyDark` 没同时落地，字典页会永久停在亮色**且毫无报错** ——
样式照常编译、构建照样绿。实机验证了暗色与亮色两种模式。

配色改造：满屏写死的 `zinc-*`/`sky-*` 换成语义 token；状态色（emerald/amber/red）保留
（调试页自己也是这么用的，那条"禁止写死"针对的是 zinc 色板）；进度条的**轨道与填充成对改**
（`bg-muted` + `bg-primary`），只改填充会让暗色下的轨道消失。

### 第 4 期：彻底移除字典页，两个拉取功能并入设置页

字典侧只剩「拉源码镜像」与「跑字典同步」两个动作。它们并进设置页「数据字典」分区的
同名卡片后，那套 SPA 就没有存在理由了 —— `web/dict/` 整体删除。

- 字典类端点从 `/dict/api/*` 升到共享层 `/api/*`。那套前缀是"两套页面各有一套 `/api/*`
  需要分开"的遗留。注意它带来一个反直觉的后果：包内的 `/api/` 兜底会接到与字典无关的
  路径，所以那条 404 的文案改成了通用的「未知接口」。
- 前端构建产物从 `dist/debug` 移到 dist 根：只剩一套 SPA，再套一层是多余的嵌套。
- 运行按钮在**改动未保存时禁用** —— 服务端用的是配置里的值而不是表单里的值，
  不这么做的话用户改了目录还没保存就点「全量重建」，镜像会拉到旧目录去而界面上看不出。

### 第 5 期：目录改名（清理）

`web/debug/` 装的是**整套应用**（工作台 + 覆盖三个工具的设置页），`web/shared/` 也只剩
一个消费者 —— 两个名字的含义都过期了。改名为 `web/app/` + `web/shared/`，
workspace 名从 `debug` 改成 `app`（`npm run build:app` / `dev:app` / `check:app`）。
`base: '/debug/'` 不变 —— 那是 URL 前缀，与目录名无关。

### 第 6 期：清理与文档

- 删掉字典子系统里已无人调用的三个写端点（`PUT /api/mirror`、`PUT /api/dbsync`、
  `PUT /api/bdldoc`）：它们写的三节现在统一由 `PUT /api/hosts` 的分节写入覆盖，
  前端只调 GET 与两个动作端点，CLI 的写命令直接改配置文件、不走 HTTP。
  留着就是三条重复的写路径。相应的测试改为直接落配置。
- 去掉前端的三行 shim（`theme.ts` / `lib/utils.ts` / `ui.tsx` 的再导出），
  调用点直接引 `shared/`。

### 这一轮的验证

- **omitempty 回归矩阵**（那个数据丢失 bug 的守门人）：逐节改一个字段保存，确认同节
  兄弟键一个都没掉 —— `debug` 节改 `termWidth` 而 `fglserver`/`persistBreakpoints`/
  `activeEnv` 原样；`hosts` 节 `sshs[0]` 10 个键进 10 个键出；`query`/`mirror`/`bdldoc`/
  `sync`/`listen` 各自提交后其余节不受影响。
- `grep` 确认 `web/app/src` 无写死色板；产物 CSS 里 `.dark` 是类驱动、
  `prefers-color-scheme` 归零、明暗两套 `--background` 都在。
- 实机核对四个分区渲染、深链接落位、明暗切换、`/dict/*` 全部 404、
  桌面冒烟全通（`desktop/scripts/smoke.mjs`）。
- **全新克隆可构建**：`web/dist/.gitkeep` 一直被 `.gitignore` 挡住、从未入库，
  而 `//go:embed all:web/dist` 在该目录不存在时会直接报错 —— 意思是从仓库克隆下来
  `go build` 过不去。已改为 `/web/dist/*` + `!/web/dist/.gitkeep`，并实测克隆后
  未构建前端也能 `go build` 且 `go test ./...` 全绿。

### 仍未做

- **设置页的搜索框**（用户给的 VS Code 参考图里有）。字段数约 30 个，够用得上，
  但用户明确说不做。
- **每环境的 `launchArgs`/`watchdogSeconds` 仍不在界面上暴露**：靠服务端保留机制保证
  不被丢掉（见第 2 期）。要真正可编辑，得给站点编辑器的每个环境加两个控件。
