package cli

// tzs_verb_test.go —— 具名动词那一层的测试。
//
// 这一层**不碰引擎**：只测纯函数（取开关、找动词、读 --args 原文、旧写法提示）与
// 分派里在"拉 manifest 之前"就返回的那几条路径。真 manifest 是
// tzs_verb_e2e_test.go 在管（TTZS_E2E=1）—— 这里刻意不去连它：
// 一条依赖本机配置的测试会时好时坏，那比不测更糟。

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tt/internal/dev/tzs"
)

// verbFixtureManifest 是给 lookupVerb / verbExample 用的小夹具。
//
// 字段名与参数**照引擎的真实形状**写（t=handle/enum/path[]/int，req，values，from），
// 因为 verbExample 是从这张表生成示例的 —— 夹具形状不对，测出来的就不是真实渲染。
const verbFixtureManifest = `[
  {"fn":"list_open","group":"会话","desc":"列出打开的句柄","writes":false,"slow":false,
   "returns":"list<el>","needsHandle":false,"args":[],"errors":[]},
  {"fn":"save","group":"会话","desc":"把句柄的模型写回新包","writes":false,"slow":false,
   "returns":"void","needsHandle":true,
   "args":[{"n":"handle","t":"handle","req":true},
           {"n":"out","t":"string","req":true,"desc":"输出 .tzs 路径"}],"errors":[]},
  {"fn":"nudge","group":"结构","desc":"按方向平移","writes":true,"slow":false,
   "returns":"delta","needsHandle":true,
   "args":[{"n":"handle","t":"handle","req":true},
           {"n":"paths","t":"path[]","req":true},
           {"n":"direction","t":"enum","req":true,"values":["up","down","left","right"]},
           {"n":"offset","t":"int","req":false,"desc":"格数，默认 1"}],"errors":[]},
  {"fn":"set_spec_attr","group":"属性","desc":"改字段规格属性","writes":true,"slow":false,
   "returns":"delta","needsHandle":true,
   "args":[{"n":"handle","t":"handle","req":true},
           {"n":"path","t":"path","req":true,"desc":"name-path"},
           {"n":"kind","t":"kind","req":true},
           {"n":"attr","t":"attr","req":true,"from":"spec:<kind>"},
           {"n":"value","t":"string","req":true,"desc":"新值"}],"errors":[]},
  {"fn":"form_tree","group":"读","desc":"结构树","writes":false,"slow":false,
   "returns":"tree","needsHandle":true,
   "args":[{"n":"handle","t":"handle","req":true},
           {"n":"depth","t":"int","req":false}],"errors":[]}
]`

func verbFixture(t *testing.T) *tzs.Manifest {
	t.Helper()
	m, err := tzs.ParseManifest([]byte(verbFixtureManifest))
	if err != nil {
		t.Fatalf("夹具 manifest 该能解析：%v", err)
	}
	return m
}

// TestTakeVerbFlags 钉住动词命令的解析：**只有**传输级开关，参数一律走 JSON。
func TestTakeVerbFlags(t *testing.T) {
	cases := []struct {
		name      string
		argv      []string
		verb      string
		extra     []string
		wantJSON  bool
		wantHelp  bool
		wantWS    string
		wantArgs  string
		wantArgsF string
		wantForm  string
		wantTO    time.Duration
		wantErr   string
	}{
		{name: "只有动词", argv: []string{"list_open"}, verb: "list_open"},
		{name: "动词 + json", argv: []string{"list_open", "--json"}, verb: "list_open", wantJSON: true},
		{name: "args", argv: []string{"nudge", "--args", `{"direction":"up"}`},
			verb: "nudge", wantArgs: `{"direction":"up"}`},
		{name: "args-file 指向标准输入", argv: []string{"nudge", "--args-file=-"},
			verb: "nudge", wantArgsF: "-"},
		{name: "form 按程序名寻址", argv: []string{"nudge", "--form", "aapp320"},
			verb: "nudge", wantForm: "aapp320"},
		{name: "form 用 = 写、给 ProgramKey", argv: []string{"nudge", "--form=aapp320|Form"},
			verb: "nudge", wantForm: "aapp320|Form"},
		{name: "workspace", argv: []string{"open", "--workspace", `D:\ws`},
			verb: "open", wantWS: `D:\ws`},
		{name: "rpc-timeout", argv: []string{"open", "--rpc-timeout", "300"},
			verb: "open", wantTO: 300 * time.Second},
		{name: "rpc-timeout 用 = 写", argv: []string{"open", "--rpc-timeout=200"},
			verb: "open", wantTO: 200 * time.Second},
		{name: "-h", argv: []string{"open", "-h"}, verb: "open", wantHelp: true},
		{name: "--help", argv: []string{"open", "--help"}, verb: "open", wantHelp: true},

		// ---- 参数一律走 JSON：flag 形式必须当场被拒 ----
		{name: "动词参数写成 flag 要报错", argv: []string{"nudge", "--handle", "h9"},
			wantErr: "未知开关 --handle"},
		{name: "json 里的 timeout 不是开关", argv: []string{"open", "--timeout", "30"},
			wantErr: "未知开关 --timeout"},
		// 裸词先**留着**：分派时可能拿它拼两级动词名（`field add` → `field_add`）；
		// 拼不上才由 runTzsVerb 报"多了位置参数"。
		{name: "动词后面的裸词先留着", argv: []string{"nudge", "h9"},
			verb: "nudge", extra: []string{"h9"}},
		{name: "两级动词名的第二个词也先留着", argv: []string{"field", "add", "--json"},
			verb: "field", extra: []string{"add"}, wantJSON: true},
		{name: "连字符动词名不算开关", argv: []string{"set-spec-attr", "--json"},
			verb: "set-spec-attr", wantJSON: true},

		// ---- 开关自己的错 ----
		{name: "rpc-timeout 不是数字", argv: []string{"open", "--rpc-timeout", "abc"}, wantErr: "--rpc-timeout"},
		{name: "rpc-timeout 是 0", argv: []string{"open", "--rpc-timeout", "0"}, wantErr: "--rpc-timeout"},
		{name: "args 缺值", argv: []string{"nudge", "--args"}, wantErr: "--args 需要参数"},
		{name: "args-file 缺值", argv: []string{"nudge", "--args-file"}, wantErr: "--args-file"},
		{name: "form 缺值", argv: []string{"nudge", "--form"}, wantErr: "--form 需要参数"},
		{name: "workspace 缺值", argv: []string{"open", "--workspace"}, wantErr: "--workspace"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, err := takeVerbFlags(c.argv)
			if c.wantErr != "" {
				if err == nil {
					t.Fatalf("该报错（含 %q）", c.wantErr)
				}
				if !strings.Contains(err.Error(), c.wantErr) {
					t.Errorf("文案里缺 %q：%v", c.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("不该报错：%v", err)
			}
			if f.verb != c.verb {
				t.Errorf("verb 得 %q，想 %q", f.verb, c.verb)
			}
			if strings.Join(f.extra, " ") != strings.Join(c.extra, " ") {
				t.Errorf("extra 得 %v，想 %v", f.extra, c.extra)
			}
			if f.asJSON != c.wantJSON {
				t.Errorf("asJSON 得 %v", f.asJSON)
			}
			if f.help != c.wantHelp {
				t.Errorf("help 得 %v", f.help)
			}
			if f.workspace != c.wantWS {
				t.Errorf("workspace 得 %q，想 %q", f.workspace, c.wantWS)
			}
			if f.args != c.wantArgs || f.argsSet != (c.wantArgs != "") {
				t.Errorf("args 得 %q（set=%v），想 %q", f.args, f.argsSet, c.wantArgs)
			}
			if f.argsFile != c.wantArgsF || f.argsFileSet != (c.wantArgsF != "") {
				t.Errorf("argsFile 得 %q（set=%v），想 %q", f.argsFile, f.argsFileSet, c.wantArgsF)
			}
			if f.form != c.wantForm {
				t.Errorf("form 得 %q，想 %q", f.form, c.wantForm)
			}
			if f.timeout != c.wantTO {
				t.Errorf("timeout 得 %v，想 %v", f.timeout, c.wantTO)
			}
		})
	}
}

// TestSplitTwoWordVerb 钉住两级动词名：纯拼接、无映射表、拼不上就不当回事。
func TestSplitTwoWordVerb(t *testing.T) {
	m := verbFixture(t)
	if f, used := splitTwoWordVerb(m, "form", []string{"tree", "extra"}); f == nil || f.Name != "form_tree" || used != 1 {
		t.Errorf("form tree 该拼成 form_tree：%+v used=%d", f, used)
	}
	if f, _ := splitTwoWordVerb(m, "set", []string{"spec", "attr"}); f != nil {
		t.Errorf("只拼两个词，set spec 不该命中：%+v", f)
	}
	if f, used := splitTwoWordVerb(m, "form", nil); f != nil || used != 0 {
		t.Errorf("没有第二个词就不该拼：%+v", f)
	}
	if f, _ := splitTwoWordVerb(m, "field", []string{"add"}); f != nil {
		t.Errorf("引擎里没有 field_add 时不该凭空造出来：%+v", f)
	}
}

// TestLookupVerb 钉住动词查找：精确优先，连字符是别名，未知返回 nil。
func TestLookupVerb(t *testing.T) {
	m := verbFixture(t)
	if f := lookupVerb(m, "form_tree"); f == nil || f.Name != "form_tree" {
		t.Errorf("精确匹配失败：%+v", f)
	}
	if f := lookupVerb(m, "form-tree"); f == nil || f.Name != "form_tree" {
		t.Errorf("连字符别名该能找到 form_tree：%+v", f)
	}
	if f := lookupVerb(m, "set-spec-attr"); f == nil || f.Name != "set_spec_attr" {
		t.Errorf("连字符别名该能找到 set_spec_attr：%+v", f)
	}
	if f := lookupVerb(m, "form"); f != nil {
		t.Errorf("半个名字不该命中：%+v", f)
	}
	if f := lookupVerb(m, "nope"); f != nil {
		t.Errorf("未知动词该返回 nil：%+v", f)
	}
}

// TestReadArgsBody 读 JSON 原文：--args 直接用、--args-file 读文件、都没给是空对象、
// 读不到是 IO 失败（退 5）。
func TestReadArgsBody(t *testing.T) {
	got, code := readArgsBody(verbFlags{args: `{"a":1}`, argsSet: true})
	if code != 0 || string(got) != `{"a":1}` {
		t.Errorf("--args 得 %q（code %d）", got, code)
	}

	// 一个参数都不给的动词（list_open 之类）不必写 --args '{}'。
	got, code = readArgsBody(verbFlags{})
	if code != 0 || string(got) != `{}` {
		t.Errorf("没给 --args 时该是空对象，得 %q（code %d）", got, code)
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "args.json")
	if err := os.WriteFile(p, []byte(`{"b":2}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, code = readArgsBody(verbFlags{argsFile: p, argsFileSet: true})
	if code != 0 || string(got) != `{"b":2}` {
		t.Errorf("--args-file 得 %q（code %d）", got, code)
	}

	code = silent(t, func() int {
		_, c := readArgsBody(verbFlags{argsFile: filepath.Join(dir, "nope.json"), argsFileSet: true})
		return c
	})
	if code != 5 {
		t.Errorf("读不到文件该退 5，得 %d", code)
	}
}

// TestOldCallExample：旧写法给一条能直接粘的例子。
func TestOldCallExample(t *testing.T) {
	if got := oldCallExample(nil); !strings.Contains(got, "--args") {
		t.Errorf("空参数时该给一条现成的例子（带 --args）：%q", got)
	}
	if got := oldCallExample([]string{"open", "--path", "x"}); got != "open --path x" {
		t.Errorf("该原样搬运函数名与参数：%q", got)
	}
}

// TestRunTzsVerbLocalFailures：在"拉 manifest 之前"就该返回的那几条路径。
//
// 它们必须是确定的（不依赖本机有没有配工作区/引擎），否则 CLI 的用法错误会
// 随着环境变成 5 —— 而"参数打错"永远是 2。
func TestRunTzsVerbLocalFailures(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"没有动词", nil, 2},
		{"只有 flag", []string{"--json"}, 2},
		{"args 缺值", []string{"open", "--args"}, 2},
		{"动词参数写成 flag", []string{"nudge", "--handle", "h9"}, 2},
		{"旧写法 call + 函数", []string{"call", "open", "--path", "x"}, 2},
		{"旧写法 call 单独出现", []string{"call"}, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := silent(t, func() int { return runTzsVerb(c.args) }); got != c.want {
				t.Errorf("该退 %d，得 %d", c.want, got)
			}
		})
	}
}

// TestUsageTextHasNoCallGateway 是文档漂移的机械防线。
//
// `call <fn>` 已经删了；一段还写着它的帮助文本会让调用方敲一条注定退 2 的命令，
// 而这类漂移不会让任何别的测试失败。
func TestUsageTextHasNoCallGateway(t *testing.T) {
	for _, text := range []string{tzsUsage, Usage} {
		if strings.Contains(text, "call <fn>") || strings.Contains(text, "tzs call") {
			t.Errorf("帮助文本里还在教 call 写法：\n%s", text)
		}
		if !strings.Contains(text, "<动词>") {
			t.Errorf("帮助文本里该出现 <动词> 的用法：\n%s", text)
		}
		if !strings.Contains(text, "--args") {
			t.Errorf("帮助文本里该出现 --args（参数只用 JSON 给）：\n%s", text)
		}
		if !strings.Contains(text, "--form") {
			t.Errorf("帮助文本里该出现 --form（按程序名寻址，免得搬运句柄）：\n%s", text)
		}
	}
}

// TestVerbExample 钉住"示例优先"的那条示例：从 manifest 生成，能直接粘。
//
// 三条要点：① 需要句柄的动词用 `--form <程序名>`（而不是 handle）；② 只列**必填**参数
// （例子要短、要一定能跑通）；③ `--args` 那一段必须是**合法 JSON** —— 用坏例子教人是最坏的。
func TestVerbExample(t *testing.T) {
	m := verbFixture(t)

	save := verbExample(m.ByName("save")) // handle + out（都必填）
	if !strings.Contains(save, "--form aapp320") {
		t.Errorf("需要句柄的动词该用 --form：%s", save)
	}
	if strings.Contains(save, "handle") {
		t.Errorf("示例里不该出现 handle（那是内部 id）：%s", save)
	}
	for _, want := range []string{"tt dev tzs save", `"out":"<值>"`, "--json"} {
		if !strings.Contains(save, want) {
			t.Errorf("示例里缺 %q：%s", want, save)
		}
	}
	assertExampleArgsIsJSON(t, save)

	// list_open 既不需要句柄、也没有参数：示例该短到只有命令本身。
	lo := verbExample(m.ByName("list_open"))
	if strings.Contains(lo, "--form") || strings.Contains(lo, "--args") {
		t.Errorf("无参动词的示例不该带 --form/--args：%s", lo)
	}

	// nudge 的 offset 是选填：不该出现在示例里（参数表里有）；必填的枚举要给合法值之一。
	nudge := verbExample(m.ByName("nudge"))
	if strings.Contains(nudge, "offset") {
		t.Errorf("选填参数不该进示例：%s", nudge)
	}
	if !strings.Contains(nudge, `"direction":"up"`) {
		t.Errorf("枚举占位该用它的第一个合法值：%s", nudge)
	}
	assertExampleArgsIsJSON(t, nudge)

	// path / kind / attr / string 各自给一个说明性的占位符（不是编造一个具体值）。
	ssa := verbExample(m.ByName("set_spec_attr"))
	for _, want := range []string{`"path":"<name-path>"`, `"kind":"<kind>"`, `"attr":"<属性名>"`, `"value":"<值>"`} {
		if !strings.Contains(ssa, want) {
			t.Errorf("示例里缺 %q：%s", want, ssa)
		}
	}
	assertExampleArgsIsJSON(t, ssa)
}

// TestPlaceholder 钉住每种参数类型的占位符（示例的正确性靠它）。
func TestPlaceholder(t *testing.T) {
	cases := []struct {
		p    tzs.Param
		want string
	}{
		{tzs.Param{Name: "n", Type: tzs.TypeInt}, "0"},
		{tzs.Param{Name: "b", Type: tzs.TypeBool}, "true"},
		{tzs.Param{Name: "e", Type: tzs.TypeEnum, Values: []string{"up", "down"}}, `"up"`},
		{tzs.Param{Name: "e2", Type: tzs.TypeEnum}, `"<值>"`},
		{tzs.Param{Name: "l", Type: tzs.TypePathList}, "[]"},
		{tzs.Param{Name: "l2", Type: tzs.TypeStrList}, "[]"},
		{tzs.Param{Name: "p", Type: tzs.TypePath}, `"<name-path>"`},
		{tzs.Param{Name: "k", Type: tzs.TypeKind}, `"<kind>"`},
		{tzs.Param{Name: "a", Type: tzs.TypeAttr}, `"<属性名>"`},
		{tzs.Param{Name: "s", Type: tzs.TypeString}, `"<值>"`},
		{tzs.Param{Name: "h", Type: tzs.TypeHandle}, `"<值>"`},
	}
	for _, c := range cases {
		if got := placeholder(&c.p); got != c.want {
			t.Errorf("placeholder(%s/%s) = %s，想 %s", c.p.Name, c.p.Type, got, c.want)
		}
	}
}

// assertExampleArgsIsJSON 取出示例里 `--args '...'` 的那一段，断言它是合法 JSON 对象。
func assertExampleArgsIsJSON(t *testing.T, example string) {
	t.Helper()
	i := strings.Index(example, "--args '")
	if i < 0 {
		return // 没有 --args（无参动词）
	}
	rest := example[i+len("--args '"):]
	j := strings.Index(rest, "'")
	if j < 0 {
		t.Fatalf("示例里的单引号没配对：%s", example)
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(rest[:j]), &obj); err != nil {
		t.Errorf("示例里的 args 不是合法 JSON：%v\n  %s", err, rest[:j])
	}
}

// TestProgressAfter 钉住慢调用那一行：到点才打，收尾之后绝不再打。
func TestProgressAfter(t *testing.T) {
	var b strings.Builder
	stop := progressAfter(&b, 10*time.Millisecond, "正在执行 validate")
	time.Sleep(60 * time.Millisecond)
	stop()
	if !strings.Contains(b.String(), "正在执行 validate") {
		t.Errorf("超过阈值该打一行：%q", b.String())
	}

	var b2 strings.Builder
	stop2 := progressAfter(&b2, time.Hour, "正在执行 list_open")
	stop2() // 立刻收尾 = 快命令
	time.Sleep(10 * time.Millisecond)
	if b2.String() != "" {
		t.Errorf("快命令不该有任何输出：%q", b2.String())
	}
}

// TestFilterHintAndNudge 钉住"先过滤"的两处提示，重点是**什么时候不提**。
//
// 判据全部来自 manifest：① 会改模型的动词不谈收窄（field_add 的 table 是输入，不是过滤器，
// 不挡就会印出一句毫无意义的"先收窄：table…"）；② 只有真会回一大块的读动词才提；
// ③ 已经过滤过就不再念。
func TestFilterHintAndNudge(t *testing.T) {
	// 有 query 的读动词 → 提。
	read := &tzs.SpecFn{Name: "list_columns", Writes: false, Returns: "list<el>",
		Args: []*tzs.Param{{Name: "table", Type: tzs.TypeString, Required: true},
			{Name: "query", Type: tzs.TypeString, Required: false}}}
	if h := filterHint(read); !strings.Contains(h, "query") {
		t.Errorf("有 query 的读动词该提示先过滤：%q", h)
	}

	// 会改模型的动词：即使有 table 这种"收窄名"，也不提。
	write := &tzs.SpecFn{Name: "field_add", Writes: true, Returns: "report",
		Args: []*tzs.Param{{Name: "table", Type: tzs.TypeString, Required: true}}}
	if h := filterHint(write); h != "" {
		t.Errorf("改模型的动词不该有收窄提示：%q", h)
	}

	big := json.RawMessage(`{"count":109,"columns":[]}`)
	if got := nudge(read, `{"table":"pmdl_t"}`, big); !strings.Contains(got, "query") {
		t.Errorf("109 条且没过滤，该提醒：%q", got)
	}
	if got := nudge(read, `{"table":"pmdl_t","query":"pmdl00"}`, big); got != "" {
		t.Errorf("已经过滤过就不该念：%q", got)
	}
	if got := nudge(read, `{"table":"pmdl_t"}`, json.RawMessage(`{"count":2}`)); got != "" {
		t.Errorf("小返回不该提醒：%q", got)
	}
	if got := nudge(write, `{"table":"pmdl_t"}`, big); got != "" {
		t.Errorf("改模型的动词不该提醒：%q", got)
	}
}

// nudge 跑一次 maybeNudgeFilter，把要打印的东西收回来。
func nudge(fn *tzs.SpecFn, args string, result json.RawMessage) string {
	var b strings.Builder
	maybeNudgeFilter(&b, fn, json.RawMessage(args), result)
	return b.String()
}

// TestArgGiven 判据：空串 / null / 空数组都不算"收窄过"。
func TestArgGiven(t *testing.T) {
	cases := []struct {
		args string
		name string
		want bool
	}{
		{`{"query":"pmdl"}`, "query", true},
		{`{"query":""}`, "query", false},
		{`{"query":null}`, "query", false},
		{`{"query":[]}`, "query", false},
		{`{}`, "query", false},
		{`{"limit":0}`, "limit", true}, // 给了就是给了；0 合不合法由引擎说
	}
	for _, c := range cases {
		if got := argGiven(json.RawMessage(c.args), c.name); got != c.want {
			t.Errorf("argGiven(%s, %s) = %v，想 %v", c.args, c.name, got, c.want)
		}
	}
}

// TestResultCount 认引擎的两种计数键（清单类返回的约定）。
func TestResultCount(t *testing.T) {
	if got := resultCount(json.RawMessage(`{"count":109}`)); got != 109 {
		t.Errorf("count 该读出来：%d", got)
	}
	if got := resultCount(json.RawMessage(`{"returned":7,"totalRows":9}`)); got != 7 {
		t.Errorf("returned 该读出来：%d", got)
	}
	if got := resultCount(json.RawMessage(`{"nodes":[]}`)); got != 0 {
		t.Errorf("没有计数键就是 0：%d", got)
	}
}

// TestRemovedCommandsAreNotAdvertised：fns / manifest 两个命令已删除
// （对外只留"动词"这一层，函数表不再整个摊出去）。
//
// 两半都要钉：① 帮助文本里不再出现它们 —— 否则调用方敲一条注定退 2 的命令；
// ② 它们不再是内建动词 —— 否则引擎那边万一有同名函数，会被内建抢走而不可达。
func TestRemovedCommandsAreNotAdvertised(t *testing.T) {
	for _, text := range []string{tzsUsage, Usage} {
		for _, gone := range []string{"tzs fns", "tzs manifest"} {
			if strings.Contains(text, gone) {
				t.Errorf("帮助文本里还在教 %q：\n%s", gone, text)
			}
		}
	}
	for _, b := range builtinVerbs {
		if b == "fns" || b == "manifest" {
			t.Errorf("%q 不该再是内建动词", b)
		}
	}
	for _, a := range []string{"fns", "manifest"} {
		if got := silent(t, func() int { return cmdTzs([]string{a}) }); got != 2 {
			t.Errorf("tt dev tzs %s 该退 2（已不是命令），得 %d", a, got)
		}
	}
}

// TestCmdTzsHelpIsStatic：`tt dev tzs --help` 不依赖引擎（拿不到 manifest 也退 0）。
//
// 这条是刻意的：`--help` 是"我还没有环境时唯一能用的一条命令"，
// 让它去 Boot 一个设计器（或在没配工作区时退 5）等于把说明书锁在门里面。
// 动词索引是**尽力而为**的附加物：拿不到就少一段，不影响退出码。
func TestCmdTzsHelpIsStatic(t *testing.T) {
	for _, a := range []string{"-h", "--help", "help"} {
		code, out := captureStdout(t, func() int { return cmdTzs([]string{a}) })
		if code != 0 {
			t.Errorf("tt dev tzs %s 该退 0，得 %d", a, code)
		}
		if !strings.Contains(out, "tt dev tzs <动词>") {
			t.Errorf("tt dev tzs %s 的输出里该有动词用法：\n%s", a, out)
		}
	}
	if got := silent(t, func() int { return cmdTzs(nil) }); got != 2 {
		t.Errorf("没有子命令该退 2，得 %d", got)
	}
}
