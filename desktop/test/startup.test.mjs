// 启动顺序的回归：这是**静态断言**，不是真机 Electron 测试。
//
// 这个顺序在纯 Node 里跑不起来（main.js 依赖 electron 运行时），但它一旦回退，
// 用户看到的就是"每次启动先弹一张加载小窗" —— 这是已经被明确去掉的行为。
// 所以按源码把两头钉住：boot() 在服务就绪前不许建窗口；启动页那套东西不许回来。

import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const srcDir = path.resolve(here, '..', 'src')
const mainSrc = fs.readFileSync(path.join(srcDir, 'main.js'), 'utf8')

// 取出一个函数的函数体（按花括号配平）。只够解析 main.js 里这几处，不求通用。
function functionBody(src, name) {
  const m = new RegExp(`(?:async\\s+)?function\\s+${name}\\s*\\(`).exec(src)
  assert.ok(m, `main.js 里找不到 function ${name}`)
  const open = src.indexOf('{', m.index + m[0].length)
  assert.ok(open >= 0, `${name} 后面没有函数体`)
  let depth = 0
  for (let i = open; i < src.length; i++) {
    if (src[i] === '{') depth++
    else if (src[i] === '}') {
      depth--
      if (depth === 0) return src.slice(open + 1, i)
    }
  }
  assert.fail(`${name} 的花括号没有配平`)
}

test('boot：服务就绪前不建任何窗口，就绪后直接开主窗', () => {
  const body = functionBody(mainSrc, 'boot')
  const ready = body.indexOf('startServer(')
  assert.ok(ready >= 0, 'boot 里应该调用 startServer')

  const before = body.slice(0, ready)
  assert.doesNotMatch(before, /new BrowserWindow/, '服务就绪前不许建窗口')
  assert.doesNotMatch(before, /showFailurePage|openMain/, '服务就绪前不许开页面')

  const after = body.slice(ready)
  assert.match(after, /openMain\(\s*''\s*\)/, '服务就绪后直接开主窗')
})

test('启动页整套已移除：只剩失败页，且只在出错的分支出现', () => {
  assert.equal(fs.existsSync(path.join(srcDir, 'splash.html')), false, 'splash.html 应该已经删掉')
  assert.doesNotMatch(mainSrc, /splash/i, 'main.js 里不该再有启动页的痕迹')
  assert.match(mainSrc, /failure\.html/, '失败页仍是独立页面')
})
