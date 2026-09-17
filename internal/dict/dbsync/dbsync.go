// Package dbsync 把 ERP 远程库的数据字典表同步到本地 SQLite(与 tt dict db sync 同一实现),
// 供 CLI 与 tt serve 的可视化页面共用。
//
// 流程:连远程库 → 逐表 SELECT → 写入临时 SQLite → 依据 dzed_t 的主键定义建唯一索引
// → 原子替换目标文件(原库备份为 <库>.bak)。中途失败只清理临时库,不动原库。
package dbsync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tt/internal/dbconfig"
	"tt/internal/dict/db"
	"tt/internal/erpdb"
)

// Family 一个数据族:一族 = 一个查询命令所需的一组字典表。
// Families 是同步清单、`tt dict db status` 与各命令 --help 里「本地数据齐不齐」
// 提示的唯一事实来源(改同步范围请改这里)。
type Family struct {
	Key      string   `json:"键"`    // 稳定标识(表族键,如 table/check/scc/prog)
	Name     string   `json:"数据族"`  // 中文名
	Commands []string `json:"依赖命令"` // 依赖该族的命令
	Tables   []string `json:"字典表"`  // 该族的字典表
}

// Families 全部数据族,顺序即同步顺序(表字典一族放在最前,先有表才有别的)。
var Families = []Family{
	{Key: "table", Name: "表字典", Commands: []string{"r.t"},
		Tables: []string{"dzea_t", "dzeal_t", "dzeb_t", "dzebl_t", "dzec_t", "dzed_t", "dzee_t", "dzef_t", "dzeg_t"}},
	{Key: "check", Name: "校验带值", Commands: []string{"r.v"},
		Tables: []string{"dzcd_t", "dzcdl_t", "dzce_t", "dzcel_t", "dzch_t"}},
	{Key: "scc", Name: "系统分类码", Commands: []string{"scc"},
		Tables: []string{"gzca_t", "gzcal_t", "gzcb_t", "gzcbl_t"}},
	{Key: "spec", Name: "字段画面规格", Commands: []string{"desc"},
		Tables: []string{"dzep_t"}},
	{Key: "win", Name: "可复用开窗", Commands: []string{"r.q"},
		Tables: []string{"dzca_t", "dzcal_t", "dzcb_t", "dzcbl_t", "dzcc_t"}},
	{Key: "msg", Name: "系统消息", Commands: []string{"msg"},
		Tables: []string{"gzze_t", "gzzal_t"}},
	{Key: "param", Name: "参数定义", Commands: []string{"sysp", "docp"},
		Tables: []string{"gzsz_t", "gzszl_t", "gzsy_t"}},
	// 程序与作业 (azzi900 程式基本資料 / azzi910 作業基本資料):
	// 程序档 + 程序名称多语言 + 作业编号设置表(作业用 gzzz002 挂程序,一个程序可被多个作业使用)
	// + 程序应用参数组设置表。gzzal_t 同时服务 msg 族,展平时去重。
	{Key: "prog", Name: "程序与作业", Commands: []string{"prog"},
		Tables: []string{"gzza_t", "gzzz_t", "gzzk_t", "gzzal_t"}},
	// 程序 ↔ 表格(gzdg_t 程序与应用表格功能分析表,由 T100 自己维护;参考作业 azzq902
	// 程式編號對應表格查詢):主键 = 程序编号 + 表格编号 + 功能类别(SCC 212: I/S/U/D)。
	// 表名与程序名分别 join dzeal_t(表字典族)与 gzzal_t(程序族),两者已在库中。
	{Key: "progtable", Name: "程序与表格", Commands: []string{"prog", "r.t"},
		Tables: []string{"gzdg_t"}},
	// 子程序与元件(gzde_t 子程序及应用元件基本数据表 + gzdel_t 说明,参考作业 azzi901
	// 子程式及元件基本資料設定作業):与 gzza_t(主程序)互不重叠——实测正式区
	// gzde_t 4,036 个 / gzza_t 4,147 个、交集 0。规格类别见 SCC 91。
	{Key: "subprog", Name: "子程序与元件", Commands: []string{"prog"},
		Tables: []string{"gzde_t", "gzdel_t"}},
}

// DictTables 同步的 T100 数据字典表 —— 由 Families 展平并按首次出现去重。
var DictTables = flattenFamilies()

func flattenFamilies() []string {
	seen := make(map[string]bool)
	var out []string
	for _, f := range Families {
		for _, t := range f.Tables {
			if seen[t] {
				continue
			}
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

// FamilyByKey 按 key 取数据族;不存在返回 nil。
func FamilyByKey(key string) *Family {
	for i := range Families {
		if Families[i].Key == key {
			return &Families[i]
		}
	}
	return nil
}

// Progress 同步过程回调(CLI 打印/Web 进度条用)。
// Phase: open(连远程库) | table(逐表拉取) | index(建索引) | replace(替换本地库) | done。
// table 阶段首尾各回调一次:开始 TableDur=0、TableRows=0;完成时带上行数与耗时。
type Progress struct {
	Phase      string
	Table      string
	TableIndex int // 1-based(当前第几张表)
	TableTotal int
	TableRows  int
	TableDur   time.Duration
	TotalRows  int
	Message    string
}

// Stats 同步结果统计。
type Stats struct {
	Target   string   // 目标 SQLite 路径
	Tables   int      // 同步表数
	Rows     int      // 总行数
	Backup   string   // 原库备份路径(原库存在时)
	Warnings []string // 非致命警告(索引创建失败等)
	Elapsed  string
}

// Run 执行同步:tables 为空时用 DictTables。conn 的直连凭据(账号列表首项)在内部填充。
// onProgress 可为 nil。返回的 Stats 在成功时为非 nil。
func Run(ctx context.Context, conn dbconfig.Connection, target string, tables []string, onProgress func(Progress)) (*Stats, error) {
	report := func(p Progress) {
		if onProgress != nil {
			onProgress(p)
		}
	}
	start := time.Now()
	if len(tables) == 0 {
		tables = DictTables
	}
	st := &Stats{Target: target}

	// 客户端直连凭据 = 账号列表首项(SelectAllSQL 用它限定 user.table schema)
	if err := conn.FillDialCred(); err != nil {
		return nil, err
	}

	report(Progress{Phase: "open", Message: "连接远程数据库…"})
	ext, err := erpdb.Open(ctx, conn)
	if err != nil {
		return nil, err
	}
	defer ext.Close()

	// 写入临时库,全部成功后原子替换目标文件,避免中途失败留下半成品
	// 目标目录不存在就创建(便携版/自定义路径首跑)
	if dir := filepath.Dir(target); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("创建目标目录失败 (%s): %w", dir, err)
		}
	}
	tmpPath := target + ".sync.tmp"
	os.Remove(tmpPath)
	sq, err := db.Open(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("创建临时数据库失败 (%s): %w", tmpPath, err)
	}
	if err := sq.SetBulkWriteMode(); err != nil {
		sq.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("配置临时数据库失败: %w", err)
	}
	fail := func(err error) (*Stats, error) {
		sq.Close()
		os.Remove(tmpPath)
		return nil, err
	}

	total := len(tables)
	totalRows := 0
	var dzedCols []string
	var dzedRows [][]string
	for i, t := range tables {
		report(Progress{Phase: "table", Table: t, TableIndex: i + 1, TableTotal: total,
			TotalRows: totalRows, Message: "拉取 " + t})
		ts := time.Now()
		sql, err := erpdb.SelectAllSQL(conn, t)
		if err != nil {
			return fail(err)
		}
		cols, rows, err := ext.Query(ctx, sql)
		if err != nil {
			return fail(fmt.Errorf("查询表 %s 失败: %w", t, err))
		}
		n, err := sq.RebuildTable(t, cols, rows)
		if err != nil {
			return fail(fmt.Errorf("写入表 %s 失败: %w", t, err))
		}
		if t == "dzed_t" {
			dzedCols, dzedRows = cols, rows
		}
		totalRows += n
		report(Progress{Phase: "table", Table: t, TableIndex: i + 1, TableTotal: total,
			TableRows: n, TableDur: time.Since(ts), TotalRows: totalRows, Message: "已写入 " + t})
	}

	// 依据 dzed_t 的 PK 定义 (dzed003='P') 为本次同步的表创建唯一索引
	synced := make(map[string]bool, len(tables))
	for _, t := range tables {
		synced[t] = true
	}
	report(Progress{Phase: "index", TableTotal: total, TotalRows: totalRows, Message: "创建主键索引…"})
	st.Warnings = createPKIndexes(sq, dzedCols, dzedRows, synced)
	sq.Close()

	report(Progress{Phase: "replace", TableTotal: total, TotalRows: totalRows, Message: "替换本地数据库…"})
	// 原子替换: 原库备份为 .bak, 临时库替换为正式库
	if _, err := os.Stat(target); err == nil {
		bak := target + ".bak"
		os.Remove(bak)
		if err := os.Rename(target, bak); err != nil {
			os.Remove(tmpPath)
			return nil, fmt.Errorf("备份原数据库失败 (%s): %w", bak, err)
		}
		st.Backup = bak
	}
	if err := os.Rename(tmpPath, target); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("替换数据库失败 (%s): %w", target, err)
	}
	// 清理可能残留的 WAL/SHM 孤儿文件 (旧库为 WAL 模式时)
	os.Remove(target + "-wal")
	os.Remove(target + "-shm")

	st.Tables = len(tables)
	st.Rows = totalRows
	st.Elapsed = time.Since(start).Round(10 * time.Millisecond).String()
	report(Progress{Phase: "done", TableTotal: total, TotalRows: totalRows,
		Message: fmt.Sprintf("同步完成:%d 张表,共 %d 行", len(tables), totalRows)})
	return st, nil
}

// createPKIndexes builds UNIQUE indexes for each synced table's primary key
// defined in dzed_t (dzed003 = 'P'). Returns non-fatal warning strings on failure.
func createPKIndexes(sq *db.DB, dzedCols []string, dzedRows [][]string, synced map[string]bool) []string {
	ti, ki, ty, fi := indexOf(dzedCols, "dzed001"), indexOf(dzedCols, "dzed002"),
		indexOf(dzedCols, "dzed003"), indexOf(dzedCols, "dzed004")
	if ti < 0 || ki < 0 || ty < 0 || fi < 0 || len(dzedRows) == 0 {
		return nil
	}
	var warnings []string
	for _, r := range dzedRows {
		if r[ty] != "P" {
			continue
		}
		if !synced[r[ti]] {
			continue // 该表本次未同步
		}
		table, key := r[ti], r[ki]
		var cols []string
		for _, f := range strings.Split(r[fi], ",") {
			f = strings.Trim(strings.TrimSpace(f), `"`)
			if f != "" {
				cols = append(cols, f)
			}
		}
		if len(cols) == 0 {
			continue
		}
		idxName := fmt.Sprintf("idx_%s_%s", table, key)
		if err := sq.CreateUniqueIndex(idxName, table, cols); err != nil {
			warnings = append(warnings, fmt.Sprintf("  [警告] 索引 %s 创建失败: %v", idxName, err))
		}
	}
	return warnings
}

func indexOf(cols []string, target string) int {
	for i, c := range cols {
		if c == target {
			return i
		}
	}
	return -1
}
