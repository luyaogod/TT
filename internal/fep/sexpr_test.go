package fep

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestVerbPrefixIsNotAnObject guards a silent failure mode: the wire carries a
// verb in front of the message (`om 3 {{…}}`, `meta Connection {{…}}`). If the
// verb were parsed as data, a whole screen would collapse into one object of
// type "om" with no name and no window.
func TestVerbPrefixIsNotAnObject(t *testing.T) {
	body := `om 3 {{an 0 Window 4097 {{name "axmt500_wf"}} {{Entry 5157 {{text "录入"}} {} {Value 260 {{col "xmda001"}} {}}}}}`
	nodes := ParseMessage(body)
	if len(nodes) == 0 {
		t.Fatal("no nodes parsed")
	}
	if nodes[0].Type != "Window" {
		t.Fatalf("root type = %q, want Window (the verb leaked into the parse)", nodes[0].Type)
	}
	m := NewModel()
	m.ApplyMessage(body)
	for _, id := range []int{4097, 5157, 260} {
		if _, ok := m.Get(id); !ok {
			t.Fatalf("object %d missing", id)
		}
	}
	chain, cycle, win := m.Chain(260)
	if cycle {
		t.Fatal("chain reports a cycle")
	}
	if win != 4097 || len(chain) < 3 {
		t.Fatalf("Value 260 must hang under Window 4097: window=%d chain=%d", win, len(chain))
	}
}

// TestBareObjectHeaderIsNotAVerb makes sure the verb skip does not eat a body
// that opens with an object header: cutting at the first brace would throw the
// type and the id away.
func TestBareObjectHeaderIsNotAVerb(t *testing.T) {
	body := `Table 501 {{col "xmdc007"}} {{TableColumn 502 {{name "xmdc_t.xmdc007"}{text "订购量"}} {}}}`
	nodes := ParseMessage(body)
	if len(nodes) < 2 {
		t.Fatalf("nodes = %d", len(nodes))
	}
	if nodes[0].Type != "Table" || nodes[0].ID != 501 {
		t.Fatalf("root = %s#%d, want Table#501", nodes[0].Type, nodes[0].ID)
	}
}

// TestTrimNoiseHandlesPrototypeRecordings covers the recordings written by the
// earlier prototype: part of the frame header stayed in front of the message.
func TestTrimNoiseHandlesPrototypeRecordings(t *testing.T) {
	noisy := "\x00\x00\x00+   +om 0 {{an 0 UserInterface 0 {{name \"axmt500_wf\"}}}}"
	got, ok := TrimNoise(noisy)
	if !ok || got[:2] != "om" {
		t.Fatalf("TrimNoise = %q,%v", got, ok)
	}
	if _, ok := TrimNoise("\x00\x00\x18\x05\x06\xff!!__cached__!!"); ok {
		t.Fatal("a proxy marker must not pass as a message")
	}
	if _, ok := TrimNoise(`meta Connection {{encoding "UTF-8"}}`); !ok {
		t.Fatal("meta must be recognized")
	}
}

// 真机上运行期的更新消息是这样发的：`om 7 {{un 271 {{value "X"}}}}` —— 属性组外面
// 还多一层花括号。少了这层，字段值就永远合并不进模型（而这是我们判断"输入生效了没有"
// 的唯一依据）。
func TestUnUpdateWithExtraBraceLevel(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(`om 7 {{un 271 {{value "X"}}}}`)
	o, ok := m.Get(271)
	if !ok {
		t.Fatal("对象没建出来")
	}
	if got := o.Attr("value"); got != "X" {
		t.Fatalf("value = %q，期望 X", got)
	}
}

// 真机序列：先定义字段，随后运行期用 un 回填值（`om 7 {{un 271 {{value "X"}}}}`）。
// 我们判断"输入生效了没有"就靠这个值。
func TestUnMergesIntoExistingDefinition(t *testing.T) {
	m := NewModel()
	m.ApplyMessage(`om 4 {{an 2858 FormField 271 {{name "xmda_t.xmda004"} {comment "cmt_xmda004"}} {}}}`)
	if o, ok := m.Get(271); !ok || o.Attr("comment") != "cmt_xmda004" {
		t.Fatalf("定义没建起来: %+v", o)
	}
	m.ApplyMessage(`om 7 {{un 271 {{value "CNJ-A50"}}}}`)
	o, _ := m.Get(271)
	if o.Attr("value") != "CNJ-A50" {
		t.Fatalf("un 没把值合进来: attrs=%v", o.Attrs)
	}
	if o.Attr("comment") != "cmt_xmda004" {
		t.Fatalf("un 把已有属性弄丢了: attrs=%v", o.Attrs)
	}
	if o.Parent != 2858 {
		t.Fatalf("un 改掉了父链: parent=%d", o.Parent)
	}
}

// TestDetectsArrayDialect 守 C6.2（契约 §1 C6.2）。
//
// 上游证据（three-threads.md §①）：同一条路径上**有两套方言**，由运行期全局开关
// （0x30c9b4）切换 —— 原生 GDC 走 S 表达式 `{an …}`，而 UR/HTML5 客户端或
// `gui.rendering` 启用时走**数组/JSON 方言**：`["a",0,"Window",178,{…}]`
// （容器 `[`、引号、`:` 分隔）。
//
// 我们只实现了 S 表达式那一套，所以数组方言必须**明确报错**：硬解它不会崩，
// 只会把窗口名/父链解析成垃圾 —— 而"解析出垃圾"和"方言不对"在调用方看起来
// 一模一样，这是本计划要防的两条静默出错路径之一（计划 §决策记录）。
func TestDetectsArrayDialect(t *testing.T) {
	// 合成输入，形状照 §① 的数组方言（不入库，只在本测试里用）。
	const array = `["a",0,"Window",178,{"name":"axmt500_wf"}]`

	_, err := ParseMessageChecked(array)
	if err == nil {
		t.Fatal("数组方言被当成 S 表达式硬解了（必须报错）")
	}
	if !errors.Is(err, ErrUnsupportedDialect) {
		t.Fatalf("错误必须能被 errors.Is 判成 ErrUnsupportedDialect，实得 %v", err)
	}
	// 原因文案要说出"是哪种方言的迹象"，而不是一句"解析失败"。
	if !strings.Contains(err.Error(), "数组") || !strings.Contains(err.Error(), "[") {
		t.Fatalf("原因文案没说清是数组/JSON 方言：%v", err)
	}

	// 真机载荷常带前置空白（帧尾换行、缩进）：去空白后是 `[` 就必须命中。
	for _, in := range []string{" \r\n\t" + array, "\n" + array} {
		if _, err := ParseMessageChecked(in); !errors.Is(err, ErrUnsupportedDialect) {
			t.Fatalf("带空白前缀的数组方言漏过了：%q → %v", in, err)
		}
	}

	// 探测本身也要能单独调用（调用方想在解析前先问一句"这是什么方言"）。
	if err := CheckDialect(array); !errors.Is(err, ErrUnsupportedDialect) {
		t.Fatalf("CheckDialect 数组方言 = %v", err)
	}
	if err := CheckDialect(`om 3 {{}}`); err != nil {
		t.Fatalf("S 表达式被判定成了别的方言：%v", err)
	}

	// 既有 S 表达式输入行为**一个字都不变**：同一份输入，节点与 ParseMessage 完全一致。
	const sexpr = `om 3 {{an 0 Window 4097 {{name "axmt500_wf"}} {{an 4097 FormField 271 {{name "xmda004"}} {}}}}}`
	got, err := ParseMessageChecked(sexpr)
	if err != nil {
		t.Fatalf("S 表达式被误判成未支持方言：%v", err)
	}
	want := ParseMessage(sexpr)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("S 表达式路径被改动了：got %+v want %+v", got, want)
	}
	if len(want) != 2 || want[0].Type != "Window" || want[1].ID != 271 {
		t.Fatalf("对照样本本身不对：%+v", want)
	}
	// 反过来：被判定为数组方言的输入**不能**产出任何节点。
	if nodes, _ := ParseMessageChecked(array); nodes != nil {
		t.Fatalf("数组方言产出了节点（硬解了）：%+v", nodes)
	}
}

// TestDetectsArrayDialectBehindVerbPrefix 守 **A16**（契约 §4：C6.2 判据放宽一格）。
//
// 原判据只认"去空白后以 `[` 开头"，于是**带动词前缀**的数组方言（`om 3 ["a",…]`）
// 会漏过探测、被交给 S 表达式解析器 —— 静默解成垃圾。放宽后判据改为"看**第一个结构字符**"：
// S 表达式的结构必然从 `{` 开始（对象声明/属性组；语料 7,805 帧无一例外），
// 数组方言的结构从 `[` 开始，谁先出现就定谁 —— 前缀是什么都不影响。
func TestDetectsArrayDialectBehindVerbPrefix(t *testing.T) {
	for _, in := range []string{
		`om 3 ["a",0,"Window",1,{}]`,         // 动词前缀 + 数组方言（A16 的原例）
		`meta Connection ["a",0,"Window",1]`, // 别的动词前缀也一样
		`  om 9 ["u",178,{"value":"X"}]`,     // 前缀 + 前置空白
	} {
		if err := CheckDialect(in); !errors.Is(err, ErrUnsupportedDialect) {
			t.Fatalf("带前缀的数组方言漏过了：%q → %v", in, err)
		}
		nodes, err := ParseMessageChecked(in)
		if !errors.Is(err, ErrUnsupportedDialect) {
			t.Fatalf("ParseMessageChecked 没拦住 %q → %v", in, err)
		}
		if nodes != nil {
			t.Fatalf("带前缀的数组方言产出了节点（硬解了）：%q → %+v", in, nodes)
		}
	}
}

// TestUnknownShapeIsReportedWithoutBlamingDialect 守 **A16**（契约 §4）。
//
// 两件事必须分得开：**"方言不对"**（`ErrUnsupportedDialect`）与**"这不是一条消息"**
// （既不是 `{` 结构的 S 表达式，也不是 `[` 结构的数组方言 → 未知形状）。
// 合成一个错误码，调用方就没法决定"是换个方言还是重取字节"。
//
// 同时钉住**不误拒**：空/纯空白照旧无错（`ParseMessage("")` 的既有行为），
// 两种合法 S 表达式形态（动词前缀开头、裸对象头开头）也照旧放行 —— 放宽判据
// 绝不能变成"把合法消息认成垃圾"。
func TestUnknownShapeIsReportedWithoutBlamingDialect(t *testing.T) {
	_, err := ParseMessageChecked(`xyz`)
	if err == nil {
		t.Fatal("未知形状被静默放行了（既不是 { 也不是 [）")
	}
	if errors.Is(err, ErrUnsupportedDialect) {
		t.Fatalf("未知形状被误报成方言错误（两件事混了）：%v", err)
	}
	if !errors.Is(err, ErrUnknownShape) {
		t.Fatalf("错误必须能被 errors.Is 判成 ErrUnknownShape，实得 %v", err)
	}

	// 空 / 纯空白：不报（钉住"不误拒"）。
	for _, in := range []string{"", " ", "\r\n\t  "} {
		nodes, err := ParseMessageChecked(in)
		if err != nil {
			t.Fatalf("空输入被拒了：%q → %v", in, err)
		}
		if len(nodes) != 0 {
			t.Fatalf("空输入产出了节点：%q → %+v", in, nodes)
		}
	}

	// 合法 S 表达式的三种形态都不能被误拒。
	for _, in := range []string{
		`om 3 {{an 0 Window 4097 {{name "w"}}}}`,                                // 动词前缀
		`Table 501 {{col "xmdc007"}} {{TableColumn 502 {{name "xmdc007"}} {}}}`, // 裸对象头（本包文档里的合法形态）
		`{an 0 Window 5 {{}}}`,
	} {
		if err := CheckDialect(in); err != nil {
			t.Fatalf("合法 S 表达式被误拒：%q → %v", in, err)
		}
	}
}
