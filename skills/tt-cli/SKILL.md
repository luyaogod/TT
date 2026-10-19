---
name: tt-cli
description: TT CLI（T100 开发命令行）的全量命令清单与选路指南：全局 --at/--session 寻址、env 环境、db 连通性与只读 SQL、dict 数据字典、file 服务器文件与源码镜像、session/debug 调试、dev tzc/tzs 设计器包、form 表单、drawio、以及 serve/install/update/config/cache。当用户问“tt 有哪些命令 / 这件事该用哪条命令 / 参数怎么写”时先读本技能；输出一律简体中文。
---

# TT CLI — 命令清单

本技能是**命令地图**：回答“有什么命令、怎么点名、参数怎么写、先看哪里”。
深入用法见同名技能：调试看 `tt-debug`，字典看 `tt-dict`，代码包看 `tt-dev-tzc`，
表单包看 `tt-dev-tzs`，读源码看 `tt-erp-read`，原型图看 `tt-drawio`。

## 0. 三条硬规则

1. **目标一律点名**：环境用 `--at <环境>` 或 `--env <环境>`；企业用 `--at <环境>@<企业>` 或 `--ent`；
   会话用 `--session`。**没有默认环境**，不点名就报错并列出可用环境。
2. **不知道填什么先问工具**：`tt <任意命令> --at-list` 会列出该命令可用的环境/企业、来源与示例；
   它是纯本地操作（读 `config.json` + ENT 清单），不连 SSH、不连库。
3. **服务器只读**：看文件用 `tt file`（浏览/读取/搜索走纯 SFTP，服务器上不起进程）；不要在服务器上 `ssh/scp`。
   通用文件命令没有 `put/rm`。写服务器只发生在有明确闸门的通道：`tt dev tzc apply`（三道闸门）、
   `tt dev tzs` 的引擎动词（工作区/包约束）、`tt debug wsdebug` 重放时的 `$TEMPDIR` 临时报文、
   以及 `tt file mirror pull` 在 `$TOP` 下写的固定命名打包临时文件（`.tdict-mirror-*`）。

## 1. 全局约定

### 1.1 目标：`--at`、`--at-list`、`--session`

- `--at <环境>`：只认环境。适用于 `tt file …`、`tt form …`、`tt debug wslogs`、`tt dict db …` 等。
- `--at <环境>@<企业>`：环境 + 企业编号。适用于要按企业解析数据库账号的命令：
  `tt db sql/probe`、`tt dict` 的查询、`tt debug start/wsdebug`、`tt session new`。
- 带 `@企业` 传给只认环境的命令会被明确拒绝（例如 `tt file` 会提示去掉 `@99`）。
- 会话类命令（`tt session …`、`tt debug exec/interrupt/wait/…`）用 `--session` 点会话；
  这里的 `--at` 只用来收窄“在哪个环境里找”，**不带企业**。
- `--session` 的取值：别名（`复现-1`）、规范串（`恒一#2`）、`环境/作业`、`环境/作业#序号`、
  `环境/企业编号`、单写环境（该环境唯一那条）。命中多条会报歧义，不会“随便挑一条”。
- `--at` 与 `--env` 给了两个不同环境会报错；`--at` 与企业 `--ent` 不一致也会报错。

### 1.2 输出

- 数据类命令（dict / db / env / session）默认 JSON 信封，含环境、企业、只读标记与行数；
  debug 视命令而定（`exec` 是原生输出，`status/logs` 等是信封）。
- `tt file` 默认**纯文本**（Unix 习惯）；`ls/stat/find/get` 支持 `--json`。
- 全局可用 `--json`、`--csv`、`--format json|csv|table`、`-v/--verbose`。
- 表格/CSV 的头部会带 `# 环境 …` 注释；脚本用 `grep -v '^#'` 过滤。

### 1.3 退出码

各命令线自己的码表在对应章节；通用约定：

- `0` 成功。
- `1` 用法/参数错；`tt file grep/find` 的“无匹配”也是 1；`tt form` 取不到表单也是 1。
- `2` `--at`/`--env`/`--ent` 等**目标**填错（环境不存在、企业不在清单、`--at` 与 `--env` 冲突），
  以及 `tt db` 的用法/身份错。
- `3` `tt db` 只读闸门拒绝；`tt dict` 缺表；`tt dev tzc` 验证失败。
- `4` 连接/传输失败（`tt db`、`tt file`）；`tt dev tzc` 写入被拒；`tt update` 拒绝自装。
- `5` 本地写失败（`tt file get`）、`tt dev tzc` IO/环境失败、`tt update` 出网/落盘失败。
- `6` `tt form verify` 与语料不一致；`tt dev tzs formcheck` 有确定差异。
- `10` `tt update check` 发现有可用更新。

### 1.4 配置、缓存与 ENT 清单

- 配置位置：`TT_CONFIG` 环境变量 → `--config` → 便携包（exe 同目录 `.portable`）→
  用户目录 `%APPDATA%\tt\config.json`。用 `tt config path` 看最终结果。
- 环境清单在 `hosts.sshs`；**ENT 清单是静态配置**（`hosts.sshs[].ents` + `entsSyncedAt`），
  默认不联网。同步用设置页「ENT」页签的“从服务器刷新”，或 `tt env ents --refresh`。
- 缓存目录（`tt cache clear` 清理）：`ents/`、`srccache/`、`execlog/`、`debug-bps/`、`spill/`、
  `update/`、`forms/`、`runtime/`。`runtime/` 是 `$TOP/$ERP/$COM` 的登录探测缓存（30 分钟，
  `tt file --refresh` 强制重探）。

## 2. 工具管理

### 2.1 `tt serve`

启动本地 Web 服务（调试工作台 + 统一设置页），默认后台常驻、单实例，打印地址后返回。

```
tt serve
tt serve --foreground
tt serve --listen 127.0.0.1:28670
tt serve --stop
```

- 默认地址 `http://127.0.0.1:28670`（端口占用自动顺延）。
- 设置页覆盖环境、ENT、数据库账号、源码镜像、缓存等；与 CLI 共用同一份 `config.json`。
- 重复执行不会起第二个实例，只提示已在运行。

### 2.2 `tt install`

```
tt install path [--dry-run]
tt install skills [--agent <名>] [--to <目录>] [--force]
```

- `path`：把 exe 所在目录加入**用户 PATH（HKCU）**，不需要管理员；永不碰系统 PATH。
- `skills`：把 exe 同目录的 `skills/` 复制到 agent 技能目录；`--force` 覆盖。
  装过哪些地方记在数据目录 `.tt-skills.json`。

### 2.3 `tt update`

查新版、下载并校验 sha256、交棒给独立更新器。默认会问一句再装；agent/脚本必须 `--yes`。

```
tt update check              # 只查；有新版退 10
tt update                    # 查 + 下载 + 校验，问一句再装
tt update --dry-run          # 走到交棒前停
tt update log [--tail N]     # 回读最近一次升级日志
```

组级开关：`--check-only`、`--dry-run`、`--pre`、`--proxy <url>`、`--yes`。

- 便携包就地覆盖程序目录（`config.json` 与用户文件不碰）；MSI 交给 `msiexec /qb`（用户级）。
- 源码态/认不出的布局只允许查，装会被拒（退 4）。

### 2.4 `tt config`

```
tt config path
tt config get <键路径> [--raw]
tt config set <键路径> <值>
tt config show [--raw]
tt config validate
```

- 键路径用点分写法，如 `hosts.sshs`、`debug.termWidth`。
- `set` 的值按 JSON 解析，解析不了当字符串。
- `show` 默认打码口令；`--raw` 原样输出（小心泄漏）。

### 2.5 `tt cache`

```
tt cache clear [--dry-run] [--older-than <duration>]
```

- 默认全清；`--older-than 168h` 只清旧的。只清缓存目录，不动 `config.json`。

### 2.6 `tt version` / `tt completion`

```
tt version
tt completion bash|zsh|fish|powershell
```

- `tt completion <shell>` 输出补全脚本；安装方法见 `--help`。

## 3. 环境：`tt env`

### 3.1 列表与详情

```
tt env list
tt env show <环境名>
```

- 共用同一份 `hosts.sshs`（调试 / 字典 / 数据库 / 文件）。
- `show` 显示连接详情，口令打码。

### 3.2 企业清单：`tt env ents`

```
tt env ents [<环境名>]
tt env ents --ent 99
tt env ents --refresh
```

- 默认**只读 `config.json` 的 ENT 清单**，不联网、离线可用。
- `--refresh` 连库读 `gzou_t` 并**写回配置**（同步）；固定用系统账号 `ds` 读。
- 同步时把清单里的账号补进该环境的 `db.accounts`（缺的才补；密码优先已有条目，
  否则按 T100 惯例“账号=密码”）。输出会点名提示补了哪些。
- 某个环境同步失败只影响它自己；`--at-list` 里也能看到“某环境没有 ENT 清单”的提示。

### 3.3 已移除的命令

`tt env use`、`tt env topent` 已移除（只剩提示）：环境一律 `--env/--at` 点名，
企业一律 `--ent`/`--at @企业` 逐个给，不存在“默认环境/环境级 TOPENT”。

## 4. 数据库：`tt db`

只做两件事：**连不连得上**、**跑一条只读 SQL**。字典内容查询（`r.t`/`desc`/`prog` …）在 `tt dict`。

退出码：`0` 成功；`2` 用法或身份错；`3` 只读闸门拒绝；`4` 连接或查询失败。

### 4.1 `tt db ping` — 本机连库

```
tt db ping --at 主机109
tt db ping --at 主机109@99
```

- 客户端直连（支持 `viaSsh` 隧道）；企业可选，不给就用该环境账号清单首项。
- `--ent <int>` 也可代替 `@企业`。

### 4.2 `tt db probe` — 服务器连库

```
tt db probe --at 主机109@99
```

- SSH 登录服务器，用该企业的账号跑一次 `sqlplus`/`ksql`，验证“服务器 → 库”。
- 企业必填：验的就是这个企业的账号。

### 4.3 `tt db sql` — 只读 SQL

```
tt db sql --at 主机109@99 "select 1 from dual"
tt db sql --at 主机109@99 --file query.sql --timeout 60
```

- 只允许单条 `SELECT`/`WITH`；只读事务；最多 200 行。
- 直连 SSH + `sqlplus`/`ksql`，**不依赖 serve**。
- `--ent <int>`、`--file <路径>`、`--timeout <秒，默认 30，上限 120>`。
- 账号来自 `config.json` 的 ENT 清单；清单里没有该企业就报错，**不会现查 `gzou_t` 猜账号**。
- 信封里带 `accountSource`（`ent(config)` / `ent(gzou_t)`）如实回显账号来源。

## 5. 数据字典：`tt dict`

查询对象是 **ERP 的业务元数据**（表、字段、校验、开窗、消息、参数、程序/作业），
不是业务数据本身。深入用法见 `tt-dict` 技能。

### 5.1 数据源

- `--env <环境>` 或 `--at <环境>`：走该环境 ERP 库直查（最新最全，不需要先同步）。
- `--env local`：走本地 SQLite 副本（`erp_data.db`）；要 `tt dict db sync` 才更新。
- **不自动切换**：在线连不上直接报错，不会偷偷查本地。
- 不点名时按 `config.json` 顶层 `query.source` 配置走；没有就本地副本。

### 5.2 表 / 字段 / 校验 / 开窗

```
tt dict r.t [table_name] [--brief] [--kw <词>] [--who]
tt dict desc <table_name> [field_name]
tt dict r.v [识别码] [--kw <词>]
tt dict r.q [开窗码] [--kw <词>]
tt dict fieldtab <模块>/<表单> [--source online|local] [--root <目录>] [--no-dict]
```

- `r.t --brief` 批量解码表名；`r.t --who` 反查“这张表被哪些程序用”。
- `--lang zh_CN` 可切语言（默认简体中文）。
- `fieldtab` 读本地 `.42f` 的字段表；`--source local` 用镜像/本地目录。

### 5.3 系统字典

```
tt dict scc [分类码] [--kw <词>]
tt dict docp <编号>[,<编号>...]
tt dict sysp <编号>[,<编号>...]
tt dict msg [编号...] [--text <词>] [--type <类型>] [--status <状态>] [--prog <程序>] [--full]
```

### 5.4 程序与作业

```
tt dict prog [程序编号] [--kw <业务词>] [--sub]
```

- `prog <编号>`：作业名、类别、归属模块、被哪些作业使用、用了哪些表（含操作类别）。
- `prog --kw <词>`：从业务词反查程序；`--sub` 查子程序/元件。
- 子程序/库/开窗也能查（`gzde_t` 独立登记，别把主程序的用途当成子程序的）。

### 5.5 本地库维护

```
tt dict db list [--show-secrets]
tt dict db discover [--env <环境>] [--host <主机>] [--type oracle|kingbase] [--save]
tt dict db status
tt dict db sync [--table <表>]
```

- `list`：列各环境的库连接（环境/类型/地址/账号）。
- `discover`：登录 SSH 探库要素；`--save` 写回配置。
- `status`：本地副本覆盖情况（缺哪族、各多少行）。
- `sync`：从 ERP 刷新本地副本；`--table` 只同步一张。

### 5.6 文档与落盘

```
tt dict bdldoc dir [<目录>]
tt dict spill list
tt dict spill show <文件|latest> [--offset <int>]
```

- `bdldoc`：BDL（4GL）语言文档目录，存在 `config.json` 顶层 `bdldoc.dir`。
- `spill`：查询被截断时，完整结果已落盘；信封里的 `localPath` 指向它。

### 5.7 组级开关

`-d/--db <路径>`（本地 SQLite）、`--ent <编号>`（企业，决定账号）、`--limit <int>`（默认跟随配置，
再缺省 20）。查询命令加 `--at <环境>[@<企业>]` 是推荐写法。

## 6. 服务器文件：`tt file`

只读 `$TOP`；浏览/读取/搜索走纯 SFTP，服务器上不执行命令；没有 `put/rm/mv/chmod`。
路径相对 `$TOP`；绝对路径也必须在 `$TOP` 内；符号链接默认不跟随，`--follow` 也只允许指向 `$TOP` 内。
（例外：`mirror pull` 是固定的服务器侧打包——只读 `find`/`tar`，并在 `$TOP` 下留打包临时文件。）

退出码：`0` 成功；`1` 无匹配或无此文件；`2` 用法或越界；`4` 连接/传输失败；`5` 本地写失败。

组级开关：`--follow`（跟随软链）、`--refresh`（重探 `$TOP`）、`--binary`（允许输出二进制）。

### 6.1 浏览与读取

```
tt file ls    [-la] [--human] [<路径>]        # 默认 $TOP
tt file cat   <文件...>                       # 多文件直接拼接
tt file head  [-n N] <文件...>                # 默认 10 行
tt file tail  [-n N] <文件...>
tt file stat  <路径...>
tt file wc    [-l] <文件...>
tt file du    [-s] [--human] [<路径>]
```

- `ls -l`：权限位/大小/时间/名字（不显示属主名）；`-a` 含隐藏；`--human` 人性化大小。
- `cat` 遇到含 NUL 的文件默认拒绝（防止打烂终端/上下文）；确要输出加 `--binary`。
- `head/tail` 多文件时按 GNU 风格打印 `==> 文件 <==`；`tail` 只读文件尾部块，不整份下载。
- `stat`：类型/大小/mtime/权限/软链目标；`--json` 可用。
- `wc`：行/词/字节；`du` 按文件大小累计（1K 块，`--human` 可读）。

### 6.2 搜索

```
tt file grep [-rinl] [-m N] [--include '*.4gl'] <模式> <路径...>
tt file find <路径...> [-name <模式>] [-type f|d|l] [-maxdepth N]
```

- `grep` 用 Go 的 **RE2** 语法；**必须给路径**（不会默认扫整个 `$TOP`）；`-r` 递归；
  `-n` 行号；`-l` 只列文件；`-m` 每文件最多匹配数；`--include` 只扫基名匹配的文件。
- 二进制文件按 NUL 判断后跳过；单文件超过 64 MB 跳过并提示。
- `find` 输出相对 `$TOP` 的路径，可直接回喂给本组命令；`-maxdepth 0` 只看起点；
  不接受 `-exec/-delete/-fprint`。Unix 单横线写法 `-name/-type/-maxdepth` 也可用（内部归一化）。
- 二者都在客户端实现，服务器上零进程；`grep` 12 路并发读、输出按路径排序。

### 6.3 下载

```
tt file get <路径> [-o <本地|->] [--max <MB>]
```

- 默认落到当前目录同名文件；`-o -` 写 stdout；`--max` 默认 200 MB（0 = 不限）。
- 先写 `.part` 再 rename，中断不留半截文件；服务器侧始终只读。

### 6.4 离线源码镜像

```
tt file mirror dir [<目录>]                  # 全局镜像根（config.json 顶层 mirror.dir）
tt file mirror path [<环境>]                 # 打印该环境镜像目录（第一行即路径）
tt file mirror pull [<环境>] [--full]        # 增量更新；--full 全量重建
```

- `dir` 不需要环境；`path/pull` 用 `--at <环境>` 或位置参数点名。镜像根必须显式设置（`dir <目录>`）。
- 只镜像白名单：各模块 `4gl/`、`4fd/`、`42s/zh_CN/`、`*.inc`；不含 `per/`、编译产物
  （`42f/42m/42r`）、其它语言、备份文件。
- 用途：断网时 AI 用本地工具读码；有网时优先 `tt file cat/grep` 现读。
- `tt dict mirror …` 只剩隐藏转发（会打印迁移提示），新写法是 `tt file mirror …`。

### 6.5 典型用法

```
tt file ls -l --at 主机109 erp/apm/4gl
tt file grep -rn --at 主机109 'apmt520' erp/apm/4gl
tt file head -n 60 --at 主机109 erp/apm/4gl/apmi004.4gl
tt file find --at 主机109 erp/apm -name '*apmi004*'
tt file cat --at 主机109 erp/apm/42s/zh_CN/apmi004.42s
tt file get --at 主机109 per/apm/apmt520.rpt -o ./apmt520.rpt
tt file mirror pull --at 主机109
```

## 7. 会话：`tt session`

会话是一级对象：一条 SSH 登录 + 当前作业 + 停站现场。`--session` 点会话，`--at` 只收窄环境。

### 7.1 建会话

```
tt session new --at 主机109@99                     # 空闲会话（已登录，无作业）
tt session new --at 主机109@99 --prog apmt520       # 同时起作业并等到入口停站
tt session new --at 主机109@99 --prog apmt520 --name 复现-1
```

- `--ent`/`--env` 必填（或用 `--at` 一次给全）；企业决定账号，不再沿用登录 shell 的 `$TOPENT`。
- 给了 `--prog` 会自动起作业；`--name` 是跨轮追踪用的别名（不能改名）。
- 其它开关：`-m/--module`、`--timeout <秒，默认 120>`。

### 7.2 查看与收尾

```
tt session ls [-a|--all] [--env <环境>]
tt session show [--session <目标>]
tt session ent <企业编号> [--session <目标>] [--clear]
tt session quit [--session <目标>]      # 结束本轮调试，会话保留、回空闲
tt session rm [--session <目标>]        # 收掉会话（断连接、清资源）
```

### 7.3 在会话里操作

```
tt session job --prog <作业> [--session <目标>] [--ent <编号>] [-m <模块>] [--fglserver <值>]
tt session exec -- "<fgldb 命令>" [--session <目标>]
```

- `job`：会话空闲时起一个作业。
- `exec`：透传一条 fgldb 命令（等价 `tt debug exec`）。

## 8. 调试：`tt debug`

调试线的深入用法（断点、停站现场、人机协作、接口日志重放）见 `tt-debug` 技能。
这里是命令清单与关键开关。

### 8.1 服务与会话生命周期

```
tt debug serve [--foreground|--listen|--stop]
tt debug start <作业编号> --at <环境>@<企业> [--module …] [--name …] [--session …]
tt debug status [--url <地址>]
tt debug quit [--session <目标>]
tt debug env [环境名] [--name <环境>|--session <目标>]
```

- `debug start` 会自动发现本地 serve；没有就报错（先 `tt serve`）。
- `start` 的 `--ent` 必填；`--zone`/`--ssh`/`--fglserver`/`--timeout` 可按需覆盖。
- `debug status` 不认 `--session`，因为它列全部会话。

### 8.2 执行、等待与现场

```
tt debug exec "<命令>" [更多命令...] [--session] [--file] [--max 2000] [--wait] [--timeout 90]
tt debug interrupt [--session]
tt debug wait [--for stopped,exit,dead,watchdog] [--timeout 300] [--session]
tt debug stop [--session]          # 读停站现场（state/位置/断点/TOPENT）
tt debug why [--no-resume] [--session]
tt debug logs [-n/--tail N] [--url]
tt debug locate <函数名或模块.函数> [--session]
```

- `exec` 透传 fgldb 标准命令（print/break/next/where/info/…）；`--file` 从文件读命令。
- `wait` 阻塞到事件或超时；`--for` 默认 `stopped,exit,dead,watchdog`。
- `why` 判断“在等用户操作”还是“在空转（慢查询/慢循环）”。

### 8.3 断点

```
tt debug bp list [--session]
tt debug bp add <位置> [--kind <类型>] [--ignore <次数>] [--session]
tt debug bp del <编号> [--session]
```

- 位置可以是行号、函数名、`file:line`；`--kind watch` 时是表达式。
- 每条断点带命中计数（Hits）。

### 8.4 模式与企业

```
tt debug mode [solo|collab] [--session]
tt debug topent [值] [--clear] [--session]
```

- `solo`：纯人工；`collab`：AI 主导协作。
- `topent`：会话级 TOPENT override（空闲时可设）；企业优先用 `--at @企业`/`--ent` 显式给。

### 8.5 源码、作业与企业清单

```
tt debug source [DVM文件] [-m <模块>] [--path <路径>] [--from N] [--to N] [--path-only] [--url]
tt debug resolve <作业编号> [-m <模块>] [--url]
tt debug ents [--ent N] [--cached] [--refresh]
tt debug probe [-m <模块>] [-p <作业>=bsft001_wf]
```

- `source`：读登录区白名单源码（可指定行段）；需要 serve。
- `resolve`：作业编号 → 实体程序/模块（`gzzz_t`），不启动会话。
- `debug ents`：问当前环境的企业→账号（快照/现查）；日常清单改用 `tt env ents`（读配置），
  只读 SQL 改用 `tt db sql`。
- `probe`：调试链路全自动自检（登录/断点/命中/源码/栈帧/步进/求值）。

### 8.6 接口日志

```
tt debug wslogs [--job <作业>] [--service <服务>] [--pid <进程>] [--fail] [--page N]
                [--from <时间>] [--to <时间>] [--show <字段>] [--save-request <文件>]
tt debug wsdebug <rowid> --at <环境>@<企业> [--request-file <文件>] [--set k=v] [--name <别名>]
```

- `wslogs`：列 wssp/awsp 接口报文流水；`--save-request` 把请求报文取回本地。
- `wsdebug`：按日志里的报文参数启动重放调试，等到入口停站。

### 8.7 已迁移

`tt debug sql` / `tt debug db` 已隐藏，执行会提示迁移；新命令是 `tt db sql` / `tt db probe`。

## 9. 设计器包：`tt dev`

### 9.1 `.tzc` 代码包：`tt dev tzc`

把 `.tzc` 渲染成带围栏的 4GL 工作区，编辑后用 `apply` 走三道闸门写回（唯一会改 `.tzc` 的命令）。

```
tt dev tzc export <pkg.tzc> [-o <目录>] [--only <点名>...] [--json]
tt dev tzc status [<工作区>] [--json]
tt dev tzc verify [<工作区>] [--strict] [--json]
tt dev tzc apply  [<工作区>] [-o <pkg.tzc>] [--dry-run] [--yes] [--json]
tt dev tzc unlock [<工作区>] [--yes] [--json]
tt dev tzc rename [<工作区>] <旧函数名> <新函数名> [--scope PUBLIC|PRIVATE] [--desc <描述>] [--json]
tt dev tzc newfn  [<工作区>] --type FUNCTION|DIALOG|REPORT [--name <名>] [--scope …] [--json]
tt dev tzc selftest [--json]
```

- 工作区目录可省略：先 `cd` 进工作区，命令直接在当前目录操作。
- `unlock` 不可逆（解开后规格调整不再自动生成程序代码），必须显式 `--yes`；AI 不得自主加。
- `apply` 是唯一会写 `.tzc` 的动词；`--dry-run` 只出 diff。
- 退出码：`0` 成功；`2` 包格式错；`3` 验证失败；`4` 写入被拒；`5` IO/环境失败。
- 深入用法见 `tt-dev-tzc` 技能。

### 9.2 `.tzs` 表单包：`tt dev tzs`

读写表单由设计器**自己的引擎**算，不是拼 XML。固定子命令：

```
tt dev tzs export <包.tzs> [-o <目录>] [--force] [--json]
tt dev tzs formcheck --form42f <表单.42f> --tzs <包.tzs|表单.4fd> [--json]
tt dev tzs doctor [--json]
tt dev tzs stop
tt dev tzs reap [--yes]
```

动态动词（约 55 个，由引擎函数表生成）：

```
tt dev tzs <动词> [--form <程序名>] --args '<JSON 对象>' [--args-file <UTF-8 文件|->] \
    [--workspace <目录>] [--rpc-timeout <秒>] [--json]
```

- 参数**只用 JSON 给**；中文/长内容用 `--args-file`（值是 `-` 读 stdin）。
- `--form` 是“哪一张已打开的表单”（程序名或 `程序|Form`），不需要搬运句柄。
- 常见动词：`open`、`form_tree`、`find_component`、`get_component`、`list_spec_nodes`、
  `describe_kind`、`set_spec_attr`、`set_spec_attrs`、`set_layout_attr`、`add_widget`、
  `add_field`、`nudge`、`validate`、`save`、`close`。全名以 `tt dev tzs --help` 为准
  （引擎可达时附在帮助末尾）。
- `formcheck` 退出码：`0` 一致（含只有未判定）；`2` 参数或读取错；`6` 有确定差异。
- 深入用法见 `tt-dev-tzs` 技能。

## 10. 出厂表单：`tt form`

按需取一张出厂表单（`.42f`，编译后的明文 XML）并读成字段表。

```
tt form ls     --at <环境> --module <模块> [--dir <目录>]
tt form show   --at <环境> --module <模块> --form <表单名> [--variant <值>] [--no-cache] [--json]
tt form fetch  --at <环境> --module <模块> [--force] [--dir <目录>] [--json]
tt form verify --at <环境> --module <模块> --form <表单名> --map <语料.json> [--window <值>] [--json]
tt form cache  [--json]
```

- 取数三条路：`--dir`（本地目录/镜像，离线）→ 环境 SSH 按需单取（只读：`find` + SFTP get）→
  `<数据目录>/forms/` 缓存。
- `fetch` 批量取进缓存；`cache` 只看概览（多少张/多少字节/最旧最新），不删任何东西。
- 退出码：`0` 一致；`1` 表单取不到或不是合法 `.42f`；`2` 用法错或无基线；
  `6` `verify` 与语料不一致。

## 11. 原型图：`tt drawio`

```
tt drawio compose <spec.json> [-o <文件>]
tt drawio lib [--catalog] [-o <文件>]
```

- `compose`：把排版规格 spec 展开成 `.drawio`。
- `lib`：导出内置 T100 形状库；`--catalog` 只出目录。
- 深入用法见 `tt-drawio` 技能。

## 12. 速查：该用哪条

- 想知道“这张表/字段/消息/参数是什么” → `tt dict`（表用 `r.t`，程序用 `prog`）。
- 想连库、跑一条只读 SQL → `tt db sql`；只想验通不通 → `tt db ping` / `tt db probe`。
- 想读服务器上的源码/文件 → `tt file`（在线）；断网 → `tt file mirror path` + 本地工具。
- 想调试作业、看停站现场、透传 fgldb → `tt session` + `tt debug`。
- 想改 `.tzc` / 读写 `.tzs` → `tt dev tzc` / `tt dev tzs`。
- 想看表单字段表、与语料核对 → `tt form`。
- 想画原型图 → `tt drawio`。
- 想改配置/缓存/升级/装技能 → `tt config` / `tt cache` / `tt update` / `tt install`。

## 13. 不要做

- 不要 `ssh`/`scp` 去服务器读文件；用 `tt file`。不要在服务器上执行写命令。
- 不要猜环境或企业；不点名就报错，先 `tt env list` / `--at-list`。
- 不要把 `--at <环境>@<企业>` 传给只认环境的命令（`tt file`、`tt form` 等）。
- 不要把 `tt dict` 的“查不到数据”当成“数据不存在”：先核对环境与企业编号。
- 不要用 `cat` 整份读大文件；先 `grep`/`head`/`find` 收窄。
