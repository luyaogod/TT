// 设置页「插件管理」：已安装插件的元数据、声明权限 vs 已授权、启停、自检与日志。
//
// 刻意**不加载插件页面**：第三方页面只在 iframe 里由外壳的 #plugin/<route> 承载
// （见 README 的「插件页面与消息桥」）。这里是管理视图，只做元数据与生命周期动作。
//
// 每个动作都打 Core 管理面的既有端点（/api/plugins*），不新开一套契约：
// 列表 = GET /api/plugins，启用 = POST …/enable（能力+环境授权），停用 = …/disable，
// 自检 = …/doctor，日志 = …/logs?tail=。
import { useCallback, useEffect, useState } from 'react'
import {
  AlertCircle, CheckCircle2, FileText, Power, PowerOff, RefreshCw, Stethoscope,
} from 'lucide-react'
import {
  api, type PluginAdminSummary, type PluginDoctorIssue, type PluginLogsReply,
} from './api'
import { Badge, Button } from '../../shared/ui'
import { Card, SectionTitle } from '../../shared/settings'
import { Checkbox, Dialog, DialogContent, DialogHeader, DialogTitle } from '../../shared/ui-radix'
import { useToast } from './toast'
import { cn } from '../../shared/utils'
import {
  capabilityLabel, contributionLabels, doctorSummary, grantsPayload, logsSummary,
  needsEnvSelection, pluginStatus, sortIssues, ungrantedPermissions,
} from './pluginAdmin'

/** 运行态的中文名（后端只给 running / stopped / degraded，未知值原样显示）。 */
const RUNTIME_LABEL: Record<string, string> = {
  running: '运行中', stopped: '已停止', degraded: '降级',
}

/** 一页的插件管理。 */
export function PluginsSettings() {
  const toast = useToast()
  const [list, setList] = useState<PluginAdminSummary[] | null>(null)
  const [envs, setEnvs] = useState<string[]>([])
  const [loadErr, setLoadErr] = useState('')
  const [loading, setLoading] = useState(false)
  const [logsFor, setLogsFor] = useState<PluginAdminSummary | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setLoadErr('')
      const r = await api.pluginsAdmin()
      setList(r.plugins || [])
    } catch (e: any) {
      setLoadErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])
  useEffect(() => {
    // 环境清单只是启用面板的下拉数据源；读不到不影响列表本身（授权会显示"未知环境"）。
    void api.settings()
      .then((c) => setEnvs((c.sshs || []).map((e) => e.name).filter((n) => !!n)))
      .catch(() => { /* 提示性数据，静默 */ })
  }, [])

  if (loadErr) {
    return (
      <div className="space-y-2 rounded-ui border border-border p-3">
        <div className="flex items-start gap-2 text-[11px] text-red-600 dark:text-red-400">
          <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span className="break-all">{loadErr}</span>
        </div>
        <Button size="sm" variant="outline" onClick={() => void load()}>
          <RefreshCw className="h-3.5 w-3.5" />重试
        </Button>
      </div>
    )
  }

  if (list === null) {
    return <div className="py-2 text-muted-foreground">读取插件清单…</div>
  }

  const enabled = list.filter((p) => p.enabled).length
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between gap-2">
        <span className="text-[11px] text-muted-foreground">
          已安装 {list.length} 个 · 已启用 {enabled} 个
        </span>
        <Button size="sm" variant="outline" disabled={loading} onClick={() => void load()}>
          <RefreshCw className={cn('h-3.5 w-3.5', loading && 'animate-spin')} />刷新
        </Button>
      </div>

      {list.length === 0 && (
        <Card title="还没有安装插件" description="命令行装好之后回到这里启用并授权。">
          <p className="text-[11px] text-muted-foreground">
            <code className="rounded-ui bg-muted px-1">tt plugin install &lt;插件目录或 zip&gt;</code>
            {' '}装完后列表会刷新出这个插件；安装本身不会执行插件代码。
          </p>
        </Card>
      )}

      {list.map((p) => (
        <PluginCard key={p.id} p={p} envs={envs} onChanged={load} onShowLogs={() => setLogsFor(p)} />
      ))}

      {logsFor && (
        <PluginLogsDialog plugin={logsFor} onClose={() => setLogsFor(null)} />
      )}
    </div>
  )
}

/** 一个插件一张卡：元数据 + 动作 + 可展开的启用面板 / 自检结论。 */
function PluginCard({ p, envs, onChanged, onShowLogs }: {
  p: PluginAdminSummary
  envs: string[]
  onChanged: () => Promise<void>
  onShowLogs: () => void
}) {
  const toast = useToast()
  const [busy, setBusy] = useState('')
  const [grantOpen, setGrantOpen] = useState(false)
  // 默认**一条能力都不勾**：声明只是诉求，授权必须是人这一次给出的（见权限矩阵）。
  const [caps, setCaps] = useState<string[]>([])
  const [pickedEnvs, setPickedEnvs] = useState<string[]>([])
  const [issues, setIssues] = useState<PluginDoctorIssue[] | null>(null)

  const status = pluginStatus(p)
  const declared = p.permissions || []
  const grantedCaps = p.capabilities || []
  const grantedEnvs = p.envs || []
  const canPickEnv = needsEnvSelection(caps)

  const fail = (title: string) => (e: any) => toast.error(title, { description: e?.message || String(e), duration: 0 })

  const run = async (key: string, title: string, fn: () => Promise<void>) => {
    setBusy(key)
    try {
      await fn()
    } catch (e: any) {
      fail(title)(e)
    } finally {
      setBusy('')
    }
  }

  const toggle = (list: string[], v: string, on: boolean): string[] =>
    on ? (list.includes(v) ? list : [...list, v]) : list.filter((x) => x !== v)

  const enable = () => run('enable', '启用失败', async () => {
    await api.pluginEnable(p.id, grantsPayload(caps, pickedEnvs))
    setGrantOpen(false)
    setCaps([])
    setPickedEnvs([])
    toast.success(`已启用 ${p.name || p.id}`)
    await onChanged()
  })

  const disable = () => run('disable', '停用失败', async () => {
    await api.pluginDisable(p.id)
    toast.success(`已停用 ${p.name || p.id}`)
    await onChanged()
  })

  const doctor = () => run('doctor', '自检失败', async () => {
    const r = await api.pluginDoctor(p.id)
    setIssues(sortIssues(r.issues || []))
  })

  return (
    <Card
      title={<span className="flex min-w-0 items-center gap-2">
        <span className="truncate">{p.name || p.id}</span>
        <Badge tone={status.tone}>{status.label}</Badge>
      </span>}
      description={<span className="font-mono">{p.id}</span>}
      right={<>
        <Button variant="ghost" size="sm" disabled={busy !== ''} onClick={() => void doctor()}>
          <Stethoscope className="h-3.5 w-3.5" />{busy === 'doctor' ? '自检中…' : '自检'}
        </Button>
        <Button variant="ghost" size="sm" onClick={onShowLogs}>
          <FileText className="h-3.5 w-3.5" />日志
        </Button>
        {p.enabled ? (
          <Button variant="outline" size="sm" disabled={busy !== ''} onClick={() => void disable()}>
            <PowerOff className="h-3.5 w-3.5" />{busy === 'disable' ? '停用中…' : '停用'}
          </Button>
        ) : (
          <Button size="sm" disabled={busy !== ''} onClick={() => setGrantOpen((v) => !v)}>
            <Power className="h-3.5 w-3.5" />启用
          </Button>
        )}
      </>}
    >
      {/* 元数据行：版本 / 上一版 / 运行态。 */}
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
        <span className="font-mono text-foreground">v{p.version}</span>
        {p.previousVersion && <span>上一版 {p.previousVersion}（可回滚）</span>}
        {p.state && <span>{RUNTIME_LABEL[p.state] || p.state}</span>}
        {p.pid ? <span className="font-mono">pid {p.pid}</span> : null}
        {p.addr && <span className="font-mono">{p.addr}</span>}
        {p.restarts ? <span>重启 {p.restarts} 次</span> : null}
      </div>

      {/* 贡献面：插件到底往宿主里加了什么。 */}
      <div className="flex flex-wrap gap-1">
        {contributionLabels(p).map((t) => (
          <span key={t} className="rounded-ui bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">{t}</span>
        ))}
      </div>

      {/* 声明 ≠ 授权：左右两栏并排，用户一眼能看出还差哪一条。 */}
      <div className="grid gap-1 text-[11px] sm:grid-cols-2">
        <div className="min-w-0">
          <span className="text-muted-foreground">声明权限：</span>
          {declared.length === 0
            ? <span className="text-muted-foreground">无</span>
            : declared.map((c) => (
              <span key={c} title={c} className="mr-1.5">{capabilityLabel(c)}</span>
            ))}
        </div>
        <div className="flex min-w-0 flex-wrap items-center gap-1">
          <span className="text-muted-foreground">已授权：</span>
          {grantedCaps.length === 0
            ? <span className="text-amber-600 dark:text-amber-400">无能力</span>
            : grantedCaps.map((c) => <Badge key={c} tone="gray" title={c}>{capabilityLabel(c)}</Badge>)}
          {grantedEnvs.length > 0 && <span className="ml-1 text-muted-foreground">环境</span>}
          {grantedEnvs.map((e) => <Badge key={e} tone="blue">{e}</Badge>)}
        </div>
      </div>

      {p.lastError && (
        <div className="flex items-start gap-1.5 rounded-ui border border-red-500/40 p-1.5 text-[11px] text-red-600 dark:text-red-400">
          <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span className="break-all">{p.lastError}</span>
        </div>
      )}

      {grantOpen && !p.enabled && (
        <div className="space-y-2 rounded-ui border border-border p-2">
          <div>
            <SectionTitle>能力</SectionTitle>
            {declared.length === 0 ? (
              <p className="text-[11px] text-muted-foreground">该插件没有声明任何权限，可以直接启用。</p>
            ) : (
              <div className="space-y-1">
                {declared.map((c) => (
                  <label key={c} className="flex cursor-pointer items-start gap-2 py-0.5">
                    <Checkbox className="mt-0.5" checked={caps.includes(c)}
                      onCheckedChange={(v) => {
                        const next = toggle(caps, c, v === true)
                        setCaps(next)
                        // 能力全部取消时，环境选择也不再有意义（后端会拒"只给环境"）。
                        if (!needsEnvSelection(next)) setPickedEnvs([])
                      }} />
                    <span className="min-w-0">
                      <span className="text-[11px] text-foreground">{capabilityLabel(c)}</span>
                      <span className="ml-1.5 font-mono text-[10px] text-muted-foreground">{c}</span>
                    </span>
                  </label>
                ))}
              </div>
            )}
          </div>

          <div>
            <SectionTitle>环境</SectionTitle>
            {envs.length === 0 ? (
              <p className="text-[11px] text-muted-foreground">配置里还没有环境。</p>
            ) : (
              <div className={cn('flex flex-wrap gap-x-3 gap-y-1', !canPickEnv && 'opacity-50')}>
                {envs.map((e) => (
                  <label key={e} className={cn('flex items-center gap-1.5 py-0.5', canPickEnv ? 'cursor-pointer' : 'cursor-not-allowed')}
                    title={canPickEnv ? '' : '先勾选至少一个能力；环境授权不会单独生效'}>
                    <Checkbox disabled={!canPickEnv} checked={pickedEnvs.includes(e)}
                      onCheckedChange={(v) => setPickedEnvs((cur) => toggle(cur, e, v === true))} />
                    <span className="text-[11px]">{e}</span>
                  </label>
                ))}
              </div>
            )}
          </div>

          <div className="flex items-center justify-end gap-2">
            <Button variant="ghost" size="sm" onClick={() => setGrantOpen(false)}>取消</Button>
            <Button size="sm" disabled={busy !== ''} onClick={() => void enable()}>
              {busy === 'enable' ? '启用中…' : '确认启用'}
            </Button>
          </div>
        </div>
      )}

      {issues !== null && (
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <SectionTitle>自检</SectionTitle>
            <span className={cn('text-[11px]', issues.length === 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400')}>
              {doctorSummary(issues.length)}
            </span>
            <button type="button" className="mb-1.5 ml-auto text-[11px] text-muted-foreground hover:text-foreground"
              onClick={() => setIssues(null)}>收起</button>
          </div>
          {issues.length === 0 ? (
            <p className="flex items-center gap-1 text-[11px] text-emerald-600 dark:text-emerald-400">
              <CheckCircle2 className="h-3.5 w-3.5" />清单、兼容性与授权都没发现问题。
            </p>
          ) : issues.map((it, i) => (
            <div key={`${it.code}-${i}`} className={cn('flex items-start gap-1.5 text-[11px]',
              it.severity === 'error' ? 'text-red-600 dark:text-red-400' : 'text-amber-600 dark:text-amber-400')}>
              <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
              <div className="min-w-0">
                <span className="font-mono text-[10px]">{it.code}</span>
                <span className="ml-1.5 break-all text-foreground">{it.message}</span>
                {it.hint && <div className="text-muted-foreground">{it.hint}</div>}
              </div>
            </div>
          ))}
        </div>
      )}
    </Card>
  )
}

/** 日志查看：读 logs/<id>/ 里最新那个文件，只读。 */
function PluginLogsDialog({ plugin, onClose }: { plugin: PluginAdminSummary; onClose: () => void }) {
  const [data, setData] = useState<PluginLogsReply | null>(null)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)

  const load = useCallback(async () => {
    setBusy(true)
    setErr('')
    try {
      setData(await api.pluginLogs(plugin.id, 500))
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setBusy(false)
    }
  }, [plugin.id])

  useEffect(() => { void load() }, [load])

  return (
    <Dialog open onOpenChange={(o) => { if (!o) onClose() }}>
      <DialogContent aria-describedby={undefined} className="max-w-4xl">
        <DialogHeader>
          <DialogTitle>日志 · {plugin.name || plugin.id}</DialogTitle>
        </DialogHeader>
        <div className="flex items-center justify-between gap-2 text-[11px] text-muted-foreground">
          <span className="min-w-0 truncate">
            {data ? logsSummary(data) : '读取中…'}
            {data?.file ? <span className="ml-2 font-mono text-[10px]">{data.file}</span> : null}
          </span>
          <Button variant="ghost" size="sm" disabled={busy} onClick={() => void load()}>
            <RefreshCw className={cn('h-3.5 w-3.5', busy && 'animate-spin')} />刷新
          </Button>
        </div>
        {err && (
          <p className="flex items-start gap-1.5 text-[11px] text-red-600 dark:text-red-400">
            <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />{err}
          </p>
        )}
        <pre className="min-h-24 flex-1 overflow-auto rounded-ui bg-muted/40 p-2 font-mono text-[11px] leading-4 whitespace-pre-wrap break-all">
          {(data?.lines || []).join('\n') || (busy ? '读取中…' : '（没有日志）')}
        </pre>
      </DialogContent>
    </Dialog>
  )
}
