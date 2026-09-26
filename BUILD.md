# 从源码构建 TT

**克隆下来什么都能构建，包括 `.tzs`** —— 设计器的程序集随仓库一起下来（`engine/designer/`），
不需要另外安装设计器，也不需要配置它的路径。

## 1. 依赖

| 要建什么 | 需要什么 | 说明 |
|---|---|---|
| 后端 `tt.exe` | **Go 1.26.5+** | 版本下限就是 `go.mod` 里那条 `go` 指令。**没有 `vendor/`**，首次构建要能取到模块 |
| 前端 `web/dist` | **Node.js + npm** | 只用 npm workspaces（`web/` 是根，`web/app` 是唯一成员），没有别的包管理器的锁文件 |
| `.tzs` 引擎 | **.NET Framework 4.0 的 `csc.exe`** + 一个 POSIX shell | 只在改动 `engine/` 时才要。编译器是 Windows 自带的 `C:\Windows\Microsoft.NET\Framework64\v4.0.30319\csc.exe`（**C# 5**），`build.sh` 是 bash 且用 `cygpath`，所以要 Git Bash 这类环境 |
| 打包 zip | Python 3（可选） | 缺了会退回 PowerShell |
| MSI 安装包 | WiX v3 工具集 + Python 3 | 只在打 MSI 时要 |

**上表没有"T100 设计器"** —— 它的 28 个程序集已经入库（`engine/designer/`），跟着 clone 一起
下来。发行包也带同样一份。

Go 的**直接**依赖只有 8 个：`coder/websocket`（调试 WebSocket）、`jackc/pgx`（Kingbase/PostgreSQL）、
`sijms/go-ora`（Oracle）、`pkg/sftp` + `x/crypto`（SSH）、`spf13/cobra`（命令树）、`x/sys`（Windows
进程与管道）、`modernc.org/sqlite`（本地字典镜像，纯 Go 无 cgo）。`go.mod` 里其余都是它们的间接依赖。
前端依赖见 `web/app/package.json`。

模块代理：`build_portable.bat` 里设了 `GOPROXY=https://goproxy.cn,direct`；手工构建时按自己的
网络环境设。

## 2. 首次搭建

```bash
git clone <repo> && cd TT

cd web && npm install && npm run build && cd ..   # ① 前端（可先跳过，见下）
go build -o tt.exe .                              # ② 后端 —— 到这里 debug / dict / dev tzc 就能用
go test ./...                                     # ③ 自检

cd engine && ./build.sh && cd ..                  # ④ .tzs 引擎（只在要用 tzs 时）
```

**① 前端可以先跳过。** 产物目录里的占位文件让后端的嵌入在产物缺失时也成立，所以 `go build`
**不会**失败 —— 但那样出来的 tt 没有界面，`tt serve` 会返回一张写着「界面未构建」和构建命令的
说明页（不是静默空白）。想省这一步就省，之后随时补构建即可。

**④ 引擎只在要用 `tt dev tzs` 时构建**，而且**只在它真的改了时才重编** —— 重编会换掉程序集的
MVID，而守护进程的管道名包含它，于是上一个构建起的守护进程再也停不掉。详见第 5 节。

`go build` 直接出来的版本号是 `devel`；`build_portable.bat` 会把版本号注进去
（`-ldflags "-X tt/internal/cli.Version=…"`），`tt version` 看得出来差别。

## 3. 三条构建链

| 链 | 工具 | 产物 | 何时跑 |
|---|---|---|---|
| 前端 | Node + vite | `web/dist` → 被后端嵌入 | 改了前端 |
| 后端 | Go | `tt.exe` | 改了 Go |
| 引擎 | `csc.exe` + bash | `engine/out/` | **只在改了 `engine/` 时** |

**引擎为什么必须单独构建**：

1. **它不属于 Go 的构建链** —— 用 `csc.exe` 引用 GAC 里的 WPF 程序集，没有工程文件、不用 MSBuild。
   打包脚本只**采集**产物，不构建它。
2. **重编会让在跑的守护进程变成孤儿**（管道名含 MVID）。
3. **设计器程序集跟着产物走**：`build.sh` 编完把 `engine/designer/` 采到 `engine/out/designer/`，
   于是引擎不需要任何环境变量就能跑起来 —— 与发行包 `<引擎目录>\designer\` 是同一个布局。

## 4. 从源码树跑 `.tzs`

源码树里没有发行包那层 `tzs\`，而 `tt` 默认去 `<tt.exe 目录>\tzs\` 找引擎，所以要指两处：

```bash
tt config set tzs.workspace  "D:\你的工作区"                              # 没有缺省，必须给
tt config set tzs.serverExe  "<仓库路径>\engine\out\tzs-server.exe"
tt dev tzs doctor                                                       # 应全绿
```

**不需要设 `TZSCLI_INSTALL`** —— 设计器就在 `engine/out/designer/`，引擎自己找得到。
`tzs.serverExe` 平时不该写（发行版靠固定布局），它存在的意义就是源码树与自定义部署这两种情况。

## 5. 引擎的构建

```bash
cd engine && ./build.sh                      # → engine/out/，并把 designer/ 采进去
./build.sh TzsCli.Designer                   # 只编一个
OUT=<dir> ./build.sh                         # 换落点
TZSCLI_INSTALL=<别的设计器目录> ./build.sh    # 用别的设计器版本覆盖仓库里那份
```

`build.sh` 干四件事：定路径 → 编两个库（并校验分层没被破坏）→ 编十几个程序 → 把
`engine/designer/` 采到产物目录。

**产物目录里除了引擎本体还有十几个探测程序**（`Probe` / `Edit` / `AddField` / `RoundTrip` /
`Test*` / `E2E`），它们只供开发、**不进包**。

**换设计器版本** = 换 `engine/designer/` 下的文件并提交，是一次可 review 的改动。
`build_portable.bat` 只采那一份，所以打发行包同样不用配任何东西；`TZSDESIGNER` 可以覆盖它，
用来打一个装别版设计器的包。

## 6. 开发模式

```bash
tt serve                     # 后端（默认 127.0.0.1:28670）
cd web && npm run dev:app    # 前端热更新，代理到后端
```

端口被占用时后端会自动顺延，用 `TT_PROXY=http://127.0.0.1:<实际端口>` 告诉前端。

## 7. 自检

```bash
go test ./...                        # Go 全量；默认跳过 .tzs 语料回归
cd web && npm run check:app          # 前端三项检查（词法 / 大纲 / 状态容器）
cd web && npm run build              # 构建含类型检查（check:app 不做类型检查）
./tt.exe dev tzc selftest            # .tzc 的内置对抗用例，不需要真实语料
./tt.exe dev tzs doctor              # .tzs 引擎环境自检
```

**深度回归是显式开关，别顺手开**：

| 开关 | 跑什么 | 代价 |
|---|---|---|
| `TDEV_DEEP=1 go test ./... -timeout 30m` | `.tzc` 全语料 export+verify / apply 仿真 | 9–11 分钟 |
| `TTZS_DEEP=1 go test ./internal/dev/tzs/ -run TestCorpus -timeout 30m` | `.tzs` 全语料回归 | 17 分钟 |
| `TDEV_CORPUS` / `TTZS_CORPUS` | 覆盖语料根（两条管线都认） | — |
| `TTZS_CORPUS_LIMIT=N` | 只跑前 N 个包（冒烟） | 秒级 |

它们对上百个真实包各跑一遍，而 `go test` 默认超时 10 分钟 —— 表现为**随机器负载时好时坏的假
失败**，所以默认跳过。两条纪律：

1. **先在副本上跑。** 缺省语料根是**真实客户目录**，而回归会**在源包旁边**写临时包。
   用 `TTZS_CORPUS=%TEMP%\ttws`（整份拷贝）跑。
2. **整轮在跑的时候不要重建引擎。** 构建会覆盖产物，而每个包都要 spawn 探测程序，撞上重写会
   得到"文件找不到"的假红，报告上看不出是构建造成的。

## 8. 打包

```bash
build_portable.bat        # → dist/tt-portable/ 与 dist/tt-portable.zip
build_msi.bat             # → dist/TT-0.1.0-x64.msi（需 WiX v3）
```

两个脚本都**只采集、不构建**：

- **便携包**依次跑：前端构建 → `go build`（带版本号）→ 暂存 `tt.exe` + 空配置骨架 +
  示例配置 + README + `skills/` → 采引擎那三个文件到 `tzs\` → 采设计器到 `tzs\designer\` → 打 zip。
  它**刻意不打包本机的 `config.json`**（含真实口令）。
- **MSI** 复用便携包的载荷（先跑一遍便携打包），再剥掉便携标记与配置文件 —— 装出来的版本配置
  应落在 `%APPDATA%\T100\tt\`，不是安装目录。文件清单由 `heat.exe` 采集，卸载要删的目录由
  `tools/wix_removefolders.py` 补。WiX 放在 `D:\tt-build-tools\wix3`，或用 `WIX_BIN` 指向它。

**进包的文件是按名字手列的**（引擎那三个 + README + skills + 设计器），缺一个都会让脚本报错退出；
`engine/out/` 里那十几个探测程序绝不能整目录拷贝。

**版本号写在多处，改版本要一起改**：`build_portable.bat`、`build_msi.bat`、
`internal/dev/cli/cli.go` 的 `ToolVersion`、以及 `web/` 与 `web/app/` 的 `package.json`。

## 9. 几条构建相关的约束

- **前端产物目录里的占位文件是承重墙**：后端的嵌入靠它成立，删掉 `go build` 直接失败。
  前端构建会清空产物目录，由构建后脚本还原。
- **`engine/out/` 不能落在 `dist/` 下**：便携打包第一件事就是删 `dist`。
- **`.bat` 必须保持纯 ASCII**：cmd.exe 按 OEM 码页解析批处理文件，UTF-8 中文会让解析器断在
  字符中间。中文只写在 Markdown 里。
- **仓库的行尾是混合的**：Go 源与 Markdown 是 LF，部分 C# 与部分文档是 CRLF。批量替换（`sed -i`、
  用文本模式写文件的脚本）会整份翻掉行尾，判据是 `git diff --stat` 与
  `git diff --ignore-cr-at-eol --stat` 的差别。
