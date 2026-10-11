# internal/fep — FEP 协议内核

T100 前端协议（GDC 与 fglrun 之间）的**纯离线**解析层：分帧、S 表达式、AUI 对象模型、
字段名解析、语义动作抽取。它是 `tt fep` / `tt view` 两条线共用的底座。

## 做什么

| 文件 | 契约 |
|---|---|
| `frame.go` | 线格式 `[BE32 长度][BE32 长度][0x01] payload '\n'` 的编解码；流式 `Decoder` 容忍**断包、粘包、字节间夹杂的二进制噪声**（靠逐字节重同步，不信任固定偏移） |
| `sexpr.go` | 消息体的对象树解析 → `[]Node`（`Type/ID/Parent/Attrs`）。支持 `{an <parent> <Type> <id>}`、裸 `<Type> <id>`、`{un}`/`{rn}` 更新、以及**子对象各自带一层花括号**的组形式 |
| `aui.go` | `Model`：把 `[]Node` 合并成对象表，维护父子链、窗口归属、`name→对象` 与窗口索引 |
| `fields.go` | `Resolver`：字段/列 → 中文名（四条规则 + 置信度 + 覆盖率统计） |
| `semantic.go` | 事件消息拆分（`SplitEvents`）与**语义动作**抽取（`SemanticActions`） |

## 不做什么（边界）

- **不连网络、不碰窗口**：一切输入都是录制下来的字节流，所以可离线单测。
- 不发事件、不做校验（发前校验在 `internal/fep/event.go`，属 M2）。
- 不解析业务语义（不做 SQL、不查字典库；`cmt_* → 中文` 只用界面自己推来的 Label 文本）。

## 不变量（改这一层必须守住）

1. **`un`/`rn` 只合并属性**：永不改 `Type`、永不改 `Parent`。
2. **父链只写一次**：先看到非 0 父级就固定；之后出现不同父级 → 记 `Conflicts` 并**拒绝**。
   实测遇到过“残缺消息把对象重挂到别的窗口/造出环”，这条是防它。
3. **子组要递归到底**：一个对象后面可以跟多个子组，且子对象各自带花括号；
   漏一层就会把树压平、把后续子对象挂到外层（原型就是这个 bug）。
4. **字段名只有“高置信”可以对外引用**：链无环且能回溯到作业 `Window`；否则降级为
   `low`，只作诊断。解析不出就是 `unresolved`，**不许猜**。
5. **继承不外扩**：R4 只沿字段类父级（`FormField/TableColumn/Edit/ButtonEdit/ComboBox/
   CheckBox/DateEdit/TextEdit`）向上；遇到容器（`VBox/HBox/Page/Folder/Form/Window/…`）立即停。

## 判据

```
go test ./internal/fep/...          # 全绿（14 个用例）
go vet ./internal/fep/...           # 干净
```

夹具是**手写合成**数据（`aui_test.go` 的 `node()` 构建器按真实格式生成），不含任何客户数据。
真实录制件不入库。

## 已知边界

- 动态开窗窗口里的 `Value/Item` 对象在某些会话里的真实父链仍可能不稳（实测 `3009/3105`
  解析不出）：那类 id 的中文名一律标 `unresolved`，由调用方决定是否人工确认。
- 覆盖率是**实测值**而非目标：本机一条 `axmt500_wf` 会话数据字段 799 → 754（94.4%）。

## M1：代理、会话与 `session open`（2026-10-11）

| 文件 | 内容 |
|---|---|
| `proxy.go` | 本地反向代理：运行期连进来 → 转发到本机 GDC。多连接、逐连接落盘（原始字节）、按连接寻址的注入口（`list` / `@<id> <事件>`） |
| `session.go` | 跨进程登记 `.tt-fep.json`（数据目录下，原子写）；`Stop` **只杀登记的 pid**；`CheckElevation` 权限体检 |

`internal/winproc` 增 `TokenElevated(pid)`（Windows；其余平台返回错误）。

### 三条硬知识（都是真机/真录制件打出来的）

1. **注入的 `_om` 序号**：`seq` 字段存"运行期期待的下一个号"，注入时**先用后加**。
   先加后用会跳号；号在**前端→运行期**方向上跟踪（前端给自己编号），不是在运行期方向上。
2. **消息前缀**：线上消息带动词（`om <n> {{…}}`、`meta Connection {{…}}`），解析要跳过它。
   但**不能"见到 `{` 就切"** —— 裸对象头（`Table 501 {…}`）会被切掉类型与编号，
   整屏塌成一批没有名字、没有父链的对象。只认已知动词（`om`/`meta`/`un`/`rn`/`event`）。
3. **两种录制件都要认**：
   - Go 代理落的是**原始帧**（带长度头）；
   - 早期原型的落盘是**文本**，而且行首还留着帧头噪声（`…\x00\x00+om 0 {{…}}`）。
   `TrimNoise` 去掉噪声并只认已知动词；`--from` 传单个文件时，自动配同一连接的另一侧
   （`-gdc2rt.log` ↔ `-rt2gdc.log`），否则对象编号对不上，动作全变"名字未解析"。

### 实测（离线，真实录制件）

```
tt fep ui snapshot --from axmt-rt2gdc.log   → 对象 4558 / 窗口 w_axmt500 / 字段名 447-492（90.9%）
tt fep ui find --from axmt-rt2gdc.log --text 客户  → 客户编号 high name-suffix（comment=cmt_xmda004）
tt fep rec steps --from teach-4-gdc2rt.log  → 原始事件 52 → 语义动作 17（点击 录入 / 输入 200 / 点击 确定）
```

退出码：0/1/2/3/4/5/6（表在 `internal/cli/README.md`）。未知子命令退 2、无会话退 3，
都实测过。

### `session open`（常驻会话）

| 文件 | 内容 |
|---|---|
| `open.go` | 起隔离实例 + 代理 + SSH 登录开作业；`Wait`/`Close`；`ErrUsage`/`ErrPreflight`/`ErrPortBusy` 三类失败供命令面映射退出码 |
| `winproc/list_windows.go` | `ListProcesses(name)`：按可执行文件名列进程（权限体检要用） |

顺序（不能换）：**权限体检 → 端口体检 → 起隔离实例 → 起代理 → 写登记 → 开作业**。
登记写在开作业之前，所以中途被打断也收得回来；`Close` 的顺序是
远端作业 → 代理 → 隔离实例（**只杀登记的 pid**）→ 清登记。

- `FGLSERVER=<本机 IP>:<代理端口-6400>`：作业客户端连的是 `FGLSERVER 数值 + 6400`，
  所以代理监听的那个端口减 6400 才是要下发的值（实测 offset 固定）。
- 本机 IP 用**到 SSH 主机的路由源地址**探测（UDP 连接不发包，只问内核选哪条路由）。
- 等隔离实例就绪用**轮询端口**（不是固定 sleep），就绪后再加 0.7s 余量：
  渲染中途取帧会得到一张看起来很像真的半成品图。
- 登录序列与调试线一致：等区域菜单 → 写 zone → 等 shell 提示符。

### 事件构造与发前校验（`event.go`，M2 第一块）

构造器逐字节对着**录制件里客户端自己发出的**事件写：

```
ActionEvent 0{{idRef "5160"}}
KeyEvent 0{{keyName "Tab"}}
ConfigureEvent 0{{idRef "60"}{ value "LUYAO"}}
ConfigureEvent 0{{idRef "6543"}{ cursor "5"}{ cursor2 "5"}}
```

两个必须照抄的细节：第一个属性紧贴 `{{`、后续属性前**有一个空格**；
**值输入不带光标对**（早期探针的"必须带 cursor/cursor2"与录制件不符 ——
真人打字产生的是 value-only，光标对是客户端自己另发的一条）。

发前校验（只判断、**不发包**）：目标存在 → 类型白名单（含 `controlp` 的 ButtonEdit）→
`actionActive` → `_om` 序号等于该连接的下一个 → `KeyEvent` 禁带 idRef。
打字目标优先取字段的 `FormField` 子对象（`ResolveInputTarget` 会把"有没有拿它"如实报告出来，
不假装验证过）。

### 动作面（`do.go` + `record.go`，M2 后半）

跨进程：常驻会话在另一个进程里，本层经**注入口**（回环 TCP）与它说话，AUI 状态从
**录制件**重建（运行期不会回答"当前界面是什么"，它只会推 `om`）。

- 注入口新增 `status`：回一段 JSON（每条连接的 id / 下一个 `_om` / 存活 / 字节数），
  别的进程因此不必持有任何内存状态，重开一个终端照样能干活。
- `Do(click/key/input)`：先校验后发送，`--dry-run` 只校验与打印（字节增长 = 0）。
- 目标可以给名字或 `#id`。一个名字在一屏里**往往命中多个对象**
  （Label + ButtonEdit + FormField 都叫「客户编号」）——这是真实情况，不是异常：
  按动作类型收敛（点击只认可点击的、输入只认可写的），输入再优先取 FormField；
  仍不唯一就报歧义让人用 `#id`，绝不自己挑一个。
- 每步写审计（动作/连接/目标/事件/是否真发/结果），成没成都记。

### 两个 Windows 上的坑（都实测到过）

1. **`DirEntry.Info()` 的大小可能是陈旧的**：目录枚举那一刻的快照，实测
   `entryInfo=0` 而 `os.Stat` 同一文件是 `356`。判断"写进去了没有"必须用 `os.Stat`。
2. 抓屏进程必须在任何 GDI 调用**之前**声明 DPI 感知（见计划 §3 顶部）。

### 实测（离线，全绿）

`go test ./internal/fep/` 36 用例：其中 `do` 的端到端用**真代理 + 假前端 + 假运行期 +
临时登记目录**跑通 —— 点击按名字命中 `录入` → 发到前端的事件字节是
`event _om 1{}{{ActionEvent 0{{idRef "5160"}}}}`；输入命中字段的 `FormField`；
`--dry-run` 下一个字节都不发；`actionActive` 为假与名字不存在都被拦且不发包。

## 真机联调（2026-10-11）：两个让它"看起来在动、其实没动"的根因

第一次真跑 `session open` + `do act` 时，事件发出去了、`seq` 也推进了，**作业一动不动**。
逐字节对比真人点击与注入的帧之后，抓到两个都属"静默失败"的真根因：

### 1. 帧的长度字段**包含结尾换行**

| | 声明长度 | 文本长度 | 末尾字节 |
|---|---|---|---|
| GDC 自己的帧 | **47** | 46 | `0x0a` |
| 修前我发的帧 | **44** | 44 | `0x7d`(`}`) |

少算那 1 个字节，运行期按声明长度切完就再也对不上下一条消息 —— 连接是好的、字节也
发出去了、**作业一动不动**；而且之后连真人的事件也一起被错位吃掉（这解释了当时"注入
一次之后整轮就哑了"）。解码侧相应地把末尾换行从载荷里切掉（载荷 = 消息文本）。

### 2. 原始录制件的第一条握手**不带帧头**

真实原始录制件的开头是 `meta Connection {{…}}`，它是整个会话里唯一不带帧头的一条。
拿"首字节是不是字母"判形态会被它骗过去，把二进制帧流当文本读 —— 结果是**安静的错答案**：
整屏解析成垃圾、字段值一个都不见（AUI 覆盖率看着还行，值全丢）。现在先按帧试、再退回文本。

### 联调判据（都实测过）

```
session open        → 代理连上：#2 alive seq=3 bytesRT=384002 remote=172.16.1.109:37506
do act 录入          → 已发 ActionEvent 0{{idRef "5160"}}；bytesRT 384002 → 446673（运行期推回整屏）
do type 客户编号 <值> → 优先发到 FormField#271；运行期回 om 7 {{un 271 {{value "…"}}}}（值被接受）
ui find --text <值>  → #271 FormField 客户编号 value="…"（输入生效可复核）
do act 流程状态      → 被第 3 条拦下：actionActive 为假（不发包）
真人点击同一条连接   → 同样生效（协议侧可见）
session close       → GDC 已结束、登记已清、进程 0、端口仅剩 TIME_WAIT
```

## M2 完成（2026-10-11）：等回包、按窗口路由、变更汇总、正式区警告

- **`Do` 会等运行期**：发完盯 `bytesRT`（运行期推来的字节），等到了再读模型并汇总变化 ——
  「事件发出去了」和「作业真的变了」是两件事，只有后者有用。`--wait` 默认 10s，负数=不等。
- **变更汇总**（`Changes`）：运行期新增字节 / 新消息条数 / 新窗口 / 值变化（带字段名）/
  是否收到应答；超时如实报 `timedOut`，不假装成功。
- **按窗口路由连接**：`--window <名字>` 会先查这个窗口在**哪条连接**上 ——
  依据是录制件文件名（`<tag>-<ID>-rt2gdc.log` 就是那条连接收到的 AUI），
  多连接同名窗口时报错让人用 `--conn` 指定，绝不猜着发（发错连接的后果实测过：窗口不动）。
- **`tt fep wait --window/--value/--text`**：等作业真的变成某个样子；轮询协议录制件
  而不是屏幕（协议是唯一权威，屏幕是它的渲染结果）。
- **正式区写类动作警告行**：依据作业自报的区域文本（`…(正式区):t35prd,99`）与
  写类判断（名单外的动作都算写；名单来自真机工具栏实际存在的项），只警告不硬禁（决策②）。
- 开关解析错也退 2（cobra 默认给 1，与这条线的表不符）。

离线判据：40+ 用例全绿，其中新增两个 —— 等待回包（假运行期回一条 `om … {{un 271 …}}`
→ `Changes.Values` 里能看到值变化）、等条件成立（`--value 客户编号=LATER`）。

## M3：回放（`replay.go`）

与 `rec steps`（给人看的步骤清单）**不是一回事**：steps 丢导航、只要语义；replay 求保真。
白名单的边界是实测出来的：

| 回放 | 丢 |
|---|---|
| `ActionEvent` / `KeyEvent` | `tabIndex` / `bufferSize` / `pageSize` |
| `ConfigureEvent{value …}` | 纯光标事件（`cursor`/`cursor2`） |
| **`ConfigureEvent{currentRow/currentColumn …}`** | |

最后一行是当年踩过的坑：把它当噪声滤掉，回放到开窗那一步网格全空，作业还弹
“传入的料件编号为空!” —— 那对属性就是“选中哪一行”的语义。

三道保险：白名单之外不发；**AUI 安全闸**（每一步都重新读界面，对象不在当次界面里就跳过
—— 用旧模型会误杀后面的步骤）；`--dry-run` 只出计划。`--stop-at N` 让人先跑一半看效果，
`--skip <子串>` 绕开某几步。每一步写审计。

计划阶段全是离线的（读录制件），所以 `--dry-run` **不需要有会话** —— 先看清要发什么，
再决定要不要开会话，这个顺序本来就该如此。实测：真机那轮教学录制件（teach-4）出 50 条计划。

### 真机第 3 条硬知识：输入**只能**发给 `FormField`（`-6327` 的根因）

把 `ConfigureEvent{value}` 发给字段**容器本身**（实测发给了 `Edit#263`），作业进程当场停下：

```
FORMS statement error number -6327
Internal error in the run time library file ../ui/indialog.c(10523):0
Program stopped at 'axm_axmt500_wf.4gl', line number 19839
```

代码里原来留了一个"找不到 FormField 就用对象自己"的兜底 —— **那个兜底就是凶手**，
现在删掉了：找不到 FormField 就**拒发**并说明去查 `FormField`。发给 `FormField`（如
`#271`、`#259`）则一路正常，运行期回 `om {{un 271 {{value …}}}}`。

一句话：**宁可不做，也不做错** —— 这条错的代价是打死作业进程。

## 【已解】开窗怎么点：先纠正一个错判据（本节的"未解"结论是错的）

早期结论是"合成点开窗都不开窗，判据是连接数不增加" —— **判据本身就错了**：
开窗**不开新连接**，它是在同一条连接里推一个新的 `Window` 对象（`an <conn> Window <id> {name "w_qry_<开窗程序>"}`），
并伴随一次 `currentWindow` 切换。按"连接数"看，永远看不到开窗。
正确的是：

- 判据 = **AUI 多一个 `Window` 对象**（`tt fep ui tree --window w_qry_` 一看就知道），加上进程窗口列表里多一个窗口；
- 触发 = 单击字段的**开窗按钮**（`ButtonEdit` 且 `action=controlp`，如供应商 `#273`；
  明细列的开窗按钮要求**光标先落在那一行的那一列**，否则 `actionActive=0` 会被拒）；
- 之后的取数序列见下一节与 `openwin.go`（`tt fep src openwin <程序>`）。

下表保留当时的事实（都是真的），错的是从它们推出的结论：

| 试法 | 当时看到的结果 |
|---|---|
| 单击 `ButtonEdit{controlp}`（客户编号 #272） | 只回 ACK（+79 字节）—— 其实**开了**，只是判据看连接数 |
| 先发焦点 `ConfigureEvent{tabIndex}` 再单击 | 同上 |
| 直接单击 `FormField`（#271） | 同上 |
| 双击（两次 `ActionEvent` 相隔 150ms） | 只回 ACK（+36 字节），连接数仍为 1 |
| 裸 `KeyEvent F4` | 只回 ACK（+17 字节） |

## 开窗：列表是前端本地渲染的（真人录制对照出来的）

一轮真人示范（`fep-20261011-123139` 的录制件）把整条序列录下来了：

```
_om  9: ActionEvent 272   ← 真人点「客户编号」的开窗按钮（我上次也发过这条）
_om 11: ActionEvent 6638  ← 接着点列表里的项
_om 12: ActionEvent 6590  ← 再点一项（选定）
_om 20: ConfigureEvent{idRef 48}{value "LUYAOWF"}     ← 在料号格子输入
_om 3x: ConfigureEvent{idRef 3009}{value "1000"}      ← 在订购量格子输入
```

**结论修正**：点开窗那一下**协议上完全没有回声** —— 列表由前端本地渲染（连接数是 1，
没有新连接、没有新事件）。我上次发完 `272` 看不到任何反应，就误判成"开窗做不到" ✗。
正确的序列是：**点开窗（协议静默）→ 再点列表里的项（这一步才有协议事件、也是真正的"选择"）**。
所以 `do act <栏位>` 紧接着 `do act <列表项>` 是可行且符合真人节奏的；
但"开出来了没有"**不能**用协议判断，只能看屏幕（`tt view shot`）。

## 不变量回归：真语料折叠（`corpus_test.go`）

AUI 同步只有三种操作（官方 1522）：`an` 加、`un` 改、`rn` **删**。
2026-10-11 现场查出来的根因就是 **`rn` 被当成 `un`**：被删的 Dialog / 工具栏项 /
模板条目永远不死 —— 症状全是间接的："同一个 id 既是 Item 又是 TableColumn"、
"`accept` 有三个候选"、focus 乱跳、事件发到已关闭的旧实例上（只回 18 字节空 ACK）。

这条 bug 靠肉眼在几百行流里找了一整天才找到，所以它必须变成每次 `go test` 都会
自己跑回来的检查。`corpus_test.go` 折叠真语料并断言：

1. **被 `rn` 的对象不许还活着**：该裸 id 的**任何一个候选**都不许剩（判据是运行期说
   "这个 id 删了"，而我的作用域归属只是我自己的结论，算错时不许让对象活下来）；
2. **被 `rn` 的 Dialog 不许还在"当前交互指令候选"里**（`ActiveDialogs`）；
3. **每个裸 id 必须能判出"现在是谁"**：运行期**会复用裸 id**（同一个 `159` 既做过
   全局 `ActionDefault`、又做过网格格子；同一个 `4937` 既做过下拉 `Item`、又做过
   `Action`），所以"同一作用域一个 id 两种类型"本身不是错误 —— 但候选之间必须能
   按 `Object.Seq`（最后声明/触碰）**严格分先后**，`Get(id)` 必须给出最新的那份；
4. **窗口登记的网格必须存在**（`winTables` 的账本要跟着 `rn` 更新）；
5. **反证**（`TestTheRnInvariantHasTeeth`）：把 `rn` 退化回 `un`，上面第 1/2 条**必须**红。
   没有这条，"不变量测试"就只是个摆设。

语料两层：

- **仓库夹具** `testdata/fep-recordings/*.txt`（必跑，剪法见 `tools/fepfix` 与该目录 README）；
- **现场语料** `%APPDATA%\tt\fep\*-rt2gdc.log`（可选，含真实业务数据，`TT_FEP_RECORD_DIR` 可改）。
  实测 94 个连接 / 910 条 AUI 消息，其中 49 个是空壳连接，最大一份 1.19MB / 80 条消息。

## 动作层：两道防线（`errors.go`）

发一个事件只有两种结果值得区分：**真的生效了**、或**作业已经停了**。
"发了但没反应"是最坏的一种 —— 它看起来像"形态不对"，会让人接着改参数接着试（今天就这样白试了十几步）。
所以这一层有两条必须过的闸：

### 发送前：动作必须属于**当前交互指令**

官方 1514：动作由 Dialog 节点持有，交互元素按 name 绑定到 Action 节点。
实测更狠 —— 运行期**每次状态变化都重新分配动作 id**（同一个"确定"：模板 `#109` →
实例 `#3535` → 查询后 `#4533` → `#4538`），所以"名字对了"不够，**这一份**也要对。
`CheckClick` 因此在 `actionActive` 之外再要一条：动作的 Dialog 必须是当前 active 的那个
（判据不足时 —— 模型里一个 active Dialog 都没有 —— **不拦**，因为那时任何判断都是猜）。

### 发送后：认得出"作业停下了"

运行期把程序停掉时会推一条界面消息（2026-10-11 实测原样）：

```
om 22 {{an 4811 Menu 39 {{text "ERROR"} {posY "0"} {style "winmsg"}
        {comment "Program stopped at 'qry_q_pmaa001_3.4gl', line number 278.\n
                  FORMS statement error number -6327.\n
                  Internal error in the run time library file ../ui/ndialog.c(10523):0.\n"}}}
```

`RuntimeErrors` 认它（`Menu{text "ERROR"}` 或 comment 里含 `Program stopped at`），
`waitChanges` 把它写进 `Changes.RuntimeError`，`Do` 于是返回**运行失败**而不是"没生效"，
`tt fep ui state` 也会在最前面打一条 `⚠ 作业程序已经停下了：<程序>:<行> (<错误号>)`。

触发它的实测原因只有一类：**把值/事件发给了容器**（`Edit`、`ValueList`、`ToolBar` 本身），
或发给不属于当前指令的动作。这正是 `writableTypes` / `ClickableTypes` 白名单与
"绝不试探"那条纪律的理由。

夹具：`testdata/fep-recordings/program-stop.txt`（现场 `20261011-170312` 剪出来的，就是那次停摆）。
`TestHealthyFixturesHaveNoProgramStop` 是反向对照：正常会话绝不能被误判成停摆，
否则这条检查会变成噪声、人会开始忽略它。

## W3：从源码读开窗语义（`openwin.go` + `tt fep src openwin`）

**同一个"确定"按钮，两个开窗的"选中"含义完全不同** —— 这是"绝不猜、回源码读"最硬的例子：

| 开窗 | `ON ACTION accept` 段 | 所以"选中"= |
|---|---|---|
| 供应商 `q_pmaa001_3`（`DISPLAY ARRAY` 那个对话框） | `IF li_ac > 0 THEN return1 = gr_qry[li_ac].pmaa001_1` | **当前行** |
| 料件 `q_imaf001_15`（`INPUT ARRAY` 那个对话框） | `CALL …_get_multiret()` → `IF …[i].check = "Y"` | **勾选** |

后者靠试是试不出来的：不勾选直接点确定，运行期**只回一个空 ACK**、界面什么都不返回 ——
看起来像"形态不对"，真实原因是"你没勾选"。我为这个判断浪费了一整天。

而且**一个开窗可以有两个对话框**（实测 `q_pmaa001_3` 与 `q_imaf001_15` 都有一个
`INPUT ARRAY … FROM s_qry.*` 与一个 `DISPLAY ARRAY … TO s_qry.*`），语义各不同，所以结论要**按对话框**给。

### 运行期怎么告诉你在哪个对话框

`dialogType`（现场语料 94 个连接的取值分布）：`DisplayArray` 21817 / `InputArray` 11328 /
`Input` 1697 / `Construct` 965。映射关系（两侧都是实测）：

| `dialogType` | 源码里对应 | "选中"语义 |
|---|---|---|
| `Construct` | `CONSTRUCT … ON <列,列,列>` | 没有"选中"：往条件列写值 → 确定执行查询 |
| `InputArray` | `INPUT ARRAY … FROM s_qry.*` | 按该对话框的 accept 读出来的（实测是勾选） |
| `DisplayArray` | `DISPLAY ARRAY … TO s_qry.*` | 按该对话框的 accept 读出来的（实测是当前行） |

`SelectModeFor(ow, dialogType)` 就是这张表；**对不上时回 `unknown`**（宁可不判，不猜）。

### 勾选要写什么值

从 `.4fd` 读：`<CheckBox name="check" valueChecked="Y" valueUnchecked="N" widget="CheckBox"/>`
（`q_imaf001_15.4fd:59` 原样）。没有它就只能猜"勾是用 Y 还是 1 还是 true"——猜错就是空 ACK。

### 命令

```bash
tt fep src openwin q_imaf001_15 --source D:/T100/源代码/主机109
```

输出：查询条件列 / 每个对话框的 accept 调用链 / 选中语义 / 回传字段 / 勾选栏字面值，
**每条结论都带出处**（`文件:行 + 原文`）。镜像根取自配置 `mirror.dir`（`tt dict mirror dir`），
可用 `--source` 或 `TT_SOURCE_ROOT` 覆盖。

回归：`openwin_test.go` 用的是两个开窗的**真实结构**（删掉注释与无关行），
断言"InputArray → 勾选 / DisplayArray → 当前行 / Construct → 没有选中 / 未知类型 → unknown"。

## 地基：对象属于哪个窗口（`WindowOf` / `TablesOfWindow`）

驱动每一句都要先按窗口收窄 —— 因为**列与格子的 id 跨实例复用**，同名同 id 的对象在模型里
同时存在多份。今天在这里踩了两次：

| 症状 | 原因 |
|---|---|
| 主窗的明细网格 `#644` 在 `--scope 178` 里看不到 | 旧写法只登记"与窗口**同一条消息**里声明"的网格 |
| 料件窗口的 `--scope` 里冒出供应商的列 | 旧写法拿 `Scope`/`Win` 兜底，而 `Win` 被**无条件**改写 |

判据是从一份真会话里把三套证据逐条对过才定的：

| 证据 | 含义 | #644（主窗明细） | #586（两窗共享的模板列） |
|---|---|---|---|
| `DeclWindow` | **最近一次声明它的那条消息**属于哪个窗口 | 178 ✅ | 687 ✅ |
| 父链 → 最近的 `Window` 祖先 | 结构真相 | 178 ✅ | 155 ❌（指到别的实例） |
| `Win` | 运行期最后一次把它当"某窗口的东西"触及 | 155 ❌ | 687 ✅ |

所以 `WindowOf = DeclWindow（若那个 id 现在确实是 Window）→ 父链 → Win → 0`，
并且 `Win` 的**改写规则也收窄了**：只在 (a) 本条消息自己声明了当前窗口（新实例构造，
模板对象被重新触及），或 (b) 它本来就属于那个窗口时才改写 —— 否则一条"开窗时顺手更新主窗字段"
的消息会把不相干的对象记到新窗口名下。

回归（`window_test.go`）：**两个活着的窗口不能共有同一个对象 id**（这条不靠具体角色名，
任何语料上都成立）、网格归属不许指向幽灵窗口（"还活着"必须是"那个 id 现在还是 Window"，
裸 id 会被复用）、开窗窗口必须能问出它的 `s_qry` 结果网格。

## 地基：源码 ↔ 运行期交叉校验（`ReconcileSelect`）

`OpenWinProgramFromWindow("w_qry_q_imaf001_15") == "q_imaf001_15"` —— 驱动**不需要人告诉它**
刚开的是哪个开窗，运行期在窗口名里写着了。

`ReconcileSelect(ow, dialogType)` 把两处权威对起来，规则是**宁可不判，不猜**：

- 运行期说 `InputArray`/`DisplayArray`，源码里有对应对话框且读出了语义 → 用源码的（带 `文件:行`）；
- 源码里没有那个对话框、或读不出语义、或**根本没有源码** → `unknown`；
- `Construct` 阶段没有"选中"这回事。

`StageDefaultSelect`（`InputArray`→勾选 / `DisplayArray`→当前行）是**观察，不是规矩** ——
它只在调用方显式 `--assume` 时才用，输出里标「推断·非源码」。`tt fep do lookup` 默认：
判不出来就**停手**，并告诉你该给 `--source`。
