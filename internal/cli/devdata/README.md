# internal/cli/devdata — `tt gen`

为 model 准备原材料的命令组。领域实现在 [../../devdata](../../devdata)，本包只做参数、
输出与退出码（与 Web 共用同一份实现，所以 `tt gen fetch` 与设置页拉出来的 sha 完全一致）。

> **与 `tt file mirror` 不是一回事。** 那个是"读代码用的镜像"（全量、落 `srccache/`、
> 会被 `tt cache` 清理）；这里是"给 model 的原材料"（按程序成套、带 manifest 指纹、
> 清理由 `tt gen clean` 负责）。概念对照见 [../../devdata/README.md](../../devdata/README.md) 的
> "与读侧的区别"。

## 命令

环境一律点名（`--at <环境>` 或 `--env <环境>`），没有默认环境。
`--at <环境>@<企业>` 里的企业决定规格库用哪个数据库账号（不给就用账号清单首项）。

### `tt gen status [--probe]`

共用资料 + 各程序的状态。**默认不出网**（只读本地 manifest 与文件，打开即快）；
`--probe` 才向服务器核对 mtime/大小。

```bash
tt gen status --at 福联测试tst
tt gen status --at 福联测试tst --probe --human
```

JSON 载荷：`{env, dir, present, dict, mdl, progs[], jobs[], warnings[]}`。
程序行带 `stale` 与 `staleReasons[]`（`missing-local` / `sha-mismatch` / `remote-newer` /
`remote-gone` / `version-changed`）。

### `tt gen mirror [--only dict|mdl] [--force]`

把本环境的规格表（53 张 `DZ*_T` → SQLite）与模板树（`$TOP/erp/mdl`）拉到
`devdata/<环境>/`。**幂等**：已同步过就跳过（`skippedPull=true`，不报错）；`--force` 重拉。
`--only` 只刷一半，另一半的 manifest 条目原样保留。

进度打在 **stderr**（每 400ms 最多一行），stdout 只留最终 JSON/表格。
Ctrl+C 可中断 —— 中断后临时目录会被清掉，正式目录保持拉取前的状态。

```bash
tt gen mirror --at 福联测试tst
tt gen mirror --at 福联测试tst --only dict
tt gen mirror --at 福联测试tst --only mdl --force
```

单表超 200 万行会被跳过并记进 `warnings`（实测 `DZBA_T` 781 万、`DZAZ_T` 462 万，
都是生成器的运行时产物/计数，不是入参）。缺表进 `missing`，不判失败。

### `tt gen fetch <程序> [--module <模块>] [--with-design] [--form] [--force]`

按程序拉一套：`tab`（蓝图）、`tsd`、`tap`、`tgl`（骨架，含 `tgx/tgx2`）、`4gl`（产物）、
`inc/*_{mask,rep}.4gl`。缺必需项（`tab`/`tgl`/`4gl`）**整次失败、不留半套**；
缺可选项只进 `missing[]`。

模块默认从程序名前三位推断（`axmt500_wf` → `axm`），推断不出时用 `--module` 指定。
已拉过的程序需要 `--force`（整包替换，不合并）。

```bash
tt gen fetch axmt500_wf --at 福联测试tst
tt gen fetch axmt500_wf --at 福联测试tst --with-design --form
tt gen fetch cxm999 --at 福联测试tst --module cxm --force
```

`--with-design` 把本地工作区（`hosts.sshs[].workspace`）里的 `<prog>*.tzs` 复制进
`design/`（只归档、不解析）；`--form` 才拉 `.42f`（二进制，体积大）。

### `tt gen ls`

只列已拉取的程序（小表）：大小、拉取时间、是否过期、缺项。等价 `status` 里的程序清单。

```bash
tt gen ls --at 福联测试tst --human
```

### `tt gen check <程序>`

不重拉，只与服务器核对并给出**过期明细**（哪一条触发了、为什么）。

```bash
tt gen check axmt500_wf --at 福联测试tst --human
```

### `tt gen clean [--prog <程序>] [--yes]`

删除开发资料。**默认 dry-run**（只列出将删的内容与体积），`--yes` 才真删。

为什么不跟 `tt cache` 一起清：见 `tt gen --help` 与 [../../devdata/README.md](../../devdata/README.md)
的 ADR-02 —— 这份材料删了要重新拉（几十 MB + 数据库查询）。

```bash
tt gen clean --at 福联测试tst                    # 看会删什么
tt gen clean --at 福联测试tst --prog axmt500_wf  # 只看这个程序
tt gen clean --at 福联测试tst --prog axmt500_wf --yes
```

## 输出形态

默认 JSON 信封（`ok/source/env/…` + `data`），`--human` 出小表。两种形态**内容一致**：
表格只换展示形态，不偷偷少字段。`--csv` 与其它命令组同一套规则。

## 退出码

| 码 | 含义 |
|---|---|
| 0 | 成功（含 `mirror` 的"已同步、跳过"） |
| 2 | 用法或身份错：环境不在清单、程序名非法、模块推断不出、已拉过没 `--force`、同类任务在跑 |
| 4 | 未拉取 / 目标不存在（`check` 与 `clean` 的 404 语义） |
| 5 | 拉取失败：网络、库、写盘；或用户中断（`CANCELLED`） |

## 帮助与验收

```bash
tt gen --help
tt gen status --at-list
tt gen fetch --at 福联测试tst axmt500_wf   # 真实拉一次（约 5 秒，4.9 MB）
tt gen check axmt500_wf --at 福联测试tst
tt gen fetch ../x --at 福联测试tst         # 退 2，程序名非法
tt gen check nope --at 福联测试tst         # 退 4，未拉取
```

测试：`go test ./internal/cli/devdata -count=1`（输出形态 + 错误码映射）。
