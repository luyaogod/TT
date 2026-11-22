package devdata

// 测试夹具：假的规格表来源与假的 SFTP 来源。
//
// 为什么值得写这套夹具：mirror/fetch 真正容易错的地方（顺序、回滚、取消、上限跳过）
// 用真库真服务器验证代价极高（要 112 万行的 DZFK_T、要一台测机），而这恰恰是要钉住的部分。

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/luyaogod/TT/pkg/erpdb"
)

// fakeTable 内存版的规格表来源。
type fakeTable struct {
	// cols 表名 → 列名
	cols map[string][]string
	// rows 表名 → 行
	rows map[string][][]string
	// missing 表名 → 当作"表不存在"（ORA-00942）
	missing map[string]bool
	// failOn 表名 → 首次 Query 返回该错误
	failOn map[string]error
	// countErr 表名 → CountRows 返回该错误
	countErr map[string]error
	// countOverride 表名 → CountRows 直接返回这个数（不落真实行，用来模拟百万行大表）
	countOverride map[string]int64
	// queried 记录查过的表（顺序）
	queried []string
	// where 记录每张表收到的过滤条件
	where map[string]string
	// orders 记录每张表收到的排序键（分页用）
	orders map[string][]string
	// noOrder 模拟"取不到主键"（分页退化）
	noOrder bool
	mu      sync.Mutex
}

func newFakeTable() *fakeTable {
	return &fakeTable{
		cols:          map[string][]string{},
		rows:          map[string][][]string{},
		missing:       map[string]bool{},
		failOn:        map[string]error{},
		countErr:      map[string]error{},
		countOverride: map[string]int64{},
		where:         map[string]string{},
		orders:        map[string][]string{},
	}
}

func (f *fakeTable) CountRows(_ context.Context, table string) (int64, error) {
	if err := f.countErr[table]; err != nil {
		return 0, err
	}
	if f.missing[table] {
		return 0, fmt.Errorf("ORA-00942: table or view does not exist")
	}
	if n, ok := f.countOverride[table]; ok {
		return n, nil
	}
	return int64(len(f.rows[table])), nil
}

// OrderColumns 假主键：固定返回表名首列（测试里够用）。
func (f *fakeTable) OrderColumns(_ context.Context, table string) ([]string, error) {
	if f.noOrder {
		return nil, nil
	}
	if cols := f.cols[table]; len(cols) > 0 {
		return cols[:1], nil
	}
	return nil, nil
}

func (f *fakeTable) TableColumns(_ context.Context, table string) (TableMeta, error) {
	if f.noOrder {
		return TableMeta{}, nil
	}
	if cols := f.cols[table]; len(cols) > 0 {
		return TableMeta{Columns: cols, LOB: map[string]bool{}}, nil
	}
	return TableMeta{}, nil
}

func (f *fakeTable) Query(ctx context.Context, table, where string, orderBy []string, _ TableMeta, fn erpdb.RowFunc) error {
	f.mu.Lock()
	f.orders[table] = orderBy
	f.mu.Unlock()
	f.mu.Lock()
	f.queried = append(f.queried, table)
	f.where[table] = where
	f.mu.Unlock()
	if err := f.failOn[table]; err != nil {
		delete(f.failOn, table)
		return err
	}
	if f.missing[table] {
		return fmt.Errorf("ORA-00942: table or view does not exist")
	}
	cols := f.cols[table]
	if cols == nil {
		cols = []string{table + "_C1"}
	}
	// 与生产实现同一契约：先给一次列名（row=nil），空表也给
	if err := fn(cols, nil); err != nil {
		return err
	}
	for _, r := range f.rows[table] {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fn(cols, r); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeTable) Close() {}

// setTable 设一张表的列与行（rows 的每行必须与 cols 等长）。
func (f *fakeTable) setTable(name string, cols []string, rows ...[]string) {
	f.cols[name] = cols
	f.rows[name] = rows
}

// fakeFiles 内存版的 SFTP 来源。
type fakeFiles struct {
	top   string
	files map[string][]byte
	// dirs 相对路径 → 是否存在（Stat 认目录）
	dirs map[string]bool
	// mtimes 相对路径 → mtime（不给就用固定的旧时间）
	mtimes map[string]time.Time
	// failRead 相对路径 → 读它时返回该错误
	failRead map[string]error
	// walkErr Walk 返回该错误
	walkErr error
	// readAll 记录读过的路径
	readAll []string
	mu      sync.Mutex
	// closed 是否被关闭
	closed bool
}

func newFakeFiles() *fakeFiles {
	return &fakeFiles{top: "/u1/t35tst", files: map[string][]byte{}, dirs: map[string]bool{},
		mtimes: map[string]time.Time{}, failRead: map[string]error{}}
}

func (f *fakeFiles) Top() string { return f.top }

func (f *fakeFiles) Stat(_ context.Context, rel string) (RemoteFile, error) {
	rel = strings.TrimPrefix(rel, "/")
	if f.dirs[rel] {
		return RemoteFile{Rel: rel, IsDir: true}, nil
	}
	if b, ok := f.files[rel]; ok {
		mt := f.mtimes[rel]
		if mt.IsZero() {
			mt = time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
		}
		return RemoteFile{Rel: rel, Size: int64(len(b)), ModTime: mt}, nil
	}
	return RemoteFile{Rel: rel}, fmt.Errorf("%w: %s", fs.ErrNotExist, rel)
}

func (f *fakeFiles) ReadFile(_ context.Context, rel string, max int64) ([]byte, error) {
	rel = strings.TrimPrefix(rel, "/")
	f.mu.Lock()
	f.readAll = append(f.readAll, rel)
	f.mu.Unlock()
	if err := f.failRead[rel]; err != nil {
		delete(f.failRead, rel)
		return nil, err
	}
	b, ok := f.files[rel]
	if !ok {
		return nil, fmt.Errorf("%w: %s", fs.ErrNotExist, rel)
	}
	if max > 0 && int64(len(b)) > max {
		return nil, fmt.Errorf("超过上限")
	}
	return append([]byte(nil), b...), nil
}

func (f *fakeFiles) Walk(ctx context.Context, rel string, fn func(RemoteFile) error) error {
	if f.walkErr != nil {
		return f.walkErr
	}
	rel = strings.TrimPrefix(rel, "/")
	prefix := rel + "/"
	keys := make([]string, 0, len(f.files))
	for k := range f.files {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return fmt.Errorf("%w: %s", fs.ErrNotExist, rel)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		mt := f.mtimes[k]
		if mt.IsZero() {
			mt = time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
		}
		if err := fn(RemoteFile{Rel: k, Size: int64(len(f.files[k])), ModTime: mt}); err != nil {
			return err
		}
	}
	return nil
}

func (f *fakeFiles) Close() { f.closed = true }

// writeConfig 造一份带环境的 config.json，返回路径。
func writeConfig(t *testing.T, hosts string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	body := `{"hosts":{"sshs":[` + hosts + `]},"devdata":{"dir":""},"schemaVersion":4}`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatalf("写测试配置失败: %v", err)
	}
	return p
}
