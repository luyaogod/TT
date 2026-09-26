# internal/dev/pkgfile — 包的只读视图与字节级重建

`.tzc` / `.tzf` / `.tzx` 包（zip 容器）的**只读视图**，以及"写回"的唯一实现。
`.tzc` 管线的每一步都以这里拿到的 `*Package` 为输入。

## 契约

- **`Package` 不可变**。只读，没有 `SetPoint` / `SetTgl` 之类的接口；写回一律
  `Build(Rebuild{...})`，产出**新字节**（红线 R5）。
- **`Plan` 不改动任何字节**，只算逐条目写回计划 —— 这是 `--dry-run` 与 apply 报告的
  数据源（`EntryAction`：每条目的旧/新大小、旧/新 sha256、是否透传、压缩方法）。
- **红线两条**（`Plan` 的注释里写着）：条目集合不增不减、名字不变、顺序不变（R2）；
  `.4gl` 永不改写（R1，它是服务器产出）。

## 条目与版本

- 条目类型**由扩展名决定**（`.4gl` / `.tap` / `.tgl` / 版本文件 / 未知），不是猜内容。
- 必需条目的检查在这里；缺必需条目 → `FormatError`（退出码 2）。
- 版本文件按**名字**取第一条，并且只比较 **Major / Minor**（补丁号不同不算不匹配）。

## 为什么写回是"字节级重建"

设计器写出来的 zip 有自己的形态：**没有数据描述符**，CRC 与大小直接写在局部头里，
局部的 extra 字段是它自己的两段。因此：

- 用 `archive/zip` 的 `Writer` 重写会变成「bit3 + 数据描述符 + Go 自己的 extra」，
  包在设计器里打不开（它会报找不到数据描述符签名）。
- 实现的作法是 `zipraw.go` 的**字节级重建**：未改动的条目连压缩数据都逐字节照抄，
  只打补丁（清标志位、写真实 CRC/大小、更新偏移，必要时更新 Zip64 字段）。
- **零改动重建与原包逐字节相同**，这是本包最重要的判据。
- 压缩方法与修改时间沿用原条目的值，**不取时钟** —— 结果确定、可复现（红线 R7）。

## 导出入口

| 入口 | 用途 |
|---|---|
| `Open(path, OpenOptions)` / `Package` / `Entry` / `Kind` | 打开并检视包 |
| `Package.Tap() / Tgl() / Full4gl()` | 取三类具名条目 |
| `Package.Plan(Rebuild)` / `Build(Rebuild)` | 算计划 / 产出新字节 |
| `Rebuild{Tap, Tgl, Replace}` | 这次要替换什么（其余条目保持原字节） |
| `ParseVersion` / `Version` | 版本文件解析与比较 |
| `UnzipTo(src, dst, force)` | `.tzs export` 用的纯解压（另一条线） |
| `FormatError` / `IOError` / `ExitCodeError` | 错误类型，各自带退出码 |

## 判据

```bash
go test ./internal/dev/pkgfile -count=1
# TestRebuildIdentityByteExact   零改动重建与原包逐字节相同
# TestRawParseMatchesStdlib      自研解析与标准库在语料条目上逐字段一致
# TestRoundtripZeroChangeCorpus  语料级 roundtrip
# TestBuildRedlines              R1/R2 红线
```
退出码映射见 [../README.md](../README.md)。

## 细节去哪

- zip 形态与权限判定的实现细节 → `zipraw.go`、`unzip.go`
- 写回之后的事（拆分与落盘） → [../split/README.md](../split/README.md)、[../store/README.md](../store/README.md)
- 三方依据 → [`docs/T100设计器-README.md`](../../../docs/T100设计器-README.md) §3.1 容器层、§3.2 条目→加载器分派
