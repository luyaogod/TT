# internal/cli/devdata — `tt gen`

为 model 准备原材料的命令组。领域实现在 [../../devdata](../../devdata)，本包只做参数、
输出与退出码（与 Web 共用同一份实现，所以 `tt gen fetch` 与设置页拉出来的 sha 完全一致）。

> **与 `tt file mirror` 不是一回事。** 那个是"读代码用的镜像"（全量、落 `srccache/`、
> 会被 `tt cache` 清理）；这里是"给 model 的原材料"（按程序成套、带 manifest 指纹、
> 清理由 `tt gen clean` 负责）。概念对照见 [../../devdata/README.md](../../devdata/README.md) 的
> "与读侧的区别"。

## 命令

环境一律点名（`--at <环境>` 或 `--env <环境>`），没有默认环境。
`--at <环境>@<企业>` 里的企业决定规格库用哪个数据库账号（不给就用账号清单首项）。

> **当前状态（2026-11-21）**：镜像功能保留但**暂不启用**，`mirror` 不会被自动执行；
> `fetch` 独立可用，它需要的版次（`dzaa002`）**默认在线查**，本地共用资料只作兜底。
> 镜像要镜哪些表/镜多大，等 model 的能力边界确定后再定（plan ADR-08）。

### `tt gen status [--probe]`

共用资料 + 各程序的状态。**默认不出网**（只读本地 manifest 与文件，打开即快）；
`--probe` 才向服务器核对 mtime/大小。

```bash
tt gen status --at 测试环境
tt gen status --at 测试环境 --probe --human
```

JSON 载荷：`{env, dir, present, dict, mdl, progs[], jobs[], warnings[]}`。
程序行带 `stale` 与 `staleReasons[]`（`missing-local` / `sha-mismatch` / `remote-newer` /
`remote-gone` / `version-changed`）。

### `tt gen mirror [--only dict|mdl] [--force]`

> **暂不启用**（见文件开头的「当前状态」）。功能保留：等 model 定型后用它按确定的范围备料。

把本环境的规格表（53 张 `DZ*_T` → SQLite）与模板树（`$TOP/erp/mdl`）拉到
`devdata/<环境>/`。**幂等**：已同步过就跳过（`skippedPull=true`，不报错）；`--force` 重拉。
`--only` 只刷一半，另一半的 manifest 条目原样保留。

进度打在 **stderr**（每 400ms 最多一行），stdout 只留最终 JSON/表格。
Ctrl+C 可中断 —— 中断后临时目录会被清掉，正式目录保持拉取前的状态。

```bash
tt gen mirror --at 测试环境
tt gen mirror --at 测试环境 --only dict
tt gen mirror --at 测试环境 --only mdl --force
```

单表超 200 万行会被跳过并记进 `warnings`（实测 `DZBA_T` 781 万、`DZAZ_T` 462 万，
都是生成器的运行时产物/计数，不是入参）。缺表进 `missing`，不判失败。

### `tt gen fetch <程序> [--module <模块>] [--with-design] [--form] [--force]`

按程序拉一套：`tab`（蓝图）、`tsd`、`tap`、`tgl`（骨架，含 `tgx/tgx2`）、`4gl`（产物）、
`inc/*_{mask,rep}.4gl`。缺必需项（`tab`/`tgl`/`4gl`）**整次失败、不留半套**；
缺可选项只进 `missing[]`。

模块默认从程序名前三位推断（`axmt500_wf` → `axm`），推断不出时用 `--module` 指定。
已拉过的程序需要 `--force`（整包替换，不合并）。

```bash
tt gen fetch axmt500_wf --at 测试环境
tt gen fetch axmt500_wf --at 测试环境 --with-design --form
tt gen fetch cxm999 --at 测试环境 --module cxm --force
```

`--with-design` 把本地工作区（`hosts.sshs[].workspace`）里的 `<prog>*.tzs` 复制进
`design/`（只归档、不解析）；`--form` 才拉 `.42f`（二进制，体积大）。

### `tt gen pack <程序> [--identity c|s] [-o <文件|目录>] [--ver 1.0] [--force]`

把**已拉到的材料**打成 `.tzc`，命名同设计器下载的形态：`<程序>(c).tzc` / `<程序>(s).tzc`。

**纯本地，不连服务器**。设计器的「下载程式包」（`adzp050_download_code_pack`）读源码看下来
只是 `sadzp050_zip_compress_to_code_zip` + `sadzp050_zip_download_code_zip`：不重产、不读库，
包内成员就是 `dzx` 里的 `tap`/`tgl` 加 `4gl` 与一个恒为 `1.0` 的 `ver`（5 个真包实测 crc
都是 `6493f4df`）。既然 `fetch` 已把这些文件拉到本地（还带 sha256），这一步就没有理由再上服务器。

包的**形态**必须像真包：`flags` 不置 bit 3、CRC/大小内联在局部头、每个条目带
`UT`+`ux` extra、外部属性 0664（`ver` 为 0775）。所以它走
[`pkgfile.WriteZip`](../../dev/pkgfile/zipcreate.go)，而不是 `archive/zip.Writer`
（后者会置 bit 3 并补数据描述符，设计器直接报「Data descriptor signature not found」）。

```bash
tt gen pack axmt500_wf --at 测试环境                 # → devdata/<环境>/pack/axmt500_wf(c).tzc
tt gen pack axmt500_wf --at 测试环境 --identity s    # → axmt500_wf(s).tzc
tt gen pack axmt500_wf --at 测试环境 -o D:\tmp\      # 落到指定目录
tt gen pack axmt500_wf --at 测试环境 --force         # 覆盖同名包
```

身份（`c`/`s`）缺省从 `.tap` 根元素的 `identity="…"` 读，读不出按客制 `c` 并给 warning。
缺 `tap`/`tgl`/`4gl` 任一项 → 退 4 并说明"材料不成套"（未被设计器接管过的标准程序本来就
没有 `.tap`，这是正常态，不是拉取失败）。产出的包会被自己的读取器重新打开自检一遍。

用途：给 `tt dev tzc` 造夹具、留可回放的快照。验证方式：

```bash
tt gen pack axmt500_wf --at 测试环境
tt dev tzc export devdata/测试环境/pack/axmt500_wf(c).tzc -o ws/axmt500_wf
tt dev tzc verify ws/axmt500_wf
```

### `tt gen map <程序> [--field <字段>|--custom]`

建**蓝图 ↔ 代码**的索引，落到材料目录的 `link.json`：

```
.tab 页面/字段  →  .tap 槽位（加点）  →  .4gl 行号区间  →  切片 aXX(Version:N)
```

**纯本地**（材料先 `fetch`），不连服务器。为什么值得有它：一个 69,254 行的 `.4gl` 里，
`.tap` 声明过内容的只有 1,483 个槽位、真正人写的客制内容约 9,000 行（13%），
而单个需求往往只涉及一两百行。没有索引时只能 grep 整份代码，且分不清"这句是客制还是
模板正文"（后者重产就没了）。

身份规则（与生成器一致，实测钉下来的）：

- 槽位名 = `<功能>.<位置>.<页 id>.<字段>`；**页用的是 `.tab` 的 `page_id`，不是页签序号**
  （页签序号是第 12 个，点名里却叫 `page73` —— 两者差一位是常见的坑）；
- 客制点可挂到标准槽位上（`mapping ≠ name`，如 `b_fill2.before_fill10` → `update_b.other`）；
- 重名合法（实测 `construct.c.page1.xmdc027` ×3），键是"名字 + 出现次序"；
- `function.*` 这类客制函数在 `.4gl` 里**没有 `#add-point` 标记**，按函数名定位；
- `.tap` 标记行的 `name` 可能被截断（实测 `name="chk_`），此时以 `{<point name="…"/>}` 洞为准。

```bash
tt gen map axmt500_wf --at 测试环境                          # 摘要 + 检查（默认，约 2.7 KB）
tt gen map axmt500_wf --at 测试环境 --field xmdduc001        # 这个字段的点在哪几行
tt gen map axmt500_wf --at 测试环境 --custom --human         # 客制点清单
tt gen map axmt500_wf --at 测试环境 --orphans --human        # 写了不生效的孤儿（带名字）
tt gen map axmt500_wf --at 测试环境 --full > /tmp/link.json  # 整份索引（脚本用）
```

> **默认不吐整份索引**：完整 doc 有 800+ KB（实测 622 KB JSON），直接灌 stdout 对调用方
> （尤其 AI）是灾难。默认给摘要与检查，索引本体留在 `link.json`；脚本要全量用 `--full`。

`--field` 的输出（真实数据）：

```
点位                            来源    行区间                         声明行数
construct.c.page73.xmdduc001  客制    axmt500_wf.4gl:18269-18282  8 行
construct.b.page73.xmdduc001  模板    axmt500_wf.4gl:18287-18289  0 行
input.a.page73.xmdduc001      客制    axmt500_wf.4gl:44107-44132  23 行
input.c.page73.xmdduc001      客制    axmt500_wf.4gl:44217-44243  18 行
```

检查项（每条带 `level`，写进 link.json 的 `checks[]`）：

| id | 含义 |
|---|---|
| `sections-pair` | 两侧 section 同名同序（对不上 = 材料不是同一版） |
| `slots-pair` | 槽位一一对应；只在 tap = 4gl 旧，只在 4gl = 生成器自加（多为注释掉的重复点） |
| `decls-orphan` | 声明了但骨架无槽位、4gl 也找不到 = **写了不生效**（字段/页被删后的残留） |
| `decls-unmarked` | `function.*` 以函数形式落地，无标记，按函数名定位 |
| `slice-version` | 产物里的模板版次 vs 当前模板（需先 `tt gen mirror --only mdl`） |
| `freshness` | `.4gl` 比 `.tap` 旧 = 没跑 F7（产生程式），行号有效但内容未必生效 |
| `blueprint` | 版型 / 页数 / `cusver`（`0000` 即从未客制） |

⚠️ 有 `.tap` 才谈得上"客制还是标准"；从未被设计器接管过的标准程序**服务器上就没有 `.tap`**，
此时退化为"只用 `.4gl`"：section / 槽位 / 切片 / 字段照旧，`points[].g4Only=true`。

### `tt gen grep <程序> <子串>`

在 `.4gl` 里检索，并把**每条命中标上归属**：落在哪个加点、客制还是标准（重产会不会保留）、
哪个 section、哪个切片。这是"拿到一个不认识的符号，第一反应是搜一下"的那条命令 ——
比 grep 多的正是"归属"这一层（跨 `.tab/.tap/.4gl` 的对应关系，grep 做不了）。

```bash
tt gen grep axmt500_wf "SET(" --at 测试环境
tt gen grep axmt500_wf xmdducsite --at 测试环境 --human
tt gen grep axmt500_wf "cl_chk_exist" --at 测试环境 --limit 200 --case
```

输出列：`位置（文件:行）| 来源 | 范围（加点内/正文）| section|切片 | 代码`。

### `tt gen show <程序> <点名|字段>`

在 link.json 里找点位，并把**代码内容**打出来。匹配规则分两种（这是试用人踩出来的坑）：

- **query 里带 `.`** → 当点名处理，**只接受精确匹配**。否则 `input.a.page73.xmdducseq`
  （一个真正丢失的孤儿）会静默命中 `…xmdducseq1`，让人以为"写了却没生效的点其实有代码"。
- **不带 `.`** → 当字段名处理，允许后缀/子串，输出里用 `提示` 行和 JSON 的 `match` 说明是哪种。

```bash
tt gen show axmt500_wf input.a.page73.xmdduc001 --at 测试环境
tt gen show axmt500_wf xmdduc001 --at 测试环境 --human
tt gen show axmt500_wf insert_b.m_insert10 --at 测试环境 --max 60
```

每条结果带：来源（客制 / 标准 / 生成器自加 / 未落地）、`.4gl` 行区间、section、`slot`、
段内出现的切片代号、以及内容（`--max` 截断）。

### `tt gen at <程序>:<行号>`

反查一行属于哪个加点 / 哪个 section / 哪个切片，并给一句**安全性结论**：

```
行号   44107
结论   在客制加点 input.a.page73.xmdduc001 内（第 44107-44132 行）——改这段重产会保留
点位   input.a.page73.xmdduc001     客制，第 44107-44132 行
section axmt500_wf.input.body       第 25853-44448 行（.tap 第 57260 行）
切片   a17                          Version 3，标记在第 44109 行
L44107   #add-point:AFTER FIELD xmdduc001 name="input.a.page73.xmdduc001"
L44108   IF NOT cl_null(g_xmdc73_d[l_ac].xmdduc001) THEN
```

落在 section 正文（不在任何加点内）时会明确警告"**模板生成的框架代码，手改会被重产覆盖**"。
编译报错（`fglcomp: 44021`）、断点行、review 时看到的一行，都直接用它。

### `tt gen fields <程序> [--pages] [--page <page_id|序号>] [--q <子串>]`

列蓝图字段（只列 `.tab` 里真实存在、且代码里有对应加点的），每个字段带它的全部点位：

```
页签/page_id  record      字段        点位数  点位
12/73        s_detail73  xmdduc001   7      construct.c… construct.b… construct.a… input.a… input.b… input.g… input.c…
```

- `--page` 同时接受 **page_id（73）** 与 **页签序号（12）**；传序号时会明说对应关系。
  这解决"客户说第 12 个页签，代码里却叫 page73"。
- `--pages` 列页面本身：`序号 / page_id / record / 栏位数 / 有逻辑的字段数`。
- 注意这是"**有逻辑**的字段"清单，不是"页面上有几个栏位"：纯数据字段不出现。
  总栏位数看 `gen map` 摘要里的 `columnsTotal`（实测 412 栏位 / 317 个有逻辑）。

### `tt gen ls`

只列已拉取的程序（小表）：大小、拉取时间、是否过期、缺项。等价 `status` 里的程序清单。

```bash
tt gen ls --at 测试环境 --human
```

### `tt gen check <程序>`

不重拉，只与服务器核对并给出**过期明细**（哪一条触发了、为什么）。

```bash
tt gen check axmt500_wf --at 测试环境 --human
```

### `tt gen clean [--prog <程序>] [--yes]`

删除开发资料。**默认 dry-run**（只列出将删的内容与体积），`--yes` 才真删。

为什么不跟 `tt cache` 一起清：见 `tt gen --help` 与 [../../devdata/README.md](../../devdata/README.md)
的 ADR-02 —— 这份材料删了要重新拉（几十 MB + 数据库查询）。

```bash
tt gen clean --at 测试环境                    # 看会删什么
tt gen clean --at 测试环境 --prog axmt500_wf  # 只看这个程序
tt gen clean --at 测试环境 --prog axmt500_wf --yes
```

## 输出形态

默认 JSON 信封（`ok/source/env/…` + `data`），`--human` 出小表。两种形态**内容一致**：
表格只换展示形态，不偷偷少字段。`--csv` 与其它命令组同一套规则。

## 退出码

| 码 | 含义 |
|---|---|
| 0 | 成功（含 `mirror` 的"已同步、跳过"） |
| 2 | 用法或身份错：环境不在清单、程序名非法、模块推断不出、已拉过没 `--force`、同类任务在跑 |
| 4 | 未拉取 / 目标不存在 / 材料不成套（`check`、`clean`、`pack` 的 404 语义） |
| 5 | 拉取失败：网络、库、写盘；或用户中断（`CANCELLED`） |

## 帮助与验收

```bash
tt gen --help
tt gen status --at-list
tt gen fetch --at 测试环境 axmt500_wf   # 真实拉一次（约 5 秒，4.9 MB）
tt gen check axmt500_wf --at 测试环境
tt gen fetch ../x --at 测试环境         # 退 2，程序名非法
tt gen check nope --at 测试环境         # 退 4，未拉取
tt gen pack axmt500_wf --at 测试环境    # 出 pack/axmt500_wf(c).tzc（纯本地，131ms）
tt gen pack aimi100 --at 测试环境       # 退 4，材料不成套（标准程序没有 tap）
```

测试：`go test ./internal/cli/devdata -count=1`（输出形态 + 错误码映射）。
