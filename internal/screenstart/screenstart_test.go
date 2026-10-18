package screenstart

// 判据：**按什么参数、以什么身份起子进程**。
//
// 真起一个进程需要 GDC 与真实环境，单测跑不了；而这里最容易错的恰恰是参数：
// 少了 `--session` 就变成"另起一条会话"（挂靠失效），少了 `--config` 子进程会去读默认配置，
// 少了 `TT_ACTOR` 子进程会以 AI 自称、被人建的纯人工会话挡在门外（实测 2026-10-18）。
// 三种结局也各钉一条：起不来 / 起来就退 / 起住了。

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// stub 把三个可替换点换成假的，测试结束自动还原。
func stub(t *testing.T, spawn func(exe string, args []string, log string, env ...string) (int, error), alive func(int) bool) *[]string {
	t.Helper()
	var got []string
	oldExec, oldSpawn, oldAlive := executable, spawnDetached, processAlive
	oldPause, oldDeadline := alivePollPause, aliveDeadline
	t.Cleanup(func() {
		executable, spawnDetached, processAlive = oldExec, oldSpawn, oldAlive
		alivePollPause, aliveDeadline = oldPause, oldDeadline
	})
	executable = func() (string, error) { return "C:/fake/tt.exe", nil }
	spawnDetached = func(exe string, args []string, log string, env ...string) (int, error) {
		got = append([]string{exe}, args...)
		got = append(got, "ENV:"+strings.Join(env, " "))
		return spawn(exe, args, log, env...)
	}
	processAlive = alive
	alivePollPause = time.Millisecond
	aliveDeadline = 5 * time.Millisecond
	return &got
}

// TestStartPassesBindArgs：挂靠模式必须带 `--session`/`--prog`，可选项按需补。
func TestStartPassesBindArgs(t *testing.T) {
	got := stub(t, func(string, []string, string, ...string) (int, error) { return 4242, nil },
		func(int) bool { return true })

	req := Request{
		CfgPath: "D:/cfg/config.json", Ref: "主机109#1", Env: "主机109",
		Ent: "99", Prog: "bsft001_wf", Tag: "R2", Actor: "human",
	}
	if _, logPath, err := New()(req); err != nil {
		t.Fatalf("起画面通道失败：%v", err)
	} else if !strings.HasSuffix(logPath, "screen.log") {
		t.Fatalf("日志落点不对：%s", logPath)
	}

	line := strings.Join(*got, " ")
	for _, want := range []string{
		"C:/fake/tt.exe", "recording start", "--session 主机109#1", "--prog bsft001_wf",
		"--ent 99", "--env 主机109", "--config D:/cfg/config.json",
		// 身份必须传下去：否则子进程自称 AI，人建的纯人工会话会把它挡在门外。
		"ENV:TT_ACTOR=human",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("参数里缺 %q：%s", want, line)
		}
	}
}

// TestStartOmitsEmptyOptionals：空的 ent/env/config/actor 不许拼成空值参数/空环境变量。
func TestStartOmitsEmptyOptionals(t *testing.T) {
	got := stub(t, func(string, []string, string, ...string) (int, error) { return 1, nil },
		func(int) bool { return true })
	if _, _, err := New()(Request{Ref: "主机109#1", Prog: "p"}); err != nil {
		t.Fatal(err)
	}
	line := strings.Join(*got, " ")
	for _, bad := range []string{"--ent", "--env", "--config"} {
		if strings.Contains(line, bad) {
			t.Errorf("空值不该拼 %q：%s", bad, line)
		}
	}
	// 没给身份 = 不设 TT_ACTOR：子进程按 CLI 的缺省算 ai（既有行为一字不改）。
	if strings.Contains(line, "TT_ACTOR=") {
		t.Errorf("没给身份时不该注入 TT_ACTOR：%s", line)
	}
}

// TestStartReportsImmediateExit：起来就退 = 启动失败，错误里要带日志路径。
func TestStartReportsImmediateExit(t *testing.T) {
	stub(t, func(string, []string, string, ...string) (int, error) { return 7, nil },
		func(int) bool { return false })
	_, logPath, err := New()(Request{Ref: "r", Env: "e", Prog: "p"})
	if err == nil || !strings.Contains(err.Error(), "起来就退") {
		t.Fatalf("该报'起来就退'：%v", err)
	}
	if !strings.Contains(err.Error(), logPath) {
		t.Errorf("错误里要带日志路径（%s）：%v", logPath, err)
	}
}

// TestStartReportsSpawnError：起不来照实报，不吞。
func TestStartReportsSpawnError(t *testing.T) {
	stub(t, func(string, []string, string, ...string) (int, error) { return 0, errors.New("Access is denied") },
		func(int) bool { return false })
	_, _, err := New()(Request{Ref: "r", Env: "e", Prog: "p"})
	if err == nil || !strings.Contains(err.Error(), "Access is denied") {
		t.Fatalf("起不来要如实报：%v", err)
	}
}

// TestStartReportsExecutableError：拿不到自身路径也要报，而不是起一个空路径。
func TestStartReportsExecutableError(t *testing.T) {
	stub(t, func(string, []string, string, ...string) (int, error) { return 1, nil }, func(int) bool { return true })
	executable = func() (string, error) { return "", errors.New("no exe") }
	if _, _, err := New()(Request{Ref: "r", Env: "e", Prog: "p"}); err == nil {
		t.Fatal("拿不到 exe 路径该报错")
	}
}
