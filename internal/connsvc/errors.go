// Package connsvc 是 TT Core 的**进程内连接服务**：第一方模块直接函数调用，
// 第三方插件经 ConnectRPC 走同一份实现（见 internal/hostrpc）。
//
// 边界（对应 .agents/plan/tt-plugin-system-plan.md 的 §4.3 / §9）：
//
//   - TT 管：哪个调用方能用哪个环境、连接何时建立/关闭、会话归谁所有、审计元数据。
//   - TT 不管：调用方用连接做什么业务；不审查 SQL、不审查 SSH 命令内容。
//   - 秘密（口令/私钥）只在 Core 内解析并用于建连；对外的元数据一律脱敏。
//   - 内置只读路径（tt db sql / 字典直查 / dbsync）继续走 erpdb 的两层防线；
//     这里的通用会话是**另一条路径**，两者互不冒充（见 erpdb/session.go 的包注释）。
//
// 句柄规则：
//   - 句柄是不透明随机串，绑定 Caller（插件 ID + 实例 ID）、环境与连接类型；
//   - 拿别人的句柄操作一律报 SESSION_NOT_FOUND —— 不区分"不存在"和"不属于你"，
//     否则句柄可以当探测手段用；
//   - 连接断连、调用方停止、授权撤销、配置变更与 TT 关闭都必须走 Close/CloseAll。
package connsvc

import (
	"errors"
	"fmt"
	"strings"
)

// 稳定错误码。已发布的码不重命名；新增行为只新增码。
const (
	// CodeEnvUnknown 环境名不存在或没配（不含任何环境时也是它）。
	CodeEnvUnknown = "ENV_UNKNOWN"
	// CodeEnvNotAuthorized 调用方没有该环境的授权（不泄露环境是否存在）。
	CodeEnvNotAuthorized = "ENV_NOT_AUTHORIZED"
	// CodePermissionDenied 调用方缺少某项能力。
	CodePermissionDenied = "PERMISSION_DENIED"
	// CodeConfigBroken 配置文件读不出来/结构不对。
	CodeConfigBroken = "CONFIG_BROKEN"
	// CodeConfigChanged 配置在会话存活期间被改写过，会话已失效（重开即可）。
	CodeConfigChanged = "CONFIG_CHANGED"
	// CodeDBNoConfig 该环境没有配数据库连接。
	CodeDBNoConfig = "DB_NOT_CONFIGURED"
	// CodeDBConnectFailed 建连/测试失败（阶段见 Stage）。
	CodeDBConnectFailed = "DB_CONNECT_FAILED"
	// CodeSSHConnectFailed SSH 建连失败。
	CodeSSHConnectFailed = "SSH_CONNECT_FAILED"
	// CodeSessionNotFound 句柄不存在、已关闭，或不属于该调用方。
	CodeSessionNotFound = "SESSION_NOT_FOUND"
	// CodeSessionKind 句柄类型不符（拿 SSH 句柄查库）。
	CodeSessionKind = "SESSION_KIND_MISMATCH"
	// CodeSessionReadOnly 只读会话上试图执行写语句。
	CodeSessionReadOnly = "DB_SESSION_READONLY"
	// CodeQueryFailed 查询失败。
	CodeQueryFailed = "DB_QUERY_FAILED"
	// CodeExecFailed 执行失败。
	CodeExecFailed = "DB_EXEC_FAILED"
	// CodeSSHExecFailed SSH 命令执行失败（非零退出码也是它，输出里带远端输出）。
	CodeSSHExecFailed = "SSH_EXEC_FAILED"
	// CodeSSHFileFailed SSH 远端文件读写失败（含超过大小上限）。
	CodeSSHFileFailed = "SSH_FILE_FAILED"
	// CodeTimeout 超时（ctx 到期）。
	CodeTimeout = "TIMEOUT"
	// CodeAccountUnknown 指定的账号不在该环境的账号清单里。
	CodeAccountUnknown = "DB_ACCOUNT_UNKNOWN"
	// CodeEntUnknown 企业编号不在该环境的启用清单里（没同步 ENT 或编号写错）。
	CodeEntUnknown = "DB_ENT_UNKNOWN"
	// CodeEntNoAccount 企业存在但 gzou_t 里没有配可用账号。
	CodeEntNoAccount = "DB_ENT_NO_ACCOUNT"
	// CodeBatchFailed 批量执行失败（消息里带失败语句下标）。
	CodeBatchFailed = "DB_BATCH_FAILED"
	// CodeUnsupported 会话/驱动不支持请求的能力（如参数绑定、批量）。
	CodeUnsupported = "UNSUPPORTED"
	// CodeClosed 服务已关闭（进程正在退出）。
	CodeClosed = "SERVICE_CLOSED"
)

// 建连失败时的阶段标记，供插件/设置页定位"卡在哪一步"。
const (
	StageResolveConfig = "resolve_config"
	StageTunnel        = "tunnel"
	StageConnect       = "connect"
	StagePing          = "ping"
	StageQuery         = "query"
	StageExec          = "exec"
)

// Error 连接服务的稳定错误。插件只看 Code，不看 Message。
type Error struct {
	Code    string
	Message string
	Hint    string
	Env     string
	// Stage 失败阶段（TestDB/TestSSH 用；其余为空）。
	Stage string
	// Cause 底层错误（不对外序列化，只用于日志与 errors.Is/As）。
	Cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Env != "" {
		return fmt.Sprintf("[%s] 环境 %q: %s", e.Code, e.Env, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 让 errors.Is/As 能穿到底层错误。
func (e *Error) Unwrap() error { return e.Cause }

// WithHint 补一句"下一步怎么办"。
func (e *Error) WithHint(format string, a ...any) *Error {
	e.Hint = fmt.Sprintf(format, a...)
	return e
}

// errf 造一个带码的错误。
func errf(code, env, format string, a ...any) *Error {
	return &Error{Code: code, Env: env, Message: fmt.Sprintf(format, a...)}
}

// AsError 把任意错误识别服务错误（不是服务错误时返回 nil,false）。
func AsError(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// CodeOf 取错误的稳定码；非服务错误返回空串（调用方自行决定怎么归类）。
func CodeOf(err error) string {
	if e, ok := AsError(err); ok {
		return e.Code
	}
	return ""
}

// IsCode 判断错误是否某个稳定码。
func IsCode(err error, code string) bool { return CodeOf(err) == code }

// normalizeEnvName 只做去空格；**不做大小写折叠、不做前缀猜测** ——
// 环境名是不同的客户/库，模糊匹配等于静默换环境。
func normalizeEnvName(s string) string { return strings.TrimSpace(s) }
