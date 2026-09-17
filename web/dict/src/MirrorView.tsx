// 源码镜像页:在浏览器里操作 tdict mirror 的 pull,并实时显示拉取进度。
//
// 后端 /dict/api/mirror 返回镜像根、各环境镜像现状与当前拉取任务;拉取在服务端后台执行
// (host.MirrorPullProgress),前端轮询该接口画进度条——通常不通过 CLI 操作本功能。
// 前缀 /dict 是挂载点带来的(统一服务把字典子系统挂在 /dict 下);见 api.ts 的 API_BASE。
//
// 镜像根目录**不在这里改**:它属于配置,已收进统一设置页(/debug/#settings/data-dict)。
// 本页只显示生效值并给一个跳过去的入口 —— 一个页面既配置又跑长任务,职责会糊在一起。
import { useCallback, useEffect, useState } from 'react'
import { RefreshCw, Download, FolderOpen, AlertCircle, CheckCircle2, Settings2 } from 'lucide-react'
import { api, type MirrorResp } from './api'
import { Button } from '../../shared/ui'
import { Card, InfoRow } from '../../shared/settings'
import { cn } from '../../shared/utils'

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

// 环境下拉:原生 <select> 换成语义 token 的样式。
// (统一设置页那边用的是 Radix 下拉;这里保留原生,因为本页只需要一个纯选择器,
//  为它引一套弹层组件不划算。)
const selectCls = 'h-7 min-w-0 flex-1 border border-input bg-transparent px-2 text-xs text-foreground outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:opacity-60'

export function MirrorView() {
  const [data, setData] = useState<MirrorResp | null>(null)
  const [env, setEnv] = useState('')
  const [err, setErr] = useState('')

  const refresh = useCallback(async () => {
    try {
      const d = await api.mirror()
      setData(d)
      setErr('')
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

  const pull = async (full: boolean) => {
    if (!env) { setErr('请先选择环境'); return }
    if (full && !window.confirm(`全量重建环境「${env}」?\n将整目录替换本地镜像(删除服务器已不存在的残留),首次或按需执行,耗时较长。`)) return
    setErr('')
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
    <div className="flex h-full min-h-0 flex-col bg-background text-foreground">
      <header className="flex shrink-0 items-center gap-3 border-b border-border bg-card px-4 py-2">
        <h1 className="shrink-0 text-sm font-semibold">源码镜像</h1>
        <span className="min-w-0 flex-1 truncate text-[11px] text-muted-foreground">
          把某环境 T100 服务器的 4gl/4fd 源码、42s(zh_CN)字符串与 *.inc 包含文件镜像到本地,供 AI 用本地文件工具读码
        </span>
        {err && <span className="flex shrink-0 items-center gap-1 text-[11px] text-red-600 dark:text-red-400"><AlertCircle className="h-3.5 w-3.5" />{err}</span>}
      </header>

      <main className="min-h-0 flex-1 overflow-auto p-4">
        <div className="mx-auto max-w-3xl space-y-3">
          <Card
            title="镜像根目录"
            description="本地保存镜像的根目录(必须显式设置)。属于配置,在统一设置页里改。"
            right={
              <a href="/debug/#settings/data-dict"
                className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
                <Settings2 className="h-3.5 w-3.5" />在设置中修改
              </a>
            }
          >
            <InfoRow label="生效值" value={data?.mirrorDir || '(未设置)'} title={data?.mirrorDir} />
            {data && !data.mirrorDir && (
              <p className="mt-1 border-l-2 border-amber-500/50 bg-amber-500/5 px-2 py-1 text-[11px] text-muted-foreground">
                尚未设置镜像根目录:设置后才能拉取。
                <a href="/debug/#settings/data-dict" className="ml-1 text-primary hover:underline">去设置</a>
              </p>
            )}
          </Card>

          <Card title="环境与操作">
            {!data && <div className="text-xs text-muted-foreground">加载中…</div>}
            {data && data.envs.length === 0 && (
              <div className="border border-dashed border-border p-3 text-xs text-muted-foreground">
                还没有 SSH 环境。先到统一设置页的「站点管理」添加服务器与数据库。
                <a href="/debug/#settings/sites" className="ml-1 text-primary hover:underline">去添加</a>
              </div>
            )}
            {data && data.envs.length > 0 && (
              <>
                <div className="flex items-center gap-2">
                  <span className="shrink-0 text-[11px] text-muted-foreground">环境</span>
                  <select value={env} disabled={running} onChange={(e) => setEnv(e.target.value)} className={selectCls}>
                    {data.envs.map((e) => (
                      <option key={e.name} value={e.name}>
                        {e.name}{e.name === data.activeEnv ? '（默认）' : ''}{e.ready ? ' · 已有镜像' : ' · 未拉取'}
                      </option>
                    ))}
                  </select>
                </div>
                {cur?.path && (
                  <div className="mt-1 flex min-w-0 items-center gap-1 text-[11px] text-muted-foreground">
                    <FolderOpen className="h-3.5 w-3.5 shrink-0" /><span className="truncate" title={cur.path}>{cur.path}</span>
                  </div>
                )}
              </>
            )}
            <div className="mt-3 flex items-center gap-2">
              <Button variant="default" size="sm" disabled={running || !env || !data?.mirrorDir} onClick={() => void pull(false)}>
                <RefreshCw className={cn('h-3.5 w-3.5', running && 'animate-spin')} />增量更新
              </Button>
              <Button variant="outline" size="sm" disabled={running || !env || !data?.mirrorDir} onClick={() => void pull(true)}>
                <Download className="h-3.5 w-3.5" />全量重建
              </Button>
            </div>
            <p className="mt-1 text-[11px] text-muted-foreground">
              增量:服务器 marker 记录基线,只拉变更文件(本地无完整镜像时自动转全量);全量:整目录替换。
            </p>
          </Card>

          {job && (job.running || job.done) && (
            <Card title="拉取进度">
              <div className="mb-2 flex items-center gap-2 text-xs">
                <span className="font-medium">{job.env}</span>
                <span className="text-muted-foreground">{job.full ? '全量' : '增量'} · {PHASE_LABEL[job.phase] || job.phase}</span>
                <span className="ml-auto text-[11px] text-muted-foreground">
                  {job.running ? job.elapsed : (job.elapsed ? '用时 ' + job.elapsed : '')}
                </span>
              </div>
              {/* 轨道与填充必须成对换 token:只改填充会让暗色下的轨道消失 */}
              <div className="h-2 w-full overflow-hidden bg-muted">
                {indeterminate
                  ? <div className="bar-indeterminate h-full w-1/3 bg-primary" />
                  : <div className={cn('h-full transition-[width] duration-300', job.phase === 'error' ? 'bg-destructive' : 'bg-primary')} style={{ width: pct + '%' }} />}
              </div>
              <div className="mt-2 grid grid-cols-3 gap-2 text-[11px] text-muted-foreground">
                <span>已传输:{fmtBytes(job.bytes)}{job.total > 0 ? ` / ${fmtBytes(job.total)}` : ''}{job.total > 0 ? ` (${pct}%)` : ''}</span>
                <span>文件数:{job.files || '—'}</span>
                <span>{job.running ? '进行中…' : '已结束'}</span>
              </div>
              {job.message && (
                <p className={cn('mt-2 flex items-center gap-1 text-xs',
                  job.phase === 'error' ? 'text-red-600 dark:text-red-400'
                    : job.phase === 'done' ? 'text-emerald-600 dark:text-emerald-400' : 'text-foreground')}>
                  {job.phase === 'done' && <CheckCircle2 className="h-3.5 w-3.5" />}
                  {job.phase === 'error' && <AlertCircle className="h-3.5 w-3.5" />}
                  {job.error || job.message}
                </p>
              )}
            </Card>
          )}
        </div>
      </main>
    </div>
  )
}
