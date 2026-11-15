package update

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Repo 更新源。**写死在代码里**，不给命令行开关：更新器是一个"下载可执行文件并运行它"
// 的路径，让它能指向别处等于把这条路径交给调用参数。
const (
	RepoSlug  = "luyaogod/TT"
	RepoPage  = "https://github.com/" + RepoSlug
	apiBase   = "https://api.github.com/repos/" + RepoSlug
	LatestAPI = apiBase + "/releases/latest"
	// ListAPI 列发布（`--pre` 用它：/releases/latest 有意跳过预发布）。
	ListAPI = apiBase + "/releases?per_page=20"
)

// Asset 一个发布资产。Digest 形如 "sha256:<hex>"。
//
// **Digest 为空就不许安装**（只允许 check）：没有摘要就没有"下载到的字节 == 发布方
// 发布的字节"这条判据，而这是整条路径上唯一不靠信任的判断。
type Asset struct {
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Digest string `json:"digest"`
	URL    string `json:"browser_download_url"`
}

// HexDigest 取出 sha256 的十六进制部分；不是 sha256 或格式不对时返回空串。
func (a Asset) HexDigest() string {
	s, ok := strings.CutPrefix(a.Digest, "sha256:")
	if !ok {
		return ""
	}
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) != 64 {
		return ""
	}
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return ""
		}
	}
	return s
}

// Release 一次发布。
type Release struct {
	Tag         string    `json:"tag_name"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []Asset   `json:"assets"`
}

// Version 取发布对应的版本号（tag 的 v 前缀由 ParseVersion 容忍）。
func (r *Release) Version() (Version, error) { return ParseVersion(r.Tag) }

// ParseRelease 解析 GitHub 的发布 JSON。
//
// 只读它需要的字段，多余字段一律忽略：GitHub 的响应体很大，而我们要的是"这个发布是
// 谁、有哪些资产、摘要是什么"。
func ParseRelease(b []byte) (*Release, error) {
	var r Release
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("解析发布信息失败: %w", err)
	}
	if r.Tag == "" {
		return nil, fmt.Errorf("发布信息里没有 tag_name")
	}
	return &r, nil
}

// ParseReleases 解析 `/releases`（数组），按发布时间从新到旧排序。
func ParseReleases(b []byte) ([]*Release, error) {
	var list []*Release
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, fmt.Errorf("解析发布列表失败: %w", err)
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].PublishedAt.After(list[j].PublishedAt) })
	return list, nil
}

// AssetFor 按安装形态与本机架构挑出该下载哪个资产（见 PickAsset / PickAssetFor）。
func (r *Release) AssetFor(k Kind) (Asset, error) { return PickAsset(r.Assets, k) }

// PickAsset 按安装形态与 runtime.GOARCH 挑资产。不想依赖本机架构时用 PickAssetFor。
func PickAsset(assets []Asset, k Kind) (Asset, error) {
	arch, err := ArchName(runtime.GOARCH)
	if err != nil {
		return Asset{}, err
	}
	return PickAssetFor(assets, k, arch)
}

// ArchName 把 runtime.GOARCH 翻成发行资产名里的架构记号。
//
// 发布只出 x64 一种：其它架构不是"这次没找到资产"，而是这一版根本不支持它，所以在这里
// 就报错误，而不是拿一个空架构名去资产清单里碰运气。arm64 的名字保留在映射里，是为了
// 将来真要补 arm64 包时只改打包脚本、不改更新器。
func ArchName(goarch string) (string, error) {
	switch goarch {
	case "amd64":
		return "x64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("发布只提供 x64，不支持本机的 %s", goarch)
	}
}

// PickAssetFor 按安装形态与架构挑出该下载哪个资产（arch 取 ArchName 的产物）。
//
// **资产名是发布契约的一部分**：安装脚本按同一套名字出包，更新器按同一套名字回挑，
// 两边任何一处改了都会在这里对不上。0.6.0 起一次发布固定 4 个资产（x64）：
//
//	CLI 安装版：  TT-<ver>-x64-Setup.exe
//	CLI 便携版：  tt-<ver>-x64-portable.zip
//	桌面安装版：  TT-Desktop-<ver>-x64-Setup.exe
//	桌面便携版：  TT-Desktop-<ver>-x64.zip
//
// 只按扩展名挑已经不够用了：发布里有 2 个 .zip、2 个 .exe（CLI 与桌面版都在里面），
// "多个同类就报错"必然命中。所以判据是**形态 + 架构**：`-Setup.exe` 且不含 `-desktop`
// 才是 CLI 安装包，`-portable.zip` 且含 `-<arch>-` 才是 CLI 便携包。桌面版不从发布里
// 取资产（它随 TT Desktop 整体升级）。名字大小写不敏感（大小写在契约里是固定的，但
// 比较时不值得为此拒绝一份只错了大小写的包）。
func PickAssetFor(assets []Asset, k Kind, arch string) (Asset, error) {
	arch = strings.ToLower(arch)
	match := func(name string) bool { return false }
	switch k {
	case KindSetup:
		// CLI 安装包：TT-<ver>-<arch>-Setup.exe。桌面安装包也叫 -Setup.exe，所以必须
		// 排掉带 -desktop 的那些。
		match = func(name string) bool {
			n := strings.ToLower(name)
			return strings.HasSuffix(n, "-setup.exe") &&
				strings.Contains(n, "-"+arch+"-") &&
				!strings.Contains(n, "-desktop")
		}
	case KindPortable:
		// CLI 便携包：tt-<ver>-<arch>-portable.zip。桌面便携包叫 .zip 但不带 -portable。
		match = func(name string) bool {
			n := strings.ToLower(name)
			return strings.HasSuffix(n, "-portable.zip") && strings.Contains(n, "-"+arch+"-")
		}
	case KindDesktop:
		return Asset{}, fmt.Errorf("桌面版不从发布里取资产：它随 TT Desktop 整体升级")
	default:
		return Asset{}, fmt.Errorf("安装形态 %s 不从发布里取资产", k)
	}

	var found []Asset
	for _, a := range assets {
		if match(a.Name) {
			found = append(found, a)
		}
	}
	if len(found) == 1 {
		return found[0], nil
	}
	// 0 个：把发布里到底有什么列出来；多个：列出命中的那几个。两种都要带上本机架构，
	// 否则排障时只能靠猜"是不是下错了架构的包"。
	names := assetNames(assets)
	if len(found) > 0 {
		names = assetNames(found)
	}
	if len(names) == 0 {
		names = []string{"（没有任何资产）"}
	}
	kind := k.String()
	if len(found) > 1 {
		return Asset{}, fmt.Errorf("发布里有多个匹配 %s（这台机器是 %s）的资产：%s；不知道该装哪个",
			kind, arch, strings.Join(names, " / "))
	}
	return Asset{}, fmt.Errorf("发布里没有匹配 %s（这台机器是 %s）的资产；发布里有：%s",
		kind, arch, strings.Join(names, " / "))
}

// assetNames 取资产名（报错用）。
func assetNames(assets []Asset) []string {
	names := make([]string, 0, len(assets))
	for _, a := range assets {
		names = append(names, a.Name)
	}
	return names
}
