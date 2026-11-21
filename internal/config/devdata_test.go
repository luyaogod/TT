package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDevDataLayout 钉住目录布局（plan §5.1）：一个环境一棵树，dict/mdl/prog 各就各位，
// 且与读侧的 <data>/dict 不共用。
func TestDevDataLayout(t *testing.T) {
	data := t.TempDir()
	env := "福联测试tst"
	envDir := DevDataEnvDir(data, env)
	if want := filepath.Join(data, "devdata", env); envDir != want {
		t.Fatalf("环境目录 = %q, want %q", envDir, want)
	}
	if got, want := DevDataProgDir(data, env, "axmt500_wf"), filepath.Join(envDir, "prog", "axmt500_wf"); got != want {
		t.Fatalf("程序目录 = %q, want %q", got, want)
	}
	if got, want := DevDataDictPath(data, env), filepath.Join(envDir, "dict", env+".db"); got != want {
		t.Fatalf("规格库 = %q, want %q", got, want)
	}
	if got, want := DevDataMdlDir(data, env), filepath.Join(envDir, "mdl"); got != want {
		t.Fatalf("模板树 = %q, want %q", got, want)
	}
	// 与读侧字典库必须不是同一个文件：同环境两个 .db 是**故意**的，混用就等于把
	// "给 model 的原材料"和"给查询用的镜像"捆在一起，改一边影响另一边。
	if DevDataDictPath(data, env) == DictDBPath(data, env) {
		t.Fatal("devdata 规格库不能落在读侧字典库的位置")
	}
}

// TestDevDataDirEmpty 空输入一律回空串（而不是拼出 "devdata/xxx" 这种半截路径）。
func TestDevDataDirEmpty(t *testing.T) {
	cases := []struct{ name, got string }{
		{"空 dataDir", DevDataDir("")},
		{"空 env", DevDataEnvDir(t.TempDir(), "")},
		{"空 prog", DevDataProgDir(t.TempDir(), "e", "")},
		{"空 env 的 prog", DevDataProgDir(t.TempDir(), "", "p")},
		{"空 dataDir 的库", DevDataDictPath("", "e")},
		{"空 cfg 的 ForEnv", DevDataEnvDirForEnv("", "")},
	}
	for _, c := range cases {
		if c.got != "" {
			t.Errorf("%s: 应为空串，得到 %q", c.name, c.got)
		}
	}
}

// TestDevDataRootDirForThreeLayers 三层解析：环境 devdataDir > 顶层 devdata.dir > 缺省。
func TestDevDataRootDirForThreeLayers(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, "config.json")
	env := "福联测试tst"

	// ① 都没有 → <data>/devdata
	if err := os.WriteFile(cfgPath, []byte(`{"hosts":{"sshs":[{"name":"福联测试tst"}]}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "devdata", env)
	if got := DevDataEnvDirForEnv(cfgPath, env); got != want {
		t.Fatalf("缺省 = %q, want %q", got, want)
	}

	// ② 顶层 devdata.dir
	topDir := filepath.Join(root, "top-devdata")
	body := fmt.Sprintf(`{"devdata":{"dir":%q},"hosts":{"sshs":[{"name":"福联测试tst"}]}}`, topDir)
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := DevDataEnvDirForEnv(cfgPath, env), filepath.Join(topDir, env); got != want {
		t.Fatalf("顶层 = %q, want %q", got, want)
	}
	if got, want := DevDataRootDirFor(cfgPath), topDir; got != want {
		t.Fatalf("根目录 = %q, want %q", got, want)
	}

	// ③ 环境级覆盖压过顶层
	envRoot := filepath.Join(root, "env-root")
	body = fmt.Sprintf(`{"devdata":{"dir":%q},"hosts":{"sshs":[{"name":"福联测试tst","devdataDir":%q}]}}`, topDir, envRoot)
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := DevDataEnvDirForEnv(cfgPath, env), filepath.Join(envRoot, env); got != want {
		t.Fatalf("环境覆盖 = %q, want %q", got, want)
	}
	// 别的环境不受这个覆盖影响
	other := "109tst"
	if got, want := DevDataEnvDirForEnv(cfgPath, other), filepath.Join(topDir, other); got != want {
		t.Fatalf("其它环境 = %q, want %q", got, want)
	}
	// prog/库 两个派生函数也跟着覆盖走
	if got, want := DevDataProgDirForEnv(cfgPath, env, "axmt500_wf"), filepath.Join(envRoot, env, "prog", "axmt500_wf"); got != want {
		t.Fatalf("程序目录 = %q, want %q", got, want)
	}
	if got, want := DevDataDictPathForEnv(cfgPath, env), filepath.Join(envRoot, env, "dict", env+".db"); got != want {
		t.Fatalf("规格库 = %q, want %q", got, want)
	}
}

// TestDevDataDirForEnvBrokenConfig 配置读不到/写坏时安静回退到缺省落点。
//
// 资料位置不该因为配置文件有问题就变成空串 —— 空串会让上层把"没配"与"配错了"
// 混成一种状态，最后表现为"材料落到了奇怪的地方"。
func TestDevDataDirForEnvBrokenConfig(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, "config.json")
	if err := os.WriteFile(cfgPath, []byte(`{not json`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := DevDataEnvDirForEnv(cfgPath, "e"), filepath.Join(root, "devdata", "e"); got != want {
		t.Fatalf("坏配置 = %q, want %q", got, want)
	}
	// 配置文件根本不存在，也一样
	missing := filepath.Join(root, "nope", "config.json")
	if got, want := DevDataEnvDirForEnv(missing, "e"), filepath.Join(root, "nope", "devdata", "e"); got != want {
		t.Fatalf("无配置 = %q, want %q", got, want)
	}
}

// TestDevDataEnvNameTrims：环境名首尾空白不该变成目录名的一部分。
func TestDevDataEnvNameTrims(t *testing.T) {
	if got := DevDataEnvName("  e  "); got != "e" {
		t.Fatalf("DevDataEnvName = %q, want e", got)
	}
	if got := DevDataEnvDir(t.TempDir(), " "); got != "" {
		t.Fatalf("空白环境名应回空串，得到 %q", got)
	}
}

// TestDevDataNotInCacheWhitelist ADR-02：devdata 不进缓存白名单。
func TestDevDataNotInCacheWhitelist(t *testing.T) {
	for _, name := range CacheSubdirs {
		if strings.EqualFold(name, DevDataDirName) {
			t.Fatalf("%q 不能进缓存白名单：tt cache clear 会整目录删掉 model 的原材料", name)
		}
	}
	// 反过来也要真的删不掉（白名单外返回空串，防手滑建目录）
	l := LocationsAt(filepath.Join(t.TempDir(), "config.json"))
	if got := l.CacheDir(DevDataDirName); got != "" {
		t.Fatalf("CacheDir(%q) = %q，应为空串", DevDataDirName, got)
	}
}
