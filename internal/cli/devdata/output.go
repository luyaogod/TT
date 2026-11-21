package devdata

// 输出：JSON 默认（agent 用），--human 出小表（人用）。
//
// 两种形态**内容一致**：表格只换展示形态，不偷偷少字段（tt-pj-add-cli 的"可信"原则）。
// 因此每张表的列都从同一份结果对象里取，没有第二套计算。

import (
	"fmt"
	"os"
	"strings"

	"github.com/luyaogod/TT/internal/cli/common"
	"github.com/luyaogod/TT/internal/devdata"
	"github.com/luyaogod/TT/internal/output"
)

// meta 本组共用的信封头（环境/来源）。企业只有在真的用了才写。
func meta(env string, ent int) output.Meta {
	m := common.CurrentMeta()
	m.Env = env
	m.Source = "local"
	if ent > 0 {
		m.Ent = ent
	}
	return m
}

// emit 统一出口。
func emit(data any, env string, ent int, columns []string, rows [][]string) error {
	return output.Emit(os.Stdout, output.Options{
		Format:  common.OutputFormat(),
		Meta:    meta(env, ent),
		Data:    data,
		Columns: columns,
		Rows:    rows,
	})
}

// emitStatus 输出环境状态（共用资料 + 程序清单 + 在跑任务）。
func emitStatus(st *devdata.EnvStatus) error {
	rows := [][]string{
		{"共用资料", "规格表", fmtCount(st.Dict.Tables, st.Dict.Skipped), fmtBytes(st.Dict.Size), fmtAge(st.Dict.SyncedAt), presentFlag(st.Dict.Present)},
		{"共用资料", "模板", fmt.Sprintf("%d / %d 个文件", st.Mdl.TemplateCount, st.Mdl.Files), fmtBytes(st.Mdl.Bytes), fmtAge(st.Mdl.SyncedAt), presentFlag(st.Mdl.Present)},
	}
	for _, p := range st.Progs {
		status := "最新"
		if p.Stale {
			status = "已过期"
			if len(p.StaleReasons) > 0 {
				status += "：" + p.StaleReasons[0].Detail
			}
		}
		rows = append(rows, []string{"按需资料", p.Prog, fmt.Sprintf("%d 个文件", p.Files), fmtBytes(p.Bytes), fmtAge(p.FetchedAt), status})
	}
	for _, j := range st.Jobs {
		rows = append(rows, []string{"任务", j.Describe(), fmt.Sprintf("%d/%d", j.Done, j.Total), "", j.Elapsed, j.Message})
	}
	// JSON 载荷与表格同源：直接给领域层的结构，避免"界面看得到、JSON 没有"
	data := map[string]any{
		"env":      st.Env,
		"dir":      st.Dir,
		"present":  st.Present,
		"dict":     st.Dict,
		"mdl":      st.Mdl,
		"progs":    st.Progs,
		"jobs":     st.Jobs,
		"warnings": st.Warnings,
	}
	return emit(data, st.Env, 0,
		[]string{"分组", "名称", "规模", "大小", "时间", "状态"}, rows)
}

// emitProgs 只输出程序清单（tt gen ls）。
func emitProgs(st *devdata.EnvStatus) error {
	rows := make([][]string, 0, len(st.Progs))
	for _, p := range st.Progs {
		status := "最新"
		if p.Stale {
			status = "已过期"
		}
		missing := ""
		if len(p.Missing) > 0 {
			missing = strings.Join(p.Missing, ",")
		}
		rows = append(rows, []string{p.Prog, p.Module, fmt.Sprintf("%d", p.Files), fmtBytes(p.Bytes),
			fmtAge(p.FetchedAt), status, missing})
	}
	return emit(map[string]any{"env": st.Env, "dir": st.Dir, "progs": st.Progs}, st.Env, 0,
		[]string{"程序", "模块", "文件数", "大小", "拉取时间", "状态", "缺失"}, rows)
}

// emitCheck 输出一个程序的新鲜度明细（tt gen check）。
func emitCheck(p *devdata.ProgStatus) error {
	rows := make([][]string, 0, len(p.StaleReasons)+1)
	if len(p.StaleReasons) == 0 {
		rows = append(rows, []string{p.Prog, "-", "材料是最新的"})
	}
	for _, r := range p.StaleReasons {
		rows = append(rows, []string{p.Prog, r.Code, r.Detail})
	}
	return emit(map[string]any{"prog": p}, p.Prog, 0,
		[]string{"程序", "原因码", "说明"}, rows)
}

// emitMirror 输出共用资料同步结果。
func emitMirror(res *devdata.MirrorResult) error {
	rows := [][]string{
		{"规格表", fmt.Sprintf("%d 张 / %d 行", res.TableCount, res.RowTotal), ""},
		{"模板", fmt.Sprintf("%d 个文件", res.MdlFiles), fmtBytes(res.MdlBytes)},
		{"耗时", res.Elapsed, ""},
	}
	if res.SkippedPull {
		rows = append([][]string{{"结果", "已同步过，跳过（--force 重拉）", ""}}, rows...)
	}
	for _, w := range res.Warnings {
		rows = append(rows, []string{"提示", w, ""})
	}
	for _, m := range res.Missing {
		rows = append(rows, []string{"缺少", m, ""})
	}
	return emit(res, res.Env, 0, []string{"项目", "值", "大小"}, rows)
}

// emitFetch 输出按程序拉取的结果。
func emitFetch(res *devdata.FetchResult) error {
	rows := [][]string{{"程序", res.Prog, res.Module}, {"大小", fmtBytes(res.Bytes), fmt.Sprintf("%d 个文件", len(res.Files))}}
	for _, f := range res.Files {
		rows = append(rows, []string{"文件", f.Rel, fmt.Sprintf("%d B", f.Size)})
	}
	for _, m := range res.Missing {
		rows = append(rows, []string{"缺少", m, ""})
	}
	for _, w := range res.Warnings {
		rows = append(rows, []string{"提示", w, ""})
	}
	return emit(res, res.Env, 0, []string{"项目", "值", "备注"}, rows)
}

// emitCleanPlan 输出清理计划（dryRun=true 表示"只看不删"）。
func emitCleanPlan(plan *devdata.CleanPlan, dryRun bool) error {
	rows := make([][]string, 0, len(plan.Items))
	for _, it := range plan.Items {
		target := it.Env
		if it.Prog != "" {
			target += "/" + it.Prog
		}
		rows = append(rows, []string{it.Kind, target, it.Path, fmt.Sprintf("%d", it.Files), fmtBytes(it.Bytes)})
	}
	data := map[string]any{"dryRun": dryRun, "items": plan.Items, "files": plan.Files, "bytes": plan.Bytes}
	return emit(data, "", 0, []string{"类型", "对象", "路径", "文件数", "大小"}, rows)
}

// ---------- 小工具 ----------

func fmtBytes(n int64) string {
	if n <= 0 {
		return "0 B"
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}

func fmtCount(tables, skipped int) string {
	if skipped > 0 {
		return fmt.Sprintf("%d 张（跳过 %d）", tables, skipped)
	}
	return fmt.Sprintf("%d 张", tables)
}

func fmtAge(stamp string) string {
	if stamp == "" {
		return "—"
	}
	return devdata.AgeText(stamp)
}

func presentFlag(ok bool) string {
	if ok {
		return "已同步"
	}
	return "未同步"
}
