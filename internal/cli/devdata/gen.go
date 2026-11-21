// Package devdata 是 `tt gen` 命令组：为 model 准备原材料。
//
// 与 `tt file mirror`（读代码用的源码镜像）**不是一回事**，帮助文案里必须讲清楚：
//
//	tt file mirror   求"查得快"：全量源码/42f/42s/inc 落到 srccache/，会被 tt cache 清理
//	tt gen …        求"成套、可追溯"：按程序一套、带 manifest 指纹，不归 tt cache 管
//
// 领域层在 internal/devdata（与 Web 同一份实现）；本包只做参数解析、输出与退出码。
//
// 退出码：0 成功 / 2 用法或身份错（环境不存在、程序名非法、模块推断不出）/
// 4 未拉取或目标不存在 / 5 拉取失败（网络、库、写盘）。
package devdata

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/devdata"
	"github.com/luyaogod/TT/internal/output"
)

// 退出码（本组自己的码表，见包注释）。
const (
	exitOK       = 0
	exitUsage    = 2
	exitNotFound = 4
	exitFailed   = 5
)

// Register 把 `tt gen` 挂到根命令上。
func Register(root *cobra.Command) {
	gen := &cobra.Command{
		Use:   "gen",
		Short: "开发资料：为 model 准备原材料（与 tt file mirror 不是一回事）",
		Long: `为 model（生成器语义层）准备原材料：规格库、模板树、按程序成套的设计档。

它**不是**"读代码用的镜像"：
  tt file mirror   把源码/42f/42s 全量拉下来方便阅读，落 srccache/，会被 tt cache 清理
  tt gen …         按程序成套、每份材料带 manifest 指纹，落 devdata/<环境>/，
                   清理由 tt gen clean 负责（不归 tt cache 管）

两类材料：
  共用资料（--only dict|mdl）  53 张 DZ*_T 规格表 → SQLite；$TOP/erp/mdl 模板树
  按需资料（fetch <程序>）      蓝图/表单规格/加点/排序骨架/产物/引子，一次拉齐成套

环境一律点名（--at <环境> 或 --env），没有默认环境。

退出码：0 成功 / 2 用法或身份错 / 4 未拉取或目标不存在 / 5 拉取失败`,
	}
	gen.AddCommand(newStatusCmd(), newMirrorCmd(), newFetchCmd(), newLsCmd(), newCheckCmd(), newCleanCmd())
	root.AddCommand(gen)
}

// ---------- status ----------

func newStatusCmd() *cobra.Command {
	var probe bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "看共用资料与各程序的状态（默认不出网）",
		Long: `输出某环境的开发资料状态：规格表/模板的同步时间与大小、各程序的拉取时间与
是否过期（stale）、当前在跑的任务。

默认**不出网**：只读本地 manifest 与文件，所以打开即快。
--probe 会向服务器核对 mtime 与大小（慢，但能发现"服务器那边更新了"）。

示例：
  tt gen status --at 福联测试tst
  tt gen status --at 福联测试tst --probe --human`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				st, err := devdata.Status(context.Background(), devdata.StatusRequest{
					CfgPath: cfgPath, Env: env, Probe: probe,
				})
				if err != nil {
					return usageOrFail(err)
				}
				return emitStatus(st)
			})
		},
	}
	cmd.Flags().BoolVar(&probe, "probe", false, "向服务器核对文件是否更新（默认只读本地）")
	return cmd
}

// ---------- mirror ----------

func newMirrorCmd() *cobra.Command {
	var only string
	var force bool
	cmd := &cobra.Command{
		Use:   "mirror",
		Short: "同步共用资料（规格表 + 模板树）",
		Long: `把本环境的规格表（53 张 DZ*_T → SQLite）与模板树（$TOP/erp/mdl）拉到
devdata/<环境>/ 下。已同步过就跳过（幂等），--force 才重拉。

进度打在 stderr（JSON 输出不受影响）；Ctrl+C 可中断，中断后不留半套材料。

示例：
  tt gen mirror --at 福联测试tst
  tt gen mirror --at 福联测试tst --only dict     # 只刷规格表
  tt gen mirror --at 福联测试tst --only mdl      # 只刷模板
  tt gen mirror --at 福联测试tst --force`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				ctx, stop := interruptContext()
				defer stop()
				res, err := devdata.Mirror(ctx, devdata.MirrorRequest{
					CfgPath: cfgPath, Env: env, Ent: entOf(cmd), Only: only, Force: force,
					TTVersion: common.Version,
				}, stderrProgress())
				if err != nil {
					return failErr(err)
				}
				return emitMirror(res)
			})
		},
	}
	cmd.Flags().StringVar(&only, "only", "", "只刷一半：dict（规格表）或 mdl（模板）")
	cmd.Flags().BoolVar(&force, "force", false, "已有材料也重拉（默认幂等跳过）")
	cmd.Flags().Int("ent", 0, "企业编号（决定用哪个数据库账号；缺省用账号列表首项）")
	return cmd
}

// ---------- fetch ----------

func newFetchCmd() *cobra.Command {
	var module string
	var withDesign, form, force bool
	cmd := &cobra.Command{
		Use:   "fetch <程序>",
		Short: "按程序拉一套材料",
		Long: `把一个程序的材料一次拉齐：蓝图 .tab、表单规格 .tsd、加点 .tap、排序骨架 .tgl
（含 .tgx/.tgx2）、产物 .4gl、编译引子 _mask/_rep.4gl。缺可选项只记进 manifest 的
missing[]，缺必需项（tab/tgl/4gl）整次失败、不留半套。

模块默认从程序名前三位推断（axmt500_wf → axm）；推断不出时用 --module 指定。
已拉过的程序要 --force 才覆盖。

示例：
  tt gen fetch axmt500_wf --at 福联测试tst
  tt gen fetch axmt500_wf --at 福联测试tst --with-design --form
  tt gen fetch cxm999 --at 福联测试tst --module cxm --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				ctx, stop := interruptContext()
				defer stop()
				res, err := devdata.Fetch(ctx, devdata.FetchRequest{
					CfgPath: cfgPath, Env: env, Prog: args[0], Module: module, Ent: entOf(cmd),
					WithDesign: withDesign, Form: form, Force: force, TTVersion: common.Version,
				}, stderrProgress())
				if err != nil {
					return failErr(err)
				}
				return emitFetch(res)
			})
		},
	}
	cmd.Flags().StringVar(&module, "module", "", "模块目录名（缺省从程序名前三位推断）")
	cmd.Flags().BoolVar(&withDesign, "with-design", false, "把本地工作区的设计稿（*.tzs）一并归档")
	cmd.Flags().BoolVar(&form, "form", false, "连表单 .42f 一起拉（二进制，体积大）")
	cmd.Flags().BoolVar(&force, "force", false, "已拉过也覆盖（整包替换）")
	cmd.Flags().Int("ent", 0, "企业编号（写进 manifest 的来源信息）")
	return cmd
}

// ---------- ls ----------

func newLsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "列出已拉取的程序（小表）",
		Long: `只列本机上已经拉过的程序：大小、拉取时间、是否过期。等价 status 里的程序清单。

示例：
  tt gen ls --at 福联测试tst
  tt gen ls --at 福联测试tst --human`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				st, err := devdata.Status(context.Background(), devdata.StatusRequest{CfgPath: cfgPath, Env: env})
				if err != nil {
					return usageOrFail(err)
				}
				return emitProgs(st)
			})
		},
	}
	return cmd
}

// ---------- check ----------

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <程序>",
		Short: "只与服务器核对，输出过期明细",
		Long: `不重拉，只比对服务器上这些文件的 mtime/大小与 manifest 里记的源信息，
给出"哪里变了、为什么判过期"。

触发项：源文件变新、大小变化、源文件消失、模板/骨架版本变化、本地文件被改。

示例：
  tt gen check axmt500_wf --at 福联测试tst
  tt gen check axmt500_wf --at 福联测试tst --human`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				st, err := devdata.Check(context.Background(), devdata.StatusRequest{
					CfgPath: cfgPath, Env: env, Prog: args[0],
				})
				if err != nil {
					return failErr(err)
				}
				return emitCheck(st)
			})
		},
	}
	return cmd
}

// ---------- clean ----------

func newCleanCmd() *cobra.Command {
	var prog string
	var yes bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "删除开发资料（默认只看不删）",
		Long: `删除某环境（或某程序）的开发资料。**默认 dry-run**：只列出将删的内容与体积，
--yes 才真删。

为什么默认不删：这里面可能是某环境唯一一份已归档的蓝图，删了要重新拉（几十 MB +
数据库查询）。清理不归 tt cache 管（那份清单里没有 devdata）。

示例：
  tt gen clean --at 福联测试tst                        # 看会删什么
  tt gen clean --at 福联测试tst --prog axmt500_wf      # 只看这个程序
  tt gen clean --at 福联测试tst --prog axmt500_wf --yes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				plan, err := devdata.PlanClean(cfgPath, env, prog)
				if err != nil {
					return failErr(err)
				}
				if !yes {
					return emitCleanPlan(plan, true)
				}
				if err := devdata.ExecuteClean(plan); err != nil {
					return failErr(err)
				}
				return emitCleanPlan(plan, false)
			})
		},
	}
	cmd.Flags().StringVar(&prog, "prog", "", "只删这个程序的材料（缺省 = 整个环境）")
	cmd.Flags().BoolVar(&yes, "yes", false, "确认删除（默认只打印将删内容）")
	return cmd
}

// ---------- 公共 ----------

// entOf 取本命令的 --ent（由 --at <环境>@<企业> 或 --ent 写入）。
//
// 读的是命令自己的 flag 而不是全局变量：`--at` 的落点就是"命令声明了 --ent 就写进去"
// （见 internal/cli/at 的 Apply），所以这里读同一个来源，不会出现两处各记一份。
func entOf(cmd *cobra.Command) int {
	if cmd.Flags().Lookup("ent") == nil {
		return 0
	}
	n, err := cmd.Flags().GetInt("ent")
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// withEnv 解析配置路径与环境名，交给 fn。环境名必须能解析出来（--at 或 --env）。
func withEnv(cmd *cobra.Command, fn func(cfgPath, env string) error) error {
	cfgPath, err := common.ResolveConfig(false)
	if err != nil {
		return output.Errorf("USAGE", exitUsage, "%v", err)
	}
	env := strings.TrimSpace(common.Env)
	if env == "" {
		return output.Errorf("USAGE", exitUsage,
			"未指定环境：用 --at <环境> 或 --env <环境>（不知道有什么：加 --at-list）")
	}
	return fn(cfgPath, env)
}

// interruptContext 让长跑命令能被 Ctrl+C 打断（打断后领域层会清掉临时目录）。
func interruptContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, func() {
		signal.Stop(ch)
		cancel()
	}
}

// stderrProgress 把进度打到 stderr（stdout 保留给 JSON/表格）。
//
// 只在节流后打印：一张表一行太多（53 张表 + 280 个模板文件），改成每 400ms 最多一行。
func stderrProgress() devdata.ProgressFunc {
	last := time.Time{}
	lastMsg := ""
	return func(p devdata.Progress) {
		// 阶段切换一定要打（否则看起来像卡住）
		phaseChanged := p.Phase != lastMsg && p.Message != ""
		if !phaseChanged && time.Since(last) < 400*time.Millisecond {
			return
		}
		last = time.Now()
		lastMsg = p.Phase
		msg := p.Message
		if p.Total > 0 {
			msg += fmt.Sprintf("  [%d/%d]", p.Done, p.Total)
		}
		fmt.Fprintln(os.Stderr, msg)
	}
}

// failErr 把领域错误映射成退出码与信封。
func failErr(err error) error {
	switch {
	case errors.Is(err, devdata.ErrEnvNotFound), errors.Is(err, devdata.ErrInvalidEnv),
		errors.Is(err, devdata.ErrInvalidProg), errors.Is(err, devdata.ErrInvalidOpt),
		errors.Is(err, devdata.ErrModuleUnknown):
		return output.Errorf("USAGE", exitUsage, "%v", err)
	case errors.Is(err, devdata.ErrProgNotFound):
		return output.Errorf("NOT_FOUND", exitNotFound, "%v", err)
	case errors.Is(err, devdata.ErrExists):
		return output.Errorf("EXISTS", exitUsage, "%v", err)
	case errors.Is(err, context.Canceled):
		return output.Errorf("CANCELLED", exitFailed, "已中断；未完成的材料没有落盘")
	default:
		return output.Errorf("FAILED", exitFailed, "%v", err)
	}
}

// usageOrFail 只用于只读汇总（status/ls）：它们不会因为"没拉过"而失败。
func usageOrFail(err error) error {
	switch {
	case errors.Is(err, devdata.ErrEnvNotFound), errors.Is(err, devdata.ErrInvalidEnv),
		errors.Is(err, devdata.ErrInvalidProg), errors.Is(err, devdata.ErrInvalidOpt):
		return output.Errorf("USAGE", exitUsage, "%v", err)
	default:
		return output.Errorf("FAILED", exitFailed, "%v", err)
	}
}
