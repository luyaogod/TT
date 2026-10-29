// Package envmove 把「环境改名」里按名字落盘的东西一起搬走。
//
// 环境名是跨文件的持久键：dict/<环境>.db、ents/<环境>.json、srccache/<环境>/。
// 只改 config.json 里的 name，这些目录就变成孤儿；查询随后报「本地库不存在」，
// 而它与「数据不存在」在输出上无法区分。所以改名必须是一次迁移：
//
//	plan → 搬文件 → 改配置（config.RenameHost）→ 失败则回滚
//
// 先搬文件后改配置：配置写失败可以回滚，配置写成功却发现文件搬不动就回不去了。
package envmove

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tt/internal/config"
	"tt/internal/entdir"
)

// Move 一份要搬的落点（文件或整个目录）。
type Move struct {
	Kind string `json:"kind"` // dict / ents / srccache
	From string `json:"from"`
	To   string `json:"to"`
}

// Plan 计算 from → to 需要搬哪些落点。只读，不落盘。
//
// 源不存在就跳过（没同步过字典、没查过企业、没调试过都很正常）；
// 目标已存在则整体失败 —— 覆盖一份别人的库比报错危险得多。
func Plan(dataDir, from, to, host, zone string) ([]Move, error) {
	from, to = strings.TrimSpace(from), strings.TrimSpace(to)
	if dataDir == "" {
		return nil, fmt.Errorf("无法定位数据目录（配置路径为空）")
	}
	if from == "" || to == "" {
		return nil, fmt.Errorf("环境名不能为空")
	}
	if from == to {
		return nil, fmt.Errorf("新旧环境名相同：%q", from)
	}

	segFrom := entdir.EnvSeg(from, host, zone)
	segTo := entdir.EnvSeg(to, host, zone)
	pairs := []struct {
		kind     string
		from, to string
	}{
		{"dict", config.DictDBPath(dataDir, from), config.DictDBPath(dataDir, to)},
		{"ents", entdir.Path(dataDir, segFrom), entdir.Path(dataDir, segTo)},
		{"srccache", filepath.Join(dataDir, "srccache", segFrom), filepath.Join(dataDir, "srccache", segTo)},
	}

	out := make([]Move, 0, len(pairs))
	for _, p := range pairs {
		if p.from == "" || p.to == "" || p.from == p.to {
			continue // 环境名里只有路径非法字符被替换时，新旧指向同一份
		}
		if _, err := os.Stat(p.from); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("检查 %s 失败：%w", p.from, err)
		}
		if _, err := os.Stat(p.to); err == nil {
			return nil, fmt.Errorf("目标已存在，拒绝覆盖：%s（先确认它是不是另一个环境的数据）", p.to)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("检查 %s 失败：%w", p.to, err)
		}
		out = append(out, Move{Kind: p.kind, From: p.from, To: p.to})
	}
	return out, nil
}

// Execute 按计划搬迁。任一失败即回滚已搬的部分，不留半套。
// 企业快照搬完后顺手把文件内的 env 字段改成新名（它只用于展示，但留着旧名会误导排查）。
func Execute(moves []Move, newName string) error {
	done := make([]Move, 0, len(moves))
	for _, m := range moves {
		if err := os.Rename(m.From, m.To); err != nil {
			undo(done)
			return fmt.Errorf("搬 %s 失败：%w", m.From, err)
		}
		done = append(done, m)
		if m.Kind == "ents" && newName != "" {
			if err := SetSnapshotName(m.To, newName); err != nil {
				undo(done)
				return fmt.Errorf("更新企业快照 %s 失败：%w", m.To, err)
			}
		}
	}
	return nil
}

// Undo 反向搬（配置写失败时回滚用），并把企业快照里的 env 字段改回旧名。
func Undo(moves []Move, oldName string) {
	for i := len(moves) - 1; i >= 0; i-- {
		m := moves[i]
		if err := os.Rename(m.To, m.From); err != nil {
			continue
		}
		if m.Kind == "ents" && oldName != "" {
			_ = SetSnapshotName(m.From, oldName)
		}
	}
}

func undo(done []Move) {
	for i := len(done) - 1; i >= 0; i-- {
		_ = os.Rename(done[i].To, done[i].From)
	}
}

// SetSnapshotName 把企业快照里的 env 字段改成新环境名（新名由调用方给）。
// 迁移后调用一次：文件名对了、内容还写着旧名，排查时会被带偏。
func SetSnapshotName(snapshotPath, name string) error {
	return setSnapshotEnv(snapshotPath, name)
}

func setSnapshotEnv(snapshotPath, name string) error {
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(snapshotPath), ".json")
	}
	b, err := os.ReadFile(snapshotPath)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	m["env"] = name
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return config.AtomicWrite(snapshotPath, out)
}
