package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newManager 造一个管理面（保留名与内置命令的一个子集一致）。
func newManager(t *testing.T) (*Manager, *Store) {
	t.Helper()
	s := newStore(t)
	m := NewManager(s, []string{"serve", "plugin", "help", "dict", "db"})
	m.Envs = func() ([]string, error) { return []string{"prod", "test"}, nil }
	m.HostAPIVersion = "1.0.0"
	return m, s
}

// hooks 记录调用顺序的钩子。
type hooks struct {
	calls      []string
	enableErr  error
	disableErr error
}

func (h *hooks) AfterEnable(*Record) error {
	h.calls = append(h.calls, "after-enable")
	return h.enableErr
}
func (h *hooks) BeforeDisable(*Record) error {
	h.calls = append(h.calls, "before-disable")
	return h.disableErr
}
func (h *hooks) AfterDisable(*Record) error {
	h.calls = append(h.calls, "after-disable")
	return nil
}
func (h *hooks) AfterUninstall(string) { h.calls = append(h.calls, "after-uninstall") }

// TestInstallRegistersDisabled 安装只解包注册：不启用、不执行插件代码。
func TestInstallRegistersDisabled(t *testing.T) {
	m, s := newManager(t)
	src := writePluginDir(t)

	rec, err := m.Install(InstallOptions{Source: src})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Enabled {
		t.Error("新装的插件必须未启用")
	}
	if rec.ActiveVersion != "0.1.0" || rec.ID != "com.example.demo" {
		t.Fatalf("记录不对: %+v", rec)
	}
	dir := s.VersionDir(rec.ID, rec.ActiveVersion)
	if _, err := os.Stat(filepath.Join(dir, "bin", "demo.bin")); err != nil {
		t.Fatalf("载荷没落进版本目录: %v", err)
	}
	// 缓存里的暂存目录要清掉（不留垃圾）。
	entries, _ := os.ReadDir(s.CacheDir())
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "stage-") {
			t.Errorf("暂存目录没清: %s", e.Name())
		}
	}
	// 注册表里也查得到。
	reg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if reg.Find("com.example.demo") == nil {
		t.Error("注册表里没有记录")
	}
}

// TestInstallSameVersionRequiresReplace 同版本重装要显式 --replace。
func TestInstallSameVersionRequiresReplace(t *testing.T) {
	m, _ := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Install(InstallOptions{Source: src}); err == nil {
		t.Fatal("同版本重装该被拒")
	} else if !strings.Contains(err.Error(), "--replace") {
		t.Fatalf("错误该提示 --replace，得 %v", err)
	}
	if _, err := m.Install(InstallOptions{Source: src, Replace: true}); err != nil {
		t.Fatalf("--replace 该能重装: %v", err)
	}
}

// TestInstallRejectsReservedNamespace 撞内置命令的命名空间装不进去。
func TestInstallRejectsReservedNamespace(t *testing.T) {
	m, _ := newManager(t)
	src := writePluginDir(t)
	// 把 namespace 改成 serve（内置命令）。
	writeFile(t, filepath.Join(src, ManifestFileName), `{
  "schemaVersion": 1, "id": "com.example.demo", "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "contributions": {"cli": {"namespace": "serve", "entry": "bin/demo.bin"}}
}`)
	if _, err := m.Install(InstallOptions{Source: src}); err == nil {
		t.Fatal("命名空间与内置命令冲突该被拒")
	} else if !strings.Contains(err.Error(), "冲突") {
		t.Fatalf("错误该说明冲突，得 %v", err)
	}
}

// TestInstallRejectsNamespaceTakenByEnabledPlugin 两个已启用插件不能抢同一个命名空间。
func TestInstallRejectsNamespaceTakenByEnabledPlugin(t *testing.T) {
	m, s := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable("com.example.demo", Grants{Capabilities: []string{"profiles.list"}}); err != nil {
		t.Fatal(err)
	}

	// 另一个 id、同样的 namespace。
	src2 := t.TempDir()
	writeFile(t, filepath.Join(src2, ManifestFileName), `{
  "schemaVersion": 1, "id": "com.example.other", "version": "0.1.0",
  "hostApi": ">=1.0.0 <2.0.0",
  "contributions": {"cli": {"namespace": "demo", "entry": "bin/x.bin"}}
}`)
	writeFile(t, filepath.Join(src2, "bin", "x.bin"), "x")
	if _, err := m.Install(InstallOptions{Source: src2}); err == nil {
		t.Fatal("与已启用插件的命名空间冲突该被拒")
	} else if !strings.Contains(err.Error(), "已被其他已启用插件占用") {
		t.Fatalf("错误该说明占用，得 %v", err)
	}
	_ = s
}

// TestEnableGrantAndDisable 启用要显式授权；停用状态可来回切。
func TestEnableGrantAndDisable(t *testing.T) {
	m, _ := newManager(t)
	h := &hooks{}
	m.Hooks = h
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}

	rec, err := m.Enable("com.example.demo", Grants{
		Capabilities: []string{"db.connect", "profiles.list"},
		Envs:         []string{"prod"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !rec.Enabled {
		t.Fatal("该已启用")
	}
	if !rec.Grants.Has("db.connect") || !rec.Grants.AllowsEnv("prod") {
		t.Fatalf("授权没落盘: %+v", rec.Grants)
	}
	if len(h.calls) != 1 || h.calls[0] != "after-enable" {
		t.Fatalf("钩子调用不对: %v", h.calls)
	}

	rec, err = m.Disable("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Enabled {
		t.Fatal("该已停用")
	}
	// 停用保留授权：重新启用时用户不必再填一遍。
	if !rec.Grants.Has("db.connect") {
		t.Error("停用不该清掉授权")
	}
	want := []string{"after-enable", "before-disable", "after-disable"}
	if strings.Join(h.calls, ",") != strings.Join(want, ",") {
		t.Fatalf("钩子顺序不对: %v", h.calls)
	}
}

// TestEnableValidatesGrants 授权要校验：未知能力、不存在的环境、只给环境不给能力。
func TestEnableValidatesGrants(t *testing.T) {
	m, _ := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}

	if _, err := m.Enable("com.example.demo", Grants{Capabilities: []string{"root.access"}}); err == nil {
		t.Error("未知能力该被拒")
	}
	if _, err := m.Enable("com.example.demo", Grants{Capabilities: []string{"db.connect"}, Envs: []string{"ghost"}}); err == nil {
		t.Error("不存在的环境该被拒")
	} else if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("错误该点名环境，得 %v", err)
	}
	if _, err := m.Enable("com.example.demo", Grants{Envs: []string{"prod"}}); err == nil {
		t.Error("只给环境不给连接能力该被拒")
	}
	// 全部被拒之后仍然是未启用。
	reg, _ := m.Store.Load()
	if reg.Find("com.example.demo").Enabled {
		t.Error("校验失败不该改变启用状态")
	}
}

// TestEnableFailsWhenManifestBroken 版本目录被改坏后启用要失败并留下原因。
func TestEnableFailsWhenManifestBroken(t *testing.T) {
	m, s := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	dir := s.VersionDir("com.example.demo", "0.1.0")
	if err := os.Remove(filepath.Join(dir, "bin", "demo.bin")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable("com.example.demo", Grants{}); err == nil {
		t.Fatal("入口不见了该启用失败")
	}
	reg, _ := m.Store.Load()
	rec := reg.Find("com.example.demo")
	if rec.Enabled {
		t.Error("失败后不该是启用状态")
	}
	if rec.LastError == "" || !strings.Contains(rec.LastError, "cli.entry") {
		t.Errorf("该把原因记在 LastError，得 %q", rec.LastError)
	}
}

// TestEnableRollsBackWhenHookFails 启用后置动作失败要回退成未启用。
func TestEnableRollsBackWhenHookFails(t *testing.T) {
	m, _ := newManager(t)
	m.Hooks = &hooks{enableErr: errHook}
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable("com.example.demo", Grants{}); err == nil {
		t.Fatal("钩子失败该让启用失败")
	}
	reg, _ := m.Store.Load()
	rec := reg.Find("com.example.demo")
	if rec.Enabled {
		t.Error("必须回退成未启用")
	}
	if !strings.Contains(rec.LastError, "钩子") {
		t.Errorf("该记录原因，得 %q", rec.LastError)
	}
}

// errHook 钩子的测试错误。
var errHook = &testErr{"钩子失败：测试用"}

type testErr struct{ s string }

func (e *testErr) Error() string { return e.s }

// TestUninstallRequiresDisabledAndKeepsData 卸载要求先停用；默认保留持久数据。
func TestUninstallRequiresDisabledAndKeepsData(t *testing.T) {
	m, s := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Enable("com.example.demo", Grants{}); err != nil {
		t.Fatal(err)
	}
	if err := m.Uninstall("com.example.demo", UninstallOptions{}); err == nil {
		t.Fatal("启用中该拒绝卸载")
	}

	// 造一份持久数据。
	dataDir := s.DataDir("com.example.demo")
	writeFile(t, filepath.Join(dataDir, "keep.txt"), "用户的数据")

	if _, err := m.Disable("com.example.demo"); err != nil {
		t.Fatal(err)
	}
	if err := m.Uninstall("com.example.demo", UninstallOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.VersionDir("com.example.demo", "0.1.0")); !os.IsNotExist(err) {
		t.Error("版本目录该被删除")
	}
	if _, err := os.Stat(filepath.Join(dataDir, "keep.txt")); err != nil {
		t.Error("默认卸载必须保留持久数据")
	}
	reg, _ := m.Store.Load()
	if reg.Find("com.example.demo") != nil {
		t.Error("注册项该被移除")
	}
}

// TestUninstallPurgeData 显式要求时才删数据。
func TestUninstallPurgeData(t *testing.T) {
	m, s := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	dataDir := s.DataDir("com.example.demo")
	writeFile(t, filepath.Join(dataDir, "keep.txt"), "x")
	if err := m.Uninstall("com.example.demo", UninstallOptions{PurgeData: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		t.Error("--purge-data 该删掉数据目录")
	}
}

// TestDoctorFindsProblems 自检要能发现"注册表有、磁盘没有"这类不一致。
func TestDoctorFindsProblems(t *testing.T) {
	m, s := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	issues, err := m.Doctor("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("健康插件不该有问题: %+v", issues)
	}

	if err := os.RemoveAll(s.VersionDir("com.example.demo", "0.1.0")); err != nil {
		t.Fatal(err)
	}
	issues, err = m.Doctor("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) == 0 || issues[0].Code != "VERSION_DIR_MISSING" {
		t.Fatalf("该发现版本目录缺失，得 %+v", issues)
	}
}

// TestFindByNamespaceAndRoute 转发与页面服务都按命名空间/路由反查。
func TestFindByNamespaceAndRoute(t *testing.T) {
	m, _ := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	// 未启用时查不到（禁用插件不出现在帮助与菜单里）。
	if rec, err := m.FindByNamespace("demo"); err != nil || rec != nil {
		t.Fatalf("未启用时不该查到: %v %v", rec, err)
	}
	if _, err := m.Enable("com.example.demo", Grants{}); err != nil {
		t.Fatal(err)
	}
	rec, err := m.FindByNamespace("demo")
	if err != nil || rec == nil || rec.ID != "com.example.demo" {
		t.Fatalf("该按命名空间查到: %v %v", rec, err)
	}
	byRoute, err := m.FindByRoute("demo")
	if err != nil || byRoute == nil || byRoute.ID != "com.example.demo" {
		t.Fatalf("该按路由查到: %v %v", byRoute, err)
	}
	if rec, _ := m.FindByNamespace("nope"); rec != nil {
		t.Error("不存在的命名空间该返回 nil")
	}
}

// TestListAndInfo 列表与详情的数据来源一致。
func TestListAndInfo(t *testing.T) {
	m, _ := newManager(t)
	src := writePluginDir(t)
	if _, err := m.Install(InstallOptions{Source: src}); err != nil {
		t.Fatal(err)
	}
	list, err := m.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != "com.example.demo" {
		t.Fatalf("列表不对: %+v", list)
	}
	info, err := m.Info("com.example.demo")
	if err != nil {
		t.Fatal(err)
	}
	if info.Manifest == nil || info.Manifest.ID != "com.example.demo" {
		t.Fatalf("详情该带清单: %+v", info)
	}
	if info.Paths.VersionDir == "" || info.Paths.DataDir == "" {
		t.Error("详情该给出落点")
	}
	if _, err := m.Info("nope"); err == nil {
		t.Error("不存在的插件该报错")
	}
}

// TestInstallRejectsZipEscapeSource 目录来源里的符号链接不被跟随（包不能把洞带进来）。
func TestInstallRejectsSymlinkInDirSource(t *testing.T) {
	m, _ := newManager(t)
	src := writePluginDir(t)
	link := filepath.Join(src, "bin", "link")
	if err := os.Symlink(filepath.Join(src, "bin", "demo.bin"), link); err != nil {
		t.Skipf("本机不支持创建符号链接: %v", err)
	}
	if _, err := m.Install(InstallOptions{Source: src}); err == nil {
		t.Fatal("含符号链接的包该被拒")
	}
}
