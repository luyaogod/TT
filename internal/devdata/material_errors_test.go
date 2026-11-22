package devdata

// 材料读写的错误路径与"清单契约"测试。
//
// 这一组回答的是：材料缺件/清单损坏/环境不存在时，gen 给的是**能照做**的错误，
// 还是把空索引/空材料当成正常结果。真机上最容易出错的就是这里 —— 标准程序本来
// 就没有 .tap，若把它当"程序不存在"，用户会以为程序编号打错。

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFixtureMap_ProgramNotFound 钉住"没拉过材料"：错误里必须是 ErrProgNotFound
// 且给出下一步命令，不能是一句 file does not exist。
func TestFixtureMap_ProgramNotFound(t *testing.T) {
	cfg := writeConfig(t, `{"name":"e","host":"127.0.0.1","port":22,"user":"u"}`)
	lr := LinkRequest{CfgPath: cfg, Env: "e", Prog: "axmt500_wf"}

	if _, err := Map(lr); !errors.Is(err, ErrProgNotFound) {
		t.Fatalf("Map: err = %v，想要 ErrProgNotFound", err)
	} else if !strings.Contains(err.Error(), "tt gen fetch") {
		t.Errorf("错误没给出下一步: %v", err)
	}
	if _, err := LoadIndex(lr); !errors.Is(err, ErrProgNotFound) {
		t.Errorf("LoadIndex: err = %v", err)
	}
	if _, err := Show(ShowRequest{LinkRequest: lr, Query: "f1"}); !errors.Is(err, ErrProgNotFound) {
		t.Errorf("Show: err = %v", err)
	}
}

// TestFixtureMap_ManifestCorrupt 钉住清单本身坏掉的两种情况：
// kind 不是 fetch（拿错材料集）与 JSON 坏了 —— 都要报 ErrProgNotFound/明确错误。
func TestFixtureMap_ManifestCorrupt(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")
	if err := os.WriteFile(filepath.Join(set.Dir, ManifestFileName),
		[]byte(`{"schema":"tt.devdata/v1","kind":"mirror","env":"e"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Map(set.LinkRequest()); !errors.Is(err, ErrProgNotFound) {
		t.Errorf("kind=mirror: err = %v，想要 ErrProgNotFound", err)
	} else if !strings.Contains(err.Error(), "按程序拉取") {
		t.Errorf("错误没说清是材料种类不对: %v", err)
	}

	if err := os.WriteFile(filepath.Join(set.Dir, ManifestFileName), []byte(`{坏 JSON`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Map(set.LinkRequest())
	if err == nil || !strings.Contains(err.Error(), "不是合法 JSON") {
		t.Errorf("坏 JSON: err = %v", err)
	}
}

// TestFixtureMap_Missing4glIsIncomplete 钉住"程序在、材料缺件"：
// 缺 .4gl 时 Map 报 ErrMaterialIncomplete，文案要说明是缺件而不是没这个程序。
func TestFixtureMap_Missing4glIsIncomplete(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")
	set.removeFile("src/axmt500_wf.4gl")
	_, err := Map(set.LinkRequest())
	if !errors.Is(err, ErrMaterialIncomplete) {
		t.Fatalf("err = %v，想要 ErrMaterialIncomplete", err)
	}
	if !strings.Contains(err.Error(), "缺 .4gl") {
		t.Errorf("文案没点出缺什么: %v", err)
	}
	// 缺 .tap 不算缺件（标准程序），Map 必须能正常建索引。
	notap := newMaterials(t, matOpts{Prog: "axmt200", NoTap: true})
	if doc := notap.mapIndex(); doc.Stats.Points == 0 && len(doc.Sections) == 0 {
		t.Errorf("缺 .tap 不该让 Map 失败: %+v", doc)
	}
}

// TestFixtureTapView_OrphanBodyAndErrors 钉住 .tap 正文视图：
// 声明正文来自 .tap（行号也是 .tap 行号）、HasSlot=false 表示重产不保留；
// 材料里没有 .tap 时要报 ErrMaterialIncomplete（不是"没有这个点"）。
func TestFixtureTapView_OrphanBodyAndErrors(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt500_wf",
		Orphans: []matOrphan{
			{Name: "input.a.page73.gone", Src: "c", Ver: "2", Body: "MESSAGE '孤兒內容'\nLET x = 1\n"},
		},
	})
	lr := set.LinkRequest()

	v, err := TapPoint(lr, "INPUT.A.PAGE73.GONE") // 大小写不敏感
	if err != nil || v == nil {
		t.Fatalf("TapPoint: %v / %v", v, err)
	}
	if v.HasSlot {
		t.Error("孤儿点不该有槽位")
	}
	if len(v.Content) != 2 || !strings.Contains(v.Content[0], "孤兒內容") {
		t.Errorf("正文 = %#v（应来自 .tap）", v.Content)
	}
	if v.TagLine == 0 || v.Line0 == 0 || v.Line1 != v.Line0+1 {
		t.Errorf("行号 = tag:%d %d-%d", v.TagLine, v.Line0, v.Line1)
	}
	if v.Src != "c" || v.Ver != "2" {
		t.Errorf("声明属性 = %+v", v)
	}

	if v, err := TapPoint(lr, "input.a.page1.f1"); err != nil || v == nil || !v.HasSlot {
		t.Errorf("有槽位的点 HasSlot 应为 true: %+v / %v", v, err)
	}
	if v, err := TapPoint(lr, "no.such.point"); err != nil || v != nil {
		t.Errorf("找不到的点应 (nil,nil): %+v / %v", v, err)
	}

	// 没有 .tap：读不了是"材料不成套"，不是"没这个点"。
	notap := newMaterials(t, matOpts{Prog: "axmt200", NoTap: true})
	if _, err := TapPoint(notap.LinkRequest(), "x.y"); !errors.Is(err, ErrMaterialIncomplete) {
		t.Errorf("没有 .tap 时 err = %v，想要 ErrMaterialIncomplete", err)
	}
	if _, err := TapSearch(notap.LinkRequest(), "x", 10, false); !errors.Is(err, ErrMaterialIncomplete) {
		t.Errorf("TapSearch: err = %v", err)
	}
}

// TestFixtureTapSearch_AttributionAndLimit 钉住 .tap 检索：命中要标明属于哪个声明、
// 该声明有没有槽位（"代码在、跑起来没有"的唯一判据），limit 只截显示不截总数。
func TestFixtureTapSearch_AttributionAndLimit(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog: "axmt500_wf",
		Orphans: []matOrphan{
			{Name: "input.a.page73.gone", Src: "c", Ver: "2", Body: "MESSAGE '孤兒'\nMESSAGE '第二行'\nMESSAGE '第三行'\n"},
		},
	})
	res, err := TapSearchBytes("axmt500_wf", "materials",
		[]byte(set.files["dzx/axmt500_wf.tap"]), "MESSAGE", 2, false)
	if err != nil {
		t.Fatalf("TapSearchBytes: %v", err)
	}
	if res.Total != 3 || res.Shown != 2 || len(res.Hits) != 2 {
		t.Fatalf("tap 命中 = total:%d shown:%d %+v", res.Total, res.Shown, res.Hits)
	}
	if h := res.Hits[0]; h.Point != "input.a.page73.gone" || h.HasSlot || h.Src != "c" {
		t.Errorf("命中归属 = %+v", h)
	}
	if !strings.Contains(res.Hits[0].Text, "孤兒") {
		t.Errorf("命中文本 = %q", res.Hits[0].Text)
	}

	// 区分大小写与"只命中声明行"的边界。
	if r, _ := TapSearchBytes("axmt500_wf", "materials",
		[]byte(set.files["dzx/axmt500_wf.tap"]), "message", 50, true); r.Total != 0 {
		t.Errorf("区分大小写时大写 MESSAGE 不该被 message 命中: %d", r.Total)
	}
}

// TestFixtureManifest_RoundTripAndFingerprint 钉住 manifest 是材料的"身份证"：
// 每个落盘文件都有 sha256/size/srcMtime，读写往返不丢字段，Add* 去重。
func TestFixtureManifest_RoundTripAndFingerprint(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog:             "axmt500_wf",
		ManifestWarnings: []string{"蓝图比骨架新"},
		ManifestMissing:  []string{"erp/cxm/dzx/tap/axmt500_wf.tap"},
		Slices:           []matSlice{{Code: "a05", Used: 2, Current: 3}},
	})
	mf := set.readManifest()

	// 角色齐全：D 侧的七类都在（tab/tsd/4fd/tap/tgl/tgx/tgx2/4gl/inc）。
	roles := map[string]bool{}
	for _, f := range mf.Files {
		roles[f.Role] = true
		if f.Rel == "" || f.SHA256 == "" || f.Size == 0 || f.SrcMtime == "" {
			t.Errorf("清单条目缺指纹/时间: %+v", f)
		}
		// 落盘的 sha 必须与磁盘内容一致（否则索引复用判定从一开始就是假的）。
		if got, err := FileSHA256(set.filePath(f.Rel)); err != nil || got != f.SHA256 {
			t.Errorf("文件 %s 指纹不符: %q/%v", f.Rel, got, err)
		}
	}
	for _, r := range []string{"tab", "tsd", "4fd", "tap", "tgl", "tgx", "tgx2", "4gl", "inc"} {
		if !roles[r] {
			t.Errorf("清单缺角色 %s: %+v", r, roles)
		}
	}
	if mf.Kind != "fetch" || mf.Schema != ManifestSchema || mf.Prog != "axmt500_wf" {
		t.Errorf("清单身份 = %+v", mf)
	}
	if mf.Bytes() <= 0 || len(mf.Missing) != 1 || len(mf.Warnings) != 1 {
		t.Errorf("清单统计 = %d / missing:%v warnings:%v", mf.Bytes(), mf.Missing, mf.Warnings)
	}
	if got := mf.Versions.TemplateVersions["a05"]; got != 3 {
		t.Errorf("模板版本锚点 = %v", mf.Versions.TemplateVersions)
	}
	if mf.FileByRel("dzx/axmt500_wf.tap") == nil || mf.FileByRel("nope") != nil {
		t.Errorf("FileByRel 口径错")
	}

	// Add* 去重。
	mf.AddMissing("erp/cxm/dzx/tap/axmt500_wf.tap")
	mf.AddWarning("蓝图比骨架新")
	mf.AddWarning("   ")
	if len(mf.Missing) != 1 || len(mf.Warnings) != 1 {
		t.Errorf("Add* 没去重: %v / %v", mf.Missing, mf.Warnings)
	}
	mf.AddFile(FileEntry{Role: "4gl", Rel: "x.4gl", Size: 3})
	if len(mf.Files) == 0 {
		t.Error("AddFile 没生效")
	}

	// 读写往返：写到另一个目录再读回，字段不丢。
	dir := t.TempDir()
	if err := WriteManifest(dir, mf); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}
	back, err := ReadManifest(dir)
	if err != nil || back == nil {
		t.Fatalf("ReadManifest: %v / %v", back, err)
	}
	if back.Prog != mf.Prog || len(back.Files) != len(mf.Files) || back.Schema != ManifestSchema {
		t.Errorf("往返后不一致: %+v", back)
	}
	if back, err := ReadManifest(t.TempDir()); back != nil || err != nil {
		t.Errorf("没有清单时应 (nil,nil): %+v / %v", back, err)
	}

	// 旧契约版本：能读出来，但必须加一条"建议重新拉取"的警告（不静默按新规则解释）。
	oldDir := t.TempDir()
	old := `{"schema":"tt.devdata/v0","kind":"fetch","env":"e","files":[],"missing":[],"warnings":[]}`
	if err := os.WriteFile(filepath.Join(oldDir, ManifestFileName), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	oldMF, err := ReadManifest(oldDir)
	if err != nil || oldMF == nil {
		t.Fatalf("旧契约清单应能读: %v / %v", oldMF, err)
	}
	if len(oldMF.Warnings) == 0 || !strings.Contains(oldMF.Warnings[0], "契约版本") {
		t.Errorf("旧契约没给警告: %v", oldMF.Warnings)
	}
}

// TestFixtureManifest_StaleAndConsistency 钉住清单的一致性判据：
//   - LocalStale：本地文件被改/被删要报（status 不出网就能看见）；
//   - ConsistencyWarnings：.tab 比 .tgl 新 → 蓝图改过但没重产；缺 tab/tgl 必须明说；
//   - RemoteStale：服务器侧消失/变大/mtime 变新/模板版本变化。
func TestFixtureManifest_StaleAndConsistency(t *testing.T) {
	set := newMaterials(t, matOpts{
		Prog:     "axmt500_wf",
		TabMTime: "2026-11-21T10:20:00+08:00",
		TglMTime: "2026-11-21T10:00:00+08:00", // tab 比 tgl 新 20 分钟
		Slices:   []matSlice{{Code: "a05", Used: 2, Current: 3}},
	})
	mf := set.readManifest()

	// LocalStale：干净时应为空；改一个文件内容（大小不变也算 sha-mismatch）后必须报。
	if got := LocalStale(set.Dir, mf); len(got) != 0 {
		t.Fatalf("干净材料不该 stale: %+v", got)
	}
	rel := "dzx/axmt500_wf.tgl"
	raw, err := os.ReadFile(set.filePath(rel))
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(raw), "測試作業", "測試作業改", 1)
	if err := os.WriteFile(set.filePath(rel), []byte(changed), 0o644); err != nil {
		t.Fatal(err)
	}
	stale := LocalStale(set.Dir, mf)
	if len(stale) != 1 || stale[0].Code != "sha-mismatch" || !strings.Contains(stale[0].Detail, rel) {
		t.Errorf("改内容后 stale = %+v", stale)
	}
	// 删文件 → missing-local。
	if err := os.Remove(set.filePath(rel)); err != nil {
		t.Fatal(err)
	}
	if got := LocalStale(set.Dir, mf); len(got) != 1 || got[0].Code != "missing-local" {
		t.Errorf("删文件后 stale = %+v", got)
	}
	if got := LocalStale(set.Dir, nil); len(got) != 1 || got[0].Code != "missing-local" {
		t.Errorf("没有 manifest 时 stale = %+v", got)
	}

	// ConsistencyWarnings：tab 比 tgl 新 → 一条"产物可能滞后"。
	// 用原始 mtime 重建一份干净 manifest，避免被上面的改动影响。
	set2 := newMaterials(t, matOpts{
		Prog:     "axmt500_wf",
		TabMTime: "2026-11-21T10:20:00+08:00",
		TglMTime: "2026-11-21T10:00:00+08:00",
	})
	warns := ConsistencyWarnings(set2.readManifest())
	found := false
	for _, w := range warns {
		if strings.Contains(w, "蓝图改过但骨架未重产") {
			found = true
		}
	}
	if !found {
		t.Errorf("tab 比 tgl 新时没报一致性警告: %v", warns)
	}
	// 缺 tab / 缺 tgl 各自必须有一条。
	bare := NewManifest("fetch", "e")
	if warns := ConsistencyWarnings(bare); len(warns) != 2 {
		t.Errorf("缺 tab+tgl 时警告 = %v", warns)
	}
	if warns := ConsistencyWarnings(nil); warns != nil {
		t.Errorf("nil manifest 应返回 nil: %v", warns)
	}

	// RemoteStale 的四种触发。
	remote := map[string]RemoteStamp{
		"gone":    {Src: "gone", Exists: false},
		"bigger":  {Src: "bigger", Exists: true, Size: mf.Files[0].Size + 10},
		"newer":   {Src: "newer", Exists: true, Size: mf.Files[0].Size, ModTime: matMustTime(t, "2026-11-22T00:00:00+08:00")},
		"unknown": {Src: "not-in-manifest", Exists: true},
	}
	// 让清单里有对应 src 的条目。
	probe := NewManifest("fetch", "e")
	probe.Files = []FileEntry{
		{Src: "gone", Rel: "a", Size: 1, SHA256: "x"},
		{Src: "bigger", Rel: "b", Size: 1, SHA256: "x"},
		{Src: "newer", Rel: "c", Size: 1, SHA256: "x", SrcMtime: "2026-11-21T00:00:00+08:00"},
	}
	// 模板版本锚点：清单记 a05=3、现在 4 → version-changed。
	probe.Versions.TemplateVersions["a05"] = 3
	reasons := RemoteStale(probe, remote, map[string]int{"a05": 4}, nil)
	codes := map[string]bool{}
	for _, r := range reasons {
		codes[r.Code] = true
	}
	for _, want := range []string{"remote-gone", "remote-newer", "version-changed"} {
		if !codes[want] {
			t.Errorf("RemoteStale 缺 %s 触发: %+v", want, reasons)
		}
	}
	if got := RemoteStale(nil, remote, nil, nil); got != nil {
		t.Errorf("nil manifest 应返回 nil: %+v", got)
	}
}

// TestFixtureMdlAndTglParsers 钉住版本锚点解析：模板 key/版本、.tgl 头部、
// 以及"哪些文件算模板"的边界（带日期的备份算文件、不算版本锚点）。
func TestFixtureMdlAndTglParsers(t *testing.T) {
	// 构建器写出的 .tgl 必须能被 ParseTglHeader 解析出 SD 版次/Description。
	set := newMaterials(t, matOpts{
		Prog:   "axmt500_wf",
		Slices: []matSlice{{Code: "a05", Used: 2, Current: 3}},
	})
	hdr := ParseTglHeader([]byte(set.files["dzx/axmt500_wf.tgl"]))
	if hdr == nil || !strings.HasPrefix(hdr.SDVer, "0015") || hdr.Build != "000037" ||
		!strings.Contains(hdr.Desc, "測試作業") {
		t.Fatalf("tgl 头部 = %+v", hdr)
	}
	if ParseTglHeader([]byte("plain tgl body")) != nil {
		t.Error("没有版次头时应返回 nil")
	}

	// 模板版本：只看前 4 KB 的 `(Version:N)`。
	if v, ok := ParseTemplateVersion([]byte("#應用 a05 樣板自動產生(Version:3)\n")); !ok || v != 3 {
		t.Errorf("模板版本 = %d/%v", v, ok)
	}
	if _, ok := ParseTemplateVersion([]byte("没有版本标记")); ok {
		t.Error("没有标记时不该报 ok")
	}
	if v, ok := ParseTemplateVersion([]byte(strings.Repeat("x", 5000) + "(Version:7)")); ok || v != 0 {
		t.Errorf("4 KB 之后的版本标记不该被扫到: %d/%v", v, ok)
	}

	// TemplateKey / IsTemplateFile 的边界：备份不算锚点、slice/ 下算模板文件。
	for _, tc := range []struct {
		rel, key string
		ok       bool
	}{
		{"slice/code_a05.template", "a05", true},
		{"code_q01.template", "q01", true},
		{"code_g01.template.141119", "", false},
		{"4rp_t01.template", "", false},
		{"code_.template", "", false},
	} {
		key, ok := TemplateKey(tc.rel)
		if key != tc.key || ok != tc.ok {
			t.Errorf("TemplateKey(%q) = %q/%v，想要 %q/%v", tc.rel, key, ok, tc.key, tc.ok)
		}
	}
	if !IsTemplateFile("code_a05.template") || !IsTemplateFile("erp/mdl/slice/x.template") || IsTemplateFile("erp/mdl/4rp_t01.template") {
		t.Error("IsTemplateFile 口径错")
	}

	// 模板树里真的写下了当前版本（否则脱版检测无从谈起）。
	tpl, err := os.ReadFile(filepath.Join(set.Mdl, "slice", "code_a05.template"))
	if err != nil {
		t.Fatalf("模板树里没有 a05: %v", err)
	}
	if v, ok := ParseTemplateVersion(tpl); !ok || v == 0 {
		t.Errorf("模板版本 = %d/%v", v, ok)
	}
}

// TestFixturePaths_RolesAndClean 钉住路径规则与清理：
// 路径只由 config 拼（不许调用方自己 Join）、清理默认给计划而不是直接删。
func TestFixturePaths_RolesAndClean(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")

	// 路径：EnvDir/ProgDir/DictPath/MdlDir/ProgRoot 都走同一套配置。
	if d, err := EnvDir(set.CfgPath, "e"); err != nil || !strings.HasSuffix(d, filepath.Join("devdata", "e")) {
		t.Errorf("EnvDir = %q/%v", d, err)
	}
	if d, err := ProgRoot(set.CfgPath, "e"); err != nil || !strings.HasSuffix(d, filepath.Join("e", "prog")) {
		t.Errorf("ProgRoot = %q/%v", d, err)
	}
	if d, err := MdlDir(set.CfgPath, "e"); err != nil || !strings.HasSuffix(d, filepath.Join("e", "mdl")) {
		t.Errorf("MdlDir = %q/%v", d, err)
	}
	if d, err := DictPath(set.CfgPath, "e"); err != nil || !strings.HasSuffix(d, ".db") {
		t.Errorf("DictPath = %q/%v", d, err)
	}
	// 环境名/程序名的错误要能区分"非法"与"不存在"。
	if _, err := EnvDir(set.CfgPath, "不存在的环境"); !errors.Is(err, ErrInvalidEnv) && !errors.Is(err, ErrEnvNotFound) {
		t.Errorf("未知环境 err = %v", err)
	}
	if _, err := EnvDir(set.CfgPath, ""); !errors.Is(err, ErrInvalidEnv) {
		t.Errorf("空环境 err = %v，想要 ErrInvalidEnv", err)
	}
	if _, err := ProgDir(set.CfgPath, "e", "../etc"); !errors.Is(err, ErrInvalidProg) {
		t.Errorf("越界程序名 err = %v，想要 ErrInvalidProg", err)
	}
	if _, err := ProgDir(set.CfgPath, "e", "ab"); !errors.Is(err, ErrInvalidProg) {
		t.Errorf("太短程序名 err = %v", err)
	}
	if _, err := DictPath(set.CfgPath, "nope"); !errors.Is(err, ErrEnvNotFound) {
		t.Errorf("不存在的环境 err = %v，想要 ErrEnvNotFound", err)
	}

	// 清理：默认是计划（不删），Execute 才删；删程序目录后 Plan 报 ErrProgNotFound。
	plan, err := PlanClean(set.CfgPath, "e", "axmt500_wf")
	if err != nil || plan.Empty() || plan.Items[0].Kind != "prog" || plan.Files == 0 {
		t.Fatalf("PlanClean = %+v / %v", plan, err)
	}
	if _, err := os.Stat(set.Dir); err != nil {
		t.Error("PlanClean 不该真删")
	}
	removed, err := RemoveProg(set.CfgPath, "e", "axmt500_wf")
	if err != nil || removed.Empty() {
		t.Fatalf("RemoveProg = %+v / %v", removed, err)
	}
	if _, err := os.Stat(set.Dir); !os.IsNotExist(err) {
		t.Error("RemoveProg 没删掉程序目录")
	}
	if _, err := PlanClean(set.CfgPath, "e", "axmt500_wf"); !errors.Is(err, ErrProgNotFound) {
		t.Errorf("删完再 PlanClean err = %v，想要 ErrProgNotFound", err)
	}
	// ExecuteClean 对空计划是幂等的。
	if err := ExecuteClean(nil); err != nil {
		t.Errorf("ExecuteClean(nil) = %v", err)
	}
	if err := ExecuteClean(&CleanPlan{}); err != nil {
		t.Errorf("ExecuteClean(空) = %v", err)
	}

	// 临时目录纪律：NewTempDir 造在目标同父目录、CommitDir 原子改名、CleanupTemp 清残留。
	final := filepath.Join(t.TempDir(), "prog_commit")
	tmp, err := NewTempDir(final)
	if err != nil {
		t.Fatalf("NewTempDir: %v", err)
	}
	if filepath.Dir(tmp) != filepath.Dir(final) || !strings.Contains(filepath.Base(tmp), ".tmp-") {
		t.Errorf("临时目录形态 = %q", tmp)
	}
	nested := filepath.Join(tmp, "材料")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := CommitDir(tmp, final, false); err != nil {
		t.Fatalf("CommitDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(final, "材料")); err != nil {
		t.Error("CommitDir 后材料不在正式目录")
	}
	if err := CommitDir(t.TempDir(), final, false); !errors.Is(err, ErrExists) {
		t.Errorf("目标已存在 err = %v，想要 ErrExists", err)
	}
	if err := CommitDir(t.TempDir(), final, true); err != nil {
		t.Errorf("force 覆盖 err = %v", err)
	}
	residual := filepath.Join(filepath.Dir(final), filepath.Base(final)+".tmp-keep")
	if err := os.MkdirAll(residual, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := TempDirs(filepath.Dir(final), filepath.Base(final)); len(got) != 1 {
		t.Errorf("TempDirs = %v", got)
	}
	CleanupTemp("", final)
	if _, err := os.Stat(residual); !os.IsNotExist(err) {
		t.Error("CleanupTemp 没清掉历史残留")
	}

	// 角色表：表名清单与按名查找（大小写不敏感）。
	if names := MirrorTableNames(); len(names) != len(MirrorTables) || names[0] != "DZAA_T" {
		t.Errorf("镜像表名 = %d 条 / %q", len(names), names[0])
	}
	if t0 := TableByName("dzaa_t"); t0 == nil || t0.Name != "DZAA_T" {
		t.Errorf("TableByName 大小写不敏感失败: %+v", t0)
	}
	if TableByName("NOT_A_TABLE") != nil {
		t.Error("不存在的表应返回 nil")
	}
	if !equalFold("abc", "ABC") || equalFold("abc", "abd") || equalFold("ab", "abc") {
		t.Error("equalFold 口径错")
	}
	if got := (FetchRoles[0]).RelPath("p"); got != "dzx/p.tab" {
		t.Errorf("RelPath = %q", got)
	}
	if got := (FileRole{Src: "erp/%s/4gl/%s.4gl"}).SrcPath("cxm", "p"); got != "erp/cxm/4gl/p.4gl" {
		t.Errorf("SrcPath = %q", got)
	}
}
