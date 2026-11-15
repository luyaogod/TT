// 插件前端贡献的检查项（纯 Node，不连后端、不需要浏览器）：
//   1. #plugin/<route> 的解析与 hash 回写；未知/停用必须落到显式错误页；
//   2. 菜单顺序（按标题，稳定）与“只列已启用”；
//   3. 管理面拒绝清单与两条受控桥路径；
//   4. 消息桥的授权：nonce、来源、未知类型、重放、跨插件、已销毁；
//   5. 体量上限：> 1 MiB 或 > 5000 行都必须显式回 error，不静默缓冲；
//   6. iframe 的 sandbox 只给 allow-scripts。
//
// 与其它 check:* 一样：由 esbuild 打成一个临时脚本再交给 Node 跑，
// 所以被加载的模块（routing / plugins / pluginBridge / pluginProtocol）顶层不能碰 DOM。
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

import { isPluginRoute, parseRoute, routeHash, routeView } from '../src/routing'
import {
  DENIED_ADMIN_PATHS,
  isDeniedAdminPath,
  normalizeHostApiMethod,
  normalizePluginRpcPath,
  normalizePlugins,
  pluginHostApiBridgePath,
  pluginMenuItems,
  pluginRPCBridgePath,
  pluginUIPath,
  resolvePluginPage,
} from '../src/plugins'
import { PluginBridge } from '../src/pluginBridge'
import {
  BRIDGE_PROTOCOL,
  BRIDGE_VERSION,
  MAX_MESSAGE_BYTES,
  MAX_RESPONSE_ROWS,
} from '../src/pluginProtocol'

// ---------------------------------------------------------------------------
// 1) 路由
// ---------------------------------------------------------------------------
{
  assert.deepEqual(parseRoute('#plugin/demo'), { kind: 'plugin', route: 'demo' })
  assert.deepEqual(parseRoute('#plugin/demo?from=deep'), { kind: 'plugin', route: 'demo' })
  assert.deepEqual(parseRoute('#plugin/demo/'), { kind: 'plugin', route: 'demo' })
  assert.equal(parseRoute('#plugin/Demo').kind, 'plugin-error', '大写必须拒')
  assert.equal(parseRoute('#plugin/').kind, 'plugin-error', '空 route 必须拒')
  assert.equal(parseRoute('#plugin/demo/extra').kind, 'plugin-error', '多余的段必须拒')
  assert.equal(parseRoute('#pluginx/demo').kind, 'home', '不是插件路由')
  assert.equal(isPluginRoute('demo-page'), true)
  assert.equal(isPluginRoute('demo_page'), false)
  assert.equal(routeHash({ kind: 'plugin', route: 'demo' }, 'sites'), '#plugin/demo')
  assert.equal(routeView({ kind: 'plugin', route: 'demo' }), 'plugin')
  assert.equal(routeView({ kind: 'plugin-error', route: 'x', reason: 'malformed' }), 'plugin')
  console.log('OK 1) #plugin/<route> 解析、回写与失败形态')
}

// ---------------------------------------------------------------------------
// 2) 未知 / 停用 → 显式错误，不回退
// ---------------------------------------------------------------------------
{
  const plugins = normalizePlugins({
    ok: true,
    plugins: [
      { id: 'com.example.on', version: '1.2.3', enabled: true, route: 'on', routeTitle: '可用页面' },
      { id: 'com.example.off', version: '0.1.0', enabled: false, route: 'off', routeTitle: '停用页面' },
      { id: 'com.example.cli', version: '1.0.0', enabled: true, namespace: 'x' }, // 没有页面贡献
    ],
  })
  assert.equal(plugins.length, 2, '没有 route 的插件不进前端描述符')
  const ok = resolvePluginPage('on', plugins)
  assert.equal(ok.kind, 'ok')
  assert.equal(ok.plugin.title, '可用页面')

  const unknown = resolvePluginPage('nope', plugins)
  assert.equal(unknown.kind, 'error')
  assert.equal(unknown.reason, 'unknown')

  const disabled = resolvePluginPage('off', plugins)
  assert.equal(disabled.kind, 'error')
  assert.equal(disabled.reason, 'disabled')
  assert.equal(disabled.plugin.route, 'off', '错误页要能说出是哪个插件')
  console.log('OK 2) 未知 / 停用各自落显式错误，不静默回退')
}

// ---------------------------------------------------------------------------
// 3) 菜单顺序与过滤
// ---------------------------------------------------------------------------
{
  const plugins = normalizePlugins([
    { id: 'com.example.beta', version: '1', enabled: true, route: 'beta', routeTitle: 'Beta' },
    { id: 'com.example.alpha', version: '1', enabled: true, route: 'alpha', routeTitle: 'Alpha' },
    { id: 'com.example.off', version: '1', enabled: false, route: 'off', routeTitle: 'Alpha' },
  ])
  assert.deepEqual(pluginMenuItems(plugins).map((p) => p.route), ['alpha', 'beta'])
  assert.deepEqual(pluginMenuItems([...plugins].reverse()).map((p) => p.route), ['alpha', 'beta'],
    '输入顺序变化不影响菜单顺序')

  // 同名时按 route 兜底，仍然是确定的
  const tie = normalizePlugins([
    { id: 'com.example.two', version: '1', enabled: true, route: 'zzz', routeTitle: '同名' },
    { id: 'com.example.one', version: '1', enabled: true, route: 'aaa', routeTitle: '同名' },
  ])
  assert.deepEqual(pluginMenuItems(tie).map((p) => p.route), ['aaa', 'zzz'])
  console.log('OK 3) 菜单只列已启用，按标题稳定排序')
}

// ---------------------------------------------------------------------------
// 4) 管理面拒绝清单 / 受控桥路径
// ---------------------------------------------------------------------------
{
  for (const p of DENIED_ADMIN_PATHS) {
    assert.equal(isDeniedAdminPath(p), true, `应拒绝 ${p}`)
    assert.equal(isDeniedAdminPath(p + '/sub'), true, `应拒绝 ${p}/sub`)
  }
  assert.equal(isDeniedAdminPath('/api/status'), true, '/api 整棵树都不给桥用')
  assert.equal(isDeniedAdminPath('/api/hostsfoo'), true, '/api 前缀本身就不允许')
  assert.equal(isDeniedAdminPath('/apix/hosts'), false, '不是 /api 段')
  assert.equal(isDeniedAdminPath('/plugins/com.example.ui/ui/'), false)

  const host = pluginHostApiBridgePath('com.example.ui', 'tt.host.v1.ProfilesService/List')
  assert.equal(host, '/plugins/com.example.ui/hostrpc/tt.host.v1.ProfilesService/List')
  assert.ok(!host.includes('/api/'), 'Host API 不走 /api/')
  assert.equal(pluginHostApiBridgePath('com.example.ui', '/api/hosts'), null)
  assert.equal(pluginHostApiBridgePath('com.example.ui', 'tt.host.v1.ProfilesService/../../api'), null)
  assert.equal(normalizeHostApiMethod('tt.host.v1.DBSessionService/Query'), 'tt.host.v1.DBSessionService/Query')
  assert.equal(normalizeHostApiMethod('tt.host.v1.EvilService/Query'), null)

  assert.equal(pluginRPCBridgePath('com.example.ui', '/echo'), '/plugins/com.example.ui/rpc/echo')
  assert.equal(pluginRPCBridgePath('com.example.ui', '/api/plugins'), null, '管理面路径不代理')
  assert.equal(pluginRPCBridgePath('com.example.ui', '/../api/hosts'), null, '路径穿越必须拒')
  assert.equal(pluginRPCBridgePath('com.example.ui', '//evil.test/x'), null, '协议相对地址必须拒')
  assert.equal(normalizePluginRpcPath('/v1/echo'), '/v1/echo')

  assert.equal(pluginUIPath('com.example.ui'), '/plugins/com.example.ui/ui/')
  assert.equal(pluginUIPath('../evil'), null)
  console.log('OK 4) 拒绝清单覆盖 13 条管理端点，两条桥路径都不经过 /api/')
}

// ---------------------------------------------------------------------------
// 消息桥测试脚手架
// ---------------------------------------------------------------------------
let seq = 0

function fakeFrame() {
  const posted = []
  const win = {
    postMessage(message, targetOrigin) {
      posted.push({ message, targetOrigin })
    },
  }
  return { frame: { contentWindow: win }, win, posted }
}

function envelope(init, type, extra = {}) {
  seq += 1
  return {
    protocol: BRIDGE_PROTOCOL,
    v: BRIDGE_VERSION,
    type,
    id: `msg-${seq}`,
    session: init.session,
    nonce: init.nonce,
    ts: Date.now(),
    ...extra,
  }
}

function hostRequest(init, extra = {}) {
  return envelope(init, 'host.request', {
    reqId: `host-${seq}`,
    method: 'tt.host.v1.ProfilesService/List',
    payload: {},
    ...extra,
  })
}

function rpcRequest(init, extra = {}) {
  return envelope(init, 'rpc.request', {
    reqId: `rpc-${seq}`,
    path: '/echo',
    payload: {},
    ...extra,
  })
}

function makeBridge(transport = {}) {
  const calls = { host: [], rpc: [] }
  const errors = []
  const full = {
    hostApi: async (call) => {
      calls.host.push(call)
      return transport.hostApi ? transport.hostApi(call) : { ok: true }
    },
    pluginRPC: async (call) => {
      calls.rpc.push(call)
      return transport.pluginRPC ? transport.pluginRPC(call) : { ok: true }
    },
  }
  const bridge = new PluginBridge({
    pluginId: 'com.example.ui',
    route: 'demo',
    transport: full,
    onError: (e) => errors.push(e),
  })
  const fake = fakeFrame()
  assert.equal(bridge.attach(fake.frame), true, 'attach 应成功')
  const init = fake.posted[0].message
  assert.equal(init.type, 'init')
  assert.equal(init.pluginId, 'com.example.ui', 'init 绑定的是父页面认定的 id')
  assert.equal(fake.posted[0].targetOrigin, '*', '不透明源只能用 targetOrigin="*"')
  assert.ok(init.nonce && init.nonce.length >= 32, 'init 必须带随机 nonce')
  return { bridge, calls, errors, ...fake, init }
}

// ---------------------------------------------------------------------------
// 5) nonce 不匹配：静默拒绝，不触达网络
// ---------------------------------------------------------------------------
{
  const ctx = makeBridge()
  const req = hostRequest(ctx.init)
  req.nonce = 'ffffffffffffffff'
  const d = ctx.bridge.handleMessage({ source: ctx.win, data: req })
  assert.equal(d.accepted, false)
  assert.equal(d.code, 'UNAUTHORIZED')
  assert.equal(ctx.calls.host.length, 0, 'nonce 不匹配不得发起 Host API 调用')
  assert.equal(ctx.errors.length, 0, '未认证消息不往界面上画')
  ctx.bridge.destroy()
  console.log('OK 5) nonce 不匹配被拒，且不触发网络/界面')
}

// ---------------------------------------------------------------------------
// 6) 伪造 event.source：别的窗口不能借这个 iframe 的身份
// ---------------------------------------------------------------------------
{
  const ctx = makeBridge()
  const d = ctx.bridge.handleMessage({ source: { other: true }, data: hostRequest(ctx.init) })
  assert.equal(d.accepted, false)
  assert.equal(d.code, 'UNAUTHORIZED')
  assert.equal(ctx.calls.host.length, 0)
  ctx.bridge.destroy()
  console.log('OK 6) event.source 伪造被拒')
}

// ---------------------------------------------------------------------------
// 7) 未知消息类型：丢弃 + 可见错误 + 回 error
// ---------------------------------------------------------------------------
{
  const ctx = makeBridge()
  const d = ctx.bridge.handleMessage({ source: ctx.win, data: envelope(ctx.init, 'host.evil') })
  assert.equal(d.accepted, false)
  assert.equal(d.code, 'UNKNOWN_TYPE')
  assert.equal(ctx.errors.length, 1)
  assert.equal(ctx.errors[0].code, 'UNKNOWN_TYPE')
  const reply = ctx.posted.find((p) => p.message.type === 'error')
  assert.ok(reply, '应向 iframe 回一条 error')
  assert.equal(reply.message.error.code, 'UNKNOWN_TYPE')
  ctx.bridge.destroy()
  console.log('OK 7) 未知类型丢弃并可见报错')
}

// ---------------------------------------------------------------------------
// 8) 重放：同一条消息 id 第二次必须拒
// ---------------------------------------------------------------------------
{
  const ctx = makeBridge()
  const req = hostRequest(ctx.init)
  assert.equal(ctx.bridge.handleMessage({ source: ctx.win, data: req }).accepted, true)
  await ctx.bridge.flush()
  const again = ctx.bridge.handleMessage({ source: ctx.win, data: req })
  assert.equal(again.accepted, false)
  assert.equal(again.code, 'REPLAY')
  assert.equal(ctx.calls.host.length, 1, '重放不得触发第二次调用')
  ctx.bridge.destroy()
  console.log('OK 8) 重放消息被拒')
}

// ---------------------------------------------------------------------------
// 9) 跨插件：消息里自称的 pluginId 不参与授权
// ---------------------------------------------------------------------------
{
  const ctx = makeBridge()
  const req = hostRequest(ctx.init, { pluginId: 'com.evil.plugin' })
  assert.equal(ctx.bridge.handleMessage({ source: ctx.win, data: req }).accepted, true)
  await ctx.bridge.flush()
  assert.equal(ctx.calls.host.length, 1)
  assert.equal(ctx.calls.host[0].pluginId, 'com.example.ui', '调用只认父页面绑定的 id')
  assert.equal(ctx.calls.host[0].method, 'tt.host.v1.ProfilesService/List')
  ctx.bridge.destroy()
  console.log('OK 9) iframe 自报的 pluginId 被忽略，调用绑定真实 id')
}

// ---------------------------------------------------------------------------
// 10) 入站消息 > 1 MiB：显式拒绝
// ---------------------------------------------------------------------------
{
  const ctx = makeBridge()
  const req = hostRequest(ctx.init, { payload: { blob: 'x'.repeat(MAX_MESSAGE_BYTES + 32) } })
  const d = ctx.bridge.handleMessage({ source: ctx.win, data: req })
  assert.equal(d.accepted, false)
  assert.equal(d.code, 'PAYLOAD_TOO_LARGE')
  assert.equal(ctx.calls.host.length, 0, '超限消息不得触达网络')
  const reply = ctx.posted.find((p) => p.message.type === 'error')
  assert.ok(reply)
  assert.equal(reply.message.error.code, 'PAYLOAD_TOO_LARGE')
  ctx.bridge.destroy()
  console.log('OK 10) 大于 1 MiB 的消息被拒并显式回报')
}

// ---------------------------------------------------------------------------
// 11) 响应 > 5000 行：显式 error，不发 response
// ---------------------------------------------------------------------------
{
  const rows = Array.from({ length: MAX_RESPONSE_ROWS + 1 }, (_, i) => ({ values: [String(i)] }))
  const ctx = makeBridge({ hostApi: async () => ({ columns: ['n'], rows, rowCount: rows.length }) })
  assert.equal(ctx.bridge.handleMessage({ source: ctx.win, data: hostRequest(ctx.init) }).accepted, true)
  await ctx.bridge.flush()
  assert.equal(ctx.calls.host.length, 1)
  assert.ok(!ctx.posted.some((p) => p.message.type === 'host.response'), '超行数不得发 response')
  const reply = ctx.posted.find((p) => p.message.type === 'error')
  assert.ok(reply)
  assert.equal(reply.message.error.code, 'TOO_MANY_ROWS')
  assert.equal(ctx.errors.at(-1).code, 'TOO_MANY_ROWS')
  ctx.bridge.destroy()
  console.log('OK 11) 超过 5000 行的响应被拒，不缓冲后照发')
}

// ---------------------------------------------------------------------------
// 12) 响应 > 1 MiB：显式 error
// ---------------------------------------------------------------------------
{
  const ctx = makeBridge({ hostApi: async () => ({ ok: true, blob: 'x'.repeat(MAX_MESSAGE_BYTES + 1) }) })
  assert.equal(ctx.bridge.handleMessage({ source: ctx.win, data: hostRequest(ctx.init) }).accepted, true)
  await ctx.bridge.flush()
  assert.ok(!ctx.posted.some((p) => p.message.type === 'host.response'))
  const reply = ctx.posted.find((p) => p.message.type === 'error')
  assert.ok(reply)
  assert.equal(reply.message.error.code, 'PAYLOAD_TOO_LARGE')
  ctx.bridge.destroy()
  console.log('OK 12) 大于 1 MiB 的响应被拒')
}

// ---------------------------------------------------------------------------
// 13) cancel：中断在途请求，回 CANCELLED
// ---------------------------------------------------------------------------
{
  let started = false
  const ctx = makeBridge({
    hostApi: (call) => new Promise((_, reject) => {
      started = true
      call.signal.addEventListener('abort', () => reject(new Error('aborted')))
    }),
  })
  const req = hostRequest(ctx.init)
  assert.equal(ctx.bridge.handleMessage({ source: ctx.win, data: req }).accepted, true)
  assert.equal(started, true)
  const cancel = envelope(ctx.init, 'cancel', { reqId: req.reqId })
  assert.equal(ctx.bridge.handleMessage({ source: ctx.win, data: cancel }).accepted, true)
  await ctx.bridge.flush()
  const reply = ctx.posted.find((p) => p.message.type === 'error')
  assert.ok(reply)
  assert.equal(reply.message.error.code, 'CANCELLED')
  ctx.bridge.destroy()
  console.log('OK 13) cancel 中断在途请求并回 CANCELLED')
}

// ---------------------------------------------------------------------------
// 14) 已销毁的 iframe：之后任何消息都拒
// ---------------------------------------------------------------------------
{
  const ctx = makeBridge()
  ctx.bridge.destroy()
  const d = ctx.bridge.handleMessage({ source: ctx.win, data: hostRequest(ctx.init) })
  assert.equal(d.accepted, false)
  assert.equal(d.code, 'DESTROYED')
  assert.equal(ctx.calls.host.length, 0)
  console.log('OK 14) 已销毁 iframe 的消息被拒')
}

// ---------------------------------------------------------------------------
// 15) sandbox 只给 allow-scripts
// ---------------------------------------------------------------------------
{
  const source = readFileSync(new URL('../src/PluginView.tsx', import.meta.url), 'utf8')
  const match = source.match(/sandbox="([^"]*)"/)
  assert.ok(match, 'PluginView 必须给 iframe 设 sandbox')
  assert.equal(match[1], 'allow-scripts')
  assert.ok(!/sandbox="[^"]*allow-same-origin/.test(source), '不能给 allow-same-origin')
  assert.ok(!/sandbox="[^"]*allow-top-navigation/.test(source), '不能给顶层导航')
  assert.ok(!/sandbox="[^"]*allow-popups/.test(source), '不能给弹窗')
  console.log('OK 15) iframe sandbox 只有 allow-scripts')
}

console.log('\n全部通过：插件路由 / 菜单 / 拒绝清单 / 消息桥共 15 组判定')
