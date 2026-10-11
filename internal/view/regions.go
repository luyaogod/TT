package view

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
)

// 命名区域：给"协议里没有坐标的东西"一条正规出路。
//
// 明细网格的格子就是这样：TableColumn 与运行时的 Value **都没有 posX/posY**
// （真机实测：带几何的 Value 对象数为 0）。没有坐标就只能**量** —— 往那个格子写一个值、
// 拍前后两张图、取像素差，量出来的矩形就是它在屏幕上的位置。
//
// 量出来的是"这一次测量的事实"，所以每条区域都带着**来源与时间**：
// 换主题/换分辨率/换作业之后它可能失效，而失效时必须能看出来（别悄悄画错框）。
type Region struct {
	Name    string          `json:"name"`
	Rect    image.Rectangle `json:"-"`
	RectTxt string          `json:"rect"`
	At      string          `json:"at"`
	Source  string          `json:"source,omitempty"` // 怎么量出来的（注入的字段/值）
	Note    string          `json:"note,omitempty"`
}

func (r Region) Bounds() image.Rectangle { return r.Rect }

// ParseRect reads "x,y,w,h".
func ParseRect(s string) (image.Rectangle, error) {
	var x, y, w, h int
	if _, err := fmt.Sscanf(s, "%d,%d,%d,%d", &x, &y, &w, &h); err != nil {
		return image.Rectangle{}, fmt.Errorf("view: 矩形要写成 x,y,w,h：%q", s)
	}
	if w <= 0 || h <= 0 {
		return image.Rectangle{}, fmt.Errorf("view: 矩形宽高要为正：%q", s)
	}
	return image.Rect(x, y, x+w, y+h), nil
}

// RegionSet is the saved set of measured regions.
type RegionSet struct {
	Regions []Region `json:"regions"`
}

// LoadRegions reads a regions file (missing file = empty set).
func LoadRegions(path string) (RegionSet, error) {
	var rs RegionSet
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return rs, nil
		}
		return rs, err
	}
	if err := json.Unmarshal(b, &rs); err != nil {
		return rs, err
	}
	for i := range rs.Regions {
		if r, perr := ParseRect(rs.Regions[i].RectTxt); perr == nil {
			rs.Regions[i].Rect = r
		}
	}
	return rs, nil
}

// SaveRegions writes the set atomically.
func SaveRegions(path string, rs RegionSet) error {
	sort.Slice(rs.Regions, func(i, j int) bool { return rs.Regions[i].Name < rs.Regions[j].Name })
	b, err := json.MarshalIndent(rs, "", " ")
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

// Put adds or replaces one region (by name).
func (rs *RegionSet) Put(r Region) {
	for i := range rs.Regions {
		if rs.Regions[i].Name == r.Name {
			rs.Regions[i] = r
			return
		}
	}
	rs.Regions = append(rs.Regions, r)
}

// Get finds one region by name.
func (rs RegionSet) Get(name string) (Region, bool) {
	for _, r := range rs.Regions {
		if r.Name == name {
			return r, true
		}
	}
	return Region{}, false
}
