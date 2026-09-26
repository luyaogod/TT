package erpdb

import (
	"strings"
	"testing"
	"time"

	"tt/internal/dbconfig"
)

// conn 造一个只填了本文件关心的那几个字段的连接（用户名词典表要的是 User/Type/Host）。
func conn(user, typ, host string) dbconfig.Connection {
	return dbconfig.Connection{Type: typ, Host: host, User: user}
}

// 本包管的是**怎么连 ERP 库**：连接（要真库，不测）、标识符与字面量转义（纯函数，这里测）、
// 只读 SQL 的文本闸门（纯函数，这里测）。
//
// 转义这一层是**安全关键**：远端库的账号密码是明文配置、查询里拼的是用户给的表名与关键字，
// 一处漏了就成注入。所以下面的判据写得比"能跑"严：**拒绝**的那些样例才是重点。
//
// **口令一律是假的**（AGENTS.md §9）：这里用 "pw-dev" 只为分辨取到了哪个值。

func TestValidIdent(t *testing.T) {
	valid := []string{"dzea_t", "DS", "a", "A1", "_x", "t35prd", "gzou003", "a_b_c", "x9"}
	for _, s := range valid {
		if !ValidIdent(s) {
			t.Errorf("%q 该算合法标识符", s)
		}
	}
	invalid := []struct{ in, why string }{
		{"", "空"},
		{"9a", "数字开头"},
		{"0", "数字开头"},
		{"a-b", "连字符"},
		{"a b", "空格"},
		{"a.b", "点（限定名要拆开各自校验，不能整个放行）"},
		{"a;b", "分号"},
		{"a'b", "单引号"},
		{`a"b`, "双引号"},
		{"a(b)", "括号"},
		{"a*", "星号"},
		{"a%", "百分号"},
		{"a/../b", "路径式"},
		{"表", "非 ASCII"},
		{"a\nb", "换行"},
		{"a\tb", "制表符"},
		{"a--", "SQL 注释符"},
		{"a/*x*/", "SQL 块注释"},
	}
	for _, c := range invalid {
		if ValidIdent(c.in) {
			t.Errorf("%q 该被拒（%s）", c.in, c.why)
		}
	}

	// 保留字是**合法标识符** —— 本函数只管字符集，不做 SQL 关键字过滤。
	// 这条特意写出来，免得后来人以为"防注入"到这里就做完了。
	for _, kw := range []string{"drop", "select", "table", "insert"} {
		if !ValidIdent(kw) {
			t.Errorf("%q 在字符集上合法（本函数不过滤保留字），实得 false", kw)
		}
	}
}

// TestQuoteIdentReturnsEmptyOnInvalid 不过校验**返回空串**而不是原样返回 ——
// 拼进 SQL 会得到 "SELECT * FROM .t"，报的错与真实原因（表名非法）完全对不上。
func TestQuoteIdentReturnsEmptyOnInvalid(t *testing.T) {
	if got := QuoteIdent("dzea_t"); got != "dzea_t" {
		t.Errorf("合法标识符该原样返回，得 %q", got)
	}
	for _, bad := range []string{"", "a-b", "a;b", "表", "1a"} {
		if got := QuoteIdent(bad); got != "" {
			t.Errorf("非法标识符 %q 该返回空串，得 %q", bad, got)
		}
	}
}

// TestQuoteLitDoublesSingleQuotes 字面量转义只有一条规则：单引号翻倍。
//
// 这是**唯一**挡住值注入的东西（erpdb 走 simple protocol，没有绑定参数）。
func TestQuoteLitDoublesSingleQuotes(t *testing.T) {
	cases := map[string]string{
		"":             "''",
		"abc":          "'abc'",
		"O'Brien":      "'O''Brien'",
		"'; DROP--":    "'''; DROP--'",
		"''":           "''''''",
		"a'b'c":        "'a''b''c'",
		"%":            "'%'", // 通配符不加处理：它由调用方按 LIKE 语义自己决定
		`a"b`:          `'a"b'`,
		"line1\nline2": "'line1\nline2'",
	}
	for in, want := range cases {
		if got := QuoteLit(in); got != want {
			t.Errorf("QuoteLit(%q) = %q，想 %q", in, got, want)
		}
	}
	// 转义后的结果再"翻倍回去"应当还原：值里不可能出现落单的单引号。
	for in := range cases {
		got := QuoteLit(in)
		inner := got[1 : len(got)-1]
		if strings.Contains(strings.ReplaceAll(inner, "''", ""), "'") {
			t.Errorf("QuoteLit(%q) 的结果里有落单的单引号：%q", in, got)
		}
	}
}

// TestSelectAllSQL 表名与用户名都要过白名单 —— 用户名来自配置、表名可能来自用户。
func TestSelectAllSQL(t *testing.T) {
	got, err := SelectAllSQL(conn("ds", "oracle", "db1"), "dzea_t")
	if err != nil {
		t.Fatalf("该成功：%v", err)
	}
	if want := "SELECT * FROM ds.dzea_t"; got != want {
		t.Errorf("得 %q，想 %q", got, want)
	}

	bad := []struct{ name, user, table string }{
		{"表名带分号", "ds", "dzea_t; DROP TABLE x"},
		{"表名带空格", "ds", "dzea t"},
		{"表名带引号", "ds", "dzea_t'"},
		{"表名为空", "ds", ""},
		{"用户名带连字符", "d-s", "dzea_t"},
		{"用户名为空（连接没填凭据）", "", "dzea_t"},
		{"表名是限定名（点号要拆开各自校验，不能整个放行）", "ds", "other.dzea_t"},
	}
	for _, c := range bad {
		t.Run(c.name, func(t *testing.T) {
			if _, err := SelectAllSQL(conn(c.user, "oracle", "db1"), c.table); err == nil {
				t.Error("该被拒")
			}
		})
	}
}

// TestCheckReadOnlySQLIsAPrefixCheckNotAParser 钉住这道闸门的**强度**，不是它的完善度。
//
// 它做的事只有两件：TrimSpace+ToUpper 之后看**开头**是不是写操作、看有没有分号。
// 所以它挡得住"手写的 INSERT/DROP 直接贴进来"这种最常见的情况，**挡不住**
// 前面加一段注释的写法 —— 它不是解析器。
//
// 真正的文本闸门是 internal/safesql（`Check`，会先剥注释），这条是纵深防御的一层；
// `live` 那条路径更是自己构造 SQL（值走 QuoteLit、表名走 ValidIdent），这里只是兜底。
// 把它当"完善"来用才是危险的，所以下面把**它的边界**也一并钉住。
func TestCheckReadOnlySQLIsAPrefixCheckNotAParser(t *testing.T) {
	t.Run("拒绝：写操作开头", func(t *testing.T) {
		for _, kw := range []string{"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE",
			"REPLACE", "ATTACH", "DETACH", "PRAGMA", "REINDEX", "VACUUM", "GRANT", "REVOKE"} {
			for _, sql := range []string{
				kw + " TABLE t",
				strings.ToLower(kw) + " table t", // 大小写不敏感
				"  \t" + kw + " TABLE t",         // 前面有空白
			} {
				if err := checkReadOnlySQL(sql); err == nil {
					t.Errorf("%q 该被拒", sql)
				}
			}
		}
	})
	t.Run("拒绝：多语句（任何位置的分号）", func(t *testing.T) {
		for _, sql := range []string{
			"SELECT 1; DROP TABLE t",
			"SELECT 1;",
			"SELECT * FROM t WHERE x = 'a;b'", // 字面量里的分号也一并拒 —— 宁可错杀
		} {
			if err := checkReadOnlySQL(sql); err == nil {
				t.Errorf("%q 该被拒", sql)
			}
		}
	})
	t.Run("放行：普通的只读 SELECT", func(t *testing.T) {
		for _, sql := range []string{
			"SELECT * FROM ds.dzea_t",
			"select 1 from dual",
			"WITH x AS (SELECT 1) SELECT * FROM x",
			"SELECT * FROM t WHERE a = 'O''Brien'",
		} {
			if err := checkReadOnlySQL(sql); err != nil {
				t.Errorf("%q 该放行：%v", sql, err)
			}
		}
	})
	t.Run("边界（它的强度就到这儿）", func(t *testing.T) {
		// 空语句：TrimSpace 之后没有前缀、没有分号 → 放行。它拦不住，也没打算拦。
		if err := checkReadOnlySQL(""); err != nil {
			t.Errorf("空串会放行（它不是解析器）：%v", err)
		}
		// 前面加注释就绕过了"开头是写操作"这一条 —— 这**不是**"这里发现了漏洞"，
		// 而是"这里从来不是拦这个的地方"（safesql 会先剥注释）。
		bypass := "/* x */ DROP TABLE t"
		if err := checkReadOnlySQL(bypass); err != nil {
			t.Logf("（这一条现在被拒了：%v —— 那是好事，说明它比注释里写的更严，请更新本注释）", err)
		}
	})
}

func TestGoOraDSNEscapesEveryComponent(t *testing.T) {
	got := goOraDSN("db1", 1521, "t35prd", "ds", "pw-dev")
	if want := "oracle://ds:pw-dev@db1:1521/t35prd"; got != want {
		t.Errorf("得 %q，想 %q", got, want)
	}
	// 两端都要能塞进去：DSN 是拼出来的字符串，口令里的 @ 与 : 会改变它解析出什么。
	got = goOraDSN("db1", 1521, "svc", "u@h", "p:w")
	if want := "oracle://u%40h:p%3Aw@db1:1521/svc"; got != want {
		t.Errorf("特殊字符该被转义，得 %q，想 %q", got, want)
	}
}

func TestURLEscape(t *testing.T) {
	cases := map[string]string{
		"plain": "plain",
		"p@ss":  "p%40ss",
		"p:ss":  "p%3Ass",
		"a/b":   "a%2Fb",
		"a?b":   "a%3Fb",
		"100%":  "100%25",
		"":      "",
		"a b":   "a b",   // 空格不转义：本函数只管 DSN 里会改变解析的那几个字符
		"%40":   "%2540", // 先转 % 再转别的，所以已是转义序列的输入不会被二次解释
	}
	for in, want := range cases {
		if got := urlEscape(in); got != want {
			t.Errorf("urlEscape(%q) = %q，想 %q", in, got, want)
		}
	}
}

// TestKingbaseDSN 金仓那条用的是 key=value 形式，与 oracle 的 URL 形式不同。
func TestKingbaseDSN(t *testing.T) {
	c := conn("ds", "kingbase", "db1")
	c.Port = 54321
	c.Database = "topprd"
	c.User = "ds"
	c.Password = "pw-dev"
	got := kingbaseDSN(c)
	for _, want := range []string{"host=db1", "port=54321", "dbname=topprd", "user=ds", "password=pw-dev"} {
		if !strings.Contains(got, want) {
			t.Errorf("DSN 里该有 %q，实得 %q", want, got)
		}
	}
}

func TestScanString(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, ""},
		{"abc", "abc"},
		{[]byte("abc"), "abc"},
		{"", ""},
		{int64(42), "42"},     // 驱动可能给非字符串
		{float64(1.5), "1.5"}, // 不带多余的尾零
		{true, "true"},
		{time.Date(2026, 9, 26, 19, 30, 0, 0, time.UTC), "2026-09-26 19:30:00"},
	}
	for _, c := range cases {
		if got := scanString(c.in); got != c.want {
			t.Errorf("scanString(%v) = %q，想 %q", c.in, got, c.want)
		}
	}
}
