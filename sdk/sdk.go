// Package sdk 是 TT 插件作者的公开入口。
//
// 一个插件就是一个普通的 Go 程序：
//
//	func main() {
//		sdk.Run(&myPlugin{})
//	}
//
//	type myPlugin struct{}
//
//	func (p *myPlugin) Start(ctx sdk.Context) error {
//		sess, err := ctx.DB().Open(ctx.Context(), "109tst", sdk.OpenOptions{Ent: 99})
//		...
//	}
//
// 这里只放**稳定契约**：服务接口、错误码、生命周期。实现由 TT 进程通过
// Host API 提供（SDK 不 import TT 的 internal 包，也不持有数据库凭据）。
//
// 版本：与 TT 主版本一起演进；新增方法/服务是兼容，改语义或删除要升主版本。
package sdk

import (
	"context"
	"encoding/json"
	"time"
)

// Plugin 一个插件的生命周期。两条都由 TT 的宿主进程调用。
type Plugin interface {
	// Start 插件进程启动后调用一次；返回错误会让 TT 把插件标记为启动失败。
	Start(ctx Context) error
	// Stop 插件被停止前调用一次；用于关闭自己打开的资源。ctx 可能已取消。
	Stop(ctx Context) error
}

// Run 启动插件并阻塞到 TT 要求退出（实现见 run.go）。

// Context 宿主注入给插件的一切。插件按需取用；取不到的服务返回 nil。
type Context interface {
	// Context 插件主生命周期；TT 要求退出时取消。
	Context() context.Context
	// PluginID / Version 本插件的标识（来自清单）。
	PluginID() string
	Version() string
	// DataDir 持久数据目录（升级/卸载默认保留）。
	DataDir() string

	Config() ConfigService
	DB() DBService
	SSH() SSHService
	FS() FSService
	Jobs() JobService
	Commands() CommandService
	Events() EventService
	Log() Logger
}

// Env 一个已配置环境的脱敏描述。
type Env struct {
	Name   string
	DBType string
	// DBAddr 展示用 host:port/service，不含凭据。
	DBAddr string
}

// Ent 企业编号 → 账号的映射（不含口令）。
type Ent struct {
	Ent         int
	Account     string
	Placeholder bool
}

// ConfigService 宿主配置的只读投影 + 本插件自己的配置。
type ConfigService interface {
	Envs() []Env
	// Ents 该环境在 gzou_t 里启用的企业；nil = 尚未同步企业目录。
	Ents(env string) []Ent
	// Values 本插件清单里 config 段的当前取值（JSON）。
	Values() json.RawMessage
}

// OpenOptions 打开一条数据库会话。
type OpenOptions struct {
	// Ent 企业编号：>0 时由宿主按企业目录解析账号；0 用 Account/默认账号。
	Ent int
	// Account 指定账号名；空 = 该环境的默认账号。
	Account string
	// ReadOnly 建立库侧只读会话（插件自己的选择，不是宿主强加）。
	ReadOnly bool
}

// DBService 数据库资源。凭据只在 TT 手里，插件拿到的是会话句柄。
type DBService interface {
	Open(ctx context.Context, env string, opt OpenOptions) (DBSession, error)
}

// DBSession 一条有所有权的数据库会话。Close 由插件负责；插件卸载时宿主兜底关闭。
type DBSession interface {
	Query(ctx context.Context, sql string, params ...string) (Rows, error)
	Exec(ctx context.Context, sql string, params ...string) (int64, error)
	// Batch 一次提交多条语句；atomic=true 时整批一个事务，失败全部回滚。
	Batch(ctx context.Context, stmts []Statement, atomic bool) (BatchResult, error)
	// Begin 开启显式事务；事务句柄跨多次调用保持，Commit/Rollback 结束。
	Begin(ctx context.Context) (Tx, error)
	Close() error
}

// Statement 一条带参数的语句。占位符统一写 `?`，方言由宿主处理。
type Statement struct {
	SQL    string
	Params []string
}

// BatchResult 批量执行结果。
type BatchResult struct {
	RowsAffected int64
	// FailedIndex 失败语句下标（0 基）；-1 = 全部成功。
	FailedIndex int
}

// Rows 查询结果。值一律为字符串（与 T100 字典/查询工具一致）。
type Rows struct {
	Columns   []string
	Values    [][]string
	Truncated bool
}

// Tx 一个显式事务。
type Tx interface {
	Exec(ctx context.Context, sql string, params ...string) (int64, error)
	Query(ctx context.Context, sql string, params ...string) (Rows, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// SSHService SSH 资源。
type SSHService interface {
	Open(ctx context.Context, env string) (SSHSession, error)
}

// SSHSession 一条 SSH 会话。非零退出码是正常结果，不是错误。
type SSHSession interface {
	Exec(ctx context.Context, command string, timeout time.Duration) (SSHOutput, error)
	// Put / Get 走 SFTP；实现随 P2 落地。
	Put(ctx context.Context, localPath, remotePath string) error
	Get(ctx context.Context, remotePath, localPath string) error
	Close() error
}

// SSHOutput 一条 SSH 命令的结果。
type SSHOutput struct {
	Output   string
	ExitCode int
	TimedOut bool
}

// FSService 插件自己的文件空间（便捷封装；插件也可以直接用 os）。
type FSService interface {
	DataDir() string
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
}

// JobService 长任务：进度、取消、日志。
type JobService interface {
	// Run 起一个任务；fn 里通过 Reporter 报进度。返回任务 ID。
	Run(ctx context.Context, name string, fn func(context.Context, Reporter) error) (string, error)
}

// Reporter 任务进度上报。
type Reporter interface {
	Progress(percent int, message string)
	Log(message string)
}

// CommandService 注册 CLI 子命令（tt <namespace> …）。
type CommandService interface {
	Register(name, summary string, handler func(ctx context.Context, args []string) error)
}

// EventService 宿主事件。
type EventService interface {
	On(name string, handler func(payload json.RawMessage))
}

// Logger 结构化日志，落在 logs/<id>/。
type Logger interface {
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

// ---- 错误 ----

// Code 稳定错误码。插件只判 Code，不解析 Message。
type Code string

const (
	// 连接与会话。
	CodeEnvUnknown       Code = "ENV_UNKNOWN"
	CodeEnvNotAuthorized Code = "ENV_NOT_AUTHORIZED"
	CodeDBConnectFailed  Code = "DB_CONNECT_FAILED"
	CodeDBNoConfig       Code = "DB_NOT_CONFIGURED"
	CodeAccountUnknown   Code = "DB_ACCOUNT_UNKNOWN"
	CodeEntUnknown       Code = "DB_ENT_UNKNOWN"
	CodeEntNoAccount     Code = "DB_ENT_NO_ACCOUNT"
	CodeSessionNotFound  Code = "SESSION_NOT_FOUND"
	CodeSessionReadOnly  Code = "DB_SESSION_READONLY"
	CodeQueryFailed      Code = "DB_QUERY_FAILED"
	CodeExecFailed       Code = "DB_EXEC_FAILED"
	CodeBatchFailed      Code = "DB_BATCH_FAILED"
	CodeTimeout          Code = "TIMEOUT"
	CodeSSHConnectFailed Code = "SSH_CONNECT_FAILED"
	CodeSSHExecFailed    Code = "SSH_EXEC_FAILED"
	CodePermissionDenied Code = "PERMISSION_DENIED"
	CodeServiceClosed    Code = "SERVICE_CLOSED"
	CodeConfigBroken     Code = "CONFIG_BROKEN"
	CodeConfigChanged    Code = "CONFIG_CHANGED"
	CodeSessionKind      Code = "SESSION_KIND_MISMATCH"
	CodeUnimplemented    Code = "UNIMPLEMENTED"
	CodeInvalidArgument  Code = "INVALID_ARGUMENT"
	CodeInternal         Code = "INTERNAL"
)

// Error 带稳定码的错误。
type Error struct {
	Code    Code
	Message string
	Hint    string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Hint != "" {
		return string(e.Code) + ": " + e.Message + "（" + e.Hint + "）"
	}
	return string(e.Code) + ": " + e.Message
}

// CodeOf 取错误的稳定码；不是 *Error 时返回空串。
func CodeOf(err error) Code {
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return ""
}
