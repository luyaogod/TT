// 插件页面消息桥的 v1 协议（父页面 ↔ 第三方 iframe）。
//
// 为什么单独一个纯模块：消息形状、体量上限与判据都要能被 Node 侧的检查项直接加载
//（scripts/plugins.test.mjs），所以这里不碰 DOM、不依赖 React。
//
// 认证模型（重要）：
//   iframe 是 sandbox="allow-scripts"、**没有** allow-same-origin，文档处于不透明源，
//   浏览器给出的 event.origin 是字面量 "null"。父页面向 iframe 发消息时因此**只能**用
//   targetOrigin="*"（没有可用的具体 origin）；反过来，父页面不拿 origin 当身份，而是用
//   「event.source === iframe.contentWindow」+ 每条消息里的随机 nonce 认证。
//   两者缺一不可：source 证明消息来自这个 iframe，nonce 证明对端读过父页面发出的 init。
//
// 版本边界：v 是协议版本，不是插件版本。新增字段可以留在 v1；改字段含义/删除字段才升 v。

export const BRIDGE_PROTOCOL = 'tt.plugin.bridge' as const
export const BRIDGE_VERSION = 1 as const

/** 一条消息的 UTF-8 上限：1 MiB。超过就显式报错，不做“截断后照发”。 */
export const MAX_MESSAGE_BYTES = 1024 * 1024
/** 单次 Host API / RPC 响应的行数上限。 */
export const MAX_RESPONSE_ROWS = 5000

/** 一个 iframe 实例的桥会话：sessionId 用来区分实例，nonce 是唯一的认证凭据。 */
export interface BridgeSession {
  pluginId: string
  route: string
  sessionId: string
  nonce: string
}

export type BridgeType =
  | 'init'
  | 'host.request'
  | 'host.response'
  | 'rpc.request'
  | 'rpc.response'
  | 'cancel'
  | 'error'

export type BridgeErrorCode =
  | 'BAD_ENVELOPE'
  | 'BAD_VERSION'
  | 'UNKNOWN_TYPE'
  | 'UNAUTHORIZED'
  | 'REPLAY'
  | 'DESTROYED'
  | 'NOT_READY'
  | 'BAD_METHOD'
  | 'DENIED'
  | 'PAYLOAD_TOO_LARGE'
  | 'TOO_MANY_ROWS'
  | 'CANCELLED'
  | 'TRANSPORT_ERROR'
  | 'PLUGIN_ERROR'

/** 错误体的稳定 code + 人读文案。detail 只放排查细节，不放任何凭据。 */
export interface BridgeErrorBody {
  code: BridgeErrorCode
  message: string
  detail?: string
}

/** 所有消息共用的信封字段。 */
export interface BridgeEnvelope {
  protocol: typeof BRIDGE_PROTOCOL
  v: typeof BRIDGE_VERSION
  type: BridgeType
  /** 消息 id（随机）。父页面用它拒绝同一条消息的重复投递。 */
  id: string
  /** 桥会话 id（不是秘密，只用于把消息对到实例上）。 */
  session: string
  /** 认证用随机串；每条 plugin→parent 消息都必须原样带回 init 里那一个。 */
  nonce: string
  /** 发送时刻（毫秒）。只用于日志/排查，不参与授权。 */
  ts: number
}

/**
 * init：父页面 → iframe。只有这一条消息携带“这一轮”的会话事实。
 * iframe 必须等收到它之后才能发请求 —— 之前父页面还没有可用来认证的 nonce。
 */
export interface InitMessage extends BridgeEnvelope {
  type: 'init'
  /** 父页面绑定的真实插件 id。iframe 自报的 id 永远不作为授权依据。 */
  pluginId: string
  route: string
  /** Host API 主版本（与 /hostrpc/ 的契约对应）。 */
  hostApiVersion: 1
  limits: {
    maxMessageBytes: number
    maxRows: number
  }
}

/** Host API 一元请求：插件页面 → 父页面，父页面再走 /hostrpc/ 的桥。 */
export interface HostRequestMessage extends BridgeEnvelope {
  type: 'host.request'
  reqId: string
  /** ConnectRPC 形态：<package.Service>/<Method>，例如 tt.host.v1.ProfilesService/List。 */
  method: string
  payload: unknown
}

export interface HostResponseMessage extends BridgeEnvelope {
  type: 'host.response'
  reqId: string
  ok: boolean
  result?: unknown
  error?: BridgeErrorBody
}

/** 插件自身服务的 RPC：插件页面 → 父页面，父页面转发到 /plugins/<id>/rpc/…。 */
export interface RpcRequestMessage extends BridgeEnvelope {
  type: 'rpc.request'
  reqId: string
  /** 相对插件服务的路径（以 / 开头）。宿主管理面路径在父页面侧拒绝。 */
  path: string
  payload: unknown
}

export interface RpcResponseMessage extends BridgeEnvelope {
  type: 'rpc.response'
  reqId: string
  ok: boolean
  result?: unknown
  error?: BridgeErrorBody
}

/** 取消一条在途请求（任一方向都可以发）。 */
export interface CancelMessage extends BridgeEnvelope {
  type: 'cancel'
  reqId: string
}

/** 错误（任一方向都可以发）：对端用它把失败原因显式说出来。 */
export interface ErrorMessage extends BridgeEnvelope {
  type: 'error'
  reqId?: string
  error: BridgeErrorBody
}

/** 父页面发给 iframe 的消息。 */
export type ParentToPluginMessage =
  | InitMessage
  | HostResponseMessage
  | RpcResponseMessage
  | CancelMessage
  | ErrorMessage

/** iframe 发给父页面、且 v1 允许父页面接收的消息类型。 */
export const INBOUND_TYPES = ['host.request', 'rpc.request', 'cancel', 'error'] as const
export type InboundType = (typeof INBOUND_TYPES)[number]

export function isInboundType(t: string): t is InboundType {
  return (INBOUND_TYPES as readonly string[]).includes(t)
}

/** 所有稳定错误码（运行时校验对端发来的 error.code 用）。 */
export const BRIDGE_ERROR_CODES: readonly BridgeErrorCode[] = [
  'BAD_ENVELOPE', 'BAD_VERSION', 'UNKNOWN_TYPE', 'UNAUTHORIZED', 'REPLAY',
  'DESTROYED', 'NOT_READY', 'BAD_METHOD', 'DENIED', 'PAYLOAD_TOO_LARGE',
  'TOO_MANY_ROWS', 'CANCELLED', 'TRANSPORT_ERROR', 'PLUGIN_ERROR',
]

export function isBridgeErrorCode(v: unknown): v is BridgeErrorCode {
  return typeof v === 'string' && (BRIDGE_ERROR_CODES as readonly string[]).includes(v)
}

/**
 * 只做“信封字段是否存在且类型正确”的读取，**不**判断协议版本、nonce 或消息类型。
 * 顺序很重要：先读信封（认不出来就静默丢弃），再用 nonce 认证，最后才谈消息语义。
 */
export interface RawEnvelope {
  protocol: string
  v: number
  type: string
  id: string
  session: string
  nonce: string
  ts: number
}

export function readEnvelope(data: unknown): RawEnvelope | null {
  if (typeof data !== 'object' || data === null || Array.isArray(data)) return null
  const m = data as Record<string, unknown>
  if (typeof m.protocol !== 'string') return null
  if (typeof m.v !== 'number' || !Number.isFinite(m.v)) return null
  if (typeof m.type !== 'string') return null
  if (typeof m.id !== 'string' || m.id === '') return null
  if (typeof m.session !== 'string' || m.session === '') return null
  if (typeof m.nonce !== 'string' || m.nonce === '') return null
  if (typeof m.ts !== 'number' || !Number.isFinite(m.ts)) return null
  return {
    protocol: m.protocol,
    v: m.v,
    type: m.type,
    id: m.id,
    session: m.session,
    nonce: m.nonce,
    ts: m.ts,
  }
}

const encoder = new TextEncoder()

/** 字符串的 UTF-8 字节数。 */
export function utf8Bytes(s: string): number {
  return encoder.encode(s).byteLength
}

/**
 * 一条消息的体量。JSON.stringify 抛错（循环引用）时按“超限”处理 ——
 * 宁可显式拒绝，也不要把一个编不出来的对象排队等 postMessage 去抛。
 */
export function messageBytes(data: unknown): number {
  if (typeof data === 'string') return utf8Bytes(data)
  try {
    const s = JSON.stringify(data)
    return s === undefined ? 0 : utf8Bytes(s)
  } catch {
    return Number.POSITIVE_INFINITY
  }
}

/**
 * 响应里的行数。只认 Host API 的 QueryResponse 形状（rows / rowCount，兼容 proto JSON
 * 的 row_count）；没有这两项就按 0 算，绝不靠猜数组下标。
 */
export function rowCountOf(value: unknown): number {
  if (typeof value !== 'object' || value === null) return 0
  const o = value as Record<string, unknown>
  let n = Array.isArray(o.rows) ? o.rows.length : 0
  const declared = typeof o.rowCount === 'number' ? o.rowCount
    : typeof o.row_count === 'number' ? o.row_count : 0
  if (Number.isFinite(declared) && declared > n) n = declared
  return n
}

/**
 * 单次响应的体量/行数判据。返回 null = 可投递；否则是应当显式回给 iframe 的错误。
 * 行数先判：几千行的小对象也可能不到 1 MiB，但下游按“行”消费会更早出问题。
 */
export function payloadLimitError(value: unknown): BridgeErrorBody | null {
  const rows = rowCountOf(value)
  if (rows > MAX_RESPONSE_ROWS) {
    return {
      code: 'TOO_MANY_ROWS',
      message: `响应有 ${rows} 行，超过单次上限 ${MAX_RESPONSE_ROWS} 行`,
    }
  }
  const bytes = messageBytes(value)
  if (bytes > MAX_MESSAGE_BYTES) {
    return {
      code: 'PAYLOAD_TOO_LARGE',
      message: `响应 ${bytes} 字节，超过单条消息上限 ${MAX_MESSAGE_BYTES} 字节`,
    }
  }
  return null
}

export function isHostRequestMessage(v: unknown): v is HostRequestMessage {
  if (typeof v !== 'object' || v === null) return false
  const m = v as Record<string, unknown>
  return m.type === 'host.request' && typeof m.reqId === 'string' && m.reqId !== ''
    && typeof m.method === 'string' && m.method !== ''
}

export function isRpcRequestMessage(v: unknown): v is RpcRequestMessage {
  if (typeof v !== 'object' || v === null) return false
  const m = v as Record<string, unknown>
  return m.type === 'rpc.request' && typeof m.reqId === 'string' && m.reqId !== ''
    && typeof m.path === 'string' && m.path !== ''
}

export function isCancelMessage(v: unknown): v is CancelMessage {
  if (typeof v !== 'object' || v === null) return false
  const m = v as Record<string, unknown>
  return m.type === 'cancel' && typeof m.reqId === 'string' && m.reqId !== ''
}

export function isErrorMessage(v: unknown): v is ErrorMessage {
  if (typeof v !== 'object' || v === null) return false
  const m = v as Record<string, unknown>
  if (m.type !== 'error') return false
  if (m.reqId !== undefined && typeof m.reqId !== 'string') return false
  const e = m.error
  if (typeof e !== 'object' || e === null) return false
  const body = e as Record<string, unknown>
  return isBridgeErrorCode(body.code) && typeof body.message === 'string'
}
