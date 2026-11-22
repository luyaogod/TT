package devdata

import (
	"fmt"
	"sort"
)

// brief：一个程序的"一页纸画像"，给刚接到任务的 AI/人做第一眼判断。
//
// 为什么不是把 link.json 直接扔过去：索引 800 KB、3,215 个点，读它等于不读。
// brief 只回答五个问题（顺序就是判断顺序）：
//
//  1. 这是什么程序 —— 判型/模块/客制身份/客制版次/是否分段落版型
//  2. 有哪些页签 —— 页序 / page_id / screen record / 栏位数 / 有逻辑的字段数
//  3. 客制逻辑在哪 —— 客制点按行数降序（重产会保留的只有这些）
//  4. 有什么风险 —— 一致性检查里的 warn/error、模板脱版、.4gl 比 .tap 旧
//  5. 下一步怎么查 —— 意图清单 + 三条命令
//
// 所有数字都来自同一次索引，不做二次推断；查不到就说查不到（notes）。

// BriefRequest 是一次画像请求。
type BriefRequest struct {
	LinkRequest
	// TopPoints 客制点最多列几个（默认 15）。
	TopPoints int
}

// BriefResult 是画像结果（字段都带 json tag，CLI 的 --json 直接吐它）。
type BriefResult struct {
	Prog     string        `json:"prog"`
	Env      string        `json:"env"`
	Dir      string        `json:"dir"`
	Mode     LinkMode      `json:"mode"`
	Stats    LinkStats     `json:"stats"`
	Pages    []LinkPage    `json:"pages"`
	Sections []LinkSection `json:"sections"`
	// CustomPoints 客制点（按声明行数降序，重产会保留的就是这些）。
	CustomPoints []LinkPoint `json:"customPoints"`
	// CustomTotal 客制点总数（CustomPoints 可能被 TopPoints 截断）。
	CustomTotal int          `json:"customTotal"`
	Orphans     []LinkOrphan `json:"orphans,omitempty"`
	// OrphansCustom 客制孤儿数（真的丢了客户代码的那些）。
	OrphansCustom int         `json:"orphansCustom"`
	Slices        LinkSlices  `json:"slices"`
	Checks        []LinkCheck `json:"checks"`
	Intents       []Intent    `json:"intents"`
	Notes         []string    `json:"notes"`
	Link          string      `json:"link"`
}

// Brief 生成画像。
func Brief(req BriefRequest) (*BriefResult, error) {
	doc, err := LoadIndex(req.LinkRequest)
	if err != nil {
		return nil, err
	}
	top := req.TopPoints
	if top <= 0 {
		top = 15
	}
	res := &BriefResult{
		Prog: doc.Prog, Env: doc.Env, Dir: req.Dir,
		Mode: doc.Mode, Stats: doc.Stats,
		Pages: doc.Pages, Sections: doc.Sections,
		Orphans: doc.Orphans, Slices: doc.Slices,
		Checks: doc.Checks, Intents: Intents, Link: doc.Path,
	}
	// 客制点：只要"有正文"的（src=c 且非空），按行数降序 —— 行数就是"这块逻辑多厚"。
	for _, p := range doc.Points {
		if !p.Custom || p.DeclLines == 0 {
			continue
		}
		res.CustomPoints = append(res.CustomPoints, p)
	}
	res.CustomTotal = len(res.CustomPoints)
	for _, o := range doc.Orphans {
		if o.Src == "c" {
			res.OrphansCustom++
		}
	}
	sortPointsByLines(res.CustomPoints)
	if len(res.CustomPoints) > top {
		res.CustomPoints = res.CustomPoints[:top]
		res.Notes = append(res.Notes,
			fmt.Sprintf("客制点共 %d 个，只列最大 %d 个（--top 调大；全量看 tt dev search / map --custom）",
				res.CustomTotal, top))
	}
	if len(doc.Pages) == 0 {
		res.Notes = append(res.Notes, "没有页面信息：缺 .tab（蓝图）")
	}
	if doc.Slices.Drift != nil {
		res.Notes = append(res.Notes, "有切片脱版：产物是旧版模板生成的，动这些段前先确认重产影响")
	}
	res.Notes = append(res.Notes, "下一步：tt dev find --intent <意图> [--field <字段>] / tt dev at <点名|行号> / tt dev search <子串>")
	// JSON 数组的空值必须是 [] 而不是 null（agent 侧按数组消费，null 会炸）。
	res.Pages = nonNil(res.Pages)
	res.Sections = nonNil(res.Sections)
	res.CustomPoints = nonNil(res.CustomPoints)
	res.Checks = nonNil(res.Checks)
	res.Intents = nonNil(res.Intents)
	res.Notes = nonNil(res.Notes)
	return res, nil
}

// nonNil 把 nil 切片换成空切片（只在 JSON 落盘/输出这一层用：nil → null 会让
// 按数组消费的脚本失败，而空数组与"没有"在语义上是同一件事）。
func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// sortPointsByLines 按声明行数降序（同行的按名字），用来回答"哪块客制最厚"。
func sortPointsByLines(ps []LinkPoint) {
	sort.SliceStable(ps, func(i, j int) bool {
		if ps[i].DeclLines != ps[j].DeclLines {
			return ps[i].DeclLines > ps[j].DeclLines
		}
		return ps[i].Name < ps[j].Name
	})
}
