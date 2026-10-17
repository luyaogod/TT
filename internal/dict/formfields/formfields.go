package formfields

// M3 第一步（dict 侧复用字段表）：**只读**的"离线字段表"数据源。
//
// 它回答一个字典线一直答不出的问题：**这张作业/表单上有哪些字段，各自的中文名与类型是什么**。
//
// 三条硬约束（都在代码里，不靠注释）：
//
//  1. **不复制解析器**：`.42f` 一律交给 `fep.ParseForm42F`（契约 C1）。本包只做
//     "字段表 × 字典"的合并与路径定位 —— `tt form show` 的输出契约（C5.2）一个键都不动，
//     本包的 JSON 键是**另一套**（见 README）。
//  2. **不猜**（契约 A15 / A18）：中文名只有两种来源 —— 字典 `dzeb_t`/`dzebl_t` 给出的
//     字段说明，或"没有"。字典查不到、多个候选各说各话、查表报错 —— 一律
//     `LabelUnresolved` 且 `Label` 为空，**绝不**回落成 `colName` / `cmt_*` / `lbl_*`。
//  3. **不连真机**：数据源只读本地目录里的 `.42f`（根目录由参数或环境变量给），
//     字典侧只吃调用方传进来的 `FieldNamer`（本地 SQLite 镜像即可离线跑）。

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tt/internal/dict/db"
	"tt/internal/fep"
)

// ── 状态常量：中文名"有没有"必须写在脸上 ─────────────────────────────────────

// LabelState 是**中文名**这一列的解析状态。
//
// 只有两个取值：`resolved`（字典给出了名字）与 `unresolved`（没给出）。
// 没有第三种 —— "大概像这样" 这种第三态就是 A15 明令禁止的启发式判定。
type LabelState string

const (
	// LabelResolved：字典查到了名字，`Field.Label` 非空。
	LabelResolved LabelState = "resolved"
	// LabelUnresolved：字典没给名字，`Field.Label` 必为空串，原因见 `Field.LabelReason`。
	LabelUnresolved LabelState = "unresolved"
)

// 未解析的原因（`Field.LabelReason`）。分开是为了让"为什么没名字"不必靠猜：
// 是没字典、是表单里没表可依、是字典真没这个字段，还是多个候选在打架。
const (
	ReasonNoDict    = "no-dict"     // 本次没接字典数据源
	ReasonNoTable   = "no-table"    // 表单里没有可依的表（RecordView/sqlTabName 都指不到）
	ReasonNotInDict = "not-in-dict" // 字典里没有这个字段
	ReasonAmbiguous = "ambiguous"   // 多个候选给出不同的名字 → 不挑
)

// 中文名整列的数据来源（`Table.LabelSource`）。
const (
	LabelSourceDict = "dict" // 接了字典（本地镜像或远程直查）
	LabelSourceNone = "none" // 没接字典 → 全列 unresolved/no-dict
)

// ── 字典侧的窄接口（接口归消费者；*db.DB 与 live.Live 都满足） ───────────────

// FieldNamer 是本包对字典侧的全部要求：**一个**方法。
//
// `internal/dict/db.Source`（本地镜像 `*db.DB` 与远程 `live.Live` 共同实现的接口）
// 已包含 `QueryTable`，所以 CLI 直接把 `GetDB()` 传进来即可 —— 中文字段名的解析
// 复用字典既有实现（`dzeb_t` JOIN `dzebl_t ... zh_CN`），本包不另写一套查名逻辑。
type FieldNamer interface {
	QueryTable(tableName string) ([]db.TableInfo, error)
}

// ── 结果结构（JSON 键即本包的新契约，见 README） ─────────────────────────────

// Field 是表单上的一个字段条目（含它从字典借来的中文名）。
type Field struct {
	ColName     string     `json:"colName"` // 字段身份（如 b_xmda001），**不是** fieldId
	FieldID     int        `json:"fieldId"` // 表单内编号（编译期分配）
	Tag         string     `json:"tag"`     // FormField | TableColumn | PhantomColumn | PhantomFormField
	Widget      string     `json:"widget,omitempty"`
	SQLType     string     `json:"sqlType,omitempty"`
	SQLTab      string     `json:"sqlTab,omitempty"`
	Required    bool       `json:"required"`
	NotNull     bool       `json:"notNull"`
	Comment     string     `json:"comment,omitempty"`     // cmt_*
	TextKey     string     `json:"textKey,omitempty"`     // lbl_*
	Label       string     `json:"label"`                 // 中文名；未解析时**必为空串**
	LabelState  LabelState `json:"labelState"`            // resolved | unresolved
	LabelFrom   string     `json:"labelFrom,omitempty"`   // 命中的字典位置：<表>.<字段>
	LabelReason string     `json:"labelReason,omitempty"` // 未解析的原因（resolved 时为空）
}

// Table 是一张表单的字段表（字段顺序 = `.42f` 里的声明顺序，不重排 —— 位置信息在顺序里）。
type Table struct {
	Module string  `json:"module"`
	Form   string  `json:"form"`
	Path   string  `json:"path,omitempty"` // 实际读到的 `.42f`（只读；便于核对版本）
	Fields []Field `json:"fields"`

	Duplicates []string `json:"duplicates"` // 同表单内重名的 colName（原样带出，不猜）
	Skipped    int      `json:"skipped"`    // 缺属性的条目数（解析器跳过并计数）

	LabelResolved   int    `json:"labelResolved"`   // 中文名解析到的条数
	LabelUnresolved int    `json:"labelUnresolved"` // 标"未解析"的条数
	LabelSource     string `json:"labelSource"`     // dict | none

	Notes []string `json:"notes,omitempty"` // 查表报错之类，如实记账
}

// ── 取件：只读本地目录 ──────────────────────────────────────────────────────

// candidates 是 42f 根目录下表单文件的候选位置（顺序即优先级，先命中先赢）：
//
//  1. <root>/<module>/42f/<form>.42f —— 计划里写的数据源形态（模块下带 42f 子目录）
//  2. <root>/<module>/<form>.42f    —— 成果目录里真样本的形态（fixtures/forms/t100/axmt500_wf.42f）
//
// module 为空时两种都退化成 <root>/[42f/]<form>.42f（Genero 自带样本就平铺在根下）。
// 这里只做"文件在不在"，不读内容、不改任何文件。
func candidates(root, module, form string) []string {
	out := []string{
		filepath.Join(root, module, "42f", form+".42f"),
		filepath.Join(root, module, form+".42f"),
	}
	seen := make(map[string]bool, len(out))
	uniq := out[:0]
	for _, p := range out {
		if seen[p] {
			continue
		}
		seen[p] = true
		uniq = append(uniq, p)
	}
	return uniq
}

// FindPath 定位一张表单的 `.42f`。找不到就报错并**列出尝试过的路径**（不猜、不换名）。
func FindPath(root, module, form string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("42f 根目录为空：用 --root 或环境变量指定（不连真机，只读本地目录）")
	}
	if strings.TrimSpace(form) == "" {
		return "", fmt.Errorf("表单名为空")
	}
	tried := candidates(root, module, form)
	for _, p := range tried {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("找不到表单 %s.%s 的 .42f（根目录 %s）。尝试过：\n  %s",
		module, form, root, strings.Join(tried, "\n  "))
}

// LoadForm 读入 `<root>/<module>[/42f]/<form>.42f`，解析成字段表，再把中文名接到字典上。
//
// namer 为 nil 时不做字典查询，整列标 `unresolved/no-dict` —— 这是**合法**结果
// （没字典就不该有名字），不是错误。
func LoadForm(root, module, form string, namer FieldNamer) (*Table, error) {
	path, err := FindPath(root, module, form)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读 %s 失败: %w", path, err)
	}
	f, err := fep.ParseForm42F(form, string(b))
	if err != nil {
		// 解析失败归 FORM_UNKNOWN（契约 C1.7 / C4.2）—— 原样上传，由命令层定退出码
		return nil, fmt.Errorf("%s 不是可用的 .42f: %w", path, err)
	}
	tab := Merge(f, module, namer)
	tab.Path = path
	return tab, nil
}

// ── 合并：字段表 × 字典 ────────────────────────────────────────────────────

// Merge 把解析结果并成字段表。**纯函数**（不读盘、不连库，只走调用方给的 namer），
// 所以"不猜"这条规矩可以在合成夹具上钉死。
func Merge(f *fep.F42Form, module string, namer FieldNamer) *Table {
	tab := &Table{
		Module:      module,
		Form:        string(f.Name),
		Fields:      make([]Field, 0, len(f.Fields)),
		Duplicates:  f.Duplicates,
		Skipped:     f.Skipped,
		LabelSource: LabelSourceNone,
	}
	if tab.Duplicates == nil {
		tab.Duplicates = []string{}
	}

	tables := formTables(f)
	cache := newTableCache(namer, tab)

	for _, fld := range f.Fields {
		out := Field{
			ColName:  string(fld.ColName),
			FieldID:  int(fld.FieldID),
			Tag:      fld.Tag,
			Widget:   fld.Widget,
			SQLType:  fld.SQLType,
			SQLTab:   fld.SQLTab,
			Required: fld.Required,
			NotNull:  fld.NotNull,
			Comment:  fld.Comment,
			TextKey:  fld.TextKey,
		}
		out.Label, out.LabelFrom, out.LabelReason = cache.resolve(fld, tables)
		if out.Label == "" {
			out.LabelState = LabelUnresolved
			tab.LabelUnresolved++
		} else {
			out.LabelState = LabelResolved
			tab.LabelResolved++
		}
		tab.Fields = append(tab.Fields, out)
	}
	if namer != nil {
		tab.LabelSource = LabelSourceDict
	}
	return tab
}

// formTables 收集**这张表单声明自己跑在哪几张表上**（证据，全部来自表单文件本身）：
//
//	① `RecordView tabName=…`（含其 `<Link colName=…>` 覆盖的字段）—— 表单的记录视图；
//	② 字段条目自己的 `sqlTabName`。
//
// 为什么**不**按 colName 逐字段定表：实测 `axmt500_wf.42f` 里明细网格的身份是
// `formonly.b_xmda001`（`sqlTabName=formonly`），而它真正的表是 `xmda_t` ——
// 表单把"画面缓冲区列"挂在 `formonly` 这个**伪表**上。所以"哪个字段属于哪张表"不能靠
// 表单里的单点声明去断言（那会得出 formonly，然后全军未解析），只能这样判：
// **表单声明的表集合 × 字典的字段名等值匹配**，命中唯一名字才算解析到。
//
// 这**不是**启发式：没有按名字反推表名（"b_xmda001 大概是 xmda_t"那种），
// 表集合是表单自己列的，字段名是 `cmt_*`/`lbl_*`/`colName` 明写的；两张表给出不同名字时
// 一律判 `ambiguous`（见 resolve）。
func formTables(f *fep.F42Form) []string {
	set := map[string]bool{}
	add := func(tab string) {
		tab = strings.TrimSpace(tab)
		if tab != "" {
			set[tab] = true
		}
	}
	for _, v := range f.RecordViews {
		add(v.TabName)
	}
	for _, fld := range f.Fields {
		add(fld.SQLTab)
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out) // 顺序固定：同样的输入给同样的输出
	return out
}

// tableCache 一次 Merge 内**每张表只查一次**（563 个字段 × 2 张表 = 1000+ 次查询太浪费，
// 而字典的 `dzeb_t` 一张表一次就够）。
type tableCache struct {
	namer FieldNamer
	tab   *Table
	rows  map[string][]db.TableInfo
	fails map[string]bool
}

func newTableCache(namer FieldNamer, tab *Table) *tableCache {
	return &tableCache{
		namer: namer,
		tab:   tab,
		rows:  map[string][]db.TableInfo{},
		fails: map[string]bool{},
	}
}

// resolve 返回 (中文名, 位置, 未解析原因)；中文名为空串即"没解析到"。
//
// 判定只看"字典给不给名字"与"名字是否唯一"，不看任何启发式（A15/A18）。
func (c *tableCache) resolve(fld fep.F42Field, tables []string) (label, from, reason string) {
	if c.namer == nil {
		return "", "", ReasonNoDict
	}
	keys := candidateKeys(fld)
	if len(tables) == 0 || len(keys) == 0 {
		return "", "", ReasonNoTable
	}

	names := map[string]string{} // 中文名 → 第一次命中的位置
	for _, table := range tables {
		for _, row := range c.rowsOf(table) {
			if row.FieldDesc == "" {
				continue
			}
			for _, key := range keys {
				if !strings.EqualFold(row.FieldName, key) {
					continue
				}
				if _, ok := names[row.FieldDesc]; !ok {
					names[row.FieldDesc] = table + "." + row.FieldName
				}
			}
		}
	}

	switch len(names) {
	case 0:
		return "", "", ReasonNotInDict
	case 1:
		for name, where := range names {
			return name, where, ""
		}
	}
	// 多个候选给出不同的名字：**不挑**（契约 C3.4 的同一条精神：绝不自己挑一个）
	return "", "", ReasonAmbiguous
}

// rowsOf 取一张表的字段档（带一次缓存；错误只记一次 note，不冒充命中）。
func (c *tableCache) rowsOf(table string) []db.TableInfo {
	if rows, ok := c.rows[table]; ok {
		return rows
	}
	rows, err := c.namer.QueryTable(table)
	if err != nil {
		c.rows[table] = nil
		// "字典里没这张表"是**正常**情形（明细列的 sqlTabName 常写 formonly/undefined），
		// 不当噪声；其它错误记账，因为"为什么没中文名"得查得下去。
		if !db.IsMissingTable(err) && !c.fails[table] {
			c.fails[table] = true
			c.tab.Notes = append(c.tab.Notes,
				fmt.Sprintf("字典查表 %s 失败：%v", table, err))
		}
		return nil
	}
	c.rows[table] = rows
	return rows
}

// candidateKeys 给出**查字典用的键**候选（顺序固定、去重）。
//
// 来源只有三处，全部是表单文件明写的属性，**不从 colName 反推表名、不猜字段**：
//
//	cmt_xmda001  → xmda001（再试 xmda001 去掉 b_ 的形式）
//	lbl_xmda001  → xmda001
//	colName b_xmda001 → xmda001（`b_` 是画面缓冲区前缀，不是字典里的字段名）
//
// 注意：`colName` **原样**不当键用（字典 `dzeb_t` 里存的是物理字段名 `xmda001`）——
// 键取错了最多查不到（→ 未解析），永远不会变成"编出来的中文名"。
func candidateKeys(fld fep.F42Field) []string {
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		for _, x := range out {
			if strings.EqualFold(x, s) {
				return
			}
		}
		out = append(out, s)
	}
	addKey := func(prefix, v string) {
		v = strings.TrimSpace(v)
		if v == "" {
			return
		}
		if len(v) >= len(prefix) && strings.EqualFold(v[:len(prefix)], prefix) {
			v = v[len(prefix):]
		}
		add(v)
		add(trimPrefixFold(v, "b_"))
	}
	addKey("cmt_", fld.Comment)
	addKey("lbl_", fld.TextKey)
	add(trimPrefixFold(string(fld.ColName), "b_"))
	return out
}

// trimPrefixFold 大小写不敏感地去掉前缀（T100 的列名都是小写，但不靠这个假设过活）。
func trimPrefixFold(s, prefix string) string {
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):]
	}
	return s
}
