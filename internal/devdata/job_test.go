package devdata

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestManagerSingleFlight 同一个 (kind,env,prog) 不许并发；不同程序可以并行。
func TestManagerSingleFlight(t *testing.T) {
	m := NewManager()
	block := make(chan struct{})
	started := make(chan struct{})
	if _, err := m.Start(KindMirror, "e1", "", func(ctx context.Context, report ProgressFunc) (any, error) {
		close(started)
		<-block
		return nil, nil
	}); err != nil {
		t.Fatalf("第一次 Start: %v", err)
	}
	<-started
	if _, err := m.Start(KindMirror, "e1", "", func(context.Context, ProgressFunc) (any, error) {
		return nil, nil
	}); !errors.Is(err, ErrBusy) {
		t.Fatalf("同类任务在跑应报 ErrBusy，得到 %v", err)
	}
	// 不同环境 / 不同种类不受影响
	if _, err := m.Start(KindMirror, "e2", "", func(context.Context, ProgressFunc) (any, error) { return nil, nil }); err != nil {
		t.Errorf("别的环境应能开: %v", err)
	}
	if _, err := m.Start(KindFetch, "e1", "axmt500_wf", func(context.Context, ProgressFunc) (any, error) { return nil, nil }); err != nil {
		t.Errorf("按程序拉取是另一个维度，应能开: %v", err)
	}
	// 同一程序重复拉取也不行
	if _, err := m.Start(KindFetch, "e1", "axmt500_wf", func(context.Context, ProgressFunc) (any, error) { return nil, nil }); !errors.Is(err, ErrBusy) {
		t.Errorf("同一程序重复拉取应 ErrBusy，得到 %v", err)
	}
	close(block)
	j := m.Wait(KindMirror, "e1", "")
	if j.Running || j.Phase != "done" {
		t.Errorf("任务应收尾成 done: %+v", j)
	}
}

// TestManagerStop 取消 → Cancelled（不是 Error），且停止端点在没有任务时报 ErrNoJob。
func TestManagerStop(t *testing.T) {
	m := NewManager()
	started := make(chan struct{})
	m.Start(KindMirror, "e1", "", func(ctx context.Context, report ProgressFunc) (any, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	<-started
	if err := m.Stop(KindMirror, "e1", ""); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	j := m.Wait(KindMirror, "e1", "")
	if !j.Cancelled || j.Phase != "cancelled" || j.Error != "" {
		t.Fatalf("取消应标成 cancelled 且没有 error: %+v", j)
	}
	if err := m.Stop(KindMirror, "e1", ""); !errors.Is(err, ErrNoJob) {
		t.Fatalf("没有任务时 Stop 应报 ErrNoJob，得到 %v", err)
	}
}

// TestManagerProgress 进度字段要能落到快照上（前端轮询看到的就是这些）。
func TestManagerProgress(t *testing.T) {
	m := NewManager()
	started := make(chan struct{})
	release := make(chan struct{})
	m.Start(KindFetch, "e1", "axmt500_wf", func(ctx context.Context, report ProgressFunc) (any, error) {
		close(started)
		report(Progress{Phase: "fetch", Message: "下载 tab", Done: 3, Total: 10, CurrentFile: "dzx/a.tab"})
		<-release
		return "ok", nil
	})
	<-started
	// 等一拍让回调落进快照（Start 的 report 在 run 的 goroutine 里同步执行）
	deadline := time.Now().Add(time.Second)
	for {
		j, _ := m.Snapshot(KindFetch, "e1", "axmt500_wf")
		if j.Done == 3 && j.Total == 10 && j.CurrentFile == "dzx/a.tab" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("进度没落到快照: %+v", j)
		}
		time.Sleep(5 * time.Millisecond)
	}
	close(release)
	j := m.Wait(KindFetch, "e1", "axmt500_wf")
	if j.Result != "ok" || j.Done != j.Total {
		t.Fatalf("结束后应带 result 且 Done=Total: %+v", j)
	}
	if got := m.Running("e1"); len(got) != 0 {
		t.Errorf("结束后 Running 应为空: %+v", got)
	}
}

// TestManagerError 失败 → Phase=error 且 Error 带原文。
func TestManagerError(t *testing.T) {
	m := NewManager()
	m.Start(KindMirror, "e1", "", func(context.Context, ProgressFunc) (any, error) {
		return nil, errors.New("模拟失败")
	})
	j := m.Wait(KindMirror, "e1", "")
	if j.Phase != "error" || j.Error != "模拟失败" || j.Cancelled {
		t.Fatalf("失败应标成 error: %+v", j)
	}
}

// TestStatusEmptyThenFilled status 在"没拉过"与"拉过"两种状态下的形状。
func TestStatusEmptyThenFilled(t *testing.T) {
	cfgPath, tables, files := newMirrorFixture(t)
	ctx := context.Background()

	st, err := Status(ctx, StatusRequest{CfgPath: cfgPath, Env: "福联测试tst"})
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if st.Present || st.Dict.Present || st.Mdl.Present || len(st.Progs) != 0 {
		t.Fatalf("没拉过时应是空状态: %+v", st)
	}
	if st.Dict.Path == "" || st.Mdl.Dir == "" {
		t.Errorf("空状态也要给出预期落点: %+v", st)
	}

	// 拉一次共用资料
	if _, err := Mirror(ctx, MirrorRequest{CfgPath: cfgPath, Env: "福联测试tst", Tables: tables, Files: files}, nil); err != nil {
		t.Fatal(err)
	}
	st, err = Status(ctx, StatusRequest{CfgPath: cfgPath, Env: "福联测试tst"})
	if err != nil {
		t.Fatal(err)
	}
	if !st.Present || !st.Dict.Present || !st.Mdl.Present {
		t.Fatalf("拉过后应三项都在: %+v", st)
	}
	if st.Dict.Tables != 51 || st.Dict.Skipped != 1 {
		t.Errorf("dict 表数/跳过数 = %d/%d，应为 51/1", st.Dict.Tables, st.Dict.Skipped)
	}
	if st.Mdl.Files != 3 || st.Mdl.TemplateCount != 2 {
		t.Errorf("mdl 文件/模板数 = %d/%d，应为 3/2", st.Mdl.Files, st.Mdl.TemplateCount)
	}
	if st.Mdl.TemplateVersions["a05"] != 3 {
		t.Errorf("模板版本 = %v", st.Mdl.TemplateVersions)
	}
	if st.Dict.SyncedAt == "" {
		t.Error("应有同步时间")
	}
	if AgeText(st.Dict.SyncedAt) == "" {
		t.Error("AgeText 应能渲染")
	}
}

// TestStatusListsProgsAndStale 磁盘上的程序目录都要出现在列表里，本地被改过要标 stale。
func TestStatusListsProgsAndStale(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	ctx := context.Background()
	if _, err := Fetch(ctx, FetchRequest{CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files}, nil); err != nil {
		t.Fatal(err)
	}
	st, err := Status(ctx, StatusRequest{CfgPath: cfgPath, Env: "福联测试tst"})
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Progs) != 1 || st.Progs[0].Prog != "axmt500_wf" {
		t.Fatalf("程序列表 = %+v", st.Progs)
	}
	if st.Progs[0].Stale {
		t.Errorf("刚拉完不该 stale: %+v", st.Progs[0])
	}
	if st.Progs[0].Module != "axm" || st.Progs[0].Files == 0 {
		t.Errorf("程序状态缺字段: %+v", st.Progs[0])
	}

	// 本地改一个文件 → 下一次 status 必须报 stale（本地 sha 不一致，不需要联网）
	progDir, _ := ProgDir(cfgPath, "福联测试tst", "axmt500_wf")
	if err := os.WriteFile(filepath.Join(progDir, "dzx", "axmt500_wf.tab"), []byte("改了"), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err = Status(ctx, StatusRequest{CfgPath: cfgPath, Env: "福联测试tst"})
	if err != nil {
		t.Fatal(err)
	}
	if !st.Progs[0].Stale {
		t.Fatal("本地改过应报 stale")
	}
	if st.Progs[0].StaleReasons[0].Code != "sha-mismatch" {
		t.Errorf("原因应为 sha-mismatch: %+v", st.Progs[0].StaleReasons)
	}
	if st.Progs[0].Probed {
		t.Error("默认不该 probe（不出网）")
	}

	// Check 强制 probe：用注入的假来源比对服务器
	prog, err := Check(ctx, StatusRequest{CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !prog.Probed {
		t.Error("Check 应 probe")
	}
}

// TestCheckNotFetched 没拉过的程序查 check → ErrProgNotFound（404 语义）。
func TestCheckNotFetched(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	_, err := Check(context.Background(), StatusRequest{
		CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files})
	if !errors.Is(err, ErrProgNotFound) {
		t.Fatalf("应报 ErrProgNotFound，得到 %v", err)
	}
}

// TestCleanDryRunThenExecute 清理默认只看不删；执行后目录消失。
func TestCleanDryRunThenExecute(t *testing.T) {
	cfgPath, files := newFetchFixture(t)
	if _, err := Fetch(context.Background(), FetchRequest{CfgPath: cfgPath, Env: "福联测试tst", Prog: "axmt500_wf", Files: files}, nil); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanClean(cfgPath, "福联测试tst", "axmt500_wf")
	if err != nil {
		t.Fatalf("PlanClean: %v", err)
	}
	if plan.Empty() || plan.Files == 0 || plan.Bytes == 0 {
		t.Fatalf("计划应有内容: %+v", plan)
	}
	progDir, _ := ProgDir(cfgPath, "福联测试tst", "axmt500_wf")
	if _, err := os.Stat(progDir); err != nil {
		t.Fatal("dry-run 不该删东西")
	}
	if err := ExecuteClean(plan); err != nil {
		t.Fatalf("ExecuteClean: %v", err)
	}
	if _, err := os.Stat(progDir); err == nil {
		t.Fatal("执行后程序目录应消失")
	}
	// 再删一次 → ErrProgNotFound
	if _, err := PlanClean(cfgPath, "福联测试tst", "axmt500_wf"); !errors.Is(err, ErrProgNotFound) {
		t.Fatalf("已删除的程序应报 ErrProgNotFound，得到 %v", err)
	}
}

// TestCleanEnvLevel 只给环境 = 删整棵环境目录。
func TestCleanEnvLevel(t *testing.T) {
	cfgPath, tables, files := newMirrorFixture(t)
	if _, err := Mirror(context.Background(), MirrorRequest{CfgPath: cfgPath, Env: "福联测试tst", Tables: tables, Files: files}, nil); err != nil {
		t.Fatal(err)
	}
	envDir, _ := EnvDir(cfgPath, "福联测试tst")
	plan, err := PlanClean(cfgPath, "福联测试tst", "")
	if err != nil {
		t.Fatalf("PlanClean: %v", err)
	}
	if len(plan.Items) != 1 || plan.Items[0].Kind != "env" {
		t.Fatalf("计划 = %+v", plan)
	}
	if err := ExecuteClean(plan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(envDir); err == nil {
		t.Fatal("环境目录应消失")
	}
	// 环境没了以后再计划 → ErrProgNotFound（"还没有任何开发资料"）
	if _, err := PlanClean(cfgPath, "福联测试tst", ""); !errors.Is(err, ErrProgNotFound) {
		t.Fatalf("应报 ErrProgNotFound，得到 %v", err)
	}
}
