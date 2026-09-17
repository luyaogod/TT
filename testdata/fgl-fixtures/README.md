# FGL 夹具（61 组）— 来源与用法

本目录是 **BDL** 项目的 `test/fixtures/`（61 个 `<name>.4gl` + 61 个 `<name>.expected.json`）的副本，
经 `D:\我的项目\TDebug\web\scripts\fgl-fixtures\` 原样转来。

- **来源**：`D:\我的项目\BDL`（VS Code 扩展，**MIT License**，同作者）。
- **用途**：给 `internal/fgl`（移植自 BDL `src/extension.ts` 的 4GL 大纲解析器）做一致性对拍，
  由 `internal/fgl/outline_fixtures_test.go` 消费。
- **期望值来源**：由 BDL 官方文档逐条推导（每个 `.expected.json` 的 `doc` / `about` 字段注明出处），
  **不是**某个实现跑出来的快照。
- **基线**：与 BDL 自身 `scripts/fixture-check.cjs` 的实测一致 ——
  `用例 61 个：通过 58，已知偏差 3，不符 0`。
- **`deviation` 字段**：声明「文档上应当如此、实现已知未做到」的偏差。本目录的判据是：
  输出与 `nodes` 不等**且**夹具带 `deviation` → 记为「已知偏差」（不算失败）；
  不等且没有 `deviation` → **失败**。
  当前 3 个已知偏差：`dialog-subdialog`、`input-by-name-record-star`、`input-nested-semicolon`。
- **改动约定**：**不要**在本目录增删或修改夹具。确实要改期望时，必须先补 `deviation` 说明，
  不许把实现输出回写当期望。

`README.source.md` 是 TDebug 侧对该夹具集的原始说明（保留以便溯源）。

工具的语料回归（`internal/fgl/corpus_test.go`）另用真实 `.tzc`：
166 个包、4,336 个自订点，4,335 个通过信封检查；唯一失败的是
`s_axmt500(s).tzc` 的 `function.memo_industry`（正文只有一行 `#` 注释，
是语料里已知的退化点，tdev 会把它标成只读且拒绝写回）。
