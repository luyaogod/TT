# T100 设计 → 代码生成 · 架构研究

> 版本：v1.0（2026-11）
> 语料：`01-server/`（服务器生成器 4GL）、`02-launcher/`（r.* 脚本）、`03-templates/`、`04-designer/`（反编译源码子集）
> 结论均标注证据位置（`文件:行`）；推断性内容标 **〔推断〕**。

---

## 0. 结论摘要（TL;DR）

1. T100 是一套 **DSL 驱动的、主动式（active）代码生成平台**：设计器里画的是「规格模型」，程序代码由模型派生，随时可重产。
2. 生成链是 **四段式**：
   **上传解包 → 蓝图（M2M）→ 模板展开（M2T）→ 填孔 + 编译**。
3. 生成器的「智能」集中在三处：
   - **判型**（`sadzp030_tab_relation`）：看结构决定用哪套骨架（i01/t01/i02/t02/i07…）
   - **关系求解**（`sadzp030_tab`）：把 PK/FK/上阶/页组翻译成 `.tab` 蓝图节点
   - **模板切片**（`adzp150` + `sadzp150_*`）：把蓝图套进「主样板 + 段落切片」展开成 `.tgl`
4. 客制代码只活在 **`.tap`（加点 / add-point）** 里，等价于 MDD 里的 **受保护区域（protected region）/ Generation Gap**；重产会覆盖其余一切。
5. 真正的「代码生成」不在设计器里，也不在 `adzp080` 里，而在 **`adzp150`（样板展开）+ `compose_path`（填孔）** 两个黑盒里。

---

## 1. 文件谱系与术语

### 1.1 设计器侧（.tzs）

| 文件 | 内容 | 备注 |
|---|---|---|
| `.tzs` | 表单包（zip） | 内含 `.tsd` `.4fd` `.bdx` `ver` |
| `.tsd` | 表单**规格**（XML：页签/字段/表关联/开窗/校验） | 生成的源头模型 |
| `.4fd` | 表单**定义**（编译后的组件树，Genero Form Definition） | 由 `.tsd` 产出 |
| `.bdx` | 画面索引/辅助 | — |

### 1.2 程序侧（.tzc）

| 文件 | 内容 | 备注 |
|---|---|---|
| `.tzc` | 代码包（zip） | 内含 `.4gl` `.tgl` `.tap` `ver` |
| `.tgl` | **模板**：带 section 与 `{<point>}` 孔位的程序骨架 | 3 层结构：`{<section>}` / `#add-point:` / `{<point>}` |
| `.tap` | **加点**：每个 point 里用户写的真实代码 | 客制代码的唯一合法居所 |
| `.4gl` | **最终程序**（由 `.tgl` + `.tap` 填孔而成） | 编译输入 |

### 1.3 生成过程中的中间物

| 文件 | 阶段 | 说明 |
|---|---|---|
| `<prog>.tab` | 蓝图 | XML：结构 + 型态 + 参数（M2M 产物） |
| `<prog>.tgx` | 展开 | 主样板代入后的程序文本 |
| `<prog>.tgx2` | 整理 | 加 section 结构（adzp175） |
| `<prog>.tgl` | 成品模板 | section + 孔位（adzp177） |
| `<prog>.4gl` | 最终 | 填孔后（compose_path） |
| `.42f/.42s/.42m/.42r` | 编译产物 | gen42f / gen42m / gen42r |

### 1.4 规格数据库（dz*，设计器上传后写入，生成器从这里读）

| 表 | 用途（据代码引用） | 关键列 |
|---|---|---|
| `dzaa_t` | 规格主档 | `dzaa001` 程序、`dzaa002` 版次、`dzaa003` 识别码、`dzaa009` 客制标示 |
| `dzag_t` | **表关联** | `dzag004` 上阶表、`dzag005` main、`dzag008` pk、`dzag009` fk_detail、`dzag010` fk_master、`dzag013/014/015` 三层二显 |
| `dzeb_t` | 栏位定义 | — |
| `dzfq_t` | 程式样板型态 | `dzfq001` 名称、`dzfq003` 版次、`dzfq004` 程序、`dzfq005` 型态 |
| `dzfi_t` | 程序基本资料 | `dzfi001` 程序代号 |
| `dzax_t` | 固定参数 | `dzax007` 参数个数 |
| `dzaf_t` | 程序/表格使用分析 | `dzaf001` 程序、`dzaf005` 型别 |

### 1.5 `r.*` 命令（`02-launcher/`）

| 命令 | 脚本 | 实际执行 |
|---|---|---|
| `r.r` | `r.r` | `$FGLRUN $AZZ/42r/azzp000.42r "$@"`（T100 作业入口） |
| `r.c3` | `tools/r.c3` | `$FGLRUN $FBIN/42m/compose_path <mod> a:<prog> b:'' c:<ver> d:<0\|2>` |
| `r.c` | `tools/r.c` | `$FGLRUN $FBIN/42m/gen42m <prog> <arg2>`（4gl → 42m） |
| `r.f` | `tools/r.f` | `$FGLRUN $FBIN/42m/gen42f "$@"`（4fd → 42f/42s） |
| `r.l` | `tools/r.l` | `$FGLRUN $FBIN/42m/gen42r.42r <arg1> <arg2>`（链接 42r） |

---

## 2. 全景流水线

```
┌─────────┐  Telnet: r.r adzp080 '<pkg>' <mod> SPECRC3|CODE '<dir>'
│ 设计器  │ ───────────────────────────────────────────────┐
└─────────┘                                                ▼
                                              ┌──────────────────────┐
                                              │ adzp080 上传/产生主控 │
                                              └──────┬───────────────┘
                        ┌────────────────────────────┴──────────────┐
                        ▼                                           ▼
              ┌──────────────────┐                        ┌──────────────────┐
              │ adzp060 上包/解包 │                        │ adzp040 产生编排  │
              │ unzip → $TEMPDIR  │                        └──────┬───────────┘
              │ → $ERP/<mod>/4gl  │                               │
              │ → dzag/dzeb… 落库 │                               │
              └──────────────────┘                               │
        ┌────────────────────────────────────────────────────────┤
        ▼                            ▼                           ▼
┌──────────────────┐        ┌──────────────────┐        ┌──────────────────┐
│ sadzp030_tab     │        │ adzp150 样板展开  │        │ r.c3 → compose   │
│ 规格 → .tab 蓝图  │ ─────▶ │ .tab+模板 → .tgl  │ ─────▶ │ .tgl+.tap → .4gl │
└──────────────────┘        └──────────────────┘        └────────┬─────────┘
                                                                 ▼
                                                     ┌──────────────────────┐
                                                     │ r.f/r.c/r.l          │
                                                     │ gen42f/gen42m/gen42r │
                                                     └──────────────────────┘
```

### 2.1 客户端只发一条命令

`02-launcher/` 之外，设计器的发送逻辑在 `04-designer`（`ConnectionManager.UploadSpec/UploadCode`）：

```
r.r adzp080 '<包名>' <模块> <SPEC|SPECRC3|CODE|GCODE|CSPEC|RSPEC> '<本地目录>' [参数]
```

- `SPECRC3` = 上传规格并产生程式（F7）
- `CODE` = 上传程式（tzc）
- 之后的全部工作在服务器，客户端只接收回传日志

### 2.2 两种主流程

**A. 规格 + 产生（SPECRC3）**

```
adzp040_start(106)
 ├─ update_sd_data(700)            规格入库（dzaa/dzag/dzeb…）
 ├─ compile_4fd(1016)              r.f → gen42f（4fd → 42f/42s）
 ├─ (非 M 类) r.r azzp191          重产 42s
 ├─ generate_tab(561)              sadzp030_tab_gen → .tab
 ├─ r.r azzp188                    程序关联（忽略错误）
 ├─ generate_tgl_4gl(858)          r.c3 '<prog> '' A' → .tgl + .4gl
 ├─ (GCODE) generate_rdd(953)      r.c <prog> rdd
 └─ chk_4gl_lines(1654)            4gl 行数异常检查
```

**B. 程式包上传（CODE）**

```
adzp040_start(106)
 ├─ update_tap(791)                .tap 加点入 dzf*（只在 .tgl 需要重建时）
 ├─ (is_regen / 模板换代 / K 类) generate_tab(561)
 ├─ generate_tgl_4gl(858)          r.c3 '<prog> '' <ver> <0|2> <dgenv>'
 └─ chk_4gl_lines(1654)
```

`r.c3` 的第 4 个参数（`adzp040.4gl:906-913`）：

| d | 含义 |
|---|---|
| `0` | 全部：产生 + 组合 + 编译 + 链接 |
| `1` | 只产生与组合 |
| `2` | 只做组合器以后：编译 + 链接（**日常 tzc 上传走这个**） |
| `3` | 组合器以后但不编译不链接 |
| `A` | SPECRC3 专用，一次做完 |

---

## 3. 阶段一：上传与解包（adzp080 / adzp060）

### 3.1 adzp080 的两件事

| 函数 | 行 | 做什么 |
|---|---|---|
| `adzp080_upload_package` | 454 | 组 `r.r adzp060 <prog> <mod> <doctype> <client_path>`（564-570），经 `adzp080_exec_prog` 执行 |
| `adzp080_build_package` | 577 | 组 `r.r adzp040 <prog> <mod> <doctype> <specver> <codever> <std> … <lang>`（598-611） |

`adzp080_exec_prog(659)` 用 `base.Channel.openPipe(cmd,"r")` 抓子进程输出（693-699），
并扫描 `ERROR/FAILED` 判定失败（763-766），把每行打上 `[COMMAND]/[MESSAGE]` 回传客户端。

### 3.2 adzp060 解包

- `.tzs/.tzc` 用 `unzip -o` 解到 `$TEMPDIR`，再落到 `$ERP/<mod>/4gl|4fd`（`sadzp060.4gl:458-530`）
- 同时把包内的**规格节点**写进 `dz*` 表——这是生成器真正的「模型输入」。

### 3.3 关键：`tbl` 节点 → `dzag_t`

`sadzp060_1.4gl:897-926` 把 `.tsd` 的 `<tbl>` 属性原样写进 `dzag_t`：

```
dzag004 = parent            ← 连线的另一头（上阶表）
dzag008 = pk
dzag009 = fk_detail
dzag010 = fk_master
dzag013 = upper_table       ← 三层二显
dzag014 = upper_key
dzag015 = this_key
```

> 这解释了「连线管树、三层二显管列」的分工：两者在数据库里是**不同列**。

---

## 4. 阶段二：蓝图生成（sadzp030_tab，M2M）

### 4.1 输入与输出

- 输入：`dzaa/dzag/dzeb/dzax/dzfq` 等规格表
- 输出：`<prog>.tab`（XML），节点树形如

```xml
<body detail="Y" master="qcbd_t" pk="…" fk="…" 32wc="qcbdseq=qcbeseq">
  <sql id="forupd_sql_detail" query="SELECT … FROM qcbe_t WHERE … FOR UPDATE"/>
  <sql id="detail_append" type="single" table="…"/>
  …
</body>
```

属性含义：

| 属性 | 含义 | 生成到 |
|---|---|---|
| `pk` / `fk` | 主键、外键（→主档） | 键数组、JOIN `ON` |
| `detail="Y"` / `master` | 这是某表的子表，上阶是谁 | 嵌套数组、索引、页签归属 |
| `32wc` | **三层二显**算出的额外条件 | JOIN 的 `AND` |
| `sql@forupd_sql_detail` | 每表的 `FOR UPDATE` 语句 | `lock_b`/`b_fill` 的游标 |

### 4.2 判型（决定骨架）

`sadzp030_tab_relation.4gl:830-956`：

| 结构 | 型态 |
|---|---|
| 有 browse(Table) + 单档 | `i01` |
| 有 browse + 双档 | `t01` |
| 无 browse 的单/双档 | `i10` / `i09` |
| 假双档 | `i07` / `i12` |
| 树状单档 / 六阶树 / 树状主从 / 树状单档+单身 | `i05` / `i13` / `i04` / `i08` |
| 单档多栏 / 双档多栏 | `i02` / `t02` |
| Web Service / WS 元件 / 报表 | `m00` / `p00` / `r01` |

名义型态（客制 `c01a/c02a/c03a/c04a`）到实际样板的映射写在 `01-server/generator/adzp150.4gl:412-442`。

### 4.3 三段生成

| 文件 | 负责 |
|---|---|
| `sadzp030_tab_ma.4gl` | 单头（master）段落 |
| `sadzp030_tab_de.4gl` | 单身/子表（detail）段落 |
| `sadzp030_tab_relation.4gl` | 表关联、PK/FK、判型 |
| `sadzp030_tab_file.4gl` | 文件路径/输出 |
| `sadzp030_tsd.4gl` | 反向：规格 → `.tsd`（设计器下载用） |

### 4.4 三层二显的「减法」

`sadzp030_tab.4gl:3466 sadzp030_tab_gen_3layer2show()`：

```
32wc = upper_key − (上阶表对主档的 FK)
```

先用 `sadzp030_tab_relation_fk(上阶表, 主档, 2)` 取上阶表对主档的外键列，
把与 `upper_key` 重复的标 `checked`（3535-3543），剩下的才拼成 `A = B AND …`。

所以 `xmdd` 的 `upper_key=(ent,docno,seq)` 最后只输出 `AND xmdcseq = xmddseq`。

---

## 5. 阶段三：模板展开（adzp150，M2T）

### 5.1 两层模板

| 层 | 文件 | 作用 |
|---|---|---|
| 主样板 | `$ERP/mdl/code_<实际样板>.template` | 整支程序骨架（i01/t01/…），内含符号 |
| 切片 | `$ERP/mdl/slice/code_<aXX>.template` | 段落级小样板（a01/a02/a03/a17/a43/a77…） |

`adzp150` 主流程（`adzp150.4gl:85 MAIN`）：

```
读 $ERP/<mod>/dzx/tab/<prog>.tab          (224-227)
 └─ read_and_replace(960)
      读 mdl/code_<type>.template          (984)
      写 <mod>/dzx/tgl/<prog>.tgx          (995-997)
 └─ adzp175 → .tgx2                        (275)
 └─ adzp176 → section 解析                 (285)
 └─ adzp177 → <prog>.tgl                   (297)
 └─ adzp155 → mask（特定样板）              (305)
```

切片加载：`sadzp150_mdl.4gl:12 adzp150_make_slice()`，
路径 `$ERP/mdl/slice/code_<mdl>.template`（24-33），符号分发在 47-111
（`#pages` / `#tables` / `#master_keys` / `#detail_keys` / `#mdls` …）。

### 5.2 样板白名单与版本

```4gl
# adzp150.4gl:401
i00,i01,i02,i03,i04,i05,i06,i07,i08,i09,i10,i12,i13,t01,t02,
c01a,c01b,c01c,c02a,c02b,c02c,c03a,c03b,c03c,c04a,p00,p01,p02,n00,m00,r01,q00
```

- 不在白名单 → `ERROR(2):目前尚未支持此種樣板檔` 并退出（403-405）
- 实际样板由名义样板映射得到（412-442），例如 `c03a→t01`、`i10→i01`、`i12→i07`、`i13→i01`
- 模板版本用 MD5 校验：`sadzi100_template_chk(type_id, cs_prog_ver)`（`adzp150.4gl:467`；实现在 `01-server/common/sadzi100_extract.4gl:162`）

### 5.3 切片清单（`03-templates/slice/`，共 118）

| 族 | 数量 | 例 |
|---|---|---|
| `a00..a77` | 77 | a01 BEFORE FIELD、a02 AFTER FIELD、a03 开窗、a04 ON CHANGE、a05 查重、a07/a08 开窗段、a17 校验带值、a43 按钮动作、a77 模糊查询 |
| `d01..d10` | 10 | 明细/资料段 |
| `qs01..qs27` | 27 | 查询段 |
| `ss01..ss04` | 4 | 子画面段 |

模板是**参数化**的：`${mdl_var1}` 单值、`#mdls - Start -` 循环；全部参数来自 `.tab`。

---

## 6. 阶段四：填孔与编译（compose_path）

### 6.1 `.tgl` 的结构

```
{<section id="axmt500_wf.global" type="s">}
  #add-point:填寫註解說明 name="global.memo"
  {<point name="global.memo" edit="s"/>}
  #end add-point
  …生成段…
{</section>}
```

- `section` = 框架段（可选解锁，解锁后不再随规格重产）
- `point` = 孔位，内容来自 `.tap`

### 6.2 填孔

`r.c3` → `compose_path`（`02-launcher/tools/r.c3`）：

```sh
$FGLRUN $FBIN/42m/compose_path $1 a:$2 b:$3 c:$4 d:$5
```

把 `.tgl` 的每个 `{<point>}` 用 `.tap` 对应内容替换，写出 `.4gl`；
`d` 参数决定是否继续走编译与链接（见 §2.2）。

### 6.3 编译与链接

| 步骤 | 命令 | 二进制 |
|---|---|---|
| 画面 | `r.f <prog> <tiptop>` | `gen42f` |
| 4gl | `r.c <prog> [tiptop\|TEMP\|ALL\|rdd]` | `gen42m` |
| 链接 | `r.l <prog>\|<mod>` | `gen42r.42r` |

错误回传链路：子工具 stdout → `cl_cmdrun_openpipe`（`adzp040.4gl:1361`）→
`adzp080` 打 `[COMMAND]/[ERROR]` → 设计器「进阶功能 → 联机记录」。
编译错误行号对应 `$ERP/<mod>/4gl/<prog>.4gl`（`adzp040_execute_program` 先 `chdir`，1305-1329）。

---

## 7. 表关系 → CRUD 模型

> 详细图解见 `diagrams/relation-to-crud.html`、`diagrams/fk-vs-upper.html`。

### 7.1 三件套 + 五支子程序

| 关系 | 生成物 |
|---|---|
| main 表 | 单头 `RECORD`（如 `g_xmda_m`）+ `s_browse` |
| 明细表 | `DYNAMIC ARRAY`（`g_xmdc_d`）+ 旧值备份 `_t` / `_o` |
| 键 | 运行时键数组 `gs_keys[]` / `gs_keys_bak[]` |
| 行为 | `insert_b` / `update_b` / `delete_b` / `key_update_b` / `key_delete_b`，外加 `lock_b` / `unlock_b` / `b_fill` |

每张表的 SQL 只出现一次，靠 `ps_keys[]` 与 `ps_page`（页组）复用。

### 7.2 页组与索引

- **页组**：`ls_group = "'1','2','8',"` → `IF ls_group.getIndexOf(ps_page,1) > 0`，
  同组表在 insert/delete 里一起处理
- **索引**：`g_detail_idx` / `g_detail_idx2` / `g_detail_idx3` 按**深度**分配；
  同深度的平行表共用一个索引

### 7.3 FK、upper、三层二显

| | 回答 | 存哪 | 生成物 |
|---|---|---|---|
| FK（fk_master/fk_detail） | 属于哪张单 | `dzag009/010` | JOIN `ON`、键数组 |
| upper（三层二显） | 属于哪一笔上层记录 | `dzag013/014/015` | JOIN `AND`（32wc） |
| parent（连线） | 树上挂谁 | `dzag004` | 第几层、索引、页签归属 |

**正常三层 vs 三层二显**（对用户操作的影响）：

| | 正常三层（parent=中间层） | 三层二显（parent=主档 + upper） |
|---|---|---|
| 页签取数 | 当前父行（`WHERE docno AND seq`） | 整单（`WHERE docno`） |
| 程序结构 | 嵌套数组 + `idx2/idx3` | 平铺数组 + `idx` |
| 页签内新增 | 可以，父键自动 | 通常禁用，走子视窗 |
| 代表 | `qcbe`（缺点原因） | `xmdd`（交期明细） |

### 7.4 字段的三种身份（设计器）

| 身份 | 控件代号 | `.4fd` fieldType | 生成行为 |
|---|---|---|---|
| 实栏位 | `表名.栏位名`（`xmdc_t.xmdc027`） | `TABLE_COLUMN` | 进 SELECT/INSERT/UPDATE/查重 |
| 借类型 | `l_xxx` 但 table/column 有值 | `COLUMN_LIKE` | 记录成员改名，DML 取该成员值 |
| 纯画面 | `l_xxx`，table/column 空 | `NON_DATABASE` | 无任何 SQL |

> 生成器以「控件代号是否带表名」作为身份的第一判据；改名等于把控件从表上摘下来。

---

## 8. 加点机制（受保护区域）

- `.tap` 里每个 point 是一段用户代码；`.tgl` 里对应一个 `{<point>}`
- 重产时：`.tgl` 重新展开，`.tap` 原样回填 → **客制不丢、生成段被覆盖**
- section 可「解锁」：解锁后该段不再随规格重产（单向、不可逆），代价是与标准补丁脱钩
- 与业界对照：
  - Eclipse JET 的 `@generated` + JMerge
  - .NET `partial class`
  - EMF `@generated NOT`
  - 学术上即 **Generation Gap 模式**

### 8.1 重产语义（务必记住）

| 改动 | 结果 |
|---|---|
| 改规格 + 重产 | `.tab/.tgl/.4gl` 重建，`.tap` 保留 |
| 手改 `.4gl` | 下次重产被覆盖 |
| 手改 `.tgl` | 破坏围栏/坐标，gate 直接拒绝 |
| 解锁 section | 该段永久脱离生成器 |

---

## 9. 版本、兼容与已知脆弱点

| 风险 | 机理 | 现象 |
|---|---|---|
| 模板与产生器版本不匹配 | `sadzi100_template_chk` MD5 校验（adzp150:467） | 产生中止 / 代码段错位 |
| `tables.xml` 过期 | 设计器启动时解析本地表字典（`04-designer/Common/SettingManager.cs`） | 新表的 PK/FK 带不出来 → 空列 |
| r.t 缺 `<fk>` | `GetFK` 找不到 `fk_table` 匹配 | 连线后 FK 仍为空 |
| FK 半填 | `32wc = upper_key − FK` 减法后仍缺列 | `AND  = xmdducseq` 空列 |
| 加点与模板版本耦合 | 模板换代后 point 名/位置可能变 | 客制代码回填失败，需人工迁移 |

---

## 10. 研究方法与证据链

1. **设计器**：反编译源码（`04-designer/`）——得到「规格如何被编辑与落库」
2. **服务器源码**：`01-server/`——得到「规格如何变成代码」
3. **启动脚本**：`02-launcher/`——得到「命令行的真实工具边界」
4. **现场包**：`.tzc`（`.tgl`/`.tap`/`.4gl` 三件套）——得到「输入输出长什么样」
5. **数据库字典与数据**：`r.t`、`tt dict`、只读 SQL——验证 FK/孤儿数据等假设
6. **交叉验证**：同一结论至少在两类证据中出现（如 32wc 同时见于源码与生成 SQL）

---

## 11. 术语表

| 中文 | 英文 | 在本项目中的对应 |
|---|---|---|
| 模型驱动开发 | MDD / MDA | 规格 → 代码 |
| 模型到模型 | M2M | `.tsd` → `.tab` |
| 模型到文本 | M2T | `.tab` + 模板 → `.tgl`/`.4gl` |
| 领域特定语言/建模 | DSL / DSM | 设计器 + `.tsd` |
| 受保护区域 | Protected Region | `.tap` / add-point |
| 生成间隙 | Generation Gap | 生成段 vs 加点 |
| 骨架生成 | Scaffolding | i01/t01 的 CRUD 骨架 |
| 主动式 / 被动式生成 | active / passive | T100 = 主动式（随时重产） |

---

## 12. 未解问题（后续研究）

1. **主样板 `$ERP/mdl/code_<type>.template` 未取得**（服务器未镜像）——i01/t01 的骨架细节只能从生成物反推，需从客户环境补采。
2. **`compose_path` 是二进制**：填孔算法（point 匹配、缩进、fence）只能通过行为观察，无法读源码。
3. **增量/差异生成的边界**：`adzp177` 的 section 合并、解锁 section 的保留策略，需更多样本。
4. **报表型态（r01/n00/k 类）** 的生成链与 ITM 差异（`adzp151/152/161`）。
5. **多语言（*l_t）与描述字段** 的自动生成规则尚未系统化。

---

## 附录 A：语料清单

| 路径 | 文件数 | 大小 | 说明 |
|---|---|---|---|
| `01-server/orchestration` | 13 | 734 KB | adzp080/adzp040/adzp060 + sadzp060_* |
| `01-server/blueprint` | 6 | 525 KB | sadzp030_tab*（蓝图/判型/关系） |
| `01-server/generator` | 13 | 691 KB | adzp150 + sadzp150_* |
| `01-server/generator-other` | 4 | 427 KB | adzp151/152/161（报表类）、adzp031（K 类） |
| `01-server/tgl` | 4 | 66 KB | adzp175/176/177/155 |
| `01-server/common` | 2 | 12 KB | 模板校验/差异 |
| `02-launcher` | 39 | 33 KB | r.* + tools/* |
| `03-templates/slice` | 118 | 162 KB | 切片模板 |
| `04-designer` | 14 | 213 KB | 表关联/控件/加点模型 |
| `diagrams` | 8 | 1.9 MB | 图解（HTML+PNG） |

逐文件 sha256 见 `MANIFEST.tsv`。

## 附录 B：关键行号速查

| 主题 | 位置 |
|---|---|
| 上传/产生编排 | `01-server/orchestration/adzp080.4gl:454/577/659` |
| 产生分派（SPECRC3/CODE） | `01-server/orchestration/adzp040.4gl:106/253/342/394` |
| r.c3 参数语义 | `01-server/orchestration/adzp040.4gl:906-913` |
| 执行子程序 | `01-server/orchestration/adzp040.4gl:1282` |
| tbl 节点落库 | `01-server/orchestration/sadzp060_1.4gl:897-926` |
| 判型 | `01-server/blueprint/sadzp030_tab_relation.4gl:830-956` |
| 蓝图主函数 | `01-server/blueprint/sadzp030_tab.4gl:184` |
| 三层二显减法 | `01-server/blueprint/sadzp030_tab.4gl:3466` |
| 样板白名单 | `01-server/generator/adzp150.4gl:401` |
| 名义→实际样板 | `01-server/generator/adzp150.4gl:412-442` |
| 主样板路径/输出 | `01-server/generator/adzp150.4gl:984/995-997` |
| 切片加载 | `01-server/generator/sadzp150_mdl.4gl:12` |
| 符号分发 | `01-server/generator/sadzp150_mdl.4gl:47-111` |

## 附录 C：命令速查

```bash
# 上传规格并产生程式（等价 F7）
r.r adzp080 'axmt500_wf(c).tzs' CXM SPECRC3 'D:\work\'

# 上传程式包（等价 F6）
r.r adzp080 'axmt500_wf(c).tzc' CXM CODE 'D:\work\'

# 单步：只产生蓝图
r.r adzp040 <prog> <mod> SPEC <ver> '' '' '' '' <env> <type> <lang>

# 单步：只展开模板 + 编译链接
r.c3 <mod> '<prog> '' A'

# 单步：编译 / 链接 / 画面
r.c <prog>
r.l <prog>
r.f <prog> tiptop
```
