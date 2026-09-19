# 任务板 — T100 `.tzs` AI 集成

契约见 `SPEC.md §11.24`。计划全文在 `C:\Users\18526\.claude\plans\crystalline-cuddling-kahan.md`。

## 现在的状态

**W0–W3 全部走完，48 个函数全部实现，W3 门已跑完并通过。** 结论与四个缺陷见
`HANDOFF.md §18`。一句话：四个批脚本回归到基线（`bsft001_wf` 一行的差异已查明是人手点弹窗，
不是回归）· `gate-w3.py` 段 A 9/9、段 B 21/21（当时的语料是 81 个文件） · `gate-w3-fns.py` 34/34 写函数被驱动、
125/125 负向对照以契约码拒绝 · **门挖出 4 个缺陷，`list_tables`/`list_columns` 不可达已修**。

> **共享 `$OUT=/c/Users/18526/AppData/Local/Temp/dep_probe` 是关卡专用**——门在用它。
> 每个 agent 必须用自己的私有 `OUT`（`dep_w2*` 等），否则跑了一半的门会被换掉的二进制污染。

## 关卡怎么跑（三个脚本，全语料是关卡手段不是迭代手段）

```bash
python gate-w3.py                     # 出货路径 + 全语料（现 67 个），约 8 分钟
python gate-w3-fns.py --files "aapp320(c).tzs,aist310_wf(c).tzs"   # 33 个写函数 × 指定样本，约 40 秒
python gate-w3-fns.py                 # 不带 --files 就是【全语料】× 33 个写函数，很久——别顺手跑它
python gate-w2.py                     # 六个请求的里程碑链，秒级
./batch.sh && ./batch-write.sh && ./batch-edit.sh && ./batch-action.sh   # 约 20 分钟
git diff --stat -- '*.tsv'
```

## 波次

| 波次 | 并发 | 任务 | 独占文件 | 依赖 | 状态 |
|---|---|---|---|---|---|
| **W0** | 串行 | W0.1 建基线 | `.gitignore` | — | ✅ `f710a51` |
| | | W0.2 冻结契约 | `SPEC.md §11.24` | — | ✅ `1187a98` `4314707` |
| | | W0.3 两个分歧裁决 | — | — | ✅ 已查证（`SpecItem` 是幽灵，`Call` 统一到强版） |
| | | W0.4 `build.sh` v2 | `build.sh` | — | ✅ 13 个 BUILT；纯净度检查会响 |
| | | W0.5 语料固定 + 重采基线 | `corpus.manifest` `make-manifest.sh` `batch.sh` | — | ✅ 当时 81 个固定文件，四脚本已重采并提交（现 67 个，基线待重采） |
| | | W0.6 close/reopen + ProgramKey 碰撞探针 | `test/ProbeReopen.cs` | — | ✅ **可以重开**；但挖出 close 泄漏订阅者（见 §11.24 g-1） |
| **W1** | 2–3 | W1-A `src/` → `TzsCli.dll` + `build.sh` 三值化 | `build.sh` `src/*.cs` | W0.4 | ✅ |
| | | W1-B `src/Designer/*` 五个抽取文件 | `src/Designer/*` | W0.2 | ✅ 5 文件、编译 + 运行时已验证；已修它报的两个问题 |
| | | **W1 门** 只跑读侧 `batch.sh`（新规则，2 分钟） | — | 全部 W1 | ✅ |
| **W2** | 3 | W2-A 命名管道守护进程 + `Rpc.cs`/`Manifest.cs` + `tzs-cli` | `test/tzs-server.cs` `test/tzs-cli.cs` `src/Designer/Rpc.cs` `Manifest.cs` | W1 | ✅ |
| | | W2-B `Fns/Read.cs` + `Fns/Attr.cs` | `src/Designer/Fns/Read.cs` `Attr.cs` | W1 | ✅ 6 项验收全过（含负向对照、夹紧、幂等） |
| | | W2-C `Fns/Validate.cs` + `Verify.cs` + `Session.cs` | 对应三个文件 | W1 | ✅ T1+T3 全过；挖出两个静默故障 |
| | | **W2 门** 一个 server 进程连续跑通六个请求 | — | 全部 W2 | ✅ **9/9 PASS**（`gate-w2.py`），含真写与 RoundTrip 0/0 |
| **W3** | 6 | W3-A `Fns/Struct.cs` 结构 12 | `src/Designer/Fns/Struct.cs` | W2 | ✅ |
| | | W3-B `Fns/PageTab.cs` 页签 2 + Tab 2 + 工具 2 | `src/Designer/Fns/PageTab.cs` | W2 | ✅ |
| | | W3-C `Fns/Action.cs` 语义 1 + Action 3 | `src/Designer/Fns/Action.cs` | W2 | ✅ |
| | | W3-D `Fns/Semantic.cs` 多语言/选项/串查 6 + 复制 1 | `src/Designer/Fns/Semantic.cs` | W2 | ✅ |
| | | W3-E **守护进程捕获挂死**（阻塞出货路径） | `test/tzs-cli.cs` `src/Designer/Rpc.cs` | — | ✅ |
| | | W3-F `E_NO_OP` 契约对齐（它应是成功不是错误） | `src/Designer/Fns/Attr.cs` | — | ✅ |
| | | **W3 门**：出货路径 + 全语料 + 33 个写函数 | — | 全部 W3 | ✅ **`gate-w3.py` 21/21、`gate-w3-fns.py` F1–F10 过（33 个写函数）**；批脚本回归到基线。挖出 4 个缺陷，1 个已修（见 `HANDOFF.md §18`） |

## 已完成任务挖出的东西（都已落进 SPEC）

- **W0.6**：close/reopen **可以**（会话模型成立）。但 close 三个 Remove **不够**——每次 open 挂 3 个
  Global 订阅者，只有发布 `TzpFileClose` 能摘掉；泄漏的 `SaveSettingEvent` 处理器按 key 解析，
  会让下次保存**把已关闭包的模型写进新包**。另外失败的 open 会**毒化 key**（`map.Add` 先于
  `SpecificationInfo.Create`），必须回滚。抛出的是 `SetInitGridY` 不是 X（契约原文记错）。
- **W1-B**：库里的 `Environment.Exit` 会杀掉守护进程——已改为抛 `TzsError`（只剩看门狗那处，那是契约）。
  `Session` 保留了三处契约禁止的行为，已按 §11.24 修正。

## 样本集 —— 日常自检只用这三个文件

**全语料（现 67 个文件）是关卡手段，不是迭代手段。** 每个文件要起一个完整设计器进程（~1 秒起步），
四个批脚本一轮 20 分钟以上。把它用在每个 agent 的每次自检上，反馈延迟就毁掉了迭代速度。

### T1 —— 默认自检（3 个文件，合计 349 个元素）

| 文件 | 元素 | tpl | Tree | progrel | rfield | 工作区 | 覆盖什么 |
|---|---|---|---|---|---|---|---|
| `asft330(c).tzs` | 203 | F | **✓** | — | ✓ (33) | **xiyuan/prd** | 跨工作区（`TZSCLI_WS`）、含 Tree 与 Table（语料里少数同时有两种的） |
| `cpmp530(c).tzs` | 32 | **Q** | ✗ | ✗ | ✗ | hengshuo/prd | tpl=Q（少数）、且是**已知 posX 漂移**的文件之一，正好测基线相对判据 |
| `aapp320(c).tzs` | 114 | **P** | ✗ | ✓ | ✓ | hengshuo/prd | tpl=P、字段/标签/描述三件套、文档里所有例子用的就是它 |

三条命令各覆盖一个轴，加起来比单个大文件还快。

> **2026-09-19 语料变动**：`xiyuan/tst` 整个模块已从工作区删除，语料从 81 个降到 **67 个**
> （hengshuo/prd 63 + xiyuan/prd 4），`corpus.manifest` 已重新钉定。原先 T1 里承担"跨工作区"
> 轴的是 `cs_excel_in_xmdl_s01(c).tzs`（也随之消失），现在由 `asft330(c).tzs` 承担。
> **四个批脚本的基线 TSV 因此作废**——它们记录的是 81 个文件，其中 14 个已不存在
> （8 个属 `xiyuan/tst`，另 6 个是早先的临时产物）。重采基线是一次约 20 分钟的运行，尚未做。

### T2 —— 加"宽"用例（+1 个文件，约 4 秒）

`aist310_wf(c).tzs`（577 元素，**Tree** + progrel + 34 个 act）。Tree 是必须覆盖的——
晋升规则、`c` 标记、`ComponentFactory.cs:72-76` 造的脚手架子节点都在这里。
只在改动**碰到 Tree / 晋升 / 结构**时才加。

### T3 —— 只有验校验基线时才用

`axmt500_wf(c).tzs`（678 元素）——它是**未改动就报 11 条 WARNING** 的那个，是"判据必须相对基线"
这条规则的证据。但 `validate` 在它上面要 **10.4 秒**，所以只在专门验校验器时跑一次，不进日常自检。

### 全语料 —— 只在关卡跑

`./batch*.sh` 只在 **W1 门 / W2 门 / 合并前**跑。任务中途不许跑。


---

## 纪律（每个 agent 的任务说明里都写了）

1. **只碰自己名下的文件。** 要改别人的，先提给协调者，不许就地改。
   没有 git worktree（本目录是独立仓库，但同一工作树），**按文件分区就是隔离**。
2. **私有 `OUT`。** 共享 `$OUT` 在基线跑完前是禁区。
3. **不跑 `batch*.sh`。** 它是共享资源，且 15 列 `SUMMARY` + `cut -f9` 基线逻辑焊死在里面。
4. **`RoundTrip.exe` 是法官。** Wave 1–3 不许改它的输出格式——被告受审期间不换法官。
5. **`[assembly: AssemblyVersion("1.0.0.251")]` 只放在 exe 的 Main 文件里**，库里一个都不许有（CS0579）。
6. **一律按 key 寻址，永不读 `TzpManager.Current`**——实测它会被改写成最后加载的包。

## 已查证的关键事实（别重复调研）

- **`ProgramKey` 不含路径**（`TzpManager.CreateKey`:229 = 程序名 + 类型）。语料里四个 `aapp320` 文件撞同一个 key；
  现有 `LoadPackage` 的 `map.Remove/Add` 会**静默丢弃前一个句柄而留下它的 `UndoRedoManager`**。→ `SPEC §11.24 (g)`
- **`batch.sh` 的旧基线含 10 个 `_ai_*.tzs` 生成物**（三个写侧脚本都排除了，只有读侧没有）→ 当时是 81 个固定文件（现 67 个，见下方语料变动）
- **布局属性只能走 `XmlElement` 索引器**，直接调 `FormAttributesUndoRedoCommand` 会跳过白名单/门禁/夹紧（P0 实测）
- **校验器要先用 `SaveToForm`/`SaveToTSD` 刷新快照**，`FormElement`/`TSDElement` 是构造时一次性赋值的
- **`ValidateForm` 必须在 `finally` 里恢复 `false`**——它是进程级，影响别的包的加载
- **`Newtonsoft.Json.dll` v9.0.0.0 在安装目录里**，直接引用，不手写解析器
- **`SpecItem` 不存在**（只出现在 `AddField.cs:470`），是 `Prop()` 静默返回 null 的死条目

## 里程碑交付物

一个 server 进程连续处理六个请求（**不是六次 CLI 调用**——那等于什么都没证明）：

```
open → find_component → get_component → set_layout_attr → validate → save
```
