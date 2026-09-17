<!-- 恢复说明（非原文） -->
<!--
本文件是 D:\我的项目\T100设计器\README.md 的逐行恢复副本。
原文件已被删除（同批被删的还有 tools\tzccli\tzc.py、tzc_audit.py）。
来源：上一轮会话 transcript
  C:\Users\18526\.workbuddy\projects\d-我的项目-T100设计器\
    1cb22e5f-3517-42cc-9db6-7faf612b01cd.jsonl  第 12 条记录（Read 工具结果）
  该 Read 结果 787 行 / 39378 字符，未截断，已逐行剥除行号前缀后原样保存。
本文件是 tdev tzc 实现的不变量依据（§3.8 I1–I15）与设计器行为依据，
实现代码中的 file:line 注释均指回本文件或反编译源码 D:\我的项目\T100设计器\。
-->
# T100Designer · tzc 包处理 & 代码编辑器

> **读者**：将要阅读 / 修改本仓库的 AI Agent 或工程师。
> **本文自包含**（不依赖任何其它文档）。所有关键结论都给出 `相对路径:行号`。
>
> **证据标记**（务必区分，别把推断当事实）：
> 🟢 反编译源码直证 ｜ 🟡 对发行版 DLL 做反射确证 ｜ 🔵 用 105 个真实生产 `.tzc` 实测确证 ｜ ⚪ 推断（未验证）
>
> 仓库根 = `D:\我的项目\T100设计器`。所有相对路径以此为基准。

---

## 0. 三条硬约束 —— 先看这个，否则会白干

### 0.1 有一大块关键代码**不在本仓库**，且不可重建

代码编辑器的「**可编辑区间模型**」（框架内 / 框架外的编辑权限）不在反编译产物里，
它在发行版的 **`ICSharpCode.AvalonEdit.dll`（被厂商打过补丁的 AvalonEdit）** 内。🟡

```
D:\APPS\T100设计器_1.0.0.251_免安装\ICSharpCode.AvalonEdit.dll
  AssemblyVersion = 4.3.1.9429   （与官方 4.3.1.9429 同号，但是魔改版）
  类型总数 415，其中厂商新增：
    ICSharpCode.AvalonEdit.ContentType                        [Flags] ADP=1 SEC=2 READONLY=4 OPEN=8
    ICSharpCode.AvalonEdit.Document.SegmentObject             区间基类
    ICSharpCode.AvalonEdit.Document.ISegmentObject / ISegmentTree
    ICSharpCode.AvalonEdit.Document.SectionObject : SegmentObject    （框架区段）
    ICSharpCode.AvalonEdit.Document.EditObject    : SegmentObject    （新增点）
    ICSharpCode.AvalonEdit.Editing.HasReadOnlyDocumentProvider      区间表 + 编辑门禁
    ICSharpCode.AvalonEdit.Editing.EditableAreaMargin              左侧可编辑区边距
    ICSharpCode.AvalonEdit.Editing.IReadOnlySectionProvider / NoReadOnlySections
    ICSharpCode.AvalonEdit.CodeCompletion.DesignerCompletionWindow / DesignerInsightWindow
    ICSharpCode.AvalonEdit.CodeCompletion.GeneroCodeCompletionBinding
    ICSharpCode.AvalonEdit.Bookmark.BookmarkBarManager / BookmarkBarMargin / BookmarkBase
    ICSharpCode.AvalonEdit.Highlighting.HighlightingDiffColorize
```

**入口**：`TextDocument` 上有一个 **public 字段**（不是属性）：

```csharp
// ICSharpCode.AvalonEdit.Document.TextDocument
public ICSharpCode.AvalonEdit.Editing.HasReadOnlyDocumentProvider SectionProvider;
```

反编译源码里有 **67 处** `SectionProvider` 调用，**0 处**定义。
`HasReadOnlyDocumentProvider` 的公开 API（🟡 反射得到，这些就是源码能用的全部能力）：

```
bool?  IsEditable(int offset)                        // 单点判权
bool   CanInsert(int offset)                         // 唯一的写入闸门
string GetText(Guid id) / GetText(ContentType, Guid)
SegmentObject Find(int lineNumber) / Find(ContentType, int) / Find(Guid) / Find(ContentType, Guid)
SegmentObject FindContains(int lineNumber)
ReadOnlyCollection<SegmentObject> FindOverlappingSegments(int offset)
IEnumerable<Guid> GetDirtySegments()   void ClearDirtySegments()
IEnumerable<SegmentObject> GetSectionSegments() / GetDeletableSegments(ISegment)
void   AddSection(int offset, SectionObject)         // 装框架区段
void   AddDeletableSegments(int offset, SegmentObject)  // 装新增点
int    GetOffset/GetEndOffset/GetSectionOffset(int)
void   Attached() / Dettached() / Clear()
```

`SegmentObject` 的成员：`Key:PackageKey`、`ID:Guid`、`Offset`、`EndOffset`、`Length`、
`Parent`、`PreviousNode`、`NextNode`、`Children`、`HasChild`、`IsEditable`。
⚠️ 源码**从不写** `IsEditable`，是 DLL 自己算的（见 §0.2）。

### 0.2 依赖方向是**倒置**的 —— 不要随便改模型层的公开形状

`ICSharpCode.AvalonEdit.dll` 引用了**应用自己的程序集**：🟡

```
ICSharpCode.AvalonEdit 4.3.1.9429  引用 →  SpecDesignerCommon 1.0.0.0
                                          Infrastructure     1.0.0.0
                                          SpecDesigner.Controls 1.0.0.0
                                          Microsoft.Practices.Prism
例：SegmentObject.Key 的类型是 SpecDesignerCommon.PackageKey
```

后果：

| 你不能做的事 | 原因 |
|---|---|
| 用官方 AvalonEdit 4.3.1 替换这个 DLL | 缺 `ContentType` / `SegmentObject` / `SectionProvider` 等类型，编译/运行直接崩 |
| 重新编译这个 DLL | 厂商没提供源码 |
| 随意改 `SpecDesignerCommon.PackageKey`、`Infrastructure` 里被它引用的类型（改名/改签名/改可见性） | 二进制 DLL 是**对着当前形状编译**的，运行时会 `TypeLoadException` / `MissingMethodException` |
| 认为「改 SpecDesignerCommon 只影响业务层」 | 编辑器的区间模型也依赖它 |

### 0.3 全仓还依赖一个无源码的厂商 DLL

`SpecDesigner.Controls.dll`（3.99 MB / **70 个类型**，🟡 已反射确认），被 **10/16** 个 `.csproj`
用 `HintPath` 引用。源码里**没有**这些类型，所以下面这些调用点是硬边界：

```
SpecDesigner.Controls.Controls.DesignerMessageBox / DesignerMessageBoxWindow   ← 全应用的消息框
SpecDesigner.Controls.Controls.AutoFilteredComboBox / WatermarkTextBox / WatermarkButtonEdit
SpecDesigner.Controls.Controls.PropertyLabel / SpecUniformGrid / TabControlEx
SpecDesigner.Controls.Controls.MasterDetailView / DetailView                   ← FormDataEditor 用
SpecDesigner.Controls.Controls.TextSearch / ScrollBarHelper / IconHelper / DialogHelper
SpecDesigner.Controls.Controls.BrushEditor.*   （21 个类型：主题/颜色编辑器）
SpecDesigner.Controls.AdornedControl.*         （4 个类型：自定义 Adorner 宿主）
SpecDesigner.Controls.Converters.*             （VisibilityConverter / InvertBoolConverter / CitedForEnabled）
SpecDesigner.Controls.Theme / GenericTheme / DarkTheme    ← 主题资源
```

**所有第三方 DLL 都在 `D:\APPS\T100设计器_1.0.0.251_免安装\`，
各 `.csproj` 用 `HintPath` 指过去，不走 NuGet。这个目录不能删、不能移。**

其余第三方依赖（同样来自该目录）：`Xceed.Wpf.AvalonDock`(+`.Themes.VS2013`)、
`Xceed.Wpf.Toolkit`（`BusyIndicator` 在这里）、`Microsoft.Practices.Prism`(+`ServiceLocation`)、
`Renci.SshNet`、`AppLimit.NetSparkle.Net40`、`ICSharpCode.SharpZipLib`、`Newtonsoft.Json`、
`DocX`(Novacode)、`System.Windows.Interactivity`、`Microsoft.Expression.Drawing`。

> 反例提示：`CustomizedRoutedCommand`（`CodeEditWindow/Helper/CustomizedRoutedCommand.cs:7`）和
> `BusyIndicator` **在源码里有**，别把它们当成闭源类型。
> 判断某个类型在哪，最快的办法是反射那个 DLL（见 §7 的方法）。

---

## 1. 30 秒定性

T100 ERP 的「**规格书设计器 + 4GL 代码编辑器**」客户端。WPF + Prism(仅 EventAggregator) +
AvalonDock。它**不直连数据库**：所有业务数据都靠 SSH/Telnet 在服务器上跑 4GL shell 程序
（`adzp050` 下载 / `adzp080` 上传 / `adzp201` 签入 …）导出成本地 `.tz*` 包，再本地解析。

| 项 | 值 | 证据 |
|---|---|---|
| 程序 | T100Designer 1.0.0.251，厂商 DSC，Copyright © 2013 | `T100Designer/Properties/AssemblyInfo.cs` |
| 描述 | Specification Designer & Code Editor | 同上 `AssemblyDescription` |
| 运行时 | .NET Framework **4.0**（16 个工程） | 各 `.csproj` 的 `TargetFrameworkVersion` |
| 规模 | 16 工程 / **749** `.cs` / **105** `.xaml` / 1188 文件 | 🔵 实测 |
| UI | WPF + AvalonDock + **魔改 AvalonEdit 4.3.1.9429** | §0.1 |
| 反编译 | dnSpy，未混淆未签名；**无 `.pdb`，故局部变量/参数名丢失**（`text`/`num`/`arg0` 是反编译损失，不是原风格） | — |

---

## 2. 工程地图

```
第4层  T100Designer (54cs, 唯一可执行/WinExe)
          │  SpecDesigner/Main/ = 启动骨架；Commands/MenuCommands.cs = 全部菜单命令
第3层  FormEditor(89) SpecEditor(76) FormDataEditor(10) TableViewer(6) Output(6) Search(11)
          │
第2层  SpecDesignerCommon(308) ──────── CodeEditWindow(68cs/16xaml)
          │                                    │
第1层  UndoRedoFramework(9) CustomException(3) Infrastructure(76) CodeEditor.FglAnalysis(12)
                                                 ↑ 只被 CodeEditWindow 源码级引用
        DifferenceEngine(14) —— 被 CodeEditWindow 引用
```

`SpecDesignerCommon` 是真正的公共层，被**所有**工程引用。子目录规模：
`ViewModel/` 76、`Events/` 65、`site/` 31、`UndoRedo/` 25、`Connection/` 20、`Helpers/` 20。

**关键入口**

| 角色 | 文件 |
|---|---|
| 启动 | `T100Designer/SpecDesigner/Main/SpecDesignerApp.xaml.cs`（`OnStartup`） |
| 主窗口 | `T100Designer/SpecDesigner/Main/SpecDesignerMainWindow.xaml`（1523 行 / 127 个 `<MenuItem>`）+ `.xaml.cs`（523 行；`BindMenuCommands()` 在 `:403`，93 处 `new CommandBinding`） |
| 命令总表 | `T100Designer/SpecDesigner/Commands/MenuCommands.cs`（**3924 行**，96 处 `new RoutedCommand(`） |
| 文档总管 | `T100Designer/SpecDesigner/ViewModels/EditorWorkspace.cs`（537 行，单例 `This`） |
| 包模型 | `SpecDesignerCommon/TzpManager.cs`（940 行） |
| 解包/打包 | `SpecDesignerCommon/PackageManager.cs`（606 行） |
| 代码程序模型 | `Infrastructure/SpecDesigner/Infrastructure/Model/ProgramInformation.cs`（1162 行）、`AddPointModel.cs`（1317 行）、`SectionModel.cs`（175 行） |
| 编辑器 | `CodeEditWindow/Helper/CodeEditorManager.cs`（1717 行）、`View/CodeTextEditor.cs`（2085 行） |

> ⚠️ `T100Designer` **的 `Main` 方法在源码里找不到**：它由
> `<ApplicationDefinition Include="SpecDesigner\Main\SpecDesignerApp.xaml">` 在编译期生成。别以为代码缺失。

**服务器命令出口**：`SpecDesignerCommon/Connection/ConnectionManager.cs`（824 行）

| 命令 | 用途 | 行 |
|---|---|---|
| `r.r adzp050 DIR '<dir>\' <TYPE> …` | 下载（`SPEC`/`CODE`/`CSPEC`/`RSPEC`/`4RP`/`GCODE`） | `:236` `:287` |
| `r.r adzp080 '<file>' <module> <TYPE> …` | 上传 | `:373` `:421` `:737` `:773` |
| `r.r adzp051 upload\|download` | 简易表单 | `:457` `:486` |
| `r.r adzp052` | 独立功能设定上传 | `:523` |
| `adzp020` / `adzp070` | 基础数据更新 / 重生成 | `:631` `:647` |
| `adzp062 <prog> <ver> …` | 重新编译 | `:683` |
| `adzp064 <prog> <type>` | **回标准**（把客制点还原） | `CodeEditWindow/View/CodeTextEditor.cs:1171` |
| `adzp085` | 设 FreeStyle | `T100Designer/…/MenuCommands.cs:2839` |
| `adzi520 <prog>` | 区段改动后的联动 | `CodeEditorManager.cs:493` |
| `adzi261` / `adzi262` | 通用函数 / 通用取值函数 | `ConnectionManager.cs:805` `:819` |

**落盘位置**：站点配置 `%APPDATA%\SpecDesigner\sites.xml`（AES，密钥硬编码在
`site/SiteFileHelper.cs:51-75`）；偏好 `Preference.xml` / 布局 `LayoutSetting.xml`（明文）；
工作区 = `Connection.Workspace`（默认 `C:\TT`），基础数据在 `<WS>\mta\`，日志在 `<WS>\log\`。

---

## 3. tzc 包处理

### 3.1 容器层

`.tzc` 是 **普通 zip**（ICSharpCode.SharpZipLib）。生产包里就 4 个条目：🔵

```
<prog>.4gl     服务器产出的完整程序
<prog>.tap     客制增量 XML
<prog>.tgl     框架骨架（含插入点/区段标记）
ver            无扩展名，版本闸门
```

**类型完全由扩展名决定**，不看内容 —— `SpecDesignerCommon/TzpManager.cs:150-182`：🟢

| 扩展名 | `TzpType` | 含义 | 上传代号 |
|---|---|---|---|
| `.tzs` / `.tzv` | `Form` | 表单规格 / 简易表单 | `SPEC` |
| `.tzc` | `Code` | 4GL 程序 | `CODE` |
| `.tzf` | `Code` | 独立功能程序（`isIndFun=true`） | `CODE` |
| `.tzx` | `Code` | 差异包（`IsDiff=true`） | `CODE` |
| `.tzd` | `CodeSpec` | 4GL 代码规格 | `CSPEC` |
| `.tzr` | `ReportSpec` | 报表规格 | `RSPEC` |
| `.tzg` | `ReportCode` | 报表程序 | `GCODE` |
| `.tzt` | `Report` | 报表模板 | `4RP` |

`ver` 由 `PackageManager.SeekReleaseVersion` 用 `ZipFile.GetEntry("ver")` **按名字取**
（`PackageManager.cs:589-604`），主次版本必须与客户端一致，否则 `VersionIncompatibleException`
（`TzpManager.cs:395-404`）。→ **这个条目名不能改、不能加目录前缀**。🔵 生产包 `ver` 值都是 `1.0`。

### 3.2 条目 → 加载器分派表

唯一入口 `PackageManager.Unpacking()`（`SpecDesignerCommon/PackageManager.cs:68-205`）。
每个类型有**必须存在的条目**清单，缺一个直接抛 `FileNotFoundException`（`:194-203`）：

| 类型 | 必须 | 可选 |
|---|---|---|
| `Code` / `Report` | `.tap` `.tgl` | `.tap2` `.4gl` `.src`+`.apt`（diff）`.bdx` `.xml` `.4ad` |
| `Form` | `.tsd` `.4fd` | `.4fdref`（简易表单）`.str` `.4ad` `.merge` `.bdx` |
| `CodeSpec` | `.csd` | `.bdx` |
| `ReportSpec` | `.rsd` | `.bdx` |

条目 → `TzpManager` 字段（`:116-191` 的 switch）：

| 条目 | 字段 | 说明 |
|---|---|---|
| `.tap` | `TAP` | 客制增量：`<other>` + `<point>` + `<section>` |
| `.tgl` | `TGL` | 框架骨架 |
| `.4gl` | **无**（`Load4glFile` 丢弃 content） | 见 §3.3 |
| `.tap2` | `UpdatedTAP` | 本次异动增量（删点 + 改点 + `status="u"` 的区段），服务器合并用 |
| `.src` (`*.4gl.src`/`*.tap.src`) | `DIFF_SRC` / `DIFF_TAP` | 差异包的「原版」 |
| `.apt` | `ADT` | 差异包的原始 TAP |
| `.bdx` | `ElementBindings` | 控件↔栏位绑定（表单） |
| `.4ad` / `.str` / `.xml` | `CustomActionDefaults` / `FormLocalizedStrings` / `infoXML` | — |
| `ver` | — | 不参与解析，纯透传 |

### 3.3 三件套的真实关系

**`.4gl` 客户端从不读回。** 证据链：🟢

1. 装载 `case ".4gl": tzpManager.Load4glFile(text2);`（`PackageManager.cs:132-134`）
2. `Load4glFile(string content) { _readingFlags["4gl"] = false; }` —— **`content` 参数根本没用**（`TzpManager.cs:378-381`）
3. `TzpManager` 里没有任何 `Load` 路径给 `FullCode` 赋值；只在 `SaveFullCode` 写（`:426-431`）
4. 保存时 `programInfo.FullCode = this.Editor.Text;`（`CodeEditorManager.cs:278`）

**`.tgl` = 框架骨架**，用 Genero 花括号注释（`{ }`）埋标记，所以骨架本身可编译：

```genero
{<point name="other.function"/>}                      ← 集合锚点：新增点从这里注入
{<point name="before_input" edit="c" mark="Y"/>}      ← 普通插入点（可带 edit/mark）
{<section id="aapp131.main" type="s" >}
   ... 标准代码 ...
{</section>}
{<section id="aapp131.other_function" readonly="Y" type="s" >}
{</section>}
```

标记正则（`CodeEditorManager.cs:1694-1701`、`:1709-1715`）：🟢

| 标记 | 正则 | 备注 |
|---|---|---|
| point | `\{<point\s+name="(\S+)".*\s*/>\}` | `edit="s\|c"` → `EditEnv`；`mark="Y"` → `IsMarkable` |
| section 起 | `\{<section\s+id="(?<name>\S*)".*>` | ⚠️ TGL 用 `type=`，TAP 用 `src=`（`SectionModel.SRC` 读的是 `src`，`SectionModel.cs:151`） |
| section 止 | `\{</section>\}` | 计数必须与起点相等 |
| readonly | `readonly="(?<value>\w)"` | — |

**`.tap` 根元素是 `<add_points>`**（不是 `<tap>` —— 那只是 `TzpManager.TAP` 属性名）。真实形态：🔵

```xml
<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<add_points prog="aapp131" std_prog="aapp131" erpver="3.0" module="AAP" ver="66"
            env="c" zone="topprd" booking="N" type="M" identity="s"
            section_flag="N" designer_ver="1.0" login_user="tiptop"
            std_section_verify="N" std_to_cus="" pre_compile="Y" topind="sd">
  <other>
    <code_template value="P" status=""/>
    <free_style value="N" status=""/>
  </other>

  <point name="function.aapp131_qbe_clear" order="1" ver="63" cite_std="N" new="Y"
         ind_fun="sd" ind_extra="N" ch="" status="" src="s" readonly=""
         mark_hard="N" modi_by_topstd="" mapping="function.aapp131_qbe_clear">
    <![CDATA[# 注释描述
PUBLIC FUNCTION aapp131_qbe_clear()
   ...正文...
END FUNCTION]]>
  </point>

  <section id="aapp131.main" src="c" status="u" ver="1" modi_by_topstd="">
    <![CDATA[ ...覆盖后的框架正文... ]]>
  </section>
</add_points>
```

⚠️ **`<point>` 的 CDATA 存的是「整块」**，不是只有正文：
`ToXML()` 写 `IsSelfDefinition ? ToString() : Content`（`AddPointModel.cs:1105`），
而 `ToString()`（`:1008-1060`）= `Description` + `Scope` + `FunctionName` + `Content` + `END xxx`。
内存里的 `Content` 属性**才只有正文**（头尾在解析时被正则剥掉，`:357-401`）。

`<point>` 属性语义（`AddPointModel.Create` `:1143-1210` + `IsEditable` `:691-747`）：🟢

| 属性 | 值 | 作用 |
|---|---|---|
| `name` | `function.` / `dialog.` / `report.` 前缀 或 裸名 | **前缀决定 `DefinitionType`**（`:545-580`）。带前缀 = 自订定义（`IsSelfDefinition`），只有函数体可编辑；裸名 = 普通插入点，整块即正文 |
| `order` | int | 注入 `other.function/dialog/report` 锚点的排序键（`SortIndex`） |
| `new` | `Y`/`N` | **只有 `new="Y"` 的点可被删除**（`CodeEditorMainWindow.xaml.cs:521`） |
| `status` | `""` / `" "` / `c` / `u` / `d` | 无变化 / 无变化 / 新增 / 修改 / 删除。⚠️ 未变的点被规范化成 **`status=" "`（一个空格）**（`ToXML` `:1086-1112`）；`ToXML` 对**恰好等于 `CREATE`** 的点 `return null`，即**没写内容的新点不会被序列化** |
| `src` | `s`/`c`/`m` | 来源：标准/客制/? —— 决定编辑权 |
| `cite_std` | `Y`/`N` | 引用标准程序时该点内容由标准程序提供 → **不可编辑**（`:703-706`） |
| `ind_fun` | 行业别码 | `.tzf` 独立功能程序里行业别不匹配 → 不可编辑（`:695-698`） |
| `readonly` | `Y` | 一票否决（`:707-710`） |
| `mark_hard` | `Y`/`N` | 硬标记书签；区段保存时被清掉并移除标记（`CodeEditorManager.cs:445-449`） |
| `mapping` | 字符串 | 引用标准点的映射名 |

### 3.4 合成管线：TGL + TAP → 编辑器文档

`CodeEditorManager.LoadContent()`（`:1070-1080`）五步：🟢

```
① SectionProvider.Dettached()                     清空区间表
② sb = TGL
③ ProcessFunctionTypes(FUNCTION/DIALOG/REPORT)    把 {<point name="other.function"/>}
                                                  替换成该类型所有点的 {<point name="X"/>} 列表
                                                  （按 order 升序，Environment.NewLine 连接）
④ Editor.Text = sb                                此刻还是「骨架 + 占位符」
⑤ ProcessAddPoints()                              逐行扫 {<point name="X"/>}，
                                                  用 AddPointModel.ToString() **整行替换**
⑥ ProcessSections()                               给 {<section>} 区间注册 SectionObject
```

- **第 ③ 步是关键**：新增点**不写在 TGL 里**，只由 `other.<type>` 锚点按 `order` 注入。
  TGL 没锚点 → 那些点根本不会出现（`:1158` 的 `if (regex.IsMatch(...))` 守卫）。🟢
- 第 ⑤ 步 `ProcessAddPoints`（`:1179-1227`）：查不到同名 `<point>` 时调
  `ProgramInformation.Initial(name)` —— **造一个空点**（`ProgramInformation.cs:655-666`）。这是常态，不是错误。
- `ProcessEditableArea`（`:1230-1253`）：自订定义点用 FGL 解析器算正文范围：

```csharp
ParseResult r = FglParserQuickHelper.ParseFunction(model.ToString());
if (r == null) throw new FormatException($"TAP：'{model.Name}' has format error");  // :1238-1242
editObject.AddChild(model.ID, startLine + r.StartLineNumber, len - r.StartLineNumber - 1, endLine);
```

  解析不出来 → **整份文档打不开**（`CodeEditWindow/Helper/FglParserQuickHelper.cs`）。
- `ProcessSections`（`:1083-1139`）：区段名对不上 TAP 时，从 `src=="c"` 的区段 clone 模板，
  没有就找 `src=="m"`，再没有就抛「找不到区段」。以下三个区段**强制只读**：

```csharp
if (name == prog + ".other_function" || prog + ".other_dialog" || prog + ".other_report")
    sectionModel.IsReadOnly = true;                                        // :1127-1130
```

### 3.5 保存管线

```
SettingManager.SaveSetting(key)                    SpecDesignerCommon/SettingManager.cs:531
  └─ Publish SaveSettingEvent
       ├─ CodeEditorManager.SaveFile(key)          CodeEditWindow/Helper/CodeEditorManager.cs:48 → :269
       │    ├─ SaveContent()  → SaveADPContent()  + （SEC 模式才）SaveSectionContent()
       │    ├─ GenerateTGL()  → 把 status="u" 的区段正文写回 TGL 字符串
       │    └─ FullCode = Editor.Text
       └─ ResourceController.SaveFile(key)         Infrastructure/.../ResourceController.cs:106-116
            ├─ SaveCodeFile(TGL)      → _readingFlags["tgl"] = true
            ├─ SaveAddPoint(AddPoint) → _readingFlags["tap"] = true   ← TAP XML 这时才序列化
            ├─ SaveUpdatedAddPoint()  → _readingFlags["tap2"] = true
            ├─ SaveFullCode()         → _readingFlags["4gl"] = true
            └─ 各 Save* 内部都调 PackChecking() → 可能触发 PackageManager.Packing() 重写 zip
```

**回写目标的差异（最容易搞错的地方）**：

| | 框架外（`EditObject` / 新增点） | 框架内（`SectionObject` / 区段） |
|---|---|---|
| 采集者 | `SaveADPContent()` `:361-397` | `SaveSectionContent()` `:400-496` |
| 采集方式 | 整个 EditObject 区间 `GetText(id)` **原样** → `AddPointModel.Content` | **逐行**遍历；遇到 `EditObject` 行就写回 `addPointModel.TglTag`（原始 `{<point .../>}` 文本），否则写该行原文 `:434-470` |
| 落到哪 | `.tap` 的 `<point>` | `.tap` 的 `<section>` **+** `.tgl` 的区段正文（`GenerateTGL` `:314-352`） |
| 落 `.tgl` 吗 | **不落**（TGL 只有占位符） | **落** |
| 副作用 | — | `IsSectionModify=true` → TAP 根 `section_flag="Y"`；客制环境 + `Booking` 时 `RunProgram("adzi520 " + prog)` `:491-494` |
| `\a` 剥离 | IsDiff 时剥 `\a\r\n`/`\a` `:369-383` | 剥 `\a\r\n`/`\r\n\a` `:471` |

> **`TglTag` 折叠机制是「不破坏」的命门**：框架区段正文里嵌的自订函数，在保存区段时会被还原成
> `{<point name="X"/>}` 占位符。所以 **`.tap` 的 `<section>` 里绝不能出现展开后的函数正文**，
> 否则下次装载会重复展开 → 代码重复 / 结构错乱。

### 3.6 打包（`PackageManager.Packing`）

**什么时候重写 zip**：`TzpManager._readingFlags`（字段在 `:886`）语义是
「该条目在包里存在，且已用内存状态重写过」。`PackChecking()`（`:251-272`）：
**只要有一个 flag 还是 `false` 就直接 return，不打包**；全 true 才 `Packing`，成功后重置为 false。
一次「保存」正好把 `tap`/`tgl`/`4gl` 三个都写一遍 → 所以**每次保存都重写整包**。

`Packing`（`PackageManager.cs:208-586`）遍历 zip 里**已有**的条目，认识的扩展名用内存状态覆盖，
其余原样拷贝。6 个「粗暴」行为（写工具时必须知道）：🟢

| # | 行为 | 影响 |
|---|---|---|
| 1 | **非原子写**：`File.Delete` → `File.Create` → 写入（`:569-572`） | 写到一半崩 = **原包彻底丢失** |
| 2 | **空条目会凭空消失**：认识的扩展名若内存状态为空串，走 `goto IL_0935; CloseEntry(); continue;`（如 `:257-260`） | 一个空 `.tgl` 会让条目从包里消失 |
| 3 | **时间戳全部重写**：`DateTime = DateTime.Now`（13 处） | 同内容每次保存 zip 二进制都不同 |
| 4 | **重新压缩**：`SetLevel(3)`（`:214`），未知条目也解压重压（`:537-548`） | 原本 store 的条目变 deflate |
| 5 | **强改 ACL**：给 `Everyone` 加 `FullControl`（`:581-584`） | 非 Windows/无权限时抛异常 |
| 6 | **注入 CRLF**：`text.Replace("><![CDATA[", ">\r\n<![CDATA[")`（`:261-262` 等 5 处） | `.tap` 变**混合换行**：元素间 CRLF，CDATA 内部 LF |

> 第 6 点的原因见 §3.3：`AddPointModel.Content` setter 第一步就是 `value.Replace("\r\n","\n")`
> （`AddPointModel.cs:328`）。**解析 `.tap` 时不能假设全局统一换行。** 🔵

另外：如果原包**本来没有** `.4gl` 条目，保存**永远不会新增一个**（循环只在遇到该条目时才写）。

### 3.7 point 名字的命名空间有**三层**（不是两层）🔵

```
other.function / other.dialog / other.report   → 集合锚点，只存在于 TGL，TAP 里没有
<裸名>（global.memo / input.a.page2.x / main.exit…）  → 框架预留插入点，TGL 通常有、TAP 通常没有
function.* / dialog.* / report.*               → 自订定义点，TAP 有，由锚点注入
```

实测（`aapp131(s).tzc`）：TGL 有 **113** 个占位符，TAP 只有 **44** 个 point。
全量 105 个包有约 **2.6 万处**「TGL 有、TAP 无」——**这是设计如此**，不是缺失。

反过来也有「TAP 有、TGL 无位置」的**孤儿点**（76/105 个包都有，如 `construct.c.pmdl002`）。
它们既不出现在编辑器里，**也不会丢** —— 因为 `AddPointModel.Create` 不设 `IsLoaded`，
而 `SaveADPContent` 的过滤条件是 `a.IsLoaded`（`:363`），所以设计器跳过它、原样 `ToXML()` 写回。

### 3.8 不变量（写 tzc 的工具必须守住）

| # | 不变量 | 检查方式 | 等级 |
|---|---|---|---|
| I1 | `{<section>}` / `{</section>}` 配对；区段 id 唯一且与 TAP `<section id>` 对应 | 计数 + 集合比对 | 错 = error |
| I2a | TGL 占位符在 TAP 里没有对应 `<point>` | — | **正常**，info |
| I2b | TAP 的 `function./dialog./report.` 点但 TGL **没有**对应 `other.<type>` 锚点 | 集合比对 | **error**（代码彻底不出现） |
| I2c | TAP 裸插入点在 TGL 里没有同名占位符 | 集合比对 | warn（惰性数据，看不到但不丢） |
| I3 | **区段正文里不得出现真实自订点的完整函数块**。判据要精确：拿 TAP 里已知的函数名去搜 `^\s*(public\|private)?\s*(FUNCTION\|DIALOG\|REPORT)\s+<名字>\s*\(`。**宽正则会误伤框架自己的 `DIALOG ATTRIBUTES(UNBUFFERED,…)`** | 用已知函数名匹配 | error |
| I4 | 自订点结构：`[# 注释行…]` + `[public\|private]` + `FUNCTION 名(参数)` + body + `END FUNCTION`。⚠️ **解析顺序：先剥 `public/private`，再匹配定义头**，否则 `PUBLIC FUNCTION x(p)` 匹配不上 `^\s*(function)+`。⚠️ 实测存在「带 `function.` 前缀但不是函数块」的点 | 设计器同款正则（`AddPointModel.cs:1243-1264`） | warn |
| I5 | `status` ∈ `{"" , " ", c, u, d}`；删点置 `d` 且**保留原内容** | 属性检查 | error |
| I7 | `ver` 必须存在且主次版本匹配客户端；**字节级原样保留** | 存在性 + 哈希 | error |
| I9 | `.tap2` 必须与 `.tap` 的 diff 一致 | 重算比对 | warn（`--strict` 下 error） |
| I10 | `.4gl` **不等于**「TGL + 展开点」（见 §5.2） | 渲染比对 | 仅 info |
| I11 | UTF-8 **无 BOM**；CDATA 里不得残留 `\a`(0x07)；CDATA 内 LF | 字节检查 | error |
| I13 | 未知条目 / `ver` 字节级原样 | 逐条目哈希 | error |
| I15 | 引用标准程序（`prog != std_prog`）的包，部分 `.tap` 内容由 `<std_prog>.tmc` 提供（`Packing` 的 `.tap` 分支在**条目基名 ≠ 程序名**时会改用 `CiteTAP`，`PackageManager.cs:320-329`）→ **这种包不能整体重写 `.tap`** | — | ⚪ |

### 3.9 本仓库自带的 tzc CLI 工具链

`tools/tzccli/`（Python 3.8+，仅标准库，无需安装）：

| 文件 | 作用 |
|---|---|
| `tzc.py` | 主工具：`info / ls / outline / cat / render / get / set / add / rm / unpack / pack / check / selftest` |
| `tzc_audit.py` | 批量体检一个目录的所有 `.tzc`（含 `--roundtrip` 字节保真测试） |

```powershell
# 只读（零风险）
python tools\tzccli\tzc.py info    <pkg.tzc> [--json]
python tools\tzccli\tzc.py outline <pkg.tzc> [--json]     # 结构树（锚点会展开出注入的点）
python tools\tzccli\tzc.py render  <pkg.tzc> --annotate   # 合成完整 4GL，带区间横幅注释
python tools\tzccli\tzc.py get     <pkg.tzc> <点名> --body-only
python tools\tzccli\tzc.py check   <pkg.tzc> [--strict] [--json]

# 目录工作流（AI 主战场：只改纯 4GL，不碰 XML）
python tools\tzccli\tzc.py unpack <pkg.tzc> -o <dir>
#   改 <dir>/points/*.4gl 与 <dir>/sections/*.4gl
python tools\tzccli\tzc.py pack   <dir> -o <pkg.tzc> [--dry-run]

# 单点外科手术
python tools\tzccli\tzc.py set <pkg.tzc> <点名> -f new.4gl [-o out.tzc] [--dry-run]

# 批量
python tools\tzccli\tzc_audit.py <目录> --roundtrip
python tools\tzccli\tzc.py selftest        # 45 项内置自测，不需要真实包
```

`unpack` 产出：`manifest.json` + `source/{*.4gl,*.tgl,*.tap}` + `points/*.4gl` +
`sections/*.4gl` + `opaque/*`。

**三个高危开关默认全关**：`--allow-tgl-edit`（允许改框架骨架）、
`--allow-4gl-edit`（允许改 `.4gl`）、`--rerender-4gl`（按 TGL 重渲染 `.4gl`）。
退出码：`0` 成功 / `2` 包格式错 / `3` 不变量失败 / `4` 写入被拒 / `5` IO 失败。

**实现要点**（要自己写工具时照抄这两条，否则 diff 会爆炸）：

1. **不要用 XML 库解析再 `ToString()`**。写一个 CDATA 感知的扫描器
   （遇到 `<![CDATA[` 跳到 `]]>`），只记录元素区间，改写时**只重建 CDATA 内部**；
   属性顺序、空白、引号、换行全部自动保真。
2. **根标签名动态识别，属性改动「有则改、无则加」**（`XElement.SetAttributeValue` 语义）。

---

## 4. 代码编辑器

### 4.1 类层次与文件地图

```
ICSharpCode.AvalonEdit.TextEditor                    （魔改版，闭源）
  └ BaseTextEditor    CodeEditWindow/View/BaseTextEditor.cs   681 行
      └ CodeTextEditor CodeEditWindow/View/CodeTextEditor.cs 2085 行
          └ DiffTextViewer  CodeEditWindow/View/DiffTextViewer.cs 467 行（只读差异视图）
```

- **`TextEditor` 实例不在 XAML 里**，是代码 `new` 出来塞进 `<Grid x:Name="dockPanel"/>`
  （`CodeEditorMainWindow.xaml.cs:211/260`）。
- 主窗口 `CodeEditorMainWindow.xaml.cs`（898 行）+ `CodeEditorMainWindow.xaml`。
- 管理器 `CodeEditWindow/Helper/CodeEditorManager.cs`（1717 行）—— 装载/保存/展开/折叠/差异全在这。
- `SectionProvider` 调用点分布（67 处）：`CodeEditorManager` 29、`CodeTextEditor` 23、
  `BaseTextEditor` 4、`BookmarkTrackerBehavior` 4、`CodeEditorMainWindow` 3、其余各 1。
- 左侧边距 `EditableAreaMargin`（`:49`）可视化可编辑区。

### 4.2 两种编辑模式

`CodeTextEditor.Mode` 是 `ContentType`（`[Flags]`，见 §0.1）。UI 上是主窗口一个 checkbox，
处理函数 `CodeEditorMainWindow.xaml.cs:557-588`：🟢

| | `ContentType.ADP`（默认，未勾选） | `ContentType.SEC`（勾选） |
|---|---|---|
| 语义 | **框架外**：只编辑新增点 | **框架内**：编辑框架区段 |
| 变更标记 | `AddPointModel.Status \|= MODIFY`（`Helper/TextChangedBehavior.cs:31-47`） | — |
| 保存采集 | `SaveADPContent()` | `SaveSectionContent()` |
| 切出时 | — | 先 `SaveSectionContent()` |
| 切入时 | — | **弹框确认**（见下） |

进入 SEC 模式的闸门（`:565-586`）：

```
IsSectionModify == true（TAP 根 section_flag="Y"）  → 直接进，不弹框
ENV=="s" 且 login_user!="topstd"
    ├─ std_section_verify=="Y" → Yes/No 确认
    └─ 否则 → 弹「标准区段校验」警告后 return（进不去）
其他                                        → Yes/No 确认「变更为客制区段属性」
```

`Helper/EditModeConverter.cs`：`ADP → false`，其余 → `true`（绑 checkbox）。

### 4.3 可编辑性决策链

**框架外**（`AddPointModel.IsEditable`，`AddPointModel.cs:691-747`），顺序短路：🟢

```
1. .tzf 独立功能程序 且 Ind_fun != Topind 且 name != "global.memo_industry"  → false
2. Topind ∈ {"sd",""} 且 ENV=="s" 且 name=="global.memo_industry"            → false
3. 程序非标准件 且 cite_std=="Y"                                            → false
4. readonly=="Y"                                                            → false
5. EditEnv=="c" 且 ENV=="s"                                                 → false
   EditEnv=="c" 且 ENV=="c" 且 IsTopstdMode                                  → false
   EditEnv=="s" 且 ENV=="c" 且 !IsTopstdMode 且 SRC!="c"                     → false
6. IsTopstdMode 且 SRC=="c"   → 仅 status==CREATE 可编辑
   IsTopstdMode 且 SRC∈{s,m}  → 仅 status==NULL，或 (modi_by_topstd && MODIFY)
7. login_user=="topstd" 且 SRC=="c"  → 仅 status==CREATE 可编辑
```

**框架内**（`SectionModel.IsEditable`，`SectionModel.cs:101-133`）：🟢

```
0. 程序 Type=="G" 且 IsSectionModify → name ∈ {prog.other_function, prog.other_report} ? false : true
1. IsReadOnly（ProcessSections 强制置位，见 §3.4）     → false
2. TGL 标记里 readonly="y"                            → false
3. IsTopstdMode 且 SRC=="c"                           → false
   IsTopstdMode 且 SRC∈{s,m}                          → 仅 (无 modi_by_topstd) 或 (modi_by_topstd=="Y" && status=="u") 或 status==""
4. login_user=="topstd" 且 SRC=="c"                   → false
```

**唯一的写入门禁**是 `SectionProvider.CanInsert(offset)`：

```csharp
private static bool IsEditable(CodeTextEditor editor) {           // CodeTextEditor.cs:1400-1427
    // 选择区逐个 offset 查 CanInsert；无选区查光标 offset
}
```

### 4.4 编辑动作的行为差异

| 动作 | 框架外（新增点） | 框架内（区段） |
|---|---|---|
| 打字 | 只有落在 `EditObject` 子区间（正文本体）才行；函数头/`END FUNCTION` 被 `CanInsert` 拒绝 | 只有 `SectionObject` 且 `IsEditable` 才能打 |
| 重命名 | `CanFunctionModify`/`ExecutedRename`（`:791-829`）要求 `IsSelfDefinition && IsEditable`，弹 `FunctionInfoWindow` | `Mode==SEC` 时直接 `CanExecute=false` |
| 删除 | `ExecutedDelete`（`:832-845`）要求 `IsSelfDefinition && IsNew && IsEditable` —— **`new="Y"` 才能删** | 不支持（`:948-951`） |
| 回标准 | `adzp064 prog type`（`:1143-1173`），条件 `StdToCus=="Y"` | 同左 |
| 引用标准 | `CiteOrNot` 切 `cite_std`，仅非标准程序可用（`:945-976`） | — |
| 插代码片段 | `CanInsertCodeSample` 要求光标在某区间内且 `IsEditable`（`:885-915`） | — |
| 全选替换 | `BaseTextEditor.ReplaceAllFromCaret`（`:585-611`）逐行判断：`SectionObject` 直接跳到 `EndOffset`（跳过），`EditObject` 且 `IsEditable` 才替换 | 同左 |
| 替换（搜索） | `BaseTextEditor.ReplaceResultInfoSelected`（`:447-489`）：命中只读 `EditObject` 时不替换文本，改走 `Modify` 改名流程 | — |
| 搜索过滤 | `EditableAreaOnly` 只搜可编辑区（`BaseTextEditor.cs:281`/`:367`） | 同左 |
| 差异 | `CanDiffSingleContent`（`:1176-1201`）要求 `IsDiff` 且命中 ADP 区间 | — |

**会抛异常的情形**：

| 情形 | 位置 | 结果 |
|---|---|---|
| TGL 里 `{<section>` / `{</section>}` 数量不等 | `CodeEditorManager.cs:1093-1096`、`:324-327` | 抛「区段错误」 |
| `<point name>` 与内容里的函数名不一致 | `AddPointModel.cs:1200-1203` | `ComplexException` |
| 自订点的正文无法被 FGL 解析 | `CodeEditorManager.cs:1238-1242` | `FormatException` |
| 内容里没有 `FUNCTION x(...)` / `END FUNCTION` | `AddPointModel.cs:361-364` | 抛「函式内容错误」 |
| `Description` 有空行不以 `#` 开头 | `AddPointModel.cs:253-260` | 表单校验错误 |
| `FunctionName` 不以程序名开头 | `AddPointModel.cs:193-203` | 仅 INFORMATION 提示，不阻断 |

### 4.5 功能核账（不要相信"有这个功能"的说法）🔵

| 功能 | 有 | 实现位置 |
|---|---|---|
| 语法高亮 | ✅ | `CodeTextEditor.LoadSyntaxHighlighting()` `:1752`（xshd 路径在 `:1760`/`:1764`） |
| 行号 | ✅ | `base.ShowLineNumbers = true;`（`BaseTextEditor.cs:91`） |
| 搜索/替换 | ✅ | 自研，**不用** AvalonEdit 的 SearchPanel（`BaseTextEditor.cs:190/492/556/585`） |
| 书签 | ✅ | `ICSharpCode.AvalonEdit.Bookmark.BookmarkBarManager`（在魔改 DLL 里）+ `View/BookmarksWindow` |
| 智能补全 / 签名提示 | ✅ | `InitializeCodeCompletion()` `:1460`；UI 类 `DesignerCompletionWindow` / `DesignerInsightWindow` / `GeneroCodeCompletionBinding` **在魔改 DLL 里** |
| 差异对比 | ✅ | `Helper/DiffManager` + `DifferenceEngine` + `DiffTextViewer` / `DiffBaseOnStandardWindow` |
| 当前行高亮 / 框选标记 | ✅ | `CaretLineRenderer` / `TextMarkedRenderer` |
| 可编辑区左侧边距 | ✅ | `EditableAreaMargin`（魔改 DLL） |
| Vi 模式 | ✅ | `CodeEditorMainWindow.xaml.cs:235`（`TextArea.Options.EnableViMode`） |
| **代码折叠** | ❌ | 全仓 grep `FoldingManager`/`FoldingStrategy`/`NewFolding` **0 命中**（`FoldingSection` 只是 AvalonEdit 自带的空壳类型） |
| **括号匹配** | ❌ | 无 `BracketSearcher` / 缩进策略 |
| **错误波浪线** | ❌ | 无 `TextMarkerService` / squiggly |

> **「程序错误检查」的真面目**：`View/ErrorCheckWindow.xaml.cs:31-33`，三条正则配平 IF：
> ```csharp
> new Regex("^ *IF",       RegexOptions.IgnoreCase | RegexOptions.Multiline);
> new Regex("^ *END *IF",  RegexOptions.IgnoreCase | RegexOptions.Multiline);
> new Regex("END *IF",     RegexOptions.IgnoreCase | RegexOptions.Multiline);
> ```
> 结果写进一个 TextBox。**完全不经过 FglParser。**

### 4.6 语法高亮

xshd 正则驱动，与 parser 无关。`CodeTextEditor.cs:1760/1764`：

```csharp
"pack://application:,,,/CodeEditWindow;component/FGL_Mode_Default.xshd"     // 默认主题
"pack://application:,,,/CodeEditWindow;component/FGL_Mode_DarkTheme.xshd"  // 深色主题
```

⚠️ 磁盘文件名是**小写** `CodeEditWindow/fgl_mode_default.xshd` / `fgl_mode_darktheme.xshd`，
但 pack 引用是**大写** `FGL_Mode_*.xshd`。**改名必须同步这两处。** 🔵
每个 xshd 有 **426** 个 `<Word>`。🔵

### 4.7 FGL 语法分析器（`CodeEditor.FglAnalysis/`，12 cs）

```
string 源码
  → FglReader   逐字符游标（NextChar/PreviousChar/LookBack）        79 行
  → FglScanner  手写字符级状态机 NextToken()，68 个 case "关键字"   530 行
  → FglToken    {Type, Text}
  → FglParser   预测式（1 token 前瞻）手写递归下降                 1121 行
  → AST         **通用树节点**（不是每种语法一个类）
       Nodes.NodeKind: int 判型；叶子是 FglTokenNode（带 Children/Parent）
```

- `TokenType` 枚举 **86** 个成员。🔵
- **真实拼写错误（反编译保留原样）**：`TokenType.FOR_TOEKN`（应为 FOR_TOKEN）、
  `TokenType.PEROID_TOKEN`（**实际代表逗号 `,`**，见 `FglScanner.cs:350`）、
  `Nodes.controlBlcok`（应为 controlBlock）。
- 节点种类 7 个：`Unknow=-1 / program=0 / function=1 / interact=3 / controlBlcok=4 / flowControl=5 / end=99`（**2 是缺号**）。
- **两个死代码**：`ParseFlowControl()`（`:191`，从未被 `Parse()` 调用 → FOR/WHILE/FOREACH 的
  flowControl 节点实际永不产生）和 `ParseErrorCheck()`（`:1088`，**方法体是空的**）。

**它唯一的两个用途**（全仓 grep 只有这两条调用链）：

1. **函数结构树 / 大纲**：`Infrastructure/.../TreeNodeFactory.Parse()` → 由
   `ProgramInformation.RefreshTreeNodes()`（`:952`）在 BackgroundWorker 里调用；
2. **自订函数可编辑区行号计算**：`CodeEditorManager.ProcessEditableArea()` `:1230-1253`
   → `CodeEditWindow/Helper/FglParserQuickHelper.ParseFunction()`。

> 所以：**它不是给高亮用的，也不是给错误检查用的。** 想要折叠 / 准确错误检查 / 括号匹配，
> 得在这个 parser 上继续做（`ParseFlowControl` 已经写了一半躺在那里）。

### 4.8 差异引擎（`DifferenceEngine/`，14 cs）

- 算法是 **「最长匹配块(LMS) + 递归锚定」**（不是 LCS-DP，也不是 Myers）：
  `DiffEngine.ProcessDiff` → `ProcessRange`（递归）→ `GetLongestSourceMatch`（暴力扫描最长匹配段）
  → `DiffState/DiffStateList` 缓存 → `DiffReport()` 转 `DiffResultSpan`。
- 三档精度 `DiffEngineLevel`：`FastImperfect`（滚动条色块）/ `Medium` / `SlowPerfect`（正式比对）。
- 4 个输入适配器都实现 `IDiffList`，只是切分粒度不同：
  `DiffList_TextFile`（按 `\n` 切行）/ `DiffList_TextContent` / `DiffList_CharData` / `DiffList_BinaryFile`。
- ⚠️ `TextLine.CompareTo` 是**用 `GetHashCode()` 比较的**（`TextLine.cs:12`）——
  对比对结果精确性有影响，慎改。
- 差异模式下文档里会插入 **`\a`（BEL, 0x07）** 标记；保存时按模式剥离（见 §3.5）。

### 4.9 撤销重做（`UndoRedoFramework/`，9 cs）

经典 Command 模式 + 双栈。接口 `IUndoRedoCommand`（`Undo`/`Execute`/`Clear`）。
栈上限在 **Manager** 里（`UndoRedoManager.cs:17/20/22/144` 的 `MAXSTACKSIZE`，`-1` = 无限；
用户可在偏好里配 `UndoTimes`）。**没有「相邻同类命令自动合并」**，替代品是**命令分组**：
`StartGroup(complexCommand)` → 组内 `AddUndo` 都 `Append` 进组 → `EndGroup()` 一次性压栈。
每次 `AddUndo` 先 `_RedoStack.Clear()`。重绘靠 `event UndoStackChanged`。

### 4.10 搜索（`SpecDesigner.Search/`，11 cs）

范围 = **当前文档 / 所有已打开文档的编辑器文本**（不是全工程检索）。
`SearchBoxControl` + `SearchBoxViewModel`。支持 `IsEditableOnly` / `IsSelectedOnly` 过滤。

---

## 5. 陷阱清单

### 5.1 我在 105 个真实包上踩过的 5 个坑

| # | 我原来的（源码推断出来的）想法 | 实测真相 |
|---|---|---|
| 1 | 「TGL 的占位符必须在 TAP 里有同名 `<point>`，否则是错误」 | **反了**。TGL 预留插入点远多于 TAP 点（113 vs 44 / 全量 ~2.6 万处），`ProcessAddPoints` 会 `Initial()` 造空点。当成 error 会导致 **105/105 个包全部被拒绝写入** |
| 2 | 「`.4gl` = TGL + 展开点的渲染结果」 | **不是**。82/105 个包的 `.4gl` 与渲染结果不同：它由服务器产出，含 TGL 里没有的内容（填好的 `PR版次`、`#此ctrlp無內容` 注记）。→ **绝不能用渲染结果覆盖它** |
| 3 | 「TAP 根元素是 `<tap>`」 | 真名是 **`<add_points>`**，还带 `<?xml ... standalone="no"?>` 声明。硬编码会导致 `section_flag` 设不上、`.tap2` 产出畸形 XML |
| 4 | 「区段正文里含 `DIALOG/REPORT` 开头的行 = 展开的自订点」 | 宽正则会命中框架自己的 `DIALOG ATTRIBUTES(UNBUFFERED,FIELD ORDER FORM)`（180 处误报）。精确判据（拿 TAP 已知函数名搜）后：2426 个区段 **0 处真命中** |
| 5 | 「区段里的 `{<point name=X/>}` 必须在 TAP 里有对应 `<point>`」 | 不需要。`aapp131.main` 里的 `main.define_customerization` 就没有 —— 就是坑 #1 的正常情形 |

另外一条：**`function.*` 前缀不保证内容是函数块**。`s_axmt500(s).tzc` 的
`function.memo_industry` 内容是纯注释 `#240401-00041#1 ...`。
按 `AddPointModel.Content` setter 的逻辑，设计器打开这个包**应该**抛 `ComplexException`
（未真机验证）。

### 5.2 ⚠️ 一条容易误判的：`.4gl` 与「合成结果」不一致**不是** bug

对**服务器下载的原始包**，`.4gl` ⊋ 「TGL + 展开点」。例如：

```
包内 .4gl : #+ Standard Version.....: SD版次:0066(...), PR版次:0066(2024-01-31 13:45:04)
TGL       : #+ Standard Version.....: SD版次:0066(...), PR版次:
包内 .4gl : #此ctrlp無內容         ON ACTION controlp INFIELD apcadocdt
TGL       :          ON ACTION controlp INFIELD apcadocdt
```

TGL 头两行自己就写着 `#該程式未解開Section, 採用最新樣板產出!` ——
**TGL 是「樣板」，`.4gl` 是 build 产物**。

### 5.3 其它容易踩的

- `Packing` **只遍历 zip 里已有的条目** → 原包没有 `.4gl` 就永远不新增。
- `.tap` 的换行是**混合**的（元素间 CRLF、CDATA 内 LF）。
- `ver` 条目**没有扩展名**，靠名字取；加了目录前缀就找不到。
- `status` 未变的点是 `" "`（空格）不是 `""`；`ToXML` 对 `status==CREATE` 的点返回 null。
- TGL 用 `type="s|c|m"`，TAP 用 `src="s|c|m"` —— 不同名。
- 改 `SpecDesignerCommon.PackageKey` / `Infrastructure` 的公开形状会**在运行时**打断编辑器（§0.2）。
- 程序里的**局部变量名/参数名丢失**（`text`/`num`/`arg0`），是反编译损失，别试图"还原"。

---

## 6. 编译与运行

1. **16 个 `.csproj` 全是 `<TargetFrameworkVersion>v4.0`**。现代 VS 没有 4.0 目标包 →
   **批量改成 `v4.8`**（原程序是 AnyCPU 纯 IL、无签名、无强名称，跑 4.8 没问题）。
2. VS 需装「**.NET 桌面开发**」工作负载（提供 4.8 目标包）。
3. 所有第三方 DLL 走 `HintPath` → `D:\APPS\T100设计器_1.0.0.251_免安装\`，**不还原 NuGet**。
   该目录被移动 → 需同步改 `HintPath`。**别删它**（里面有 Sparkle 自动更新组件，可能覆盖改动，改原程序前整目录备份）。
4. WPF 的 `Main` 入口由 `SpecDesignerApp.xaml` 编译期生成（见 §2 末尾）。
5. BAML 还原的 XAML 偶有问题：编译报错先查 `.xaml` ↔ `.xaml.cs` 配对和 `x:Class`。
6. 界面文案在 `.resx`（`Strings.xxx` / `{StaticResource menu_FileOpen}`），**不是 XAML 字面量**。
7. dnSpy 导航技巧：源码里每个类型/方法上方的 `// Token: 0x0200002B RID: 43` 是 .NET 元数据令牌，
   在 dnSpy 打开原始 `T100Designer.exe` → 输入该令牌 → 精确跳到同一类型。

---

## 7. 验证状态

| 结论 | 标记 |
|---|---|
| 章节 3 的容器/分派/合成/保存/打包管线，文件行号 | 🟢 反编译源码直证 |
| `ContentType` 位标志、`HasReadOnlyDocumentProvider` API、`SegmentObject`/`SectionObject`/`EditObject`、xshd 类型、Bookmark/Completion 类型、AvalonEdit 的引用列表、`SegmentObject.Key` 类型 | 🟡 对 `ICSharpCode.AvalonEdit.dll` / `SpecDesigner.Controls.dll` 反射确证 |
| 文件/行数统计、426 个 xshd 关键字、无折叠/括号匹配、`unpack→pack` 105/105 逐条目 sha256 一致、`check` 0 error、`.4gl` 82/105 不一致、TGL 113 vs TAP 44、孤儿点 76/105 | 🔵 实测确证 |
| `FoldingManager` 0 命中、`SectionProvider` 67 处调用 0 处定义、闭源类型 0 定义 | 🔵 grep 确证 |
| §4.8 DifferenceEngine 的开源出处（ClassProject "A Generic, Reusable Diff Algorithm in C# - II"） | ⚪ 工程内无署名，无法证实 |
| `.tap2` 的重算逻辑（生产包里一个 `.tap2` 都没有，只有合成样本验证过） | ⚪ |
| `.tzs` 表单包 / `.tzx` 差异包 / `cite_std="Y"` 引用标准程序的包 | ⚪ **没有样本，未验证** |
| `function.memo_industry` 是否真的会让设计器抛异常 | ⚪ 未真机验证 |
| 服务器端 `adzp050/adzp080/adzp201/adzp064` 对 `.tap`/`.tgl`/`.tap2` 的实际消费方式 | ⚪ 在服务器上，不在本仓库 |

**未做的事**：把改完的包拿**真实设计器打开**验证 —— 这一步只能在装了客户端的机器上做。

**怎么自己复核本文的技术断言**（`🟡` 类结论的复现方法）：

```powershell
# 反射一个发行版 DLL，查类型是否存在、成员签名、ContentType 位标志
$asm = [System.Reflection.Assembly]::LoadFrom("D:\APPS\T100设计器_1.0.0.251_免安装\ICSharpCode.AvalonEdit.dll")
try { $t = $asm.GetTypes() } catch { $t = $_.Exception.Types | Where-Object {$_} }   # 依赖缺失时用异常里的部分结果
$t | Where-Object { $_.Name -match 'Section|Segment|Editable' } | Select-Object FullName
$t | Where-Object { $_.FullName -eq 'ICSharpCode.AvalonEdit.ContentType' } |
     ForEach-Object { $_.GetFields() | Where-Object IsLiteral | ForEach-Object { $_.Name + '=' + $_.GetRawConstantValue() } }
$asm.GetReferencedAssemblies() | Select-Object Name, Version    # 看依赖方向（§0.2）
```

---

> 本工程为第三方商业软件（厂商标识 DSC）的**反编译还原产物**，
> 仅用于个人学习、排障与接口对接研究；请勿用于重制发布或绕过授权。

