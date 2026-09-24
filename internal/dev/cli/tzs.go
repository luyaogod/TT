// `tt dev tzs` —— 表单包（`.tzs` / `.tzv`）的入口。
//
// 与 `tt dev tzc` 的区别：
//   - tzc 是**代码包**管线：渲染带围栏的 prog.full.4gl、跑三道闸门、apply 原子写回；
//   - tzs 的 `export` 只有一件事：**纯解压**，不解围栏、不校验、不产生工作区。
//     要**读写表单**走 `tt dev tzs <动词>`（open / set_spec_attr / save / …），
//     那一条由 engine/ 里那个 C# 引擎驱动，动词表来自引擎自己的函数表。
//
// 红线曾经是「永远不写回」，理由是：
//
//	表单（SPEC）由设计器的表单设计器维护，tdev 没有对应的模型与验收样本；硬写就是拿真实包赌。
//
// 那个前提现在不成立了 —— engine/ 就是**设计器自己的代码**（它 `Assembly.LoadFrom` 设计器的
// 程序集，布局属性走设计器自己的 `XmlElement` 索引器，验收用设计器自己的校验器加 RoundTrip
// 不动点）。所以红线改成「**导出只读、要写走引擎**」：`export` 的产物仍然是只读参考，
// 不要手工改完再塞回包；改表单走具名动词，让设计器自己算。
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"tt/internal/dev/model"
	"tt/internal/dev/pkgfile"
	"tt/internal/dev/tzs"
)

const tzsUsage = `tt dev tzs —— 表单包工具（导出只读；读写表单由一个常驻引擎跑）

用法：
  tt dev tzs export <pkg.tzs> [-o <dir>] [--force] [--json]
        # 纯解压：把 zip 原样摊到目录里，不做任何处理
        # -o 省略时默认解压到 <包所在目录>/<程序名>-unzip（身份后缀 (c)/(s) 会去掉）
        # 目标已存在且非空时拒绝，加 --force 覆盖同名文件

  tt dev tzs <动词> [--form <程序名>] --args '<JSON 对象>' [--args-file <UTF-8 文件>] [--workspace <dir>] [--rpc-timeout <秒>] [--json]
        # 读写表单：**唯一**写路径，由设计器自己的引擎算，不是我们拼 XML
        # 52 个动词全部由引擎的函数表生成，所以这里没有"函数名"这一层要填：
        #   open form_tree find_component get_component list_spec_nodes describe_kind
        #   set_spec_attr set_spec_attrs set_layout_attr set_layout_attrs add_widget add_field
        #   field_add nudge validate save close …
        #   动词全名：tt dev tzs --help（引擎可达时附在帮助后面）
        #   某个动词的参数与示例：tt dev tzs <动词> --help
        # 参数**只用 JSON 给** —— 数组就是数组、布尔就是布尔，不必学命令行的引号与切分：
        #   tt dev tzs open           --args '{"path":"D:\\pkg\\aapp320(c).tzs"}' --json
        #   tt dev tzs list_open --json                      # 无参数的动词可以省掉 --args
        #   tt dev tzs form_tree      --form aapp320 --args '{"depth":2}' --json
        #   tt dev tzs set_spec_attr  --form aapp320 --args '{"path":"<path>","kind":"field","attr":"can_edit","value":"true"}'
        #   tt dev tzs validate       --form aapp320 --json
        #   tt dev tzs save           --form aapp320 --args '{"out":"D:\\pkg\\_ai.tzs"}' --json   # 写**新**包
        # --form 是"哪一张已打开的表单"：写程序名（aapp320）或 ProgramKey（aapp320|Form），
        #   由引擎解析 —— 所以**不必搬运句柄**（它每次 open 都换号、永不复用）。
        # 中文/长内容请用 --args-file（UTF-8 文件；值是单个 - 表示读标准输入）
        # 传输开关（不是动词参数）：--form / --workspace / --rpc-timeout / --json / -h

  tt dev tzs doctor [--json]        # 环境自检（引擎产物、设计器目录、工作区、管道名）
  tt dev tzs stop                   # 停掉本工作区的常驻引擎（不启动）
  tt dev tzs reap [--yes]           # 清理引擎重编后停不掉的孤儿守护进程

  支持的输入：.tzs（表单包）、.tzv（简易表单包）。
  .tzc/.tzf/.tzx 等**代码包**请用 ` + "`tt dev tzc export`" + `（那条管线有围栏渲染与三道闸门）。

保留名（不是动词参数）：--form / --args / --args-file / --workspace / --rpc-timeout / --json / -h。
动词名可以用连字符写（form-tree = form_tree）；**参数一律写进 --args 的 JSON 里**，
--handle h9 这种写法会当场报错（handle 不是开关，它是 JSON 里的一个键）。

红线：**export 的产物是只读参考** —— 它就是一包文件，没有 manifest/围栏，不要手工改完再
塞回包。要改表单走 ` + "`tt dev tzs <动词>`" + `：那是设计器自己的模型在算，改完设计器打得开。
`

func cmdTzs(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, tzsUsage)
		return 2
	}
	switch args[0] {
	case "export":
		return cmdTzsExport(args[1:])
	case "doctor":
		return cmdTzsDoctor(args[1:])
	case "stop":
		return cmdTzsStop(args[1:])
	case "reap":
		return cmdTzsReap(args[1:])
	case "-h", "--help", "help":
		// 静态用法总是打得出（没有引擎、没有工作区也行），动词索引是**尽力而为**：
		// 拉不到就只留一句指引。--help 绝不能因为环境不全而失败。
		fmt.Print(tzsUsage)
		printVerbIndexIfReachable(os.Stdout)
		return 0
	default:
		// 其余第一个词一律当**具名动词**（open / save / set_spec_attr / …）：
		// 名字与参数来自引擎 manifest，这里不做任何硬编码（见 tzs_engine.go 的说明）。
		return runTzsVerb(args)
	}
}

// printVerbIndexIfReachable 在引擎可达时把动词索引（按组列名字）附在帮助后面。
//
// 为什么允许"可有可无"：`--help` 是**还没有环境时唯一能用的一条命令**。要求引擎可达
// 就等于把说明书锁在门里面（而那条门锁正是 `tt dev tzs doctor` 要告诉你怎么开的）。
func printVerbIndexIfReachable(w io.Writer) {
	exe, err := tzsEngineExe()
	if err != nil {
		return
	}
	m, err := tzs.FetchManifest(tzsCtx(), exe)
	if err != nil {
		fmt.Fprintln(w, "\n（这次没能列出动词名：引擎不可达。tt dev tzs doctor 看差什么）")
		return
	}
	fmt.Fprintln(w)
	m.Index(w)
}

// defaultUnzipDir 是 -o 省略时的默认解压目录：<包所在目录>/<程序名>-unzip。
//
// 用 -unzip 后缀（而不是 tzc 的 -ws）是为了让「工作区」和「纯解压产物」一眼可分：
// 前者有 manifest/.tdev/git、能 apply；后者就是一包文件，只读。
func defaultUnzipDir(pkgPath string) string {
	dir := filepath.Dir(pkgPath)
	base := strings.TrimSuffix(filepath.Base(pkgPath), filepath.Ext(pkgPath))
	if m := reIdentitySuffix.FindString(base); m != "" {
		base = strings.TrimSuffix(base, m)
	}
	if base == "" {
		base = "pkg"
	}
	return filepath.Join(dir, base+"-unzip")
}

func cmdTzsExport(args []string) int {
	fs := flag.NewFlagSet("tzs export", flag.ContinueOnError)
	out := fs.String("o", "", "输出目录（可省略，默认 <包所在目录>/<程序名>-unzip）")
	force := fs.Bool("force", false, "目标非空时也解压：覆盖同名文件")
	asJSON := fs.Bool("json", false, "输出 JSON")
	if err := parseArgs(fs, args, "o"); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "用法：tt dev tzs export <pkg.tzs> [-o <dir>] [--force] [--json]")
		return 2
	}
	pkgPath := fs.Arg(0)
	w := os.Stdout

	// 输入类型闸门：只吃表单包。把代码包指回 tzc，避免「用错管线」这类简单错误。
	switch strings.ToLower(filepath.Ext(pkgPath)) {
	case ".tzs", ".tzv":
		// 表单包，纯解压
	case ".tzc", ".tzf", ".tzx":
		return fail(&pkgfile.FormatError{
			Msg: "这是代码包，请用 `tt dev tzc export`（tt dev tzs 只纯解压表单包）",
			Detail: []string{"输入：" + pkgPath,
				"代码包要走围栏渲染 + 三道闸门：`tt dev tzc export <pkg> [-o <dir>]`"},
		}, *asJSON)
	default:
		return fail(&pkgfile.FormatError{
			Msg:    "tt dev tzs export 只支持表单包 .tzs / .tzv",
			Detail: []string{"输入：" + pkgPath},
		}, *asJSON)
	}

	dst := *out
	if dst == "" {
		// -o 缺省：先用 config.json 的 tdev.defaultOut；没配就退回
		// <包所在目录>/<程序名>-unzip（旧行为）。配置缺失时 loadTdevSettings 返回零值，
		// 等价于「没配」，因此没有配置文件的用户行为与合并前逐字一致。
		if s := loadTdevSettings(); s.DefaultOut != "" {
			dst = s.DefaultOut
		} else {
			dst = defaultUnzipDir(pkgPath)
		}
	}
	entries, err := pkgfile.UnzipTo(pkgPath, dst, *force)
	if err != nil {
		return fail(err, *asJSON)
	}
	pkgSha, _ := model.Sha256File(pkgPath)
	abs, _ := filepath.Abs(dst)
	total := 0
	for _, e := range entries {
		total += e.Size
	}
	if *asJSON {
		emitJSON(w, map[string]any{
			"ok": true, "pkg": pkgPath, "pkg_sha256": pkgSha, "dir": abs,
			"entries": entries, "count": len(entries), "bytes": total, "readonly": true,
		})
		return 0
	}
	line(w, "已纯解压（不做任何处理，也不产生工作区）：%s", abs)
	line(w, "  来源：%s（%d B，sha256=%s）", pkgPath, fileSize(pkgPath), short(pkgSha))
	line(w, "  文件：%d 个，共 %d B", len(entries), total)
	for _, e := range entries {
		line(w, "    %-28s %8d B  sha256=%s", e.Name, e.Size, short(e.Sha256))
	}
	line(w, "")
	line(w, "提醒：这是**只读参考**（没有 tzs apply）；要改表单用 `tt dev tzs <动词>`（open/add_field/save…），由设计器自己的引擎算。")
	line(w, "下一步：直接读上面的文件即可；不要把它当成 tzc 工作区去 apply。")
	return 0
}

// fileSize 取文件大小（取不到返回 0）。
func fileSize(p string) int64 {
	st, err := os.Stat(p)
	if err != nil {
		return 0
	}
	return st.Size()
}
