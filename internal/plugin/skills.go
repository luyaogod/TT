package plugin

// 插件贡献的技能目录：只做「注册表 + 版本清单」的投影，**不判定技能是否合规**。
//
// 为什么不在这个包里判：frontmatter、目录名、assets 指针与撞名规则只有 internal/skill
// 一份实现（P4-04）。这里多写一套就会养出两种口径，而两套口径的后果是"命令行说能装、
// 设置页说不能装"。
//
// 本包也不 import internal/skill：插件管理面只管元数据与载荷，技能怎么装出去是命令层
// 接线的事（见 internal/cli/skills_sync.go）。返回的只是路径，谁校验由调用方决定。
//
// 落点口径（P4-02）：目录一律取**活跃版本目录**（Store.ActiveDir）里的贡献目录 ——
// 注册表已记录版本与启停状态，升级换版本后这里自然指向新版本，不需要另存一份技能清单。
// 反过来，**能从平台目录删除什么只由 .agents/.tt-manifest.json 决定**：注册表只说明
// "哪些技能应当出现"，撤销安装也只删平台清单记过的条目。

import "path/filepath"

// SkillSource 一个已启用插件贡献的技能目录。
type SkillSource struct {
	Plugin string `json:"plugin"`
	// Dir 是活跃版本目录内的**绝对路径**（由包内相对路径拼出并确认没有越界）。
	Dir string `json:"dir"`
}

// EnabledSkillSources 列出全部已启用插件贡献的技能目录（按插件 ID 稳定排序）。
//
// 顺序取注册表的 Enabled()（已按 ID 排序），插件内取清单里的声明顺序；计划里插件
// 之间同名的判定也就按这个顺序"先出现者赢"。
//
// 读不出清单/版本目录的插件直接跳过：启用时已经校验过一遍，此时读不出来说明包被人
// 动过；跳过它不影响其它插件，也不该让整份技能计划失败。
func (s *Store) EnabledSkillSources() ([]SkillSource, error) {
	reg, err := s.Load()
	if err != nil {
		return nil, err
	}
	var out []SkillSource
	for _, rec := range reg.Enabled() {
		dir := s.ActiveDir(rec)
		if dir == "" {
			continue
		}
		man, merr := LoadManifest(dir)
		if merr != nil {
			continue
		}
		rels, rerr := man.SkillDirs()
		if rerr != nil {
			continue
		}
		for _, rel := range rels {
			// rel 已由 normalizeRelPath 校验（无绝对路径/盘符/..）；这里再把拼出来的
			// 绝对路径关回版本目录一次 —— 路径拼装是最后一道门，不靠上游的自觉。
			full := filepath.Join(dir, filepath.FromSlash(rel))
			if !withinDir(dir, full) {
				continue
			}
			out = append(out, SkillSource{Plugin: rec.ID, Dir: full})
		}
	}
	return out, nil
}

// DisabledSkillNames 返回「已安装但未启用」的插件贡献的技能名集合。
//
// 用途只有一个：期望态重算时判断平台清单里的某个技能该不该撤走。启用的插件不需要它
// （它们的技能还在计划里）；它专治停用与卸载 —— 此时源已经不在计划里，而"源被撤走"
// 与"源暂时不可用"（移动硬盘没插）在盘上长得一模一样，只有注册表说得清是哪种。
//
// 名字取目录名：skill 侧的合规规则要求 SKILL.md 的 name 等于目录名，两边因此同名。
func (s *Store) DisabledSkillNames() (map[string]bool, error) {
	reg, err := s.Load()
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, rec := range reg.Plugins {
		if rec.Enabled {
			continue
		}
		dir := s.ActiveDir(rec)
		if dir == "" {
			continue
		}
		man, merr := LoadManifest(dir)
		if merr != nil {
			continue
		}
		rels, rerr := man.SkillDirs()
		if rerr != nil {
			continue
		}
		for _, rel := range rels {
			full := filepath.Join(dir, filepath.FromSlash(rel))
			if !withinDir(dir, full) {
				continue
			}
			out[filepath.Base(full)] = true
		}
	}
	return out, nil
}
