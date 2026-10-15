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

```bash
go test ./internal/fep/...          # 全绿（含真语料夹具上的不变量回归）
go vet ./internal/fep/...           # 干净
go test ./internal/fep -run TestGate -count=1 -v   # 指纹闸：阈值两侧各一条
```

- **期望**：`TestFixtureRecordingsKeepInvariants` 在三份仓库夹具（`testdata/fep-recordings/`）上逐条过；
  `TestTheRnInvariantHasTeeth` 把 `rn` 退化回 `un` 就得红；`TestGateRefusesTwoDifferentRealInterfaces`
  在两份真夹具上判"拒"且给出差异条数。
- **这一层负责证明**：协议解析与 AUI 合并在真录制件上守住不变量，且"这份录制配不配得上这个界面"
  的判决只在**表对得上**时放行。真机行为（起作业、发事件、断点）不在这里证。
- **故意不覆盖**：现场语料（含真实业务数据，不随仓库走）与真机 E2E —— 它们分别由
  `TT_FEP_RECORD_DIR` 那条可选路径和 `tt debug` 那条线覆盖。

夹具是**手写合成**数据（`aui_test.go` 的 `node()` 构建器按真实格式生成）与现场录制的裁剪件
（`testdata/fep-recordings/`，裁剪方式见该目录 README）。原始录制件不入库。

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

#### 启动时下发的两样东西（`jobLaunchCmd`）

`export TOPENT=…`（给了 `--ent` 时）→ `export FGLSERVER=…` → `r.r <作业>`，**顺序不能换**：
两样都必须在 `r.r` 之前生效，反了作业会带着上一轮的值跑起来。

- **环境必须点名**：`session open` 走 `config.Hosts.Resolve`，不给 `--env` 就报
  「必须指定环境(--env <环境名>);可用环境: …」，与 `tt debug`/`tt dict` 同一句。
  这里**刻意不用** `debug.Config.EnvName()` 兜底 —— 它最后一级是 `SSH.Host + "-" + Zone`，
  环境没选中时 SSH 是零值，会拼出一个叫 `-` 的幽灵环境名，报错成了
  「环境 "-" 不存在」（实测踩过，人看不出该补什么）。
- **企业编号只能每次显式给**：`config.NamedSsh` 刻意没有 `topent` 字段，
  所以配置里就算留着 `topent` 也是死键、没有任何代码读它。作业落在哪个企业只由
  登录 shell 的 `$TOPENT` 决定 —— 不给 `--ent` 就如实回显"沿用登录 shell 自己的"，
  不猜、也不擅自 unset。
- 值不限数字/文本：单引号包裹**并剔除内嵌单引号**，与 `internal/debug` 的
  `topentShellCmd` 同一套规则（两处必须一致，改一处就要改另一处）。

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
所以回放**按事件类型整类放行**（`AllowedForReplay`：`ActionEvent` / `KeyEvent` / `ConfigureEvent`），
不做逐属性过滤 —— 焦点（`ConfigureEvent{cursor…}`）看着像导航噪声，但运行期正靠它确认
"光标现在在这个栏位"，滤掉它后面的动作就全打在空气上。"人做过哪几步"要用 `rec steps` 看。

三道保险：**AUI 安全闸**（每一步都重新读界面，对象不在当次界面里就跳过
—— 用旧模型会误杀后面的步骤）；`--dry-run` 只出计划；`--stop-at N` 让人先跑一半看效果、
`--skip <子串>` 绕开某几步。每一步写审计。在这三道之前还有一道**指纹闸**：它先判
"这份录制配不配得上现在这个界面"，判据见下节。

计划阶段全是离线的（读录制件），所以 `--dry-run` **不需要有会话** —— 先看清要发什么，
再决定要不要开会话，这个顺序本来就该如此。实测：真机那轮教学录制件（teach-4）出 50 条计划。

### 录制件的适用范围：id 是启动历史的函数

AUI 的裸 id 由**运行期**按对象声明顺序分配，被 `rn` 删掉后会被后续声明**复用**。
所以 id 只在**同一段启动历史**里有意义：换一条启动路径、换一份 `.42r`、换一个企业，
同一批对象的编号就全变。实测同一个作业的两条路径：对象集 1556/1558 相同，
而编号在共同条目里 **0 条相同**（偏移从 −31 到 −4274 都有，不是等量平移）。

作业怎么起由 T100 自己的脚本决定，路径的编号互不相同：

| 路径 | 服务器上真正执行的命令 | 在命令面上吗 |
|---|---|---|
| `r.r <作业>` | `fglrun $AZZ/42r/azzp000.42r <作业>`（azzp000 是 T100 的作业启动器，**不带 `-d`**） | 是（`session open`，标准运行录制） |
| debug 线 | `fglrun -d <LaunchRef>.42r <LaunchArgs>`，见 [../debug/README.md](../debug/README.md) | 是（`tt fep rec start` / `tt debug start`） |
| `r.d <作业>` | `$FBIN/42m/gendbg <作业>` → `$FGLDEBPATH/fgldeb <作业>.42r …`（fgldeb 是 `fglrun -d` 的前端、**T100 的调试器**） | **否**（已从命令面拿掉） |

**录制走哪条**：要"录了以后在断点下重放"就必须走一条带 `-d` 的路（`r.r` 的表与它不同），
而那条路用本仓库自己的 fgldb 就够 —— `tt fep rec start --env … --prog … [--ent …]`
一条命令做完：起隔离实例与代理 → `tt debug start --fglserver <本轮探测到的地址>` →
自动 `continue` 放行 → 前台常驻；用户面前**只有作业窗口**，收工落身份并入库。
要自己先下断点就加 `--no-continue`。

为什么不用 `r.d`（实测 2026-10-15，主机109）：`r.d` 起的是 **fgldeb**，而 fgldeb 自己的界面
也走 FGLSERVER —— 它在我们的隔离 GDC 上渲染成一个普通 AUI 表单
（`Window #9 name "fgldeb"`，297 条对象、70 个 `Action`、标题 `Debugger - …/asf_bsft001_wf.4gl`）。
于是那个会话录下来是**两棵树**：`#1 e2f009591dc3b368（298 条，调试器）`
+ `#2 493b95ef0b0df2af（3976 条，作业界面）`。它有三个代价：画面上多一个 T100 的调试器
（用户会误触）、录制件里混进调试器那棵树、断点台账/求值那套（我们只认 fgldb）用不上。
顺带两条实测：fgldeb 启动时报的 `_deb_getOmIdXPath: did not find an id for //Dialog[@active="1"]`
是它**启动顺序**上的查找失败 —— 它的 Dialog 18:04:04.384 才声明，而 AUI 流只从 18:04:02.335
到 18:04:04.467，那条错就在这 2 秒里；界面本身建全了、点得动（实测点它的 `Run/Continue`
就把它自己的状态栏推成 `currfunc="main()"` / `cstatus="running"`）。

**关键**：`r.d` 起的作业与调试线起的作业**编号表相同**（`493b95ef0b0df2af`，3976 条）——
两条都是 `fglrun -d <同一个 .42r> <同一套参数>`，区别只在谁当前端。所以：

| 路线 | 起作业的命令 | 调试器 | 录制件里有哪些连接 |
|---|---|---|---|
| `r.r` | `azzp000`（无 `-d`） | 无 | 作业界面 |
| debug 线 | `fglrun -d …` | 本仓库的 fgldb（无界面，全由 `tt debug` 驱动） | 作业界面 |
| `r.d`（只认历史件） | `gendbg` → `fgldeb` → `fglrun -d …` | T100 原生的 fgldeb（界面也进录制） | 调试器界面 + 作业界面（两条） |

⇒ 以前用 `r.d` 录的件**照样能在调试线上重放**（编号表一致）；而"录制 + 回放都用我们自己的
调试器"只需要 debug 线一条路。`LaunchPath` 里的 `r.d` 常量因此只为**识别**那些历史件而留。
**跨路径回放不会报错**：安全闸只判"这个 id 在不在当前界面"，而另一个会话的编号几乎总能"找到"
—— 实测 6/6 全部落到**别的**对象上（`#327` 从「加工类型编号」变成「重测/返工否」的勾选框）。
所以回放必须与录制走同一条路径，而"这一份录制还算不算数"由回放前的**指纹闸**回答（见下）。

判据（前置：一台能连的 T100 环境 + 一个作业名；约 3 分钟）：

```bash
tt fep session open --env <环境> --prog <作业>     # 第一次
# 约 10s 等界面稳定（tt fep fe status --ctl 127.0.0.1:7899 连续两次不变）
tt fep ui tree > a.txt
# 收工，再原地重复一次，得到 b.txt
diff a.txt b.txt                                  # 期望：无差异
```

同一段启动历史必得逐条相同的编号；换成另一条路径再取一次，则与 `a.txt` **逐条不同**
（对象集仍然相同）。

### 录制件的身份与指纹闸

一份录制有两样东西：**字节**（`<ID>-<连接>-{rt2gdc,gdc2rt,timeline}.log`）与**身份**
（`<ID>.meta.json`，收工时由 `Open`/`Close` 落，见 `open.go` 的 `writeMeta`）。身份里记的是
从字节里推不出来的那些：启动路径（`path.go`）、作业/环境/区域/ENT、画面接线（FGLSERVER）、
编译件标识（`BinaryID`，用 `--binary-ref` 触发一次只读 `ls -ln`）、连接清单，以及收工那一刻的指纹。
没有身份文件的录制件照样能用（指纹是**从字节复算**的），只是路径与编译件一栏空着 —— `rec show` 会这么说。

**指纹 = 整张"坐标 → 编号"表**（`fingerprint.go`）：每条对象取
`(种类, 作用域, 编号, 角色名)`，角色名取协议 `name` → `comment`，两样都没有就留空。
它**不含业务值**（`text`/`value`）：同一路径连录两轮，单号会从 `…003` 变成 `…004`，
拿值当签名会把能用的录制判成失效。指纹里因此也没有任何业务数据，可以随便复制流转。
摘要（`Fingerprint.Hash`）是这条路线的身份：**同一条启动历史必得同一张表**。

**闸门**（`gate.go`）在回放之前判两件事，任一不符就拒（`rec replay` 里落地）：

| 判什么 | 判决 | 报文 |
|---|---|---|
| 两边都知道启动路径且**不是同一条** | `path` | 说出录制件是哪条路径录的、当前会话是哪条 |
| **逐条连接**比：当前这条连接的编号表里，能在录制件的**某一条**连接里原样找到的比例 `< 99%` | `fingerprint` | 报最接近的那一对连接号、对得上/编号不同/只在一边的条数，外加前几条"录制里是哪个编号、现在这个坐标是哪个编号" |
| 录制件里没有界面 / 当前会话还没有界面 | `no-recording` / `no-live` | 分别说清是哪一边空着（含读不动的原因） |

**按连接比，不是把整个会话揉成一张表**：一个会话本来就有多条连接，而且它们可以是完全不同的
两棵树 —— 实测 `r.d` 会话是 `#1 e2f009591dc3b368（298 条，fgldeb 调试器的界面）` +
`#2 493b95ef0b0df2af（3976 条，作业界面）`。揉起来是 4273 条，拿它去比单连接的录制件，
同一条路径也会判成 93% 不过。所以判据是"存在一条连接对得上"，报文里会印出比过的每一对：
`#1↔#2 0% ／ #2↔#2 100%`。

阈值是量出来的：同路径、当前会话刚起完作业（录制件的**前缀**）时是 99–100%；
两份真界面不同的夹具（`testdata/fep-recordings/id-reuse.txt` 与 `openwin-with-dialogs.txt`）是 79%；
换一条启动路径（同一个作业 `bsft001_wf`、同一个环境与 ENT，`r.r` 的件 vs 调试线的会话）是 0%。
99% 落在这条空档里。判定由 `TestGateComparesPerConnection`、`TestGateReportsPerConnectionReadFailures`
与 `TestGateRefusesTwoDifferentRealInterfaces` 钉住（阈值两侧各有一条）。

现场量到的两条（主机109，2026-10-15，作业 `bsft001_wf`）：**同一条路径**上录的几份
（五份字节各不相同的调试线录制，跨两轮 `tt serve` 重启）指纹**逐条相同**：`493b95ef0b0df2af`（3976 条）；
同一个作业换成 `r.r` 路径录一轮就是另一张表：`d9431dc4b8f87339`（4961 条）。
要自己复现就照上面 M3 那条配方做两轮，把 `diff a.txt b.txt` 换成 `tt fep rec show <ID>` 比指纹。

`tt fep rec list` 是清单（默认把缓存与长期库都列出来，`--from` 可以指到别处），每份都当场复算指纹。

### 长期库：按 环境 + ENT

录制的**寿命由归属决定**（策略见 `retention.go`，用户 2026-10-15 定）：

| 谁录的 | 住哪 | 怎么清 |
|---|---|---|
| **用户主动**录的（面板点"开始录制"、`tt fep rec start`） | 长期库 `<数据目录>/records/<环境>/<ENT>/<录制 ID>/` | **不参与自动清理**：前端手动删（`tt fep rec rm <ID>`）、环境被删时**级联清** |
| **系统自动**录的（调试会话顺手录的上下文） | 缓存目录 `<数据目录>/fep/` | 走**统一清理框架**：`tt serve` 启动清超过 7 天的、`tt cache clear` 全清 |

用"住哪"表达寿命是有意的：清理框架只认 `config.CacheSubdirs` 里列出的目录
（`config.CleanCache`），所以"在不在那份清单里"就是"会不会被自动清"的唯一判据 ——
多一个"过期时间"字段就会有两个判据，而它们迟早对不上。身份文件里也照写
`origin` / `retention` 两栏，看一份录像为什么会被清掉时答案就在那儿。

```
<数据目录>/records/<环境>/<ENT>/<录制 ID>/
├─ <录制 ID>-<连接>-rt2gdc.log      界面（回放与指纹都从它重建）
├─ <录制 ID>-<连接>-gdc2rt.log      用户/注入的事件
├─ <录制 ID>-<连接>-timeline.log    两个方向按时间对齐
└─ <录制 ID>.meta.json              身份（含归属与留存）
```

**目录树就是索引**：一份录制是一个目录，环境与 ENT 是路径本身。刻意不另写一份 `index.json` ——
那会有两个真相，而它们迟早对不上（谁的 ID 列表、谁的时间戳）。录制 ID 不带环境与 ENT，
所以按 ID 找人是遍历三层（`Library.Find`），这个成本可以忽略，换来的是"ID 复制出去就能用"：
`rec show <ID>` / `rec replay --id <ID>` 会在缓存与长期库两处找，`rec list` 每行标出它住哪。

没给 ENT 的那一轮落 `ENT-未指定`：没给 `--ent` 表示沿用登录 shell 自己的 `$TOPENT`，
那个值我们不知道，不能假装知道。

搬的是**缓存里那一份**（`Promote`，移动不是拷贝），身份文件最后搬 —— 它是"这一份齐了"的标记。
`rec keep <录制 ID>` 可以手工补搬（当时用了 `--no-keep`、或从别处拿回来的一份）；
它要求身份文件在场，因为"进哪个环境、哪个 ENT 的层"只能从身份里读，猜不得。

**删除**两条路（`Library.Remove` / `Library.RemoveEnv`，都走 `config.RemoveTree` ——
逐文件自底向上删，不用 `os.RemoveAll`，理由见 `internal/config/cache.go` 的实测）：

- `tt fep rec rm <ID>`：删库里的一份（前端"手动清除"落地的那一下）。传进来的不在库里就拒；
  在缓存里的那类会提示它归统一清理框架管，不在那里删。**破坏性动作要 `--yes`**
  （与 `tt dev tzc unlock` / `tt update` 同一条规矩：录制不可再生，AI 不得自主加）；
- 环境被删 ⇒ 它的整段库跟着删：设置页保存环境列表时**级联**（`internal/web/server.go` 的
  `hHostsPut`：写配置成功之后再删文件，失败只记日志），**手改 config.json** 删掉的那种走
  `tt fep rec prune`（拿库里的环境目录名与配置里的环境列表对一遍，多出来的就是没人认领的）。

这条路线的判据（约 1 分钟，要一台能连的 T100 环境）：

```bash
tt fep rec start --env <环境> --prog <作业> --ent <ENT>   # 录一轮（用户主动录的）
tt fep session close                                       # 收工：落身份 + 自动入库
tt fep rec list                                            # 期望：那一行标着「长期库」并带路径与指纹
tt fep rec show <录制 ID>                                   # 期望：归属「用户主动录的」／留存「长期库」
tt fep rec rm <录制 ID> --dry-run                           # 期望：只说要删什么，一个字节都不删
tt fep rec rm <录制 ID>                                    # 期望：退 2 并要求 --yes（录制不可再生）
```

`config.CleanCache` 只碰 `CacheSubdirs` 里列出的目录，`records/` **不**在其中 ——
`TestCleanCacheNeverTouchesTheLibrary` 就是这句话的牙（清缓存与按年龄清理各验一次）；
级联那条由 `internal/web/cascade_test.go` 钉住（删环境删库、只改 host 不动库）。
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

## 细节去哪

- 命令面与用法（**录制 → 回放的配方**在这里）→ [`skills/tt-debug/SKILL.md`](../../skills/tt-debug/SKILL.md)
- 服务器连接与登录后的环境 → [../host/README.md](../host/README.md)
- 调试线的会话模型、协议锚点与画面接线 → [../debug/README.md](../debug/README.md)
- 窗口与截图 → [../view/README.md](../view/README.md)

协议的权威是录制件本身，不是这份文档；下面这条离线读一份仓库夹具：

```bash
tt fep ui tree --from testdata/fep-recordings/openwin-with-dialogs.txt
```
