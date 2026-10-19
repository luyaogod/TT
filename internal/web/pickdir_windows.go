//go:build windows

package web

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// pickDirPS 用 Windows PowerShell 的 FolderBrowserDialog 弹原生目录选择框。
// 路径通过环境变量传进去，不拼进命令文本 —— 初始目录里带引号/分号也不会变成脚本。
const pickDirPS = `
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Windows.Forms
Add-Type @"
using System;
using System.Runtime.InteropServices;
public class TTPickWin {
  [DllImport("kernel32.dll")] public static extern IntPtr GetConsoleWindow();
  [DllImport("user32.dll")] public static extern bool ShowWindow(IntPtr hWnd, int nCmdShow);
}
"@
$d = New-Object System.Windows.Forms.FolderBrowserDialog
$d.Description = if ($env:TT_PICK_TITLE) { $env:TT_PICK_TITLE } else { '选择文件夹' }
$d.ShowNewFolderButton = $true
if ($env:TT_PICK_INITIAL -and (Test-Path -LiteralPath $env:TT_PICK_INITIAL -PathType Container)) {
  $d.SelectedPath = $env:TT_PICK_INITIAL
}
$console = [TTPickWin]::GetConsoleWindow()
$timer = New-Object System.Windows.Forms.Timer
$timer.Interval = 100
$timer.add_Tick({ [void][TTPickWin]::ShowWindow($console, 0); $timer.Stop() })
$timer.Start()
try {
  $r = $d.ShowDialog()
  if ($r -eq [System.Windows.Forms.DialogResult]::OK) { [Console]::Out.Write($d.SelectedPath) }
} finally {
  $d.Dispose()
}
`

func pickDirectory(ctx context.Context, initial, title string) (string, bool, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe",
		"-NoProfile", "-STA", "-Command", pickDirPS)
	cmd.Env = append(os.Environ(),
		"TT_PICK_INITIAL="+initial,
		"TT_PICK_TITLE="+title,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", false, ctx.Err()
		}
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", false, fmt.Errorf("打开文件夹选择器失败: %s", msg)
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return "", true, nil // 用户取消：PowerShell 正常退出但没有输出
	}
	return path, false, nil
}
