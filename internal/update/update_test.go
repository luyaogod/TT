package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"tt/internal/config"
)

// ---------- 版本模型 ----------

func TestParseVersionAndCompare(t *testing.T) {
	ok := []struct {
		in   string
		want Version
	}{
		{"0.2.1", Version{0, 2, 1, ""}},
		{"v0.2.1", Version{0, 2, 1, ""}}, // tag 带 v，VERSION 文件不带
		{"1.0.0", Version{1, 0, 0, ""}},
		{"10.20.30", Version{10, 20, 30, ""}},
		{"v0.3.0-rc1", Version{0, 3, 0, "rc1"}}, // 预发布 tag：注入二进制的版本永远不会长这样
		{"0.3.0+build7", Version{0, 3, 0, "build7"}},
	}
	for _, c := range ok {
		got, err := ParseVersion(c.in)
		if err != nil {
			t.Errorf("ParseVersion(%q) 报错: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseVersion(%q) = %+v，想要 %+v", c.in, got, c.want)
		}
	}
	// 空串必须报错：本地 go build 的 Version 就是空的，悄悄参与比较会让"是否需要更新"
	// 这件事永久为假。
	for _, bad := range []string{"", "0.2", "0.2.1.0", "1.2.x", "v", "1.2.3 ", "release-1"} {
		if _, err := ParseVersion(bad); err == nil {
			t.Errorf("ParseVersion(%q) 该报错", bad)
		}
	}

	cmp := []struct {
		a, b string
		want int
	}{
		{"0.2.1", "0.2.1", 0},
		{"0.2.0", "0.2.1", -1},
		{"0.2.1", "0.2.0", 1},
		{"0.9.9", "0.10.0", -1}, // 数字比较，不是字符串比较
		{"1.0.0", "0.99.99", 1},
		// 同一个数字段上，带后缀的比正式的旧：已装正式版的人不该被劝去"升级"到更早的候选版。
		{"0.3.0-rc1", "0.3.0", -1},
		{"0.3.0", "0.3.0-rc1", 1},
		{"0.2.1", "0.3.0-rc1", -1},
		{"0.3.0-rc1", "0.3.0-rc2", -1},
	}
	for _, c := range cmp {
		va, _ := ParseVersion(c.a)
		vb, _ := ParseVersion(c.b)
		if got := va.Compare(vb); got != c.want {
			t.Errorf("Compare(%s, %s) = %d，想要 %d", c.a, c.b, got, c.want)
		}
		if got := va.NewerThan(vb); got != (c.want > 0) {
			t.Errorf("NewerThan(%s, %s) = %v，想要 %v", c.a, c.b, got, c.want > 0)
		}
	}
}

// ---------- 发布与资产 ----------

// fakeReleaseJSON 一次"完整发布"：8 个发行资产（4 个 CLI + 4 个桌面）外加一个非制品。
// 名字与大小写就是发布契约（见 release.go 的 PickAssetFor）：TT 桌面版的安装包也叫
// -Setup.exe，靠 -desktop 区分；CLI 便携包与桌面便携包都叫 .zip，靠 -portable 区分。
var fakeReleaseJSON = `{
  "tag_name": "v0.2.2",
  "prerelease": false,
  "published_at": "2026-10-07T08:59:16Z",
  "assets": [
    {"name":"t100-v0.1.0.xml","size":10,"digest":"sha256:` + strings.Repeat("a", 64) + `","browser_download_url":"https://example.invalid/a.xml"},
    {"name":"TT-0.2.2-x64-Setup.exe","size":41297555,"digest":"sha256:` + strings.Repeat("b", 64) + `","browser_download_url":"https://example.invalid/x64-setup.exe"},
    {"name":"TT-0.2.2-arm64-Setup.exe","size":40297555,"digest":"sha256:` + strings.Repeat("c", 64) + `","browser_download_url":"https://example.invalid/arm64-setup.exe"},
    {"name":"tt-0.2.2-x64-portable.zip","size":123,"digest":"sha256:` + strings.Repeat("d", 64) + `","browser_download_url":"https://example.invalid/x64-portable.zip"},
    {"name":"tt-0.2.2-arm64-portable.zip","size":124,"digest":"sha256:` + strings.Repeat("e", 64) + `","browser_download_url":"https://example.invalid/arm64-portable.zip"},
    {"name":"TT-Desktop-0.2.2-x64-Setup.exe","size":61297555,"digest":"sha256:` + strings.Repeat("f", 64) + `","browser_download_url":"https://example.invalid/desktop-x64-setup.exe"},
    {"name":"TT-Desktop-0.2.2-arm64-Setup.exe","size":60297555,"digest":"sha256:` + strings.Repeat("a", 64) + `","browser_download_url":"https://example.invalid/desktop-arm64-setup.exe"},
    {"name":"TT-Desktop-0.2.2-x64.zip","size":223,"digest":"sha256:` + strings.Repeat("b", 64) + `","browser_download_url":"https://example.invalid/desktop-x64.zip"},
    {"name":"TT-Desktop-0.2.2-arm64.zip","size":224,"digest":"sha256:` + strings.Repeat("c", 64) + `","browser_download_url":"https://example.invalid/desktop-arm64.zip"}
  ]
}`

func TestParseReleaseAndPickAsset(t *testing.T) {
	rel, err := ParseRelease([]byte(fakeReleaseJSON))
	if err != nil {
		t.Fatalf("ParseRelease: %v", err)
	}
	v, err := rel.Version()
	if err != nil || v.String() != "0.2.2" {
		t.Fatalf("Version() = %v, %v；想要 0.2.2", v, err)
	}
	if rel.PublishedAt.IsZero() || rel.Prerelease {
		t.Errorf("发布时间/预发布标记没读出来: %+v", rel)
	}
	if len(rel.Assets) != 9 {
		t.Fatalf("该有 9 个资产（8 个发行资产 + 1 个非制品），得到 %d", len(rel.Assets))
	}

	// 8 个发行资产里，按形态 + 架构各挑各的。
	cases := []struct {
		kind Kind
		arch string
		want string
	}{
		{KindSetup, "x64", "TT-0.2.2-x64-Setup.exe"},
		{KindSetup, "arm64", "TT-0.2.2-arm64-Setup.exe"},
		{KindPortable, "x64", "tt-0.2.2-x64-portable.zip"},
		{KindPortable, "arm64", "tt-0.2.2-arm64-portable.zip"},
	}
	for _, c := range cases {
		got, err := PickAssetFor(rel.Assets, c.kind, c.arch)
		if err != nil || got.Name != c.want {
			t.Errorf("PickAssetFor(%s, %s) = %v, %v；想要 %s", c.kind, c.arch, got.Name, err, c.want)
		}
		// 桌面资产不许被误选：CLI 安装包的判据必须排掉 -desktop。
		if strings.Contains(strings.ToLower(got.Name), "desktop") {
			t.Errorf("PickAssetFor(%s, %s) 挑到了桌面版资产 %s", c.kind, c.arch, got.Name)
		}
	}

	// AssetFor 保留原来的调用形状（内部按 runtime.GOARCH 判架构）。
	if arch, err := ArchName(runtime.GOARCH); err == nil {
		got, err := rel.AssetFor(KindSetup)
		if err != nil || got.Name != "TT-0.2.2-"+arch+"-Setup.exe" {
			t.Errorf("AssetFor(安装版) = %v, %v；想要 TT-0.2.2-%s-Setup.exe", got.Name, err, arch)
		}
	}
	// 认不出的形态与桌面版都不许从发布里取资产（否则会给 KindOther 配一个"随便解压"
	// 的动作，或让桌面版吃到一份 CLI 载荷）。
	if _, err := rel.AssetFor(KindOther); err == nil {
		t.Error("AssetFor(KindOther) 该报错")
	}
	if _, err := PickAssetFor(rel.Assets, KindDesktop, "x64"); err == nil {
		t.Error("桌面版不该从发布里取资产")
	}

	// 多个命中：宁可报错也不猜，错误里要列出候选名与本机架构。
	multi := []Asset{
		{Name: "TT-0.2.1-x64-Setup.exe"},
		{Name: "TT-0.2.2-x64-Setup.exe"},
	}
	_, err = PickAssetFor(multi, KindSetup, "x64")
	if err == nil {
		t.Fatal("同类资产有两个时该报错")
	}
	for _, want := range []string{"多个", "x64", "TT-0.2.1-x64-Setup.exe", "TT-0.2.2-x64-Setup.exe"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("多个命中的错误信息该含 %q: %v", want, err)
		}
	}

	// 0 个命中：错误里要说出"这台机器是 <arch>"，并列出发布里到底有什么。
	desktopOnly := []Asset{
		{Name: "TT-Desktop-0.2.2-x64-Setup.exe"},
		{Name: "TT-Desktop-0.2.2-x64.zip"},
	}
	_, err = PickAssetFor(desktopOnly, KindSetup, "x64")
	if err == nil {
		t.Fatal("只有桌面资产时该报错")
	}
	for _, want := range []string{"这台机器是 x64", "TT-Desktop-0.2.2-x64-Setup.exe"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("0 个命中的错误信息该含 %q: %v", want, err)
		}
	}
	// 架构对不上也是 0 个命中：错误里必须点明本机架构，否则排障只能猜。
	_, err = PickAssetFor([]Asset{{Name: "TT-0.2.2-x64-Setup.exe"}}, KindSetup, "arm64")
	if err == nil || !strings.Contains(err.Error(), "arm64") {
		t.Errorf("架构对不上时该报错并写明本机架构，得到 %v", err)
	}

	// 一个都没有。
	rel3, _ := ParseRelease([]byte(`{"tag_name":"v1.0.0","assets":[]}`))
	if _, err := rel3.AssetFor(KindPortable); err == nil {
		t.Error("没有对应资产时该报错")
	}
}

// ArchName 是"发布只出 x64 与 arm64"这条约束的唯一出口：其它架构当场报错，
// 而不是拿空架构名去资产清单里碰运气。
func TestArchName(t *testing.T) {
	for goarch, want := range map[string]string{"amd64": "x64", "arm64": "arm64"} {
		got, err := ArchName(goarch)
		if err != nil || got != want {
			t.Errorf("ArchName(%q) = %q, %v；想要 %q", goarch, got, err, want)
		}
	}
	for _, bad := range []string{"386", "arm", "ppc64", ""} {
		if got, err := ArchName(bad); err == nil {
			t.Errorf("ArchName(%q) 该报错，得到 %q", bad, got)
		}
	}
}

func TestAssetHexDigest(t *testing.T) {
	want := strings.Repeat("0f", 32)
	if got := (Asset{Digest: "sha256:" + want}).HexDigest(); got != want {
		t.Errorf("HexDigest = %q，想要 %q", got, want)
	}
	// 没有摘要 / 摘要格式不对 → 空串，调用方据此拒绝安装。
	for _, bad := range []string{"", "sha256:", "md5:abc", "sha256:xyz", "sha256:" + strings.Repeat("a", 63)} {
		if got := (Asset{Digest: bad}).HexDigest(); got != "" {
			t.Errorf("Digest=%q 该得到空串，得到 %q", bad, got)
		}
	}
}

func TestParseReleasesSortsNewestFirst(t *testing.T) {
	b := `[
	  {"tag_name":"v0.1.0","published_at":"2025-01-01T00:00:00Z"},
	  {"tag_name":"v0.3.0","published_at":"2026-01-01T00:00:00Z","prerelease":true},
	  {"tag_name":"v0.2.0","published_at":"2025-06-01T00:00:00Z"}
	]`
	list, err := ParseReleases([]byte(b))
	if err != nil {
		t.Fatalf("ParseReleases: %v", err)
	}
	if len(list) != 3 || list[0].Tag != "v0.3.0" {
		t.Fatalf("排序不对: %+v", list)
	}
}

// ---------- 安装形态 ----------

func TestDetectKindAndInstallRefusal(t *testing.T) {
	base := t.TempDir()

	// 便携包：exe 同目录有 .portable。
	portable := filepath.Join(base, "tt-portable")
	if err := os.MkdirAll(portable, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(portable, ".portable"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DetectKind(filepath.Join(portable, "tt.exe")); got != KindPortable {
		t.Errorf("便携标记在时 DetectKind = %v，想要 KindPortable", got)
	}
	if why := InstallRefusal(KindPortable, filepath.Join(portable, "tt.exe"), "0.2.1"); why != "" {
		t.Errorf("便携包该允许自装，却被拒: %s", why)
	}

	// 安装版：exe 在 %LOCALAPPDATA%\Programs\TT（installer/tt-cli.iss 的 DefaultDirName，
	// 也是旧 MSI 用过的目录 —— 旧 MSI 用户升级后自然落进这一类）。
	local := filepath.Join(base, "Local")
	setupDir := filepath.Join(local, "Programs", "TT")
	if err := os.MkdirAll(setupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOCALAPPDATA", local)
	if got := DetectKind(filepath.Join(setupDir, "tt.exe")); got != KindSetup {
		t.Errorf("安装版目录里 DetectKind = %v，想要 KindSetup", got)
	}
	// 大小写与结尾分隔符不该影响判定（Windows 路径大小写不敏感）。
	if got := DetectKind(strings.ToUpper(filepath.Join(setupDir, "tt.exe"))); got != KindSetup {
		t.Errorf("大写路径 DetectKind = %v，想要 KindSetup", got)
	}
	if why := InstallRefusal(KindSetup, filepath.Join(setupDir, "tt.exe"), "0.2.1"); why != "" {
		t.Errorf("安装版该允许自装，却被拒: %s", why)
	}
	// 形态的名字与跨进程记号是两套东西，后者是计划文件的契约（见 kind.go）。
	if got := KindSetup.String(); got != "安装版" {
		t.Errorf("KindSetup.String() = %q，想要\"安装版\"", got)
	}
	if got := KindSetup.Token(); got != TokenSetup {
		t.Errorf("KindSetup.Token() = %q，想要 TokenSetup(\"setup\")", got)
	}
	if TokenSetup != "setup" {
		t.Errorf("TokenSetup = %q，跨进程记号必须是 \"setup\"", TokenSetup)
	}
	if KindPortable.Token() != "portable" || KindDesktop.Token() != "desktop" {
		t.Error("便携包/桌面版的跨进程记号不该跟着改名而变")
	}

	// 桌面版：exe 在 <安装目录>\resources\tt，同目录有 .desktop。
	// 它**不允许**自装（那份二进制要跟着桌面版整体升级）。
	desktopDir := filepath.Join(base, "Programs", "TT Desktop", "resources", "tt")
	if err := os.MkdirAll(desktopDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(desktopDir, config.DesktopMark), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DetectKind(filepath.Join(desktopDir, "tt.exe")); got != KindDesktop {
		t.Errorf("桌面版标记在时 DetectKind = %v，想要 KindDesktop", got)
	}
	deskWhy := InstallRefusal(KindDesktop, filepath.Join(desktopDir, "tt.exe"), "0.5.3")
	if deskWhy == "" || !strings.Contains(deskWhy, "TT Desktop") {
		t.Errorf("桌面版该拒绝自装并指向桌面版升级，得到 %q", deskWhy)
	}
	// 形态判定优先于版本号：一份没注入版本号的桌面载荷，提示也该指向桌面版升级
	// （否则会报成"源码态，用 make build 更新"——指错了地方）。
	if why := InstallRefusal(KindDesktop, filepath.Join(desktopDir, "tt.exe"), ""); !strings.Contains(why, "TT Desktop") {
		t.Errorf("没有版本号的桌面版该仍指向桌面版升级，得到 %q", why)
	}

	// 认不出的布局：有版本号也不让自装。
	other := filepath.Join(base, "somewhere")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := DetectKind(filepath.Join(other, "tt.exe")); got != KindOther {
		t.Errorf("认不出的布局 DetectKind = %v，想要 KindOther", got)
	}
	why := InstallRefusal(KindOther, filepath.Join(other, "tt.exe"), "0.2.1")
	if why == "" || !strings.Contains(why, "认不出") {
		t.Errorf("认不出的布局该给出解释，得到 %q", why)
	}

	// 源码态（没有注入版本号）：先于布局判定，提示 make build，并指向发行版的两种形态。
	why = InstallRefusal(KindOther, filepath.Join(other, "tt.exe"), "")
	if why == "" || !strings.Contains(why, "make build") {
		t.Errorf("没有版本号时该指向 make build，得到 %q", why)
	}
	if !strings.Contains(why, "安装版或便携包") {
		t.Errorf("源码态该指向安装版/便携包，得到 %q", why)
	}
}

// ---------- 检查与缓存 ----------

func TestCheckAndCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Error("GitHub API 要求带 User-Agent")
		}
		fmt.Fprint(w, fakeReleaseJSON)
	}))
	defer srv.Close()

	c, err := NewClient("", "tt-test")
	if err != nil {
		t.Fatal(err)
	}
	c.LatestAPI = srv.URL
	res, err := Check(context.Background(), c, "0.2.1", false)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !res.Newer || res.Latest != "0.2.2" || res.Current != "0.2.1" {
		t.Fatalf("检查结论不对: %+v", res)
	}
	if !strings.Contains(res.Message(), "0.2.2") {
		t.Errorf("给使用者的一句话该提到新版本号: %q", res.Message())
	}

	dataDir := t.TempDir()
	if err := SaveCheck(dataDir, res); err != nil {
		t.Fatalf("SaveCheck: %v", err)
	}
	// 检查结果落在缓存清单里的 update/ 下（不是数据目录根）。
	if _, err := os.Stat(filepath.Join(dataDir, "update", "check.json")); err != nil {
		t.Errorf("检查结果该落 <数据目录>/update/check.json: %v", err)
	}
	now := time.Now()
	if got := LoadCachedCheck(dataDir, "0.2.1", now); got == nil || got.Latest != "0.2.2" {
		t.Errorf("同一个版本号的缓存该被用上: %+v", got)
	}
	// 版本变了（升级过）→ 旧缓存不许再提示"有新版本"。
	if got := LoadCachedCheck(dataDir, "0.2.2", now); got != nil {
		t.Errorf("当前版本已变，缓存该作废，得到 %+v", got)
	}
	// 过期 → 作废。
	if got := LoadCachedCheck(dataDir, "0.2.1", now.Add(CheckTTL+time.Hour)); got != nil {
		t.Errorf("超过保鲜期，缓存该作废，得到 %+v", got)
	}
	// 时钟倒退（CheckedAt 在未来）也不许当新鲜缓存用。
	if got := LoadCachedCheck(dataDir, "0.2.1", now.Add(-2*time.Hour)); got != nil {
		t.Errorf("缓存时间是未来，该作废，得到 %+v", got)
	}
	// 没有缓存文件时不报错、不提示。
	if got := LoadCachedCheck(t.TempDir(), "0.2.1", now); got != nil {
		t.Errorf("没有缓存时该返回 nil，得到 %+v", got)
	}
	// 已经是最新时，Message 不该提示升级。
	if m := (&CheckResult{Current: "0.2.2", Latest: "0.2.2"}).Message(); m != "" {
		t.Errorf("已是最新时不该给提示，得到 %q", m)
	}
}

func TestCheckRejectsUnknownLocalVersion(t *testing.T) {
	c, err := NewClient("", "tt-test")
	if err != nil {
		t.Fatal(err)
	}
	// 本地没有版本号（go build 产物）→ 不该去问远端，当场报错说清。
	if _, err := Check(context.Background(), c, "", false); err == nil {
		t.Error("本地版本号为空时该报错")
	}
}

func TestLatestPreFallsBackToList(t *testing.T) {
	listHit := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/latest"):
			http.Error(w, "预发布不在 latest 里", http.StatusNotFound)
		default:
			listHit++
			fmt.Fprint(w, `[{"tag_name":"v0.1.0","published_at":"2025-01-01T00:00:00Z"},
			                 {"tag_name":"v0.3.0-rc","prerelease":true,"published_at":"2026-05-01T00:00:00Z"}]`)
		}
	}))
	defer srv.Close()

	c, err := NewClient("", "tt-test")
	if err != nil {
		t.Fatal(err)
	}
	c.LatestAPI = srv.URL + "/latest"
	c.ListAPI = srv.URL + "/list"
	if _, err := c.Latest(context.Background(), false); err == nil {
		t.Error("不带 --pre 时走 latest，这里该因 404 失败")
	}
	rel, err := c.Latest(context.Background(), true)
	if err != nil {
		t.Fatalf("带 --pre 该走列表: %v", err)
	}
	if rel.Tag != "v0.3.0-rc" || !rel.Prerelease {
		t.Errorf("--pre 该取到预发布，得到 %+v", rel)
	}
	if listHit != 1 {
		t.Errorf("列表接口被调用 %d 次，想要 1 次", listHit)
	}
}

// ---------- 下载与校验 ----------

func TestDownloadVerifiesDigest(t *testing.T) {
	payload := []byte("假装这是一个 40MB 的安装包")
	sum := sha256.Sum256(payload)
	goodHex := hex.EncodeToString(sum[:])

	// 第一次给对的内容，后面故意给坏内容（用来测摘要不符）。
	serve := payload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(serve)
	}))
	defer srv.Close()

	c, err := NewClient("", "tt-test")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	dest := filepath.Join(dir, "a.zip")
	a := Asset{Name: "tt-portable.zip", URL: srv.URL, Size: int64(len(payload)), Digest: "sha256:" + goodHex}

	var lastGot int64
	if err := c.Download(context.Background(), a, dest, goodHex, func(got, total int64) { lastGot = got }); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if b, err := os.ReadFile(dest); err != nil || string(b) != string(payload) {
		t.Fatalf("下到的内容不对: %v / %q", err, b)
	}
	if lastGot != int64(len(payload)) {
		t.Errorf("进度回调最后拿到 %d，想要 %d", lastGot, len(payload))
	}
	// 幂等：第二次不再请求（把服务换成 500，仍应成功）。
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "不该再被请求", http.StatusInternalServerError)
	})
	if err := c.Download(context.Background(), a, dest, goodHex, nil); err != nil {
		t.Errorf("已有合格文件时不该重新下载: %v", err)
	}

	// 摘要不符 → 报错，且不留半成品，也不碰目标文件。
	badSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("被掉包的内容"))
	}))
	defer badSrv.Close()
	bad := a
	bad.URL = badSrv.URL
	badDest := filepath.Join(dir, "bad.zip")
	err = c.Download(context.Background(), bad, badDest, goodHex, nil)
	if err == nil {
		t.Fatal("摘要不符时该报错")
	}
	if !strings.Contains(err.Error(), "校验不通过") {
		t.Errorf("错误信息该说清是校验问题: %v", err)
	}
	for _, p := range []string{badDest, badDest + ".part"} {
		if _, serr := os.Stat(p); serr == nil {
			t.Errorf("%s 不该留下", p)
		}
	}

	// 没有摘要 → 拒绝下载（门槛在 Asset.HexDigest）。
	if err := c.Download(context.Background(), a, filepath.Join(dir, "c.zip"), "", nil); err == nil {
		t.Error("没有摘要时该拒绝下载")
	}
}

func TestClientRejectsBadProxy(t *testing.T) {
	if _, err := NewClient("不是个地址", "tt-test"); err == nil {
		t.Error("代理地址不合法时该报错")
	}
}

func TestExplicitProxyPrecedence(t *testing.T) {
	// 参数 > 配置 > 空（空 = 交给环境变量规则）。
	if got := ExplicitProxy("http://flag:1", "http://cfg:2"); got != "http://flag:1" {
		t.Errorf("参数该优先，得到 %q", got)
	}
	if got := ExplicitProxy("", "http://cfg:2"); got != "http://cfg:2" {
		t.Errorf("配置该次之，得到 %q", got)
	}
	if got := ExplicitProxy("  ", "  "); got != "" {
		t.Errorf("两处都是空白该得到空串，得到 %q", got)
	}
}

// ---------- 状态机 ----------

func TestStateRoundTripAndInProgress(t *testing.T) {
	dir := t.TempDir()
	if LoadState(dir) != nil {
		t.Error("没有状态文件时该返回 nil")
	}
	st := &State{Phase: PhaseHandedOff, From: "0.2.1", Target: "0.2.2", Kind: "便携包", StartedAt: time.Now()}
	if err := SaveState(dir, st); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	got := LoadState(dir)
	if got == nil || got.Phase != PhaseHandedOff || got.Target != "0.2.2" {
		t.Fatalf("状态没读回来: %+v", got)
	}
	if !got.InProgress() {
		t.Error("handed-off 该算进行中")
	}
	for _, done := range []Phase{PhaseDone, PhaseFailed, PhaseIdle} {
		if (&State{Phase: done}).InProgress() {
			t.Errorf("%s 不该算进行中", done)
		}
	}
	// 状态文件是"不是缓存"的那类：落数据目录根，不进 update/ 子目录。
	if _, err := os.Stat(filepath.Join(dir, ".tt-update.json")); err != nil {
		t.Errorf("状态文件该落数据目录根的 .tt-update.json: %v", err)
	}
	// 日志可追写，且坏路径不 panic（日志写不进不该让升级失败）。
	AppendLog(LogPath(dir), "阶段 %s → %s", PhaseIdle, PhaseChecking)
	AppendLog("", "没有日志路径也不该炸")
	if b, err := os.ReadFile(LogPath(dir)); err != nil || !strings.Contains(string(b), "阶段 idle → checking") {
		t.Errorf("日志内容不对: %v / %q", err, b)
	}
}

// 缓存目录名必须真的在 config.CacheSubdirs 清单里 —— 不在的话 UpdateDir 返回空串，
// 所有下载与检查结果都会无处可落。这条钉住"名字改了就当场坏掉"这件事。
func TestUpdateDirIsInCacheList(t *testing.T) {
	dir := UpdateDir(t.TempDir())
	if dir == "" {
		t.Fatal("UpdateDir 返回空串：update/ 不在 config.CacheSubdirs 清单里")
	}
	if filepath.Base(dir) != "update" {
		t.Errorf("UpdateDir = %q，想要以 update 结尾", dir)
	}
}

// 发布 JSON 里字段缺失时不该崩：GitHub 的响应会变，我们只依赖少数几个字段。
func TestParseReleaseTolerantToMissingFields(t *testing.T) {
	rel, err := ParseRelease([]byte(`{"tag_name":"v1.0.0"}`))
	if err != nil {
		t.Fatalf("ParseRelease: %v", err)
	}
	if len(rel.Assets) != 0 || !rel.PublishedAt.IsZero() {
		t.Errorf("缺失字段该是零值: %+v", rel)
	}
	if _, err := ParseRelease([]byte(`{}`)); err == nil {
		t.Error("没有 tag_name 时该报错")
	}
	if _, err := ParseRelease([]byte(`不是 JSON`)); err == nil {
		t.Error("坏 JSON 该报错")
	}
	// 检查结果的 JSON 形态是契约（tt version 与设置页都读它），所以钉住字段名。
	b, _ := json.Marshal(&CheckResult{Current: "1.0.0", Latest: "1.1.0", Newer: true})
	for _, key := range []string{`"current"`, `"latest"`, `"newer"`, `"checkedAt"`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("检查结果的 JSON 里该有 %s: %s", key, b)
		}
	}
}
