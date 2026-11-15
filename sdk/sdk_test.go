package sdk

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"
)

// fakeHost 用最少的 JSON 应答模拟 TT 的 Host API（Connect unary JSON）。
func fakeHost(t *testing.T, gotAuth *string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	jsonReply := func(w http.ResponseWriter, body string) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
	mux.HandleFunc("/hostrpc/tt.host.v1.ProfilesService/List", func(w http.ResponseWriter, r *http.Request) {
		*gotAuth = r.Header.Get("Authorization")
		jsonReply(w, `{"profiles":[{"env":"109tst","dbType":"oracle","dbAddress":"db:1521/t35tst",
			"ents":[{"ent":99,"account":"dsdemo"},{"ent":2222,"account":"dsstd"}]}]}`)
	})
	mux.HandleFunc("/hostrpc/tt.host.v1.ConnectionsService/OpenDB", func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, `{"handle":"h1","readOnly":false,"dbType":"oracle","account":"dsdemo","ent":99}`)
	})
	mux.HandleFunc("/hostrpc/tt.host.v1.ConnectionsService/Close", func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, `{}`)
	})
	mux.HandleFunc("/hostrpc/tt.host.v1.DBSessionService/Query", func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, `{"columns":["A"],"rows":[{"values":["1"]},{"values":["2"]}],"rowCount":"2"}`)
	})
	mux.HandleFunc("/hostrpc/tt.host.v1.DBSessionService/Batch", func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, `{"rowsAffected":"2","failedIndex":-1}`)
	})
	mux.HandleFunc("/hostrpc/tt.host.v1.SSHSessionService/GetFile", func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, `{"data":"aGk="}`)
	})
	mux.HandleFunc("/hostrpc/tt.host.v1.SSHSessionService/PutFile", func(w http.ResponseWriter, r *http.Request) {
		jsonReply(w, `{}`)
	})
	return httptest.NewServer(mux)
}

// TestSDKConfigAndDB 配置读取、开库、查询、批量走同一条链路。
func TestSDKConfigAndDB(t *testing.T) {
	var auth string
	srv := fakeHost(t, &auth)
	defer srv.Close()

	c := newContext(context.Background(), srv.URL+"/hostrpc", "tok-1", "com.example.p", "0.1.0", t.TempDir())
	if envs := c.Config().Envs(); len(envs) != 1 || envs[0].Name != "109tst" {
		t.Fatalf("环境清单 = %+v", envs)
	}
	if auth != "Bearer tok-1" {
		t.Fatalf("Authorization = %q", auth)
	}
	ents := c.Config().Ents("109tst")
	if len(ents) != 2 || ents[0].Ent != 99 || ents[0].Account != "dsdemo" {
		t.Fatalf("企业清单 = %+v", ents)
	}

	sess, err := c.DB().Open(context.Background(), "109tst", OpenOptions{Ent: 99})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := sess.Query(context.Background(), "SELECT A FROM t WHERE x = ?", "v")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows.Values) != 2 || rows.Columns[0] != "A" {
		t.Fatalf("查询结果 = %+v", rows)
	}
	res, err := sess.Batch(context.Background(), []Statement{{SQL: "INSERT INTO t VALUES (?)", Params: []string{"1"}}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.RowsAffected != 2 || res.FailedIndex != -1 {
		t.Fatalf("批量结果 = %+v", res)
	}
	if err := sess.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestMapErrKeepsStableCode Connect 错误翻成带码的 *Error（无详情时按状态码兜底）。
func TestMapErrKeepsStableCode(t *testing.T) {
	err := mapErr(connect.NewError(connect.CodePermissionDenied, errors.New("没有权限")))
	if CodeOf(err) != CodePermissionDenied {
		t.Fatalf("错误码 = %q（%v）", CodeOf(err), err)
	}
	err = mapErr(connect.NewError(connect.CodeDeadlineExceeded, errors.New("超时")))
	if CodeOf(err) != CodeTimeout {
		t.Fatalf("超时该映射成 TIMEOUT，得 %q", CodeOf(err))
	}
}

// TestFSDataDir 插件数据目录：相对路径落在 DataDir 下。
func TestFSDataDir(t *testing.T) {
	dir := t.TempDir()
	c := newContext(context.Background(), "http://127.0.0.1:1", "tok", "p", "0.1.0", dir)
	if err := c.FS().WriteFile("a/b.txt", []byte("hi")); err != nil {
		t.Fatal(err)
	}
	got, err := c.FS().ReadFile("a/b.txt")
	if err != nil || string(got) != "hi" {
		t.Fatalf("读写失败: %q %v", got, err)
	}
	if c.DataDir() != dir {
		t.Fatalf("DataDir = %q", c.DataDir())
	}
}

// TestSSHFileRoundTrip 远端文件读写走 Host API（字节 base64 由 protojson 处理）。
func TestSSHFileRoundTrip(t *testing.T) {
	var auth string
	srv := fakeHost(t, &auth)
	defer srv.Close()
	c := newContext(context.Background(), srv.URL+"/hostrpc", "tok", "p", "0.1.0", t.TempDir())
	s := &sshSession{cli: c.cli, handle: "h"}

	local := filepath.Join(t.TempDir(), "a.png")
	if err := s.Get(context.Background(), "/remote/a.png", local); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(local)
	if err != nil || string(got) != "hi" {
		t.Fatalf("读回 = %q %v", got, err)
	}

	src := filepath.Join(t.TempDir(), "b.png")
	if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Put(context.Background(), src, "/remote/b.png"); err != nil {
		t.Fatal(err)
	}
}
// TestJobsRunReportsProgress ctx.jobs 进程内任务：立即返回 id，fn 异步执行。
func TestJobsRunReportsProgress(t *testing.T) {
	c := newContext(context.Background(), "http://127.0.0.1:1", "tok", "p", "0.1.0", t.TempDir())
	js := c.Jobs()
	if js == nil {
		t.Fatal("Jobs() 不该是 nil")
	}
	done := make(chan struct{})
	id, err := js.Run(context.Background(), "demo", func(ctx context.Context, r Reporter) error {
		r.Progress(50, "halfway")
		r.Log("step")
		close(done)
		return nil
	})
	if err != nil || id == "" {
		t.Fatalf("Run = %q, %v", id, err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("任务没有执行")
	}
}