package plugin

// 升级/回滚的失败路径（P6-04/P6-08）：任何一次失败都必须让"当前可用的那个版本"完好如初。
//
// 这一组用例故意都把失败放在不同的阶段 —— 索引读不到、资产下不来、摘要不符、zip 坏了、
// 注册表写不进去、服务起不来、技能刷不进平台目录 —— 因为"失败不破坏可用版本"是一条
// 按阶段成立的性质，不是一句笼统的保证。

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"tt/internal/connsvc"
	"tt/internal/hostrpc"
)

//---- 夹具 ----

// writePluginDirV 造一个指定版本的最小可用插件（cli + frontend + skills）。
func writePluginDirV(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ManifestFileName), fmt.Sprintf(`{
  "schemaVersion": 1,
  "id": "com.example.demo",
  "version": %q,
  "hostApi": ">=1.0.0 <2.0.0",
  "name": "演示插件",
  "contributions": {
    "cli": {"namespace": "demo", "entry": "bin/demo.bin", "summary": "演示"},
    "frontend": {"route": "demo", "title": "演示页面", "entry": "ui/index.html"},
    "skills": ["skills/demo-skill"]
  },
  "permissions": ["profiles.list", "db.connect"]
}`, version))
	writeFile(t, filepath.Join(dir, "bin", "demo.bin"), "#!/bin/sh\necho demo "+version+"\n")
	writeFile(t, filepath.Join(dir, "ui", "index.html"), "<!doctype html><title>demo "+version+"</title>")
	writeFile(t, filepath.Join(dir, "skills", "demo-skill", "SKILL.md"),
		"---\nname: demo-skill\ndescription: 演示技能\n---\n\n正文 "+version+"\n")
	return dir
}

// writeServicePluginDirV 造一个带常驻服务的插件包（升级时要被停/起的那个）。
func writeServicePluginDirV(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ManifestFileName), fmt.Sprintf(`{
  "schemaVersion": 1,
  "id": "com.example.demo",
  "version": %q,
  "hostApi": ">=1.0.0 <2.0.0",
  "contributions": {
    "cli": {"namespace": "demo", "entry": "bin/demo.bin"},
    "service": {"entry": "bin/demo.bin", "args": ["service"]}
  }
}`, version))
	writeFile(t, filepath.Join(dir, "bin", "demo.bin"), "#!/bin/sh\necho svc "+version+"\n")
	return dir
}

// zipDir 把目录打成 zip（相对路径入包，模拟发布产物）。
func zipDir(t *testing.T, src, dest string) {
	t.Helper()
	f, err := os.Create(dest)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(src, path)
		if rerr != nil {
			return rerr
		}
		w, werr := zw.Create(filepath.ToSlash(rel))
		if werr != nil {
			return werr
		}
		in, oerr := os.Open(path)
		if oerr != nil {
			return oerr
		}
		defer in.Close()
		_, cerr := io.Copy(w, in)
		return cerr
	})
	if err != nil {
		t.Fatalf("打包失败: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("关闭 zip 失败: %v", err)
	}
}

// fixtureAsset 索引里声明的一条资产（sha/size 可以与真实字节故意不一致）。
type fixtureAsset struct {
	path string
	sha  string
	size int64
	raw  []byte // 非 nil 时用它当响应体（模拟坏包/中断）
}

// upgradeFixture 升级用例的共用环境：装好 0.1.0 的一个插件 + 一台托管索引与资产的服务器。
type upgradeFixture struct {
	t   *testing.T
	m   *Manager
	s   *Store
	srv *httptest.Server
	// hostAPI 索引里声明的兼容范围（可以故意设成不兼容）。
	hostAPI string
	// assets 版本 → 资产
	assets map[string]*fixtureAsset
	// order 版本登记顺序（索引里按它输出）
	order []string
	// skills 技能同步的调用记录
	skills []string
	// runtime 升级时注入的运行态替身
	runtime *fakeUpdateRuntime
}

// fakeUpdateRuntime 记录"停/起/撤凭证"的顺序，并可按版本注入失败。
type fakeUpdateRuntime struct {
	calls []string
	// failStartFor 这个版本的 StartService 返回错误（空 = 都成功）。
	failStartFor string
}

func (r *fakeUpdateRuntime) StopService(id string) error {
	r.calls = append(r.calls, "stop "+id)
	return nil
}

func (r *fakeUpdateRuntime) StartService(rec *Record) error {
	r.calls = append(r.calls, "start "+rec.ActiveVersion)
	if r.failStartFor != "" && rec.ActiveVersion == r.failStartFor {
		return fmt.Errorf("模拟：服务在 %s 上起不来", rec.ActiveVersion)
	}
	return nil
}

func (r *fakeUpdateRuntime) InvalidateInstances(id string) {
	r.calls = append(r.calls, "invalidate "+id)
}

// newUpgradeFixture 造夹具。withService=true 时插件声明常驻服务。
func newUpgradeFixture(t *testing.T, withService bool) *upgradeFixture {
	t.Helper()
	m, s := newManager(t)
	f := &upgradeFixture{t: t, m: m, s: s, hostAPI: ">=1.0.0 <2.0.0", assets: map[string]*fixtureAsset{}}
	m.SkillsSync = func(reason string) error {
		f.skills = append(f.skills, reason)
		return nil
	}

	var dir string
	if withService {
		dir = writeServicePluginDirV(t, "0.1.0")
	} else {
		dir = writePluginDirV(t, "0.1.0")
	}
	if _, err := m.Install(InstallOptions{Source: dir}); err != nil {
		t.Fatalf("装初始版本失败: %v", err)
	}
	f.skills = nil // 安装那一次不算

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index.json" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(f.indexJSON(srv.URL))
			return
		}
		v := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), ".zip")
		a, ok := f.assets[v]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if a.raw != nil {
			_, _ = w.Write(a.raw)
			return
		}
		http.ServeFile(w, r, a.path)
	}))
	t.Cleanup(srv.Close)
	f.srv = srv
	return f
}

// indexJSON 按当前登记的资产造一份发布索引。
func (f *upgradeFixture) indexJSON(base string) []byte {
	idx := ReleaseIndex{SchemaVersion: IndexSchemaVersion}
	ip := IndexPlugin{ID: "com.example.demo"}
	for _, v := range f.order {
		a := f.assets[v]
		ip.Versions = append(ip.Versions, IndexVersion{
			Version: v, HostAPI: f.hostAPI,
			Assets: map[string]IndexAsset{
				AssetKey(): {URL: base + "/" + v + ".zip", SHA256: a.sha, Size: a.size},
			},
		})
	}
	idx.Plugins = append(idx.Plugins, ip)
	b, err := json.Marshal(idx)
	if err != nil {
		f.t.Fatal(err)
	}
	return b
}

// publish 把一个版本目录打成 zip 并正确地登记进索引。
func (f *upgradeFixture) publish(version string) *fixtureAsset {
	f.t.Helper()
	dir := writePluginDirV(f.t, version)
	return f.publishDir(version, dir)
}

// publishDir 把指定目录作为某版本的资产。
func (f *upgradeFixture) publishDir(version, dir string) *fixtureAsset {
	f.t.Helper()
	p := filepath.Join(f.t.TempDir(), version+".zip")
	zipDir(f.t, dir, p)
	sha, size := fileDigest(f.t, p)
	a := &fixtureAsset{path: p, sha: sha, size: size}
	f.assets[version] = a
	f.order = append(f.order, version)
	return a
}

// publishBytes 用自定义字节当某个版本的下载内容（sha/size 按真实字节算，用于坏包场景）。
func (f *upgradeFixture) publishBytes(version string, b []byte) *fixtureAsset {
	f.t.Helper()
	sum := sha256.Sum256(b)
	a := &fixtureAsset{raw: b, sha: hex.EncodeToString(sum[:]), size: int64(len(b))}
	f.assets[version] = a
	f.order = append(f.order, version)
	return a
}

// fileDigest 算文件的 sha256 与大小。
func fileDigest(t *testing.T, path string) (string, int64) {
	t.Helper()
	sha, err := fileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return sha, st.Size()
}

// update 跑一次升级（默认 hostApi 与平台都正常，索引用夹具那台服务器）。
//
// AllowInsecureHTTP 在这里必须打开：httptest 是普通 HTTP，这是**测试缝**不是生产策略
// （生产路径的 https-only 规则由 TestIndexRejectsBadShapes 单独钉住）。
func (f *upgradeFixture) update(opt UpdateOptions) []UpdateResult {
	f.t.Helper()
	opt.IndexURL = f.srv.URL + "/index.json"
	opt.AllowInsecureHTTP = true
	res, err := f.m.Update(context.Background(), opt)
	if err != nil {
		f.t.Fatalf("Update 整体失败: %v", err)
	}
	return res
}

// active 当前活跃版本。
func (f *upgradeFixture) active() *Record {
	f.t.Helper()
	reg, err := f.s.Load()
	if err != nil {
		f.t.Fatal(err)
	}
	rec := reg.Find("com.example.demo")
	if rec == nil {
		f.t.Fatal("注册表里没有插件")
	}
	return rec
}

// 保证夹具的 Store 字段名一致。
func (f *upgradeFixture) Store() *Store { return f.s }

// versionDirExists 某版本目录在不在。
func (f *upgradeFixture) versionDirExists(v string) bool {
	_, err := os.Stat(f.s.VersionDir("com.example.demo", v))
	return err == nil
}

//---- 成功路径 ----

// TestUpdateInstallsNewVersionAndKeepsPrevious 升级成功：切版本、留上一个、重算技能。
func TestUpdateInstallsNewVersionAndKeepsPrevious(t *testing.T) {
	f := newUpgradeFixture(t, false)
	f.publish("0.2.0")

	res := f.update(UpdateOptions{})
	if len(res) != 1 || !res[0].Updated {
		t.Fatalf("该升级成功，得 %+v", res)
	}
	rec := f.active()
	if rec.ActiveVersion != "0.2.0" || rec.PreviousVersion != "0.1.0" {
		t.Fatalf("版本记录不对: active=%s previous=%s", rec.ActiveVersion, rec.PreviousVersion)
	}
	if !f.versionDirExists("0.2.0") || !f.versionDirExists("0.1.0") {
		t.Error("当前与上一个版本的目录都该留着（回滚要用）")
	}
	if len(f.skills) == 0 || !strings.Contains(f.skills[len(f.skills)-1], "update") {
		t.Errorf("升级后该重算技能，得 %v", f.skills)
	}
	// 来源要跟着换成发布索引：下次升级才知道回哪个源问。
	if rec.Source.Kind != SourceURL || rec.Source.Ref != f.srv.URL+"/index.json" {
		t.Errorf("来源没跟着版本切换: %+v", rec.Source)
	}
}

// TestUpdateDryRunWritesNothing --dry-run 不下载、不切版本、不留垃圾。
func TestUpdateDryRunWritesNothing(t *testing.T) {
	f := newUpgradeFixture(t, false)
	f.publish("0.2.0")

	res := f.update(UpdateOptions{DryRun: true})
	if len(res) != 1 || res[0].Error != "" || res[0].Updated {
		t.Fatalf("dry-run 该只给计划，得 %+v", res)
	}
	if len(res[0].Notes) == 0 {
		t.Error("dry-run 该说明会做什么")
	}
	if f.active().ActiveVersion != "0.1.0" {
		t.Error("dry-run 不该改活跃版本")
	}
	if f.versionDirExists("0.2.0") {
		t.Error("dry-run 不该装任何版本目录")
	}
	entries, _ := os.ReadDir(f.s.CacheDir())
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "dl-") {
			t.Errorf("dry-run 不该写下载缓存: %s", e.Name())
		}
	}
}

// TestUpdatePrunesToCurrentAndPrevious 只保留当前与上一个程序版本（持久数据不动）。
func TestUpdatePrunesToCurrentAndPrevious(t *testing.T) {
	f := newUpgradeFixture(t, false)
	// 数据目录里有用户的东西，升级与清理都不许碰。
	dataFile := filepath.Join(f.s.DataDir("com.example.demo"), "state.json")
	writeFile(t, dataFile, `{"keep":true}`)

	f.publish("0.2.0")
	if res := f.update(UpdateOptions{}); !res[0].Updated {
		t.Fatalf("第一跳失败: %+v", res)
	}
	f.publish("0.3.0")
	if res := f.update(UpdateOptions{}); !res[0].Updated {
		t.Fatalf("第二跳失败: %+v", res)
	}
	rec := f.active()
	if rec.ActiveVersion != "0.3.0" || rec.PreviousVersion != "0.2.0" {
		t.Fatalf("版本记录不对: active=%s previous=%s", rec.ActiveVersion, rec.PreviousVersion)
	}
	if f.versionDirExists("0.1.0") {
		t.Error("0.1.0 既不是当前也不是上一个，该被清理")
	}
	if rec.HasVersion("0.1.0") {
		t.Error("注册表里也不该再记着 0.1.0")
	}
	if _, err := os.Stat(dataFile); err != nil {
		t.Errorf("持久数据被动了: %v", err)
	}
}

// TestUpdateUpToDate 没有更新版本不是错误。
func TestUpdateUpToDate(t *testing.T) {
	f := newUpgradeFixture(t, false)
	f.publish("0.1.0") // 与已装版本相同
	res := f.update(UpdateOptions{})
	if len(res) != 1 || !res[0].UpToDate || res[0].Error != "" {
		t.Fatalf("该报「已是最新」，得 %+v", res)
	}
}

//---- P6-08：失败不破坏可用版本 ----

// TestUpdateNoNetworkKeepsActiveVersion 索引都拉不到时什么都不变。
func TestUpdateNoNetworkKeepsActiveVersion(t *testing.T) {
	f := newUpgradeFixture(t, false)
	// 本机一个没人监听的端口：连接被拒（不碰真实网络）。
	_, err := f.m.Update(context.Background(), UpdateOptions{
		IndexURL: "http://127.0.0.1:1/index.json", AllowInsecureHTTP: true, Retries: -1,
	})
	if err == nil {
		t.Fatal("没有网络时该报错")
	}
	if f.active().ActiveVersion != "0.1.0" {
		t.Error("拉不到索引不该改活跃版本")
	}
}

// TestUpdateMissingAssetKeepsActiveVersion 索引在、资产 404：旧版本照跑。
func TestUpdateMissingAssetKeepsActiveVersion(t *testing.T) {
	f := newUpgradeFixture(t, false)
	// 索引里登记一个不存在的资产路径。
	f.assets["0.2.0"] = &fixtureAsset{sha: strings.Repeat("a", 64), size: 10}
	f.order = append(f.order, "0.2.0")

	res := f.update(UpdateOptions{})
	if len(res) != 1 || res[0].Error == "" {
		t.Fatalf("下不到资产该报错，得 %+v", res)
	}
	if f.active().ActiveVersion != "0.1.0" || f.versionDirExists("0.2.0") {
		t.Error("下载失败不该改变活跃版本或留下版本目录")
	}
}

// TestUpdateHashMismatchKeepsActiveVersion 摘要不符：拒绝并删掉半成品。
func TestUpdateHashMismatchKeepsActiveVersion(t *testing.T) {
	f := newUpgradeFixture(t, false)
	a := f.publish("0.2.0")
	a.sha = strings.Repeat("b", 64) // 索引声明的不是这份字节

	res := f.update(UpdateOptions{})
	if len(res) != 1 || !strings.Contains(res[0].Error, "校验不通过") {
		t.Fatalf("摘要不符该被拒，得 %+v", res)
	}
	if f.active().ActiveVersion != "0.1.0" || f.versionDirExists("0.2.0") {
		t.Error("摘要不符不该改变活跃版本或留下版本目录")
	}
	if _, err := os.Stat(filepath.Join(f.s.CacheDir(), "dl-com.example.demo-0.2.0.zip")); err == nil {
		t.Error("校验不过的下载物不该留在缓存里")
	}
}

// TestUpdateCorruptZipKeepsActiveVersion 字节对得上但 zip 坏了：同样不许动旧版本。
func TestUpdateCorruptZipKeepsActiveVersion(t *testing.T) {
	f := newUpgradeFixture(t, false)
	f.publishBytes("0.2.0", []byte("这不是一个 zip"))

	res := f.update(UpdateOptions{})
	if len(res) != 1 || res[0].Error == "" {
		t.Fatalf("坏 zip 该被拒，得 %+v", res)
	}
	if f.active().ActiveVersion != "0.1.0" || f.versionDirExists("0.2.0") {
		t.Error("坏包不该改变活跃版本或留下版本目录")
	}
}

// TestUpdateRegistryWriteFailureKeepsActiveVersion 注册表写不进去（锁被占）：载荷撤掉、版本不变。
func TestUpdateRegistryWriteFailureKeepsActiveVersion(t *testing.T) {
	f := newUpgradeFixture(t, false)
	f.publish("0.2.0")

	// 一把"新鲜"的锁：另一个 tt 进程正拿着注册表。它会让每一次 WithLock 都失败。
	lock := f.s.RegistryPath() + ".lock"
	writeFile(t, lock, "pid=999999\n")
	t.Cleanup(func() { _ = os.Remove(lock) })
	f.s.lockWait = 30 * time.Millisecond // 30ms，别让用例白等 2 秒

	res := f.update(UpdateOptions{})
	if len(res) != 1 || res[0].Error == "" {
		t.Fatalf("注册表写不进去该报错，得 %+v", res)
	}
	if f.active().ActiveVersion != "0.1.0" {
		t.Error("注册表写失败不该改变活跃版本")
	}
	if f.versionDirExists("0.2.0") {
		t.Error("登记失败的载荷该被撤掉，不留孤儿目录")
	}
}

// TestUpdateRollsBackWhenServiceFails 切换后服务起不来：切回旧版本并重新起旧服务。
func TestUpdateRollsBackWhenServiceFails(t *testing.T) {
	f := newUpgradeFixture(t, true)
	rt := &fakeUpdateRuntime{failStartFor: "0.2.0"}
	f.m.Runtime = rt
	f.publishDir("0.2.0", writeServicePluginDirV(t, "0.2.0"))

	res := f.update(UpdateOptions{})
	if len(res) != 1 || res[0].Error == "" {
		t.Fatalf("服务起不来该报升级失败，得 %+v", res)
	}
	if !res[0].RolledBack {
		t.Errorf("该明确报告已回滚，得 %+v", res[0])
	}
	rec := f.active()
	if rec.ActiveVersion != "0.1.0" {
		t.Fatalf("该退回 0.1.0，得 %s", rec.ActiveVersion)
	}
	if rec.PreviousVersion != "" {
		t.Errorf("回滚后不该再指着一个刚被证明不可用的版本，得 %q", rec.PreviousVersion)
	}
	if !strings.Contains(rec.LastError, "回滚") {
		t.Errorf("LastError 该说明为什么退回来，得 %q", rec.LastError)
	}
	// 顺序：停旧服务 → 撤凭证 → 切新版 → 起新版失败 → 停 → 撤 → 切回 → 起旧版
	joined := strings.Join(rt.calls, " | ")
	for _, want := range []string{"stop com.example.demo", "invalidate com.example.demo", "start 0.2.0", "start 0.1.0"} {
		if !strings.Contains(joined, want) {
			t.Errorf("运行态操作里缺少 %q（实际: %s）", want, joined)
		}
	}
	if !f.versionDirExists("0.2.0") {
		t.Error("失败的新版本目录留着（供 --to 重试/人工检查），不该被删")
	}
}

// TestUpdateRollsBackWhenSkillsSyncFails 技能刷不进平台目录：同样回滚（P6-04）。
func TestUpdateRollsBackWhenSkillsSyncFails(t *testing.T) {
	f := newUpgradeFixture(t, false)
	f.m.SkillsSync = func(reason string) error {
		if strings.Contains(reason, "update") {
			return fmt.Errorf("模拟：平台目录写不进去")
		}
		f.skills = append(f.skills, reason)
		return nil
	}
	f.m.Runtime = &fakeUpdateRuntime{}
	f.publish("0.2.0")

	res := f.update(UpdateOptions{})
	if len(res) != 1 || res[0].Error == "" || !res[0].RolledBack {
		t.Fatalf("技能刷新失败该回滚，得 %+v", res)
	}
	if f.active().ActiveVersion != "0.1.0" {
		t.Error("技能失败后该退回 0.1.0")
	}
}

// TestUpdateRejectsPackageWhoseUIEntryIsMissing 包里的页面入口不在：切换前就拒，不留半成品。
func TestUpdateRejectsPackageWhoseUIEntryIsMissing(t *testing.T) {
	f := newUpgradeFixture(t, false)
	dir := writePluginDirV(t, "0.2.0")
	if err := os.Remove(filepath.Join(dir, "ui", "index.html")); err != nil {
		t.Fatal(err)
	}
	f.publishDir("0.2.0", dir)

	res := f.update(UpdateOptions{})
	if len(res) != 1 || !strings.Contains(res[0].Error, "ui/index.html") {
		t.Fatalf("缺 UI 入口该被拒，得 %+v", res)
	}
	if f.active().ActiveVersion != "0.1.0" || f.versionDirExists("0.2.0") {
		t.Error("校验不过的包不该留下版本目录")
	}
}

// TestUpdateRejectsIncompatibleHostAPI 索引里的新版与当前 Host API 不兼容：跳过而不是硬装。
func TestUpdateRejectsIncompatibleHostAPI(t *testing.T) {
	f := newUpgradeFixture(t, false)
	f.hostAPI = ">=2.0.0 <3.0.0"
	f.publish("0.2.0")

	res := f.update(UpdateOptions{})
	if len(res) != 1 || !res[0].UpToDate {
		t.Fatalf("不兼容的版本该被跳过（已是最新），得 %+v", res)
	}
	if f.active().ActiveVersion != "0.1.0" {
		t.Error("不该装不兼容的版本")
	}
}

// TestUpdateRefusesWhenServeOwnsTheService 别的 tt 进程正跑着服务时命令行拒绝升级。
func TestUpdateRefusesWhenServeOwnsTheService(t *testing.T) {
	f := newUpgradeFixture(t, true)
	f.publishDir("0.2.0", writeServicePluginDirV(t, "0.2.0"))

	// 造一份"正在跑"的实例状态：PID 用本进程（一定活着），版本是旧版本。
	st := fmt.Sprintf(`{"pluginId":"com.example.demo","version":"0.1.0","instanceId":"svc_x","pid":%d,"state":"running"}`, os.Getpid())
	writeFile(t, filepath.Join(f.s.RunDir("com.example.demo"), "instance.json"), st)

	res := f.update(UpdateOptions{})
	if len(res) != 1 || !strings.Contains(res[0].Error, "tt serve") {
		t.Fatalf("该拒绝并提示先停 serve，得 %+v", res)
	}
	if f.active().ActiveVersion != "0.1.0" || f.versionDirExists("0.2.0") {
		t.Error("被拒时不该装新版本")
	}
}

//---- 回滚 ----

// TestRollbackSwitchesBackAndRestarts 回滚：切回上一个版本、重新起服务、重算技能。
func TestRollbackSwitchesBackAndRestarts(t *testing.T) {
	f := newUpgradeFixture(t, true)
	rt := &fakeUpdateRuntime{}
	f.m.Runtime = rt
	f.publishDir("0.2.0", writeServicePluginDirV(t, "0.2.0"))
	if res := f.update(UpdateOptions{}); !res[0].Updated {
		t.Fatalf("升级该成功: %+v", res)
	}
	rt.calls = nil

	res, err := f.m.Rollback("com.example.demo", RollbackOptions{})
	if err != nil {
		t.Fatalf("回滚失败: %v", err)
	}
	if !res.Updated {
		t.Fatalf("回滚该成功，得 %+v", res)
	}
	rec := f.active()
	if rec.ActiveVersion != "0.1.0" || rec.PreviousVersion != "0.2.0" {
		t.Fatalf("回滚后版本记录不对: active=%s previous=%s", rec.ActiveVersion, rec.PreviousVersion)
	}
	joined := strings.Join(rt.calls, " | ")
	if !strings.Contains(joined, "stop com.example.demo") || !strings.Contains(joined, "start 0.1.0") {
		t.Errorf("回滚要真停真起，实际: %s", joined)
	}
	if len(f.skills) == 0 {
		t.Error("回滚后该重算技能")
	}
}

// TestRollbackWithoutPreviousVersion 没切过版本就没有可回滚的版本。
func TestRollbackWithoutPreviousVersion(t *testing.T) {
	f := newUpgradeFixture(t, false)
	if _, err := f.m.Rollback("com.example.demo", RollbackOptions{}); err == nil {
		t.Fatal("没有上一个版本时该报错")
	} else if !strings.Contains(err.Error(), "没有可回滚的版本") {
		t.Fatalf("错误该说清原因，得 %v", err)
	}
}

// TestRollbackFailureReturnsToCurrent 回滚本身失败时把活跃版本切回原处。
func TestRollbackFailureReturnsToCurrent(t *testing.T) {
	f := newUpgradeFixture(t, true)
	rt := &fakeUpdateRuntime{}
	f.m.Runtime = rt
	f.publishDir("0.2.0", writeServicePluginDirV(t, "0.2.0"))
	if res := f.update(UpdateOptions{}); !res[0].Updated {
		t.Fatalf("升级该成功: %+v", res)
	}
	// 现在让"旧版本"起不来：回滚的切换后验收失败 → 应该退回 0.2.0。
	rt.failStartFor = "0.1.0"

	if _, err := f.m.Rollback("com.example.demo", RollbackOptions{}); err == nil {
		t.Fatal("回滚失败该报错")
	}
	if got := f.active().ActiveVersion; got != "0.2.0" {
		t.Errorf("回滚失败后该退回 0.2.0，得 %s", got)
	}
}

//---- 真进程：切版本、换实例、撤旧令牌 ----

// supRuntime 用真 supervisor 当 UpdateRuntime —— 等价于 tt serve 进程里那一份。
type supRuntime struct {
	sup    *Supervisor
	tokens *hostrpc.TokenStore
	conn   *connsvc.Service
}

func (r supRuntime) StopService(id string) error { return r.sup.Stop(id) }

func (r supRuntime) StartService(rec *Record) error {
	if rec == nil || rec.Contributions.Service == nil {
		return nil
	}
	if _, err := r.sup.Start(context.Background(), rec); err != nil {
		return err
	}
	return nil
}

func (r supRuntime) InvalidateInstances(id string) {
	r.tokens.RevokePlugin(id)
	r.conn.CloseForPlugin(id)
}

// mustInstanceState 读 run/<id>/instance.json（没有就失败）。
func mustInstanceState(t *testing.T, s *Store, id string) *InstanceState {
	t.Helper()
	st, err := s.ReadInstanceState(id)
	if err != nil {
		t.Fatal(err)
	}
	if st == nil {
		t.Fatal("没有运行状态文件")
	}
	return st
}

// copyPackageAsVersion 拷一份包并把清单版本改掉（造"新版本"）。
func copyPackageAsVersion(t *testing.T, src, version string) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "pkg-"+version)
	if err := copyTree(src, dst, DefaultZipLimits); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dst, ManifestFileName)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var man Manifest
	if err := json.Unmarshal(b, &man); err != nil {
		t.Fatal(err)
	}
	man.Version = version
	nb, err := json.MarshalIndent(man, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(nb, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return dst
}

// TestUpdateWithRealServiceSwitchesVersionAndInstance 真服务升级：
// 停旧服务 → 切版本 → 起新服务 → run/<id>/instance.json 指向新实例 → 旧令牌撤销（P6-03/P6-05）。
func TestUpdateWithRealServiceSwitchesVersionAndInstance(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}

	cfgPath := s.Locations().ConfigPath
	tokens := hostrpc.NewTokenStore()
	conn := connsvc.New(connsvc.Options{ConfigPath: cfgPath, Authorizer: allowAll{}})
	sup := NewSupervisor(s, tokens, conn, "", "test")
	sup.Logf = t.Logf
	sup.StartTimeout = 10 * time.Second
	m.Runtime = supRuntime{sup: sup, tokens: tokens, conn: conn}
	t.Cleanup(sup.StopAll)

	// tt serve 启动时的那一步。
	if _, err := sup.Start(context.Background(), rec); err != nil {
		t.Fatalf("起旧版本服务失败: %v", err)
	}
	first := mustInstanceState(t, s, rec.ID)
	if first.Version != "0.1.0" || !first.Alive() {
		t.Fatalf("旧实例状态不对: %+v", first)
	}
	if tokens.Count() != 1 {
		t.Fatalf("旧实例该有一个令牌，得 %d", tokens.Count())
	}

	// 造 0.2.0 的包并起一台只托管它的索引。
	pkgV2 := copyPackageAsVersion(t, pkg, "0.2.0")
	zipPath := filepath.Join(t.TempDir(), "0.2.0.zip")
	zipDir(t, pkgV2, zipPath)
	sha, size := fileDigest(t, zipPath)
	idxBytes, err := json.Marshal(ReleaseIndex{
		SchemaVersion: IndexSchemaVersion,
		Plugins: []IndexPlugin{{ID: rec.ID, Versions: []IndexVersion{{
			Version: "0.2.0", HostAPI: ">=1.0.0 <2.0.0",
			Assets: map[string]IndexAsset{AssetKey(): {URL: "PLACEHOLDER", SHA256: sha, Size: size}},
		}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/0.2.0.zip" {
			http.ServeFile(w, r, zipPath)
			return
		}
		body := strings.Replace(string(idxBytes), "PLACEHOLDER", srv.URL+"/0.2.0.zip", 1)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	res, uerr := m.Update(context.Background(), UpdateOptions{
		IndexURL: srv.URL + "/index.json", AllowInsecureHTTP: true, Retries: -1,
	})
	if uerr != nil {
		t.Fatalf("升级失败: %v", uerr)
	}
	if len(res) != 1 || !res[0].Updated {
		t.Fatalf("该升级成功，得 %+v", res)
	}

	second := mustInstanceState(t, s, rec.ID)
	if second.Version != "0.2.0" {
		t.Errorf("instance.json 该指向新版本，得 %+v", second)
	}
	if second.InstanceID == "" || second.InstanceID == first.InstanceID {
		t.Error("该是一个新实例（旧实例必须作废）")
	}
	if !second.Alive() {
		t.Error("新实例该在跑")
	}
	// 旧令牌撤销了，只剩新实例那一个。
	if tokens.Count() != 1 {
		t.Errorf("升级后该只剩新实例的令牌，得 %d", tokens.Count())
	}
	if rec := fRecord(t, s, rec.ID); rec.ActiveVersion != "0.2.0" || rec.PreviousVersion != "0.1.0" {
		t.Errorf("注册表版本记录不对: active=%s previous=%s", rec.ActiveVersion, rec.PreviousVersion)
	}
	// 上一个版本的程序仍在（回滚要用）。
	if _, err := os.Stat(s.VersionDir(rec.ID, "0.1.0")); err != nil {
		t.Errorf("上一个版本的目录该留着: %v", err)
	}
}

// TestUpdateHashMismatchKeepsRealServiceRunning P6-08 的正面证据：真服务在跑的时候，
// 一次摘要不符的升级不能碰它一个字节 —— 活跃版本、实例、进程都不变。
func TestUpdateHashMismatchKeepsRealServiceRunning(t *testing.T) {
	m, s := newManager(t)
	pkg := buildReferencePlugin(t)
	rec := installRef(t, m, pkg)
	if _, err := m.Enable(rec.ID, Grants{}); err != nil {
		t.Fatal(err)
	}
	cfgPath := s.Locations().ConfigPath
	tokens := hostrpc.NewTokenStore()
	conn := connsvc.New(connsvc.Options{ConfigPath: cfgPath, Authorizer: allowAll{}})
	sup := NewSupervisor(s, tokens, conn, "", "test")
	sup.Logf = t.Logf
	m.Runtime = supRuntime{sup: sup, tokens: tokens, conn: conn}
	t.Cleanup(sup.StopAll)
	if _, err := sup.Start(context.Background(), rec); err != nil {
		t.Fatalf("起服务失败: %v", err)
	}
	before := mustInstanceState(t, s, rec.ID)

	// 索引声明的摘要与实际字节不符。
	pkgV2 := copyPackageAsVersion(t, pkg, "0.2.0")
	zipPath := filepath.Join(t.TempDir(), "0.2.0.zip")
	zipDir(t, pkgV2, zipPath)
	_, size := fileDigest(t, zipPath)
	badIdx, err := json.Marshal(ReleaseIndex{
		SchemaVersion: IndexSchemaVersion,
		Plugins: []IndexPlugin{{ID: rec.ID, Versions: []IndexVersion{{
			Version: "0.2.0", HostAPI: ">=1.0.0 <2.0.0",
			Assets: map[string]IndexAsset{AssetKey(): {
				URL: "PLACEHOLDER", SHA256: strings.Repeat("c", 64), Size: size,
			}},
		}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/0.2.0.zip" {
			http.ServeFile(w, r, zipPath)
			return
		}
		_, _ = w.Write([]byte(strings.Replace(string(badIdx), "PLACEHOLDER", srv.URL+"/0.2.0.zip", 1)))
	}))
	t.Cleanup(srv.Close)

	res, uerr := m.Update(context.Background(), UpdateOptions{
		IndexURL: srv.URL + "/index.json", AllowInsecureHTTP: true, Retries: -1,
	})
	if uerr != nil {
		t.Fatalf("整批失败（这里期望「单个插件失败」）: %v", uerr)
	}
	if len(res) != 1 || !strings.Contains(res[0].Error, "校验不通过") {
		t.Fatalf("摘要不符该被拒，得 %+v", res)
	}

	after := mustInstanceState(t, s, rec.ID)
	if after.InstanceID != before.InstanceID || after.PID != before.PID || !after.Alive() {
		t.Errorf("旧服务被打断了: before=%+v after=%+v", before, after)
	}
	if after.Version != "0.1.0" {
		t.Errorf("实例版本该还是 0.1.0，得 %s", after.Version)
	}
	if got := fRecord(t, s, rec.ID).ActiveVersion; got != "0.1.0" {
		t.Errorf("活跃版本该还是 0.1.0，得 %s", got)
	}
}

// fRecord 读一条最新记录。
func fRecord(t *testing.T, s *Store, id string) *Record {
	t.Helper()
	reg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	rec := reg.Find(id)
	if rec == nil {
		t.Fatal("注册表里没有插件")
	}
	return rec
}

//---- 索引 ----

// TestIndexRejectsBadShapes 索引的严格性：版本、未知字段、http 地址都要拒。
func TestIndexRejectsBadShapes(t *testing.T) {
	good := `{"schemaVersion":1,"plugins":[{"id":"com.example.a","versions":[{"version":"1.0.0",
	  "hostApi":">=1.0.0 <2.0.0","assets":{"%s":{"url":"https://example.com/a.zip",
	  "sha256":"%s","size":10}}}]}]}`
	asset := fmt.Sprintf(good, AssetKey(), strings.Repeat("a", 64))
	if _, err := ParseReleaseIndex([]byte(asset), false); err != nil {
		t.Fatalf("合法索引该通过: %v", err)
	}

	cases := map[string]string{
		"schemaVersion 更高": strings.Replace(asset, `"schemaVersion":1`, `"schemaVersion":2`, 1),
		"schemaVersion 缺失": strings.Replace(asset, `"schemaVersion":1`, `"schemaVersion":0`, 1),
		"未知字段":             strings.Replace(asset, `"assets"`, `"assetz"`, 1),
		"http 地址":          strings.Replace(asset, "https://example.com", "http://example.com", 1),
		"摘要不是 64 位":        strings.Replace(asset, strings.Repeat("a", 64), "abc", 1),
		"大小缺失":             strings.Replace(asset, `"size":10`, `"size":0`, 1),
		"版本不是 SemVer":      strings.Replace(asset, `"version":"1.0.0"`, `"version":"1.0"`, 1),
		"资产键不是 os-arch":    strings.Replace(asset, AssetKey(), AssetKey()+"-x", 1),
	}
	for name, body := range cases {
		if _, err := ParseReleaseIndex([]byte(body), false); err == nil {
			t.Errorf("%s: 该被拒绝", name)
		}
	}
	// 测试缝：允许 http 只在显式放开时生效。
	httpBody := strings.Replace(asset, "https://example.com", "http://example.com", 1)
	if _, err := ParseReleaseIndex([]byte(httpBody), true); err != nil {
		t.Errorf("显式放开 http 时该通过: %v", err)
	}
}

// TestIndexPickRules 选版本：最新兼容、跳过不兼容、--to 必须点名存在。
func TestIndexPickRules(t *testing.T) {
	asset := func() map[string]IndexAsset {
		return map[string]IndexAsset{AssetKey(): {URL: "https://e/a.zip", SHA256: strings.Repeat("a", 64), Size: 1}}
	}
	ip := &IndexPlugin{ID: "com.example.a", Versions: []IndexVersion{
		{Version: "0.1.0", HostAPI: ">=1.0.0 <2.0.0", Assets: asset()},
		{Version: "0.2.0", HostAPI: ">=1.0.0 <2.0.0", Assets: asset()},
		{Version: "0.3.0", HostAPI: ">=2.0.0 <3.0.0", Assets: asset()},
	}}
	v, err := ip.Pick("0.1.0", "1.0.0", "")
	if err != nil || v == nil || v.Version != "0.2.0" {
		t.Fatalf("该选 0.2.0（0.3.0 与 Host API 不兼容），得 %v / %v", v, err)
	}
	if v, err := ip.Pick("0.2.0", "1.0.0", ""); err != nil || v != nil {
		t.Fatalf("已是最新时该返回空，得 %v / %v", v, err)
	}
	if _, err := ip.Pick("0.1.0", "1.0.0", "9.9.9"); err == nil {
		t.Error("--to 指一个不存在的版本该报错")
	}
	if _, err := ip.Pick("0.1.0", "1.0.0", "0.3.0"); err == nil {
		t.Error("--to 指一个与 Host API 不兼容的版本该报错")
	}
	if v, err := ip.Pick("0.1.0", "1.0.0", "0.2.0"); err != nil || v == nil || v.Version != "0.2.0" {
		t.Errorf("--to 点名的版本该被选中，得 %v / %v", v, err)
	}
}

// TestIndexFromLocalFile 索引与资产都可以是本机文件（离线/共享盘）。
func TestIndexFromLocalFile(t *testing.T) {
	f := newUpgradeFixture(t, false)
	dir := writePluginDirV(t, "0.2.0")
	zipPath := filepath.Join(t.TempDir(), "0.2.0.zip")
	zipDir(t, dir, zipPath)
	sha, size := fileDigest(t, zipPath)

	idx := ReleaseIndex{SchemaVersion: IndexSchemaVersion, Plugins: []IndexPlugin{{
		ID: "com.example.demo",
		Versions: []IndexVersion{{
			Version: "0.2.0", HostAPI: ">=1.0.0 <2.0.0",
			Assets: map[string]IndexAsset{AssetKey(): {URL: zipPath, SHA256: sha, Size: size}},
		}},
	}}}
	b, err := json.Marshal(idx)
	if err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(t.TempDir(), "index.json")
	if err := os.WriteFile(indexPath, b, 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := f.m.Update(context.Background(), UpdateOptions{IndexURL: indexPath, Retries: -1})
	if err != nil {
		t.Fatalf("本机索引该能用: %v", err)
	}
	if len(res) != 1 || !res[0].Updated {
		t.Fatalf("该升级成功，得 %+v", res)
	}
	if f.active().ActiveVersion != "0.2.0" {
		t.Error("活跃版本该是 0.2.0")
	}
}
