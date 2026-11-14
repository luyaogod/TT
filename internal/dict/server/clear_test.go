package server

// 「清理」端点的判据：删得掉、只删该删的、该拒绝时拒绝。
//
// 这两个动作不可撤销（删的是本地那份镜像/字典库），所以测试的重点不是"删成功"，
// 而是**边界**：别人的镜像不能跟着没、根目录不能整棵删、任务在跑时不能删。

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func postClear(h http.HandlerFunc, env string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest("POST", "/api/clear", bytes.NewBufferString(fmt.Sprintf(`{"env":%q}`, env))))
	return rec
}

// 正常清理：只删该环境那一棵，兄弟环境原样保留。
func TestMirrorClear(t *testing.T) {
	root := t.TempDir()
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"h","user":"u"},{"name":"b","host":"h2","user":"u"}]}}`)
	setSection(t, p, "mirror", map[string]any{"dir": root})
	s := New(p)

	aDir := filepath.Join(root, "a")
	if err := os.MkdirAll(filepath.Join(aDir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(aDir, "sub", "x.4gl"), bytes.Repeat([]byte("x"), 128), 0o644); err != nil {
		t.Fatal(err)
	}
	bDir := filepath.Join(root, "b")
	if err := os.MkdirAll(bDir, 0o755); err != nil {
		t.Fatal(err)
	}

	rec := postClear(s.hMirrorClear, "a")
	if rec.Code != http.StatusOK {
		t.Fatalf("清理 a 应 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(aDir); !os.IsNotExist(err) {
		t.Errorf("a 的镜像目录应已删除: %v", err)
	}
	if _, err := os.Stat(bDir); err != nil {
		t.Errorf("b 的镜像目录不该被动: %v", err)
	}
	var resp struct {
		OK    bool   `json:"ok"`
		Path  string `json:"path"`
		Freed int64  `json:"freed"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.OK || filepath.Clean(resp.Path) != filepath.Clean(aDir) || resp.Freed <= 0 {
		t.Errorf("响应不符: %+v", resp)
	}
}

// 两个环境指向同一个目录：拒绝（否则一次点击会删掉别人的镜像）。
func TestMirrorClearRefusesSharedDir(t *testing.T) {
	shared := t.TempDir()
	p := writeTempConfig(t, fmt.Sprintf(
		`{"hosts":{"sshs":[{"name":"a","host":"h","user":"u","mirrorDir":%q},{"name":"b","host":"h2","user":"u","mirrorDir":%q}]}}`,
		shared, shared))
	s := New(p)
	if err := os.WriteFile(filepath.Join(shared, "keep.4gl"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := postClear(s.hMirrorClear, "a")
	if rec.Code != http.StatusConflict {
		t.Fatalf("共用目录应 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(shared, "keep.4gl")); err != nil {
		t.Errorf("拒绝时不该删任何东西: %v", err)
	}
}

// 镜像目录就是根目录本身：拒绝（那不是"某个环境的镜像"）。
func TestMirrorClearRefusesRootItself(t *testing.T) {
	root := t.TempDir()
	p := writeTempConfig(t, fmt.Sprintf(
		`{"mirror":{"dir":%q},"hosts":{"sshs":[{"name":"a","host":"h","user":"u","mirrorDir":%q}]}}`, root, root))
	s := New(p)

	rec := postClear(s.hMirrorClear, "a")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("根目录本身应 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(root); err != nil {
		t.Errorf("拒绝时不该删根目录: %v", err)
	}
}

// 目录本来就不存在：算成功（幂等），不是错误。
func TestMirrorClearMissingDirIsOK(t *testing.T) {
	root := t.TempDir()
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"h","user":"u"}]}}`)
	setSection(t, p, "mirror", map[string]any{"dir": root})
	s := New(p)

	rec := postClear(s.hMirrorClear, "a")
	if rec.Code != http.StatusOK {
		t.Fatalf("不存在的目录应 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// 正常清理本地字典库：主文件与伴生文件一起删，目录保留。
func TestDBSyncClear(t *testing.T) {
	dbDir := t.TempDir()
	p := writeTempConfig(t, fmt.Sprintf(`{"dict":{"dir":%q},"hosts":{"sshs":[{"name":"a","host":"h","user":"u"}]}}`, dbDir))
	s := New(p)

	target := s.dbPathFor("a")
	if target == "" {
		t.Fatal("算不出库路径")
	}
	for _, f := range []string{target, target + "-wal", target + ".bak"} {
		if err := os.WriteFile(f, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	rec := postClear(s.hDBSyncClear, "a")
	if rec.Code != http.StatusOK {
		t.Fatalf("清理应 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, f := range []string{target, target + "-wal", target + ".bak"} {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("%s 应已删除", f)
		}
	}
	if _, err := os.Stat(dbDir); err != nil {
		t.Errorf("库目录不该被删: %v", err)
	}
}

// 没有本地库时也算成功（幂等）。
func TestDBSyncClearMissingIsOK(t *testing.T) {
	dbDir := t.TempDir()
	p := writeTempConfig(t, fmt.Sprintf(`{"dict":{"dir":%q},"hosts":{"sshs":[{"name":"a","host":"h","user":"u"}]}}`, dbDir))
	s := New(p)

	rec := postClear(s.hDBSyncClear, "a")
	if rec.Code != http.StatusOK {
		t.Fatalf("没有本地库应 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// 任务在跑时不许删（同步正写着那个文件 / 拉取正写着那棵树）。
func TestClearRefusesWhileJobRunning(t *testing.T) {
	root := t.TempDir()
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"h","user":"u"}]}}`)
	setSection(t, p, "mirror", map[string]any{"dir": root})
	s := New(p)

	s.syncMu.Lock()
	s.sync = dbSyncJob{Running: true, Env: "a"}
	s.syncMu.Unlock()
	if rec := postClear(s.hDBSyncClear, "a"); rec.Code != http.StatusConflict {
		t.Errorf("同步在跑时应 409, got %d: %s", rec.Code, rec.Body.String())
	}

	s.mirrorMu.Lock()
	s.mirror = mirrorJob{Running: true, Env: "a"}
	s.mirrorMu.Unlock()
	if rec := postClear(s.hMirrorClear, "a"); rec.Code != http.StatusConflict {
		t.Errorf("拉取在跑时应 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

// 未知环境：400，不猜。
func TestClearUnknownEnv(t *testing.T) {
	p := writeTempConfig(t, `{"hosts":{"sshs":[{"name":"a","host":"h","user":"u"}]}}`)
	s := New(p)
	if rec := postClear(s.hDBSyncClear, "nope"); rec.Code != http.StatusBadRequest {
		t.Errorf("未知环境应 400, got %d", rec.Code)
	}
	if rec := postClear(s.hMirrorClear, "nope"); rec.Code != http.StatusBadRequest {
		t.Errorf("未知环境应 400, got %d", rec.Code)
	}
}
