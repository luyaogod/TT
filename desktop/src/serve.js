// 桌面壳与 tt 服务之间的启动/收尾契约。
//
// 这一层**不含 Electron API**（只依赖 node:child_process 与 fs），所以能用 `node --test`
// 直接测 —— 壳里最容易出错的一段正是"等它就绪、判断该不该收掉它"，而这段一旦错了，
// 表现是窗口白屏或关不掉的隐形进程，没人会想到是启动逻辑的问题。
//
// 契约的另一半在 Go：`tt serve --foreground --ready-json --exit-on-stdin-close`
// （见 internal/cli/debug/serve_ready.go）。三件事都以那边的输出为准，这里不许猜：
//
//   真实地址  端口被占用时服务会 +1 顺延，只有子进程知道最后落在哪
//   所有权    attached=true 表示"同目录已有实例，我没启动新的"→ 退出时不能收它
//   收尾      stdin 关闭 = 优雅退出（壳被强杀时靠它带走子进程）

import { spawn as nodeSpawn } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'

/** 等就绪的上限。与服务自己的后台启动口径一致（15 秒，见 servebg.go）。 */
export const READY_TIMEOUT_MS = 15000
/** 收尾等待：先请它自己退，超时才硬杀。 */
export const STOP_TIMEOUT_MS = 5000
/** 诊断用的尾部行数。 */
const TAIL_LINES = 40

/**
 * 桌面版载荷里的 tt.exe 在哪。
 *
 * 打包后固定在 <resources>\tt\tt.exe（electron-builder 的 extraResources）。
 * 开发时用仓库根的 tt.exe；TT_DESKTOP_EXE 可以指到别处（调试便携包里的那份）。
 */
export function resolveExe({ isPackaged, resourcesPath, repoRoot, env = {}, exists = fs.existsSync }) {
  if (isPackaged) return path.join(resourcesPath, 'tt', 'tt.exe')
  const override = String(env.TT_DESKTOP_EXE || '').trim()
  if (override) return override
  const local = path.join(repoRoot, 'tt.exe')
  // 开发态没先 `make build` 时给出可照做的路径，而不是让 spawn 报 ENOENT
  return exists(local) ? local : local
}

/** 发给服务的参数。**顺序与开关名是契约**，改这里等于改 Go 侧的行为。 */
export function serveArgs() {
  return ['serve', '--foreground', '--ready-json', '--format', 'json', '--exit-on-stdin-close']
}

/** 把一段可能含多行的文本切成完整行，返回 [完整行[], 余下不完整的部分]。 */
export function splitLines(buffer) {
  const lines = []
  let rest = buffer
  let i
  while ((i = rest.indexOf('\n')) >= 0) {
    lines.push(rest.slice(0, i).replace(/\r$/, ''))
    rest = rest.slice(i + 1)
  }
  return [lines, rest]
}

/** 从错误信封里取出人话。信封形状见 internal/output：{ok:false,error:{code,message}}。 */
export function errorText(envelope) {
  const e = envelope && envelope.error
  if (typeof e === 'string') return e
  if (e && typeof e === 'object' && typeof e.message === 'string') return e.message
  return '服务启动失败（没有给出原因）'
}

/**
 * 解析一行就绪输出。
 *
 * 返回 null = 这行不是就绪契约（是日志或人读块），调用方按日志处理。
 * 返回 {ok:false} = 契约里的失败信封，调用方要把 error 顶到失败页。
 */
export function parseReadyLine(line) {
  const s = String(line ?? '').trim()
  if (!s.startsWith('{')) return null
  let obj
  try {
    obj = JSON.parse(s)
  } catch {
    return null
  }
  if (!obj || typeof obj !== 'object') return null
  if (obj.ok === false) return { ok: false, error: errorText(obj) }
  if (obj.ok !== true) return null
  if (typeof obj.url !== 'string' || !/^https?:\/\//.test(obj.url)) return null
  return {
    ok: true,
    attached: obj.attached === true,
    url: obj.url,
    pid: Number(obj.pid) || 0,
    config: typeof obj.config === 'string' ? obj.config : '',
  }
}

/** 由服务地址推页面地址：SPA 挂在 /debug/ 下（见 internal/web 的 routes）。 */
export function appUrl(serverUrl, hash = '') {
  const base = String(serverUrl || '').replace(/\/+$/, '')
  return `${base}/debug/${hash ? '#' + hash.replace(/^#/, '') : ''}`
}

/** 由服务地址推 origin（外链判定与同源限制用）。 */
export function originOf(serverUrl) {
  try {
    return new URL(serverUrl).origin
  } catch {
    return ''
  }
}

/** 启动失败时抛这个：带上尾巴日志，失败页才有东西可看。 */
export class ServeStartError extends Error {
  constructor(message, tail = '') {
    super(message)
    this.name = 'ServeStartError'
    this.tail = tail
  }
}

function isAlive(child) {
  return !!child && child.exitCode === null && !child.killed
}

/**
 * 起服务并等它就绪。
 *
 * 返回的句柄是**唯一**知道"该不该收掉它"的地方：
 *   attached=true  → 属于别的进程（命令行先起的），stop() 什么也不做
 *   attached=false → 我们拉起来的，stop() 走优雅收尾
 */
export function startServer({
  exePath,
  env = process.env,
  spawn = nodeSpawn,
  timeoutMs = READY_TIMEOUT_MS,
  onLog = () => {},
  onExit = () => {},
  platform = process.platform,
}) {
  return new Promise((resolve, reject) => {
    let child
    try {
      child = spawn(exePath, serveArgs(), {
        windowsHide: true,
        stdio: ['pipe', 'pipe', 'pipe'],
        env: { ...env, TT_DESKTOP: '1' },
      })
    } catch (err) {
      reject(new ServeStartError(`启动 ${exePath} 失败: ${err?.message || err}`))
      return
    }

    const tail = []
    const push = (line) => {
      if (!line) return
      tail.push(line)
      if (tail.length > TAIL_LINES) tail.shift()
      onLog(line)
    }
    const tailText = () => tail.slice(-TAIL_LINES).join('\n')

    let settled = false
    let handle = null
    const timer = setTimeout(() => {
      if (settled) return
      settled = true
      try {
        child.kill()
      } catch {
        /* 已经退出就算了 */
      }
      reject(new ServeStartError(`服务 ${timeoutMs / 1000} 秒内没有就绪（端口可能全被占用）`, tailText()))
    }, timeoutMs)

    const settleOk = (ready) => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      handle = {
        ...ready,
        child,
        // 只有自己拉起来的那份才由壳收尾
        get owned() {
          return !ready.attached
        },
        stop: () => stopServer(handle, { platform }),
      }
      resolve(handle)
    }

    const settleErr = (err) => {
      if (settled) return
      settled = true
      clearTimeout(timer)
      reject(err instanceof ServeStartError ? err : new ServeStartError(String(err?.message || err), tailText()))
    }

    // 按行处理 **必须跨 chunk 累积**：一行就绪 JSON 完全可能被拆进两次 read
    // （管道读的边界与行边界无关）。早期版本按 chunk 各切一次，就绪行被拆开时
    // 永远认不出来 —— 表现是"服务明明起来了，壳却等到超时"。
    let outBuf = ''
    const feed = (chunk) => {
      outBuf += chunk
      const [lines, rest] = splitLines(outBuf)
      outBuf = rest
      for (const line of lines) {
        const ready = parseReadyLine(line)
        if (ready && !settled) {
          if (ready.ok) settleOk(ready)
          else settleErr(new ServeStartError(ready.error, tailText()))
          continue
        }
        push(line)
      }
    }

    let errBuf = ''
    const feedErr = (chunk) => {
      errBuf += chunk
      const [lines, rest] = splitLines(errBuf)
      errBuf = rest
      for (const line of lines) push(line)
    }

    child.stdout?.setEncoding?.('utf8')
    child.stdout?.on?.('data', feed)
    child.stderr?.setEncoding?.('utf8')
    child.stderr?.on?.('data', feedErr)
    child.on?.('error', (err) => settleErr(err))
    child.on?.('exit', (code, signal) => {
      if (!settled) {
        settleErr(new ServeStartError(`服务提前退出（退出码 ${code ?? signal}）`, tailText()))
        return
      }
      // 附着的那份本来就会在打完就绪行后立刻退出（它只是来报地址的）
      if (handle?.attached) return
      onExit({ code, signal })
    })

    // 起手就把句柄挂上，方便 stop() 在极端时序下也能拿到 child
    queueMicrotask(() => {
      if (settled && !handle) {
        try {
          child.kill()
        } catch {
          /* ignore */
        }
      }
    })
  })
}

function waitExit(child, ms) {
  return new Promise((resolve) => {
    if (!isAlive(child)) {
      resolve(true)
      return
    }
    let done = false
    const finish = (v) => {
      if (done) return
      done = true
      clearTimeout(timer)
      resolve(v)
    }
    const timer = setTimeout(() => finish(false), ms)
    child.once('exit', () => finish(true))
  })
}

/**
 * 收尾：附着的什么都不做；自己拉起来的先请它退，退不掉再硬杀。
 *
 * 顺序是有讲究的：先 POST /api/shutdown（让 Go 侧把调试会话与 SSH 连接收口），
 * 同时关掉 stdin —— 两条独立的路，任一条先到都算成功。
 */
export async function stopServer(handle, { platform = process.platform, timeoutMs = STOP_TIMEOUT_MS } = {}) {
  if (!handle || handle.attached) return
  const child = handle.child
  if (!isAlive(child)) return

  try {
    await fetch(`${handle.url.replace(/\/+$/, '')}/api/shutdown`, { method: 'POST' })
  } catch {
    /* 服务可能已经没了：下面按退出判定 */
  }
  try {
    child.stdin?.destroy?.()
  } catch {
    /* ignore */
  }

  if (await waitExit(child, timeoutMs)) return

  try {
    child.kill()
  } catch {
    /* ignore */
  }
  if (platform === 'win32' && handle.pid > 0) {
    // 进程树：tzs 引擎是被 tt 拉起来的，只杀 tt 会留下它
    try {
      nodeSpawn('taskkill', ['/pid', String(handle.pid), '/T', '/F'], { windowsHide: true, stdio: 'ignore' })
    } catch {
      /* ignore */
    }
  }
}
