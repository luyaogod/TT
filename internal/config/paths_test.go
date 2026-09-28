package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// isolateEnv 把配置发现相关的环境变量全部指到临时目录，让测试不受本机
// 真实存在的 %APPDATA%\T100\{tdebug,tdict,tt} 与 %APPDATA%\TDebug 影响。
//
// 返回两个根 —— 它们**不再是同一个目录**（改落点之前是）：
//   - legacyRoot 旧产品根：tdebug/、tdict/ 以及"旧落点" tt/ 都种在这里
//   - dataDir    新数据目录（TT_HOME）：配置直接落在它下面
//
// 两个都要给：只给 legacyRoot 的话新落点会落回真实的 %APPDATA%\TT。
func isolateEnv(t *testing.T) (legacyRoot, dataDir string) {
	t.Helper()
	legacyRoot, dataDir = t.TempDir(), t.TempDir()
	t.Setenv("T100_HOME", legacyRoot)
	t.Setenv("TT_HOME", dataDir)
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("TT_CONFIG", "")
	t.Setenv("TDEBUG_CONFIG", "")
	t.Setenv("TDICT_CONFIG", "")
	return legacyRoot, dataDir
}

// TT_HOME 是数据目录本身。
func TestUserConfigDir_TTHomeOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TT_HOME", dir)

	if got := UserConfigDir(); got != dir {
		t.Errorf("UserConfigDir() = %q, 期望 %q", got, dir)
	}
	want := filepath.Join(dir, DefaultConfigName)
	if got := UserConfigPath(); got != want {
		t.Errorf("UserConfigPath() = %q, 期望 %q", got, want)
	}
	if got := DefaultConfigPathHint(); got != want {
		t.Errorf("DefaultConfigPathHint() = %q, 期望 %q", got, want)
	}
}

// 旧名 T100_HOME 仍被识别。
//
// **语义与改落点之前不同**：那时 T100_HOME=D:\x 指的是 D:\x\tt（产品根多一层），
// 现在指 D:\x 本身。所以设过它的机器会走一次迁移（来源 <T100_HOME>\tt\config.json）。
func TestUserConfigDir_T100HomeStillWorks(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TT_HOME", "")
	t.Setenv("T100_HOME", dir)

	if got := UserConfigDir(); got != dir {
		t.Errorf("UserConfigDir() = %q, 期望旧名指到的 %q", got, dir)
	}
}

// 两个都设时 TT_HOME 优先。
func TestUserConfigDir_TTHomeBeatsT100Home(t *testing.T) {
	tt, legacy := t.TempDir(), t.TempDir()
	t.Setenv("TT_HOME", tt)
	t.Setenv("T100_HOME", legacy)

	if got := UserConfigDir(); got != tt {
		t.Errorf("UserConfigDir() = %q, 期望 TT_HOME 的 %q", got, tt)
	}
}

// 两个环境变量都不设时落在 %APPDATA%\TT —— 不再多一层 T100。
func TestUserConfigDir_DefaultsToAppDataTT(t *testing.T) {
	t.Setenv("TT_HOME", "")
	t.Setenv("T100_HOME", "")
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skip("拿不到用户配置目录")
	}
	if got, want := UserConfigDir(), filepath.Join(base, "TT"); got != want {
		t.Errorf("UserConfigDir() = %q, 期望 %q", got, want)
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

// 端到端迁移：旧工具目录下两份配置 → 合并成新落点上的 config.json。
func TestMigrate_EndToEnd(t *testing.T) {
	legacyRoot, _ := isolateEnv(t)

	write := func(tool, content string) {
		dir := filepath.Join(legacyRoot, tool)
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
		src := filepath.Join(legacyRoot, tool, DefaultConfigName)
		if _, err := os.Stat(src); err != nil {
			t.Errorf("源配置被删除了: %s", src)
		}
		if _, err := os.Stat(src + ".pre-merge.bak"); err != nil {
			t.Errorf("未留下备份: %s", src+".pre-merge.bak")
		}
	}
}

// 改落点的核心用例：旧落点（<旧产品根>\tt\config.json，本机即 %APPDATA%\T100\tt\config.json）
// 那份**现役**配置必须被搬到新落点，而且一个键都不能少。
//
// 为什么"一个键都不能少"是重点：旧落点是 0.2.0 的现役配置，已经是 SchemaVersion 2，
// 里面 tzs.workspace 这类键没有缺省值。**它该被搬运，不该被合并** —— MergeConfigs 只认识
// hosts/debug/listen/query/mirror/bdldoc/sync/tdev，走合并会把 tzs 整节丢掉，
// 用户升级后 `tt dev tzs` 直接不能跑。
//
// 同时放一份 tdict 在旧工具目录里，是为了把"搬运"与"合并"两种结果区分开：
// 若走的是合并，mirror 节会冒出来、环境会变成两个、tzs 会消失。
func TestMigrate_FromOldDefaultLocation(t *testing.T) {
	legacyRoot, dataDir := isolateEnv(t)

	oldDir := filepath.Join(legacyRoot, ToolDirName)
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(oldDir, DefaultConfigName)
	seed := `{
  "schemaVersion": 2,
  "listen": "127.0.0.1:8899",
  "hosts": {"activeEnv": "现役", "sshs": [{"name": "现役", "host": "h", "user": "u"}]},
  "query": {"source": "local"},
  "tzs": {"workspace": "D:\\t100_wrok_dir", "serverExe": "D:\\tt\\out\\ttzs.exe"}
}`
	if err := os.WriteFile(old, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}
	// 旧工具目录里还躺着一份没被删掉的 tdict（仓库规矩：迁移不删源文件）
	if err := os.MkdirAll(filepath.Join(legacyRoot, "tdict"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacyRoot, "tdict", DefaultConfigName), []byte(tdictStyle), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := ResolvePath("", false)
	if err != nil {
		t.Fatalf("ResolvePath: %v", err)
	}
	if want := UserConfigPath(); got != want {
		t.Fatalf("落点 = %q，期望新落点 %q", got, want)
	}
	if got == old {
		t.Fatal("落点仍是旧位置 —— 没有搬家")
	}
	// 判据是"新路径上真的有文件"，不是"没报错"
	b, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("新落点上没有配置文件: %v", err)
	}
	if filepath.Dir(got) != dataDir {
		t.Errorf("落点目录 = %q，期望数据目录 %q", filepath.Dir(got), dataDir)
	}

	r, err := Load(got)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if r.Listen != "127.0.0.1:8899" {
		t.Errorf("listen = %q", r.Listen)
	}
	if r.Query.Source != "local" {
		t.Errorf("query.source = %q", r.Query.Source)
	}
	if len(r.Hosts.SSHs) != 1 || r.Hosts.SSHs[0].Name != "现役" {
		t.Errorf("环境清单 = %+v，期望只有现役那一个", r.Hosts.SSHs)
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	// tzs 整节原样带过来 —— 这正是"走搬运而不是合并"的判据
	tzs, _ := raw["tzs"].(map[string]any)
	if tzs == nil || tzs["workspace"] != "D:\\t100_wrok_dir" {
		t.Errorf("tzs 节丢失或被改写（MergeConfigs 不搬它）: %v", raw["tzs"])
	}
	// 反过来：旧工具那份的内容不该被并进来
	if _, has := raw["mirror"]; has {
		t.Error("tdict 的 mirror 节被并进来了 —— 目标是空的新落点，不该发生合并")
	}

	// 源文件不删除，且各留一份备份
	if _, err := os.Stat(old); err != nil {
		t.Errorf("旧落点的配置被删除了: %v", err)
	}
	if _, err := os.Stat(old + ".pre-merge.bak"); err != nil {
		t.Errorf("旧落点的配置没有留备份: %v", err)
	}
}

// 旧落点在 collectSources 里必须**紧邻 dst**（排在所有旧工具来源之后）。
//
// 这条顺序不是随手排的：MergeConfigs 里 query/mirror/bdldoc/sync/tdev 是"首个非空者
// 胜出"，而 hosts.sshs / debug / listen 是"靠后者胜出"。旧落点占的正是 dst 改落点之前
// 在顺序里的那一格，才能让旧结构配置的合并优先级与改落点之前逐条一致。
// （已经是当前结构的那份走的是原样搬运，不看这个位置 —— 见 migrationResult。）
func TestCollectSources_OldLocationSitsRightBeforeDst(t *testing.T) {
	legacyRoot, dataDir := isolateEnv(t)

	write := func(dir, content string) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, DefaultConfigName), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(legacyRoot, "tdebug"), tdebugStyle)
	write(filepath.Join(legacyRoot, "tdict"), tdictStyle)
	// 旧落点放一份**旧结构**的配置：当前结构的那份走搬运，不参与这段顺序
	write(filepath.Join(legacyRoot, ToolDirName), `{"query":{"source":"local"}}`)

	dst := filepath.Join(dataDir, DefaultConfigName) // 不存在，所以不会出现在来源里
	var got []string
	for _, s := range collectSources(dst) {
		got = append(got, filepath.Base(filepath.Dir(s.Path))+"/"+filepath.Base(s.Path))
	}
	want := []string{"tdebug/config.json", "tdict/config.json", ToolDirName + "/config.json"}
	if len(got) != len(want) {
		t.Fatalf("来源 = %v，期望 %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("来源顺序 = %v，期望 %v", got, want)
		}
	}
}

// 迁移是幂等的：第二次调用不该再动任何东西。
func TestMigrate_Idempotent(t *testing.T) {
	legacyRoot, _ := isolateEnv(t)

	dir := filepath.Join(legacyRoot, "tdebug")
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
	legacyRoot, dataDir := isolateEnv(t)

	dst := filepath.Join(dataDir, DefaultConfigName)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte(`{"schemaVersion":2,"hosts":{"activeEnv":"现役","sshs":[{"name":"现役"}]}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	// 同时放一份旧配置在旁边
	old := filepath.Join(legacyRoot, "tdict")
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
	legacyRoot, _ := isolateEnv(t)

	// 一个长得像 Node 项目的 config.json 放在旧工具目录里
	foreign := filepath.Join(legacyRoot, "tdict")
	if err := os.MkdirAll(foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(foreign, DefaultConfigName),
		[]byte(`{"name":"my-app","dependencies":{"react":"^18"}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	srcs := collectSources(filepath.Join(legacyRoot, ToolDirName, DefaultConfigName))
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
	isolateEnv(t)

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

// 不是配置的文件不能被就地迁移改写 —— 这是个凭据泄漏的口子，不只是一处不整洁。
//
// migrateInPlace 的职责是"把合并前的旧结构升级到当前结构"，而"旧结构"的前提是它**得先是一份
// 配置**。修之前，任何解不出 JSON 的文件都会一路掉进 PlanMigration 并被当成迁移目标，而来源里
// 包含 ToolsHome 下 legacyTools 的那几份配置（**含真实 SSH / 数据库口令**）。后果很具体：
// `TT_CONFIG=D:\tmp\scratch.json` 里有个笔误，你的口令就被复制到了那个临时文件里。
//
// 这个测试**必须先放一份可合并的旧配置**：没有来源时 plan.Sources 为空，谁都不会去写目标，
// 那样测了等于没测 —— 无论修没修都会绿。
func TestMigrateInPlace_CorruptFileIsNotATarget(t *testing.T) {
	legacyRoot, _ := isolateEnv(t)

	legacy := filepath.Join(legacyRoot, "tdict", DefaultConfigName)
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	seed := `{"hosts":{"activeEnv":"e","sshs":[{"name":"e","host":"h","user":"u","password":"SECRET"}]}}`
	if err := os.WriteFile(legacy, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	p := filepath.Join(dir, DefaultConfigName)
	const broken = "这不是配置，是写坏的 JSON {\n"
	if err := os.WriteFile(p, []byte(broken), 0o600); err != nil {
		t.Fatal(err)
	}

	if got := migrateInPlace(p); got != p {
		t.Fatalf("migrateInPlace 返回 %q，期望原路径", got)
	}

	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != broken {
		t.Errorf("不是配置的文件被迁移改写了：\n  现在 = %q\n  应该 = %q", string(b), broken)
	}
	if contains(string(b), "SECRET") {
		t.Error("旧配置里的口令被复制进了这个文件")
	}
}

// looksLikeConfig 与 isCurrentSchema 的 false 含义不同：前者是"不是配置"，后者是
// "是配置但结构旧"。只有后者该被迁移改写 —— 见上面那个测试。
func TestLooksLikeConfig(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{`{"hosts":{"sshs":[]}}`, true}, // 旧结构，但是配置
		{`{"schemaVersion":2}`, true},   // 当前结构
		{``, false},                     // 空文件
		{`[]`, false},                   // 是 JSON，但不是对象
		{`"x"`, false},                  // 同上
		{`{`, false},                    // 坏 JSON
		{"这不是 JSON", false},             // 根本不是
	}
	for _, c := range cases {
		if got := looksLikeConfig([]byte(c.in)); got != c.want {
			t.Errorf("looksLikeConfig(%q) = %v，期望 %v", c.in, got, c.want)
		}
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
