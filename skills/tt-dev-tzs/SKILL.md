---
name: tt-dev-tzs
description: 读写 T100 设计器**表单包**（.tzs/.tzv）：由设计器自己的引擎驱动（49 个函数，命名管道 JSON-RPC），不是拼 XML。流程 open → 读/改 → validate → save --out 写新包。当用户要改表单（改属性、加字段、加页签、调布局、换容器、多语言/选项/串查、Tab 顺序）、从数据字典按表加字段、或查看表单结构与字段时使用。**代码包（.tzc/.tzf/.tzx）不归这里，用 tt-dev-tzc。**
license: 与 tt 仓库一致（见随包 README.md）
metadata:
  tool: tdev
  command: tt dev tzs / tt install
---

# tt dev tzs：T100 设计器表单包的读写

**这是表单包（`.tzs`/`.tzv`）的技能。** 代码包（`.tzc`/`.tzf`/`.tzx`）是**另一条互不相通的管线**，
用 `tt-dev-tzc`：

| | `tt dev tzs`（表单包，本技能） | `tt dev tzc`（代码包，见 tt-dev-tzc） |
|---|---|---|
| 干什么 | 由设计器自己的引擎读写**表单模型** | 把包渲染成带围栏的 `.4gl` 工作区给人改 |
| 唯一写路径 | `call save --out`（写**新**包，不动原包） | `apply`（闸门校验 + 原子写回原包） |

拿错入口会被挡住：`.tzs` 跑 `tzc export` → 退出码 2 并提示改用 `tzs export`。

表单的读写**由设计器自己的代码算**：`engine/` 里那个 C# 引擎反射驱动**已安装的设计器自己的程序集**
（布局属性走设计器自己的 `XmlElement` 索引器，`.tsd` 由设计器从模型重算）。我们一个字节的 `.tzs`
格式都没实现 —— 所以**手工拼 XML 这条路是不存在的**，改表单只有 `call` 一条路。

## 一次改动的完整形状

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

**不要跳过读那几步**：`describe_kind` 给出的白名单是**运行时**从这个包算出来的
（同一类节点在不同包里属性集不同），照着它写 `set_spec_attr` 才不会撞 `E_ATTR_NOT_WHITELIST`。

纯解压（只读参考、写文档用；**不依赖引擎也不依赖设计器**）：

```powershell
tt dev tzs export "D:\pkg\aapt300(c).tzs"    # → D:\pkg\aapt300-unzip\
```

要配两样（机器级依赖，没有合理缺省）：

- **设计器**：不用配。它的程序集随仓库与发行包自带（`engine/designer/` 与 `tzs\designer\` 是同一组），
  引擎默认从自己旁边的 `designer\` 加载 —— 所以同一份 tt 在任何机器上跑的是**同一版设计器**，
  不存在"你装的那版"和"我装的那版"。`TZSCLI_INSTALL` 只是覆盖手段（拿别版验证时才用）。
- **用哪个工作区**：`tzs.workspace` 或 `TZSCLI_WS`。**这条没有缺省、也不回落** —— 引擎内置的
  默认工作区是一个**真实客户目录**，落上去等于拿别人的表单当草稿纸。三层都空时拒绝启动（退 5）。

## 句柄（handle）：最容易踩的一组

- **每个需要句柄的函数都必须显式给 `--handle`**。没有「自动沿用上次句柄」这回事：
  状态文件（`%APPDATA%\T100\tt\.tt-tzs.json`）里虽然记了一份最近 `open` 的结果，
  但命令层不读它，省掉 `--handle` 只会得到退 2「缺必填参数 handle」。
- **句柄永不复用**：`close` 掉一个包再 `open` 同一个包拿到的是**新号**（不是原来那个）。
  拿旧号去调 → `E_NOT_FOUND`「句柄不存在或已关闭」（退 2）。这是安全失败，不是 bug。
- **同一个程序已经开着再 `open` → `E_KEY_IN_USE`（退 4）**，报错会说清是哪个文件占着。
  先 `close` 占用者（`list_open` 看是谁），或 `open --force`（只对 `Loaded` 状态的占用者有效，
  `Mutable` 的会被拒）。
- 句柄只活在常驻守护进程里。**进程一死全部失效** —— 守护进程挂了就重新 `open`，别重放写请求。

## `validate` 是**基线相对**的，首调必然「零增量」

这是读输出时最容易误判的一处：`validate` 在**句柄上第一次被调用时，那次运行本身就是基线**
（`Fns/Validate.cs:139`）。所以第一次调用 `newErrors`/`newWarnings` **按构造就是空** ——
它证明不了任何事。要看「我改坏了没有」，必须在**同一个句柄**上先 `validate` 一次建立基线。

实测：`aapt300(c).tzs`（未做任何改动）首调 `baseline = 13 条 WARNING`，`after = 13`，
`newErrors = 0`、`newWarnings = 0`。**语料里本来就不干净** —— 所以判据永远是
「改动后的增量」，而不是「有没有 WARNING」。

返回体：`{baseline[], after[], newErrors[], newWarnings[], elapsedMs}`。
加 `--json` 会把**整帧**（`{id, ok, result, error, ms}`）原样打出，而不是只打 `result`。

## 参数语法（本地只做**语法**校验，语义交给引擎）

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

## 其它必知

- **`--depth 1` 意思是「只有这个节点本身」**，不是「往下看一层」。要子节点至少 `--depth 2`。
- **`save --out` 写到新包，不动原包**。不要 `--out` 指向源包 —— 那是唯一会毁掉原始素材的写法。
- **请求一旦上线绝不重试**：协议无幂等键，而这些函数都在改设计器内存里的模型，重试是在赌
  「上一次写进去了没有」。守护进程死了就重新 `open`。
- `stop` **绝不 spawn**（让 stop 去起一个服务器是件可笑的事）；`reap` 收的是**引擎重编后**
  停不掉的孤儿守护进程（管道名含 MVID，重编就换名，旧进程的 pid 只能靠状态文件的 `orphans` 找到）。
   `reap` 不加 `--yes` 只列不杀。

## 红线（碰了就出事，没有例外）

1. **`export` 的产物是只读参考**：它就是一包文件，没有 manifest/围栏。不要手工改完再塞回包，
   也不要拿它当 `tzc` 工作区去 `apply`。要改表单用 `tt dev tzs call` —— 那是设计器自己的模型在算，
   改完设计器打得开；手工拼 XML 则不然（`.tsd` 由设计器从模型重算，`.4fd` 的 Record 段要重建）。
2. **请求一旦发出不要重试**（同「其它必知」第二条）。
3. **`save --out` 永远指向新包**，绝不指向源包；`export` 只读源包。实验一律用副本或临时目录。
4. **不要在没有工作区的地方跑** —— 引擎的缺省工作区是真实客户目录，命令会拒绝启动而不是回落；
   不要试图绕过那个拒绝。

## 退出码

**没有 `3`**（`3` 是 `.tzc` 那条线的「验证失败」），多一个 `1`：

| 码 | 含义 |
|---|---|
| `0` | 帧 `ok:true` |
| `1` | 引擎内部错（`kind=internal`，含 `E_NOT_IMPLEMENTED`） |
| `2` | 参数/环境不对：本地参数错、未知函数、manifest 拉不到、引擎的 `validation` 与 `not_found` |
| `4` | **设计器拒绝**（`kind=designer`，含 `E_KEY_IN_USE`） |
| `5` | 传输或环境失败（含加载超时、没配工作区） |

## 动词速查

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
`--container Table` 才得到「一个 Table 装 N 列」；一次一列会让大表单慢两个数量级
（实测 84 列从 137 秒降到 1.17 秒）。

## 常见简单错误（自查表）

- ❌ 调 `call` 时省掉 `--handle` → 退 2「缺必填参数 handle」。**没有自动沿用上次句柄**这件事。
- ❌ 上一个包 `close` 之后仍拿旧句柄调 → 退 2「句柄不存在或已关闭」。**句柄永不复用**，
  重新 `open` 拿新号。
- ❌ 同一个程序已经开着又 `open` → 退 4 `E_KEY_IN_USE`。先 `close` 占用者（`list_open` 看是谁），
  或对 `Loaded` 的占用者用 `open --force`。
- ❌ 看到第一次 `validate` 报 `newErrors: []` 就以为改对了 → 首调**就是建立基线的那次**，
  零增量是构造性的。要比对增量必须在同一句柄上先建立基线。
- ❌ 拿 `baseline` 里的 WARNING 数当「改坏了」 → 语料**本来就不干净**（aapt300 未改动即有 13 条）。
- ❌ `form_tree --depth 1` 想看子节点 → 那是「只有这个节点本身」，至少 `--depth 2`。
- ❌ 不看 `describe_kind` 就写 `set_spec_attr` → 撞 `E_ATTR_NOT_WHITELIST`；白名单是每个包现算的。
- ❌ `--paths a b c` 希望是三个 → 列表要么 `--paths a,b,c`，要么 `--paths a --paths b --paths c`。
- ❌ 写参数时用位置参数（裸词）→ 一律报错，参数都是 `--名 值`。
- ❌ `--offset 2 --offset 3` 这种标量给两次 → 报错（只有 `path[]`/`string[]` 可重复）。
- ❌ 以为参数打错会立刻在本地报 → `call` **先拉 manifest 再校验参数**，引擎 exe 不在就要退 5。
- ❌ `--timeout 30` 想快点失败 → 会被抬到 120 秒（下限是给加载看门狗留的）。
- ❌ `save --out` 指向源包 → 唯一会毁掉原始素材的写法。永远写**新**包。

## 背景与细节

引擎为什么单独构建（不在 Go 构建链里、设计器目录是构建期与运行期都要的依赖、重编会让在跑的
守护进程变孤儿）见 [engine/BUILD.md](../../engine/BUILD.md)。
`.tzs` 格式契约见 [engine/SPEC.md](../../engine/SPEC.md)；
命令面与验收清单见 [docs/WIKI.md](../../docs/WIKI.md#6-tt-dev-tzs-表单包)。
