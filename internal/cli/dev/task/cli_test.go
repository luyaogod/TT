package task

// 命令层的"契约面"测试：参数校验、退出码映射、JSON 信封、工作区定位，
// 以及三个只读命令（status/brief/find/at/search 中属于本文件的那部分）的驱动。
//
// 为什么这些要单独钉：tt dev task 这一组是 agent 的第一入口，它的**失败形态**比成功
// 形态更常被消费（参数打错、环境没点名、工作区找错、退出码被脚本当分支）。这些行为
// 全在 cli.go 里，领域层（taskws/devdata）的测试覆盖不到。
//
// 所有用例都：t.TempDir() 造工作区 + TT_CONFIG 指到临时文件（不读真实 config.json）
// + 不连 SSH/DB/网络。open/update 的成功路径要 SFTP（见 fixture_test.go 顶部说明），
// 这里只驱动它们"出网之前"的分支。

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/dev/taskws"
	"github.com/luyaogod/TT/internal/dev/testutil"
	"github.com/luyaogod/TT/internal/devdata"
	"github.com/luyaogod/TT/internal/output"
)

// TestVerbsCommandSurface 钉住命令面清单：少一个动词 = 用户/agent 在帮助里看不到它。
func TestVerbsCommandSurface(t *testing.T) {
	want := []string{
		"open", "update", "status", "brief", "find",
		"at", "search", "apply", "verify", "relink", "unlock",
	}
	verbs := Verbs()
	if len(verbs) != len(want) {
		t.Fatalf("Verbs() 返回 %d 条命令，期望 %d 条（%v）", len(verbs), len(want), want)
	}
	got := map[string]bool{}
	for _, c := range verbs {
		name := strings.Fields(c.Use)[0]
		got[name] = true
		if strings.TrimSpace(c.Short) == "" {
			t.Errorf("%s 没有 Short（组帮助里会是一条空行）", name)
		}
		if strings.TrimSpace(c.Long) == "" {
			t.Errorf("%s 没有 Long（--help 会只剩一行）", name)
		}
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("命令面缺动词 %q", name)
		}
	}
}

// TestParseLine 钉住 `at` 的行号判定：它决定一条输入走"行号反查"还是"点名查内容"，
// 判错会把 50710 当成点名去查（返回 NOT_FOUND，而用户看到的是一个合法行号）。
func TestParseLine(t *testing.T) {
	good := []struct {
		in   string
		want int
	}{
		{"50710", 50710},
		{" 50710 ", 50710},
		{"axmt500_wf:50710", 50710},
		{"a:b:12", 12}, // 从最后一个冒号切
		{"1", 1},
	}
	for _, c := range good {
		got, ok := parseLine(c.in)
		if !ok || got != c.want {
			t.Errorf("parseLine(%q) = (%d,%v)，期望 (%d,true)", c.in, got, ok, c.want)
		}
	}
	bad := []string{"", "   ", "0", "-3", "abc", "axmt500_wf:", "prog:abc", "1.5"}
	for _, in := range bad {
		if n, ok := parseLine(in); ok {
			t.Errorf("parseLine(%q) 应判为「不是行号」，得到 (%d,true)", in, n)
		}
	}
}

// TestStringSliceFlag 钉住 --only 的"可重复 + 逗号分隔 + 去空白"：
// 少切一段，用户就会以为自己只导出了一个点。
func TestStringSliceFlag(t *testing.T) {
	var s stringSlice
	if err := s.Set("a, b ,, c"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got := strings.Join(s, "|"); got != "a|b|c" {
		t.Fatalf("逗号/空白没切干净：%q", got)
	}
	if err := s.Set("d"); err != nil {
		t.Fatalf("Set(单值): %v", err)
	}
	if got := s.String(); got != "a,b,c,d" {
		t.Errorf("String() = %q，期望 a,b,c,d", got)
	}
	if s.Type() != "strings" {
		t.Errorf("Type() = %q，期望 strings（帮助里的类型名）", s.Type())
	}
}

// TestEntOf 钉住 --ent 的取值口：--at <环境>@<企业> 最终落在这个 flag 上；
// 负数必须当"没给"（否则会拼进材料路径/查询里）。
func TestEntOf(t *testing.T) {
	open := newOpenCmd()
	if got := entOf(open); got != 0 {
		t.Errorf("未指定时 entOf = %d，期望 0", got)
	}
	if err := open.Flags().Set("ent", "99"); err != nil {
		t.Fatal(err)
	}
	if got := entOf(open); got != 99 {
		t.Errorf("--ent 99 时 entOf = %d，期望 99", got)
	}
	if err := open.Flags().Set("ent", "-3"); err != nil {
		t.Fatal(err)
	}
	if got := entOf(open); got != 0 {
		t.Errorf("--ent -3 时 entOf = %d，期望 0（当没给）", got)
	}
	// 没有 ent flag 的命令返回 0（不能 panic）。
	if got := entOf(newStatusCmd()); got != 0 {
		t.Errorf("没有 --ent 的命令 entOf = %d，期望 0", got)
	}
}

// TestFailExitCodeMapping 钉住领域错误 → 退出码码表的映射。
//
// 这是脚本/agent 唯一稳定的分支依据：2 用法、4 未找到/拒绝、5 IO/失败。
// 尤其是"拒绝"与"失败"必须分开 —— 拒绝是用户能改的（dirty/包不一致），失败是环境问题。
func TestFailExitCodeMapping(t *testing.T) {
	cases := []struct {
		name  string
		err   error
		code  string
		exit  int
		inMsg string
	}{
		{"无工作区", taskws.ErrNoWorkspace, "USAGE", exitUsage, "不是工作区"},
		{"工作区已存在", taskws.ErrWorkspaceExists, "USAGE", exitUsage, "工作区已存在"},
		{"任务名非法", taskws.ErrTaskName, "USAGE", exitUsage, "任务名非法"},
		{"程序名非法", devdata.ErrInvalidProg, "USAGE", exitUsage, "程序名非法"},
		{"环境名非法", devdata.ErrInvalidEnv, "USAGE", exitUsage, "环境名非法"},
		{"环境不存在", devdata.ErrEnvNotFound, "USAGE", exitUsage, "环境不存在"},
		{"选项非法", devdata.ErrInvalidOpt, "USAGE", exitUsage, "选项取值非法"},
		{"模块未知", devdata.ErrModuleUnknown, "USAGE", exitUsage, "无法确定程序所属模块"},
		{"视图脏", taskws.ErrDirtyWorkspace, "REFUSED", exitRefused, "未落盘"},
		{"包与基线不一致", taskws.ErrPackageMismatch, "REFUSED", exitRefused, "包与视图基线不一致"},
		{"程序不存在", devdata.ErrProgNotFound, "NOT_FOUND", exitNotFound, "该程序不存在"},
		{"材料不成套", devdata.ErrMaterialIncomplete, "NOT_FOUND", exitNotFound, "材料不成套"},
		{"点位不存在", devdata.ErrPointNotFound, "NOT_FOUND", exitNotFound, "没有匹配的点位"},
		{"被中断", context.Canceled, "CANCELLED", exitFailed, "已中断"},
		{"其他失败", errors.New("磁盘满了"), "FAILED", exitFailed, "磁盘满了"},
		{"包装后的拒绝", fmt.Errorf("refresh: %w", taskws.ErrDirtyWorkspace), "REFUSED", exitRefused, "refresh"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := fail(c.err)
			var oe *output.Error
			if !errors.As(err, &oe) {
				t.Fatalf("fail(%v) 不是 *output.Error：%T", c.err, err)
			}
			if oe.Code != c.code || oe.Exit != c.exit {
				t.Errorf("fail(%v) = %s/%d，期望 %s/%d", c.err, oe.Code, oe.Exit, c.code, c.exit)
			}
			if !strings.Contains(oe.Message, c.inMsg) {
				t.Errorf("错误文案 = %q，应含 %q", oe.Message, c.inMsg)
			}
			if got := exitOf(err); got != c.exit {
				t.Errorf("exitOf = %d，期望 %d", got, c.exit)
			}
		})
	}
}

// TestExitCodeHelpers 钉住"把本组算出的退出码交回 cobra"的那条通路：
// 0 必须是 nil（否则成功会被根命令当失败），非 0 必须带 AlreadyReported
// （否则 stdout 上会出现两个 JSON 对象）。
func TestExitCodeHelpers(t *testing.T) {
	if err := exitCode(0); err != nil {
		t.Fatalf("exitCode(0) = %v，期望 nil", err)
	}
	err := exitCode(4)
	if err == nil {
		t.Fatal("exitCode(4) 不该是 nil")
	}
	if got := exitOf(err); got != 4 {
		t.Errorf("exitCode(4) 的退出码 = %d，期望 4", got)
	}
	var reported interface{ AlreadyReported() bool }
	if !errors.As(err, &reported) || !reported.AlreadyReported() {
		t.Error("非 0 退出码必须实现 AlreadyReported()=true（根命令据此不重复打错误）")
	}
	if err.Error() != "退出码 4" {
		t.Errorf("Error() = %q，期望 退出码 4", err.Error())
	}
	if got := exitCodeOf(5).ExitCode(); got != 5 {
		t.Errorf("exitCodeOf(5).ExitCode() = %d，期望 5", got)
	}
}

// TestResolveRoot 钉住"地址"这条规则：显式 --dir 优先；否则从当前目录往上找 task.json；
// 都找不到就交回当前目录（让 Load 给出"这里不是工作区 + 怎么建"的提示）。
func TestResolveRoot(t *testing.T) {
	base := t.TempDir()

	// ① 显式 --dir（不存在也照用：路径拼错要由后面的 Load 报，不能在这里瞎猜）
	got, err := resolveRoot(filepath.Join(base, "某个", "工作区"))
	if err != nil {
		t.Fatalf("显式 --dir 不该报错: %v", err)
	}
	if want := filepath.Join(base, "某个", "工作区"); got != want {
		t.Errorf("resolveRoot(显式) = %q，期望 %q", got, want)
	}

	// ② 从子目录往上找 task.json
	root := filepath.Join(base, "ws")
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, taskws.TaskFileName), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(nested)
	got, err = resolveRoot("")
	if err != nil {
		t.Fatalf("往上找不该报错: %v", err)
	}
	if got != root {
		t.Errorf("从子目录往上找 = %q，期望 %q", got, root)
	}

	// ③ 找不到：交回当前目录
	empty := t.TempDir()
	t.Chdir(empty)
	got, err = resolveRoot("")
	if err != nil {
		t.Fatalf("找不到工作区不该报错（交给 Load 报）: %v", err)
	}
	if got != empty {
		t.Errorf("找不到时返回 %q，期望当前目录 %q", got, empty)
	}
}

// TestEmitJSONEnvelope 钉住统一信封：ok + 环境信息 + data + totalRows/returned。
//
// 这是 agent 唯一能机读的契约：data 里放结果，totalRows 是**命中总数**（可能被截断），
// returned 是本次回了几条。两者混成一个数，调用方就分不清"只有这些"与"被截断了"。
func TestEmitJSONEnvelope(t *testing.T) {
	setFormat(t, false)
	oldMeta := common.MetaProvider
	common.MetaProvider = func() output.Meta {
		return output.Meta{Account: "tester", Ent: 9}
	}
	t.Cleanup(func() { common.MetaProvider = oldMeta })

	out := renderOutput(t, func() {
		if err := emitJSON("测试区", map[string]any{"k": "v", "n": 3}, 7, 2); err != nil {
			t.Fatalf("emitJSON: %v", err)
		}
	})
	if n := strings.Count(out, "\n"); n != 1 {
		t.Errorf("信封应为单行紧凑 JSON，得到 %d 行：%q", n, out)
	}
	env := decodeEnvelope(t, out)
	if ok, _ := env["ok"].(bool); !ok {
		t.Errorf("ok 应为 true：%v", env)
	}
	if env["source"] != "workspace" {
		t.Errorf("source = %v，期望 workspace", env["source"])
	}
	if env["env"] != "测试区" {
		t.Errorf("env = %v，期望 测试区", env["env"])
	}
	if env["account"] != "tester" || num(t, env["ent"]) != 9 {
		t.Errorf("MetaProvider 的环境信息没带进信封：%v", env)
	}
	if num(t, env["totalRows"]) != 7 || num(t, env["returned"]) != 2 {
		t.Errorf("totalRows/returned = %v/%v，期望 7/2", env["totalRows"], env["returned"])
	}
	d := dataOf(t, env)
	if d["k"] != "v" || num(t, d["n"]) != 3 {
		t.Errorf("data 载荷不对：%v", d)
	}

	// 空结果：data 必须是空数组（而不是缺键 / null），计数为 0。
	out = renderOutput(t, func() {
		if err := emitJSON("测试区", []string{}, 0, 0); err != nil {
			t.Fatalf("emitJSON(空): %v", err)
		}
	})
	env = decodeEnvelope(t, out)
	if num(t, env["totalRows"]) != 0 || num(t, env["returned"]) != 0 {
		t.Errorf("空结果的计数应为 0：%v", env)
	}
	arr, ok := env["data"].([]any)
	if !ok || len(arr) != 0 {
		t.Errorf("空结果的 data 应为空数组，得到 %T %v", env["data"], env["data"])
	}
}

// TestLinkReq 钉住"工作区三个落点显式给，查询就不碰配置"：
// 少给一个落点，brief/find/at/search 就会回落去读 devdata/<环境>/，工作区里的材料被无视。
func TestLinkReq(t *testing.T) {
	task := &taskws.Task{Env: "e", Prog: "p"}
	p := taskws.Paths{Root: filepath.Join("R", "root")}
	lr := linkReq("/cfg/c.json", task, p)
	if lr.CfgPath != "/cfg/c.json" || lr.Env != "e" || lr.Prog != "p" {
		t.Errorf("身份字段丢了：%+v", lr)
	}
	if lr.Dir != p.Materials() || lr.Out != p.Index() || lr.Mdl != p.Mdl() {
		t.Errorf("三个落点应显式给：Dir=%q Out=%q Mdl=%q", lr.Dir, lr.Out, lr.Mdl)
	}
}

// TestInterruptAndProgress 钉住两条"打到 stderr、不影响 stdout"的旁路：
// Ctrl+C 能被领域层看见（否则临时目录不清），进度只在 stderr（stdout 留给结果）。
func TestInterruptAndProgress(t *testing.T) {
	ctx, stop := interruptContext()
	if ctx.Err() != nil {
		t.Fatalf("刚建好就取消了: %v", ctx.Err())
	}
	stop()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("stop() 后 ctx.Err() = %v，期望 Canceled", ctx.Err())
	}

	progress := stderrProgress()
	out := captureStderr(t, func() {
		progress(devdata.Progress{})
		progress(devdata.Progress{Message: "拉材料"})
		progress(devdata.Progress{Message: "下载", Done: 2, Total: 5})
	})
	if !strings.Contains(out, "拉材料") || !strings.Contains(out, "下载  [2/5]") {
		t.Errorf("进度文案不对：%q", out)
	}
	if strings.Contains(out, "\n\n") {
		t.Errorf("空消息不该打空行：%q", out)
	}
}

// TestOpenArgValidation 覆盖 `open` 出网之前的所有用法校验。
//
// open 的成功路径要 SFTP（真实服务器），命令层没有注入缝 —— 这一条在报告里显式说明。
// 但它的**前置校验**恰恰是日常最常踩的：环境没点名、程序名/任务名非法、目录已存在。
func TestOpenArgValidation(t *testing.T) {
	t.Run("没点名环境", func(t *testing.T) {
		isolateConfig(t, false)
		setFormat(t, false)
		common.Env = ""
		code, _, _, err := runCmd(t, newOpenCmd, fxProg)
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2", code)
		}
		if err == nil || !strings.Contains(err.Error(), "未指定环境") {
			t.Errorf("错误文案应点名「未指定环境」：%v", err)
		}
	})

	t.Run("配置文件找不到", func(t *testing.T) {
		isolateConfig(t, false)
		setFormat(t, false)
		common.Env = fxEnv
		code, _, _, err := runCmd(t, newOpenCmd, fxProg)
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2", code)
		}
		if err == nil || !strings.Contains(err.Error(), "配置文件未找到") {
			t.Errorf("错误文案应含「配置文件未找到」：%v", err)
		}
	})

	t.Run("程序名非法", func(t *testing.T) {
		isolateConfig(t, true)
		setFormat(t, false)
		common.Env = fxEnv
		code, _, _, err := runCmd(t, newOpenCmd, "BAD-PROG")
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2", code)
		}
		if err == nil || !strings.Contains(err.Error(), "程序名非法") {
			t.Errorf("错误文案应含「程序名非法」：%v", err)
		}
	})

	t.Run("任务名非法", func(t *testing.T) {
		isolateConfig(t, true)
		setFormat(t, false)
		common.Env = fxEnv
		code, _, _, err := runCmd(t, newOpenCmd, fxProg, "--task", "../逃出去",
			"--dir", filepath.Join(t.TempDir(), "w"))
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2", code)
		}
		if err == nil || !strings.Contains(err.Error(), "任务名非法") {
			t.Errorf("错误文案应含「任务名非法」：%v", err)
		}
	})

	t.Run("目录已存在", func(t *testing.T) {
		isolateConfig(t, true)
		setFormat(t, false)
		common.Env = fxEnv
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "别动我.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		code, _, _, err := runCmd(t, newOpenCmd, fxProg, "--dir", dir)
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2", code)
		}
		if err == nil || !strings.Contains(err.Error(), "工作区已存在") {
			t.Errorf("错误文案应含「工作区已存在」：%v", err)
		}
	})

	t.Run("环境不存在且失败不留半套", func(t *testing.T) {
		isolateConfig(t, true) // 空配置：没有任何环境
		setFormat(t, false)
		common.Env = "不存在的环境"
		dir := filepath.Join(t.TempDir(), "w")
		code, _, _, err := runCmd(t, newOpenCmd, fxProg, "--dir", dir)
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（环境不存在属用法错）", code)
		}
		if err == nil || !strings.Contains(err.Error(), "环境不存在") {
			t.Errorf("错误文案应含「环境不存在」：%v", err)
		}
		if _, serr := os.Stat(dir); serr == nil {
			t.Error("拉料失败后不该留下半套工作区目录")
		}
	})

	t.Run("缺参数与多余参数", func(t *testing.T) {
		isolateConfig(t, false)
		setFormat(t, false)
		code, _, _, err := runCmd(t, newOpenCmd)
		if code == 0 || err == nil {
			t.Fatalf("缺参数应失败，得到 code=%d err=%v", code, err)
		}
		code, _, _, err = runCmd(t, newOpenCmd, "a", "b")
		if code == 0 || err == nil {
			t.Fatalf("多余参数应失败，得到 code=%d err=%v", code, err)
		}
	})
}

// TestStatusCommand 驱动 status 的两条输出形态 + 工作区定位。
func TestStatusCommand(t *testing.T) {
	f := newFx(t)
	isolateConfig(t, false)

	t.Run("JSON信封", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newStatusCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		env := decodeEnvelope(t, out)
		if env["env"] != f.env || env["source"] != "workspace" {
			t.Errorf("信封环境信息不对：env=%v source=%v", env["env"], env["source"])
		}
		d := dataOf(t, env)
		if d["prog"] != f.prog || d["dir"] != f.root {
			t.Errorf("data 身份不对：prog=%v dir=%v", d["prog"], d["dir"])
		}
		mats := d["materials"].(map[string]any)
		if mats["fetchedAt"] == "" || num(t, mats["files"]) != 5 || num(t, mats["dzaa002"]) != 7 {
			t.Errorf("材料状态不对：%v", mats)
		}
		idx := d["index"].(map[string]any)
		if !idx["present"].(bool) || !idx["fresh"].(bool) {
			t.Errorf("索引应存在且新鲜：%v", idx)
		}
		ws := d["ws"].(map[string]any)
		if !ws["present"].(bool) || ws["dirty"].(bool) {
			t.Errorf("视图应存在且干净：%v", ws)
		}
		pkg := d["pkg"].(map[string]any)
		if !pkg["present"].(bool) || !pkg["matchesBase"].(bool) {
			t.Errorf("包应存在且与基线一致：%v", pkg)
		}
		mdl := d["mdl"].(map[string]any)
		if !mdl["present"].(bool) || num(t, mdl["templates"]) != 1 {
			t.Errorf("模板切片状态不对：%v", mdl)
		}
	})

	t.Run("人读输出", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newStatusCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		for _, want := range []string{
			f.prog, "材料", "索引", "最新", "视图", "干净", "包", "未 apply", "模板", "切片",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("人读输出缺 %q：\n%s", want, out)
			}
		}
	})

	t.Run("工作区里省略dir", func(t *testing.T) {
		setFormat(t, false)
		t.Chdir(filepath.Join(f.ws, "..")) // 工作区根
		code, out, _, err := runCmd(t, newStatusCmd)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		if got := dataOf(t, decodeEnvelope(t, out))["prog"]; got != f.prog {
			t.Errorf("往上找 task.json 失败：prog=%v", got)
		}
	})

	t.Run("非工作区退2", func(t *testing.T) {
		setFormat(t, false)
		t.Chdir(t.TempDir())
		code, _, _, err := runCmd(t, newStatusCmd)
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2", code)
		}
		if err == nil || !strings.Contains(err.Error(), "不是工作区") {
			t.Errorf("错误文案应含「不是工作区」：%v", err)
		}
	})
}

// TestUpdateCommand 覆盖 update 在**出网之前**的全部闸门。
//
// 成功路径要 devdata.Check（强制 probe → SFTP），命令层无注入缝，故这里只钉：
// 非工作区、视图脏拒绝、包被换拒绝、--check 仍要先有环境。
func TestUpdateCommand(t *testing.T) {
	t.Run("非工作区退2", func(t *testing.T) {
		isolateConfig(t, false)
		setFormat(t, false)
		t.Chdir(t.TempDir())
		code, _, _, err := runCmd(t, newUpdateCmd)
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（err=%v）", code, err)
		}
		if err == nil || !strings.Contains(err.Error(), "不是工作区") {
			t.Errorf("错误文案应含「不是工作区」：%v", err)
		}
	})

	t.Run("视图脏时拒绝", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false)
		edited := filepath.Join(f.ws, "prog.full.4gl")
		b, err := os.ReadFile(edited)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(edited, append(b, []byte("\n# 本地未落盘一行\n")...), 0o644); err != nil {
			t.Fatal(err)
		}
		code, _, _, uerr := runCmd(t, newUpdateCmd, "--dir", f.root)
		if code != exitRefused {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, uerr)
		}
		if uerr == nil || !strings.Contains(uerr.Error(), "未落盘的改动") {
			t.Errorf("错误文案应说明未落盘改动：%v", uerr)
		}
	})

	t.Run("包与基线不一致时拒绝", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false)
		replacePkg(t, f, testutil.NormalEntriesWith(fxProg, map[string]string{"env": "s"}))
		code, _, _, uerr := runCmd(t, newUpdateCmd, "--dir", f.root)
		if code != exitRefused {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, uerr)
		}
		if uerr == nil || !strings.Contains(uerr.Error(), "包与视图基线不一致") {
			t.Errorf("错误文案应说明包不一致：%v", uerr)
		}
	})

	t.Run("check仍要环境", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, true) // 空配置：Check 第一步就因环境不存在失败（不连 SFTP）
		setFormat(t, false)
		code, _, _, uerr := runCmd(t, newUpdateCmd, "--dir", f.root, "--check")
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（err=%v）", code, uerr)
		}
		if uerr == nil || !strings.Contains(uerr.Error(), "环境不存在") {
			t.Errorf("错误文案应含「环境不存在」：%v", uerr)
		}
	})

	t.Run("多余参数", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false)
		code, _, _, err := runCmd(t, newUpdateCmd, "--dir", f.root, "多余")
		if code == 0 || err == nil {
			t.Fatalf("update 不接受位置参数，得到 code=%d err=%v", code, err)
		}
	})
}

// TestBriefCommand 驱动 brief：JSON 载荷的关键计数 + 人读的五个判断块。
func TestBriefCommand(t *testing.T) {
	f := newFx(t)
	isolateConfig(t, false)

	t.Run("JSON载荷", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newBriefCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		env := decodeEnvelope(t, out)
		if env["env"] != f.env {
			t.Errorf("env = %v，期望 %v", env["env"], f.env)
		}
		d := dataOf(t, env)
		if d["prog"] != f.prog {
			t.Errorf("prog = %v", d["prog"])
		}
		mode := d["mode"].(map[string]any)
		if mode["type"] != "t01" || mode["module"] != "CXM" || mode["identity"] != "c" {
			t.Errorf("判型信息不对：%v", mode)
		}
		if pages := d["pages"].([]any); len(pages) != 1 {
			t.Errorf("页签数 = %d，期望 1", len(pages))
		}
		custom := d["customPoints"].([]any)
		if len(custom) != 2 || num(t, d["customTotal"]) != 2 {
			t.Errorf("客制点 = %d / total %v，期望 2/2", len(custom), d["customTotal"])
		}
		if len(d["orphans"].([]any)) != 2 {
			t.Errorf("孤儿声明数 = %d，期望 2", len(d["orphans"].([]any)))
		}
		drift := d["slices"].(map[string]any)["drift"].([]any)
		if len(drift) != 1 {
			t.Errorf("切片脱版数 = %d，期望 1", len(drift))
		}
	})

	t.Run("top截断", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newBriefCmd, "--dir", f.root, "--top", "1")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		d := dataOf(t, decodeEnvelope(t, out))
		if got := len(d["customPoints"].([]any)); got != 1 {
			t.Errorf("--top 1 后客制点 = %d，期望 1", got)
		}
		notes := fmt.Sprint(d["notes"])
		if !strings.Contains(notes, "只列最大 1 个") {
			t.Errorf("截断必须在 notes 里说明（不能静默）：%v", notes)
		}
	})

	t.Run("人读输出", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newBriefCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		for _, want := range []string{"页签", "客制逻辑", "风险", "切片脱版", "意图 → 落点", "函数"} {
			if !strings.Contains(out, want) {
				t.Errorf("人读输出缺 %q：\n%s", want, out)
			}
		}
	})
}

// TestFindCommand 驱动 find：意图 + 字段命中的口径、label 命中、limit 截断、无参清单。
func TestFindCommand(t *testing.T) {
	f := newFx(t)
	isolateConfig(t, false)

	t.Run("意图加字段", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newFindCmd,
			"--dir", f.root, "--intent", "check", "--field", "xmdduc001")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		env := decodeEnvelope(t, out)
		if num(t, env["totalRows"]) != 1 || num(t, env["returned"]) != 1 {
			t.Errorf("totalRows/returned = %v/%v，期望 1/1", env["totalRows"], env["returned"])
		}
		d := dataOf(t, env)
		if d["fieldBy"] != "column" {
			t.Errorf("fieldBy = %v，期望 column", d["fieldBy"])
		}
		if pts := d["points"].([]any); len(pts) != 1 {
			t.Errorf("命中点 = %d，期望 1", len(pts))
		}
	})

	t.Run("画面中文名命中", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newFindCmd, "--dir", f.root, "--field", "刻号")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		d := dataOf(t, decodeEnvelope(t, out))
		if d["fieldBy"] != "label" {
			t.Errorf("fieldBy = %v，期望 label", d["fieldBy"])
		}
		labels, _ := d["labels"].(map[string]any)
		if labels["xmdduc001"] != "刻号" {
			t.Errorf("labels 缺 刻号：%v", labels)
		}
		if len(d["orphans"].([]any)) == 0 {
			t.Error("同页上的孤儿声明应被带出来（现场最常见的「写了不生效」）")
		}
	})

	t.Run("limit截断", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newFindCmd, "--dir", f.root, "--intent", "check", "--limit", "1")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		if !strings.Contains(out, "命中 2 个（列 1）") {
			t.Errorf("截断时必须同时给总数与列数：\n%s", out)
		}
	})

	t.Run("无参列意图清单", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newFindCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		for _, want := range []string{"意图定位", "没有命中", "可用意图", "check", "lookup"} {
			if !strings.Contains(out, want) {
				t.Errorf("无参输出缺 %q：\n%s", want, out)
			}
		}
	})

	t.Run("包有apply改动时加说明", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false)
		// 包内容被换（=apply 过）→ 点位数/行数仍来自 materials，必须给人一句说明，
		// 否则"改了怎么统计没变"会被读成"改动丢了"。
		replacePkg(t, f, testutil.NormalEntriesWith(fxProg, map[string]string{"env": "s"}))
		code, out, _, err := runCmd(t, newFindCmd, "--dir", f.root, "--field", "xmdduc001")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		d := dataOf(t, decodeEnvelope(t, out))
		note, _ := d["note"].(string)
		if !strings.Contains(note, "不计入统计") {
			t.Errorf("包内已 apply 时应加统计口径说明：%q", note)
		}
	})
	t.Run("JSON空结果", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newFindCmd, "--dir", f.root)
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		env := decodeEnvelope(t, out)
		if num(t, env["totalRows"]) != 0 || num(t, env["returned"]) != 0 {
			t.Errorf("空结果计数应为 0：%v", env)
		}
		// 空结果的 JSON 数组必须是 `[]` 而不是 `null`：agent 侧按数组消费，
		// null 会直接炸（find/search/brief 已统一为空切片）。
		if pts, ok := dataOf(t, env)["points"].([]any); !ok || len(pts) != 0 {
			t.Errorf("空结果的 points 应为 []：%v（类型 %T）", dataOf(t, env)["points"], dataOf(t, env)["points"])
		}
	})
}

// TestAtCommand 覆盖 at 的三条主要分支：行号反查、点名看包内 .tap、--from 4gl。
func TestAtCommand(t *testing.T) {
	f := newFx(t)
	isolateConfig(t, false)
	line := f.lineInside(t, "input.a.page73.xmdduc001")

	t.Run("行号反查JSON", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, strconv.Itoa(line))
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		d := dataOf(t, decodeEnvelope(t, out))
		if num(t, d["line"]) != line {
			t.Errorf("line = %v，期望 %d", d["line"], line)
		}
		pt := d["point"].(map[string]any)
		if pt["name"] != "input.a.page73.xmdduc001" {
			t.Errorf("归属点 = %v", pt["name"])
		}
		if verdict, _ := d["verdict"].(string); !strings.Contains(verdict, "客制") {
			t.Errorf("结论应说明是客制（重产保留）：%q", verdict)
		}
	})

	t.Run("行号反查人读带切片", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, strconv.Itoa(line))
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		for _, want := range []string{fmt.Sprintf("L%d", line), "客制", "切片 a00 v1"} {
			if !strings.Contains(out, want) {
				t.Errorf("输出缺 %q：\n%s", want, out)
			}
		}
	})

	t.Run("行号越界退2", func(t *testing.T) {
		setFormat(t, false)
		code, _, _, err := runCmd(t, newAtCmd, "--dir", f.root, "99999")
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（err=%v）", code, err)
		}
		if err == nil || !strings.Contains(err.Error(), "超出范围") {
			t.Errorf("错误文案应含「超出范围」：%v", err)
		}
	})

	t.Run("空参数退2", func(t *testing.T) {
		setFormat(t, false)
		code, _, _, err := runCmd(t, newAtCmd, "--dir", f.root, "")
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（err=%v）", code, err)
		}
		if err == nil || !strings.Contains(err.Error(), "要写成") {
			t.Errorf("错误文案应指路两种写法：%v", err)
		}
	})

	t.Run("点名看包内tap", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "global.memo")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		for _, want := range []string{"global.memo", "包内 .tap", "标准"} {
			if !strings.Contains(out, want) {
				t.Errorf("输出缺 %q：\n%s", want, out)
			}
		}
	})

	t.Run("点名JSON", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "global.memo")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		d := dataOf(t, decodeEnvelope(t, out))
		if d["source"] != "pkg" {
			t.Errorf("source = %v，期望 pkg（默认查包内最新真相）", d["source"])
		}
		if pts, _ := d["points"].([]any); len(pts) != 1 {
			t.Errorf("points = %v，期望 1 条", d["points"])
		}
	})

	t.Run("中文名点名", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "刻号")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		if !strings.Contains(out, "input.a.page73.xmdduc001") || !strings.Contains(out, "包内 .tap") {
			t.Errorf("画面中文名应能落到列名并打到包内正文：\n%s", out)
		}
	})

	t.Run("孤儿声明", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "function."+fxProg+"_calc")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		if !strings.Contains(out, "重产不会保留") {
			t.Errorf("孤儿必须明确说重产不保留：\n%s", out)
		}
	})

	t.Run("模板洞退回产物", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "input.b.page73.extra001")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		if !strings.Contains(out, "没有 .tap 声明（模板洞）") {
			t.Errorf("包里没有声明时应退回产物并说明原因：\n%s", out)
		}
	})

	t.Run("from4gl看产物人读", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "global.memo", "--from", "4gl")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		if !strings.Contains(out, "匹配方式") || !strings.Contains(out, "global.memo") {
			t.Errorf("--from 4gl 的人读输出应带匹配方式与正文：\n%s", out)
		}
	})

	t.Run("from非法退2", func(t *testing.T) {
		setFormat(t, false)
		code, _, _, err := runCmd(t, newAtCmd, "--dir", f.root, "global.memo", "--from", "bogus")
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（err=%v）", code, err)
		}
		if err == nil || !strings.Contains(err.Error(), "--from 只能是") {
			t.Errorf("错误文案应点名 --from 的取值：%v", err)
		}
	})

	t.Run("from4gl找不到退4", func(t *testing.T) {
		setFormat(t, false)
		code, _, _, err := runCmd(t, newAtCmd, "--dir", f.root, "nope.zzz", "--from", "4gl")
		if code != exitNotFound {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, err)
		}
	})

	t.Run("from4gl JSON", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "global.memo", "--from", "4gl", "--max", "2")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		d := dataOf(t, decodeEnvelope(t, out))
		if d["match"] == "" || len(d["points"].([]any)) == 0 {
			t.Errorf("--from 4gl 的 JSON 应带匹配方式与点位：%v", d)
		}
	})

	t.Run("孤儿JSON", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "function."+fxProg+"_calc")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		env := decodeEnvelope(t, out)
		if num(t, env["returned"]) != 1 {
			t.Errorf("孤儿 JSON 的 returned = %v，期望 1", env["returned"])
		}
	})

	t.Run("模板洞JSON", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "input.b.page73.extra001")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		d := dataOf(t, decodeEnvelope(t, out))
		if d["query"] != "input.b.page73.extra001" {
			t.Errorf("退回产物时应带 query：%v", d)
		}
	})

	t.Run("多命中说明无声明点", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newAtCmd, "--dir", f.root, "page73")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		// page73 子串能命中三个点，其中 extra001 是生成器自加、包里没有声明。
		// 有正文的点照打，同时必须说明"另外还有几个没声明"（而不是静默少打）。
		if !strings.Contains(out, "没有 .tap 声明") {
			t.Errorf("多命中时要点名哪些没有声明：\n%s", out)
		}
	})

	t.Run("材料读不出来退5", func(t *testing.T) {
		f := newFx(t)
		isolateConfig(t, false)
		setFormat(t, false)
		// 清单坏成非法 JSON：这不是"没有材料"（那是 ErrProgNotFound→4），
		// 而是"读不动"——必须退 5，别把 IO 损坏说成程序不存在。
		if err := os.WriteFile(filepath.Join(f.mats, "manifest.json"), []byte("{坏 JSON"), 0o644); err != nil {
			t.Fatal(err)
		}
		code, _, _, err := runCmd(t, newAtCmd, "--dir", f.root, "global.memo")
		if code != exitFailed {
			t.Fatalf("退出码 = %d，期望 5（err=%v）", code, err)
		}
	})
	t.Run("点名不存在退4", func(t *testing.T) {
		setFormat(t, false)
		code, _, _, err := runCmd(t, newAtCmd, "--dir", f.root, "nope.zzz")
		if code != exitNotFound {
			t.Fatalf("退出码 = %d，期望 4（err=%v）", code, err)
		}
		if err == nil || !strings.Contains(err.Error(), "包里没有匹配") {
			t.Errorf("错误文案应说明包里没有匹配点：%v", err)
		}
	})
}

// TestSearchCommand 驱动 search：产物侧 + 包内 .tap 侧两段命中、scope、大小写、limit。
func TestSearchCommand(t *testing.T) {
	f := newFx(t)
	isolateConfig(t, false)

	t.Run("JSON两段命中", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newSearchCmd, "--dir", f.root, "xmdduc")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		env := decodeEnvelope(t, out)
		d := dataOf(t, env)
		total := num(t, d["total"])
		if total == 0 {
			t.Error("产物侧应有命中")
		}
		if num(t, d["tapTotal"]) == 0 {
			t.Error(".tap 侧应有命中（声明/占位符里都有 xmdduc）")
		}
		if d["tapSource"] != "pkg" {
			t.Errorf("tapSource = %v，期望 pkg（工作区优先用包内 .tap）", d["tapSource"])
		}
		if num(t, env["totalRows"]) != total {
			t.Errorf("信封 totalRows = %v，应与 data.total %d 一致", env["totalRows"], total)
		}
	})

	t.Run("只看tap范围", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newSearchCmd, "--dir", f.root, "xmdduc", "--scope", "tap")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		if strings.Contains(out, "产物命中") {
			t.Errorf("--scope tap 不该打「产物命中 0 条」（那是被排除，不是结果）：\n%s", out)
		}
		if !strings.Contains(out, ".tap 命中") {
			t.Errorf("应打 .tap 命中：\n%s", out)
		}
	})

	t.Run("大小写敏感", func(t *testing.T) {
		setFormat(t, false)
		code, out, _, err := runCmd(t, newSearchCmd, "--dir", f.root, "XMDDUC", "--case")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		env := decodeEnvelope(t, out)
		if num(t, env["totalRows"]) != 0 {
			t.Errorf("--case 下大写串不该命中，totalRows=%v", env["totalRows"])
		}
	})

	t.Run("limit截断提示", func(t *testing.T) {
		setFormat(t, true)
		code, out, _, err := runCmd(t, newSearchCmd, "--dir", f.root, "xmdduc", "--limit", "1")
		if code != 0 {
			t.Fatalf("退出码 = %d（err=%v）", code, err)
		}
		if !strings.Contains(out, "只列前") {
			t.Errorf("截断必须提示（不能静默）：\n%s", out)
		}
	})

	t.Run("空检索串退2", func(t *testing.T) {
		setFormat(t, false)
		code, _, _, err := runCmd(t, newSearchCmd, "--dir", f.root, "")
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（err=%v）", code, err)
		}
		if err == nil || !strings.Contains(err.Error(), "检索串不能为空") {
			t.Errorf("错误文案应含「检索串不能为空」：%v", err)
		}
	})

	t.Run("非工作区退2", func(t *testing.T) {
		setFormat(t, false)
		t.Chdir(t.TempDir())
		code, _, _, err := runCmd(t, newSearchCmd, "MAIN")
		if code != exitUsage {
			t.Fatalf("退出码 = %d，期望 2（err=%v）", code, err)
		}
	})
}

// TestReadCommandsBadWorkspace 钉住四个读命令对"工作区坏了"的收场。
//
// 它们都在 resolveRoot 之后先 Load(task.json)：schema 不对（手改过/版本不符）必须
// 退用法错 2 并说清是 schema —— 而不是让后面的 devdata 报一个"材料缺失"的假原因。
func TestReadCommandsBadWorkspace(t *testing.T) {
	isolateConfig(t, false)
	setFormat(t, false)

	root := t.TempDir()
	// 合法 JSON、但 schema 不是本版认的那个。
	bad := `{"schema":"tt.dev.ws/v9","prog":"p","env":"e"}`
	if err := os.WriteFile(filepath.Join(root, taskws.TaskFileName), []byte(bad), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func() *cobra.Command{
		"status": newStatusCmd,
		"brief":  newBriefCmd,
		"find":   newFindCmd,
		"at":     newAtCmd,
		"search": newSearchCmd,
		"apply":  newApplyCmd,
		"verify": newVerifyCmd,
		"relink": newRelinkCmd,
		"unlock": newUnlockCmd,
	}
	for name, mk := range cases {
		t.Run(name, func(t *testing.T) {
			args := []string{"--dir", root}
			switch name {
			case "at":
				args = append(args, "50710")
			case "search":
				args = append(args, "MAIN")
			}
			code, _, _, err := runCmd(t, mk, args...)
			// task.json 的 schema 不匹配是"本版工具读不懂这份工作区"——用户可处理
			// （换新版 tt / 重新 open），按用法/版本错退 2，不是 IO 失败 5。
			// 领域侧用 taskws.ErrWorkspaceVersion 标记，CLI 在 fail() 里映射。
			if code != exitUsage {
				t.Fatalf("退出码 = %d，期望 2（err=%v）", code, err)
			}
			if err == nil || !strings.Contains(err.Error(), "schema") {
				t.Errorf("错误文案应点名 schema：%v", err)
			}
		})
	}
}

// TestReadCommandsBrokenMaterials 钉住"task.json 没问题、但材料读不动"的收场：
// 必须退 5（IO/环境失败）并保留领域错误原文，而不是假装工作区不存在（2）或没有命中（4）。
func TestReadCommandsBrokenMaterials(t *testing.T) {
	isolateConfig(t, false)
	setFormat(t, false)

	f := newFx(t)
	// 先确保证 task.json 合法（newFx 已写好），只把材料清单弄坏。
	if err := os.WriteFile(filepath.Join(f.mats, "manifest.json"), []byte("{坏 JSON"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, mk := range map[string]func() *cobra.Command{
		"brief": newBriefCmd,
		"find":  newFindCmd,
	} {
		t.Run(name, func(t *testing.T) {
			code, _, _, err := runCmd(t, mk, "--dir", f.root)
			if code != exitFailed {
				t.Fatalf("退出码 = %d，期望 5（err=%v）", code, err)
			}
			if err == nil || !strings.Contains(err.Error(), "manifest") {
				t.Errorf("错误文案应点名材料清单：%v", err)
			}
		})
	}
}

// TestAtBrokenPackageFails 钉住"包读不出来"时 at 的收场：退 5（IO 失败），
// 而不是装作没有这个点位（那会让人以为代码不存在）。
func TestAtBrokenPackageFails(t *testing.T) {
	f := newFx(t)
	isolateConfig(t, false)
	setFormat(t, false)
	if err := os.Remove(f.pkg); err != nil {
		t.Fatal(err)
	}
	code, _, _, err := runCmd(t, newAtCmd, "--dir", f.root, "global.memo")
	if code != exitFailed {
		t.Fatalf("退出码 = %d，期望 5（err=%v）", code, err)
	}
	if err == nil {
		t.Fatal("包丢了必须报错，不能静默当成没有点位")
	}
}
