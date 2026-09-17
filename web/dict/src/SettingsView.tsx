// TDict 配置页:可视化维护 config.json 的 SSH 环境与数据库连接。
//
// 布局与交互模仿 TDebug 设置页「环境」:左侧环境列表 + 右侧「SSH 服务器 / 数据库」
// 两个 Tab 编辑同一个环境(一对一挂载)。SSH 连接与「从服务器获取数据库配置」由后端
// tdict/host 承担(与 TDebug 同源):host.Dial 连服务器 → 按 zone 加载 T100 环境 →
// oracle 读 tnsnames / kingbase 实例发现,解析出 host/port/service(库名)。
//
// 环境清单走**共享**端点 /api/hosts(顶层,不带 API_BASE 前缀):一个进程同时挂着
// 两套 SPA,而环境只有一份数据源 —— config.json 的 hosts 节,调试工作台的环境页
// 读写的也是它。合并前这里是本包私有的 /api/config,现在两边看到的是同一份清单。
import { useCallback, useEffect, useRef, useState } from 'react'
import {
  Plus, Trash2, Server, Database, Eye, EyeOff, RefreshCw, Zap, Save, Star, AlertCircle, CheckCircle2,
} from 'lucide-react'
import { api, type DBAcct, type DBConnection, type SshEnv } from './api'
import { Button, cn, Field, Input, SectionTitle, Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui'

type EnvTab = 'ssh' | 'db'

// 可编辑表格的单元格:聚焦时给整格一圈内描边,提示「正在编辑这一格」
const CELL_EDITING = 'focus-within:ring-1 focus-within:ring-inset focus-within:ring-sky-500/60'

// 格内密码可见性切换(浮在 TableCell 右上,故父格需 relative)
function PwdEye({ shown, onClick }: { shown: boolean; onClick: () => void }) {
  return (
    <button type="button" title={shown ? '隐藏密码' : '显示密码'} onClick={onClick}
      className="absolute right-0.5 top-1/2 flex h-6 w-6 -translate-y-1/2 items-center justify-center text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100">
      {shown ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
    </button>
  )
}

// T100 标准 schema 账号(密码=账号):新增环境时预置,可自由增删改
const DEFAULT_ACCOUNTS: DBAcct[] = [
  { account: 'ds', password: 'ds' },
  { account: 'dsdata', password: 'dsdata' },
  { account: 'dsdemo', password: 'dsdemo' },
]

const blankDb = (): DBConnection => ({
  type: 'oracle', host: '', port: 1521, service: '', database: '',
  accounts: DEFAULT_ACCOUNTS.map((a) => ({ ...a })), // 复制,避免各环境共享同一数组
})
const blankSsh = (): SshEnv => ({
  name: '', host: '', port: 22, user: '', password: '', zone: '36', topent: '',
  db: { ...blankDb() },
})

const sshName = (e: SshEnv) => e.name || `${e.host}-${e.zone || ''}`.replace(/-$/, '')

export function SettingsView() {
  const [cfgPath, setCfgPath] = useState('')
  const [activeEnv, setActiveEnv] = useState('')
  const [sshs, setSshs] = useState<SshEnv[]>([])
  const [sel, setSel] = useState(0)
  const [tab, setTab] = useState<EnvTab>('ssh')
  const [loaded, setLoaded] = useState(false)
  const [err, setErr] = useState('')
  const [saveState, setSaveState] = useState<'idle' | 'saving' | 'saved'>('idle')
  const dirtyRef = useRef(false)
  const [dirty, setDirty] = useState(false)
  const [showPwd, setShowPwd] = useState(false)
  // DB Tab 操作状态
  const [busy, setBusy] = useState('')
  const [note, setNote] = useState<{ kind: 'ok' | 'err' | 'info'; text: string } | null>(null)
  const [accProbe, setAccProbe] = useState<{ i: number; state: 'testing' | 'ok' | 'err'; msg: string } | null>(null)
  const [addAcct, setAddAcct] = useState({ account: '', password: '' })

  const markDirty = () => { dirtyRef.current = true; setDirty(true) }
  const cleanDirty = () => { dirtyRef.current = false; setDirty(false) }

  const load = useCallback(() => {
    api.hosts()
      .then((c) => {
        setCfgPath(c.config)
        setActiveEnv(c.activeEnv || '')
        setSshs((c.sshs || []).map((e) => ({
          name: e.name || '', host: e.host || '', port: e.port || 22,
          user: e.user || '', password: e.password || '',
          zone: e.zone || '', topent: e.topent != null ? String(e.topent) : '',
          db: e.db ? {
            type: e.db.type || 'oracle', host: e.db.host || '', port: e.db.port || 0,
            service: e.db.service || '', database: e.db.database || '',
            accounts: (e.db.accounts || []).map((a) => ({ account: a.account || '', password: a.password || '' })),
            readonlySql: e.db.readonlySql === false ? false : undefined,
          } : null,
        })))
        cleanDirty(); setSaveState('idle'); setErr('')
      })
      .catch((e) => setErr(e.message))
      .finally(() => setLoaded(true))
  }, [])

  useEffect(() => { load() }, [load])

  // ---- 统一保存(整环境:SSH 字段 + db;list 可覆盖当前列表,删除/设为默认后立即落盘用) ----
  const saveAll = async (list?: SshEnv[], def?: string) => {
    if (saveState === 'saving') return
    const src = list || sshs
    if (!src.length) { setErr('至少保留一个 SSH 环境'); return }
    setErr(''); setSaveState('saving')
    try {
      const target = def ?? activeEnv
      const sshsOut: SshEnv[] = src.filter((x) => x.host.trim()).map((x) => {
        const o: SshEnv = {
          name: sshName(x), host: x.host.trim(), port: x.port || 22,
          user: x.user.trim(), password: x.password,
        }
        if (x.zone?.trim()) o.zone = x.zone.trim()
        if (x.topent?.trim()) o.topent = x.topent.trim()
        if (x.db) {
          const db: DBConnection = { type: x.db.type || 'oracle', host: x.db.host.trim(), port: x.db.port || 0 }
          if (x.db.type === 'oracle') { if (x.db.service?.trim()) db.service = x.db.service.trim() }
          else { if (x.db.database?.trim()) db.database = x.db.database.trim() }
          const accounts = (x.db.accounts || []).map((a) => ({ account: a.account.trim(), password: a.password })).filter((a) => a.account)
          if (accounts.length) db.accounts = accounts
          // 只读 SQL 开关由调试工作台的环境页维护,本页没有对应控件 —— 但两页写同一个
          // hosts 节,原样带上,否则在本页点一次保存就把用户关掉的开关悄悄打开了
          if (x.db.readonlySql === false) db.readonlySql = false
          o.db = db
        }
        return o
      })
      const activeOut = sshsOut.some((e) => e.name === target) ? target : (sshsOut[0]?.name || '')
      // 只提交 hosts 节:query/mirror/bdldoc/sync 等节后端保持原样,
      // 不会被这里的陈旧快照覆盖
      await api.saveHosts({ hosts: { activeEnv: activeOut, sshs: sshsOut } })
      setActiveEnv(activeOut)
      setSshs((prev) => prev.map((x) => ({ ...x, name: sshName(x) })))
      cleanDirty(); setSaveState('saved')
    } catch (ex) {
      // 服务端是校验的权威(环境名重复 / 端口越界 / 至少要有一个环境…),
      // 它给的中文说明原样显示,不要用本地判断把它盖掉
      setErr(ex instanceof Error ? ex.message : String(ex)); setSaveState('idle')
    }
  }

  const patchSsh = (i: number, patch: Partial<SshEnv>) => {
    setSshs((prev) => prev.map((x, j) => (j === i ? { ...x, ...patch } : x)))
    markDirty()
  }
  const patchDb = (i: number, patch: Partial<DBConnection>) => {
    const cur = sshs[i]
    if (!cur) return
    patchSsh(i, { db: cur.db ? { ...cur.db, ...patch } : { ...blankDb(), ...patch } })
  }
  const patchAcct = (i: number, j: number, patch: Partial<{ account: string; password: string }>) => {
    const cur = sshs[i]
    if (!cur?.db) return
    patchDb(i, { accounts: (cur.db.accounts || []).map((a, k) => (k === j ? { ...a, ...patch } : a)) })
  }
  const delAcct = (i: number, j: number) => {
    const cur = sshs[i]
    if (!cur?.db) return
    patchDb(i, { accounts: (cur.db.accounts || []).filter((_, k) => k !== j) })
  }
  const pushAcct = () => {
    const cur = sshs[sel]
    if (!cur?.db) return
    const account = addAcct.account.trim()
    if (!account) return
    patchDb(sel, { accounts: [...(cur.db.accounts || []), { account, password: addAcct.password }] })
    setAddAcct({ account: '', password: '' })
  }
  const addSsh = () => { setSel(sshs.length); setSshs([...sshs, blankSsh()]); markDirty() }
  const delSsh = async (i: number) => {
    const cur = sshs[i]
    if (!cur) return
    const label = sshName(cur) || '(未命名环境)'
    if (!window.confirm(`删除环境「${label}」?\n该环境的 SSH 与数据库配置将一并移除,并立即写入 config.json,不可撤销。`)) return
    const next = sshs.filter((_, j) => j !== i)
    setSel(Math.max(0, Math.min(i, next.length - 1)))
    setSshs(next)
    await saveAll(next)
  }
  // 设为默认环境并立即落盘(mirror/db sync/--conn 缺省取它)
  const setDefault = async (name: string) => {
    if (!name || name === activeEnv) return
    setActiveEnv(name)
    await saveAll(sshs, name)
  }

  // ---- DB Tab 操作(针对当前环境的 db) ----
  // 从服务器获取:登录当前环境 SSH 探测连接要素回填(host.ProbeDBConfig,只读)
  const fetchFromServer = async () => {
    const s = sshs[sel]
    const d = s?.db
    if (!s?.host || !s?.user) { setNote({ kind: 'err', text: '请先填写 SSH 主机与账号' }); return }
    if (!d) return
    setBusy('fetch'); setNote(null)
    try {
      const r = await api.probeDB({ host: s.host, port: s.port || 22, user: s.user, password: s.password, zone: s.zone || '', type: d.type || 'oracle' })
      const patch: Partial<DBConnection> = {}
      // 服务器解析出 IP 直接用;内部主机名客户端多半不可达,回填 SSH 主机(T100 库通常与应用服务器同机)
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
      if (Object.keys(patch).length) patchDb(sel, patch)
    } catch (ex) {
      setNote({ kind: 'err', text: '获取失败: ' + (ex instanceof Error ? ex.message : String(ex)) })
    } finally { setBusy('') }
  }
  // 账号行验证:服务器侧以 账号+密码 连当前环境的显式目标库(host.VerifyDBAcct,只读)
  const verifyAcct = async (j: number) => {
    const s = sshs[sel]
    const d = s?.db
    const row = d?.accounts?.[j]
    if (!s?.host || !s?.user) { setAccProbe({ i: j, state: 'err', msg: '请先填写 SSH 主机与账号' }); return }
    if (!d?.host.trim()) { setAccProbe({ i: j, state: 'err', msg: '请先填写数据库主机地址' }); return }
    if (!row?.account.trim()) { setAccProbe({ i: j, state: 'err', msg: '账号不能为空' }); return }
    setAccProbe({ i: j, state: 'testing', msg: '' })
    try {
      await api.dbAccVerify({
        host: s.host, port: s.port || 22, user: s.user, password: s.password, zone: s.zone || '',
        type: d.type || 'oracle', account: row.account.trim(), acctPassword: row.password,
        dbHost: d.host.trim(), dbPort: d.port || 0,
        ...(d.type === 'kingbase' ? { dbDatabase: (d.database || '').trim() } : { dbSvc: (d.service || '').trim() }),
      })
      setAccProbe({ i: j, state: 'ok', msg: '连接正常 ✓' })
    } catch (ex) {
      setAccProbe({ i: j, state: 'err', msg: ex instanceof Error ? ex.message : String(ex) })
    }
  }
  // 客户端直连测试(erpdb.Open,与 db ping 同链路)
  const testConn = async () => {
    const d = sshs[sel]?.db
    if (!d) return
    setBusy('conn'); setNote(null)
    try {
      const r = await api.connTest(d)
      setNote({ kind: 'ok', text: '连接正常: ' + (r.serverVersion || '').split('\n')[0] })
    } catch (ex) {
      setNote({ kind: 'err', text: '连接失败: ' + (ex instanceof Error ? ex.message : String(ex)) })
    } finally { setBusy('') }
  }

  // 「已保存」提示短暂停留后回到空闲
  useEffect(() => {
    if (saveState !== 'saved') return
    const t = window.setTimeout(() => setSaveState('idle'), 2000)
    return () => window.clearTimeout(t)
  }, [saveState])

  const cur = sshs[sel]
  const curDb = cur?.db
  const acctList = curDb?.accounts || []

  return (
    <div className="flex h-full min-h-0 flex-col bg-zinc-50 text-zinc-800 dark:bg-zinc-950 dark:text-zinc-100">
      {/* 顶栏:标题 + 配置文件路径 + 保存状态/按钮 */}
      <header className="flex shrink-0 items-center gap-3 border-b border-zinc-200 bg-white px-4 py-2 dark:border-zinc-800 dark:bg-zinc-900">
        <h1 className="shrink-0 text-sm font-semibold">TDict 配置</h1>
        <span className="min-w-0 flex-1 truncate text-[11px] text-zinc-500 dark:text-zinc-400" title={cfgPath}>
          {cfgPath || 'config.json'}
        </span>
        {err && <span className="flex shrink-0 items-center gap-1 text-[11px] text-red-600 dark:text-red-400"><AlertCircle className="h-3.5 w-3.5" />{err}</span>}
        {!err && (
          <span className={cn('shrink-0 text-[11px]',
            saveState === 'saved' ? 'text-emerald-600 dark:text-emerald-400'
              : dirty ? 'text-amber-600 dark:text-amber-400' : 'text-zinc-400 dark:text-zinc-600')}>
            {saveState === 'saving' ? '保存中…' : saveState === 'saved' ? '已保存 ✓' : dirty ? '有未保存的修改' : ''}
          </span>
        )}
        <Button variant="primary" disabled={!dirty || saveState === 'saving'} onClick={() => void saveAll()}>
          <Save className="h-3.5 w-3.5" />{saveState === 'saving' ? '保存中…' : '保存'}
        </Button>
      </header>

      <div className="flex min-h-0 flex-1">
        {/* 环境列表 */}
        <aside className="flex w-56 shrink-0 flex-col overflow-auto border-r border-zinc-200 bg-white p-2 dark:border-zinc-800 dark:bg-zinc-900">
          {!loaded && <div className="p-2 text-xs text-zinc-500">加载中…</div>}
          {loaded && sshs.length === 0 && <div className="p-2 text-xs text-zinc-500">(无环境，点下方新增)</div>}
          <div className="space-y-0.5">
            {sshs.map((e, i) => (
              <button key={i} type="button" onClick={() => setSel(i)}
                className={cn('flex w-full items-center gap-1.5 px-2 py-1.5 text-left transition-colors',
                  sel === i ? 'bg-zinc-100 text-zinc-900 dark:bg-zinc-800 dark:text-zinc-50' : 'hover:bg-zinc-50 dark:hover:bg-zinc-800/60')}>
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-xs font-medium">{e.name || e.host || '(新环境)'}</span>
                  <span className="block truncate text-[11px] text-zinc-500 dark:text-zinc-400">
                    {e.zone ? `${e.zone} · ` : ''}{e.port || 22}{e.db ? ' · 库' : ''}
                  </span>
                </span>
                {activeEnv === sshName(e) && <Star className="h-3.5 w-3.5 shrink-0 fill-amber-400 text-amber-400" />}
              </button>
            ))}
          </div>
          <Button variant="outline" className="mt-2 w-full" onClick={addSsh}>
            <Plus className="h-3.5 w-3.5" />新增环境
          </Button>
        </aside>

        {/* 表单区 */}
        <main className="min-h-0 flex-1 overflow-auto p-4">
          {!cur && loaded && (
            <div className="mx-auto max-w-3xl border border-dashed border-zinc-300 p-6 text-xs text-zinc-500 dark:border-zinc-700">
              还没有 SSH 环境。点击左侧「新增环境」填写服务器连接,再在「数据库」Tab 配置该环境挂载的库。
            </div>
          )}
          {cur && (
            <div className="mx-auto max-w-3xl">
              <div className="mb-2 flex items-center gap-2">
                <h2 className="text-sm font-medium">环境参数</h2>
                {activeEnv === sshName(cur)
                  ? <span className="flex items-center gap-1 text-[11px] text-amber-600 dark:text-amber-400"><Star className="h-3.5 w-3.5 fill-amber-400 text-amber-400" />默认环境</span>
                  : <Button variant="ghost" size="xs" title="设为默认环境(mirror / db sync / --conn 缺省取它)" onClick={() => void setDefault(sshName(cur))}>
                      <Star className="h-3.5 w-3.5" />设为默认
                    </Button>}
                <Button variant="ghost" className="ml-auto text-zinc-500 hover:text-red-600 dark:hover:text-red-400"
                  title="删除该环境(确认后立即保存)" onClick={() => void delSsh(sel)}>
                  <Trash2 className="h-3.5 w-3.5" />删除环境
                </Button>
              </div>

              {/* SSH / 数据库 两个 Tab 编辑同一环境 */}
              <div className="mb-3 flex items-center gap-1 border-b border-zinc-200 dark:border-zinc-800">
                {([{ key: 'ssh' as const, label: 'SSH 服务器', Icon: Server }, { key: 'db' as const, label: '数据库', Icon: Database }]).map(({ key, label, Icon }) => (
                  <button key={key} type="button" onClick={() => setTab(key)}
                    className={cn('relative flex items-center gap-1.5 px-3 py-1.5 text-xs transition-colors',
                      tab === key ? 'text-zinc-900 dark:text-zinc-50' : 'text-zinc-500 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-100')}>
                    <Icon className="h-3.5 w-3.5" />{label}
                    {tab === key && <span className="absolute inset-x-0 bottom-0 h-px bg-sky-500" />}
                  </button>
                ))}
              </div>

              {tab === 'ssh' && (
                <div className="grid grid-cols-2 gap-2">
                  <Field label="环境名称(留空自动为主机-区域)" className="col-span-2">
                    <Input value={cur.name} placeholder="如 正式区"
                      onChange={(e) => patchSsh(sel, { name: e.target.value })} />
                  </Field>
                  <Field label="IP 主机"><Input value={cur.host} placeholder="如 10.0.0.1" onChange={(e) => patchSsh(sel, { host: e.target.value })} /></Field>
                  <Field label="端口"><Input type="number" value={cur.port || ''} placeholder="22" onChange={(e) => patchSsh(sel, { port: Number(e.target.value) || 22 })} /></Field>
                  <Field label="登录区域(zone)"><Input value={cur.zone || ''} placeholder="31开发/35测试/36正式/39PATCH/t出货" onChange={(e) => patchSsh(sel, { zone: e.target.value })} /></Field>
                  <Field label="账号"><Input value={cur.user} placeholder="如 youruser" onChange={(e) => patchSsh(sel, { user: e.target.value })} /></Field>
                  <Field label="密码">
                    <div className="relative">
                      <Input className="pr-8" type={showPwd ? 'text' : 'password'} value={cur.password}
                        onChange={(e) => patchSsh(sel, { password: e.target.value })} />
                      <button type="button" title={showPwd ? '隐藏密码' : '显示密码'} onClick={() => setShowPwd((v) => !v)}
                        className="absolute inset-y-0 right-0.5 flex w-6 items-center justify-center text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-100">
                        {showPwd ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
                      </button>
                    </div>
                  </Field>
                  <Field label="TOPENT(默认企业,数字或文本)" className="col-span-2">
                    <Input value={cur.topent || ''} placeholder="如 99 / YOURENT" onChange={(e) => patchSsh(sel, { topent: e.target.value })} />
                  </Field>
                  <p className="col-span-2 text-[11px] text-zinc-500 dark:text-zinc-400">
                    SSH 用于登录 T100 服务器并按区域探测环境;该环境一对一挂载的数据库在「数据库」Tab 维护。
                  </p>
                </div>
              )}

              {tab === 'db' && (
                <div>
                  {!curDb ? (
                    <div className="flex flex-col items-start gap-2 border border-dashed border-zinc-300 p-4 text-xs text-zinc-500 dark:border-zinc-700">
                      该环境未配置数据库(远程直查 / db sync / db discover 需要)。
                      <Button variant="outline" onClick={() => patchDb(sel, {})}>
                        <Plus className="h-3.5 w-3.5" />添加数据库
                      </Button>
                    </div>
                  ) : (
                    <div className="grid grid-cols-2 gap-2">
                      <div className="col-span-2 flex items-center justify-end gap-1.5">
                        <Button variant="outline" disabled={busy === 'fetch' || !cur.host || !cur.user} onClick={() => void fetchFromServer()}>
                          <RefreshCw className={cn('h-3.5 w-3.5', busy === 'fetch' && 'animate-spin')} />从服务器获取数据库配置
                        </Button>
                        <Button variant="outline" disabled={busy === 'conn' || !curDb.host} onClick={() => void testConn()}>
                          <Zap className={cn('h-3.5 w-3.5', busy === 'conn' && 'animate-pulse')} />测试连接
                        </Button>
                      </div>
                      <Field label="类型">
                        <select value={curDb.type || 'oracle'} onChange={(e) => patchDb(sel, { type: e.target.value })}
                          className="h-7 w-full min-w-0 border border-zinc-300 bg-white px-2 text-xs text-zinc-800 outline-none focus:border-sky-500 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-100">
                          <option value="oracle">Oracle</option>
                          <option value="kingbase">人大金仓(PG 引擎)</option>
                        </select>
                      </Field>
                      <Field label="主机地址"><Input placeholder="客户端与服务器均可达" value={curDb.host} onChange={(e) => patchDb(sel, { host: e.target.value })} /></Field>
                      <Field label="端口"><Input type="number" placeholder={curDb.type === 'oracle' ? '1521' : '54321'} value={curDb.port || ''} onChange={(e) => patchDb(sel, { port: Number(e.target.value) || 0 })} /></Field>
                      {curDb.type === 'oracle' ? (
                        <Field label="服务名 (SERVICE_NAME)"><Input placeholder="YOUR_SERVICE" value={curDb.service || ''} onChange={(e) => patchDb(sel, { service: e.target.value })} /></Field>
                      ) : (
                        <Field label="库名 (database)"><Input placeholder="your_database" value={curDb.database || ''} onChange={(e) => patchDb(sel, { database: e.target.value })} /></Field>
                      )}

                      {/* 账号列表(无主账号;客户端直连取首项,服务器侧按 TOPENT 解析) */}
                      <div className="col-span-2 mt-2">
                        <SectionTitle>账号列表(账号=schema)</SectionTitle>
                        {/* 账号表:可编辑表格(每格直接是输入框,网格线由 TableCell 承担) */}
                        <Table className="mt-1.5">
                          <TableHeader>
                            <TableRow>
                              <TableHead className="w-40">账号(schema)</TableHead>
                              <TableHead>密码(缺省=账号)</TableHead>
                              <TableHead className="w-20 text-center">操作</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {/* 新增行放第一行:Enter 或「添加」落到下方列表末尾 */}
                            <TableRow>
                              <TableCell className={CELL_EDITING}>
                                <input className="cell-input font-mono" placeholder="如 ds" value={addAcct.account}
                                  onChange={(e) => setAddAcct({ ...addAcct, account: e.target.value })}
                                  onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); pushAcct() } }} />
                              </TableCell>
                              <TableCell className={cn('relative', CELL_EDITING)}>
                                <input className="cell-input pr-7 font-mono" type={showPwd ? 'text' : 'password'} placeholder="密码(缺省=账号)" value={addAcct.password}
                                  onChange={(e) => setAddAcct({ ...addAcct, password: e.target.value })}
                                  onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); pushAcct() } }} />
                                <PwdEye shown={showPwd} onClick={() => setShowPwd((v) => !v)} />
                              </TableCell>
                              <TableCell className="text-center">
                                <Button variant="outline" size="xs" disabled={!addAcct.account.trim()} onClick={pushAcct}>
                                  <Plus className="h-3 w-3" />添加
                                </Button>
                              </TableCell>
                            </TableRow>
                            {acctList.map((a, j) => (
                              <TableRow key={j}>
                                <TableCell className={CELL_EDITING}>
                                  <input className="cell-input font-mono" placeholder="如 ds" value={a.account}
                                    onChange={(e) => patchAcct(sel, j, { account: e.target.value })} />
                                </TableCell>
                                <TableCell className={cn('relative', CELL_EDITING)}>
                                  <input className="cell-input pr-7 font-mono" type={showPwd ? 'text' : 'password'} placeholder="密码(缺省=账号)" value={a.password}
                                    onChange={(e) => patchAcct(sel, j, { password: e.target.value })} />
                                  <PwdEye shown={showPwd} onClick={() => setShowPwd((v) => !v)} />
                                </TableCell>
                                <TableCell className="text-center">
                                  <div className="flex items-center justify-center gap-1">
                                    <button type="button" title="服务器上以该账号+密码连目标库验证(只读)"
                                      className={cn('flex h-6 w-6 items-center justify-center',
                                        accProbe?.i === j && accProbe.state === 'testing' ? 'animate-pulse text-sky-500' : 'text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100')}
                                      onClick={() => void verifyAcct(j)}>
                                      <Zap className="h-3.5 w-3.5" />
                                    </button>
                                    <button type="button" title="删除该账号"
                                      className="flex h-6 w-6 items-center justify-center text-zinc-500 hover:text-red-600 dark:hover:text-red-400"
                                      onClick={() => delAcct(sel, j)}>
                                      <Trash2 className="h-3.5 w-3.5" />
                                    </button>
                                  </div>
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                        {accProbe && acctList[accProbe.i] && (
                          <div className={cn('mt-1 flex items-center gap-1 text-[11px]',
                            accProbe.state === 'ok' ? 'text-emerald-600 dark:text-emerald-400'
                              : accProbe.state === 'err' ? 'text-red-600 dark:text-red-400' : 'text-sky-600 dark:text-sky-400')}>
                            {accProbe.state === 'ok' ? <CheckCircle2 className="h-3 w-3" /> : <AlertCircle className="h-3 w-3" />}
                            账号 {acctList[accProbe.i].account}: {accProbe.msg || '验证中…'}
                          </div>
                        )}
                        {!acctList.length && (
                          <div className="mt-1 text-[11px] text-zinc-500 dark:text-zinc-400">
                            客户端直连(远程直查 / db ping / db sync)取列表首项;未收录账号按「账号=密码」兜底。
                          </div>
                        )}
                      </div>

                      {note && (
                        <div className={cn('col-span-2 mt-1 px-3 py-2 text-xs',
                          note.kind === 'ok' ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
                            : note.kind === 'err' ? 'bg-red-500/10 text-red-600 dark:text-red-400'
                              : 'bg-sky-500/10 text-sky-700 dark:text-sky-300')}>
                          {note.text}
                        </div>
                      )}
                      <p className="col-span-2 text-[11px] text-zinc-500 dark:text-zinc-400">
                        「从服务器获取」登录该环境 SSH 只读探测:oracle 解析 tnsnames 的服务名/端口/地址,金仓发现实例库名/端口;服务器执行工具(sqlplus/ksql)路径自动探测,无需配置。
                      </p>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
