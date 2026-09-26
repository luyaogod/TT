# internal/dev/synth — 合成与权限判定

把包**合成**成一份可编辑文档（`Package → Document`），以及决定"哪一块能改"的两条权限判定链。

合成逐条复刻设计器的装载流程：锚点展开 → 占位符替换 / 造空点 / 记标记 → 区段按序号配对并
强制只读。合成出来的 `Document` 是 [fence](../fence/README.md) 渲染围栏的输入。

**权限判定的两条链是本包最重要的东西** —— 机械重写最大的风险不是写错字节，而是**写了一个
设计器本来不许改的地方**。

## 点：G1–G7（`ResolvePoint`）

| 门 | 条件 | 结果 |
|---|---|---|
| G1 | 独立功能程序，且行业别不匹配 | 不可编辑 |
| G2 | 标准环境 + 行业别 `sd`/空 + 行业 memo 点 | 不可编辑 |
| G3 | 非标准件且 `cite_std="Y"` | 不可编辑 |
| G4 | `readonly="Y"` | 不可编辑 |
| G5 | 插入点 `edit=` 与环境的组合（`edit=c` 遇上标准环境 / `edit=s` 遇上客制环境…） | 不可编辑 |
| G6 | topstd 编辑模式：按 `src` 与 `status` 收窄到只有 `CREATE` / `NULL` / `MODIFY` 可编辑 | 受限 |
| G7 | 登录用户是 `topstd` 且 `src="c"` 且状态不是新建 | 不可编辑 |

外加一条**本工具的政策**：自订定义点的正文解析不出块信封 → 一律不写回（设计器装载会抛异常）。

`new="Y"` 不参与可编辑性判定（它只关删除授权）；是否自订定义点也不参与（只关改名）。

## 区段：政策层 + S1–S5（`ResolveSection`）

前两层是本工具自己的政策，先于设计器链生效：

1. **三个集合锚点区段的正文永不写** —— 那是注入的点块，写它等于把展开后的函数写进框架（红线 R4）。
2. **框架未解锁 → 一律只读**（先 `tt dev tzc unlock`）。

然后是设计器链 S1–S5（`type="G"` 且已解开时的例外、运行期只读、`readonly` 属性、
topstd 模式、登录用户门禁）。

> **解锁改变的是门，不是所有房间。** 解锁之后仍然逐条走设计器链 —— 锚点区段、TGL 标了
> `readonly="Y"` 的、topstd 模式下的 `src` 规则，一个都不放松。

## 解锁闸门（`CheckUnlockPermission`）

逐条重放设计器那个勾选框的判定，三个分支：

| 情形 | 判定 |
|---|---|
| 包本来就是解开态（根标志已是 `Y`） | 允许（设计器对这类包不设拦截） |
| 标准环境 + 非 topstd：无授权事实 | **拒绝**，附授权原文 |
| 标准环境 + 非 topstd：有授权事实 | 允许，但需二次确认 |
| 客制环境或 topstd | 允许，但需二次确认（打印代价警告） |

**没有绕过**：不提供强制开关，不改环境与登录用户，也不写"假装有授权"的标志 ——
那是服务器侧的权限域，本地工具自我授权等于伪造权限。是否解锁必须由人决定。

## 入口

| 入口 | 作用 |
|---|---|
| `Synthesize(pkg, Options)` | 合成文档 |
| `ResolvePoint` / `ResolveSection` | 判定单个点 / 区段的权限，返回"可否"与**拒绝原因码** |
| `CheckUnlockPermission(env, yes)` / `UnlockDecision` | 解锁闸门 |
| `DeniedError` / `SynthesisError` | 拒绝类错误的退出码是 4 |

## 判据

```bash
go test ./internal/dev/synth -count=1
# TestResolvePointTruthTable    点的权限链真值表
# TestResolveSectionTruthTable  区段的权限链真值表
```

## 细节去哪

- 围栏怎么把权限渲染成可读的标注 → [../fence/README.md](../fence/README.md)
- 拒绝原因怎么翻成中文 → [../model/README.md](../model/README.md)（`DenyReasonText`）
- 三方依据 → [`docs/T100设计器-README.md`](../../../docs/T100设计器-README.md) §3.4、§3.7、§4.3，
  以及反编译源码的 `CodeEditorManager.cs` / `AddPointModel.cs` / `SectionModel.cs`（源码树不在本仓库）
