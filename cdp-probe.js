// 最小 CDP 客户端（零依赖）：连无头 Edge，导航到调试台，执行 JS 读 DOM。
// 用法: node cdp-probe.js <url> <js表达式...>
const http = require('http')

const PORT = Number(process.env.CDP_PORT || 9333)
const URL_TO_OPEN = process.argv[2] || 'about:blank'
const EVAL = process.argv.slice(3)

function getJSON(path) {
  return new Promise((resolve, reject) => {
    http.get({ host: '127.0.0.1', port: PORT, path }, (res) => {
      let b = ''
      res.on('data', (d) => (b += d))
      res.on('end', () => {
        try { resolve(JSON.parse(b)) } catch (e) { reject(new Error('bad json: ' + b.slice(0, 200))) }
      })
    }).on('error', reject)
  })
}

async function main() {
  const list = await getJSON('/json/list')
  const page = list.find((t) => t.type === 'page')
  if (!page) throw new Error('没有 page target')
  const ws = new WebSocket(page.webSocketDebuggerUrl)
  let id = 0
  const pending = new Map()
  const send = (method, params = {}) => new Promise((resolve, reject) => {
    const myId = ++id
    pending.set(myId, { resolve, reject })
    ws.send(JSON.stringify({ id: myId, method, params }))
  })
  ws.addEventListener('message', (ev) => {
    const msg = JSON.parse(ev.data)
    if (msg.id && pending.has(msg.id)) {
      const p = pending.get(msg.id)
      pending.delete(msg.id)
      if (msg.error) p.reject(new Error(JSON.stringify(msg.error)))
      else p.resolve(msg.result)
    }
  })
  await new Promise((r) => ws.addEventListener('open', r, { once: true }))

  if (URL_TO_OPEN !== '-') {
    await send('Page.enable')
    await send('Page.navigate', { url: URL_TO_OPEN })
    // 等前端挂载（Vite 产物大，给足时间）
    await new Promise((r) => setTimeout(r, 6000))
  }

  for (const expr of EVAL) {
    const r = await send('Runtime.evaluate', {
      expression: expr, awaitPromise: true, returnByValue: true,
    })
    if (r.exceptionDetails) {
      console.log('EXC:', JSON.stringify(r.exceptionDetails.exception?.description || r.exceptionDetails).slice(0, 500))
    } else {
      console.log(JSON.stringify(r.result.value))
    }
  }
  ws.close()
}

main().catch((e) => { console.error('ERR', e.message); process.exit(1) })
