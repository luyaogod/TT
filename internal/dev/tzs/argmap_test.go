package tzs

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

//---------------------------------------------------------------------------
// SplitArgs：纯语法
//---------------------------------------------------------------------------

// TestSplitArgs 钉住 `--k v` 的拆分规则。
//
// 最关键的一条是「负号不是 flag」：`-1` 不以 `--` 开头，所以它是**值**。
// 少了这一条，`--value -1` 会把 -1 拆成一个独立的位置参数（或未知参数），
// 于是一个合法的负值永远传不进去 —— 而报出来的错是「未知参数 -1」，
// 看起来像用户打错了。
func TestSplitArgs(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []Arg
	}{
		{"k v", []string{"--value", "-1"}, []Arg{{"value", "-1"}}},
		{"k=v", []string{"--value=-1"}, []Arg{{"value", "-1"}}},
		{"空值", []string{"--desc="}, []Arg{{"desc", ""}}},
		{"k=v 里有 =", []string{"--desc=a=b"}, []Arg{{"desc", "a=b"}}},
		{"裸布尔", []string{"--force"}, []Arg{{"force", "true"}}},
		{"裸布尔后面跟 flag", []string{"--force", "--json"}, []Arg{{"force", "true"}, {"json", "true"}}},
		{"显式 false", []string{"--force", "false"}, []Arg{{"force", "false"}}},
		{"可重复", []string{"--paths", "a", "--paths", "b"}, []Arg{{"paths", "a"}, {"paths", "b"}}},
		{"以 -- 开头的值要用 =", []string{"--path=--weird"}, []Arg{{"path", "--weird"}}},
		{"空列表", nil, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SplitArgs(c.in)
			if err != nil {
				t.Fatalf("不该报错：%v", err)
			}
			if len(got) != len(c.want) {
				t.Fatalf("拆出 %d 个，想要 %d 个：%+v", len(got), len(c.want), got)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("第 %d 个：得 %+v，想要 %+v", i, got[i], c.want[i])
				}
			}
		})
	}
}

// TestSplitArgsRejects 位置参数与 `--` 必须**当场报错**（退出 2）。
//
// 这是与引擎自带 tzs-cli 的一处有意差异：那边把不带 `--` 的词「忽略并继续」。
// 无声忽略正是引擎 Manifest.Check 双向卡 arity 要治的病（`--attrr` 打错了和一次
// 合法调用在结果上无法区分，直到有人去读文件）。
func TestSplitArgsRejects(t *testing.T) {
	for _, in := range [][]string{{"foo"}, {"--paths", "a", "foo"}, {"--"}, {"--paths", "--"}} {
		_, err := SplitArgs(in)
		if err == nil {
			t.Errorf("输入 %v 该被拒", in)
			continue
		}
		var ue *UsageError
		if !errors.As(err, &ue) {
			t.Errorf("输入 %v 该是用法错，得 %T", in, err)
		}
	}
}

//---------------------------------------------------------------------------
// BuildArgs：定型 + 本地校验
//---------------------------------------------------------------------------

// TestBuildArgsTable 是 argmap 的主表。
//
// 表里既有「必须本地拦住」的（枚举越界、类型不对、缺必填、未知参数、标量重复），
// 也有「**绝不能**本地拦住」的三条反向断言（kind / attr / string）——
// 后三条更重要：越界拦截是**无声**发生的，写错了不会有任何测试失败，
// 只会让某些合法调用永远到不了引擎。
func TestBuildArgsTable(t *testing.T) {
	cases := []struct {
		name    string
		fn      string
		argv    []string
		want    string // 成功时期望的 args JSON（空 = 应当失败）
		wantErr string // 失败时文案里必须出现的子串
	}{
		// ---- int / bool / 数组的定型 ----
		{
			"int 变数字", "nudge",
			[]string{"--handle", "h1", "--paths", "a", "--direction", "up", "--offset", "2"},
			`{"handle":"h1","paths":["a"],"direction":"up","offset":2}`, "",
		},
		{
			"选填省略", "nudge",
			[]string{"--handle", "h1", "--paths", "a", "--direction", "up"},
			`{"handle":"h1","paths":["a"],"direction":"up"}`, "",
		},
		{
			"逗号拆数组", "nudge",
			[]string{"--handle", "h1", "--paths", "a,b", "--direction", "down"},
			`{"handle":"h1","paths":["a","b"],"direction":"down"}`, "",
		},
		{
			"数组项去空白", "nudge",
			[]string{"--handle", "h1", "--paths", " a , b ", "--direction", "up"},
			`{"handle":"h1","paths":["a","b"],"direction":"up"}`, "",
		},
		{
			"可重复的数组", "nudge",
			[]string{"--handle", "h1", "--paths", "a", "--paths", "b", "--direction", "up"},
			`{"handle":"h1","paths":["a","b"],"direction":"up"}`, "",
		},
		{
			"= 写法 + 负整数", "nudge",
			[]string{"--handle=h1", "--paths=a", "--direction=up", "--offset=-3"},
			`{"handle":"h1","paths":["a"],"direction":"up","offset":-3}`, "",
		},
		{
			"裸布尔", "set_excluded",
			[]string{"--handle", "h1", "--path", "p", "--excluded"},
			`{"handle":"h1","path":"p","excluded":true}`, "",
		},
		{
			"布尔 false", "set_excluded",
			[]string{"--handle", "h1", "--path", "p", "--excluded", "false"},
			`{"handle":"h1","path":"p","excluded":false}`, "",
		},
		{
			"布尔 0", "set_excluded",
			[]string{"--handle", "h1", "--path", "p", "--excluded", "0"},
			`{"handle":"h1","path":"p","excluded":false}`, "",
		},
		{
			"布尔大写", "set_excluded",
			[]string{"--handle", "h1", "--path", "p", "--excluded", "YES"},
			`{"handle":"h1","path":"p","excluded":true}`, "",
		},
		{
			"无参函数", "list_open", nil,
			`{}`, "",
		},
		{
			"path 里的反斜杠", "save",
			[]string{"--handle", "h1", "--out", `D:\x\y.tzs`},
			`{"handle":"h1","out":"D:\\x\\y.tzs"}`, "",
		},
		{
			"数组给空串", "nudge",
			[]string{"--handle", "h1", "--paths", "", "--direction", "up"},
			`{"handle":"h1","paths":[],"direction":"up"}`, "",
		},

		// ---- 反向断言：本地**绝不**拦（越界拦截是无声发生的） ----
		{
			// from:"spec:<kind>" 的合法集要现场从模型里取，静态列不出来。
			// 本地拦它 = 让描述器白名单永远到不了引擎。
			"attr(from:*) 放行", "set_spec_attr",
			[]string{"--handle", "h1", "--path", "p", "--kind", "field", "--attr", "bogus_attr", "--value", "v"},
			`{"handle":"h1","path":"p","kind":"field","attr":"bogus_attr","value":"v"}`, "",
		},
		{
			// kind 的合法集（七种）**不在 manifest 里** —— Go 侧抄一份就是第二份真源。
			// 引擎会回 E_BAD_PARAM（kind=validation，退 2），和本地拦下来一模一样。
			"kind 放行", "set_spec_attr",
			[]string{"--handle", "h1", "--path", "p", "--kind", "alien", "--attr", "a", "--value", "v"},
			`{"handle":"h1","path":"p","kind":"alien","attr":"a","value":"v"}`, "",
		},
		{
			// add_action 的「型态」是一张**每表单自己的**词表（该表单的 s_detail<n>），
			// 在 manifest 里它就是一个普通 string。在本地拦它会让「加一个 db* 型态」
			// 永远到不了引擎。
			"per-form 词表放行", "add_action",
			[]string{"--handle", "h1", "--path", "p", "--type", "Bogus"},
			`{"handle":"h1","path":"p","type":"Bogus"}`, "",
		},
		{
			// string 参数不按逗号切（只有 path[]/string[] 才切）。
			"string 里的逗号不切", "add_action",
			[]string{"--handle", "h1", "--path", "p", "--type", "di3,db4"},
			`{"handle":"h1","path":"p","type":"di3,db4"}`, "",
		},
		{
			// 契约点名的：负号不是 flag，`--value -1` 得到字符串 "-1"。
			"value -1 是字符串", "set_spec_attr",
			[]string{"--handle", "h1", "--path", "p", "--kind", "field", "--attr", "a", "--value", "-1"},
			`{"handle":"h1","path":"p","kind":"field","attr":"a","value":"-1"}`, "",
		},
		{
			"= 给空串", "set_spec_attr",
			[]string{"--handle", "h1", "--path", "p", "--kind", "field", "--attr", "a", "--value="},
			`{"handle":"h1","path":"p","kind":"field","attr":"a","value":""}`, "",
		},

		// ---- string[]：切逗号但保 | ----
		{
			"string[] 切逗号", "set_items",
			[]string{"--handle", "h1", "--path", "p", "--items", "lbl_a|A|d1,lbl_b|B|d2"},
			`{"handle":"h1","path":"p","items":["lbl_a|A|d1","lbl_b|B|d2"]}`, "",
		},
		{
			"string[] 可重复", "set_items",
			[]string{"--handle", "h1", "--path", "p", "--items", "x", "--items", "y"},
			`{"handle":"h1","path":"p","items":["x","y"]}`, "",
		},

		// ---- 本地必须拦 ----
		{
			"枚举越界", "nudge",
			[]string{"--handle", "h1", "--paths", "a", "--direction", "Bogus"},
			"", "direction",
		},
		{
			"枚举越界要列合法值", "nudge",
			[]string{"--handle", "h1", "--paths", "a", "--direction", "Bogus"},
			"", "up|down|left|right",
		},
		{
			"缺必填", "nudge",
			[]string{"--handle", "h1", "--paths", "a"},
			"", "缺必填参数 direction",
		},
		{
			"缺必填 handle", "nudge",
			[]string{"--paths", "a", "--direction", "up"},
			"", "缺必填参数 handle",
		},
		{
			"未知参数", "nudge",
			[]string{"--handle", "h1", "--paths", "a", "--direction", "up", "--bogus", "1"},
			"", "bogus",
		},
		{
			"整数解析失败", "nudge",
			[]string{"--handle", "h1", "--paths", "a", "--direction", "up", "--offset", "2x"},
			"", "offset",
		},
		{
			"布尔值不认识", "set_excluded",
			[]string{"--handle", "h1", "--path", "p", "--excluded", "maybe"},
			"", "excluded",
		},
		{
			"标量给两次", "nudge",
			[]string{"--handle", "h1", "--paths", "a", "--direction", "up", "--offset", "2", "--offset", "3"},
			"", "给了一次以上",
		},
		{
			"无参函数收到参数", "list_open",
			[]string{"--handle", "h1"},
			"", "handle",
		},
		{
			"stop 不在 manifest 里", "stop",
			[]string{},
			"", "传输级命令",
		},
		{
			"未知函数", "nope",
			[]string{},
			"", "未知函数 nope",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := mustManifest(t)
			got, err := BuildArgs(m, c.fn, argsOf(c.argv))
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("不该报错：%v", err)
				}
				if string(got) != c.want {
					t.Fatalf("定型结果不对\n  得 %s\n想 %s", got, c.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("该报错（含 %q），却得到 %s", c.wantErr, got)
			}
			var ue *UsageError
			if !errors.As(err, &ue) {
				t.Fatalf("该是用法错（退 2），得 %T：%v", err, err)
			}
			if ue.ExitCode() != ExitUsage {
				t.Errorf("该退 2，得 %d", ue.ExitCode())
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("文案里缺 %q：%v", c.wantErr, err)
			}
		})
	}
}

// argsOf 把 argv 过一遍 SplitArgs（测试里就写命令行原样，不手搓 Arg）。
func argsOf(argv []string) []Arg {
	a, err := SplitArgs(argv)
	if err != nil {
		panic(err)
	}
	return a
}

// TestBuildArgsReverseRequestCarried 是反向断言的**第二半**：
// 本地放行不算数，还得真的进了请求帧。
//
// 只说「本地没拦」是不够的 —— 一个把值悄悄丢掉的实现也能过那一半。
func TestBuildArgsReverseRequestCarried(t *testing.T) {
	m := mustManifest(t)
	cases := []struct {
		fn    string
		argv  []string
		inReq string // 请求帧里必须原样出现的值
	}{
		{"set_spec_attr",
			[]string{"--handle", "h1", "--path", "p", "--kind", "field", "--attr", "bogus_attr", "--value", "v"},
			`"attr":"bogus_attr"`},
		{"set_spec_attr",
			[]string{"--handle", "h1", "--path", "p", "--kind", "alien", "--attr", "a", "--value", "v"},
			`"kind":"alien"`},
		{"add_action",
			[]string{"--handle", "h1", "--path", "p", "--type", "Bogus"},
			`"type":"Bogus"`},
		{"nudge",
			[]string{"--handle", "h1", "--paths", "a", "--direction", "up", "--offset", "-7"},
			`"offset":-7`},
	}
	for _, c := range cases {
		args, err := BuildArgs(m, c.fn, argsOf(c.argv))
		if err != nil {
			t.Fatalf("%s：本地不该拦：%v", c.fn, err)
		}
		frame, err := MarshalRequest(1, c.fn, args)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(frame), c.inReq) {
			t.Errorf("%s：请求帧里缺 %s\n帧：%s", c.fn, c.inReq, frame)
		}
	}
}

// TestBuildArgsOrderIsManifestOrder：输出顺序固定 = manifest 声明顺序。
//
// 固定顺序是为了可复现：同一份输入永远产出同一串字节，日志、--json、测试断言
// 才能逐字节比对（命令行给参数的顺序不该影响帧）。
func TestBuildArgsOrderIsManifestOrder(t *testing.T) {
	m := mustManifest(t)
	a, err := BuildArgs(m, "nudge", argsOf([]string{
		"--offset", "1", "--direction", "up", "--paths", "a", "--handle", "h1",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"handle":"h1","paths":["a"],"direction":"up","offset":1}`
	if string(a) != want {
		t.Errorf("参数顺序没按 manifest：\n  得 %s\n想 %s", a, want)
	}
	if !json.Valid(a) {
		t.Errorf("产出的不是合法 JSON：%s", a)
	}
}
