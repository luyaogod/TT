// 设置页「技能 → 分发」：项目行 + 五平台矩阵 + 同步。
//
// 页面刻意只留三块：项目行、矩阵、底部计数 + 同步。没有图例、没有说明段、
// 没有版本号：勾选状态就代表“这个平台该不该有它”，点同步就是覆盖。
import { useCallback, useEffect, useMemo, useState } from 'react'
import { AlertCircle, Check, Minus, RefreshCw } from 'lucide-react'
import {
  api, type SkillDeployResult, type SkillMatrixRow, type SkillsOverview, type SkillTargetState,
} from './api'
import {
  Badge, Button, Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '../../shared/ui'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '../../shared/ui-radix'
import { DirPicker } from './DirPicker'
import { useToast } from './toast'
import { cn } from '../../shared/utils'

const PROJECT_KEY = 'tt.skills.project'

const key = (target: string, skill: string) => target + '::' + skill

function savedProject(): string {
  try { return localStorage.getItem(PROJECT_KEY) || '' } catch { return '' }
}

function rememberProject(project: string) {
  try { localStorage.setItem(PROJECT_KEY, project) } catch { /* 隐私模式/禁用存储：不影响本次 */ }
}

/** 首次进入的勾选：有清单按清单；没有清单就替用户把 `.agents` + 已检测平台全勾上。 */
function initialPicked(data: SkillsOverview): Set<string> {
  const avail = new Set(data.skills.filter((s) => s.available).map((s) => s.id))
  const out = new Set<string>()
  for (const t of data.targets) {
    for (const name of t.skills) {
      if (avail.has(name)) out.add(key(t.id, name))
    }
  }
  if (!data.manifest) {
    for (const t of data.targets) {
      if (t.id !== 'agents' && !t.detected) continue
      for (const s of data.skills) {
        if (s.available) out.add(key(t.id, s.id))
      }
    }
  }
  return out
}

export function SkillsDeploy() {
  const toast = useToast()
  const [data, setData] = useState<SkillsOverview | null>(null)
  const [picked, setPicked] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [picking, setPicking] = useState(false)
  const [confirm, setConfirm] = useState<{ removals: { target: string; skill: string }[]; targets: { id: string; skills: string[] }[] } | null>(null)

  const load = useCallback(async (project?: string) => {
    setLoading(true)
    setErr('')
    try {
      const r = await api.skillsOverview(project)
      setData(r)
      rememberProject(r.project)
      setPicked(initialPicked(r))
    } catch (e: any) {
      setErr(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load(savedProject() || undefined) }, [load])

  const avail = useMemo(() => (data?.skills ?? []).filter((s) => s.available), [data])

  if (loading && !data) {
    return <div className="px-1 py-6 text-[11px] text-muted-foreground">加载中…</div>
  }
  if (err && !data) {
    return (
      <div className="space-y-2 rounded-ui border border-border p-3">
        <div className="flex items-start gap-2 text-[11px] text-red-600 dark:text-red-400">
          <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span className="break-all">{err}</span>
        </div>
        <Button size="sm" variant="outline" onClick={() => void load(savedProject() || undefined)}>
          <RefreshCw className="h-3.5 w-3.5" />重试
        </Button>
        <Button size="sm" variant="ghost" onClick={() => {
          try { localStorage.removeItem(PROJECT_KEY) } catch { /* 忽略 */ }
          void load()
        }}>用当前目录</Button>
      </div>
    )
  }
  if (!data) return null

  const toggle = (t: SkillTargetState, s: SkillMatrixRow) => {
    if (!s.available) return
    setPicked((prev) => {
      const next = new Set(prev)
      const k = key(t.id, s.id)
      if (next.has(k)) next.delete(k)
      else next.add(k)
      return next
    })
  }

  const toggleColumn = (t: SkillTargetState) => {
    setPicked((prev) => {
      const next = new Set(prev)
      const all = avail.length > 0 && avail.every((s) => next.has(key(t.id, s.id)))
      for (const s of avail) {
        const k = key(t.id, s.id)
        if (all) next.delete(k)
        else next.add(k)
      }
      return next
    })
  }

  const selTargets = data.targets.filter((t) => avail.some((s) => picked.has(key(t.id, s.id)))).length
  const selSkills = avail.filter((s) => data.targets.some((t) => picked.has(key(t.id, s.id)))).length

  const availNames = new Set(avail.map((s) => s.id))
  const removals = data.targets.flatMap((t) => t.skills
    .filter((n) => availNames.has(n) && !picked.has(key(t.id, n)))
    .map((n) => ({ target: t.id, skill: n })))

  const buildTargets = () => data.targets.map((t) => ({
    id: t.id,
    skills: avail.filter((s) => picked.has(key(t.id, s.id))).map((s) => s.id),
  }))

  const run = async (targets: { id: string; skills: string[] }[]) => {
    if (busy) return
    setBusy(true)
    try {
      const r: SkillDeployResult = await api.skillsDeploy(data.project, targets)
      const files = r.written.reduce((n, w) => n + (w.files?.length || 0), 0)
      const removed = r.removed.length
      toast.success(`已同步：写入 ${files} 个文件` + (removed ? `，删除 ${removed} 个` : ''))
      await load(data.project)
    } catch (e: any) {
      toast.error(e?.message || String(e))
    } finally {
      setBusy(false)
    }
  }

  const sync = () => {
    const targets = buildTargets()
    if (removals.length > 0) {
      setConfirm({ removals, targets })
      return
    }
    void run(targets)
  }

  const targetName = (id: string) => data.targets.find((t) => t.id === id)?.dir || id

  const tri = (t: SkillTargetState) => {
    if (avail.length === 0) return 'off'
    const n = avail.filter((s) => picked.has(key(t.id, s.id))).length
    return n === 0 ? 'off' : n === avail.length ? 'on' : 'ind'
  }

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-[11px] text-muted-foreground">项目</span>
        <code className="max-w-[560px] truncate rounded-ui border border-border bg-muted/40 px-1.5 py-0.5 font-mono text-[11px]"
          title={data.project}>{data.project}</code>
        <Button size="sm" variant="outline" disabled={busy} onClick={() => setPicking(true)}>更换</Button>
      </div>

      {data.officialError && (
        <div className="flex items-start gap-2 rounded-ui border border-red-500/40 p-2 text-[11px] text-red-600 dark:text-red-400">
          <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          <span className="break-all">{data.officialError}</span>
        </div>
      )}

      <Table className="min-w-[880px] text-[11px]">
        <TableHeader>
          <TableRow>
            <TableHead className="w-[240px] align-top">
              <div className="px-2 py-1.5 text-muted-foreground">技能</div>
            </TableHead>
            {data.targets.map((t) => (
              <TableHead key={t.id} className="min-w-[128px] align-top"
                title={t.detected ? '已存在' : '未检测到；勾选后同步会创建目录'}>
                <div className="px-2 py-1.5">
                  <div className="flex items-center gap-1.5">
                    <span className={cn('size-1.5 rounded-full', t.detected ? 'bg-emerald-500' : 'bg-muted-foreground/40')} />
                    <span className="text-[11px] font-medium text-foreground">{t.name}</span>
                  </div>
                  <div className="mt-0.5 truncate font-mono text-[10px] font-normal text-muted-foreground" title={t.dir}>{t.dir}</div>
                  <button type="button" onClick={() => toggleColumn(t)} title="全选 / 清空"
                    className="mt-1 flex w-full flex-col items-center gap-0.5 text-[10px] text-muted-foreground hover:text-foreground">
                    <CheckBox checked={tri(t)} />
                    <span>全选</span>
                  </button>
                </div>
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.skills.map((s) => (
            <TableRow key={s.id}>
              <TableCell className="w-[240px] whitespace-normal">
                <div className="px-2 py-1.5">
                  <div className="flex items-center gap-1.5">
                    <code className="truncate font-mono text-[11px] text-foreground" title={s.name}>{s.name}</code>
                    {!s.available && <Badge tone="red">不可用</Badge>}
                  </div>
                  {!s.available && s.why && (
                    <div className="mt-0.5 line-clamp-2 text-[10px] leading-4 text-muted-foreground">{s.why}</div>
                  )}
                </div>
              </TableCell>
              {data.targets.map((t) => {
                const on = picked.has(key(t.id, s.id))
                return (
                  <TableCell key={t.id} className="p-0 text-center">
                    <button type="button" disabled={!s.available} onClick={() => toggle(t, s)}
                      title={s.available ? (on ? '取消勾选：同步时删除' : '勾选：同步时写入') : s.why}
                      className="flex h-7 w-full items-center justify-center hover:bg-muted/40 disabled:cursor-not-allowed">
                      <CheckBox checked={on ? 'on' : 'off'} disabled={!s.available} />
                    </button>
                  </TableCell>
                )
              })}
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {data.skills.length === 0 && (
        <div className="text-[11px] text-muted-foreground">没有可用技能</div>
      )}

      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="text-[11px] text-muted-foreground">
          已选 {selTargets} 个平台 · {selSkills} 个技能
          {removals.length > 0 && <span className="ml-2 text-red-600 dark:text-red-400">将删除 {removals.length} 个</span>}
        </span>
        <Button size="sm" onClick={sync} disabled={busy || !!data.officialError || (selTargets === 0 && removals.length === 0)}>
          {busy ? '同步中…' : '同步'}
        </Button>
      </div>

      <DirPicker
        open={picking}
        initial={data.project}
        title="选择项目目录"
        onPick={(p) => { setPicking(false); void load(p) }}
        onClose={() => setPicking(false)}
      />

      <AlertDialog open={confirm !== null} onOpenChange={(o) => { if (!o) setConfirm(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>同步会删除 {confirm?.removals.length} 个目录</AlertDialogTitle>
            <AlertDialogDescription>
              这些条目在清单里，是 TT 装的；取消勾选后同步会删掉它们。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="max-h-40 overflow-auto rounded-ui border border-border p-2 font-mono text-[11px] text-muted-foreground">
            {confirm?.removals.map((r) => (
              <div key={r.target + '::' + r.skill}>{targetName(r.target)}/{r.skill}</div>
            ))}
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => { const c = confirm; setConfirm(null); if (c) void run(c.targets) }}>
              继续同步
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

/** 自绘勾选框：整格可点，不依赖 Radix 的 16px 命中区。 */
function CheckBox({ checked, disabled }: { checked: 'on' | 'off' | 'ind'; disabled?: boolean }) {
  return (
    <span className={cn(
      'flex size-3.5 shrink-0 items-center justify-center rounded-[2px] border',
      disabled
        ? 'border-input/50 opacity-40'
        : checked === 'on'
          ? 'border-primary bg-primary text-primary-foreground'
          : checked === 'ind'
            ? 'border-primary bg-primary/20 text-primary'
            : 'border-input bg-transparent',
    )}>
      {checked === 'on' && <Check className="size-2.5" strokeWidth={3} />}
      {checked === 'ind' && <Minus className="size-2.5" strokeWidth={3} />}
    </span>
  )
}
