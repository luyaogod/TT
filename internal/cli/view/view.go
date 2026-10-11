// Package view 承载 tt view 命令组（窗口与图像：抓真图、标注落点）。
//
// 它只做"看"：不碰键鼠、不改窗口样式、不置顶。抓图后端与两项环境要求
// （DPI 感知必须最先声明、枚举只按 pid）见 internal/view 的包注释。
package view

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/fep"
	"tt/internal/output"
)

// 退出码（表在 internal/cli/README.md）。
const (
	exitOK      = 0
	exitFailed  = 1 // 运行失败：IO / 抓图失败
	exitUsage   = 2 // 用法错
	exitTarget  = 3 // 找不到目标窗口
	exitQuality = 4 // 产物自查未通过（全白图、黑边、尺寸异常）
)

var Group = &cobra.Command{
	Use:   "view",
	Short: "窗口与图像：抓真图、查尺寸、标注落点",
	Long: `抓运行时窗口的**真图**（QT 渲染的客户所见画面），并回答"这张图能不能用"。

抓图后端：PrintWindow + PW_RENDERFULLCONTENT（不受遮挡、含标题栏与滚动条、
不会透出背后窗口）。抓完自动裁掉 DWM 的纯黑阴影边，并过一遍交付体检：

  ① 几乎只有一两种颜色 → 全白/全空图（多半是漏了 PW_RENDERFULLCONTENT）
  ② 右端出现纯黑列 → 没裁阴影，或抓图区域被剪（DPI 感知没声明）
  ③ 尺寸异常（过小/过大）

体检不过退 4：**图看起来像成功**是最容易骗人的一种失败。

退出码：
  0  成功
  1  运行失败（IO / 抓图失败）
  2  用法错
  3  找不到目标窗口
  4  产物自查未通过`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return fail(exitUsage, "未知子命令 %q", args[0])
	},
}

func Register(root *cobra.Command) { root.AddCommand(Group) }

func init() {
	Group.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return fail(exitUsage, "%v", err)
	})
	Group.AddCommand(lsCmd, shotCmd)
}

func fail(code int, format string, args ...any) error {
	c := output.CodeUsage
	switch code {
	case exitTarget:
		c = output.CodeTargetMissing
	case exitQuality:
		c = output.CodeQueryFailed
	}
	return &output.Error{Code: c, Exit: code, Message: fmt.Sprintf(format, args...)}
}

// sessionGDCPID 从 fep 的会话登记里取隔离实例的 pid（不要求必须先跑 fep）。
func sessionGDCPID() (uint32, bool) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return 0, false
	}
	dir := config.LocationsAt(cfgPath).DataDir()
	if dir == "" {
		return 0, false
	}
	s, err := fep.NewRegistry(dir).Load()
	if err != nil {
		return 0, false
	}
	return uint32(s.GDCPID), s.GDCPID != 0
}

// cacheDir tt view 的产物落点（截图都在 fep/ 缓存目录下）。
func cacheDir() (string, error) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return "", err
	}
	d := config.LocationsAt(cfgPath).CacheDir("fep")
	if d == "" {
		return "", fail(exitFailed, "缓存目录没解析出来（检查配置路径）")
	}
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", fail(exitFailed, "建缓存目录失败: %v", err)
	}
	return d, nil
}

func printJSONOrText(v any, text func()) error {
	if common.OutputFormat() == output.FormatJSON {
		return common.PrintJSON(v)
	}
	text()
	return nil
}

// noArgs 参数校验：cobra 自带的参数错退 1，而这张线的表里 2 = 用法错。
var noArgs = func(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fail(exitUsage, "%s 不接受位置参数", cmd.CommandPath())
	}
	return nil
}
