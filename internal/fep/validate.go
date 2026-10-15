package fep

import (
	"fmt"
	"net"
	"regexp"
	"strconv"

	"tt/internal/host"
)

// 本文件是**人给的字符串进远端 shell 之前的唯一一道门**。
//
// 为什么必须有：作业名与 FGLSERVER 地址都会被裸插进登录 shell 的命令行
// （见 jobLaunchCmd），区域代码会被直接写进登录 PTY（见 openJob）。这三样以后
// 都由前端面板提供 —— 那就是一条"从界面到生产服务器"的注入路径。同类洞仓库
// 已经修过一处（host.ChenvCmd 的 `--zone "36; id"` 实测能在服务器上执行命令），
// 这里照它做：**只放行白名单字符**，而且在发往服务器之前就拒绝。
//
// 为什么是白名单而不是转义：这三个值本来就只由字母数字与下划线构成
// （实测的作业名是 axmt500_wf / bsft001_wf，区域是 36）。合法字符集越小，
// 出错的空间越小；转义要做对，得把外层引号、PTY 行编辑、`r.r` 后面那层脚本全算上。

// reProgName 作业程序名：字母开头 + 字母/数字/下划线，最长 64。
//
// 必须字母开头：T100 生成器产出的都是 `<字母>…_wf`，而允许数字开头就会把
// `42r` 这种看起来像路径片段的串也放进来。
var reProgName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

// reBinaryRef 作业编译件在服务器上的绝对路径白名单（与 host 的 reToolPath 同一类）。
// 它只用来拼一条**只读的** `ls -ln`，但同样是"人给的字符串进远端 shell"，
// 所以照样过白名单：没有引号、没有分号、没有空格，就没有拆词的机会。
var reBinaryRef = regexp.MustCompile(`^/[A-Za-z0-9_./-]{1,300}$`)

// ValidateBinaryRef 校验 --binary-ref。
func ValidateBinaryRef(ref string) error {
	if ref == "" {
		return nil // 空 = 不探测（r.r 那条路本来就不知道路径）
	}
	if !reBinaryRef.MatchString(ref) {
		return fmt.Errorf("%w: --binary-ref %q 不合法（要服务器上的绝对路径，只允许字母/数字/下划线/点/斜杠/连字符）", ErrUsage, ref)
	}
	return nil
}

// ValidateProg 校验作业程序名（空值在这里也算不合法：调用方若允许"不起作业"，
// 该走 NoJob 那条路，而不是给一个空名字）。
func ValidateProg(prog string) error {
	if !reProgName.MatchString(prog) {
		return fmt.Errorf("%w: 作业程序名 %q 不合法（只允许字母开头，字母/数字/下划线，最长 64）", ErrUsage, prog)
	}
	return nil
}

// ValidateFGLServerIP 校验 FGLSERVER 里那个地址。
//
// 只收 IP 字面量（可带端口）：这个值会被插进 `export FGLSERVER=<ip>:<port>`，
// 而它的语义就是"本机在服务器那侧的可达地址"—— 域名在这里没有合法用途
// （服务器侧要回连的是一个具体地址，域名解析在服务器上做只会引入歧义）。
func ValidateFGLServerIP(ip string) error {
	if ip == "" {
		return nil // 空 = 按到 SSH host 的路由探测（见 localAddrFor）
	}
	if net.ParseIP(ip) != nil {
		return nil
	}
	// 先判整串再判 host:port：IPv6 字面量自己带冒号，顺序反了 SplitHostPort 会把它判死。
	// 端口也要自己判数字：`1.2.3.4:1300; id` 这种 SplitHostPort 是**成功**的
	// （它只找最后一个冒号，不检查端口长什么样），漏判它等于白设这道门。
	if h, p, err := net.SplitHostPort(ip); err == nil && net.ParseIP(h) != nil {
		if n, perr := strconv.Atoi(p); perr == nil && n > 0 && n <= 65535 {
			return nil
		}
	}
	return fmt.Errorf("%w: --fgserver-ip %q 不是 IP 字面量（域名不收：这一项只用本机可达地址）", ErrUsage, ip)
}

// ValidateZone 校验区域代码（与 host 拼 chenv 用的是同一份白名单）。
func ValidateZone(zone string) error {
	if zone == "" {
		return nil // 空 = 登录后没有区域菜单，走不到选区域那一步
	}
	if !host.ZoneOK(zone) {
		return fmt.Errorf("%w: 区域代码 %q 不合法（只允许字母/数字/下划线/连字符，最长 16）", ErrUsage, zone)
	}
	return nil
}

// ValidatePath 校验启动路径是不是已知的那几个值之一。
//
// 它进录制身份、也进闸门的路径比对：写错了不会报错，只会让闸门比出一条"路径不符"，
// 而人看着两个值都像是对的。所以宁可在入口就拒。
// 空值是合法的：它表示"还没定"，由 fillDefaults 按起没起作业推默认。
func ValidatePath(p LaunchPath) error {
	switch p {
	case "", PathRunnerRR, PathDebug, PathRunnerRD, PathNone:
		return nil
	}
	return fmt.Errorf("%w: 启动路径 %q 不在已知的四条里（%s / %s / %s / %s）",
		ErrUsage, p, PathRunnerRR, PathDebug, PathRunnerRD, PathNone)
}

// launchCommand 拼出"在登录 shell 里把作业起起来"的那一行（不含 TOPENT 与 FGLSERVER）。
//
// 只有 `r.r` 一条：T100 的标准启动器。另外两条不从这里走 ——
// 调试线由 `internal/debug` 自己起 `fglrun -d`（见它的 README），
// 而 T100 原生的 `r.d`（gendbg → fgldeb）**已从命令面拿掉**：那是 T100 的调试器，
// 不是我们的；录制与调试都走调试线（`tt fep rec start` / `tt debug start`）。
// `PathRunnerRD` 这个常量仍留着 —— 它用来**识别**以前录的 r.d 录制件。
func launchCommand(path LaunchPath, prog string) (string, error) {
	switch path {
	case "", PathRunnerRR:
		// T100 的标准启动器：`r.r <作业>` → `fglrun $AZZ/42r/azzp000.42r <作业>`（不可调试）
		return "r.r " + prog, nil
	case PathRunnerRD:
		return "", fmt.Errorf("%w: T100 原生调试器（r.d → fgldeb）已从命令面拿掉；要录要调都用调试线（tt debug start / tt fep rec start）", ErrUsage)
	case PathDebug:
		return "", fmt.Errorf("%w: 调试线不用登录 shell 起作业 —— 那是 `tt debug start`（它自己起 fglrun -d）", ErrUsage)
	case PathNone:
		return "", fmt.Errorf("%w: 路径 none 表示不起作业（配合 --no-job）", ErrUsage)
	}
	return "", fmt.Errorf("%w: 启动路径 %q 不在已知的四条里", ErrUsage, path)
}

// validateLaunch 把 OpenConfig 里所有会被拼进远端 shell 或写进登录 PTY 的值一次查完。
//
// 放在 Open 的入口（起隔离实例之前）：拒绝要发生在**任何东西被发出去之前**，
// 否则"命令已经下发了"与"参数不合法"会同时成立，而前者是不可撤回的。
func validateLaunch(cfg OpenConfig) error {
	if !cfg.NoJob {
		if cfg.Prog == "" {
			return fmt.Errorf("%w: 没给作业程序（--prog）", ErrUsage)
		}
		if err := ValidateProg(cfg.Prog); err != nil {
			return err
		}
	}
	if err := ValidateFGLServerIP(cfg.FGLServerIP); err != nil {
		return err
	}
	if err := ValidateBinaryRef(cfg.BinaryRef); err != nil {
		return err
	}
	if err := ValidatePath(cfg.Path); err != nil {
		return err
	}
	return ValidateZone(cfg.Zone)
}
