package taskws

// Relink 与 Export 的测试。
//
// 这两条都是"包 ↔ 视图"的搬运：
//
//   - Export：包 → 围栏工作区（tzc 与工作区共用同一实现，差一个字符就会让两条命令
//     产出不同视图）。这里钉住它的失败面：包打不开、TAP 解析失败、拒绝覆盖已有工作区、
//     git 不可用时的降级。
//   - Relink：把"包被换过、工作区成死局"拉回来 —— 以当前包为新基线重导视图。
//     钉住：没包拒绝、包坏拒绝、脏视图默认拒绝、--force 会丢改动、事件要落进 task.json。

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/model"
	"github.com/luyaogod/TT/internal/dev/store"
	"github.com/luyaogod/TT/internal/dev/testutil"
)

// TestRelink_LoadError: 目录不是工作区时 relink 报 ErrNoWorkspace。
func TestRelink_LoadError(t *testing.T) {
	if _, err := Relink(RelinkRequest{Dir: t.TempDir()}); !errors.Is(err, ErrNoWorkspace) {
		t.Fatalf("err = %v，想要 ErrNoWorkspace", err)
	}
}

// TestRelink_NoPackage: 工作区里没有包就没有"新基线"，必须拒绝并说清放回哪。
func TestRelink_NoPackage(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(f.Pkg); err != nil {
		t.Fatal(err)
	}
	_, err := Relink(RelinkRequest{Dir: f.Root})
	if !errors.Is(err, ErrPackageMismatch) {
		t.Fatalf("err = %v，想要 ErrPackageMismatch", err)
	}
	if !strings.Contains(err.Error(), "没有包") {
		t.Fatalf("提示没说清本地没有包: %v", err)
	}
	// 视图还在（relink 没动手）
	if _, serr := os.Stat(filepath.Join(f.P.Ws(), "prog.full.4gl")); serr != nil {
		t.Errorf("拒绝时不该删视图: %v", serr)
	}
}

// TestRelink_CorruptPackage: 包字节在（sha 非空）但不是合法 .tzc 时，
// 重导视图会失败 —— 此时工作区仍处于死局，错误必须冒出来而不是留一个空视图。
func TestRelink_CorruptPackage(t *testing.T) {
	f := newFixture(t)
	if err := os.WriteFile(f.Pkg, []byte("不是 zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Relink(RelinkRequest{Dir: f.Root})
	if err == nil {
		t.Fatal("坏包时 relink 应当失败")
	}
}

// TestRelink_ForceDropsDirtyEdits: 脏视图默认拒绝（改动会丢）；--force 才按包重导，
// 并且改动**确实**被丢掉 —— 这是"代价写在文档里"的那条纪律。
func TestRelink_ForceDropsDirtyEdits(t *testing.T) {
	f := newFixture(t)
	f.dirtyView()
	if _, err := Relink(RelinkRequest{Dir: f.Root}); !errors.Is(err, ErrDirtyWorkspace) {
		t.Fatalf("脏视图默认应报 ErrDirtyWorkspace: %v", err)
	}

	res, err := Relink(RelinkRequest{Dir: f.Root, Force: true})
	if err != nil {
		t.Fatalf("Relink --force: %v", err)
	}
	if res.Pkg != f.Pkg || res.PkgSha == "" {
		t.Errorf("结果包 = %q sha=%q", res.Pkg, res.PkgSha)
	}
	if res.Regions == 0 {
		t.Error("重导后区域数不应为 0")
	}
	// 改动丢了 = 重导出来的视图里没有那行
	edited, err := os.ReadFile(filepath.Join(f.P.Ws(), "prog.full.4gl"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(edited), "本地随手一行") {
		t.Error("--force 后未落盘改动仍在（重导没生效）")
	}
	// 重导后视图与包一致，且工作区不再脏
	st := f.status()
	if st.Ws.Dirty || !st.Pkg.MatchesBase {
		t.Errorf("relink 后状态 = ws:%+v pkg:%+v", st.Ws, st.Pkg)
	}
	// 事件落进 task.json（交接时要能回答"这视图怎么来的"）
	task := f.loadTask()
	if len(task.Events) == 0 || task.Events[len(task.Events)-1].Action != "relink" {
		t.Errorf("没记 relink 事件: %+v", task.Events)
	}
	if st.Pkg.Sha256 != res.PkgSha {
		t.Errorf("status 的基线 sha 没跟上: %s vs %s", st.Pkg.Sha256, res.PkgSha)
	}
}

// TestRelink_AdoptsReplacedPackage: 包被换成另一份（apply 的"源包未变"校验会永久失败）
// —— relink 以它为新基线，status 从此说"一致"。
func TestRelink_AdoptsReplacedPackage(t *testing.T) {
	f := newFixture(t)
	f.writePkg(testutil.NormalEntriesWith(f.Prog, map[string]string{"env": "s"}))
	if st := f.status(); st.Pkg.MatchesBase {
		t.Fatal("换包后仍报一致")
	}
	res, err := Relink(RelinkRequest{Dir: f.Root})
	if err != nil {
		t.Fatalf("Relink: %v", err)
	}
	st := f.status()
	if !st.Pkg.MatchesBase {
		t.Errorf("relink 后仍不一致: %+v", st.Pkg)
	}
	if st.Pkg.Sha256 != res.PkgSha {
		t.Errorf("新基线 sha 没记进视图: %s vs %s", st.Pkg.Sha256, res.PkgSha)
	}
}

// TestExport_Errors: 导出器的四个失败面。关键在于"失败不落半套工作区"：
// 包打不开/解析失败时不能留下一个看起来可用的 ws/。
func TestExport_Errors(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(t.TempDir(), "ws")

	// 1) 包不存在
	if _, err := Export(ExportRequest{Dir: wsDir, Pkg: filepath.Join(root, "没有.tzc")}); err == nil {
		t.Error("包不存在时应当失败")
	}
	// 2) 包不是 zip
	badZip := filepath.Join(root, "坏(c).tzc")
	if err := os.WriteFile(badZip, []byte("不是 zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Export(ExportRequest{Dir: wsDir, Pkg: badZip}); err == nil {
		t.Error("坏 zip 应当失败")
	}
	// 3) TAP 解析失败（包结构合法、内容不是 XML）
	tapBad, err := testutil.WritePackage(root, fixProg+"(c).tzc", map[string]string{
		fixProg + ".tap": "<不是合法 XML",
		fixProg + ".tgl": "x\n",
		fixProg + ".4gl": "MAIN\nEND MAIN\n",
		"ver":            "1.0\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Export(ExportRequest{Dir: wsDir, Pkg: tapBad}); err == nil {
		t.Error("TAP 解析失败应当报错")
	}
	// 失败后没有留下工作区
	if _, serr := os.Stat(wsDir); !os.IsNotExist(serr) {
		t.Errorf("导出失败却留下了工作区: %v", serr)
	}

	// 4) 已有工作区：绝不静默覆盖
	f := newFixture(t)
	if _, err := Export(ExportRequest{Dir: f.P.Ws(), Pkg: f.Pkg}); err == nil {
		t.Error("已有工作区时应当拒绝覆盖")
	}
}

// TestExportWarnings_GitMissing: git 不在 PATH 时导出要降级（工作区仍可用）并给出
// "改动不便于回溯"的提示 —— 不能让"没有 git"变成建不了工作区。
func TestExportWarnings_GitMissing(t *testing.T) {
	t.Setenv("PATH", "") // 让 exec 找不到 git

	pkg, err := fixPackageAt(t, filepath.Join(t.TempDir(), "pkgs"), fixProg)
	if err != nil {
		t.Fatal(err)
	}
	wsDir := filepath.Join(t.TempDir(), "ws")
	res, err := Export(ExportRequest{Dir: wsDir, Pkg: pkg})
	if err != nil {
		t.Fatalf("git 不可用时 Export 应降级而不是失败: %v", err)
	}
	if res.Commit != "" {
		t.Errorf("git 不可用却报 commit = %q", res.Commit)
	}
	if res.Regions == 0 {
		t.Error("视图没导出")
	}
	warns := ExportWarnings(res)
	if len(warns) == 0 || !strings.Contains(strings.Join(warns, "；"), "git 提交失败") {
		t.Errorf("缺少 git 降级提示: %v", warns)
	}
	// 工作区文件必须齐全（降级的是版本留痕，不是导出本身）
	if _, serr := os.Stat(filepath.Join(wsDir, "prog.full.4gl")); serr != nil {
		t.Errorf("工作区没落盘: %v", serr)
	}
}

// TestExportWarnings_LockedAndUnlocked: 框架锁着时提示"改 section 正文要先 unlock"
// （单向、有代价）；包本身就是解锁态时不再重复提示。
func TestExportWarnings_LockedAndUnlocked(t *testing.T) {
	locked := &ExportResult{Commit: "abc", Section: store.SectionInfo{State: string(model.SectionLocked)}}
	warns := ExportWarnings(locked)
	if len(warns) != 1 || !strings.Contains(warns[0], "框架未解锁") {
		t.Errorf("Locked 提示 = %v", warns)
	}
	unlocked := &ExportResult{Commit: "abc", Section: store.SectionInfo{State: string(model.SectionUnlocked)}}
	if warns := ExportWarnings(unlocked); len(warns) != 0 {
		t.Errorf("Unlocked 不该有提示: %v", warns)
	}
	// 没有 commit 但框架解锁：只有 git 那条
	noCommit := &ExportResult{Commit: "", Section: store.SectionInfo{State: string(model.SectionUnlocked)}}
	if warns := ExportWarnings(noCommit); len(warns) != 1 || !strings.Contains(warns[0], "git 提交失败") {
		t.Errorf("只应报 git: %v", warns)
	}
}

// TestExport_PartialOnly: --only 是部分导出：只把点名的点渲染成可编辑。
// 导出结果与视图清单都要能回答"这次导的是哪几个点"。
func TestExport_PartialOnly(t *testing.T) {
	pkg, err := fixPackageAt(t, filepath.Join(t.TempDir(), "pkgs"), fixProg)
	if err != nil {
		t.Fatal(err)
	}
	only := "function." + fixProg + "_calc"
	res, err := Export(ExportRequest{Dir: filepath.Join(t.TempDir(), "only-ws", DirWs), Pkg: pkg, Only: []string{only}})
	if err != nil {
		t.Fatalf("Export --only: %v", err)
	}
	if res.Manifest == nil {
		t.Fatal("导出结果没带清单")
	}
	if len(res.Manifest.Export.Only) != 1 || res.Manifest.Export.Only[0] != only {
		t.Errorf("export.only = %v，想要 [%s]", res.Manifest.Export.Only, only)
	}
	if res.Regions == 0 {
		t.Error("部分导出也应有可编辑区域（至少 .4gl 骨架）")
	}
}

// TestRelink_TaskJSONUnwritableFailsLoudly: 视图重导成功但 task.json 写不回去时，
// relink 必须报错（历史没记上 = 交接时说不清这视图怎么来的）。Windows 上只读文件
// 会让原子替换失败。
func TestRelink_TaskJSONUnwritableFailsLoudly(t *testing.T) {
	f := newFixture(t)
	if err := os.Chmod(f.P.TaskFile(), 0o444); err != nil {
		t.Fatalf("设只读: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(f.P.TaskFile(), 0o644) })
	if _, err := Relink(RelinkRequest{Dir: f.Root}); err == nil {
		t.Fatal("task.json 写不回去时应报错")
	}
}

// TestRelink_ReportsRemoveViewFailure: 旧视图删不掉时必须报错 —— 不能在半旧半新的
// 视图上继续导出。用"当前工作目录删不掉"来稳定地制造这个失败（Windows 上进程 cwd
// 不能被 RemoveAll 删掉）。
func TestRelink_ReportsRemoveViewFailure(t *testing.T) {
	f := newFixture(t)
	t.Chdir(f.P.Ws())
	_, err := Relink(RelinkRequest{Dir: f.Root})
	if err == nil || !strings.Contains(err.Error(), "删旧视图失败") {
		t.Fatalf("err = %v，想要删旧视图失败", err)
	}
}
