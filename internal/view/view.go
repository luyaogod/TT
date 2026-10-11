// Package view 是"窗口与图像"层：枚举窗口、抓真图、裁掉不该有的东西。
//
// 它只做**看**：不碰键鼠、不改窗口样式、不置顶（那些是驱动层的事，本包不提供）。
//
// 抓图后端与理由（都在真机上量过）：
//
//  1. PrintWindow(hwnd, hdc, PW_RENDERFULLCONTENT)：**首选**。不受遮挡、含标题栏与
//     滚动条、不会透出背后窗口的像素。漏掉 PW_RENDERFULLCONTENT 会得到一张全白图 ——
//     这一条当年踩过，是"QT 抓不了"这个错误结论的来源。
//  2. GetDC(hwnd) 客户区 DC：降级。缺标题栏/滚动条。
//  3. 屏幕抓取：所见即所得，但**受遮挡**（实测抓到过电子书窗口和作业窗口完全重合）。
//
// 两条必须记住的规矩：
//   - **DPI 感知必须在任何 GDI 调用之前**声明，否则虚拟桌面尺寸不对，窗口右侧被剪、
//     左侧出现黑边（实测 1920×1200 变成 1536×960 的坐标空间）。
//   - 枚举窗口**只按 pid**（+可见+尺寸）过滤，绝不按类名：作业窗口与开窗的窗口类不同
//     （`Qt5QWindowToolSaveBits` ≠ `Qt5QWindowIcon`），按类名过滤会漏掉开窗。
package view

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Window is one top-level window of a process.
type Window struct {
	HWND    uintptr
	PID     uint32
	Title   string
	Class   string
	Visible bool
	Rect    Rect
	Area    int
}

// Rect is a window rectangle in physical pixels.
type Rect struct{ X, Y, W, H int }

func (r Rect) String() string { return fmt.Sprintf("%dx%d+%d+%d", r.W, r.H, r.X, r.Y) }

// PickWindows filters by pid, keeps visible windows with a sane size, and orders
// them biggest-first (the job window is the big one; popups are small).
//
// pid == 0 means "any process" — 只用于排查，正常路径必须给 pid：
// 用户自己的实例与本会话的实例混在一起时，"抓错了窗口"是最难发现的一种错。
func PickWindows(all []Window, pid uint32) []Window {
	var out []Window
	for _, w := range all {
		if pid != 0 && w.PID != pid {
			continue
		}
		if !w.Visible {
			continue
		}
		if w.Rect.W < 80 || w.Rect.H < 60 {
			continue // 工具提示、阴影窗之类
		}
		w.Area = w.Rect.W * w.Rect.H
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Area > out[j].Area })
	return out
}

// FindWindow picks one window by title substring; empty name means "the biggest".
func FindWindow(wins []Window, name string) (Window, error) {
	if len(wins) == 0 {
		return Window{}, fmt.Errorf("view: 没有可选窗口")
	}
	if name == "" {
		return wins[0], nil
	}
	var hits []Window
	for _, w := range wins {
		if strings.Contains(w.Title, name) {
			hits = append(hits, w)
		}
	}
	switch len(hits) {
	case 0:
		return Window{}, fmt.Errorf("view: 没有标题含 %q 的窗口", name)
	case 1:
		return hits[0], nil
	default:
		return Window{}, fmt.Errorf("view: 标题含 %q 的有 %d 个（%s）：请给更长的名字",
			name, len(hits), titlesOf(hits))
	}
}

func titlesOf(wins []Window) string {
	var t []string
	for _, w := range wins {
		t = append(t, fmt.Sprintf("%q(%s)", w.Title, w.Rect))
	}
	return strings.Join(t, ", ")
}

// TrimShadow cuts the uniform black border DWM leaves around a captured window.
//
// PrintWindow 把窗口阴影区渲染成**纯黑**（实测左/右/下各 9px）。不裁掉，合成到画布上
// 就是一圈黑边。标题栏是亮的（243,243,243），所以**左右列的检测必须跳过标题栏行** ——
// 否则标题栏那几行会把整列判成"非黑"，于是一列都不裁。
func TrimShadow(img *image.RGBA) *image.RGBA {
	if img == nil {
		return nil
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 8 || h < 8 {
		return img
	}
	skip := h / 6 // 顶部这一段当标题栏，不参与左右判定
	isBlackCol := func(x int) bool {
		for y := skip; y < h; y++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			if r>>8 > 8 || g>>8 > 8 || bl>>8 > 8 {
				return false
			}
		}
		return true
	}
	isBlackRow := func(y int) bool {
		for x := 0; x < w; x++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			if r>>8 > 8 || g>>8 > 8 || bl>>8 > 8 {
				return false
			}
		}
		return true
	}
	left, right, top, bottom := 0, 0, 0, 0
	for left < w-1 && isBlackCol(left) {
		left++
	}
	for right < w-1-left && isBlackCol(w-1-right) {
		right++
	}
	for top < h-1 && isBlackRow(top) {
		top++
	}
	for bottom < h-1-top && isBlackRow(h-1-bottom) {
		bottom++
	}
	if left+right+top+bottom == 0 {
		return img
	}
	out := image.NewRGBA(image.Rect(0, 0, w-left-right, h-top-bottom))
	draw.Draw(out, out.Bounds(), img, image.Pt(b.Min.X+left, b.Min.Y+top), draw.Src)
	return out
}

// SavePNG writes an image to path (creating parent directories).
func SavePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// LoadPNG reads a PNG.
func LoadPNG(path string) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba, nil
	}
	out := image.NewRGBA(img.Bounds())
	draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Src)
	return out, nil
}

// QualityReport is the result of the delivery-time check list (计划 §M4).
//
// 抓图是"看起来对"最容易骗人的一步：全白图、被裁掉右缘的图、带黑边的图，
// 都像是成功的截图。所以每次抓完都过一遍体检，并把结论打出来。
type QualityReport struct {
	Width           int      `json:"width"`
	Height          int      `json:"height"`
	RightBlackCols  int      `json:"rightBlackCols"`
	LeftBlackCols   int      `json:"leftBlackCols"`
	BlackBorderRows int      `json:"blackBorderRows"`
	DistinctColors  int      `json:"distinctColors"`
	Problems        []string `json:"problems,omitempty"`
}

// Check runs the delivery checks on a captured image.
func Check(img *image.RGBA) QualityReport {
	rep := QualityReport{}
	if img == nil {
		rep.Problems = append(rep.Problems, "没有图像")
		return rep
	}
	b := img.Bounds()
	rep.Width, rep.Height = b.Dx(), b.Dy()
	blackCol := func(x int) bool {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			if r>>8 > 8 || g>>8 > 8 || bl>>8 > 8 {
				return false
			}
		}
		return true
	}
	for x := b.Max.X - 1; x >= b.Min.X && blackCol(x); x-- {
		rep.RightBlackCols++
	}
	for x := b.Min.X; x < b.Max.X && blackCol(x); x++ {
		rep.LeftBlackCols++
	}
	colors := map[uint32]bool{}
	for y := b.Min.Y; y < b.Max.Y; y += 3 {
		for x := b.Min.X; x < b.Max.X; x += 3 {
			r, g, bl, _ := img.At(x, y).RGBA()
			colors[(r>>8)<<16|(g>>8)<<8|(bl>>8)] = true
		}
	}
	rep.DistinctColors = len(colors)
	if rep.DistinctColors <= 2 {
		rep.Problems = append(rep.Problems, "几乎只有一两种颜色：像是全白/全黑的空图（PW_RENDERFULLCONTENT 漏了？）")
	}
	if rep.RightBlackCols > 0 {
		rep.Problems = append(rep.Problems, fmt.Sprintf("右端有 %d 列纯黑：要么没裁阴影，要么抓图区域被剪（DPI 感知没声明？）", rep.RightBlackCols))
	}
	return rep
}

// Overlay is one annotation mark on a captured image.
type Overlay struct {
	Rect   image.Rectangle
	Color  color.RGBA // 描边色
	Width  int        // 描边宽度
	Number int        // >0：在左上角画序号
}

// Compose draws overlays onto a copy of base. The canvas is never scaled:
// PowerShell/GDI+ 的 DrawImage 会按 DPI 比自动缩放（96/120 → 画布右侧 20% 变黑），
// 所以合成一律用 Go 的 image/draw，逐像素、无缩放。
func Compose(base *image.RGBA, marks []Overlay) *image.RGBA {
	out := image.NewRGBA(base.Bounds())
	draw.Draw(out, out.Bounds(), base, base.Bounds().Min, draw.Src)
	for _, m := range marks {
		thick := m.Width
		if thick <= 0 {
			thick = 3
		}
		col := m.Color
		if col.A == 0 {
			col = color.RGBA{R: 255, G: 64, B: 0, A: 255}
		}
		drawRect(out, m.Rect, col, thick)
		if m.Number > 0 {
			drawBadge(out, m.Rect, m.Number, col)
		}
	}
	return out
}

func drawRect(img *image.RGBA, r image.Rectangle, col color.RGBA, thick int) {
	for t := 0; t < thick; t++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			setPx(img, x, r.Min.Y+t, col)
			setPx(img, x, r.Max.Y-1-t, col)
		}
		for y := r.Min.Y; y < r.Max.Y; y++ {
			setPx(img, r.Min.X+t, y, col)
			setPx(img, r.Max.X-1-t, y, col)
		}
	}
}

// drawBadge draws a small filled square with a digit inside, at p.
func drawBadge(img *image.RGBA, box image.Rectangle, n int, col color.RGBA) {
	const size = 22
	// 徽标画在框**外面**（左侧、垂直居中），别压住框里的内容 —— 实测反馈：
	// 画在框内左上角会盖住第一个字符（那正是手册要让人看清的东西）。
	x0 := box.Min.X - size - 4
	if x0 < 0 {
		x0 = box.Min.X + 2
	}
	y0 := box.Min.Y + (box.Dy()-size)/2
	if y0 < 0 {
		y0 = 0
	}
	for y := y0; y < y0+size; y++ {
		for x := x0; x < x0+size; x++ {
			setPx(img, x, y, col)
		}
	}
	digits := [10][5]string{
		{"111", "101", "101", "101", "111"},
		{"010", "110", "010", "010", "111"},
		{"111", "001", "111", "100", "111"},
		{"111", "001", "111", "001", "111"},
		{"101", "101", "111", "001", "001"},
		{"111", "100", "111", "001", "111"},
		{"111", "100", "111", "101", "111"},
		{"111", "001", "010", "010", "010"},
		{"111", "101", "111", "101", "111"},
		{"111", "101", "111", "001", "111"},
	}
	s := fmt.Sprintf("%d", n)
	scale := 3
	totalW := len(s)*3*scale + (len(s)-1)*scale
	x := x0 + (size-totalW)/2
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for _, ch := range s {
		pat := digits[int(ch-'0')]
		for row := 0; row < 5; row++ {
			for cIdx := 0; cIdx < 3; cIdx++ {
				if pat[row][cIdx] != '1' {
					continue
				}
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						setPx(img, x+cIdx*scale+dx, y0+6+row*scale+dy, white)
					}
				}
			}
		}
		x += 3*scale + scale
	}
}

func setPx(img *image.RGBA, x, y int, col color.RGBA) {
	if !(image.Point{X: x, Y: y}).In(img.Bounds()) {
		return
	}
	img.SetRGBA(x, y, col)
}
