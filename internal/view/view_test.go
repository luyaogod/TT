package view

import (
	"image"
	"image/color"
	"testing"
)

func fill(img *image.RGBA, r image.Rectangle, c color.RGBA) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

// 阴影裁切：DWM 的阴影是纯黑边框，标题栏（亮）那几行**不能**参与左右判定 ——
// 否则整列判成非黑，一列都裁不掉。
func TestTrimShadowKeepsTitleBarAndCutsBlackBorder(t *testing.T) {
	const shadow = 9
	w, h := 200, 120
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	fill(img, img.Bounds(), color.RGBA{0, 0, 0, 255})
	// 亮标题栏 + 白色内容区，四周留 shadow 像素黑边
	fill(img, image.Rect(shadow, shadow, w-shadow, shadow+24), color.RGBA{243, 243, 243, 255})
	fill(img, image.Rect(shadow, shadow+24, w-shadow, h-shadow), color.RGBA{255, 255, 255, 255})
	got := TrimShadow(img)
	if got.Bounds().Dx() != w-2*shadow || got.Bounds().Dy() != h-2*shadow {
		t.Fatalf("裁切结果 = %v，期望 %dx%d", got.Bounds(), w-2*shadow, h-2*shadow)
	}
	if r, g, b, _ := got.At(0, 0).RGBA(); r>>8 < 200 || g>>8 < 200 || b>>8 < 200 {
		t.Fatalf("裁完后左上角应是标题栏的亮色，实际 %d,%d,%d", r>>8, g>>8, b>>8)
	}
}

// 没有黑边时不能乱裁。
func TestTrimShadowNoOpWhenNoBorder(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	fill(img, img.Bounds(), color.RGBA{200, 200, 200, 255})
	got := TrimShadow(img)
	if got.Bounds().Dx() != 50 || got.Bounds().Dy() != 50 {
		t.Fatalf("不该裁：%v", got.Bounds())
	}
}

// 体检要能把"全白空图"和"右端黑列"这两种最像成功的失败指出来。
func TestCheckFlagsWhiteImageAndBlackColumns(t *testing.T) {
	white := image.NewRGBA(image.Rect(0, 0, 40, 30))
	fill(white, white.Bounds(), color.RGBA{255, 255, 255, 255})
	rep := Check(white)
	if len(rep.Problems) == 0 {
		t.Fatal("全白图必须报问题")
	}
	ok := image.NewRGBA(image.Rect(0, 0, 40, 30))
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			ok.SetRGBA(x, y, color.RGBA{uint8(x * 5), uint8(y * 5), 128, 255})
		}
	}
	rep2 := Check(ok)
	if len(rep2.Problems) != 0 {
		t.Fatalf("正常图不该报问题：%+v", rep2.Problems)
	}
	bad := image.NewRGBA(image.Rect(0, 0, 40, 30))
	fill(bad, bad.Bounds(), color.RGBA{10, 20, 30, 255})
	fill(bad, image.Rect(36, 0, 40, 30), color.RGBA{0, 0, 0, 255})
	rep3 := Check(bad)
	found := false
	for _, p := range rep3.Problems {
		if len(p) > 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("右端黑列必须被指出")
	}
}

// 窗口筛选：只留可见、尺寸合理的，按面积从大到小。
func TestPickWindowsFiltersAndSorts(t *testing.T) {
	wins := []Window{
		{PID: 7, Title: "small", Visible: true, Rect: Rect{W: 20, H: 20}},
		{PID: 7, Title: "hidden", Visible: false, Rect: Rect{W: 900, H: 700}},
		{PID: 9, Title: "other-pid", Visible: true, Rect: Rect{W: 900, H: 700}},
		{PID: 7, Title: "job", Visible: true, Rect: Rect{W: 900, H: 700}},
		{PID: 7, Title: "popup", Visible: true, Rect: Rect{W: 400, H: 300}},
	}
	got := PickWindows(wins, 7)
	if len(got) != 2 || got[0].Title != "job" || got[1].Title != "popup" {
		t.Fatalf("筛选/排序不对：%+v", got)
	}
	if _, err := FindWindow(got, "popup"); err != nil {
		t.Fatalf("按名字找不到：%v", err)
	}
	if _, err := FindWindow(got, "不存在"); err == nil {
		t.Fatal("找不到时必须报错")
	}
}
