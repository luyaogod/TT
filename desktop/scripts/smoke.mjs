// 桌面链路冒烟（无 GUI）：验证 Electron 外壳依赖的 Go 侧契约。
//
//   1) `tt serve --desktop` 在空目录首启：建出 config.json，打印 TT_READY {json}；
//   2) READY 里的地址真实可达：/api/health 报 ok，/debug/ 返回界面 HTML；
//   3) POST /api/shutdown → 优雅退出（外壳关窗时走的就是它）。
//
// 用法：node desktop/scripts/smoke.mjs [--bin <tt.exe>]
// 退出码 0 = 全部通过；失败时打印日志尾部便于定位。
//
// 合并说明：合并前外壳拉的是 `tdebug desktop`（调试专用服务），并有一条
// "同数据目录再起一个 → attached:true" 的接管断言与 .tt-serve.json 状态文件断言。
// 统一服务 `tt serve` 是一次性的前台服务，没有接管语义、也不写状态文件
// （写状态文件的是后台守护 `tt debug serve`），这两条断言随之取消。
import { spawn } from 'node:child_process'
import fs from 'node:fs'
import http from 'node:http'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const root = path.resolve(here, '..', '..')
const binArg = process.argv.indexOf('--bin')
const BIN = binArg >= 0 ? path.resolve(process.argv[binArg + 1]) : path.join(root, 'tt.exe')

const log = (...a) => console.log('[smoke]', ...a)
const fail = (msg, extra = '') => {
  console.error('[smoke] ✗ ' + msg + (extra ? '\n' + extra : ''))
  process.exit(1)
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

function request(url, { method = 'GET', body } = {}) {
  return new Promise((resolve, reject) => {
    const u = new URL(url)
    const req = http.request(
      { hostname: u.hostname, port: u.port, path: u.pathname + u.search, method,
        headers: body ? { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(body) } : {} },
      (res) => {
        let data = ''
        res.on('data', (c) => (data += c))
        res.on('end', () => resolve({ status: res.statusCode, body: data }))
      },
    )
    req.on('error', reject)
    req.end(body)
  })
}

// 起一个桌面模式服务，返回 {child, ready, lines, exited}
function startDesktop({ dataDir, extraArgs = [] }) {
  if (!fs.existsSync(BIN)) fail(`未找到后端程序：${BIN}\n  请先在仓库根目录执行：go build -o tt.exe .`)
  const cfg = path.join(dataDir, 'config.json')
  const child = spawn(BIN, ['serve', '--desktop', '--config', cfg, '--listen', '127.0.0.1:0', ...extraArgs],
    { cwd: dataDir, windowsHide: true })
  const state = { child, ready: null, lines: [], exited: null }
  child.on('exit', (code) => { state.exited = code })
  const scan = (buf) => {
    for (const line of buf.toString('utf8').split(/\r?\n/)) {
      if (!line.trim()) continue
      state.lines.push(line)
      const i = line.indexOf('TT_READY ')
      if (i >= 0) {
        try { state.ready = JSON.parse(line.slice(i + 'TT_READY '.length)) } catch { /* 非 JSON 行忽略 */ }
      }
    }
  }
  child.stdout.on('data', scan)
  child.stderr.on('data', (b) => state.lines.push('[stderr] ' + b.toString('utf8').trim()))
  return state
}

async function waitReady(state, ms = 15000) {
  const t0 = Date.now()
  while (Date.now() - t0 < ms) {
    if (state.ready) return state.ready
    if (state.exited !== null) fail(`serve 提前退出(code=${state.exited})`, state.lines.join('\n'))
    await sleep(100)
  }
  fail('15s 内没有 TT_READY', state.lines.join('\n'))
}

async function waitExit(state, ms = 8000) {
  const t0 = Date.now()
  while (Date.now() - t0 < ms) {
    if (state.exited !== null) return state.exited
    await sleep(100)
  }
  return null
}

const dataDir = fs.mkdtempSync(path.join(os.tmpdir(), 'tt-desktop-smoke-'))
log('数据目录:', dataDir)
log('后端:', BIN)

let first = null
try {
  // ---- 1) 首启：空目录也能起来，并建出 config.json ----
  first = startDesktop({ dataDir })
  const ready = await waitReady(first)
  log('READY:', JSON.stringify(ready))
  if (!ready.url || !ready.pid) fail('READY 缺少 url/pid', JSON.stringify(ready))
  const cfg = path.join(dataDir, 'config.json')
  if (!fs.existsSync(cfg)) fail('未自动创建 config.json: ' + cfg)
  log('✓ 首启建配置并打出 READY')

  // ---- 2) READY 的地址真实可达：健康检查 + 页面 ----
  const health = JSON.parse((await request(ready.url + '/api/health')).body)
  if (!health.ok) fail('/api/health 未报 ok: ' + JSON.stringify(health))
  if (!health.debug || !health.dict) fail('/api/health 应同时报 debug 与 dict 已接入: ' + JSON.stringify(health))
  if (health.config !== cfg) fail(`健康检查报的配置路径不对: ${health.config} != ${cfg}`)
  log('✓ /api/health 可达且与 READY 同端口、同配置')

  {
    const res = await request(ready.url + '/debug/')
    if (res.status !== 200) fail(`GET /debug/ → HTTP ${res.status}`)
    if (!/<div id="root">/.test(res.body)) fail('GET /debug/ 未返回界面 HTML')
    if (!res.body.includes('TT')) fail('GET /debug/ 的页面标题不含「TT」')
  }
  log('✓ /debug/ 返回界面 HTML')

  // 根路径应跳到调试工作台
  const home = await request(ready.url + '/')
  if (home.status !== 302) fail(`GET / 应 302 到 /debug/，实际 HTTP ${home.status}`)
  log('✓ GET / 跳转到 /debug/')

  // ---- 3) 优雅停止 ----
  const sd = await request(ready.url + '/api/shutdown', { method: 'POST', body: '{}' })
  if (sd.status !== 200) fail('shutdown 应返回 200，实际 ' + sd.status)
  const code = await waitExit(first, 8000)
  if (code === null) fail('shutdown 后进程未退出', first.lines.join('\n'))
  log(`✓ /api/shutdown 优雅退出(code=${code})`)

  log('全部通过 ✓')
  fs.rmSync(dataDir, { recursive: true, force: true })
} catch (e) {
  try { if (first?.child && first.exited === null) first.child.kill() } catch {}
  fail(String(e?.stack || e), first ? first.lines.join('\n') : '')
}
