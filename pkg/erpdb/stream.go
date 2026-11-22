package erpdb

// 逐行消费结果集：给"要把整张表搬到本地"的调用方用（开发资料的规格库镜像）。
//
// 为什么不能直接用 Connector.Query：它把整张表读成 [][]string。`DZFK_T` 实测 112 万行、
// `DZBC_T` 97 万行 —— 一次读进内存就是几百 MB 的 Go 切片，再加上逐行 string 的开销，
// 机器先 OOM 再谈落盘。逐行回调之后内存只跟"一批多少行"有关，与表大小无关。
//
// 为什么是**独立接口**而不是给 Connector 加方法：Connector 是导出接口，插件与
// `tools/refplugin` 那边也有实现；往里加方法等于要求所有外部实现跟着改（编译期断掉）。
// 用可选接口（type assert）表达之后，不实现它的连接器自动退回 `Query`，语义不变。

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
)

// RowFunc 收到一行（列名一次给出，之后每行一次调用）。返回错误即中止查询。
//
// **空结果集也会先回调一次**：`fn(cols, nil)` —— 列名来自结果集的元数据，与有没有行
// 无关。没有这一条时，"这张表是空的"和"这张表没有列"在下游分不出来（实测踩过：
// 0 行的 DZAS_T 让镜像报"表没有列"）。
//
// 约定与 Query 一致：列值一律文本化，NULL 是空串。
type RowFunc func(cols []string, row []string) error

// Streamer 可选接口：能逐行吐结果集的连接器实现它。
//
// 注意它**不是** Connector 的一部分（见包注释）。调用方写法：
//
//	if s, ok := conn.(erpdb.Streamer); ok {
//	    err = s.QueryStream(ctx, sql, fn)
//	} else {
//	    cols, rows, err := conn.Query(ctx, sql)
//	    ...
//	}
type Streamer interface {
	QueryStream(ctx context.Context, sql string, fn RowFunc) error
}

// ---------- Oracle ----------

// QueryStream 逐行执行只读查询（Oracle 侧，两层防线与 Query 完全一致）。
//
// 与 Query 的差别只有"不攒结果"：边扫边回调。库侧只读事务、文本层校验、连接归还点
// 都保持原样 —— 多一条执行路径不能多一套安全语义。
func (o *OracleConnector) QueryStream(ctx context.Context, sql string, fn RowFunc) error {
	if o.readOnly {
		if err := checkReadOnlyStatement(sql); err != nil {
			return err
		}
		return o.streamReadOnly(ctx, sql, fn)
	}
	rows, err := o.db.QueryContext(ctx, sql)
	if err != nil {
		return wrapQueryErr(ctx, err)
	}
	defer rows.Close()
	return streamSQLRows(ctx, rows, fn)
}

// streamReadOnly 带只读事务的流式执行体（约束见 queryReadOnly 的注释）。
func (o *OracleConnector) streamReadOnly(ctx context.Context, sql string, fn RowFunc) error {
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return wrapQueryErr(ctx, err)
	}
	// 只读事务的终局一律 Rollback；它同时是连接归还点（漏了就是连接泄漏）。
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, "SET TRANSACTION READ ONLY"); err != nil {
		return wrapQueryErr(ctx, err)
	}
	rows, err := tx.QueryContext(ctx, sql)
	if err != nil {
		return wrapQueryErr(ctx, err)
	}
	defer rows.Close()
	return streamSQLRows(ctx, rows, fn)
}

// ---------- 金仓 ----------

// QueryStream 逐行执行只读查询（金仓侧）。
func (k *KingbaseConnector) QueryStream(ctx context.Context, sql string, fn RowFunc) error {
	// 文本层只在只读会话上过（与 Query 同一个口径与同一个理由）
	if k.readOnly {
		if err := checkReadOnlyStatement(sql); err != nil {
			return err
		}
	}
	rows, err := k.pool.Query(ctx, sql)
	if err != nil {
		return wrapQueryErr(ctx, err)
	}
	defer rows.Close()
	return streamPgxRows(ctx, rows, fn)
}

// ---------- 内部 ----------

// wrapQueryErr 把驱动的取消报错还原成 ctx.Err()。
//
// 不还原的话调用方判不出"用户主动停止"（errors.Is(err, context.Canceled)），
// 界面上会把一次正常取消显示成红色失败 —— 与 dbsync 踩过的是同一个坑。
func wrapQueryErr(ctx context.Context, err error) error {
	if cerr := ctx.Err(); cerr != nil {
		return cerr
	}
	return err
}

// streamSQLRows 把 database/sql 结果集逐行文本化后回调。
func streamSQLRows(ctx context.Context, rows *sql.Rows, fn RowFunc) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	// 先给一次列名（row=nil）：空结果集也要让调用方拿到表的形状
	if err := fn(cols, nil); err != nil {
		return err
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return wrapQueryErr(ctx, err)
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			row[i] = scanString(v)
		}
		if err := fn(cols, row); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return wrapQueryErr(ctx, err)
	}
	return nil
}

// streamPgxRows 把金仓结果集逐行文本化后回调。
func streamPgxRows(ctx context.Context, rows pgx.Rows, fn RowFunc) error {
	fields := rows.FieldDescriptions()
	cols := make([]string, len(fields))
	for i, f := range fields {
		cols[i] = f.Name
	}
	// 先给一次列名（row=nil）：空结果集也要让调用方拿到表的形状
	if err := fn(cols, nil); err != nil {
		return err
	}
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		raw := rows.RawValues()
		row := make([]string, len(raw))
		for i, v := range raw {
			if v != nil {
				row[i] = string(v)
			}
		}
		if err := fn(cols, row); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return wrapQueryErr(ctx, err)
	}
	return nil
}
