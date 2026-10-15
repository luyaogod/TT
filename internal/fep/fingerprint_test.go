package fep

import (
	"path/filepath"
	"strings"
	"testing"

	"tt/internal/testkit"
)

// 指纹（录制件的身份）这几条规则，每一条都对着一个实测教训。
// 现场测出来的数字与命令见 internal/fep/README 的「指纹闸」一节；
// 这里钉的是**规则本身**：哪张表算同一张、业务值算不算数、闸门什么时候拒。

// 夹具（demoAUI）里的对象全都要进表：跨路径实测的 1556 条就是这个口径。
func TestEntriesCoverEveryObjectAndCarryNoBusinessValues(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(demoAUI())
	es := EntriesOf(m)
	if len(es) != m.Len() {
		t.Fatalf("编号表要有模型里的每一条：表 %d ／ 模型 %d", len(es), m.Len())
	}
	// 名字取 name → comment；业务值（text/value）**不进表** ——
	// 同一条路径连录两轮，单号会从 …003 变成 …004，拿值当签名会把能用的录制判失效。
	for _, e := range es {
		if strings.Contains(e.Role, "演示字段一") || strings.Contains(e.Role, "无键说明") {
			t.Fatalf("界面上的文本/值不许进指纹：%+v", e)
		}
	}
	// name 优先：FormField 400 的协议名是 demo_t.demo001
	var got string
	for _, e := range es {
		if e.ID == 400 {
			got = e.Role
		}
	}
	if got != "demo_t.demo001" {
		t.Fatalf("角色要取协议 name，got %q", got)
	}
	// 没有 name 的那条退回 comment
	for _, e := range es {
		if e.ID == 401 && e.Role != "cmt_demo001" {
			t.Fatalf("没有 name 要退到 comment，got %q", e.Role)
		}
	}
}

// 值变了、表不变 ⇒ 指纹不变。这是"同一条路径可以录一次反复用"的前提。
func TestFingerprintIgnoresValuesAndDeclarationOrder(t *testing.T) {
	a := NewModel()
	a.ApplyMessage(demoAUI())
	base := FingerprintOf(EntriesOf(a))

	// 同样的树，换一批业务值（数量/文本/单号）
	changed := strings.NewReplacer(
		`{value ""}`, `{value "2026-10-15-000004"}`,
		`{text "演示数量"}`, `{text "演示数量（改过）"}`,
	).Replace(demoAUI())
	b := NewModel()
	b.ApplyMessage(changed)
	if got := FingerprintOf(EntriesOf(b)); got != base {
		t.Fatalf("值变了不该改指纹：%s → %s", base, got)
	}

	// 表一样但顺序（读数顺序）不同 ⇒ 还是同一个指纹
	es := EntriesOf(a)
	rev := make([]FPEntry, 0, len(es))
	for i := len(es) - 1; i >= 0; i-- {
		rev = append(rev, es[i])
	}
	if got := FingerprintOf(rev); got != base {
		t.Fatalf("同一条目不同顺序该是同一个指纹：%s → %s", base, got)
	}

	// 编号错一个 ⇒ 另一个指纹（这就是"启动历史不同"的最小体现）
	one := append([]FPEntry(nil), es...)
	one[0].ID += 1000
	if got := FingerprintOf(one); got == base {
		t.Fatal("编号动了指纹必须跟着动")
	}
}

func TestCompareFPCountsBothDirections(t *testing.T) {
	rec := []FPEntry{{Kind: "Value", Scope: 1, ID: 1, Role: "a"}, {Kind: "Value", Scope: 1, ID: 2, Role: "b"}}
	live := []FPEntry{{Kind: "Value", Scope: 1, ID: 1, Role: "a"}, {Kind: "Value", Scope: 1, ID: 9, Role: "b"}}
	d := CompareFP(rec, live)
	if d.Same != 1 || d.Differ != 1 || d.OnlyNew != 0 || d.OnlyRec != 1 {
		t.Fatalf("逐条口径不对：%+v", d)
	}
	if d.MatchPct() != 50 {
		t.Fatalf("MatchPct = %d，应为 50", d.MatchPct())
	}
	// 跨路径的样子：坐标一个都对不上（编号全不同）
	other := []FPEntry{{Kind: "Value", Scope: 7, ID: 41, Role: "a"}}
	d2 := CompareFP(rec, other)
	if d2.Same != 0 || d2.OnlyNew != 1 {
		t.Fatalf("跨路径该是 0 对得上：%+v", d2)
	}
}

// 空表必须是"没有指纹"，而不是空串的 sha256：后者看着像个真指纹，会让人以为
// "这份录制有指纹，只是对不上"。
func TestEmptyTableHasNoFingerprint(t *testing.T) {
	f := FingerprintOf(nil)
	if f.Total != 0 || f.Hash != "" {
		t.Fatalf("空表不该有指纹：%+v", f)
	}
	if !strings.Contains(f.String(), "没有指纹") {
		t.Fatalf("空指纹要如实说：%s", f.String())
	}
	if got := FingerprintOf([]FPEntry{}); got.Hash != "" {
		t.Fatalf("空切片同理：%+v", got)
	}
}

// gate 把"整会话一张表"当单连接面用（闸门本身是按连接比的，见 TestGateComparesPerConnection）。
func gate(rec, live Snapshot) GateVerdict { return GateReplay(OfSnapshot(rec), OfSnapshot(live)) }

// 闸门**按连接比**：一个会话里各条连接可以是完全不同的两棵树。
//
// 实测来源：`r.d` 会话有两条连接 —— #1 是 fgldeb 调试器的界面（297 条），
// #2 才是作业界面（3976 条）。把两条揉成一张表去比单连接的录制件，
// 同一条路径也会判成 93% 不过；按连接比，#2 对 #2 是 100%。
func TestGateComparesPerConnection(t *testing.T) {
	job := func(shift int) []FPEntry {
		var out []FPEntry
		for i := 1; i <= 200; i++ {
			out = append(out, FPEntry{Kind: "Value", Scope: 5, ID: i + shift, Role: "f"})
		}
		return out
	}
	dbg := []FPEntry{{Kind: "Edit", Scope: 0, ID: 11, Role: "debugger.cmd"}}

	rec := MultiSnapshot{Label: "件", Path: PathRunnerRD, Conns: []ConnSnapshot{
		{Conn: 1, Snapshot: Snapshot{Entries: dbg, FP: FingerprintOf(dbg)}},
		{Conn: 2, Snapshot: Snapshot{Entries: job(0), FP: FingerprintOf(job(0))}},
	}}
	// 当前会话只起了作业那一条连接（例如我们要在调试线上重放 r.d 录的件）
	live := MultiSnapshot{Label: "会话", Path: PathRunnerRD, Conns: []ConnSnapshot{
		{Conn: 2, Snapshot: Snapshot{Entries: job(0), FP: FingerprintOf(job(0))}},
	}}
	v := GateReplay(rec, live)
	if !v.OK {
		t.Fatalf("按连接比时 #2↔#2 该通过：%+v", v)
	}
	if v.RecConn != 2 || v.LiveConn != 2 {
		t.Fatalf("该采用 #2↔#2 那一对：rec=%d live=%d", v.RecConn, v.LiveConn)
	}
	// 同一条路径但作业换了（编号整体平移）⇒ 一条连接都对不上 ⇒ 拒
	other := MultiSnapshot{Label: "会话", Path: PathRunnerRD, Conns: []ConnSnapshot{
		{Conn: 2, Snapshot: Snapshot{Entries: job(1000), FP: FingerprintOf(job(1000))}},
	}}
	if v := GateReplay(rec, other); v.OK {
		t.Fatalf("换了作业该拒：%+v", v)
	} else if !strings.Contains(v.Reason, "录制件 #") || !strings.Contains(v.Reason, "当前连接") {
		t.Fatalf("拒绝报文要说清比的是哪一对、两边各有哪些连接：%s", v.Reason)
	}
}

// 一条连接读不动不该把整份录制判死，也不该静默当成"没有界面"。
func TestGateReportsPerConnectionReadFailures(t *testing.T) {
	good := []FPEntry{{Kind: "Window", Scope: 0, ID: 1, Role: "w"}}
	rec := MultiSnapshot{Label: "件", Conns: []ConnSnapshot{
		{Conn: 1, Snapshot: EmptySnapshot("件 #1", "读不动：坏帧")},
		{Conn: 2, Snapshot: Snapshot{Entries: good, FP: FingerprintOf(good)}},
	}}
	live := MultiSnapshot{Label: "会话", Conns: []ConnSnapshot{
		{Conn: 2, Snapshot: Snapshot{Entries: good, FP: FingerprintOf(good)}},
	}}
	if v := GateReplay(rec, live); !v.OK {
		t.Fatalf("#1 读不动、#2 对得上 ⇒ 该通过：%+v", v)
	}
	// 全都读不动时才说"没有界面"，并且把原因带上
	allBad := MultiSnapshot{Label: "件", Conns: []ConnSnapshot{
		{Conn: 1, Snapshot: EmptySnapshot("件 #1", "读不动：坏帧")},
	}}
	v := GateReplay(allBad, live)
	if v.Code != GateNoRecording || !strings.Contains(v.Reason, "坏帧") {
		t.Fatalf("空面要把原因说出来：%+v", v)
	}
}

// 闸门：路径不符直接拒（不必等指纹）。
func TestGateRefusesDifferentLaunchPath(t *testing.T) {
	es := []FPEntry{{Kind: "Window", Scope: 0, ID: 10, Role: "w"}}
	same := FingerprintOf(es)
	v := gate(
		Snapshot{Label: "件", Path: PathRunnerRR, Entries: es, FP: same},
		Snapshot{Label: "会话", Path: PathDebug, Entries: es, FP: same},
	)
	if v.OK || v.Code != GatePathDiff {
		t.Fatalf("跨路径必须拒：%+v", v)
	}
	if !strings.Contains(v.Reason, "r.r") || !strings.Contains(v.Reason, "debug") {
		t.Fatalf("报文要写清是哪两条路径：%s", v.Reason)
	}
	if strings.TrimSpace(v.Report()) == "" {
		t.Fatal("判决要能打给人看")
	}
}

// 闸门：编号表对不上就拒，并给出差异条数与前几条现场。
func TestGateRefusesWhenTablesDiffer(t *testing.T) {
	rec := make([]FPEntry, 0, 100)
	live := make([]FPEntry, 0, 100)
	for i := 1; i <= 100; i++ {
		rec = append(rec, FPEntry{Kind: "Value", Scope: 3, ID: i, Role: "f"})
		live = append(live, FPEntry{Kind: "Value", Scope: 3, ID: i + 500, Role: "f"}) // 另一条启动历史
	}
	v := gate(
		Snapshot{Label: "件", Path: PathRunnerRR, Entries: rec, FP: FingerprintOf(rec)},
		Snapshot{Label: "会话", Path: PathRunnerRR, Entries: live, FP: FingerprintOf(live)},
	)
	if v.OK || v.Code != GateFPNoMatch {
		t.Fatalf("编号表对不上必须拒：%+v", v)
	}
	if v.Diff.Same != 0 || v.Diff.LiveTotal != 100 {
		t.Fatalf("差异条数要如实报：%+v", v.Diff)
	}
	if len(v.Examples) == 0 {
		t.Fatal("拒绝时要给出「录制里是哪个编号、现在这个坐标是哪个编号」的头几条")
	}
}

// 闸门：同一条路径、当前界面是录制件的一个前缀（起完作业就录的样子）⇒ 放行。
//
// 这是**回放的真实场景**：录制件是"整个会话"，当前会话只有刚打开的那一屏。
func TestGatePassesWhenLiveIsAPrefixOfTheRecording(t *testing.T) {
	var rec []FPEntry
	for i := 1; i <= 300; i++ {
		rec = append(rec, FPEntry{Kind: "Value", Scope: 3, ID: i, Role: "f"})
	}
	live := append([]FPEntry(nil), rec[:300]...)
	v := gate(
		Snapshot{Label: "件", Path: PathRunnerRR, Entries: rec, FP: FingerprintOf(rec)},
		Snapshot{Label: "会话", Path: PathRunnerRR, Entries: live, FP: FingerprintOf(live)},
	)
	if !v.OK || !v.PathChecked {
		t.Fatalf("同路径前缀该放行：%+v", v)
	}

	// 界面的 1% 允许对不上（实测同路径会有极少数裸 id 复用导致的重声明）：
	// 3/300 = 1% 仍在门内，4/300 就出去。
	bad := append([]FPEntry(nil), live...)
	bad[0].ID, bad[1].ID, bad[2].ID = 9001, 9002, 9003
	if v = gate(Snapshot{Entries: rec, Path: PathRunnerRR, FP: FingerprintOf(rec)},
		Snapshot{Label: "会话", Path: PathRunnerRR, Entries: bad, FP: FingerprintOf(bad)}); !v.OK {
		t.Fatalf("1%%（3/300）该放行：%+v", v)
	}
	bad[3].ID = 9004
	if v = gate(Snapshot{Entries: rec, Path: PathRunnerRR, FP: FingerprintOf(rec)},
		Snapshot{Label: "会话", Path: PathRunnerRR, Entries: bad, FP: FingerprintOf(bad)}); v.OK {
		t.Fatalf("超过 1%% 该拒：%+v", v)
	}
}

// 有一边不知道路径时**不比路径**，但判决完全靠指纹（不是"路径没比就放过"）。
func TestGateFallsBackToFingerprintWhenPathUnknown(t *testing.T) {
	es := []FPEntry{{Kind: "Window", Scope: 0, ID: 10, Role: "w"}}
	v := gate(
		Snapshot{Label: "件", Entries: es, FP: FingerprintOf(es)},
		Snapshot{Label: "会话", Path: PathRunnerRR, Entries: es, FP: FingerprintOf(es)},
	)
	if !v.OK || v.PathChecked {
		t.Fatalf("路径不知道时不比路径、但要如实报没比：%+v", v)
	}
	other := []FPEntry{{Kind: "Window", Scope: 0, ID: 99, Role: "w"}}
	v2 := gate(
		Snapshot{Label: "件", Entries: other, FP: FingerprintOf(other)},
		Snapshot{Label: "会话", Path: PathRunnerRR, Entries: es, FP: FingerprintOf(es)},
	)
	if v2.OK || v2.Code != GateFPNoMatch {
		t.Fatalf("路径不知道就更要靠指纹：%+v", v2)
	}
}

// 空的两面各有各的说法：录制件里没界面 ≠ 当前会话没界面。
func TestGateDistinguishesEmptySides(t *testing.T) {
	es := []FPEntry{{Kind: "Window", Scope: 0, ID: 10, Role: "w"}}
	if v := gate(Snapshot{Label: "件"}, Snapshot{Label: "会话", Entries: es, FP: FingerprintOf(es)}); v.Code != GateNoRecording {
		t.Fatalf("录制件没界面要报 no-recording：%+v", v)
	}
	if v := gate(Snapshot{Label: "件", Entries: es, FP: FingerprintOf(es)}, Snapshot{Label: "会话"}); v.Code != GateNoLive {
		t.Fatalf("当前会话没界面要报 no-live：%+v", v)
	}
}

// **真夹具**上的判决：两份不同会话/不同作业的界面，闸门必须拒。
//
// 为什么用夹具而不是合成表：合成表只能证明"我写的比较函数按我写的规则算"，
// 而这里的两张表是从两份真录制件里重建出来的（各 1500+ 条），
// 它们对不上是因为**界面真的不一样**。实测（2026-10-15，主机109）这一对是 79%。
func TestGateRefusesTwoDifferentRealInterfaces(t *testing.T) {
	dir := filepath.Join(testkit.RepoRoot(t), "testdata", "fep-recordings")
	rec, err := SnapshotOfFiles([]string{filepath.Join(dir, "id-reuse.txt")}, "id-reuse")
	if err != nil {
		t.Fatalf("读夹具: %v", err)
	}
	live, err := SnapshotOfFiles([]string{filepath.Join(dir, "openwin-with-dialogs.txt")}, "openwin")
	if err != nil {
		t.Fatalf("读夹具: %v", err)
	}
	if rec.FP.Total < 1000 || live.FP.Total < 1000 {
		t.Fatalf("夹具太小，这条测试会变成空转：rec=%d live=%d", rec.FP.Total, live.FP.Total)
	}
	if rec.FP.Hash == live.FP.Hash {
		t.Fatal("两份不同界面的指纹不该相同")
	}
	v := gate(rec, live)
	if v.OK {
		t.Fatalf("两份真界面不同，闸门必须拒：%+v", v)
	}
	if v.Diff.Same*100/v.Diff.LiveTotal >= MinLiveMatchPct {
		t.Fatalf("这个夹具对已经跑到门口了（%d%%），换一份：%+v", v.Diff.MatchPct(), v.Diff)
	}
	// 同一份自己跟自己：一定放行（前缀关系的最简情形）
	if v := gate(rec, rec); !v.OK {
		t.Fatalf("同一份录制自己跟自己该放行：%+v", v)
	}
}
