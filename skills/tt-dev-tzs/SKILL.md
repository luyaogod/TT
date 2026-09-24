---
name: tt-dev-tzs
description: 读写 T100 设计器**表单包**（.tzs/.tzv）：由设计器自己的引擎驱动（52 个具名动词，命名管道 JSON-RPC），参数一律用 JSON 给，不是拼 XML。按数据表加字段用任务级动词 field_add（一条命令做完挑容器+建字段+校验+另存）；改属性用 set_spec_attr / set_layout_attr，改多个用它们的复数形式（一次请求、先全量校验再全量写）；布局/页签等用细粒度动词（open → 读 → 改 → validate → save）。要改表单、按表加字段、查表单结构或字段时用。**前提：先配工作区，且包与 out 都必须用绝对路径落在工作区目录之下。代码包（.tzc/.tzf/.tzx）不归这里，用 tt-dev-tzc。**
license: 与 tt 仓库一致（见随包 README.md）
metadata:
  tool: tdev
  command: tt dev tzs <动词> / tt install
---

# tt dev tzs：T100 设计器表单包的读写

表单包（`.tzs`/`.tzv`）的读写**由设计器自己的代码算**：引擎反射加载随包分发的设计器程序集，
`.tsd` 由设计器从模型重算。我们一个字节的 `.tzs` 格式都没实现 —— 所以**手工拼 XML 这条路不存在**。

代码包（`.tzc`/`.tzf`/`.tzx`）是**另一条互不相通的管线**，用 `tt-dev-tzc`：

| | `tt dev tzs`（本技能） | `tt dev tzc` |
|---|---|---|
| 干什么 | 由设计器引擎读写**表单模型** | 把包渲染成带围栏的 `.4gl` 工作区给人改 |
| 怎么调 | `tt dev tzs <动词> --args '<JSON>'` | `tt dev tzc export/apply …` |
| 唯一写路径 | `field_add … out` 或 `save`（都写**新**包） | `apply`（闸门校验 + 原子写回原包） |

拿错入口会被挡住：`.tzs` 跑 `tzc export` → 退 2 并提示改用 `tt dev tzs export`。

## 1. 三条路怎么选

| 你要做的事 | 用哪个 |
|---|---|
| **按某张表的列加字段**（最高频） | **`field add`** —— 挑容器 + 建字段 + 报校验增量 + 存新包，一条命令（§2） |
| 改属性 / 挪布局 / 换容器 / 页签 / 多语言 / Tab 顺序 | **细粒度动词链**（§3）—— 先看清结构，再改一处 |
| 只看结构、字段、开着的表单 | `form_tree` / `find_component` / `get_component` / `list_spec_nodes` / `list_columns` / `list_open` |

**跑之前要配一样东西**：工作区（`tt config set tzs.workspace "D:\ws"` 或环境变量 `TZSCLI_WS`；
**本文档的 `D:\ws` 一律指工作区**）。**这条没有缺省、也不回落** —— 引擎内置的默认工作区是一个
**真实客户目录**，落上去等于拿别人的表单当草稿纸；三层都空时运行类动词直接拒绝启动（退 5）。
工作区还**圈定了你能碰哪些包**：包的目录必须在它之下，`out` 也一样（§4.4）。
设计器**不用配**（它的程序集随仓库与发行包自带）。开工前先 `tt dev tzs doctor` 自检。

## 2. 任务级动词：`field_add`

一次做完：**挑容器 → 按列建字段 → 报校验增量 →（给了 `out` 就）存新包**。

```powershell
# ① 最省事：只给 file（没开着就顺手开，已开着就复用 —— 不会撞 E_KEY_IN_USE）
tt dev tzs field add --args '{"file":"D:\\ws\\apmt500_wf(c).tzs","table":"pmdl_t","columns":["pmdlent","pmdlsite"],"out":"D:\\ws\\_ai.tzs"}' --json

# ② 表单已经开着：用程序名寻址，不必抄句柄
tt dev tzs field_add --args '{"handle":"apmt500_wf","table":"pmdl_t","columns":["pmdlent"]}' --json

# ③ 容器不想让它挑：显式给 into（name-path 用 form_tree 看）
tt dev tzs field_add --args '{"file":"D:\\ws\\x.tzs","table":"pmdl_t","columns":["pmdlent"],"into":"managedform/x/HBoxT1/worksheet"}' --json
```

| 参数 | 必填 | 说明 |
|---|---|---|
| `handle` | 否 | 已在开的表单：句柄 `h1`、程序名 `aapp320`、或 ProgramKey `aapp320\|Form` |
| `file` | 否 | `.tzs` 路径；**与 `handle` 二选一** |
| `table` | **是** | 表名，如 `pmdl_t` |
| `columns` | **是** | 列名数组；一次构造 N 列（列名用 `list_columns` 取，**记得先过滤**，§4.7） |
| `into` | 否 | 父容器的 name-path；省略时自动挑（见下） |
| `container` | 否 | 容器模式，默认 `None`（`None` = 每个字段配一个 Label 控件；`Table` = 一个 Table 装 N 列） |
| `out` | 否 | 给了就把结果存成这个**新**包（绝不写源包）；**必须也落在同一工作区内**（§4.4） |

> ⚠️ `container` 的**枚举在 `--help` 里比函数体接受的宽**：`HBox` / `VBox` / `Page` / `Folder`
> 目前会被拒（`E_BAD_PARAM`「未知容器模式 HBox；合法值: None, Grid, Group, ScrollGrid, Table, Tree」）。
> 用 `None` 或 `Table` 最稳。（引擎里两份清单该对齐，属已知待办。）

**容器怎么自动挑**（`into` 可覆盖）：① 叫 `worksheet` 的容器（设计器模板里放字段的那块）→
② 名字含 `layout` 的容器 → ③ 根 `<Form>` 下唯一的容器 → ④ 都不成立就**报错要你显式指定**
（宁可不做，也不把字段塞进一个猜出来的盒子里）。

**它返回什么**（刻意瘦，约 1.4 KB）：`form` / `key` / `handle` / `opened`（这次是不是它开的）/
`container` / `added`（加了哪些节点）/ `validate`（`newErrorCount`/`newWarningCount` + 明细）/
`baselineCached` / `saved`（`out`、字节数）。**不回表单全量，也不回校验的 baseline/after 全表** ——
要看全量就单独调 `validate`。

**边界**：它只做"按表的列加字段"。改属性、挪布局、页签、多语言仍是细粒度动词 —— 走 §3。

## 3. 细粒度：一次调用改一处

```powershell
tt dev tzs doctor                                        # 先自检：引擎/设计器目录/工作区/管道名
tt dev tzs <动词> --help                                 # 参数表 + 一条能直接粘的示例

# open → 读 → 改 → validate → save → close（全程 --form 寻址，不搬运句柄）
tt dev tzs open           --args '{"path":"D:\\ws\\aapt300(c).tzs"}' --json   # → {"program":"aapt300",…}
tt dev tzs form_tree      --form aapt300 --args '{"depth":3}' --json            # 结构树，每节点带 name-path
tt dev tzs find_component --form aapt300 --args '{"query":"worksheet"}' --json  # 控件代号 → name-path
tt dev tzs get_component  --form aapt300 --args '{"path":"managedform/aapt300/HBoxT1/worksheet"}' --json
tt dev tzs describe_kind  --form aapt300 --args '{"kind":"field"}' --json       # 这类节点**运行时**能写哪些属性
tt dev tzs set_spec_attr  --form aapt300 --args '{"path":"<path>","kind":"field","attr":"can_edit","value":"N"}' --json
tt dev tzs set_spec_attrs --form aapt300 --args '{"path":"<path>","kind":"field","attrs":{"can_edit":"Y","can_query":"N"}}' --json
tt dev tzs validate       --form aapt300 --json                                 # 慢（实测 3–5 秒）；报**增量**
tt dev tzs save           --form aapt300 --args '{"out":"D:\\ws\\_ai.tzs"}' --json  # 写**新**包；原包一字节不动
tt dev tzs close          --form aapt300 --json
tt dev tzs stop                                          # 停本工作区的常驻引擎（内建命令，不是引擎动词）
```

> ⚠️ **上面这个顺序里那次 `validate` 证明不了任何事。** 一个会话上的**第一次** validate
> **就是建基线的那次**（§5），而这里它排在改动**之后** —— 所以 `newErrors` 按构造就是空。
> 想看真增量要**改之前**先 validate 一次（多花 3–8 秒）：
> `open → 读 → validate(建基线) → 改 → validate(真增量) → save`。
> 真正的"改对了"由**回读**证明，不是由 validate 证明。

**要改多个属性就用复数形式**（`set_spec_attrs` / `set_layout_attrs`）：一次请求、一次寻址、**先全量校验再全量写**
—— 名字或值有一个不合法，一个都不会写。它省掉的是 N 遍那条 100 多字符的 name-path。

| | 单属性 | 多属性 |
|---|---|---|
| 规格（`.tsd`） | `set_spec_attr` + `attr`/`value` | `set_spec_attrs` + `attrs:{…}` |
| 布局（`.4fd`） | `set_layout_attr`（也能同属性刷多节点，用 `paths`） | `set_layout_attrs` + `attrs:{…}` |

`attrs` 的值**必须是字符串**（属性在文件里都是文本）：`{"gridWidth":"20"}`，不是 `{"gridWidth":20}`。
复数形式合成**一步撤销**只对 `set_spec_attrs` 成立；`set_layout_attrs` 是 N 步 —— 布局写入必须走
`XmlElement` 索引器（§11.24(d) 1），而它自己造命令，我们拿不到那个命令来做分组。返回体的
`undoSteps` 是**实测的**（撤销栈前后差），不是承诺的。

**读的动词各自干什么**：`form_tree` 看层级与 name-path；`get_component` 看一个节点的属性与规格；
`describe_kind` 给**这一类节点此刻能写哪些属性**（白名单是每个包现算的，`attr` 必须照着它写，
否则撞 `E_ATTR_NOT_WHITELIST`）；`list_spec_nodes` 列字段/动作等规格节点；`list_tables` /
`list_columns` 查数据字典；`list_records` / `list_local_strings` 看记录与多语言。

**两个返回形状不一样，别按一个猜**：`get_component` 的 `spec` 是**按 kind 分层的**
（`spec.field.attrs.can_edit`），而 `layout` 是**扁平的**（`layout.noEntry`，没有 `.attrs`）。

**规格动词怎么寻址**（这一条不写下来会白撞一次）：

- `set_spec_attr` / `set_spec_attrs` 要的 `path` **是那个布局元素的 name-path**，再用 `kind` 指出
  它身上挂的哪个规格节点。规格节点活在 `.tsd` 里，本身**没有** name-path。
- `list_spec_nodes` 回的是 `{kind, name, status}`，**不含 path** —— 它不是寻址工具，
  是"这张表单有哪些规格节点"的清单。要 path 用 `find_component`。
- **path 的根不等于 `--form` 的程序名**：`--form apmt500_wf` 但 path 是
  `managedform/`**`apmt500`**`/HBoxT1/…`。拿程序名去拼 path 会 `E_NOT_FOUND`。照 `find_component`
  回的那条原样用。

**几处容易记混的**：`move` **只改 Z 序**（同一父容器内前后挪），换父容器要用 `reparent`
（设计器的拖拽命令，仅限同表单）；`add_field` 是加字段、`wrap` 是拿选中元素**包一层新容器**；
`add_field` 的 `columns` 一次构造多列，一次一列会让大表单慢两个数量级（实测 84 列从 137 秒降到
1.17 秒）。

## 4. 调用形状（所有动词通用）

### 4.1 动词名怎么找

```powershell
tt dev tzs --help            # 静态用法 + 按组列出动词名（引擎可达时；[工作流] 排在最前）
tt dev tzs <动词> --help     # 该动词的示例 + 参数 + 错误码
tt dev tzs nudgee            # 敲错：退 2，并把完整动词清单打到 stderr
```

两级写法与连字符都认：`field add` = `field_add`、`form-tree` = `form_tree`。**没有"打印整张函数表"
的命令** —— 参数表是每个动词自己的契约，一次全摊开只会淹没人。

### 4.2 参数只有一种给法：JSON

```powershell
tt dev tzs <动词> --args '{"<参数>": <值>}' [--json]   # 数组就是数组、布尔就是布尔、数字不带引号
tt dev tzs <动词> --args-file <UTF-8 文件>              # 长内容/中文；值写 - 表示读标准输入
tt dev tzs list_open --json                            # 无参数的动词可以省掉 --args
```

没有 `--handle h9` 这种写法：参数名与类型是**运行时**从引擎 manifest 来的，写成 flag 就得让调用方
去学一套只存在于命令行的语法（逗号切数组、能不能重复、负号算不算值…）。JSON 里这些都不存在。
**中文/长内容一律走 `--args-file`**：Windows 管道可能按控制台代码页重编码。

### 4.3 寻址：`--form <程序名>`，不要搬运句柄

```powershell
tt dev tzs set_spec_attr --form aapp320 --args '{"path":"…","kind":"field","attr":"can_edit","value":"N"}' --json
```

`--form` 写**程序名**（`aapp320`）或 **ProgramKey**（`aapp320|Form`），引擎自己解析成会话 ——
`open` 的返回里虽然有 `handle:"h1"`，但你可以全程忽略它。想看现在开着哪些：`list_open --json`。

- `--form` 与 args 里的 `handle` **只能给一个**（不做优先级猜测）；
- `--form` 只对**需要句柄**的动词有意义，给 `list_open` 这类会报「不接受 form」；
- 名字不存在 → 退 2，错误里列出当前开着的候选（含可用的 `key`）并说明三种写法；
- 程序名不唯一 → 报 `bad_param` 让你改用 ProgramKey，**不猜**。

### 4.4 路径与工作区：包必须**在工作区目录之下**

工作区不是包的属性，是**目录**的属性。设计器只做一次字符串前缀比较（`TzpManager.InCurrentWorkspace`
→ `ConnectionSetting.InWorkspace`）：

> 取**包所在的目录**，两边补上尾 `\`，看它是否以 `<工作区>\` 开头（大小写不敏感）。是就放行，否则抛
> `NotInCurrentWorkspaceException`。

由此而来的四条，都实测过：

| | |
|---|---|
| 包内容与此无关 | `.tzs` 里**没有任何**工作区/路径信息（5 个条目，只有一个 4 字节的 `ver` 装着发行版本号）。同一份包，放对目录就开，放错就拒。 |
| 子目录可以 | `<工作区>\任意\子\目录\x.tzs` 也过 —— 前缀语义如此。（兄弟目录不行：`…\prd2\` 不以 `…\prd\` 开头。） |
| **`out` 必须也在工作区内** | 闸只装在**加载**路径，写盘不查。`save`/`field_add` 往区外写会**成功返回**（退 0），那个包之后 `open` 才被拒。 |
| **路径必须绝对** | 相对路径由**守护进程**的 cwd 解释（它继承自第一次调用它的那个 `tt` 进程），不是你的 cwd；裸文件名更必拒（`Path.GetDirectoryName("x.tzs")` 是空串）。 |

守护进程 `Boot` 一次就绑死一个工作区、不可重指，所以 **`field_add` 的 `file` 与 `out` 必须同属一个工作区**。
"改 A 模块的表单、存到 B 模块"在无头这条路里做不到 —— 那需要两个进程。

### 4.5 传输开关（不是动词参数）

`--form` / `--args` / `--args-file` / `--workspace` / `--rpc-timeout <秒>` / `--json` / `-h`。

### 4.6 慢动词会有动静

超过约 0.4 秒先在 **stderr** 打一行 `… 正在执行 validate；还在等引擎，慢是正常的`。
那是提示不是错误；**数据仍只在 stdout**。

### 4.7 先过滤：清单类动词不给过滤会回一大块

**最容易白花 context 的一处**（实测）：`list_columns --table pmdl_t` 不给 `query` 会回 109 列的
完整元数据 **44,655 字节**；`"query":"pmdl00"` → **3,833 字节**；`"query":"site"` → **517 字节**。
动词自己的 `--help` 里有这句提示；调用后如果回了一大块又本来能收窄，stderr 还会提一句。
**会改模型的动词不谈"收窄"** —— 它们的参数是输入，不是过滤器。

## 5. `validate` 与"增量"

`validate` 在**一个会话上第一次被调用时，那次运行本身就是基线**（`Fns/Validate.cs:139`），
所以首调 `newErrors`/`newWarnings` **按构造就是空**，它证明不了任何事。

实测 `aapt300(c).tzs`（**未做任何改动**）：首调 `baseline = 13 条 WARNING`、`after = 13`、
`newErrors = 0`、`newWarnings = 0`（耗时 5,029 ms）；第二次调用 3,379 ms 起才是"相对基线的增量"。
**语料本来就不干净** —— 判据永远是「改动后的增量」，不是「有没有 WARNING」。

返回体 `{baseline[], after[], newErrors[], newWarnings[], elapsedMs}`；`field_add` 会把这一坨
**瘦身**成计数 + 增量（见 §2）。`--json` 打的是**整帧** `{id, ok, result, error, ms}`，不是只打 `result`。

## 6. 会话与句柄（你只需要知道的最小集）

- **句柄永不复用**：`close` 后再 `open` 同一个包拿到的是新号。旧号去调 → `E_NOT_FOUND`（退 2，安全失败）。
- **同一个程序已经开着又 `open` → `E_KEY_IN_USE`（退 4）**：先 `close` 占用者（`list_open` 看是谁），
  或对 `Loaded` 状态的占用者用 `"force":true`。**用 `field_add` + `file` 不会有这个问题**（已开着就复用）。
- 会话只活在常驻守护进程里：**进程一死全部失效**，重新 `open` 即可。
- **请求一旦上线绝不重试**：协议没有幂等键，这些动词都在改设计器内存里的模型，重试是在赌
  「上一次写进去了没有」。
- `stop` 停本工作区的常驻引擎（**绝不 spawn**）；`reap [--yes]` 收引擎**重编后**停不掉的孤儿
  守护进程（管道名含 MVID，重编即换名）。`reap` 不加 `--yes` 只列不杀。

## 7. 本地校验口径（语法在本地，语义交给引擎）

| 情况 | 行为 |
|---|---|
| 未知动词 / 未知参数 / 缺必填 | 退 2 并点名（未知参数**一次报全部**） |
| int 收到 `"2"` 或 `1.5`、bool 收到 `"yes"`、枚举越界、数组元素非字符串 | 退 2，点名参数与合法值 |
| `attrs` 收到非对象 / 值不是字符串 / 空对象 | 退 2，点名是哪个键 |
| 参数写 `null` | 视同**没给**（缺必填照报） |
| 参数写成 flag（`--handle h9`） | 退 2「未知开关」 |
| 寻址写进 JSON（`"form":"aapp320"`） | 退 2「参数 form 不在 manifest 里」+ 提醒用 `--form` |
| `--form` 与 `handle` 同时给 / 给不需要句柄的动词 | 退 2（不猜优先级 / 不接受 form） |
| **`kind` / `attr` 的取值、`add_action` 的 `type`** | **本地放行**，由引擎用它自己的白名单/该表单的 `s_detail<n>` 拒绝（退 2，`detail.legal` 给合法集） |
| **布局属性的「值」** | **引擎按工作区的 `mta/mod-fd.spec` 拒**（退 2，`detail.legal` 给值集、`detail.hint` 给近似值） |
| **`req` / `can_edit` / `can_query` 的值** | **引擎只收 `Y` / `N` / 空串**（退 2，`detail.legal`）。这三个是设计器面板上的勾选位，**不要写 `"true"`** |

最后三行是**故意的**：那几处的合法集要从**活着的模型**、**该表单自己的**记录、或**工作区的规范文件**里取，
静态判不了。在本地拦下来只会让一个引擎本会接受的调用永远到不了引擎，而且拦得静悄悄（写错了不会有测试失败）。

**`req` / `can_edit` / `can_query` 只写 `Y` / `N`（或空串）。** 这三个值得单独说，因为写错**不报错**：

- 面板的复选框走 `CheckedValueConverter`，它的 `ConvertBack` 只写 `"Y"` 或 `"N"`。
- 但它的 `Convert`（值→勾选态）认 `"Y"` **和 `"TRUE"`**（大小写不敏感）—— 所以 `can_edit="true"`
  在面板上**显示为勾着**。
- 而 `SpecNodeTransform.TransformCanEdit` 是**精确比较**：
  `formElement.SetAttribute("noEntry", ("Y" == specFieldNode.CanEdit) ? "false" : "true")`。
  于是 `can_edit="true"` → `noEntry="true"` → **面板说能编辑，运行时说不能编辑**。

语料（65 个包的 `.tsd`）里这三个属性只出现过 `Y` / `N` / 空串，从没有别的写法。**写别的值引擎现在会拒。**

**值校验只覆盖布局侧**（`set_layout_attr` / `set_layout_attrs`），依据是 `<工作区>/mta/mod-fd.spec` 的
`<PropertyInfo type=… editorInfo="contains:a|b|c">`。三条边界要知道：

- **空值永远放行** —— 语料里 `hidden=""` 出现 2649 次，它不在自己的声明集里，但表示"没设/继承基础数据"。
- **spec 侧不查**（`set_spec_attr` / `set_spec_attrs`）：`mod-fd.spec` 是**表单设计器**的规范，写的是
  `.4fd` 元素属性；`.tsd` 规格属性是另一套，值集也不同 —— `widget="Label"` 在语料里出现 94 次，
  而声明集里没有 `Label`。
- **`range:` 不查**（`gridWidth` 是 `range:0|4000`）：下限已经由设计器夹住并回 `E_ATTR_CLAMPED`，
  上限是模型限制还是 UI 输入框限制没有依据。

参数打错的报错**也要引擎在**（命令先拉 manifest 再校验参数）。`--rpc-timeout` **最小 120 秒**
（下限是给加载看门狗留的）；引擎自己的加载看门狗是另一个参数：`open --args '{"path":"…","timeout":30}'`。

## 8. 退出码

**没有 `3`**（`3` 是 `.tzc` 那条线的「验证失败」），多一个 `1`：

| 码 | 含义 |
|---|---|
| `0` | 帧 `ok:true` |
| `1` | 引擎内部错（`kind=internal`，含 `E_NOT_IMPLEMENTED`） |
| `2` | 参数/环境不对：本地参数错、未知动词、manifest 拉不到、引擎的 `validation` 与 `not_found` |
| `4` | **设计器拒绝**（`kind=designer`，含 `E_KEY_IN_USE`） |
| `5` | 传输或环境失败（含加载超时、没配工作区、`--args-file` 读不到） |

## 9. 只读解压：`export`

```powershell
tt dev tzs export "D:\\ws\\aapt300(c).tzs"    # 纯解压到 <包目录>\aapt300-unzip（只读参考）
```

`export` **不依赖引擎也不依赖设计器**；`-o <dir>` 换落点，目标非空时拒绝（`--force` 覆盖同名文件）。
产物就是**一包文件**（不是 `tzc` 那种带围栏的工作区、不能 apply）；要改表单走动词，不要手工改完塞回去。

## 10. 动词全表（52 个）

`tt dev tzs --help` 会列出它们（`[工作流]` 在最前；会改模型的标 `[写]`，慢的标 `[slow]`）：

| 组 | 动词 |
|---|---|
| 工作流 | `field_add` |
| 会话 | `open` `save` `close` `verify` `list_open` |
| 读 | `form_tree` `find_component` `get_component` `list_spec_nodes` `describe_kind` `list_tables` `list_columns` `list_records` `list_local_strings` |
| 属性 | `set_spec_attr` `set_spec_attrs` `set_layout_attr` `set_layout_attrs` `set_tree_source` `rename_component` |
| 结构 | `add_widget` `add_field` `insert_at` `delete` `move` `reparent` `nudge` `align` `fit_size` `wrap` `break_layout` `convert_widget` `convert_container` |
| 页签 | `add_page` `delete_page` |
| 语义/Action | `insert_semantic` `add_action` `delete_action` `set_action_types` |
| 多语言/选项/串查 | `set_local_string` `set_items` `set_progrel_programs` `set_table_association` `set_spec_description` `set_cited` |
| Tab 顺序 | `set_tab_order` `tab_action` |
| 校验/工具 | `validate` `base_data` `set_excluded` `set_code_template` |

**这 52 个来自引擎的函数表。另有 4 个是 `tt` 自己的内建命令，不在表里、也不接受 `--help`：**
`export`（§9，本地解压，不用引擎）、`doctor`（自检）、`stop`、`reap`。`tt dev tzs stop --help`
会退 2 —— 它走的是另一条路径。

## 11. 常见错误（自查表）

**任务级**
- ❌ 加字段还自己一步步串（挑列 → 开包 → 探容器 → 两遍 validate → save）→ 直接用 `field add`；
  手工链只在要精细控制时才走（§3）。
- ❌ `field_add` 没给 `table` 或 `columns` → 退 2 点名缺哪个。
- ❌ 让它猜容器，它报"挑不出容器" → 那是**故意的**；给 `into`（`form_tree` 看 name-path）。
- ❌ 给 `field_add` 的 `container` 写 `HBox`/`VBox`/`Page`/`Folder` → 函数体不收（见 §2 的警示）。

**参数**
- ❌ 布尔写字符串、数字写字符串、数组写逗号串 → 退 2（`"excluded":"yes"` / `"offset":"2"` / `"paths":"a,b"`）。
- ❌ 参数写成 flag（`--handle h9`）或用位置参数（裸词）→ 退 2。参数一律进 `--args` 的 JSON。
- ❌ 中文/长内容直接写在命令行里被重编码 → 用 `--args-file` 写 UTF-8 文件。
- ❌ 给 `path` / `file` / `out` 写相对路径 → 由**守护进程**的 cwd 解释（不是你的 cwd），裸文件名必拒；一律写绝对路径（§4.4）。
- ❌ `list_columns` / `list_local_strings` 不给过滤直接打 → 可能回几十 KB（先看 `--help` 的"先过滤"提示）。
- ❌ `attrs` 写成数组、写空对象、或值不带引号（`{"gridWidth":20}`）→ 退 2，点名是哪个键；值一律是字符串。

**改属性**
- ❌ 一次只改一个属性、还一条条发 → 用复数形式（`set_spec_attrs` / `set_layout_attrs`），省掉 N 遍 name-path。
- ❌ **猜布局属性的值** → 引擎按工作区的 `mta/mod-fd.spec` 拒（退 2，`detail.legal` 给值集、`detail.hint` 给近似值）。
  实测过的坑：`case="UPPER"`（合法的是 `upper`）、`scroll="MAYBE"`（BOOLEAN 只收 true/false）。
- ❌ 把布局属性和规格属性混在一个动词里 → 两个面板、两个文件、两个动词，混用会被白名单拒
  （`case` 是布局、`can_query` 是规格）。
- ❌ 复数形式改一半失败还当成功 → `detail` 里有 `applied` / `failed` 两栏，`E_ATTR_PARTIAL` 说明
  模型已经不是原样了；只重发 `failed` 里那几个。

**寻址与会话**
- ❌ 需要句柄的动词既没给 `--form` 也没在 JSON 里给 `handle` → 退 2「缺必填参数 handle」。
  **正解是 `--form <程序名>`**，不必抄句柄。
- ❌ `--form` 与 `handle` 同时给 / 给 `list_open` 传 `--form` / 把 `form` 写进 JSON → 各退 2（文案会指路）。
- ❌ 用已经失效的程序名（`close` 过、或守护进程重启过）→ 退 2 并列出当前候选；重新 `open`。
- ❌ 同一个程序已经开着又 `open` → 退 4 `E_KEY_IN_USE`（或改用 `field_add` + `file`，它会复用）。

**读与校验**
- ❌ 第一次 `validate` 报 `newErrors: []` 就以为改对了 → 首调**就是建基线的那次**，零增量是构造性的。
- ❌ 拿 `baseline` 的 WARNING 数当"改坏了" → 语料**本来就不干净**（aapt300 未改动即有 13 条）。
- ❌ `form_tree` 用 `--depth 1` 想看子节点 → 那是"只有这个节点本身"，至少 `2`。
- ❌ 不看 `describe_kind` 就写 `set_spec_attr` → 撞 `E_ATTR_NOT_WHITELIST`；白名单是每个包现算的。

**红线**
- ❌ 把 `save`/`field_add` 的 `out` 指向**源包** → 唯一会毁掉原始素材的写法。永远写**新**包。
- ❌ 把 `out` 写到工作区**外** → 写的时候退 0 一切正常，那个包**之后** `open` 才被设计器拒（§4.4）。
- ❌ 把 `export` 的产物当工作区去 `apply` / 手工改完塞回包 → 它只是只读参考。
- ❌ 在没有工作区的地方跑运行类动词 → 引擎缺省工作区是真实客户目录，拒绝启动是保护你的（别绕过）。
- ❌ 请求超时后重发 → 不重试；重新 `open` 或直接看结果。
- ❌ 看到 stderr 的 `… 正在执行 validate；还在等引擎，慢是正常的` 就以为出错 → 那是提示行。

## 12. 背景与细节

- 引擎为什么单独构建（不在 Go 构建链里、设计器目录是构建期与运行期都要的依赖、重编会让在跑的
  守护进程变孤儿）：[engine/BUILD.md](../../engine/BUILD.md)
- `.tzs` 格式与引擎契约：[engine/SPEC.md](../../engine/SPEC.md)
- 命令面、验收清单、设计依据：[docs/WIKI.md](../../docs/WIKI.md#7-tt-dev-tzs-表单包)
