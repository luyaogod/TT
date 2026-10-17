package fep

import (
	"context"
	"errors"
	"fmt"
	"os"
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
)

var (
	openProg     string
	openEnv      string
	openEnt      string
	openBinary   string
	openGDCExe   string
	openGDCPort  int
	openProxyPr  int
	openInjectPr int
	openFGLIP    string
	openNoJob    bool
	openNoKeep   bool
	openWaitGDC  time.Duration
)

var sessionOpenCmd = &cobra.Command{
	Use:   "open",
	Short: "起一个隔离会话（隔离实例 + 代理 + 打开作业），前台常驻到收工",
	Long: `起一个隔离的作业会话，全程不打扰用户：不抢焦点、不碰键鼠、只在本机回环上说话。

它做四件事，顺序不能换：
  1. 环境体检：已经在跑的 gdc.exe 若与 TT 权限不一致，直接失败（那种情况下窗口驱动不可靠）；
  2. 端口体检：GDC / 代理 / 注入口三个端口都必须空着；
  3. 起隔离实例（gdc.exe -n -M -A 0 -p <port>）与本地代理（逐连接落原始字节）；
  4. 写会话登记（.tt-fep.json）——**先登记再开作业**，这样中途被打断也能收干净。

然后经 SSH 登录、选区域、下发 TOPENT（给了 --ent 时）与 FGLSERVER=<本机>:<代理端口-6400>，
再启动作业客户端。两者都必须在 r.r 之前生效：企业编号错了的症状是"查不到数据"，
与"这个企业本来就没有数据"完全一样。

环境**必须点名**（--env），没有默认环境可回落；不给会报错并列出可用环境。
企业编号（--ent）不给则沿用登录 shell 自己的 $TOPENT（如实回显，不猜）。

作业一开着，本命令就前台常驻；Ctrl-C 或另开一个终端跑 tt fep session close 都能收工，
收工顺序是：远端作业 → 代理 → 隔离实例 → 清登记（只杀登记里的 pid）。

注意：作业窗口是**普通窗口**，用户看得见、也能自己操作；TT 不碰键鼠。`,
	Args: noArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSessionOpen(cmd, openParams{
			Env: openEnv, Prog: openProg, Ent: openEnt, Binary: openBinary,
			GDCExe: openGDCExe, GDCPort: openGDCPort, ProxyPort: openProxyPr, InjectPort: openInjectPr,
			FGLServerIP: openFGLIP, WaitGDC: openWaitGDC, NoJob: openNoJob, NoKeep: openNoKeep,
		})
	},
}


// openEnvName 解出这次要开哪个环境。**必须点名**，没有默认环境可回落。
//
// 为什么不用 debug.Config.EnvName() 兜底：它最后一级是 `SSH.Host + "-" + Zone`，
// 环境没选中时 SSH 是零值，于是拼出一个叫 `-` 的幽灵环境名 —— 报错就成了
// 「环境 "-" 不存在（见 tt env list）」，人看不出该补什么（这条实测踩过）。
// 改用统一解析器后，报的是与 tt debug / tt dict 同一句
// 「必须指定环境(--env <环境名>);可用环境: …」，拼错时也如实报错、绝不静默换环境。
func openEnvName(sshs []config.NamedSsh, name string) (string, error) {
	ref, err := (&config.Hosts{SSHs: sshs}).Resolve(name)
	if err != nil {
		return "", err
	}
	return ref.Name, nil
}

// mapOpenErr 把内核的错误映射成这条线的退出码。
func mapOpenErr(err error) error {
	switch {
	case errors.Is(err, fep.ErrUsage):
		return fail(exitUsage, "%v", err)
	case errors.Is(err, fep.ErrPreflight):
		return fail(exitPreflight, "%v", err)
	case errors.Is(err, fep.ErrPortBusy):
		return fail(exitState, "%v", err)
	default:
		return fail(exitFailed, "%v", err)
	}
}

func init() {
	sessionOpenCmd.Flags().StringVar(&openProg, "prog", "", "作业程序名（必填，如 axmt500_wf）")
	sessionOpenCmd.Flags().StringVar(&openEnv, "env", "", "环境名（config.json hosts.sshs 里的 name；必填 —— 环境一律点名，没有默认环境可回落）")
	sessionOpenCmd.Flags().StringVar(&openEnt, "ent", "", "企业编号（ENT）：作业落在哪个企业；不给 = 不覆盖，沿用登录 shell 已有的 $TOPENT")
	sessionOpenCmd.Flags().StringVar(&openBinary, "binary-ref", "",
		"作业编译件（.42r）在服务器上的绝对路径：给了就只读地探一次它的标识记进录制身份，\n用来发现「程序重编译过、界面一模一样」这种录制失效（r.r 路径下程序路径由登录环境决定，探不到就留空）")
	sessionOpenCmd.Flags().StringVar(&openGDCExe, "gdc-exe", "", "GDC 可执行文件（默认 C:\\Program Files\\FourJs\\gdc\\bin\\gdc.exe）")
	sessionOpenCmd.Flags().IntVar(&openGDCPort, "gdc-port", 0, "隔离实例监听端口（默认 7400）")
	sessionOpenCmd.Flags().IntVar(&openProxyPr, "proxy-port", 0, "代理监听端口（默认 7600；FGLSERVER 数值 = 它 - 6400）")
	sessionOpenCmd.Flags().IntVar(&openInjectPr, "inject-port", 0, "注入口端口（默认 7899）")
	sessionOpenCmd.Flags().StringVar(&openFGLIP, "fgserver-ip", "", "覆盖 FGLSERVER 里的本机地址（默认按路由探测）")
	sessionOpenCmd.Flags().BoolVar(&openNoJob, "no-job", false, "只起隔离实例与代理，不打开作业（排障用）")
	sessionOpenCmd.Flags().BoolVar(&openNoKeep, "no-keep", false,
		"收工**不**把这一轮录制搬进长期库（默认搬：<数据目录>/records/<环境>/<ENT>/<录制 ID>/。\n"+
			"缓存目录会被启动清理与 tt cache clean 删掉，而录制删了就造不回来）")
	sessionOpenCmd.Flags().DurationVar(&openWaitGDC, "wait-gdc", 0, "等隔离实例监听的最长时间（默认 20s）")
	sessionCmd.AddCommand(sessionOpenCmd)
}
// openParams 是"起一个隔离会话"要的那些参数（`session open` 与 `recording start --launch standard` 共用）。
//
// 用结构体而不是一堆包级变量：两个入口的默认值不同，各自组装一份交给同一个实现，
// 比"改全局变量再调用"清楚得多。
type openParams struct {
	Env, Prog, Ent, Binary, GDCExe string
	GDCPort, ProxyPort, InjectPort int
	FGLServerIP                    string
	WaitGDC                        time.Duration
	NoJob, NoKeep                  bool
}

// path 这一次作业走哪条启动路径：只走 T100 标准启动器（r.r）；--no-job 时如实报 none。
func (p openParams) path() fep.LaunchPath {
	if p.NoJob {
		return fep.PathNone
	}
	return fep.PathRunnerRR
}

// runSessionOpen 起一个隔离会话（隔离实例 + 代理 + 打开作业），前台常驻到收工。
//
// 两个入口共用它：`tt fep session open`（旧）与 `tt recording start --launch standard`（新）。
// 复制一份的后果是"两条路对同一轮录制留下不同的现场"，而现场差一点回放就用不了。
func runSessionOpen(cmd *cobra.Command, p openParams) error {

	// 用法错要在读配置、起任何东西之前就报出来
	if p.Prog == "" && !p.NoJob {
		return fail(exitUsage, "没给作业程序（--prog axmt500_wf）")
	}
	// 企业编号（ENT）：值不限数字/文本，剔除两侧空白后透传。
	// 留空 = 不覆盖，沿用登录 shell 自己的 $TOPENT（见 internal/fep 的 jobLaunchCmd）。
	ent := strings.TrimSpace(p.Ent)
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return err
	}
	dcfg, err := debug.LoadConfig(cfgPath)
	if err != nil {
		return fail(exitFailed, "%v", err)
	}
	// LoadConfig 只把环境清单读进来，运行时的 SSH/zone 要**应用一次**才生效
	// （debug 线自己会在起会话时做；这里必须显式做，否则 SSH.Host 是空的，
	// 症状是「dial tcp :22 ... refused」这种看不出根因的错）。
	// 环境一律**点名**：没有默认环境可回落（见 p.EnvName）。
	name, rerr := openEnvName(dcfg.SSHs, p.Env)
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
	launchPath := p.path()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logf := func(format string, a ...any) {
		fmt.Fprintf(os.Stderr, "· "+format+"\n", a...)
	}
	rt, err := fep.Open(ctx, fep.OpenConfig{
		Env:         dcfg.EnvName(),
		Zone:        dcfg.Zone,
		Prog:        p.Prog,
		Path:        launchPath,
		Ent:         ent,
		BinaryRef:   strings.TrimSpace(p.Binary),
		SSH:         dcfg.SSH,
		GDCExe:      p.GDCExe,
		GDCPort:     p.GDCPort,
		ProxyPort:   p.ProxyPort,
		InjectPort:  p.InjectPort,
		FGLServerIP: p.FGLServerIP,
		DataDir:     loc.DataDir(),
		RecordDir:   recDir,
		Tag:         tag,
		GDCLog:      filepath.Join(recDir, tag+"-gdc.log"),
		WaitGDC:     p.WaitGDC,
		NoJob:       p.NoJob,
		NoKeep:      p.NoKeep,
	}, logf)
	if err != nil {
		return mapOpenErr(err)
	}

	s := rt.Session
	if err := printJSONOrText(map[string]any{
		"prog": s.Prog, "env": s.Env, "zone": s.Zone, "ent": s.Ent,
		"gdcPid": s.GDCPID, "gdcPort": s.GDCPort,
		"proxyPort": s.ProxyPort, "injectPort": s.InjectPort,
		"tag": s.Tag, "registry": rt.RegistryPath(),
	}, func() {
		fmt.Printf("会话已起：%s（环境 %s / 区域 %s / 启动路径 %s）\n", s.Prog, s.Env, s.Zone, s.Path)
		if s.Ent != "" {
			fmt.Printf("  企业编号：TOPENT=%s（已下发到登录 shell）\n", s.Ent)
		} else {
			fmt.Println("  企业编号：没给 --ent，沿用登录 shell 自己的 $TOPENT")
		}
		fmt.Printf("  隔离实例 pid %d，端口 %d\n", s.GDCPID, s.GDCPort)
		fmt.Printf("  代理监听 %d（注入口 127.0.0.1:%d）\n", s.ProxyPort, s.InjectPort)
		fmt.Printf("  录制 ID  %s\n", s.Tag)
		fmt.Printf("  录制件   %s/%s-*-*.log\n", recDir, s.Tag)
		fmt.Printf("  取这一轮：tt fep rec show %s（收工时落 %s）\n", s.Tag, s.Tag+fep.MetaSuffix)
		fmt.Println("收工：Ctrl-C，或另开终端跑 tt fep session close")
	}); err != nil {
		return err
	}

	waitErr := rt.Wait(ctx)
	closeErr := rt.Close()
	if waitErr != nil {
		return fail(exitFailed, "%v", waitErr)
	}
	if closeErr != nil {
		return fail(exitFailed, "%v", closeErr)
	}
	fmt.Fprintln(os.Stderr, "· 已收工，无残留")
	return nil
}