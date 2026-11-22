package taskws

// GetStatus 的测试：状态**全部现算**（一个字段都不从 task.json 读），所以每条用例
// 都去改磁盘上的那一份真文件，再看 status 会不会跟着变。这是本包最容易悄悄说谎的
// 地方 —— 存进 task.json 的"是否脏/包 sha"第一次忘了刷新就开始骗人。
//
// 覆盖四件事：材料、索引新鲜度、视图（脏/存在）、包（apply 过/被换过/不存在）。

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/model"
	"github.com/luyaogod/TT/internal/dev/testutil"
)

// TestGetStatus_FullFixture: 刚建好的工作区（材料齐全 / 索引新鲜 / 视图干净 / 包未被
// 换过）应当是"一切对得上"。这条用例顺带把 status 的每个字段都对着夹具算一遍 ——
// 字段错位（比如把包字节数算成材料字节数）在别的用例里很难发现。
func TestGetStatus_FullFixture(t *testing.T) {
	f := newFixture(t)
	st := f.status()

	if st.Dir != f.Root || st.Prog != f.Prog || st.Env != "e" || st.Module != "cxm" {
		t.Errorf("身份字段 = %+v", st)
	}
	if st.Identity != "c" || st.Task != "epi" || st.Ent != 7 {
		t.Errorf("身份字段（identity/task/ent）= %+v", st)
	}
	if st.CreatedAt == "" || st.UpdatedAt == "" {
		t.Errorf("时间字段没读出来: created=%q updated=%q", st.CreatedAt, st.UpdatedAt)
	}

	// 材料：4 个文件（tab/tgl/tap/4gl），字节数与 manifest 合计一致
	if st.Materials.Files != len(f.Manifest.Files) || st.Materials.Files != 4 {
		t.Errorf("材料文件数 = %d，想要 4", st.Materials.Files)
	}
	if st.Materials.Bytes != f.Manifest.Bytes() {
		t.Errorf("材料字节 = %d，想要 %d", st.Materials.Bytes, f.Manifest.Bytes())
	}
	if st.Materials.FetchedAt != f.Manifest.FetchedAt || st.Materials.Module != "cxm" {
		t.Errorf("材料清单字段 = %+v", st.Materials)
	}
	if st.Materials.DZAA002 != 3 {
		t.Errorf("版次锚点 = %d，想要 3", st.Materials.DZAA002)
	}
	if len(st.Notes) != 0 {
		t.Errorf("一切对得上时不该有提示: %v", st.Notes)
	}

	// 索引：存在且新鲜，摘要与 link.json 一致
	if !st.Index.Present || !st.Index.Fresh {
		t.Fatalf("索引状态 = %+v，想要存在且新鲜", st.Index)
	}
	if st.Index.Path != f.P.Index() || st.Index.Schema != f.Doc.Schema || st.Index.GeneratedAt == "" {
		t.Errorf("索引抬头 = %+v", st.Index)
	}
	if st.Index.Points != f.Doc.Stats.Points || st.Index.Custom != f.Doc.Stats.PointsCustom {
		t.Errorf("索引点数 = %d/%d，想要 %d/%d", st.Index.Points, st.Index.Custom,
			f.Doc.Stats.Points, f.Doc.Stats.PointsCustom)
	}
	if st.Index.Sections != len(f.Doc.Sections) || st.Index.Slices != f.Doc.Slices.Distinct {
		t.Errorf("索引区段/切片 = %d/%d", st.Index.Sections, st.Index.Slices)
	}

	// 视图：存在、干净、框架锁着、区间数来自 ws/manifest.json
	if !st.Ws.Present || st.Ws.Dirty {
		t.Fatalf("视图状态 = %+v，想要存在且干净", st.Ws)
	}
	if st.Ws.SectionState != string(model.SectionLocked) {
		t.Errorf("框架状态 = %q，想要 Locked", st.Ws.SectionState)
	}
	wsMf, err := os.ReadFile(filepath.Join(f.P.Ws(), "manifest.json"))
	if err != nil || len(wsMf) == 0 {
		t.Fatalf("读视图清单: %v", err)
	}
	if st.Ws.PkgSha == "" {
		t.Error("视图没记源包 sha")
	}

	// 包：存在、与视图基线一致、没被 apply 过
	if !st.Pkg.Present || st.Pkg.Bytes == 0 || st.Pkg.Sha256 == "" {
		t.Fatalf("包状态 = %+v", st.Pkg)
	}
	if st.Pkg.Path != f.Pkg || st.Pkg.Path != f.P.Pkg(f.Prog, "c") {
		t.Errorf("包路径 = %q，想要 %q", st.Pkg.Path, f.Pkg)
	}
	if st.Pkg.Modified {
		t.Error("刚导出的包不该算 apply 过")
	}
	if !st.Pkg.MatchesBase {
		t.Errorf("刚导出的包应与基线一致: %+v", st.Pkg)
	}
	if st.Pkg.Sha256 != st.Ws.PkgSha {
		t.Errorf("包 sha 与视图基线 sha 不同: %s vs %s", st.Pkg.Sha256, st.Ws.PkgSha)
	}

	// 模板切片
	if !st.Mdl.Present || st.Mdl.Templates != 2 {
		t.Errorf("模板切片 = %+v，想要 2 片", st.Mdl)
	}

	// 人读定位
	if got, want := st.Describe(), f.Prog+" · epi（e）"; got != want {
		t.Errorf("Describe = %q，想要 %q", got, want)
	}
}

// TestGetStatus_AppliesVisibleInPkg: apply 之后包里的 .tap 与材料不同 → Modified；
// 同时包 sha 与视图基线不同 → MatchesBase=false 并且**必须**给出一条出路提示
// （否则就成了"静默可交付"，README 里记着实测事故）。
func TestGetStatus_AppliedVisibleInPkg(t *testing.T) {
	f := newFixture(t)
	// 换一份 TAP（改根属性 env → 字节变了，等价于"apply 落了一版改动"）
	f.writePkg(testutil.NormalEntriesWith(f.Prog, map[string]string{"env": "s"}))

	st := f.status()
	if !st.Pkg.Modified {
		t.Error("包里的 .tap 与材料不同，应判 apply 过")
	}
	if st.Pkg.MatchesBase {
		t.Error("包被换过却报与基线一致")
	}
	if st.Pkg.Sha256 == st.Ws.PkgSha {
		t.Error("sha 相同却判 apply 过，判据搞错了")
	}
	notes := strings.Join(st.Notes, "；")
	if !strings.Contains(notes, "包与视图基线不一致") {
		t.Errorf("没有不可交付提示: %v", st.Notes)
	}
	if !strings.Contains(notes, "tt dev relink") {
		t.Errorf("提示里没有出路（relink）: %v", st.Notes)
	}
	if st.Ws.Dirty {
		t.Error("只换包不该把视图判脏")
	}
}

// TestGetStatus_DirtyView: 视图与基线不同就是脏，并给出"先 apply/verify"的下一步。
func TestGetStatus_DirtyView(t *testing.T) {
	f := newFixture(t)
	f.dirtyView()
	st := f.status()
	if !st.Ws.Dirty {
		t.Fatal("改了视图却没判脏")
	}
	if !strings.Contains(strings.Join(st.Notes, "；"), "未落盘的改动") {
		t.Errorf("没有脏提示: %v", st.Notes)
	}
	// 视图脏不影响包与材料的判断
	if !st.Pkg.MatchesBase || st.Pkg.Modified {
		t.Errorf("脏视图不该改变包状态: %+v", st.Pkg)
	}
}

// TestGetStatus_StaleIndexAfterMaterialEdit: 材料变了索引没重建 → Fresh=false 且提示
// 重建；同时"包 vs 材料"的字节比对会立刻发现包落后（材料里能取到 tap/tgl 时按字节比，
// 而不是拿包 sha 与基线比）。
func TestGetStatus_StaleIndexAfterMaterialEdit(t *testing.T) {
	f := newFixture(t)
	if st := f.status(); !st.Index.Fresh {
		t.Fatalf("夹具初始索引应新鲜: %+v", st.Index)
	}
	f.editMaterialTap()
	st := f.status()
	if !st.Index.Present {
		t.Error("索引文件还在，Present 应为 true")
	}
	if st.Index.Fresh {
		t.Error("材料变过，索引应判过期")
	}
	if !strings.Contains(strings.Join(st.Notes, "；"), "索引与材料不一致") {
		t.Errorf("没有索引过期提示: %v", st.Notes)
	}
	if !st.Pkg.Modified {
		t.Error("材料 tap 变了、包没变 → 包应判落后于材料")
	}
}

// TestGetStatus_NoMaterialsManifest: 材料清单丢了（没拉过 / 被删）时状态要照常给出，
// 只是材料侧标空 + 三条提示；**不能**因为材料不完整就让 status 整体报错。
func TestGetStatus_NoMaterialsManifest(t *testing.T) {
	f := newFixture(t)
	f.removeMatManifest()
	st := f.status()
	if st.Materials.FetchedAt != "" || st.Materials.Files != 0 {
		t.Errorf("没有清单时材料字段应空: %+v", st.Materials)
	}
	notes := strings.Join(st.Notes, "；")
	for _, want := range []string{"没有材料清单", "索引新鲜度查不出来"} {
		if !strings.Contains(notes, want) {
			t.Errorf("缺少提示 %q: %v", want, st.Notes)
		}
	}
	// 索引文件本身还在 → Present=true，但新鲜度答不了 → Fresh=false
	if !st.Index.Present || st.Index.Fresh {
		t.Errorf("索引状态 = %+v", st.Index)
	}
}

// TestGetStatus_CorruptMaterialsManifest: 清单坏成非法 JSON 与"没有清单"要分开说
// （一个是材料被改坏，一个是从来没拉过）。
func TestGetStatus_CorruptMaterialsManifest(t *testing.T) {
	f := newFixture(t)
	f.corruptMatManifest()
	st := f.status()
	if !strings.Contains(strings.Join(st.Notes, "；"), "材料清单读不出来") {
		t.Errorf("坏清单被当成没清单: %v", st.Notes)
	}
	if !strings.Contains(strings.Join(st.Notes, "；"), "包内容与材料对比失败") {
		t.Errorf("清单是坏的（读出错）时包比对失败要留痕: %v", st.Notes)
	}
	if st.Materials.Files != 0 {
		t.Errorf("坏清单不该读出文件数: %+v", st.Materials)
	}
	if st.Index.Fresh {
		t.Error("清单读不出来时新鲜度应判 false")
	}
}

// TestGetStatus_CorruptIndex: link.json 坏掉时 Present=false（读不出来就是没有），
// 不阻塞其它三块状态。
func TestGetStatus_CorruptIndex(t *testing.T) {
	f := newFixture(t)
	if err := os.WriteFile(f.P.Index(), []byte("{坏"), 0o644); err != nil {
		t.Fatal(err)
	}
	st := f.status()
	if st.Index.Present {
		t.Error("坏索引不该报 Present")
	}
	if st.Index.Points != 0 {
		t.Errorf("坏索引不该有统计: %+v", st.Index)
	}
	if st.Materials.FetchedAt == "" || st.Materials.Files == 0 {
		t.Error("材料本身没动，不该被判坏")
	}
	if !st.Ws.Present {
		t.Error("索引坏掉不该影响视图状态")
	}
}

// TestGetStatus_MissingIndexFile: 从没建过索引时 Present=false 且不报"过期"
// （Present 都没有就谈不上过期）。
func TestGetStatus_MissingIndexFile(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(f.P.Index()); err != nil {
		t.Fatal(err)
	}
	st := f.status()
	if st.Index.Present || st.Index.Fresh {
		t.Errorf("没有索引文件: %+v", st.Index)
	}
	if strings.Contains(strings.Join(st.Notes, "；"), "索引与材料不一致") {
		t.Errorf("没有索引不该报不一致: %v", st.Notes)
	}
}

// TestGetStatus_ViewUnreadable: 视图目录里缺必需文件（export 没完成）时给一条
// "视图读不出来"，而不是安静地当没视图。
func TestGetStatus_ViewUnreadable(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(filepath.Join(f.P.Ws(), ".tdev", "regions.json")); err != nil {
		t.Fatal(err)
	}
	st := f.status()
	if st.Ws.Present {
		t.Error("缺必需文件的视图不该报 Present")
	}
	if !strings.Contains(strings.Join(st.Notes, "；"), "视图读不出来") {
		t.Errorf("没有视图损坏提示: %v", st.Notes)
	}
}

// TestGetStatus_CorruptViewManifest: 视图文件都在、manifest 内容坏了 → Present 为真，
// 但摘要字段取不到（不能把上一次的缓存值当现状）。
func TestGetStatus_CorruptViewManifest(t *testing.T) {
	f := newFixture(t)
	if err := os.WriteFile(filepath.Join(f.P.Ws(), "manifest.json"), []byte("{坏"), 0o644); err != nil {
		t.Fatal(err)
	}
	st := f.status()
	if !st.Ws.Present {
		t.Error("文件齐全时视图应报 Present")
	}
	if st.Ws.SectionState != "" || st.Ws.Regions != 0 || st.Ws.PkgSha != "" {
		t.Errorf("坏清单不该读出摘要: %+v", st.Ws)
	}
}

// TestGetStatus_NoPackageNoMdl: 包被删 / 模板切片没复制时两块状态各自标空，
// 且 identity 为空的老工作区按 c 兜底找包。
func TestGetStatus_NoPackageNoMdl(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(f.Pkg); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(f.P.Mdl()); err != nil {
		t.Fatal(err)
	}
	task := f.loadTask()
	task.Identity = ""
	if err := Save(task, f.P); err != nil {
		t.Fatal(err)
	}
	st := f.status()
	if st.Pkg.Present || st.Pkg.Sha256 != "" {
		t.Errorf("包删掉后仍报存在: %+v", st.Pkg)
	}
	if st.Pkg.Path != f.P.Pkg(f.Prog, "c") {
		t.Errorf("无身份工作区的包路径 = %q，想要 c 兜底", st.Pkg.Path)
	}
	if st.Mdl.Present || st.Mdl.Templates != 0 {
		t.Errorf("没有切片时报了存在: %+v", st.Mdl)
	}
	if !st.Ws.Present || st.Ws.PkgSha == "" {
		t.Errorf("视图状态不该受包删除影响: %+v", st.Ws)
	}
}

// TestDescribe_WithoutTask: 没记任务名时定位串只有"程序（环境）"。
func TestDescribe_WithoutTask(t *testing.T) {
	st := &Status{Prog: fixProg, Env: "e"}
	if got, want := st.Describe(), fixProg+"（e）"; got != want {
		t.Errorf("Describe = %q，想要 %q", got, want)
	}
}

// TestGetStatus_ManifestWithoutTapTgl: 材料清单里没有 tap/tgl 角色时，包改动判据要
// 退回到"包 sha ≠ 视图基线"（README：材料缺件时才退回 sha 比对），不能拿空角色去比字节。
func TestGetStatus_ManifestWithoutTapTgl(t *testing.T) {
	f := newFixture(t)
	f.replaceManifest(f.filesOfRole("tab", "4gl"))
	st := f.status()
	if st.Pkg.Modified {
		t.Error("清单里没有 tap/tgl 角色时，包与基线一致，不该判 apply 过")
	}
	if st.Materials.Files != 2 {
		t.Errorf("材料文件数 = %d，想要 2", st.Materials.Files)
	}
}

// TestGetStatus_ManifestRelFileGone: 清单记着 tap 的落点、但文件被删了 —— 读不到就跳过
// 这个角色（不能把"材料缺件"误判成"包里改过"）；tgl 仍能比，所以结论是不可信但不算脏。
func TestGetStatus_ManifestRelFileGone(t *testing.T) {
	f := newFixture(t)
	rel, err := f.materialRel("tap")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(f.P.Materials(), filepath.FromSlash(rel))); err != nil {
		t.Fatal(err)
	}
	st := f.status()
	if st.Pkg.Modified {
		t.Error("tap 文件读不到时应跳过该角色，而不是判 apply 过")
	}
	if strings.Contains(strings.Join(st.Notes, "；"), "包内容与材料对比失败") {
		t.Errorf("清单本身可读，不该报比对失败: %v", st.Notes)
	}
}

// TestGetStatus_SliceSubdirectorySkipped: 数模板切片时目录不算，只数 code_*.template 文件。
func TestGetStatus_SliceSubdirectorySkipped(t *testing.T) {
	f := newFixture(t)
	if err := os.MkdirAll(filepath.Join(f.P.Mdl(), "slice", "某目录"), 0o755); err != nil {
		t.Fatal(err)
	}
	st := f.status()
	if !st.Mdl.Present || st.Mdl.Templates != 2 {
		t.Errorf("切片统计 = %+v，想要 2（子目录不算）", st.Mdl)
	}
}

// TestGetStatus_LoadError: 目录不是工作区时状态查询报 ErrNoWorkspace。
func TestGetStatus_LoadError(t *testing.T) {
	if _, err := GetStatus(StatusRequest{Dir: t.TempDir()}); !errors.Is(err, ErrNoWorkspace) {
		t.Fatalf("err = %v，想要 ErrNoWorkspace", err)
	}
}

// TestPackagePathAndPkgPoint_LoadError: 两条读命令在非工作区目录上都要报
// ErrNoWorkspace（CLI 据此提示"用 tt dev open 建一个"）。
func TestPackagePathAndPkgPoint_LoadError(t *testing.T) {
	dir := t.TempDir()
	if _, err := PackagePath(dir); !errors.Is(err, ErrNoWorkspace) {
		t.Errorf("PackagePath err = %v，想要 ErrNoWorkspace", err)
	}
	if _, _, err := PkgPoint(dir, "x"); !errors.Is(err, ErrNoWorkspace) {
		t.Errorf("PkgPoint err = %v，想要 ErrNoWorkspace", err)
	}
}

// TestPackagePath_FallsBackToManifestPath: 本地 pkg/ 下没有包、但视图清单里记着源包路径时，
// PackagePath 必须像 PkgPoint 一样回退找到它 —— 否则 `search --scope pkg` 会静默搜不到，
// 而 `at --scope pkg` 又能看到（同一份工作区两个结论）。
func TestPackagePath_FallsBackToManifestPath(t *testing.T) {
	f := newFixture(t)
	elsewhere, err := fixPackageAt(t, filepath.Join(f.Root, "elsewhere"), f.Prog)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(f.Pkg); err != nil {
		t.Fatal(err)
	}
	setManifestPkgPath(t, f.P.Ws(), elsewhere)

	got, err := PackagePath(f.Root)
	if err != nil {
		t.Fatalf("PackagePath: %v", err)
	}
	if got != elsewhere {
		t.Errorf("回退路径 = %q，想要 %q", got, elsewhere)
	}
	// PkgPoint 与 PackagePath 口径一致
	_, pkgPointPath, perr := PkgPoint(f.Root, "global.memo")
	if perr != nil || pkgPointPath != elsewhere {
		t.Errorf("PkgPoint 回退 = %q / %v，想要与 PackagePath 相同", pkgPointPath, perr)
	}

	// 清单里记的路径也不存在时，仍然安静返回空串（不报错）
	setManifestPkgPath(t, f.P.Ws(), filepath.Join(f.Root, "nowhere.tzc"))
	if got, err := PackagePath(f.Root); err != nil || got != "" {
		t.Errorf("清单路径也不存在时 = %q / %v，想要空串且无错", got, err)
	}

	// 清单里记的是个目录：不算包（返回空串），别让消费者打开时才炸
	dirAsPkg := filepath.Join(f.Root, "not-a-pkg")
	if err := os.MkdirAll(dirAsPkg, 0o755); err != nil {
		t.Fatal(err)
	}
	setManifestPkgPath(t, f.P.Ws(), dirAsPkg)
	if got, err := PackagePath(f.Root); err != nil || got != "" {
		t.Errorf("清单指向目录时 = %q / %v，想要空串且无错", got, err)
	}
}
