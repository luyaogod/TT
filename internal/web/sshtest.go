package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"tt/internal/host"
)

// sshTestReq 「SSH 服务器」页签的「测试连接」:用表单里的连接参数真登录一次。
//
// 为什么单独一个端点,而不是复用 dbprobe:后者登进去还要跑 chenv/tnsnames(慢),
// 且会把库要素带出来;这里只回答一句话 —— 这台机器用这组账号口令**能不能登进去**。
// 参数由前端从**表单当前值**给出(不必先保存),所以这里不读 config.json。
type sshTestReq struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// sshTestResp 成功时的凭据:地址 + 远端主机名(+ 主机名取不到时的说明)。
type sshTestResp struct {
	OK        bool   `json:"ok"`
	Address   string `json:"address"`
	Hostname  string `json:"hostname,omitempty"`
	ElapsedMs int64  `json:"elapsedMs"`
	Note      string `json:"note,omitempty"`
}

// hSSHTest 测一次 SSH 链路:登录 → 取主机名 → 关掉。
//
// 只读:不写配置、不落盘、不建隧道;失败一律 502 交给前端 toast(与 dbprobe 同口径)。
func (s *Server) hSSHTest(w http.ResponseWriter, r *http.Request) {
	var req sshTestReq
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	req.Host = strings.TrimSpace(req.Host)
	req.User = strings.TrimSpace(req.User)
	if req.Host == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("IP 主机不能为空"))
		return
	}
	if req.User == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("账号不能为空"))
		return
	}
	// 端口缺省按 22:表单里清空端口时不该变成 "host:0" 那种连不上的地址。
	if req.Port <= 0 || req.Port > 65535 {
		req.Port = 22
	}

	cfg := host.SSHConfig{Host: req.Host, Port: req.Port, User: req.User, Password: req.Password}
	start := time.Now()
	out, err := s.testSSH(cfg)
	if err != nil {
		writeErr(w, http.StatusBadGateway, fmt.Errorf("SSH 连接失败: %w", err))
		return
	}
	writeJSON(w, http.StatusOK, sshTestResp{
		OK: true, Address: cfg.Addr(), Hostname: out.Hostname, Note: out.Note,
		ElapsedMs: time.Since(start).Milliseconds(),
	})
}

// testSSH 走注入的实现;没注入就用真的(与 probeDB/verifyAcc 同一手法)。
func (s *Server) testSSH(cfg host.SSHConfig) (*host.SSHTestOut, error) {
	if s.opt.SSHTest != nil {
		return s.opt.SSHTest(cfg)
	}
	return host.TestSSH(cfg)
}
