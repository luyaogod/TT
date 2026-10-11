package view

import (
	"fmt"
	"image"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/view"
)

var (
	shotPID    uint32
	shotWindow string
	shotOut    string
	shotNoTrim bool
	shotMarks  []string
)

var shotCmd = &cobra.Command{
	Use:   "shot",
	Short: "抓一个窗口的真图（含体检，不过退 4）",
	Long: `抓目标窗口的图并落盘，同时给出交付体检结论。

--window 给标题片段（不给就是最大的那个窗口）。--mark "x,y,w,h[,序号]" 可重复，
在图上画出标注框（坐标是**窗口内的像素**，落点由调用方从协议里拿到，不靠猜）。

产物默认落在 fep 缓存目录：<缓存>/shot-<时间>.png`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		pid := shotPID
		if pid == 0 {
			p, ok := sessionGDCPID()
			if !ok {
				return fail(exitUsage, "没有 fep 会话可以推断 pid：给 --pid <pid>")
			}
			pid = p
		}
		wins, err := view.ListTimeout(pid, 5*time.Second)
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		picked := view.PickWindows(wins, pid)
		if len(picked) == 0 {
			return fail(exitTarget, "pid %d 没有可见的候选窗口（作业还没开？）", pid)
		}
		target, ferr := view.FindWindow(picked, shotWindow)
		if ferr != nil {
			return fail(exitTarget, "%v", ferr)
		}
		img, cerr := view.Capture(target.HWND)
		if cerr != nil {
			return fail(exitFailed, "%v", cerr)
		}
		if !shotNoTrim {
			img = view.TrimShadow(img)
		}
		marks, merr := parseMarks(shotMarks)
		if merr != nil {
			return merr
		}
		if len(marks) > 0 {
			img = view.Compose(img, marks)
		}
		out := shotOut
		if out == "" {
			dir, derr := cacheDir()
			if derr != nil {
				return derr
			}
			out = filepath.Join(dir, "shot-"+time.Now().Format("20060102-150405")+".png")
		}
		if err := view.SavePNG(out, img); err != nil {
			return fail(exitFailed, "写图失败: %v", err)
		}
		rep := view.Check(img)
		res := map[string]any{"path": out, "window": target.Title, "hwnd": fmt.Sprintf("%x", target.HWND),
			"rect": target.Rect.String(), "size": fmt.Sprintf("%dx%d", rep.Width, rep.Height), "quality": rep}
		if perr := printJSONOrText(res, func() {
			fmt.Printf("已抓：%s\n", out)
			fmt.Printf("  窗口 %q %s（hwnd %x）\n", target.Title, target.Rect, target.HWND)
			fmt.Printf("  图片 %dx%d，颜色数 %d\n", rep.Width, rep.Height, rep.DistinctColors)
			for _, p := range rep.Problems {
				fmt.Printf("  ⚠ %s\n", p)
			}
		}); perr != nil {
			return perr
		}
		if len(rep.Problems) > 0 {
			return fail(exitQuality, "产物自查未通过（见上面 ⚠）：%s", strings.Join(rep.Problems, "；"))
		}
		return nil
	},
}

// parseMarks 解析 "x,y,w,h[,序号]"。
func parseMarks(specs []string) ([]view.Overlay, error) {
	var out []view.Overlay
	for _, s := range specs {
		parts := strings.Split(s, ",")
		if len(parts) < 4 {
			return nil, fail(exitUsage, "--mark 要写成 x,y,w,h[,序号]：%q", s)
		}
		var v [5]int
		for i := 0; i < 4; i++ {
			n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
			if err != nil {
				return nil, fail(exitUsage, "--mark 里的数字读不出来：%q", s)
			}
			v[i] = n
		}
		if len(parts) >= 5 {
			n, err := strconv.Atoi(strings.TrimSpace(parts[4]))
			if err != nil {
				return nil, fail(exitUsage, "--mark 的序号读不出来：%q", s)
			}
			v[4] = n
		}
		out = append(out, view.Overlay{Rect: image.Rect(v[0], v[1], v[0]+v[2], v[1]+v[3]), Number: v[4]})
	}
	return out, nil
}

func init() {
	shotCmd.Flags().Uint32Var(&shotPID, "pid", 0, "目标进程（默认取 fep 会话的隔离实例）")
	shotCmd.Flags().StringVar(&shotWindow, "window", "", "窗口标题片段（默认最大的那个）")
	shotCmd.Flags().StringVar(&shotOut, "out", "", "输出路径（默认 <缓存>/shot-<时间>.png）")
	shotCmd.Flags().BoolVar(&shotNoTrim, "no-trim", false, "不裁 DWM 阴影黑边")
	shotCmd.Flags().StringArrayVar(&shotMarks, "mark", nil, "标注框 x,y,w,h[,序号]（可重复）")
}
