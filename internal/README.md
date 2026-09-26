# internal — Go 侧的全部能力

按**四层**组织：入口与命令层、共享能力层、能力层，加一块横向支撑。上层依赖下层，
**没有反向边**。

## 分层

| 层 | 目录 | 关系 |
|---|---|---|
| 入口与命令层 | 根、`cli/`、`cli/{debug,dev,dict}/` | 三个能力组并列且互不依赖；统一命令组也在这一层 |
| 共享能力层 | `config/`、`host/`、`output/`、`dbconfig/`、`erpdb/`、`entdir/`、`safesql/`、`sshtun/`、`winproc/`、`pathinstall/` | 被上下两层共同依赖，**自身不反向依赖业务包** |
| 能力层 | `debug/`、`web/`、`dev/`、`dict/` | 各自独立，之间只通过注入相连 |
| 横向支撑 | `dev/tzs/` | 引擎客户端，依赖面刻意只有 `config` + `winproc` |

## 三条规矩

1. **三个命令组互不 import**（`cli/debug`、`cli/dev`、`cli/dict` 之间没有任何依赖边）。
   共享能力一律下沉到共享层。
2. **往上通信靠注入，不让叶子反向 import**。共享的 CLI 上下文住在 `cli/common`（叶子包），
   需要命令组提供的东西以函数变量注入。
3. **命令层只有一个汇合点**：统一服务、调试内核、字典动作三者的接线只发生在 `cli/` 里。

**没有任何内部依赖的叶子包**（可独立测试、可替换的基座）：
`dbconfig`、`output`、`pathinstall`、`safesql`、`sshtun`、`winproc`、`dict/db`、
`dev/model`、`dev/fgl`、`dev/tapfile`、`dev/tglfile`、`dev/testutil`。

## 依赖形状

```
命令层 ─┬─ cli/common ── config, output
        ├─ cli/debug ── config, dbconfig, debug, output, winproc
        ├─ cli/dev   ── dev/cli
        ├─ cli/dict  ── config, dbconfig, dict/{db,dbsync,live}, entdir, host, output
        ├─ debug ───── config, dbconfig, entdir, host, safesql
        ├─ dict/server ─ config, dict/dbsync, host
        └─ web ─────── config, dbconfig, erpdb, host, pathinstall
```

**`web` 与 `debug` 之间没有任何 import 边** —— 耦合全靠统一服务在装配处注入处理器。

## 谁管什么

| 东西 | 谁管 | 存在哪 |
|---|---|---|
| SSH 环境（主机/用户/口令/区域/企业编号） | `host` + 配置的 `hosts` 节 | `config.json` |
| 数据库连接（类型/地址/账号清单） | `dbconfig` | `config.json` |
| **用哪个账号** | 解析出来的，不是配的（`entdir`） | 运行时快照 |
| 查询走哪个数据源 | `cli/dict` 的钩子 | `config.json` + `--env` |
| 本地字典副本 | `dict/dbsync` | 由同步目标决定 |
| 输出形状 | `output`（唯一出口） | —— |
| 缓存目录清单 | `config`（唯一定义） | —— |

## 四条硬规矩

1. **命令组之间互不 import。**
2. **SSH 主机密钥当前不校验**（与"手工登录内网机器"同一信任级别）。要改这条等于改安全模型，
   须单独讨论，不要顺手改。
3. **口令一律不进命令行**（客户端直连路径）。服务器侧金仓是唯一例外，代码注释里如实标注。
4. **往配置目录写文件时，要么登记进缓存清单，要么说清它为什么不是缓存** ——
   那个目录里住着唯一删了不能自愈的文件。

## 加新东西时的自查

- **环境从哪来？** 一律走配置层的解析入口，**不要自己写"名字 → 活动环境 → 首条"**。
- **账号从哪来？** 一律走企业目录解析，**不要自己取账号列表第一项**。
- **输出怎么走？** 一律走统一输出出口，**不要自建 JSON/CSV 发射器**。
- **写配置？** 一律走唯一写入口。
- **往配置目录写文件？** 登记进缓存清单，或说明为什么不是缓存。
- **拼远端命令？** 值必须过白名单或引号；**SQL 一律走标准输入**，不拼进命令串。
- **新增位置参数？** 负整数一类的值会被开关解析吃掉，要给出可操作的提示。

## 判据

```bash
go test ./...                 # 35 个包，其中 22 个带测试
go build -o tt.exe .
```

## 细节去哪

- 各层内部的细节 → 各目录的 README（见上表的目录名）
- 整体结构与关系 → 根 [DESIGN_DOC.md](../DESIGN_DOC.md)
- 怎么改、改哪块先读什么 → 根 [AGENTS.md](../AGENTS.md)
