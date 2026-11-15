package web

import (
	"path/filepath"
	"testing"

	"tt/internal/connsvc"
)

// testConnSvc 造一个注入了假外部依赖的**进程内连接服务**。
//
// 设置页的 /api/conntest、/api/sshtest 现在走 connsvc（与 Host API 同一份实现），
// 所以测试也从"注入一个 OpenDB 函数"变成"注入一个连接服务的依赖"：
// 被测的是真实的 connsvc 路径（含隧道改写、阶段归类、关闭收口），
// 只有最底层那次拨号被替身换掉。
//
// 配置路径指向临时目录里一个**不存在**的文件：config.Load 把它当空配置，
// 这些端点本来也不读配置（设置页传的是表单当前值）。
func testConnSvc(t *testing.T, opt connsvc.Options) *connsvc.Service {
	t.Helper()
	if opt.ConfigPath == "" {
		opt.ConfigPath = filepath.Join(t.TempDir(), "config.json")
	}
	return connsvc.New(opt)
}

// fakeTunnel 是 connsvc.Tunnel 的假实现：只回答本地端口，不真建连接。
type fakeTunnel struct {
	port   int
	closed bool
}

func (f *fakeTunnel) LocalPort() int { return f.port }
func (f *fakeTunnel) Close() error   { f.closed = true; return nil }
