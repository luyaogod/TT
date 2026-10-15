package fep

import (
	"context"
	"errors"
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
