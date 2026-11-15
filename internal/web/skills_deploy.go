package web

// 「分发」页的后端：把技能同步到五个平台的项目目录。
//
// 与 `tt install skills` / `tt skills` 共用 internal/skill：校验、覆盖、清单、assets
// 指针都不在这里写。这一层只做三件事：解析项目路径、读配置里的第三方登记、把结果
// 原样带出去。
//
// 与旧「技能」页的分工：那条线只登记第三方技能（GET /api/skills）；这条线负责
// “装到哪儿”，是同一个入口的第二个视图，不是第二条安装路径 —— 目标由 internal/skill
// 的平台表推导，前端不能传任意目录。

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"tt/internal/config"
	"tt/internal/skill"
)

// matrixSkill 是矩阵里的一行。不可用的第三方技能也占一行（显示原因），
// id 用 `external:<序号>`，避免和真实技能名混在一起。
type matrixSkill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Origin      string `json:"origin"`
	Dir         string `json:"dir,omitempty"`
	Available   bool   `json:"available"`
	Why         string `json:"why,omitempty"`
}

// matrixTarget 是矩阵的一列。Skills 来自清单（= 已分发）。
type matrixTarget struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Dir      string   `json:"dir"`
	Detected bool     `json:"detected"`
	Skills   []string `json:"skills"`
}

// pluginSkillSources 已启用插件贡献的技能目录 → 计划来源。
//
// 与命令行走同一份注册表投影（plugin.Store.EnabledSkillSources）：目录取各插件的活跃
// 版本目录，合不合法、撞不撞名由 internal/skill 判。没有插件管理面（裁剪版服务）时为空。
func (s *Server) pluginSkillSources() ([]skill.Source, []skill.Warning) {
	if s.opt.PluginMgr == nil {
		return nil, nil
	}
	store := s.opt.PluginMgr.Store
	srcs, err := store.EnabledSkillSources()
	if err != nil {
		// 注册表坏了不该让整份技能视图打不开：官方与第三方照常列，插件这一栏缺席，
		// 但把原因带出去（否则"插件技能怎么没了"无从查起）。
		return nil, []skill.Warning{{
			Kind: skill.StateInvalid, Path: store.RegistryPath(),
			Why: "插件注册表读不出来，本次不含插件技能：" + err.Error(),
		}}
	}
	out := make([]skill.Source, 0, len(srcs))
	for _, p := range srcs {
		out = append(out, skill.Source{Origin: skill.OriginPlugin, Plugin: p.Plugin, Dir: p.Dir})
	}
	return out, nil
}

// skillsPlan 组计划：官方 + 登记第三方 + 已启用插件。与命令行的 buildSkillsPlan 同构，
// 共用 internal/skill 的合并与校验 —— 页面与命令行不该有两套"哪些技能可用"的口径。
//
// 第二个返回值是"来源级"告警（注册表读不出来这类），即使计划失败也要带出去。
func (s *Server) skillsPlan(officialDir string, externals []string) (skill.Plan, []skill.Warning, error) {
	srcs, pre := s.pluginSkillSources()
	plan, err := skill.BuildPlanWithPluginSkills(officialDir, externals, srcs)
	if err != nil {
		return skill.Plan{}, pre, err
	}
	plan.Warnings = append(plan.Warnings, pre...)
	return plan, pre, nil
}

// pluginSkillRows 插件技能行（与第三方那一栏同形：不合规的也占一行并带 why）。
func (s *Server) pluginSkillRows() []skillRow {
	srcs, _ := s.pluginSkillSources()
	rows := make([]skillRow, 0, len(srcs))
	for _, src := range srcs {
		rows = append(rows, skillRow{Info: skill.InspectPlugin(src)})
	}
	return rows
}

// hSkillsOverview 返回某个范围（项目 / 用户目录）的技能源、平台检测与清单。
//
// 不写盘、不联网。项目默认取服务进程的 cwd；scope=global 时根目录换成用户目录。
func (s *Server) hSkillsOverview(w http.ResponseWriter, r *http.Request) {
	base, scope, err := resolveSkillsBase(r.URL.Query().Get("project"), r.URL.Query().Get("scope"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	cfgPath, err := s.configPath(true)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	root, err := config.Load(cfgPath)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}

	rows, officialDir, officialErr := skillMatrix(root)

	// 插件技能排在官方/第三方之后，与命令行的计划顺序一致（同名的进 warnings）。
	// 官方树读不出来时组不出计划 —— 此时 officialError 已经说明了原因，不硬凑。
	var warnings = []skill.Warning{}
	if officialDir != "" {
		if plan, pre, perr := s.skillsPlan(officialDir, root.Skills.Paths()); perr == nil {
			warnings = plan.Warnings
			for _, src := range plan.Skills {
				if src.Origin != skill.OriginPlugin {
					continue
				}
				rows = append(rows, matrixSkill{
					ID: src.Name, Name: src.Name,
					Origin: string(skill.OriginPlugin), Dir: src.Dir, Available: true,
				})
			}
		} else {
			warnings = pre
		}
	}

	m, exists, err := skill.LoadManifest(base)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	targets := make([]matrixTarget, 0, len(skill.Targets))
	for _, t := range skill.Targets {
		targets = append(targets, matrixTarget{
			ID:       t.ID,
			Name:     t.Name,
			Dir:      t.Dir,
			Detected: t.Detected(base),
			Skills:   m.Skills(t.ID),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"scope":         scope,
		"project":       base,
		"officialDir":   officialDir,
		"officialError": officialErr,
		"manifest":      exists,
		"skills":        rows,
		"targets":       targets,
		"warnings":      warnings,
	})
}

// hSkillsDeploy 按页面提交的期望态同步多个平台。
//
// 请求里的 targets 是完整的期望态；没出现的平台不动。dryRun 只计算不写盘。
// scope=global 时根目录是用户目录，清单在 <用户目录>/.agents/.tt-manifest.json。
func (s *Server) hSkillsDeploy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Scope   string               `json:"scope"`
		Project string               `json:"project"`
		Targets []skill.DeployTarget `json:"targets"`
		DryRun  bool                 `json:"dryRun,omitempty"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	base, scope, err := resolveSkillsBase(req.Project, req.Scope)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	cfgPath, err := s.configPath(true)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	root, err := config.Load(cfgPath)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	officialDir, err := officialSkillsDirFn()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	// 与命令行同一份计划（官方 + 登记第三方 + 已启用插件）：插件技能也走这条既有路径，
	// 不另写一套复制/校验。
	plan, _, err := s.skillsPlan(officialDir, root.Skills.Paths())
	if err != nil {
		// 官方技能树坏了 = 这个包坏了，不是用户配错了。
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	rep, err := skill.Deploy(plan, skill.DeployRequest{Project: base, Targets: req.Targets, DryRun: req.DryRun})
	if err != nil {
		// 走到这里的错误基本是请求本身（未知平台/技能、项目不存在、清单坏了）。
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"scope":    scope,
		"project":  rep.Project,
		"dryRun":   rep.DryRun,
		"written":  rep.Written,
		"removed":  rep.Removed,
		"skipped":  rep.Skipped,
		"warnings": rep.Warnings,
	})
}

// skillMatrix 把官方技能 + 第三方登记合成矩阵行。
//
// 官方树不合规时返回 officialError（页面显示一条横幅），第三方照常列出来；
// 两者都不影响页面打开。
func skillMatrix(root *config.Root) (rows []matrixSkill, officialDir, officialErr string) {
	rows = make([]matrixSkill, 0)
	dir, derr := officialSkillsDirFn()
	if derr != nil {
		officialErr = derr.Error()
	} else {
		officialDir = dir
		infos, _, ierr := skill.InspectOfficial(dir)
		if ierr != nil {
			officialErr = ierr.Error()
		} else {
			for _, info := range infos {
				rows = append(rows, matrixSkill{
					ID: info.Name, Name: info.Name, Description: info.Description,
					Origin: string(skill.OriginOfficial), Dir: info.Dir, Available: true,
				})
			}
		}
	}

	var officialNames []string
	for _, row := range rows {
		if row.Origin == string(skill.OriginOfficial) {
			officialNames = append(officialNames, row.Name)
		}
	}

	entries := root.Skills.Entries()
	paths := root.Skills.Paths()
	for i, info := range skill.InspectExternal(paths, officialNames) {
		if info.State == skill.StateOK {
			rows = append(rows, matrixSkill{
				ID: info.Name, Name: info.Name, Description: info.Description,
				Origin: string(skill.OriginExternal), Dir: info.Dir, Available: true,
			})
			continue
		}
		label := ""
		if i < len(entries) {
			label = strings.TrimSpace(entries[i].Label)
		}
		if label == "" {
			label = filepath.Base(filepath.Clean(info.Dir))
		}
		if label == "" || label == "." {
			label = fmt.Sprintf("第 %d 条", info.Index)
		}
		rows = append(rows, matrixSkill{
			ID: fmt.Sprintf("external:%d", info.Index), Name: label,
			Origin: string(skill.OriginExternal), Dir: info.Dir,
			Available: false, Why: info.Why,
		})
	}
	return rows, officialDir, officialErr
}

// resolveProject 把请求里的项目目录规范成绝对路径，并要求它真的存在。
func resolveProject(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("取不到当前目录: %w", err)
		}
		raw = wd
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", fmt.Errorf("解析项目目录失败 %s: %w", raw, err)
	}
	st, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("项目目录不存在: %s", abs)
		}
		return "", fmt.Errorf("读项目目录失败 %s: %w", abs, err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("项目路径不是目录: %s", abs)
	}
	return abs, nil
}

// userHomeDir 可替换（测试用）：--scope=global 的用例不该写开发者真实的用户目录。
var userHomeDir = os.UserHomeDir

// resolveSkillsBase 把 project / scope 解析成“分发根目录 + 范围名”。
//
// 全局就是根目录换成用户目录，平台相对目录不变：Claude 的全局落点是
// <用户目录>/.claude/skills，清单在 <用户目录>/.agents/.tt-manifest.json。
func resolveSkillsBase(project, scope string) (base, kind string, err error) {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "", "project":
		base, err = resolveProject(project)
		return base, "project", err
	case "global":
		if strings.TrimSpace(project) != "" {
			return "", "", fmt.Errorf("scope=global 时不能再给 project")
		}
		home, herr := userHomeDir()
		if herr != nil {
			return "", "", fmt.Errorf("取不到用户目录: %w", herr)
		}
		st, serr := os.Stat(home)
		if serr != nil || !st.IsDir() {
			return "", "", fmt.Errorf("用户目录不存在: %s", home)
		}
		return home, "global", nil
	default:
		return "", "", fmt.Errorf("不认识的 scope: %s（project / global）", scope)
	}
}
