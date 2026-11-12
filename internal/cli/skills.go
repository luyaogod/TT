// `tt skills` —— 把技能分发到项目里的 AI 工具目录。
//
//	tt skills list   [--project <dir>] [--brief] [--json]
//	tt skills deploy [--project <dir>] [--all | --agent <id> ...]
//	                 [--skill <name> ...] [--dry-run] [--json]
//	tt skills remove [--project <dir>] [--agent <id> ...] [--yes] [--dry-run]
//
// 这是「分发」的管理面：平台表、清单、覆盖/删除都在 internal/skill。`tt install skills`
// 仍是单目标底层（`tt update` 刷新用它），两条命令共用同一份复制与 assets 指针实现。
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
	"tt/internal/output"
	"tt/internal/skill"
)

func newSkillsCmd() *cobra.Command {
	var (
		project string
		global  bool
	)
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "把技能分发到 AI 工具目录（项目级或全局）",
		Long: `把 TT 的技能分发到五个 AI 工具目录：默认落在当前项目，加 --global 落在用户目录。

平台（目录相对项目根 / 用户目录）：

  agents      .agents/skills           跨客户端公约数
  dsh         .dsh/skills              DeepSeek Harness
  workbuddy   .codebuddy/skills        WorkBuddy
  claude      .claude/skills           Claude Code
  codex       .codex/skills            Codex

同步是**覆盖更新**：勾选的技能直接覆盖同名目录；取消勾选的（清单里记过的）删除。
不做版本检测，不比较内容，也不需要 --force。

清单只记“哪个平台装了哪些技能”，落在对应范围的 .agents/.tt-manifest.json：
项目级在 <项目>/.agents/.tt-manifest.json，全局在 <用户目录>/.agents/.tt-manifest.json。
`,
		Example: `  tt skills list
  tt skills deploy --all
  tt skills deploy --agent claude --agent codex
  tt skills deploy --global --agent claude
  tt skills deploy --agent dsh --skill tt-erp-read --dry-run
  tt skills remove --global --agent claude --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("请指定子命令: list / deploy / remove\n\n%s", cmd.UsageString())
			}
			return common.UnknownSubcommand(cmd, args)
		},
	}
	cmd.PersistentFlags().StringVar(&project, "project", "", "项目目录（默认当前目录）")
	cmd.PersistentFlags().BoolVar(&global, "global", false, "操作用户目录（全局），而不是当前项目")
	cmd.AddCommand(newSkillsListCmd(&project, &global), newSkillsDeployCmd(&project, &global), newSkillsRemoveCmd(&project, &global))
	return cmd
}

//---------------------------------------------------------------------------
// list
//---------------------------------------------------------------------------

func newSkillsListCmd(project *string, global *bool) *cobra.Command {
	var brief bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "看技能源、五个平台的检测与已分发情况",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			base, scope, err := resolveSkillsBase(*project, *global)
			if err != nil {
				return err
			}
			plan, err := buildSkillsPlan()
			if err != nil {
				return err
			}
			m, _, err := skill.LoadManifest(base)
			if err != nil {
				return err
			}

			targets := make([]skillListTarget, 0, len(skill.Targets))
			for _, t := range skill.Targets {
				targets = append(targets, skillListTarget{
					ID: t.ID, Name: t.Name, Dir: t.Dir,
					Detected: t.Detected(base), Skills: m.Skills(t.ID),
				})
			}

			if common.IsJSON() {
				if brief {
					return common.PrintJSON(map[string]any{
						"ok": true, "scope": scope, "project": base,
						"skills":  skillNames(plan.Skills),
						"targets": targets,
					})
				}
				rows := make([]skillListRow, 0, len(plan.Skills))
				for _, s := range plan.Skills {
					rows = append(rows, skillListRow{
						Name: s.Name, Origin: string(s.Origin), Dir: s.Dir, Available: true,
					})
				}
				for _, w := range plan.Warnings {
					rows = append(rows, skillListRow{
						Name: w.Name, Origin: string(skill.OriginExternal), Dir: w.Path,
						Available: false, Why: w.Why,
					})
				}
				return common.PrintJSON(map[string]any{
					"ok": true, "scope": scope, "project": base, "skills": rows, "targets": targets, "warnings": plan.Warnings,
				})
			}

			fmt.Printf("%s %s\n", scopeLabel(scope), base)
			if brief {
				fmt.Printf("技能 %d 个：%s\n", len(plan.Skills), strings.Join(skillNames(plan.Skills), "、"))
			} else {
				fmt.Printf("技能 %d 个\n", len(plan.Skills))
				for _, s := range plan.Skills {
					fmt.Printf("  %-24s %s\n", s.Name, originLabel(s.Origin))
				}
				for _, w := range plan.Warnings {
					fmt.Printf("  %-24s 不可用：%s\n", w.Name, w.Why)
				}
			}
			fmt.Println("平台")
			for _, t := range targets {
				mark := "未检测"
				if t.Detected {
					mark = "已检测"
				}
				fmt.Printf("  %-10s %-22s %-6s %d 个\n", t.ID, t.Dir, mark, len(t.Skills))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&brief, "brief", false, "只列技能名与平台计数")
	return cmd
}

//---------------------------------------------------------------------------
// deploy
//---------------------------------------------------------------------------

func newSkillsDeployCmd(project *string, global *bool) *cobra.Command {
	var (
		all    bool
		agents []string
		skills []string
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "把技能覆盖写入选中的平台目录",
		Long: `把技能覆盖写入选中的平台目录。

默认平台 = 已检测到的 + .agents；--all = 五个全写（没检测到的会建目录）。
默认技能 = 全部可用技能（官方 + 登记且合规的第三方）；--skill 可只挑几个。
取消勾选（不带某个 --skill）不会删除已装技能 —— 删除用 tt skills remove。
`,
		Example: `  tt skills deploy --all
  tt skills deploy --agent claude --agent codex
  tt skills deploy --agent dsh --skill tt-erp-read --dry-run`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if all && len(agents) > 0 {
				return output.Errorf("USAGE", 2, "--all 与 --agent 只能给一个")
			}
			base, scope, err := resolveSkillsBase(*project, *global)
			if err != nil {
				return err
			}
			plan, err := buildSkillsPlan()
			if err != nil {
				return err
			}
			names, err := pickSkillNames(plan, skills)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				return output.Errorf("USAGE", 2, "没有可用技能")
			}
			targets, err := pickTargets(base, all, agents)
			if err != nil {
				return err
			}
			req := skill.DeployRequest{Project: base, DryRun: dryRun}
			for _, t := range targets {
				req.Targets = append(req.Targets, skill.DeployTarget{ID: t.ID, Skills: names})
			}
			rep, err := skill.Deploy(plan, req)
			if err != nil {
				return err
			}
			return printDeployReport(rep, scope)
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "五个平台全部写入（含未检测，会建目录）")
	cmd.Flags().StringArrayVar(&agents, "agent", nil, "平台："+strings.Join(skill.TargetIDs(), " / ")+"（可重复）")
	cmd.Flags().StringArrayVar(&skills, "skill", nil, "只分发这些技能（可重复；默认全部可用）")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只算不写")
	return cmd
}

//---------------------------------------------------------------------------
// remove
//---------------------------------------------------------------------------

func newSkillsRemoveCmd(project *string, global *bool) *cobra.Command {
	var (
		agents []string
		yes    bool
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "删除清单里记录的技能目录（只删 TT 装的）",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			base, scope, err := resolveSkillsBase(*project, *global)
			if err != nil {
				return err
			}
			plan, err := buildSkillsPlan()
			if err != nil {
				return err
			}
			targets, err := pickTargetsForRemove(agents)
			if err != nil {
				return err
			}
			if !dryRun && !yes {
				return output.Errorf("USAGE", 2, "删除要加 --yes（先加 --dry-run 看一眼会删什么）")
			}
			req := skill.DeployRequest{Project: base, DryRun: dryRun}
			for _, t := range targets {
				req.Targets = append(req.Targets, skill.DeployTarget{ID: t.ID, Skills: nil})
			}
			rep, err := skill.Deploy(plan, req)
			if err != nil {
				return err
			}
			return printDeployReport(rep, scope)
		},
	}
	cmd.Flags().StringArrayVar(&agents, "agent", nil, "只清这些平台（默认全部）")
	cmd.Flags().BoolVar(&yes, "yes", false, "确认删除")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只算不删")
	return cmd
}

//---------------------------------------------------------------------------
// 共用
//---------------------------------------------------------------------------

type skillListRow struct {
	Name      string `json:"name"`
	Origin    string `json:"origin"`
	Dir       string `json:"dir,omitempty"`
	Available bool   `json:"available"`
	Why       string `json:"why,omitempty"`
}

type skillListTarget struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Dir      string   `json:"dir"`
	Detected bool     `json:"detected"`
	Skills   []string `json:"skills"`
}

// buildSkillsPlan 读 <exe>/skills 与登记表，组一份安装计划。
func buildSkillsPlan() (skill.Plan, error) {
	src, err := skillsSource()
	if err != nil {
		return skill.Plan{}, err
	}
	if !isDir(src) {
		return skill.Plan{}, fmt.Errorf("找不到 skills 目录: %s\n\nskills/ 与 tt.exe 同目录（免安装包里自带）；"+
			"从源码运行时先执行 make build", src)
	}
	externals, err := externalSkillPaths()
	if err != nil {
		return skill.Plan{}, err
	}
	return skill.BuildPlan(src, externals)
}

// resolveProjectDir 解析 --project；缺省取当前目录，必须是已存在的目录。
func resolveProjectDir(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
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
			return "", output.Errorf("NOT_FOUND", 4, "项目目录不存在: %s", abs)
		}
		return "", fmt.Errorf("读项目目录失败 %s: %w", abs, err)
	}
	if !st.IsDir() {
		return "", output.Errorf("NOT_FOUND", 4, "项目路径不是目录: %s", abs)
	}
	return abs, nil
}

// homeDir 可替换（测试用）：否则 --global 的用例会写到开发者真实的用户目录。
var homeDir = os.UserHomeDir

// resolveSkillsBase 把 --project / --global 解析成一个分发根目录与范围名。
//
// 全局就是“根目录换成用户目录”，平台相对目录不变 —— 例如 Claude 的全局落点是
// <用户目录>/.claude/skills。清单也跟着根目录走：<用户目录>/.agents/.tt-manifest.json。
func resolveSkillsBase(project string, global bool) (base, scope string, err error) {
	if global {
		if strings.TrimSpace(project) != "" {
			return "", "", output.Errorf("USAGE", 2, "--global 与 --project 只能给一个")
		}
		home, herr := homeDir()
		if herr != nil {
			return "", "", fmt.Errorf("取不到用户目录: %w", herr)
		}
		st, serr := os.Stat(home)
		if serr != nil || !st.IsDir() {
			return "", "", output.Errorf("NOT_FOUND", 4, "用户目录不存在: %s", home)
		}
		return home, "global", nil
	}
	base, err = resolveProjectDir(project)
	if err != nil {
		return "", "", err
	}
	return base, "project", nil
}

func scopeLabel(scope string) string {
	if scope == "global" {
		return "全局"
	}
	return "项目"
}

func pickSkillNames(plan skill.Plan, want []string) ([]string, error) {
	if len(want) == 0 {
		out := make([]string, 0, len(plan.Skills))
		for _, s := range plan.Skills {
			out = append(out, s.Name)
		}
		sort.Strings(out)
		return out, nil
	}
	byName := make(map[string]bool, len(plan.Skills))
	for _, s := range plan.Skills {
		byName[s.Name] = true
	}
	seen := make(map[string]bool, len(want))
	out := make([]string, 0, len(want))
	for _, raw := range want {
		n := strings.TrimSpace(raw)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		if !byName[n] {
			why := ""
			for _, w := range plan.Warnings {
				if w.Name == n {
					why = w.Why
					break
				}
			}
			if why != "" {
				return nil, output.Errorf("NOT_FOUND", 4, "技能 %s 不可用：%s", n, why)
			}
			return nil, output.Errorf("NOT_FOUND", 4, "技能 %s 不在可用技能里", n)
		}
		out = append(out, n)
	}
	sort.Strings(out)
	return out, nil
}

func pickTargets(proj string, all bool, agents []string) ([]skill.Target, error) {
	if all {
		return append([]skill.Target(nil), skill.Targets...), nil
	}
	if len(agents) > 0 {
		return targetsFromNames(agents)
	}
	out := make([]skill.Target, 0, len(skill.Targets))
	for _, t := range skill.Targets {
		if t.ID == "agents" || t.Detected(proj) {
			out = append(out, t)
		}
	}
	return out, nil
}

func pickTargetsForRemove(agents []string) ([]skill.Target, error) {
	if len(agents) == 0 {
		return append([]skill.Target(nil), skill.Targets...), nil
	}
	return targetsFromNames(agents)
}

func targetsFromNames(names []string) ([]skill.Target, error) {
	seen := make(map[string]bool, len(names))
	out := make([]skill.Target, 0, len(names))
	for _, raw := range names {
		t, ok := skill.TargetByID(raw)
		if !ok {
			return nil, output.Errorf("USAGE", 2, "不认识的平台: %s（可选：%s）",
				raw, strings.Join(skill.TargetIDs(), " / "))
		}
		if seen[t.ID] {
			continue
		}
		seen[t.ID] = true
		out = append(out, t)
	}
	return out, nil
}

func skillNames(src []skill.Source) []string {
	out := make([]string, 0, len(src))
	for _, s := range src {
		out = append(out, s.Name)
	}
	sort.Strings(out)
	return out
}

func originLabel(o skill.Origin) string {
	if o == skill.OriginOfficial {
		return "官方"
	}
	return "第三方"
}

func printDeployReport(rep skill.DeployReport, scope string) error {
	if common.IsJSON() {
		return common.PrintJSON(map[string]any{
			"ok": true, "scope": scope, "project": rep.Project, "dryRun": rep.DryRun,
			"written": rep.Written, "removed": rep.Removed,
			"skipped": rep.Skipped, "warnings": rep.Warnings,
		})
	}
	fmt.Printf("%s %s\n", scopeLabel(scope), rep.Project)
	if rep.DryRun {
		fmt.Println("（试算，未写入任何文件）")
	}
	files, removed := 0, len(rep.Removed)
	perTarget := make(map[string]int, len(rep.Written))
	for _, a := range rep.Written {
		files += len(a.Files)
		perTarget[a.Target] += len(a.Files)
	}
	verb := "已同步"
	if rep.DryRun {
		verb = "将同步"
	}
	fmt.Printf("%s：写入 %d 个文件，删除 %d 个目录\n", verb, files, removed)
	for _, t := range skill.Targets {
		if n := perTarget[t.ID]; n > 0 {
			fmt.Printf("  %-10s %-22s 写入 %d 个文件\n", t.ID, t.Dir, n)
		}
	}
	for _, a := range rep.Removed {
		fmt.Printf("  删除 %s/%s\n", a.Target, a.Skill)
	}
	if len(rep.Skipped) > 0 {
		fmt.Fprintf(os.Stderr, "跳过 %d 条：\n", len(rep.Skipped))
		for _, s := range rep.Skipped {
			fmt.Fprintf(os.Stderr, "  %s/%s：%s\n", s.Target, s.Skill, s.Why)
		}
	}
	if len(rep.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "告警 %d 条（不影响其它技能）：\n", len(rep.Warnings))
		for _, w := range rep.Warnings {
			fmt.Fprintf(os.Stderr, "  %s\n", w.Text())
		}
	}
	return nil
}
