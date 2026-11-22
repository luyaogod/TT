// Package task 是 `tt dev` 的工作区命令组：open / update / status / brief / find / at / search
// + 三个直通 tzc 的动词（apply / verify / unlock）。
//
// 这一组与前两组的分工：
//
//	tt gen …        面向"材料"：按环境成套拉取、建索引、查询（地址是 devdata/<环境>/）
//	tt dev tzc …    面向"包/工作区"：.tzc 的导出与三道闸门（地址是某个工作区目录）
//	tt dev <动词>   面向"一次开发任务"：一个自包含目录（材料+索引+视图+包），
//	                地址是**当前目录**（或 --dir），环境只在 open/update 时需要
//
// 所以这组的读命令（brief/find/at/search/status）**完全离线**：不连服务器、不读配置，
// 只读工作区里的三份文件（材料、索引、视图）。
package task

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/cli/dev/tzc"
	"github.com/luyaogod/TT/internal/dev/taskws"
	"github.com/luyaogod/TT/internal/devdata"
	"github.com/luyaogod/TT/internal/output"
)

// 退出码（本组自己的码表，与 tzc 线同形）：
//
//	0 成功 / 2 用法或身份错 / 4 未找到或写入被拒 / 5 IO·网络失败
const (
	exitOK       = 0
	exitUsage    = 2
	exitNotFound = 4
	exitRefused  = 4
	exitFailed   = 5
)

// stringSlice 是可重复的字符串选项（--only 可给多次）。
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }
func (s *stringSlice) Type() string   { return "strings" }
func (s *stringSlice) Set(v string) error {
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			*s = append(*s, p)
		}
	}
	return nil
}

// Verbs 返回本组的全部命令（由 dev 组挂到 cobra 树上）。
func Verbs() []*cobra.Command {
	return []*cobra.Command{
		newOpenCmd(), newUpdateCmd(), newStatusCmd(), newBriefCmd(),
		newFindCmd(), newAtCmd(), newSearchCmd(),
		newApplyCmd(), newVerifyCmd(), newRelinkCmd(), newUnlockCmd(),
	}
}

// ---------- open ----------

func newOpenCmd() *cobra.Command {
	var dir, name, module, identity string
	var only stringSlice
	var noMdl, form bool
	cmd := &cobra.Command{
		Use:   "open <程序>",
		Short: "建任务工作区（拉材料 + 建索引 + 打包 + 导出可编辑视图）",
		Long: `把一个程序的一个任务所需的全部输入，落成一个自包含目录：

  <目录>/task.json          身份与历史（实时状态一律现算，不存这里）
  <目录>/materials/         蓝图/表单规格/加点/骨架/产物 + manifest 指纹
  <目录>/mdl/slice/         模板切片（可选，切片脱版检测用）
  <目录>/index/link.json    蓝图 ↔ 代码 索引
  <目录>/ws/                可编辑视图（唯一编辑文件 ws/prog.full.4gl）
  <目录>/pkg/<程序>(c).tzc  可直接交给设计器的产物
  <目录>/notes/             人的备注（工具不碰）

目录默认是当前目录下的 <程序>（给了 --task 就是 <程序>-<任务名>）。
之后所有命令都在这个目录里跑（或 --dir 指过来），**不再需要 --at**。

环境必须点名（--at <环境> 或 --env）。拉材料的规则与 tt gen fetch 完全一致
（客制模块自动优先；可选项缺失只记 missing[]，必需项缺失整次失败且不留半套）。

示例：
  tt dev open axmt500_wf --at 测试环境
  tt dev open axmt500_wf --at 测试环境 --task <任务名> --dir D:\work\axmt500_wf-<任务名>
  tt dev open axmt500_wf --at 测试环境 --only xmdduc001 --no-mdl`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			env := strings.TrimSpace(common.Env)
			if env == "" {
				return output.Errorf("USAGE", exitUsage, "未指定环境：用 --at <环境>（不知道有什么：加 --at-list）")
			}
			cfgPath, err := common.ResolveConfig(false)
			if err != nil {
				return output.Errorf("USAGE", exitUsage, "%v", err)
			}
			ctx, stop := interruptContext()
			defer stop()
			res, cerr := taskws.Create(ctx, taskws.CreateRequest{
				CfgPath: cfgPath, Env: env, Prog: args[0], Module: module,
				Name: name, Dir: dir, Identity: identity, Only: only,
				NoMdl: noMdl, Form: form, Ent: entOf(cmd), TTVersion: common.Version,
				Progress: stderrProgress(),
			})
			if cerr != nil {
				return fail(cerr)
			}
			if common.IsJSON() {
				return emitJSON(env, openPayload(res), 0, 0)
			}
			printOpen(res)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：当前目录/<程序>[-<任务名>]）")
	cmd.Flags().StringVar(&name, "task", "", "任务名（拼进默认目录名，并记进 task.json）")
	cmd.Flags().StringVar(&module, "module", "", "模块目录名（缺省从前三位推断；客制模块自动优先）")
	cmd.Flags().StringVar(&identity, "identity", "", "包身份 c（客制）/ s（标准）；缺省从 .tap 推断")
	cmd.Flags().Var(&only, "only", "部分导出：只把这些点渲染成可编辑（可重复/逗号分隔）")
	cmd.Flags().BoolVar(&noMdl, "no-mdl", false, "不复制模板切片（放弃切片脱版检测）")
	cmd.Flags().BoolVar(&form, "form", false, "连表单 .42f 一起拉（二进制，体积大）")
	cmd.Flags().Int("ent", 0, "企业编号（缺省用账号列表首项）")
	return cmd
}

// openPayload 是 open 的 JSON 载荷（不吐 store.Manifest 这类大对象）。
type openPayloadT struct {
	Dir       string   `json:"dir"`
	Prog      string   `json:"prog"`
	Env       string   `json:"env"`
	Module    string   `json:"module,omitempty"`
	Identity  string   `json:"identity,omitempty"`
	Task      string   `json:"task,omitempty"`
	Materials int      `json:"materials"`
	Points    int      `json:"points"`
	Custom    int      `json:"custom"`
	Sections  int      `json:"sections"`
	Pkg       string   `json:"pkg"`
	PkgBytes  int64    `json:"pkgBytes"`
	Ws        string   `json:"ws"`
	Regions   int      `json:"regions"`
	Editable  int      `json:"editable"`
	Warnings  []string `json:"warnings,omitempty"`
}

func openPayload(res *taskws.CreateResult) openPayloadT {
	p := openPayloadT{Dir: res.Dir, Prog: res.Task.Prog, Env: res.Task.Env,
		Module: res.Task.Module, Identity: res.Task.Identity, Task: res.Task.Task,
		Materials: len(res.Fetch.Files), Pkg: res.Pack.Path, PkgBytes: res.Pack.Bytes,
		Ws: res.Export.Dir, Regions: res.Export.Regions, Editable: res.Export.Stats.Editable,
		Warnings: res.Warnings}
	if res.Index != nil {
		p.Points, p.Custom, p.Sections = res.Index.Points, res.Index.Custom, res.Index.Sections
	}
	return p
}

// ---------- update ----------

func newUpdateCmd() *cobra.Command {
	var dir string
	var full, check, noMdl, force bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "刷新工作区（材料/索引跟上服务器；有未落盘改动时拒绝）",
		Long: `向服务器核对材料是否变了，变了才重拉；然后重建索引。视图与包按"有没有活"决定：

  视图干净且包没被 apply 过  → 重新打包 + 重新导出视图（整体对齐到最新材料）
  视图有未落盘改动            → **拒绝**（先 tt dev apply / verify；或 --force 只刷材料）
  包里已有 apply 过的改动     → 只更新材料与索引（视图落后于服务器，别再在上面改）

--check 只核对不写盘；--full 不比对直接整包重拉。

示例：
  tt dev update
  tt dev update --check
  tt dev update --full`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			cfgPath, _ := common.ResolveConfig(false)
			ctx, stop := interruptContext()
			defer stop()
			res, uerr := taskws.Update(ctx, taskws.UpdateRequest{
				CfgPath: cfgPath, Dir: root, Full: full, CheckOnly: check,
				NoMdl: noMdl, Force: force, Progress: stderrProgress(),
			})
			if uerr != nil {
				return fail(uerr)
			}
			if common.IsJSON() {
				return emitJSON(res.Status.Env, res, 0, 0)
			}
			printUpdate(res)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	cmd.Flags().BoolVar(&full, "full", false, "不比对，直接整包重拉")
	cmd.Flags().BoolVar(&check, "check", false, "只向服务器核对，不写盘")
	cmd.Flags().BoolVar(&noMdl, "no-mdl", false, "不更新模板切片")
	cmd.Flags().BoolVar(&force, "force", false, "有未落盘改动也继续（只刷材料与索引，不动视图/包）")
	return cmd
}

// ---------- status ----------

func newStatusCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "看工作区状态（材料/索引/视图/包，全部现算，不出网）",
		Long: `报告工作区四件事的现状：材料是哪次拉的、索引是否与材料一致、视图有没有未落盘的
改动、包是否被 apply 过。**纯本地**，打开即快（几十毫秒）。

示例：
  tt dev status
  tt dev status --dir D:\work\axmt500_wf --human`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			st, serr := taskws.GetStatus(taskws.StatusRequest{Dir: root})
			if serr != nil {
				return fail(serr)
			}
			if common.IsJSON() {
				return emitJSON(st.Env, st, 0, 0)
			}
			printStatus(st)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	return cmd
}

// ---------- brief ----------

func newBriefCmd() *cobra.Command {
	var dir string
	var top int
	cmd := &cobra.Command{
		Use:   "brief",
		Short: "一页纸画像：判型/页签/客制逻辑/风险/下一步（离线）",
		Long: `回答刚接到任务时最先要判断的五件事：

  1. 这是什么程序（判型/模块/身份/客制版次/是否分段落版型）
  2. 有哪些页签（页序 / page_id / screen record / 栏位数 / 有逻辑的字段数）
  3. 客制逻辑在哪（按行数降序 —— 重产会保留的只有这些）
  4. 有什么风险（一致性检查的 warn/error、模板脱版、.4gl 比 .tap 旧）
  5. 下一步怎么查（意图清单 + 三条命令）

**纯本地**。默认列前 15 个客制点，--top 调整。

示例：
  tt dev brief
  tt dev brief --top 30 --human`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			t, p, lerr := taskws.Load(root)
			if lerr != nil {
				return fail(lerr)
			}
			cfgPath, _ := common.ResolveConfig(false)
			res, berr := devdata.Brief(devdata.BriefRequest{
				LinkRequest: linkReq(cfgPath, t, p), TopPoints: top,
			})
			if berr != nil {
				return fail(berr)
			}
			if common.IsJSON() {
				return emitJSON(t.Env, res, 0, 0)
			}
			printBrief(res)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	cmd.Flags().IntVar(&top, "top", 15, "最多列几个客制点（按行数降序）")
	return cmd
}

// ---------- find ----------

func newFindCmd() *cobra.Command {
	var dir, intent, field, page string
	var limit int
	cmd := &cobra.Command{
		Use:   "find",
		Short: "按意图 + 字段定位：该改哪个加点、参考哪个切片（离线）",
		Long: `把"我要做什么"翻成"改哪里"：

  --intent check      栏位检核     → input.a（AFTER FIELD）+ 模板 a17/a05
  --intent lookup     开窗/带值    → construct.c / browser_fill + a07/a08
  --intent default    预带默认值   → input.b（BEFORE FIELD）/ b_fill / init
  --intent save       保存前       → insert_b / update_b / key_*_b
  --intent display    显示/刷新    → show / detail_show / ui_dialog
  --intent init       初始化       → init / global / construct
  --intent query      查询取数     → query / sql 段 / filter
  --intent sign       签核复核     → signature / signing / state_change

给 --field 收窄到某个字段（可只给后缀，如 xmdduc001）。两个都不给就列意图清单。

示例：
  tt dev find --intent check --field xmdduc001
  tt dev find --field 刻号
  tt dev find --intent 保存前 --human`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			t, p, lerr := taskws.Load(root)
			if lerr != nil {
				return fail(lerr)
			}
			cfgPath, _ := common.ResolveConfig(false)
			res, ferr := devdata.Find(devdata.FindRequest{
				LinkRequest: linkReq(cfgPath, t, p), Intent: intent, Field: field, Page: page, Limit: limit,
			})
			if ferr != nil {
				return fail(ferr)
			}
			// 点位数/行数来自 materials：apply 之后包里的改动不会改变这些统计（实测被当成"没改"）。
			if st, serr := taskws.GetStatus(taskws.StatusRequest{Dir: root}); serr == nil && st.Pkg.Modified {
				res.Note = strings.TrimSpace(res.Note + "\n\n（说明：点位数/行数来自 materials，包内已 apply 的改动不计入统计；看包内正文用 tt dev at <点名>）")
			}
			if common.IsJSON() {
				return emitJSON(t.Env, res, res.Total, len(res.Points))
			}
			printFind(res)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	cmd.Flags().StringVar(&intent, "intent", "", "意图 key 或关键词（check/lookup/default/save/display/init/query/sign 或中文）")
	cmd.Flags().StringVar(&field, "field", "", "字段名（可只给后缀）或画面中文名（如 刻号）")
	cmd.Flags().StringVar(&page, "page", "", "只要这个页签的点：page_id（73）或页签序号（12）")
	cmd.Flags().IntVar(&limit, "limit", 20, "最多列几个点")
	return cmd
}

// ---------- at ----------

func newAtCmd() *cobra.Command {
	var dir, from string
	var contextLines, maxLines int
	cmd := &cobra.Command{
		Use:   "at <行号|点名>",
		Short: "定位：这一行属于哪个点（或这个点是什么内容）",
		Long: `给**行号**（如 50710 或 axmt500_wf:50710）：回答这在哪个加点、是不是客制、
属于哪个 section、由哪个切片生成 —— 编译报错、断点、review 时看到的一行都能直接用。

给**点名/字段名**（如 input.a.page73.xmdduc001 或 xmdduc001）：直接打这段内容（带行号）。

行号是 **.4gl 产物的行号**（编译报错那种）。要注意视图 ws/prog.full.4gl 行号不同，
在视图里看到的行请用 tt dev search 或点名定位。

示例：
  tt dev at 50710
  tt dev at xmdduc003
  tt dev at input.c.page73.xmdduc003 --context 5`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			t, p, lerr := taskws.Load(root)
			if lerr != nil {
				return fail(lerr)
			}
			cfgPath, _ := common.ResolveConfig(false)
			lr := linkReq(cfgPath, t, p)
			if line, ok := parseLine(args[0]); ok {
				res, aerr := devdata.At(devdata.AtRequest{LinkRequest: lr, Line: line, Context: contextLines})
				if aerr != nil {
					return fail(aerr)
				}
				if common.IsJSON() {
					return emitJSON(t.Env, res, 0, 0)
				}
				printAt(res)
				return nil
			}
			query := strings.TrimSpace(args[0])
			if query == "" {
				return output.Errorf("USAGE", exitUsage, "要写成 <行号> 或 <点名>")
			}
			if from != "4gl" {
				return atFromPkg(t, root, lr, query, from)
			}
			res, serr := devdata.Show(devdata.ShowRequest{LinkRequest: lr, Query: query, MaxLines: maxLines})
			if serr != nil {
				return fail(serr)
			}
			if common.IsJSON() {
				return emitJSON(t.Env, res, len(res.Points), len(res.Points))
			}
			printShow(res)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	cmd.Flags().StringVar(&from, "from", "", "点名模式看哪里：pkg（缺省，包内 .tap，能验证 apply 结果）/ 4gl（服务器 build 产物）")
	cmd.Flags().IntVar(&contextLines, "context", 2, "行号模式：附带几行上下文")
	cmd.Flags().IntVar(&maxLines, "max", 200, "点名模式：每个点最多打多少行")
	return cmd
}

// parseLine 判断参数是不是行号（`50710` / `axmt500_wf:50710`）。
func parseLine(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	if i := strings.LastIndex(s, ":"); i >= 0 {
		s = s[i+1:]
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}

// ---------- search ----------

func newSearchCmd() *cobra.Command {
	var dir, scope string
	var limit int
	var caseSensitive bool
	cmd := &cobra.Command{
		Use:   "search <子串>",
		Short: "检索产物与 .tap（apply 后查包内），每条命中标出归属与有无槽位",
		Long: `比 grep 多给三样东西：命中落在哪个加点、是客制还是模板（重产会不会保留）、
属于哪个 section / 切片。**离线**。

范围（--scope，缺省 all）：
  4gl   服务器 build 的产物（fglcomp 报错那种行号）
  tap   材料里拉下来的 .tap
  pkg   工作区包内的 .tap —— apply 之后**这才是最新真相**（all 在工作区里优先用它）
  all   产物 + 最新 .tap

--limit 控制打印条数（默认 50），命中总数照给；--case 区分大小写。

示例：
  tt dev search "SET("
  tt dev search xmdducsite --limit 200 --human`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			t, p, lerr := taskws.Load(root)
			if lerr != nil {
				return fail(lerr)
			}
			cfgPath, _ := common.ResolveConfig(false)
			pkgPath, _ := taskws.PackagePath(root)
			res, gerr := devdata.Grep(devdata.GrepRequest{
				LinkRequest: linkReq(cfgPath, t, p), Pattern: args[0], Limit: limit, Case: caseSensitive,
				Scope: scope, PkgPath: pkgPath,
			})
			if gerr != nil {
				return fail(gerr)
			}
			if common.IsJSON() {
				return emitJSON(t.Env, res, res.Total, len(res.Hits))
			}
			printGrep(res)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	cmd.Flags().StringVar(&scope, "scope", "all", "检索范围：all（产物 + 最新 .tap，缺省）/ 4gl / tap（材料）/ pkg（包内）")
	cmd.Flags().IntVar(&limit, "limit", 50, "每个范围最多打印多少条（总数照给）")
	cmd.Flags().BoolVar(&caseSensitive, "case", false, "区分大小写")
	return cmd
}

// ---------- tzc 直通 ----------

func newApplyCmd() *cobra.Command {
	var dir string
	var dryRun, yes bool
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "把视图改动写进工作区的包（= tt dev tzc apply <ws>）",
		Long: `跑完整验证管线（gate1 + gate2 + gate3）后原子写包。唯一会改 .tzc 的命令。

包路径沿用导出时记的那一个（pkg/<程序>(c).tzc），写完它就是要交给设计器的产物。
--dry-run 只输出逐条目 diff；--yes 对"写区段（解开框架）"二次确认。

示例：
  tt dev apply --dry-run
  tt dev apply`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			if _, _, lerr := taskws.Load(root); lerr != nil {
				return fail(lerr)
			}
			// 把"写回哪个包"绑回**当前工作区**：清单里记的是绝对路径，复制/移动过的工作区
			// 会静默写进原目录的包（实测事故）。不一致就拒绝，见 RebindPackage。
			if _, rebased, rerr := taskws.RebindPackage(root); rerr != nil {
				return fail(rerr)
			} else if rebased {
				// 固定走 stderr：JSON 消费方不看它，但人**必须**知道"这次写的是哪个包"。
				fmt.Fprintf(os.Stderr, "注意：视图清单里的源包路径原本指向别处，已改到当前工作区：%s\n", root)
			}
			args := []string{filepath.Join(root, taskws.DirWs)}
			if dryRun {
				args = append(args, "--dry-run")
			}
			if yes {
				args = append(args, "--yes")
			}
			if common.IsJSON() {
				args = append(args, "--json")
			}
			code := tzc.Run("apply", args)
			if code == 0 && !common.IsJSON() {
				fmt.Fprintln(os.Stderr, "注意：apply 只保证围栏/结构/装载模拟通过，**不编译 BDL**；语法错要到设计器 F7 才暴露。")
				fmt.Fprintln(os.Stderr, "      包内 .4gl 是服务器 build 产物，本次不会改；上传后由设计器 F7 按新 .tap 重生。")
			}
			return exitCode(code)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只输出逐条目 diff，不落盘")
	cmd.Flags().BoolVar(&yes, "yes", false, "对写区段（解开框架）二次确认")
	return cmd
}

func newVerifyCmd() *cobra.Command {
	var dir string
	var strict bool
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "只审计不写包（= tt dev tzc verify <ws>）",
		Long: `跑 gate1 + gate2 并列出全部发现，不写盘。--strict 把 warn 也算失败（CI 用）。

示例：
  tt dev verify
  tt dev verify --strict --human`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			if _, _, lerr := taskws.Load(root); lerr != nil {
				return fail(lerr)
			}
			// 包被换过时 verify 给"通过"是假绿灯（实测）：先拦，并指向 relink。
			if st, serr := taskws.GetStatus(taskws.StatusRequest{Dir: root}); serr == nil &&
				st.Pkg.Present && st.Ws.PkgSha != "" && !st.Pkg.MatchesBase {
				return output.Errorf("REFUSED", exitRefused,
					"包与视图基线不一致（%s ≠ %s）：verify 已拒绝。改动落到哪份包上不可判定；先 tt dev relink 或把基线包放回来",
					taskws.Short(st.Pkg.Sha256), taskws.Short(st.Ws.PkgSha))
			}
			args := []string{filepath.Join(root, taskws.DirWs)}
			if strict {
				args = append(args, "--strict")
			}
			if common.IsJSON() {
				args = append(args, "--json")
			}
			code := tzc.Run("verify", args)
			if code == 0 && !common.IsJSON() {
				fmt.Fprintln(os.Stderr, "注意：verify 只查围栏结构与不变量，**不编译 BDL**；语法错要到设计器 F7 才暴露。")
			}
			return exitCode(code)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	cmd.Flags().BoolVar(&strict, "strict", false, "把 warn 级发现也算作失败")
	return cmd
}

func newRelinkCmd() *cobra.Command {
	var dir string
	var force bool
	cmd := &cobra.Command{
		Use:   "relink",
		Short: "以当前工作区的包为基线，重新导出视图（解「包被换过」的死局）",
		Long: `apply 会校验"源包自 export 之后没被改过"。包被替换（拷错、手工覆盖、别人给了一份）
之后这个校验永远失败，而它建议的"放回基线包"通常你手里没有 —— 工作区就卡死了。

relink 承认"当前这个包就是新基线"，按它重导视图。**视图里未落盘的改动会丢**，
所以默认拒绝脏工作区（要 --force）。之后 apply 恢复正常。

示例：
  tt dev relink
  tt dev relink --force`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			res, rerr := taskws.Relink(taskws.RelinkRequest{Dir: root, Force: force})
			if rerr != nil {
				return fail(rerr)
			}
			if common.IsJSON() {
				return emitJSON("", res, 0, 0)
			}
			fmt.Printf("已以当前包重建视图：%s\n  包 %s（%s）\n  视图 %d 个区域（可编辑 %d）\n",
				res.Dir, filepathBase(res.Pkg), taskws.Short(res.PkgSha), res.Regions, res.Editable)
			for _, w := range res.Warnings {
				fmt.Printf("  ! %s\n", w)
			}
			fmt.Println("\n下一步：确认视图内容，再 tt dev apply")
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	cmd.Flags().BoolVar(&force, "force", false, "视图有未落盘改动也重导（改动会丢）")
	return cmd
}

func newUnlockCmd() *cobra.Command {
	var dir string
	var yes bool
	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "框架解锁：允许改 section 正文（= tt dev tzc unlock <ws>，单向、有代价）",
		Long: `解锁后可以改 section 正文（框架区段），但**解开不可逆**：之后规格调整不再自动
生成程序代码。需要显式 --yes —— AI 不得自主加。

示例：
  tt dev unlock --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, err := resolveRoot(dir)
			if err != nil {
				return fail(err)
			}
			if _, _, lerr := taskws.Load(root); lerr != nil {
				return fail(lerr)
			}
			args := []string{filepath.Join(root, taskws.DirWs)}
			if yes {
				args = append(args, "--yes")
			}
			if common.IsJSON() {
				args = append(args, "--json")
			}
			return exitCode(tzc.Run("unlock", args))
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "工作区目录（缺省：从当前目录往上找 task.json）")
	cmd.Flags().BoolVar(&yes, "yes", false, "确认解锁（单向、有代价）")
	return cmd
}

// ---------- 公共 ----------

// linkReq 把工作区映射成 devdata 的查询请求（三个落点都显式给，查询就不碰配置）。
func linkReq(cfgPath string, t *taskws.Task, p taskws.Paths) devdata.LinkRequest {
	return devdata.LinkRequest{
		CfgPath: cfgPath, Env: t.Env, Prog: t.Prog,
		Dir: p.Materials(), Out: p.Index(), Mdl: p.Mdl(),
	}
}

// resolveRoot 定位工作区根：显式 --dir，否则从当前目录往上找 task.json。
func resolveRoot(flagDir string) (string, error) {
	if d := strings.TrimSpace(flagDir); d != "" {
		return filepath.Abs(d)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("取当前目录失败（用 --dir 指定工作区）: %w", err)
	}
	dir := wd
	for {
		if _, serr := os.Stat(filepath.Join(dir, taskws.TaskFileName)); serr == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// 没找到就把当前目录交回去，让 Load 给出"这里不是工作区 + 怎么建"的提示。
	return wd, nil
}

// emitJSON 统一信封（与 tt gen 同形：ok + 环境信息 + data）。
func emitJSON(env string, data any, total, returned int) error {
	m := common.CurrentMeta()
	m.Env = env
	m.Source = "workspace"
	m.TotalRows = total
	m.Returned = returned
	return output.WriteJSON(os.Stdout, output.Options{Meta: m, Data: data})
}

// entOf 取本命令的 --ent（由 --at <环境>@<企业> 或 --ent 写入）。
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

// fail 把领域错误映射成退出码与信封。
func fail(err error) error {
	switch {
	case errors.Is(err, taskws.ErrNoWorkspace), errors.Is(err, taskws.ErrWorkspaceExists),
		errors.Is(err, taskws.ErrTaskName), errors.Is(err, taskws.ErrWorkspaceVersion),
		errors.Is(err, devdata.ErrInvalidProg), errors.Is(err, devdata.ErrInvalidEnv),
		errors.Is(err, devdata.ErrEnvNotFound), errors.Is(err, devdata.ErrInvalidOpt),
		errors.Is(err, devdata.ErrModuleUnknown):
		return output.Errorf("USAGE", exitUsage, "%v", err)
	case errors.Is(err, taskws.ErrDirtyWorkspace), errors.Is(err, taskws.ErrPackageMismatch):
		return output.Errorf("REFUSED", exitRefused, "%v", err)
	case errors.Is(err, devdata.ErrProgNotFound), errors.Is(err, devdata.ErrMaterialIncomplete),
		errors.Is(err, devdata.ErrPointNotFound):
		return output.Errorf("NOT_FOUND", exitNotFound, "%v", err)
	case errors.Is(err, context.Canceled):
		return output.Errorf("CANCELLED", exitFailed, "已中断；未完成的文件没有落盘")
	default:
		return output.Errorf("FAILED", exitFailed, "%v", err)
	}
}

// exitCode 把本组算出的退出码交回 cobra（根命令的 AlreadyReported 通路）。
func exitCode(code int) error {
	if code == 0 {
		return nil
	}
	return exitCodeOf(code)
}

// exitCodeOf 是带退出码的错误：根命令的 exitCodeOf 解得出，且 AlreadyReported
// 告诉它"文案已经打过了，别再抄一份"（否则 stdout 上会出现两个 JSON 对象）。
type exitCodeOf int

func (e exitCodeOf) Error() string { return fmt.Sprintf("退出码 %d", int(e)) }

// ExitCode 实现根命令期望的退出码接口。
func (e exitCodeOf) ExitCode() int { return int(e) }

// AlreadyReported 告诉根命令"这次失败我已经打出去了"。
func (e exitCodeOf) AlreadyReported() bool { return true }

// interruptContext 让长跑命令能被 Ctrl+C 打断（领域层会清掉临时目录）。
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

// stderrProgress 把进度打到 stderr（stdout 留给结果）。
func stderrProgress() devdata.ProgressFunc {
	return func(p devdata.Progress) {
		if p.Message == "" {
			return
		}
		msg := p.Message
		if p.Total > 0 {
			msg += fmt.Sprintf("  [%d/%d]", p.Done, p.Total)
		}
		fmt.Fprintln(os.Stderr, msg)
	}
}

// atFromPkg 从**包内 .tap** 取点名正文（apply 之后复核"改动进包没有"）。
//
// 为什么默认走这里：工作区里 `.tap` 才是设计源，`.4gl` 是上次服务器 build 的产物。
// 刚 apply 完去看 `.4gl` 只会看到旧内容（实测试用者因此以为改动没生效）。
func atFromPkg(t *taskws.Task, root string, lr devdata.LinkRequest, query, from string) error {
	if from != "" && from != "pkg" {
		return output.Errorf("USAGE", exitUsage, "--from 只能是 pkg 或 4gl，给的是 %q", from)
	}
	// 孤儿点优先：它只存在于 .tap，包里当然没有 —— 但"没有"不是客户要的答案。
	if res, serr := devdata.Show(devdata.ShowRequest{LinkRequest: lr, Query: query}); serr == nil && res.Match == "orphan" {
		if common.IsJSON() {
			return emitJSON(t.Env, res, 1, 1)
		}
		printShow(res)
		return nil
	}
	doc, err := devdata.LoadIndex(lr)
	if err != nil {
		return fail(err)
	}
	pts, _ := devdata.MatchPoints(doc, query)
	if len(pts) == 0 {
		if cols, by, _ := devdata.ResolveColumns(doc, query); by == "label" {
			for _, p := range doc.Points {
				if containsString(cols, devdata.ColumnOfPoint(p.Name)) {
					pts = append(pts, p)
				}
			}
		}
	}
	if len(pts) == 0 {
		return output.Errorf("NOT_FOUND", exitNotFound,
			"包里没有匹配 %q 的点位（试 --from 4gl 看服务器产物，或 tt dev brief 看整体）", query)
	}
	const maxShow = 5
	var payload []map[string]any
	shown, noDecl := 0, 0
	for _, p := range pts {
		v, pkgPath, perr := taskws.PkgPoint(root, p.Name)
		if perr != nil {
			return fail(perr)
		}
		if v == nil {
			// 模板点只有"洞"、没有 <point> 声明，包内天然没有正文。别把它当错误一行行打印。
			noDecl++
			continue
		}
		if shown >= maxShow {
			continue
		}
		shown++
		if common.IsJSON() {
			payload = append(payload, map[string]any{
				"name": v.Name, "src": v.Src, "ver": v.Ver, "hasSlot": v.HasSlot,
				"line0": v.Line0, "line1": v.Line1, "content": v.Content, "pkg": pkgPath,
			})
			continue
		}
		printPkgPoint(v, pkgPath)
	}
	if shown == 0 {
		// 包里没有声明（全是模板洞）：退回产物内容，并说明为什么。
		res, serr := devdata.Show(devdata.ShowRequest{LinkRequest: lr, Query: query})
		if serr != nil {
			return fail(serr)
		}
		if common.IsJSON() {
			return emitJSON(t.Env, res, len(res.Points), len(res.Points))
		}
		fmt.Printf("这些点在包里没有 .tap 声明（模板洞），下面是服务器产物里的内容：\n\n")
		printShow(res)
		return nil
	}
	if !common.IsJSON() && noDecl > 0 {
		fmt.Printf("\n（另有 %d 个匹配点没有 .tap 声明：模板洞里没有正文，看产物请加 --from 4gl）\n", noDecl)
	}
	if common.IsJSON() {
		return emitJSON(t.Env, map[string]any{"source": "pkg", "points": payload, "noDecl": noDecl}, len(pts), shown)
	}
	return nil
}

// containsString 判列表里是否有某个值（忽略大小写）。
func containsString(list []string, v string) bool {
	for _, s := range list {
		if strings.EqualFold(s, v) {
			return true
		}
	}
	return false
}
