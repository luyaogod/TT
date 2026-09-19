package tzs

// 需要**真引擎**的测试：默认 `go test ./...` 全部跳过（会打印跳过原因）。
//
// 为什么要这个开关：真引擎要 Boot 一次 WPF 设计器（约 1 s）、要装一份设计器、
// 要一个 .tzs 工作区 —— 这三样都不是 CI 或别人的机器上一定有的东西。
// 而把它们写进默认测试，`go test ./...` 就会变成「时好时坏」，
// 那种假失败比不跑更糟（corpus_test.go 的 requireDeep 是同一个理由）。
//
// 开启方式（三样都得给，工作区**绝不留缺省** —— 引擎自己的缺省是一个真实客户目录）：
//
//	$env:TTZS_E2E="1"
//	$env:TTZS_EXE="D:\...\engine\out\tzs-server.exe"
//	$env:TTZS_WS="D:\t100_wrok_dir\<模块>"        # 一个真的 .tzs 工作区
//	$env:TTZS_INSTALL="D:\APPS\T100设计器_1.0.0.251_免安装"   # 可省：引擎有内置缺省
//	go test ./internal/dev/tzs/ -run 'TestE2E' -timeout 5m -v

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// requireE2E 把「对真引擎」的测试设为显式开关。
func requireE2E(t *testing.T) (exe, ws, install string) {
	t.Helper()
	if os.Getenv("TTZS_E2E") == "" {
		t.Skip("需要真引擎：设 TTZS_E2E=1 + TTZS_EXE + TTZS_WS（可选 TTZS_INSTALL），见本文件顶部注释")
	}
	exe = os.Getenv("TTZS_EXE")
	if exe == "" {
		// 从包目录出发的相对缺省（本仓库里引擎的产物就在那儿），只是省事，不是保证。
		exe = filepath.Join("..", "..", "..", "engine", "out", "tzs-server.exe")
	}
	if _, err := os.Stat(exe); err != nil {
		t.Skipf("找不到引擎 exe（TTZS_EXE=%s）：%v", exe, err)
	}
	ws = os.Getenv("TTZS_WS")
	if strings.TrimSpace(ws) == "" {
		t.Skip("没给 TTZS_WS。工作区绝不替你选：引擎自己的缺省是一个真实客户目录")
	}
	return exe, ws, os.Getenv("TTZS_INSTALL")
}

// TestE2EManifest 用真 manifest 校验我们的解析层：两个「不 Boot」的开关，
// 所以这条既便宜又最能证明「我们读的字段名没写错」。
func TestE2EManifest(t *testing.T) {
	exe, ws, _ := requireE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	m, err := FetchManifest(ctx, exe)
	if err != nil {
		t.Fatalf("--manifest 失败: %v", err)
	}
	// 契约说 ~49 个函数。给个下界即可（引擎会先声明全部、再慢慢补实现）。
	if len(m.Fns) < 40 {
		t.Errorf("函数表只有 %d 个，看着不像完整表", len(m.Fns))
	}
	for _, fn := range []string{"open", "save", "close", "list_open", "form_tree", "set_spec_attr", "validate"} {
		if m.ByName(fn) == nil {
			t.Errorf("函数表里该有 %s", fn)
		}
	}
	// stop 不在 manifest 里（传输级）。这条是契约里最容易写错的一条。
	if m.ByName(FnStop) != nil {
		t.Errorf("%s 不该在 manifest 里", FnStop)
	}
	// 真 manifest 里的类型必须全部被我们认识 —— 尤其是契约清单里漏掉的 handle。
	for _, f := range m.Fns {
		for _, a := range f.Args {
			if !KnownType(a.Type) {
				t.Errorf("%s.%s 的类型 %q 我们不认识（引擎加了新类型？）", f.Name, a.Name, a.Type)
			}
		}
	}
	// 实测过的两处：from:* 真的只在 attr 上；add_action 的 type 是普通 string。
	if a := m.ByName("set_spec_attr").Param("attr"); a == nil || a.From != "spec:<kind>" {
		t.Errorf("set_spec_attr.attr 该带 from=spec:<kind>：%+v", a)
	}
	if ty := m.ByName("add_action").Param("type"); ty == nil || ty.From != "" {
		t.Errorf("add_action.type 该是普通 string（无 from）：%+v", ty)
	}
	if m.ByName("list_open").NeedsHandle {
		t.Errorf("list_open 必须 needsHandle=false —— 就绪握手就靠它")
	}

	name, err := PipeName(ctx, exe, ws)
	if err != nil {
		t.Fatalf("--pipe-name 失败: %v", err)
	}
	if !regexp.MustCompile(`^tzs-cli-[0-9a-f]{8}-[0-9a-f]{8}$`).MatchString(name) {
		t.Errorf("管道名形状不对：%q", name)
	}
	// 换工作区名字就该换管道名（工作区是 hash 的输入之一）。
	other, err := PipeName(ctx, exe, filepath.Join("D:", "绝对不存在的模块"))
	if err != nil {
		t.Fatalf("--pipe-name（另一个工作区）失败: %v", err)
	}
	if other == name {
		t.Errorf("不同工作区该得到不同管道名（%q）", name)
	}
	t.Logf("真 manifest：%d 个函数；工作区 %s → 管道 %s", len(m.Fns), ws, name)
}

// TestE2EEnsureCallStop 走一遍完整生命周期：spawn → 就绪 → 一次真调用 → stop。
//
// 用 list_open 当那次调用：它是 needsHandle=false 的只读函数，不需要打开任何包，
// 所以这条测试不会改动工作区里的任何东西。
func TestE2EEnsureCallStop(t *testing.T) {
	exe, ws, install := requireE2E(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	o := Options{Exe: exe, Workspace: ws, InstallDir: install}

	// 先确保干净：上一次跑留下的守护进程不该影响这次。
	_ = Stop(ctx, o)

	pipe, err := Ensure(ctx, o)
	if err != nil {
		t.Fatalf("Ensure 失败: %v", err)
	}
	info, err := LookupDaemon(ctx, o)
	if err != nil {
		t.Fatalf("LookupDaemon 失败: %v", err)
	}
	if !info.Running {
		t.Fatalf("Ensure 之后该有一个应答的守护进程：%+v", info)
	}
	if info.Pipe != pipe {
		t.Errorf("Ensure 返回的管道名与快照不一致：%q vs %q", pipe, info.Pipe)
	}

	// 就绪握手已经证明 list_open 能答；这里再走一遍完整的 Call 路径（带参数定型）。
	args, err := BuildArgs(mustManifestFromEngine(t, exe), "list_open", nil)
	if err != nil {
		t.Fatalf("list_open 的 args 该是空的：%v", err)
	}
	r, err := Call(ctx, DefaultDialer(), pipe, 1, "list_open", args, DefaultTimeout)
	if err != nil {
		t.Fatalf("list_open 调用失败: %v", err)
	}
	if !r.OK {
		t.Fatalf("list_open 该成功：%+v", r.Error)
	}
	if got := ExitCode(r, nil); got != ExitOK {
		t.Errorf("该退 0，得 %d", got)
	}

	// 再 Ensure 一次：必须**复用**（不重启），否则每次调用都要重 Boot 一次。
	pipe2, err := Ensure(ctx, o)
	if err != nil {
		t.Fatalf("第二次 Ensure 失败: %v", err)
	}
	if pipe2 != pipe {
		t.Errorf("同一个工作区该复用同一个管道：%q → %q", pipe, pipe2)
	}

	// stop：幂等，且绝不 spawn。
	if err := Stop(ctx, o); err != nil {
		t.Fatalf("Stop 失败: %v", err)
	}
	if err := Stop(ctx, o); err != nil {
		t.Fatalf("第二次 Stop 该是幂等的：%v", err)
	}
	if info, _ := LookupDaemon(ctx, o); info != nil && info.Running {
		t.Errorf("stop 之后不该还有应答的守护进程：%+v", info)
	}
}

// mustManifestFromEngine 取真 manifest（E2E 里已经证明能取到）。
func mustManifestFromEngine(t *testing.T, exe string) *Manifest {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	m, err := FetchManifest(ctx, exe)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
