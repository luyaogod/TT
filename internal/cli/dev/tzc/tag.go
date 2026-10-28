package tzc

// 本次修改的"唯一标识"门禁。
//
// 为什么要有这一层：`tt-dev-tzc` 技能要求开工前向用户索要 `<单号>#<项次>` 与工号
// （用户拒绝时用 16 位 uuid + agent），并把它们写进 global.memo 的摘要行与修改块。
// 但**只写在技能里是拦不住的** —— agent 可以不读、读了也可以跳过，apply 照样成功。
// 所以这里把同一规则变成 apply 的前置闸门：没有标记/工号、标记格式不对、或
// prog.full.4gl 里找不到对应摘要行，一律拒绝（退出码 3，与验证失败同类）。
//
// 只校验"有没有、对不对得上"，不替用户编标记：编造单号比缺标记更难追查。

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// reModTag 单号#项次：两段都允许字母数字与 _-，项次允许点号（如 A12345#01、DOC-7#1.2）。
	reModTag = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,15}#[A-Za-z0-9][A-Za-z0-9_.-]{0,15}$`)
	// reModUUID 用户拒绝提供时自动生成的标记：不带分隔符的 16 位十六进制。
	reModUUID = regexp.MustCompile(`^[0-9a-f]{16}$`)
	// reModOperator 工号：字母数字与 _-.；用户拒绝时固定写 agent。
	reModOperator = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,31}$`)
	// reModBlock 修改块的开/闭行：`#<tag> -s` / `#<tag> -e`（行尾可有空白）。
	reModBlock = regexp.MustCompile(`^[ \t]*#(\S+)[ \t]+-(s|e)[ \t]*$`)
)

// checkModTag 校验本次修改的标记与工号是否已写进工作区源码。
// 返回 nil 表示通过；否则是给人看的问题列表（每条一句，可直接打印）。
func checkModTag(source []byte, tag, operator string) []string {
	tag = strings.TrimSpace(tag)
	operator = strings.TrimSpace(operator)

	var issues []string
	switch {
	case tag == "":
		issues = append(issues, "缺少 --tag：本次修改标记 <单号>#<项次>（用户拒绝提供时用 16 位 uuid）")
	case !reModTag.MatchString(tag) && !reModUUID.MatchString(tag):
		issues = append(issues, fmt.Sprintf("--tag %q 格式不对：应为 <单号>#<项次>（如 A12345#01），或用户拒绝时的 16 位小写十六进制 uuid", tag))
	}
	switch {
	case operator == "":
		issues = append(issues, "缺少 --operator：执行人工号（用户拒绝提供时写 agent）")
	case !reModOperator.MatchString(operator):
		issues = append(issues, fmt.Sprintf("--operator %q 格式不对：只允许字母数字与 _-.（如 12345）", operator))
	}
	if len(issues) > 0 {
		return issues
	}
	if reModUUID.MatchString(tag) && operator != "agent" {
		issues = append(issues, "16 位 uuid 标记只在用户拒绝提供标记时使用，对应 --operator agent")
	}

	// global.memo 里的摘要行：`#<tag> <作业日> By <工号> <一句说明>`
	// 允许行首缩进；"By <工号>" 后必须还有说明文字（光有标记不算写完）。
	memo := regexp.MustCompile(`(?m)^[ \t]*#` + regexp.QuoteMeta(tag) +
		`[ \t]+.*\bBy[ \t]+` + regexp.QuoteMeta(operator) + `\b[ \t]+\S`)
	if !memo.Match(source) {
		issues = append(issues, fmt.Sprintf(
			"prog.full.4gl 里找不到带标记的摘要行：应为 `#%s <作业日> By %s <一句说明>`（加在 global.memo 的 add-point 内）",
			tag, operator))
	}

	if msg := checkModBlocks(source, tag); msg != "" {
		issues = append(issues, msg)
	}
	return issues
}

// checkModBlocks 校验本次标记的修改块是否成对（-s 在前、-e 在后，不嵌套、不悬空）。
// 一个块都没写不算错：只改 memo 的修改也可能存在；写了就必须配平。
func checkModBlocks(source []byte, tag string) string {
	depth := 0
	for i, line := range strings.Split(string(source), "\n") {
		m := reModBlock.FindStringSubmatch(strings.TrimRight(line, "\r"))
		if m == nil || m[1] != tag {
			continue
		}
		switch m[2] {
		case "s":
			depth++
		case "e":
			depth--
			if depth < 0 {
				return fmt.Sprintf("第 %d 行的 `#%s -e` 没有对应的 `-s`", i+1, tag)
			}
		}
	}
	if depth > 0 {
		return fmt.Sprintf("修改块没有闭合：`#%s -s` 比 `-e` 多 %d 个", tag, depth)
	}
	return ""
}
