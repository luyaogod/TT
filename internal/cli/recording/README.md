# internal/cli/recording — `tt recording` 命令组

录制件的命令面：起一轮、停一轮、列（正在录的 / 录过件的）、回放、留、删、清、建索引、标失效。

实现在 [`../fep`](../fep)（`rec_*.go`：起停、体检、回放、留存、索引）与
[`../../fep`](../../fep)（帧流、索引、身份文件）。本组**不复制实现**：每个动词都是
`fep.RecordingCommands()` 里那条命令的镜像（同一个 `RunE`、同一批标志），
`recording_test.go` 逐条钉住"名字对得上、标志一个不丢"。

## 命令

| 命令 | 作用 |
|---|---|
| `tt recording start --env <环境> --ent <ENT> --prog <作业>` | 起一轮录制（默认走调试线，可下断点） |
| `tt recording start --session <ref|别名> --prog …` | 把录制挂到一条**已有空闲且有名**的会话上 |
| `tt recording start --launch standard …` | 走 T100 标准启动器（`r.r`）：界面同用户平时所见，**不可调试** |
| `tt recording stop <录制ID>` | 停一轮：先请常驻进程自退（写完帧、落身份），再只杀登记的 pid，最后摘登记 |
| `tt recording active` | 列**正在录的**（读登记 + 判本机 GDC 是否还活着） |
| `tt recording ls` | 列录制件（缓存 + 长期库） |
| `tt recording show <ID>` | 身份、指纹、路径、现在能不能用 |
| `tt recording replay <ID>` | 在调试器下回放：指纹闸 → 身份闸 → 发包；`--no-wait` 只发不等停站 |
| `tt recording keep / rm / prune / index / mark-stale` | 留件 / 删件 / 清理 / 建帧索引 / 标记失效 |

## 改代码必须知道的三件事

1. **收尾顺序不能换**（`rec_stop.go` 与 `cli/fep/session.go` 的 `closeRegisteredSession`）：
   **先结束调试**（按登记里的 `ref` 收口）→ 常驻进程自退 → 只杀**登记里**的 pid → 最后摘登记。
   - 调试先收：代理是作业画面唯一的出口，先杀代理会让作业断连、会话以 exit 收场，
     "收工后宿主保留、下一轮免登录"当场失效（真机复现过：stop 之后 `session ls` 空）。
   - 常驻进程先自退：它要把帧写完、把身份落盘；
   - 只杀登记里的 pid：反过来会把远端作业留在服务器上，或者把别人的登记一起清掉。
2. **挂靠的三条前置**（`bindRecSession`）：会话必须空闲、必须有名字、环境不许打架。
   理由是三条安静的失败：作业在跑时指 `FGLSERVER` 不生效；没名字回不到同一条会话；
   环境打架会落到别的会话上。
3. **标准启动的录制不可回放**：`r.r` 不带 `-d`，没有调试协议；回放前路径比对会拒它。

## 判据

```bash
go test ./internal/cli/recording/ ./internal/cli/fep/ -count=1
```

- `recording_test.go`：`tt recording` 逐条镜像 `fep.RecordingCommands()` —— 名字映射、
  标志一个不丢、没有重名；
- `rec_start_test.go`：录制动作的**调用序列**（switch → topent → launch → 等停站 → break → 回填 ref → 放行）、
  `--no-continue` 不放行、`--launch` 三条用法规矩、挂靠的三条前置；
- `rec_stop` 与录制的端到端（起 → 登记 → 停 → 登记清空）在真机验证，命令见下：
  `tt recording start --env <环境> --ent <ENT> --prog <作业>` 后
  `tt recording active` 应看到该 tag，`tt recording stop <tag>` 后应清空。

## 细节去哪

- 起停与体检的实现：[`../fep`](../fep)（`rec_start.go` / `rec_stop.go` / `rec_preflight.go` / `rec_reproduce.go`）
- 帧流、索引、身份文件：[`../../fep/README.md`](../../fep/README.md)
- 会话侧：[`../session`](../session)
- 命令树与退出码总表：[`../README.md`](../README.md)
