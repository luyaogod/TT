# internal/dev/verify — 三道闸门

写回之前的**机械证明**：改了的东西是不是只有该改的那些，产出的包设计器打不打得开。

| 闸门 | 检查 | 失败 |
|---|---|---|
| **gate1** | 围栏行、围栏外字节、只读区内容、可编辑区里的结构行 —— 逐字节相等；以及删除/追加的授权 | 校验失败；权限类 → 4 |
| **gate2** | 不变量全集 I1–I15（区段配对、命名空间错配、状态取值、编码、未知条目透传…） | 校验失败 |
| **gate3** | 对**将要产出的新包**重跑合成 = "设计器打得开吗"的判决（此时磁盘还没动） | 校验失败 |

管线顺序即契约（取自 `cli.go` 的 `cmdApply`）：
`解析围栏 → gate1 → 源包未变校验 → 回写拆分 → gate2 → 改写字节 → gate3 → 原子写 → 提交`。
gate3 在写之前、对内存里的新包执行。**任何一步失败，原包字节不变。**

## 发现码（`Finding.Code`）

| 前缀 | 含义 |
|---|---|
| `gate1.*` | 字节恒等与授权：`fence-line`（围栏行锚定部分被改）、`outside-fence`、`readonly-region`、`region-count`、`delete-*` / `append-*`（授权）、`kind`、`eol-normalized`（**info**：受保护字节只有行尾差异） |
| `I*` | gate2 的不变量：`I1` 区段配对、`I2a` 占位符无点是**正常**（info）、`I2b` 有点无锚点是 error、`I2c` 孤儿点（warn，必须透传）、`I3` 区段正文里混进完整函数块、`I4` 定义头形态、`I5` 状态取值、`I10`（info）、`I11`/`I11b` 编码与 CDATA 合法性、`I13` 未知条目与版本文件字节原样、`I-structure` 可编辑点的结构行未改 |
| `V*` | 结构事务（见下） |
| `gate3.*` | 装载模拟各步：`open` / `synthesize` / `render` / `fence` / `tap` / `build` / `invariants` |
| `structtx` | 结构事务的汇总条目（info） |

等级三档：`info` / `warn` / `error`。**退出码由发现决定**：只要有 error 级且标记为"写入被拒"的
发现就退 **4**，否则退 **3**（`VerifyErrorFromFindings.ExitCode`）。`info` 与 `warn` 默认不失败 ——
`verify --strict` 把 warn 也算作失败（供 CI 用）。

## 结构事务 V1–V7（`structtx.go`）

改名 / 改 scope / 改描述这类"点的身份变了"的操作，要跨六处保持一致，所以单独一套校验：

| 码 | 判定 | 等级 |
|---|---|---|
| V1 | 围栏里的 `fn` 与签名行的函数名一致 | error |
| V2 | 旧名残留引用清零：可编辑区里由你自己改；**只读区段里改不了 → 写入被拒（4）** | error |
| V3 | 新名不与现存的点/占位符重名（含本次会话内的冲突与交换式改名） | error |
| V4 | 命名规范（前缀）——与设计器一致，只提示 | warn |
| V5 | scope 与程序类型的默认关系 | warn |
| V6 | 描述块里非空行必须以 `#` 开头（**空行合法**）；等级按"是否本次改动过"区分 | error / warn |
| V7 | 目标必须是可编辑的自订定义点 | error |

本包放在 `verify` 里而不是独立包，是因为要返回 `Finding`，独立包会造成循环依赖。

## 判据

本包没有自己的测试文件；行为由使用者覆盖：

```bash
./tt.exe dev tzc selftest                       # 对抗用例（改只读区、删围栏、塞非法字符…）
TDEV_DEEP=1 go test ./internal/dev/cli -run TestCorpusExportVerify
                                               # 全语料验证（9–11 分钟）
```

`verify` 的输出带定位：每条发现带文件、1-based 行号与该行内容（`--json` 下给结构化字段）。

## 细节去哪

- 解析结果从哪来 → [../fence/README.md](../fence/README.md)
- 发现怎么变了新包的字节 → [../split/README.md](../split/README.md)
- 命令行与退出码 → [../cli/README.md](../cli/README.md)
