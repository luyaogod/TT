# internal/cli — 根命令与统一命令组

命令树的装配处：把三个能力组挂上去，再加一组所有工具共用的东西
（环境、配置、缓存、服务、安装、版本）。**它也是唯一的接线处** —— 统一服务、
调试内核、字典动作三者的连线只发生在这里。

## 命令树

| 组 | 实现 | 说明 |
|---|---|---|
| `tt debug …` | [./debug](./debug/README.md) | 调试线 |
| `tt dev …` | [./dev](./dev/README.md) | 设计器包线（cobra 树挂动词；参数解析权在 tdev 自己的解析器，见其 README） |
| `tt dict …` | [./dict](./dict/README.md) | 字典线 |
| `tt drawio …` | [./drawio](./drawio/README.md) | 原型图线 |
| `tt session …` | [./session](./session) | 会话：建 / 列 / 看 / 改 ENT / 结束本轮 / 收掉 / 透传（`--record` 一步建会话并起录制） |
| `tt recording …` | [./recording](./recording) | 录制：起 / 停 / 列（正在录的、录过件的）/ 回放 / 留 / 删 / 清 / 索引 / 标失效 |
| `tt fep …` | [./fep](./fep/README.md) | 前端协议线（动作 / 界面快照 / 等待 / 源码 / 作业 / 无头前端） |
| `tt form …` | [./form](./form/README.md) | 出厂表单线（`.42f` 字段表：`ls` / `show` / `verify` / `fetch` / `cache`） |
| `tt env / config / cache / serve / install / version` | 本包 | 各组共用 |

各能力组各自独立子包、**互不 import**；共享的 CLI 上下文在
[./common](./common/README.md)（叶子包，所以不会与根命令成环）。旧的命令名仍作为别名被接受。

`tt form` 的用法（细节见 [./form](./form/README.md)）：

```
tt form ls     --env <环境> --module <模块>                     列出该模块下的表单名
tt form show   --env … --module … --form <表单名> [--json]      字段表：工号 colName / 表单内
                                                                编号 fieldId / 控件 / 几何 / 必填
tt form verify --env … --module … --form … --map <语料表.json>  与归档语料表逐条比 →
                                                                打印「可比 N 条 / 一致 N 条」
tt form fetch  --env … --module … [--force] [--json]            批量取进缓存：新取 / 命中缓存 /
                                                                失败（逐条列名+原因）
tt form cache  [--json]                                         缓存只读概览：多少张 / 多少字节 /
                                                                最旧·最新（不删任何东西）
```

取件按需单取、**只读**（`--dir <本地目录>` 可离线；`fetch` 也只是 list + get，绝不写服务器）；
`--json` 里 `show` 的键名冻结在 `tt-genero-formdef-contract.md` 的 C5.2，`fetch` / `cache`
用的是各自新增的信封（`failures[]` / `forms,bytes,oldest,newest`）。缓存落 `<数据目录>/forms/`
（`config.CacheSubdirs` 里已登记），清理走统一的 `tt cache clear`。

## 退出码：同一个数字在不同线上含义不同

**这张表是唯一的权威**（根帮助里也是它）：

```
tt debug      0 成功 / 1 用法或运行失败
tt dev tzc    0 成功 / 2 包格式或用法错 / 3 校验失败 / 4 拒绝写入 / 5 IO·环境失败
tt dev tzs    0 成功 / 1 引擎内部错 / 2 参数或环境错 / 4 设计器拒绝 / 5 传输或环境失败
              （没有 3；2 也包括"manifest 拉不到"）
tt dict       0 成功（含"查无结果"）/ 1 用法错 / 2 数据源错 / 3 缺表
tt session    0 成功 / 1 运行失败（IO·网络·环境）/ 2 用法错 / 3 无会话
              4 状态不符（会话没空 / 不是空闲 / 端口被占）/ 5 目标不存在 / 6 环境体检不过
tt recording  0 成功 / 1 运行失败 / 3 无会话登记（active、stop 找不到）/ 5 找不到该录制件
tt drawio     0 成功 / 1 用法或运行失败 / 2 输入错（spec 不合法、包读不出）/ 3 产物自查未通过
tt view       0 成功 / 1 运行失败（IO·抓图失败）/ 2 用法错 / 3 找不到目标窗口 / 4 产物自查未通过
tt form       0 一致 / 1 取件失败（FORM_UNKNOWN：取不到表单，或它不是合法 .42f）
              2 用法错或无基线（缺 --env/--module/--form；既没 --map 也没 --recording；
                语料里没有那一节窗口）/ 6 与语料逐条比对不一致（体检不过）
顶层          0 成功 / 1 用法错（未知命令、未知开关）
```

**包装脚本要按子命令辨认，别只看码**：`3` 在 `.tzc` 线是"校验失败"、在字典线是"缺表"、
在 drawio 线是"产物自查未通过"、在 `.tzs` 线根本不用；`4` 只在设计器两条线上有；
`6` 在 `tt session`/`tt recording start` 是"环境体检不过"（权限不一致）、在 `tt form` 是"字段表与语料不一致"。
各线的细化含义在各自 README 里。

**回放的身份闸（契约 A23/A24）**：`tt recording replay` 与 `tt fep replay`
在**发包之前**先过身份闸 —— 步骤没有 `colName`（该件早于身份上线）就退 `5`
（`FIELD_NOT_IN_SCOPE`）；有 `colName` 但解析不过则按四码退 `5/5/1/4`。
`--dry-run` 是唯一例外（不发包）：照旧展示，但每条无身份的步骤标
`无身份（真跑会被拒）`。旧件的处置只有两条：重录，或
`tt recording mark-stale <ID> [--reason <文字>]`（只改身份文件，列表随即显示
`不可回放（需重录）`）。**没有** `--skip-gate` / `--force` 这类跳过开关。

**前置体检与「一次说清」（[./fep](./fep/README.md) 实现，`tt recording start` 触发）**：
四项（数据目录可写 / 调试服务 / 调试会话 / 环境·作业）一次查完；它只检查、不录制、
不起会话、不写任何录制文件。`tt recording start` 在**起代理之前**跑这份体检：缺几项就列几项，每条带可直接复制的命令
（`tt debug serve`、`tt debug start <作业> --env <环境>`），已知的 `--prog`/`--env` 会填进去。
缺前置一律退 **2**（用法/环境不满足）—— 不新造码，也不占 `6`（那是"体检结论：权限不一致"）。

**无目标动作不受"目标互斥"约束（`tt fep`，契约 A28）**："目标 / `--field` 互斥"只对
**要目标**的动词生效 —— `tt fep do act`（click）、`tt fep do type`（input）及其变体；
`tt fep do key` 发的是 `KeyEvent`，**不带 `idRef`，没有目标可指**，所以不检查目标。
给它 `--field` 是**用法错（2）**：显式拒绝，不静默忽略。判据只有一处
（`internal/cli/fep/do_field.go` 的 `needsTarget`，与 `fieldKindFor` 同一张表），
调用点不许自己加条件。

## 统一的错误出口

根命令的 `reportError` 是唯一出口：

- **JSON 模式下失败信息走 stdout**（信封带稳定错误码与退出码），其余模式走 stderr 的散文
  并多打一行"提示"把修复建议拎出来。**这会改变"stderr 有内容 = 失败"这个既有判据。**
- 错误里没带环境信息时，自动补上**当前数据源**的 —— 连不上库时，"连的是哪个环境"比错误文本
  本身更难猜。
- **`tt dev` 不经过这里**：它自己是一套完整的 CLI（自己的用法、退出码、失败形状），
  返回的错误标记为"已报告"。否则同一个失败会在 stdout 上出现两份信封。

## 字典通道（`tt dict fieldtab`，缺省在线）

`tt dict fieldtab` 的字典有**两条通道**，开关 `--source online|local`，**缺省 `online`**：
在线 = 直查服务器字典表（只发 `select`），本地 = 本地 SQLite 镜像（只读）。
**不静默降级**：在线没环境/连不上 ⇒ 报错并给出下一步（`--env <环境名>` 或 `--source local`），
绝不偷用本地库；本地库不存在 ⇒ 报错并指向 `tt dict sync`（契约 A15/A31）。
用了哪条写在输出里：JSON 的 `dictSource`、人读的一行 `字典源: …`。细节见 [./dict](./dict/README.md)。

## 统一服务（`tt serve`）

**一个进程、一个端口、一份配置、一套页面**：

| 路径 | 内容 |
|---|---|
| `/debug/` | 调试工作台（含统一设置页） |
| `/debug/api/*` | 调试子系统的接口 |
| `/api/*` | 统一接口：环境与库配置、配置派生状态、安装、字典的长跑动作 |

**默认后台常驻、单实例**：把自身以"前台"模式重新拉起来后立即返回，终端不被占住；
`--foreground` 看实时日志，`--stop` 停止。运行状态与调试服务共用同一份状态文件，
所以控制端命令对着哪种装法都能用。

前台的装配顺序（每一处都有理由）：

1. **允许空环境**加载配置 —— 首次运行还没有任何环境，服务要能先起来；
   而命令行那条路径仍然是"没配环境就报错"。
2. **启动清理缓存**：只删超过保留期的，失败只报一行、不让服务起不来（清缓存是维护动作，
   不是启动前提）。`spill` 里那些文件恰恰是"刚才那条查询的完整结果"，启动即全清会把上一条
   命令刚告诉用户的东西删掉。
3. 建调试服务内核（内嵌前端产物也交给它，因为 `tt debug serve` 单独跑时也要能出页面）。
4. 定**同步目标缺省值**（当前目录下的本地字典库），并**同时**交给统一层与字典子系统 ——
   两边各算一次的话，设置页显示的路径可能不是实际写入的那个。
5. 装配统一服务：注入调试与字典两个 handler、配置重载器、停止回调，然后监听。
   停止路径有两条（Ctrl+C 与接口调用），统一收口到 `sync.Once`。

## 统一命令组

| 命令 | 作用 |
|---|---|
| `tt env` | 环境清单的列出 / 查看 / 切换 / 企业编号（**唯一的环境管理入口**） |
| `tt config` | 位置 / 查看 / 读写 / 迁移 / 校验；`--raw` 不过滤口令 |
| `tt cache` | 配置目录下**可再生**中间数据的查看与清理；"哪些算缓存"由配置层定义一次 |
| `tt install` | 安装 AI 技能到目标目录，或把 tt 加进用户 PATH |
| `tt version` | 版本；本地构建时回落到构建信息里的修订号（回答"这份二进制对应哪次提交"） |

## 判据

```bash
go test ./internal/cli/... -count=1
# 含：根帮助与 README 的漂移防线、安装面（技能复制 / PATH 合并）、
#     后台常驻的启动与停止、控制端命令的自动寻址
```

## 改动影响面

| 改什么 | 会波及 |
|---|---|
| 退出码表 | 所有脚本的判据；五条线各有一套语义，不能"顺手统一" |
| `reportError` | 一切命令的失败形状；JSON/散文两种模式的分界 |
| 服务装配顺序 | 首次运行体验（空环境）、缓存清理的边界、设置页显示与实际写入是否一致 |
| 统一命令组的行为 | 三条线共用，改一处影响全局 |

## 细节去哪

- 调试线 → [./debug](./debug/README.md)；字典线 → [./dict](./dict/README.md)
- 设计器包线 → [./dev](./dev/README.md)；前端协议线 → [./fep](./fep/README.md)
- 服务的 HTTP 层 → [../web/README.md](../web/README.md)；配置层 → [../config/README.md](../config/README.md)
