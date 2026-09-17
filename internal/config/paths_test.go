package config

import (
	"os"
	"path/filepath"
	"testing"
)

// isolateEnv 把配置发现相关的环境变量全部指到临时目录，让测试不受本机
// 真实存在的 %APPDATA%\T100\{tdebug,tdict} 与 %APPDATA%\TDebug 影响。
// 返回统一用户目录（T100_HOME）。
func isolateEnv(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("T100_HOME", home)
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("TT_CONFIG", "")
	t.Setenv("TDEBUG_CONFIG", "")
	t.Setenv("TDICT_CONFIG", "")
	return home
}

// T100_HOME 整体改写统一目录。
func TestToolsHome_T100HomeOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("T100_HOME", dir)

	if got := ToolsHome(); got != dir {
		t.Errorf("ToolsHome() = %q, 期望 %q", got, dir)
	}
	want := filepath.Join(dir, ToolDirName)
	if got := UserConfigDir(); got != want {
		t.Errorf("UserConfigDir() = %q, 期望 %q", got, want)
	}
	if got := UserConfigPath(); got != filepath.Join(want, DefaultConfigName) {
		t.Errorf("UserConfigPath() = %q", got)
	}
}

// 没有 T100_HOME 时落在 %APPDATA%\T100 下。
func TestToolsHome_DefaultsUnderUserConfigDir(t *testing.T) {
	t.Setenv("T100_HOME", "")
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skip("拿不到用户配置目录")
	}
	if got, want := ToolsHome(), filepath.Join(base, "T100"); got != want {
		t.Errorf("ToolsHome() = %q, 期望 %q", got, want)
	}
}

// TT_CONFIG 优先级最高，且旧名 TDEBUG_CONFIG / TDICT_CONFIG 仍然生效 ——
// 既有脚本和 AI skill 不改也能用。
func TestResolvePath_EnvVars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "my.json")
	if err := os.WriteFile(path, []byte(`{"hosts":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, env := range []string{"TT_CONFIG", "TDEBUG_CONFIG", "TDICT_CONFIG"} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("TT_CONFIG", "")
			t.Setenv("TDEBUG_CONFIG", "")
			t.Setenv("TDICT_CONFIG", "")
			t.Setenv(env, path)

			got, err := ResolvePath("", false)
			if err != nil {
				t.Fatalf("ResolvePath: %v", err)
			}
			if got != path {
				t.Errorf("解析到 %q, 期望 %q", got, path)
			}
		})
	}

	// TT_CONFIG 应当压过另外两个旧名
	t.Setenv("TT_CONFIG", path)
	t.Setenv("TDEBUG_CONFIG", filepath.Join(dir, "不存在.json"))
	got, err := ResolvePath("", false)
	if err != nil || got != path {
		t.Errorf("TT_CONFIG 未压过旧名: got=%q err=%v", got, err)
	}
}

// --config 显式指定：绝对路径直接用，相对路径按 exe 目录 / 当前目录找。
func TestResolvePath_ExplicitFlag(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	isolateEnv(t)

	rel := "custom.json"
	if err := os.WriteFile(filepath.Join(dir, rel), []byte(`{"hosts":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ResolvePath(rel, false)
	if err != nil {
		t.Fatalf("相对路径未按当前目录解析: %v", err)
	}
	if filepath.Base(got) != rel {
		t.Errorf("解析到 %q", got)
	}
}

// 全部落空时报错，错误里要含尝试过的路径与设置环境变量的提示。
func TestResolvePath_NoneFound(t *testing.T) {
	isolateEnv(t)

	_, err := ResolvePath("", false)
	if err == nil {
		t.Fatal("找不到配置时应当报错")
	}
	msg := err.Error()
	for _, want := range []string{"配置文件未找到", "TT_CONFIG", "尝试了以下路径"} {
		if !contains(msg, want) {
			t.Errorf("错误信息缺少 %q:\n%s", want, msg)
		}
	}

	// allowMissing 用于写路径：首次保存要能拿到落点而不是报错
	got, err := ResolvePath("", true)
	if err != nil {
		t.Fatalf("allowMissing=true 不应报错: %v", err)
	}
	if got != UserConfigPath() {
		t.Errorf("落点 = %q, 期望 %q", got, UserConfigPath())
	}
}

// 端到端迁移：旧工具目录下两份配置 → 合并成 %T100_HOME%\tt\config.json。
func TestMigrate_EndToEnd(t *testing.T) {
	home := isolateEnv(t)

	write := func(tool, content string) {
		dir := filepath.Join(home, tool)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, DefaultConfigName), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("tdebug", tdebugStyle)
	write("tdict", tdictStyle)

	dst, err := ResolvePath("", false)
	if err != nil {
		t.Fatalf("ResolvePath: %v", err)
	}
	if dst != UserConfigPath() {
		t.Fatalf("落点 = %q, 期望 %q", dst, UserConfigPath())
	}

	r, err := Load(dst)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if r.SchemaVersion != SchemaVersion {
		t.Errorf("迁移后 schemaVersion = %d", r.SchemaVersion)
	}
	if len(r.Hosts.SSHs) != 2 {
		t.Errorf("合并后环境数 = %d, 期望 2", len(r.Hosts.SSHs))
	}
	if r.Debug.WatchdogSeconds != 240 {
		t.Errorf("调试设置丢失: %+v", r.Debug)
	}
	if r.Query.Source != "local" {
		t.Errorf("字典设置丢失: %+v", r.Query)
	}

	// 源文件不删除，且各留一份备份
	for _, tool := range []string{"tdebug", "tdict"} {
		src := filepath.Join(home, tool, DefaultConfigName)
		if _, err := os.Stat(src); err != nil {
			t.Errorf("源配置被删除了: %s", src)
		}
		if _, err := os.Stat(src + ".pre-merge.bak"); err != nil {
			t.Errorf("未留下备份: %s", src+".pre-merge.bak")
		}
	}
}

// 迁移是幂等的：第二次调用不该再动任何东西。
func TestMigrate_Idempotent(t *testing.T) {
	home := isolateEnv(t)

	dir := filepath.Join(home, "tdebug")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, DefaultConfigName), []byte(tdebugStyle), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := ResolvePath("", false); err != nil {
		t.Fatalf("首次解析: %v", err)
	}
	before, err := os.ReadFile(UserConfigPath())
	if err != nil {
		t.Fatal(err)
	}

	// 第二次：目标已是当前结构，Migrate 应当直接放弃
	if got := Migrate(); got != nil {
		t.Errorf("第二次迁移本应放弃, 却报告了来源 %v", got)
	}
	after, err := os.ReadFile(UserConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("第二次运行改动了配置内容")
	}
}

// 目标已是当前结构时不可被旧配置覆盖。
func TestPlanMigration_CurrentWins(t *testing.T) {
	home := isolateEnv(t)

	dst := filepath.Join(home, ToolDirName, DefaultConfigName)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte(`{"schemaVersion":2,"hosts":{"activeEnv":"现役","sshs":[{"name":"现役"}]}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	// 同时放一份旧配置在旁边
	old := filepath.Join(home, "tdict")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, DefaultConfigName), []byte(tdictStyle), 0o600); err != nil {
		t.Fatal(err)
	}

	p, err := PlanMigration(dst)
	if err != nil {
		t.Fatalf("PlanMigration: %v", err)
	}
	if p != nil {
		t.Error("目标已是当前结构，不该再生出迁移计划")
	}
}

// LooksLikeOwnConfig 必须挡住无关的 config.json，避免误迁。
func TestCollectSources_IgnoresForeignConfig(t *testing.T) {
	home := isolateEnv(t)

	// 一个长得像 Node 项目的 config.json 放在旧工具目录里
	foreign := filepath.Join(home, "tdict")
	if err := os.MkdirAll(foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(foreign, DefaultConfigName),
		[]byte(`{"name":"my-app","dependencies":{"react":"^18"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	srcs := collectSources(filepath.Join(home, ToolDirName, DefaultConfigName))
	for _, s := range srcs {
		if s.Path == filepath.Join(foreign, DefaultConfigName) {
			t.Error("无关的 config.json 被当作可迁移来源")
		}
	}
}

// 显式指定的路径即使文件还不存在，也必须原样用 —— 那是"在这里新建"，不是"没找到"。
// 写偏了会让便携版把配置写到用户目录，也会让 --config <临时目录> 的调用落在别处。
func TestResolvePath_ExplicitWinsWhenMissing(t *testing.T) {
	isolateEnv(t)

	dir := t.TempDir()
	want := filepath.Join(dir, "sub", "config.json")

	got, err := ResolvePath(want, true)
	if err != nil {
		t.Fatalf("allowMissing 时不该报错: %v", err)
	}
	if got != want {
		t.Errorf("落点 = %q, 期望用户显式指定的 %q", got, want)
	}
	if got == DefaultConfigPath() {
		t.Error("退回了默认落点 —— 显式路径被忽略")
	}
}

// 环境变量指定的路径同理（TT_CONFIG 指向一个还不存在的文件）。
func TestResolvePath_EnvWinsWhenMissing(t *testing.T) {
	isolateEnv(t)

	want := filepath.Join(t.TempDir(), "from-env.json")
	t.Setenv("TT_CONFIG", want)

	got, err := ResolvePath("", true)
	if err != nil {
		t.Fatalf("allowMissing 时不该报错: %v", err)
	}
	if got != want {
		t.Errorf("落点 = %q, 期望 %q", got, want)
	}
}

// 便携包与 --config 显式指定都不会走"统一用户目录"那条迁移分支，
// 所以旧结构的配置必须在选定路径上就地迁移 —— 否则用户把上一版的 config.json
// 拷进新包后，环境清单会是空的。
func TestMigrateInPlace_OldSchema(t *testing.T) {
	home := isolateEnv(t)
	_ = home

	dir := t.TempDir()
	p := filepath.Join(dir, DefaultConfigName)
	if err := os.WriteFile(p, []byte(tdebugStyle), 0o600); err != nil {
		t.Fatal(err)
	}

	if got := migrateInPlace(p); got != p {
		t.Fatalf("migrateInPlace 返回 %q, 期望原路径", got)
	}

	// 迁移后必须是当前结构，且环境还在
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !isCurrentSchema(b) {
		t.Fatal("旧结构没有被迁移")
	}
	r, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(r.Hosts.SSHs) != 2 {
		t.Errorf("环境数 = %d, 期望 2（旧 debug.sshs 里的两个）", len(r.Hosts.SSHs))
	}
	if r.Debug.WatchdogSeconds != 240 {
		t.Errorf("调试设置丢失: %+v", r.Debug)
	}
	if r.Hosts.ByName("开发环境") == nil {
		t.Error("环境名丢失")
	}
}

// 已是当前结构的配置不该被就地迁移动到。
func TestMigrateInPlace_AlreadyCurrent(t *testing.T) {
	isolateEnv(t)

	dir := t.TempDir()
	p := filepath.Join(dir, DefaultConfigName)
	seed := `{"schemaVersion":2,"listen":"127.0.0.1:9999","hosts":{"activeEnv":"x","sshs":[{"name":"x","host":"h","user":"u"}]}}`
	if err := os.WriteFile(p, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(p)

	migrateInPlace(p)

	after, _ := os.ReadFile(p)
	if string(before) != string(after) {
		t.Error("已是当前结构的配置被改动了")
	}
}

// 文件不存在且没有可合并的旧配置时，migrateInPlace 不该凭空造文件 ——
// 首次运行的骨架由 EnsureExists 负责，那是调用方的决定。
func TestMigrateInPlace_NothingToDo(t *testing.T) {
	isolateEnv(t)

	p := filepath.Join(t.TempDir(), DefaultConfigName)
	migrateInPlace(p)
	if _, err := os.Stat(p); err == nil {
		t.Error("无可合并内容时不该创建文件")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
