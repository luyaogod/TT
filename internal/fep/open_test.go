package fep

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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
