---
name: tt-dev
description: 做一次 T100 开发、或改 .tzc 代码包里的 FGL 程序时使用。一个任务一个自包含工作区，主线 open → brief → find → at → 改 → verify → apply；工作区建好后读命令全离线。开工前必须先从用户拿到修改标记 <单号>#<项次>、工号、环境三样，缺一不动手。
---

# tt-dev：一次开发 = 一个工作区

`tt dev` 把"一个程序的一个任务"落成一个自包含目录：材料、索引、可编辑视图、包、备注都在里面。
之后所有命令都在这个目录里跑，只改 `ws/prog.full.4gl`，只有一条写回路径 `apply`。

记住主线：**open → brief → find → at → 改 → verify → apply →（交付前）update --check**。

| 场景 | 入口 |
|---|---|
| 做一次开发（推荐） | `tt dev open …`，用本文主线 |
| 拿到一个 `.tzc`、没有工作区 | 直通层 `tt dev tzc …`，见 §12 |
| `.tzs` 表单包 | 另一条管线 `tt dev tzs …`；拿错入口会退 2 |

## 0. 开工前：标记、工号、环境，三样都要用户给

**读代码、定位阶段不需要**；**改正文之前三样必须齐**。

| 要什么 | 规则 |
|---|---|
| **修改标记** `<单号>#<项次>`（如 `A12345#01`） | 用户给就原样用，**不要自己编**；用户明确拒绝时跑 `tt dev tzc uuid` 生成 16 位 id 当标记，并**告知用户用了哪个 id** |
| **工号**（如 `12345`） | 用户给；拒绝时写 `agent` |
| **环境** `<环境>` 或 `<环境>@<企业>` | 也必须用户明确给出。不知道有什么：`tt dev open --at-list`（纯本地，不联网）列出来再问用户。**禁止自己挑环境，尤其禁止拿生产环境当任务基线** |

拿不到就问，不要边问边改。

## 1. open：建工作区（唯一需要在命令行点环境的一步）

```powershell
tt dev open <程序> --at <环境>        # 环境一律点名：<环境> 或 <环境>@<企业>
cd <程序>                                  # 之后所有命令都不再需要 --at
```

常用开关：`--task <任务名>`（拼进默认目录 `<程序>-<任务名>` 并记进 task.json）、
`--dir <目录>`、`--identity c|s`（包身份，缺省从 `.tap` 推断）、
`--only <点名>…`（只把这些点渲染成可编辑）、`--no-mdl`（不复制模板切片）、
`--form`（连表单 `.42f` 一起拉，体积大）。

产出目录（**整个目录就是交接物**，别只拷其中几项）：

```
task.json           身份与历史（prog / env@ent / task / identity / 事件）；实时状态一律现算
materials/          蓝图 .tab、规格 .tsd/.4fd、加点 .tap、骨架 .tgl、产物 .4gl + manifest 指纹
mdl/slice/          模板切片（切片脱版检测用；--no-mdl 则没有）
index/link.json     蓝图 ↔ 代码索引（页签/字段 → 加点 → 行号 → 切片）
ws/                 可编辑视图：prog.full.4gl（唯一编辑文件）、.tdev/base.full.4gl（基线）、manifest.json
pkg/<程序>(<身份>).tzc   可直接交给设计器的产物（identity=c 时是 (c).tzc）
notes/              人的备注，工具不碰
```

- 一次 open 约 30–40 s（11.8 MB 量级）；建好之后读命令 50–120 ms。
- **没有被设计器接管的程序**（服务器上没有 `.tap`/`.tgl`）不能建工作区：退 4，
  文案会给出只读出路 `tt gen fetch … && tt gen map …`。身份 `s` 的标准件只要有这两件材料就能 open，
  产物是 `<程序>(s).tzc`。
- 失败不留半套：中途出错会删掉本次创建的目录。
- 工作区被复制/移动后，`apply` 会把源包路径自动绑回当前工作区（stderr 会明说），
  不要手工改 `ws/manifest.json`。

## 2. 先读再改：五条离线命令

`status / brief / find / at / search` 建好工作区后**完全离线**（不连服务器），地址是当前目录或 `--dir`。
这些工作区动词默认输出 JSON（脚本/agent 用），给人看加 `--human`；`tt dev tzc …` 那层反过来 —— 默认人读，要 JSON 加 `--json`。

| 我想知道 | 命令 | 说明 |
|---|---|---|
| 这是什么程序/有哪些页签/客制逻辑在哪/有什么风险 | `tt dev brief` | 判型、页序、page_id、screen record、客制点（按行数降序，默认前 15：`--top`）、warn/error、脱版、`.4gl` 是否比 `.tap` 旧 |
| 要改哪个加点 | `tt dev find --intent check --field 刻号` | 意图 × 字段。`--field` 收列名或**画面中文名**；`--page` 收 page_id（73）或页签序号（12）；`--limit` 调条数 |
| 这一行属于哪个点（编译报错/断点/review 的一行） | `tt dev at 44107` | 行号只认 **`.4gl` 产物行号**（编译报错那种），不是视图行号、也不是 `.tap` 行号；`--context` 带上下文 |
| 这个点的内容（点名或字段名） | `tt dev at xmdduc003` | 缺省看**包内 `.tap`**（apply 后这才是真相，行号是 `.tap` 行号）；`--from 4gl` 看服务器产物（产物行号）；模板洞没有 `.tap` 声明时会退到产物并明说 |
| 这段代码在哪、是谁生成的 | `tt dev search "SET("` | 每条命中标出加点/客制还是模板/section/切片；`--scope all`（缺省）在工作区里优先查包内 `.tap` |
| 当前工作区什么状态 | `tt dev status` | 材料/索引/视图/包四件事现算：脏不脏、包有没有 apply 过、能不能交付 |
| 是不是孤儿（写了却不生效） | `tt dev search <名字>` 或 `tt dev at <点名>` | 有 `.tap` 正文、骨架没槽位的点重产不保留，工具会明说 |

**find 的 8 个意图 → 该改哪个段位**（改错段位 verify/apply 照样过，但运行行为是错的）：

| 意图 | 典型段位 |
|---|---|
| `check` 栏位检核 | `input.a`（AFTER FIELD）+ 模板 a17/a05 |
| `lookup` 开窗/带值 | `construct.c` / `browser_fill` + a07/a08 |
| `default` 预带默认值 | `input.b`（BEFORE FIELD）/ `b_fill` / `init` |
| `save` 保存前 | `insert_b` / `update_b` / `key_*_b` |
| `display` 显示刷新 | `show` / `detail_show` / `ui_dialog` |
| `init` 初始化 | `init` / `global` / `construct` |
| `query` 查询取数 | `query` / sql 段 / `filter` |
| `sign` 签核复核 | `signature` / `signing` / `state_change` |

`find` 的输出只给点名 + 行号 + 客制/模板，**没有逐点的段位列**（只有一行意图提示，如"栏位检核改 input.a"）：
段位看点名后缀（`input.a`/`b`/`c`/`g`、`construct.*`）或再跑 `tt dev at <点名>`；
`find` 里的 `L…` 是 **`.4gl` 产物行号**，不是视图行号。
**模板洞**：`find` 可能返回 `Kind="模板"` 的点 —— 骨架里有槽位、包内 `.tap` 里没有声明
（例如某程序第 73 页的刻号栏位 `input.a.page73.xmdduc003`，视图里是 0 行空点）。
这类点在 `find`/`points` 里是 **`Kind=模板` + 0 行 = 可填的洞**（不是"不能改"）；填正文用 `tt dev tzc set ws <点名> --stdin`（或 `--file`，先 `--dry-run` 预览）：
填完 `show` 确认 → 成为客制点 → `apply` 才进包。`tt dev at <点名>` 对它们会退到产物并说明，
不是"没逻辑、不能改"。

**行号有三套，别混**（互不通用，工具自己也只会提示"不通用"）：

| 行号 | 用在哪 | 怎么拿 |
|---|---|---|
| **视图行号** | 改代码（`ws/prog.full.4gl`） | `tt dev tzc show ws <点名>`（非空点打印 `L<行号>`）；0 行模板洞不打印行号，用 `points ws --json` 的 `line0/line1` 或 `show --json` 的 `region.line0` |
| **`.4gl` 产物行号** | 编译报错、断点、`tt dev at <行号>`、`tt dev search` 的 4gl 命中、`find` 输出的 `L…` | `tt dev at <行号>` / `tt dev search` |
| **包内 `.tap` 行号** | `tt dev at <点名>`（默认为包内声明）；`at --from pkg` 的 `line0` | `tt dev at <点名> --from pkg` |

- `at`/`search` **都不给视图行号**；把视图行号喂给 `tt dev at <行号>` 会落到别的段（实测 50678 落到 `lock_b`）。
- 视图比产物长（每块区域加围栏、按包重排），**长度随工作区变，不要背数字**：以当前工作区 `points ws --json` 为准。
- 改行用 `tt dev tzc patch ws <点名> --line N --text …`（N 是视图行号）。

## 3. 改哪里、怎么安全地改

**只改 `ws/prog.full.4gl` 里 `EDITABLE` 的围栏正文。** 三条红线：

1. 围栏行（`{//@tdev:begin …}` / `{//@tdev:end …}`）是坐标，改了整份映射作废（gate1 直接拒）。
2. 结构行、围栏外字节（文件头尾、区段间空行、缩进）必须逐字节不变 ——
   **不要用格式化工具重排整个文件**。
3. 不要手改 `.tap`、不要绕开工具改 zip、不要编辑 `ws/.tdev/base.full.4gl`（那是 gate1 的比对基线，
   改它等于伪造验证）。

视图动辄数万行，**不要整份读/整份改**（行数随工作区变，以 `points --json` 为准）。区域级动词把"区域"变成一等对象：

```powershell
tt dev tzc points ws --editable                 # 列能改的区域（点名/行数/拒绝码）
tt dev tzc points ws --match xmdduc003          # 按点名找（含 0 行的模板洞）
tt dev tzc show   ws input.a.page73.xmdduc001   # 只看这一块正文（非空点带视图行号，如 L50678）
tt dev tzc find   ws input.a.page73.xmdduc001 "IF" --context 3
tt dev tzc patch  ws input.a.page73.xmdduc001 --line 50678 --text "   ..." --dry-run
tt dev tzc set    ws input.a.page73.xmdduc003 --stdin   # 整段替换（空点/小区域；大区域用 patch）
```

它们**只改 workspace，不写包**；落包仍走 `verify` / `apply`。
`patch` 与 `set` 都有 `--dry-run`，正式写之前先跑一遍；围栏行绝不能被 `--old`/`--new` 碰到。
输出：区域级 verbs **默认人读表格**，要 JSON 加 `--json`；它们不认全局 `--human`/`--format`（加了会退 2）。
0 行空点（模板洞）的 `show` 只会告诉你"0 行"，视图行号从 `points --json` 的 `line0/line1` 取。

**结构变化**不走手改，且注意地址是 **`ws` 这个 tzc 工作区**（在任务工作区根直接敲会报"当前目录不是 tdev 工作区"）：

- 改函数/对话/报表名：`tt dev tzc rename ws <旧名> <新名>`（围栏 `fn` + 签名行 + 描述块三处原子同步）
- 加新函数：`tt dev tzc newfn ws --type FUNCTION|DIALOG|REPORT --name <名>`

### 写业务代码时的三个样板（省得从零试）

**同类检核先抄现成的**：`tt dev search "cl_err" --human` 找本程序的报错写法，
再用 `tt dev tzc find ws <同页参考点> "cl_err" --context 8` 看上下文。

判空报错的最小写法（字段/数组变量名照同页其他栏位抄，别自己造）：

```bdl
IF cl_null(<字段>) THEN
   INITIALIZE g_errparam TO NULL
   LET g_errparam.code = '<消息编号>'
   LET g_errparam.popup = FALSE
   CALL cl_err()
   NEXT FIELD CURRENT
END IF
```

**消息编号不能用文案代替**：`g_errparam.code` 必须是字典里的消息号。查法：
`tt dict msg --text <关键字> --at <环境>`（缺省查 zh_CN；提示只有 zh_TW 时加 `--lang zh_TW`），
`tt dict msg <编号>` 看详情。字典里没有合适消息时如实告诉用户（新增消息属字典侧数据，
工作区/CLI 改不了），不要自己编号。
**被拒时先看 `deny=` 码，不是所有拒绝都能 unlock**：

| `deny` 码 | 含义 | unlock 有用吗 |
|---|---|---|
| `section-locked` | 框架区段未解锁 | ✅ 只有这一条走 unlock |
| `section-anchor` / `section-readonly` | 锚点区段 / 只读区段 | ❌ 解锁了也还是只读 |
| `readonly-attr` | 点属性 `readonly=Y` | ❌ |
| `env-edit-env` | 插入点 `edit=` 与当前环境不匹配（如标准点 `edit="s"` 在客制环境） | ❌ 这是点级门禁，不是区段门 |
| `cite-std` / `ind-fun-mismatch` / `industry-memo` / `topstd-mode` / `login-topstd` | 引用标准 / 行业别 / 顶层标准 / 登录用户等门禁 | ❌ |

**unlock 的流程（单向、不可逆、有代价）**：只有 `deny=section-locked` 才考虑。

1. 先跑**不带 `--yes`** 的 `tt dev unlock`：工具打印代价后以退 4 结束，**不产生任何变更**；
2. 把"解开后规格调整不再自动生成程序代码"这个代价明确告诉用户；
3. **用户明确确认后**才加 `--yes`。AI **绝不自行加 `--yes`**。
   锚点区段即使解锁也依然只读。

## 4. 摘要与修改标记

拿到标记/工号后，全文用同一个标记。

**摘要在客制 memo 点里写，不要写 `global.memo`**：

- `global.memo` 是标准点（`edit="s"`、`src="s"`）；在客制环境（`env=c`）下是
  `deny=env-edit-env`，**verify/apply 必拒，unlock 也解不开点**。
- 客制摘要写 `global.memo_customerization`（`src="c"`、`edit="c"`，可编辑）；
  行业件摘要写 `global.memo_industry`。
- 用 `tt dev tzc show ws global.memo_customerization` 先看现状再改。

摘要行格式（加在 memo 点正文里）：

```
#<单号>#<项次> <作业日> By <工号> <一句说明>
```

例：`#A12345#01 20261028 By 12345 新增客户信用额度校验`；用 uuid 兜底时 `#3f9a1c07b2d84e56 20261028 By agent …`。

本次涉及的代码改动用**同一标记**成对包裹：`#A12345#01 -s` … `#A12345#01 -e`。

> 同一 memo 里可能已有旧格式行（如 `#EE90A83C7FF1F3A5 2026/11/21 By agent …`）：
> 不要改旧行，本次按上面格式**追加一行**即可。
写完自检：memo 摘要行与所有 `-s/-e` 块用的是同一个标记。

## 5. verify → apply：验证与写回

```powershell
tt dev verify                 # 只审计，不写盘（gate1 + gate2）
tt dev verify --strict        # warn 也算失败（CI 用）
tt dev apply --dry-run        # 先看逐条目 diff，不落盘
tt dev apply                  # gate1+gate2+gate3 全过才原子写包；这是唯一会改 .tzc 的命令
```

- **apply 没有二次确认**（`--yes` 只是为兼容旧脚本保留、实际被忽略）。
  要写区段（解开框架）必须先 `tt dev unlock --yes`，否则 gate1 直接退 4。
- **apply 不编译 BDL**：它保证围栏/结构/装载模拟通过；语法错要到设计器 F7 才暴露。
- 包内 `.4gl` 是服务器 build 产物，apply 不会改它；上传后由设计器 F7 按新 `.tap` 重生。
- `at --from pkg` 的信封注意双层：外层 `source` 恒为 `workspace`，真实来源在 `data.source`（`pkg`），
  每个点的 `pkg` 字段给出包路径。
- **apply 成功后视图会被规范化重写**（围栏字段/行尾/行号可能变）。继续改第二处之前要重新定位，
  不要沿用旧行号；复核"改动进包没有"用 `tt dev at <点名> --from pkg`。
- **什么算"可交付"**：包必须是 **apply 过的**。刚 `open` 出来、`status` 显示"包未 apply"时，
  包内容就是原始基线（与服务器一样），交出去等于没做；判据看 `status`：人读"已 apply"，
  JSON 里 `ws.pkgSha` 与 `pkg.sha256` **不相等**且 `pkg.matchesBase=true`。`apply` 之后才谈交付。
- 交给设计器之前跑 `tt dev update --check` 确认服务器材料没变。

**update 是四条规则（第 0 条先在 `--check` 之外拦住不一致）**：

0. 包与视图基线不一致（包被换过/工作区复制来的）→ **拒绝**退 4，先 `relink`；`--check` 例外。
1. 视图有未落盘改动 → **拒绝**（先 `apply`；`--force` 只刷材料与索引，不动视图/包）。
   **`verify` 不会清 dirty**（它只审计），别靠 verify 解锁 update。
2. 包里已有 apply 过的改动 → 只更新材料与索引（成果保留，视图落后于服务器，别再在上面改）。
3. 视图干净且包没被 apply 过 → 重新打包 + 重新导出视图，整体对齐到最新材料。

`update --check` **只要命令成功就退 0**：材料过期与否看结果里的 `stale`（人读显示"材料已过期/一致"）。
过期就先 `update`（或重新 open）再交付，不能拿旧材料直接交付。

## 6. 危险动作与出路

| 现象 | 处理 |
|---|---|
| `apply` 退 4，"源包自 export 之后已被改动" | 包被替换/工作区是复制来的：先 `tt dev status --human` 看丢了什么、能 `apply` 就先 apply 保存改动；确认丢弃后 `tt dev relink`（**未落盘改动会丢**），或把基线包放回来 |
| `relink` 被拒（dirty） | 默认拒脏工作区。`--force` 才丢改动 —— **执行前必须让用户知情确认** |
| `update` 退 4，视图有未落盘改动 | 先 `apply`（verify 不清 dirty）；确认不要视图/包才 `--force` |
| `status` 说"与视图基线不一致（不可交付）" | 包被换过，`apply/verify` 都会拒；先 `relink` |
| 退 2，工作区 schema 不是本版认的 | 换新版 tt，或重新 `open`；不要手改 task.json 的 schema |
| 被拒且 `deny=section-locked` | 只有这条能 unlock：见 §3 的三步，必须人确认 |
| 被拒且 `deny=` 是别的 | 先看码（§3 表），不是 unlock 问题，别乱试 |
| `at` 行号对不上 | `at` 是 `.4gl` 产物行号；视图行号用 `tt dev tzc show ws <点名>`（§2） |
| 工作区在别的目录 | 任务层命令加 `--dir <工作区>`；tzc 区域动词用位置参数 `ws` |

## 7. 退出码

| 码 | 含义 |
|---|---|
| 0 | 成功（含 `update --check`：过期与否看结果，不看退出码） |
| 2 | 用法/身份错（不是工作区、目录非空、程序名/任务名非法、环境不存在、工作区版本不兼容…） |
| 3 | 校验失败（verify/apply 的 gate 发现 error 级问题） |
| 4 | 未找到（程序/材料/点位）或写入被拒（dirty、包与基线不一致、缺 `--yes`、权限） |
| 5 | IO/网络失败（含"当前目录不是 tdev 工作区"这类地址错） |

`.tzs` 那条线没有 3，另有一个 1（引擎内部错）；`tzs formcheck` 有 6（有确定差异）。

## 8. 输出与报错怎么读

- 默认 **JSON 到 stdout**；`--human` 才是人读表格；进度/提醒/警告一律 stderr。
- 失败也是 **stdout 一行 JSON**，`ok:false`。键名有两套，别取错：
  - 工作区动词（status/brief/find/at/apply/verify/unlock…）：`ok` / `code` / `exitCode` / `error`
  - tzc 直通动词（`tt dev tzc status … --json`）：`ok` / `exit_code` / `error`
- `tt dev status` 的关键 JSON 字段：`materials.{fetchedAt,missing}`、`index.{fresh,schema}`、
  `ws.{dirty,sectionState,regions,pkgSha}`、`pkg.{present,sha256,modified,matchesBase}`。
  有没有 apply 过：看 `ws.pkgSha` 与 `pkg.sha256` 是否相等（相等=未 apply）或人读的"已 apply"。
- `verify`/`apply` 被拒会打位置：`prog.full.4gl:<行号>` + 该行内容 + `deny=<原因>`；
  基线侧问题给 `ws/.tdev/base.full.4gl:<行号>`；JSON 里每条发现带 `file`/`line`/`snippet`。

## 9. 交接给别人/另一个 agent

- 交**整个工作区目录**（材料、索引、视图、包、notes 都在里面），不要只拷 `pkg/`。
- 接手第一步：`tt dev status --human` 看四态 → `tt dev brief` 看全貌 → 再决定在哪个点继续。
- 在新机器/新目录上**不要对着旧视图直接 `apply`**：先 `status`，需要时 `relink`。
- 交付前在 `notes/` 写清：改了什么、验证到哪一步、还剩什么没做。`notes/` 是**唯一允许直接用编辑器/重定向写**的地方（工具不碰它，也没有写它的命令）。

## 10. 反模式

1. 手改 `.tap` / 手改 zip / 用文本编辑器改包 —— 唯一写回路径是 `apply`。
2. 格式化整个 `prog.full.4gl` 或重排围栏外字节 —— gate1 必拒。
3. 编辑 `ws/.tdev/base.full.4gl` 或 `ws/manifest.json` 来"让校验通过"。
4. 交付前不跑 `verify` 就直接 `apply`：apply 自己会跑三道闸门，但 verify 能更早、更便宜地把问题列全。
5. AI 自己加 `--yes` 解锁框架；或没拿到标记/工号/环境就开始改正文。
6. 把真实客户包或线上环境当试验品：open 只读拉材料，实验一律用新目录/副本。
7. 忽略 `status` 的"不可交付/不一致"继续改 —— 后面 apply 一定被拒。
8. 摘要写 `global.memo`（标准点）然后怪工具拒写 —— 见 §4。

## 11. 一次完整会话（示意；尖括号换成你自己的值）

```
tt dev open <程序> --at <环境>              # 产出 pkg/<程序>(c).tzc
cd <程序>
tt dev brief --human                       # 判型/页签/客制点/风险；页签序号→page_id 的对应看这里
tt dev find --intent check --field <字段>   # 列候选加点；包内没有声明的就是模板洞
tt dev tzc points ws --match <字段>         # 确认 0 行空点、可编辑
tt dev tzc show ws <点名>                   # 看这块正文；0 行洞不打印行号（用 points --json）
# set/patch 写正文（先 --dry-run 预览），按 §4 加 memo 摘要与 -s/-e 标记
tt dev verify                              # gate1+gate2
tt dev apply                               # 写回 pkg/<程序>(c).tzc
tt dev tzc points ws --match <字段>         # apply 后视图被规范化，行号可能位移：重新拿
tt dev at <点名> --from pkg                 # 复核改动真的进包了
tt dev update --check                      # 交付前确认服务器材料没变
```

> 页签序号→page_id→screen record 的对应、客制点数/行数、切片脱版，都以你自己工作区的
> `tt dev brief` 输出为准，不要照搬别人的数字。

## 12. 拿到现成 `.tzc`、没有工作区时（直通层）

直通层 = `tt dev tzc`，只有一条链：**export 读包 → 改工作区 → verify → apply 写包**。
它**不提供** brief/find/at/search 那套蓝图索引；要"先看懂再改"还是走 §1 的 open。

```powershell
tt dev tzc export "D:\pkg\capt110(c).tzc"   # 默认导出到 <包目录>/<程序名>-ws（-o 改；身份后缀 (c)/(s) 会去掉）
cd D:\pkg\capt110-ws                        # 之后省略 <dir>
tt dev tzc status                           # 改了什么、在第几行、会不会被拒
tt dev tzc verify                           # 不写盘预检
tt dev tzc apply                            # 校验通过才写回：**默认原地覆盖源包**；-o 另写一份
```

- tzc 工作区 = `prog.full.4gl`（唯一编辑文件）+ `manifest.json`（源包身份与 sha256）+ `snapshot/`（逐条目快照）
  + `.tdev/`（比对基线 `base.full.4gl`、上一版包）+ `.git/`（每次 apply 一次提交）。
- 区域级 verbs（`points/show/find/patch/set`）用法同 §3，只是地址换成这个工作区路径。
- 围栏红线、`deny=` 码、摘要标记、退出码、报错怎么读：全部同 §3–§8，直通层没有另一套规则。
- 源包在 export 之后被别人动过：重新 `export`（工作区里未落盘改动会丢），或先 `apply` 保住能保的。
- 内置自检：`tt dev tzc selftest`（31 项合成包对抗用例，不需要真实语料）。
- 拿不准某个点/区段能不能改：`tt dev tzc points <dir> --match <点名>` 直接给 `deny` 码与原因。

## 13. 其余看帮助

每个开关与默认值都在命令行里：`tt dev --help`、`tt dev <动词> --help`、
`tt dev tzc --help`、`tt dev tzc <动词> --help`、`tt dev tzs --help`。
本文与帮助冲突时：**命令行为以 `--help`/实测为准；退出码以本文 §7 为准**
（`tt dev --help` 的码表漏了 3，但 verify/apply 确实会返回 3）。
