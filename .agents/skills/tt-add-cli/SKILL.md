---
name: tt-add-cli
description: 在 TT 中新增或改造 CLI 命令时使用。触发词：加命令、命令组、Agent 友好、JSON 输出、--brief、投影、批量、错误码、CLI 契约。
version: 0.1.1
---

# 给 TT 加 CLI：只做 Agent 真正需要的命令

TT 的 CLI 不是数据导出器，而是 Agent 的查询 API。

每条命令只回答一个明确问题，并返回：

- 足够回答问题的内容；
- 可继续裁剪的粒度；
- 失败时能自己纠正的线索。

---

## 一、设计四问

动手前先回答：

1. **Agent 要用它回答什么问题？**
2. **数据从哪里来，默认是否联网？**
3. **默认输出多大，调用方怎么缩小？**
4. **查不到、查不了、查错地方，分别返回什么？**

答不清，不要开始写 Go。

---

## 二、三条原则

### 1. 明确

- 环境、企业、文件、目录一律点名，没有默认环境。
- 默认只读；远程写操作必须显式、单独设计。
- 失败不能伪装成成功：未知子命令非 0，精确 ID 缺失是 `NOT_FOUND`。
- 不静默切换数据源，不猜企业，不猜列名。

### 2. 可控

默认输出必须有界：

```text
列表/搜索：--limit
树：--depth
长文本：字段级截断
大结果：先落盘，再截断
```

超限必须返回：

```text
totalRows / returned / truncated / 查看全量的下一步
```

投影不是截断：

```text
r.t   --col apca035,apca036 / --kw 暂估
prog  --tables none|names|full
scc   --brief
docp  --doclist none|count|full
```

被选中的字段/节点保持完整，只缩短数组。

### 3. 可信

- `--human` 只改变展示形态，不能比 JSON 少内容。
- 摘要必须叫 `--brief`，不能靠 `--human` 偷偷降级。
- `--brief` 在 JSON、CSV、表格三种形态都生效。
- 不从定宽文本、截断表头、猜测 token 里取结构；拿不到可信元数据就明确失败。
- 批量中单个缺失不拖垮整批；单项精确缺失仍报 `NOT_FOUND`。

---

## 三、命令放在哪

**单动词不要新建组。**

- 一个稳定动作：放现有组，或顶层单命令。
- 只有多个**独立数据面**才建组；不是多个视图就建组。

```text
tt interface show/pages/fields   # 页面/字段/标签是不同数据面 → 组
tt outline <file>                # 只有一棵块树 → 单命令 + flags
```

`outline` 的典型 flags：

```text
--flat     扁平列表
--kind     按块类型过滤
--depth    控制层级
--brief    精简节点字段
--human    人读树
```

不要把“都读文件”当成同组理由。`tt file` 是服务器文件线；纯本地源码分析优先顶层，避免绕开它的 SSH 生命周期。

新组必须：

- 有至少两个稳定动作；
- 在 `internal/cli/at/scopes.go` 认领 `--at` 作用域；
- 有 `RunE`，未知子命令走 `common.UnknownSubcommand`。

---

## 四、输出契约

数据命令默认 JSON：

```text
默认 / --json / --format json   → JSON
--human / --format table        → 同一份数据的人读形态
--brief                         → 明确的摘要/投影
```

优先使用统一信封：

```text
ok
source / env / ent / account / target
totalRows / returned / truncated / localPath / notes
data
```

错误用稳定结构：

```json
{
  "ok": false,
  "code": "NOT_FOUND",
  "error": "表 xyz 不存在",
  "hint": "列出全部表：tt dict r.t --kw ...",
  "exitCode": 4
}
```

`hint` 必须指向当前存在的命令，并给出下一步；不要只写“参数错误”。

---

## 五、批量与精确查询

多 ID 查询统一：

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

- 批量单个缺失 → 进 `missing[]`，整批仍成功；
- 单个精确 ID 缺失 → `NOT_FOUND` / exit 4；
- `--kw` 搜索无结果 → 成功空数组；
- 每项带 `found`，不要靠数组长度猜。

---

## 六、帮助、测试、文档

`--help` 至少写：

- 2–4 条可复制示例；
- 默认输出形态；
- 关键 JSON 形状；
- `--brief` 与投影语义；
- 退出码；
- 数据源选择。

测试至少覆盖：

```text
空输入 / 缺目标 / 未知子命令
精确缺失 / 搜索空结果
batch items + missing + found
--brief 在 JSON/CSV/table 都生效
--human 不丢详情
截断元数据
```

默认验证：

```bash
go test ./... -count=1
```

手工回读：

```bash
tt <cmd> --help
tt <cmd> --at-list
tt <cmd> ...
tt <cmd> ... --human
tt <cmd> ... --brief
tt <cmd> <不存在 ID>
tt <cmd> --kw <无匹配词>
```

文档更新：

```text
命令组说明       internal/cli/<组>/README.md
命令树与退出码   internal/cli/README.md
整体分层         docs/DESIGN.md
仓库结构         AGENTS.md
```

---

## 七、提交前检查

- [ ] 单动词没有新建组
- [ ] 默认 JSON，`--human` 不丢数据
- [ ] `--brief` 是显式摘要，所有格式生效
- [ ] 有字段/子对象投影，精确缺字段报 `NOT_FOUND`
- [ ] 默认有界，截断不静默
- [ ] 批量有 `items[]` / `missing[]` / `found`
- [ ] 错误有稳定 code / exit / hint
- [ ] `--at` 已登记 scopes，没有默认环境
- [ ] 本地/远程不静默切换
- [ ] 帮助、README、测试同步
- [ ] 不提交他人未完成的技能或其他会话文件

---

## 八、反模式

- 默认打印表格。
- 用 `--human` 省 token，结果只回摘要。
- 只给表格版 `--brief`，JSON 仍是全量。
- 默认输出全部字段/全部行。
- 用行号、数组下标、可重复 textKey 当唯一 ID。
- 从 sqlplus/psql 定宽表头猜列名。
- 查不到就静默换数据源。
- 未知子命令退 0。
- 为“功能丰富”堆只换名字的子命令。

---

## 参考

- [docs/DESIGN.md](../../../docs/DESIGN.md)
- [internal/cli/README.md](../../../internal/cli/README.md)
- [internal/cli/common/README.md](../../../internal/cli/common/README.md)
- [internal/output/README.md](../../../internal/output/README.md)
- [internal/cli/at/scopes.go](../../../internal/cli/at/scopes.go)
- [internal/cli/interfacecmd/README.md](../../../internal/cli/interfacecmd/README.md)
- [internal/cli/dict/README.md](../../../internal/cli/dict/README.md)
