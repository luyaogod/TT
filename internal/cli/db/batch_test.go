package db

// 批量（--batch）的契约：输入解析、逐条成败不互相拖累、以及“只输出 JSON”。
// 全部离线：执行部分用假会话喂，不连任何库。

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/debug"
	"github.com/luyaogod/TT/internal/output"
)

type fakeQuerier struct {
	res  map[string]*debug.DBQueryResult
	errs map[string]error
	seen []string
	// seenMax 记下每条实际拿到的行数上限：--max-rows 必须真的传到会话层，
	// 而不是只在 CLI 层校验完就丢掉。
	seenMax []int
}

func (f *fakeQuerier) Query(_ context.Context, sql string, _ int, maxRows int) (*debug.DBQueryResult, error) {
	f.seen = append(f.seen, sql)
	f.seenMax = append(f.seenMax, maxRows)
	if err := f.errs[sql]; err != nil {
		return nil, err
	}
	if r, ok := f.res[sql]; ok {
		return r, nil
	}
	return &debug.DBQueryResult{Columns: []string{"X"}, Rows: [][]string{{"1"}}, TotalRows: 1}, nil
}

func TestParseBatchDefaultsAndValidation(t *testing.T) {
	stmts, err := parseBatch([]byte(`[{"sql":"select 1 from dual"},{"id":"b","sql":"select 2 from dual"}]`))
	if err != nil {
		t.Fatalf("不该报错: %v", err)
	}
	if len(stmts) != 2 || stmts[0].ID != "1" || stmts[1].ID != "b" {
		t.Errorf("id 缺省应按序号补（1..N）: %+v", stmts)
	}

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"空数组", `[]`, "没有任何语句"},
		{"非数组", `{"sql":"select 1 from dual"}`, "JSON 数组"},
		{"缺 sql", `[{"id":"a"}]`, "缺 sql"},
		{"id 重复", `[{"id":"a","sql":"select 1 from dual"},{"id":"a","sql":"select 2 from dual"}]`, "id 重复"},
	}
	for _, c := range cases {
		_, err := parseBatch([]byte(c.in))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: 期望错误含 %q，得到 %v", c.name, c.want, err)
		}
	}
}

func TestRunBatchQueriesIsolatesPerStatementFailure(t *testing.T) {
	boom := errors.New("ORA-00942: table or view does not exist")
	q := &fakeQuerier{
		res: map[string]*debug.DBQueryResult{
			"select 1 from dual": {
				Columns: []string{"X"}, Rows: [][]string{{"1"}}, TotalRows: 1, ServerLimited: true,
				Truncated: false,
			},
		},
		errs: map[string]error{"select bad from nope": boom},
	}
	stmts := []batchStmt{
		{ID: "ok", SQL: "select 1 from dual"},
		{ID: "boom", SQL: "select bad from nope"},
		{ID: "write", SQL: "delete from imaa_t"},
	}
	items, failed := runBatchQueries(context.Background(), q, stmts, 30, 200)

	if len(items) != 3 || failed != 2 {
		t.Fatalf("应有 3 条、2 条失败，得到 %d 条、failed=%d", len(items), failed)
	}
	if !items[0].OK || items[0].ID != "ok" || items[0].TotalRows != 1 || !items[0].ServerLimited {
		t.Errorf("成功那条的结果没带全: %+v", items[0])
	}
	if items[1].OK || items[1].Code != "DB_FAILED" || !strings.Contains(items[1].Error, "ORA-00942") {
		t.Errorf("查询失败应记 DB_FAILED 且带原因: %+v", items[1])
	}
	if items[2].OK || items[2].Code != "SQL_REJECTED" {
		t.Errorf("写语句应被只读闸门逐条拒掉（不是整批拒）: %+v", items[2])
	}
	// 被闸门拒的那条不该发到库上：发出去的只有前两条。
	if len(q.seen) != 2 || strings.Contains(strings.Join(q.seen, "|"), "delete") {
		t.Errorf("被拒语句不应执行，实发: %v", q.seen)
	}
}

func TestRunBatchQueriesKeepsOrderAndIDs(t *testing.T) {
	q := &fakeQuerier{}
	stmts := []batchStmt{{ID: "a", SQL: "select 1 from dual"}, {ID: "b", SQL: "select 2 from dual"}}
	items, failed := runBatchQueries(context.Background(), q, stmts, 30, 200)
	if failed != 0 || items[0].ID != "a" || items[1].ID != "b" {
		t.Fatalf("顺序与 id 必须原样回: %+v failed=%d", items, failed)
	}
	if q.seen[0] != "select 1 from dual" || q.seen[1] != "select 2 from dual" {
		t.Errorf("执行顺序应保持输入顺序: %v", q.seen)
	}
	if q.seenMax[0] != 200 || q.seenMax[1] != 200 {
		t.Errorf("行数上限要一路传到会话层: %v", q.seenMax)
	}
}

func TestClampMaxRows(t *testing.T) {
	cases := []struct {
		in      int
		want    int
		wantErr bool
	}{
		{0, 200, false},   // 不填 = 默认
		{-1, 200, false},  // 负数当默认，不报错
		{1, 1, false},     // 下限 1 行也合法
		{200, 200, false}, // 默认值本身
		{5000, 5000, false},
		{5001, 0, true}, // 超硬上限必须拒，不能静默夹到 5000
	}
	for _, c := range cases {
		got, err := clampMaxRows(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("clampMaxRows(%d) 应报错，得到 %d", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("clampMaxRows(%d) 不该报错: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("clampMaxRows(%d) = %d，期望 %d", c.in, got, c.want)
		}
	}
}

// 截断必须留下"下一步怎么办"，只有真截断（且原因是行数）才加 note。
func TestAddRowLimitNote(t *testing.T) {
	trunc := &debug.DBQueryResult{TotalRows: 201, Rows: make([][]string, 200), Truncated: true, TruncReason: "rows", RowLimit: 200}
	addRowLimitNote(trunc)
	if len(trunc.Notes) != 1 || !strings.Contains(trunc.Notes[0], "COUNT(*)") || !strings.Contains(trunc.Notes[0], "200") {
		t.Errorf("截断 note 要带上限与三条出路: %v", trunc.Notes)
	}

	notTrunc := &debug.DBQueryResult{TotalRows: 3, Rows: make([][]string, 3)}
	addRowLimitNote(notTrunc)
	if len(notTrunc.Notes) != 0 {
		t.Errorf("没截断就不该加 note: %v", notTrunc.Notes)
	}

	otherReason := &debug.DBQueryResult{Truncated: true, TruncReason: "bytes", RowLimit: 200}
	addRowLimitNote(otherReason)
	if len(otherReason.Notes) != 0 {
		t.Errorf("只有按行数截断才加这条 note: %v", otherReason.Notes)
	}
	addRowLimitNote(nil) // 不该 panic
}

func TestQueryBatchRefusesNonJSONFormats(t *testing.T) {
	oldCSV, oldJSON := common.CSV, common.JSON
	common.CSV, common.JSON = true, false
	t.Cleanup(func() { common.CSV, common.JSON = oldCSV, oldJSON })

	err := QueryBatch("q.json", 99, 30, SQLRouteAuto, 200)
	var oe *output.Error
	if !errors.As(err, &oe) || oe.Exit != 2 || oe.Code != "USAGE" {
		t.Fatalf("--csv 下的批量应是 USAGE/2，得到 %v", err)
	}
	if !strings.Contains(oe.Error(), "批量只输出 JSON") {
		t.Errorf("错误要说明原因，得到 %q", oe.Error())
	}
}
