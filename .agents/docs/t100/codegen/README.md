# T100 设计 → 代码生成 · 研究语料

本目录收集 T100（鼎捷 ERP）**设计器规格 → 程序代码** 这条链路上的核心代码、模板与研究结论。
只放「生成器基础设施」，**不含业务程序**（如 axmt500_wf、aqct300_wf 这类 4GL 业务代码不在此列）。

## 目录

| 目录 | 内容 | 说明 |
|---|---|---|
| `01-server/` | 服务器端生成器（4GL） | 上传编排、蓝图、产生器、tgl 整理 |
| `02-launcher/` | `r.*` 命令脚本 | 设计器实际调用的入口（r.r / r.c3 / r.c / r.f / r.l …） |
| `03-templates/slice/` | 118 个段落切片模板 | `code_a00..a77 / d01..d10 / qs01..qs27 / ss01..ss04` |
| `04-designer/` | 设计器反编译源码（子集） | 表关联、控件、加点模型等关键类 |
| `diagrams/` | 4 张 HTML+SVG 图解 | 流水线 / 关系→CRUD / FK vs upper / aqct300 案例 |

主文档：**[ARCHITECTURE.md](./ARCHITECTURE.md)** —— 从设计到代码生成的完整架构说明。

## 来源与快照

| 来源 | 路径 | 快照时间 |
|---|---|---|
| 服务器源码镜像（测试环境） | `D:\APPS\TT Desktop\mirror\测试环境\erp\adz\4gl` | 2026-09-03 |
| 设计器反编译源码 | `docs\t100\designer-src` | — |
| `r.*` 启动脚本 | `docs\t100\r.t\launcher` | — |
| 切片模板 | `D:\t100_wrok_dir\fulian\172prd\slice`（客户工作区镜像） | 2026-11-20 |

逐文件清单（大小 + sha256 前 12 位 + 来源）见 [MANIFEST.tsv](./MANIFEST.tsv)。

## 未收录（有意排除）

- **业务程序**：`erp/*/4gl/*.4gl`（除 adz 模块的生成器）
- **整支样例母版** `$ERP/mdl/code_<type>.template`（i01/t01/…）：服务器未镜像到本地，暂缺，见 ARCHITECTURE「未解问题」
- **`azzp188`（823 KB，规格上传时的程序关联作业）**：属于流程配套而非生成核心
- **`adzp050`（569 KB，签出/下载）**：属于版本与下载通道，与生成链弱相关

需要时可从客户环境镜像按 MANIFEST 的同名路径补齐。

## 怎么读

1. `ARCHITECTURE.md` 第 2 节（全景流水线）——先建立四段式印象
2. `diagrams/pipeline-overview.html`——把流水线和文件谱系看一遍
3. `02-launcher/tools/r.c3`、`r.c`、`r.f`、`r.l`——四个编译入口到底调了什么
4. `01-server/orchestration/adzp080.4gl` + `adzp040.4gl`——上传与产生编排
5. `01-server/blueprint/sadzp030_tab*.4gl`——判型与关系求解（.tab 从哪来）
6. `01-server/generator/adzp150.4gl` + `sadzp150_*`——模板展开（.tgl 从哪来）
7. `04-designer/Relation/*.cs`——设计器里的表关联/三层二显如何写进规格

## 约定

- 文件名保持原样（与服务器一致），便于回查
- 本目录是**只读参考语料**，不参与任何构建
- 代码行号引用均相对本目录内文件；`file:line` 可直接跳转
