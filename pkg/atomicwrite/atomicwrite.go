// Package atomicwrite 原子写文件：同目录临时文件 → 写入 → fsync → 保留原权限位 →
// 重命名替换。配置、快照、插件数据都用它，避免读者看到写了一半的文件。
//
// 与 TDev internal/store/atomic.go 同一套语义，合并后只有这一份。
package atomicwrite

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Write 原子写入 path。
//
// 已存在的文件按原权限位写，新建的用 0600 —— 配置里有明文口令。
// 父目录不存在时自动创建。
func Write(path string, data []byte) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建配置目录失败(%s): %w", dir, err)
		}
	}

	mode := os.FileMode(0o600)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode().Perm()
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp*")
	if err != nil {
		if isAccessDenied(err) {
			return fmt.Errorf("创建临时文件失败(%s): %w\n"+
				"  这个目录当前不允许写入。配置由运行中的 tt serve 保存 —— "+
				"若它是从某个受限环境(沙箱/只读挂载)启动的,请换成普通终端重新启动再试。", dir, err)
		}
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
	// 用重命名替换目标。Windows 上这一步会因为**目标文件暂时被别的手占用**
	// (杀软扫描/索引器/另一个读者)而报 Access is denied —— 那是瞬时的,
	// 所以给几次短重试,而不是让用户的"保存"直接失败。
	if err := renameWithRetry(tmpName, path); err != nil {
		return fmt.Errorf("更新配置文件失败: %w", err)
	}
	return nil
}

// renameAttempts / renameBackoff 是重命名替换的重试策略。
//
// 为什么值得重试:实测(2026-10-15)在 Windows 上,只要目标文件那一刻
// 被另一个句柄占着,os.Rename 就报 `Access is denied` —— 而占用往往是杀软/索引器
// 的瞬时扫描,几十毫秒后就没了。没有重试的话,用户看到的是"删除环境失败:
// 更新配置文件失败: Access is denied",而他什么都没做错。
//
// 次数与间隔刻意小:配置保存是交互动作,不能让人干等;真被长期占用时
// 仍要如实报错(那说明有东西一直占着,重试再多次也没用)。
const (
	renameAttempts = 5
	renameBackoff  = 40 * time.Millisecond
)

func renameWithRetry(from, to string) error {
	var err error
	for i := 0; i < renameAttempts; i++ {
		if err = os.Rename(from, to); err == nil {
			return nil
		}
		if !isAccessDenied(err) {
			return err // 不是"被占用"这一类:重试没有意义
		}
		time.Sleep(renameBackoff * time.Duration(i+1)) // 递增退避
	}
	return err
}

// isAccessDenied 判断错误是不是"拒绝访问"这一类。
//
// 基础判据只用标准库(errors.Is(err, fs.ErrPermission))—— 它跨平台成立,
// 而 Go 在 Windows 上会把 ERROR_ACCESS_DENIED 映射到它。平台特有的补充判据
// (syscall.Errno 直出的情况)放在 isAccessDeniedPlatform 里,按平台分文件。
func isAccessDenied(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, fs.ErrPermission) {
		return true
	}
	return isAccessDeniedPlatform(err)
}
