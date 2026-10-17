package fep

// 作业连接（P4）的判据：边界在 job.go，且空作业收尾安全。
//
// 边界那条和画面通道同一条道理：别处再冒出"自己拨 SSH / 开 PTY"的地方，就会有第二个
// 作业连接 —— 谁关它、它的 PTY 归谁、编译件标识记在哪，全都没人管了。

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestJobIsSafeWhenEmpty：没打开的 Job 收尾不炸（收尾路径必须幂等）。
func TestJobIsSafeWhenEmpty(t *testing.T) {
	var j *Job
	if err := j.Close(); err != nil {
		t.Fatalf("nil Job 收尾不该报错：%v", err)
	}
	if j.Binary().Ref != "" {
		t.Fatal("空 Job 不该有编译件标识")
	}
	if j.Done() != nil {
		// Done 在未打开时也返回一个 channel（watch 关闭它）—— 这里只要求不为 nil。
		t.Log("空 Job 的 Done 有值，符合预期")
	}
}

// TestNewJobStartsUnopened：新建的 Job 没有连接、done 未关闭。
func TestNewJobStartsUnopened(t *testing.T) {
	j := newJob(OpenConfig{}, nopLog)
	if j == nil {
		t.Fatal("newJob 返回 nil")
	}
	if j.Binary().Ref != "" {
		t.Fatal("还没探编译件就不该有值")
	}
	select {
	case <-j.Done():
		t.Fatal("还没打开就结束了？")
	default:
	}
	if err := j.Close(); err != nil {
		t.Fatalf("收一个没打开的 Job 不该报错：%v", err)
	}
}

// TestOnlyJobConnectsToTheHost：拨 SSH / 开 PTY 只许在 job.go 里发生。
func TestOnlyJobConnectsToTheHost(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("读包目录失败：%v", err)
	}
	forbidden := []string{"host.Dial(", ".NewPTY("}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if name == "job.go" {
			continue
		}
		b, rerr := os.ReadFile(filepath.Clean(name))
		if rerr != nil {
			t.Fatalf("读 %s 失败：%v", name, rerr)
		}
		for _, line := range strings.Split(string(b), "\n") {
			code := strings.TrimSpace(line)
			if strings.HasPrefix(code, "//") {
				continue
			}
			for _, f := range forbidden {
				if strings.Contains(code, f) {
					t.Errorf("%s 里出现了 %q —— 拨连接 / 开 PTY 只许在 job.go 里（P4 的边界）", name, f)
				}
			}
		}
	}
}
