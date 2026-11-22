# cli/dev/task —— 工作区命令组

`tt dev` 下的工作区动词。地址是**工作区目录**（缺省从当前目录往上找 `task.json`），
环境只在 `open` 时需要一次。

## 命令

| 命令 | 作用 | 联网 |
|---|---|---|
| `open <程序>` | 拉材料 + 建索引 + 打包 + 导出可编辑视图 + 写 `task.json` | 是 |
| `update` | 向服务器核对，变了才重拉；重建索引；按"有没有活"决定重建视图/包 | 是 |
| `status` | 材料/索引/视图/包 四项现状（全部现算） | 否 |
| `brief` | 一页纸画像：判型/页签/客制逻辑/风险/意图清单 | 否 |
| `find` | `--intent`（check/lookup/default/save/display/init/query/sign）+ `--field`（列名**或画面中文名**）+ `--page` 定位落点 | 否 |
| `at <行号\|点名>` | 行号 → 归属；点名 → 内容（缺省看**包内 .tap**，`--from 4gl` 看服务器产物） | 否 |
| `search <子串>` | 检索产物与 .tap（`--scope all` 缺省 / 4gl / tap / pkg），apply 后自动查**包内** .tap，每条命中标出归属与"有无槽位" | 否 |
| `apply` | = `tt dev tzc apply <ws>`（写回 `pkg/`；包与基线不一致时拒绝） | 否 |
| `verify` | = `tt dev tzc verify <ws>`（包被换过时给"通过"是假绿灯，已拦） | 否 |
| `relink` | 以当前包为基线重导视图（解"包被换过"的死局；`--force` 才丢未落盘改动） | 否 |
| `unlock` | = `tt dev tzc unlock <ws>`（框架解锁，单向、有代价） | 否 |

公共开关：`--dir <工作区>`（缺省从 cwd 往上找）、`--json`（缺省）/ `--human`。

## 退出码

| 码 | 含义 |
|---|---|
| 0 | 成功 |
| 2 | 用法/身份错（不是工作区、目录非空、程序名/任务名非法、环境不存在、工作区 schema 不是本版认的…） |
| 3 | 校验失败（verify/apply 的 gate 发现 error 级问题；apply/verify 直通 tzc 线） |
| 4 | 未找到（没有这个程序/材料/点位）或写入被拒（有未落盘改动、包与视图基线不一致、缺 `--yes`） |
| 5 | IO/网络失败 |

## 一次典型会话

```
tt dev open axmt500_wf --at 测试环境        # ~30 s（拉 11.8 MB）
cd axmt500_wf
tt dev brief                                   # 页签标题/客制点/风险
tt dev find --intent check --field 刻号         # 中文名也行（对应 xmdduc003）
tt dev at 44107                                # 编译报错的行属于哪个点（.4gl 行号）
tt dev at input.a.page73.xmdduc003             # 直接看包内 .tap 里的正文（可验证 apply 结果）
tt dev at xmdducseq                            # 孤儿点：会说明"有代码但不生效"
tt dev search 不允許改寫 --scope all            # .4gl 搜不到时，.tap 里找（apply 后查包内）
# 编辑 ws/prog.full.4gl（唯一可编辑文件）
tt dev verify                                  # 只审计
tt dev apply                                   # 写 pkg/axmt500_wf(c).tzc
tt dev update --check                          # 交出去前确认服务器没变
```

## 输出约定

- `--json`（缺省）：`{ok, source, env, totalRows, returned, data}`，与 `tt gen` 同形。
- `--human`：每个数字都来自同一份 JSON 结果对象，不另算一份。
- 进度一律打 stderr，stdout 只有结果。

## open 的产出证明

`open` 之后可以直接回答三件事（全离线，具体数字以你手里那个程序的材料为准）：

- `brief`：程序判型、页签清单（页序 / page_id / screen record / 栏位数）、
  客制点数与行数、切片种类、模板脱版告警。
- `find --intent check --field <字段>`：该意图下的候选落点（input.a/b/c/g…）、
  参考切片与用量；包内没有声明的点是"模板洞"。
- `at <行号>`：落在哪个加点、是不是客制、属于哪个 section、由哪个切片生成。

## 包被换过之后怎么办

`apply` 会校验"源包自 export 之后没被改过"。包被替换（拷错、手工覆盖、别人给了一份）后
这个校验永远失败，而旧文案建议的"放回基线包"通常你手里没有。两条真出路：

- **把基线那份包放回来**（有备份的话）；
- **`tt dev relink`**：承认当前包就是新基线，按它重导视图。视图里未落盘的改动会丢，
  所以默认拒绝脏工作区。之后 `apply` 恢复正常。

`status` 里 `pkg.matchesBase=false` 就是这种状态，摘要行会直接写"与视图基线不一致（不可交付）"，
不会再出现"可交付"字样。

## 三条"看起来通过、其实没验"的边界（实测试用者踩过）

1. **verify/apply 不编译 BDL**。它们只查围栏结构、不变量与装载模拟；语法错要到设计器 F7
   才暴露。两条命令成功时都会在 stderr 提醒这一点。
2. **`at <行号>` 的行号是 .4gl 产物的行号**（`fglcomp` 报错那种）。视图 `ws/prog.full.4gl`
   的行号不同，别混用；点名模式默认读包内 `.tap`，打的是 `.tap` 行号。
3. **孤儿点**（有 `.tap` 声明、骨架里没有槽位）：`brief` 会点名列出客制孤儿，
   `at <点名>` 会给出正文并标注"重产不保留"，`search --scope tap` 能搜到它们。
## 行号有三种，别混（实测试用者因此误判过两次）

| 你在看什么 | 行号 | 怎么查 |
|---|---|---|
| `fglcomp` 报错 / 服务器 `.4gl` 产物 | `.4gl` 行号（69,269 行） | `tt dev at <行号>`，头部会写明"按 .4gl 产物行号解读" |
| 可编辑视图 `ws/prog.full.4gl` | 视图行号（77,xxx 行） | 用点名查（`tt dev at <点名>` 会按包内 .tap 给内容） |
| 包内 `.tap`（apply 之后的最新真相） | `.tap` 行号 | `tt dev at <点名>`（缺省）/ `tt dev search --scope pkg` |

`at <行号>` 只认 `.4gl` 产物行号；给别的行号时它不会静默对齐，头部会说明依据。