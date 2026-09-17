// tdict 配置服务 REST 客户端
//
// 路径分两类 —— 合并后一个进程同时挂着两套 SPA,后端把字典子系统用
// http.StripPrefix 挂在 /dict 下,于是包内注册的 /api/mirror 对外是 /dict/api/mirror:
//
//   `${API_BASE}/…`  字典页**私有**接口(镜像、数据同步、PATH 安装、BDL 文档)
//   `/api/…`         两套页面**共用**的顶层端点(hosts 配置、dbprobe、conntest、health)
//
// 共用端点不带前缀;私有调用少了前缀就会落到共享层或调试页去。私有端点必须走
// API_BASE,不要在调用处手写字符串。
export const API_BASE = '/dict/api'

export interface DBAcct { account: string; password: string }
export interface DBConnection {
  type: string // oracle | kingbase
  host: string
  port: number
  service?: string // oracle: SERVICE_NAME
  database?: string // kingbase: 库名
  accounts?: DBAcct[]
  // 是否允许 AI 执行只读 SQL。undefined = 未配置 = 开启;显式 false 才关。
  // **载入/保存都必须带上它**:它由调试工作台的环境页维护,本页没有对应控件,
  // 但两页写的是同一个 hosts 节 —— 丢掉它就等于把用户关掉的开关悄悄打开。
  readonlySql?: boolean
}
export interface SshEnv {
  name: string
  host: string
  port: number
  user: string
  password: string
  zone?: string
  topent?: string
  db?: DBConnection | null
}
// 调试工作台的设置(debug 节):本页只读不写,只为把 GET /api/hosts 拆干净
export interface DebugSettingsView {
  activeEnv?: string
  launchArgs?: string
  watchdogSeconds?: number
  fglserver?: string
  termWidth?: number
  termHeight?: number
  printElements?: number
  persistBreakpoints?: boolean
}
// 共享配置快照:GET /api/hosts 的响应。
// 两套配置页读的是这同一份(环境清单只有 config.json 的 hosts 节一个数据源),
// 与 tt/internal/config 的 schema.go 一一对应。
export interface HostsView {
  config: string
  activeEnv: string
  sshs: SshEnv[]
  listen: string // 监听地址在顶层(合并前它藏在调试页的 debug.listen 里)
  debug: DebugSettingsView
  query: { source?: string } // ""(缺省/在线默认环境) | "local" | <环境名>
  mirror: { dir?: string }
  bdldoc: { dir?: string }
  sync: { target?: string }
  tdev: { workspaceSuffix?: string; defaultOut?: string }
  version: string
}
// PUT /api/hosts 的请求体:只带要改的节,省略的节后端保持原样 ——
// 两套页面各改各的部分时不会用陈旧快照覆盖对方刚改好的节。
export interface HostsPatch {
  hosts?: { activeEnv: string; sshs: SshEnv[] }
  listen?: string
  debug?: DebugSettingsView
  query?: { source: string }
  mirror?: { dir?: string }
  bdldoc?: { dir?: string }
  sync?: { target?: string }
  tdev?: { workspaceSuffix?: string; defaultOut?: string }
}
// 共享健康检查(GET /api/health):服务是否就绪、版本、配置文件路径、两个子系统是否在线
export interface HealthView {
  ok: boolean
  version: string
  config: string
  debug: boolean
  dict: boolean
}
// 服务器侧探测结果(host.ProbeDBConfig)
export interface DBProbeOut {
  type: string
  tns?: string
  port?: number
  database?: string
  sqlplus?: string
  oracleHome?: string
  twoTask?: string
  host?: string
  service?: string
  note?: string
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
  mirrorDir: string
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
  target: string
  configured: string
  defaultTarget: string
  exists: boolean
  activeEnv: string
  envs: DBSyncEnv[]
  job: DBSyncJob
}

// 命令行安装:把可执行文件所在目录加入用户 PATH
export interface InstallStatus {
  supported: boolean // 本平台是否支持自动写入
  exePath: string
  exeDir: string
  inUserPath: boolean
  userPath?: string
  manual?: string
  note?: string
}

// BDL(4GL)语言文档目录(等价 tdict bdldoc dir)
export interface BdldocStatus {
  ok: boolean
  dir: string
  exists: boolean
  configPath: string
}

async function req<T>(url: string, opts?: RequestInit): Promise<T> {
  const r = await fetch(url, { headers: { 'Content-Type': 'application/json' }, ...opts })
  const data = await r.json().catch(() => ({}))
  if (!r.ok || data.ok === false) throw new Error(data.error || `HTTP ${r.status}`)
  return data as T
}

export const api = {
  // 运行信息:共享的健康检查。合并前这里是本包的 GET /api/status,那个端点已经没了
  // (listen 从 hosts 拿 —— 它属于共享配置)。
  health: () => req<HealthView>('/api/health'),
  // 配置元信息:配置文件缺省落点、便携标记、支持的库类型…
  configMeta: () => req<{
    config: string; defaultConfig: string; portable: boolean; toolsHome: string
    schemaVersion: number; defaultListen: string; legacyTools: string[] | null; supportedTypes: string[]
  }>('/api/config/meta'),
  // 环境清单:共享端点(顶层)。合并前这里是本包私有的 GET /api/config,
  // 合并后 hosts 节只有这一份读写入口,调试页的环境页读写的也是它。
  hosts: () => req<HostsView>('/api/hosts'),
  // 保存:只发要改的节(hosts / query / bdldoc / …),省略的节后端保持原样
  saveHosts: (cfg: HostsPatch) => req<HostsView>('/api/hosts', { method: 'PUT', body: JSON.stringify(cfg) }),
  // 服务器侧探测连接要素(登录该环境 SSH 只读执行;note 说明未获取到的原因)
  probeDB: (body: { host: string; port: number; user: string; password: string; zone: string; type: string }) =>
    req<DBProbeOut>('/api/dbprobe', { method: 'POST', body: JSON.stringify(body) }),
  // 客户端直连测试(凭据取账号列表首项)。共享端点要求 {connection: …} 包一层
  // (与调试页同形),不再是平铺连接字段。
  connTest: (connection: DBConnection) =>
    req<{ ok: boolean; type?: string; serverVersion?: string; address?: string; stage?: string; error?: string }>(
      '/api/conntest', { method: 'POST', body: JSON.stringify({ connection }) }),
  // 账号清单「验证」:SSH 上服务器以该账号+密码连显式目标库 select 1(只读)
  dbAccVerify: (body: {
    host: string; port: number; user: string; password: string; zone: string; type: string
    account: string; acctPassword: string; dbHost?: string; dbPort?: number; dbSvc?: string; dbDatabase?: string
  }) => req<{ ok: boolean; error?: string }>('/api/dbaccverify', { method: 'POST', body: JSON.stringify(body) }),
  // 源码镜像:根目录读写 + 拉取任务(单实例)
  mirror: () => req<MirrorResp>(`${API_BASE}/mirror`),
  saveMirrorDir: (dir: string) =>
    req<{ ok: boolean; mirrorDir?: string }>(`${API_BASE}/mirror`, { method: 'PUT', body: JSON.stringify({ dir }) }),
  mirrorPull: (env: string, full: boolean) =>
    req<{ ok: boolean }>(`${API_BASE}/mirror/pull`, { method: 'POST', body: JSON.stringify({ env, full }) }),
  // 数据库同步:远程库字典 → 本地 SQLite(单实例任务)
  dbsync: () => req<DBSyncResp>(`${API_BASE}/dbsync`),
  dbsyncRun: (env: string) =>
    req<{ ok: boolean }>(`${API_BASE}/dbsync`, { method: 'POST', body: JSON.stringify({ env }) }),
  // 设置同步目标(config.json 顶层 sync.target;空串=清除,回到默认 exe 同目录)
  saveDBSyncTarget: (target: string) =>
    req<DBSyncResp>(`${API_BASE}/dbsync`, { method: 'PUT', body: JSON.stringify({ target }) }),
  // 命令行安装:查看/加入/移出用户 PATH(用户级,无需管理员)
  installStatus: () => req<InstallStatus>(`${API_BASE}/install`),
  installAdd: () => req<InstallStatus>(`${API_BASE}/install`, { method: 'POST' }),
  installRemove: () => req<InstallStatus>(`${API_BASE}/install`, { method: 'DELETE' }),
  // BDL 语言文档目录。读取仍走本包的 /dict/api/bdldoc —— 它比共享的 hosts 多返回
  // 「目录是否存在于本机」(exists),设置页靠它给提示;写入则统一走 PUT /api/hosts
  // 的顶层 bdldoc 节(与 query 一样,本包不再有 PUT /api/bdldoc 的写入入口)。
  bdldoc: () => req<BdldocStatus>(`${API_BASE}/bdldoc`),
  saveBdldocDir: async (dir: string) => {
    await req<HostsView>('/api/hosts', { method: 'PUT', body: JSON.stringify({ bdldoc: { dir } }) })
    return req<BdldocStatus>(`${API_BASE}/bdldoc`)
  },
}
