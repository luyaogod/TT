# TDict — ERP 数据字典 CLI

面向 **T100 / Genero(4GL) ERP** 的本地命令行工具：

**ERP 数据字典查询**（给"读代码、做配置"提供上下文）：查 ERP 数据字典与业务元数据——数据表结构与字段中文含义（`tt dict r.t`）、字段校验规则（`tt dict r.v`）、下拉选项/系统分类码（`scc`）、字段画面规格（`desc`）、可复用开窗（`tt dict r.q`）、报错消息文本（`msg`）、系统与单据参数说明（`sysp`/`docp`）。查询**缺省直连 ERP 在线库**（数据最新最全，不用先同步），也可切到本地 SQLite 副本（`tt dict db sync` 刷新，离线可用）——见「查询数据源切换」。

此外保留 SSH 环境与数据库连接管理（`tt serve` 可视化配置页 / `tt env` / `tt dict db`）与服务器源码镜像（`tt dict mirror`）。

所有输出默认使用**简体中文 (zh_CN)**。

## 能查到的数据（规模）

查询命令覆盖的数据与当前体量（以正式区数据为例，同步后合计约 85 万行）：

| 内容 | 数据是什么 | 体量 | 查询命令 |
|---|---|---|---|
| 数据表字典 | 系统 **3,882 张表**的登记与字段定义：每张表做什么、字段中文含义/类型/长度/主键、键值与索引 | 约 50 万行 | `tt dict r.t <表名>`（表名示例：`dzea_t` 表登记档、`dzeb_t` 字段档、`oobd_t` 业务表） |
| 校验带值 (r.v) | 字段校验规则模板：校验 SQL、外部参数、判断条件、错误讯息 | 1,442 条 | `tt dict r.v [识别码]` |
| 系统分类码 (SCC) | 下拉/选项字典：分类码 →「值 → 说明」 | 2,066 个分类码、1.1 万分类值 | `tt dict scc [分类码]` |
| 字段画面规格 | 字段在画面的控件/下拉来源/格式/必填等设计参考 | 15.6 万条 | `tt dict desc <表名> [字段]` |
| 可复用开窗 (r.q) | 开窗选单定义：SQL、参数、显现/回传列 | 5,141 个 | `tt dict r.q [开窗码]` |
| 系统消息 | 所有提示/报错编号的文本、建议处理、技术细节（多语言） | 约 6 万条 | `tt dict msg <编号>` |
| 参数定义 | 系统/企业/据点/单据参数的用途说明、型态、值域、预设值 | 1,034 个参数（含 2,578 条单据性质绑定） | `tt dict sysp <编号>` / `tt dict docp <编号>` |
| 程序与作业 | 程序登记（中文作业名、程序类别、归属模块、客制、引用主程序）与作业挂程序的关联（**一个程序可被多个作业使用**） | 4,147 个程序、5,416 个作业、1.1 万条名称 | `tt dict prog [程序编号]` |
| 程序与表格 | **程序用哪些表、每张表做了什么操作**（T100 自己维护的代码分析索引，参考作业 azzq902） | 18.2 万条（14,448 个程序 × 3,825 张表） | `tt dict prog [编号]` / `tt dict r.t <表名> --who` |
| 子程序与元件 | 子程序 / 应用元件 / 报表元件 / 组件 / WebService 的登记与中文说明（与主程序登记**互补**，实测交集 0） | 4,036 个（子程序 1,316、应用元件 1,491、报表元件 1,130、组件 95、WS 4） | `tt dict prog <编号>` / `tt dict prog --sub` |

查询读的是本地数据；`tt dict db sync` 从 ERP 刷新后即包含以上全部内容（详见各命令节与「数据维护」）。

## 安装

### 方式一：下载便携版（Windows，免编译）

不想装 Go 的话，直接下载打包好的便携版：

**<https://github.com/luyaogod/TDictCli/releases/latest>** → `tt-portable-vX.Y.Z.zip`，解压即用。包含 `tt.exe`（内嵌配置页前端）、空的 `config.json`、`config.example.json`、`README.md`、`skills/`；**发布包里不含任何真实凭据与业务数据**，也**不含 `erp_data.db`**——首次解压后：

```bash
tt.exe serve            # 浏览器里「环境配置」填 SSH 环境与数据库
# 再到「数据同步」拉字典数据(默认写到 exe 同目录的 erp_data.db)
```

每个版本同时提供 `.sha256` 校验文件。可执行文件**未做代码签名**，首次运行可能被 SmartScreen 拦截，选「更多信息 → 仍要运行」即可。

### 方式二：全局安装（推荐）

安装到用户 PATH 后，可在任意目录直接调用 `tt`（合并后只有一个二进制，一次安装覆盖调试 / 设计器包 / 数据字典三个工具）：

```bash
# 1. 编译
cd tt
GOPROXY=https://goproxy.cn,direct go build -o tt.exe .

# 2. 把 exe 所在目录加入用户 PATH（HKCU，不需要管理员；先 --dry-run 可预览）
tt install path

#    或手动复制到 GOPATH/bin（要求它已在 PATH 中）
#    cp tt.exe $GOPATH/bin/

# 3. 设置数据库路径环境变量（永久生效，需重启终端）
setx TDICT_DB "D:\我的项目\TT\erp_data.db"

# 或者为当前会话设置
export TDICT_DB="D:\我的项目\TT\erp_data.db"
```

数据库查找优先级：`$TDICT_DB` > `-d` 参数 > 可执行文件同目录 > 当前工作目录。

### 方式三：手动指定数据库路径

将 `tt.exe` 和 `erp_data.db` 放到同一目录：
```bash
./tt dict r.t dzea_t -d ./erp_data.db
```

### 方式四：从源码编译

```bash
# 要求 Go 1.21+
cd tt
GOPROXY=https://goproxy.cn,direct go build -o tt.exe .
```

### 安装 Skill（推荐）

AI 技能文件（`tdict`、`erp-code-reader`）以普通目录 **`skills/`** 与 `tt.exe` 放在一起（便携版/仓库里都是），**不内嵌进二进制**，可直接编辑。每个技能是一个带 `SKILL.md` 的目录：

```bash
# 装到 Claude Code 直接读取的位置
tt install skills --to .claude/skills

# 或先装到当前目录(生成 ./skills/<技能名>/SKILL.md),之后自己摆放
tt install skills
```

## 快速上手

```bash
# 查版本与本地数据覆盖情况（先看这两个，再决定要不要同步）
tt version
tt dict db status

# 查询单张表的完整字典（表名、字段、键值、索引）
tt dict r.t dzea_t

# 查询多个表（逗号分隔）；只要表级信息时加 --brief，避免打出上千行字段明细
tt dict r.t "dzea_t,dzeb_t,dzed_t"
tt dict r.t --brief "apca_t,apcb_t,glab_t"

# 按中文业务词找表（读代码时最常用）
tt dict r.t --kw 应收

# 输出 JSON / CSV
tt dict r.t dzea_t --json
tt dict r.t dzea_t --csv

# 程序用了哪些表 / 这张表被哪些程序用（改表前的影响分析）
tt dict prog aapi011
tt dict r.t glab_t --who

# 从 ERP 实时刷新数据
tt dict db sync

# 列出全部校验带值定义，并查看指定 dzcd001 的详情
tt dict r.v
tt dict r.v v_ooba002_07

# 列出全部系统分类码 (SCC)，并查看指定 gzca001 的详情
tt dict scc
tt dict scc 4

# 查询指定表的字段规格（控件/SCC码/格式），以及单字段完整规格
tt dict desc oobd_t
tt dict desc oobd_t oobd002

# 列出全部可复用开窗，并查看指定 dzca001 的详情
tt dict r.q
tt dict r.q q_apca001

# 报错/参数编号 → 消息文本与处理建议、参数用途说明
tt dict msg std-00006
tt dict sysp A-SYS-0040
tt dict docp D-MFG-0076

# 数据较旧先刷新;或临时直查某环境(不需要本地数据)
tt dict db sync
tt dict msg aoo-00120 --env 正式区
```

## 命令

> 命名说明：`r.t`（数据表，对应表格设计器 r.t）、`r.v`（校验带值 r.v）、`r.q`（开窗 r.q）、`desc`（字段规格）、`scc`（分类码）均对齐 T100 原生工具/术语；`prog`（程序与作业，数据来自 azzi900 程式基本資料 / azzi910 作業基本資料，别名 `program`/`job`）。短名 `rt`/`rv`/`rq` 与旧名 `table`/`check`/`win`/`spec` 仍作为别名可用。

### `tt dict r.t [表名]`

查询一张或多张表的完整字典：这张表在系统里做什么（表说明/所属模块/表类型）、每个字段的中文含义、类型与主键必填、键值与索引。读代码、看 SQL、查界面字段含义时用它。

返回内容包含：

- **表名/表说明**：表名与中文说明、所属模块、表类型（取值含义见文末「输出中的类型码」）
- **字段**：序号、字段名、字段说明、数据类型、长度、主键、必填、备注
- **键值**：键名、类型（主键/外键/唯一）、键值字段、外键表、外键字段
- **索引**：索引名、类型、索引字段

```
$ tt dict r.t dzea_t
=== dzea_t ===
表说明: 数据表主档
模块: ADZ    类型: X

字段 (31):
序号  字段名        字段说明      数据类型       长度  主键  必填  备注
--  ---------  ----------  ---------  ---  --  --  --
1   dzeastus   状态码       varchar2   10
2   dzea001    table编号   varchar2   15   PK  Y
...

键值 (1):
键名       类型  键值字段   外键表  外键字段
-------  --  -------  ---  ----
dzea_pk  主键  dzea001

索引 (1):
索引名      类型  索引字段
-------  --  -------
dzea_pk  U   dzea001
```

`--json` 输出结构化对象（`表名/表说明/模块/类型/字段/键值/索引`）；`--csv` 输出字段明细（含表名列）。

**不给表名时是列表模式**：列出全部表（表名/表说明/模块/类型/字段数），`--kw <关键字>` 按表名或中文表说明过滤——读代码时"从业务词找表"就走它：

```
$ tt dict r.t --kw 应收
表名      表说明        模块   类型  字段数
------  ---------  ---  --  ---
aphc_t  应收票据转付明细档  AAP  D   42
nmcn_t  应收票据主档     ANM  M   89
...

共 8 张表
```

> 表说明有简繁两种写法（`zh_CN` 列与原始繁体档），同一词也可能有异体（如 `对账`/`对帐`/`對帳`）：搜不到时换个写法、更短的词，或加 `--lang` 切换语言别。

**只要表级信息时加 `--brief`**：指定多张表时默认会打印每张表的**全部字段**（7 张表 ≈ 700 行），`--brief` 只给「表名/表说明/模块/类型/字段数」，与列表模式同样的表——读代码时"这几张表分别是什么"就用它，要看字段再单表查：

```bash
tt dict r.t --brief "glab_t,glaa_t,glad_t,apca_t,apcb_t,oocq_t,ooag_t"
```

### `tt dict r.v [dzcd001]`

查询 **字段校验规则 (r.v)**——系统保存数据前做的各种检查是可复用的"校验模板"，每条校验一个识别码：含要执行的校验 SQL（SQL 里用 `<field>`、`arg1~9`、`:TODAY` 等占位符，运行时代入实际字段与参数）、外部参数定义与判断条件。想知道某字段会被怎样校验、校验不过显示什么错误讯息，或要给字段绑定校验时用它。

校验识别码示例：`v_ooba002_07`（形如 `v_表_用途`）；也可 `tt dict r.v --kw 料号` 按识别码/说明搜索。

```bash
# 列出全部校验定义（识别码/客制/说明/型态/错误讯息/备注）
tt dict r.v

# 按关键字过滤（识别码或说明）
tt dict r.v --kw 料号

# 指定 dzcd001 的完整详情
tt dict r.v v_ooba002_07

# 指定说明语言别（默认 zh_CN）
tt dict r.v v_ooba002_07 --lang zh_TW

# JSON 输出（列表/详情）；--csv 仅列表模式
tt dict r.v v_ooba002_07 --json
tt dict r.v --csv
```

详情包含（标准/客制变体分别显示）：

- **单头**：型态（1=检查存在 / 2=带值 / 3=检查存在并带值）、错误讯息代号、行业别、状态码、说明、备注
- **SQL 指令**：校验 SQL 原文，附标签图例（`<field>` `<table>` `<wc>` `<count>`、`arg1~9`、`:TODAY`、`:DEPT` 等全局变量）
- **参数**：顺序、参数名称（对应 arg1~9）、日期型态（1=年月日 / 2=年月日时分秒 / 3=毫秒 / 其他=字符串）、说明、备注
- **判断条件**：顺序、条件 SQL、条件成立时显示的错误讯息

### `tt dict scc [gzca001]`

查询 **系统分类码 (SCC)**——系统里各种下拉/选项的"选项字典"：币别、单据类型、料件属性等分类统一按分类码组织，一个分类码 = 一组「值 → 说明」，画面上的下拉选项就来自这里。查字段可选项、下拉数据来源、按分类码取说明时用它。

分类码示例：`4`、`241`（多为数字）；`--kw 币别` 按分类码/名称搜索。

```bash
# 列出全部分类码（分类码/群组/状态/名称/值数）
tt dict scc

# 按关键字过滤（分类码或名称）
tt dict scc --kw 币别

# 指定 gzca001 的完整详情
tt dict scc 4

# 指定说明语言别（默认 zh_CN）
tt dict scc 4 --lang zh_TW

# JSON 输出（列表/详情）；--csv 仅列表模式
tt dict scc 4 --json
tt dict scc --csv
```

详情包含：

- **单头**：群组、状态（Y=启用/N=停用）、名称、说明、说明2
- **分类值**：值、说明（即下拉选项显示文字）、排序（下拉按此排序）、标准/客制
- **扩展数据列**（`gzcb003`~`gzcb015`，仅显示非空值）：含义由各分类码自己定义（如某分类码用它们存群组编号区间、另一个存格式），结合该分类码的语境理解

### `tt dict desc <表名> [字段名]`

查询**字段的画面规格**——每个字段在画面上用什么控件（输入框/下拉/日期…）、下拉取哪个系统分类码、显示格式与宽度、必填与否、默认值、开窗程序、最大/最小值等，画面设计器按它生成画面。要复刻/理解画面字段行为、查字段编辑控件或下拉来源时用它。

示例：`tt dict desc oobd_t`（整表全部字段规格）；`tt dict desc oobd_t oobd002`（单字段完整规格）。

```bash
# 列出该表全部字段规格（序号/字段/字段名/控件/SCC码/必填/宽度/格式/默认值/校验带值）
tt dict desc oobd_t

# 单字段完整规格（含 SCC 名称、取值范围、开窗、串查、报表栏宽等全部列）
tt dict desc oobd_t oobd002

# 指定说明语言别（默认 zh_CN）
tt dict desc oobd_t oobd002 --lang zh_TW

# JSON 输出（列表/详情）；--csv 输出全列明细
tt dict desc oobd_t oobd002 --json
tt dict desc oobd_t --csv
```

输出语义（每条规格的字段含义）：

| 输出项 | 含义 |
|---|---|
| 控件 | 字段在画面的控件类型：03=ComboBox（下拉）05=Edit（输入）01=ButtonEdit（带开窗按钮）04=DateEdit 09=RadioGroup 11=SpinEdit 12=TextEdit 13=TimeEdit 34=DateTimeEdit 等 |
| SCC码 | 下拉/单选控件的选项来源（哪个系统分类码；用 `tt dict scc <码>` 查可选项） |
| 必填 | Y/N |
| 显示宽度 | 控件宽度（"整数,小数"） |
| 格式 | 显示格式 format |
| 大小写 | U/L |
| 默认值 | 默认值 |
| 最大值/最小值 | 取值范围（配比较符号） |
| 编辑开窗 / 查询开窗 | 字段上开窗按钮弹出的程序 |
| 校验带值 | 该字段绑定的校验识别码（用 `tt dict r.v <码>` 查校验内容） |
| 串查型态 / 串查程序 | 串查/参考程序 |
| 报表栏宽 / 小数位 | 报表输出参考 |
| 客制识别 | 标准/客制 |

### `tt dict r.q [dzca001]`

查询**可复用开窗 (r.q)**——代码里 `CALL q_xxx()` 弹出的查寻选单定义：每条开窗 = 一段带占位符的 SQL（选取字段/表/条件区用 `<field>/<table>/<wc>` 标记）+ 外部参数（对应 arg1~9）+ 显现与回传列。读代码遇到开窗调用、想知道它查什么表、回传哪些值时用它。

开窗码示例：`q_apca001`（`q_` 开头）；`--kw 料号` 按说明搜索。

```bash
# 列出全部开窗（识别码/客制/说明/状态/每页笔数/HardCode/行业别）
tt dict r.q

# 按关键字过滤（识别码或说明）
tt dict r.q --kw 料号

# 指定 dzca001 的完整详情
tt dict r.q q_apca001

# 指定说明语言别（默认 zh_CN）
tt dict r.q q_apca001 --lang zh_TW

# JSON 输出（列表/详情）；--csv 仅列表模式
tt dict r.q q_apca001 --json
tt dict r.q --csv
```

详情包含（标准/客制变体分别显示）：

- **单头**：状态、SQL 指令（原文，附标签图例）、每页笔数、作业串查编号、HardCode（Y=跳过自动产生）、行业别、说明、助记码
- **参数**：顺序、参数名称（对应 argN）、日期型态（1=年月日 / 2=年月日时分秒 / 3=毫秒 / 其他=字符串）、说明
- **显现设定**：显现顺序、字段编号、表格别名、显示控件、是否回传（Y 的字段按显现顺序构成 return1~9）、大小写、显示格式、标签缀字

### `tt dict msg <编号>`

查询**系统消息**——所有提示/报错消息的记录，每个消息一个编号。程序报错、日志里出现的编号，都能查到它的完整文本、建议处理方式与技术细节。消息编号格式：`std-00001`、`azz-00041`、`lib-xxxxx`（类型-流水），负整数（SQLCODE 如 `-100`）需用 `--` 分隔：`tt dict msg -- -100`。

```bash
# 默认语言(zh_CN)行
tt dict msg std-00006

# 指定语言
tt dict msg azz-00041 --lang zh_TW

# 多编号、JSON
tt dict msg "std-00006,azz-00041" --json

# 远程直查
tt dict msg aoo-00120 --env 正式区
```

输出长这样：

```
=== std-00006 (zh_CN) ===
类型:     错误 (1)
状态:     启用
文本:     输入的资料已存在
建议处理: 请重新审核您所需要输入的资料
```

返回：**文本**（报错原文）、**建议处理**、**建议作业**（附作业名称）、**技术细节**（给程式人员的详细讯息）、**类型**（0=警告 / 1=错误 / 2=资讯）、**状态**（启用/停用）。消息按语言各一条：默认只显示 `--lang`（zh_CN）行，该编号无此语言时命令会列出可用语言（与源系统一致：精确匹配、无自动回退）。

### `tt dict sysp <编号>` / `tt dict docp <编号>`

查询**系统/单据参数说明**——决定系统行为与单据流程的可配置项。每个参数有编号、名称、中文说明与"怎么填"的定义（型态/值域/预设值）。

- `tt dict sysp` 查**系统/企业/据点级参数**，编号格式 = 型态码 + 领域 + 4 位流水：`A-SYS-0040`（A=系统级）、`E-CIR-0001`（E=企业级）、`S-BAS-0028`（S=据点级）；
- `tt dict docp` 查**单据别参数**（编号 `D-` 开头，如 `D-MFG-0076`），并附该参数适用的单据性质（模块 + 单据性质清单）。

```bash
tt dict sysp A-SYS-0040              # 系统参数(默认 zh_CN 说明)
tt dict sysp S-FIN-3014 --lang zh_TW
tt dict docp D-MFG-0101              # 单据别参数 + 单据性质绑定
tt dict docp "D-MFG-0076,D-BAS-0058" --json
tt dict docp A-SYS-0040              # 群不对会提示改用 sysp
```

输出长这样：

```
=== A-SYS-0040 (zh_CN) ===
名称:     系统是否允许LIB相关资源被签出
说明:     系统是否允许LIB相关资源被签出
群:       gzsa_t  系统级参数(系统全域参数)
型态:     Y/N (1)
领域:     SYS
预设值:   N
异常处理: 改抓设定时的预设值 (1)
状态:     启用
```

返回：名称/说明（多语言）、参数群与级别、**型态**（1=Y/N 2=整数选项 3=范围设定 4=字符或SCC 5=日期）、领域、预设值、值域、SCC 选项、校核/开窗引用、取参异常处理、修改频度、即时抓取、状态与长备注；`tt dict docp` 另附**单据性质绑定**表。注意：这里查的是参数的定义与说明，各环境下参数**实际设的值**在参数值维护作业/画面里，不在本命令范围。

### `tt dict prog [程序编号]` — 程序与作业

查询 T100 的**程序与作业登记**（azzi900 程式基本資料設定作業 / azzi910 作業基本資料維護）：程序编号对应的中文作业名、程序类别、归属模块、是否客制、引用主程序与系统运行指令，以及**哪些作业用了这个程序**。

关键关系：作业通过 `gzzz_t.gzzz002` 挂到程序上，**一个程序可以被多个作业使用**（共用维护程序可达数百个作业）；作业的显示名称取它所挂程序的名称（`gzzal_t.gzzal003`，与 azzi910 的取法一致）。拿作业编号也能查，会自动跟到它挂的程序。

读代码时的第一个问题"这个程序做什么"就用它：

```
$ tt dict prog aapi011
=== aapi011 ===
程序名称: 应付账款类别依账套设置科目作业
程序类别: I(基本资料维护)
归属模块: AAP
客制: s
状态码: Y
系统运行指令: $FGLRUN $AAPi/aapi011

使用它的作业 (1):
作业编号     作业名称             归属模块  应用参数组  参数组说明  默认单据性质
-------  ---------------  ----  -----  -----  ------
aapi011  应付账款类别依账套设置科目作业  AAP   0
```

```
$ tt dict prog aooi701          # 作业编号 → 它挂的程序
=== aooi701 ===
该编号不是程序登记,是作业:挂的程序 = aooi301
```

- `tt dict prog --kw 对账` —— 按程序编号或中文名称搜索（从业务词找程序/作业）

**「程序**↔**表格」两个方向都能查**（数据来自 `gzdg_t` 程序与应用表格功能分析表，由 T100 自己维护，参考作业 azzq902「程式編號對應表格查詢」）：

```
$ tt dict prog aapi011
…（程序信息、作业清单）…
使用的表格 (11):
表格编号     表说明            操作
-------  -------------  -------
glab_t   账套应用会计科目设置档    S/I/U/D
glaa_t   账别数据档          S
aapi011  应付账款类别依账套设置科目作业  S/I/U/D
操作: S=SELECT 查询 / I=INSERT 新增 / U=UPDATE 修改 / D=DELETE 删除

$ tt dict r.t glab_t --who          # 反查:改这张表会影响谁
=== glab_t (账套应用会计科目设置档) ===
使用它的程序 (376):
程序编号        程序名称             操作
aapi011     应付账款类别依账套设置科目作业  S/I/U/D
aapi201     结算费用依账套设置科目作业    S/I/U/D
…（--limit 默认 20，0 = 全部）
```

- **比 grep 准**：`grep` 会把被注释掉的 SQL、字符串里的表名、`LIKE 表名.字段` 的变量声明都算进来，还混进 `type_t`；这个索引只登记**真实 SQL 访问**，并带操作类别（S/I/U/D）。
- **覆盖库与元件**：不只作业主程序——实测 `cl_abi`（库）22 条、`s_apcp300`（元件）8 条、`q_adzi052`（开窗）9 条。所以 `tt dict prog <库名/元件名/开窗码>` 也能用它。
- **边界**：**子程序**（`aapq110_01` 这类，实测 0 条）与**动态 SQL**（`CURSOR FROM 变量`）不入索引，这些仍需 `grep` 兜底。

**子程序 / 元件 / 库的登记**（数据来自 `gzde_t` 子程序及应用元件基本数据表，参考作业 azzi901「子程式及元件基本資料設定作業」）——与主程序登记**互补且不重叠**（实测 `gzza_t` 4,147 个主程序、`gzde_t` 4,036 个子程序/元件，交集 0），所以 `tt dict prog <编号>` 对两类编号都能答：

```
$ tt dict prog aapq110_01
=== aapq110_01 ===
说明: 供应商对账单明细查询报表打印
规格类别: S(子程序)
归属模块: AAP
程序类别: Q(查询)
客制: s
(登记在「子程序及应用元件基本数据表」gzde_t,不在程序/作业登记里)
它的主程序: aapq110(供应商对账单明细查询);子程序自己做的事以文件头 `#+ Description:` 为准。

$ tt dict prog --sub --kw ABI        # 在子程序/元件里按业务词搜
规格编号              说明               规格类别  归属模块  客制
cl_abi            ABI Library      B     LIB   s
sadzp188_abi_rep  亿信ABI过单导出入函式库    S     ADZ   s
…
```

规格类别（SCC 91）：`B`=应用元件 / `S`=子程序 / `G`=报表元件-GR类 / `X`=报表元件-XG,FR类 / `K`=报表组件-XR类 / `W`=WebService元件。**注意子程序与它的主程序是两回事**（上面 `aapq110_01` 是"报表打印"、主程序 `aapq110` 是"明细查询"），工具的降级提示会明说这一点。


- 程序名称有简体（`--lang zh_CN`，默认）与繁体（`--lang zh_TW`）两份，搜索用字要对应
- 一个程序被很多作业使用时默认只列前 20 个作业（`--limit 0` 显示全部，`--json` 导出）

### `tt dict db status` — 本地数据覆盖情况

`tt dict` 的各查询命令依赖不同**数据族**（表字典/校验带值/系统分类码/字段画面规格/可复用开窗/系统消息/参数定义/程序与作业）。命令报「本地库尚未包含 XXX 数据」时，用它一次看清哪个族没同步、各多少行、缺哪些表：

```
$ tt dict db status
本地库: D:\APPS\tt\erp_data.db  (56.5 MB, 2026-08-26 13:18)

状态  数据族     行数      缺表                    依赖命令
--  ------  ------  --------------------  ---------
✓   表字典     498833  -                     r.t
✗   校验带值    0       dzcd_t,dzcdl_t,...    r.v
✗   程序与作业   0       gzza_t,gzzz_t,...     prog
...

未同步的数据族用 `tt dict db sync` 补齐(需能连 ERP);临时也可 `tt dict <命令> --env <环境名>` 远程直查。
```

只读、不改任何文件，也不需要 `config.json`。各数据命令的 `--help` 末尾也会带一行本命令数据族的状态（`tt dict --help` 给全部数据族的概览）。

**读法**：`✓`=该族已完整同步；`✗`/`✗ 部分`=该族字典表不全，**依赖它的命令会整体报错**（不是部分可用——所以行数列显示 `—`，因为那个行数没有意义）。补齐两条路：`tt dict db sync`，或临时用 `tt dict <命令> --env <环境名>` 免写盘直查。

### `tt version`

打印版本与构建来源：打包脚本用 `-ldflags` 注入版本号（如 `0.1.1`），并附上 commit 短号、提交日期、是否有未提交改动；本地 `go build` 出来的则显示 `devel (commit <短号>+未提交改动, <日期>)`。

```bash
$ tt version
tt 0.1.1 (commit 8c57e1d, 2026-09-17)
```

用来判断「手上的二进制与文档是不是同一版」；每个命令的 `--help` 末尾也会带一行（`版本: tt dict …`）。

### `tt dict db sync` — 数据维护（刷新本地查询数据）

> 通常不必用 CLI：`tt serve` 的「数据同步」视图可选环境并显示逐表进度（见「可视化配置页」）。下列命令等价，便于脚本/自动化。

上文各查询命令读的是本地数据；`tt dict db sync` 从 ERP 把它刷新到最新（覆盖表字典、校验、分类码、画面规格、开窗、消息、参数等全部内容）。

```bash
# 全量刷新（约 85 万行，视网络约 2~4 分钟）
tt dict db sync

# 只刷部分（一般不必）
tt dict db sync --table dzea_t

# 指定数据源环境与库文件
tt dict db sync --env 正式区 -d D:/path/to/erp_data.db
```

- 同步需能访问 ERP 服务器（内网/VPN）；采用临时库整体替换，失败不影响原库，完成后原库自动备份为 `<库>.bak`。
- 缺省数据源 = 默认环境（`hosts.activeEnv`，用 `tt env list` / `tt env use` 查看/切换）的库；环境与库配置见「数据库连接与查询数据源」。
- 不想维护本地数据时，查询命令加 `--env <环境名>` 直接查远程（与本地同一批数据、同一输出）。

### `tt install skills` / `tt install path`

把 `tt` 装进使用者的环境。合并后三个工具共用同一个二进制、同一条 install 命令。

**`install skills`** —— 把**与 `tt.exe` 同目录的 `skills/`** 复制到**当前工作目录**（命令在哪运行就装到哪）。
每个技能是一个带 `SKILL.md` 的目录（Claude 技能规范：目录名必须等于 frontmatter 里的 `name`）；
`skills/` 是普通 markdown、可直接编辑，不内嵌二进制。

```bash
# 装到当前目录(生成 ./skills/<技能名>/SKILL.md)
tt install skills

# 装到 Claude Code 直接读取的位置
tt install skills --to .claude/skills

# 已存在同名技能目录时默认拒绝(列出冲突);确认要刷新再加 --force
tt install skills --force
```

**`install path`** —— 把 `tt.exe` 所在目录追加到**用户** PATH（`HKCU\Environment\Path`，不需要管理员，
绝不碰系统 PATH）。已存在时幂等，原有的 `%USERPROFILE%` 之类可展开变量原样保留、不重排顺序。
与设置页「加入 PATH」是同一份实现（`pathinstall` 包）。

```bash
tt install path --dry-run   # 只预览将要写入的内容,不碰注册表
tt install path             # 实际写入(新开的终端生效)
```

### 环境管理（`tt env`）

环境清单是**三个工具共用**的（`config.json` 顶层 `hosts.sshs`），统一由顶层 `tt env` 组查看 / 切换 / 设置：

```bash
tt env list                  # 列出全部环境与当前默认(* 标记默认环境)
tt env show 正式区            # 看某环境的连接详情(口令打码)
tt env use 正式区             # 把默认环境写入 config.json 的 hosts.activeEnv
tt env topent 正式区 99       # 设置该环境的默认企业编号(TOPENT)
```

缺省数据源（`db sync`、`mirror`、`--env` 直查）都取该默认环境；`--env <环境名>` 可对单次调用覆盖。`tt dict env [<环境名>]` 保留为字典侧的等价命令。

## 可视化配置页（`tt serve`）

在浏览器里维护 `config.json` 的 SSH 环境与数据库连接，并可直接跑源码镜像与数据同步（通常不必走 CLI）：

```bash
tt serve                              # 默认 127.0.0.1:28670(占用自动顺延),Ctrl+C 停止
tt serve --listen 127.0.0.1:9123
```

左侧活动栏切换四个视图：

- **环境配置**（左侧环境列表 + 右侧「SSH 服务器 / 数据库」Tab 编辑同一环境）：

| 操作 | 说明 |
|---|---|
| SSH 服务器 | 环境名、主机、端口、账号、密码、登录区域(zone)、TOPENT；可增删环境、设为默认环境 |
| 数据库 | 每个环境一对一挂载：类型(oracle/kingbase)、主机、端口、服务名(service)/库名(database)、账号列表；**新增环境时账号列表默认预置 `ds` / `dsdata` / `dsdemo` 三个账号(密码=账号)**，可自行增删改。账号列表是可编辑表格（对齐 TDebug/TbmLite 的 Excel 风格：网格线由单元格承担、控件嵌在格内、聚焦时整格描边），行内 👁 切换该行密码明文 |
| 从服务器获取数据库配置 | 用该环境 SSH 登录服务器**只读探测**并回填连接要素：oracle 按区域加载环境后读 `ORACLE_HOME`/`sqlplus`/`TWO_TASK` 并解析 `tnsnames.ora` 得到 host/port/service；kingbase 发现运行中实例的数据目录、端口与库名。解析出内部主机名(客户端不可达)时回填 SSH 主机 |
| 账号行 ⚡ 验证 | 服务器侧以该「账号/密码」连显式目标库执行 `select 1`(只读) |
| 数据库「测试连接」 | 客户端直连测试（与 `tt dict db ping` 同链路：账号取列表首项） |
| 保存 | 整块写回顶层 `hosts` 键（`query`/`mirror`/`bdldoc` 原样保留），并兼容迁移旧键 `debug` |

- **源码镜像**（等价 `tt dict mirror dir` / `pull`，带实时进度）：

| 操作 | 说明 |
|---|---|
| 镜像根目录 | 查看/设置 `config.json` 顶层 `mirror.dir` |
| 环境下拉框 | 选择要拉取的环境，**默认选中默认环境**（`hosts.activeEnv`，标「（默认）」）；下方显示该环境的镜像目录与是否已拉取（有无完整基线 `.tdict-mirror.ok`，增量前提） |
| 增量更新 / 全量重建 | 后台执行 `host.MirrorPullProgress`；全量重建需二次确认（整目录替换）。同一时间只允许一个拉取任务 |
| 拉取进度 | 分阶段显示（连接 → 探测 T100 目录 → 服务器打包 → 下载并解压 → 完成）：打包阶段轮询服务器归档大小，下载阶段按压缩字节显示 `已传输/总量 (百分比)`、文件数与用时 |

- **数据同步**（等价 `tt dict db sync`，带实时进度）：

| 操作 | 说明 |
|---|---|
| 目标数据库 | **可在页面上编辑并保存**（写入 `config.json` 顶层 `sync.target`）；默认是 **exe 同目录** 的 `erp_data.db`——便携版分发到任何机器都成立，**不再受宿主机 `TDICT_DB` 影响**。文件（含父目录）不存在时同步会自动创建；「恢复默认」清空自定义值。原库自动备份为 `.bak` |
| 环境下拉框 | 只列**已挂数据库**的环境，**默认选中默认环境**（`hosts.activeEnv`） |
| 开始同步 | 后台执行 `dbsync.Run`：远程逐表拉取 → 写临时库 → 建主键索引 → 原子替换本地库。需二次确认；同一时间只允许一个同步任务 |
| 同步进度 | 数据表 `第 x/y 张`、当前表与行数、累计行数、阶段（连接 → 拉取 → 建索引 → 替换 → 完成）、用时、原库备份路径与索引警告 |

- **设置**（命令行安装 / BDL 文档目录 / 运行信息）：

| 操作 | 说明 |
|---|---|
| 查询数据源 | 下拉切换查询命令读哪里：**在线**（默认环境或指定环境的 ERP 库直查，数据最新），或**本地 SQLite**（`erp_data.db` 副本，快、可离线）。写入 `config.json` 顶层 `query.source`；两者不自动切换 |
| 命令行安装 | 显示当前 `tt` 可执行文件与所在目录，一键**加入用户 PATH**（Windows 写注册表 `HKCU\Environment\Path`，用户级、无需管理员），之后任意位置可直接运行 `tt`；也可一键移除。已加入/未加入有明确状态 |
| 生效时机 | 新开的终端立即可用；**已打开的终端需重开**（写入后广播 `WM_SETTINGCHANGE`） |
| 非 Windows | 不自动改 PATH，页面给出等价的 `export PATH="$PATH:<目录>"` 手动命令 |
| BDL 语言文档目录 | 查看/设置 `config.json` 顶层 `bdldoc.dir`（等价 `tt dict bdldoc dir <目录>`），并提示该目录在本机是否存在；接口 `GET/PUT /api/bdldoc` |
| 运行信息 | 配置文件路径、当前服务地址 |

- 配置文件取 `--config` / `TT_CONFIG`（旧名 `TDICT_CONFIG` 仍识别）；缺省位置是**统一用户目录** `%APPDATA%\T100\tt\config.json`（可用 `T100_HOME` 整体改写），**三个工具共用这一个文件**。文件不存在时首次保存自动创建，目录也一并建好。
- **便携版发布为空配置、且不含业务数据**：打包脚本用 `config.empty.json` 生成空的 `config.json`（`hosts.sshs` 为空），首次运行用本页添加自己的环境；同时也**不打包 `erp_data.db`**（含客户表字典/schema/企业码等数据），配好环境后自行用「数据同步」或 `tt dict db sync` 拉取。技能以普通目录 `skills/` 随包提供（**不内嵌二进制**，可直接编辑）。仓库中不提交 `config.json` 与 `erp_data.db`（见 `.gitignore`）。
  > 注意区分**发布包**与**你自己那份安装副本**：发布包里是空配置；而你把工具解压/安装到某个目录、配置过环境并同步过数据之后，**那个目录里当然会有你的真实凭据与 `erp_data.db`**——那是你的工作副本，不是发布物，别拿它跟上面这句话对照。
- 服务器执行工具（sqlplus/ksql）路径自动探测，无需配置；SSH/DB 探测逻辑与 `tt dict db discover` 同源（`host` 包）；镜像逻辑与 `tt dict mirror pull` 同源，同步逻辑与 `tt dict db sync` 同源（`dbsync` 包）。
- 该服务只做配置读写、只读探测与同步拉取。

## 本地源码镜像（tt dict mirror）

> 通常不必用 CLI：`tt serve` 的「源码镜像」视图可直接设置镜像根、执行增量/全量拉取并显示实时进度（见「可视化配置页」）。下列命令等价，便于脚本/自动化。

把某环境(T100 服务器)的源代码镜像到本地目录,供 **AI 用本地文件工具读码**(ls/rg/带行号读),避免每次读取都经服务器往返、也避免给 AI 服务器端查询权限。**只镜像**:各模块 `4gl`(源码)、`4fd`(前端字段描述)、`42s`(编译字符串,**只取 `zh_CN` 语言目录**——文件所在目录的最后一个目录必须是 `zh_CN`),以及 **`*.inc`(4GL include,erp/com 下任意位置)**。per(界面源)、编译产物(42m/42r/42f)、其它语言、设计器辅助目录一律不拉;目录与服务器同构(`<镜像根>/<环境名>/erp/<模块>/{4gl,4fd,42s/zh_CN,inc}/...`、`com/{lib,sub,qry,wss}/...`),客制模块 `c<mod>` 与标准模块并列,同路径下客制优先。

> **新增文件类型会自动补齐**:白名单带版本号(记录在本地基线标记 `.tdict-mirror.ok` 里)。白名单升级后,下次「增量更新」检测到版本不一致会**自动转全量**一次,把新增类型的历史文件补齐(不必手动 `--full`);进度/日志会提示本次实际为全量。

> **备份/临时文件也排除**:`*.bak`、`*.bck`、`*.bck1`、`*.old`、`*.orig`、`*.tmp`、`*.swp`、`*~` 等(服务器侧 `find` 与本地清理共用同一份后缀列表)。历史上已拉取到本地的备份文件会在下次拉取时自动清理(无需 `--full`)。

镜像根目录配置在 config.json 顶层 `mirror` 键(**必须显式设置**,不设默认避免隐式落盘):

```json
{ "hosts": { "sshs": [ ... ], "activeEnv": "正式区" },
  "query": { "source": "local" }, "mirror": { "dir": "D:\\dev\\erp-src" },
  "bdldoc": { "dir": "D:\\path\\to\\docs\\bdl" } }
```

```bash
tt dict mirror dir                     # 查看当前镜像根
tt dict mirror dir D:\dev\erp-src      # 设置/更换镜像根(写入 config.json;相对路径存为绝对)
tt dict mirror pull                    # 默认环境(activeEnv/首条)增量更新
tt dict mirror pull 正式区          # 指定环境增量更新
tt dict mirror pull 正式区 --full   # 全量重建(整目录替换,删除服务器已不存在的残留)
tt dict mirror path                    # 打印镜像目录绝对路径,第一行即路径(AI 前往)
tt dict mirror path 正式区
```

- **增量机制**:服务器 TOP 下保留 `.tdict-mirror-<环境>.mark` 基线,默认只打包 `find -newer` 的变更文件(秒级);首次拉取与 `--full` 走全量(数百 MB~数 GB,视站点规模,服务器 gzip 打包 + SFTP 流式下载,完成后清理归档、保留 marker)。
- **打包根(TOP)解析**:T100 路径**不允许静态配置**——打包根按登录区域在服务器上动态探测获取(登录 zone → 环境脚本回读 TOP/ERP/COM),探测失败即报错(检查该环境 SSH 登录与 zone 设置)。
- AI 工作流:`tt dict mirror path` 拿目录 → 本地工具读码。

## BDL 语言参考文档(tt dict bdldoc)

项目内随带 **Genero BDL(4GL)语言参考文档**(markdown,已排除图片):`docs/bdl/`(约 5,200 篇,含 01 用户指南 ~ 19 版权、`llms.txt` 全文索引)。供 AI/开发者查询 BDL 语法、内置函数、fgldb 调试器命令等语言级问题——写 4GL 代码、读调试器输出遇到语言问题先查它,而不是猜。

文档存放路径记录在 config.json 顶层 `bdldoc.dir`,用命令查看/设置(**只改配置,不移动文件**):

```bash
tt dict bdldoc dir                    # 输出文档目录(第一行即绝对路径;未设置时给出提示)
tt dict bdldoc dir D:\path\to\docs    # 设置目录(写入 config.json;相对路径存为绝对)
```

- 默认值指向仓库内 `docs/bdl`(克隆/解压后按需用 `tt dict bdldoc dir` 校正)。
- 相关章节速查:BDL 语言基础/进阶在 `08_language-basics`/`09_advanced-features`;SQL 支持在 `10_sql-support`;画面/报表在 `11_user-interface`/`12_reports`;fgldb 调试器与工具在 `13_programming-tools`。

## 数据库连接与查询数据源（config.json）

合并后**只有一个配置文件**，三个工具共用，缺省 `%APPDATA%\T100\tt\config.json`（JSON，开发阶段明文，勿用于生产；可用 `T100_HOME` 整体改写父目录）。**环境清单是三个工具共用的**：它住在顶层 `hosts.sshs`，同一份列表同时驱动字典查询、调试与源码镜像 —— 加一台机器一次，三处都能用。数据库连接按 SSH 环境**一对一挂载**：每个环境在 `hosts.sshs[]` 内嵌一个 `db`（显式 host/port + `service`(oracle) 或 `database`(kingbase) + 账号列表）；服务器侧 sqlplus/ksql 工具路径自动探测，无需配置。顶层 `query` 键记录查询命令的默认数据源（与 hosts 平级）；另有顶层 `schemaVersion`、`listen`（统一 Web 服务地址）与 `debug`/`tdev`/`mirror`/`bdldoc`/`sync` 等节：

```json
{
  "hosts": {
    "sshs": [
      { "name": "正式区", "host": "10.0.0.1", "port": 22, "user": "youruser",
        "password": "yourpassword", "zone": "36", "topent": "99",
        "db": {
          "type": "oracle",
          "host": "10.0.0.1", "port": 1521,
          "service": "YOUR_SERVICE",
          "accounts": [
            { "account": "your_schema", "password": "your_password" },
            { "account": "your_schema2", "password": "your_password2" }
          ]
        } }
    ],
    "activeEnv": "正式区"
  },
  "query": { "source": "local" }
}
```

- **T100 路径不静态配置**:没有 `topDir`/`moduleRoots` 配置项——源码查找根/打包根等路径一律登录该环境后按其 zone 动态探测获取(登录 zone → 环境脚本回读 TOP/ERP/COM),探测失败即报错,请检查 SSH 登录与 zone 设置。
- **账号规则**：无"主账号"，所有账号都在 `accounts` 列表，不区分默认。客户端直连（查询数据源 / `db ping` / `db sync` / 连接测试）取列表**首项**；未收录的账号按"账号=密码"惯例兜底。
- `viaSsh`（可选）：客户端不可达 DB、但 DB 对 SSH 服务器可达时，经 SSH 端口转发再直连（本地起转发端口 → 驱动连 `127.0.0.1:本地端口`）；缺省远端取连接自身 host/port。
- 支持类型：`kingbase`（人大金仓，PostgreSQL 协议，pgx）、`oracle`（go-ora 纯 Go 驱动）。
- 配置文件查找优先级：`$TT_CONFIG`（旧名 `TDICT_CONFIG` 仍识别） > `--config` 参数 > 便携包内（exe 同目录有 `.portable` 标记时）> **统一用户目录** `%APPDATA%\T100\tt\config.json` > 旧位置（exe 同目录、当前工作目录）。首次运行会把合并前的旧位置配置（含 TDebug 的 `%APPDATA%\T100\tdebug\config.json`）**合并**到统一用户目录（旧文件只读、不删除，确认无误后自行清理）；预览用 `tt config migrate --dry-run`。位置与内容用 `tt config path` / `tt config show`（口令打码）看，`tt config validate` 校验。

### 查询数据源切换（在线 ⇄ 本地库）

`r.t/r.v/desc/scc/r.q/prog` 这些查询命令统一走同一查询接口（`db.Source`）：本地 SQLite 镜像与远程 ERP 库查的是**同一批表**，输出完全一致；命令层不感知数据源。解析优先级：

1. `--env <环境名>` / `--env local`：本次调用**强制**走该数据源；
2. `config.json` 顶层 `query.source`：**环境名** → 在线直查该环境；**`"local"`** → 本地 SQLite（`-d`/`TDICT_DB` 定位的 `erp_data.db`）；
3. **缺省 = 在线**：用默认环境（`hosts.activeEnv`）直查；一个环境都没配（新装/便携包首次）时才用本地库。

**两者不自动切换**：在线就是在线、本地就是本地——选在线时连不上就**直接报错**（不会偷偷改查本地），报错里会提示怎么切。

三种切换方式（等效）：

```bash
tt dict r.t dzea_t --env local              # ① 单次覆盖(命令行)
tt serve                                # ② 前端:「设置 → 查询数据源」下拉切换
# ③ 直接改 config.json 顶层(与 hosts 平级)
{"query": {"source": "local"}}             # 或 {"source": "正式区"} 指定在线环境
```

为什么缺省在线：**远程是最新且完整的数据**，本地副本要靠 `db sync` 更新（新机器/便携包上甚至还没有）。离线作业就在设置页切成「本地 SQLite」，或长期固定本地。

远程直查使用客户端驱动直连 `db.host:port`（账号取列表首项，与 `query.source` 的取值无关），不要求本地已 sync；金仓与 Oracle 均为完整支持（列表 `--kw` 过滤大小写不敏感，与本地一致）。查询是只读单条 SELECT，值经白名单/转义内联。

`db` 子命令（连接管理，与查询数据源独立；`--env` 语义同为环境名）：

```bash
tt dict db list                                  # 按环境列出挂载的库(类型/地址/账号数)
tt dict db ping [--env <环境名>]                # 验证连接可达(只读;缺省 activeEnv 环境的库)
tt dict db sync [--env <环境名>]                # 拉取该环境字典数据写入本地 SQLite(见前节)
tt dict db discover --env 正式区 --type oracle [--save]   # SSH 自动发现连接要素 → 预览或写入该环境 db
```

- `db discover` 给出"服务器视角"连接要素；客户端不可达时改 `db.host` 或补 `viaSsh`。
- `--env`（旧名 `--conn` 仍作别名）是全局 flag，查询数据源切换与 `db sync` / `db ping` 都用它，值取 `hosts.sshs` 里的环境名。

## 输出中的类型码

`tt dict r.t` 输出的取值含义。

### 表类型(r.t 的「类型」列:每张表在系统里的角色)

| 码 | 含义 |
|---|---|
| B | 基础数据 (Basic) |
| M | 主档 (Master) |
| T | 交易单头 (Transaction header) |
| D | 交易单身/明细 (Detail) |
| L | 多语言 (Language) |
| V | 提速档/视图 (View) |
| X | 系统/交叉 (Cross-reference) |
| H | 历史/暂存 (History/Temp) |

### 字段数据类型(r.t 字段表的「数据类型」列常见值)

| 码 | 含义 |
|---|---|
| `N004`, `N101` | 数字 (number) |
| `C003`, `C004`, `C105` | 字符 (varchar2) |
| `D001` | 日期 (date) |
| `Z012` | 时间戳 (timestamp) |
| `Z501` | 一般 flag (Y/N) |
| `Z509` | 短说明 |
| `Z510` | 长说明 |
| `Z521` | 文件编号 (Table 引用) |
| `Z522` | 字段编号 (Field 引用) |
| `Z504` | 人员编号 |
| `Z505` | 组织编号 |

## 项目结构

合并后本工具的代码不再是独立仓库，而是统一项目 TT 里的一个命令组（`tt dict`）。
完整布局见 [ARCHITECTURE.md](ARCHITECTURE.md)，这里只列字典相关的部分：

```
internal/dict/
  db/         本地 SQLite 数据源（Source 接口 + 各查询族）
  live/       远程库数据源（金仓 / Oracle，同一 Source 接口）
  dbsync/     字典表同步（Family 表是"哪个命令需要哪几张表"的唯一出处）
  server/     字典页的 HTTP 接口（由统一服务挂在 /dict/api/ 下）
internal/cli/dict/   cobra 的 tt dict 命令组
internal/host/       ★ 远程服务器共享层（与调试侧共用）：SSH/PTY、环境探测、源码镜像
internal/config/     ★ 统一配置层：环境清单、query/mirror/bdldoc/sync 各节
internal/erpdb/      Oracle(go-ora) / Kingbase(pgx) 连接器
internal/sshtun/     SSH 端口转发隧道
internal/output/     表格 / JSON / CSV 输出（CJK 宽度感知）
web/dict/            前端(React 18 + Vite 6 + Tailwind v4)
```


## 技术栈

| 项目 | 选择 |
|---|---|
| 语言 | Go 1.26 |
| SQLite 驱动 | [modernc.org/sqlite](https://modernc.org/sqlite)（纯 Go，无 CGO） |
| Kingbase 驱动 | [jackc/pgx](https://github.com/jackc/pgx)（纯 Go，PostgreSQL 协议） |
| Oracle 驱动 | [sijms/go-ora/v2](https://github.com/sijms/go-ora)（纯 Go） |
| SSH/SFTP | [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) + [pkg/sftp](https://github.com/pkg/sftp) |
| CLI 框架 | [cobra](https://github.com/spf13/cobra) |
| Web 前端 | React + Vite + Tailwind（SSH/数据库配置页，产物 `go:embed` 嵌入） |
| 输出 | `text/tabwriter` + `encoding/json` |
