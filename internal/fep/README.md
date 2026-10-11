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
