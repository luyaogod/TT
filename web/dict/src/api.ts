// 字典页的 REST 客户端。
//
// 路径分两类 —— 合并后一个进程同时挂着两套 SPA,后端把字典子系统用
// http.StripPrefix 挂在 /dict 下,于是包内注册的 /api/mirror 对外是 /dict/api/mirror:
//
//   `${API_BASE}/…`  字典页**私有**接口:只剩两个长跑动作(源码镜像、字典同步)
//   `/api/…`         统一层的共享端点(hosts 配置、config/status、install、health)
//
// 配置类的东西(环境/数据库、查询数据源、镜像目录、同步目标、BDL 文档目录、PATH 安装)
// 全部收进了调试工作台里的统一设置页,本页不再读写它们 —— 只剩下这两个动作各自的
// 「跑一次 + 看进度」。共用端点不带前缀;私有调用少了前缀就会落到共享层去。
export const API_BASE = '/dict/api'

// 共享健康检查(GET /api/health):服务是否就绪、版本、配置文件路径、两个子系统是否在线。
// 本页只用它的 debug 字段决定要不要显示「设置」入口。
export interface HealthView {
  ok: boolean
  version: string
  config: string
  debug: boolean
  dict: boolean
}

// 源码镜像(host.MirrorPullProgress 的进度经 /dict/api/mirror 轮询返回)
export interface MirrorEnv {
  name: string
  zone: string
  path: string
  ready: boolean // 本地已有完整基线(增量前提)
}
export interface MirrorJob {
  running: boolean
  env: string
  full: boolean
  phase: string // connect|probe|pack|download|done|error
  message: string
  bytes: number
  total: number // 下载阶段为归档总字节;0=未知(pack 阶段)
  files: number
  elapsed: string
  error?: string
  note?: string
  done: boolean
  startedAt?: string
}
export interface MirrorResp {
  mirrorDir: string // 只读展示:镜像根目录在设置页里改
  activeEnv: string
  envs: MirrorEnv[]
  job: MirrorJob
}

// 数据库同步(dbsync.Run 的进度经 /dict/api/dbsync 轮询返回)
export interface DBSyncEnv {
  name: string
  type: string
  address: string
}
export interface DBSyncJob {
  running: boolean
  env: string
  phase: string // open|table|index|replace|done|error
  message: string
  table: string
  tableIndex: number
  tableTotal: number
  tableRows: number
  totalRows: number
  tables: number
  elapsed: string
  target: string
  backup?: string
  warning?: string
  error?: string
  done: boolean
  startedAt?: string
}
export interface DBSyncResp {
  target: string // 生效目标
  configured: string // 配置里显式的值(空=用默认位置)
  defaultTarget: string
  exists: boolean
  activeEnv: string
  envs: DBSyncEnv[]
  job: DBSyncJob
}

async function req<T>(url: string, opts?: RequestInit): Promise<T> {
  const r = await fetch(url, { headers: { 'Content-Type': 'application/json' }, ...opts })
  const data = await r.json().catch(() => ({}))
  if (!r.ok || data.ok === false) throw new Error(data.error || `HTTP ${r.status}`)
  return data as T
}

export const api = {
  // 运行信息:共享的健康检查(本页只用 debug 字段判断设置入口该不该显示)
  health: () => req<HealthView>('/api/health'),

  // 源码镜像:读现状 + 触发一次拉取(单实例任务,进度在 job 里)
  mirror: () => req<MirrorResp>(`${API_BASE}/mirror`),
  mirrorPull: (env: string, full: boolean) =>
    req<{ ok: boolean }>(`${API_BASE}/mirror/pull`, { method: 'POST', body: JSON.stringify({ env, full }) }),

  // 字典同步:读现状 + 触发一次同步(单实例任务,进度在 job 里)
  dbsync: () => req<DBSyncResp>(`${API_BASE}/dbsync`),
  dbsyncRun: (env: string) =>
    req<{ ok: boolean }>(`${API_BASE}/dbsync`, { method: 'POST', body: JSON.stringify({ env }) }),
}
