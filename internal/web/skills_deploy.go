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
	plan, err := skill.BuildPlan(officialDir, root.Skills.Paths())
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
