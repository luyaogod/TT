package plugin

// 插件发布索引：**与 TT 自更新完全分开**的一份契约（P6-01）。
//
// 为什么不复用 internal/update 的发布模型：那里固定了 GitHub 仓库、`tag_name`、按扩展名
// 挑 `.msi`/`.zip` 的整包规则 —— 那是一台机器装一份 TT 的语义。插件是"同一个 ID 有很多
// 版本、每个版本每种平台一个资产"，判据是 id + version + {os}-{arch}，两者没有可共用的
// 决策，硬套只会让两边都变形。可共用的只有字节级原语，那部分已抽到 internal/fetch。

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"

	"tt/internal/fetch"
)

// IndexSchemaVersion 发布索引的结构版本。更高的主版本一律拒绝（不猜字段含义）。
const IndexSchemaVersion = 1

// ReleaseIndex 一份发布索引。
//
// 形状（见 .agents/docs/plugin-system/13-release-index.md）：
//
//	{"schemaVersion":1,"plugins":[{"id":"com.example.p","versions":[
//	  {"version":"0.2.0","hostApi":">=1.0.0 <2.0.0",
//	   "assets":{"windows-amd64":{"url":"https://…/p.zip","sha256":"…","size":123}}}]}]}
type ReleaseIndex struct {
	SchemaVersion int           `json:"schemaVersion"`
	Plugins       []IndexPlugin `json:"plugins"`
}

// IndexPlugin 一个插件在索引里的全部版本。
type IndexPlugin struct {
	ID       string         `json:"id"`
	Versions []IndexVersion `json:"versions"`
}

// IndexVersion 一个可发布的版本。
type IndexVersion struct {
	Version string `json:"version"`
	// HostAPI 该版本兼容的 Host API 范围（与清单里的 hostApi 同一种写法）。
	HostAPI string `json:"hostApi"`
	// Assets 按 {os}-{arch} 分平台的资产。
	Assets map[string]IndexAsset `json:"assets"`
}

// IndexAsset 一个平台资产。
//
// sha256 与 size 都**必填**：没有摘要就没有"下到的字节 == 发布方发布的字节"这条判据；
// 没有大小就没法在解包前先挡掉明显不对的包。
type IndexAsset struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// AssetKey 当前平台在索引里的键（windows-amd64 / linux-arm64 …）。
func AssetKey() string { return runtime.GOOS + "-" + runtime.GOARCH }

// AssetFor 取当前平台的资产。
func (v *IndexVersion) AssetFor() (IndexAsset, bool) {
	if v == nil {
		return IndexAsset{}, false
	}
	a, ok := v.Assets[AssetKey()]
	return a, ok
}

// IndexError 索引内容不合法（整份索引被拒，不是"某个版本跳过"）。
type IndexError struct{ msg string }

func (e *IndexError) Error() string { return e.msg }

func indexErrf(format string, a ...any) error {
	return &IndexError{msg: "发布索引不合法：" + fmt.Sprintf(format, a...)}
}

// IsIndexError 是否是索引内容错误。
func IsIndexError(err error) bool {
	var ie *IndexError
	return errors.As(err, &ie)
}

// ParseReleaseIndex 严格解码并校验一份发布索引。
//
// 为什么严格：索引决定"下哪个包、跑哪个可执行文件"。写错的键（url vs href）被静默忽略
// 会变成"索引看着有、下载时才发现没地址"。allowInsecure 只给测试用（httptest 的普通
// HTTP 服务端）；生产命令一律不放开。
func ParseReleaseIndex(b []byte, allowInsecure bool) (*ReleaseIndex, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var idx ReleaseIndex
	if err := dec.Decode(&idx); err != nil {
		return nil, indexErrf("不是合法 JSON 或含未知字段: %v", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, indexErrf("第一个 JSON 对象之后还有多余内容")
	}
	if idx.SchemaVersion != IndexSchemaVersion {
		return nil, indexErrf("schemaVersion = %d，本版 TT 只支持 %d（索引来自更新的 TT 时请先升级 tt）",
			idx.SchemaVersion, IndexSchemaVersion)
	}
	seen := map[string]bool{}
	for pi := range idx.Plugins {
		p := &idx.Plugins[pi]
		if err := ValidateID(p.ID); err != nil {
			return nil, indexErrf("plugins[%d]: %v", pi, err)
		}
		if seen[p.ID] {
			return nil, indexErrf("插件 %q 出现了两次", p.ID)
		}
		seen[p.ID] = true
		if len(p.Versions) == 0 {
			return nil, indexErrf("插件 %q 没有任何版本", p.ID)
		}
		vseen := map[string]bool{}
		for vi := range p.Versions {
			v := &p.Versions[vi]
			if err := ValidateSemver(v.Version); err != nil {
				return nil, indexErrf("插件 %q 的 %v", p.ID, err)
			}
			if vseen[v.Version] {
				return nil, indexErrf("插件 %q 的版本 %s 出现了两次（不知道该信哪一个）", p.ID, v.Version)
			}
			vseen[v.Version] = true
			if err := ValidateHostAPIRange(v.HostAPI); err != nil {
				return nil, indexErrf("插件 %q 的版本 %s: %v", p.ID, v.Version, err)
			}
			if len(v.Assets) == 0 {
				return nil, indexErrf("插件 %q 的版本 %s 没有任何平台资产", p.ID, v.Version)
			}
			for key, a := range v.Assets {
				if !validAssetKey(key) {
					return nil, indexErrf("插件 %q 的版本 %s: 资产键 %q 不是 {os}-{arch} 形态", p.ID, v.Version, key)
				}
				if err := validateIndexAsset(a, allowInsecure); err != nil {
					return nil, indexErrf("插件 %q 的版本 %s 的 %s 资产: %v", p.ID, v.Version, key, err)
				}
			}
		}
	}
	return &idx, nil
}

// validateIndexAsset 校验一个资产的三要素。
func validateIndexAsset(a IndexAsset, allowInsecure bool) error {
	if strings.TrimSpace(a.URL) == "" {
		return errors.New("url 不能为空")
	}
	if err := fetch.RequireHTTPS(a.URL, allowInsecure); err != nil {
		return err
	}
	if !isHex64(a.SHA256) {
		return fmt.Errorf("sha256 %q 不是 64 位十六进制摘要", a.SHA256)
	}
	if a.Size <= 0 {
		return fmt.Errorf("size = %d，必须给出正整数字节数", a.Size)
	}
	return nil
}

// validAssetKey 资产键必须是 {os}-{arch}（小写，两段）。
func validAssetKey(k string) bool {
	parts := strings.Split(k, "-")
	if len(parts) != 2 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if !(c >= 'a' && c <= 'z') && !(c >= '0' && c <= '9') {
				return false
			}
		}
	}
	return true
}

// isHex64 是否是 64 位十六进制。
func isHex64(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

// Find 按 ID 找索引条目。
func (idx *ReleaseIndex) Find(id string) *IndexPlugin {
	if idx == nil {
		return nil
	}
	for i := range idx.Plugins {
		if idx.Plugins[i].ID == id {
			return &idx.Plugins[i]
		}
	}
	return nil
}

// Version 按版本号找条目。
func (p *IndexPlugin) Version(v string) *IndexVersion {
	for i := range p.Versions {
		if p.Versions[i].Version == v {
			return &p.Versions[i]
		}
	}
	return nil
}

// SortedVersions 版本从新到旧（展示与"有哪些可升"用）。
func (p *IndexPlugin) SortedVersions() []IndexVersion {
	out := append([]IndexVersion(nil), p.Versions...)
	sort.SliceStable(out, func(i, j int) bool {
		a, ea := semver.StrictNewVersion(out[i].Version)
		b, eb := semver.StrictNewVersion(out[j].Version)
		if ea != nil || eb != nil {
			return out[i].Version > out[j].Version
		}
		return a.GreaterThan(b)
	})
	return out
}

// Pick 选这次要升到哪个版本。
//
// 规则（按优先级）：
//  1. pin 非空：必须是它。不兼容当前 Host API、没有本平台资产、或就是当前版本 → 明确报错
//     （用户点名了一个版本，静默换成别的版本比失败更难查）。
//  2. 否则取**比 current 新**、与 hostAPI 兼容、且有本平台资产的最高版本。
//  3. 一个都没有 → (nil, nil)，表示"已是最新或没有可用的新版"，调用方按 upToDate 处理。
//
// 已经下载到本机但**没激活**的版本仍会被选中：那条路径不再下载，只有一次切换（见
// Manager.updateOne）。否则"升级失败 → 重试"会永远得到"已是最新"，把人钉死在旧版本上。
func (p *IndexPlugin) Pick(current, hostAPI, pin string) (*IndexVersion, error) {
	if pin != "" {
		v := p.Version(pin)
		if v == nil {
			return nil, fmt.Errorf("发布索引里没有 %s 的版本 %s（可用: %s）",
				p.ID, pin, strings.Join(p.versionStrings(), " / "))
		}
		if ok, err := HostAPISatisfies(v.HostAPI, hostAPI); err != nil || !ok {
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("版本 %s 要求 hostApi %q，当前 TT 的 Host API 是 %s", pin, v.HostAPI, hostAPI)
		}
		if _, ok := v.AssetFor(); !ok {
			return nil, fmt.Errorf("版本 %s 没有 %s 平台的资产（索引里有: %s）",
				pin, AssetKey(), strings.Join(assetKeys(v), " / "))
		}
		if pin == current {
			return nil, fmt.Errorf("%s 已经是 %s，不用升", p.ID, pin)
		}
		return v, nil
	}

	cur, err := semver.StrictNewVersion(current)
	if err != nil {
		// 当前版本不合法（注册表被手改过）：不猜"要不要升"，直接报出来。
		return nil, fmt.Errorf("已装版本 %q 不是严格 SemVer，无法判断该升到哪一版: %w", current, err)
	}
	var best *IndexVersion
	var bestVer *semver.Version
	for i := range p.Versions {
		v := &p.Versions[i]
		sv, verr := semver.StrictNewVersion(v.Version)
		if verr != nil {
			continue
		}
		if !sv.GreaterThan(cur) {
			continue
		}
		if ok, herr := HostAPISatisfies(v.HostAPI, hostAPI); herr != nil || !ok {
			continue
		}
		if _, ok := v.AssetFor(); !ok {
			continue
		}
		if best == nil || sv.GreaterThan(bestVer) {
			best, bestVer = v, sv
		}
	}
	return best, nil
}

// versionStrings 索引里登记的全部版本（排序后）。
func (p *IndexPlugin) versionStrings() []string {
	sorted := p.SortedVersions()
	out := make([]string, 0, len(sorted))
	for i := range sorted {
		out = append(out, sorted[i].Version)
	}
	return out
}

// assetKeys 一个版本声明的平台键（排序）。
func assetKeys(v *IndexVersion) []string {
	out := make([]string, 0, len(v.Assets))
	for k := range v.Assets {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
