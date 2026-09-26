# internal/dev/tapfile — `.tap` 的字节保真改写

`.tap` 是包里**唯一被服务端消费**的设计文件（`.4gl` 是服务器 build 产物、`.tgl` 是框架骨架）。
本包负责它的解析与改写，方式只有一种：**CDATA 感知的字节级扫描**。

**红线 R3：禁止用 XML 库把 `.tap` 整体序列化。** 解析成 DOM 再序列化会重排属性、改写引号、
归一空白 —— 对设计器来说那是另一份文件。

## 形状

```go
doc, err := tapfile.Parse(raw)                 // 只读视图：元素、属性、CDATA 区间
newRaw, err := tapfile.Rewrite(orig, ops...)   // 产出新字节；除改写点外逐字节不动
```

`Op` 是一组显式的改写动作，**没有"改任意元素"的通用口子**：

| Op | 作用 |
|---|---|
| `SetRootAttr{Key, Value}` | 根元素属性（例如区段解锁要写的标志） |
| `SetPointAttr{Name, Key, Value}` | 某个点的属性（`status` / `new` / `order`…） |
| `SetPointCDATA{...}` | 某个点的正文（自订定义点的函数体就在这里） |
| `AddPoint{...}` | 新增点（新增自订定义点走它） |
| `MarkDeleted{Name}` | 把点标成删除态（**保留原正文**） |
| `RemoveTombstones{Name}` | 清掉同名的旧墓碑 |
| `RenamePoint{From, To}` | 改名事务用 |
| `SetSectionCDATA{...}` / `SetSectionAttr{ID, Key, Value}` | 区段正文与属性 |

## 保真规则

- **属性改动"有则改、无则加"**：属性顺序、空白、引号风格全部原样保留。
- **除显式改写点之外，其余字节逐字节不动** —— 这是"不破坏"的物理基础之一。
- `.tap` 的换行是**混合**的：元素之间是 CRLF，CDATA 内部是 LF。写入时按原文各自的形态处理，
  不做全局归一。

## 判据

```bash
go test ./internal/dev/tapfile -count=1     # 约 48 秒，跑真实语料
# TestIdentityRewriteIsByteExact    恒等改写逐字节一致
# TestParseAllCorpusTaps            语料里的 .tap 全部可解析
# TestCDATAMatchesIndependentRegex  与一条独立正则对拍，确认 CDATA 区间找得对
# TestAttrSetIfPresentElseAdd       "有则改、无则加" 的属性语义
# TestRewriteSelfClosingPointExpands 自闭合元素的展开
```

这几条会打印**当前的语料规模**（`.tap` 个数、`<point>` / `<section>` / 带 CDATA 元素数与
恒等写次数）。实测值随语料增删而变，**别把它当判据** —— 判据是"恒等改写逐字节一致"。

## 细节去哪

- 谁产出这些 Op（回写拆分） → [../split/README.md](../split/README.md)
- 权限链与合成（哪些点可编辑） → [../synth/README.md](../synth/README.md)
- 三方依据 → [`docs/T100设计器-README.md`](../../../docs/T100设计器-README.md) §3.3、§3.6、§3.9
