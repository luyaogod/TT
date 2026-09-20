---
name: tt-dev
description: 安全编辑 T100 设计器包：.tzc 代码包（4GL/TAP/TGL）走 export→改围栏工作区→verify→apply；.tzs 表单包用 tzs call 读写（由设计器自己的引擎驱动，不是拼 XML）。当用户要改 T100 客制程序、新增/改名/改签名自订函数、解锁框架区段、修改表单（改属性/加字段/加页签/调布局）、或查看 .tzc/.tzs 包里有什么时使用。
license: 与 tt 仓库一致（见随包 README.md）
metadata:
  tool: tdev
  command: tt dev tzc / tt dev tzs / tt install
---

# tdev：T100 设计器包的安全工具

`tdev` 有**两条互不相通的管线**，先分清再动手 —— 这是最容易犯的错：

| | `tt dev tzc`（代码包 `.tzc`/`.tzf`/`.tzx`） | `tt dev tzs`（表单包 `.tzs`/`.tzv`） |
|---|---|---|
| 干什么 | 把包渲染成**带围栏的 `.4gl` 工作区**给人/AI 编辑，改完写回 | `export` 纯解压；读写表单走 `call` |
| 产物 | 工作区：`prog.full.4gl` + `manifest.json` + `snapshot/` + `.tdev/` + `.git/` | 就是一包文件（`.tsd`/`.4fd`/`ver`…），默认目录带 `-unzip` 后缀 |
| 能写回吗 | 能，**唯一写路径是 `apply`** | 能，**唯一写路径是 `call save`** —— 由设计器自己的代码算，不是我们拼 XML（`export` 的产物本身仍只读） |
| 用途 | 改 4GL 客制逻辑 | 改表单：改属性 / 加字段 / 加页签 / 调布局 / 改多语言 |

拿错入口会被挡住：`.tzc` 跑 `tzs export` → 退出码 2 并提示改用 `tzc export`。

## 标准流程（`.tzc` 代码包）

```powershell
tt dev tzc export "D:\pkg\capt110(c).tzc"    # → D:\pkg\capt110-ws（默认，-o 可改）
cd /d D:\pkg\capt110-ws                    # cd 进去之后所有命令都能省 <dir>
# 只改围栏内标着 EDITABLE 的正文
tt dev tzc status                            # 改了哪些点、在第几行、会被拒还是放行
tt dev tzc verify                            # 不写盘预检
tt dev tzc apply                             # 唯一会改 .tzc 的命令（先备份到 .tdev\prev.tzc）
```

## 标准流程（`.tzs` 表单包）

表单的读写**由设计器自己的代码算**：`engine/` 里那个 C# 引擎反射驱动**已安装的设计器自己的程序集**
（布局属性走设计器自己的 `XmlElement` 索引器，`.tsd` 由设计器从模型重算）。我们一个字节的 `.tzs`
格式都没实现 —— 所以**手工拼 XML 这条路是不存在的**，改表单只有 `call` 一条路。

### 一次改动的完整形状

```powershell
tt dev tzs doctor                            # 先自检：引擎 exe、设计器目录、工作区、管道名
tt dev tzs fns                               # 49 个函数（按组列）；`fns <fn>` 看单个函数的参数

# open → 读 → 改 → validate → save → close
tt dev tzs call open           --path "D:\pkg\aapt300(c).tzs"   # → {"handle":"h9", "state":"Loaded", …}
tt dev tzs call form_tree      --handle h9 --depth 3            # 结构树，每节点带 name-path
tt dev tzs call find_component --handle h9 --query worksheet    # 控件代号 → name-path
tt dev tzs call get_component  --handle h9 --path managedform/aapt300/HBoxT1/worksheet
tt dev tzs call describe_kind  --handle h9 --kind field         # 这类节点**运行时**能写哪些属性
tt dev tzs call set_spec_attr  --handle h9 --path <path> --kind field --attr can_edit --value true
tt dev tzs call validate       --handle h9                      # 慢（实测 3–7 秒）；报**增量**
tt dev tzs call save           --handle h9 --out "D:\pkg\_ai.tzs" # 写到**新**包；原包一个字节不动
tt dev tzs call close          --handle h9
tt dev tzs stop                                                 # 停本工作区的常驻引擎
```

纯解压（只读参考、写文档用；**不依赖引擎也不依赖设计器**）：

```powershell
tt dev tzs export "D:\pkg\aapt300(c).tzs"    # → D:\pkg\aapt300-unzip\
```

要配两样（机器级依赖，没有合理缺省）：

- **设计器装在哪**：`tzs.installDir` 或 `TZSCLI_INSTALL`。第三方商业软件、**不随 tt 分发**。
- **用哪个工作区**：`tzs.workspace` 或 `TZSCLI_WS`。**这条没有缺省、也不回落** —— 引擎内置的
  默认工作区是一个**真实客户目录**，落上去等于拿别人的表单当草稿纸。三层都空时拒绝启动（退 5）。

### 句柄（handle）：最容易踩的一组

- **每个需要句柄的函数都必须显式给 `--handle`**。没有「自动沿用上次句柄」这回事：
  状态文件（`%APPDATA%\T100\tt\.tt-tzs.json`）里虽然记了一份最近 `open` 的结果，
  但命令层不读它，省掉 `--handle` 只会得到退 2「缺必填参数 handle」。
- **句柄永不复用**：`close` 掉一个包再 `open` 同一个包拿到的是**新号**（不是原来那个）。
  拿旧号去调 → `E_NOT_FOUND`「句柄不存在或已关闭」（退 2）。这是安全失败，不是 bug。
- **同一个程序已经开着再 `open` → `E_KEY_IN_USE`（退 4）**，报错会说清是哪个文件占着。
  先 `close` 占用者，或 `open --force`（只对 `Loaded` 状态的占用者有效，`Mutable` 的会被拒）。
- 句柄只活在常驻守护进程里。**进程一死全部失效** —— 守护进程挂了就重新 `open`，别重放写请求。

### `validate` 是**基线相对**的，首调必然「零增量」

这是读输出时最容易误判的一处：`validate` 在**句柄上第一次被调用时，那次运行本身就是基线**
（`Fns/Validate.cs:139`）。所以第一次调用 `newErrors`/`newWarnings` **按构造就是空** ——
它证明不了任何事。要看「我改坏了没有」，必须在**同一个句柄**上先 `validate` 一次建立基线。

实测：`aapt300(c).tzs`（未做任何改动）首调 `baseline = 13 条 WARNING`，`after = 13`，
`newErrors = 0`、`newWarnings = 0`。**语料里本来就不干净** —— 所以判据永远是
「改动后的增量」，而不是「有没有 WARNING」。

返回体：`{baseline[], after[], newErrors[], newWarnings[], elapsedMs}`。
加 `--json` 会把**整帧**（`{id, ok, result, error, ms}`）原样打出，而不是只打 `result`。

### 参数语法（本地只做**语法**校验，语义交给引擎）

| 写法 | 含义 |
|---|---|
| `--handle h9` / `--handle=h9` | 两种都行；`=` 后允许空串（`--desc=` 是「清空」） |
| `--paths a,b,c` | 列表按逗号切 |
| `--paths a --paths b` | 列表**可重复**，多次出现会合并（`a,b,c`） |
| `--force` | 裸开关 = `true`（`--excluded` / `--cited` 同理） |
| 裸词 `foo` | **报错**。位置参数一律不接受（这一条与 `tzs-cli` 有意不同） |
| `--offset 2` | 按 manifest 定型成数字 2 |
| 标量给两次 | 报错（只有 `path[]`/`string[]` 可重复） |

**本地绝不拦的**（拦了就是「本地拒绝了一个引擎本会接受的调用」）：`attr` 白名单、
`kind` 的取值、`add_action` 的 `type` —— 这些合法集要从**活着的模型**或**该表单自己的**
`s_detail<n>` 记录里取，静态判不了。所以 `--attr bogus_attr` 会照发，由引擎用它自己的白名单拒绝。

两个附带事实：**参数打错的报错也要引擎在**（`call` 先拉 manifest 再校验参数，manifest 拉不到按
用法错退 2）；`--timeout` **最小 120 秒**，低于此值会被抬到 120（下限是给加载看门狗留的）。

### 其它必知

- **`--depth 1` 意思是「只有这个节点本身」**，不是「往下看一层」。要子节点至少 `--depth 2`。
- **`save --out` 写到新包，不动原包**。不要 `--out` 指向源包 —— 那是唯一会毁掉原始素材的写法。
- **请求一旦上线绝不重试**：协议无幂等键，而这些函数都在改设计器内存里的模型，重试是在赌
  「上一次写进去了没有」。守护进程死了就重新 `open`。
- `stop` **绝不 spawn**（让 stop 去起一个服务器是件可笑的事）；`reap` 收的是**引擎重编后**
  停不掉的孤儿守护进程（管道名含 MVID，重编就换名，旧进程的 pid 只能靠状态文件的 `orphans` 找到）。
   `reap` 不加 `--yes` 只列不杀。

## 红线（碰了就出事，没有例外）

1. **不许改围栏行**：`{//@tdev:begin point …}` / `{//@tdev:end point}` / `{//@tdev:end section}`。
   它们不是内容，是 tt dev 的坐标（区间记账），改了 gate1 直接拒。
2. **不许改围栏外的任何字节**：文件头尾、区段之间的空行、缩进对齐 —— 都不行。
   也**不要用编辑器格式化/重排整个文件**（只允许改围栏正文；行尾被归一成 LF 没事，改空白有事）。
3. **不许改 `END FUNCTION` / `END DIALOG` / `END REPORT` 行**（结构行，逐字节比对）。
4. **签名行的改动要走 `rename`**（它同步围栏 `fn` + 签名行 + 描述块，并在 `.tap` 里复刻墓碑事务）；
   手工只改一处会被 gate2 的 V1–V7 拒。
5. **`.4gl` 条目永远不被写回**（服务器 build 产物，渲染结果装不下它）。
6. **`tzs export` 的产物是只读参考**：它就是一包文件，没有 manifest/围栏。不要手工改完再塞回包，
   也不要拿它当 `tzc` 工作区去 `apply`。要改表单用 `tt dev tzs call` —— 那是设计器自己的模型在算，
   改完设计器打得开；手工拼 XML 则不然（`.tsd` 由设计器从模型重算，`.4fd` 的 Record 段要重建）。
7. **表单的请求一旦发出不要重试**：协议无幂等键，且这些函数都在改设计器内存里的模型，
   重试是在赌「上一次写进去了没有」。引擎的守护进程死掉时重新 `open` 即可，别重放写请求。
8. **不要手改 `.tap` / 不要直接改 zip**：`.tzc` 的唯一写路径是 `apply`（它做闸门校验 + 原子写）。
9. **不要编辑 `.tdev/base.full.4gl`**：那是 gate1 的比对基线，改它等于伪造验证。
   要改内容就改 `prog.full.4gl`，要改意图就用 `unlock`/`rename`/`newfn`。
10. **只读区段不要硬改**：先 `tt dev tzc unlock`（单向、有代价）；锚点区段永远只读，改不了。
   **`--yes` 必须由人确认，AI 不得自己加**。
11. **不要写真实语料/客户包**：`export` 只读源包；实验一律用副本或临时目录。
   `.tzs` 侧同理：`save --out` **永远指向新包**，绝不指向源包。

## 退出码

`.tzc` 那条线：`0` 成功 · `2` 包格式/用法错 · `3` 验证失败 · `4` 写入被拒（权限/确认不足）· `5` IO·环境失败。

`.tzs` 这条线**没有 `3`**，多一个 `1`：

| 码 | 含义 |
|---|---|
| `0` | 帧 `ok:true` |
| `1` | 引擎内部错（`kind=internal`，含 `E_NOT_IMPLEMENTED`） |
| `2` | 参数/环境不对：本地参数错、未知函数、manifest 拉不到、引擎的 `validation` 与 `not_found` |
| `4` | **设计器拒绝**（`kind=designer`，含 `E_KEY_IN_USE`） |
| `5` | 传输或环境失败（含加载超时、没配工作区） |

## 报错怎么读（先看行号）

`apply` / `verify` 拒绝时会打位置，直接照着改：

```
apply 被拒绝（gate1 error=1 warn=0 info=1）：
  [error] gate1.readonly-region  只读 Region 内容被改动（写入被拒）
            位置：prog.full.4gl:4          ← 你编辑的那份文件 + 行号
            该行：   # 我加的字             ← 该行内容
            deny=section-locked            ← 原因：框架未解开，先 unlock
```

删除类问题给 `.tdev/base.full.4gl:<行号>`（编辑后的文档里它已经不存在了）。
`status` 的改动清单也带行号：`改动的点  function.foo（prog.full.4gl:1234）`。
`--json` 里每条发现有 `file`/`line`/`snippet`；`status --json` 有带行号的
`changed_regions`/`added_regions`/`deleted_regions`。

## 常见简单错误（自查表）

- ❌ 拿 `.tzs` 去跑 `tzc export`，或拿 `.tzc` 去跑 `tzs export` → 看清扩展名，两个入口不同。
- ❌ 以为 `export` 改了包 → 它**只读**源包；写回只有 `apply`。
- ❌ 在没有工作区的目录里跑 `status`/`apply` → 退出码 5；先 `cd` 进工作区或显式给 `<dir>`。
- ❌ 改围栏行 / `END` 行 / 围栏外空行 → gate1 拒（退出码 3/4）。
- ❌ 用格式化工具重排 `prog.full.4gl` → 围栏外字节变了，被拒。
- ❌ 改只读区段不 `unlock` → 退出码 4；锚点区段即使 unlock 也只读。
- ❌ 一次改一大片再 `apply` → 先用 `status` 看清单与行号，小步走。
- ❌ 把 `tzs` 解压产物当工作区去 `apply` → 没有 `tzs apply`；解压产物是只读参考，
  改表单走 `tt dev tzs call`。
- ❌ 源包在 `export` 之后被别人动过 → `apply` 拒写（退出码 5），重新 `export`。
- ❌ 新增函数后以为要给 `--desc` → `newfn` 没有 `--desc`，函数头用设计器模板（含顶部空行），
  自己填 `# Descriptions...:` 与正文。

`.tzs` 侧：

- ❌ 调 `call` 时省掉 `--handle` → 退 2「缺必填参数 handle」。**没有自动沿用上次句柄**这件事。
- ❌ 上一个包 `close` 之后仍拿旧句柄调 → 退 2「句柄不存在或已关闭」。**句柄永不复用**，
  重新 `open` 拿新号。
- ❌ 同一个程序已经开着又 `open` → 退 4 `E_KEY_IN_USE`。先 `close` 占用者（`list_open` 看是谁），
  或对 `Loaded` 的占用者用 `open --force`。
- ❌ 看到第一次 `validate` 报 `newErrors: []` 就以为改对了 → 首调**就是建立基线的那次**，
  零增量是构造性的。要比对增量必须在同一句柄上先建立基线。
- ❌ 拿 `baseline` 里的 WARNING 数当「改坏了」 → 语料**本来就不干净**（aapt300 未改动即有 13 条）。
- ❌ `form_tree --depth 1` 想看子节点 → 那是「只有这个节点本身」，至少 `--depth 2`。
- ❌ `--paths a b c` 希望是三个 → 列表要么 `--paths a,b,c`，要么 `--paths a --paths b --paths c`。
- ❌ 写参数时用位置参数（裸词）→ 一律报错，参数都是 `--名 值`。
- ❌ `--offset 2 --offset 3` 这种标量给两次 → 报错（只有 `path[]`/`string[]` 可重复）。
- ❌ 以为参数打错会立刻在本地报 → `call` **先拉 manifest 再校验参数**，引擎 exe 不在就要退 5。
- ❌ `--timeout 30` 想快点失败 → 会被抬到 120 秒（下限是给加载看门狗留的）。
- ❌ `save --out` 指向源包 → 唯一会毁掉原始素材的写法。永远写**新**包。

## 动词速查（`.tzc`）

| 命令 | 作用 | 改 `.tzc` 吗 |
|---|---|---|
| `export <pkg.tzc> [-o <dir>]` | 包 → 工作区（含 git init 与逐条目快照） | 否 |
| `status [<dir>]` | 报告改动（含行号），不写盘 | 否 |
| `verify [<dir>] [--strict]` | 不写盘跑 gate1 + gate2 | 否 |
| `apply [<dir>] [-o <pkg>] [--dry-run] [--yes]` | 全阶段校验 + 原子写包 + git commit | **是（唯一）** |
| `unlock [<dir>] --yes` | 框架解锁（单向、不可逆；先不带 `--yes` 看代价警告） | 否（只改工作区） |
| `rename [<dir>] <旧> <新> [--scope …] [--desc …]` | 结构事务：围栏 + 签名行 + 描述块原子同步 | 否 |
| `newfn [<dir>] --type FUNCTION\|DIALOG\|REPORT [--name …] [--scope …]` | 在 APPEND 锚点插入新空函数（带模板） | 否 |

## 动词速查（`.tzs`）

| 命令 | 作用 | 要引擎吗 | 改包吗 |
|---|---|---|---|
| `export <pkg> [-o <dir>] [--force]` | 纯解压（只读参考） | 否 | 否 |
| `fns [<fn>] [--json]` | 函数表 / 单个函数的参数 | 是（拉 manifest） | 否 |
| `manifest` | 函数表 JSON，原样转发 | 是 | 否 |
| `call <fn> --<参数> …` | 读写表单，**唯一写路径** | 是 | 只经 `save --out` 写到**新**包 |
| `doctor [--json]` | 环境自检（引擎/设计器目录/工作区/管道名/守护进程） | 是 | 否 |
| `stop` | 停本工作区的常驻引擎（不启动新的） | 否（要配工作区） | 否 |
| `reap [--yes]` | 清理引擎重编后停不掉的孤儿守护进程 | 否（要配工作区） | 否 |

改表单的函数按组分（`tt dev tzs fns` 看全表，会改模型的标 `[写]`，慢的标 `[slow]`）：

| 组 | 函数 |
|---|---|
| 会话 | `open` `save` `close` `verify` `list_open` |
| 读 | `form_tree` `find_component` `get_component` `list_spec_nodes` `describe_kind` `list_tables` `list_columns` `list_records` `list_local_strings` |
| 属性 | `set_spec_attr` `set_layout_attr` `set_tree_source` `rename_component` |
| 结构 | `add_widget` `add_field` `insert_at` `delete` `move` `reparent` `nudge` `align` `fit_size` `wrap` `break_layout` `convert_widget` `convert_container` |
| 页签 | `add_page` `delete_page` |
| 语义/Action | `insert_semantic` `add_action` `delete_action` `set_action_types` |
| 多语言/选项/串查 | `set_local_string` `set_items` `set_progrel_programs` `set_table_association` `set_spec_description` `set_cited` |
| Tab 顺序 | `set_tab_order` `tab_action` |
| 校验/工具 | `validate` `base_data` `set_excluded` `set_code_template` |

几处容易记混的：**`move` 只改 Z 序**（同一个父容器内前后挪），**换父容器用 `reparent`**
（设计器的拖拽 `DragComponentsUndoRedoCommand`，仅限同表单）。**`add_field` 加字段，
`wrap` 是拿选中元素包一层新容器**（不是加字段）。`add_field --columns a,b,c` 一次构造多列，
`--container Table` 才得到「一个 Table 装 N 列」。

## 自检与安装

```powershell
tt dev tzc selftest      # 31 项内置对抗用例，不需要真实语料
tt install path          # 把 tt.exe 所在目录加进用户 PATH（只动 HKCU）；合并后一次安装覆盖全部工具
tt install skills        # 把技能文件复制到当前目录的 skills/
```

细节（围栏协议、不变量 I1–I15、结构事务 V1–V7、与设计器的全部偏差 D-1…D-10/DV-2/DV-3、
真机验收清单 S7）见随包 `README.md` 与 `docs/tzc-model.md`。
