package devdata

// 领域层其余分支的补充测试（长跑任务状态机、status/probe、pack 边界、解析器畸形输入）。
//
// 这些不是"为覆盖率而测"：每一条都对应一次真机/交互中会走到的分支 ——
// 任务被取消时不能算失败、probe 时 design/dict 类素材不能去服务器找、pack 的
// 身份推断失败要警告、.tab/.4fd 在传输中被截断时不能整份崩掉。

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/config"
)

// TestJobManager_StateMachine 钉住长跑任务的状态机：单飞、取消≠失败、失败原因、
// 结束顺序（运行中在前）以及 Describe 的四种结论文案。
func TestJobManager_StateMachine(t *testing.T) {
	m := NewManager()
	base := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return base }

	// 成功路径：进度回报要能更新 phase/message/total/currentFile。
	started, err := m.Start(KindFetch, "e", "p1", func(_ context.Context, report ProgressFunc) (any, error) {
		report(Progress{Phase: "fetch", Message: "拉取中", Total: 10, CurrentFile: "DZFK_T"})
		report(Progress{Done: 5})
		return "结果", nil
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !started.Running || started.Phase != "start" || started.Kind != KindFetch || started.Env != "e" {
		t.Errorf("启动快照 = %+v", started)
	}
	done := m.Wait(KindFetch, "e", "p1")
	if done.Running || done.Phase != "done" || done.Done != 10 || done.Result != "结果" {
		t.Errorf("完成快照 = %+v", done)
	}
	if done.FinishedAt == "" || done.Message != "拉取中" {
		t.Errorf("完成时间/消息 = %+v", done)
	}
	if !strings.Contains(started.Describe(), "/p1 运行中") || !strings.Contains(done.Describe(), "完成") {
		t.Errorf("Describe = %q / %q", started.Describe(), done.Describe())
	}
	if got := m.JobsFor("e"); len(got) != 1 || got[0].Prog != "p1" {
		t.Errorf("JobsFor = %+v", got)
	}
	if all := m.All(); len(all) != 1 {
		t.Errorf("All = %+v", all)
	}
	if jobs := m.Running("e"); len(jobs) != 0 {
		t.Errorf("已结束的任务不该在 Running 里: %+v", jobs)
	}
	if _, ok := m.Snapshot(KindFetch, "e", "nope"); ok {
		t.Error("没跑过的键 Snapshot 应为 false")
	}
	if z := m.Wait(KindFetch, "e", "nope"); z.Kind != "" {
		t.Errorf("没跑过的键 Wait 应返回零值: %+v", z)
	}

	// 单飞 + 取消：取消是"已停止"，不是失败。
	block := make(chan struct{})
	cancelStarted, err := m.Start(KindMirror, "e", "", func(ctx context.Context, report ProgressFunc) (any, error) {
		close(block)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	if err != nil {
		t.Fatalf("Start(mirror): %v", err)
	}
	<-block
	if cancelStarted.Phase != "start" {
		t.Errorf("取消中的启动快照 = %+v", cancelStarted)
	}
	if _, err := m.Start(KindMirror, "e", "", func(context.Context, ProgressFunc) (any, error) {
		return nil, nil
	}); !errors.Is(err, ErrBusy) {
		t.Errorf("同类任务并发 err = %v，想要 ErrBusy", err)
	}
	if err := m.Stop(KindMirror, "e", ""); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	cancelled := m.Wait(KindMirror, "e", "")
	if !cancelled.Cancelled || cancelled.Phase != "cancelled" || cancelled.Error != "" {
		t.Errorf("取消后的快照 = %+v（取消不是失败）", cancelled)
	}
	if !strings.Contains(cancelled.Describe(), "已停止") {
		t.Errorf("Describe = %q", cancelled.Describe())
	}
	if err := m.Stop(KindMirror, "e", ""); !errors.Is(err, ErrNoJob) {
		t.Errorf("没有任务时 Stop err = %v，想要 ErrNoJob", err)
	}

	// 失败路径：Error/Message 都要带上原因。
	if _, err := m.Start(KindFetch, "e", "p2", func(context.Context, ProgressFunc) (any, error) {
		return nil, fmt.Errorf("连不上库")
	}); err != nil {
		t.Fatalf("Start(fail): %v", err)
	}
	failed := m.Wait(KindFetch, "e", "p2")
	if failed.Phase != "error" || failed.Error != "连不上库" || !strings.Contains(failed.Describe(), "失败") {
		t.Errorf("失败快照 = %+v / %q", failed, failed.Describe())
	}

	// 其余小工具：nil 回调可安全调用、单飞键区分 fetch 的程序。
	progressReporter(nil)(Progress{Phase: "x"})
	if jobKey(KindFetch, "e", "p") == jobKey(KindFetch, "e", "q") {
		t.Error("fetch 的单飞键必须带程序名")
	}
	if jobKey(KindMirror, "e", "p") != jobKey(KindMirror, "e", "") {
		t.Error("mirror 的单飞键不该按程序区分")
	}
	if nj := NewManager(); nj == nil || nj.jobs == nil {
		t.Error("NewManager 没初始化 jobs")
	}
}

// TestStatus_MatrixAndProbe 钉住 status 的两条路：
//   - 默认只读本地（不出网）：环境目录/字典/模板树/程序清单，manifest 警告要带出来；
//   - Check 强制 probe：本地文件与服务器现状比对（remote-gone/newer/version-changed），
//     design/dict 这类"不是服务器素材"的条目要跳过（否则每次 check 都误报）。
func TestStatus_MatrixAndProbe(t *testing.T) {
	ctx := context.Background()
	set := newMaterials(t, matOpts{
		Prog:             "axmt500_wf",
		ManifestWarnings: []string{"蓝图比骨架新"},
		ManifestMissing:  []string{"erp/cxm/dzx/tap/axmt500_wf.tap"},
		Slices:           []matSlice{{Code: "a05", Used: 3, Current: 3}},
	})
	set.mapIndex()

	// 字典文件与模板树都造出来，status 才能报 present/文件数。
	dictPath, _ := DictPath(set.CfgPath, "e")
	if err := os.MkdirAll(filepath.Dir(dictPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dictPath, []byte("SQLite format 3\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 模板树文件由构建器写好；这里只需要环境 manifest 里登记它（status 按登记计数）。
	mdlRel := filepath.ToSlash(filepath.Join("mdl", "slice", "code_a05.template"))
	tplBody := "#應用 a05 樣板自動產生(Version:3)\n"
	// 环境级 manifest（M 侧）：status 的 Dict/Mdl 汇总只读它。
	envDir := filepath.Dir(filepath.Dir(set.Dir))
	dictRel := filepath.ToSlash(filepath.Join(config.DevDataSubdirDict, config.DictDBFileName("e")))
	envMF := NewManifest("mirror", "e")
	envMF.FetchedAt = "2026-11-21T10:00:00+08:00"
	envMF.Files = []FileEntry{
		{Role: "mdl", Rel: mdlRel, Src: "erp/mdl/slice/code_a05.template", Size: int64(len(tplBody)), SHA256: shaOfString(tplBody)},
		{Role: "dict", Rel: dictRel, Src: "dict", Size: 16},
	}
	envMF.Tables = []TableEntry{{Table: "DZAA_T", Rows: 10, Skipped: true}}
	envMF.AddWarning("蓝图比骨架新")
	if err := WriteManifest(envDir, envMF); err != nil {
		t.Fatalf("写环境 manifest: %v", err)
	}
	// 程序 manifest 里追加一条本地归档（design）与 dict 角色：probe 时必须跳过它们。
	designRel := "design/x.json"
	designBody := `{"local":true}`
	if err := os.MkdirAll(filepath.Dir(set.filePath(designRel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(set.filePath(designRel), []byte(designBody), 0o644); err != nil {
		t.Fatal(err)
	}
	mdlProgRel := "mdl/local-note.txt"
	mdlProgBody := "note"
	if err := os.MkdirAll(filepath.Dir(set.filePath(mdlProgRel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(set.filePath(mdlProgRel), []byte(mdlProgBody), 0o644); err != nil {
		t.Fatal(err)
	}
	set.rewriteManifest(func(mf *Manifest) {
		mf.Files = append(mf.Files,
			FileEntry{Role: "design", Rel: designRel, Src: "local:x", Size: int64(len(designBody)), SHA256: shaOfString(designBody)},
			FileEntry{Role: "dict", Rel: mdlProgRel, Src: "dict", Size: int64(len(mdlProgBody)), SHA256: shaOfString(mdlProgBody)},
		)
		mf.Versions.TglHeader = &TglHeader{SDVer: "0001"}
	})

	st, err := Status(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "e", Jobs: []Job{{Kind: KindFetch}}})
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !st.Present || st.Env != "e" || st.Dir != envDir {
		t.Errorf("环境状态 = %+v", st)
	}
	if st.Dict == nil || !st.Dict.Present || st.Dict.Size == 0 || st.Dict.Path != dictPath {
		t.Errorf("字典状态 = %+v", st.Dict)
	}
	if st.Dict.Skipped != 1 || st.Dict.Rows != 10 {
		t.Errorf("字典表统计 = %+v（skipped 的表不算进 Tables）", st.Dict)
	}
	if st.Dict.Tables != 0 {
		t.Errorf("全被 skip 的表数 = %+v", st.Dict)
	}
	if st.Mdl == nil || !st.Mdl.Present || st.Mdl.Files != 1 || st.Mdl.TemplateCount != 1 {
		t.Errorf("模板树状态 = %+v", st.Mdl)
	}
	if len(st.Progs) != 1 || st.Progs[0].Prog != "axmt500_wf" || !st.Progs[0].Exists {
		t.Errorf("程序清单 = %+v", st.Progs)
	}
	if st.Progs[0].Files == 0 || st.Progs[0].Bytes == 0 || st.Progs[0].Stale {
		t.Errorf("程序状态 = %+v", st.Progs[0])
	}
	if len(st.Progs[0].Missing) != 1 || len(st.Progs[0].Warnings) == 0 {
		t.Errorf("程序 missing/warnings = %+v", st.Progs[0])
	}
	hasWarn := false
	for _, w := range st.Warnings {
		if strings.Contains(w, "蓝图比骨架新") {
			hasWarn = true
		}
	}
	if !hasWarn {
		t.Errorf("环境警告没带出 manifest warnings: %v", st.Warnings)
	}
	if len(st.Jobs) != 1 {
		t.Errorf("jobs 没带出: %+v", st.Jobs)
	}
	// 只查单个程序。
	if one, _ := Status(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt500_wf"}); len(one.Progs) != 1 {
		t.Errorf("按程序过滤 = %+v", one.Progs)
	}
	if none, _ := Status(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt999"}); len(none.Progs) != 0 {
		t.Errorf("不存在的程序不该出现: %+v", none.Progs)
	}
	// 临时目录残留不进程序清单。
	if err := os.MkdirAll(filepath.Join(set.Dir, "..", "x.tmp-dead"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got, _ := Status(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "e"}); len(got.Progs) != 1 {
		t.Errorf("临时目录被当成程序: %+v", got.Progs)
	}
	// 环境目录不存在：返回"未拉取"的完整空状态，不报错。
	empty := writeConfig(t, `{"name":"e","host":"127.0.0.1","port":22,"user":"u"}`)
	if got, err := Status(ctx, StatusRequest{CfgPath: empty, Env: "e"}); err != nil || got.Present {
		t.Errorf("未拉取环境 = %+v / %v", got, err)
	}
	// 环境名错误：区分非法与不存在。
	if _, err := Status(ctx, StatusRequest{CfgPath: set.CfgPath, Env: ""}); !errors.Is(err, ErrInvalidEnv) {
		t.Errorf("空环境 err = %v", err)
	}
	if _, err := Status(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "nope"}); !errors.Is(err, ErrEnvNotFound) {
		t.Errorf("未知环境 err = %v", err)
	}
	// manifest 坏掉：status 要能继续给其余信息，并明确警告。
	bad := newMinimalMaterials(t, "axmt501")
	if err := os.WriteFile(filepath.Join(bad.Dir, ManifestFileName), []byte("{坏"), 0o644); err != nil {
		t.Fatal(err)
	}
	bs, err := Status(ctx, StatusRequest{CfgPath: bad.CfgPath, Env: "e", Prog: "axmt501"})
	if err != nil {
		t.Fatalf("Status(bad manifest): %v", err)
	}
	// 坏的程序 manifest：status 仍要给程序条目，并标 stale + 原因（不是整页失败）。
	if len(bs.Progs) != 1 || !bs.Progs[0].Stale {
		t.Errorf("坏 manifest 的程序状态 = %+v", bs.Progs)
	}
	if !hasStaleCode(bs.Progs[0].StaleReasons, "missing-local") ||
		!strings.Contains(bs.Progs[0].StaleReasons[0].Detail, "JSON") {
		t.Errorf("坏 manifest 原因 = %+v", bs.Progs[0].StaleReasons)
	}

	// Check = 强制 probe。用内存 FileSource（fakeFiles）核对服务器现状。
	mf := set.readManifest()
	ff := newFakeFiles()
	for _, f := range mf.Files {
		if f.Src == "" || strings.HasPrefix(f.Src, "local:") || f.Role == "dict" {
			continue
		}
		body := set.files[f.Rel]
		if body == "" {
			raw, _ := os.ReadFile(set.filePath(f.Rel))
			body = string(raw)
		}
		ff.files[f.Src] = []byte(body)
	}
	ps, err := Check(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt500_wf", Files: ff})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !ps.Probed {
		t.Error("Check 应该做过 probe")
	}
	// design/local 与 dict 不参与服务器比对，不该产生 remote-gone。
	for _, r := range ps.StaleReasons {
		if r.Code == "remote-gone" && strings.Contains(r.Detail, "local:") {
			t.Errorf("design 素材被当成服务器文件: %+v", r)
		}
	}
	// tgl 头部 SD 版次变化 → version-changed。
	if !hasStaleCode(ps.StaleReasons, "version-changed") {
		t.Errorf("tgl 版次变化没被 check 出来: %+v", ps.StaleReasons)
	}
	// 服务器上少一个文件 → remote-gone；文件变新 → remote-newer。
	delete(ff.files, mf.Files[0].Src)
	for src := range ff.files {
		if strings.HasSuffix(src, ".4gl") {
			ff.mtimes[src] = matMustTime(t, "2027-01-01T00:00:00+08:00")
			break
		}
	}
	ps2, err := Check(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt500_wf", Files: ff})
	if err != nil {
		t.Fatalf("Check(mutated): %v", err)
	}
	if !hasStaleCode(ps2.StaleReasons, "remote-gone") || !hasStaleCode(ps2.StaleReasons, "remote-newer") {
		t.Errorf("probe 结论 = %+v", ps2.StaleReasons)
	}
	// Check 的程序不存在/名字非法。
	if _, err := Check(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt999"}); !errors.Is(err, ErrProgNotFound) {
		t.Errorf("不存在程序 err = %v", err)
	}
	if _, err := Check(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "e", Prog: "BAD"}); !errors.Is(err, ErrInvalidProg) {
		t.Errorf("非法程序 err = %v", err)
	}
	// 没有注入 FileSource 时会去连 SSH；测试环境的 config 没配 zone，必须快速失败而不是挂住。
	if _, err := Check(ctx, StatusRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt500_wf"}); err == nil ||
		!strings.Contains(err.Error(), "登录区域") {
		t.Errorf("未配 zone 的 probe err = %v", err)
	}

	// AgeText：相对时间文案（空/非法/刚刚/分钟/小时/天）。
	if AgeText("") != "" || AgeText("not-a-time") != "not-a-time" {
		t.Errorf("AgeText 空/非法 = %q / %q", AgeText(""), AgeText("not-a-time"))
	}
	now := time.Now()
	if got := AgeText(now.Format(time.RFC3339)); got != "刚刚" {
		t.Errorf("AgeText(刚刚) = %q", got)
	}
	if got := AgeText(now.Add(-5 * time.Minute).Format(time.RFC3339)); !strings.Contains(got, "分钟前") {
		t.Errorf("AgeText(5m) = %q", got)
	}
	if got := AgeText(now.Add(-2 * time.Hour).Format(time.RFC3339)); !strings.Contains(got, "小时前") {
		t.Errorf("AgeText(2h) = %q", got)
	}
	if got := AgeText(now.Add(-48 * time.Hour).Format(time.RFC3339)); got != "2 天前" {
		t.Errorf("AgeText(48h) = %q", got)
	}
}

// hasStaleCode 判断 stale 原因里有没有某个 code。
func hasStaleCode(reasons []StaleReason, code string) bool {
	for _, r := range reasons {
		if r.Code == code {
			return true
		}
	}
	return false
}

// TestPack_Edges 钉住打包的边界：身份推断失败要警告、材料被本地改过要警告、
// --out 的目录/文件语义、以及"目标已存在"的强制覆盖。
func TestPack_Edges(t *testing.T) {
	set := newMaterials(t, matOpts{Prog: "axmt500_wf", Identity: "s"})

	// 身份从 .tap 根属性读出（s）。
	res, err := Pack(PackRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt500_wf"})
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if res.Identity != "s" || filepath.Base(res.Path) != "axmt500_wf(s).tzc" {
		t.Errorf("身份/包名 = %q / %q", res.Identity, filepath.Base(res.Path))
	}
	if len(res.Entries) != 4 || res.Entries[3].Name != "ver" || res.Entries[3].Role != "ver" {
		t.Errorf("包内成员 = %+v", res.Entries)
	}
	if res.Elapsed == "" || res.Bytes <= 0 || res.SHA256 == "" || res.Path == "" {
		t.Errorf("打包结果 = %+v", res)
	}

	// --out 是目录（带/不带分隔符两种写法）：放进目录里，文件名不变。
	outDir := t.TempDir()
	toDir, err := Pack(PackRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt500_wf", Out: outDir + string(os.PathSeparator), Force: true})
	if err != nil {
		t.Fatalf("Pack(dir): %v", err)
	}
	if filepath.Dir(toDir.Path) != outDir || filepath.Base(toDir.Path) != "axmt500_wf(s).tzc" {
		t.Errorf("目录输出 = %q", toDir.Path)
	}
	toDir2, err := Pack(PackRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt500_wf", Out: outDir, Force: true})
	if err != nil || filepath.Dir(toDir2.Path) != outDir {
		t.Errorf("目录输出(无分隔符) = %+v / %v", toDir2, err)
	}
	// 目标已存在：不给 force 拒绝。
	again := PackRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt500_wf", Out: filepath.Join(outDir, "x.tzc")}
	if _, err := Pack(again); err != nil {
		t.Fatalf("Pack(x.tzc): %v", err)
	}
	if _, err := Pack(again); !errors.Is(err, ErrExists) {
		t.Errorf("重复写 err = %v，想要 ErrExists", err)
	}
	again.Force = true
	if _, err := Pack(again); err != nil {
		t.Errorf("--force err = %v", err)
	}
	// 输出路径的父路径是个文件：必须报"创建输出目录失败"。
	f := filepath.Join(t.TempDir(), "notadir")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Pack(PackRequest{CfgPath: set.CfgPath, Env: "e", Prog: "axmt500_wf", Out: filepath.Join(f, "p.tzc")}); err == nil ||
		!strings.Contains(err.Error(), "创建输出目录失败") {
		t.Errorf("坏父路径 err = %v", err)
	}
	// 非法环境：打包前就该报环境不存在。
	if _, err := Pack(PackRequest{CfgPath: set.CfgPath, Env: "nope", Prog: "axmt500_wf"}); !errors.Is(err, ErrEnvNotFound) {
		t.Errorf("非法环境 err = %v", err)
	}

	// 身份推断失败：.tap 没有 identity 属性 → 警告并按客制命名。
	noID := newMaterials(t, matOpts{Prog: "axmt500_wf"})
	tapBody := strings.Replace(noID.files["dzx/axmt500_wf.tap"], `identity="c" `, "", 1)
	noID.writeFile("dzx/axmt500_wf.tap", tapBody)
	guess, err := Pack(PackRequest{CfgPath: noID.CfgPath, Env: "e", Prog: "axmt500_wf"})
	if err != nil {
		t.Fatalf("Pack(no identity): %v", err)
	}
	if guess.Identity != "c" || len(guess.Warnings) == 0 ||
		!strings.Contains(strings.Join(guess.Warnings, "；"), "identity") {
		t.Errorf("身份推断 = %+v / %v", guess.Identity, guess.Warnings)
	}

	// 材料被本地改过（manifest 的 sha 对不上）：必须警告，不能默默打包。
	tamper := newMaterials(t, matOpts{Prog: "axmt500_wf"})
	tapPath := filepath.Join(tamper.Dir, "dzx", "axmt500_wf.tap")
	raw, err := os.ReadFile(tapPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tapPath, append(raw, []byte("<!-- 本地改动 -->\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	tampered, err := Pack(PackRequest{CfgPath: tamper.CfgPath, Env: "e", Prog: "axmt500_wf"})
	if err != nil {
		t.Fatalf("Pack(tampered): %v", err)
	}
	if !strings.Contains(strings.Join(tampered.Warnings, "；"), "sha256 不一致") {
		t.Errorf("本地改动没警告: %v", tampered.Warnings)
	}

	// 小工具口径。
	if d, err := PackMaterialDir(PackRequest{Dir: set.Dir}); err != nil || d != set.Dir {
		t.Errorf("PackMaterialDir = %q/%v", d, err)
	}
	if PackFileName("p", "c") != "p(c).tzc" {
		t.Error("PackFileName 口径错")
	}
	if !isDirLike("a/") || !isDirLike("a"+string(os.PathSeparator)) || isDirLike("a") {
		t.Error("isDirLike 口径错")
	}
	if got := detectIdentity([]packMember{{role: "4gl", data: []byte(`identity="s"`)}}); got != "" {
		t.Errorf("没有 .tap 时不该推断出身份: %q", got)
	}
	if fileExt("other") != ".other" || fileExt("inc") != ".inc" {
		t.Errorf("fileExt = %q/%q", fileExt("other"), fileExt("inc"))
	}
}

// TestLink_RequestOverridesAndMalformed 钉住 LinkRequest 的三个落点覆盖（Dir/Out/Mdl）
// 与畸形材料的错误路径：读不到某份文件、.tap 坏掉、写索引失败。
func TestLink_RequestOverridesAndMalformed(t *testing.T) {
	set := newMinimalMaterials(t, "axmt500_wf")
	out := filepath.Join(t.TempDir(), "custom-index.json")
	doc, err := Map(LinkRequest{Dir: set.Dir, Out: out, Mdl: set.Mdl, Env: "e", Prog: "axmt500_wf"})
	if err != nil {
		t.Fatalf("Map(overrides): %v", err)
	}
	if doc.Path != out {
		t.Errorf("Out 覆盖 = %q", doc.Path)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("Out 指定的索引没写出来: %v", err)
	}

	// 拿不到模板树根（配置文件坏）：脱版检测应当跳过而不是报错。
	dirOnly := LinkRequest{Dir: set.Dir, CfgPath: filepath.Join(t.TempDir(), "no.json"), Env: "e", Prog: "axmt500_wf"}
	if d, err := Map(dirOnly); err != nil {
		t.Fatalf("Map(no cfg): %v", err)
	} else if matCheckLevel(d, "slice-version") != "missing" {
		t.Errorf("没有模板树时不该有 slice-version 检查: %+v", d.Checks)
	}

	// 各角色文件被删（manifest 还在）：报"读材料失败"并点出文件。
	for _, tc := range []struct{ role, rel string }{
		{"tap", "dzx/axmt500_wf.tap"},
		{"4gl", "src/axmt500_wf.4gl"},
		{"tab", "dzx/axmt500_wf.tab"},
		{"tsd", "dzx/axmt500_wf.tsd"},
		{"4fd", "dzx/axmt500_wf.4fd"},
	} {
		t.Run("missing_"+tc.role, func(t *testing.T) {
			s := newMinimalMaterials(t, "axmt500_wf")
			if err := os.Remove(s.filePath(tc.rel)); err != nil {
				t.Fatal(err)
			}
			_, err := Map(s.LinkRequest())
			if err == nil || !strings.Contains(err.Error(), "读材料失败") {
				t.Errorf("缺 %s 的 err = %v", tc.role, err)
			}
		})
	}

	// .tap 坏掉：解析错误要直说（而不是当成空骨架继续建索引）。
	bad := newMinimalMaterials(t, "axmt500_wf")
	bad.writeFile("dzx/axmt500_wf.tap", "<add_points><point")
	if _, err := Map(bad.LinkRequest()); err == nil || !strings.Contains(err.Error(), "解析 .tap 失败") {
		t.Errorf("坏 .tap err = %v", err)
	}

	// 写索引失败：Out 指向一个非空目录 → 改名失败，不能返回"看起来成功"的索引。
	fail := newMinimalMaterials(t, "axmt500_wf")
	outDir := filepath.Join(fail.Dir, "outdir")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "keep"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Map(LinkRequest{Dir: fail.Dir, Out: outDir, Env: "e", Prog: "axmt500_wf"}); err == nil ||
		!strings.Contains(err.Error(), "改名失败") {
		t.Errorf("写索引失败 err = %v", err)
	}

	// 模板（无声明）点位：Kind 应判"模板"（不是客制/标准）。
	tplSet := newMaterials(t, matOpts{
		Prog: "axmt510",
		Sections: []matSection{{
			ID: "p.input", Ver: "1", Src: "s",
			Slots: []matSlot{{Name: "input.a.page1.f1", Label: "模板点", Body: []string{"  LET f1 = 1"}}},
		}},
		NoTsd: true, No4fd: true,
	})
	tpl := tplSet.mapIndex()
	if len(tpl.Points) != 1 || tpl.Points[0].Kind() != "模板" {
		t.Errorf("模板点 = %+v / %q", tpl.Points, tpl.Points[0].Kind())
	}

	// 超长正文：Preview 要截断（否则 brief 里一行几千字）。
	long := newMaterials(t, matOpts{
		Prog: "axmt511",
		Sections: []matSection{{
			ID: "p.input", Ver: "1", Src: "s",
			Slots: []matSlot{{
				Name: "input.a.page1.f1", Label: "长正文", Src: "c",
				DeclBody: "LET x = '" + strings.Repeat("长", 120) + "'\n",
				Body:     []string{"  LET x = 1"},
			}},
		}},
		NoTab: true, NoTsd: true, No4fd: true,
	})
	ld := long.mapIndex()
	if len(ld.Points) != 1 || !strings.HasSuffix(ld.Points[0].Preview, "…") {
		t.Errorf("超长预览没截断: %q", ld.Points[0].Preview)
	}

	// srcMtime 不是合法时间：freshness 检查跳过（不能因此报错）。
	badTime := newMinimalMaterials(t, "axmt512")
	badTime.rewriteManifest(func(mf *Manifest) {
		for i := range mf.Files {
			if mf.Files[i].Role == "tap" {
				mf.Files[i].SrcMtime = "not-a-time"
			}
		}
	})
	if d := badTime.mapIndex(); matCheckLevel(d, "freshness") != "missing" {
		t.Errorf("非法 srcMtime 时不该有 freshness 检查: %+v", d.Checks)
	}

	// 客制函数没有 END FUNCTION：行号一直算到文件尾（不能返回 0）。
	openFn := newMaterials(t, matOpts{
		Prog:  "axmt513",
		Funcs: []matFunc{{Name: "p_feature", Body: []string{"  LET y = 1"}}},
		NoTab: true, NoTsd: true, No4fd: true,
	})
	openFn.mapIndex()
	body := strings.Replace(openFn.files["src/axmt513.4gl"], "END FUNCTION\n", "", 1)
	openFn.writeFile("src/axmt513.4gl", body)
	od := openFn.mapIndex()
	var fn *LinkPoint
	for i := range od.Points {
		if od.Points[i].Unmarked {
			fn = &od.Points[i]
		}
	}
	if fn == nil || fn.Line1 != od.Stats.G4Lines {
		t.Errorf("没有 END FUNCTION 时行号 = %+v / total=%d", fn, od.Stats.G4Lines)
	}
}

// TestLinkParse_EdgeForms 钉住解析器对真机畸形/少见表单的容忍：
// 截断的标签、单引号属性、没有 text 的 4fd 页、空 sfield、空 value。
func TestLinkParse_EdgeForms(t *testing.T) {
	// .tap 非法 XML：必须报错（不能当成空骨架）。
	if _, err := parseTap([]byte("<add_points><point")); err == nil {
		t.Error("非法 .tap 应报错")
	}
	// section 的 type/readonly 属性要读进来（设计器只读段靠它）。
	set := newMaterials(t, matOpts{
		Prog: "axmt520",
		Sections: []matSection{{
			ID: "p.readonly", Ver: "2", Src: "s", Type: "s", Readonly: "Y",
			Slots: []matSlot{{Name: "input.a.page1.f1", Label: "只读段", Body: []string{"  LET f1 = 1"}}},
		}},
	})
	tap, err := parseTap([]byte(set.files["dzx/axmt520.tap"]))
	if err != nil {
		t.Fatalf("parseTap: %v", err)
	}
	if len(tap.Sections) != 1 || tap.Sections[0].Type != "s" || tap.Sections[0].Readon != "Y" {
		t.Errorf("section 属性 = %+v", tap.Sections)
	}

	// .tab 被截断（传输中断）：不报错、能读到的部分照读。
	truncated := `<assembly name="p" module="m" type="t01"><structure><var id="body" page="1" page_id="7" record="s7" value="f1,f2"/>` +
		`<var id="head" value="h1"`
	tab, err := parseTab([]byte(truncated))
	if err != nil {
		t.Fatalf("parseTab(truncated): %v", err)
	}
	if tab.Assembly.Type != "t01" || len(tab.Pages) != 1 || len(tab.Pages[0].Columns) != 2 {
		t.Errorf("截断 .tab = %+v", tab)
	}
	// 属性里用单引号 + head 的 value 为空。
	quoted := `<assembly name='p' type='q01' cusver='0000'><structure>` +
		`<var id='head' value=''/><var id='body' page='1' page_id='7' record='s7' value='f1(chr10),f2'/></structure></assembly>`
	tab2, err := parseTab([]byte(quoted))
	if err != nil {
		t.Fatalf("parseTab(single quote): %v", err)
	}
	if tab2.Assembly.Type != "q01" || tab2.Assembly.CusVer != "0000" || len(tab2.Head) != 0 {
		t.Errorf("单引号 .tab = %+v head=%v", tab2.Assembly, tab2.Head)
	}

	// .tsd：没有 text 的 sfield 跳过（不能写进标签表）。
	tsd := `<spec><strings><sfield name="lbl_f1"/><sfield name="lbl_f2" text="二号"/><sfield name="page_1" text="基本"/></strings></spec>`
	spec := parseTsd([]byte(tsd))
	if _, ok := spec.Fields["f1"]; ok {
		t.Error("没有 text 的标签不该进表")
	}
	if spec.Fields["f2"] != "二号" || spec.Pages["page_1"] != "基本" {
		t.Errorf("tsd = %+v", spec)
	}

	// .4fd：截断的 <Page 属性 → 停止解析（不 panic、已读的保留）。
	fd := `<ManagedForm><Page name="page_1" text="page_1"><Table name="s1"/></Page>` +
		`<Page name="page_2" text="page_2"><Table name="s2"`
	fdSpec := parse4fd([]byte(fd))
	if fdSpec.PageByRecord["s1"] != "page_1" {
		t.Errorf("截断 4fd 已读部分 = %+v", fdSpec.PageByRecord)
	}
	// Page 没有 text：退回 name 当键；`<PageX>` 不是 Page，不能被误认。
	fd2 := parse4fd([]byte(`<ManagedForm><PageX name="bad"><Table name="sx"/></PageX>` +
		`<Page name="page_9"><Table name="s9"/></Page><Table name="orphan"/></ManagedForm>`))
	if fd2.PageByRecord["sx"] != "" {
		t.Errorf("<PageX> 被误当成 Page: %+v", fd2.PageByRecord)
	}
	if got := fd2.PageByRecord["s9"]; got != "page_9" {
		t.Errorf("没有 text 时页签键 = %q，想要 name", got)
	}
	if _, ok := fd2.PageByRecord["orphan"]; ok {
		t.Error("Page 外的 Table 不该被归属")
	}

	// 空 .tsd：applyFieldLabels 直接返回（不能把字段标签写成空串覆盖）。
	noLabels := newMaterials(t, matOpts{Prog: "axmt521"})
	noLabels.writeFile("dzx/axmt521.tsd", `<spec><strings></strings></spec>`)
	d := noLabels.mapIndex()
	if len(d.Fields) != 1 || d.Fields[0].Label != "" {
		t.Errorf("空 tsd 时的标签 = %+v", d.Fields)
	}
}

// TestQuery_EdgeForms 钉住查询层的兜底分支：孤儿子串命中、无页面名的孤儿提示、
// 没切片的命中、以及非法程序的错误传播。
func TestQuery_EdgeForms(t *testing.T) {
	// 只有孤儿、没有任何点位：Show 仍要能按子串找到它（最后一路兜底）。
	orphanOnly := newMaterials(t, matOpts{
		Prog:      "axmt530",
		Freestyle: true, // 没有 section 锚 → 没有点位
		Orphans: []matOrphan{
			{Name: "input.a.page73.xmdducseq", Src: "c", Ver: "2", Body: "LET x = 1\n"},
		},
		NoTab: true, NoTsd: true, No4fd: true,
	})
	orphanOnly.mapIndex()
	sub, err := Show(ShowRequest{LinkRequest: orphanOnly.LinkRequest(), Query: "ducseq"})
	if err != nil {
		t.Fatalf("Show(子串): %v", err)
	}
	if sub.Match != "orphan" || len(sub.Points) != 1 || sub.Points[0].Name != "input.a.page73.xmdducseq" {
		t.Errorf("孤儿子串命中 = %+v", sub)
	}
	// 同页没有其它控件时，不给"相近控件"提示（宁缺勿错）。
	if strings.Contains(sub.Note, "同页相近的控件") {
		t.Errorf("没有相近控件却给了提示: %q", sub.Note)
	}

	// ResolveColumns 的空查询/标签子串/无页面列分支。
	doc := &LinkDoc{Fields: []LinkField{
		{PageID: "1", Column: "xmdduc001", Label: "刻号"},
		{PageID: "1", Column: "xmdduc002", Label: "料号"},
	}}
	if cols, by, lm := ResolveColumns(doc, "  "); cols != nil || by != "" || lm != nil {
		t.Errorf("空查询 = %v/%q/%v", cols, by, lm)
	}
	if cols, by, _ := ResolveColumns(doc, "号"); by != "label" || len(cols) != 2 {
		t.Errorf("标签子串 = %v/%s", cols, by)
	}
	if got := labelOfColumn(doc, "xmdduc001"); got != "刻号" {
		t.Errorf("labelOfColumn = %q", got)
	}
	if got := labelOfColumn(nil, "x"); got != "" {
		t.Errorf("nil doc labelOfColumn = %q", got)
	}
	if got := labelOfColumn(doc, ""); got != "" {
		t.Errorf("空列 labelOfColumn = %q", got)
	}
	if columnOfPoint("no-page") != "" {
		t.Error("非 page 点名不该有列名")
	}

	// 没有切片的材料：grep 命中的 SliceCode 留空（不能凭空给一个）。
	noSlice := newMinimalMaterials(t, "axmt531")
	noSlice.mapIndex()
	gres, err := Grep(GrepRequest{LinkRequest: noSlice.LinkRequest(), Pattern: "#add-point"})
	if err != nil {
		t.Fatalf("Grep: %v", err)
	}
	if len(gres.Hits) == 0 || gres.Hits[0].SliceCode != "" {
		t.Errorf("无切片命中的归属 = %+v", gres.Hits)
	}

	// 没有 .tap 却要查 tap 范围：报材料不成套（不是 0 条）。
	notap := newMaterials(t, matOpts{Prog: "axmt532", NoTap: true})
	if _, err := Grep(GrepRequest{LinkRequest: notap.LinkRequest(), Pattern: "x", Scope: "tap"}); !errors.Is(err, ErrMaterialIncomplete) {
		t.Errorf("tap 范围缺 .tap err = %v", err)
	}

	// 非法程序名：四条查询都要在打开材料前挡住。
	bad := LinkRequest{CfgPath: newMinimalMaterials(t, "axmt533").CfgPath, Env: "e", Prog: "BAD"}
	if _, err := Show(ShowRequest{LinkRequest: bad, Query: "x"}); !errors.Is(err, ErrInvalidProg) {
		t.Errorf("Show 非法程序 err = %v", err)
	}
	if _, _, err := Fields(FieldsRequest{LinkRequest: bad}); !errors.Is(err, ErrInvalidProg) {
		t.Errorf("Fields 非法程序 err = %v", err)
	}
	if _, err := At(AtRequest{LinkRequest: bad, Line: 1}); !errors.Is(err, ErrInvalidProg) {
		t.Errorf("At 非法程序 err = %v", err)
	}
	if _, err := Grep(GrepRequest{LinkRequest: bad, Pattern: "x"}); !errors.Is(err, ErrInvalidProg) {
		t.Errorf("Grep 非法程序 err = %v", err)
	}
}
