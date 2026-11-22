package taskws

// 合成夹具：在 t.TempDir() 里现造一份**devdata 能解析**的工作区落盘形态
// （task.json + materials/ + mdl/slice + index/link.json + pkg/*.tzc + ws/ + notes/），
// 让"open 之后"的全流程（status / relink / update 的本地闸门）能在断网环境里被钉住。
//
// 为什么不用真实材料：工作区的正确性几乎全在"文件形态"上（manifest 指纹、包 sha、
// 视图基线、围栏工作区布局）。用真实语料只会把网络、客户目录、配置一起变成隐性依赖，
// 而单测的价值恰恰是"同一份输入永远给同一个结论"。这里每个文件都是最小但合法的一套。
//
// 材料文本照 internal/devdata/link_test.go 的合成形态（TAP 根 <add_points>、TGL 只放
// 区段标记、4GL 是服务器产物），保证 devdata.Map 真的能从它建出索引 —— 不是随手写
// 几个"看起来像"的文件。

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/luyaogod/TT/internal/dev/testutil"
	"github.com/luyaogod/TT/internal/devdata"
)

// fixProg 是夹具统一的程序名（合法形态：小写字母开头 + 下划线）。
const fixProg = "axmt500_wf"

// fixEnv 是一份**合成**配置 + 它指向的开发资料根。
type fixEnv struct {
	Cfg     string // 临时 config.json 的绝对路径（永不是用户真实的那个）
	DevData string // <tmp>/devdata：环境 e 的 devdataDir
}

// newFixEnv 写一份只含环境 e 的 config.json。
//
// 环境**不配 SSH 主机**是刻意的：任何走到 OpenFileSource 的代码都会在拨号之前
// 报"没有配置 SSH 主机"，测试因此不可能真的出网（失败点就是"没联网"的证据）。
func newFixEnv(t *testing.T) fixEnv {
	t.Helper()
	dir := t.TempDir()
	devData := filepath.Join(dir, "devdata")
	body := `{"schemaVersion":4,"hosts":{"sshs":[{"name":"e","port":22,"user":"u","devdataDir":` +
		fixJSONString(devData) + `}]}}`
	cfg := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfg, []byte(body), 0o644); err != nil {
		t.Fatalf("写测试配置: %v", err)
	}
	return fixEnv{Cfg: cfg, DevData: devData}
}

// fixJSONString 把路径安全地放进 JSON 字符串里（Windows 反斜杠要转义）。
func fixJSONString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// writeSlice 在环境 e 的模板树下造两片 slice 模板（versionAnchors/脱版检测用）。
func (e fixEnv) writeSlice(t *testing.T) string {
	t.Helper()
	slice := filepath.Join(e.DevData, "e", "mdl", "slice")
	if err := os.MkdirAll(slice, 0o755); err != nil {
		t.Fatalf("建模板切片目录: %v", err)
	}
	for _, name := range []string{"code_a05.template", "code_qs01.template"} {
		body := "#應用 a05 樣板自動產生(Version:3)\n"
		if err := os.WriteFile(filepath.Join(slice, name), []byte(body), 0o644); err != nil {
			t.Fatalf("写模板切片: %v", err)
		}
	}
	return filepath.Join(e.DevData, "e", "mdl")
}

// fixTab 是最小可解析的 .tab（蓝图）：一个 page + 一个 screen record 字段。
// 形态照 internal/devdata/link_test.go 的 tabDocText。
func fixTab(prog string) string {
	return `<?xml version='1.0' encoding='UTF-8'?>` + "\n" +
		`<assembly name="` + prog + `" module="cxm" type="t01" page="1" cusver="0003">` + "\n" +
		"  <structure>\n" +
		`    <var id="body" page="1" page_id="1" record="s_detail1" value="f1,f2"/>` + "\n" +
		"  </structure>\n</assembly>\n"
}

// fixMaterials 写一套 devdata 能解析的材料 + manifest（指纹与大小按内容现算）。
//
// 角色与落点按 devdata.FetchRoles 的真实形态：tab/tgl 在 dzx/、4gl 在 src/。
// 返回写出的 manifest，供调用方断言"材料侧状态"。
func fixMaterials(t *testing.T, matDir, prog string) *devdata.Manifest {
	t.Helper()
	files := map[string]string{
		"dzx/" + prog + ".tab": fixTab(prog),
		"dzx/" + prog + ".tgl": testutil.NormalTGL(prog),
		"dzx/" + prog + ".tap": testutil.NormalTAP(prog),
		"src/" + prog + ".4gl": testutil.Normal4GL(prog),
	}
	roleOf := func(rel string) string {
		switch filepath.Ext(rel) {
		case ".tab":
			return "tab"
		case ".tgl":
			return "tgl"
		case ".tap":
			return "tap"
		}
		return "4gl"
	}
	mf := devdata.NewManifest("fetch", "e")
	mf.Prog = prog
	mf.Module = "cxm"
	mf.FetchedAt = "2026-11-21T20:00:00+08:00"
	mf.Files = nil
	mf.Versions.DZAA002 = 3
	for rel, body := range files {
		p := filepath.Join(matDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("建材料目录: %v", err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("写材料 %s: %v", rel, err)
		}
		mf.AddFile(devdata.FileEntry{
			Role: roleOf(rel), Src: "erp/cxm/" + rel, Rel: rel,
			SHA256: devdata.BytesSHA256([]byte(body)), Size: int64(len(body)),
			SrcMtime: "2026-11-21T19:00:00+08:00",
		})
	}
	if err := devdata.WriteManifest(matDir, mf); err != nil {
		t.Fatalf("写材料清单: %v", err)
	}
	return mf
}

// fixMdl 在工作区里造 mdl/slice（与 Create 的 copyMdl 落点同形）。
func fixMdl(t *testing.T, root string) string {
	t.Helper()
	slice := filepath.Join(root, DirMdl, "slice")
	if err := os.MkdirAll(slice, 0o755); err != nil {
		t.Fatalf("建工作区模板切片: %v", err)
	}
	for _, name := range []string{"code_a05.template", "code_qs01.template"} {
		body := "#應用 a05 樣板自動產生(Version:3)\n"
		if err := os.WriteFile(filepath.Join(slice, name), []byte(body), 0o644); err != nil {
			t.Fatalf("写工作区模板切片: %v", err)
		}
	}
	return filepath.Join(root, DirMdl)
}

// fixPackageAt 在任意目录写一个正常包（"包放在别处"的回退用例用）。
func fixPackageAt(t *testing.T, dir, prog string) (string, error) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return testutil.WritePackage(dir, prog+"(c).tzc", testutil.NormalEntries(prog))
}

// fixIndex 用 devdata.Map 从材料建 index/link.json（与 Create 第 3 步同一实现）。
func fixIndex(t *testing.T, p Paths, prog string) *devdata.LinkDoc {
	t.Helper()
	doc, err := devdata.Map(devdata.LinkRequest{
		Env: "e", Prog: prog, Dir: p.Materials(), Out: p.Index(), Mdl: p.Mdl(),
	})
	if err != nil {
		t.Fatalf("devdata.Map（夹具建索引）: %v", err)
	}
	return doc
}

// fixPackage 在 pkg/ 下写一个合法 .tzc；entries 为 nil 时用 testutil.NormalEntries。
func fixPackage(t *testing.T, p Paths, prog string, entries map[string]string) string {
	t.Helper()
	if err := os.MkdirAll(p.PkgDir(), 0o755); err != nil {
		t.Fatalf("建包目录: %v", err)
	}
	if entries == nil {
		entries = testutil.NormalEntries(prog)
	}
	path, err := testutil.WritePackage(p.PkgDir(), prog+"(c).tzc", entries)
	if err != nil {
		t.Fatalf("造合成包: %v", err)
	}
	return path
}

// fixture 是一个"open 之后"的完整工作区。它等价于 Create 成功时的落盘结果，
// 只是材料由合成夹具直接写（Create 的拉料步骤必须联网，单测里替代为现造文件）。
type fixture struct {
	t        *testing.T
	Root     string
	Prog     string
	Env      string
	Cfg      string
	P        Paths
	Task     *Task
	Pkg      string
	Doc      *devdata.LinkDoc
	Manifest *devdata.Manifest
}

// newFixture 造完整工作区：材料 → 模板块 → 索引 → 包 → 视图 → task.json → notes。
func newFixture(t *testing.T) *fixture {
	t.Helper()
	root := t.TempDir()
	env := newFixEnv(t)
	f := &fixture{
		t: t, Root: root, Prog: fixProg, Env: "e", Cfg: env.Cfg,
		P: Paths{Root: root},
	}
	f.Manifest = fixMaterials(t, f.P.Materials(), f.Prog)
	fixMdl(t, root)
	f.Doc = fixIndex(t, f.P, f.Prog)
	f.Pkg = fixPackage(t, f.P, f.Prog, nil)
	if _, err := Export(ExportRequest{Dir: f.P.Ws(), Pkg: f.Pkg}); err != nil {
		t.Fatalf("Export（夹具建视图）: %v", err)
	}
	task := &Task{
		Schema: TaskSchema, Prog: f.Prog, Env: f.Env, Module: "cxm", Identity: "c",
		Task: "epi", Ent: 7, CreatedAt: nowStamp(), Layout: DefaultLayout(),
	}
	task.AddEvent("open", "夹具")
	if err := Save(task, f.P); err != nil {
		t.Fatalf("写 task.json: %v", err)
	}
	if err := os.MkdirAll(f.P.Notes(), 0o755); err != nil {
		t.Fatalf("建 notes/: %v", err)
	}
	f.Task = task
	return f
}

// status 现算一次工作区状态（失败直接终止用例）。
func (f *fixture) status() *Status {
	f.t.Helper()
	st, err := GetStatus(StatusRequest{Dir: f.Root})
	if err != nil {
		f.t.Fatalf("GetStatus: %v", err)
	}
	return st
}

// loadTask 重新读 task.json（断言事件/时间戳用）。
func (f *fixture) loadTask() *Task {
	f.t.Helper()
	task, _, err := Load(f.Root)
	if err != nil {
		f.t.Fatalf("Load: %v", err)
	}
	return task
}

// readTaskJSON 读 task.json 原始字节（断言"只读命令一个字都没写"用）。
func (f *fixture) readTaskJSON() []byte {
	f.t.Helper()
	b, err := os.ReadFile(f.P.TaskFile())
	if err != nil {
		f.t.Fatalf("读 task.json: %v", err)
	}
	return b
}

// removeMatManifest 删掉材料清单（造"材料不成套 / 没拉过"的状态）。
func (f *fixture) removeMatManifest() {
	f.t.Helper()
	if err := os.Remove(filepath.Join(f.P.Materials(), "manifest.json")); err != nil {
		f.t.Fatalf("删材料清单: %v", err)
	}
}

// corruptMatManifest 把材料清单写成非法 JSON。
func (f *fixture) corruptMatManifest() {
	f.t.Helper()
	if err := os.WriteFile(filepath.Join(f.P.Materials(), "manifest.json"), []byte("{不是 JSON"), 0o644); err != nil {
		f.t.Fatalf("写坏材料清单: %v", err)
	}
}

// writePkg 用给定条目覆盖本地包（模拟"apply 过"或"包被换过"）。
func (f *fixture) writePkg(entries map[string]string) {
	f.t.Helper()
	if err := os.MkdirAll(f.P.PkgDir(), 0o755); err != nil {
		f.t.Fatalf("建包目录: %v", err)
	}
	p, err := testutil.WritePackage(f.P.PkgDir(), f.Prog+"(c).tzc", entries)
	if err != nil {
		f.t.Fatalf("覆写包: %v", err)
	}
	f.Pkg = p
}

// dirtyView 在可编辑视图末尾追加一行（造"视图有未落盘改动"）。
func (f *fixture) dirtyView() {
	f.t.Helper()
	p := filepath.Join(f.P.Ws(), "prog.full.4gl")
	b, err := os.ReadFile(p)
	if err != nil {
		f.t.Fatalf("读视图: %v", err)
	}
	if err := os.WriteFile(p, append(b, []byte("\n# 本地随手一行\n")...), 0o644); err != nil {
		f.t.Fatalf("改视图: %v", err)
	}
}

// editMaterialTap 在材料 .tap 末尾加一行（造"材料变过、索引过期"）。
func (f *fixture) editMaterialTap() {
	f.t.Helper()
	rel, err := f.materialRel("tap")
	if err != nil {
		f.t.Fatal(err)
	}
	p := filepath.Join(f.P.Materials(), filepath.FromSlash(rel))
	b, rerr := os.ReadFile(p)
	if rerr != nil {
		f.t.Fatalf("读材料 tap: %v", rerr)
	}
	if werr := os.WriteFile(p, append(b, []byte("<!-- 改过 -->\n")...), 0o644); werr != nil {
		f.t.Fatalf("改材料 tap: %v", werr)
	}
}

// replaceManifest 用给定文件条目重写材料清单（造"清单里某些角色缺/文件对不上"）。
func (f *fixture) replaceManifest(files []devdata.FileEntry) {
	f.t.Helper()
	mf := devdata.NewManifest("fetch", "e")
	mf.Prog = f.Prog
	mf.Module = "cxm"
	mf.FetchedAt = f.Manifest.FetchedAt
	mf.Files = files
	if err := devdata.WriteManifest(f.P.Materials(), mf); err != nil {
		f.t.Fatalf("重写材料清单: %v", err)
	}
}

// filesOfRole 从夹具材料清单里挑出指定角色的条目。
func (f *fixture) filesOfRole(roles ...string) []devdata.FileEntry {
	want := map[string]bool{}
	for _, r := range roles {
		want[r] = true
	}
	var out []devdata.FileEntry
	for _, e := range f.Manifest.Files {
		if want[e.Role] {
			out = append(out, e)
		}
	}
	return out
}

// materialRel 从材料清单里取某个 role 的落点（顺带证明清单本身可读）。
func (f *fixture) materialRel(role string) (string, error) {
	for _, e := range f.Manifest.Files {
		if e.Role == role {
			return e.Rel, nil
		}
	}
	return "", os.ErrNotExist
}
