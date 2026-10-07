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

> 启动Serve后在本地打开配置网页` http://127.0.0.1:28675`
>
> 至少需要配置一个T100远程Linux环境

---
