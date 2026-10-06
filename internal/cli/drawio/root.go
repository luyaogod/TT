package drawio

import (
	"fmt"
	"os"
	"path/filepath"
)

// 本命令组的退出码。**这条线自己定，别拿别条线的数字套** ——
// 与 debug / dev / dict 一样，同一个数字在各线含义不同（权威表在 internal/cli/README.md）。
const (
	exitOK        = 0
	exitUsage     = 1 // 用法或运行失败
	exitInput     = 2 // 输入错：形状源校验失败、spec 不合法、包读不出
	exitSelfCheck = 3 // 产物自查未通过 —— 工具自己的 bug，与"你的输入错"分开
)

// exitCode 是带退出码的错误，由根命令的 exitCodeOf 解出（与 internal/cli/dev 同一形态）。
type exitCode int

func (e exitCode) Error() string { return fmt.Sprintf("退出码 %d", int(e)) }

// ExitCode 实现根命令期望的退出码接口。
func (e exitCode) ExitCode() int { return int(e) }

// inputErr 造一个「输入错」（退 2）：形状源坏了、spec 不合法、包读不出来。
//
// 与"用法错"分开是给脚本用的：`1` 是"你命令敲错了"，`2` 是"命令没错，喂进来的东西不对"。
func inputErr(format string, a ...any) error {
	return &codedError{code: exitInput, err: fmt.Errorf(format, a...)}
}

// selfCheckErr 造一个「产物自查未通过」（退 3）。
//
// 这一类**不是调用方的问题** —— 是工具自己算错了（id 撞了、parent 悬空、几何里出现
// NaN）。单列一个码，下游脚本才能区分"我的 spec 写错了"与"该给 tt 提个 bug"。
func selfCheckErr(format string, a ...any) error {
	return &codedError{code: exitSelfCheck, err: fmt.Errorf(format, a...)}
}

// codedError 把一句话错误挂上退出码。
type codedError struct {
	code int
	err  error
}

func (e *codedError) Error() string { return e.err.Error() }
func (e *codedError) Unwrap() error { return e.err }
func (e *codedError) ExitCode() int { return e.code }

// outDirName 是缺省落点的目录名。
const outDirName = "dist"

// libraryDir 决定形状库导出到哪：**缺省 `<当前目录>/dist`**，`-o` 覆盖。
//
// 为什么不落数据目录：库文件是**给人拿走的**（拖进 drawio 或放进项目），不是
// 可再生的缓存。放 cwd 下的 dist/ 有一个好处 —— 在哪儿跑就在哪儿拿到，不必
// 先 `tt config path` 查一遍路径。它与仓库的构建产物目录同名同性质，也同样
// **不入库**（`.gitignore` 罩着 `/dist/`）。
func libraryDir(flag string) (string, error) {
	if flag != "" {
		return filepath.Abs(flag)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("取不到当前目录: %w", err)
	}
	return filepath.Join(cwd, outDirName), nil
}
