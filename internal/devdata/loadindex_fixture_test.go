package devdata

// 索引复用（LoadIndex / IndexFresh / ReadLinkDoc）的磁盘材料测试。
//
// 为什么值得单独测：交互式问答里每条查询都先取索引，而索引是 800 KB、重建要解析
// 几 MB 的 .tap/.4gl。复用判定错了有两种坏方向 —— 该重建时复用（行号悄悄错）与
// 不该重建时重建（每次 5 秒）。判据只有一条：manifest 登记的三份源文件 sha256。

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

// TestFixtureLoadIndex_ReusesWhenSourcesUnchanged 钉住快路径：源文件 sha 未变时
// LoadIndex 直接返回**同一份**索引（GeneratedAt 不变），IndexFresh=true。
func TestFixtureLoadIndex_ReusesWhenSourcesUnchanged(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")
	first := set.mapIndex()

	// 索引刚建好：IndexFresh 必须为真（否则每次查询都会重建）。
	fresh, err := IndexFresh(set.LinkRequest())
	if err != nil {
		t.Fatalf("IndexFresh: %v", err)
	}
	if !fresh {
		t.Fatal("刚建完索引就报不新鲜（sha 比对口径错了）")
	}

	second := set.loadIndex()
	if second.GeneratedAt != first.GeneratedAt {
		t.Errorf("源未变却重建了索引: %q → %q", first.GeneratedAt, second.GeneratedAt)
	}
	if second.Sources.G4.SHA256 != first.Sources.G4.SHA256 {
		t.Errorf("复用路径不该改来源指纹")
	}
}

// TestFixtureLoadIndex_RebuildsWhenSourceChanged 钉住慢路径：源文件改了（这里改 .4gl）
// → IndexFresh=false 且 LoadIndex 重建出新指纹/新行数。
func TestFixtureLoadIndex_RebuildsWhenSourceChanged(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")
	first := set.mapIndex()

	set.writeFile("src/axmt500_wf.4gl",
		set.files["src/axmt500_wf.4gl"]+"\nDISPLAY '追加一段';\n")

	fresh, err := IndexFresh(set.LinkRequest())
	if err != nil {
		t.Fatalf("IndexFresh: %v", err)
	}
	if fresh {
		t.Fatal("源文件改了，IndexFresh 仍为 true（会复用旧索引、行号悄悄错）")
	}
	second := set.loadIndex()
	if second.Sources.G4.SHA256 == first.Sources.G4.SHA256 {
		t.Error("重建后 G4 指纹没变")
	}
	if second.Stats.G4Lines <= first.Stats.G4Lines {
		t.Errorf("行数没跟上: %d → %d", first.Stats.G4Lines, second.Stats.G4Lines)
	}
	// 重建后应恢复"新鲜"。
	if fresh, _ := IndexFresh(set.LinkRequest()); !fresh {
		t.Error("重建后 IndexFresh 应为 true")
	}
}

// TestFixtureLoadIndex_CorruptIndexRebuilds 钉住"索引损坏"：link.json 不是 JSON、
// 或 schema 是旧版本，都必须自动重建，不能让查询直接失败。
func TestFixtureLoadIndex_CorruptIndexRebuilds(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")
	set.mapIndex()

	// 情况 1：不是合法 JSON。
	if err := os.WriteFile(set.linkJSONPath(), []byte("{ 这不是 JSON"), 0o644); err != nil {
		t.Fatal(err)
	}
	if fresh, err := IndexFresh(set.LinkRequest()); err != nil || fresh {
		t.Errorf("损坏索引 IndexFresh = %v/%v，想要 false/nil", fresh, err)
	}
	doc := set.loadIndex()
	if doc.Schema != LinkSchema || doc.Stats.Points != 1 {
		t.Errorf("损坏索引没重建出正确结果: %+v", doc)
	}
	back := set.readLinkJSON()
	if back.Schema != LinkSchema {
		t.Errorf("重建后的 link.json 仍不合 schema: %q", back.Schema)
	}

	// 情况 2：合法 JSON 但 schema 是旧契约（v3）。
	old, err := json.Marshal(map[string]any{"schema": "tt.gen.link/v3", "prog": "axmt500_wf"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(set.linkJSONPath(), old, 0o644); err != nil {
		t.Fatal(err)
	}
	if fresh, _ := IndexFresh(set.LinkRequest()); fresh {
		t.Error("旧 schema 的索引不该判新鲜")
	}
	if got := set.loadIndex(); got.Schema != LinkSchema {
		t.Errorf("旧 schema 索引没升版: %q", got.Schema)
	}
}

// TestFixtureMap_RebuildOverwritesIndex（对应 CLI 的 --rebuild）：Map 不读旧索引，
// 每次都重写 link.json。这里把旧文件内容改掉后重建，必须被覆盖。
func TestFixtureMap_RebuildOverwritesIndex(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")
	set.mapIndex()
	if err := os.WriteFile(set.linkJSONPath(), []byte(`{"schema":"tt.gen.link/v4","prog":"旧内容"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	doc := set.mapIndex()
	if doc.Prog != "axmt500_wf" {
		t.Errorf("重建后的 prog = %q", doc.Prog)
	}
	back := set.readLinkJSON()
	if back.Prog != "axmt500_wf" || len(back.Points) != 1 {
		t.Errorf("--rebuild 没覆盖旧索引: %+v", back)
	}
}

// TestReadLinkDoc_Compatibility 钉住只读入口：不存在返回 (nil,nil)（status 据此说
// "未建索引"而不是报错）、损坏报错、合法则带上 Path。
func TestReadLinkDoc_Compatibility(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")

	if d, err := ReadLinkDoc(set.linkJSONPath()); d != nil || err != nil {
		t.Errorf("不存在时应 (nil,nil)，得到 %v/%v", d, err)
	}

	if err := os.WriteFile(set.linkJSONPath(), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadLinkDoc(set.linkJSONPath()); err == nil || !strings.Contains(err.Error(), "合法 JSON") {
		t.Errorf("损坏索引应明确报错，得到 %v", err)
	}

	set.mapIndex()
	d, err := ReadLinkDoc(set.linkJSONPath())
	if err != nil || d == nil {
		t.Fatalf("ReadLinkDoc: %v / %v", d, err)
	}
	if d.Path != set.linkJSONPath() || d.Schema != LinkSchema {
		t.Errorf("只读入口 = %+v", d)
	}
	if len(d.Points) != 1 || d.Points[0].Name != "input.a.page1.f1" {
		t.Errorf("只读入口读出的点位 = %+v", d.Points)
	}
}

// TestFixtureIndexFresh_DetectsTsdAndTabChange 钉住"画面侧材料也算源"：只改 .tsd 的
// 字段标签也足以让索引过期（否则标签会一直缺）。这是 v4 schema 引入 Sources.Tsd 的原因。
func TestFixtureIndexFresh_DetectsTsdAndTabChange(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog:      "axmt500_wf",
		TSDLabels: map[string]string{"f1": "料号"},
	})
	set.mapIndex()
	if fresh, _ := IndexFresh(set.LinkRequest()); !fresh {
		t.Fatal("初始索引应新鲜")
	}

	// 只改 .tsd（标签文案）→ 索引必须过期；重建后字段标签要更新。
	set.writeFile("dzx/axmt500_wf.tsd", `<spec><strings><sfield name="lbl_f1" text="新料号名" /></strings></spec>`)
	if fresh, _ := IndexFresh(set.LinkRequest()); fresh {
		t.Error("改了 .tsd，索引却还判新鲜")
	}
	doc := set.loadIndex()
	if len(doc.Fields) != 1 || doc.Fields[0].Label != "新料号名" {
		t.Errorf("重建后字段标签 = %+v", doc.Fields)
	}

	// 删掉 .tsd（材料本来就没有）：sha 从有到无，同样必须过期。
	set.removeFile("dzx/axmt500_wf.tsd")
	if fresh, _ := IndexFresh(set.LinkRequest()); fresh {
		t.Error("删了 .tsd，索引却还判新鲜")
	}
	if doc := set.loadIndex(); doc.Sources.Tsd != nil {
		t.Errorf("没有 .tsd 却记了来源: %+v", doc.Sources.Tsd)
	}
}

// TestLoadIndex_Missing4glIsMaterialIncomplete 钉住错误路径：重建时缺 .4gl
// 必须报 ErrMaterialIncomplete（"材料缺件"），而不是把索引写成空壳。
func TestLoadIndex_Missing4glIsMaterialIncomplete(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")
	set.mapIndex()
	set.removeFile("src/axmt500_wf.4gl")
	// manifest 里也摘掉，让 sourceSHAs 认为没有 4gl，从而走重建。
	set.removeFile("dzx/axmt500_wf.tap") // 顺便让 fast path 直接失效
	_, err := LoadIndex(set.LinkRequest())
	if !errors.Is(err, ErrMaterialIncomplete) {
		t.Fatalf("err = %v，想要 ErrMaterialIncomplete", err)
	}
	if !strings.Contains(err.Error(), "缺 .4gl") {
		t.Errorf("错误文案没点出缺件: %v", err)
	}
}
