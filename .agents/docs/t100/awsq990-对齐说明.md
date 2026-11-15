# awsq990（整合服務端檢測工具）对齐说明

> 本文件是 TT「服务信息查询 / 集成服务检测」的调研底稿与口径来源。
> 蓝本：T100 `erp/aws/awsq990`，本地镜像 `D:\APPS\TT Desktop\mirror\109prd`。
> 所有行号指 `erp/aws/4gl/awsq990.4gl`，另有注明的除外。

## 1. 页面与页签

| Page | 页签 | 可见 | 内容 |
|---|---|---|---|
| `page_qbe` | 服务信息查询 | 是 | QBE（4 行可见）+ 列表 `s_detail1` |
| `page_1` | 集成服务检测 | 是 | 接口方式 / 集成方式 / 测试 / 结果表 / Request / Response |
| `page_2` | 集成常用指令 | 否（`.42f` `hidden="1"`） | `l_cmdcontent` + `s_detail3` |
| `page_3` | 相关设置 | 否 | AWS Log / Run Log / GAS Log / 设置 IP |
| `page_4` | 集成接口测试 | 否 | WebService / JavaAPI / Restful 三个空壳按钮 |

TT 只复刻前两个页签。

## 2. QBE 栅格（`.42f` 的 posY/posX）

| 行 | 列 1 | 列 2 | 按钮列（posX=80） |
|---|---|---|---|
| 0 | 区域 `l_location`（不进查询） | `wsfa002`（静态 hidden） | 查询 |
| 1 | 发起端 `l_originator` | 服务端 `l_server` | 消息内容 |
| 2 | 起始时间 `wsfa003` | 服务名称 `wsfa001`（ButtonEdit） | 执行 debug |
| 3 | 结束时间 `wsfa004` | 处理结果 `l_result`（SCC 239） | 执行性能测试 |
| 4 | 关键字 `l_datakey`（静态 hidden） | 报表种类 `l_report`（运行时隐藏） | 注意提示 |

TT 映射：区域→企业编号；`wsfa002` 格留空；列 2 从第 2 行起 服务端/服务名称/处理结果；查询放第 3 列第 1 行。

## 3. 查询口径

主查询（1260-1262）：

```sql
SELECT wsfa001,wsfa005,wsfa006,wsfa004,wsfa002,wsfa003,
       wsfa012,wsfa013,wsfa014,wsfa015,wsfa018
FROM wsfa_t
WHERE 1=1 AND wsfa001 != 'docno.storage' <g_new_wc>
ORDER BY wsfa003 DESC, wsfa004 DESC
```

`awsq990_afterField(1)`（2364-2514）生成：

| QBE | SQL |
|---|---|
| 起始时间 | `wsfa003 >= '起'` |
| 结束时间 | `wsfa004 <= '止 23:59:59.99999'` |
| 服务名称 | `wsfa001 = '值'`；多选 `wsfa001 in (...)` |
| 处理结果 | `wsfa006 = '值'` |
| 发起端 | `wsfa013 = '值'` |
| 服务端 | `wsfa018 = '值'` |
| 关键字/报表/区域 | 不产生条件 |

- 清空日期 = 不限时间，不是回退今天。
- `docno.storage` 是 SSO 记录（`wsfa003` 不是时间），所有入口都必须排除；原生 `b_fill` 漏了，TT 统一加。
- 原生 `afterField` 不认 `like`，模糊条件会被丢掉；`b_fill` 用 FORMONLY 名拼 SQL，翻页可能 ORA-00904。TT 只翻译一次成真实列名，不复刻这两个缺陷。

处理结果 SCC 239：`000` 执行正常 / `200` 执行后有错误信息 / `100` Service 无法执行。

## 4. 服务名称开窗 q_gzja005_1

- 触发：486-495，`g_qryparam.state='c'`（多选），`CALL q_gzja005_1()`，`DISPLAY return3 TO wsfa001`。
- `return1=gzja001`（服务规格编号）、`return2=gzjal003`（服务名称）、`return3=gzja004`（**API服务名称**）。
- 多选以 `|` 串联（`q_gzja005_1.4gl:894-910`）。
- SQL（`q_gzja005_1.4gl:586-651`）：

```sql
SELECT gzja001, gzjal003, gzja004 FROM gzja_t
LEFT JOIN gzjal_t ON gzja001=gzjal001 AND gzjal002=:g_lang
WHERE 1=1
UNION ALL
SELECT gzjc002, gzjcl003, gzjc001 FROM gzjc_t
LEFT JOIN gzjcl_t ON gzjc001=gzjcl001 AND gzjcl002=:g_lang
WHERE 1=1
```

- 本环境 `gzja_t` 2129 行、`gzjc_t` 0 行；每页 500、选择上限 1000。
- 三列表头即三个过滤字段：`gzja001/gzjal003/gzja004`，`AFTER FIELD` 调 `cl_ap_code_fuzzyquery` 自动补 `*`。

## 5. 集成服务检测

- `wsfc001` 是 `.42f` 静态下拉：`1 Web Service`、`3 Restful`、`4 OpenApi Restful`、`5 OpenApi Web Service`；默认 `3`；**不接 SCC 240**。
- URL（522-531）：

```
ZONE=topprd：http://<APIP>/ws<ZONE>/ws/r/awsp920
其余区      ：http://<APIP>/w<ZONE>/ws/r/awsp920
```

- 方法映射：`1→awsp900`、`3→awsp920`、`4→awsp940`、`5→awsp930`。
- 测试（1014-1072）：
  - `1/5`：`< >` 转义后包 SOAP 信封（`soapenv:Envelope → tip:invokeSrv → request`），`SOAPAction: ""`，POST。
  - `3/4`：原文 POST（UTF-8，原生未设 Content-Type）。
  - 仅 HTTP 200 判成功；非 200 显示 `HTTPS Error (<code>) <reason>` 并丢弃 body；200 的 SOAP 响应再反转义并截 `<response>…</response>`（2218-2238）。
- 结果小表（2329-2348）：`起始时间 / HTTP code / 运行结果 / 处理时间(秒)`，新行插到第 1 行，关窗即消失。
- **测试不写 `wsfa_t`**；日志由 `awsp900_01_insertInto`、`s_awsp306_ins_log_into_wsfa`、`cl_openapi_ins_wsfa` 在真实调用时写。
- 原生缺陷：异常时 `l_resp` 可能为 NULL 仍取 `getStatusCode()`；非 200 丢 body，看不到 Proxy error 细节。TT 保留 body，并在传输失败时给明确文案。

## 6. 全局变量来源

`$TOP/cfg/top_global.inc`（本地副本 `.agents/docs/t100/r.t/deps/top_global.inc`）：

| 变量 | 行 | 用途 |
|---|---|---|
| `g_argv` | 30 | `default_search:335-344`：argv[1..3] → `wsfa001/wsfa002/wsfa003` 预置条件 |
| `g_enterprise` | 34 | 登录企业；`awsq990` 不直接引用，经 `cl_get_para` 读参数（如 A-SYS-0093） |
| `g_lang` | 28 | `gzjal_t/gzjcl_t` 语言别 join |
| `g_qryparam.state` | 142 | `c` = CONSTRUCT（多选开窗） |
| `g_qryparam.return1..3` | 166-168 | 开窗回传；`return3 = gzja004` |

- 没有 `g_zone`；区域读 `FGL_GETENV('ZONE')`。
- `wsfa_t` 是 `DS.WSFA_T`（`DSDEMO.WSFA_T` 同义词），18 字段**没有 ent 列**；T100 的「固定 ent」不体现在日志查询条件上。TT 的企业编号只作运行语境（重放 TOPENT、结果回显）。
- `APIP` = `cl_getenv_tpserver_ip()`，空则 `127.0.0.1`；TT 默认用配置的 SSH host。

## 7. 对 TT 的结论

1. 页面只做两个页签；条件区四行三列；服务名称开窗回传 `gzja004`，多选 `|`。
2. 查询只用一个 WHERE 构造器，真实列名 + 所有入口都排除 `docno.storage`。
3. 检测从 T100 服务器发 `curl`（与原生同网络侧）；URL 由后端按区域生成、可编辑；请求体走 stdin；不落临时文件。
4. 检测不写 `wsfa_t`；真实调用让框架自己写日志。
5. 非 200 保留 body；SOAP 截取失败回退整段。
