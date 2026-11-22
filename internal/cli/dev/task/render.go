package task

// 人读输出。**只换展示形态，不换口径**：每个数字都直接来自 JSON 里的同一份结果对象。
//
// 为什么默认不是文本：tt 的缺省是 JSON（见 internal/output 的说明，agent 消费更稳）。
// 想读人话就加 --human。这里每个函数都只被 --human 分支调用。

import (
	"fmt"
	"sort"
	"strings"

	"github.com/luyaogod/TT/internal/dev/taskws"
	"github.com/luyaogod/TT/internal/devdata"
)

// printOpen 是 open 完成后的摘要。
func printOpen(res *taskws.CreateResult) {
	t := res.Task
	var bytes int64
	for _, f := range res.Fetch.Files {
		bytes += f.Size
	}
	name := t.Prog
	if t.Task != "" {
		name += " · " + t.Task
	}
	fmt.Printf("已建工作区：%s\n", res.Dir)
	fmt.Printf("  %s（%s）· 模块 %s · 身份 %s\n", name, t.Env, t.Module, t.Identity)
	fmt.Printf("  材料 %d 个文件 / %s%s\n", len(res.Fetch.Files), fmtBytes(bytes), missingText(res.Fetch.Missing))
	if res.Index != nil {
		fmt.Printf("  索引 %s 点 · 客制 %s 点 · %d 个 section\n",
			fmtCount(res.Index.Points), fmtCount(res.Index.Custom), res.Index.Sections)
	}
	fmt.Printf("  包   %s（%s）\n", res.Pack.Path, fmtBytes(res.Pack.Bytes))
	fmt.Printf("  视图 %s（%d 个区域，可编辑 %d）\n", res.Export.Dir, res.Export.Regions, res.Export.Stats.Editable)
	for _, w := range res.Warnings {
		fmt.Printf("  ! %s\n", w)
	}
	fmt.Printf("\n下一步：cd %s\n", res.Dir)
	fmt.Printf("        tt dev brief                       # 先看整体\n")
	fmt.Printf("        tt dev find --intent check         # 再定位落点\n")
	fmt.Printf("        <编辑 ws/prog.full.4gl> → tt dev verify → tt dev apply\n")
}

// printUpdate 是 update 的结果。
func printUpdate(res *taskws.UpdateResult) {
	if res.Checked {
		if res.Stale {
			fmt.Printf("材料已过期（%d 条原因）：\n", len(res.StaleReasons))
			for _, r := range res.StaleReasons {
				fmt.Printf("  - %-18s %s\n", r.Code, r.Detail)
			}
		} else {
			fmt.Println("材料与服务器一致：没有过期项")
		}
	}
	switch {
	case res.Fetch != nil:
		fmt.Printf("材料已重拉：%d 个文件\n", len(res.Fetch.Files))
	case res.IndexRebuilt:
		fmt.Println("材料未变")
	}
	if res.IndexRebuilt && res.Status != nil {
		fmt.Printf("索引已重建：%s 点 / 客制 %s 点\n",
			fmtCount(res.Status.Index.Points), fmtCount(res.Status.Index.Custom))
	}
	// "保留了包/视图"只在真的动过东西时才说 —— 否则每次 update 都念一遍，
	// 会让人以为刚刚发生过什么。
	did := res.Fetch != nil || res.IndexRebuilt
	switch {
	case res.Repacked && res.Reexported:
		fmt.Println("包与视图已按新材料重建（视图干净，没有丢失改动）")
	case did && res.Status != nil && res.Status.Pkg.Modified:
		fmt.Println("包里有已 apply 的成果：保留包与视图，只更新了材料与索引")
	case did && res.Status != nil && res.Status.Ws.Dirty:
		fmt.Println("视图有未落盘改动：保留视图，只更新了材料与索引（--force）")
	}
	for _, w := range res.Warnings {
		fmt.Printf("  ! %s\n", w)
	}
	for _, n := range res.Notes {
		fmt.Printf("  · %s\n", n)
	}
	if res.Status != nil {
		fmt.Println()
		printStatus(res.Status)
	}
}

// printStatus 是工作区状态。
func printStatus(st *taskws.Status) {
	name := st.Prog
	if st.Task != "" {
		name += " · " + st.Task
	}
	env := st.Env
	if st.Ent > 0 {
		env += fmt.Sprintf("@%d", st.Ent)
	}
	fmt.Printf("%s（%s）\n", name, env)
	fmt.Printf("  工作区 %s\n", st.Dir)
	fmt.Printf("  材料   %s  %s  %d 个文件 / %s%s\n",
		orDash(st.Materials.FetchedAt), orDash(st.Materials.Module), st.Materials.Files,
		fmtBytes(st.Materials.Bytes), missingText(st.Materials.Missing))
	if st.Materials.DZAA002 > 0 {
		fmt.Printf("         规格版次 dzaa002=%d\n", st.Materials.DZAA002)
	}
	idxState := "缺失"
	switch {
	case st.Index.Present && st.Index.Fresh:
		idxState = "最新"
	case st.Index.Present:
		idxState = "需重建"
	}
	fmt.Printf("  索引   %s 点 / 客制 %s 点  ·  %s\n",
		fmtCount(st.Index.Points), fmtCount(st.Index.Custom), idxState)
	wsState := "缺失"
	switch {
	case st.Ws.Present && st.Ws.Dirty:
		wsState = "有未落盘改动"
	case st.Ws.Present:
		wsState = "干净"
	}
	fmt.Printf("  视图   %s · %s · %d 区域（可编辑 %d）\n",
		wsState, orDash(st.Ws.SectionState), st.Ws.Regions, st.Ws.Editable)
	pkgState := "未 apply"
	switch {
	case st.Pkg.Present && st.Ws.PkgSha != "" && !st.Pkg.MatchesBase:
		// 包与基线不一致时，绝不能说"可交付" —— 上一轮事故里 status 骗人的地方就是这里。
		pkgState = "与视图基线不一致（不可交付）"
	case st.Pkg.Modified:
		pkgState = "已 apply（与材料不同，可交付）"
	}
	fmt.Printf("  包     %s  %s  ·  %s\n", filepathBase(st.Pkg.Path), fmtBytes(st.Pkg.Bytes), pkgState)
	if st.Mdl.Present {
		fmt.Printf("  模板   %d 个切片\n", st.Mdl.Templates)
	} else {
		fmt.Printf("  模板   未复制（无切片脱版检测）\n")
	}
	for _, n := range st.Notes {
		fmt.Printf("  提示   %s\n", n)
	}
}

// printBrief 是一页纸画像。
func printBrief(b *devdata.BriefResult) {
	m := b.Mode
	head := fmt.Sprintf("%s（%s）", b.Prog, b.Env)
	var traits []string
	for _, p := range []struct{ k, v string }{
		{"主模板", m.Type}, {"模块", m.Module}, {"身份", m.Identity},
		{"客制版次", m.CusVer}, {"行业", m.Industry},
	} {
		if p.v != "" {
			traits = append(traits, p.k+" "+p.v)
		}
	}
	if m.UnfoldedSection {
		traits = append(traits, "未开 Section")
	}
	if m.SectionFlag != "" {
		traits = append(traits, "section_flag="+m.SectionFlag)
	}
	fmt.Printf("%s · %s\n", head, strings.Join(traits, " · "))
	s := b.Stats
	fmt.Printf("代码 %s 行 / %d 段 / %s 槽位 · 客制 %s 点（%s 行）· 未标记函数 %d · 孤儿声明 %d\n",
		fmtCount(s.G4Lines), len(b.Sections), fmtCount(s.SlotsTap),
		fmtCount(s.PointsCustom), fmtCount(s.CustomLines), s.DeclUnmarked, s.DeclOrphans)
	fmt.Printf("蓝图 %d 页 · %s 栏位（有逻辑 %s）· 切片 %d 种 / %d 处\n",
		len(b.Pages), fmtCount(s.ColumnsTotal), fmtCount(s.FieldsLinked), s.SlicesDistinct, s.SlicesTotal)

	if len(b.Pages) > 0 {
		fmt.Println("\n页签")
		for _, p := range b.Pages {
			title := p.Title
			if title == "" {
				title = "?"
			}
			fmt.Printf("  %2s  %-10s page_id=%-4s %-12s %3d 栏位 / %3d 有逻辑",
				p.Page, title, p.PageID, orDash(p.Record), p.Columns, p.Fields)
			if p.PageName != "" {
				fmt.Printf("  (%s)", p.PageName)
			}
			fmt.Println()
		}
	}

	if len(b.CustomPoints) > 0 {
		fmt.Printf("\n客制逻辑（按行数，共 %d 个，列前 %d）\n", b.CustomTotal, len(b.CustomPoints))
		for _, p := range b.CustomPoints {
			fmt.Printf("  %-42s L%-6d %4d 行  %s\n", p.Name, p.Line0, p.DeclLines, clip(p.Preview, 60))
		}
	}

	if len(b.Orphans) > 0 {
		custom := 0
		for _, o := range b.Orphans {
			if o.Src != "c" {
				continue
			}
			custom++
			fmt.Printf("\n客制孤儿（只有 .tap 声明、骨架里没有槽位，重产不会保留）：%s（%d 行）%s\n",
				o.Name, o.Lines, orphanWhy(o.Reason))
			if o.Preview != "" {
				fmt.Printf("  %s\n", clip(o.Preview, 90))
			}
		}
		if rest := len(b.Orphans) - custom; rest > 0 {
			fmt.Printf("\n其余 %d 个孤儿是标准声明的残留（tt gen map --orphans 看全部）\n", rest)
		}
	}

	if warn := warnChecks(b.Checks); len(warn) > 0 {
		fmt.Println("\n风险")
		for _, c := range warn {
			fmt.Printf("  %-5s %-16s %s\n", c.Level, c.ID, c.Detail)
		}
	}
	if len(b.Slices.Drift) > 0 {
		fmt.Println("\n切片脱版（产物是旧版模板生成的）")
		for _, d := range b.Slices.Drift {
			fmt.Printf("  %-5s 产物 v%d / 当前 v%d\n", d.Code, d.Used, d.Current)
		}
	}

	fmt.Println("\n意图 → 落点（tt dev find --intent <key>）")
	for _, in := range b.Intents {
		fmt.Printf("  %-8s %s\n", in.Key, clip(in.Hint, 100))
	}
	for _, n := range b.Notes {
		fmt.Printf("  · %s\n", n)
	}
}

// printFind 是意图定位结果。
func printFind(f *devdata.FindResult) {
	title := "意图定位"
	if f.Intent != "" {
		title += "：" + f.Intent
	}
	if f.Field != "" {
		title += "  字段：" + f.Field
	}
	fmt.Println(title)
	for _, h := range f.Intents {
		var tpl []string
		for _, t := range h.Templates {
			tpl = append(tpl, fmt.Sprintf("%s×%d", t.Code, t.Count))
		}
		fmt.Printf("  %-8s %d 个点", h.Key, h.Points)
		if len(tpl) > 0 {
			fmt.Printf("  参考切片 %s", strings.Join(tpl, " "))
		}
		fmt.Println()
		fmt.Printf("           %s\n", h.Hint)
	}
	if len(f.Points) == 0 {
		fmt.Printf("\n没有命中。%s\n", f.Note)
		if f.Intent == "" && f.Field == "" {
			fmt.Println("\n可用意图：")
			for _, in := range devdata.Intents {
				fmt.Printf("  %-8s %s\n", in.Key, clip(in.Hint, 100))
			}
		}
		return
	}
	fmt.Printf("\n命中 %d 个（列 %d）：\n", f.Total, len(f.Points))
	for _, p := range f.Points {
		name := p.Name
		if lb := labelFor(f.Labels, p.Name); lb != "" {
			name += "（" + lb + "）"
		}
		fmt.Printf("  %-46s L%-6d %-8s %4d 行  %s\n",
			name, p.Line0, p.Kind(), p.DeclLines, clip(p.Preview, 46))
	}
	if len(f.Orphans) > 0 {
		fmt.Println("\n孤儿声明（有 .tap 正文、骨架里没有槽位：重产不保留）：")
		for _, o := range f.Orphans {
			fmt.Printf("  %-44s %4d 行  %s\n", o.Name, o.Lines, clip(o.Preview, 46))
		}
	}
	if f.Note != "" {
		fmt.Printf("\n%s\n", f.Note)
	}
}

// printAt 是行号反查。
func printAt(a *devdata.AtResult) {
	where := a.Verdict
	if a.Point != nil {
		where = fmt.Sprintf("%s（%s · %s）", a.Point.Name, a.Point.Kind(), pointSpan(a.Point))
	}
	fmt.Printf("L%d → %s\n", a.Line, where)
	fmt.Printf("  （按 .4gl 产物行号解读；.4gl 共 %d 行。视图/包内 .tap 的行号不通用）\n", a.TotalLine)
	if a.Point == nil && a.Section == nil {
		fmt.Println("  ! 这一行不在任何加点/section 内：如果你给的是视图或 .tap 行号，请改用 tt dev at <点名>")
	}
	var parts []string
	if a.Section != nil {
		parts = append(parts, "section "+a.Section.ID)
	}
	if a.Slice != nil {
		rel := ""
		if a.SliceRelation != "" {
			rel = "（" + a.SliceRelation + "）"
		}
		parts = append(parts, fmt.Sprintf("切片 %s v%d%s", a.Slice.Code, a.Slice.Ver, rel))
	}
	if len(parts) > 0 {
		fmt.Println("  " + strings.Join(parts, " · "))
	}
	if a.StaleNote != "" {
		fmt.Println("  ! " + a.StaleNote)
	}
	if len(a.Content) > 0 {
		fmt.Println(strings.Repeat("-", 60))
		start := a.ContentStart
		if start < 1 {
			start = a.Line
		}
		for i, ln := range a.Content {
			fmt.Printf("%d: %s\n", start+i, ln)
		}
	}
}

// printShow 是点名内容。
func printShow(s *devdata.ShowResult) {
	if len(s.Points) == 0 {
		fmt.Printf("没有匹配的点位：%s\n", s.Query)
		if s.Note != "" {
			fmt.Println(s.Note)
		}
		return
	}
	fmt.Printf("%s（匹配方式：%s）\n", s.Query, s.Match)
	if s.Note != "" {
		fmt.Println(s.Note)
	}
	for i, p := range s.Points {
		if i > 0 {
			fmt.Println()
		}
		head := p.Name
		if p.Label != "" {
			head += "（" + p.Label + "）"
		}
		kind := p.Kind()
		if p.Orphan {
			kind = "孤儿（无槽位，重产不保留）"
		}
		sec := sectionOrDash(p.Section)
		if p.Orphan {
			sec = "正文在 .tap（.4gl 里没有）"
		}
		fmt.Printf("%s  %s · %s · %s\n", head, kind, pointSpan(&p.LinkPoint), sec)
		if len(p.SliceCodes) > 0 {
			fmt.Printf("  切片 %s\n", strings.Join(p.SliceCodes, " "))
		}
		fmt.Println(strings.Repeat("-", 60))
		for j, ln := range p.Content {
			fmt.Printf("%d: %s\n", p.Line0+j, ln)
		}
		if p.Truncated {
			fmt.Printf("…（只打前 %d 行，--max 调大）\n", len(p.Content))
		}
	}
}

// printGrep 是检索结果（产物 + .tap 声明两段，行号各自独立）。
func printGrep(g *devdata.GrepResult) {
	if g.Scope == "tap" || g.Scope == "pkg" {
		// 只看 .tap 时别打"产物命中 0 条"：那个 0 是"本范围被排除"，不是结果。
		fmt.Printf("检索 %q：.tap 命中 %d 条（列 %d）\n", g.Pattern, g.TapTotal, len(g.TapHits))
	} else {
		fmt.Printf("检索 %q：产物命中 %d 条（列 %d）", g.Pattern, g.Total, g.Shown)
		fmt.Printf("；.tap 命中 %d 条（列 %d）", g.TapTotal, len(g.TapHits))
		fmt.Println()
	}
	if g.StaleNote != "" {
		fmt.Println("! " + g.StaleNote)
	}
	if g.Note != "" {
		fmt.Println("· " + g.Note)
	}
	for _, h := range g.Hits {
		who := h.Point
		if who == "" {
			who = "（不在加点内）"
		}
		extra := h.Kind
		if h.Section != "" {
			extra += " · " + h.Section
		}
		if h.SliceCode != "" {
			extra += " · " + h.SliceCode
		}
		fmt.Printf("%6d  %-42s %-22s %s\n", h.Line, clip(who, 42), extra, clip(strings.TrimSpace(h.Text), 70))
	}
	if g.TapTotal > 0 {
		where := "材料里的 .tap"
		if g.TapSource == "pkg" {
			where = "包内 .tap（工作区最新真相）"
		}
		fmt.Printf("\n.tap 声明 —— 来源：%s；行号是 .tap 行号，与产物不可比\n", where)
		for _, h := range g.TapHits {
			who := h.Point
			if who == "" {
				who = "（不在任何声明里）"
			}
			state := "有槽位"
			if !h.HasSlot {
				state = "**无槽位，重产不保留**"
			}
			src := h.Src
			if src == "" {
				src = "-"
			}
			fmt.Printf("%6d  %-42s %-6s %-18s %s\n", h.Line, clip(who, 42), src, state, clip(strings.TrimSpace(h.Text), 50))
		}
	}
	if g.Total > g.Shown || g.TapTotal > len(g.TapHits) {
		fmt.Printf("\n只列前 %d 条/范围：--limit 调大看全部\n", g.Shown)
	}
}

// ---------- 小工具 ----------

// labelFor 取某点对应的画面中文名（Labels 的键是列名）。
func labelFor(labels map[string]string, point string) string {
	for col, lb := range labels {
		if col != "" && strings.HasSuffix(point, "."+col) {
			return lb
		}
	}
	return ""
}

// pointSpan 是"L50710-50734 / 25 行"。
func pointSpan(p *devdata.LinkPoint) string {
	if p.Line0 == 0 {
		return "无行号"
	}
	if p.Line1 > p.Line0 {
		return fmt.Sprintf("L%d-%d", p.Line0, p.Line1)
	}
	return fmt.Sprintf("L%d", p.Line0)
}

// sectionOrDash 给 section 名（空则 ?）。
func sectionOrDash(s string) string {
	if s == "" {
		return "?"
	}
	return "section " + s
}

// warnChecks 只留 warn/error（ok 不占篇幅）。
func warnChecks(cs []devdata.LinkCheck) []devdata.LinkCheck {
	var out []devdata.LinkCheck
	for _, c := range cs {
		if c.Level == "warn" || c.Level == "error" {
			out = append(out, c)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Level != out[j].Level {
			return out[i].Level == "error"
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// clip 截断长文本（按 rune，避免切坏中文）。
func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// orDash 空值给 `-`。
func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

// missingText 是"缺 N 项"。
func missingText(missing []string) string {
	if len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf("（可选素材缺 %d 项，不影响开工：%s）", len(missing), strings.Join(missing, ", "))
}

// filepathBase 取文件名（不引 path/filepath，免得为一行再 import）。
func filepathBase(p string) string {
	if p == "" {
		return "-"
	}
	if i := strings.LastIndexAny(p, `\/`); i >= 0 {
		return p[i+1:]
	}
	return p
}

// fmtBytes 人类可读体积。
func fmtBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// fmtCount 千分位。
func fmtCount(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}

// orphanWhy 把孤儿的 reason 翻成一句人话。
func orphanWhy(reason string) string {
	switch reason {
	case "no-slot":
		return "（原因：.tap 里有声明，骨架里没有对应的洞）"
	case "no-code":
		return "（原因：函数名在 .4gl 里也找不到）"
	default:
		return ""
	}
}

// printPkgPoint 打"包内 .tap 里的这个点"（apply 之后复核用）。
func printPkgPoint(v *devdata.TapPointView, pkgPath string) {
	kind := "客制"
	if v.Src == "s" {
		kind = "标准"
	}
	fmt.Printf("%s  %s · 包内 .tap L%d-%d · %s\n", v.Name, kind, v.Line0, v.Line1, filepathBase(pkgPath))
	if !v.HasSlot {
		fmt.Println("! 骨架里没有这个点的槽位：重产不会保留它")
	}
	fmt.Println(strings.Repeat("-", 60))
	for i, ln := range v.Content {
		fmt.Printf("%d: %s\n", v.Line0+i, ln)
	}
}
