package fep

import (
	"strings"
	"testing"

	"tt/internal/cli/common"
)

// 「一条命令开始录」的顺序与开关：少一条或顺序反了，症状都是**安静**的
// （界面不出来、或作业跑在别的企业上），而那种错在真机上才发现代价很大。
func TestRecStartStepsOrderAndFlags(t *testing.T) {
	steps := recStartSteps("主机109", "99", "bsft001_wf", "asf", "192.168.14.193:1200")
	if len(steps) != 3 {
		t.Fatalf("该是 环境 → 企业 → 起作业 三步，got %d: %v", len(steps), steps)
	}
	if got := strings.Join(steps[0], " "); got != "debug env 主机109" {
		t.Fatalf("第一步必须先把环境点明白：%q", got)
	}
	if got := strings.Join(steps[1], " "); got != "debug topent 99" {
		t.Fatalf("第二步是企业编号：%q", got)
	}
	start := strings.Join(steps[2], " ")
	for _, want := range []string{"debug start bsft001_wf", "--env 主机109", "--fglserver 192.168.14.193:1200", "-m asf"} {
		if !strings.Contains(start, want) {
			t.Fatalf("起作业那一步缺 %q：%q", want, start)
		}
	}
	// 画面接线必须落在起作业这一步里（它才是"这一轮画面去哪"的开关）
	if !strings.HasSuffix(start, "-m asf") {
		t.Fatalf("模块要接在最后（-m 后面还有别的参数容易被当成它的值）：%q", start)
	}
	// 放行**不在**这里：它是 runRecStartSteps 的最后一步，--no-continue 时整步不跑
	for _, st := range steps {
		if strings.Contains(strings.Join(st, " "), "continue") {
			t.Fatalf("放行不该混在起作业那几步里（那样 --no-continue 就没法拦住它）：%v", st)
		}
	}
}

// 没给 ENT 时不许下发 TOPENT：那是"沿用登录 shell 自己的"，下发个空值会把作业挪走。
func TestRecStartStepsSkipTopentWithoutEnt(t *testing.T) {
	steps := recStartSteps("主机109", "", "bsft001_wf", "", "10.0.0.1:1200")
	if len(steps) != 2 {
		t.Fatalf("没给 ENT 就只有 环境 → 起作业 两步：%v", steps)
	}
	for _, st := range steps {
		if strings.Contains(strings.Join(st, " "), "topent") {
			t.Fatalf("没给 ENT 不许出现 topent：%v", st)
		}
	}
	// 模块为空时不传 -m（让调试线用它自己的默认值），但 --fglserver 必须在
	start := strings.Join(steps[1], " ")
	if strings.Contains(start, "-m") {
		t.Fatalf("模块为空时不该补一个空 -m：%q", start)
	}
	if !strings.Contains(start, "--fglserver 10.0.0.1:1200") {
		t.Fatalf("起作业必须带这一轮的画面接线：%q", start)
	}
}

// selfArgs 必须把 --config 跟着传下去：调用方可能用的是非缺省配置，
// 子进程默默换一份配置的后果是"作业起在另一个环境/另一个企业上"。
func TestSelfArgsPropagatesConfigPath(t *testing.T) {
	old := common.ConfigPath
	defer func() { common.ConfigPath = old }()

	common.ConfigPath = `C:\somewhere\config.json`
	got := selfArgs([]string{"debug", "env", "主机109"})
	want := []string{"--config", `C:\somewhere\config.json`, "debug", "env", "主机109"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("--config 没跟着传：%v", got)
	}
	common.ConfigPath = ""
	got = selfArgs([]string{"debug", "quit"})
	if len(got) != 2 || got[0] != "debug" {
		t.Fatalf("没有 --config 时不该凭空造一个：%v", got)
	}
}
