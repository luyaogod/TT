package fep

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"tt/internal/testkit"
)

// W1：**真实录制语料上的不变量回归**。
//
// 为什么必须有它：2026-10-11 查了一整天才找到的根因是 —— AUI 同步的三种操作
// `an`/`un`/`rn` 里，`rn`（删节点）被当成 `un` 处理，于是被删掉的 Dialog / 工具栏项 /
// 模板条目**永远不死**。"同一个 id 既是 Item 又是 TableColumn""accept 有三个候选"
// "focus 乱跳" 全是这一个 bug 的表现，而我当时只能靠肉眼在几百行流里看出来。
//
// 手工发现一次不算修好 —— 要把它钉成每次 `go test` 都会自己跑回来的检查。
//
// 语料分两层：
//
//  1. **仓库夹具** `testdata/fep-recordings/*.txt`（必跑）：从现场录制剪出来的增量子流，
//     任何机器上都在，所以这条测试不会变成空转。裁剪方式见 tools/fepfix。
//  2. **现场语料** `%APPDATA%\tt\fep\*-rt2gdc.log`（可选）：活体作业录的会话语料，
//     含真实业务数据、不随仓库走；用 TT_FEP_RECORD_DIR 可指到别处。
//     实测规模（2026-10-11）：94 个连接 / 910 条 AUI 消息，其中 49 个是空壳
//     （会话刚起来就死的连接），最大一份 1.19MB / 80 条消息。非 AUI 帧（type=5）不参与。

// fixtureFiles 返回仓库里的夹具（必有，所以这条测试在哪儿都会真的跑）。
func fixtureFiles(t *testing.T) []string {
	t.Helper()
	dir := filepath.Join(testkit.RepoRoot(t), "testdata", "fep-recordings")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("夹具目录读不到（%s）: %v", dir, err)
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	if len(out) == 0 {
		t.Fatalf("%s 里没有夹具（*.txt）—— 夹具是这条回归的牙，不能没有", dir)
	}
	return out
}

// recordDir 找本机语料目录（测试里只用它做诊断打印）。
func recordDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("TT_FEP_RECORD_DIR")
	if dir == "" {
		dir = filepath.Join(os.Getenv("APPDATA"), "tt", "fep")
	}
	return dir
}

// maybeRecordFiles 找现场语料；没有就返回 (nil,false)，**不跳过**。
//
// 为什么要一个不跳过的版本：「必跑的夹具测试」在有现场语料的机器上应该顺手多覆盖
// 一份真语料，而不是因为"语料可有可无"就整条退化成只跑夹具。
func maybeRecordFiles(t *testing.T) ([]string, bool) {
	t.Helper()
	dir := os.Getenv("TT_FEP_RECORD_DIR")
	if dir == "" {
		if app := os.Getenv("APPDATA"); app != "" {
			dir = filepath.Join(app, "tt", "fep")
		}
	}
	if dir == "" {
		return nil, false
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return nil, false
	}
	files, err := RecordingFiles(dir, "-rt2gdc.log")
	if err != nil || len(files) == 0 {
		return nil, false
	}
	var out []string
	for _, f := range files {
		if st, err := os.Stat(f); err == nil && st.Size() > 8<<20 {
			continue
		}
		out = append(out, f)
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

// recordFiles 收集现场录制件（跳过过大的，避免测试时间失控）；一个都没有就跳过。
func recordFiles(t *testing.T) []string {
	t.Helper()
	files, ok := maybeRecordFiles(t)
	if !ok {
		t.Skip("没有现场录制语料（TT_FEP_RECORD_DIR 或数据目录 fep 都没有）")
	}
	return files
}

// corpusStat 是一个录制件的折叠统计（诊断用）。
type corpusStat struct {
	file  string
	size  int64
	msgs  int
	objs  int
	reuse int // 同作用域内一个裸 id 被用过多种类型的次数
	noSeq int // 候选之间分不出先后的次数（真错误）
}

// victim 记录"本批消息要删的对象，删之前长什么样"。
type victim struct {
	id    int
	scope int
	typ   string
}

// foldRecording 按消息顺序折叠录制件，并在**每一步**做独立检查。
//
// **一个录制件 = 一个连接 = 一个模型**：不同连接/会话的 AUI id 毫无关系（每个作业实例
// 各有一套 id），混在一个模型里会造出成千上万条假冲突 —— 这个测试的第一版就是这么错的，
// 所以这里一个文件一个模型。
//
// rnAsUn=true 时把 `rn` 改写成 `un` —— 用来反证检查确实抓得住那个 bug（见 TestTheRnInvariantHasTeeth）。
//
// 检查刻意不依赖模型自己的结论：删之前先问模型"这个 id 是什么"，删之后要求它真的不在。
func foldRecording(t *testing.T, files []string, rnAsUn bool) (stats []corpusStat, violations []string) {
	t.Helper()
	add := func(v string) {
		if len(violations) < 12 {
			violations = append(violations, v)
		}
	}
	for _, f := range files {
		bodies, err := ReadRecording(f)
		if err != nil {
			t.Fatalf("读录制件 %s: %v", f, err)
		}
		base := filepath.Base(f)
		m := NewModel()
		st := corpusStat{file: base}
		if fi, err := os.Stat(f); err == nil {
			st.size = fi.Size()
		}
		for _, body := range bodies {
			nodes := ParseMessage(body)
			if len(nodes) == 0 {
				continue
			}
			st.msgs++
			// 1) 先把"本批要删谁、删之前是什么"记下来
			var victims []victim
			for _, n := range nodes {
				if n.Type != "rn" || n.ID == 0 {
					continue
				}
				if o, ok := m.Get(n.ID); ok {
					victims = append(victims, victim{id: n.ID, scope: o.Scope, typ: o.Type})
				}
			}
			// 2) 需要反证时把 rn 退化成 un（复现旧 bug）
			if rnAsUn {
				for i := range nodes {
					if nodes[i].Type == "rn" {
						nodes[i].Type = "un"
					}
				}
			}
			m.Apply(nodes)
			// 3) 本批重新声明过的 id（同批"先删再加"是合法形态）
			redeclared := map[int]bool{}
			for _, n := range nodes {
				if n.ID != 0 && !n.Bare && n.Type != "un" && n.Type != "rn" {
					redeclared[n.ID] = true
				}
			}
			// 4) 不变量：被 rn 的东西不许还活着 —— **该裸 id 的任何一个候选**都不许
			// （同作用域、其它作用域、全局表都不行）。rn 说的是"运行期把这个 id 删了"，
			// 它是硬事实；我的作用域归属只是我自己的结论，可能算错。
			for _, v := range victims {
				if redeclared[v.id] {
					continue
				}
				if cands := m.Candidates(v.id); len(cands) > 0 {
					add(fmt.Sprintf("%s: 被 rn 的 #%d(%s,scope=%d) 还活着（候选 %d 份，最新是 %s,scope=%d）",
						base, v.id, v.typ, v.scope, len(cands), cands[0].Type, cands[0].Scope))
				}
				if v.typ != "Dialog" {
					continue
				}
				for _, d := range m.ActiveDialogs() {
					if d.ID == v.id && d.Scope == v.scope {
						add(fmt.Sprintf("%s: 被 rn 的 Dialog #%d(scope=%d) 仍是当前交互指令的候选",
							base, v.id, v.scope))
					}
				}
			}
		}
		// 5) 这一个连接折叠完之后的全局不变量。
		//
		// 关于"同一作用域一个 id 用过多种类型"：实测运行期**会复用裸 id**
		// （同一个 159 既做过全局 ActionDefault、又做过网格格子；同一个 4937 既做过
		// 下拉 Item、又做过 Dialog 的 Action），所以这本身不是错误 —— 但**必须可判**：
		// 候选之间要能分出先后（Seq 严格递减），Get(id) 才能给出"现在是谁"。
		// 分不出先后 = 驱动只能猜 = 真错误。
		for _, id := range allIDs(m) {
			cands := m.Candidates(id)
			for i := 1; i < len(cands); i++ {
				if cands[i].Seq == cands[i-1].Seq {
					st.noSeq++
					add(fmt.Sprintf("%s: #%d 的两份候选 Seq 相同（%d），分不出\"现在是谁\"",
						base, id, cands[i].Seq))
				}
			}
			if len(cands) > 1 {
				seen := map[int]string{}
				for _, c := range cands {
					if prev, ok := seen[c.Scope]; ok && prev != c.Type {
						st.reuse++
					}
					seen[c.Scope] = c.Type
				}
			}
			got, ok := m.Get(id)
			if !ok || got.ID != id {
				add(fmt.Sprintf("%s: Get(#%d) 取不到（ok=%v）", base, id, ok))
			} else if len(cands) > 0 && got.Seq != cands[0].Seq {
				add(fmt.Sprintf("%s: Get(#%d) 没给出最新那份（%d != %d）", base, id, got.Seq, cands[0].Seq))
			}
		}
		// Objects() 必须与 Candidates 一致，且不能有已删的幽灵。
		for _, o := range m.Objects() {
			if len(m.Candidates(o.ID)) == 0 {
				add(fmt.Sprintf("%s: Objects() 里的 #%d 不在 Candidates 里", base, o.ID))
			}
		}
		// 网格登记的目标必须存在（winTables 是"窗口↔它的网格"的折叠结论，
		// 曾经因为它指向已删对象而把事件发到旧实例上）。
		for w, grids := range m.winTables {
			for _, g := range grids {
				if _, ok := m.Get(g); !ok {
					add(fmt.Sprintf("%s: 窗口 #%d 登记的网格 #%d 不在模型里", base, w, g))
				}
			}
		}
		st.objs = m.Len()
		stats = append(stats, st)
	}
	if len(stats) == 0 {
		t.Skip("没有录制件：跳过")
	}
	return stats, violations
}

// allIDs 收集模型里出现过的所有裸 id。
func allIDs(m *Model) []int {
	seen := map[int]bool{}
	for id := range m.objs {
		seen[id] = true
	}
	for k := range m.locals {
		seen[k.ID] = true
	}
	out := make([]int, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}

// reportStats 打印语料规模（前几名按大小），让"测试跑的是不是真东西"一眼可见。
func reportStats(t *testing.T, stats []corpusStat) {
	t.Helper()
	sort.Slice(stats, func(i, j int) bool { return stats[i].size > stats[j].size })
	total, reuse := 0, 0
	for _, s := range stats {
		total += s.msgs
		reuse += s.reuse
	}
	t.Logf("语料 %d 个连接 / 共 %d 条 AUI 消息 / id 复用（同作用域多类型）%d 处", len(stats), total, reuse)
	for i, s := range stats {
		if i >= 5 {
			break
		}
		t.Logf("  %9d 字节  %5d 条  %6d 对象  id 复用 %4d  %s", s.size, s.msgs, s.objs, s.reuse, s.file)
	}
	empty := 0
	for _, s := range stats {
		if s.msgs == 0 {
			empty++
		}
	}
	if empty > 0 {
		t.Logf("  其中 %d 个录制件没有任何可用消息（多半是会话刚起来就死的空壳）", empty)
	}
}

// TestFixtureRecordingsKeepInvariants 是**必跑**的那半：仓库夹具上不许出现
// "被 rn 的对象还活着"，且每个裸 id 必须能判出"现在是谁"。
func TestFixtureRecordingsKeepInvariants(t *testing.T) {
	files := fixtureFiles(t)
	stats, violations := foldRecording(t, files, false)
	reportStats(t, stats)
	for _, v := range violations {
		t.Errorf("不变量被破坏：%s", v)
	}
}

// TestRealRecordingsNeverKeepRemovedObjects 是**现场语料**那半（可选，含真实数据）：
// 每个连接各自折叠，全程不许出现同样的问题。语料不在就跳过（台账已登记）。
func TestRealRecordingsNeverKeepRemovedObjects(t *testing.T) {
	files := recordFiles(t)
	stats, violations := foldRecording(t, files, false)
	reportStats(t, stats)
	for _, v := range violations {
		t.Errorf("不变量被破坏：%s", v)
	}
	total := 0
	for _, s := range stats {
		total += s.msgs
	}
	if total == 0 {
		t.Fatal("折叠完一条消息都没有：解析链断了")
	}
}

// TestTheRnInvariantHasTeeth 反证：把 `rn` 退化成 `un`（旧 bug）后，上面那条不变量
// **必须**被打破 —— 否则说明这个检查是摆设。跑夹具（必跑）与现场语料（有则跑）。
func TestTheRnInvariantHasTeeth(t *testing.T) {
	campaigns := map[string][]string{"夹具": fixtureFiles(t)}
	if files, ok := maybeRecordFiles(t); ok {
		campaigns["现场语料"] = files
	}
	for name, files := range campaigns {
		_, violations := foldRecording(t, files, true)
		if len(violations) == 0 {
			t.Errorf("%s：把 rn 当成 un 之后居然一条都不违反 —— 这个不变量抓不住那个 bug，需要重写", name)
			continue
		}
		t.Logf("%s：反证成立，rn→un 时抓到 %d 条（示例：%s）", name, len(violations), violations[0])
	}
}
