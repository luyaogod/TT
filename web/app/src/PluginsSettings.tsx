// 设置页「插件管理」：一张卡只回答三件事 —— 叫什么、做什么、开没开。
//
// 刻意**不加载插件页面**：第三方页面只在 iframe 里由外壳的 #plugin/<route> 承载
// （见 README 的「插件页面与消息桥」）。这里是管理视图，只做元数据与生命周期动作。
//
// 每个动作都打 Core 管理面的既有端点（/api/plugins*），不新开一套契约：
// 列表 = GET /api/plugins，启用 = POST …/enable（能力+环境授权），停用 = …/disable。
//
// 卡片只留名称/描述/开关：版本、运行态、pid/端口、贡献面、声明 vs 已授权、自检、日志
// 都不再铺在卡上（诊断有 `tt plugin doctor` / `tt plugin logs`，不必占设置页的版面）。
// 声明了权限的插件，打开开关时仍会弹一次授权面板 —— 那是启用的前置条件，不是元数据。
import { useCallback, useEffect, useState } from 'react'
import { AlertCircle, RefreshCw } from 'lucide-react'
import {
  api, type PluginAdminSummary,
} from './api'
import { Button, Switch } from '../../shared/ui'
import { Card, SectionTitle } from '../../shared/settings'
import { Checkbox, Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '../../shared/ui-radix'
import { useToast } from './toast'
import { cn } from '../../shared/utils'
import { capabilityLabel, grantsPayload, needsEnvSelection } from './pluginAdmin'

/** 一页的插件管理。 */
export function PluginsSettings() {
  const toast = useToast()
  const [list, setList] = useState<PluginAdminSummary[] | null>(null)
  const [envs, setEnvs] = useState<string[]>([])
  const [loadErr, setLoadErr] = useState('')
  const [loading, setLoading] = useState(false)

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
    <div className="space-y-3">
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
        <PluginCard key={p.id} p={p} envs={envs} onChanged={load} />
      ))}
    </div>
  )
}

/** 一个插件一行：名称 + 描述 + 开关。 */
function PluginCard({ p, envs, onChanged }: {
  p: PluginAdminSummary
  envs: string[]
  onChanged: () => Promise<void>
}) {
  const toast = useToast()
  const [busy, setBusy] = useState('')
  const [grantOpen, setGrantOpen] = useState(false)
  // 默认**一条能力都不勾**：声明只是诉求，授权必须是人这一次给出的（见权限矩阵）。
  const [caps, setCaps] = useState<string[]>([])
  const [pickedEnvs, setPickedEnvs] = useState<string[]>([])

  const declared = p.permissions || []
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

  const toggleIn = (arr: string[], v: string, on: boolean): string[] =>
    on ? (arr.includes(v) ? arr : [...arr, v]) : arr.filter((x) => x !== v)

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

  const onSwitch = (next: boolean) => {
    if (busy !== '') return
    if (!next) {
      void disable()
      return
    }
    // 没声明权限的直接启用；声明了就先让人选这次要授予哪些（不勾 = 不授权）。
    if (declared.length === 0) {
      void enable()
    } else {
      setCaps([])
      setPickedEnvs([])
      setGrantOpen(true)
    }
  }

  return (
    <section className="flex items-start justify-between gap-4 rounded-ui border border-border bg-card px-3 py-2.5">
      <div className="min-w-0">
        <div className="truncate text-xs font-medium text-foreground">{p.name || p.id}</div>
        {p.description && (
          <p className="mt-0.5 line-clamp-2 text-[11px] leading-4 text-muted-foreground" title={p.description}>
            {p.description}
          </p>
        )}
      </div>
      <Switch
        className="mt-0.5"
        checked={p.enabled}
        disabled={busy !== ''}
        onCheckedChange={onSwitch}
        aria-label={`${p.enabled ? '停用' : '启用'} ${p.name || p.id}`}
        title={p.lastError || undefined}
      />

      {/* 声明了权限才会打开：授权是启用的前置条件，不是常驻信息。 */}
      <Dialog open={grantOpen} onOpenChange={(o) => {
        setGrantOpen(o)
        if (!o) { setCaps([]); setPickedEnvs([]) }
      }}>
        <DialogContent aria-describedby={undefined} className="max-w-md">
          <DialogHeader>
            <DialogTitle>启用 {p.name || p.id}</DialogTitle>
          </DialogHeader>

          <div className="space-y-3">
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
                          const next = toggleIn(caps, c, v === true)
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
                        onCheckedChange={(v) => setPickedEnvs((cur) => toggleIn(cur, e, v === true))} />
                      <span className="text-[11px]">{e}</span>
                    </label>
                  ))}
                </div>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button variant="ghost" size="sm" onClick={() => setGrantOpen(false)}>取消</Button>
            <Button size="sm" disabled={busy !== ''} onClick={() => void enable()}>
              {busy === 'enable' ? '启用中…' : '确认启用'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  )
}
