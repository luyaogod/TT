# internal/dev/cli — `tt dev` 的命令行入口

`tt dev` 的**完整 CLI**：`.tzc` 的八个动词、`tt dev tzs` 的转发、`tt dev install`。
它不走 cobra 树 —— 根命令那层只做参数原样转发与退出码透传（见
[../../cli/dev/README.md](../../cli/dev/README.md)）。

## 分发

`Run(args)` 是两级手写分发：一级 `install` / `tzs` / `tzc`（外加 `help`），二级是 `.tzc` 的动词。
未知子命令或未知动词一律打 `Usage` 并退 2。

| 动词 | 作用 | 写盘吗 |
|---|---|---|
| `export <pkg> [-o <dir>] [--only <点名>…]` | 渲染工作区；`-o` 省略时用默认后缀（见下） | 新建工作区 |
| `status [<dir>]` | 看改了什么 | 否 |
| `verify [<dir>] [--strict]` | 解析围栏 + gate1 + gate2 | 否 |
| `apply [<dir>] [-o <pkg>] [--dry-run] [--yes]` | 完整管线：见下 | **原地覆盖源包**（原子写） |
| `unlock [<dir>] [--yes]` | 框架解锁：单向状态迁移 | 只改工作区 |
| `rename [<dir>] <旧名> <新名> [--scope …] [--desc …]` | 结构事务：改名 | 只改工作区 |
| `newfn [<dir>] --type FUNCTION\|DIALOG\|REPORT [--name …]` | 新增自订定义点 | 只改工作区 |
| `selftest` | 内置对抗用例（不需要真实语料） | 临时目录 |

`rename` / `newfn` / `unlock` 都**当场跑一遍结构事务预检**（与 `apply` 用同一套校验），
避免"命令行过了、apply 又拒"。它们改的都是工作区，仍须走 `apply` 才落到包里。

## `apply` 的管线顺序（顺序即契约）

```
解析围栏 → gate1 → 源包未变校验 → I15 拒绝 → split → gate2 → 改写字节 → gate3 → 原子写 → git 提交
```

逐项说明：

- **解析围栏**：配对、归属、嵌套深度（结构性第一道）。
- **gate1**：字节恒等 + 写入授权 → 失败退 3，权限类退 4。
- **源包未变校验**：对照 `manifest` 里记的 sha256，源包在 export 之后被改过就拒绝写回（退 5），
  避免对着错误的包写回。
- **I15 拒绝**：TAP 条目基名与根 `prog` 不一致 = 引用标准程序的包，设计器在那种情况下会改用
  另一套内容，所以拒绝整体重写。
- **split → gate2**：先算出要打哪些补丁，再对**原包**的 TAP 与编辑后的文档做不变量判定。
- **gate3**：对**将要产出的新包**重跑合成（"设计器打得开吗"），此时磁盘仍未动。
- **原子写**：临时文件 + Rename，然后备份上一版包、刷新工作区基线、提交一次 git。

**任何一步失败，原包字节不变。**

## 输出

`--json` 下 stdout **恰好一个 JSON 对象**：

- 成功：`{"ok":true, …}`（各动词字段不同，例如 `apply` 给新包的 sha256 与逐条目动作）
- 失败：`{"ok":false,"exit_code":N,"error":"…"}`（**走 stdout**，不在 stderr）

不加 `--json` 时失败写 stderr（`错误（退出码 N）：…`）。**退出码**：
`0` 成功 / `2` 包格式或用法错 / `3` 验证失败 / `4` 写入被拒 / `5` IO·环境失败
（`1` 只作未分类内部错误的兜底）。

`--strict`（只对 `verify`）把 warn 级发现也算作失败，供 CI 用。

## 配置接缝

`config.json` 的 `tdev` 节只在这里被读，且**只在省略 flag 时**补默认值：

| 键 | 作用 |
|---|---|
| `workspaceSuffix` | `tzc export` 省略 `-o` 时的默认后缀（缺省 `-ws`） |
| `defaultOut` | `tzs export` 省略 `-o` 时的默认目录 |

三条纪律写死在 `settings.go`：**命令行 flag 永远优先**；**绝不因为配置失败而失败**
（读不到配置就静默退回硬编码行为）；**不进热路径**（只在确实要用默认值时读一次）。

## 判据

```bash
go test ./internal/dev/cli          # 含语料用例；corpus_test.go 顶部说明了深度回归为何要显式开关
./tt.exe dev tzc selftest           # 内置对抗用例：通过 31，失败 0
```

`selftest` 用的包是**合成**的（不依赖真实语料），覆盖改围栏行、改只读区、删围栏、
塞非法字符、结构行被改、越权删除与追加、改名事务、新函数模板等被拒的场景，
并且断言**被拒的 apply 不改包**。

## 细节去哪

- 各动词的用法、常见错误、退出码速查 → [`skills/tt-dev-tzc/SKILL.md`](../../../skills/tt-dev-tzc/SKILL.md)
- 闸门与不变量的含义 → [../verify/README.md](../verify/README.md)
- 工作区长什么样 → [../store/README.md](../store/README.md)
