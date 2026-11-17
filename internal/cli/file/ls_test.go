package file

import (
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/host"
)

func mkEntries(n int) []host.Entry {
	out := make([]host.Entry, 0, n)
	base := time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		out = append(out, host.Entry{
			Name:    string(rune('a'+i%26)) + string(rune('0'+i/26)),
			Size:    int64(i * 10),
			ModTime: base.Add(time.Duration(i) * time.Minute),
		})
	}
	return out
}

// 默认有界：超过 --limit 必须截断，且总数照实报（截断不能是静默的）。
func TestApplyLsOptionsLimit(t *testing.T) {
	entries := mkEntries(5)
	shown, total, truncated := applyLsOptions(entries, lsOptions{limit: 3, sort: "name"})
	if total != 5 || len(shown) != 3 || !truncated {
		t.Fatalf("limit=3 于 5 项：total=%d shown=%d truncated=%v，期望 5/3/true", total, len(shown), truncated)
	}
	// limit=0 = 不限：不截断、不报 truncated。
	shown, total, truncated = applyLsOptions(entries, lsOptions{limit: 0, sort: "name"})
	if total != 5 || len(shown) != 5 || truncated {
		t.Fatalf("limit=0 应不截断，实得 total=%d shown=%d truncated=%v", total, len(shown), truncated)
	}
	// 正好等于 limit 不算截断。
	shown, total, truncated = applyLsOptions(entries, lsOptions{limit: 5, sort: "name"})
	if total != 5 || len(shown) != 5 || truncated {
		t.Fatalf("正好 5 项、limit=5 不该算截断，实得 %v", truncated)
	}
}

func TestApplyLsOptionsSort(t *testing.T) {
	entries := []host.Entry{
		{Name: "b", Size: 300, ModTime: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)},
		{Name: "a", Size: 100, ModTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "c", Size: 200, ModTime: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
	}
	names := func(es []host.Entry) string {
		s := ""
		for _, e := range es {
			s += e.Name
		}
		return s
	}
	if got := names(mustShow(t, entries, lsOptions{sort: "name"})); got != "abc" {
		t.Errorf("--sort name = %q，期望 abc", got)
	}
	if got := names(mustShow(t, entries, lsOptions{sort: "name", desc: true})); got != "cba" {
		t.Errorf("--sort name --desc = %q，期望 cba", got)
	}
	if got := names(mustShow(t, entries, lsOptions{sort: "time"})); got != "acb" {
		t.Errorf("--sort time = %q，期望 acb", got)
	}
	if got := names(mustShow(t, entries, lsOptions{sort: "time", desc: true})); got != "bca" {
		t.Errorf("--sort time --desc = %q，期望 bca", got)
	}
	if got := names(mustShow(t, entries, lsOptions{sort: "size"})); got != "acb" {
		t.Errorf("--sort size = %q，期望 acb", got)
	}
	if got := names(mustShow(t, entries, lsOptions{sort: "size", desc: true})); got != "bca" {
		t.Errorf("--sort size --desc = %q，期望 bca", got)
	}
	// 入参不能被就地处决（调用方还要用它算别的）。
	if entries[0].Name != "b" {
		t.Errorf("applyLsOptions 改动了入参：%v", entries)
	}
}

func mustShow(t *testing.T, entries []host.Entry, opt lsOptions) []host.Entry {
	t.Helper()
	shown, _, _ := applyLsOptions(entries, opt)
	return shown
}

// 截断提示必须同时说清"少了多少"和"怎么要看全"。
func TestLsTruncNote(t *testing.T) {
	n := lsTruncNote(1200, 200)
	for _, want := range []string{"1000", "1200", "200", "--limit 0"} {
		if !strings.Contains(n, want) {
			t.Errorf("截断提示 %q 缺少 %q", n, want)
		}
	}
}
