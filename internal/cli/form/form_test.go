package form

// WP2：`tt form` 命令面的**测试先行**（契约 C5 / C7）。
//
// 纪律（本文件是这套测试的地基）：
//   - **不连真机**：取件一律走假 `Source` 或本地临时目录（`--dir`），
//     真样本/真语料只在门控用例里出现（`verify_test.go`）；
//   - 真机取件实现（`internal/host/formsource.go`）只用纯函数测试（`hostfinder_test.go`）；
//   - `--json` 的输出形状由 golden 快照钉死（`golden_test.go`，牙齿 b）。

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/output"
	"tt/internal/testkit"
)

// ── 假件 ────────────────────────────────────────────────────────────────────

// fakeSource 是消费方接口（`Source`，契约 A8）的假实现：取件全部从内存里拿。
type fakeSource struct {
	forms map[string][]byte // 键：module + "/" + form
	names []string          // List 的返回值
	err   error             // 非 nil 时 Form/List 一起失败
}

func (f *fakeSource) Form(_ context.Context, _, module, form string) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	b, ok := f.forms[module+"/"+form]
	if !ok {
		return nil, fmt.Errorf("假件里没有 %s/%s", module, form)
	}
	return b, nil
}

func (f *fakeSource) List(_ context.Context, _, _ string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.names, nil
}

// injectSource 把本次用例的取件实现换成 fake，用例结束自动还回去。
//
// 取件的裁决点只有 `sourceFor` 一处（生产实现在 `productionSource`），所以这里
// 换一个函数就等于换掉了取件通道 —— 不必为了测试给命令加"依赖注入"参数。
func injectSource(t *testing.T, s Source) {
	t.Helper()
	old := sourceFor
	sourceFor = func(string, sourceOpts) (Source, error) { return s, nil }
	t.Cleanup(func() { sourceFor = old })
}

// ── 运行命令 ────────────────────────────────────────────────────────────────

// runForm 从**真实的命令树**跑一条 tt form 子命令，把 stdout 收回来。
//
// 捕获 stdout 走 testkit.CaptureStdout（边写边读的管道实现，仓内独此一份），
// flag 重置走 testkit.ResetFlags —— 命令树是包级单例，不重置会让上一条用例的
// `--form` 留到下一条（真机上一条命令一个进程，所以这不是产品缺陷）。
func runForm(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var runErr error
	_, out := testkit.CaptureStdout(t, func() int {
		root := newTestRoot()
		root.SetArgs(append([]string{"form"}, args...))
		runErr = root.Execute()
		testkit.ResetFlags(root)
		return 0
	})
	return out, runErr
}

// newTestRoot 造一棵最小但**忠实**的命令树：全局开关（--env/--json/--format）
// 与 internal/cli 的根命令同名同默认值，这样 `--json` / `--format table`
// 走的就是真路径（而不是测试里凭空多出的一套开关）。
func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "tt", SilenceUsage: true, SilenceErrors: true}
	pf := root.PersistentFlags()
	pf.StringVar(&common.Env, "env", "", "指定环境名")
	pf.BoolVar(&common.JSON, "json", false, "以 JSON 输出")
	pf.StringVar(&common.Format, "format", "json", "输出形态: json | table")
	root.AddCommand(Group)
	return root
}

// tempDataDir 造一份临时"数据目录"（config.json 旁边就是缓存目录）。
//
// `common.ConfigPath` 是包级 var（命令层的单例），用完必须还回去：不还的话下一条
// 用例会拿到这条留下的 t.TempDir()，而它早被删了（症状是"目录找得到、里面没了"）。
func tempDataDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	old := common.ConfigPath
	t.Cleanup(func() { common.ConfigPath = old })
	common.ConfigPath = filepath.Join(dir, "config.json")
	if err := os.WriteFile(common.ConfigPath, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// exitOf 取错误里的退出码（与 internal/cli 的 exitCodeOf 同一条链：ExitCode() int）。
func exitOf(t *testing.T, err error) int {
	t.Helper()
	var ec interface{ ExitCode() int }
	if !errors.As(err, &ec) {
		t.Fatalf("错误没带退出码（@json 下 agent 只能靠码分支）: %v", err)
	}
	return ec.ExitCode()
}

// codeOf 取错误里的稳定错误码（agent 靠它分支，不啃中文）。
func codeOf(t *testing.T, err error) string {
	t.Helper()
	var oe *output.Error
	if !errors.As(err, &oe) {
		t.Fatalf("错误不是 *output.Error（稳定码会丢）: %T %v", err, err)
	}
	return oe.Code
}

// fixtureBasic 是 internal/fep 的合成夹具（几百字节，随仓库走）。
// 复用它而不是在两边各写一份 XML：`tt form show` 的解析结果就是 `fep` 的解析结果，
// 两份夹具迟早会漂移成"看起来对、其实对不上"。
func fixtureBasic(t *testing.T) string {
	t.Helper()
	p := filepath.Join("..", "..", "fep", "testdata", "form42f", "basic.xml")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("读合成夹具失败: %v", err)
	}
	return string(b)
}

// ── show ────────────────────────────────────────────────────────────────────

func TestFormShowUsesSource(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{forms: map[string][]byte{
		"demomod/demoform": []byte(fixtureBasic(t)),
	}})

	out, err := runForm(t, "show", "--env", "demoenv", "--module", "demomod", "--form", "demoform", "--json")
	if err != nil {
		t.Fatalf("show 失败: %v", err)
	}
	doc := decodeShow(t, out)
	if doc["env"] != "demoenv" || doc["module"] != "demomod" || doc["form"] != "demoform" {
		t.Errorf("信封头的三件不对: %v %v %v", doc["env"], doc["module"], doc["form"])
	}
	if doc["variant"] != "gui" {
		t.Errorf("variant = %v, want gui（C3.8 暂定：文件名后缀，认不出就是 gui）", doc["variant"])
	}
	if doc["build"] != "3.21.03" {
		t.Errorf("build = %v, want 3.21.03（来自 <Form build=…>）", doc["build"])
	}
	fields, _ := doc["fields"].([]any)
	if len(fields) != 3 {
		t.Fatalf("fields = %d 条, want 3", len(fields))
	}
	first, _ := fields[0].(map[string]any)
	if first["colName"] != "demo001" || first["fieldId"] != float64(0) || first["tag"] != "FormField" {
		t.Errorf("fields[0] = %v", first)
	}
}

func TestFormShowTextIsReadable(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{forms: map[string][]byte{
		"demomod/demoform": []byte(fixtureBasic(t)),
	}})

	out, err := runForm(t, "show", "--env", "demoenv", "--module", "demomod",
		"--form", "demoform", "--format", "table")
	if err != nil {
		t.Fatalf("show 失败: %v", err)
	}
	if strings.Contains(out, "\"colName\"") {
		t.Errorf("--format table 不该是 JSON:\n%s", out)
	}
	for _, want := range []string{"demo001", "demo002", "demo003", "TableColumn", "demo_t"} {
		if !strings.Contains(out, want) {
			t.Errorf("人读输出里缺 %q:\n%s", want, out)
		}
	}
}

func TestFormShowFetchFailureIsNonZero(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{err: errors.New("SFTP: no such file")})

	_, err := runForm(t, "show", "--env", "demoenv", "--module", "demomod", "--form", "demoform")
	if err == nil {
		t.Fatal("取件失败必须非 0 退出，不能退 0")
	}
	if got := exitOf(t, err); got != exitFormUnknown {
		t.Errorf("退出码 = %d, want %d（FORM_UNKNOWN，契约 C4.2）", got, exitFormUnknown)
	}
	if got := codeOf(t, err); got != CodeFormUnknown {
		t.Errorf("错误码 = %q, want %q", got, CodeFormUnknown)
	}
	if !strings.Contains(err.Error(), "demoform") {
		t.Errorf("错误应点名是哪张表单取不到: %v", err)
	}
}

func TestFormShowRejectsBrokenForm(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{forms: map[string][]byte{
		"demomod/demoform": []byte("这不是 XML，只是一段文本"),
	}})

	_, err := runForm(t, "show", "--env", "demoenv", "--module", "demomod", "--form", "demoform")
	if err == nil {
		t.Fatal("不是合法 .42f 必须非 0 退出")
	}
	if got := exitOf(t, err); got != exitFormUnknown {
		t.Errorf("退出码 = %d, want %d", got, exitFormUnknown)
	}
	if !strings.Contains(err.Error(), "不是 XML") {
		t.Errorf("错误该说清「不是 XML」: %v", err)
	}
}

func TestFormShowRequiresEnvModuleForm(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{})

	for _, args := range [][]string{
		{"show", "--module", "demomod", "--form", "demoform"},
		{"show", "--env", "demoenv", "--form", "demoform"},
		{"show", "--env", "demoenv", "--module", "demomod"},
	} {
		_, err := runForm(t, args...)
		if err == nil {
			t.Fatalf("%v 缺必填参数必须报错", args)
		}
		if got := exitOf(t, err); got != exitUsage {
			t.Errorf("%v 的退出码 = %d, want %d（用法错）", args, got, exitUsage)
		}
	}
}

// ── ls ──────────────────────────────────────────────────────────────────────

func TestFormLsUsesSource(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{names: []string{"axmt500_wf", "axmt501_wf"}})

	out, err := runForm(t, "ls", "--env", "demoenv", "--module", "axm", "--format", "table")
	if err != nil {
		t.Fatalf("ls 失败: %v", err)
	}
	for _, want := range []string{"axmt500_wf", "axmt501_wf"} {
		if !strings.Contains(out, want) {
			t.Errorf("ls 输出里缺 %q:\n%s", want, out)
		}
	}

	jout, err := runForm(t, "ls", "--env", "demoenv", "--module", "axm", "--json")
	if err != nil {
		t.Fatalf("ls --json 失败: %v", err)
	}
	if !strings.Contains(jout, "\"forms\"") || !strings.Contains(jout, "axmt500_wf") {
		t.Errorf("ls --json 的形状不对:\n%s", jout)
	}
}

func TestFormLsFetchFailureIsNonZero(t *testing.T) {
	tempDataDir(t)
	injectSource(t, &fakeSource{err: errors.New("列目录失败")})

	_, err := runForm(t, "ls", "--env", "demoenv", "--module", "axm")
	if err == nil {
		t.Fatal("列不到表单必须非 0 退出")
	}
	if got := exitOf(t, err); got != exitFormUnknown {
		t.Errorf("退出码 = %d, want %d", got, exitFormUnknown)
	}
}

// ── 组本身 ──────────────────────────────────────────────────────────────────

func TestFormUnknownSubcommandIsNonZero(t *testing.T) {
	out, err := runForm(t, "nope")
	if err == nil {
		t.Fatalf("不认识的子命令必须非 0（实测过 `tt dict nope` 退 0 的那种失败形态）: %s", out)
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("错误该点名那个子命令: %v", err)
	}
}

// ── 小工具与契约常量 ────────────────────────────────────────────────────────

func TestVariantOf(t *testing.T) {
	cases := map[string]string{
		"axmt500_wf":  "gui", // 契约 C5.2 的样例就是 gui（C3.8 待核：真机若有更权威来源以真机为准）
		"fglfind_tui": "tui",
		"demoform":    "gui",
		"fglfind":     "gui",
	}
	for in, want := range cases {
		if got := VariantOf(in); got != want {
			t.Errorf("VariantOf(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWindowOf(t *testing.T) {
	cases := map[string]string{
		"axmt500_wf":  "w_axmt500",
		"fglfind_tui": "w_fglfind",
		"apmt500_wf":  "w_apmt500",
	}
	for in, want := range cases {
		if got := WindowOf(in); got != want {
			t.Errorf("WindowOf(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestGroupRegistersVerbs 钉住命令面本身：`tt form` 三个动词一个都不能少。
// 命令组挂在根命令上（`internal/cli/root.go` 里一行 Register），这条在**本包**里
// 用一棵最小的根命令验，避免"注册漏了但没有任何测试会红"。
func TestGroupRegistersVerbs(t *testing.T) {
	root := &cobra.Command{Use: "tt"}
	Register(root)

	forms := ""
	for _, c := range root.Commands() {
		if c.Name() == "form" {
			forms = c.Name()
		}
	}
	if forms != "form" {
		t.Fatalf("Register 没把 form 挂到根命令上: %v", root.Commands())
	}
	got := map[string]bool{}
	for _, c := range Group.Commands() {
		got[c.Name()] = true
	}
	for _, want := range []string{"ls", "show", "verify"} {
		if !got[want] {
			t.Errorf("tt form 缺子命令 %q（现有：%v）", want, got)
		}
	}
}

func TestExitCodesAreTheFrozenOnes(t *testing.T) {
	// 码表（本组）：0 成功 / 1 FORM_UNKNOWN（取件·解析失败）/ 2 用法错 /
	// 6 体检不过（verify 与语料不一致）。改码必须先改契约。
	if exitOK != 0 || exitFormUnknown != 1 || exitUsage != 2 || exitMismatch != 6 {
		t.Errorf("退出码被改了: ok=%d unknown=%d usage=%d mismatch=%d",
			exitOK, exitFormUnknown, exitUsage, exitMismatch)
	}
	if CodeFormUnknown != "FORM_UNKNOWN" {
		t.Errorf("错误码 = %q, want FORM_UNKNOWN（契约 C4.2 的冻结名）", CodeFormUnknown)
	}
}
