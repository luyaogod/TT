// 开发态启动器：`npm run dev`。
//
// 它只做两件容易忘的事：确认 tt.exe 已经构建好；把 TT_WEB_DIR 之外的活交给壳。
// 前端热更新走 TT_WEB_DIR（见 main.go 的注释）：把 web/dist 指给服务，改完前端
// 重建 dist、窗口里按 Ctrl+R 就能看到，不用重启服务（重启会清掉内存里的调试会话）。

import { spawn } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const desktopDir = path.resolve(here, '..')
const repoRoot = path.resolve(desktopDir, '..')

const exe = (process.env.TT_DESKTOP_EXE || '').trim() || path.join(repoRoot, 'tt.exe')
if (!fs.existsSync(exe)) {
  console.error(`没有找到 tt.exe：${exe}\n先在仓库根运行 make build`)
  process.exit(1)
}

// 前端产物：仓库里通常已经构建过；没构建就提示一句（不自动构建，免得每次启动都等）
const dist = path.join(repoRoot, 'web', 'dist')
const env = { ...process.env, TT_DESKTOP_EXE: exe }
if (fs.existsSync(path.join(dist, 'index.html'))) {
  env.TT_WEB_DIR = env.TT_WEB_DIR || dist
} else {
  console.warn(`[dev] 没有 ${path.join(dist, 'index.html')}：界面会是一张引导页。先在 web/ 里 npm run build`)
}

const electronPath = (await import('electron')).default
const child = spawn(electronPath, ['.'], { cwd: desktopDir, env, stdio: 'inherit' })
child.on('exit', (code) => process.exit(code ?? 0))
