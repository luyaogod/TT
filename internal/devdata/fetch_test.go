package devdata

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newFetchFixture 造一套 axmt500_wf 的服务器文件。
func newFetchFixture(t *testing.T) (cfgPath string, files *fakeFiles) {
	t.Helper()
	cfgPath = writeConfig(t, hostStub)
	files = newFakeFiles()
	files.dirs["erp/axm/dzx"] = true
	files.files["erp/axm/dzx/tab/axmt500_wf.tab"] = []byte(`<table name="axmt500_wf">&#xA;</table>`)
	files.files["erp/axm/dzx/tsd/axmt500_wf.tsd"] = []byte("<tsd/>")
	files.files["erp/axm/dzx/tgl/axmt500_wf.tgl"] = []byte("#+ Standard Version.....: SD版次:0025(2025-11-13 17:33:55), PR版次:\n" +
		"#+ Build......: 000037\n#+ Description: 訂單維護作業\n")
	files.files["erp/axm/dzx/tgl/axmt500_wf.tgx"] = []byte("tgx")
	files.files["erp/axm/dzx/tgl/axmt500_wf.tgx2"] = []byte("tgx2")
	files.files["erp/axm/4gl/axmt500_wf.4gl"] = []byte("FUNCTION axmt500_wf()\nEND FUNCTION\n")
	files.files["com/inc/erp/axm/axmt500_wf_mask.4gl"] = []byte("mask")
	files.files["erp/axm/42f/axmt500_wf.42f"] = []byte("\x00\x01binary")
	return cfgPath, files
}

// TestFetchHappyPath 一次成功拉取：文件齐全、manifest 的 sha 与实际文件一致。
func TestFetchHappyPath(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	res, err := Fetch(context.Background(), FetchRequest{
		CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files}, nil)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if res.Module != "axm" {
		t.Errorf("模块推断 = %q，应为 axm", res.Module)
	}
	progDir, _ := ProgDir(cfgPath, "福联测试tst", "axmt500_wf")

	// 每个落盘文件的字节必须与服务器一致（原样保真），manifest 的 sha 必须能对上
	m, err := ReadManifest(progDir)
	if err != nil || m == nil {
		t.Fatalf("manifest: %v", err)
	}
	if len(m.Files) != 7 { // tab tsd tgl tgx tgx2 4gl inc(mask)；_rep 与 tap 不存在
		t.Errorf("落盘文件数 = %d，应为 7: %+v", len(m.Files), m.Files)
	}
	for _, f := range m.Files {
		p := filepath.Join(progDir, filepath.FromSlash(f.Rel))
		if _, err := os.Stat(p); err != nil {
			t.Errorf("manifest 登记了 %s，但文件不存在", f.Rel)
			continue
		}
		sha, err := FileSHA256(p)
		if err != nil || sha != f.SHA256 {
			t.Errorf("%s 的 sha 对不上（manifest %s，实际 %s, %v）", f.Rel, f.SHA256, sha, err)
		}
	}

	// 保真：tab 里的实体与换行没被改动
	b, _ := os.ReadFile(filepath.Join(progDir, "dzx", "axmt500_wf.tab"))
	if string(b) != `<table name="axmt500_wf">&#xA;</table>` {
		t.Errorf("tab 被改写了: %q", b)
	}
	// 可选缺失进 missing，不判失败
	if !containsStr(res.Missing, "dzx/axmt500_wf.tap") || !containsStr(res.Missing, "inc/axmt500_wf_rep.4gl") {
		t.Errorf("可选缺失应进 missing: %v", res.Missing)
	}
	// 默认不拉 .42f
	if _, err := os.Stat(filepath.Join(progDir, "form", "axmt500_wf.42f")); err == nil {
		t.Error(".42f 默认不该拉")
	}
	// tgl 头部版本解析
	if m.Versions.TglHeader == nil || m.Versions.TglHeader.SDVer != "0025(2025-11-13 17:33:55)" {
		t.Errorf("tglHeader = %+v", m.Versions.TglHeader)
	}
	// 共用资料没同步 → 版本锚点缺失要开口，但不是失败
	if len(m.Warnings) == 0 {
		t.Error("没有版本锚点时应给 warning")
	}
	// 本地一致性：刚拉完不该 stale
	if reasons := LocalStale(progDir, m); len(reasons) != 0 {
		t.Errorf("刚拉完不该 stale: %+v", reasons)
	}
	// 无临时残留
	if left := TempDirs(filepath.Dir(progDir), "axmt500_wf"); len(left) != 0 {
		t.Errorf("还有临时目录: %v", left)
	}
}

// TestFetchRequiredMissingFails 必需素材（蓝图）缺失 = 整套失败，且不留半套。
//
// 同时钉住错误语义：缺必需素材必须能被判成"这个程序没有材料"（ErrProgNotFound → 404），
// 而不是一条普通的 500 —— 调用方要能据它区分"程序名写错了/没产过"与"网络挂了"。
func TestFetchRequiredMissingFails(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	delete(files.files, "erp/axm/dzx/tab/axmt500_wf.tab")
	_, err := Fetch(context.Background(), FetchRequest{
		CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files}, nil)
	if err == nil {
		t.Fatal("缺 tab 应失败")
	}
	if !strings.Contains(err.Error(), "必需素材") || !strings.Contains(err.Error(), "axmt500_wf") {
		t.Errorf("错误应点名缺的素材: %v", err)
	}
	if !errors.Is(err, ErrProgNotFound) {
		t.Errorf("缺必需素材应包 ErrProgNotFound（→ 404），得到 %v", err)
	}
	progDir, _ := ProgDir(cfgPath, "福联测试tst", "axmt500_wf")
	if _, serr := os.Stat(progDir); serr == nil {
		t.Error("失败后不该留下程序目录")
	}
	if left := TempDirs(filepath.Dir(progDir), "axmt500_wf"); len(left) != 0 {
		t.Errorf("失败后还有临时目录: %v", left)
	}
}

// TestFetchUnknownProgram 整个程序都不存在（tab/tgl/4gl 全没有）→ 明确说"找不到程序"。
func TestFetchUnknownProgram(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	for _, k := range []string{"erp/axm/dzx/tab/axmt500_wf.tab", "erp/axm/dzx/tgl/axmt500_wf.tgl", "erp/axm/4gl/axmt500_wf.4gl"} {
		delete(files.files, k)
	}
	// 模块目录还在（说明模块装过），但这个程序一条材料都没有
	_, err := Fetch(context.Background(), FetchRequest{
		CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files}, nil)
	if !errors.Is(err, ErrProgNotFound) {
		t.Fatalf("应报 ErrProgNotFound，得到 %v", err)
	}
	if !strings.Contains(err.Error(), "找不到程序") {
		t.Errorf("错误应说清是程序不存在: %v", err)
	}
}

// TestFetchModuleResolution 模块推断：推断对了就用，推断错了要给出可操作的提示。
func TestFetchModuleResolution(t *testing.T) {
	_, files := newFetchFixture(t)

	// ① 显式指定模块：直接用
	if mod, err := resolveModule(context.Background(), files, "axmt500_wf", "AXM"); err != nil || mod != "axm" {
		t.Fatalf("显式模块 = %q, %v", mod, err)
	}
	// ② 非法模块名
	if _, err := resolveModule(context.Background(), files, "axmt500_wf", "../x"); err == nil {
		t.Fatal("非法模块名应被拒绝")
	}
	// ③ 推断出来的模块下没有蓝图 → 报"请用 --module"
	if _, err := resolveModule(context.Background(), files, "zzz500_wf", ""); err == nil {
		t.Fatal("推断不出模块时应报错")
	} else if !errors.Is(err, ErrModuleUnknown) {
		t.Fatalf("错误应包 ErrModuleUnknown: %v", err)
	}
	// ④ 隐式推断成功
	if mod, err := resolveModule(context.Background(), files, "axmt500_wf", ""); err != nil || mod != "axm" {
		t.Fatalf("推断模块 = %q, %v", mod, err)
	}
}

// TestFetchExistsNeedsForce 重复拉取要 --force（整包替换是显式动作）。
func TestFetchExistsNeedsForce(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	req := FetchRequest{CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files}
	if _, err := Fetch(context.Background(), req, nil); err != nil {
		t.Fatalf("第一次: %v", err)
	}
	if _, err := Fetch(context.Background(), req, nil); !errors.Is(err, ErrExists) {
		t.Fatalf("第二次应报 ErrExists，得到 %v", err)
	}
	req.Force = true
	if _, err := Fetch(context.Background(), req, nil); err != nil {
		t.Fatalf("force 应当成功: %v", err)
	}
}

// TestFetchWithForm --form 才拉 .42f。
func TestFetchWithForm(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	res, err := Fetch(context.Background(), FetchRequest{
		CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Form: true, Files: files}, nil)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	progDir, _ := ProgDir(cfgPath, "福联测试tst", "axmt500_wf")
	b, err := os.ReadFile(filepath.Join(progDir, "form", "axmt500_wf.42f"))
	if err != nil || string(b) != "\x00\x01binary" {
		t.Fatalf("二进制没原样落盘: %q, %v", b, err)
	}
	if !containsRel(res.Files, "form/axmt500_wf.42f") {
		t.Errorf("结果里应有 form 条目: %+v", res.Files)
	}
}

// TestFetchWithDesign --with-design 归档本地工作区的 `.tzs`（只复制、不解析）。
func TestFetchWithDesign(t *testing.T) {
	ws := t.TempDir()
	design := []byte("<tzs>design</tzs>")
	if err := os.WriteFile(filepath.Join(ws, "axmt500_wf(c).tzs"), design, 0o644); err != nil {
		t.Fatal(err)
	}
	cfgPath := writeConfig(t, `{"name":"福联测试tst","host":"10.0.0.1","zone":"2","workspace":`+
		quote(ws)+`}`)
	_, files := newFetchFixture(t)
	res, err := Fetch(context.Background(), FetchRequest{
		CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", WithDesign: true, Files: files}, nil)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	progDir, _ := ProgDir(cfgPath, "福联测试tst", "axmt500_wf")
	b, err := os.ReadFile(filepath.Join(progDir, "design", "axmt500_wf(c).tzs"))
	if err != nil || string(b) != string(design) {
		t.Fatalf("设计稿没归档: %q, %v", b, err)
	}
	var ok bool
	for _, f := range res.Files {
		if f.Role == "design" {
			ok = true
		}
	}
	if !ok {
		t.Errorf("结果里应有 design 条目: %+v", res.Files)
	}
}

// TestFetchVersionAnchors 共用资料在的时候，版本锚点要能取到（模板版本 + DZAA 版次）。
func TestFetchVersionAnchors(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	// 先造一份共用资料的 manifest（不必真拉表）
	envDir, _ := EnvDir(cfgPath, "福联测试tst")
	m := NewManifest(KindMirror, "福联测试tst")
	m.FetchedAt = "2026-11-21T10:00:00+08:00"
	m.Versions.TemplateVersions["a05"] = 3
	m.Versions.TemplateVersions["i02"] = 12
	if err := WriteManifest(envDir, m); err != nil {
		t.Fatal(err)
	}
	res, err := Fetch(context.Background(), FetchRequest{
		CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files}, nil)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if res.TemplateVersions["a05"] != 3 || res.TemplateVersions["i02"] != 12 {
		t.Errorf("模板版本锚点 = %v", res.TemplateVersions)
	}
	progDir, _ := ProgDir(cfgPath, "福联测试tst", "axmt500_wf")
	pm, _ := ReadManifest(progDir)
	if pm.Versions.TemplateVersions["a05"] != 3 {
		t.Errorf("程序 manifest 里应带模板版本: %v", pm.Versions.TemplateVersions)
	}
}

// TestFetchCancelNoResidue 取消后不落半套。
func TestFetchCancelNoResidue(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	progDir, _ := ProgDir(cfgPath, "福联测试tst", "axmt500_wf")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立刻取消：连第一个文件都不该落盘
	_, err := Fetch(ctx, FetchRequest{CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("取消应返回 context.Canceled，得到 %v", err)
	}
	if _, serr := os.Stat(progDir); serr == nil {
		t.Error("取消后不该留下程序目录")
	}
	if left := TempDirs(filepath.Dir(progDir), "axmt500_wf"); len(left) != 0 {
		t.Errorf("取消后还有临时目录: %v", left)
	}
}

// containsRel 判文件清单里有没有某个相对路径。
func containsRel(list []FileEntry, rel string) bool {
	for _, f := range list {
		if f.Rel == rel {
			return true
		}
	}
	return false
}

// quote 把字符串包成 JSON 字符串（测试里拼配置用）。
func quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
