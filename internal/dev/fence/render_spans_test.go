package fence

import (
	"testing"

	"tt/internal/dev/fgl"
	"tt/internal/dev/model"
)

// spanTextIn 取区间原文（越界或空区间给 ""）。
func spanTextIn(text []byte, sp *model.ByteRange) string {
	if sp == nil || sp.End <= sp.Start || sp.End > len(text) {
		return ""
	}
	return string(text[sp.Start:sp.End])
}

func flattenRegions(rs []*model.Region) []*model.Region {
	var out []*model.Region
	for _, r := range rs {
		out = append(out, r)
		out = append(out, flattenRegions(r.Children)...)
	}
	return out
}

// TestRenderSpansCoverSignatureAndDesc 钉住渲染后的区间记账：插入 begin 围栏行后，
// 签名行与描述块必须跟着正文一起平移。
//
// 漏平移时 Render 会把「未围栏坐标系」的偏移写进 .tdev/regions.json，读基线时
// 读到的不是签名行（实测 2.5MB 处差 316KB），gate2 于是把每个点都判成「签名被改」，
// 行业版 _wf 包的 70 个 cite_std 点就是这样全军覆没的。
func TestRenderSpansCoverSignatureAndDesc(t *testing.T) {
	_, fenced, regions, _ := synthDoc(t, "adzi999")
	checked := 0
	for _, r := range flattenRegions(regions) {
		if r.Kind != model.RegionPoint || r.Plain || r.SignatureSpan == nil {
			continue
		}
		kind, ok := fgl.EnvelopeKindFor(r.Name)
		if !ok {
			continue
		}
		content := fenced[r.ContentSpan.Start:r.ContentSpan.End]
		blk, err := fgl.ParseBlock(string(content), kind)
		if err != nil || blk == nil {
			t.Fatalf("%s 点到正文不是良构信封：%v", r.Name, err)
		}
		wantSig := string(content[blk.Header.StartByte:blk.Header.EndByte])
		if got := spanTextIn(fenced, r.SignatureSpan); got != wantSig {
			t.Errorf("%s 的 SignatureSpan 没落在签名行上：\n  期望 %q\n  实得 %q",
				r.Name, wantSig, got)
		}
		wantDesc := string(content[:blk.Header.StartByte])
		if got := spanTextIn(fenced, r.DescSpan); got != wantDesc {
			t.Errorf("%s 的 DescSpan 偏了：\n  期望 %q\n  实得 %q", r.Name, wantDesc, got)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("合成包里没有自订定义点，这条测试没有检查到任何东西")
	}
	t.Logf("检查 %d 个自订定义点的签名行/描述块落点", checked)
}
