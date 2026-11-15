// Package live 提供"远程 ERP 库直查"的数据访问层。
// *Live 实现 db.Source:与本地 SQLite 镜像(db 包)同语义 —— 同一套字典表
// (dzea_t 等 24 张)与 JOIN SQL,直接连远程库执行,rt/rv/desc/scc/rq 的数据源
// 切到某环境时即经本包(金仓 + Oracle 双方言,见 source.go)。
//
// 设计要点:
//   - erpdb.Connector 走简单协议(simple protocol),不支持绑定参数 → 所有值经
//     QuoteLit/ValidIdent 白名单后内联,仍保持单条只读 SELECT;
//   - viaSsh 隧道:客户端不可达 DB 时先建 SSH 端口转发,连接 host 换成 127.0.0.1:本地端口。
package live

import (
	"context"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/pkg/erpdb"
)

// Live 一个远程数据源:底层连接器 + 由 Core 连接服务给出的释放函数(含 SSH 隧道)。
type Live struct {
	conn    erpdb.Connector
	release func()
}

// Dialect 返回底层连接方言("kingbase"|"oracle")。
func (l *Live) Dialect() string { return l.conn.Type() }

// Connector 暴露底层连接器(SelectAllSQL 等通用能力)。
func (l *Live) Connector() erpdb.Connector { return l.conn }

// Close 释放连接;若开了 SSH 隧道一并关闭。
func (l *Live) Close() {
	if l.release != nil {
		l.release()
	}
}

// Open 按连接配置建立远程数据源。若配置了 viaSsh,先起 SSH 隧道再连本地转发端口。
//
// 隧道编排与 erpdb 的只读语义都在 connsvc.OpenBuiltin 一处(与 dbsync、设置页的
// 测试连接同一条链路),所以"测试通了、直查不通"这种分叉不会从本包长出来。
func Open(ctx context.Context, c dbconfig.Connection) (*Live, error) {
	conn, release, err := connsvc.OpenBuiltin(ctx, c)
	if err != nil {
		return nil, err
	}
	return &Live{conn: conn, release: release}, nil
}
