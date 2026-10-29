package config

import (
	"fmt"
	"strings"
)

// RenameHost 把 hosts.sshs 里的环境名 from 改成 to，并在**同一次原子写**里把
// query.source 中指向 from 的引用（`from` 或 `local:from`）一并改成 to。
// 返回 query.source 是否被改写。
//
// 为什么不是"只改 name"：环境名是跨文件的持久引用 —— query.source、dict/<环境>.db、
// ents/<环境>.json、srccache/<环境>/ 都拿它当键。只改 name 会留下悬空引用，而悬空
// 引用的症状是「查不到数据」，与「数据不存在」在输出上分不出来。按名字落盘的那几份
// 由 envmove 搬走；这个函数保证配置本身是自洽的。
func RenameHost(path, from, to string) (bool, error) {
	from, to = strings.TrimSpace(from), strings.TrimSpace(to)
	if from == "" || to == "" {
		return false, fmt.Errorf("环境名不能为空")
	}
	if from == to {
		return false, fmt.Errorf("新旧环境名相同：%q", from)
	}

	rewritten := false
	err := Edit(path, func(root map[string]any) error {
		names := hostNames(root)
		if !containsName(names, from) {
			return fmt.Errorf("没有名为 %q 的环境（用 tt env list 看全部）", from)
		}
		if containsName(names, to) {
			return fmt.Errorf("环境 %q 已存在，不能改名成它", to)
		}
		return nil
	}, func(root map[string]any) error {
		hosts, _ := root["hosts"].(map[string]any)
		if hosts == nil {
			return fmt.Errorf("配置里没有 hosts 节")
		}
		list, _ := hosts["sshs"].([]any)
		for _, item := range list {
			m, _ := item.(map[string]any)
			if m == nil {
				continue
			}
			if name, _ := m["name"].(string); strings.TrimSpace(name) == from {
				m["name"] = to
			}
		}

		q, _ := root["query"].(map[string]any)
		if q == nil {
			return nil
		}
		rewritten = RewriteQuerySource(root, from, to)
		return nil
	})
	if err != nil {
		return false, err
	}
	return rewritten, nil
}

// RewriteQuerySource 把 root 里 query.source 指向 from 的引用改成 to
// （`from` 或 `local:from`）。返回是否改写。设置页在"同一次保存里改名"时用它。
func RewriteQuerySource(root map[string]any, from, to string) bool {
	q, _ := root["query"].(map[string]any)
	if q == nil {
		return false
	}
	src, _ := q["source"].(string)
	src = strings.TrimSpace(src)
	switch {
	case src == from:
		q["source"] = to
		return true
	case strings.HasPrefix(strings.ToLower(src), "local:"):
		if env := strings.TrimSpace(src[len("local:"):]); env == from {
			q["source"] = "local:" + to
			return true
		}
	}
	return false
}

// hostNames 取 hosts.sshs 里的全部环境名（保持顺序）。
func hostNames(root map[string]any) []string {
	hosts, _ := root["hosts"].(map[string]any)
	if hosts == nil {
		return nil
	}
	list, _ := hosts["sshs"].([]any)
	out := make([]string, 0, len(list))
	for _, item := range list {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		if name, _ := m["name"].(string); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func containsName(names []string, want string) bool {
	for _, n := range names {
		if strings.TrimSpace(n) == want {
			return true
		}
	}
	return false
}

// ValidateQuerySource 校验 root 里的 query.source 只引用**存在的**环境。
//
// 这是"保存时就拦住"的那道闸：悬空引用不会当场报错，只会在下次查询时表现为
// 「本地库不存在」或「查不到数据」—— 与真实的数据缺失长得一模一样。
func ValidateQuerySource(root map[string]any) error {
	q, _ := root["query"].(map[string]any)
	if q == nil {
		return nil
	}
	src, _ := q["source"].(string)
	src = strings.TrimSpace(src)
	if src == "" || strings.EqualFold(src, "auto") || strings.EqualFold(src, "local") {
		return nil // 未固定 / 旧值：由命令行点名环境
	}
	env := src
	if strings.HasPrefix(strings.ToLower(src), "local:") {
		env = strings.TrimSpace(src[len("local:"):])
	}
	if env == "" {
		return fmt.Errorf("query.source = %q 缺少环境名", src)
	}
	names := hostNames(root)
	if !containsName(names, env) {
		return fmt.Errorf("query.source = %q 指向不存在的环境（现有：%s）", src, strings.Join(names, "、"))
	}
	return nil
}
