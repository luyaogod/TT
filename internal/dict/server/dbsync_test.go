package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// GET /api/dbsync:目标库、默认环境、只列挂了库的环境。
func TestDBSyncGet(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"activeEnv":"b","sshs":[
      {"name":"a","host":"1.2.3.4","user":"u","password":"p"},
      {"name":"b","host":"5.6.7.8","user":"u","password":"p",
       "db":{"type":"oracle","host":"5.6.7.8","port":1521,"service":"s","accounts":[{"account":"x","password":"x"}]}}]}}`)
	s := New(p)
	s.SetDBTarget("D:/tmp/erp_data.db")

	rec := httptest.NewRecorder()
	s.hDBSyncGet(rec, httptest.NewRequest("GET", "/api/dbsync", nil))
	var resp dbSyncResp
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Target != "D:/tmp/erp_data.db" {
		t.Fatalf("target = %q", resp.Target)
	}
	if resp.ActiveEnv != "b" {
		t.Fatalf("activeEnv = %q, want b", resp.ActiveEnv)
	}
	if len(resp.Envs) != 1 || resp.Envs[0].Name != "b" || resp.Envs[0].Type != "oracle" {
		t.Fatalf("应只列挂库环境: %+v", resp.Envs)
	}
	if resp.Job.Running {
		t.Fatal("初始不应有任务在跑")
	}
}

// 拉取前置校验:空环境/未知环境/未挂库均 400。
func TestDBSyncPostValidation(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"nodb","host":"1.2.3.4","user":"u","password":"p"}]}}`)
	s := New(p)

	for _, body := range []string{`{"env":""}`, `{"env":"nope"}`, `{"env":"nodb"}`} {
		rec := httptest.NewRecorder()
		s.hDBSyncPost(rec, httptest.NewRequest("POST", "/api/dbsync", bytes.NewBufferString(body)))
		if rec.Code != 400 {
			t.Fatalf("body %s 应 400, got %d (%s)", body, rec.Code, rec.Body.String())
		}
	}
}

// 已有任务在跑时再启动返回 409(不触发真实同步)。
func TestDBSyncRunningConflict(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"1.2.3.4","user":"u","password":"p",
      "db":{"type":"oracle","host":"1.2.3.4","service":"s","accounts":[{"account":"x","password":"x"}]}}]}}`)
	s := New(p)
	s.sync.Running = true

	rec := httptest.NewRecorder()
	s.hDBSyncPost(rec, httptest.NewRequest("POST", "/api/dbsync", bytes.NewBufferString(`{"env":"a"}`)))
	if rec.Code != 409 {
		t.Fatalf("已有任务应 409, got %d", rec.Code)
	}
}

// 目标库可设置:PUT 写 config.json 顶层 sync.target;空值清除、回到默认(exe 同目录)。
func TestDBSyncTargetSetting(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"1.2.3.4","user":"u","password":"p",
      "db":{"type":"oracle","host":"1.2.3.4","service":"s","accounts":[{"account":"x","password":"x"}]}}]}}`)
	s := New(p)
	s.SetDBTarget(`D:\default-dir\erp_data.db`)

	rec := httptest.NewRecorder()
	s.hDBSyncGet(rec, httptest.NewRequest("GET", "/api/dbsync", nil))
	var r1 dbSyncResp
	if err := json.Unmarshal(rec.Body.Bytes(), &r1); err != nil {
		t.Fatal(err)
	}
	if r1.Configured != "" || r1.Target != `D:\default-dir\erp_data.db` || r1.DefaultTarget != `D:\default-dir\erp_data.db` {
		t.Fatalf("默认状态不符: %+v", r1)
	}

	custom := filepath.Join(t.TempDir(), "sub", "erp.db")
	body, _ := json.Marshal(map[string]string{"target": custom})
	rec2 := httptest.NewRecorder()
	s.hDBSyncPut(rec2, httptest.NewRequest("PUT", "/api/dbsync", bytes.NewBuffer(body)))
	var r2 dbSyncResp
	if err := json.Unmarshal(rec2.Body.Bytes(), &r2); err != nil {
		t.Fatal(err)
	}
	if filepath.Clean(r2.Configured) != filepath.Clean(custom) || filepath.Clean(r2.Target) != filepath.Clean(custom) {
		t.Fatalf("设置后不符: %+v", r2)
	}
	if r2.Exists {
		t.Fatal("目标文件尚不存在,exists 应为 false")
	}
	root := readRoot(t, p)
	if _, ok := root["hosts"]; !ok {
		t.Fatal("hosts 键应保留")
	}
	if sec, _ := root["sync"].(map[string]any); sec == nil || sec["target"] == "" {
		t.Fatalf("sync.target 未写入: %v", root["sync"])
	}

	rec3 := httptest.NewRecorder()
	s.hDBSyncPut(rec3, httptest.NewRequest("PUT", "/api/dbsync", bytes.NewBufferString(`{"target":"  "}`)))
	var r3 dbSyncResp
	if err := json.Unmarshal(rec3.Body.Bytes(), &r3); err != nil {
		t.Fatal(err)
	}
	if r3.Configured != "" || r3.Target != `D:\default-dir\erp_data.db` {
		t.Fatalf("清除后不符: %+v", r3)
	}
	if _, ok := readRoot(t, p)["sync"]; ok {
		t.Fatal("sync 键应被删除")
	}
}
