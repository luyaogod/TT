# internal/cli/dev/tzc — `.tzc` 代码包线的动词

八个动词的实现体（`cli.go` 的 `cmdXxx`）与 `uuid` 生成（`uuid.go`）、31 项对抗自检（`selftest.go`）、
动词接线清单（`verbs.go`，父包的 cobra 树从这里取 Use/Short/Long/Run）。
参数解析权在动词自己的解析器（`common.ParseArgs`，位置无关）——契约见 [../README.md](../README.md)。

**安装不在这条线上**：skills 与 PATH 统一走 `tt install skills` / `tt install path`；
`install` 子命令在父包留了一块墓碑。

## 动词

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
- **源包未变校验**：对照 `manifest` 里记的 sha256，源包在 export 之后被改过就拒绝写回（退 4，
  文案给出 `tt dev relink` 出路），避免对着错误的包写回。
- **I15 拒绝**：TAP 条目基名与根 `prog` 不一致 = 引用标准程序的包，设计器在那种情况下会改用
  另一套内容，所以拒绝整体重写。
- **split → gate2**：先算出要打哪些补丁，再对**原包**的 TAP 与编辑后的文档做不变量判定。
- **gate3**：对**将要产出的新包**重跑合成（"设计器打得开吗"），此时磁盘仍未动。
- **原子写**：临时文件 + Rename，然后备份上一版包、刷新工作区基线、提交一次 git。

**任何一步失败，原包字节不变。**

## 输出

`--json` 下 stdout **恰好一个 JSON 对象**：

- 成功：`{"ok":true, …}`（各动词字段不同，例如 `apply` 给新包的 sha256 与逐条目动作）
- 失败：`{"ok":false,"exit_code":N,"error":"…"}`（**走 stdout**，不在 stderr；
  实现在 [../common/scaffold.go](../common/scaffold.go) 的 `Fail`，两条线同一份）

不加 `--json` 时失败写 stderr（`错误（退出码 N）：…`）。
`--strict`（只对 `verify`）把 warn 级发现也算作失败，供 CI 用。

## 配置接缝

`config.json` 的 `tdev` 节只在这里被读（[../common/scaffold.go](../common/scaffold.go)
的 `LoadTdevSettings`），且**只在省略 flag 时**补默认值：

| 键 | 作用 |
|---|---|
| `workspaceSuffix` | `tzc export` 省略 `-o` 时的默认后缀（缺省 `-ws`） |

三条纪律：**命令行 flag 永远优先**；**绝不因为配置失败而失败**；**不进热路径**。

## 判据

```bash
go test ./internal/cli/dev/tzc -count=1   # 含语料用例与那 31 项自检（约 20 秒）
./tt.exe dev tzc selftest                 # 命令行口：通过 31，失败 0
```

`corpus_test.go` 顶部说明了深度回归为何要显式开关（9–11 分钟 > `go test` 默认超时）。

**那 31 项内置对抗用例现在有两入口，同一个定义处**：

```bash
go test ./internal/cli/dev/tzc -run TestSelftestCases -v -count=1   # 默认档：31 个子测试
./tt.exe dev tzc selftest                                           # 命令行：通过 31，失败 0
```

`TestSelftestCases` **只写驱动不写断言** —— `selftestCases()` 是那 31 项的唯一定义处，
谁加一项 `go test` 自动跟着长。它们走的是 `cmdExport` / `cmdApply` / `cmdVerify` /
`cmdRename` / `cmdNewfn` / `cmdUnlock` 的**完整命令路径**（tzs 那一项走 `tzs.Run`，
所以本包 import tzs），用的包是**合成**的（不依赖真实语料），覆盖改围栏行、改只读区、
删围栏、塞非法字符、结构行被改、越权删除与追加、改名事务、新函数模板等被拒的场景，
并且断言**被拒的 apply 不改包**。

**故意不覆盖**：真实语料的形状分布 —— 那归 `TDEV_DEEP=1` 的两条。

## 细节去哪

- 结构与接线、根开关桥 → [../README.md](../README.md)
- 闸门与不变量的含义 → [../../../dev/verify/README.md](../../../dev/verify/README.md)
- 工作区长什么样 → [../../../dev/store/README.md](../../../dev/store/README.md)
- 各动词的用法、常见错误、退出码速查 → [`skills/tt-dev/SKILL.md`](../../../../skills/tt-dev/SKILL.md)

## 区域级动词（2026-11-21 新增，实测在 axmt500_wf 上验证）

`prog.full.4gl` 有 7.7 万行，而一次改动通常只碰一个点里的几行。这五个动词把「围栏围起来的
点 / 区段」变成一等对象，**只改 workspace，不写包**（落包仍走 `verify` / `apply`，三道闸门不变）：

```
tt dev tzc points [<dir>] [--editable] [--nonempty] [--kind point|section] [--match <子串>]
tt dev tzc show   [<dir>] <点名> [--max N]
tt dev tzc find   [<dir>] <点名> <子串> [--context N] [--case] [--limit N]
tt dev tzc patch  [<dir>] <点名> (--line N --text S | --old S --new S | --old-file F --new-file F) [--dry-run]
tt dev tzc set    [<dir>] <点名> (--file F | --stdin) [--max-lines 300] [--dry-run]
```

规矩（都是试用踩出来的，别省）：

1. **改一行用 `--line N --text S`**（N 用 `show`/`find` 打印的绝对行号），工具保留原缩进。
   子串模式 `--old/--new` 只在"唯一且行结构不破坏"时可用：old/new 任一侧含围栏直接拒，
   会拼出 `##`、或凭空多出空行的替换也拒。
2. **先 `--dry-run` 再落盘**：纯子串替换曾静默拼出 `##` 且退出码 0，现在有预览 + 结构判据。
3. **大区域用 `find` + `patch`，不要 `set`**：`set` 超过 `--max-lines`（默认 300）拒；
   含内嵌加点的区段也拒（整段替换会丢围栏）。
4. 改区域**不写 `.4gl`**：实测 `apply --dry-run` 只改 `.tap`（`.4gl` 不变），与设计器一致（产物由 F7 重产）。
   改**区段正文**例外：会连带 `.tgl` 补丁 + `section_flag=Y`（`status` 会单列"改动的区段"）。
5. 退出码按组文档：2 用法 / 3 验证失败 / 4 写入被拒（不可编辑、不唯一、找不到、围栏）/ 5 IO。

实测（axmt500_wf，未解开）：`points --editable --nonempty` → 1,269 个区域；
`show input.a.page73.xmdduc001` → 23 行；`find axmt500_wf.input.body 'cl_chk_exist'` → 2.1 万行的
区段里 16 条命中；`patch` 后 `status` 报 `prog.full.4gl:50587`，`apply --dry-run` 计划「改动 1 个点」。
