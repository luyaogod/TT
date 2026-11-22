package tzc

// 区域级的五个动词：points / show / find / patch / set。
//
// 背景：`prog.full.4gl` 有 77,140 行（实测 axmt500_wf），而一次改动通常只碰一个点里的
// 几行。整份文件当编辑面时，AI 要么读不下、要么在围栏与行号上出错。
// 这五个动词把"区域"（围栏围起来的点 / 区段）变成一等对象：
//
//	points            列区域（点名 / 行数 / 能不能改 / 为什么不能）
//	show    <name>    只看这一个区域（带行号；行号即 prog.full.4gl 的行号）
//	find    <name> p  在区域里按子串定位（带上下文）
//	patch   <name>    改一处：--line N --text S（按行，保留缩进）或 --old/--new（子串，须唯一）
//	set     <name>    整段替换（只对小区域、且不含内嵌加点）
//
// 共同约束：**只改 workspace，不写包**。落包仍走 `verify` / `apply`（三道闸门不变）。
//
// 两条血泪规则（来自试用）：
//  1. patch 必须能预览（--dry-run）——纯子串替换会静默拼出 `##` 这种坏行，退出码还是 0；
//  2. 围栏行绝不可被 old 或 new 触碰——删掉一行围栏，verify 要等到下一个孤立 end 才报错，行号还指错。
//
// 退出码与会话文档一致：2 用法 / 3 验证失败 / 4 写入被拒 / 5 IO。

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/luyaogod/TT/internal/cli/dev/common"
	"github.com/luyaogod/TT/internal/dev/fence"
	"github.com/luyaogod/TT/internal/dev/model"
	"github.com/luyaogod/TT/internal/dev/store"
)

const (
	exitUsage  = 2
	exitDenied = 4
	exitIO     = 5
)

// failEdit 按文档化退出码报错（不再一律 1）。
func failEdit(code int, err error, asJSON bool) int {
	if asJSON {
		common.EmitJSON(os.Stdout, map[string]any{"ok": false, "exitCode": code, "error": err.Error()})
		return code
	}
	fmt.Fprintf(os.Stderr, "错误（退出码 %d）：%v\n", code, err)
	return code
}

// editCtx 是一次区域编辑的上下文。
type editCtx struct {
	ws     *store.Workspace
	base   *model.Document
	edited []byte
	parsed *fence.ParseResult
}

// openForEdit 打开工作区并解析出"当前编辑文本 + 围栏"。
func openForEdit(dir string) (*editCtx, error) {
	ws, err := store.Open(dir)
	if err != nil {
		return nil, err
	}
	base, _, err := loadBase(ws)
	if err != nil {
		return nil, err
	}
	edited, err := ws.ReadEdited()
	if err != nil {
		return nil, err
	}
	parsed, err := fence.Parse(base, edited)
	if err != nil {
		return nil, err
	}
	return &editCtx{ws: ws, base: base, edited: edited, parsed: parsed}, nil
}

// resolveDirAndName 解析「[<dir>] <name>」：第一个位置参数是工作区目录时才算 dir。
func resolveDirAndName(args []string) (dir, name string, err error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("缺 <点名>（先 `tt dev tzc points` 看有哪些）")
	}
	if looksLikeWorkspaceArg(args[0]) {
		if len(args) < 2 {
			return "", "", fmt.Errorf("缺 <点名>（先 `tt dev tzc points` 看有哪些）")
		}
		return args[0], args[1], nil
	}
	d, derr := resolveWorkspaceDir("")
	if derr != nil {
		return "", "", derr
	}
	return d, args[0], nil
}

// findRegion 按点名找区域（含子区域）。
func findRegion(d *model.Document, name string) *model.Region {
	var walk func(rs []*model.Region) *model.Region
	walk = func(rs []*model.Region) *model.Region {
		for _, r := range rs {
			if r.Name == name {
				return r
			}
			if got := walk(r.Children); got != nil {
				return got
			}
		}
		return nil
	}
	return walk(d.Regions)
}

// bodyRange 取区域正文的字节区间。
//
// 点有 BodySpan；**section 没有**（渲染层只给点记 span，区段的正文在两行围栏之间）。
// 所以区段要在文本里按围栏名定位。注意：结束围栏是 `{//@tdev:end section}`，**不带名字**。
func bodyRange(text []byte, r *model.Region) (int, int, bool) {
	if r == nil {
		return 0, 0, false
	}
	if r.BodySpan != nil {
		s, e := r.BodySpan.Start, r.BodySpan.End
		if s >= 0 && e >= s && e <= len(text) {
			return s, e, true
		}
		return 0, 0, false
	}
	if r.Kind != model.RegionSection {
		return 0, 0, false
	}
	begin := []byte("{//@tdev:begin section " + r.Name)
	bi := bytes.Index(text, begin)
	if bi < 0 {
		return 0, 0, false
	}
	nl := bytes.IndexByte(text[bi:], '\n')
	if nl < 0 {
		return 0, 0, false
	}
	start := bi + nl + 1
	ei := bytes.Index(text[start:], []byte("{//@tdev:end section}"))
	if ei < 0 {
		return 0, 0, false
	}
	return start, start + ei, true
}

// bodyOf 取区域的正文切片（文档坐标）。
func bodyOf(text []byte, r *model.Region) ([]byte, bool) {
	s, e, ok := bodyRange(text, r)
	if !ok {
		return nil, false
	}
	return text[s:e], true
}

// innerPointAt 返回该偏移处最内层的"点"区域（区段正文里的改动会落到某个内嵌点上）。
// 写回时 split 按点处理，所以回执必须告诉用户"实际落在哪个点"。
func innerPointAt(text []byte, d *model.Document, off int) *model.Region {
	var best *model.Region
	bestLen := 1 << 30
	var walk func(rs []*model.Region)
	walk = func(rs []*model.Region) {
		for _, r := range rs {
			if r.Kind == model.RegionPoint && r.BodySpan != nil {
				s, e := r.BodySpan.Start, r.BodySpan.End
				if off >= s && off <= e && e-s < bestLen {
					best, bestLen = r, e-s
				}
			}
			walk(r.Children)
		}
	}
	walk(d.Regions)
	return best
}

// lineOfOffset 把文档偏移换成 1 起行号。
func lineOfOffset(text []byte, off int) int {
	if off < 0 {
		return 0
	}
	if off > len(text) {
		off = len(text)
	}
	return bytes.Count(text[:off], []byte("\n")) + 1
}

// regionLines 数区域正文的行数（不含末尾空行）。
func regionLines(body []byte) int {
	s := strings.TrimRight(string(body), "\r\n")
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// lineSpan 取正文里第 idx 行（0 起）的 [内容起, 内容止) 与行尾结束位置。
func lineSpan(body []byte, idx int) (start, contentEnd, lineEnd int, ok bool) {
	if idx < 0 {
		return 0, 0, 0, false
	}
	pos, cur := 0, 0
	for {
		nl := bytes.IndexByte(body[pos:], '\n')
		var end int
		if nl < 0 {
			end = len(body)
		} else {
			end = pos + nl + 1
		}
		if cur == idx {
			ce := end
			if nl >= 0 {
				ce = pos + nl
			}
			if ce > pos && body[ce-1] == '\r' {
				ce--
			}
			return pos, ce, end, true
		}
		if nl < 0 {
			return 0, 0, 0, false
		}
		pos, cur = end, cur+1
	}
}

// regionMeta 是 points/show 输出用的区域视图。
type regionMeta struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Editable bool   `json:"editable"`
	DenyCode string `json:"denyCode,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Lines    int    `json:"lines"`
	Bytes    int    `json:"bytes"`
	Line0    int    `json:"line0"`
	Line1    int    `json:"line1"`
}

func describeRegion(text []byte, r *model.Region) regionMeta {
	m := regionMeta{Name: r.Name, Kind: string(r.Kind), Editable: r.Editable,
		DenyCode: r.DenyCode, Reason: r.Reason}
	if s, e, ok := bodyRange(text, r); ok {
		m.Lines = regionLines(text[s:e])
		m.Bytes = e - s
		m.Line0 = lineOfOffset(text, s)
		m.Line1 = lineOfOffset(text, e)
	}
	return m
}

// ---------- points ----------

func cmdPoints(args []string) int {
	fs := flag.NewFlagSet("points", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "输出 JSON")
	kind := fs.String("kind", "", "只列某类：point / section")
	match := fs.String("match", "", "只列点名包含该子串的区域（大小写不敏感）")
	editable := fs.Bool("editable", false, "只列可编辑的区域")
	withBody := fs.Bool("nonempty", false, "只列正文非空的区域")
	if err := common.ParseArgs(fs, args, "kind", "match"); err != nil {
		return exitUsage
	}
	dir, derr := resolveWorkspaceDir(fs.Arg(0))
	if derr != nil {
		fmt.Fprintln(os.Stderr, "用法：tt dev tzc points [<dir>] [--kind point|section] [--match <子串>] [--editable] [--nonempty] [--json]")
		return failEdit(exitUsage, derr, *asJSON)
	}
	ec, err := openForEdit(dir)
	if err != nil {
		return failEdit(exitIO, err, *asJSON)
	}
	needle := strings.ToLower(strings.TrimSpace(*match))
	var out []regionMeta
	var walk func(rs []*model.Region)
	walk = func(rs []*model.Region) {
		for _, r := range rs {
			m := describeRegion(ec.edited, r)
			ok := true
			if *kind != "" && m.Kind != *kind {
				ok = false
			}
			if needle != "" && !strings.Contains(strings.ToLower(m.Name), needle) {
				ok = false
			}
			if *editable && !m.Editable {
				ok = false
			}
			if *withBody && m.Lines == 0 {
				ok = false
			}
			if ok {
				out = append(out, m)
			}
			walk(r.Children)
		}
	}
	walk(ec.parsed.Doc.Regions)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Line0 < out[j].Line0 })

	if *asJSON {
		common.EmitJSON(os.Stdout, map[string]any{"ok": true, "prog": ec.base.Prog, "count": len(out), "regions": out})
		return 0
	}
	w := os.Stdout
	common.Line(w, "工作区：%s（程序 %s，共 %d 个区域）", dir, ec.base.Prog, len(out))
	common.Line(w, "%-52s %-8s %-6s %6s %8s  %s", "点名", "类型", "可编辑", "行数", "字节", "说明")
	for _, m := range out {
		note := ""
		if !m.Editable {
			note = "不可改：" + m.DenyCode
			if m.Reason != "" {
				note += "（" + m.Reason + "）"
			}
		}
		common.Line(w, "%-52s %-8s %-6s %6d %8d  %s", m.Name, m.Kind, yesNo(m.Editable), m.Lines, m.Bytes, note)
	}
	return 0
}

func yesNo(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

// ---------- show ----------

func cmdShowRegion(args []string) int {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "输出 JSON")
	maxLines := fs.Int("max", 0, "最多打印多少行（0 = 全部）")
	if err := common.ParseArgs(fs, args, "max"); err != nil {
		return exitUsage
	}
	dir, name, derr := resolveDirAndName(fs.Args())
	if derr != nil {
		fmt.Fprintln(os.Stderr, "用法：tt dev tzc show [<dir>] <点名> [--max N] [--json]")
		return failEdit(exitUsage, derr, *asJSON)
	}
	ec, err := openForEdit(dir)
	if err != nil {
		return failEdit(exitIO, err, *asJSON)
	}
	r := findRegion(ec.parsed.Doc, name)
	if r == nil {
		return failEdit(exitUsage, fmt.Errorf("工作区里没有点名 %q（先 tt dev tzc points --match %s）", name, name), *asJSON)
	}
	body, ok := bodyOf(ec.edited, r)
	if !ok {
		return failEdit(exitUsage, fmt.Errorf("%q 没有正文区间（%s）", name, r.Kind), *asJSON)
	}
	meta := describeRegion(ec.edited, r)
	text := strings.TrimRight(string(body), "\r\n")
	lines := []string{}
	if text != "" {
		lines = strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	}
	truncated := false
	if *maxLines > 0 && len(lines) > *maxLines {
		lines = lines[:*maxLines]
		truncated = true
	}

	if *asJSON {
		common.EmitJSON(os.Stdout, map[string]any{
			"ok": true, "region": meta, "content": lines, "truncated": truncated,
		})
		return 0
	}
	w := os.Stdout
	common.Line(w, "区域：%s（%s，%d 行，%d 字节，%s）", meta.Name, meta.Kind, meta.Lines, meta.Bytes, editableText(meta))
	// 用 "| " 分隔行号与正文：直接拼接会看不清正文的首行缩进（做 --old 锚点时这点很要命）
	for i, ln := range lines {
		common.Line(w, "L%-6d| %s", meta.Line0+i, ln)
	}
	if truncated {
		common.Line(w, "…（超过 --max %d，已截断；剩余 %d 行）", *maxLines, meta.Lines-*maxLines)
	}
	return 0
}

func editableText(m regionMeta) string {
	if m.Editable {
		return "可编辑"
	}
	if m.Reason != "" {
		return "不可改：" + m.DenyCode + "（" + m.Reason + "）"
	}
	return "不可改：" + m.DenyCode
}

// ---------- find ----------

func cmdFindInRegion(args []string) int {
	fs := flag.NewFlagSet("find", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "输出 JSON")
	context := fs.Int("context", 2, "每条命中附几行上下文（文本输出也打印）")
	caseSensitive := fs.Bool("case", false, "区分大小写")
	limit := fs.Int("limit", 30, "最多几条命中")
	if err := common.ParseArgs(fs, args, "context", "limit"); err != nil {
		return exitUsage
	}
	rest := fs.Args()
	if len(rest) < 2 {
		fmt.Fprintln(os.Stderr, "用法：tt dev tzc find [<dir>] <点名> <子串> [--context N] [--case] [--limit N] [--json]")
		return exitUsage
	}
	dir, name := "", ""
	if looksLikeWorkspaceArg(rest[0]) {
		if len(rest) < 3 {
			fmt.Fprintln(os.Stderr, "用法：tt dev tzc find [<dir>] <点名> <子串> …")
			return exitUsage
		}
		dir, name = rest[0], rest[1]
		rest = rest[2:]
	} else {
		d, derr := resolveWorkspaceDir("")
		if derr != nil {
			return failEdit(exitUsage, derr, *asJSON)
		}
		dir, name = d, rest[0]
		rest = rest[1:]
	}
	pattern := rest[0]

	ec, err := openForEdit(dir)
	if err != nil {
		return failEdit(exitIO, err, *asJSON)
	}
	r := findRegion(ec.parsed.Doc, name)
	if r == nil {
		return failEdit(exitUsage, fmt.Errorf("工作区里没有点名 %q（先 tt dev tzc points --match %s）", name, name), *asJSON)
	}
	body, ok := bodyOf(ec.edited, r)
	if !ok {
		return failEdit(exitUsage, fmt.Errorf("%q 没有正文区间", name), *asJSON)
	}
	meta := describeRegion(ec.edited, r)
	lines := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")

	type hit struct {
		Line    int      `json:"line"`
		Text    string   `json:"text"`
		Context []string `json:"context,omitempty"`
	}
	needle := pattern
	if !*caseSensitive {
		needle = strings.ToLower(pattern)
	}
	var hits []hit
	total := 0
	for i, ln := range lines {
		hay := ln
		if !*caseSensitive {
			hay = strings.ToLower(ln)
		}
		if !strings.Contains(hay, needle) {
			continue
		}
		total++
		if len(hits) >= *limit {
			continue
		}
		lo := i - *context
		if lo < 0 {
			lo = 0
		}
		hi := i + *context
		if hi >= len(lines) {
			hi = len(lines) - 1
		}
		hits = append(hits, hit{Line: meta.Line0 + i, Text: ln, Context: lines[lo : hi+1]})
	}

	if *asJSON {
		common.EmitJSON(os.Stdout, map[string]any{
			"ok": true, "region": meta, "pattern": pattern,
			"total": total, "shown": len(hits), "hits": hits,
		})
		return 0
	}
	w := os.Stdout
	if total == 0 {
		common.Line(w, "区域 %s（%d 行）里没有 %q", meta.Name, meta.Lines, pattern)
		return 0
	}
	common.Line(w, "区域 %s（%d 行）：命中 %d 条，显示 %d 条", meta.Name, meta.Lines, total, len(hits))
	for _, h := range hits {
		if *context > 0 {
			for k, c := range h.Context {
				no := h.Line - (len(h.Context) / 2) + k
				mark := " "
				if no == h.Line {
					mark = ">"
				}
				common.Line(w, "%sL%-6d| %s", mark, no, c)
			}
			common.Line(w, "")
			continue
		}
		common.Line(w, "L%-6d| %s", h.Line, strings.TrimRight(h.Text, "\r"))
	}
	return 0
}

// ---------- patch ----------

func cmdPatchRegion(args []string) int {
	fs := flag.NewFlagSet("patch", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "输出 JSON")
	dryRun := fs.Bool("dry-run", false, "只预览替换结果，不写盘")
	oldText := fs.String("old", "", "要替换的原文（与 --new 配对；必须在区域内唯一）")
	newText := fs.String("new", "", "替换后的文本")
	oldFile := fs.String("old-file", "", "从文件读原文（多行用）")
	newFile := fs.String("new-file", "", "从文件读新文本（多行用）")
	line := fs.Int("line", 0, "按行改：prog.full.4gl 的行号（show/find 打印的就是它）")
	text := fs.String("text", "", "按行改时的整行新内容（保留原缩进）")
	if err := common.ParseArgs(fs, args, "old", "new", "old-file", "new-file", "line", "text"); err != nil {
		return exitUsage
	}
	dir, name, derr := resolveDirAndName(fs.Args())
	if derr != nil {
		fmt.Fprintln(os.Stderr, "用法：tt dev tzc patch [<dir>] <点名> (--line N --text S | --old S --new S | --old-file F --new-file F) [--dry-run] [--json]")
		return failEdit(exitUsage, derr, *asJSON)
	}
	ec, err := openForEdit(dir)
	if err != nil {
		return failEdit(exitIO, err, *asJSON)
	}
	r := findRegion(ec.parsed.Doc, name)
	if r == nil {
		return failEdit(exitUsage, fmt.Errorf("工作区里没有点名 %q（先 tt dev tzc points --match %s）", name, name), *asJSON)
	}
	if !r.Editable {
		return failEdit(exitDenied, fmt.Errorf("%q 不可编辑：%s（%s）", name, r.DenyCode, r.Reason), *asJSON)
	}
	body, ok := bodyOf(ec.edited, r)
	if !ok {
		return failEdit(exitUsage, fmt.Errorf("%q 没有正文区间", name), *asJSON)
	}
	meta := describeRegion(ec.edited, r)

	var newBody string
	var changedOffset int // 改动落点（用于报"落在哪个内嵌点"）
	switch {
	case *line > 0:
		if meta.Line0 == 0 || *line < meta.Line0 || *line > meta.Line1 {
			return failEdit(exitUsage, fmt.Errorf("--line %d 不在区域 %s 的正文范围（%d-%d）",
				*line, name, meta.Line0, meta.Line1), *asJSON)
		}
		idx := *line - meta.Line0
		s, ce, le, ok := lineSpan(body, idx)
		if !ok {
			return failEdit(exitUsage, fmt.Errorf("--line %d 定位不到正文行", *line), *asJSON)
		}
		// 保留原行缩进：--text 只给"内容"，缩进沿用原行
		indent := int64(0)
		for int(s)+int(indent) < ce && (body[s+int(indent)] == ' ' || body[s+int(indent)] == '\t') {
			indent++
		}
		var b bytes.Buffer
		b.Write(body[:s])
		b.Write(body[s : s+int(indent)])
		b.WriteString(*text)
		b.Write(body[ce:le])
		b.Write(body[le:])
		newBody = b.String()
		changedOffset = int(s) + int(indent)
	default:
		oldStr, err := pickText(*oldText, *oldFile, "--old", "--old-file")
		if err != nil {
			return failEdit(exitUsage, err, *asJSON)
		}
		// --new 允许显式空串（= 删除）：用 fs.Visit 区分"没给"与"给了空"
		seen := map[string]bool{}
		fs.Visit(func(f *flag.Flag) { seen[f.Name] = true })
		newStr := *newText
		if strings.TrimSpace(*newFile) != "" {
			b, rerr := os.ReadFile(*newFile)
			if rerr != nil {
				return failEdit(exitUsage, fmt.Errorf("读 --new-file 失败（%s）: %w", *newFile, rerr), *asJSON)
			}
			newStr = string(b)
		} else if !seen["new"] {
			return failEdit(exitUsage, fmt.Errorf("缺文本：给 --new 或 --new-file"), *asJSON)
		}
		if strings.TrimSpace(oldStr) == "" {
			return failEdit(exitUsage, fmt.Errorf("原文不能为空（整段替换用 set）"), *asJSON)
		}
		// 围栏两侧都禁：old 里带围栏 = 想删围栏；new 里带围栏 = 想造假围栏
		if bad := fenceToken(oldStr); bad != "" {
			return failEdit(exitDenied, fmt.Errorf("原文里含围栏标记 %q：围栏是坐标，不能作为替换对象", bad), *asJSON)
		}
		if bad := fenceToken(newStr); bad != "" {
			return failEdit(exitDenied, fmt.Errorf("新文本里含围栏标记 %q：不能手写围栏", bad), *asJSON)
		}
		n := strings.Count(string(body), oldStr)
		switch {
		case n == 0:
			hint := nearestLines(body, oldStr)
			msg := fmt.Errorf("%q 里找不到要替换的原文", name)
			if hint != "" {
				msg = fmt.Errorf("%s；区域内最接近的行：\n%s", msg, hint)
			}
			return failEdit(exitDenied, msg, *asJSON)
		case n > 1:
			return failEdit(exitDenied, fmt.Errorf("%q 里命中 %d 处，不唯一；请带上更多上下文，或改用 --line N --text", name, n), *asJSON)
		}
		off := strings.Index(string(body), oldStr)
		// 子串模式的行边界检查：替换后若会拼出"同一行两个 #"或吃掉/多出换行，直接拒
		if why := lineAsymmetry(string(body), off, oldStr, newStr); why != "" {
			return failEdit(exitDenied, fmt.Errorf("%s（改用 --line N --text 更安全）", why), *asJSON)
		}
		newBody = strings.Replace(string(body), oldStr, newStr, 1)
		changedOffset = off
	}

	if strings.TrimRight(string(body), "\r\n") != string(body) && !strings.HasSuffix(newBody, "\n") {
		newBody += "\n"
	}
	start, end, _ := bodyRange(ec.edited, r)
	out := make([]byte, 0, len(ec.edited)+len(newBody)-len(body))
	out = append(out, ec.edited[:start]...)
	out = append(out, newBody...)
	out = append(out, ec.edited[end:]...)

	// 改动落点实际属于哪个内嵌点（区段里的改动写回时按点处理）
	inner := ""
	if p := innerPointAt(ec.edited, ec.parsed.Doc, start+changedOffset); p != nil && p.Name != r.Name {
		inner = p.Name
	}

	if *dryRun {
		preview := previewAround(newBody, changedOffset, 3)
		if *asJSON {
			common.EmitJSON(os.Stdout, map[string]any{"ok": true, "dryRun": true, "region": name,
				"innerPoint": inner, "beforeBytes": len(body), "afterBytes": len(newBody), "preview": preview})
			return 0
		}
		common.Line(os.Stdout, "预览（未写盘）：%s%s", name, innerNote(inner))
		for _, ln := range preview {
			common.Line(os.Stdout, "  %s", ln)
		}
		common.Line(os.Stdout, "字节：%d → %d；确认无误后去掉 --dry-run 重跑", len(body), len(newBody))
		return 0
	}

	if err := ec.ws.WriteEdited(out); err != nil {
		return failEdit(exitIO, err, *asJSON)
	}
	note := sectionRewriteNote(r)
	if *asJSON {
		common.EmitJSON(os.Stdout, map[string]any{"ok": true, "region": name, "innerPoint": inner,
			"beforeBytes": len(body), "afterBytes": len(newBody), "note": note})
		return 0
	}
	common.Line(os.Stdout, "已改 %s%s：%d → %d 字节；%s", name, innerNote(inner), len(body), len(newBody), note)
	return 0
}

// innerNote 说明"改动实际落在哪个内嵌点"。
func innerNote(inner string) string {
	if inner == "" {
		return ""
	}
	return "（落在内嵌点 " + inner + "，写回时按点处理）"
}

// sectionRewriteNote 改区段正文的代价提醒。
func sectionRewriteNote(r *model.Region) string {
	if r.Kind == model.RegionSection {
		return "注意：改区段正文会触发整段重写（TglTag 折叠 + .tgl 补丁 + section_flag=Y），写回面可能明显增删字节"
	}
	return "落包走 verify/apply"
}

// fenceToken 找出文本里的围栏标记片段。
func fenceToken(s string) string {
	for _, bad := range []string{"{//@tdev:", "{/*@tdev:"} {
		if i := strings.Index(s, bad); i >= 0 {
			end := i + len(bad)
			for end < len(s) && s[end] != '\n' {
				end++
			}
			return s[i:end]
		}
	}
	return ""
}

// lineAsymmetry 拦住"子串替换把行结构拼坏"的典型情形。
//
// 实测踩过的坑：原文是 `#檢查失敗時後續處理`（不含行首 #、不含行尾换行），
// 新文本自带 `#` 与换行，结果拼出 `##檢查失敗時後續處理` + 多一个空行，退出码还是 0。
func lineAsymmetry(body string, off int, oldStr, newStr string) string {
	// 先判"行尾多出换行"，再退化成多行判断（多行替换的其它判据不适用）
	if !strings.HasSuffix(oldStr, "\n") && strings.HasSuffix(newStr, "\n") {
		return "新文本以换行结尾而原文不是：会凭空多出一个空行"
	}
	if strings.Contains(oldStr, "\n") || strings.Contains(newStr, "\n") {
		return ""
	}
	// 实测踩过的坑：old 不含行首 `#`、new 自带 `#` → 拼出 `##`
	if off > 0 && off <= len(body) && body[off-1] == '#' && strings.HasPrefix(newStr, "#") {
		return "替换后会出现连续的 ##（原文前面已经有一个 #）"
	}
	return ""
}

// nearestLines 在区域里找与原文"忽略首尾空白后"最接近的行，供定位失败时参考。
func nearestLines(body []byte, oldStr string) string {
	key := strings.TrimSpace(oldStr)
	if key == "" {
		return ""
	}
	var out []string
	lines := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if t == "" || len(t) < 3 {
			continue
		}
		if strings.Contains(ln, key) || strings.Contains(key, t) {
			out = append(out, fmt.Sprintf("  第 %d 行（区域内）：%s", i+1, ln))
			if len(out) >= 3 {
				break
			}
		}
	}
	return strings.Join(out, "\n")
}

// previewAround 取改动点周围几行。
func previewAround(body string, off, ctx int) []string {
	if off > len(body) {
		off = len(body)
	}
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	at := strings.Count(body[:off], "\n")
	lo, hi := at-ctx, at+ctx
	if lo < 0 {
		lo = 0
	}
	if hi >= len(lines) {
		hi = len(lines) - 1
	}
	var out []string
	for i := lo; i <= hi; i++ {
		mark := " "
		if i == at {
			mark = ">"
		}
		out = append(out, fmt.Sprintf("%s%s", mark, lines[i]))
	}
	return out
}

// ---------- set ----------

func cmdSetRegion(args []string) int {
	fs := flag.NewFlagSet("set", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "输出 JSON")
	dryRun := fs.Bool("dry-run", false, "只预览，不写盘")
	file := fs.String("file", "", "新正文所在文件（- 表示 stdin）")
	fromStdin := fs.Bool("stdin", false, "从 stdin 读新正文")
	maxLines := fs.Int("max-lines", 300, "允许整段替换的最大行数（超过请用 patch）")
	if err := common.ParseArgs(fs, args, "file", "max-lines"); err != nil {
		return exitUsage
	}
	dir, name, derr := resolveDirAndName(fs.Args())
	if derr != nil {
		fmt.Fprintln(os.Stderr, "用法：tt dev tzc set [<dir>] <点名> (--file F | --stdin) [--max-lines N] [--dry-run] [--json]")
		return failEdit(exitUsage, derr, *asJSON)
	}
	var data []byte
	var err error
	switch {
	case *fromStdin || *file == "-":
		data, err = readAll(os.Stdin)
	case *file != "":
		data, err = os.ReadFile(*file)
	default:
		err = fmt.Errorf("要给新正文：--file <路径> 或 --stdin")
	}
	if err != nil {
		return failEdit(exitUsage, err, *asJSON)
	}
	if bad := fenceToken(string(data)); bad != "" {
		return failEdit(exitDenied, fmt.Errorf("新正文里含围栏标记 %q：不能手写围栏", bad), *asJSON)
	}

	ec, err := openForEdit(dir)
	if err != nil {
		return failEdit(exitIO, err, *asJSON)
	}
	r := findRegion(ec.parsed.Doc, name)
	if r == nil {
		return failEdit(exitUsage, fmt.Errorf("工作区里没有点名 %q", name), *asJSON)
	}
	if !r.Editable {
		return failEdit(exitDenied, fmt.Errorf("%q 不可编辑：%s（%s）", name, r.DenyCode, r.Reason), *asJSON)
	}
	body, ok := bodyOf(ec.edited, r)
	if !ok {
		return failEdit(exitUsage, fmt.Errorf("%q 没有正文区间", name), *asJSON)
	}
	if bytes.Contains(body, []byte("{//@tdev:begin point")) {
		return failEdit(exitDenied, fmt.Errorf("%q 是含内嵌加点的区段：整段替换会丢围栏（写回时无法折叠回占位符）；"+
			"请用 find + patch 改某一段", name), *asJSON)
	}
	if n := regionLines(body); n > *maxLines {
		return failEdit(exitDenied, fmt.Errorf("%q 现有 %d 行，超过整段替换上限 %d；"+
			"大区域请用 find + patch（不要求整段重写）", name, n, *maxLines), *asJSON)
	}

	newBody := string(data)
	if strings.HasSuffix(string(body), "\n") && !strings.HasSuffix(newBody, "\n") {
		newBody += "\n"
	}
	if *dryRun {
		if *asJSON {
			common.EmitJSON(os.Stdout, map[string]any{"ok": true, "dryRun": true, "region": name,
				"beforeLines": regionLines(body), "afterLines": regionLines([]byte(newBody)),
				"beforeBytes": len(body), "afterBytes": len(newBody)})
			return 0
		}
		common.Line(os.Stdout, "预览（未写盘）：%s %d 行 → %d 行，%d → %d 字节",
			name, regionLines(body), regionLines([]byte(newBody)), len(body), len(newBody))
		return 0
	}
	start, end, _ := bodyRange(ec.edited, r)
	out := make([]byte, 0, len(ec.edited)+len(newBody)-len(body))
	out = append(out, ec.edited[:start]...)
	out = append(out, newBody...)
	out = append(out, ec.edited[end:]...)
	if err := ec.ws.WriteEdited(out); err != nil {
		return failEdit(exitIO, err, *asJSON)
	}
	if *asJSON {
		common.EmitJSON(os.Stdout, map[string]any{"ok": true, "region": name,
			"beforeLines": regionLines(body), "afterLines": regionLines([]byte(newBody))})
		return 0
	}
	common.Line(os.Stdout, "已整段替换 %s：%d 行 → %d 行；落包走 verify/apply",
		name, regionLines(body), regionLines([]byte(newBody)))
	return 0
}

// ---------- 小工具 ----------

func pickText(inline, file, inlineFlag, fileFlag string) (string, error) {
	if strings.TrimSpace(file) != "" {
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("读 %s 失败（%s）: %w", fileFlag, file, err)
		}
		return string(b), nil
	}
	if inline == "" {
		return "", fmt.Errorf("缺文本：给 %s 或 %s", inlineFlag, fileFlag)
	}
	return inline, nil
}

func readAll(f *os.File) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
