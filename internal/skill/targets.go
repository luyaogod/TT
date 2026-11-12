package skill

import (
	"os"
	"path/filepath"
	"strings"
)

// Target 是一个 agent 工具的技能落点（项目级与全局共用这张表）。
//
// 表驱动：加一个平台只加一行，前端、CLI、同步逻辑共用这张表。目录一律相对**根目录**：
// 项目级时根是项目目录，--global / 设置页「全局」时根是用户目录（如 ~/.claude/skills）。
type Target struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Root    string   `json:"-"`   // 检测路标：<project>/<Root>
	Dir     string   `json:"dir"` // 技能目录：<project>/<Dir>
	Aliases []string `json:"-"`
	Who     string   `json:"who,omitempty"`
}

// Targets 是 v1 适配的五个平台。
//
// 路径依据：
//   - `.agents/skills` 是 Agent Skills 的跨客户端公约数；DeepSeek Harness 也读它（rank 200）。
//   - DeepSeek Harness 的项目级目录是 `.dsh/skills`（rank 100，优先于 `.agents`）。
//   - WorkBuddy 读工作目录下的 `.codebuddy/skills`（官方「项目」文档）。
//   - Claude Code 读 `.claude/skills`。
//   - Codex 读 repo-scoped `.codex/skills`。
//
// 顺序就是设置页的列顺序；`.agents` 放第一列（读它的客户端最多）。
var Targets = []Target{
	{ID: "agents", Name: ".agents", Root: ".agents", Dir: ".agents/skills", Aliases: []string{"universal"},
		Who: "跨客户端公约数"},
	{ID: "dsh", Name: "DeepSeek Harness", Root: ".dsh", Dir: ".dsh/skills",
		Aliases: []string{"deepseek", "deepseek-harness"}, Who: "DeepSeek Harness 项目技能目录"},
	{ID: "workbuddy", Name: "WorkBuddy", Root: ".codebuddy", Dir: ".codebuddy/skills",
		Aliases: []string{"wb", "work-buddy"}, Who: "WorkBuddy 项目技能目录"},
	{ID: "claude", Name: "Claude Code", Root: ".claude", Dir: ".claude/skills",
		Aliases: []string{"claude-code"}, Who: "Claude Code 项目技能目录"},
	{ID: "codex", Name: "Codex", Root: ".codex", Dir: ".codex/skills",
		Aliases: []string{"codex-cli"}, Who: "Codex 项目技能目录"},
}

// TargetByID 按 id 或别名找平台（大小写不敏感）；id 是唯一权威。
func TargetByID(name string) (Target, bool) {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, t := range Targets {
		if t.ID == n {
			return t, true
		}
		for _, a := range t.Aliases {
			if a == n {
				return t, true
			}
		}
	}
	return Target{}, false
}

// TargetIDs 返回全部 id（帮助与错误信息用）。
func TargetIDs() []string {
	out := make([]string, 0, len(Targets))
	for _, t := range Targets {
		out = append(out, t.ID)
	}
	return out
}

// AbsDir 返回某个项目下这个平台的技能目录（绝对路径）。
func (t Target) AbsDir(project string) string {
	return filepath.Join(project, filepath.FromSlash(t.Dir))
}

// Detected 判断平台是否已经在项目里露过面。
//
// 对 `.agents` 来说，清单文件也会把它建出来 —— 那也算“存在”：后续同步本来就要往这里写。
func (t Target) Detected(project string) bool {
	st, err := os.Stat(filepath.Join(project, filepath.FromSlash(t.Root)))
	return err == nil && st.IsDir()
}

// DetectedTargets 返回项目里已检测到的平台（保持表顺序）。
func DetectedTargets(project string) []Target {
	out := make([]Target, 0, len(Targets))
	for _, t := range Targets {
		if t.Detected(project) {
			out = append(out, t)
		}
	}
	return out
}
