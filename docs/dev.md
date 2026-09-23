# tt dev — T100 设计器 `.tzc` 包的安全编辑工具

给 AI 和工程师用的命令行工具：把「改 T100 设计器里的 4GL 客制」变成一条**可机械证明不会破坏**的管线。

```
tt dev tzc export <pkg.tzc> [-o <dir>] [--only <点名>...] [--json]
tt dev tzc status [<dir>] [--json]
tt dev tzc verify [<dir>] [--json] [--strict]
tt dev tzc apply  [<dir>] [-o <pkg.tzc>] [--dry-run] [--yes] [--json]
tt dev tzc unlock [<dir>] [--yes] [--json]
tt dev tzc rename [<dir>] <旧函数名> <新函数名> [--scope PUBLIC|PRIVATE] [--desc <描述>] [--json]
tt dev tzc newfn  [<dir>] --type FUNCTION|DIALOG|REPORT [--name <名>] [--scope …] [--json]
tt dev tzc selftest [--json]

tt dev tzs export <pkg.tzs> [-o <dir>] [--force] [--json]   # 表单包纯解压（只读参考）
tt dev tzs call <fn> [--<参数> <值>…]                        # 读写表单（由设计器自己的引擎算）
tt dev tzs fns | manifest | doctor | stop | reap            # 函数表 / 环境自检 / 守护进程
tt install skills [--to <dir>] [--force] [--json]   # 复制 exe 旁边的 skills/ 到 <当前目录>/skills
tt install path [--dry-run] [--json]                # 把 exe 目录加进用户 PATH（HKCU，免管理员）
```

**两条管线别用错**：`.tzc` 是**代码包**，走 `tzc export`（渲染围栏工作区，改完 `apply` 写回）；
`.tzs` 是**表单包**：`tzs export` 纯解压、只读；**要读写表单走 `tzs call`**，那是由设计器自己的代码算的。

**四个生命周期动词**（export/status/verify/apply）——写包只有 `apply`；
`unlock` 是**单向状态迁移**（框架解锁，只改 workspace）；`rename`/`newfn` 是**编辑辅助**
（把设计器弹窗 CLI 化，只改 workspace）。后三类产出的都是普通的工作区改动，仍须走 `apply` 才落盘。

`-o` 可省略：默认导出到 **`<包所在目录>/<程序名>-ws`**（身份后缀 `(c)`/`(s)` 会去掉）。
例如 `D:\pkg\capt110(c).tzc` → `D:\pkg\capt110-ws\`。
同目录下放多个包时各自建各自的工作区，不会互撞；默认目录已被占用时会明确报错并提示用 `-o` 换位置。

`<dir>` 也可以省略：**先 `cd` 进工作区，命令就不用再写目录**。

```bat
cd /d D:\pkg\capt110-ws
tt dev tzc status                 :: = tt dev tzc status D:\pkg\capt110-ws
tt dev tzc verify
tt dev tzc apply
tt dev tzc rename capt110_calc capt110_count   :: 省略 <dir> 时 rename 收 2 个位置参数
tt dev tzc newfn --type FUNCTION --name capt110_added
```

判定规则只有三条，没有隐藏行为：显式给了 `<dir>` 就用给的；没给且当前目录含
`prog.full.4gl` + `manifest.json`（即确实是工作区）就用当前目录；没给且当前目录不是工作区
就**明确报错退出 5**（提示 `cd` 或显式给出 `<dir>`），绝不猜别处的目录。
`rename` 的第一个位置参数若是个已存在的工作区目录，仍按 `<dir> <旧> <新>` 解释，
否则按 `<旧> <新>` + 当前目录解释。

四个动词，与三条设计公理一一对应，没有第五个：

| 公理 | 含义 |
|---|---|
| **A1 唯一编辑界面** | AI 只面对一个带围栏标注的 `.4gl` 文件，和人在设计器里看到的画面一致 |
| **A2 唯一写路径** | 任何修改都走 `apply`；不存在「单点手术」式命令 |
| **A3 不破坏是机械证明** | 围栏外逐字节比对 → 不变量 + FGL 解析 → 新包的装载模拟 |

## 为什么不是直接改 zip

- `.tzc` 是普通 zip，但里面的 `<prog>.tap` 才是**唯一被服务端消费**的设计文件；`.tgl` 是框架骨架、`.4gl` 是服务器 build 产物（**客户端从不读回**）。
- 设计器保存时会把整包**非原子重写**（`File.Delete` → `File.Create`，时间戳/压缩级别/ACL 全变），并且 `.tap` 的换行是**混合**的（元素间 CRLF、CDATA 内 LF）。所以任何「解析成 XML 再序列化」的做法都会让 diff 爆炸、甚至丢数据。
- tt dev 的对策：**CDATA 感知的字节级扫描器**，只重建被改的 CDATA 内部；其余字节（含属性顺序、引号风格、空白、未知条目、`ver`）逐字节透传。

## 包容器（zip）形态：照抄设计器，不是「写个合法 zip」

**合法不够，得像它。** 设计器写出来的 `.tzc` 有自己的 zip 形态，tt dev 写回时必须一模一样：

| 特征 | 设计器（.NET）的包 | 必须避免的写法 |
|---|---|---|
| general purpose flag | `0x0000`（**没有**数据描述符） | `0x0008`：bit 3 + 数据描述符 |
| CRC / 压缩后大小 / 原大小 | **写在局部头里** | 局部头填 0，真值只放数据描述符 |
| 局部头 extra | 原样保留（`UT-time` 9 B + `ux-infozip` 11 B） | 被替换成写入器自己的时间戳字段 |
| 中央目录 extra | 原样保留（`UT-time` 5 B + `ux-infozip` 11 B） | 同上 |

> **真机事故（S7 抓到的第一个不兼容）**：早期实现用 Go 标准库 `archive/zip` 的 `Writer` 重写整包。
> 它**无条件**置 bit 3 并补数据描述符，还会用自己的 9 字节 extra 顶掉原 extra。
> 结果：包在 tt dev 里读得回来（Go 看中央目录），拿到设计器里一打开就报
> **`Data descriptor signature not found`**。这正是「S7 真机验收不能省」的原因。

现在的写回是**字节级重建**（`internal/pkgfile/zipraw.go`）：局部头与中央目录记录都从原包逐字节拷贝，
只打补丁 —— 清 bit 3、写真实 CRC/大小、更新局部头偏移、必要时更新 Zip64 extra 里的 8 字节字段；
未改动的条目连**压缩数据都逐字节照抄**。于是：

- **零改动重建 = 与原包逐字节相同**（语料 166 个真实包：165 个逐字节相同，1 个是 TDev 早期产出的 bit3 包，被归一化成设计器形态；归一化之后是不动点）；
- **只改 `.tap` 时，差异只出现在该条目的 CRC/尺寸字段与其载荷、以及它之后条目的局部头偏移字段**（`.4gl`/`.tgl`/`ver` 连局部头和 extra 都和原来一样）；
- Zip64：原包用 `0xFFFFFFFF` 标记的字段保持标记，真值就地写回 8 字节字段（语料里有 302 个这样的条目）；不静默去掉 Zip64，也不在超 4 GiB 时乱写。


## 工作区

```
<dir>/
├── prog.full.4gl        # 唯一编辑文件：完整文档 + 围栏标注
├── manifest.json        # 人机共读索引（含每个 Region 的 editable 与**不可编辑原因**）
├── snapshot/
│   ├── index.json       # 原包条目清单（name/sha256/size/role）
│   └── entries/<条目名>  # 原包每个条目的逐字节拷贝（恢复源；apply 不从这里读）
├── .tdev/
│   ├── base.full.4gl    # 导出时 prog.full.4gl 的字节拷贝（gate1 的比对基线）
│   ├── base.sha256
│   ├── regions.json     # Region 表：名字/类型/权限/原因/meta/字节区间/tgl_tag
│   └── lock             # 防止两个 tt dev 进程同时操作
└── .git/                # export 时 init + 首次 commit；apply 成功 = 新 commit
```

`manifest.json` 里每个 Region 都带 `reason`（中文），例如
`"SEC 模式未开启（--allow-sec）"`、`"不在本次 --only 导出范围内"`、`"框架集合锚点区段强制只读"`。
AI 看到 `editable:false` 的同时就知道为什么，不必用试错法探测权限边界。

## 表单包（`.tzs`）：导出只读，读写走引擎

`.tzs`（表单包）/ `.tzv`（简易表单包）走的是**完全不同的入口**：

```powershell
tt dev tzs export "D:\pkg\aapp320(c).tzs"        # → D:\pkg\aapp320-unzip\（默认）
# 已纯解压（不做任何处理，也不产生工作区）：D:\pkg\aapp320-unzip
#   来源：D:\pkg\aapp320(c).tzs（10323 B，sha256=aca66b97015a）
#   文件：4 个，共 109807 B
#     aapp320.4fd    79272 B  sha256=e72f38bd8416
#     aapp320.tsd    30468 B  …
```

它做的事**只有一件**：把 zip 逐条目解压到目录里（`pkgfile.UnzipTo`）。因此：

| | `tt dev tzc export`（代码包） | `tt dev tzs export`（表单包） |
|---|---|---|
| 产物 | **工作区**：`prog.full.4gl`（带围栏）+ `manifest.json` + `snapshot/` + `.tdev/` + `.git/` | **就是一包文件**（`.tsd` / `.4fd` / `ver` …） |
| 特殊处理 | 合成 + 围栏渲染 + 权限判定 | **没有**：不解围栏、不校验必需条目、不解析 `ver` |
| 能不能改回来 | 改 `prog.full.4gl` → `verify` → `apply`（唯一写路径） | **不走 export**：文本改完塞回包这条路不存在，要用 `tzs call` 让设计器自己算 |
| 用途 | 改 4GL 客制 | **只读参考**：读表单结构、查字段定义、写文档 |

配套细节：

- 默认目录 `<包所在目录>/<程序名>-unzip`（身份后缀 `(c)`/`(s)` 去掉），和 `-ws` 区分开，
  一眼能看出「这是解压产物还是工作区」；`-o` 可改。
- 目标目录非空 → 拒绝（退出码 5），确认要覆盖加 `--force`（只覆盖同名文件，不删别的）。
- zip-slip 防护：条目名带 `../`、盘符、绝对路径、NUL → 整体拒绝（退出码 2），不会写到目标目录之外。
- 输入类型闸门：拿 `.tzc`/`.tzf`/`.tzx` 跑 `tzs export` 会被挡下并指回 `tt dev tzc export`。
- `--json` 给 `{ok, pkg, pkg_sha256, dir, entries[], count, bytes, readonly:true}`。

### 读写表单：`tt dev tzs call`

表单的写路径是引擎（`engine/`，一个 C# exe），它反射驱动**已安装的设计器自己的程序集**：
布局属性走设计器自己的 `XmlElement` 索引器，`.tsd` 由设计器自己从模型重算，验收用设计器
自己的校验器加 RoundTrip 不动点。**我们一个字节的 `.tzs` 格式都没实现。**

```powershell
tt dev tzs fns                                  # 49 个函数（由 --manifest 派生，不手写）
tt dev tzs fns add_field                        # 单个函数的参数表
tt dev tzs doctor                               # 引擎/设计器目录/工作区/管道名自检

tt dev tzs call open           --path "D:\pkg\aapp320(c).tzs"
tt dev tzs call find_component --handle h1 --query l_apcasite
tt dev tzs call nudge          --handle h1 --paths <path> --direction right --offset 1
tt dev tzs call validate       --handle h1
tt dev tzs call save           --handle h1 --out "D:\pkg\_ai.tzs"
tt dev tzs stop
```

需要配置两样（都是机器级依赖，没有合理缺省）：

- **设计器**：不用配。它的程序集**随包分发**（`tzs\designer\`），引擎默认从那儿加载 —— 所以同一份
  tt 在任何机器上跑的是同一版设计器，不需要谁去对齐配置。`TZSCLI_INSTALL` 只在开发时用
  （`engine/out/` 不是包，本地引擎与探测程序靠它指向机器上装的那份）。
- **用哪个工作区**：`tzs.workspace` 或 `TZSCLI_WS`。**这条没有缺省、也不回落** —— 引擎内置
  的默认工作区是一个真实客户目录，落上去等于拿别人的表单当草稿纸。三层都空时报错退出。

退出码沿用 tdev 那套：`0` 成功 / `2` 参数或环境不对（引擎的 `validation`、`not_found`）/
`4` 设计器拒绝（`designer`）/ `1` 引擎内部错 / `5` 传输或环境失败（含「包加载不完」）。

**请求一旦上线绝不重试**：协议无幂等键，而这些函数都在改设计器内存里的模型，重试是在赌
「上一次写进去了没有」。

> **红线**：`export` 的产物是**只读参考** —— 它就是一包文件，没有 manifest/围栏，不要手工
> 改完再塞回包，也不要拿它当 `tzc` 工作区去 `apply`。要改表单用 `tt dev tzs call`：
> 那是设计器自己的模型在算，改完设计器打得开。
>
> 旧版本这里写的是「tt dev 不写回（没有对应模型与验收样本）」。那句话的前提现在不成立了 ——
> `engine/` 就是设计器自己的代码。

## 围栏协议

围栏是**单行注释**，不改变 4GL 语义，且与 TGL 原生的 `{<point/>}`/`{<section>}` 标记区分开：

```
{//@tdev:begin section adzi999.main [READONLY deny="sec-off" src="s"]}
MAIN
  {//@tdev:begin point main.define_customerization [EDITABLE src="s" new="Y" order=""]}
  CALL ccl_init()
  {//@tdev:end point}
END MAIN
{//@tdev:end section}
```

规则：

1. `begin`/`end` 严格配对；允许 `section → point` 一层嵌套（真实 TGL 里自订点就住在区段内），深度 > 2 报错。
2. **围栏行本身属于「围栏外」**：AI 不许增删改围栏行（meta 由 tt dev 推导，不由作者填写）。
3. 自订定义点内部再分**结构行**（注释头 / `PUBLIC FUNCTION x(...)` / `END FUNCTION`）与**正文本体**；
   只有正文本体可改，结构行改动一律拦下（对齐设计器 `CanInsert` 只允许改正文本体的行为）。
4. 新增自订点的唯一方式：在 `[APPEND]` 锚点区段内追加一个完整的 point 围栏块。
5. 删除自订点 = 删除其整个围栏块，且**仅 `new="Y"` 可删**。
6. Region 之外的字节（文件头尾、区段之间的空行）不许增删改，逐字节比对。

## 三道闸门

| 闸门 | 检查 | 失败 |
|---|---|---|
| **gate1** | 围栏行、围栏外字节、只读 Region 内容、可编辑区内的结构行 —— **逐字节相等**；删除/追加的授权 | 退出码 3；权限类 → 4 |
| **gate2** | README §3.8 的不变量 I1–I15：区段配对/命名空间三层错配/`TglTag` 折叠/status 取值/UTF-8 无 BOM/`]]>`/ver/未知条目透传… | 退出码 3（warn 仅在 `--strict` 下失败） |
| **gate3** | 对**将要产出的新包**重跑合成 = 「设计器打得开吗」的判决；磁盘此刻仍未动 | 退出码 3 |

`apply` 的管线顺序即契约：

```
读 edited → parse_fenced → gate1 → 源包未变校验 → gate2 → split → 原子写 → git commit
                                                              ↑
                                            gate3（写之前，对内存里的新包）
```

任何一步失败，原包字节不变 —— 这不是目标，是 `atomic_write` + 管线顺序的必然结果。

**apply 到底会改哪些东西**（明确清单）：

| 对象 | 动作 |
|---|---|
| `.tzc` 包（`manifest.pkg.path` 记的路径） | **原地覆盖**（原子写：临时文件 + rename） |
| └ 未改动条目（`.4gl` / `.tgl` / `ver` / 未知条目） | **逐字节透传**，一个字节都不变（红线 R1/R2） |
| └ `.tap` | 只替换被改区域的 CDATA 区间 + 相应属性 |
| `.tdev/prev.tzc` | 覆盖前把**上一版包**整文件备份一份（可回滚） |
| 工作区 `.tdev/base.full.4gl`、`regions.json`、`manifest.json`、`snapshot/` | 刷新为新状态（基线前进一格） |
| 工作区 `.git` | 一次 commit（`tdev apply: <prog>`） |

> **迭代循环**：`改 → apply → 再改 → 再 apply` 连续多次是支持的。这依赖 apply 成功后刷新
> `manifest.pkg.sha256`；不刷新的话第二次 apply 会被「源包自 export 之后已被改动」误拒（退出码 5）。
> 该缺陷已修复，并由自检项「真机事故-迭代循环」钉住。

> **改动的条目会被重新压缩**：`.tap` 变了就得重压（Go 标准库 deflate），所以它的条目字节
> 与设计器的 SharpZipLib（level 3）不同；**其余条目（`.4gl` / `.tgl` / `ver` / 未知条目）
> 连压缩数据都逐字节照抄**，零改动重建更是与原包逐字节相同（见上节「包容器形态」）。
> 判据始终是**逐条目内容 sha256**（设计指南 §8 同口径）：`.4gl` / `.tgl` / `ver` 的条目内容永远不变。

## 报错怎么定位（行号 + 该行内容）

`apply` / `verify` 拒绝时，每条发现都会带上**出错位置**——出错文件、1-based 行号、该行内容：

```
apply 被拒绝（gate1 error=1 warn=0 info=1）：
  [error] gate1.readonly-region  只读 Region 内容被改动（写入被拒）
            位置：prog.full.4gl:5
            该行：#應用 a00 樣板自動產生(Version:3)  # 我加的字
            deny=section-locked
```

- 行号给的是**你编辑的那份** `prog.full.4gl`；删除类问题（「删掉了不该删的点」）给
  `.tdev/base.full.4gl:<行号>`（编辑后的文档里它已经不存在了）。
- 定位用**行尾等价**比较：编辑器把 CRLF 归一成 LF 不会把位置指到区段开头，
  而是指到你真正改的那一行。
- `status` 的改动清单同样带行号：`改动的点  aapp131.description（prog.full.4gl:5）`。
- `--json` 里每条发现有 `file` / `line` / `snippet` 字段，`status --json` 有
  `changed_regions` / `added_regions` / `deleted_regions` 三个数组（都带行号）。
- `fence.Parse` 的解析错误本来就带「（第 N 行）」；TAP 层的字节偏移会补上落在哪个
  `<point>` / `<section>`（例如 `TAP 字节偏移 12345，位于 <point name="function.foo">`）。

## 框架解锁（Unlocked）—— 取代 v1 的 `--allow-sec`

「解开框架」不是渲染效果，是**记录在 TAP 根属性上的单向状态**（`section_flag="Y"`）。
解开之后，**规格（SPEC）的任何调整都不会再产生对应的程序代码** —— 代码与规格从此脱钩。
所以它是一次显式、有代价、不可逆的状态迁移，不是开关：

```powershell
tt dev tzc export "pkg.tzc" -o ws     # Locked：SectionRegion 全只读
tt dev tzc unlock ws                  # 需要二次确认 → 打印设计器代价警告原文，退出码 4
tt dev tzc unlock ws --yes            # 确认后迁移（只改 workspace）
# 编辑区段正文 → apply（此时才把 TAP 根 section_flag 置 Y）
tt dev tzc apply ws
```

`cd ws` 之后 `ws` 一样可以省掉（`unlock --yes` / `apply` 都用当前目录）。

授权闸门逐条重放设计器 `checkBox_PreviewMouseLeftButtonDown`：

| 情形 | 判定 |
|---|---|
| 包本来就是解开态（`section_flag="Y"`） | Allow（设计器对这类包不设拦截，`unlock` 为 no-op） |
| `env="s"` 且 `login_user != "topstd"`，`std_section_verify="Y"` | Allow，但需 `--yes` |
| `env="s"` 且非 topstd，无 `std_section_verify` | **Deny，退出码 4**，打印 `adzi052` 授权原文 |
| 其他（`env="c"` 或 topstd） | Allow，但需 `--yes`（打印代价警告） |

> **R9：没有绕过。** 不提供 `--force`，不改 env/login_user，也不写 `std_section_verify`
> 假装有授权 —— 那是服务器端 `adzi052` 管理的**权限域**，本地工具自我授权等于伪造权限。
> **AI 不得自主加 `--yes`**：框架是否解开必须由人决定。

解锁 ≠ 所有区段可编辑（**解锁改变的是门，不是所有房间**）：三个锚点区段
（`other_function`/`other_dialog`/`other_report`）**永远只读**，TGL 标 `readonly="Y"` 的仍只读，
topstd 模式下的 src 规则仍然生效。围栏呈现：可编辑区段标 `[EDITABLE-SEC]`。

- `unlock` **不碰 `.tzc`**：只改 workspace（围栏旗标 + `.tdev/section-state`）+ git commit。
- Locked 态改区段 → 退出码 4，文案指向 `unlock`。
- `apply` 触发区段写入时提示「下次上传会触发服务器 `adzi520` 联动」（tt dev 不代为执行）。
- **反向迁移（重新锁上）不做**（R10）：那是服务器 `adzp064`（回标准）的职责。

## 结构事务（改名 / 改签名 / 改描述 / 改 scope）

设计器把自订定义点的「身份」拆在**六处**（点名、签名行、scope、描述块、墓碑、调用点），
任何一处单独漂移都是故障。所以 tt dev 不禁止你改，而是把「识别意图 → 证明一致 → 复刻事务」做成一等公民：

```powershell
tt dev tzc rename ws aapp131_calc aapp131_calc_v2            # 原子同步围栏 fn + 签名行
tt dev tzc rename ws aapp131_calc aapp131_calc_v2 --desc "#+ 新描述"
tt dev tzc rename ws aapp131_calc aapp131_calc_v2 --scope PRIVATE
tt dev tzc newfn  ws --type FUNCTION --name aapp131_added     # 在 APPEND 锚点插入新点
```

已经 `cd` 进 `ws` 时，`ws` 可以省掉（`rename` 变成只收「旧名 新名」两个位置参数）：

```powershell
tt dev tzc rename aapp131_calc aapp131_calc_v2 --desc "#+ 新描述"
```

`rename`/`newfn` **只改 workspace**（不碰 `.tzc`），并当场跑一遍结构事务预检
（与 `apply` 用的同一套校验，避免"命令行过了、apply 又拒"）。

**`newfn` 造出来的函数带设计器形态的函数头模板**（与 `FunctionGenerator.cs:20-67`、
真实包里的 `desc="\n####…"` 同形），且**顶部有一个空行**，所以落进 `.tap` 后
新函数不会与上一个函数的 `END` 挨着：

```
(空行)
################################################################################   ← 80 个 #
# Descriptions...: 描述说明
# Memo...........:
# Usage..........: CALL <新函数名>(传入参数)
#                  RETURNING 回传参数
# Input parameter: 传入参数变量1   传入参数变量说明1
#                : 传入参数变量2   传入参数变量说明2
# Return code....: 回传参数变量1   回传参数变量说明1
#                : 回传参数变量2   回传参数变量说明2
# Date & Author..: 日期 By 作者
# Modify.........:
################################################################################
```

模板只填**已知值**（`Usage` 行的函数名），`日期 By 作者` 等占位符逐字保留交给人/AI 填
—— 不自动填日期是为了让同一输入产出逐字节相同的工作区（红线 R7：时间轴不是功能）。
`newfn` **没有 `--desc`**：描述块固定为模板，要改既有函数的描述用 `rename --desc`。

改名落盘时 `apply` 复刻设计器 `ProgramInformation.Modify` 的事务序列：

```
.tap 里一次改名 = 一对 <point>
  function.旧名  status="d"   ← 墓碑，**保留原 CDATA 逐字节**
  function.新名  status="u"   ← 活点，签名行已改为新名
```

校验（V1–V7，全部 error 级，命中即拦且原包不变）：

| 码 | 判定 | 退出码 |
|---|---|---|
| V1 | 围栏 `fn` == 签名行的函数名（设计器 `ComplexException` 的前置防线） | 3 |
| V2 | 旧名在所有 Region 正文里的残留引用：可编辑区内 → 你自己改调用点 | 3 |
| V2 | 旧名残留在**只读区段**（含 Locked 态框架）→ 改不了，拒绝 | **4** |
| V3 | 新名与现存 point/TGL 占位符裸名冲突；含本次会话的重复目标与交换式改名 | 3 |
| V4 | 命名规范：普通程序 `<prog>_` 前缀、`.tzf` 为 `<prog>_<topind>_`（设计器仅 INFORMATION） | warn |
| V5 | scope 违反程序类型：`type=B` 强制 PUBLIC；M/G/X/Z/Q 强制 PRIVATE；S/W 可改 | 3 |
| V6 | 描述块里非空行必须以 `#` 开头（**允许空行**，见偏差 DV-2） | 3 |
| V7 | 目标必须是自订定义点（`function./dialog./report.`）且可编辑 | **4** |

描述与 `desc` 字段的关系：**块可直改**（它是可编辑区，apply 后 `desc` 由块重算）；
**`desc` 字段也可改**（视作显式意图，apply 用它重写块）；两者同时改且不一致 → 报 V6 让你只改一处。

## 围栏协议 v2

```
{//@tdev:begin point function.aapp131_qbe_clear [EDITABLE src="s" new="Y" order="1"
   fn="aapp131_qbe_clear()" scope="PRIVATE" desc="# 注释描述\n# 第二行"]}
PUBLIC FUNCTION aapp131_qbe_clear()
   ...正文...
END FUNCTION
{//@tdev:end point}

{//@tdev:begin point global.memo [READONLY PLAIN deny="env-edit-env" edit="s"]}
{//@tdev:begin section aapp131.main [EDITABLE-SEC src="s" ver="1"]}
{//@tdev:begin section aapp131.other_function [READONLY APPEND src="s" reason="框架集合锚点区段…"]}
```

- 旗标：`EDITABLE` / `READONLY` / `EDITABLE-SEC`（已解锁且可编辑的区段）/ `APPEND`（锚点可追加）/ `PLAIN`（裸名插入点）。
- **锚定部分不许改**（point/section 名、旗标、`status`/`src`/`new`/`order`…）；
  `fn`/`scope`/`desc` 是**唯一允许改**的字段，改了就触发 V1–V7。
- `PLAIN` 点（裸名插入点）**没有函数身份**：整块即正文，除只读判定外不做结构校验
  —— 特殊的是自订定义点，不是所有 point。
- 自订定义点的三类权限区：**描述块可编辑**（V6 校验形态）/ **签名行**由结构事务治理 /
  **`END` 行绝对不可改** / **正文本体自由编辑**。

## 行尾策略（重要，编辑器相关）

真实 `.tzc` 的 `.tap` 是**混合行尾**的（元素间 CRLF、CDATA 内 LF；capt110 实测 866 个 CRLF + 8466 个 LF），
渲染出来的 `prog.full.4gl` 继承了这一点。而**编辑器会自作主张把整份文件的行尾归一**
（VS Code 保存 capt110 时 866 个 CRLF → LF，文件正好少 866 字节）。

因此 tt dev 的字节恒等校验对**行尾等价**：

- **受保护字节**（围栏行、只读区内容、结构行、围栏外字节）：CRLF↔LF 的差异**不算改动**。
  这不是放宽安全性 —— 这些字节 tt dev **从不写回包**（`tapfile.Rewrite` 只在原 `.tap` 的字节区间上做替换），
  包里的它们永远保持原样。`verify` 会给一条 `gate1.eol-normalized` 的 info 说明。
- **任何其它字节差异**照旧按逐字节拦截。
- **可编辑区**：只有"行尾归一之外还有实质改动"的区域才会被重写；被重写的区域按**你文档里的行尾**写入包。
  未改动区域在包里的字节**一字不变**（自检里的真机事故用例会断言这一点）。

> 复盘记录（对抗用例集已收入，见 `selftest` 的「真机事故-编辑器把行尾归一」）：
> 旧实现逐字节比对，导致 VS Code 用户保存文件后 596/605 个 Region 被误判为改动、
> 只读区段"被改" → `apply` 以退出码 4 拒绝，用户只是加了一行注释却什么都写不回去。

## 退出码

| 码 | 含义 |
|---|---|
| 0 | 成功 |
| 2 | 包格式错误（zip 损坏、缺必须条目、`ver` 不存在/不匹配、TGL 区段标记不配对） |
| 3 | 验证失败（围栏外被改 / 结构行被改 / 不变量 / FGL 信封 / 装载模拟） |
| 4 | 写入被拒（改只读区、未开 `--allow-sec` 改区段、`--only` 范围外、删非 `new="Y"` 的点） |
| 5 | IO / 环境失败（磁盘、锁冲突、源包已被改动、git） |

## 危险开关

| 开关 | 语义 | 约束 |
|---|---|---|
| `--allow-sec` | **已废弃**（v2 起由 `unlock` 状态机取代） | 传入会报错并指向 `tt dev tzc unlock` |
| `unlock --yes` | 框架解锁的代价确认 | 无授权（`env=s` 且无 `std_section_verify`）时**加 `--yes` 也拒**（R9 无绕过） |
| `--yes`（apply） | 兼容保留，不再用于区段 | 区段的二次确认已上移到 `unlock` |
| `--only` | 部分导出（控制上下文体积） | 未导出的 Region 一律 READONLY，改了退出码 4 |

> ⚠️ `--allow-sec` + `--yes` 的后果是不可逆的：**解开框架 = 主动放弃「跟随原厂样板自动重产代码」的能力**。
> 这个决定必须由人做，AI 不得自主触发。

## 与设计指南的偏差（均已核实，逐条给出理由）

| # | 指南原文 | 实现 | 理由 |
|---|---|---|---|
| **D-1** | 新增点写 `status="c"` | **写 `status="u"` + `new="Y"`** | `AddPointModel.ToXML()` 在 `Status == CREATE` 时直接 `return null`（`AddPointModel.cs:1090-1093`）；设计器下一次保存会**静默丢弃**该点。设计器自身新增点也走 `CREATE\|MODIFY → "u"` |
| **D-2** | 围栏严格配对、**不嵌套** | 允许 `section → point` 一层嵌套，深度 > 2 报错 | 真实 TGL 里自订点住在区段内（3 个锚点区段的正文就是注入的点块）；平面围栏无法表达 |
| **D-3** | 禁止无法归属 Region 的散行 | 允许未归属字节段（prefix/gap/suffix），但**全部纳入 gate1 字节恒等**并登记在 `regions.json` | 真实包里必然存在（区段之间的空行）；禁止它们会让 105/105 个真实包直接失败。机械安全性不变 |
| **D-4** | `fgl_parse_function` 报解析错误 | 契约收窄为**块信封良构性**（7 个错误码），不做语句级语法校验 | 三个候选实现都不做语句级校验，且本机无 `fglcomp`/`fglgo` |
| **D-5** | 工作区只有 `.tdev/base.sha256` | 增加 `.tdev/base.full.4gl` | gate1 必须与「AI 实际看到的那份」比对，且 `status` 不应强依赖 git |
| **D-6** | 未定义 | `apply` 前校验源包 sha256 未变，变了 → 退出码 5 | 避免对着错误的包写回 |
| **D-7** | `apply [--yes]` 无语义 | 写区段除 `--allow-sec` 外必须 `--yes` | 「解开框架」是不可逆架构决策，需人工二次确认 |
| **D-8** | 不把时间戳当功能 | 写 zip 时沿用每条目原始 method/modtime（不取时钟），未知条目字节透传 | 结果确定、可复现 |
| **D-9** | CDATA 改写 | 新内容含 `]]>` → 退出码 3（I11b） | XML CDATA 装不下 `]]>`（语料 36,324 段实测 0 例），设计器会拆成两段 CDATA |
| **D-10** | — | 区段 id 与点名同名时按**配对**处理，不按名字查表 | 真实包存在 `prog.process` 既是 `<section id>` 又是区段内裸名插入点 |

### v2 新增的三条「文档 vs 源码」偏差（源码直证）

| # | 指南 | 源码 | 实现 |
|---|---|---|---|
| **DV-2** | §4.1 V6：描述块「每行 `#` 开头、**无空行**」 | `AddPointModel.cs:253` 的判据是 `!Regex.IsMatch(line,"^\s*#") && line.Trim() != ""` → **空行合法** | V6 允许空行；且只对**本次改动过描述块**的点报 error，存量数据降为 warn（真实语料里有历史遗留的不规范行） |
| **DV-3** | §4.1 V5：「type=B 强制 PUBLIC；M/G/X/Z/Q 强制 PRIVATE」= error | `FunctionInfoWindow.xaml.cs:143-192` 的 type→scope 映射是**新建点对话框的默认值**，整段被 `useDefaultScope` 门控 —— **不是对既有文本的硬约束**（真实语料 `aapt300(c).tzc` 就有 type=M 而函数声明 PUBLIC 的点） | 只在**本次发生 ScopeChange** 时给 **warn**；`newfn` 造新点时按该默认值填 scope |
| — | §4.1 要求「事务等价测试与设计器手工改名产物做语义 diff」 | 设计器无法进 CI | 自检用**独立参考断言**墓碑对形状（旧名 `status="d"` + 原 CDATA 逐字节；新名 `status="u"` + 签名行已改）；真机对拍列入 S7 人工项 |

> 这三条与 D-1 同源：**把真实数据里的常态当 error，会拒掉真实包**（README 陷阱 #1 的同一类错误）。

## 测试与验收

```powershell
# 单元 / 夹具 / 围栏回归（缺语料时自动跳过；约 2–3 分钟）
$env:GOCACHE="$PWD\.gocache"; $env:GOTMPDIR="$PWD\.gotmp"
go test ./...

# 全量深度语料回归：166 个真实包逐个 export + verify + apply 仿真（约 9–11 分钟）
$env:TDEV_DEEP="1"; go test ./... -timeout 30m

# 内置自检 + 对抗用例集（不需要真实语料）
.\tt.exe dev tzc selftest

# 覆盖语料目录（默认 D:\t100_wrok_dir）
$env:TDEV_CORPUS="D:\t100_wrok_dir"; go test ./...
```

> **为什么深度回归要显式开关**：`TestCorpusExportVerify`（S3，275 s）与
> `TestCorpusApplySimulation`（S5，约 250–380 s）合计 9–11 分钟，而 `go test` 默认
> `-timeout=10m`（go 命令在 10m+1m 处杀进程，报 `*** Test killed: ran too long (11m0s)`）。
> 也就是说默认命令下这两个用例会**随机器负载时好时坏**——同一份代码有时 530 s 通过、
> 有时 660 s 被杀。这种假失败比不跑更糟，所以默认跳过（会打印跳过原因与开启方法），
> 需要全量证据时用 `TDEV_DEEP=1` + `-timeout 30m` 显式开启。
> 测试二进制内部的 `-test.timeout` 拦不住 go 命令的杀进程，改它没有意义。

已实测的验收数据（本机）：

| 层 | 结果 | 复现命令 |
|---|---|---|
| 零改动 roundtrip | **165/165 个真实包**逐条目 sha256 一致 | `go test ./internal/pkgfile` |
| zip 容器保真 | **165/165 个真实包的零改动重建与原包逐字节相同**；原始解析与 `archive/zip` 在 660 个条目上逐字段一致；Zip64 包改动后仍可被独立读者读全 | `go test ./internal/pkgfile` |
| TAP 字节保真 | 32,571 个 `<point>` / 3,641 个 `<section>`；36,212 个带 CDATA 元素与独立正则对拍一致（36,208 通过、4 个重名点）；**36,373 个恒等写操作全部逐字节一致** | `go test ./internal/tapfile -count=1` |
| 围栏可逆性 | **165 个包、73,107 个 Region**，`parse_fenced(render(synthesize(pkg))) ≡ synthesize(pkg)` 逐项相等且区间记账对称 | `go test ./internal/fence` |
| FGL 解析器 | BDL 61 组夹具「通过 58 / 已知偏差 3 / 不符 0」；语料 4,318 个自订点 4,317 个通过（唯一失败是 `function.memo_industry` 这种「前缀是 function 但正文只有一行注释」的退化点） | `go test ./internal/fgl -count=1` |
| `verify` 语料 | **165 个包全部 0 error**（125 warn / 302 info） | `$env:TDEV_DEEP="1"; go test ./internal/cli -run TestCorpusExportVerify` |
| `apply` 仿真 | **137 个包**实际写回：只动目标点，`.4gl`/`ver` 与其它点字节不变 | `$env:TDEV_DEEP="1"; go test ./internal/cli -run TestCorpusApplySimulation` |
| `.tzs` 语料回归 | **67 个包 0 跳过 0 失败**；pin 67 条与磁盘一致；1285 帧请求、0 行异物输出 | `$env:TTZS_DEEP="1"; go test ./internal/dev/tzs/ -run TestCorpus -timeout 30m`（约 17 分钟） |
| 对抗用例 | **31 项自检全绿**（改围栏行/改只读区/删 end 围栏/塞 `]]>`/结构行/围栏外插行/`--only` 范围外/缺条目/ver 不匹配/区段不配对/改名事务/V1–V7/newfn 模板/newfn 无 `--desc`…），且**被拒的 apply 不改包** | `.\tt.exe dev tzc selftest` |
| 报错定位 | 只读区被改 → 打印 `prog.full.4gl:<行号>` + 该行内容（`--json` 带 `file`/`line`/`snippet`）；`status` 改动清单逐条带行号；行尾归一化不会把位置指到区段开头 | `go test ./internal/model ./internal/cli -run 'LineAt\|FirstDiffEOL\|TestApplyReportsReadonlyLine\|TestStatusReportsChangedLine\|TestVerifyReportsPosition'` |
| 安装面 | `install skills` 复制到 `<当前目录>/skills`（冲突拒绝 / `--force` 刷新 / 源=目标拒绝）；`install path` 解析真实折行的用户 PATH、幂等、不改写 `%USERPROFILE%`（写入前 `--dry-run` 可预览） | `go test ./internal/cli -run 'TestInstall\|TestMergeUserPath\|TestParseRegQueryPath'` |

> 前 4 行随默认 `go test ./...` 一起跑；`verify` 语料 / `apply` 仿真属于深度语料回归
> （PowerShell 里先 `$env:TDEV_DEEP="1"`，并给足 `-timeout 30m`）。
> 语料是**活的目录**（人也在里面干活）：包数/Region/点数的绝对值随语料增删而变，
> 换机器或语料变动后重跑上面的命令即可刷新；判据（0 error、逐字节一致）不随规模变。

### 真机验收清单（S7，需人工执行）

自动化只到 gate3（「模拟设计器装载」）。最后一步请在装了客户端的机器上做：

1. 取一个 `apply` 产出的包（`apply` 报告会打印路径与新 sha256）。
2. **先关掉设计器里该程序的文档**（否则设计器内存里的旧模型会在保存时覆盖你的改动）。
3. 在设计器里打开该 `.tzc`：应**无异常对话框**；改动内容可见。
4. 保存并关闭，再重新打开一次确认改动仍在。
5. 建议留证：改动前后的 `.tap` 的 `<point>` CDATA 逐字节比对。

> **输入必须是设计器产出的包**。如果输入本身就是 TDev 早期版本产出的（局部头 bit 3 + 数据描述符，
> 见上节），那它本来就不是设计器形态；tt dev 会把它归一化成设计器形态再写，
> 但**用这种包做真机验收等于同时验两件事**，容易误判。验收请从语料/客户端里原生的 `.tzc` 出发。

## 已知边界

- **不写 `.4gl`**（红线 R1）。它是服务器 build 产物，与「TGL + 展开点」的渲染结果**本来就不一致**（实测 82/105），覆盖即丢数据。
- **不动 `.tgl`**，除了 `--allow-sec` 改区段时同步打补丁（`.tap` 的 `<section>` 与 `.tgl` 的同一区段必须逐字节一致）。
- **不增删 zip 条目**、不改条目名、不给 `ver` 加目录前缀。
- 引用标准程序的包（TAP 条目基名 ≠ 根 `prog`）**拒绝写回**（I15）：设计器在这种包里会用 `CiteTAP` 内容，我们没有样本验证。
- **不做表单包（`.tzs`/`.tzv`）的写回**：只提供 `tt dev tzs export` 纯解压（只读参考），表单由设计器的表单设计器维护。
- `.tzg`（ReportCode）在设计器里没有必需条目检查分支，tt dev 只保证字节透传。

## 安装（免安装包 / 从源码）

免安装包就是「exe + README + skills」，没有安装程序、不写系统目录。合并后只有一个二进制 `tt.exe`，**一次安装覆盖全部工具**（调试 / 设计器包 / 数据字典）：

```powershell
# ① 解压 dist\tt-portable.zip 到一个固定目录，例如 D:\APPS\tt
# ② 把该目录加进**用户** PATH（只动 HKCU\Environment\Path，不需要管理员）
D:\APPS\tt\tt.exe install path
#    先看看会写什么、不落盘：
D:\APPS\tt\tt.exe install path --dry-run

# ③ 新开一个终端，然后在**你的项目目录**里装技能文件
cd /d D:\work\my-project
tt install skills                   # → D:\work\my-project\skills\tdev\SKILL.md
tt dev tzc selftest                 # 31 项自检
```

要点：

- `install path` **不用 setx**：setx 会把 `%USERPROFILE%` 这类展开式变量展开、还会把长 PATH
  截断到 1024 字符。这里直接读改写 `HKCU\Environment\Path`（保持原有 `REG_EXPAND_SZ`/`REG_SZ` 类型），
  并广播 `WM_SETTINGCHANGE` 让新开的终端立刻生效；已经是 PATH 里的目录则原样不动（幂等）。
- `install skills` 的源是**与 exe 同目录的 `skills/`**（不内嵌进二进制）；目标是
  **执行命令时的当前目录**下的 `skills/`，可用 `--to <dir>` 改（例如 Claude Code 项目级
  技能目录 `--to .claude/skills`）。目标已有同名技能目录时**默认拒绝**，确认要刷新加 `--force`
  （只删目标下同名的那一层，不动别的东西）。
- 从源码直接跑也一样：`go build -o tt.exe .` 之后，`skills/` 就在仓库根目录。

## 目录结构

合并后本工具的代码不再是独立仓库，而是统一项目 TT 里的一个命令组（`tt dev`）。
完整布局见 [ARCHITECTURE.md](ARCHITECTURE.md)，这里只列设计器包相关的部分：

```
internal/dev/              管线全部实现，与合并前 internal/ 一一对应：
  pkgfile/  tapfile/  tglfile/   zip / TAP / TGL 三层（字节保真的读写）
  fgl/      synth/     fence/    FGL 解析 / 合成 / 围栏渲染
  verify/   split/     store/   三道闸门 / 写回拆分 / 工作区
  model/    testutil/  cli/     域类型 / 夹具 / 命令行（tt dev 的派发层）
internal/cli/dev/          cobra 接入层：参数与退出码原样转发给 internal/dev/cli
testdata/fgl-fixtures/    FGL 夹具（61 组）
```

`tt dev` 之外的部分（配置、SSH、Web 服务）属于统一项目，不在这里展开。


## 许可与出处

- 本工具的实现依据是 T100 设计器（第三方商业软件，厂商标识 DSC）的**反编译源码**，仅用于个人学习、排障与接口对接研究；请勿用于重制发布或绕过授权。
- `testdata/fgl-fixtures` 与 `internal/fgl/outline.go` 移植自同作者的 MIT 项目 BDL（`D:\我的项目\BDL`）。
