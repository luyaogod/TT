// 插件技能与「分发」管线的接线处：计划来源的组装 + 状态变更后的期望态重算。
//
// 为什么接在这一层：
//   - internal/plugin 管注册表与程序载荷，不该知道"技能怎么装出去"（它连
//     internal/skill 都不 import）；
//   - internal/skill 是纯管线：不读配置、不认插件、不打印；
//   - 命令层是唯一同时看得见两边的地方 —— 它把"启用插件的技能目录"翻成计划来源。
//
// 删除纪律（P4-02 / P4-10）：平台目录里能够删除什么**只由 .agents/.tt-manifest.json
// 决定**。注册表只说明"哪些技能应当出现"；撤销只发生在清单记过、且源确实撤走了的
// 条目上，清单外的目录（用户自己建的、别的工具装的）永远不动。插件技能也**永远不写进
// <exe>/skills**：那条路径是发版包，计划只从它读，不往它写。
package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"tt/internal/cli/common"
	"tt/internal/config"
	"tt/internal/plugin"
	"tt/internal/skill"
)

// errNoOfficialSkills 官方技能树不在（源码构建、或这个包没带 skills/）。
//
// 单独一个哨兵是为了把"这台机器没有技能管线"与"技能管线坏了"分开：前者下插件启停
// 不该失败（本来就没装过技能），后者要如实报错。
var errNoOfficialSkills = errors.New("找不到 skills 目录")

// officialSkillsDir 解析 <exe 目录>/skills 并要求它存在。
func officialSkillsDir() (string, error) {
	src, err := skillsSource()
	if err != nil {
		return "", err
	}
	if !isDir(src) {
		return "", fmt.Errorf("%w: %s\n\nskills/ 与 tt.exe 同目录（免安装包里自带）；"+
			"从源码运行时先执行 make build（产物 tt.exe 与 skills/ 同目录）", errNoOfficialSkills, src)
	}
	return src, nil
}

// buildSkillsPlan 组一份安装计划：官方 + 用户登记的第三方 + 已启用插件贡献的技能。
//
// 三条来源的合并、优先级与校验只有 internal/skill 一份实现；这里只负责取来源。
// `tt install skills`、`tt skills list/deploy/remove` 与插件状态变更后的重算都走它 ——
// 这是 P4-09 要的"两条既有安装路径共用同一个组装点"。
func buildSkillsPlan() (skill.Plan, string, error) {
	src, err := officialSkillsDir()
	if err != nil {
		return skill.Plan{}, "", err
	}
	externals, err := externalSkillPaths()
	if err != nil {
		return skill.Plan{}, "", err
	}
	pluginSkills, pre := pluginSkillSources()
	plan, err := skill.BuildPlanWithPluginSkills(src, externals, pluginSkills)
	if err != nil {
		return skill.Plan{}, "", err
	}
	// 插件侧"连来源都取不到"的问题（注册表坏了）排在技能级告警之后：
	// 逐条技能的告警才是用户要动手的，整体性问题的严重度更高但不需要逐项修。
	plan.Warnings = append(plan.Warnings, pre...)
	return plan, src, nil
}

// pluginStoreForSkills 造一个只读插件注册表面（技能管线只用到 Store）。
//
// 解析不出配置落点不是错误：便携包第一次运行、或压根不装插件时，官方技能照样要装。
func pluginStoreForSkills() (*plugin.Store, error) {
	cfgPath, err := common.ResolveConfig(true)
	if err != nil {
		return nil, err
	}
	return plugin.NewStore(config.LocationsAt(cfgPath)), nil
}

// pluginSkillSources 把已启用插件贡献的技能目录翻成计划来源。
//
// 只做投影：目录来自各插件的**活跃版本目录**（plugin.Store.ActiveDir），升级换版本后
// 自然指向新版本；合不合法、撞不撞名由 internal/skill 判（InspectPlugin/计划）。
func pluginSkillSources() ([]skill.Source, []skill.Warning) {
	store, err := pluginStoreForSkills()
	if err != nil {
		return nil, nil // 连落点都解析不出 = 没有插件可谈
	}
	srcs, serr := store.EnabledSkillSources()
	if serr != nil {
		// 注册表坏了不该让整份技能计划失败：官方与第三方技能照装，插件技能这次缺席，
		// 但要让用户看见原因（否则"插件技能怎么没装"无从查起）。
		return nil, []skill.Warning{{
			Kind: skill.StateInvalid, Path: store.RegistryPath(),
			Why: "插件注册表读不出来，本次不含插件技能：" + serr.Error(),
		}}
	}
	out := make([]skill.Source, 0, len(srcs))
	for _, s := range srcs {
		out = append(out, skill.Source{Origin: skill.OriginPlugin, Plugin: s.Plugin, Dir: s.Dir})
	}
	return out, nil
}

// disabledPluginSkillNames 「已停用/已装未启用」插件贡献的技能名（撤走判据）。
func disabledPluginSkillNames() (map[string]bool, error) {
	store, err := pluginStoreForSkills()
	if err != nil {
		return map[string]bool{}, nil
	}
	return store.DisabledSkillNames()
}

// skillsSyncRoots 可能装过技能的根：当前项目目录 + 用户目录（全局）。
//
// 为什么两个都算：插件的启停是全局事件，而技能可以装到项目级也可以装到全局。
// 只有**清单文件存在**的根会被重算：没装过的根一个字节都不碰（P4-07 的"不新建平台目录"）。
func skillsSyncRoots() []string {
	var out []string
	seen := make(map[string]bool, 2)
	add := func(p string) {
		if p == "" {
			return
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			return
		}
		key := abs
		if runtime.GOOS == "windows" {
			key = strings.ToLower(abs)
		}
		if seen[key] {
			return
		}
		seen[key] = true
		out = append(out, abs)
	}
	if wd, err := os.Getwd(); err == nil {
		add(wd)
	}
	if home, err := homeDir(); err == nil {
		add(home)
	}
	return out
}

// syncPluginSkills 在插件状态变更后重算期望态：把已部署的技能刷成当前计划的源。
//
// 只碰"已经部署过技能"的平台（平台清单里有记录、且目录还在）：插件启停不该在用户的
// 工程里顺手长出 .claude/ 之类的新目录。
//
// 返回的 error 是**部分失败的报告**：某个平台写不进去不影响其它平台，全部试过之后
// 一次说出来（reason 带上是哪次动作触发的，排障时能对上）。
func syncPluginSkills(reason string) error {
	plan, _, err := buildSkillsPlan()
	if err != nil {
		if errors.Is(err, errNoOfficialSkills) {
			// 这台机器没有官方技能树 = 没有技能管线；插件启停不该因此失败。
			return nil
		}
		return err
	}
	removable, rerr := disabledPluginSkillNames()
	if rerr != nil {
		return fmt.Errorf("读插件注册表失败，无法判断该撤走哪些技能：%w", rerr)
	}

	var errs []error
	for _, base := range skillsSyncRoots() {
		if serr := syncSkillsAt(base, plan, removable); serr != nil {
			errs = append(errs, fmt.Errorf("%s: %w", base, serr))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("技能同步失败（%s）：%w", reason, errors.Join(errs...))
	}
	return nil
}

// syncSkillsAt 重算一个根目录下的已部署平台。
//
// 每个平台各自收口：一个平台写不进去（目录被删、被换成同名文件、权限不够）只把它
// 记进错误清单，其余平台照常刷新 —— 插件状态已经变了，用户不该因为一个平台坏掉
// 就以为"整个技能分发都失败了"。
func syncSkillsAt(base string, plan skill.Plan, removable map[string]bool) error {
	m, exists, err := skill.LoadManifest(base)
	if err != nil {
		return err
	}
	if !exists {
		return nil // 这个范围从没部署过：不建目录、不写清单
	}

	available := make(map[string]bool, len(plan.Skills))
	var pluginNames []string
	for _, s := range plan.Skills {
		available[s.Name] = true
		if s.Origin == skill.OriginPlugin {
			pluginNames = append(pluginNames, s.Name)
		}
	}
	sort.Strings(pluginNames)

	var errs []error
	for _, t := range skill.Targets {
		recorded := m.Skills(t.ID)
		if len(recorded) == 0 {
			continue // 这个平台没装过技能：绝不新建它的目录
		}
		dir := t.AbsDir(base)
		st, serr := os.Stat(dir)
		switch {
		case os.IsNotExist(serr):
			errs = append(errs, fmt.Errorf("%s：技能目录已不在 %s，跳过（不新建平台目录）", t.ID, dir))
			continue
		case serr != nil:
			errs = append(errs, fmt.Errorf("%s：读不到 %s: %w", t.ID, dir, serr))
			continue
		case !st.IsDir():
			errs = append(errs, fmt.Errorf("%s：%s 不是目录，跳过这个平台", t.ID, dir))
			continue
		}

		// 期望态 = 清单里仍然可用的（尊重用户当初的勾选）+ 这次新可用的插件技能。
		// 后者的理由是"启用插件"本身就意味着"让它的贡献生效"（CLI/页面/服务同理）。
		keep := make([]string, 0, len(recorded)+len(pluginNames))
		for _, n := range recorded {
			if available[n] {
				keep = append(keep, n)
			}
		}
		keep = append(keep, pluginNames...)

		// 先撤走"源确实没了"的条目（只认清单 + 只认可撤名单），再把期望态写一遍。
		if _, werr := skill.Withdraw(skill.WithdrawRequest{
			Project:   base,
			Targets:   []skill.DeployTarget{{ID: t.ID, Skills: keep}},
			Removable: removable,
		}); werr != nil {
			errs = append(errs, fmt.Errorf("%s：撤走旧技能失败: %w", t.ID, werr))
			continue
		}
		if _, derr := skill.Deploy(plan, skill.DeployRequest{
			Project: base,
			Targets: []skill.DeployTarget{{ID: t.ID, Skills: keep}},
		}); derr != nil {
			errs = append(errs, fmt.Errorf("%s：写入失败: %w", t.ID, derr))
			continue
		}
	}
	return errors.Join(errs...)
}
