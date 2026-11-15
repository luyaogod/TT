# Host API 公共契约（proto/tt/host/v1）

第三方插件只依赖这里与 `internal/gen` 的生成物；`internal/*` 的 Go 包不是公共 API。

## 目录

| 路径 | 作用 |
|---|---|
| `tt/host/v1/common.proto` | `ErrorInfo`、`Capabilities`、`Feature`、`Profile` |
| `tt/host/v1/host.proto` | `HostService.GetCapabilities` |
| `tt/host/v1/profiles.proto` | `ProfilesService.List` |
| `tt/host/v1/connections.proto` | `ConnectionsService.TestDB/TestSSH/OpenDB/OpenSSH/Close` |
| `tt/host/v1/dbsession.proto` | `DBSessionService.Query/Exec` |
| `tt/host/v1/sshsession.proto` | `SSHSessionService.Exec` |
| `baseline/v1.json` | 已发布契约的 FileDescriptorSet 快照，`make proto-breaking` 的对照基线 |

## 改契约的规矩

1. 只增不改：新增字段给新编号；删字段/改语义必须升主版本（出现 `tt.host.v2`，v1 原样保留）。
2. 改完 `.proto` 必须跑 `make proto-gen` 并一起提交生成物；`make proto-check` 会抓"改了没重生成"。
3. `make proto-breaking` 对照 `baseline/v1.json`。基线**只在发版并有意做增量扩充后**刷新：
   `buf build -o proto/baseline/v1.json --as-file-descriptor-set`。
4. 错误一律带 `ErrorInfo` 详情；插件按 `code` 分支，不解析 `message`。

## 本机怎么装工具链

`make proto-tools`（`go install` 装 buf 与两个 protoc 插件）。**不要**下 buf/protoc 的 exe 包：
企业终端防护策略会让下载来的可执行文件无法写盘，症状是 `Access is denied` 而看不出原因。
Buf 的缓存目录由 Makefile 固定到仓库内（`.tools/buf-cache`），同样是为了绕开只读的用户目录。
