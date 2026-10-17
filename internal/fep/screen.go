// 画面通道（Screen）：隔离的 GDC 实例 + 录制代理。
//
// # 为什么把它单独提出来
//
// 从前这三件事（起 GDC、起代理、登记会话）都揉在 `Open()` 里，`Runtime` 一个对象同时
// 持有"作业连接的 PTY""GDC 进程""录制代理""会话登记"。后果是任何一件要单独用都不行：
//
//   - 只想录一段画面（不要作业）→ 只能借用整套 Runtime；
//   - 想查"这一轮的代理录了多少"→ 得先有一个 Runtime；
//   - Web/CLI 想单说"把画面通道关掉"→ 没有这个词。
//
// 现在 `Screen` 就是"画面通道"这一个概念：它自己起停 GDC 与代理，暴露三个端口与注入入口。
// `Runtime` 变成**组合**（会话 + 作业 + 一块 Screen），而不是"什么都自己干"。
//
// # 边界
//
// Screen 不认识会话、不认识作业、不认识注册表 —— 它只知道"GDC 在哪、代理在哪、注入口在哪"。
// 这正是 P4 要的：能力分组，互不内嵌。
package fep

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"tt/internal/winproc"
)

// ScreenConfig 是起一块画面通道要的东西（全部来自 OpenConfig 的对应字段）。
type ScreenConfig struct {
	GDCExe   string
	GDCLog   string
	GDCPort  int
	ProxyPort int
	InjectPort int
	RecordDir string
	Tag       string
	WaitGDC   time.Duration
}

// Screen 是一块画面通道：隔离 GDC 实例 + 录制代理（+ 注入口）。
type Screen struct {
	gdcPID   int
	gdcPort  int
	proxy    *Proxy
	proxyPort int
	injectPort int
}

// StartScreen 起 GDC 实例与代理。失败时**自己收干净**（不留半块画面通道）。
func StartScreen(cfg ScreenConfig, logf func(string, ...any)) (*Screen, error) {
	pid, err := winproc.SpawnDetached(cfg.GDCExe,
		[]string{"-n", "-M", "-A", "0", "-p", strconv.Itoa(cfg.GDCPort)}, cfg.GDCLog)
	if err != nil {
		return nil, fmt.Errorf("fep: 起隔离实例失败: %w", err)
	}
	logf("GDC 已起：pid %d，端口 %d", pid, cfg.GDCPort)
	if err := waitPort(cfg.GDCPort, cfg.WaitGDC); err != nil {
		winproc.Kill(pid)
		return nil, err
	}
	time.Sleep(gdcWarmup)

	p := NewProxy(ProxyConfig{
		Listen:  fmt.Sprintf("0.0.0.0:%d", cfg.ProxyPort),
		Target:  net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.GDCPort)),
		LogDir:  cfg.RecordDir,
		Tag:     cfg.Tag,
		Control: net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.InjectPort)),
	})
	if err := p.Start(); err != nil {
		winproc.Kill(pid)
		return nil, err
	}
	return &Screen{
		gdcPID: pid, gdcPort: cfg.GDCPort,
		proxy: p, proxyPort: cfg.ProxyPort, injectPort: cfg.InjectPort,
	}, nil
}

// Proxy 返回底下的录制代理（连接清单、注入、关闭都从它走）。
func (s *Screen) Proxy() *Proxy { return s.proxy }

// GDCPID 是隔离实例的进程号（收尾只杀它）。
func (s *Screen) GDCPID() int { return s.gdcPID }

// Ports 返回三个端口：GDC 自己听的、代理对外听的、注入口听的。
func (s *Screen) Ports() (gdc, proxy, inject int) {
	return s.gdcPort, s.proxyPort, s.injectPort
}

// Inject 往一条连接里注入一段事件文本（走代理的控制口）。
func (s *Screen) Inject(connID int, eventText string) error {
	return s.proxy.Inject(connID, eventText)
}

// Close 收掉整块通道：代理先关（它要把缓冲写完），再杀 GDC。
func (s *Screen) Close() error {
	var firstErr error
	if s.proxy != nil {
		if err := s.proxy.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if s.gdcPID != 0 && winproc.Alive(s.gdcPID) {
		if err := winproc.Kill(s.gdcPID); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Screen 返回这一轮的画面通道（可能为 nil：只做会话管理、没起画面）。
//
// 给上层（REST / CLI）一个正式的入口：以后 `/api/screen` 这类接口要问"这一轮的代理录了多少、
// 注入口在哪"，答案都在这个对象上，不必再从 Runtime 的一堆字段里拼。
func (r *Runtime) Screen() *Screen { return r.screen }
