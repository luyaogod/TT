package screenstart

// 判据：**按什么参数起子进程**。
//
// 真起一个进程需要 GDC 与真实环境，单测跑不了；而这里最容易错的恰恰是参数：
// 少了 `--session` 就变成"另起一条会话"（挂靠失效），少了 `--config` 子进程会去读默认配置。
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

	if _, logPath, err := New()("D:/cfg/config.json", "主机109#1", "主机109", "99", "bsft001_wf", "R2"); err != nil {
		t.Fatalf("起画面通道失败：%v", err)
	} else if !strings.HasSuffix(logPath, "screen.log") {
		t.Fatalf("日志落点不对：%s", logPath)
	}

	line := strings.Join(*got, " ")
	for _, want := range []string{
		"C:/fake/tt.exe", "recording start", "--session 主机109#1", "--prog bsft001_wf",
		"--ent 99", "--env 主机109", "--config D:/cfg/config.json",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("参数里缺 %q：%s", want, line)
		}
	}
}

// TestStartOmitsEmptyOptionals：空的 ent/env/config 不许拼成空值参数。
func TestStartOmitsEmptyOptionals(t *testing.T) {
	got := stub(t, func(string, []string, string, ...string) (int, error) { return 1, nil },
		func(int) bool { return true })
	if _, _, err := New()("", "主机109#1", "", "", "p", ""); err != nil {
		t.Fatal(err)
	}
	line := strings.Join(*got, " ")
	for _, bad := range []string{"--ent", "--env", "--config"} {
		if strings.Contains(line, bad) {
			t.Errorf("空值不该拼 %q：%s", bad, line)
		}
	}
}

// TestStartReportsImmediateExit：起来就退 = 启动失败，错误里要带日志路径。
func TestStartReportsImmediateExit(t *testing.T) {
	stub(t, func(string, []string, string, ...string) (int, error) { return 7, nil },
		func(int) bool { return false })
	_, logPath, err := New()("", "r", "e", "", "p", "")
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
	_, _, err := New()("", "r", "e", "", "p", "")
	if err == nil || !strings.Contains(err.Error(), "Access is denied") {
		t.Fatalf("起不来要如实报：%v", err)
	}
}

// TestStartReportsExecutableError：拿不到自身路径也要报，而不是起一个空路径。
func TestStartReportsExecutableError(t *testing.T) {
	stub(t, func(string, []string, string, ...string) (int, error) { return 1, nil }, func(int) bool { return true })
	executable = func() (string, error) { return "", errors.New("no exe") }
	if _, _, err := New()("", "r", "e", "", "p", ""); err == nil {
		t.Fatal("拿不到 exe 路径该报错")
	}
}
