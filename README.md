# TT — T100 工具集

T100 日常工作的统一入口：**一个二进制 `tt`、一份配置、一个本地 Web 服务。**

```
tt debug …     作业调试器：SSH 驱动 fglrun -d 的 (fgldb) 文本调试协议，
               本地 Web 界面（源码/断点/调用栈/变量/接口日志）+ 命令行控制端
tt dev tzc …   设计器代码包（.tzc/.tzf/.tzx）：渲染成带围栏的 4GL 工作区给人/AI 改，
               改完走三道闸门写回
tt dev tzs …   设计器表单包（.tzs/.tzv）：export 纯解压只读；读写表单走**具名动词**
               (`tt dev tzs <动词> --args '<JSON>'`)，由设计器自己的引擎算，不是拼 XML
tt dict …      ERP 数据字典查询：r.t / r.v / desc / scc / r.q / prog，本地镜像与远程直查

tt env …       环境管理（各命令组共用同一份环境清单）
tt config …    配置管理：位置 / 查看 / 读写 / 迁移 / 校验
tt serve       启动本地 Web 服务（工作台 + 统一设置页；默认后台常驻，打印地址后返回）
tt install     安装 AI skills 到当前目录，或把 tt 加进用户 PATH
tt version
```

TT 是一款**面向 Agent 的 CLI 开发工具**，用于开发基于 Genero BDL 技术栈的大型 ERP 系统
—— 鼎捷数智旗下的 T100。它由 TDebug、TDev、TDictCli 三个工具合并而来，之后又长出了第四块：
**`.tzs` 表单引擎**（`engine/`，一个 C# 程序，反射驱动设计器自己的程序集）。

有一条边界贯穿全篇：**tt 不实现 T100 设计器的任何私有格式**。`.tzc` 靠设计器的公开发行物反推；
`.tzs` 更彻底 —— 它直接反射调用设计器自己的程序集，格式那部分是设计器自己在算。
设计器的 28 个程序集**随本仓库与发行包分发**，所以运行 tt 不需要另外安装设计器。

## 这份 README 管什么

**本文件只管两件事：把 tt 装起来、把它构建出来。** 别的内容各有归属：

| 想做什么 | 去哪 |
|---|---|
| 装 tt / 构建 tt / 知道配置放在哪 | **本文件** |
| 搞清设计、契约、不变量，以及"为什么是这样" | [docs/WIKI.md](docs/WIKI.md)（11 章，带章节索引） |
| 用某个命令 —— 怎么改 `.tzc`、怎么读写 `.tzs`、怎么查字典、怎么调程序 | [skills/](skills/) 下对应的 `SKILL.md` |
| 精确的参数、flag、退出码 | `tt <命令> --help`，或 `tt dev tzs <动词> --help` 看单个动词的参数 |
| 改 `.tzs` 引擎（C#） | [engine/BUILD.md](engine/BUILD.md) 与 [engine/SPEC.md](engine/SPEC.md) |
| 查 `.tzc` 那些 `file:line` 依据指向的第三方材料 | [docs/T100设计器-README.md](docs/T100设计器-README.md)（设计器反编译源码树的 README 副本） |

三份材料各管一段，互不重复：**README 管装和构建，WIKI 管为什么，skills 管怎么用。**

> 合并前的命令名仍是别名（`tt tdebug …` = `tt debug …`，`tt tdev …` = `tt dev …`，
> `tt tdict …` = `tt dict …`），`TDEBUG_CONFIG` / `TDICT_CONFIG` 仍被识别，`--conn` 仍是
> `--env` 的别名 —— 既有脚本不改即可运行。

## 安装

### 便携包

把 `tt-portable.zip` 解压到任意目录，双击或命令行运行 `tt.exe` 即可。
包内有 `.portable` 标记，配置就近留在包内（`config.json`），不写用户目录。

包内还带 `tzs\`（引擎三个文件 + `designer\` 里的设计器程序集）与五套 AI 技能。
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

### AI 技能（也就是使用手册）

```
tt install skills                       # 复制到 <当前目录>/skills
tt install skills --to .claude/skills   # 装到 Claude Code 直接读的位置
```

`skills/` 下是五套：`tt-debug`（调试）、`tt-dev-tzc`（代码包 `.tzc`）、`tt-dev-tzs`（表单包 `.tzs`）、
`tt-dict`（数据字典）、`erp-read`（读 ERP 代码）。一个技能一个目录、目录里是 `SKILL.md`
（Claude 技能规范：目录名必须等于 frontmatter 里的 `name`）。装一次全部到位。

**这五份也是给人看的操作手册** —— 每个命令组怎么用、哪些坑，都在里面，别处不再重复。

## 快速上手

```bash
tt serve                                    # 起服务，浏览器打开它打印的地址
# 在「设置 → 站点管理」里加一台环境（SSH + 数据库），之后：

tt debug start <作业> -m <模块>              # 调程序 → 细节见 skills/tt-debug
tt dev tzc export "D:\pkg\x.tzc"            # 改 4GL 客制 → 细节见 skills/tt-dev-tzc
tt dev tzs field add --args '{"file":"D:\\pkg\\x.tzs","table":"pmdl_t","columns":["pmdlent","pmdlsite"],"out":"D:\\pkg\\_ai.tzs"}'
                                            # 改表单 → 细节见 skills/tt-dev-tzs（53 个具名动词 + JSON 参数）
tt dict r.t --kw 应收                        # 查字典     → 细节见 skills/tt-dict
```

`.tzs` 还要配一样：`tt config set tzs.workspace "D:\你的工作区"`（**这条没有缺省**，
引擎内置的默认工作区是一个真实客户目录）。`tt dev tzs doctor` 会告诉你还差什么。

## 配置在哪

**只有一个配置文件**，所有命令组共用，位置按优先级（第一个存在的胜出）：

1. `TT_CONFIG` 环境变量（旧名 `TDEBUG_CONFIG` / `TDICT_CONFIG` 仍可用）
2. `--config <路径>`
3. `<exe 目录>\.portable` 存在 → 便携包，配置留在包内
4. `%APPDATA%\T100\tt\config.json` —— 默认
5. 旧位置兜底（首次运行自动合并迁移）

`T100_HOME` 整体改写统一目录（如 `T100_HOME=D:\t100`）。**字段的完整说明、迁移规则、
以及"哪些能写在界面上"见 [docs/WIKI.md](docs/WIKI.md#2-统一配置)。**

## 依赖

| 要建什么 | 需要什么 | 说明 |
|---|---|---|
| 后端 `tt.exe` | **Go 1.26.5+** | 版本下限就是 `go.mod` 里那条 `go` 指令。**没有 `vendor/`**，首次构建要能取到模块（`go mod download`；`build_portable.bat` 里设了 `GOPROXY=https://goproxy.cn,direct`） |
| 前端 `web/dist` | **Node.js + npm** | 只用 npm workspaces（`web/` 是根，`web/app` 是唯一 workspace），没有 pnpm/yarn 的锁文件 |
| 打包 | **Python 3**（可选） | `tools/zip.py` 打 zip、`tools/wix_removefolders.py` 补卸载目录。**缺了不会失败** —— 打 zip 会退回 PowerShell，卸载目录那段才需要它 |
| `.tzs` 引擎 | **.NET Framework 4.0 的 `csc.exe`** + 一个 POSIX shell | 只在改动 `engine/` 时才要。编译器是 Windows 自带的 `C:\Windows\Microsoft.NET\Framework64\v4.0.30319\csc.exe`（**C# 5**），`build.sh` 是 bash 且用 `cygpath`，所以要 Git Bash 这类环境 |
| MSI | **WiX v3 工具集** | 只在打 MSI 时要，见下 |

上表**没有"T100 设计器"** —— 它的 28 个程序集已经入库（`engine/designer/`），跟着 clone 一起下来，
不需要谁去装一份。发行包也带同样一份。要换设计器版本就换那几个文件并提交，见「构建」。

Go 的**直接**依赖只有 8 个：`coder/websocket`（调试 WebSocket）、`jackc/pgx`（Kingbase/PostgreSQL）、
`sijms/go-ora`（Oracle）、`pkg/sftp`+`x/crypto`（SSH）、`spf13/cobra`（命令树）、`x/sys`（Windows 进程/管道）、
`modernc.org/sqlite`（本地字典镜像，纯 Go 无 cgo）。`go.mod` 里其余都是它们的间接依赖。

前端依赖见 `web/app/package.json`（React 18 + Vite 6 + Tailwind 4 + Radix + Monaco + zustand）。

## 构建

**克隆下来什么都能构建，包括 `.tzs`** —— 设计器的 28 个程序集随仓库一起下来（`engine/designer/`）：

| 功能 | 需要 |
|---|---|
| `tt debug` / `tt dict` / `tt dev tzc` / Web 界面 | Go + Node |
| `go test ./...` | Go |
| `tt dev tzs`（表单读写，含 `doctor`） | 上面这些 + `cd engine && ./build.sh` |
| 打发行包（便携 zip / MSI） | 上面这些；MSI 另要 WiX |

T100 设计器是第三方商业软件（厂商标识 DSC）。它的程序集**已入库**在 `engine/designer/`，
发行包里也带一份（`tzs\designer\`）。所以无论 clone 还是装包，**都不用另装设计器、也不用配它的
路径** —— 引擎默认从自己旁边的 `designer\` 加载（`engine/src/Designer/Bootstrap.cs`）。

### 一条命令都不少

```bash
git clone <repo> && cd TT

cd web && npm install && npm run build && cd ..   # 前端（可先跳过，见下）
go build -o tt.exe .                              # 后端 —— 到这里 tt debug / dict / dev tzc 就能用
go test ./...                                     # 21 个包；语料回归默认跳过

cd engine && ./build.sh && cd ..                  # .tzs 引擎（C#）→ engine/out/
```

**前端可以先跳过**：`web/dist/.gitkeep` 这个占位文件让 `//go:embed all:web/dist` 在没有产物时
也成立，所以 `go build` **不会**失败 —— 但那样出来的 tt 没有界面，`tt serve` 会返回一张写着
「界面未构建」和构建命令的说明页（不是静默空白）。

`./build.sh` 会把仓库里的 `engine/designer/` 一并采到 `engine/out/designer/` —— 那正是引擎默认
去找的地方，所以编完就能跑，**没有任何环境变量要设**。

`build_portable.bat` 会把版本号注进去（`-ldflags "-X tt/internal/cli.Version=…"`），
直接 `go build` 出来的是 `devel`，`tt version` 看得出来差别。

### `.tzs` 引擎（`engine/`）

**只在引擎真的改了时才重编。** 它的产物不进 Go 的构建链，而且重编会换掉引擎程序集的 MVID，
让所有在跑的守护进程变成停不掉的孤儿（`engine/BUILD.md` 解释了这条约束）。

```bash
cd engine && ./build.sh                      # → engine/out/，并把 designer/ 采进去
./build.sh TzsCli.Designer                   # 只编一个
OUT=<dir> ./build.sh                         # 换落点
TZSCLI_INSTALL=<别的设计器目录> ./build.sh    # 用别的版本覆盖仓库里那份
```

`build.sh` 干两件事：**编译期**从设计器那份里取 `Newtonsoft.Json.dll`，然后把
`engine/designer/` 采到 `engine/out/designer/`。于是 `engine/out/tzs-server.exe` 不需要任何环境
变量就能找到设计器 —— 与发行包 `<引擎目录>\designer\` 是同一个布局。

`build_portable.bat` / `build_msi.bat` **只采集产物、不构建它**。引擎那三个文件按名字采，
`engine/out/` 里还有十几个探测程序（`Probe` / `Edit` / `AddField` / `RoundTrip` / `Test*` / `E2E`），
xcopy 整个目录会把它们一起打进包里。缺任何一个都会让打包脚本报错退出。
引擎自带的 C# 客户端 `tzs-cli.exe` 不采 —— 对外只有 `tt` 一个入口（理由见
[engine/BUILD.md](engine/BUILD.md)）。

打包时设计器的来源默认也是 `engine\designer\`，所以**打发行包同样不用配任何东西**；
`TZSDESIGNER` 可以覆盖它，用来打一个装别版设计器的包。发行版因此钉在仓库里那一版 ——
换设计器版本 = 换 `engine/designer/` 下的文件并提交，是一次可 review 的改动。

### 从源码跑 `.tzs`

编完就能用，只差把 `tt` 指到那个引擎 —— `tt` 默认去 `<tt.exe 目录>\tzs\` 找它，那是**发行包**的
布局，源码树里没有这一层：

```bash
tt config set tzs.workspace  "D:\你的工作区"                  # 没有缺省，必须给
tt config set tzs.serverExe  "<仓库路径>\engine\out\tzs-server.exe"
```

然后 `tt dev tzs doctor` 应该全绿：引擎 exe / 设计器目录 / 工作区 / 管道名。
**不需要设 `TZSCLI_INSTALL`** —— 设计器就在 `engine/out/designer/`，引擎自己找得到。

`tzs.serverExe` 平时不该写（发行版靠 `<tt.exe 目录>\tzs\` 的固定布局），它存在的意义就是
源码树与自定义部署这两种情况。

### 便携包

```bash
build_portable.bat        # → dist/tt-portable/ 与 dist/tt-portable.zip
```

它会依次跑：前端构建 → `go build`（带版本号）→ 暂存 `tt.exe` + `config.empty.json`（作为包内的
`config.json`）+ `config.example.json` + `README.md` + `skills/` → 采引擎那三个文件到 `tzs\` →
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
