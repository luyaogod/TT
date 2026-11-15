# 原始代码索引

本目录下的 `u1/`、`u3/`、`etc/` 是**服务器绝对路径的镜像**：
`u3/pub/bin/chenv` 即服务器上的 `/u3/pub/bin/chenv`，`etc/httpd/conf/fastcgi.conf`
即 `/etc/httpd/conf/fastcgi.conf`。共 **68 份**。

> **⚠ 已脱敏**：镜像文件**不是逐字节原样** —— 口令值已替换为 `<REDACTED>`
> （`fglprofile.*` 的 `password`、`multipoijdbc.ini` 的 `pass=`、`poijdbc.ini` 的口令行、
> `azzi085.4gl` 里注释残留的 `pass=` 样例）。除此之外的内容与服务器一致。
`u3/pub/bin/chenv` 即服务器上的 `/u3/pub/bin/chenv`。
文件名、内容与服务器逐字节一致（经 `base64` 回传解码，无换行转换）。

## 登录 / 环境变量链路

| 文件 | 服务器路径 | 作用 |
|---|---|---|
| `u3/pub/bin/chenv` | `/u3/pub/bin/chenv` | **登录入口**：菜单、选 ZONE、分发 env、`FGLSERVER`、`tr.*/pr.*` 别名 |
| `u3/pub/etc/chenv` | `/u3/pub/etc/chenv` | 按 ZONE 出 `TOPCHKOUT/TOPALM/TOPIND/CUST/DGENV/ORG`（含失效的 `k35prd`） |
| `u3/pub/etc/topenv` | `/u3/pub/etc/topenv` | **env 引擎（交互侧）**：Oracle/Genero/TOP/路径/CLASSPATH |
| `u3/pub/etc/topenv.kingbase` | 同目录 | Kingbase 变体（unset Oracle，改 `KINGBASE_*`），当前不可达 |
| `u3/pub/etc/topenv.0824` | 同目录 | 历史版本（0824），用于比对变更 |
| `u3/pub/etc/topsys` | `/u3/pub/etc/topsys` | 生成 205 个模块变量 + `FGLLDPATH` |
| `u1/etc/topenv` | `/u1/etc/topenv` | 旧版 env 引擎：`FGLPROFILE=/u1/etc/…`，`ORACLE_SID/TWO_TASK` 为 readonly。**当前无调用者**（只被旧的 `/u3/pub/etc/webrun` 引用） |
| `u1/etc/chenv` | `/u1/etc/chenv` | Web 侧旧版旗标：按 `LOGNAME` 判 `DGENV`，`TOPIND=sd`、`TOPCHKOUT=Y` 恒定 |
| `u1/etc/topsys` | `/u1/etc/topsys` | 两份 `topsys` 内容相同（1084 B） |
| `u1/etc/topalias` | `/u1/etc/topalias` | **命令别名**：`menu=r.r azzi000`、`r.t`、`account` 等 |
| `u3/pub/etc/webrun` | `/u3/pub/etc/webrun` | 旧版 Web 启动器（source `/u1/etc/*`，**已不被引用**） |
| `u1/etc/webrun` | `/u1/etc/webrun` | **GAS 实际调用的启动器**（source `/u3/pub/etc/*`，追加 `fglprofile.cpu` / `fgllicense.cpu`） |
| `u3/pub/etc/profile.tiptop` | 同目录 | 另一版登录脚本（`chenv` 的完整内联版，额外打印 `TOPENT`） |
| `u1/etc/profile.tiptop` | `/u1/etc/profile.tiptop` | 同上（/u1 变体） |
| `u1/etc/profile.topapp` | 同目录 | topapp 用户版登录脚本 |
| `u1/etc/profile.topman` | 同目录 | topman 用户版 |
| `u1/etc/profile.topstd` | 同目录 | topstd 用户版（`DGENV=s`） |
| `u1/usr/tiptop/.profile.st` | `/u1/usr/tiptop/.profile.st` | **老版**菜单：码 `1/2/3/4` = topprd/toptst/topstd/toppth |
| `u1/usr/tiptop/.kshrc` | 同目录 | ksh 交互配置（只 source `/etc/kshrc` + emacs 模式） |
| `u3/pub/bin/tjob` | `/u3/pub/bin/tjob` | **排程作业**：`su tiptop -c '.profile 35; $FGLRUN …'` 用法范例 |
| `u3/pub/etc/auto_start` | 同目录 | 开机自启（FLM/mount/jetty），同样用 `.profile <码>` |
| `u1/etc/auto_start` | `/u1/etc/auto_start` | 同上（/u1 变体） |
| `u3/pub/bin/adduser.sh` | 同目录 | 建用户脚本（说明新用户也套同一套 profile） |

## 数据库连接信息

| 文件 | 服务器路径 | 作用 |
|---|---|---|
| `u3/pub/etc/fglprofile.t35tst` | 同目录 | **库别名/账号/口令（明文）** — 测试区 |
| `u3/pub/etc/fglprofile.t35dev` | 同目录 | 开发区 |
| `u3/pub/etc/fglprofile.t35prd` | 同目录 | 正式区 |
| `u3/pub/etc/fglprofile.t35pth` | 同目录 | PATCH 区 |
| `u3/pub/etc/fglprofile.topprd` | 同目录 | 出货区 |
| `u1/etc/fglprofile.t35tst` | `/u1/etc/fglprofile.t35tst` | Web 侧用的同一份别名表 |
| `u1/t35tst/utl/java/jar/multipoijdbc.ini` | 同路径 | Java 工具连接（`<zone>.<acct>.host/port/user/pass`，**base64**）；**本机这份是 t30prd 的残留** |
| `u1/t35tst/utl/java/jar/poijdbc.ini` | 同路径 | POI/Excel 的 JDBC 目标；同样过期 |
| `u1/genero/as-t35tst.xcf` | `/u1/genero/as-t35tst.xcf` | GAS 配置（`ZONE=t35tst`，offset 95） |

## Web 入口链路（`wt35prd` 之谜，见 [06](06-Web入口wt35prd.md)）

| 文件 | 服务器路径 | 作用 |
|---|---|---|
| `etc/httpd/conf/fastcgi.conf` | `/etc/httpd/conf/fastcgi.conf` | **`ProxyPass /wt35prd/ → fcgi://…:6398/`**；`/prd/`、`/tst/` 等短链跳转 |
| `etc/httpd/conf/httpd.conf` | `/etc/httpd/conf/httpd.conf` | 同样含 `/prd/`→`/wtopprd/…gwc_azzi000?Arg=99` 的 `Redirect` |
| `u1/genero/as-t35prd.xcf` | `/u1/genero/as-t35prd.xcf` | **`ZONE=t35prd`**、`res.zone=/u1/t35prd`、`base.port 6300 + offset 98`、`GROUP app` |
| `u1/t35prd/erp/cfg/xcf/web/gwc_azzi000.xcf` | 同路径 | **作业定义**：`DVM=$(res.envdir)/webrun`、`MODULE=azzi000`、`GUIMODE=GWC`、`AllowUrlParameters` |
| `u1/t35prd/erp/cfg/xcf/web/gdc_azzi000.xcf` | 同路径 | 桌面版对照：`GUIMODE=GDC`、`Parent=defaultgdc` |
| `u1/t35tst/erp/cfg/xcf/web/gwc_azzi000.xcf` | 同路径 | 跨区对照（已验证与 t35prd **逐字节相同**） |
| `u1/usr/login.sql` | `/u1/usr/login.sql` | sqlplus 登录脚本（`SQLPATH=/u1/usr`） |
| `u1/etc/oraq` | `/u1/etc/oraq` | sqlplus 包装脚本 |
| `u1/etc/q_opid` | `/u1/etc/q_opid` | 查 OPID 的小工具 |

## T100 程序（环境内源码）

| 文件 | 服务器路径 | 作用 |
|---|---|---|
| `u1/t35tst/com/cfg/top_global.inc` | `$TOP/com/cfg/` | **`SCHEMA ds` + 全部全局变量**（`g_enterprise`/`g_dbs`/`g_clienttype`…） |
| `u1/t35tst/erp/azz/4gl/azzi000.4gl` | `$TOP/erp/azz/4gl/` | **T100 系统首頁/登录作业**（`SCHEMA ds`，v1.3.01） |
| `u1/t35tst/erp/azz/4gl/azzi000_03.4gl` | 同上 | 登录/帐号权限子程序，用 `g_enterprise` 过滤 `gzxa_t`/`gzba_t`… |
| `u1/t35tst/erp/azz/42r/azzp000.42r` | `$TOP/erp/azz/42r/` | **作业加载器（只有编译件）** |
| `u1/t35tst/erp/azz/4gl/azzp820.4gl` | `$TOP/erp/azz/4gl/` | **连接切换范本**：`cl_db_connect(gzou003)` / `cl_db_disconnect_current()` / `cl_db_setconnect()` |
| `u1/t35tst/erp/azz/4gl/azzq080.4gl` | 同上 | 多库处理（`cl_db_connect("dscom")`、按 `gzou003` 切换） |
| `u1/t35tst/erp/azz/4gl/azzi085.4gl` | 同上 | **维护 `gzda_t`**（ERP 库 → 中间库映射），并应重新生成 `multipoijdbc.ini` |
| `u1/t35tst/com/lib/4gl/cl_eai.4gl` | `$TOP/com/lib/4gl/` | **`cl_eai_db_connect()`**：`gzda_t` 查中间库 + 兜底 dsdemo→dsawst |
| `u1/t35tst/com/lib/4gl/cl_get.4gl` | 同上 | **跨 schema 读**：`g_dbs="ds"` 时用 `gzou_t.gzou003` 拼 `schema.table` |
| `u1/t35tst/com/lib/4gl/cl_gr.4gl` | 同上 | 注释直白解释「AZZ/ADZ 跑在 ds 时抓不到 dsdemo 数据」 |
| `u1/t35tst/com/lib/4gl/cl_xg.4gl` | 同上 | `SET CONNECTION g_dbs` / `cl_db_connect(g_report,…)` |
| `u1/t35tst/com/lib/4gl/cl_cmdask.4gl` | 同上 | 命令/查询里用 `gzou003` 定位数据库 |
| `u1/t35tst/erp/adz/4gl/adzi140.4gl` | `$TOP/erp/adz/4gl/` | 表设计器：`cl_db_connect("ds",TRUE)`、`TOPIND` 由 `INDSD` 覆盖并 `FGL_SETENV` |
| `u1/t35tst/erp/adz/4gl/adzi888_global.inc` | 同上 | `g_sys_schema2 = "dsdemo"`（工具的双重 schema 常量） |

## 启动器 / 工具

| 文件 | 服务器路径 | 作用 |
|---|---|---|
| `u1/t35tst/utl/bin/r.r` | `$TOP/utl/bin/` | **运行作业**：`cd $TEMPDIR; fglrun $AZZ/42r/azzp000.42r "$@"` |
| `u1/t35tst/utl/bin/r.l` | 同上 | 链接 42r（`TOPENT=99` 在此处显式设） |
| `u1/t35tst/utl/bin/r.c` | 同上 | 编译（前置 `fglprofile.dev` + dev 授权；注意 `export FGLPROFIE` 拼写错误） |
| `u1/t35tst/utl/bin/r.cs` | 同上 | 编译（字符串/42s 相关） |
| `u1/t35tst/utl/bin/r.s` | 同上 | 生成 schema（`ds.sch`/`ds.42d`），前置 dev profile |
| `u1/t35tst/utl/bin/r.f` | 同上 | 编译表单 |
| `u1/t35tst/utl/bin/r.d` | 同上 | 差异/部署工具 |
| `u1/t35tst/utl/bin/r.h` | 同上 | 帮助/说明 |
| `u1/t35tst/utl/bin/r.patch` | 同上 | 补丁流程 |
| `u1/t35tst/utl/bin/r.ent` | 同上 | **设 `TOPENTERPRISE`**（与 `TOPENT` 不同！且 `ans="$1"` 覆盖了 `read`，实际不可交互） |
| `u1/t35tst/utl/bin/appq` | 同上 | 查询工具入口 |

## 说明

- 未纳入的：`$TOP` 下的 `42m/`（编译中间物）、`www/`（设计器/文档，约 3 GB）、
  `utl/java/jar/*.jar`、`apse/`（Java 排程）。
- `azzp000.4gl` 与 `cl_db*.4gl` **在服务器上就不存在**（只有 `.42r`/`.42m`），
  所以无法纳入——见 [05](05-实证记录.md#四未解--只能推断的部分)。
