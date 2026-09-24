package config

import (
	"fmt"
	"strings"
)

// 环境解析的唯一入口。
//
// 合并前「name → hosts.activeEnv → 首条」这个选择在四处各写了一遍
// (cli/dict/db.go、cli/dict/dbops.go、cli/dict/source.go、internal/debug/config.go),
// 每处的错误文案与边界行为还都不一样。这里收成一份:Hosts.pickName 决定"选哪个",
// Resolve 负责"找到它并说清是怎么选中的"。
//
// 选中的方式会随结果一起回给调用方(EnvRef.From),最终打进查询输出的环境头 ——
// "这个环境是命令行点名的,还是兜底取的首条"对排查"查错了地方"是必需信息。

// EnvRef 一次环境解析的结果。
type EnvRef struct {
	Env  *NamedSsh // 指向 Hosts.SSHs 里的元素(不是拷贝)
	Name string    // 生效环境名
	From string    // 它怎么被选中的:flag / tool-override / activeEnv / first
}

func (r *EnvRef) String() string {
	if r == nil {
		return ""
	}
	return r.Name
}

// pickName 环境名的唯一选择顺序:显式 name → hosts.activeEnv → 首条。
// 返回 (名字, 来源);一个环境都没有时返回 ("", "")。
func (h *Hosts) pickName(name string) (string, string) {
	if name != "" {
		return name, "flag"
	}
	if h.ActiveEnv != "" {
		return h.ActiveEnv, "activeEnv"
	}
	if len(h.SSHs) > 0 {
		return h.SSHs[0].Name, "first"
	}
	return "", ""
}

// Resolve 按名解析环境;name 为空取 activeEnv,再空取首条。
// 名字不存在时报错并列出可用环境名 —— 显式点名的环境拼错了应当报错,
// 绝不静默换一个环境连(那正是"查错了地方却当成没有数据"的根因)。
func (h *Hosts) Resolve(name string) (*EnvRef, error) {
	target, from := h.pickName(name)
	if target == "" {
		return nil, fmt.Errorf("尚未配置 SSH 环境(运行 tt serve 添加,或编辑 config.json hosts.sshs)")
	}
	for i := range h.SSHs {
		if h.SSHs[i].Name == target {
			return &EnvRef{Env: &h.SSHs[i], Name: target, From: from}, nil
		}
	}
	if names := h.Names(); len(names) > 0 {
		return nil, fmt.Errorf("未找到环境 %q;可用环境: %s", target, strings.Join(names, "、"))
	}
	return nil, fmt.Errorf("未找到环境 %q(配置里没有任何环境,运行 tt serve 添加)", target)
}

// ResolveForTool 在 Resolve 之上叠加工具级环境覆盖。目前只有 debug 有
// (debug.activeEnv);它指向的环境已被删除时,落回 hosts 的默认选择。
func (r *Root) ResolveForTool(tool, name string) (*EnvRef, error) {
	if name == "" {
		switch tool {
		case "debug":
			if r.Debug.ActiveEnv != "" {
				if ref, err := r.Hosts.Resolve(r.Debug.ActiveEnv); err == nil {
					ref.From = "tool-override"
					return ref, nil
				}
			}
		}
	}
	return r.Hosts.Resolve(name)
}
