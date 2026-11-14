# r.t（Table Designer / 数据表设计器）实现说明

## 1. 一句话结论

**`r.t` 不是服务器上的一个文件，而是登录 shell 里的一个 alias：`r.t → r.r adzi140`。**

真正干活的是 ADZ 模块的 **`adzi140`** 程序，用 Genero BDL（4GL）编写。它的核心逻辑是：

```
设计元数据（DZE* 表）
   → 与真实数据库比对（列 / 键值 / 索引 / 权限）
   → 生成 ALTER/CREATE 等 DDL
   → 交给 sqlplus / psql / gsql / ksql / sqlcmd / disql 执行
   → 记录快照、版本、异动码
   → 重产 .sch / .tbl / tables.xml / AlterTableList.xml / 多语言 4GL
```

本目录保存了 `adzi140` 的入口源码、全部链接模块、表单/菜单/工具栏/字符串资源、构建脚本，以及 `r.t` 启动链上的 shell 脚本。

## 2. 启动链：从 `r.t` 到 `adzi140.42r`

```text
用户 ksh（登录时 ~/.profile 选择 1=topprd / 2=toptst / 3=topstd / 4=toppth）
  │  . /u1/etc/chenv
  │  . /u1/etc/topenv
  │      └─ . /u1/etc/topalias
  │            alias r.t="r.r adzi140"
  ▼
r.t
  └─ /u1/toptst/utl/bin/r.r                 （ksh 包装脚本）
       PW=`pwd`; cd $TEMPDIR
       $FGLRUN $AZZ/42r/azzp000.42r adzi140
       cd $PW; exit $?
  ▼
$TOP/erp/azz/42r/azzp000.42r                （T100 框架通用启动器，仅二进制）
  ▼
$TOP/erp/adz/42r/adzi140.42r                （r.t 主程序，编译产物）
```

关键点：

| 环节 | 文件 / 变量 | 说明 |
|---|---|---|
| alias | `launcher/topalias:16` | `alias r.t="r.r adzi140"`；同一文件还有 `r.a/r.q/r.v/r.p/r.b/menu` 等设计器入口 |
| 启动器 | `launcher/r.r:11-19` | 切到 `$TEMPDIR`，用 `$FGLRUN` 执行 `$AZZ/42r/azzp000.42r`，把 `adzi140` 作为参数传入；结束后切回原目录 |
| `$AZZ` | `$ERP/azz` | 由 `/u1/etc/topsys` 按模块目录导出；`$AZZ/42r` 即 `$TOP/erp/azz/42r` |
| 框架入口 | `$TOP/erp/azz/42r/azzp000.42r` | T100 通用程序启动器/作业检查层，源码不在本环境（只有 42r）；负责装载目标程序 |
| 主程序 | `$TOP/erp/adz/42r/adzi140.42r` | 由本目录 `program/*.4gl` 编译链接而来 |
| 运行环境 | `/u1/etc/topenv` | 设置 `TOP/ERP/COM/UTL/FBIN/FGLRUN/FGLRESOURCEPATH/PATH` 等 |

> 注意：非交互式 SSH（如 `ssh host 'which r.t'`）不会加载 `.profile`，此时 `TOP` 为空、也没有 alias，所以看不到 `r.t`。这是 T100 登录脚本的行为，不是 `r.t` 缺失。

## 3. 目录内容

```text
r.t/
├─ README.md                 本文
├─ MANIFEST.tsv              134 个文件的本地路径 / 服务器原路径 / 大小 / SHA256
├─ launcher/                 启动链与 r.* 工具脚本
│  ├─ profile                登录脚本（选 1/2/3/4 决定 topprd/toptst/topstd/toppth）
│  ├─ topalias               全部 T100 alias（r.t 的定义在这里）
│  ├─ topenv                 T100 环境变量脚本（TOP/ERP/AZZ/FGLRUN/DB 等）
│  ├─ r.r                    r.t 的实际启动脚本
│  ├─ readme / r.help        T100 工具清单与帮助
│  └─ tools/                 r.42s、r.c、r.f、r.l、r.s、r.d、r.v 等 33 个工具
├─ program/                  adzi140 主程序 + 22 个模块 + 3 个 include
├─ forms/                    14 个 Genero 表单源文件（.4fd）
├─ deps/                     公共 include（sadzp000_*、sadzp200_*、top_global.inc）
├─ resources/                菜单/工具栏/样式/字符串/多语言资源
│  ├─ 4ad/                    Action Defaults
│  ├─ 4tm/                    TopMenu（DBA专用 / 工具 / 其他）
│  ├─ 4st/                    样式
│  ├─ 4tb/                    4 种语言的工具栏
│  ├─ per/                    由 .4fd 生成的表单布局文件
│  ├─ str/zh_CN/              中文简体字符串源
│  └─ xcf/                    GDC/GWC 客户端配置
└─ related/adzp090/          “批量执行异动”入口 adzp090（r.t 顶层菜单会启动它）
```

规模：`program/` 下 28 个文件，约 **51,839 行**、**720 个 FUNCTION**（含注释/空行）。

## 4. 运行环境与关键开关

来自 `launcher/topenv`、`launcher/topalias` 与代码中的 `FGL_GETENV`：

| 变量 | 典型值 | 用途 |
|---|---|---|
| `TOP` | `/u1/toptst` | 当前区（测试区）；生产为 `/u1/topprd`，标准区 `/u1/topstd`，Patch 区 `/u1/toppth` |
| `ERP` / `COM` / `UTL` / `FBIN` | `$TOP/erp` 等 | 模块、公共、工具、Genero 编译产物目录 |
| `FGLRUN` | `fglrun` | Genero 运行时 |
| `FGLRESOURCEPATH` | `$ERP:$COM` | 42f/42s/资源查找路径 |
| `DGENV` | `c` / `s` | `c`=客户客制环境，`s`=标准环境；影响可签出性、字段删除、行业别等分支 |
| `TOPCHKOUT` | `Y/N` | 是否允许签出 |
| `TOPALM` | `Y/N` | 是否启用 ALM（`DZLM_T` 签出流程） |
| `TOPIND` / `INDSD` | `sd` 或行业别 | 行业别；非 `sd` 时表名/字段会带行业后缀 |
| `ERPVER` | `3.0` | 版本分支（影响 `ds.sch` 的处理） |
| `CUST` | `DCS`/`DIGIWIN` 等 | 客户别 |
| `TEMPDIR` | `/ut/toptst/tmp` | `r.r` 的启动目录，也是 DDL/日志脚本的落盘目录 |

数据库：测试区是 Oracle 19c（`toptst`），但 `sadzi140_db.4gl` 同时包含 **Oracle / PostgreSQL / MSSQL / Kingbase / DM / Informix** 分支（`cs_db_driver_type_*`，见 `program/sadzi140_cnst.inc:405-411`）。

## 5. 程序结构

### 5.1 入口

`program/adzi140.4gl`（程序版本 `1.1.093`，源文件头标注 2025-01-13）：

```text
MAIN                                :377
  adzi140_initialize()              :384   环境/参数/连库/建存储过程/ALM 检查
  adzi140_initial_form()            :610   装载表单、TopMenu、ToolBar、样式、下拉框
  adzi140_start()                   :753   主 DIALOG / 全部 UI 动作
  adzi140_finalize(TRUE)            :2781
```

`initialize()` 的关键动作：

1. 读取 `DGENV`、`ERPVER`、`TOPIND`、`TOPCHKOUT`、`TOPALM`，连接 `ds`。
2. 从系统参数读取 `MasterDB`、`MasterUser`、`EnableCustomize`、`EnableIndustry`、`MaxTableNameLength` 等。
3. 校验设计数据完整性（`sadzi140_util_identify_db_data`）。
4. 在 Oracle/达梦下创建取字段默认值用的存储过程（`sadzi140_util_create_procedure`）。
5. 若启用 ALM，检查 `DZLM_T` 是否存在。
6. 计算表名/字段尾码（客户码 + 行业码 + `_t` 等）。

### 5.2 模块分工

| 文件 | 行数 | FUNCTION | 职责 |
|---|---:|---:|---|
| `adzi140.4gl` | 9,940 | 109 | 主程序：初始化、DIALOG、全部菜单/工具栏动作、调用各模块 |
| `sadzi140_db.4gl` | 10,119 | 126 | 数据库元数据查询、DDL 生成、SQL 执行、`.sch`/重建脚本 |
| `sadzi140_exim.4gl` | 9,528 | 191 | `.tdi` 导入/导出：设计表数据的 JSON 序列化/反序列化 |
| `sadzi140_util.4gl` | 4,492 | 82 | 异动编排、表格重建、快照、授权、导入导出胶水、文件操作 |
| `sadzi140_rdif.4gl` | 3,492 | 15 | 按版本查看差异（Revision Diff） |
| `sadzi140_shot.4gl` | 3,201 | 25 | 设计数据/真实 DB 的快照与差异比对 |
| `sadzi140_diff.4gl` | 2,178 | 15 | “表格差异比对”界面（设计 vs DB） |
| `sadzi140_crud.4gl` | 1,593 | 24 | 对 `DZEA/DZEB/DZEC/DZED/DZEN/DZEP/DZEF/...` 的增删改 |
| `sadzi140_xml.4gl` | 1,353 | 19 | `tables.xml`、`.tbl`、`AlterTableList.xml` 生成 |
| `sadzi140_rev.4gl` | 1,294 | 27 | 版本列表、SQL 预览、回滚 |
| `sadzi140_clone.4gl` | 672 | 10 | 表格复制/克隆 |
| `sadzi140_vcs.4gl` | 346 | 8 | 版本号（major/minor/rev/build）序列 |
| `sadzi140_log.4gl` / `logs` | 345 / 223 | 13 / 6 | 异动日志、日志查看 |
| `sadzi140_cols.4gl` | 324 | 12 | 字段挑选式导入的左/右选择界面 |
| `sadzi140_lbl.4gl` | 301 | 14 | 字符串/多语言标签生成（`.str` → `.42s`） |
| `sadzi140_msg.4gl` / `inp` / `dlg` | 236 / 223 / 61 | 7 / 6 / 4 | 消息、输入、文件对话框辅助 |
| `sadzi140_prc.4gl` | 201 | 3 | Oracle 存储过程 DDL |
| `sadzi140_lang` / `bkrv` / `hcpy` | 80 / 22 / 19 | 2 / 1 / 1 | 语言、备份/硬拷贝占位实现 |
| `sadzi140_cnst.inc` / `type.inc` / `mcr.inc` | 335 / 1,045 / 7 | — | 常量、记录类型、宏 |

> 编译脚本还会带上 `sadzi888_04`、`sadzp220`、`adzp220`；`sadzi140_alm` 只在 42m 产物中存在，未找到 `.4gl` 源，故未收录。

### 5.3 UI 结构

表单 `forms/adzi140.4fd` 的页面：

- `page_MasterTable`：模块下拉、搜索、表格清单（表名/说明/类型/是否多语言/系统表/Trigger/签出者等）。
- `page_ColumnDefineList`：共用字段定义清单，可拖拽到栏位表。
- `page_PageTableContents`：栏位内容（字段名、说明、主键、必填、属性、类型、长度、控件、默认值、开窗、检查、取值等）。
- `page_PageKeys`：键值（PK/FK/RK）。
- `page_PageIndex`：索引。
- `page_PagePrivilege`：表格权限。

顶层菜单 `resources/4tm/adzi140.4tm`：

- **DBA专用**：产出 `.sch`、`tables.xml`、`.tbl`、APS 权限；批次产出；批次执行异动；批次表格重建；版本比对及回复；删除非正规字段；表格重建；修正设计数据字段序号。
- **工具**：表格设计数据导出/导入/快速复制导入、简易数据操作工具、查看使用表格程序清单、分镜档表格导入、字段挑选式导入、删除标准字段、批次签出。
- **其他**：关于。

工具栏 `resources/4tb/toolbar_adzi140_*.4tb`：新建表、附属表、复制、删除、修改/确认/取消、执行异动、差异比对、表格重建、导出 Excel、栏位设计资料维护（`adzi150`）、离开。

## 6. 设计数据模型（`ds` schema）

`r.t` 把“表长什么样”存成一组元数据表，再据此生成真实 DDL。常用表（说明取自 `tt dict r.t` 实时字典）：

| 表 | 中文说明 | 在 r.t 中的作用 |
|---|---|---|
| `dzea_t` | 数据表主档 | 表名、说明、模块、表类型、多语言/系统表/Trigger 标记、签出/异动状态 |
| `dzeb_t` | 数据表字段档 | 字段名、说明、PK/必填、属性、类型、长度、控件、默认值、开窗、检查、取值 |
| `dzebl_t` | 数据表字段多语言档 | 字段说明的多语言 |
| `dzeal_t` | 数据表多语言档 | 表说明的多语言 |
| `dzec_t` | 数据表索引档 | 索引名、类型、字段、唯一性 |
| `dzed_t` | 数据表键值档 | PK/FK/RK 键值及外来键关系 |
| `dzen_t` | 表格权限赋予档 | 表级权限（SELECT/INSERT/UPDATE/DELETE/REFERENCES/ALTER/INDEX） |
| `dzep_t` | 字段规格设计数据档 | 控件、格式、显示宽度、百分比等前端规格 |
| `dzef_t` | 字段参考设置档 | 检查存在、取值/开窗参照 |
| `dzer_t` | 多语言关联设置档 | 多语言表与主表的关联 |
| `dzet_t` | 助记码设置档 | 助记码 |
| `dzeq_t` | 树状结构设置档 | 树状表设置 |
| `dzeu_t` | 表格.同义字设置档 | Table/Synonym/View 类型 |
| `dzev_t` | 实际表格与设计数据对照设置档 | 每次异动后的“设计 vs 实际”快照 |
| `dzew_t` | 索引/键快照档 | 索引/键快照 |
| `dzey_t` | Table异动更新档 | 待异动/已异动表清单 |
| `dzhj_t` | 客户端标准字段删除纪录档 | 记录被删除的标准字段 |
| `dzlm_t` | 签出信息档 | ALM 签出/签入/版本/需求单号 |
| `dzlu_t` | 多语言表格设置主档 | 代码中也用于 ALM 签出者/行业别信息 |
| `gztd_t` | 字段属性定义档 | 数据类型、长度、默认控件等属性字典 |
| `dzeo_t` | 表格版本控制码档 | VCS 的 major/minor/rev/build 序号（代码 `sadzi140_vcs.4gl`） |

> 真实表可能分布在多个 schema（`ds`、`dsdemo`、`dsdata`、`dsstd`、行业 schema 等）；`r.t` 会先用 `sadzi140_db_get_table_in_db_type` 找出该表以 Table/Synonym 形式存在的所有库，再逐个执行异动。

## 7. 关键流程

### 7.1 启动与编辑

1. 选模块 → 读 `dzea_t` 填表清单；选中表后读 `dzeb_t/dzec_t/dzed_t/dzen_t/dzep_t/dzef_t` 填各页签。
2. 栏位页支持拖拽 `dzel_t` 的共用字段；修改时在内存数组中打 `+ / * / - / >` 标记。
3. `MODIFY` 动作按当前焦点分派：
   - 表头 → `adzi140_modify_master_table`
   - 栏位 → `adzi140_alter_column_data`
   - 键值 → `adzi140_alter_key_data`
   - 索引 → `adzi140_alter_index_data`
   - 权限 → `adzi140_alter_privilege_data`
4. 写库统一走 `sadzi140_crud_insert_update_*`，落到 `DZEA/DZEB/DZEC/DZED/DZEN/DZEP/DZEF/...`。
5. 写库前会做大量规则校验：PK 数量/长度、标准字段保护、行业别、客户客制字段、ALM 签出状态、表名长度/尾码、索引是否已存在等。

### 7.2 “执行异动”（核心）

入口 `adzi140.4gl:2194`（`ON ACTION tbi_exec_alter`），实际调用 `sadzi140_util_make_alter_table`（`program/sadzi140_util.4gl:665`）。

```text
adzi140_check_db_data_diff / sadzi140_shot_check_diff   先判断设计数据与 DB 是否有差异
  ▼
sadzi140_util_make_alter_table
  for 每个存在该表的 DB/schema:
    ├─ 检查表锁
    ├─ PostgreSQL：备份并删除依赖 View
    ├─ 生成/执行 “Table ↔ Synonym/View” 类型调整脚本
    ├─ 设计侧差异（TableEditor）
    │    ├─ 列：sadzi140_db_get_alter_column_sql
    │    ├─ 约束：sadzi140_db_get_alter_constraint_sql
    │    └─ 索引：sadzi140_db_get_alter_index_sql
    ├─ 数据库侧差异（Database）
    │    ├─ 删列：sadzi140_db_get_drop_column_sql
    │    ├─ 删约束：sadzi140_db_get_drop_constraint_sql
    │    └─ 删索引：sadzi140_db_get_drop_index_sql
    ├─ 外键/注释/授权/LOB 表空间等
    ├─ sadzi140_rev_gen_sql_script_file  写 $TEMPDIR/<owner>_<table>_<guid>.<type>.sql
    └─ sadzi140_db_sqlplus               调 sqlplus / psql / gsql / ksql / sqlcmd / disql 执行
  ▼
sadzi140_db_update_alter_code              更新异动码
sadzi140_util_create_curr_snapshot         写 DZEV_T / DZEW_T 快照
  ▼
对每种语言:
  sadzi140_xml_gen_table_datas             生成 .tbl + tables.xml
  sadzi140_db_gen_table_schema             生成 $ERP/<模块>/sch/<lang>/<表>.sch
  adzi140_set_altered_table_info           写 DZEY_T
sadzi140_db_gen_db_schema                  r.s ds <表> → 合并到 $TOP/sch/ds.sch
sadzi140_xml_gen_web_table_datas          生成 AlterTableList.xml
sadzi140_util_gen_multi_lang              多语言表 → 生成多语言 4GL
sadzi140_util_gen_extend_inc              生成扩展 include
sadzi140_util_recompile_invalid_db_objects 重编译失效对象
sadzi140_util_grant_revoke_privileges     授权 / APS 权限
```

要点：

- **DDL 不是拼一条 SQL 就执行**，而是按“设计侧要加的”和“数据库侧要删的”两轮生成，避免误删客户自建字段；`DGENV=c` 时对客制字段更保守。
- 每种 DDL 都有数据库分支（Oracle 的 `ALTER TABLE ... MODIFY`、PostgreSQL 的 `ALTER COLUMN ... TYPE ... USING`、MSSQL 的 `nvarchar(max)` 等）。
- 执行走外部 DB 客户端（`sadzi140_db_sqlplus`，`program/sadzi140_db.4gl:1345`），脚本与日志落在 `$TEMPDIR`，便于失败后排查。
- 成功后 `sadzi140_db_gen_db_schema` 调 `r.s ds <table>` 更新 `ds.sch`；这是 T100 编译器/建库使用的 schema 汇总文件。

### 7.3 表格重建

当 `ALTER TABLE` 无法完成（例如类型/长度改不动、需要保留数据重建）时使用。

入口 `adzi140.4gl:2340`，调用 `sadzi140_util_table_rebuild`（`program/sadzi140_util.4gl:1631`）：

1. 检查 `dzeb` 的 column id 是否需要重建；检查参数 `Rebuild/CheckDataExists`，有数据则拒绝。
2. `sadzi140_db_gen_rebuild_script`（`program/sadzi140_db.4gl:1791`）生成重建脚本：
   - Oracle/达梦：PL/SQL 块，建 `<表>_rebuild`、搬数据、删旧表、改名、重建 Trigger/索引/约束，并用 `SPOOL` 收集日志。
   - PostgreSQL/MSSQL：不能直接使用该 PL/SQL 过程，脚本按各自的数据库分支处理。
3. 执行脚本 → 成功则写快照、更新异动码、重产 `.sch`、多语言、扩展 include、重编译失效对象、授权。

### 7.4 导入 / 导出 / 复制

- 导出：`tm_export` → `sadzi140_util_export_table_schema`（`sadzi140_util.4gl:2132`）→ `sadzi140_exim_run` 把设计表数据序列化成 JSON，打成 `.tdi`（tar）包，再下载到客户端。
- 导入：`tm_import` → `sadzi140_util_import_table_schema`（`:2185`）→ 上传 `.tdi` → 解析表信息 → 校验 ERPVER/构造版本 → 逐表导入 `DZEA/DZEB/...` → 执行异动 → 快照/产物。
- 快速复制导入：`tm_clone_import` 解包后改表名/内容，再按新表名重新打包导入。
- 字段挑选式导入：`sadzi140_cols` 提供左右选择界面，只导入选中字段。

### 7.5 版本、快照、签出

- **版本号**：`sadzi140_vcs` 用 `DZEO_T` 维护 major/minor/rev/build 序列，`sadzi140_vcs_get_new_version_code`（`:104`）生成 `x.y.z`。
- **快照**：`sadzi140_shot` 把“设计数据”和“真实 DB”分别快照到 `DZEV_T`（表/列）与 `DZEW_T`（索引/键），用于后续 diff 和回滚。
- **版本比对/回复**：`sadzi140_rev` 列出历史版本、预览 SQL、回滚；`sadzi140_rdif` 看某一版本的差异；`sadzi140_diff` 看当前设计与 DB 的差异。
- **签出**：`btn_check_out/check_in/release` 走 ALM（`DZLM_T`，`sadzp200_alm_*`）；是否允许由 `TOPALM`、`TOPCHKOUT`、系统参数共同决定。标准环境下未签出不能改；客户环境（`DGENV=c`）可放宽。

## 8. 构建与发布

`program/adzi140` 是 ADZ 模块的构建脚本（134 行），顺序为：

```text
r.l adz                         # 先链接模块公共库
r.c sadzi140_<module>           # 逐模块编译 4gl → 42m
r.c adzi140
r.c sadzi888_04 / sadzp220 / adzp220
mv adzi140.4fd sadzi140_*.4fd ../4fd   # 13 个表单源移到 4fd
mv adzi140.4st $ERP/cfg/4st             # 样式
cp adzi140.4ad ../4ad/zh_TW,zh_CN       # Action Defaults（简繁两份）
cp toolbar_adzi140_*.4tb $ERP/cfg/4tb   # 工具栏（4 种语言）
cp adzi140_<lang>.4tm $ERP/adz/4tm/<lang>  # TopMenu（4 种语言）
cd ../4fd; r.f sadzi140_<form> ...; r.f adzi140   # 编译表单 → 42f
r.l adz                         # 重新链接公共库
r.l adzi140                     # 链接主程序 → 42r
r.l adzp220/adzp230/adzp240/adzp250/adzp999      # 关联程序
```

工具链（本目录 `launcher/tools/`）：

| 脚本 | 底层 | 作用 |
|---|---|---|
| `r.c` | `$FBIN/42m/gen42m` | 编译 `.4gl` → `.42m` |
| `r.f` | `$FBIN/42m/gen42f` | 编译 `.4fd` → `.42f` |
| `r.l` | `$FBIN/42m/gen42r.42r` | 链接 `.42m` → `.42x/.42r` |
| `r.42s` | `gen42s_all.42m` | 重产字符串 `.42s` |
| `r.s` | `gensch.42r` 等 | 重产数据库 schema（`ds.sch`） |
| `r.d/r.dg` | `gendbg` | 调试器 |

## 9. 建议的阅读顺序

1. `launcher/topalias`、`launcher/r.r`：先弄清 `r.t` 只是别名。
2. `program/adzi140.4gl:377` 的 `MAIN` 与 `:384/:610/:753` 三个初始化函数。
3. `program/adzi140.4gl:933` 开始的 `DIALOG`，按 `ON ACTION` 找功能；重点看 `tbi_exec_alter`（`:2194`）、`tbi_table_rebuild`（`:2340`）、`tm_export/tm_import`（`:1933/:1937`）、`tm_rev_list`（`:2462`）。
4. `program/sadzi140_util.4gl:665` 的 `make_alter_table`，配合 `sadzi140_db.4gl` 的 DDL 函数读。
5. `program/sadzi140_shot.4gl:520` 的 `check_diff` 与 `sadzi140_vcs.4gl:104`，理解快照/版本。
6. 最后看 `sadzi140_exim.4gl`（导入导出）与 `forms/adzi140.4fd`（界面字段与页签）。

## 10. 边界与未包含内容

- **框架层没有源码**：`azzp000.42r`、`cl_*`、`sadzp000_*`/`sadzp200_*` 的运行时实现属于 T100 框架；本目录只收了 `adzi140` 直接 `&include` 的公共 include。
- **编译产物未收**：`adzi140.42r/.42m/.42f/.42s` 等二进制没有下载；需要时到 `$TOP/erp/adz/42r`、`42m`、`42f`、`42s` 取。
- **单表源不等于可独立编译**：`adzi140.4gl` 依赖 `SCHEMA ds` 的数据库字典、Genero 运行时和框架库；本目录用于阅读/审计，不是可独立构建工程。
- **目录做了归类**：为便于阅读，本目录按 `launcher/program/forms/resources/deps/related` 分类，没有保留 `4gl/4fd` 的原始相对层级；源码里的 `&include "../4gl/..."` 不能直接在本地解析。需要还原原始树时，按 `MANIFEST.tsv` 的 `remote_path` 放回对应目录即可。
- **环境差异**：本目录来自 `toptst`（测试区）。生产区 `topprd`、标准区 `topstd` 的表结构/参数/客制分支可能不同；核对时以对应区的 `$TOP` 为准。
- **只读采集**：服务器侧只执行了 `ls/find/grep/head/cat/strings` 与 `scp` 下载，没有改文件、没有执行 r.t 的异动功能。

## 11. 文件溯源

`MANIFEST.tsv` 逐文件记录：

```
local_path <TAB> remote_path <TAB> bytes <TAB> sha256
```

例如：

```
program/adzi140.4gl    /u1/toptst/erp/adz/4gl/adzi140.4gl    473199    <sha256>
launcher/r.r           /u1/toptst/utl/bin/r.r                 334       <sha256>
```

需要还原某个文件时，按 `remote_path` 回服务器取即可。
