package erpdb

import "testing"

// TestRewritePlaceholdersDialects 方言占位符改写。
func TestRewritePlaceholdersDialects(t *testing.T) {
	cases := []struct {
		dialect string
		in      string
		want    string
		count   int
	}{
		{"oracle", "SELECT * FROM t WHERE a=? AND b=?", "SELECT * FROM t WHERE a=:1 AND b=:2", 2},
		{"kingbase", "INSERT INTO t(a,b) VALUES(?,?)", "INSERT INTO t(a,b) VALUES($1,$2)", 2},
		// 字面量/注释/引号标识符里的 ? 不动。
		{"oracle", "SELECT '?' FROM t WHERE a=?", "SELECT '?' FROM t WHERE a=:1", 1},
		{"oracle", `SELECT "a?b" FROM t WHERE c=?`, `SELECT "a?b" FROM t WHERE c=:1`, 1},
		{"oracle", "SELECT 1 -- ?\nFROM t WHERE a=?", "SELECT 1 -- ?\nFROM t WHERE a=:1", 1},
		{"oracle", "SELECT 1 /* ? */ FROM t WHERE a=?", "SELECT 1 /* ? */ FROM t WHERE a=:1", 1},
		// 转义的 '' 不结束字符串。
		{"oracle", "SELECT 'it''s ?' FROM t WHERE a=?", "SELECT 'it''s ?' FROM t WHERE a=:1", 1},
	}
	for _, c := range cases {
		got, n, err := RewritePlaceholders(c.in, c.dialect)
		if err != nil {
			t.Fatalf("%s/%q: %v", c.dialect, c.in, err)
		}
		if got != c.want || n != c.count {
			t.Errorf("%s/%q => %q (n=%d), want %q (n=%d)", c.dialect, c.in, got, n, c.want, c.count)
		}
	}
}

// TestBindSQLRejectsParamCountMismatch 占位符与参数数量不一致时早失败。
func TestBindSQLRejectsParamCountMismatch(t *testing.T) {
	if _, _, err := bindSQL("SELECT ? FROM t", "oracle", nil); err == nil {
		t.Fatal("占位符 1 个、参数 0 个应报错")
	}
	if _, _, err := bindSQL("SELECT 1 FROM t", "oracle", []string{"x"}); err == nil {
		t.Fatal("占位符 0 个、参数 1 个应报错")
	}
	if _, args, err := bindSQL("SELECT ? FROM t", "kingbase", []string{"x"}); err != nil || len(args) != 1 || args[0] != "x" {
		t.Fatalf("正常绑定失败: args=%v err=%v", args, err)
	}
}
