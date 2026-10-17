package fep

// 画面通道（P4）的最小判据：它必须能单独存在、能安全收尾、端口自洽。
// 真正的起停要 GDC 可执行件（真机才有），所以这里钉"空对象也不炸"这一层 ——
// 收尾路径上出 panic 的代价是"进程退了一半、GDC 留在后台"。

import "testing"

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

