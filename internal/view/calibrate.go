package view

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
)

// 校准：协议格 → 像素的换算**靠实测**，不靠推断。
//
// 为什么必须实测：真实 AUI 里几何属性既不是一套单位、也不是一个固定比例 ——
// 同一屏里 `ButtonEdit 272 {posX "15"}{width "8"}`（格）与 `Window 178 {width "1912"}`
// （像素）并存；而且同一个 posX 在不同行上的像素位置还要叠上客户区的偏移。
// 写死一个比例，换主题/换字体/换 DPI 就全错，**而且错得很像对的**（框还画在图上，
// 只是不在栏位上）。
//
// 所以：往两个已知栏位各写一个值、拍两张图、取**像素差**当锚点，解出 (格宽, 格高, 原点)。

// DiffBox returns the bounding box and count of changed pixels between two captures.
//
// 这是整套校准的地基：注入一次输入之后，**变化的那些像素**就是那个栏位在屏幕上的真实位置 ——
// 不是推出来的，是量出来的。
func DiffBox(a, b *image.RGBA, threshold int) (image.Rectangle, int) {
	if a == nil || b == nil || a.Bounds() != b.Bounds() {
		return image.Rectangle{}, 0
	}
	if threshold <= 0 {
		threshold = 30
	}
	bounds := a.Bounds()
	minX, minY := 1<<30, 1<<30
	maxX, maxY := -1, -1
	changed := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			ar, ag, ab, _ := a.At(x, y).RGBA()
			br, bg, bb, _ := b.At(x, y).RGBA()
			d := absInt(int(ar>>8)-int(br>>8)) + absInt(int(ag>>8)-int(bg>>8)) + absInt(int(ab>>8)-int(bb>>8))
			if d < threshold {
				continue
			}
			changed++
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if changed == 0 {
		return image.Rectangle{}, 0
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), changed
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// obs is one (unit, pixel) observation.
type obs struct{ unit, px float64 }

// Sample is one measured anchor: writing into the field whose protocol box is Box
// changed the pixels in Pixels.
type Sample struct {
	Box    Box             // 协议坐标（含 posX/posY）
	Pixels image.Rectangle // 实测像素位置
}

// SolveMapping solves scale and origin from measured samples.
//
// 两个样本就够（两个未知数、两个方程）；样本多了取最小二乘，不一致就报出来 ——
// 不一致通常意味着"用量到的不是同一个字体/行"或"中间被别的重绘污染了"，
// 那种时候宁可报错，也不要给一个勉强拟合出来的映射。
func SolveMapping(samples []Sample) (FrameMap, []string) {
	var problems []string
	fm := FrameMap{ScaleX: 1, ScaleY: 1}
	var xs, ys []obs
	for _, s := range samples {
		if s.Pixels.Empty() {
			problems = append(problems, fmt.Sprintf("%s：没量到像素变化", s.Box.Label))
			continue
		}
		xs = append(xs, obs{unit: float64(s.Box.X), px: float64(s.Pixels.Min.X)})
		ys = append(ys, obs{unit: float64(s.Box.Y), px: float64(s.Pixels.Min.Y)})
	}
	if len(xs) < 2 || len(ys) < 2 {
		return fm, append(problems, "至少需要两个不同位置的样本才能解出比例")
	}
	sx, ox, okX := fitLine(xs)
	sy, oy, okY := fitLine(ys)
	if !okX || !okY {
		return fm, append(problems, "所有样本在同一格上：解不出比例（换一行/一列再取一个）")
	}
	if sx <= 0 || sy <= 0 {
		return fm, append(problems, fmt.Sprintf("解出来的比例不是正数（格宽 %.2f，格高 %.2f）", sx, sy))
	}
	fm.ScaleX, fm.ScaleY = sx, sy
	fm.OriginX, fm.OriginY = int(math.Round(ox)), int(math.Round(oy))
	// 一致性：残差太大说明样本不是同一种重绘，报出来
	if rx := maxResidual(xs, sx, ox); rx > 6 {
		problems = append(problems, fmt.Sprintf("横向样本残差 %.1fpx 偏大：两次取值可能不是同一列", rx))
	}
	if ry := maxResidual(ys, sy, oy); ry > 6 {
		problems = append(problems, fmt.Sprintf("纵向样本残差 %.1fpx 偏大：两次取值可能不是同一行", ry))
	}
	return fm, problems
}

// fitLine returns slope and intercept for px = slope*unit + intercept.
func fitLine(obs []obs) (float64, float64, bool) {
	var meanU, meanP float64
	for _, o := range obs {
		meanU += o.unit
		meanP += o.px
	}
	meanU /= float64(len(obs))
	meanP /= float64(len(obs))
	var num, den float64
	for _, o := range obs {
		num += (o.unit - meanU) * (o.px - meanP)
		den += (o.unit - meanU) * (o.unit - meanU)
	}
	if den == 0 {
		return 0, 0, false
	}
	slope := num / den
	return slope, meanP - slope*meanU, true
}

func maxResidual(obs []obs, slope, intercept float64) float64 {
	worst := 0.0
	for _, o := range obs {
		if r := math.Abs(o.px - (slope*o.unit + intercept)); r > worst {
			worst = r
		}
	}
	return worst
}

// Calibration is a saved mapping plus the evidence it was derived from.
//
// 存"证据"而不只存结果：换主题/换字体之后映射会失效，而**样本与残差**是判断
// "这份校准还能不能用"的唯一依据。
type Calibration struct {
	ScaleX  float64  `json:"scaleX"`
	ScaleY  float64  `json:"scaleY"`
	OriginX int      `json:"originX"`
	OriginY int      `json:"originY"`
	At      string   `json:"at"`            // 什么时候量的
	Env     string   `json:"env,omitempty"` // 哪个环境/作业
	Samples []Sample `json:"samples,omitempty"`
	Notes   []string `json:"notes,omitempty"`
}

// Map converts a protocol box with this calibration.
func (c Calibration) Map(b Box) image.Rectangle {
	return FrameMap{ScaleX: c.ScaleX, ScaleY: c.ScaleY, OriginX: c.OriginX, OriginY: c.OriginY}.Map(b)
}

// SaveCalibration writes the calibration as JSON.
func SaveCalibration(path string, c Calibration) error {
	b, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadCalibration reads a calibration file.
func LoadCalibration(path string) (Calibration, error) {
	var c Calibration
	b, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return c, err
	}
	if c.ScaleX <= 0 || c.ScaleY <= 0 {
		return c, fmt.Errorf("view: 校准文件里的比例不合法（%g, %g）", c.ScaleX, c.ScaleY)
	}
	return c, nil
}
