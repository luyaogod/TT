package web

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// 环境被删掉时，它的**录制库要跟着收**（用户 2026-10-15 定的策略：
// 用户主动录的那些长期留着，但环境没了它们就是没人认领的字节）。
// 环境可能是设置页删的（这条），也可能是手改 config.json 删的
// （那条走 `tt fep rec prune` 兜底，见 internal/cli/fep）。
func TestHostsPutCascadesRecordingLibrary(t *testing.T) {
	seed := `{"schemaVersion":3,"hosts":{"sshs":[
	  {"name":"要删的环境","host":"10.0.0.9","port":22,"user":"u","password":"p","zone":"36","db":{"type":"oracle","host":"10.0.0.9","port":1521,"service":"t35prd","accounts":[{"account":"ds","password":"x"}]}},
	  {"name":"留着的环境","host":"10.0.0.8","port":22,"user":"u","password":"p","zone":"36","db":{"type":"oracle","host":"10.0.0.8","port":1521,"service":"t35prd","accounts":[{"account":"ds","password":"x"}]}}]}}`
	s, cfgPath := newTestServer(t, seed)
	dataDir := filepath.Dir(cfgPath)

	// 两个环境各有一份录制（用户主动录的：住在 records/<环境>/<ENT>/<ID>/）
	seedRecordingDir := func(env, id string) string {
		dir := filepath.Join(dataDir, "records", env, "99", id)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, id+"-1-rt2gdc.log"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	doomed := seedRecordingDir("要删的环境", "fep-doomed")
	kept := seedRecordingDir("留着的环境", "fep-kept")

	// 设置页保存：只留一个环境（等价于把另一个删掉）
	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
		"hosts": map[string]any{
			"sshs": []any{map[string]any{
				"name": "留着的环境", "host": "10.0.0.8", "port": 22, "user": "u", "password": "p", "zone": "36",
				"db": map[string]any{"type": "oracle", "host": "10.0.0.8", "port": 1521, "service": "t35prd",
					"accounts": []any{map[string]any{"account": "ds", "password": "x"}}},
			}},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("保存环境失败：%d %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(doomed); !os.IsNotExist(err) {
		t.Fatalf("被删环境的录制库该跟着删掉：%v", err)
	}
	if _, err := os.Stat(kept); err != nil {
		t.Fatalf("没被删的环境，它的录制不该动：%v", err)
	}
	// 库里那个环境的整段都没了（不是只删了一个录制目录）
	if _, err := os.Stat(filepath.Dir(doomed)); !os.IsNotExist(err) {
		t.Fatalf("该环境在库里整段都不该留：%v", err)
	}
}

// 改配置但**没删环境**（只改 host）时，一份录制都不许动。
func TestHostsPutKeepsRecordingsWhenNothingRemoved(t *testing.T) {
	seed := `{"schemaVersion":3,"hosts":{"sshs":[
	  {"name":"甲","host":"10.0.0.9","port":22,"user":"u","password":"p","zone":"36","db":{"type":"oracle","host":"10.0.0.9","port":1521,"service":"t35prd","accounts":[{"account":"ds","password":"x"}]}}]}}`
	s, cfgPath := newTestServer(t, seed)
	dir := filepath.Join(filepath.Dir(cfgPath), "records", "甲", "99", "fep-x")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	rec, _ := doJSON(t, s, http.MethodPut, "/api/hosts", map[string]any{
		"hosts": map[string]any{
			"sshs": []any{map[string]any{
				"name": "甲", "host": "10.0.0.10", "port": 22, "user": "u", "password": "p", "zone": "36",
				"db": map[string]any{"type": "oracle", "host": "10.0.0.10", "port": 1521, "service": "t35prd",
					"accounts": []any{map[string]any{"account": "ds", "password": "x"}}},
			}},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("保存失败：%d", rec.Code)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("只是改了 host，录制不该动：%v", err)
	}
}

func TestLostNames(t *testing.T) {
	got := lostNames([]string{"甲", "乙", "丙"}, []string{"乙", "丁"})
	if len(got) != 2 || got[0] != "甲" || got[1] != "丙" {
		t.Fatalf("该报出 甲/丙：%v", got)
	}
	if n := lostNames(nil, []string{"甲"}); len(n) != 0 {
		t.Fatalf("旧的为空时没有丢失的名字：%v", n)
	}
	if n := lostNames([]string{"甲"}, nil); len(n) != 1 {
		t.Fatalf("全删时每个都算丢失：%v", n)
	}
}

func TestEnvNamesInSkipsNameless(t *testing.T) {
	sec := map[string]any{"sshs": []any{
		map[string]any{"name": "甲"},
		map[string]any{"host": "无名"},
		map[string]any{"name": ""},
		"不是对象",
	}}
	got := envNamesIn(sec)
	if len(got) != 1 || got[0] != "甲" {
		t.Fatalf("只该取到有名字的那个：%v", got)
	}
}
