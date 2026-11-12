package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DeployRequest 是一次分发请求：项目 + 期望态（每个平台要哪些技能）。
//
// `Targets` 是**完整的期望态**：出现的平台按这里同步，没出现的平台一概不碰。
// 设置页一次发五个平台；CLI 只发它点名的那些。
type DeployRequest struct {
	Project string         `json:"project"`
	Targets []DeployTarget `json:"targets"`
	DryRun  bool           `json:"dryRun,omitempty"`
}

// DeployTarget 是某个平台期望装上的技能名。
type DeployTarget struct {
	ID     string   `json:"id"`
	Skills []string `json:"skills"`
}

// DeployAction 是一次写入或删除。Files 是目标目录内的相对路径（/ 分隔）。
type DeployAction struct {
	Target string   `json:"target"`
	Skill  string   `json:"skill"`
	Files  []string `json:"files,omitempty"`
}

// DeploySkip 是一条“请求了但没做”的记录。
type DeploySkip struct {
	Target string `json:"target"`
	Skill  string `json:"skill"`
	Why    string `json:"why"`
}

// DeployReport 是一次同步的结果。Warnings 是技能源本身的问题（第三方失联等），
// 与本次勾选无关，原样带出去让界面/命令行提示。
type DeployReport struct {
	Project  string         `json:"project"`
	DryRun   bool           `json:"dryRun,omitempty"`
	Written  []DeployAction `json:"written"`
	Removed  []DeployAction `json:"removed"`
	Skipped  []DeploySkip   `json:"skipped,omitempty"`
	Warnings []Warning      `json:"warnings,omitempty"`
}

// Deploy 按期望态同步多个平台：勾选的覆盖写入，取消勾选的删除。
//
// 覆盖不需要判断版本，也不需要 --force：清单里的条目就是 TT 装的，直接替换；
// 不在清单里的同名目录同样替换（旧命令装的、手拷的），同步后纳入清单。
// 删除只发生在“清单里有、这次没勾”的条目上 —— 清单外的东西永远不动。
func Deploy(plan Plan, req DeployRequest) (DeployReport, error) {
	rep := DeployReport{
		Written:  []DeployAction{},
		Removed:  []DeployAction{},
		Warnings: plan.Warnings,
	}
	project, err := filepath.Abs(req.Project)
	if err != nil {
		return rep, fmt.Errorf("解析项目目录失败 %s: %w", req.Project, err)
	}
	st, err := os.Stat(project)
	if err != nil {
		if os.IsNotExist(err) {
			return rep, fmt.Errorf("项目目录不存在: %s", project)
		}
		return rep, fmt.Errorf("读项目目录失败 %s: %w", project, err)
	}
	if !st.IsDir() {
		return rep, fmt.Errorf("项目路径不是目录: %s", project)
	}
	rep.Project = project
	rep.DryRun = req.DryRun

	if len(req.Targets) == 0 {
		return rep, fmt.Errorf("没有选择平台")
	}

	byName := make(map[string]Source, len(plan.Skills))
	for _, s := range plan.Skills {
		byName[s.Name] = s
	}

	m, _, err := LoadManifest(project)
	if err != nil {
		return rep, err
	}

	seen := make(map[string]bool, len(req.Targets))
	for _, dt := range req.Targets {
		t, ok := TargetByID(dt.ID)
		if !ok {
			return rep, fmt.Errorf("不认识的平台: %s（可选：%s）", dt.ID, strings.Join(TargetIDs(), " / "))
		}
		if seen[t.ID] {
			return rep, fmt.Errorf("平台重复: %s", t.ID)
		}
		seen[t.ID] = true

		names, err := requestedSkills(dt.Skills, byName, plan.Warnings)
		if err != nil {
			return rep, err
		}

		// 写入：只把这次选中的技能组一个计划，Force = 覆盖同名条目。
		if len(names) > 0 {
			srcs := make([]Source, 0, len(names))
			for _, n := range names {
				srcs = append(srcs, byName[n])
			}
			res, aerr := Apply(Plan{Skills: srcs}, t.AbsDir(project), Options{Force: true, DryRun: req.DryRun})
			if aerr != nil {
				return rep, aerr
			}
			bySkill := make(map[string][]string, len(names))
			for _, f := range res.Copied {
				name := strings.SplitN(f, "/", 2)[0]
				bySkill[name] = append(bySkill[name], f)
			}
			for _, n := range names {
				files, ok := bySkill[n]
				if !ok {
					continue
				}
				rep.Written = append(rep.Written, DeployAction{Target: t.ID, Skill: n, Files: files})
			}
		}

		// 删除：清单里有、这次没选的条目。
		//
		// 但“技能源当前不可用”的条目不删：第三方目录可能只是移动硬盘没插，
		// 此时把已装好的副本删掉是净损失。清单里继续留着它，等源回来再处理。
		selected := make(map[string]bool, len(names))
		for _, n := range names {
			selected[n] = true
		}
		keep := make([]string, 0)
		for _, old := range m.Skills(t.ID) {
			if selected[old] {
				continue
			}
			if _, known := byName[old]; !known {
				rep.Skipped = append(rep.Skipped, DeploySkip{
					Target: t.ID, Skill: old, Why: "技能源当前不可用，保留旧副本",
				})
				keep = append(keep, old)
				continue
			}
			if !ValidName(old) {
				rep.Skipped = append(rep.Skipped, DeploySkip{Target: t.ID, Skill: old, Why: "清单里的技能名不合规，未删除"})
				continue
			}
			full := filepath.Join(t.AbsDir(project), old)
			if filepath.Dir(full) != t.AbsDir(project) {
				rep.Skipped = append(rep.Skipped, DeploySkip{Target: t.ID, Skill: old, Why: "路径越界，未删除"})
				continue
			}
			if !req.DryRun {
				if rerr := os.RemoveAll(full); rerr != nil {
					return rep, fmt.Errorf("删除旧技能失败 %s: %w", full, rerr)
				}
			}
			rep.Removed = append(rep.Removed, DeployAction{Target: t.ID, Skill: old})
		}

		// 清单跟着走：完成一个平台记一个，后面的平台失败也不会把前面的写入变成“无主”。
		if m.Set(t.ID, append(append([]string(nil), names...), keep...)) && !req.DryRun {
			if serr := m.Save(project); serr != nil {
				return rep, serr
			}
		}
	}
	return rep, nil
}

// requestedSkills 校验并规整一次请求里的技能名。
//
// 未知名字直接报错：请求是界面/命令行生成的，出现未知名字说明两边口径不一致，
// 静默跳过会让用户以为装上了。不可用的第三方给一句 why，便于一眼定位。
func requestedSkills(names []string, byName map[string]Source, warnings []Warning) ([]string, error) {
	seen := make(map[string]bool, len(names))
	out := make([]string, 0, len(names))
	for _, raw := range names {
		n := strings.TrimSpace(raw)
		if n == "" {
			continue
		}
		if !ValidName(n) {
			return nil, fmt.Errorf("技能名不合规: %q", n)
		}
		if seen[n] {
			continue
		}
		seen[n] = true
		if _, ok := byName[n]; !ok {
			why := ""
			for _, w := range warnings {
				if w.Name == n {
					why = w.Why
					break
				}
			}
			if why != "" {
				return nil, fmt.Errorf("技能 %s 不可用：%s", n, why)
			}
			return nil, fmt.Errorf("技能 %s 不在可用技能里（官方技能目录 + 已登记的第三方）", n)
		}
		out = append(out, n)
	}
	sort.Strings(out)
	return out, nil
}
