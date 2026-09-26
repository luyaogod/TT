# internal/dev/fence — 围栏协议

`Document ⇄ FencedText`：把合成出来的文档渲染成**一个带标注的 4GL 文件**（人/AI 唯一编辑界面），
以及把编辑后的文件解析回来。

**契约：`render` 与 `parse` 互为逆。** 对任意合法包，
`parse(render(synthesize(pkg)))` 必须还原出同一个文档（Region 序列逐项相等）——
这是"不破坏"能被机械证明的前提。

围栏是**单行注释**，不改变 4GL 语义，与框架原生的 `{<point/>}` / `{<section>}` 标记区分开。

## 五种旗标

| 旗标 | 含义 |
|---|---|
| `EDITABLE` | 可编辑的自订定义点 |
| `READONLY` | 只读 |
| `EDITABLE-SEC` | 框架解锁后可编辑的**区段** |
| `APPEND` | 锚点区段：允许在其中追加新的点块 |
| `PLAIN` | 裸名插入点：没有函数身份，整块即正文 |

## 规则

1. `begin` / `end` 严格配对；**允许 `section → point` 一层嵌套**（真实框架里自订点就住在区段内），
   深度 > 2 直接报错。这是与上游设计指南的一处**有意偏差**（D-2）。
2. **围栏行本身属于"围栏外"** —— 不许增删改。锚定部分（点名、旗标、`status`/`src`/`new`/`order`
   等）由工具推导，不由作者填写；`fn` / `scope` / `desc` 是唯一允许改的字段，改了就触发
   [verify](../verify/README.md) 的结构事务校验（V1 / V5 / V6）。
3. 自订定义点内部再分**结构行**（注释头 / 签名行 / `END`）与**正文本体**，只有正文本体可改；
   结构行被改一律拦下。
4. 新增点只能出现在 `APPEND` 锚点区段的围栏块里，且类型要匹配、名字不与基线重名。
5. 删除点只允许删掉本次标记为新增的那些。

解析阶段就把这些违反面分别报出来（改结构行、未配对围栏、锚定部分被改、越权删除/追加），
不留给后面的闸门去猜。

## 入口

| 入口 | 作用 |
|---|---|
| `Render(doc)` | 文档 → 带围栏的字节 + Region 表 + 区间记账 |
| `Parse(base, fenced)` | 编辑后的字节 → `ParseResult`（含新增/删除/变动的区间） |
| `ApplyEdits(doc, edits)` | 按编辑列表产出新的围栏文本 |
| `ParseFenceMeta` / `FenceMeta` / `AnchoredSignature` / `BeginFenceLine` | 围栏行与元数据的读写 |
| `Pair` / `FenceError` | 配对检查与错误 |

## 判据

```bash
go test ./internal/dev/fence -count=1     # 约 48 秒，跑真实语料
# TestFenceRoundTripCorpus            可逆性：语料级 render/parse 往返一致
# TestRenderContainsFencesAndMetadata 渲染出的围栏与元数据
# TestParseRejectsEditedStructureLine 改结构行被拒
# TestParseRejectsUnpairedFence       未配对围栏被拒
# TestParseDetectsDeletion / TestParseDetectsAppendInsideAnchor / TestParseRejectsAppendOutsideAnchor
```

## 细节去哪

- 行尾：受保护字节的 CRLF↔LF 不算改动，判定在 [../model/README.md](../model/README.md)（`EqualEOL`）
- 闸门怎么用解析结果 → [../verify/README.md](../verify/README.md)
- 合成从哪来 → [../synth/README.md](../synth/README.md)
