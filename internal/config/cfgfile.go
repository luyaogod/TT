package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Open 读取并解析配置文件；文件不存在返回空配置（不是错误 —— 首次运行没有文件），
// 空文件视为空配置，非法 JSON 报错。
//
// 返回的 map 保留所有未知顶层键：调用方只该改自己那一节，其余原样带走。
func Open(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("读取配置文件失败(%s): %w", path, err)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return map[string]any{}, nil // 空文件（或只有空白）→ 空配置
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("解析配置文件失败(%s): %v", path, err)
	}
	if root == nil {
		root = map[string]any{} // 字面量 null → 空配置
	}
	return root, nil
}

// Save 原子写回：MarshalIndent + 同目录临时文件 + fsync + 重命名，
// 失败不留半截文件，也不会让读者看到只写了一半的内容。
//
// 配置目录不存在时自动创建 —— 缺省位置在 %APPDATA%\T100\tt\ 下，
// 新机器上首次保存时该目录还没有。
func Save(path string, root map[string]any) error {
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	out = append(out, '\n')
	return AtomicWrite(path, out)
}

// AtomicWrite 同目录临时文件 → 写入 → fsync → 保留原权限位 → 重命名替换。
// 与 TDev internal/store/atomic.go 同一套语义，合并后只有这一份。
func AtomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建配置目录失败(%s): %w", dir, err)
		}
	}

	// 已存在的文件按原权限位写，新建的用 0600 —— 配置里有明文口令。
	mode := os.FileMode(0o600)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode().Perm()
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp*")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // 成功路径下已被 Rename 掉，这里是失败清理

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("同步配置文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭配置文件失败: %w", err)
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return fmt.Errorf("设置配置文件权限失败: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("更新配置文件失败: %w", err)
	}
	return nil
}

// Edit 统一修改入口：Open → validate(root)（可空）→ mutate(root) → Save。
// validate/mutate 任一失败都不落盘。
//
// config.json 的一切修改都应走本函数，避免各调用方重复"读-改-写"样板
// 与错误语义漂移。注意 mutate 只改自己那一节，Open 保留的其他顶层键会原样写回。
func Edit(path string, validate func(root map[string]any) error, mutate func(root map[string]any) error) error {
	root, err := Open(path)
	if err != nil {
		return err
	}
	if validate != nil {
		if err := validate(root); err != nil {
			return err
		}
	}
	if mutate != nil {
		if err := mutate(root); err != nil {
			return err
		}
	}
	return Save(path, root)
}

// NewSkeleton 返回一份最小可用的新配置：空环境清单 + 缺省监听地址。
// 首次运行（以及桌面版首启）用它落一个骨架，用户随后在配置页里填。
func NewSkeleton() map[string]any {
	return map[string]any{
		"schemaVersion": SchemaVersion,
		"listen":        DefaultListen,
		"hosts": map[string]any{
			"activeEnv": "",
			"sshs":      []any{},
		},
		"query": map[string]any{"source": "auto"},
	}
}

// EnsureExists 在配置不存在时落一份骨架（已存在则什么都不做）。
//
// 服务启动路径需要它：桌面/首次运行时还没有任何环境，但配置页要有文件可编辑，
// 而 Electron 外壳也会检查数据目录里是否建出了 config.json。
// 沿用与 Save 相同的原子写，所以不会留下半截文件。
func EnsureExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return Save(path, NewSkeleton())
}

// Section 返回 root 的顶层对象节；缺失或非对象返回统一错误。
// key 为节名，如 "hosts" / "debug"。
func Section(root map[string]any, key string) (map[string]any, error) {
	s, _ := root[key].(map[string]any)
	if s == nil {
		return nil, fmt.Errorf("config.json 缺少 %q 配置节", key)
	}
	return s, nil
}

// SectionOrEmpty 同 Section，但缺失时返回一个已挂回 root 的空节。
// 写路径用它，读路径用 Section。
func SectionOrEmpty(root map[string]any, key string) map[string]any {
	if s, ok := root[key].(map[string]any); ok && s != nil {
		return s
	}
	s := map[string]any{}
	root[key] = s
	return s
}
