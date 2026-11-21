package devdata

// 清理：**只走 `tt gen clean`**，不进 `tt cache`（plan ADR-02）。
//
// 默认 dry-run（把要删的算出来给人看），`--yes` / `?yes=1` 才真删。理由：devdata 里可能
// 是某环境唯一一份已归档的蓝图，而生产环境甚至可能已经重新产生过 —— 删了要重新拉，
// 且拉取成本是几十 MB 加数据库查询。这与"清缓存"是两种动作，不该共用一个确认。

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/luyaogod/TT/internal/config"
)

// CleanItem 一项待删内容。
type CleanItem struct {
	// Kind env | prog
	Kind  string `json:"kind"`
	Env   string `json:"env"`
	Prog  string `json:"prog,omitempty"`
	Path  string `json:"path"`
	Files int    `json:"files"`
	Bytes int64  `json:"bytes"`
}

// CleanPlan 一次清理计划（dry-run 与执行共用同一份）。
type CleanPlan struct {
	Items []CleanItem `json:"items"`
	Files int         `json:"files"`
	Bytes int64       `json:"bytes"`
}

// Empty 没有任何可删的内容。
func (p *CleanPlan) Empty() bool { return p == nil || len(p.Items) == 0 }

// PlanClean 算一次清理计划（**不删任何东西**）。
//
// 三种粒度：给了 prog 删该程序目录；只给 env 删整个环境目录；都没给 = 报错
// （"删所有环境"这种一步到位的动作必须点名，见 tt-pj-add-cli 的"环境一律点名"）。
func PlanClean(cfgPath, env, prog string) (*CleanPlan, error) {
	e, err := envInfo(cfgPath, env)
	if err != nil {
		return nil, err
	}
	envDir, err := EnvDir(cfgPath, e.Name)
	if err != nil {
		return nil, err
	}
	plan := &CleanPlan{Items: []CleanItem{}}
	if strings.TrimSpace(prog) != "" {
		if err := ValidProgName(prog); err != nil {
			return nil, err
		}
		dir := filepath.Join(envDir, config.DevDataSubdirProg, strings.TrimSpace(prog))
		if _, serr := os.Stat(dir); serr != nil {
			return nil, fmt.Errorf("%w（本地没有该程序的材料）: %s", ErrProgNotFound, prog)
		}
		files, bytes := dirUsage(dir)
		plan.Items = append(plan.Items, CleanItem{Kind: "prog", Env: e.Name, Prog: strings.TrimSpace(prog), Path: dir, Files: files, Bytes: bytes})
	} else {
		if _, serr := os.Stat(envDir); serr != nil {
			return nil, fmt.Errorf("%w: 环境 %q 还没有任何开发资料", ErrProgNotFound, e.Name)
		}
		files, bytes := dirUsage(envDir)
		plan.Items = append(plan.Items, CleanItem{Kind: "env", Env: e.Name, Path: envDir, Files: files, Bytes: bytes})
	}
	for _, it := range plan.Items {
		plan.Files += it.Files
		plan.Bytes += it.Bytes
	}
	return plan, nil
}

// ExecuteClean 真删。**顺序按路径长度倒序**：先删深的（prog 目录），再删浅的 ——
// 先删父目录会让后面的条目"不存在"，而那种失败很容易被当成"删了一半"。
func ExecuteClean(plan *CleanPlan) error {
	if plan.Empty() {
		return nil
	}
	items := append([]CleanItem(nil), plan.Items...)
	sort.Slice(items, func(i, j int) bool { return len(items[i].Path) > len(items[j].Path) })
	for _, it := range items {
		if err := os.RemoveAll(it.Path); err != nil {
			return fmt.Errorf("删除失败 (%s): %w", it.Path, err)
		}
	}
	return nil
}

// RemoveProg 删一个程序的材料目录（Web 的 DELETE /api/devdata/prog 用）。
func RemoveProg(cfgPath, env, prog string) (*CleanPlan, error) {
	plan, err := PlanClean(cfgPath, env, prog)
	if err != nil {
		return nil, err
	}
	if err := ExecuteClean(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// dirUsage 递归统计一个目录的文件数与字节数（不存在返回 0,0）。
func dirUsage(dir string) (int, int64) {
	var files int
	var bytes int64
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}
		files++
		if info, ierr := d.Info(); ierr == nil {
			bytes += info.Size()
		}
		return nil
	})
	return files, bytes
}
