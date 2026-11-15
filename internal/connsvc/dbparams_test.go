package connsvc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/luyaogod/TT/pkg/dbconfig"
	"github.com/luyaogod/TT/pkg/erpdb"
)

// fakeArgsSession 在 fakeSession 上补参数绑定与批量接口。
type fakeArgsSession struct {
	*fakeSession
	gotStmts   []erpdb.Stmt
	gotAtomic  bool
	batchN     int64
	batchFail  int
	batchErr   error
	lastSQL    string
	lastParams []string
}

func (f *fakeArgsSession) QueryArgs(_ context.Context, sql string, params []string) ([]string, [][]string, error) {
	f.lastSQL, f.lastParams = sql, params
	return f.cols, f.rows, f.queryErr
}

func (f *fakeArgsSession) ExecArgs(_ context.Context, sql string, params []string) (int64, error) {
	f.lastSQL, f.lastParams = sql, params
	return f.execN, f.execErr
}

func (f *fakeArgsSession) ExecBatch(_ context.Context, stmts []erpdb.Stmt, atomic bool) (int64, int, error) {
	f.gotStmts, f.gotAtomic = stmts, atomic
	return f.batchN, f.batchFail, f.batchErr
}

func openArgsSession(t *testing.T, fs *fakeArgsSession, readOnly bool) (*Service, string) {
	t.Helper()
	s := newTestService(t, Options{
		OpenDB: func(context.Context, dbconfig.Connection, erpdb.SessionOptions) (erpdb.Session, error) {
			return fs, nil
		},
	})
	h, _, err := s.OpenDB(context.Background(), BuiltinCaller(), "prod", "", readOnly)
	if err != nil {
		t.Fatalf("开库失败: %v", err)
	}
	return s, h
}

// TestDBQueryArgsPassesParams 带参查询把参数原样交给会话，不自己拼 SQL。
func TestDBQueryArgsPassesParams(t *testing.T) {
	fs := &fakeArgsSession{fakeSession: &fakeSession{
		typ: "oracle", readOnly: true, cols: []string{"c"}, rows: [][]string{{"v"}},
	}}
	s, h := openArgsSession(t, fs, true)
	res, err := s.DBQueryArgs(context.Background(), BuiltinCaller(), h, "SELECT ? FROM t WHERE a = ?", []string{"x", "y"}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 1 || res.Rows[0][0] != "v" {
		t.Fatalf("结果不对: %+v", res)
	}
	if fs.lastSQL != "SELECT ? FROM t WHERE a = ?" {
		t.Errorf("SQL 被改动了: %q", fs.lastSQL)
	}
	if len(fs.lastParams) != 2 || fs.lastParams[0] != "x" || fs.lastParams[1] != "y" {
		t.Errorf("参数没传到: %v", fs.lastParams)
	}
}

// TestDBBatchAtomicAndFailureIndex 批量带 atomic 下发；失败时返回失败下标与稳定码。
func TestDBBatchAtomicAndFailureIndex(t *testing.T) {
	fs := &fakeArgsSession{fakeSession: &fakeSession{typ: "oracle"}, batchN: 5}
	s, h := openArgsSession(t, fs, false)
	stmts := []erpdb.Stmt{
		{SQL: "INSERT INTO t VALUES (?, ?)", Params: []string{"1", "2"}},
		{SQL: "DELETE FROM t WHERE a = ?", Params: []string{"3"}},
	}
	n, failed, err := s.DBBatch(context.Background(), BuiltinCaller(), h, stmts, true)
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 || failed != -1 {
		t.Fatalf("n=%d failed=%d", n, failed)
	}
	if !fs.gotAtomic || len(fs.gotStmts) != 2 {
		t.Fatalf("批量参数没传对: atomic=%v stmts=%d", fs.gotAtomic, len(fs.gotStmts))
	}

	fs.batchErr = errors.New("dup key")
	fs.batchFail = 1
	_, failed, err = s.DBBatch(context.Background(), BuiltinCaller(), h, stmts, true)
	if !IsCode(err, CodeBatchFailed) {
		t.Fatalf("该报 DB_BATCH_FAILED，得 %v", err)
	}
	if failed != 1 {
		t.Fatalf("失败下标 = %d, want 1", failed)
	}
	if !strings.Contains(err.Message, "第 2 条") {
		t.Errorf("错误消息该带失败语句序号: %q", err.Message)
	}
}

// TestDBBatchRejectsReadOnlySession 只读会话上批量写一律拒绝。
func TestDBBatchRejectsReadOnlySession(t *testing.T) {
	fs := &fakeArgsSession{fakeSession: &fakeSession{typ: "oracle", readOnly: true}}
	s, h := openArgsSession(t, fs, true)
	if _, _, err := s.DBBatch(context.Background(), BuiltinCaller(), h, []erpdb.Stmt{{SQL: "DELETE FROM t"}}, true); !IsCode(err, CodeSessionReadOnly) {
		t.Fatalf("只读会话该报 DB_SESSION_READONLY，得 %v", err)
	}
}

// TestOpenDBWithEntResolvesAccount 按企业编号解析账号；未知企业/无账号给稳定码。
func TestOpenDBWithEntResolvesAccount(t *testing.T) {
	var got dbconfig.Connection
	s := newTestService(t, Options{
		OpenDB: func(_ context.Context, c dbconfig.Connection, _ erpdb.SessionOptions) (erpdb.Session, error) {
			got = c
			return &fakeSession{typ: "oracle", version: "v"}, nil
		},
	})
	_, res, err := s.OpenDBWithEnt(context.Background(), BuiltinCaller(), "prod", "", false, 99)
	if err != nil {
		t.Fatal(err)
	}
	if got.Account != "ds" {
		t.Errorf("实际账号 = %q, want ds", got.Account)
	}
	if res.Account != "ds" || res.Ent != 99 {
		t.Errorf("结果回执 = %+v", res)
	}

	if _, _, err := s.OpenDBWithEnt(context.Background(), BuiltinCaller(), "prod", "", false, 1234); !IsCode(err, CodeEntUnknown) {
		t.Fatalf("未知企业该报 DB_ENT_UNKNOWN，得 %v", err)
	}
	if _, _, err := s.OpenDBWithEnt(context.Background(), BuiltinCaller(), "prod", "", false, 2222); !IsCode(err, CodeEntNoAccount) {
		t.Fatalf("无账号企业该报 DB_ENT_NO_ACCOUNT，得 %v", err)
	}
}
