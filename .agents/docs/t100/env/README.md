# T100 登录环境链路 / ENT 与数据库账号 —— 实勘报告

- **目标机**：`172.16.1.109:22`（账号 `tiptop`，只读探索）
- **主机**：`top109ind`，Linux 4.18 / RHEL8，Oracle 19c 单实例多 service
- **实勘时间**：2026-11-14
- 本目录同时存放**直接相关的原始代码**（按服务器绝对路径原样镜像在 `u1/`、`u3/`、`etc/` 下），
  清单见 [FILE-INDEX.md](FILE-INDEX.md)，实测数据见 [env-matrix.csv](env-matrix.csv)

> **⚠ 已脱敏**：所有口令值一律替换为 `<REDACTED>`，涉及
> `fglprofile.*` 的 `dbi.database.*.password`、`multipoijdbc.ini` 的 `pass=`、
> `poijdbc.ini` 的口令行，以及文档里的口令样例与 base64 字面量。
> **`<REDACTED>` 不是真值**，不要当配置用。
> 账号名（`ds`/`dsdemo`/`dsdata`/`dsstd`…）、内网地址、主机名保留——它们是本报告的分析对象，
> 且不含凭据。若要一并打码，告诉我。

## 一句话结论

**"选菜单"决定的是 `ZONE`（=Oracle 服务名 = `$TOP` 目录 = `FGLPROFILE` 文件），
"登录后选企业"决定的是 `gzou_t` 里的 schema 名（`dsdemo`/`dsstd`…），
前者改环境变量、后者只改进程内数据库连接——两者不是一回事，容易混淆。**

## 目录

| 文件 | 内容 |
|---|---|
| [01-登录链路.md](01-登录链路.md) | 从 SSH 登录到拿到环境变量的完整调用链，三个同名 `chenv` 的分工 |
| [02-环境变量.md](02-环境变量.md) | 环境变量全清单、六个菜单选项的**实测**差异、两份 `topenv`/`webrun` |
| [03-ENT与数据库账号.md](03-ENT与数据库账号.md) | `azzi000`/`azzp000` 作业、ENT 切换、`gzou_t`、连接切换 API |
| [04-数据库连接信息位置.md](04-数据库连接信息位置.md) | 连接信息到底存在哪 9 个地方，各自的权威度 |
| [05-实证记录.md](05-实证记录.md) | 怎么验的、原始输出、未解之谜 |
| **[06-Web入口wt35prd.md](06-Web入口wt35prd.md)** | **`wt35prd` 之谜**：URL 前缀 → Apache → fastcgi 端口 → GAS 的 xcf → ZONE |
| [FILE-INDEX.md](FILE-INDEX.md) | 68 个原始文件的逐个说明（服务器路径 + 作用） |
| [env-matrix.csv](env-matrix.csv) | 六个菜单选项的环境变量矩阵（机读） |
| **[architecture.html](architecture.html)** | **一页图看懂全链路**（HTML + 内联 SVG，自包含、可打印）<br>同内容的位图：[architecture.png](architecture.png) |

## 服务器上的关键坐标

```
$TOP          = /u1/<ZONE>          # 例 /u1/t35tst
$ERP          = $TOP/erp
$COM          = $TOP/com
登录脚本       = /u3/pub/bin/chenv  # 被 ~/.profile 符号链接
env 引擎       = /u3/pub/etc/topenv
账号口令表     = /u3/pub/etc/fglprofile.<ZONE>
菜单/作业入口  = r.r azzi000  →  azzp000.42r
```

本机能区分出来的 6 个"区"（同一台服务器、同一个库实例、不同 service/schema 组合）：

| 菜单码 | ZONE | 用途 | Oracle service |
|---|---|---|---|
| `31` | `t35dev` | 开发区 | `t35dev` |
| `35` | `t35tst` | 测试区（tt 的 `109tst`） | `t35tst` |
| `36` | `t35prd` | 正式区（tt 的 `109prd`） | `t35prd` |
| `39` | `t35pth` | PATCH 区 | `t35pth` |
| `t` | `topprd` | 出货区（标准版） | `topprd` |
| `36k` | `k35prd` | 人大金仓版 —— **入口已断，选到直接退出** | 无 |

## 三个最容易踩的坑

1. **`36k`（Kingbase）是死路**：`/u3/pub/bin/chenv` 里分发 `topenv` 的 `36k)` 分支被注释掉了，
   只剩 `*) exit`。实测选 `36k` 时子壳**一个环境变量都没拿到就退出**。见 [02](02-环境变量.md#36k-为什么是死路)。
2. **Web 的 `wt35prd` 不含任何库信息**：它只是 Apache 的 `ProxyPass` 前缀，指向一个
   fastcgi 端口（`6300 + offset`，t35prd=98→6398）；该端口上的进程启动时就用
   `-f as-t35prd.xcf` 把 `ZONE` 写死了。**区域是在"服务启动"那一刻绑定的，不在 URL 里。**
   见 [06](06-Web入口wt35prd.md)。
3. **ENT 切换不动环境变量**：`TOPENT` 只是登录前 shell 给的一个**默认企业号**（缺省 99），
   T100 运行期切 ENT 是 `cl_db_connect()` 换物理连接 + 改 `g_enterprise` 全局。见 [03](03-ENT与数据库账号.md)。

## 采集方式与边界

- 全程**只读**：只用了 `ls / cat / grep / find / ps / diff / base64` 与只读 SQL（`tt db sql`，
  只读事务）；远端没有落任何文件（脚本经 SSH stdin 送入 `bash -s`，不写磁盘）。
- 唯一"执行"的动作是 `. /u3/pub/bin/chenv <选项>`，用于抓取登录后真实环境；
  它自己只做赋值、`cd $TOP` 和一次 `fglWrt -a info user`（查授权人数，无副作用）。
- 编译件（`azzp000.42r`、`lib_cl_db.42m`）**源码不在环境里**，相关结论标为"推断"并给出依据。
