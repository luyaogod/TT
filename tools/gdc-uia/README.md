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

- **自绘的表格格子不在 UIA 里**：实测明细网格 `s_detail1` 在 UIA 里**只有 Table 本身**，
  行列与格子都是 GDC 内部绘制的 —— 所以"往某个格子写值/勾选某行"仍然要走 FEP 协议。
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
