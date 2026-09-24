# TT 项目 wiki

> 装 tt、构建 tt 的步骤在 [README.md](../README.md)；每个命令组**怎么用**在
> [skills/](../skills/) 下对应的 `SKILL.md`。这里放的是**设计、契约、依据**：
> 为什么这么写、哪里不能碰、判决标准是什么 —— 以及配置字段的完整说明。

TT 是一款面向 Agent 的 CLI 开发工具，用于开发基于 Genero BDL 技术栈的大型 ERP 系统——鼎捷数智旗下的 T100。

```
tt debug …     作业调试器：SSH 驱动 fglrun -d 的 (fgldb) 协议 + 本地 Web 工作台
tt dev tzc …   设计器代码包（.tzc/.tzf/.tzx）：渲染带围栏的 4GL 工作区，apply 写回
tt dev tzs …   设计器表单包（.tzs/.tzv）：由设计器自己的引擎读写，不是拼 XML
tt dict …      ERP 数据字典查询：表/字段/校验/分类码/开窗/消息/参数/程序
tt env / config / serve / install / version
```

## 0. 章节索引

| 章 | 内容 | 什么时候看 |
|---|---|---|
| [1. 总览](#1-总览) | 三块能力的边界、目录布局、术语 | 刚接手时 |
| [2. 统一配置](#2-统一配置) | 单一 `config.json`、位置解析、旧配置迁移、唯一写入口 | 动配置层之前 |
| [3. 环境、账号与输出](#3-环境账号与输出规范) | **SSH 登录 → 环境变量 → 企业 → 账号** 的一条链、谁管什么、查询输出的契约 | **动 `host`/`dbconfig`/`erpdb`/`entdir`/`output` 之前必读** |
| [4. 统一 Web 服务](#4-统一-web-服务) | 单进程单端口、API 前缀为什么这么分、共用的配置端点 | 动 `internal/web` 或前端之前 |
| [5. `tt debug` 调试器](#5-tt-debug-调试器) | 命令面、会话模型、环境变量、运行时产物 | 用或改调试器时 |
| [6. `tt dev tzc` 代码包](#6-tt-dev-tzc-代码包) | 三条公理、zip 保真、围栏协议、三道闸门、结构事务、退出码、**偏差与不变量依据** | 改 `.tzc` 管线时**必读** |
| [7. `tt dev tzs` 表单包](#7-tt-dev-tzs-表单包) | 引擎与设计器程序集、49 个函数、句柄与基线规则 | 用或改表单读写时 |
| [8. `tt dict` 数据字典](#8-tt-dict-数据字典) | 各查询命令的返回与读法、数据源切换、类型码 | 查 ERP 元数据时 |
| [9. 测试与验收](#9-测试与验收) | 每层的判据、语料回归、S7 真机清单 | 交付前 |
| [10. 构建与发布](#10-构建与发布) | 只留"为什么"，步骤见 README | 发版时 |
| [11. 设计史与取舍](#11-设计史与取舍) | 合并动机、发现的分叉、修掉的 bug、已废弃的东西 | 想知道"为什么是这样"时 |
| [12. 许可与出处](#12-许可与出处) | 第三方材料与来源 | 引用/分发前 |

两条贯穿全篇的红线：

1. **tt 不实现设计器的私有格式。** `.tzc` 靠设计器发行物反推；`.tzs` 更彻底 —— 反射调用
   设计器自己的程序集，格式那部分是设计器自己在算。
2. **不破坏要能机械证明。** 凡是声称"安全"的地方，本文都给出判据与复现命令，而不是形容词。

---

## 1. 总览

### 1.1 三块能力与它们的历史边界

| 命令组 | 原项目 | 职责 | 规模（合并时） |
|---|---|---|---|
| `tt debug` | TDebug | T100 作业调试器：SSH 驱动 `fglrun -d` 的 `(fgldb)` 文本调试协议，本地 Web 调试界面（源码/断点/调用栈/变量/接口日志）+ 命令行控制端 | Go 16.3k 行 + TS 6.4k 行 |
| `tt dev` | TDev | 设计器 `.tzc`/`.tzs` 包的安全编辑管线（export/status/verify/apply/unlock/rename/newfn） | Go 18.1k 行 |
| `tt dict` | TDictCli | ERP 数据字典查询（`r.t`/`r.v`/`desc`/`scc`/`r.q`/`prog`…）、DB 同步、源码镜像、BDL 文档 | Go 12.6k 行 + TS 1.5k 行 |

四块能力里，`tt dev tzs` 是合并之后新长出来的第四块，不在上面这张表里。

### 1.2 目录布局

```
TT/
├─ main.go                  //go:embed all:web/dist → cli.Execute
├─ README.md                使用者与构建者的入口（安装/构建/配置字段）
├─ docs/WIKI.md             本文
├─ docs/T100设计器-README.md 第三方材料的恢复副本（见 §12）
├─ engine/                  ★ .tzs 引擎（C#，单独构建）
│  ├─ designer/             设计器的 28 个程序集（随仓库分发）
│  └─ src/ test/            我们的两万行 C#；探测程序只供开发，不进包
├─ installer/tt.wxs         MSI 定义
├─ tools/zip.py             打包 zip（递归，见 §10）
│
├─ internal/
│  ├─ config/               ★ 统一配置层
│  │  ├─ paths.go           位置解析：TT_CONFIG / .portable / %APPDATA%\T100\tt
│  │  ├─ cfgfile.go         Open / Save(原子写) / Edit —— 单一写入口
│  │  ├─ schema.go          配置根结构体 + 各节类型
│  │  ├─ pathutil.go        路径型取值的解析与派生状态
│  │  └─ migrate.go         旧配置迁移与合并（tdebug + tdict → tt）
│  ├─ cli/                  cobra 根命令
│  │  ├─ env.go config.go serve.go install.go version.go
│  │  ├─ common/            各命令组共享的 CLI 上下文（叶子包）
│  │  └─ debug/ dev/ dict/  三个命令组
│  ├─ debug/                调试内核：fgldb 驱动、会话、REST+WS API
│  ├─ dev/                  设计器包管线 → 见 §6.1
│  │  └─ tzs/               .tzs 引擎的 Go 客户端（manifest/wire/client/server/state/doctor）
│  ├─ dict/                 数据源与查询：db（本地）/ live（远程）/ dbsync / server
│  ├─ host/                 ★ 合并后唯一的 SSH/PTY/终端/探测/镜像层
│  ├─ dbconfig/ erpdb/      数据库连接模型与连接器（两处合并）
│  ├─ safesql/ sshtun/ output/ pathinstall/ atomic/ winproc/
│  └─ web/                  统一 HTTP 服务与共享 /api/*
│
├─ web/                     前端（React 18 + Vite 6 + TS + Tailwind 4）
│  ├─ app/                  唯一一套 SPA：调试工作台 + 统一设置页
│  └─ shared/               共享层：主题变量、UI 基元、设置页布局件
├─ skills/                  AI 技能：tt-debug / tt-dev-tzc / tt-dev-tzs / tt-dict / erp-read
└─ testdata/                FGL 夹具
```

`★` = 为合并或为 `.tzs` 而真正重组的部分。

### 1.3 术语

| 词 | 含义 |
|---|---|
| **包 / 工作区** | 包 = `.tzc`/`.tzs` 文件；工作区 = `tzc export` 出来的可编辑目录（`-ws` 后缀） |
| **围栏** | 写进 `prog.full.4gl` 的单行注释标记，圈出可编辑区间；**不是内容，是坐标** |
| **Region** | 一段被围栏圈出的字节区间，带权限与类型（见 §6.6） |
| **点 / 区段** | 设计器对 `.tap` 内容的两种组织单位；自订定义点 = `function.*`/`dialog.*`/`report.*` |
| **entry / 条目** | zip 里的一个文件（`.4gl`/`.tap`/`.tgl`/`ver`/未知） |
| **句柄（handle）** | `.tzs` 引擎里一个已加载包的会话标识，永不复用，进程内有效 |
| **守护进程** | `.tzs` 引擎的常驻进程；管道名含程序集 MVID，重编即换名 |

---

## 2. 统一配置

### 2.1 单一配置文件

合并后**只有一个** `config.json`，所有命令组共用（操作见 README）。顶层节：

```jsonc
{
  "schemaVersion": 2,
  "listen": "127.0.0.1:28670",   // 统一 Web 服务

  "hosts": {                      // ★ 共用环境清单：所有命令组读同一份
    "activeEnv": "开发环境",
    "sshs": [{
      "name": "开发环境", "host": "10.0.0.1", "port": 22,
      "user": "tiptop", "password": "…", "zone": "31", "topent": "10001",
      "launchArgs": "", "watchdogSeconds": 0,   // 可选：覆盖全局调试参数
      "db": {
        "type": "oracle",         // oracle | kingbase
        "host": "10.0.0.2", "port": 1521,
        "service": "t100dev",     // oracle: service；kingbase: database
        "readonlySql": true,      // 省略 = 启用只读守卫
        "viaSsh": { },            // 可选：经 SSH 端口转发再直连
        "accounts": [ { "account": "ds", "password": "…" } ]
      }
    }]
  },

  "debug": { "activeEnv": "", "launchArgs": "…", "watchdogSeconds": 1800,
             "fglserver": "", "termWidth": 200, "termHeight": 50,
             "printElements": 1000, "persistBreakpoints": true },
  "query":  { "source": "auto" },  // auto（缺省=在线优先）| local | <环境名>
  "mirror": { "dir": "" },
  "bdldoc": { "dir": "" },
  "sync":   { "target": "" },
  "tdev":   { "workspaceSuffix": "-ws", "defaultOut": "" },
  "tzs":    { "workspace": "" }    // 设计器程序集随仓库/包自带，不在这里配
}
```

要点：

- **`hosts` 是唯一的环境/连接登记处。** 原来 `tdebug` 把环境放在 `debug.sshs`、`tdict` 放在
  `hosts.sshs`，两份 schema 完全一致，合并后只留 `hosts.sshs`。这是合并里最关键的数据结构变更。
- **`debug` 节降级为纯工具设置**，不再承载环境列表。`debug.activeEnv` 可覆盖 `hosts.activeEnv`。
- **`tdev` 节**：原 TDev 完全没有配置系统，所有参数都是每次调用的 flag。新节只放跨调用稳定的
  默认值，**flag 仍然优先**；配置读取失败绝不让命令失败。
- **`tzs` 节只有 `workspace`**（外加可选的 `serverExe` 覆盖）。设计器程序集不在配置里 ——
  它由仓库与发行包自带（§7.1）。
- **账号无"主账号"**：客户端直连取 `accounts` 列表**首项**；未收录的账号按"账号=密码"惯例兜底。

### 2.2 配置位置

单一实现 `internal/config/paths.go`，按优先级（第一个存在的胜出）：

1. `TT_CONFIG` 环境变量（旧名 `TDEBUG_CONFIG` / `TDICT_CONFIG` 仍识别）
2. `--config <路径>`
3. `<exe 目录>\.portable` 存在 → 便携包，配置留在包内
4. `%APPDATA%\T100\tt\config.json` —— 默认统一位置
5. 旧位置兜底：`<exe 目录>`、`<cwd>`、`%APPDATA%\T100\tdebug`、`%APPDATA%\T100\tdict`、
   `%APPDATA%\TDebug`

`T100_HOME` 整体改写统一目录（如 `T100_HOME=D:\t100`）。数据目录 = 配置所在目录，所以
`ents/`、`srccache/`、`execlog/`、`debug-bps/`、`spill/`（可再生的缓存，`tt cache` 管）与 `.tt-serve.json`（运行中服务的状态文件）都跟着落在一起。

> **显式指定的路径就是答案。** `--config` / `TT_CONFIG` 指向一个还不存在的文件时，**不得**
> 悄悄改用默认落点 —— 那会让便携版把配置写进用户目录，也会让测试脚本落在别处。

### 2.3 旧配置迁移（`internal/config/migrate.go`）

首次运行且目标配置不存在（或 `schemaVersion < 2`）时执行一次，结果落盘并写 `schemaVersion: 2`：

1. 读入候选旧配置：`%APPDATA%\T100\tdebug\config.json`、`%APPDATA%\T100\tdict\config.json`、
   便携包/工作目录下的 `config.json`。
2. **拆解 TDebug 的 `debug` 节**：`debug.sshs` → `hosts.sshs`（提升）、`debug.listen` → 顶层
   `listen`（提升），**其余键留在 `debug`**。
3. 合并两份 `hosts.sshs`：按 `name` 去重，**同名以 TDictCli 那份为准**（它是字典查询的现役
   配置，保留它能让查询侧行为完全不变；TDebug 独有的环境仍会并进来）。
4. `hosts.activeEnv` 取 TDictCli 的值；缺失时取第一台环境名。
5. `listen` 取 TDebug 的 `debug.listen`，否则 `127.0.0.1:28670`。
6. **原文件不删除**，各留一份 `.pre-merge.bak`。
7. 预览：`tt config migrate --dry-run`。

> **最容易出错的一处**：TDictCli 原有的迁移规则是「顶层 `debug` 键**整节重命名**为 `hosts`」。
> 在合并后的配置里 `debug` 是**真节**（放 `launchArgs`/`watchdogSeconds`/`termWidth`…），
> 照搬那条规则会把调试设置一并吞进 `hosts` 并从 `debug` 删掉。所以换成了拆解规则，
> 并由 `config_test.go` 的 `TestMergeConfigs_BothTools` 盯着（断言 `debug` 节仍存在、
> `debug.sshs` 不残留、`debug.listen` 已提升）。

### 2.4 唯一写入口

所有写操作走 `internal/config/cfgfile.go` 的 `Edit(path, validate, mutate)`：
`Open → validate → mutate → Save`。`Save` = 同目录 `CreateTemp` + `Write` + `Sync` + 保留原
权限位 + `Rename`，失败不留半截文件，**也不会覆盖自己那一节之外的任何键**。

禁止任何调用方自己 `json.Unmarshal` 后 `os.WriteFile`。读取侧由 `schema.go` 的类型化 schema
收敛（原先散落在 6 个文件里的匿名 struct 反序列化）。

**未知键一律保留**：持久化以 map 为准，`Root` 只是读侧的类型化视图。这是「打开设置页点一下
保存，手工配的东西就没了」那类 bug 的根治办法 —— 不过 `viaSsh` 还需要服务端额外兜底，见 §4.3。

---

## 3. 环境、账号与输出（规范）

> 这一章是**规范**：给后续开发定"什么从哪来、谁管什么、返回长什么样"。
> 动 `internal/host` / `internal/dbconfig` / `internal/erpdb` / `internal/entdir` /
> `internal/output` / `internal/cli/dict` 之前**必读**。
> 别处只讲"怎么用"，这里讲"为什么只能这样"。

### 3.1 一条链：登录 → 环境变量 → 企业 → 账号 → 库

```
配置里的环境(hosts.sshs[i])
  │  name / host / port / user / password / zone / topent / db
  ├─① SSH 登录           密码认证,20s 超时,不校验主机密钥(内网工具,见 3.5)
  ├─② 加载 T100 环境变量   source /u3/pub/bin/chenv <zone>
  │                      → TOP / ERP / COM / FGLDIR / FGLRESOURCEPATH / TOPENT
  ├─③ 企业编号 TOPENT      ← 这条链上唯一的"业务身份"
  ├─④ 查 gzou_t           企业编号 → 账号名(gzou_t.gzou003)
  └─⑤ 库地址              **来自配置**(hosts.sshs[i].db),不从 TOPENT 推
```

**最要紧的一句：企业编号决定的是账号(schema)，不是库。** 库一对一挂在环境上
(`hosts.sshs[i].db`)，企业编号只决定"用这个库里的哪个账号连"。所以 `internal/entdir`
开头那句注释是对的：**"企业 → 数据库"这个说法不准**，解析出来的永远是**账号**。

`topent` 与 `db.accounts` 是**两个独立字段**，不互相推导：

| 字段 | 谁写 | 作用 |
|---|---|---|
| `hosts.sshs[i].topent` | 设置页「站点管理 → SSH」、`tt env topent` | 默认企业编号；登录后 T100 自己也会回显一个 `$TOPENT`，两者应当一致 |
| `hosts.sshs[i].db` | 设置页「站点管理 → 数据库」、`tt dict db discover --save` | 库地址 + 账号清单 |

### 3.2 登录之后拿到什么环境变量

**T100 没有 `$TOPDIR`。** 登录区域(`zone`：31 开发 / 35 测试 / 36 正式 / 39 PATCH /
t 出货)经站点 profile 的 case 表得到 `ZONE` 目录名，`topenv` 再派生全部路径。
所以 T100 路径**不允许静态配置** —— 登录后动态获取是唯一权威来源
(`internal/host/tenv.go` 的 `RuntimeEnv`)，取不到就报错，不猜、不回退到某个缺省。

回读的六个变量(`TEnvProbe`)：

| 变量 | 含义 | 谁用 |
|---|---|---|
| `TOP` | `/u1/<ZONE>` | 派生下面两个 |
| `ERP` / `COM` | `$TOP/erp`、`$TOP/com` | 源码路径、模块根 |
| `FGLDIR` | Genero 运行时目录 | 启动 `fglrun` |
| `FGLRESOURCEPATH` | `$ERP:$COM` | 同上 |
| **`TOPENT`** | **当前企业编号** | **喂给 3.3 的账号解析** |

⚠️ **探针必须用定界符包起来**(`TDBG-BEGIN` / `TDBG-END`)。这不是洁癖：登录脚本
**自己也会打印** `ZONE = t35prd`、`TOPENT   = 99` 这类行。PTY 只有纯文本，没有区间就
分不清"我们请求的回显"和"服务器自己说的话"。改探针时这两行不能删。

⚠️ **zone 是拼进远端 shell 的字符串，必须过白名单**(`reZone`：
`^[A-Za-z0-9_-]{1,16}$`)。历史教训：从前它是裸插的，实测
`tdebug start --zone "36; id"` 能在服务器上执行任意命令。

### 3.3 环境变量与数据库账号的关系

账号**不是配出来的，是解析出来的**：

```
TOPENT(企业编号) --查 gzou_t--> gzou003(账号名) --> 拿这个账号去连库
```

- 表：`gzou_t`（`gzou001` = 企业编号，`gzou003` = 账号名，`gzoustus='Y'` 才算启用）
- 结果缓存到 `<配置目录>/ents/<环境段>.json`，新鲜期 10 分钟
- **两条访问路径共用同一个快照文件**（`internal/entdir` 是这个共享件的唯一实现）
- 密码：在 `db.accounts` 里按账号名查（`PasswordFor`）；未收录则回退 T100 惯例
  **账号 = 密码**

**两条路径必须给出同一个账号**（2026-09 定下的规矩，此前 `tt dict` 固定取
`accounts[0]`、`tt debug sql` 才查 `gzou_t`，于是同一张字典表可能被两个 schema 读到而
结果里没有任何信号）：

| 路径 | 谁在连 | 传输 | 用哪次连接读 gzou_t |
|---|---|---|---|
| **服务器侧** | T100 服务器 | SSH exec + `sqlplus`/`ksql`，SQL 走 **stdin** | 借服务器上的 sqlplus |
| **客户端直连** | 本机 | `erpdb`（pgx / go-ora），可选 SSH 隧道 | 先用 `accounts[0]` 连一次读 |

客户端直连路径的降级 —— **每一步都记进输出信封，绝不静默**：

| 情形 | 行为 | 信封里怎么体现 |
|---|---|---|
| `topent` 没配 | 用 `accounts[0]` | `accountSource: "accounts[0]"` |
| `topent` 不是数字（如把据点码 `DSCNJ` 填在那儿） | 同上 + note | note 写明原因 |
| 企业不在 `gzou_t` 启用清单里 | 同上 + note | 同上 |
| 解析出的账号不在账号清单里 | **仍用它**，密码走惯例 + note | `accountSource: "ent(gzou_t)"` + note |

**为什么允许降级而不是报错**：查一条字典不该因为企业目录查不动就整个失败
（企业目录要现查，冷启动时多一次连接，见下）。但降级**必须可见** —— 这是
`accountSource` 与 `notes` 存在的全部理由。

代价也说清楚：**快照未命中时要多连一次**（先以 `accounts[0]` 读 `gzou_t`，再以解析出的
账号重连）；命中快照则零额外开销。

### 3.4 两条路径的边界（别把对方的能力搬过去）

- **服务器侧**（`internal/debug`）只服务 `tt debug sql` / `db` / `ents`：它要在**那台
  服务器**上跑 SQL。`db.viaSsh` 对它**无意义**（它本来就连着服务器），静默忽略是对的。
- **客户端直连**（`internal/dict/live` + `internal/erpdb`）服务所有字典查询与
  `tt dict db ping/sync`。`viaSsh` 只在**客户端到不了库**时才配：先起 SSH 端口转发，
  驱动改连 `127.0.0.1:<本地端口>`。
- **不要**为了"统一"把字典查询改走 SSH：那会把一次直连查询变成"登录 + 探测工具路径 +
  拼连接串"三次往返，还把库的可见性绑死在服务器上。
- 反过来，**不要**给服务器侧加 `viaSsh` 支持 —— 它是给客户端用的。

### 3.5 谁管什么（责任表）

| 东西 | 谁管 | 存在哪 | 怎么改 |
|---|---|---|---|
| **SSH 连接**（host/port/user/password/zone/TOPENT） | 环境（`hosts.sshs[]`） | `config.json` | 设置页「站点管理 → SSH」、`tt env list/use/topent` |
| **数据库连接**（type/host/port/service\|database/accounts） | 环境（`hosts.sshs[].db`） | `config.json` | 设置页「站点管理 → 数据库」、`tt dict db discover --save` |
| **用哪个账号** | **解析出来的，不是配的** | 快照 `ents/` | 由 `topent` + `gzou_t` 决定（3.3） |
| **查询走哪个数据源** | `query.source` / `--env` | `config.json` | 设置页「数据字典 → 查询数据源」、`--env local\|<环境名>` |
| **本地字典副本** | `sync.target` | `config.json` | 设置页「数据字典 → 数据同步」、`tt dict db sync` |
| **返回条数上限** | `query.limit` | `config.json` | 设置页、`--limit`（单次覆盖） |
| **缓存**（企业快照/源码镜像/执行日志/断点/查询落盘） | 可再生，**不是数据** | `<配置目录>/<子目录>` | `tt cache`、设置页「应用设置 → 缓存」 |
| **配置写入** | 只有 `config.Edit` 一条路 | — | 新增写路径就是 bug（见 §2.4） |

几条硬规矩：

1. **命令组之间互不 import。** `internal/cli/debug` 与 `internal/cli/dict` 不许互相引用；
   共享能力一律下沉到 `internal/host` / `config` / `dbconfig` / `entdir` / `output`。
   （合并前同一件事各写一份的后果，见 §2.1 与 §11.2。）
2. **SSH 主机密钥当前不校验**（`InsecureIgnoreHostKey`）—— 与"手工 `ssh` 登录内网机器"
   同一个信任级别。要改这条等于改安全模型，**须单独讨论，不要顺手改**。
3. **口令一律不进命令行**（客户端直连路径）。服务器侧金仓是唯一例外：
   `KINGBASE_PASSWORD` 仍会出现在远端 argv 里（`ps` 可见），`host.KbCmd` 的注释如实写了。
4. **往配置目录写文件时，要么登记进 `internal/config/cache.go` 的 `CacheSubdirs`，
   要么在注释里说清它为什么不是缓存。** 那个目录里住着 `config.json` —— 唯一删了不能
   自愈的东西。

### 3.6 查询输出的契约

所有 DB 查询命令（`tt dict` 的 r.t/r.v/scc/desc/r.q/msg/sysp/docp/prog、`tt debug sql`）
共用一个出口 `internal/output`；`Meta` 是"我这次落在哪"的唯一表述：

```jsonc
{
  "ok": true,
  "source": "live", "env": "主机正式区", "sshHost": "…", "zone": "36",
  "topent": "99", "ent": 99, "account": "dsdemo", "accountSource": "ent(gzou_t)",
  "target": "…:1521/t35prd", "dialect": "oracle", "route": "client-direct",
  "readonly": true,
  "totalRows": 28005, "returned": 20, "truncated": true,
  "localPath": "…/spill/query-….json", "elapsedSeconds": 1.2, "notes": [],
  "data": [ … ]
}
```

**默认 JSON**；`--format csv` 给 `# ` 头 + 裸 CSV；`--format table` 给人看。
四条不许破：

1. **空段不打印。** 本地源拿不到 SSH，就不出现 `环境  · SSH  · 区域 ` 这种空壳行。
2. **企业编号与账号同在一行** —— 它们是一条链，拆开会被读成两件独立的事。
3. **截断绝不静默，结果绝不丢**：先落盘完整结果，再截；**落盘失败就不截**
   （宁可给一坨大的，也不给一份看不出少了东西的假结果）。四处同时说：信封字段、
   `notes`、CSV 头三行、`# 落盘` 行。
4. **stdout 只放数据**（JSON 模式）：诊断进 `notes` 或 stderr。CSV 的 `# ` 头留在 stdout
   是刻意例外 —— 头必须随数据走，管道里没法把 stderr 和 stdout 重新配对。

错误同样是信封：`{"ok":false,"code","error","hint","exitCode"}`，`code` 是**稳定契约**
（`USAGE` / `ENUM_INVALID` / `ENV_UNKNOWN` / `TABLE_MISSING` / `CONNECT_FAILED` /
`QUERY_FAILED`），退出码 `0 成功`（含"查无结果"）`/ 1 用法错 / 2 数据源错 / 3 缺表`。
逐条规矩与依据见 §8.5。

### 3.7 加新东西时的检查清单

- [ ] **环境从哪来？** 一律 `config.Hosts.Resolve` / `Root.ResolveForTool`；
      **不要自己写 "name → activeEnv → 首条"**（历史上抄了五份，错误文案与边界行为
      全不一样，见 §11.2）
- [ ] **账号从哪来？** 一律走 `entdir` + `gzou_t`；**不要自己取 `accounts[0]`**
- [ ] **输出怎么走？** 一律 `output.Emit`；**不要自建 JSON/CSV 发射器**（曾经有四套）
- [ ] **写配置？** 一律 `config.Edit`
- [ ] **往配置目录写文件？** 登记进 `CacheSubdirs`，或说清为什么不是缓存
- [ ] **拼接服务器命令？** 值必须过白名单/引号（`reZone` / `reToolPath` / `reDBAcct` /
      `shQuote`）；**SQL 一律走 stdin**，不拼进命令串
- [ ] **命令行新增位置参数？** 负整数之类的值会被 flag 解析吃掉，给出可操作的提示
      （见 `tt dict msg` 的 `SetFlagErrorFunc`）

---

## 4. 统一 Web 服务

### 4.1 一个进程、一个端口、一套页面

`tt serve` 启动，默认 `127.0.0.1:28670`（占用时自动向后探测），**默认后台常驻**（单实例），
`--foreground` 前台、`--stop` 停止。

| 路径 | 内容 |
|---|---|
| `/debug/` | 调试工作台 SPA（源码/断点/调用栈/变量/接口日志） |
| `/debug/#settings/*` | 统一设置页：站点管理 / 数据字典 / DEBUG / 应用设置 |
| `/debug/api/*` | 调试子系统 API（内部仍按 `/api/…` 注册，靠 `StripPrefix` 挂在前缀下） |
| `/api/*` | 统一接口：`/api/hosts` 配置分节读写、`/api/config/status` 派生状态、`/api/install` PATH，以及字典类长跑动作（镜像拉取 / 字典同步 / BDL 文档） |
| ~~`/dict/`~~ | 字典页 SPA **已删除**，它的两个动作并入设置页「数据字典」分区 |

**为什么必须按前缀分 API**：合并前两个工具各自起服务，两边都注册了 `/api/status`、
`/api/dbprobe`、`/api/dbaccverify`、`/api/conntest` —— 放进同一个 mux 会直接冲突。

**为什么子系统只挂 API、页面归统一服务**：一开始把整个子系统 handler 用 `StripPrefix` 挂在
`/debug/` 下（子系统自带 `/` 兜底）。这条路走不通 —— Go 的 `ServeMux` 一旦匹配到 `/debug/`
前缀就**不再向外回落**，子系统的 `/` 兜底会把该前缀下所有请求吃掉，页面反而出不来。

配套两个坑，都注释在 `internal/web/spa.go`：

- **挂载点与目录不能交给 `http.FileServer`**：它对目录返回一次到 `/` 的 301，前缀被丢掉，
  用户会从 `/debug/` 被弹到根路径。目录一律直接给 `index.html`。
- **未命中的路径要回落 `index.html`**，否则刷新页面或粘深链接都是 404。

`//go:embed all:web/dist` 的 FS 根是 `web/dist`，而所有消费者按"dist 根"理解路径，所以
`main.go` 用 `fs.Sub` 剥掉一层；**"前端放在哪"只在 `common.WebSub` 里说一次**。

### 4.2 服务的单实例与自寻址

运行状态写在与 `tt debug serve` **同一份** `.tt-serve.json` 里，其中 `apiBase` 记着调试 API
挂在哪 —— `tt serve` 是 `/debug/api`，独立的 `tt debug serve` 是 `/api`。控制端命令
（`tt debug start/exec/status/…`）据此寻址，所以对着哪个实例都能用，不必知道对面是哪种装法。

> 同一份配置目录下只允许一个实例：两条进程会争同一份状态文件，`--stop` 与自动寻址就都指错人了。
>
> 这条是修出来的：前缀不同而控制端命令永远拼 `/api/…`，于是 `tt debug start/exec/status`
> 对着 `tt serve` 全是 404。现在前缀记进状态文件，按文件寻址。

### 4.3 共用的配置端点

`GET/PUT /api/hosts`：读回 `hosts`/`listen`/`debug`/`query`/`mirror`/`bdldoc`/`sync`/`tdev`；
PUT 只接受要改的节（缺省 = 不动），交给 `internal/config` 的原子写路径。两个页面都写它。

服务端替前端兜底三件事：

- **校验归服务端**：拒绝空环境列表、重名、端口越界、库类型非法、`activeEnv` 指向不存在的
  环境；400 里给中文说明，前端原样展示。
- **保存时保留表单不管理的 `db.viaSsh`**（SSH 端口转发隧道）。配置页没有它的输入控件，
  整节替换会把它抹掉 —— 那是"打开设置页点一下保存，手工配的隧道就没了"，而用户完全看不出
  发生过什么。按环境名对齐补齐；代价是通过页面删不掉 `viaSsh`，想删就手改 `config.json`。
  **宁可选"删不掉"。**
- **写入成功后调用子系统的 `ReloadConfig()`**（`internal/web.ConfigReloader`），让持有配置
  内存态的调试服务重新加载。否则从别处改了环境，调试侧仍按旧环境连，表现是"改了没生效"。
  重新加载失败只记日志：配置已经落盘成功，不该把成功的写报成失败，但必须让用户看见。

### 4.4 设置页的四块

| 分区 | 内容 |
|---|---|
| 站点管理 | 环境列表 + 「SSH 服务器 / 数据库」两个 Tab；从服务器只读探测回填连接要素；账号行 ⚡ 验证、数据库「测试连接」 |
| 数据字典 | 查询数据源（在线 ⇄ 本地）、镜像根目录与增量/全量拉取、同步目标与字典同步、BDL 文档目录 |
| DEBUG | 调试参数（`launchArgs`/看门狗/终端宽高/print 上限/断点持久化） |
| 应用设置 | 命令行集成（加/移除用户 PATH）、明暗色、运行信息 |

前端**只有一套页面**：合并完成后曾有两套 SPA 并存互链，后来字典页被删掉、两个动作并进设置页
（经过见 §11.6）。

---

## 5. `tt debug` 调试器

通过 SSH 在 T100 服务器上驱动 `fglrun -d` 的 `(fgldb)` 文本调试协议，提供本地 Web 调试界面与
命令行控制端，实现"人操作 GDC 界面 + AI 借助命令行检查分析"的人机协同调试。

### 5.1 能力

- **Web 界面**（`tt serve` 后浏览器打开 `/debug/`）：源码 + 断点 + 调用栈 + 变量监视
  （Monaco、悬停求值、大纲、运行到光标），接口报文日志与重放调试，服务测试，环境/数据库/参数设置。
  主题支持**亮色 / 暗色 / 跟随系统**；还没有配置任何环境时，打开界面直接落到「设置 → 环境」，
  不会把用户丢在空调试页。
- **命令行控制端**：自动发现后台服务地址；`exec` 透传全部 fgldb 标准命令并返回原生文本。
- **人机协同**：程序跑到 INPUT/MENU 等交互语句时会阻塞在 GDC 等人操作（从调试器看与死循环
  无法区分）—— `tt debug why` 专门回答"它在等人还是在空转"，判断信号与交接话术见
  `skills/tt-debug/SKILL.md`。
- **防呆**：停站停留超时看门狗自动放行（默认 1800s，保护生产区行锁）、断点持久化、单实例服务。

### 5.2 命令一览

| 命令 | 作用 |
|---|---|
| `serve` | 启动本地调试服务（默认后台常驻单实例；`--listen`/`--foreground`/`--stop`） |
| `status` | 查看服务状态与活动会话 |
| `start <作业>` | 连接 SSH、启动调试并等到入口停站（`-m/--module`、`--zone`、`--ssh`、`--timeout`） |
| `exec "<命令>" [更多…]` | 透传标准调试命令（print/break/next/where/info/watch…），原样返回输出；**可一次给多条**（或 `--file`）省掉每条一次进程启动，逐条标注 `[i/N]`；resume 类命令执行后读一次现场，停住了就继续往下跑，没停住才中止并把剩余标为未执行；resume 类用 `--wait N` 软等待（到点返回，不发 SIGINT），`--timeout` 才是会发 SIGINT 的硬超时 |
| `quit` / `stop` | 结束会话 / 查看当前停站现场（可反复取用，带 `waitingForUser`/`waitingKind`） |
| `why` | 探测程序此刻在**等用户操作**还是**空转**（interrupt → where → 判定停站行是否交互语句；默认探完自动放回，`--no-resume` 保留现场） |
| `wait` | 等会话事件（`--for stopped,exit,dead,watchdog`、`--timeout`）：事件一到即返回，超时返回 `timedOut`（不是错误） |
| `source` | 读服务器源码（登录区源码目录白名单只读，`--from/--to` 取行段，`--path-only` 只要行数与本地副本路径）；每次读取落一份本地副本供整读 |
| `logs` / `locate <函数>` / `resolve <作业>` / `interrupt` | 会话事件 / 定位函数到 文件:行 / 解析作业编号 → 实体程序（不建会话）/ 中断运行中或卡住的程序 |
| `sql "<语句>"` | 执行一条**只读** SQL（单条 SELECT/WITH；账号由 TOPENT 经 `gzou_t` 解析）。**默认输出 CSV**：头部若干行 `# ` 注释回显**环境信息**（环境名/SSH/区域、**本次用的企业(ENT)→账号**、库与耗时、行数），其后是纯净 CSV 主体（表头 + 数据），`grep -v '^# '` 即得标准 CSV；`--json` 才回结构化详情。白名单 + 库侧只读事务双重约束，最多 200 行；`hosts.sshs[].db.readonlySql=false` 可关闭 |
| `wslogs` | 接口报文日志列表（`--job` 支持通配、`--service`/`--server`/`--origin`/`--result`/`--pid`/`--from`/`--to`/`--fail`/`--page`/`--show <rowid>`；条件口径对齐 T100 原生 awsq990） |
| `wsdebug <rowid>` | 按日志报文参数重放调试，停在入口；`--set 路径=值` 改入参再重放（可重复）、`--request-file` 整份替换报文 |
| `db [--ent N]` | 数据库连接探查：企业(TOPENT) → 账号映射与连接验证（`--refresh` 跳过缓存强制现查） |
| `ents [--ent N]` | 企业目录：当前环境有哪些企业编号、各用哪个账号（`--refresh` 现查 / `--cached` 只用快照 / `--env` 指定环境）。**带落盘快照，断网也能答**，用 `stale`/`fetchedAt` 标注新鲜度 |
| `probe` | 协议驱动器自检尖刺：登录→启动→下断点→步进→求值 |

全局参数 `--config` / `--json` / `-v`；控制端命令另有 `--url` 覆盖自动发现的地址。
控制端命令默认输出**给人看的文本**；唯一例外是 `sql` —— 它的默认输出是 **CSV**
（头部 `# ` 注释 + 纯净主体，理由见上表），`--json` 在那条命令上是"显式要结构化详情"。

**「当前环境」有两套语义，都保留**：`tt env use <名称>` 改的是**配置默认**
（`hosts.activeEnv`）；`tt debug env <名称>` 切的是**活动调试会话**（会重连）。
`tt debug topent <值>` 是会话级覆盖，与 `tt env topent <名称> <编号>` 设的配置默认不同。

### 5.3 环境变量

| 变量 | 作用 |
|---|---|
| `TT_CONFIG` | 覆盖配置文件路径（优先级高于 `--config`）；旧名 `TDEBUG_CONFIG`/`TDICT_CONFIG` 仍识别 |
| `T100_HOME` | 改写统一用户目录的父目录（缺省 `%APPDATA%\T100`） |
| `TT_PROXY` | 前端开发模式的后端地址（端口顺延时用；原 `TDEBUG_PROXY`） |
| `TDEBUG_SERVE_LOG` | 后台服务子进程写入的日志路径（由 `serve` 自动设置） |
| `TDBG_RAW=1` | 把 fgldb 协议原始行打到服务日志（排障用） |

### 5.4 运行时产物（均在 `.gitignore` 中）

- `.tt-serve.json` —— 后台实例状态（pid/地址/**调试 API 前缀**/日志）；控制端据此自动寻址
- `.tt-serve.log` —— 后台服务日志
- `debug-bps/<模块>__<作业>.json` —— 断点持久化
- `ents/<环境>.json` —— 企业目录快照。**是可丢弃的缓存**：换了机器/区域/库（指纹不符）或超过
  10 分钟就当没有，重新现查；文件坏了也不影响命令失败

**控制端命令**：`start/exec/status/quit/stop/source/logs/locate/resolve/interrupt/env/topent/
wslogs/wsdebug`。

---

## 6. `tt dev tzc` 代码包

把「改 T100 设计器里的 4GL 客制」变成一条**可机械证明不会破坏**的管线。完整命令与参数见
README 与 `tt dev tzc --help`；本章讲契约与依据。

### 6.1 三条公理

| 公理 | 含义 |
|---|---|
| **A1 唯一编辑界面** | AI 只面对一个带围栏标注的 `.4gl` 文件，和人在设计器里看到的画面一致 |
| **A2 唯一写路径** | 任何修改都走 `apply`；不存在「单点手术」式命令 |
| **A3 不破坏是机械证明** | 围栏外逐字节比对 → 不变量 + FGL 解析 → 新包的装载模拟 |

生命周期四个动词 `export/status/verify/apply` 与三条公理一一对应，没有第五个。
`unlock` 是**单向状态迁移**（只改 workspace）；`rename`/`newfn` 是**编辑辅助**（把设计器弹窗
CLI 化，只改 workspace）。后三类产出的都是普通的工作区改动，**仍须走 `apply` 才落盘**。

`<dir>` 可省略：先 `cd` 进工作区即可。判定规则只有三条，没有隐藏行为 —— 显式给了就用给的；
没给且当前目录确实是工作区（含 `prog.full.4gl` + `manifest.json`）就用当前目录；否则
**明确报错退出 5**，绝不猜别处的目录。

包相关的实现全在 `internal/dev/`：

```
pkgfile/ tapfile/ tglfile/   zip / TAP / TGL 三层（字节保真的读写）
fgl/     synth/    fence/    FGL 解析 / 合成 / 围栏渲染
verify/  split/    store/    三道闸门 / 写回拆分 / 工作区
model/   testutil/ cli/      域类型 / 夹具 / 命令行派发
```

### 6.2 为什么不是直接改 zip

- `.tzc` 是普通 zip，但里面的 `<prog>.tap` 才是**唯一被服务端消费**的设计文件；`.tgl` 是框架
  骨架、`.4gl` 是服务器 build 产物（**客户端从不读回**）。
- 设计器保存时会把整包**非原子重写**（`File.Delete` → `File.Create`，时间戳/压缩级别/ACL 全变），
  并且 `.tap` 的换行是**混合**的（元素间 CRLF、CDATA 内 LF）。所以任何「解析成 XML 再序列化」
  的做法都会让 diff 爆炸、甚至丢数据。
- 对策：**CDATA 感知的字节级扫描器**，只重建被改的 CDATA 内部；其余字节（属性顺序、引号风格、
  空白、未知条目、`ver`）逐字节透传。

### 6.3 包容器形态：照抄设计器，不是"写个合法 zip"

**合法不够，得像它。** 设计器写出来的 `.tzc` 有自己的 zip 形态，写回时必须一模一样：

| 特征 | 设计器（.NET）的包 | 必须避免的写法 |
|---|---|---|
| general purpose flag | `0x0000`（**没有**数据描述符） | `0x0008`：bit 3 + 数据描述符 |
| CRC / 压缩后大小 / 原大小 | **写在局部头里** | 局部头填 0，真值只放数据描述符 |
| 局部头 extra | 原样保留（`UT-time` 9 B + `ux-infozip` 11 B） | 被替换成写入器自己的时间戳字段 |
| 中央目录 extra | 原样保留（`UT-time` 5 B + `ux-infozip` 11 B） | 同上 |

> **真机事故（S7 抓到的第一个不兼容）**：早期实现用 Go 标准库 `archive/zip` 的 `Writer` 重写
> 整包 —— 它**无条件**置 bit 3 并补数据描述符，还会用自己的 9 字节 extra 顶掉原 extra。结果：
> 包在 tt dev 里读得回来（Go 看中央目录），拿到设计器里一打开就报
> **`Data descriptor signature not found`**。这正是「S7 真机验收不能省」的原因。

现在的写回是**字节级重建**（`internal/pkgfile/zipraw.go`）：局部头与中央目录记录都从原包逐字节
拷贝，只打补丁 —— 清 bit 3、写真实 CRC/大小、更新局部头偏移、必要时更新 Zip64 extra 里的
8 字节字段；未改动的条目连**压缩数据都逐字节照抄**。于是：

- **零改动重建 = 与原包逐字节相同**（166 个真实包：165 个逐字节相同，1 个是早期产出的 bit3 包，
  被归一化成设计器形态；归一化之后是不动点）；
- **只改 `.tap` 时，差异只出现在该条目的 CRC/尺寸字段与其载荷、以及它之后条目的局部头偏移**；
- **Zip64**：原包用 `0xFFFFFFFF` 标记的字段保持标记，真值就地写回 8 字节字段（语料里有 302 个
  这样的条目）；不静默去掉 Zip64，也不在超 4 GiB 时乱写。

**改动的条目会被重新压缩**（Go 标准库 deflate），所以它的字节与设计器的 SharpZipLib（level 3）
不同；其余条目连压缩数据都逐字节照抄。判据始终是**逐条目内容 sha256**。

### 6.4 工作区

```
<dir>/
├── prog.full.4gl        # 唯一编辑文件：完整文档 + 围栏标注
├── manifest.json        # 人机共读索引（每个 Region 的 editable 与**不可编辑原因**）
├── snapshot/
│   ├── index.json       # 原包条目清单（name/sha256/size/role）
│   └── entries/<条目名>  # 原包每个条目的逐字节拷贝（恢复源；apply 不从这里读）
├── .tdev/
│   ├── base.full.4gl    # 导出时 prog.full.4gl 的字节拷贝（gate1 的比对基线）
│   ├── base.sha256
│   ├── regions.json     # Region 表：名字/类型/权限/原因/meta/字节区间/tgl_tag
│   └── lock             # 防止两个 tt dev 进程同时操作
└── .git/                # export 时 init + 首次 commit；apply 成功 = 新 commit
```

`manifest.json` 里每个 Region 都带 `reason`（中文），例如 `"SEC 模式未开启"`、
`"不在本次 --only 导出范围内"`、`"框架集合锚点区段强制只读"`。**AI 看到 `editable:false` 的
同时就知道为什么**，不必用试错法探测权限边界。

### 6.5 两条管线别用错

| | `tzc export`（代码包） | `tzs export`（表单包） |
|---|---|---|
| 产物 | 工作区：`prog.full.4gl` + `manifest.json` + `snapshot/` + `.tdev/` + `.git/` | 就是一包文件（`.tsd`/`.4fd`/`ver`…） |
| 特殊处理 | 合成 + 围栏渲染 + 权限判定 | **没有**：不解围栏、不校验必需条目、不解析 `ver` |
| 能改回来吗 | 改 `prog.full.4gl` → `verify` → `apply` | **不走 export**：要用 `tzs call` 让设计器自己算（§7） |
| 用途 | 改 4GL 客制 | 只读参考：读表单结构、查字段定义 |

拿错入口会被挡住：`.tzc` 跑 `tzs export` → 退出码 2 并提示改用 `tzc export`。

`tzs export` 的配套细节：默认目录 `<包目录>/<程序名>-unzip`（与 `-ws` 区分开，一眼能看出是
解压产物还是工作区）；目标非空则拒绝（退出码 5），`--force` 只覆盖同名文件；**zip-slip 防护**
（条目名带 `../`、盘符、绝对路径、NUL → 整体拒绝，退出码 2）。

### 6.6 围栏协议

围栏是**单行注释**，不改变 4GL 语义，且与 TGL 原生的 `{<point/>}`/`{<section>}` 标记区分开：

```
{//@tdev:begin section adzi999.main [READONLY deny="sec-off" src="s"]}
MAIN
  {//@tdev:begin point main.define_customerization [EDITABLE src="s" new="Y" order=""]}
  CALL ccl_init()
  {//@tdev:end point}
END MAIN
{//@tdev:end section}
```

规则：

1. `begin`/`end` 严格配对；允许 `section → point` **一层嵌套**（真实 TGL 里自订点就住在区段内），
   深度 > 2 报错。
2. **围栏行本身属于「围栏外」**：不许增删改围栏行（meta 由 tt dev 推导，不由作者填写）。
3. 自订定义点内部再分**结构行**（注释头 / `PUBLIC FUNCTION x(...)` / `END FUNCTION`）与
   **正文本体**；只有正文本体可改，结构行改动一律拦下（对齐设计器 `CanInsert` 的行为）。
4. 新增自订点的唯一方式：在 `[APPEND]` 锚点区段内追加一个完整的 point 围栏块。
5. 删除自订点 = 删除其整个围栏块，且**仅 `new="Y"` 可删**。
6. Region 之外的字节（文件头尾、区段之间的空行）不许增删改，逐字节比对。

**v2 围栏**多带了身份字段：

```
{//@tdev:begin point function.aapp131_qbe_clear [EDITABLE src="s" new="Y" order="1"
   fn="aapp131_qbe_clear()" scope="PRIVATE" desc="# 注释描述\n# 第二行"]}
```

- 旗标：`EDITABLE` / `READONLY` / `EDITABLE-SEC`（已解锁且可编辑的区段）/ `APPEND`（锚点可追加）/
  `PLAIN`（裸名插入点）。
- **锚定部分不许改**（point/section 名、旗标、`status`/`src`/`new`/`order`…）；
  `fn`/`scope`/`desc` 是**唯一允许改**的字段，改了就触发 V1–V7。
- `PLAIN` 点（裸名插入点）**没有函数身份**：整块即正文，除只读判定外不做结构校验 ——
  特殊的是自订定义点，不是所有 point。

### 6.7 三道闸门与 apply 管线

| 闸门 | 检查 | 失败 |
|---|---|---|
| **gate1** | 围栏行、围栏外字节、只读 Region 内容、可编辑区内的结构行 —— **逐字节相等**；删除/追加的授权 | 退出码 3；权限类 → 4 |
| **gate2** | 不变量 I1–I15（§6.14）：区段配对 / 命名空间三层错配 / `TglTag` 折叠 / status 取值 / UTF-8 无 BOM / `]]>` / `ver` / 未知条目透传… | 退出码 3（warn 仅在 `--strict` 下失败） |
| **gate3** | 对**将要产出的新包**重跑合成 = 「设计器打得开吗」的判决；**磁盘此刻仍未动** | 退出码 3 |

管线顺序即契约：

```
读 edited → parse_fenced → gate1 → 源包未变校验 → gate2 → split → 原子写 → git commit
                                                              ↑
                                            gate3（写之前，对内存里的新包）
```

> **任何一步失败，原包字节不变** —— 这不是目标，是 `atomic_write` + 管线顺序的必然结果。

**apply 到底会改哪些东西**（明确清单）：

| 对象 | 动作 |
|---|---|
| `.tzc` 包 | **原地覆盖**（原子写：临时文件 + rename） |
| └ 未改动条目（`.4gl`/`.tgl`/`ver`/未知） | **逐字节透传**，一个字节都不变 |
| └ `.tap` | 只替换被改区域的 CDATA 区间 + 相应属性 |
| `.tdev/prev.tzc` | 覆盖前把**上一版包**整文件备份一份（可回滚） |
| 工作区基线（`base.full.4gl`/`regions.json`/`manifest.json`/`snapshot/`） | 刷新为新状态（基线前进一格） |
| 工作区 `.git` | 一次 commit（`tdev apply: <prog>`） |

> **迭代循环**：`改 → apply → 再改 → 再 apply` 连续多次是支持的。这依赖 apply 成功后刷新
> `manifest.pkg.sha256`；不刷新的话第二次 apply 会被「源包自 export 之后已被改动」误拒（退出码 5）。
> 该缺陷已修复，并由自检项「真机事故-迭代循环」钉住。

### 6.8 报错怎么定位（行号 + 该行内容）

`apply` / `verify` 拒绝时，每条发现都带**出错位置** —— 出错文件、1-based 行号、该行内容：

```
apply 被拒绝（gate1 error=1 warn=0 info=1）：
  [error] gate1.readonly-region  只读 Region 内容被改动（写入被拒）
            位置：prog.full.4gl:5
            该行：#應用 a00 樣板自動產生(Version:3)  # 我加的字
            deny=section-locked
```

- 行号给的是**你编辑的那份** `prog.full.4gl`；删除类问题（「删掉了不该删的点」）给
  `.tdev/base.full.4gl:<行号>`（编辑后的文档里它已经不存在了）。
- 定位用**行尾等价**比较：编辑器把 CRLF 归一成 LF 不会把位置指到区段开头，而是指到你真正改的
  那一行。
- `--json` 里每条发现有 `file`/`line`/`snippet`；`status --json` 有带行号的
  `changed_regions`/`added_regions`/`deleted_regions`。
- TAP 层的字节偏移会补上落在哪个 `<point>`/`<section>`。

### 6.9 框架解锁（Unlocked）—— 取代 v1 的 `--allow-sec`

「解开框架」不是渲染效果，是**记录在 TAP 根属性上的单向状态**（`section_flag="Y"`）。解开之后，
**规格（SPEC）的任何调整都不会再产生对应的程序代码** —— 代码与规格从此脱钩。所以它是一次显式、
有代价、不可逆的状态迁移，不是开关：

```powershell
tt dev tzc export "pkg.tzc" -o ws     # Locked：SectionRegion 全只读
tt dev tzc unlock ws                  # 需要二次确认 → 打印设计器代价警告原文，退出码 4
tt dev tzc unlock ws --yes            # 确认后迁移（只改 workspace）
# 编辑区段正文 → apply（此时才把 TAP 根 section_flag 置 Y）
```

授权闸门逐条重放设计器 `checkBox_PreviewMouseLeftButtonDown`：

| 情形 | 判定 |
|---|---|
| 包本来就是解开态（`section_flag="Y"`） | Allow（设计器对这类包不设拦截，`unlock` 为 no-op） |
| `env="s"` 且 `login_user != "topstd"`，有 `std_section_verify="Y"` | Allow，但需 `--yes` |
| `env="s"` 且非 topstd，无 `std_section_verify` | **Deny，退出码 4**，打印 `adzi052` 授权原文 |
| 其他（`env="c"` 或 topstd） | Allow，但需 `--yes`（打印代价警告） |

> **R9：没有绕过。** 不提供 `--force`，不改 env/login_user，也不写 `std_section_verify` 假装有
> 授权 —— 那是服务器端 `adzi052` 管理的**权限域**，本地工具自我授权等于伪造权限。
> **AI 不得自主加 `--yes`**：框架是否解开必须由人决定。

解锁 ≠ 所有区段可编辑（**解锁改变的是门，不是所有房间**）：三个锚点区段
（`other_function`/`other_dialog`/`other_report`）**永远只读**，TGL 标 `readonly="Y"` 的仍只读，
topstd 模式下的 src 规则仍然生效。围栏呈现为 `[EDITABLE-SEC]`。

- `unlock` **不碰 `.tzc`**：只改 workspace（围栏旗标 + `.tdev/section-state`）+ git commit。
- Locked 态改区段 → 退出码 4，文案指向 `unlock`。
- `apply` 触发区段写入时提示「下次上传会触发服务器 `adzi520` 联动」（tt dev 不代为执行）。
- **反向迁移（重新锁上）不做**（R10）：那是服务器 `adzp064`（回标准）的职责。

### 6.10 结构事务（改名 / 新函数）

设计器把自订定义点的「身份」拆在**六处**（点名、签名行、scope、描述块、墓碑、调用点），
任何一处单独漂移都是故障。所以不禁止改，而是把「识别意图 → 证明一致 → 复刻事务」做成一等公民：

```powershell
tt dev tzc rename ws aapp131_calc aapp131_calc_v2                        # 原子同步围栏 fn + 签名行
tt dev tzc rename ws aapp131_calc aapp131_calc_v2 --desc "#+ 新描述"     # 同步描述块
tt dev tzc rename ws aapp131_calc aapp131_calc_v2 --scope PRIVATE        # 同步 scope
tt dev tzc newfn  ws --type FUNCTION --name aapp131_added                # 在 APPEND 锚点插入新点
```

`rename`/`newfn` **只改 workspace**（不碰 `.tzc`），并当场跑一遍结构事务预检（与 `apply` 用同一套
校验，避免"命令行过了、apply 又拒"）。

**`newfn` 造出来的函数带设计器形态的函数头模板**（与 `FunctionGenerator.cs:20-67` 同形），
且**顶部有一个空行**，所以落进 `.tap` 后不会与上一个函数的 `END` 挨着：

```
(空行)
################################################################################   ← 80 个 #
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL <新函数名>(传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
# Return code....: 回传参数变量1   回传参数变量说明1
# Date & Author..: 日期 By 作者
################################################################################
```

模板只填**已知值**（`Usage` 行的函数名），`日期 By 作者` 等占位符逐字保留交给人填 ——
**不自动填日期**是为了让同一输入产出逐字节相同的工作区（红线 R7：时间轴不是功能）。
`newfn` **没有 `--desc`**：描述块固定为模板，要改既有函数的描述用 `rename --desc`。

改名落盘时 `apply` 复刻设计器 `ProgramInformation.Modify` 的事务序列：

```
.tap 里一次改名 = 一对 <point>
  function.旧名  status="d"   ← 墓碑，**保留原 CDATA 逐字节**
  function.新名  status="u"   ← 活点，签名行已改为新名
```

校验 V1–V7（全部 error 级，命中即拦且原包不变）：

| 码 | 判定 | 退出码 |
|---|---|---|
| V1 | 围栏 `fn` == 签名行的函数名 | 3 |
| V2 | 旧名残留引用：可编辑区内 → 你自己改调用点；**只读区段内 → 退出码 4**（改不了，拒绝） | 3 / 4 |
| V3 | 新名与现存 point / TGL 占位符裸名冲突；含本次会话的重复目标与交换式改名 | 3 |
| V4 | 命名规范：普通程序 `<prog>_` 前缀、`.tzf` 为 `<prog>_<topind>_`（设计器仅 INFORMATION） | warn |
| V5 | scope 违反程序类型：`type=B` 强制 PUBLIC；M/G/X/Z/Q 强制 PRIVATE；S/W 可改 | 3 |
| V6 | 描述块里非空行必须以 `#` 开头（**允许空行**，见偏差 DV-2） | 3 |
| V7 | 目标必须是自订定义点（`function./dialog./report.`）且可编辑 | **4** |

描述与 `desc` 字段的关系：**块可直改**（它是可编辑区，apply 后 `desc` 由块重算）；**`desc` 字段
也可改**（视作显式意图，apply 用它重写块）；两者同时改且不一致 → 报 V6 让你只改一处。

### 6.11 行尾策略（编辑器相关）

真实 `.tzc` 的 `.tap` 是**混合行尾**的（元素间 CRLF、CDATA 内 LF；capt110 实测 866 个 CRLF +
8466 个 LF），渲染出的 `prog.full.4gl` 继承了这一点。而**编辑器会自作主张把整份文件的行尾归一**
（VS Code 保存 capt110 时 866 个 CRLF → LF，文件正好少 866 字节）。

因此字节恒等校验对**行尾等价**：

- **受保护字节**（围栏行、只读区内容、结构行、围栏外字节）：CRLF↔LF 的差异**不算改动**。
  这不是放宽安全性 —— 这些字节 tt dev **从不写回包**，包里的它们永远保持原样。`verify` 会给一条
  `gate1.eol-normalized` 的 info 说明。
- **任何其它字节差异**照旧按逐字节拦截。
- **可编辑区**：只有"行尾归一之外还有实质改动"的区域才会被重写；被重写的区域按**你文档里的行尾**
  写入包。未改动区域在包里的字节**一字不变**。

> 复盘：旧实现逐字节比对，导致 VS Code 用户保存文件后 596/605 个 Region 被误判为改动、只读区段
> "被改" → `apply` 以退出码 4 拒绝，**用户只是加了一行注释却什么都写不回去**。
> 对抗用例已收入 `selftest`（「真机事故-编辑器把行尾归一」）。

### 6.12 退出码与危险开关

| 码 | 含义 |
|---|---|
| 0 | 成功 |
| 2 | 包格式错误（zip 损坏、缺必须条目、`ver` 不存在/不匹配、TGL 区段标记不配对） |
| 3 | 验证失败（围栏外被改 / 结构行被改 / 不变量 / FGL 信封 / 装载模拟） |
| 4 | 写入被拒（改只读区、Locked 态改区段、`--only` 范围外、删非 `new="Y"` 的点） |
| 5 | IO / 环境失败（磁盘、锁冲突、源包已被改动、git） |

| 开关 | 语义 | 约束 |
|---|---|---|
| `--allow-sec` | **已废弃**（v2 起由 `unlock` 状态机取代） | 传入会报错并指向 `tt dev tzc unlock` |
| `unlock --yes` | 框架解锁的代价确认 | 无授权时**加 `--yes` 也拒**（R9 无绕过） |
| `--yes`（apply） | 兼容保留，不再用于区段 | 区段的二次确认已上移到 `unlock` |
| `--only` | 部分导出（控制上下文体积） | 未导出的 Region 一律 READONLY，改了退出码 4 |

> ⚠️ 解开框架的后果不可逆：**等于主动放弃「跟随原厂样板自动重产代码」的能力**。
> 这个决定必须由人做，AI 不得自主触发。

### 6.13 已知边界

- **不写 `.4gl`**（红线 R1）。它是服务器 build 产物，与「TGL + 展开点」的渲染结果**本来就不一致**
  （实测 82/105），覆盖即丢数据。
- **不动 `.tgl`**，除了改区段时同步打补丁（`.tap` 的 `<section>` 与 `.tgl` 的同一区段必须逐字节一致）。
- **不增删 zip 条目**、不改条目名、不给 `ver` 加目录前缀。
- 引用标准程序的包（TAP 条目基名 ≠ 根 `prog`）**拒绝写回**（I15）：设计器在这种包里会用
  `CiteTAP` 内容，我们没有样本验证。
- **`.tzs` 的写回不走这条管线** —— 那是 §7 的引擎在做。（v1 的文档写的是"不做 `.tzs` 写回"，
  那句话在 `engine/` 落地之后就不成立了。）
- `.tzg`（ReportCode）在设计器里没有必需条目检查分支，只保证字节透传。

### 6.14 实现 ↔ 设计器源码 对照

本节是 `tt dev tzc` 的**逐条依据表**：每一处行为都指回反编译源码的 `file:line`，或指回
`docs/T100设计器-README.md`（第三方 README 的恢复副本，下文用 `README §x.y` 指代）。
反编译源码根：`D:\我的项目\T100设计器\`。

| 层 | 实现包 | 关键入口 | 设计器依据 |
|---|---|---|---|
| Package 层 | `internal/pkgfile` | `Open` / `Package.Tap/Tgl/Full4gl/Plan/Build` | `TzpManager.cs:150-182`（类型由扩展名定）、`PackageManager.cs:70-100`（必需条目）、`:589-604`（`ver` 按名字取第一行）、`TzpManager.cs:395-404`（只比 Major/Minor） |
| TAP 层 | `internal/tapfile` | `Parse` / `Rewrite` / `SetPointCDATA` / `AddPoint` / `MarkDeleted` / `SetSectionCDATA` / `SetRootAttr` | README §3.3（根元素真名 `add_points`）、§3.6（混合换行）、§3.9（只用 CDATA 感知扫描器；属性「有则改、无则加」） |
| TGL 标记层 | `internal/tglfile` | `TrimEnd` / `FindSections` / `FindPlaceholders` / `FindAnchor` / `ReplaceAnchor` / `PatchSection` | `CodeEditorManager.cs:1694-1715`（四个正则原文）、`:314-352`（GenerateTGL 补丁格式）、`TzpManager.cs:453-457` |
| FGL | `internal/fgl` | `ParseOutline` / `ParseBlock` / `ParseFunction` / `EnvelopeKindFor` | README §4.7、`FglParserQuickHelper.cs:18-23`、`CodeEditorManager.cs:1230-1253` |
| 合成层 | `internal/synth` | `Synthesize` / `ResolvePoint` / `ResolveSection` | `CodeEditorManager.cs:1070-1080`（LoadContent 六步）、`:1142-1176`（锚点展开）、`:1179-1227`（占位符替换/造空点/记 TglTag）、`:1083-1139`、`ProgramInformation.cs:655-666` |
| 围栏层 | `internal/fence` | `Render` / `Parse` / `Pair` | 本文 §6.6 与 §6.15 的偏差表 |
| 验证层 | `internal/verify` | `Gate1` / `Gate2` / `Gate3` | README §3.8（I1–I15）、§5.1（105 包实测的五个坑） |
| 回写层 | `internal/split` | `Split` / `ApplyTglPatches` | `CodeEditorManager.cs:361-397`（SaveADPContent）、`:400-496`（TglTag 折叠）、`:408`（`section_flag="Y"`）、`AddPointModel.cs:1086-1112`（ToXML）、`ProgramInformation.cs:96-108`（order = max+1） |
| 包容器 | `internal/pkgfile/zipraw.go` | `parseRawZip` / `rebuildZip` | **S7 真机实测**（不是源码），见 §6.3 |
| 审计层 | `internal/store` | `Create` / `Open` / `Lock` / `GitInit` / `GitCommit` / `AtomicWrite` | 红线 R6（禁止 `File.Delete`→`File.Create`：`PackageManager.cs:569-572` 是事故模板） |
| 报错定位 | `internal/model/lines.go` + `verify.Finding` | `LineOf` / `LineAt` / `FirstDiffEOL` / `fillPositions` | 见 §6.8 |
| 新函数模板 | `internal/cli/newfn_template.go` | `newFnHeaderTemplate` | `FunctionGenerator.cs:20-67` + 真实包 `desc="\n####…"`（语料 armt100.tap） |

**权限判定链**（点，`AddPointModel.IsEditable`，`AddPointModel.cs:691-747`）—— 实现全在
`synth.ResolvePoint`：

| 门 | 条件 | 结果 |
|---|---|---|
| G1 | `isIndFun && ind_fun != topind && name != "global.memo_industry"` | false |
| G2 | `(topind=="sd" \|\| topind=="") && env=="s" && name=="global.memo_industry"` | false |
| G3 | `!isStandard && cite_std=="Y"` | false |
| G4 | `readonly=="Y"` | false |
| G5a | `edit=="c" && env=="s"`，或 `edit=="c" && env=="c" && IsTopstdMode` | false |
| G5b | `edit=="s" && env=="c" && !IsTopstdMode && src!="c"` | false |
| G6 | `IsTopstdMode`：`src=="c"` → 仅 `status==CREATE`；`src∈{s,m}` → 仅 `status==NULL` 或（`modi_by_topstd=="Y"` 且 MODIFY） | 受限 |
| G7 | `login_user=="topstd" && src=="c" && status!=CREATE` | false |
| — | 自订定义点正文解析不出信封 | false（`structure-unparsable`） |

> `new="Y"` **不参与** `IsEditable`（只网关删除）；`IsSelfDefinition` **不参与**（只关改名）。
> `edit=` 取自 **TGL 占位符行**（`:1203-1206`）。

**权限判定链**（区段，`SectionModel.IsEditable`，`SectionModel.cs:101-133`）：

| 门 | 条件 | 结果 | 备注 |
|---|---|---|---|
| 政策 | 三个锚点区段 | false | **不复刻**设计器在 `type=="G" && section_flag=="Y"` 下对 `other_dialog` 的放行 —— 写锚点正文会把展开后的函数写进框架（红线 R4） |
| 政策 | 未 Unlock | false（`sec-off`） | 对应设计器 SEC 模式未勾选 |
| S1 | `type=="G" && section_flag=="Y"` → 除 `other_function`/`other_report` 外可编辑 | true | 该特例在政策层之后才生效 |
| S2 | `IsReadOnly`（运行期派生：3 锚点硬编码 `:1127-1130` + TGL `readonly="y"` `:1131-1135`） | false | 永不落盘 |
| S3 | 区段属性 `readonly=="Y"` | false | |
| S4 | `IsTopstdMode`：`src=="c"` → false；`src∈{s,m}` → 见 `:124-129` | 受限 | |
| S5 | `login_user=="topstd" && src=="c"` | false | |
| 政策 | `--only` 部分导出 | false（`only-points`） | |

**命名空间三层**（README §3.7）：

```
other.function / other.dialog / other.report   → 集合锚点（只在 TGL 里）
<裸名>（global.memo / input.a.page2.x / …）      → 框架预留插入点（TGL 通常有、TAP 通常没有）
function.* / dialog.* / report.*               → 自订定义点（TAP 有，由锚点注入）
```

- TGL 有占位符而 TAP 无点 → **造空点**（`ProgramInformation.cs:655-666`），常态，`I2a` info。
- TAP 有点而 TGL 无锚点 → `I2b` **error**（代码彻底不出现）。
- TAP 裸插入点无同名占位符 → 孤儿点，`I2c` warn，**必须原样透传**（`AddPointModel.Create` 不设
  `IsLoaded`，`SaveADPContent` 过滤 `IsLoaded` 所以设计器跳过它但仍 `ToXML()` 写回）。
- **同名两次**：真实包里 `status="d"` 的墓碑与 `status="u"` 的 live 会同名共存，内容改写一律命中
  第一个**未删除**元素（`CodeEditorManager.cs:1188` 的过滤条件）。

**不变量 I1–I15 落地表**：

| # | 不变量 | 实现 | 等级 | 退出码 |
|---|---|---|---|---|
| I1 | `{<section>}` 配对、id 唯一、与 TAP `<section id>` 对应 | `tglfile.FindSections` + `verify.Gate2` | error | 2/3 |
| I2a | TGL 占位符在 TAP 无对应点 | `synth` 造空点 + `Gate2` info | **info（正常）** | 0 |
| I2b | TAP 有自订点但 TGL 无锚点 | `Gate2` | **error** | 3 |
| I2c | TAP 裸插入点无同名占位符 | `Gate2` | warn（孤儿点，不丢） | 0（`--strict` → 3） |
| I3 | 区段正文不得含真实自订点的完整函数块 | `Gate2`（用 TAP 已知函数名匹配定义头，只扫区段**自身字节**） | error | 3 |
| I4 | 自订点结构：先剥 `public/private` 再匹配定义头 | `fgl.ParseBlock` + `synth.ResolvePoint` | warn | 0（写它 → 4） |
| I5 | `status ∈ {"", " ", c, u, d}`；删点置 `d` 且保留原内容 | `tapfile.MarkDeleted` + `Gate2` | error | 3 |
| I7 | `ver` 存在且主次版本匹配、字节原样 | `pkgfile.Open` | error | 2 |
| I9 | `.tap2` 与 `.tap` 的 diff 一致 | 语料无 `.tap2`；只做透传 | — | — |
| I10 | `.4gl ≠ TGL + 展开点` | `Gate2` info（提醒不要试图同步） | **info（正常）** | 0 |
| I11 | UTF-8 无 BOM；CDATA 内无 `\a`；新内容无 `]]>` | `Gate2`（`I11` + `I11b`） | error | 3 |
| I13 | 未知条目 / `ver` 字节原样 | `pkgfile.Plan/Build` + `Gate3` 逐条目 sha256 | error | 3 |
| I15 | 引用标准程序的包（条目基名 ≠ 根 `prog`） | `cli.cmdApply` 直接拒绝写回 | error | 2 |
| I-structure | 可编辑点的结构行未改 | `verify.Gate1`（结构行区间与可写区间**不相交**，逐字节比对） | error | 3 |

补充的机械检查（指南未列，属 A3 的实现）：

| 码 | 内容 |
|---|---|
| `gate1.fence-line` | 围栏行的**锚定部分**被改（`fn`/`scope`/`desc` 豁免，交由 V1/V5/V6） |
| `gate1.outside-fence` | 围栏外字节（prefix/gap/suffix）拼接后不等 |
| `gate1.readonly-region` | 只读 Region 的**自身字节**被改（按配对比较，剔除子区间） |
| `gate1.eol-normalized` | info：受保护字节只有行尾差异（CRLF↔LF）→ 按等价处理（见 §6.11） |
| `gate1.region-count` | 基线总数 = 有基线配对的 Region + 被删除的 Region |
| `gate1.delete-*` | 删除授权：只有 `new="Y"` 且可编辑的点可删 |
| `gate1.append-*` | 追加必须在 `[APPEND]` 锚点内、类型匹配、`new="Y"`、名字不与基线重名 |

**v2 状态机与事务的落地位置**：

| 概念 | 实现 | 关键依据 |
|---|---|---|
| `SectionState`（Locked/Unlocked） | `model.SectionState` + `.tdev/section-state` + `manifest.section` | `ProgramInformation.cs:219-230` |
| 解锁授权闸门 | `synth.CheckUnlockPermission`（三分支 + 设计器原文常量） | `CodeEditorMainWindow.xaml.cs:557-588`；文案 `langs/zh-cn.xaml:448/641/688/689` |
| 解锁重渲染 | `cli.cmdUnlock`（只翻围栏旗标；**编辑文件与基线都翻**，绝不把待落盘的正文编辑吸进基线） | 唯一允许工具改写围栏行的操作 |
| 区段写入联动 | `split`（TglTag 折叠 + `.tap`/`.tgl` 双写 + `section_flag="Y"`）+ apply 末尾 adzi520 提示 | `CodeEditorManager.cs:400-496`、`:408` |
| 结构事务检测 | `verify.DetectStructural`（复用 `fence.Pair.StructFieldChanged`） | |
| 改名事务 | `split`：RemoveTombstones → RenamePoint → SetPointCDATA → SetPointAttr(status=u) → AddPoint(墓碑 status=d) | `ProgramInformation.cs:116-134`；`SetName` `AddPointModel.cs:810-830` |
| 新增点 | `cli.cmdNewfn` + `split` 的 AddPoint | `AddPointModel.cs:1008-1060` |
| apply 基线重算 | **以写出的新包为准重新 synthesize + render**，并同步刷新 `prog.full.4gl` | 改名会切换点身份，只有从包重渲染基线才与包一致 |

### 6.15 与设计指南的偏差（均已核实，逐条给出理由）

| # | 指南原文 | 实现 | 理由 |
|---|---|---|---|
| **D-1** | 新增点写 `status="c"` | **写 `status="u"` + `new="Y"`** | `AddPointModel.ToXML()` 在 `Status == CREATE` 时直接 `return null`（`AddPointModel.cs:1090-1093`）；设计器下一次保存会**静默丢弃**该点。设计器自身新增点也走 `CREATE\|MODIFY → "u"` |
| **D-2** | 围栏严格配对、**不嵌套** | 允许 `section → point` 一层嵌套，深度 > 2 报错 | 真实 TGL 里自订点住在区段内；平面围栏无法表达 |
| **D-3** | 禁止无法归属 Region 的散行 | 允许未归属字节段（prefix/gap/suffix），但**全部纳入 gate1 字节恒等**并登记在 `regions.json` | 真实包里必然存在（区段之间的空行）；禁止它们会让 105/105 个真实包直接失败 |
| **D-4** | `fgl_parse_function` 报解析错误 | 契约收窄为**块信封良构性**（7 个错误码），不做语句级语法校验 | 三个候选实现都不做语句级校验，且本机无 `fglcomp`/`fglgo` |
| **D-5** | 工作区只有 `.tdev/base.sha256` | 增加 `.tdev/base.full.4gl` | gate1 必须与「AI 实际看到的那份」比对，且 `status` 不应强依赖 git |
| **D-6** | 未定义 | `apply` 前校验源包 sha256 未变，变了 → 退出码 5 | 避免对着错误的包写回 |
| **D-7** | `apply [--yes]` 无语义 | 写区段必须走 `unlock --yes` | 「解开框架」是不可逆架构决策，需人工二次确认 |
| **D-8** | 不把时间戳当功能 | 写 zip 时沿用每条目原始 method/modtime（不取时钟），未知条目字节透传 | 结果确定、可复现 |
| **D-9** | CDATA 改写 | 新内容含 `]]>` → 退出码 3（I11b） | XML CDATA 装不下 `]]>`（语料 36,324 段实测 0 例），设计器会拆成两段 CDATA |
| **D-10** | — | 区段 id 与点名同名时按**配对**处理，不按名字查表 | 真实包存在 `prog.process` 既是 `<section id>` 又是区段内裸名插入点 |

**v2 新增的三条「文档 vs 源码」偏差（源码直证）**：

| # | 文档 | 源码 | 实现 |
|---|---|---|---|
| **DV-2** | V6：描述块「每行 `#` 开头、**无空行**」= error | `AddPointModel.cs:253` 的判据是 `!Regex.IsMatch(line,"^\s*#") && line.Trim() != ""` → **空行合法** | V6 允许空行；且只对**本次改动过描述块**的点报 error，存量数据降为 warn（真实语料有历史遗留的不规范行） |
| **DV-3** | V5：「type=B 强制 PUBLIC；M/G/X/Z/Q 强制 PRIVATE」= error | `FunctionInfoWindow.xaml.cs:143-192` 的 type→scope 映射是**新建点对话框的默认值**，整段被 `useDefaultScope` 门控 —— **不是对既有文本的硬约束**（真实语料 `aapt300(c).tzc` 就有 type=M 而函数声明 PUBLIC 的点） | 只在**本次发生 ScopeChange** 时给 **warn**；`newfn` 造新点时按该默认值填 scope |
| — | 要求「事务等价测试与设计器手工改名产物做语义 diff」 | 设计器无法进 CI | 自检用**独立参考断言**墓碑对形状；真机对拍列入 S7 人工项 |

> 这几条与 D-1 同源：**把真实数据里的常态当 error，会拒掉真实包。**

### 6.16 实测基线

| 项 | 数据 |
|---|---|
| 语料 | `D:\t100_wrok_dir`：166 个 `.tzc`（hengshuo 107 / xiyuan 34 / wq 25），全部固定 4 条目 `4gl/tap/tgl/ver` |
| `ver` | 全部 `"1.0\n"`（4 字节，无 BOM） |
| 换行 | `.tgl` 纯 LF；`.tap` 混合（元素间 CRLF、CDATA 内 LF）；105 包实测 tap CRLF 289,763 / 独立 LF 1,108,074 |
| CDATA | 36,324 段；**0 段含 `]]>`**；**0 个 0x07**；**0 个 BOM** |
| `<point>` / `<section>` | 32,678 / 3,646 |
| 同名点（tombstone+live） | 4 处 |
| 区段 id 与点名同名 | 8 个包（如 `wssp00316.process`） |
| 退化点（`function.*` 但正文只有注释） | 1 处：`s_axmt500(s).tzc` / `function.memo_industry` |

> 语料是**活的目录**（人也在里面干活）：包数/Region/点数的绝对值随语料增删而变，
> 换机器或语料变动后重跑 §9 的命令即可刷新；**判据（0 error、逐字节一致）不随规模变**。

---

## 7. `tt dev tzs` 表单包

### 7.1 引擎与设计器程序集

`tt dev tzs call` 背后是 `engine/` 里一个 **C# 引擎**（`tzs-server.exe`）。它**不实现** `.tzs`
格式 —— 它 `Assembly.LoadFrom` **设计器自己的程序集**，布局属性走设计器自己的 `XmlElement`
索引器，`.tsd` 由设计器从模型重算，验收用设计器自己的校验器加 RoundTrip 不动点。
我们这部分总共 200 KB（`TzsCli.dll` 26 KB + `TzsCli.Designer.dll` 180 KB）。

`tt` 通过命名管道上的 JSON-RPC 驱动它（`internal/dev/tzs/`，`tt` 自己实现的 Go 客户端）。

**设计器的 28 个程序集随仓库分发**，在 `engine/designer/`；发行包带同样一份（`tzs\designer\`）。
引擎的解析规则只有两条：`TZSCLI_INSTALL`（覆盖用），否则 `<引擎自己的目录>\designer`。
**没有"回落到别处装的那份"这一条** —— 少带一个文件是包/仓库不完整，不是"去别处找找"。

> **为什么入库而不是让用户自己装。** 同一份 tt 在两台机器上，如果设计器目录来自各自的配置，
> 两边跑的就是两版设计器 —— 同一个 `.tzs` 行为不同，而报错里看不出来。跟着仓库/包走之后，
> "运行环境一致"是**分发这件事本身**保证的。代价是版本变更进历史（约 11 MB / 版，git 压缩后
> 约 4.4 MB）；好处是那次提交就是"这一版钉在哪一版"的记录。
>
> 只采 `*.dll`。GUI 的 `T100Designer.exe`、会连厂商更新通道的 `AutoUpdater.exe` 与 Sparkle
> 组件、它的 `.config` 和一个快捷方式都不在内。

**引擎为什么单独构建**（见 `engine/BUILD.md`）：

1. **它不属于 Go 的构建链。** 用 `csc.exe`（Framework64 v4.0.30319，**C# 5** —— 没有模式匹配、
   没有 `nameof`、没有字符串插值）编译，引用 GAC 里的 WPF 程序集。打包脚本只**采集**产物。
2. **重编会让所有在跑的守护进程变成孤儿。** 守护进程的管道名 = `hash(工作区)` + **本程序集的
   MVID 前 8 位**，MVID 每次重编都变。客户端因此**永远够不到**跑着陈旧字节的守护进程
   （刻意的，否则你会和旧行为对话而看不出来）；代价是每次重编后，上一个构建起的守护进程
   **再也停不掉**（新名字没人监听，旧名字没人知道）。清理由 `tt dev tzs reap` 兜底。

   把引擎构建挂在 `tt` 的每次构建上，等于每次发版都制造一批停不掉的进程 —— 一个和 `tt` 无关的
   构建步骤造成用户可见的后果。

### 7.2 命令面

| 命令 | 作用 | 要引擎吗 | 改包吗 |
|---|---|---|---|
| `export <pkg> [-o <dir>] [--force]` | 纯解压（只读参考，不依赖引擎也不依赖设计器） | 否 | 否 |
| `fns [<fn>]` / `manifest` | 函数表 / 单个函数的参数；函数表 JSON 原样转发 | 是 | 否 |
| `call <fn> --<参数> …` | 读写表单，**唯一写路径** | 是 | 只经 `save --out` 写到**新**包 |
| `doctor` | 环境自检（引擎 exe / 设计器目录 / 工作区 / 管道名 / 守护进程） | 是 | 否 |
| `stop` / `reap [--yes]` | 停本工作区的常驻引擎（不启动新的）/ 清理重编后停不掉的孤儿 | 否（要配工作区） | 否 |

**49 个函数**按组分（`tt dev tzs fns` 看全表；会改模型的标 `[写]`，慢的标 `[slow]`）：

| 组 | 函数 |
|---|---|
| 会话 | `open` `save` `close` `verify` `list_open` |
| 读 | `form_tree` `find_component` `get_component` `list_spec_nodes` `describe_kind` `list_tables` `list_columns` `list_records` `list_local_strings` |
| 属性 | `set_spec_attr` `set_layout_attr` `set_tree_source` `rename_component` |
| 结构 | `add_widget` `add_field` `insert_at` `delete` `move` `reparent` `nudge` `align` `fit_size` `wrap` `break_layout` `convert_widget` `convert_container` |
| 页签 | `add_page` `delete_page` |
| 语义/Action | `insert_semantic` `add_action` `delete_action` `set_action_types` |
| 多语言/选项/串查 | `set_local_string` `set_items` `set_progrel_programs` `set_table_association` `set_spec_description` `set_cited` |
| Tab 顺序 | `set_tab_order` `tab_action` |
| 校验/工具 | `validate` `base_data` `set_excluded` `set_code_template` |

几处容易记混的：**`move` 只改 Z 序**（同一父容器内前后挪），**换父容器用 `reparent`**
（设计器的拖拽 `DragComponentsUndoRedoCommand`，仅限同表单）。**`add_field` 加字段，
`wrap` 是拿选中元素包一层新容器**。`add_field --columns a,b,c` 一次构造多列，`--container Table`
才得到"一个 Table 装 N 列" —— 一次一列会让大表单慢两个数量级（实测 84 列从 137 秒降到 1.17 秒，
而且会在 Table 里给每个控件塞一个 Label，真实表单里没有那种形态）。

### 7.3 句柄

- **每个需要句柄的函数都必须显式给 `--handle`。** 没有"自动沿用上次句柄"这回事。
- **句柄永不复用**：`close` 掉一个包再 `open` 同一个包拿到的是**新号**。拿旧号去调 →
  `E_NOT_FOUND`「句柄不存在或已关闭」（退 2）。这是安全失败，不是 bug。
- **同一个程序已经开着再 `open` → `E_KEY_IN_USE`（退 4）**，报错会说清是哪个文件占着。
  先 `close` 占用者（`list_open` 看是谁），或 `open --force`（只对 `Loaded` 状态的占用者有效）。
- 句柄只活在常驻守护进程里。**进程一死全部失效** —— 守护进程挂了就重新 `open`，别重放写请求。

### 7.4 `validate` 是基线相对的

`validate` 在**句柄上第一次被调用时，那次运行本身就是基线**（`Fns/Validate.cs:139`）。所以第一次
调用 `newErrors`/`newWarnings` **按构造就是空** —— 它证明不了任何事。要看"我改坏了没有"，
必须在**同一个句柄**上先 `validate` 一次建立基线。

实测：`aapt300(c).tzs`（未做任何改动）首调 `baseline = 13 条 WARNING`，`after = 13`，
`newErrors`/`newWarnings` 都是 0。**语料里本来就不干净** —— 判据永远是「改动后的增量」，
而不是「有没有 WARNING」。

返回体 `{baseline[], after[], newErrors[], newWarnings[], elapsedMs}`；加 `--json` 打**整帧**
（`{id, ok, result, error, ms}`）而不是只打 `result`。

### 7.5 参数语法

本地只做**语法**校验，语义交给引擎：

| 写法 | 含义 |
|---|---|
| `--handle h9` / `--handle=h9` | 两种都行；`=` 后允许空串 |
| `--paths a,b,c` | 列表按逗号切 |
| `--paths a --paths b` | 列表**可重复**，多次出现会合并 |
| `--force` | 裸开关 = `true`（`--excluded`/`--cited` 同理） |
| 裸词 `foo` | **报错**。位置参数一律不接受 |
| 标量给两次 | 报错（只有 `path[]`/`string[]` 可重复） |

**本地绝不拦的**：`attr` 白名单、`kind` 的取值、`add_action` 的 `type` —— 这些合法集要从
**活着的模型**或**该表单自己的** `s_detail<n>` 记录里取，静态判不了。所以 `--attr bogus_attr`
会照发，由引擎用它自己的白名单拒绝。

两个附带事实：**参数打错的报错也要引擎在**（`call` 先拉 manifest 再校验参数）；`--timeout`
**最小 120 秒**，低于此值会被抬到 120（下限是给加载看门狗留的）。

### 7.6 退出码

**没有 `3`**（`3` 是 `.tzc` 那条线的"验证失败"），多一个 `1`：

| 码 | 含义 |
|---|---|
| 0 | 帧 `ok:true` |
| 1 | 引擎内部错（`kind=internal`，含 `E_NOT_IMPLEMENTED`） |
| 2 | 参数/环境不对：本地参数错、未知函数、manifest 拉不到、引擎的 `validation` 与 `not_found` |
| 4 | **设计器拒绝**（`kind=designer`，含 `E_KEY_IN_USE`） |
| 5 | 传输或环境失败（含加载超时、没配工作区） |

**红线**：`export` 的产物是**只读参考**，不要手工改完再塞回包，也不要拿它当 `tzc` 工作区去
`apply`；`save --out` **永远指向新包**，绝不指向源包；**请求一旦上线绝不重试**（协议无幂等键，
重试是在赌"上一次写进去了没有"）。

---

## 8. `tt dict` 数据字典

### 8.1 能力与数据族

各查询命令依赖不同**数据族**：表字典 / 校验带值 / 系统分类码 / 字段画面规格 / 可复用开窗 /
系统消息 / 参数定义 / 程序与作业。`tt dict db status` 一次看清哪个族没同步、各多少行、缺哪些表
（只读，不改文件，也不需要 `config.json`）。

**读法**：`✓` = 该族已完整同步；`✗`/`✗ 部分` = 该族字典表不全，**依赖它的命令会整体报错**
（不是部分可用 —— 所以行数列显示 `—`，那个行数没有意义）。补齐两条路：`tt dict db sync`，
或临时用 `tt dict <命令> --env <环境名>` 免写盘直查。

### 8.2 查询命令

> 命名对齐 T100 原生工具/术语：`r.t`（数据表）、`r.v`（校验带值）、`r.q`（开窗）、`desc`（字段
> 规格）、`scc`（分类码）、`prog`（程序与作业，别名 `program`/`job`）。短名 `rt`/`rv`/`rq` 与旧名
> `table`/`check`/`win`/`spec` 仍作别名可用。所有命令支持 `--json`（结构化）、多数支持 `--csv`、
> `--lang zh_CN|zh_TW`（默认 `zh_CN`；搜索用字要跟语言别对应）。

**`tt dict r.t [表名]`** —— 一张或多张表的完整字典：表说明/所属模块/表类型、每个字段的中文含义
与类型主键必填、键值与索引。读代码、看 SQL、查界面字段含义时用它。不给表名是**列表模式**
（`--kw <关键字>` 按表名或中文说明过滤 —— "从业务词找表"走它）；指定多张表时默认打印每张表的
**全部字段**（7 张表 ≈ 700 行），**只要表级信息加 `--brief`**；`--who` 反查"改这张表会影响谁"
（见下 `prog`）。

> 表说明有简繁两种写法，同一词也可能有异体（`对账`/`对帐`/`對帳`）：搜不到时换个写法、更短的词，
> 或加 `--lang` 切换语言别。

**`tt dict r.v [dzcd001]`** —— **字段校验规则**：系统保存数据前做的各种检查是可复用的"校验模板"，
每条校验一个识别码（形如 `v_ooba002_07`）。详情含：单头（型态 1=检查存在 / 2=带值 / 3=检查存在
并带值、错误讯息代号、行业别、状态码）、**SQL 指令**原文（附标签图例：`<field>` `<table>` `<wc>`
`<count>`、`arg1~9`、`:TODAY`、`:DEPT` 等全局变量）、参数（对应 arg1~9、日期型态）、判断条件
（条件 SQL + 成立时显示的错误讯息）。

**`tt dict scc [gzca001]`** —— **系统分类码**：系统里各种下拉/选项的"选项字典"（币别、单据类型、
料件属性…），一个分类码 = 一组「值 → 说明」，画面上的下拉选项就来自这里。详情含单头（群组、
状态、名称）、分类值（值、说明、排序、标准/客制）、**扩展数据列** `gzcb003~gzcb015`（含义由各分类码
自己定义 —— 如某分类码用它们存群组编号区间、另一个存格式，要结合该分类码的语境理解）。

**`tt dict desc <表名> [字段名]`** —— **字段的画面规格**：每个字段在画面上用什么控件、下拉取哪个
系统分类码、显示格式与宽度、必填与否、默认值、开窗程序、最大/最小值等。画面设计器按它生成画面。
输出项含义：**控件**（03=ComboBox 05=Edit 01=ButtonEdit 04=DateEdit 09=RadioGroup 11=SpinEdit
12=TextEdit 13=TimeEdit 34=DateTimeEdit 等）、**SCC码**（下拉来源，用 `tt dict scc <码>` 查可选项）、
显示宽度/格式/大小写/默认值/最大最小值、编辑与查询开窗、**校验带值**（用 `tt dict r.v <码>` 查内容）、
串查型态与程序、报表栏宽与小数位、客制识别。

**`tt dict r.q [dzca001]`** —— **可复用开窗**：代码里 `CALL q_xxx()` 弹出的查寻选单定义。
每条开窗 = 一段带占位符的 SQL（`<field>/<table>/<wc>`）+ 外部参数（arg1~9）+ 显现与回传列。
详情含单头（状态、SQL 原文、每页笔数、作业串查编号、HardCode）、参数（日期型态）、**显现设定**
（显现顺序、字段编号、表格别名、显示控件、**是否回传** —— Y 的字段按显现顺序构成 `return1~9`）。

**`tt dict msg [编号]`** —— **系统消息**：提示/报错消息的完整文本、建议处理、建议作业、技术细节。
查询条件即作业 `azzi920` 查询画面的那几栏、同时成立：**编号**（gzze001，位置参数、逗号分隔可给多个）、
**信息语句**（gzze003，`--text <关键字>`）、**信息类型**（gzze007，`--type 0|1|2`，SCC 106）、
**状态**（gzzestus，`--status Y|N`）、**建议运行作业**（gzze005，`--prog <作业编号>`，用来反查
"哪些消息建议跑这个作业"）、**语言别**（gzze002，`--lang`，缺省 `zh_CN`——它在 WHERE 里，不是
事后过滤）。类型/状态/建议作业都可逗号分隔多选（同栏 OR），不同栏之间是 AND。
编号格式 `std-00001`/`azz-00041`/`lib-xxxxx`；负整数（SQLCODE 如
`-100`）需 `--` 分隔：`tt dict msg -- -100`。返回**文本**、**建议处理**、建议作业、技术细节、
**类型**（0=警告/1=错误/2=资讯）、状态。命中一条打详情块，命中多条打列表（编号/类型/状态/语句）
——搜索能命中几十上百条，要看某条的完整内容再拿编号精确查一次（`--full` 可强制全部打详情）。
编号栏**默认精确、写 `*` 才通配**，这是与 azzi920 画面唯一的行为差异（画面里编号栏被子模板
`cl_ap_code_fuzzyquery` 加了 `*…*` 变成子串）：CLI 的常见用途之一是查 SQLCODE，模糊匹配 `-263`
会把 `-1263`/`-22263`/`-26300…-26312` 一并捞回来。信息语句两边一致，默认子串。
**枚举一律写码不写中文**（`--type 1`，不是 `--type 错误`）——与家族其它命令一致；填错时报错自带
`0=警告 1=错误 2=资讯` 的对照表。语言别不算"筛行条件"：只给 `--lang` 等于没筛（那是把整张表换个
语言看一遍），其它五个任给其一即可。源系统是精确 (编号, 语言) 匹配、无自动回退；零结果时命令会
同条件去掉语言补查一次，把可用语言列出来。

**`tt dict sysp <编号>` / `tt dict docp <编号>`** —— **系统/单据参数说明**（决定系统行为与单据流程
的可配置项）。`sysp` 查系统/企业/据点级参数（编号 = 型态码 + 领域 + 4 位流水：`A-SYS-0040` A=系统级、
`E-CIR-0001` E=企业级、`S-BAS-0028` S=据点级）；`docp` 查单据别参数（`D-` 开头，如 `D-MFG-0076`），
并附该参数适用的单据性质。返回名称/说明（多语言）、参数群与级别、**型态**（1=Y/N 2=整数选项
3=范围设定 4=字符或SCC 5=日期）、领域、预设值、值域、SCC 选项、校核/开窗引用、取参异常处理、
修改频度、即时抓取、状态与长备注。

> 注意：这里查的是参数的**定义与说明**，各环境下参数**实际设的值**在参数值维护作业/画面里，
> 不在本命令范围。

**`tt dict prog [程序编号]`** —— **程序与作业登记**（azzi900 / azzi910）：程序编号对应的中文作业名、
程序类别、归属模块、是否客制、引用主程序与系统运行指令，以及**哪些作业用了这个程序**。

- 作业通过 `gzzz_t.gzzz002` 挂到程序上，**一个程序可以被多个作业使用**（共用维护程序可达数百个
  作业）；作业的显示名称取它所挂程序的名称（`gzzal_t.gzzal003`）。拿作业编号也能查，会自动跟到
  它挂的程序。
- **「程序 ↔ 表格」两个方向都能查**（数据来自 `gzdg_t` 程序与应用表格功能分析表，由 T100 自己
  维护，参考作业 azzq902）：`tt dict prog <程序>` 给"用了哪些表 + 操作 S/I/U/D"，`tt dict r.t
  <表> --who` 反查"改这张表会影响谁"（`--limit` 默认 20，0 = 全部）。
- **比 grep 准**：grep 会把被注释掉的 SQL、字符串里的表名、`LIKE 表名.字段` 的变量声明都算进来，
  还混进 `type_t`；这个索引只登记**真实 SQL 访问**并带操作类别。覆盖库与元件（实测 `cl_abi` 22 条、
  `s_apcp300` 8 条、`q_adzi052` 9 条），所以 `tt dict prog <库名/元件名/开窗码>` 也能用。
- **边界**：**子程序**（`aapq110_01` 这类）与**动态 SQL**（`CURSOR FROM 变量`）不入索引，
  这些仍需 `grep` 兜底。
- **子程序/元件/库的登记**来自 `gzde_t`，与主程序登记（`gzza_t`）**互补且不重叠**（实测
  `gzza_t` 4,147 / `gzde_t` 4,036，交集 0）。规格类别（SCC 91）：`B`=应用元件 / `S`=子程序 /
  `G`=报表元件-GR类 / `X`=报表元件-XG,FR类 / `K`=报表组件-XR类 / `W`=WebService元件。
  **子程序与它的主程序是两回事**（`aapq110_01` 是"报表打印"、主程序 `aapq110` 是"明细查询"）。
- `tt dict prog --kw 对账` 按编号或中文名称搜索；`--sub` 在子程序/元件里搜。

### 8.3 数据维护与镜像

**`tt dict db sync`** —— 从 ERP 把本地数据刷到最新（覆盖全部数据族）。全量约 85 万行，视网络
2~4 分钟。采用临时库整体替换，**失败不影响原库**，完成后原库自动备份为 `<库>.bak`。
`--table` 只刷部分；`--env <环境名>` 指定数据源；`-d` 指定库文件。**通常不必用 CLI** ——
`tt serve` 的「数据同步」视图可选环境并显示逐表进度。

**`tt dict db list|ping|discover`** —— 连接管理（与查询数据源独立）：按环境列出挂载的库 /
验证连接可达（只读）/ SSH 自动发现连接要素（`--save` 写入该环境 `db`）。`db discover` 给出的是
**"服务器视角"**的连接要素；客户端不可达时改 `db.host` 或补 `viaSsh`。

**`tt dict mirror`** —— 把某环境的源代码镜像到本地，供 **AI 用本地文件工具读码**，避免每次读取
都经服务器往返、也避免给 AI 服务器端查询权限。

- **只镜像**：各模块 `4gl`（源码）、`4fd`（前端字段描述）、`42s`（编译字符串，**只取 `zh_CN`
  语言目录** —— 文件所在目录的最后一个目录必须是 `zh_CN`）、以及 **`*.inc`**。per（界面源）、
  编译产物（42m/42r/42f）、其它语言、设计器辅助目录一律不拉。目录与服务器同构，客制模块
  `c<mod>` 与标准模块并列，**同路径下客制优先**。
- **白名单判的是目录，不是扩展名**：`-path '*/4gl/*'` 命中 4gl 树下的**所有**文件，所以服务器
  4gl 目录里混放的副产物会跟着进镜像（实测 `erp/azz/4gl/azzi920.4gl.src`、`azzi310_del_str.42r`、
  `adzp030.42m`）；只有放在正常编译输出位置的 42m/42r/42f 才真的不会被拉。
- **新增文件类型自动补齐**：白名单带版本号（记录在 `.tdict-mirror.ok`）。白名单升级后，下次
  「增量更新」检测到版本不一致会**自动转全量**一次，不必手动 `--full`。
- **备份/临时文件也排除**（`*.bak`/`*.bck`/`*.bck1`/`*.old`/`*.orig`/`*.tmp`/`*.swp`/`*~` 等）；
  历史上已拉取到本地的备份文件会在下次拉取时自动清理。
- **增量机制**：服务器 TOP 下保留 `.tdict-mirror-<环境>.mark` 基线，默认只打包 `find -newer` 的
  变更文件（秒级）；首次与 `--full` 走全量（数百 MB~数 GB，gzip 打包 + SFTP 流式下载）。
- **打包根（TOP）解析**：T100 路径**不允许静态配置** —— 打包根按登录区域在服务器上动态探测获取
  （登录 zone → 环境脚本回读 TOP/ERP/COM），**探测失败即报错**（检查该环境 SSH 登录与 zone 设置）。
- AI 工作流：`tt dict mirror path` 拿目录 → 本地工具读码。

**`tt dict bdldoc`** —— 项目内随带 **Genero BDL(4GL) 语言参考文档**（markdown，已排除图片，
约 5,200 篇，含 `llms.txt` 全文索引），供 AI/开发者查询 BDL 语法、内置函数、fgldb 调试器命令等
语言级问题 —— **写 4GL 代码、读调试器输出遇到语言问题先查它，而不是猜**。存放路径记录在
`bdldoc.dir`，命令只改配置、不移动文件。速查：语言基础/进阶在 `08_language-basics`/
`09_advanced-features`，SQL 在 `10_sql-support`，画面/报表在 `11_user-interface`/`12_reports`，
fgldb 与工具在 `13_programming-tools`。

### 8.4 查询数据源切换（在线 ⇄ 本地库）

`r.t/r.v/desc/scc/r.q/prog` 统一走同一查询接口（`db.Source`）：本地 SQLite 镜像与远程 ERP 库
查的是**同一批表**，输出完全一致；命令层不感知数据源。解析优先级：

1. `--env <环境名>` / `--env local`：本次调用**强制**走该数据源；
2. `config.json` 顶层 `query.source`：**环境名** → 在线直查该环境；**`"local"`** → 本地 SQLite；
3. **缺省 = 在线**：用默认环境（`hosts.activeEnv`）直查；一个环境都没配（新装/便携包首次）时
   才用本地库。

**两者不自动切换**：在线就是在线、本地就是本地 —— 选在线时连不上就**直接报错**（不会偷偷改查
本地），报错里会提示怎么切。

> **为什么缺省在线**：远程是最新且完整的数据，本地副本要靠 `db sync` 更新（新机器/便携包上甚至
> 还没有）。离线作业就在设置页切成「本地 SQLite」，或长期固定本地。

远程直查使用客户端驱动直连 `db.host:port`，不要求本地已 sync；金仓与 Oracle
均为完整支持（列表 `--kw` 过滤大小写不敏感，与本地一致）。查询是只读单条 SELECT，值经白名单/
转义内联。

**用哪个账号**（2026-09 起统一）：由环境的 **`topent`（企业编号）查 `gzou_t` 解析**，与
`tt debug sql` 同一条规则、同一张表、同一个快照文件（`internal/entdir`，`<配置目录>/ents/`，
10 分钟新鲜期）。两条路径的区别只在传输方式：debug 在 T100 服务器侧跑 sqlplus/ksql，
dict 从客户端直连。

- `topent` 未配或不是数字（如把据点码 `DSCNJ` 填在了那里）→ 退回**账号列表首项**，
  并在输出的 `notes` 与 `accountSource` 里如实标出，不静默；
- 解析出的账号不在环境账号清单里 → 仍按它连（密码走 T100 惯例「账号=密码」），同样记一条 note；
- **快照未命中时要多连一次**（先以列表首项读 `gzou_t`，再以解析出的账号重连），命中后零开销。

这条规则改变的是**行为而不只是显示**：若某环境的 `accounts[0]` 不是该企业对应的 schema，
升级后 `tt dict` 会读到另一个 schema 的数据 —— 这正是要修的（同一张字典表被两个 schema 读到
而结果里没有任何信号）。看 `account`/`accountSource` 两个字段即可确认本次到底落在哪。

### 8.5 查询输出：一个信封，三种形态

> **本节是细节，契约在 §3.6。** 两者的分工：§3.6 定"不许破的四条"（给改的人看），
> 本节展开"为什么是这四套东西合过来的"（给查历史的人看）。改输出层时两处都要对得上。

所有 DB 查询命令（`tt dict` 的 r.t/r.v/scc/desc/r.q/msg/sysp/docp/prog、`tt debug sql`）
共用**同一个输出出口**（`internal/output`）与同一份环境信息块（`output.Meta`）。此前这里是
四套并存的东西：JSON 发射器四个（其中两个字节相同地重复）、CSV 写入器三个（一个死代码、
一个吞掉所有错误、行尾还各不相同）、表格渲染四种。

**默认形态是 JSON**（`--format json`，也是 `--json` 的糖）。依据是表格类格式对 LLM 的理解力
实测（JSON 52.3% / Markdown 表格 51.9% / CSV 44.3%），而 JSON 又省掉了解析歧义。三种形态：

| `--format` | 输出 |
|---|---|
| `json`（默认） | 一个信封：`ok` + 环境信息（平铺）+ `data` |
| `csv` | `# ` 前缀的环境头 + 裸 CSV 主体（`grep -v '^# '` 一刀切下去就是干净 CSV） |
| `table` | 人读对齐表格（分隔线按**显示宽度**算，中文表头不再歪） |

```jsonc
{
  "ok": true,
  "source": "live", "env": "主机正式区", "sshHost": "172.16.1.109", "zone": "36",
  "topent": "DSCNJ", "ent": 99, "account": "dsdemo", "accountSource": "accounts[0]",
  "target": "172.16.1.109:1521/t35prd", "dialect": "oracle", "route": "client-direct",
  "readonly": true,
  "totalRows": 65, "returned": 65, "elapsedSeconds": 0.18,
  "notes": ["环境的 topent 不是数字(DSCNJ)…"],
  "data": [ { "编号": "azz-00110", … } ]
}
```

CSV 的 `# ` 头（DEBUG 侧先有的写法，现在两处同一个实现）：

```
# 环境 主机正式区 · SSH 172.16.1.109 · 区域 36
# 账号 dsdemo · 来源 accounts[0] · 库 oracle · 172.16.1.109:1521/t35prd · client-direct · 只读事务 · 用时 0.18s
# 行数 65
```

几条定死的规矩：

- **返回条数有全局上限**（`config.json` 的 `query.limit`，缺省 20；`--limit` 覆盖本次，`0` = 不限）。
  它不是"显示上限"而是**返回上限**，三种形态一视同仁：一次宽查询足以打爆 agent 的上下文，
  实测 `tt dict msg --type 1 --status Y` 是 28005 条 / 9.3 MB。
- **截断绝不静默，结果绝不丢**。被截断时**先**把完整结果落盘到 `<配置目录>/spill/`，再截；
  信封里同时给 `truncated` / `returned` / `totalRows` / `localPath`，`notes` 里再写一句人话，
  CSV 的 `# ` 头写三遍（行数行、落盘行、注）。**落盘失败就不截断** —— 宁可给一坨大的，
  也不给一份看不出少了东西的假结果。
- **单实体详情不截**（`emitDetail`）：`tt dict r.t dzea_t` 是"这一张表的 31 个字段"，
  按 20 行切下去会把字段切一半，那不是结果集的一页。
- **环境字段全可空，空的整段不跳出来** —— 本地源拿不到 SSH 就不会出现 `环境  · SSH  · 区域 `
  这种空壳行。
- **企业编号与账号同在一行** —— 「哪个企业 → 哪个账号」是一条链，拆开会被读成两件独立的事。
- **0 行要点名"上错号/上错环境也是 0 行"**，截断要给总数下界（`≥N`）。
- **`--json` 时 stdout 只有那一份信封**：诊断与提示一律进 `notes` 或 stderr。CSV 模式的 `# `
  头留在 stdout 是刻意的例外 —— 头必须随数据一起走，管道里没法重新配对。
- **错误也是信封**：`{"ok":false,"code":"TABLE_MISSING","error":…,"hint":…,"exitCode":3}`，
  `code` 是稳定契约（`USAGE`/`ENUM_INVALID`/`ENV_UNKNOWN`/`TABLE_MISSING`/`CONNECT_FAILED`/
  `QUERY_FAILED`），agent 靠它分支、人读 message。
- **退出码**：0 成功（含"查无结果"）/ 1 用法错 / 2 数据源错 / 3 缺表。0 与 1 沿用既有语义，
  只新增分类码 —— 重新编号会把既有脚本的判据改掉。
- **缺表 = 退出 3**（`TABLE_MISSING`）。从前它在十来处是"打一句提示、退出 0"，脚本会把
  "没查过"读成"查过了，没有"。

⚠️ 这是**破坏性变更**：`--json` 的顶层从裸数组/裸对象变成了信封，载荷移进 `data`。
消费方要改 `.[]` → `.data[]`。

### 8.6 输出中的类型码

**表类型**（`r.t` 的「类型」列，每张表在系统里的角色）：

| 码 | 含义 | 码 | 含义 |
|---|---|---|---|
| B | 基础数据 (Basic) | L | 多语言 (Language) |
| M | 主档 (Master) | V | 提速档/视图 (View) |
| T | 交易单头 (Transaction header) | X | 系统/交叉 (Cross-reference) |
| D | 交易单身/明细 (Detail) | H | 历史/暂存 (History/Temp) |

**字段数据类型**（`r.t` 字段表的「数据类型」列常见值）：`N004`/`N101` 数字；`C003`/`C004`/`C105`
字符；`D001` 日期；`Z012` 时间戳；`Z501` 一般 flag (Y/N)；`Z509` 短说明；`Z510` 长说明；
`Z521` 文件编号 (Table 引用)；`Z522` 字段编号 (Field 引用)；`Z504` 人员编号；`Z505` 组织编号。

---

## 9. 测试与验收

```powershell
go test ./...                        # 全量；缺语料时自动跳过（约 2–3 分钟）
.\tt.exe dev tzc selftest            # 31 项内置对抗用例，不需要真实语料
.\tt.exe dev tzs doctor              # .tzs 引擎环境自检

# 深度语料回归（需显式开启，见下）
$env:TDEV_DEEP="1";  go test ./... -timeout 30m                       # .tzc，约 9–11 分钟
$env:TTZS_DEEP="1";  go test ./internal/dev/tzs/ -run TestCorpus -timeout 30m   # .tzs，约 17 分钟

$env:TDEV_CORPUS="D:\t100_wrok_dir"  # 覆盖语料目录（默认即此）
$env:TTZS_CORPUS="D:\t100_wrok_dir"
```

> **为什么深度回归要显式开关**：`.tzc` 的两个大用例（`TestCorpusExportVerify` 275 s +
> `TestCorpusApplySimulation` 250–380 s）合计 9–11 分钟，而 `go test` 默认 `-timeout=10m`
> （go 命令在 10m+1m 处杀进程，报 `*** Test killed: ran too long`）。默认命令下这两个用例会
> **随机器负载时好时坏** —— 同一份代码有时 530 s 通过、有时 660 s 被杀。这种假失败比不跑更糟，
> 所以默认跳过（会打印跳过原因与开启方法）。测试二进制内部的 `-test.timeout` 拦不住 go 命令的
> 杀进程，改它没有意义。

### 9.1 已实测的验收数据

| 层 | 结果 | 复现命令 |
|---|---|---|
| 零改动 roundtrip | **165/165 个真实包**逐条目 sha256 一致 | `go test ./internal/pkgfile` |
| zip 容器保真 | **165/165 个真实包的零改动重建与原包逐字节相同**；原始解析与 `archive/zip` 在 660 个条目上逐字段一致；Zip64 包改动后仍可被独立读者读全 | `go test ./internal/pkgfile` |
| TAP 字节保真 | 32,571 个 `<point>` / 3,641 个 `<section>`；36,212 个带 CDATA 元素与独立正则对拍一致（36,208 通过、4 个重名点）；**36,373 个恒等写操作全部逐字节一致** | `go test ./internal/tapfile -count=1` |
| 围栏可逆性 | **165 个包、73,107 个 Region**，`parse_fenced(render(synthesize(pkg))) ≡ synthesize(pkg)` 逐项相等且区间记账对称 | `go test ./internal/fence` |
| FGL 解析器 | BDL 61 组夹具「通过 58 / 已知偏差 3 / 不符 0」；语料 4,318 个自订点 4,317 个通过（唯一失败是退化点） | `go test ./internal/fgl -count=1` |
| `verify` 语料 | **165 个包全部 0 error**（125 warn / 302 info） | `TDEV_DEEP=1 go test ./internal/cli -run TestCorpusExportVerify` |
| `apply` 仿真 | **137 个包**实际写回：只动目标点，`.4gl`/`ver` 与其它点字节不变 | `TDEV_DEEP=1 go test ./internal/cli -run TestCorpusApplySimulation` |
| `.tzs` 语料回归 | **67 个包 0 跳过 0 失败**；pin 67 条与磁盘一致；1285 帧请求、0 行异物输出 | 见上 |
| 对抗用例 | **31 项自检全绿**（改围栏行/改只读区/删 end 围栏/塞 `]]>`/结构行/围栏外插行/`--only` 范围外/缺条目/ver 不匹配/区段不配对/改名事务 V1–V7/newfn 模板/newfn 无 `--desc`…），且**被拒的 apply 不改包** | `tt dev tzc selftest` |
| 报错定位 | 只读区被改 → 打印 `prog.full.4gl:<行号>` + 该行内容（`--json` 带 `file`/`line`/`snippet`）；行尾归一化不会把位置指到区段开头 | `go test ./internal/model ./internal/cli -run 'LineAt\|FirstDiffEOL\|TestApplyReportsReadonlyLine'` |
| 安装面 | `install skills` 复制到 `<当前目录>/skills`（冲突拒绝 / `--force` 刷新 / 源=目标拒绝）；`install path` 解析真实折行的用户 PATH、幂等、不改写 `%USERPROFILE%` | `go test ./internal/cli -run 'TestInstall\|TestMergeUserPath'` |

> 前 4 行随默认 `go test ./...` 一起跑；`verify` 语料 / `apply` 仿真属于深度回归。

### 9.2 真机验收清单（S7，需人工执行）

自动化只到 gate3（「模拟设计器装载」）。最后一步请在装了客户端的机器上做：

1. 取一个 `apply` 产出的包（`apply` 报告会打印路径与新 sha256）。
2. **先关掉设计器里该程序的文档**（否则设计器内存里的旧模型会在保存时覆盖你的改动）。
3. 在设计器里打开该 `.tzc`：应**无异常对话框**；改动内容可见。
4. 保存并关闭，再重新打开一次确认改动仍在。
5. 建议留证：改动前后的 `.tap` 的 `<point>` CDATA 逐字节比对。

> **输入必须是设计器产出的包。** 如果输入本身就是早期版本产出的（局部头 bit 3 + 数据描述符，
> 见 §6.3），那它本来就不是设计器形态；tt dev 会把它归一化再写，但**用这种包做真机验收等于
> 同时验两件事**，容易误判。验收请从原生的 `.tzc` 出发。

---

## 10. 构建与发布

**操作步骤见 [README 的构建一节](../README.md#构建)**。这里只留"为什么"。

- **前端必须先构建，但 `go build` 不会因此失败**：`main.go` 用 `//go:embed all:web/dist`，
  而 `web/dist/.gitkeep` 这个占位文件让 embed 在产物缺失时也成立（否则全新克隆连 `go build`
  都过不去）。代价是那样出来的 tt 没有界面 —— `tt serve` 会返回一张写着「界面未构建」和构建
  命令的说明页，不是静默空白。
- **`.tzs` 引擎单独构建**，因为它在 Go 的构建链之外，而且重编会让在跑的守护进程变孤儿（§7.1）。
- **打包脚本只采集、不构建**：`build_portable.bat` 按名字采引擎那四个文件与
  `engine\designer\`（`engine/out/` 里还有十几个探测程序，xcopy 整个目录会把它们一起打进包里）。
  MSI 复用同一份载荷，用 `heat.exe` 采集文件（新增目录自动跟着走，不用改 `tt.wxs`）。
- **`tools/zip.py` 为什么单独一个脚本**：原先是在 `.bat` 里塞一句 `python -c`，用的是
  `os.listdir` + `z.write` —— 那个组合**不会递归**，目录只会写进一个空条目：`skills/` 下明明有
  `SKILL.md`，打出来的包里却只有一个空的 `skills/` 目录，便携包于是缺了技能文档（而它正是给 AI
  用的说明书）。
- **MSI 是用户级安装**（装到 `%LOCALAPPDATA%\Programs\TT`、只追加 HKCU 的用户 PATH、免管理员），
  刻意**不带** `.portable` 与 `config.json`：装出来的版本配置该落在 `%APPDATA%\T100\tt\`，
  不是安装目录。文件清单由 `heat.exe` 采集，卸载要删的目录由 `tools/wix_removefolders.py` 补
  （MSI 的 ICE64/ICE38 对用户级安装的要求）。需要 WiX v3 工具集，`WIX_BIN` 指向它。
- **便携包与 MSI 都不含**：本机的 `config.json`（含真实口令）、`erp_data.db`（含客户表字典/
  schema/企业码）、设计器里的 GUI 与自动更新组件。

---

## 11. 设计史与取舍

> 本章是"为什么现在是这个样子"的记录。上面各章写的是**现状**，这里写的是**经过**。

### 11.1 合并动机：三处"请手工同步"

三个项目互相依赖，靠注释提醒人工保持同步：

| 位置 | 原文 |
|---|---|
| `TDebug/cli/root.go:53` | 存放规则(与 TDictCli 保持一致，改动请两边同步) |
| `TDictCli/cli/root.go:173` | 存放规则(与 TDebug 保持一致，改动请两边同步) |
| `TDebug/cli/install.go:6` | 命令形态与 TDictCli / TDev 一致(改动请三边同步) |

这三处现在各只有一份实现。另外两份 `cfgfile`、两份 `dbconfig`、两份 `host`、两份 `pathinstall`、
两份 `erpdb` 也合并了。合并**不做功能取舍，只消除重复实现**。

### 11.2 合并中发现的分叉

两个 `host` 包**不是**简单的复制关系，已经实质分叉。逐文件比对后，**每个有分叉的文件都取更
成熟、更安全的那一版**：

| 文件 | 发现 | 取哪版 |
|---|---|---|
| `ssh.go` | TDebug 版修掉一处数据竞争（把结果写进外层变量，超时分支返回后 goroutine 仍在写，`-race` 可复现），改用带缓冲 channel；并新增 `OutputStdin` | TDebug |
| **`dbprobe.go`** | **这次合并最重要的发现** —— 见下 | TDebug |
| `tenv.go` | TDictCli 版用变量名前缀解析探针回显，但 PTY 只有纯文本，登录脚本自己也会打印 `ZONE = t35prd` 这类行，前缀法无法区分"我方请求的值"与"服务器自己打印的内容"。TDebug 版用 `TDBG-BEGIN`/`TDBG-END` 定界符区间 + 状态机，超宽换行产生的碎片也落在区间之外 | TDebug |
| `mirror.go` | TDebug 没有这个文件（源码镜像只服务于字典查询） | 从 TDictCli 搬入 |
| `dbconfig` | TDictCli 版少了 `ReadonlySQL *bool` 与 `ReadonlySQLEnabled()` | TDebug（超集） |
| `erpdb` / `pathinstall` | TDictCli 版是超集（`erpdb` 多了 `SelectAllSQL`/`ValidIdent`/`QuoteIdent`/`QuoteLit`；`pathinstall` 已是完整包而 TDebug 版是两个散文件） | TDictCli |
| `cfgfile` | 两份逐字节相同，只差一个节访问器（`Debug(root)` vs `Hosts(root)`） | 重写为 `internal/config` |

**`dbprobe.go` 的命令注入面。** TDictCli 版的 SQL 执行是**完整的命令注入面**：

```go
// 旧版（TDictCli）
func SqlplusRun(zone, sqlplusPath, connStr, sql string) string {
    q := strings.ReplaceAll(sql, "'", "'\\''")          // 只转义单引号
    return fmt.Sprintf(`bash -lc '%s; echo "%s" | %s -S %s'`, ChenvCmd(zone), q, p, connStr)
}
```

SQL 拼进命令串后再整体套一层 `bash -lc`，**两层解析**：数据里的 `$(...)`、反引号、双引号都会
在远端 shell 里展开，一个引号就能把后面的 `;` `|` 变成命令分隔符。而 `tdict` 执行 SQL 的输入
来自用户查询与 AI 生成的语句，**不是常量**。

TDebug 版换成了「返回不含 SQL 的命令行 + SQL 经 stdin 送入」：

```go
func SqlplusCmd(zone, sqlplusPath, connStr string, killAfterSec int) (string, error)
func KbCmd(ksqlPath, host, port, db, connStr string) (string, error)
// 调用方：conn.OutputStdin(cmd, []byte(sql), timeout)
```

并补了 `reToolPath` / `reDBAcct` 白名单校验（工具路径、数据库账号），拒绝含换行的口令；
`killAfterSec > 0` 时给远端命令套 `timeout -s TERM` —— **Oracle 侧没有任何服务端超时机制**，
本地 SSH 层超时只是"不再等"、不杀进程，这一步是必须的。

**TDictCli 那套不安全的 SQL 拼接没有保留。** `tt dict` 的调用点相应改写，这是合并里唯一需要
改动调用方行为的地方。

### 11.3 统一配置的结构变更

两份 `resolveConfigPath` / `toolsHome` / `isPortable` / `migrateLegacyConfig` / `looksLikeOwnConfig`
/ `samePath` / `defaultConfigPath` 逐行相同，只有 `toolDirName`（`tdebug`/`tdict`）与环境变量名
不同。合并为 `internal/config/paths.go` 一份：

| | 合并前 | 合并后 |
|---|---|---|
| 工具目录 | `%APPDATA%\T100\tdebug`、`%APPDATA%\T100\tdict` | `%APPDATA%\T100\tt` |
| 显式指定 | `TDEBUG_CONFIG` / `TDICT_CONFIG` | `TT_CONFIG`（旧的两个名仍识别） |
| 便携标记 | `.portable` | 不变 |
| 覆盖根目录 | `T100_HOME` | 不变 |

迁移规则与理由见 §2.3。另外两处顺带修好的：

- `Open` 的注释写"空文件视为空配置"，实现却对空文件抛 `unexpected end of JSON input` ——
  现在按文档行为处理。
- `Save` 从「`path + ".tmp"` + rename」换成 TDev 那套更完整的原子写（同目录 `CreateTemp` +
  `Write` + `Sync` + 保留原权限位 + `Rename`）。

### 11.4 命令面与兼容

| 合并前 | 合并后 |
|---|---|
| `tdebug start` | `tt debug start` |
| `tdev tzc export` | `tt dev tzc export` |
| `tdict r.t` | `tt dict r.t` |
| `tdebug serve` / `tdict serve` | `tt serve`（一个服务，一套页面） |
| 三份 `install skills` | `tt install skills` |
| `tdebug env` + `tdebug topent` + `tdict env` | `tt env list/show/use/topent` |
| — | `tt config path/show/get/set/migrate/validate` |

旧命令名保留为别名（`tt tdebug …` = `tt debug …`），既有脚本与 AI skill 不改即可运行。
`--conn` 保留为 `--env` 的别名。

TDev 的两点被完整保留：位置无关的参数解析（`-o`/`--json` 可出现在位置参数之后，标准库 `flag`
做不到），以及明确的退出码契约。`tt dev` 用 `DisableFlagParsing` 把参数原样交给 TDev 自己的
解析器，退出码经 `exitCode` 透传。

### 11.5 合并过程中修掉的两个 bug

两处都是"整节替换"这个动作的副作用，都由测试抓出来：

1. **类型化 nil 指针装进 `interface{}` 不等于 nil。** `PUT /api/hosts` 原本用
   `for key, v := range map[string]any{"query": req.Query, …}` 加 `if v == nil { continue }`
   判断"这一节没提交"。但那些字段是指针类型，nil 指针装进 `any` 后接口值非 nil，于是**本次没提交
   的节会被序列化成 `null` 再当成空对象写回去，把用户原有的 `query`/`mirror`/`debug` 全清掉**。
   现在逐个类型化判空。
2. **`viaSsh` 被配置页静默丢弃**（见 §4.3）。

两个 bug 都有回归测试盯着：`TestHostsPut_LeavesOtherSectionsAlone` 与
`TestHostsPut_PreservesViaSSH`。

### 11.6 前端统一设置页（5 期重构）

三工具合并完成后，前端仍是两套 SPA（`web/debug` 与 `web/dict`），各带一个设置页，两边的环境
编辑器近乎重复。要求是**只保留一个设置页签**，组织成站点管理 / 数据字典 / DEBUG / 应用设置四块。

| 期 | 做了什么 |
|---|---|
| 0 | 抽出共享层 `web/shared/`：主题变量与机制、UI 基元、设置页布局件 |
| 1 | 设置页外壳 + 深链接（`/debug/#settings/<分区>`），并**修掉一处数据丢失 bug**（见 §11.5 第 1 条） |
| 2 | 设置页只跟共享层说话；服务端补上"保留表单不管理的字段"（`viaSsh`、每环境的 `launchArgs`/`watchdogSeconds`） |
| 3 | 字典页接入共享主题层 |
| 4 | **彻底移除字典页**：设置统一之后字典页只剩两个动作，于是那个 SPA 被删掉，动作并进设置页 |
| 5 | 目录改名（清理） |

`grep` 确认 `web/app/src` 无写死色板；产物 CSS 里 `.dark` 是类驱动、`prefers-color-scheme` 归零、
明暗两套 `--background` 都在。

### 11.7 后续变更

- **移除桌面版（Electron 外壳）**：`desktop/` 与 `build_desktop.bat` 整体删除；`tt serve --desktop`
  隐藏开关、`TT_READY {json}` 就绪行、`TT_DESKTOP_DATA`/`TT_DESKTOP_PORT`/`ELECTRON_MIRROR` 三个
  环境变量一并去掉。界面统一走浏览器。
  - 桌面目录里唯一还被别处引用的文件是 MSI 的图标，已搬到 `installer/icon.ico`。
  - `paths.go` 里"桌面版旧数据目录"的兜底路径**保留**：那是给曾装过桌面版的机器读旧配置用的，
    删掉会让它们的配置突然读不到。
  - `POST /api/shutdown` 保留：它原本的主要调用方是桌面壳的关窗动作，现在没有界面在调，但它仍是
    这个服务对外的优雅停止入口（也可 curl），且带测试，删了只有净损失。
- **`tt serve` 改为默认后台常驻**：与 `tt debug serve` 共用同一套 spawn / 状态文件 / 停止机制。
  随之修掉两个既有缺陷：两种服务的调试 API 前缀不同而控制端永远拼 `/api/…`（于是对着 `tt serve`
  全是 404，现在按状态文件的 `apiBase` 寻址）；以及 `root.go` 的 `Execute` 注释写着"错误只由这里
  打一次"但那里并没有打印，于是任何命令失败都是静默退出 1 —— 已补上 stderr 输出。
- **`.tzs` 从"不写回"变成"走引擎"**：v1 的红线是"永远不写回"，理由是"表单由设计器的表单设计器
  维护，tdev 没有对应的模型与验收样本"。那个前提在 `engine/` 落地后不成立 —— 引擎就是设计器
  自己的代码。红线改成"导出只读、要写走引擎"。
- **设计器程序集入库**：原先由用户在配置里指路（`tzs.installDir`），现改为随仓库与发行包分发，
  见 §7.1。

### 11.8 未做 / 已知取舍

- **`viaSsh` 无法通过配置页删除。** 为了不"保存一次就静默丢配置"，服务端保存时会把表单不管理的
  `db.viaSsh` 从旧配置补回来。要删就手改 `config.json`。
- **`tt config set` 不做节内校验。** 深路径写入会绕过环境页的校验（环境名唯一、端口范围等）。
  改环境清单请用 `tt env` 或配置页。
- **`tt dev` 的 config 默认值只覆盖 `-o` 省略的场景。** 新增的 `tdev` 节刻意只放两个稳定的默认值。
- **`tt --config` 对 `tt dev` 需要额外一步**：该命令组是 `DisableFlagParsing`，cobra 不会替它解析
  全局 flag，所以 `--config` 会被摘出来转成 `TT_CONFIG` 再转发。三种写法都已验证可用。
- **`internal/dev` 的 `store.ToolName` 仍写作 `"tdev tzc"`**：那是写进 `manifest.json` 的持久化
  字段，没有任何代码读它；保持稳定意味着重命名二进制不会让已有工作区的 manifest 变身份。
- **`.tzs` 引擎的语料回归没有全函数面覆盖**：`engine/gate-w3-fns.py` 驱动全部 34 个写函数跑在
  一个长驻进程里（只有这样才能看见 `ComponentTabIndexService` 的状态泄漏），每组带负向对照；
  Go 侧测试驱动 8–10 个函数，是**语料的回归网，不是函数面的覆盖网**。它没有 Go 等价物，
  所以保留在 `engine/` 里。`gate-w3.py` 的 B-neg（"不写直接存"与"被拒的写之后存"必须逐字节
  相同）也还没有搬过来。
- **在线/离线**：本项目只做离线（本地镜像、本地 SQLite）。在线那一层不在范围里。
- **设置页没有搜索框**（用户明确说不做）；**每环境的 `launchArgs`/`watchdogSeconds` 不在界面上
  暴露**，靠服务端保留机制保证不被丢掉。

---

## 12. 许可与出处

- **本工具的实现依据是 T100 设计器**（第三方商业软件，厂商标识 **DSC**）——`.tzc` 一侧来自其
  发行物的反推，`.tzs` 一侧直接反射调用其程序集。仅供个人学习、排障与接口对接研究；
  **请勿用于重制发布或绕过授权**。
- `docs/T100设计器-README.md` 是该设计器反编译源码树 README 的**逐行恢复副本**（第三方文本），
  是 §6.14 各条 `file:line` 依据的来源。
- **设计器的 28 个程序集**（`engine/designer/`，第三方商业软件）经明确决定随本仓库分发，
  见 §7.1。
- `testdata/fgl-fixtures` 与 `internal/fgl/outline.go` 移植自同作者的 MIT 项目 **BDL**
  （`D:\我的项目\BDL`）。
- Go 依赖见 [README 的依赖一节](../README.md#依赖)；前端依赖见 `web/app/package.json`。



