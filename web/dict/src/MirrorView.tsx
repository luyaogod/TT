// 源码镜像页:在浏览器里操作 tdict mirror 的 dir / pull,并实时显示拉取进度。
//
// 后端 /dict/api/mirror 返回镜像根、各环境镜像现状与当前拉取任务;拉取在服务端后台执行
// (host.MirrorPullProgress),前端轮询该接口画进度条——通常不通过 CLI 操作本功能。
// 前缀 /dict 是挂载点带来的(统一服务把字典子系统挂在 /dict 下);见 api.ts 的 API_BASE。
import { useCallback, useEffect, useRef, useState } from 'react'
import { RefreshCw, Download, FolderOpen, Save, AlertCircle, CheckCircle2 } from 'lucide-react'
import { api, type MirrorResp } from './api'
import { Button, cn, Field, Input, SectionTitle } from './ui'

const PHASE_LABEL: Record<string, string> = {
  connect: '连接服务器…',
  probe: '探测 T100 目录…',
  pack: '服务器打包…',
  download: '下载并解压…',
  done: '完成',
  error: '失败',
}

function fmtBytes(n: number): string {
  if (!n) return '0 B'
  if (n >= 1 << 30) return (n / (1 << 30)).toFixed(2) + ' GB'
  if (n >= 1 << 20) return (n / (1 << 20)).toFixed(1) + ' MB'
  if (n >= 1 << 10) return (n / (1 << 10)).toFixed(1) + ' KB'
  return n + ' B'
}

export function MirrorView() {
  const [data, setData] = useState<MirrorResp | null>(null)
  const [dirInput, setDirInput] = useState('')
  const [env, setEnv] = useState('')
  const [err, setErr] = useState('')
  const [notice, setNotice] = useState('')
  const [busy, setBusy] = useState('')
  const dirInit = useRef(false)

  const refresh = useCallback(async () => {
    try {
      const d = await api.mirror()
      setData(d)
      setErr('')
      if (!dirInit.current) {
        setDirInput(d.mirrorDir)
        dirInit.current = true
      }
      setEnv((prev) => {
        if (prev && d.envs.some((e) => e.name === prev)) return prev
        if (d.activeEnv && d.envs.some((e) => e.name === d.activeEnv)) return d.activeEnv
        return d.envs[0]?.name || ''
      })
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    }
  }, [])

  useEffect(() => { void refresh() }, [refresh])

  const running = !!data?.job?.running
  // 拉取进行中每 800ms 刷新进度;结束后停止轮询
  useEffect(() => {
    if (!running) return
    const t = window.setInterval(() => { void refresh() }, 800)
    return () => window.clearInterval(t)
  }, [running, refresh])

  const saveDir = async () => {
    if (!dirInput.trim()) { setErr('请填写镜像根目录'); return }
    setBusy('dir'); setErr(''); setNotice('')
    try {
      const r = await api.saveMirrorDir(dirInput.trim())
      setDirInput(r.mirrorDir || dirInput.trim())
      setNotice('镜像根目录已保存:' + (r.mirrorDir || dirInput.trim()))
      await refresh()
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally { setBusy('') }
  }

  const pull = async (full: boolean) => {
    if (!env) { setErr('请先选择环境'); return }
    if (full && !window.confirm(`全量重建环境「${env}」?\n将整目录替换本地镜像(删除服务器已不存在的残留),首次或按需执行,耗时较长。`)) return
    setErr(''); setNotice('')
    try {
      await api.mirrorPull(env, full)
      await refresh()
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    }
  }

  const job = data?.job
  const indeterminate = !!job?.running && job.total === 0
  const pct = !job ? 0
    : job.total > 0 ? Math.min(100, Math.round((job.bytes * 100) / job.total))
      : job.phase === 'done' ? 100 : 0
  const cur = data?.envs.find((e) => e.name === env)

  return (
    <div className="flex h-full min-h-0 flex-col bg-zinc-50 text-zinc-800 dark:bg-zinc-950 dark:text-zinc-100">
      <header className="flex shrink-0 items-center gap-3 border-b border-zinc-200 bg-white px-4 py-2 dark:border-zinc-800 dark:bg-zinc-900">
        <h1 className="shrink-0 text-sm font-semibold">源码镜像</h1>
        <span className="min-w-0 flex-1 truncate text-[11px] text-zinc-500 dark:text-zinc-400">
          把某环境 T100 服务器的 4gl/4fd 源码、42s(zh_CN)字符串与 *.inc 包含文件镜像到本地,供 AI 用本地文件工具读码
        </span>
        {err && <span className="flex shrink-0 items-center gap-1 text-[11px] text-red-600 dark:text-red-400"><AlertCircle className="h-3.5 w-3.5" />{err}</span>}
      </header>

      <main className="min-h-0 flex-1 overflow-auto p-4">
        <div className="mx-auto max-w-3xl space-y-4">
          {/* 镜像根目录 */}
          <section>
            <SectionTitle>镜像根目录</SectionTitle>
            <div className="mt-2 flex items-end gap-2">
              <Field label="本地保存镜像的根目录(必须显式设置)" className="min-w-0 flex-1">
                <Input value={dirInput} placeholder="如 D:\dev\erp-src" disabled={running}
                  onChange={(e) => setDirInput(e.target.value)}
                  onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); void saveDir() } }} />
              </Field>
              <Button variant="outline" disabled={running || busy === 'dir'} onClick={() => void saveDir()}>
                <Save className="h-3.5 w-3.5" />{busy === 'dir' ? '保存中…' : '保存'}
              </Button>
            </div>
            {data && !data.mirrorDir && (
              <p className="mt-1 text-[11px] text-amber-600 dark:text-amber-400">尚未设置镜像根目录:设置后才能拉取。</p>
            )}
            {notice && <p className="mt-1 text-[11px] text-emerald-600 dark:text-emerald-400">{notice}</p>}
          </section>

          {/* 环境与操作 */}
          <section>
            <SectionTitle>环境</SectionTitle>
            {!data && <div className="mt-2 text-xs text-zinc-500">加载中…</div>}
            {data && data.envs.length === 0 && (
              <div className="mt-2 border border-dashed border-zinc-300 p-3 text-xs text-zinc-500 dark:border-zinc-700">
                还没有 SSH 环境。先到「环境配置」添加服务器与数据库。
              </div>
            )}
            {data && data.envs.length > 0 && (
              <>
                <div className="mt-2 flex items-center gap-2">
                  <span className="shrink-0 text-[11px] text-zinc-500 dark:text-zinc-400">环境</span>
                  <select value={env} disabled={running} onChange={(e) => setEnv(e.target.value)}
                    className="h-7 min-w-0 flex-1 border border-zinc-300 bg-white px-2 text-xs text-zinc-800 outline-none focus:border-sky-500 disabled:opacity-60 dark:border-zinc-700 dark:bg-zinc-950 dark:text-zinc-100">
                    {data.envs.map((e) => (
                      <option key={e.name} value={e.name}>
                        {e.name}{e.name === data.activeEnv ? '（默认）' : ''}{e.ready ? ' · 已有镜像' : ' · 未拉取'}
                      </option>
                    ))}
                  </select>
                </div>
                {cur?.path && (
                  <div className="mt-1 flex min-w-0 items-center gap-1 text-[11px] text-zinc-500 dark:text-zinc-400">
                    <FolderOpen className="h-3.5 w-3.5 shrink-0" /><span className="truncate" title={cur.path}>{cur.path}</span>
                  </div>
                )}
              </>
            )}
            <div className="mt-2 flex items-center gap-2">
              <Button variant="primary" disabled={running || !env || !data?.mirrorDir} onClick={() => void pull(false)}>
                <RefreshCw className={cn('h-3.5 w-3.5', running && 'animate-spin')} />增量更新
              </Button>
              <Button variant="outline" disabled={running || !env || !data?.mirrorDir} onClick={() => void pull(true)}>
                <Download className="h-3.5 w-3.5" />全量重建
              </Button>
            </div>
            <p className="mt-1 text-[11px] text-zinc-500 dark:text-zinc-400">
              增量:服务器 marker 记录基线,只拉变更文件(本地无完整镜像时自动转全量);全量:整目录替换。
            </p>
          </section>

          {/* 进度 */}
          {job && (job.running || job.done) && (
            <section>
              <SectionTitle>拉取进度</SectionTitle>
              <div className="mt-2 border border-zinc-200 p-3 dark:border-zinc-800">
                <div className="mb-2 flex items-center gap-2 text-xs">
                  <span className="font-medium">{job.env}</span>
                  <span className="text-zinc-500 dark:text-zinc-400">{job.full ? '全量' : '增量'} · {PHASE_LABEL[job.phase] || job.phase}</span>
                  <span className="ml-auto text-[11px] text-zinc-500 dark:text-zinc-400">
                    {job.running ? job.elapsed : (job.elapsed ? '用时 ' + job.elapsed : '')}
                  </span>
                </div>
                <div className="h-2 w-full overflow-hidden bg-zinc-200 dark:bg-zinc-800">
                  {indeterminate
                    ? <div className="bar-indeterminate h-full w-1/3 bg-sky-500" />
                    : <div className={cn('h-full transition-[width] duration-300', job.phase === 'error' ? 'bg-red-500' : 'bg-sky-500')} style={{ width: pct + '%' }} />}
                </div>
                <div className="mt-2 grid grid-cols-3 gap-2 text-[11px] text-zinc-500 dark:text-zinc-400">
                  <span>已传输:{fmtBytes(job.bytes)}{job.total > 0 ? ` / ${fmtBytes(job.total)}` : ''}{job.total > 0 ? ` (${pct}%)` : ''}</span>
                  <span>文件数:{job.files || '—'}</span>
                  <span>{job.running ? '进行中…' : '已结束'}</span>
                </div>
                {job.message && (
                  <p className={cn('mt-2 flex items-center gap-1 text-xs',
                    job.phase === 'error' ? 'text-red-600 dark:text-red-400'
                      : job.phase === 'done' ? 'text-emerald-600 dark:text-emerald-400' : 'text-zinc-600 dark:text-zinc-300')}>
                    {job.phase === 'done' && <CheckCircle2 className="h-3.5 w-3.5" />}
                    {job.phase === 'error' && <AlertCircle className="h-3.5 w-3.5" />}
                    {job.error || job.message}
                  </p>
                )}
              </div>
            </section>
          )}
        </div>
      </main>
    </div>
  )
}
