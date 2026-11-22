package model

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 本文件补的是 model 包**纯数据/纯函数**的地基：区间语义、拒绝原因文案、
// 解锁状态推导、文档遍历与「自身字节」剔除。
//
// 为什么这些值得单独钉：它们全是**口径**（半开区间、半开还是闭、CRLF 等价、
// 深拷不能共享切片），口径错了不会 panic，只会让上层 gate/split 静默误判 ——
// 例如 OwnBytes 少剥一个换行，gate1 就会把「父区段没改」误报成被改。
// 上面的 lines_test.go 已覆盖行号/首差异，这里不重复。

// TestByteRangeHalfOpen 钉住半开区间 [Start, End) 的四个方法。
//
// 半开是整套字节坐标的约定：End 不属于区间。差一个等于号，
// Contains/Overlaps 就会在边界上给出相反答案，进而影响「可写区间」判定。
func TestByteRangeHalfOpen(t *testing.T) {
	r := ByteRange{Start: 2, End: 5}
	if r.Len() != 3 {
		t.Errorf("Len = %d，想 3", r.Len())
	}
	if r.Empty() {
		t.Error("[2,5) 不是空区间")
	}
	// 半开：Start 在、End 不在。
	for off, want := range map[int]bool{1: false, 2: true, 4: true, 5: false, 6: false} {
		if got := r.Contains(off); got != want {
			t.Errorf("Contains(%d) = %v，想 %v", off, got, want)
		}
	}
	// 空区间：Len=0、Empty=true、不含任何偏移。
	for _, e := range []ByteRange{{Start: 3, End: 3}, {Start: 7, End: 3}, {}} {
		if e.Len() != 0 || !e.Empty() {
			t.Errorf("%+v 该是空区间（Len=0 Empty=true），得 Len=%d Empty=%v", e, e.Len(), e.Empty())
		}
		if e.Contains(e.Start) {
			t.Errorf("空区间 %+v 不该包含任何偏移（含 Start）", e)
		}
	}

	// Overlaps 也是半开：相邻区间 [2,5) 与 [5,9) 不相交（共享端点不算）。
	cases := []struct {
		a, b ByteRange
		want bool
	}{
		{ByteRange{2, 5}, ByteRange{5, 9}, false}, // 首尾相接不算交
		{ByteRange{2, 5}, ByteRange{4, 9}, true},
		{ByteRange{2, 5}, ByteRange{0, 2}, false},
		{ByteRange{2, 5}, ByteRange{0, 3}, true},
		{ByteRange{2, 5}, ByteRange{3, 4}, true}, // 包含
		{ByteRange{2, 5}, ByteRange{2, 5}, true}, // 相同
		{ByteRange{2, 5}, ByteRange{6, 9}, false},
		{ByteRange{}, ByteRange{}, false}, // 空区间不相交
	}
	for _, c := range cases {
		if got := c.a.Overlaps(c.b); got != c.want {
			t.Errorf("%+v.Overlaps(%+v) = %v，想 %v", c.a, c.b, got, c.want)
		}
		// s 对称：相交关系不该依赖参数顺序。
		if got := c.b.Overlaps(c.a); got != c.want {
			t.Errorf("Overlaps 不对称：%+v vs %+v = %v，想 %v", c.b, c.a, got, c.want)
		}
	}
}

// TestDenyReasonTextEveryCode 每个原因码都必须给出**中文解释**（manifest 的 reason 字段），
// 漏一个就会让用户在 manifest 里看到机器码看不懂；未知码原样返回（不吞掉）。
func TestDenyReasonTextEveryCode(t *testing.T) {
	codes := []string{
		DenyNone, DenyReadonlyAttr, DenyCiteStd, DenyIndFunMismatch, DenyIndustryMemo,
		DenyEnvEditEnv, DenyTopstdMode, DenyLoginTopstd, DenySectionAnchor,
		DenySectionReadonly, DenySecLocked, DenyNotExported, DenyStructureUnparsable,
		DenySectionTemplateMissing, DenyOnlyPoints,
	}
	seen := map[string]bool{}
	for _, c := range codes {
		if seen[c] {
			t.Errorf("原因码 %q 重复出现", c)
		}
		seen[c] = true
		if c == DenyNone {
			if got := DenyReasonText(c); got != "" {
				t.Errorf("DenyNone 该给空串，得 %q", got)
			}
			continue
		}
		got := DenyReasonText(c)
		if got == "" {
			t.Errorf("原因码 %q 没有中文解释", c)
		}
		if got == c {
			t.Errorf("原因码 %q 被原样返回，等于没写文案", c)
		}
		if !strings.ContainsAny(got, "区段点框") && !strings.ContainsAny(got, "程序标准环") {
			t.Errorf("原因码 %q 的解释不像中文：%q", c, got)
		}
	}
	// 未知码原样返回：宁可显示机器码，也不要显示空白或误配的文案。
	if got := DenyReasonText("no-such-code"); got != "no-such-code" {
		t.Errorf("未知码该原样返回，得 %q", got)
	}
}

// TestDenyCodesAreDistinct 原因码字符串是**对外契约**（写进围栏行 deny= 与 regions.json）。
// 一旦重复，两个不同的拒绝会变成同一个，排查时分不清是权限不够还是没解锁。
func TestDenyCodesAreDistinct(t *testing.T) {
	all := []string{
		DenyNone, DenyReadonlyAttr, DenyCiteStd, DenyIndFunMismatch, DenyIndustryMemo,
		DenyEnvEditEnv, DenyTopstdMode, DenyLoginTopstd, DenySectionAnchor,
		DenySectionReadonly, DenySecLocked, DenyNotExported, DenyStructureUnparsable,
		DenySectionTemplateMissing, DenyOnlyPoints,
	}
	seen := map[string]string{}
	for _, c := range all {
		if prev, dup := seen[c]; dup {
			t.Errorf("%q 与 %q 是同一个原因码", prev, c)
		}
		seen[c] = c
	}
	// 机器码用 kebab-case，便于 grep 与写进围栏。
	for _, c := range all {
		if c == "" {
			continue
		}
		if strings.ContainsAny(c, " _") || strings.ToLower(c) != c {
			t.Errorf("原因码 %q 不是小写 kebab-case", c)
		}
	}
}

// TestEffectiveSectionState 钉住解锁状态的单一事实来源：
// 包级 section_flag="Y" 优先；否则看 workspace 的 pending_unlock；两者都无 → Locked。
//
// 这条错了会把「没解锁的包」放行去改框架区段（红线 R4 的入口）。
func TestEffectiveSectionState(t *testing.T) {
	cases := []struct {
		flag, pending bool
		wantState     SectionState
		wantBy        string
	}{
		{true, true, SectionUnlocked, UnlockedByPkg}, // 包级事实优先
		{true, false, SectionUnlocked, UnlockedByPkg},
		{false, true, SectionUnlocked, UnlockedByUnlockCmd},
		{false, false, SectionLocked, ""},
	}
	for _, c := range cases {
		if got := EffectiveSectionState(c.flag, c.pending); got != c.wantState {
			t.Errorf("EffectiveSectionState(flag=%v,pending=%v) = %q，想 %q", c.flag, c.pending, got, c.wantState)
		}
		if got := UnlockedByOf(c.flag, c.pending); got != c.wantBy {
			t.Errorf("UnlockedByOf(flag=%v,pending=%v) = %q，想 %q", c.flag, c.pending, got, c.wantBy)
		}
	}
	// 常量值是对外契约（写进 regions.json / manifest）。
	if SectionLocked != "Locked" || SectionUnlocked != "Unlocked" {
		t.Errorf("SectionState 取值变了：%q %q", SectionLocked, SectionUnlocked)
	}
	if UnlockedByPkg != "pkg" || UnlockedByUnlockCmd != "unlock-cmd" {
		t.Errorf("UnlockedBy 取值变了：%q %q", UnlockedByPkg, UnlockedByUnlockCmd)
	}
}

// regionFixture 造一棵 section → point 嵌套与一个平级 section，供遍历/区间测试用。
func regionFixture() *Document {
	point := &Region{
		Name: "function.a", Kind: RegionPoint, Editable: true,
		Meta:           map[string]string{"new": "Y"},
		TglTag:         []byte(`{<point name="function.a"/>}`),
		ContentSpan:    ByteRange{Start: 10, End: 25},
		FullSpan:       ByteRange{Start: 10, End: 25},
		BodySpan:       &ByteRange{Start: 12, End: 20},
		SignatureSpan:  &ByteRange{Start: 10, End: 12},
		DescSpan:       &ByteRange{Start: 8, End: 10},
		StructureSpans: []ByteRange{{Start: 20, End: 25}},
	}
	section := &Region{
		Name: "p.main", Kind: RegionSection, Editable: true,
		Meta:        map[string]string{"src": "s"},
		ContentSpan: ByteRange{Start: 0, End: 40},
		FullSpan:    ByteRange{Start: 0, End: 45},
		Children:    []*Region{point},
	}
	roSection := &Region{
		Name: "p.ro", Kind: RegionSection, Editable: false,
		DenyCode: DenySecLocked, Reason: DenyReasonText(DenySecLocked),
		ContentSpan: ByteRange{Start: 50, End: 60},
		FullSpan:    ByteRange{Start: 50, End: 65},
	}
	return &Document{
		Text:    bytes.Repeat([]byte("x"), 80),
		Regions: []*Region{section, roSection},
	}
}

// TestDeepCopyIsDeep 深拷必须**不共享**任何可变字段（fence 靠它在不改基线的前提下改区间）。
// 共享切片/映射最阴险：测试里改副本看起来没问题，真机上会把基线一起改掉。
func TestDeepCopyIsDeep(t *testing.T) {
	orig := regionFixture()
	sectionCopy := orig.Regions[0].DeepCopy()
	pointCopy := sectionCopy.Children[0]

	// 改副本身上的所有可变字段，原件必须纹丝不动。
	sectionCopy.Meta["src"] = "changed"
	pointCopy.Meta["new"] = "N"
	pointCopy.TglTag[0] = 'Z'
	pointCopy.StructureSpans[0].Start = 999
	pointCopy.BodySpan.Start = 999
	pointCopy.SignatureSpan.Start = 998
	pointCopy.DescSpan.Start = 997
	pointCopy.Name = "changed"

	p := orig.Regions[0]
	if p.Meta["src"] != "s" {
		t.Error("改副本的 Meta 污染了原件")
	}
	pp := p.Children[0]
	if pp.Meta["new"] != "Y" {
		t.Error("改副本子点的 Meta 污染了原件")
	}
	if pp.TglTag[0] != '{' {
		t.Error("改副本的 TglTag 污染了原件")
	}
	if pp.StructureSpans[0].Start != 20 {
		t.Error("改副本的 StructureSpans 污染了原件")
	}
	if pp.BodySpan.Start != 12 || pp.SignatureSpan.Start != 10 || pp.DescSpan.Start != 8 {
		t.Error("改副本的 span 指针污染了原件")
	}
	if pp.Name != "function.a" {
		t.Error("改副本的点名污染了原件")
	}
	// nil 字段深拷后仍为 nil；无子区间的区段不 panic。
	ro := orig.Regions[1].DeepCopy()
	if ro.Children != nil || ro.TglTag != nil || ro.BodySpan != nil {
		t.Errorf("nil 字段深拷后该仍为 nil：%+v", ro)
	}
}

// TestWalkAndAllRegions 遍历是**先序**（父在子前），AllRegions 按文档顺序展平 ——
// 顺序错了会让 fillPositions、I 系列检查的「第一个命中」变成另一个区段。
func TestWalkAndAllRegions(t *testing.T) {
	doc := regionFixture()
	var names []string
	doc.Regions[0].Walk(func(r *Region) { names = append(names, r.Name) })
	if got := strings.Join(names, ","); got != "p.main,function.a" {
		t.Errorf("Walk 顺序 = %q，想先父后子", got)
	}
	all := doc.AllRegions()
	if len(all) != 3 {
		t.Fatalf("AllRegions = %d 个，想 3", len(all))
	}
	if got := strings.Join([]string{all[0].Name, all[1].Name, all[2].Name}, ","); got != "p.main,function.a,p.ro" {
		t.Errorf("AllRegions 顺序 = %q", got)
	}
	// FindRegion 要能下钻到嵌套子区间；找不到返回 nil。
	if r := doc.FindRegion("function.a"); r == nil || r.Kind != RegionPoint {
		t.Errorf("FindRegion 没找到嵌套子区间：%+v", r)
	}
	if r := doc.FindRegion("nope"); r != nil {
		t.Errorf("找不到该返回 nil，得 %+v", r)
	}
	// 空文档不 panic。
	var empty Document
	if len(empty.AllRegions()) != 0 || empty.FindRegion("x") != nil {
		t.Error("空文档应有空结果")
	}
}

// TestEditableSpans 钉住 gate1 的可写区间口径：
//   - 非 Editable 的区段/点一个字节都不算可写；
//   - 自订定义点 = DescSpan + BodySpan（**不含** SignatureSpan / StructureSpans）；
//   - 裸名点（无 BodySpan）= 整个 ContentSpan；
//   - 空区间不登记；结果按起点升序。
func TestEditableSpans(t *testing.T) {
	point := &Region{
		Name: "function.a", Kind: RegionPoint, Editable: true,
		ContentSpan:   ByteRange{Start: 10, End: 30},
		DescSpan:      &ByteRange{Start: 10, End: 15},
		BodySpan:      &ByteRange{Start: 16, End: 28},
		SignatureSpan: &ByteRange{Start: 15, End: 16},
		StructureSpans: []ByteRange{
			{Start: 28, End: 30},
		},
	}
	plain := &Region{
		Name: "global.memo", Kind: RegionPoint, Editable: true, Plain: true,
		ContentSpan: ByteRange{Start: 40, End: 50},
		BodySpan:    &ByteRange{Start: 40, End: 50},
	}
	ro := &Region{Name: "p.ro", Kind: RegionSection, Editable: false,
		ContentSpan: ByteRange{Start: 0, End: 8}}
	emptyBody := &Region{Name: "global.empty", Kind: RegionPoint, Editable: true,
		ContentSpan: ByteRange{Start: 60, End: 60},
		BodySpan:    &ByteRange{Start: 60, End: 60}}
	doc := &Document{
		Text:    bytes.Repeat([]byte("x"), 100),
		Regions: []*Region{{Kind: RegionSection, Name: "p.s", Editable: true, ContentSpan: ByteRange{Start: 0, End: 8}, Children: []*Region{point}}, plain, ro, emptyBody},
	}
	got := doc.EditableSpans()
	want := []ByteRange{{0, 8}, {10, 15}, {16, 28}, {40, 50}}
	if len(got) != len(want) {
		t.Fatalf("EditableSpans = %+v，想 %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("区间[%d] = %+v，想 %+v", i, got[i], want[i])
		}
	}
	// 结构行（END 行）与签名行都**不在**可写面里。
	for _, sp := range got {
		if sp.Start == 28 && sp.End == 30 {
			t.Error("StructureSpans 不该出现在可写区间里")
		}
		if sp.Start == 15 && sp.End == 16 {
			t.Error("SignatureSpan 不该出现在可写区间里")
		}
	}
}

// TestOwnBytesExciseChildren 区段的「自身字节」= 正文剔除所有子区间（含其行尾）。
//
// 这是区段与子点责任分离的地基：子点改了不该把父区段判成被改，
// 但父区段自己的字节改动必须暴露。剔除时**吃掉子区间后面的那一个行尾**是关键 ——
// 不吃掉的话子点合法改动会留下/带走换行，父区段就被误判。
func TestOwnBytesExciseChildren(t *testing.T) {
	// 文本：内容 [0,14)，子区间 A=[2,4)、B=[8,10)；A 后跟 \r\n，B 后跟 \n。
	text := []byte("xxAy\r\nzzBq\nww!")
	sec := &Region{
		Name: "p.s", Kind: RegionSection,
		ContentSpan: ByteRange{Start: 0, End: 14},
		Children: []*Region{
			{FullSpan: ByteRange{Start: 2, End: 4}},
			{FullSpan: ByteRange{Start: 8, End: 10}},
		},
	}
	doc := &Document{Text: text}
	// 期望：保留 "xx" + "zz" + "ww!"（A/B 及各自的 \r\n / \n 都被剔除）。
	if got := string(doc.OwnBytes(sec)); got != "xxzzww!" {
		t.Errorf("OwnBytes = %q，想 %q（子区间及其后一个行尾都要剔除）", got, "xxzzww!")
	}

	// 无子区间 = 整个 ContentSpan；边界越界/倒置返回 nil（不 panic）。
	leaf := &Region{ContentSpan: ByteRange{Start: 0, End: 5}}
	if got := doc.OwnBytes(leaf); !bytes.Equal(got, text[:5]) {
		t.Errorf("无子区间该返回 ContentSpan 原文，得 %q", got)
	}
	if got := doc.OwnBytes(nil); got != nil {
		t.Errorf("nil region 该返回 nil，得 %q", got)
	}
	bad := &Region{ContentSpan: ByteRange{Start: 2, End: 99}}
	if got := doc.OwnBytes(bad); got != nil {
		t.Errorf("末端越界该返回 nil，得 %q", got)
	}
	rev := &Region{ContentSpan: ByteRange{Start: 9, End: 3}}
	if got := doc.OwnBytes(rev); got != nil {
		t.Errorf("起止倒置该返回 nil，得 %q", got)
	}

	// 子区间在文本末尾、后面没有行尾：只剔除子区间本身。
	text2 := []byte("abXY")
	sec2 := &Region{ContentSpan: ByteRange{Start: 0, End: 4},
		Children: []*Region{{FullSpan: ByteRange{Start: 2, End: 4}}}}
	doc2 := &Document{Text: text2}
	if got := string(doc2.OwnBytes(sec2)); got != "ab" {
		t.Errorf("末尾子区间：得 %q，想 %q", got, "ab")
	}
}

// TestNormalizeEOLAndEqualEOL 行尾等价是**真机事故的修复**：编辑器把 CRLF 归一成 LF
// 不应被当成改动。注意单独 CR 也折成 LF；没有 CR 时必须原样返回同一片字节（快路径）。
func TestNormalizeEOLAndEqualEOL(t *testing.T) {
	// 无 CR → 原样（甚至允许共享底层数组，这里只要求内容相同）。
	in := []byte("a\nb\nc")
	if got := NormalizeEOL(in); !bytes.Equal(got, in) {
		t.Errorf("无 CR 时该原样返回，得 %q", got)
	}
	cases := []struct{ in, want string }{
		{"a\r\nb", "a\nb"},
		{"a\rb", "a\nb"},
		{"\r\n\r\n", "\n\n"},
		{"a\n\r\nb\rc", "a\n\nb\nc"},
		{"", ""},
	}
	for _, c := range cases {
		if got := string(NormalizeEOL([]byte(c.in))); got != c.want {
			t.Errorf("NormalizeEOL(%q) = %q，想 %q", c.in, got, c.want)
		}
	}
	// EqualEOL：完全相同当然相等，纯行尾差异相等，真实内容差异不等。
	if !EqualEOL([]byte("a\r\nb"), []byte("a\r\nb")) {
		t.Error("相同字节该相等")
	}
	if !EqualEOL([]byte("a\r\nb"), []byte("a\nb")) {
		t.Error("纯 CRLF↔LF 差异该算相等")
	}
	if !EqualEOL([]byte("a\rb"), []byte("a\nb")) {
		t.Error("单独 CR↔LF 差异该算相等")
	}
	if EqualEOL([]byte("a\nb"), []byte("a\nc")) {
		t.Error("内容不同不该相等")
	}
	if EqualEOL([]byte("a\nb"), []byte("a\nb\n")) {
		t.Error("多一个行尾不算等价")
	}
}

// TestSha256Helpers 摘要口径（小写 hex、File 版本要能读文件/报错）。
func TestSha256Helpers(t *testing.T) {
	// 空串的 SHA-256 是众所周知的常量，钉死口径（大小写 + 长度）。
	const emptySHA = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := Sha256Bytes(nil); got != emptySHA {
		t.Errorf("空字节摘要 = %q，想 %q", got, emptySHA)
	}
	if got := Sha256Bytes([]byte("abc")); got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Errorf("abc 摘要 = %q", got)
	}
	if len(Sha256Bytes([]byte("x"))) != 64 {
		t.Error("摘要该是 64 位小写 hex")
	}

	dir := t.TempDir()
	p := filepath.Join(dir, "f.bin")
	if err := os.WriteFile(p, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Sha256File(p)
	if err != nil {
		t.Fatalf("读文件算摘要失败：%v", err)
	}
	if got != Sha256Bytes([]byte("abc")) {
		t.Errorf("文件摘要应与字节摘要一致，得 %q", got)
	}
	if _, err := Sha256File(filepath.Join(dir, "missing.bin")); err == nil {
		t.Error("文件不存在该报错")
	}
}

// TestMarshalJSONStable 稳定序列化是 manifest/regions.json 的字节契约：
// 两空格缩进 + 末尾换行；编码失败要原样返回错误（不能吞）。
func TestMarshalJSONStable(t *testing.T) {
	b, err := MarshalJSONStable(map[string]string{"b": "2", "a": "1"})
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"a\": \"1\",\n  \"b\": \"2\"\n}\n"
	if string(b) != want {
		t.Errorf("得 %q，想 %q", string(b), want)
	}
	// 结果必须可被标准库解回等价结构。
	var back map[string]string
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("输出不是合法 JSON：%v", err)
	}
	if back["a"] != "1" || back["b"] != "2" {
		t.Errorf("回读不对：%+v", back)
	}
	// 无法编码的值（通道）要报错而不是返回半截。
	if _, err := MarshalJSONStable(make(chan int)); err == nil {
		t.Error("不可编码的值该报错")
	}
}
