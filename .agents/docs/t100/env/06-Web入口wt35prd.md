# 06 · Web 登录：`wt35prd` 是怎么变成 `t35prd` 的

> 这是整条链里唯一的"黑盒"。结论先给：
> **`wt35prd` 这个路径段里不含任何数据库/账号信息。它只是 Apache 的一个前缀，
> 被 `ProxyPass` 硬映射到一个 fastcgi 端口；那个端口上跑的进程，在**启动时**就用
> `-f as-t35prd.xcf` 把 ZONE 写死了。链路在请求到达之前就已经定好。**

## 一、URL 拆开看

```
http://172.16.1.109/wt35prd/da/r/app/gwc_azzi000?Arg=99
                    └──┬──┘ └┬┘ │ └──┬───────┘ └──┬──┘
                       │     │  │    │            └ 程序参数（ENT = 99）
                       │     │  │    └ 应用 id：GROUP "app" + gwc_azzi000.xcf
                       │     │  └ 语言段（r = 默认）
                       │     └ 接口模式段（da=desktop / wa=web，见 §五）
                       └ ★ 唯一的路由开关
```

## 二、六段链路（每一段都已实地核对）

| # | 环节 | 事实 | 证据 |
|---|---|---|---|
| 1 | Apache 前缀匹配 | `ProxyPass /wt35prd/ fcgi://172.16.1.109:6398/ timeout=300` | `/etc/httpd/conf/fastcgi.conf` |
| 2 | fastcgi 端口 | `6398` 上监听的是 `fastcgidispatch` | `ss -lntp` |
| 3 | 该进程用哪份配置 | `fastcgidispatch -s -f /u1/genero/as-t35prd.xcf` | `ps -ef` / `/proc/<pid>/cmdline` |
| 4 | xcf 里写死区域 | `<ENVIRONMENT_VARIABLE Id="ZONE">t35prd</ENVIRONMENT_VARIABLE>` | `as-t35prd.xcf:155` |
| 5 | 启动器 | `gwc_azzi000.xcf` 的 `<DVM>$(res.envdir)/webrun</DVM>`，`res.envdir=/u1/etc` | `gwc_azzi000.xcf`、`as-t35prd.xcf:134` |
| 6 | webrun 读环境 | `. /u3/pub/etc/chenv` + `. /u3/pub/etc/topenv`（`ZONE` 已由 GAS 注入） | `/u1/etc/webrun:22-23` |

于是：

```
ZONE=t35prd
  ├─ TOP        = /u1/t35prd
  ├─ ORACLE_SID = TWO_TASK = t35prd
  ├─ FGLPROFILE = /u3/pub/etc/fglprofile.t35prd
  └─ 程序路径    = /u1/t35prd/erp/azz/42r/azzi000
```

`FGLPROFILE` 里 `dbi.database.<别名>.source = "t35prd"` →
`tnsnames.ora` 的 `T35PRD` → **`top109ind:1521/t35prd`**。
账号仍由 `gzou_t` 按 ENT 决定（见 [03](03-ENT与数据库账号.md)）。

## 三、端口不是随机的

`as-t35prd.xcf` 里两个资源相加就是端口：

```xml
<RESOURCE Id="res.ic.base.port"   Source="INTERNAL">6300</RESOURCE>
<RESOURCE Id="res.ic.port.offset" Source="INTERNAL">98</RESOURCE>
...
<TCP_BASE_PORT>$(res.ic.base.port)</TCP_BASE_PORT>
<TCP_PORT_OFFSET>$(res.ic.port.offset)</TCP_PORT_OFFSET>
```

**6300 + 98 = 6398**。每个区一个 offset：

| 区 | offset | fcgi 端口 | 监听进程用的是 |
|---|---:|---:|---|
| `topprd` | 94 | 6394 | `as-topprd.xcf` |
| `t35tst` | 95 | 6395 | `as-t35tst.xcf` |
| `t35dev` | 96 | 6396 | `as-t35dev.xcf` |
| `t35pth` | 97 | 6397 | `as-t35pth.xcf` |
| **`t35prd`** | **98** | **6398** | **`as-t35prd.xcf`** |

（`-ws` 变体是另一组端口，如 `wstopprd → 6370`，由 `as-*-ws.xcf` 提供 Web Service。）

> 所以 `wt35prd` 本质上只是"6398 号门"的门牌。**把 `wt35prd` 指向 6395，
> 你就会拿到 t35tst 的程序去连 t35prd 的库**——路径名本身没有任何校验作用。

## 四、作业定义在哪：`$TOP/erp/cfg/xcf/web/`

`as-t35prd.xcf` 里把 GROUP `app` 指向了这里：

```xml
<GROUP Id="app">$(res.zone.config)/xcf/web</GROUP>
<!-- res.zone.config = $(res.zone.erp)/cfg = /u1/t35prd/erp/cfg -->
```

于是 `app/gwc_azzi000` → **`/u1/t35prd/erp/cfg/xcf/web/gwc_azzi000.xcf`**（463 B，全文）：

```xml
<APPLICATION Parent="defaultgwc">
  <EXECUTION AllowUrlParameters="TRUE">          <!-- ← 这句让 ?Arg=99 能传进去 -->
    <ENVIRONMENT_VARIABLE Id="GUIMODE">GWC</ENVIRONMENT_VARIABLE>
    <PATH>$(res.zone.erp)/azz/42r</PATH>          <!-- /u1/t35prd/erp/azz/42r -->
    <DVM>$(res.envdir)/webrun</DVM>               <!-- /u1/etc/webrun -->
    <MODULE>azzi000</MODULE>
  </EXECUTION>
</APPLICATION>
```

目录里有 **10675 个** `gwc_*.xcf` / `gdc_*.xcf`——每个作业两份（Web 版 / 桌面版）。
两份唯一的差别是 `GUIMODE` 与 `Parent`：

| | `gwc_azzi000.xcf` | `gdc_azzi000.xcf` |
|---|---|---|
| `Parent` | `defaultgwc` | `defaultgdc` |
| `GUIMODE` | **`GWC`** | **`GDC`** |
| `DVM` / `MODULE` / `PATH` | 完全相同 | 完全相同 |

**这些文件在各区之间逐字节相同**（已 diff 验证）：作业定义不含区域信息，
区域只存在于 `as-<ZONE>.xcf` 的 `res.zone` 与 `ZONE`。这是这套设计的关键分层。

## 五、`da` 与 `wa` 的区别

| 段 | httpd 的跳转用它 | 实际效果 |
|---|---|---|
| `wa` | `Redirect /tprd/ …/wt35prd/**wa**/r/app/gwc_azzi000?Arg=99` | Web/浏览器 |
| `da` | 你手上的 URL 用 `da` | desktop（GDC 客户端） |

真正决定 GUI 形态的**不是**这个段，而是选中那份 xcf 里的 `GUIMODE`。
你给的是 `/da/r/app/gwc_azzi000`，加载的仍是 `gwc_azzi000.xcf`（`GUIMODE=GWC`），
现场进程里的 `"guimode":"GWC"` 正是这么来的。

## 六、`?Arg=99` 去哪了

`AllowUrlParameters="TRUE"` 让 URL 参数变成程序参数。现场进程的命令行直接印证了这一点：

```
fglrun ... /u1/t35prd/erp/azz/42r/azzi000 \
  {"sessionkey":"NKEP350675266b","parent":"NKEP350395006b","prog":"azzi000",
   "account":"tiptop",
   "param":["99","T100AISSO","V0VCTkFOQQ==","MjAy…"],
   "schedule_seq":0,
   "static_ent":99,                 ← ★ Arg=99 变成了 static_ent
   "ip":"10.1.25.216",
   "guimode":"GWC",
   "frontEndId":"586593b5-…"}
```

- `param[0] = "99"` ← 就是 `?Arg=99`
- `static_ent = 99` ← 它被当作**企业（ENT）**，不是据点
- 据点（`營運據點`，如截图里的「鼎捷南京 DSCNJ」）走登录画面/另一个参数，不在这条链上

所以：**`Arg` = ENT，进 `gzou_t` 查 schema；据点只是运行期变量。**

## 七、两条登录路径的对照

| | 终端 / 2-tier | Web·GDC / 4-tier |
|---|---|---|
| 入口 | `menu` → `r.r azzi000` | URL `/wt35prd/...` |
| 谁定 ZONE | `chenv` 的菜单（人选的） | `as-<ZONE>.xcf`（系统启动时定的） |
| 启动器 | `$TOP/utl/bin/r.r` | `/u1/etc/webrun`（由 `<DVM>` 指定） |
| 传给 `azzp000` 的第一个参数 | **作业名** `azzi000` | **完整路径** `/u1/t35prd/erp/azz/42r/azzi000`（GAS 用 `PATH`+`MODULE` 拼出来，推断） |
| 作业定义 | `topalias` 里的别名 | `$TOP/erp/cfg/xcf/web/gwc_<prog>.xcf` |
| 环境变量 | `/u3/pub/etc/chenv` + `topenv` | **同一份** + 额外 `fglprofile.cpu` / `fgllicense.cpu` |
| GUI | GDC 直连 | GAS 代管，`GUIMODE=GWC` |

两条路最后都收敛到同一个 `azzp000.42r` + 同一份 `topenv`，所以**库、账号、service 的解析规则完全一致**；
不同的只有"谁替你把 ZONE 填好"。

## 八、⚠ 修正：两份 `webrun`，实际用的是 `/u1/etc/webrun`

本目录早先根据 `/u3/pub/etc/webrun` 写过"Web 侧走 `/u1/etc/topenv`"，**这个结论要改**。
两份 `webrun` 确实都存在，但 GAS 只调用 `res.envdir` 指向的那一份：

| | `/u1/etc/webrun`（**在用**，2602 B，2025-07） | `/u3/pub/etc/webrun`（旧，2131 B，2023-08） |
|---|---|---|
| 被谁调用 | `as-<ZONE>.xcf` 的 `<DVM>$(res.envdir)/webrun</DVM>`，`res.envdir=/u1/etc` | 未在 xcf/systemd 中引用（历史遗留） |
| source 哪套 env | **`/u3/pub/etc/chenv` + `/u3/pub/etc/topenv`** | `/u1/etc/chenv` + `/u1/etc/topenv` |
| `T100NANAWEB` | `Y`（影响 azzi000 窗口样式） | 不设 |
| `FGLPROFILE` | **无条件**前置 `$FGLDIR/etc/fglprofile.cpu` | 仅当 `prog != azzi000` 时前置 |
| `FGLLICENSE` | `$FGLDIR/etc/fgllicense.cpu`（CPU 授权） | 不设 |
| `azzi000→adzi444` 改名 | 已注释 | 生效（当 `$TOPKEY/flc2.sh`=1） |
| `CUST` | 先设 `T100DEMO3`，随后被 `chenv` 按 ZONE 覆盖 | `DIGIWIN` |

`fglprofile.cpu` 是 Genero 自带的**全量默认配置**文件，里面没有任何 `dbi.database.*` 别名，
所以前置它不会影响连库；真正的库别名仍来自 `/u3/pub/etc/fglprofile.<ZONE>`。

**结论：Web 与交互登录用的是同一份 `topenv`、同一份 `fglprofile.<ZONE>`。**
两者的环境差异只有三条：`T100NANAWEB=Y`、`FGLPROFILE` 多前置一个 `cpu` 文件、
`FGLLICENSE` 指向 `fgllicense.cpu`。

## 九、回到你的四个问题

| 问题 | 答案 |
|---|---|
| **登录不同区域，怎么判断执行哪份代码** | `wt35prd` → Apache `ProxyPass` → `:6398` → `fastcgidispatch -f as-t35prd.xcf` → xcf 的 `ZONE`/`res.zone`=/u1/t35prd → `TOP=/u1/t35prd`；作业定义在 `$TOP/erp/cfg/xcf/web/gwc_<prog>.xcf` |
| **查哪个数据库** | `ORACLE_SID`/`TWO_TASK`=`$ZONE`；`FGLPROFILE=/u3/pub/etc/fglprofile.<ZONE>` 里的 `source="<ZONE>"` |
| **用哪个账号** | `gzou_t` 按 ENT 查 `gzou003`（`dsdemo`/`dsdata`/`dsstd`），再取 `fglprofile` 里同名别名的 `username`/`password` |
| **连哪个 Oracle service_name** | `tnsnames.ora` 里与 `$ZONE` 同名的别名 → `top109ind:1521/<SERVICE_NAME>` |

一条命令验证某区到底连哪个 service（用本机 tt）：

```powershell
tt db sql --at 109tst@99 --format table "select sys_context('USERENV','SERVICE_NAME') svc, user from dual"
```

## 十、本目录新增的证据文件

| 文件 | 服务器路径 |
|---|---|
| `etc/httpd/conf/fastcgi.conf` | `/etc/httpd/conf/fastcgi.conf`（`ProxyPass` ↔ 端口） |
| `etc/httpd/conf/httpd.conf` | `/etc/httpd/conf/httpd.conf`（`/prd/`、`/tst/`… 短链跳转） |
| `u1/etc/webrun` | `/u1/etc/webrun`（**GAS 实际调用的启动器**） |
| `u1/genero/as-t35prd.xcf` | `/u1/genero/as-t35prd.xcf`（`ZONE`、`res.zone`、端口 offset） |
| `u1/genero/as-t35tst.xcf` | `/u1/genero/as-t35tst.xcf`（对照：offset 95） |
| `u1/t35prd/erp/cfg/xcf/web/gwc_azzi000.xcf` | 同上（`DVM=webrun`、`MODULE=azzi000`） |
| `u1/t35prd/erp/cfg/xcf/web/gdc_azzi000.xcf` | 同上（对照：`GUIMODE=GDC`） |
| `u1/t35tst/erp/cfg/xcf/web/gwc_azzi000.xcf` | 同上（已验证与 t35prd 逐字节相同） |
