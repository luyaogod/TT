package view

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"
)

// 协议坐标 → 图像坐标的映射，以及“说明里点名的栏位在这一帧里真的看得见吗”。
//
// 这一层只认几何与像素，不认识 AUI：把 AUI 属性翻译成 Box 是**调用方**（CLI 层）的事，
// 这样 internal/form42f 与 internal/view 谁都不 import 谁（跨线只走 JSON 产物与会话登记）。
//
// 为什么需要"派生"而不是写死比例：AUI 的 posX/posY/gridWidth/gridHeight 是协议自己的
// 坐标系（栅格/字符单位，随主题与字体而变），抓到的图是物理像素。写死一个比例，
// 换一台机器或换一个主题就全错，而且**错得很像对的**（框还画在图上，只是不在栏位上）。
// 所以映射从**参照物**推出来：拿 AUI 里那个顶层窗口/表单的尺寸，与抓到的客户区尺寸对齐。

// Box is a rectangle in the protocol's own coordinate system.
type Box struct {
	X, Y, W, H int
	Label      string // 让人看得懂的名字（栏位名/窗口号），只用于报错与标注
}

// FrameMap maps protocol coordinates onto image pixels.
type FrameMap struct {
	OriginX, OriginY int
	ScaleX, ScaleY   float64
}

// Derive builds the mapping from a reference box (the AUI's top-level window or
// form) and the captured image size.
//
// 纵向的做法值得说明：抓的是**整窗**（含标题栏），而 AUI 的参照物是**客户区**，
// 两者差一个标题栏高度，而这个高度协议里没有。这里不去猜它，而是用宽度定比例
// （宽度不受标题栏影响），再由"参照物应该占满客户区"反解出整体的纵向下移量。
// 解出来的东西如果离谱（负比例、偏移超过高度的四分之一），就**报出来**而不是接着画。
func Derive(ref Box, captured image.Rectangle) (FrameMap, []string) {
	var problems []string
	fm := FrameMap{ScaleX: 1, ScaleY: 1}
	if ref.W <= 0 || ref.H <= 0 {
		return fm, []string{fmt.Sprintf("参照物尺寸不合法（%dx%d）", ref.W, ref.H)}
	}
	if captured.Dx() <= 0 || captured.Dy() <= 0 {
		return fm, []string{"抓到的图尺寸不合法"}
	}
	fm.ScaleX = float64(captured.Dx()) / float64(ref.W)
	fm.ScaleY = fm.ScaleX // 同一套栅格，纵横比例必然一致
	// 客户区高度（像素）与参照物高度（协议单位）之比应等于 ScaleY；
	// 差额就是标题栏+边框的像素数。
	clientH := float64(ref.H) * fm.ScaleY
	offsetY := float64(captured.Dy()) - clientH
	if offsetY < 0 {
		offsetY = 0
	}
	fm.OriginY = int(offsetY + 0.5)
	fm.OriginX = 0
	if fm.ScaleX <= 0 {
		problems = append(problems, "推出来的横向比例不是正数：参照物与图像的对应关系不成立")
	}
	if float64(fm.OriginY) > float64(captured.Dy())/4 {
		problems = append(problems, fmt.Sprintf(
			"推出来的纵向下移量 %dpx 过大（超过图高的 1/4）：参照物多半不是客户区（选错对象了？）", fm.OriginY))
	}
	if fm.ScaleX < 0.2 || fm.ScaleX > 8 {
		problems = append(problems, fmt.Sprintf(
			"推出来的比例 %.3f 不像同一套坐标系：参照物与图像可能不同源", fm.ScaleX))
	}
	return fm, problems
}

// Map turns a protocol box into an image rectangle.
func (fm FrameMap) Map(b Box) image.Rectangle {
	x0 := fm.OriginX + int(float64(b.X)*fm.ScaleX)
	y0 := fm.OriginY + int(float64(b.Y)*fm.ScaleY)
	x1 := x0 + int(float64(b.W)*fm.ScaleX)
	y1 := y0 + int(float64(b.H)*fm.ScaleY)
	return image.Rect(x0, y0, x1, y1)
}

// Inside reports whether a rectangle lies (mostly) inside the image.
func Inside(r image.Rectangle, img image.Rectangle) bool {
	if r.Empty() {
		return false
	}
	inter := r.Intersect(img)
	return inter.Dx()*inter.Dy() >= (r.Dx()*r.Dy())*3/4
}

// RegionVerdict checks the pixels under one mapped rectangle: a field that the
// manual names must actually be **drawn there**.
//
// 这一条直接防住"标题与画面不一致"那类错：说明里写“在这里输入订购量”，
// 而那一帧根本没走到那个栏位 —— 框还会画在图上，看着像真的。
func RegionVerdict(img *image.RGBA, r image.Rectangle, label string) string {
	if img == nil {
		return "没有图像"
	}
	if r.Empty() {
		return fmt.Sprintf("%s：映射出来的矩形是空的", label)
	}
	if !Inside(r, img.Bounds()) {
		return fmt.Sprintf("%s：映射出的区域 %v 有一半以上在图外（坐标映射或窗口选择不对）", label, r)
	}
	// 逐像素采样：隔行采样会漏掉细内容（一行文字、一条边框），
	// 而栏位矩形本身很小，全采也不贵。
	colors := map[uint32]int{}
	total := 0
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			colors[(cr>>8)<<16|(cg>>8)<<8|(cb>>8)]++
			total++
		}
	}
	if total == 0 {
		return fmt.Sprintf("%s：区域里没有采样点", label)
	}
	// 全黑或只有一种颜色 = 那块地方什么都没画
	if len(colors) <= 1 {
		return fmt.Sprintf("%s：这块区域是纯色（%d 种颜色）—— 该栏位在那一帧里没画出来", label, len(colors))
	}
	black := 0
	for c, n := range colors {
		if r, g, b := (c>>16)&0xff, (c>>8)&0xff, c&0xff; r < 8 && g < 8 && b < 8 {
			black += n
		}
	}
	if black*10 > total*9 {
		return fmt.Sprintf("%s：区域里九成以上是纯黑", label)
	}
	return ""
}

// Verdict is the result of the visibility gate for one field.
type Verdict struct {
	Label   string          `json:"label"`
	Rect    image.Rectangle `json:"-"`
	RectTxt string          `json:"rect"`
	Problem string          `json:"problem,omitempty"`
	Colors  int             `json:"colors,omitempty"`
}

// GateFields runs the visibility gate over a list of mapped fields.
//
// 判据（计划 §2.12）：每个被点名的栏位，必须在**那一帧**里看得见 ——
// 看不见就换帧、用放大图兜底，或者改说明；不许把看不见的东西写进说明。
func GateFields(img *image.RGBA, fields []Box, fm FrameMap) []Verdict {
	out := make([]Verdict, 0, len(fields))
	for _, b := range fields {
		r := fm.Map(b)
		v := Verdict{Label: b.Label, Rect: r, RectTxt: fmt.Sprintf("%dx%d+%d+%d", r.Dx(), r.Dy(), r.Min.X, r.Min.Y)}
		// 采样一下颜色数，方便报错时给出证据
		if img != nil && !r.Empty() {
			seen := map[uint32]bool{}
			for y := r.Min.Y; y < r.Max.Y; y++ {
				for x := r.Min.X; x < r.Max.X; x++ {
					if !(image.Point{X: x, Y: y}).In(img.Bounds()) {
						continue
					}
					cr, cg, cb, _ := img.At(x, y).RGBA()
					seen[(cr>>8)<<16|(cg>>8)<<8|(cb>>8)] = true
				}
			}
			v.Colors = len(seen)
		}
		v.Problem = RegionVerdict(img, r, b.Label)
		out = append(out, v)
	}
	return out
}

// Frame is a captured picture plus where the protocol says things are.
type Frame struct {
	Image  *image.RGBA
	Map    FrameMap
	Fields []Box
	Notes  []string
}

// Gate runs the visibility gate and returns the failing verdicts (nil = all good).
func (f Frame) Gate() []Verdict {
	var bad []Verdict
	for _, v := range GateFields(f.Image, f.Fields, f.Map) {
		if v.Problem != "" {
			bad = append(bad, v)
		}
	}
	return bad
}

// Describe renders the gate result as lines for humans.
func Describe(verdicts []Verdict) string {
	if len(verdicts) == 0 {
		return "所有被点名的栏位在这一帧里都看得见"
	}
	var lines []string
	sort.Slice(verdicts, func(i, j int) bool { return verdicts[i].Label < verdicts[j].Label })
	for _, v := range verdicts {
		lines = append(lines, fmt.Sprintf("%s（%s，%d 色）：%s", v.Label, v.RectTxt, v.Colors, v.Problem))
	}
	return strings.Join(lines, "；")
}

// HighlightColor is the default mark colour (orange: visible on T100's grey UI).
var HighlightColor = color.RGBA{R: 255, G: 96, B: 0, A: 255}