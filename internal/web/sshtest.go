package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/luyaogod/TT/internal/connsvc"
	"github.com/luyaogod/TT/internal/host"
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
	res := s.connSvc().TestSSHConn(cfg)
	if !res.OK {
		writeErr(w, http.StatusBadGateway, fmt.Errorf("SSH 连接失败: %s", res.Message))
		return
	}
	writeJSON(w, http.StatusOK, sshTestResp{
		OK: true, Address: cfg.Addr(), Hostname: bannerOrEmpty(res.Banner), Note: noteOf(res.Banner),
		ElapsedMs: time.Since(start).Milliseconds(),
	})
}

// bannerOrEmpty / noteOf 把连接服务的 Banner 拆回这个端点的两个字段。
//
// connsvc 的 Banner 语义是"能拿到主机名就给主机名，拿不到就给说明"；
// 这个端点的历史形状是 hostname + note 两个字段，前端按它们分别显示，
// 所以这里按同一条规则拆开（而不是把说明塞进 hostname 里）。
func bannerOrEmpty(banner string) string {
	if strings.HasPrefix(banner, "已登录") {
		return ""
	}
	return banner
}

func noteOf(banner string) string {
	if strings.HasPrefix(banner, "已登录") {
		return banner
	}
	return ""
}

// testSSH 走连接服务（与 Host API 的 ConnectionsService.TestSSH 同一份实现）。
func (s *Server) testSSH(cfg host.SSHConfig) *connsvc.TestSSHResult {
	return s.connSvc().TestSSHConn(cfg)
}
