# TT 架构与统一配置设计

TT 是 `TDebug`、`TDev`、`TDictCli` 三个工具合并后的统一项目。

- 一个仓库、一个二进制 `tt`、一份配置、一个本地 Web 服务。
- 三个工具以**子命令组**形式并存：`tt debug …` / `tt dev …` / `tt dict …`。
- 三套原有能力不做功能取舍，只消除重复实现。

## 1. 三个来源项目

| 原项目 | 原模块/二进制 | 职责 | 规模 |
|---|---|---|---|
| TDebug | `tdebug` / `tdebug.exe` | T100 作业调试器：SSH 驱动 `fglrun -d` 的 `(fgldb)` 文本调试协议，提供本地 Web 调试界面（源码/断点/调用栈/变量/接口日志）与命令行控制端 | Go 16.3k 行 + TS 6.4k 行 |
| TDev | `tdev` / `tdev.exe` | T100 设计器 `.tzc`/`.tzs` 包的安全编辑管线（export/status/verify/apply/unlock/rename/newfn），纯 stdlib | Go 18.1k 行 |
| TDictCli | `tdict` / `tdict.exe` | T100 ERP 数据字典查询（`r.t`/`r.v`/`desc`/`scc`/`r.q`/`prog`…）、DB 同步、源码镜像、可视化配置页 | Go 12.6k 行 + TS 1.5k 行 |

## 2. 合并前已存在的重复

合并不是"把三个目录塞一起"，而是要清掉三处**已明确标注需要手工同步**的重复：

1. **配置路径解析**：`TDebug/cli/root.go` 与 `TDictCli/cli/root.go` 各有一份
   `toolsHome` / `userConfigDir` / `isPortable` / `migrateLegacyConfig`，
   两处注释都写着"与对方保持一致，改动请两边同步"。
2. **`cfgfile` 包**：两份近乎逐字节相同，只差一个节访问器
   （TDebug 的 `Debug(root)` vs TDictCli 的 `Hosts(root)`）。
3. **`dbconfig` / `host` / `erpdb` / `pathinstall` / `install` 命令**：
   两边各一份，`install.go` 注释写着"与 TDev / TDictCli 一致（改动请三边同步）"。

合并后这些都只有一份实现。

## 3. 目录布局

```
TT/
├─ main.go                  //go:embed all:web/dist → cli.Execute(webFS)
├─ go.mod                   module tt
├─ README.md
├─ build_portable.bat       构建便携包
├─ build_msi.bat            构建用户级 MSI（需要 WiX v3）
├─ installer/tt.wxs         MSI 的产品/目录/PATH/升级定义
├─ tools/zip.py             打包用 zip 助手
├─ tools/wix_removefolders.py  给采集出的组件补删除目录(ICE64/ICE38)
│
├─ internal/
│  ├─ config/               ★ 统一配置层（本设计的核心）
│  │  ├─ paths.go           配置位置解析：T100_HOME / .portable / %APPDATA%\T100\tt
│  │  ├─ cfgfile.go         Open / Save(原子写) / Edit —— 单一写入口
│  │  ├─ schema.go          配置根结构体 + 各节类型 + 缺省值
│  │  ├─ pathutil.go        路径型取值的解析与派生状态（AbsPath / DirStatusOf / SyncTarget）
│  │  └─ migrate.go         旧配置迁移与合并（tdebug + tdict → tt）
│  ├─ cli/                  cobra 根命令与公共约定
│  │  ├─ root.go            tt 根命令、全局 flag、--help 约定
│  │  ├─ env.go             tt env …       统一环境管理（原 tdict env + tdebug env/topent）
│  │  ├─ config.go          tt config …    路径/查看/修改/迁移/校验（校验也在这里）
│  │  ├─ serve.go           tt serve …     统一本地 Web 服务
│  │  ├─ install.go         tt install skills|path
│  │  ├─ debug/             tt debug …     ← 原 tdebug cli/
│  │  ├─ dev/               tt dev …       ← 原 tdev internal/cli/
│  │  └─ dict/              tt dict …      ← 原 tdict cli/
│  │
│  ├─ debug/                ← 原 TDebug debug/   调试内核：fgldb 驱动、会话、REST+WS API
│  ├─ dev/                  ← 原 TDev internal/  管线：pkgfile/tapfile/tglfile/fgl/synth/fence/verify/split/store/model
│  ├─ dict/                 ← 原 TDictCli db/ + live/ + dbsync/ + server/  数据源、字典查询与两个长跑动作的 HTTP 面
│  │
│  ├─ host/                 ★ 合并 TDebug/host + TDictCli/host：SSH/PTY、终端解析、环境探测、DB 探测、镜像引擎
│  ├─ dbconfig/             ★ 数据库连接模型（两边合并）
│  ├─ erpdb/                ★ Oracle(go-ora) / Kingbase(pgx) 连接器（两边合并）
│  ├─ safesql/              ← 原 TDebug safesql：只读 SQL 守卫
│  ├─ sshtun/               ← 原 TDictCli sshtun：SSH 端口转发隧道
│  ├─ output/               ← 原 TDictCli output：表格/JSON/CSV 输出（CJK 宽度感知）
│  ├─ pathinstall/          ★ 用户 PATH 安装（两边合并，含 Windows 注册表实现）
│  └─ web/                  统一 HTTP 服务：路由装配、配置与派生状态端点、SPA 静态服务
│
├─ web/                     前端（React 18 + Vite 6 + TS + Tailwind 4）
│  ├─ app/                  ← 原 TDebug/web  唯一一套 SPA：调试工作台 +
│  │                        覆盖三个工具的统一设置页（站点管理/数据字典/DEBUG/应用设置）
│  └─ shared/               共享层：主题变量与机制、UI 基元、设置页布局件、cn
│
├─ desktop/                 Electron 外壳（原 TDebug/desktop）
├─ skills/                  AI 技能：tt-debug / tt-dev / tt-dict / erp-read
├─ docs/                    本设计与迁移记录
└─ testdata/                FGL 夹具（原 TDev/testdata）
```

`★` = 需要真正合并（而非搬移）的包；`←` = 基本原样搬入。

## 4. 统一配置管理

### 4.1 单一配置文件

合并后**只有一个** `config.json`，顶层节：

```jsonc
{
  "schemaVersion": 2,

  // 共享：SSH 环境 + 数据库连接。三个工具读同一份。
  "hosts": {
    "activeEnv": "开发环境",
    "sshs": [
      {
        "name": "开发环境",
        "host": "10.0.0.1", "port": 22, "user": "tiptop", "password": "…",
        "zone": "T100", "topent": "BBDL512840855a",
        "db": {
          "type": "oracle",              // oracle | kingbase
          "host": "10.0.0.2", "port": 1521,
          "service": "t100",             // oracle: service；kingbase: database
          "database": "",
          "readonlySql": true,           // 省略 = 启用只读守卫
          "viaSsh": { "host": "", "port": 0, "user": "", "password": "",
                      "bindPort": 0, "remoteHost": "", "remotePort": 0 },
          "accounts": [ { "account": "ds", "password": "…" } ]
        }
      }
    ]
  },

  // 本地 Web 服务（统一：原 tdebug debug.listen 与 tdict server 同为 127.0.0.1:28670）
  "listen": "127.0.0.1:28670",

  // 工具自有设置
  "debug": { "activeEnv": "", "launchArgs": "…", "watchdogSeconds": 1800,
             "fglserver": "", "termWidth": 200, "termHeight": 50,
             "printElements": 1000, "persistBreakpoints": true },
  "query": { "source": "auto" },        // auto（缺省=在线优先）| local | <环境名>
  "mirror": { "dir": "" },
  "bdldoc": { "dir": "" },
  "sync":   { "target": "" },
  "tdev":   { "workspaceSuffix": "-ws", "defaultOut": "" }
}
```

要点：

- **`hosts` 是唯一的环境/连接登记处**。原来 `tdebug` 把环境放在 `debug.sshs`、`tdict` 放在 `hosts.sshs`，schema 完全一致（`name/host/port/user/password/zone/topent/db`），合并后只留 `hosts.sshs`。
- **`debug` 节降级为纯工具设置**，不再承载环境列表。这是本次合并最关键的数据结构变更。
- **`tdev` 首次获得配置节**：原 TDev 完全没有配置系统，所有参数都是每次调用的 flag。新增的 `tdev` 节只放"跨调用稳定"的默认值，flag 仍然优先。
- 每个工具可用 `<工具>.activeEnv` 覆盖 `hosts.activeEnv`；留空即继承。

### 4.2 配置位置

沿用两边已一致的规则，但收敛为单一实现 `internal/config/paths.go`：

1. `TT_CONFIG` 环境变量（兼容旧名 `TDEBUG_CONFIG` / `TDICT_CONFIG`）
2. `--config <路径>`
3. `<exe 目录>\.portable` 存在 → 便携包，配置留在包内 `<exe 目录>\config.json`
4. `%APPDATA%\T100\tt\config.json` —— 默认统一位置
5. 旧位置兜底：`<exe 目录>\config.json`、`<cwd>\config.json`、`%APPDATA%\T100\tdebug\config.json`、`%APPDATA%\T100\tdict\config.json`、`%APPDATA%\TDebug\config.json`

`T100_HOME` 环境变量整体改写统一目录（如 `T100_HOME=D:\t100`）。

数据目录 = 配置所在目录，所以 `srccache/`、`debug-bps/`、`.tt-serve.json`、`logs/` 都跟着落在一起。

### 4.3 旧配置迁移（`internal/config/migrate.go`）

首次运行且目标配置不存在（或 `schemaVersion < 2`）时执行一次，结果落盘并写 `schemaVersion: 2`：

1. **读入所有候选旧配置**：`%APPDATA%\T100\tdebug\config.json`、`%APPDATA%\T100\tdict\config.json`、便携包/工作目录下的 `config.json`。
2. **拆解 TDebug 的 `debug` 节**：把 `debug.sshs` 提升为 `hosts.sshs`，其余键（`listen`/`launchArgs`/`watchdogSeconds`/`termWidth`/`termHeight`/`printElements`/`persistBreakpoints`/`fglserver`）留在 `debug`。
   *注意*：这正是 TDictCli 原有的 `debug → hosts` 整节重命名会踩的坑——那条规则在合并后会把 TDebug 的调试设置一并吞掉，必须替换。
3. **合并两份 `hosts.sshs`**：按 `name` 去重；同名时优先保留 TDictCli 那份（其结构已含 `activeEnv` 且是字典查询的现役配置）。
4. **`hosts.activeEnv`**：取 TDictCli 的值；缺失时取第一台环境名。
5. **`listen`**：取 TDebug 的 `debug.listen`，否则 `127.0.0.1:28670`。
6. **原文件不删除**，仅在新位置写合并结果；`.bak` 保留一份。
7. 迁移前打印将合并的文件清单，`tt config migrate --dry-run` 可预览。

### 4.4 唯一写入口

所有写操作走 `internal/config/cfgfile.go` 的 `Edit(path, validate, mutate)`：
`Open → validate → mutate → Save`，`Save` 为 临时文件 + `os.Rename` 原子替换，失败不留半截文件。
禁止任何调用方自己 `json.Unmarshal` 后 `os.WriteFile`。

读取侧新增**类型化 schema**（`internal/config/schema.go`）收敛原本散落在 6 个文件里的匿名 struct 反序列化
（`host/env.go`、`cli/source.go`、`server/server.go`、`server/bdldoc.go`、`server/mirror.go`、`cli/bdldoc.go`、`cli/mirror.go`）。

## 5. 统一 Web 服务与配置页面

### 5.1 服务

单进程、单端口（默认 `127.0.0.1:28670`，占用时自动向后探测），由 `tt serve` 启动，
一套页面 + 一个统一 API：

- `/debug/`             → 调试工作台 SPA
- `/debug/#settings/*`  → 统一设置页（站点管理 / 数据字典 / DEBUG / 应用设置）
- `/api/*`              → 统一 REST API：`/api/hosts` 读写配置的各节，
                          `/api/config/status` 给派生状态（路径存不存在），
                          `/api/install` 管 PATH，`/api/mirror` `/api/dbsync` `/api/bdldoc`
                          是字典类的长跑动作

**配置**（环境与数据库、查询数据源、镜像目录、同步目标、BDL 目录）全部走 `/api/hosts`
的分节写入；**动作**（拉源码镜像、跑字典同步）留在设置页「数据字典」分区的对应卡片里。
初始设计是两套页面并存互链，但设置一旦统一，字典页就只剩这两个动作 —— 于是那个
SPA 被删掉，动作并进设置页。

### 5.2 环境配置的单一数据源

原来 `SettingsView.tsx`（TDebug，650 行）读写 `debug` 节，`SettingsView.tsx`（TDict，496 行）读写 `hosts` 节。
合并后两边都读写 `hosts` 节：

- `GET  /api/hosts` → `{ activeEnv, sshs: [...] }`
- `PUT  /api/hosts` → 整节替换，但**保留顶层其他节**（沿用原 `hSettingsPut` 只替换自己那节的做法）
- `POST /api/dbprobe`、`POST /api/dbaccverify`、`POST /api/conntest` → 环境页的三个辅助端点，两边共用

TDebug 的设置页因此改为：环境列表取自 `hosts`，外观/高级页仍写 `debug`。
TDict 的设置页保持原样（它本来就是 `hosts`），并新增指向调试工作台的链接。

## 6. 命令面

```
tt debug …                 原 tdebug：probe / serve / db / desktop
                           start / exec / status / quit / wslogs / sql / wsdebug / why / wait / mode
                           stop / source / logs / locate / resolve / interrupt
tt dev …                   原 tdev：tzc export|status|verify|apply|unlock|rename|newfn|selftest
                           tzs export
tt dict …                  原 tdict：r.t / r.v / desc / scc / r.q / msg / sysp / docp / prog
                           db status|sync|list|ping|discover / mirror / bdldoc
tt env …                   统一环境管理：list / show / use / topent   （原 tdict env + tdebug env/topent）
tt config …                统一配置：path / show / get / set / migrate / validate
tt serve                   统一本地 Web 服务（工作台 + 统一设置页）
tt install skills|path     统一安装（原三份同形命令合并为一份）
tt version
```

全局 flag：`--config`、`--json`、`--csv`、`-v/--verbose`、`--env`（替代 `--conn`）。
原 `--conn` 保留为 `--env` 的别名。

**旧命令名兼容**：`tt` 在 `args[0]` 处接受 `tdebug`/`tdev`/`tdict` 三个别名，
分别等价于 `tt debug` / `tt dev` / `tt dict`，方便既有脚本与 AI skill 不改就可用。

## 7. 构建与打包

- 前端：`cd web && npm install && npm run build` → `web/dist/`
- 后端：`go build -trimpath -ldflags "-X tt/internal/cli.Version=…" -o tt.exe .`
- 便携包：`build_portable.bat` → `dist/tt-portable.zip`（`tt.exe` + `config.example.json` + `README.md` + `skills/` + `.portable`）
- MSI：`build_msi.bat` → `dist/TT-<版本>-x64.msi`。**用户级安装**（装到
  `%LOCALAPPDATA%\Programs\TT`、只追加用户 PATH、免管理员），刻意不带 `.portable`
  与 `config.json`：配置该落在 `%APPDATA%\T100\tt\config.json`。定义在
  `installer/tt.wxs`，文件清单由 `heat.exe` 采集，卸载要删的目录由
  `tools/wix_removefolders.py` 补（MSI 的 ICE64/ICE38 对用户级安装的要求）。
  需要 WiX v3 工具集，`WIX_BIN` 指向它。
- 桌面版：`build_desktop.bat` → Electron 安装包（复用原 TDebug desktop 外壳，改为拉起 `tt.exe serve`）

`//go:embed all:web/dist` 要求前端先构建，`main.go` 已带空目录占位与引导页兜底。

## 8. 实施顺序

见 `docs/MIGRATION.md`。
