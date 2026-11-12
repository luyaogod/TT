package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// ManifestRel 是清单的落点（相对分发根目录：项目目录或用户目录）。
//
// 放在 `.agents/` 而不是项目根：它描述的就是这棵树的部署事实，跟 README/AGENTS.md
// 混在根目录只会更难找。文件名刻意不叫 `.tt-skills.json` —— 配置目录里已经有一个同名
// 的版本戳（update.SkillsStamp），两个不同形状的文件同名会让人以为是同一份。
const ManifestRel = ".agents/.tt-manifest.json"

// manifestSchema 是清单的结构版本；本版只有 1。
const manifestSchema = 1

// Manifest 只记“哪个平台装了哪些技能”。
//
// 不记版本、哈希、时间：本功能不做版本检测，同步就是覆盖；清单的唯一职责是
// “这些目录是 TT 装的，取消勾选时可以删”。少了这些字段，盘上这份文件就不会
// 因为升级/重装/换机器而变得“看起来过期”。
type Manifest struct {
	Schema  int                 `json:"schema"`
	Targets map[string][]string `json:"targets,omitempty"`
}

// ManifestPath 返回某个项目的清单路径。
func ManifestPath(project string) string {
	return filepath.Join(project, filepath.FromSlash(ManifestRel))
}

// LoadManifest 读清单。不存在时返回空清单 + exists=false，不是错误。
func LoadManifest(project string) (m Manifest, exists bool, err error) {
	p := ManifestPath(project)
	b, rerr := os.ReadFile(p)
	if rerr != nil {
		if os.IsNotExist(rerr) {
			return Manifest{Schema: manifestSchema}, false, nil
		}
		return Manifest{}, false, fmt.Errorf("读技能清单失败 %s: %w", p, rerr)
	}
	if uerr := json.Unmarshal(b, &m); uerr != nil {
		return Manifest{}, true, fmt.Errorf("技能清单格式不对 %s: %w（修好或删掉它再试）", p, uerr)
	}
	if m.Targets == nil {
		m.Targets = map[string][]string{}
	}
	return m, true, nil
}

// Skills 返回某个平台记过的技能名（副本）。没有记录时是空数组而不是 nil：
// JSON 里 `null` 和 `[]` 对调用方是两种形状，前者会让“有没有”得再判一次。
func (m Manifest) Skills(targetID string) []string {
	list := m.Targets[targetID]
	if list == nil {
		return []string{}
	}
	return append([]string(nil), list...)
}

// Set 更新某个平台的技能清单，返回是否真的变了。
//
// 空列表 = 这个平台不再由 TT 管（删掉键，不留 `[]`）。名字排序去重，让重复同步写出
// 来的字节完全一致。
func (m *Manifest) Set(targetID string, names []string) bool {
	next := normalizeSkillNames(names)
	if m.Targets == nil {
		m.Targets = map[string][]string{}
	}
	prev := m.Targets[targetID]
	if prev == nil && len(next) == 0 {
		return false
	}
	if slices.Equal(prev, next) {
		return false
	}
	if len(next) == 0 {
		delete(m.Targets, targetID)
	} else {
		m.Targets[targetID] = next
	}
	return true
}

// Save 原子写清单。只写项目相对信息，可直接入 git。
func (m Manifest) Save(project string) error {
	p := ManifestPath(project)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("建 .agents 目录失败: %w", err)
	}
	out := m
	out.Schema = manifestSchema
	if out.Targets == nil {
		out.Targets = map[string][]string{}
	}
	for id, names := range out.Targets {
		out.Targets[id] = normalizeSkillNames(names)
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("编码技能清单失败: %w", err)
	}
	b = append(b, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(p), ".tt-manifest-*.tmp")
	if err != nil {
		return fmt.Errorf("建临时清单失败: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // rename 成功后这里已经是 no-op
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("写临时清单失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("收尾临时清单失败: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("设置清单权限失败: %w", err)
	}
	if err := os.Rename(tmpName, p); err != nil {
		return fmt.Errorf("替换技能清单失败 %s: %w", p, err)
	}
	return nil
}

// normalizeSkillNames 排序、去重，并丢掉不合规的名字。
//
// 名字不合规的条目只可能来自手改的清单；留着它会在删除时拼出目录外的路径。
func normalizeSkillNames(names []string) []string {
	seen := make(map[string]bool, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || seen[n] || !ValidName(n) {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return nil
	}
	return out
}
