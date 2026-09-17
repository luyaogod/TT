package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"tt/internal/cli"
)

// webDist 前端构建产物（web/dist）。
//
// 两套 SPA 分别构建到 web/dist/debug 与 web/dist/dict，由 tt serve 挂在
// /debug/ 与 /dict/ 下。未构建时目录由 web/dist/.gitkeep 占位，go build 仍能通过，
// 服务返回引导页告诉用户去构建 —— 这样后端开发不被前端依赖卡住。
//
// AI 技能文件不内嵌：以仓库根目录的 skills/ 随发行包（便携版）一起分发，
// 用户可直接编辑；`tt install skills` 把它复制到目标目录。
// 与合并前的 TDebug / TDev / TDictCli 三边约定一致。
//
//go:embed all:web/dist
var webDist embed.FS

func main() {
	// 把 FS 的根收敛到 dist 目录本身。embed 的根是模块路径（web/dist/...），
	// 而所有消费者（internal/web 的 SubFS、调试服务的静态兜底）都按
	// "dist 根"理解路径 —— 在这里剥一层，免得每个调用方各自去猜该不该剥。
	dist, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "内嵌前端目录异常(web/dist): %v\n", err)
		os.Exit(1)
	}
	cli.Execute(dist)
}
