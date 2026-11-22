package task

// 离线工作区夹具：把 `tt dev open` 的产物（materials + index + pkg + ws + task.json）
// 用**合成材料**在 t.TempDir() 里手工拼出来。
//
// 为什么必须自己拼：命令层的 open/update 走 devdata.Fetch / devdata.Check，那两条
// 都要真 SFTP（FetchRequest.Files 的注入缝只到领域层，CLI 这层拿不到），所以在命令层
// 没法离线跑通 open 的成功路径。但 open 之后的全部读命令（status/brief/find/at/search）
// 以及 apply/verify/relink/unlock 只依赖磁盘上的工作区，于是**用同一条领域管线的
// 公开函数**（devdata.Pack / taskws.Export / devdata.Map / taskws.Save）造出等价产物，
// 就能把命令层的行为钉住，而完全不碰网络与真实 config.json。
//
// 材料内容刻意做成"最小但真实"：.tap 声明 + .tgl 骨架 + .4gl 产物 + .tab 蓝图 +
// .tsd 画面文案，覆盖客制点 / 标准点 / 孤儿声明 / 生成器自加槽位 / 切片标记 / 页签字段，
// 这样 brief/find/at/search 的每一条分支都有真实数据可走（而不是拿假结构体绕过）。

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/dev/store"
	"github.com/luyaogod/TT/internal/dev/taskws"
	"github.com/luyaogod/TT/internal/dev/testutil"
	"github.com/luyaogod/TT/internal/devdata"
	"github.com/luyaogod/TT/internal/testkit"
)

const (
	fxProg = "axmt500_wf"
	fxEnv  = "测试区"
)

// fx 是一个已经"open 完成"的工作区。
type fx struct {
	root   string
	prog   string
	env    string
	link   string // index/link.json
	pkg    string // pkg/<程序>(c).tzc
	mats   string // materials/
	ws     string // ws/
	task   *taskws.Task
	points []devdata.LinkPoint
}

// newFx 造一个可用的离线工作区。t.TempDir 每次全新，用例之间不共享。
func newFx(t *testing.T) *fx {
	t.Helper()
	root := t.TempDir()
	f := &fx{root: root, prog: fxProg, env: fxEnv}
	mats := filepath.Join(root, taskws.DirMaterials)
	writeFxMaterials(t, mats, fxProg)
	f.mats = mats

	// 1) 包：从材料本地打包（与 open 的第 4 步同一条管线）。
	pkgDir := filepath.Join(root, taskws.DirPkg)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("建包目录: %v", err)
	}
	pack, err := devdata.Pack(devdata.PackRequest{
		Env: fxEnv, Prog: fxProg, Identity: "c", Out: pkgDir, Force: true, Dir: mats,
	})
	if err != nil {
		t.Fatalf("打开展用的包失败（合成材料不成包？）: %v", err)
	}
	f.pkg = pack.Path

	// 2) 视图：从包导出围栏工作区（与 open 的第 5 步同一条管线）。
	f.ws = filepath.Join(root, taskws.DirWs)
	if _, err := taskws.Export(taskws.ExportRequest{Dir: f.ws, Pkg: f.pkg}); err != nil {
		t.Fatalf("导出围栏视图失败（合成包不是合法 tzc？）: %v", err)
	}

	// 3) task.json（open 的第 6 步）。
	f.task = &taskws.Task{
		Schema: taskws.TaskSchema, Prog: fxProg, Env: fxEnv, Module: "CXM",
		Identity: "c", CreatedAt: time.Now().Format(time.RFC3339),
		Layout: taskws.DefaultLayout(),
	}
	if err := taskws.Save(f.task, taskws.Paths{Root: root}); err != nil {
		t.Fatalf("写 task.json: %v", err)
	}

	// 4) 模板切片：只为造出"切片脱版"分支（产物用 a05 v3，当前模板 v5）。
	sliceDir := filepath.Join(root, taskws.DirMdl, "slice")
	if err := os.MkdirAll(sliceDir, 0o755); err != nil {
		t.Fatalf("建模板切片目录: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sliceDir, "code_a05.template"),
		[]byte("#應用 a05 樣板自動產生(Version:5)\n"), 0o644); err != nil {
		t.Fatalf("写模板切片: %v", err)
	}

	// 5) 索引（open 的第 3 步；显式给 Dir/Out/Mdl，不碰配置）。
	p := taskws.Paths{Root: root}
	f.link = p.Index()
	doc, err := devdata.Map(devdata.LinkRequest{
		Env: fxEnv, Prog: fxProg, Dir: mats, Out: f.link, Mdl: p.Mdl(),
	})
	if err != nil {
		t.Fatalf("建索引失败: %v", err)
	}
	f.points = doc.Points
	return f
}

// pointByName 取夹具里某个点位（不存在直接失败 —— 夹具坏了比用例假绿好）。
func (f *fx) pointByName(t *testing.T, name string) devdata.LinkPoint {
	t.Helper()
	for _, p := range f.points {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("夹具里没有点位 %q（现有 %d 个）", name, len(f.points))
	return devdata.LinkPoint{}
}

// lineInside 给一个落在某客制点区间内的行号（at 行号模式用）。
func (f *fx) lineInside(t *testing.T, name string) int {
	t.Helper()
	p := f.pointByName(t, name)
	if p.Line0 <= 0 {
		t.Fatalf("点位 %q 没有 .4gl 行号（Line0=%d）", name, p.Line0)
	}
	return p.Line0
}

// writeFxMaterials 写出一整套合成材料 + manifest（kind=fetch，与 fetch 落盘同形）。
func writeFxMaterials(t *testing.T, matDir, prog string) {
	t.Helper()
	if err := os.MkdirAll(matDir, 0o755); err != nil {
		t.Fatalf("建材料目录: %v", err)
	}
	files := map[string]string{
		prog + ".tgl": fxTGL(prog),
		prog + ".4gl": fx4GL(prog),
		prog + ".tap": fxTAP(prog),
		prog + ".tab": fxTAB(prog),
		prog + ".tsd": fxTSD(),
	}
	mf := devdata.NewManifest("fetch", fxEnv)
	mf.Prog, mf.Module = prog, "CXM"
	mf.FetchedAt = "2026-11-21T20:00:00+08:00"
	mf.Versions.DZAA002 = 7

	roles := []struct{ role, ext, srcDir string }{
		{"tab", ".tab", "dzx/tab"},
		{"tsd", ".tsd", "dzx/tsd"},
		{"tap", ".tap", "dzx/tap"},
		{"tgl", ".tgl", "dzx/tgl"},
		{"4gl", ".4gl", "dzx/src"},
	}
	for _, r := range roles {
		rel := prog + r.ext
		body := []byte(files[rel])
		if err := os.WriteFile(filepath.Join(matDir, rel), body, 0o644); err != nil {
			t.Fatalf("写材料 %s: %v", rel, err)
		}
		mf.AddFile(devdata.FileEntry{
			Role: r.role, Rel: rel, Src: "erp/cxm/" + r.srcDir + "/" + rel,
			SHA256: devdata.BytesSHA256(body), Size: int64(len(body)),
		})
	}
	if err := devdata.WriteManifest(matDir, mf); err != nil {
		t.Fatalf("写材料 manifest: %v", err)
	}
}

// fxTGL 是框架骨架：section 标记 + `#add-point` 占位符（含三个锚点区段）。
func fxTGL(prog string) string {
	return "#該程式未解開Section, 採用最新樣板產出!\n" +
		"#該程式非freestyle程式!\n" +
		"{<section id=\"" + prog + ".description\" type=\"s\" >}\n" +
		"#應用 a00 樣板自動產生(Version:1)\n" +
		"{</section>}\n" +
		"{<section id=\"" + prog + ".global\" type=\"s\" >}\n" +
		"#add-point:填寫註解說明 name=\"global.memo\"\n" +
		"{<point name=\"global.memo\" edit=\"s\"/>}\n" +
		"#end add-point\n" +
		"#add-point:增加匯入項目 name=\"global.import\"\n" +
		"{<point name=\"global.import\"/>}\n" +
		"#end add-point\n" +
		"IMPORT os\n" +
		"{</section>}\n" +
		"{<section id=\"" + prog + ".input\" type=\"s\" >}\n" +
		"#add-point:過欄位檢查 name=\"input.a.page73.xmdduc001\"\n" +
		"{<point name=\"input.a.page73.xmdduc001\" edit=\"c\"/>}\n" +
		"#end add-point\n" +
		"#add-point:開窗帶值 name=\"construct.c.page73.xmdduc002\"\n" +
		"{<point name=\"construct.c.page73.xmdduc002\" edit=\"c\"/>}\n" +
		"#end add-point\n" +
		"{</section>}\n" +
		"{<section id=\"" + prog + ".other_function\" readonly=\"Y\" type=\"s\" >}\n" +
		"{<point name=\"other.function\"/>}\n" +
		"{</section>}\n" +
		"{<section id=\"" + prog + ".other_dialog\" readonly=\"Y\" type=\"s\" >}\n" +
		"{<point name=\"other.dialog\"/>}\n" +
		"{</section>}\n" +
		"{<section id=\"" + prog + ".other_report\" readonly=\"Y\" type=\"s\" >}\n" +
		"{<point name=\"other.report\"/>}\n" +
		"{</section>}\n"
}

// fx4GL 是"服务器 build 产物"：与 TGL 同构，但多一个生成器自加槽位
// （input.b.page73.extra001，.tap 里没有对应的洞 → 包内也无声明 → at 的模板洞分支）、
// 并带一条 a05 v3 的切片标记（模板当前是 v5 → 脱版）。
func fx4GL(prog string) string {
	return "#該程式未解開Section, 採用最新樣板產出!\n" +
		"#該程式非freestyle程式!\n" +
		"#應用 a05 樣板自動產生(Version:3)\n" +
		"{<section id=\"" + prog + ".description\" type=\"s\" >}\n" +
		"#應用 a00 樣板自動產生(Version:1)\n" +
		"{</section>}\n" +
		"{<section id=\"" + prog + ".global\" type=\"s\" >}\n" +
		"#add-point:填寫註解說明 name=\"global.memo\"\n" +
		"{<point name=\"global.memo\" edit=\"s\"/>}\n" +
		"#end add-point\n" +
		"#add-point:增加匯入項目 name=\"global.import\"\n" +
		"{<point name=\"global.import\"/>}\n" +
		"#end add-point\n" +
		"IMPORT os\n" +
		"{</section>}\n" +
		"{<section id=\"" + prog + ".input\" type=\"s\" >}\n" +
		"#add-point:過欄位檢查 name=\"input.a.page73.xmdduc001\"\n" +
		"{<point name=\"input.a.page73.xmdduc001\" edit=\"c\"/>}\n" +
		"#end add-point\n" +
		"#add-point:開窗帶值 name=\"construct.c.page73.xmdduc002\"\n" +
		"{<point name=\"construct.c.page73.xmdduc002\" edit=\"c\"/>}\n" +
		"#end add-point\n" +
		"#add-point:生成器自加 name=\"input.b.page73.extra001\"\n" +
		"{<point name=\"input.b.page73.extra001\" edit=\"c\"/>}\n" +
		"#end add-point\n" +
		"{</section>}\n" +
		"{<section id=\"" + prog + ".other_function\" readonly=\"Y\" type=\"s\" >}\n" +
		"{<point name=\"other.function\"/>}\n" +
		"{</section>}\n"
}

// fxTAP 是设计源：根属性（identity=c / section_flag=N）+ 声明 + 区段占位符。
// 刻意留两个"只声明、没有洞"的点：function.axmt500_wf_calc（no-code 孤儿）与
// input.a.page73.gone001（no-slot 孤儿），供孤儿分支使用。
func fxTAP(prog string) string {
	hdr := "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?>\r\n" +
		"<add_points prog=\"" + prog + "\" std_prog=\"" + prog + "\" erpver=\"3.0\" module=\"CXM\" ver=\"1\"" +
		" env=\"c\" zone=\"topprd\" booking=\"N\" type=\"M\" identity=\"c\" section_flag=\"N\"" +
		" designer_ver=\"1.0\" login_user=\"tiptop\" std_section_verify=\"N\" std_to_cus=\"\"" +
		" pre_compile=\"Y\" topind=\"sd\">\r\n" +
		"  <other>\r\n" +
		"    <code_template value=\"F\" status=\"\"/>\r\n" +
		"    <free_style value=\"N\" status=\"\"/>\r\n" +
		"  </other>\r\n"
	decl := func(name, src, ver, body string) string {
		return "  <point name=\"" + name + "\" order=\"\" ver=\"" + ver + "\" cite_std=\"N\" new=\"Y\"" +
			" ind_fun=\"sd\" ind_extra=\"N\" ch=\"\" status=\"\" src=\"" + src + "\" readonly=\"\"" +
			" mark_hard=\"N\" modi_by_topstd=\"\">\r\n<![CDATA[" + body + "]]>\r\n  </point>\r\n"
	}
	sec := func(id, body string) string {
		return "  <section id=\"" + id + "\" src=\"s\" status=\"\" ver=\"1\" ch=\"\" readonly=\"\" modi_by_topstd=\"\">\r\n" +
			"<![CDATA[" + body + "]]>\r\n  </section>\r\n"
	}
	slot := func(label, name string) string {
		return "#add-point:" + label + " name=\"" + name + "\"\r\n" +
			"{<point name=\"" + name + "\" edit=\"c\"/>}\r\n#end add-point\r\n"
	}
	return hdr +
		decl("global.memo", "s", "1", "\r\n# 標準註解\r\n") +
		decl("input.a.page73.xmdduc001", "c", "2", "\r\nIF xmdduc001 = \"\" THEN\r\n  LET xmdduc001 = \"X\"\r\nEND IF\r\n") +
		decl("construct.c.page73.xmdduc002", "c", "3", "\r\nCALL cl_get_02(xmdduc002)\r\n") +
		decl("function."+prog+"_calc", "c", "4", "\r\nPRIVATE FUNCTION "+prog+"_calc(p_a)\r\n  RETURN p_a\r\nEND FUNCTION\r\n") +
		decl("input.a.page73.gone001", "c", "5", "\r\nLET gone001 = 1\r\n") +
		sec(prog+".description", "#應用 a00 樣板自動產生(Version:1)\r\n#+ Description: 合成測試程式\r\n") +
		sec(prog+".global", slot("填寫註解說明", "global.memo")+slot("增加匯入項目", "global.import")+"IMPORT os\r\n") +
		sec(prog+".input", slot("過欄位檢查", "input.a.page73.xmdduc001")+slot("開窗帶值", "construct.c.page73.xmdduc002")) +
		sec(prog+".other_function", "{<point name=\"other.function\"/>}\r\n") +
		"</add_points>\r\n"
}

// fxTAB 是蓝图：一个页签（page 1 / page_id 73），两个字段。
func fxTAB(prog string) string {
	return "<?xml version='1.0' encoding='UTF-8'?>\n" +
		"<assembly name=\"" + prog + "\" module=\"CXM\" type=\"t01\" industry=\"sd\"" +
		" page=\"1\" sdver=\"0001\" cusver=\"0003\" tabver=\"1\">\n" +
		"  <structure>\n" +
		"    <var id=\"body\" page=\"1\" page_id=\"73\" record=\"xmdduc_t\"" +
		" value=\"xmdduc001,xmdduc002\"/>\n" +
		"  </structure>\n" +
		"</assembly>\n"
}

// fxTSD 是画面文案：字段中文名 + 页签标题。
func fxTSD() string {
	return "<?xml version='1.0' encoding='UTF-8'?>\n" +
		"<strings>\n" +
		"  <sfield name=\"lbl_xmdduc001\" text=\"刻号\" />\n" +
		"  <sfield name=\"lbl_xmdduc002\" text=\"品名\" />\n" +
		"  <sfield name=\"page_73\" text=\"测试页签B\" />\n" +
		"  <sfield name=\"btn_ok\" text=\"确定\" />\n" +
		"</strings>\n"
}

// TestOfflineFixtureIsUsable 先给夹具本身一条实质断言：
// 合成材料必须真的能建出索引、能打成包、能导出视图，否则后面每一条用例都会变成
// "因为夹具坏了而失败"，分不清是产品问题还是测试问题。
func TestOfflineFixtureIsUsable(t *testing.T) {
	f := newFx(t)
	if len(f.points) == 0 {
		t.Fatal("夹具索引里一个点位都没有")
	}
	for _, name := range []string{
		"input.a.page73.xmdduc001", "construct.c.page73.xmdduc002",
		"global.memo", "input.b.page73.extra001",
	} {
		f.pointByName(t, name)
	}
	custom := f.pointByName(t, "input.a.page73.xmdduc001")
	if !custom.Custom || custom.DeclLines == 0 {
		t.Errorf("客制点应带 src=c 与声明行数: %+v", custom)
	}
	for _, p := range []string{
		filepath.Join(f.root, taskws.TaskFileName),
		filepath.Join(f.ws, "prog.full.4gl"),
		f.pkg,
		f.link,
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("工作区缺落点 %s: %v", p, err)
		}
	}
	// 索引必须新鲜：status 的"索引 最新"分支靠它。
	if fresh, err := devdata.IndexFresh(devdata.LinkRequest{
		Env: f.env, Prog: f.prog, Dir: f.mats, Out: f.link,
	}); err != nil || !fresh {
		t.Errorf("索引不新鲜：fresh=%v err=%v", fresh, err)
	}
}

// ---------- 命令驱动与输出捕获 ----------

// isolateConfig 把 TT_CONFIG 指到一个临时文件（exists=false 时该文件不存在），
// 保证被测命令**永远不会**回落到这台机器上真实的 config.json —— 那会让测试依赖
// 机器状态，也可能连到真实环境。返回这个临时路径。
func isolateConfig(t *testing.T, exists bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if exists {
		if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
			t.Fatalf("写临时 config.json: %v", err)
		}
	}
	t.Setenv("TT_CONFIG", path)
	return path
}

// setFormat 固定本次用例的输出形态（命令层直接读 common.JSON/Human 这两个包级开关，
// 没有 root 来绑定它们）。用例结束原样恢复，避免跨用例串味。
func setFormat(t *testing.T, human bool) {
	t.Helper()
	oldJSON, oldHuman, oldFormat := common.JSON, common.Human, common.Format
	oldEnv := common.Env
	common.JSON, common.Human, common.Format = !human, human, ""
	t.Cleanup(func() {
		common.JSON, common.Human, common.Format = oldJSON, oldHuman, oldFormat
		common.Env = oldEnv
	})
}

// runCmd 执行一个新建的命令并同时抓 stdout/stderr。
//
// 两条管道必须**并发排干**（见 internal/cli/dev/register_test.go 里那次死等事故）：
// 命令层直接写 os.Stdout/os.Stderr，串行读会与写端互相阻塞。
// 退出码按根命令的契约从 err 里解：带 ExitCode() 的按它，普通 err（cobra 的参数校验
// 错）等同于根命令的 os.Exit(1)。
func runCmd(t *testing.T, mk func() *cobra.Command, args ...string) (int, string, string, error) {
	t.Helper()
	var outBuf, errBuf strings.Builder
	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("建 stdout 管道: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("建 stderr 管道: %v", err)
	}
	os.Stdout, os.Stderr = wOut, wErr
	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(&outBuf, rOut); done <- struct{}{} }()
	go func() { _, _ = io.Copy(&errBuf, rErr); done <- struct{}{} }()

	cmd := mk()
	if args == nil {
		// 关键：空参数必须显式给空切片。cobra 的 SetArgs(nil) 会退回 os.Args[1:]，
		// 于是 `go test -coverprofile=...` 这类**测试二进制自己的参数**会被当成
		// 子命令参数（实测报 unknown command ".cover"）——普通 go test 看不出来，
		// 带 -coverprofile 跑就会红。
		args = []string{}
	}
	cmd.SetArgs(args)
	execErr := cmd.Execute()

	os.Stdout, os.Stderr = oldOut, oldErr
	_ = wOut.Close()
	_ = wErr.Close()
	<-done
	<-done
	_ = rOut.Close()
	_ = rErr.Close()
	return exitOf(execErr), outBuf.String(), errBuf.String(), execErr
}

// exitOf 按根命令 internal/cli.exitCodeOf 的契约取退出码。
func exitOf(err error) int {
	if err == nil {
		return 0
	}
	var ec interface{ ExitCode() int }
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return 1
}

// renderOutput 抓住一段直接 fmt.Printf 到 os.Stdout 的渲染输出。
//
// 它只是 testkit.CaptureStdout 的一层**别名**（把 (code,out) 收成 out，渲染用例不关心码）：
// 捕获实现单源在 internal/testkit，这里不另写管道（仓库的"测试助手单源"规则会扫函数名，
// 所以特意不叫 captureStdout）。
func renderOutput(t *testing.T, fn func()) string {
	t.Helper()
	_, out := testkit.CaptureStdout(t, func() int { fn(); return 0 })
	return out
}

// captureStderr 抓住一段直接写 os.Stderr 的输出（进度/警告都在 stderr）。
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("建 stderr 管道: %v", err)
	}
	done := make(chan string, 1)
	go func() {
		var b strings.Builder
		_, _ = io.Copy(&b, r)
		done <- b.String()
	}()
	old := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = old }()
	fn()
	_ = w.Close()
	out := <-done
	_ = r.Close()
	return out
}

// decodeEnvelope 把一条 `--json` 输出解析成信封 map（ok/env/data/… 平铺一层）。
func decodeEnvelope(t *testing.T, out string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &m); err != nil {
		t.Fatalf("输出不是合法 JSON：%v\n原文：%s", err, out)
	}
	return m
}

// dataOf 取信封里的 data 对象（本组所有成功信封都有一层 data）。
func dataOf(t *testing.T, env map[string]any) map[string]any {
	t.Helper()
	d, ok := env["data"].(map[string]any)
	if !ok {
		t.Fatalf("信封里没有 data 对象：%v", env)
	}
	return d
}

// num 把 JSON 解出来的数字（float64）取成 int。
func num(t *testing.T, v any) int {
	t.Helper()
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("不是数字：%T %v", v, v)
	}
	return int(f)
}

// replacePkg 用另一份合成包覆盖工作区的包（模拟"包被拷错/别人给了一份"）。
// 覆盖后 status 会报"与视图基线不一致"，apply/verify 拒绝 —— 正是 relink 要解的死局。
func replacePkg(t *testing.T, f *fx, entries map[string]string) {
	t.Helper()
	other, err := testutil.WritePackage(filepath.Join(t.TempDir(), "other"), "other.tzc", entries)
	if err != nil {
		t.Fatalf("造另一个包: %v", err)
	}
	raw, err := os.ReadFile(other)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.pkg, raw, 0o644); err != nil {
		t.Fatalf("覆盖工作区的包: %v", err)
	}
}

// reexportWsFromPkg 删掉视图、从当前包重新导出（用于把 env=s 的包变成当前工作区）。
func reexportWsFromPkg(t *testing.T, f *fx) {
	t.Helper()
	if err := os.RemoveAll(f.ws); err != nil {
		t.Fatalf("删旧视图: %v", err)
	}
	if _, err := taskws.Export(taskws.ExportRequest{Dir: f.ws, Pkg: f.pkg}); err != nil {
		t.Fatalf("从当前包重导视图: %v", err)
	}
}

// repointWsManifest 只把视图清单里记的"源包绝对路径"改到别处（包内容不动）。
// 这是"工作区被复制/移动"的最小形态：内容 sha 仍与基线一致，只是路径不再是本地包。
func repointWsManifest(t *testing.T, f *fx) {
	t.Helper()
	manifestPath := filepath.Join(f.ws, "manifest.json")
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("读视图清单: %v", err)
	}
	var m store.Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("解析视图清单: %v", err)
	}
	m.Pkg.Path = filepath.Join(t.TempDir(), "old", "elsewhere.tzc")
	out, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, out, 0o644); err != nil {
		t.Fatalf("改写视图清单: %v", err)
	}
}

// breakPkgBinding 造出"包被换过且路径指向别处"的形态（task 层 apply 的 RebindPackage 闸门）：
// 覆盖包内容（sha 对不上基线），再把清单路径挪走。
// 只覆盖包内容而不动路径是**另一种**形态：路径没变，RebindPackage 不报错，
// 由 tzc 自己的 sha 校验兜着 —— 两条不要混。
func breakPkgBinding(t *testing.T, f *fx) {
	t.Helper()
	replacePkg(t, f, testutil.NormalEntriesWith(fxProg, map[string]string{"env": "s"}))
	repointWsManifest(t, f)
}
