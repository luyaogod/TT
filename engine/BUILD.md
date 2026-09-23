# engine/ — `.tzs` 表单设计器引擎（C#）

这个目录是 **tt 的引擎**，不是 tt 的一部分：它读写的 `.tzs` 格式是 T100 设计器私有的，
而这里的代码**不实现那个格式**——它反射调用**已安装的设计器自己的程序集**。我们这部分
只有 200 KB（`TzsCli.dll` 26 KB + `TzsCli.Designer.dll` 180 KB），设计器那部分是 10.5 MB。

`tt dev tzs` 通过 JSON-RPC 驱动它（见 `internal/dev/tzs/`）。

## 为什么必须单独构建

**1. 它不属于 Go 的构建链。** 用 `csc.exe`（Framework64 v4.0.30319，**C# 5**——没有模式匹配、
没有 `nameof`、没有字符串插值）编译，引用 GAC 里的 WPF 程序集。把它塞进 `.bat` 等于用 cmd.exe
重写一份已经踩平过 CS0433 的脚本，得到两份会漂移的实现。`build_portable.bat` 只**采集**产物。

**2. 设计器程序集：构建期向机器要一份，运行期自带一份。**

```
构建期  INSTALL = TZSCLI_INSTALL 或 D:\APPS\T100设计器_1.0.0.251_免安装
运行期  TZSCLI_INSTALL（开发覆盖）或 <引擎自己的目录>\designer（随包分发的那份）
```

- **构建期**：`-r:$INSTALL\Newtonsoft.Json.dll` —— 一个编译期引用，构建机上得有；
- **运行期**：`Assembly.LoadFrom` 设计器的 `SpecDesignerCommon.dll` / `FormEditor.dll` /
  `UndoRedoFramework.dll`，以及语言字典（见下）。这些**由发行包自带**。

**为什么运行期要自带。** 同一份 tt 装在两台机器上，如果设计器目录来自各自的配置，两边跑的
就是两版设计器 —— 同一个 `.tzs` 在两台机器上行为不同，而报错里看不出来。把版本钉进包里之后，
"运行环境一致"是**分发这件事本身**保证的，不需要谁去对齐配置。

代价也要说清楚：**设计器的版本从此由打包时采进去的那份决定**。换版本 = 重打包，
`build_portable.bat` 的 `TZSDESIGNER`（或 `TZSCLI_INSTALL`）指到新目录，脚本只采 `*.dll`。

所以 `config.json` 里**没有** `tzs.installDir` 这个键。`TZSCLI_INSTALL` 保留为开发/构建期的
逃生口 —— `engine/out/` 不是包，本地跑的引擎和 `test/` 下的探测程序靠它指向机器上装的那份。
设计器本身仍然不入库（`.gitignore` 的 `*.dll`），它只是**打包输入**。

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
cd engine && ./build.sh          # → engine/out/，15 个单元
OUT=<dir> ./build.sh             # 换落点
./build.sh TzsCli.Designer       # 只编一个
```

跨机器时先设好设计器目录：

```bash
TZSCLI_INSTALL='D:\APPS\某版本设计器' ./build.sh
```

## 进 tt 分包的是哪四个文件

```
tzs-server.exe        服务端（命名管道 / --stdio 两种模式）
tzs-cli.exe           客户端（独立可用；tt 自己实现了一份 Go 客户端）
TzsCli.dll            纯文本/zip 层，不反射
TzsCli.Designer.dll   反射管线 + 49 个函数
```

`build_portable.bat` 把它们采到 `<stage>\tzs\`，MSI 由 `heat.exe` 自动采集（不用改 `tt.wxs`）。

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
