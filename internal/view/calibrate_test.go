package view

import (
	"image"
	"image/color"
	"testing"
)

func TestDiffBoxFindsTheTypedText(t *testing.T) {
	a := image.NewRGBA(image.Rect(0, 0, 200, 100))
	b := image.NewRGBA(image.Rect(0, 0, 200, 100))
	fill(a, a.Bounds(), color.RGBA{240, 240, 240, 255})
	fill(b, b.Bounds(), color.RGBA{240, 240, 240, 255})
	// 在 (120..150, 40..52) 画上"文字"
	for y := 40; y < 52; y++ {
		for x := 120; x < 150; x++ {
			b.SetRGBA(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	box, n := DiffBox(a, b, 30)
	if n != 30*12 || box != image.Rect(120, 40, 150, 52) {
		t.Fatalf("diff = %v（%d 像素）", box, n)
	}
	if _, n0 := DiffBox(a, a, 30); n0 != 0 {
		t.Fatal("同一张图不该有差异")
	}
}

// 两个实测锚点解出映射：真机那一对（posX=15 → x=526；posY=2 → y=215，
// posY=1 → y=183 → 格高 32）就是这条测试的来源。
func TestSolveMappingFromTwoSamples(t *testing.T) {
	samples := []Sample{
		{Box: Box{X: 15, Y: 2, Label: "客户编号"}, Pixels: image.Rect(526, 215, 574, 228)},
		{Box: Box{X: 12, Y: 1, Label: "订单单号"}, Pixels: image.Rect(505, 183, 545, 196)},
	}
	fm, problems := SolveMapping(samples)
	if len(problems) != 0 {
		t.Fatalf("不该报问题：%v", problems)
	}
	// 横向：两点解出 7px/格、原点 421；纵向：32px/格、原点 151
	if fm.ScaleX < 6.9 || fm.ScaleX > 7.1 || fm.ScaleY != 32 {
		t.Fatalf("映射 = %+v", fm)
	}
	// 用它预测第一个样本的位置，应落在实测框里
	got := fm.Map(Box{X: 15, Y: 2, W: 10, H: 1})
	if got.Min.X != 526 || got.Min.Y != 215 {
		t.Fatalf("预测 %v，实测锚点 526,215", got)
	}
}

func TestSolveMappingRefusesBadSamples(t *testing.T) {
	if _, p := SolveMapping([]Sample{{Box: Box{X: 1}}}); len(p) == 0 {
		t.Fatal("样本不足必须报问题")
	}
	same := []Sample{
		{Box: Box{X: 5, Y: 5, Label: "a"}, Pixels: image.Rect(100, 100, 120, 110)},
		{Box: Box{X: 5, Y: 5, Label: "b"}, Pixels: image.Rect(100, 100, 130, 110)},
	}
	if _, p := SolveMapping(same); len(p) == 0 {
		t.Fatal("同一格上的样本解不出比例，必须报问题")
	}
}
