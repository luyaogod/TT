package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// EditSection 只能动自己那一节，其他顶层键（含未知键）必须原样保留。
func TestEditSection_PreservesOtherKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	seed := `{
  "schemaVersion": 2,
  "listen": "127.0.0.1:28670",
  "hosts": {"activeEnv": "a", "sshs": [{"name": "a", "host": "h", "port": 22, "user": "u", "password": "p"}]},
  "query": {"source": "local"},
  "未来的键": {"x": 1}
}`
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}

	err := EditSection(path, "query", nil, func(sec map[string]any) error {
		sec["source"] = "生产环境"
		return nil
	})
	if err != nil {
		t.Fatalf("EditSection: %v", err)
	}

	root, _ := Open(path)
	if root["未来的事"] != nil {
		t.Error("不该出现这个键")
	}
	if _, has := root["未来的键"]; !has {
		t.Error("未知顶层键被抹掉了")
	}
	if root["listen"] != "127.0.0.1:28670" {
		t.Error("其他已知节被改动了")
	}
	q, _ := root["query"].(map[string]any)
	if q["source"] != "生产环境" {
		t.Errorf("目标节未写入: %v", q)
	}
	hosts, _ := root["hosts"].(map[string]any)
	if hosts["activeEnv"] != "a" {
		t.Error("hosts 节被误改")
	}
}

// EditHosts 读-改-写环境清单。
func TestEditHosts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// 文件不存在时也能写（首次运行路径）
	err := EditHosts(path, func(h *Hosts) error {
		h.SSHs = append(h.SSHs, NamedSsh{
			Name:      "新环境",
			SSHConfig: SSHConfig{Host: "h", Port: 22, User: "u", Password: "p"},
		})
		return nil
	})
	if err != nil {
		t.Fatalf("EditHosts: %v", err)
	}

	r, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(r.Hosts.SSHs) != 1 || r.Hosts.SSHs[0].Name != "新环境" {
		t.Fatalf("环境写入不正确: %+v", r.Hosts.SSHs)
	}
	if r.Hosts.ActiveEnv != "新环境" {
		t.Errorf("activeEnv 未兜底为首条: %q", r.Hosts.ActiveEnv)
	}
	if r.SchemaVersion != SchemaVersion {
		t.Errorf("写入未打上 schemaVersion: %d", r.SchemaVersion)
	}
}

// SaveHosts 整体替换 hosts 节，其他节不动。
func TestSaveHosts_ReplacesOnlyHosts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	seed := `{"schemaVersion":2,"hosts":{"sshs":[]},"query":{"source":"local"}}`
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}

	err := SaveHosts(path, Hosts{
		ActiveEnv: "e1",
		SSHs: []NamedSsh{{
			Name:      "e1",
			SSHConfig: SSHConfig{Host: "h", Port: 2222, User: "u", Password: "p"},
		}},
	})
	if err != nil {
		t.Fatalf("SaveHosts: %v", err)
	}

	root, _ := Open(path)
	q, _ := root["query"].(map[string]any)
	if q["source"] != "local" {
		t.Error("SaveHosts 误改了 query 节")
	}
	r, _ := Load(path)
	if len(r.Hosts.SSHs) != 1 || r.Hosts.SSHs[0].Port != 2222 {
		t.Errorf("hosts 未正确写入: %+v", r.Hosts)
	}
}

// Open 对缺失文件返回空配置，对非法 JSON 报错，对空文件返回空配置。
func TestOpen(t *testing.T) {
	dir := t.TempDir()

	root, err := Open(filepath.Join(dir, "missing.json"))
	if err != nil {
		t.Fatalf("缺失文件应当返回空配置而不是错误: %v", err)
	}
	if len(root) != 0 {
		t.Errorf("缺失文件应返回空 map, 得到 %v", root)
	}

	empty := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if root, err = Open(empty); err != nil || len(root) != 0 {
		t.Errorf("空文件应视为空配置: root=%v err=%v", root, err)
	}

	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{oops"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = Open(bad); err == nil {
		t.Error("非法 JSON 应当报错")
	}
}

// Save 必须是原子的：写完不留 .tmp 残渣，且保留原文件权限位。
func TestSave_AtomicAndKeepsMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := Save(path, map[string]any{"a": 1}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		names := []string{}
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("目录里应当只有一个文件, 实为 %v", names)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// Windows 的权限模型只有只读位，Go 的 os.Stat 一律报 0666/0444，
	// 断言 Unix 权限位没有意义；那里只验证文件可写（即没被写成只读）。
	if runtime.GOOS == "windows" {
		if fi.Mode().Perm()&0o200 == 0 {
			t.Errorf("新建配置不应是只读: %v", fi.Mode().Perm())
		}
	} else if fi.Mode().Perm() != 0o600 {
		t.Errorf("新建配置权限 = %v, 期望 0600（含明文口令）", fi.Mode().Perm())
	}
}

// 配置目录不存在时要能自动创建 —— 缺省位置在新机器上还没有。
func TestSave_CreatesMissingDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "T100", "tt", "config.json")
	if err := Save(path, map[string]any{"a": 1}); err != nil {
		t.Fatalf("Save 应当自动创建目录: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("文件未创建: %v", err)
	}
}

// LoadHosts 是读环境清单的入口（合并前 TDictCli/host.LoadHosts 的替代）。
// 它要求至少配了一个环境 —— CLI 命令都按"有环境可连"的前提工作。
//
// 合并前那两份 LoadHosts 测试里有一条断言"旧键 debug 也能读"，那条随旧结构一起
// 取消了：兼容性由 migrate.go 一次性做掉，读路径只认 hosts。
func TestLoadHosts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, DefaultConfigName)
	seed := `{
  "schemaVersion": 2,
  "hosts": {"activeEnv": "b", "sshs": [
    {"name": "a", "host": "1.1.1.1", "user": "u"},
    {"name": "b", "host": "2.2.2.2", "user": "u", "db": {"type": "oracle", "host": "d", "service": "t35prd"}}
  ]}
}`
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}

	h, err := LoadHosts(path)
	if err != nil {
		t.Fatalf("LoadHosts: %v", err)
	}
	if h.ActiveEnv != "b" {
		t.Errorf("ActiveEnv = %q, 期望 b", h.ActiveEnv)
	}
	if len(h.SSHs) != 2 {
		t.Fatalf("环境数 = %d, 期望 2", len(h.SSHs))
	}
	if got := h.ByName(""); got == nil || got.Name != "b" {
		t.Errorf("ByName(\"\") 应回退 activeEnv: %+v", got)
	}
	// db 要能正常解析出来
	e := h.ByName("b")
	if e.DB == nil || e.DB.Type != "oracle" || e.DB.Svc() != "t35prd" {
		t.Errorf("db 未正确解析: %+v", e.DB)
	}
}

func TestLoadHosts_Errors(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	// 没有 hosts 节
	if _, err := LoadHosts(write("a.json", `{"schemaVersion":2,"query":{"source":"local"}}`)); err == nil {
		t.Error("缺少 hosts 节应报错")
	}
	// 有节但一个环境都没有
	if _, err := LoadHosts(write("b.json", `{"schemaVersion":2,"hosts":{"sshs":[]}}`)); err == nil {
		t.Error("sshs 为空应报错")
	}
}

// 首次运行要能落一份骨架：配置页得有个文件可编辑。
func TestEnsureExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.json")

	if err := EnsureExists(path); err != nil {
		t.Fatalf("EnsureExists: %v", err)
	}
	r, err := Load(path)
	if err != nil {
		t.Fatalf("骨架应当可被读回: %v", err)
	}
	if r.SchemaVersion != SchemaVersion {
		t.Errorf("骨架的 schemaVersion = %d, 期望 %d（写对了才不会被就地迁移）", r.SchemaVersion, SchemaVersion)
	}
	if r.Listen != DefaultListen {
		t.Errorf("骨架的 listen = %q, 期望 %q", r.Listen, DefaultListen)
	}
	if len(r.Hosts.SSHs) != 0 {
		t.Errorf("骨架不该预置环境: %+v", r.Hosts.SSHs)
	}

	// 已存在时原样不动
	if err := Save(path, map[string]any{"schemaVersion": SchemaVersion, "listen": "127.0.0.1:9999", "hosts": map[string]any{"sshs": []any{}}}); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)
	if err := EnsureExists(path); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Error("EnsureExists 覆盖了已存在的配置")
	}
}

func TestHostsByName(t *testing.T) {
	h := &Hosts{
		ActiveEnv: "b",
		SSHs: []NamedSsh{
			{Name: "a"}, {Name: "b"}, {Name: "c"},
		},
	}
	if got := h.ByName("c"); got == nil || got.Name != "c" {
		t.Error("按名取环境失败")
	}
	if got := h.ByName(""); got == nil || got.Name != "b" {
		t.Error("空名应取 activeEnv")
	}
	if got := h.ByName("不存在"); got != nil {
		t.Error("未命中应返回 nil")
	}

	h.ActiveEnv = ""
	if got := h.ByName(""); got == nil || got.Name != "a" {
		t.Error("activeEnv 为空时应取首条")
	}
}

// TDev 原来完全没有配置系统；这个节是新增的，缺省值必须能兜住。
func TestTdevSettingsDefault(t *testing.T) {
	var s TdevSettings
	if got := s.WorkspaceSuffixOrDefault(); got != DefaultWorkspaceSuffix {
		t.Errorf("缺省后缀 = %q, 期望 %q", got, DefaultWorkspaceSuffix)
	}
	s.WorkspaceSuffix = "-work"
	if got := s.WorkspaceSuffixOrDefault(); got != "-work" {
		t.Errorf("显式后缀 = %q", got)
	}
}
