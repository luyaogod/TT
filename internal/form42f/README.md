# internal/form42f — `.42f` 表单文件解析（纯函数，零 IO）

只做一件事：把编译后的 `.42f`（明文 XML）解析成**字段表** —— 每个字段的
`colName` / `fieldId` / 控件类型 / 几何 / 必填等，供 `tt form` / `tt dict fieldtab` /
`tt drawio` / `tt dev tzs` 共用。

```go
f, err := form42f.ParseForm42F("axmt500_wf", xmlText)
f.ByColName[form42f.ColName("sfajwf012")] // → fieldId 等
```

## 三条硬约束（写死在代码里）

1. **字段载体不止 `<FormField>`** —— 明细网格的列是 `<TableColumn>`，还有
   `<PhantomColumn>` / `<PhantomFormField>`。只收 FormField 会得到"看起来对、其实错"的结果
   （实测 `axmt500_wf.42f`：563 条 = TableColumn 397 + FormField 132 + PhantomFormField 22 +
   PhantomColumn 12）。由 `TestParseForm42F_CollectsTableColumns`（反证测试）守着。
2. **绝不用 `.4fd` 的 `fieldId`** —— 设计期编号每次保存全量重排（`engine/SPEC.md:257`）。
   本文件解析的是**编译后**的 `.42f`，它的 `fieldId` 才稳定（表单变体内；带作用域后 100%）。
3. **本解析器不做业务语义** —— 不查字典、不猜中文名；`cmt_*` / `lbl_*` 只是原样带出来。

## 契约与判据

- 契约（唯一口径）：`.agents/plan/tt-genero-formdef-contract.md` 的 C1 / C2 / C7。
- 上游事实：`.42f` 是明文 XML（`GDC-逆向报告.md` §17）；真 T100 表单同样（§18）；
  离线解析与运行期语料逐条比对 **509/509 = 100%**。
- 判据：`go test ./internal/form42f -count=1`（合成夹具在 `testdata/form42f/`）。

## 边界

本包**不许** import 命令层或任何带 IO 的包（`internal/host` / `internal/cli/*`）——
"取 `.42f` 的字节"是调用方的事（见 `internal/host/formsource.go`）。
