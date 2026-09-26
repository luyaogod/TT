# 在 tt 上跑测试

**这份文档是"测试"这件事的单一来源**：跑什么、怎么配真环境、哪些开关、怎么回读结果。
`BUILD.md` 只管构建与打包，各目录的 `README.md` 只管那个目录**内部**的判据。

## 1. 默认档：不带任何环境变量

```bash
go test ./... -count=1
```

**这条命令在任何一台机器上都必须全绿**，包括没有语料、没有引擎、没有 git、没有
用户配置目录的机器。**一条依赖本机配置的测试会时好时坏，那比不测更糟** —— 那是这个
仓库对测试的第一条判断（原话在 `internal/dev/cli/corpus_test.go` 顶部）。

跑不了全量的东西一律推到显式开关后面，并且**必须打印为什么跳过**。
想验证"这次到底跑了多少、跳了多少"：

```bash
go test ./... -count=1 -v 2>&1 | grep -E '^--- (PASS|SKIP)' | sort | uniq -c
```

> 注意这条 grep **只数顶层测试**：子测试那行带前导空格，匹配不上。
> 想连子测试一起数，把模式放宽成 `'^ *--- (PASS|SKIP)'`。

## 2. 四层 + 编译期

| 层 | 跑一次多少钱 | 干净克隆上跑得起来吗 | 红了是什么意思 | 入口 |
|---|---|---|---|---|
| **L0 编译期** | `go build` ~10s；前端 `npm run build` ~30s | 跑得起来（前端需先 `npm ci`） | 改坏了 | `go build` / `npm run build` |
| **L1 纯逻辑** | 毫秒级 | 跑得起来 | **只可能是改坏了** | `go test ./...` |
| **L2 本地资源** | 秒级 | 跑得起来 | **只可能是改坏了** | 同上 |
| **L3 真语料** | 9–11 分 / 17 分 | 跑不起来（会 skip） | 改坏了 **或** 语料在两次运行之间变了 —— 要人分诊 | `TDEV_DEEP=1` / `TTZS_DEEP=1` |
| **L4 真机 E2E** | 分钟级到十几分钟 | 跑不起来 | 改坏了 **或** 环境不对 —— 要人分诊 | `TTZS_E2E=1` + 真引擎 / 真 SSH |

**默认档 = L0 + L1 + L2。** L2 的"本地资源"指：`t.TempDir()`、真 SQLite、
`httptest`、`127.0.0.1:0` 回读、临时注册表键、合成 `.tzc` 包 —— 都不依赖外部世界。

**L3/L4 永远靠环境变量开关 + 显式 `-timeout 30m`，不靠调大默认超时。**
理由被实测过：杀进程来自 **go 命令**（默认 `-timeout=10m`），
"测试二进制内部改 `-test.timeout` 是拦不住的"（`internal/dev/cli/corpus_test.go`）。

## 3. 命令表

```bash
go build -o tt.exe .                     # 后端（前端未构建也能过）
go test ./... -count=1                   # 默认档；深档语料回归默认跳过（约 1 分钟）
cd web && npm run check:app              # 前端三项检查：fgltokens / fgloutline / store
cd web && npm run build                  # 含 tsc --noEmit（check:app 不做类型检查）
./tt.exe dev tzc selftest                # .tzc 的内置对抗用例，不需要真实语料
./tt.exe dev tzs doctor                  # .tzs 引擎环境自检（含语料根一项）
```

包数与测试文件数会漂移，别写死在文档里：

```bash
go list ./... | wc -l                                                  # 包数
go list -f '{{if or .TestGoFiles .XTestGoFiles}}T{{end}}' ./... | grep -c T   # 带测试的包数
```

**没有 CI、没有 Makefile、没有 linter** —— 上面这一串加上 `AGENTS.md §7` 的按改动面
矩阵，就是全部的闸门。

## 4. 开关总表

全是**环境变量**，设了就生效，不设就走缺省。

| 开关 | 打开什么 | 代价 |
|---|---|---|
| `TDEV_DEEP=1` | `.tzc` 全语料回归（export+verify / apply 仿真） | 9–11 分钟 |
| `TTZS_DEEP=1` | `.tzs` 全语料回归 | 17 分钟 |
| `TTZS_VALIDATE=all` | 把 `validate` 从"每轴一个样本"扩到全部包 | 再加十几分钟 |
| `TTZS_E2E=1` | 真引擎测试（要配 `TTZS_EXE`、`TTZS_WS`，可选 `TTZS_INSTALL`） | 分钟级 |
| `TTZS_PKG=<真 .tzs 的绝对路径>` | 逻辑键寻址 / 任务级动词那两条 | 秒级 |
| `TTZS_CORPUS_LIMIT=N` | 只跑前 N 个包（冒烟） | 秒级 |

**为什么深档必须显式**：这些用例对上百个真实包各跑一遍，而 `go test` 默认超时 10 分钟
—— 表现为**随机器负载时好时坏的假失败**（同一份代码有时 530s 通过、有时 660s 被杀）。
那种假失败会污染"全绿"这条验收证据。

## 5. 真环境怎么配

要连真引擎 / 真语料 / 真工作区时，**四样机器路径**可以写进仓库根的
`config.local.json`（`.gitignore` 已经忽略它），也可以照旧用环境变量：

```json
{
  "corpusRoot":  "D:\\t100_wrok_dir",
  "engineExe":   "D:\\...\\engine\\out\\tzs-server.exe",
  "workspace":   "D:\\t100_wrok_dir\\<模块>",
  "designerDir": "D:\\APPS\\T100设计器_1.0.0.251_免安装"
}
```

| 字段 | 环境变量 | 缺省 |
|---|---|---|
| `corpusRoot` | `TDEV_CORPUS` / `TTZS_CORPUS`（**谁先设谁说话**） | `D:\t100_wrok_dir` |
| `engineExe` | `TTZS_EXE` | 无 |
| `workspace` | `TTZS_WS` | 无 |
| `designerDir` | `TTZS_INSTALL` | 无（用随包分发的那份） |

**优先级：环境变量 > `config.local.json` > 内置缺省。** 解析只有一处实现：
`internal/testenv`（它**不带 `testing`**，所以连生产代码也能用它问语料根）。

**覆盖项设了却指不到目录 → 返回空，不落回缺省。** 这是刻意的：静默换一份别的语料去跑，
会留下"我以为跑的是这份"这种错误结论。

**这份配置只放机器路径，不放口令。** 口令在 `config.json` 的 `hosts.sshs[].db` 里，
那是**凭据**（`AGENTS.md §9`），是另一件事。

配置写对了吗：

```bash
./tt.exe dev tzs doctor     # 末项「语料根」会报出它是怎么定下来的、卡在哪一处
```

## 6. 两条硬纪律

都是踩出来的事故（2026-09-26）：

1. **先在副本上跑。** 缺省语料根 `D:\t100_wrok_dir` 是**真实客户目录**，而回归会
   **在源包旁边**写 `_tdev_*` / `_tt_dry_*` 临时包。用副本跑：
   `TTZS_CORPUS=%TEMP%\ttws`。要单独证明真语料一个字节没动，跑 `TestCorpusPin`
   且**不设** `TTZS_CORPUS`。
2. **整轮跑的时候不要重建引擎。** `engine/build.sh` 会覆盖 `out/*.exe`，而每个包都要
   spawn `RoundTrip.exe` —— 撞上重写那一瞬间会得到 `The system cannot find the file
   specified`，表现为两三条**假红**，报告上看不出是构建造成的。

## 7. 跳过是被盯着的

跳过的测试与通过的测试在报告里长得一样无害。所以每一处 `t.Skip` / `t.Skipf` 都登记在
**跳过台账**里（`internal/testkit/skipledger_test.go`）：

- 台账与代码**严格对上**：多一处、少一处、**文案改了**都红
- 每条带层标记（L3 / L3-缺件 / L4 / L0-环境 / 无对象 / 前提）
- L3/L4 的条目要在理由里**点名开关名** —— 人看到 skip 就要能知道怎么打开

数跳过的正确方式（**别用 `grep "t.Skip" | wc -l`** —— 那按行数，会把注释与文案里的
提及也算进去）：

```bash
grep -roE 't\.Skipf?\(' --include=*.go internal/ | wc -l    # 真实调用点
```

## 8. 细节去哪

| 想知道 | 去哪 |
|---|---|
| **某个目录**该怎么验、覆盖了什么、故意不覆盖什么 | 该目录 `README.md` 的「判据」节 |
| 改哪块代码要跑到哪一步、改完的定义 | `AGENTS.md §7` |
| 从源码构建、依赖、打包 | `BUILD.md` |
| 跨包的测试辅助（语料发现 / 抓 stdout / 重置 flag / 跳过台账） | `internal/testkit/README.md` |
| 本机测试配置（`config.local.json` 怎么解析、几个字段） | `internal/testenv/README.md` |
| 语料的**遍历**规则（哪些文件算语料、跳过哪些草稿产物） | `internal/dev/testutil/README.md` |
| 语料清单的固定（`engine/corpus.manifest`） | `engine/BUILD.md` |
