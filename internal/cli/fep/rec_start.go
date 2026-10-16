package fep

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/debug"
	"tt/internal/fep"
	"tt/internal/winproc"
)

var (
	recStartProg   string
	recStartEnv    string
	recStartEnt    string
	recStartModule string
	recStartNoCont bool
	recStartBinary string
)

// tt fep rec start —— 一条命令开始录，走**我们自己的调试线**。
//
// 为什么是这条形态（而不是 T100 原生调试器 `r.d`/fgldeb）：
//
//   - 更无感：fgldb 是我们经 SSH 驱动的，不出现在画面上；屏幕里只有隔离 GDC 里那个作业窗口。
//   - 不误触：用户面前没有任何调试器界面可点（fgldeb 的菜单/对话框一个都不在）。
//   - 能自动放行：放行就是 `debug exec "continue"` —— 用户不必会调试，纯录制的用户
//     只要"看作业、复现问题"。
//   - 录出来的件照样能在断点下重放：实测两条调试路径（fgldeb 与 fgldb）下**作业的编号表相同**
//     （都是 `fglrun -d`），所以走哪条都能对上闸门。
//
// 顺序不能换：先起代理（画面得有地方去），再把作业挂上去（`--fglserver` 指到代理）。
// 中间任何一步失败都要把已经起来的代理收掉，不然会留下一个占着三个端口的空会话。
var recStartCmd = &cobra.Command{
	Use:   "start",
	Short: "起一轮录制（走我们自己的调试线，作业自动放行，画面上只有作业）",
	Long: `一条命令开始录：

  1. 起隔离实例 + 代理（这一轮的录制身份写 path=debug、作业、环境/区域/ENT）；
  2. 让调试线把作业挂到这条代理上（tt debug start --fglserver <这一轮探测到的地址>）；
  3. 自动放行（continue）—— 用户不需要会调试，屏幕上只有作业窗口；
  4. 前台常驻，直到 Ctrl-C 或另开一个终端跑 tt fep session close。
     收工落身份，并把录制搬进长期库（<数据目录>/records/<环境>/<ENT>/<录制 ID>/）。

环境与企业编号都要点名（作业落在哪个企业只由它决定），模块默认 asf。

要自己下断点就先加 --no-continue，放行前用 tt debug exec 下好再 continue。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(recStartProg) == "" {
			return fail(exitUsage, "没给作业程序（--prog bsft001_wf）")
		}
		ent := strings.TrimSpace(recStartEnt)
		cfgPath, err := common.ResolveConfig(true)
		if err != nil {
			return err
		}
		dcfg, err := debug.LoadConfig(cfgPath)
		if err != nil {
			return fail(exitFailed, "%v", err)
		}
		name, rerr := openEnvName(dcfg.SSHs, recStartEnv)
		if rerr != nil {
			return fail(exitUsage, "%v", rerr)
		}
		clone := dcfg.CloneEnv(name)
		if clone == nil {
			return fail(exitUsage, "环境 %q 不存在（见 tt env list）", name)
		}
		dcfg = clone
		if dcfg.SSH.Host == "" {
			return fail(exitPreflight, "环境 %q 没配 SSH host", name)
		}
		loc := config.LocationsAt(cfgPath)
		recDir := loc.CacheDir("fep")
		if recDir == "" {
			return fail(exitPreflight, "缓存目录没解析出来（检查配置路径）")
		}
		if err := os.MkdirAll(recDir, 0o755); err != nil {
			return fail(exitFailed, "建录制目录失败: %v", err)
		}
		tag := "fep-" + time.Now().Format("20060102-150405")

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		logf := func(format string, a ...any) {
			fmt.Fprintf(cmd.ErrOrStderr(), "· "+format+"\n", a...)
		}

		// ① 代理先起来：作业的画面得有个地方去
		rt, err := fep.Open(ctx, fep.OpenConfig{
			Env: dcfg.EnvName(), Zone: dcfg.Zone,
			// 这一轮的身份：作业是调试线起的（见 path.go）。身份写对了，回放前的路径比对才有双方的值。
			Prog: recStartProg, Path: fep.PathDebug, Ent: ent,
			// 这一轮是**录**：面板据此说"正在录"（回放那一轮登记成 replay）
			Kind: fep.SessionKindRec,
			// 用户主动录的 → 长期库（策略见 internal/fep/retention.go）
			Origin: fep.OriginUser,
			// 编译件标识：这一条不是所有路径都知道（走 r.r 时程序位置由登录环境决定），
			// 所以照旧由调用方给 —— 给了就只读探一次，录进身份里用来发现"程序重编译过"。
			BinaryRef: strings.TrimSpace(recStartBinary),
			SSH:       dcfg.SSH, DataDir: loc.DataDir(), RecordDir: recDir, Tag: tag,
			GDCLog: filepath.Join(recDir, tag+"-gdc.log"),
			NoJob:  true,
		}, logf)
		if err != nil {
			return mapOpenErr(err)
		}
		fgl := rt.FGLServer()
		if fgl == "" {
			_ = rt.Close()
			return fail(exitPreflight, "算不出这一轮的画面接线（FGLSERVER）：加 --fgserver-ip，或给环境配 SSH host")
		}
		logf("这一轮画面接线：FGLSERVER=%s（调试线按它把作业画到本代理上）", fgl)

		// ② 让调试线把作业挂上去 ③ 自动放行
		if serr := runRecStartSteps(cmd, recStartSteps(dcfg.EnvName(), ent, recStartProg, recStartModule, fgl), nil, recStartNoCont); serr != nil {
			// 半路失败：把已经起来的代理收掉（它占着 GDC/代理/注入三个端口）
			if cerr := rt.Close(); cerr != nil {
				logf("收代理时还有失败：%v", cerr)
			}
			return serr
		}

		s := rt.Session
		if werr := printJSONOrText(map[string]any{
			"recording": s.Tag, "prog": s.Prog, "env": s.Env, "zone": s.Zone, "ent": s.Ent,
			"path": s.Path, "fglserver": fgl, "injectPort": s.InjectPort, "recordDir": recDir,
		}, func() {
			fmt.Printf("录制已开始：%s\n", s.Tag)
			fmt.Printf("  作业 %s（环境 %s / 区域 %s / %s）\n", s.Prog, s.Env, s.Zone, entText(s.Ent))
			fmt.Printf("  启动路径 %s（我们自己的调试线，用户面前只有作业窗口）\n", s.Path)
			fmt.Printf("  画面接线 FGLSERVER=%s\n", fgl)
			if recStartNoCont {
				fmt.Println("  ⚠ --no-continue：作业还停在调试器里，下好断点后跑 tt debug exec \"continue\" 放行")
			}
			fmt.Println("  取这一轮：tt fep rec show " + s.Tag)
			fmt.Println("收工：Ctrl-C，或另开终端跑 tt fep session close（收工落身份并入库）")
		}); werr != nil {
			return werr
		}

		waitErr := rt.Wait(ctx)
		closeErr := rt.Close()
		if waitErr != nil {
			return fail(exitFailed, "%v", waitErr)
		}
		if closeErr != nil {
			return fail(exitFailed, "%v", closeErr)
		}
		fmt.Fprintln(cmd.ErrOrStderr(), "· 已收工：身份已落，录制已入长期库（无残留）")
		return nil
	},
}

// recStartSteps 是"把作业挂到代理上"要跑的调试线命令。**顺序不能换**：
// 先点名环境（否则可能挂到上一次的环境上），再点名企业，最后起作业。
//
// 拆成纯函数是为了能钉住它：少一条或顺序反了，症状都是**安静**的
// （界面不出来、或作业跑在别的企业上），而那种错在真机上才发现代价很大。
func recStartSteps(env, ent, prog, module, fglserver string) [][]string {
	steps := [][]string{{"debug", "env", env}}
	if strings.TrimSpace(ent) != "" {
		steps = append(steps, []string{"debug", "topent", ent})
	}
	start := []string{"debug", "start", prog, "--env", env, "--fglserver", fglserver}
	if m := strings.TrimSpace(module); m != "" {
		start = append(start, "-m", m)
	}
	return append(steps, start)
}

// runRecStartSteps 依次跑这些命令，再（默认）放行。
//
// 为什么走子进程而不是 import 调试那个包：调试服务的协议（HTTP API、apiBase 前缀、
// X-Actor 头）只该有一份实现，它在 internal/cli/debug 里；跨命令组 import 会把那份实现
// 变成两处共用、谁都不敢改的东西。代价是每条命令一次进程启动（毫秒级）。
func runRecStartSteps(cmd *cobra.Command, steps [][]string, breaks []string, noContinue bool) error {
	for _, st := range steps {
		if err := runSelf(cmd, st...); err != nil {
			return fail(exitFailed, "调试线这一步失败（%s）：%v", strings.Join(st, " "), err)
		}
	}
	// 断点必须在**放行之前**下：程序停在入口时下才有意义，放行之后那一段已经跑过去了。
	for _, b := range breaks {
		if err := runSelf(cmd, "debug", "exec", "break "+b); err != nil {
			return fail(exitFailed, "下断点失败（break %s）：%v", b, err)
		}
	}
	if noContinue {
		return nil
	}
	if err := runSelf(cmd, "debug", "exec", "continue", "--wait", "20"); err != nil {
		return fail(exitFailed, "放行失败（tt debug exec continue）：%v", err)
	}
	return nil
}

// selfArgs 拼出"调自己另一个子命令"的完整参数：`--config` 跟着走。
//
// 为什么必须跟着：调用方可能用的是非缺省配置（`--config D:\...\config.json`），
// 子进程默默换一份配置的后果是"作业起在另一个环境/另一个企业上"。
func selfArgs(args []string) []string {
	if common.ConfigPath == "" {
		return args
	}
	return append([]string{"--config", common.ConfigPath}, args...)
}

// runSelf 调我们自己的另一个子命令（同一条命令行界面，另一条线）。
//
// 三条要注意的：
//   - `--config` 必须跟着传（见 selfArgs）；
//   - 子进程的 stdout 并到**本命令的 stderr**：这些是"过程"，不是本命令的输出，
//     混进 stdout 会把 --format json 的输出毁掉；
//   - stdin 继承：调试线的命令有的要读终端（例如提示确认）。
func runSelf(cmd *cobra.Command, args ...string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	c := exec.Command(self, selfArgs(args)...)
	c.Stdout = cmd.ErrOrStderr()
	c.Stderr = cmd.ErrOrStderr()
	c.Stdin = os.Stdin
	// 父进程没有控制台时（`tt serve` 里那条脱离的 rec start）别让每个子进程弹一个新的
	// 控制台窗口 —— 用户实测点一下「开始录制」弹出一大堆终端标签页（见 winproc.HideConsole）。
	winproc.HideConsole(c)
	if err := c.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return fmt.Errorf("退出码 %d", ee.ExitCode())
		}
		return err
	}
	return nil
}

// runSelfOut 同 runSelf，但把子进程的 **stdout 收下来**（要解析它的 JSON 时用）。
//
// 为什么不接管道：受限沙箱里子进程开管道会被拒（EPERM），而临时文件哪都能用。
// 输出落临时文件再读回来，代价是一次磁盘写（几十 KB 级）。
func runSelfOut(cmd *cobra.Command, args ...string) ([]byte, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	// 落**本项目的暂存目录**，不用系统临时目录：沙箱里 %TEMP% 可能写不了
	// （实测 `os.CreateTemp("")` 报 Access is denied），而数据目录是本来就允许写的地方。
	scratch := ""
	if d, cerr := cacheDir(); cerr == nil {
		scratch = d
	}
	f, err := os.CreateTemp(scratch, "tt-self-*.out")
	if err != nil {
		return nil, err
	}
	path := f.Name()
	defer os.Remove(path)
	c := exec.Command(self, selfArgs(args)...)
	c.Stdout = f
	c.Stderr = cmd.ErrOrStderr()
	c.Stdin = os.Stdin
	winproc.HideConsole(c) // 同上：没有控制台的父进程下，别让子进程弹控制台窗口
	runErr := c.Run()
	cerr := f.Close()
	if runErr != nil {
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			runErr = fmt.Errorf("退出码 %d", ee.ExitCode())
		}
		return nil, runErr
	}
	if cerr != nil {
		return nil, cerr
	}
	return os.ReadFile(path)
}

func init() {
	recStartCmd.Flags().StringVar(&recStartEnv, "env", "", "环境名（必填 —— 环境一律点名，没有默认环境可回落）")
	recStartCmd.Flags().StringVar(&recStartProg, "prog", "", "作业程序名（必填，如 bsft001_wf）")
	recStartCmd.Flags().StringVar(&recStartEnt, "ent", "", "企业编号（ENT）：作业落在哪个企业；不给 = 沿用登录 shell 已有的 $TOPENT")
	recStartCmd.Flags().StringVarP(&recStartModule, "module", "m", "asf", "T100 模块目录名（调试线要用）")
	recStartCmd.Flags().StringVar(&recStartBinary, "binary-ref", "",
		"作业编译件（.42r）在服务器上的绝对路径（可选）：给了就只读探一次，记进录制身份")
	recStartCmd.Flags().BoolVar(&recStartNoCont, "no-continue", false,
		"先别放行：作业停在调试器里，等你自己 tt debug exec 下好断点再 continue")
	recCmd.AddCommand(recStartCmd)
}
