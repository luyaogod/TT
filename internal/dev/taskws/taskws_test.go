package taskws

// 工作区的状态与闸门。这几条测试的共同点：**不联网**。
//
// 能联网的路径（Create/Update 的拉料）留给端到端跑真环境；这里只钉住三件不依赖网络、
// 但错了代价很大的事：
//
//  1. 目录守卫：已存在的非空目录不许覆盖，非法任务名不进目录名。
//  2. 状态全现算：脏/已 apply/索引新鲜度都从文件推，不从 task.json 读。
//  3. update 的前置闸门：视图有未落盘改动时，**在联网之前**就拒绝。

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/store"
	"github.com/luyaogod/TT/internal/dev/testutil"
)

// writeTask 写一份最小 task.json。
func writeTask(t *testing.T, root, prog, env string) *Task {
	t.Helper()
	task := &Task{
		Schema: TaskSchema, Prog: prog, Env: env, Module: "cxm", Identity: "c",
		CreatedAt: nowStamp(), Layout: DefaultLayout(),
	}
	if err := Save(task, Paths{Root: root}); err != nil {
		t.Fatalf("写 task.json: %v", err)
	}
	return task
}

// writeMaterials 写一份最小材料目录（manifest 为空清单）。
func writeMaterials(t *testing.T, root, prog string) {
	t.Helper()
	dir := filepath.Join(root, DirMaterials)
	if err := os.MkdirAll(filepath.Join(dir, "dzx"), 0o755); err != nil {
		t.Fatalf("建材料目录: %v", err)
	}
	mf := map[string]any{
		"schema": "tt.devdata/v1", "kind": "fetch", "env": "e", "prog": prog,
		"fetchedAt": "2026-11-21T20:00:00+08:00", "module": "cxm",
		"files": []any{}, "missing": []any{}, "warnings": []any{},
	}
	b, _ := json.MarshalIndent(mf, "", " ")
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), b, 0o644); err != nil {
		t.Fatalf("写 manifest: %v", err)
	}
}

func TestLoad_MissingTaskFileSaysHowToCreate(t *testing.T) {
	root := t.TempDir()
	_, _, err := Load(root)
	if !errors.Is(err, ErrNoWorkspace) {
		t.Fatalf("err = %v，想要 ErrNoWorkspace", err)
	}
	if !strings.Contains(err.Error(), "tt dev open") {
		t.Errorf("提示里没有下一步: %v", err)
	}
}

func TestLoad_RejectsUnknownSchema(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, TaskFileName), []byte(`{"schema":"tt.dev.ws/v9","prog":"p"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := Load(root)
	if err == nil || !strings.Contains(err.Error(), "schema") {
		t.Fatalf("err = %v，想要 schema 不匹配", err)
	}
}

func TestCreate_RefusesNonEmptyDir(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "别动我.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Create(context.Background(), CreateRequest{
		Env: "e", Prog: "axmt500_wf", Dir: root,
	})
	if !errors.Is(err, ErrWorkspaceExists) {
		t.Fatalf("err = %v，想要 ErrWorkspaceExists", err)
	}
	// 原文件必须还在（拒绝就是拒绝，不能清场）
	if _, serr := os.Stat(filepath.Join(root, "别动我.txt")); serr != nil {
		t.Errorf("非空目录被动了: %v", serr)
	}
}

func TestCreate_RejectsIllegalTaskName(t *testing.T) {
	for _, name := range []string{"../逃出去", "a/b", `x\y`, "a..b"} {
		_, err := Create(context.Background(), CreateRequest{
			Env: "e", Prog: "axmt500_wf", Name: name, Dir: filepath.Join(t.TempDir(), "w"),
		})
		if !errors.Is(err, ErrTaskName) {
			t.Errorf("任务名 %q: err = %v，想要 ErrTaskName", name, err)
		}
	}
}

func TestStatus_ComputesDirtyAndAppliedLive(t *testing.T) {
	root := t.TempDir()
	prog := "axmt500_wf"
	writeTask(t, root, prog, "e")
	writeMaterials(t, root, prog)
	p := Paths{Root: root}

	// 包：用 testutil 造一个合法 .tzc，放到 pkg/ 下。
	if err := os.MkdirAll(p.PkgDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	built, err := testutil.WritePackage(p.PkgDir(), prog+"(c).tzc", testutil.NormalEntries(prog))
	if err != nil {
		t.Fatalf("造包: %v", err)
	}
	// 视图：从这份包导出。
	if _, err := Export(ExportRequest{Dir: p.Ws(), Pkg: built}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	st, err := GetStatus(StatusRequest{Dir: root})
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if !st.Ws.Present || st.Ws.Dirty {
		t.Errorf("视图应存在且干净: %+v", st.Ws)
	}
	if st.Ws.SectionState != "Locked" {
		t.Errorf("框架状态 = %q，想要 Locked", st.Ws.SectionState)
	}
	if !st.Pkg.Present || st.Pkg.Sha256 == "" {
		t.Errorf("包状态 = %+v", st.Pkg)
	}
	if st.Pkg.Modified {
		t.Error("刚导出的包不该算 apply 过")
	}
	if st.Materials.FetchedAt == "" {
		t.Errorf("材料时间没读到: %+v", st.Materials)
	}

	// 1) 改视图 → 脏
	edited, err := os.ReadFile(filepath.Join(p.Ws(), "prog.full.4gl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Ws(), "prog.full.4gl"), append(edited, []byte("\n# 本地随手一行\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err = GetStatus(StatusRequest{Dir: root})
	if err != nil {
		t.Fatalf("GetStatus(脏): %v", err)
	}
	if !st.Ws.Dirty {
		t.Error("改了视图却没判脏")
	}

	// 2) update 必须在联网之前拒绝（CfgPath 为空，真去联网会报环境不存在）
	if _, uerr := Update(context.Background(), UpdateRequest{Dir: root}); !errors.Is(uerr, ErrDirtyWorkspace) {
		t.Fatalf("update err = %v，想要 ErrDirtyWorkspace", uerr)
	}

	// 3) 包被换掉（相当于 apply 过）→ Applied，且视图恢复干净也不影响这个判断
	if err := os.WriteFile(filepath.Join(p.Ws(), "prog.full.4gl"), edited, 0o644); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(root, "other.tzc")
	if _, err := testutil.WritePackage(root, "other.tzc", testutil.NormalEntriesWith(prog, map[string]string{"env": "s"})); err != nil {
		t.Fatalf("造第二个包: %v", err)
	}
	raw, err := os.ReadFile(other)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p.Pkg(prog, "c"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	st, err = GetStatus(StatusRequest{Dir: root})
	if err != nil {
		t.Fatalf("GetStatus(apply 后): %v", err)
	}
	if st.Ws.Dirty {
		t.Error("恢复原文后不该还判脏")
	}
	if !st.Pkg.Modified {
		t.Error("包与视图基线不一致时应判为已 apply")
	}
}

func TestCountSliceTemplates(t *testing.T) {
	dir := t.TempDir()
	if _, ok := countSliceTemplates(dir); ok {
		t.Error("没有 slice/ 时应报不存在")
	}
	if err := os.MkdirAll(filepath.Join(dir, "slice"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"code_a05.template", "code_a17.template", "readme.txt"} {
		if err := os.WriteFile(filepath.Join(dir, "slice", n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	n, ok := countSliceTemplates(dir)
	if !ok || n != 2 {
		t.Errorf("切片数 = %d / %v，想要 2 / true", n, ok)
	}
}

// setManifestPkgPath 直接改工作区清单里的源包路径（模拟"工作区被复制/移动"）。
func setManifestPkgPath(t *testing.T, wsDir, pkgPath string) {
	t.Helper()
	p := filepath.Join(wsDir, "manifest.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var m store.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	m.Pkg.Path = pkgPath
	out, _ := json.MarshalIndent(m, "", " ")
	if err := os.WriteFile(p, out, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRebindPackage_RepointsToCurrentWorkspace(t *testing.T) {
	root := t.TempDir()
	prog := "axmt500_wf"
	writeTask(t, root, prog, "e")
	writeMaterials(t, root, prog)
	p := Paths{Root: root}
	if err := os.MkdirAll(p.PkgDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	built, err := testutil.WritePackage(p.PkgDir(), prog+"(c).tzc", testutil.NormalEntries(prog))
	if err != nil {
		t.Fatalf("造包: %v", err)
	}
	if _, err := Export(ExportRequest{Dir: p.Ws(), Pkg: built}); err != nil {
		t.Fatalf("Export: %v", err)
	}
	// 复制/移动之后：清单里的绝对路径还指着别处
	elsewhere := filepath.Join(root, "old", prog+"(c).tzc")
	setManifestPkgPath(t, p.Ws(), elsewhere)

	local, rebased, rerr := RebindPackage(root)
	if rerr != nil {
		t.Fatalf("RebindPackage: %v", rerr)
	}
	if !rebased || local != p.Pkg(prog, "c") {
		t.Fatalf("local=%q rebased=%v", local, rebased)
	}
	// 绑完之后 status 必须说"一致"
	st, serr := GetStatus(StatusRequest{Dir: root})
	if serr != nil {
		t.Fatalf("GetStatus: %v", serr)
	}
	if !st.Pkg.MatchesBase {
		t.Errorf("绑回本地包后仍报不一致: %+v", st.Pkg)
	}
}

func TestRebindPackage_RefusesWhenLocalPkgDiffers(t *testing.T) {
	root := t.TempDir()
	prog := "axmt500_wf"
	writeTask(t, root, prog, "e")
	writeMaterials(t, root, prog)
	p := Paths{Root: root}
	if err := os.MkdirAll(p.PkgDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	built, err := testutil.WritePackage(p.PkgDir(), prog+"(c).tzc", testutil.NormalEntries(prog))
	if err != nil {
		t.Fatalf("造包: %v", err)
	}
	if _, err := Export(ExportRequest{Dir: p.Ws(), Pkg: built}); err != nil {
		t.Fatalf("Export: %v", err)
	}
	// 本地包被换成另一份（与视图基线 sha 不同）
	other, err := testutil.WritePackage(root, "other.tzc", testutil.NormalEntriesWith(prog, map[string]string{"env": "s"}))
	if err != nil {
		t.Fatalf("造第二个包: %v", err)
	}
	raw, _ := os.ReadFile(other)
	if err := os.WriteFile(p.Pkg(prog, "c"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	setManifestPkgPath(t, p.Ws(), filepath.Join(root, "old", "x.tzc"))

	if _, _, rerr := RebindPackage(root); !errors.Is(rerr, ErrPackageMismatch) {
		t.Fatalf("err = %v，想要 ErrPackageMismatch", rerr)
	}
	// status 也要报出来（不然就是"静默可交付"）
	st, serr := GetStatus(StatusRequest{Dir: root})
	if serr != nil {
		t.Fatalf("GetStatus: %v", serr)
	}
	if st.Pkg.MatchesBase {
		t.Error("包与基线不同却报一致")
	}
	if !strings.Contains(strings.Join(st.Notes, "；"), "基线不一致") {
		t.Errorf("状态里没有提示: %v", st.Notes)
	}
}

func TestRelink_AdoptsCurrentPackage(t *testing.T) {
	root := t.TempDir()
	prog := "axmt500_wf"
	writeTask(t, root, prog, "e")
	writeMaterials(t, root, prog)
	p := Paths{Root: root}
	if err := os.MkdirAll(p.PkgDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	built, err := testutil.WritePackage(p.PkgDir(), prog+"(c).tzc", testutil.NormalEntries(prog))
	if err != nil {
		t.Fatalf("造包: %v", err)
	}
	if _, err := Export(ExportRequest{Dir: p.Ws(), Pkg: built}); err != nil {
		t.Fatalf("Export: %v", err)
	}
	// 包被换成另一份（正是 relink 要解的死局）
	other, err := testutil.WritePackage(root, "other.tzc", testutil.NormalEntriesWith(prog, map[string]string{"env": "s"}))
	if err != nil {
		t.Fatalf("造第二个包: %v", err)
	}
	raw, _ := os.ReadFile(other)
	if err := os.WriteFile(p.Pkg(prog, "c"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	st, _ := GetStatus(StatusRequest{Dir: root})
	if st.Pkg.MatchesBase {
		t.Fatal("换包后仍报一致")
	}
	res, rerr := Relink(RelinkRequest{Dir: root})
	if rerr != nil {
		t.Fatalf("Relink: %v", rerr)
	}
	if res.PkgSha == "" {
		t.Fatal("relink 没报新基线 sha")
	}
	st, _ = GetStatus(StatusRequest{Dir: root})
	if !st.Pkg.MatchesBase {
		t.Errorf("relink 后仍不一致: %+v", st.Pkg)
	}
	if st.Pkg.Sha256 != res.PkgSha {
		t.Errorf("基线 sha 没跟着走: %s vs %s", st.Pkg.Sha256, res.PkgSha)
	}
}

func TestRelink_RefusesDirtyWithoutForce(t *testing.T) {
	root := t.TempDir()
	prog := "axmt500_wf"
	writeTask(t, root, prog, "e")
	writeMaterials(t, root, prog)
	p := Paths{Root: root}
	if err := os.MkdirAll(p.PkgDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	built, err := testutil.WritePackage(p.PkgDir(), prog+"(c).tzc", testutil.NormalEntries(prog))
	if err != nil {
		t.Fatalf("造包: %v", err)
	}
	if _, err := Export(ExportRequest{Dir: p.Ws(), Pkg: built}); err != nil {
		t.Fatalf("Export: %v", err)
	}
	edited, err := os.ReadFile(filepath.Join(p.Ws(), "prog.full.4gl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Ws(), "prog.full.4gl"), append(edited, []byte("\n# 脏\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Relink(RelinkRequest{Dir: root}); !errors.Is(err, ErrDirtyWorkspace) {
		t.Fatalf("err = %v，想要 ErrDirtyWorkspace", err)
	}
}

func TestUpdate_RefusesPackageMismatch(t *testing.T) {
	root := t.TempDir()
	prog := "axmt500_wf"
	writeTask(t, root, prog, "e")
	writeMaterials(t, root, prog)
	p := Paths{Root: root}
	if err := os.MkdirAll(p.PkgDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	built, err := testutil.WritePackage(p.PkgDir(), prog+"(c).tzc", testutil.NormalEntries(prog))
	if err != nil {
		t.Fatalf("造包: %v", err)
	}
	if _, err := Export(ExportRequest{Dir: p.Ws(), Pkg: built}); err != nil {
		t.Fatalf("Export: %v", err)
	}
	other, err := testutil.WritePackage(root, "other.tzc", testutil.NormalEntriesWith(prog, map[string]string{"env": "s"}))
	if err != nil {
		t.Fatalf("造第二个包: %v", err)
	}
	raw, _ := os.ReadFile(other)
	if err := os.WriteFile(p.Pkg(prog, "c"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	// CfgPath 为空：如果它先去联网，会报环境错；先报 ErrPackageMismatch 说明拦在了前面
	if _, uerr := Update(context.Background(), UpdateRequest{Dir: root}); !errors.Is(uerr, ErrPackageMismatch) {
		t.Fatalf("err = %v，想要 ErrPackageMismatch", uerr)
	}
}
