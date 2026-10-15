package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tt/internal/fep"
)

// 录制面板的服务端：列 / 看 / 删 + "现在在录吗"。
// "起一轮/停一轮"是**另起进程**（见 recordings.go 头注释），所以这里不测它 ——
// 那条链在命令行上真机验过（`tt fep rec start` / `session close`），
// 面板这一侧只是把它包成一个接口。

func seedLibRecording(t *testing.T, dataDir, env, ent, id, prog string) string {
	t.Helper()
	cache := filepath.Join(dataDir, "fep")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	body := []byte("om 3 {{}}\n")
	for _, n := range []string{id + "-1-rt2gdc.log", id + "-1-gdc2rt.log"} {
		if err := os.WriteFile(filepath.Join(cache, n), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := fep.WriteMeta(cache, fep.RecordingMeta{
		ID: id, Path: fep.PathDebug, Prog: prog, Env: env, Ent: ent, Zone: "36",
		Origin: fep.OriginUser, Retention: fep.RetainLibrary,
		Started: time.Now(), Finished: time.Now().Add(time.Minute),
		Fingerprint: fep.Fingerprint{Total: 3, Hash: "abcdef0123456789"},
		Conns:       []fep.MetaConn{{ID: 1, BytesRT: 100, BytesFE: 20}},
	}); err != nil {
		t.Fatal(err)
	}
	item, err := fep.NewLibrary(dataDir).Promote(cache, id)
	if err != nil {
		t.Fatal(err)
	}
	return item.Dir
}

func getJSON(t *testing.T, s *Server, target string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

// 列表按 环境 + ENT 过滤，两处落点都列出来。
func TestRecordingsListFiltersByEnvAndEnt(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	dataDir := filepath.Dir(cfgPath)
	seedLibRecording(t, dataDir, "主机109", "99", "fep-a", "bsft001_wf")
	seedLibRecording(t, dataDir, "主机109", "77", "fep-b", "bsft001_wf")
	seedLibRecording(t, dataDir, "尚阳通", "1", "fep-c", "axmt500_wf")

	code, out := getJSON(t, s, "/api/recordings?env=主机109&ent=99")
	if code != http.StatusOK {
		t.Fatalf("列表该成功：%d %v", code, out)
	}
	rows, _ := out["recordings"].([]any)
	if len(rows) != 1 {
		t.Fatalf("按 环境+ENT 过滤后该只剩一份：%v", out)
	}
	row, _ := rows[0].(map[string]any)
	for k, want := range map[string]string{
		"id": "fep-a", "where": "长期库", "prog": "bsft001_wf", "origin": "user", "retention": "library",
	} {
		if got, _ := row[k].(string); got != want {
			t.Errorf("%s = %q，应为 %q", k, got, want)
		}
	}
	if got, _ := row["bytes"].(float64); got != 120 {
		t.Errorf("体积该把两个方向加起来（100+20）：%v", row["bytes"])
	}
	// 不给过滤条件 = 全部列出
	if _, all := getJSON(t, s, "/api/recordings"); len(all["recordings"].([]any)) != 3 {
		t.Fatalf("不给条件该列出全部三份：%v", all)
	}
}

// 详情：按 ID 找得到，且带上身份。
func TestRecordingShow(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	seedLibRecording(t, filepath.Dir(cfgPath), "主机109", "99", "fep-a", "bsft001_wf")
	code, out := getJSON(t, s, "/api/recordings/fep-a")
	if code != http.StatusOK {
		t.Fatalf("详情该成功：%d %v", code, out)
	}
	if out["meta"] == nil {
		t.Fatal("详情要带身份（前端要拿它显示路径/企业/编译件）")
	}
	if code, _ := getJSON(t, s, "/api/recordings/没有这一份"); code != http.StatusNotFound {
		t.Fatalf("找不到该回 404：%d", code)
	}
}

// 删：必须显式确认；删完库里没了，别的环境不动。
func TestRecordingDeleteNeedsConfirm(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	dataDir := filepath.Dir(cfgPath)
	seedLibRecording(t, dataDir, "主机109", "99", "fep-a", "bsft001_wf")
	seedLibRecording(t, dataDir, "主机109", "99", "fep-b", "bsft001_wf")

	del := func(target string) (int, map[string]any) {
		req := httptest.NewRequest(http.MethodDelete, target, nil)
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, req)
		out := map[string]any{}
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return rec.Code, out
	}
	if code, _ := del("/api/recordings/fep-a"); code != http.StatusBadRequest {
		t.Fatalf("没确认就该拒（录制不可再生）：%d", code)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "records", "主机109", "99", "fep-a")); err != nil {
		t.Fatalf("拒了就不该动文件：%v", err)
	}
	code, out := del("/api/recordings/fep-a?confirm=1")
	if code != http.StatusOK {
		t.Fatalf("确认后该删掉：%d %v", code, out)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "records", "主机109", "99", "fep-a")); !os.IsNotExist(err) {
		t.Fatalf("删完该没了：%v", err)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "records", "主机109", "99", "fep-b")); err != nil {
		t.Fatalf("别的录制不该动：%v", err)
	}
}

// "现在在录吗"看**会话登记**：录制件要收工才落身份，看文件会在录到一半时说"没有"。
func TestRecordingsCurrentFollowsTheRegistry(t *testing.T) {
	s, cfgPath := newTestServer(t, `{"schemaVersion":3,"hosts":{"sshs":[]}}`)
	dataDir := filepath.Dir(cfgPath)
	if _, out := getJSON(t, s, "/api/recordings/current"); out["recording"] != nil {
		t.Fatalf("没在录时该是 null：%v", out)
	}
	reg := fep.NewRegistry(dataDir)
	if err := reg.Save(&fep.Session{
		Env: "主机109", Zone: "36", Prog: "bsft001_wf", Ent: "99",
		Path: fep.PathDebug, Tag: "fep-now", Started: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	_, out := getJSON(t, s, "/api/recordings/current")
	cur, _ := out["recording"].(map[string]any)
	if cur == nil || cur["id"] != "fep-now" || cur["path"] != "debug" {
		t.Fatalf("该报出正在录的那一轮：%v", out)
	}
}
