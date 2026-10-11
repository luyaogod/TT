package view

import (
	"image"
	"image/color"
	"testing"
)

// 映射从参照物派生：参照物 1000x800 协议单位、图 1200x1000 像素
// → 比例 1.2，客户区高 960，标题栏 40px。
func TestDeriveAndMap(t *testing.T) {
	fm, problems := Derive(Box{W: 1000, H: 800}, image.Rect(0, 0, 1200, 1000))
	if len(problems) != 0 {
		t.Fatalf("不该报问题：%v", problems)
	}
	if fm.ScaleX != 1.2 || fm.OriginY != 40 {
		t.Fatalf("派生结果 = %+v", fm)
	}
	got := fm.Map(Box{X: 10, Y: 20, W: 100, H: 30})
	want := image.Rect(12, 40+24, 12+120, 40+24+36)
	if got != want {
		t.Fatalf("映射 = %v，期望 %v", got, want)
	}
}

// 参照物离谱时必须报出来，而不是接着画框。
func TestDeriveReportsImplausibleReference(t *testing.T) {
	if _, problems := Derive(Box{W: 10, H: 10}, image.Rect(0, 0, 1200, 1000)); len(problems) == 0 {
		t.Fatal("比例严重不符必须报问题")
	}
	if _, problems := Derive(Box{W: 0, H: 0}, image.Rect(0, 0, 100, 100)); len(problems) == 0 {
		t.Fatal("尺寸不合法必须报问题")
	}
}

// 像素校验：那块地方必须真的画了东西（纯色/纯黑 = 没画出来）。
func TestRegionVerdictCatchesBlankAreas(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 60, 40))
	fill(img, img.Bounds(), color.RGBA{240, 240, 240, 255})
	if got := RegionVerdict(img, image.Rect(5, 5, 25, 15), "订购量"); got == "" {
		t.Fatal("纯色区域必须报问题（那一帧里没画出来）")
	}
	// 画一个带边框的框（真实栏位就是这个样子）
	for x := 6; x < 24; x++ {
		img.SetRGBA(x, 6, color.RGBA{0, 0, 0, 255})
		img.SetRGBA(x, 14, color.RGBA{0, 0, 0, 255})
	}
	for y := 6; y < 15; y++ {
		img.SetRGBA(6, y, color.RGBA{0, 0, 0, 255})
		img.SetRGBA(23, y, color.RGBA{0, 0, 0, 255})
	}
	if got := RegionVerdict(img, image.Rect(5, 5, 25, 15), "订购量"); got != "" {
		t.Fatalf("画了东西就不该报：%s", got)
	}
	// 超出图外
	if got := RegionVerdict(img, image.Rect(50, 30, 300, 300), "订购量"); got == "" {
		t.Fatal("区域大半在图外必须报问题")
	}
}

// 门：看不出问题的栏位集合为空，出问题的给出证据行。
func TestGateFieldsReportsNamedProblemFields(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	fill(img, img.Bounds(), color.RGBA{245, 245, 245, 255})
	for x := 20; x < 120; x++ {
		img.SetRGBA(x, 25, color.RGBA{0, 0, 0, 255})
		img.SetRGBA(x, 34, color.RGBA{0, 0, 0, 255})
	}
	for y := 25; y < 35; y++ {
		img.SetRGBA(20, y, color.RGBA{0, 0, 0, 255})
		img.SetRGBA(119, y, color.RGBA{0, 0, 0, 255})
	}
	fm := FrameMap{ScaleX: 1, ScaleY: 1}
	f := Frame{Image: img, Map: fm, Fields: []Box{
		{X: 20, Y: 25, W: 100, H: 10, Label: "客户编号"},
		{X: 20, Y: 60, W: 100, H: 10, Label: "订购量"},
	}}
	bad := f.Gate()
	if len(bad) != 1 || bad[0].Label != "订购量" {
		t.Fatalf("应只有订购量不过：%+v", bad)
	}
	if s := Describe(bad); !contains(s, "订购量") {
		t.Fatalf("描述里要有栏位名：%s", s)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && (indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
