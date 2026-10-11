// Package fep 承载 tt fep 命令组（前端协议线：读写运行中的 T100 作业界面）。
//
// 本组是 internal/fep 的命令面：内核（帧编解码、AUI 模型、字段名解析、语义动作）
// 不碰 socket、不碰窗口，全部输入是录制下来的字节流；本组负责把"录制件/会话登记"
// 这类输入交进去，并把结果打到输出出口上。
//
// 退出码见 internal/cli/README.md（唯一权威）；本文件只做实现。
package fep

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/fep"
	"tt/internal/output"
)

// 退出码。与其它命令组的数字含义**不通用**：调用方必须连子命令一起判。
const (
	exitOK        = 0
	exitFailed    = 1 // 运行失败：IO / 网络 / 环境
	exitUsage     = 2 // 用法错：未知子命令、参数不成立
	exitNoSession = 3 // 没有会话：未登记，或登记已经过期
	exitState     = 4 // 会话状态不符：已有会话、端口被占、GDC 未就绪
	exitTarget    = 5 // 目标不存在：连接号 / 窗口 / 对象
	exitPreflight = 6 // 环境体检不过：权限不一致
)

// Group 是 tt fep 命令组。
var Group = &cobra.Command{
	Use:   "fep",
	Short: "T100 前端协议：读写运行中的作业界面（无感）",
	Long: `与运行中的 T100 作业界面打交道：读它的界面结构（AUI）、发它的用户事件。

内核只认协议，不看屏幕：字段名从协议里解析（注释键 → 表.字段名后缀 → 自身文本 → 受限继承），
解析不出来就标"未解析"，绝不猜。动作按语义抽取（点击 / 输入），只是路过的栏位不算一步。

退出码：
  0  成功
  1  运行失败（IO / 网络 / 环境）
  2  用法错（未知子命令、参数不成立）
  3  没有会话（未登记，或登记已经过期）
  4  会话状态不符（已有会话、端口被占、GDC 未就绪）
  5  目标不存在（连接号 / 窗口 / 对象）
  6  环境体检不过（权限不一致等）

录制件（--from）就是代理逐连接落下的原始字节（.log）：本组全部命令都能离线跑，
只要给一份录制件，不需要真机在场。`,

	// 未知子命令必须非 0：打错命令读成成功，是调用方最难自己发现的失败形态。
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return fail(exitUsage, "未知子命令 %q（试 tt fep --help）", args[0])
	},
}

// Register 把本命令组挂到根命令上。
func Register(root *cobra.Command) { root.AddCommand(Group) }

func init() {
	// 开关解析错也退 2：cobra 默认给 1，而这条线的表里 2 = 用法错
	Group.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return fail(exitUsage, "%v", err)
	})

	Group.AddCommand(sessionCmd, uiCmd, recCmd, doCmd, waitCmd, srcCmd, jobCmd, feCmd)
	sessionCmd.AddCommand(sessionPsCmd, sessionCloseCmd)
	uiCmd.AddCommand(uiSnapshotCmd, uiFindCmd, uiTreeCmd, uiRolesCmd, uiStateCmd)
	recCmd.AddCommand(recStepsCmd)
}

// fail 造一个带退出码的错误。根命令通过 ExitCode() 取码（见 internal/cli/root.go）。
func fail(code int, format string, args ...any) error {
	c := output.CodeUsage
	switch code {
	case exitNoSession:
		c = output.CodeNoSession
	case exitState:
		c = output.CodeSessionState
	case exitTarget:
		c = output.CodeTargetMissing
	case exitPreflight:
		c = output.CodePreflight
	case exitFailed:
		c = output.CodeQueryFailed
	}
	return &output.Error{Code: c, Exit: code, Message: fmt.Sprintf(format, args...)}
}

// ---------------------------------------------------------------------------
// 共同件：会话登记与录制件

// registry 打开会话登记（落点由配置所在目录决定，不自行拼路径）。
func registry() (*fep.Registry, string, error) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return nil, "", err
	}
	loc := config.LocationsAt(cfgPath)
	dir := loc.DataDir()
	if dir == "" {
		return nil, "", fail(exitPreflight, "配置路径没解析出数据目录，先跑 tt config init 或指定 --config")
	}
	return fep.NewRegistry(dir), dir, nil
}

// cacheDir 返回 fep 的缓存落点（录制件/截图都在这里）。
func cacheDir() (string, error) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return "", err
	}
	dir := config.LocationsAt(cfgPath).CacheDir("fep")
	if dir == "" {
		return "", fail(exitPreflight, "缓存目录没解析出来（检查配置路径）")
	}
	return dir, nil
}

// recordingFiles 按后缀挑录制件：--from 可以是单个 .log，也可以是存放它的目录。
//
// 目录模式默认只取最新的一份：一个目录里往往躺着好几轮录制，混着读会把不同会话
// 的对象、事件堆进同一个模型（实测过：动作数将近翻倍，名字全对不上）。要合起来读用 --all。
func recordingFiles(from, suffix string, all bool) ([]string, error) {
	if from == "" {
		dir, err := cacheDir()
		if err != nil {
			return nil, err
		}
		from = dir
	}
	st, err := os.Stat(from)
	if err != nil {
		return nil, fail(exitFailed, "读不到 %s: %v", from, err)
	}
	if !st.IsDir() {
		return []string{from}, nil
	}
	files, err := fep.RecordingFiles(from, suffix) // 按时间倒序
	if err != nil {
		return nil, fail(exitFailed, "读不到 %s: %v", from, err)
	}
	if len(files) == 0 {
		return nil, fail(exitFailed, "%s 下没有 *%s（先跑 tt fep session open 录一轮）", from, suffix)
	}
	if !all && len(files) > 1 {
		files = files[:1]
	}
	return files, nil
}

// readRecording 把一份录制件读成一串"消息体"（两种形态都吃，见 internal/fep/record.go）。
func readRecording(path string) ([]string, error) {
	msgs, err := fep.ReadRecording(path)
	if err != nil {
		return nil, fail(exitFailed, "读 %s: %v", filepath.Base(path), err)
	}
	return msgs, nil
}

// pairedRecording 把一侧录制件换成同一连接的另一侧。
//
// 同一轮会话里，每个窗口连接都有一对文件（rt2gdc / gdc2rt）。拿事件文件去配别的
// 连接（或别的会话）的 AUI，对象编号对不上：动作全变"名字未解析"。
func pairedRecording(from, want, other string) string {
	if from == "" || !strings.HasSuffix(from, other) {
		return from
	}
	p := strings.TrimSuffix(from, other) + want
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return from
}

// loadModel 把 AUI（运行期→前端方向）喂进模型，返回模型、字段名解析器与消息数。
func loadModel(from string) (*fep.Model, *fep.Resolver, int, error) {
	files, err := recordingFiles(pairedRecording(from, "-rt2gdc.log", "-gdc2rt.log"), "-rt2gdc.log", recAll)
	if err != nil {
		return nil, nil, 0, err
	}
	m := fep.NewModel()
	msgs := 0
	for _, f := range files {
		bodies, err := readRecording(f)
		if err != nil {
			return nil, nil, 0, err
		}
		for _, body := range bodies {
			m.ApplyMessage(body)
			msgs++
		}
	}
	// 窗口归属只向下传播一层：运行期标定过的父对象，其子对象也属于同一个窗口。
	m.Propagate()
	return m, fep.NewResolver(m), msgs, nil
}

// loadEvents 把前端事件（前端→运行期方向）读成事件文本列表。
func loadEvents(from string) ([]string, error) {
	files, err := recordingFiles(pairedRecording(from, "-gdc2rt.log", "-rt2gdc.log"), "-gdc2rt.log", recAll)
	if err != nil {
		return nil, err
	}
	var events []string
	for _, f := range files {
		bodies, err := readRecording(f)
		if err != nil {
			return nil, err
		}
		for _, body := range bodies {
			events = append(events, fep.SplitEvents(body)...)
		}
	}
	return events, nil
}

// printJSONOrText 按 --json 决定走输出出口还是人读的文本。
func printJSONOrText(v any, text func()) error {
	if common.OutputFormat() == output.FormatJSON {
		return common.PrintJSON(v)
	}
	text()
	return nil
}

// exactArgs / noArgs 是本组自己的参数校验：cobra 自带的参数错退出码是 1，
// 而这张线的表里 2 = 用法错。参数错是最常见的调用错误，读成"运行失败"会误导排障。
func exactArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return fail(exitUsage, "%s 需要 %d 个参数，收到 %d 个", cmd.CommandPath(), n, len(args))
		}
		return nil
	}
}

var noArgs = exactArgs(0)
