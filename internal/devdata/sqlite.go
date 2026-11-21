package devdata

// 规格库（SQLite）的写入。**本包自己读自己写**，不 import internal/dict/db ——
// 那是读侧（查询用）的实现，两边共用一份就等于 ADR-01 想拆开的东西又粘回去了。
// 落盘形态刻意与 dbsync 一致（全 TEXT 列、逐表重建），这样 `tt gen status` 与
// 后续 model 侧都能用同一套 SQL 读它。

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // 注册 sqlite 驱动（与 internal/dict/db 同一个实现）
)

// insertBatchRows 一批多少行提交一次（plan §10：分批事务，每 1000 行 commit）。
const insertBatchRows = 1000

// sqliteWriter 一个正在构建的规格库。
type sqliteWriter struct {
	db   *sql.DB
	path string

	// 当前表的插入语句（列名在建表时确定）
	insertSQL string
	cols      []string
}

// openSQLite 打开（必要时创建）一个 SQLite 文件。
func openSQLite(path string) (*sqliteWriter, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("创建目录失败 (%s): %w", dir, err)
		}
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败 (%s): %w", path, err)
	}
	// SQLite 单写者：连接池大于 1 只会换来 "database is locked"
	conn.SetMaxOpenConns(1)
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("打开数据库失败 (%s): %w", path, err)
	}
	// 临时存储放内存（理由同 internal/dict/db：某些目录下建临时文件会失败）
	_, _ = conn.Exec("PRAGMA temp_store=MEMORY")
	return &sqliteWriter{db: conn, path: path}, nil
}

// SetBulkMode 关日志与 fsync。**只对一次性的构建库安全**：崩溃时它可能损坏，
// 而这里的库正在临时目录里、坏了就整个重拉，不存在"丢失用户数据"。
func (w *sqliteWriter) SetBulkMode() error {
	for _, p := range []string{"PRAGMA journal_mode=OFF", "PRAGMA synchronous=OFF"} {
		if _, err := w.db.Exec(p); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	}
	return nil
}

// Close 关闭连接。
func (w *sqliteWriter) Close() {
	if w.db != nil {
		_ = w.db.Close()
	}
}

// validIdent 判一个 SQL 标识符（表名/列名）是否安全。
//
// 表名来自我们自己的角色表、列名来自**数据库返回的元数据** —— 都不是用户输入，
// 但仍然过一遍：这是拼 SQL 的唯一入口，宁可在这里多一道，也别让某天一个新来源绕过去。
func validIdent(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '$' || c == '#') {
			return false
		}
	}
	// 纯数字不是合法列名（Oracle 的 ROWID 之类会带字母，但保险起见挡掉）
	if s[0] >= '0' && s[0] <= '9' {
		return false
	}
	return true
}

func quoteIdent(s string) string { return `"` + s + `"` }

// BeginTable 重建一张表（DROP + CREATE，全 TEXT 列）。
//
// 全 TEXT 的理由：源侧文本化之后（`erpdb` 的列值一律字符串），写入不做任何类型推断
// 就不会"猜错类型"。数值比较在 SQLite 里靠 CAST，查询侧（r.t 那些）本来也是这么做的。
func (w *sqliteWriter) BeginTable(table string, cols []string) error {
	if !validIdent(table) {
		return fmt.Errorf("非法表名: %s", table)
	}
	if len(cols) == 0 {
		return fmt.Errorf("表 %s 没有列", table)
	}
	defs := make([]string, len(cols))
	quoted := make([]string, len(cols))
	for i, c := range cols {
		if !validIdent(c) {
			return fmt.Errorf("表 %s 有非法列名: %q", table, c)
		}
		defs[i] = quoteIdent(c) + " TEXT"
		quoted[i] = quoteIdent(c)
	}
	if _, err := w.db.Exec("DROP TABLE IF EXISTS " + quoteIdent(table)); err != nil {
		return fmt.Errorf("删除表 %s 失败: %w", table, err)
	}
	if _, err := w.db.Exec("CREATE TABLE " + quoteIdent(table) + " (" + strings.Join(defs, ", ") + ")"); err != nil {
		return fmt.Errorf("建表 %s 失败: %w", table, err)
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(cols)), ", ")
	w.cols = cols
	w.insertSQL = "INSERT INTO " + quoteIdent(table) + " (" + strings.Join(quoted, ", ") + ") VALUES (" + placeholders + ")"
	return nil
}

// InsertRows 写一批行（整批一个事务）。
//
// 每批一次事务而不是整表一个事务：整表事务在 112 万行时要攒一个巨大的 rollback journal
// （journal_mode=OFF 时不攒，但事务本身会持有大量页缓存）；每 1000 行提交一次，
// 内存与耗时都稳定。取消/失败时已提交的批次留在临时库里，而临时库最终会被删掉 ——
// 所以"半张表"永远不会出现在正式落点上。
func (w *sqliteWriter) InsertRows(rows [][]string) error {
	if len(rows) == 0 {
		return nil
	}
	if w.insertSQL == "" {
		return fmt.Errorf("InsertRows 前要先 BeginTable")
	}
	n := len(w.cols)
	tx, err := w.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(w.insertSQL)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, row := range rows {
		vals := make([]any, n)
		for i := 0; i < n; i++ {
			if i < len(row) {
				vals[i] = row[i]
			} else {
				vals[i] = ""
			}
		}
		if _, err := stmt.Exec(vals...); err != nil {
			_ = stmt.Close()
			_ = tx.Rollback()
			return err
		}
	}
	_ = stmt.Close()
	return tx.Commit()
}

// CountRows 数一张表的行数（写 manifest 时用；只查本地）。
func (w *sqliteWriter) CountRows(table string) (int64, error) {
	if !validIdent(table) {
		return 0, fmt.Errorf("非法表名: %s", table)
	}
	var n int64
	if err := w.db.QueryRow("SELECT COUNT(*) FROM " + quoteIdent(table)).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// Exec 执行一条内部 SQL（建索引用）。
func (w *sqliteWriter) Exec(query string) error {
	_, err := w.db.Exec(query)
	return err
}

// CreatePKIndexes 按本地 `dzed_t` 的主键定义建唯一索引，返回警告（尽力而为）。
//
// 为什么值得做：53 张表里 `DZEB_T`（15 万行）按 `dzeb001` 查是 model 侧最常见的动作，
// 没索引就是全表扫。失败不判整次 mirror 失败 —— 索引只影响速度，不影响"材料齐不齐"。
//
// 「一条都没建成」必须开口（与 dbsync 同一个教训）：dzed_t 没同步到、列名大小写对不上，
// 都会让这里静默地一个索引都不建，而镜像照样显示成功。
func (w *sqliteWriter) CreatePKIndexes(synced map[string]bool) []string {
	rows, err := w.db.Query(`SELECT dzed001, dzed002, dzed004 FROM dzed_t WHERE dzed003='P'`)
	if err != nil {
		return []string{"建主键索引跳过：读不到本地 dzed_t（" + err.Error() + "）"}
	}
	defer rows.Close()
	var warnings []string
	expected, created := 0, 0
	for rows.Next() {
		var table, key, fields string
		if err := rows.Scan(&table, &key, &fields); err != nil {
			warnings = append(warnings, "建主键索引跳过：读 dzed_t 出错（"+err.Error()+"）")
			break
		}
		if !synced[strings.ToLower(table)] {
			continue
		}
		var cols []string
		for _, f := range strings.Split(fields, ",") {
			f = strings.Trim(strings.TrimSpace(f), `"`)
			if f != "" && validIdent(f) {
				cols = append(cols, quoteIdent(f))
			}
		}
		if len(cols) == 0 {
			continue
		}
		expected++
		idx := "idx_" + strings.ToLower(table) + "_" + strings.ToLower(key)
		if !validIdent(idx) {
			continue
		}
		q := "CREATE UNIQUE INDEX IF NOT EXISTS " + quoteIdent(idx) + " ON " + quoteIdent(strings.ToLower(table)) +
			" (" + strings.Join(cols, ", ") + ")"
		if err := w.Exec(q); err != nil {
			warnings = append(warnings, fmt.Sprintf("索引 %s 建失败（%v）", idx, err))
			continue
		}
		created++
	}
	if expected > 0 && created == 0 {
		warnings = append(warnings, fmt.Sprintf(
			"本次应建 %d 个主键索引、实际建成 0 个：规格库查询会全表扫（先确认 dzed_t 已镜到，再看列名是否对得上）", expected))
	}
	return warnings
}

// CloseAndRename 关闭连接并把临时库改名成正式库（同一父目录 → 原子）。
func (w *sqliteWriter) CloseAndRename(final string, force bool) error {
	w.Close()
	if _, err := os.Stat(final); err == nil {
		if !force {
			return fmt.Errorf("%w: %s（加 --force 覆盖）", ErrExists, final)
		}
		if err := os.Remove(final); err != nil {
			return fmt.Errorf("删除旧库失败 (%s): %w", final, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(final), 0o755); err != nil {
		return fmt.Errorf("创建目录失败 (%s): %w", filepath.Dir(final), err)
	}
	if err := os.Rename(w.path, final); err != nil {
		return fmt.Errorf("改名失败 (%s → %s): %w", w.path, final, err)
	}
	// 清掉可能残留的伴随文件（旧库为 WAL 模式时）
	for _, s := range []string{"-wal", "-shm"} {
		_ = os.Remove(final + s)
	}
	return nil
}

// 时间格式化只用一处，方便以后换形态。
func stampTime(t time.Time) string { return t.Format(time.RFC3339) }
