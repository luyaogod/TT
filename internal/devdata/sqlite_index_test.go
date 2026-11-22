package devdata

// CreatePKIndexes 的自锁回归（实测事故：单连接池上"边扫 rows 边 Exec"会永久卡住）。
//
// 这条测试同时是**超时哨**：没有修复时 CreatePKIndexes 永不返回，靠 goroutine + 超时
// 把它变成一条普通的红灯，而不是让整个 go test 挂到 CI 超时。
// 判定依据是"能不能在有限时间内返回并真的建出索引"，不是内部实现细节。

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCreatePKIndexes_NoSelfDeadlockOnSingleConn(t *testing.T) {
	w, err := openSQLite(filepath.Join(t.TempDir(), "x.db"))
	if err != nil {
		t.Fatalf("openSQLite: %v", err)
	}
	defer w.Close()

	// dzed_t 是"数据表键值档"：dzed003='P' 表示主键定义，dzed004 是列清单。
	if err := w.BeginTable("dzed_t", []string{"dzed001", "dzed002", "dzed003", "dzed004"}); err != nil {
		t.Fatalf("建 dzed_t: %v", err)
	}
	// 一行指向**本次已同步**的表（delivery_t），一行指向没同步的表（other_t，应被跳过）。
	if err := w.InsertRows([][]string{
		{"delivery_t", "pk1", "P", "id"},
		{"other_t", "pk2", "P", "id"},
	}); err != nil {
		t.Fatalf("写 dzed_t: %v", err)
	}
	if err := w.BeginTable("delivery_t", []string{"id", "name"}); err != nil {
		t.Fatalf("建 delivery_t: %v", err)
	}
	if err := w.InsertRows([][]string{{"1", "a"}}); err != nil {
		t.Fatalf("写 delivery_t: %v", err)
	}

	type result struct {
		warns []string
	}
	done := make(chan result, 1)
	go func() {
		done <- result{warns: w.CreatePKIndexes(map[string]bool{"delivery_t": true})}
	}()

	var got result
	select {
	case got = <-done:
	case <-time.After(15 * time.Second):
		// 自锁时这个 goroutine 永远不返回；测试进程随后退出，泄漏可接受。
		t.Fatal("CreatePKIndexes 15s 未返回：单连接池上自锁了（先读完 dzed_t 并 Close，再建索引）")
	}
	if len(got.warns) != 0 {
		t.Errorf("正常建索引不该有警告: %v", got.warns)
	}

	// 索引真的建成了（唯一索引会在重复键上失败，这里只验证存在与列序）。
	var n int
	if err := w.db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_delivery_t_pk1'`).Scan(&n); err != nil {
		t.Fatalf("查 sqlite_master: %v", err)
	}
	if n != 1 {
		t.Errorf("索引 idx_delivery_t_pk1 没建出来（n=%d）", n)
	}

	// 再跑一次必须幂等（IF NOT EXISTS），且不产生"应建但一个都没建成"的噪音。
	done2 := make(chan []string, 1)
	go func() { done2 <- w.CreatePKIndexes(map[string]bool{"delivery_t": true}) }()
	select {
	case warns := <-done2:
		if len(warns) != 0 {
			t.Errorf("重复建索引不该有警告: %v", warns)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("CreatePKIndexes 第二次调用 15s 未返回（自锁）")
	}

	// 没同步的表不能被建索引：CREATE ... ON other_t 会报 no such table。
	var other int
	if err := w.db.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_other_t%'`).Scan(&other); err != nil {
		t.Fatalf("查 other_t 索引: %v", err)
	}
	if other != 0 {
		t.Errorf("未同步的表不该建索引，实际建了 %d 个", other)
	}
}
