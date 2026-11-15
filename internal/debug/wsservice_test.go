package debug

import (
	"strings"
	"testing"
)

func TestWSServiceLikePattern(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"", "", true},
		{"wssp", "%wssp%", true},
		{"wssp*", "wssp%", true},
		{"wssp?00", "wssp_00", true},
		{"a%b", `%a\%b%`, true},
		{"a_b", `%a\_b%`, true},
		{"a*b_c", `a%b\_c`, true},
		{"O'Brien", "", false},
		{`a\b`, "", false},
		{"  easy.query  ", "%easy.query%", true},
	}
	for _, c := range cases {
		got, ok := wsServiceLikePattern(c.in)
		if ok != c.ok || got != c.want {
			t.Errorf("wsServiceLikePattern(%q) = (%q,%v), want (%q,%v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestWSServiceWhere(t *testing.T) {
	cases := []struct {
		name string
		f    WSServiceFilter
		want []string
		not  []string
	}{
		{"空条件", WSServiceFilter{}, []string{"1=1"}, []string{"LIKE", "1=0"}},
		{
			"三字段 AND",
			WSServiceFilter{Spec: "wssp*", Name: "查询", API: "easy.query"},
			[]string{
				`UPPER(svc_spec) LIKE UPPER('wssp%') ESCAPE '\'`,
				`UPPER(svc_name) LIKE UPPER('%查询%') ESCAPE '\'`,
				`UPPER(svc_api) LIKE UPPER('%easy.query%') ESCAPE '\'`,
			},
			nil,
		},
		{"单引号退化为空集", WSServiceFilter{Name: "a'; DROP"}, []string{"AND 1=0"}, []string{"DROP"}},
	}
	for _, c := range cases {
		got := wsServiceWhere(c.f)
		for _, w := range c.want {
			if !strings.Contains(got, w) {
				t.Errorf("%s: 缺少 %q\n实际: %s", c.name, w, got)
			}
		}
		for _, n := range c.not {
			if strings.Contains(got, n) {
				t.Errorf("%s: 不应出现 %q\n实际: %s", c.name, n, got)
			}
		}
	}
}

func TestWSServicePagingSQL(t *testing.T) {
	ora := wsServiceDataSQL(false, "1=1", 3, 200)
	for _, w := range []string{"OFFSET 400 ROWS FETCH FIRST 200 ROWS ONLY", "ORDER BY svc_spec", "gzja_t", "gzjc_t"} {
		if !strings.Contains(ora, w) {
			t.Errorf("Oracle SQL 缺少 %q:\n%s", w, ora)
		}
	}
	kb := wsServiceDataSQL(true, "1=1", 2, 50)
	for _, w := range []string{"OFFSET 50 LIMIT 50", "ORDER BY svc_spec"} {
		if !strings.Contains(kb, w) {
			t.Errorf("Kingbase SQL 缺少 %q:\n%s", w, kb)
		}
	}
	if !strings.Contains(wsServiceCountSQL("1=1"), "COUNT(1)") {
		t.Error("count SQL 缺少 COUNT(1)")
	}
}

func TestParseWSServiceRows(t *testing.T) {
	out := "wssp000|简易查询服务API|easy.query\r\nwssp00001|创建销售订单变更单数据|t100.oapi.sales.order.change\r\n\r\nbadline\nwsspX||\n"
	items := parseWSServiceRows(out)
	if len(items) != 3 {
		t.Fatalf("想要 3 条，得到 %d: %#v", len(items), items)
	}
	if items[0].Spec != "wssp000" || items[0].API != "easy.query" {
		t.Errorf("第 1 条解析错: %#v", items[0])
	}
	if items[1].Spec != "wssp00001" {
		t.Errorf("第 2 条解析错: %#v", items[1])
	}
	if items[2].Spec != "wsspX" || items[2].Name != "" || items[2].API != "" {
		t.Errorf("空列解析错: %#v", items[2])
	}
}

func TestParseWSServiceCount(t *testing.T) {
	if got := parseWSServiceCount("2129\n"); got != 2129 {
		t.Errorf("count = %d, want 2129", got)
	}
	if got := parseWSServiceCount("ORA-00942\n"); got != 0 {
		t.Errorf("错误行应得 0，得到 %d", got)
	}
}
