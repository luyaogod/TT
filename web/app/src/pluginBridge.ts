// 第三方插件页面与宿主之间的消息桥（父页面一侧）。
//
// 这个模块是**权限执行点**：它不接受 iframe 自报的插件 id/路由/权限，只认自己绑定的
// 那份事实（pluginId + route + 本次 iframe 的 session/nonce）。所有入站消息必须先过
// 「来源 + nonce + 类型白名单」，再过方法/路径白名单，最后才可能到达网络。
//
// 为什么 targetOrigin 是 "*"：见 pluginProtocol.ts 顶部的说明 —— iframe 是不透明源，
// 没有可指定的具体 origin。认证由 source + 一次性 nonce 承担，而不是 origin 字符串。
//
// 纯 Node 可加载：模块顶层不碰 window/document；只有 createFetchTransport 里会用到 fetch。

import {
  BRIDGE_PROTOCOL,
  BRIDGE_VERSION,
  MAX_MESSAGE_BYTES,
  MAX_RESPONSE_ROWS,
  isCancelMessage,
  isErrorMessage,
  isHostRequestMessage,
  isInboundType,
  isRpcRequestMessage,
  messageBytes,
  payloadLimitError,
  readEnvelope,
  utf8Bytes,
  type BridgeEnvelope,
  type BridgeErrorBody,
  type BridgeErrorCode,
  type BridgeSession,
  type BridgeType,
  type ErrorMessage,
  type HostRequestMessage,
  type HostResponseMessage,
  type InboundType,
  type ParentToPluginMessage,
  type RpcRequestMessage,
  type RpcResponseMessage,
} from './pluginProtocol'
import {
  normalizeHostApiMethod,
  normalizePluginRpcPath,
  pluginHostApiBridgePath,
  pluginRPCBridgePath,
} from './plugins'

/** 一次 iframe 的桥会话。nonce 是 32 字节随机串，只出现在 init 与该 iframe 的后续消息里。 */
export function createBridgeSession(pluginId: string, route: string): BridgeSession {
  return {
    pluginId,
    route,
    sessionId: randomHex(16),
    nonce: randomHex(32),
  }
}

function randomHex(bytes: number): string {
  const c = globalThis.crypto
  if (!c || typeof c.getRandomValues !== 'function') {
    // 退化成可预测的随机数会让 nonce 失去意义；宁可不建立桥。
    throw new Error('当前环境没有 crypto.getRandomValues，无法建立插件消息桥')
  }
  const buf = new Uint8Array(bytes)
  c.getRandomValues(buf)
  let out = ''
  for (const b of buf) out += b.toString(16).padStart(2, '0')
  return out
}

/** 能拿到 contentWindow 的 iframe（测试里用假对象即可）。 */
export interface BridgeFrameLike {
  contentWindow: unknown
}

/** 只需要 source/data 的消息事件（Node 检查项不构造真的 MessageEvent）。 */
export interface BridgeMessageEventLike {
  source: unknown
  data: unknown
  origin?: string
}

export interface HostApiCall {
  pluginId: string
  method: string
  payload: unknown
  signal: AbortSignal
}

export interface PluginRpcCall {
  pluginId: string
  path: string
  payload: unknown
  signal: AbortSignal
}

/**
 * 网络出口。桥只调用这两个方法 —— 没有“通用 URL 代理”的口子。
 * 生产实现在 createFetchTransport；检查项注入桩来观察调用。
 */
export interface PluginBridgeTransport {
  hostApi(call: HostApiCall): Promise<unknown>
  pluginRPC(call: PluginRpcCall): Promise<unknown>
}

export interface PluginBridgeOptions {
  /** 父页面绑定的真实插件 id。 */
  pluginId: string
  /** 父页面绑定的真实路由（同一插件未来可挂多条页面时用来区分实例）。 */
  route: string
  transport: PluginBridgeTransport
  /** 可见错误回调：未知消息类型、体量超限、对端报错等会走到这里。 */
  onError?: (error: BridgeErrorBody) => void
}

export type BridgeDecision =
  | { accepted: true; type: InboundType }
  | { accepted: false; code: BridgeErrorCode; message: string }

/** 已见消息 id 的上限。只用于拒绝重复投递，不做审计留存。 */
const SEEN_LIMIT = 512

export class PluginBridge {
  private readonly transport: PluginBridgeTransport
  private readonly onError?: (error: BridgeErrorBody) => void
  private session: BridgeSession
  private frameWindow: unknown = null
  private destroyed = false
  private readonly seen = new Map<string, true>()
  private readonly pending = new Map<string, AbortController>()
  private readonly jobs = new Set<Promise<void>>()

  constructor(opts: PluginBridgeOptions) {
    this.transport = opts.transport
    this.onError = opts.onError
    this.session = createBridgeSession(opts.pluginId, opts.route)
  }

  get sessionState(): Readonly<BridgeSession> {
    return this.session
  }

  get isDestroyed(): boolean {
    return this.destroyed
  }

  /**
   * 绑定一个已挂载的 iframe 并发出 init。
   *
   * 每次真正 load 都换一套 session/nonce：同一个组件里 iframe 重新加载后，
   * 旧文档与新文档不应共享认证串。调用方在 iframe 的 onLoad 里调用它。
   */
  attach(frame: BridgeFrameLike | null): boolean {
    if (this.destroyed) return false
    const win = frame?.contentWindow
    if (!win) return false
    // 重新加载：上一份文档的在途请求已经失去意义，先取消。
    for (const ac of this.pending.values()) ac.abort()
    this.pending.clear()
    this.frameWindow = win
    this.session = createBridgeSession(this.session.pluginId, this.session.route)
    this.post({
      ...this.envelope('init'),
      type: 'init',
      pluginId: this.session.pluginId,
      route: this.session.route,
      hostApiVersion: 1,
      limits: { maxMessageBytes: MAX_MESSAGE_BYTES, maxRows: MAX_RESPONSE_ROWS },
    })
    return true
  }

  /**
   * 处理一条 window message。返回值只描述“是否受理”；被受理的请求会异步跑完。
   * 检查顺序是刻意的：销毁 → 来源 → 信封 → nonce → 版本 → 类型 → 体量 → 形状。
   */
  handleMessage(event: BridgeMessageEventLike): BridgeDecision {
    if (this.destroyed) return reject('DESTROYED', '插件 iframe 已销毁')

    // 来源不是这个 iframe 的窗口：别的 frame/窗口在说话，静默丢弃。
    // 这里不看 event.origin —— 沙箱 iframe 的 origin 是字面量 "null"，它证明不了身份。
    if (!this.frameWindow || event.source !== this.frameWindow) {
      return reject('UNAUTHORIZED', '消息来源不是当前插件 iframe')
    }

    const raw = readEnvelope(event.data)
    if (!raw) {
      // 已经能确认来源，但连信封都读不出来：这是插件实现问题，给可见错误。
      this.emit({ code: 'BAD_ENVELOPE', message: '插件消息不是合法的 v1 信封' })
      return reject('BAD_ENVELOPE', '消息信封不合法')
    }

    // nonce/session 不匹配一律静默：任何本机页面都能向本窗口 postMessage，
    // 如果把它们都画到界面上，一个无关页面就能反复污染宿主 UI。
    if (raw.nonce !== this.session.nonce || raw.session !== this.session.sessionId) {
      return reject('UNAUTHORIZED', 'nonce 或会话不匹配')
    }

    if (raw.protocol !== BRIDGE_PROTOCOL || raw.v !== BRIDGE_VERSION) {
      const body: BridgeErrorBody = {
        code: 'BAD_VERSION',
        message: `不支持的桥协议 ${raw.protocol} v${raw.v}（本机只支持 ${BRIDGE_PROTOCOL} v${BRIDGE_VERSION}）`,
      }
      this.emit(body)
      this.sendError(undefined, body)
      return reject(body.code, body.message)
    }

    if (!isInboundType(raw.type)) {
      // 未知类型必须可见，并回一条 error 让插件知道自己的消息被拒了。
      const body: BridgeErrorBody = {
        code: 'UNKNOWN_TYPE',
        message: `v1 不接受插件页面发来的消息类型「${raw.type}」`,
      }
      this.emit(body)
      this.sendError(undefined, body)
      return reject(body.code, body.message)
    }

    if (!this.remember(raw.id)) {
      const body: BridgeErrorBody = { code: 'REPLAY', message: '同一条消息被重复投递' }
      this.emit(body)
      this.sendError(undefined, body)
      return reject(body.code, body.message)
    }

    const size = messageBytes(event.data)
    if (size > MAX_MESSAGE_BYTES) {
      const body: BridgeErrorBody = {
        code: 'PAYLOAD_TOO_LARGE',
        message: `消息 ${size} 字节，超过上限 ${MAX_MESSAGE_BYTES} 字节`,
      }
      this.emit(body)
      this.sendError(undefined, body)
      return reject(body.code, body.message)
    }

    if (raw.type === 'host.request') {
      if (!isHostRequestMessage(event.data)) {
        return this.badShape('host.request')
      }
      const method = normalizeHostApiMethod(event.data.method)
      if (!method) {
        const body: BridgeErrorBody = {
          code: 'BAD_METHOD',
          message: `Host API 方法不合法：${event.data.method}`,
        }
        this.fail(event.data.reqId, body, true)
        return reject(body.code, body.message)
      }
      this.track(this.runHost(event.data, method))
      return { accepted: true, type: 'host.request' }
    }

    if (raw.type === 'rpc.request') {
      if (!isRpcRequestMessage(event.data)) {
        return this.badShape('rpc.request')
      }
      const rpcPath = normalizePluginRpcPath(event.data.path)
      if (!rpcPath) {
        const body: BridgeErrorBody = {
          code: 'DENIED',
          message: `插件 RPC 路径不合法或指向宿主管理面：${event.data.path}`,
        }
        this.fail(event.data.reqId, body, true)
        return reject(body.code, body.message)
      }
      this.track(this.runRPC(event.data, rpcPath))
      return { accepted: true, type: 'rpc.request' }
    }

    if (raw.type === 'cancel') {
      if (!isCancelMessage(event.data)) return this.badShape('cancel')
      this.pending.get(event.data.reqId)?.abort()
      return { accepted: true, type: 'cancel' }
    }

    if (!isErrorMessage(event.data)) return this.badShape('error')
    // 插件自报的错误：显示出来，不回声（否则两边可以互相刷 error）。
    this.emit({
      code: event.data.error.code,
      message: event.data.error.message,
      detail: event.data.error.detail,
    })
    return { accepted: true, type: 'error' }
  }

  /** 等所有在途请求落定（检查项用；生产调用方不需要）。 */
  async flush(): Promise<void> {
    while (this.jobs.size > 0) {
      await Promise.allSettled(Array.from(this.jobs))
    }
  }

  /** 销毁：取消在途请求、摘掉窗口引用，之后任何消息都按 DESTROYED 拒绝。 */
  destroy(): void {
    if (this.destroyed) return
    this.destroyed = true
    this.frameWindow = null
    for (const ac of this.pending.values()) ac.abort()
    this.pending.clear()
    this.seen.clear()
  }

  // ---- 内部 ----

  private envelope(type: BridgeType): BridgeEnvelope {
    return {
      protocol: BRIDGE_PROTOCOL,
      v: BRIDGE_VERSION,
      type,
      id: randomHex(8),
      session: this.session.sessionId,
      nonce: this.session.nonce,
      ts: Date.now(),
    }
  }

  /**
   * 投递到 iframe。targetOrigin 只能是 "*"（不透明源）；安全性由 nonce 与 source 校验保证。
   */
  private post(msg: ParentToPluginMessage): void {
    if (this.destroyed || !this.frameWindow) return
    const target = this.frameWindow as { postMessage?: (message: unknown, targetOrigin: string) => void }
    try {
      target.postMessage?.(msg, '*')
    } catch (e) {
      this.emit({ code: 'TRANSPORT_ERROR', message: `向插件页面投递消息失败：${errText(e)}` })
    }
  }

  private sendError(reqId: string | undefined, body: BridgeErrorBody): void {
    const msg: ErrorMessage = {
      ...this.envelope('error'),
      type: 'error',
      ...(reqId ? { reqId } : {}),
      error: body,
    }
    this.post(msg)
  }

  private emit(body: BridgeErrorBody): void {
    try {
      this.onError?.(body)
    } catch {
      // 界面回调自己抛错不能拖垮消息桥。
    }
  }

  private rejectWithError(reqId: string, body: BridgeErrorBody, visible: boolean): void {
    if (visible) this.emit(body)
    this.sendError(reqId, body)
  }

  private badShape(type: string): BridgeDecision {
    const body: BridgeErrorBody = {
      code: 'BAD_ENVELOPE',
      message: `「${type}」消息缺字段或字段类型不对`,
    }
    this.emit(body)
    this.sendError(undefined, body)
    return reject(body.code, body.message)
  }

  private fail(reqId: string, body: BridgeErrorBody, visible: boolean): void {
    this.rejectWithError(reqId, body, visible)
  }

  private remember(id: string): boolean {
    if (this.seen.has(id)) return false
    this.seen.set(id, true)
    if (this.seen.size > SEEN_LIMIT) {
      const oldest = this.seen.keys().next().value
      if (oldest !== undefined) this.seen.delete(oldest)
    }
    return true
  }

  private track(job: Promise<void>): void {
    this.jobs.add(job)
    void job.finally(() => this.jobs.delete(job)).catch(() => {})
  }

  private async runHost(msg: HostRequestMessage, method: string): Promise<void> {
    const ac = new AbortController()
    this.pending.set(msg.reqId, ac)
    try {
      const result = await this.transport.hostApi({
        pluginId: this.session.pluginId,
        method,
        payload: msg.payload,
        signal: ac.signal,
      })
      if (this.destroyed) return
      if (ac.signal.aborted) {
        this.fail(msg.reqId, { code: 'CANCELLED', message: '请求已取消' }, false)
        return
      }
      const limit = payloadLimitError(result)
      if (limit) {
        this.fail(msg.reqId, limit, true)
        return
      }
      const response: HostResponseMessage = {
        ...this.envelope('host.response'),
        type: 'host.response',
        reqId: msg.reqId,
        ok: true,
        result,
      }
      this.post(response)
    } catch (e) {
      if (this.destroyed) return
      const body = errorBodyOf(e, ac.signal.aborted)
      this.fail(msg.reqId, body, body.code !== 'CANCELLED')
    } finally {
      this.pending.delete(msg.reqId)
    }
  }

  private async runRPC(msg: RpcRequestMessage, path: string): Promise<void> {
    const ac = new AbortController()
    this.pending.set(msg.reqId, ac)
    try {
      const result = await this.transport.pluginRPC({
        pluginId: this.session.pluginId,
        path,
        payload: msg.payload,
        signal: ac.signal,
      })
      if (this.destroyed) return
      if (ac.signal.aborted) {
        this.fail(msg.reqId, { code: 'CANCELLED', message: '请求已取消' }, false)
        return
      }
      const limit = payloadLimitError(result)
      if (limit) {
        this.fail(msg.reqId, limit, true)
        return
      }
      const response: RpcResponseMessage = {
        ...this.envelope('rpc.response'),
        type: 'rpc.response',
        reqId: msg.reqId,
        ok: true,
        result,
      }
      this.post(response)
    } catch (e) {
      if (this.destroyed) return
      const body = errorBodyOf(e, ac.signal.aborted)
      this.fail(msg.reqId, body, body.code !== 'CANCELLED')
    } finally {
      this.pending.delete(msg.reqId)
    }
  }
}

function reject(code: BridgeErrorCode, message: string): BridgeDecision {
  return { accepted: false, code, message }
}

function errText(e: unknown): string {
  return e instanceof Error ? e.message : String(e)
}

/** 传输层的可识别错误：code 原样回给 iframe，不让它退化成模糊的“失败”。 */
export class BridgeTransportError extends Error {
  readonly code: BridgeErrorCode

  constructor(code: BridgeErrorCode, message: string) {
    super(message)
    this.name = 'BridgeTransportError'
    this.code = code
  }
}

function errorBodyOf(e: unknown, aborted: boolean): BridgeErrorBody {
  if (e instanceof BridgeTransportError) return { code: e.code, message: e.message }
  if (aborted) return { code: 'CANCELLED', message: '请求已取消' }
  return { code: 'TRANSPORT_ERROR', message: errText(e) }
}

// ---------------------------------------------------------------------------
// 生产出口：同源 fetch 到两条受控桥路径
// ---------------------------------------------------------------------------

/**
 * 父页面自己的网络出口。两条路径都不是 /api/：
 *   · Host API → /plugins/<id>/hostrpc/<service>/<method>（后端换实例令牌后转 /hostrpc/）
 *   · 插件服务 → /plugins/<id>/rpc/<path>（后端只转发到该插件实例）
 */
export function createFetchTransport(): PluginBridgeTransport {
  return {
    hostApi: (call) => {
      const url = pluginHostApiBridgePath(call.pluginId, call.method)
      if (!url) return Promise.reject(new BridgeTransportError('BAD_METHOD', 'Host API 方法不合法'))
      return postJSON(url, call.payload, call.signal)
    },
    pluginRPC: (call) => {
      const url = pluginRPCBridgePath(call.pluginId, call.path)
      if (!url) return Promise.reject(new BridgeTransportError('DENIED', '插件 RPC 路径不合法'))
      return postJSON(url, call.payload, call.signal)
    },
  }
}

async function postJSON(url: string, payload: unknown, signal: AbortSignal): Promise<unknown> {
  const r = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload ?? null),
    signal,
  })
  const text = await readBoundedText(r, MAX_MESSAGE_BYTES)
  if (!r.ok) {
    throw new BridgeTransportError('TRANSPORT_ERROR', `HTTP ${r.status}${text ? '：' + text.slice(0, 300) : ''}`)
  }
  if (!text) return null
  try {
    return JSON.parse(text)
  } catch {
    // 插件自身服务的响应不归宿主理解，非 JSON 就原样交给页面。
    return text
  }
}

/**
 * 有上限地读响应体：先看 Content-Length，再边读边计数。
 * 不先 `await r.text()` 再判断 —— 那等于先把大响应完整缓冲进内存，正好违背“不静默缓冲”。
 */
async function readBoundedText(r: Response, limit: number): Promise<string> {
  const declared = Number(r.headers.get('content-length') || 0)
  if (Number.isFinite(declared) && declared > limit) {
    throw new BridgeTransportError('PAYLOAD_TOO_LARGE', `响应 ${declared} 字节，超过上限 ${limit} 字节`)
  }
  if (!r.body) {
    const text = await r.text()
    if (utf8Bytes(text) > limit) {
      throw new BridgeTransportError('PAYLOAD_TOO_LARGE', `响应超过上限 ${limit} 字节`)
    }
    return text
  }
  const reader = r.body.getReader()
  const chunks: Uint8Array[] = []
  let total = 0
  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    total += value.byteLength
    if (total > limit) {
      await reader.cancel().catch(() => {})
      throw new BridgeTransportError('PAYLOAD_TOO_LARGE', `响应超过上限 ${limit} 字节`)
    }
    chunks.push(value)
  }
  const merged = new Uint8Array(total)
  let offset = 0
  for (const c of chunks) {
    merged.set(c, offset)
    offset += c.byteLength
  }
  return new TextDecoder().decode(merged)
}
