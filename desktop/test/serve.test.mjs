// 桌面壳最容易错、又最难从现象上看出来的那一段：等它就绪、判断所有权、收尾。
// 这里用假的子进程把三条路都走一遍 —— 不需要真的 Electron，也不需要真的 tt.exe。

import test from 'node:test'
import assert from 'node:assert/strict'
import { EventEmitter } from 'node:events'
import path from 'node:path'
import { PassThrough } from 'node:stream'

import {
  appUrl,
  originOf,
  parseReadyLine,
  resolveExe,
  serveArgs,
  splitLines,
  startServer,
  stopServer,
} from '../src/serve.js'

/** 一个够真的假子进程：三个 stdio 流 + exit/kill 语义。 */
function fakeChild() {
  const child = new EventEmitter()
  child.stdout = new PassThrough()
  child.stderr = new PassThrough()
  child.stdin = { destroy() {}, end() {} }
  child.exitCode = null
  child.killed = false
  child.signals = []
  child.kill = (sig) => {
    child.signals.push(sig ?? 'SIGTERM')
    child.killed = true
    child.exitCode = 0
    child.emit('exit', 0, sig ?? null)
  }
  return child
}

const READY_OWN = '{"ok":true,"attached":false,"url":"http://127.0.0.1:28670","pid":4242,"config":"C:\\\\Users\\\\u\\\\AppData\\\\Roaming\\\\tt\\\\config.json"}'
const READY_ATTACHED = '{"ok":true,"attached":true,"url":"http://127.0.0.1:28671","pid":7,"config":"c.json"}'

test('resolveExe：打包后固定取 resources/tt/tt.exe', () => {
  const got = resolveExe({
    isPackaged: true,
    resourcesPath: path.join('C:', 'Apps', 'TT Desktop', 'resources'),
    repoRoot: 'ignored',
    env: { TT_DESKTOP_EXE: 'C:\\elsewhere\\tt.exe' },
  })
  assert.equal(got, path.join('C:', 'Apps', 'TT Desktop', 'resources', 'tt', 'tt.exe'))
})

test('resolveExe：开发态优先 TT_DESKTOP_EXE，其次仓库根', () => {
  assert.equal(
    resolveExe({ isPackaged: false, repoRoot: 'D:\\repo', env: { TT_DESKTOP_EXE: 'C:\\x\\tt.exe' } }),
    'C:\\x\\tt.exe',
  )
  assert.equal(resolveExe({ isPackaged: false, repoRoot: 'D:\\repo', env: {} }), path.join('D:\\repo', 'tt.exe'))
})

test('serveArgs 是与 Go 侧对齐的契约', () => {
  const args = serveArgs()
  assert.deepEqual(args, ['serve', '--foreground', '--ready-json', '--format', 'json', '--exit-on-stdin-close'])
  // 这几项各自对应一条约定，拆开断言是为了让"谁被拿掉了"一眼看出来
  assert.ok(args.includes('--ready-json'), '没有 --ready-json 壳就读不到真实地址')
  assert.ok(args.includes('--exit-on-stdin-close'), '没有退出兜底，壳被强杀会留下隐形服务')
  assert.ok(args.includes('--foreground'), '后台模式由父进程派生子进程，壳拿到的不是真实地址')
})

test('splitLines 保留不完整的尾巴', () => {
  assert.deepEqual(splitLines('a\nb\nc'), [['a', 'b'], 'c'])
  assert.deepEqual(splitLines('done\r\n'), [['done'], ''])
  assert.deepEqual(splitLines(''), [[], ''])
})

test('parseReadyLine：正常就绪', () => {
  assert.deepEqual(parseReadyLine(READY_OWN), {
    ok: true,
    attached: false,
    url: 'http://127.0.0.1:28670',
    pid: 4242,
    config: 'C:\\Users\\u\\AppData\\Roaming\\tt\\config.json',
  })
})

test('parseReadyLine：附着（别的进程起的实例）', () => {
  const got = parseReadyLine(READY_ATTACHED)
  assert.equal(got.ok, true)
  assert.equal(got.attached, true)
  assert.equal(got.url, 'http://127.0.0.1:28671')
})

test('parseReadyLine：错误信封给出人话', () => {
  const got = parseReadyLine('{"ok":false,"error":{"code":"USAGE","message":"配置坏了"}}')
  assert.deepEqual(got, { ok: false, error: '配置坏了' })
})

test('parseReadyLine：不是就绪行的一律返回 null（日志不能被误认成契约）', () => {
  assert.equal(parseReadyLine(''), null)
  assert.equal(parseReadyLine('   '), null)
  assert.equal(parseReadyLine('[tt] 端口 28670 被占用，改用 28671'), null)
  assert.equal(parseReadyLine('{"ok":true}'), null, '缺 url 不能当成就绪')
  assert.equal(parseReadyLine('{"ok":true,"url":"ftp://x"}'), null, '只认 http(s)')
  assert.equal(parseReadyLine('{坏 JSON'), null)
})

test('appUrl 把 /debug/ 与哈希拼对', () => {
  assert.equal(appUrl('http://127.0.0.1:28670'), 'http://127.0.0.1:28670/debug/')
  assert.equal(appUrl('http://127.0.0.1:28670/'), 'http://127.0.0.1:28670/debug/')
  assert.equal(appUrl('http://127.0.0.1:28670', 'settings/sites'), 'http://127.0.0.1:28670/debug/#settings/sites')
  assert.equal(appUrl('http://127.0.0.1:28670', '#settings/sites'), 'http://127.0.0.1:28670/debug/#settings/sites')
})

test('originOf 用于同源判定', () => {
  assert.equal(originOf('http://127.0.0.1:28670/debug/'), 'http://127.0.0.1:28670')
  assert.equal(originOf('不是地址'), '')
})

test('startServer：读到就绪行 → 自持句柄', async () => {
  const child = fakeChild()
  const logs = []
  const p = startServer({
    exePath: 'tt.exe',
    env: {},
    spawn: () => child,
    timeoutMs: 2000,
    onLog: (l) => logs.push(l),
  })
  child.stdout.write('启动中…\n')
  child.stdout.write(READY_OWN + '\n')
  const handle = await p

  assert.equal(handle.url, 'http://127.0.0.1:28670')
  assert.equal(handle.pid, 4242)
  assert.equal(handle.attached, false)
  assert.equal(handle.owned, true)
  assert.deepEqual(logs, ['启动中…'], '非契约行要进日志，不能吞掉')
})

test('startServer：就绪行被拆成两次 chunk 也要认出来', async () => {
  const child = fakeChild()
  const p = startServer({ exePath: 'tt.exe', env: {}, spawn: () => child, timeoutMs: 2000 })
  const cut = Math.floor(READY_OWN.length / 2)
  child.stdout.write(READY_OWN.slice(0, cut))
  child.stdout.write(READY_OWN.slice(cut) + '\n')
  const handle = await p
  assert.equal(handle.attached, false)
})

test('startServer：attached=true 的句柄不可收尾，且它退出时不算崩溃', async () => {
  const child = fakeChild()
  let exits = 0
  const p = startServer({
    exePath: 'tt.exe',
    env: {},
    spawn: () => child,
    timeoutMs: 2000,
    onExit: () => exits++,
  })
  child.stdout.write(READY_ATTACHED + '\n')
  const handle = await p
  assert.equal(handle.attached, true)
  assert.equal(handle.owned, false)

  // 附着的那条进程打完就绪行就退（它只是来报地址的）—— 这不是崩溃
  child.exitCode = 0
  child.emit('exit', 0, null)
  await new Promise((r) => setImmediate(r))
  assert.equal(exits, 0)

  // 收尾不动别人的实例
  await stopServer(handle, { platform: 'linux', timeoutMs: 20 })
  assert.deepEqual(child.signals, [])
})

test('startServer：进程提前退出 → 报错带退出码与尾巴日志', async () => {
  const child = fakeChild()
  const p = startServer({ exePath: 'tt.exe', env: {}, spawn: () => child, timeoutMs: 2000 })
  child.stderr.write('错误: 端口全被占用\n')
  child.exitCode = 3
  child.emit('exit', 3, null)
  await assert.rejects(p, (err) => {
    assert.match(err.message, /提前退出/)
    assert.match(err.tail || '', /端口全被占用/)
    return true
  })
})

test('startServer：失败信封 → 直接透出人话，不等到超时', async () => {
  const child = fakeChild()
  const p = startServer({ exePath: 'tt.exe', env: {}, spawn: () => child, timeoutMs: 5000 })
  child.stdout.write('{"ok":false,"error":{"code":"USAGE","message":"监听地址格式不正确"}}\n')
  await assert.rejects(p, /监听地址格式不正确/)
})

test('startServer：迟迟不就绪 → 超时报错', async () => {
  const child = fakeChild()
  const p = startServer({ exePath: 'tt.exe', env: {}, spawn: () => child, timeoutMs: 30 })
  await assert.rejects(p, /没有就绪/)
  assert.equal(child.killed, true, '超时要把挂着的子进程杀掉')
})

test('startServer：spawn 本身失败（exe 不在）也要给出可读的错误', async () => {
  const p = startServer({
    exePath: 'C:\\nope\\tt.exe',
    env: {},
    spawn: () => {
      throw Object.assign(new Error('spawn ENOENT'), { code: 'ENOENT' })
    },
    timeoutMs: 1000,
  })
  await assert.rejects(p, /启动 .*tt\.exe 失败/)
})

test('stopServer：自己拉起来的，退不掉时硬杀', async () => {
  const child = fakeChild()
  const p = startServer({ exePath: 'tt.exe', env: {}, spawn: () => child, timeoutMs: 2000 })
  child.stdout.write(READY_OWN + '\n')
  const handle = await p

  // 假服务的地址连不上：fetch 会立刻失败，走"退不掉"的那条路
  await stopServer(handle, { platform: 'linux', timeoutMs: 30 })
  assert.deepEqual(child.signals, ['SIGTERM'])
})
