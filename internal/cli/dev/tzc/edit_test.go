package tzc

// 区域级动词（points/show/find/patch/set）的测试。
//
// 这层不新增写回引擎：patch/set 只改 workspace，落包仍走 verify/apply。
// 所以断言落在两件事上：
//  1. 区域定位对不对（尤其 section 的正文在两行围栏之间，且结束围栏不带名字）；
//  2. 改完之后 apply 仍能跑通（证明没破坏三道闸门的前提）。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/dev/model"
	"github.com/luyaogod/TT/internal/testkit"
)

func TestBodyRange_SectionFences(t *testing.T) {
	text := []byte("head\n{//@tdev:begin section s.a [READONLY deny=\"x\"]}\nLINE1\nLINE2\n{//@tdev:end section}\ntail\n")
	r := &model.Region{Name: "s.a", Kind: model.RegionSection}
	s, e, ok := bodyRange(text, r)
	if !ok {
		t.Fatal("区段正文没定位到（结束围栏 `{//@tdev:end section}` 不带名字）")
	}
	if got := string(text[s:e]); got != "LINE1\nLINE2\n" {
		t.Errorf("正文 = %q", got)
	}
}

func TestBodyRange_PointUsesBodySpan(t *testing.T) {
	text := []byte("AAAABBBBCCCC")
	r := &model.Region{Name: "p", Kind: model.RegionPoint, BodySpan: &model.ByteRange{Start: 4, End: 8}}
	if b, ok := bodyOf(text, r); !ok || string(b) != "BBBB" {
		t.Fatalf("点正文 = %q ok=%v", b, ok)
	}
}

func TestRegionLines(t *testing.T) {
	cases := map[string]int{"": 0, "\n": 0, "a": 1, "a\n": 1, "a\nb": 2, "a\r\nb\r\n": 2}
	for in, want := range cases {
		if got := regionLines([]byte(in)); got != want {
			t.Errorf("regionLines(%q) = %d，想要 %d", in, got, want)
		}
	}
}

func TestFenceToken(t *testing.T) {
	if fenceToken("普通代码 LET x = 1") != "" {
		t.Error("正常代码被误拒")
	}
	for _, bad := range []string{"{//@tdev:begin point x}", "a\n{//@tdev:end point}\nb"} {
		if fenceToken(bad) == "" {
			t.Errorf("围栏注入应被拒：%q", bad)
		}
	}
}

// 子串替换拼接坏行：old 前面已有一个 #，new 又以 # 开头 → 必须拒。
func TestLineAsymmetry(t *testing.T) {
	body := "x\n  #檢查失敗時後續處理\ny\n"
	off := strings.Index(body, "檢查失敗")
	if why := lineAsymmetry(body, off, "檢查失敗時後續處理", "#新提示"); why == "" {
		t.Error("会产生 ## 的替换应被拒")
	}
	if why := lineAsymmetry(body, off, "檢查失敗時後續處理", "檢查失敗（改）"); why != "" {
		t.Errorf("正常替换被误拒：%s", why)
	}
	if why := lineAsymmetry(body, 0, "x", "x\n"); why == "" {
		t.Error("凭空多空行的替换应被拒")
	}
}

// pickEditablePoint 在 fixture 工作区里挑一个"可编辑、正文非空且不太大"的点。
func pickEditablePoint(t *testing.T, ws string) (*model.Region, []byte, *editCtx) {
	t.Helper()
	ec, err := openForEdit(ws)
	if err != nil {
		t.Fatalf("openForEdit: %v", err)
	}
	var found *model.Region
	var foundBody []byte
	var walk func(rs []*model.Region)
	walk = func(rs []*model.Region) {
		for _, r := range rs {
			if found == nil && r.Kind == model.RegionPoint && r.Editable {
				if b, ok := bodyOf(ec.edited, r); ok && regionLines(b) <= 20 {
					found, foundBody = r, b
				}
			}
			walk(r.Children)
		}
	}
	walk(ec.parsed.Doc.Regions)
	if found != nil {
		return found, foundBody, ec
	}
	t.Fatal("fixture 里找不到可编辑的非空点")
	return nil, nil, nil
}

func TestSetRegion_WritesAndApplyStillWorks(t *testing.T) {
	dir := t.TempDir()
	ws, _, _ := exportFixture(t, dir)
	r, _, _ := pickEditablePoint(t, ws)

	newFile := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(newFile, []byte("// edited-by-test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := testkit.Silent(t, func() int { return cmdSetRegion([]string{ws, r.Name, "--file", newFile}) }); code != 0 {
		t.Fatalf("set 退出码 %d", code)
	}
	ec, err := openForEdit(ws)
	if err != nil {
		t.Fatalf("改后解析失败: %v", err)
	}
	b, _ := bodyOf(ec.edited, findRegion(ec.parsed.Doc, r.Name))
	if !strings.Contains(string(b), "edited-by-test") {
		t.Fatalf("正文没写进去：%q", b)
	}
	// 关键：区域级改动之后，原来的写回管线仍然可用
	if code := testkit.Silent(t, func() int { return cmdApply([]string{ws, "--dry-run"}) }); code != 0 {
		t.Fatalf("apply --dry-run 退出码 %d（区域编辑破坏了管线）", code)
	}
}

func TestPatchRegion_UniqueThenAmbiguous(t *testing.T) {
	dir := t.TempDir()
	ws, _, _ := exportFixture(t, dir)
	r, _, _ := pickEditablePoint(t, ws)

	// fixture 的点是空的：先 set 两行已知内容，再用它当锚点
	seed := filepath.Join(dir, "seed.txt")
	if err := os.WriteFile(seed, []byte("AAA_anchor\nBBB_line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := testkit.Silent(t, func() int { return cmdSetRegion([]string{ws, r.Name, "--file", seed}) }); code != 0 {
		t.Fatalf("seed set 退出码 %d", code)
	}
	anchor := "AAA_anchor"
	if code := testkit.Silent(t, func() int {
		return cmdPatchRegion([]string{ws, r.Name, "--old", anchor, "--new", anchor + " // patched"})
	}); code != 0 {
		t.Fatalf("patch 退出码 %d", code)
	}
	ec, _ := openForEdit(ws)
	b, _ := bodyOf(ec.edited, findRegion(ec.parsed.Doc, r.Name))
	if !strings.Contains(string(b), "// patched") {
		t.Fatal("patch 没生效")
	}

	// 不唯一要拒：拿正文的第一行去当锚点之外，再试一个必然多处/不存在的串
	code := testkit.Silent(t, func() int {
		return cmdPatchRegion([]string{ws, r.Name, "--old", "END IF", "--new", "ENDIF"})
	})
	if code == 0 {
		t.Fatal("不唯一/不存在的锚点不该成功")
	}
}

// 整段替换的大小闸门：用 --max-lines 0 把阈值压到"任何非空区域都超"，结果是确定的，
// 不依赖 fixture 里恰好有没有 300 行的大点（不写 skip：跳过会让这条守卫静默空转）。
func TestSetRegion_RefusesBigRegion(t *testing.T) {
	dir := t.TempDir()
	ws, _, _ := exportFixture(t, dir)
	r, _, _ := pickEditablePoint(t, ws)

	seed := filepath.Join(dir, "seed2.txt")
	if err := os.WriteFile(seed, []byte("ONE_LINE\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := testkit.Silent(t, func() int { return cmdSetRegion([]string{ws, r.Name, "--file", seed}) }); code != 0 {
		t.Fatalf("seed set 退出码 %d", code)
	}
	code := testkit.Silent(t, func() int {
		return cmdSetRegion([]string{ws, r.Name, "--file", seed, "--max-lines", "0"})
	})
	if code == 0 {
		t.Fatal("正文非空时，--max-lines 0 必须拒（闸门失效）")
	}
}
