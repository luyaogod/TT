package skill

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Source 是一条**要装出去的技能**：官方目录里的一个、用户登记且判定通过的一个，
// 或某个已启用插件贡献的一个。
type Source struct {
	Origin Origin
	Dir    string // 源目录（绝对路径）
	Name   string // 技能名 = SKILL.md 的 name = 源目录名
	// Plugin 是这条技能的来源插件 ID（Origin=OriginPlugin 时非空）。
	//
	// 只用于把话说清（"这个技能是哪个插件带来的"）。技能本身的身份仍然只有 name
	// 一个：装到平台目录里的副本不带任何插件标记，撤走时也只认平台清单。
	Plugin string
}

// Warning 是一条「这条技能没装成，但整次安装照常继续」的记录。
//
// 三条失败路径都落在这里：路径没了、内容不合规、跟别人撞名。用户明确要的是
// 「只警告，不影响其它技能安装」—— 因为第三方路径可能指向一个今天没插的移动硬盘，
// 插件包也可能被人手改过。
type Warning struct {
	Kind   State  `json:"kind"`            // missing / invalid / conflict
	Index  int    `json:"index,omitempty"` // 第三方登记的序号（1 起）
	Path   string `json:"path,omitempty"`
	Name   string `json:"name,omitempty"`
	Plugin string `json:"plugin,omitempty"` // 插件来源的告警带上插件 ID
	Why    string `json:"why"`
}

// Text 是给人看的一行，含修复动作。
//
// 命令行与设置页共用这一句：同一个问题两处各写一段文案，改一处忘一处之后，
// 用户看到的两个入口对同一件事给出不同说法。
func (w Warning) Text() string {
	var b strings.Builder
	switch {
	case w.Plugin != "":
		// 插件来源先报"是哪个插件带来的"：技能名由插件包决定，用户改不了。
		fmt.Fprintf(&b, "插件 %s 的技能", w.Plugin)
		if w.Name != "" {
			fmt.Fprintf(&b, " %s", w.Name)
		}
		if w.Path != "" {
			fmt.Fprintf(&b, "（%s）", w.Path)
		}
	case w.Index > 0 || w.Path != "" || w.Name != "":
		if w.Index > 0 {
			fmt.Fprintf(&b, "第 %d 条 ", w.Index)
		}
		if w.Path != "" {
			b.WriteString(w.Path)
			if w.Name != "" && w.Name != filepath.Base(w.Path) {
				fmt.Fprintf(&b, "（%s）", w.Name)
			}
		} else if w.Name != "" {
			b.WriteString(w.Name)
		}
	}
	fmt.Fprintf(&b, "：%s", w.Why)
	switch {
	case w.Plugin != "":
		b.WriteString("  —— 插件的技能来自它的版本目录；停用/升级该插件会重算，包里的 name 由作者决定")
	case w.Index > 0:
		b.WriteString("  —— 挪过位置就改 skills.externals 里的路径，删掉了就把这一条去掉")
	}
	return b.String()
}

// Plan 是一次安装的完整计划：要装哪些技能、官方树的顶层文件、以及所有告警。
//
// 计划与执行分开是为了 --dry-run：它只跑 BuildPlan 并把计划打印出来，一个字节都不写。
type Plan struct {
	Skills     []Source
	ExtraFiles []string  // 官方树顶层的非目录条目（README.md 等），原样复制
	Warnings   []Warning // 第三方/插件里没通过判定的那些，按登记顺序
}

// BuildPlan 组出安装计划（官方 + 用户登记的第三方）。官方树不合规 → 返回错误
// （包坏了，不是用户配错了）；第三方的任何问题都降级成 Warnings。
//
// externals 是配置里 skills.externals 的顺序，顺序即优先级（同名时先登记者赢）。
//
// 行为与引入插件来源之前**完全一致**：插件那一侧传空，见 BuildPlanWithPluginSkills。
func BuildPlan(officialDir string, externals []string) (Plan, error) {
	return buildPlan(officialDir, externals)
}

// BuildPlanWithPluginSkills 在 BuildPlan 之上追加**已启用插件**贡献的技能。
//
// 追加顺序就是优先级：官方 → 用户登记的第三方 → 插件。插件排最后是有意的：
// 官方技能是发版包的一部分，第三方是用户显式登记的选择，插件的技能是"启用插件
// 顺带带来的"；同一名字出现多次时先出现的赢。
//
// 插件侧的每一项问题（目录没了、SKILL.md 坏了、与已计划的技能撞名）都只降级成
// Warnings（与第三方同一策略）：插件包可能被手改、版本目录可能被删，那不该让官方
// 与第三方技能装不上，也不该阻断插件自己的 CLI/UI（那两条线读的是注册表，不是计划）。
//
// pluginSkills 由调用方从插件注册表投影出来（Origin=OriginPlugin、Dir 为活跃版本
// 目录内的绝对路径、Plugin 为插件 ID）；本函数不读注册表、不认配置 —— 那保持
// internal/skill 是纯管线的性质。
func BuildPlanWithPluginSkills(officialDir string, externals []string, pluginSkills []Source) (Plan, error) {
	p, err := buildPlan(officialDir, externals)
	if err != nil {
		return Plan{}, err
	}
	if len(pluginSkills) == 0 {
		return p, nil
	}

	// taken 记「这个名字已经被谁占了」：插件永远排最后，所以它撞上的一定是官方、
	// 先登记的第三方，或排在它前面的插件。
	taken := make(map[string]string, len(p.Skills))
	for _, s := range p.Skills {
		taken[s.Name] = ownerLabel(s)
	}
	for _, raw := range pluginSkills {
		src := raw
		src.Origin = OriginPlugin
		info := InspectPlugin(src)
		if info.State != StateOK {
			p.Warnings = append(p.Warnings, Warning{
				Kind: info.State, Path: info.Dir, Name: info.Name,
				Plugin: src.Plugin, Why: info.Why,
			})
			continue
		}
		if prev, ok := taken[info.Name]; ok {
			// 撞名只告警、不失败：插件自己的 CLI/UI 照常挂载，只是这个技能不装出去。
			p.Warnings = append(p.Warnings, Warning{
				Kind: StateConflict, Path: info.Dir, Name: info.Name, Plugin: src.Plugin,
				Why: fmt.Sprintf("与%s同名，先出现的优先，已跳过", prev),
			})
			continue
		}
		taken[info.Name] = ownerLabel(Source{Origin: OriginPlugin, Plugin: src.Plugin})
		p.Skills = append(p.Skills, Source{
			Origin: OriginPlugin, Dir: info.Dir, Name: info.Name, Plugin: src.Plugin,
		})
	}
	return p, nil
}

// ownerLabel 给"这个技能已经是谁的了"一句人话（冲突文案里用）。
func ownerLabel(s Source) string {
	switch s.Origin {
	case OriginOfficial:
		return "官方技能"
	case OriginPlugin:
		if s.Plugin != "" {
			return "插件 " + s.Plugin + " 的技能"
		}
		return "插件带来的技能"
	default:
		return "已登记的第三方技能"
	}
}

// buildPlan 是唯一的实现：BuildPlan 与 BuildPlanWithPluginSkills 都走这里，
// 保证"没有插件时"两条入口给出同一份计划（P4-03 的兼容要求）。
func buildPlan(officialDir string, externals []string) (Plan, error) {
	officialDir = absOr(officialDir)
	official, extraFiles, err := InspectOfficial(officialDir)
	if err != nil {
		return Plan{}, err
	}

	names := make([]string, 0, len(official))
	for _, s := range official {
		names = append(names, s.Name)
	}

	p := Plan{ExtraFiles: extraFiles}
	p.Skills = make([]Source, 0, len(official)+len(externals))
	for _, s := range official {
		p.Skills = append(p.Skills, Source{Origin: s.Origin, Dir: s.Dir, Name: s.Name})
	}
	for _, info := range InspectExternal(externals, names) {
		if info.State == StateOK {
			p.Skills = append(p.Skills, Source{Origin: info.Origin, Dir: info.Dir, Name: info.Name})
			continue
		}
		p.Warnings = append(p.Warnings, Warning{
			Kind: info.State, Index: info.Index, Path: info.Dir, Name: info.Name, Why: info.Why,
		})
	}
	// 顶层文件按名字排一下：ReadDir 已经排过，但计划是给人看的，别指望上游。
	sort.Strings(p.ExtraFiles)
	return p, nil
}

func absOr(p string) string {
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}
