package devdata

// link.json 之上的四条只读查询：show（点名 → 内容）、at（行号 → 归属）、
// fields（字段/页面清单）、grep（在产物里检索并标出归属）。
//
// 为什么值得有：索引本身 800 KB，AI 不可能整份读；真正的问题总是"这一个点/这一行是什么"。
// 四条命令都把答案压到几十行以内，并且**默认复用已建索引**（源文件 sha 未变就跳过重建）。

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/dev/pkgfile"
)

// ---------- 索引加载（带 sha 复用） ----------

// LoadIndex 取索引：源文件三份 sha 与 link.json 里记的一致就直接用，否则重建。
//
// 实测：每次重建要解析 3.6 MB 的 .tap + 3 MB 的 .4gl，约 5 秒；而 sha 只需读文件（毫秒级）。
// 交互式问答里 5 秒/次是明显负担，所以这条快路径值得。
func LoadIndex(req LinkRequest) (*LinkDoc, error) {
	progDir, err := req.matDir()
	if err != nil {
		return nil, err
	}
	mf, err := readPackManifest(progDir, req.Prog)
	if err != nil {
		return nil, err
	}
	shas, err := sourceSHAs(progDir, mf)
	if err != nil {
		return nil, err
	}
	linkPath := req.indexPath(progDir)
	if b, rerr := os.ReadFile(linkPath); rerr == nil {
		var d LinkDoc
		if json.Unmarshal(b, &d) == nil && d.Schema == LinkSchema && sameSources(&d, shas) {
			d.Path = linkPath
			return &d, nil
		}
	}
	return Map(req)
}

// ReadLinkDoc 只读索引文件（不校验新鲜度、不重建）。文件不存在返回 (nil, nil)。
//
// 供 status 这类"只想知道有没有、多大"的地方用 —— 它们不该为了显示摘要触发一次重建。
func ReadLinkDoc(path string) (*LinkDoc, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("读索引失败 (%s): %w", path, err)
	}
	var d LinkDoc
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("索引不是合法 JSON (%s): %w", path, err)
	}
	d.Path = path
	return &d, nil
}

// IndexFresh 报告索引是否与当前材料一致（只读 sha，不重建）。
//
// 供 status 用：重建要 5 秒，而"索引该不该重建"只需要毫秒级的 sha 比对。
// 返回 false 且 err=nil 表示"需要重建"（索引不存在/版本不符/源文件变了）。
func IndexFresh(req LinkRequest) (bool, error) {
	progDir, err := req.matDir()
	if err != nil {
		return false, err
	}
	mf, err := readPackManifest(progDir, req.Prog)
	if err != nil {
		return false, err
	}
	shas, err := sourceSHAs(progDir, mf)
	if err != nil {
		return false, err
	}
	b, err := os.ReadFile(req.indexPath(progDir))
	if err != nil {
		return false, nil
	}
	var d LinkDoc
	if json.Unmarshal(b, &d) != nil {
		return false, nil
	}
	return d.Schema == LinkSchema && sameSources(&d, shas), nil
}

// sourceSHAs 算 .tab/.tsd/.tap/.4gl 的 sha256（缺的留空）。
func sourceSHAs(progDir string, mf *Manifest) (map[string]string, error) {
	out := map[string]string{}
	for _, role := range []string{"tab", "tsd", "tap", "4gl"} {
		rel, _ := relOfRole(mf, role)
		if rel == "" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(progDir, rel))
		if err != nil {
			return nil, fmt.Errorf("读材料失败 (%s): %w", rel, err)
		}
		out[role] = hashOf(b)
	}
	return out, nil
}

// sameSources 比对索引里记的 sha 与当前文件。
func sameSources(d *LinkDoc, shas map[string]string) bool {
	cmp := func(recorded, role string) bool {
		want := shas[role]
		if want == "" {
			return recorded == ""
		}
		return recorded == want
	}
	if !cmp(d.Sources.Tap.SHA256, "tap") || !cmp(d.Sources.G4.SHA256, "4gl") {
		return false
	}
	tabSHA, tsdSHA := "", ""
	if d.Sources.Tab != nil {
		tabSHA = d.Sources.Tab.SHA256
	}
	if d.Sources.Tsd != nil {
		tsdSHA = d.Sources.Tsd.SHA256
	}
	return cmp(tabSHA, "tab") && cmp(tsdSHA, "tsd")
}

// loadForQuery 取索引 + .4gl 原文。
func loadForQuery(req LinkRequest) (*LinkDoc, []string, error) {
	doc, err := LoadIndex(req)
	if err != nil {
		return nil, nil, err
	}
	progDir, err := req.matDir()
	if err != nil {
		return nil, nil, err
	}
	mf, err := readPackManifest(progDir, req.Prog)
	if err != nil {
		return nil, nil, err
	}
	rel, _ := relOfRole(mf, "4gl")
	b, err := os.ReadFile(filepath.Join(progDir, rel))
	if err != nil {
		return nil, nil, fmt.Errorf("读材料失败 (%s): %w", rel, err)
	}
	return doc, splitLines(b), nil
}

// ---------- show ----------

// ShowRequest 查一个点位。
type ShowRequest struct {
	LinkRequest
	Query    string // 点名（可只给字段名或点名的后缀）
	MaxLines int    // 内容最多打印多少行（0 = 默认 200）
}

// ShownPoint 是一个点位的完整视图（含内容）。
type ShownPoint struct {
	LinkPoint
	// Label 该点对应的画面中文名（.tsd 的 lbl_<字段>；没有就空）。
	Label string `json:"label,omitempty"`
	// Orphan true = 这个点只有 .tap 声明、骨架里没有槽位（重产不会保留）。
	// 正文来自 .tap（Line0/Line1 是 .tap 行号），不是 .4gl。
	Orphan       bool     `json:"orphan,omitempty"`
	OrphanReason string   `json:"orphanReason,omitempty"`
	Content      []string `json:"content"`
	Truncated    bool     `json:"truncated,omitempty"`
	SliceCodes   []string `json:"sliceCodes,omitempty"` // 这段里（或其前一条）出现的切片
}

// ShowResult 是 show 的结果。
type ShowResult struct {
	Prog   string       `json:"prog"`
	Env    string       `json:"env"`
	Query  string       `json:"query"`
	Match  string       `json:"match"` // exact / suffix / substring
	Points []ShownPoint `json:"points"`
	Note   string       `json:"note,omitempty"`
	Link   string       `json:"link"`
}

// Show 按点名（或其后缀/画面中文名）取点位内容。
//
// 匹配规则：query 里带 `.` 的按**点名**处理，只接受精确匹配（否则 `xmdducseq`
// 会静默命中 `xmdducseq1`，让人以为"写了却没生效的点其实有代码"）；
// 不带 `.` 的先按字段名（后缀/子串），查不到再按**画面中文名**（.tsd 标签）。
// Match 里写明是哪种，标签命中时 Note 会给出对应列名。
func Show(req ShowRequest) (*ShowResult, error) {
	doc, lines, err := loadForQuery(req.LinkRequest)
	if err != nil {
		return nil, err
	}
	// 孤儿优先：`xmdducseq` 与 `xmdducseq1` 只差一个字符，而前者是"有代码但不生效"的
	// 孤儿声明。这时给出后者（甚至建议照抄）是最坏的一种"看起来对"。
	q := strings.TrimSpace(req.Query)
	dotted := strings.Contains(q, ".")
	if sp, level := orphanShown(req.LinkRequest, doc, q); sp != nil {
		if level == "exact" || (!dotted && level == "suffix") {
			return orphanResult(req, doc, q, sp), nil
		}
	}

	pts, how := matchPoints(doc, q)
	if len(pts) == 0 && !dotted {
		// 字段名不中，试画面中文名（"刻号" → xmdduc003）。
		if cols, by, _ := ResolveColumns(doc, q); by == "label" && len(cols) > 0 {
			seen := map[string]bool{}
			for _, p := range doc.Points {
				if !containsFold(cols, columnOfPoint(p.Name)) || seen[p.Name] {
					continue
				}
				seen[p.Name] = true
				pts = append(pts, p)
			}
			how = "label"
		}
	}
	res := &ShowResult{Prog: req.Prog, Env: req.Env, Query: q, Match: how, Link: doc.Path,
		Points: []ShownPoint{}}
	if len(pts) == 0 {
		// 最后一路：孤儿点的子串命中（前面已放行 exact/suffix）。
		if sp, _ := orphanShown(req.LinkRequest, doc, q); sp != nil {
			return orphanResult(req, doc, q, sp), nil
		}
		return nil, fmt.Errorf("%w: %q（字段名或画面中文名都对不上；tt dev fields %s 可看清单）",
			ErrPointNotFound, q, req.Prog)
	}
	switch how {
	case "label":
		var cols []string
		for _, p := range pts {
			cols = append(cols, columnOfPoint(p.Name))
		}
		res.Note = fmt.Sprintf("%q 是画面中文名，按列 %s 找到 %d 个点", q,
			strings.Join(dedupeStrings(cols), ", "), len(pts))
	case "suffix", "substring":
		res.Note = fmt.Sprintf("按%s命中 %d 个点（不是精确点名）", map[string]string{
			"suffix": "字段名后缀", "substring": "子串",
		}[how], len(pts))
	}
	if dotted && how != "exact" {
		return nil, fmt.Errorf("%w: %q 看着像完整点名，但只匹配到近似的 %q；"+
			"完整点名请照抄（tt dev fields 可看），按字段名查就别带点",
			ErrPointNotFound, q, pts[0].Name)
	}

	max := req.MaxLines
	if max <= 0 {
		max = 200
	}
	for _, p := range pts {
		sp := ShownPoint{LinkPoint: p, Label: labelOfColumn(doc, columnOfPoint(p.Name))}
		if p.Line0 > 0 && p.Line1 >= p.Line0 && p.Line0 <= len(lines) {
			end := p.Line1
			if end > len(lines) {
				end = len(lines)
			}
			body := lines[p.Line0-1 : end]
			if len(body) > max {
				body = body[:max]
				sp.Truncated = true
			}
			sp.Content = body
			sp.SliceCodes = sliceCodesIn(doc, p.Line0, p.Line1)
		}
		res.Points = append(res.Points, sp)
	}
	return res, nil
}

// orphanResult 把孤儿声明包成 Show 的结果（正文来自 .tap）。
func orphanResult(req ShowRequest, doc *LinkDoc, q string, sp *ShownPoint) *ShowResult {
	return &ShowResult{
		Prog: req.Prog, Env: req.Env, Query: q, Match: "orphan", Link: doc.Path,
		Points: []ShownPoint{*sp},
		Note: fmt.Sprintf("%q 只有 .tap 声明、骨架里没有槽位：**重产不会保留它**，"+
			"运行期也不会执行（正文来自 .tap，行号是 .tap 行号）%s", sp.Name, orphanSiblingHint(doc, sp.Name)),
	}
}

// orphanSiblingHint 给一句"同页相近的控件是什么" —— 孤儿最常见的成因就是控件改名/删列，
// 只说"没有槽位"不够，得让人知道这一页上真正能改的是哪个点。
func orphanSiblingHint(doc *LinkDoc, orphan string) string {
	m := rePointPage.FindStringSubmatch(orphan)
	if m == nil {
		return ""
	}
	pageID, col := m[1], m[2]
	var sib []string
	for _, f := range doc.Fields {
		if f.PageID != pageID || f.Column == col {
			continue
		}
		if strings.HasPrefix(f.Column, col) || strings.HasPrefix(col, f.Column) {
			label := ""
			if f.Label != "" {
				label = "（" + f.Label + "）"
			}
			sib = append(sib, fmt.Sprintf("%s%s", f.Column, label))
		}
	}
	if len(sib) == 0 {
		return ""
	}
	return "；同页相近的控件是 " + strings.Join(dedupeStrings(sib), ", ") + "，要改画面上的那个请用它"
}

// ResolveColumns 把"字段查询"解析成一组字段名。
//
// 先按字段名（精确 → 后缀 → 子串），再按**画面标签**（.tsd 的 `lbl_<字段>` →
// LinkField.Label）。加标签这一路是因为现场的语言是"刻号"，而代码里叫 xmdduc003 ——
// 让用户先去翻 .tsd 查列名，等于把工具该干的活推回给人（实测两个试用者都卡在这一步）。
//
// 返回：命中的列名（去重、保序）、依据（column / label）、以及这些列的中文名。
func ResolveColumns(doc *LinkDoc, q string) (cols []string, by string, labels map[string]string) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, "", nil
	}
	lq := strings.ToLower(q)
	labels = map[string]string{}
	// 只做精确与后缀，**不做列名子串**：`xmdducseq` 是 `xmdducseq1` 的子串，
	// 子串匹配会把"孤儿字段"答成"另一个活着的字段"（实测差一个字符就是两种结论）。
	var exact, suffix []string
	for _, f := range doc.Fields {
		c := strings.ToLower(f.Column)
		switch {
		case c == lq:
			exact = append(exact, f.Column)
		case strings.HasSuffix(c, lq):
			suffix = append(suffix, f.Column)
		}
	}
	pick := func(list []string, kind string) ([]string, string, map[string]string) {
		if len(list) == 0 {
			return nil, "", nil
		}
		out := dedupeStrings(list)
		lm := map[string]string{}
		for _, f := range doc.Fields {
			for _, c := range out {
				if f.Column == c && f.Label != "" {
					lm[c] = f.Label
				}
			}
		}
		return out, kind, lm
	}
	if got, kind, lm := pick(exact, "column"); got != nil {
		return got, kind, lm
	}
	if got, kind, lm := pick(suffix, "column"); got != nil {
		return got, kind, lm
	}

	// 字段名都不中，才按标签找：精确 → 子串（标签通常很短，"刻号"这种精确居多）。
	var lexact, lsub []string
	for _, f := range doc.Fields {
		if f.Label == "" {
			continue
		}
		if strings.EqualFold(f.Label, q) {
			lexact = append(lexact, f.Column)
		} else if strings.Contains(f.Label, q) {
			lsub = append(lsub, f.Column)
		}
	}
	if got, _, lm := pick(lexact, "label"); got != nil {
		return got, "label", lm
	}
	if got, _, lm := pick(lsub, "label"); got != nil {
		return got, "label", lm
	}
	return nil, "", nil
}

// labelOfColumn 取某个列的画面中文名（没有就空串）。
func labelOfColumn(doc *LinkDoc, col string) string {
	if col == "" || doc == nil {
		return ""
	}
	for _, f := range doc.Fields {
		if f.Column == col {
			return f.Label
		}
	}
	return ""
}

// dedupeStrings 保序去重。
func dedupeStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// ColumnOfPoint 取点名里的字段（`input.a.page73.xmdduc001` → xmdduc001）。
func ColumnOfPoint(name string) string { return columnOfPoint(name) }

// columnOfPoint 取点名里的字段（`input.a.page73.xmdduc001` → xmdduc001）。
func columnOfPoint(name string) string {
	if m := rePointPage.FindStringSubmatch(name); m != nil {
		return m[2]
	}
	return ""
}

// matchPoints 按精确 → 后缀 → 子串三级匹配（都忽略大小写）。
func matchPoints(doc *LinkDoc, q string) ([]LinkPoint, string) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, ""
	}
	lq := strings.ToLower(q)
	var exact, suffix, sub []LinkPoint
	for _, p := range doc.Points {
		n := strings.ToLower(p.Name)
		switch {
		case n == lq:
			exact = append(exact, p)
		case strings.HasSuffix(n, "."+lq):
			suffix = append(suffix, p)
		case strings.Contains(n, lq):
			sub = append(sub, p)
		}
	}
	switch {
	case len(exact) > 0:
		return sortPoints(exact), "exact"
	case len(suffix) > 0:
		return sortPoints(suffix), "suffix"
	default:
		return sortPoints(sub), "substring"
	}
}

// MatchPoints 供 CLI 复用同一套匹配规则（精确 → 后缀 → 子串）。
func MatchPoints(doc *LinkDoc, q string) ([]LinkPoint, string) { return matchPoints(doc, q) }

func sortPoints(ps []LinkPoint) []LinkPoint {
	out := append([]LinkPoint(nil), ps...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].Line0, out[j].Line0
		if a == 0 || b == 0 {
			return b == 0 && a != 0
		}
		return a < b
	})
	return out
}

// sliceCodesIn 取某行区间内出现过的切片代号（去重、保序）。
func sliceCodesIn(doc *LinkDoc, l0, l1 int) []string {
	seen := map[string]bool{}
	var out []string
	add := func(code string) {
		if !seen[code] {
			seen[code] = true
			out = append(out, code)
		}
	}
	var before string
	for _, m := range doc.Slices.Markers {
		if m.Line < l0 {
			before = m.Code
			continue
		}
		if m.Line > l1 {
			break
		}
		add(m.Code)
	}
	if len(out) == 0 && before != "" {
		add(before)
	}
	return out
}

// ---------- at ----------

// AtRequest 查一行属于哪个点/段/切片。
type AtRequest struct {
	LinkRequest
	Line    int
	Context int // 附带几行上下文（0 = 默认 2）
}

// AtResult 是一行的归属。
type AtResult struct {
	Prog string `json:"prog"`
	Env  string `json:"env"`
	Line int    `json:"line"`
	// Inside=true 表示落在某个加点内（重产会保留内容）。
	Inside  bool         `json:"inside"`
	Point   *LinkPoint   `json:"point,omitempty"`
	Section *LinkSection `json:"section,omitempty"`
	Slice   *g4Slice     `json:"slice,omitempty"`
	// SliceRelation：in-point（切片标记在这个点内）/ before-line（只是该行之前最近的）
	SliceRelation string   `json:"sliceRelation,omitempty"`
	Verdict       string   `json:"verdict"`
	Stale         bool     `json:"stale,omitempty"` // .4gl 比 .tap 旧：行号可能已漂
	StaleNote     string   `json:"staleNote,omitempty"`
	Content       []string `json:"content,omitempty"`
	// ContentStart 是 Content[0] 对应的行号（= Line-context，至少 1）。
	// 少这个字段就会把上下文行标成 `Line` 那几行 —— 实测整段行号整体偏移，行号定位直接失效。
	ContentStart int    `json:"contentStart,omitempty"`
	Link         string `json:"link"`
	TotalLine    int    `json:"totalLines"`
}

// At 反查行号。
func At(req AtRequest) (*AtResult, error) {
	doc, lines, err := loadForQuery(req.LinkRequest)
	if err != nil {
		return nil, err
	}
	if req.Line < 1 || req.Line > len(lines) {
		return nil, fmt.Errorf("%w: 行号 %d 超出范围（1-%d）", ErrInvalidOpt, req.Line, len(lines))
	}
	res := &AtResult{Prog: req.Prog, Env: req.Env, Line: req.Line, Link: doc.Path, TotalLine: len(lines)}
	res.Stale, res.StaleNote = staleNote(doc)

	// 点：区间包含该行（多个命中取区间最小者 —— 内层更具体）
	for i := range doc.Points {
		p := &doc.Points[i]
		if p.Line0 == 0 || req.Line < p.Line0 || req.Line > p.Line1 {
			continue
		}
		if res.Point == nil || (p.Line1-p.Line0) < (res.Point.Line1-res.Point.Line0) {
			res.Point = p
		}
	}
	for i := range doc.Sections {
		s := &doc.Sections[i]
		if s.Line0 > 0 && req.Line >= s.Line0 && req.Line <= s.Line1 {
			res.Section = s
			break
		}
	}
	// 切片：优先"这个点区间内"的第一条（能说明这段是谁生成的），
	// 其次才是该行之前最近的一条（并标明只是"之前"）。
	var before *g4Slice
	for i := range doc.Slices.Markers {
		m := doc.Slices.Markers[i]
		if res.Point != nil && m.Line >= res.Point.Line0 && m.Line <= res.Point.Line1 {
			res.Slice = &g4Slice{Code: m.Code, Ver: m.Ver, Line: m.Line}
			res.SliceRelation = "in-point"
			break
		}
		if m.Line <= req.Line {
			before = &g4Slice{Code: m.Code, Ver: m.Ver, Line: m.Line}
		}
	}
	if res.Slice == nil && before != nil {
		res.Slice = before
		res.SliceRelation = "before-line"
	}

	ctx := req.Context
	if ctx <= 0 {
		ctx = 2
	}
	lo := req.Line - ctx
	if lo < 1 {
		lo = 1
	}
	hi := req.Line + ctx
	if hi > len(lines) {
		hi = len(lines)
	}
	res.Content = lines[lo-1 : hi]
	res.ContentStart = lo

	// 结论：一句话说清"改它安不安全"
	switch {
	case res.Point != nil && res.Point.Unmarked && res.Point.Custom:
		res.Inside = true
		res.Verdict = fmt.Sprintf("在客制函数 %s 内（第 %d-%d 行，无加点标记）——改这段重产会保留",
			res.Point.Name, res.Point.Line0, res.Point.Line1)
	case res.Point != nil && res.Point.Custom:
		res.Inside = true
		res.Verdict = fmt.Sprintf("在客制加点 %s 内（第 %d-%d 行）——改这段重产会保留",
			res.Point.Name, res.Point.Line0, res.Point.Line1)
	case res.Point != nil && res.Point.Unmarked:
		res.Inside = true
		res.Verdict = fmt.Sprintf("在函数 %s 内（第 %d-%d 行，无加点标记，来源标准侧）",
			res.Point.Name, res.Point.Line0, res.Point.Line1)
	case res.Point != nil && res.Point.Extra:
		res.Verdict = fmt.Sprintf("在生成器自加的槽位 %s 内（被注释掉的重复点）——不是你的代码",
			res.Point.Name)
	case res.Point != nil && res.Point.G4Only:
		res.Inside = true
		res.Verdict = fmt.Sprintf("在加点 %s 内（第 %d-%d 行）——该程序没有 .tap，无法判断内容是客制还是标准",
			res.Point.Name, res.Point.Line0, res.Point.Line1)
	case res.Point != nil:
		res.Inside = true
		res.Verdict = fmt.Sprintf("在标准加点 %s 内（第 %d-%d 行）——内容来自标准/模板，重产按模板重建",
			res.Point.Name, res.Point.Line0, res.Point.Line1)
	case res.Section != nil:
		res.Verdict = fmt.Sprintf("在 section %s 的正文里（不在任何加点内）——**模板生成的框架代码，手改会被重产覆盖**",
			res.Section.ID)
	default:
		res.Verdict = "不在任何 section 内（文件头/尾或非生成器产物）"
	}
	return res, nil
}

// staleNote 说明 .4gl 是否比 .tap 旧（有旧值时给出人读提示）。
func staleNote(doc *LinkDoc) (bool, string) {
	tap, err1 := time.Parse(time.RFC3339, doc.Sources.Tap.ModTime)
	g4, err2 := time.Parse(time.RFC3339, doc.Sources.G4.ModTime)
	if err1 != nil || err2 != nil {
		return false, ""
	}
	if g4.Before(tap) {
		return true, fmt.Sprintf("注意：.4gl（%s）比 .tap（%s）旧，可能没跑 F7，行号/内容未必是当前规格",
			doc.Sources.G4.ModTime, doc.Sources.Tap.ModTime)
	}
	return false, ""
}

// ---------- fields / pages ----------

// FieldsRequest 列字段。
type FieldsRequest struct {
	LinkRequest
	Page    string // page_id 或页签序号（两者都接受，见 Fields）
	Pattern string // 按字段名子串过滤
}

// Fields 列蓝图字段（已与 .tab 对齐）。
func Fields(req FieldsRequest) ([]LinkField, *LinkDoc, error) {
	doc, _, err := loadForQuery(req.LinkRequest)
	if err != nil {
		return nil, nil, err
	}
	pageID := strings.TrimSpace(req.Page)
	note := ""
	if pageID != "" {
		// 人最容易传页签序号（12），而不是 page_id（73）。先按 page_id 试，
		// 没有再按序号试，并明确告诉调用方"你传的是序号"。
		if !docHasPageID(doc, pageID) {
			if id, ok := pageIDOfSeq(doc, pageID); ok {
				note = fmt.Sprintf("你传的是页签序号 %s，对应的 page_id 是 %s（点名用的是 page_id）", pageID, id)
				pageID = id
			}
		}
	}
	out := make([]LinkField, 0, len(doc.Fields))
	p := strings.ToLower(strings.TrimSpace(req.Pattern))
	for _, f := range doc.Fields {
		if pageID != "" && f.PageID != pageID {
			continue
		}
		if p != "" && !strings.Contains(strings.ToLower(f.Column), p) {
			continue
		}
		out = append(out, f)
	}
	if pageID != "" && len(out) == 0 {
		if note == "" {
			note = fmt.Sprintf("page_id=%s 上没有带加点的字段；可用 --pages 看有哪些页", pageID)
		}
	}
	if note != "" {
		doc.Notes = append(doc.Notes, note)
	}
	return out, doc, nil
}

func docHasPageID(doc *LinkDoc, id string) bool {
	for _, p := range doc.Pages {
		if p.PageID == id {
			return true
		}
	}
	return false
}

func pageIDOfSeq(doc *LinkDoc, seq string) (string, bool) {
	for _, p := range doc.Pages {
		if p.Page == seq {
			return p.PageID, true
		}
	}
	return "", false
}

// ---------- grep ----------

// GrepRequest 在 .4gl 里按子串检索，并把命中标上归属。
type GrepRequest struct {
	LinkRequest
	Pattern string
	Limit   int  // 默认 50；命中总数照样给
	Case    bool // true = 区分大小写
	// Scope 检索范围：4gl（产物，缺省）| tap（材料里的 .tap）| pkg（工作区包内的 .tap）| all。
	// tap 这一路不可少：orphan 点的正文只在 .tap 里，"代码在、跑起来没有"搜产物永远搜不到。
	// 工作区里 apply 之后，**包内的 .tap 才是最新真相**，所以 all 优先查包。
	Scope string
	// PkgPath 工作区包的路径（只有工作区命令会给）；给了才能查 pkg 内 .tap。
	PkgPath string
}

// GrepHit 是一条命中。
type GrepHit struct {
	Line      int    `json:"line"`
	Text      string `json:"text"`
	Point     string `json:"point,omitempty"` // 空 = 不在任何加点内
	Kind      string `json:"kind,omitempty"`  // 客制 / 标准 / 模板 / 生成器自加 / 未落地
	Inside    bool   `json:"inside"`
	Section   string `json:"section,omitempty"`
	SliceCode string `json:"slice,omitempty"`
}

// GrepResult 是 grep 的结果。
type GrepResult struct {
	Prog    string    `json:"prog"`
	Env     string    `json:"env"`
	Pattern string    `json:"pattern"`
	Total   int       `json:"total"` // 命中总数（含被 limit 截掉的）
	Shown   int       `json:"shown"`
	Hits    []GrepHit `json:"hits"`
	// Scope 本次查了哪些范围；TapHits 是 .tap 侧的命中（行号是 .tap 行号，两个文件的行号不可比）。
	Scope    string   `json:"scope,omitempty"`
	TapTotal int      `json:"tapTotal,omitempty"`
	TapHits  []TapHit `json:"tapHits,omitempty"`
	// TapSource .tap 命中来自哪：pkg（工作区包）/ materials（拉下来的材料）。
	TapSource string `json:"tapSource,omitempty"`
	// Note 解读提示（例如"命中的声明没有槽位，重产也不会保留"）。
	Note      string `json:"note,omitempty"`
	Link      string `json:"link"`
	StaleNote string `json:"staleNote,omitempty"`
}

// Grep 在产物里检索。这是"第一个动作永远是搜一下"的那条命令：
// 它比 grep 多条东西 —— 命中落在哪个加点、是客制还是模板、重产会不会保留。
func Grep(req GrepRequest) (*GrepResult, error) {
	doc, lines, err := loadForQuery(req.LinkRequest)
	if err != nil {
		return nil, err
	}
	needle := req.Pattern
	if needle == "" {
		return nil, fmt.Errorf("%w: 检索串不能为空", ErrInvalidOpt)
	}
	if !req.Case {
		needle = strings.ToLower(needle)
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	res := &GrepResult{Prog: req.Prog, Env: req.Env, Pattern: req.Pattern, Link: doc.Path,
		// JSON 数组空值给 [] 不给 null（agent 侧按数组消费）。
		Hits: []GrepHit{}}
	scope := strings.ToLower(strings.TrimSpace(req.Scope))
	if scope == "" {
		scope = "4gl"
	}
	_, res.StaleNote = staleNote(doc)

	for i, ln := range lines {
		if scope == "tap" {
			break // 只看 .tap 声明：产物侧不用扫
		}
		hay := ln
		if !req.Case {
			hay = strings.ToLower(ln)
		}
		if !strings.Contains(hay, needle) {
			continue
		}
		res.Total++
		if len(res.Hits) >= limit {
			continue
		}
		line := i + 1
		hit := GrepHit{Line: line, Text: strings.TrimRight(ln, "\r")}
		if p := innermostPointAt(doc.Points, line); p != nil {
			hit.Point = p.Name
			hit.Kind = p.Kind()
			hit.Inside = true
		}
		if s := sectionAt(doc.Sections, line); s != nil {
			hit.Section = s.ID
		}
		if sl := sliceBefore(doc.Slices.Markers, line); sl != "" {
			hit.SliceCode = sl
		}
		res.Hits = append(res.Hits, hit)
	}
	res.Shown = len(res.Hits)
	res.Scope = scope
	if scope == "tap" || scope == "all" || scope == "pkg" {
		tr, terr := grepTap(req, scope)
		if terr != nil {
			return nil, terr
		}
		res.TapTotal, res.TapHits = tr.Total, tr.Hits
		res.TapSource = tr.Env // 复用 Env 字段传来源：TapSearchResult 只在这里内部用
	}
	if res.TapTotal > 0 {
		noSlot := 0
		for _, h := range res.TapHits {
			if !h.HasSlot {
				noSlot++
			}
		}
		if res.Total == 0 {
			res.Note = fmt.Sprintf("产物里 0 条、.tap 里 %d 条：这段代码只在设计源里", res.TapTotal)
		}
		if noSlot > 0 {
			// 这一句是给"没跑 F7"那条告警打补丁：没有槽位的声明，跑多少次 F7 都不会出现。
			res.Note = joinNonEmpty(res.Note, fmt.Sprintf(
				"其中 %d 条落在**没有槽位**的声明里：重产也不会保留，不是 freshness 问题（跑 F7 也不会出现）", noSlot))
		}
	}
	return res, nil
}

func innermostPointAt(ps []LinkPoint, line int) *LinkPoint {
	var best *LinkPoint
	for i := range ps {
		p := &ps[i]
		if p.Line0 == 0 || line < p.Line0 || line > p.Line1 {
			continue
		}
		if best == nil || (p.Line1-p.Line0) < (best.Line1-best.Line0) {
			best = p
		}
	}
	return best
}

func sectionAt(ss []LinkSection, line int) *LinkSection {
	for i := range ss {
		s := &ss[i]
		if s.Line0 > 0 && line >= s.Line0 && line <= s.Line1 {
			return s
		}
	}
	return nil
}

func sliceBefore(ms []SliceMark, line int) string {
	code := ""
	for _, m := range ms {
		if m.Line > line {
			break
		}
		code = m.Code
	}
	return code
}

// orphanShown 把一个孤儿声明做成可显示的点位（正文取 .tap）。
func orphanShown(req LinkRequest, doc *LinkDoc, q string) (*ShownPoint, string) {
	q = strings.TrimSpace(q)
	if q == "" || len(doc.Orphans) == 0 {
		return nil, ""
	}
	lq := strings.ToLower(q)
	var exact, suffix, sub *LinkOrphan
	for i := range doc.Orphans {
		o := &doc.Orphans[i]
		n := strings.ToLower(o.Name)
		switch {
		case n == lq:
			exact = o
		case strings.HasSuffix(n, "."+lq):
			if suffix == nil {
				suffix = o
			}
		case strings.Contains(n, lq):
			if sub == nil {
				sub = o
			}
		}
	}
	o := exact
	if o == nil {
		o = suffix
	}
	if o == nil {
		o = sub
	}
	level := "exact"
	switch {
	case o == nil:
		return nil, ""
	case o == exact:
		level = "exact"
	case o == suffix:
		level = "suffix"
	default:
		level = "substring"
	}
	sp := &ShownPoint{
		LinkPoint: LinkPoint{Name: o.Name, Src: o.Src, Ver: o.Ver, DeclLines: o.Lines, Preview: o.Preview},
		Orphan:    true, OrphanReason: o.Reason,
	}
	if v, err := TapPoint(req, o.Name); err == nil && v != nil {
		sp.Content = v.Content
		sp.Line0, sp.Line1 = v.Line0, v.Line1
	}
	return sp, level
}

// grepTap 决定查哪份 .tap：工作区给了 PkgPath 且包存在时优先查包内（apply 之后
// 那才是最新真相；材料里的 .tap 还停在 fetch 那一刻）。
func grepTap(req GrepRequest, scope string) (*TapSearchResult, error) {
	if req.PkgPath != "" && (scope == "pkg" || scope == "all") {
		if b, err := pkgTapBytes(req.PkgPath); err == nil {
			tr, terr := TapSearchBytes(req.Prog, "pkg", b, req.Pattern, req.Limit, req.Case)
			if terr != nil {
				return nil, terr
			}
			return tr, nil
		} else if scope == "pkg" {
			return nil, err
		}
	}
	tr, err := TapSearch(req.LinkRequest, req.Pattern, req.Limit, req.Case)
	if err != nil {
		return nil, err
	}
	tr.Env = "materials"
	return tr, nil
}

// pkgTapBytes 取包内 .tap 的字节。
func pkgTapBytes(path string) ([]byte, error) {
	pkg, err := pkgfile.Open(path, pkgfile.OpenOptions{AcceptVer: "1.0"})
	if err != nil {
		return nil, err
	}
	tap := pkg.Tap()
	if tap == nil {
		return nil, fmt.Errorf("包里没有 .tap")
	}
	return tap.Data, nil
}
