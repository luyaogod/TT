# internal/cli/fep — `tt fep` 命令组

T100 前端协议线（读写运行中的作业界面）的命令面。内核在
[`../../fep`](../../fep/README.md)：帧编解码、AUI 模型、字段名解析、语义动作都在那边，
**不碰窗口**；本组负责把「会话登记 / 录制件」这类输入交进去，把结果打到输出出口。

录制件的起停、回放与维护在 [`tt recording`](../recording)；会话的建/列/看/收在
[`tt session`](../session)。本组只留**前端协议**本身：动作、界面快照、等待、源码、作业、无头前端。

## 命令

| 命令 | 作用 |
|---|---|
| `tt fep do …` | 往前端发动作：`act`（点击）/ `type`（输入）/ `key`（按键）/ `cell` / `pick` / `lookup` / `role` |
| `tt fep ui …` | 读界面：`snapshot` / `tree` / `find` / `roles` / `state` / `focus`（都从录制件的字节折出来，离线可用） |
| `tt fep wait …` | 等运行中的作业变化：窗口出现、字段变值、文本出现 |
| `tt fep src …` | 读本地源码镜像：开窗程序的交互语义、字段类型与必填 |
| `tt fep job …` | 作业级：从源码推导操作计划，按计划闭环执行（每步现读状态、判可用、验效果） |
| `tt fep fe …` | 无头前端：我们自己当 GDC（`listen` / `status` / `send`），绕开 GDC 与注入口 |

`--format text` 出人读态；不加（或 `--format json`）出 JSON —— 本组一直如此
（`printJSONOrText`）。**`--json` 下 stdout 恰好一个对象**，失败时也是它。

## 前置体检

`tt recording start` 在**起代理之前**跑这份体检（实现在本包 `rec_preflight.go`）。
真机上常见的情形是失败信息**逐层挤出来**：先"服务没起"，起了再"没有会话"，
随后又冒一条"切换会话：没有 SSH 会话" —— 使用者被迫来回试错凑前置。体检查的就是那四项：

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

会话清单的解析**认多段 JSON**：`tt debug status --json` 可能连打几段（状态段 +
`{"ok":true,"sessions":[…]}`），解析器逐段 `Decode` 到 `io.EOF`，取**带 `sessions` 数组**
的那一段（段数不是契约：两段/三段/空行/CRLF 都认）；一段都读不懂才是 `unknown`。
把它改回单次 `Unmarshal`，`TestParseDebugSessionsReadsTwoDocuments` 会红。

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

缺几项就列几项，每条都带**可直接复制**的命令（已知的 `--prog` / `--env` 会填进去；
没给时留 `<作业>` / `<环境>` 占位符）：

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
`Sessions` 的生产实现走 [`../../debugclient`](../../debugclient)（调试面 REST 的唯一客户端）；
本包**不自己拼** HTTP 请求。

## 退出码

**唯一权威是 [`../README.md`](../README.md) 的那张表**，本包只做实现（`fep.go` 的常量）：

```
0 成功 / 1 运行失败（IO·网络·环境）/ 2 用法错 / 3 无会话 / 4 会话状态不符
5 目标不存在 / 6 环境体检不过（权限不一致）
```

缺前置一律退 **2**（用法/环境不满足）—— **不新造码**，也**不**用 6：6 是"体检跑过了但权限
不一致"那类结论，而缺 `tt debug serve` 属于环境没满足。
清单打出去之后再失败时，preflight 用 `AlreadyReported()` 的记号让根命令跳过 `reportError`，
否则 `--json` 下 stdout 上会出现两个对象（与 `tt form verify` 同一个手法）。

## 判据

```bash
go test ./internal/cli/fep/ -count=1     # 含：一次说清、体检全就绪/不可写、体检不落盘、多段 JSON
go vet ./internal/cli/fep/
```

钉住的几条（`rec_preflight_test.go` / `rec_start_missing_test.go` /
`rec_preflight_unknown_test.go` / `rec_session_multidoc_test.go` /
`no_shellout_test.go`）：

1. **一次说清**：注入"服务不可达 + 无会话"，报错里**同时**出现 `tt debug serve` 与
   `tt debug start <作业> --env <环境>`；
2. 体检全就绪 → 清单四项都 ok、`preflightExit` 返 nil（退 0）；
3. 目录不可写（假 `Writable` 注入）→ 退 2，提示里含「完整性/权限」与 `icacls`；
4. 体检**只检查**：临时目录前后文件集一致、会话清单只读一次（列 ≠ 起）；
5. **A31**：多段 JSON 也能读出会话；解析失败 ⇒ `unknown`（`Missing` 不增、
   `preflightExit` 返 nil），且人读态必须打印"查不了（原因）"；
6. **本包不许再起子进程调调试面**（`TestFepDoesNotShellOutToDebug`）：出现 `runSelf(` /
   `"os/exec"` 即红 —— 这类回退能编译、能跑，只在多会话或码不一致时才露出来。

## 细节去哪

- 内核（帧、AUI、语义动作）：[`../../fep/README.md`](../../fep/README.md)
- 录制件的起停与回放：[`../recording`](../recording)、[`../../cli/fep/rec_start.go`](./rec_start.go)
- 会话的建/列/看/收：[`../session`](../session)
- 调试面客户端：[`../../debugclient`](../../debugclient)
