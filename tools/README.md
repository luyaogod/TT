# tools — 打包辅助脚本与评测装置

全部是 Go（`go run` 直跑，随 `go build ./...` 编译期把关）。2026-10 从 Python 脚本
整体迁移。

| 项 | 是什么 |
|---|---|
| `zip/` | 把暂存目录打成 zip（`-flat` 打桌面版免安装包） |
| `refplugin/` | 参考插件（测试夹具） |
| `tzsmini/` | 重新生成 `testdata/tzs-mini/ws` 最小语料夹具（含 `manifest.txt`） |
| `eval/` | 工具面评测装置（见它自己的 README） |

## `zip/`

**必须递归。** 早先是在批处理里塞一句 `python -c`，用的是"列目录 + 逐个写"的写法 ——
那个组合**不会递归**：目录只会被写成一个空条目，于是包里明明该有技能文档，却只有一个空的
技能目录，而技能文档正是给 AI 用的说明书。单独成工具就是为了这一段能被读、被改、被审。

路径分隔符固定用 `/`（zip 规范），不依赖平台。**没有降级回退**：打包失败就该让构建停下来，
而不是悄悄换一个会把条目名写成反斜杠的降级实现（旧 PowerShell 回退正是这么干的）。

两种布局都由它出：

```bash
# CLI 便携包：条目顶层是暂存目录名（tt-portable-x64/…）
go run ./tools/zip .tmp-build/tt-portable-x64 dist/tt-0.6.0-x64-portable.zip

# 桌面版免安装包：条目直接落在 zip 根（解开就是应用目录，和 electron-builder 的 zip 一致）
go run ./tools/zip -flat desktop/release/win-unpacked dist/TT-Desktop-0.6.0-x64.zip
```

`-root <名字>` 可指定条目顶层目录名（缺省 = 暂存目录的 basename）。

## `refplugin/`

参考插件：同一个二进制既是 CLI 又是常驻服务，带可控故障开关，给插件系统的测试与文档当夹具。

## `tzsmini/`

从一份真工作区重新生成 `testdata/tzs-mini/ws` 与 `manifest.txt`（一次做两件事）。
生成规则与"为什么留哪些文件"见 [main.go](./tzsmini/main.go) 的文件头（每条都注明是实测）。

## 判据

```bash
go build ./...                                        # 全部工具随全仓编译把关
go run ./tools/zip -flat desktop/release/win-unpacked %TEMP%\tt-desktop-check.zip
go run ./tools/tzsmini --src <真工作区> --out %TEMP%\tzsmini-check   # 对照 manifest 行数与内容
make dist                                             # 走完整条打包链（4 个产物）
```

## 细节去哪

- CLI 安装包定义与用户级安装的理由 → [../installer/README.md](../installer/README.md)
- 桌面版安装包定义 → [../desktop/installer/tt-desktop.iss](../desktop/installer/tt-desktop.iss)
- 最小语料夹具本身 → [../testdata/tzs-mini/README.md](../testdata/tzs-mini/README.md)
