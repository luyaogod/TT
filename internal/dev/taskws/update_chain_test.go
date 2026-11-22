package taskws

// Update 三条刷新分支的离线端到端测试（用假 $TOP 驱动 Check/Fetch）。
//
// README 的三条分支：
//
//	① 视图干净、包没 apply 过 → 重拉材料 → 重建索引 → 重新打包 → 重新导出视图
//	② 视图有未落盘改动      → 拒绝；--force 只更新材料与索引
//	③ 包里已有 apply 过的成果 → 只更新材料与索引（成果不能冲掉）
//
// 外加"材料与索引都是最新的，什么都没改""--check 一个字节都不写""材料没变但索引该重建"
// 这三条边界；以及失败面（Fetch 中途失败、建包目录失败）要冒出来而不是留半套。

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/dev/testutil"
	"github.com/luyaogod/TT/internal/devdata"
)

// snapshotTree 把一棵目录里的每个文件折成 rel → sha256（断言"零写盘"用）。
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return rerr
		}
		out[filepath.ToSlash(rel)] = devdata.BytesSHA256(b)
		return nil
	})
	if err != nil {
		t.Fatalf("快照 %s: %v", root, err)
	}
	return out
}

// assertSameSnapshot 断言两次快照逐文件一致（--check / 无变化分支必须零写盘）。
func assertSameSnapshot(t *testing.T, before, after map[string]string, why string) {
	t.Helper()
	for rel, sha := range before {
		got, ok := after[rel]
		if !ok {
			t.Errorf("%s：文件被删了 %s", why, rel)
			continue
		}
		if got != sha {
			t.Errorf("%s：文件被改了 %s", why, rel)
		}
	}
	for rel := range after {
		if _, ok := before[rel]; !ok {
			t.Errorf("%s：多出了文件 %s", why, rel)
		}
	}
}

// bumpServer4GL 在假服务器上把 .4gl 换成新内容并变新 mtime（造"服务器有新规格"）。
func bumpServer4GL(c *chain) string {
	newG4 := testutil.Normal4GL(fixProg) + "// 服务器改过 v2\n"
	c.src.put("erp/axm/4gl/"+fixProg+".4gl", newG4, fakeBaseTime.Add(2*time.Hour))
	return newG4
}

// TestUpdate_NoChangeBranch: 服务器与本地都没变时，Update 只回报"什么都没改"，
// 且**一个字节都不写**（这是"刷新"最容易被写成"每次都重写"的地方）。
func TestUpdate_NoChangeBranch(t *testing.T) {
	c := openViaSource(t)
	before := snapshotTree(t, c.Root())

	res, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !res.Checked {
		t.Error("Checked 未置位")
	}
	if res.Stale || len(res.StaleReasons) != 0 {
		t.Errorf("什么都没变却报 stale: %+v", res.StaleReasons)
	}
	if res.Fetch != nil || res.IndexRebuilt || res.Repacked || res.Reexported {
		t.Errorf("无变化分支动了东西: %+v", res)
	}
	if !strings.Contains(strings.Join(res.Notes, "；"), "材料与索引都是最新的") {
		t.Errorf("没有'最新'提示: %v", res.Notes)
	}
	if res.Status == nil {
		t.Error("Update 没回状态")
	}
	assertSameSnapshot(t, before, snapshotTree(t, c.Root()), "无变化分支")
}

// TestUpdate_RepullsRebuildsAndReexports: 服务器有新规格时，三条动作要一起发生：
// 拉材料 → 重建索引 → 重新打包 + 重新导出视图（视图干净、包未 apply 的正常情形）。
// 顺带钉住 Fetch 的告警透传与 task.json 的 update 事件。
func TestUpdate_RepullsRebuildsAndReexports(t *testing.T) {
	c := openViaSource(t)
	newG4 := bumpServer4GL(c)
	c.src.resetLog()

	res, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !res.Checked || !res.Stale {
		t.Fatalf("服务器变了却没报 stale: %+v", res)
	}
	if len(res.StaleReasons) == 0 || res.StaleReasons[0].Code != "remote-newer" {
		t.Errorf("stale 原因 = %+v，想要 remote-newer", res.StaleReasons)
	}
	if res.Fetch == nil || len(res.Fetch.Files) != 4 {
		t.Fatalf("没重拉材料: %+v", res.Fetch)
	}
	if !res.IndexRebuilt || !res.Repacked || !res.Reexported {
		t.Errorf("索引/包/视图没重建: index=%v pack=%v export=%v",
			res.IndexRebuilt, res.Repacked, res.Reexported)
	}
	// 材料真的换成了服务器上的新字节
	p := c.P()
	got, rerr := os.ReadFile(filepath.Join(p.Materials(), "src", fixProg+".4gl"))
	if rerr != nil || string(got) != newG4 {
		t.Errorf("材料 .4gl 没更新: %v / %q", rerr, string(got))
	}
	if !strings.Contains(strings.Join(c.src.readLog(), "\n"), "erp/axm/4gl/"+fixProg+".4gl") {
		t.Errorf("Update 没从假服务器读 .4gl: %v", c.src.readLog())
	}
	// 包与视图整体对齐：新包与视图基线一致，索引新鲜，视图干净
	st := c.status()
	if !st.Pkg.MatchesBase || st.Ws.Dirty || !st.Index.Fresh {
		t.Errorf("刷新后状态 = pkg:%+v ws:%+v index:%+v", st.Pkg, st.Ws, st.Index)
	}
	// task.json 的 update 事件要把"做了什么"说清
	task := c.Task()
	last := task.Events[len(task.Events)-1]
	if last.Action != "update" ||
		!strings.Contains(last.Detail, "材料已更新") ||
		!strings.Contains(last.Detail, "包与视图已重建") {
		t.Errorf("update 事件 = %+v", last)
	}
}

// TestUpdate_KeepsAppliedResults: 包里已有 apply 过的成果时，刷新只能动材料与索引 ——
// 包与视图必须原样保留（重建会把任务成果冲掉）。
func TestUpdate_KeepsAppliedResults(t *testing.T) {
	c := openViaSource(t)
	p := c.P()
	// 模拟 apply：包被写进一版改动，再 relink 把它认作新基线（apply 后基线会跟着更新）
	if _, err := testutil.WritePackage(p.PkgDir(), fixProg+"(c).tzc",
		testutil.NormalEntriesWith(fixProg, map[string]string{"env": "s"})); err != nil {
		t.Fatalf("覆写包: %v", err)
	}
	if _, err := Relink(RelinkRequest{Dir: c.Root()}); err != nil {
		t.Fatalf("Relink: %v", err)
	}
	st := c.status()
	if !st.Pkg.Modified || !st.Pkg.MatchesBase {
		t.Fatalf("夹具状态不对（要已 apply 且与基线一致）: %+v", st.Pkg)
	}
	pkgSha := st.Pkg.Sha256
	wsBefore := snapshotTree(t, p.Ws())

	bumpServer4GL(c)
	res, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !res.Checked || !res.Stale || !res.IndexRebuilt {
		t.Errorf("材料/索引该更新: checked=%v stale=%v index=%v", res.Checked, res.Stale, res.IndexRebuilt)
	}
	if res.Repacked || res.Reexported {
		t.Errorf("有 apply 成果时不该重建包/视图: %+v", res)
	}
	if !strings.Contains(strings.Join(res.Warnings, "；"), "视图或包里已有改动") {
		t.Errorf("没提示成果被保留: %v", res.Warnings)
	}
	// 包与视图逐字节不变
	if got, err := ReadPkgSha(p); err != nil || got != pkgSha {
		t.Errorf("包被改了: %q vs %q (%v)", got, pkgSha, err)
	}
	assertSameSnapshot(t, wsBefore, snapshotTree(t, p.Ws()), "apply 成果保留")
	st2 := c.status()
	if !st2.Pkg.Modified || !st2.Pkg.MatchesBase {
		t.Errorf("刷新后包状态变了: %+v", st2.Pkg)
	}
}

// ReadPkgSha 算包的 sha（测试辅助；生产侧 pkgSha 是包内私有）。
func ReadPkgSha(p Paths) (string, error) {
	sha, _ := pkgSha(p.Pkg(fixProg, "c"))
	return sha, nil
}

// TestUpdate_ForceOnlyRefreshesMaterialsAndIndex: 有未落盘改动时普通 update 拒绝；
// --force 只刷材料与索引，视图（含未落盘改动）与包原样保留。
func TestUpdate_ForceOnlyRefreshesMaterialsAndIndex(t *testing.T) {
	c := openViaSource(t)
	p := c.P()
	wsFile := filepath.Join(p.Ws(), "prog.full.4gl")
	base, err := os.ReadFile(wsFile)
	if err != nil {
		t.Fatal(err)
	}
	edited := append(append([]byte(nil), base...), []byte("\n# 本地改动\n")...)
	if err := os.WriteFile(wsFile, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	newG4 := bumpServer4GL(c)

	if _, err := Update(context.Background(), UpdateRequest{Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src}); !errors.Is(err, ErrDirtyWorkspace) {
		t.Fatalf("脏工作区普通 update: err = %v，想要 ErrDirtyWorkspace", err)
	}

	res, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src, Force: true,
	})
	if err != nil {
		t.Fatalf("Update --force: %v", err)
	}
	if res.Fetch == nil || !res.IndexRebuilt {
		t.Errorf("--force 该刷材料与索引: fetch=%v index=%v", res.Fetch, res.IndexRebuilt)
	}
	if res.Repacked || res.Reexported {
		t.Errorf("--force 不该重建视图/包: %+v", res)
	}
	if !strings.Contains(strings.Join(res.Warnings, "；"), "视图或包里已有改动") {
		t.Errorf("没提示视图保留: %v", res.Warnings)
	}
	// 未落盘改动还在（没有被重导冲掉）
	after, err := os.ReadFile(wsFile)
	if err != nil || !strings.Contains(string(after), "# 本地改动") {
		t.Errorf("--force 把未落盘改动冲掉了: %v", err)
	}
	// 材料确实更新了
	if got, err := os.ReadFile(filepath.Join(p.Materials(), "src", fixProg+".4gl")); err != nil || string(got) != newG4 {
		t.Errorf("材料没更新: %v", err)
	}
	if st := c.status(); !st.Ws.Dirty {
		t.Error("--force 后视图应仍是脏的")
	}
}

// TestUpdate_CheckOnlyWritesNothing: --check 承诺只核对不写盘。用整棵工作区的文件哈希
// 前后对比来兑现这个承诺（服务器已变、视图还脏，正是最想核对的时候）。
func TestUpdate_CheckOnlyWritesNothing(t *testing.T) {
	c := openViaSource(t)
	bumpServer4GL(c)
	before := snapshotTree(t, c.Root())

	res, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src, CheckOnly: true,
	})
	if err != nil {
		t.Fatalf("Update --check: %v", err)
	}
	if !res.Checked || !res.Stale || len(res.StaleReasons) == 0 {
		t.Errorf("--check 应核对出新规格: %+v", res)
	}
	if res.Fetch != nil || res.IndexRebuilt || res.Repacked || res.Reexported {
		t.Errorf("--check 动了东西: %+v", res)
	}
	if !strings.Contains(strings.Join(res.Notes, "；"), "只核对") {
		t.Errorf("没有只核对提示: %v", res.Notes)
	}
	assertSameSnapshot(t, before, snapshotTree(t, c.Root()), "--check")

	// 脏视图里 --check 也必须放行，并且同样零写盘
	c2 := openViaSource(t)
	wsFile := filepath.Join(c2.P().Ws(), "prog.full.4gl")
	b, _ := os.ReadFile(wsFile)
	if err := os.WriteFile(wsFile, append(b, []byte("\n# 脏\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	bumpServer4GL(c2)
	before2 := snapshotTree(t, c2.Root())
	if _, err := Update(context.Background(), UpdateRequest{
		Dir: c2.Root(), CfgPath: c2.env.Cfg, Files: c2.src, CheckOnly: true,
	}); err != nil {
		t.Fatalf("脏工作区 --check 应放行: %v", err)
	}
	assertSameSnapshot(t, before2, snapshotTree(t, c2.Root()), "脏工作区 --check")
}

// TestUpdate_RebuildsStaleIndexOnly: 材料没变但索引该重建（缺/过期）时，只重建索引，
// 不重新拉料、不动视图与包。顺带钉住 task.json 写失败**只警告不失败**。
func TestUpdate_RebuildsStaleIndexOnly(t *testing.T) {
	c := openViaSource(t)
	p := c.P()
	if err := os.Remove(p.Index()); err != nil {
		t.Fatal(err)
	}
	wsBefore := snapshotTree(t, p.Ws())

	taskFile := p.TaskFile()
	if err := os.Chmod(taskFile, 0o444); err != nil {
		t.Fatalf("设只读: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(taskFile, 0o644) })

	res, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if res.Stale || res.Fetch != nil {
		t.Errorf("材料没变不该重拉: stale=%v fetch=%v", res.Stale, res.Fetch)
	}
	if !res.IndexRebuilt || res.Repacked || res.Reexported {
		t.Errorf("只该重建索引: index=%v pack=%v export=%v",
			res.IndexRebuilt, res.Repacked, res.Reexported)
	}
	if _, err := os.Stat(p.Index()); err != nil {
		t.Errorf("索引没重建: %v", err)
	}
	if !strings.Contains(strings.Join(res.Notes, "；"), "索引重建") {
		t.Errorf("没有索引重建提示: %v", res.Notes)
	}
	if !strings.Contains(strings.Join(res.Warnings, "；"), "task.json 没写成功") {
		t.Errorf("task.json 写失败应只警告: %v", res.Warnings)
	}
	assertSameSnapshot(t, wsBefore, snapshotTree(t, p.Ws()), "只重建索引时视图不动")
}

// TestUpdate_FullForcesFetch: --full 不比对，直接整包重拉；材料没变也会重新打包并导出，
// 让工作区整体对齐到服务器当前状态。
func TestUpdate_FullForcesFetch(t *testing.T) {
	c := openViaSource(t)
	res, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src, Full: true,
	})
	if err != nil {
		t.Fatalf("Update --full: %v", err)
	}
	if res.Stale {
		t.Errorf("服务器没变，不该报 stale: %+v", res.StaleReasons)
	}
	if res.Fetch == nil || !res.IndexRebuilt || !res.Repacked || !res.Reexported {
		t.Errorf("--full 应整链重建: %+v", res)
	}
}

// TestUpdate_ReportsFetchFailure: 重拉中途失败要原样冒出来（此刻材料目录已被替换到一半
// 由 devdata 的临时目录纪律保证，本用例只钉 Update 不吞错）。
func TestUpdate_ReportsFetchFailure(t *testing.T) {
	c := openViaSource(t)
	bumpServer4GL(c)
	c.src.failRead("erp/axm/4gl/"+fixProg+".4gl", errors.New("合成读失败"))
	_, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src,
	})
	if err == nil || !strings.Contains(err.Error(), "合成读失败") {
		t.Fatalf("err = %v，想要把 Fetch 的错误带出来", err)
	}
}

// TestUpdate_ReportsPkgDirFailure: 重建包这一步的前提（pkg/ 是目录）被破坏时要报
// "建包目录失败"，不能把目录路径当包文件名写出去。
func TestUpdate_ReportsPkgDirFailure(t *testing.T) {
	c := openViaSource(t)
	p := c.P()
	bumpServer4GL(c)
	if err := os.RemoveAll(p.PkgDir()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.PkgDir(), []byte("不是目录"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src,
	})
	if err == nil || !strings.Contains(err.Error(), "建包目录失败") {
		t.Fatalf("err = %v，想要建包目录失败", err)
	}
}

// TestUpdate_ReportsIndexError: 重拉回来的 .tap 解析不了时，重建索引要失败并冒出来
// （此时材料已被替换成坏的那份，绝不能当成功往下走）。
func TestUpdate_ReportsIndexError(t *testing.T) {
	c := openViaSource(t)
	c.src.put("erp/axm/dzx/tap/"+fixProg+".tap", "<不是合法 XML", fakeBaseTime.Add(2*time.Hour))
	_, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src,
	})
	if err == nil {
		t.Fatal("坏 .tap 重建索引应当失败")
	}
	if !strings.Contains(err.Error(), "XML") {
		t.Errorf("err = %v，想要 .tap 解析错误", err)
	}
}

// TestUpdate_WarnsWhenMdlMissingOnRefresh: 环境里没镜像模板切片时，刷新只降级成一条提示
// （少一条脱版检查），不阻断材料/索引更新。
func TestUpdate_WarnsWhenMdlMissingOnRefresh(t *testing.T) {
	c := openViaSource(t)
	if err := os.RemoveAll(filepath.Join(c.env.DevData, "e", "mdl")); err != nil {
		t.Fatal(err)
	}
	bumpServer4GL(c)
	res, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if res.Fetch == nil {
		t.Error("没重拉材料")
	}
	if !strings.Contains(strings.Join(res.Warnings, "；"), "模板切片没镜像") {
		t.Errorf("应提示没镜像（跳过脱版检查）: %v", res.Warnings)
	}
}

// TestUpdate_ReportsRemoveViewFailure: 旧视图删不掉时必须报错 —— 不能在半旧半新的
// 视图上继续导出。用"当前工作目录删不掉"稳定制造这个失败（Windows 上进程 cwd 不可删）。
func TestUpdate_ReportsRemoveViewFailure(t *testing.T) {
	c := openViaSource(t)
	bumpServer4GL(c)
	t.Chdir(c.P().Ws())
	_, err := Update(context.Background(), UpdateRequest{
		Dir: c.Root(), CfgPath: c.env.Cfg, Files: c.src,
	})
	if err == nil || !strings.Contains(err.Error(), "删旧视图失败") {
		t.Fatalf("err = %v，想要删旧视图失败", err)
	}
}
