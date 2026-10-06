package drawio

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"tt/internal/testkit"
)

// TestCobraRouting 钉住命令组这一层的路由与退出码契约。
//
// 为什么值得一条：组命令没有 RunE 时，cobra 会打一份帮助就**退 0** —— 调用方会把
// "我打错了命令"读成成功，而这是最难自己发现的失败形态（`tt dict nope` 实测过一次）。
// 每次用 newGroup() 造新树，避免 cobra/pflag 的命令状态在用例间串味。
func TestCobraRouting(t *testing.T) {
	t.Run("裸命令打帮助退0", func(t *testing.T) {
		code, out, _ := runGroup(t)
		if code != 0 {
			t.Errorf("退出码 = %d，期望 0", code)
		}
		if !strings.Contains(out, "Usage:") {
			t.Errorf("裸命令该打帮助，得到：%s", out)
		}
	})

	t.Run("未知子命令退1并列出可用", func(t *testing.T) {
		code, _, errText := runGroup(t, "nope")
		if code != 1 {
			t.Errorf("退出码 = %d，期望 1（与顶层「用法错」一致）", code)
		}
		if !strings.Contains(errText, `未知子命令 "nope"`) {
			t.Errorf("该报未知子命令，得到：%s", errText)
		}
		if !strings.Contains(errText, "lib") {
			t.Errorf("该把可用子命令列出来，得到：%s", errText)
		}
	})
}

// TestLibCatalogToStdout 钉住 --catalog：不落盘、只打清单，且清单里认得出形状。
func TestLibCatalogToStdout(t *testing.T) {
	code, out, _ := runGroup(t, "lib", "--catalog")
	if code != 0 {
		t.Fatalf("退出码 = %d，期望 0", code)
	}
	for _, want := range []string{"# T100 组件库 · 形状清单", "## controls", "## business",
		"`ButtonEdit`", "`Table`", "文字槽位"} {
		if !strings.Contains(out, want) {
			t.Errorf("清单里缺 %q", want)
		}
	}
}

// TestLibWritesLoadableLibraries 钉住导出：文件落盘后**按 drawio 的方式能读回来**。
//
// 只断言"文件存在"是不够的 —— 库文件坏掉的典型形态正是"文件在、drawio 打不开"
// （转义少一层），所以这里走一遍 parseXml → JSON.parse 那两步。
func TestLibWritesLoadableLibraries(t *testing.T) {
	dir := t.TempDir()
	code, _, errText := runGroup(t, "lib", "-o", dir)
	if code != 0 {
		t.Fatalf("退出码 = %d，期望 0（%s）", code, errText)
	}

	xmls, err := filepath.Glob(filepath.Join(dir, "*.xml"))
	if err != nil || len(xmls) != 2 {
		t.Fatalf("期望导出 2 个库文件，得到 %v（%v）", xmls, err)
	}

	total := 0
	for _, p := range xmls {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("读不了 %s：%v", p, err)
		}
		var root struct {
			Inner string `xml:",chardata"`
		}
		if err := xml.Unmarshal(b, &root); err != nil {
			t.Fatalf("%s 不是合法 XML：%v", p, err)
		}
		var entries []map[string]any
		if err := json.Unmarshal([]byte(root.Inner), &entries); err != nil {
			t.Fatalf("%s 的文本节点不是合法 JSON（drawio 会拒收）：%v", p, err)
		}
		for i, e := range entries {
			if s, _ := e["xml"].(string); s == "" {
				t.Errorf("%s 第 %d 条的 xml 是空的", p, i)
			}
		}
		total += len(entries)
	}
	if total != 27 {
		t.Errorf("两个库共 %d 个形状，期望 27（controls 18 + business 9）", total)
	}
}

// TestLibDefaultsToDistUnderCwd 钉住缺省落点：**当前目录下的 dist/**。
//
// 必须换目录再跑 —— 测的就是"相对于谁"。不换的话它会把库写进包目录，
// 那既污染了源码树，也让这条断言失去意义（包目录下正好也有个 dist 的话就假绿了）。
func TestLibDefaultsToDistUnderCwd(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	code, _, errText := runGroup(t, "lib")
	if code != 0 {
		t.Fatalf("退出码 = %d，期望 0（%s）", code, errText)
	}
	xmls, _ := filepath.Glob(filepath.Join(dir, outDirName, "*.xml"))
	if len(xmls) != 2 {
		t.Errorf("期望 <当前目录>/%s 下拿到 2 个库文件，得到 %v", outDirName, xmls)
	}
}

// TestComposeWritesFile 钉住 compose 这条路的出口：文件写出来了，而且**能被 XML 解析器读回来**。
//
// 只断言"文件存在"不够：compose 最典型的坏产物是"文件在、drawio 打不开"（转义或
// 引用出了错），所以这里走一遍真解析。
func TestComposeWritesFile(t *testing.T) {
	spec := filepath.Join(t.TempDir(), "spec.json")
	if err := os.WriteFile(spec, []byte(`{"title":"t","items":[{"shape":"ButtonEdit","col":0,"row":0,"text":{"label":"供应商"}}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "x.drawio")

	code, _, errText := runGroup(t, "compose", spec, "-o", out)
	if code != 0 {
		t.Fatalf("退出码 = %d，期望 0（%s）", code, errText)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("产物没写出来：%v", err)
	}
	var root struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(b, &root); err != nil {
		t.Fatalf("产物不是良构 XML：%v", err)
	}
	if root.XMLName.Local != "mxfile" {
		t.Errorf("根元素该是 mxfile，得到 %s", root.XMLName.Local)
	}
	if !strings.Contains(string(b), `value="供应商"`) {
		t.Error("文字槽位没被写进产物")
	}
}

// TestComposeExitCodes 钉住这条线的退出码分类 —— 脚本靠它区分"我写错了"与"工具坏了"。
func TestComposeExitCodes(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.json")
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte(`{"items":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		args []string
		want int
	}{
		{"文件不存在退2", []string{"compose", missing, "-o", filepath.Join(dir, "a")}, 2},
		{"spec 不合法退2", []string{"compose", bad, "-o", filepath.Join(dir, "b")}, 2},
		{"--json 不带 -o 退1", []string{"compose", bad, "--json"}, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if code, _, _ := runGroup(t, c.args...); code != c.want {
				t.Errorf("退出码 = %d，期望 %d", code, c.want)
			}
		})
	}
}

// TestComposeToStdout 钉住不带 -o 时的形状：.drawio 走 stdout，摘要走 stderr ——
// 这样 `tt drawio compose spec.json > x.drawio` 直接可用。
func TestComposeToStdout(t *testing.T) {
	spec := filepath.Join(t.TempDir(), "s.json")
	if err := os.WriteFile(spec, []byte(`{"items":[{"shape":"Label","col":0,"row":0}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out, _ := runGroup(t, "compose", spec)
	if code != 0 {
		t.Fatalf("退出码 = %d，期望 0", code)
	}
	if !strings.HasPrefix(out, "<?xml version=\"1.0\"") {
		t.Errorf("stdout 该以 XML 声明开头，得到：%.60s", out)
	}
	if strings.Contains(out, "个控件 /") {
		t.Error("摘要不该混进 stdout —— 它会污染重定向出来的 .drawio")
	}
}

// TestLibIsReproducible 钉住"同一份形状源产出同一份字节"。
//
// 形状库里没有任何时钟或随机源，所以两次导出必须逐字节相同 —— 这条一红就说明
// 有东西把构建环境漏进了产物（原项目就栽在生成物里写死本机绝对路径上）。
func TestLibIsReproducible(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	for _, dir := range []string{a, b} {
		if code, _, errText := runGroup(t, "lib", "-o", dir); code != 0 {
			t.Fatalf("导出到 %s 失败（%d）：%s", dir, code, errText)
		}
	}
	xmls, _ := filepath.Glob(filepath.Join(a, "*.xml"))
	if len(xmls) == 0 {
		t.Fatal("没导出任何文件")
	}
	for _, p := range xmls {
		x, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		y, err := os.ReadFile(filepath.Join(b, filepath.Base(p)))
		if err != nil {
			t.Fatalf("第二次没产出同名文件 %s：%v", filepath.Base(p), err)
		}
		if string(x) != string(y) {
			t.Errorf("%s 两次导出不一致", filepath.Base(p))
		}
	}
}

// runGroup 用一棵**新树**跑一条命令，返回 (退出码, stdout, 错误文本)。
//
// 必须把组挂在一个假根下面，不能直接当根执行：cobra 的 legacyArgs 只在命令
// **没有父级**时拦"未知命令"，直接当根跑会拿到 cobra 那句 `unknown command`，
// 而真实调用路径上（组挂在 rootCmd 下）走的是我们自己的 RunE。
//
// 退出码的算法与根命令的 exitCodeOf 一致：错误带 ExitCode() 就用它，否则算 1。
// SilenceErrors/SilenceUsage 也照根命令设 —— 根命令是这么配的，测试要跑同一套。
func runGroup(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var errText string
	code, out := testkit.CaptureStdout(t, func() int {
		root := &cobra.Command{
			Use:           "tt",
			SilenceErrors: true,
			SilenceUsage:  true,
		}
		root.AddCommand(newGroup())
		root.SetArgs(append([]string{"drawio"}, args...))

		err := root.Execute()
		if err == nil {
			return 0
		}
		errText = err.Error()
		if ec, ok := err.(interface{ ExitCode() int }); ok {
			return ec.ExitCode()
		}
		return 1
	})
	return code, out, errText
}
