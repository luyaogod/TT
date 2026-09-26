# internal/dev/tglfile — `.tgl` 里的标记

`.tgl` 是设计器的**框架骨架**。本包只做一件事：在里面找到并操作三类标记。

| 标记 | 是什么 | 谁用 |
|---|---|---|
| **区段边界**（`{<section>}`） | 框架区段的起止，带 id | 合成时与 `.tap` 的 `<section>` 配对；改区段时两边必须一致 |
| **占位符**（裸名插入点） | 框架预留的插入位置 | 合成时与 `.tap` 的点配对；TGL 有占位符而 TAP 无点 → 造空点 |
| **集合锚点**（`other.function` / `other.dialog` / `other.report`） | 自订定义点被注入的位置 | 新增点只能插进这类锚点 |

## 入口

| 函数 | 作用 |
|---|---|
| `TrimEnd` | 读入时按设计器的习惯去掉尾部空白 |
| `FindSections` / `SectionBody` | 找区段并取出正文 |
| `FindPlaceholders` | 找占位符（含它是哪一类） |
| `FindAnchor(kind)` / `ReplaceAnchor(kind, repl)` | 找/替换集合锚点（新增自订点走这里） |
| `AnchorSectionID(prog, id)` | 判断某个区段 id 是不是锚点区段 |
| `PatchSection(tgl, id, content, eol)` | 按区段打补丁（改区段时同步 `.tgl`） |

## 关系

- 被 [synth](../synth/README.md)（占位符 ↔ 点配对）、[fence](../fence/README.md)、
  [verify](../verify/README.md)、[split](../split/README.md)（打补丁）、[cli](../cli/README.md) 使用。
- **改区段要写两处**：`.tap` 的 `<section>` 与 `.tgl` 的同一区段必须逐字节一致 ——
  这条约束由 [split](../split/README.md) 落地、由 `selftest` 里的对抗用例盯着。

## 判据

本包没有自己的测试文件（`go test ./internal/dev/tglfile` 报 `no test files`）。它的行为由
使用者覆盖：

```bash
go test ./internal/dev/...           # synth / fence / verify / split 的用例
./tt.exe dev tzc selftest            # 含"改区段后 .tap 与 .tgl 必须一致"的对抗用例
```

## 细节去哪

- 三类标记怎么与 TAP 配对、配对不上各是什么后果 → [../synth/README.md](../synth/README.md)、[../verify/README.md](../verify/README.md)
- 三方依据（四个正则原文与锚点展开逻辑的出处） → [`docs/T100设计器-README.md`](../../../docs/T100设计器-README.md) §3.3、§3.4、§3.5
