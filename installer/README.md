# installer — CLI 安装版（Inno Setup）

`tt-cli.iss` 是 CLI 安装包定义，`zh-CN-messages.iss` 是它与桌面版安装包共用的中文文案，
`icon.ico` 是它用的图标。

0.6.0 起**安装版统一走 Inno Setup**（CLI 与桌面版都是），MSI / WiX 已删除。理由：

- 编译期就生成卸载器，打包机上不需要"先跑一次安装器"（NSIS 正是死在这一点）；
- CLI 与桌面版共用一套中文向导文案，用户看到的两份安装程序是一致的。

## 为什么是"用户级"

- **装到用户目录**，不碰 `Program Files` —— **全程不需要管理员**。这与工具一贯的取向一致：
  把 tt 加进 PATH 也只写用户级环境变量、绝不碰系统 PATH。
- **PATH 由 tt 自己加/摘**：安装时 `[Run]` 调 `tt install path`，卸载时 `[UninstallRun]`
  调 `tt install path --remove`。这与 `tt install path` 命令、设置页「加入 PATH」是同一份
  实现（`internal/pathinstall`），安装包里不再复制一套注册表逻辑。
- **装完不写便携标记**：配置该落在 `%APPDATA%\tt`（用户可写），而不是程序目录
  （升级会整目录替换，且普通用户写不进去）。构建脚本暂存时已删掉 `.portable` 与
  `config.json`，`.iss` 里再断言一次。

## 只管四件事

产品、目录、PATH、升级关系。**文件清单不在这里** —— 它由 `build_cli_setup.bat` 把便携包
暂存目录整棵铺进去，所以技能增减、引擎文件变化都不需要改这个文件。

## 从 0.5.x 的 MSI 升上来

旧 MSI 在「应用和功能」里留了一条登记，而它的卸载程序删的正是同一个安装目录。安装器
`[Code]` 里的 `RemoveLegacyMSI` 会在铺文件之前按 `DisplayName='TT'` + `Publisher='luyao'` +
`WindowsInstaller=1` 找到它，静默 `msiexec /x` 卸掉；失败只记日志，不阻断安装。

## 判据

```bash
build_cli_setup.bat     # → dist/TT-<版本>-x64-Setup.exe
```

装一次、卸一次，确认：不需要管理员、PATH 被加上又被摘掉、程序目录被清干净、
配置落在用户配置目录而不是程序目录、卸载后「应用和功能」里没有残留。

## 细节去哪

- 桌面版安装包定义 → [../desktop/installer/tt-desktop.iss](../desktop/installer/tt-desktop.iss)
- 打包脚本 → [../build_cli_setup.bat](../build_cli_setup.bat)、[../build_dist.bat](../build_dist.bat)
