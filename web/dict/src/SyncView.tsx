// 数据同步页:在浏览器里选环境执行 tdict db sync(远程 ERP 字典 → 本地 SQLite),并显示进度。
//
// 后端 POST /dict/api/dbsync 在服务端后台执行(dbsync.Run),前端轮询 /dict/api/dbsync 画进度:
// 逐表拉取(第 x/y 张表 + 当前表行数/累计行数)→ 建索引 → 原子替换本地库(原库备份 .bak)。
// 前缀 /dict 是挂载点带来的(统一服务把字典子系统挂在 /dict 下);见 api.ts 的 API_BASE。
//
// 同步目标**不在这里改**:它属于配置,已收进统一设置页(/debug/#settings/data-dict)。
// 本页只显示生效值并给一个跳过去的入口。
import { useCallback, useEffect, useState } from 'react'
import { RefreshCw, AlertCircle, CheckCircle2, Settings2 } from 'lucide-react'
import { api, type DBSyncResp } from './api'
import { Button } from '../../shared/ui'
import { Card, InfoRow } from '../../shared/settings'
import { cn } from '../../shared/utils'

const PHASE_LABEL: Record<string, string> = {
  open: '连接远程数据库…',
  table: '拉取数据表…',
  index: '创建主键索引…',
  replace: '替换本地数据库…',
  done: '完成',
  error: '失败',
}

const selectCls = 'h-7 min-w-0 flex-1 border border-input bg-transparent px-2 text-xs text-foreground outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] disabled:opacity-60'

export function SyncView() {
  const [data, setData] = useState<DBSyncResp | null>(null)
  const [env, setEnv] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState('')

  const refresh = useCallback(async () => {
    try {
      const d = await api.dbsync()
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
  // 同步进行中每 800ms 刷新进度;结束后停止轮询
  useEffect(() => {
    if (!running) return
    const t = window.setInterval(() => { void refresh() }, 800)
    return () => window.clearInterval(t)
  }, [running, refresh])

  const run = async () => {
    if (!env) { setErr('请先选择环境'); return }
    if (!window.confirm(`从环境「${env}」拉取字典数据到:\n${data?.target || ''}\n\n将覆盖本地 SQLite(原库自动备份为 .bak),约 85 万行,需数分钟。继续?`)) return
    setBusy('run'); setErr('')
    try {
      await api.dbsyncRun(env)
      await refresh()
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally { setBusy('') }
  }

  const job = data?.job
  const pct = !job ? 0
    : job.phase === 'done' ? 100
      : job.tableTotal > 0 ? Math.min(100, Math.round((job.tableIndex * 100) / job.tableTotal)) : 0
  const cur = data?.envs.find((e) => e.name === env)

  return (
    <div className="flex h-full min-h-0 flex-col bg-background text-foreground">
      <header className="flex shrink-0 items-center gap-3 border-b border-border bg-card px-4 py-2">
        <h1 className="shrink-0 text-sm font-semibold">数据同步</h1>
        <span className="min-w-0 flex-1 truncate text-[11px] text-muted-foreground" title={data?.target}>
          从远程 ERP 拉取数据字典(表字典/校验/分类码/画面规格/开窗/消息/参数)到本地 SQLite
        </span>
        {err && <span className="flex shrink-0 items-center gap-1 text-[11px] text-red-600 dark:text-red-400"><AlertCircle className="h-3.5 w-3.5" />{err}</span>}
      </header>

      <main className="min-h-0 flex-1 overflow-auto p-4">
        <div className="mx-auto max-w-3xl space-y-3">
          <Card
            title="目标数据库"
            description="同步写入的本地 SQLite。属于配置,在统一设置页里改。"
            right={
              <a href="/debug/#settings/data-dict"
                className="inline-flex items-center gap-1 text-[11px] text-primary hover:underline">
                <Settings2 className="h-3.5 w-3.5" />在设置中修改
              </a>
            }
          >
            <InfoRow label="生效值" value={data?.target || '—'} title={data?.target} />
            <InfoRow label="来源" value={data?.configured ? '自定义(设置页里配的)' : '默认位置'} />
            {data?.configured && <InfoRow label="默认位置" value={data.defaultTarget || '—'} title={data.defaultTarget} />}
            <div className="mt-1 flex items-center gap-2 text-[11px]">
              <span className={cn('flex items-center gap-1', data?.exists ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400')}>
                {data?.exists ? <CheckCircle2 className="h-3.5 w-3.5" /> : <AlertCircle className="h-3.5 w-3.5" />}
                {data?.exists ? '目标文件已存在' : '目标文件尚未创建(同步时自动创建)'}
              </span>
            </div>
            <p className="mt-2 text-[11px] text-muted-foreground">
              查询命令读的也是它;原库自动备份为 <code>.bak</code>,失败不影响原库。
            </p>
          </Card>

          <Card title="环境与操作">
            {!data && <div className="text-xs text-muted-foreground">加载中…</div>}
            {data && data.envs.length === 0 && (
              <div className="border border-dashed border-border p-3 text-xs text-muted-foreground">
                没有可同步的环境(需先在统一设置页的「站点管理」为环境挂上数据库)。
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
                        {e.name}{e.name === data.activeEnv ? '（默认）' : ''}
                      </option>
                    ))}
                  </select>
                </div>
                {cur && (
                  <div className="mt-1 text-[11px] text-muted-foreground">{cur.type} · {cur.address}</div>
                )}
              </>
            )}
            <div className="mt-3">
              <Button variant="default" size="sm" disabled={running || busy !== '' || !env || !data?.envs.length} onClick={() => void run()}>
                <RefreshCw className={cn('h-3.5 w-3.5', running && 'animate-spin')} />{running ? '同步中…' : '开始同步'}
              </Button>
            </div>
          </Card>

          {job && (job.running || job.done) && (
            <Card title="同步进度">
              <div className="mb-2 flex items-center gap-2 text-xs">
                <span className="font-medium">{job.env}</span>
                <span className="text-muted-foreground">{PHASE_LABEL[job.phase] || job.phase}</span>
                <span className="ml-auto text-[11px] text-muted-foreground">
                  {job.running ? job.elapsed : (job.elapsed ? '用时 ' + job.elapsed : '')}
                </span>
              </div>
              {/* 轨道与填充必须成对换 token:只改填充会让暗色下的轨道消失 */}
              <div className="h-2 w-full overflow-hidden bg-muted">
                <div className={cn('h-full transition-[width] duration-300', job.phase === 'error' ? 'bg-destructive' : 'bg-primary')}
                  style={{ width: pct + '%' }} />
              </div>
              <div className="mt-2 grid grid-cols-3 gap-2 text-[11px] text-muted-foreground">
                <span>数据表:{job.tableTotal > 0 ? `${job.tableIndex}/${job.tableTotal}` : '—'}{job.tables > 0 ? `(完成 ${job.tables})` : ''}</span>
                <span>当前表:{(job.table || '—') + (job.tableRows > 0 ? ` ${job.tableRows} 行` : '')}</span>
                <span>累计行数:{job.totalRows > 0 ? job.totalRows.toLocaleString() : '—'}</span>
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
              {job.backup && <p className="mt-1 text-[11px] text-muted-foreground">原库已备份:{job.backup}</p>}
              {job.warning && <pre className="mt-1 whitespace-pre-wrap text-[11px] text-amber-600 dark:text-amber-400">{job.warning}</pre>}
            </Card>
          )}
        </div>
      </main>
    </div>
  )
}
