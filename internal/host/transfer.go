package host

// 远端文件的**读写**能力：与 fsops.go（只读原语）分开放，写意图是显式的。
//
// 用途：插件把资源文件（例如流程图节点图片）在环境之间搬运。写只发生在插件
// 明确调用的路径上；`tt file` 那类只读命令仍然只经 host.FS。

import (
	"errors"
	"fmt"
	"io"
	"path"
)

// ErrFileTooLarge 远端文件超过调用方给的上限。
var ErrFileTooLarge = errors.New("远端文件超过大小上限")

// GetFile 读一个远端文件（SFTP）。maxBytes<=0 表示不限；超限返回 ErrFileTooLarge。
func (c *SSHConn) GetFile(remote string, maxBytes int64) ([]byte, error) {
	cl, err := c.SFTP()
	if err != nil {
		return nil, fmt.Errorf("打开 SFTP 通道失败: %w", err)
	}
	defer cl.Close()
	f, err := cl.Open(remote)
	if err != nil {
		return nil, fmt.Errorf("打开远端文件失败(%s): %w", remote, err)
	}
	defer f.Close()
	if maxBytes > 0 {
		b, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
		if err != nil {
			return nil, fmt.Errorf("读取远端文件失败(%s): %w", remote, err)
		}
		if int64(len(b)) > maxBytes {
			return nil, ErrFileTooLarge
		}
		return b, nil
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("读取远端文件失败(%s): %w", remote, err)
	}
	return b, nil
}

// PutFile 写一个远端文件（SFTP），父目录不存在时自动创建。
func (c *SSHConn) PutFile(remote string, data []byte) error {
	cl, err := c.SFTP()
	if err != nil {
		return fmt.Errorf("打开 SFTP 通道失败: %w", err)
	}
	defer cl.Close()
	if dir := path.Dir(remote); dir != "" && dir != "." && dir != "/" {
		if err := cl.MkdirAll(dir); err != nil {
			return fmt.Errorf("创建远端目录失败(%s): %w", dir, err)
		}
	}
	f, err := cl.Create(remote)
	if err != nil {
		return fmt.Errorf("创建远端文件失败(%s): %w", remote, err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("写入远端文件失败(%s): %w", remote, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("关闭远端文件失败(%s): %w", remote, err)
	}
	return nil
}
