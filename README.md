# TT — T100 工具集

TDebug、TDev、TDictCli 三个工具合并后的统一入口。

**一个仓库、一个二进制 `tt`、一份配置、一个本地 Web 服务。**

```
tt debug …   作业调试器：SSH 驱动 fglrun -d 的 (fgldb) 文本调试协议
             本地 Web 界面（源码/断点/调用栈/变量/接口日志）+ 命令行控制端
tt dev …     T100 设计器包工具：.tzc 安全编辑（export/status/verify/apply），.tzs 只读解压
tt dict …    ERP 数据字典查询：r.t / r.v / desc / scc / r.q / prog，本地镜像与远程直查

tt env …     环境管理（三个工具共用同一份环境清单）
tt config …  配置管理：位置 / 查看 / 读写 / 迁移 / 校验
tt serve     启动本地 Web 服务（工作台 + 统一设置页）
tt install   安装 AI skills 到当前目录，或把 tt 加进用户 PATH
tt version
```

## 为什么合并

三个工具本来就是同一套东西的三块，且已经互相依赖：

- 都要 SSH 连同一批 T100 服务器、都读同一批环境与数据库连接；
- `TDebug/cfgfile/cfgfile.go` 与 `TDictCli/cfgfile/cfgfile.go` 近乎逐字节相同；
- 两边的配置路径解析（`toolsHome` / `isPortable` / 旧配置迁移）是同一份实现抄了两遍，
  注释里直接写着**「与对方保持一致，改动请两边同步」**；
- `install` 命令的注释写着**「与另外两个一致（改动请三边同步）」**。

三处"请手工同步"就是合并的理由。现在这些各只有一份实现。

合并过程中发现原两份 `host` 包已实质分叉，且 TDebug 那份在每个文件上都更成熟：

| 文件 | 差异 |
|---|---|
| `ssh.go` | TDebug 用 channel 收结果，修掉了超时分支后 goroutine 仍写外层变量的数据竞争；并新增 `OutputStdin` |
| `dbprobe.go` | TDebug 对工具路径/账号做白名单校验，SQL 走 stdin 而非拼进命令行 —— 旧版是**完整的命令注入面**（`echo "<SQL>" \| sqlplus` 再套 `bash -lc`，两层解析）；并用 `timeout -s TERM` 真正杀掉 Oracle 会话 |
| `tenv.go` | TDebug 用定界符区间解析探针回显，挡掉登录脚本自己打印的 `ZONE = t35prd` 之类污染行 |

所以合并以 TDebug 版为基，补入 TDictCli 独有的 `mirror.go`（源码镜像引擎）。
旧版那条不安全的 SQL 路径没有保留。

## 安装

### 便携包

把 `tt-portable.zip` 解压到任意目录，双击或命令行运行 `tt.exe` 即可。
包内有 `.portable` 标记，配置就近留在包内（`config.json`），不写用户目录。

### 加进 PATH

```
tt install path          # 把 tt.exe 所在目录加进用户 PATH（HKCU，免管理员）
tt install path --dry-run   # 只预览将要写入的内容
```

### AI 技能

```
tt install skills                    # 复制到 <当前目录>/skills
tt install skills --to .claude/skills   # 装到 Claude Code 直接读的位置
```

合并后 `skills/` 下是四套：`tdebug-debug`（调试）、`tdev`（设计器包）、
`tdict` 与 `erp-code-reader`（数据字典）。装一次全部到位。

## 配置

**只有一个配置文件**：`%APPDATA%\T100\tt\config.json`，三个工具共用。

位置解析顺序（第一个存在的胜出）：

1. `TT_CONFIG` 环境变量（旧名 `TDEBUG_CONFIG` / `TDICT_CONFIG` 仍可用）
2. `--config <路径>`
3. `<exe 目录>\.portable` 存在 → 便携包，配置留在包内
4. `%APPDATA%\T100\tt\config.json`
5. 旧位置兜底（首次运行自动合并迁移到 4）

`T100_HOME` 环境变量整体改写统一目录（如 `T100_HOME=D:\t100`）。
数据目录 = 配置所在目录，所以 `srccache/`、`debug-bps/`、`logs/` 都跟着落在一起。

### 结构

```jsonc
{
  "schemaVersion": 2,
  "listen": "127.0.0.1:28670",     // 本地 Web 服务；端口占用时自动顺延

  "hosts": {                        // ★ 共用环境清单：三个工具都读这一份
    "activeEnv": "开发环境",
    "sshs": [{
      "name": "开发环境", "host": "10.0.0.1", "port": 22,
      "user": "tiptop", "password": "…", "zone": "31", "topent": "10001",
      "db": { "type": "oracle", "host": "10.0.0.2", "port": 1521,
              "service": "t100dev", "readonlySql": true,
              "accounts": [{ "account": "ds", "password": "…" }] }
    }]
  },

  "debug": { "activeEnv": "", "launchArgs": "…", "watchdogSeconds": 180,
             "fglserver": "", "termWidth": 200, "termHeight": 50,
             "printElements": 1000, "persistBreakpoints": true },
  "query":  { "source": "auto" },   // auto（缺省=在线优先）| local | <环境名>
  "mirror": { "dir": "" },
  "bdldoc": { "dir": "" },
  "sync":   { "target": "" },
  "tdev":   { "workspaceSuffix": "-ws", "defaultOut": "" }
}
```

**这就是"统一配置管理"的核心**：原来 TDebug 把环境放在 `debug.sshs`、TDictCli 放在
`hosts.sshs`（两份 schema 完全一样，只是节名不同），现在只留 `hosts.sshs` 一处。
`debug` 节降级为纯工具设置。`tdev` 是新增节 —— 原 TDev 完全没有配置系统，
所有参数都是每次调用的 flag；新节只放跨调用稳定的默认值，**flag 仍然优先**。

完整示例见 `config.example.json`。

### 旧配置迁移

首次运行会**自动**把合并前的两份配置合并成一份，规则：

- 环境清单取**并集**，同名环境以 TDictCli 那份为准（它是字典查询的现役配置）；
- TDebug 的 `debug` 节被**拆开**：`sshs` 提升为 `hosts.sshs`，其余键留在 `debug`；
- `query` / `mirror` / `bdldoc` / `sync` 原样带过来；
- **原文件不删除**，各留一份 `.pre-merge.bak`。

> 注意 TDictCli 原有的一条迁移规则是「`debug` 整节重命名为 `hosts`」。那条规则在合并后会
> 把 TDebug 的 `launchArgs` / `watchdogSeconds` / `termWidth` 等设置一并吞进 `hosts`，
> 必须换成本次的拆解规则 —— 这是合并里最容易出错的一处，`internal/config/migrate.go`
> 的注释和测试都盯着它。

预览迁移结果：

```bash
tt config migrate --dry-run
```

### 配置页

```
tt serve
```

一个进程、一个端口、一套页面：

- `/debug/` — 调试工作台
- `/debug/#settings` — 统一设置页：站点管理 / 数据字典 / DEBUG / 应用设置
- `/api/*` — 统一接口：环境与数据库配置（`/api/hosts`）、配置派生状态、
  PATH 安装，以及字典类动作（源码镜像拉取、字典同步、BDL 文档）

**三个工具的配置都在设置页里**：环境与数据库、查询数据源、镜像目录、同步目标、
BDL 文档目录、调试参数、明暗色。环境清单只有一份数据源（config.json 的 hosts 节），
调试、字典查询与源码镜像读的是同一份。

### 命令行也能改配置

```bash
tt env list                       # 列出全部环境（* 标记默认环境）
tt env show 开发环境               # 看某个环境的连接详情（口令打码）
tt env use 正式环境                # 切换默认环境
tt env topent 开发环境 10001       # 设置默认企业编号

tt config path                    # 配置在哪
tt config show                    # 看配置（口令打码；--raw 看原样）
tt config validate                # 校验：节是否齐全、环境是否完整
tt config get hosts.activeEnv     # 按点分路径读
tt config set debug.termWidth 240 # 按点分路径写（值按 JSON 解析）
tt config migrate [--dry-run]     # 合并旧配置
```

写操作走与配置页完全相同的原子写路径（同目录临时文件 + fsync + 重命名），
失败不留半截文件，也不会覆盖自己那一节之外的任何键。

## 命令面

### `tt debug` — 作业调试器

```
tt debug serve [--stop|--foreground]     启动/停止调试服务
tt debug probe                           探测：SSH / 区域 / 数据库
tt debug db status|discover|ping         数据库连接检查

tt debug start <作业> [-m <模块>] [-z <区域>]   拉起调试会话
tt debug exec "<命令>"                    下断点/继续/求值（fgldb 原生命令）
tt debug status | quit | wait | why       会话管理
tt debug source | logs | locate | resolve | interrupt
tt debug wslogs | wsdebug | sql | wstest  接口日志与报文回放
tt debug mode | topent                    切换调试模式 / 企业
```

### `tt dev` — 设计器包工具

```
tt dev tzc export <pkg.tzc> [-o <dir>] [--only <点名>...] [--json]
tt dev tzc status | verify | apply | unlock | rename | newfn | selftest
tt dev tzs export <pkg.tzs> [-o <dir>] [--force] [--json]
```

`.tzc` 是**代码包**，走 `export` 渲染围栏工作区、改完 `apply` 写回（唯一写路径）；
`.tzs` 是**表单包**，走 `export` 纯解压、只读参考（没有 `tzs apply`）。
退出码：`0` 成功 / `2` 包格式或用法错 / `3` 校验失败 / `4` 拒绝写入 / `5` IO 与环境失败。

### `tt dict` — 数据字典

```
tt dict r.t [表名] | r.v [识别码] | desc <表> [字段] | scc [分类码] | r.q [开窗码]
tt dict msg | sysp | docp | prog
tt dict env
tt dict db status | sync | list | ping | discover
tt dict mirror dir|pull|path
tt dict bdldoc dir
```

数据源缺省是**在线优先**：用默认环境（`hosts.activeEnv`）的远程库直查；
一个环境都没配时才用本地 SQLite 镜像（`erp_data.db`，由 `tt dict db sync` 同步）。
用 `--env <环境名|local>` 或 `config.json` 的 `query.source` 固定数据源。
连不上不会静默回落本地 —— 否则你会以为查到的是实时数据，实际是几天前的镜像。

## 项目结构

```
TT/
├─ main.go                  //go:embed all:web/dist → cli.Execute
├─ internal/
│  ├─ config/               ★ 统一配置层：位置解析 / 读写 / schema / 迁移
│  ├─ cli/                  cobra 根命令
│  │  ├─ env.go config.go serve.go install.go version.go
│  │  ├─ common/            三个命令组共享的 CLI 上下文（叶子包）
│  │  ├─ debug/ dev/ dict/  三个命令组
│  ├─ debug/                调试内核：fgldb 驱动、会话、REST+WS
│  ├─ dev/                  设计器包管线：pkgfile/tapfile/tglfile/fgl/synth/fence/verify/split/store/model
│  ├─ dict/                 数据源与查询：db（本地）/ live（远程）/ dbsync / server
│  ├─ web/                  统一 HTTP 服务与共享 /api/hosts
│  ├─ host/                 ★ 合并后唯一的 SSH/PTY/终端/探测/镜像层
│  ├─ dbconfig/ erpdb/      数据库连接模型与连接器（两处合并）
│  ├─ safesql/ sshtun/ output/ pathinstall/ atomic/
├─ web/
│  ├─ app/                  调试工作台 SPA（React + Radix + zustand + Monaco）
│  │                        其中的设置视图是三个工具的统一配置页
│  ├─ shared/               整套 SPA 共用的主题层 / UI 基元 / 设置页布局件
│  └─ package.json          npm workspace 根
├─ desktop/                 Electron 外壳
├─ skills/                  AI 技能：tdebug-debug / tdev / tdict / erp-code-reader
├─ docs/                    分册文档与设计说明
└─ testdata/                FGL 夹具
```

`★` = 为合并而真正重组的部分。

## 构建

```bash
# 前端（必须先构建：main.go 的 //go:embed all:web/dist 依赖产物）
cd web && npm install && npm run build

# 后端
cd .. && go build -o tt.exe .

# 便携包
build_portable.bat        # → dist/tt-portable.zip
```

开发模式：

```bash
tt serve                  # 后端（默认 127.0.0.1:28670）
cd web && npm run dev:debug   # 工作台与设置页，热更新，代理到后端
```

端口被占用时后端会自动顺延，用 `TT_PROXY=http://127.0.0.1:<实际端口>` 告诉前端。

## 文档

| 文档 | 内容 |
|---|---|
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | 合并架构与统一配置设计 |
| [docs/MIGRATION.md](docs/MIGRATION.md) | 迁移记录：改了什么、为什么 |
| [docs/debug.md](docs/debug.md) | 调试器手册（原 TDebug README） |
| [docs/dev.md](docs/dev.md) | 设计器包工具手册（原 TDev README） |
| [docs/dict.md](docs/dict.md) | 数据字典手册（原 TDictCli README） |
| [docs/tzc-model.md](docs/tzc-model.md) | `.tzc` 包模型与不变量 |
| [desktop/README.md](desktop/README.md) | Electron 桌面外壳 |

## 兼容性

合并前的命令名仍可直接当子命令组用，既有脚本与 AI skill 不改即可运行：

```
tt tdebug …  =  tt debug …
tt tdev …    =  tt dev …
tt tdict …   =  tt dict …
```

环境变量同理：`TDEBUG_CONFIG` / `TDICT_CONFIG` 仍被识别，`--conn` 仍是 `--env` 的别名。
