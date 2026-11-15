// 会话层：连接建立与**只读策略**在这里显式分开。
//
// 为什么需要这一层（而不是让所有调用方继续用 Open）：
//
//   - 内置只读路径（`tt db sql`、字典直查、dbsync 校验）必须在**文本层 + 库侧**双防线
//     之内跑，`Open` 就是那条路径的入口，语义不变；
//   - 插件通用连接（Host API 的 DBSession）是另一回事：TT 不审查插件用连接做什么业务，
//     数据库账号自身权限就是边界。它**不能**悄悄继承内置只读策略，否则 capabilities 里
//     写着能写却每条语句都被拒；反过来也**不能**让插件路径的存在削弱内置路径的两层防线。
//
// 于是 `OpenSession` 要一个显式 `ReadOnly` 选择，两条路各自成立、互不冒充。
package erpdb

import (
	"context"
	"errors"
	"fmt"

	"tt/internal/dbconfig"
)

// SessionOptions 打开连接时的策略选择。
type SessionOptions struct {
	// ReadOnly 建立**库侧只读**会话，并在 Query 上加文本层校验（内置只读语义）。
	//
	// false 时两层都不加 —— 这是插件通用会话的路径，边界是数据库账号权限。
	// 零值 false 是**故意的**：将来新增调用方必须自己写清楚要哪种，
	// 漏写时得到的是"普通会话"（与驱动默认一致），而不是偷偷变成只读或者偷偷变成可写。
	ReadOnly bool
}

// Session 一条可执行的数据库会话。
//
// 与 Connector 的关系：Connector 是历史上"只读查询"的最小面（大批调用方依赖它），
// 保留不动；Session 是它 + 写能力 + 策略自述。新代码用 Session。
type Session interface {
	Connector
	// Exec 执行一条写语句，返回受影响行数；驱动给不出时返回 -1。
	// 只读会话上调用它一律报错（不是静默变成 SELECT）。
	Exec(ctx context.Context, sql string) (int64, error)
	// ReadOnly 本会话是否库侧只读。
	ReadOnly() bool
}

// ErrReadOnlySession 只读会话上试图执行写语句。
var ErrReadOnlySession = errors.New("只读会话不允许执行写语句（该会话以 ReadOnly 建立）")

// OpenSession 按类型打开一条 Session。
func OpenSession(ctx context.Context, c dbconfig.Connection, opt SessionOptions) (Session, error) {
	if err := c.FillDialCred(); err != nil {
		return nil, err
	}
	switch c.Type {
	case "kingbase":
		return openKingbase(ctx, c, opt.ReadOnly)
	case "oracle":
		return openOracle(ctx, c, opt.ReadOnly)
	default:
		return nil, fmt.Errorf("未知的连接类型 \"%s\"", c.Type)
	}
}

// ---- Kingbase ----

// Exec 只在非只读会话上可用。
func (k *KingbaseConnector) Exec(ctx context.Context, sql string) (int64, error) {
	if k.readOnly {
		return 0, ErrReadOnlySession
	}
	tag, err := k.pool.Exec(ctx, sql)
	if err != nil {
		return -1, fmt.Errorf("exec: %w", err)
	}
	return tag.RowsAffected(), nil
}

// ReadOnly 本会话是否库侧只读。
func (k *KingbaseConnector) ReadOnly() bool { return k.readOnly }

// ---- Oracle ----

// Exec 只在非只读会话上可用。
func (o *OracleConnector) Exec(ctx context.Context, sql string) (int64, error) {
	if o.readOnly {
		return 0, ErrReadOnlySession
	}
	res, err := o.db.ExecContext(ctx, sql)
	if err != nil {
		return -1, fmt.Errorf("exec: %w", err)
	}
	if n, err := res.RowsAffected(); err == nil {
		return n, nil
	}
	return -1, nil
}

// ReadOnly 本会话是否库侧只读。
func (o *OracleConnector) ReadOnly() bool { return o.readOnly }
