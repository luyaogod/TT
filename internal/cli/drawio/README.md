# internal/cli/drawio — tt drawio 命令组

原型图线的命令面：开关、落点、输出、退出码。能力实现在 [`../../drawio`](../../drawio/README.md)。

## 命令

| 动词 | 做什么 |
|---|---|
| `lib` | 把内置的形状库导出成 drawio 能加载的 `.xml`；`--catalog` 只把形状清单打到 stdout |
| `compose <spec.json> [-o <文件>]` | 把排版规格展开成 `.drawio`；不带 `-o` 时打到 stdout（摘要走 stderr） |

## 落点

`lib` 默认写到**当前目录下的 `dist/`**（在仓库里跑就是仓库根的 `dist/`），`-o <目录>` 覆盖。
它与仓库的其它构建产物同一处，也**同样不入库**（`.gitignore` 罩着 `/dist/`）。

`compose` 的 `-o` 指**文件**；不写就打到 stdout，**不落盘** —— 因为图纸是给人看的，
写在哪儿由用的人决定（`tt drawio compose spec.json > x.drawio` 直接可用）。

为什么不落数据目录：库文件是**给人拿走的**（拖进 drawio、或放进项目当素材），不是
可再生的缓存；放当前目录下，在哪儿跑就在哪儿拿到，不必先 `tt config path` 查一遍路径。

## 退出码

```
tt drawio   0 成功 / 1 用法或运行失败 / 2 输入错（形状源坏了、spec 不合法、包读不出）/ 3 产物自查未通过
```

`2` 与 `1` 分开是给脚本用的：`1` 是"你命令敲错了"，`2` 是"命令没错，喂进来的东西不对"。
权威表在 [`../README.md`](../README.md)，根帮助里那张是它的复述。

## 判据

```bash
go test ./internal/cli/drawio -count=1
./tt.exe drawio nope; echo $?          # 期望 1，且把可用子命令列出来
./tt.exe drawio lib && ls dist/        # 期望 dist/ 下多出两个 .xml
cmp dist/t100-controls-v0.1.0.xml testdata/drawio/t100-controls-v0.1.0.xml
./tt.exe drawio compose testdata/drawio/purchase-order-form.json -o "$TEMP/po.drawio"
./tt.exe drawio compose testdata/drawio/ainq120-spec.json | cmp - testdata/drawio/ainq120.drawio
```

期望：`ok`；`nope` 退 1 并列出两个动词；导出的库与 `compose` 的产物都与 `testdata/drawio/`
里那两份 **Node 版基线**一致（库解码后逐条相同，`.drawio` 逐字节相同）。

覆盖的是：cobra 路由与退出码（`2` 输入错 / `3` 产物自查未通过）、缺省落点是 `<cwd>/dist`、
`compose` 不带 `-o` 时 XML 走 stdout 而摘要走 stderr（否则重定向出来的文件会被污染）、
导出物可被 `parseXml → JSON.parse` 两步读回来、以及"两次导出逐字节相同"。

**故意不覆盖**：形状拖进 drawio 长什么样 —— 那是人打开 drawio 看一眼的事，本层证不了。

## 细节去哪

- 形状源、合成规则、转义那两层为什么不能改 → [`../../drawio/README.md`](../../drawio/README.md)
- 对照基线是什么、为什么不许重新生成 → [`../../../testdata/drawio/README.md`](../../../testdata/drawio/README.md)
