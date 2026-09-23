# TT — T100 工具集

T100 日常工作的统一入口：**一个二进制 `tt`、一份配置、一个本地 Web 服务。**

```
tt debug …     作业调试器：SSH 驱动 fglrun -d 的 (fgldb) 文本调试协议，
               本地 Web 界面（源码/断点/调用栈/变量/接口日志）+ 命令行控制端
tt dev tzc …   设计器代码包（.tzc/.tzf/.tzx）：渲染成带围栏的 4GL 工作区给人/AI 改，
               改完走三道闸门写回
tt dev tzs …   设计器表单包（.tzs/.tzv）：export 纯解压只读；读写表单走 call，
               由设计器自己的引擎驱动（见「.tzs 表单引擎」一节）
tt dict …      ERP 数据字典查询：r.t / r.v / desc / scc / r.q / prog，本地镜像与远程直查

tt env …       环境管理（各命令组共用同一份环境清单）
tt config …    配置管理：位置 / 查看 / 读写 / 迁移 / 校验
tt serve       启动本地 Web 服务（工作台 + 统一设置页；默认后台常驻，打印地址后返回）
tt serve --foreground / --stop    前台运行看实时日志 / 停止后台实例
tt install     安装 AI skills 到当前目录，或把 tt 加进用户 PATH
tt version
```

TT 最初是 TDebug、TDev、TDictCli 三个工具合并的结果 —— 合并的动机、发现的分叉与取舍
（包括旧版那条**命令注入面**为什么没有保留）记在 [docs/MIGRATION.md](docs/MIGRATION.md)。
合并之后又长出了第四块：**`.tzs` 表单引擎**，它不在 Go 的构建链里，而且**运行期要的设计器程序集
由发行包自带**（不靠用户机器上恰好装着的那份），所以单独用一节说清楚。

有一条边界贯穿全篇，先说在前面：**tt 不实现、也不分发 T100 设计器的任何私有格式**。
`.tzc` 靠设计器的公开发行物反推（[docs/tzc-model.md](docs/tzc-model.md)）；
`.tzs` 更彻底 —— 它**直接反射调用已安装的设计器自己的程序集**，格式那部分是设计器自己在算。

## 安装

### 便携包

把 `tt-portable.zip` 解压到任意目录，双击或命令行运行 `tt.exe` 即可。
包内有 `.portable` 标记，配置就近留在包内（`config.json`），不写用户目录。

包内还带 `tzs\`（引擎四个文件 + `designer\` 里的设计器程序集，见下）与五套 AI 技能。
引擎跑起来**不需要你装设计器、也不需要配它的路径** —— 包里自带的那份就是它用的那份，
所以同一份包在任何机器上跑的是同一版设计器。

### MSI 安装包（用户级，免管理员）

`TT-0.1.0-x64.msi` 双击即可安装，**全程不需要管理员**：

- 装到 `%LOCALAPPDATA%\Programs\TT`，不碰 `Program Files`；
- 安装目录追加到**用户** PATH（HKCU），卸载时自动摘掉；
- 刻意**不带** `.portable`，所以配置落在 `%APPDATA%\T100\tt\config.json`
  而不是安装目录（程序目录不是放用户数据的地方）；
- 也不带 `config.json`——装完是干净的一份，首次运行自己生成骨架。

升级：装了新版 MSI 会先移除旧版（固定 UpgradeCode + 每次构建新 ProductCode）。

### 加进 PATH

```
tt install path             # 把 tt.exe 所在目录加进用户 PATH（HKCU，免管理员）
tt install path --dry-run   # 只预览将要写入的内容
```

### AI 技能

```
tt install skills                       # 复制到 <当前目录>/skills
tt install skills --to .claude/skills   # 装到 Claude Code 直接读的位置
```

`skills/` 下是五套：`tt-debug`（调试）、`tt-dev-tzc`（代码包 `.tzc`）、`tt-dev-tzs`（表单包 `.tzs`）、
`tt-dict`（数据字典）、`erp-read`（读 ERP 代码）。一个技能一个目录、目录里是 `SKILL.md`
（Claude 技能规范：目录名必须等于 frontmatter 里的 `name`）。装一次全部到位。

## 配置

**只有一个配置文件**：`%APPDATA%\T100\tt\config.json`，所有命令组共用。

位置解析顺序（第一个存在的胜出）：

1. `TT_CONFIG` 环境变量（旧名 `TDEBUG_CONFIG` / `TDICT_CONFIG` 仍可用）
2. `--config <路径>`
3. `<exe 目录>\.portable` 存在 → 便携包，配置留在包内
4. `%APPDATA%\T100\tt\config.json`
5. 旧位置兜底（首次运行自动合并迁移到 4）

`T100_HOME` 环境变量整体改写统一目录（如 `T100_HOME=D:\t100`）。
数据目录 = 配置所在目录，所以 `srccache/`、`debug-bps/`、`logs/`、`.tt-tzs.json` 都跟着落在一起。

### 结构

```jsonc
{
  "schemaVersion": 2,
  "listen": "127.0.0.1:28670",     // 本地 Web 服务；端口占用时自动顺延

  "hosts": {                        // ★ 共用环境清单：所有命令组都读这一份
    "activeEnv": "开发环境",
    "sshs": [{
      "name": "开发环境", "host": "10.0.0.1", "port": 22,
      "user": "tiptop", "password": "…", "zone": "31", "topent": "10001",
      "db": { "type": "oracle", "host": "10.0.0.2", "port": 1521,
              "service": "t100dev", "readonlySql": true,
              "accounts": [{ "account": "ds", "password": "…" }] }
    }]
  },

  "debug": { "activeEnv": "", "launchArgs": "…", "watchdogSeconds": 1800,
             "fglserver": "", "termWidth": 200, "termHeight": 50,
             "printElements": 1000, "persistBreakpoints": true },
  "query":  { "source": "auto" },   // auto（缺省=在线优先）| local | <环境名>
  "mirror": { "dir": "" },
  "bdldoc": { "dir": "" },
  "sync":   { "target": "" },
  "tdev":   { "workspaceSuffix": "-ws", "defaultOut": "" },

  "tzs": {                          // .tzs 引擎（设计器程序集随包自带，不在这里配）
    "workspace": "D:\\t100_wrok_dir\\某客户\\prd"
  }
}
```

**这就是"统一配置管理"的核心**：原来 TDebug 把环境放在 `debug.sshs`、TDictCli 放在
`hosts.sshs`（两份 schema 完全一样，只是节名不同），现在只留 `hosts.sshs` 一处。
`debug` 节降级为纯工具设置。`tdev` / `tzs` 是新增节 —— 原 TDev 完全没有配置系统，
所有参数都是每次调用的 flag。

`tdev` 与 `tzs` 的口径**故意不同**，值得分清：

- `tdev` 放的是**跨调用稳定的默认值**，命令行 flag 永远优先；
- `tzs` 只剩**工作区**这一个不能由 flag 取代的机器级依赖，而且它没有合理缺省 ——
  引擎内置的默认工作区是一个**真实客户目录**，落到它上面会去 Boot 别人的包，然后报一个
  和你意图完全无关的错。**三层都空时拒绝启动**（`--workspace` → `TZSCLI_WS` → `tzs.workspace`），
  不回落。
- **设计器不在配置里**：它的程序集随包分发，引擎默认从 `<引擎目录>\designer\` 加载。
  所以同一份 tt 在任何机器上跑的是同一版设计器 —— 这是分发本身保证的，不需要谁去对齐配置。

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

**默认后台常驻**（单实例）：打印实际地址后立即返回，终端可以接着敲别的命令；
`tt debug start/exec/status` 等控制命令会自动找到它。已在运行时打印它的地址后返回。

```
tt serve --foreground    # 前台运行，日志直出终端（Ctrl+C 停止）
tt serve --stop          # 停止后台实例
```

一个进程、一个端口、一套页面：

- `/debug/` — 调试工作台
- `/debug/#settings` — 统一设置页：站点管理 / 数据字典 / DEBUG / 应用设置
- `/api/*` — 统一接口：环境与数据库配置（`/api/hosts`）、配置派生状态、
  PATH 安装，以及字典类动作（源码镜像拉取、字典同步、BDL 文档）

**所有配置都在设置页里**：环境与数据库、查询数据源、镜像目录、同步目标、
BDL 文档目录、调试参数、`.tzs` 引擎依赖、明暗色。环境清单只有一份数据源
（config.json 的 hosts 节），调试、字典查询与源码镜像读的是同一份。

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
tt debug ents [--ent N]                  企业目录：有哪些企业(ENT)、各用哪个账号（带快照，离线可答）
tt debug db [--ent N]                    数据库连接体检：真连一次库验证账号可用

tt debug start <作业> [-m <模块>] [-z <区域>]   拉起调试会话
tt debug exec "<命令>"                    下断点/继续/求值（fgldb 原生命令）
tt debug status | quit | wait | why       会话管理
tt debug source | logs | locate | resolve | interrupt
tt debug wslogs | wsdebug | sql | wstest  接口日志与报文回放
tt debug mode | topent                    切换调试模式 / 企业
```

### `tt dev tzc` — 设计器代码包

```
tt dev tzc export <pkg.tzc> [-o <dir>] [--only <点名>...] [--json]
tt dev tzc status | verify | apply | unlock | rename | newfn | selftest
```

`export` 把包渲染成**带围栏的 4GL 工作区**（`prog.full.4gl` + `manifest.json` + 快照 + git），
改完 `apply` 走 gate1/gate2/gate3 写回 —— **`apply` 是唯一的写路径**，它做闸门校验与原子写。
围栏协议与不变量见 [docs/tzc-model.md](docs/tzc-model.md) 与 [docs/dev.md](docs/dev.md)。

退出码：`0` 成功 / `2` 包格式或用法错 / `3` 校验失败 / `4` 拒绝写入 / `5` IO 与环境失败。

### `tt dev tzs` — 设计器表单包

```
tt dev tzs export <pkg.tzs> [-o <dir>] [--force] [--json]   # 纯解压（只读参考）
tt dev tzs call <fn> [--<参数> <值>…]                        # 读写表单（设计器自己的引擎）
tt dev tzs fns [<fn>] | manifest | doctor | stop | reap
```

- **`export` 是纯解压**：把 zip 逐条目摊到目录里（默认 `<包目录>/<程序名>-unzip`），
  不解围栏、不校验、不产生工作区。产物是**只读参考**，不要改完再塞回包。
  它**不依赖引擎，也不依赖设计器** —— 没装设计器也能用。
- **读写表单走 `call`**，49 个函数（`fns` 看全表）。这是设计器自己的模型在算，
  改完设计器打得开；手工拼 XML 则不然。

退出码与 `tzc` 那套**不一样**：没有 `3`，多一个 `1`。
`0` 成功 / `1` 引擎内部错 / `2` 参数或环境不对 / `4` 设计器拒绝（含 `E_KEY_IN_USE`）/
`5` 传输或环境失败。

细节（句柄语义、`validate` 的基线规则、参数语法、逐函数说明）见
[skills/tt-dev-tzs/SKILL.md](skills/tt-dev-tzs/SKILL.md)。

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

## `.tzs` 表单引擎

`tt dev tzs call` 背后是 `engine/` 里一个 **C# 引擎**（`tzs-server.exe`）。它**不实现**
`.tzs` 格式 —— 它 `Assembly.LoadFrom` **设计器自己的程序集**，布局属性走设计器自己的
`XmlElement` 索引器，`.tsd` 由设计器从模型重算，验收用设计器自己的校验器加 RoundTrip 不动点。
我们这部分总共 200 KB（`TzsCli.dll` 26 KB + `TzsCli.Designer.dll` 180 KB）。
**设计器那 10.5 MB 由发行包自带**，在 `tzs\designer\` 下。

`tt` 通过命名管道上的 JSON-RPC 驱动它（`internal/dev/tzs/`，`tt` 自己实现的 Go 客户端）。

### 为什么单独构建

1. **它不属于 Go 的构建链。** 用 `csc.exe`（Framework64 v4.0.30319，**C# 5**——没有模式匹配、
   没有 `nameof`、没有字符串插值）编译，引用 GAC 里的 WPF 程序集。`build_portable.bat`
   只**采集**产物，不构建它。**只在引擎真的改了时才重编。**
2. **设计器程序集：构建期向机器要一份，运行期自带一份。** 构建期 `-r:` 它的
   `Newtonsoft.Json.dll`；运行期 `LoadFrom` 它的 `SpecDesignerCommon.dll` / `FormEditor.dll` /
   `UndoRedoFramework.dll`，以及分散在多个程序集里的语言字典 —— **这些由包自带**。
   引擎的解析规则只有两条：`TZSCLI_INSTALL`（开发/构建期的逃生口），否则 `<自己的目录>\designer`。
   **没有"回落到用户装的那份"这一条**，所以少带一个文件是打包坏了，不是"去别处找找"。
3. **重编会让所有在跑的守护进程变成孤儿。** 守护进程的管道名 = `hash(工作区)` +
   **本程序集的 MVID 前 8 位**，MVID 每次重编都变。客户端因此**永远够不到**跑着陈旧字节的
   守护进程（刻意的，否则你会和旧行为对话而看不出来）；代价是每次重编后，上一个构建起的
   守护进程**再也停不掉**（新名字没人监听，旧名字没人知道）。清理由 `tt dev tzs reap` 兜底。

   把引擎构建挂在 `tt` 的每次构建上，等于每次发版都制造一批停不掉的进程 ——
   一个和 `tt` 无关的构建步骤造成用户可见的后果。

### 进分包的是什么

```
tzs-server.exe        服务端（命名管道 / --stdio 两种模式）
tzs-cli.exe           客户端（独立可用；tt 自己实现了一份 Go 客户端）
TzsCli.dll            纯文本/zip 层，不反射
TzsCli.Designer.dll   反射管线 + 49 个函数
designer\             设计器的 28 个 .dll（第三方商业软件，见下）
```

`engine/out/` 里还有十几个探测程序（`Probe` / `Edit` / `AddField` / `RoundTrip` / `Test*` / `E2E`），
**不要 xcopy 整个目录** —— `build_portable.bat` 显式按名字采那四个到 `<stage>\tzs\`。

`designer\` 只采 `*.dll`：那个目录里还有 `T100Designer.exe`（GUI，不在引擎的依赖闭包里）、
`AutoUpdater.exe` 与 AppLimit 的 Sparkle 更新组件（会连厂商的更新通道，**不发它**）、
`AutoUpdater.exe.config` 和一个快捷方式。

```bash
cd engine && ./build.sh                      # → engine/out/，15 个单元
TZSCLI_INSTALL='D:\APPS\某版本设计器' ./build.sh  # 换机器时先指设计器目录（只是编译期引用）

TZSCLI_INSTALL='D:\APPS\某版本设计器' build_portable.bat   # 打包时采进 tzs\designer\
```

**打包脚本的 `TZSCLI_INSTALL` 没有缺省**，这是故意的：采进去的那份就是这份发行版钉住的版本，
所以应当由人指明，而不是从某个写死的路径猜。设计器本身**不入库**（`.gitignore` 的 `*.dll`），
它只是打包输入。

其余（`SPEC.md` 格式契约、`HANDOFF.md` 交接、`TASKS.md` 任务板）见 [engine/BUILD.md](engine/BUILD.md)。

## 项目结构

```
TT/
├─ main.go                  //go:embed all:web/dist → cli.Execute
├─ internal/
│  ├─ config/               ★ 统一配置层：位置解析 / 读写 / schema / 迁移
│  ├─ cli/                  cobra 根命令
│  │  ├─ env.go config.go serve.go install.go version.go
│  │  ├─ common/            各命令组共享的 CLI 上下文（叶子包）
│  │  ├─ debug/ dev/ dict/  命令组
│  ├─ debug/                调试内核：fgldb 驱动、会话、REST+WS
│  ├─ dev/
│  │  ├─ tzc 管线           pkgfile/tapfile/tglfile/fgl/synth/fence/verify/split/store/model
│  │  └─ tzs/               ★ .tzs 引擎的 Go 客户端：manifest/wire/client/server/state/doctor
│  ├─ dict/                 数据源与查询：db（本地）/ live（远程）/ dbsync / server
│  ├─ web/                  统一 HTTP 服务与共享 /api/*
│  ├─ host/                 ★ 合并后唯一的 SSH/PTY/终端/探测/镜像层
│  ├─ dbconfig/ erpdb/      数据库连接模型与连接器（两处合并）
│  ├─ safesql/ sshtun/ output/ pathinstall/ atomic/ winproc/
├─ engine/                  ★ .tzs 表单引擎（C#，单独构建，见 engine/BUILD.md）
├─ web/
│  ├─ app/                  调试工作台 SPA（React + Radix + zustand + Monaco）
│  │                        其中的设置视图是所有命令组的统一配置页
│  ├─ shared/               整套 SPA 共用的主题层 / UI 基元 / 设置页布局件
│  └─ package.json          npm workspace 根
├─ skills/                  AI 技能：tt-debug / tt-dev-tzc / tt-dev-tzs / tt-dict / erp-read
├─ docs/                    分册文档与设计说明
└─ testdata/                FGL 夹具
```

`★` = 为合并或为 `.tzs` 而真正重组的部分。

## 依赖

| 要建什么 | 需要什么 | 说明 |
|---|---|---|
| 后端 `tt.exe` | **Go 1.26.5+** | 版本下限就是 `go.mod` 里那条 `go` 指令。**没有 `vendor/`**，首次构建要能取到模块（`go mod download`；`build_portable.bat` 里设了 `GOPROXY=https://goproxy.cn,direct`） |
| 前端 `web/dist` | **Node.js + npm** | 只用 npm workspaces（`web/` 是根，`web/app` 是唯一 workspace），没有 pnpm/yarn 的锁文件 |
| 打包 | **Python 3**（可选） | `tools/zip.py` 打 zip、`tools/wix_removefolders.py` 补卸载目录。**缺了不会失败** —— 打 zip 会退回 PowerShell，卸载目录那段才需要它 |
| `.tzs` 引擎 | **.NET Framework 4.0 的 `csc.exe`** + 一个 POSIX shell | 只在改动 `engine/` 时才要。编译器是 Windows 自带的 `C:\Windows\Microsoft.NET\Framework64\v4.0.30319\csc.exe`（**C# 5**），`build.sh` 是 bash 且用 `cygpath`，所以要 Git Bash 这类环境 |
| `.tzs` 引擎 | **已安装的 T100 设计器**（只在你重打发行包时要） | 打包时采它的 `*.dll` 进 `tzs\designer\`，那份就是这一版发行钉住的版本。编译期只从它取 `Newtonsoft.Json.dll`。引擎跑起来**不需要**机器上装着设计器 —— 它用包里自带的那份 |
| MSI | **WiX v3 工具集** | 只在打 MSI 时要，见下 |

Go 的**直接**依赖只有 8 个：`coder/websocket`（调试 WebSocket）、`jackc/pgx`（Kingbase/PostgreSQL）、
`sijms/go-ora`（Oracle）、`pkg/sftp`+`x/crypto`（SSH）、`spf13/cobra`（命令树）、`x/sys`（Windows 进程/管道）、
`modernc.org/sqlite`（本地字典镜像，纯 Go 无 cgo）。`go.mod` 里其余都是它们的间接依赖。

前端依赖见 `web/app/package.json`（React 18 + Vite 6 + Tailwind 4 + Radix + Monaco + zustand）。

## 构建

### 后端

**前端必须先构建**：`main.go` 用 `//go:embed all:web/dist` 把前端产物嵌进二进制，
`web/dist` 不存在时 `go build` 直接失败。

```bash
cd web && npm install && npm run build   # → web/dist（顺带跑 tsc --noEmit 做类型检查）
cd .. && go build -o tt.exe .            # → tt.exe
```

`build_portable.bat` 会把版本号注进去（`-ldflags "-X tt/internal/cli.Version=…"`），
直接 `go build` 出来的是 `devel`，`tt version` 看得出来差别。

### `.tzs` 引擎（`engine/`）

**只在引擎真的改了时才重编。** 它的产物不进 Go 的构建链，而且重编会换掉引擎程序集的 MVID，
让所有在跑的守护进程变成停不掉的孤儿（`engine/BUILD.md` 解释了这条约束）。

```bash
cd engine && ./build.sh                      # → engine/out/，15 个单元
TZSCLI_INSTALL='D:\APPS\某版本设计器' ./build.sh  # 换一台机器时先指设计器目录
./build.sh TzsCli.Designer                   # 只编一个
OUT=<dir> ./build.sh                         # 换落点
```

`build_portable.bat` / `build_msi.bat` **只采集产物、不构建它**。引擎那四个文件按名字采，
`engine/out/` 里还有十几个探测程序（`Probe` / `Edit` / `AddField` / `RoundTrip` / `Test*` / `E2E`），
xcopy 整个目录会把它们一起打进包里。缺任何一个都会让打包脚本报错退出。

同一份 `TZSCLI_INSTALL` 在打包时还有第二个作用：`build_portable.bat` 用它找到设计器目录，
把其中的 `*.dll` 采进 `<stage>\tzs\designer\`。**它没有缺省** —— 采的是哪一版，这一版发行就钉在
哪一版，所以由人指明比从写死的路径猜更正确（也因为 `.bat` 必须保持纯 ASCII，而常见的设计器路径
是中文的）。

### 便携包

```bash
build_portable.bat        # → dist/tt-portable/ 与 dist/tt-portable.zip
```

它会依次跑：前端构建 → `go build`（带版本号）→ 暂存 `tt.exe` + `config.empty.json`（作为包内的
`config.json`）+ `config.example.json` + `README.md` + `skills/` → 采引擎那四个文件到 `tzs\` →
打 zip。注意它**刻意不打包本机的 `config.json`**（含真实口令），包里放的是空骨架。

### MSI

```bash
build_msi.bat             # → dist/TT-0.1.0-x64.msi
```

依赖 WiX v3 工具集（**不需要装 .NET SDK**，解压即用）：把 `candle.exe` / `light.exe` / `heat.exe`
放到 `D:\tt-build-tools\wix3`，或用环境变量 `WIX_BIN` 指向它们所在的目录。脚本会先跑一遍
`build_portable.bat` 复用同一份载荷，再剥掉 `.portable` 与 `config.json`（装出来的版本配置应落在
`%APPDATA%\T100\tt\`，不是安装目录），用 `heat.exe` 采集文件（`tzs\` 自动跟着走，不用改 `tt.wxs`），
最后编译链接成 MSI。

### 自检

```bash
go test ./...                        # Go 全量；默认跳过 .tzs 语料回归（那个要 17 分钟）
cd web && npm run check:app          # 前端三项：fgltokens / fgloutline / store
tt dev tzc selftest                  # .tzc 的 31 项内置对抗用例，不需要真实语料
TTZS_DEEP=1 go test ./internal/dev/tzs/ -run TestCorpus -timeout 30m   # .tzs 语料回归
```

### 开发模式

```bash
tt serve                     # 后端（默认 127.0.0.1:28670）
cd web && npm run dev:app    # 前端热更新，代理到后端
```

端口被占用时后端会自动顺延，用 `TT_PROXY=http://127.0.0.1:<实际端口>` 告诉前端。

## 文档

| 文档 | 内容 |
|---|---|
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | 合并架构与统一配置设计 |
| [docs/MIGRATION.md](docs/MIGRATION.md) | 迁移记录：改了什么、为什么 |
| [docs/debug.md](docs/debug.md) | 调试器手册（原 TDebug README） |
| [docs/dev.md](docs/dev.md) | 设计器包工具手册（原 TDev README；`.tzs` 与验收清单都在里面） |
| [docs/dict.md](docs/dict.md) | 数据字典手册（原 TDictCli README） |
| [docs/tzc-model.md](docs/tzc-model.md) | `.tzc` 包模型与不变量 |
| [engine/BUILD.md](engine/BUILD.md) | `.tzs` 引擎：为什么单独构建、采哪四个文件 |
| [engine/SPEC.md](engine/SPEC.md) | `.tzs` 格式与契约的完整记录 |
| [skills/tt-dev-tzc/SKILL.md](skills/tt-dev-tzc/SKILL.md) | 给 AI 的操作手册：`.tzc` 代码包怎么改、哪些坑 |
| [skills/tt-dev-tzs/SKILL.md](skills/tt-dev-tzs/SKILL.md) | 给 AI 的操作手册：`.tzs` 表单包怎么读写、哪些坑 |

## 兼容性

合并前的命令名仍可直接当子命令组用，既有脚本与 AI skill 不改即可运行：

```
tt tdebug …  =  tt debug …
tt tdev …    =  tt dev …
tt tdict …   =  tt dict …
```

环境变量同理：`TDEBUG_CONFIG` / `TDICT_CONFIG` 仍被识别，`--conn` 仍是 `--env` 的别名。
