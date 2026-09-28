package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tt config migrate --dry-run 必须打码 —— 它是唯一会把整份配置内容吐出来的迁移入口。
//
// plan.Merged 里有 hosts.sshs[].password 与 db.accounts[].password，全是明文；修之前这一条
// 直接把它们原样打到终端，而同一份数据在 config show / config get 里是打码的
// （AGENTS.md §9：输出侧打码只有一个实现，新增"把配置吐出来"的命令要走它）。
//
// 这条断言的对象是"输出里不能出现夹具里那个假口令"，不是"输出长得像什么样" ——
// 打码的占位符以后换了也不该让这条测试变红。
func TestConfigMigrateDryRunRedactsPasswords(t *testing.T) {
	appdata := t.TempDir()
	t.Setenv("APPDATA", appdata)
	t.Setenv("TT_HOME", t.TempDir()) // 新落点：空的，于是迁移真的有话可说
	t.Setenv("T100_HOME", "")
	t.Setenv("TT_CONFIG", "")
	t.Setenv("TDEBUG_CONFIG", "")
	t.Setenv("TDICT_CONFIG", "")

	// 一份合并前的旧配置，口令是仓库里惯用的假口令（AGENTS.md §9：夹具一律用假口令）
	dir := filepath.Join(appdata, "T100", "tdebug")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	seed := `{"debug":{"sshs":[{"name":"e","host":"h","user":"u","password":"SECRET"}]}}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}

	// 用 SetOut 而不是 testkit.CaptureStdout：cobra 的 Print/Printf/Println 走的是
	// OutOrStderr（不是 OutOrStdout），换掉 os.Stdout 抓不到它。
	cmd := newConfigCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"migrate", "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("dry-run 失败: %v", err)
	}
	out := buf.String()

	if strings.Contains(out, "SECRET") {
		t.Errorf("--dry-run 把明文口令打出来了:\n%s", out)
	}
	// 反向断言：得确认它是**打着码**吐出来的，而不是"什么都没吐"
	if !strings.Contains(out, "password") {
		t.Errorf("输出里连 password 这个键都没有，这条测试失去了对象:\n%s", out)
	}
}
