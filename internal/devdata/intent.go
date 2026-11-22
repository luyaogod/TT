package devdata

import (
	"fmt"
	"strings"
)

// 意图 → 落点：把"我要做什么"翻成"改哪个加点/参考哪个切片"。
//
// 为什么需要这张表：T100 的代码落在**命名槽位**里（`input.a.page73.xmdduc001`），
// 而需求是业务语言（"保存前加个检核"）。两者之间那一步（哪个段、哪个位置码、哪块
// 模板）是新手最容易猜错的地方 —— 猜错的后果不是报错，而是"写了但不生效"（改到
// 4gl 正文，F7 一重产就没了）或"改到错的槽位，永远不触发"。
//
// 表里的每一行都对应实测过的段名（axmt500_wf，57 个 section / 3,215 个槽位）。
// 位置码的语义（设计器术语）：
//
//	a = AFTER FIELD   离开字段之后（检核/带值最常见）
//	b = BEFORE FIELD  进入字段之前（预带默认值）
//	c = ON ACTION     按钮/动作触发（开窗、跳转）
//	g = ON CHANGE     字段值变化

// Intent 是一条"意图 → 落点"的知识条目。
type Intent struct {
	Key string `json:"key"`
	// Aliases 检索用的别名（中文/英文关键词，find 按子串匹配）。
	Aliases []string `json:"aliases,omitempty"`
	// Sections 落点的段名前缀（点名第一段）：input / construct / insert_b / ui_dialog …
	Sections []string `json:"sections,omitempty"`
	// Slots 更细的落点标记（点名里任意一段命中即可）：browser_fill / before_insert …
	Slots []string `json:"slots,omitempty"`
	// Templates 这些段的正文由哪些切片模板生成（改的时候要跟模板写法一致）。
	Templates []string `json:"templates,omitempty"`
	// Hint 一句话：该改哪里、别改哪里。
	Hint string `json:"hint"`
}

// Intents 是意图表（顺序即帮助里的展示顺序）。
//
// 只收"落点确定"的意图。像"加打印"这种要看报表设计的，不硬编一张表 —— 给错的
// 落点比不给更坏。
var Intents = []Intent{
	{
		Key:       "check",
		Aliases:   []string{"检核", "校验", "检查", "check", "validate", "栏位检核"},
		Sections:  []string{"input", "chk_data_exist", "fill_chk", "action_chk", "idx_chk"},
		Templates: []string{"a17", "a05"},
		Hint:      "栏位检核改 input.a（AFTER FIELD）；跨字段/整单检核落 chk_data_exist；别改 .4gl 正文（F7 会重产覆盖）。",
	},
	{
		Key:       "lookup",
		Aliases:   []string{"开窗", "查询窗", "带值", "browse", "lookup", "选单"},
		Sections:  []string{"construct", "browser_fill", "ui_browser_refresh"},
		Templates: []string{"a07", "a08"},
		Hint:      "开窗改 construct.c（ON ACTION）或 browser_fill；用 r.q 定义的开窗不要在这里重写 SQL。",
	},
	{
		Key:       "default",
		Aliases:   []string{"预带", "默认值", "带值默认", "default", "初始化值"},
		Sections:  []string{"input", "b_fill", "b_fill2", "init", "default_search"},
		Templates: []string{"a05", "a06"},
		Hint:      "进入字段时预带改 input.b（BEFORE FIELD）；整段/整页预带落 b_fill / b_fill2；同步预带落 init。",
	},
	{
		Key:       "save",
		Aliases:   []string{"保存前", "存盘前", "异动前", "before_save", "before_insert", "before_update"},
		Sections:  []string{"insert_b", "update_b", "delete_b", "insert", "modify", "delete"},
		Slots:     []string{"before_insert", "before_update", "before_delete", "key_insert_b", "key_update_b"},
		Templates: []string{"a09", "a10"},
		Hint:      "保存前改 insert_b / update_b（key_*_b 是按键触发的那一对）；删除前改 delete_b。",
	},
	{
		Key:       "display",
		Aliases:   []string{"显示前", "画面", "显示", "show", "display", "refresh"},
		Sections:  []string{"show", "detail_show", "ui_detailshow", "ui_headershow", "ui_dialog", "status_show"},
		Templates: []string{"a20", "a21"},
		Hint:      "画面显示/刷新改 show / detail_show / ui_dialog；别在 ui_dialog 里做数据异动（它会被重复调用）。",
	},
	{
		Key:       "init",
		Aliases:   []string{"初始化", "启动", "init", "construct", "进入作业"},
		Sections:  []string{"init", "global", "main", "construct", "description"},
		Templates: []string{"a01", "a02"},
		Hint:      "作业初始化改 init；全局变量/常量落 global 的 define 区；construct 是画面构造，最容易被重产。",
	},
	{
		Key:       "query",
		Aliases:   []string{"查询", "取数", "query", "search", "筛选", "filter"},
		Sections:  []string{"query", "construct", "filter", "filter_parser", "fetch", "default_search"},
		Slots:     []string{"sql", "define_sql", "after_define_sql"},
		Templates: []string{"a03", "a04"},
		Hint:      "查询条件/取数改 query 或 construct 的 sql 段（define_sql）；画面筛选条件改 filter。",
	},
	{
		Key:       "sign",
		Aliases:   []string{"签核", "复核", "审核", "sign", "approve", "reject"},
		Sections:  []string{"signature", "signing", "approved", "rejection", "withdraw"},
		Templates: []string{"a30"},
		Hint:      "签核流程改 signature / signing；状态迁移改 state_change。",
	},
}

// IntentByKey 按 key 或别名取一条意图（大小写不敏感，允许子串）。
func IntentByKey(q string) []Intent {
	q = lowerTrim(q)
	if q == "" {
		return nil
	}
	var out []Intent
	for _, in := range Intents {
		if lowerTrim(in.Key) == q {
			return []Intent{in}
		}
	}
	for _, in := range Intents {
		hit := false
		for _, a := range in.Aliases {
			if strings.Contains(lowerTrim(a), q) || strings.Contains(q, lowerTrim(a)) {
				hit = true
				break
			}
		}
		if hit {
			out = append(out, in)
		}
	}
	return out
}

// FindRequest 是一次意图定位请求。
type FindRequest struct {
	LinkRequest
	// Intent 意图 key 或关键词（"check"/"保存前"…）。空 = 只按字段/内容过滤。
	Intent string
	// Field 字段名（可只给后缀，如 xmdduc001）**或画面中文名**（如 刻号）。
	Field string
	// Page 只要这个页签的点：page_id（73）或页签序号（12），两者都接受。
	Page string
	// Limit 最多返回多少个点（默认 20）。
	Limit int
}

// IntentHit 是一个命中的意图，附带本程序里这些落点的实际数量。
type IntentHit struct {
	Intent
	// Points 命中该意图的槽位数（本程序实测）。
	Points int `json:"points"`
	// Templates 命中的切片代号（含本程序里的用量）；模板没被用到的不列。
	Templates []SliceUse `json:"templates,omitempty"`
}

// FindResult 是意图定位的结果。
type FindResult struct {
	Prog   string `json:"prog"`
	Env    string `json:"env"`
	Intent string `json:"intent,omitempty"`
	Field  string `json:"field,omitempty"`
	// FieldBy 这次 --field 是按什么命中的：column（列名）/ label（画面中文名）。
	FieldBy string `json:"fieldBy,omitempty"`
	// Labels 列名 → 中文名（命中的那些列），供输出时并排显示。
	Labels  map[string]string `json:"labels,omitempty"`
	Page    string            `json:"page,omitempty"`
	Intents []IntentHit       `json:"intents"`
	Points  []LinkPoint       `json:"points"`
	// Orphans 同字段/同页上的孤儿声明（有 .tap 正文、骨架里没有槽位）。
	// 单独一栏是刻意的：它们不是"点位"，混进 Points 会被当成可改的落点。
	Orphans []LinkOrphan `json:"orphans,omitempty"`
	// Total 命中总数（Points 可能被 Limit 截断）。
	Total int    `json:"total"`
	Note  string `json:"note,omitempty"`
	Link  string `json:"link"`
}

// Find 按意图 + 可选字段定位落点。
//
// --field 先当列名（精确/后缀/子串），再当**画面中文名**（.tsd 的标签）。
// 这一步是必须的：现场说的是"刻号"，代码里叫 xmdduc003，中间那层翻译不该由人来查。
func Find(req FindRequest) (*FindResult, error) {
	doc, err := LoadIndex(req.LinkRequest)
	if err != nil {
		return nil, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	res := &FindResult{Prog: req.Prog, Env: doc.Env, Intent: req.Intent, Field: req.Field,
		Page: req.Page, Link: doc.Path,
		// JSON 数组的空值必须是 [] 而不是 null：agent 侧按数组消费，
		// null 会直接炸（find/search/brief 同一口径）。
		Intents: []IntentHit{}, Points: []LinkPoint{}}

	intents := IntentByKey(req.Intent)
	if req.Intent != "" && len(intents) == 0 {
		res.Note = "没有这条意图；先 tt dev brief 看可用的意图清单（check/lookup/default/save/display/init/query/sign）"
	}
	sliceUse := map[string]SliceUse{}
	for _, u := range doc.Slices.ByCode {
		sliceUse[u.Code] = u
	}

	// 字段解析：列名优先，其次画面标签。
	var cols []string
	if req.Field != "" {
		cols, res.FieldBy, res.Labels = ResolveColumns(doc, req.Field)
		if len(cols) > 0 && res.FieldBy == "label" {
			res.Note = joinNonEmpty(res.Note,
				fmt.Sprintf("%q 是画面中文名，对应列 %s", req.Field, strings.Join(cols, ", ")))
		}
	}
	// 页签收窄：接受 page_id，也接受页签序号（会明确换算出 page_id）。
	pageID := ""
	if p := strings.TrimSpace(req.Page); p != "" {
		if docHasPageID(doc, p) {
			pageID = p
		} else if id, ok := pageIDOfSeq(doc, p); ok {
			pageID = id
			res.Note = joinNonEmpty(res.Note,
				fmt.Sprintf("页签序号 %s = page_id %s", p, id))
		} else {
			res.Note = joinNonEmpty(res.Note, fmt.Sprintf("没有页签 %q（--page 接受 page_id 或页签序号）", p))
		}
	}
	match := func(p LinkPoint) bool {
		if len(cols) > 0 {
			if !containsFold(cols, columnOfPoint(p.Name)) {
				return false
			}
		} else if f := lowerTrim(req.Field); f != "" && !strings.HasSuffix(lowerTrim(p.Name), "."+f) {
			return false
		}
		if pageID != "" {
			if m := rePointPage.FindStringSubmatch(p.Name); m == nil || m[1] != pageID {
				return false
			}
		}
		return true
	}

	// 候选过滤：每个意图各自算一遍，命中的点按名字去重。
	seen := map[string]bool{}
	for _, in := range intents {
		hit := IntentHit{Intent: in}
		for _, u := range in.Templates {
			if su, ok := sliceUse[u]; ok {
				hit.Templates = append(hit.Templates, su)
			}
		}
		for _, p := range doc.Points {
			if !intentMatches(in, p) || !match(p) {
				continue
			}
			hit.Points++
			if seen[p.Name] {
				continue
			}
			seen[p.Name] = true
			res.Points = append(res.Points, p)
		}
		if hit.Points > 0 || req.Field == "" {
			res.Intents = append(res.Intents, hit)
		}
	}
	if req.Intent == "" && req.Field != "" {
		// 只给字段：不套意图，直接按字段找（等价于"这个栏位的逻辑散在哪些段"）。
		for _, p := range doc.Points {
			if !match(p) || seen[p.Name] {
				continue
			}
			seen[p.Name] = true
			res.Points = append(res.Points, p)
		}
	}
	res.Total = len(res.Points)
	res.Orphans = collectOrphans(doc, req.Field, res.Points)
	if len(res.Orphans) > 0 {
		var names []string
		for i, o := range res.Orphans {
			if i >= 3 {
				names = append(names, fmt.Sprintf("…共 %d 个", len(res.Orphans)))
				break
			}
			names = append(names, o.Name)
		}
		res.Note = joinNonEmpty(res.Note, fmt.Sprintf(
			"另有 %d 个**孤儿声明**（只有 .tap 正文、骨架里没有槽位：重产不保留、运行期不执行）：%s；用 tt dev at <点名> 看正文",
			len(res.Orphans), strings.Join(names, ", ")))
	}
	if res.Total > limit {
		omitted := omittedPages(res.Points[limit:])
		res.Points = res.Points[:limit]
		res.Note = joinNonEmpty(res.Note,
			fmt.Sprintf("命中 %d 个，只列前 %d 个%s（--limit 调大；或加 --field/--page 收窄）",
				res.Total, limit, omitted))
	}
	if res.Total == 0 && req.Field != "" {
		// "这个字段有代码但不生效"是现场最常见的一种"没有命中"。别只说"换个字段名"。
		f := lowerTrim(req.Field)
		var orph []string
		for _, o := range doc.Orphans {
			n := lowerTrim(o.Name)
			if n == f || strings.HasSuffix(n, "."+f) {
				orph = append(orph, o.Name)
			}
		}
		if len(orph) > 0 {
			res.Note = joinNonEmpty(res.Note, fmt.Sprintf(
				"这个字段有 %d 个声明**没有槽位**（孤儿：重产不保留、运行期不执行）：%s；用 tt dev at <点名> 看正文",
				len(orph), strings.Join(orph, ", ")))
		}
	}
	if res.Total == 0 && res.Note == "" {
		res.Note = "没有命中：字段可以给列名或画面中文名；也可以不带 --intent 先看这个字段有哪些点"
	}
	return res, nil
}

// containsFold 判断列表里是否含某个值（忽略大小写）。
func containsFold(list []string, v string) bool {
	for _, s := range list {
		if strings.EqualFold(s, v) {
			return true
		}
	}
	return false
}

// intentMatches 判断一个点是否落在该意图的段/槽位上。
func intentMatches(in Intent, p LinkPoint) bool {
	if len(in.Sections) == 0 && len(in.Slots) == 0 {
		return false
	}
	segs := strings.Split(p.Name, ".")
	// 段名匹配看第一段（input.…）与前两段（insert_b.before_insert.…）。
	head := segs[0]
	for _, s := range in.Sections {
		if strings.EqualFold(head, s) {
			return true
		}
	}
	if len(in.Slots) > 0 {
		for _, seg := range segs[1:] {
			for _, s := range in.Slots {
				if strings.EqualFold(seg, s) {
					return true
				}
			}
		}
	}
	return false
}

// lowerTrim 是本地小工具（避免为两处调用在多个文件里重复写）。
func lowerTrim(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// joinNonEmpty 拼两段提示（空串跳过）。
func joinNonEmpty(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return a + "；" + b
	}
}

// collectOrphans 找出与本次查询相关的孤儿声明：同名/同后缀/同列，或落在命中的那些页上。
func collectOrphans(doc *LinkDoc, field string, pts []LinkPoint) []LinkOrphan {
	if len(doc.Orphans) == 0 {
		return nil
	}
	f := lowerTrim(field)
	pages := map[string]bool{}
	for _, p := range pts {
		if m := rePointPage.FindStringSubmatch(p.Name); m != nil {
			pages[m[1]] = true
		}
	}
	// 只有"命中收敛到一页"时才按页带出孤儿：像「项次」这种跨 8 页的标签，
	// 按页带孤儿会把十几条无关声明倒给用户（实测），比不给更糟。
	pageMates := len(pages) == 1
	var out []LinkOrphan
	seen := map[string]bool{}
	for _, o := range doc.Orphans {
		n := lowerTrim(o.Name)
		direct := f != "" && (n == f || strings.HasSuffix(n, "."+f) || strings.EqualFold(columnOfPoint(o.Name), f))
		onPage := false
		if f != "" {
			if m := rePointPage.FindStringSubmatch(o.Name); m != nil && pages[m[1]] {
				onPage = true
			}
		}
		if !direct && !(pageMates && onPage) {
			continue
		}
		if seen[o.Name] {
			continue
		}
		seen[o.Name] = true
		out = append(out, o)
		if len(out) >= 10 {
			break
		}
	}
	return out
}

// omittedPages 把被 limit 截掉的那些点涉及的页签列出来（不然"没显示 page73"没人知道）。
func omittedPages(rest []LinkPoint) string {
	seen := map[string]bool{}
	var ids []string
	for _, p := range rest {
		m := rePointPage.FindStringSubmatch(p.Name)
		if m == nil || seen[m[1]] {
			continue
		}
		seen[m[1]] = true
		ids = append(ids, "page_id "+m[1])
		if len(ids) >= 5 {
			break
		}
	}
	if len(ids) == 0 {
		return ""
	}
	return "（未显示的涉及 " + strings.Join(ids, ", ") + "…）"
}
