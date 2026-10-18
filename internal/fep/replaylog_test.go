package fep

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// 回放遥测（计划 C 的 D2）。三条判据：
//  1. 一行一步、**写一行 flush 一行**（边跑边读）；
//  2. `respAt` 没有就是 **null**，不是 0（0 会被读成"瞬回"，那是另一件事）；
//  3. 同一个 n 可能有多行（waiting → 终态、暂停 → 继续），读侧取**最后一行**。

func TestReplayLogRoundTripAndLastLineWins(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ReplayLogName("fep-1", time.Now()))
	w, err := NewReplayLogWriter(path, "fep-1", 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	at := func(ms int64) *int64 { return &ms }
	// 第 1 步：中间态 + 终态（等到回包）
	w.Step(ReplayStep{N: 1, Total: 3, Event: `ActionEvent 0{{idRef "42"}}`, Kind: "ActionEvent", Target: 42, SentAt: at(1000), Result: ReplayWaiting})
	w.Step(ReplayStep{N: 1, Total: 3, Event: `ActionEvent 0{{idRef "42"}}`, Kind: "ActionEvent", Target: 42, SentAt: at(1000), RespAt: at(1380), Result: ReplaySent, RuntimeBytes: 512})
	// 第 2 步：没有回包，而且是"本该有"的点击类
	w.Step(ReplayStep{N: 2, Total: 3, Event: `ActionEvent 0{{idRef "77"}}`, Kind: "ActionEvent", Target: 77, SentAt: at(2000), Result: ReplayNoResponse, Suspect: true, Reason: "8 秒内没有回包"})
	// 第 3 步：被 AUI 安全闸跳过（没发出去 → sentAt 必须是 null）
	w.Step(ReplayStep{N: 3, Total: 3, Event: `ConfigureEvent 0{{idRef "9"}}{{value "x"}}`, Kind: "ConfigureEvent", Result: ReplaySkipped, Reason: "对象 #9 不在当前界面（AUI 安全闸）"})
	if err := w.Close(ReplayResult{Sent: 1, Skipped: 1, Failed: 0, RuntimeBytes: 900}, "跑完"); err != nil {
		t.Fatal(err)
	}

	// 中间态与终态都在文件里（这就是"边跑边读"要的东西）
	raw, rerr := os.ReadFile(path)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if n := strings.Count(string(raw), "\n"); n != 6 { // 头 + 3~5 步 + 尾 = 6 行（第 1 步两行）
		t.Fatalf("该落 6 行（头 / 4 条步骤行 / 尾），得到 %d", n)
	}
	// null 而不是 0
	if !strings.Contains(string(raw), `"respAt":null`) {
		t.Fatalf("没有回包时 respAt 必须是 null（0 会被读成瞬回）：\n%s", raw)
	}
	if strings.Contains(string(raw), `"sentAt":0`) {
		t.Fatalf("跳过的步没有 sentAt，不该写成 0：\n%s", raw)
	}

	lg, err := ReadReplayLog(path)
	if err != nil {
		t.Fatal(err)
	}
	if lg.ID != "fep-1" || lg.Conn != 2 || lg.Total != 3 {
		t.Fatalf("头行没读对：%+v", lg)
	}
	if !lg.Done || lg.Note != "跑完" {
		t.Fatalf("尾行没读对：done=%v note=%q", lg.Done, lg.Note)
	}
	// 同一个 n 取最后一行：第 1 步是 sent（不是 waiting）
	st := lg.States()
	if len(st) != 3 {
		t.Fatalf("该收敛出 3 步：%d", len(st))
	}
	if st[0].Result != ReplaySent || st[0].RespAt == nil || *st[0].RespAt != 1380 {
		t.Fatalf("第 1 步该取最后那一行（sent + respAt=1380）：%+v", st[0])
	}
	if st[0].RuntimeBytes != 512 {
		t.Fatalf("这一步的运行期字节要留下：%+v", st[0])
	}
	if st[1].Result != ReplayNoResponse || !st[1].Suspect || st[1].RespAt != nil {
		t.Fatalf("第 2 步该是「没回包且可疑」：%+v", st[1])
	}
	if st[2].SentAt != nil {
		t.Fatalf("跳过的那一步没发出去，sentAt 必须是 null：%+v", st[2])
	}

	p := lg.Progress()
	if p.Running || p.Total != 3 || p.Done != 3 || p.Sent != 1 || p.NoResp != 1 || p.Suspect != 1 || p.Skipped != 1 {
		t.Fatalf("进度汇总不对：%+v", p)
	}

	// 分页：since = 已经拿到多少条（**按行位置**，不是 n —— 跳过/暂停会产生额外的行）
	if got := lg.From(0); len(got) != 4 {
		t.Fatalf("从头该拿到 4 条步骤行：%d", len(got))
	}
	if got := lg.From(4); len(got) != 0 {
		t.Fatalf("已经拿完了就不该再给：%d", len(got))
	}
	if got := lg.From(2); len(got) != 2 || got[0].N != 2 {
		t.Fatalf("从第 2 条起该拿到第 2、3 步：%+v", got)
	}
}

// 尾行不在 = "还在跑"（界面据此把进度条画成进行中）。
func TestReplayLogWithoutTrailerIsStillRunning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ReplayLogName("fep-2", time.Now()))
	w, err := NewReplayLogWriter(path, "fep-2", 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	w.Step(ReplayStep{N: 1, Total: 5, Event: "ActionEvent 0{}", Kind: "ActionEvent", Result: ReplayWaiting})
	// 不写尾行就放开句柄 = 回放进程被杀在半路时那个文件的样子。
	// 故意不走 Close：Close 会补上尾行，那样就测不到"没有尾行"这一支了。
	_ = w.Abandon()
	lg, err := ReadReplayLog(path)
	if err != nil {
		t.Fatal(err)
	}
	p := lg.Progress()
	if !p.Running || !p.Pending {
		t.Fatalf("没有尾行 = 还在跑，而且有一步正在等回包：%+v", p)
	}
	if p.Done != 0 {
		t.Fatalf("中间态不该算「已完成」：%+v", p)
	}
}

// 控制文件：一个词，覆盖式（只有最新一条有意义）。
// TestReplayCtlOneShotIsConsumed：step / skip 是**动作**，读完就删；pause/resume/stop 是状态，留着。
//
// 不消费的后果（真机 2026-10-18）：回放每步都重读文件里那条 `skip`，
// "跳过这一步"变成"从这里开始全跳过"（94 步全被跳过）；`step` 同理会一路跑到底。
func TestReplayCtlOneShotIsConsumed(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "fep-x-20260101-000000"+ReplayLogSuffix)
	for _, c := range []ReplayCtl{CtlStep, CtlSkip} {
		if err := WriteReplayCtl(log, c); err != nil {
			t.Fatal(err)
		}
		if got := ReadReplayCtl(log); got != c {
			t.Fatalf("第一次读该是 %q，实得 %q", c, got)
		}
		if got := ReadReplayCtl(log); got != "" {
			t.Fatalf("%q 是一次性命令，读完该没了，实得 %q", c, got)
		}
	}
	// 状态类命令读多少次都还在
	if err := WriteReplayCtl(log, CtlPause); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if got := ReadReplayCtl(log); got != CtlPause {
			t.Fatalf("pause 是状态，第 %d 次读该还是它，实得 %q", i+1, got)
		}
	}
}

func TestReplayCtlRoundTrip(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, ReplayLogName("fep-3", time.Now()))
	if got := ReadReplayCtl(log); got != "" {
		t.Fatalf("没有控制文件时该返回空：%q", got)
	}
	for _, c := range []ReplayCtl{CtlPause, CtlStep, CtlResume, CtlStop} {
		if err := WriteReplayCtl(log, c); err != nil {
			t.Fatal(err)
		}
		if got := ReadReplayCtl(log); got != c {
			t.Fatalf("写 %q 读回 %q", c, got)
		}
	}
	if !strings.HasSuffix(ReplayCtlPath(log), ".ctl") {
		t.Fatalf("控制文件该与日志同一个干：%s", ReplayCtlPath(log))
	}
}

// 找日志：新的在前；`fep-1` 不许把 `fep-10` 的日志认成自己的。
func TestFindReplayLogsIsNewestFirstAndPrefixed(t *testing.T) {
	dir := t.TempDir()
	mk := func(id, stamp string) string {
		p := filepath.Join(dir, id+"-"+stamp+ReplayLogSuffix)
		if err := os.WriteFile(p, []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	mk("fep-1", "20261016-100000")
	mk("fep-1", "20261016-120000")
	mk("fep-10", "20261016-130000")

	got, err := FindReplayLogs(dir, "fep-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("fep-1 该只有 2 份日志（不许把 fep-10 的算进来）：%v", got)
	}
	if !strings.Contains(got[0], "120000") {
		t.Fatalf("该新的在前：%v", got)
	}
	if lg, lerr := LatestReplayLog(dir, "fep-none"); lerr != ErrNoReplayLog {
		t.Fatalf("没有日志该回 ErrNoReplayLog，得到 %v / %+v", lerr, lg)
	}
}

// 头行丢了（被杀在半路）也要读得动：ID 从文件名切、步数按实际行数。
func TestReplayLogSurvivesMissingHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fep-9-20261016-120000"+ReplayLogSuffix)
	body := `{"n":1,"total":2,"event":"ActionEvent 0{}","kind":"ActionEvent","sentAt":5,"respAt":null,"result":"no-response"}` + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	lg, err := ReadReplayLog(path)
	if err != nil {
		t.Fatal(err)
	}
	if lg.ID != "fep-9-20261016-120000" || lg.Total != 1 || len(lg.Steps) != 1 {
		t.Fatalf("头行丢了也不该读不动：%+v", lg)
	}
	if lg.Done {
		t.Fatal("没有尾行就不是「结束了」")
	}
	// 坏行只计数，不报错
	if err := os.WriteFile(path, []byte(body+"{坏行\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	lg2, err := ReadReplayLog(path)
	if err != nil || lg2.Bad != 1 {
		t.Fatalf("坏行该照实计数：err=%v bad=%d", err, lg2.Bad)
	}
}

// 事件类别与 idRef：都从事件文本里读，读不到就不填（不猜）。
func TestEventKindAndTarget(t *testing.T) {
	cases := []struct{ ev, kind string }{
		{`ActionEvent 0{{idRef "42"}}`, "ActionEvent"},
		{`KeyEvent 0{{keyName "Tab"}}`, "KeyEvent"},
		{`ConfigureEvent 0{{idRef "9"}}{{value "x"}}`, "ConfigureEvent"},
		{`  ActionEvent 0{}`, "ActionEvent"},
	}
	for _, c := range cases {
		if got := eventKind(c.ev); got != c.kind {
			t.Errorf("eventKind(%q) = %q，该是 %q", c.ev, got, c.kind)
		}
	}
	// 回放步骤的 JSON 形状（前端按字段渲染，字段名就是契约）
	b, _ := json.Marshal(ReplayStep{N: 1, Result: ReplayNoResponse})
	for _, want := range []string{`"n":1`, `"respAt":null`, `"result":"no-response"`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("步骤的 JSON 里该有 %s：%s", want, b)
		}
	}
}
