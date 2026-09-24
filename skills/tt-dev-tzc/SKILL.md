---
name: tt-dev-tzc
description: 安全编辑 T100 设计器**代码包**（.tzc/.tzf/.tzx）：export 渲染成带围栏的 4GL/TGL 工作区 → 只改围栏内正文 → verify 预检 → apply 原子写回。当用户要改 T100 客制程序、新增/改名/改签名自订函数、解锁只读框架区段、或查看 .tzc 包里有什么时使用。**表单包（.tzs/.tzv）不归这里，用 tt-dev-tzs。**
license: 与 tt 仓库一致（见随包 README.md）
metadata:
  tool: tdev
  command: tt dev tzc / tt install
---

# tt dev tzc：T100 设计器代码包的安全编辑

**这是代码包（`.tzc`/`.tzf`/`.tzx`）的技能。** 表单包（`.tzs`/`.tzv`）是**另一条互不相通的管线**，
用 `tt-dev-tzs`：

| | `tt dev tzc`（代码包，本技能） | `tt dev tzs`（表单包，见 tt-dev-tzs） |
|---|---|---|
| 干什么 | 把包渲染成**带围栏的 `.4gl` 工作区**给人/AI 编辑 | 由设计器自己的引擎读写表单模型 |
| 产物 | 工作区：`prog.full.4gl` + `manifest.json` + `snapshot/` + `.tdev/` + `.git/` | 就是一包文件（`.tsd`/`.4fd`/`ver`…） |
| 唯一写路径 | `apply`（闸门校验 + 原子写） | `call save --out`（写**新**包） |

拿错入口会被挡住：`.tzc` 跑 `tzs export` → 退出码 2 并提示改用 `tzc export`。

## 标准流程

```powershell
tt dev tzc export "D:\pkg\capt110(c).tzc"    # → D:\pkg\capt110-ws（默认，-o 可改）
cd /d D:\pkg\capt110-ws                      # cd 进去之后所有命令都能省 <dir>
# 只改围栏内标着 EDITABLE 的正文
tt dev tzc status                            # 改了哪些点、在第几行、会被拒还是放行
tt dev tzc verify                            # 不写盘预检
tt dev tzc apply                             # 唯一会改 .tzc 的命令（先备份到 .tdev\prev.tzc）
```

`status` 的改动清单带行号：`改动的点  function.foo（prog.full.4gl:1234）`。
小步走 —— 一次改一大片再 `apply`，被拒时定位成本高得多。

## 红线（碰了就出事，没有例外）

1. **不许改围栏行**：`{//@tdev:begin point …}` / `{//@tdev:end point}` / `{//@tdev:end section}`。
   它们不是内容，是 tt dev 的坐标（区间记账），改了 gate1 直接拒。
2. **不许改围栏外的任何字节**：文件头尾、区段之间的空行、缩进对齐 —— 都不行。
   也**不要用编辑器格式化/重排整个文件**（只允许改围栏正文；行尾被归一成 LF 没事，改空白有事）。
3. **不许改 `END FUNCTION` / `END DIALOG` / `END REPORT` 行**（结构行，逐字节比对）。
4. **签名行的改动要走 `rename`**（它同步围栏 `fn` + 签名行 + 描述块，并在 `.tap` 里复刻墓碑事务）；
   手工只改一处会被 gate2 的 V1–V7 拒。
5. **`.4gl` 条目永远不被写回**（服务器 build 产物，渲染结果装不下它）。
6. **不要手改 `.tap` / 不要直接改 zip**：`.tzc` 的唯一写路径是 `apply`。
7. **不要编辑 `.tdev/base.full.4gl`**：那是 gate1 的比对基线，改它等于伪造验证。
   要改内容就改 `prog.full.4gl`，要改意图就用 `unlock`/`rename`/`newfn`。
8. **只读区段不要硬改**：先 `tt dev tzc unlock`（单向、有代价）；锚点区段永远只读，改不了。
   **`--yes` 必须由人确认，AI 不得自己加**。
9. **不要写真实语料/客户包**：`export` 只读源包；实验一律用副本或临时目录。

## 退出码

`0` 成功 · `2` 包格式/用法错 · `3` 验证失败 · `4` 写入被拒（权限/确认不足）· `5` IO·环境失败。

（`.tzs` 那条线**没有 `3`**、多一个 `1` —— 见 `tt-dev-tzs`。）

## 报错怎么读（先看行号）

`apply` / `verify` 拒绝时会打位置，直接照着改：

```
apply 被拒绝（gate1 error=1 warn=0 info=1）：
  [error] gate1.readonly-region  只读 Region 内容被改动（写入被拒）
            位置：prog.full.4gl:4          ← 你编辑的那份文件 + 行号
            该行：   # 我加的字             ← 该行内容
            deny=section-locked            ← 原因：框架未解开，先 unlock
```

删除类问题给 `.tdev/base.full.4gl:<行号>`（编辑后的文档里它已经不存在了）。
`--json` 里每条发现有 `file`/`line`/`snippet`；`status --json` 有带行号的
`changed_regions`/`added_regions`/`deleted_regions`。

## 常见简单错误（自查表）

- ❌ 拿 `.tzc` 去跑 `tzs export` → 退出码 2；代码包走 `tzc export`。
- ❌ 以为 `export` 改了包 → 它**只读**源包；写回只有 `apply`。
- ❌ 在没有工作区的目录里跑 `status`/`apply` → 退出码 5；先 `cd` 进工作区或显式给 `<dir>`。
- ❌ 改围栏行 / `END` 行 / 围栏外空行 → gate1 拒（退出码 3/4）。
- ❌ 用格式化工具重排 `prog.full.4gl` → 围栏外字节变了，被拒。
- ❌ 改只读区段不 `unlock` → 退出码 4；锚点区段即使 unlock 也只读。
- ❌ 源包在 `export` 之后被别人动过 → `apply` 拒写（退出码 5），重新 `export`。
- ❌ 一次改一大片再 `apply` → 先用 `status` 看清单与行号，小步走。
- ❌ 新增函数后以为要给 `--desc` → `newfn` 没有 `--desc`，函数头用设计器模板（含顶部空行），
  自己填 `# Descriptions...:` 与正文。

## 动词速查

| 命令 | 作用 | 改 `.tzc` 吗 |
|---|---|---|
| `export <pkg.tzc> [-o <dir>] [--only <点名>...]` | 包 → 工作区（含 git init 与逐条目快照） | 否 |
| `status [<dir>]` | 报告改动（含行号），不写盘 | 否 |
| `verify [<dir>] [--strict]` | 不写盘跑 gate1 + gate2 | 否 |
| `apply [<dir>] [-o <pkg>] [--dry-run] [--yes]` | 全阶段校验 + 原子写包 + git commit | **是（唯一）** |
| `unlock [<dir>] --yes` | 框架解锁（单向、不可逆；先不带 `--yes` 看代价警告） | 否（只改工作区） |
| `rename [<dir>] <旧> <新> [--scope …] [--desc …]` | 结构事务：围栏 + 签名行 + 描述块原子同步 | 否 |
| `newfn [<dir>] --type FUNCTION\|DIALOG\|REPORT [--name …] [--scope …]` | 在 APPEND 锚点插入新空函数（带模板） | 否 |
| `selftest [--json]` | 31 项内置对抗用例，不需要真实语料 | 否 |

工作区动词的 `<dir>` 可以省略：先 `cd` 进工作区，命令就不用再写目录。

```powershell
cd D:\pkg\capt110-ws
tt dev tzc status          # = tt dev tzc status D:\pkg\capt110-ws
tt dev tzc rename adzi999_calc adzi999_count   # 省略 <dir> 时 rename 收 2 个位置参数
```

## 安装

```powershell
tt install path      # 把 tt.exe 所在目录加进用户 PATH（只动 HKCU）；一次安装覆盖全部命令
tt install skills    # 把技能文件复制到当前目录的 skills/
```

细节（围栏协议、不变量 I1–I15、结构事务 V1–V7、与设计器的全部偏差 D-1…D-10/DV-2/DV-3、
真机验收清单 S7）见随包 `README.md` 与 [docs/WIKI.md](../../docs/WIKI.md#6-tt-dev-tzc-代码包)。
