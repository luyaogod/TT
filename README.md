# TT （T100 开发CLI）

TT 是用于开发鼎捷数智旗下的大型ERP系统（T100）所构建的CLI工具集

---

## 安装

参考 https://github.com/luyaogod/TT/releases

- 便携包
    ```powershell
    #手动将tt安装到用户PATH
    tt install path
    ```
- MSI安装包

---

## 升级

```powershell
tt update check     # 查有没有新版（不下载不安装）
tt update           # 查 + 下载校验，问一句再装
tt update log       # 上一次升级走到哪一步
```

`tt update` 按你现在的安装形态装：便携包就地覆盖程序目录（`config.json` 与目录里
你自己的文件不碰），MSI 安装交给 `msiexec`（用户级安装，不需要管理员）。升级全程
需要能访问 GitHub；内网要经代理时用 `tt update --proxy http://host:port`，
或把地址写进 `config.json` 的 `net.proxy`。

升级会把 agent 目录里的技能树一并刷成同版（`tt install skills` 装过哪些地方记在
数据目录的 `.tt-skills.json`）。如果那一步没成，`tt version` 会提示你手动跑
`tt install skills --force`。

MCP/Agent 环境里注意：默认会问一句再装，脚本里要用 `tt update --yes`；
先看它打算做什么可以 `tt update --dry-run`。

---

## 初次配置

```
tt serve
```

> 启动Serve后在本地打开配置网页` http://127.0.0.1:28670`
>
> 至少需要配置一个T100远程Linux环境

---

## 命令清单

全部命令都在这里（`tt <命令> --help` 看参数与示例，`tt help <命令>` 等价）。每条后面的
一句话就是那条命令自己的提要 —— 细节、数字、动词数以它的 `--help` 为准。
`tdebug` / `tdev` / `tdict` 这三个名字也可以叫（分别等价于 `debug` / `dev` / `dict`）；
`tt env use` 与 `tt env topent` 只打印迁移提示，不做任何事。

全局开关（所有命令都认）：`--env <环境名>`（环境一律点名，没有默认环境可回落）、
`--config <路径>`、`--format json|table|csv`（`--json` / `--csv` 是语法糖）、`-v`。
各条线的退出码含义不同，见 `tt --help` 末尾那张表。
shell 补全：`tt completion bash|zsh|fish|powershell`。

### `tt debug` — TDebug - T100 作业调试器

- `tt debug start` — 连接 SSH 并对指定程序启动调试，等到入口停站
- `tt debug exec` — 透传 fgldb 标准调试命令(print/break/next/where/info/...),可一次多条
- `tt debug wait` — 阻塞等待会话事件(停站/退出/掉线/看门狗),或等到超时
- `tt debug status` — 查看调试服务状态与活动会话
- `tt debug stop` — 查看当前调试会话的停站现场(state/位置/断点/TOPENT,可复取)
- `tt debug quit` — 结束指定调试会话(作业窗口随之关闭)
- `tt debug interrupt` — 中断运行中的程序(回到调试器,等效 SIGINT)
- `tt debug why` — 探测程序此刻在干什么:在等用户操作,还是在空转(慢查询/慢循环)
- `tt debug logs` — 查看会话最近事件(停站/状态/日志/断开;tail 查询)
- `tt debug source` — 读取服务器源码(登录区源码目录白名单只读,可指定行段)
- `tt debug locate` — 定位函数定义到文件:行(需停站;info line)
- `tt debug resolve` — 解析作业编号对应的实体程序/模块(gzzz_t;不启动会话)
- `tt debug db` — 数据库连接探查:按企业(TOPENT)查账号并尝试连接
- `tt debug ents` — 列出当前环境的企业编号(ENT)→ 数据库账号(schema);带快照,离线可答
- `tt debug sql` — 执行一条只读 SQL(默认 JSON 信封,含环境与 ENT;账号由 TOPENT 决定)
- `tt debug wslogs` — 获取接口日志列表(wssp/awsp 报文流水)
- `tt debug wsdebug` — 按接口日志的报文参数启动重放调试,等到入口停站
- `tt debug bp` — 断点面：list（带命中计数）/ add / del
  - `tt debug bp list` — 列出断点（每条带命中计数 Hits）
  - `tt debug bp add` — 下一条断点（位置：行号 / 函数名 / file:line；`--kind watch` 时是表达式）
  - `tt debug bp del` — 删除一条断点（编号见 bp list）
- `tt debug mode` — 查看或切换会话模式:纯人工(solo) / 协作(collab,AI 主导)
- `tt debug topent` — 查看/设置会话 TOPENT override(需 serve 在运行,会话空闲可设)
- `tt debug env` — 查看/切换当前调试环境(SSH 配置,需 serve 在运行)
- `tt debug probe` — 调试链路自检:全自动验一遍协议驱动器(登录/断点/命中/源码/栈帧/步进/求值)
- `tt debug serve` — 后台常驻调试服务（`tt debug start/exec/status` 等控制命令会自动发现的守护进程）

### `tt session` — 会话：建 / 列 / 看 / 改 / 收（会话是一级对象）

目标一律用 `--session` 点名；多条不替你挑。

- `tt session new` — 建会话（可同时起作业）
- `tt session ls` — 列会话
- `tt session show` — 看一条会话的详情（身份 / 作业 / 停站位置）
- `tt session job` — 在这条会话里起一个作业（会话本来就空闲）
- `tt session exec` — 透传一条 fgldb 命令（等价 `tt debug exec`）
- `tt session ent` — 改会话的企业编号（仅空闲时可设）
- `tt session quit` — 结束本轮调试（会话保留，回空闲）
- `tt session rm` — 收掉会话（断连接、清资源）

### `tt dict` — TDict - ERP 数据字典查询工具

- `tt dict r.t` — 查询数据表字典 (r.t)
- `tt dict r.q` — 查询可复用开窗 (r.q)
- `tt dict r.v` — 查询校验带值定义 (r.v)
- `tt dict scc` — 查询系统分类码 (SCC)
- `tt dict desc` — 查询数据库字段对应的前端规格
- `tt dict prog` — 查询程序/作业字典（程序做什么、被哪些作业使用）
- `tt dict msg` — 查询系统消息（编号/语句/类型/状态/建议作业 多条件）
- `tt dict sysp` — 查询系统参数说明
- `tt dict docp` — 查询单据参数说明
- `tt dict fieldtab` — 字段表：表单有哪些字段、中文名与类型（读本地 `.42f`；字典在线/本地两种通道）
- `tt dict db` — 管理 ERP 数据库连接与数据同步
  - `tt dict db list` — 列出各环境的数据库连接（环境名/类型/地址/账号）
  - `tt dict db ping` — 验证环境的数据库可达（只读 `SELECT version`；支持 viaSsh 隧道）
  - `tt dict db discover` — SSH 自动发现数据库连接要素并保存为连接
  - `tt dict db status` — 查看本地查询数据的覆盖情况（缺什么、各多少行）
  - `tt dict db sync` — 从 ERP 刷新本地查询数据
- `tt dict mirror` — 管理 T100 本地源码镜像（服务器 4gl/4fd/42s/*.inc → 本地目录）
  - `tt dict mirror dir` — 查看或设置本地镜像根目录（`config.json` 顶层 `mirror.dir`）
  - `tt dict mirror path` — 打印该环境镜像目录的绝对路径（AI 直接前往）
  - `tt dict mirror pull` — 下载/更新该环境源码镜像（默认增量；`--full` 全量重建）
- `tt dict bdldoc` — BDL(4GL) 语言文档目录
  - `tt dict bdldoc dir` — 查看或设置文档目录（`config.json` 顶层 `bdldoc.dir`）
- `tt dict spill` — 查看被截断查询落盘的完整结果（截断时信封里的 `localPath`）
  - `tt dict spill list` — 列出落盘过的结果（新 → 旧）
  - `tt dict spill show` — 读一份落盘结果（可翻页）

### `tt dev` — T100 设计器包工具（.tzc 代码包 + .tzs 表单包）

- `tt dev tzc` — .tzc 代码包：export 渲染围栏工作区，apply 三道闸门写回
  - `tt dev tzc export` — 渲染带围栏的 4GL 工作区（只读包）
  - `tt dev tzc status` — 报告工作区相对上次 export/apply 的改动（不写盘）
  - `tt dev tzc verify` — 不写盘跑完整验证管线（gate1 + gate2）
  - `tt dev tzc apply` — 验证管线全阶段 + 原子写包 + git commit（唯一会改 .tzc 的命令）
  - `tt dev tzc unlock` — 框架解锁：单向状态迁移，只改 workspace
  - `tt dev tzc rename` — 结构事务：改自订定义点的函数名（只改 workspace）
  - `tt dev tzc newfn` — 新增自订定义点（只改 workspace）
  - `tt dev tzc selftest` — 内置 31 项对抗用例（全部合成包）
- `tt dev tzs` — .tzs 表单包：export 纯解压只读；读写表单走引擎动词
  - `tt dev tzs export <包.tzs>` — 纯解压（不做任何处理）
  - `tt dev tzs formcheck` — 静态比对：应然（`.42f`）↔ 实然（生成物里的 `.4fd`），不开引擎
  - `tt dev tzs <动词> --args '<JSON 对象>'` — 读写表单的**唯一**路径，由设计器自己的引擎算；
    动词面由引擎的函数表生成（`tt dev tzs --help` 列全），参数只用 JSON 给

### `tt form` — 出厂表单：列出 / 查看字段表 / 与语料核对 / 预热缓存

- `tt form ls` — 列出该模块下的表单名
- `tt form show` — 查看一张表单的字段表（工号 / 编号 / 控件 / 几何 / 必填）
- `tt form verify` — 与语料核对：离线字段表 vs 归档语料表（逐条，必须 100%）
- `tt form fetch` — 批量把该模块的表单取进缓存（只缓原始 XML，绝不写服务器）
- `tt form cache` — 表单缓存概览：多少张、多少字节、最旧/最新（只读，不删任何东西）

### `tt drawio` — 用 T100 组件库画原型图（.drawio）

- `tt drawio compose` — 把排版规格 spec 展开成 .drawio
- `tt drawio lib` — 把形状库导出成 drawio 能加载的文件

### `tt env` — 管理 SSH 环境（三个工具共用同一份环境清单）

- `tt env list` — 列出全部环境（同不带参数）
- `tt env show` — 显示某个环境的连接详情（口令打码）

### `tt serve` — 启动本地 Web 服务（工作站 + 统一设置页）

默认后台常驻（单实例）：打印地址后立即返回。端口取配置 `listen`（缺省 `127.0.0.1:28670`），
被占用时自动向后顺延。`--foreground` 前台运行、`--stop` 停止后台实例。

### `tt install` — 把 tt 装进你的环境（skills / path）

- `tt install skills` — 把 exe 同目录的 `skills/` 复制到目标目录
- `tt install path` — 把 exe 所在目录加入用户 PATH（HKCU，不需要管理员）

### `tt config` — 配置管理：位置 / 查看 / 读写 / 校验

- `tt config path` — 打印配置文件的解析结果
- `tt config show` — 打印配置内容（口令默认打码，`--raw` 原样输出）
- `tt config get` — 按点分路径读一个值（如 `hosts.sshs`、`debug.termWidth`）
- `tt config set` — 按点分路径写一个值（值按 JSON 解析，解析不了当字符串）
- `tt config validate` — 校验配置：能否解析、环境是否完整、缺省值是否被补全

### `tt cache` — 查看/清理缓存（企业快照、源码镜像、查询落盘、断点存档）

- `tt cache clear` — 清理缓存（默认全清；`--older-than` 只清旧的）

### `tt update` — 更新 tt 自身（查新版、下载校验、交棒给更新器）

- `tt update` — 查 + 下载校验 + 问一句再装（`--yes` 不问、`--dry-run` 只看、`--proxy` 走代理）
- `tt update check` — 查有没有新版（不下载、不安装）
- `tt update log` — 回读最近一次升级的阶段与结果

### `tt version` — 显示版本信息

### `tt completion` — 生成 shell 补全脚本

- `tt completion bash|zsh|fish|powershell` — 输出对应 shell 的补全脚本
