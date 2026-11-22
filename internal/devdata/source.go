package devdata

// 两个"来源"抽象：规格表（数据库）与文件（SFTP）。
//
// 为什么要接口而不是直接调：mirror/fetch 的正确性全在**顺序与回滚**上（先写临时目录、
// 成功才改名、取消不留半套、超上限跳过）。这些用真库真服务器验证代价太高（要 112 万行
// 的 DZFK_T、要一台测机），而它们恰恰是最容易写错的地方。接口化之后单测能注入
// "第 3 张表查询失败"、"第 5 个文件读一半取消" 这种场景，而生产路径的实现只有一处。

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/host"
	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/pkg/erpdb"
)

// tablePageRows 分页大小。取 2 万：一页 ~2 万行 × 二三十列，几 MB 内存，
// 又不会把往返次数推高到几千次（大表 112 万行 = 56 页）。
const tablePageRows = 20000

// RemoteFile 服务器侧一个文件/目录的属性（相对 $TOP 的路径）。
type RemoteFile struct {
	// Rel 相对 $TOP 的 POSIX 路径（`erp/mdl/code_a05.template`）。
	Rel     string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// TableSource 规格表的来源（生产实现走 erpdb 只读连接；测试注入内存实现）。
type TableSource interface {
	// CountRows 表的行数（含进度与上限判断都要它）。
	CountRows(ctx context.Context, table string) (int64, error)
	// OrderColumns 主键列（分页用的排序键）；不知道就返回 nil。
	//
	// 为什么需要它：Oracle 驱动 **不会流式返回结果集** —— 它先把整张表读进内存再交给
	// database/sql（实测 `DZBB_T` 52 万行时进程工作集涨到 1.1 GB，且几乎不前进）。
	// 所以大表必须**按页查**：每页一条独立 SQL，内存只跟一页有关。分页要有稳定的
	// 排序键，否则翻页会重复/漏行 —— 主键就是那个键。
	OrderColumns(ctx context.Context, table string) ([]string, error)
	// TableColumns 取表的列清单与大字段（CLOB/BLOB）标记；不知道就返回空。
	//
	// 为什么要它：Oracle 的 LOB 列在 go-ora 里是**逐行额外往返**取的。`DZBB_T` 有一列
	// `dzbb098 CLOB`（52 万行），照 `SELECT *` 拉会变成 52 万次额外往返，实测几分钟
	// 都不动。改成在 SQL 里 `DBMS_LOB.SUBSTR(col, 4000, 1)`（服务端就地转成 VARCHAR2）
	// 之后就是普通字符串，往返次数不再跟行数挂钩。
	TableColumns(ctx context.Context, table string) (TableMeta, error)
	// Query 逐行吐一张表（orderBy 非空时按它排序；cols 首次给出，row=nil 是表头通知）。
	Query(ctx context.Context, table, where string, orderBy []string, meta TableMeta, fn erpdb.RowFunc) error
	// Close 释放连接。
	Close()
}

// TableMeta 一张表的列形态（列名按字段序号；LOB 列单独标出来）。
type TableMeta struct {
	Columns []string
	// LOB 列名（小写）→ 是不是 LOB。命中时 SELECT 列表里那一列换成 DBMS_LOB.SUBSTR。
	LOB map[string]bool
}

// Known 是否取到了列清单。取不到时一律退回 `SELECT *`。
func (m TableMeta) Known() bool { return len(m.Columns) > 0 }

// lobCapChars LOB 列每次取多少**字符**（截断上限）。
//
// 实测（2026-11-21 测试环境）：`DZBB_T.dzbb098`（设计点内容）单列合计 **2.63 GB**
// （单行最长 687,979 字符），`DZFK_T.dzfk009` 823 MB。整份材料会变成 GB 级，与本计划
// "几十 MB"的量级假设完全不同；而且这些内容属于"逐行大字段"，与"按程序成套的原材料"
// 不是一回事（详见 plan 的 ADR-07）。
//
// 取 1000 **字符**，写法 `TO_CHAR(SUBSTR(col, 1, 1000))`。两个坑都实测过：
//   - `SUBSTRB`（按字节）在多字节字符集上直接失败：`ORA-22998: CLOB or NCLOB in
//     multibyte character set not supported`；
//   - `SUBSTR` 到 4000 字符再 `TO_CHAR` 会 `ORA-64203: destination buffer too small`
//     （字符集转换后超过 VARCHAR2 的 4000 字节上限）。1000 字符在最坏 4 字节/字符时
//     是 4000 字节，正好不溢出。
//
// 截断**不是静默的**：命中 LOB 的表会在 manifest 的 warnings 里逐表记一句（见 mirror.go），
// `tt gen status` 也会带出来。要看完整内容请直查库（`tt db sql` / `tt dict`）。
const lobCapChars = 1000

// FileSource 服务器文件的来源（生产实现走 host.FS 只读 SFTP；测试注入内存实现）。
type FileSource interface {
	// Top $TOP（写进 manifest.source.topDir）。
	Top() string
	// Stat 看一个相对路径（不存在时返回 os.ErrNotExist 包装的错误）。
	Stat(ctx context.Context, rel string) (RemoteFile, error)
	// ReadFile 读一个文件（max<=0 不限；超限返回 host.ErrFileTooLarge）。
	ReadFile(ctx context.Context, rel string, max int64) ([]byte, error)
	// Walk 递归列出一棵子树里的**文件**（不跟随符号链接）。
	Walk(ctx context.Context, rel string, fn func(RemoteFile) error) error
	// Close 关闭连接。
	Close()
}

// ---------- 数据库实现 ----------

// dbTableSource 用一条只读连接查规格表。
type dbTableSource struct {
	conn erpdb.Connector
}

// CountRows 数一张表。表不存在时返回的错误里带 ORA-00942（调用方据此判 missing）。
func (s *dbTableSource) CountRows(ctx context.Context, table string) (int64, error) {
	if !validIdent(table) {
		return 0, fmt.Errorf("非法表名: %s", table)
	}
	_, rows, err := s.conn.Query(ctx, "SELECT COUNT(*) FROM "+table)
	if err != nil {
		return 0, err
	}
	n, err := parseCount(rows)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Query 逐行吐一张表。
//
// Oracle 走**分页**（见 OrderColumns 的注释）；金仓/其它连接器走单条流式查询
// （pgx 本身就是逐行吐的，内存有界）。
func (s *dbTableSource) Query(ctx context.Context, table, where string, orderBy []string, meta TableMeta, fn erpdb.RowFunc) error {
	if !validIdent(table) {
		return fmt.Errorf("非法表名: %s", table)
	}
	if s.conn.Type() == "oracle" && len(orderBy) > 0 {
		return s.queryPaged(ctx, table, where, orderBy, meta, fn)
	}
	sql := "SELECT " + selectList(meta) + " FROM " + table
	if strings.TrimSpace(where) != "" {
		sql += " WHERE " + where
	}
	return s.stream(ctx, sql, fn)
}

// selectList 生成 SELECT 列表：有 LOB 列就逐列写（LOB 用 DBMS_LOB.SUBSTR 就地转文本），
// 否则退回 `*`。
//
// 只在**知道列清单**时才逐列写。逐列写的好处是 LOB 不再是逐行往返；代价是列顺序
// 依赖数据字典里记的字段序号（DZEB021）—— 那正是设计器的字段顺序。
func selectList(meta TableMeta) string {
	if !meta.Known() || len(meta.LOB) == 0 {
		return "*"
	}
	parts := make([]string, 0, len(meta.Columns))
	for _, c := range meta.Columns {
		q := `"` + strings.ToUpper(c) + `"`
		if meta.LOB[strings.ToLower(c)] {
			// 就地转 VARCHAR2：LOB 列逐行往返是这里唯一的性能陷阱。
			//
			// 为什么是 `TO_CHAR(SUBSTR(...))` 而不是 `DBMS_LOB.SUBSTR`：
			//   1. `DBMS_*` 被只读文本层（internal/safesql）当受限函数拒绝 ——
			//      那是 tt 的安全边界，不能为了这条查询开口子；
			//   2. 先按字符截断再 TO_CHAR，绕开 `ORA-22835` / `ORA-64203`（CLOB 超过 4000
			//      字节时直接 TO_CHAR 会报"缓冲区太小"）。
			parts = append(parts, fmt.Sprintf("TO_CHAR(SUBSTR(%s, 1, %d)) AS %s", q, lobCapChars, q))
			continue
		}
		parts = append(parts, q)
	}
	return strings.Join(parts, ", ")
}

// stream 单条查询逐行回调（先一次表头通知）。连接器没有 Streamer 时退回一次性读。
func (s *dbTableSource) stream(ctx context.Context, sql string, fn erpdb.RowFunc) error {
	if st, ok := s.conn.(erpdb.Streamer); ok {
		return st.QueryStream(ctx, sql, fn)
	}
	cols, rows, err := s.conn.Query(ctx, sql)
	if err != nil {
		return err
	}
	// 与流式路径保持同一个契约：先给一次列名（row=nil），空结果集也一样
	if err := fn(cols, nil); err != nil {
		return err
	}
	for _, r := range rows {
		if err := fn(cols, r); err != nil {
			return err
		}
	}
	return nil
}

// queryPaged 按 ROWNUM 窗口翻页（Oracle）。
//
// SQL 形态是 Oracle 的标准分页写法：
//
//	SELECT * FROM (
//	    SELECT t.*, ROWNUM AS tt_rn FROM (<inner> ORDER BY <pk>) t WHERE ROWNUM <= <hi>
//	) WHERE tt_rn > <lo>
//
// 外层 `*` 会把 `tt_rn` 也带出来，所以在回调里把这一列剥掉（列名固定是 `tt_rn`，
// 不会与 T100 的列名撞车：那边全是 dzaa001 这种）。
//
// **没有 ORDER BY 就不分页**（调用方给 nils 时走单条流式）：无序分页会重复/漏行，
// 那比慢或占内存更糟 —— 静默错数据没人能发现。
func (s *dbTableSource) queryPaged(ctx context.Context, table, where string, orderBy []string, meta TableMeta, fn erpdb.RowFunc) error {
	inner := "SELECT " + selectList(meta) + " FROM " + table
	if strings.TrimSpace(where) != "" {
		inner += " WHERE " + where
	}
	inner += " ORDER BY " + strings.Join(quoteAll(orderBy), ", ")
	headerSent := false
	strip := func(cols []string, row []string) ([]string, []string) {
		if len(cols) == 0 {
			return cols, row
		}
		last := strings.ToLower(cols[len(cols)-1])
		if last != "tt_rn" {
			return cols, row
		}
		cols = cols[:len(cols)-1]
		if row != nil {
			row = row[:len(row)-1]
		}
		return cols, row
	}
	for lo := int64(0); ; lo += tablePageRows {
		hi := lo + tablePageRows
		q := fmt.Sprintf("SELECT * FROM (SELECT t.*, ROWNUM AS tt_rn FROM (%s) t WHERE ROWNUM <= %d) WHERE tt_rn > %d",
			inner, hi, lo)
		got := int64(0)
		err := s.stream(ctx, q, func(cols []string, row []string) error {
			cols, row = strip(cols, row)
			if row == nil {
				if !headerSent {
					headerSent = true
					return fn(cols, nil)
				}
				return nil
			}
			got++
			return fn(cols, row)
		})
		if err != nil {
			return err
		}
		if got < tablePageRows {
			return nil // 最后一页（含空页）
		}
	}
}

// TableColumns 从数据字典的字段档（DZEB_T）取列清单与 LOB 标记。
//
// 取不到就返回空 TableMeta（→ `SELECT *` + 单条流式），不报错：本环境没有字典档时
// 镜像仍应能跑完，只是大表/LOB 表会慢且占内存。
func (s *dbTableSource) TableColumns(ctx context.Context, table string) (TableMeta, error) {
	if !validIdent(table) {
		return TableMeta{}, fmt.Errorf("非法表名: %s", table)
	}
	_, rows, err := s.conn.Query(ctx,
		"SELECT DZEB002, UPPER(DZEB007) FROM DZEB_T WHERE DZEB001 = '"+strings.ToLower(table)+
			"' ORDER BY CAST(DZEB021 AS INTEGER)")
	if err != nil {
		return TableMeta{}, nil
	}
	meta := TableMeta{LOB: map[string]bool{}}
	for _, r := range rows {
		if len(r) < 2 || !validIdent(r[0]) {
			continue
		}
		col := strings.ToLower(r[0])
		meta.Columns = append(meta.Columns, col)
		switch r[1] {
		case "CLOB", "NCLOB", "BLOB", "LONG", "LONG RAW":
			meta.LOB[col] = true
		}
	}
	return meta, nil
}

// OrderColumns 从数据字典的键值档（DZED_T）取主键列。
//
// 取不到就返回 nil（→ 单条流式），不报错：本环境没有 DZED_T 时镜像仍应能跑完，
// 只是大表会占更多内存。
func (s *dbTableSource) OrderColumns(ctx context.Context, table string) ([]string, error) {
	if !validIdent(table) {
		return nil, fmt.Errorf("非法表名: %s", table)
	}
	_, rows, err := s.conn.Query(ctx,
		"SELECT dzed004 FROM DZED_T WHERE dzed001 = '"+strings.ToLower(table)+"' AND dzed003 = 'P'")
	if err != nil {
		return nil, nil // 字典表不可用 = 没有排序键，不算错误
	}
	if len(rows) == 0 {
		return nil, nil
	}
	var cols []string
	for _, f := range strings.Split(rows[0][0], ",") {
		f = strings.Trim(strings.TrimSpace(f), `"`)
		if validIdent(f) {
			cols = append(cols, strings.ToLower(f))
		}
	}
	// 超过 20 列的主键不现实；真遇到就当没有（ORDER BY 太长反而是负担）
	if len(cols) == 0 || len(cols) > 8 {
		return nil, nil
	}
	return cols, nil
}

// quoteAll 给排序键加引号并转大写。
//
// **必须大写**：Oracle 的标识符默认按大写解释，加引号的 `"dzaa009"` 会被当成另一个
// 名字（实测：`ORA-00904: "dzaa009": invalid identifier`，整次 mirror 失败）。
// 分页只对 Oracle 生效（金仓走单条流式、不看 orderBy），所以这里一律按 Oracle 的口径写。
func quoteAll(cols []string) []string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = `"` + strings.ToUpper(c) + `"`
	}
	return out
}

func (s *dbTableSource) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

// parseCount 从 `SELECT COUNT(*)` 的结果里取数。
func parseCount(rows [][]string) (int64, error) {
	if len(rows) == 0 || len(rows[0]) == 0 {
		return 0, fmt.Errorf("COUNT(*) 没有返回结果")
	}
	var n int64
	if _, err := fmt.Sscanf(strings.TrimSpace(rows[0][0]), "%d", &n); err != nil {
		return 0, fmt.Errorf("COUNT(*) 返回值不是数字: %q", rows[0][0])
	}
	return n, nil
}

// OnlineProgramVersion 在线取某程序当前的有效版次（`MAX(DZAA002) WHERE DZAASTUS='Y'`）。
//
// identity="c"（客制材料）时优先取 `DZAA009='c'` 的版次，取不到再退回不限身份：
// 客制版次通常停在签出那一刻，而标准侧会继续升版 —— 不限身份的 MAX 会把标准
// 的新版次记到客制材料头上（见 fetch.go 的 materialIdentity）。
//
// 这是本包里唯一"查库取一个值"的地方，也是「查询默认在线」这条口径的落点：
// 版本锚点**先问在线库**，只有在线取不到时才看本地共用资料（见 fetch.go 的
// programVersionAnchor）。理由：镜像暂不启用（材料范围等 model 定型后再定），
// 而"这个程序现在第几版"这件事本来就该以库为准。
//
// 程序名已过白名单校验，内联进 SQL 是安全的（与 dict 侧同样的做法）。
func OnlineProgramVersion(ctx context.Context, e *config.NamedSsh, ent int, prog, identity string) (int, error) {
	if err := ValidProgName(prog); err != nil {
		return 0, err
	}
	src, err := OpenTableSource(ctx, e, ent)
	if err != nil {
		return 0, err
	}
	defer src.Close()
	db, ok := src.(*dbTableSource)
	if !ok {
		return 0, fmt.Errorf("取版次失败：连接器不可用")
	}
	_, rows, err := db.conn.Query(ctx, fmt.Sprintf(
		"SELECT COALESCE(MAX(CASE WHEN DZAA009 = 'c' THEN CAST(DZAA002 AS DECIMAL(10,0)) END),0),"+
			" COALESCE(MAX(CAST(DZAA002 AS DECIMAL(10,0))),0)"+
			" FROM DZAA_T WHERE DZAA001 = '%s' AND DZAASTUS = 'Y'",
		strings.TrimSpace(prog)))
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 || len(rows[0]) == 0 {
		return 0, nil
	}
	cust := parseVersionCell(rows[0], 0)
	anyVer := parseVersionCell(rows[0], 1)
	if identity == "c" && cust > 0 {
		return cust, nil
	}
	return anyVer, nil
}

// parseVersionCell 取聚合结果里第 i 列并转成整数；空值/NULL 返回 0。
func parseVersionCell(row []string, i int) int {
	if i >= len(row) {
		return 0
	}
	s := strings.TrimSpace(row[i])
	if s == "" {
		return 0
	}
	var v int
	if _, err := fmt.Sscanf(s, "%d", &v); err != nil {
		return 0
	}
	return v
}

// OpenTableSource 按环境配置打开规格表来源（客户端直连，支持 viaSsh 隧道）。
//
// 与 `tt dict db sync` 走的是同一条链路（connsvc.OpenBuiltin）：只读事务、隧道编排、
// 账号解析都只有一份实现。不同之处只有"读哪张表"。
func OpenTableSource(ctx context.Context, e *config.NamedSsh, ent int) (TableSource, error) {
	if e == nil || e.DB == nil {
		return nil, fmt.Errorf("环境 %q 未配置数据库，无法镜规格表", envNameOf(e))
	}
	c := *e.DB
	if ent > 0 {
		acct := ""
		for _, it := range e.Ents {
			if it.Ent == ent {
				acct = strings.TrimSpace(it.Account)
				break
			}
		}
		if acct == "" {
			return nil, fmt.Errorf("企业 %d 不在环境 %q 的企业清单里（先用 tt env ents --refresh 同步）", ent, e.Name)
		}
		c.Account = acct
	}
	if err := c.FillDialCred(); err != nil {
		return nil, err
	}
	conn, err := openReadOnly(ctx, c)
	if err != nil {
		return nil, err
	}
	return &dbTableSource{conn: conn}, nil
}

// openReadOnly 抽出成变量是为了让单测能替换（生产路径就是 connsvc.OpenBuiltin）。
var openReadOnly = func(ctx context.Context, c dbconfig.Connection) (erpdb.Connector, error) {
	conn, release, err := connsvc.OpenBuiltin(ctx, c)
	if err != nil {
		return nil, err
	}
	// 连接释放由 wrapper 负责：调用方只认 Close()，不需要知道隧道。
	return &releasedConn{Connector: conn, release: release}, nil
}

// releasedConn 让"关连接"与"关隧道"合成一次 Close（漏一个就是端口泄漏）。
type releasedConn struct {
	erpdb.Connector
	release func()
}

func (r *releasedConn) Close() {
	if r.Connector != nil {
		r.Connector.Close()
	}
	if r.release != nil {
		r.release()
	}
}

// ---------- SFTP 实现 ----------

// sftpSource 用一条只读 SFTP 通道读文件（`tt file` 的同一套原语）。
type sftpSource struct {
	fs  *host.FS
	top string
}

// Top 返回 $TOP。
func (s *sftpSource) Top() string { return s.top }

// Stat 看一个相对路径；不存在时把错误统一成 fs.ErrNotExist。
func (s *sftpSource) Stat(_ context.Context, rel string) (RemoteFile, error) {
	rel = strings.TrimPrefix(path.Clean("/"+rel), "/")
	ent, err := s.fs.Stat(path.Join(s.top, rel))
	if err != nil {
		if host.NotFound(err) {
			return RemoteFile{Rel: rel}, fmt.Errorf("%w: %s", fs.ErrNotExist, rel)
		}
		return RemoteFile{Rel: rel}, host.WrapPathErr(rel, err)
	}
	return RemoteFile{Rel: rel, Size: ent.Size, ModTime: ent.ModTime, IsDir: ent.IsDir()}, nil
}

// ReadFile 读一个相对路径的文件。
func (s *sftpSource) ReadFile(_ context.Context, rel string, max int64) ([]byte, error) {
	rel = strings.TrimPrefix(path.Clean("/"+rel), "/")
	f, err := s.fs.Open(path.Join(s.top, rel))
	if err != nil {
		if host.NotFound(err) {
			return nil, fmt.Errorf("%w: %s", fs.ErrNotExist, rel)
		}
		return nil, host.WrapPathErr(rel, err)
	}
	defer f.Close()
	if max > 0 {
		// 先看属性：比"读完才发现超了"省一次传输（.42f 可以到几十 MB）
		if e, serr := s.Stat(context.Background(), rel); serr == nil && e.Size > max {
			return nil, host.ErrFileTooLarge
		}
	}
	b, truncated, err := readAllLimit(f, max)
	if err != nil {
		return nil, host.WrapPathErr(rel, err)
	}
	if truncated {
		return nil, host.ErrFileTooLarge
	}
	return b, nil
}

// Walk 递归列文件（目录不回调；符号链接不跟随）。
func (s *sftpSource) Walk(_ context.Context, rel string, fn func(RemoteFile) error) error {
	rel = strings.TrimPrefix(path.Clean("/"+rel), "/")
	return s.walk(path.Join(s.top, rel), rel, fn)
}

func (s *sftpSource) walk(absDir, relDir string, fn func(RemoteFile) error) error {
	entries, err := s.fs.List(absDir)
	if err != nil {
		if host.NotFound(err) {
			return fmt.Errorf("%w: %s", fs.ErrNotExist, relDir)
		}
		return host.WrapPathErr(relDir, err)
	}
	for _, e := range entries {
		rel := path.Join(relDir, e.Name)
		if e.IsSymlink() {
			// 不跟随：T100 的 mdl 目录里没有链接，而跟随链接会让"材料目录"越出 $TOP
			continue
		}
		if e.IsDir() {
			if err := s.walk(e.Path, rel, fn); err != nil {
				return err
			}
			continue
		}
		if err := fn(RemoteFile{Rel: rel, Size: e.Size, ModTime: e.ModTime}); err != nil {
			return err
		}
	}
	return nil
}

func (s *sftpSource) Close() {
	if s.fs != nil {
		_ = s.fs.Close()
	}
}

// OpenFileSource 连 SSH、探测 $TOP、打开只读 SFTP。
//
// $TOP **不缓存**（与 `tt file` 的 30 分钟缓存不同）：一份材料的 source.topDir 要能
// 事后复核，用几天前的缓存值会让"材料来自哪台机的哪个目录"这件事说不准。
// 代价是每次多一次登录探测（2–5 秒），而 mirror/fetch 本来就是一分钟级的动作。
func OpenFileSource(e *config.NamedSsh) (FileSource, func(), error) {
	if e == nil || e.Host == "" {
		return nil, nil, fmt.Errorf("环境 %s 没有配置 SSH 主机", envNameOf(e))
	}
	if strings.TrimSpace(e.Zone) == "" {
		return nil, nil, fmt.Errorf("环境 %s 没有配置登录区域（zone），无法探测 $TOP", e.Name)
	}
	conn, err := host.Dial(e.SSHConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("连接 %s 失败: %w", e.SSHConfig.Addr(), err)
	}
	rt, err := host.ProbeTEnv(conn, e.Zone)
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("登录探测 T100 目录失败（环境 %s，zone %s）: %w", e.Name, e.Zone, err)
	}
	if rt == nil || strings.TrimSpace(rt.TOP) == "" {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("登录探测没有回读到 $TOP（环境 %s，zone %s）", e.Name, e.Zone)
	}
	fsys, err := conn.FS()
	if err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("打开 SFTP 通道失败: %w", err)
	}
	top := strings.TrimSuffix(strings.TrimSpace(rt.TOP), "/")
	src := &sftpSource{fs: fsys, top: top}
	closeAll := func() {
		_ = fsys.Close()
		_ = conn.Close()
	}
	return src, closeAll, nil
}

// envNameOf 给错误文案取一个名字（nil 安全）。
func envNameOf(e *config.NamedSsh) string {
	if e == nil {
		return "(未指定)"
	}
	return e.Name
}

// readAllLimit 读满一个 reader，超过 limit 时返回 truncated=true。
func readAllLimit(r io.Reader, limit int64) ([]byte, bool, error) {
	if limit <= 0 {
		b, err := io.ReadAll(r)
		return b, false, err
	}
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return b, false, err
	}
	if int64(len(b)) > limit {
		return b[:limit], true, nil
	}
	return b, false, nil
}

// writeLocalFile 把一段字节原样写到本地（自动建父目录，先 .tmp 再 rename）。
//
// **不做任何转换**（ADR-05）：不 TrimSpace、不补换行、不转编码 ——
// `.tab` 里的 `&apos;`、`CRLF`、末尾空行都按字节保留。
func writeLocalFile(p string, data []byte) error {
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败 (%s): %w", dir, err)
	}
	tmp := p + ".part"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("写文件失败 (%s): %w", tmp, err)
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("写文件失败 (%s): %w", p, err)
	}
	return nil
}
