# TT （T100 开发CLI）
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

TT 是用于开发鼎捷数智旗下的大型ERP系统（T100）所构建的CLI工具集

---

## 安装

参考 https://github.com/luyaogod/TT/releases —— 每个版本 4 个产物：**CLI / 桌面版** ×
**安装版 / 便携版**（x64）。

- 桌面版（推荐给不常用命令行的同学）
    - 下载 `TT-Desktop-<版本>-x64-Setup.exe`，双击安装，不需要管理员，装完双击「TT Desktop」即可
    - 第一次打开直接进工作台，在「设置」里连上 T100 环境就能开工（技能与命令行也在那里装）
    - 它和下面两种发行版共用同一份配置（`%APPDATA%\tt`），所以随时可以换来换去
- CLI 安装版
    - 下载 `TT-<版本>-x64-Setup.exe`，双击安装：装到 `%LOCALAPPDATA%\Programs\TT`，
      自动加入**用户** PATH，不需要管理员；卸载时会把 PATH 里那一项一起摘掉
- CLI 便携版
    - 下载 `tt-<版本>-x64-portable.zip`，解压到任意目录即用（配置留在包内，不写用户目录）
    ```powershell
    #手动将tt安装到用户PATH
    tt install path
    ```

---

## 升级

```powershell
tt update check     # 查有没有新版（不下载不安装）
tt update           # 查 + 下载校验，问一句再装
tt update log       # 上一次升级走到哪一步
```

`tt update` 按你现在的安装形态装：便携包就地覆盖程序目录（`config.json` 与目录里
你自己的文件不碰），安装版重新下载同一架构的 `Setup.exe` 静默安装（仍是用户级，不需要
管理员）。升级全程需要能访问 GitHub；
内网要经代理时用 `tt update --proxy http://host:port`，或把地址写进 `config.json` 的
`net.proxy`。

> 0.6.0 起安装版统一是 Inno Setup，不再发布 MSI。0.5.x 用 MSI 装过的机器可以直接装
> 0.6.0 的安装版：安装器会先静默摘掉旧 MSI 的登记，再铺新文件，不会留下两条「应用和功能」。

升级会把 agent 目录里的技能树一并刷成同版（`tt install skills` 装过哪些地方记在
数据目录的 `.tt-skills.json`）。如果那一步没成，`tt version` 会提示你手动跑
`tt install skills --force`。

MCP/Agent 环境里注意：默认会问一句再装，脚本里要用 `tt update --yes`；
先看它打算做什么可以 `tt update --dry-run`。

桌面版**不参与**上面这套自更新：它内置的那份 `tt` 跟着桌面版整体升级（在应用里
菜单「帮助 → 检查更新」，或重新下载安装包覆盖安装）。命令行的 `tt update` 会拒绝
替换桌面版的载荷，这是有意的 —— 否则桌面版下次启动加载到的就不是它发布的那一份了。

---

## 初次配置

```
tt serve
```

> 启动Serve后在本地打开配置网页` http://127.0.0.1:28670`
>
> 至少需要配置一个T100远程Linux环境

## 技能（skills）

`tt install skills` 把官方技能与登记的第三方技能装到**一个**目录（`tt update` 刷新时用它）。

要一次分发到多个 AI 工具，用 `tt skills deploy`：

```bash
tt skills list                  # 看技能源、五个平台的检测与已分发情况
tt skills deploy --all          # 覆盖写入全部五个平台（当前项目）
tt skills deploy --global --agent claude   # 装到用户目录，所有项目共用
tt skills deploy --agent claude --agent codex
tt skills remove --agent claude --yes
```

平台目录（相对根目录）：`.agents/skills`、`.dsh/skills`、`.workbuddy/skills`、
`.claude/skills`、`.codex/skills`。默认根是当前项目；`--global` 或设置页「全局」把根换成
用户目录（`~/.workbuddy/skills`、`~/.claude/skills` 等）。同步是**覆盖更新**，不做版本检测；清单在
`<根>/.agents/.tt-manifest.json`（项目级或用户目录各一份）。设置页「技能 → 分发」是同一份
实现的图形面。

## 插件（plugins）

TT 支持第三方插件：一个 zip 可以贡献 CLI 命名空间（`tt <namespace> …`）、常驻服务、
独立页面和 skills，并通过 Host API 复用 TT 已经配好的数据库/SSH 连接（插件拿不到
明文口令，能用哪些环境由你逐个授权）。

```bash
tt plugin install ./myplugin.zip     # 只解包注册，不执行插件代码
tt plugin enable  com.example.myplugin --allow profiles.list,db.connect --env prod
tt plugin list --all                 # 版本、启停、贡献面
tt plugin doctor                     # 清单/兼容/页面/服务/技能 自检
tt plugin logs com.example.myplugin  # 最新一次运行的日志
tt plugin update --index <发布索引> --dry-run   # 手动升级（不自动）：先看计划
tt plugin rollback com.example.myplugin         # 切回上一个版本
tt plugin disable com.example.myplugin
```

- **装插件 = 运行别人的本机程序**，TT 不做沙箱；只装可信来源。
- 升级是手动的：TT 不后台检查、不自动下载。升级失败会自动回滚到原来的版本。
- Host API 独立挂在 `/hostrpc/`（ConnectRPC v1），与设置页的连接测试是同一份实现。

作者与用户文档：[插件作者指南](.agents/docs/plugin-system/11-plugin-author-guide.md)、
[用户指南](.agents/docs/plugin-system/12-user-guide.md)、
[发布索引格式](.agents/docs/plugin-system/13-release-index.md)。

## TT SKILL HUB

如果需要贡献skill，请遵循下面的格式，将你的skill仓库连接、摘要和描述通过PR的形式推送到本文档本章节

|name&link|abstract|description|
|---|---|---|
|[bdl-tutorial](https://github.com/luyaogod/bdl-tutorial.git)|BDL语法SKILL|当不理解T100项目中的`.4gl`文件里的BDL语法规范时，请激活此技能。|

---

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/liuwpa"><img src="https://avatars.githubusercontent.com/u/61258913?v=4?s=100" width="100px;" alt="liuwpa"/><br /><sub><b>liuwpa</b></sub></a><br /><a href="https://github.com/luyaogod/TT/commits?author=liuwpa" title="Code">💻</a> <a href="https://github.com/luyaogod/TT/commits?author=liuwpa" title="Documentation">📖</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/songyuliang1"><img src="https://avatars.githubusercontent.com/u/111861624?v=4?s=100" width="100px;" alt="songyuliang1"/><br /><sub><b>songyuliang1</b></sub></a><br /><a href="https://github.com/luyaogod/TT/issues?q=author%3Asongyuliang1" title="Bug reports">🐛</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/Rainco"><img src="https://avatars.githubusercontent.com/u/211707777?v=4?s=100" width="100px;" alt="Rainco"/><br /><sub><b>Rainco</b></sub></a><br /><a href="https://github.com/luyaogod/TT/issues?q=author%3ARainco" title="Bug reports">🐛</a></td>
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!