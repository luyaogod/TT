package host

import "time"

// SSHTestOut 一次 SSH 链路自检的结果。
//
// Hostname 是"登进去看到的那台机器"的证据 —— 跳板/端口转发环境下,
// 只回一句 ok 看不出连到的到底是谁。拿不到主机名不算失败(部分环境没有 shell,
// 也有把 hostname 摘掉的加固镜像),这种情况落在 Note 里,不影响"链路是通的"。
type SSHTestOut struct {
	Hostname string
	Note     string
}

// TestSSH 登录一次、取主机名、断开。只读:不写配置、不建隧道、不碰库。
//
// 认证失败/连接超时由 Dial 原样返回(20s 握手超时),调用方负责包装成对外错误。
// 单独一个函数而不是让 web 层摆弄 *SSHConn:测试要能注入假实现,
// 真实现又只需要这一条最短路径。
func TestSSH(cfg SSHConfig) (*SSHTestOut, error) {
	conn, err := Dial(cfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	out := &SSHTestOut{}
	if s, err := conn.Output("hostname 2>/dev/null || uname -n", 5*time.Second); err == nil {
		out.Hostname = FirstLines(s, 1)
	} else {
		out.Note = "已登录，但读取主机名失败: " + err.Error()
	}
	return out, nil
}
