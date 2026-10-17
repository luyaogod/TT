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