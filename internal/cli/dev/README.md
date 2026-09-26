# internal/cli/dev — 把 `tt dev` 接进根命令

一层很薄的适配：让 `tt dev …` 出现在统一的命令树里，同时**不改变**这条线自己的参数写法与
退出码契约。全部实现就是两个文件。

## 为什么是适配层而不是 cobra 树

| 事 | 做法 | 理由 |
|---|---|---|
| 参数解析 | `DisableFlagParsing: true`，原样交给 [../../dev/cli/README.md](../../dev/cli/README.md) | 这条线的 flag 解析是**位置无关**的（`-o` / `--json` 可以出现在位置参数之后），标准库的解析器做不到，cobra 也做不到 |
| 退出码 | 把 `Run()` 返回的码包成 `exitCode` 往上传 | 0/2/3/4/5 是这条线的既有语义，必须原样透出 |
| 帮助 | `devLong` = 一句"在 tt 里的调用方式" + 这条线自己的 `Usage` | 帮助只有一份，不抄第二份 |

`Run()` 返回非 0 时，根命令不再自己打一遍错误 —— 这条线自己已经把失败写成了它约定的形状
（stdout 上的 JSON 对象，或 stderr 上的一行）。否则 stdout 上会出现两个对象。

## 根开关的翻译（`takeRootFlags`）

根命令的常驻开关写在自己的 `--help` 里，但本组不解析任何 flag，所以它们会原样落进这条线的
解析器 —— 而那边不认识，结果是"未知子命令 `--config`"直接退 2。于是这里做一次翻译：

| 根开关 | 翻译成 |
|---|---|
| `--config <路径>`（或 `--config=<路径>`） | 环境变量 `TT_CONFIG` —— 这条线本来就读它，不必为一个开关开新接口 |
| `--json` | 攒到参数**末尾**再放（见下） |
| `--format json` | 同上 |
| `--format table` / `text` | 丢掉（人读文本就是这条线的缺省形态） |
| `--format csv` | **明确拒绝** —— 这条线没有 CSV，"要 CSV 却拿到 JSON"是最该避免的静默走偏 |
| `--env` / `--conn` | **明确拒绝**，并点名它属于 `tt dict` / `tt debug` 那条线；本线用 `--workspace` 或配置里的工作区 |

**翻译出来的开关必须挂到参数末尾。** 就地保留的话，`tt --json dev tzc selftest` 里的 `--json`
本来就在最前面，这条线会把 `args[0]` 当成子命令名 —— 报"未知子命令 `--json`"。
两种写法的结果是同一个，所以不实例化位置的区别。

## 判据

```bash
go test ./internal/cli -run TestRootHelp      # 根帮助与 README 的漂移防线（在根包）
go build -o tt.exe . && ./tt.exe dev tzc selftest
./tt.exe --json dev tzc selftest               # 根开关翻译：应退 0，且 stdout 只有一个对象
./tt.exe dev tzc selftest --format csv         # 应被明确拒绝（不是静默走偏）
```

## 细节去哪

- 这条线自己的命令面、退出码、管线顺序 → [../../dev/cli/README.md](../../dev/cli/README.md)
- 用法与示例 → [`skills/tt-dev-tzc/SKILL.md`](../../../skills/tt-dev-tzc/SKILL.md)
