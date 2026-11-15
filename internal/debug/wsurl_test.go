package debug

// wsURLPrefix 决定 `/api/status` 里 `wsPrefix` 的值（前端拿它拼占位符服务地址）。
//
// 这组用例钉的是「**不能**用 host.ZoneTNSName 推导」这件事：
// 那张表现在只覆盖旧式菜单码 1/2/3/4、对不上返回空串，而本机出貨區的菜单码是 `t`。
// 用 ZoneTNSName 推导会让 zone=t 的站点拿到 `/w`（前端拼 /wtopprd，后端要 /wstopprd）。

import "testing"

func TestWSURLPrefix(t *testing.T) {
	cases := []struct{ zone, want string }{
		{"t", "/ws"},      // 本机出貨區菜单码
		{"1", "/ws"},      // 旧式出貨區菜单码
		{"topprd", "/ws"}, // 直接给 ZONE 字符串
		{"35", "/w"},      // t35tst
		{"36", "/w"},      // t35prd
		{"31", "/w"},      // t35dev
		{"39", "/w"},      // t35pth
		{"2", "/w"},       // toptst 不是出貨區
		{"3", "/w"},       // topstd
		{"4", "/w"},       // toppth
		{"36k", "/w"},     // 金仓版不是出貨區
		{"", "/w"},
	}
	for _, c := range cases {
		if got := wsURLPrefix(c.zone); got != c.want {
			t.Errorf("wsURLPrefix(%q) = %q, want %q", c.zone, got, c.want)
		}
	}
}
