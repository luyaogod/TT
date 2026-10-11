package view

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestRegionSetRoundTripAndReplace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "regions.json")
	var rs RegionSet
	rs.Put(Region{Name: "明细料号(第1行)", Rect: image.Rect(100, 400, 220, 420), RectTxt: "100,400,120,20", At: "2026-10-11T13:00:00Z", Source: "注入 料号=LUYAOWF888"})
	rs.Put(Region{Name: "明细订购量(第1行)", Rect: image.Rect(300, 400, 400, 420), RectTxt: "300,400,100,20", At: "2026-10-11T13:00:01Z"})
	if err := SaveRegions(path, rs); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("临时文件没清掉")
	}
	got, err := LoadRegions(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Regions) != 2 {
		t.Fatalf("区域数 = %d", len(got.Regions))
	}
	if r, ok := got.Get("明细订购量(第1行)"); !ok || r.Rect.Dx() != 100 {
		t.Fatalf("读回来的矩形不对：%+v", r)
	}
	// 重名覆盖
	rs.Put(Region{Name: "明细料号(第1行)", Rect: image.Rect(1, 2, 3, 4), RectTxt: "1,2,2,2", At: "later"})
	if err := SaveRegions(path, rs); err != nil {
		t.Fatal(err)
	}
	got2, _ := LoadRegions(path)
	if len(got2.Regions) != 2 {
		t.Fatalf("重名应覆盖而不是追加：%d", len(got2.Regions))
	}
	if r, _ := got2.Get("明细料号(第1行)"); r.Rect.Min.X != 1 {
		t.Fatalf("覆盖没生效：%+v", r)
	}
}

func TestParseRectRejectsBadInput(t *testing.T) {
	if _, err := ParseRect("1,2,3"); err == nil {
		t.Fatal("缺一个数必须报错")
	}
	if _, err := ParseRect("1,2,0,5"); err == nil {
		t.Fatal("宽为 0 必须报错")
	}
	if r, err := ParseRect("10,20,30,40"); err != nil || r != image.Rect(10, 20, 40, 60) {
		t.Fatalf("正常输入解错：%v %v", r, err)
	}
}
