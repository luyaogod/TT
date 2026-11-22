package devdata

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hostStub 一个只够路径解析用的环境（不真的连）。
const hostStub = `{"name":"测试环境","host":"10.0.0.1","user":"tiptop","zone":"2","password":"x"}`

// newMirrorFixture 造一套"服务器上有一份标准材料"的假来源。
func newMirrorFixture(t *testing.T) (cfgPath string, tables *fakeTable, files *fakeFiles) {
	t.Helper()
	cfgPath = writeConfig(t, hostStub)
	tables = newFakeTable()
	files = newFakeFiles()
	for _, spec := range MirrorTables {
		// 每张表一行一列（真实的 53 张表结构差异不重要，这里要验的是流程）
		tables.setTable(spec.Name, []string{strings.ToLower(spec.Name[:4]) + "001"}, []string{"v-" + spec.Name})
	}
	// 一张超上限的表 + 一张本环境不存在的表
	tables.countOverride["DZBA_T"] = MaxTableRows + 1
	tables.missing["DZAS_T"] = true
	// 模板树：一个 code_ 模板 + 一个 slice 切片 + 一个 snippet
	files.files["erp/mdl/code_i02.template"] = []byte("#應用 i02 樣板自動產生(Version:12)\n...")
	files.files["erp/mdl/slice/code_a05.template"] = []byte("#應用 a05 樣板自動產生(Version:3)\n...")
	files.files["erp/mdl/F001_00.snippet"] = []byte("snippet-bytes")
	return cfgPath, tables, files
}

// TestMirrorHappyPath 一次成功的共用资料同步：规格库落盘、模板树落盘、manifest 齐全。
func TestMirrorHappyPath(t *testing.T) {
	cfgPath, tables, files := newMirrorFixture(t)
	envDir, err := EnvDir(cfgPath, "测试环境")
	if err != nil {
		t.Fatal(err)
	}
	var phases []string
	res, err := Mirror(context.Background(), MirrorRequest{
		CfgPath: cfgPath, Env: "测试环境", TTVersion: "test",
		Tables: tables, Files: files,
	}, func(p Progress) { phases = append(phases, p.Phase) })
	if err != nil {
		t.Fatalf("Mirror: %v", err)
	}

	// 51 张落库（53 - DZBA 超限 - DZAS 缺表），2 张分别进 skipped / missing
	if res.TableCount != 51 {
		t.Errorf("落库表数 = %d，应为 51", res.TableCount)
	}
	if res.RowTotal != 51 {
		t.Errorf("总行数 = %d，应为 51（每表一行）", res.RowTotal)
	}
	var skipped []string
	for _, tb := range res.Tables {
		if tb.Skipped {
			skipped = append(skipped, tb.Table)
		}
	}
	if len(skipped) != 1 || skipped[0] != "DZBA_T" {
		t.Errorf("skipped 应为 [DZBA_T]，实际 %v", skipped)
	}
	if !containsStr(res.Missing, "DZAS_T") {
		t.Errorf("DZAS_T 应进 missing: %v", res.Missing)
	}
	if res.MdlFiles != 3 {
		t.Errorf("模板文件数 = %d，应为 3", res.MdlFiles)
	}
	// 模板版本：i02=12、a05=3（snippet 没有版本标记）
	if res.TemplateVersions["i02"] != 12 || res.TemplateVersions["a05"] != 3 {
		t.Errorf("模板版本 = %v", res.TemplateVersions)
	}

	// 落盘检查
	dbPath := filepath.Join(envDir, "dict", "测试环境.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("规格库没落盘: %v", err)
	}
	if _, err := os.Stat(filepath.Join(envDir, "mdl", "slice", "code_a05.template")); err != nil {
		t.Fatalf("模板树没落盘: %v", err)
	}
	// 原样保真：字节一致
	got, err := os.ReadFile(filepath.Join(envDir, "mdl", "F001_00.snippet"))
	if err != nil || string(got) != "snippet-bytes" {
		t.Fatalf("模板内容被改了: %q, %v", got, err)
	}
	// 无临时残留
	if left := TempDirs(filepath.Join(envDir, "dict"), "build"); len(left) != 0 {
		t.Errorf("dict 下还有临时目录: %v", left)
	}
	if left := TempDirs(envDir, "mdl-build"); len(left) != 0 {
		t.Errorf("env 下还有临时目录: %v", left)
	}

	// manifest
	m, err := ReadManifest(envDir)
	if err != nil || m == nil {
		t.Fatalf("manifest: %v", err)
	}
	if m.Kind != KindMirror || m.Env != "测试环境" || m.TTVersion != "test" {
		t.Errorf("manifest 基本字段: %+v", m)
	}
	if len(m.Tables) != 53 {
		t.Errorf("manifest.tables 应有 53 条，实际 %d", len(m.Tables))
	}
	if m.FileByRel(filepath.ToSlash(filepath.Join("dict", "测试环境.db"))) == nil {
		t.Error("manifest 里缺 dict 条目")
	}
	if countRole(m, "mdl") != 3 {
		t.Errorf("manifest 里 mdl 条目 = %d，应为 3", countRole(m, "mdl"))
	}
	if m.Source.TopDir != files.Top() {
		t.Errorf("source.topDir = %q，应为 %q", m.Source.TopDir, files.Top())
	}
	// 进度：至少经过 table 与 mdl 两个阶段
	seen := map[string]bool{}
	for _, p := range phases {
		seen[p] = true
	}
	for _, want := range []string{"table", "mdl", "index", "done"} {
		if !seen[want] {
			t.Errorf("进度里缺阶段 %q: %v", want, phases)
		}
	}
}

// TestMirrorIdempotent 第二次同步默认跳过（不重写、不报错、不再连库）。
func TestMirrorIdempotent(t *testing.T) {
	cfgPath, tables, files := newMirrorFixture(t)
	req := MirrorRequest{CfgPath: cfgPath, Env: "测试环境", Tables: tables, Files: files}
	if _, err := Mirror(context.Background(), req, nil); err != nil {
		t.Fatalf("第一次: %v", err)
	}
	queriedBefore := len(tables.queried)
	readBefore := len(files.readAll)

	res, err := Mirror(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("第二次: %v", err)
	}
	if !res.SkippedPull {
		t.Fatal("第二次应跳过（幂等）")
	}
	if res.TableCount == 0 || res.MdlFiles == 0 {
		t.Errorf("跳过时也要回报已有状态: %+v", res)
	}
	if len(tables.queried) != queriedBefore {
		t.Errorf("跳过时不该再查库（多了 %d 次）", len(tables.queried)-queriedBefore)
	}
	if len(files.readAll) != readBefore {
		t.Errorf("跳过时不该再读文件（多了 %d 次）", len(files.readAll)-readBefore)
	}

	// --force 必须重拉
	res, err = Mirror(context.Background(), MirrorRequest{
		CfgPath: cfgPath, Env: "测试环境", Force: true, Tables: tables, Files: files}, nil)
	if err != nil {
		t.Fatalf("force: %v", err)
	}
	if res.SkippedPull {
		t.Fatal("force 不该跳过")
	}
}

// TestMirrorOnlyKeepsOtherHalf --only dict 不能把 mdl 的记录抹掉。
func TestMirrorOnlyKeepsOtherHalf(t *testing.T) {
	cfgPath, tables, files := newMirrorFixture(t)
	req := MirrorRequest{CfgPath: cfgPath, Env: "测试环境", Tables: tables, Files: files}
	if _, err := Mirror(context.Background(), req, nil); err != nil {
		t.Fatalf("第一次: %v", err)
	}
	envDir, _ := EnvDir(cfgPath, "测试环境")
	res, err := Mirror(context.Background(), MirrorRequest{
		CfgPath: cfgPath, Env: "测试环境", Only: OnlyDict, Force: true,
		Tables: tables, Files: files}, nil)
	if err != nil {
		t.Fatalf("only dict: %v", err)
	}
	if res.MdlFiles != 0 {
		t.Errorf("only dict 不该重拉模板，实际 %d", res.MdlFiles)
	}
	m, err := ReadManifest(envDir)
	if err != nil || m == nil {
		t.Fatalf("manifest: %v", err)
	}
	if countRole(m, "mdl") != 3 {
		t.Errorf("only dict 后 mdl 条目应保留 3 条，实际 %d（抹掉就等于说模板没同步）", countRole(m, "mdl"))
	}
	if m.Versions.TemplateVersions["a05"] != 3 {
		t.Errorf("only dict 后模板版本锚点应保留，实际 %v", m.Versions.TemplateVersions)
	}
}

// TestMirrorCancelNoResidue 取消后：不报"失败"（是 context.Canceled）、不落半套。
func TestMirrorCancelNoResidue(t *testing.T) {
	cfgPath, tables, files := newMirrorFixture(t)
	envDir, _ := EnvDir(cfgPath, "测试环境")
	ctx, cancel := context.WithCancel(context.Background())
	// 数到第 5 张表就取消（模拟用户点"停止"）
	n := 0
	origCount := tables.countOverride
	_ = origCount
	tables.countErr = map[string]error{}
	go func() {}()

	// 用 Query 的取消路径：让第 5 张表的 Query 前 ctx 已被取消
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := Mirror(ctx, MirrorRequest{CfgPath: cfgPath, Env: "测试环境", Tables: tables, Files: files},
			func(p Progress) {
				if p.Phase == "table" {
					n++
					if n == 5 {
						cancel()
					}
				}
			})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("取消应返回 context.Canceled，得到 %v", err)
		}
	}()
	<-done

	if _, err := os.Stat(filepath.Join(envDir, "dict", "测试环境.db")); err == nil {
		t.Error("取消后不该留下规格库")
	}
	if _, err := os.Stat(filepath.Join(envDir, "manifest.json")); err == nil {
		t.Error("取消后不该留下 manifest")
	}
	if left := TempDirs(filepath.Join(envDir, "dict"), "build"); len(left) != 0 {
		t.Errorf("取消后还有临时目录: %v", left)
	}
	if left := TempDirs(envDir, "mdl-build"); len(left) != 0 {
		t.Errorf("取消后还有模板临时目录: %v", left)
	}
}

// TestMirrorTableWriteFailureRollsBack 写库失败 → 整体失败、不留正式落点。
func TestMirrorTableWriteFailureRollsBack(t *testing.T) {
	cfgPath, tables, files := newMirrorFixture(t)
	envDir, _ := EnvDir(cfgPath, "测试环境")
	tables.failOn["DZAC_T"] = errors.New("模拟查询失败")
	_, err := Mirror(context.Background(), MirrorRequest{
		CfgPath: cfgPath, Env: "测试环境", Tables: tables, Files: files}, nil)
	if err == nil {
		t.Fatal("应当失败")
	}
	if !strings.Contains(err.Error(), "DZAC_T") {
		t.Errorf("错误里应点名失败的表: %v", err)
	}
	if _, serr := os.Stat(filepath.Join(envDir, "dict", "测试环境.db")); serr == nil {
		t.Error("失败后不该留下规格库")
	}
	if m, _ := ReadManifest(envDir); m != nil {
		t.Error("失败后不该留下 manifest")
	}
}

// TestMirrorVersionWhere DZAA_T 用"当前有效版"过滤，其它表整表。
func TestMirrorVersionWhere(t *testing.T) {
	cfgPath, tables, files := newMirrorFixture(t)
	if _, err := Mirror(context.Background(), MirrorRequest{
		CfgPath: cfgPath, Env: "测试环境", Tables: tables, Files: files}, nil); err != nil {
		t.Fatal(err)
	}
	if w := tables.where["DZAA_T"]; !strings.Contains(w, "DZAASTUS='Y'") || !strings.Contains(w, "MAX(B.DZAA002)") {
		t.Errorf("DZAA_T 应带当前有效版过滤，实际 %q", w)
	}
	if w := tables.where["DZAB_T"]; w != "" {
		t.Errorf("其它表不该过滤（通用 STUS 过滤会静默清空 DZEB/DZED），实际 %q", w)
	}
}

// TestMirrorEmptyTableIsKept 0 行的表也要建出来（"空表"≠"缺表"）。
//
// 回归用例（真机实测踩过）：DZAS_T 有 0 行，早先的实现在"没有行就不知道列名"处报
// "表 DZAS_T 没有列"，整次 mirror 失败。正确行为是：结果集哪怕没有行，也要先给一次
// 列名（RowFunc 的表头通知），照常建表、行数记 0、**不进 missing**。
func TestMirrorEmptyTableIsKept(t *testing.T) {
	cfgPath, tables, files := newMirrorFixture(t)
	tables.missing["DZAS_T"] = false
	tables.setTable("DZAS_T", []string{"dzas001", "dzas002"}) // 有列、没有行
	res, err := Mirror(context.Background(), MirrorRequest{
		CfgPath: cfgPath, Env: "测试环境", Tables: tables, Files: files}, nil)
	if err != nil {
		t.Fatalf("空表不该让整次镜像失败: %v", err)
	}
	var got *TableEntry
	for i := range res.Tables {
		if res.Tables[i].Table == "DZAS_T" {
			got = &res.Tables[i]
		}
	}
	if got == nil {
		t.Fatal("空表也要出现在 tables[] 里（否则 status 的 53 张对不上）")
	}
	if got.Missing || got.Skipped || got.Rows != 0 {
		t.Fatalf("空表应记 0 行、既不是 missing 也不是 skipped: %+v", got)
	}
	if containsStr(res.Missing, "DZAS_T") {
		t.Error("空表不该进 missing")
	}
	// 空表真的建出来了（status 要能区分"空表"与"缺表"）
	envDir, _ := EnvDir(cfgPath, "测试环境")
	if _, err := os.Stat(filepath.Join(envDir, "dict", "测试环境.db")); err != nil {
		t.Fatalf("规格库没落盘: %v", err)
	}
}

// containsStr 小工具。
func containsStr(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
