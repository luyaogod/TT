# TT （T100 开发CLI）
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-2-orange.svg?style=flat-square)](#contributors-)
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

## 技能（skills）

tt 的技能有两套来源，`tt install skills` 一次装完：

| 来源 | 从哪来 |
|---|---|
| 官方 | 与 `tt.exe` 同目录的 `skills/`，随包发布，走专用路径 |
| 第三方 | 你登记在 `config.json` 的 `skills.externals` 里的路径，一行一个技能目录 |

登记第三方技能在设置页「技能」分区，或直接编辑配置：

```jsonc
"skills": {
  "externals": [
    { "path": "D:\\我的技能库\\kb-faq", "label": "kb-faq" }
  ]
}
```

技能目录放在硬盘任意位置，tt 只记路径：

```
kb-faq/
├── SKILL.md      # 必须有；frontmatter 的 name 要等于目录名
└── assets/       # 可选
```

```powershell
tt install skills --to auto          # 装到 agent 会读的位置（.agents/skills 优先）
tt install skills --dry-run          # 只列计划与告警，不写盘
tt install skills --to <dir> --force # 刷新目标里的同名技能
```

三条要知道的事：

1. **第三方路径失效只告警，不拦安装**。路径换过位置、移动硬盘没插、`SKILL.md` 不合规、
   跟别的技能撞名 —— 都只跳过那一条，官方技能与其它第三方照装，退出码仍是 0。
   要在 CI 里把「装漏了一个」当失败，加 `--strict`。
2. **根级 `assets/` 不复制**。目标里的 `SKILL.md` 正文首行会写一行指回源目录的绝对路径
   （重复安装会替换那行，不会堆）。这样知识库型技能不会把目标目录撑大；
   代价是那行路径是本机的、挪了目录就失效。
3. **撞名时官方优先**；两条第三方同名，先登记的优先（顺序就是配置里的顺序）。

---

## 初次配置

```
tt serve
```

> 启动Serve后在本地打开配置网页` http://127.0.0.1:28670`
>
> 至少需要配置一个T100远程Linux环境

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
    </tr>
  </tbody>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!