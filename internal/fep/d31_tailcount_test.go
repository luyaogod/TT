package fep

import (
	"fmt"
	"strings"
	"testing"
)

// TestListTrueCount 钉住台账 D31 的判据：
//
// 判决行里的「X 个身份 / Y 条边」必须是**真实条数**。真机出现过
// 「明细 13 条、尾部写着 118 条」—— 因为 `diffSide` 截断时把余下的折成
// `… 等 N 条`（那一句自己也算一个元素），而判决行直接写了 `len(list)`：
// 既把标记算成一条，又漏掉被折掉的 N。
func TestListTrueCount(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want int
	}{
		{"空清单", nil, 0},
		{"没有截断", []string{"a", "b", "c"}, 3},
		{"截断（条）", append(rep("e", 13), "… 等 118 条"), 131},
		{"截断（类）", append(rep("k", 5), "… 等 7 类"), 12},
		{"只有标记", []string{"… 等 9 条"}, 9},
		{"标记不完整（当普通条目）", []string{"a", "… 等 x 条"}, 2},
		{"标记没有数字（当普通条目）", []string{"… 等 条"}, 1},
	}
	for _, c := range cases {
		if got := listTrueCount(c.in); got != c.want {
			t.Errorf("%s: listTrueCount = %d, want %d（清单 %v）", c.name, got, c.want, c.in)
		}
	}
}

func rep(s string, n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, fmt.Sprintf("%s%d", s, i))
	}
	return out
}

// TestJudgeReportCountsAreTrue 是 D31 的端到端判据：
// 造出**会被截断**的差异，判决行的数字必须等于"明细条数 + 尾部折掉的条数"。
func TestJudgeReportCountsAreTrue(t *testing.T) {
	const limit = 3
	// 每条边都给一个不同的"父"，保证是 8 条互不相同的差异边。
	var rec, live []string
	for i := 0; i < 8; i++ {
		rec = append(rec, fmt.Sprintf("b_live#%d→b_child#%d", i, i))
	}
	missing, extra := multisetDiff(rec, live, limit)
	// rec 有、live 没有 ⇒ 归「尚未加载」那一栏（missing）。
	if got := listTrueCount(missing); got != 8 {
		t.Fatalf("真实条数该是 8，得 %d（清单 %v）", got, missing)
	}
	if len(missing) != limit+1 {
		t.Fatalf("截断后明细应是 limit+1 条（%d 条 + 1 句标记），得 %d：%v", limit, len(missing), missing)
	}
	if !strings.Contains(missing[len(missing)-1], "等 5 条") {
		t.Fatalf("尾部应写「… 等 5 条」（8-3），得 %q", missing[len(missing)-1])
	}
	if len(extra) != 0 {
		t.Fatalf("另一边不该有差异：%v", extra)
	}
	// 明细 + 尾部标记 == 真实条数（这正是人读时会对上的两个数）
	if len(missing)-1+5 != listTrueCount(missing) {
		t.Fatalf("明细 %d + 尾部 5 ≠ 真实 %d", len(missing)-1, listTrueCount(missing))
	}
}
