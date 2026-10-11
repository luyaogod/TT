package view

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
	"tt/internal/view"
)

// tt view measure：给"协议里没有坐标的东西"量一个位置出来。
//
// 明细网格的格子就是这种：TableColumn 与运行时的 Value 都没有 posX/posY（真机实测：
// 带几何的 Value 对象数为 0）。没有坐标就不能推 —— 只能量：往那个格子写一个值、
// 拍前后两张图、取**像素差**。量出来的矩形连同**来源与时间**一起存下来，
// 换主题/分辨率之后失效时能看出来（而不是悄悄画错框）。

var (
	measureName    string
	measureTarget  string
	measureValue   string
	measureOut     string
	measureMinPix  int
	measureNote    string
	measureTimeout time.Duration
)

var measureCmd = &cobra.Command{
	Use:   "measure",
	Short: "实测一个区域的位置（注入一个值 + 取像素差），存成命名区域",
	Long: `把一个"协议里没有坐标"的东西量出来，供 marks --region 使用。

做法与 calibrate 同一套地基：拍下前后两张真图，差异的包围盒就是它在屏幕上的位置。
判据：差异像素太少（默认 < 20）判为"没量到"，直接失败 —— 宁可没有区域，也不要一个空壳。

存下来的每条区域都带来源与时间：换主题/分辨率/作业之后可能失效，
所以 marks 用它出框时会把它标成"实测区域"而不是"协议坐标"。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if measureTarget == "" && measureName == "" {
			return fail(exitUsage, "至少给 --target（要往哪写值）或 --name（区域名）")
		}
		c, err := newFEPClient()
		if err != nil {
			return err
		}
		before, berr := captureForCalibration()
		if berr != nil {
			return berr
		}
		value := measureValue
		if value == "" {
			value = "13579"
		}
		if _, derr := fep.Do(c, fep.DoConfig{Kind: "input", Target: measureTarget,
			Value: value, Wait: measureTimeout}); derr != nil {
			return fail(exitFailed, "注入失败（没量到东西）：%v", derr)
		}
		after, aerr := captureForCalibration()
		if aerr != nil {
			return aerr
		}
		rect, changed := view.DiffBox(before, after, 30)
		if changed == 0 {
			return fail(exitQuality, "写了值但屏幕没有变化：这个目标写不进去，或者它不在当前这一帧里")
		}
		// 太薄/太扁的差异框不是栏位：多半是光标闪动，或者“写进去的值与原来的相同”
		// （实测踩过：往已经是 2026/10/11 的日期栏位再写同一个值，量到 1x23 的细条）。
		if rect.Dx() < 8 || rect.Dy() < 8 {
			return fail(exitQuality, "差异框 %dx%d 太薄，不像一个栏位：换一个与当前值不同的值再量", rect.Dx(), rect.Dy())
		}
		if measureMinPix > 0 && changed < measureMinPix {
			return fail(exitQuality, "只变了 %d 个像素（少于 %d）：这个差异不像一个栏位/格子", changed, measureMinPix)
		}
		name := measureName
		if name == "" {
			name = strings.TrimPrefix(measureTarget, "#")
		}
		path := measureOut
		if path == "" {
			dir, derr := cacheDir()
			if derr != nil {
				return derr
			}
			path = filepath.Join(dir, "regions.json")
		}
		rs, lerr := view.LoadRegions(path)
		if lerr != nil {
			return fail(exitFailed, "读区域文件失败: %v", lerr)
		}
		// 体检：一行文字的变化框不该有两行高；高了说明这一枪把邻居也重画了，
		// 量出来的框会比目标栏位大 —— 照实存，但**记下来**，别让人误以为它是栏位边界。
		note := measureNote
		if rect.Dy() > 40 && !strings.Contains(note, "偏大") {
			note = strings.TrimSpace(note + "；差异框高 " + fmt.Sprint(rect.Dy()) + "px 超过一行：可能含邻居重绘，框偏大")
		}
		reg := view.Region{Name: name, Rect: rect,
			RectTxt: fmt.Sprintf("%d,%d,%d,%d", rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy()),
			At:      time.Now().Format(time.RFC3339),
			Source:  fmt.Sprintf("注入 %s=%q 后取像素差", measureTarget, value),
			Note:    note}
		rs.Put(reg)
		if serr := view.SaveRegions(path, rs); serr != nil {
			return fail(exitFailed, "写区域失败: %v", serr)
		}
		return printJSONOrText(map[string]any{"path": path, "name": name, "rect": reg.RectTxt,
			"changedPixels": changed, "regions": len(rs.Regions)}, func() {
			fmt.Printf("已量到区域 %q：%s（变化 %d 像素）\n", name, reg.RectTxt, changed)
			fmt.Printf("  存档：%s（共 %d 个区域）\n", path, len(rs.Regions))
		})
	},
}

func init() {
	measureCmd.Flags().StringVar(&measureName, "name", "", "区域名（默认用目标名）")
	measureCmd.Flags().StringVar(&measureTarget, "target", "", "往哪写值：栏位名或 #id（如 #864）")
	measureCmd.Flags().StringVar(&measureValue, "value", "", "写入的值（默认 13579）")
	measureCmd.Flags().StringVar(&measureOut, "out", "", "区域文件（默认 <缓存>/regions.json）")
	measureCmd.Flags().IntVar(&measureMinPix, "min-pixels", 20, "差异像素下限，少于它就判为没量到")
	measureCmd.Flags().StringVar(&measureNote, "note", "", "备注（写进区域档案）")
	measureCmd.Flags().DurationVar(&measureTimeout, "wait", 5*time.Second, "注入后等运行期多久")
	Group.AddCommand(measureCmd)
}
