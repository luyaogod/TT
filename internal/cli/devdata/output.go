package devdata

// 输出：JSON 默认（agent 用），--human 出小表（人用）。
//
// 两种形态**内容一致**：表格只换展示形态，不偷偷少字段（tt-pj-add-cli 的"可信"原则）。
// 因此每张表的列都从同一份结果对象里取，没有第二套计算。

import (
	"fmt"
	"os"
	"strings"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/devdata"
	"github.com/luyaogod/TT/internal/output"
)

// meta 本组共用的信封头（环境/来源）。企业只有在真的用了才写。
func meta(env string, ent int) output.Meta {
	m := common.CurrentMeta()
	m.Env = env
	m.Source = "local"
	if ent > 0 {
		m.Ent = ent
	}
	return m
}

// emit 统一出口。
func emit(data any, env string, ent int, columns []string, rows [][]string) error {
	return output.Emit(os.Stdout, output.Options{
		Format:  common.OutputFormat(),
		Meta:    meta(env, ent),
		Data:    data,
		Columns: columns,
		Rows:    rows,
	})
}

// emitStatus 输出环境状态（共用资料 + 程序清单 + 在跑任务）。
func emitStatus(st *devdata.EnvStatus) error {
	rows := [][]string{
		{"共用资料", "规格表", fmtCount(st.Dict.Tables, st.Dict.Skipped), fmtBytes(st.Dict.Size), fmtAge(st.Dict.SyncedAt), presentFlag(st.Dict.Present)},
		{"共用资料", "模板", fmt.Sprintf("%d / %d 个文件", st.Mdl.TemplateCount, st.Mdl.Files), fmtBytes(st.Mdl.Bytes), fmtAge(st.Mdl.SyncedAt), presentFlag(st.Mdl.Present)},
	}
	for _, p := range st.Progs {
		status := "最新"
		if p.Stale {
			status = "已过期"
			if len(p.StaleReasons) > 0 {
				status += "：" + p.StaleReasons[0].Detail
			}
		}
		rows = append(rows, []string{"按需资料", p.Prog, fmt.Sprintf("%d 个文件", p.Files), fmtBytes(p.Bytes), fmtAge(p.FetchedAt), status})
	}
	for _, j := range st.Jobs {
		rows = append(rows, []string{"任务", j.Describe(), fmt.Sprintf("%d/%d", j.Done, j.Total), "", j.Elapsed, j.Message})
	}
	// JSON 载荷与表格同源：直接给领域层的结构，避免"界面看得到、JSON 没有"
	data := map[string]any{
		"env":      st.Env,
		"dir":      st.Dir,
		"present":  st.Present,
		"dict":     st.Dict,
		"mdl":      st.Mdl,
		"progs":    st.Progs,
		"jobs":     st.Jobs,
		"warnings": st.Warnings,
	}
	return emit(data, st.Env, 0,
		[]string{"分组", "名称", "规模", "大小", "时间", "状态"}, rows)
}

// emitProgs 只输出程序清单（tt gen ls）。
func emitProgs(st *devdata.EnvStatus) error {
	rows := make([][]string, 0, len(st.Progs))
	for _, p := range st.Progs {
		status := "最新"
		if p.Stale {
			status = "已过期"
		}
		missing := ""
		if len(p.Missing) > 0 {
			missing = strings.Join(p.Missing, ",")
		}
		rows = append(rows, []string{p.Prog, p.Module, fmt.Sprintf("%d", p.Files), fmtBytes(p.Bytes),
			fmtAge(p.FetchedAt), status, missing})
	}
	return emit(map[string]any{"env": st.Env, "dir": st.Dir, "progs": st.Progs}, st.Env, 0,
		[]string{"程序", "模块", "文件数", "大小", "拉取时间", "状态", "缺失"}, rows)
}

// emitCheck 输出一个程序的新鲜度明细（tt gen check）。
func emitCheck(p *devdata.ProgStatus) error {
	rows := make([][]string, 0, len(p.StaleReasons)+1)
	if len(p.StaleReasons) == 0 {
		rows = append(rows, []string{p.Prog, "-", "材料是最新的"})
	}
	for _, r := range p.StaleReasons {
		rows = append(rows, []string{p.Prog, r.Code, r.Detail})
	}
	return emit(map[string]any{"prog": p}, p.Prog, 0,
		[]string{"程序", "原因码", "说明"}, rows)
}

// emitMirror 输出共用资料同步结果。
func emitMirror(res *devdata.MirrorResult) error {
	rows := [][]string{
		{"规格表", fmt.Sprintf("%d 张 / %d 行", res.TableCount, res.RowTotal), ""},
		{"模板", fmt.Sprintf("%d 个文件", res.MdlFiles), fmtBytes(res.MdlBytes)},
		{"耗时", res.Elapsed, ""},
	}
	if res.SkippedPull {
		rows = append([][]string{{"结果", "已同步过，跳过（--force 重拉）", ""}}, rows...)
	}
	for _, w := range res.Warnings {
		rows = append(rows, []string{"提示", w, ""})
	}
	for _, m := range res.Missing {
		rows = append(rows, []string{"缺少", m, ""})
	}
	return emit(res, res.Env, 0, []string{"项目", "值", "大小"}, rows)
}

// emitFetch 输出按程序拉取的结果。
func emitFetch(res *devdata.FetchResult) error {
	rows := [][]string{{"程序", res.Prog, res.Module}, {"大小", fmtBytes(res.Bytes), fmt.Sprintf("%d 个文件", len(res.Files))}}
	if label := moduleOriginLabel(res.ModuleOrigin); label != "" {
		rows = append(rows, []string{"模块来源", res.Module, label})
	}
	for _, f := range res.Files {
		rows = append(rows, []string{"文件", f.Rel, fmt.Sprintf("%d B", f.Size)})
	}
	for _, m := range res.Missing {
		rows = append(rows, []string{"缺少", m, ""})
	}
	for _, w := range res.Warnings {
		rows = append(rows, []string{"提示", w, ""})
	}
	return emit(res, res.Env, 0, []string{"项目", "值", "备注"}, rows)
}

// emitMap 输出「蓝图 ↔ 代码」索引的摘要（索引本体在 link.json；--full 才整份吐）。
func emitMap(doc *devdata.LinkDoc, full bool) error {
	st := doc.Stats
	mode := fmt.Sprintf("版型 %s", doc.Mode.Type)
	if doc.Mode.Type == "" {
		mode = "版型未知（无 .tab）"
	}
	if doc.Mode.Identity != "" {
		mode += fmt.Sprintf("，身份 %s", doc.Mode.Identity)
	}
	if doc.Mode.CusVer == "0000" {
		mode += "，从未客制"
	}
	if doc.Mode.UnfoldedSection {
		mode += "，未解開Section"
	}
	if doc.Mode.Freestyle {
		mode += "，freestyle"
	}

	pointNote := fmt.Sprintf("客制 %d（%d 行）、注释 %d、缺 %d、多 %d",
		st.PointsCustom, st.CustomLines, st.PointsCommented, st.PointsMissing, st.PointsExtra)
	if st.SlotsTap == 0 {
		pointNote = "无 .tap：点位全部来自 .4gl（客制/标准未知）"
	}
	fieldNote := "按 .tab 的 page_id + 字段归组"
	if st.FieldsLinked == 0 {
		fieldNote = "该版型没有页面字段（报表/批次）"
	}
	rows := [][]string{
		{"程序", doc.Prog, mode},
		{"规模", fmt.Sprintf("%d 行 / %d sections", st.G4Lines, st.SectionsG4),
			fmt.Sprintf("槽位 .tap %d / .4gl %d", st.SlotsTap, st.SlotsG4)},
		{"点位", fmt.Sprintf("%d 个", st.Points), pointNote},
		{"声明", fmt.Sprintf("%d 个", st.DeclTotal),
			fmt.Sprintf("客制 %d（%d 行）、无标记的函数 %d、孤儿 %d（客制 %d）",
				st.DeclCustom, st.DeclCustomLines, st.DeclUnmarked, st.DeclOrphans, st.DeclOrphansCustom)},
		{"字段", fmt.Sprintf("%d 个有逻辑 / %d 个栏位", st.FieldsLinked, st.ColumnsTotal), fieldNote},
		{"页面", fmt.Sprintf("%d 个", len(doc.Pages)), "序号与 page_id 不同；--pages 可列"},
		{"切片", fmt.Sprintf("%d 处 / %d 种", st.SlicesTotal, st.SlicesDistinct), ""},
		{"索引", doc.Path, "整份索引；脚本用 --full 或直接读它"},
	}
	for _, c := range doc.Checks {
		rows = append(rows, []string{"检查", c.ID + " [" + c.Level + "]", c.Detail})
	}
	for _, d := range doc.Slices.Drift {
		rows = append(rows, []string{"脱版", d.Code, fmt.Sprintf("产物 v%d / 模板 v%d", d.Used, d.Current)})
	}
	// 客制点前 20 条（人读时最想看的就是这个）
	shown := 0
	for _, p := range doc.Points {
		if !p.Custom || shown >= 20 {
			continue
		}
		shown++
		rows = append(rows, []string{"客制点", p.Name,
			fmt.Sprintf("%s:%d-%d  %s", doc.Prog+".4gl", p.Line0, p.Line1, p.Preview)})
	}
	if st.PointsCustom > shown {
		rows = append(rows, []string{"客制点", fmt.Sprintf("…另 %d 个", st.PointsCustom-shown),
			"清单：--custom（或 link.json 的 points[custom=true]）"})
	}
	if st.DeclOrphans > 0 {
		rows = append(rows, []string{"孤儿", fmt.Sprintf("%d 个（客制 %d）", st.DeclOrphans, st.DeclOrphansCustom),
			"名单：--orphans"})
	}

	// JSON 默认只给摘要：整份 doc 有几百 KB，直接吐到 stdout 对调用方是灾难。
	data := any(map[string]any{
		"schema": doc.Schema, "prog": doc.Prog, "env": doc.Env,
		"generatedAt": doc.GeneratedAt, "path": doc.Path,
		"mode": doc.Mode, "stats": st, "checks": doc.Checks,
		"drift": doc.Slices.Drift, "pages": doc.Pages,
		"notes": append(doc.Notes, "整份索引在 path；--full 输出到 stdout"),
	})
	if full {
		data = doc
	}
	return emit(data, doc.Env, 0, []string{"项目", "值", "备注"}, rows)
}

// emitOrphans 列出"写了不生效"的声明（--orphans）。
func emitOrphans(doc *devdata.LinkDoc) error {
	rows := make([][]string, 0, len(doc.Orphans))
	for _, o := range doc.Orphans {
		rows = append(rows, []string{o.Name, o.Src, fmt.Sprintf("%d", o.Lines), o.Reason, o.Preview})
	}
	return emit(map[string]any{
		"prog": doc.Prog, "count": len(doc.Orphans),
		"custom": doc.Stats.DeclOrphansCustom, "orphans": doc.Orphans, "link": doc.Path,
	}, doc.Env, 0, []string{"点名", "来源", "声明行", "原因", "内容首行"}, rows)
}

// emitPages 列页面（--pages）：解决"客户说的第 12 个页签是哪个 page_id"。
func emitPages(prog string, doc *devdata.LinkDoc) error {
	rows := make([][]string, 0, len(doc.Pages))
	for _, p := range doc.Pages {
		rows = append(rows, []string{p.Page, p.PageID, p.Record,
			fmt.Sprintf("%d", p.Columns), fmt.Sprintf("%d", p.Fields)})
	}
	return emit(map[string]any{
		"prog": prog, "pages": doc.Pages, "link": doc.Path,
		"note": "点名用的是 page_id（第三列起）：input.a.page<page_id>.<字段>",
	}, doc.Env, 0, []string{"页签序号", "page_id", "record", "栏位数", "有逻辑的字段"}, rows)
}

// emitGrep 输出检索命中（tt gen grep）。
func emitGrep(res *devdata.GrepResult) error {
	rows := make([][]string, 0, len(res.Hits))
	kind := map[bool]string{true: "加点内", false: "正文"}
	for _, h := range res.Hits {
		src := "—"
		if h.Point != "" {
			src = h.Kind
		}
		rows = append(rows, []string{
			fmt.Sprintf("%s:%d", res.Prog+".4gl", h.Line),
			src, kind[h.Inside],
			fmt.Sprintf("%s | %s", h.Section, h.SliceCode),
			h.Text,
		})
	}
	data := res
	if res.StaleNote != "" {
		rows = append(rows, []string{"提示", res.StaleNote, "", "", ""})
	}
	return emit(data, res.Env, 0, []string{"位置", "来源", "范围", "section|切片", "代码"}, rows)
}

// emitMapPoints 只列点位（--field / --custom）：给"我要改这个字段"这类问题用。
func emitMapPoints(doc *devdata.LinkDoc, field string, customOnly bool) error {
	var rows [][]string
	// 与 gen show 用同一套匹配（精确 → 后缀 → 子串），避免同一个词在两个命令里结论相反。
	pts, how := doc.Points, ""
	if field != "" {
		pts, how = devdata.MatchPoints(doc, field)
		if len(pts) == 0 {
			rows = append(rows, []string{"提示",
				fmt.Sprintf("没有匹配 %q 的点（用 gen fields 看字段名，或 gen grep 按内容找）", field), "", "", ""})
		} else if how != "" && how != "exact" {
			rows = append(rows, []string{"提示",
				fmt.Sprintf("按%s匹配 %d 个点（不是精确点名）", map[string]string{"suffix": "字段名后缀", "substring": "子串"}[how], len(pts)), "", "", ""})
		}
	}
	n := 0
	for _, p := range pts {
		if customOnly && !p.Custom {
			continue
		}
		n++
		where := fmt.Sprintf("%s:%d-%d", doc.Prog+".4gl", p.Line0, p.Line1)
		if p.Line0 == 0 {
			where = "（4gl 里没有）"
		}
		kind := "模板"
		switch {
		case p.Custom:
			kind = "客制"
		case p.Src == "s":
			kind = "标准"
		case p.Unmarked:
			kind = "客制函数"
		case p.Extra:
			kind = "生成器自加"
		case p.Missing:
			kind = "未落地"
		}
		rows = append(rows, []string{p.Name, kind, where,
			fmt.Sprintf("%d 行", p.DeclLines), p.Preview})
	}
	return emit(map[string]any{
		"prog": doc.Prog, "field": field, "customOnly": customOnly,
		"count": n, "points": rows, "link": doc.Path,
	}, doc.Env, 0, []string{"点位", "来源", "行区间", "声明行数", "内容首行"}, rows)
}

// emitShow 输出点位内容（tt gen show）。
func emitShow(res *devdata.ShowResult) error {
	rows := [][]string{}
	if res.Note != "" {
		rows = append(rows, []string{"提示", res.Note, "", ""})
	}
	for _, p := range res.Points {
		decl := "—"
		if p.DeclLines > 0 {
			decl = fmt.Sprintf("%d 行非空", p.DeclLines)
		}
		rows = append(rows, []string{
			p.Name, p.Kind(),
			fmt.Sprintf("%d-%d", p.Line0, p.Line1),
			fmt.Sprintf("section=%s  slot=%s  slice=%s  声明=%s",
				p.Section, p.Slot, strings.Join(p.SliceCodes, ","), decl),
		})
		for i, ln := range p.Content {
			rows = append(rows, []string{"", fmt.Sprintf("L%d", p.Line0+i), "", ln})
		}
		if p.Truncated {
			rows = append(rows, []string{"", "…", "", "（内容超过 --max，已截断）"})
		}
	}
	return emit(res, res.Env, 0, []string{"点名", "来源", "行号", "代码 / 属性"}, rows)
}

// emitAt 输出行号归属（tt gen at）。
func emitAt(res *devdata.AtResult) error {
	rows := [][]string{
		{"行号", fmt.Sprintf("%d", res.Line), fmt.Sprintf("文件共 %d 行", res.TotalLine)},
		{"结论", res.Verdict, ""},
	}
	if res.Stale && res.StaleNote != "" {
		rows = append(rows, []string{"注意", res.StaleNote, ""})
	}
	if res.Point != nil {
		rows = append(rows, []string{"点位", res.Point.Name,
			fmt.Sprintf("%s，第 %d-%d 行", res.Point.Kind(), res.Point.Line0, res.Point.Line1)})
	}
	if res.Section != nil {
		tapNote := ""
		if res.Section.TapLine > 0 {
			tapNote = fmt.Sprintf("（.tap 第 %d 行）", res.Section.TapLine)
		}
		rows = append(rows, []string{"section", res.Section.ID,
			fmt.Sprintf("第 %d-%d 行%s", res.Section.Line0, res.Section.Line1, tapNote)})
	}
	if res.Slice != nil {
		rel := "该行之前最近的标记（不保证覆盖此行）"
		if res.SliceRelation == "in-point" {
			rel = "在这个点内"
		}
		rows = append(rows, []string{"切片", res.Slice.Code,
			fmt.Sprintf("Version %d，标记在第 %d 行（%s）", res.Slice.Ver, res.Slice.Line, rel)})
	}
	for i, ln := range res.Content {
		no := res.Line - len(res.Content)/2 + i
		rows = append(rows, []string{fmt.Sprintf("L%d", no), ln, ""})
	}
	return emit(res, res.Env, 0, []string{"项目", "值", "备注"}, rows)
}

// emitFields 输出字段清单（tt gen fields）。
func emitFields(prog string, fs []devdata.LinkField, doc *devdata.LinkDoc) error {
	rows := make([][]string, 0, len(fs))
	for _, f := range fs {
		rows = append(rows, []string{
			fmt.Sprintf("%s/%s", f.Page, f.PageID), f.Record, f.Column,
			fmt.Sprintf("%d", len(f.Points)), strings.Join(f.Points, ", "),
		})
	}
	return emit(map[string]any{
		"prog": prog, "count": len(fs), "fields": fs, "link": doc.Path,
	}, doc.Env, 0, []string{"页签/page_id", "record", "字段", "点位数", "点位"}, rows)
}

// emitPack 输出打包结果（tt gen pack）。
func emitPack(res *devdata.PackResult) error {
	rows := [][]string{
		{"包", res.Path, fmtBytes(res.Bytes)},
		{"程序", res.Prog, res.Identity},
	}
	for _, e := range res.Entries {
		rows = append(rows, []string{"成员", e.Name, fmt.Sprintf("%d B  %s", e.Size, e.SHA256[:12])})
	}
	for _, w := range res.Warnings {
		rows = append(rows, []string{"提示", w, ""})
	}
	return emit(res, res.Env, 0, []string{"项目", "值", "备注"}, rows)
}

// moduleOriginLabel 材料取自标准还是客制模块（给人看的一列）。
func moduleOriginLabel(origin string) string {
	switch origin {
	case "custom":
		return "客制模块"
	case "standard":
		return "标准模块"
	case "explicit":
		return "--module 指定"
	}
	return ""
}

// emitCleanPlan 输出清理计划（dryRun=true 表示"只看不删"）。
func emitCleanPlan(plan *devdata.CleanPlan, dryRun bool) error {
	rows := make([][]string, 0, len(plan.Items))
	for _, it := range plan.Items {
		target := it.Env
		if it.Prog != "" {
			target += "/" + it.Prog
		}
		rows = append(rows, []string{it.Kind, target, it.Path, fmt.Sprintf("%d", it.Files), fmtBytes(it.Bytes)})
	}
	data := map[string]any{"dryRun": dryRun, "items": plan.Items, "files": plan.Files, "bytes": plan.Bytes}
	return emit(data, "", 0, []string{"类型", "对象", "路径", "文件数", "大小"}, rows)
}

// ---------- 小工具 ----------

func fmtBytes(n int64) string {
	if n <= 0 {
		return "0 B"
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}

func fmtCount(tables, skipped int) string {
	if skipped > 0 {
		return fmt.Sprintf("%d 张（跳过 %d）", tables, skipped)
	}
	return fmt.Sprintf("%d 张", tables)
}

func fmtAge(stamp string) string {
	if stamp == "" {
		return "—"
	}
	return devdata.AgeText(stamp)
}

func presentFlag(ok bool) string {
	if ok {
		return "已同步"
	}
	return "未同步"
}
