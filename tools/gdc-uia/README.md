# gdc-uia — 通过 **GDC 自己的可访问性接口**（Qt QAccessible → Windows UI Automation）读写它的运行时 AUI

## 这是什么，为什么它比"猜协议形状"强

GDC 是 Qt 程序。Qt 会把**自己构建好的控件树**通过 QAccessible 暴露给 Windows 的
UI Automation（UIA）—— 那棵树就是**运行期 AUI 的落地形态**，而且：

| UIA 给的 | 例子（实测 apmt500_wf） |
|---|---|
| **AUI 对象身份**：`AutomationId` = 窗口 + AUI 名 | `w_apmt500_Window…w_apmt500_ToolBarItem_accept` |
| **动作名**（与 AUI `name` 一致） | `accept` / `insert` / `modify` / `query` / `cancel` |
| **控件类型**（决定能做什么） | `Button`(44) / `Edit`(16) / `ComboBox`(3) / `CheckBox` / `Table` / `Header` / `ListItem` |
| **栏位键**（就是 T100 的 `cmt_*`） | `Edit name='cmt_pmdl012'`、`ComboBox name='cmt_pmdl005'` |
| **屏幕矩形** | `rect=x,y w×h`（手册标注直接用） |
| **语义动作** | `InvokePattern` / `TogglePattern` / `ValuePattern.SetValue` |

关键点：**UIA 的语义动作走的是 Qt 自己的代码路径**（`InvokePattern` 落到
`QAbstractButton::click()`），所以 GDC 后续发出的 FEP 事件**就是它给用户点击时发的那一条** ——
不需要我们再猜"勾选框该发什么形状"。而且它**不需要抢焦点**（不占前台、不闪），符合无感要求。

### 已实测（2026-10-11）

```
powershell -File invoke.ps1 -TargetPid <gdcPid> -AutoIdSuffix "ToolBarItem_insert"
→ INVOKED w_apmt500_Window…w_apmt500_ToolBarItem_insert
→ 协议侧效果与 `tt fep do act '#3435'`（录入）逐项一致：采购日期 → "2026/10/11"、版次 → "0"
```

## 边界（同样重要）

- **表格格子确实在 UIA 里**（此处前版写错了，2026-10-11 纠正）：格子暴露为 `ControlType.DataItem`，
  带 `GridItemPattern`（**给出 row/col**）＋ `TableItemPattern` ＋ `ValuePattern` ＋ `InvokePattern`。
  两个实测注意点：① 它们**不挂在 `Table` 元素下面**（在树里是平的，别只看 Table 的子节点就下结论）；
  ② **只有当前行的格子被实例化**（实测 12 个 = 可见列数），滚动/换行后重新枚举。
  另：**在只读（DisplayArray/浏览）网格上 Invoke 格子不会产生任何 FEP 事件**（实测）；
  要验证"点格子会不会让 GDC 发出 currentRow/currentColumn"，必须先把作业推进到**可编辑**状态。
- UIA 树里 `name` 有时是中文标签（`单价含税否`）、有时是 AUI 名（`cmt_pmdl012`）——
  按 `AutomationId` 的后缀匹配最稳（脚本都这么做）。
- 需要有可访问性桥。实测 GDC 3.20.25 在 `QT_ACCESSIBILITY=1` 下就能被 UIA 看到。
  无感前提下这不增加任何东西（只是让它把已有的树报出来）。

## 分工建议

| 场景 | 走哪条 |
|---|---|
| 工具条动作、按钮、菜单 | **UIA Invoke**（形状天然正确） |
| 栏位写值、下拉选择、勾选框（**非表格**） | **UIA ValuePattern.SetValue / TogglePattern** |
| 表格里的格子（值、勾选、光标） | **FEP 协议**（UIA 看不到格子）＋ `{网格}{currentRow}{currentColumn}` 定位 |
| 读数（"这一步到底变了什么"） | 两边都行：UIA 读控件值更直观，协议读数更权威（运行期推的就是真相） |
| 手册标注需要的矩形 | **UIA 的 rect**（比截图量像素稳） |

## 用法

```powershell
# 1) 列出控件树（写入 uia-tree.txt，UTF-8）
powershell -NoProfile -File tree.ps1 -TargetPid <gdcPid> -MaxNodes 1000

# 2) 调用一个动作（AutomationId 后缀匹配）
powershell -NoProfile -File invoke.ps1 -TargetPid <gdcPid> -AutoIdSuffix "ToolBarItem_insert"

# 3) 读一个控件的值 / 勾选状态
powershell -NoProfile -File read.ps1 -TargetPid <gdcPid> -AutoIdSuffix "cmt_pmdl012"

# 4) 写值或勾选（有 ValuePattern 就 SetValue，否则 TogglePattern）
powershell -NoProfile -File setvalue.ps1 -TargetPid <gdcPid> -AutoIdSuffix "cmt_pmdl012" -Value "ABC"
```

GDC 的 pid 用 `tt fep session ps` 拿。

脚本一律 **纯 ASCII**（PowerShell 5.1 读无 BOM UTF-8 会乱码，中文只在数据里出现）。

## 实测结论：**能不能控、好不好控**（2026-10-11 普查，250 个元素）

| 控件类型 | 数量 | 可寻址 | 可操控 | 便利度 | 实测备注 |
|---|---|---|---|---|---|
| `Button`（工具条动作） | 42 | ✔ `AutomationId` 后缀 | ✔ **Invoke** | **高** | **已验**：Invoke `ToolBarItem_insert` 与 `tt fep do act '#3435'`（录入）协议效果逐项一致 |
| `MenuItem` | 8 | ✔ | ✔ Invoke | 高 | 同 Button |
| `CheckBox` | 2 | ✔ `name`（中文标签） | ⚠ Toggle **不产生 FEP 事件** | 中 | Qt 的 `TogglePattern` 走 `setChecked()`，而 GDC 监听的是 `click()`；**用 Invoke**（= `click()`） |
| `Edit`（栏位） | 16 | ✔ `name` = `cmt_*` | ✗ **SetValue 报告成功但不落地**（UIA 侧读回为空、协议侧也没动） | 低 | 写值仍走 FEP 协议（已验证可行） |
| `ComboBox` | 3 | ✔ `name` = `cmt_*` | ✗ 只有 Value，没有 Expand/Selection | 低 | 选项走 FEP 协议 |
| `Table` + `DataItem`（表格格子） | 1 + 12 | ⚠ **无 name/id**，只有 rect；但 `Table` 带 Grid/TableItemPattern（可按 (row,col) 取项） | ✗ 未见 Value/Toggle | 中 | 写格子/勾选行走 FEP 协议 |
| `Header`（表头） | 52 | ✔ | 只读（合理） | — | 带 rect，手册标注可用 |
| `Text`/`Group`/`Custom` | 94 | ✔ | Invoke | 中 | 容器与标签 |

**结论（分工，不是替代）**：

- **UIA 负责**：动作调用（形状天然正确、不抢焦点）、读数、**精确矩形**（手册标注直接用）；
- **FEP 协议负责**：写栏位值、选下拉、表格里的一切（格子值/勾选/光标定位）。

两个坑（脚本已处理）：

1. **按名字寻址会撞容器**：`FormField` 与 `Edit` 共享 `name=cmt_xxx` —— 必须加类型过滤
   （脚本的 `-Type`），否则 `SetValue` 会打到 FormField 上：**报告成功、值不落地**。
2. **勾选框不能用 Toggle**：要 `Invoke`（见上表）。

## GDC 二进制取证：能不能编译/反编译？源码拿得到吗？

**事实（2026-10-11 实测）**：

| 问题 | 答案 |
|---|---|
| 能拿到源码吗 | **不能**。Four Js 闭源发行（`C:\Program Files\FourJs\gdc`，v3.20.25），**无 PDB/调试符号**。可读的源码只有 3 个 JS（`gbc.bootstrap.js` / `webchannelbridge.js` / `webchannelurbridge.js`，属 Web/GBC 侧） |
| 能全量反编译吗 | **技术上能，但没有必要**。`gdc.exe` 只有 **7.5 MB**（MSVC x64，Qt5），**5353 个修饰名/类名（RTTI）**、**63 个 .cpp 源文件名** 都在；定向反编译 1–2 个函数可行，但拿到的只是伪代码 |
| 那还要不要猜 | **不需要**。二进制自己就把答案写着，再加上真机语料，两者互证 |

### 从 gdc.exe 里直接抠出来的（无需反编译）

**① GDC 能发出的全部事件——只有 7 种**（这才是"形状空间"的全集）：

```
{ActionEvent 0{{idRef …}}}                {KeyEvent 0{{keyName …}}}
{ConfigureEvent 0{{idRef …}}              {DestroyEvent 0{{status …}{message …}}}
{DragDropEvent 0{{idRef …}}}              {FunctionCallEvent 0 {{result …}}{errorMessage …}}
{RowSelectionEvent 0{{idRef …}}}
```

**② 架构图（63 个源文件名）**：

- `N*` = AUI 节点类：`NDialog` `NForm` `NFormField` `NGroup` `NTable` `NScrollGrid` `NMatrix` `NMenuAction` `NToolBarItem` `NWindow`
- `W*` = Qt 控件：`WTableView` `WFrozenTableView` `WLineEdit` `WComboBox` `WPushButton` `WScrollView` `WToolButton`
- `F*` = 栏位控件：`FButtonEdit` `FComboBox` `FRadioGroup` `FSpinEdit` **`FTableEdit`** `FTextEdit`
- **`CFEAui.cpp`** = FEP 序列化层（"事件形状"就在这层）；`Node.cpp` / `ActionNode.cpp` / `NodeSelection.cpp`
- **`QA.cpp` / `QAEventPlayer.cpp`** = QAccessible（本目录脚本用的那座桥，GDC 自带）

**③ 勾选框的权威字面值**：AUI 类型声明自己写着（真机原样）：

```
{CheckBox 340 {{width "10"} {text "单价含税否"} {valueChecked "Y"} {valueUnchecked "N"} {sizePolicy "dynamic"} {comment "cmt…"}}}
```

### 真机语料里"勾选"的原始字节（这是最终判据，不是推断）

```
event _om 21{}{{ConfigureEvent 0{{idRef "5195"}{ currentColumn "0"}{ currentRow "4"}}}{ConfigureEvent 0{{idRef "5196"}{ cursor "0"}{ cursor2 "0"}}}}
event _om 22{}{{ConfigureEvent 0{{idRef "6587"}{ value "Y"}}}}
event _om 23{}{{ActionEvent 0{{idRef "6539"}}}}
```

对应运行期侧：

```
un 5195 {{size "5"} {dialogType "InputArray"} {active "1"} {currentColumn "0"}}   ← 5195 = 明细网格；勾选列 = 第 0 列
un 6587 {{value "N"}}  →（前端写入）→ un 6587 {{value "Y"}}                      ← 6587 = 第 4 行"勾选格"的 Value 节点
```

**规则（有出处，可机械执行）**：

> 勾选某一行 = ① `ConfigureEvent{idRef <网格>}{ currentRow "<r>"}{ currentColumn "<勾选列下标>" }`
> ② `ConfigureEvent{idRef "<该行勾选格的 Value 节点>"}{ value "<valueChecked 字面值>" }`
> ③ `ActionEvent{idRef <accept>}`

注意：**失败往往不是形状错，而是目标节点选错** —— 同一窗口下会残留旧实例的同名 Value 节点，
必须取**最新声明（Seq 最大）**那一个（这正是 `internal/fep` 的 `Seq` 规则存在的原因）。
