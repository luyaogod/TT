---
name: tt-dev-tzs
description: 读写 T100 设计器**表单包**（.tzs/.tzv）：由设计器自己的引擎驱动（55 个具名动词，命名管道 JSON-RPC），参数一律用 JSON 给，不是拼 XML。按数据表加字段用任务级动词 field_add（一条命令做完挑容器+建字段+校验+另存）；改属性用 set_spec_attr / set_layout_attr，改多个用它们的复数形式（一次请求、先全量校验再全量写）；布局/页签等用细粒度动词（open → 读 → 改 → validate → save）。要改表单、按表加字段、查表单结构或字段时用。**前提：先配工作区，且包与 out 都必须用绝对路径落在工作区目录之下。代码包（.tzc/.tzf/.tzx）不归这里，用 tt-dev-tzc。**
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
| **按某张表的列加字段**（最高频） | **`field_add`** —— 挑容器 + 建字段 + 报校验增量 + 存新包，一条命令（§2） |
| 改属性 / 挪布局 / 换容器 / 页签 / 多语言 / Tab 顺序 | **细粒度动词链**（§3）—— 先看清结构，再改一处 |
| 只看结构、字段、开着的表单 | `form_tree` / `find_component` / `get_component` / `list_spec_nodes` / `list_columns` / `list_open` |
| **还不知道该开哪个文件** | `list_packages`（按程序名筛，见 §1 那段） |

**跑之前要配一样东西**：工作区（`tt config set tzs.workspace "D:\ws"` 或环境变量 `TZSCLI_WS`；
**本文档的 `D:\ws` 一律指工作区**）。**这条没有缺省、也不回落** —— 引擎内置的默认工作区是一个
**真实客户目录**，落上去等于拿别人的表单当草稿纸；三层都空时运行类动词直接拒绝启动（退 5）。
工作区还**圈定了你能碰哪些包**：包的目录必须在它之下，`out` 也一样（§4.4）。

**"工作区在哪"从哪问**（`out` 必须是绝对路径，所以这条迟早要用）：`tt dev tzs doctor` 的输出里
就有当前工作区的**绝对路径**那一行；`list_packages` 的返回里也有一个 `workspace` 字段。
（两处都实测过。文档从前只说"工作区配在 config 里"，没说在命令行里怎么看它。）

**包就在工作区里，而"程序名"不是文件名**：`aapt300` 的包叫 `aapt300(c).tzs`
（`(c)` = 客户版、`(s)` = 标准版，一般读写用 `(c)`）。**要问"哪个文件是 aapt300"就用
`list_packages`**（不需要句柄，与 `list_tables` 同族）：

```bash
tt dev tzs list_packages --args '{"query":"aapt300"}' --json
# → {"count":2,"returned":2,"packages":[{"name":"aapt300(c).tzs","path":"…","program":"aapt300",
#                                        "sizeBytes":57818,"modified":"2026-09-25 12:00:00"}, …]}
```

`query` 是**文件名子串**，`limit` 默认 200。里面的 `program` 是**从文件名推的**
（`<程序名>(c).tzs` 那条约定），拿来挑文件用；**权威答案永远是 `open`**（它回的 `program`
是设计器自己算的）。手上有程序名时，照约定直接拼 `<工作区>/<程序名>(c).tzs` 也一样快。

> ⚠️ **你 `save` 出来的包会被标错 `program`**（实测）：`out` 叫 `_ai_e2.tzs` 时，
> `list_packages` 推出来的 `program` 就是 `"_ai_e2"`，而那包里真正的程序是 `aapt300`
> （`save` 出的新包**沿用源包**的 program，§2）。这不是 bug —— 它只按文件名推 ——
> 但意味着**别拿 `list_packages` 去找自己刚存的中间包**；按文件名找它就行。

> ⚠️ **别裸 `ls <工作区>`。** 一个真实工作区里几十上百个文件（`.tzs`/`.tzc`/`.bak` 混着），
> 实测一次回 **29.9 KB** —— 那是纯噪音，而 §4.7 那条"先过滤"的规矩正是对同一件事立的
>（`list_packages` 就是为了替掉这一步而加的）。非要看目录也按名字收窄：`ls <工作区>/<程序名>*`。
>
> （`open` 只收 `path`；`--form` 收的是**程序名**，两个不是一回事。）

配置**文件**的位置可以用环境变量 `TT_CONFIG=<config.json>` 指定（便携版、或"不想动默认配置"
时用）；`tt config path` 会告诉你当前实际读的是哪一份。工作区写在那个文件的 `tzs.workspace` 里。
设计器**不用配**（它的程序集随仓库与发行包自带）。开工前先 `tt dev tzs doctor` 自检。

> ⚠️ **`TT_CONFIG` 的值在 bash / Git Bash 下也要写正斜杠（或整个加单引号）。**
> `export TT_CONFIG=C:\Users\…\config.json` 会被 shell 把反斜杠**全吃掉**（变成
> `C:Users…`），于是配置根本没被读到 —— 而症状看起来像**环境装坏了**（`doctor` 报"工作区
> 未配置 / 引擎 exe 不存在"），不像"你的环境变量写错了"。两种写法都对：
> `export TT_CONFIG='C:\Users\…\config.json'` 或 `export TT_CONFIG=C:/Users/…/config.json`。
> 与 §4.2 那条 `--args` 里的路径是同一个坑。

## 2. 任务级动词：`field_add`

一次做完：**挑容器 → 按列建字段 → 报校验增量 →（给了 `out` 就）存新包**。

```powershell
# ① 最省事：只给 file（没开着就顺手开，已开着就复用 —— 不会撞 E_KEY_IN_USE）
tt dev tzs field_add --args '{"file":"D:/ws/apmt500_wf(c).tzs","table":"pmdl_t","columns":["pmdlent","pmdlsite"],"out":"D:/ws/_ai.tzs"}' --json

# ② 表单已经开着：用程序名寻址，不必抄句柄
tt dev tzs field_add --args '{"handle":"apmt500_wf","table":"pmdl_t","columns":["pmdlent"]}' --json

# ③ 容器不想让它挑：显式给 into（name-path 用 form_tree 看）
tt dev tzs field_add --args '{"file":"D:/ws/x.tzs","table":"pmdl_t","columns":["pmdlent"],"into":"managedform/x/HBoxT1/worksheet"}' --json
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

> `container` 的枚举**就是函数体校验的那一份**（2026-09-25 起，同一个数组）：
> `None` / `Grid` / `Group` / `ScrollGrid` / `Table` / `Tree` —— 照着 `--help` 写一定过得去。
> `HBox`/`VBox`/`Page`/`Folder` 从来不是**创建目标**（它们在设计器里只作为命令存在），
> 从前 `--help` 把它们列出来、函数体再拒 —— 那是声明那一侧的错，现在两边是同一份。
> `convert_container` 的 `type` 是另一件事：只收 `Grid` / `Group`（设计器的转换命令只能在这两者之间转）。

**容器怎么自动挑**（`into` 可覆盖）：① 叫 `worksheet` 的容器（设计器模板里放字段的那块）→
② 名字含 `layout` 的容器 → ③ 根 `<Form>` 下唯一的容器 → ④ 都不成立就**报错要你显式指定**
（宁可不做，也不把字段塞进一个猜出来的盒子里）。

**它返回什么**（刻意瘦，约 1.4 KB）：`form` / `key` / `handle` / `opened`（这次是不是它开的）/
`container` / `added`（**对象**：`{tag,name,path,container,table,columns,specNodeType,promoted,bound,added[]}`）/
`validate`（`newErrorCount`/`newWarningCount` + 明细）/
`baselineCached` / `saved`（与 `save` 的返回同一个形状，见下）。**不回表单全量，也不回校验的
baseline/after 全表** —— 要看全量就单独调 `validate`。

> ⚠️ **`saved` 里有两条路径，别拿错**：`saved.path` 是**源包**（它是那次 `save` 的输入），
> `saved.out` 才是**这次写出来的新包**。要比 sha256、要 `open` 回读，用的都是 `saved.out`
> —— 拿 `path` 去比会变成"拿源包跟自己比"，看起来还挺像对了。

> ⚠️ **`added` 里有两个 `container`，含义不同**（实测）：
> 顶层的 `added.container` 是**容器模式字符串**（`"None"`/`"Table"`…，与那个**参数** `container` 同义），
> 而它旁边的 `added.path` 才是**被挑中的容器**的 name-path。`promoted` = 被提升/补建的节点数、
> `bound` = 真正绑到表列的字段数。
>
> ⚠️ **里层 `added[]` 每个节点带自己的绑定**（2026-09-25 起）：字段节点有 `table` 与 `column`
> （如 `{"path":"…/pmdl_t.pmdlent","name":"pmdl_t.pmdlent","tag":"Edit","table":"pmdl_t","column":"pmdlent"}`），
> 而配套的 `Label` 节点**两个都没有** —— **缺席本身就是信息**：它是同伴，不是字段。
> 外层那个 `columns` 是**你请求的那份列清单**，里层每个节点上的 `column` 才是**它自己绑到哪一列**。
> （要更全的绑定信息 —— 比如 `fieldType` —— 仍然回读：`find_component` 会给。）

> **`field_add` 的 `validate` 增量是有效的**（与手工链不同）：它内部**先建基线、再加字段、再跑一次**
> （`Fns/Session.cs`：`Validate.Run` 在 `AddFieldFn` 之前），所以 `newErrorCount`/`newWarningCount`
> 回的就是**这次加字段引入的**新问题。`baselineCached` 只告诉你"会话里本来就有基线吗"：
> `true` = 基线是**你的**（或上一轮留下的），`false` = 基线是它刚建的 —— 两种情况下增量都算数。
> （2026-09-25 更正：上一版文档写的是"`false` 时那个 0 什么都没证明"，那是把 `field_add` 与手工链
> 搞混了 —— 手工链的**首调**才是"建基线那次"。判"改对了"仍然用回读，但**别忽略这个计数**。）

**"落了盘没有"和"盘上内容对不对"是两个问题，各有便宜的问法。**

| 你想知道 | 怎么问 | 代价 |
|---|---|---|
| 改动**落了盘**（文件里就是引擎刚写的那份字节） | `sha256sum <out>` 与 `save` 返回里的 `sha256` 比一次 | **一条 shell 命令，不碰引擎** |
| 盘上那个包**内容**是什么 | `close` 源会话 → `open <新包>` → 读 | 三次引擎调用，但看得到内容 |
| **源包**有没有被碰（§11 的红线） | 开工前 `sha256sum <源包>`，收工再比一次 | 两条 shell 命令，不碰引擎 |

前者为什么成立：`save` 渲染的就是**当前模型**（同一个包再存一次，逐条目 sha256 应当一致 ——
语料回归在每个真实包上守着这条），
而 `sha256` 是它刚写下去那份字节的摘要 —— 两个已知事实一叠，"盘上就是这个包"就闭合了。
（`get_component` / `verify` 证明不了这一条：它们读的都是**内存模型**，而 save 正是把那个模型写盘。）

**要 `open` 新包就必须先 `close`**：新包 **program 名与源包相同**（`aapt300(c).tzs` → `_ai.tzs` 都叫
`aapt300`），而 `open` 的 key 是 `程序名|Form`、**不含路径** —— 不先 `close` 源会话就撞
`E_KEY_IN_USE`（退 4）。`save` 的返回里**就写着这个 key**（`key` 字段）与一句 `note`，
不必自己去推。顺序是 `close --form <程序名>` → `open --args '{"path":"<新包>"}'`，
拿到的是**新句柄**（§6：句柄永不复用）。

**边界**：它只做"按表的列加字段"。改属性、挪布局、页签、多语言仍是细粒度动词 —— 走 §3。

## 3. 细粒度：一次调用改一处

> **下面这些例子都写成裸 `tt dev tzs …`** —— 它们假设 `tt` 在 PATH 上、工作区配在默认配置里。
> 便携版 / 用 `TT_CONFIG` 指向另一份配置 / `tt` 不在 PATH 时，每条示例前面都要加它的前缀：
>
> ```bash
> export TT_CONFIG=C:/path/to/config.json     # 正斜杠，见 §1 的警示
> TT=C:/path/to/tt.exe                        # 换掉示例里的 `tt`
> ```
>
> 这是本次评测里**每个执行者**都要做的一步（他们的提示词就长这样），写在这里省掉那一遍改写。

```powershell
tt dev tzs doctor                                        # 先自检：引擎/设计器目录/工作区/管道名
tt dev tzs <动词> --help                                 # 参数表 + 一条**带占位符的**示例（<包路径>/<path> 要你自己换）

# 正解顺序（全程 --form 寻址，不搬运句柄）：
#   open → 读 → validate(建基线) → 改 → validate(增量) → save → 落盘证明
# 落盘证明有两种问法，按你要问什么选（§2 有表）：
#   · "文件里就是引擎刚写的那份字节" → save 返回的 `sha256` 与 `sha256sum <out>` 比一次（不碰引擎）
#   · "盘上那个包的内容" → close → open(新包) → 读（三次调用，且必须先 close）
# 先看一眼再动手（可选）：写动词都收 dry_run，见 §6 —— 真跑一遍、答案在 preview 里、模型放回去
# 反悔：reload 丢内存改动、从盘重读（保 validate 基线），见 §6
# 超时了不确定写没写进去：给写起个 op 名，事后 list_ops 问它，见 §6
# 三处容易走错的：
#   · validate 想报增量，第一次必须排在**改之前**（§5：一个会话上的首调就是建基线那次）
#   · `save` 出的新包 program 名与源包**相同**（返回里的 `key` 就写着它），要 open 它必须先
#     close，否则 E_KEY_IN_USE（§6）
#   · `get_component` 与 `verify` 读的都是**句柄里的内存模型**，而 save 正是把那个模型写盘
#     —— 拿它们回读证明不了**盘上**的字节，那是自证循环
tt dev tzs open           --args '{"path":"D:/ws/aapt300(c).tzs"}' --json   # → {"program":"aapt300",…}
tt dev tzs form_tree      --form aapt300 --args '{"depth":3}' --json            # 结构树，每节点带 name-path
tt dev tzs find_component --form aapt300 --args '{"query":"worksheet"}' --json  # 控件代号 → name-path
tt dev tzs get_component  --form aapt300 --args '{"query":"worksheet"}' --json   # 代号直取，省掉一次 find_component
#   （重名时它退 2 并列出候选让你改用 path —— 不会替你挑；找不到退 2 E_NOT_FOUND）
tt dev tzs describe_kind  --form aapt300 --args '{"kind":"field"}' --json       # 这类节点**运行时**能写哪些属性
tt dev tzs describe_kind  --form aapt300 --args '{"kind":"layout"}' --json      # 布局侧：名字 + 类型 + 取值集（约 9 KB）
tt dev tzs validate       --form aapt300 --json                                 # ① 建基线（大包慢：aapt300 实测 3–5 秒；aapp320 那种 0.6 秒）
tt dev tzs set_spec_attr  --form aapt300 --args '{"path":"<path>","kind":"field","attr":"can_edit","value":"N"}' --json
tt dev tzs set_spec_attrs --form aapt300 --args '{"path":"<path>","kind":"field","attrs":{"can_edit":"Y","can_query":"N"}}' --json
# 布局侧少一个 kind（规格节点才分 kind，布局元素不分 —— §7 有对照表）：
tt dev tzs set_layout_attr  --form aapt300 --args '{"path":"<path>","attr":"hidden","value":"true"}' --json
tt dev tzs set_layout_attrs --form aapt300 --args '{"path":"<path>","attrs":{"hidden":"true","gridWidth":"20"}}' --json
tt dev tzs validate       --form aapt300 --json                                 # ② 真增量（①已经建过基线）
tt dev tzs save           --form aapt300 --args '{"out":"D:/ws/_ai.tzs"}' --json  # 写**新**包；原包一字节不动
                                                                                #   → 记下返回里的 sha256 与 key
# ③ 落盘证明（二选一；下面这条不碰引擎）
sha256sum D:/ws/_ai.tzs                                                         # 与上一步的 sha256 比一次即证
#   （不在 Git Bash 里就用 PowerShell：Get-FileHash -Algorithm SHA256 D:\ws\_ai.tzs）
# 或者要看内容（三次调用：close + open + 读，且 close 是必须的）：
tt dev tzs close          --form aapt300 --json
tt dev tzs open           --args '{"path":"D:/ws/_ai.tzs"}' --json               # 从盘重载新包
tt dev tzs get_component  --form aapt300 --args '{"path":"<path>"}' --json       # 回读它内容
tt dev tzs stop                                          # 停本工作区的常驻引擎（内建命令，不是引擎动词）
```

> **规格写入（`set_spec_attr` 那类）不算"纯布局改动"**：设计器的 transform 会把它映射到布局
> 属性（`can_edit` → `noEntry`），所以它**有**布局侧副作用 —— 按上面的正解顺序建基线再改，
> 别省那次 validate。（实测过一次：`set_spec_attr` 的返回里 `layoutDelta` 带着它牵动的布局变化。）

> ⚠️ **validate 的增量只对"改动之后的那次"成立**，而一个会话上第一次调用永远是建基线那次（§5）。
> 所以上面那个 ① 不能省：省掉它，改动后那次就是首调，`newErrors` 按构造是空 —— 那不是"没改坏"，
> 是没测。反过来，**纯布局改动可以不跑 validate**（落盘证明 + 回读就是证明，见 ③），
> 它值得跑的时候是你怀疑这次改动有语义副作用。
> validate 只回答"有没有改坏"；"改对了"由**回读**回答，而"落盘了"由 `sha256` 那条一比回答。

**要改多个属性就用复数形式**（`set_spec_attrs` / `set_layout_attrs`）：一次请求、一次寻址、**先全量校验再全量写**
—— 名字或值有一个不合法，一个都不会写。它省掉的是 N 遍那条 100 多字符的 name-path。

| | 单属性 | 多属性 |
|---|---|---|
| 规格（`.tsd`） | `set_spec_attr` + `attr`/`value` | `set_spec_attrs` + `attrs:{…}` |
| 布局（`.4fd`） | `set_layout_attr`（也能同属性刷多节点，用 `paths`） | `set_layout_attrs` + `attrs:{…}` |

`attrs` 的值**必须是字符串**（属性在文件里都是文本）：`{"gridWidth":"20"}`，不是 `{"gridWidth":20}`。
复数形式合成**一步撤销**只对 `set_spec_attrs` 成立；`set_layout_attrs` 是 N 步 —— 布局写入必须走
`XmlElement` 索引器（设计器自己的写入路径，值校验也发生在那里），而它自己造命令，
我们拿不到那个命令来做分组。返回体的
`undoSteps` 是**实测的**（撤销栈前后差），不是承诺的。

**读的动词各自干什么**：`form_tree` 看层级与 name-path；`get_component` 看一个节点的属性与规格
（用 name-path 寻址，**也可以直接给控件代号 `query`** —— 那时它自己解析成路径，省掉
`find_component → get_component` 两次调用；两个都不给会退 2 说明这一点）；`describe_kind` 给
**这一类规格节点此刻能写哪些属性**（白名单是每个包现算的，`attr` 必须照着它写，
否则撞 `E_ATTR_NOT_WHITELIST`）；`list_spec_nodes` 列字段/动作等规格节点；`list_tables` /
`list_columns` 查数据字典；`list_records` / `list_local_strings` 看记录与多语言。

> ⚠️ **要改布局属性（`set_layout_attr` 的 `attr`）时，先问一次 `describe_kind --args '{"kind":"layout"}'`**
> （**§7 已经写死值集的那几个例外**：`hidden` / `req` / `can_edit` / `can_query` 不必先问 ——
> 它们的值与含义在 §7 说死了；其余不眼熟的属性才值得花这 ~9 KB）
> —— 它给的是这张表单上布局属性的**并集**，每条是 `{name, type, values, initial}`：
>
> ```json
> {"name":"hidden","type":"ENUM","values":["false","true"],"initial":"false"}
> {"name":"invisible","type":"BOOLEAN","values":["true","false"],"initial":"false"}
> {"name":"gridWidth","type":"INTEGER"}
> ```
>
> `values` 就是**引擎会接受的那一份**（与 `set_layout_attr` 拒绝你时给的 `detail.legal` 是同一张表，
> 语料回归在每个真实工作区上验它俩一致）；**没有 `values` 键 = 没有任何地方声明取值集**
> （`TEXT`/`FDSTYLE` 之类自由格式，以及 `gridWidth` 这种只有范围的）。所以"这个属性收什么值"
> **不必靠试**：先看这里。实测 `aapt300` 那份并集 132 条、约 9 KB —— 大，但一次调用换掉一串试错。
> 某个元素**具体能用哪些**比这份并集窄，写错时 `E_ATTR_NOT_WHITELIST` 的 `detail.legal` 会给该元素那一份。
> （两个日期各修过一次：2026-09-24 之前函数体实现了 `layout` 分支而参数校验收不下；
> 2026-09-25 之前它只给名字，不给类型与取值集。）

**两个返回形状不一样，别按一个猜**：`get_component` 的 `spec` 是**按 kind 分层的**
（`spec.field.attrs.can_edit`），而 `layout` 是**扁平的**（`layout.noEntry`，没有 `.attrs`）。

**两个写动词的返回：形状按分支不同**（实测两个分支都跑过 —— 上一版这里列的是两个分支的并集，
还漏了 `applied`，照它解析会在写入成功时去找 `noop`）：

```jsonc
// 真的写了（成功帧，退出码 0）
set_spec_attr   // {"path","kind","attr","old","value","specStatus","layoutDelta"}
set_layout_attr // {"path","attr","old","value","written","applied":true,"changed":true,"layoutDelta"}
// 值本来就一样（也是成功帧，退出码 0）
set_spec_attr   // 同上去掉 specStatus 之外，另加 "noop":true,"code":"E_NO_OP","note"
set_layout_attr // 同上，但 "changed":false 且**没有** applied，另有 "noop":true,"code":"E_NO_OP","note"
```

**判"这次写落地了没有"看 `applied`**（写入分支才有）；**判"是不是走了 no-op"看 `noop`**。
两者互斥，各带一个 —— 与 §8 那张三结局表同源。

两处容易读错的：

- **`set_spec_attr`（改规格）的返回里也有 `layoutDelta`** —— 不是笔误：设计器的 transform 会把
  规格属性映射成 `.4fd` 上的属性（`can_edit="Y"` → `noEntry="false"` 就是一例），那个增量报的
  正是**这次规格写入牵动的布局侧变化**。所以"两个面板别混"说的是**参数**别混（`kind`/`attr`
  的合法集不同），不是说改了这边那边一定不动。
- **`specStatus` 是设计器给那个规格节点的状态字母**：`c`=新建 / `u`=改动 / `d`=墓碑（删除）。
  ⚠️ **刚加载的包里也可能一堆 `c`**（实测 `aapt300` 的 269 个字段节点全是 `c`）—— 它不是
  "你这次改过"的标记，只是模型里的现状。`layoutDelta` 的形状是
  `{path, attrs:{改动的属性→新值}, rename}`（没动就是空/`null`）。

**规格动词怎么寻址**（这一条不写下来会白撞一次）：

- `set_spec_attr` / `set_spec_attrs` 要的 `path` **是那个布局元素的 name-path**，再用 `kind` 指出
  它身上挂的哪个规格节点。规格节点活在 `.tsd` 里，本身**没有** name-path。
- `list_spec_nodes` 回的是 `{kind, name, status}`，**不含 path** —— 它不是寻址工具，
  是"这张表单有哪些规格节点"的清单。要 path 用 `find_component`。
- **path 的根不等于 `--form` 的程序名**：`--form apmt500_wf` 但 path 是
  `managedform/`**`apmt500`**`/HBoxT1/…`。拿程序名去拼 path 会 `E_NOT_FOUND`。照 `find_component`
  回的那条原样用。
- 改**程序级**规格描述（SD 描述里的 `all` / `mi_all` / `db_all` / `di_all`）用
  `set_spec_description`，**不需要 `kind`** —— 那四个是程序自己的规格节点，不是某个元素的。
  给它元素路径时才要 `kind`（缺了会退 2 并提示这两种写法）；给错 kind 退 2 `E_NO_SPEC_NODE`
  （`not_found`，不是"设计器拒绝"：换一个 kind 就行）。

**几处容易记混的**：`move` **只改 Z 序**（同一父容器内前后挪），换父容器要用 `reparent`
（设计器的拖拽命令，仅限同表单）；`add_field` 是加字段、`wrap` 是拿选中元素**包一层新容器**；
`add_field` 的 `columns` 一次构造多列，一次一列会让大表单慢两个数量级（实测 84 列从 137 秒降到
1.17 秒）。`delete` / `move` / `nudge` / `align` / `fit_size` / `wrap` 六个的寻址是
**`path` 与 `paths` 二选一**（都可选、至少给一个，函数体判；两者都给时以 `paths` 为准）。

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

> ⚠️ **`<动词> --help` 的参数表里那个 `handle` 是线上参数名，命令行写法是 `--form`。**
> 表只列**动词参数**，传输开关（`--form`/`--args`/`--json`…）不在里面 —— 所以看到
> `{"n":"handle",…}` 不必疑惑"这个动词能不能用 `--form`"：需要句柄的动词一律可以，
> 而且推荐就用它（§4.3）。反过来，把 `handle` 写进 JSON 也可以，但那就得自己抄句柄号。

> ⚠️ **路径写正斜杠，别写反斜杠。** 内联 `--args` 里的 Windows 反斜杠路径
> （`'{"path":"D:\\pkg\\x.tzs"}'`）在 bash / Git Bash 下会被**吃掉一层转义**，引擎报
> `invalid character 'X' in string escape code` —— `X` 是路径里反斜杠后面那个字符
> （实测见过 `'U'`、`'w'`，所以别拿某个具体字母当特征去对号）。单反斜杠和双反斜杠**都**会中招。
> **本节上面所有示例因此一律写成正斜杠**：本文档自己是踩过这个坑之后改过来的（2026-09-25 的
> 评测里，四个执行者全都撞上过旧版示例，其中一个的结论是「知道坑在哪，却把坑留在示例里」）。
> 两条出路，都不用猜：
>
> ① **路径写成正斜杠**：`'{"path":"D:/pkg/x.tzs"}'` ← **推荐**。引擎接受它（包路径、`out`、
> `file` 都是），而且不必和引号层数较劲；
> ② 或者把整份 JSON 写进文件，走 `--args-file`。
>
> 这不是"小心一点就好"的坑：四个互不相干的执行者在这上面各花了额外调用，全都靠试出①才通。

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
| **`out` 必须也在工作区内** | **写盘时也查**（引擎侧闸门，2026-09-24 起）：`save`/`field_add` 往区外写**当场退 2**（`E_BAD_PARAM`，`detail.reason=out-outside-workspace`）。从前是"退 0 成功、那个包之后 `open` 才被拒" —— 那个坑没了。 |
| **路径必须绝对** | 相对路径由**守护进程**的 cwd 解释（它继承自第一次调用它的那个 `tt` 进程），不是你的 cwd；裸文件名更必拒（`Path.GetDirectoryName("x.tzs")` 是空串）。 |

守护进程 `Boot` 一次就绑死一个工作区、不可重指，所以 **`field_add` 的 `file` 与 `out` 必须同属一个工作区**。
"改 A 模块的表单、存到 B 模块"在无头这条路里做不到 —— 那需要两个进程。

### 4.5 传输开关（不是动词参数）

`--form` / `--args` / `--args-file` / `--workspace` / `--rpc-timeout <秒>` / `--json` / `-h`。

### 4.6 慢动词会有动静

超过约 0.4 秒先在 **stderr** 打一行 `… 正在执行 <动词>（<程序名>）；还在等引擎，慢是正常的` ——
动词名与程序名都是**插值**的（实测：`… 正在执行 validate（aapt300）；…`），不是写死 `validate`。
那是提示不是错误；**数据仍只在 stdout**。

### 4.7 先过滤：清单类动词不给过滤会回一大块

**最容易白花 context 的一处**（实测）：`list_columns --table pmdl_t` 不给 `query` 会回 109 列的
完整元数据 **44,655 字节**；`"query":"pmdl00"` → **3,833 字节**；`"query":"site"` → **517 字节**。

还有一处**最该先过滤的**（2026-09-25 实测）：`list_packages` 不给 `query` 会把这一个工作区的
全部包回给你 —— 实测 67 个包 **11.8 KB**，而 `"query":"aapt"` 只剩 5 条。它的 `--help`
里也有这句提示。

另外两个大头（2026-09-25 实测，`aapt300(c)`）：`list_spec_nodes` 不过滤回 **701 个节点 /
29,918 字节**，`"query":"app"` → **11 个 / 555 字节**（差 54 倍）；`list_local_strings` 337 条 /
13,807 字节，`"query":"lbl_apca"` → 84 条 / 3,312 字节，再叠 `"path":"<子树>"` → **6 条 / 398 字节**
（两个参数是**与**关系）。这几个 `query` 是 2026-09-25 才接上线的：引擎比这份 SKILL 旧时
会报「参数 query 不在 manifest 里」—— 那就是这个。

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
  **被拒的调用也吃号**（实测：`open`→`h2`，再 `open` 被 `E_KEY_IN_USE` 拒，第三次 `open`→`h4` ——
  中间那个号给了被拒的那次）。所以**别拿句柄号去数会话数**。
  （想自己复核：记下 `open` 回的号 → 故意撞一次 → 再 `open` 另一个包，看新号比"上一个 + 1"多跳一格。）
- **同一个程序已经开着又 `open` → `E_KEY_IN_USE`（退 4）**。**先问一句"我要的到底是什么"**：
  被拒本身就说明**它已经开着** —— 目标就是"让这张表单开着"的话，**什么都不用做**。
  `close` → `open` 只在你要一个**新会话**时才有意义：源包在盘上变了、你要读的是**另一个**包
  （新包与源包同 program 名，见 §2）、或者你要一个干净的句柄去读回。
  调用方**不需要**先 `list_open` 找它是谁 —— 那个错误帧里已经写着占用者的**完整路径**和处方
  （「先 close 它，或换一个文件」），`list_open` 只在你想看全部会话时才有必要。
  （`open` 曾有一个 `force` 参数，**已于 2026-09-25 从参数表删除** —— 它从未实现，而"接管占用者"
  正是契约禁止的静默驱逐，留着只会引诱人照参数表去试。拿一个被占用的 key 只能先 `close` 再 `open`。）
  **用 `field_add` + `file` 不会有这个问题**（已开着就复用）——
  但 `field_add --out` 存完新包之后，**回读新包**要先 `close`（那新包沿用了同一个程序名，见 §2）。
- 会话只活在常驻守护进程里：**进程一死全部失效**，重新 `open` 即可。
- **请求一旦上线绝不重试 —— 但超时之后可以问。** 协议没有幂等键，这些动词都在改设计器内存里的
  模型，重试是在赌「上一次写进去了没有」。所以：给每次写起个名字（`op`），超时之后用 `list_ops`
  问那个名字，答案只有三种：**没有记录** = 请求从未到达（重发安全）；**`pending`** = 看见了、
  还在做（别盲目重发）；**`ok` / `error`** = 已经结束（重发安全，而且结果摘要就写在里面）。
  ```bash
  tt dev tzs set_spec_attr --form aapp320 --args '{"path":"…","kind":"field","attr":"can_edit","value":"N","op":"task7-a"}'
  tt dev tzs list_ops --args '{"op":"task7-a"}' --json      # 超时之后问这一句
  ```
  ⚠️ 日志**只在守护进程内存里**（它只回答"这个进程看见过什么"）：进程换了就是一份新日志，
  返回里的 `daemonStartedAt` 就是给你判断这件事的。"没有记录"要配上"守护进程没换过"才叫结论。
  不带 `op` 的写不进日志 —— 那就回到"绝不重试"。
- **`reload`：丢掉内存改动、从盘重读，句柄不变。** 这是 `close` + `open` 的替代品，比那条老路
  多两样东西：① **保住 validate 基线**（文件字节与打开时相同时；不同时如实说
  `baseline:"dropped"`）；② **告诉你盘上那个包在你打开之后被换过**（`fileChanged:true` ——
  那之前你手里的基线增量与 diff 都在拿两个不同的文档比）。
  ```bash
  tt dev tzs reload --form aapp320 --json
  ```
- **写动词可以干跑：`dry_run: true`。** 它**真跑一遍**这个动词，把答案原样放进 `preview`
  （所以 `preview.noop` / `preview.applied` 就是"真做的话会是哪种结果"），然后把模型放回调用前。
  想知道"这一改会新增什么校验问题""这个容器挑得对不对"而不想先改再后悔时用它。
  ```bash
  tt dev tzs add_field --form aapp320 --args '{"path":"…","table":"pmda_t","columns":["pmda001"],"dry_run":true}'
  ```
  ⚠️ 代价写在返回里（`reverted`）：回滚是**重建会话**（把模型渲染成一份临时包 → 跑一遍 →
  从那份临时包重读回来；撤销栈回不去这次写，实测见 §11.9 第 18 条），所以**句柄串不变，
  但 `loadMs` 重置、`state` 回到 `Loaded`**。盘上什么都不写。
  `reverted.complete:false` = **回滚没走完**（这时 `reverted.error` 有原因，`reverted.scratch`
  是你调用前那一刻的包，`open` 它就是回到调用前）；`reverted.recovered:true` = 这份临时包
  丢过一次、用留底字节重写过 —— 模型照样回来了，只是这台机器上发生过一次怪事。
  `save --dry-run` 是另一条路（它不改模型，只是不落盘）：照样拼包、照回 `sha256`，`out` 不出现。
  实测开销：一次干跑 ≈ 一次 `open`（render 2–11 ms + load 70–350 ms）。
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

**「是否可编辑」有两个面，改之前先想清要改哪个**（用户嘴里那句话是含糊的）。

> **判决规则**（用户只说了"是否可编辑"、没说是哪个面时）：**按字面最贴的那个 —— 面板勾选位
> `can_edit` 为准**（它就是设计器面板上那个框），改它；**只有当你要声称"运行时也关掉"时才连带
> `noEntry`**。而**改之前先读一眼两边现在各是什么**，已经是"否"的那一面不要写（写了就是一次
> `E_NO_OP` 成功帧，白花一次调用）。下面 `net108` 就是这种情形。

| 说法 | 落在哪 | 谁读它 |
|---|---|---|
| 面板上的勾选位 | 规格 `can_edit`（`Y`/`N`） | 设计器面板；`TransformCanEdit` 由它推出布局的 `noEntry` |
| 运行时真的能不能改 | 布局 `noEntry`（`"true"` = 只读） | 渲染/运行时（`noEntry=="true"` 就是不让改） |

两者**会不一致**，而且这不是理论问题：语料里就有（`aapt300` 的 `net108`：`can_edit="Y"` 而
`noEntry="true"`）。它们是**两个属性、两套动词**（`set_spec_attr` 改前者、`set_layout_attr` 改后者），
面板勾选只驱动后者，所以"把可编辑关掉"要看你想要的是**面板显示**还是**运行时行为**。

**这种不一致不是特例**：`field_add` **新建**的字段默认就会这样 —— 实测新字段的
`can_edit:"N"` 而布局 `noEntry:"false"`（面板说"不可编辑"、运行时说"可编辑"）。所以"加一个
可编辑字段"要做两件事：加字段 + 把规格那一侧改对（§3）。

**做法：先用 `get_component` 读一眼两边现在各是什么，再只改需要改的那一边。**
拿 `net108` 说：它的 `noEntry` **本来就已经是 `"true"`**（运行时已经只读），所以"要两边都关"
只差 `can_edit` 那一边 —— 照着"改两次"去写，第二次必然是一次 `E_NO_OP`（成功帧，但白花一次）。

**「隐藏」用布局属性 `hidden`，别用 `invisible`。** 这两个现在**问得出来**（§3 那条：
`describe_kind --kind layout` 会告诉你 `hidden` 是 `ENUM false|true`、`invisible` 是 `BOOLEAN`），
但类型相同也可能语义不同 —— `hidden` 是隐藏，而 `invisible` 在 `mod-fd.spec` 里的 4.2 名是
`isPassword`（掩码）：拿它当"藏起来"会得到一个不报错但语义不对的结果。
语料里 `hidden="true"` 是多数的隐藏写法（空串 = 没设）。**它对容器同样成立** —— 隐藏一个
`HBox`/`Grid` 就是在那个容器元素上写 `hidden="true"`（实测：`get_component` 里容器节点带
`hidden`，`set_layout_attr` 写它同样回 `applied:true`）。

**值校验只覆盖布局侧**（`set_layout_attr` / `set_layout_attrs`），依据是 `<工作区>/mta/mod-fd.spec` 的
`<PropertyInfo type=… editorInfo="contains:a|b|c">`。**同一张表现在也读得到前面**：
`describe_kind --kind layout` 每条都带 `type` / `values` / `initial`（§3），所以下面这些边界
是"万一你绕过了那个出口"的兜底，不是唯一的信息来源。三条边界要知道：

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
| `1` | 引擎内部错（`kind=internal`，含 `E_NOT_IMPLEMENTED`）。**`E_NO_OP` 不属于这一格** —— 它是成功码，见下 |
| `2` | 参数/环境不对：本地参数错、未知动词、manifest 拉不到、引擎的 `validation` 与 `not_found` |
| `4` | **设计器拒绝**（`kind=designer`，含 `E_KEY_IN_USE`） |
| `5` | 传输或环境失败（含加载超时、没配工作区、`--args-file` 读不到） |

**`E_NO_OP` 是成功码，不在错误那一侧。** "你请求的值就是当前值"回**成功帧**：
`ok:true` + `result.code="E_NO_OP"` + `noop:true` + `changed:false`，**退出码 0**。
会这样答的写入动词：`set_spec_attr` / `set_spec_attrs` / `set_layout_attr` /
`set_layout_attrs` / `set_tree_source` / `rename_component` / `add_action` / `delete_action` /
`set_action_types` / `set_local_string` / `set_spec_description` / `set_code_template`。
（2026-09-25 起全部一致：在那之前后三个各有自己的形状，`E_NO_OP` 甚至会以**错误帧**回来、
落进上表的 `1` —— 所以看到 `ok:false` + `code:E_NO_OP` 只可能是引擎比你的 SKILL 旧。）
含义只有一个：**请求的值与当前值相同、什么都没改** —— 不是失败，也不用重试。
同理 `E_ATTR_CLAMPED` 也是成功帧：应用了，但值被栅格门禁改过，`result.written` 才是真值。

**收到 no-op 不用改流程。** 它说的是"**这一次调用**没动模型"，不是"你可以跳过剩下的步骤"：
接着改别的属性、照样 `save`（`save` 写的是**当前模型**，与它前面有没有 no-op 无关）。
唯一要小心的是别把"no-op"当成"我这次改成功了"的凭据 —— 值本来就在模型里，所以后面回读
也确实会是你想要的结果；真正证明"改对了"的仍然是回读，不是这一次的返回码。

**报错默认就会把"怎么改对"打出来**（不用加 `--json`）：`合法值`（`detail.legal`）、
`提示`（近似值）、`候选（key 可直接拿去重试）`、以及复数写入的 `已写入`/`失败` 两栏，
都跟在错误那一行下面。看到这些就别自己猜 —— 它们就是从**活着的模型**里取出来的那个集合。
`--json` 拿到的仍是整帧。

**`--json` 下 stdout 恰好一帧**（形状永远是 `{id,ok,result,error,ms}`）。传输失败、
manifest 拉不到、守护进程起不来也各**合成**一帧，形状一样，所以你不必分辨是谁发的。
例外是**本地失败** —— 它们只写 stderr、stdout 为空：用法屏（没有动词 / 动词名打错）、
参数形状错（`--args` 与 `--args-file` 同时给、多了位置参数）、
环境没就绪（引擎 exe 找不到、工作区没配、`--args-file` 读不到）。

⚠️ **"未知参数"不在这一列**（如给 `open` 传 `force`）：它要**先拉 manifest 才知道名字不对**，
所以走的是"给一帧"那条路 —— 实测 stdout 上有一整帧（`code:E_BAD_REQUEST`、退 2），stderr 空。
判据仍然成立："有帧"不等于"发出去过"（那是 §7 表里"参数打错的报错也要引擎在"的同一件事）。

⚠️ **`<动词> --help` 不在这一列**：它是**成功**（退 0），参数表打在 **stdout** 上
（实测 `open --help`：退 0 / stdout 449 字节 / stderr 空）。它和上面那几种"本地失败"不像，
但那句判据对它仍然成立：`--help` 没向引擎发过请求，所以 stdout 有东西 ≠ 发出去过。

由此得到一条**只往一个方向成立**的判据：

> **stdout 为空 = 这次调用没有发出去过** —— 改参数重来是安全的。
> 反过来不成立（manifest 拉不到也会给一帧），所以"有帧"不等于"发出去过"。
> **看到 `code=E_SERVER_DIED` 的那一帧绝不要重试**：请求可能已经上线，只是应答没回来。

## 9. 只读解压：`export`

```powershell
tt dev tzs export "D:/ws/aapt300(c).tzs"    # 纯解压到 <包目录>\aapt300-unzip（只读参考）
tt dev tzs export "D:/ws/aapt300(c).tzs" -o D:/out            # 换落点
tt dev tzs export "D:/ws/aapt300(c).tzs" -o D:/out --force    # 目标非空时覆盖同名文件
```

`export` **不依赖引擎也不依赖设计器**；`-o <dir>` 换落点，目标非空时拒绝（`--force` 覆盖同名文件）。
产物就是**一包文件**（不是 `tzc` 那种带围栏的工作区、不能 apply）；要改表单走动词，不要手工改完塞回去。

> ⚠️ **新包通常比源包小，那不是丢数据。** `save` / `field_add --out` 的返回里 `bytesIn` 是源包、
> `bytesOut` 是写出来的新包，两者常差一截（实测 57,818 → 45,480；13,138 → 11,167）—— 设计器
> 按模型**重算**了各条目并去冗余。这也是"别手改 `export` 的产物再塞回包"的另一个理由：
> 包不是原样拷贝的容器。**要证明落盘就比 `sha256`**（返回里那个与 `sha256sum <out>`），
> 不必靠字节数推断。

## 10. 动词全表（55 个）

`tt dev tzs --help` 会列出它们（`[工作流]` 在最前；会改模型的标 `[写]`，慢的标 `[slow]`）：

| 组 | 动词 |
|---|---|
| 工作流 | `field_add` |
| 会话 | `open` `save` `close` `reload` `verify` `list_open` `list_ops` |
| 读 | `form_tree` `find_component` `get_component` `list_spec_nodes` `describe_kind` `list_packages` `list_tables` `list_columns` `list_records` `list_local_strings` |
| 属性 | `set_spec_attr` `set_spec_attrs` `set_layout_attr` `set_layout_attrs` `set_tree_source` `rename_component` |
| 结构 | `add_widget` `add_field` `insert_at` `delete` `move` `reparent` `nudge` `align` `fit_size` `wrap` `break_layout` `convert_widget` `convert_container` |
| 页签 | `add_page` `delete_page` |
| 语义/Action | `insert_semantic` `add_action` `delete_action` `set_action_types` |
| 多语言/选项/串查 | `set_local_string` `set_items` `set_progrel_programs` `set_table_association` `set_spec_description` `set_cited` |
| Tab 顺序 | `set_tab_order` `tab_action` |
| 校验/工具 | `validate` `base_data` `set_excluded` `set_code_template` |

> `open` **没有 `force` 参数**（2026-09-25 删掉的，原因见 §6）；`export` 的 `-o` / `--force`
> 见 §9 —— 它是内建命令，`--help` 走不通，只能照文档写。
>
> **写动词（标 `[写]` 的 36 个）与 `save` 另有两个横切参数**：`dry_run` 与 `op`，见 §6。
> 其余动词没有它们 —— 读动词挂 `dry_run` 是空话，所以没有。
> `list_ops` 自己也有一个 `op`，但那是**过滤器**（"我要问哪个名字"），同词不同位。

**这 55 个来自引擎的函数表。另有 4 个是 `tt` 自己的内建命令，不在表里、也不接受 `--help`：**
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
- ❌ `list_columns` / `list_spec_nodes` / `list_local_strings` / `list_packages` 不给过滤直接打
  → 可能回几十 KB（先看 `--help` 的"先过滤"提示；`list_spec_nodes` 不过滤实测 29,918 字节，
  `list_packages` 实测 67 个包 11.8 KB）。
- ❌ `attrs` 写成数组、写空对象、或值不带引号（`{"gridWidth":20}`）→ 退 2，点名是哪个键；值一律是字符串。

**改属性**
- ❌ 一次只改一个属性、还一条条发 → 用复数形式（`set_spec_attrs` / `set_layout_attrs`），省掉 N 遍 name-path。
- ❌ **猜布局属性的值** → 先 `describe_kind --kind layout` 看 `values`（§3）；真写错了引擎会拒
  （退 2，`detail.legal` 给值集、`detail.hint` 给近似值）。实测过的坑：`case="UPPER"`（合法的是
  `upper`）、`scroll="MAYBE"`（BOOLEAN 只收 true/false）—— 这两个现在问一次就知道，不必撞。
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
- ❌ 不看 `describe_kind` 就写**不眼熟**的 `set_spec_attr` → 撞 `E_ATTR_NOT_WHITELIST`；白名单是每个包现算的。
  （`req` / `can_edit` / `can_query` 这三个勾选位**不必先问** —— 它们的值与含义在 §7 说死了，直接写。）

**红线**
- ❌ 把 `save`/`field_add` 的 `out` 指向**源包** → **引擎当场拒**（`E_BAD_PARAM`，`detail.reason=out-is-open-package`，
  2026-09-24 起的闸门）。别拿它当兜底：`out` 永远写**新**包。
- ❌ 把 `out` 写到工作区**外** → 现在也是**当场退 2**（同上，`out-outside-workspace`）；
  从前它退 0 成功、之后那个包才 `open` 不了（§4.4）。
- ❌ 把 `export` 的产物当工作区去 `apply` / 手工改完塞回包 → 它只是只读参考。
- ❌ 在没有工作区的地方跑运行类动词 → 引擎缺省工作区是真实客户目录，拒绝启动是保护你的（别绕过）。
- ❌ 请求超时后重发 → 不重试；重新 `open` 或直接看结果。
- ❌ 看到 stderr 的 `… 正在执行 <动词>（<程序名>）；还在等引擎，慢是正常的` 就以为出错 → 那是提示行
  （不是错误：数据仍只在 stdout，退出码也没变）。

## 12. 背景与细节

- 引擎为什么单独构建（不在 Go 构建链里、设计器目录是构建期与运行期都要的依赖、重编会让在跑的
  守护进程变孤儿）：[engine/BUILD.md](../../engine/BUILD.md)
- `.tzs` 格式与引擎契约：[engine/SPEC.md](../../engine/SPEC.md)
- 命令面、验收清单、设计依据：[docs/WIKI.md](../../docs/WIKI.md#7-tt-dev-tzs-表单包)
