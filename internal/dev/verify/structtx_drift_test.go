package verify

import (
	"bytes"
	"strings"
	"testing"

	"tt/internal/dev/fence"
	"tt/internal/dev/fgl"
	"tt/internal/dev/model"
	"tt/internal/dev/pkgfile"
	"tt/internal/dev/synth"
	"tt/internal/dev/tapfile"
	"tt/internal/dev/testutil"
)

// driftFixture 造一个**老工作区**：围栏里没有 fn/desc 字段，且 regions.json 的
// SignatureSpan / DescSpan 停在「未插围栏」的坐标系上（修好 Render 之前 export 的
// 工作区就是这个形状）。
func driftFixture(t *testing.T) (base *model.Document, fenced []byte, point *model.Region) {
	t.Helper()
	p, err := testutil.NormalPkgPath(t.TempDir(), "adzi999")
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := pkgfile.Open(p, pkgfile.OpenOptions{})
	if err != nil {
		t.Fatal(err)
	}
	doc, err := synth.Synthesize(pkg, synth.Options{})
	if err != nil {
		t.Fatal(err)
	}
	clearFnDesc(doc.Regions)

	fenced, regions, spans, err := fence.Render(doc)
	if err != nil {
		t.Fatal(err)
	}
	// 把结构区间改回未围栏坐标：这正是旧版本 Render 留下的坏表。
	legacyDrift(doc, regions)

	point = firstPoint(regions)
	if point == nil {
		t.Fatal("合成包里没有自订定义点")
	}
	b := *doc
	b.Text = fenced
	b.Regions = regions
	b.Spans = spans
	return &b, fenced, point
}

func clearFnDesc(rs []*model.Region) {
	for _, r := range rs {
		if r.Kind == model.RegionPoint {
			r.Fn = ""
			r.Desc = ""
		}
		clearFnDesc(r.Children)
	}
}

func legacyDrift(doc *model.Document, regions []*model.Region) {
	for _, r := range regions {
		if orig := doc.FindRegion(r.Name); orig != nil {
			if orig.SignatureSpan != nil {
				r.SignatureSpan = &model.ByteRange{Start: orig.SignatureSpan.Start, End: orig.SignatureSpan.End}
			}
			if orig.DescSpan != nil {
				r.DescSpan = &model.ByteRange{Start: orig.DescSpan.Start, End: orig.DescSpan.End}
			}
		}
		legacyDrift(doc, r.Children)
	}
}

func firstPoint(rs []*model.Region) *model.Region {
	for _, r := range rs {
		if r.Kind == model.RegionPoint && !r.Plain {
			return r
		}
		if p := firstPoint(r.Children); p != nil {
			return p
		}
	}
	return nil
}

// TestDriftBaseline_UnchangedWorkspace：老工作区一个字节都不改，不应检出结构事务。
func TestDriftBaseline_UnchangedWorkspace(t *testing.T) {
	base, fenced, _ := driftFixture(t)
	parsed, err := fence.Parse(base, fenced)
	if err != nil {
		t.Fatal(err)
	}
	if edits := DetectStructural(base, parsed); len(edits) != 0 {
		t.Fatalf("未修改的工作区检出 %d 个结构事务，应为 0：%+v", len(edits), edits)
	}
	if rep := ValidateStructural(base, parsed, &tapfile.Doc{}); rep.HasError() {
		t.Fatalf("未修改的工作区报了 %d 条 error：%+v", rep.Errors, rep.Findings)
	}
}

// TestDriftBaseline_OldWorkspaceDescEdit：老工作区里只改 desc（与签名无关的结构事务），
// 不应因为基线签名行坐标漂移而报假的 V1。
func TestDriftBaseline_OldWorkspaceDescEdit(t *testing.T) {
	base, fenced, point := driftFixture(t)

	oldLine := fence.BeginFenceLine(point)
	newLine := strings.TrimSuffix(oldLine, "]}") + ` desc="probe desc"]}`
	edited := bytes.Replace(fenced, []byte(oldLine), []byte(newLine), 1)
	if bytes.Equal(edited, fenced) {
		t.Fatalf("围栏行替换失败，原行=%q", oldLine)
	}
	parsed, err := fence.Parse(base, edited)
	if err != nil {
		t.Fatal(err)
	}
	var descChanged bool
	for _, e := range DetectStructural(base, parsed) {
		if e.Kind == StructDescChange {
			descChanged = true
		}
	}
	if !descChanged {
		t.Fatal("夹具没有触发 desc 结构事务，这条测试会白过")
	}

	rep := ValidateStructural(base, parsed, &tapfile.Doc{})
	for _, f := range rep.Findings {
		if f.Code == "V1" {
			t.Errorf("只改 desc 却报了假 V1（%s）：%s", f.Region, f.Message)
		}
	}
}

// TestDriftBaseline_RealSignatureChange：老工作区里真改了签名行，必须仍然检出，
// 且基线侧取到的要是**真的**签名行（不是漂移坐标读出来的乱码）。
func TestDriftBaseline_RealSignatureChange(t *testing.T) {
	base, fenced, point := driftFixture(t)

	kind, ok := fgl.EnvelopeKindFor(point.Name)
	if !ok {
		t.Fatalf("%s 不是信封点", point.Name)
	}
	content := fenced[point.ContentSpan.Start:point.ContentSpan.End]
	blk, perr := fgl.ParseBlock(string(content), kind)
	if perr != nil || blk == nil {
		t.Fatalf("%s 正文不是良构信封：%v", point.Name, perr)
	}
	sig := string(content[blk.Header.StartByte:blk.Header.EndByte])
	name := bareName(point.Name)
	newSig := strings.Replace(sig, name, flipLastByte(name), 1)
	if newSig == sig {
		t.Fatalf("没构造出签名改动：sig=%q name=%q", sig, name)
	}
	edited := bytes.Replace(fenced, []byte(sig), []byte(newSig), 1)
	parsed, err := fence.Parse(base, edited)
	if err != nil {
		t.Fatal(err)
	}

	var sigEdits []StructEdit
	for _, e := range DetectStructural(base, parsed) {
		if e.Kind == StructSignatureChange {
			sigEdits = append(sigEdits, e)
		}
	}
	if len(sigEdits) != 1 {
		t.Fatalf("真改一条签名应恰好检出 1 条，实得 %d：%+v", len(sigEdits), sigEdits)
	}
	if sigEdits[0].Region != point.Name {
		t.Errorf("检出的点不对：期望 %s，实得 %s", point.Name, sigEdits[0].Region)
	}
	if sigEdits[0].From != sig {
		t.Errorf("基线侧签名行取错：\n  期望 %q\n  实得 %q", sig, sigEdits[0].From)
	}
	if !strings.Contains(sigEdits[0].To, flipLastByte(name)) {
		t.Errorf("编辑后签名行不对：%q", sigEdits[0].To)
	}
}

// flipLastByte 改掉标识符最后一个字节，得到等长的新名字。
func flipLastByte(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	switch last := b[len(b)-1]; {
	case last == 'z':
		b[len(b)-1] = 'y'
	case last >= 'a' && last < 'z', last >= 'A' && last < 'Z':
		b[len(b)-1] = last + 1
	default:
		return s + "x"
	}
	return string(b)
}
