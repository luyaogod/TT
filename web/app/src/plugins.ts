// 插件前端贡献的运行时投影，以及插件页面/桥路径的边界规则。
//
// 主 SPA **不静态 import 任何第三方代码**：插件页面只在 iframe 里从
// /plugins/<id>/ui/ 加载。这里拿到的只是后端注册表的摘要（GET /api/plugins），
// 所以模块本身是纯函数 + 常量，能被 Node 侧检查项直接加载。
//
// 两个数据来源不要混：
//   · 后端 Summary 里 route/routeTitle 是“这个插件声明了哪个页面”；
//   · 本模块把它压成 PluginDescriptor，并丢掉没有页面贡献（没有 route）的条目。
//     插件管理（安装/启停）是设置页的事，不在外壳菜单里做。

import { isPluginRoute } from './routing'

/** 宿主支持的槽位（与后端 plugin.KnownUISlots 同一份口径）。 */
export const KNOWN_UI_SLOTS = ['nav', 'page', 'header.actions', 'settings.section', 'statusbar', 'overlay'] as const
export type PluginUISlot = (typeof KNOWN_UI_SLOTS)[number]

/** 一条同页 UI 贡献（清单 contributions.ui 的运行时投影）。 */
export interface PluginUIEntry {
  slot: PluginUISlot
  id: string
  route: string
  title: string
  icon: string
  order: number
  /** 包内相对路径（如 client.js）；由 pluginUIAssetURL 拼成可 import 的地址。 */
  entry: string
}

/** 一条前端贡献的运行时描述。健康状态可选：后端暂时只给 lastError 时由它推导。 */
export interface PluginDescriptor {
  id: string
  name: string
  version: string
  enabled: boolean
  /** 外壳路由名（#plugin/<route>），后端已保证小写/数字/连字符。 */
  route: string
  /** 菜单与页面标题；缺省按 name → route 兜底。 */
  title: string
  /** 同页 UI 贡献；空 = 旧式 iframe 页面。 */
  ui: PluginUIEntry[]
  /** undefined = 后端没说；false = 明确不健康（lastError 非空也按 false 处理）。 */
  healthy?: boolean
  lastError?: string
}

/** 一条 ui 贡献的原始形状（只列本模块关心的字段）。 */
export interface PluginSummaryUI {
  slot?: string
  id?: string
  route?: string
  title?: string
  icon?: string
  order?: number
  entry?: string
}

/** GET /api/plugins 的一条摘要（只列本模块关心的字段，其余忽略）。 */
export interface PluginSummary {
  id: string
  name?: string
  version?: string
  enabled?: boolean
  namespace?: string
  route?: string
  routeTitle?: string
  ui?: PluginSummaryUI[]
  healthy?: boolean
  lastError?: string
}

/** GET /api/plugins 的应答。 */
export interface PluginsReply {
  ok: boolean
  count?: number
  plugins: PluginSummary[]
}

/**
 * 插件 id 的形态：反向 DNS（至少两段、小写字母/数字/连字符）。
 * 它会被拼进 /plugins/<id>/…，所以**先校验再拼**，不把后端字段直接当路径用。
 */
const PLUGIN_ID_RE = /^[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,62}[a-z0-9])?)+$/

export function isPluginID(v: string): boolean {
  return v.length > 0 && v.length <= 190 && PLUGIN_ID_RE.test(v)
}

/** 一条摘要 → 描述符；id 必须合法；有 page 时 route 必须合法，只有 nav 等槽位时允许无 route。 */
function toDescriptor(raw: PluginSummary): PluginDescriptor | null {
  if (!raw || typeof raw !== 'object') return null
  const id = typeof raw.id === 'string' ? raw.id.trim() : ''
  const route = typeof raw.route === 'string' ? raw.route.trim() : ''
  if (!isPluginID(id)) return null
  // 没有 page 的插件（只贡献 nav/statusbar…）没有 route；有 page 时 route 必须合法。
  const uiAll = normalizeUIEntries(raw.ui)
  const hasPage = uiAll.some((u) => u.slot === 'page')
  if (hasPage && !isPluginRoute(route)) return null
  if (!hasPage && route !== '' && !isPluginRoute(route)) return null
  const name = typeof raw.name === 'string' && raw.name.trim() ? raw.name.trim() : id
  const title = (typeof raw.routeTitle === 'string' && raw.routeTitle.trim()) || name || route
  const version = typeof raw.version === 'string' ? raw.version : ''
  const enabled = raw.enabled === true
  const ui = uiAll
  // 健康：后端给了布尔就用；没给但有 lastError，按“明确不健康”处理（错误是事实）。
  const healthy = typeof raw.healthy === 'boolean' ? raw.healthy : (raw.lastError ? false : undefined)
  return {
    id, name, version, enabled, route, title, ui, healthy,
    lastError: typeof raw.lastError === 'string' && raw.lastError ? raw.lastError : undefined,
  }
}

/** 包内相对路径：不接受绝对路径、`..`、空段与 NUL。 */
function normalizeEntryPath(raw: unknown): string | null {
  if (typeof raw !== 'string') return null
  const s = raw.trim().replace(/\\/g, '/')
  if (!s || s.startsWith('/') || s.includes('\0')) return null
  const parts = s.split('/')
  if (parts.some((p) => p === '' || p === '.' || p === '..')) return null
  return parts.join('/')
}

/** 逐条校验 ui 贡献；坏条目直接丢（不猜、不把半截路径塞给 import）。 */
function normalizeUIEntries(raw: PluginSummaryUI[] | undefined): PluginUIEntry[] {
  if (!Array.isArray(raw)) return []
  const out: PluginUIEntry[] = []
  for (const item of raw) {
    if (!item || typeof item !== 'object') continue
    const slot = typeof item.slot === 'string' ? item.slot.trim() : ''
    if (!(KNOWN_UI_SLOTS as readonly string[]).includes(slot)) continue
    const route = typeof item.route === 'string' ? item.route.trim() : ''
    if (route && !isPluginRoute(route)) continue
    const entry = normalizeEntryPath(item.entry)
    // page 与组件类槽位必须有入口；nav/settings.section 可以只有标题。
    if (!entry && (slot === 'page' || slot === 'header.actions' || slot === 'statusbar' || slot === 'overlay')) continue
    out.push({
      slot: slot as PluginUISlot,
      id: typeof item.id === 'string' ? item.id.trim() : '',
      route,
      title: typeof item.title === 'string' ? item.title.trim() : '',
      icon: normalizeEntryPath(item.icon) ?? '',
      order: typeof item.order === 'number' && Number.isFinite(item.order) ? item.order : 0,
      entry: entry ?? '',
    })
  }
  return out
}

/** 插件的页面贡献（同页渲染的入口）；没有就返回 null（走 legacy iframe）。 */
export function pluginPageEntry(plugin: PluginDescriptor): PluginUIEntry | null {
  return plugin.ui.find((u) => u.slot === 'page' && u.route === plugin.route) ?? null
}

/** ui 贡献的浏览器模块地址：/plugins/<id>/ui/<entry>。参数不合法返回 null。 */
export function pluginUIAssetURL(id: string, entry: string): string | null {
  if (!isPluginID(id)) return null
  const e = normalizeEntryPath(entry)
  return e ? `/plugins/${id}/ui/${e}` : null
}

/** 把应答压成菜单/页面能直接用的描述符清单。 */
export function normalizePlugins(input: PluginsReply | PluginSummary[] | null | undefined): PluginDescriptor[] {
  const list = Array.isArray(input) ? input : input?.plugins
  if (!Array.isArray(list)) return []
  const out: PluginDescriptor[] = []
  for (const raw of list) {
    const d = toDescriptor(raw)
    if (d) out.push(d)
  }
  return out
}

/**
 * 菜单顺序：按标题排；标题相同再按 route、最后按 id。
 *
 * 用码点比较而不是 localeCompare：Node 检查项与浏览器必须得到同一个顺序，
 * 而 locale 数据在不同运行环境里可能不同。稳定比“像拼音”更重要。
 */
function compareDescriptors(a: PluginDescriptor, b: PluginDescriptor): number {
  if (a.title !== b.title) return a.title < b.title ? -1 : 1
  if (a.route !== b.route) return a.route < b.route ? -1 : 1
  if (a.id !== b.id) return a.id < b.id ? -1 : 1
  return 0
}

/** 外壳菜单条目：只含已启用、且确有 page 贡献的插件，顺序稳定。 */
export function pluginMenuItems(plugins: PluginDescriptor[]): PluginDescriptor[] {
  return plugins
    .filter((p) => p.enabled && isPluginID(p.id) && p.ui.some((u) => u.slot === 'page'))
    .slice()
    .sort(compareDescriptors)
}

/** 插件页面的路由解析结果。error 必须落到显式错误页，绝不回退到别的页面。 */
export type PluginPageErrorReason = 'malformed' | 'unknown' | 'disabled' | 'invalid'

export type PluginPageResolution =
  | { kind: 'ok'; plugin: PluginDescriptor }
  | { kind: 'error'; route: string; reason: PluginPageErrorReason; plugin?: PluginDescriptor }

export function resolvePluginPage(route: string, plugins: PluginDescriptor[]): PluginPageResolution {
  const plugin = plugins.find((p) => p.route === route)
  if (!plugin) return { kind: 'error', route, reason: 'unknown' }
  if (!isPluginID(plugin.id)) return { kind: 'error', route, reason: 'invalid', plugin }
  if (!plugin.enabled) return { kind: 'error', route, reason: 'disabled', plugin }
  return { kind: 'ok', plugin }
}

/** 插件 iframe 的地址。id 不合法时返回 null（调用方渲染错误页，不拼一个半截路径）。 */
export function pluginUIPath(id: string): string | null {
  return isPluginID(id) ? `/plugins/${id}/ui/` : null
}

// ---------------------------------------------------------------------------
// 管理面拒绝清单
// ---------------------------------------------------------------------------

/**
 * 插件消息桥**永远不代理**的宿主管理端点。
 *
 * 为什么还要单独一张表：桥只该做两件事 —— 走 /hostrpc/ 的 Host API、走
 * /plugins/<id>/rpc/ 的插件自身服务。任何形状像管理面的路径都是配置错误或越权尝试；
 * 显式列出来，测试与代码审查才有同一个口径。
 */
export const DENIED_ADMIN_PATHS: readonly string[] = [
  '/api/plugins',
  '/api/hosts',
  '/api/update',
  '/api/shutdown',
  '/api/skills',
  '/api/install',
  '/api/pickdir',
  '/api/fs/list',
  '/api/config',
  '/api/cache',
  '/api/dbprobe',
  '/api/conntest',
  '/api/sshtest',
]

function normalizePathForDeny(path: string): string {
  let p = path.trim()
  const cut = p.search(/[?#]/)
  if (cut >= 0) p = p.slice(0, cut)
  if (!p.startsWith('/')) p = '/' + p
  p = p.replace(/\/{2,}/g, '/').toLowerCase()
  if (p.length > 1 && p.endsWith('/')) p = p.slice(0, -1)
  return p
}

/**
 * 这个路径是否落在管理面上。按**段边界**匹配：/api/hosts 命中，
 * /api/hosts/export 也命中；/api/hostsfoo 不命中（它不是那个端点）。
 * 另外 `/api` 整棵树一律算管理面 —— 桥没有理由请求它。
 */
export function isDeniedAdminPath(path: string): boolean {
  const p = normalizePathForDeny(path)
  if (p === '/api' || p.startsWith('/api/')) return true
  return DENIED_ADMIN_PATHS.some((d) => p === d || p.startsWith(d + '/'))
}

// ---------------------------------------------------------------------------
// 两条受控桥路径
// ---------------------------------------------------------------------------

/** Host API 的 ConnectRPC 方法与服务名白名单（v1）。 */
const HOST_API_METHOD_RE = /^tt\.host\.v1\.(HostService|ProfilesService|ConnectionsService|DBSessionService|SSHSessionService)\/[A-Za-z][A-Za-z0-9]*$/

/** 校验 Host API 方法名；不合法返回 null。 */
export function normalizeHostApiMethod(raw: string): string | null {
  const method = raw.trim()
  if (!HOST_API_METHOD_RE.test(method)) return null
  if (isDeniedAdminPath(method)) return null
  return method
}

/**
 * Host API 的桥路径：/plugins/<id>/hostrpc/<service>/<method>。
 *
 * **不是** /api/：父页面把请求交给这个桥，由后端用该插件的实例令牌去 /hostrpc/ 调用。
 * 令牌由后端签发与持有，插件页面永远看不到它。
 */
export function pluginHostApiBridgePath(pluginId: string, method: string): string | null {
  if (!isPluginID(pluginId)) return null
  const m = normalizeHostApiMethod(method)
  if (!m) return null
  const url = `/plugins/${pluginId}/hostrpc/${m}`
  return isDeniedAdminPath(url) ? null : url
}

const RPC_PATH_RE = /^\/[A-Za-z0-9._~/-]*$/

/**
 * 插件 RPC 路径：相对插件自身服务的路径（后端把它转发到该实例）。
 *
 * 拒绝：绝对 URL、`//host`、`.`/`..` 段、以及 /api/… 形状 ——
 * 最后一条是刻意的：桥不该出现任何“像在代理宿主管理面”的路径。
 */
export function normalizePluginRpcPath(raw: string): string | null {
  let p = raw.trim()
  if (!p.startsWith('/')) p = '/' + p
  if (p.includes('//')) return null
  if (!RPC_PATH_RE.test(p)) return null
  if (p.split('/').some((seg) => seg === '.' || seg === '..')) return null
  if (isDeniedAdminPath(p)) return null
  return p
}

/** 插件自身服务的 RPC 桥路径：/plugins/<id>/rpc/<suffix>。 */
export function pluginRPCBridgePath(pluginId: string, path: string): string | null {
  if (!isPluginID(pluginId)) return null
  const p = normalizePluginRpcPath(path)
  if (!p) return null
  const url = `/plugins/${pluginId}/rpc${p}`
  return isDeniedAdminPath(url) ? null : url
}
