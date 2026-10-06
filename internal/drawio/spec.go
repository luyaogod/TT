package drawio

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 排版规格（spec）：人（或 AI）只写「用哪个控件、放在第几列第几行、写什么字」，
// 查库、算栅格、平移坐标、重新编号、拼 XML 这些机械活全交给 compose。
//
// 这是本模块的核心分工：**模型不碰 XML**，就不会写坏转义，也不会撞 id；库改了也不用
// 重新教它。写 spec 的人只需要读形状清单（`tt drawio lib --catalog`）。
//
// 字段用指针的地方是"没写"与"写了零值"的区分：col 写 0 是合法的（第一列），
// x 写 0 也是合法的（贴左边缘）。

// Spec 是一份排版规格。
type Spec struct {
	// Library 用哪个库：`controls`（表单控件，缺省）或 `business`（业务组件）。
	Library string `json:"library"`
	Title   string `json:"title"`
	Grid    *Grid  `json:"grid"`
	Page    *Page  `json:"page"`
	Items   []SpecItem `json:"items"`
}

// Grid 是栅格参数。**用 col/row 排，不要自己算像素** —— 让模型做 `104 + 209*n`
// 这种加法，错是迟早的事。
type Grid struct {
	OriginX *float64 `json:"originX"`
	OriginY *float64 `json:"originY"`
	ColGap  *float64 `json:"colGap"`
	RowGap  *float64 `json:"rowGap"`
	// ColWidths 显式指定列宽；不写就取该列最宽控件的宽度。
	// 宽控件（TextEdit 是 314）会把整列撑开，那时用它钉死。
	ColWidths []float64 `json:"colWidths"`
}

// Page 是画布尺寸，缺省 850×1100。
type Page struct {
	Width  *float64 `json:"width"`
	Height *float64 `json:"height"`
}

// SpecItem 是图上的一个控件。
type SpecItem struct {
	Shape string `json:"shape"`
	// Col/Row 是**零基**的列号/行号，脚本自动算像素坐标。
	Col *int `json:"col"`
	Row *int `json:"row"`
	// X/Y 给了就无视栅格，直接绝对定位（整屏/分区布局用）。
	X *float64 `json:"x"`
	Y *float64 `json:"y"`
	// W/H 覆盖控件默认尺寸。拉伸行为见 compose 的三类子格。
	W *float64 `json:"w"`
	H *float64 `json:"h"`
	// Text 覆盖控件的文字，键是**文字槽位**名（见形状清单）。
	Text map[string]any `json:"text"`
	// Table 仅 Table：在 spec 里直接定义任意列行。
	Table *Table `json:"table"`
	// Pages/Active 仅 Folder：多页签。Pages 给全所有页名，Active（缺省 0）是当前页。
	Pages  []string `json:"pages"`
	Active *int     `json:"active"`
}

// ParseSpec 解析一份 spec 并做最基本的形状校验（能不能读、有没有内容）。
//
// 这里只拦"这份 spec 根本不成立"的错（空 items、缺 shape、列号为负）；
// 控件名对不对、槽位名对不对要等查库之后，那部分在 compose 里报 —— 那里的报错
// 能顺带列出可用值。
func ParseSpec(b []byte) (*Spec, error) {
	var s Spec
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("spec 不是合法 JSON：%w", err)
	}
	if len(s.Items) == 0 {
		return nil, fmt.Errorf("spec.items 必须是非空数组")
	}
	if s.Library == "" {
		s.Library = "controls"
	}
	for i, it := range s.Items {
		if strings.TrimSpace(it.Shape) == "" {
			return nil, fmt.Errorf("items[%d] 缺 shape 字段", i)
		}
		if it.Col != nil && *it.Col < 0 {
			return nil, fmt.Errorf("items[%d].col 不能是负数（列号零基）", i)
		}
		if it.Row != nil && *it.Row < 0 {
			return nil, fmt.Errorf("items[%d].row 不能是负数（行号零基）", i)
		}
	}
	return &s, nil
}
