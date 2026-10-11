package fep

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

var (
	waitWindow  string
	waitValue   string
	waitText    string
	waitTimeout time.Duration
)

var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "等作业发生变化（等窗口出现 / 等字段变值 / 等文本出现）",
	Long: `等到作业真的变成了某个样子再继续。

轮询的是**协议录制件**而不是屏幕：协议是唯一权威，屏幕是它的渲染结果。
三个条件可以叠（都必须成立）；默认 10s 超时。

用法举例：
  tt fep wait --window w_axmt503          # 等某个开窗出现
  tt fep wait --value 客户编号=CNJ-A50    # 等某字段变成某个值（名字按协议解析）
  tt fep wait --text 请选择               # 等界面上出现某段文本`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cond := fep.WaitCond{Window: waitWindow, Value: waitValue, Text: waitText}
		if cond.Window == "" && cond.Value == "" && cond.Text == "" {
			return fail(exitUsage, "至少给一个条件（--window / --value / --text）")
		}
		client, err := newClient()
		if err != nil {
			return err
		}
		start := time.Now()
		m, werr := client.WaitFor(cond, waitTimeout)
		if werr != nil {
			return fail(exitFailed, "%v", werr)
		}
		type out struct {
			Waited    string `json:"waited"` // 实际等了多久
			Condition string `json:"condition"`
			Windows   int    `json:"windows"`
			Objects   int    `json:"objects"`
		}
		res := out{Waited: time.Since(start).Round(time.Millisecond).String(), Condition: cond.String()}
		if m != nil {
			res.Windows, res.Objects = len(m.Windows()), m.Len()
		}
		return printJSONOrText(res, func() {
			fmt.Printf("条件成立（等了 %s）：%s\n", res.Waited, res.Condition)
			if m != nil {
				fmt.Printf("  当前界面：窗口 %d 个 / 对象 %d 个\n", res.Windows, res.Objects)
			}
		})
	},
}

// newClient 按会话登记与缓存落点造一个 Client（跨进程与常驻会话说话）。
func newClient() (*fep.Client, error) {
	reg, _, err := registry()
	if err != nil {
		return nil, err
	}
	s, err := reg.Load()
	if errors.Is(err, fep.ErrNoSession) {
		return nil, fail(exitNoSession, "没有登记中的会话（先跑 tt fep session open）")
	}
	if err != nil {
		return nil, fail(exitFailed, "%v", err)
	}
	recDir := s.RecordDir
	if recDir == "" {
		if recDir, err = cacheDir(); err != nil {
			return nil, err
		}
	}
	return &fep.Client{
		InjectAddr: net.JoinHostPort("127.0.0.1", strconv.Itoa(s.InjectPort)),
		RecordDir:  recDir,
		Tag:        s.Tag,
	}, nil
}

func init() {
	waitCmd.Flags().StringVar(&waitWindow, "window", "", "等这个窗口名出现")
	waitCmd.Flags().StringVar(&waitValue, "value", "", "等某字段变成某个值：字段名=值")
	waitCmd.Flags().StringVar(&waitText, "text", "", "等界面上出现这段文本")
	waitCmd.Flags().DurationVar(&waitTimeout, "timeout", 10*time.Second, "等多久（默认 10s）")
	Group.AddCommand(waitCmd)
}
