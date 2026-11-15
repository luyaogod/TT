// 桌面壳的第三方页面回归：这是**静态断言**，不是真机 Electron 测试。
//
// 为什么用静态断言：这个仓库的 desktop/test 全部是纯 Node（不启 Electron），
// 而起一个 Electron 实例在 CI/本机都要下载运行时。插件页面的安全属性恰恰全在
// 配置里（BrowserWindow 的三个开关 + webview 禁用 + 窗口打开策略 + iframe sandbox），
// 这些一旦被谁"顺手放宽"，现象是**看不出来**的：页面照样能用，隔离没了。
// 所以这里把它们钉成"改了就红"的断言；真机验证（iframe 真跑起来后的 postMessage
// 语义）仍需人跑一次，已在 .agents/docs/plugin-system/12-user-guide.md 的验收清单里注明。

import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(here, '..', '..')
const mainSrc = fs.readFileSync(path.join(repoRoot, 'desktop', 'src', 'main.js'), 'utf8')
const pluginViewSrc = fs.readFileSync(
  path.join(repoRoot, 'web', 'app', 'src', 'PluginView.tsx'),
  'utf8',
)

// browserWindows 把 main.js 里所有 new BrowserWindow({...}) 的选项文本切出来。
//
// 只做文本切分而不求值：main.js 依赖 electron 运行时，import 它会直接失败；
// 而这里要守的恰好是"源码里写着什么"。
function browserWindows(src) {
  const out = []
  const marker = 'new BrowserWindow({'
  let i = src.indexOf(marker)
  while (i >= 0) {
    const start = i + marker.length
    let depth = 1
    let j = start
    for (; j < src.length && depth > 0; j++) {
      if (src[j] === '{') depth++
      else if (src[j] === '}') depth--
    }
    out.push(src.slice(start, j - 1))
    i = src.indexOf(marker, j)
  }
  return out
}

test('每个 BrowserWindow 都关掉 Node 集成、开上下文隔离与沙箱', () => {
  const wins = browserWindows(mainSrc)
  assert.ok(wins.length >= 2, `该找到主窗口与标题栏窗口，实际 ${wins.length}`)
  for (const [idx, opts] of wins.entries()) {
    assert.match(opts, /nodeIntegration:\s*false/, `第 ${idx + 1} 个窗口 nodeIntegration 必须 false`)
    assert.match(opts, /contextIsolation:\s*true/, `第 ${idx + 1} 个窗口 contextIsolation 必须 true`)
    assert.match(opts, /sandbox:\s*true/, `第 ${idx + 1} 个窗口 sandbox 必须 true`)
  }
})

test('preload 只有两个白名单文件，且都不给第三方页面用', () => {
  const wins = browserWindows(mainSrc)
  const pres = wins.map((o) => (o.match(/preload:\s*path\.join\([^)]*'([^']+\.cjs)'\)/) || [])[1])
  assert.deepEqual(
    pres.filter(Boolean).sort(),
    ['preload.cjs', 'shell-preload.cjs'],
    `preload 只允许这两个，实际 ${JSON.stringify(pres)}`,
  )
  // 插件页面跑在 iframe 里：它既不是 BrowserWindow，也拿不到任何 preload ——
  // 这条断言守的是"别哪天给插件加一个专用窗口/专用 preload"。
  assert.doesNotMatch(mainSrc, /plugin.*preload|preload.*plugin/i)
})

test('webview 标签禁用、webview 附加被拒、窗口打开被拦', () => {
  const wins = browserWindows(mainSrc)
  for (const [idx, opts] of wins.entries()) {
    assert.doesNotMatch(opts, /webviewTag:\s*true/, `第 ${idx + 1} 个窗口不该开 webviewTag`)
  }
  assert.match(mainSrc, /webviewTag:\s*false/, 'webviewTag 必须显式 false')
  assert.match(mainSrc, /will-attach-webview',\s*\(e\)\s*=>\s*e\.preventDefault\(\)/, 'webview 附加必须被拒')
  assert.match(mainSrc, /setWindowOpenHandler\(/, '必须有窗口打开策略')
})

test('插件 iframe 的 sandbox 只有 allow-scripts（不给同源）', () => {
  const m = pluginViewSrc.match(/sandbox="([^"]*)"/)
  assert.ok(m, 'PluginView 里该有 iframe sandbox 属性')
  const perms = m[1].split(/\s+/).filter(Boolean)
  assert.deepEqual(perms, ['allow-scripts'], `sandbox = ${JSON.stringify(perms)}`)
  for (const bad of ['allow-same-origin', 'allow-top-navigation', 'allow-popups', 'allow-forms', 'allow-modals']) {
    assert.ok(!perms.includes(bad), `sandbox 不该含 ${bad}`)
  }
})
