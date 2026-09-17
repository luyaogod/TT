# TDebug — T100 作业调试器

通过 SSH 在 T100 服务器上驱动 `fglrun -d` 的 `(fgldb)` 文本调试协议，提供**本地 Web 调试界面**
与**命令行控制端**，实现"人操作 GDC 界面 + AI 借助命令行检查分析"的人机协同调试。

本手册对应合并后的 `tt debug` 命令组（原 TDebug，独立成库时是从 TDictCli 剥离出来的纯调试项目）：只保留调试能力，数据字典/镜像/查询等由 `tt dict` 提供。

## 功能

- **Web 界面**（`tt serve` 后浏览器打开）：源码 + 断点 + 调用栈 + 变量监视（Monaco 编辑器、
  悬停求值、大纲、运行到光标），接口报文日志与重放调试，服务测试，环境/数据库/参数设置。
  主题支持**亮色 / 暗色 / 跟随系统**（「设置 → 外观」三张卡片，跟随系统会随操作系统实时切换）；
  还没有配置任何服务器环境时，打开界面直接落到「设置 → 环境」，不会把用户丢在空调试页。
- **命令行控制端**：`start` / `exec` / `stop` / `source` / `logs` / `locate` / `resolve` / `interrupt`
  等，自动发现后台服务地址；`exec` 透传全部 fgldb 标准调试命令并返回原生文本。
- **人机协同**：程序跑到 INPUT/MENU 等交互语句时会阻塞在 GDC 等人操作（从调试器看与死循环无法区分），
  技能文档给出了判断信号与交接话术（见 `skills/tt-debug/SKILL.md`）。
- **防呆**：停站停留超时看门狗自动放行（默认 1800s，保护生产区行锁）、断点持久化、单实例服务。

## 构建

需要 Go 1.26+ 与 Node 18+（前端产物经 `go:embed` 打进二进制，必须先构建前端）。

```bash
cd web && npm install && npm run build   # 生成 web/dist
cd .. && go build -o tt.exe .
```

Windows 一键打包，两种交付形态任选（互不影响，用的是同一个 `tt.exe`）：

| 形态 | 命令 | 产物 |
| --- | --- | --- |
| **便携包**（exe + config + README，浏览器用界面） | `build_portable.bat` | `dist/tt-portable.zip` |
| **MSI 安装包**（用户级安装，免管理员） | `build_msi.bat` | `dist/TT-<版本>-x64.msi` |

界面在浏览器里（`tt serve` 后打开 http://127.0.0.1:28670/debug/）。MSI 的细节见
[README 的安装一节](../README.md#msi-安装包用户级免管理员)。

## 快速开始

1. 配置环境（二选一）：
   - 复制 `config.example.json` 为 `config.json`，填写 `hosts.sshs`（SSH/区域/企业/数据库）；
   - 或启动服务后在 Web 界面「设置 → 环境」页里增删改。
2. 启动服务并打开界面：

   ```bash
   tt serve                 # 后台常驻(单实例),打印地址后返回
   # 浏览器打开 http://127.0.0.1:28670(端口被占用会自动顺延,以打印地址为准)
   tt serve --stop          # 停止后台实例
   tt serve --foreground    # 前台运行,日志直出终端
   ```

3. 命令行调试：

   ```bash
   tt debug start bsft001_wf -m asf      # 连 SSH + 启动作业 + 等入口停站(返回 JSON 快照)
   tt debug exec "break 4450"            # 透传 fgldb 命令
   tt debug exec "print ls_sql"
   tt debug exec "print g_req_param" "print g_status"   # 一次多条(批量)
   tt debug exec --file cmds.txt                        # 或从文件读,一行一条
   tt debug exec "continue" --timeout 300
   tt debug stop                         # 可复取的停站现场
   tt debug quit                         # 结束会话(作业窗口随之关闭)
   ```

## 命令一览

| 命令 | 作用 |
| --- | --- |
| `serve` | 启动本地调试服务（默认后台常驻单实例；`--listen`/`--foreground`/`--stop`） |
| `status` | 查看服务状态与活动会话 |
| `start <作业>` | 连接 SSH、启动调试并等到入口停站（`--module/-m`、`--zone`、`--ssh`、`--timeout`） |
| `exec "<命令>" [更多…]` | 透传标准调试命令（print/break/next/where/info/watch…），原样返回输出；**可一次给多条**（或 `--file` 从文件读）省掉每条一次进程启动，逐条标注 `[i/N]`；resume 类命令执行后读一次现场，停住了就继续往下跑（「走一步再取一批值」一次调用做完），没停住才中止并把剩余标为未执行；resume 类命令用 `--wait N` 软等待（到点返回，不发 SIGINT），`--timeout` 才是会发 SIGINT 的硬超时 |
| `quit` | 结束当前调试会话 |
| `stop` | 查看当前停站现场（状态/位置/函数/断点/TOPENT，可反复取用；带 `waitingForUser`/`waitingKind`） |
| `why` | 探测程序此刻在**等用户操作**还是在**空转**（interrupt → where → 判定停站行是否交互语句；默认探完自动放回，`--no-resume` 保留现场） |
| `wait` | 等会话事件（`--for stopped,exit,dead,watchdog`、`--timeout`）：事件一到即返回，超时返回 `timedOut`（不是错误） |
| `source` | 读服务器源码（登录区源码目录白名单只读，`--from/--to` 取行段，`--path-only` 只要行数与本地副本路径）；每次读取都会落一份本地副本供整读 |
| `logs` | 会话最近事件（`--tail N`） |
| `locate <函数>` | 定位函数定义到 文件:行（需停站） |
| `resolve <作业>` | 解析作业编号 → 实体程序/模块（不建会话） |
| `interrupt` | 中断运行中/卡住的程序，回到调试器 |
| `tt env list` / `tt env use <环境名>` | 查看全部环境 / 切换会话所在环境（SSH 配置）；环境清单与 `tt dict`、源码镜像共用（调试本地仍保留 `tt debug env [环境名]`） |
| `tt env show <环境名>` / `tt env topent <环境名> <编号>` | 查看某环境连接详情 / 设置它的 TOPENT 默认值（调试本地仍保留 `tt debug topent [值]`，`--clear` 清除会话级 override） |
| `sql "<语句>"` | 执行一条**只读** SQL 查业务数据（单条 SELECT/WITH；账号由 TOPENT 经 `gzou_t` 解析并在结果头回显「企业→账号」；白名单 + 库侧只读事务双重约束，最多回 200 行；`--ent` 指定企业、`--file` 从文件读；可用 `hosts.sshs[].db.readonlySql=false` 关闭） |
| `wslogs` | 接口报文日志列表（`--job` 作业编号(支持通配，按作业找日志用这个)、`--service` 服务名、`--server`、`--origin`、`--result`、`--pid`、`--from`/`--to`、`--fail`、`--page`、`--show <rowid>` 看单条报文；条件口径对齐 T100 原生 awsq990） |
| `wsdebug <rowid>` | 按日志报文参数重放调试，停在入口；`--set 路径=值` 改入参再重放（可重复）、`--request-file` 整份替换报文；启动后打印实际生效的 TOPENT（它不是数字时会提示） |
| `db [--ent N]` | 数据库连接探查：企业(TOPENT) → 账号映射与连接验证 |
| `probe` | 协议驱动器自检尖刺：登录→启动→下断点→步进→求值（`-m/-p/-l`） |
| `tt install skills` | 把 exe 同目录的 `skills/` 复制到目标目录（`--to <dir>` 换目标、`--force` 覆盖；默认 `<当前目录>/skills`） |
| `tt install path` | 把 exe 所在目录加入**用户** PATH（HKCU，不需管理员；`--dry-run` 只预览）；合并后一次安装覆盖全部工具 |

全局参数：`--config <路径>`（缺省取统一用户目录 `%APPDATA%\T100\tt\config.json`，见「config.json」）、`--json`、`-v`；
控制端命令（start/exec/status/quit/stop/source/logs/locate/resolve/interrupt/env/topent/wslogs/wsdebug）另有 `--url` 覆盖自动发现的地址
（`tt serve` 与 `tt debug serve` 的调试 API 挂在不同前缀下，自动寻址会从状态文件里带上；手工指定的地址若已写到 `/api`，按原样用）。

## config.json

合并后**只有一个**配置文件，三个工具共用，缺省 `%APPDATA%\T100\tt\config.json`（可用 `T100_HOME` 整体改写父目录）。顶层节如下；Web「设置」页保存时只改写它负责的那一节，其余键原样保留。

```jsonc
{
  "schemaVersion": 2,
  "listen": "127.0.0.1:28670",      // 统一 Web 服务的监听地址(工作台 /debug/ + 设置页 /debug/#settings)
  "hosts": {                        // ★ 共用环境清单：三个工具都读这一份
    "activeEnv": "示例测试区",        // 默认环境；调试侧可用下面的 debug.activeEnv 覆盖
    "sshs": [                       // 环境列表(设置-环境 页维护；原来在 debug.sshs)
      {
        "name": "示例测试区",
        "host": "10.0.0.2", "port": 22, "user": "tiptop", "password": "tiptop",
        "zone": "2",                // 登录菜单选项号(31开发/35测试/36正式/…)
        "topent": "9999",           // 默认企业编号(连接会话即下发到 shell;会话内可覆盖)
        "launchArgs": "",           // 覆盖全局启动参数模板({prog} 替换为作业名)
        "watchdogSeconds": 0,       // 覆盖全局看门狗秒数
        "db": {                     // 该环境一对一挂的数据库(作业解析/日志查询/字典查询用)
          "type": "oracle",         // oracle | kingbase
          "host": "10.0.0.2", "port": 1521, "service": "t35prd",
          "readonlySql": true,      // 省略 = 启用只读 SQL 守卫
          "accounts": [{ "account": "ds", "password": "ds" }]
        }
      }
    ]
  },
  "debug": {                        // 调试设置(不再承载环境列表)
    "activeEnv": "",                // 覆盖 hosts.activeEnv；留空即继承
    "launchArgs": "BBDL512840855a 2 12345 'N' {prog}",
    "watchdogSeconds": 1800,        // 停站停留超时自动放行(保护生产区行锁)
    "fglserver": "",                // 留空由 T100 按 SSH 来源 IP 自动设置
    "termWidth": 200, "termHeight": 50,
    "printElements": 1000,          // fgldb 单次 print 的数组元素上限
    "persistBreakpoints": null      // 断点持久化(默认开)
  }
}
```

**环境清单是三个工具共用的**：它住在 `hosts.sshs`（原 TDebug 放在 `debug.sshs`、TDictCli 放在 `hosts.sshs`，两份 schema 一样，合并后只留一处），同一份列表同时驱动 `tt debug`、`tt dict` 与源码镜像 —— 加一台机器一次，三处都能用。合并前的两份配置首次运行会自动合并迁移，预览用 `tt config migrate --dry-run`。

**当前环境**：缺省取 `hosts.activeEnv`（调试侧可用 `debug.activeEnv` 覆盖）；运行中的会话由「会话」面板或 CLI `tt debug env <名称>` 切换，无会话时取 `sshs` 首条，服务重启后回到首条。接口日志按当前会话所属环境的数据库查询。配置的位置与内容用 `tt config path` / `tt config show`（口令打码）看，`tt config validate` 校验。

### 环境变量

| 变量 | 作用 |
| --- | --- |
| `TT_CONFIG` | 覆盖配置文件路径（优先级高于 `--config`）；旧名 `TDEBUG_CONFIG` / `TDICT_CONFIG` 仍识别 |
| `T100_HOME` | 改写**统一用户目录**的父目录（缺省 `%APPDATA%\T100`）；三个工具共用，改一处一起生效 |
| `TT_PROXY` | 前端开发模式的后端地址（端口顺延时用，如 `TT_PROXY=http://127.0.0.1:<实际端口>`；原 `TDEBUG_PROXY`） |
| `TDEBUG_SERVE_LOG` | 后台服务子进程写入的日志路径（由 `serve` 自动设置） |
| `TDBG_RAW=1` | 把 fgldb 协议原始行打到服务日志（排障用） |

### 运行时产物（均已在 .gitignore 中）

- `.tt-serve.json` —— 后台实例状态（pid/地址/**调试 API 前缀**/日志），控制端命令据此自动寻址：
  `tt serve` 与 `tt debug serve` 写同一份，`apiBase` 区分对面把调试面挂在 `/debug/api` 还是 `/api`
- `.tt-serve.log` —— 后台服务日志
- `debug-bps/<模块>__<作业>.json` —— 断点持久化

## 目录结构

合并后本工具的代码不再是一个独立仓库，而是统一项目 TT 里的一个命令组。
完整布局见 [ARCHITECTURE.md](ARCHITECTURE.md)，这里只列调试相关的部分：

```
internal/debug/            调试核心：fgldb 协议驱动(session.go)、会话管理(manager.go)、
                           REST+WS 服务(api.go)、配置(config.go)、报文日志(wslog.go)、
                           服务测试(wstest.go)、DB 探查(db.go)、协议正则(parser.go)、断点存储(bpsstore.go)
internal/cli/debug/        命令行：cobra 的 tt debug 命令组 + 后台守护(servebg_*)
internal/host/             ★ 远程服务器共享层（与字典侧共用）：SSH/PTY、终端行解析、
                           登录区动态路径探测、DB 环境探测、环境模型、源码镜像
internal/config/           ★ 统一配置层：路径解析 / 读写 / schema / 旧配置迁移
internal/web/              统一 HTTP 服务：页面 + 共享 /api/*（配置读写、派生状态、字典类动作）
web/app/                   前端(React 18 + Vite 6 + Monaco + Tailwind v4 + zustand)
```
