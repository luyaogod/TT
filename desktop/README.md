# desktop — TT Desktop（Electron 壳）

**只做壳，不做业务。** 一个窗口、一个单实例锁、一个后台子进程、一次收尾 —— 界面与全部接口
来自 Go 服务（`tt serve` 的同一套 `/debug/` 与 `/api/*`）。浏览器版与桌面版因此永远看到同一份
界面、同一份 `%APPDATA%\tt` 配置、同一批调试会话。

```
desktop/
├─ package.json          私有包；version 与仓库 VERSION 同步（make version-check 管）
├─ electron-builder.yml  appId / productName / 用户级 NSIS / extraResources
├─ .npmrc                npm 镜像（electron 运行时另由 ELECTRON_MIRROR 指，见下）
├─ build/icon.ico        窗口与安装包图标（沿用 installer/icon.ico）
├─ src/main.js           生命周期、单实例、窗口、菜单、外链、日志、IPC
├─ src/serve.js          起服务 / 等就绪 / 判断所有权 / 收尾（纯 Node，可单测）
├─ src/splash.html       启动页与失败页（loading / error / crashed 三态）
├─ src/welcome.html      首次启动的三步入口（不做第二套配置表单）
├─ src/preload.cjs       只挂给上面两个本地窗口（主窗口刻意不挂）
├─ scripts/dev.mjs       npm run dev
├─ scripts/check-syntax.mjs  语法闸门（main.js 没法单测，但语法错会让它双击没反应）
└─ test/serve.test.mjs   node --test
```

## 与 Go 侧的契约

三个开关，一条都不能改，改了就是壳与服务对不上（见 [../internal/cli/debug/serve_ready.go](../internal/cli/debug/serve_ready.go)）：

| 契约 | 谁提供 | 壳拿它做什么 |
|---|---|---|
| `--ready-json` | `tt serve --foreground` | 读**一行** JSON 拿真实地址（端口会顺延）、pid、配置路径；`attached:true` 表示同目录已有实例，壳只开窗、退出时**不**收它 |
| `--exit-on-stdin-close` | 同上 | 壳被强杀 → 管道关闭 → 服务自己优雅退出，不留隐形 `tt.exe` 占着端口 |
| `<exe目录>\.desktop` | `build_desktop.bat` | 让 `tt update` 认出"这份二进制随桌面版安装"，拒绝命令行自替（`KindDesktop`） |

## 跑起来

```bash
cd web && npm run build          # 前端产物（后端把它嵌进二进制；也可用 TT_WEB_DIR 指磁盘）
make build                       # 仓库根的 tt.exe
cd desktop && npm install        # 首次
node node_modules/electron/install.js   # 首次：electron 45 起运行时按需下载
npm run dev
```

开发时改前端不用重启壳：服务认 `TT_WEB_DIR`（见 [../main.go](../main.go)），重建 `web/dist`
后在窗口里 `Ctrl+R` 即可。壳的日志在 `%APPDATA%\TT Desktop\logs\desktop.log`。

## 打包

```bash
make desktop-dist   # = build_desktop.bat：复用便携包载荷 → 依赖 → 检查 → electron-builder
```

产物落在仓库 `dist\`：`TT-Desktop-<版本>-x64-Setup.exe`（用户级安装，不需要管理员）与
`TT-Desktop-<版本>-x64.zip`。载荷来自 `dist\tt-portable`，但**去掉** `config.json` 与
`.portable`、**加上** `.desktop` —— 桌面版的配置必须落在 `%APPDATA%\tt`，与 CLI 共用。

镜像：`electron` 运行时与 electron-builder 的 NSIS 二进制默认从 GitHub 取，
`build_desktop.bat` 里用 `ELECTRON_MIRROR` / `ELECTRON_BUILDER_BINARIES_MIRROR` 指到 npmmirror；
换网络环境改那两行即可。

## 两个目录别混

| 目录 | 放什么 |
|---|---|
| `%APPDATA%\tt` | **业务数据**：`config.json`、字典库、企业快照、源码镜像、`.tt-serve.json`（与 CLI 共用） |
| `%APPDATA%\TT Desktop` | **壳自己的**：窗口位置、`welcomed.json`、`logs\desktop.log` |

## 判据

```bash
cd desktop && npm test     # 语法闸门 + serve.js 的单元测试（纯 Node，不启 Electron）
make version-check         # desktop/package.json 与 VERSION 同版
```

Electron 的窗口、菜单、退出收尾没法在 CI 里跑 —— **每次改 `main.js` 都要在真桌面上过一遍**
这几条（`node --test` 覆盖不到它们）：

1. 冷启动到出界面，全程不出现终端窗口；
2. 关窗后任务管理器里没有残留的 `tt.exe`；
3. 任务管理器强杀壳，`tt.exe` 随之退出（stdin EOF）；
4. 命令行先 `tt serve` 再开壳：附着，关窗后命令行那条仍在跑；
5. 拔掉载荷（改名 `resources\tt`）后启动：出失败页，而不是白屏。

## 细节去哪

- 服务的启动/单实例/端口顺延 → [../internal/cli/README.md](../internal/cli/README.md)
- 页面的挂载与接口 → [../internal/web/README.md](../internal/web/README.md)
- 安装形态与自更新 → [../internal/update/README.md](../internal/update/README.md)
- 前端 → [../web/README.md](../web/README.md)
