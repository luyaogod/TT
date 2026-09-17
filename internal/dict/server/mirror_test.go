package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// GET/PUT /api/mirror:镜像根读写 + 环境路径/基线状态。
func TestMirrorGetPut(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"activeEnv":"a","sshs":[{"name":"a","host":"1.2.3.4","port":22,"user":"u","password":"p","zone":"36"}]}}`)
	s := New(p)
	dir := t.TempDir()

	// 未设置镜像根
	rec := httptest.NewRecorder()
	s.hMirrorGet(rec, httptest.NewRequest("GET", "/api/mirror", nil))
	var resp mirrorResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.MirrorDir != "" || len(resp.Envs) != 1 || resp.Envs[0].Ready {
		t.Fatalf("初始 resp = %+v", resp)
	}
	if resp.ActiveEnv != "a" {
		t.Fatalf("activeEnv = %q, want a", resp.ActiveEnv)
	}
	if resp.Envs[0].Name != "a" || resp.Envs[0].Zone != "36" {
		t.Fatalf("环境信息不符: %+v", resp.Envs[0])
	}

	// 设置镜像根
	body, _ := json.Marshal(map[string]string{"dir": dir})
	rec2 := httptest.NewRecorder()
	s.hMirrorPut(rec2, httptest.NewRequest("PUT", "/api/mirror", bytes.NewBuffer(body)))
	if !bytes.Contains(rec2.Body.Bytes(), []byte(`"ok":true`)) {
		t.Fatalf("PUT 失败: %s", rec2.Body.String())
	}

	// 写基线标记后应显示 ready + 环境路径
	envDir := filepath.Join(dir, "a")
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, ".tdict-mirror.ok"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec3 := httptest.NewRecorder()
	s.hMirrorGet(rec3, httptest.NewRequest("GET", "/api/mirror", nil))
	var resp2 mirrorResp
	if err := json.Unmarshal(rec3.Body.Bytes(), &resp2); err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(resp2.MirrorDir) != filepath.Clean(dir) || !resp2.Envs[0].Ready {
		t.Fatalf("设置后 resp = %+v", resp2)
	}
	if filepath.Clean(resp2.Envs[0].Path) != filepath.Clean(envDir) {
		t.Fatalf("path = %q, want %q", resp2.Envs[0].Path, envDir)
	}
}

// 拉取前置校验:未设镜像根 -> ok:false;空环境 -> 400。
func TestMirrorPullValidation(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"1.2.3.4","user":"u","password":"p"}]}}`)
	s := New(p)

	rec := httptest.NewRecorder()
	s.hMirrorPull(rec, httptest.NewRequest("POST", "/api/mirror/pull", bytes.NewBufferString(`{"env":"a"}`)))
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"ok":false`)) {
		t.Fatalf("未设镜像根应拒绝: %s", rec.Body.String())
	}

	rec2 := httptest.NewRecorder()
	s.hMirrorPull(rec2, httptest.NewRequest("POST", "/api/mirror/pull", bytes.NewBufferString(`{"env":""}`)))
	if rec2.Code != 400 {
		t.Fatalf("空环境应 400, got %d", rec2.Code)
	}
}

// 镜像根为空时 PUT 拒绝。
func TestMirrorPutEmptyDir(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"1.2.3.4"}]}}`)
	s := New(p)
	rec := httptest.NewRecorder()
	s.hMirrorPut(rec, httptest.NewRequest("PUT", "/api/mirror", bytes.NewBufferString(`{"dir":"  "}`)))
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"ok":false`)) {
		t.Fatalf("空目录应拒绝: %s", rec.Body.String())
	}
}
