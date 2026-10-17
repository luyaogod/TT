# internal/cli/fep — `tt fep` 命令组

T100 前端协议线（读写运行中的作业界面）的命令面。内核在
[`../../fep`](../../fep/README.md)：帧编解码、AUI 模型、字段名解析、语义动作都在那边，
**不碰 socket、不碰窗口**；本组负责把「录制件 / 会话登记」这类输入交进去，把结果打到输出出口。

## 命令

| 命令 | 作用 |
|---|---|
| `tt fep session open` / `session ps` / `session close` | 隔离实例 + 代理的开关与登记 |
| `tt fep rec start` / `rec list` / `rec show` / `rec steps` / `rec keep` / `rec rm` / `rec prune` | 录制：起一轮、看清单、看步骤、留件与清理 |
| `tt fep rec preflight [--prog <作业>] [--env <环境>]` | **前置体检**：目录/服务/会话/环境一次查清（**只检查**，不录制、不起会话） |
| `tt fep rec reproduce <ID>` / `rec mark-stale <ID>` | 回放取证与旧件处置 |
| `tt fep do …` / `ui …` / `wait …` / `src …` / `job …` / `fe …` | 动作、界面快照、等待、源码、作业、前端事件 |

`--format text` 出人读态；不加（或 `--format json`）出 JSON —— 本组一直如此
（`printJSONOrText`）。**`--json` 下 stdout 恰好一个对象**，失败时也是它。

## 前置体检（`rec preflight`）

真机冒烟（[`../../../.agents/plan/L1-结果-2026-10-18.md`](../../../.agents/plan/L1-结果-2026-10-18.md)）
里 `tt fep rec start` 连试两次都失败，而且失败信息是**逐层挤出来**的：第一次"服务没起"，
起了服务再跑"没有会话"，随后又冒一条"切换会话：没有 SSH 会话"。使用者被迫来回试错凑前置。
体检查的就是那四项：

| 检查项 | key | 判据 |
|---|---|---|
| 数据目录可写 | `dir` | **真去** `<数据目录>/fep` 下写一个临时文件再删（`os.Stat` 的权限位在 Windows 上说明不了任何事） |
| 调试服务 | `service` | TCP 800ms 连一下 + 900ms 查一次 `/debug/api/status`、`/api/status` 的 `tdebug-debug` 标记（**不与失败本身比慢**） |
| 调试会话 | `session` | 有没有、是**哪条**（别名/句柄 + 环境 / 作业 / 状态） |
| 环境/作业 | `env` | 给了没有、在不在配置里（解析器与 `tt debug`/`tt dict` 同一句） |

### 三态：`ok` / `missing` / `unknown`（契约 A31）

每一项只有三种状态，**不是"过/不过"两值**：

| state | 含义 | 计入 `missing` | 退出码 |
|---|---|---|---|
| `ok` | 明确就绪 | 否 | — |
| `missing` | **确定**缺失（说得出缺什么 + 给得出修复命令） | 是 | 2 |
| `unknown` | **检查本身失败**（解析不了 / 超时 / 依赖故障） | **否** | **不变** |

`unknown` 必须**打印**"这一项查不了（原因）"，不许静默当 ok，也不许升级成失败。
退出码只由 `missing` 决定（`preflightReport.Unknown` / `preflightItem.State` 就是这两件事的机器判据）。

真机事故（2026-10-18）：`tt debug status --json` 实测输出**两段** JSON（状态段 + `{"ok":true,"sessions":[…]}`），
老 `parseDebugSessions` 只 `json.Unmarshal` 一次 ⇒ `invalid character '{' after top-level value`
⇒ session 项被当成"没有 SSH 会话"计入 missing ⇒ **服务起着也 exit 2**，而 `rec start`
开跑前过同一份体检 ⇒ **`rec start` 必然 exit 2**，`path=debug` 的件根本录不下来。
现在 `parseDebugSessions` 用 `json.Decoder` 逐段 `Decode` 到 `io.EOF`，取**带 `sessions` 数组**的那一段
（段数不是契约：两段/三段/空行/CRLF 都认；一段都读不懂才是 `unknown`）。

```
  ? 调试会话          列会话失败：读不懂 `tt debug status --json` 的报文（第 1 段）：…
有 1 项没检查成功（查不了：不计入缺失，也不影响退出码）：
  ? 调试会话          这一项查不了：列会话失败（读不懂 …）
```

不可写时给的处置是具体的：

```
看起来是从低完整性/受限进程启动的（例如 tt.exe 带 Low 标签）：
请改用普通权限终端，或 icacls <exe> /setintegritylevel Medium
```

为什么值得单列：Low 完整性标签下目录 ACE 明明允许写，落盘照样 `Access is denied`，
而错误信息只说 "起隔离实例失败: open …: Access is denied"，看不出是**完整性/权限**问题。

### 「一次说清」

`rec start` 在**起代理之前**跑同一份体检：缺几项就列几项，每条都带**可直接复制**的命令
（已知的 `--prog` / `--env` 会填进去；没给时留 `<作业>` / `<环境>` 占位符）：

```
还缺 2 项前置：
  1) 调试服务没起：        tt debug serve
  2) 没有 SSH 会话：       tt debug start axmt500_wf --env 主机109
（都满足后再跑本命令）
```

位置也在**起代理之前**：代理占 GDC/代理/注入三个端口，半路失败要收，收不干净就留一个空会话。

### 全部可注入

四项检查都走 `preflightDeps`（`Dir` / `Writable` / `Service` / `Sessions` / `Hosts`），
生产实现是 `defaultPreflightDeps()`，测试换成假的 —— **不起服务、不连真机、不改 ACL**。
`Sessions` 默认就是 `rec_session.go` 的 `debugSessions`（`tt debug status --json` 子进程），
调试服务的 HTTP 协议仍然只有 `internal/cli/debug` 那一份实现。

## 退出码

**唯一权威是 [`../README.md`](../README.md) 的那张表**（`tt fep` 行），本包只做实现（`fep.go:25-33`）：

```
0 成功 / 1 运行失败（IO·网络·环境）/ 2 用法错 / 3 无会话 / 4 会话状态不符
5 目标不存在 / 6 环境体检不过（权限不一致）
```

`rec preflight` 与 `rec start` 的缺前置都退 **2**（用法/环境不满足）—— **不新造码**，
也**不**用 6：6 是"体检跑过了但权限不一致"那类结论，而缺 `tt debug serve` 属于环境没满足。
清单打出去之后再失败时，preflight 用 `AlreadyReported()` 的记号让根命令跳过 `reportError`，
否则 `--json` 下 stdout 上会出现两个对象（与 `tt form verify` 同一个手法）。

## 判据

```bash
go test ./internal/cli/fep/ -count=1     # 含：一次说清、体检全就绪/不可写、体检不落盘
go vet ./internal/cli/fep/
```

钉住的几条（`rec_preflight_test.go` / `rec_start_missing_test.go` /
`rec_preflight_unknown_test.go` / `rec_session_multidoc_test.go`）：

1. **一次说清**：注入"服务不可达 + 无会话"，报错里**同时**出现 `tt debug serve` 与
   `tt debug start <作业> --env <环境>`（改回"只报第一个缺失"这条必红 —— 反证跑过）；
2. 体检全就绪 → 清单四项都 ok、`preflightExit` 返 nil（退 0）；
3. 目录不可写（假 `Writable` 注入）→ 退 2，提示里含「完整性/权限」与 `icacls`；
4. 体检**只检查**：临时目录前后文件集一致、会话清单只读一次（列 ≠ 起）；
5. **A31**：两段 JSON 也能读出会话（改回单次 `Unmarshal` 这条必红 —— 反证跑过）；
   解析失败 ⇒ `unknown`：`Missing` 不增、`preflightExit` 返 nil、`rec start` 不被拦，
   但人读态必须打印"查不了（原因）"。
