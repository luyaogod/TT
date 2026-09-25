package cli

// tzs_verb_e2e_test.go —— 拿**真 manifest** 验动词面。
//
// 为什么放在 cli 包而不是 internal/dev/tzs 的 E2E 里：要证的几件事都是**命令层**的性质，
// 而且它们只需要 `tzs-server --manifest`（不 Boot、不连守护进程、不需要工作区），
// 所以这条 E2E 很便宜。
//
// 开启方式：
//	$env:TTZS_E2E="1"
//	$env:TTZS_EXE="D:\...\engine\out\tzs-server.exe"   # 可省：默认取仓库里的 engine/out
//	go test ./internal/dev/cli/ -run TestE2EVerbSurface -v

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tt/internal/dev/tzs"
)

func requireRealManifest(t *testing.T) *tzs.Manifest {
	t.Helper()
	if os.Getenv("TTZS_E2E") == "" {
		t.Skip("需要真引擎：设 TTZS_E2E=1（可选 TTZS_EXE），见本文件顶部注释")
	}
	exe := os.Getenv("TTZS_EXE")
	if exe == "" {
		exe = filepath.Join("..", "..", "..", "engine", "out", "tzs-server.exe")
	}
	if _, err := os.Stat(exe); err != nil {
		t.Skipf("找不到引擎 exe（TTZS_EXE=%s）：%v", exe, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	m, err := tzs.FetchManifest(ctx, exe)
	if err != nil {
		t.Fatalf("拉不到真 manifest：%v", err)
	}
	return m
}

// TestE2EVerbSurface 拿真 manifest 证命令层的几条性质。
//
// 它们的共同后果都是"某个动词/参数悄悄不可达"，而现象看起来像引擎的问题：
//
//	① 内建动词与引擎函数撞名 → 内建赢，那个函数再也敲不出来；
//	② 有函数名却敲不出来（连字符别名也要能敲）；
//	③ 有参数却定型不了（类型没被认识 → 合法调用到不了引擎）；
//	④ 参数还能用 flag 形式给 —— 那就说明"只用 JSON"没有被真正落实。
func TestE2EVerbSurface(t *testing.T) {
	m := requireRealManifest(t)
	t.Logf("真 manifest：%d 个动词，%d 个参数", len(m.Fns), countParams(m))

	// ① 内建动词不能与引擎函数同名（同名时 warnBuiltinCollisions 会警告，这里直接判失败：
	//    真发生了就该改设计，而不是靠一句警告糊过去）。
	for _, b := range builtinVerbs {
		if m.ByName(b) != nil {
			t.Errorf("引擎有函数叫 %q，与内建动词撞名：内建优先，该函数无法用动词调用", b)
		}
	}

	switches := transportSwitchNames()
	for _, f := range m.Fns {
		// ② 本名与连字符形式都要找得到。
		if lookupVerb(m, f.Name) == nil {
			t.Errorf("动词 %q 用本名找不到", f.Name)
		}
		if kebab := strings.ReplaceAll(f.Name, "_", "-"); kebab != f.Name {
			if lookupVerb(m, kebab) == nil {
				t.Errorf("动词 %q 的连字符形式 %q 找不到", f.Name, kebab)
			}
		}

		// ③ 每个声明的参数都要能被 JSON 入口定型（给齐必填就是一次合法调用）。
		body := dummyArgsJSON(t, f)
		args, err := tzs.BuildArgsFromJSON(m, f.Name, body)
		if err != nil {
			t.Errorf("%s：给齐参数后 JSON 定型仍失败（%s）：%v", f.Name, body, err)
		} else if !json.Valid(args) {
			t.Errorf("%s：定型结果不是合法 JSON：%s", f.Name, args)
		}

		// ④ 同一个参数用 flag 形式给，必须当场被拒（参数只用 JSON）。
		for _, p := range f.Args {
			if switches[p.Name] {
				continue // 名字恰好是传输开关：那不是参数形式的错
			}
			if _, err := takeVerbFlags([]string{f.Name, "--" + p.Name, "v"}); err == nil {
				t.Errorf("%s --%s：参数只用 JSON 给，flag 形式该被拒", f.Name, p.Name)
			}
		}

		// ⑤ `--help` 里那条示例：`--args` 那一段必须真的是合法 JSON ——
		//    用坏例子教人是最坏的一种错，而所有示例都是生成的，一次全检很便宜。
		ex := verbExample(f)
		if f.NeedsHandle && !strings.Contains(ex, "--form") {
			t.Errorf("%s 的示例该用 --form 寻址：%s", f.Name, ex)
		}
		if !f.NeedsHandle && strings.Contains(ex, "--form") {
			t.Errorf("%s 不需要句柄，示例里不该有 --form：%s", f.Name, ex)
		}
		if i := strings.Index(ex, "--args '"); i >= 0 {
			rest := ex[i+len("--args '"):]
			if j := strings.Index(rest, "'"); j < 0 {
				t.Errorf("%s 的示例单引号没配对：%s", f.Name, ex)
			} else if !json.Valid([]byte(rest[:j])) {
				t.Errorf("%s 的示例里 args 不是合法 JSON：%s", f.Name, rest[:j])
			}
		}

		// ⑥ 示例里的占位符必须与参数的**角色**一致。包路径教成 `<name-path>` 是这条测试
		//    存在的理由：`open.path` 与 `field_add.out` 从前都被教错，而它们都是包路径
		//    （2026-09-24 的设计评审发现，修法是引擎声明里的 `role`）。
		if args := exampleArgsOf(t, ex); args != nil {
			for _, p := range f.Args {
				if p.Role != tzs.RolePackagePath {
					continue
				}
				if v, ok := args[p.Name].(string); ok && v == "<name-path>" {
					t.Errorf("%s：参数 %s 的角色是包路径，示例却教成 <name-path>：%s", f.Name, p.Name, ex)
				}
			}
		}

		// 动词的参数表要能渲染（`<动词> --help` 走的就是 printFn）。
		if f.Name == "list_open" {
			printFn(io.Discard, f)
		}
	}
}

// transportSwitchNames 把传输级开关的名字（不带 --）取成集合。
func transportSwitchNames() map[string]bool {
	out := map[string]bool{}
	for _, s := range verbTransportFlags {
		out[strings.TrimLeft(s, "-")] = true
	}
	return out
}

// dummyArgsJSON 给一个动词的**全部**参数各造一个类型正确的值。
//
// 给全部（不只必填）：这样"某个可选参数的类型没被认识"也会被抓到 ——
// 而那种错恰恰只在有人真的用它时才暴露。
func dummyArgsJSON(t *testing.T, f *tzs.SpecFn) json.RawMessage {
	t.Helper()
	obj := map[string]any{}
	for _, p := range f.Args {
		obj[p.Name] = dummyValue(p)
	}
	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("%s：造参数失败：%v", f.Name, err)
	}
	return b
}

func dummyValue(p *tzs.Param) any {
	switch p.Type {
	case tzs.TypeInt:
		return 1
	case tzs.TypeBool:
		return true
	case tzs.TypeEnum:
		if len(p.Values) > 0 {
			return p.Values[0]
		}
		return "x"
	case tzs.TypePathList, tzs.TypeStrList:
		return []string{"x"}
	case tzs.TypeAttrs:
		// `attrs` 是**对象**（属性名 → 字符串值），不是字符串。少了这一支，set_spec_attrs /
		// set_layout_attrs 会给本地定型判死 —— 这条测试从此一直红，而它归 TTZS_E2E，默认不跑，
		// 所以直到 2026-09-25 才被看见（见到它红的那一刻，先怀疑的对象是当天的改动，不是它）。
		return map[string]string{"can_edit": "N"}
	}
	return "x"
}

func countParams(m *tzs.Manifest) int {
	n := 0
	for _, f := range m.Fns {
		n += len(f.Args)
	}
	return n
}

// TestE2EDocsVerbCountMatchesManifest 把"动词数"这个数字钉在 manifest 上。
//
// 为什么值得一条 E2E：同一个数字曾经在四处各不相同（49 / 50 / 52 同时存在），
// 而 `Index()` 打的那份**永远是对的**（它是 `len(m.Fns)`）—— 漂的只有手写的散文。
// 这就是"能问的就不抄"在文档上的形态：既然跑的时候引擎就在手边，数字就该被问一次。
//
// 声明怎么收集、为什么只认「N 个动词」这一种说法，见 tzs_verb_test.go 的
// docVerbCountClaims；"彼此一致"那一半由那边的 TestDocVerbCountsAgree 在不碰引擎的
// 情况下守着，这条只补最后一问：那个数字**真的是引擎的数吗**。
func TestE2EDocsVerbCountMatchesManifest(t *testing.T) {
	m := requireRealManifest(t)
	want := len(m.Fns)

	claims := docVerbCountClaims(t)
	for _, c := range claims {
		if c.n != want {
			t.Errorf("%s 写着 %d 个动词，manifest 里是 %d 个：\n    %s", c.where, c.n, want, c.line)
		}
	}
	// 一处都没找到不判失败：把数字整句删掉是允许的改法之一。但要让人在输出里看得见
	// 这条断言到底查了几处 —— 否则它会悄悄退化成空转。
	t.Logf("动词数 %d：查了 %d 处「N 个动词」的声明", want, len(claims))
}
