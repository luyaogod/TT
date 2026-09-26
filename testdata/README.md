# testdata — 测试数据

只有一套夹具：[fgl-fixtures](./fgl-fixtures/README.md) —— 4GL 大纲解析器的文档级对拍用例
（61 组 `.4gl` + 期望值）。

**同一套夹具有两份**：本目录是 Go 侧那份，前端侧还有一份在
[../web/app/scripts/fgl-fixtures/](../web/app/scripts/fgl-fixtures/README.md)（它的检查项
跑在 Node 里，从自己的目录读夹具）。**两份内容必须逐字节相同**，改一处要同步另一处。

真实语料（客户的 `.tzc` / `.tzs` 包）**不在仓库里**：它是活目录、是客户数据，
由环境变量指向，见根 [BUILD.md](../BUILD.md) 的两条纪律。

## 细节去哪

- 夹具的来源、判据与改动约定 → [fgl-fixtures/README.md](./fgl-fixtures/README.md)
- 语料回归怎么跑 → 根 [BUILD.md](../BUILD.md)
