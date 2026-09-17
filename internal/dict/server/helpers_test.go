package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeTempConfig 在临时目录写一份 config.json,返回其路径。
func writeTempConfig(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("写入临时配置失败: %v", err)
	}
	return p
}

// readRoot 把落盘的 config.json 读成 map(测试断言顶层键用)。
func readRoot(t *testing.T, p string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(b, &root); err != nil {
		t.Fatal(err)
	}
	return root
}
