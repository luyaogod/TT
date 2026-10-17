package fep

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"tt/internal/config"
)

// 「一次说清」的唯一验收（真机冒烟连试两次才凑齐前置，见 L1-结果-2026-10-18.md）：
//
// 第 1 次：服务没起 → USAGE/USAGE 两条报文
// 第 2 次：起了服务 → 没有 SSH 会话
// 随后又冒一条：切换会话：没有 SSH 会话…
//
// 使用者被迫来回试错。判据就一条：**把两个前置都缺的现场交进去，报错里要同时
// 出现两条可直接复制的命令** —— 少一条就又是"跑一次才知道下一个"。
func TestRecStartSaysAllMissingPreconditionsAtOnce(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)
	// 现场：服务没起（第一次失败），于是会话必然也没有（第二次失败）
	d.Service = func() (bool, string) { return false, "127.0.0.1:28670 连不上 —— 服务没起" }
	d.Sessions = func(*cobra.Command) ([]dbgSessionView, error) { return nil, nil }

	err := recStartPreflightWith(new(cobra.Command), "axmt500_wf", "主机109", d)
	if err == nil {
		t.Fatal("两个前置都缺时必须报错，不能开跑（半路失败会白起一个占三个端口的代理）")
	}
	if got := exitOf(t, err); got != exitUsage {
		t.Fatalf("缺前置的退出码 = %d，想要 %d（用法/环境不满足，与 preflight 同一个码）", got, exitUsage)
	}
	msg := err.Error()
	// 两条可复制的命令都必须在场
	for _, want := range []string{
		"还缺 2 项前置",
		"tt debug serve",
		"tt debug start axmt500_wf --env 主机109",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("报错里缺 %q —— 那就又变成「跑一次才知道下一个前置」：\n%s", want, msg)
		}
	}
	// 编号要两条都在（1) 2)），否则人以为只有一项
	if !strings.Contains(msg, "1)") || !strings.Contains(msg, "2)") {
		t.Fatalf("缺 2 项就要列 2 条编号：\n%s", msg)
	}
}

// 反向判据的另一半：只缺一项时不许把"还缺"写成两项（免得人以为要补两处）。
func TestRecStartSaysSingleMissingPrecondition(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)
	d.Sessions = func(*cobra.Command) ([]dbgSessionView, error) { return nil, nil }

	err := recStartPreflightWith(new(cobra.Command), "axmt500_wf", "主机109", d)
	if err == nil {
		t.Fatal("没有会话时必须报错")
	}
	msg := err.Error()
	if !strings.Contains(msg, "还缺 1 项前置") {
		t.Fatalf("只缺会话一项：\n%s", msg)
	}
	if strings.Contains(msg, "tt debug serve") {
		t.Fatalf("服务是好的，不该叫用户去起服务：\n%s", msg)
	}
}

// 全就绪时不许拦：真实路径上这一步在起代理之前，拦错了等于把能跑的命令拦死。
func TestRecStartPreflightPassesWhenReady(t *testing.T) {
	dir := t.TempDir()
	d, _ := readyDeps(t, dir)
	if err := recStartPreflightWith(new(cobra.Command), "axmt500_wf", "主机109", d); err != nil {
		t.Fatalf("四项都就绪时不该拦：%v", err)
	}
}

// 没给 --prog/--env 时，命令里要留下占位符（而不是拼出半条命令）。
func TestMissingTextKeepsPlaceholders(t *testing.T) {
	d := preflightDeps{
		Dir:      func() (string, error) { return t.TempDir(), nil },
		Writable: func(string) error { return nil },
		Service:  func() (bool, string) { return false, "连不上" },
		Sessions: func(*cobra.Command) ([]dbgSessionView, error) { return nil, nil },
		Hosts:    func() ([]config.NamedSsh, error) { return hostsWith("主机109"), nil },
	}
	rep := runPreflight(new(cobra.Command), "", "", d)
	msg := missingText(rep.missing())
	for _, want := range []string{"tt debug start <作业> --env <环境>", "tt env list"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("没给作业/环境时要留占位符 %q：\n%s", want, msg)
		}
	}
}
