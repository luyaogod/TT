package file

import (
	"reflect"
	"testing"
)

// -e 的两种写法都要能拆出多个模式：重复给、逗号分隔。
// 逗号写法存在的唯一理由，是绕开 PowerShell/CMD 把 `--include 'a|b'` 当管道。
func TestSplitIncludePatterns(t *testing.T) {
	cases := []struct {
		in   []string
		want []string
	}{
		{nil, []string{}},
		{[]string{"*.4gl"}, []string{"*.4gl"}},
		{[]string{"*.4gl", "*.4fd"}, []string{"*.4gl", "*.4fd"}},
		{[]string{"*.4gl,*.4fd"}, []string{"*.4gl", "*.4fd"}},
		{[]string{" *.4gl , *.4fd "}, []string{"*.4gl", "*.4fd"}},
		{[]string{"*.4gl,,", "", "  "}, []string{"*.4gl"}},
	}
	for _, c := range cases {
		got := splitIncludePatterns(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitIncludePatterns(%q) = %q，期望 %q", c.in, got, c.want)
		}
	}
}

// TestIncludeOK 的既有用例在 text_test.go（那里已覆盖 nil / 多模式 / 非法模式）；
// 这里只补 -e 拆分后仍能正确过滤。
func TestIncludeOKAfterSplit(t *testing.T) {
	pats := splitIncludePatterns([]string{"*.4fd, *.4gl"})
	if !includeOK("s_apmt510.4gl", pats) {
		t.Errorf("-e '*.4fd, *.4gl' 拆分后应命中 .4gl，实得 %q", pats)
	}
	if includeOK("s_apmt510.42m", pats) {
		t.Errorf("-e '*.4fd, *.4gl' 拆分后不该命中 .42m，实得 %q", pats)
	}
}
