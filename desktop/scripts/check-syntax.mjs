// 语法闸门。
//
// main.js / preload.cjs 没法用 `node --test` 直接跑（它们 import 'electron'，在纯 Node 里
// 拿到的不是 Electron 的 API），但一个语法错误就足以让桌面版根本起不来 —— 而它的表现是
// "双击没反应"，没有任何日志。`node --check` 能便宜地把这类错误挡在打包之前。
//
// 这不是替代真机冒烟：窗口、菜单、退出收尾仍然要在真桌面里跑一遍（见 desktop/README.md）。

import { execFileSync } from 'node:child_process'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(here, '..')

const files = [
  'src/main.js',
  'src/serve.js',
  'src/preload.cjs',
  'scripts/dev.mjs',
  'scripts/check-syntax.mjs',
]

let bad = 0
for (const rel of files) {
  const file = path.join(root, rel)
  try {
    execFileSync(process.execPath, ['--check', file], { stdio: 'pipe' })
  } catch (err) {
    bad++
    console.error(`✖ ${rel}`)
    console.error(String(err.stderr || err.message))
  }
}
if (bad > 0) process.exit(1)
console.log(`语法检查通过（${files.length} 个文件）`)
