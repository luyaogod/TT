package devdata

// 材料角色表：**唯一事实来源**。改"拉什么"请改这里，别在 mirror.go/fetch.go 里加散装路径。
//
// 分两类（依据是"是否绑定具体程序"，不是体积 —— plan §1.1 约束 3）：
//
//	M（mirror，环境级）：53 张 DZ*_T 规格表 + 数据字典表 + $TOP/erp/mdl 模板树
//	D（fetch，程序级） ：tab/tsd/tap/tgl/4gl/inc/42f/design，一次拉齐成套
//
// 为什么"按程序成套"而不是散装下载（ADR-04）：model 的第一类结论（"这一行为什么长这样"）
// 依赖 tab（蓝图）与 tgl（排序骨架）来自同一时刻。散装会让两者错配，而错配的产物看起来
// 完全正常，只是答案微妙地不对 —— 这类错误没有报错可循。

import "fmt"

// 种类（单飞锁的维度，也是 HTTP/CLI 的 kind 取值）。
const (
	// KindMirror 共用资料（环境级，plan 里的 M）。
	KindMirror = "mirror"
	// KindFetch 按需资料（程序级，plan 里的 D）。
	KindFetch = "fetch"
)

// 单表 / 单文件 / 总量的上限（plan §5.4、§10）。
const (
	// MaxTableRows 单表行数上限：超过则该表记为 skipped 并写 warnings（不判失败）。
	//
	// 实测（2026-11-21 测试环境）：53 张表合计 18,371,187 行，其中 DZBA_T 7,810,349、
	// DZAZ_T 4,622,785 —— 这两张是生成器的**运行时产物/计数**，不是入参。上限挡住它们，
	// 也挡住任何一次"某环境数据量暴涨把镜像拖死"。
	MaxTableRows = 2_000_000
	// MaxMdlBytes 模板树总量上限：超了直接拒绝（而不是跳过某些文件 —— 模板缺一片
	// 就是"材料不齐"，与"材料很全但多到拉不动"是两回事）。
	MaxMdlBytes = 200 << 20
	// MaxFileBytes 单个文件上限（.42f 这类二进制尤其要拦）。
	MaxFileBytes = 64 << 20
)

// TableSpec 一张要镜的规格表。
type TableSpec struct {
	// Name 大写表名（如 DZAA_T）。SQL 里不带 schema：T100 的 DZ*_T 有同义词，
	// dict 侧（tt dict r.t 等）也是同样的写法，两边不会因为 schema 前缀分叉。
	Name string
	// Desc 中文表说明（取自 DZEAL_T zh_CN，2026-11-21 实测）。
	Desc string
	// Fields 字段数（取自 DZEB_T，2026-11-21 实测）。
	Fields int
	// Rows 首次实测行数（2026-11-21 测试环境@1108）。只作参考基线：
	// 真正的行数以每次 mirror 的实测为准，写进 manifest。
	Rows int64
	// Note 生成器侧用途的一句话说明（详见 .agents/docs/t100/devdata-materials.md）。
	Note string
	// LatestOnly 只镜"当前有效版"（ADR-03 的补充口径）：`*STUS='Y'` 且版次为该程序最大值。
	//
	// 只有 DZAA_T 用得上：它是"规格与内容版本对应表"，全表 70 万行 / 5,445 个程序 /
	// 每程序平均百来行内容项。其余表**不做**通用 STUS 过滤 —— 实测 DZEB_T/DZED_T 的
	// `*STUS='Y'` 行数为 0，说明状态码并非全表统一，通用过滤会静默清空整张表。
	LatestOnly bool
}

// MirrorTables 首版镜像的 53 张 DZ*_T（顺序即镜像顺序，按表名字典序）。
//
// 清单来源：对 `.agents/docs/t100/codegen/01-server` 做 `FROM dz[a-z0-9_]+_t` 的提取
// （plan 附录 A）。roles_test.go 会对着计划文件里的附录 A 做一致性断言 —— 这份清单
// 与文档不许漂移。
//
// 「数据字典」那几张（r.t 的 dzea_t/dzeal_t/dzeb_t/dzed_t…）已经在上面这 53 张里
// （DZEB/DZED 在列），所以不再单独列一遍。
var MirrorTables = []TableSpec{
	{Name: "DZAA_T", Desc: "规格与内容版本对应表", Fields: 18, Rows: 702438, LatestOnly: true,
		Note: "程式设计主档：dzaa001 程序 / dzaa002 版次 / dzaa003 内容项 / dzaa009 identity；生成器先读它判版本"},
	{Name: "DZAB_T", Desc: "规格整体设计表", Fields: 15, Rows: 5921,
		Note: "程序的整体设计设定（判型/样板选择入口）"},
	{Name: "DZAC_T", Desc: "字段规格设计表", Fields: 30, Rows: 212109,
		Note: "栏位检核（min/max/compare），对应产物的 cl_ap_chk_range 段"},
	{Name: "DZAD_T", Desc: "Action规格设计表", Fields: 16, Rows: 22976,
		Note: "Action（新增/修改/删除/查询…）规格"},
	{Name: "DZAF_T", Desc: "规格与程序版本历程对应表", Fields: 10, Rows: 47299,
		Note: "规格 → 程序版本的历史对应"},
	{Name: "DZAG_T", Desc: "规格Table设置表", Fields: 23, Rows: 7139,
		Note: "表关联：parent / fk_detail / fk_master / upper_table / upper_key / this_key"},
	{Name: "DZAI_T", Desc: "字段参考设计表", Fields: 19, Rows: 38855,
		Note: "栏位明细的参考来源"},
	{Name: "DZAJ_T", Desc: "字段数据多语言设计表", Fields: 20, Rows: 867,
		Note: "栏位多语言文案"},
	{Name: "DZAK_T", Desc: "字段助记码设计表", Fields: 19, Rows: 89462,
		Note: "栏位助记码"},
	{Name: "DZAL_T", Desc: "程序串查设计表", Fields: 17, Rows: 1441,
		Note: "串查（drill-down）设定"},
	{Name: "DZAM_T", Desc: "规格画面元件排除设置", Fields: 14, Rows: 7337,
		Note: "画面元件排除清单"},
	{Name: "DZAQ_T", Desc: "行业程序同步纪录表", Fields: 22, Rows: 64,
		Note: "行业程序同步记录"},
	{Name: "DZAS_T", Desc: "行业非主程序与标准程序关系表", Fields: 8, Rows: 0,
		Note: "行业非主程序 ↔ 标准程序；本环境为空"},
	{Name: "DZAT_T", Desc: "设计数据脱机档备份", Fields: 20, Rows: 0,
		Note: "设计数据的脱机备份；本环境为空"},
	{Name: "DZAU_T", Desc: "行业对应之标准规格引用清单", Fields: 6, Rows: 1254,
		Note: "行业 → 标准规格的引用清单"},
	{Name: "DZAX_T", Desc: "编程基本设置表", Fields: 15, Rows: 9748,
		Note: "程式设计基本设定（判型入口）"},
	{Name: "DZAY_T", Desc: "dzax_t 的暂存表（取消签出使用）", Fields: 5, Rows: 152,
		Note: "DZAX_T 的暂存/回滚副本"},
	{Name: "DZAZ_T", Desc: "生成器生成计数表", Fields: 4, Rows: 4622785,
		Note: "生成器运行时计数（产物侧，不是入参）；行数超上限会被跳过"},
	{Name: "DZBA_T", Desc: "程序与内容版本对应表", Fields: 20, Rows: 7810349,
		Note: "内容项 ↔ 版本对应（代码段级别）；行数超上限会被跳过"},
	{Name: "DZBB_T", Desc: "编程点设计表", Fields: 16, Rows: 611403,
		Note: "加点定义：point 名 / 版本 / 内容"},
	{Name: "DZBC_T", Desc: "section标准数据档", Fields: 19, Rows: 971136,
		Note: "section 标准资料，决定 .tgl 的排序"},
	{Name: "DZBD_T", Desc: "section设计表", Fields: 14, Rows: 242683,
		Note: "section 设计结果"},
	{Name: "DZBH_T", Desc: "Autocomplete功能SQL维护表", Fields: 12, Rows: 19250,
		Note: "autocomplete（a77 片段）的 SQL 维护"},
	{Name: "DZBK_T", Desc: "差异比对Patch编号条件设置", Fields: 10, Rows: 0,
		Note: "差异比对 Patch 编号条件；本环境为空"},
	{Name: "DZBU_T", Desc: "行业对应之标准add point引用清单", Fields: 5, Rows: 2407,
		Note: "行业 → 标准 add point 引用清单"},
	{Name: "DZCA_T", Desc: "开窗数据表", Fields: 16, Rows: 5261,
		Note: "可复用开窗（r.q）的元数据"},
	{Name: "DZCB_T", Desc: "开窗设计参数设计表", Fields: 12, Rows: 4444,
		Note: "开窗参数"},
	{Name: "DZCC_T", Desc: "开窗显现设置表格", Fields: 18, Rows: 32327,
		Note: "开窗显现（显示哪些列）"},
	{Name: "DZCD_T", Desc: "校验带值设计表", Fields: 15, Rows: 1459,
		Note: "栏位检核的带值（r.v）"},
	{Name: "DZCE_T", Desc: "校验带值参数设计表", Fields: 15, Rows: 2402,
		Note: "带值参数"},
	{Name: "DZCP_T", Desc: "公式设置主表", Fields: 13, Rows: 0,
		Note: "公式设定；本环境为空"},
	{Name: "DZCT_T", Desc: "带值设置主表", Fields: 17, Rows: 0,
		Note: "带值设定主档；本环境为空"},
	{Name: "DZCY_T", Desc: "参数值设置参数表", Fields: 15, Rows: 0,
		Note: "参数值设定；本环境为空"},
	{Name: "DZDY_T", Desc: "样板基本数据表", Fields: 13, Rows: 155,
		Note: "样板（template）基本资料"},
	{Name: "DZDZ_T", Desc: "样板及版本更新基本数据表", Fields: 10, Rows: 1648,
		Note: "样板版本与更新"},
	{Name: "DZEB_T", Desc: "数据表字段档", Fields: 28, Rows: 150665,
		Note: "数据字典的字段档（r.t/desc 的来源）"},
	{Name: "DZED_T", Desc: "数据表键值档", Fields: 20, Rows: 5540,
		Note: "数据字典的键值档（主键/外键定义）"},
	{Name: "DZEM_T", Desc: "SA分镜规格导入接口档", Fields: 11, Rows: 0,
		Note: "SA 分镜规格汇入接口；本环境为空"},
	{Name: "DZEP_T", Desc: "字段规格设计数据档", Fields: 38, Rows: 158023,
		Note: "栏位画面规格（desc 的来源）"},
	{Name: "DZFF_T", Desc: "树状结构属性设置档", Fields: 17, Rows: 449,
		Note: "树状结构属性"},
	{Name: "DZFI_T", Desc: "画面结构档", Fields: 25, Rows: 410088,
		Note: "page ↔ field ↔ tbl 挂载"},
	{Name: "DZFJ_T", Desc: "画面元件组合档", Fields: 19, Rows: 899991,
		Note: "画面元件组合"},
	{Name: "DZFK_T", Desc: "画面元件属性数据档", Fields: 9, Rows: 1120856,
		Note: "画面元件属性"},
	{Name: "DZFQ_T", Desc: "画面结构设计主档", Fields: 24, Rows: 5192,
		Note: "画面结构设计主档"},
	{Name: "DZFS_T", Desc: "程序Table与Screen Record对应档", Fields: 20, Rows: 8435,
		Note: "程序 Table ↔ Screen Record 对应"},
	{Name: "DZFU_T", Desc: "画面样版ToolBar功能键设置档", Fields: 10, Rows: 98,
		Note: "ToolBar 功能键"},
	{Name: "DZGC_T", Desc: "报表元件设计-数据模型字段明细档", Fields: 9, Rows: 78228,
		Note: "报表元件的数据模型字段"},
	{Name: "DZGD_T", Desc: "报表元件设计-数据模型自订义字段明细档", Fields: 8, Rows: 26338,
		Note: "报表元件的自订字段"},
	{Name: "DZGG_T", Desc: "报表元件设计-样板单头档(GR)", Fields: 17, Rows: 724,
		Note: "报表样板单头"},
	{Name: "DZKC_T", Desc: "报表组件设计-数据模型字段明细档(XR)", Fields: 18, Rows: 23769,
		Note: "XR 报表组件字段"},
	{Name: "DZKD_T", Desc: "报表组件设计-自订字段明细档(XR)", Fields: 16, Rows: 5196,
		Note: "XR 报表自订字段"},
	{Name: "DZYE_T", Desc: "设计数据运行结果记录表", Fields: 24, Rows: 2824,
		Note: "设计数据运行结果"},
	{Name: "DZZC_T", Desc: "简易画面数据设置", Fields: 13, Rows: 0,
		Note: "简易画面数据设定；本环境为空"},
}

// MirrorTableNames 取表名清单（给 SQL 拼装与测试用）。
func MirrorTableNames() []string {
	out := make([]string, len(MirrorTables))
	for i := range MirrorTables {
		out[i] = MirrorTables[i].Name
	}
	return out
}

// TableByName 按（大小写不敏感的）表名取一条；没有返回 nil。
func TableByName(name string) *TableSpec {
	for i := range MirrorTables {
		if equalFold(MirrorTables[i].Name, name) {
			return &MirrorTables[i]
		}
	}
	return nil
}

// FileRole 一个"文件素材"的角色定义（D 侧按程序成套，M 侧的模板树也用它描述落点）。
type FileRole struct {
	// Role manifest 里的角色名（tab/tsd/tap/tgl/tgx/tgx2/4gl/inc/form/design/mdl）。
	Role string
	// Src 服务器侧相对 $TOP 的路径模板（%s 依次为 mod、prog）。含 `<prog>` 的地方由
	// fmt.Sprintf 填；M 侧的模板树用 SrcDir + 逐文件相对路径，不看这里。
	Src string
	// Rel 本地相对程序目录的落点（%s 为 prog）。
	Rel string
	// Required 必需：缺失即整个 fetch 失败（不会留下半套）。
	Required bool
	// Form 只在 --form 时拉（.42f 是二进制、体积大，默认不拉）。
	Form bool
}

// FetchRoles 按程序成套拉取的文件角色表（顺序即落盘顺序）。
//
// 必需项的判据是"model 少了它就无法判断"：蓝图 tab、排序骨架 tgl、产物 4gl。
// tsd/tap/tgx/tgx2/inc 是"存在才有"（有的程序不产这些），缺失只进 manifest 的
// missing[]，不判失败 —— 把可选缺失当失败会让 fetch 在一半程序上不可用。
var FetchRoles = []FileRole{
	{Role: "tab", Src: "erp/%s/dzx/tab/%s.tab", Rel: "dzx/%s.tab", Required: true},
	{Role: "tsd", Src: "erp/%s/dzx/tsd/%s.tsd", Rel: "dzx/%s.tsd"},
	// 4fd 是表单定义（页面结构）：它把 .tab 的 screen record 与 .tsd 的页签键接起来，
	// 是"客户说的第 12 个页签到底是哪个 page_id"的唯一硬证据。约 0.9 MB，可缺。
	{Role: "4fd", Src: "erp/%s/4fd/%s.4fd", Rel: "dzx/%s.4fd"},
	{Role: "tap", Src: "erp/%s/dzx/tap/%s.tap", Rel: "dzx/%s.tap"},
	{Role: "tgl", Src: "erp/%s/dzx/tgl/%s.tgl", Rel: "dzx/%s.tgl", Required: true},
	{Role: "tgx", Src: "erp/%s/dzx/tgl/%s.tgx", Rel: "dzx/%s.tgx"},
	{Role: "tgx2", Src: "erp/%s/dzx/tgl/%s.tgx2", Rel: "dzx/%s.tgx2"},
	{Role: "4gl", Src: "erp/%s/4gl/%s.4gl", Rel: "src/%s.4gl", Required: true},
	{Role: "inc", Src: "com/inc/erp/%s/%s_mask.4gl", Rel: "inc/%s_mask.4gl"},
	{Role: "inc", Src: "com/inc/erp/%s/%s_rep.4gl", Rel: "inc/%s_rep.4gl"},
	{Role: "form", Src: "erp/%s/42f/%s.42f", Rel: "form/%s.42f", Form: true},
}

// MdlRoot 模板树在服务器上的相对位置（保持相对路径整体落盘，plan §5.1）。
const MdlRoot = "erp/mdl"

// SrcPath 把一个角色的源路径模板填成实际路径。
func (r FileRole) SrcPath(mod, prog string) string {
	return fmt.Sprintf(r.Src, mod, prog)
}

// RelPath 把一个角色的落点模板填成实际相对路径。
func (r FileRole) RelPath(prog string) string {
	return fmt.Sprintf(r.Rel, prog)
}

// equalFold 是 strings.EqualFold 的本地包装（避免只为一次比较 import strings 到角色表里）。
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
