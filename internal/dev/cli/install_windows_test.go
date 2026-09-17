//go:build windows

package cli

import (
	"strings"
	"testing"
)

// 本文件只放**依赖 Windows 注册表**的用例（reg query 解析、真实 PATH 读取）。
// 非 Windows 上这些符号不存在（见 install_other.go），所以用构建标签隔开，
// 保证 `GOOS=linux go vet ./...` / 跨平台编译同样干净。
// TestParseRegQueryPath 解析 reg query 输出：含**长值折行**（本机真实形态）。
func TestParseRegQueryPath(t *testing.T) {
	// 真实布局：值按控制台宽度折行，续行从行首继续，不插空格
	out := "\r\nHKEY_CURRENT_USER\\Environment\r\n" +
		"    Path    REG_EXPAND_SZ    C:\\Users\\x\\.cargo\\bin;C:\\Users\\x\\AppData\\Local\\Programs\\\r\n" +
		"Python311\\Scripts\\;%USERPROFILE%\\AppData\\Local\\Microsoft\\WindowsApps;D:\\APPS\\tdict-portable\r\n\r\n"
	val, typ, err := parseRegQueryPath(out)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if typ != "REG_EXPAND_SZ" {
		t.Errorf("类型 = %q，想要 REG_EXPAND_SZ", typ)
	}
	want := `C:\Users\x\.cargo\bin;C:\Users\x\AppData\Local\Programs\Python311\Scripts\;%USERPROFILE%\AppData\Local\Microsoft\WindowsApps;D:\APPS\tdict-portable`
	if val != want {
		t.Errorf("值 = %q\n想要 %q", val, want)
	}
	if !strings.Contains(val, "%USERPROFILE%") {
		t.Error("展开式变量必须原样保留（不能被展开）")
	}

	// REG_SZ 空值
	val2, typ2, err := parseRegQueryPath("HKEY_CURRENT_USER\\Environment\r\n    Path    REG_SZ    \r\n")
	if err != nil {
		t.Fatalf("空值解析失败: %v", err)
	}
	if val2 != "" || typ2 != "REG_SZ" {
		t.Errorf("空值 = (%q, %q)，想要 (\"\", \"REG_SZ\")", val2, typ2)
	}

	// 找不到值 → 报错（值不存在由调用方当成空处理）
	if _, _, err := parseRegQueryPath("HKEY_CURRENT_USER\\Environment\r\n"); err == nil {
		t.Error("输出里没有 Path 值时应报错")
	}
}

// TestCmdInstallPathDryRun --dry-run 绝不写注册表：跑得通、有输出、退出码 0。
func TestCmdInstallPathDryRun(t *testing.T) {
	if code := silent(t, func() int { return cmdInstall([]string{"path", "--dry-run"}) }); code != 0 {
		t.Errorf("install path --dry-run 退出码 %d，想要 0", code)
	}
}
