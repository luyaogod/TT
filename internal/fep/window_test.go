package fep

import (
	"strings"
	"testing"
)

// 地基：**对象属于哪个窗口实例**（`WindowOf`）与**窗口有哪些网格**（`TablesOfWindow`）。
//
// 为什么这是一条地基：驱动每一句都要先按窗口收窄（列与格子的 id 跨实例复用，
// 同名同 id 的对象在模型里同时存在多份）。今天两次踩的都是它：
//   - 主窗的明细网格 #644 在 `--scope` 里看不到（旧写法只登记"与窗口同一条消息里声明"的网格）；
//   - 料件窗口的 `--scope` 里冒出供应商的列（旧写法拿 Scope/Win 兑底，而 Win 被无条件改写）。
//
// 判据是从一份真会话里把**三套证据逐条对过**才定的（见 WindowOf 的注释）：
// DeclWindow（最近一次声明的消息所属窗口）最准，父链次之，Win 最后。

// TestLiveWindowsHaveDisjointControls 是这个地基的核心不变式：
// **两个活着的窗口不能共有同一个对象 id**。
//
// 理由：一个裸 id 在任一时刻只属于一个实例（运行期复用 id，但复用时旧的已经 rn 掉了）。
// 所以“料件窗口的 --scope 里能看到供应商的列”这种串窗，一定意味着归属算错了。
// 这条不变式不靠具体角色名，因此在任何语料上都成立。
func TestLiveWindowsHaveDisjointControls(t *testing.T) {
	for _, name := range append(fixtureFiles(t), maybeLiveFiles(t)...) {
		m := foldFixture(t, name)
		if m == nil {
			continue // 空壳录制
		}
		cat := Catalog(m, NewResolver(m))
		byWin := map[int]map[int]string{}
		for _, c := range cat {
			if c.WindowID == 0 {
				continue
			}
			if byWin[c.WindowID] == nil {
				byWin[c.WindowID] = map[int]string{}
			}
			byWin[c.WindowID][c.ID] = c.Role
		}
		wins := make([]int, 0, len(byWin))
		for w := range byWin {
			wins = append(wins, w)
		}
		for i := 0; i < len(wins); i++ {
			for j := i + 1; j < len(wins); j++ {
				a, b := wins[i], wins[j]
				for id, role := range byWin[a] {
					if brole, dup := byWin[b][id]; dup {
						t.Errorf("%s：控件 #%d 同时属于窗口 #%d(%s) 与 #%d(%s) —— 有一个是串窗",
							name, id, a, role, b, brole)
					}
				}
			}
		}
	}
}

// TestEveryGridBelongsToALiveWindow：网格的归属不许指向不存在的窗口（指向幽灵 =
// 之后所有按窗口的收窄都会失效）。
func TestEveryGridBelongsToALiveWindow(t *testing.T) {
	for _, name := range fixtureFiles(t) {
		m := foldFixture(t, name)
		if m == nil {
			continue
		}
		windows := map[int]bool{}
		for _, w := range m.Windows() {
			windows[w.ID] = true
		}
		grids := m.ByType("Table")
		if len(grids) == 0 {
			t.Fatalf("%s：一个网格都没折叠出来", name)
		}
		orphans := 0
		for _, g := range grids {
			w := m.WindowOf(g.ID)
			if w == 0 {
				// 归属问不出来 = 它属于一个已经关掉的实例（运行期不复位这些对象的归属）——
				// 这是**如实结果**，不是错误；错误是"归给一个不存在的窗口"。
				orphans++
				continue
			}
			if !windows[w] {
				t.Errorf("%s：网格 #%d(%s) 归属窗口 #%d，但那个窗口不在模型里",
					name, g.ID, g.Attr("name"), w)
			}
			// 归给它的窗口必须是个 Window（不是被复用的同名裸 id）
			if o, ok := m.Get(w); !ok || o.Type != "Window" {
				t.Errorf("%s：网格 #%d 归属的 #%d 不是 Window", name, g.ID, w)
			}
		}
		for _, w := range m.Windows() {
			a, b := m.TablesOfWindow(w.ID), m.TablesOfWindow(w.ID)
			if len(a) != len(b) {
				t.Errorf("%s：窗口 #%d 的网格列表不稳定", name, w.ID)
			}
			seen := map[int]bool{}
			for _, g := range a {
				if seen[g] {
					t.Errorf("%s：窗口 #%d 的网格 #%d 重复出现", name, w.ID, g)
				}
				seen[g] = true
			}
		}
		t.Logf("%s：网格 %d 个，其中归属已关闭实例的 %d 个", name, len(grids), orphans)
	}
}

// TestOpenWinWindowOwnsItsResultGrid：开窗窗口必须能问出"它的结果网格"——
// 驱动每次都要拿它发 pageSize/currentRow，问不出来就只能瞎猜。
func TestOpenWinWindowOwnsItsResultGrid(t *testing.T) {
	for _, name := range fixtureFiles(t) {
		m := foldFixture(t, name)
		if m == nil {
			continue
		}
		for _, w := range m.Windows() {
			wname := w.Attr("name")
			if !strings.HasPrefix(wname, "w_qry_") {
				continue
			}
			grids := m.TablesOfWindow(w.ID)
			if len(grids) == 0 {
				t.Errorf("%s：开窗窗口 #%d %s 没有网格", name, w.ID, wname)
				continue
			}
			found := false
			for _, g := range grids {
				if o, ok := m.Get(g); ok && o.Attr("name") == "s_qry" {
					found = true
				}
			}
			if !found {
				t.Errorf("%s：开窗窗口 #%d %s 的网格 %v 里没有 s_qry", name, w.ID, wname, grids)
			}
		}
	}
}

// maybeLiveFiles 找现场语料（有就跑、没有就只跑夹具）；不跳过。
func maybeLiveFiles(t *testing.T) []string {
	t.Helper()
	files, _ := maybeRecordFiles(t)
	return files
}
