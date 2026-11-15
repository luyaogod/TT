// 参数绑定与批量执行：插件通用会话的写路径。
//
// 占位符统一由插件写 `?`，按方言改写成驱动认得的形态（Oracle `:1..:n`、
// Kingbase `$1..$n`）；参数值由驱动绑定，不做字符串拼 SQL。
//
// 批量 = 一条连接上的多条语句。atomic=true 时整批一个事务：任一条失败全部回滚，
// 并返回失败语句的下标（0 基）。

package erpdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Stmt 一条带参数的语句。
type Stmt struct {
	SQL    string
	Params []string
}

// RewritePlaceholders 把 SQL 里的 `?` 换成方言占位符。
//
// 只替换**字符串/引号标识符/注释之外**的 `?`；`?` 出现在字面量里时原样保留。
func RewritePlaceholders(sql, dialect string) (string, int, error) {
	var b strings.Builder
	b.Grow(len(sql) + 8)
	count := 0
	inSingle, inDouble, inLine, inBlock := false, false, false, false
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		if inLine {
			b.WriteByte(c)
			if c == '\n' {
				inLine = false
			}
			continue
		}
		if inBlock {
			b.WriteByte(c)
			if c == '*' && i+1 < len(sql) && sql[i+1] == '/' {
				b.WriteByte('/')
				i++
				inBlock = false
			}
			continue
		}
		if inSingle {
			b.WriteByte(c)
			if c == '\'' {
				if i+1 < len(sql) && sql[i+1] == '\'' { // 转义的 ''''
					b.WriteByte('\'')
					i++
				} else {
					inSingle = false
				}
			}
			continue
		}
		if inDouble {
			b.WriteByte(c)
			if c == '"' {
				inDouble = false
			}
			continue
		}
		switch c {
		case '\'':
			inSingle = true
			b.WriteByte(c)
		case '"':
			inDouble = true
			b.WriteByte(c)
		case '-':
			if i+1 < len(sql) && sql[i+1] == '-' {
				inLine = true
				b.WriteString("--")
				i++
			} else {
				b.WriteByte(c)
			}
		case '/':
			if i+1 < len(sql) && sql[i+1] == '*' {
				inBlock = true
				b.WriteString("/*")
				i++
			} else {
				b.WriteByte(c)
			}
		case '?':
			count++
			switch dialect {
			case "oracle":
				fmt.Fprintf(&b, ":%d", count)
			case "kingbase":
				fmt.Fprintf(&b, "$%d", count)
			default:
				return "", 0, fmt.Errorf("未知方言 %q", dialect)
			}
		default:
			b.WriteByte(c)
		}
	}
	return b.String(), count, nil
}

// bindSQL 改写占位符并返回 []any；占位符个数与参数个数必须一致（早失败，别等驱动报错）。
func bindSQL(sql, dialect string, params []string) (string, []any, error) {
	q, n, err := RewritePlaceholders(sql, dialect)
	if err != nil {
		return "", nil, err
	}
	if n != len(params) {
		return "", nil, fmt.Errorf("占位符 %d 个，参数 %d 个，数量不一致", n, len(params))
	}
	args := make([]any, len(params))
	for i, p := range params {
		args[i] = p
	}
	return q, args, nil
}

// ---- Oracle ----

// QueryArgs 带参查询；只读会话上仍走文本层 + 库侧只读事务。
func (o *OracleConnector) QueryArgs(ctx context.Context, sql string, params []string) ([]string, [][]string, error) {
	q, args, err := bindSQL(sql, "oracle", params)
	if err != nil {
		return nil, nil, err
	}
	if o.readOnly {
		if err := checkReadOnlyStatement(q); err != nil {
			return nil, nil, err
		}
		return o.queryReadOnly(ctx, q, args...)
	}
	return o.queryPlain(ctx, q, args...)
}

// ExecArgs 带参写语句（只读会话拒绝）。
func (o *OracleConnector) ExecArgs(ctx context.Context, sql string, params []string) (int64, error) {
	if o.readOnly {
		return 0, ErrReadOnlySession
	}
	q, args, err := bindSQL(sql, "oracle", params)
	if err != nil {
		return 0, err
	}
	res, err := o.db.ExecContext(ctx, q, args...)
	if err != nil {
		return -1, fmt.Errorf("exec: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil {
		return n, nil
	}
	return -1, nil
}

// ExecBatch 批量执行；atomic=true 时整批一个事务。
func (o *OracleConnector) ExecBatch(ctx context.Context, stmts []Stmt, atomic bool) (int64, int, error) {
	if o.readOnly {
		return 0, -1, ErrReadOnlySession
	}
	if len(stmts) == 0 {
		return 0, -1, nil
	}
	if !atomic {
		var total int64
		for i, s := range stmts {
			n, err := o.ExecArgs(ctx, s.SQL, s.Params)
			if err != nil {
				return total, i, err
			}
			if n > 0 {
				total += n
			}
		}
		return total, -1, nil
	}
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, -1, fmt.Errorf("begin: %w", err)
	}
	var total int64
	for i, s := range stmts {
		q, args, berr := bindSQL(s.SQL, "oracle", s.Params)
		if berr != nil {
			_ = tx.Rollback()
			return 0, i, berr
		}
		res, xerr := tx.ExecContext(ctx, q, args...)
		if xerr != nil {
			_ = tx.Rollback()
			return 0, i, fmt.Errorf("exec[%d]: %w", i, xerr)
		}
		if n, rerr := res.RowsAffected(); rerr == nil && n > 0 {
			total += n
		}
	}
	if cerr := tx.Commit(); cerr != nil {
		return 0, -1, fmt.Errorf("commit: %w", cerr)
	}
	return total, -1, nil
}

// ---- Kingbase ----

// QueryArgs 带参查询；只读会话上仍走文本层检查。
func (k *KingbaseConnector) QueryArgs(ctx context.Context, sql string, params []string) ([]string, [][]string, error) {
	q, args, err := bindSQL(sql, "kingbase", params)
	if err != nil {
		return nil, nil, err
	}
	if k.readOnly {
		if err := checkReadOnlyStatement(q); err != nil {
			return nil, nil, err
		}
	}
	rows, err := k.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()
	return scanPgxRows(rows)
}

// ExecArgs 带参写语句（只读会话拒绝）。
func (k *KingbaseConnector) ExecArgs(ctx context.Context, sql string, params []string) (int64, error) {
	if k.readOnly {
		return 0, ErrReadOnlySession
	}
	q, args, err := bindSQL(sql, "kingbase", params)
	if err != nil {
		return 0, err
	}
	tag, err := k.pool.Exec(ctx, q, args...)
	if err != nil {
		return -1, fmt.Errorf("exec: %w", err)
	}
	return tag.RowsAffected(), nil
}

// ExecBatch 批量执行；atomic=true 时整批一个事务。
func (k *KingbaseConnector) ExecBatch(ctx context.Context, stmts []Stmt, atomic bool) (int64, int, error) {
	if k.readOnly {
		return 0, -1, ErrReadOnlySession
	}
	if len(stmts) == 0 {
		return 0, -1, nil
	}
	if !atomic {
		var total int64
		for i, s := range stmts {
			n, err := k.ExecArgs(ctx, s.SQL, s.Params)
			if err != nil {
				return total, i, err
			}
			if n > 0 {
				total += n
			}
		}
		return total, -1, nil
	}
	tx, err := k.pool.Begin(ctx)
	if err != nil {
		return 0, -1, fmt.Errorf("begin: %w", err)
	}
	var total int64
	for i, s := range stmts {
		q, args, berr := bindSQL(s.SQL, "kingbase", s.Params)
		if berr != nil {
			_ = tx.Rollback(ctx)
			return 0, i, berr
		}
		tag, xerr := tx.Exec(ctx, q, args...)
		if xerr != nil {
			_ = tx.Rollback(ctx)
			return 0, i, fmt.Errorf("exec[%d]: %w", i, xerr)
		}
		if n := tag.RowsAffected(); n > 0 {
			total += n
		}
	}
	if cerr := tx.Commit(ctx); cerr != nil {
		return 0, -1, fmt.Errorf("commit: %w", cerr)
	}
	return total, -1, nil
}

// scanPgxRows 把 pgx 结果集文本化（与 queryReadOnly 同一套 RawValues 逻辑）。
func scanPgxRows(rows pgx.Rows) ([]string, [][]string, error) {
	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, f := range fields {
		columns[i] = f.Name
	}
	var result [][]string
	for rows.Next() {
		raw := rows.RawValues()
		row := make([]string, len(raw))
		for i, v := range raw {
			if v != nil {
				row[i] = string(v)
			}
		}
		result = append(result, row)
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}
	return columns, result, nil
}
