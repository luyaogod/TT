# internal/drawio — 原型图能力（形状库 + 规格展开）

把 T100 的画面画成 drawio 图纸。本层管两件事：**形状库的合成**，与**排版规格 spec → .drawio**。

整套是从 `D:\我的项目\T100Drawio`（原独立的零依赖 Node 项目）搬进来的。搬的原则是
**换语言不换产物**：形状源与合成规则逐条照搬，交给 drawio 的库文件与原实现逐字节相同。

## 职责边界

**做什么**

- 形状源的解析、校验、合成：`shapes/` → catalog（带文字槽位与拉伸元数据）
- catalog → drawio 能加载的 `<mxlibrary>`
- catalog → 给 AI 看的紧凑形状清单（markdown）
- 表格的三层生成（容器 → 行 → 列，用 drawio 原生的 `tableLayout`）
- **排版规格 spec → `.drawio`**（`compose.go` / `spec.go`）

**不做什么**（边界比职责更值得写）

- **不落盘、不碰配置、不碰网络。** 要写文件的是命令层（`internal/cli/drawio`），落点也由它决定。
- **不依赖引擎。** 低保真原型图不需要设计器程序集、不需要工作区、不需要活引擎。
- **不解析输入的 XML。** 形状源是 JSON；本层唯一读的 XML 是 `shapes/raw/` 里手写的那一个片段。

## 形状源

`shapes/` 整棵被 `go:embed` 进去（见 [embed.go](./embed.go)）：

```
shapes/
├─ library.json        库级元信息：目录名 → 库名 / 版本 / 输出文件名
├─ presets.json        部件样式预设（kind → drawio 样式串）
├─ controls/*.json     18 个表单控件（id 与设计器对齐）
├─ business/*.json     9 个业务组件（模块 / 单据 / 接口 / 外部系统）
├─ icons/*.svg         图标，合成时内联成 data URI
└─ raw/*.xml           复杂组件的原始 mxGraphModel（1 个）
```

**任何生成物都不落库、不入库**：catalog、mxlibrary、形状清单全是从上面这些源派生的，
按 `docs/AGENTS.md` 的"一个事实只有一个家"，它们是第二个家，一定会漂。运行期在内存里合成
（27 个形状，微秒级）。代价是**加形状要改源码重编 tt** —— 这是有意的：形状源是代码级数据，
不是给用户编辑的文档。

## 组件的三种写法

`style`（单图元）/ `parts`（组合：外层 group + 每个部件一个子单元格）/ `xmlFile`（原始 XML），
三者互斥，`validateShape` 会拦。校验里最值钱的两条：

- **文字槽位撞车**：两个部件声明同一个 `slot` 时，后者会静默盖掉前者，用的人还找不到原因。
- **id 重复**：两个形状抢同一个 id，图形面板里只看得见一个。

## 转义（改这里之前先读）

drawio 加载一个库走的是 `mxUtils.parseXml(data)` → `JSON.parse(mxUtils.getTextContent(root))`
—— **先当 XML 解析，再取文本当 JSON**。所以库文件是两层转义，顺序不能颠倒：

1. `MarshalJSON` 把 `"` 写成 `\"`（见 [jsonx.go](./jsonx.go)）
2. `MxLibrary` 把整段 JSON 当 XML 文本节点转义，只转 `& < >`（见 [xml.go](./xml.go)）

**不要换成 `encoding/xml` 的 `EscapeText`**：它会把 `"` 写成 `&#34;`、还会动制表符与换行。
原项目为这两层踩过两次坑，报错是 drawio 那句
`Unexpected token 'T', "This page" ... is not valid JSON`。

同理，`MarshalJSON` 关掉了 Go 默认的 HTML 转义（`SetEscapeHTML(false)`）—— 开着的话，
形状源里那些 XML 片段会被转成一堆 `&#x3c;`。

## spec 与 compose

**写 spec 的人（多半是 AI）只决定「用哪个控件、放哪、写什么字」**，其余全是机械活。
这样它不碰 XML，就不会写坏转义、不会撞 id；库改了也不用重新教它。

展开一份 spec 分五步，顺序都不能换（`placeItem` 的注释里有完整理由）：

| 步 | 做什么 | 为什么这个顺序 |
|---|---|---|
| ① 多页签 | `pages`/`active` → 页签条画全、当前页白底 | 当前页是**原地替换**库里的页签格（保留原 id），兄弟页才是新增 |
| ② 换文字 | 按控件的文字槽位替换 | **必须在加 id 前缀之前** —— 槽位里存的是原始单元格 id |
| ③ 重编号 | 剥掉库条目自带的 0/1 图层根格，其余 id 加 `i{序号}-` 前缀 | 不加前缀，两个同名控件会在同一张图上撞 id |
| ④ 拉伸 | 三类子格行为（见下） | 要在定位外层之前做 —— 几何得先按原尺寸算完 |
| ⑤ 定位 | 给外层格补 x/y/w/h | 外层格的 geometry **只写 width/height**，就是留给这一步的 |

三类子格由形状源里 `pfill` / `pdeco` 标出（值是**1 基**的部件序号）：

- **fill** —— 容器框体，锚定 x/y 拉伸到填满外层（减去另三边的内缩）。边框跟着动。
- **deco** —— 装饰件（页签、图例、树节点），一个像素都不动，**不被拉变形**。
- **其余** —— 等比缩放。字段行、表格内容随尺寸走。

## 判据

```bash
go test ./internal/drawio -count=1
```

期望 `ok`。每条测试钉住什么：

| 测试 | 钉住 |
|---|---|
| `TestComposeMatchesNodeGolden` | 一份**真实的**全屏规格（67 控件 / 881 单元格）与**原 Node 实现**的产物**逐字节**相同 |
| `TestComposeSimpleGrid` | 文档里那份两列栅格样例：9 个控件、列宽报告、产物良构 |
| `TestComposeStretchClasses` | 三类子格：fill 撑开、deco 不动、外层尺寸正确 |
| `TestComposeDynamicTable` | 表格尺寸由列宽与行数**算出来**，表头文字来自 spec |
| `TestComposeMultiTab` | 当前页保留原 id、兄弟页签带序号与灰底 |
| `TestComposeErrors` | 七条输入错的报错里带着可用 id / 可用槽位 |
| `TestSelfCheckCatches` | 自查真的会响（id 重复 / parent 悬空 / NaN），且**不误报** `grid="1"` 这类子串 |
| `TestPortEquivalence` | Go 合成结果与**原 Node 版产出的库**逐条相同（形状数、几何、样式串、XML 全文） |
| `TestSynthShapeCounts` | controls 18 + business 9 —— 防 embed 少文件导致的**静默掉件** |
| `TestLibraryEntriesCarryWhatDrawioNeeds` | 每条都带 `xml/w/h/title/aspect`，且 `xml` 非空 |

对照基线在 `testdata/drawio/`：库文件与那份 `.drawio` 都**由原 Node 实现产出**，
搬进 TT 之后不再重新生成，所以能钉住"移植没有改变交给 drawio 的东西"。
库比的是**解码后的语义**，`.drawio` 比的是**逐字节**（那条链没有时钟也没有随机源）。

**故意不覆盖**：图"好不好看"、在 drawio 里能不能拖 —— 那是人打开 drawio 看一眼的事，
本层证不了。`TestSynthShapeCounts` 是**数数**，证明不了"每个形状的部件都画对了"。

## 细节去哪

- 命令面（开关、落点、退出码） → [`../cli/drawio/README.md`](../cli/drawio/README.md)
- 库文件为什么长这样、形状源的三种写法 → 本目录各文件的包注释；`.gitattributes` 与生成物无关（生成物不落库）
