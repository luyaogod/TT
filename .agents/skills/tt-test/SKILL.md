---
name: tt-test
description: 在真实 T100 服务器上跑 TT 的集成测试与链路验证。需要"必须真连 T100"（真 SSH / 真 fgldb / 真库）才能回答问题时触发：验证某环境的调试链路是否通、把待上线的改动拿到真机上复现、跑通了再收口。触发词：真机测试、连 T100 跑一遍、集成测试、验证链路、probe。本技能先要求用户明确授权才能读本机环境配置，并给出取配置、跑测试、收口的真实命令与红线。
---

# 真机测试 — 跑"必须真连 T100"的测试

## 什么时候用

要用**真实 T100 服务器**（真 SSH、真 fgldb、真库）回答问题时才用它。能靠默认档回答的问题不要进这条线。

- 默认档 `go test ./... -count=1` 是**离线**的：真 SSH / 真 fgldb / 真库那几层在代码里没有测试落点 ——
  判据见 `internal/debug/README.md` 的「判据」（`session.go` 主流程要真 SSH PTY + 真 fgldb 交互，
  故意不覆盖；`api.go` 的快乐路径要 `*host.SSHConn`，其 `cli` 是私有字段，测试只注入必然失败的拨号）。
- 全仓 38 处跳过登记在 `internal/testkit/skipledger_test.go` 的 `skipLedger`（L3 语料 / L4 真引擎 /
  L0-环境 / 前提），**没有一处以"有台真 T100"为门槛**。所以现在**没有**哪条已入库的测试会因为
  "给了环境配置"就去连真机；想让测试连真机是**写新测试**，不是设一个开关。
- 所以本技能管三件事：**取环境配置**、**用已在命令面上的真机入口验链路**、**跑完收口**。

## 前提：先向用户取授权

**动手前必须先拿到用户明确的授权**，才能读本机现有的环境配置。

**"授权"在这份技能里只意味着一句话**：用户明确说"**可以用本机这个环境跑**"，并**点出是哪个环境**
（`config.json` 里 `hosts.sshs[].name` 的那个名字）。用户没说环境名就是没授权 —— 环境一律点名，
没有"默认环境"可回落（`internal/config/schema.go:92` 的 `Hosts`、`internal/cli/root.go:95` 的 `--env`）。

授权之后允许读的，就是下面这两条命令；**不许**再往外扩（不 dump 整份配置、不加 `--raw`）：

```bash
./tt.exe config path --json   # 配置文件的绝对路径：{"path","default","portable","exists"}
                              # internal/cli/config.go:40-67
./tt.exe env list --json      # 该文件里的环境清单：名称/主机/端口/账号/区域/库
                              # internal/cli/env.go:36-41、129-171
./tt.exe env show <环境名>     # 单个环境详情（口令打码）internal/cli/env.go:173-226
```

两条都自带打码：`env views` 的结构体里没有口令字段（`internal/cli/env.go:96-105`），
`env show` 打的是 `internal/cli/env.go:228-234` 的 `mask`。**输出可以照抄；`--raw` 不可以。**

没拿到授权时：不要读、不要连，**直接把"要用户授权"这件事说出来**。

## 没有现成配置时：要用户提供，不许猜

`tt config path` 说"配置文件尚未创建"、或 `config.json validate` 报 `hosts.sshs 为空`，就是没有现成配置。

**这时要用户给，缺一不可**（字段形状 `internal/config/schema.go:30-90`、`internal/dbconfig/dbconfig.go:41-61`）：

| 要给什么 | 字段 | 不给会怎样 |
|---|---|---|
| 环境名 | `hosts.sshs[].name` | 命令报错并列出可用环境；**绝不替你挑一个** |
| SSH 主机 / 端口 | `.host` / `.port`（0 → 22，`schema.go:127-134`） | 无从连起 |
| SSH 账号 | `.user` | 同上 |
| SSH 口令或密钥 | `.password` | 同上 |
| 登录区域 | `.zone`（如 31 开发 / 36 正式） | 登录后取不到区域环境 |
| 库（要连库才需要） | `.db.type`（`oracle`/`kingbase`）+ `.db.host` + `.db.port` + `.db.service` 或 `.db.database` + `.db.accounts[]` | `tt debug sql` / `ents` / `db` 不可用；`probe` 与 `start` 仍可用 |
| 企业编号 ENT | 每次命令上给 `--ent` | 用错账号查数，**症状是"查不到数据"**，与"数据不存在"一样 |

**放哪**：`internal/testenv` 认的是**仓库根的 `config.local.json`**（`internal/testenv/testenv.go:14`、
`README.md:40`），被 `.gitignore:19` 忽略。但**要如实知道它的边界**（`internal/testenv/README.md:16`）：
它只解析四个**机器路径**字段——`corpusRoot` / `engineExe` / `workspace` / `designerDir`
（`internal/testenv/testenv.go:20-25`），**它不读 `hosts`、也刻意不碰口令**（`README.md:16-18`）。
所以：

- **首选**：让用户把环境配进本机那份 `config.json`（设置页 `tt serve` 是官方入口），
  或用户给一份 JSON 文件路径，测试命令用 `--config <路径>` / `TT_CONFIG=<路径>` 指过去 ——
  优先级与"显式路径就是答案"见 `internal/config/paths.go:104-169`。
- 若用户坚持把环境写进 `config.local.json`：那里**只有一个家**是四个机器路径；往里加 `hosts`
  等于开出第二份配置，`testenv` 不会读它（`json.Unmarshal` 忽略未知键）。真要这么用，**待核实**的
  就是"读它的那段测试代码" —— 那一行不在这份技能里。

**明确不许做的事**：

- 不猜主机名/IP/端口/区域/账号。
- 不扫内网、不试常见端口、不遍历网段。
- 不拿**别的环境**的配置顶替（"反正都是 T100"——那是把测试打到另一台机/另一个企业上，
  输出看起来完全正常，结论会彻底反向；`internal/config/schema.go:92-100`）。
- 不把口令写进任何测试文件、命令行参数、脚本或环境变量持久化。

## 怎么跑

### 1. 先验链路（真机入口，已在命令面上）

```bash
./tt.exe debug probe -m <模块> -p <作业> --env <环境名>
# 一条命令跑完：登录 → 起会话 → 下断点 → run → where → next → 求值 → 退出
# 实现 internal/cli/debug/debug.go:28-142；参数见 :277-279
```

- **它会真的登录、真的启动那个作业**。所以这是"拿授权换来的"动作里最重的一个：目标机由用户点名，
  不许自己换一个（`internal/cli/debug/debug.go:40-52`）。
- 不点名环境时 `probe` **直接拒绝并指向 `--env`**，不是"挑了台空的去连"
  （`internal/cli/debug/probe_test.go:57-74`）。

跑通了，链路就是通的。要停站到具体位置再取证，转 [`skills/tt-debug/SKILL.md`](../../../skills/tt-debug/SKILL.md)
的人机交接一节（`tt debug start` / `exec` / `wait` / `why`）。

### 2. 真机上的只读取数（可选）

```bash
./tt.exe debug ents --env <环境名> --json      # 有哪些企业(ENT)、各自账号（快照 10 分钟）
./tt.exe debug sql "<单条 SELECT>" --ent <ENT> # 只读 SQL：输出 CSV，# 头回显环境/企业/账号
```

- `sql` 只放行**单条** SELECT / WITH，**最多回 200 行**（`internal/safesql/safesql.go:25`）。
- 用户没授权时不要用；`hosts.sshs[].db.readonlySql` 为假的环境直接关掉了这条能力
  （`internal/dbconfig/dbconfig.go:61`）。

### 3. 真机跑 `go test`（当前可用形态）

```bash
make test        # 全量 go test ./... -count=1（约 1 分钟，离线）  Makefile:20-21
make skips       # 跑了多少、跳了多少（会真跑一遍全量）          Makefile:54-55

go build -o tt.exe .                                   # Makefile:17-18
go test ./internal/<pkg> -count=1 -v                   # 单个包，看真实执行
go test ./internal/<pkg> -count=1 -run <用例名> -v      # 钉一条（-run 的用法见 Makefile:48、52）
```

`-count=1` 是必须的：默认档的命令逐字用它，`go test` 的缓存会让"没跑"长得像"跑过了"。

给真机测试预留的开关只存在于**别的层**，`TT_CONFIG` 与它们无关：
`TDEV_CORPUS` / `TTZS_CORPUS` / `TTZS_EXE` / `TTZS_WS` / `TTZS_INSTALL` / `TTZS_PKG` / `TDEV_DEEP` /
`TTZS_DEEP` / `TTZS_E2E` / `TTZS_FNS` / `TT_FEP_RECORD_DIR`（`internal/testkit/skipledger_test.go:62-66`、
`internal/testenv/README.md:42-47`）。它们一个都不连 T100。

**要断言"真机上的行为"，得先有一位测试连真机**（见上文：现在一条都没有）。那是改代码的事：
只新增 `.agents/skills/tt-test/` 下的文件，**不改 `.go`、不改 `Makefile`、不改 `config.example.json`、不碰 `skills/`**。
新增跳过必须登记进 `skipLedger`，否则 `make test` 会红（`internal/testkit/skipledger_test.go:204-217`）。

## 跑完怎么收口

会话与窗口：

```bash
./tt.exe debug quit        # 结束调试会话，作业窗口随之关闭
./tt.exe debug start --help # start 的开关（--env / --ent / -m / --fglserver …）
./tt.exe serve --stop      # 停止后台服务；它会断开全部会话（internal/debug/manager.go:724-734）
```

留在本机的东西（都在**数据目录**，即 `config.json` 所在目录，`internal/config/paths.go:28`）：

| 残留 | 落点 | 何时清 |
|---|---|---|
| 源码镜像 `srccache/`、命令输出 `execlog/` | `<数据目录>/srccache`、`/execlog`（`internal/debug/srcmirror.go:90-92`） | 每轮 `debug start` 前自动清空（`srcmirror.go:121-126`） |
| 断点存档 `debug-bps/`、企业快照 `ents/` | `<数据目录>/debug-bps`、`/ents`（`internal/debug/bpsstore.go:4`、`ents.go:80`） | `tt cache clear`（清单见 `internal/config/cache.go:21-29`） |
| 长期库 `records/` | `<数据目录>/records/<环境>/<ENT>/<录制 ID>`（`internal/fep/library.go:21-31`） | **不在缓存清单里**，`tt cache clear` 不碰它；`tt fep rec rm <ID> --yes` |
| 服务状态与日志 | `.tt-serve.json` / `.tt-serve.log` | **刻意不是缓存**（`internal/config/cache.go:17-20`） |

- 清缓存**连 `ents/` 快照一起清**：下次问企业会慢一次（那是正常重查）。缓存目录名不在清单里时
  `CacheDir` 返回空串，不会悄悄建出清单外的新目录（`internal/config/cache.go:31-49`）。
- 服务器侧：`probe` / `start` **真的起过作业**、库里有行锁、用户 GDC 上会弹过窗口 —— 这些不随本机收口
  消失。停站过夜会导致行锁长期被占，有一条看门狗到点自动放行（默认 1800 秒，
  `internal/config/schema.go:20` 与 `internal/debug/README.md:40-41`）。
- 测试配置本身不落任何东西；`config.local.json` 在不在都不影响跑完的状态。

## 红线

**⚡ 任何环境配置都不允许提交到仓库，这会导致严重的泄露事故。**

1. **绝不**把 `config.json` / `config.local.json` / 任何含明文口令的文件纳入版本库、贴进对话、
   写进日志、或作为测试固件提交。忽略规则已经写死在 `.gitignore:16-19`（`/config.json`、
   `/config.json.bak`、`/config.local.json`）与 `:44`（`/.agents/`）—— **不改它，也不要用 `git add -f` 绕它**。
2. **口令只在运行时从本机配置读**，不得复制进任何会被提交的文件（不得写进测试文件、`config.local.json`
   之外的夹具、`.md`、脚本、命令历史）。测试文件里出现的只能是 `<占位符>`。
3. **输出与错误里的口令必须打码**。打码只有一份实现：
   `internal/config/redact.go` —— `IsSecretKey`（:18-24，按键名 `password`/`passwd`/`pwd`）、
   `SecretPath`（:28-35，点分路径点到口令）、`RedactSecrets`（:39-62，深拷贝整棵树）、
   `RedactValue`（:67-77，给结构体载荷用）。给 `tt` 加任何"把配置吐出来"的输出，都要调它，
   不要自己写第二份（`internal/config/README.md:73-78`）。
   已知的两处例外要当心：`tt env show` 用的是 `internal/cli/env.go:228-234` 的 `mask`；
   **`GET /api/hosts` 返回明文口令**（设置页要就地编辑），任何 dump 或粘贴它的行为都是泄密
   （`internal/config/README.md:80-81`）。
4. **提交前自查**：`git status --porcelain` 里**不得**出现配置类文件。
   本文末尾「自查」一节的四条命令就是这件事的可重放版本。
5. **一旦发现口令已经进了 git 历史、或进了即将提交的东西里：先停下来报告用户**，
   不要自行"删掉就当没发生" —— 历史里的口令要按泄露事故处理，不是一次 revert。

按这条线的既有事实记着两处**同类风险**（不是新问题，但测试时别把它们放大）：

- 只读 SQL 与 `db` 探查会把 `账号/口令@//主机:端口/服务` 这类连接串拼进**远端命令行**，
  远端 `ps` 可见（`internal/debug/db.go:148-159`、`internal/debug/db.go:176-198`；
  [`skills/tt-debug/SKILL.md`](../../../skills/tt-debug/SKILL.md) 的只读 SQL 一节把它记为既有行为）。
  SQL 文本本身走 stdin、不拼进命令行（`internal/host/README.md:13`）。
- 报错时会带上被截断的远端输出（`internal/debug/db.go:215`、`:224`）—— 贴报错前后照样扫一遍口令。

## 自查（交之前逐条过）

```bash
git status --porcelain
git status --porcelain | Select-String -Pattern 'config\.json|config\.local\.json|\.bak$'
grep -nE '[0-9]{1,3}(\.[0-9]{1,3}){3}' .agents/skills/tt-test/SKILL.md
grep -nE 'password|passwd|pwd' .agents/skills/tt-test/SKILL.md
# 后两条**没有输出**才是对的：命中只允许是 `file:line` 引用与 `<占位符>` 这两类
```

1. `git status` 除了本次新增的文件之外干净？有没有配置类文件被纳入？（第二条命令没有输出 = 没有）
2. 全文里没有真实主机 / IP / 账号 / 口令：命中只允许是**源码文件:行号引用**与 `<占位符>`。
3. 技能里引用的每条命令、每个路径，都在源码里见过（本文的 `file:line` 就是出处）。
4. "没配置时要求用户提供"这条路径写在「没有现成配置时」一节，含必需字段清单与不猜/不扫/不顶替三条禁令。

## 待核实

- **`config.local.json` 是否承载 `hosts`**：`internal/testenv` 只认四个机器路径字段
  （`testenv.go:20-25`），**不读 `hosts`**。要用它装环境，得先有一段读它的测试代码；那段代码不在
  `internal/**` 里，所以这里按"首选指到 `config.json` 或 `--config <路径>`"写。
- **"在真机上跑 `go test`"的现成形态**：默认档没有一条测试连真机（判据见上），
  仓库现有 `make` 目标里也没有真 T100 的测试目标（`Makefile:10` 的目标清单）。
