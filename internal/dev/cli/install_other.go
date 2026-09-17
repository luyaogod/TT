//go:build !windows

package cli

import (
	"tt/internal/dev/store"
)

// 非 Windows 平台：`tt dev install path` 没有 HKCU 可写。
// 保留同样签名让 tdz 的 CLI 层跨平台可编译（go vet ./... 在别的 GOOS 下也能过）。

func userPath() (string, string, error) {
	return "", "", &store.IOError{Msg: "tt dev install path 仅支持 Windows（T100 设计器是 Windows 客户端）"}
}

func setUserPath(string, string) error {
	return &store.IOError{Msg: "tt dev install path 仅支持 Windows"}
}

func broadcastEnvChange() error { return nil }
