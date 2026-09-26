# AGENTS.md

TT（`tt`）是面向 Agent 的 T100 / Genero BDL 开发工具：**一个 Go 二进制 + 一套 React 工作台 + 一个 C# 引擎**
（引擎反射驱动 T100 设计器自己的程序集，不实现它的私有格式）。

本文件是**在仓库里改代码**的入口：怎么跑、改哪块之前先读什么、什么不能碰、怎么证明没改坏。

> **本文不重复别处的材料**，只做三件事：路由、边界、验证。凡是别处已经写清的，这里只给指路，
> 不抄第二份 —— 抄了就会漂移。

| 想知道 | 去哪 |
|---|---|
| 装 tt / 构建 tt / 配置文件放哪 | [README.md](README.md) |
| **某个目录内部**的设计、契约、不变量与实测数据 | 该目录的 `README.md`（每目录一份，外层讲关系、内层讲细节）；总索引见 [docs/README.md](docs/README.md) |
| 怎么**用**某个命令 | [skills/](skills/) 下对应的 `SKILL.md` |
| `.tzs` 引擎（C#） | [engine/BUILD.md](engine/BUILD.md)、[engine/SPEC.md](engine/SPEC.md)、[engine/HANDOFF.md](engine/HANDOFF.md) |
| 工具面评测装置 | [tools/eval/README.md](tools/eval/README.md) + [docs/eval-baseline.md](docs/eval-baseline.md) |

---

## 1. 一张图

| 目录 | 是什么 | 语言 / 工具链 | 怎么单独验 |
|---|---|---|---|
| 根 | `main.go` 只做 `//go:embed all:web/dist` → `cli.Execute` | Go 1.26.5+，**无 vendor**，直接依赖 8 个 | `go build -o tt.exe .` |
| `internal/` | 全部 Go 侧能力。四个命令组：`cli/`（cobra 树）· `debug/` · `dev/`（`.tzc` 管线 + `tzs/` 引擎客户端）· `dict/` | Go | `go test ./...` |
| `internal/host` | **唯一的 SSH / 远程服务器层**（debug 与 dict 共用） | Go | `go test ./internal/host` |
| `web/` | `app/` 是唯一 SPA（调试工作台 + 统一设置页）；`shared/` 是主题变量、UI 基元、设置页布局件 | React 18 + Vite 6 + TS + Tailwind 4，npm workspaces | `cd web && npm run check:app` |
| `engine/` | `.tzs` 引擎（C#）。**不在 Go 构建链里**，单独构建 | `csc.exe`（**C# 5**）+ Git Bash | `cd engine && ./build.sh` |
| `engine/designer/` | 设计器的 28 个程序集，**入库、随仓库分发** | 二进制，别碰 | — |
| `skills/` | 五套 AI 技能，同时是**给人看的操作手册** | Markdown，`SKILL.md` 的目录名必须 == frontmatter `name` | — |
| `tools/eval/` | 派"只拿 SKILL 的干净执行者"做真事、再回读产物的评测装置 | Python 3 | `python tools/eval/grade.py` |

**没有的东西**（别去找，也别顺手加）：CI、Makefile、linter、formatter、`.editorconfig`、pre-commit 钩子、`vendor/`。
**没有任何自动门禁替你拦错** —— 下面第 7 节那条手工序列就是全部的闸门。

---

## 2. 命令

```bash
go build -o tt.exe .                     # 后端（前端未构建也能过，见下方 .gitkeep 那条）
go test ./...                            # 全量；21 个包，深档语料回归默认跳过（约 2–3 分钟）
cd web && npm run check:app              # 前端三项检查：fgltokens / fgloutline / store
cd web && npm run build                  # 含 tsc --noEmit（**check:app 不做类型检查**）
cd engine && ./build.sh                  # 只在改了 engine/ 时才跑！理由见第 4 节
./tt.exe dev tzc selftest                # .tzc 的 31 项内置对抗用例，不需要真实语料
./tt.exe dev tzs doctor                  # .tzs 引擎环境自检（引擎 exe / 设计器 / 工作区 / 管道名）
```

### 深度回归：显式开关，别顺手开

| 开关 | 跑什么 | 代价 |
|---|---|---|
| `TDEV_DEEP=1 go test ./... -timeout 30m` | `.tzc` 全语料 export+verify / apply 仿真 | 9–11 分钟 |
| `TTZS_DEEP=1 go test ./internal/dev/tzs/ -run TestCorpus -timeout 30m` | `.tzs` 全语料回归 | 17 分钟 |
| `TTZS_VALIDATE=all` | 把 `validate` 从"每轴一个样本"扩到全部包 | 再加十几分钟 |
| `TTZS_CORPUS_LIMIT=N` | 只跑前 N 个包（冒烟） | 秒级 |

**为什么必须显式**：这些用例对上百个真实包各跑一遍，`go test` 默认 `-timeout=10m`，表现为
**随机器负载时好时坏的假失败** —— "那种假失败比不跑更糟"（`internal/dev/cli/corpus_test.go` 顶部）。
改这个开关之前先读那里的注释。

### 跑语料的两条硬纪律（都是踩出来的事故，2026-09-26）

1. **先在副本上跑。** 缺省语料根 `D:\t100_wrok_dir` 是**真实客户目录**，而回归会**在源包旁边**写
   `_tdev_*` / `_tt_dry_*` 临时包。用 `TTZS_CORPUS=%TEMP%\ttws`（整份拷贝）跑。要单独证明真语料
   一个字节没动，就跑 `TestCorpusPin` 且**不设** `TTZS_CORPUS`。
2. **整轮在跑的时候不要重建引擎。** `engine/build.sh` 会覆盖 `out/*.exe`，而每个包都要 spawn
   `RoundTrip.exe` —— 撞上重写那一瞬间得到 `fork/exec … RoundTrip.exe: The system cannot find the
   file specified`，表现为两三条**假红**，报告上看不出是构建造成的。

语料相关的环境变量：`TDEV_CORPUS` / `TTZS_CORPUS`（**两个都认，谁先设谁说话**）、`TTZS_E2E=1`
（真机引擎测试，配 `TTZS_EXE` / `TTZS_WS` / `TTZS_INSTALL`）。语料根的解析只有一处
（`internal/dev/testutil/corpus.go`，两条管线共用）—— 别再写第二份：漂移的结果是
**"0 个包全部通过"这种最坏的假绿**。注意覆盖变量**设了但指不到目录时返回空**，不会回落到缺省。

---

## 3. 改哪块之前先读哪份

| 你要改 | 先读（**必读**） |
|---|---|
| 任何东西 | 该目录的 `README.md`。**契约与"为什么"现在住在被改的那一层旁边**，不再有一份集中文档 |
| `internal/config/**` | [internal/config/README.md](internal/config/README.md)（位置解析、迁移、唯一写入口、缓存清单） |
| `internal/host` `dbconfig` `erpdb` `entdir` `output` | 各自的 README；那条链的形状见 [DESIGN_DOC.md](DESIGN_DOC.md) |
| `internal/web` 或 `web/**` | [internal/web/README.md](internal/web/README.md) + [web/shared/README.md](web/shared/README.md)（设计系统规则） |
| `internal/debug/**` | [internal/debug/README.md](internal/debug/README.md) |
| `internal/dev/**`（`.tzc` 管线） | [internal/dev/README.md](internal/dev/README.md)（**必读**：公理、围栏协议、三道闸门、红线 R1–R7、编号体系）+ 你要改的那个子包的 README |
| `internal/dev/tzs/**`（引擎客户端） | [internal/dev/tzs/README.md](internal/dev/tzs/README.md) + `engine/SPEC.md`（冻结契约） |
| `internal/cli/dict/**` `internal/dict/**` | [internal/dict/README.md](internal/dict/README.md)（数据族与数据源） |
| `engine/**`（C#） | [engine/README.md](engine/README.md) → `engine/BUILD.md`（为什么必须单独构建）→ `SPEC.md` → `HANDOFF.md` → `TASKS.md`（**阶段记录，不是现状**，看它怎么跑会被带到沟里） |
| `build_*.bat` / `installer/` | [BUILD.md](BUILD.md) + [installer/README.md](installer/README.md)、[tools/README.md](tools/README.md) |
| `skills/**` | 它是**对外文档**：改了工具面就要重跑 `tools/eval`（见第 7 节） |

---

## 4. 硬边界

1. **不实现设计器的私有格式。** `.tzc` 靠设计器发行物反推；`.tzs` 更彻底 —— 反射调用设计器自己的
   程序集，格式由它自己算。任何"我来解析/我来拼这段 XML"的想法都违反这条。
2. **设计器程序集入库是有意的**（`engine/designer/`，28 个 dll，约 10.5 MB）。它们保证"同一份 tt
   在任何机器上跑的是同一版设计器"。换版本 = 换文件 + 提交，**是一次可 review 的动作**；不要
   "清理"它们，也不要让它们走上行尾转换（`.gitattributes` 已明示 `binary`）。
3. **引擎只在真的改了 `engine/` 时才重建。** 守护进程的管道名 = `hash(工作区)` + **本程序集 MVID 前 8 位**，
   所以每次重编都会让**上一个构建起的守护进程再也停不掉**（新名字没人监听）。这是"一个与 tt 无关的
   构建步骤造成用户可见后果"的典型。清理靠 `tt dev tzs reap --yes`；没被记录过 pid 的守护进程
   `stop`/`reap` 都够不到，只能手工 `taskkill`。
4. **生成的产物不许手工编辑**：
   - `web/dist/*` 除 `.gitkeep`；**`.gitkeep` 是承重墙** —— `main.go` 的 `//go:embed all:web/dist`
     靠它成立，删了 `go build` 直接失败；vite 的 `emptyOutDir` 每次构建会删它，由 `postbuild` 还原。
   - `engine/out/`（里面有十几个探测程序 `Probe`/`RoundTrip`/`Test*`/`E2E`，**只供开发，不进包**）；
     它也不能落在 `dist/` 下（`build_portable.bat` 第一件事就是删 `dist`）。
   - `web/package-lock.json` 是入库的；`web/app/scripts/.tmp-*.mjs` 是中间产物。
5. **不许进包的东西**：本机 `config.json`（含明文口令）、`erp_data.db`（客户表字典/schema/企业码）、
   引擎探测程序、`tzs-cli.exe`（对外只有 `tt` 一个入口 —— 两个客户端就是两套命令面与两套退出码）。
   便携包的文件是**按名字手列**在 `build_portable.bat` 里的（引擎那三个 + `README.md` + `skills/`）；
   MSI 侧不用改 `tt.wxs`（`heat.exe` 采集），但 `build_portable.bat` 要手工加。
6. **`tt dev tzc unlock --yes` 必须由人确认。** 解开框架**不可逆**，等于主动放弃"跟随原厂样板自动
   重产代码"的能力。**AI 不得自主触发**；没授权时加 `--yes` 也拒。
7. **别在真客户目录上写**。工作区之外的写回、`out` 指向源包（会覆盖原始素材）、语料回归写在源包旁边
   —— 前两条**代码里已有闸门**，第三条靠你按第 2 节的纪律做。
8. **`config.json` 是唯一配置文件**，含明文 SSH / DB 口令（新建权限 0600）。不要新增第二个配置文件、
   不要新增第二份位置解析（历史上那两份逐行相同、靠注释"改动请两边同步"，合并就是为了删掉它们）。

---

## 5. 单一来源清单（别抄第二份）

这个仓库反复出现的坏味道是"同一件事有第二份实现，然后在某次改动里静默过期"。以下都是**故意只留一份**的：

| 事 | 唯一入口 | 抄第二份的后果 |
|---|---|---|
| 写配置 | `config.Edit` / `EditSection`（`internal/config/cfgfile.go`） | 未知顶层/节内键被悄悄抹掉 |
| 环境解析（name → activeEnv → 首条） | `config.Hosts.Resolve` / `Root.ResolveForTool` | 历史上抄过五份，错误文案与边界全不一样 |
| 账号取法 | `entdir` + `gzou_t` | 别自己取 `accounts[0]` |
| 查询输出 | `output.Emit`（唯一的输出出口） | 曾经有四套 JSON 发射器、三套 CSV 写入器 |
| 打码 | `config.RedactSecrets` / `RedactValue`（按键名，不按 tag） | 未知节的未知键漏出明文口令 |
| 工作区落盘 | `store.AtomicWrite`（同目录临时文件 → Rename） | 红线 R6：出现"半个文件"或"原文件已丢" |
| tzs 动词参数 | **引擎的 manifest**（`internal/dev/tzs/manifest.go`） | Go 侧抄一份 `map[string][]Param`，引擎一改就静默过期 |
| tzs 动词数 | 只在 `tzsUsage` 常量与 `skills/tt-dev-tzs` 的元数据里声明（`TestDocVerbCountsAgree` 盯着） | 散在多处就会互相漂移；README 一律写"数见 `tt dev tzs --help`" |
| tzs 管道名 | **问引擎**（`--pipe-name --workspace`） | 自己算 → 静默失败，看起来像"冷启动 60 秒没就绪" |
| 语料根发现 | `internal/dev/testutil/corpus.go` | 一边认 `TDEV_CORPUS`、一边只认 `TTZS_CORPUS` → 假绿 |
| 危险字符/SQL | `safesql` + `shQuote`/`bashLC`；**SQL 一律走 stdin** | `--zone "36; id"` 曾在远端执行任意命令 |

**分层规则**：三个命令组（`cli/debug`、`cli/dev`、`cli/dict`）**互不 import**，共享上下文在叶子包
`internal/cli/common`；往上通信靠**注入**（`common.MetaProvider` / `common.WebFS`），不是让叶子反向 import。
`internal/dev/tzs` 不 import `internal/dev/cli` —— 那边 import 它。

---

## 6. 会静默咬人的地方

| 症状 | 原因 / 规矩 |
|---|---|
| 退出码按"通用含义"读 | **每个子命令一条线、语义各不同**：`tzc` 0/2/3/4/5；`tzs` 0/1/2/4/5（**没有 3**）；`dict` 0/1/2/3（3 = 缺表，是错误不是提示）；`debug` 0/1。`tt --help` 里有整张表 |
| 脚本只看 `stderr` 非空判失败 | JSON 模式下**错误信封走 stdout**（`{ok:false,code,error,hint,exitCode}`）；`code` 是**冻结契约**，能加不能改名 |
| 失败信封"形状不对" | 两条线**故意不同形**：`tzc` 是 `{"ok":false,"exit_code":…}`（snake）；`tzs` 是合成一个**帧**（`{id,ok,result,error,ms}`）；`dict`/`debug` 走 `output` 信封（camelCase） |
| `grep -v '^# '` 把数据行切掉了 | 注释头**井号后必须有一个空格**（值本身以 `#` 开头是常事，如颜色 `#FF0000`） |
| `tt dict nope` 退 0 却什么都没干 | cobra 组命令**没有 `RunE`** 时未知子命令返回成功。`dict`/`debug` 都加了未知子命令兜底，新增命令组照做 |
| `tt dev … --csv` 被拒 | `tt dev` 是 `DisableFlagParsing` 的独立 CLI，不是 cobra 树；`--csv`/`--env` 被**故意拒绝**并给出正确指引（"要 CSV 拿到 JSON"是最该避免的静默走偏）；`--config` 靠转成 `TT_CONFIG` 转发 |
| 同一程序的第二个包打不开 | `open` 的 key 是 `程序名|Form`，**不含路径**；新包与源包同名，必须先 `close` 再 `open`（`E_KEY_IN_USE`，退 4）。**别读 `TzpManager.Current`**，一律按 key 寻址 |
| 无界面驱动卡住 90 秒 | 某些操作会弹**没有消息泵的模态框**（如缺基础资料、只拷 `mta/` 的工作区）。看门狗跑不掉这类；只能整份工作区 |
| `git diff --stat` 行数是改动的几百倍 | **行尾被翻了**。仓库行尾是**混合**的（Go/`docs/`/`README.md` 是 LF，部分 C# 与 `engine/SPEC.md`、`engine/TASKS.md` 是 CRLF）。`Edit` 工具保留原行尾，**`sed -i` 与 Python 文本模式不会**。判据：`git diff --stat` vs `git diff --ignore-cr-at-eol --stat`，并比对 `file -b` |
| 前端"构建绿但缺样式" | 共享层只被 `../shared` 的 `@source` 扫到；`web/app/src/index.css` 里那条 `@source "../../shared"` 动一下就是这个结果 |
| 改了 `tokens.css` 主题后卡在亮色 | 暗色是 **class 驱动**（`@custom-variant dark`），导入方必须同时有首屏防闪脚本与 `applyDark`；存储键 `tt.theme` 在 `index.html` 与 `shared/theme.ts` 里各写了一次 |
| `TestCorpusPin` 红 | pin（`engine/corpus.manifest`）与实际语料**逐条**比对，磁盘多出三个从未入册的 `*_test.tzs` 就红；它归 `TTZS_DEEP`，默认不跑所以平时看不见。重钉用 `engine/make-manifest.sh [root]`（**默认根是真客户目录**，别在已被改动的语料上重钉） |
| 打包脚本报错退出 | 引擎那三个文件**按名字**采集，缺一个就报错；`engine/out/` 里的探测程序绝不能 xcopy 整个目录 |

---

## 7. 做完的定义

**没有 CI，所以每个改动都要自己跑完这条序列，并把结果如实报出来**（跑不了就说跑不了）：

| 改了什么 | 至少跑到 |
|---|---|
| 任何 Go 代码 | `go build -o tt.exe .` + `go test ./...` |
| `internal/dev/**`、`internal/pkgfile|tapfile|fence|fgl/**` | 上一行 + 相关包 `-count=1`；动了写回再考虑 `tt dev tzc selftest` |
| `web/**` | `cd web && npm run check:app` + `npm run build`（类型检查在构建里） |
| `engine/**` | `cd engine && ./build.sh` + `tt dev tzs doctor`，**再**考虑 `TTZS_DEEP`（先满足第 2 节两条纪律） |
| `skills/**` 或**任何工具面**（动词、参数、退出码、错误文案） | 重跑工具面评测：`python tools/eval/setup.py --src <语料> --exe tt.exe` → 派只读 SKILL 的执行者 → `python tools/eval/grade.py --run %TEMP%\ttrun` |
| 打包/发行 | README 的构建节 + `build_portable.bat` / `build_msi.bat`（版本号硬编码在**四处**，一起改） |

**判据是回读，不是自述。** 这是评测装置立的规矩，也适用于你自己的改动：说"改好了"之前，
用一条**能被别人重放**的命令证明它 —— `sha256sum` 比字节、`open` 回读内容、测试输出、`git diff --stat`。
"看起来对"不算证据；文里凡是声称安全的地方都要给判据与复现命令，而不是形容词。

**语料是关卡手段，不是迭代手段。** 如果 corpus_test 顶部的判据是"基线相对"而不是"等于零"，
别把它改成硬编码零 —— 实测有三个包在**没人动过**的时候就报 stale。

---

## 8. 提交与文档

- **提交信息**：`<范围>: <一句话说清改了什么、代价是什么>`，范围用 `tzs` / `tzc` / `eval` / `docs` / `README` / `config` 这类词；
  正文写**为什么**、**实测数字**、**被推翻的旧结论**、**踩到的坑**。这个仓库的历史是它的第二份文档，值得对齐。
- **文档与代码同一个提交**。文档责任是分区明确的：**根 README 管装与配置，BUILD 管构建，DESIGN_DOC 管整体结构
  与关系，AGENTS 管改代码的路由与边界，各目录 README 管该目录内部的细节与契约，skills 管怎么用**。
  改了行为就去改**被改的那一层旁边**那份文档；外层只给结论并引用内层，**别把内容抄第二份**。
- **`skills/**` 与源代码同级重要**：它是对外文档，也是评测里执行者唯一能读的东西。改它的措辞要当成
  改 API 文档来对待（"说明书写得好不好"正是评测要测的东西）。
- **中文**：面向人的输出（帮助、错误、提示、文档、注释）一律简体中文；JSON 键用英文 lowerCamelCase
  （信封/元信息/错误面），**字典行数据的中文键是有意的**（对齐 T100 列名），两个方向都不要"顺手统一"。

---

## 9. 凭据与安全

- `config.json` 里有**明文** SSH / DB 口令。**不要**把它粘进对话、提交、issue、测试夹具；
  测试一律用假口令（仓库里现成的是 `pw-dev` / `SECRET` 这类）。
- **`GET /api/hosts` 返回明文口令**（设置页要能就地编辑与探测）。**任何 dump / 粘贴 `/api/hosts` 响应的行为
  都是泄密**，包括贴进聊天窗口和 issue。
- 输出侧打码只有一个实现（`internal/config/redact.go`，按**键名** `password`/`passwd`/`pwd` 判断）。
  新增"把配置吐出来"的命令，走它。
- `erp_data.db` 含客户字典与 schema，同样是数据而非代码：不进仓库、不进发行包。
- 远端命令拼接：`zone` 之类要过白名单（历史事故：`--zone "36; id"` 在远端执行了任意命令）；
  **SQL 一律走 stdin**，不拼进命令串。
- `tt install path` 只写 **HKCU 用户 PATH**，永不碰系统 PATH。
