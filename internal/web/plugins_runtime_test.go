package web

// 插件运行态与日志端点（P6-07）：都是**只读增量**，不改变既有响应形状。

import (
	"path/filepath"
	"testing"

	"github.com/luyaogod/TT/internal/plugin"
)

// TestPluginListCarriesRuntimeStatus 列表里附带 supervisor 状态（state/addr/pid/restarts）。
func TestPluginListCarriesRuntimeStatus(t *testing.T) {
	s, m := pluginTestServer(t)
	if _, err := m.Install(plugin.InstallOptions{Source: writePluginFixture(t)}); err != nil {
		t.Fatal(err)
	}
	// 没有运行态来源时：字段不出现（老客户端看到的形状不变）。
	rec, body := doJSON(t, s, "GET", "/api/plugins", nil)
	if rec.Code != 200 {
		t.Fatalf("列表该成功，得 %d", rec.Code)
	}
	first := body["plugins"].([]any)[0].(map[string]any)
	if _, has := first["state"]; has {
		t.Errorf("没装配 supervisor 时不该有 state: %v", first)
	}

	// 装配运行态来源后：字段出现且值来自它。
	s2 := New(Options{
		Version: "test", ConfigPath: "", PluginMgr: m,
		PluginStatuses: func() []plugin.Status {
			return []plugin.Status{{
				PluginID: "com.example.demo", State: "running", Addr: "127.0.0.1:12345",
				PID: 4242, Restarts: 1,
			}}
		},
	})
	rec, body = doJSON(t, s2, "GET", "/api/plugins", nil)
	if rec.Code != 200 {
		t.Fatalf("列表该成功，得 %d", rec.Code)
	}
	first = body["plugins"].([]any)[0].(map[string]any)
	for k, want := range map[string]any{
		"state": "running", "addr": "127.0.0.1:12345",
		"pid": float64(4242), "restarts": float64(1),
	} {
		if got := first[k]; got != want {
			t.Errorf("%s 该是 %v，得 %v", k, want, got)
		}
	}
}

// TestPluginLogsEndpoint 只读日志视图：最新文件的最后 N 行，未见过的插件 404。
func TestPluginLogsEndpoint(t *testing.T) {
	s, m := pluginTestServer(t)
	if _, err := m.Install(plugin.InstallOptions{Source: writePluginFixture(t)}); err != nil {
		t.Fatal(err)
	}
	logDir := m.Store.LogsDir("com.example.demo")
	writeTestFile(t, filepath.Join(logDir, "service-20260101-000000.log"), "第一行\n第二行\n第三行\n")

	rec, body := doJSON(t, s, "GET", "/api/plugins/com.example.demo/logs?tail=2", nil)
	if rec.Code != 200 {
		t.Fatalf("日志端点该成功，得 %d（%v）", rec.Code, body)
	}
	if body["ok"] != true || body["id"] != "com.example.demo" {
		t.Fatalf("响应不对: %v", body)
	}
	lines, _ := body["lines"].([]any)
	if len(lines) != 2 || lines[1] != "第三行" {
		t.Fatalf("该回最后两行，得 %v", body["lines"])
	}
	if body["file"] == "" || body["file"] == nil {
		t.Error("该给出实际读的文件")
	}

	// tail 不合法 → 400（不是静默回 50 行）。
	if rec, _ := doJSON(t, s, "GET", "/api/plugins/com.example.demo/logs?tail=abc", nil); rec.Code != 400 {
		t.Errorf("非法 tail 该 400，得 %d", rec.Code)
	}
	// 没有这个插件 → 404（而不是"没有日志"）。
	if rec, _ := doJSON(t, s, "GET", "/api/plugins/nope/logs", nil); rec.Code != 404 {
		t.Errorf("不存在的插件该 404，得 %d", rec.Code)
	}
}

// TestPluginLogsNotWiredWhenNoManager 没有管理面时日志端点也不存在（503/405/404 都可，不能 200）。
func TestPluginLogsNotWiredWhenNoManager(t *testing.T) {
	s := New(Options{Version: "test"})
	rec, _ := doJSON(t, s, "GET", "/api/plugins/com.example.demo/logs", nil)
	if rec.Code == 200 {
		t.Fatal("没装配插件管理面时不该有日志端点")
	}
}
