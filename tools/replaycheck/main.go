// replaycheck：离线校验「按身份回放」这个思路能不能成立。
//
// 它**不连界面、不发一个字节**，只做一件事：把一份录制里的每条事件，用**这份录制自己的**
// 界面模型（-rt2gdc 侧）把身份（colName/fieldId）解析出来，逐条告诉你：
//
//	✔ 解析到 → 录制时的 id 与本模型里的 id 是否一致、是什么类型/名字
//	✘ 拒绝   → 四码之一（FIELD_NOT_IN_SCOPE / FIELD_AMBIGUOUS / FORM_UNKNOWN / FIELD_NOT_ACTIONABLE）
//	－ 无身份 → KeyEvent 这类本来就没有目标对象的步骤
//
// 判据与真回放完全同源：`fep.PlanReplay` + `fep.ResolveStepTarget`（不另写一套）。
//
// 用法：
//
//	go run ./tools/replaycheck -dir <录制目录>          # 目录里要有 *-gdc2rt.log 与 *-rt2gdc.log
//	go run ./tools/replaycheck -root <records 根> -id fep-20261018-175926
//
// 读完之后想上真机看画面，用：
//
//	tt recording replay <id> --attached --interval 1500ms
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tt/internal/fep"
)

func main() {
	dir := flag.String("dir", "", "录制目录（含 *-gdc2rt.log 与 *-rt2gdc.log）")
	root := flag.String("root", "", "records 根目录（配合 -id 自动定位）")
	id := flag.String("id", "", "录制 ID（配合 -root）")
	maxShow := flag.Int("max", 40, "最多逐条打印多少步")
	scan := flag.Bool("scan", false, "批量扫描 -root 下所有录制，列出「有几条真正可回放的步骤」")
	flag.Parse()

	if *scan {
		scanAll(*root)
		return
	}

	target := strings.TrimSpace(*dir)
	if target == "" && *root != "" && *id != "" {
		found, err := findRecording(*root, *id)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		target = found
	}
	if target == "" {
		fmt.Fprintln(os.Stderr, "给 -dir，或给 -root 与 -id")
		os.Exit(2)
	}

	evFile, uiFile := pickSides(target)
	if evFile == "" || uiFile == "" {
		fmt.Fprintf(os.Stderr, "%s 里缺 -gdc2rt.log（要回放的事件侧）或 -rt2gdc.log（界面模型侧）\n", target)
		os.Exit(1)
	}

	events, err := fep.ReadRecording(evFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读事件侧失败: %v\n", err)
		os.Exit(1)
	}
	uiMsgs, err := fep.ReadRecording(uiFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读界面侧失败: %v\n", err)
		os.Exit(1)
	}

	m := fep.NewModel()
	for _, msg := range uiMsgs {
		m.ApplyMessage(msg)
	}

	fmt.Printf("录制目录：%s\n", target)
	fmt.Printf("  事件侧 %s：%d 条\n", filepath.Base(evFile), len(events))
	fmt.Printf("  界面侧 %s：%d 条 → 模型对象 %d 个\n", filepath.Base(uiFile), len(uiMsgs), m.Len())
	fmt.Println()

	plan := fep.PlanReplay(events, fep.ReplayOptions{Model: m})

	var resolved, noIdent, refused, notInPlan int
	var withIdentity, byRawID int
	var refuseCodes = map[string]int{}
	shown := 0

	fmt.Printf("计划：扫描 %d 条 → 进计划 %d 条、跳过 %d 条\n",
		plan.Scanned, len(plan.Events), len(plan.Skips))
	if len(plan.Skips) > 0 {
		fmt.Println("  跳过的原因：")
		reasons := map[string]int{}
		for _, s := range plan.Skips {
			reasons[s.Reason]++
		}
		for r, n := range reasons {
			fmt.Printf("    %-40s %d\n", r, n)
		}
	}
	fmt.Println()

	fmt.Printf("%-4s %-6s %-26s %-10s %s\n", "#", "步骤", "colName", "fieldId", "结论")
	for _, ev := range plan.Events {
		tgt, rerr := fep.ResolveStepTarget(m, ev, fep.Scope{})
		col := strings.TrimSpace(string(ev.ColName))
		if col == "" {
			col = "—"
		}
		fid := "—"
		if ev.FieldID != 0 {
			fid = fmt.Sprint(ev.FieldID)
		}
		verdict := ""
		switch {
		case rerr != nil:
			refused++
			code := "REFUSED"
			if r, ok := rerr.(*fep.Refusal); ok {
				code = string(r.Code)
			}
			refuseCodes[code]++
			verdict = "✘ 拒绝：" + code + "（" + firstLine(rerr.Error()) + "）"
		case tgt.ID == 0:
			noIdent++
			verdict = "－ 无目标对象（KeyEvent 之类，不需要身份）"
		default:
			resolved++
			desc := fmt.Sprintf("→ 解析到 #%d", tgt.ID)
			if o, ok := m.Get(tgt.ID); ok {
				desc += fmt.Sprintf(" (%s", o.Type)
				if n := o.Attr("name"); n != "" {
					desc += " " + n
				}
				desc += ")"
			}
			if tgt.FromIdentity {
				desc += " [按身份重解析]"
				withIdentity++
			} else {
				desc += " [沿用录制 id]"
				byRawID++
			}
			verdict = "✔ " + desc
		}
		if shown < *maxShow {
			fmt.Printf("%-4d %-6s %-26s %-10s %s\n", ev.N, kindOf(ev.Event), shorten(col, 26), fid, verdict)
			shown++
		}
	}
	if shown < len(plan.Events) {
		fmt.Printf("… 其余 %d 步未打印（-max 调大）\n", len(plan.Events)-shown)
	}
	notInPlan = plan.Scanned - len(plan.Events)

	fmt.Println()
	total := len(plan.Events)
	fmt.Printf("汇总：进计划 %d 步 —— 解析到 %d、无目标对象 %d、拒绝 %d；另跳过噪声 %d\n",
		total, resolved, noIdent, refused, notInPlan)
	if total > 0 {
		fmt.Printf("      可回放率（解析到+无目标对象）：%.1f%%；需要拒绝的：%.1f%%\n",
			100*float64(resolved+noIdent)/float64(total), 100*float64(refused)/float64(total))
		fmt.Printf("      **身份覆盖**：按身份重解析 %d 步、沿用录制裸 id %d 步（后者换会话必漂）\n",
			withIdentity, byRawID)
	}
	if len(refuseCodes) > 0 {
		keys := make([]string, 0, len(refuseCodes))
		for k := range refuseCodes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Println("      拒绝码分布：")
		for _, k := range keys {
			fmt.Printf("        %-24s %d\n", k, refuseCodes[k])
		}
	}
	fmt.Println()
	fmt.Println("下一步（想上真机看画面）：先起一条会话，然后")
	fmt.Println("  tt recording replay <录制ID> --attached --interval 1500ms")
}

func kindOf(ev string) string {
	f := strings.Fields(strings.TrimSpace(ev))
	for _, w := range f {
		switch w {
		case "Action", "Click", "KeyEvent", "Entry", "Focus":
			return w
		}
	}
	if len(f) > 0 {
		return strings.Trim(f[0], "{}")
	}
	return "?"
}

func shorten(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i > 0 {
		return s[:i]
	}
	return s
}

// pickSides 在目录里挑出事件侧（gdc2rt）与界面侧（rt2gdc）。
func pickSides(dir string) (ev, ui string) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "", ""
	}
	for _, e := range ents {
		n := e.Name()
		switch {
		case strings.Contains(n, "gdc2rt") && (strings.HasSuffix(n, ".log")):
			if ev == "" {
				ev = filepath.Join(dir, n)
			}
		case strings.Contains(n, "rt2gdc") && (strings.HasSuffix(n, ".log")):
			if ui == "" {
				ui = filepath.Join(dir, n)
			}
		}
	}
	return ev, ui
}

// scanAll 批量扫描：哪几份录制里真有"可回放的步骤"。
//
// 这是选样用的 —— 一份录制可能全是 Tab/焦点这类噪声（`AllowedForReplay` 会全丢掉），
// 拿它验任何东西都验不出来。
func scanAll(root string) {
	if root == "" {
		fmt.Fprintln(os.Stderr, "-scan 需要 -root <records 根>")
		os.Exit(2)
	}
	type row struct {
		id, prog, env           string
		events, inPlan, refused int
	}
	var rows []row
	seen := map[string]bool{}
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// 两种布局都认：
		//   · 长期库：<root>/<env>/<ent>/<id>/<id>-1-gdc2rt.log（目录形式）
		//   · 缓存：  <root>/fep-<tag>-1-gdc2rt.log（**扁平**，同一目录里很多份）
		n := info.Name()
		if !strings.HasSuffix(n, "gdc2rt.log") {
			return nil
		}
		evFile := p
		uiFile := strings.TrimSuffix(p, "gdc2rt.log") + "rt2gdc.log"
		if _, serr := os.Stat(uiFile); serr != nil {
			return nil // 没有界面侧，回放不了
		}
		id := strings.TrimSuffix(filepath.Base(evFile), "-1-gdc2rt.log")
		if id == filepath.Base(evFile) { // 名字不合约定，用目录名兜底
			id = filepath.Base(filepath.Dir(evFile))
		}
		if seen[id] {
			return nil
		}
		seen[id] = true

		events, e1 := fep.ReadRecording(evFile)
		uiMsgs, e2 := fep.ReadRecording(uiFile)
		if e1 != nil || e2 != nil {
			return nil
		}
		m := fep.NewModel()
		for _, msg := range uiMsgs {
			m.ApplyMessage(msg)
		}
		plan := fep.PlanReplay(events, fep.ReplayOptions{Model: m})
		refused := 0
		for _, ev := range plan.Events {
			if _, rerr := fep.ResolveStepTarget(m, ev, fep.Scope{}); rerr != nil {
				refused++
			}
		}
		r := row{id: id, events: len(events), inPlan: len(plan.Events), refused: refused}
		if meta, merr := fep.ReadMeta(filepath.Dir(evFile), id); merr == nil {
			r.prog, r.env = meta.Prog, meta.Env
		}
		rows = append(rows, r)
		return nil
	})
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].inPlan != rows[j].inPlan {
			return rows[i].inPlan > rows[j].inPlan
		}
		return rows[i].id > rows[j].id
	})
	fmt.Printf("%-28s %-14s %-8s %6s %7s %7s\n", "录制 ID", "作业", "环境", "事件", "进计划", "拒绝")
	for _, r := range rows {
		fmt.Printf("%-28s %-14s %-8s %6d %7d %7d\n", r.id, trunc(r.prog, 14), trunc(r.env, 8), r.events, r.inPlan, r.refused)
	}
	fmt.Printf("\n共 %d 份。进计划 >0 的才值得上真机看画面。\n", len(rows))
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// findRecording 在 records 根下按 id 找目录（<root>/<env>/<ent>/<id>）。
func findRecording(root, id string) (string, error) {
	var found string
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if filepath.Base(p) == id {
			found = p
			return filepath.SkipAll
		}
		return nil
	})
	if found == "" {
		return "", fmt.Errorf("%s 下找不到录制 %s", root, id)
	}
	return found, nil
}
