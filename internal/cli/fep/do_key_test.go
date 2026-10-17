package fep

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"tt/internal/cli/common"
	"tt/internal/fep"
)

// 契约 A28 的回归钉：**无目标动作**（`do key`，KeyEvent 不带 idRef）不得被
// "目标 / --field 互斥"检查拦；而 `act` / `type` 的互斥一条都不许松。
//
// 纪律（切片边界）：这一组**不连真机、不起会话、不碰服务器**。注入口是本地假件
// （回环 TCP，按 internal/fep/do.go 的行协议应答），界面从**合成录制件**重建 ——
// 与 `tt fep` 的其它子命令同一份输入，没有第二条真相。

// cliDoAUI 是合成读侧夹具：与 internal/fep 的 do 测试同一形状（Label + ButtonEdit
// 挂一个 FormField），这样"按名字点名"走的是既有解析路径，不是特例。
const cliDoAUI = `om 1 {{an 0 Window 4097 {{name "w_axmt500"}} {{` +
	`{an 4097 ToolBarItem 5160 {{text "录入"}{actionActive "1"}} {}} ` +
	`{an 4097 Label 270 {{comment "cmt_xmda004"}{text "客户编号"}} {}} ` +
	`{an 4097 ButtonEdit 272 {{comment "cmt_xmda004"}} {{an 272 FormField 271 {{value ""}} {}}} {}} ` +
	`{an 4097 ToolBarItem 5156 {{text "确定"}{actionActive "0"}} {}}}}}`

// fakeInjector 是**假注入口**：只认两条请求 ——
//
//	status → 一条活着的连接（id 0）
//	@0 <事件体> → 记下来，回 ACK
//
// 记下来的就是"这一步到底发了什么"，测试断言的是它以 KeyEvent 形态出现。
type fakeInjector struct {
	addr string
	mu   sync.Mutex
	got  []string
}

func newFakeInjector(t *testing.T) *fakeInjector {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("假注入口起不来：%v", err)
	}
	t.Cleanup(func() { ln.Close() })
	f := &fakeInjector{addr: ln.Addr().String()}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go f.serve(c)
		}
	}()
	return f
}

func (f *fakeInjector) serve(c net.Conn) {
	defer c.Close()
	sc := bufio.NewScanner(c)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		var reply string
		switch {
		case line == "status":
			reply = `{"conns":[{"id":0,"seq":1,"alive":true}]}`
		case strings.HasPrefix(line, "@0 "):
			f.mu.Lock()
			f.got = append(f.got, strings.TrimPrefix(line, "@0 "))
			f.mu.Unlock()
			reply = "ACK"
		default:
			reply = "ERR 假注入口不认这条"
		}
		fmt.Fprintln(c, reply)
	}
}

func (f *fakeInjector) events() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.got...)
}

// awaitEvents 等到至少 n 条事件（假件是异步收的，别用 sleep 猜）。
func (f *fakeInjector) awaitEvents(t *testing.T, n int) []string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if ev := f.events(); len(ev) >= n {
			return ev
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("假注入口只收到 %d 条事件（想要 %d 条）：%v", len(f.events()), n, f.events())
	return nil
}

// cliDoEnv 备一个**假会话环境**（配置落点 + 合成录制件 + 全局开关复位），
// 返回数据目录；登记由 cliDoSession 视需要写 —— "没有会话"也要能测。
func cliDoEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	if err := os.WriteFile(cfg, []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("写临时配置：%v", err)
	}
	// TT_CONFIG 优先级最高（见 config.ResolvePath），配置路径 = 数据目录 = 这个临时目录
	t.Setenv("TT_CONFIG", cfg)
	oldCfg := common.ConfigPath
	common.ConfigPath = cfg
	oldWait, oldDry, oldConn, oldField, oldCommit := doWait, doDryRun, doConn, doField, doCommitOff
	doWait, doDryRun, doConn, doField, doCommitOff = -1, false, 0, "", false
	t.Cleanup(func() {
		common.ConfigPath = oldCfg
		doWait, doDryRun, doConn, doField, doCommitOff = oldWait, oldDry, oldConn, oldField, oldCommit
	})

	recDir := filepath.Join(dir, "rec")
	if err := os.MkdirAll(recDir, 0o700); err != nil {
		t.Fatalf("建录制目录：%v", err)
	}
	if err := os.WriteFile(filepath.Join(recDir, "t-0-rt2gdc.log"), []byte(cliDoAUI+"\n"), 0o600); err != nil {
		t.Fatalf("写合成录制件：%v", err)
	}
	return dir
}

// cliDoSession 在假环境上再写一份会话登记（注入口指向假件）。
func cliDoSession(t *testing.T, dir string, inj *fakeInjector) {
	t.Helper()
	_, portStr, err := net.SplitHostPort(inj.addr)
	if err != nil {
		t.Fatalf("假注入口地址：%v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("假注入口端口：%v", err)
	}
	reg := fep.NewRegistry(dir)
	if serr := reg.Save(&fep.Session{
		Prog: "axmt500_wf", Tag: "t", RecordDir: filepath.Join(dir, "rec"), InjectPort: port,
	}); serr != nil {
		t.Fatalf("写会话登记：%v", serr)
	}
}

// TestKeyNeedsNoTarget：`key` 是**无目标动作**（契约 A28）。
//
// 这是真机回归的那一条：`tt fep do key Tab` 曾报 `USAGE 没给目标`。
func TestKeyNeedsNoTarget(t *testing.T) {
	if needsTarget("key") {
		t.Fatal("key 没有目标对象（KeyEvent 不带 idRef）：needsTarget 必须是 false")
	}
	for _, kind := range []string{"click", "input", "pick"} {
		if !needsTarget(kind) {
			t.Fatalf("%s 是要目标的动作：needsTarget 必须是 true", kind)
		}
	}
	if err := checkFieldExclusive("key", "", ""); err != nil {
		t.Fatalf("key 没给目标也不该报错：%v", err)
	}
}

// TestDoKeyTabReachesTheInjector：`do key Tab` 走通到"发了 KeyEvent"。
func TestDoKeyTabReachesTheInjector(t *testing.T) {
	dir := cliDoEnv(t)
	inj := newFakeInjector(t)
	cliDoSession(t, dir, inj)

	if err := doKeyCmd.RunE(doKeyCmd, []string{"Tab"}); err != nil {
		t.Fatalf("`do key Tab` 不该报错（被目标互斥拦了就是 A28 回归）：%v", err)
	}
	ev := inj.awaitEvents(t, 1)[0]
	if !strings.Contains(ev, `KeyEvent 0{{keyName "Tab"}}`) {
		t.Fatalf("发出去的不是 KeyEvent Tab：%q", ev)
	}
	if strings.Contains(ev, "idRef") {
		t.Fatalf("KeyEvent 不该带 idRef：%q", ev)
	}
}

// TestDoKeyWithoutSessionSaysNoSession：没有会话就该说没有会话（3），
// 而不是把一条**合法命令**读成用法错（2） —— 排障方向完全相反。
func TestDoKeyWithoutSessionSaysNoSession(t *testing.T) {
	cliDoEnv(t) // 故意不写登记

	err := doKeyCmd.RunE(doKeyCmd, []string{"Tab"})
	if err == nil {
		t.Fatal("没有会话时必须失败")
	}
	if got := exitOf(t, err); got != exitNoSession {
		t.Fatalf("退出码 = %d，想要 %d（3 = 没有会话；2 = 用法错）", got, exitNoSession)
	}
	if strings.Contains(err.Error(), "没给目标") {
		t.Fatalf("key 没有目标对象，不该被目标检查拦：%v", err)
	}
}

// TestActAndTypeStillMutuallyExclusive：互斥一条都不许松（A28 只给 key 开的口子）。
func TestActAndTypeStillMutuallyExclusive(t *testing.T) {
	cliDoEnv(t) // 用法错先于"有没有会话"判定，所以不需要登记

	for _, kind := range []string{"click", "input"} {
		if err := checkFieldExclusive(kind, "客户编号", "b_xmda004"); err == nil {
			t.Fatalf("%s：目标与 --field 同时给必须报用法错", kind)
		} else if got := exitOf(t, err); got != exitUsage {
			t.Fatalf("%s：退出码 = %d，想要 %d", kind, got, exitUsage)
		}
		if err := checkFieldExclusive(kind, "", ""); err == nil {
			t.Fatalf("%s：两个都没给必须报用法错", kind)
		} else if got := exitOf(t, err); got != exitUsage {
			t.Fatalf("%s：退出码 = %d，想要 %d", kind, got, exitUsage)
		}
	}

	// 走真正的入口（regression 就发生在 runDo 的第一次调用上）
	doField = "b_xmda004"
	defer func() { doField = "" }()
	if err := runDo(doActCmd, fep.DoConfig{Kind: "click", Target: "客户编号"}); err == nil {
		t.Fatal("act：目标 + --field 必须退 2，不能被后面的「没会话」盖过去")
	} else if got := exitOf(t, err); got != exitUsage {
		t.Fatalf("act：退出码 = %d，想要 %d", got, exitUsage)
	}
	if err := runDo(doTypeCmd, fep.DoConfig{Kind: "input", Target: "客户编号", Value: "A50"}); err == nil {
		t.Fatal("type：目标 + --field 必须退 2")
	} else if got := exitOf(t, err); got != exitUsage {
		t.Fatalf("type：退出码 = %d，想要 %d", got, exitUsage)
	}
}

// TestDoKeyRejectsFieldFlagExplicitly：`key + --field` 的口径是**拒绝**（不静默忽略）。
//
// 为什么选拒绝：`--field` 的含义是"按字段身份点名目标"，而 key 根本没有目标可点；
// 忽略它等于替人猜"你大概想按键"——猜错那次会安静地按错键（契约 A9/A11 的同一原则：
// 拿不准就拒，并说清怎么改）。挂上这个开关而不是放任 cobra 回 `unknown flag`，
// 也是为了让报错是**我们的话**（讲清为什么），退出码仍是用法错 2。
func TestDoKeyRejectsFieldFlagExplicitly(t *testing.T) {
	if doKeyCmd.Flags().Lookup("field") == nil {
		t.Fatal("--field 必须挂在 key 上：拒绝要讲清理由，而不是一句 unknown flag")
	}
	if err := checkFieldExclusive("key", "", "b_xmda004"); err == nil {
		t.Fatal("key + --field 必须报用法错（口径：拒绝）")
	} else {
		if got := exitOf(t, err); got != exitUsage {
			t.Fatalf("退出码 = %d，想要 %d", got, exitUsage)
		}
		if msg := err.Error(); !strings.Contains(msg, "--field") || !strings.Contains(msg, "目标") {
			t.Fatalf("报错要说清「key 没有目标、不能用 --field」：%q", msg)
		}
	}

	// 走真正的入口：参数校验（exactArgs(1)）过了之后，runDo 必须当场拒
	cliDoEnv(t)
	doField = "b_xmda004"
	defer func() { doField = "" }()
	err := doKeyCmd.RunE(doKeyCmd, []string{"Tab"})
	if err == nil {
		t.Fatal("key + --field 必须退 2，不能静默忽略也不能真发")
	}
	if got := exitOf(t, err); got != exitUsage {
		t.Fatalf("退出码 = %d，想要 %d", got, exitUsage)
	}
}

// TestDoTypeCommitsWithTabByDefault：`type` 写完栏位后默认发 KeyEvent Tab 离开 ——
// AFTER FIELD（校验 / 自动带值）只在光标离开那一刻跑，不离开就看不见系统的动作。
func TestDoTypeCommitsWithTabByDefault(t *testing.T) {
	dir := cliDoEnv(t)
	inj := newFakeInjector(t)
	cliDoSession(t, dir, inj)

	if err := doTypeCmd.RunE(doTypeCmd, []string{"客户编号", "CNJ-A50"}); err != nil {
		t.Fatalf("do type：%v", err)
	}
	ev := inj.awaitEvents(t, 2)
	if !strings.Contains(ev[0], "ConfigureEvent") {
		t.Fatalf("第一条该是写值：%q", ev[0])
	}
	if !strings.Contains(ev[1], `KeyEvent 0{{keyName "Tab"}}`) {
		t.Fatalf("第二条该是离开栏位的 KeyEvent Tab：%q", ev[1])
	}
}

// TestDoTypeNoCommitSendsNoKeyEvent：`--no-commit` 时不许再发那个 Tab。
func TestDoTypeNoCommitSendsNoKeyEvent(t *testing.T) {
	dir := cliDoEnv(t)
	inj := newFakeInjector(t)
	cliDoSession(t, dir, inj)
	doCommitOff = true // = --no-commit

	if err := doTypeCmd.RunE(doTypeCmd, []string{"客户编号", "CNJ-A50"}); err != nil {
		t.Fatalf("do type --no-commit：%v", err)
	}
	inj.awaitEvents(t, 1)
	time.Sleep(700 * time.Millisecond) // 给"本该发的 Tab"留出足够机会（它在 400ms 后发）
	if ev := inj.events(); len(ev) != 1 {
		t.Fatalf("--no-commit 只该发写值这一条，实收 %d 条：%v", len(ev), ev)
	}
}
