package at

// --at-list 的渲染。输出是**固定分段**的契约（文法 / 来源 / 可用值 / 示例 / 备注），
// 人和 agent 读同一份；--json 是同一份数据，不是另一套说法。

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"tt/internal/cli/common"
)

type contractValue struct {
	Handle     string `json:"handle"`
	Env        string `json:"env"`
	Ent        int    `json:"ent,omitempty"`
	Account    string `json:"account,omitempty"`
	Zone       string `json:"zone,omitempty"`
	DBTarget   string `json:"dbTarget,omitempty"`
	Source     string `json:"source,omitempty"` // live | snapshot | snapshot-stale
	Stale      bool   `json:"stale,omitempty"`
	AgeSeconds int    `json:"ageSeconds,omitempty"`
	Note       string `json:"note,omitempty"`
}

type contract struct {
	OK          bool            `json:"ok"`
	Command     string          `json:"command"`
	Scope       string          `json:"scope"`
	EntRequired bool            `json:"entRequired,omitempty"`
	Grammar     string          `json:"grammar"`
	Sources     []string        `json:"sources,omitempty"`
	Values      []contractValue `json:"values"`
	Example     string          `json:"example,omitempty"`
	Notes       []string        `json:"notes,omitempty"`
	Message     string          `json:"message,omitempty"`
}

// PrintContract 打一条命令的 --at 契约。w 由调用方给（测试里换成 buffer）。
func PrintContract(w io.Writer, cmd *cobra.Command) error {
	scope := ScopeOf(cmd)
	sc := scope.String()
	if scope == EnvEntReq {
		sc = "env+ent"
	}
	c := contract{OK: true, Command: cmd.CommandPath(), Scope: sc, EntRequired: scope == EnvEntReq}

	switch scope {
	case None:
		c.Message = noneMessage(cmd)
	case Env, EnvFilter:
		c.Grammar = "<环境>"
		c.Sources = []string{"<环境> ← config.json 的 hosts.sshs[].name（本机配置，无网络）"}
		c.Example = exampleOf(cmd, "<环境>")
		c.Notes = append(c.Notes, "环境一律点名：没有默认环境可回落")
		if scope == EnvFilter {
			c.Notes = append(c.Notes,
				"本命令是会话寻址：--at 只用来收窄“在哪个环境里找会话”",
				"指名会话用 --session（tt session ls --json 是它的下拉）；要断言企业用 --session <环境>/<企业>")
		}
		values, err := envValues()
		if err != nil {
			c.Notes = append(c.Notes, "读不到环境清单: "+err.Error())
			break
		}
		c.Values = values
	case EnvEnt, EnvEntReq:
		c.Grammar = "<环境>@<企业>"
		if scope == EnvEnt {
			c.Grammar += "（也可以只写 <环境>，企业可选）"
		}
		c.Sources = []string{
			"<环境> ← config.json 的 hosts.sshs[].name（本机配置，无网络）",
			"<企业> ← 该环境库里的 gzou_t（gzoustus='Y' 的 gzou001；账号取 gzou003）",
			"组合   = 环境名 + \"@\" + 企业编号（企业编号是纯数字）",
		}
		c.Example = exampleOf(cmd, "<环境>@<企业>")
		c.Notes = append(c.Notes,
			"企业部分只读快照，不拨 SSH；没有快照的环境会标出来",
			"要现况：tt env ents --refresh")
		if scope == EnvEntReq {
			c.Notes = append(c.Notes, "本命令必须带 @企业：企业编号决定用哪个账号连库")
		}
		values, notes, err := entValues()
		c.Notes = append(c.Notes, notes...)
		if err != nil {
			c.Notes = append(c.Notes, "读不到企业目录: "+err.Error())
			break
		}
		c.Values = values
	}

	if common.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(c)
	}
	writeHuman(w, c)
	return nil
}

func writeHuman(w io.Writer, c contract) {
	fmt.Fprintf(w, "%s —— --at 取值\n\n", c.Command)
	if c.Message != "" {
		fmt.Fprintf(w, "  %s\n", strings.ReplaceAll(c.Message, "\n", "\n  "))
		return
	}
	fmt.Fprintf(w, "  文法    %s\n", c.Grammar)
	if len(c.Sources) > 0 {
		fmt.Fprintf(w, "  来源\n")
		for _, s := range c.Sources {
			fmt.Fprintf(w, "          %s\n", s)
		}
	}
	if len(c.Values) > 0 {
		fmt.Fprintf(w, "  可用值\n")
		for _, v := range c.Values {
			line := v.Handle
			extra := make([]string, 0, 3)
			if v.Account != "" {
				extra = append(extra, v.Account)
			}
			if v.Zone != "" {
				extra = append(extra, "区域 "+v.Zone)
			}
			if v.Note != "" {
				extra = append(extra, v.Note)
			} else if v.Source != "" {
				fresh := v.Source
				if v.Stale {
					fresh += "（已过期）"
				} else if v.AgeSeconds > 0 {
					fresh += fmt.Sprintf("（%d 分钟前）", v.AgeSeconds/60)
				}
				extra = append(extra, fresh)
			}
			if len(extra) > 0 {
				line += "  " + strings.Join(extra, " · ")
			}
			fmt.Fprintf(w, "          %s\n", line)
		}
	}
	if c.Example != "" {
		fmt.Fprintf(w, "  示例    %s\n", c.Example)
	}
	if len(c.Notes) > 0 {
		fmt.Fprintf(w, "  备注\n")
		for _, n := range c.Notes {
			fmt.Fprintf(w, "          %s\n", n)
		}
	}
}

func envValues() ([]contractValue, error) {
	envs, _, err := LoadEnvs()
	if err != nil {
		return nil, err
	}
	out := make([]contractValue, 0, len(envs))
	for _, e := range envs {
		out = append(out, contractValue{
			Handle: e.Name, Env: e.Name, Zone: e.Zone, DBTarget: e.DBTarget,
		})
	}
	return out, nil
}

// entValues 逐环境读快照。坏一个环境不影响其它环境（返回值里带 note）。
func entValues() ([]contractValue, []string, error) {
	envs, _, err := LoadEnvs()
	if err != nil {
		return nil, nil, err
	}
	out := make([]contractValue, 0, len(envs))
	notes := make([]string, 0, 1)
	for _, e := range envs {
		l, err := SnapshotEnts(e.Name)
		if err != nil {
			out = append(out, contractValue{Handle: e.Name, Env: e.Name, Note: "读企业目录失败: " + err.Error()})
			continue
		}
		if l == nil {
			out = append(out, contractValue{Handle: e.Name, Env: e.Name, Note: "无快照（先跑 tt env ents --refresh）"})
			notes = append(notes, "有环境还没有企业快照 —— 用 tt env ents --refresh 拉一次")
			continue
		}
		for _, m := range l.Ents {
			v := contractValue{
				Handle: fmt.Sprintf("%s@%d", e.Name, m.Ent), Env: e.Name, Ent: m.Ent,
				Zone: e.Zone, DBTarget: e.DBTarget,
				Source: l.Source, Stale: l.Stale, AgeSeconds: l.AgeSeconds,
			}
			if m.Account != "" && m.Account != "-" {
				v.Account = m.Account
			} else {
				v.Account = "未配账号"
			}
			out = append(out, v)
		}
	}
	return out, notes, nil
}

// exampleOf 用命令自己的路径与位置参数拼一条可复制的例子。
func exampleOf(cmd *cobra.Command, atValue string) string {
	usage := cmd.Use
	if i := strings.Index(usage, " "); i >= 0 {
		usage = strings.TrimSpace(usage[i+1:])
	} else {
		usage = ""
	}
	rest := ""
	if usage != "" {
		rest = " " + usage
	}
	return fmt.Sprintf("%s --at %s%s", cmd.CommandPath(), atValue, rest)
}

// noneMessage 解释“这条命令为什么没有 --at”，并给出可执行的替代写法。
func noneMessage(cmd *cobra.Command) string {
	path := pathOf(cmd)
	switch {
	case path == "":
		return "--at 是给具体命令用的：先选命令，再加 --at-list（例如 tt debug start --at-list）。"
	case strings.HasPrefix(path, "session"):
		return "本命令不接受 --at：会话身份来自 --session（tt session ls --json 可看全部）；\n" +
			"环境与企业跟着会话走，不用单独给。"
	case strings.HasPrefix(path, "dev tzs"):
		return "本命令不接受 --at：它的参数由 tzs 自己的解析器接管 —— 环境用 --env，工作区用 --workspace。"
	default:
		return "本命令没有目标参数（不需要环境/企业）。"
	}
}
