# internal/dict — ERP 数据字典

查 T100 的元数据：表与字段、校验带值、系统分类码、字段画面规格、可复用开窗、系统消息、
参数定义、程序与作业。**两种数据源、一个接口、一份数据族清单。**

## 两种数据源

| 源 | 实现 | 说明 |
|---|---|---|
| **本地镜像** | [./db](./db/README.md) | 本地 SQLite 副本，离线可用、快 |
| **远程直查** | [./live](./live/README.md) | 连客户的 ERP 库现查，数据总是最新的 |

**两者实现同一组查询方法、返回同一批结构**（[./db](./db/README.md) 的 `Source` 接口），
所以命令层不区分本地与远程 —— 由数据源策略决定，命令本身不关心。

数据源的选取顺序（`cli/dict` 的钩子）：

1. `--env` / `--conn <环境名|local>` —— **显式指定即强制**；
2. 配置里的查询数据源（环境名或 `local`）；
3. 都没有 → **在线优先**（只有连环境都没配时才落到本地）。

## 数据族

**`Families` 是唯一的事实来源**：哪些表属于哪一族、谁依赖它、同步顺序 ——
同步清单、数据覆盖状态、各命令帮助里的"本地数据齐不齐"提示全部从它生成。

| 族 | 中文名 | 依赖命令 | 表 |
|---|---|---|---|
| `table` | 表字典 | `r.t` | 表档 / 字段档 / 键值档 / 索引档等 9 张 |
| `check` | 校验带值 | `r.v` | `dzcd_t` 系 5 张 |
| `scc` | 系统分类码 | `scc` | `gzca_t` 系 4 张 |
| `spec` | 字段画面规格 | `desc` | `dzep_t` |
| `win` | 可复用开窗 | `r.q` | `dzca_t` 系 5 张 |
| `msg` | 系统消息 | `msg` | `gzze_t`、`gzzal_t` |
| `param` | 参数定义 | `sysp` / `docp` | `gzsz_t` 系 3 张 |
| `prog` | 程序与作业 | `prog` | 程序档 / 作业表 / 应用参数组等 4 张 |
| `progtable` | 程序与表格 | `prog` / `r.t` | `gzdg_t` |
| `subprog` | 子程序与元件 | `prog` | `gzde_t`、`gzdel_t` |

**缺族不是"部分可用"**：依赖它的命令会**整体报错** —— 所以状态表里那种情况行数显示为
"无意义"，而不是给一个会让人误以为能用的数。

## 字段表（`fieldtab`，M3）

`tt dict fieldtab <模块>/<表单>`：把**编译后的表单文件** `.42f`（明文 XML）解析成字段表 ——
每个字段的身份、编号、载体、控件、SQL 类型、必填，以及**中文名**。

* **解析不复制**：一律走 [`../form42f/form42f.go`](../form42f/form42f.go) 的 `ParseForm42F`
  （`tt form show` 用的同一个，`internal/form42f`）。本包只做"字段表 × 字典"的合并与路径定位。
* **表单只读本地目录**：`--root <目录>` 或环境变量 `TT_FORM_DIR`；目录下按
  `<模块>/42f/<表单>.42f` 或 `<模块>/<表单>.42f` 找。表单这一侧**不连服务器、不起会话**。
* **中文名走字典既有解析**：把 `.42f` 里的 `cmt_*` / `lbl_*` / `colName` 键接到
  `dzeb_t`+`dzebl_t`（`tt dict r.t` 的同一套字段说明）。
* **两条字典通道，缺省在线**（开关 `--source online|local`，实现见
  [`../cli/dict/fieldtab_source.go`](../cli/dict/fieldtab_source.go)）：

  | 通道 | 读哪儿 | 不可用时 |
  |---|---|---|
  | `online`（**缺省**） | 直查服务器字典表，只发 `select`；环境取 `--env/--conn`，其次配置 `query.source` | **报错 + 下一步**（`--env <环境名>` 或 `--source local`），**绝不**偷用本地库（A15） |
  | `local` | 本地 SQLite 镜像（`-d/--db` / `TDICT_DB` / 数据目录下的 `erp_data.db`），只读 | 报错并指向 `tt dict sync`；按 **A31 归 `missing`（确定缺失）**，不是 `unknown` |
  | `none`（`--no-dict`） | 不接字典，整列 `unresolved/no-dict` | 不算错误（显式要求）；优先于 `--source` |

  两条通道**都只读**：查询路径里不写任何东西（同步是另一条命令 `tt dict sync`）。
  "这一轮用的是哪种"必须写在脸上：JSON 看 `dictSource`，人读输出多一行 `字典源: …`。
* **不猜**（契约 A15/A18/A31）：字典查不到 → `label` 为**空串**、`labelState=unresolved`，
  并在 `labelReason` 写清是哪一种"查不到"。**绝不**回落成 `colName` / `cmt_*` / `lbl_*`
  当中文名；多个候选给出不同名字时判 `ambiguous`，绝不挑一个。

### 输出键（JSON,C5.2 的键一个都没动）

```json
{"module":"t100","form":"axmt500_wf","path":"…/t100/axmt500_wf.42f",
 "fields":[{"colName":"b_xmda001","fieldId":18,"tag":"TableColumn","widget":"Edit",
            "sqlType":"SMALLINT","sqlTab":"formonly","required":false,"notNull":false,
            "comment":"cmt_xmda001","textKey":"lbl_xmda001",
            "label":"料号","labelState":"resolved","labelFrom":"xmda_t.xmda001"}],
 "duplicates":[],"skipped":0,
 "labelResolved":1,"labelUnresolved":562,"labelSource":"dict","notes":[],
 "dictSource":"online"}
```

| 键 | 含义 |
|---|---|
| `colName` / `fieldId` / `tag` / `widget` / `sqlType` / `sqlTab` | 表单文件原样带出（与 `tt form show` 同名同义，但**各自独立**） |
| `required` / `notNull` | 同上 |
| `comment` / `textKey` | `cmt_*` / `lbl_*` 键原值（中文名的**线索**，本身不是中文名） |
| `label` | 中文名。**未解析时必为空串** |
| `labelState` | `resolved` / `unresolved`（只有这两种） |
| `labelFrom` | 命中的字典位置 `<表>.<字段>`，可回溯到 `tt dict r.t <表>` |
| `labelReason` | 未解析原因：`no-dict` / `no-table` / `not-in-dict` / `ambiguous` |
| `labelResolved` / `labelUnresolved` / `labelSource` | 整表统计与字典来源（`dict` / `none`） |
| `dictSource` | **本轮新增**：这一轮用的通道 `online` / `local` / `none`（既有键一个没改名、没删除） |
| `notes` | 查表**故障**（连接/权限/SQL）如实记账；"字典里没这张表"属正常，不刷 note |

人读形态里未解析行直接打印"未解析"，不留空单元格（空单元格会被读成"这一列没意义"）。

## 长跑动作

| 动作 | 实现 |
|---|---|
| 源码镜像拉取（从服务器拉 ERP 源码到本地） | [../host/README.md](../host/README.md) |
| 字典同步（远程库 → 本地 SQLite） | [./dbsync](./dbsync/README.md) |

服务端只为这两个动作提供接口（[./server](./server/README.md)），它们也是设置页里那两张卡片。

## 判据

```bash
go test ./internal/dict/... -count=1
go test ./internal/cli/dict -count=1
```

`fieldtab` 的**反证钉子**在 `formfields/formfields_test.go`：
`TestMerge_UnresolvedNeverFallsBackToColName` —— 把"未解析"改成"回落 colName"必红
（真样本上还有 `TestLoadFormAxmt500FieldTable` 同一个钉子）。

两条通道的反证钉子在 `cli/dict/fieldtab_source_test.go`：把 `--source` 的缺省改成 `local`
（`TestFieldtabSourceDefaultIsOnline`）、或把"在线不可用"改成静默回落本地
（`TestFieldtabOnlineUnavailableIsErrorNotFallback` + `TestFieldtabE2EDefaultOnlineNeverUsesLocalMirror`）
都**必红**。

## 细节去哪

- 查询命令的用法与逐条说明 → [`skills/tt-dict/SKILL.md`](../../skills/tt-dict/SKILL.md)
- 输出的形状（信封里为什么有那么多字段） → [../output/README.md](../output/README.md)
- 结果的落盘与翻页 → [./db/README.md](./db/README.md) 与 `tt dict spill`
