// 设置页「插件管理」的纯逻辑：状态、贡献面、授权选择的唯一口径。
//
// 为什么单独一个模块而不是写在组件里：这些判定要被 Node 侧的 check:plugins 直接加载
// （裸打包器、不启浏览器），所以模块顶层不能碰 DOM/fetch。组件只负责把它们画出来。
//
// 这里不复制后端的校验规则：什么能力合法、环境存不存在、只给环境不给能力会不会被拒，
// 都以 /api/plugins* 的返回为准；本模块只决定"提交什么形状"与"怎么读回来"。
import type { PluginAdminSummary, PluginDoctorIssue } from './api'

/** 已知能力 → 一句功能名。未知能力原样显示（后端会拒绝未知授权，这里只是展示）。 */
export const CAPABILITY_LABELS: Record<string, string> = {
  'profiles.list': '读取环境清单',
  'db.connect': '数据库连接',
  'ssh.connect': 'SSH 连接',
}

/** 能力名 → 给人看的标签；不认识的返回原名（宁可露出后端标识，也不猜含义）。 */
export function capabilityLabel(cap: string): string {
  const k = cap.trim()
  return CAPABILITY_LABELS[k] || k
}

/** 去空、去重，保持输入顺序。 */
function uniq(list: string[] | undefined): string[] {
  const out: string[] = []
  for (const raw of list || []) {
    const v = typeof raw === 'string' ? raw.trim() : ''
    if (v && !out.includes(v)) out.push(v)
  }
  return out
}

/** 去空、去重、按码点排序（写进注册表的值要稳定：同一次选择永远得到同一份授权）。 */
function sortedUnique(list: string[] | undefined): string[] {
  return uniq(list).sort()
}

export type PluginTone = 'green' | 'yellow' | 'red' | 'gray'

/**
 * 卡片右上角的状态徽标。
 *
 * 运行态优先于启用态：一个 enabled 的服务插件如果已经是 degraded，"已启用"会骗人。
 * lastError 不塞进这个徽标 —— 它有自己的红色一行（徽标只回答"现在算不算正常"）。
 */
export function pluginStatus(s: PluginAdminSummary): { label: string; tone: PluginTone } {
  if (s.state === 'degraded') return { label: '降级', tone: 'red' }
  if (s.enabled && s.state === 'running') return { label: '运行中', tone: 'green' }
  if (s.enabled) return { label: '已启用', tone: 'green' }
  return { label: '未启用', tone: 'gray' }
}

/** 贡献面（CLI / 服务 / 页面 / 技能）；一条都没有时给一句说明，不留空白。 */
export function contributionLabels(s: PluginAdminSummary): string[] {
  const out: string[] = []
  if (s.namespace) out.push(`CLI tt ${s.namespace}`)
  if (s.hasService) out.push('常驻服务')
  if (s.route) out.push(`页面 #plugin/${s.route}`)
  const skills = uniq(s.skillDirs).length
  if (skills > 0) out.push(`技能 × ${skills}`)
  return out.length > 0 ? out : ['（无贡献）']
}

/** 声明了但还没授予的能力（启用面板里逐条勾选的就是它）。 */
export function ungrantedPermissions(s: PluginAdminSummary): string[] {
  const granted = new Set(uniq(s.capabilities))
  return uniq(s.permissions).filter((p) => !granted.has(p))
}

/** 是否需要在启用面板里选环境：没有任何能力时，环境授权不会单独生效（后端会拒）。 */
export function needsEnvSelection(capabilities: string[]): boolean {
  return uniq(capabilities).length > 0
}

/**
 * 启用请求体：去重 + 排序；没有能力时不带环境。
 *
 * 这一条与后端 validateGrants 对齐：只给环境不给能力会被拒，与其让用户点了确认再看到
 * 一句错误，不如在提交前就把"单独的环境"丢掉（界面上也把环境勾选置灰）。
 */
export function grantsPayload(capabilities: string[], envs: string[]): { capabilities: string[]; envs: string[] } {
  const caps = sortedUnique(capabilities)
  return { capabilities: caps, envs: caps.length > 0 ? sortedUnique(envs) : [] }
}

/** 自检问题排序：error 在前，同级保持后端顺序（稳定排序）。 */
export function sortIssues(issues: PluginDoctorIssue[]): PluginDoctorIssue[] {
  return issues.slice().sort((a, b) => (a.severity === 'error' ? 0 : 1) - (b.severity === 'error' ? 0 : 1))
}

/** 自检结论的一句话。 */
export function doctorSummary(count: number): string {
  return count === 0 ? '没有发现问题' : `${count} 个问题`
}

/** 日志行数说明：尾部被截断时必须说出来，别让人以为日志只有这么长。 */
export function logsSummary(r: { count: number; truncated?: boolean }): string {
  return r.truncated ? `显示最新 ${r.count} 行（更早的已截断）` : `${r.count} 行`
}
