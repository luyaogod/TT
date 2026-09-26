# TT — T100 工具集

T100 日常工作的统一入口：**一个二进制 `tt`、一份配置、一个本地 Web 服务。**

TT 是一款面向 Agent 的 CLI 开发工具，用于开发基于 Genero BDL 技术栈的大型 ERP 系统 ——
鼎捷数智旗下的 T100。它把"改 ERP 客制代码"这件事拆成四块能力，供人或 AI 在同一个入口下使用：

| 命令组 | 做什么 |
|---|---|
| `tt debug` | 作业调试器：经 SSH 驱动服务器上 `fglrun -d` 的调试协议，提供本地 Web 调试界面（源码 / 断点 / 调用栈 / 变量 / 接口日志）与命令行控制端 |
| `tt dev tzc` | 设计器**代码包**（`.tzc`/`.tzf`/`.tzx`）：渲染成带围栏的 4GL 工作区给人或 AI 修改，改完经三道闸门写回 |
| `tt dev tzs` | 设计器**表单包**（`.tzs`/`.tzv`）：以具名动词读写表单，由设计器自己的引擎计算，不是拼 XML |
| `tt dict` | ERP 数据字典查询：表 / 字段 / 校验 / 分类码 / 开窗 / 消息 / 参数 / 程序，支持本地镜像与远程直查 |

另有 `tt env` / `tt config` / `tt serve` / `tt install` / `tt version` 负责环境、配置、本地服务与安装。

**tt 不实现 T100 设计器的任何私有格式** —— `.tzc` 的行为依据设计器的公开发行物反推，`.tzs` 直接
调用设计器自己的程序集。设计器的程序集随发行包一起分发，所以**不需要另外安装设计器**，也不需要
配置它的路径。

## 安装

### 便携包

把 `tt-portable.zip` 解压到任意目录，双击或命令行运行 `tt.exe` 即可。包内带便携标记，配置就近
留在包内（`config.json`），不写用户目录。包内另有 `tzs\`（引擎与设计器程序集）与 `skills\`。

### MSI 安装包

`TT-0.1.0-x64.msi` 双击安装，**全程不需要管理员**：装到 `%LOCALAPPDATA%\Programs\TT`，
安装目录追加到用户 PATH，卸载时自动摘掉。

### 把 tt 加进 PATH

```
tt install path             # 把 tt.exe 所在目录加进用户 PATH
tt install path --dry-run   # 只预览将要写入的内容
```

### AI 技能

```
tt install skills                       # 复制到 <当前目录>/skills
tt install skills --to .claude/skills   # 装到 Claude Code 直接读的位置
```

`skills/` 下是五套：`tt-debug`、`tt-dev-tzc`、`tt-dev-tzs`、`tt-dict`、`erp-read`。装一次全部到位。
**这五份也是给人看的操作手册** —— 每个命令组怎么用、有哪些坑，都在里面。

## 首次配置

### 1. 起服务，加一台环境

```
tt serve
```

浏览器打开它打印的地址，在「设置 → 站点管理」里加一台环境，填 SSH（主机 / 端口 / 用户 / 口令 /
区域 / 企业编号）与数据库（类型 / 地址 / 库名 / 账号）两部分。界面上可以就地验证：账号行有验证
按钮，数据库有「测试连接」。

加完确认一下：

```
tt env list          # 环境清单，星号是当前环境
tt env use <名称>    # 切换当前环境
```

### 2. 配 `.tzs` 工作区（必须）

表单包读写的引擎需要一个工作区目录，**这一项没有缺省值**（引擎内置的默认值是一个真实客户目录）：

```
tt config set tzs.workspace "D:\你的工作区"
tt dev tzs doctor        # 自检：引擎 exe / 设计器 / 工作区 / 管道名都应该通过
```

### 3. 按需配置其余项

以下都在设置页里，或 `tt config set <键> <值>`：

| 项 | 键 | 说明 |
|---|---|---|
| 查询数据源 | `query.source` | `auto`（在线优先）/ `local` / 某个环境名 |
| 本地字典副本 | `sync.target` | 离线查询用的本地库文件位置 |
| 源码镜像目录 | `mirror.dir` | 拉取 ERP 源码的本地目录 |
| BDL 文档目录 | `bdldoc.dir` | 查 BDL 文档时的本地目录 |

### 4. 试一下

```
tt debug start <作业> -m <模块>              # 调程序 → 细节见 skills/tt-debug
tt dev tzc export "D:\pkg\x.tzc"            # 改 4GL 客制 → 见 skills/tt-dev-tzc
tt dict r.t --kw 应收                        # 查字典 → 见 skills/tt-dict
tt dev tzs field_add --args '{"file":"D:\\pkg\\x.tzs","table":"pmdl_t","columns":["pmdlent","pmdlsite"],"out":"D:\\pkg\\_ai.tzs"}'
                                            # 改表单 → 见 skills/tt-dev-tzs
```

**只有一个配置文件**，所有命令组共用，位置按优先级（第一个存在的胜出）：

1. `TT_CONFIG` 环境变量
2. `--config <路径>`
3. `<exe 目录>\.portable` 存在 → 便携包，配置留在包内
4. `%APPDATA%\T100\tt\config.json` —— 默认（MSI 安装的版本用这个）

`T100_HOME` 可以整体改写统一目录（如 `T100_HOME=D:\t100`）。配置里含明文口令，新建的文件权限
是 0600，**不要提交、不要外发**。可参考随包分发的 `config.example.json` 的字段结构。

## 文档地图

| 想知道 | 去哪 |
|---|---|
| 从源码构建、依赖、跑测试、打包 | [BUILD.md](BUILD.md) |
| 整体结构、分层、关系、术语 | [DESIGN_DOC.md](DESIGN_DOC.md) |
| 在仓库里改代码：改哪块先读什么、什么不能碰、做完的定义 | [AGENTS.md](AGENTS.md) |
| 怎么**用**某个命令（含各命令的坑） | [skills/](skills/) 下对应的 `SKILL.md` |
| **某个目录内部**是什么、契约是什么、怎么验 | 该目录的 `README.md`（**每个目录一份，外层讲关系、内层讲细节**） |
| 索引与写作约定 | [docs/README.md](docs/README.md) |
