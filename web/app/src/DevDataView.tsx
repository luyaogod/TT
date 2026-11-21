// 「开发资料」分区：按环境准备 model 需要的原材料。
//
// 它与「数据字典」分区不是同一件事，所以页面也分开：
//   数据字典 → 一份可供查询的本地库（读侧）
//   开发资料 → 按程序成套、带指纹的原材料（写侧/model 侧）
// 用户看到的是「共用资料（一次拉全环境的规格表与模板）」和「按需资料（一个程序一套）」
// 两件事，各自带进度与停止。
//
// 轮询策略：任务在跑时 1 秒一次 status，结束后停手；离开分区（组件卸载）也停手 ——
// **不取消任务**（拉一半就走的场景很常见，比如用户去看别的分区）。

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Download, RefreshCw, Trash2, Square, FolderCog } from 'lucide-react'
import { Button, Input, Badge, Separator } from '../../shared/ui'
import { Card } from '../../shared/settings'
import { Progress, Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
  Checkbox, AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '../../shared/ui-radix'
import { api, type DevDataStatus, type DevDataJob, type DevDataProg } from './api'
import { fmtBytes } from './progress'
import { useStore } from './store'
import { useToast } from './toast'
import { cn } from '../../shared/utils'

const POLL_MS = 1000

export function DevDataView() {
  const envs = useStore((s) => s.envs)
  const envsLoaded = useStore((s) => s.envsLoaded)
  const loadEnvs = useStore((s) => s.loadEnvs)
  const toast = useToast()

  // 当前环境：默认第一个（与设置页其它分区一致，不写全局 store）。
  const [env, setEnv] = useState('')
  const [st, setSt] = useState<DevDataStatus | null>(null)
  const [err, setErr] = useState('')
  // 拉取表单
  const [module, setModule] = useState('')
  const [prog, setProg] = useState('')
  const [withDesign, setWithDesign] = useState(false)
  const [form, setForm] = useState(false)
  // 删除确认
  const [pendingDel, setPendingDel] = useState<DevDataProg | null>(null)
  // 正在核对服务器（check）的程序
  const [checking, setChecking] = useState('')

  const alive = useRef(true)
  useEffect(() => {
    alive.current = true
    return () => { alive.current = false }
  }, [])

  useEffect(() => {
    if (!envsLoaded) void loadEnvs()
  }, [envsLoaded, loadEnvs])

  useEffect(() => {
    if (!env && envs.length > 0) setEnv(envs[0].name)
  }, [envs, env])

  const refresh = useCallback(async (name: string, quiet = false) => {
    if (!name) return
    try {
      const s = await api.devdataStatus(name)
      if (!alive.current) return
      setSt(s)
      if (!quiet) setErr('')
    } catch (e: any) {
      if (!alive.current) return
      // 读不到状态时不要清空旧数据（那会让界面闪成"未同步"，与"读不到"混淆）
      setErr(e?.message || String(e))
    }
  }, [])

  useEffect(() => { void refresh(env) }, [env, refresh])

  const running = useMemo(() => (st?.jobs ?? []).filter((j) => j.running), [st])
  const runningMirror = running.find((j) => j.kind === 'mirror')
  const runningFetch = running.find((j) => j.kind === 'fetch')

  // 任务在跑时轮询；结束即停（不空转）
  useEffect(() => {
    if (running.length === 0 || !env) return
    const t = window.setInterval(() => { void refresh(env, true) }, POLL_MS)
    return () => window.clearInterval(t)
  }, [running.length, env, refresh])

  // 任务结束时给一句结果（成功/取消/失败三种语气不同）。
  //
  // status 里会带**已结束**的任务（后端 JobsFor），所以不能"看到结束态就提示"——
  // 否则一进设置页就会为上一次跑完的任务弹提示。只在**见过它跑**之后收尾时提示。
  const seenRunning = useRef<Set<string>>(new Set())
  useEffect(() => {
    for (const j of st?.jobs ?? []) {
      const key = `${j.kind}:${j.prog ?? ''}`
      if (j.running) {
        seenRunning.current.add(key)
        continue
      }
      if (!seenRunning.current.has(key)) continue
      seenRunning.current.delete(key)
      if (j.phase === 'done') toast.success(j.message || '完成')
      else if (j.phase === 'cancelled') toast.info('已停止')
      else if (j.phase === 'error') toast.error(j.error || '失败')
    }
  }, [st, toast])

  const startMirror = async (only?: 'dict' | 'mdl') => {
    if (!env) return
    try {
      const r = await api.devdataMirror(env, { only })
      toast.info(only === 'dict' ? '正在同步规格表…' : only === 'mdl' ? '正在同步模板…' : '正在同步共用资料…')
      setSt((s) => (s ? { ...s, jobs: [...(s.jobs ?? []), r.job] } : s))
      void refresh(env, true)
    } catch (e: any) {
      toast.error(e?.message || String(e))
    }
  }

  const stopJob = async (kind: 'mirror' | 'fetch', p?: string) => {
    try {
      await api.devdataStop(env, kind, p)
      toast.info('正在停止…')
    } catch (e: any) {
      toast.error(e?.message || String(e))
    }
  }

  const startFetch = async (p: string, force = false) => {
    if (!env || !p.trim()) return
    try {
      const r = await api.devdataFetch(env, p.trim(), {
        module: module.trim() || undefined, withDesign, form, force,
      })
      toast.info(force ? '正在重新拉取…' : '正在拉取…')
      setSt((s) => (s ? { ...s, jobs: [...(s.jobs ?? []), r.job] } : s))
      void refresh(env, true)
    } catch (e: any) {
      toast.error(e?.message || String(e))
    }
  }

  const checkProg = async (p: string) => {
    setChecking(p)
    try {
      const r = await api.devdataCheck(env, p)
      const reasons = r.prog.staleReasons ?? []
      if (r.prog.stale) toast.info(`已过期：${reasons.map((x) => x.detail).join('；')}`)
      else toast.success('材料是最新的')
      void refresh(env, true)
    } catch (e: any) {
      toast.error(e?.message || String(e))
    } finally {
      setChecking('')
    }
  }

  const removeProg = async () => {
    const p = pendingDel
    setPendingDel(null)
    if (!p) return
    try {
      await api.devdataRemoveProg(env, p.prog)
      toast.success(`已删除 ${p.prog}`)
      void refresh(env, true)
    } catch (e: any) {
      toast.error(e?.message || String(e))
    }
  }

  const dict = st?.dict
  const mdl = st?.mdl
  const tableLabel = dict && dict.tables > 0
    ? `${dict.tables}${dict.skipped > 0 ? `（跳过 ${dict.skipped}）` : ''}`
    : '—'
  const totalBytes = (dict?.size ?? 0) + (mdl?.bytes ?? 0)

  return (
    <>
      <Card id="card-devdata-env" title="环境"
        right={env ? (
          <Button variant="ghost" size="sm" onClick={() => void refresh(env)} title="重新读取本机状态">
            <RefreshCw className="h-3.5 w-3.5" />
          </Button>
        ) : undefined}>
        <div className="flex items-center gap-2">
          <Select value={env} onValueChange={setEnv}>
            <SelectTrigger className="h-7 w-64 text-xs"><SelectValue placeholder="选择环境" /></SelectTrigger>
            <SelectContent>
              {(envs ?? []).map((e) => (
                <SelectItem key={e.name} value={e.name}>{e.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          {st?.dir && <span className="truncate text-[11px] text-muted-foreground" title={st.dir}>{st.dir}</span>}
        </div>
        {err && <div className="text-[11px] text-red-600 dark:text-red-400">读取状态失败：{err}</div>}
        {!envsLoaded && <div className="text-[11px] text-muted-foreground">正在读取环境…</div>}
        {envsLoaded && (envs ?? []).length === 0 && (
          <div className="text-[11px] text-muted-foreground">还没有配置环境（在「站点管理」里添加）。</div>
        )}
      </Card>

      {/* ============ 共用资料 ============ */}
      <Card id="card-devdata-shared" title="共用资料"
        description="整个环境共用：规格表与程序样板。" right={
          runningMirror ? (
            <Button variant="outline" size="sm" onClick={() => void stopJob('mirror')}>
              <Square className="mr-1 h-3 w-3" />停止
            </Button>
          ) : (
            <div className="flex items-center gap-1.5">
              <Button size="sm" onClick={() => void startMirror()} disabled={!env}>
                <Download className="mr-1 h-3 w-3" />同步共用资料
              </Button>
              <Button variant="outline" size="sm" onClick={() => void startMirror('dict')} disabled={!env}>仅规格表</Button>
              <Button variant="outline" size="sm" onClick={() => void startMirror('mdl')} disabled={!env}>仅模板</Button>
            </div>
          )
        }>
        <div className="grid grid-cols-4 gap-3">
          <Stat label="规格表" value={tableLabel} />
          <Stat label="模板" value={mdl && mdl.files > 0 ? `${mdl.templateCount} / ${mdl.files} 个文件` : '—'} />
          <Stat label="占用" value={totalBytes > 0 ? fmtBytes(totalBytes) : '—'} />
          <Stat label="上次同步" value={fmtTime(dict?.syncedAt || mdl?.syncedAt)} />
        </div>
        {dict && dict.skipped > 0 && (
          <div className="text-[11px] text-muted-foreground">
            有 {dict.skipped} 张表因体积过大被跳过（详见 tt gen status）。
          </div>
        )}
        {runningMirror && <JobProgress job={runningMirror} />}
      </Card>

      {/* ============ 按需资料 ============ */}
      <Card id="card-devdata-fetch" title="按需资料"
        description="一个程序一套：蓝图、表单规格、加点、排序骨架、产物与引子。">
        <div className="flex flex-wrap items-center gap-2">
          <Input className="h-7 w-28 text-xs" value={module} placeholder="模块（可自动）"
            onChange={(e) => setModule(e.target.value)} />
          <Input className="h-7 w-44 text-xs" value={prog} placeholder="程序编号，如 axmt500_wf"
            onChange={(e) => setProg(e.target.value)} />
          <label className="flex items-center gap-1 text-[11px] text-muted-foreground">
            <Checkbox checked={withDesign} onCheckedChange={(v) => setWithDesign(v === true)} />含设计稿
          </label>
          <label className="flex items-center gap-1 text-[11px] text-muted-foreground">
            <Checkbox checked={form} onCheckedChange={(v) => setForm(v === true)} />含 .42f
          </label>
          {runningFetch ? (
            <Button variant="outline" size="sm" onClick={() => void stopJob('fetch', runningFetch.prog)}>
              <Square className="mr-1 h-3 w-3" />停止
            </Button>
          ) : (
            <Button size="sm" onClick={() => {
              if (!prog.trim()) { toast.error('请填写程序编号'); return }
              void startFetch(prog)
            }} disabled={!env}>
              <Download className="mr-1 h-3 w-3" />拉取
            </Button>
          )}
        </div>

        {runningFetch && <JobProgress job={runningFetch} />}

        <Separator />
        {(st?.progs ?? []).length === 0 ? (
          <div className="flex items-center gap-2 py-2 text-[11px] text-muted-foreground">
            <FolderCog className="h-3.5 w-3.5" />这个环境还没有按程序拉取过。
          </div>
        ) : (
          <div className="space-y-1">
            {(st?.progs ?? []).map((p) => (
              <div key={p.prog} className="flex items-center gap-2 rounded-ui px-1 py-1 text-[11px] hover:bg-accent/40">
                <span className="w-40 shrink-0 truncate font-medium text-foreground" title={p.prog}>{p.prog}</span>
                <span className="w-24 shrink-0 text-muted-foreground">{fmtBytes(p.bytes)}</span>
                <span className="w-32 shrink-0 text-muted-foreground">{fmtTime(p.fetchedAt)}</span>
                <span className="w-20 shrink-0">
                  {p.stale
                    ? <Badge tone="yellow" title={(p.staleReasons ?? []).map((r) => r.detail).join('\n')}>已过期</Badge>
                    : <Badge tone="green">最新</Badge>}
                </span>
                <span className="min-w-0 flex-1 truncate text-muted-foreground" title={(p.warnings ?? []).join('\n')}>
                  {(p.missing ?? []).length > 0 ? `缺 ${(p.missing ?? []).length} 项` : ''}
                </span>
                <Button variant="ghost" size="sm" disabled={checking === p.prog}
                  onClick={() => void checkProg(p.prog)} title="与服务器核对是否过期">
                  {checking === p.prog ? '核对中…' : '核对'}
                </Button>
                <Button variant="ghost" size="sm" disabled={!!runningFetch || !env}
                  onClick={() => void startFetch(p.prog, true)} title="重新拉取（整包替换）">重新拉取</Button>
                <Button variant="ghost" size="sm" onClick={() => setPendingDel(p)} title="删除这个程序的材料">
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            ))}
          </div>
        )}
      </Card>

      <AlertDialog open={pendingDel !== null} onOpenChange={(o) => { if (!o) setPendingDel(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除 {pendingDel?.prog} 的材料？</AlertDialogTitle>
            <AlertDialogDescription>
              会删除本机上这份程序材料（{fmtBytes(pendingDel?.bytes ?? 0)}）。删除后要重新拉取才能恢复。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void removeProg()}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}

/** 一项统计（标签 + 值）。 */
function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <div className="text-[11px] text-muted-foreground">{label}</div>
      <div className="truncate text-xs text-foreground" title={value}>{value}</div>
    </div>
  )
}

/** 任务进度条：显示当前文件与"已完成/总数"。 */
function JobProgress({ job }: { job: DevDataJob }) {
  const pct = job.total > 0 ? Math.min(100, Math.round((job.done / job.total) * 100)) : 0
  return (
    <div className="space-y-1">
      <Progress value={pct} className={cn('h-1.5')} />
      <div className="flex items-center justify-between gap-2 text-[11px] text-muted-foreground">
        <span className="min-w-0 truncate">{job.message || '处理中…'}</span>
        <span className="shrink-0">
          {job.total > 0 ? `${job.done}/${job.total}` : ''}{job.currentFile ? ` · ${job.currentFile}` : ''}
        </span>
      </div>
    </div>
  )
}

/** 时间显示：manifest 里是 RFC3339，界面上只要"哪天几点"。 */
function fmtTime(s?: string): string {
  if (!s) return '—'
  const d = new Date(s)
  if (Number.isNaN(d.getTime())) return s
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}`
}
