package fep

import (
	"errors"
	"strings"
	"testing"
)

// R6 的验收：`--prog` / `--fgserver-ip` / 区域代码里的注入尝试必须在**拼出远端命令之前**被拒。
//
// 依据是仓库里已经修过的那一处（internal/host 的 ChenvCmd）：同一类洞的实测症状是
// `--zone "36; id"` 能在生产服务器上执行任意命令 —— 而这些值现在由前端面板提供，
// 面板一接入它就是活的风险。

func TestValidateProgRejectsShellMetacharacters(t *testing.T) {
	bad := []string{
		`axmt500_wf; id`,
		`axmt500_wf && rm -rf /`,
		`axmt500_wf|id`,
		"axmt500_wf\nid",
		"axmt500_wf`id`",
		`axmt500_wf$(id)`,
		`../axmt500_wf`,
		`42r`,
		``,
		`axmt500_wf'`,
		strings.Repeat("a", 65),
	}
	for _, p := range bad {
		if err := ValidateProg(p); err == nil {
			t.Errorf("作业名 %q 必须被拒", p)
		} else if !errors.Is(err, ErrUsage) {
			t.Errorf("作业名 %q 的错要能判成用法错: %v", p, err)
		}
	}
	good := []string{"axmt500_wf", "bsft001_wf", "a", "a_b9", strings.Repeat("a", 64)}
	for _, p := range good {
		if err := ValidateProg(p); err != nil {
			t.Errorf("作业名 %q 应放行: %v", p, err)
		}
	}
}

func TestValidateFGLServerIPOnlyAcceptsLiterals(t *testing.T) {
	good := []string{"", "192.168.14.193", "10.0.0.1:1300", "::1", "[::1]:1300"}
	for _, ip := range good {
		if err := ValidateFGLServerIP(ip); err != nil {
			t.Errorf("%q 应放行: %v", ip, err)
		}
	}
	bad := []string{
		"1.2.3.4; id",
		"1.2.3.4:1300; id",
		"example.com",
		"localhost:1300",
		"192.168.14.193 1300",
		"`id`",
		"$HOST",
	}
	for _, ip := range bad {
		if err := ValidateFGLServerIP(ip); err == nil {
			t.Errorf("%q 必须被拒", ip)
		}
	}
}

// 作业名与地址是**裸插**进这条命令的（r.r <prog> / FGLSERVER=<ip>:<port>），
// 所以拒绝必须发生在 jobLaunchCmd 自己里面 —— 调用方忘了查也发不出去。
func TestJobLaunchCmdRefusesInjection(t *testing.T) {
	if got, err := jobLaunchCmd(PathRunnerRR, "10.0.0.1", 1200, "x; id", ""); err == nil {
		t.Fatalf("带分号的作业名必须被拒，却拼出了 %q", got)
	}
	if got, err := jobLaunchCmd(PathRunnerRR, "1.2.3.4; id", 1200, "axmt500_wf", ""); err == nil {
		t.Fatalf("带分号的地址必须被拒，却拼出了 %q", got)
	}
}

// 入口处查：这些值里有一个不合法，整轮就不能起（不能"先起了再说"——
// 隔离实例与代理一旦起来，命令就已经在路上了）。
func TestValidateLaunchCoversEveryRemoteValue(t *testing.T) {
	if err := validateLaunch(OpenConfig{Prog: "axmt500_wf", Zone: "36", FGLServerIP: "10.0.0.1"}); err != nil {
		t.Fatalf("全合法应放行: %v", err)
	}
	// Zone 会被直接写进登录 PTY（openJob），所以它也在白名单里
	if err := validateLaunch(OpenConfig{Prog: "axmt500_wf", Zone: "36; id"}); err == nil {
		t.Fatal("区域代码里的注入必须被拒")
	}
	if err := validateLaunch(OpenConfig{Prog: "x; id", Zone: "36"}); err == nil {
		t.Fatal("作业名里的注入必须被拒")
	}
	if err := validateLaunch(OpenConfig{Prog: "axmt500_wf", Zone: "36", FGLServerIP: "a b"}); err == nil {
		t.Fatal("地址里的注入必须被拒")
	}
	// NoJob：不起作业，所以允许没有作业名（--no-job 是排障用的一条正路）
	if err := validateLaunch(OpenConfig{NoJob: true, Zone: "36"}); err != nil {
		t.Fatalf("--no-job 不该要求作业名: %v", err)
	}
}
