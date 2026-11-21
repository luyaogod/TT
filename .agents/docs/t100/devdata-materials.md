# TT 开发资料（devdata）· P0 材料事实册

- **状态：P0 完成**（口径冻结与测量；本文件即 `tt-devdata-tasks.md` P0 一节 7 个勾选项的交付物）
- **测量环境**：`福联测试tst@1108`（企业 1108，Oracle 19c，账号 `dsdemo`，**只读**）
- **测量时间**：2026-11-21（本机 14:31 起）
- **只读手段**：`tt db sql --batch`（53 张 `DZ*` 的 `COUNT(*)`、字段数、`DZEAL_T` 中文名、`DZAA_T` 版次统计）、`tt file ls` / `tt file head`（`dzx/tbl`、`dzx/refd`、`$TOP/erp/mdl` 的清单与首个文件内容）
- **语料依据**：`.agents/docs/t100/codegen/`（`01-server` 生成器 4GL、`ARCHITECTURE.md`、`MANIFEST.tsv`、`README.md`）
- **冲突裁决**：本文件与 `.agents/plan/tt-devdata-plan.md` 冲突时，**以 plan 的 ADR 为准**；且必须**先在 plan 里改**，再回头改本文件（plan §12.3）。
- **范围**：只写事实与口径，不含实现代码。所有数字来自上文实测或 `codegen` 语料中的 grep 证据；查不到的写「待确认」，不编。

---

## 0. 摘要（十行速览）

1. 规格库 `DZ*` 首版 **53 张**，实测合计 **18,371,187 行**，字段数 4–38 列不等。
2. 其中两张超 200 万行：`DZBA_T` 7,810,349、`DZAZ_T` 4,622,785 —— 都是生成器的**运行时产物/计数**，不是入参，本期 `skipped`。
3. 剩余 51 张合计 **5,938,053 行（≈5.9M）**；再套 `DZAA_T` 有效版过滤（702,438 → 272,355）后预计落库 **5,507,970 行（≈5.5M）**。
4. 53 张表里 **52 张**在生成器语料里找到明确读写证据；**1 张（`DZBH_T`）待确认**（只有一段被注释掉的 SQL）。
5. 版本维度结论：1837 万 ≫ 50 万 → **只镜当前有效版**；但只有 `DZAA_T` 做版次过滤，其余 52 张整表镜（通用 `*STUS='Y'` 会静默清空表）。
6. `mdl` 实测：顶层 154 个文件（`code_*` 45 + `*.snippet` 90 + `4rp_*` 3 + 其它 16）+ `slice/` 126 个；总量 3,204,623 B（≈3.0 MiB）。
7. 「模板 171」= `code_*`(45) + `slice`(126)；但其中真正 `code_*.template` 只有 **38** 个，另 7 个是带日期的备份。
8. `dzx/tbl/`、`dzx/refd/` 各看一眼后**本期默认排除**：两者都是**表级** XML（设计器本地定义/缓存），不是程序级、不是生成器输入。
9. 目录布局、manifest schema、5 个 HTTP 端点、6 条 CLI 命令在本文件第 6、7 节冻结，与 plan §5.1/§5.2/§5.5/§5.7 一致。
10. 所有测量均为 `SELECT`；未写 T100 服务器任何文件。

---

## 1. 角色表（角色 → 源路径 → 落点 → 必需性）

### 1.1 M 类：全量镜像，env 级

落点根：`<data>/devdata/<env>/`，`<data>` = `Locations.DataDir()`（`config.json` 旁）。源路径一律写 `$TOP` 相对形式。

| 角色 | 源路径（`$TOP` 相对） | 落点 | 必需性 | 上限 / 备注 |
|---|---|---|---|---|
| 规格库 `DZ*` 53 张表 | schema `DS`（Oracle），表名清单见附录 A | `<data>/devdata/<env>/dict/<env>.db`（SQLite，逐表原样列） | **必需**（单表超限 → 该表 `skipped`，不判整体失败） | **单表 200 万行**；实测命中 `DZBA_T`、`DZAZ_T` |
| 数据字典来源表（`r.t` / `r.v` / `r.q` / `scc` 的底表） | schema `DS`（具体表清单**待确认**；语料里以 `gztz_t` 当「表/栏位字典」判存在性，见 `sadzp030_tab_ma.4gl:764` `SELECT COUNT(1) … FROM gztz_t WHERE gztz001=…`） | 同上 db | **必需** | 行数**待确认**（未做 `COUNT(*)`） |
| 主样板 | `erp/mdl/code_*.template` | `<data>/devdata/<env>/mdl/code_*.template` | **必需** | 单文件 64 MB；实测 `code_*` 共 45 个，其中 `.template` 后缀 38 个 |
| 切片 | `erp/mdl/slice/*`（`code_a00..a77`、`d01..d10`、`qs01..qs27`、`ss01..ss04`） | `mdl/slice/`（保持相对路径） | **必需** | 实测 126 个 |
| snippet | `erp/mdl/*.snippet` | `mdl/` | **必需** | 实测 90 个 |
| 4rp 样板 | `erp/mdl/4rp_*.template` | `mdl/` | **必需** | 实测 3 个 |
| XML 设置（plan 明列） | `erp/mdl/ComboxItemSetting.xml` | `mdl/` | **必需** | plan §2.5 明列 |
| 其余顶层文件 | `erp/mdl/*`：`*.4fd` 表单样板、`UITemplateSetting.xml`、`xbarr.xlsx`、`test-rc3.sh` 等共 16 个 | **不拉** | **可选**（plan §2.5 未列） | 不计入「模板 171」；是否纳入**待确认** |
| `xrp/` 子目录 | `erp/mdl/xrp/*` | **不拉** | **待确认** | 文件数与用途未测（`待确认`） |
| —— | 整个 `mdl/` 总量 | —— | —— | **总量 200 MB**；实测仅 3.0 MiB，远低于上限 |

补充：

- 单文件上限 **64 MB**（plan §10 / §5.4），超限拒绝并写 `warnings[]`。
- M 类不拉 `erp/<mod>/*.42x`、`42m`、`42r`、`azzp188/190/191/193/195`、`GZOL_T/GZDD_T`、`mta/*`（plan 附录 B）。
- 数据字典与规格库同库不同表、同一个 SQLite 文件（plan §2.5 落点列）。

### 1.2 D 类：按需，按程序成套

落点根：`<data>/devdata/<env>/prog/<prog>/`。`<mod>` 为程序所属模块（如 `axmt500_wf` → `CXM`）。

| 角色 | 源路径（`$TOP` 相对） | 落点 | 必需性 | 上限 / 备注 |
|---|---|---|---|---|
| 蓝图 `.tab` | `erp/<mod>/dzx/tab/<prog>.tab` | `prog/<prog>/dzx/<prog>.tab` | **必需**（缺 → 整套失败） | 单文件 64 MB |
| 表单规格 `.tsd` | `erp/<mod>/dzx/tsd/<prog>.tsd` | `prog/<prog>/dzx/<prog>.tsd` | **必需** | 单文件 64 MB |
| 加点 `.tap` | `erp/<mod>/dzx/tap/<prog>.tap` | `prog/<prog>/dzx/<prog>.tap` | **必需**（plan §4.4 要求 `dzx/{tab,tsd,tap,tgl}` 齐全；真机上是否每个程序都有 `.tap` 未逐个验证，见 §9） | 单文件 64 MB |
| 排序骨架 `.tgl` | `erp/<mod>/dzx/tgl/<prog>.tgl` | `prog/<prog>/dzx/<prog>.tgl` | **必需** | 单文件 64 MB |
| 展开中间物 `.tgx` | `erp/<mod>/dzx/tgl/<prog>.tgx` | `prog/<prog>/dzx/<prog>.tgx` | **存在才拉** | 缺失进 `missing[]` |
| 整理中间物 `.tgx2` | `erp/<mod>/dzx/tgl/<prog>.tgx2` | `prog/<prog>/dzx/<prog>.tgx2` | **存在才拉** | plan §5.1 显式列出 |
| 产物 `.4gl` | `erp/<mod>/4gl/<prog>.4gl` | `prog/<prog>/src/<prog>.4gl` | **必需** | 单文件 64 MB |
| 编译引子 mask | `com/inc/erp/<mod>/<prog>_mask.4gl` | `prog/<prog>/inc/<prog>_mask.4gl` | **存在才拉** | 缺失进 `missing[]` |
| 编译引子 rep | `com/inc/erp/<mod>/<prog>_rep.4gl` | `prog/<prog>/inc/<prog>_rep.4gl` | **存在才拉** | 缺失进 `missing[]` |
| 表单 `.42f` | `erp/<mod>/42f/<prog>.42f` | `prog/<prog>/form/<prog>.42f` | **可选**（`--form` 才拉） | Genero 二进制；单文件 64 MB |
| 设计稿 `.tzs` | 用户本地同名目录（zip：`.4fd`/`.tsd`/`.bdx`） | `prog/<prog>/design/<prog>.tzs` | **可选**（`--with-design` 才归档） | 只复制，不解析 |

补充（与 plan 一致，勿自行扩张）：

- 一律**按程序成套**，不提供散装下载（ADR-04）；缺 `.tab`/`.tsd`/`.tap`/`.tgl`/`.4gl` → 整套失败；`missing[]` 只承载可选项（ADR-04、plan §5.2）。
- 目录名 `<prog>` 强校验 `^[a-z][a-z0-9_]{2,31}$`（plan §5.3）；环境名来自配置清单白名单，中文环境名按原样建目录。

> 注：plan §2.5 对 `.tap` 未标「存在才拉」，故本表按「必需」记。若真机验证发现标准程序确实无 `.tap`，应**先改 plan**（ADR-04 或 §2.5）再改本表，不在此处私自放宽。

---

## 2. 53 张 `DZ*` 表逐张用途

说明：

- 「中文说明」照抄 `DZEAL_T`（zh_CN，繁体原样）。
- 「字段数」来自 `DZEB_T` 逐表 `COUNT(*)`；「行数」为 2026-11-21 实测 `COUNT(*)`。
- 「生成器角色」取值：**读** / **写** / **读+写** / **未在语料出现**。判定口径是 `codegen/01-server/` 里是否有有效（非注释）的 `SELECT/FROM` 与 `INSERT/UPDATE/DELETE`。上传/同步路径（`sadzp060_*`）的写入也算「写」，用途列会注明是生成期读还是上传期写。
- 证据格式：`文件:行`（可带短引文）。文件均在 `.agents/docs/t100/codegen/01-server/` 下。

| 表 | 中文说明 | 字段数 | 行数 | 生成器角色 | 用途与证据 |
|---|---|---:|---:|---|---|
| `DZAA_T` | 规格与内容版本对应表 | 18 | 702,438 | 读+写 | 规格主档：`dzaa001` 程序、`dzaa002` 版次、`dzaa003` 识别码、`dzaa005` 类型、`dzaa009` 客制标示、`dzaastus` 有效码。生成器所有取数都以它为驱动与版次来源：`sadzp030_tab.4gl:525` `SELECT MAX(dzaa002) INTO lc_sd_ver FROM dzaa_t WHERE dzaa001=g_prog_id AND dzaa009="s"` → `:530` 写成 `.tab` 的 `sdver`；上传入库 `sadzp060_1.4gl:1994` `insert into dzaa_t(...)` |
| `DZAB_T` | 规格整体设计表 | 15 | 5,921 | 读+写 | 整体规格 CLOB（`dzab099`），按 `dzaa003` 分成 `ALL`/`MI_ALL`/`DB_ALL`/`DI_ALL` 四段。生成 `.tsd` 时读取：`sadzp030_tsd.4gl:957` `SELECT … dzab099 FROM dzab_t RIGHT JOIN dzaa_t …`，`:967-977` 按 `ALL`/`MI_ALL`/`DB_ALL`/`DI_ALL` 建节点；上传写入 `sadzp060_1.4gl:1630` |
| `DZAC_T` | 字段规格设计表 | 30 | 212,109 | 读+写 | 每字段的检核/画面属性：`req`、`i_zoom`/`c_zoom`、`chk_ref`、`default`、max/min 及其运算符、`can_edit`/`can_query`。生成蓝图时整批读入 `g_dzac_m`：`sadzp030_tab_ma.4gl:1391` `SELECT dzac005,dzac002,dzac009,dzac010,dzac011,dzac014,dzac015,dzac020,dzac016,dzac021,dzac017,dzac018 …`；`sadzp030_tab.4gl:2508-2539` 用 max/min 组 `<check id="range" st=… bt=…>`，配切片 `03-templates/slice/code_a15.template` 的 `cl_ap_chk_range(...)`；上传写入 `sadzp060_1.4gl:1837` |
| `DZAD_T` | Action规格设计表 | 16 | 22,976 | 读+写 | Action 清单与「触发时机」`dzad006`、是否「产生标准程式框架」`dzad007`。生成按钮/事件时读取：`sadzp030_tab.4gl:4467` `SELECT DISTINCT dzad002,dzad007 FROM dzaa_t,dzad_t … AND (dzad006='all' OR dzad006 LIKE 'all,%' …)`；上传写入 `sadzp060_1.4gl:2751` |
| `DZAF_T` | 规格与程序版本历程对应表 | 10 | 47,299 | 读+写 | 规格/程序版本历程。未给版次时用它反推规格版次：`adzp031.4gl:85` `SELECT MAX(dzaf002) FROM dzaf_t WHERE dzaf001=? AND dzaf007=? AND dzaf010=?`，`:94`/`:102` 取 `dzaf003`（规格版次）或 `dzaf004`（报表元件用程序版次）；上传/签出写入 `sadzp060_4.4gl:719` |
| `DZAG_T` | 规格Table设置表 | 23 | 7,139 | 读+写 | 表关联：`dzag004` parent、`dzag008` pk、`dzag009/010` fk_detail/fk_master、`dzag013/014/015` 三层二显。判型与 JOIN 依据：`sadzp030_tab_relation.4gl:263` `SELECT dzag008 … FROM dzaa_t,dzag_t`；三层二显 `sadzp030_tab.4gl:3485` `SELECT dzag013,dzag014,dzag015 FROM dzaa_t,dzag_t`，`:3535-3539` 拼 `upper_key - FK` 的 `AND`；上传由 `sadzp060_1.4gl:918` 写（`.tsd` 的 `<tbl>` 属性 → `dzag004/008/009/010/013/014/015`，见 `sadzp060_1.4gl:897-926`） |
| `DZAI_T` | 字段参考设计表 | 19 | 38,855 | 读+写 | reference（参照）设定：参照表 `dzai008`、被参照栏位 `dzai005`、取值栏位 `dzai011`、对照给值 `dzai007`。生成 `<reference ref_sql=…>`：`sadzp030_tab.4gl:2603` `SELECT dzai_t.* INTO lc_dzai.* FROM dzai_t,dzaa_t …`，`:2712` 拼 `SELECT <refcol> FROM <reftable> WHERE <wc>`；上传写入 `sadzp060_1.4gl:2118` |
| `DZAJ_T` | 字段数据多语言设计表 | 20 | 867 | 读+写 | 多语言字段（`dzaa005='7'`）的依附控件/多语言 table/FK/语系栏位/回传栏位。生成 `<multilang target=… lang_table=… fk=…>`：`sadzp030_tab_ma.4gl:1484` `SELECT dzaj005,dzaj007…dzaj011 FROM dzaj_t WHERE dzaj001=g_prog_id …`，`sadzp030_tab.4gl:2449` `createChild("multilang")`；上传写入 `sadzp060_1.4gl:2263` |
| `DZAK_T` | 字段助记码设计表 | 19 | 89,462 | 读+写 | 助记码（mnemonic）设定：搜索栏位 `dzak007`、助记码表 `dzak008`、回传栏位 `dzak009`、语系条件 `dzak010`、对应控件 `dzak011`。生成 `<mnemonic table=… field=… wc=…>`：`sadzp030_tab.4gl:2406` `sadzp030_tab_ma_2(ls_columnid,"dzak007")`、`:2421` `createChild("mnemonic")`、`:2423-2433` 取 `dzak008/007/009/011`；上传写入 `sadzp060_1.4gl:2416` |
| `DZAL_T` | 程序串查设计表 | 17 | 1,441 | 读+写 | 字段的多笔串查（一个字段可挂多个查询 action），按 `dzal008` 排序。生成串查定义时读取：`sadzp030_tab_ma.4gl:1234` `SELECT dzal002,dzal005,dzal006,dzal007 FROM dzal_t WHERE dzal001=g_prog_id AND dzal005=? … AND dzalstus='Y' ORDER BY dzal008`，`:1441-1460` 过滤掉「没有程式参考设定」的串查；上传写入 `sadzp060_1.4gl:2554` |
| `DZAM_T` | 规格画面元件排除设置 | 14 | 7,337 | 读+写 | 「排除」设定：被排除的字段/元件不进生成结果。生成时先查排除再决定是否跳过：`sadzp030_tab_ma.4gl:318` `SELECT COUNT(1) INTO li_deny FROM dzaa_t,dzam_t WHERE … dzam003=lc_dzfj005 …`，`:334-339` 命中即 `CONTINUE FOREACH`；判型侧另读 `dzam003`（`sadzp030_tab_relation.4gl:872` `SELECT dzam003 FROM dzaa_t,dzam_t`）；上传写入 `sadzp060_1.4gl:2941` |
| `DZAQ_T` | 行业程序同步纪录表 | 22 | 64 | 读+写 | 行业程序同步日志：签出时按程序删旧插新，事后回填行业建构版次。`sadzp060_4.4gl:355` `DELETE FROM dzaq_t WHERE dzaq001=l_prog`、`:357` `INSERT INTO dzaq_t(...)`、`:420` `UPDATE dzaq_t SET dzaq004=?`（行业建构版次）、`:417` `SELECT COUNT(1) … FROM dzaq_t` |
| `DZAS_T` | 行业非主程序与标准程序关系表 | 8 | 0 | 读 | 行业程序 ↔ 标准程序的对应，用来判断「某程序是否被引用」。`sadzp060_ind.4gl:834` `SELECT COUNT(1) INTO li_cnt FROM dzas_t …`（`sadzp060_ind_is_cite_by_dzas_t`），`:870` `select * from dzas_t where dzas001=?`；`sadzp060_4.4gl:227` `EXISTS (SELECT 1 FROM dzas_t WHERE dzas003=dzye002 …)` |
| `DZAT_T` | 设计数据脱机档备份 | 20 | 0 | 写 | 签出/取消签出时把设计数据备份成 `.tzs`/`.tzt` 的索引：`sadzp060_3.4gl:197` `UPDATE dzat_t SET dzat009=l_file,dzat010=l_tzt`（旧版次）、`:210` `INSERT INTO dzat_t (dzat001,dzat002,…,dzat009,dzat010)`（新版次）。语料中无读取方 |
| `DZAU_T` | 行业对应之标准规格引用清单 | 6 | 1,254 | 读+写 | 行业规格引用标准规格的清单（加速引用比对）。同步时按「行业程序 + 行业版次」删旧插新：`sadzp060_4.4gl:1256` `DELETE FROM dzau_t WHERE dzau001=g_ind_prog AND dzau002=g_ind_sver`、`:1259` `INSERT INTO dzau_t(dzau001..dzau006)`；读取上一版做差异：`:1044` `(SELECT DISTINCT(au1.dzau006) FROM dzau_t au1 … MAX(au2.dzau002) …)` |
| `DZAX_T` | 编程基本设置表 | 15 | 9,748 | 读+写 | 程式类别 `dzax002`、free style 指标 `dzax003`、固定参数个数 `dzax007`，是**判型入口**：`sadzp030_tab_relation.4gl:1172` `SELECT dzax002,dzax003,dzax007 INTO … FROM dzax_t WHERE dzax001=g_prog_id AND dzax006=g_dgenv AND dzaxstus="Y"`（`sadzp030_tab.4gl:510-511` 写成 `.tab` 的 `fix_arg`）；生成器侧 `adzp150.4gl:185` `SELECT dzax003 INTO lc_dzax003 FROM dzax_t`（`.tgl` 的 free-style 分支）；上传写入 `sadzp060_1.4gl:661` |
| `DZAY_T` | dzax_t 的暂存表（取消签出使用） | 5 | 152 | 读+写 | 签出时把 `DZAX_T` 复制到 `DZAY_T`；取消签出时再写回。`sadzp060_2.4gl:2207-2221`（签出：`DELETE FROM dzay_t` + `INSERT INTO dzay_t(...)`）、`:2229` `SELECT * INTO lr_dzay.* FROM dzay_t WHERE dzay001=p_prog` → `:2232` `UPDATE dzax_t …`（取消签出）、`:2304` `UPDATE dzay_t SET dzay005=p_regen` |
| `DZAZ_T` | 生成器生成计数表 | 4 | 4,622,785 | 读+写 | 每个程序每次产生递增一次 build 次数（运行时产物）：`adzp150.4gl:1198` `SELECT MAX(dzaz002) INTO li_dzaz002 FROM dzaz_t WHERE dzaz001=lc_dzaz001`、`:1207` `INSERT INTO dzaz_t(dzaz001,dzaz002,dzaz003,dzaz004)`（`dzaz003`=OS 用户、`dzaz004`=时间）。报表类生成器同构：`adzp151.4gl:570/579`、`adzp152.4gl:923/932`、`adzp161.4gl:630/639` |
| `DZBA_T` | 程序与内容版本对应表 | 20 | 7,810,349 | 读+写 | 程序内容版次对照（含使用版次 `dzba004`、`dzba005='c'` 标示）。编排侧判断「是否首版」：`adzp040.4gl:1484` `select count(distinct ba.dzba002) ver_count,min(ba.dzba002) min_ver from dzba_t ba where ba.dzba001=…`；上传时建/改版次：`sadzp060_2.4gl:1736` `insert into dzba_t(...)`、`:1779` `update dzba_t set dzba005='c',dzba004=1`；行业同步 `sadzp060_4.4gl:1442/1456/1493` |
| `DZBB_T` | 编程点设计表 | 16 | 611,403 | 读+写 | add-point（编程点）定义：point 名 `dzbb002`、版次 `dzbb003`、内容 `dzbb098`。判「free style 程序是否已转成加点程序」：`sadzp030_tab_relation.4gl:1079` `SELECT COUNT(1) INTO li_cnt FROM dzbb_t WHERE dzbb001=g_prog_id AND dzbb002="free_style.variable"`，为 0 则调 `sadzp180_free` 自动补加点；上传写入 `sadzp060_1.4gl:3615`、行业同步 `sadzp060_4.4gl:1534` |
| `DZBC_T` | section标准数据档 | 19 | 971,136 | 读+写 | section 的顺序/型态/只读：`dzbc003` 名称、`dzbc004` 版次、`dzbc005` 型态、`dzbc010` 顺序、`dzbc011` readonly。adzp177 组装 `.tgl` 时先看库里有没有 section 顺序：`adzp177.4gl:104` `SELECT COUNT(1) … FROM dzbc_t … dzbc010 IS NOT NULL`，`:119-125` `SELECT dzbc003,dzbc011,dzbc005 FROM dzbc_t … ORDER BY dzbc010`；adzp176 维护：`adzp176.4gl:256` `INSERT INTO dzbc_t`、`:304` `UPDATE dzbc_t SET dzbc004=…`；上传 `sadzp060_2.4gl:1837` |
| `DZBD_T` | section设计表 | 14 | 242,683 | 读+写 | section 内容（CLOB `dzbd098`）。adzp177 取 section 内容并与骨架合并：`adzp177.4gl:208` `FROM dzbd_t`；`gs_wmode='2'` 时清掉暂存：`:71` `DELETE FROM dzbd_t WHERE dzbd001=gs_prog_name AND dzbd004='p'`；adzp176 写入 `:382/:442/:536` `INSERT INTO dzbd_t`、`:596` `UPDATE dzbd_t SET dzbd098=…`；上传 `sadzp060_1.4gl:3971` |
| `DZBK_T` | 差异比对Patch编号条件设置 | 10 | 0 | 写 | 语料中唯一有效语句是 Patch 同步收尾的整表清空：`sadzp060_4.4gl:273` `DELETE FROM dzbk_t`（在 `sadzp060_4_batch_ins_dzaq_t` 末尾，注释 `#170726-00035`）。插入方/读取方**未在语料出现** → 用途细节**待确认** |
| `DZBH_T` | Autocomplete功能SQL维护表 | 12 | 19,250 | 未在语料出现 | **待确认**。代码里只有一段被整体注释掉的 SQL：`sadzp150_fieldchk.4gl:415` `#LET ls_sql = "SELECT COUNT(1) FROM dzbh_t WHERE dzbhstus='Y' AND dzbh001=… AND dzbh002=?…"`（`PREPARE`/`EXECUTE` 同样被注释，见 `:419`、`:604`、`:783`、`:857`、`:1078`）。语料研究结论另记「`DZBH_T` autocomplete（`a77` 片段）」（`ARCHITECTURE.md:435`），但**没有生效代码可验证**，故不据此下结论 |
| `DZBU_T` | 行业对应之标准add point引用清单 | 5 | 2,407 | 读+写 | `DZAU_T` 的 add-point 版：行业程序引用了哪些标准加点。`sadzp060_4.4gl:1566` `DELETE FROM dzbu_t WHERE dzbu001=g_ind_prog AND dzbu002=g_ind_cver`、`:1569` `INSERT INTO dzbu_t(dzbu001..dzbu005)`；读上一版 `:1350` `(SELECT DISTINCT(bu1.dzbu005) FROM dzbu_t bu1 … MAX(bu2.dzbu002) …)` |
| `DZCA_T` | 开窗数据表 | 16 | 5,261 | 读 | Q 类开窗程序的规格（`dzca002` c/s、`dzca006` free-style、创建人/日期）。判型时识别开窗程序：`sadzp030_tab_relation.4gl:1131` `SELECT dzca006 INTO lc_dzax003 FROM dzca_t WHERE dzca001=g_prog_id AND dzca002="c"`（找不到再取 `"s"`），`:1139` 置 `gzde003="Q"`；取创建人 `:738` `SELECT dzcacrtid,dzcacrtdt FROM dzca_t …` |
| `DZCB_T` | 开窗设计参数设计表 | 12 | 4,444 | 读 | 开窗参数说明：生成开窗切片时按 `dzcb003` 顺序取 `dzcb004` 当参数说明（`mdl_arg_desc1..N`）：`sadzp150_ctrlp.4gl:76-80` `SELECT dzcb004 FROM dzcb_t WHERE dzcb001=<form> AND dzcb003 IS NOT NULL ORDER BY dzcb003`，`:85-94` 逐参数编号 |
| `DZCC_T` | 开窗显现设置表格 | 18 | 32,327 | 读 | 开窗回传/显现字段：`dzcc003` 字段、`dzcc005='Y'` 表示要显现。`sadzp150_ctrlp.4gl:100-104` `SELECT dzcc003,dzeb003 FROM dzcc_t LEFT OUTER JOIN dzeb_t ON dzcc003=dzeb002 WHERE dzcc001=<form> AND dzcc005='Y'`，`:110-121` 组合成 `field/var/desc` 三元组 |
| `DZCD_T` | 校验带值设计表 | 15 | 1,459 | 读 | 校验带值主档：`chkid` → 型态 `dzcd005`。生成 `<check id=…>` 时决定是「纯检查 / 纯带值 / 两者」：`sadzp030_tab.4gl:2548` `SELECT dzcd005 INTO lc_dzcd005 FROM dzcd_t WHERE dzcd001=lc_dzcd001 AND dzcd002="c"`（找不到取 `"s"`），`:2570-2581` `1→isExist`、`2→reference`、`3→chkandReturn`；注释明确「`dzcd_t` 表不须检查状态码」（`:2551`） |
| `DZCE_T` | 校验带值参数设计表 | 15 | 2,402 | 读 | 校验带值的参数档：数 `chkid` 有几个参数 → 校验带值切片 `a17` 的 `mdl_mdls`。`sadzp150_object.4gl:883-886` `SELECT COUNT(1) INTO li_tmp FROM dzce_t WHERE dzce001=ls_tmp` → `addAttribute("mdl_mdls", …)`，`:896` `adzp150_make_slice("a17")`；另两处同构 `:975`、`:1045` |
| `DZCP_T` | 公式设置主表 | 13 | 0 | 读 | 「简易带值」的**目标控件**侧（`pi_type` 8/9）：判断某字段有没有被设定，决定 INPUT 要不要生成该字段。`sadzp030_tab_ma.4gl:677` `SELECT COUNT(1) … FROM dzcp_t WHERE dzcp001=g_prog_id AND dzcp002=lc_dzfj005 AND dzcpstus='Y'`（先查 `gztz_t` 确认表存在，`:672`）；读公式内容 `sadzp150_fieldchk.4gl:1351` `SELECT dzcp006 FROM dzcp_t` |
| `DZCT_T` | 带值设置主表 | 17 | 0 | 读 | 「公式」的**目标控件**侧（与 `DZCP_T` 成对，代码用 `dzct_t` + `dzcw_t`）。`sadzp030_tab_ma.4gl:721` `SELECT COUNT(1) … FROM dzct_t WHERE dzct001=g_prog_id AND dzct002=lc_dzfj005 AND dzctstus='Y'`，`:742` 用控件侧 `FROM dzct_t,dzcw_t … dzcw003=lc_dzfj005`；读公式内容 `sadzp150_fieldchk.4gl:1356` `SELECT dzct007 FROM dzct_t` |
| `DZCY_T` | 参数值设置参数表 | 15 | 0 | 读 | 带值/校验的参数值与使用控件：`dzcy005` 参数顺序、`dzcy007` 使用控件。生成 `g_qryparam.argN = <控件>`：`sadzp150_ctrlp.4gl:163/169` `SELECT COUNT(*)/SELECT dzcy005,dzcy007 FROM dzcy_t WHERE dzcy003=… AND dzcy001=app_id AND dzcy004=<form> AND dzcy002 LIKE '%<column>%'`，`:225` `LET g_qryparam.arg<N> = …`；校验侧 `sadzp150_object.4gl:1603/1609`（`dzcy003='3'`） |
| `DZDY_T` | 样板基本数据表 | 13 | 155 | 读 | 样板编号 → 存放路径 `dzdy004`，模板校验/取模板时用：`sadzi100_extract.4gl:28` `SELECT dzdy004 INTO lc_dzdy004 FROM dzdy_t WHERE dzdy001=lc_dzdy001` |
| `DZDZ_T` | 样板及版本更新基本数据表 | 10 | 1,648 | 读+写 | 样板版次与内容 MD5：`sadzi100_extract.4gl:70` 取 `dzdz003`（CLOB）、`:140` `SELECT dzdz004 INTO ls_md5_o … WHERE dzdz002=(SELECT MAX(dzdz002) …)` 做 MD5 比对，`:213` `UPDATE dzdz_t SET dzdz010=pi_ver`（把版本号写回）；`sadzp030_tab_file.4gl:50` `SELECT MAX(dzdz002) FROM dzdz_t` 取样板最新版 |
| `DZEB_T` | 数据表字段档 | 28 | 150,665 | 读 | 表字段字典：`dzeb001` 表、`dzeb002` 字段、`dzeb003` 中文说明、`dzeb004` 显示名、`dzeb006` 栏目属性、`dzeb007` 数据类型。生成 DEFINE 类型：`sadzp030_tab_ma.4gl:529` `SELECT dzeb006 INTO lc_dzeb006 FROM dzeb_t WHERE dzeb001=dzac005 AND dzeb002=dzac002`（`:543` `sadzp030_tab_ma_datatype2`）；多语言显示名 `sadzp150_lang.4gl:280`；模糊查询只挑字符型 `sadzp030_tab_ma.4gl:638-641` `dzeb007 IN ('varchar','varchar2','clob','blob')` |
| `DZED_T` | 数据表键值档 | 20 | 5,540 | 读 | 表键值字典：`dzed003='P'` 为 PK。求主键拼键数组：`sadzp030_tab_relation.4gl:296` `SELECT dzed004 FROM dzed_t WHERE dzed001=p_dzag002 AND dzed003="P"`（先试设计器给的 PK，失败再回落 `r.t`，见 `:284-306`）；另一路 `adzp150.4gl:3738` `SELECT dzed004 INTO lc_chk_site_pk FROM dzed_t`（site pk 校验） |
| `DZEM_T` | SA分镜规格导入接口档 | 11 | 0 | 读 | SA 分镜规格：按 `dzem003` 分类型（`1` 字段规格→`sa_field`、`2` Action 规格→`sa_act`、`3` 整体规格→`sa_all`）。生成 `.tsd` 时读取：`sadzp030_tsd.4gl:2152` `SELECT dzem_t.* FROM dzem_t WHERE dzem001=gs_prog`，`:2161-2171` 按类型建节点 |
| `DZEP_T` | 字段规格设计数据档 | 38 | 158,023 | 读+写 | 字段的「系统分类码」`dzep011`（SCC），落到 `.tab` 的 `def_scc`：`sadzp030_tab.4gl:2399-2402` `lc_dzep011 = sadzp030_tab_ma_2(ls_columnid,"dzep011")` → `setAttribute("def_scc", …)`；取数 `sadzp030_tab_ma.4gl:1422` `SELECT dzep011 FROM dzep_t WHERE dzep001=… AND dzep002=…`；上传更新 `sadzp060_1.4gl:1908` `UPDATE dzep_t` |
| `DZFF_T` | 树状结构属性设置档 | 17 | 449 | 读+写 | 树状画面的属性（`s_browse` 的 tag 属性 → `type/pid/slid/…`）。生成 `i04/i05/i08` 树结构时读取：`sadzp030_tab_relation.4gl:1582` `SELECT dzff005,dzff006,dzff007 FROM dzaa_t,dzff_t WHERE dzaa003='s_browse' AND dzaa005='5' …`，`:1603-1606` 按属性名分派；上传写入 `sadzp060_1.4gl:1351`、行业同步 `sadzp060_4.4gl:2151` |
| `DZFI_T` | 画面结构档 | 25 | 410,088 | 读 | 画面控件层级：`dzfi004` 群组、`dzfi005` 顺序、`dzfi006` 元件组、`dzfi007` 节点型态。按画面树递归取字段清单：`sadzp030_tab_ma.4gl:220` `SELECT dzfi004,dzfi005,dzfi006,dzfi007 FROM dzfi_t WHERE dzfi001=g_prog_id AND dzfi002=g_sd_ver AND dzfi009=g_dgenv AND dzfi004=pc_mastername`，`:938` 递归下阶；判 qbe 型态 `sadzp030_tab_relation.4gl:1195/1213`（`dzfi006='qbe'/'vb_master'`） |
| `DZFJ_T` | 画面元件组合档 | 19 | 899,991 | 读 | 控件 → 字段：`dzfj004` 顺序、`dzfj005` 控件/字段名、`dzfj006` 控件型态。逐控件产字段清单：`sadzp030_tab_ma.4gl:259` `SELECT dzfj004,dzfj005,dzfj006 FROM dzfj_t WHERE dzfj001=g_prog_id AND dzfj002=g_sd_ver AND dzfj003=sr[i].dzfi006 AND dzfj017=g_dgenv ORDER BY dzfj004`，`:268-269` 只留 Edit/ButtonEdit/ComboBox…；特殊控件判断 `sadzp150_fieldchk.4gl:1243` `COUNT(*) … dzfj006='DateTimeEdit'` |
| `DZFK_T` | 画面元件属性数据档 | 9 | 1,120,856 | 读+写 | 控件属性 CLOB `dzfk009`。解析 4fd 时把属性解成 `format/case/required/hidden/noEntry` 回写 `dzci_t`：`adzp040.4gl:1161-1167` `select dzfk003, dzfk009 from dzfk_t … where dzfk001=<prog> and dzfk002=<ver> and dzfk007=<dgenv>`，`:1175` `sadzp252_1_parse_property(dzfk009)`；行业同步回写 `sadzp060_4.4gl:2852` `UPDATE dzfk_t` |
| `DZFQ_T` | 画面结构设计主档 | 24 | 5,192 | 读 | 样板型态 `dzfq001` 与创建人。决定 form style：`adzp040.4gl:678-681` `select fq.dzfq001 form_style from dzfq_t fq where fq.dzfq004=<prog> and fq.dzfq003='1'`；取创建人 `sadzp030_tab_relation.4gl:721` `SELECT dzfqcrtid,dzfqcrtdt FROM dzfq_t WHERE dzfq003="1" AND dzfq004=g_prog_id`；生成 `.tsd` 时 `sadzp030_tsd.4gl:349` `SELECT dzfq_t.* FROM dzfq_t` |
| `DZFS_T` | 程序Table与Screen Record对应档 | 20 | 8,435 | 读+写 | 程序里的 Table 与画面 Screen Record 对应；`s_browse` 的记录型态存 `dzfs010`。判「有没有 browse / browse 型态」：`sadzp030_tab_relation.4gl:1253` `SELECT dzfs010 … FROM dzaa_t,dzfs_t WHERE … dzfs003="s_browse" … dzfsstus="Y"`（`SQLCODE=100` 时置 `"N"`），`:1369` `SELECT dzfs004 …`；上传写入 `sadzp060_1.4gl:1197` |
| `DZFU_T` | 画面样版ToolBar功能键设置档 | 10 | 98 | 读 | ToolBar/ON ACTION 清单：`dzfu002` 为 action。生成 `.tsd` 的 ToolBar：`sadzp030_tsd.4gl:706` `FROM dzfu_t`；上传/改版时取标准 action 清单：`sadzp060_8.4gl:81` `SELECT DISTINCT(dzfu002) FROM dzfu_t WHERE dzfustus='Y'` |
| `DZGC_T` | 报表元件设计-数据模型字段明细档 | 9 | 78,228 | 读 | 报表（GR/XR）数据模型字段 → 实际表栏位：`dzgc006='Y'` 表示自订字段，`dzgc007` 为表名。生成报表元件时读取：`adzp151.4gl:2025` `SELECT dzgc006 INTO ls_dzgc006 FROM dzgc_t WHERE dzgc001=g_prog_id AND dzgc002=g_prog_ver AND (dzgc005='1' OR dzgc005='3') AND dzgc004=p_field AND dzgc009=g_code_ide`，`:2088-2094` 自订字段转查 `dzgd_t`；`adzp161.4gl:3062/3081` 同构 |
| `DZGD_T` | 报表元件设计-数据模型自订义字段明细档 | 8 | 26,338 | 读 | 自订字段的对应栏位 `dzgd004` 与公式 `dzgd006`。`adzp151.4gl:2088-2090` `SELECT dzgd004,dzgd006 FROM dzgd_t WHERE dzgd001=g_prog_id AND dzgd002=g_prog_ver AND dzgd003=p_field AND dzgd008=g_code_ide`，`:2094` 组 `field.table.column`；`adzp161.4gl:3126` 同构 |
| `DZGG_T` | 报表元件设计-样板单头档(GR) | 17 | 724 | 读 | GR 报表样板单头：`dzgg001` 程序、`dzgg002` 规格版次、`dzgg006` 是否签核。`adzp151.4gl:948` `SELECT dzgg006 INTO l_dzgg006 FROM dzgg_t …`（`:925-927` 定义），`adzp161.4gl:1324` 同构 |
| `DZKC_T` | 报表组件设计-数据模型字段明细档(XR) | 18 | 23,769 | 读 | XR 报表数据模型字段（`dzkc004` 字段、`dzkc006` 型态、`dzkc007` 表名、`dzkc010/011` 比对）。`adzp161.4gl:1875` `FROM dzkc_t WHERE dzkc001=g_prog_id AND dzkc002=g_prog_ver`，`:1862` 取明细、`:1976-1989` 取字段型态 |
| `DZKD_T` | 报表组件设计-自订字段明细档(XR) | 16 | 5,196 | 读 | XR 自订字段：`dzkd010`/`dzkd011` 对应栏位与公式。`adzp161.4gl:1927` `FROM dzkd_t WHERE dzkd001=g_prog_id AND dzkd002=g_prog_ver`，`:1850-1851` 定义 `dzkd010/dzkd011`，`:2024` 取明细 |
| `DZYE_T` | 设计数据运行结果记录表 | 24 | 2,824 | 读 | Patch 设计数据的运行结果（程序 `dzye002`、建构类型 `dzye003`、来源版次 `dzye014/020/021`）。同步前先读它决定哪些程序要处理：`sadzp060_4.4gl:214-231` `SELECT dzye002,dzye003,dzye005,dzye014,dzye019,dzye020,dzye021 FROM dzye_t WHERE dzye001=? AND dzye003='M' … UNION … dzye003 IN ('S','F','Q')` |
| `DZZC_T` | 简易画面数据设置 | 13 | 0 | 读 | 简易画面的「实际存放模块」`dzzc004`（简易画面可能不在本模块）：`adzp150.4gl:5107` `SELECT COUNT(1) INTO ls_cnt FROM dzzc_t`，`:5118` `SELECT dzzc004 INTO ls_dzzc004 FROM dzzc_t`（注释说明「所以简易画面需以 dzzc_t 中纪录的实际模组(dzzc004) 为主」`adzp150.4gl:16`）；`adzp152.4gl:4024/4034` 同构 |

### 2.1 本节统计

| 项 | 数量 | 表 |
|---|---:|---|
| 有明确读/写证据 | **52** | 除 `DZBH_T` 外的全部 |
| 其中「读+写」 | 26 | `DZAA DZAB DZAC DZAD DZAF DZAG DZAI DZAJ DZAK DZAL DZAM DZAQ DZAU DZAX DZAY DZAZ DZBA DZBB DZBC DZBD DZBU DZDZ DZEP DZFF DZFK DZFS` |
| 其中「只写」 | 2 | `DZAT`（备份）、`DZBK`（清空） |
| 其中「只读」 | 24 | `DZAS DZCA DZCB DZCC DZCD DZCE DZCP DZCT DZCY DZDY DZEB DZED DZEM DZFI DZFJ DZFQ DZFU DZGC DZGD DZGG DZKC DZKD DZYE DZZC` |
| **待确认** | **1** | **`DZBH_T`**（只有注释掉的 SQL） |

另外 6 张表虽然证据明确，但只出现在**上传/同步**链路而非 `.tab/.tgl` 生成链路上，读用途要打折看：`DZAS_T`、`DZAT_T`、`DZAQ_T`、`DZAU_T`、`DZBU_T`、`DZBK_T`。

---

## 3. 行数探测（实测，2026-11-21）

- 环境：`福联测试tst@1108`（企业 1108）
- 库：Oracle 19c，账号 `dsdemo`，**只读**
- 手段：`tt db sql --batch`，逐表执行 `SELECT COUNT(*) FROM DS.<表>;`（53 次）
- 单位：行

| 表 | 行数 |
|---|---:|
| `DZAA_T` | 702,438 |
| `DZAB_T` | 5,921 |
| `DZAC_T` | 212,109 |
| `DZAD_T` | 22,976 |
| `DZAF_T` | 47,299 |
| `DZAG_T` | 7,139 |
| `DZAI_T` | 38,855 |
| `DZAJ_T` | 867 |
| `DZAK_T` | 89,462 |
| `DZAL_T` | 1,441 |
| `DZAM_T` | 7,337 |
| `DZAQ_T` | 64 |
| `DZAS_T` | 0 |
| `DZAT_T` | 0 |
| `DZAU_T` | 1,254 |
| `DZAX_T` | 9,748 |
| `DZAY_T` | 152 |
| `DZAZ_T` | 4,622,785 |
| `DZBA_T` | 7,810,349 |
| `DZBB_T` | 611,403 |
| `DZBC_T` | 971,136 |
| `DZBD_T` | 242,683 |
| `DZBH_T` | 19,250 |
| `DZBK_T` | 0 |
| `DZBU_T` | 2,407 |
| `DZCA_T` | 5,261 |
| `DZCB_T` | 4,444 |
| `DZCC_T` | 32,327 |
| `DZCD_T` | 1,459 |
| `DZCE_T` | 2,402 |
| `DZCP_T` | 0 |
| `DZCT_T` | 0 |
| `DZCY_T` | 0 |
| `DZDY_T` | 155 |
| `DZDZ_T` | 1,648 |
| `DZEB_T` | 150,665 |
| `DZED_T` | 5,540 |
| `DZEM_T` | 0 |
| `DZEP_T` | 158,023 |
| `DZFF_T` | 449 |
| `DZFI_T` | 410,088 |
| `DZFJ_T` | 899,991 |
| `DZFK_T` | 1,120,856 |
| `DZFQ_T` | 5,192 |
| `DZFS_T` | 8,435 |
| `DZFU_T` | 98 |
| `DZGC_T` | 78,228 |
| `DZGD_T` | 26,338 |
| `DZGG_T` | 724 |
| `DZKC_T` | 23,769 |
| `DZKD_T` | 5,196 |
| `DZYE_T` | 2,824 |
| `DZZC_T` | 0 |
| **合计（53 张）** | **18,371,187** |

### 3.1 补充实测

| 指标 | 值 | 说明 |
|---|---:|---|
| `DZAA_T` 不同程序数 | 5,445 | 按程序代号 `dzaa001` 去重 |
| `DZAA_T` 中 `DZAASTUS='Y'` 行数 | 665,449 | 有效码行 |
| `DZAA_T` 按「每程序 `MAX(DZAA002)` 且 `DZAASTUS='Y'`」筛后 | **272,355** | 本期实际要镜的行数（见 §4） |
| 0 行的表 | **8 张** | `DZAS_T`、`DZAT_T`、`DZBK_T`、`DZCP_T`、`DZCT_T`、`DZCY_T`、`DZEM_T`、`DZZC_T` |
| 超 200 万行的表 | 2 张 | `DZBA_T` 7,810,349、`DZAZ_T` 4,622,785 |

---

## 4. 版本维度结论（ADR-03 补充段）

### 4.1 结论

- 53 张合计 **18,371,187 行**，远大于 plan ADR-03 的 50 万行分界线 → **不做全版次镜像，只镜「当前有效版」**。
- 但「有效版」不是对 53 张表统一套 `*STUS='Y'`：
  - `DZAA_T` 有版次维度（`dzaa002`）与有效码（`dzaastus`），按程序取最新有效版；
  - 其余 52 张表**整表镜**，**不做**通用 `*STUS='Y'` 过滤（其中 `DZBA_T`/`DZAZ_T` 另因超 200 万行转 `skipped`，见 §4.2 第 3 条）。
- 反证（实测）：`DZEB_T`、`DZED_T` 的 `*STUS='Y'` 行数为 **0**。这两张是字段/键值字典，状态码语义与 `DZAA_T` 不同（或长期为 `N`/空）。若套通用过滤，会在无告警的情况下把整张表镜成空表 —— 这是必须避免的静默数据缺失。

### 4.2 本期可实施口径（实现按这一条做）

1. **`DZAA_T` 只镜每程序最新有效版**：

   ```sql
   SELECT * FROM DS.DZAA_T
    WHERE DZAASTUS = 'Y'
      AND DZAA002 = (SELECT MAX(B.DZAA002)
                       FROM DS.DZAA_T B
                      WHERE B.DZAA001 = DZAA_T.DZAA001
                        AND B.DZAASTUS = 'Y')
   ```

   实测命中 **272,355 行**（基准：全表 702,438；`DZAASTUS='Y'` 665,449；不同程序 5,445）。

2. **其余 52 张整表镜**：`SELECT * FROM DS.<表>`，不加任何 `*STUS` 条件；实际落库时 `DZBA_T`/`DZAZ_T` 按第 3 条跳过。
3. **单表超 200 万行 → 该表 `skipped` + 写 `warnings[]`**，不中断其他表：
   - `DZBA_T` 7,810,349 → skipped
   - `DZAZ_T` 4,622,785 → skipped
   - 这两张都是生成器的**运行时产物/计数**（`DZAZ_T` 是 build 次数计数，`adzp150.4gl:1198-1207`；`DZBA_T` 是程序内容版次对照，`adzp040.4gl:1484`、`sadzp060_1.4gl:3726`），**不是生成入参**，跳过不影响 model 的输入完整性。
4. 探测量与跳过原因写进 M 侧 `manifest.json` 的 `warnings[]`（plan ADR-03 要求可追溯）。

### 4.3 后果与算式

| 项 | 行数 | 算式 |
|---|---:|---|
| 53 张全量 | 18,371,187 | 实测 |
| 减 `DZBA_T` | −7,810,349 | 超 200 万行 → skipped |
| 减 `DZAZ_T` | −4,622,785 | 超 200 万行 → skipped |
| **其余 51 张合计** | **5,938,053** | 18,371,187 − 7,810,349 − 4,622,785 = **≈5.9M** |
| 其中 `DZAA_T` 由整表 702,438 缩为 272,355 | −430,083 | 702,438 − 272,355 |
| **本期 dict 侧预计实际落库** | **5,507,970** | 5,938,053 − 430,083 = **≈5.5M** |

结论口径：

- 对外表述用「**≈5.9M 行左右**」（= 51 张整表上限，不含 `DZAA_T` 版次过滤）；实现完成后的真实行数应落在 **≈5.51M**，两者同一数量级，差异只来自 `DZAA_T` 的版次过滤。
- 另有「数据字典来源表」要进同一个 `.db`，其行数**待确认**（§1.1），不在上述 1837 万/593 万之内。

### 4.4 `tt gen status` 必须能显示 skipped 表清单

- M 侧 `manifest.json` 与 `status` 载荷都要带 skipped 信息（表名 + 实测行数 + 原因），否则用户会以为「53 张都齐了」而实际少两张。
- 建议固定文案：`skipped: DZBA_T (7,810,349 行 > 200 万)、DZAZ_T (4,622,785 行 > 200 万)`，同时原样进 `warnings[]`。
- 表数字段口径：`status` 里「表数」应同时给出 `total=53`、`mirrored=51`、`skipped=2`，避免把 51 当 53。

---

## 5. `dzx/tbl/`、`dzx/refd/` 格式结论（本期默认排除）

### 5.1 实测（环境 `福联测试tst`，模块 `axm`）

| 目录 | 语言子目录 | 文件 | 内容格式 |
|---|---|---|---|
| `erp/axm/dzx/tbl/` | 仅 `zh_CN`、`zh_TW` 两个 | `zh_CN` 下为逐表的 `<table>.tbl`，**150 个** | XML：`<table name="xmaas_t" version="1"><column name="xmaasent" text="" attribute="N802" type="number(5)" pk="Y" req="Y"/>…` |
| `erp/axm/dzx/refd/` | 8 个（`en_US`/`ja_JP`/`ko_KR`/`th_TH`/`vi_VN`/`zh_CN`/`zh_TW` 等，其中多个为空） | 逐表的 `<table>.refd`，**107 个** | XML：`<xmaa_t><reference name="xmaa001" column="xmaaent,296,xmaa001" return-column="oocql004" ref-table="oocql_t" …/>` |

### 5.2 结论

1. 两者都是**表级**（不是程序级）的 XML：文件名就是 `<table>`，内容描述该表的列/参照，与「某个程序」无绑定关系。
2. 内容性质是**设计器侧的本地定义/缓存**（列属性、reference 定义），与 `mta/*`（设计器本机缓存，plan 附录 B）同类；不是生成器输入 —— 生成器的输入是 `DZ*` 规格表与 `.tab`/`.tgl`，对 `01-server/` 做 `(?i)refd|\.tbl` 的 grep **零命中**，说明 `sadzp030_*`、`adzp150`、`adzp175-177` 没有任何一处读这两类文件。
3. 因此**本期排除**：与 plan §3.2「不碰 `dzx/{tbl,refd}`」、附录 B「格式未定性（P0 看一眼再定）」一致，P0 的「看一眼」已做，结论是继续排除。
4. 若将来要纳入：**按表组织**（如 `tbl/<table>.tbl`），不要按程序塞进 `prog/<prog>/`；也不要与 `mdl/` 混放，因为它们的更新节奏与来源（设计器）和模板（服务器 `$TOP/erp/mdl`）不同。
5. 待确认：8 个语言目录中哪些为空、`refd` 的 107 个是否覆盖全部业务表、以及 `tbl` 与 `r.t`/`DZEB_T` 的权威顺序。这些都不影响本期排除结论。

---

## 6. manifest schema 与目录布局冻结

> 本节**原样固化** plan §5.1 与 §5.2；与 plan 冲突时以 plan 为准。下面只做要点复述与字段补齐，不改语义。

### 6.1 目录布局（冻结，同 plan §5.1）

```
<data>/devdata/                       # <data> = Locations.DataDir()（config.json 旁）
└─ <env>/                             # 环境名，与环境清单同名（含中文，按原样建目录）
   ├─ manifest.json                   # M 侧清单：每表的行数/版本/sha、每个模板的 sha、同步时间
   ├─ dict/<env>.db                   # M：DZ* 53 张 + 数据字典，SQLite（独立于 <data>/dict/）
   ├─ mdl/                            # M：保持 $TOP/erp/mdl 的相对路径
   │  ├─ code_i02.template …
   │  └─ slice/code_a05.template …
   └─ prog/<prog>/                    # D：一个程序一目录，成套
      ├─ manifest.json
      ├─ design/<prog>.tzs            # 可选（--with-design）
      ├─ dzx/{<prog>.tab,.tsd,.tap,.tgl,.tgx2}
      ├─ src/<prog>.4gl
      ├─ inc/                         # *_mask.4gl / *_rep.4gl（存在才有）
      └─ form/<prog>.42f              # 可选（--form）
```

落地要点（与 ADR 一致）：

- 先写 `<prog>.tmp-<rand>/`，全部成功后**原子改名**；失败/取消删临时目录（ADR-02 之外的 plan §5.4 纪律）。
- **不**把 `devdata` 加进 `pkg/cachedir.Subdirs`（ADR-02）；清理只走 `tt gen clean`。
- 与读侧不共用任何目录（ADR-01）：`<data>/dict/<env>.db` 是字典库，`<data>/devdata/<env>/dict/<env>.db` 是 devdata 自己的库。

### 6.2 manifest 契约（`tt.devdata/v1`，冻结，同 plan §5.2）

```json
{
  "schema": "tt.devdata/v1",
  "env": "福联测试tst", "module": "CXM", "prog": "axmt500_wf",
  "fetchedAt": "2026-11-21T14:30:00+08:00", "ttVersion": "x.y.z",
  "source": { "topDir": "/u1/toptst", "ent": 1108 },
  "versions": {
    "dzaa002": 3,
    "tglHeader": { "sdVer": "0003(2025-11-13 17:33:55)", "prVer": "…" },
    "templateVersions": { "code_i02": 123, "a05": 3 }
  },
  "files": [
    { "role": "tab", "src": "erp/cxm/dzx/tab/axmt500_wf.tab",
      "rel": "dzx/axmt500_wf.tab", "sha256": "…", "size": 93468,
      "srcMtime": "2026-11-21T12:27:32+08:00" }
  ],
  "missing": ["dzx/axmt500_wf.tgx2", "inc/axmt500_wf_rep.4gl"],
  "warnings": ["tab(12:27) 比 tgl(12:27) 早 0 分，可能未重产"]
}
```

规则（plan §5.2 原文要点）：

- `files[]` 是「这份材料的身份证」，**每个文件都要有 `sha256`**；缺失的进 `missing[]`（不判失败）。
- `versions` 是给 model 判「规则是否仍适用」的锚点（模板版本、`DZAA` 版次）。
- `warnings[]` 承载一致性提示（如 `tab` 比 `tgl` 新 → 产物滞后），不阻断。
- M 侧 manifest 也用同一 schema（`prog` 置空）；**表清单见下面第 3 条的落地修正**。

**本文件追加的冻结项（不冲突，需 P1 实现时遵守）：**

1. `schema` 取值固定为 **`tt.devdata/v1`**；不得另造版本号。
2. `files[].role` 的取值集合见第 1 节角色表，共 9 个（**落地修正，见第 3 条**）：

   | 侧 | role | 对应 |
   |---|---|---|
   | M | `dict` | 规格库本身（`dict/<env>.db`，一个） |
   | M | `mdl` | 模板树里落盘的每个文件（`code_*`、`slice/*`、`*.snippet`、`4rp_*`、xml、…） |
   | D | `tab` / `tsd` / `tap` / `tgl` / `tgx` / `tgx2` | `.tab`/`.tsd`/`.tap`/`.tgl`/`.tgx`/`.tgx2` |
   | D | `4gl` | `src/<prog>.4gl` |
   | D | `inc` | `inc/*_mask.4gl`、`inc/*_rep.4gl`（两者都用 `inc`，靠 `rel` 区分） |
   | D | `form` | `form/<prog>.42f` |
   | D | `design` | `design/<prog>.tzs` |

3. **M 侧表清单的落地修正**（P1 实现时定，plan §6.1「补充 1」已记录）：M 侧的 `files[]`
   **只登记真正落盘的文件**（`role=dict` 一个 + `role=mdl` 每个模板一个，每个都有真实 sha256），
   表清单另放 `tables[]: [{table, rows, skipped?, missing?, desc}]`。

   为什么不按原计划把每张表塞进 `files[]`：表没有独立文件（整库一个 `.db`），硬给每张表
   编一个 sha256 只会变成「看起来校验过、其实校验的是整库」的假证据。`tables[]` 的长度
   恒等于 53（角色表张数），所以「镜了几张/缺几张/跳过几张」不需要集合差集。
   skipped 的表同时进 `warnings[]`。

---

## 7. HTTP 契约与 CLI 命令面（草案）

> 本节内容之后进 `internal/devdata/README.md`；与 plan §5.5 / §5.7 **逐条一致**，只补 plan 未写死的细节并标注「待确认」。

### 7.1 挂载

- `web.Options` 增 `DevData http.Handler`；`internal/web/server.go` 加 `m.Handle("/api/devdata/", s.opt.DevData)`；`internal/cli/serve.go` 构造并注入。
- `nil` 时该前缀 404（与 `Debug`/`Dict` 同规则）。
- 错误信封沿用共享层：`{ok:false, code, error}`。

### 7.2 5 个端点

| 方法 | 路径 | 请求 | 成功 | 错误 |
|---|---|---|---|---|
| GET | `/api/devdata/status?env=` | query：`env`（必填）、`probe=1`（可选，向服务器 stat 比对 stale） | `200`：M 状态 + 程序列表（含 stale）+ `skipped[]` + 运行中进度；不触发网络（除非 `probe=1`） | `404 NOT_FOUND` env 不存在；`400 USAGE` 缺 `env` |
| POST | `/api/devdata/mirror` | body `{env, only?: "dict" 或 "mdl", force?}` | `200`：受理并返回进度载荷；同类在跑 → `409` | `404 NOT_FOUND` env 不存在；`400 USAGE` `only` 非法；`409 CONFLICT` 已在跑 |
| POST | `/api/devdata/fetch` | body `{env, prog, module?, withDesign?, form?, force?}` | `200`：受理；产物落 `prog/<prog>/` | `400 USAGE` prog 非法（不匹配 `^[a-z][a-z0-9_]{2,31}$`）；`404 NOT_FOUND` env 不存在或程序在服务器不存在；`409 CONFLICT` 同类在跑 |
| POST | `/api/devdata/stop` | body `{env, kind}`（`kind` = `dict`/`mdl`/`fetch`） | `200`：取消受理 | `409 CONFLICT` 没在跑；`404 NOT_FOUND` env 不存在 |
| DELETE | `/api/devdata/prog?env=&prog=` | query：`env`、`prog`、`yes=1`（**必须**） | `200`：删除该程序目录 | `400 USAGE` prog 非法或缺 `yes`；`404 NOT_FOUND` 未拉取 |

错误码口径（plan §5.5）：

- `env` 不在配置清单 → `404 NOT_FOUND`
- `prog` 非法 → `400 USAGE`
- 未拉取查 status / 删未拉取的 prog → `404`
- 同类任务在跑 → `409`
- 409 的 `code` 名 plan 未写死，本节暂定 `CONFLICT`（**待确认**，实现时与 dict 侧现有 conflict code 对齐）。

进度载荷（plan §5.4）：

```json
{ "running": true, "kind": "fetch", "env": "福联测试tst", "prog": "axmt500_wf",
  "done": 3, "total": 7, "currentFile": "dzx/axmt500_wf.tgl",
  "startedAt": "2026-11-21T14:30:00+08:00", "error": null }
```

### 7.3 6 条 CLI

| 命令 | 参数 | 行为 |
|---|---|---|
| `tt gen status` | `[--at <env>] [--probe] [--json]` | M 状态 + 程序列表 + stale；`--probe` 才向服务器 stat；`--json` 为默认信封（`output.Emit`），`--human` 出小表 |
| `tt gen mirror` | `--at <env> [--only dict 或 mdl] [--force]` | 全量镜像；前台跑并输出进度 |
| `tt gen fetch` | `<prog> --at <env> [--module CXM] [--with-design] [--form] [--force]` | 按需整套拉取 |
| `tt gen ls` | `[--at <env>]` | 已拉取程序清单（小表） |
| `tt gen check` | `<prog> --at <env>` | 只比对服务器 mtime/版本，输出 stale 明细 |
| `tt gen clean` | `[--at <env>] [--prog <prog>] [--yes]` | 删除；无 `--yes` 只打印将删内容（dry-run） |

约定：

- CLI 与 Web **共用 `internal/devdata` 领域层**，不复制逻辑；Web 走 HTTP，CLI 直接调包（plan §5.7）。
- 错误码对齐 HTTP：程序不存在 `404`、程序名非法 `400`、任务在跑 `409`、未拉取 `404`。
- 帮助首句固定：「为 model 准备的原材料；与 `tt file mirror`（阅读用）不是一回事」。
- `tt gen status` 必须显示 §4.4 的 `mirrored/skipped` 计数与 skipped 表清单。

---

## 8. `$TOP/erp/mdl/` 实测清单

### 8.1 总量与结构

| 位置 | 文件数 | 字节 | 说明 |
|---|---:|---:|---|
| `erp/mdl/` 顶层 | 154 | 3,026,050 | 含 2 个子目录 `slice/`、`xrp/` |
| `erp/mdl/slice/` | 126 | 178,573 | `code_a00..a77`、`d01..d10`、`qs01..qs27`、`ss01..ss04` |
| `erp/mdl/xrp/` | **待确认** | **待确认** | 本次未测文件数与大小 |
| **合计（已测部分）** | **280** | **3,204,623** | 3,204,623 B ≈ 3.0 MiB |

### 8.2 顶层 154 个文件的细分

| 类别 | 数量 | 说明 |
|---|---:|---|
| `code_*` | **45** | 其中 **38** 个是 `code_*.template`；另 **7** 个是带日期的备份，如 `code_g01.template.141119` |
| `4rp_*.template` | 3 | 报表类主样板 |
| `*.snippet` | 90 | 代码片段（`#mdls` 循环外的复用块） |
| 其它 | 16 | `*.4fd` 表单样板、`ComboxItemSetting.xml`、`UITemplateSetting.xml`、`xbarr.xlsx`、`test-rc3.sh` 等 |
| **合计** | **154** | 45 + 3 + 90 + 16 = 154 |

### 8.3 「模板 171」的口径

- plan / 验收口径的 **171** = `code_*`（**45**）+ `slice/`（**126**）。
- 注意：`code_*` 里真正 `.template` 后缀的是 **38** 个，其余 7 个是历史备份。也就是说「171」是**plan 口径的计数**（含备份文件），不是「38 + 126 = 164」。
- `tt gen status` 报模板数固定用 **171**；若将来想只数有效模板，需**先改 plan**，不要在实现里悄悄改成 164（否则与验收标准 §4.2/§8.4 冲突）。
- `*.snippet`(90)、`4rp_*.template`(3)、`ComboxItemSetting.xml`(1) 是 M 类必拉项，但**不计入 171**。

### 8.4 镜像可行性

- 实测总量 **3,204,623 B（≈3.0 MiB）**，远低于 plan §5.4 的 `mdl/` 总量 200 MB 上限；单文件也远低于 64 MB。
- 因此 `mdl` 镜像**不会**触发任何上限跳过；上限只对 `DZ*` 的 `DZBA_T`/`DZAZ_T` 生效。

---

## 9. 待确认清单（P1 前需要定的）

| # | 事项 | 现状 | 影响 |
|---|---|---|---|
| 1 | `DZBH_T` 的生成器用途 | 只有注释掉的 SQL；`ARCHITECTURE.md:435` 记为 autocomplete a77，但没有生效代码 | 用途列写「待确认」；不影响镜像（整表镜） |
| 2 | `DZBK_T` 的插入方/读取方 | 语料只有 `DELETE FROM dzbk_t` | 用途细节待确认；不影响镜像 |
| 3 | 数据字典来源表清单（`r.t`/`r.v`/`r.q`/`scc` 的底表） | 只知道语料以 `gztz_t` 判表存在（`sadzp030_tab_ma.4gl:764`）；完整清单与行数未测 | 影响 M 侧「53 张之外还要镜什么」与 db 体积 |
| 4 | `mdl/xrp/` 是否纳入 | 文件数与用途未测 | 影响模板数与 M 侧体积 |
| 5 | `mdl/` 其余 16 个顶层文件是否纳入 | plan 只明列 `ComboxItemSetting.xml` | 影响「模板 171」之外的口径 |
| 6 | 每个程序是否都有 `.tap` | plan §4.4 要求齐全；未逐个验证 | 若某些程序无 `.tap`，需回改 plan（ADR-04/§2.5）再放宽 |
| 7 | `prog` → `<mod>` 的映射来源 | plan 只给 `--module CXM` 可选参数 | 影响 fetch 的源路径拼装（不能靠猜） |
| 8 | 409 的 `code` 名 | plan 未写死 | 影响 HTTP 契约文档与前端判断 |
| 9 | `DZ*` 表内容 `sha256` 的算法 | M manifest 要求每表 sha，plan 未规定对什么取 sha | 影响 M manifest 实现与幂等判断 |

---

## 附录：证据文件索引

`codegen` 路径均相对 `.agents/docs/t100/codegen/`：

| 文件 | 在本文中的用途 |
|---|---|
| `01-server/blueprint/sadzp030_tab.4gl` | `DZAG` 判型/三层二显、`DZAC` range、`DZAI` reference、`DZAJ` multilang、`DZAK` mnemonic、`DZCD` 校验型态、`DZEP` def_scc、`DZAD` action、`DZAA` 版次 |
| `01-server/blueprint/sadzp030_tab_ma.4gl` | `DZFI`/`DZFJ` 画面树、`DZAM` 排除、`DZAC` 字段属性、`DZAL` 串查、`DZAJ` 多语言、`DZEP` 分类码、`DZAK` 助记码、`DZCP`/`DZCT` 带值公式门控 |
| `01-server/blueprint/sadzp030_tab_de.4gl` | 单身/子表段落（`DZCP`/`DZCT`/`DZFI`/`DZFJ`/`DZAI` 同构读取） |
| `01-server/blueprint/sadzp030_tab_relation.4gl` | `DZAX` 判型入口、`DZCA` 开窗程序、`DZFF` 树属性、`DZFS` browse 型态、`DZED` PK、`DZBB` 加点检查 |
| `01-server/blueprint/sadzp030_tab_file.4gl` | `DZDZ` 模板最新版次 |
| `01-server/blueprint/sadzp030_tsd.4gl` | `DZAB`/`DZEM`/`DZFU`/`DZFQ` 的 `.tsd` 反向生成 |
| `01-server/generator/adzp150.4gl` | `DZAZ` build 计数、`DZAX` free style、`DZCP`/`DZCT` 简式带值 |
| `01-server/generator/sadzp150_ctrlp.4gl` | `DZCB`/`DZCC`/`DZCY` 开窗切片 |
| `01-server/generator/sadzp150_object.4gl` | `DZCE` 参数个数、`DZCY` 校验参数 |
| `01-server/generator/sadzp150_fieldchk.4gl` | `DZFJ` DateTimeEdit、`DZCP`/`DZCT` 公式、`DZBH`（仅注释） |
| `01-server/generator/sadzp150_lang.4gl` | `DZEB` 多语言显示名 |
| `01-server/generator/sadzp150_multitable.4gl` | `DZEB`/`DZED` 多表样板 |
| `01-server/tgl/adzp176.4gl` / `adzp177.4gl` | `DZBC`/`DZBD` section 组装与维护 |
| `01-server/orchestration/adzp040.4gl` | `DZBA` 首版判断、`DZFK` 属性解析、`DZFQ` form style |
| `01-server/orchestration/adzp031.4gl` | `DZAF` 版次反推 |
| `01-server/orchestration/sadzp060_1/2/3/4/8/ind.4gl` | 各表的**上传/同步写入**证据、`DZAS`/`DZAT`/`DZAQ`/`DZAU`/`DZBU`/`DZYE` |
| `01-server/generator-other/adzp151.4gl`、`adzp152.4gl`、`adzp161.4gl` | `DZGC`/`DZGD`/`DZGG`/`DZKC`/`DZKD` 报表元件；`DZAZ` 计数 |
| `01-server/common/sadzi100_extract.4gl` | `DZDY`/`DZDZ` 模板路径与 MD5 校验 |
| `03-templates/slice/code_a15.template` | `cl_ap_chk_range(...)`（`DZAC` max/min 的落点） |
| `ARCHITECTURE.md` | 流程分段、`DZAG` 三层二显、`DZBH`（研究结论，未验证） |
| `README.md` / `MANIFEST.tsv` | 语料来源/快照时间/逐文件 sha |

复现方法（只读）：

```bash
# 表清单来源（plan 附录 A 同法）
grep -rhoiE 'FROM[[:space:]]+dz[a-z0-9_]+_t' .agents/docs/t100/codegen/01-server | sort -u

# 单表读写定位
grep -rniE '(select|insert into|update|delete from)[^;]*dzac_t' .agents/docs/t100/codegen/01-server
```
