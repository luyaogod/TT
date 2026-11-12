# TT （T100 开发CLI）
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

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

平台目录（相对根目录）：`.agents/skills`、`.dsh/skills`、`.codebuddy/skills`、
`.claude/skills`、`.codex/skills`。默认根是当前项目；`--global` 或设置页「全局」把根换成
用户目录（`~/.agents/skills` 等）。同步是**覆盖更新**，不做版本检测；清单在
`<根>/.agents/.tt-manifest.json`（项目级或用户目录各一份）。设置页「技能 → 分发」是同一份
实现的图形面。

## TT SKILL HUB

如果需要贡献skill，请遵循下面的格式，将你的skill仓库连接、摘要和描述通过PR的形式推送到本文档本章节

|name&link|abstract|description|
|---|---|---|
|[bdl-tutorial](https://github.com/luyaogod/bdl-tutorial.git)|BDL语法SKILL|当不理解T100项目中的`.4gl`文件里的BDL语法规范时，请激活此技能。|
|[tt-erp-query](https://github.com/songyuliang1/tt-erp-query.git)|T100通用取数|当需要查 T100 任意表的结构、字段含义，或跑只读 SQL 批量取数时，请激活此技能。|

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