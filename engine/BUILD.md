# engine/ — `.tzs` 表单设计器引擎（C#）

这个目录是 **tt 的引擎**，不是 tt 的一部分：它读写的 `.tzs` 格式是 T100 设计器私有的，
而这里的代码**不实现那个格式**——它反射调用**已安装的设计器自己的程序集**。我们这部分
只有 200 KB（`TzsCli.dll` 26 KB + `TzsCli.Designer.dll` 180 KB），设计器那部分是 10.5 MB。

`tt dev tzs` 通过 JSON-RPC 驱动它（见 `internal/dev/tzs/`）。

## 为什么必须单独构建

**1. 它不属于 Go 的构建链。** 用 `csc.exe`（Framework64 v4.0.30319，**C# 5**——没有模式匹配、
没有 `nameof`、没有字符串插值）编译，引用 GAC 里的 WPF 程序集。把它塞进 `.bat` 等于用 cmd.exe
重写一份已经踩平过 CS0433 的脚本，得到两份会漂移的实现。`build_portable.bat` 只**采集**产物。

**2. 设计器程序集：入库在 `engine/designer/`，跟着仓库走。**

```
构建期  INSTALL = TZSCLI_INSTALL（可选覆盖）或 <本目录>\designer（仓库里那份）
运行期  TZSCLI_INSTALL（可选覆盖）或 <引擎自己的目录>\designer
```

- **构建期**：`-r:$INSTALL\Newtonsoft.Json.dll` —— 一个编译期引用；
- **运行期**：`Assembly.LoadFrom` 设计器的 `SpecDesignerCommon.dll` / `FormEditor.dll` /
  `UndoRedoFramework.dll`，以及语言字典（见下）。

两处指向的是**同一份东西**：`engine/designer/` 里那 28 个 dll。`build.sh` 编完之后把它们采到
`$OUT/designer/`，于是 `engine/out/tzs-server.exe` 不需要任何环境变量就能跑起来 —— 那个布局
（`<引擎目录>\designer\`）与发行包里的完全一致。

**为什么入库而不是让用户自己装。** 同一份 tt 在两台机器上，如果设计器目录来自各自的配置，两边
跑的就是两版设计器 —— 同一个 `.tzs` 行为不同，而报错里看不出来。设计器跟着仓库/包走之后，
"运行环境一致"是**分发这件事本身**保证的，不需要谁去对齐配置；从 clone 到能跑 `.tzs` 也只有
`go build` + `./build.sh` 两步。

代价是版本变更会进历史：换设计器 = 换 `engine/designer/` 下的文件并提交（约 11 MB，git 压缩后
约 4.4 MB）。好处是那次提交就是"这一版钉在哪一版"的记录，可 review。

`config.json` 里因此**没有** `tzs.installDir` 这个键。`TZSCLI_INSTALL` 保留为**覆盖**手段 ——
拿另一版设计器来验证时用，平时不用设。

**3. 重编会让所有在跑的守护进程变成孤儿。** 这是最硬的一条。

守护进程的管道名 = `hash(工作区)` + **本程序集的 MVID 前 8 位**（`src/Designer/Rpc.cs`）。
MVID 每次重编都变，所以：

- 客户端**永远够不到**跑着陈旧字节的守护进程——这是刻意的，否则你会和一个旧行为对话而看不出来；
- 代价是：**每一次重编，上一个构建起的守护进程就再也停不掉了**（新名字没人监听，旧名字没人知道）。

把引擎构建挂在 tt 的每次构建上，等于每次发版都制造一批停不掉的进程——一个和 tt 无关的构建
步骤造成用户可见的后果。所以**只在引擎真的改了时才构建**。

清理由 `tt dev tzs reap` 兜底（枚举 exe 路径等于我们自己 payload 的 `tzs-server.exe`）。

## 构建

```bash
cd engine && ./build.sh          # → engine/out/，并把 designer/ 采到 out/designer/
OUT=<dir> ./build.sh             # 换落点
./build.sh TzsCli.Designer       # 只编一个
TZSCLI_INSTALL=<目录> ./build.sh  # 用别的设计器换掉仓库里那份（可选）
```

设计器默认取**本目录下的 `designer\`**（仓库里那份，跟着 clone 一起下来），所以不需要配任何东西。
`TZSCLI_INSTALL` 只是覆盖手段。

编完之后脚本会把 `designer\*.dll` 采到 `$OUT/designer/` —— 那正是引擎默认去找的位置，所以
`engine/out/tzs-server.exe` 开箱即跑，不需要环境变量。

## 进 tt 分包的是什么

```
tzs-server.exe        服务端（命名管道 / --stdio 两种模式）
tzs-cli.exe           客户端（独立可用；tt 自己实现了一份 Go 客户端）
TzsCli.dll            纯文本/zip 层，不反射
TzsCli.Designer.dll   反射管线 + 49 个函数
designer\             设计器的 28 个 .dll（就是本目录下 designer\ 那一份）
```

`build_portable.bat` 把它们采到 `<stage>\tzs\`（`designer\` 来自 `engine\designer\`，`TZSDESIGNER`
可覆盖），MSI 由 `heat.exe` 自动采集（不用改 `tt.wxs`）。

**`out/` 里还有十几个探测程序**（`Probe` / `Edit` / `AddField` / `RoundTrip` / `Test*` / `E2E`），
**不要 xcopy 整个目录**——只显式采那四个。

## 语言字典：曾经依赖反编译源码树，现在不依赖了

`Bootstrap.MergeLanguages()` 从 `INSTALL` 下的 `SpecDesigner*.dll` 的 `.g.resources` 里取每一份
`langs/zh-cn.xaml`。实测：`SpecDesignerCommon.dll` 里 88,737 字节、`SpecDesigner.Controls.dll`
里另有一份 450 字节——**两份都要**，字典是分散在多个程序集里的。

旧实现读的是设计器的**反编译源码树**（`D:\我的项目\T100设计器`），注释写着"安装的那份被编成
baml 资源、按路径取不到"。那句是对的，但停在了一步之前：它不在路径上，在**程序集里面**，
`ResourceReader` 按原始 XAML 交回来。所以源码树依赖已经删掉，字符串永远跟着用户装的版本。

找不到任何一份时**明确抛异常**——字典缺失会让 `FindResource("Message_...")` 返回 null，
然后在很远的地方以 `ArgumentNullException` 炸出来。

## 其余文档

| 文件 | 内容 |
|---|---|
| `SPEC.md` | 格式与契约的完整记录（§11.24 是冻结的 agent 契约，§11.22 是函数清单） |
| `HANDOFF.md` | 交接文档；§18 是最后一次关卡的结论与四个缺陷 |
| `TASKS.md` | 任务板与语料样本集 |
| `build.sh` 头部 | 两个库的分工、`LINK_SRC` 三值模式、纯净度守卫 |
