package fep

// 画面通道（P4）的最小判据：它必须能单独存在、能安全收尾、端口自洽。
// 真正的起停要 GDC 可执行件（真机才有），所以这里钉"空对象也不炸"这一层 ——
// 收尾路径上出 panic 的代价是"进程退了一半、GDC 留在后台"。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEmptyScreenIsSafeToClose：没起来的 Screen 收尾不炸（收尾路径必须幂等）。
func TestEmptyScreenIsSafeToClose(t *testing.T) {
	var s Screen
	if err := s.Close(); err != nil {
		t.Fatalf("空 Screen 收尾不该报错：%v", err)
	}
	if s.Proxy() != nil {
		t.Fatal("空 Screen 不该有代理")
	}
	if s.GDCPID() != 0 {
		t.Fatalf("空 Screen 不该有 GDC 进程：%d", s.GDCPID())
	}
}

// TestScreenPortsAreTheOnesGiven：三个端口按配置原样暴露（上层拿它拼 FGLSERVER）。
func TestScreenPortsAreTheOnesGiven(t *testing.T) {
	s := Screen{gdcPort: 7400, proxyPort: 7600, injectPort: 7899}
	gdc, proxy, inject := s.Ports()
	if gdc != 7400 || proxy != 7600 || inject != 7899 {
		t.Fatalf("端口 = %d/%d/%d", gdc, proxy, inject)
	}
}


// TestOnlyScreenStartsTheScreen：起 GDC 与建代理**只许在 screen.go 里发生**。
//
// 这条守的是 P4 的拆分本身：一旦别处又直接 `winproc.SpawnDetached(GDC…)` 或 `NewProxy(…)`，
// 就会冒出第二个"自己管画面"的地方 —— 端口谁分配、收尾谁负责、登记怎么写全部走样，
// 而且编译能过、单机能跑，只在多会话或收尾时露出来。
func TestOnlyScreenStartsTheScreen(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("读包目录失败：%v", err)
	}
	forbidden := []string{"SpawnDetached(", "NewProxy("}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if name == "screen.go" || name == "proxy.go" {
			continue // 定义处
		}
		b, rerr := os.ReadFile(filepath.Clean(name))
		if rerr != nil {
			t.Fatalf("读 %s 失败：%v", name, rerr)
		}
		for _, line := range strings.Split(string(b), "\n") {
			code := strings.TrimSpace(line)
			if strings.HasPrefix(code, "//") {
				continue
			}
			for _, f := range forbidden {
				if strings.Contains(code, f) {
					t.Errorf("%s 里出现了 %q —— 起 GDC / 建代理只许在 screen.go 里（P4 的边界）", name, f)
				}
			}
		}
	}
}
