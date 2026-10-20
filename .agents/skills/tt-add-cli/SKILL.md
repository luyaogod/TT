---
name: tt-add-cli
description: 在 TT 中新增或改造 CLI 命令、子命令、参数、JSON 输出、批量语义、投影、错误码与退出码时使用。触发词：加命令、加子命令、agent 友好、JSON 输出、--brief、投影、批量、退出码、CLI 契约、命令组。
version: 0.1.0
---

# 给 TT 加 CLI：Agent 友好契约

这份技能面向**改 TT 的人**，不是 TT 的最终用户。  
目标不是“多一个命令”，而是让后续 Agent 能：

1. 知道该调用哪条命令；
2. 用少量上下文拿到完整答案；
3. 失败时能自己纠正；
4. 永远不会把“成功但少给了字段”当成完整答案。

先读 [docs/DESIGN.md](../../../docs/DESIGN.md) 的整体分层，再读 [internal/cli/README.md](../../../internal/cli/README.md) 的命令树与退出码。

---

## 0. 先回答八个问题

新增命令前，先把下面八件事写清楚。写不清就不要开始写 Go。

1. **消费者要回答什么问题？**  
   不是“提供什么数据”，而是“Agent 拿它做什么判断”。
2. **数据源是什么？**  
   本地文件、镜像、SSH、数据库、Web？默认是否联网？
3. **目标如何点名？**  
   `--at <环境>[@<企业>]`、`--env`、`--dir`、显式文件，还是本地无目标？
4. **默认输出是什么？**  
   数据命令默认 JSON；人类可读必须显式 `--human`。
5. **默认输出是否有界？**  
   多少条、多少层、多少字段；超限如何告知、如何翻页。
6. **调用方能否选择字段/子对象？**  
   有没有 `--brief`、`--col`、`--kw`、`--tables` 这类投影。
7. **批量怎么表达？**  
   逗号分隔、`items[]`、`missing[]`、单项 `found`；单个精确 ID 缺失是否 `NOT_FOUND`。
8. **失败是否可机读？**  
   稳定错误码、退出码、可复制的修复建议；未知子命令是否非 0。

---

## 1. 命令放在哪

### 默认：单命令优先

只有一个稳定动词时，**不要为了“看起来整齐”新建命令组**。

- 能作为现有组的叶子命令 → 放现有组。
- 是一个独立能力、但只有一个动词 → 顶层单命令即可。
- 只有出现**多个相关动词**或**多个独立数据面**时，才值得建组。

反例：把一棵块树拆成 `show` / `pages` / `fields` 三个只换名字的动词。  
如果新增 `outline` 这类单数据面命令，用一个动词，加 `--flat`、`--kind`、`--depth`、`--brief` 做视图即可。

正例：

```text
tt interface show/pages/fields
```

它拆开是因为页面、字段、42s 标签是不同数据面，不是同一份数据的三个摘要。

### 放哪个现有组

- **远程文件只读浏览、下载、镜像** → `tt file`
- **服务器文件的结构化派生视图**（且需要远端读取） → 先评估 `tt file`，不要默认塞进 `tt dev`
- **设计器包/表单包操作** → `tt dev tzc` / `tt dev tzs`
- **字典查询** → `tt dict`
- **数据库连通性/只读 SQL** → `tt db`
- **跨 42f + 42s 的前端画面合成** → `tt interface`
- **纯本地源码结构分析**（如 4GL outline） → 优先顶层单命令，避免 `tt file` 的 SSH 预连接职责

判断标准是**职责**，不是“都读文件所以放一起”。  
`tt file` 的既有语义是服务器文件；如果一个命令主要读本地镜像/显式文件，把它塞进 `file` 往往要绕开它的连接生命周期，反而增加特例。

### 新组必须满足

新建组前确认：

- 至少有 2 个稳定的用户动作，且这些动作共享同一数据模型；
- 组的 `Long` 能说清“边界是什么、不做什么”；
- 组命令有 `RunE`；未知子命令用 `common.UnknownSubcommand`，不能默认退 0；
- 在 [internal/cli/at/scopes.go](../../../internal/cli/at/scopes.go) 认领 `--at` 作用域。

不要建“只有一个动词的组”，也不要把内部包边界（`internal/dev/fgl`）直接暴露成用户命令组。

---

## 2. 输出契约

### 默认 JSON，人类显式 `--human`

TT 的数据命令默认 JSON：

```text
默认 / --json / --format json  → JSON 信封
--human / --format table       → 人读视图
--csv / --format csv           → CSV（仅当命令真的支持）
```

新命令不要默认打印表格。  
`--human` 是**同一份数据的可读形态**，不是省 token 的模式。

### 信封

优先走统一的 `internal/output`：

```text
ok
source / env / sshHost / ent / account / target / dialect / route
totalRows / returned / truncated / offset / localPath / notes
data
```

`ok:false` 使用稳定错误码：

```json
{
  "ok": false,
  "code": "NOT_FOUND",
  "error": "表 xyz 不存在",
  "hint": "列出全部表：tt dict r.t --kw ...",
  "exitCode": 4
}
```

错误必须让调用方能分支：

- 稳定 `code`；
- 稳定 `exitCode`；
- 可复制的 `hint` / `remediation`；
- 不把堆栈、SQL 原文、口令当错误正文。

### 不要静默降级

这是硬规则：

- **`--human` 不能比 JSON 少内容而不告知**；
- 真正的摘要必须叫 `--brief` / `--summary`；
- 批量 `--human` 要么逐条展示详情，要么明确报错，不能只给一张汇总表；
- `--brief` 必须同时作用于 JSON、CSV、表格；不能只在表格里生效；
- 不得把“字段没解析”回退成列名、token 或空字符串冒充中文。

---

## 3. 粒度：投影、分页、截断

### 投影

投影 = 让调用方选“要哪些字段/子对象”，不是先把全量生成出来再截掉。

优先提供**语义明确的开关**，不要只给一个万能 `--brief`：

| 场景 | 开关 |
|---|---|
| 表字段 | `--col apca035,apca036` / `--kw 暂估` |
| 程序用表矩阵 | `--tables none\|names\|full` |
| 分类码值 | `--brief`（只留值→说明） |
| 参数引用清单 | `--doclist none\|count\|full` |
| 画面结构 | `--depth` / `--fields` / `--page` |

规则：

- 精确投影缺字段 → `NOT_FOUND`，不能静默省略；
- 模糊筛选零命中 → 成功空数组，不能报错；
- 被选中的字段/节点保持完整对象，只缩短数组；
- 投影不改变既有 JSON 键名。

### 有界

默认必须有界：

- 列表/搜索有 `--limit`；
- 树有 `--depth`；
- 长文本有字段级截断；
- 超限返回 `totalRows` / `returned` / `truncated` / 查看全量的下一步；
- 大结果先落盘再截断，落盘失败则不静默截断。

不要用“默认打印全部”来省设计，Agent 的上下文比磁盘贵。

---

## 4. 输入与目标

### 环境一律点名

没有默认环境：

```bash
--at <环境>[@<企业>]
--env <环境>
--at-list
```

- 需要企业的命令用 `EnvEnt` / `EnvEntReq`；
- 只连服务器的用 `Env`；
- 本地文件/镜像用 `--dir` 或显式路径；
- `--at` 与位置参数不一致时报错，不二选一猜测。

在 [internal/cli/at/scopes.go](../../../internal/cli/at/scopes.go) 登记作用域；没登记就是 `None`，给了 `--at` 必须报错，不能静默忽略。

### 本地优先

能本地读就不要联网：

```text
显式文件 > --dir 镜像 > --at 对应镜像 > 远程
```

远程命令要明确只读；不要为了“方便”偷偷降级到另一个源，也不要在没有点名环境时猜一个。

---

## 5. 错误与退出码

沿用 [internal/cli/README.md](../../../internal/cli/README.md) 的码表，不新造一套数字。

常见语义：

```text
0  成功
2  用法错 / 目标歧义
4  精确 ID 不存在（NOT_FOUND）
```

具体命令组可以有自己稳定的码，但必须在组 README 和 `--help` 中写清楚。  
未知子命令必须返回非 0，并列出可用子命令。

错误提示必须：

- 指向**当前存在**的命令；
- 给出下一步能直接复制的命令；
- 区分“没查到”“查不了”“查错了地方”；
- 不把 env/ent/账号错误说成数据不存在。

### 不能从不可信文本猜结构

列名、表头、类型不要从定宽文本或截断行里猜。  
要拿结构化元数据；拿不到就明确失败，不能返回看起来正确的错字段。

反例：从 sqlplus 固定宽度表头取列名，`'Y' as abcdef` 得到 `A`，但命令仍 `ok:true`。

---

## 6. 批量语义

多 ID 查询统一形状：

```json
{
  "items": [
    {"id": "a", "found": true, "data": {}},
    {"id": "b", "found": false}
  ],
  "missing": ["b"]
}
```

规则：

- 批量里单个缺失不拖垮整批；
- 单个精确 ID 缺失 → `NOT_FOUND` / exit 4；
- `--kw` / 搜索零命中 → 成功空数组；
- 每个 item 带 `found`，不要只靠数组长度猜；
- 批量 + `--human` 不许静默只给摘要；要摘要就显式 `--brief`。

---

## 7. 帮助与文档

### `--help` 是 Agent 的第一份文档

新命令的 `Long` 至少包含：

- 一句话职责；
- 2–4 条真实示例；
- 默认输出形态；
- 关键 JSON 形状；
- `--brief` / 投影的语义；
- 退出码；
- 本地/远程数据源选择；
- 错误时的下一步。

不要只写“查询数据”这种无法执行的描述。

### 文档更新位置

- 命令组说明 → `internal/cli/<组>/README.md`
- 命令树与退出码 → [internal/cli/README.md](../../../internal/cli/README.md)
- 整体分层/结构 → [docs/DESIGN.md](../../../docs/DESIGN.md)
- 仓库结构 → [AGENTS.md](../../../AGENTS.md)
- 对外用法 → `skills/<命令面>/SKILL.md`（技能由维护者管理；不要擅自改别人的技能）

所有文档只写当前状态；历史留在 git。

---

## 8. 测试与验收

至少覆盖：

- 空输入、缺目标、未知子命令；
- 精确 ID 不存在（code/exit）；
- 搜索无结果（成功空数组）；
- 批量 `items[]` / `missing[]` / `found`；
- `--brief` 在 JSON、CSV、表格三种形态都生效；
- 投影缺字段、投影零命中；
- 截断 `totalRows` / `returned` / `truncated`；
- 人读批量不静默丢详情；
- 不可信列名/表头不会返回错结构；
- 真实样本可选门控，不能把本机路径设为必需。

默认执行：

```bash
go test ./... -count=1
```

改了 `web/` 再加：

```bash
cd web && npm run build
```

手工回读至少跑：

```bash
tt <cmd> --help
tt <cmd> ...
tt <cmd> ... --human
tt <cmd> ... --brief
tt <cmd> ... --at-list
```

看输出字符数，不只看“有没有输出”。

---

## 9. 提交前检查表

- [ ] 命令放在正确的组/层级，没有为单个动词新建组
- [ ] 默认 JSON；`--human` 不丢数据
- [ ] `--brief` 是明确摘要，且所有格式生效
- [ ] 有投影开关，且精确缺字段报 `NOT_FOUND`
- [ ] 默认输出有界，截断不静默
- [ ] 批量有 `items[]` / `missing[]` / `found`
- [ ] 错误有稳定 code / exit / hint
- [ ] `--at` 在 scopes 表登记；没有默认环境
- [ ] 本地/远程数据源不静默切换
- [ ] 帮助里有示例、JSON 形状、退出码
- [ ] 组 README、CLI README、结构文档同步
- [ ] 测试覆盖空/缺失/批量/投影/截断
- [ ] 不提交用户自己的未完成技能或其他会话文件

---

## 10. 反模式清单

不要做：

- 默认打印表格，让 Agent 解析人类排版；
- 用 `--human` 当“省 token”开关，结果只回摘要；
- 只给 `--brief` 的表格实现，JSON 仍是全量；
- 默认输出全部字段/全部行，再要求调用方自己 grep；
- 用行号/数组下标当稳定 ID；
- 把 `page_1` 这类可重复的 textKey 当唯一 ID；
- 从 sqlplus / psql 的定宽文本猜列名；
- 查不到就静默回退到另一个数据源；
- 未知子命令退 0；
- 错误提示指向已删除的命令；
- 把内部实现的 `internal/...` 包名直接当用户命令层级；
- 为“功能丰富”堆多个只换名字的子命令，而没有新增数据面。

---

## 11. 判据

```bash
# 结构、参数、输出与错误契约
go test ./... -count=1

# 至少手工回读一次：
tt <cmd> --help
tt <cmd> --at-list
tt <cmd> <正常输入>
tt <cmd> <正常输入> --human
tt <cmd> <正常输入> --brief
tt <cmd> <不存在的精确 ID>
tt <cmd> --kw <无匹配词>
```

期望：

- 未知子命令非 0；
- 精确不存在是 `NOT_FOUND` / 4，不是成功空；
- 搜索无匹配是成功空；
- `--brief` 不改变被选对象的字段语义；
- 输出大小与问题规模相称；
- 错误提示可直接照抄执行。

---

## 细节去哪

- 命令树、退出码 → [internal/cli/README.md](../../../internal/cli/README.md)
- 全局开关与输出分流 → [internal/cli/common/README.md](../../../internal/cli/common/README.md)
- 输出信封 → [internal/output/README.md](../../../internal/output/README.md)
- 环境/企业作用域 → [internal/cli/at/scopes.go](../../../internal/cli/at/scopes.go)
- 前端画面合成参考 → [internal/cli/interfacecmd/README.md](../../../internal/cli/interfacecmd/README.md)
- 字典投影参考 → [internal/cli/dict/README.md](../../../internal/cli/dict/README.md)
- 4GL 大纲解析器 → [internal/dev/fgl/README.md](../../../internal/dev/fgl/README.md)
