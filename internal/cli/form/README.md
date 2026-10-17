# internal/cli/form — `tt form` 命令组

出厂表单（`.42f`，编译后的**明文 XML**）的列出 / 查看 / 校验。契约的唯一口径是
[`../../../.agents/plan/tt-genero-formdef-contract.md`](../../../.agents/plan/tt-genero-formdef-contract.md)
的 **C2（取件）/ C5（CLI 面）/ C7（测试）** 与 **§4 的 A8（接口归消费方）**。

## 命令

| 命令 | 作用 |
|---|---|
| `tt form ls --env <环境> --module <模块>` | 列该模块下的表单名（可直接喂 `--form`） |
| `tt form show --env … --module … --form … [--json]` | 字段表：工号 / 表单内编号 / 控件 / tabIndex / SQL 类型 / 必填 / `cmt_*`+`lbl_*` / 几何 |
| `tt form verify --env … --module … --form … --map <colname-fieldid.json>` | 与归档语料表（或 `--recording <录制目录>`）逐条比：**可比 N 条 / 一致 N 条** |
| `tt form fetch --env … --module … [--force] [--json]` | 批量把该模块的表单取进缓存：**新取 / 命中缓存 / 失败**（失败逐条列名+原因） |
| `tt form cache [--json]` | 缓存的**只读**概览：多少张 / 多少字节 / 最旧·最新（不删任何东西） |

## 取件（C2 / A8）

消费方接口（`Source`）定义在**本包**（A8：接口归消费者，只 2 个方法），实现在两处：

| 实现 | 什么时候用 | 性质 |
|---|---|---|
| `DirSource` | `--dir <目录>` 或 `TT_FORM_DIR` | 本地目录/镜像，**离线**、测试全走它 |
| `host.RemoteFormSource` | 给了 `--env` 时（默认） | SSH 按需**单取**：`find` + SFTP `get`，**绝不写、不删、不起作业**（H6） |

真机那一路先把模块/表单名过白名单（`host.FormNameOK`）再拼进远端命令 —— 名字来自命令行，
直接拼就是命令注入；候选不唯一就报错并列候选，绝不自己挑一个（D4）。

## 缓存（C2.3 / C2.4 / C2.5）

* 落点 `<数据目录>/forms/`，目录名走 `config.CacheDir` 校验（`forms` 已登记进
  `config.CacheSubdirs` —— 清单外名字拿空串，漏登记的表现是"缓存静默失效"）；
* **只缓存原始 XML**：一个表单一个 `.42f` 文件，没有元数据旁车（解析每次现做，D5）；
* 缓存键含 `env/module/form`（进路径）+ `size/mtime`（用"把缓存文件的 mtime 设成源那边的"
  表达）：源实现了 `Identify` 时一次 Stat 就知道表有没有被重编译；问不到源那边时回落到
  **TTL**（`DefaultCacheTTL` = 10 分钟）；
* **三段名字过 UTF-8 白名单**（`cacheNameOK`，台账 D37）：中英文 / 数字 / `-` `_` `.` 放行，
  `.` 与 `..` 单独出现、任何含 `..`、路径分隔符 `/` `\`、控制字符（含 `\x00`）、
  Windows 保留字符 `: * ? " < > |` 一律拒。放宽到 CJK 是**必须**的：现场环境名就是 `主机109`，
  原来只认 ASCII ⇒ `tt form fetch` 90/90 全部写缓存失败且一声不响。
  卡住的不是字符集而是**结构**——放行的那批字符里没有一个能表达"往哪儿走"；
  拒绝时走 `Cache.PathError`，报错带病因（哪一段、为什么），不是静默返回空串。

写入点只有一处：`Cache.Put`（[cache.go](./cache.go)）。`show` 顺手写、`fetch` 批量写，
两者判"新鲜"的判据同为 `Cache.Lookup` —— 所以预热过的表单，`show` 立刻命中。

### `tt form fetch`（批量预热）

`ls` 拿到清单 → 逐张 `Form` → `Put`。**只读**：只 `list` + `get`，没有写服务器/删服务器的动作
（`Source` 接口本身长不出这些）。三条口径：

* **计数分开**：新取 / 命中缓存 / 失败三个数各自统计。要分开就必须自己 `Lookup`+`Put`
  （`CachingSource.Form` 把"命中"与"取到"合并成一个返回值），所以 `fetch` 拿的是**未包装**
  的取件通道，缓存由它显式读写；
* **`--force`**：跳过 `Lookup`，全部重取；
* **失败不静默**：哪张取不到、原因是什么，逐条进输出（人读与 `--json` 的 `failures[]` 都有），
  其它表单照常写入；退出码按 A13 取 **1**（`FORM_UNKNOWN`），**不是 6**。
  报告已经打到 stdout，所以返回的错误实现 `AlreadyReported`（与 `verify` 的 `mismatchError`
  同一手法），`--json` 下 stdout 上仍然只有一个对象。

### `tt form cache`（概览）

`forms` / `files` / `bytes` / `oldest` / `newest` + 逐份清单，全部是 `Stat` 出来的真值；
目录不存在时报 0 张（不是错误，也不顺手把目录建出来）。**它不删东西**：
清理走统一的 `tt cache clear`（`config.CleanCache` 只碰 `CacheSubdirs` 里的目录）。

## 输出与退出码

`--json` 走契约 **C5.2 的冻结键名**（`env/module/form/variant/build/fields[]/recordViews[]/duplicates`），
由 golden 快照守着（`testdata/golden/form_show_basic.json`，牙齿 b）：改键名必红。
**不加 `--json`（也没显式 `--format json`）时给人读**（表格）—— 本组默认与人读对齐，
与"本仓默认 json"的通行取向不同，判据在 `wantJSON`。

码表（本组，与 `tt fep` 同族；`internal/cli/README.md` 的整表尚未加这一行）：

```
tt form    0 成功
           1 FORM_UNKNOWN：表单取不到（网络/权限/路径）或不是合法 .42f；
             fetch 有表单取不到也退这一个码（失败逐条列在输出里）
           2 用法错（缺 --env/--module/--form、没有基线、语料里没有这个窗口）
           6 体检不过：verify 与语料逐条比对出现不一致（fetch/cache **不用**这个码）
```

`verify` 的失败信息**就是那份报告**，所以它返回一个声明"已报告"的错误
（`mismatchError`，见 `internal/cli/root.go` 的 `alreadyReported`）——
否则 JSON 模式下 stdout 上会出现两个对象。

## 测试

```bash
go test ./internal/cli/form/ -count=1        # 假 Source + 本地临时目录，不连真机
go test ./internal/cli/form/ -count=1 -v -run TestFormVerifyCorpusGated   # 门控（真样本+真语料）
```

* 假件：`form_test.go` 的 `fakeSource` / `injectSource`（替换 `sourceFor` 这一个裁决点）；
* `fetch` / `cache`：`fetch_test.go`（假件 `multiSource` 会**数底层被调用了几次** ——
  `--force` 与"二次命中"的唯一判据就是它，光看输出里的计数挡不住假装）；
  反证两条：缓存里出现解析结果（JSON）必红、`--force` 被忽略必红；
  `cache` 概览另有一条"跑前后文件集合逐项相同"的断言（它只读，不删不改）；
* 牙齿 b：`golden_test.go`；
* 比对口径与 `internal/fep/form42f_test.go` 的 `TestFormMatchesRecording` 同一套；
* 门控：真样本 `TT_FORM_SAMPLE_DIR`（默认 `../../.agents/Genero Original Factory Compatible/fixtures/forms/`）、
  真语料 `TT_FORM_CORPUS_MAP`（默认同目录 `out/semantic/colname-fieldid.json`）；
  缺失 `t.Skip` 并写明原因。**验收口径：窗口 `w_axmt500` 可比 509 条 → 必须 100% 一致。**
