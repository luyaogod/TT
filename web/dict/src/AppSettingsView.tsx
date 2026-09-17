// 设置页:命令行安装(把可执行文件加入用户 PATH)、BDL 语言文档目录,以及运行信息。
//
// 数据来源分两类:PATH 安装状态读共享端点 /api/install(合并后已从本包私有端点移走),
// BDL 文档目录读本包私有端点(
// /dict/api/bdldoc);配置文件路径、监听地址、环境名与查询数据源读**共享**的
// GET /api/hosts,运行信息(版本)读共享的 GET /api/health。写入则一律走
// PUT /api/hosts 的对应节 —— 只发改过的那一节,后端对省略的节保持原样。
import { useCallback, useEffect, useRef, useState } from 'react'
import { Terminal, CheckCircle2, AlertCircle, Plus, Trash2, BookOpen, Save, Database } from 'lucide-react'
import { api, type BdldocStatus, type HostsView, type InstallStatus } from './api'
import { Button, cn, Field, Input, SectionTitle } from './ui'

export function AppSettingsView() {
  const [st, setSt] = useState<InstallStatus | null>(null)
  const [configPath, setConfigPath] = useState('')
  const [listen, setListen] = useState('')
  const [version, setVersion] = useState('')
  const [bd, setBd] = useState<BdldocStatus | null>(null)
  const [cfg, setCfg] = useState<HostsView | null>(null)
  const [srcSel, setSrcSel] = useState('')
  const srcInit = useRef(false)
  const [bdInput, setBdInput] = useState('')
  const bdInit = useRef(false)
  const [err, setErr] = useState('')
  const [notice, setNotice] = useState('')
  const [busy, setBusy] = useState('')

  const refresh = useCallback(async () => {
    try {
      const [s, cfg, health, b] = await Promise.all([api.installStatus(), api.hosts(), api.health(), api.bdldoc()])
      setSt(s)
      setCfg(cfg)
      setConfigPath(cfg.config)
      if (!srcInit.current) { setSrcSel(cfg.query?.source || ''); srcInit.current = true }
      setListen(cfg.listen)
      setVersion(health.version)
      setBd(b)
      if (!bdInit.current) { setBdInput(b.dir); bdInit.current = true }
      setErr('')
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    }
  }, [])

  useEffect(() => { void refresh() }, [refresh])

  const add = async () => {
    setBusy('add'); setErr(''); setNotice('')
    try {
      setSt(await api.installAdd())
      setNotice('已加入用户 PATH。新开的终端可直接运行 tdict;已打开的终端需重开。')
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally { setBusy('') }
  }

  const remove = async () => {
    if (!window.confirm('从用户 PATH 中移除该目录?')) return
    setBusy('remove'); setErr(''); setNotice('')
    try {
      setSt(await api.installRemove())
      setNotice('已从用户 PATH 移除(需重开终端生效)。')
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally { setBusy('') }
  }

  const supported = !!st?.supported
  const inPath = !!st?.inUserPath

  const saveBdldoc = async () => {
    if (!bdInput.trim()) { setErr('请填写 BDL 文档目录'); return }
    setBusy('bdldoc'); setErr(''); setNotice('')
    try {
      const r = await api.saveBdldocDir(bdInput.trim())
      setBd(r)
      setBdInput(r.dir)
      setNotice('BDL 文档目录已保存到 config.json(等价 tdict bdldoc dir)。')
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally { setBusy('') }
  }

  const saveSrc = async () => {
    setBusy('src'); setErr(''); setNotice('')
    try {
      // 只传 query 节:后端对省略的节(hosts/mirror/bdldoc…)保持原样,
      // 不会用陈旧快照覆盖环境配置页刚改好的环境
      await api.saveHosts({ query: { source: srcSel } })
      setNotice(srcSel === '' ? '查询数据源已设为「在线(默认环境)」。' : srcSel === 'local'
        ? '查询数据源已设为「本地 SQLite」。' : `查询数据源已设为「在线(${srcSel})」。`)
      await refresh()
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally { setBusy('') }
  }

  const envNames = (cfg?.sshs || []).map((e) => e.name || e.host)

  return (
    <div className="flex h-full min-h-0 flex-col bg-zinc-50 text-zinc-800 dark:bg-zinc-950 dark:text-zinc-100">
      <header className="flex shrink-0 items-center gap-3 border-b border-zinc-200 bg-white px-4 py-2 dark:border-zinc-800 dark:bg-zinc-900">
        <h1 className="shrink-0 text-sm font-semibold">设置</h1>
        <span className="min-w-0 flex-1 truncate text-[11px] text-zinc-500 dark:text-zinc-400">
          查询数据源、命令行安装、BDL 语言文档目录与运行信息
        </span>
        {err && <span className="flex shrink-0 items-center gap-1 text-[11px] text-red-600 dark:text-red-400"><AlertCircle className="h-3.5 w-3.5" />{err}</span>}
      </header>

      <main className="min-h-0 flex-1 overflow-auto p-4">
        <div className="mx-auto max-w-3xl space-y-5">
          {/* 查询数据源 */}
          <section>
            <SectionTitle>查询数据源</SectionTitle>
            <div className="mt-2 border border-zinc-200 p-3 dark:border-zinc-800">
              <div className="flex items-start gap-2">
                <Database className="mt-0.5 h-4 w-4 shrink-0 text-zinc-500" />
                <div className="min-w-0 flex-1">
                  <p className="text-xs text-zinc-600 dark:text-zinc-300">
                    查询命令(r.t / r.v / desc / scc / r.q / prog)读哪里,写入 <code>config.json</code> 顶层 <code>query.source</code>:
                    <strong>在线</strong>=每次直连该环境的 ERP 库(数据最新,需要网络);<strong>本地</strong>=查 <code>erp_data.db</code> 副本(快,靠「数据同步」更新)。
                    两者**不自动切换**——选在线时连不上就直接报错,不会偷偷改查本地。
                  </p>
                  <div className="mt-2 flex items-end gap-2">
                    <Field label="查询数据源" className="min-w-0 flex-1">
                      <select value={srcSel} onChange={(e) => setSrcSel(e.target.value)}
                        className="h-7 w-full min-w-0 border border-zinc-300 bg-white px-2 text-xs text-zinc-800 outline-none focus:border-sky-500 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-100">
                        <option value="">在线（默认环境{cfg?.activeEnv ? `:${cfg.activeEnv}` : ''}）</option>
                        <option value="local">本地 SQLite（erp_data.db）</option>
                        {envNames.map((n) => <option key={n} value={n}>在线（{n}）</option>)}
                      </select>
                    </Field>
                    <Button variant="primary" disabled={busy === 'src'} onClick={() => void saveSrc()}>
                      <Save className="h-3.5 w-3.5" />{busy === 'src' ? '保存中…' : '保存'}
                    </Button>
                  </div>
                  <p className="mt-2 text-[11px] text-zinc-500 dark:text-zinc-400">
                    当前:<code>{cfg?.query?.source || '(在线 · 默认环境)'}</code> · 命令行可单次覆盖:
                    <code> --conn local</code> 或 <code>--conn &lt;环境名&gt;</code>
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* 命令行安装 */}
          <section>
            <SectionTitle>命令行安装</SectionTitle>
            <div className="mt-2 border border-zinc-200 p-3 dark:border-zinc-800">
              <div className="flex items-start gap-2">
                <Terminal className="mt-0.5 h-4 w-4 shrink-0 text-zinc-500" />
                <div className="min-w-0 flex-1">
                  <p className="text-xs text-zinc-600 dark:text-zinc-300">
                    把 tdict 所在目录加入<strong>用户 PATH</strong>(当前用户,无需管理员),之后任意位置都能直接运行 <code>tdict</code>。
                  </p>
                  <div className="mt-2 space-y-1 text-[11px] text-zinc-500 dark:text-zinc-400">
                    <div className="min-w-0 truncate" title={st?.exePath}>可执行文件:{st?.exePath || '…'}</div>
                    <div className="min-w-0 truncate" title={st?.exeDir}>将加入的目录:{st?.exeDir || '…'}</div>
                  </div>
                  <div className="mt-2 flex items-center gap-2">
                    <span className={cn('flex items-center gap-1 text-[11px]',
                      inPath ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400')}>
                      {inPath ? <CheckCircle2 className="h-3.5 w-3.5" /> : <AlertCircle className="h-3.5 w-3.5" />}
                      {inPath ? '已在用户 PATH 中' : '尚未加入用户 PATH'}
                    </span>
                  </div>
                  <div className="mt-3 flex items-center gap-2">
                    <Button variant="primary" disabled={!supported || inPath || busy !== ''} onClick={() => void add()}>
                      <Plus className="h-3.5 w-3.5" />{busy === 'add' ? '添加中…' : '添加到用户 PATH'}
                    </Button>
                    <Button variant="outline" disabled={!supported || !inPath || busy !== ''} onClick={() => void remove()}>
                      <Trash2 className="h-3.5 w-3.5" />{busy === 'remove' ? '移除中…' : '从用户 PATH 移除'}
                    </Button>
                  </div>
                  {notice && <p className="mt-2 text-[11px] text-emerald-600 dark:text-emerald-400">{notice}</p>}
                  {st?.note && <p className="mt-2 text-[11px] text-zinc-500 dark:text-zinc-400">{st.note}</p>}
                  {!supported && st?.manual && (
                    <pre className="mt-2 overflow-auto bg-zinc-100 px-3 py-2 text-[11px] text-zinc-700 dark:bg-zinc-900 dark:text-zinc-200">{st.manual}</pre>
                  )}
                  <details className="mt-2">
                    <summary className="cursor-pointer text-[11px] text-zinc-500 dark:text-zinc-400">查看当前用户 PATH</summary>
                    <pre className="mt-1 max-h-40 overflow-auto whitespace-pre-wrap break-all bg-zinc-100 px-3 py-2 text-[11px] text-zinc-700 dark:bg-zinc-900 dark:text-zinc-200">{st?.userPath || '(空)'}</pre>
                  </details>
                </div>
              </div>
            </div>
          </section>

          {/* BDL 语言文档目录 */}
          <section>
            <SectionTitle>BDL 语言文档目录</SectionTitle>
            <div className="mt-2 border border-zinc-200 p-3 dark:border-zinc-800">
              <div className="flex items-start gap-2">
                <BookOpen className="mt-0.5 h-4 w-4 shrink-0 text-zinc-500" />
                <div className="min-w-0 flex-1">
                  <p className="text-xs text-zinc-600 dark:text-zinc-300">
                    Genero BDL(4GL)语言参考文档(markdown)的存放目录,写入 <code>config.json</code> 顶层 <code>bdldoc.dir</code>,
                    供 AI/工具查语法与内置函数时定位(等价 <code>tdict bdldoc dir &lt;目录&gt;</code>)。
                  </p>
                  <div className="mt-2 flex items-end gap-2">
                    <Field label="文档目录(绝对路径)" className="min-w-0 flex-1">
                      <Input value={bdInput} placeholder="如 D:\T100\4gl文档\BDL-Markdown"
                        onChange={(e) => setBdInput(e.target.value)}
                        onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); void saveBdldoc() } }} />
                    </Field>
                    <Button variant="primary" disabled={busy === 'bdldoc'} onClick={() => void saveBdldoc()}>
                      <Save className="h-3.5 w-3.5" />{busy === 'bdldoc' ? '保存中…' : '保存'}
                    </Button>
                  </div>
                  <div className="mt-2 flex items-center gap-2 text-[11px]">
                    {bd?.dir
                      ? <span className="flex items-center gap-1 text-emerald-600 dark:text-emerald-400"><CheckCircle2 className="h-3.5 w-3.5" />已设置</span>
                      : <span className="flex items-center gap-1 text-amber-600 dark:text-amber-400"><AlertCircle className="h-3.5 w-3.5" />未设置</span>}
                    {bd?.dir && (
                      <span className={cn(bd.exists ? 'text-zinc-500 dark:text-zinc-400' : 'text-red-600 dark:text-red-400')}>
                        {bd.exists ? '目录存在' : '目录不存在(仍可保存,请确认路径)'}
                      </span>
                    )}
                  </div>
                  <p className="mt-1 text-[11px] text-zinc-500 dark:text-zinc-400">
                    仅改设置,不移动文档文件;仓库自带文档在仓库的 <code>docs/bdl</code> 下。
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* 运行信息(版本来自共享的 /api/health,路径与监听地址来自共享的 /api/hosts) */}
          <section>
            <SectionTitle>运行信息</SectionTitle>
            <div className="mt-2 grid grid-cols-1 gap-2 text-xs sm:grid-cols-3">
              <div className="border border-zinc-200 p-3 dark:border-zinc-800">
                <div className="text-[11px] text-zinc-500 dark:text-zinc-400">配置文件</div>
                <div className="mt-0.5 break-all font-mono text-[11px]" title={configPath}>{configPath || '—'}</div>
              </div>
              <div className="border border-zinc-200 p-3 dark:border-zinc-800">
                <div className="text-[11px] text-zinc-500 dark:text-zinc-400">服务地址</div>
                <div className="mt-0.5 break-all font-mono text-[11px]">{listen ? `http://${listen}` : '—'}</div>
              </div>
              <div className="border border-zinc-200 p-3 dark:border-zinc-800">
                <div className="text-[11px] text-zinc-500 dark:text-zinc-400">版本</div>
                <div className="mt-0.5 break-all font-mono text-[11px]">{version || '—'}</div>
              </div>
            </div>
          </section>
        </div>
      </main>
    </div>
  )
}
