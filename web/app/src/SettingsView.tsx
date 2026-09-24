// 统一设置页:左侧四个分区的导航,右侧按分区渲染卡片。
//
// 「站点管理」环境清单 + SSH/数据库(原调试页「环境」节与原字典页「环境配置」的并集,只留一份);
// 「数据字典」查询数据源 / 源码镜像 / 数据同步 / BDL 文档;
// 「DEBUG」调试参数 / 默认环境;
// 「应用设置」外观 / 服务 / 运行信息。
//
// 环境清单走**共享**端点 /api/hosts(顶层,不带 API_BASE 前缀):一个进程里它是唯一的数据源
// —— config.json 的 hosts 节。合并前这份清单藏在 debug 节里
// (所以老代码收发的是整个 debug 节点),现在收发的是 hosts 节。
//
// 保存**按节提交**:设置页的各张卡片各改各的节,不拿陈旧快照覆盖对方刚改好的节
// (后端约定:PUT /api/hosts 里省略的节保持原样)。
//
// ⚠ 后端的 PUT 是**整节替换**,而这些结构体每个字段都带 omitempty —— 省略的键等于删除。
// 所以每张卡片的保存都必须提交**完整的节**(从读到的快照 + 编辑项构造),绝不能只发表单里
// 那几个字段:老代码在「高级」里只发了 5 个键,于是 fglserver / persistBreakpoints /
// activeEnv 在用户每次保存时被静默抹掉。debugPatchOf 就是为了堵这个。
import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react'
import {
  AlertCircle, CheckCircle2, Download, Eye, EyeOff,
  Monitor, Plus, RefreshCw, RotateCcw, Star, Sun, Moon,
  Terminal, Trash2, Zap,
} from 'lucide-react'
import {
  api, type ConfigMeta, type ConfigStatus, type DBSyncJob, type DBSyncResp,
  type HostsDb, type HostsPatch, type HostsSsh, type HostsView, type MirrorJob, type MirrorResp,
} from './api'
import {
  Button, Input, Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '../../shared/ui'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '../../shared/ui-radix'
import { Card, Field, InfoRow, SettingRow } from '../../shared/settings'
import { useStore, type ThemeMode } from './store'
import { cn } from '../../shared/utils'
import { SETTINGS_SECTIONS, settingsHash } from './routing'

// ---------- 表单模型 ----------

interface SshDb {
  type: string // oracle | kingbase
  host: string
  port: number
  service: string // oracle: SERVICE_NAME
  database: string // kingbase: 库名
  accounts: { account: string; password: string }[]
  // 是否允许 AI 执行只读 SQL(排查业务数据)。undefined = 未配置 = 开启;
  // 显式 false 才关。**载入/保存都必须带上它** —— 否则在设置页点一次保存,
  // 这个开关就被静默抹掉、悄悄变回开启(我实现时实测过)。
  readonlySql?: boolean
}

interface SshItem {
  name: string
  host: string
  port: number
  user: string
  password: string
  zone: string
  topent: string
  db: SshDb | null
}

// debug 节的完整表单。字段与 internal/config/schema.go 的 DebugSettings 一一对应 ——
// 少一个键就意味着保存时把它删了(整节替换),所以这里必须是全集。
interface DebugForm {
  activeEnv: string // 留空 = 继承 hosts.activeEnv
  launchArgs: string
  watchdogSeconds: number
  fglserver: string
  termWidth: number
  termHeight: number
  printElements: number
  // *bool 三态:nil = 跟随默认(开启)。复选框写不了三态,所以用下拉。
  persistBreakpoints: 'default' | 'on' | 'off'
}

interface DictForm {
  querySource: string // '' | 'local' | <环境名>
  mirrorDir: string
  syncTarget: string
  bdldocDir: string
}

/** 可保存的配置节。脏标记与保存按钮都按它划分。 */
type ConfigKey = 'hosts' | 'debug' | 'listen' | 'query' | 'mirror' | 'sync' | 'bdldoc'

// 主题切换卡片(shadcn 主题切换卡样式):三张卡各带一张迷你界面预览,选中卡描边+底色高亮。
const THEME_OPTIONS: { value: ThemeMode; label: string; icon: typeof Sun }[] = [
  { value: 'light', label: '亮色', icon: Sun },
  { value: 'dark', label: '暗色', icon: Moon },
  { value: 'system', label: '跟随系统', icon: Monitor },
]

// 迷你界面:左侧栏条 + 内容卡(用被预览主题的配色,尺寸固定,只示意明暗关系)
function ThemePreview({ bg, bar, panel }: { bg: string; bar: string; panel: string }) {
  return (
    <span className="flex h-full w-full items-stretch gap-1 p-1.5" style={{ background: bg }}>
      <span className="h-full w-1.5 shrink-0 rounded-[2px]" style={{ background: bar }} />
      <span className="flex h-full min-w-0 flex-1 flex-col justify-center gap-1 rounded-[2px]" style={{ background: panel }}>
        <span className="mx-1 h-1 rounded-full" style={{ background: bar }} />
        <span className="mx-1 h-1 w-2/3 rounded-full" style={{ background: bar }} />
      </span>
    </span>
  )
}

const LIGHT_PREVIEW = { bg: '#ffffff', bar: '#d4d4d4', panel: '#f4f4f4' }
const DARK_PREVIEW = { bg: '#1f1f1f', bar: '#4a4a4a', panel: '#2b2b2b' }

function ThemeCards({ value, onChange }: { value: ThemeMode; onChange: (t: ThemeMode) => void }) {
  return (
    <div className="grid max-w-md grid-cols-3 gap-2">
      {THEME_OPTIONS.map((o) => {
        const Icon = o.icon
        const active = value === o.value
        return (
          <button
            key={o.value}
            type="button"
            aria-pressed={active}
            title={`切换到「${o.label}」`}
            onClick={() => onChange(o.value)}
            className={cn('flex flex-col gap-2 rounded-md border p-2 text-left transition-colors',
              active ? 'border-foreground/50 bg-accent/60' : 'border-border hover:bg-accent/40')}
          >
            <span className="flex h-10 w-full overflow-hidden rounded border border-border/70">
              {o.value === 'light' && <ThemePreview {...LIGHT_PREVIEW} />}
              {o.value === 'dark' && <ThemePreview {...DARK_PREVIEW} />}
              {o.value === 'system' && (
                <>
                  <span className="flex h-full w-1/2"><ThemePreview {...LIGHT_PREVIEW} /></span>
                  <span className="flex h-full w-1/2 border-l border-border/70"><ThemePreview {...DARK_PREVIEW} /></span>
                </>
              )}
            </span>
            <span className={cn('flex items-center gap-1.5 text-xs', active ? 'text-foreground' : 'text-muted-foreground')}>
              <Icon className="h-3.5 w-3.5" />
              {o.label}
            </span>
          </button>
        )
      })}
    </div>
  )
}

// 长跑任务的进度块:镜像拉取与字典同步的阶段语义不同,样式与骨架相同。
function JobProgress({ env, phaseLabel, pct, indeterminate, elapsed, error, message, children }: {
  env: string; phaseLabel: string; pct: number; indeterminate: boolean; elapsed: string
  error?: string; message?: string; children: ReactNode
}) {
  const failed = !!error
  return (
    <div className="border border-border bg-muted/20 p-2">
      <div className="mb-1.5 flex items-center gap-2 text-xs">
        <span className="font-medium">{env}</span>
        <span className="text-muted-foreground">{phaseLabel}</span>
        <span className="ml-auto text-[11px] text-muted-foreground">{elapsed}</span>
      </div>
      {/* 轨道与填充成对用 token:只改填充会让暗色下的轨道消失 */}
      <div className="h-2 w-full overflow-hidden bg-muted">
        {indeterminate
          ? <div className="bar-indeterminate h-full w-1/3 bg-primary" />
          : <div className={cn('h-full transition-[width] duration-300', failed ? 'bg-destructive' : 'bg-primary')} style={{ width: pct + '%' }} />}
      </div>
      <div className="mt-1.5 grid grid-cols-3 gap-2 text-[11px] text-muted-foreground">{children}</div>
      {message && (
        <p className={cn('mt-1.5 flex items-center gap-1 text-xs',
          failed ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400')}>
          {failed ? <AlertCircle className="h-3.5 w-3.5" /> : <CheckCircle2 className="h-3.5 w-3.5" />}
          {error || message}
        </p>
      )}
    </div>
  )
}

// 阶段标签与字节格式化:原先在字典页的两个视图里,随功能一起搬进来。
const MIRROR_PHASE: Record<string, string> = {
  connect: '连接服务器…', probe: '探测 T100 目录…', pack: '服务器打包…',
  download: '下载并解压…', done: '完成', error: '失败',
}
const SYNC_PHASE: Record<string, string> = {
  open: '连接远程数据库…', table: '拉取数据表…', index: '创建主键索引…',
  replace: '替换本地数据库…', done: '完成', error: '失败',
}
function fmtBytes(n: number): string {
  if (!n) return '0 B'
  if (n >= 1 << 30) return (n / (1 << 30)).toFixed(2) + ' GB'
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB'
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(1) + ' KB'
  return n + ' B'
}

const input = 'h-7 text-xs'
const cell = 'h-7 w-full min-w-0 text-xs'
// 单元格「正在编辑」提示环:用 ring 语义 token,不写死颜色(共享层里也只有这一份)
const CELL_EDITING = 'focus-within:ring-1 focus-within:ring-inset focus-within:ring-ring/60'

const blankDb = (): SshDb => ({ type: 'oracle', host: '', port: 1521, service: '', database: '', accounts: [] })
// 新增环境默认带三个常用账号(账号=密码),省去手工录入
const defaultAccounts = () => ['ds', 'dsdata', 'dsdemo'].map((a) => ({ account: a, password: a }))
const blankSsh = (): SshItem => ({
  name: '', host: '', port: 22, user: '', password: '', zone: '36', topent: '',
  db: { ...blankDb(), accounts: defaultAccounts() },
})

// ---------- 表单 ↔ 配置节的互转 ----------

/**
 * 环境清单 → hosts 节。**整节提交**,所以这里必须给出完整的 sshs 列表。
 *
 * 仍未覆盖的字段:`hosts.sshs[].launchArgs` / `watchdogSeconds`(每环境的调试覆盖项,
 * 本页不暴露)。它们由服务端的保留机制补回(见 internal/web 的 preserveUngovernedDBFields)。
 * 纯客户端没法保住它们 —— 表单模型里根本没有这两个字段。
 */
function hostsPatchOf(list: SshItem[], prevActiveEnv: string) {
  const sshsOut: HostsSsh[] = list.filter((x) => x.host).map((x) => {
    const o: HostsSsh = {
      name: x.name || `${x.host}-${x.zone}`.replace(/-$/, ''),
      host: x.host, port: x.port || 22, user: x.user, password: x.password,
    }
    if (x.zone.trim()) o.zone = x.zone.trim()
    if (x.topent.trim()) o.topent = x.topent.trim()
    if (x.db) {
      const db: HostsDb = { type: x.db.type || 'oracle', host: x.db.host.trim(), port: x.db.port || 0 }
      if (x.db.type === 'oracle') { if (x.db.service.trim()) db.service = x.db.service.trim() }
      else { if (x.db.database.trim()) db.database = x.db.database.trim() }
      const accounts = x.db.accounts.map((a) => ({ account: a.account.trim(), password: a.password })).filter((a) => a.account)
      if (accounts.length) db.accounts = accounts
      // 只有"显式关掉"才写:没配 = 默认开启,不必往 JSON 里塞一堆 true
      if (x.db.readonlySql === false) db.readonlySql = false
      o.db = db
    }
    return o
  })
  // 默认环境本页只在「环境清单」里改它,但 PUT 提交的是**整节**,必须原样回传:
  // 若它刚好被这次删除带走了,就顺延到首条 —— 否则后端会以「默认环境不在环境列表里」整次 400 拒掉。
  return {
    activeEnv: sshsOut.some((e) => e.name === prevActiveEnv) ? prevActiveEnv : (sshsOut[0]?.name || ''),
    sshs: sshsOut,
  }
}

/**
 * debug 节 → 完整节对象。
 *
 * **这里是那个数据丢失 bug 的修法**:不按表单里"用户改过哪些"挑选,而是把 debug 节的
 * 每个键都给出来(空/默认的键省略 = 恢复该键的默认值)。老代码只发 5 个键,
 * 于是 fglserver / persistBreakpoints / activeEnv 每次保存都被删掉。
 */
function debugPatchOf(d: DebugForm): NonNullable<HostsPatch['debug']> {
  const out: NonNullable<HostsPatch['debug']> = {}
  if (d.launchArgs) out.launchArgs = d.launchArgs
  if (d.watchdogSeconds) out.watchdogSeconds = d.watchdogSeconds
  if (d.fglserver.trim()) out.fglserver = d.fglserver.trim()
  if (d.termWidth) out.termWidth = d.termWidth
  if (d.termHeight) out.termHeight = d.termHeight
  if (d.printElements) out.printElements = d.printElements
  if (d.persistBreakpoints === 'on') out.persistBreakpoints = true
  else if (d.persistBreakpoints === 'off') out.persistBreakpoints = false
  if (d.activeEnv) out.activeEnv = d.activeEnv
  return out
}

export function SettingsView() {
  const section = useStore((s) => s.settingsSection)
  const setSection = useStore((s) => s.setSettingsSection)
  const theme = useStore((s) => s.theme)
  const setTheme = useStore((s) => s.setTheme)

  const [envTab, setEnvTab] = useState<'ssh' | 'db'>('ssh')
  const [cfg, setCfg] = useState<HostsView | null>(null)
  const [sshs, setSshs] = useState<SshItem[]>([])
  const [dbg, setDbg] = useState<DebugForm>({
    activeEnv: '', launchArgs: '', watchdogSeconds: 0, fglserver: '',
    termWidth: 200, termHeight: 50, printElements: 1000, persistBreakpoints: 'default',
  })
  const [dict, setDict] = useState<DictForm>({ querySource: '', mirrorDir: '', syncTarget: '', bdldocDir: '' })
  const [listen, setListen] = useState('')
  // 派生状态(路径存不存在只有服务端算得出来)与配置元信息,以及 PATH 安装状态
  const [status, setStatus] = useState<ConfigStatus | null>(null)
  const [meta, setMeta] = useState<ConfigMeta | null>(null)
  const [installBusy, setInstallBusy] = useState(false)
  const [installNote, setInstallNote] = useState('')
  const [cacheBusy, setCacheBusy] = useState(false)
  const [cacheNote, setCacheNote] = useState('')
  // 源码镜像 / 字典同步的运行态:它们是**动作**(长跑任务 + 进度),不再是独立页面,
  // 就住在「数据字典」分区的对应卡片里。
  const [mirror, setMirror] = useState<MirrorResp | null>(null)
  const [mirrorEnv, setMirrorEnv] = useState('')
  const [sync, setSync] = useState<DBSyncResp | null>(null)
  const [syncEnv, setSyncEnv] = useState('')
  const [opBusy, setOpBusy] = useState('')
  const [opNote, setOpNote] = useState<{ kind: 'ok' | 'err'; text: string } | null>(null)
  const [selSsh, setSelSsh] = useState(0)
  const [err, setErr] = useState('')
  const [saveState, setSaveState] = useState<'idle' | 'saving' | 'saved'>('idle')
  // 脏标记按**配置节**记。保存时只提交改过的节 —— 设置页的各张卡片共用 /api/hosts,
  // 把没改的节一起发过去就是用陈旧快照覆盖别张卡片刚改好的节。
  const dirtyRef = useRef<Record<ConfigKey, boolean>>({
    hosts: false, debug: false, listen: false, query: false, mirror: false, sync: false, bdldoc: false,
  })
  const [dirty, setDirty] = useState(false)
  const markDirty = (...keys: ConfigKey[]) => {
    keys.forEach((k) => { dirtyRef.current[k] = true })
    setDirty(true)
  }
  const cleanDirty = (keys?: ConfigKey[]) => {
    if (keys) keys.forEach((k) => { dirtyRef.current[k] = false })
    else (Object.keys(dirtyRef.current) as ConfigKey[]).forEach((k) => { dirtyRef.current[k] = false })
    setDirty((Object.keys(dirtyRef.current) as ConfigKey[]).some((k) => dirtyRef.current[k]))
  }
  const isDirty = (...keys: ConfigKey[]) => keys.some((k) => dirtyRef.current[k])

  const [showPwd, setShowPwd] = useState(false)
  // 数据库页操作状态
  const [busy, setBusy] = useState<string>('')
  const [note, setNote] = useState<{ kind: 'ok' | 'err' | 'info'; text: string } | null>(null)
  const [accProbe, setAccProbe] = useState<{ i: number; state: 'testing' | 'ok' | 'err'; msg: string } | null>(null)
  const [addAcct, setAddAcct] = useState({ account: '', password: '' })
  const [delTarget, setDelTarget] = useState<number | null>(null) // 待删除环境的下标(弹窗确认)

  const sshName = (e: SshItem) => e.name || `${e.host}-${e.zone}`.replace(/-$/, '')

  const loadSettings = useCallback(() => {
    api.settings().then((c) => {
      setCfg(c)
      setSshs((c.sshs || []).map((e) => {
        const d = e.db
        return {
          name: e.name || '', host: e.host || '', port: e.port || 22, user: e.user || '', password: e.password || '',
          zone: e.zone || '', topent: e.topent != null ? String(e.topent) : '',
          db: d ? {
            type: d.type || 'oracle', host: d.host || '', port: d.port || 0,
            service: d.service || '', database: d.database || '',
            accounts: (d.accounts || []).map((a) => ({ account: a.account || '', password: a.password || '' })),
            readonlySql: d.readonlySql === false ? false : undefined,
          } : null,
        }
      }))
      setDbg({
        activeEnv: c.debug?.activeEnv || '',
        launchArgs: c.debug?.launchArgs || '',
        watchdogSeconds: c.debug?.watchdogSeconds || 0,
        fglserver: c.debug?.fglserver || '',
        termWidth: c.debug?.termWidth || 200,
        termHeight: c.debug?.termHeight || 50,
        printElements: c.debug?.printElements || 1000,
        persistBreakpoints: c.debug?.persistBreakpoints === true ? 'on'
          : c.debug?.persistBreakpoints === false ? 'off' : 'default',
      })
      setDict({
        querySource: c.query?.source || '',
        mirrorDir: c.mirror?.dir || '',
        syncTarget: c.sync?.target || '',
        bdldocDir: c.bdldoc?.dir || '',
      })
      setListen(c.listen || '')
      cleanDirty(); setSaveState('idle'); setErr('')
    }).catch((e) => setErr(e.message))
    // 派生状态与元信息:取不到不影响主流程(它们在卡片里只是提示)
    void api.configStatus().then(setStatus).catch(() => { /* 静默:提示性信息 */ })
    void api.configMeta().then(setMeta).catch(() => { /* 静默:提示性信息 */ })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => { loadSettings() }, [loadSettings])
  // 本页被 keep-alive 常驻挂载(切走只是 display:none),所以**不能轮询** ——
  // 改为"每次切回本页、且没有未保存修改时"回读一次,拿别处(命令行)改过的值。
  const dirtyRefAny = () => (Object.keys(dirtyRef.current) as ConfigKey[]).some((k) => dirtyRef.current[k])
  const visibleRef = useRef(false)
  const view = useStore((s) => s.view)
  useEffect(() => {
    const visible = view === 'settings'
    if (visible && !visibleRef.current && !dirtyRefAny()) void loadSettings()
    visibleRef.current = visible
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [view])

  // 分区变化 → 同步到 URL 片段(用 replace 而不是 push:否则浏览器后退会一格一格走遍所有分区)
  useEffect(() => {
    const want = settingsHash(section)
    if (window.location.hash !== want) window.history.replaceState(null, '', want)
  }, [section])

  // ---- 统一保存:按节提交;keys 指明这次要提交哪些节 ----
  const save = async (keys: ConfigKey[], list?: SshItem[]) => {
    if (saveState === 'saving' || !cfg) return
    const body: HostsPatch = {}
    if (keys.includes('hosts')) body.hosts = hostsPatchOf(list || sshs, cfg.activeEnv)
    if (keys.includes('debug')) body.debug = debugPatchOf(dbg)
    if (keys.includes('listen')) body.listen = listen.trim()
    if (keys.includes('query')) body.query = { source: dict.querySource }
    if (keys.includes('mirror')) body.mirror = { dir: dict.mirrorDir.trim() }
    if (keys.includes('sync')) body.sync = { target: dict.syncTarget.trim() }
    if (keys.includes('bdldoc')) body.bdldoc = { dir: dict.bdldocDir.trim() }
    setErr(''); setSaveState('saving')
    try {
      const next = await api.saveSettings(body)
      setCfg(next)
      setSshs((prev) => prev.map((x) => ({ ...x, name: sshName(x) })))
      cleanDirty(keys); setSaveState('saved')
      // 路径型取值的 exists 是服务端算的,保存后要回读一次才对得上新值
      void api.configStatus().then(setStatus).catch(() => { /* 提示性信息,取不到不影响 */ })
    } catch (ex: any) {
      // 服务端是校验的权威(环境名重复 / 端口越界 / 至少要有一个环境…),
      // 它给的中文说明原样显示,不要用本地判断把它盖掉
      setErr(ex.message || String(ex)); setSaveState('idle')
    }
  }

  // ---- 环境编辑 ----
  const patchSsh = (i: number, patch: Partial<SshItem>) => {
    const next = sshs.map((x, j) => (j === i ? { ...x, ...patch } : x))
    setSshs(next); markDirty('hosts')
  }
  const patchDb = (i: number, patch: Partial<SshDb>) => {
    const cur = sshs[i]
    if (!cur) return
    patchSsh(i, { db: cur.db ? { ...cur.db, ...patch } : { ...blankDb(), ...patch } })
  }
  const patchAcct = (i: number, j: number, patch: Partial<{ account: string; password: string }>) => {
    const cur = sshs[i]
    if (!cur?.db) return
    patchDb(i, { accounts: cur.db.accounts.map((a, k) => (k === j ? { ...a, ...patch } : a)) })
  }
  const delAcct = (i: number, j: number) => {
    const cur = sshs[i]
    if (!cur?.db) return
    patchDb(i, { accounts: cur.db.accounts.filter((_, k) => k !== j) })
  }
  const pushAcct = () => {
    const cur = sshs[selSsh]
    if (!cur?.db) return
    const account = addAcct.account.trim()
    if (!account) return
    patchDb(selSsh, { accounts: [...cur.db.accounts, { account, password: addAcct.password }] })
    setAddAcct({ account: '', password: '' })
  }
  const addSsh = () => { setSelSsh(sshs.length); setSshs([...sshs, blankSsh()]); markDirty('hosts') }
  const setDefaultEnv = (name: string) => {
    if (!cfg || name === cfg.activeEnv) return
    setCfg({ ...cfg, activeEnv: name })
    void save(['hosts'])
  }
  // 删除环境:先弹窗二次确认(误删代价高且不可撤销),确定后立即落盘保存
  const askDelSsh = () => { if (sshs[selSsh]) setDelTarget(selSsh) }
  const confirmDelSsh = async () => {
    const i = delTarget
    setDelTarget(null)
    if (i === null || !sshs[i]) return
    const next = sshs.filter((_, j) => j !== i)
    setSelSsh(Math.max(0, Math.min(i, next.length - 1)))
    setSshs(next)
    await save(['hosts'], next) // 点击「确定」= 删除并直接保存,不再需要手动点保存
  }

  // ---- 数据库页操作(针对当前环境的 db) ----
  // 从服务器获取:登录当前环境的 SSH 自动探测连接要素回填(辅助)
  const fetchFromServer = async () => {
    const s = sshs[selSsh]
    const d = s?.db
    if (!s?.host || !s?.user) { setNote({ kind: 'err', text: '请先填写 SSH 主机与账号' }); return }
    if (!d) return
    setBusy('fetch'); setNote(null)
    try {
      const r = await api.probeDB({ host: s.host, port: s.port || 22, user: s.user, password: s.password, zone: s.zone, type: d.type || 'oracle' })
      const patch: Partial<SshDb> = {}
      // 主机地址:服务器解析出的是 IP 就直接用;是服务器内部主机名(客户端多半不可达)则回填 SSH 主机
      // —— T100 的库通常与应用服务器同机,SSH 能连上就意味着该地址客户端可达。
      const probedHost = (r.host || '').trim()
      const isIP = /^\d{1,3}(\.\d{1,3}){3}$/.test(probedHost)
      const hostVal = (isIP ? probedHost : '') || (s.host || '').trim()
      if (hostVal) patch.host = hostVal
      if (r.type === 'kingbase') {
        if (r.database) patch.database = r.database
        if (r.port) patch.port = r.port
        setNote(r.note
          ? { kind: 'info', text: r.note }
          : { kind: 'ok', text: `获取成功:金仓 ${r.database || '?'} @ ${hostVal || '?'}:${r.port || '?'}(主机地址按 SSH 主机回填,如需从别处连库请手工调整)` })
      } else {
        if (r.service || r.tns) patch.service = r.service || r.tns || ''
        if (r.port) patch.port = r.port
        const inner = probedHost && !isIP ? `(服务器内部名 ${probedHost} 客户端多半不可达,已改用 SSH 主机)` : ''
        setNote({
          kind: r.note ? 'err' : 'info',
          text: (r.note ? r.note + ';' : '') +
            `服务器解析地址 ${probedHost || '未知'}:${r.port || '?'}(service ${r.service || r.tns || '?'});主机地址已回填 ${hostVal || '?'}${inner}`,
        })
      }
      if (Object.keys(patch).length) patchDb(selSsh, patch)
    } catch (ex: any) {
      setNote({ kind: 'err', text: '获取失败: ' + (ex.message || String(ex)) })
    } finally { setBusy('') }
  }
  // 附加账号行验证:服务器侧以 账号+密码 连当前环境的显式目标库
  const verifyAcct = async (j: number) => {
    const i = selSsh
    const s = sshs[i]
    const d = s?.db
    const row = d?.accounts?.[j]
    if (!s?.host || !s?.user) { setAccProbe({ i: j, state: 'err', msg: '请先填写 SSH 主机与账号' }); return }
    if (!d?.host.trim()) { setAccProbe({ i: j, state: 'err', msg: '请先填写数据库主机地址' }); return }
    if (!row?.account.trim()) { setAccProbe({ i: j, state: 'err', msg: '账号不能为空' }); return }
    setAccProbe({ i: j, state: 'testing', msg: '' })
    try {
      await api.dbAccVerify({
        host: s.host, port: s.port || 22, user: s.user, password: s.password, zone: s.zone,
        type: d.type || 'oracle', account: row.account.trim(), acctPassword: row.password,
        dbHost: d.host.trim(), dbPort: d.port || 0,
        ...(d.type === 'kingbase' ? { dbDatabase: d.database.trim() } : { dbSvc: d.service.trim() }),
      })
      setAccProbe({ i: j, state: 'ok', msg: '连接正常 ✓' })
    } catch (ex: any) {
      setAccProbe({ i: j, state: 'err', msg: ex.message || String(ex) })
    }
  }
  // 客户端直连测试:验证「本机 → 库」这条链路真的通(与服务器侧验证互补)
  const testConn = async () => {
    const d = sshs[selSsh]?.db
    if (!d?.host.trim()) { setNote({ kind: 'err', text: '请先填写数据库主机地址' }); return }
    setBusy('conn'); setNote(null)
    try {
      const r = await api.connTest({
        type: d.type || 'oracle', host: d.host.trim(), port: d.port || 0,
        service: d.service.trim(), database: d.database.trim(),
        accounts: d.accounts.map((a) => ({ account: a.account.trim(), password: a.password })).filter((a) => a.account),
      })
      const ver = (r.serverVersion || '').split('\n')[0]
      setNote({ kind: r.ok ? 'ok' : 'err', text: r.ok ? `连接正常:${ver}` : `${r.stage === 'version' ? '取版本失败' : '连接失败'}:${r.error}` })
    } catch (ex: any) {
      setNote({ kind: 'err', text: '连接失败: ' + (ex.message || String(ex)) })
    } finally { setBusy('') }
  }

  // ---- 命令行集成(用户 PATH) ----
  const doInstall = async (add: boolean) => {
    setInstallBusy(true); setInstallNote('')
    try {
      const st = add ? await api.installAdd() : await api.installRemove()
      setStatus((prev) => (prev ? { ...prev, install: st } : prev))
      setInstallNote(add ? `已把 ${st.exeDir} 加入用户 PATH,新开的终端即可直接敲 tt` : `已从用户 PATH 移除 ${st.exeDir}`)
    } catch (ex: any) {
      setInstallNote('操作失败: ' + (ex.message || String(ex)))
    } finally { setInstallBusy(false) }
  }

  // ---- 缓存清理 ----
  // 只删可再生的中间数据(后端只认 internal/config 那份清单);config.json 不碰。
  // 后端把清完的状态一起回,所以这里不必再打一次 configStatus。
  const doClearCache = async () => {
    setCacheBusy(true); setCacheNote('')
    try {
      const r = await api.cacheClear()
      setStatus((prev) => (prev ? { ...prev, cache: r.cache } : prev))
      setCacheNote(r.removed > 0
        ? `已清理 ${r.removed} 个文件,释放 ${fmtBytes(r.freed)}`
        : '没有可清理的文件')
    } catch (ex: any) {
      setCacheNote('清理失败: ' + (ex.message || String(ex)))
    } finally { setCacheBusy(false) }
  }

  // 「已保存」提示短暂停留后回到空闲
  useEffect(() => {
    if (saveState !== 'saved') return
    const t = window.setTimeout(() => setSaveState('idle'), 2000)
    return () => window.clearTimeout(t)
  }, [saveState])

  // ---- 源码镜像 / 字典同步(动作) ----
  // 只在「数据字典」分区且正在跑的时候轮询:本页是常驻挂载的,无脑轮询会一直打后端。
  const refreshMirror = useCallback(async () => {
    try {
      const d = await api.mirror()
      setMirror(d)
      setMirrorEnv((prev) => {
        if (prev && d.envs.some((e) => e.name === prev)) return prev
        if (d.activeEnv && d.envs.some((e) => e.name === d.activeEnv)) return d.activeEnv
        return d.envs[0]?.name || ''
      })
    } catch { /* 取不到就不显示操作区,不打扰 */ }
  }, [])
  const refreshSync = useCallback(async () => {
    try {
      const d = await api.dbsync()
      setSync(d)
      setSyncEnv((prev) => {
        if (prev && d.envs.some((e) => e.name === prev)) return prev
        if (d.activeEnv && d.envs.some((e) => e.name === d.activeEnv)) return d.activeEnv
        return d.envs[0]?.name || ''
      })
    } catch { /* 同上 */ }
  }, [])
  useEffect(() => {
    if (section !== 'data-dict') return
    void refreshMirror(); void refreshSync()
  }, [section, refreshMirror, refreshSync])
  const mirrorRunning = !!mirror?.job?.running
  const syncRunning = !!sync?.job?.running
  useEffect(() => {
    if (!mirrorRunning) return
    const t = window.setInterval(() => { void refreshMirror() }, 800)
    return () => window.clearInterval(t)
  }, [mirrorRunning, refreshMirror])
  useEffect(() => {
    if (!syncRunning) return
    const t = window.setInterval(() => { void refreshSync() }, 800)
    return () => window.clearInterval(t)
  }, [syncRunning, refreshSync])

  const pullMirror = async (full: boolean) => {
    if (!mirrorEnv) return
    if (full && !window.confirm(`全量重建环境「${mirrorEnv}」?
将整目录替换本地镜像(删除服务器已不存在的残留),首次或按需执行,耗时较长。`)) return
    setOpBusy('mirror'); setOpNote(null)
    try {
      await api.mirrorPull(mirrorEnv, full)
      await refreshMirror()
      setOpNote({ kind: 'ok', text: `已开始${full ? '全量重建' : '增量更新'}:${mirrorEnv}` })
    } catch (ex: any) {
      setOpNote({ kind: 'err', text: '拉取失败: ' + (ex.message || String(ex)) })
    } finally { setOpBusy('') }
  }
  const runSync = async () => {
    if (!syncEnv) return
    if (!window.confirm(`从环境「${syncEnv}」拉取字典数据到:
${sync?.target || ''}

将覆盖本地 SQLite(原库自动备份为 .bak),约 85 万行,需数分钟。继续?`)) return
    setOpBusy('sync'); setOpNote(null)
    try {
      await api.dbsyncRun(syncEnv)
      await refreshSync()
      setOpNote({ kind: 'ok', text: `已开始同步:${syncEnv}` })
    } catch (ex: any) {
      setOpNote({ kind: 'err', text: '同步失败: ' + (ex.message || String(ex)) })
    } finally { setOpBusy('') }
  }

  if (!cfg) return <div className="p-6 text-sm text-muted-foreground">{err || '加载配置中…'}</div>
  const cur = sshs[selSsh]
  const curDb = cur?.db
  const envNames = sshs.map((e) => sshName(e)).filter(Boolean)

  // 卡片的保存按钮:只在真的有改动时才出现。
  // 常驻一个灰着的「保存」挂在每张卡右上角,既不传达信息(它一直是灰的),又让人以为
  // 有什么东西没保存。有没有未保存的改动看内容区右上角那行状态文字就够了,这条按钮
  // 是「动手保存」的入口,没得保存时它就不该在。
  const SaveBtn = ({ keys }: { keys: ConfigKey[] }) => {
    if (!isDirty(...keys)) return null
    return (
      <Button size="sm" disabled={saveState === 'saving'} onClick={() => void save(keys)}>
        {saveState === 'saving' ? '保存中…' : '保存'}
      </Button>
    )
  }

  return (
    // flex-1 + min-w-0:填满活动栏右侧的整块区域。少了它这个根会按内容宽度撑开,
    // 于是外层的滚动条落在内容列的右缘(页面中间),右边留一大片空白 —— 滚动条看着像没贴着边。
    // 填满之后里面的 mx-auto max-w-3xl 才真正居中,标题与卡片始终对齐在同一条竖线上。
    <div className="flex h-full min-h-0 min-w-0 flex-1 text-xs">
      {/* 左侧导航:四个分区各一行。
          分区不带图标,也不给「设置」这类标题 —— 这一栏本身就在设置页里,
          重复一遍栏名是废话,而图标只是把每个分区名往右推了一格。
          原先选中项下面还会展开该分区的卡片清单,已经去掉:右边一屏就是那个分区的全部卡片,
          清单只是把它们再念一遍。 */}
      <div className="flex w-44 shrink-0 flex-col overflow-auto border-r border-border">
        {SETTINGS_SECTIONS.map((s) => {
          const active = section === s.key
          return (
            <button
              key={s.key}
              onClick={() => setSection(s.key)}
              className={cn('flex w-full items-center px-2 py-1.5 text-left transition-colors',
                active ? 'bg-accent font-medium text-accent-foreground' : 'text-muted-foreground hover:bg-accent/60')}
            >
              <span className="min-w-0 flex-1 truncate">{s.label}</span>
            </button>
          )
        })}
      </div>

      {/* 右侧内容区:当前分区的全部卡片,整体滚动。
          站点管理是左右两栏(左边环境清单、右边连接表单),列宽放宽一档给表单留出原来的舒适宽度;
          其余分区都是单列设置列表,仍然收在 max-w-3xl。 */}
      <div className="min-h-0 flex-1 overflow-auto">
        <div className={cn('mx-auto space-y-3 p-4', section === 'sites' ? 'max-w-5xl' : 'max-w-3xl')}>
          <div className="flex items-center justify-between gap-2">
            <h2 className="text-sm font-medium text-foreground">
              {SETTINGS_SECTIONS.find((s) => s.key === section)?.label}
            </h2>
            <span className="flex min-w-0 items-center gap-2">
              {err && <span className="truncate text-[11px] text-red-600 dark:text-red-400" title={err}>{err}</span>}
              {!err && (
                <span className={cn('text-[11px]',
                  saveState === 'saved' ? 'text-emerald-600 dark:text-emerald-400'
                    : dirty ? 'text-amber-600 dark:text-amber-400' : '')}>
                  {saveState === 'saving' ? '保存中…' : saveState === 'saved' ? '已保存 ✓' : dirty ? '有未保存的修改' : ''}
                </span>
              )}
              {/* 本页常驻挂载、不轮询(轮询会一直 os.Stat 配置里的路径),所以给一个显式回读入口:
                  命令行改过配置后,点它就能看到最新值 */}
              <Button variant="ghost" size="sm" disabled={dirty || saveState === 'saving'}
                title={dirty ? '有未保存的修改,先保存或放弃再回读' : '从 config.json 重新读取'}
                onClick={() => loadSettings()}>
                <RefreshCw className="h-3.5 w-3.5" />
              </Button>
            </span>
          </div>

          {/* ============ 站点管理(左右两栏:环境清单 | 连接表单) ============ */}
          {section === 'sites' && (
            <div className="grid h-[40rem] grid-cols-[16rem_1fr] gap-3">
              {/* 两栏定高:跟环境数量、表单长短都无关,两栏永远一样高,页面也不会跟着内容忽长忽短。
                  高度取 40rem —— 常见窗口高度下装得下,也不需要跟着视口算。 */}
              <Card
                id="card-sites-list"
                panel
                title="环境清单"
                right={<SaveBtn keys={['hosts']} />}
              >
                {/* 列表撑满卡片余下的高度、自己滚:环境再多也不顶高卡片,
                    「新增环境」钉在卡片底部,不用翻到列表末尾去找 */}
                <div className="min-h-0 flex-1 space-y-0.5 overflow-auto">
                  {sshs.length === 0 && <div className="py-2 text-muted-foreground">(空)</div>}
                  {sshs.map((e, i) => {
                    const isDefault = cfg.activeEnv === sshName(e)
                    return (
                      <div key={i}
                        className={cn('flex items-center gap-1 px-1.5 py-1 transition-colors',
                          i === selSsh ? 'bg-accent text-accent-foreground' : 'hover:bg-accent/50')}>
                        <button className="min-w-0 flex-1 text-left" onClick={() => setSelSsh(i)}>
                          <span className="block truncate">{sshName(e) || '(新环境)'}</span>
                          <span className="block truncate text-[11px] text-muted-foreground">
                            {e.zone || '-'} · {e.port || '-'} · {e.db ? (e.db.type === 'kingbase' ? '金仓' : 'Oracle') : '无库'}
                          </span>
                        </button>
                        <button
                          title={isDefault ? '当前默认环境' : '设为默认环境(会话未指定环境时用它)'}
                          disabled={isDefault}
                          onClick={() => setDefaultEnv(sshName(e))}
                          className="shrink-0 p-0.5 transition-colors hover:bg-accent/60 disabled:opacity-100"
                        >
                          <Star className={cn('h-3.5 w-3.5', isDefault ? 'fill-amber-400 text-amber-400' : 'text-muted-foreground')} />
                        </button>
                      </div>
                    )
                  })}
                </div>
                <div className="flex items-center gap-2">
                  <Button variant="outline" size="sm" onClick={addSsh}><Plus className="h-3.5 w-3.5" />新增环境</Button>
                  <span className="text-[11px] text-muted-foreground">星标 = 默认环境({cfg.activeEnv || '未设置'})</span>
                </div>
              </Card>

              <Card
                id="card-sites-conn"
                panel
                title="服务器与数据库"
                description="SSH 登录与数据库连接按环境一对一挂载。运行时用哪个数据库账号由 TOPENT 决定(服务器侧解析),客户端直连取账号列表首项。"
                right={<SaveBtn keys={['hosts']} />}
              >
                {!cur ? (
                  <div className="py-2 text-muted-foreground">先在左边的「环境清单」里选一个环境,或新增一个。</div>
                ) : (
                  <>
                    <div className="flex items-center justify-between gap-2">
                      <span className="min-w-0 truncate text-xs font-medium text-foreground">{sshName(cur) || '(新环境)'}</span>
                      <Button variant="ghost" size="sm" title="删除该环境" onClick={askDelSsh}>
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                    {/* 两个子页签做成一排小按钮(选中给 accent 底色),不画下边框 ——
                        卡片内部不打分割线,靠底色区分选中态 */}
                    <div className="flex gap-1">
                      {([['ssh', 'SSH 服务器'], ['db', '数据库']] as const).map(([k, label]) => (
                        <button key={k} onClick={() => setEnvTab(k)}
                          className={cn('px-2.5 py-1 text-xs transition-colors',
                            envTab === k ? 'bg-accent font-medium text-foreground' : 'text-muted-foreground hover:bg-accent/60 hover:text-foreground')}>
                          {label}
                        </button>
                      ))}
                    </div>

                    {envTab === 'ssh' && (
                      <>
                        <div className="grid grid-cols-2 gap-x-4 gap-y-3">
                          <Field label="环境名称(留空自动为主机-区域)">
                            <Input className={input} value={cur.name} placeholder="如 正式区"
                              onChange={(e) => patchSsh(selSsh, { name: e.target.value })} />
                          </Field>
                          <Field label="IP 主机">
                            <Input className={input} value={cur.host} onChange={(e) => patchSsh(selSsh, { host: e.target.value })} />
                          </Field>
                          <Field label="端口">
                            <Input className={input} type="number" value={cur.port || ''}
                              onChange={(e) => patchSsh(selSsh, { port: Number(e.target.value) || 22 })} />
                          </Field>
                          <Field label="登录区域">
                            <Input className={input} value={cur.zone} placeholder="31开发/35测试/36正式/39PATCH/t出货"
                              onChange={(e) => patchSsh(selSsh, { zone: e.target.value })} />
                          </Field>
                          <Field label="账号">
                            <Input className={input} value={cur.user} onChange={(e) => patchSsh(selSsh, { user: e.target.value })} />
                          </Field>
                          <Field label="密码">
                            <div className="relative">
                              <Input className={input} type={showPwd ? 'text' : 'password'} value={cur.password}
                                onChange={(e) => patchSsh(selSsh, { password: e.target.value })} />
                              <button type="button" title={showPwd ? '隐藏密码' : '显示密码'}
                                onClick={() => setShowPwd((v) => !v)}
                                className="absolute top-1/2 right-1 -translate-y-1/2 p-0.5 text-muted-foreground hover:text-foreground">
                                {showPwd ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                              </button>
                            </div>
                          </Field>
                          <Field label="TOPENT(默认企业;连接会话即下发,可数字或文本)" className="col-span-2">
                            <Input className={input} value={cur.topent} onChange={(e) => patchSsh(selSsh, { topent: e.target.value })} />
                          </Field>
                        </div>
                        <p className="mt-2 text-[11px] text-muted-foreground">
                          调试会话按该服务器登录区域(区域 + TOPENT)。该环境的数据库连接在「数据库」Tab 维护,一对一。
                        </p>
                      </>
                    )}

                    {envTab === 'db' && (
                      curDb == null ? (
                        <div className="border border-dashed border-border p-4 text-center text-muted-foreground">
                          <p className="mb-2">该环境尚未挂载数据库连接</p>
                          <Button variant="outline" size="sm" onClick={() => patchDb(selSsh, {})}>添加数据库</Button>
                        </div>
                      ) : (
                        <>
                          <div className="grid grid-cols-2 gap-x-4 gap-y-3">
                            <Field label="类型">
                              <Select value={curDb.type || 'oracle'} onValueChange={(v) => patchDb(selSsh, { type: v })}>
                                <SelectTrigger className="h-7 w-full text-xs"><SelectValue /></SelectTrigger>
                                <SelectContent>
                                  <SelectItem value="oracle">Oracle</SelectItem>
                                  <SelectItem value="kingbase">金仓(KingbaseES)</SelectItem>
                                </SelectContent>
                              </Select>
                            </Field>
                            <Field label="主机地址">
                              <Input className={input} value={curDb.host} placeholder="客户端与服务器侧均可达"
                                onChange={(e) => patchDb(selSsh, { host: e.target.value })} />
                            </Field>
                            <Field label="端口">
                              <Input className={input} type="number" value={curDb.port || ''}
                                placeholder={curDb.type === 'kingbase' ? '54321' : '1521'}
                                onChange={(e) => patchDb(selSsh, { port: Number(e.target.value) || 0 })} />
                            </Field>
                            {curDb.type === 'kingbase' ? (
                              <Field label="库名">
                                <Input className={input} value={curDb.database} placeholder="如 topprd"
                                  onChange={(e) => patchDb(selSsh, { database: e.target.value })} />
                              </Field>
                            ) : (
                              <Field label="服务名(SERVICE_NAME)">
                                <Input className={input} value={curDb.service} placeholder="如 t35prd"
                                  onChange={(e) => patchDb(selSsh, { service: e.target.value })} />
                              </Field>
                            )}
                          </div>
                          <label className="mt-3 flex items-center gap-2 text-[11px] text-muted-foreground">
                            <input type="checkbox" checked={curDb.readonlySql !== false}
                              onChange={(e) => patchDb(selSsh, { readonlySql: e.target.checked ? undefined : false })} />
                            允许 AI 执行只读 SQL(默认开启;关掉后调试端点的 SQL 查询直接 403)
                          </label>
                          <div className="mt-3 flex items-center gap-2">
                            <Button variant="outline" size="sm" disabled={busy !== '' || !cur.host || !cur.user}
                              title={!cur.host || !cur.user ? '先填 SSH 主机与账号' : '按区域登录服务器,解析库地址/service 并回填'}
                              onClick={() => void fetchFromServer()}>
                              <RefreshCw className={cn('h-3.5 w-3.5', busy === 'fetch' && 'animate-spin')} />从服务器获取
                            </Button>
                            <Button variant="outline" size="sm" disabled={busy !== '' || !curDb.host.trim()}
                              title="从本机直连该库,验证网络与账号"
                              onClick={() => void testConn()}>
                              <Zap className={cn('h-3.5 w-3.5', busy === 'conn' && 'animate-spin')} />测试连接
                            </Button>
                          </div>
                          {note && (
                            <p className={cn('mt-2 border-l-2 px-2 py-1 text-[11px]',
                              note.kind === 'ok' ? 'border-emerald-500/50 bg-emerald-500/5 text-emerald-700 dark:text-emerald-300'
                                : note.kind === 'err' ? 'border-red-500/50 bg-red-500/5 text-red-700 dark:text-red-300'
                                  : 'border-sky-500/50 bg-sky-500/5 text-muted-foreground')}>
                              {note.text}
                            </p>
                          )}

                          <div className="mt-4">
                            <div className="mb-1 text-[11px] text-muted-foreground">
                              账号列表(账号即 schema 名;客户端直连取首项,服务器侧调试按 TOPENT 解析出账号后在此查密码)
                            </div>
                            {/* 账号可能很多:表格封顶 16rem 后就地滚,表头 sticky 常驻。
                                滚动容器由我们自己提供(container={false}) —— 用 Table 自带的
                                那个 overflow-x-auto 包一层的话,它就成了滚动祖先,th 的 sticky 会失效。
                                表头底色必须不透明(默认 bg-muted/50 是半透明的),否则行会从表头透出来。 */}
                            <div className="max-h-64 overflow-auto">
                              <Table container={false}>
                                <TableHeader>
                                  <TableRow className="hover:bg-transparent">
                                    <TableHead className="sticky top-0 z-10 w-1/2 bg-muted">账号(schema)</TableHead>
                                    <TableHead className="sticky top-0 z-10 bg-muted">密码(缺省 = 账号)</TableHead>
                                    <TableHead className="sticky top-0 z-10 w-16 bg-muted">操作</TableHead>
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  <TableRow>
                                    <TableCell className={CELL_EDITING}>
                                      <input className={cn('cell-input font-mono')} placeholder="新增账号…" value={addAcct.account}
                                        onChange={(e) => setAddAcct((s) => ({ ...s, account: e.target.value }))}
                                        onKeyDown={(e) => { if (e.key === 'Enter') pushAcct() }} />
                                    </TableCell>
                                    <TableCell className={CELL_EDITING}>
                                      <input className="cell-input font-mono" placeholder="留空 = 与账号相同" value={addAcct.password}
                                        onChange={(e) => setAddAcct((s) => ({ ...s, password: e.target.value }))}
                                        onKeyDown={(e) => { if (e.key === 'Enter') pushAcct() }} />
                                    </TableCell>
                                    <TableCell className="text-center">
                                      <Button variant="ghost" size="sm" title="添加账号" disabled={!addAcct.account.trim()} onClick={pushAcct}>
                                        <Plus className="h-3.5 w-3.5" />
                                      </Button>
                                    </TableCell>
                                  </TableRow>
                                  {curDb.accounts.map((a, j) => (
                                    <TableRow key={j}>
                                      <TableCell className={CELL_EDITING}>
                                        <input className="cell-input font-mono" value={a.account}
                                          onChange={(e) => patchAcct(selSsh, j, { account: e.target.value })} />
                                      </TableCell>
                                      <TableCell className={cn(CELL_EDITING, 'relative')}>
                                        <input className="cell-input font-mono" value={a.password}
                                          onChange={(e) => patchAcct(selSsh, j, { password: e.target.value })} />
                                      </TableCell>
                                      <TableCell className="text-center whitespace-nowrap">
                                        <Button variant="ghost" size="sm" title="在服务器侧验证该账号"
                                          disabled={busy !== '' || accProbe?.state === 'testing'}
                                          onClick={() => void verifyAcct(j)}>
                                          <Zap className={cn('h-3.5 w-3.5', accProbe?.i === j && accProbe.state === 'testing' && 'animate-spin')} />
                                        </Button>
                                        <Button variant="ghost" size="sm" title="删除该账号" onClick={() => delAcct(selSsh, j)}>
                                          <Trash2 className="h-3.5 w-3.5" />
                                        </Button>
                                      </TableCell>
                                    </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </div>
                            {accProbe && (
                              <p className={cn('mt-1 text-[11px]',
                                accProbe.state === 'ok' ? 'text-emerald-600 dark:text-emerald-400'
                                  : accProbe.state === 'err' ? 'text-red-600 dark:text-red-400' : 'text-muted-foreground')}>
                                {accProbe.msg || '验证中…'}
                              </p>
                            )}
                          </div>
                          <p className="mt-2 text-[11px] text-muted-foreground">
                            库连接显式给出地址与 service|库名,不依赖服务器侧的 TNS 配置;服务器上的 sqlplus/ksql 路径自动探测,不存入配置。
                          </p>
                        </>
                      )
                    )}
                  </>
                )}
              </Card>
            </div>
          )}

          {/* ============ 数据字典 ============ */}
          {section === 'data-dict' && (
            <>
              <Card id="card-dict-query" title="查询数据源"
                description="r.t / r.v / desc / scc / r.q 等查询命令用哪个数据源。缺省是在线优先:用默认环境的远程库直查;一个环境都没配时才用本地镜像。"
                right={<SaveBtn keys={['query']} />}>
                <SettingRow
                  id="query.source"
                  label="数据源"
                  description="本地镜像由 tt dict db sync 生成;连不上不会静默回落本地 —— 那会让你以为查到的是实时数据。"
                  control={
                    <Select value={dict.querySource || '__auto__'}
                      onValueChange={(v) => { setDict((s) => ({ ...s, querySource: v === '__auto__' ? '' : v })); markDirty('query') }}>
                      <SelectTrigger className="h-7 w-full text-xs"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="__auto__">在线(默认环境)</SelectItem>
                        <SelectItem value="local">本地 SQLite 镜像</SelectItem>
                        {envNames.map((n) => <SelectItem key={n} value={n}>在线({n})</SelectItem>)}
                      </SelectContent>
                    </Select>
                  }
                />
              </Card>

              <Card id="card-dict-mirror" title="源码镜像"
                description="把 T100 服务器的源码(4gl / 4fd / 42s 中文 / *.inc)拉到本地,供 AI 直接检索。"
                right={<SaveBtn keys={['mirror']} />}>
                <SettingRow
                  id="mirror.dir"
                  label="镜像根目录"
                  description={
                    !dict.mirrorDir.trim()
                      ? '绝对路径。留空 = 未设置,镜像功能不可用。'
                      : status?.mirror.exists
                        ? '该目录存在于本机。'
                        : '该目录在本机不存在 —— 拉取时会自动创建。'
                  }
                  control={<Input className={input} value={dict.mirrorDir} placeholder="如 D:\t100\mirror"
                    onChange={(e) => { setDict((s) => ({ ...s, mirrorDir: e.target.value })); markDirty('mirror') }} />}
                />
                <SettingRow
                  label="拉取环境"
                  control={
                    <Select value={mirrorEnv} onValueChange={setMirrorEnv} disabled={mirrorRunning}>
                      <SelectTrigger className="h-7 w-full text-xs"><SelectValue placeholder="选择环境" /></SelectTrigger>
                      <SelectContent>
                        {(mirror?.envs || []).map((e) => (
                          <SelectItem key={e.name} value={e.name}>
                            {e.name}{e.name === mirror?.activeEnv ? '（默认）' : ''}{e.ready ? ' · 已有镜像' : ' · 未拉取'}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  }
                />
                {mirror && mirror.envs.length === 0 && (
                  <p className="text-[11px] text-muted-foreground">还没有 SSH 环境,先去「站点管理」添加。</p>
                )}
                <div className="flex flex-wrap items-center gap-2">
                  <Button size="sm" disabled={mirrorRunning || opBusy !== '' || !mirrorEnv || !dict.mirrorDir.trim() || isDirty('mirror')}
                    title={isDirty('mirror') ? '镜像根目录有未保存的修改,先保存' : '只拉服务器上变过的文件'}
                    onClick={() => void pullMirror(false)}>
                    <RefreshCw className={cn('h-3.5 w-3.5', mirrorRunning && 'animate-spin')} />增量更新
                  </Button>
                  <Button variant="outline" size="sm" disabled={mirrorRunning || opBusy !== '' || !mirrorEnv || !dict.mirrorDir.trim() || isDirty('mirror')}
                    title={isDirty('mirror') ? '镜像根目录有未保存的修改,先保存' : '整目录替换(含服务器上已删的残留)'}
                    onClick={() => void pullMirror(true)}>
                    <Download className="h-3.5 w-3.5" />全量重建
                  </Button>
                  <span className="text-[11px] text-muted-foreground">增量按服务器 marker 记基线;本地无完整镜像时自动转全量。</span>
                </div>
                {mirror?.job && (mirror.job.running || mirror.job.done) && (
                  <JobProgress
                    env={mirror.job.env}
                    phaseLabel={MIRROR_PHASE[mirror.job.phase] || mirror.job.phase}
                    pct={mirror.job.total > 0
                      ? Math.min(100, Math.round((mirror.job.bytes * 100) / mirror.job.total))
                      : mirror.job.phase === 'done' ? 100 : 0}
                    indeterminate={mirror.job.running && mirror.job.total === 0}
                    elapsed={mirror.job.running ? mirror.job.elapsed : (mirror.job.elapsed ? '用时 ' + mirror.job.elapsed : '')}
                    error={mirror.job.phase === 'error' ? (mirror.job.error || mirror.job.message) : undefined}
                    message={mirror.job.phase === 'done' ? mirror.job.message : undefined}
                  >
                    <span>已传输:{fmtBytes(mirror.job.bytes)}{mirror.job.total > 0 ? ` / ${fmtBytes(mirror.job.total)}` : ''}</span>
                    <span>文件数:{mirror.job.files || '—'}</span>
                    <span>{mirror.job.full ? '全量' : '增量'} · {mirror.job.running ? '进行中…' : '已结束'}</span>
                  </JobProgress>
                )}
              </Card>

              <Card id="card-dict-sync" title="数据同步"
                description="把远程 ERP 的字典表同步成本地 SQLite,查询命令再读它。"
                right={<SaveBtn keys={['sync']} />}>
                <SettingRow
                  id="sync.target"
                  label="目标数据库文件"
                  description={
                    dict.syncTarget.trim()
                      ? (status?.sync.exists ? '该文件已存在(同步时会自动备份为 .bak)。' : '该文件尚未创建,同步时自动创建。')
                      : `绝对路径。留空 = 用默认位置:${status?.sync.defaultTarget || 'erp_data.db'}`
                  }
                  control={
                    <div className="flex items-center gap-1.5">
                      <Input className={input} value={dict.syncTarget} placeholder="留空 = 默认位置"
                        onChange={(e) => { setDict((s) => ({ ...s, syncTarget: e.target.value })); markDirty('sync') }} />
                      <Button variant="outline" size="sm" title="清空 = 恢复默认位置"
                        disabled={!dict.syncTarget}
                        onClick={() => { setDict((s) => ({ ...s, syncTarget: '' })); markDirty('sync') }}>
                        <RotateCcw className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  }
                />
                <SettingRow
                  label="同步环境"
                  control={
                    <Select value={syncEnv} onValueChange={setSyncEnv} disabled={syncRunning}>
                      <SelectTrigger className="h-7 w-full text-xs"><SelectValue placeholder="选择环境" /></SelectTrigger>
                      <SelectContent>
                        {(sync?.envs || []).map((e) => (
                          <SelectItem key={e.name} value={e.name}>
                            {e.name}{e.name === sync?.activeEnv ? '（默认）' : ''} · {e.type}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  }
                />
                {sync && sync.envs.length === 0 && (
                  <p className="text-[11px] text-muted-foreground">
                    没有可同步的环境 —— 需要先在「站点管理」为环境挂上数据库连接。
                  </p>
                )}
                <div className="flex flex-wrap items-center gap-2">
                  <Button size="sm" disabled={syncRunning || opBusy !== '' || !syncEnv || isDirty('sync')}
                    title={isDirty('sync') ? '同步目标有未保存的修改,先保存' : '从远程 ERP 拉取字典数据到本地 SQLite'}
                    onClick={() => void runSync()}>
                    <RefreshCw className={cn('h-3.5 w-3.5', syncRunning && 'animate-spin')} />{syncRunning ? '同步中…' : '开始同步'}
                  </Button>
                  <span className="text-[11px] text-muted-foreground">
                    逐表拉取 → 建索引 → 原子替换;原库自动备份为 <code>.bak</code>,失败不影响原库。
                  </span>
                </div>
                {sync?.job && (sync.job.running || sync.job.done) && (
                  <JobProgress
                    env={sync.job.env}
                    phaseLabel={SYNC_PHASE[sync.job.phase] || sync.job.phase}
                    pct={sync.job.phase === 'done' ? 100
                      : sync.job.tableTotal > 0 ? Math.min(100, Math.round((sync.job.tableIndex * 100) / sync.job.tableTotal)) : 0}
                    indeterminate={sync.job.running && sync.job.tableTotal === 0}
                    elapsed={sync.job.running ? sync.job.elapsed : (sync.job.elapsed ? '用时 ' + sync.job.elapsed : '')}
                    error={sync.job.phase === 'error' ? (sync.job.error || sync.job.message) : undefined}
                    message={sync.job.phase === 'done' ? sync.job.message : undefined}
                  >
                    <span>数据表:{sync.job.tableTotal > 0 ? `${sync.job.tableIndex}/${sync.job.tableTotal}` : '—'}{sync.job.tables > 0 ? `(完成 ${sync.job.tables})` : ''}</span>
                    <span>当前表:{(sync.job.table || '—') + (sync.job.tableRows > 0 ? ` ${sync.job.tableRows} 行` : '')}</span>
                    <span>累计行数:{sync.job.totalRows > 0 ? sync.job.totalRows.toLocaleString() : '—'}</span>
                  </JobProgress>
                )}
                {sync?.job?.backup && <p className="text-[11px] text-muted-foreground">原库已备份:{sync.job.backup}</p>}
                {sync?.job?.warning && <pre className="whitespace-pre-wrap text-[11px] text-amber-600 dark:text-amber-400">{sync.job.warning}</pre>}
              </Card>

              <Card id="card-dict-bdldoc" title="BDL 文档"
                description="Genero BDL 语言文档的本地落点(随发行包不附带,指向你自己的文档目录)。"
                right={<SaveBtn keys={['bdldoc']} />}>
                <SettingRow
                  id="bdldoc.dir"
                  label="文档目录"
                  description={
                    !dict.bdldocDir.trim()
                      ? '绝对路径。留空 = 未设置。'
                      : status?.bdldoc.exists
                        ? '该目录存在于本机。'
                        : '该目录在本机不存在。'
                  }
                  control={<Input className={input} value={dict.bdldocDir} placeholder="如 D:\t100\bdldoc"
                    onChange={(e) => { setDict((s) => ({ ...s, bdldocDir: e.target.value })); markDirty('bdldoc') }} />}
                />
              </Card>
            </>
          )}

          {/* ============ DEBUG ============ */}
          {section === 'debug' && (
            <>
              <Card id="card-debug-params" title="调试参数"
                description="本机调试会话的默认参数。清空某一项 = 恢复该项的内置默认值。"
                right={<SaveBtn keys={['debug']} />}>
                <SettingRow id="debug.launchArgs" label="启动参数模板"
                  description="{prog} 会替换为作业编号。清空 = 用内置默认。"
                  control={<Input className={input} value={dbg.launchArgs} placeholder="BBDL512840855a 2 12345 'N' {prog}"
                    onChange={(e) => { setDbg((s) => ({ ...s, launchArgs: e.target.value })); markDirty('debug') }} />} />
                <SettingRow id="debug.watchdogSeconds" label="停站看门狗(秒)"
                  description="停站后原地停留超过该时长就判定会话失联。"
                  control={<Input className={input} type="number" value={dbg.watchdogSeconds || ''} placeholder="1800"
                    onChange={(e) => { setDbg((s) => ({ ...s, watchdogSeconds: Number(e.target.value) || 0 })); markDirty('debug') }} />} />
                <SettingRow id="debug.fglserver" label="FGLServer"
                  description="留空 = 由 T100 按 SSH 来源 IP 自动设置;自定义时才填。"
                  control={<Input className={input} value={dbg.fglserver} placeholder="留空 = 自动"
                    onChange={(e) => { setDbg((s) => ({ ...s, fglserver: e.target.value })); markDirty('debug') }} />} />
                <SettingRow id="debug.printElements" label="print 数组元素上限"
                  description="fgldb 单次 print 的数组元素上限。清空 = 内置默认 1000。"
                  control={<Input className={input} type="number" value={dbg.printElements || ''} placeholder="1000"
                    onChange={(e) => { setDbg((s) => ({ ...s, printElements: Number(e.target.value) || 0 })); markDirty('debug') }} />} />
                <SettingRow id="debug.term" label="终端尺寸(宽 × 高)"
                  description="驱动 PTY 时用的终端大小,影响远端输出换行。"
                  control={
                    <div className="flex items-center gap-1.5">
                      <Input className={input} type="number" value={dbg.termWidth || ''} placeholder="200"
                        onChange={(e) => { setDbg((s) => ({ ...s, termWidth: Number(e.target.value) || 0 })); markDirty('debug') }} />
                      <span className="text-muted-foreground">×</span>
                      <Input className={input} type="number" value={dbg.termHeight || ''} placeholder="50"
                        onChange={(e) => { setDbg((s) => ({ ...s, termHeight: Number(e.target.value) || 0 })); markDirty('debug') }} />
                    </div>
                  } />
                <SettingRow id="debug.persistBreakpoints" label="断点持久化"
                  description="停站断点按 模块/作业 存到数据目录,重开会话时恢复。"
                  control={
                    <Select value={dbg.persistBreakpoints}
                      onValueChange={(v) => { setDbg((s) => ({ ...s, persistBreakpoints: v as DebugForm['persistBreakpoints'] })); markDirty('debug') }}>
                      <SelectTrigger className="h-7 w-full text-xs"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="default">跟随默认(开启)</SelectItem>
                        <SelectItem value="on">开启</SelectItem>
                        <SelectItem value="off">关闭</SelectItem>
                      </SelectContent>
                    </Select>
                  } />
              </Card>

              <Card id="card-debug-env" title="默认环境"
                description="调试会话未显式指定环境时用哪个。与「站点管理」的默认环境是两层:这里是调试工具的覆盖,留空即继承站点默认。"
                right={<SaveBtn keys={['debug']} />}>
                <SettingRow id="debug.activeEnv" label="调试默认环境"
                  control={
                    <Select value={dbg.activeEnv || '__inherit__'}
                      onValueChange={(v) => { setDbg((s) => ({ ...s, activeEnv: v === '__inherit__' ? '' : v })); markDirty('debug') }}>
                      <SelectTrigger className="h-7 w-full text-xs"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="__inherit__">继承站点默认({cfg.activeEnv || '未设置'})</SelectItem>
                        {envNames.map((n) => <SelectItem key={n} value={n}>{n}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  } />
                <p className="border-l-2 border-amber-500/50 bg-amber-500/5 px-2 py-1 text-[11px] text-muted-foreground">
                  注意:「调试参数」与「默认环境」两张卡改的是**同一个** debug 配置节,任一张点保存都会提交整节
                  (后端的 PUT 是整节替换)。所以两处会一起落盘,不会互相覆盖。
                </p>
              </Card>
            </>
          )}

          {/* ============ 应用设置 ============ */}
          {section === 'app' && (
            <>
              <Card id="card-app-theme" title="外观"
                description="各视图共用同一份选择:这里切换后,整个界面都跟着变。">
                <ThemeCards value={theme} onChange={setTheme} />
              </Card>

              <Card id="card-app-service" title="服务"
                description="统一 Web 服务的监听地址 —— 这套 SPA(调试工作台)就挂在它上面。"
                right={<SaveBtn keys={['listen']} />}>
                <SettingRow id="listen" label="监听地址"
                  description="改完需要重启 tt serve 才生效(监听只在启动时绑定一次)。端口被占用时会自动向后顺延。"
                  control={<Input className={input} value={listen} placeholder="127.0.0.1:28670"
                    onChange={(e) => { setListen(e.target.value); markDirty('listen') }} />} />
              </Card>

              <Card id="card-app-install" title="命令行集成"
                description="把 tt.exe 所在目录加入用户 PATH,之后在任意终端直接敲 tt。只改当前用户的环境变量,不需要管理员。">
                <SettingRow
                  id="app.install"
                  label="用户 PATH"
                  description={
                    !status?.install.supported
                      ? (status?.install.note || '本平台不支持自动写入')
                      : status.install.inUserPath
                        ? '已在用户 PATH 中'
                        : '尚未加入 —— 加完之后新开的终端才能直接敲 tt'
                  }
                  control={
                    <div className="flex items-center gap-1.5">
                      {status?.install.inUserPath ? (
                        <Button variant="outline" size="sm" disabled={installBusy}
                          onClick={() => void doInstall(false)}>
                          <Trash2 className="h-3.5 w-3.5" />{installBusy ? '处理中…' : '从 PATH 移除'}
                        </Button>
                      ) : (
                        <Button size="sm" disabled={installBusy || !status?.install.supported}
                          onClick={() => void doInstall(true)}>
                          <Terminal className="h-3.5 w-3.5" />{installBusy ? '处理中…' : '加入用户 PATH'}
                        </Button>
                      )}
                    </div>
                  }
                />
                {status?.install.exeDir && <InfoRow label="程序目录" value={status.install.exeDir} />}
                {!status?.install.supported && status?.install.manual && (
                  <pre className="overflow-x-auto border border-border bg-muted/30 p-2 text-[11px]">{status.install.manual}</pre>
                )}
                {installNote && (
                  <p className="text-[11px] text-muted-foreground">{installNote}</p>
                )}
              </Card>

              <Card id="card-app-info" title="运行信息">
                <InfoRow label="版本" value={cfg.version || '(未知)'} />
                <InfoRow label="配置文件" value={cfg.config} />
                <InfoRow label="缺省配置位置" value={meta?.defaultConfig || '—'} />
                <InfoRow label="便携模式" value={meta ? (meta.portable ? '是(配置留在程序目录)' : '否') : '—'} />
                <InfoRow label="统一工具目录" value={meta?.toolsHome || '—'} />
                <InfoRow label="配置结构版本" value={meta ? String(meta.schemaVersion) : '—'} />
                <InfoRow label="支持的库类型" value={meta?.supportedTypes?.join(' / ') || '—'} />
                <InfoRow label="默认环境" value={cfg.activeEnv || '(未设置)'} />
                <InfoRow label="环境数" value={String(cfg.sshs?.length ?? 0)} />
              </Card>

              <Card id="card-app-cache" title="缓存">
                <p className="text-[11px] text-muted-foreground">
                  跑出来的可再生数据(企业快照、源码镜像、查询落盘、断点存档)。它们与 config.json
                  同目录,但性质相反 —— 清除只动这些,配置文件一律不碰。
                </p>
                <InfoRow label="缓存目录" value={status?.cache?.dir || '—'} />
                {(status?.cache?.dirs ?? []).map((d) => (
                  <div key={d.name} className="flex items-center justify-between gap-2 text-[11px]">
                    <span className="font-mono">{d.name}</span>
                    <span className="text-muted-foreground">
                      {d.exists ? `${fmtBytes(d.bytes)} · ${d.files} 个文件` : '无'}
                    </span>
                  </div>
                ))}
                <InfoRow
                  label="共"
                  value={`${fmtBytes(status?.cache?.bytes ?? 0)} · ${status?.cache?.files ?? 0} 个文件`}
                />
                <div className="flex flex-wrap items-center gap-2 pt-1">
                  <Button size="sm" variant="outline" disabled={cacheBusy}
                    onClick={() => void doClearCache()}>
                    <Trash2 className="mr-1 h-3.5 w-3.5" />
                    {cacheBusy ? '清理中…' : '清除缓存'}
                  </Button>
                  {cacheNote && <span className="text-[11px] text-muted-foreground">{cacheNote}</span>}
                </div>
                <p className="text-[11px] text-muted-foreground">
                  tt serve 启动时会自动清理超过{' '}
                  {Math.max(1, Math.round((status?.cache?.maxAgeHours ?? 168) / 24))} 天的缓存
                  —— 刻意不"启动即全清":查询落盘的那些文件正是"刚才那条查询的完整结果",
                  启动就删会把上一条命令刚告诉你的东西删掉。
                </p>
              </Card>
            </>
          )}
        </div>
      </div>

      {/* 删除环境二次确认 */}
      <AlertDialog open={delTarget !== null} onOpenChange={(o) => { if (!o) setDelTarget(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除环境「{delTarget !== null ? sshName(sshs[delTarget]) : ''}」?</AlertDialogTitle>
            <AlertDialogDescription>
              该环境的 SSH 连接、登录区域、TOPENT 与数据库账号配置将一并移除。
              点击「删除」后立即写入 config.json,不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void confirmDelSsh()}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
