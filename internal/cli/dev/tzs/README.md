# internal/cli/dev/tzs — `.tzs` 表单包线

`tt dev tzs` 的整条命令面。与 tzc 线的本质区别：

- tzc 是**代码包**管线：渲染带围栏的 prog.full.4gl、跑三道闸门、apply 原子写回；
- tzs 的 `export` 只有一件事：**纯解压**，不解围栏、不校验、不产生工作区。
  要**读写表单**走 `tt dev tzs <动词>`（open / set_spec_attr / save / …），
  由 engine/ 里那个 C# 引擎驱动，**动词表来自引擎自己的函数表**。

## 结构

- `tzs.go`：`cmdTzs` 分发（export / doctor / stop / reap / help / 具名动词）+ `Usage`（命令面文本）
- `tzs_engine.go`：引擎定位、manifest 拉取、动词调用、stop / reap / doctor
- `tzs_detail.go`：把引擎错误帧的 `detail` 渲染成人读的几行（只按 JSON 类型分派，不抄 schema）

**整条线 `DisableFlagParsing` 透传**（父包接线）：动词面由引擎 manifest 定义、
`<动词> --help` 是**引擎**的说明书（附动词索引），所以连动词都不挂成 cobra 子命令 ——
这是"参数定义只认引擎 manifest"这条单一来源红线的直接后果。

红线：**export 的产物是只读参考** —— 它就是一包文件，没有 manifest/围栏，不要手工改完再
塞回包。要改表单走 `tt dev tzs <动词>`：那是设计器自己的模型在算，改完设计器打得开。

## 输出与退出码

`--json` 下 stdout **恰好一个对象**：查询/写动词是帧 `{id, ok, result, error, ms}`；
export 是 `{"ok":true, …}`；失败是 `{"ok":false,"exit_code":N,"error":"…"}`
（实现在 [../common/scaffold.go](../common/scaffold.go)，两条线同一份）。
退出码：`0` 成功 / `1` 引擎内部错 / `2` 参数或环境错 / `4` 设计器拒绝 / `5` 传输或环境失败
（**没有 3**；2 也包括"manifest 拉不到"）。

## 判据

```bash
go test ./internal/cli/dev/tzs -count=1            # 默认档：合成包 + 用法契约（不需要引擎）
TTZS_E2E=1 go test ./internal/cli/dev/tzs -run TestE2E -count=1 -timeout 30m
                                                   # 需要真引擎（配 TTZS_EXE / TTZS_WS / TTZS_INSTALL）
./tt.exe dev tzs doctor                            # 环境自检：引擎 / 设计器目录 / 工作区 / 管道名
```

用法文本漂移的机械防线也在本包：`TestUsageTextHasNoStaleAdvice` 扫 `tzsUsage` 与总帮助
（[../common/usage.go](../common/usage.go)），禁掉已删除的 `call` 网关，以及把
`can_edit` / `can_query` / `req` 写成 `true`、`"force":true` 这几处教错的值。

**故意不覆盖**：引擎内部的格式算法 —— 那是设计器自己的程序集在算（红线 1：
不实现设计器的私有格式），Go 侧只测传输、契约与文案。

## 细节去哪

- 引擎客户端协议（管道、manifest、守护进程） → [../../../dev/tzs/README.md](../../../dev/tzs/README.md)
- 结构与接线、根开关桥 → [../README.md](../README.md)
- 各动词的参数与示例 → [`skills/tt-dev-tzs/SKILL.md`](../../../../skills/tt-dev-tzs/SKILL.md)
