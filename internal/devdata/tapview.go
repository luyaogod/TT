package devdata

// .tap 的点位正文视图：给"代码在、跑起来没有"这类问题用。
//
// 为什么必须单独有这一层：`.tap` 是设计源，`.4gl` 是产物。一个点位可以**只有声明、
// 没有槽位** —— 正文老老实实躺在 `.tap` 里，重产却永远不会把它拼进 `.4gl`。只查 `.4gl`
// 的工具（grep/at/show）在这种情况下会回答"没有"，而正确答案是"有，但它不生效"。
// 实测（axmt500_wf page 73）：`input.a.page73.xmdducseq` 就是这样，客户看到的却是
// "代码好像在、跑起来没有"。

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TapPointView 是 `.tap` 里一个声明点的正文。
type TapPointView struct {
	Name    string   `json:"name"`
	Src     string   `json:"src,omitempty"`
	Ver     string   `json:"ver,omitempty"`
	Status  string   `json:"status,omitempty"`
	Slot    string   `json:"slot,omitempty"`
	TagLine int      `json:"tagLine"` // `<point` 标签所在行（.tap 行号）
	Line0   int      `json:"line0,omitempty"`
	Line1   int      `json:"line1,omitempty"`
	Content []string `json:"content"`
	// HasSlot false = 骨架里没有这个点位的槽位（重产不会保留它的内容）。
	HasSlot bool `json:"hasSlot"`
}

// TapHit 是 `.tap` 里的一条命中。
type TapHit struct {
	Line    int    `json:"line"`
	Text    string `json:"text"`
	Point   string `json:"point,omitempty"`
	Src     string `json:"src,omitempty"`
	HasSlot bool   `json:"hasSlot"`
}

// TapSearchResult 是 `.tap` 检索结果。
type TapSearchResult struct {
	Prog    string   `json:"prog"`
	Env     string   `json:"env"`
	Pattern string   `json:"pattern"`
	Total   int      `json:"total"`
	Shown   int      `json:"shown"`
	Hits    []TapHit `json:"hits"`
}

// readTap 读材料里的 `.tap`（没有就报"材料不成套"）。
func readTap(req LinkRequest) ([]byte, *tapDoc, error) {
	b, err := readTapBytes(req)
	if err != nil {
		return nil, nil, err
	}
	tap, perr := parseTap(b)
	if perr != nil {
		return nil, nil, perr
	}
	return b, tap, nil
}

// readTapBytes 只读 `.tap` 字节（包内 .tap 也走它）。
func readTapBytes(req LinkRequest) ([]byte, error) {
	dir, err := req.matDir()
	if err != nil {
		return nil, err
	}
	mf, err := readPackManifest(dir, req.Prog)
	if err != nil {
		return nil, err
	}
	rel, _ := relOfRole(mf, "tap")
	if rel == "" {
		return nil, fmt.Errorf("%w: %s 没有 .tap（未被设计器接管）", ErrMaterialIncomplete, req.Prog)
	}
	b, rerr := os.ReadFile(filepath.Join(dir, rel))
	if rerr != nil {
		return nil, fmt.Errorf("读材料失败 (%s): %w", rel, rerr)
	}
	return b, nil
}

// slotNames 收集骨架里所有槽位名（判"这个声明有没有洞"）。
func slotNames(tap *tapDoc) map[string]bool {
	out := map[string]bool{}
	for _, s := range tap.Sections {
		for _, sl := range s.Slots {
			if sl.Name != "" {
				out[sl.Name] = true
			}
		}
	}
	return out
}

// TapPoint 按点名取 `.tap` 里的声明正文（精确匹配，大小写不敏感）。
//
// 找不到时返回 nil（不是错误）：调用方要能区分"没有这个点"与"读不动材料"。
func TapPoint(req LinkRequest, name string) (*TapPointView, error) {
	b, _, err := readTap(req)
	if err != nil {
		return nil, err
	}
	return TapPointBytes(b, name)
}

// TapPointBytes 在**给定的 .tap 字节**里按点名取声明正文。
// 单独开这个入口是为了包内 .tap（apply 之后想看"我的改动进包没有"）。
func TapPointBytes(data []byte, name string) (*TapPointView, error) {
	tap, err := parseTap(data)
	if err != nil {
		return nil, err
	}
	slots := slotNames(tap)
	for _, d := range tap.Decls {
		if !strings.EqualFold(d.Name, strings.TrimSpace(name)) {
			continue
		}
		lines := strings.Split(strings.TrimSuffix(d.Content, "\n"), "\n")
		v := &TapPointView{
			Name: d.Name, Src: d.Src, Ver: d.Ver, Status: d.Status, Slot: d.Slot,
			TagLine: d.Line, Content: lines, HasSlot: slots[d.Name],
		}
		if len(lines) > 0 && d.ContentLine > 0 {
			v.Line0 = d.ContentLine
			v.Line1 = d.ContentLine + len(lines) - 1
		}
		return v, nil
	}
	return nil, nil
}

// TapSearch 在**材料里的** `.tap` 里检索。
func TapSearch(req LinkRequest, pattern string, limit int, caseSensitive bool) (*TapSearchResult, error) {
	b, err := readTapBytes(req)
	if err != nil {
		return nil, err
	}
	return TapSearchBytes(req.Prog, req.Env, b, pattern, limit, caseSensitive)
}

// TapSearchBytes 在给定的 `.tap` 字节里检索（包内 .tap 走这个入口）。
func TapSearchBytes(prog, env string, b []byte, pattern string, limit int, caseSensitive bool) (*TapSearchResult, error) {
	tap, err := parseTap(b)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 50
	}
	lines := strings.Split(string(b), "\n")
	slots := slotNames(tap)
	// 声明按标签行排序，检索时顺序定位"这一行属于哪个声明"。
	decls := append([]tapDecl(nil), tap.Decls...)
	sort.Slice(decls, func(i, j int) bool { return decls[i].Line < decls[j].Line })

	needle := pattern
	if !caseSensitive {
		needle = strings.ToLower(pattern)
	}
	res := &TapSearchResult{Prog: prog, Env: env, Pattern: pattern}
	for i, ln := range lines {
		hay := ln
		if !caseSensitive {
			hay = strings.ToLower(ln)
		}
		if !strings.Contains(hay, needle) {
			continue
		}
		res.Total++
		if len(res.Hits) >= limit {
			continue
		}
		lineNo := i + 1
		hit := TapHit{Line: lineNo, Text: strings.TrimSpace(ln)}
		for _, d := range decls {
			if d.Line <= lineNo {
				hit.Point, hit.Src = d.Name, d.Src
				hit.HasSlot = slots[d.Name]
			} else {
				break
			}
		}
		res.Hits = append(res.Hits, hit)
	}
	res.Shown = len(res.Hits)
	return res, nil
}
