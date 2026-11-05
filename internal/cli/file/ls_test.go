package file

// ls 的筛选与截断是**纯函数**，所以能不必起 SSH 就测：
// 它们决定了"目录有 200 条时调用方收到多少、被截掉多少、下一步怎么说"。

import (
	"fmt"
	"strings"
	"testing"
)

func sampleEntries(n int) []lsEntry {
	out := make([]lsEntry, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, lsEntry{Name: fmt.Sprintf("f%03d.42m", i), Type: "file"})
	}
	return out
}

func TestFilterEntriesGlob(t *testing.T) {
	entries := []lsEntry{
		{Name: "cwss_wssp00120_wf.42m"},
		{Name: "cwss_wssp00120_wf.4gl"},
		{Name: "cwss_wssp00120_wf.bak"},
		{Name: "cwss_wssp00160_wf.42m"},
	}
	got, err := filterEntries(entries, "*wssp00120*")
	if err != nil {
		t.Fatalf("合法 glob 不该报错: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("想要 3 条，得到 %d: %+v", len(got), got)
	}

	got, err = filterEntries(entries, "*.42m")
	if err != nil {
		t.Fatalf("合法 glob 不该报错: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("*.42m 该 2 条，得到 %d", len(got))
	}
}

// 服务器上真实存在大小写混排的文件名：筛选不许因此漏掉。
func TestFilterEntriesCaseInsensitive(t *testing.T) {
	entries := []lsEntry{{Name: "WSSP00120_WF.42M"}}
	if got, err := filterEntries(entries, "*wssp00120*"); err != nil || len(got) != 1 {
		t.Fatalf("大小写不该影响匹配，得到 %+v, err=%v", got, err)
	}
}

func TestFilterEntriesEmptyPatternIsNoop(t *testing.T) {
	entries := sampleEntries(5)
	if got, err := filterEntries(entries, "  "); err != nil || len(got) != len(entries) {
		t.Fatalf("空 pattern 应原样返回，得到 %d, err=%v", len(got), err)
	}
}

func TestFilterEntriesNoMatchIsEmptyNotError(t *testing.T) {
	if got, err := filterEntries(sampleEntries(3), "nope*"); err != nil || len(got) != 0 {
		t.Fatalf("筛不到应是空数组（成功），得到 %d, err=%v", len(got), err)
	}
}

// P2：畸形 glob（如 "[）必须作为用法错误显式报出，不许静默滤空。
func TestFilterEntriesBadPatternIsError(t *testing.T) {
	if _, err := filterEntries(sampleEntries(3), "["); err == nil {
		t.Fatal("畸形 glob 应报 ErrBadPattern，而不是吞掉")
	}
}

func TestLimitEntries(t *testing.T) {
	entries := sampleEntries(10)

	got, cut := limitEntries(entries, 4)
	if len(got) != 4 || cut != 6 {
		t.Fatalf("limit=4 应返回 4 且截掉 6，得到 %d/%d", len(got), cut)
	}
	// 截断不许静默：cut 必须等于"没给你的条数"。
	if len(got)+cut != len(entries) {
		t.Fatalf("返回 %d + 截掉 %d 应等于总数 %d", len(got), cut, len(entries))
	}

	if got, cut = limitEntries(entries, 0); len(got) != 10 || cut != 0 {
		t.Fatalf("limit=0 应不限，得到 %d/%d", len(got), cut)
	}
	if got, cut = limitEntries(entries, 10); len(got) != 10 || cut != 0 {
		t.Fatalf("limit=总数 应全部返回，得到 %d/%d", len(got), cut)
	}
	if got, cut = limitEntries(entries, 99); len(got) != 10 || cut != 0 {
		t.Fatalf("limit>总数 应全部返回，得到 %d/%d", len(got), cut)
	}
}

// P3：截断提示按"是否设了 --filter"分情况 —— 没设 filter 时 --filter 是空操作，
// 正解是 --limit 0 看全量；设了 filter 则给"看全部匹配"的可复制命令 + 收窄选项，
// 且不许把"看更多"说成"缩小"。
func TestTruncHintWordingDependsOnFilter(t *testing.T) {
	withFilter := truncHint("com/lib/42m", "*.42m", 100, 80)
	if !strings.Contains(withFilter, "--filter") {
		t.Fatalf("设了 filter 应给带 --filter 的可复制命令，得到: %s", withFilter)
	}
	if !strings.Contains(withFilter, "--limit 0") {
		t.Fatalf("设了 filter 时截断也该建议 --limit 0 看全部，得到: %s", withFilter)
	}
	if strings.Contains(withFilter, "缩小") {
		t.Fatalf("不许把'看更多'说成'缩小'，得到: %s", withFilter)
	}

	noFilter := truncHint("com/lib/42m", "", 100, 80)
	if strings.Contains(noFilter, "--filter") {
		t.Fatalf("没设 filter 时 --filter 是空操作，不该出现，得到: %s", noFilter)
	}
	if !strings.Contains(noFilter, "--limit 0") {
		t.Fatalf("没设 filter 时应建议 --limit 0 看全量，得到: %s", noFilter)
	}
}

// 帮助文本必须说明新增契约：调用方靠 --help 判断有没有这个能力。
func TestLsHelpMentionsFilterAndLimit(t *testing.T) {
	c := newLsCmd()
	long := c.Long
	for _, want := range []string{"--filter", "--limit", "truncated", "total", "returned"} {
		if !strings.Contains(long, want) {
			t.Errorf("ls --help 缺 %q（这是 --filter/--limit 的契约，调用方靠它发现能力）", want)
		}
	}
	if c.Flags().Lookup("filter") == nil {
		t.Error("ls 没有 --filter flag")
	}
	if c.Flags().Lookup("limit") == nil {
		t.Error("ls 没有 --limit flag")
	}
	if c.Example == "" {
		t.Error("ls --help 缺可复制示例")
	}
}
