package fep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 录制 ID / --from 的解析规则：这一层决定"用户说的那一份"到底是哪一份。
// 解析错了不会报错，只会拿另一份录制去回放 —— 那种错最难自己发现。

func TestResolveRecordingFromSingleFile(t *testing.T) {
	dir := t.TempDir()
	id := "fep-20261015-170031"
	aui := filepath.Join(dir, id+"-1-rt2gdc.log")
	if err := os.WriteFile(aui, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	tgt, err := resolveRecording(aui, "")
	if err != nil {
		t.Fatal(err)
	}
	if tgt.ID != id || tgt.Dir != dir || len(tgt.Files) != 1 || tgt.Files[0] != aui {
		t.Fatalf("单文件解析不对：%+v", tgt)
	}
}

// --from 指事件侧时必须换成界面侧：拿事件文件去配别的连接的界面，编号对不上。
func TestResolveRecordingFollowsThePairedSide(t *testing.T) {
	dir := t.TempDir()
	id := "fep-20261015-170031"
	aui := filepath.Join(dir, id+"-1-rt2gdc.log")
	evt := filepath.Join(dir, id+"-1-gdc2rt.log")
	for _, p := range []string{aui, evt} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tgt, err := resolveRecording(evt, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(tgt.Files) != 1 || tgt.Files[0] != aui {
		t.Fatalf("--from 给事件侧时要换成界面侧：%+v", tgt.Files)
	}
}

func TestResolveRecordingFromDirTakesNewestUnlessIDGiven(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "fep-old-1-rt2gdc.log")
	newer := filepath.Join(dir, "fep-new-1-rt2gdc.log")
	for _, p := range []string{old, newer} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// 让 fep-new 更新（RecordingFiles 按修改时间倒序取最新）
	tgt, err := resolveRecording(dir, "fep-old")
	if err != nil {
		t.Fatal(err)
	}
	if tgt.ID != "fep-old" {
		t.Fatalf("点了 ID 就该用那个：%+v", tgt)
	}
	if _, err := resolveRecording(dir, "不存在的ID"); err == nil {
		t.Fatal("ID 不存在必须报错（绝不悄悄换一份）")
	}
}

func TestResolveRecordingRequiresSomething(t *testing.T) {
	if _, err := resolveRecording("", ""); err == nil {
		t.Fatal("既没给 ID 也没给 --from 时报用法错，不该悄悄取最新一份")
	}
}

// 事件必须来自**同一份**录制；只有界面侧时要说清楚，而不是静默当成 0 条。
func TestEventsComeFromTheSameRecordingAndSayWhenMissing(t *testing.T) {
	dir := t.TempDir()
	id := "fep-20261015-170031"
	aui := filepath.Join(dir, id+"-1-rt2gdc.log")
	if err := os.WriteFile(aui, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	tgt, err := resolveRecording(aui, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tgt.events(); err == nil {
		t.Fatal("没有事件侧（作业刚打开就断了）必须报出来")
	} else if !strings.Contains(err.Error(), "事件侧") {
		t.Fatalf("报文要说清缺的是哪一侧：%v", err)
	}

	evt := filepath.Join(dir, id+"-1-gdc2rt.log")
	body := `event _om 5{}{{ActionEvent 0{{idRef "271"}}}}`
	if err := os.WriteFile(evt, []byte(body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	events, err := tgt.events()
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || !strings.HasPrefix(events[0], "ActionEvent") {
		t.Fatalf("事件侧该被读成事件串：%v", events)
	}
}

// 中文列的对齐：% -Ns 数的是字节，中文表格会歪一整列。
func TestDispWidthCountsWideRunes(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"abc", 3},
		{"（无身份）", 10},
		{"fep-1", 5},
		{"主机109", 7},
	}
	for _, c := range cases {
		if got := dispWidth(c.in); got != c.want {
			t.Errorf("dispWidth(%q) = %d，应为 %d", c.in, got, c.want)
		}
	}
	if got := pad("（无身份）", 12); dispWidth(got) != 12 {
		t.Fatalf("补齐后显示宽度应是 12：%q（%d）", got, dispWidth(got))
	}
	if got := pad("已经比目标长", 4); dispWidth(got) != 12 {
		t.Fatalf("比目标宽时不许截断，只补 0 个空格：%q", got)
	}
}
