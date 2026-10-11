package fep

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/fep"
)

// tt fep fe —— **我们自己当 GDC**：一个无头前端。
//
// 为什么要它：运行期只把界面画给"一个前端"看。以前我们借 GDC 当前端 ✗，
// 于是焦点/状态/事件形状都得从 GDC 的转述里反推（今天大半时间花在这上面）。
// 换成我们自己当那个前端之后，**收什么、发什么、什么时候发**全在我们手里。
//
// 接线是现成的（实测）：运行期按 `FGLSERVER 端口 + 6400` 回连前端，
// 而 internal/debug 的启动器会自动导出 FGLSERVER（session.go:550）。
// 所以：FGLSERVER 指到 `<我们的IP>:1300` → 它连我们的 7700。
//
// 用法（先起前端，再用调试器把作业跑起来）：
//
//	tt fep fe listen --port 1300 --ctl 7898          # 起前端（前台；实际用后台跑）
//	tt debug start apmt500_wf -m apm --env 主机109    # 作业连过来
//	tt fep fe status --ctl 127.0.0.1:7898            # 看连上没有、收了多少
//	tt fep fe send --ctl 127.0.0.1:7898 'ActionEvent 0{{idRef "3435"}}'
//
// 收到的字节会原样落成录制文件，所以既有的 `tt fep ui state/find/roles --from <它>`
// 全部照用（不用新写一套读界面的东西）。

var (
	fePort    int
	feCtl     string
	feLogDir  string
	feCtlAddr string
	feEvent   string
)

var feCmd = &cobra.Command{
	Use:   "fe",
	Short: "无头前端：我们自己当 GDC（收 AUI / 发事件，绕开 GDC 与注入口）",
}

var feListenCmd = &cobra.Command{
	Use:   "listen",
	Short: "起一个无头前端：等运行期连入，收 AUI 并落成录制件，同时开一个控制口",
	Long: `起前端并等作业连上来。

接线（实测）：运行期按 FGLSERVER 端口 + 6400 回连前端。所以

    --port 1300      → 运行期会连到 7700（本进程监听的就是它）
    FGLSERVER        → 用调试启动器时它自动导出（config 里的 debug.fglserver）

控制口（--ctl，本地回环）是行协议，与注入口同形态：
    status                    一行 JSON：是否已连、序号、收了多少字节/消息、录制件路径
    send <事件体>              发一条事件（自动推进 _om 序号）
    quit                      退出`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := feLogDir
		if dir == "" {
			dir, _ = os.UserCacheDir()
			dir = filepath.Join(dir, "tt", "fe")
		}
		addr := fmt.Sprintf("0.0.0.0:%d", fePort+6400)
		s, err := fep.ListenFE(addr, feCtl, dir)
		if err != nil {
			return fail(exitState, "%v", err)
		}
		fmt.Printf("无头前端已就绪\n")
		fmt.Printf("  监听      : %s   （运行期按 FGLSERVER 端口+6400 回连到这里）\n", s.Addr())
		fmt.Printf("  控制口    : %s\n", s.CtlAddr())
		fmt.Printf("  录制件    : %s\n", s.RecPath())
		fmt.Printf("  提示      : config 里 debug.fglserver 要指向「本机IP:%d」\n", fePort)
		fmt.Printf("  等运行期连入…\n")
		if err := s.Accept(); err != nil {
			return fail(exitFailed, "%v", err)
		}
		st := s.Status()
		fmt.Printf("运行期已连入（收到 %d 字节）\n", st.BytesRT)
		// 保持进程活着：收流在 readLoop 里持续跑
		for {
			time.Sleep(2 * time.Second)
			if st := s.Status(); st.Err != "" {
				fmt.Printf("连接结束：%s（共收 %d 字节 / %d 条消息）\n", st.Err, st.BytesRT, st.Messages)
				return nil
			}
		}
	},
}

var feStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "看无头前端的状态（连上没有、序号、收了多少、录制件在哪）",
	RunE: func(cmd *cobra.Command, args []string) error {
		line, err := feCtlDial(feCtlAddr, "status")
		if err != nil {
			return fail(exitNoSession, "%v", err)
		}
		fmt.Println(strings.TrimSpace(line))
		return nil
	},
}

var feSendCmd = &cobra.Command{
	Use:   "send <事件体>",
	Short: "经无头前端发一条事件（自动推进 _om 序号，与注入口同一套规矩）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		line, err := feCtlDial(feCtlAddr, "send "+args[0])
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		fmt.Println(strings.TrimSpace(line))
		return nil
	},
}

// feCtlDial 是一条短连接的行协议调用。
func feCtlDial(addr, cmdline string) (string, error) {
	if addr == "" {
		return "", errors.New("没给控制口：加 --ctl 127.0.0.1:7898")
	}
	c, err := netDial(addr)
	if err != nil {
		return "", fmt.Errorf("连不上控制口 %s：%w", addr, err)
	}
	defer c.Close()
	if _, err := c.Write([]byte(cmdline + "\n")); err != nil {
		return "", err
	}
	buf := make([]byte, 4096)
	n, _ := c.Read(buf)
	return string(buf[:n]), nil
}

func init() {
	feListenCmd.Flags().IntVar(&fePort, "port", 1300, "FGLSERVER 里写的端口（本进程实际监听 port+6400）")
	feListenCmd.Flags().StringVar(&feCtl, "ctl", "127.0.0.1:7898", "控制口地址")
	feListenCmd.Flags().StringVar(&feLogDir, "log", "", "录制件目录（缺省 <缓存>/tt/fe）")
	feStatusCmd.Flags().StringVar(&feCtlAddr, "ctl", "127.0.0.1:7898", "控制口地址")
	feSendCmd.Flags().StringVar(&feCtlAddr, "ctl", "127.0.0.1:7898", "控制口地址")
	feCmd.AddCommand(feListenCmd, feStatusCmd, feSendCmd)
}

// netDial 是 net.DialTimeout 的薄封装（控制口的调用很短，3 秒足够）。
func netDial(addr string) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, 3*time.Second)
}
