package host

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMirrorEnvDirAndReady(t *testing.T) {
	dir := t.TempDir()
	if got := MirrorEnvDir(dir, "envA"); got != filepath.Join(dir, "envA") {
		t.Fatalf("MirrorEnvDir = %q", got)
	}
	if MirrorEnvDir("", "envA") != "" || MirrorEnvDir(dir, "") != "" {
		t.Fatal("空参数应返回空串")
	}
	if MirrorReady(dir, "envA") {
		t.Fatal("无基线标记应为 false")
	}
	if err := writeLocalMark(filepath.Join(dir, "envA")); err != nil {
		t.Fatalf("writeLocalMark: %v", err)
	}
	if !MirrorReady(dir, "envA") {
		t.Fatal("有基线标记应为 true")
	}
	if MirrorReady("", "envA") {
		t.Fatal("空镜像根应为 false")
	}
}

// countingReader 统计压缩字节,并在 EOF 时回调一次进度(小数据只触发收尾回调)。
func TestCountingReaderProgress(t *testing.T) {
	var got []MirrorProgress
	cr := &countingReader{
		r:      strings.NewReader("hello world"),
		total:  11,
		report: func(p MirrorProgress) { got = append(got, p) },
	}
	buf := make([]byte, 4)
	for {
		if _, err := cr.Read(buf); err != nil {
			break
		}
	}
	if cr.n != 11 {
		t.Fatalf("已读字节 = %d, want 11", cr.n)
	}
	if len(got) == 0 {
		t.Fatal("应至少回调一次进度")
	}
	last := got[len(got)-1]
	if last.Phase != "download" || last.Bytes != 11 || last.Total != 11 {
		t.Fatalf("末次进度 = %+v", last)
	}
}

// 备份/临时文件识别:用户反馈的实际样例 + 正常源码不应误判。
func TestIsBackupFile(t *testing.T) {
	backups := []string{
		"s_icd_wf_s02.4fd.bak", "s_icd_wf_s04.4fd.bak",
		"cs_aapp320.bck", "cs_aapp320.bck1", "cs_aapp320.bck2",
		"cs_axmt500_wf_excel.bck", "cs_axmt500_wf_excel.bck1", "cs_axmt500_wf_excel.bck2",
		"foo.4gl~", "foo.old", "foo.orig", "bar.tmp", "x.swp", "y.bak3",
	}
	for _, n := range backups {
		if !isBackupFile(n) {
			t.Errorf("应判为备份: %s", n)
		}
	}
	kept := []string{
		"s_icd_wf_s02.4fd", "s_icd_wf_s04.4fd",
		"cs_aapp320.4gl", "cs_axmt500_wf_excel.4gl",
		"apmt520_bak_service.4gl",          // 含 "_bak" 但不是备份文件
		"lib_common.inc", "b_sysparam.inc", // .inc 是源码,不是备份
		localMarkName, "dzea_t",
	}
	for _, n := range kept {
		if isBackupFile(n) {
			t.Errorf("不应判为备份: %s", n)
		}
	}
}

// 服务器侧 find 排除条件按同一份后缀列表生成。
func TestMirrorBackupFindExpr(t *testing.T) {
	expr := mirrorBackupFindExpr()
	for _, suf := range mirrorBackupSuffixes {
		if !strings.Contains(expr, `-not -iname '*`+suf+`*'`) {
			t.Errorf("排除条件缺少 %s: %s", suf, expr)
		}
	}
	if !strings.Contains(expr, `-not -iname '*~'`) {
		t.Errorf("排除条件缺少 ~: %s", expr)
	}
}

// 服务器侧 find 白名单:4gl/4fd + 42s 的 zh_CN 语言目录 + *.inc。
func TestMirrorWhitelistFindExpr(t *testing.T) {
	expr := mirrorWhitelistFindExpr()
	for _, want := range []string{`-path '*/4gl/*'`, `-path '*/4fd/*'`,
		`-path '*/42s/zh_CN/*'`, `-path '*/42s/*/zh_CN/*'`, `-iname '*.inc'`} {
		if !strings.Contains(expr, want) {
			t.Errorf("白名单缺少 %s: %s", want, expr)
		}
	}
}

// 基线标记记录白名单版本;旧格式(纯时间戳)与缺失都视为 0(触发自动全量)。
func TestLocalMarkVersion(t *testing.T) {
	dir := t.TempDir()
	if got := localMarkVersion(dir); got != 0 {
		t.Fatalf("无标记应为 0, got %d", got)
	}
	if err := writeLocalMark(dir); err != nil {
		t.Fatal(err)
	}
	if got := localMarkVersion(dir); got != mirrorWhitelistVersion {
		t.Fatalf("版本 = %d, want %d", got, mirrorWhitelistVersion)
	}
	// 旧格式(仅 RFC3339 时间戳)
	legacy := t.TempDir()
	if err := os.WriteFile(filepath.Join(legacy, localMarkName), []byte("2026-01-01T00:00:00Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := localMarkVersion(legacy); got != 0 {
		t.Fatalf("旧格式应为 0, got %d", got)
	}
}

// 本地残留备份清理:只删备份,不碰源码与目录。
func TestPruneBackupFiles(t *testing.T) {
	dir := t.TempDir()
	mk := func(rel string) string {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	keepA := mk("erp/asf/4gl/cs_aapp320.4gl")
	keepB := mk("erp/asf/4fd/s_icd_wf_s02.4fd")
	mk("erp/asf/4gl/cs_aapp320.bck")
	mk("erp/asf/4gl/cs_aapp320.bck1")
	mk("erp/asf/4fd/s_icd_wf_s02.4fd.bak")
	mk("erp/asf/4gl/sub/cs_x.4gl~")

	if got := pruneBackupFiles(dir); got != 4 {
		t.Fatalf("应删除 4 个备份, got %d", got)
	}
	for _, p := range []string{keepA, keepB} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("源码不应被删: %s (%v)", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "erp/asf/4gl/cs_aapp320.bck")); err == nil {
		t.Error("备份应已删除")
	}
	// 幂等:再清一次为 0
	if got := pruneBackupFiles(dir); got != 0 {
		t.Fatalf("二次清理应为 0, got %d", got)
	}
	// 不存在目录不报错
	if got := pruneBackupFiles(filepath.Join(dir, "nope")); got != 0 {
		t.Fatalf("不存在目录应为 0, got %d", got)
	}
}
