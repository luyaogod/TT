// tzs_engine.go —— `tt dev tzs` 的**读写**动词：把 tt 接到 engine/ 那个 C# 引擎上。
//
// 与 tzs.go 里 `export` 的分工：export 是纯解压、不碰引擎，只读；这里全部要引擎。
//
// 为什么这里可以有写路径，而 tzs.go 的头注释写着「永远不写回」——那条红线是**有前提**的：
//
//	表单（SPEC）由设计器的表单设计器维护，tdev 没有对应的模型与验收样本；硬写就是拿真实包赌。
//
// 现在那个前提不成立了：engine/ 就是**设计器自己的代码**（它 LoadFrom 设计器的程序集，
// 布局属性走设计器自己的 XmlElement 索引器，验收用设计器自己的校验器与不动点）。所以在
// tdev 里第一次有了「对应模型」；红线随之改成「导出仍只读、要写走引擎」，而不是「不许写」。
//
// 三条纪律（与 A1/A2/A3 同源）：
//
//  1. **不自己重算引擎的任何东西。** 函数表、参数类型、管道名全部问引擎（--manifest /
//     --pipe-name）。管道名含引擎程序集的 MVID，重算一百行且算错了不响。
//  2. **请求一旦上线，绝不重试。** 协议无幂等键，而这些函数都在改设计器内存里的模型；
//     守护进程中途死掉时重试是在赌「上一次写进去了没有」。
//  3. **工作区没有缺省。** 引擎内置的默认工作区是一个真实客户目录，落到它上面会去 Boot
//     别人的包。解析顺序末端是**拒绝**，不是回落。
package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tt/internal/config"
	"tt/internal/dev/tzs"
)

// tzsEngineOptions 解析出一次引擎调用需要的四样东西。
//
// 工作区的解析顺序：--workspace flag > TZSCLI_WS 环境变量 > config.json 的 tzs.workspace。
// **末端拒绝**：三层都没有时退出 2 并说清该配哪个键，绝不 spawn —— 引擎的缺省是一个真实
// 客户目录，落上去等于拿别人的表单当草稿纸。
func tzsEngineOptions(wsFlag string) (tzs.Options, error) {
	var cfg config.TzsSettings
	if path, err := config.ResolvePath("", true); err == nil && path != "" {
		if root, err := config.Load(path); err == nil {
			cfg = root.Tzs
		}
	}

	exe := cfg.ServerExe
	if exe == "" {
		self, err := os.Executable()
		if err != nil {
			return tzs.Options{}, fmt.Errorf("定位不到 tt.exe 自身：%w", err)
		}
		// 与 skills/ 同一种分发形态：exe 同目录的子目录，不进二进制。
		exe = filepath.Join(filepath.Dir(self), "tzs", "tzs-server.exe")
	}

	ws := wsFlag
	if ws == "" {
		ws = os.Getenv("TZSCLI_WS")
	}
	if ws == "" {
		ws = cfg.Workspace
	}
	if ws == "" {
		return tzs.Options{}, fmt.Errorf(
			"未配置工作区：--workspace、TZSCLI_WS 与配置里的 tzs.workspace 都是空的。\n" +
				"  引擎的缺省工作区是一个真实客户目录，所以这里拒绝启动而不是回落。\n" +
				"  配置：tt config set tzs.workspace \"D:\\\\你的工作区\"")
	}

	workDir := ""
	if path, err := config.ResolvePath("", true); err == nil && path != "" {
		workDir = filepath.Dir(path) // 状态文件与守护进程日志落在 config.json 旁边
	}

	// 设计器目录与工作区同一条规矩：环境变量优先，配置兜底，都空则交给引擎自己的缺省。
	// 这里读一次环境变量的意义是让 doctor / 报错文案说真话 —— 引擎本来就会继承这个变量，
	// 而 `Options.InstallDir` 非空时会经 spawn 的环境覆盖它（见 tzs.Options.extraEnv）。
	installDir := os.Getenv("TZSCLI_INSTALL")
	if installDir == "" {
		installDir = cfg.InstallDir
	}

	return tzs.Options{
		Exe:        exe,
		InstallDir: installDir,
		Workspace:  ws,
		WorkDir:    workDir,
	}, nil
}

// ---------- manifest / fns：两个只读的"问引擎"动词 ----------

func cmdTzsManifest(args []string) int {
	fs := flag.NewFlagSet("tzs manifest", flag.ContinueOnError)
	if err := parseArgs(fs, args); err != nil {
		return 2
	}
	o, err := tzsEngineOptions("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	m, err := tzs.FetchManifest(tzsCtx(), o.Exe)
	if err != nil {
		return fail(err, false)
	}
	// 逐字节转发引擎的输出：这是 AI 直接消费的东西，Go 侧改写它就是在制造第二份函数表。
	raw, err := json.MarshalIndent(m.Fns, "", "  ")
	if err != nil {
		return fail(err, false)
	}
	os.Stdout.Write(append(raw, '\n'))
	return 0
}

func cmdTzsFns(args []string) int {
	fs := flag.NewFlagSet("tzs fns", flag.ContinueOnError)
	if err := parseArgs(fs, args); err != nil {
		return 2
	}
	o, err := tzsEngineOptions("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	m, err := tzs.FetchManifest(tzsCtx(), o.Exe)
	if err != nil {
		return fail(err, false)
	}
	if fs.NArg() == 0 {
		m.Help(os.Stdout) // 全表，按组渲染
		return 0
	}
	fn := m.ByName(fs.Arg(0))
	if fn == nil {
		fmt.Fprintf(os.Stderr, "没有这个函数：%s\n\n可用的：%s\n", fs.Arg(0), strings.Join(m.Names(), " "))
		return 2
	}
	printFn(os.Stdout, fn)
	return 0
}

// printFn 渲染单个函数的参数表。参数本身用 JSON 打 —— 它的字段（t/req/values/from）
// 就是 manifest 的原样，手写一遍等于抄第二份，抄了就会漂移。
func printFn(w io.Writer, fn *tzs.SpecFn) {
	fmt.Fprintf(w, "%s — %s\n", fn.Name, fn.Desc)
	marks := []string{}
	if fn.Writes {
		marks = append(marks, "写")
	}
	if fn.Slow {
		marks = append(marks, "慢")
	}
	if fn.NeedsHandle {
		marks = append(marks, "需要句柄")
	}
	fmt.Fprintf(w, "  组 %s   返回 %s   %s\n", fn.Group, fn.Returns, strings.Join(marks, " "))
	if len(fn.Args) > 0 {
		fmt.Fprintln(w, "  参数：")
		for _, p := range fn.Args {
			b, _ := json.Marshal(p)
			fmt.Fprintf(w, "    %s\n", b)
		}
	}
	if len(fn.Errors) > 0 {
		fmt.Fprintf(w, "  错误码：%s\n", strings.Join(fn.Errors, " "))
	}
}

// ---------- call：把命令行翻成一次 JSON-RPC ----------

// takeEngineFlags 摘出 call 自己的三个开关，剩下的原样交给 tzs.SplitArgs。
//
// 不能用标准库 flag 解析整个 argv：函数参数的名字是**运行时**从 manifest 来的（`--paths`、
// `--can_edit`…），flag 包遇到不认识的名字会直接报错。
func takeEngineFlags(argv []string) (rest []string, ws string, timeout time.Duration, asJSON bool, err error) {
	take := func(i *int) (string, bool) {
		a := argv[*i]
		if eq := strings.IndexByte(a, '='); eq >= 0 {
			return a[eq+1:], true
		}
		if *i+1 < len(argv) && !strings.HasPrefix(argv[*i+1], "--") {
			*i++
			return argv[*i], true
		}
		return "", false
	}
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		switch {
		case a == "--json":
			asJSON = true
		case a == "--workspace" || strings.HasPrefix(a, "--workspace="):
			v, ok := take(&i)
			if !ok {
				return nil, "", 0, false, fmt.Errorf("--workspace 需要参数")
			}
			ws = v
		case a == "--timeout" || strings.HasPrefix(a, "--timeout="):
			v, ok := take(&i)
			if !ok {
				return nil, "", 0, false, fmt.Errorf("--timeout 需要参数")
			}
			n, e := strconv.Atoi(v)
			if e != nil || n <= 0 {
				return nil, "", 0, false, fmt.Errorf("--timeout 需要一个正整数秒数，收到 %q", v)
			}
			timeout = time.Duration(n) * time.Second
		default:
			rest = append(rest, a)
		}
	}
	return rest, ws, timeout, asJSON, nil
}

func cmdTzsCall(args []string) int {
	rest, wsFlag, timeout, asJSON, err := takeEngineFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "用法：tt dev tzs call <fn> [--<参数> <值>…] [--workspace <dir>] [--timeout <秒>] [--json]")
		fmt.Fprintln(os.Stderr, "  函数与参数见：tt dev tzs fns")
		return 2
	}
	fnName, fnArgs := rest[0], rest[1:]

	o, err := tzsEngineOptions(wsFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}

	m, err := tzs.FetchManifest(tzsCtx(), o.Exe)
	if err != nil {
		return fail(err, asJSON)
	}
	parsed, err := tzs.SplitArgs(fnArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	// 本地校验只做 manifest 明文声明的事（必填 / 类型 / 静态取值 / 未知参数）。多做的每一分
	// 都会变成「本地拒绝了一个引擎本会接受的调用」——from:* 的参数（attr、kind）与自由字符串
	// （add_action 的 type）本来就是运行时的，本地一个字符都不碰。
	raw, err := tzs.BuildArgs(m, fnName, parsed)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	pipe, err := tzs.Ensure(tzsCtx(), o)
	if err != nil {
		return fail(err, asJSON)
	}

	d := tzs.DefaultDialer()
	reply, err := tzs.Call(tzsCtx(), d, pipe, 1, fnName, raw, tzs.NormalizeTimeout(timeout))
	code := tzs.ExitCode(reply, err)

	// open 成功后记住句柄，让随后的 call 可以省掉 --handle（**本地方便，不改协议**：
	// 句柄始终是引擎的东西，进程一死全部失效，所以它只当提示用，不当缓存）。
	if code == 0 && fnName == "open" && reply != nil {
		if h := replyString(reply, "handle"); h != "" {
			if p := replyString(reply, "path"); p != "" {
				_ = tzs.Remember(o, h, p)
			}
		}
	}

	if asJSON {
		printRawReply(reply, err)
		return code
	}
	return printHumanReply(fnName, reply, err, code)
}

// ---------- stop / reap / doctor ----------

func cmdTzsStop(args []string) int {
	fs := flag.NewFlagSet("tzs stop", flag.ContinueOnError)
	ws := fs.String("workspace", "", "工作区（默认取 TZSCLI_WS / 配置）")
	_ = fs.Bool("all", false, "停掉状态文件里记录的全部守护进程")
	_ = fs.Bool("yes", false, "兼容保留")
	if err := parseArgs(fs, args, "workspace"); err != nil {
		return 2
	}
	o, err := tzsEngineOptions(*ws)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	// stop 绝不 spawn：让 stop 去起一个服务器是件可笑的事。
	if err := tzs.Stop(tzsCtx(), o); err != nil {
		return fail(err, false)
	}
	return 0
}

func cmdTzsReap(args []string) int {
	fs := flag.NewFlagSet("tzs reap", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "确认结束这些进程")
	ws := fs.String("workspace", "", "工作区（默认取 TZSCLI_WS / 配置）")
	if err := parseArgs(fs, args, "workspace"); err != nil {
		return 2
	}
	o, err := tzsEngineOptions(*ws)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	pids, err := tzs.Reap(tzsCtx(), o, *yes)
	if err != nil {
		return fail(err, false)
	}
	if !*yes && len(pids) > 0 {
		fmt.Fprintf(os.Stderr, "加 --yes 才会真的结束它们。\n")
	}
	return 0
}

func cmdTzsDoctor(args []string) int {
	fs := flag.NewFlagSet("tzs doctor", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "输出 JSON")
	if err := parseArgs(fs, args); err != nil {
		return 2
	}
	o, err := tzsEngineOptions("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 5
	}
	rep := tzs.Doctor(tzsCtx(), o)
	if *asJSON {
		b, _ := json.MarshalIndent(rep, "", "  ")
		os.Stdout.Write(append(b, '\n'))
	} else {
		// 命令层渲染：Report 只带结构化条目，措辞留给这里。
		for _, c := range rep.Checks {
			fmt.Printf("  [%-4s] %-22s %s\n", c.Level, c.Name, c.Message)
		}
		if !rep.OK() {
			fmt.Fprintf(os.Stderr, "\n自检不通过：%s\n", strings.Join(rep.Failed(), ", "))
		}
	}
	if !rep.OK() {
		return 5
	}
	return 0
}

// ---------- 小工具 ----------

// tzsCtx 是一次命令的 context。引擎那侧的超时由 tzs.Call 自己管（命名管道不支持 deadline），
// 所以这里只用来传递取消。
func tzsCtx() context.Context { return context.Background() }

func replyString(r *tzs.Reply, key string) string {
	var m map[string]any
	if json.Unmarshal(r.Result, &m) != nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}

func printRawReply(r *tzs.Reply, err error) {
	b, _ := json.Marshal(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Stdout.Write(append(b, '\n'))
}

// printHumanReply 把一帧渲染成人看的样子，并返回该给的退出码。
func printHumanReply(fn string, r *tzs.Reply, err error, code int) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return code
	}
	if r.OK {
		b, _ := json.Marshal(r.Result)
		fmt.Printf("%s: %s\n", fn, b)
		return code
	}
	if r.Error != nil {
		fmt.Fprintf(os.Stderr, "%s (%s): %s\n", r.Error.Code, r.Error.Kind, r.Error.Message)
	}
	return code
}
