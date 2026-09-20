# tzc 模型落地对照表（六层 API ↔ 设计器源码 ↔ 不变量）

本文是 `tdev tzc` 实现的**逐条依据表**：每一处行为都指回反编译源码的 `file:line`，
或指回 `docs/T100设计器-README.md`（原仓库 README 的恢复副本）的章节。

反编译源码根：`D:\我的项目\T100设计器\`
前置文档：`docs/T100设计器-README.md`（§1–§7；本文用 `README §x.y` 指代）
设计契约：`tdev-tzc-design.md`（用 `指南 §x` 指代）

---

## 1. 六层 API 对照

| 指南 | 实现包 | 关键入口 | 设计器依据 |
|---|---|---|---|
| §5.1 Package 层 | `internal/pkgfile` | `Open` / `Package.Tap/Tgl/Full4gl/Plan/Build` | `TzpManager.cs:150-182`（类型由扩展名定）、`PackageManager.cs:70-100`（必需条目）、`:589-604`（`ver` 按名字取第一行）、`TzpManager.cs:395-404`（只比 Major/Minor） |
| §5.4 TAP 层 | `internal/tapfile` | `Parse` / `Rewrite` / `SetPointCDATA` / `AddPoint` / `MarkDeleted` / `SetSectionCDATA` / `SetRootAttr` | README §3.3（根元素真名 `add_points`、point/section 结构）、§3.6 第 6 点（混合换行）、§3.9（只用 CDATA 感知扫描器；属性「有则改、无则加」） |
| TGL 标记层 | `internal/tglfile` | `TrimEnd` / `FindSections` / `FindPlaceholders` / `FindAnchor` / `ReplaceAnchor` / `PatchSection` | `CodeEditorManager.cs:1694-1715`（四个正则原文）、`:314-352`（GenerateTGL 补丁格式）、`TzpManager.cs:453-457`（TGL 一次 `TrimEnd`） |
| §5.5 FGL | `internal/fgl` | `ParseOutline` / `ParseBlock` / `ParseFunction` / `EnvelopeKindFor` | README §4.7（`CodeEditor.FglAnalysis` 的用途与局限）、`FglParserQuickHelper.cs:18-23`、`CodeEditorManager.cs:1230-1253`（正文范围只用于算可编辑区） |
| §5.2 合成层 | `internal/synth` | `Synthesize` / `ResolvePoint` / `ResolveSection` | `CodeEditorManager.cs:1070-1080`（LoadContent 六步）、`:1142-1176`（锚点展开）、`:1179-1227`（占位符替换/造空点/记 TglTag）、`:1083-1139`（ProcessSections）、`ProgramInformation.cs:655-666`（Initial 造空点） |
| §5.3 围栏层 | `internal/fence` | `Render` / `Parse` / `Pair` | 指南 §4；本文 §4 的偏差表 |
| §5.5 验证层 | `internal/verify` | `Gate1` / `Gate2` / `Gate3` | README §3.8（I1–I15）、§5.1（105 包实测的五個坑） |
| §5.6 回写层 | `internal/split` | `Split` / `ApplyTglPatches`（在 cli） | `CodeEditorManager.cs:361-397`（SaveADPContent）、`:400-496`（SaveSectionContent 的 TglTag 折叠）、`:408`（`section_flag="Y"`）、`AddPointModel.cs:1086-1112`（ToXML）、`ProgramInformation.cs:96-108`（order = max+1） |
| §5.6 包容器 | `internal/pkgfile/zipraw.go` | `parseRawZip` / `rebuildZip` | **S7 真机实测**（不是源码）：设计器的包局部头 `flags=0x0000`（无数据描述符）、CRC/大小在局部头、extra 原样（`UT-time` + `ux-infozip`）。用 `archive/zip` 的 `Writer` 重写会变成 bit3 + 描述符 → 设计器报 `Data descriptor signature not found`。故写回改为字节级重建（清 bit3、打补丁、逐字节照抄），Zip64 字段按 AppNote 4.5.3 就地更新 |
| §5.7 审计层 | `internal/store` | `Create` / `Open` / `Lock` / `GitInit` / `GitCommit` / `AtomicWrite` | 指南 §3；红线 R6（禁止 `File.Delete`→`File.Create`：`PackageManager.cs:569-572` 是事故模板） |
| §2 命令行 | `internal/cli` | `Run` / `cmdExport` / `cmdStatus` / `cmdVerify` / `cmdApply` / `cmdUnlock` / `cmdRename` / `cmdNewfn` / `cmdSelftest` / `resolveWorkspaceDir` / `cmdInstall` / `cmdTzs` | 指南 §2、§6；`resolveWorkspaceDir` 是 CLI 层的便利规则：`<dir>` 省略即用当前目录（当前目录不是工作区则退出 5，不猜路径）。`cmdInstall`（`install skills` / `install path`）是工具自身的安装面，与包内容无关：skills 不内嵌，只复制 exe 旁边的 `skills/`；PATH 只写 HKCU 且保持原注册表类型（不用 setx，避免展开 `%VAR%` 与 1024 字符截断）。`cmdTzs` 是**表单包（`.tzs`/`.tzv`）入口**：`export` 是纯解压（`pkgfile.UnzipTo`，无围栏/无校验/无工作区、永不写回，代码包误用会指回 `tzc export`）；读写表单走 `call`/`fns`/`manifest`/`doctor`/`stop`/`reap`，由 `engine/` 那个 C# 引擎驱动（见 `docs/dev.md` 与 `skills/tt-dev/SKILL.md`） |
| §2 报错定位 | `internal/model/lines.go` + `verify.Finding` | `LineOf` / `LineAt` / `FirstDiffEOL` / `fillPositions` | 报错要给「哪一行」：字节偏移来自 Region 区间或首个差异字节（行尾等价，见 `NormalizeEOL` 口径），`Finding.File/Line/Snippet` 落到 JSON 与终端输出；TAP 层偏移用 `tapfile.PointAt` 说清落在哪个 point |
| §4.1 新函数模板 | `internal/cli/newfn_template.go` | `newFnHeaderTemplate` | `FunctionGenerator.cs:20-67` + 真实包 `desc="\n####…"`（语料 armt100.tap 的 `function.armt100_qrystr`）；顶部空行、80 个 `#` 的框、占位符逐字保留（日期/作者不自动填：红线 R7 无时钟） |

---

## 2. 权限判定链落地

### 2.1 点（`AddPointModel.IsEditable`，`AddPointModel.cs:691-747`）

| 门 | 条件 | 结果 | 实现位置 |
|---|---|---|---|
| G1 | `isIndFun && ind_fun != topind && name != "global.memo_industry"` | false | `synth.ResolvePoint` |
| G2 | `(topind=="sd" \|\| topind=="") && env=="s" && name=="global.memo_industry"` | false | 同上 |
| G3 | `!isStandard && cite_std=="Y"` | false | 同上 |
| G4 | `readonly=="Y"` | false | 同上 |
| G5a | `edit=="c" && env=="s"`，或 `edit=="c" && env=="c" && IsTopstdMode` | false | 同上（`edit=` 取自 **TGL 占位符行**，`:1203-1206`） |
| G5b | `edit=="s" && env=="c" && !IsTopstdMode && src!="c"` | false | 同上 |
| G6 | `IsTopstdMode`：`src=="c"` → 仅 `status==CREATE`；`src∈{s,m}` → 仅 `status==NULL` 或 (`modi_by_topstd=="Y"` 且 `MODIFY`) | 受限 | 同上 |
| G7 | `login_user=="topstd" && src=="c" && status!=CREATE` | false | 同上 |
| — | 自订定义点正文解析不出信封 | false（`structure-unparsable`） | `fgl.ParseBlock` + `EnvelopeKindFor` |

> `new="Y"` **不参与** `IsEditable`（只网关删除）；`IsSelfDefinition` **不参与**（只关重命名/改名）。

### 2.2 区段（`SectionModel.IsEditable`，`SectionModel.cs:101-133`）

| 门 | 条件 | 结果 | 备注 |
|---|---|---|---|
| 政策 | 三个锚点区段（`other_function`/`other_dialog`/`other_report`） | false | 我们**不复刻**设计器在 `type=="G" && section_flag=="Y"` 下对 `other_dialog` 的放行 —— 写锚点正文会把展开后的函数写进框架（红线 R4） |
| 政策 | 未开 `--allow-sec` | false（`sec-off`） | 对应设计器 SEC 模式未勾选 |
| S1 | `type=="G" && section_flag=="Y"` → 除 `other_function`/`other_report` 外可编辑 | true | 该特例在政策层之后才生效 |
| S2 | `IsReadOnly`（运行期派生：3 锚点硬编码 `:1127-1130` + TGL `readonly="y"` `:1131-1135`） | false | 永不落盘 |
| S3 | 区段属性 `readonly=="Y"` | false | |
| S4 | `IsTopstdMode`：`src=="c"` → false；`src∈{s,m}` → 见 `SectionModel.cs:124-129` | 受限 | |
| S5 | `login_user=="topstd" && src=="c"` | false | |
| 政策 | `--only` 部分导出 | false（`only-points`） | 设计指南 §2/§4 的 `--only` 约束 |

### 2.3 命名空间三层（README §3.7）

```
other.function / other.dialog / other.report   → 集合锚点（只在 TGL 里）
<裸名>（global.memo / input.a.page2.x / …）      → 框架预留插入点（TGL 通常有、TAP 通常没有）
function.* / dialog.* / report.*               → 自订定义点（TAP 有，由锚点注入）
```

- TGL 有占位符而 TAP 无点 → **造空点**（`ProgramInformation.cs:655-666`），常态，`I2a` info。
- TAP 有点而 TGL 无锚点 → `I2b` **error**（代码彻底不出现）。
- TAP 裸插入点无同名占位符 → 孤儿点，`I2c` warn，**必须原样透传**（`AddPointModel.Create` 不设 `IsLoaded`，
  `SaveADPContent` 过滤 `IsLoaded` 所以设计器跳过它但仍 `ToXML()` 写回）。
- **同名两次**：真实包里 `status="d"` 的 tombstone 与 `status="u"` 的 live 会同名共存，
  内容改写一律命中第一个**未删除**元素（`CodeEditorManager.cs:1188` 的过滤条件）。

---

## 3. 不变量 I1–I15 落地表

| # | 不变量 | 实现 | 等级 | 退出码 |
|---|---|---|---|---|
| I1 | `{<section>}`/`{</section>}` 配对、id 唯一、与 TAP `<section id>` 对应 | `tglfile.FindSections`（数量不等直接报错）+ `verify.Gate2` | error / 2（不配对） | 2 / 3 |
| I2a | TGL 占位符在 TAP 无对应点 | `synth` 造空点 + `Gate2` info | **info（正常）** | 0 |
| I2b | TAP 有 `function./dialog./report.` 点但 TGL 无锚点 | `Gate2`（`liveByKind` × `base.Anchors`） | **error** | 3 |
| I2c | TAP 裸插入点无同名占位符 | `Gate2` | warn（孤儿点，不丢） | 0（`--strict` → 3） |
| I3 | 区段正文不得含真实自订点的完整函数块 | `Gate2`，判据 = 拿 TAP 已知函数名匹配 `^\s*(public\|private)?\s*(FUNCTION\|DIALOG\|REPORT)\s+<名>\s*\(`，只扫区段**自身字节** | error | 3 |
| I4 | 自订点结构：先剥 `public/private` 再匹配定义头；`function.*` 前缀不保证是函数块 | `fgl.ParseBlock` + `synth.ResolvePoint` → `structure-unparsable`（只读+不可写） | warn | 0（写它 → 4） |
| I5 | `status ∈ {"", " ", c, u, d}`；删点置 `d` 且保留原内容 | `tapfile.MarkDeleted`（只改 status）+ `Gate2` | error | 3 |
| I7 | `ver` 存在且主次版本匹配、字节原样 | `pkgfile.Open`（`ParseVersion` 只比 Major/Minor） | error | 2 |
| I9 | `.tap2` 与 `.tap` 的 diff 一致 | 语料无 `.tap2`；只做透传（不生成） | — | — |
| I10 | `.4gl ≠ TGL + 展开点` | `Gate2` info（提醒不要试图同步） | **info（正常）** | 0 |
| I11 | UTF-8 无 BOM；CDATA 内无 `\a`；新内容无 `]]>` | `Gate2`（`I11` + `I11b`） | error | 3 |
| I13 | 未知条目 / `ver` 字节原样 | `pkgfile.Plan/Build`（按扩展名分派，其余透传）+ `Gate3` 逐条目 sha256 | error | 3 |
| I15 | 引用标准程序的包（条目基名 ≠ 根 `prog`） | `cli.cmdApply` 直接拒绝写回 | error | 2 |
| I-structure | 可编辑点的结构行未改 | `verify.Gate1`（结构行区间与可写区间**不相交**，逐字节比对） | error | 3 |

补充的机械检查（指南未列，但属于 A3 的实现）：

| 码 | 内容 |
|---|---|
| `gate1.fence-line` | 围栏行本身被改动（指南 §4 规则 2） |
| `gate1.outside-fence` | 围栏外字节（prefix/gap/suffix）拼接后不等 |
| `gate1.readonly-region` | 只读 Region 的**自身字节**被改（按配对比较，剔除子区间） |
| `gate1.eol-normalized` | info：受保护字节只有行尾差异（CRLF↔LF）→ 按等价处理。受保护字节从不写回包，故不影响安全性 |
| `gate1.region-count` | 基线总数 = 有基线配对的 Region + 被删除的 Region |
| `gate1.delete-*` | 删除授权：只有 `new="Y"` 且可编辑的点可删 |
| `gate1.append-*` | 追加必须在 `[APPEND]` 锚点内、类型匹配、`new="Y"`、名字不与基线重名 |

---

## 4. 与设计指南的偏差

见 `README.md` 的偏差表（D-1 … D-10），每条都给出源码依据。最要紧的一条：

**D-1**：指南 §4 规则 4 说新增点写 `status="c"`；实现写 **`status="u"`**。
依据：`AddPointModel.ToXML()`（`AddPointModel.cs:1090-1093`）在 `Status == CREATE` 时
`return null`，而 `ProgramInformation.AddPoint` getter 用 `_tap.Add(ToXML())` 组装 `.tap`。
写 `status="c"` 的点会在设计器下一次保存时**被静默丢弃**（数据丢失，不是格式错误）。

### 4.1 v2 新增的三条「文档 vs 源码」偏差

| # | 文档 | 源码 | 实现 |
|---|---|---|---|
| **DV-2** | §4.1 V6：描述块「每行 `#` 开头、**无空行**」 | `AddPointModel.cs:253`：`if (!Regex.IsMatch(line, "^\s*#") && line.Trim() != "")` → **空行合法** | V6 允许空行，只有非空行必须以 `\s*#` 开头；**并且**只对「本次改动过描述块」的点报 error，存量数据降为 warn（真实语料有历史遗留的不规范行，当 error 会拒掉真实包） |
| **DV-3** | §4.1 V5：「type=B 强制 PUBLIC；M/G/X/Z/Q 强制 PRIVATE」= error 级 | `FunctionInfoWindow.xaml.cs:143-192`：type→scope 映射是**新建点对话框的默认值**，整段被 `useDefaultScope` 门控，**不是对既有文本的硬约束**（真实语料 `aapt300(c).tzc` 就有 type=M 而函数声明 PUBLIC 的点） | 只在**本次发生 ScopeChange** 时给 **warn**；`newfn` 造新点时按该默认值填 scope |
| — | §4.1 事务等价测试要求「与设计器手工改名后保存的 `.tap` 做语义 diff」 | 设计器无法进 CI | 自检用**独立参考断言**墓碑对形状（旧名 `status="d"` + 原 CDATA 逐字节；新名 `status="u"` + 签名行已改），真机对拍列入 S7 人工项 |

### 4.2 v2 新增的机械约束（指南未列，属 A3 的实现）

| 码 | 内容 |
|---|---|
| `gate1.fence-line` | 围栏行**锚定部分**（kind/name/旗标/status/src/new/order…）不许改；`fn`/`scope`/`desc` 豁免，交由 V1/V5/V6 |
| `V1` | 围栏 `fn` == 签名行函数名；旧工作区（围栏无 `fn` 字段）则要求签名行未被改动 |
| `V2` | 旧名残留在可编辑区 → 3（列位置）；残留在只读区段 → **4** |
| `V3` | 新名冲突：现存 point / TGL 占位符裸名 / 本次会话重复目标 / 交换式改名 |
| `V7` | 目标必须是自订定义点（`function./dialog./report.`）且可编辑 |
| `gate1.eol-normalized` | info：受保护字节只有行尾差异（CRLF↔LF）→ 按等价处理 |

## 5. v2 状态机与事务的落地位置

| 概念 | 实现 | 关键依据 |
|---|---|---|
| `SectionState`（Locked/Unlocked） | `model.SectionState` + `.tdev/section-state` + `manifest.section` | 指南 §4.2；`ProgramInformation.cs:219-230` |
| 解锁授权闸门 | `synth.CheckUnlockPermission`（三分支 + 设计器原文常量） | `CodeEditorMainWindow.xaml.cs:557-588`；文案 `langs/zh-cn.xaml:448/641/688/689` |
| 解锁重渲染 | `cli.cmdUnlock`（只翻围栏旗标；**编辑文件与基线都翻**，绝不把待落盘的正文编辑吸进基线） | 指南 §4.2「唯一允许工具改写围栏行的操作」 |
| 区段写入联动 | `split`（TglTag 折叠 + `.tap`/`.tgl` 双写 + `section_flag="Y"`）+ apply 末尾 adzi520 提示 | `CodeEditorManager.cs:400-496`、`:408` |
| 结构事务检测 | `verify.DetectStructural`（复用 `fence.Pair.StructFieldChanged`） | 指南 §4.1 |
| 改名事务 | `split`：RemoveTombstones → RenamePoint → SetPointCDATA → SetPointAttr(status=u) → AddPoint(墓碑 status=d) | `ProgramInformation.cs:116-134`；`SetName` `AddPointModel.cs:810-830` |
| 新增点 | `cli.cmdNewfn`（插进 `[APPEND]` 锚点，不加前导空行）+ `split` 的 AddPoint | `AddPointModel.cs:1008-1060`（块模板：FUNCTION 恒写 scope；DIALOG/REPORT 的 PUBLIC 省略；REPORT 空正文注入默认 FORMAT） |
| apply 基线重算 | **以写出的新包为准重新 synthesize + render**，并同步刷新 `prog.full.4gl` | 改名会切换点身份（`function.旧名`→`function.新名`），只有从包重渲染基线才与包一致；同时顺带规范化围栏字段与行尾 |

---

## 6. 实测基线（本机，166 个真实包）

| 项 | 数据 |
|---|---|
| 语料 | `D:\t100_wrok_dir`：166 个 `.tzc`（hengshuo 107 / xiyuan 34 / wq 25），全部固定 4 条目 `4gl/tap/tgl/ver` |
| `ver` | 全部 `"1.0\n"`（4 字节，无 BOM） |
| 换行 | `.tgl` 纯 LF；`.tap` 混合（元素间 CRLF、CDATA 内 LF）；105 包实测 `tap` CRLF 289,763 / 独立 LF 1,108,074 |
| CDATA | 36,324 段；**0 段含 `]]>`**；**0 个 0x07**；**0 个 BOM** |
| `<point>` / `<section>` | 32,678 / 3,646 |
| 同名点（tombstone+live） | 4 处 |
| 区段 id 与点名同名 | 8 个包（如 `wssp00316.process`） |
| 退化点（`function.*` 但正文只有注释） | 1 处：`s_axmt500(s).tzc` / `function.memo_industry` |
