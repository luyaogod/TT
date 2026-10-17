package fep

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/debug"
	"tt/internal/debugclient"
	"tt/internal/fep"
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
		// 开跑前一次查完四项前置（目录可写 / 服务 / 会话 / 环境）。
		//
		// 位置在**起代理之前**：代理占三个端口，半路失败要收，收不干净就留一个空会话；
		// 而且真机冒烟里失败信息是逐层挤出来的（先"服务没起"，起了再"没有会话"）——
		// 每跑一次才知道下一个前置，使用者被迫来回试错。这里一次列完。
		if err := recStartPreflight(cmd, recStartProg, recStartEnv); err != nil {
			return err
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
		// D29：没给 `--ent` 时用**环境自己的 topent**（配置里的 db.ent）。
		//
		// 为什么必须补：录制身份里 ENT 空 = "未指定"，可这一轮明明知道是哪个企业
		// （环境配置里就写着 topent）。以前只在 `--ent` 显式给了才落值，于是
		// `--env 主机109` 录出来的件里写着"ENT-未指定"，事后追溯环境就少了一半信息。
		// 取不到（topent 未配）就仍然是空 —— 不猜（A15）。
		if ent == "" {
			if t := dcfg.TopentInt(); t > 0 {
				ent = strconv.Itoa(t)
			}
		}
		// 到这里还是空 = 谁都没说这一轮跑在哪个企业。**不许再往下走**：调试线那边已经
		// 不再沿用登录 shell 的 $TOPENT（requireExplicitEnt），这里提前一句说清楚，
		// 免得先起代理占了三个端口、再在"起作业"那一步失败。
		if ent == "" {
			return fail(exitUsage, "录制必须点名企业编号(ENT)：tt fep rec start --env <环境> --ent <企业编号> --prog <作业>"+
				"（不再沿用登录 shell 的 $TOPENT —— 那会让「忘了点名」与「点错了」长得一样）")
		}
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
		// alias（T5.4 / C7.3）：把这一轮的录制身份当**固定别名**传进子进程 ——
		// 子进程之后靠它落回同一条会话（Q23：别名空闲就复用），而不是靠"有会话就收"。
		alias := recSessionAlias(tag)
		cl, cerr := debugClient()
		if cerr != nil {
			_ = rt.Close()
			return fail(exitFailed, "连不上调试面：%v", cerr)
		}
		ref, serr := runRecPlan(cl, recRunPlan{
			Alias: alias, Env: dcfg.EnvName(), Ent: ent,
			Prog: recStartProg, Module: recStartModule, FGL: fgl, NoContinue: recStartNoCont,
		}, logf)
		if serr != nil {
			// 半路失败：把已经起来的代理收掉（它占着 GDC/代理/注入三个端口）
			if cerr := rt.Close(); cerr != nil {
				logf("收代理时还有失败：%v", cerr)
			}
			return serr
		}
		// 把"落到哪条调试会话"写进登记（P3）：两条线从此认同一个地址。
		if rerr := rt.SetRef(ref); rerr != nil {
			logf("回填会话地址失败（不影响本轮录制）：%v", rerr)
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

// recRunPlan 是"把作业挂到代理上"要做的动作序列。**顺序不能换**：
//
//	① 收口自己上一轮（按别名找；没有就跳过）
//	② 点名环境并命名 → 等到 idle
//	③ 定 ENT（会话级）
//	④ 起作业（复用 ② 建好的那条宿主）
//	⑤ 下断点（必须在放行之前）
//	⑥ 放行
//
// 为什么是"计划"而不是"命令列表"：这套动作从前是靠**起子进程** `tt debug …` 表达的，
// 于是顺序与参数全在 CLI 层用 argv 拼；现在直连调试面（见 `internal/debugclient`），
// 计划就是数据，执行只有一处。
type recRunPlan struct {
	Alias      string
	Env        string
	Ent        string
	Prog       string
	Module     string
	FGL        string
	Breaks     []string
	NoContinue bool
}

// runRecPlan 按上面的顺序直连调试面执行。
//
// 失败信息带**是哪一步**：这条链上每一步都有过"安静失败"的现场（界面不出来、
// 作业跑在别的企业上、断点没下上），笼统一句"录制失败"会让人从头猜。
func runRecPlan(cl *debugclient.Client, p recRunPlan, logf func(string, ...any)) (string, error) {
	// ① 收口自己上一轮：调试线只在空闲时才让设 ENT，而人点「开始录制」时常常正停着一条。
	//    按别名找得到才收 —— 收不到就是"本来就没有"，不是错误。
	if a := strings.TrimSpace(p.Alias); a != "" {
		if b, rerr := cl.Resolve(a); rerr == nil {
			if qerr := cl.Quit(b.ID); qerr != nil {
				logf("收上一轮会话失败（继续）：%v", qerr)
			}
		}
	}
	// ② 点名环境并命名：后面几步要靠这个名字点到"自己那条"。
	id, err := cl.SwitchToIdle(p.Env, p.Ent, p.Alias, 120*time.Second)
	if err != nil {
		return "", fail(exitFailed, "点名环境失败（%s）：%v", p.Env, err)
	}
	// ③ 定 ENT（会话级 override）
	if e := strings.TrimSpace(p.Ent); e != "" {
		if terr := cl.Topent(id, e); terr != nil {
			return "", fail(exitFailed, "设置 ENT 失败（%s）：%v", e, terr)
		}
	}
	// ④ 起作业：同环境同别名 → 复用 ② 那条宿主，不重新登录。
	fgl := p.FGL
	sid, err := cl.Launch(debugclient.LaunchReq{
		Env: p.Env, Prog: p.Prog, Module: p.Module, Name: p.Alias, Ent: p.Ent, FGLServer: &fgl,
	})
	if err != nil {
		return "", fail(exitFailed, "起作业失败（%s）：%v", p.Prog, err)
	}
	if _, ok, werr := cl.WaitStopped(sid, 120*time.Second); werr != nil {
		return "", fail(exitFailed, "等入口停站失败：%v", werr)
	} else if !ok {
		return "", fail(exitFailed, "起作业后没等到入口停站（%s）", p.Prog)
	}
	// ⑤ 断点必须在放行之前：程序停不住，下断点就没意义。
	for _, b := range p.Breaks {
		if _, berr := cl.Exec(sid, "break "+b); berr != nil {
			return "", fail(exitFailed, "下断点失败（break %s）：%v", b, berr)
		}
	}
	// 回填地址：两条线从此认同一个 `ref`（P3 的落点）。取不到不算失败 —— 老服务可能没有。
	ref := ""
	if snap, serr := cl.Snapshot(sid); serr == nil {
		ref = str(snap["ref"])
	}
	if p.NoContinue {
		return ref, nil
	}
	// ⑥ 放行
	if cerr := cl.Continue(sid); cerr != nil {
		return ref, fail(exitFailed, "放行失败：%v", cerr)
	}
	return ref, nil
}

func init() {
	recStartCmd.Flags().StringVar(&recStartEnv, "env", "", "环境名（必填 —— 环境一律点名，没有默认环境可回落）")
	recStartCmd.Flags().StringVar(&recStartProg, "prog", "", "作业程序名（必填，如 bsft001_wf）")
	recStartCmd.Flags().StringVar(&recStartEnt, "ent", "", "企业编号（ENT，**必填**）：作业落在哪个企业。不给会回落到环境配置里的 db.ent；两者都没有则直接报错 —— 不再沿用登录 shell 的 $TOPENT")
	recStartCmd.Flags().StringVarP(&recStartModule, "module", "m", "asf", "T100 模块目录名（调试线要用）")
	recStartCmd.Flags().StringVar(&recStartBinary, "binary-ref", "",
		"作业编译件（.42r）在服务器上的绝对路径（可选）：给了就只读探一次，记进录制身份")
	recStartCmd.Flags().BoolVar(&recStartNoCont, "no-continue", false,
		"先别放行：作业停在调试器里，等你自己 tt debug exec 下好断点再 continue")
	recCmd.AddCommand(recStartCmd)
}
