# TT （T100 开发CLI）

TT 是用于开发鼎捷数智旗下的大型ERP系统（T100）所构建的CLI工具集

---

## 安装

参考 https://github.com/luyaogod/TT/releases

- 便携包
    ```powershell
    #手动将tt安装到用户PATH
    tt install path
    ```
- MSI安装包

---

## 升级

```powershell
tt update check     # 查有没有新版（不下载不安装）
tt update           # 查 + 下载校验，问一句再装
tt update log       # 上一次升级走到哪一步
```

`tt update` 按你现在的安装形态装：便携包就地覆盖程序目录（`config.json` 与目录里
你自己的文件不碰），MSI 安装交给 `msiexec`（用户级安装，不需要管理员）。升级全程
需要能访问 GitHub；内网要经代理时用 `tt update --proxy http://host:port`，
或把地址写进 `config.json` 的 `net.proxy`。

升级会把 agent 目录里的技能树一并刷成同版（`tt install skills` 装过哪些地方记在
数据目录的 `.tt-skills.json`）。如果那一步没成，`tt version` 会提示你手动跑
`tt install skills --force`。

MCP/Agent 环境里注意：默认会问一句再装，脚本里要用 `tt update --yes`；
先看它打算做什么可以 `tt update --dry-run`。

---

## 初次配置

```
tt serve
```

> 启动Serve后在本地打开配置网页` http://127.0.0.1:28670`
>
> 至少需要配置一个T100远程Linux环境

---

## 命令清单

全部命令按顶层命令分组，每张命令表三列：**命令**（完整写法：`<尖括号>` = 必填位置参数，
`[方括号]` = 可选）、**说明**、**参数**（该命令自己的开关；`<类型>` 与 `=默认值` 照抄
`tt <命令> --help`，标（必填）的不给自己会被拒；一格内多行 = 多个开关）。

每个命令都另外认下面这组全局开关，表里不再重复：

| 全局开关 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `--env <环境名>` | string | — | 这次操作落在哪个环境（`config.json` 的 `hosts.sshs[].name`）；环境一律点名，没有默认环境可回落 |
| `--config <路径>` | string | 数据目录下 `config.json` | 配置文件路径 |
| `--format <形态>` | json / table / csv | json | 输出形态：json（带环境信封）/ table（人读表格）/ csv（带 `#` 环境头） |
| `--json` | 开关 | — | `--format json` 的语法糖（也是默认） |
| `--csv` | 开关 | — | `--format csv` 的语法糖 |
| `-v, --verbose` | 开关 | — | 显示解析细节（如实际使用的数据源路径） |

参数的含义、示例与细节，以 `tt <命令> --help` 为准（`tt help <命令>` 等价）。
`tdebug` / `tdev` / `tdict` 这三个名字也可以叫（分别等价于 `debug` / `dev` / `dict`）；
`tt env use` 与 `tt env topent` 只打印迁移提示，不在下表。
`tt session` 的目标一律用 `--session` 点名（ref / 别名 / 句柄），池里多条时不替你挑。
`tt serve` 默认后台常驻（单实例），端口取配置 `listen`（缺省 `127.0.0.1:28670`），被占用时顺延。

### `tt debug` — TDebug - T100 作业调试器

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt debug bp <子命令>` | 断点面：list（带命中计数）/ add / del | — |
| `tt debug bp add <位置>` | 下一条断点（位置：行号 / 函数名 / file:line；--kind watch 时是表达式） | `--ignore <int>`<br>`--kind <string>`<br>`--session <string>` |
| `tt debug bp del <编号>` | 删除一条断点（编号见 bp list） | `--session <string>` |
| `tt debug bp list` | 列出断点（每条带命中计数 Hits） | `--session <string>` |
| `tt debug db` | 数据库连接探查:按企业(TOPENT)查账号并尝试连接 | `--ent <int>`<br>`--refresh` |
| `tt debug ents` | 列出当前环境的企业编号(ENT)→ 数据库账号(schema);带快照,离线可答 | `--cached`<br>`--ent <int>`<br>`--refresh` |
| `tt debug env [环境名]` | 查看/切换当前调试环境(SSH 配置,需 serve 在运行) | `--name <string>`<br>`--session <string>`<br>`--timeout <int>=120`<br>`--url <string>` |
| `tt debug exec "<命令>" [更多命令...]` | 透传 fgldb 标准调试命令(print/break/next/where/info/...),可一次多条 | `--file <string>`<br>`--max <int>=2000`<br>`--session <string>`<br>`--timeout <int>=90`<br>`--url <string>`<br>`--wait <int>` |
| `tt debug interrupt` | 中断运行中的程序(回到调试器,等效 SIGINT) | `--session <string>`<br>`--url <string>` |
| `tt debug locate <函数名或模块.函数>` | 定位函数定义到文件:行(需停站;info line) | `--session <string>`<br>`--url <string>` |
| `tt debug logs` | 查看会话最近事件(停站/状态/日志/断开;tail 查询) | `-n/--tail <int>`<br>`--url <string>` |
| `tt debug mode [solo\|collab]` | 查看或切换会话模式:纯人工(solo) / 协作(collab,AI 主导) | `--session <string>`<br>`--url <string>` |
| `tt debug probe` | 调试链路自检:全自动验一遍协议驱动器(登录/断点/命中/源码/栈帧/步进/求值) | `-m/--module <string>`<br>`-p/--prog <string>=bsft001_wf` |
| `tt debug quit` | 结束指定调试会话(作业窗口随之关闭) | `--session <string>`<br>`--url <string>` |
| `tt debug resolve <作业编号>` | 解析作业编号对应的实体程序/模块(gzzz_t;不启动会话) | `-m/--module <string>`<br>`--url <string>` |
| `tt debug serve` | 后台常驻调试服务(tt debug start/exec/status 等控制命令自动发现的守护进程) | `--foreground`<br>`--listen <string>`<br>`--stop` |
| `tt debug source [DVM文件]` | 读取服务器源码(登录区源码目录白名单只读,可指定行段) | `--from <int>`<br>`-m/--module <string>`<br>`--path <string>`<br>`--path-only`<br>`--to <int>`<br>`--url <string>` |
| `tt debug sql "<查询语句>"` | 执行一条只读 SQL(默认 JSON 信封,含环境与 ENT;账号由 TOPENT 决定) | `--ent <int>`<br>`--file <string>`<br>`--timeout <int>`<br>`--url <string>` |
| `tt debug start <作业编号>` | 连接 SSH 并对指定程序启动调试,等到入口停站 | `--ent <string>`（必填）<br>`--env <string>`<br>`--fglserver <string>`<br>`-m/--module <string>`<br>`--name <string>`<br>`--session <string>`<br>`--ssh <string>`<br>`--timeout <int>=120`<br>`--url <string>`<br>`--zone <string>` |
| `tt debug status` | 查看调试服务状态与活动会话 | `--url <string>` |
| `tt debug stop` | 查看当前调试会话的停站现场(state/位置/断点/TOPENT,可复取) | `--session <string>`<br>`--url <string>` |
| `tt debug topent [值]` | 查看/设置会话 TOPENT override(需 serve 在运行,会话空闲可设) | `--clear`<br>`--session <string>`<br>`--url <string>` |
| `tt debug wait` | 阻塞等待会话事件(停站/退出/掉线/看门狗),或等到超时 | `--for <string>=stopped,exit,dead,watchdog`<br>`--session <string>`<br>`--timeout <int>=300`<br>`--url <string>` |
| `tt debug why` | 探测程序此刻在干什么:在等用户操作,还是在空转(慢查询/慢循环) | `--no-resume`<br>`--session <string>`<br>`--url <string>` |
| `tt debug wsdebug <rowid>` | 按接口日志的报文参数启动重放调试,等到入口停站 | `--ent <string>`（必填）<br>`--env <string>`<br>`--fglserver <string>`<br>`--name <string>`<br>`--request-file <string>`<br>`--set <stringArray>`<br>`--ssh <string>`<br>`--timeout <int>=150`<br>`--url <string>` |
| `tt debug wslogs` | 获取接口日志列表(wssp/awsp 报文流水) | `--fail`<br>`--from <string>`<br>`--job <string>`<br>`--origin <string>`<br>`--page <int>=1`<br>`--pid <string>`<br>`--result <string>`<br>`--save-request <string>`<br>`--server <string>`<br>`--service <string>`<br>`--show <string>`<br>`--to <string>`<br>`--url <string>` |

### `tt session` — 会话：建 / 列 / 看 / 改 / 收（会话是一级对象）

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt session ent <企业编号>` | 改会话的企业编号（仅空闲时可设） | `--clear`<br>`--session <string>`<br>`--value <string>` |
| `tt session exec -- "<fgldb 命令>"` | 透传一条 fgldb 命令（等价 tt debug exec） | `--session <string>` |
| `tt session job --prog <作业>` | 在这条会话里起一个作业（会话本来就空闲） | `--ent <string>`<br>`--fglserver <string>`<br>`-m/--module <string>`<br>`--prog <string>`（必填）<br>`--session <string>` |
| `tt session ls` | 列会话 | `-a/--all`<br>`--env <string>` |
| `tt session new` | 建会话（可同时起作业） | `--ent <string>`（必填）<br>`--env <string>`（必填）<br>`-m/--module <string>`<br>`--name <string>`<br>`--prog <string>`<br>`--timeout <int>=120` |
| `tt session quit` | 结束本轮调试（会话保留，回空闲） | `--session <string>` |
| `tt session rm` | 收掉会话（断连接、清资源） | `--session <string>` |
| `tt session show` | 看一条会话的详情（身份 / 作业 / 停站位置） | `--session <string>` |

### `tt dict` — TDict - ERP 数据字典查询工具

组级参数（`tt dict` 的每个子命令都认）：`-d/--db <string>`、`--ent <string>`、`--limit <int>=-1`

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt dict bdldoc <子命令>` | BDL(4GL)语言文档目录(config.json 顶层 bdldoc.dir 查看/设置) | — |
| `tt dict bdldoc dir [<目录>]` | 查看或设置 BDL 文档目录(config.json 顶层 bdldoc.dir) | — |
| `tt dict db <子命令>` | 管理 ERP 数据库连接与数据同步 | — |
| `tt dict db discover` | SSH 自动发现数据库连接要素并保存为连接(--env <环境名>) | `--host <string>`<br>`--save`<br>`--type <string>` |
| `tt dict db list` | 列出各环境的数据库连接(环境名/类型/地址/账号) | `--show-secrets` |
| `tt dict db ping` | 验证环境的数据库可达(--env <环境名>;只读 SELECT version;支持 viaSsh 隧道) | — |
| `tt dict db status` | 查看本地查询数据的覆盖情况(缺什么、各多少行) | — |
| `tt dict db sync` | 从 ERP 刷新本地查询数据 | `--table <string>` |
| `tt dict desc <table_name> [field_name]` | 查询数据库字段对应的前端规格 | `--lang <string>=zh_CN` |
| `tt dict docp <编号>[,<编号>...]` | 查询单据参数说明 | `--lang <string>=zh_CN` |
| `tt dict fieldtab <模块>/<表单>` | 字段表:表单有哪些字段、中文名与类型(读本地 .42f;字典在线/本地两种通道) | `--no-dict`<br>`--root <string>`<br>`--source <string>=online` |
| `tt dict mirror <子命令>` | 管理 T100 本地源码镜像(服务器 4gl/4fd/42s/*.inc → 本地目录) | — |
| `tt dict mirror dir [<目录>]` | 查看或设置本地镜像根目录(config.json 顶层 mirror.dir) | — |
| `tt dict mirror path [<环境名>]` | 打印该环境镜像目录的绝对路径(AI 直接前往) | — |
| `tt dict mirror pull [<环境名>]` | 下载/更新该环境源码镜像(4gl/4fd/42s-zh_CN/*.inc;默认增量;--full 全量重建) | `--full` |
| `tt dict msg [编号[,<编号>...]]` | 查询系统消息 (编号/语句/类型/状态/建议作业 多条件) | `--full`<br>`--lang <string>=zh_CN`<br>`--prog <string>`<br>`--status <string>`<br>`--text <string>`<br>`--type <string>` |
| `tt dict prog [程序编号]` | 查询程序/作业字典 (程序做什么、被哪些作业使用) | `--kw <string>`<br>`--lang <string>=zh_CN`<br>`--sub` |
| `tt dict r.q [开窗码]` | 查询可复用开窗 (r.q) | `--kw <string>`<br>`--lang <string>=zh_CN` |
| `tt dict r.t [table_name]` | 查询数据表字典 (r.t) | `--brief`<br>`--kw <string>`<br>`--lang <string>=zh_CN`<br>`--who` |
| `tt dict r.v [识别码]` | 查询校验带值定义 (r.v) | `--kw <string>`<br>`--lang <string>=zh_CN` |
| `tt dict scc [分类码]` | 查询系统分类码 (SCC) | `--kw <string>`<br>`--lang <string>=zh_CN` |
| `tt dict spill <子命令>` | 查看被截断查询落盘的完整结果（截断时信封里的 localPath） | — |
| `tt dict spill list` | 列出落盘过的结果（新 → 旧） | — |
| `tt dict spill show <文件\|latest>` | 读一份落盘结果（可翻页） | `--offset <int>` |
| `tt dict sysp <编号>[,<编号>...]` | 查询系统参数说明 | `--lang <string>=zh_CN` |

### `tt dev` — T100 设计器包工具（.tzc 代码包 + .tzs 表单包）

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt dev tzc` | .tzc 代码包：export 渲染围栏工作区，apply 三道闸门写回 | — |
| `tt dev tzc apply` | 验证管线全阶段 + 原子写包 + git commit（唯一会改 .tzc 的命令） | — |
| `tt dev tzc export` | 渲染带围栏的 4GL 工作区（只读包） | — |
| `tt dev tzc newfn` | 新增自订定义点（只改 workspace） | — |
| `tt dev tzc rename` | 结构事务：改自订定义点的函数名（只改 workspace） | — |
| `tt dev tzc selftest` | 内置 31 项对抗用例（全部合成包） | — |
| `tt dev tzc status` | 报告工作区相对上次 export/apply 的改动（不写盘） | — |
| `tt dev tzc unlock` | 框架解锁：单向状态迁移，只改 workspace | — |
| `tt dev tzc verify` | 不写盘跑完整验证管线（gate1 + gate2） | — |
| `tt dev tzs export <包.tzs> [-o <目录>]` | 纯解压（不做任何处理，产物是只读参考） | `-o/--out <string>`<br>`--force`<br>`--json` |
| `tt dev tzs formcheck --form42f <表单.42f> --tzs <包.tzs>` | 静态比对 `.42f`（应然）↔ 生成物里的 `.4fd`（实然），不开引擎、不连服务器 | `--form42f <string>`<br>`--tzs <string>`<br>`--json` |
| `tt dev tzs <动词> --args '<JSON 对象>'` | 读写表单的唯一路径（由设计器自己的引擎算），动词面由引擎的函数表生成 | `--form <string>`<br>`--args <string>`<br>`--args-file <string>`<br>`--workspace <string>`<br>`--rpc-timeout <int>`<br>`--json` |
| `tt dev tzs doctor` | 环境自检（引擎产物、设计器目录、工作区、管道名） | `--json` |
| `tt dev tzs stop` | 停掉本工作区的常驻引擎（不启动） | — |
| `tt dev tzs reap` | 清理引擎重编后停不掉的孤儿守护进程 | `--yes` |

### `tt form` — 出厂表单：列出 / 查看字段表 / 与语料核对 / 预热缓存

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt form cache` | 表单缓存概览：多少张、多少字节、最旧/最新（只读，不删任何东西） | — |
| `tt form fetch` | 批量把该模块的表单取进缓存（只缓原始 XML，绝不写服务器） | `--dir <string>`<br>`--force`<br>`--module <string>` |
| `tt form ls` | 列出该模块下的表单名 | `--dir <string>`<br>`--module <string>` |
| `tt form show` | 查看一张表单的字段表（工号 / 编号 / 控件 / 几何 / 必填） | `--dir <string>`<br>`--form <string>`<br>`--module <string>`<br>`--no-cache`<br>`--variant <string>` |
| `tt form verify` | 与语料核对：离线字段表 vs 归档语料表（逐条，必须 100%） | `--dir <string>`<br>`--form <string>`<br>`--map <string>`<br>`--module <string>`<br>`--no-cache`<br>`--window <string>` |

### `tt drawio` — 用 T100 组件库画原型图（.drawio）

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt drawio compose <spec.json> [-o <文件>]` | 把排版规格 spec 展开成 .drawio | `-o/--out <string>` |
| `tt drawio lib` | 把形状库导出成 drawio 能加载的文件 | `--catalog`<br>`-o/--out <string>` |

### `tt env` — 管理 SSH 环境（三个工具共用同一份环境清单）

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt env list` | 列出全部环境（同不带参数） | — |
| `tt env show <环境名>` | 显示某个环境的连接详情（口令打码） | — |

### `tt serve` — 启动本地 Web 服务（工作站 + 统一设置页）

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt serve` | 启动本地 Web 服务（工作站 + 统一设置页） | `--foreground`<br>`--listen <string>`<br>`--stop` |

### `tt install` — 把 tt 装进你的环境（skills / path）

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt install path` | 把 exe 所在目录加入用户 PATH（HKCU，不需要管理员） | `--dry-run` |
| `tt install skills` | 把 exe 同目录的 skills/ 复制到目标目录 | `--agent <string>`<br>`--force`<br>`--to <string>` |

### `tt update` — 更新 tt 自身（查新版、下载校验、交棒给更新器）

组级参数（`tt update` 的每个子命令都认）：`--check-only`、`--dry-run`、`--pre`、`--proxy <string>`、`--yes`

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt update check` | 查有没有新版（不下载、不安装） | `--pre`<br>`--proxy <string>` |
| `tt update log` | 回读最近一次升级的阶段与结果 | `--tail <int>=20` |

### `tt config` — 配置管理：位置 / 查看 / 读写 / 校验

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt config get <键路径>` | 按点分路径读一个值（如 hosts.sshs、debug.termWidth） | `--raw` |
| `tt config path` | 打印配置文件的解析结果 | — |
| `tt config set <键路径> <值>` | 按点分路径写一个值（值按 JSON 解析，解析不了当字符串） | — |
| `tt config show` | 打印配置内容（口令默认打码，--raw 原样输出） | `--raw` |
| `tt config validate` | 校验配置：能否解析、环境是否完整、缺省值是否被补全 | — |

### `tt cache` — 查看/清理缓存(企业快照、源码镜像、查询落盘、断点存档)

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt cache clear` | 清理缓存(默认全清;--older-than 只清旧的) | `--dry-run`<br>`--older-than <duration>` |

### `tt version` — 显示版本信息

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt version` | 显示版本信息 | — |

### `tt completion` — 生成 shell 补全脚本

| 命令 | 说明 | 参数 |
|---|---|---|
| `tt completion bash\|zsh\|fish\|powershell` | 输出对应 shell 的补全脚本（`tt completion <shell> --help` 里有安装方法） | — |
