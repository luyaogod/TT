package plugin

// P6-16：参考插件的**一次跑完**的全生命周期。
//
// 安装 → 启用（含授权）→ CLI 转发（取退出码）→ 用本机 file:// 发布索引升到第二版
// → 回滚 → 停用 → 卸载，每一步都断言注册表/磁盘状态，最后断言持久数据没有被动过。
//
// 为什么只用一条用例串完：分段用例各自都绿，也可能拼出一条走不通的路（例如升级之后
// 注册表里还留着旧来源、回滚把数据目录一起删了）。这条用例就是验收清单里那行
// "参考插件全生命周期"的可执行版本。
//
// 全程离线：索引与资产都是本机 file:// 文件，不碰网络。

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/luyaogod/TT/internal/hostrpc"
)

// localFileURL 把一个本机路径变成 fetch 认识的 file:// 地址（空格等会被转义）。
//
// Windows 的盘符路径（C:\x）要补一个前导斜杠，url.URL.String() 才会给出 file:///C:/x；
// 补错了 fetch.resolveSource 会判成"没有路径"。
func localFileURL(t *testing.T, path string) string {
	t.Helper()
	p := filepath.ToSlash(path)
	u := &url.URL{Scheme: "file", Path: p}
	if !strings.HasPrefix(p, "/") {
		u.Path = "/" + p
	}
	if u.String() == "" {
		t.Fatalf("造不出 file:// 地址: %q", path)
	}
	return u.String()
}

// versionDirExists 某插件某版本的程序目录在不在。
func versionDirExists(s *Store, id, version string) bool {
	dir := s.VersionDir(id, version)
	if dir == "" {
		return false
	}
	_, err := os.Stat(dir)
	return err == nil
}

// mustRecord 读注册表里的某条记录（不存在就失败）。
func mustRecord(t *testing.T, s *Store, id string) *Record {
	t.Helper()
	reg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	rec := reg.Find(id)
	if rec == nil {
		t.Fatalf("注册表里没有插件 %s", id)
	}
	return rec
}

// TestReferencePluginFullLifecycle 一次跑完参考插件的全生命周期（P6-16）。
func TestReferencePluginFullLifecycle(t *testing.T) {
	started := time.Now()
	m, s := newManager(t)
	var skillsSync []string
	m.SkillsSync = func(reason string) error {
		skillsSync = append(skillsSync, reason)
		return nil
	}

	// ---- 1) 安装 ----
	pkg := buildReferencePlugin(t)
	rec, err := m.Install(InstallOptions{Source: pkg})
	if err != nil {
		t.Fatalf("安装失败: %v", err)
	}
	if rec.ID != "com.example.refplugin" || rec.Enabled || rec.ActiveVersion != "0.1.0" {
		t.Fatalf("安装后的记录不对: %+v", rec)
	}
	if !versionDirExists(s, rec.ID, "0.1.0") {
		t.Fatal("版本目录没落盘")
	}
	// 用户数据：写一份，最后要证明它活过了升级/回滚/卸载。
	dataFile := filepath.Join(s.DataDir(rec.ID), "state.json")
	writeFile(t, dataFile, `{"keep":true,"by":"lifecycle-test"}`)

	// ---- 2) 启用（含授权）----
	enabled, err := m.Enable(rec.ID, Grants{Capabilities: []string{"profiles.list"}, Envs: []string{"prod"}})
	if err != nil {
		t.Fatalf("启用失败: %v", err)
	}
	if !enabled.Enabled {
		t.Fatal("启用后 enabled 该是 true")
	}
	if !enabled.Grants.Has("profiles.list") || !enabled.Grants.AllowsEnv("prod") {
		t.Fatalf("授权没落盘: %+v", enabled.Grants)
	}
	if rec := mustRecord(t, s, rec.ID); !rec.Enabled {
		t.Fatal("注册表里的启用状态没写下去")
	}

	// ---- 3) CLI 转发：参数与退出码原样 ----
	tokens := hostrpc.NewTokenStore()
	d := refDispatcher(t, s, tokens)
	var out strings.Builder
	d.Stdout = &out
	res, err := d.Run(context.Background(), rec, "prod", "", []string{"--exit", "7", "全生命周期"})
	if err != nil {
		t.Fatalf("CLI 转发失败: %v", err)
	}
	if res.ExitCode != 7 {
		t.Fatalf("插件退出码该原样是 7，得 %d", res.ExitCode)
	}
	if !strings.Contains(out.String(), "全生命周期") {
		t.Fatalf("参数没原样转给插件: %s", out.String())
	}
	if tokens.Count() != 0 {
		t.Errorf("一次性调用结束后令牌该撤销，剩 %d", tokens.Count())
	}

	// ---- 4) 用本机 file:// 索引升级到 0.2.0 ----
	pkgV2 := copyPackageAsVersion(t, pkg, "0.2.0")
	zipPath := filepath.Join(t.TempDir(), "refplugin-0.2.0.zip")
	zipDir(t, pkgV2, zipPath)
	sha, size := fileDigest(t, zipPath)
	indexPath := filepath.Join(t.TempDir(), "index.json")
	idx := ReleaseIndex{SchemaVersion: IndexSchemaVersion, Plugins: []IndexPlugin{{
		ID: rec.ID,
		Versions: []IndexVersion{{
			Version: "0.2.0", HostAPI: ">=1.0.0 <2.0.0",
			Assets: map[string]IndexAsset{AssetKey(): {
				URL: localFileURL(t, zipPath), SHA256: sha, Size: size,
			}},
		}},
	}}}
	b, err := json.Marshal(idx)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(indexPath, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	ups, err := m.Update(context.Background(), UpdateOptions{
		ID: rec.ID, IndexURL: localFileURL(t, indexPath), Retries: -1,
	})
	if err != nil {
		t.Fatalf("升级整体失败: %v", err)
	}
	if len(ups) != 1 || !ups[0].Updated || ups[0].Error != "" {
		t.Fatalf("升级没成功: %+v", ups)
	}
	afterUpdate := mustRecord(t, s, rec.ID)
	if afterUpdate.ActiveVersion != "0.2.0" || afterUpdate.PreviousVersion != "0.1.0" {
		t.Fatalf("升级后的版本记录不对: active=%s previous=%s",
			afterUpdate.ActiveVersion, afterUpdate.PreviousVersion)
	}
	if !versionDirExists(s, rec.ID, "0.1.0") || !versionDirExists(s, rec.ID, "0.2.0") {
		t.Error("升级后当前与上一个版本的程序目录都该留着（回滚要用）")
	}
	if afterUpdate.Source.Kind != SourceURL {
		t.Errorf("来源该跟着换成发布索引，得 %+v", afterUpdate.Source)
	}
	if !strings.Contains(strings.Join(skillsSync, " | "), "update") {
		t.Errorf("升级后该重算技能，记录: %v", skillsSync)
	}

	// ---- 5) 回滚 ----
	rb, err := m.Rollback(rec.ID, RollbackOptions{})
	if err != nil {
		t.Fatalf("回滚失败: %v", err)
	}
	if !rb.Updated {
		t.Fatalf("回滚没成功: %+v", rb)
	}
	afterRollback := mustRecord(t, s, rec.ID)
	if afterRollback.ActiveVersion != "0.1.0" || afterRollback.PreviousVersion != "0.2.0" {
		t.Fatalf("回滚后的版本记录不对: active=%s previous=%s",
			afterRollback.ActiveVersion, afterRollback.PreviousVersion)
	}
	if !versionDirExists(s, rec.ID, "0.2.0") {
		t.Error("回滚后新版本的程序该留着（还能再切回去）")
	}

	// ---- 6) 停用 ----
	disabled, err := m.Disable(rec.ID)
	if err != nil {
		t.Fatalf("停用失败: %v", err)
	}
	if disabled.Enabled {
		t.Fatal("停用后 enabled 该是 false")
	}
	if rec := mustRecord(t, s, rec.ID); rec.Enabled {
		t.Fatal("注册表里的停用状态没写下去")
	}

	// ---- 7) 卸载（默认保留数据）----
	if err := m.Uninstall(rec.ID, UninstallOptions{}); err != nil {
		t.Fatalf("卸载失败: %v", err)
	}
	reg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if reg.Find(rec.ID) != nil {
		t.Fatal("卸载后注册表里不该还有这条记录")
	}
	for _, v := range []string{"0.1.0", "0.2.0"} {
		if versionDirExists(s, rec.ID, v) {
			t.Errorf("卸载后版本 %s 的程序目录该被删掉", v)
		}
	}
	// 持久数据是用户的，不是程序的一部分：升级/回滚/卸载都不许动。
	got, err := os.ReadFile(dataFile)
	if err != nil {
		t.Fatalf("持久数据被删了: %v", err)
	}
	if !strings.Contains(string(got), `"keep":true`) {
		t.Fatalf("持久数据被改了: %s", got)
	}

	elapsed := time.Since(started)
	t.Logf("全生命周期耗时 %s（技能重算 %d 次）", elapsed.Round(time.Millisecond), len(skillsSync))
	if elapsed > 15*time.Second {
		t.Errorf("全生命周期耗时 %s，超过约定的 ~15s（变慢了：先看是不是有真实等待/重试）", elapsed)
	}
}
