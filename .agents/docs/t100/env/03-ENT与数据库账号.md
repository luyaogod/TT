# 03 · T100 登录、`azzi000` 作业、ENT 切换与数据库账号

## 一、作业是怎么起来的

```
菜单别名 menu           →  r.r azzi000
r.r（$TOP/utl/bin/r.r） →  cd $TEMPDIR ; $FGLRUN $AZZ/42r/azzp000.42r  azzi000
azzp000.42r（作业加载器）→  读登录参数、建数据库连接、再启动 azzi000
azzi000（T100 系统首頁）→  SCHEMA ds
```

- `azzi000` 源码在环境里：`$TOP/erp/azz/4gl/azzi000.4gl`（532 行，`CONSTANT ls_version = "1.3.01 build-20260327"`）。
  首行注释写明「該程式已解開Section, 不再透過樣板產出!」「該程式為freestyle程式!」。
- `azzp000` **只有编译件** `$TOP/erp/azz/42r/azzp000.42r`（79391 B），源码不在环境里。
- `webrun`（Web 侧）走同一个 `azzp000`。旧版 `/u3/pub/etc/webrun` 里有一段"授权满则把首页换成 `adzi444`"：

```ksh
if [ "$prog" == "azzi000" -a `$TOPKEY/flc2.sh` == "1" ] ; then prog='adzi444' ; fi
```

  ⚠ 但**实际在跑的 `/u1/etc/webrun` 里这段是注释掉的**（且引用的是 `flc.sh`），
  所以当前 Web 登录不会被改道到 `adzi444`。见 [06](06-Web入口wt35prd.md#八--修正两份-webrun实际用的是-u1etcwebrun)。

### `azzi000` 做了什么（与数据库相关）

| 行 | 内容 |
|---|---|
| `azzi000.4gl:65` | `SCHEMA ds` —— **登录界面自己走 `ds` 系统账号** |
| `:67` | `GLOBALS "../../cfg/top_global.inc"` |
| `:140` | `loadActionDefaults(FGL_GETENV("ERP")+"/cfg/4ad/homepage.4ad")` —— 环境变量直接参与路径拼装 |
| `:147` | `FGL_GETENV("T100HOMEADDON")` 决定 V1/V3 首页 |
| `:111-121` | 支持 `-V`/`-v` 参数打印版本后直接 `EXIT PROGRAM` |

`azzi000` 本身**不选企业、不切库**——它收到的 `g_enterprise` 已经是上游设好的。
真正把「企业」翻成「数据库账号」的是下面两层。

## 二、企业（ENT）在哪、怎么影响数据库

### 2.1 三层概念，别混

| 概念 | 载体 | 谁能改 | 影响 |
|---|---|---|---|
| `TOPENT` | **环境变量**，`topenv` 里 `${TOPENT:=99}` | 登录前的 shell | 工具/批处理的**默认企业号** |
| `g_enterprise` | `top_global.inc` 的 GLOBAL（`LIKE type_t.num5`） | T100 运行期 | 每条业务 SQL 的 `ent` 条件 |
| `g_dbs` | `top_global.inc` 的 GLOBAL，注释「實體資料庫名稱」 | T100 运行期 | **当前数据库连接名** |

`TOPENT` 只在"没有值时"取 99。代码里的读法统一是"缺省 99"：

```4gl
LET li_topent = NVL(FGL_GETENV("TOPENT"),"99")     # sadzi140_db.4gl:10602 / sadzp148.4gl:2236
LET g_enterprise = FGL_GETENV("TOPENT")            # sadzi440_tool_init.4gl:231
```

所以：**`TOPENT` 只是"开门时站在哪个企业"，不是"切换企业"的开关。**

### 2.2 ENT → 账号的权威表：`gzou_t`

字段（按用法反推）：

| 字段 | 含义 |
|---|---|
| `gzou001` | 企业编号（ENT） |
| `gzou003` | **实体数据库名 / schema 名**（如 `dsdemo`、`dsdata`、`dsstd`） |
| `gzou005` | 日期年月格式之类，`cl_abi` 里 `SELECT gzou005 ... WHERE gzou001=g_enterprise` 取 |
| `gzoustus` | `Y` 启用 |

**`gzou_t` 永远用系统账号 `ds` 读**——它是跨企业共享的系统表。
（tt 自身的两条路线都是这么做的：客户端直连与服务器 sqlplus 都用 `ds`。）

### 2.3 实测数据

`t35tst`（区域 35，18 个启用企业）：

```
SCHEMA_NAME  ENT_CNT
dsdemo       17
dsstd         1        ← ENT 2222
```

`t35prd`（区域 36，27 个启用企业）：

```
SCHEMA_NAME  ENT_CNT
dsdata        7
dsdemo       19
dsstd         1        ← ENT 2222
```

**这两个分布正好对上 `fglprofile` 里的 4 个别名**（`ds`/`dsdemo`/`dsdata`/`dsstd`），
其中 `ds` 是系统账号（`gzou_t` 的入口），另外三个是业务数据 schema。
所以「不同 ENT 用不同数据库账号」在这台机器上是**事实**，不是配置可能性。

### 2.4 库别名表：`fglprofile.<ZONE>`

`/u3/pub/etc/fglprofile.t35tst` 全文只有 4 组、每组 4 行：

```ini
dbi.database.ds.source   = "t35tst"      dbi.database.ds.username = "ds"      dbi.database.ds.password = "<REDACTED>"      dbi.database.ds.schema = "ds"
dbi.database.dsdemo.source = "t35tst"    dbi.database.dsdemo.username = "dsdemo"  ...
dbi.database.dsdata.source = "t35tst"    dbi.database.dsdata.username = "dsdata"  ...
dbi.database.dsstd.source  = "t35tst"    dbi.database.dsstd.username  = "dsstd"   ...
```

`source` 是 **TNS 别名**，与 `$ZONE` 同名（`t35tst`），由
`$ORACLE_HOME/network/admin/tnsnames.ora` 解析到 `top109ind:1521/<SERVICE_NAME>`。
`username`/`password` 是**明文**，且口令多与账号名相同（样例值已脱敏为 `<REDACTED>`）。
Genero 的 `dbi.database.<别名>` 就是「连接名」，BDL 里 `SET CONNECTION <别名>` 用它。

## 三、运行期怎么切（这是"ENT 切换"的真面目）

### 3.1 三个底层 API

源码里找不到 `cl_db_connect` 的定义，只有调用——实现在编译库
**`$TOP/com/lib/42m/lib_cl_db.42m`**（源码未随环境提供）。三个动词：

| API | 作用 | 证据 |
|---|---|---|
| `cl_db_connect(<dbname>, <bool>)` | 连接到 `fglprofile` 里的某个别名 | `cl_eai.4gl:497`、`azzp820.4gl:828`、`adzi140.4gl:447` |
| `cl_db_disconnect_current()` | 断开当前连接 | `azzp820.4gl:859` |
| `cl_db_setconnect(<name>)` | 把当前连接切回指定别名（不断开） | `azzp820.4gl:861` |

上层再包一层 BDL 语句：`SET CONNECTION g_dbs`（`cl_xg.4gl:2774`、`cl_xr.4gl:1311`、
`azzi768.4gl:385`）——**`g_dbs` 就是当前连接名**。

### 3.2 标准切换套路（`azzp820.4gl` 的循环是范本）

```4gl
LET g_dbs_old = g_dbs                                    # :823  记住原连接
FOR li_cnt = 1 TO ga_gzou.getLength()                    # 遍历所有企业
    IF g_dbs_old = ga_gzou[li_cnt].gzou003 THEN
    ELSE
        CALL cl_db_connect(ga_gzou[li_cnt].gzou003,FALSE) # :828 切到该企业的 schema
    END IF
    ... 干活 ...
    IF g_dbs_old = ga_gzou[li_cnt].gzou003 THEN
    ELSE
        CALL cl_db_disconnect_current()                   # :859 断开
        IF cl_db_setconnect(g_dbs_old) THEN END IF         # :861 切回
    END IF
END FOR
```

这段代码同时证明了两件事：
1. 连接名的取值就是 `gzou003`（即 schema 名）；
2. 切换是**进程内的连接切换**，全程没有动任何环境变量。

### 3.3 跨 schema 读的另一条路（不改连接）

`cl_get.4gl:806-815`：

```4gl
LET ls_gzou003 = g_dbs
IF g_dbs = "ds" THEN                                     # 当前是系统库时
   SELECT gzou003 INTO ls_gzou003 FROM gzou_t WHERE gzou001 = lc_enterprise
   IF cl_null(ls_gzou003) THEN LET ls_gzou003 = g_dbs END IF
END IF
LET ls_sql = "SELECT ",ls_tableid,"002 FROM ",ls_gzou003 CLIPPED,".",lc_gzsz001 CLIPPED, ...
```

即：**当前连接是 `ds` 时，用 `gzou_t` 换出目标 schema，再用 `schema.table` 限定名读**。
`cl_gr.4gl:8111` 的注释把这条规则讲得最直白：

> `例如:AZZ/ADZ模組的作業執行時,g_dbs='ds',則此時要抓聯絡對象資料會抓不到,需要改抓dsdemo的資料`

**这解释了为什么登录/系统类作业跑在 `ds` 上，而业务作业跑在 `dsdemo` 上。**

### 3.4 第二张映射表：`gzda_t`（ERP 库 → 中间/AWS 库）

`cl_eai.4gl:453` 的 `cl_eai_db_connect(ps_dbs)`：

```4gl
SELECT gzda010 INTO ls_middb FROM gzda_t WHERE gzda001 = ls_dbs AND gzdastus = 'Y'
IF cl_null(ls_middb) THEN                    # 取不到就退回硬编码
   CASE ps_dbs
     WHEN 'dsdemo' LET ps_middb = 'dsawst'
     WHEN 'dsdata' LET ps_middb = 'dsaws'
     OTHERWISE     LET ps_middb = 'dsawst'
   END CASE
ELSE LET ps_middb = ls_middb
CALL cl_db_connect(ps_middb, FALSE)
```

实测 `t35tst` 的 `gzda_t`（由作业 **`azzi085`** 维护）：

| gzda001 | gzda010 | stus |
|---|---|---|
| `dsdemo` | **`dsawst`** | Y |
| `dsdata` | **`dsaws`** | Y |
| `ds` `dsaps` `dsaws` `dsawst` `dscom` `dspatch` `dsrept` `dsstd` `mongo` | （空） | 多数 Y |
| `dsc_trfdb` | （空） | **N** |

注意 `dsdemo→dsawst`、`dsdata→dsaws` 与代码里的兜底**完全一致**，说明这张表当前就是按默认值配的。

## 四、这台库上到底有哪些账号

`t35tst` 的 Oracle 用户（`all_users like 'DS%'`，12 个）：

```
DS  DSAPS  DSAWS  DSAWST  DSBACK  DSC_TRFDB  DSDATA  DSDEMO  DSPATCH  DSREPT  DSSTD  DSTOOL
```

表数量（`all_tables` 按 owner）：

| schema | 表数 | 角色 |
|---|---|---|
| `DS` | 4,009 | 系统/字典（`gzou_t`、`gzza_t`、`gzze_t`…） |
| `DSDEMO` | 29,296 | **业务数据主库** |
| `DSDATA` | 3,267 | 另一套业务数据（t35prd 才启用 7 个企业） |
| `DSSTD` | 3,304 | 标准/样板库 |
| `DSAPS` `DSAWS` `DSAWST` `DSREPT` `DSPATCH` `DSTOOL` `DSBACK` `DSC_TRFDB` | — | 排程/集成/报表/补丁/工具/备份/传输 |

连接账号 = Oracle 用户（实测 `select user from dual` 在 `dsdemo` 账号下返回 `DSDEMO`）。

## 五、结论：ENT 切换对"环境变量"的影响

**几乎没有影响——这是最重要的结论。**

| 问题 | 答案 |
|---|---|
| 切 ENT 会改 `TOPENT` 吗 | 不会。`TOPENT` 由登录 shell 给，运行期没有 `FGL_SETENV("TOPENT")` 的调用点（全树 grep 无命中） |
| 切 ENT 会改 `FGLPROFILE`/`TWO_TASK` 吗 | 不会。这些是 `$ZONE` 级，一个区一个值 |
| 切 ENT 实际改了什么 | ① `g_enterprise` 全局；② 数据库**连接**（`cl_db_connect(gzou003)`）+ `g_dbs` |
| 不同 ENT 用不同数据库账号吗 | **是**，账号由 `gzou_t.gzou003` 决定，实测同区并存 `dsdemo`/`dsdata`/`dsstd` |
| 想改默认企业怎么办 | 改 shell 的 `TOPENT`（或 `topenv` 里的默认 99），**不是**改菜单 |

因此排查"数据看起来在另一个库"时，顺序应该是：
`ZONE`（菜单）→ `FGLPROFILE` → `gzou_t` 的 `gzou003` → 当前 `g_dbs`/连接。

## 六、`top_global.inc` 里的相关声明

`$TOP/com/cfg/top_global.inc` 第一行就是 `SCHEMA ds`，然后：

```4gl
DEFINE g_enterprise  LIKE type_t.num5    #體系 Enterprise 代碼
DEFINE g_dbs         LIKE type_t.chr21   #實體資料庫名稱
DEFINE g_site        LIKE type_t.chr10   #營運據點名稱
DEFINE g_account     LIKE type_t.chr20   #使用者編號/帳號
DEFINE g_user        LIKE type_t.chr20   #使用者工號
DEFINE g_clienttype  LIKE type_t.chr1    #A:GMA / W:GWC / D:GDC
```

`g_clienttype` 说明同一套程序要同时服务 2-tier（GDC）与 4-tier（GWC/GMA），
这也是为什么连接与环境的处理都收在库里、而不是散在各作业里。

另有一处专用常量（设计器/工具用双重 schema）：

```4gl
CONSTANT g_sys_schema2 = "dsdemo"    # adzi888_global.inc:66  「系統參照的DB schema」
CONSTANT g_exp_para2   = ...         # 对应的 dsdemo 口令
```
