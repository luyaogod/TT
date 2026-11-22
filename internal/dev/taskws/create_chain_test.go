package taskws

// Create 全链（拉料 → 建索引 → 打包 → 导出视图 → task.json）的离线测试。
//
// 这一整条链路原先只能连真 SFTP 才能跑；生产侧给 CreateRequest 加了
// `Files devdata.FileSource` 注入缝之后，用 fakesource_test.go 的假 $TOP 就能在
// 断网环境里端到端跑完，并且可以精确断言"每一步落到磁盘上是什么"。
//
// 这条测试也是"Create 失败不留半套 / 无 .tap 标准程序被拦 / 身份与选项透传"的宿主。

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/pkgfile"
	"github.com/luyaogod/TT/internal/dev/testutil"
	"github.com/luyaogod/TT/internal/devdata"
)

// chain 是一条"用假 $TOP 建出来的工作区"，后续 Update 分支测试复用。
type chain struct {
	t   *testing.T
	env fixEnv
	src *fakeSource
	res *CreateResult
}

// Root / P 是工作区根与路径计算器（方便断言）。
func (c *chain) Root() string { return c.res.Dir }
func (c *chain) P() Paths     { return Paths{Root: c.res.Dir} }
func (c *chain) Task() *Task {
	t, _, err := Load(c.Root())
	if err != nil {
		c.t.Fatalf("Load: %v", err)
	}
	return t
}
func (c *chain) status() *Status {
	c.t.Helper()
	st, err := GetStatus(StatusRequest{Dir: c.Root()})
	if err != nil {
		c.t.Fatalf("GetStatus: %v", err)
	}
	return st
}

// newOpenFixture 造"环境 + 假 $TOP（已放一套正常程序材料）"。
func newOpenFixture(t *testing.T) (fixEnv, *fakeSource) {
	t.Helper()
	env := newFixEnv(t)
	env.writeSlice(t) // 环境里已镜像模板切片 → Create 会把它复制进工作区
	src := newFakeSource("/u1/top")
	seedProgram(t, src, fixProg, testutil.NormalEntries(fixProg))
	return env, src
}

// runCreate 用给定输入跑一次 Create（只补公共字段，其余由调用方决定）。
func runCreate(t *testing.T, env fixEnv, src *fakeSource, root string, patch func(*CreateRequest)) *CreateResult {
	t.Helper()
	req := CreateRequest{
		CfgPath: env.Cfg, Env: "e", Prog: fixProg, Dir: root, Name: "epi",
		Ent: 7, TTVersion: "9.9.9", Files: src,
	}
	if patch != nil {
		patch(&req)
	}
	res, err := Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	return res
}

// openViaSource 是 Update 分支测试的入口：用假 $TOP 建一个完整工作区。
func openViaSource(t *testing.T) *chain {
	t.Helper()
	env, src := newOpenFixture(t)
	root := filepath.Join(t.TempDir(), "ws")
	res := runCreate(t, env, src, root, nil)
	return &chain{t: t, env: env, src: src, res: res}
}

// TestCreate_FullChainWithInjectedSource: 一次 open 的每一步都要留下可复核的落盘证据：
// 材料文件 + manifest 指纹、索引摘要、包成员、围栏工作区的 .tdev 布局、task.json 身份
// 与 open 事件、告警透传、进度回调。任何一步"看起来成功但没落盘"都会在这里露出来。
func TestCreate_FullChainWithInjectedSource(t *testing.T) {
	env, src := newOpenFixture(t)
	root := filepath.Join(t.TempDir(), "ws")
	var phases []string
	var currentFiles []string
	res := runCreate(t, env, src, root, func(req *CreateRequest) {
		req.Progress = func(p devdata.Progress) {
			phases = append(phases, p.Phase)
			if p.CurrentFile != "" {
				currentFiles = append(currentFiles, p.CurrentFile)
			}
		}
	})

	p := Paths{Root: res.Dir}
	if res.Dir != root {
		t.Errorf("工作区根 = %q，想要 %q", res.Dir, root)
	}

	// ---- 1) 材料落盘 + manifest 指纹 ----
	for _, rel := range []string{
		"dzx/" + fixProg + ".tab", "dzx/" + fixProg + ".tap",
		"dzx/" + fixProg + ".tgl", "src/" + fixProg + ".4gl",
	} {
		full := filepath.Join(p.Materials(), filepath.FromSlash(rel))
		if _, err := os.Stat(full); err != nil {
			t.Errorf("材料没落盘 %s: %v", rel, err)
		}
	}
	// 内容必须逐字节等于服务器上的原字节（fetch 不做任何转换）
	got, err := os.ReadFile(filepath.Join(p.Materials(), "dzx", fixProg+".tgl"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != testutil.NormalTGL(fixProg) {
		t.Error("落盘的 .tgl 与服务器字节不一致")
	}
	mf, err := devdata.ReadManifest(p.Materials())
	if err != nil || mf == nil {
		t.Fatalf("读材料清单: mf=%v err=%v", mf, err)
	}
	if len(mf.Files) != 4 || mf.Module != "axm" || mf.Source.TopDir != "/u1/top" {
		t.Errorf("材料清单 = files:%d module:%q top:%q", len(mf.Files), mf.Module, mf.Source.TopDir)
	}
	if mf.Versions.TemplateVersions["a05"] != 3 {
		t.Errorf("模板版本锚点 = %v", mf.Versions.TemplateVersions)
	}
	if res.Fetch == nil || len(res.Fetch.Files) != 4 {
		t.Fatalf("Fetch 结果 = %+v", res.Fetch)
	}
	if res.Fetch.Module != "axm" || res.Fetch.ModuleOrigin != "standard" {
		t.Errorf("模块判定 = %q/%q，想要 axm/standard", res.Fetch.Module, res.Fetch.ModuleOrigin)
	}
	if len(res.Fetch.Missing) == 0 {
		t.Error("服务器上没给的 tsd/4fd/inc 等可选素材应进 missing[]")
	}

	// ---- 2) 索引摘要（IndexSummary 不把大 doc 带出去）----
	if res.Index == nil {
		t.Fatal("Create 没回索引摘要")
	}
	if res.Index.Path != p.Index() || res.Index.Schema == "" || res.Index.GeneratedAt == "" {
		t.Errorf("索引抬头 = %+v", res.Index)
	}
	if res.Index.Points == 0 || res.Index.Sections == 0 {
		t.Errorf("索引摘要 = %+v（点数/区段都要非 0）", res.Index)
	}

	// ---- 3) 包：落点按身份命名，成员齐全 ----
	if res.Pack == nil {
		t.Fatal("Create 没回打包结果")
	}
	wantPkg := filepath.Join(p.PkgDir(), fixProg+"(c).tzc")
	if res.Pack.Path != wantPkg || res.Pack.Identity != "c" {
		t.Errorf("包 = %q identity=%q，想要 %q/c", res.Pack.Path, res.Pack.Identity, wantPkg)
	}
	if res.Pack.Bytes == 0 || res.Pack.SHA256 == "" {
		t.Errorf("包统计 = bytes:%d sha:%q", res.Pack.Bytes, res.Pack.SHA256)
	}
	pkg, err := pkgfile.Open(res.Pack.Path, pkgfile.OpenOptions{AcceptVer: "1.0"})
	if err != nil {
		t.Fatalf("打开产出的包: %v", err)
	}
	names := map[string]bool{}
	for _, e := range pkg.Entries {
		names[e.Name] = true
	}
	for _, want := range []string{fixProg + ".4gl", fixProg + ".tap", fixProg + ".tgl", "ver"} {
		if !names[want] {
			t.Errorf("包里缺成员 %s（实际 %v）", want, names)
		}
	}

	// ---- 4) 视图：围栏工作区 + .tdev 布局 ----
	if res.Export == nil || res.Export.Dir != p.Ws() || res.Export.Regions == 0 {
		t.Fatalf("导出结果 = %+v", res.Export)
	}
	if res.Export.Manifest == nil {
		t.Error("导出结果没带工作区清单")
	}
	for _, rel := range []string{
		"prog.full.4gl", "manifest.json", "snapshot/index.json",
		".tdev/base.full.4gl", ".tdev/base.sha256", ".tdev/regions.json", ".tdev/section-state",
	} {
		if _, err := os.Stat(filepath.Join(p.Ws(), filepath.FromSlash(rel))); err != nil {
			t.Errorf("视图缺 %s: %v", rel, err)
		}
	}

	// ---- 5) task.json：身份 + open 事件 ----
	task := &Task{}
	{
		loaded, _, lerr := Load(root)
		if lerr != nil {
			t.Fatalf("Load: %v", lerr)
		}
		task = loaded
	}
	if task.Schema != TaskSchema || task.Prog != fixProg || task.Env != "e" ||
		task.Module != "axm" || task.Identity != "c" || task.Task != "epi" || task.Ent != 7 ||
		task.TTVersion != "9.9.9" {
		t.Errorf("task.json 身份 = %+v", task)
	}
	if task.CreatedAt == "" || task.UpdatedAt == "" {
		t.Error("task.json 缺时间戳")
	}
	if len(task.Events) != 1 || task.Events[0].Action != "open" {
		t.Fatalf("open 事件 = %+v", task.Events)
	}
	if !strings.Contains(task.Events[0].Detail, "材料 4 个文件") ||
		!strings.Contains(task.Events[0].Detail, fixProg+"(c).tzc") {
		t.Errorf("open 事件详情 = %q", task.Events[0].Detail)
	}
	if _, err := os.Stat(p.Notes()); err != nil {
		t.Errorf("notes/ 没建: %v", err)
	}

	// ---- 6) 告警透传（fetch 的 + export 的）----
	warns := strings.Join(res.Warnings, "；")
	if !strings.Contains(warns, "取不到该程序的版次") {
		t.Errorf("fetch 的版本锚点告警没透传: %v", res.Warnings)
	}
	if !strings.Contains(warns, "框架未解锁") {
		t.Errorf("export 的解锁提示没透传: %v", res.Warnings)
	}

	// ---- 7) 进度回调与"真的去拉了"（假来源的访问日志）----
	if len(phases) == 0 || phases[len(phases)-1] != "done" {
		t.Errorf("进度阶段 = %v，最后应是 done", phases)
	}
	if strings.Join(phases, ",") == "connect" || containsPhase(phases, "connect") {
		t.Error("注入了 FileSource 时不应有 connect 阶段（没有真的连服务器）")
	}
	if len(currentFiles) == 0 {
		t.Error("fetch 阶段没有回报当前文件")
	}
	read := strings.Join(src.readLog(), "\n")
	for _, want := range []string{
		"erp/axm/dzx/tab/" + fixProg + ".tab",
		"erp/axm/dzx/tap/" + fixProg + ".tap",
		"erp/axm/dzx/tgl/" + fixProg + ".tgl",
		"erp/axm/4gl/" + fixProg + ".4gl",
	} {
		if !strings.Contains(read, want) {
			t.Errorf("没从假服务器读 %s（读了 %v）", want, src.readLog())
		}
	}
}

// containsPhase 判进度阶段里有没有某个值。
func containsPhase(phases []string, want string) bool {
	for _, p := range phases {
		if p == want {
			return true
		}
	}
	return false
}

// TestCreate_RejectsProgramWithoutTap: 没被设计器接管过的标准程序（服务器上没有 .tap）
// 建不了工作区 —— 视图与包都从 .tzc 来，服务器上根本没有这个包。必须报
// ErrMaterialIncomplete，并且不留半套目录。
func TestCreate_RejectsProgramWithoutTap(t *testing.T) {
	env := newFixEnv(t)
	env.writeSlice(t)
	src := newFakeSource("/u1/top")
	entries := testutil.NormalEntries(fixProg)
	delete(entries, fixProg+".tap") // 有 tab/tgl/4gl，就是没有 .tap
	seedProgram(t, src, fixProg, entries)

	root := filepath.Join(t.TempDir(), "ws")
	_, err := Create(context.Background(), CreateRequest{
		CfgPath: env.Cfg, Env: "e", Prog: fixProg, Dir: root, Files: src,
	})
	if !errorsIs(err, devdata.ErrMaterialIncomplete) {
		t.Fatalf("err = %v，想要 ErrMaterialIncomplete", err)
	}
	if !strings.Contains(err.Error(), "没有被设计器接管") {
		t.Errorf("提示没说清是没被设计器接管: %v", err)
	}
	if _, serr := os.Stat(root); !os.IsNotExist(serr) {
		t.Errorf("失败后没清掉本次建的目录: %v", serr)
	}
	if strings.Contains(strings.Join(src.readLog(), "\n"), ".tap") {
		t.Error("服务器上没有 .tap，不该去读它")
	}
}

// errorsIs 是 errors.Is 的短包装（本文件只需要匹配一个哨兵）。
func errorsIs(err, target error) bool {
	if err == nil {
		return false
	}
	for err != nil {
		if err == target {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

// TestCreate_IdentityFromTapAndOverride: 包身份（c/s）默认从 .tap 根属性推断，
// 显式 --identity 时以显式值为准 —— 包文件名与 task.json 必须同口径。
func TestCreate_IdentityFromTapAndOverride(t *testing.T) {
	// 1) TAP 里写 identity="s" → 包按 (s) 命名，task.json 记 s
	env := newFixEnv(t)
	src := newFakeSource("/u1/top")
	seedProgram(t, src, fixProg, testutil.NormalEntriesWith(fixProg, map[string]string{"identity": "s"}))
	root := filepath.Join(t.TempDir(), "ws1")
	res := runCreate(t, env, src, root, nil)
	if res.Pack.Identity != "s" {
		t.Errorf("身份 = %q，想要从 .tap 推断出 s", res.Pack.Identity)
	}
	if base := filepath.Base(res.Pack.Path); base != fixProg+"(s).tzc" {
		t.Errorf("包名 = %q，想要 %q", base, fixProg+"(s).tzc")
	}
	task := &Task{}
	loaded, _, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	task = loaded
	if task.Identity != "s" {
		t.Errorf("task.json 身份 = %q，想要 s", task.Identity)
	}
	if st := (&chain{t: t, env: env, src: src, res: res}).status(); st.Pkg.Path != res.Pack.Path {
		t.Errorf("status 包路径 = %q，想要 %q", st.Pkg.Path, res.Pack.Path)
	}

	// 2) 显式 Identity=c 覆盖 TAP 的 s
	env2 := newFixEnv(t)
	src2 := newFakeSource("/u1/top")
	seedProgram(t, src2, fixProg, testutil.NormalEntriesWith(fixProg, map[string]string{"identity": "s"}))
	root2 := filepath.Join(t.TempDir(), "ws2")
	res2 := runCreate(t, env2, src2, root2, func(req *CreateRequest) { req.Identity = "c" })
	if res2.Pack.Identity != "c" || filepath.Base(res2.Pack.Path) != fixProg+"(c).tzc" {
		t.Errorf("显式身份没生效: %+v", res2.Pack)
	}
}

// TestCreate_NoMdlSkipsSliceCopy: --no-mdl 表示放弃切片脱版检测：不复制模板切片、
// 不报"没镜像"警告（那是没开 --no-mdl 时的提示），status 的 mdl 一块标不存在。
func TestCreate_NoMdlSkipsSliceCopy(t *testing.T) {
	env, src := newOpenFixture(t)
	root := filepath.Join(t.TempDir(), "ws")
	res := runCreate(t, env, src, root, func(req *CreateRequest) { req.NoMdl = true })
	p := Paths{Root: root}
	if _, err := os.Stat(p.Mdl()); !os.IsNotExist(err) {
		t.Errorf("--no-mdl 却建了 mdl/: %v", err)
	}
	if strings.Contains(strings.Join(res.Warnings, "；"), "模板切片没镜像") {
		t.Errorf("--no-mdl 不该报没镜像: %v", res.Warnings)
	}
	st, err := GetStatus(StatusRequest{Dir: root})
	if err != nil {
		t.Fatal(err)
	}
	if st.Mdl.Present {
		t.Errorf("status 说模板切片存在: %+v", st.Mdl)
	}
	// 对照：不带 --no-mdl 时切片确实被复制进来（否则上面那条断言没有意义）
	root2 := filepath.Join(t.TempDir(), "ws2")
	runCreate(t, env, src, root2, nil)
	if st2, err := GetStatus(StatusRequest{Dir: root2}); err != nil || !st2.Mdl.Present || st2.Mdl.Templates != 2 {
		t.Errorf("默认应复制 2 片模板: st=%+v err=%v", st2, err)
	}
}

// TestCreate_FormOptionPassthrough: --form 是"连 .42f 一起拉"，必须真的出现在请求里
// （假来源的访问日志是唯一能证明"请求参数没被吞"的地方）。
func TestCreate_FormOptionPassthrough(t *testing.T) {
	env, src := newOpenFixture(t)
	seedForm(t, src, fixProg)

	// 默认不拉 .42f
	root := filepath.Join(t.TempDir(), "ws-noform")
	res := runCreate(t, env, src, root, nil)
	if strings.Contains(strings.Join(src.readLog(), "\n"), ".42f") {
		t.Error("没给 --form 却读了 .42f")
	}
	if _, err := os.Stat(filepath.Join(Paths{Root: root}.Materials(), "form", fixProg+".42f")); !os.IsNotExist(err) {
		t.Errorf("没给 --form 却落了 .42f: %v", err)
	}
	_ = res

	// --form：请求里要出现 42f 路径，并落进 materials/form/
	src.resetLog()
	root2 := filepath.Join(t.TempDir(), "ws-form")
	res2 := runCreate(t, env, src, root2, func(req *CreateRequest) { req.Form = true })
	if !strings.Contains(strings.Join(src.readLog(), "\n"), "erp/axm/42f/"+fixProg+".42f") {
		t.Errorf("--form 没把 .42f 加进请求（读了 %v）", src.readLog())
	}
	if _, err := os.Stat(filepath.Join(Paths{Root: root2}.Materials(), "form", fixProg+".42f")); err != nil {
		t.Errorf("--form 的 .42f 没落盘: %v", err)
	}
	formRole := false
	for _, f := range res2.Fetch.Files {
		if f.Role == "form" {
			formRole = true
		}
	}
	if !formRole {
		t.Errorf("Fetch 结果里没有 form 角色: %+v", res2.Fetch.Files)
	}
}

// TestCreate_ReportsIndexError: 材料拉下来了但 .tap 解析不了时，建索引会失败 ——
// Create 必须把错误冒出来并清掉本次建的目录（半套工作区比没有更坏）。
func TestCreate_ReportsIndexError(t *testing.T) {
	env := newFixEnv(t)
	src := newFakeSource("/u1/top")
	entries := testutil.NormalEntries(fixProg)
	entries[fixProg+".tap"] = "<不是合法 XML"
	seedProgram(t, src, fixProg, entries)

	root := filepath.Join(t.TempDir(), "ws")
	_, err := Create(context.Background(), CreateRequest{
		CfgPath: env.Cfg, Env: "e", Prog: fixProg, Dir: root, Files: src,
	})
	if err == nil {
		t.Fatal("坏 .tap 建索引应当失败")
	}
	if _, serr := os.Stat(root); !os.IsNotExist(serr) {
		t.Errorf("失败后没清掉本次建的目录: %v", serr)
	}
}
