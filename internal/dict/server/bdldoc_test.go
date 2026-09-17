package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// GET 默认未设置;PUT 写入 bdldoc.dir 且保留其余顶层键。
func TestBdldocGetPut(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"activeEnv":"a","sshs":[{"name":"a","host":"1.2.3.4","user":"u","password":"p"}]},"query":{"source":"local"}}`)
	s := New(p)
	dir := t.TempDir()

	rec := httptest.NewRecorder()
	s.hBdldocGet(rec, httptest.NewRequest("GET", "/api/bdldoc", nil))
	var st bdldocStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if st.Dir != "" || st.Exists {
		t.Fatalf("初始应未设置: %+v", st)
	}

	body, _ := json.Marshal(map[string]string{"dir": dir})
	rec2 := httptest.NewRecorder()
	s.hBdldocPut(rec2, httptest.NewRequest("PUT", "/api/bdldoc", bytes.NewBuffer(body)))
	if !bytes.Contains(rec2.Body.Bytes(), []byte(`"ok":true`)) {
		t.Fatalf("PUT 失败: %s", rec2.Body.String())
	}
	var st2 bdldocStatus
	if err := json.Unmarshal(rec2.Body.Bytes(), &st2); err != nil {
		t.Fatal(err)
	}
	if st2.Dir == "" || !st2.Exists {
		t.Fatalf("PUT 应返回已存在目录: %+v", st2)
	}

	// 落盘且其余键保留
	root := readRoot(t, p)
	if _, ok := root["hosts"]; !ok {
		t.Fatal("hosts 键应保留")
	}
	if _, ok := root["query"]; !ok {
		t.Fatal("query 键应保留")
	}
	sec, _ := root["bdldoc"].(map[string]any)
	if sec == nil || sec["dir"] == "" {
		t.Fatalf("bdldoc.dir 未写入: %v", root["bdldoc"])
	}

	// 再 GET 应读到
	rec3 := httptest.NewRecorder()
	s.hBdldocGet(rec3, httptest.NewRequest("GET", "/api/bdldoc", nil))
	var st3 bdldocStatus
	_ = json.Unmarshal(rec3.Body.Bytes(), &st3)
	if st3.Dir == "" || !st3.Exists {
		t.Fatalf("GET 结果不符: %+v", st3)
	}
}

// 目录不存在时 exists=false,但设置仍成功。
func TestBdldocPutNonExistentDir(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"1.2.3.4"}]}}`)
	s := New(p)
	body, _ := json.Marshal(map[string]string{"dir": `D:\definitely\not\here\bdl`})
	rec := httptest.NewRecorder()
	s.hBdldocPut(rec, httptest.NewRequest("PUT", "/api/bdldoc", bytes.NewBuffer(body)))
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"ok":true`)) {
		t.Fatalf("应允许设置不存在的目录: %s", rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(`"exists":true`)) {
		t.Fatalf("不存在目录 exists 应为 false: %s", rec.Body.String())
	}
}

// 空目录拒绝。
func TestBdldocPutEmpty(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"1.2.3.4"}]}}`)
	s := New(p)
	rec := httptest.NewRecorder()
	s.hBdldocPut(rec, httptest.NewRequest("PUT", "/api/bdldoc", bytes.NewBufferString(`{"dir":"   "}`)))
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"ok":false`)) {
		t.Fatalf("空目录应拒绝: %s", rec.Body.String())
	}
}
