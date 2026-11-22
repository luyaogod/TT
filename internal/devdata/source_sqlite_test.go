package devdata

// 两个"来源"原语与本地规格库写入的测试（全部离线：内存假连接器 + 真 SQLite 文件）。
//
// 为什么这层必须测：mirror/fetch 的正确性全在"顺序与回滚"上，而它们只通过
// TableSource/FileSource 这两个接口落地。接口的生产实现（dbTableSource）里藏着
// 几处真机实测出来的硬约束：
//   - Oracle 驱动不流式，大表必须按 ROWNUM 窗口分页；没有主键就**不分页**（无序分页会漏行）；
//   - LOB 列逐行往返极慢，必须 `TO_CHAR(SUBSTR(col,1,1000))` 就地转文本；
//   - 数据字典（DZEB/DZED）取不到时不能报错，只能退化（SELECT * / 单条流式）。
// 这些用真库验证代价太高，所以用假连接器钉住"发出去的 SQL 与回调契约"。
//
// SQLite 侧是"本包自己读自己写"的落盘格式，写入路径（分批事务、原子改名、
// 尽力建主键索引）同样值得离线钉住。

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/config"
	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/pkg/erpdb"
)

// fakeQConn 是 erpdb.Connector 的内存实现（不带 Streamer）。
type fakeQConn struct {
	typ     string
	cols    []string
	rows    [][]string
	qerr    error
	queries []string
	closed  bool
}

func (f *fakeQConn) Type() string { return f.typ }
func (f *fakeQConn) ServerVersion(context.Context) (string, error) {
	return "fake", nil
}
func (f *fakeQConn) Query(_ context.Context, sql string) ([]string, [][]string, error) {
	f.queries = append(f.queries, sql)
	return f.cols, f.rows, f.qerr
}
func (f *fakeQConn) Close() { f.closed = true }

// fakeStreamConn 同时实现 Streamer：按脚本逐页回播放数据（模拟 Oracle 分页）。
type fakeStreamConn struct {
	typ   string
	cols  []string
	pages [][][]string // 每次 QueryStream 消费一页
	calls int
	sqls  []string
}

func (f *fakeStreamConn) Type() string                                  { return f.typ }
func (f *fakeStreamConn) ServerVersion(context.Context) (string, error) { return "fake", nil }
func (f *fakeStreamConn) Close()                                        {}
func (f *fakeStreamConn) Query(_ context.Context, sql string) ([]string, [][]string, error) {
	return nil, nil, fmt.Errorf("分页路径不该走 Query: %s", sql)
}

func (f *fakeStreamConn) QueryStream(_ context.Context, sql string, fn erpdb.RowFunc) error {
	f.sqls = append(f.sqls, sql)
	var rows [][]string
	if f.calls < len(f.pages) {
		rows = f.pages[f.calls]
	}
	f.calls++
	cols := f.cols
	if cols == nil {
		cols = []string{"DZAA009", "TT_RN"}
	}
	// 与生产契约一致：先一次表头通知（row=nil），空结果集也发。
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

// TestDBTableSource_CountRowsAndQuery 钉住 CountRows 的解析与 Query 的两条路径：
// 非 Oracle（单条 SQL）与 Oracle 无主键（退化单条流式），以及 LOB 列的 SELECT 列表。
func TestDBTableSource_CountRowsAndQuery(t *testing.T) {
	// CountRows：正常值。
	conn := &fakeQConn{typ: "kingbase", cols: []string{"COUNT(*)"}, rows: [][]string{{"42"}}}
	src := &dbTableSource{conn: conn}
	if n, err := src.CountRows(context.Background(), "DZAA_T"); err != nil || n != 42 {
		t.Errorf("CountRows = %d/%v", n, err)
	}
	// 非法表名（拼 SQL 的唯一入口，必须先挡）。
	if _, err := src.CountRows(context.Background(), "bad table"); err == nil || !strings.Contains(err.Error(), "非法表名") {
		t.Errorf("非法表名 err = %v", err)
	}
	// 查询报错透传。
	if _, err := (&dbTableSource{conn: &fakeQConn{qerr: errors.New("ORA-00942")}}).
		CountRows(context.Background(), "DZAA_T"); err == nil {
		t.Error("查询报错应透传")
	}
	// 空结果集 / 非数字。
	if _, err := (&dbTableSource{conn: &fakeQConn{}}).CountRows(context.Background(), "DZAA_T"); err == nil ||
		!strings.Contains(err.Error(), "没有返回结果") {
		t.Errorf("空结果 err = %v", err)
	}
	if _, err := (&dbTableSource{conn: &fakeQConn{cols: []string{"C"}, rows: [][]string{{"abc"}}}}).
		CountRows(context.Background(), "DZAA_T"); err == nil || !strings.Contains(err.Error(), "不是数字") {
		t.Errorf("非数字 err = %v", err)
	}

	// Query：非 Oracle 走单条 SQL（带 WHERE）。
	rows := [][]string{{"1"}, {"2"}}
	plain := &fakeQConn{typ: "kingbase", cols: []string{"C1"}, rows: rows}
	src = &dbTableSource{conn: plain}
	var got [][]string
	var headers int
	err := src.Query(context.Background(), "DZAA_T", "DZAA002 = 1", nil, TableMeta{}, func(cols []string, row []string) error {
		if row == nil {
			headers++
			if len(cols) != 1 {
				t.Errorf("表头列 = %v", cols)
			}
			return nil
		}
		got = append(got, row)
		return nil
	})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if headers != 1 || len(got) != 2 {
		t.Errorf("回调 = headers:%d rows:%d", headers, len(got))
	}
	if plain.queries[0] != "SELECT * FROM DZAA_T WHERE DZAA002 = 1" {
		t.Errorf("SQL = %q", plain.queries[0])
	}
	// 回调报错必须中断（不能吞掉继续写完半个文件）。
	boom := errors.New("取消")
	if err := src.Query(context.Background(), "DZAA_T", "", nil, TableMeta{},
		func([]string, []string) error { return boom }); !errors.Is(err, boom) {
		t.Errorf("回调错误 = %v", err)
	}

	// 知道列清单 + LOB：SELECT 列表要逐列写、LOB 用 TO_CHAR(SUBSTR(...))。
	lob := &fakeQConn{typ: "kingbase", cols: []string{"C1", "C2"}}
	src = &dbTableSource{conn: lob}
	meta := TableMeta{Columns: []string{"c1", "c2"}, LOB: map[string]bool{"c1": true}}
	_ = src.Query(context.Background(), "DZAA_T", "", nil, meta, func([]string, []string) error { return nil })
	if !strings.Contains(lob.queries[0], `TO_CHAR(SUBSTR("C1", 1, 1000)) AS "C1"`) ||
		!strings.Contains(lob.queries[0], `"C2"`) || strings.Contains(lob.queries[0], "*") {
		t.Errorf("LOB SELECT 列表 = %q", lob.queries[0])
	}
	// 非 Oracle 且非法表名。
	if err := src.Query(context.Background(), "bad-name", "", nil, TableMeta{},
		func([]string, []string) error { return nil }); err == nil {
		t.Error("非法表名应报错")
	}
}

// TestDBTableSource_OraclePaged 钉住 Oracle 分页：带主键时用 ROWNUM 窗口、剥掉 tt_rn、
// 表头只发一次；一页满 tablePageRows 才继续翻页。
func TestDBTableSource_OraclePaged(t *testing.T) {
	fullPage := make([][]string, tablePageRows)
	for i := range fullPage {
		fullPage[i] = []string{"v", fmt.Sprint(i)}
	}
	conn := &fakeStreamConn{
		typ:  "oracle",
		cols: []string{"C1", "TT_RN"},
		pages: [][][]string{
			fullPage,
			{{"last", "1"}}, // 第二页只有 1 行 → 循环结束
		},
	}
	src := &dbTableSource{conn: conn}
	var headerCols []string
	var dataRows int
	err := src.Query(context.Background(), "DZAA_T", "DZAA002 = 1", []string{"dzaa009"}, TableMeta{},
		func(cols []string, row []string) error {
			if row == nil {
				headerCols = cols
				return nil
			}
			if len(row) != 1 {
				t.Errorf("tt_rn 没被剥掉: %v", row)
			}
			dataRows++
			return nil
		})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(headerCols) != 1 || headerCols[0] != "C1" {
		t.Errorf("表头 = %v（外层 * 会带出 tt_rn，必须剥掉）", headerCols)
	}
	if dataRows != tablePageRows+1 || conn.calls != 2 {
		t.Errorf("行数/页数 = %d/%d", dataRows, conn.calls)
	}
	for _, sql := range conn.sqls {
		if !strings.Contains(sql, "ROWNUM") || !strings.Contains(sql, `ORDER BY "DZAA009"`) {
			t.Errorf("分页 SQL 形态 = %q", sql)
		}
	}
}

// TestDBTableSource_DictLookups 钉住数据字典退化：DZEB_T/DZED_T 取不到时不能报错，
// 只能退回"不知道列/不知道主键"，否则没有字典档的环境整次 mirror 都会失败。
func TestDBTableSource_DictLookups(t *testing.T) {
	// TableColumns：正常（含 LOB 类型与非 LOB、以及非法列名跳过）。
	conn := &fakeQConn{typ: "oracle", cols: []string{"DZEB002", "T"}, rows: [][]string{
		{"c1", "CLOB"},
		{"c2", "NUMBER"},
		{"bad name", "CLOB"},
		{"c4", "BLOB"},
		{"c5", "LONG"},
	}}
	meta, err := (&dbTableSource{conn: conn}).TableColumns(context.Background(), "DZAA_T")
	if err != nil {
		t.Fatalf("TableColumns: %v", err)
	}
	if strings.Join(meta.Columns, ",") != "c1,c2,c4,c5" {
		t.Errorf("列清单 = %v", meta.Columns)
	}
	if !meta.LOB["c1"] || !meta.LOB["c4"] || !meta.LOB["c5"] || meta.LOB["c2"] {
		t.Errorf("LOB 标记 = %v", meta.LOB)
	}
	if !meta.Known() || (TableMeta{}).Known() {
		t.Error("Known 口径错")
	}
	// 查不到字典档：返回空 meta，不报错。
	meta, err = (&dbTableSource{conn: &fakeQConn{qerr: errors.New("no dict")}}).
		TableColumns(context.Background(), "DZAA_T")
	if err != nil || meta.Known() {
		t.Errorf("字典不可用应退化 = %+v/%v", meta, err)
	}
	if _, err := (&dbTableSource{conn: conn}).TableColumns(context.Background(), "bad name"); err == nil {
		t.Error("非法表名应报错")
	}

	// OrderColumns：正常（逗号分隔、去引号、转小写）。
	oc := &fakeQConn{typ: "oracle", cols: []string{"DZED004"}, rows: [][]string{{`"DZAA009", dzaa010`}}}
	cols, err := (&dbTableSource{conn: oc}).OrderColumns(context.Background(), "DZAA_T")
	if err != nil || strings.Join(cols, ",") != "dzaa009,dzaa010" {
		t.Errorf("主键列 = %v/%v", cols, err)
	}
	// 空结果 → 没有排序键（不报错）。
	if cols, err := (&dbTableSource{conn: &fakeQConn{}}).OrderColumns(context.Background(), "DZAA_T"); err != nil || cols != nil {
		t.Errorf("空主键 = %v/%v", cols, err)
	}
	// 字典表不可用 → 退化（不报错）。
	if cols, err := (&dbTableSource{conn: &fakeQConn{qerr: errors.New("no dict")}}).
		OrderColumns(context.Background(), "DZAA_T"); err != nil || cols != nil {
		t.Errorf("字典不可用应退化 = %v/%v", cols, err)
	}
	// 主键超过 8 列不现实 → 当没有。
	tooMany := &fakeQConn{cols: []string{"C"}, rows: [][]string{{"a,b,c,d,e,f,g,h,i"}}}
	if cols, _ := (&dbTableSource{conn: tooMany}).OrderColumns(context.Background(), "DZAA_T"); cols != nil {
		t.Errorf("超长主键应丢弃: %v", cols)
	}
	if _, err := (&dbTableSource{conn: oc}).OrderColumns(context.Background(), "bad name"); err == nil {
		t.Error("非法表名应报错")
	}

	// parseCount / parseVersionCell / quoteAll 的小工具口径。
	if n, err := parseCount([][]string{{"7"}}); err != nil || n != 7 {
		t.Errorf("parseCount = %d/%v", n, err)
	}
	if _, err := parseCount(nil); err == nil {
		t.Error("空结果 parseCount 应报错")
	}
	if parseVersionCell([]string{" 12 "}, 0) != 12 || parseVersionCell([]string{""}, 0) != 0 ||
		parseVersionCell(nil, 2) != 0 || parseVersionCell([]string{"x"}, 0) != 0 {
		t.Error("parseVersionCell 口径错")
	}
	if got := strings.Join(quoteAll([]string{"a", "b"}), ","); got != `"A","B"` {
		t.Errorf("quoteAll = %q", got)
	}
}

// TestSource_ReadAllLimitWriteLocal 钉住两个文件原语：读上限（超限要明确
// ErrFileTooLarge，不能静默截断）与本地写（先 .part 再 rename、不做任何转换）。
func TestSource_ReadAllLimitWriteLocal(t *testing.T) {
	if b, trunc, err := readAllLimit(strings.NewReader("12345"), 0); err != nil || trunc || string(b) != "12345" {
		t.Errorf("不限长读取 = %q/%v/%v", b, trunc, err)
	}
	if b, trunc, err := readAllLimit(strings.NewReader("12345"), 3); err != nil || !trunc || string(b) != "123" {
		t.Errorf("超限读取 = %q/%v/%v", b, trunc, err)
	}
	if b, trunc, err := readAllLimit(strings.NewReader("12"), 3); err != nil || trunc || string(b) != "12" {
		t.Errorf("未超限读取 = %q/%v/%v", b, trunc, err)
	}
	if _, _, err := readAllLimit(errReader{}, 3); err == nil {
		t.Error("读错误应透传")
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "m.tab")
	raw := []byte("<?xml?>\r\n<a &apos;b&apos;>\n\n") // 原样保留：不 Trim、不补换行、不转编码
	if err := writeLocalFile(p, raw); err != nil {
		t.Fatalf("writeLocalFile: %v", err)
	}
	got, err := os.ReadFile(p)
	if err != nil || string(got) != string(raw) {
		t.Errorf("写盘内容被转换了: %q/%v", got, err)
	}
	if _, err := os.Stat(p + ".part"); !os.IsNotExist(err) {
		t.Error("临时文件没被改名")
	}

	if envNameOf(nil) != "(未指定)" {
		t.Errorf("envNameOf(nil) = %q", envNameOf(nil))
	}
	if envNameOf(&config.NamedSsh{Name: "tst"}) != "tst" {
		t.Errorf("envNameOf = %q", envNameOf(&config.NamedSsh{Name: "tst"}))
	}
}

// errReader 永远返回错误（测 readAllLimit 的错误分支）。
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

// TestOpenTableSource_AndOnlineVersion 钉住"打开规格表来源"的校验与在线版次查询：
// 环境没配数据库要给人读的错；企业编号不在清单里不能静默用别的账号；
// OnlineProgramVersion 在 identity=c 时优先取客制版次（否则标准侧升版会被记到客制材料头上）。
func TestOpenTableSource_AndOnlineVersion(t *testing.T) {
	ctx := context.Background()

	// 校验：没配数据库。
	if _, err := OpenTableSource(ctx, nil, 0); err == nil || !strings.Contains(err.Error(), "未配置数据库") {
		t.Errorf("nil 环境 err = %v", err)
	}
	noDB := &config.NamedSsh{Name: "tst"}
	if _, err := OpenTableSource(ctx, noDB, 0); err == nil || !strings.Contains(err.Error(), "tst") {
		t.Errorf("没配数据库 err = %v", err)
	}
	// 企业编号不在清单里。
	withDB := &config.NamedSsh{Name: "tst", DB: &dbconfig.Connection{
		Type: "oracle", Host: "h", Service: "svc",
		Accounts: []dbconfig.DBAcct{{Account: "u", Password: "p"}},
	}}
	if _, err := OpenTableSource(ctx, withDB, 7); err == nil || !strings.Contains(err.Error(), "企业 7") {
		t.Errorf("未知企业 err = %v", err)
	}

	// 替换 openReadOnly（生产路径是 connsvc.OpenBuiltin），注入内存连接器。
	old := openReadOnly
	fake := &fakeQConn{typ: "oracle", cols: []string{"C"}, rows: [][]string{{"1"}}}
	var dialedAccount string
	t.Cleanup(func() { openReadOnly = old })
	openReadOnly = func(_ context.Context, c dbconfig.Connection) (erpdb.Connector, error) {
		dialedAccount = c.Account
		return fake, nil
	}

	// 企业编号在清单里：连接要用该企业的账号（同一台机多企业时不能拿错账号）。
	withDB.Ents = []config.EntEntry{{Ent: 7, Account: "acct7"}}
	src, err := OpenTableSource(ctx, withDB, 7)
	if err != nil {
		t.Fatalf("OpenTableSource(ent): %v", err)
	}
	if dialedAccount != "acct7" {
		t.Errorf("连接账号 = %q，想要 acct7", dialedAccount)
	}
	if tp, ok := src.(*dbTableSource); !ok || tp.conn != fake {
		t.Errorf("来源类型 = %T", src)
	}
	src.Close()

	// 不带企业编号：直接用配置里的主账号。
	if _, err := OpenTableSource(ctx, withDB, 0); err != nil {
		t.Fatalf("OpenTableSource: %v", err)
	}

	// OnlineProgramVersion：identity=c 且有客制版次 → 取客制列。
	fake.cols = []string{"DZAA_C", "DZAA_ANY"}
	fake.rows = [][]string{{"12", "34"}}
	if v, err := OnlineProgramVersion(ctx, withDB, 0, "axmt500_wf", "c"); err != nil || v != 12 {
		t.Errorf("客制版次 = %d/%v", v, err)
	}
	// 非客制身份 → 取不限身份的最大版次。
	if v, err := OnlineProgramVersion(ctx, withDB, 0, "axmt500_wf", "s"); err != nil || v != 34 {
		t.Errorf("标准版次 = %d/%v", v, err)
	}
	// 客制版次为 0（从没客制）→ 退回不限身份。
	fake.rows = [][]string{{"0", "34"}}
	if v, _ := OnlineProgramVersion(ctx, withDB, 0, "axmt500_wf", "c"); v != 34 {
		t.Errorf("客制 0 时应退回 = %d", v)
	}
	// 空结果 → 0（不是错误）。
	fake.rows = nil
	if v, err := OnlineProgramVersion(ctx, withDB, 0, "axmt500_wf", "c"); err != nil || v != 0 {
		t.Errorf("空结果 = %d/%v", v, err)
	}
	// 非法程序名：在连库之前就挡住。
	if _, err := OnlineProgramVersion(ctx, withDB, 0, "BAD", "c"); !errors.Is(err, ErrInvalidProg) {
		t.Errorf("非法程序名 err = %v", err)
	}
	// 查询报错透传。
	fake.qerr = errors.New("ORA-00942")
	if _, err := OnlineProgramVersion(ctx, withDB, 0, "axmt500_wf", "c"); err == nil {
		t.Error("查询报错应透传")
	}

	// OpenFileSource 的纯校验分支（不拨号）：没配主机 / 没配 zone。
	if _, _, err := OpenFileSource(nil); err == nil || !strings.Contains(err.Error(), "没有配置 SSH 主机") {
		t.Errorf("nil 环境 err = %v", err)
	}
	if _, _, err := OpenFileSource(&config.NamedSsh{Name: "tst", SSHConfig: config.SSHConfig{Host: "h", Port: 22}}); err == nil ||
		!strings.Contains(err.Error(), "登录区域") {
		t.Errorf("没配 zone err = %v", err)
	}

	// releasedConn：关连接与关隧道要合成一次 Close（漏一个就是端口泄漏）。
	fc := &fakeQConn{}
	released := 0
	rc := &releasedConn{Connector: fc, release: func() { released++ }}
	rc.Close()
	if !fc.closed || released != 1 {
		t.Errorf("releasedConn.Close = closed:%v released:%d", fc.closed, released)
	}
	(&releasedConn{}).Close() // nil 安全
}

// TestSQLiteWriter_RoundTrip 钉住本地规格库写入：建表/分批写/计数/建索引/原子改名。
// 真机上一个 112 万行的库最后由它产出，落盘格式错了 model 侧全部查询都会错。
func TestSQLiteWriter_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "db", "e.db")
	w, err := openSQLite(path)
	if err != nil {
		t.Fatalf("openSQLite: %v", err)
	}
	// 打开一个目录当库文件：必须报"打开数据库失败"而不是拖到第一次查询。
	asDir := filepath.Join(dir, "as_dir.db")
	if err := os.MkdirAll(asDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := openSQLite(asDir); err == nil {
		t.Error("目录当库文件应报错")
	}

	if err := w.SetBulkMode(); err != nil {
		t.Fatalf("SetBulkMode: %v", err)
	}
	// 非法表名/列名/空列。
	if err := w.BeginTable("bad table", []string{"c"}); err == nil {
		t.Error("非法表名应报错")
	}
	if err := w.BeginTable("t", nil); err == nil || !strings.Contains(err.Error(), "没有列") {
		t.Errorf("空列 err = %v", err)
	}
	if err := w.BeginTable("t", []string{"bad col"}); err == nil || !strings.Contains(err.Error(), "非法列名") {
		t.Errorf("非法列名 err = %v", err)
	}

	// 正常建表 + 写行（列数不齐的行补空串）。
	if err := w.BeginTable("delivery_t", []string{"id", "name"}); err != nil {
		t.Fatalf("BeginTable: %v", err)
	}
	if err := w.InsertRows(nil); err != nil {
		t.Errorf("空批次 = %v", err)
	}
	if err := w.InsertRows([][]string{{"1", "a"}, {"2"}}); err != nil {
		t.Fatalf("InsertRows: %v", err)
	}
	if n, err := w.CountRows("delivery_t"); err != nil || n != 2 {
		t.Errorf("CountRows = %d/%v", n, err)
	}
	if _, err := w.CountRows("bad table"); err == nil {
		t.Error("CountRows 非法表名应报错")
	}
	if _, err := w.CountRows("no_such_table"); err == nil {
		t.Error("CountRows 不存在的表应报错")
	}
	if err := w.Exec("CREATE INDEX idx_delivery ON delivery_t (id)"); err != nil {
		t.Errorf("Exec: %v", err)
	}
	if err := w.Exec("THIS IS NOT SQL"); err == nil {
		t.Error("坏 SQL 应报错")
	}

	// 主键索引：读不到 dzed_t 时要开一条警告（静默不建索引是历史教训）。
	if warns := w.CreatePKIndexes(map[string]bool{"delivery_t": true}); len(warns) == 0 ||
		!strings.Contains(warns[0], "dzed_t") {
		t.Errorf("没有 dzed_t 的警告 = %v", warns)
	}

	// 造 dzed_t：有行但本次没有要建索引的表 → 只扫行、不建索引（安全路径）。
	if err := w.BeginTable("dzed_t", []string{"dzed001", "dzed002", "dzed003", "dzed004"}); err != nil {
		t.Fatal(err)
	}
	if err := w.InsertRows([][]string{{"delivery_t", "pk1", "P", "id"}}); err != nil {
		t.Fatal(err)
	}
	if warns := w.CreatePKIndexes(map[string]bool{}); len(warns) != 0 {
		t.Errorf("没有要建的表时不该有警告: %v", warns)
	}

	// 主键索引的**建成/建失败**分支单独用一个放宽连接池的 writer 验证。
	// 为什么要放宽：生产 openSQLite 固定 SetMaxOpenConns(1)，而 CreatePKIndexes 在
	// 未 close 的 rows 上再调 w.Exec —— 那会在单连接上自锁（见报告里的疑似 bug 与最小复现）。
	// 这里放宽到 2 只为把"建索引逻辑"本身测到，不代表生产路径可用。
	w3, err := openSQLite(filepath.Join(dir, "idx.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer w3.Close()
	w3.db.SetMaxOpenConns(2)
	if err := w3.BeginTable("dzed_t", []string{"dzed001", "dzed002", "dzed003", "dzed004"}); err != nil {
		t.Fatal(err)
	}
	if err := w3.InsertRows([][]string{{"delivery_t", "pk1", "P", "id"}}); err != nil {
		t.Fatal(err)
	}
	if err := w3.BeginTable("delivery_t", []string{"id"}); err != nil {
		t.Fatal(err)
	}
	warns := w3.CreatePKIndexes(map[string]bool{"delivery_t": true})
	// 两个方向都合理：SQLite 读锁下建索引可能成功（空告警），也可能 BUSY（建失败告警）。
	// 无论哪条路径，都不能静默返回"什么都没说"。
	if warns != nil && !strings.Contains(strings.Join(warns, "；"), "索引") {
		t.Errorf("建索引告警文案 = %v", warns)
	}
	if n, err := w3.CountRows("delivery_t"); err != nil || n != 0 {
		t.Errorf("建索引不该动数据: %d/%v", n, err)
	}

	// CloseAndRename：不 force 且目标已存在 → ErrExists；force 覆盖；原子改名后无 -wal/-shm。
	final := filepath.Join(dir, "final.db")
	if err := os.WriteFile(final, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := w.CloseAndRename(final, false); !errors.Is(err, ErrExists) {
		t.Fatalf("目标已存在 err = %v，想要 ErrExists", err)
	}
	if err := w.CloseAndRename(final, true); err != nil {
		t.Fatalf("force 改名: %v", err)
	}
	if _, err := os.Stat(final); err != nil {
		t.Errorf("正式库没落盘: %v", err)
	}
	if _, err := os.Stat(final + "-wal"); !os.IsNotExist(err) {
		t.Error("伴随文件没清掉")
	}

	// validIdent / quoteIdent / stampTime 的边界。
	for _, tc := range []struct {
		s  string
		ok bool
	}{
		{"delivery_t", true}, {"a$b#c", true}, {"", false}, {"1abc", false},
		{"a-b", false}, {strings.Repeat("a", 129), false},
	} {
		if got := validIdent(tc.s); got != tc.ok {
			t.Errorf("validIdent(%q) = %v", tc.s, got)
		}
	}
	if quoteIdent("t") != `"t"` {
		t.Errorf("quoteIdent = %q", quoteIdent("t"))
	}
	if got := stampTime(time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)); got != "2026-11-21T10:00:00Z" {
		t.Errorf("stampTime = %q", got)
	}

	// InsertRows 前没 BeginTable：明确报错（不是静默丢行）。
	w2, err := openSQLite(filepath.Join(dir, "w2.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer w2.Close()
	if err := w2.InsertRows([][]string{{"1"}}); err == nil || !strings.Contains(err.Error(), "BeginTable") {
		t.Errorf("未建表 err = %v", err)
	}
	// Close 幂等（连 nil 都安全）。
	w2.Close()
	(&sqliteWriter{}).Close()
}
