package form

// `tt form fetch`：把一个模块下的表单**批量预热进缓存**（契约 C2 / A13）。
//
// 它做的事只有一件：用既有的 `List` 拿到该模块的表单名，逐个走既有的取件通道取回来，
// 再走既有的 `Cache.Put` 写进 `<数据目录>/forms/`（**只缓原始 XML**，C2.4）。
//
// 三条硬约束：
//
//  1. **只读（C2.2 / H6）**：只 `list` + `get`。本命令没有任何写服务器/删服务器的动作 ——
//     `Source` 接口本身就长不出那些动作（见 source.go 的 `Source` 注释）。
//  2. **只缓存原始 XML（C2.4）**：`Put` 写的就是取回来的那串字节。这里**刻意不解析**：
//     解析是 `show` / `verify` 的事（D5：解析是毫秒级，缓存里出现"解析结果"就多一个真相）。
//  3. **失败不许静默吞**：哪张取不到、为什么取不到，逐条进输出（人读与 `--json` 都有），
//     退出码按 A13 取 `exitFormUnknown`（1）—— **不是 6**（6 是 `verify` 的"体检不过"）。
//
// 与 `show` 的分工：`show` 是"按需单取一张"（一次取一张，顺手缓存）；`fetch` 是"先去把
// 这一模块的表单都取回来"，这样后面一串 `show`/`verify` 不必每次等网络。

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/output"
)

var (
	fetchModule string
	fetchDir    string
	fetchForce  bool
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "批量把该模块的表单取进缓存（只缓原始 XML，绝不写服务器）",
	Long: `列出该模块下的表单，逐个取件并写进表单缓存（<数据目录>/forms/，只缓原始 XML）。

  tt form fetch --env <环境> --module <模块> [--force] [--json]

  --force   忽略已有缓存，全部重取（服务器上重编译过表单时用）
  --dir     从本地目录/镜像取件（离线；也可用 TT_FORM_DIR）

输出：总数 / 新取 / 命中缓存 / 失败（失败逐条列出表单名与原因，不静默吞）。
取件通道与 tt form show 相同，**只读**：只 list + get，绝不写、不删、不起作业。

退出码（契约 A13）：0 全部取到（含全部命中缓存）/ 1 有表单取不到或清单列不出
（FORM_UNKNOWN）/ 2 用法错（缺 --env/--module）。`,
	Args: noArgs,
	RunE: runFetch,
}

func init() {
	fetchCmd.Flags().StringVar(&fetchModule, "module", "", "模块名（如 axm）")
	fetchCmd.Flags().StringVar(&fetchDir, "dir", "", "从本地目录/镜像取件（离线；也可用 TT_FORM_DIR）")
	fetchCmd.Flags().BoolVar(&fetchForce, "force", false, "忽略缓存，全部重取")
	Group.AddCommand(fetchCmd)
}

// formCache 定位本次运行的表单缓存落点（`<数据目录>/forms/`，走 config.CacheDir 校验）。
//
// `fetch` 与 `cache` 共用这一处：两边各算一遍落点，迟早会出现"fetch 写这、cache 读那"，
// 而那种错的表现是"缓存永远是空的"（C2.5：漏登记的表现正是静默失效）。
func formCache() (*Cache, error) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return nil, err
	}
	return NewCacheFromDataDir(config.LocationsAt(cfgPath).DataDir())
}

// fetchFailure 是一条失败：**表单名 + 原因**（原因不许省略成"失败了"）。
type fetchFailure struct {
	Form   string `json:"form"`
	Reason string `json:"reason"`
}

// fetchReport 是 `--json` 的信封。
//
// 键是**新增**的（契约 C5.2 只冻结了 `show` 的信封；那条冻结一个键都不动）。
// `ok` 与 `failed` 同时给：agent 既能 `jq .ok`，也能直接数失败条数。
type fetchReport struct {
	Env      string         `json:"env"`
	Module   string         `json:"module"`
	Total    int            `json:"total"`
	Fetched  int            `json:"fetched"`
	Cached   int            `json:"cached"`
	Failed   int            `json:"failed"`
	Failures []fetchFailure `json:"failures"`
	Forced   bool           `json:"forced"`
	OK       bool           `json:"ok"`
	CacheDir string         `json:"cacheDir"`
	Source   string         `json:"source"`
}

func runFetch(cmd *cobra.Command, _ []string) error {
	env, err := requireEnv()
	if err != nil {
		return err
	}
	module := strings.TrimSpace(fetchModule)
	if module == "" {
		return failUsage("fetch 需要 --module <模块>（还要 --env <环境>）")
	}
	dir := firstNonEmpty(fetchDir, os.Getenv("TT_FORM_DIR"))

	// 取件通道**不带缓存包装**：缓存由本命令显式读写 —— 只有自己 Lookup/Put，才分得清
	// "这次是新取的"与"这次是命中的"（CachingSource.Form 把两者合并成一个返回值）。
	src, err := sourceFor(env, sourceOpts{Dir: dir, NoCache: true})
	if err != nil {
		return formUnknown("检查 --env 对应的环境配置，或用 --dir 指向本地镜像/目录", "%v", err)
	}
	cache, err := formCache()
	if err != nil {
		return formUnknown("用 --config 指向配置文件；并确认 config.CacheSubdirs 里有 \""+
			CacheSubdirForm+"\"（清单外名字拿空串，见 C2.3）", "算不出表单缓存落点：%v", err)
	}

	ctx := cmd.Context()
	names, err := src.List(ctx, env, module)
	if err != nil {
		return formUnknown("检查模块名与取件通道（--dir / 环境配置）；"+
			"确认通道本身没问题可以试：tt form ls --env "+env+" --module "+module,
			"列不出 %s 模块下的表单（环境 %s）：%v", module, env, err)
	}
	if len(names) == 0 {
		// 空清单会被读成"这个模块没有表单"，而真相多半是路径/模块名不对（与 ls 同一条口径）。
		return formUnknown("检查 --module 与取件通道；先跑 tt form ls --env "+env+" --module "+module+" 看清单",
			"%s 模块下一张表单都没有（环境 %s）；拒绝把「列不到」当成「取完了」", module, env)
	}
	sort.Strings(names)

	rep := fetchReport{
		Env:      env,
		Module:   module,
		Total:    len(names),
		Failures: []fetchFailure{},
		Forced:   fetchForce,
		CacheDir: cache.Dir,
		Source:   sourceLabel(dir),
	}
	for _, name := range names {
		if !fetchForce {
			// 命中判据与 show 完全相同：Cache.Lookup（源报得出身份时按 size+mtime，否则按 TTL）
			if _, ok := cache.Lookup(ctx, src, env, module, name); ok {
				rep.Cached++
				continue
			}
		}
		b, ferr := src.Form(ctx, env, module, name)
		if ferr != nil {
			rep.Failures = append(rep.Failures, fetchFailure{Form: name, Reason: ferr.Error()})
			continue
		}
		if len(b) == 0 {
			// 取件"成功"但内容为空：那不是一张表单，而是取件通道指错了路（空文件 / 空镜像）。
			// 归到失败里并写清病因 —— 别让 Cache.Put 的白名单替我们报"不缓存空内容"，
			// 那句话听起来像缓存的问题，而病人不在缓存。
			rep.Failures = append(rep.Failures, fetchFailure{
				Form:   name,
				Reason: "取回的内容是空的（0 字节）：空文件不是合法 .42f，查一下取件通道是不是指错了目录",
			})
			continue
		}
		var mtime time.Time
		if id, ok := src.(Identifiable); ok {
			if _, mt, ierr := id.Identify(ctx, env, module, name); ierr == nil {
				mtime = mt
			}
		}
		// 唯一写入点：Cache.Put（cache.go）。写失败同样进失败清单 ——
		// "取到了但没缓存上"与"没取到"对下一次调用是同一件事（都得重取）。
		if perr := cache.Put(env, module, name, b, mtime); perr != nil {
			rep.Failures = append(rep.Failures, fetchFailure{Form: name, Reason: "写入缓存失败：" + perr.Error()})
			continue
		}
		rep.Fetched++
	}
	rep.Failed = len(rep.Failures)
	rep.OK = rep.Failed == 0

	if wantJSON(cmd) {
		if perr := common.PrintJSON(rep); perr != nil {
			return perr
		}
	} else {
		printFetchTable(rep)
	}
	if rep.OK {
		return nil
	}
	return failedFetch(rep)
}

// failedFetch 造"已把失败逐条打出去"的错误：退出码 1（A13 的取件失败），**不是 6**。
//
// 为什么实现 alreadyReported：失败清单**就是**这次的输出（人读与 JSON 各一份），
// 若再让根命令 reportError 打一次，`--json` 下 stdout 上会出现两个对象 ——
// 那是 `tt form verify`（mismatchError）与 `tt dev` 都立过的同一条规矩。
// 错误本身仍然带稳定码 `FORM_UNKNOWN`（errors.As 经 Unwrap 取到），
// 所以"失败原因在输出里、失败分类在码上"两件事都在。
type fetchError struct{ oe *output.Error }

func (e *fetchError) Error() string         { return e.oe.Message }
func (e *fetchError) ExitCode() int         { return e.oe.Exit }
func (e *fetchError) AlreadyReported() bool { return true }
func (e *fetchError) Unwrap() error         { return e.oe }

// failedFetch 把失败清单压成一句可核对的话（含每张表单名，不含噪音）。
func failedFetch(rep fetchReport) error {
	names := make([]string, 0, len(rep.Failures))
	for _, f := range rep.Failures {
		names = append(names, f.Form)
	}
	oe := &output.Error{
		Code: CodeFormUnknown,
		Exit: exitFormUnknown,
		Message: fmt.Sprintf("%d/%d 张表单取件失败（环境 %s / 模块 %s）：%s（原因见本次输出）",
			rep.Failed, rep.Total, rep.Env, rep.Module, strings.Join(names, "、")),
	}
	hint := fmt.Sprintf("先跑 tt form fetch --env %s --module %s 看逐条原因；"+
		"取件通道本身有问题时可以试 --dir <本地镜像/目录> 走离线", rep.Env, rep.Module)
	oe.WithHint(hint)
	return &fetchError{oe: oe}
}

// printFetchTable 给人读：一行总数，一行计数，失败逐条列出（带原因），最后一行提示。
func printFetchTable(r fetchReport) {
	force := ""
	if r.Forced {
		force = "（--force：忽略缓存全部重取）"
	}
	writeOut("%s / %s：共 %d 张%s\n", r.Env, r.Module, r.Total, force)
	writeOut("  新取 %d · 命中缓存 %d · 失败 %d\n", r.Fetched, r.Cached, r.Failed)
	writeOut("  缓存落点 %s（取件通道 = %s）\n", r.CacheDir, r.Source)
	for _, f := range r.Failures {
		writeOut("  失败 %s：%s\n", f.Form, f.Reason)
	}
	if r.Failed > 0 {
		writeOut("提示：失败的表单没有进缓存；查明原因后重跑 tt form fetch --env %s --module %s\n",
			r.Env, r.Module)
	}
}

// 编译期自证：fetch 的失败错误必须同时满足"带退出码"与"已报告"两件事。
var (
	_ interface{ ExitCode() int }         = (*fetchError)(nil)
	_ interface{ AlreadyReported() bool } = (*fetchError)(nil)
	_ error                               = (*fetchError)(nil)
)
