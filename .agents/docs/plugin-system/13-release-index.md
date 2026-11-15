# 插件发布索引（v1）

`tt plugin update` 只认这一份契约：一份**严格 JSON**，说明"某个插件的某个版本，在某个
平台上，能从哪儿下载到哪一份字节"。

插件分发**不复用** TT 自更新的发布模型：那一套固定了 GitHub 仓库、`tag_name`、按扩展名
挑 `.msi`/`.zip` 的整包规则，是"一台机器装一份 TT"的语义；插件是"同一个 ID 有很多版本、
每个版本每种平台一个资产"，判据是 `id + version + {os}-{arch}`。
可共用的只有字节级原语（下载、重试、大小与 SHA-256 校验），那部分在 `internal/fetch`，
TT 自更新与插件更新用的是同一份实现。

## 1. 形状

```json
{
  "schemaVersion": 1,
  "plugins": [
    {
      "id": "com.example.myplugin",
      "versions": [
        {
          "version": "0.2.0",
          "hostApi": ">=1.0.0 <2.0.0",
          "assets": {
            "windows-amd64": {
              "url": "https://plugins.example.com/com.example.myplugin/0.2.0/windows-amd64.zip",
              "sha256": "9f2c…（64 位小写十六进制）",
              "size": 1234567
            },
            "linux-amd64": {
              "url": "https://plugins.example.com/com.example.myplugin/0.2.0/linux-amd64.zip",
              "sha256": "1ab3…",
              "size": 1200000
            }
          }
        }
      ]
    }
  ]
}
```

| 字段 | 必填 | 规则 |
|---|---|---|
| `schemaVersion` | 是 | **必须等于 1**。0 或更高主版本一律拒绝，不猜字段含义。 |
| `plugins[].id` | 是 | 反向 DNS（与 `plugin.json` 的 `id` 同一规则）。同一 ID 不能出现两次。 |
| `plugins[].versions` | 是 | 至少一个版本。同一插件里版本号不能重复。 |
| `versions[].version` | 是 | **严格 SemVer**（`0.2.0`；不接受 `v0.2.0`、`0.2`）。 |
| `versions[].hostApi` | 是 | 与清单里的 `hostApi` 同一种范围写法（`>=1.0.0 <2.0.0`）。 |
| `versions[].assets` | 是 | 至少一个平台资产；键必须是 `{os}-{arch}`。 |
| `assets[].url` | 是 | **必须 https**（或本机文件路径 / `file://`，见下）。 |
| `assets[].sha256` | 是 | 64 位十六进制摘要。缺失就不是"可校验的下载"，一律拒绝。 |
| `assets[].size` | 是 | 正整数字节数。解包前先用它挡掉明显不对的包。 |

未知字段一律报错（`DisallowUnknownFields`）：把 `url` 写成 `href` 这种错如果被静默忽略，
就会变成"索引看着有、更新时才发现没地址"。

## 2. 资产键 = `{os}-{arch}`

取自 `runtime.GOOS` / `runtime.GOARCH`，例如 `windows-amd64`、`linux-amd64`、
`darwin-arm64`。**没有匹配的资产 = 这个版本在当前平台不可用**，不是"随便挑一个"。

同一个包里也可以按平台分目录放不同入口（`bin/{os}-{arch}/tool.exe`，见
[11-plugin-author-guide.md](11-plugin-author-guide.md)）—— 索引决定"下哪个 zip"，
清单占位决定"zip 里跑哪个文件"，两者互不替代。

## 3. 选版本规则

`tt plugin update` 对每个插件：

1. 给了 `--to <version>`：必须是它。版本不存在、与本机 Host API 不兼容、没有本平台资产、
   或就是当前版本 → **明确报错**（用户点名了一个版本，静默换别的比失败更难查）。
2. 否则：取**比当前活跃版本新**、与本机 `connsvc.HostAPIVersion` 兼容、有本平台资产的
   **最高**版本。
3. 一个都没有 → 报"已是最新"（这不是错误，退出码 0）。

已经下载到本机但没激活的版本（上一轮升级失败留下的）仍会被选中：那条路径不再下载，
只做一次切换 —— 否则"失败后重试"会永远得到"已是最新"，把人钉死在旧版本上。

## 4. URL 策略

- `https://…`：正常发布形态。
- 本机绝对路径（`D:\share\plugins\a.zip`）与 `file:///…`：内网共享盘、离线安装。
  它不经过网络，因此不涉及 https 判断。
- `http://…`：**拒绝**。明文传输等于把可执行文件交给中间人。
  测试可以用 `UpdateOptions.AllowInsecureHTTP`（只存在于内部 API 与用例里），
  生产命令不暴露这个开关。

## 5. 索引从哪儿来（没有内置市场）

`tt plugin update` 的 `--index` 优先；不给时用**该插件上次安装时记下的索引地址**
（注册表里 `source.kind = "url"` 的 `ref`）。两者都没有就报错：

```text
插件 com.example.myplugin 是本机来源（dir:…）安装的，没有记录发布索引地址：加 --index <地址>
```

TT 不内置任何插件市场地址：装插件时用的是哪个源，升级就回哪个源。
批量升级（不给 `--id`）必须显式 `--index`，因为不同插件可能来自不同源。

## 6. 与升级流程的关系

索引只回答"下哪个包"。装上去之前，包内 `plugin.json` 还要过一遍清单校验，且：

- 包内清单的 `id` 必须等于索引里的 `id`；
- 包内清单的 `version` 必须等于索引里选中的 `version`。

不一致就拒绝 —— 那是"拿别的包冒充这个版本"，不是打包手滑。

升级/回滚的状态机见 [12-user-guide.md](12-user-guide.md) 与
`internal/plugin/upgrade.go` 顶部注释；失败不破坏可用版本的验证见
`internal/plugin/upgrade_test.go`（P6-08 那一组）。
