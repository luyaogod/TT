package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"tt/internal/config"
)

// newStore 造一个落在临时数据目录里的 Store。
func newStore(t *testing.T) *Store {
	t.Helper()
	loc := config.LocationsAt(filepath.Join(t.TempDir(), "config.json"))
	s := NewStore(loc)
	s.now = time.Now
	s.lockWait = 2 * time.Second
	return s
}

// TestLoadMissingRegistryIsEmpty 没装过插件不是错误。
func TestLoadMissingRegistryIsEmpty(t *testing.T) {
	s := newStore(t)
	reg, err := s.Load()
	if err != nil {
		t.Fatalf("空注册表不该报错: %v", err)
	}
	if len(reg.Plugins) != 0 || reg.SchemaVersion != RegistrySchemaVersion {
		t.Fatalf("空注册表内容不对: %+v", reg)
	}
}

// TestSaveLoadRoundTripAndBackup 写读一致，且第二次写会留 .bak。
func TestSaveLoadRoundTripAndBackup(t *testing.T) {
	s := newStore(t)
	reg := &Registry{SchemaVersion: RegistrySchemaVersion}
	reg.Put(&Record{ID: "com.example.b", ActiveVersion: "0.2.0", Enabled: true})
	reg.Put(&Record{ID: "com.example.a", ActiveVersion: "0.1.0"})
	if err := s.Save(reg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.RegistryPath()); err != nil {
		t.Fatalf("注册表没写出来: %v", err)
	}
	// 再写一次 → .bak 出现。
	if err := s.Save(reg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.RegistryPath() + ".bak"); err != nil {
		t.Fatalf("第二次写该留备份: %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Plugins) != 2 {
		t.Fatalf("插件数 = %d", len(got.Plugins))
	}
	// Save 按 ID 排序，输出才稳定。
	if got.Plugins[0].ID != "com.example.a" {
		t.Errorf("该按 ID 排序，得 %s", got.Plugins[0].ID)
	}
}

// TestLoadRecoversFromBackup 注册表损坏时从 .bak 恢复并把损坏文件留档。
func TestLoadRecoversFromBackup(t *testing.T) {
	s := newStore(t)
	reg := &Registry{SchemaVersion: RegistrySchemaVersion}
	reg.Put(&Record{ID: "com.example.a"})
	if err := s.Save(reg); err != nil {
		t.Fatal(err)
	}
	bak, err := os.ReadFile(s.RegistryPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.RegistryPath()+".bak", bak, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.RegistryPath(), []byte("{ 这不是 json"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("有备份时该能恢复: %v", err)
	}
	if got.RecoveredFrom == "" {
		t.Error("该标记 RecoveredFrom")
	}
	if len(got.Plugins) != 1 {
		t.Fatalf("恢复后插件数 = %d", len(got.Plugins))
	}
	// 损坏文件被留档（不是被悄悄删掉）。
	entries, _ := os.ReadDir(filepath.Dir(s.RegistryPath()))
	foundAside := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".corrupt-") {
			foundAside = true
		}
	}
	if !foundAside {
		t.Error("损坏的注册表该留档")
	}
}

// TestLoadCorruptWithoutBackupFailsLoudly 备份也没有时必须报错，不能当成"没装插件"。
func TestLoadCorruptWithoutBackupFailsLoudly(t *testing.T) {
	s := newStore(t)
	if err := os.MkdirAll(filepath.Dir(s.RegistryPath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.RegistryPath(), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); err == nil {
		t.Fatal("无备份的损坏注册表该报错")
	}
}

// TestLoadRejectsNewerSchema 更高主版本拒绝读（不猜字段含义）。
func TestLoadRejectsNewerSchema(t *testing.T) {
	s := newStore(t)
	if err := os.MkdirAll(filepath.Dir(s.RegistryPath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.RegistryPath(),
		[]byte(fmt.Sprintf(`{"schemaVersion": %d, "plugins": []}`, RegistrySchemaVersion+1)), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := s.Load()
	if err == nil || !strings.Contains(err.Error(), "schemaVersion") {
		t.Fatalf("该拒绝更高版本，得 %v", err)
	}
}

// TestWithLockSerializesWriters 并发写入不能互相覆盖。
//
// 这是 CLI 与 Web 同时操作注册表时唯一的安全保证；没有它，典型后果是
// "刚装的插件被另一侧的旧快照覆盖掉，而两边都报成功"。
func TestWithLockSerializesWriters(t *testing.T) {
	s := newStore(t)
	const n = 12
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := s.WithLock(func(reg *Registry) error {
				reg.Put(&Record{ID: fmt.Sprintf("com.example.p%02d", i), ActiveVersion: "0.1.0"})
				return nil
			})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("并发写入失败: %v", err)
		}
	}
	reg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Plugins) != n {
		t.Fatalf("该有 %d 条记录，得 %d（说明有覆盖）", n, len(reg.Plugins))
	}
}

// TestStaleLockIsBroken 陈旧锁会被清掉：一次崩溃不该把插件管理永久钉死。
func TestStaleLockIsBroken(t *testing.T) {
	s := newStore(t)
	s.lockWait = 300 * time.Millisecond
	if err := os.MkdirAll(filepath.Dir(s.RegistryPath()), 0o755); err != nil {
		t.Fatal(err)
	}
	lock := s.RegistryPath() + ".lock"
	if err := os.WriteFile(lock, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-10 * time.Minute)
	if err := os.Chtimes(lock, old, old); err != nil {
		t.Fatal(err)
	}
	err := s.WithLock(func(reg *Registry) error {
		reg.Put(&Record{ID: "com.example.a"})
		return nil
	})
	if err != nil {
		t.Fatalf("陈旧锁该被清掉: %v", err)
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Error("锁文件该在离开时被删掉")
	}
}

// TestFreshLockTimesOut 新鲜的锁会让写入等待后失败（而不是硬闯）。
func TestFreshLockTimesOut(t *testing.T) {
	s := newStore(t)
	s.lockWait = 100 * time.Millisecond
	if err := os.MkdirAll(filepath.Dir(s.RegistryPath()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.RegistryPath()+".lock", []byte("held"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := s.WithLock(func(*Registry) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "占用") {
		t.Fatalf("该报锁占用，得 %v", err)
	}
	// 别人持有的锁不能被我们删掉。
	if _, err := os.Stat(s.RegistryPath() + ".lock"); err != nil {
		t.Error("不该删别人的锁")
	}
}

// TestNamespaceAndRouteUsage 冲突检查只看已启用的插件。
func TestNamespaceAndRouteUsage(t *testing.T) {
	reg := &Registry{}
	reg.Put(&Record{ID: "a", Enabled: true, Contributions: Contributions{
		CLI:      &CLIContribution{Namespace: "ns-a"},
		Frontend: &FrontendContribution{Route: "route-a"},
	}})
	reg.Put(&Record{ID: "b", Enabled: false, Contributions: Contributions{
		CLI:      &CLIContribution{Namespace: "ns-b"},
		Frontend: &FrontendContribution{Route: "route-b"},
	}})
	if got := reg.NamespacesInUse(""); len(got) != 1 || got[0] != "ns-a" {
		t.Fatalf("该只统计已启用，得 %v", got)
	}
	if got := reg.NamespacesInUse("a"); len(got) != 0 {
		t.Fatalf("排除自身后该为空，得 %v", got)
	}
	if got := reg.RoutesInUse(""); len(got) != 1 || got[0] != "route-a" {
		t.Fatalf("路由统计不对: %v", got)
	}
}

// TestSummarizeKeepsGrantsSeparate 展示投影里"声明权限"与"已授权"不混。
func TestSummarizeKeepsGrantsSeparate(t *testing.T) {
	rec := &Record{
		ID: "com.example.a", Name: "A", ActiveVersion: "1.0.0", Enabled: true,
		Permissions: []string{"db.connect", "profiles.list"},
		Grants:      Grants{Capabilities: []string{"profiles.list"}, Envs: []string{"prod"}},
		Contributions: Contributions{
			CLI:      &CLIContribution{Namespace: "a", Summary: "示例"},
			Frontend: &FrontendContribution{Route: "a", Title: "A 页面"},
			Skills:   []string{"skills/x"},
		},
	}
	s := Summarize(rec)
	if strings.Join(s.Permissions, ",") != "db.connect,profiles.list" {
		t.Errorf("声明权限 = %v", s.Permissions)
	}
	if strings.Join(s.Capabilities, ",") != "profiles.list" {
		t.Errorf("已授权 = %v", s.Capabilities)
	}
	if s.Namespace != "a" || s.Route != "a" || len(s.SkillDirs) != 1 {
		t.Errorf("贡献投影不对: %+v", s)
	}
	if line := ContributionSummaryString(rec.Contributions); !strings.Contains(line, "cli:a") || !strings.Contains(line, "ui:a") {
		t.Errorf("贡献摘要不对: %q", line)
	}
}
