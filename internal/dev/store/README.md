# internal/dev/store — 工作区

工作区的落盘与读取：`export` 写出它，`apply` 从它读基线、成功后刷新它。布局是固定的，
不随环境变化。

```
<dir>/
  prog.full.4gl        唯一编辑文件：带围栏的渲染文档
  manifest.json        人机共读索引（每个 Region 的权限与不可编辑原因）
  snapshot/
    index.json         原包条目清单：name / sha256 / size / role
    entries/<条目名>    原包每个条目的逐字节拷贝（恢复源；apply 不从这里读）
  .tdev/
    base.full.4gl      导出时 prog.full.4gl 的字节拷贝（gate1 的比对基线）
    base.sha256        该基线的 sha256
    regions.json       Region 表 + 字节区间 + 基线摘要
    lock               互斥锁（防两个进程同时操作）
  .git/                export 时 init 并首次提交；apply 成功后再提交
```

**判据之一是"同一输入产出逐字节相同的工作区"** —— 所以 `manifest.json` 与 `regions.json` 里
**不含任何时钟**；唯一允许出现时钟的地方是 `.tdev/lock`（只用于诊断锁的归属）。

## 三条硬约束

| | 约束 | 实现 |
|---|---|---|
| R6 | 一切落盘走**原子写**（同目录临时文件 → Rename），禁止先删后建 | `AtomicWrite` |
| R7 | 时间戳不是功能 | 序列化里不带时钟 |
| — | **不用库调 git**：git 不是依赖，只是版本留痕 | `git` 命令行，并显式关掉 `core.autocrlf`，让 git 里的字节 == 工作区里的字节 |

git 是**可选**的：环境里没有 git 可执行文件时，`export` 照常产出完整工作区，只是不建版本库。

## 锁

`Lock()` 用 `O_CREATE|O_EXCL` 原子抢占 `.tdev/lock`（两个进程不可能同时成功）。
已存在且未陈旧 → IO 错误（退出码 5）；**mtime 超过 30 分钟**视为上次进程已死 → 删掉重试一次。

## 入口

| 入口 | 作用 |
|---|---|
| `Create(dir, doc, pkg, fenced)` / `Open(dir)` | 建 / 开工作区（目标已存在、或工作区不完整 → 拒绝） |
| `ReadEdited` / `WriteEdited` / `ReadBase` | 编辑文件与基线 |
| `Manifest()` / `Regions()` / `EffectiveUnlockState()` | 读索引、Region 表、合并后的解锁状态 |
| `WriteSectionState` / `ReadSectionState` | 解锁状态（`unlock` 只改工作区，写在这里） |
| `UpdateAfterApply(doc, fenced)` | apply 成功后把基线推进一格 |
| `SnapshotPkg` / `RefreshPackage` / `BackupPackage` / `SnapshotEntry` | 快照与回滚备份（`BackupPackage` 产出上一版包的整文件备份） |
| `Lock` | 互斥锁 |
| `GitInit` / `GitCommit` / `GitHead` | 版本留痕 |

**旧工作区的读取**：早期版本有 `--allow-sec` 时留下的 `export.allow_sec` 字段只被**读**，
用于把老工作区认成"已解锁"（`LegacyAllowSec`）；新写的工作区不再产出该字段。

## 判据

```bash
go test ./internal/dev/store -count=1
# TestCreateWorkspaceLayout / TestCreateRefusesExistingWorkspace / TestOpenRejectsIncompleteWorkspace
# TestAtomicWriteOverwrite / TestAtomicWriteRenameOntoDirectoryFails
# TestLockMutualExclusionAndStale / TestGitLifecycle
# TestRefreshPackageUpdatesRecordedPkg / TestBackupPackageKeepsPrevious
```

## 细节去哪

- 工作区里那份文档是怎么合成/渲染的 → [../synth/README.md](../synth/README.md)、[../fence/README.md](../fence/README.md)
- 逐条目写回计划从哪来 → [../split/README.md](../split/README.md)
