package fep

import (
	"strings"
	"testing"

	"tt/internal/config"
	"tt/internal/fep"
)

// 不给 --env 时必须报错并**列出可用环境**。
//
// 这条钉的是一个真实踩过的症状：以前这里用 debug.Config.EnvName() 兜底，
// 而它最后一级是 `SSH.Host + "-" + Zone`；环境没选中时 SSH 是零值，
// 于是拼出一个叫 `-` 的幽灵环境名，报错成了「环境 "-" 不存在（见 tt env list）」——
// 人看不出该补什么，也不知道有哪些环境可选。
func TestOpenEnvNameRequiresExplicitEnv(t *testing.T) {
	envs := []config.NamedSsh{{Name: "主机109"}, {Name: "尚阳通"}}

	_, err := openEnvName(envs, "")
	if err == nil {
		t.Fatal("不给环境名必须报错（环境一律点名，没有默认环境可回落）")
	}
	if !strings.Contains(err.Error(), "必须指定环境") {
		t.Fatalf("报错要告诉人怎么补（须含「必须指定环境」），实际: %v", err)
	}
	for _, n := range []string{"主机109", "尚阳通"} {
		if !strings.Contains(err.Error(), n) {
			t.Fatalf("报错要列出可用环境，缺 %q: %v", n, err)
		}
	}
	if strings.Contains(err.Error(), `"-"`) {
		t.Fatalf("不许再出现幽灵环境名 -: %v", err)
	}
}

// 点名了就必须按那个名字解出来；拼错必须报错，绝不静默换一个环境去连
// （连错环境的表现是"查不到数据"，与"数据不存在"一样）。
func TestOpenEnvNameResolvesByNameAndRejectsTypos(t *testing.T) {
	envs := []config.NamedSsh{{Name: "主机109"}, {Name: "尚阳通"}}

	got, err := openEnvName(envs, "主机109")
	if err != nil {
		t.Fatalf("按名解析应成功: %v", err)
	}
	if got != "主机109" {
		t.Fatalf("got %q, want 主机109", got)
	}

	_, err = openEnvName(envs, "主机110")
	if err == nil {
		t.Fatal("环境名拼错必须报错")
	}
	if !strings.Contains(err.Error(), "主机109") {
		t.Fatalf("报错要列出可用环境: %v", err)
	}
}

// 一个环境都没配时也要有话说（设置页可以被删空）。
func TestOpenEnvNameWithoutAnyEnv(t *testing.T) {
	_, err := openEnvName(nil, "")
	if err == nil {
		t.Fatal("没有任何环境时必须报错")
	}
	if !strings.Contains(err.Error(), "环境") {
		t.Fatalf("报错应提到环境: %v", err)
	}
}

// --ent 必须挂在 session open 上，且取值来源只有一个（openEnt）。
// 这条防的是"旗标加了但没接进 OpenConfig"——
// 那种情况下命令照常成功，作业却跑在别的企业上，全程没有一个字提示。
func TestSessionOpenHasEntFlag(t *testing.T) {
	if f := sessionOpenCmd.Flags().Lookup("ent"); f == nil {
		t.Fatal("session open 必须有 --ent（企业编号由用户每次显式给）")
	}
	if f := sessionOpenCmd.Flags().Lookup("env"); f == nil {
		t.Fatal("session open 必须有 --env（环境一律点名）")
	} else if strings.Contains(f.Usage, "activeEnv") {
		t.Fatalf("--env 的帮助还写着已删除的 activeEnv: %q", f.Usage)
	}
}

// entText 没给 ENT 时不许留空：留空读起来像"这个会话没有企业"，
// 而"我没指定（沿用登录 shell）"与"没有"是两件事。
func TestEntTextSaysWhatHappensWithoutEnt(t *testing.T) {
	if got := entText("99"); got != "ENT 99" {
		t.Fatalf("entText(99) = %q", got)
	}
	got := entText("")
	if got == "" || !strings.Contains(got, "TOPENT") {
		t.Fatalf("没给 ENT 时要如实说沿用登录 shell: %q", got)
	}
}

// 启动路径要**跟着会话一起记下来**（录制身份与回放闸门都靠它）：
// 这条命令只走 T100 标准启动器（r.r），--no-job 时如实报 none。
// T100 原生的 r.d（fgldeb）不在命令面上 —— 要录要调都走调试线。
func TestOpenPathReflectsHowTheJobIsStarted(t *testing.T) {
	_ = openNoJob
	if got := (openParams{}).path(); got != fep.PathRunnerRR {
		t.Fatalf("session open 只走标准启动器，got %q", got)
	}

	if got := (openParams{NoJob: true}).path(); got != fep.PathNone {
		t.Fatalf("--no-job 没起作业，要如实报 none，got %q", got)
	}

	// 命令面上不许再有 --path：r.d 这条路线已拿掉（它的常量只为识别老录制件而留）
	if f := sessionOpenCmd.Flags().Lookup("path"); f != nil {
		t.Fatal("--path 已从命令面拿掉：录制与调试都走调试线")
	}
}

// --binary-ref 必须挂着：录制身份里的"编译件标识"只能由调用方给（r.r 路径下
// 程序路径由登录环境决定，我们不知道）。
func TestSessionOpenHasBinaryRefFlag(t *testing.T) {
	f := sessionOpenCmd.Flags().Lookup("binary-ref")
	if f == nil {
		t.Fatal("session open 必须有 --binary-ref（否则身份里的编译件一栏永远是空的）")
	}
	if !strings.Contains(f.Usage, "重编译") {
		t.Fatalf("帮助要写清它为什么存在: %q", f.Usage)
	}
	if err := f.Value.Set("/u1/t35prd/erp/asf/42r/bsft001_wf.42r"); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(openBinary) != "/u1/t35prd/erp/asf/42r/bsft001_wf.42r" {
		t.Fatalf("旗标没接进 openBinary: %q", openBinary)
	}
	_ = f.Value.Set("")
}
