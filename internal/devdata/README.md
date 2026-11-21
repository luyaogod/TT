# internal/devdata — 开发资料（为 model 准备原材料）

> 一句话：**这不是"读代码用的镜像"，是给 model 的原材料。**
> 读侧是 [../file/mirror](../../internal/host/mirror.go)（`tt file mirror`）与
> [../dict](../../internal/dict)（`tt dict db sync`）；本包与它们**不共用目录、不共用代码、
> 不共用清理口径**。

## 当前状态（2026-11-21，产品决定）

- **镜像功能保留，但暂不启用**：`tt gen mirror` 与设置页的「同步共用资料」都还在，但**不做
  默认动作**，也不要求在 model 之前先把材料拉好。
- **查询默认在线**：需要从库里取的事实（例如某程序的当前版次 `dzaa002`）**先问在线库**，
  本地共用资料只作兜底（`programVersionAnchor`）。理由：镜像要镜哪些表、镜多大，等 model
  的能力边界确定之后再定（见 plan 的 ADR-08）。
- 因此首次使用 `tt gen` 时**不需要**先跑 mirror：`fetch` 能独立工作，只是模板版本锚点为空
  （记一条 warning，不判失败）。

## 为什么要有这一条线（ADR-01）

TT 原有的两组镜像能力都是为"**读**"服务的：查得快、看得全、过期就重拉。

model（生成器语义层，`tt gen model/why/expect`）要的不是全量源码，而是**按生成器口径组织
的原材料**：蓝图 `.tab`、中间档 `.tgl/.tap`、模板骨架 `mdl/code_*.template`、规格库
`DZ*_T`、数据字典。这类材料有三个特点：

1. **与某个程序强绑定** —— 所以按程序成套（ADR-04），不做散装下载；
2. **过期极快，且过期方式不显眼** —— 所以每份材料带 manifest 指纹与版本锚点；
3. **必须成套** —— `tab` 与 `tgl` 版本错配会得到"看起来对"的错误结论。

三者读侧都不提供（它不需要），所以本期新建一条**与读侧完全隔离**的资料线。

## 目录布局

```
<data>/devdata/                      <data> = config.Locations.DataDir()（config.json 旁）
└─ <env>/                            环境名，与 hosts.sshs 里的名字一致（可含中文）
   ├─ manifest.json                  M 侧清单：逐表行数 + 逐文件 sha256 + 版本锚点
   ├─ dict/<env>.db                  M：53 张 DZ*_T → SQLite（**与 <data>/dict/ 不是一份**）
   ├─ mdl/…                          M：$TOP/erp/mdl 整棵（保持相对路径）
   └─ prog/<prog>/                   D：一个程序一套
      ├─ manifest.json
      ├─ dzx/<prog>.{tab,tsd,tap,tgl,tgx,tgx2}
      ├─ src/<prog>.4gl
      ├─ inc/<prog>_{mask,rep}.4gl
      ├─ form/<prog>.42f             （--form 才有）
      └─ design/<prog>*.tzs          （--with-design 才有，本地工作区归档）
```

落点只由 [../config/devdata.go](../config/devdata.go) 的 `DevData*` 函数拼（与 dict 库
同一个纪律：调用方不许自己 `filepath.Join` 出这些路径）。配置项：顶层 `devdata.dir`，
环境级 `hosts.sshs[].devdataDir`（只覆盖根目录，环境目录名仍是环境名）。

## 材料角色表

**唯一事实来源是 [roles.go](roles.go)**（`MirrorTables` / `FetchRoles`）。
`roles_test.go` 会拿它跟**入库文档** `.agents/docs/t100/devdata-materials.md` 的 §2 表做
一致性断言 —— 代码与文档不许漂移（计划文件不入库，所以不能拿它当断言对象）。

### M（共用资料，环境级）

| 角色 | 来源 | 落点 | 必需性 | 上限 |
|---|---|---|---|---|
| 规格表 53 张 `DZ*_T` | schema `DS`（同义词，不带 schema 前缀） | `dict/<env>.db` | 表缺失 → `missing[]` | 单表 200 万行 → `skipped` |
| 模板树 | `$TOP/erp/mdl/**` | `mdl/…` | 缺目录 = 失败 | 总量 200 MB |

### D（按需资料，程序级）

| 角色 | 来源（相对 `$TOP`） | 落点 | 必需性 |
|---|---|---|---|
| `tab` | `erp/<mod>/dzx/tab/<prog>.tab` | `dzx/<prog>.tab` | **必需** |
| `tgl` | `erp/<mod>/dzx/tgl/<prog>.tgl` | `dzx/<prog>.tgl` | **必需** |
| `4gl` | `erp/<mod>/4gl/<prog>.4gl` | `src/<prog>.4gl` | **必需** |
| `tsd` | `erp/<mod>/dzx/tsd/<prog>.tsd` | `dzx/<prog>.tsd` | 存在才拉 |
| `tap` | `erp/<mod>/dzx/tap/<prog>.tap` | `dzx/<prog>.tap` | 存在才拉 |
| `tgx` / `tgx2` | `erp/<mod>/dzx/tgl/<prog>.{tgx,tgx2}` | `dzx/…` | 存在才拉 |
| `inc` | `com/inc/erp/<mod>/<prog>_{mask,rep}.4gl` | `inc/…` | 存在才拉 |
| `form` | `erp/<mod>/42f/<prog>.42f` | `form/<prog>.42f` | `--form` |
| `design` | 本地工作区 `<prog>*.tzs` | `design/` | `--with-design` |

单文件上限 64 MB（`.42f` 这类二进制尤其要拦）。

## 为什么 Oracle 要分页（内存纪律）

`pkg/erpdb` 有个可选的 `Streamer` 接口（逐行回调），但 **Oracle 驱动不真的流式**：
它先把整个结果集读进 `resultSet.rows` 再交给 `database/sql`。实测 `DZBB_T` 52 万行时
进程工作集涨到 **1.1 GB** 且几乎不前进（GC 抖动）。

所以 `DZAA`/大表一律**按页查**（`dbTableSource.queryPaged`）：每页一条独立 SQL，
`ROWNUM` 窗口 + 主键排序，一页 2 万行，内存与表大小无关。两个硬约束：

- **没有主键就不分页**（回退单条流式）：无序分页会重复/漏行 —— 静默错数据比慢更糟；
- 排序键要 `"UPPER"` 引号形态，否则 Oracle 把 `"dzaa009"` 当成另一个名字（实测 `ORA-00904`）。

金仓走 pgx，本身逐行，不需要分页。

## manifest 契约（`tt.devdata/v1`）

见 [manifest.go](manifest.go) 的 `Manifest`。规则四条：

1. **每个落盘文件都有 sha256** + 源 mtime（前者判"内容被改过"，后者判"源变新了"）；
2. 缺的东西进 `missing[]`，**不判失败**（有的程序本来就不产 `.tap`/`_rep.4gl`）；
3. `versions` 是"规则是否仍适用"的锚点：模板 `(Version:N)`、`.tgl` 头部的 SD/PR 版次、
   该程序的 `DZAA002` 版次；
4. `warnings[]` 只承载一致性提示（**tab 比 tgl 新 → 蓝图改过但骨架未重产，产物可能滞后**），
   不阻断。

M 侧用同一个 schema：`files[]` 只登记真正落盘的文件（`role=dict` 一个 + `role=mdl` 每片一个），
`tables[]` 登记库里的表（`table/rows/skipped/missing/desc`）。不让每张表在 `files[]` 里各占一条
（并各给一个 sha256）是刻意的：表没有独立文件，硬编一个 sha 只会变成"看起来校验过、
其实校验的是整库"的假证据。

### 版本维度（ADR-03 的实测口径）

53 张表在 `福联测试tst@1108`（2026-11-21）实测合计 **18,371,187 行**，远超 plan 的 50 万阈值，
所以结论是**只镜当前有效版**：

- `DZAA_T`（702,438 行 / 5,445 个程序）：`DZAASTUS='Y'` 且版次为该程序最大值 → 实测 272,355 行；
- 其余 52 张**整表镜**。不做通用 `*STUS='Y'` 过滤 —— 实测 `DZEB_T`/`DZED_T` 的 Y 行数为 0，
  通用过滤会静默清空整张表；
- 单表超 200 万行 → `skipped` + `warnings`。实测会命中 `DZBA_T`(781 万) 与 `DZAZ_T`(462 万)，
  两张都是生成器的**运行时产物/计数**，不是入参。

## 失败语义（CLI 退出码 / HTTP 码）

| 情况 | 哨兵错误 | CLI | HTTP |
|---|---|---|---|
| 环境不在 `hosts.sshs` | `ErrEnvNotFound` | 2 `USAGE` | 404 `NOT_FOUND` |
| 程序名非法（`../x`、大写…） | `ErrInvalidProg` | 2 `USAGE` | 400 `USAGE` |
| 模块推断不出 | `ErrModuleUnknown` | 2 `USAGE` | 400 `USAGE` |
| 程序没拉过 | `ErrProgNotFound` | 4 `NOT_FOUND` | 404 `NOT_FOUND` |
| 已拉过但没 `--force` | `ErrExists` | 2 `USAGE` | 409 `CONFLICT` |
| 同类任务在跑 | `ErrBusy` | 2 `USAGE` | 409 `CONFLICT` |
| 没有可停的任务 | `ErrNoJob` | 2 `USAGE` | 409 `CONFLICT` |
| 网络/库/写盘失败 | — | 5 `FAILED` | 500 `INTERNAL` |
| 用户取消 | `context.Canceled` | 5 `CANCELLED` | — |

**校验先于落盘与拨号**：非法程序名不会先建目录再报错（`server_test.go` 钉住这条）。

## 回滚与单飞

- **先写临时目录、成功才原子改名**：`<目标>.tmp-<rand>` → rename。取消/失败删临时目录，
  并把同前缀的历史残留一起清掉（`CleanupTemp`）。半套材料比没有材料更坏。
- **单飞**（`Manager`）：`(kind, env[, prog])` 一把锁；`mirror` 按环境、`fetch` 按程序。
  排队会让"点了没反应"与"在跑"混在一起，所以直接 409。
- **幂等**：`mirror` 已同步过默认跳过（不重写、不报错）；`fetch` 已拉过需要 `--force`。

## 与读侧的区别（一次说清）

| | `tt file mirror` | `tt dict db sync` | `tt gen …` |
|---|---|---|---|
| 给谁用 | 给人/agent 读代码 | 给查询用 | 给 **model** 用 |
| 落点 | `srccache/` | `dict/<env>.db` | `devdata/<env>/` |
| 粒度 | 全量 | 整库 | **按程序成套** |
| 指纹 | 无 | 无 | manifest（sha256 + 版本锚点） |
| 清理 | `tt cache`（7 天/整目录） | `tt dict db clear` | `tt gen clean`（默认 dry-run） |

**不进 `pkg/cachedir` 白名单**（ADR-02）：登记成缓存就会被 `tt cache clear` 整目录删掉、
被启动清理按 7 天删旧文件。而 devdata 里可能是某环境唯一一份已归档的蓝图，生产环境甚至
可能已经重新产生过 —— 删了要重新拉，成本是几十 MB + 数据库查询。

## 测试

```bash
go test ./internal/devdata/... -count=1
```

- 假 SFTP / 假库（`fake_test.go`）：缺 `tab` 失败、缺 `_rep` 只记 `missing`、取消无残留、
  超上限跳过、写库失败整体回滚、幂等跳过、`--only` 不抹另一半；
- manifest：往返、`stale` 五种触发、`tab 比 tgl 新` 的警告、模板版本与 `.tgl` 头部解析；
- `roles`：与 plan 附录 A 的一致性断言（53 张、行数合计、`LatestOnly` 只有 `DZAA_T`）；
- server：400/404/409 契约、校验先于领域层调用、未知端点 404 JSON。

## 细节去哪

- 命令行入口 → [../cli/devdata/README.md](../cli/devdata/README.md)
- HTTP 挂载 → [../web/README.md](../web/README.md) 的路由表
- 前端分区 → [../../../web/app/src/README.md](../../../web/app/src/README.md)
- 材料口径与实测数据 → [../../../.agents/docs/t100/devdata-materials.md](../../../.agents/docs/t100/devdata-materials.md)
