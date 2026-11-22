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
	"strconv"
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
  打包（pack <程序>）          把已拉到的材料打成 .tzc 夹具（纯本地，复刻设计器下载）

环境一律点名（--at <环境> 或 --env），没有默认环境。

退出码：0 成功 / 2 用法或身份错 / 4 未拉取或目标不存在 / 5 拉取失败`,
	}
	gen.AddCommand(newStatusCmd(), newMirrorCmd(), newFetchCmd(), newMapCmd(), newGrepCmd(), newShowCmd(), newAtCmd(), newFieldsCmd(), newPackCmd(), newLsCmd(), newCheckCmd(), newCleanCmd())
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
  tt gen status --at 测试环境
  tt gen status --at 测试环境 --probe --human`,
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
  tt gen mirror --at 测试环境
  tt gen mirror --at 测试环境 --only dict     # 只刷规格表
  tt gen mirror --at 测试环境 --only mdl      # 只刷模板
  tt gen mirror --at 测试环境 --force`,
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

模块默认从程序名前三位推断（axmt500_wf → axm）。客制程序会自动优先：标准模块
有 a→c 的客制模块（axm→cxm）、且客制模块下有该程序的蓝图时，取客制那份，并在
结果里给出 warning；要标准材料用 --module 指定。
已拉过的程序要 --force 才覆盖。

示例：
  tt gen fetch axmt500_wf --at 测试环境
  tt gen fetch axmt500_wf --at 测试环境 --with-design --form
  tt gen fetch axmt500_wf --at 测试环境 --module axm   # 两份都有时取标准`,
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
	cmd.Flags().StringVar(&module, "module", "", "模块目录名（缺省从前三位推断；客制模块 a→c 自动优先）")
	cmd.Flags().BoolVar(&withDesign, "with-design", false, "把本地工作区的设计稿（*.tzs）一并归档")
	cmd.Flags().BoolVar(&form, "form", false, "连表单 .42f 一起拉（二进制，体积大）")
	cmd.Flags().BoolVar(&force, "force", false, "已拉过也覆盖（整包替换）")
	cmd.Flags().Int("ent", 0, "企业编号（写进 manifest 的来源信息）")
	return cmd
}

// ---------- map ----------

func newMapCmd() *cobra.Command {
	var field string
	var customOnly, orphans, full, rebuild bool
	cmd := &cobra.Command{
		Use:   "map <程序> [--field <字段>|--custom|--orphans]",
		Short: "建「蓝图 ↔ 代码」索引（.tab → .tap 加点 → .4gl 行号 → 切片模板）",
		Long: `把已拉到的三份材料对成一张表，落到材料目录的 link.json：

  .tab 页面/字段  →  .tap 槽位（加点）  →  .4gl 行号区间  →  切片 aXX(Version:N)

**纯本地**，不连服务器（材料先 tt gen fetch）。产出供按"字段/页面/行号/内容"检索：
某个字段的代码在哪一行、这段是客制还是模板正文（重产会不会没）、是哪个切片生成的哪一版。

身份规则与生成器一致：槽位名 = <功能>.<位置>.<页 id>.<字段>，页用的是 .tab 的 page_id
（不是页签序号）；客制点可挂到标准槽位上（mapping ≠ name）；重名按出现次序区分。

输出（默认是摘要 + 检查，不吐整份索引）：
  --field <字段>  只列这个字段的点（点名后缀/子串，与 gen show 同一套匹配）
  --custom        只列客制点
  --orphans       只列"写了不生效"的孤儿声明（名字/原因/行数）
  --full          输出整份 link.json（几百 KB，脚本用；人读别用）

同时做一致性检查（每条带 level）：
  sections-pair   两侧 section 同名同序
  slots-pair      槽位一一对应（只在 tap = 4gl 旧；只在 4gl = 生成器自加的注释点）
  decls-orphan    声明了但骨架里没有槽位、4gl 里也找不到 = 写了不生效
  decls-unmarked  function.* 这类以函数形式落地、没有 #add-point 标记的点
  slice-version   产物里的模板版次 vs 当前模板（要先 tt gen mirror --only mdl）
  freshness       .4gl 是否比 .tap 旧（改了 tap 没跑 F7 的直证）
  blueprint       .tab 的版型/页数/客制版次（cusver=0000 即从未客制）

示例：
  tt gen map axmt500_wf --at 测试环境
  tt gen map axmt500_wf --at 测试环境 --field xmdduc001
  tt gen map axmt500_wf --at 测试环境 --orphans --human
  tt gen map axmt500_wf --at 测试环境 --full > link.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				req := devdata.LinkRequest{CfgPath: cfgPath, Env: env, Prog: args[0]}
				// 默认"确保索引是最新的"：源文件 sha 未变就复用（60ms 级），变了自动重建。
				// --rebuild 强制重解析（改了解析逻辑、或怀疑索引坏了时用）。
				var doc *devdata.LinkDoc
				var err error
				if rebuild {
					doc, err = devdata.Map(req)
				} else {
					doc, err = devdata.LoadIndex(req)
				}
				if err != nil {
					return failErr(err)
				}
				switch {
				case orphans:
					return emitOrphans(doc)
				case field != "" || customOnly:
					return emitMapPoints(doc, field, customOnly)
				default:
					return emitMap(doc, full)
				}
			})
		},
	}
	cmd.Flags().StringVar(&field, "field", "", "只列这个字段的点（后缀/子串匹配）")
	cmd.Flags().BoolVar(&customOnly, "custom", false, "只列客制点（src=c 且有内容）")
	cmd.Flags().BoolVar(&orphans, "orphans", false, "只列写了不生效的孤儿声明")
	cmd.Flags().BoolVar(&rebuild, "rebuild", false, "强制重建索引（默认源文件没变就复用）")
	cmd.Flags().BoolVar(&full, "full", false, "输出整份 link.json（默认只给摘要，避免几百 KB 灌到 stdout）")
	return cmd
}

// ---------- grep ----------

func newGrepCmd() *cobra.Command {
	var limit int
	var caseSensitive bool
	cmd := &cobra.Command{
		Use:   "grep <程序> <子串>",
		Short: "在 .4gl 里检索，并把每条命中标上归属（客制/标准、行号、section、切片）",
		Long: `这是"第一个动作永远是搜一下"的那条命令：比 grep 多给三样东西 ——
命中落在哪个加点、是客制还是模板（重产会不会保留）、属于哪个 section / 切片。

--limit 控制打印条数（默认 50），命中总数照样给；--case 区分大小写。

示例：
  tt gen grep axmt500_wf "SET(" --at 测试环境
  tt gen grep axmt500_wf xmdducsite --at 测试环境 --human
  tt gen grep axmt500_wf "cl_chk_exist" --at 测试环境 --limit 200`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				res, err := devdata.Grep(devdata.GrepRequest{
					LinkRequest: devdata.LinkRequest{CfgPath: cfgPath, Env: env, Prog: args[0]},
					Pattern:     args[1], Limit: limit, Case: caseSensitive,
				})
				if err != nil {
					return failErr(err)
				}
				return emitGrep(res)
			})
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "最多打印多少条命中（总数照给）")
	cmd.Flags().BoolVar(&caseSensitive, "case", false, "区分大小写")
	return cmd
}

// ---------- show ----------

func newShowCmd() *cobra.Command {
	var maxLines int
	cmd := &cobra.Command{
		Use:   "show <程序> <点名|字段>",
		Short: "按点名取内容（可只给字段名）",
		Long: `在 link.json 里找点位并把**代码内容**打出来。query 可以是完整点名
（input.a.page73.xmdduc001），也可以只给字段名（xmdduc001）或点名后缀。

匹配三级：精确 → 按 ".后缀" → 子串；每条结果带来源（客制/标准/生成器自加）、
行区间、所在 section、涉及切片。

示例：
  tt gen show axmt500_wf input.a.page73.xmdduc001 --at 测试环境
  tt gen show axmt500_wf xmdduc001 --at 测试环境 --human
  tt gen show axmt500_wf insert_b.m_insert10 --at 测试环境 --max 60`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				res, err := devdata.Show(devdata.ShowRequest{
					LinkRequest: devdata.LinkRequest{CfgPath: cfgPath, Env: env, Prog: args[0]},
					Query:       args[1], MaxLines: maxLines,
				})
				if err != nil {
					return failErr(err)
				}
				return emitShow(res)
			})
		},
	}
	cmd.Flags().IntVar(&maxLines, "max", 200, "每个点的内容最多打印多少行")
	return cmd
}

// ---------- at ----------

func newAtCmd() *cobra.Command {
	var context int
	cmd := &cobra.Command{
		Use:   "at <程序>:<行号>",
		Short: "反查一行属于哪个点/段/切片（报错行定位）",
		Long: `给一个 .4gl 行号，回答：这在哪个加点上、是不是客制（重产会不会保留）、
属于哪个 section、由哪个切片模板生成。

编译报错（fglcomp/cxm_xxx.4gl:44021）、运行时断点、review 时看到的一行，都能直接用。

示例：
  tt gen at axmt500_wf:44107 --at 测试环境
  tt gen at axmt500_wf 44107 --at 测试环境 --context 5`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			prog, line, err := parseProgLine(args)
			if err != nil {
				return output.Errorf("USAGE", exitUsage, "%v", err)
			}
			return withEnv(cmd, func(cfgPath, env string) error {
				res, aerr := devdata.At(devdata.AtRequest{
					LinkRequest: devdata.LinkRequest{CfgPath: cfgPath, Env: env, Prog: prog},
					Line:        line, Context: context,
				})
				if aerr != nil {
					return failErr(aerr)
				}
				return emitAt(res)
			})
		},
	}
	cmd.Flags().IntVar(&context, "context", 2, "附带几行上下文")
	return cmd
}

// parseProgLine 解析 `<程序>:<行号>` 或 `<程序> <行号>`。
func parseProgLine(args []string) (string, int, error) {
	prog, num := "", ""
	if len(args) == 2 {
		prog, num = args[0], args[1]
	} else if i := strings.LastIndex(args[0], ":"); i > 0 {
		prog, num = args[0][:i], args[0][i+1:]
	} else {
		return "", 0, fmt.Errorf("要写成 <程序>:<行号>（如 axmt500_wf:44107）")
	}
	n, err := strconv.Atoi(strings.TrimSpace(num))
	if err != nil || n < 1 {
		return "", 0, fmt.Errorf("行号 %q 不是正整数", num)
	}
	return strings.TrimSpace(prog), n, nil
}

// ---------- fields ----------

func newFieldsCmd() *cobra.Command {
	var page, pattern string
	var pages bool
	cmd := &cobra.Command{
		Use:   "fields <程序> [--pages]",
		Short: "列蓝图字段（page_id + 字段 → 它的点位），或列页面",
		Long: `列出 .tab 里真实存在的字段，以及与每个字段关联的点位（a/b/c/g 各位置）。

**只列"有加点关联"的字段**：纯数据字段（.tab 里有、代码里没有对应加点）不出现，
所以这不是"这张页面有几个栏位"的清单，而是"哪些栏位有逻辑"。要看全部栏位数用 gen map 的
stats.columnsTotal。

--page 同时接受 page_id（73）与页签序号（12），传序号会明确告诉你对应关系。
--pages 改成列页面（序号 / page_id / record / 栏位数 / 有逻辑的字段数），
用于"客户说的第 12 个页签到底是哪个 page_id"。

示例：
  tt gen fields axmt500_wf --at 测试环境 --pages
  tt gen fields axmt500_wf --at 测试环境 --page 73
  tt gen fields axmt500_wf --at 测试环境 --page 12
  tt gen fields axmt500_wf --at 测试环境 --q xmdd`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				fs, doc, err := devdata.Fields(devdata.FieldsRequest{
					LinkRequest: devdata.LinkRequest{CfgPath: cfgPath, Env: env, Prog: args[0]},
					Page:        page, Pattern: pattern,
				})
				if err != nil {
					return failErr(err)
				}
				if pages {
					return emitPages(args[0], doc)
				}
				return emitFields(args[0], fs, doc)
			})
		},
	}
	cmd.Flags().StringVar(&page, "page", "", "page_id 或页签序号（两者都接受）")
	cmd.Flags().StringVar(&pattern, "q", "", "按字段名子串过滤")
	cmd.Flags().BoolVar(&pages, "pages", false, "列出页面（序号/page_id/record/栏位数）而不是字段")
	return cmd
}

// ---------- pack ----------

func newPackCmd() *cobra.Command {
	var identity, out, ver string
	var force bool
	cmd := &cobra.Command{
		Use:   "pack <程序>",
		Short: "把已拉到的材料打成 .tzc（纯本地，复刻设计器下载）",
		Long: `用本地已拉取的材料（.4gl / .tap / .tgl + ver）重新打出一个 .tzc，命名同设计器
下载的形态：<程序>(c).tzc 或 <程序>(s).tzc。

**纯本地**：不复刻服务器动作，也不需要服务器。设计器的"下载程式包"
（adzp050_download_code_pack）本来就只是压缩+传输，不重产、不读库；包内成员
就是 dzx 里的 tap/tgl 加 4gl 与一个恒为 1.0 的 ver。所以材料到手之后，这一步
在本地做出来的包与下载来的同构（flags、CRC、extra 都按真包形态写）。

用途：给 tt dev tzc 造测试夹具、留一份可回放的快照。

示例：
  tt gen pack axmt500_wf --at 测试环境
  tt gen pack axmt500_wf --at 测试环境 --identity s -o D:\tmp\
  tt gen pack axmt500_wf --at 测试环境 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withEnv(cmd, func(cfgPath, env string) error {
				res, err := devdata.Pack(devdata.PackRequest{
					CfgPath: cfgPath, Env: env, Prog: args[0],
					Identity: identity, Out: out, Ver: ver, Force: force,
				})
				if err != nil {
					return failErr(err)
				}
				return emitPack(res)
			})
		},
	}
	cmd.Flags().StringVar(&identity, "identity", "", "身份：c（客制，缺省从 .tap 读）/ s（标准）")
	cmd.Flags().StringVarP(&out, "out", "o", "", "目标文件或目录（缺省 <环境>/pack/<程序>(<身份>).tzc）")
	cmd.Flags().StringVar(&ver, "ver", "", "ver 条目内容（缺省 1.0，与设计器一致）")
	cmd.Flags().BoolVar(&force, "force", false, "目标已存在时覆盖")
	return cmd
}

// ---------- ls ----------

func newLsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "列出已拉取的程序（小表）",
		Long: `只列本机上已经拉过的程序：大小、拉取时间、是否过期。等价 status 里的程序清单。

示例：
  tt gen ls --at 测试环境
  tt gen ls --at 测试环境 --human`,
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
  tt gen check axmt500_wf --at 测试环境
  tt gen check axmt500_wf --at 测试环境 --human`,
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
  tt gen clean --at 测试环境                        # 看会删什么
  tt gen clean --at 测试环境 --prog axmt500_wf      # 只看这个程序
  tt gen clean --at 测试环境 --prog axmt500_wf --yes`,
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
	case errors.Is(err, devdata.ErrProgNotFound), errors.Is(err, devdata.ErrMaterialIncomplete),
		errors.Is(err, devdata.ErrPointNotFound):
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
