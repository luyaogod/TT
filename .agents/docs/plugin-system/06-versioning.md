# 版本与兼容

## 归属

| 对象 | 所有者 | 规则 |
|---|---|---|
| Host API 版本（`hostApi`） | TT | 语义化版本，当前 `1.0.0`；插件 manifest 声明范围（如 `>=1.0.0 <2.0.0`）；major 才可破坏，minor 只增 |
| manifest `schemaVersion` | TT | 整数，从 1 开始；未知的更高版本直接拒绝 |
| Protobuf 包 | TT | `tt.host.v1`；破坏性变更让 `tt.host.v2` 并存，不原地改 v1 |
| bridge 消息协议 | TT | 带版本字段的消息协议，当前 v1 |
| 插件 service 健康/就绪协议 | 插件 + TT | TT 定义的 ready JSON 行 + `GET /healthz` |
| 插件自身版本 | 插件 | 严格 SemVer |

## 兼容规则

- proto 字段只增不改：删字段、改字段语义、复用字段编号都算破坏性变更。
- 破坏性变更提升到新包（`tt.host.v2`），v1 保留；旧插件继续按 v1 工作，不被强制升级。
- Host API 的 minor 只做加法；插件 manifest 的范围决定它接受哪些 minor。
- `schemaVersion` 高于 TT 已知版本时拒绝，不猜测未知字段的含义。旧 TT 不应“尽力解析”新 manifest。
- 插件版本必须是严格 SemVer；范围、预发布等语义由 manifest 解析器统一处理。

## 生成物与闸门

- 生成的 Go/Connect 代码提交在 `internal/gen`，不手改。
- `make proto-check` 必须干净：buf lint + 重新生成 + `git diff -- internal/gen` 为空。改了 `.proto` 忘记重生成会在这里红。
- `make proto-breaking` 相对 `main` 检查公共契约兼容性；破坏性变更必须先升包版本，不能靠豁免蒙混。
- 生成工具链由 `make proto-tools` 安装，不在仓库里塞下载的 exe。

## 尚未实现

manifest 的 `hostApi` 范围解析、`schemaVersion` 拒绝路径、插件升级/回滚尚未实现。当前只有 Host API `1.0.0` 与 proto 生成物。
