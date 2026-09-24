---
name: erp-read
description: 阅读并分析 TIPTOP ERP Genero BDL 源代码。使用 TDict CLI 将表/字段编号解码为中文含义。当用户打开包含 Genero BDL 代码的 .4gl/.per/.txt 文件，或询问 ERP 程序逻辑、代码中引用的数据库表结构、ERP 源文件中的字段含义时触发本技能。
---

# ERP 代码阅读 — Genero BDL 源代码阅读指南

## 获取源码:本地镜像优先

阅读服务器源码前,先确认本地源码镜像是否存在并新鲜:

```bash
tt dict mirror path [<环境名>]    # 拿到镜像目录绝对路径(如 D:\dev\erp-src\正式区)
tt dict mirror pull  [<环境名>]   # 下载/更新该环境源码镜像(增量;--full 全量重建)
```

- 镜像只含各模块 **4gl**(源码)与 **4fd**(前端字段描述)两棵树,目录与服务器同构(`<镜像根>/<环境名>/erp/<模块>/{4gl,4fd}/...`、`com/{lib,sub,qry,wss}/...`);per/编译产物(42m/42r)/多语言等不在镜像内。
- 拿到路径后用本地文件工具直接读/搜索(ls/rg/按行读),不要 ssh/scp 去服务器;镜像未覆盖的文件(如 per)本地读不到。
- 镜像更新时间滞后于服务器时(他人改过代码),先 `tt dict mirror pull` 再读。
- 环境清单为三个工具共用(`config.json` 的 `hosts.sshs`,用 `tt env list` / `tt env use <名称>` 查看与切换):调试台里加过的机器,这里的镜像也能直接用。

## 读一个 .4gl 的顺序

1. **先问工具"这是什么"**:`tt dict prog <程序编号>`给出中文作业名、程序类别、归属模块、**哪些作业用了它**、**用了哪些表**(见第 4 步)。反向从业务词找程序用 `tt dict prog --kw <业务词>`(找子程序/元件用 `--sub --kw <业务词>`)。
   - **数据源**:缺省走**在线**(默认环境的 ERP 库直查,数据最全,不用先同步);在线连不上会直接报错、不会自动改查本地。要离线/求快就用本地副本:`tt serve`「设置 → 查询数据源」切成本地,或命令加 `--env local`。切到本地后若某族缺失,报错会明说,先 `tt dict db sync`。
   - **子程序/元件/库也查得到**:`gzde_t`(参考作业 azzi901)登记子程序(`aapq110_01`)、应用元件(`cl_abi`)、报表元件等,与主程序表 `gzza_t` **互不重叠**(实测交集 0)。所以 `tt dict prog aapq110_01` 会给出它**自己的**说明(「供应商对账单明细查询报表打印」),并附它的主程序——**别把主程序的用途当成子程序的**,两者常不是一回事(主程序"明细查询" vs 子程序"报表列印")。查不到的就是真没登记(`q_adzi052` 这类开窗走 `tt dict r.q`)。
2. **文件头的修改记录**(`#+ Modifier...` 与 `#add-point:填寫註解說明` 里的历年单号)—— 往往一眼看出这个程序的坑与历史包袱。
2. **文件头的修改记录**(`#+ Modifier...` 与 `#add-point:填寫註解說明` 里的历年单号)—— 往往一眼看出这个程序的坑与历史包袱。
3. **`SCHEMA` / `GLOBALS` / `IMPORT` 段** —— 用到哪些表与公共变量;`IMPORT FGL <lib>` / `CALL s_xxx(...)` 说明真正的逻辑在别的文件里(见第 4 步末)。
4. **先用工具的用表索引(权威,含操作类别)**:`tt dict prog <去掉后缀的编号>` 会给出「使用的表格」。数据来自 T100 自己维护的 `gzdg_t`(参考作业 azzq902),只登记**真实 SQL 访问**,不受 `type_t`、注释、字符串里的表名干扰,并标明操作类别(S=查询/I=新增/U=修改/D=删除)。
   - **两个方向**:`tt dict prog <编号>` = 这个编号用哪些表;`tt dict r.t <表名> --who` = 这张表被哪些程序用(**改表前的影响分析**,实测 `glab_t` 被 376 个程序使用)。表名含义用 `tt dict r.t --brief "a_t,b_t,c_t"` 批量解码(**不加 `--brief` 会打出每张表的全部字段,7 张表就 700 行**);要看字段时再单表 `tt dict r.t <表名>`。
   - **对库/元件/开窗也先试**:索引不只覆盖作业主程序——实测 `cl_abi`(库) 22 条、`s_apcp300`(元件) 8 条、`q_adzi052`(开窗) 9 条。所以读 `com/lib/**`、`com/sub/**` 时,`tt dict prog <库名>` 可能直接给出它用了哪些表。
   - ⚠️ **索引的边界**:子程序(`aapq110_01` 这类,实测 0 条)与动态 SQL(`CURSOR FROM 变量`)不在里面;未登记的新程序也没有。这些才用下面的 grep 兜底。
   - **grep 兜底**:`grep -oE '\b[a-z][a-z0-9]{1,6}_t\b' <文件> | sort | uniq -c | sort -rn`,再交给 `tt dict r.t --brief`。
     - ⚠️ 它的假阳性不少:`type_t` 是 4GL 类型记录**不是表**;被注释掉的 SQL、字符串里的表名(`"dzeb001 IN ('apca_t')"`)、`LIKE 表名.字段` 的变量声明都会算进来(实测 `aapi011` 的 16 个候选里有 5 个是这类噪音——`gzdg_t` 索引不会)。`ds.xxx_t` 这种 schema 限定写法反而会被漏掉。
   - ⚠️⚠️ **grep 出 0 张表 ≠ 这个程序不碰表**。T100 常见"主程序 + 应用元件"结构:明细逻辑在 `com/sub/4gl/s_<程序>.4gl`(元件)或 `com/lib/**` 里,主文件只剩排程/画面骨架,SQL 还可能是动态拼的。**先试 `tt dict prog s_<程序编号>`**——元件自己常在索引里(实测 `s_apcp300` 直接给出 `imaa_t/indc_t/inde_t/indf_t/pcca_t(S/I/U)/rtdx_t`);索引里没有再用 `find <源码树> -iname "*<程序编号>*"` 找同编号文件 grep。实测 `erp/apc/4gl/apcp300.4gl`(691 行)grep 出 **0 张真实表**。**注意元件文件名是 `s_apcp300.4gl`,代码里 `s_apcp300_send` 只是它里面的函数名**。
   - ⚠️ **还找不到就再往上一层找母程序**。`_01`/`_02` 子程序常常只做画面输入/输出骨架,SQL 留在母程序里。实测 `erp/aap/4gl/aapq110_01.4gl`(224 行)grep 出 0 张表,`find -iname "*aapq110_01*"` 只找到 67 行的 `com/inc/erp/aap/aapq110_01_mask.4gl`(同样是空骨架),`s_aapq110*` 根本不存在——真实数据依赖(`apbb_t` 36 次、`apba_t` 11 次、`pmaa_t` 9 次)全在**母程序 `erp/aap/4gl/aapq110.4gl`(3,104 行)**。套路:`grep 0 表 → find 同编号 → 仍 0 表 → grep 去掉 _NN 的母程序`。
   - ⚠️ `_mask`/`_rep`/`_wf` 这类文件**常常是空骨架**(`_mask` 多是遮罩函数空壳,那行 `CALL cl_mask_trans_method(...)` 还可能是注释掉的)。**先 `wc -l` + 上面那条 grep 探一下再决定读不读**;行数少又 0 张表,直接跳到母程序。
5. **查被调用的开窗**:源码里的 `q_xxxxx` / 开窗调用,用 `tt dict r.q q_xxxxx` 查登记说明与 SQL 全文。
6. **样板噪音**:`{<section ...>}` 与 `#add-point:` 以外的部分多是样版产物,有效信息集中在文件头与 add-point 内。

> **重要:T100 的业务语义经常不在 `.4gl` 里,但也不全在元数据里——两边都要看。**
>
> 以 `com/qry/4gl/q_adzi052.4gl`(1,274 行) 为例,文件里是两层:
> - **通用引擎层**:分页/多选/结果过滤(函数名 `q_adzi052_pagedata_fill`、`q_adzi052_rsfilter`、`q_adzi052_sel`…),与业务无关;
> - **内嵌的硬编码 SQL 层**:实测第 573 行起 `FROM ds.dzaf_t af0 ...`,`tt dict r.q` 打出来的 SQL 与它一致——这就是它的业务取数逻辑。
>
> 源码里**没有**的是三样:**中文说明**、**显现/回传列的元数据**、以及「这个文件 = 哪条开窗登记」的**绑定关系**(程序自己在第 142 行 `SELECT dzca001,dzcal003 FROM dzca_t` 向数据库读自己的说明)。
> 所以:`tt dict r.q q_adzi052` 拿「中文说明 + 显现/回传列 + 该取哪些表」,**打开文件**拿「分页/多选/覆写等实现细节」——别只看一边,也别因为一句话就不打开文件。同理 `tt dict desc <表> [字段]` 给出字段绑定的开窗程序与校验码,那是源码里没有的。

> **反向也要知道工具答不出什么,以及源码树自带文档更强的地方。**
>
> - **库与元件有"是什么"的条目**:`tt dict prog cl_abi` → 说明「ABI Library」、规格类别 `B`(应用元件)、归属模块 `LIB`,以及它用了哪些表(实测 21 张);`s_apcp300` 同理。**但它不给"里面有哪些公开函数、各自做什么"**——那是源码与 `docs/` 的活(见下)。
> - **但源码树 `docs/` 下可能自带更完整的说明**:如 `docs/com_lib_4gl公共工具函数库说明.md`(14,109 行 / 103 个 lib 文件 / 1,591 个函数)逐函数给出**用途 + 入参 + 返回 + 输出参数**,还标注了"某函数体已整体注释停用"这类源码深处才知道的坑;另有 `docs/apmt500_framework_guide.html`、`docs/导入导出模板函数开发指南.md`。**读 `com/lib/**` 前先 `ls docs/`**,别急着靠工具或裸读拼。
> - **谁都答不出"为什么这么写"**——那在文件头的 `#+ Modifier...` 修改记录里,逐条读它。

## 命令规范说明

### 模块命名原则

模块名称由三个英文字母组成（适用于 ERP 模块及共用 COM 模块等）。

**第一码**为标准识别码：A=标准 ERP，B=行业专用 ERP，C=客制 ERP，D=客制行业专用 ERP，E=客制自创 ERP。**第二、三码**采用系统英文缩写，若已被使用则第二码改以 X 或其他相关缩写替代。

格式：**AXX**（XX 为 ERP 模块英文缩写代码）

**范例**：AIM（料件主档及库存管理，IM 取自 Item Master / Inventory Management）、AXM（销售管理，因 ASM 已被系统基本资料管理使用而改之）。

**共用模块（$COM）代码**：

| 代码 | 说明 | 英文名称 |
| :---: | :--- | :--- |
| lib | 共用程序 | Common Library |
| lng | 语言资料与提速档副程序 | Lingual sub function |
| sub | 一般副程序 | Common Sub function |
| qry | 查询副程序 | Common Query Function |
| wss | 整合与 Web Service 程序 | Web Service Subsystem |


### 程序命名原则

#### 主程序编号

格式：**SSSQ999_MM**（七码：三码英文 + 四码数字，均小写）

- SSS：模块代码
- Q：程序类别 — i=基本资料维护，m=主档维护，t=交易处理，s=参数设定，p=批次处理，q=查询，r=报表
- 999：三码流水号
- _MM：行业别专用代码（如 _ic=IC 设计业，_sc=鞋服饰，_ph=制药业；省略则通用）

**范例**：aimi100（料件基本资料维护）、cimi001（客制料件维护）、bphi100_ph（食品添加物登记证维护）

#### 子程序与子画面

- **子程序**：SSSQ999**_01**（一般,多支依序 `_02`、`_03`…）/ SSSQ999_MM**_01**（行业专用）
- **子画面**：SSSQ999_s**01** 或 SSSQ999_**01**_s**01**

**范例**：aimi100_01、cimi100_01；aimi100_s01、aimi100_01_s01

> **后缀是 `_01` 系列,不是 `_88`**：实测 `_01` 575 个、`_02` 258 个、`_03` 130 个,而 `_88` 全树只有 1 个(旧规格写法)。按 `_88` 去找子程序会找不到文件。

#### Library 文件

- **标准**：cl_xxxxxxx（共用程序，如 cl_err、cl_about），长度 1–17 字符
- **客制**：ccl_xxxxxxxxx（如 ccl_trim），长度 1–17 字符

> 客制共用程序暂不开放，lib 功能已完备。

#### 应用元件

- **标准**：s_XXXXXXX 或 s_SSSQ999_XXXXXX（如 s_transaction）
- **客制**：CS_XXXXXXXXX 或 cs_SSSQ999_XXXXXX（如 cs_date）

#### 报表相关文件

报表主程序命名同 2.1。报表元件格式（后缀同为 `_01` 系列,`_g88`/`_x88` 在本树不存在）：

| 类型 | 标准 | 行业专用 |
| :---: | :--- | :--- |
| 报表元件 | SSSQ999_g01（多件依序 `_g02`…） | SSSQ999_MM_g01 |
| 凭单列印元件 | SSSQ999_k01 | SSSQ999_MM_k01 |
| 查询列印元件 | SSSQ999_x01（多件依序 `_x02`…） | SSSQ999_MM_x01 |

**实测例**：`aapp350_g01`(退貨折讓單列印作業)、`aapr110_k01`(供應商貨款對帳憑單列印)、`aapq110_x01`(供應商對帳單明細查詢列印)。

- **报表结构档（.rdd）**：与元件同名，如 axmr500_g01
- **报表样板（.4rp）**：主报表 SSSQ999_g01（多样板加 _77），子报表加 _subrep66

#### 其他常见后缀（实测）

| 后缀 | 含义 | 例 | 本树数量 |
| :---: | :--- | :--- | ---: |
| `_wf` | Workflow 流程相关（`com/lib/cl_bpm_wf`、`q_*_wf`） | axmi125_wf、q_bxmi002_wf | 487 |
| `_rep` | 报表子程序片段，在 `com/inc/erp/<模块>/<母程序>_<NN>_rep.4gl` | aapt110_01_rep | 1,342 |
| `_01`…`_09` | 子程序（见上） | aapq110_01 | 575 / 258 / 130 |

**范例汇总**：

| 场景 | 主程序 | 报表元件 | 样板 |
| :--- | :--- | :--- | :--- |
| 标准报表 | axmr500 | axmr500_g01 | axmr500_g01 / axmr500_g01_subrep01 |
| 多样件 | axmr500 | axmr500_g02 | axmr500_g02 / axmr500_g02_subrep01 |
| 多样板 | axmr500 | axmr500_g01 | axmr500_g01_02 / axmr500_g01_02_subrep01 |
| 行业别 | axmr500_ph | bxmr500_ph_g01 | bxmr500_ph_g01 / bxmr500_ph_g01_subrep01 |
| 客制标准 | axmr500 | axmr500_g01 | axmr500_g01 / axmr500_g01_subrep01 |
| 新增客制 | cxmr501 | cxmr501_g01 | cxmr501_g01 / cxmr501_g01_subrep01 |

#### 查询元件

- **标准**：q_XXXXXXX（如 q_imaa001），长度 1–18 字符，以表格名命名
- **客制**：cq_XXXXXXXXX，长度 1–17 字符
- **行业专用**：q_XXXXXXXXX_mm（如 q_ooea001_ph），前缀长度 1–15 字符

> 避免与动态查询副程序名称冲突。

#### Web Service 程序

- **主程序**：wssp999（标准）/ cwssp999（客制）
- **子程序**：wssp999**_01**（标准）/ cwssp999**_01**（客制）（实测 `com/wss/4gl/wssp000_01.4gl`）

存放于 $COM/WSS，子程序不对外供一般 ERP 调用。

#### 共用参数文件（.inc）

- **一般用途**：SSSQ999_**01**.inc（放 4g1 目录）
- **跨模块**：top_XXXXXX.inc（放 $ERP/cfg，链接至 $COM/cfg）

#### 扩展名说明

| 类别 | 扩展名 | 说明 |
| :--- | :--- | :--- |
| 原始程序 | 4g1 / inc / 42m / 42r | 源码 / 共用参数 / 编译目标 / 可执行 |
| 画面 | 4fd / per / 42f | 设计档 / 对照档 / 编译档 |
| 报表 | rdd / 4rp | 结构档 / 样板档 |
| 资源 | str / 42s / sch / 4tm/4ad/4tb / 4pw | 翻译原始档 / 翻译编译档 / 参考内容 / Genero 设定 / 专案设定 |


### 函数命名原则

格式：**xxxxxxxxx_yyyy**（前 7 码为程序文件名，yyyy 为功能描述）

**范例**：azzi100_insert

**常用函数**：_insert、_delete、_show、fetch、_modify、_input、_query、set_entry、set_no_entry

> 同一主程序下函数名不可重复。


### 变量命名原则

| 类型 | 格式 | 说明 |
| :--- | :--- | :--- |
| 全局变量 | g_XXXXXXXXX | 定义于 $TOP/cfg/top_global.inc，如 g_gui_type、g_errno |
| 局部变量 | l_XXXXXXXXX | 仅限当前 Function 有效 |
| 传递参数 | p_XXXXXXXXX | 跨函数传递 |
| 屏幕变量 | s_detailN / s_browser | N 为流水号，异动可能影响代码生成 |


### 数据库表格命名原则

#### 表格名称

格式：**xxxx_t**（4 码小写 + _t），前两码为模块名，后两码流水号。如 imaa_t（料件主档）。

- **行业包辅助表**：xxxxmm_t（如 imaaic_t、imaasl_t）
- **客制表**：标准表不可删，新建客制表加 uc（如 imaauc_t）

#### 字段名称

格式：**xxxx999z**（表格编号 + 3 码流水号 + 阶层码）。如 imaa001（料件编号）、imaa001a（子阶）。

- **客制字段**：客制表用 apqquc001，标准表加 ud（如 apqqua001、imaaud001）
- **行业包辅助字段**：xxxxmm099（如 imaaic001）
- **自定义字段**：主/明/交易档加 ud，001–010 文字型（C003），011–020 数值型（N101），021–030 日期型（D002）

**固定字尾规范**：

| 用途 | 字尾 | 说明 |
| :--- | :--- | :--- |
| 建立 | crtid / crtdt / crtdp | 员工 / 日期 / 部门 |
| 拥有 | ownid / owndp | 员工 / 部门 |
| 修改 | modid / moddt | 员工 / 日期 |
| 确认 | cnfid / cnfdt | 员工 / 日期（adp 负责） |
| 过账 | pstd / pstdt | 员工 / 日期（adp 负责） |

**通用字段**：

| 字段 | 字尾 | 类型 | 备注 |
| :--- | :--- | :--- | :--- |
| 状态码 | stus | C001 | |
| 单号 | docno | C203 | |
| 单据日期 | docdt | D001 | T 类表单号后必跟日期 |
| 项次 | seq / seq1 / seq2 | N004 | 需逐层依赖 |
| 企业编号 | ent | N802 | |
| 法人 | comp | C813 | |
| 营运据点 | site | C813 | |
| 账别 | ld | C501 | 小写 |
| 组织对象 | unit | C007 | |
| 账务归属 | orga | C007 | |
| 核算组织 | legl | C007 | |
| 时间戳记 | stamp | D003 | 不可为 KEY |

**必要字段**：主/参数/基础资料档需 crtid、crtdp、crtdt、ownid、owndp、modid、moddt、stus；交易档另加 cnfid、cnfdt。

#### 索引与键值

- **索引**：xxxx_Qy（Q=n/u，y 流水号），如 imaa_n01、imaa_u01
  - 客制新表比照标准（imaauc_n01）；标准表加客制索引前缀 tic_（tic_imaa_n01）
  - 行业包辅助表：imaaic_n01
- **Primary Key**：xxxx_pk（r.t 自动命名，唯一不可改），如 imaa_pk
  - 客制：imaauc_pk；行业包：imaaic_pk
- **Footing Data**：xxxxxx_fdy（r.t 自动命名，仅供 r.a 及产生器参考，不实际建库），如 imaa_fd
  - 客制：imaauc_fd；行业包：imaaic_fd
  
## 工具参考

- 查询数据字典(表/字段中文含义):参见 [docs/WIKI.md](../../docs/WIKI.md#7-tt-dict-数据字典)
- BDL 语言语法/库函数参考:项目内 `docs/bdl`(Genero BDL 文档;路径用 `tt dict bdldoc dir` 查/设)