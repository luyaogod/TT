package fep

import (
	"strings"
	"testing"

	"tt/internal/cli/common"
)

// 起录那几步的顺序与开关。第一步是**先收口当前这一轮**（见下），它是 optional：
// 手上没有会话时 `debug quit` 会报"没有活动会话"，那不是错。
func TestRecStartStepsOrderAndFlags(t *testing.T) {
	steps := recStartSteps("主机109", "99", "bsft001_wf", "asf", "192.168.14.193:1200", "")
	if len(steps) != 4 {
		t.Fatalf("该是 收口 → 环境 → 企业 → 起作业 四步，got %d: %v", len(steps), steps)
	}
	if got := strings.Join(steps[0].args, " "); got != "debug quit" {
		t.Fatalf("第一步必须先把上一轮收口（否则第二步的 TOPENT 会被拒：只有空闲会话才让设）：%q", got)
	}
	if !steps[0].optional {
		t.Fatal("收口那一步必须是 optional：没有会话时 quit 报错是正常结果，不该让整轮录制失败")
	}
	if got := strings.Join(steps[1].args, " "); got != "debug env 主机109" {
		t.Fatalf("第二步把环境点明白：%q", got)
	}
	if got := strings.Join(steps[2].args, " "); got != "debug topent 99" {
		t.Fatalf("第三步是企业编号：%q", got)
	}
	start := strings.Join(steps[3].args, " ")
	for _, want := range []string{"debug start bsft001_wf", "--env 主机109", "--fglserver 192.168.14.193:1200", "-m asf"} {
		if !strings.Contains(start, want) {
			t.Fatalf("起作业那一步缺 %q：%q", want, start)
		}
	}
	// 画面接线必须落在起作业这一步里（它才是"这一轮画面去哪"的开关）
	if !strings.HasSuffix(start, "-m asf") {
		t.Fatalf("模块要接在最后（-m 后面还有别的参数容易被当成它的值）：%q", start)
	}
	// 没给别名时**一个字都不许多**（H2：既有调用方的子进程参数逐字不变）
	if strings.Contains(start, "--name") {
		t.Fatalf("没给别名时不许拼 --name：%q", start)
	}
	// 放行**不在**这里：它是 runRecStartSteps 的最后一步，--no-continue 时整步不跑
	for _, st := range steps {
		if strings.Contains(strings.Join(st.args, " "), "continue") {
			t.Fatalf("放行不该混在起作业那几步里（那样 --no-continue 就没法拦住它）：%v", st.args)
		}
	}
}

// T5.4（契约 C7.3）：`fep` 跨进程传目标 = 把**固定别名**拼进起作业那一步的 `--name`。
//
// 修前这里没有别名这一路（`recStartSteps` 连参数都没有），所以这一条在修前**编译不过**
// —— 那就是"修前红"：子进程根本收不到"该落到哪条会话"。
func TestRecStartStepsCarryAlias(t *testing.T) {
	steps := recStartSteps("主机109", "99", "bsft001_wf", "asf", "10.0.0.1:1200", "fep-fep-20261016-120000")
	start := steps[len(steps)-1].args
	joined := strings.Join(start, " ")
	if !strings.Contains(joined, "--name fep-fep-20261016-120000") {
		t.Fatalf("起作业那一步必须带上跨进程传目标用的固定别名：%q", joined)
	}
	// 别名只挂在注册了 --name 的**起会话**命令上（契约 C3.5）；其余几步今天不拼它
	for _, st := range steps[:len(steps)-1] {
		if strings.Contains(strings.Join(st.args, " "), "--name") {
			t.Fatalf("--name 只该挂在起会话那一步：%v", st.args)
		}
	}
	// 同一个录制身份算出的别名必须**同一个**（否则"重发同一条命令"会另起一条会话，
	// Q23 的幂等键就失效了）
	if recSessionAlias("fep-20261016-120000") != recSessionAlias("fep-20261016-120000") {
		t.Fatal("同一个录制 ID 必须算出同一个别名（别名是幂等键，Q23）")
	}
	if recSessionAlias("") != "" {
		t.Fatal("没有录制 ID 时不该编一个别名出来")
	}
	if a := recSessionAlias("fep-20261016-120000"); strings.ContainsAny(a, `/#`) {
		t.Fatalf("别名不许含 / 与 #（契约 C2.5 / Q17）：%q", a)
	}
}

// 没给 ENT 时不许下发 TOPENT：那是"沿用登录 shell 自己的"，下发个空值会把作业挪走。
func TestRecStartStepsSkipTopentWithoutEnt(t *testing.T) {
	steps := recStartSteps("主机109", "", "bsft001_wf", "", "10.0.0.1:1200", "")
	if len(steps) != 3 {
		t.Fatalf("没给 ENT 就是 收口 → 环境 → 起作业 三步：%v", steps)
	}
	for _, st := range steps {
		if strings.Contains(strings.Join(st.args, " "), "topent") {
			t.Fatalf("没给 ENT 不许出现 topent：%v", st.args)
		}
	}
	// 模块为空时不传 -m（让调试线用它自己的默认值），但 --fglserver 必须在
	start := strings.Join(steps[2].args, " ")
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
}
