package taskws

// open / load / save 与几个"纯本地"原语的测试。
//
// Create 的**拉料**一步必须连 SSH（devdata.Fetch 没有可注入的 FileSource 入口），
// 断网环境里只能钉它"进网络之前的全部判断"与"失败后不留半套"。
// 拉料之后的那半段（打包/导出/写 task.json）由 fixtures_test.go 的夹具现造出同样的
// 落盘形态，再由 status/relink 那条线覆盖 —— 见 workflow_*_test.go 的注释。

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/testutil"
	"github.com/luyaogod/TT/internal/devdata"
)

// TestCreate_RejectsInvalidProgram: 程序名会拼进工具链的每一层路径，非法名必须在
// 建目录之前就拒绝（这是唯一一处能在联网前拦下来的机会）。
func TestCreate_RejectsInvalidProgram(t *testing.T) {
	for _, prog := range []string{"", "AXMT500", "../x", "ab", "a b"} {
		_, err := Create(context.Background(), CreateRequest{
			Env: "e", Prog: prog, Dir: filepath.Join(t.TempDir(), "ws"),
		})
		if !errors.Is(err, devdata.ErrInvalidProg) {
			t.Errorf("程序名 %q: err = %v，想要 ErrInvalidProg", prog, err)
		}
	}
}

// TestCreate_RejectsLongOrSpaceTaskName: 任务名进目录名，超长与**首尾空白**都要拒绝。
//
// 回归：Create 曾经先 strings.TrimSpace 再校验，导致 `--name " epi "` 被静默修剪成
// `epi`，validTaskName 的"首尾不要空白"分支永远不可达（契约与实现打架）。现在 Create
// 按**原值**校验，这里两条一起钉住 —— 谁再把顺序调回去，这条就会红。
func TestCreate_RejectsLongOrSpaceTaskName(t *testing.T) {
	for name, why := range map[string]string{
		strings.Repeat("a", 41): "太长",
		" epi ":                 "首尾空白",
		"\tepi":                 "首部空白",
		"epi ":                  "尾部空白",
	} {
		_, err := Create(context.Background(), CreateRequest{
			Env: "", Prog: fixProg, Name: name, Dir: filepath.Join(t.TempDir(), "ws"),
		})
		if !errors.Is(err, ErrTaskName) {
			t.Errorf("任务名 %q（%s）: err = %v，想要 ErrTaskName", name, why, err)
		}
	}
}

// TestCreate_RefusesRootThatIsFile: --dir 指到一个文件时不能当成空目录用。
func TestCreate_RefusesRootThatIsFile(t *testing.T) {
	root := filepath.Join(t.TempDir(), "ws")
	if err := os.WriteFile(root, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Create(context.Background(), CreateRequest{Env: "e", Prog: fixProg, Dir: root})
	if !errors.Is(err, ErrWorkspaceExists) {
		t.Fatalf("err = %v，想要 ErrWorkspaceExists", err)
	}
	if !strings.Contains(err.Error(), "文件") {
		t.Errorf("提示没说清是文件: %v", err)
	}
}

// TestCreate_RefusesExistingWorkspace: 目录里有 task.json 时提示去看 status，
// 而不是笼统的"不是空目录"（两者的下一步完全不同）。
func TestCreate_RefusesExistingWorkspace(t *testing.T) {
	root := t.TempDir()
	writeTask(t, root, fixProg, "e")
	_, err := Create(context.Background(), CreateRequest{Env: "e", Prog: fixProg, Dir: root})
	if !errors.Is(err, ErrWorkspaceExists) {
		t.Fatalf("err = %v，想要 ErrWorkspaceExists", err)
	}
	if !strings.Contains(err.Error(), "tt dev status") {
		t.Errorf("提示里没有下一步: %v", err)
	}
}

// TestCreate_CleansUpOnlyDirItCreated: 中途失败时删掉**本次创建**的目录，
// 但用户本来就有（空的）目录不能动 —— "失败不留半套"与"不越权删除"是两条相反的纪律。
func TestCreate_CleansUpOnlyDirItCreated(t *testing.T) {
	// 1) 本次新建的目录（含父目录也是新建的）：失败后整棵本次目录消失
	base := t.TempDir()
	created := filepath.Join(base, "a", "b", "ws")
	_, err := Create(context.Background(), CreateRequest{Env: "", Prog: fixProg, Dir: created})
	if err == nil {
		t.Fatal("空环境下 Create 应当失败（未指定环境）")
	}
	if _, serr := os.Stat(created); !os.IsNotExist(serr) {
		t.Errorf("本次创建的目录没被清掉: %v", serr)
	}

	// 2) 用户已存在的空目录：失败后必须原样保留
	existing := t.TempDir()
	_, err = Create(context.Background(), CreateRequest{Env: "", Prog: fixProg, Dir: existing})
	if err == nil {
		t.Fatal("空环境下 Create 应当失败")
	}
	if st, serr := os.Stat(existing); serr != nil || !st.IsDir() {
		t.Errorf("用户已有的空目录被删了: %v", serr)
	}
}

// TestCreate_DefaultDirIsProgDashTask: 不给 --dir 时目录名是 `<程序>[-<任务名>]`，
// 落在**当前目录**下；失败时同样收干净。用 t.Chdir 把当前目录换到临时目录，
// 保证这条测试不在仓库里留下任何东西。
func TestCreate_DefaultDirIsProgDashTask(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)
	_, err := Create(context.Background(), CreateRequest{Env: "", Prog: fixProg, Name: "epi"})
	if err == nil {
		t.Fatal("空环境下 Create 应当失败")
	}
	if _, serr := os.Stat(filepath.Join(wd, fixProg+"-epi")); !os.IsNotExist(serr) {
		t.Errorf("默认目录没被清掉（或建到了别处）: %v", serr)
	}

	// 没给任务名时目录名就是程序名。
	if _, err := Create(context.Background(), CreateRequest{Env: "", Prog: fixProg}); err == nil {
		t.Fatal("空环境下 Create 应当失败")
	}
	if _, serr := os.Stat(filepath.Join(wd, fixProg)); !os.IsNotExist(serr) {
		t.Errorf("无任务名的默认目录没被清掉: %v", serr)
	}
}

// TestLoad_NonExistentPathSaysPath: 路径打错与"目录不是工作区"必须分开说 ——
// 前者要改路径，后者要 tt dev open。
func TestLoad_NonExistentPathSaysPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "根本不存在")
	_, _, err := Load(missing)
	if !errors.Is(err, ErrNoWorkspace) {
		t.Fatalf("err = %v，想要 ErrNoWorkspace", err)
	}
	if !strings.Contains(err.Error(), "路径不存在") {
		t.Errorf("提示没说清是路径不存在: %v", err)
	}
}

// TestLoad_InvalidJSON: task.json 坏了要报"不是合法 JSON"，不能静默当成空工作区。
func TestLoad_InvalidJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, TaskFileName), []byte("{不是 JSON"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := Load(root)
	if err == nil || !strings.Contains(err.Error(), "不是合法 JSON") {
		t.Fatalf("err = %v，想要 JSON 解析错误", err)
	}
}

// TestLoad_UnreadableTaskFile: task.json 是个目录（读不动）属于 IO 错误，
// 不能误报成"不是工作区"（后者会劝人去 open，而 open 只会再撞一次）。
func TestLoad_UnreadableTaskFile(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, TaskFileName), 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, err := Load(root)
	if err == nil {
		t.Fatal("读目录应当失败")
	}
	if errors.Is(err, ErrNoWorkspace) {
		t.Errorf("IO 错误被误报成 ErrNoWorkspace: %v", err)
	}
	if !strings.Contains(err.Error(), "读 ") {
		t.Errorf("提示没说是读失败: %v", err)
	}
}

// TestSave_RoundTrip: Save→Load 必须原样回来（含 Layout/Events/Ent），
// 且 Save 会刷新 UpdatedAt（task.json 只存身份与历史，时间戳是唯一会动的量）。
func TestSave_RoundTrip(t *testing.T) {
	root := t.TempDir()
	p := Paths{Root: root}
	task := &Task{
		Schema: TaskSchema, Prog: fixProg, Env: "e", Module: "cxm", Identity: "c",
		Task: "epi", Ent: 7, CreatedAt: "2026-11-21T20:00:00+08:00", Layout: DefaultLayout(),
	}
	task.AddEvent("open", "材料 4 个文件")
	if err := Save(task, p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if task.UpdatedAt == "" {
		t.Error("Save 没写 UpdatedAt")
	}
	back, p2, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if p2.Root != p.Root {
		t.Errorf("根路径 = %s，想要 %s", p2.Root, p.Root)
	}
	if back.Prog != fixProg || back.Env != "e" || back.Identity != "c" || back.Task != "epi" || back.Ent != 7 {
		t.Errorf("身份字段没原样回来: %+v", back)
	}
	if back.Layout != DefaultLayout() {
		t.Errorf("Layout 没原样回来: %+v", back.Layout)
	}
	if len(back.Events) != 1 || back.Events[0].Action != "open" {
		t.Errorf("事件没原样回来: %+v", back.Events)
	}
}

// TestSave_RejectsDirectoryTarget: task.json 位置被目录占住时，Save 必须报错
// （原子写的目标检查），不能假装写成功。
func TestSave_RejectsDirectoryTarget(t *testing.T) {
	root := t.TempDir()
	p := Paths{Root: root}
	if err := os.MkdirAll(p.TaskFile(), 0o755); err != nil {
		t.Fatal(err)
	}
	task := &Task{Schema: TaskSchema, Prog: fixProg, Env: "e", Layout: DefaultLayout()}
	if err := Save(task, p); err == nil {
		t.Fatal("目标目录存在时 Save 应当失败")
	}
}

// TestAddEvent_BoundsHistory: 大事记只留最近 20 条 —— 它记录"这一步为什么发生"，
// 不是日志；无限增长会让 task.json 变成没人读的垃圾场。
func TestAddEvent_BoundsHistory(t *testing.T) {
	task := &Task{Schema: TaskSchema, Prog: fixProg, Env: "e", Layout: DefaultLayout()}
	for i := 1; i <= 25; i++ {
		task.AddEvent("update", itoa(i))
	}
	if len(task.Events) != 20 {
		t.Fatalf("事件条数 = %d，想要 20", len(task.Events))
	}
	// 丢掉的是最早的 5 条（1..5），留下 6..25
	if task.Events[0].Detail != "6" || task.Events[19].Detail != "25" {
		t.Errorf("保留区间不对: 首=%s 尾=%s", task.Events[0].Detail, task.Events[19].Detail)
	}
}

// itoa 避免为了一个数字再引 strconv（本包测试里只用这一处）。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// TestSummarizeLink_NilAndFields: 摘要只带走小字段（不把 800 KB 的 doc 带出去），
// nil 进 nil 出（调用方据此区分"没建索引"）。
func TestSummarizeLink_NilAndFields(t *testing.T) {
	if got := SummarizeLink(nil); got != nil {
		t.Errorf("nil doc 的摘要 = %+v，想要 nil", got)
	}
	f := newFixture(t)
	sum := SummarizeLink(f.Doc)
	if sum == nil {
		t.Fatal("摘要为 nil")
	}
	if sum.Path != f.Doc.Path || sum.Schema != f.Doc.Schema {
		t.Errorf("路径/schema 没带上: %+v", sum)
	}
	if sum.Points != f.Doc.Stats.Points || sum.Custom != f.Doc.Stats.PointsCustom {
		t.Errorf("点数 = %d/%d，想要 %d/%d", sum.Points, sum.Custom, f.Doc.Stats.Points, f.Doc.Stats.PointsCustom)
	}
	if sum.Sections != len(f.Doc.Sections) || sum.Slices != f.Doc.Slices.Distinct {
		t.Errorf("sections/slices = %d/%d", sum.Sections, sum.Slices)
	}
	if sum.GeneratedAt == "" {
		t.Error("GeneratedAt 没带上")
	}
}

// TestShort_TruncatesTo12: 拒绝文案里的 sha 一律前 12 位，两个入口同一口径。
func TestShort_TruncatesTo12(t *testing.T) {
	if got := Short("abc"); got != "abc" {
		t.Errorf("短 sha = %q，想要原样", got)
	}
	long := strings.Repeat("0123456789", 7)
	if got := Short(long); got != long[:12] {
		t.Errorf("长 sha = %q，想要前 12 位", got)
	}
	if Short("") != "" {
		t.Error("空 sha 应原样")
	}
}

// TestHasRole: 材料完整性判断（有没有 .tap/.tgl）只看 role，不看文件名。
func TestHasRole(t *testing.T) {
	files := []devdata.FileEntry{{Role: "tab"}, {Role: "tap"}}
	if !hasRole(files, "tap") {
		t.Error("tap 没认出来")
	}
	if hasRole(files, "tgl") {
		t.Error("没有的 tgl 被认成了有")
	}
	if hasRole(nil, "tap") {
		t.Error("空清单不该判有")
	}
}

// TestSortedKeys_Stable: 提示文案里的键顺序必须稳定（否则同一份状态两次输出不一样）。
func TestSortedKeys_Stable(t *testing.T) {
	got := sortedKeys(map[string]string{"b": "2", "a": "1", "c": "3"})
	if strings.Join(got, ",") != "a,b,c" {
		t.Errorf("顺序 = %v", got)
	}
	if len(sortedKeys(nil)) != 0 {
		t.Error("空 map 应回空切片")
	}
}

// TestPaths_AbsAndNotes: 路径计算只有 Paths 一处；这里钉住最容易被漏掉的两个。
func TestPaths_AbsAndNotes(t *testing.T) {
	p := Paths{Root: filepath.Join("x", "ws")}
	if p.Abs() != p.Root {
		t.Errorf("Abs = %q", p.Abs())
	}
	if p.Notes() != filepath.Join(p.Root, DirNotes) {
		t.Errorf("Notes = %q", p.Notes())
	}
}

// TestCopyMdl_SkipsAndWarning: 模板切片是可选输入（少了只是放弃脱版检查），
// 三条"跳过"路径必须给可照做的提示，而不是静默成功。
func TestCopyMdl_SkipsAndWarning(t *testing.T) {
	dst := filepath.Join(t.TempDir(), DirMdl)

	// --no-mdl：明确不复制，不给警告
	if dir, warns := copyMdl("", "e", dst, true); dir != "" || len(warns) != 0 {
		t.Errorf("noMdl: dir=%q warns=%v", dir, warns)
	}
	// 环境名为空 → 取不到源，安静跳过
	if dir, warns := copyMdl("", "", dst, false); dir != "" || len(warns) != 0 {
		t.Errorf("空环境: dir=%q warns=%v", dir, warns)
	}
	// 环境里没镜像模板切片 → 一条提示
	env := newFixEnv(t)
	dir, warns := copyMdl(env.Cfg, "e", dst, false)
	if dir != "" || len(warns) != 1 || !strings.Contains(warns[0], "模板切片没镜像") {
		t.Errorf("没镜像: dir=%q warns=%v", dir, warns)
	}
}

// TestCopyMdl_CopiesSliceTree: 复制成功时落点是 `<工作区>/mdl/slice/**`，
// 且只带 slice/（整棵 3.2 MB 模板树不进工作区）。
func TestCopyMdl_CopiesSliceTree(t *testing.T) {
	env := newFixEnv(t)
	env.writeSlice(t)
	dst := filepath.Join(t.TempDir(), DirMdl)

	dir, warns := copyMdl(env.Cfg, "e", dst, false)
	if dir != dst {
		t.Fatalf("落点 = %q，想要 %q", dir, dst)
	}
	if len(warns) != 0 {
		t.Errorf("成功时不该有提示: %v", warns)
	}
	for _, name := range []string{"code_a05.template", "code_qs01.template"} {
		if _, err := os.Stat(filepath.Join(dst, "slice", name)); err != nil {
			t.Errorf("切片 %s 没复制过来: %v", name, err)
		}
	}
}

// TestCopyMdl_WarnsWhenTempDirFails: 目标父路径被文件占住时，复制失败只警告
// （模板切片不是建工作区的必要条件），但必须说清"失败并跳过"。
func TestCopyMdl_WarnsWhenTempDirFails(t *testing.T) {
	env := newFixEnv(t)
	env.writeSlice(t)
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir, warns := copyMdl(env.Cfg, "e", filepath.Join(blocker, DirMdl), false)
	if dir != "" {
		t.Errorf("失败时 dir = %q，想要空", dir)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "复制模板切片失败") {
		t.Errorf("warns = %v，想要一条复制失败提示", warns)
	}
}

// TestCopyTree_ErrorOnMissingSource: copyTree 失败要把错误带出去（调用方据此降级，
// 不能把"半个模板树"当成功）。
func TestCopyTree_ErrorOnMissingSource(t *testing.T) {
	dir := t.TempDir()
	err := copyTree(filepath.Join(dir, "没有这个目录"), filepath.Join(dir, "out"))
	if err == nil {
		t.Fatal("源目录不存在时应当报错")
	}
}

// TestRebindPackage_Errors: 工作区/视图缺失时 Rebind 必须报错，而不是返回空路径
// 让调用方照着空路径去写包。
func TestRebindPackage_Errors(t *testing.T) {
	// 1) 不是工作区
	if _, _, err := RebindPackage(t.TempDir()); !errors.Is(err, ErrNoWorkspace) {
		t.Errorf("非工作区: err = %v，想要 ErrNoWorkspace", err)
	}

	// 2) 是工作区但没有视图（store.Open 失败）
	root := t.TempDir()
	writeTask(t, root, fixProg, "e")
	if _, _, err := RebindPackage(root); err == nil {
		t.Error("没有视图时应当报错")
	}

	// 3) 视图文件在、manifest 内容坏了（store.Open 通过，Manifest 失败）
	f := newFixture(t)
	if err := os.WriteFile(filepath.Join(f.P.Ws(), "manifest.json"), []byte("{坏"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RebindPackage(f.Root); err == nil || !strings.Contains(err.Error(), "manifest") {
		t.Errorf("坏 manifest: err = %v，想要 manifest 解析错误", err)
	}
}

// TestRebindPackage_NoLocalPackage: 清单指着别处、本地又没有包时，
// 必须拒绝（ErrPackageMismatch）并提示两条出路，而不是把 apply 指向一个不存在的文件。
func TestRebindPackage_NoLocalPackage(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(f.Pkg); err != nil {
		t.Fatal(err)
	}
	setManifestPkgPath(t, f.P.Ws(), filepath.Join(f.Root, "别处", "x.tzc"))
	_, _, err := RebindPackage(f.Root)
	if !errors.Is(err, ErrPackageMismatch) {
		t.Fatalf("err = %v，想要 ErrPackageMismatch", err)
	}
	if !strings.Contains(err.Error(), "没有包") {
		t.Errorf("提示没说清本地没有包: %v", err)
	}
}

// TestRebindPackage_AlreadyBoundIsNoop: 已经绑回本地的路径再绑一次要回 rebased=false
// （幂等；CLI 据此不重复打印"已改绑"）。
func TestRebindPackage_AlreadyBoundIsNoop(t *testing.T) {
	f := newFixture(t)
	local, rebased, err := RebindPackage(f.Root)
	if err != nil {
		t.Fatalf("RebindPackage: %v", err)
	}
	if rebased {
		t.Fatalf("视图基线本来就指着本地包，不该报改绑")
	}
	if local != f.P.Pkg(f.Prog, "c") {
		t.Errorf("本地包路径 = %q", local)
	}
}

// TestPackagePath_PresentAndMissing: 包路径查询"不存在返回空、不报错"
// （status/brief 拿它决定要不要显示 pkg 一行）。
func TestPackagePath_PresentAndMissing(t *testing.T) {
	f := newFixture(t)
	got, err := PackagePath(f.Root)
	if err != nil {
		t.Fatalf("PackagePath: %v", err)
	}
	if got != f.Pkg {
		t.Errorf("路径 = %q，想要 %q", got, f.Pkg)
	}
	if err := os.Remove(f.Pkg); err != nil {
		t.Fatal(err)
	}
	got, err = PackagePath(f.Root)
	if err != nil || got != "" {
		t.Errorf("包删掉后 = %q / %v，想要空路径且无错", got, err)
	}
	// identity 没记的老工作区按 c 兜底
	task := f.loadTask()
	task.Identity = ""
	if err := Save(task, f.P); err != nil {
		t.Fatal(err)
	}
	f.Pkg = fixPackage(t, f.P, f.Prog, nil)
	got, err = PackagePath(f.Root)
	if err != nil || got != f.P.Pkg(f.Prog, "c") {
		t.Errorf("无身份工作区 = %q / %v", got, err)
	}
}

// TestPkgPoint_FromPackage: apply 之后"我的改动进包没有"只能看包里的 .tap
// （.4gl 还是服务器 build 的旧产物）。这里钉住正常取到正文与"没有这个点"。
func TestPkgPoint_FromPackage(t *testing.T) {
	f := newFixture(t)

	// 自订定义点在包内的 .tap 里（testutil.NormalTAP 的 function.<prog>_calc）
	view, path, err := PkgPoint(f.Root, "function."+f.Prog+"_calc")
	if err != nil {
		t.Fatalf("PkgPoint: %v", err)
	}
	if path != f.Pkg {
		t.Errorf("包路径 = %q，想要 %q", path, f.Pkg)
	}
	if view == nil || view.Name != "function."+f.Prog+"_calc" {
		t.Fatalf("点位视图 = %+v", view)
	}
	if len(view.Content) == 0 || !strings.Contains(strings.Join(view.Content, "\n"), "PRIVATE FUNCTION") {
		t.Errorf("正文没取到: %+v", view.Content)
	}

	// 没有这个点：nil 视图 + 不是错误（与"材料读不动"区分开）
	view, _, err = PkgPoint(f.Root, "根本没有这个点")
	if err != nil {
		t.Fatalf("查不存在的点不该报错: %v", err)
	}
	if view != nil {
		t.Errorf("不存在的点返回了 %+v", view)
	}
}

// TestPkgPoint_FallsBackToManifestPath: 老工作区可能把包放在别处（清单里记着），
// 本地 pkg/ 下没有时按清单找。
func TestPkgPoint_FallsBackToManifestPath(t *testing.T) {
	f := newFixture(t)
	elsewhere, err := fixPackageAt(t, filepath.Join(f.Root, "别处"), f.Prog)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(f.Pkg); err != nil {
		t.Fatal(err)
	}
	setManifestPkgPath(t, f.P.Ws(), elsewhere)
	view, path, err := PkgPoint(f.Root, "global.memo")
	if err != nil {
		t.Fatalf("PkgPoint(回退): %v", err)
	}
	if path != elsewhere {
		t.Errorf("回退路径 = %q，想要 %q", path, elsewhere)
	}
	if view == nil {
		t.Error("回退后没取到点")
	}
}

// TestPkgPoint_Errors: 包损坏 / .tap 解析失败都必须报错（此时"改动进包没有"是
// 回答不了的，不能给一个空视图让人以为没改动）。
func TestPkgPoint_Errors(t *testing.T) {
	f := newFixture(t)

	// 1) 包损坏
	if err := os.WriteFile(f.Pkg, []byte("不是 zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := PkgPoint(f.Root, "global.memo"); err == nil {
		t.Error("损坏的包应当报错")
	}

	// 2) 包能打开但 .tap 解析失败
	entries := map[string]string{
		f.Prog + ".tap": "<不是合法 XML",
		f.Prog + ".tgl": "x\n",
		f.Prog + ".4gl": "MAIN\nEND MAIN\n",
		"ver":           "1.0\n",
	}
	bad, err := testutil.WritePackage(f.P.PkgDir(), f.Prog+"(c).tzc", entries)
	if err != nil {
		t.Fatal(err)
	}
	if bad != f.Pkg {
		t.Fatalf("包落点 = %q", bad)
	}
	if _, _, err := PkgPoint(f.Root, "global.memo"); err == nil {
		t.Error(".tap 解析失败应当报错")
	}
}

// TestPkgPoint_NoPackage: 包与清单都没有时，报的是"打不开那个路径"，
// 并且把尝试过的路径带回去（用户要知道该把包放回哪）。
func TestPkgPoint_NoPackage(t *testing.T) {
	f := newFixture(t)
	if err := os.Remove(f.Pkg); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(f.P.Ws(), "manifest.json")); err != nil {
		t.Fatal(err)
	}
	_, path, err := PkgPoint(f.Root, "global.memo")
	if err == nil {
		t.Fatal("没有包时应当报错")
	}
	if path != f.P.Pkg(f.Prog, "c") {
		t.Errorf("返回的包路径 = %q，想要本地包落点", path)
	}
}

// TestValidTaskName_Edges: 任务名规则的边界单独钉一遍（Create 里还有"程序名先校验"
// 的顺序问题，散在别处）。
func TestValidTaskName_Edges(t *testing.T) {
	ok := []string{"", "epi", "修 30 字以内的任务"}
	for _, name := range ok {
		if err := validTaskName(name); err != nil {
			t.Errorf("合法任务名 %q 被拒: %v", name, err)
		}
	}
	bad := []string{strings.Repeat("x", 41), " x", "x ", "a/b", `a\b`, "a..b", `a:b`}
	for _, name := range bad {
		if err := validTaskName(name); !errors.Is(err, ErrTaskName) {
			t.Errorf("非法任务名 %q: err = %v，想要 ErrTaskName", name, err)
		}
	}
}

// TestCreate_ReportsDirCreationFailures: 目标目录建不出来时要报"创建目录失败"
// （父路径被文件占住是最常见的形态：用户 --dir 手滑指到了某个文件下面）。
func TestCreate_ReportsDirCreationFailures(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 父路径是文件 → MkdirAll 失败
	_, err := Create(context.Background(), CreateRequest{
		Env: "", Prog: fixProg, Dir: filepath.Join(blocker, "ws"),
	})
	if err == nil || !strings.Contains(err.Error(), "创建目录失败") {
		t.Errorf("父路径是文件: err = %v，想要创建目录失败", err)
	}
}

// TestRebindPackage_RefusesWhenManifestUnwritable: 视图清单写不回去时 Rebind 必须报错
// （否则 apply 会照着一份没改绑的清单写进别人的包）。用只读清单在 Windows 上让
// 原子替换失败。
func TestRebindPackage_RefusesWhenManifestUnwritable(t *testing.T) {
	f := newFixture(t)
	manifest := filepath.Join(f.P.Ws(), "manifest.json")
	setManifestPkgPath(t, f.P.Ws(), filepath.Join(f.Root, "旧目录", "x.tzc"))
	if err := os.Chmod(manifest, 0o444); err != nil {
		t.Fatalf("设只读: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(manifest, 0o644) })
	if _, _, err := RebindPackage(f.Root); err == nil {
		t.Fatal("清单写不回去时应报错")
	}
}

// TestCopyMdl_WarnsWhenReplacingExistingTreeFails: 工作区里已有 mdl/ 且旧目录删不掉时，
// 复制模板切片要降级成一条提示（不影响建工作区），而不是把失败当成功。
// 用"目标目录是当前工作目录"制造 RemoveAll 失败（Windows 上 cwd 删不掉）。
func TestCopyMdl_WarnsWhenReplacingExistingTreeFails(t *testing.T) {
	env := newFixEnv(t)
	env.writeSlice(t)
	dst := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dst, "slice"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dst)

	dir, warns := copyMdl(env.Cfg, "e", dst, false)
	if dir != "" {
		t.Errorf("失败时 dir = %q，想要空", dir)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "替换模板切片失败") {
		t.Errorf("warns = %v，想要一条替换失败提示", warns)
	}
}

// TestLoad_UnusablePath: 路径本身无法绝对化（Windows 上含 NUL 的路径）时要报
// "工作区路径不可用"，而不是笼统地报"不是工作区"。
func TestLoad_UnusablePath(t *testing.T) {
	_, _, err := Load("a\x00b")
	if err == nil || !strings.Contains(err.Error(), "工作区路径不可用") {
		t.Fatalf("err = %v，想要路径不可用", err)
	}
}
