package fep

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// freePort asks the kernel for a free port. Small race, fine for tests.
func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func TestOpenRejectsBusyPortWithoutResidue(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("隔离实例只在 Windows 上有意义：这条测的是端口被占时不留残余")
	}
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	busyPort := busy.Addr().(*net.TCPAddr).Port

	dir := t.TempDir()
	_, err = Open(context.Background(), OpenConfig{
		Prog:      "axmt500_wf",
		DataDir:   dir,
		GDCExe:    filepath.Join(dir, "not-there.exe"),
		GDCPort:   freePort(t),
		ProxyPort: busyPort,
		// 端口体检在起实例之前，所以这里不该留下任何东西
		InjectPort: freePort(t),
		NoJob:      true,
	}, nil)
	if !errors.Is(err, ErrPortBusy) {
		t.Fatalf("err = %v, want ErrPortBusy", err)
	}
	if _, serr := os.Stat(filepath.Join(dir, SessionFileName)); !os.IsNotExist(serr) {
		t.Fatal("失败路径不该留下会话登记")
	}
}

func TestOpenReportsMissingInstanceAndKeepsNothingBehind(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("隔离实例只在 Windows 上有意义：这条测的是实例 exe 不存在时不留残余")
	}
	dir := t.TempDir()
	gdcPort, proxyPort, injectPort := freePort(t), freePort(t), freePort(t)
	_, err := Open(context.Background(), OpenConfig{
		Prog:       "axmt500_wf",
		DataDir:    dir,
		GDCExe:     filepath.Join(dir, "not-there.exe"),
		GDCPort:    gdcPort,
		ProxyPort:  proxyPort,
		InjectPort: injectPort,
		NoJob:      true,
	}, nil)
	if err == nil {
		t.Fatal("起不来的隔离实例必须报错")
	}
	if _, serr := os.Stat(filepath.Join(dir, SessionFileName)); !os.IsNotExist(serr) {
		t.Fatal("失败路径不该留下会话登记")
	}
	// 三个端口都必须还空着
	for _, p := range []int{gdcPort, proxyPort, injectPort} {
		if perr := portFree(p); perr != nil {
			t.Fatalf("端口 %d 被留下了: %v", p, perr)
		}
	}
}

func TestCheckElevationWithZeroPidIsSelfComparison(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("TokenElevation（提权检查）只在 Windows 上有意义")
	}
	e, err := CheckElevation(0)
	if err != nil {
		t.Fatalf("CheckElevation(0): %v", err)
	}
	if !e.OK {
		t.Fatalf("自己跟自己比必须一致: %+v", e)
	}
}

func TestWaitPortFailsFastWhenNothingListens(t *testing.T) {
	port := freePort(t)
	start := time.Now()
	err := waitPort(port, 700*time.Millisecond)
	if err == nil {
		t.Fatal("没人在监听的端口不能算等到")
	}
	if elapsed := time.Since(start); elapsed < 500*time.Millisecond {
		t.Fatalf("等太短了: %s", elapsed)
	}
	_ = strconv.Itoa(port)
}

// 不给 ENT 时**一个字节都不许碰 TOPENT**：登录 shell 自己的 $TOPENT 就是权威，
// 擅自 unset 或给空值会把作业挪到别的企业上，而症状只是"查不到数据"。
func TestJobLaunchCmdLeavesTopentAloneWithoutEnt(t *testing.T) {
	got, err := jobLaunchCmd(PathRunnerRR, "10.0.0.1", 1200, "axmt500_wf", "")
	if err != nil {
		t.Fatalf("合法参数不该报错: %v", err)
	}
	if strings.Contains(got, "TOPENT") {
		t.Fatalf("没给 ENT 时不许出现 TOPENT: %q", got)
	}
	if got != "export FGLSERVER=10.0.0.1:1200\r\nr.r axmt500_wf\r\n" {
		t.Fatalf("没给 ENT 时命令应逐字不变, got %q", got)
	}
}

// 给了 ENT 时顺序不能反：TOPENT → FGLSERVER → r.r。
// 反了作业会带着上一轮的企业跑起来，而结果看起来完全正常。
func TestJobLaunchCmdSendsTopentBeforeLaunch(t *testing.T) {
	got, err := jobLaunchCmd(PathRunnerRR, "10.0.0.1", 1200, "axmt500_wf", "99")
	if err != nil {
		t.Fatalf("合法参数不该报错: %v", err)
	}
	iEnt := strings.Index(got, "export TOPENT='99'")
	iFgl := strings.Index(got, "export FGLSERVER=10.0.0.1:1200")
	iRun := strings.Index(got, "r.r axmt500_wf")
	if iEnt < 0 {
		t.Fatalf("给了 ENT 就必须下发: %q", got)
	}
	if !(iEnt < iFgl && iFgl < iRun) {
		t.Fatalf("顺序必须是 TOPENT → FGLSERVER → r.r，实际: %q", got)
	}
}

// 值不限数字/文本：内嵌单引号必须被剔除，任何值都拆不开这条命令。
// 与 tt debug 的 topentShellCmd 同一套规则（见两处注释）。
func TestTopentExportLineStripsQuotes(t *testing.T) {
	cases := []struct{ in, want string }{
		{"99", "export TOPENT='99'\r\n"},
		{"DSCNJ", "export TOPENT='DSCNJ'\r\n"},
		{"a'; rm -rf /; echo 'b", "export TOPENT='a; rm -rf /; echo b'\r\n"},
		{" 99 ", "export TOPENT=' 99 '\r\n"}, // 剔除空白是调用方的事（CLI 已 trim）
	}
	for _, c := range cases {
		if got := topentExportLine(c.in); got != c.want {
			t.Errorf("topentExportLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// 只由引号组成的值：退化成空赋值，但引号必须被剔干净（不许跑出命令之外）
	if got := topentExportLine("'''"); strings.Count(got, "'") != 2 {
		t.Fatalf("引号必须被剔除干净: %q", got)
	}
}

// 命令行里塞不进第二个 export：整串必须是三条独立命令，不能靠分号粘成一条
// （PTY 回显解析依赖行结构，粘起来会让"下发成功"与"命令回显"分不开）。
func TestJobLaunchCmdKeepsOneCommandPerLine(t *testing.T) {
	got, err := jobLaunchCmd(PathRunnerRR, "10.0.0.1", 1200, "axmt500_wf", "99")
	if err != nil {
		t.Fatalf("合法参数不该报错: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(got, "\r\n"), "\r\n")
	if len(lines) != 3 {
		t.Fatalf("应是 3 行（TOPENT / FGLSERVER / r.r），实际 %d 行: %q", len(lines), got)
	}
}

// 默认那一组端口不通时该**自动换一组**（三个端口的相对关系不能变：画面接线里那个数值
// 恒等于代理端口 − 6400），并把换端口这件事写进日志；调用方点名要的端口则照旧只报错。
//
// 为什么值得钉住：Windows 会把成片的端口划给 Hyper-V/WSL 保留（excludedportrange），
// 落在里面的端口 bind 直接 WSAEACCES，**没有进程在监听也一样失败** —— 实测 7899 落在
// 7809–7908 里，Get-NetTCPConnection 看是"空闲"，而默认那组端口因此整组不可用。
func TestPickPortsFallsBackWhenDefaultUnusable(t *testing.T) {
	// 占住"这一组"的第一个端口（模拟"整组用不了"）。
	//
	// 为什么不是写死 7400：它不一定是空的（这台机器上它有时落在 Windows 给 Hyper-V/WSL
	// 保留的段里），而**这条测试不该因此变成空转** —— 它测的是"该换一组"这个行为，
	// 与端口号具体是几无关。候选都在低位、且 +7000 仍在合法范围内（换组最多挪 8 组）。
	var ln net.Listener
	base := 0
	for _, cand := range []int{7400, 17400, 27400, 37400, 47400, 57400} {
		l, lerr := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(cand)))
		if lerr != nil {
			continue
		}
		ln, base = l, cand
		break
	}
	if ln == nil {
		t.Fatalf("这条用例要占住一个基准端口，六个候选都占不住（这台机器的端口保留段太宽？）")
	}
	defer ln.Close()

	cfg := OpenConfig{GDCPort: base, ProxyPort: base + fepPortOffsets[1], InjectPort: base + fepPortOffsets[2]}
	var logged []string
	logf := func(f string, a ...any) { logged = append(logged, fmt.Sprintf(f, a...)) }
	if err := pickPorts(&cfg, false, logf); err != nil {
		t.Fatalf("这一组端口不通时该自动换一组: %v", err)
	}
	if cfg.GDCPort == base {
		t.Fatalf("%d 被占着，不该还挑它: %+v", base, cfg)
	}
	if cfg.ProxyPort-cfg.GDCPort != fepPortOffsets[1] || cfg.InjectPort-cfg.GDCPort != fepPortOffsets[2] {
		t.Fatalf("三个端口的相对关系必须成套（代理−6400 就是画面接线里那个数值）: %+v", cfg)
	}
	for _, p := range []int{cfg.GDCPort, cfg.ProxyPort, cfg.InjectPort} {
		if err := portFree(p); err != nil {
			t.Fatalf("换出来的 %d 还是不能用: %v", p, err)
		}
	}
	if len(logged) == 0 {
		t.Fatal("换了端口就得说一声（日志里要能看见）")
	}

	// 点名要的端口：不通就报错，不许自作主张换
	pinned := OpenConfig{GDCPort: base, ProxyPort: base + fepPortOffsets[1], InjectPort: base + fepPortOffsets[2]}
	if err := pickPorts(&pinned, true, logf); err == nil {
		t.Fatal("调用方点名要的端口不通时该报错，而不是偷偷换一组")
	}
	if pinned.GDCPort != base {
		t.Fatalf("报错路径也不该改动调用方给的端口: %+v", pinned)
	}
}

// 点名要的端口要**按调用方给的那三个**检查，不是按 GDC 端口 +200/+499 推。
//
// 为什么值得钉住：从 GDC 端口推另外两个只对"默认那一组"成立。按偏移推有两个后果，
// 一个比一个难查：
//   - 调用方给的 ProxyPort 被占**查不出来** —— 隔离实例照起，等到代理那一步才失败，
//     白起一个 gdc（正好是这条测试要钉的那件事，TestOpenRejectsBusyPortWithoutResidue
//     就是照着"端口体检在起实例之前"写的，实测挂在"起隔离实例失败"上）；
//   - GDCPort+200 恰好被占时**误报**，用户点名的端口明明是空的。
func TestPickPortsChecksThePinnedPortsThemselves(t *testing.T) {
	// ① 调用方点名的代理端口被占 → 必须报 ErrPortBusy，且不许改调用方的值
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()
	proxyPort := busy.Addr().(*net.TCPAddr).Port
	gdcPort := freePort(t)
	cfg := OpenConfig{GDCPort: gdcPort, ProxyPort: proxyPort, InjectPort: freePort(t)}
	if err := pickPorts(&cfg, true, func(string, ...any) {}); !errors.Is(err, ErrPortBusy) {
		t.Fatalf("点名的代理端口 %d 被占就该报 ErrPortBusy，实得 %v", proxyPort, err)
	}
	if cfg.GDCPort != gdcPort || cfg.ProxyPort != proxyPort {
		t.Fatalf("报错路径不该改动调用方给的端口: %+v", cfg)
	}

	// ② 反面：GDCPort+200 被占，但调用方要的代理端口是别的空闲端口 → 不该误报。
	//
	// 基准端口要**同时**满足两件事：它自己空着、而它 +200 能占住。找不到就是这台机器的
	// 端口保留段太宽（那是"假设不成立"），照实红，别静默跳过变成空转。
	var blocker net.Listener
	gdc2 := 0
	for _, cand := range []int{17400, 27400, 37400, 47400, 57400, 63000} {
		p, perr := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(cand)))
		if perr != nil {
			continue
		}
		p.Close()
		l, lerr := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(cand+fepPortOffsets[1])))
		if lerr != nil {
			continue
		}
		blocker, gdc2 = l, cand
		break
	}
	if blocker == nil {
		t.Fatalf("找不到「自己空着、+%d 又能占住」的基准端口：这条反面用例的假设不成立", fepPortOffsets[1])
	}
	defer blocker.Close()
	cfg2 := OpenConfig{GDCPort: gdc2, ProxyPort: freePort(t), InjectPort: freePort(t)}
	if err := pickPorts(&cfg2, true, func(string, ...any) {}); err != nil {
		t.Fatalf("点名的三个端口都空着，不该报错（%d 被占跟它们无关）: %v", gdc2+fepPortOffsets[1], err)
	}
}
