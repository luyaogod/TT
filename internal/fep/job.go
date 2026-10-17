package fep

// 作业连接（Job）：一条 SSH 连接 + 它的 PTY + "作业什么时候结束" + 编译件标识。
//
// # 为什么从 Runtime 里分出来
//
// P4 之前 `Runtime` 同时持有：画面通道（GDC + 代理）、作业连接、会话登记。三件事的
// 生命周期并不相同 ——
//
//   - 画面通道可以**没有作业**（`--no-job`：先起代理，再把调试线的作业挂上来）；
//   - 作业可以**没有画面通道**（回放时不需要 GDC）；
//   - 会话登记比两者都长（收工时它最后清）。
//
// 把它们绑在一个对象上，谁想单独用一个都得借整套 Runtime。现在 Job 只回答一件事：
// "这条作业还活着吗、它的编译件是哪一份"。

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"tt/internal/host"
)

// Job 是一条被驱动起来的远端作业。
type Job struct {
	cfg   OpenConfig
	logf  func(string, ...any)
	pty   *host.PTYSession
	conn  *host.SSHConn
	done  chan struct{}
	binary BinaryID
}

// newJob 建一个未打开的作业连接（done 在 watch 结束时关闭）。
func newJob(cfg OpenConfig, logf func(string, ...any)) *Job {
	return &Job{cfg: cfg, logf: logf, done: make(chan struct{})}
}

// Done 在远端作业结束时关闭。
func (j *Job) Done() <-chan struct{} { return j.done }

// Binary 是作业编译件的标识（探不到就是空值）。
func (j *Job) Binary() BinaryID { return j.binary }

// Open 登录、把 FGLSERVER 指向我们的代理、启动作业。
//
// setFGL 用来把这轮接线记到会话上（Runtime 的事）：Job 不持有 fglServer，
// 因为它只负责"让作业把画面画到那个地址"，不负责"这轮接线是什么"。
func (j *Job) Open(ip string, setFGL func(string) error) error {
	conn, err := host.Dial(j.cfg.SSH)
	if err != nil {
		return fmt.Errorf("fep: SSH 连接失败: %w", err)
	}
	j.conn = conn
	pty, err := conn.NewPTY(120, 40)
	if err != nil {
		return fmt.Errorf("fep: 开 PTY 失败: %w", err)
	}
	j.pty = pty

	// 登录后可能先出区域菜单，也可能直接给 shell 提示符（两台机器就不一样）：
	// 两种都等，再按实际情况决定要不要选区域。
	which, err := waitAny(pty, 25*time.Second, host.ReLoginMenu, host.ReShellPrompt)
	if err != nil {
		return err
	}
	if which == 0 {
		if j.cfg.Zone == "" {
			return errors.New("fep: 这个环境没配 zone（区域菜单没法选）")
		}
		if err := pty.Write(j.cfg.Zone + "\r"); err != nil {
			return err
		}
		if err := waitRegexp(pty, host.ReShellPrompt, 25*time.Second, "shell 提示符"); err != nil {
			return err
		}
	}

	port := j.cfg.ProxyPort - FGLSERVEROffset
	if port <= 0 || port > 65535 {
		return fmt.Errorf("%w: 代理端口 %d 减 %d 后不是合法端口", ErrUsage, j.cfg.ProxyPort, FGLSERVEROffset)
	}
	if err := setFGL(ip); err != nil {
		return err
	}
	cmdLine, cerr := jobLaunchCmd(j.cfg.Path, ip, port, j.cfg.Prog, j.cfg.Ent)
	if cerr != nil {
		// 入口已经查过一遍（validateLaunch），走到这里说明两条路的口径不一致 ——
		// 那就**什么都别发**：半条命令出去了比报错坏得多。
		return fmt.Errorf("fep: 启动命令拼不出来: %w", cerr)
	}
	if err := pty.Write(cmdLine); err != nil {
		return fmt.Errorf("fep: 下发作业启动命令失败: %w", err)
	}
	// The job client keeps running while the GUI is open; its connection ending
	// means the job is gone.
	go j.watch()
	return nil
}

// Close 断开作业连接（PTY 先关，再关 SSH）。
func (j *Job) Close() error {
	if j == nil {
		return nil
	}
	if j.pty != nil {
		j.pty.Close()
	}
	if j.conn != nil {
		return j.conn.Close()
	}
	return nil
}

// watch 把 PTY 读空，并在作业结束时关闭 done。
func (j *Job) watch() {
	defer close(j.done)
	buf := make([]byte, 4096)
	for {
		n, err := j.pty.Read(buf)
		if n > 0 {
			line := strings.TrimSpace(string(buf[:n]))
			if line != "" {
				j.logf("远端: %s", firstLine(line))
			}
		}
		if err != nil {
			return
		}
	}
}

// probeBinary 只读地取一次作业编译件的标识（大小 + 修改时间）。
//
// 为什么要有它：指纹能发现"界面结构变了"，但**发现不了**"程序只是重编译了一次、
// 界面一模一样"（那种情况下编号表逐条相同）。两者合起来才是完整的"这份录制还算不算数"。
//
// 只读：一条 `ls -ln`。探不到就照实留空 —— 走 `r.r` 时程序路径由登录环境决定
// （`$AZZ`/`$ASF`），我们只给了作业名，所以那条路要调用方显式给 --binary-ref。
func (j *Job) probeBinary() BinaryID {
	ref := j.cfg.BinaryRef
	if ref == "" || j.conn == nil {
		return BinaryID{}
	}
	if err := ValidateBinaryRef(ref); err != nil {
		j.logf("编译件标识不探：%v", err)
		return BinaryID{}
	}
	// 实测（主机109，2026-10-15）：GNU ls 的 --time-style=+FORMAT 可用，输出形如
	//   -rwxrwxr-x 1 1012 1012 310855 2026-09-24T10:06:28 /u1/t35prd/erp/asf/42r/bsft001_wf.42r
	out, err := j.conn.Output("ls -ln --time-style=+%Y-%m-%dT%H:%M:%S "+ref, 15*time.Second)
	if err != nil {
		j.logf("探测编译件标识失败（不影响本轮）：%v", err)
		return BinaryID{}
	}
	f := strings.Fields(out)
	if len(f) < 6 {
		return BinaryID{}
	}
	size, _ := strconv.ParseInt(f[4], 10, 64)
	return BinaryID{Ref: f[len(f)-1], Size: size, ModTime: f[5]}
}

// ProbeBinary 探一次并记下（Open 之后调用；探不到不影响本轮）。
func (j *Job) ProbeBinary() {
	j.binary = j.probeBinary()
}

// jobDone 给 Wait 用：没有作业时返回一个**永不关闭**的 channel
// （只起画面/只做会话管理的会话，等的是 ctx 或代理退出）。
func jobDone(j *Job) <-chan struct{} {
	if j == nil {
		return nil
	}
	return j.Done()
}

// jobBinary 取作业编译件标识（没有作业就是空值）。
func jobBinary(j *Job) BinaryID {
	if j == nil {
		return BinaryID{}
	}
	return j.Binary()
}
