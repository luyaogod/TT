package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// 稳定错误码。**这是对外契约,不随版本改名** —— agent 靠 Code 分支,
// 人读 Message。加新码可以,改旧码的名字不行。
const (
	CodeUsage        = "USAGE"          // 用法/参数错
	CodeEnumInvalid  = "ENUM_INVALID"   // 枚举取值非法
	CodeEnvUnknown   = "ENV_UNKNOWN"    // 环境名不存在
	CodeConfigBroken = "CONFIG_BROKEN"  // 配置缺失/坏了
	CodeTableMissing = "TABLE_MISSING"  // 字典表不在库里(该数据族未同步)
	CodeConnFailed   = "CONNECT_FAILED" // 连不上库/服务器
	CodeSQLRejected  = "SQL_REJECTED"   // 语句没过白名单
	CodeQueryFailed  = "QUERY_FAILED"   // 查询本身失败
	CodeNotFound     = "NOT_FOUND"      // 查到了,但没有这条数据
)

// 退出码。0 与 1 **沿用 tt 既有语义**(1 = 用法/参数错),只新增分类码 ——
// 重新编号会把既有脚本的一次性判据改掉,而换来的只是对齐一个没有权威性的约定。
const (
	ExitOK      = 0
	ExitUsage   = 1 // 用法 / 参数校验错
	ExitSource  = 2 // 数据源错(连不上、配置坏)
	ExitMissing = 3 // 缺表 / 查不到该族数据
)

// Error 带稳定码与退出码的错误。它让 --json 下的错误也是机器可读的:
// 在此之前,所有 cobra 命令的错误都是 stderr 上的一句中文散文。
type Error struct {
	Code    string
	Exit    int
	Message string
	Hint    string // 修复建议;空则不打
	Meta    Meta   // 出错前已解析出的环境信息 —— 排查"哪个环境连不上"全靠它
}

func (e *Error) Error() string { return e.Message }

// ExitCode 让 tt 的退出码管线(internal/cli 的 exitCoder)认得出它。
func (e *Error) ExitCode() int { return e.Exit }

// Errorf 构造一个带码的错误。
func Errorf(code string, exit int, format string, a ...any) *Error {
	return &Error{Code: code, Exit: exit, Message: fmt.Sprintf(format, a...)}
}

// WithHint 补一句修复建议,返回自身便于链式。
func (e *Error) WithHint(h string) *Error { e.Hint = h; return e }

// WithMeta 补上出错时的环境信息,返回自身便于链式。
func (e *Error) WithMeta(m Meta) *Error { e.Meta = m; return e }

// errorEnvelope --json 时的错误形状。Meta 匿名嵌入 = 字段平铺,
// 与成功信封共用同一组键,"哪个环境出的错"在两种情况下写法一致。
type errorEnvelope struct {
	OK    bool   `json:"ok"`
	Code  string `json:"code"`
	Error string `json:"error"`
	Hint  string `json:"hint,omitempty"`
	Exit  int    `json:"exitCode"`
	Meta
}

// WriteError 把错误写成 JSON 信封。与成功信封共用 Meta 的字段,
// 所以"哪个环境出的错"在两种情况下是同一组键。
func WriteError(w io.Writer, e *Error) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(errorEnvelope{
		OK: false, Code: e.Code, Error: e.Message, Hint: e.Hint, Exit: e.Exit, Meta: e.Meta,
	})
}
