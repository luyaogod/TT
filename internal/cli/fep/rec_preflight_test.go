package fep

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"tt/internal/config"
)

// 本文件钉三件事（真机冒烟的痛点，见 .agents/plan/L1-结果-2026-10-18.md）：
//
//	① 体检全就绪 → 清单每项 ok，退出码 0；
//	② 目录不可写 → 退出码 2（用法/环境不满足，不新造码），提示里要出现"完整性/权限"；
//	③ 体检**只检查**：不起会话、不写任何录制文件 —— 用临时目录前后比对。
//
// 全部走假依赖：不起服务、不连真机、不真去改 ACL。

// hostsWith 造一份"配置里有这些环境"的假清单（判"环境是否存在"用）。
func hostsWith(names ...string) []config.NamedSsh {
	out := make([]config.NamedSsh, 0, len(names))
	for _, n := range names {
		out = append(out, config.NamedSsh{Name: n})
	}
	return out
}

// readyDeps 造一份"全部就绪"的假依赖；返回的指针指向计数用的那份提交人。
func readyDeps(t *testing.T, dir string) (preflightDeps, *int) {
	t.Helper()
	calls := 0
	return preflightDeps{
		Dir:      func() (string, error) { return dir, nil },
		Writable: probeWritable, // 真实现：真写一个临时文件再删（这个副作用要靠 ③ 钉住）
		Service:  func() (bool, string) { return true, "127.0.0.1:28670 在跑" },
		Sessions: func(*cobra.Command) ([]dbgSessionView, error) {
			calls++
			return []dbgSessionView{{ID: "s1", Env: "主机109", Prog: "axmt500_wf", State: "idle"}}, nil
		},
		Hosts: func() ([]config.NamedSsh, error) { return hostsWith("主机109", "测试1012"), nil },
	}, &calls
}

// snapshotDir 记下目录里现在有哪些文件（体检前后比对用）。
func snapshotDir(t *testing.T, dir string) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读不到 %s：%v", dir, err)
	}
	for _, e := range ents {
		info, ierr := e.Info()
		if ierr != nil {
			t.Fatalf("读不到 %s 的属性：%v", e.Name(), ierr)
		}
		out[e.Name()] = info.Size()
	}
	return out
}

// exitCodeOf 取任意带 ExitCode() 的错误的退出码。
//
// 为什么不用同包那个 exitOf：preflight 的失败**已经自己把清单打出去了**
// （AlreadyReported，与 tt form verify 同一个手法），所以它故意不是 *output.Error ——
// 走那条路会让根命令 reportError 把同一份东西再打一遍。
func exitCodeOf(t *testing.T, err error) int {
	t.Helper()
	var ec interface{ ExitCode() int }
	if !errors.As(err, &ec) {
		t.Fatalf("错误必须自带退出码，得到 %T: %v", err, err)
	}
	return ec.ExitCode()
}

// ② 全就绪：四项都 ok，退出码 0，人读态说「就绪」。
func TestPreflightAllReady(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)

	rep := runPreflight(new(cobra.Command), "axmt500_wf", "主机109", d)
	if !rep.OK {
		t.Fatalf("四项都该 ok，实得：%+v", rep.Checks)
	}
	if rep.Missing != 0 {
		t.Fatalf("全就绪时 Missing 该是 0，实得 %d", rep.Missing)
	}
	// 清单的顺序就是判据表里的顺序：目录 → 服务 → 会话 → 环境/作业
	want := []string{preflightKeyDir, preflightKeyService, preflightKeySession, preflightKeyEnv}
	if len(rep.Checks) != len(want) {
		t.Fatalf("清单该有 %d 项，实得 %d：%+v", len(want), len(rep.Checks), rep.Checks)
	}
	for i, it := range rep.Checks {
		if it.Key != want[i] {
			t.Fatalf("第 %d 项 key = %q，期望 %q（清单顺序变了，调用方的分支会错位）", i+1, it.Key, want[i])
		}
		if !it.OK {
			t.Fatalf("%s 该 ok：%+v", it.Title, it)
		}
		if it.Why != "" {
			t.Fatalf("%s 通过了就不该留「为什么没过」那一句：%q", it.Title, it.Why)
		}
	}
	if err := preflightExit(rep); err != nil {
		t.Fatalf("全就绪必须退 0，实得错误：%v", err)
	}
	txt := rep.text()
	for _, s := range []string{"就绪", "数据目录可写", "调试服务", "调试会话", "环境/作业"} {
		if !strings.Contains(txt, s) {
			t.Fatalf("人读态里缺 %q：\n%s", s, txt)
		}
	}
	// 环境名要解析成配置里的规范名（拼错时这里会变成原样输入，环境那项就不 ok）
	if rep.Env != "主机109" || rep.Prog != "axmt500_wf" {
		t.Fatalf("清单头该带上这一轮的环境/作业：%+v", rep)
	}
}

// ③ 体检**不起会话、不写录制文件**：目录前后一模一样。
//
// 这一条是"只检查"这个承诺的唯一证据：探活写临时文件是真副作用（必须当场删掉），
// 而录制件落盘是最容易顺手写出去的东西。
func TestPreflightDoesNotStartSessionOrWriteRecordings(t *testing.T) {
	dir := t.TempDir()
	d, calls := readyDeps(t, dir)

	before := snapshotDir(t, dir)
	rep := runPreflight(new(cobra.Command), "axmt500_wf", "主机109", d)
	after := snapshotDir(t, dir)
	if !rep.OK {
		t.Fatalf("假依赖都是就绪的：%+v", rep.Checks)
	}
	if len(after) != len(before) {
		t.Fatalf("体检往目录里写了东西（前后 %d → %d 个文件）：%v", len(before), len(after), after)
	}
	for name, size := range after {
		if old, ok := before[name]; !ok || old != size {
			t.Fatalf("多出来/变了的是 %s（%d 字节）—— 体检不许落任何录制件：%v", name, size, after)
		}
	}
	// 列会话一次（列 ≠ 起）；起会话是要连真机的动作，体检里一条都不许发
	if *calls != 1 {
		t.Fatalf("会话清单该只读一次，实得 %d 次（体检只查有没有，不起、不切）", *calls)
	}
}

// ② 目录不可写：退出码 2，提示里要出现"完整性/权限"字样，且给出具体处置。
//
// 用假实现注入不可写 —— 真去改 ACL 会把跑测试的人自己的目录弄坏。
func TestPreflightUnwritableDirIsUsage(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)
	d.Writable = func(string) error {
		// 真机上 Low 完整性标签与 ACL 拒绝给的是同一句（发现 A 的原文）
		return errors.New(`open C:\Users\x\AppData\Roaming\tt\fep\tt-preflight-974902201.tmp: Access is denied.`)
	}

	rep := runPreflight(new(cobra.Command), "axmt500_wf", "主机109", d)
	if rep.OK {
		t.Fatal("目录不可写时清单不许说 ok —— 后面每个落盘的命令都会 Access is denied")
	}
	if rep.Missing != 1 {
		t.Fatalf("只有目录那一项不 ok，实得 %d 项：%+v", rep.Missing, rep.Checks)
	}
	it := rep.Checks[0]
	if it.Key != preflightKeyDir || it.OK {
		t.Fatalf("第一项该是「目录可写」且不 ok：%+v", it)
	}
	if !strings.Contains(it.Why, "完整性/权限") {
		t.Fatalf("不可写的诊断必须点名「完整性/权限」（否则人只看到 Access is denied，查不出原因）：%q", it.Why)
	}
	// 处置要具体到能抄：Low 标签 + icacls
	for _, s := range []string{"低完整性", "Low", "icacls", "setintegritylevel Medium"} {
		if !strings.Contains(it.Fix, s) {
			t.Fatalf("提示里缺 %q：%q", s, it.Fix)
		}
	}
	err := preflightExit(rep)
	if err == nil {
		t.Fatal("不可写必须非 0 退出")
	}
	if got := exitCodeOf(t, err); got != exitUsage {
		t.Fatalf("不可写的退出码 = %d，想要 %d（用法/环境不满足；码表见 fep.go:25-33，不新造码）", got, exitUsage)
	}
	if !strings.Contains(err.Error(), "完整性/权限") || !strings.Contains(err.Error(), "还缺 1 项前置") {
		t.Fatalf("报错正文要把诊断说清：%q", err.Error())
	}
	// 人读态（不进根命令的 reportError 那条路）也得有这句
	if !strings.Contains(rep.text(), "完整性/权限") {
		t.Fatalf("人读态里没有完整性/权限的提示：\n%s", rep.text())
	}
	// 清单里每一项都带 key（脚本靠 key 分支，不靠标题字面）
	for _, c := range rep.Checks {
		if c.Key == "" || c.Title == "" {
			t.Fatalf("清单项缺 key/title：%+v", c)
		}
	}
}

// 服务在跑但没有会话：会话那一项不 ok，Fix 要把已知的 --prog/--env 填进去。
func TestPreflightNoSessionFillsKnownArgs(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)
	d.Sessions = func(*cobra.Command) ([]dbgSessionView, error) { return nil, nil }

	rep := runPreflight(new(cobra.Command), "axmt500_wf", "主机109", d)
	if rep.OK {
		t.Fatal("没有会话时清单不许说 ok")
	}
	var got preflightItem
	for _, it := range rep.Checks {
		if it.Key == preflightKeySession {
			got = it
		}
	}
	if got.OK {
		t.Fatalf("会话那一项该不 ok：%+v", got)
	}
	want := "tt debug start axmt500_wf --env 主机109"
	if !strings.Contains(got.Fix, want) {
		t.Fatalf("Fix 要能直接复制（把已知的作业/环境填进去）：%q，缺 %q", got.Fix, want)
	}
	if !strings.Contains(missingText(rep.missing()), want) {
		t.Fatalf("「一次说清」里也要出现这条命令：\n%s", missingText(rep.missing()))
	}
}

// 环境不在配置里：环境那一项不 ok，提示要给可用的环境清单（与 tt debug 同一句）。
func TestPreflightUnknownEnv(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)

	rep := runPreflight(new(cobra.Command), "axmt500_wf", "主机909", d)
	if rep.OK {
		t.Fatal("环境不存在时清单不许说 ok —— 后面 debug env 那一枪会白打")
	}
	last := rep.Checks[len(rep.Checks)-1]
	if last.Key != preflightKeyEnv || last.OK {
		t.Fatalf("最后一项该是「环境/作业」且不 ok：%+v", last)
	}
	if !strings.Contains(last.Detail, "主机109") {
		t.Fatalf("要列出可用环境（否则人不知道拼错在哪）：%q", last.Detail)
	}
}
