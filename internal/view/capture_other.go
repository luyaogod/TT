//go:build !windows

package view

import (
	"errors"
	"image"
	"time"
)

// 抓图后端只有 Windows 实现：目标平台就是 Windows 上的 GDC。
func List(pid uint32) ([]Window, error) {
	return nil, errors.New("view: 窗口枚举只在 Windows 上实现")
}

func ListTimeout(pid uint32, d time.Duration) ([]Window, error) {
	return nil, errors.New("view: 窗口枚举只在 Windows 上实现")
}

func Capture(hwnd uintptr) (*image.RGBA, error) {
	return nil, errors.New("view: 抓图只在 Windows 上实现")
}
