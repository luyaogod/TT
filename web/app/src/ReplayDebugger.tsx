// 回放会话的**两栏调试器**（用户 2026-10-18 定的形态）。
//
// 左：整份计划（像代码行一样，回放开始前就铺满），行首是**断点**（点一下 = 到这一步
//      "发之前"停住），焦点行跟着进度滚动。
// 右：这一轮的**实时报文流**（三泳道 + 帧表），不跟随焦点 —— 它回答的是"现在运行期在吐什么"。
//
// 顶上是**调试器手柄**（继续 / 暂停 / 发一步 / 跳过 / 停止），与 Debug 页同一套肌肉记忆：
// 左边看"要执行哪一行"，右边看"执行结果"，中间用断点/单步控制节奏。
import { useEffect, useMemo, useRef, useState } from 'react'
import { Pause, Play, SkipForward, Square } from 'lucide-react'
import { api, type ReplayPlanStep, type ReplayStep, type ReplayStepsResp } from './api'
import { FepFrames } from './FepMonitor'
import { useToast } from './toast'

type Cmd = 'pause' | 'resume' | 'step' | 'skip' | 'stop'

const STATUS_LABEL: Record<string, { text: string; cls: string }> = {
  sent: { text: '已回包', cls: 'text-emerald-600 dark:text-emerald-400' },
  waiting: { text: '等待中', cls: 'text-sky-600 dark:text-sky-400' },
  'no-response': { text: '无应答', cls: 'text-amber-600 dark:text-amber-400' },
  skipped: { text: '跳过', cls: 'text-muted-foreground' },
  refused: { text: '拒绝', cls: 'text-red-600 dark:text-red-400' },
  failed: { text: '失败', cls: 'text-red-600 dark:text-red-400' },
  paused: { text: '暂停', cls: 'text-violet-600 dark:text-violet-400' },
}

export function ReplayDebugger({ sessionId, name }: { sessionId: string; name?: string }) {
  const toast = useToast()
  const [rp, setRp] = useState<ReplayStepsResp | null>(null)
  const [plan, setPlan] = useState<ReplayPlanStep[]>([])
  const [skips, setSkips] = useState<{ raw: string; reason: string }[]>([])
  const [planErr, setPlanErr] = useState('')
  const listRef = useRef<HTMLDivElement | null>(null)

  // 进度：每秒问一次（与录制明细页同一个端点语义）。
  useEffect(() => {
    let cancelled = false
    const tick = async () => {
      try {
        const r = await api.replayLatest(name || '')
        if (!cancelled) setRp(r)
      } catch { /* 下一拍再试 */ }
    }
    void tick()
    const h = window.setInterval(tick, 1000)
    return () => { cancelled = true; window.clearInterval(h) }
  }, [name])

  // 计划：左栏要"提前铺满"——录制 ID 一变就重拉（回放开始之前就能看）。
  const recId = rp?.id || ''
  useEffect(() => {
    if (!recId) return
    let cancelled = false
    setPlanErr('')
    void api.replayPlan(recId).then((r) => {
      if (cancelled) return
      setPlan(r.steps || [])
      setSkips(r.skips || [])
    }).catch((e: any) => {
      if (!cancelled) setPlanErr(e?.message || String(e))
    })
    return () => { cancelled = true }
  }, [recId])

  const progress = rp?.progress
  const running = !!rp?.running
  const paused = !!progress?.paused
  const breaks = useMemo(() => new Set(rp?.breaks || []), [rp?.breaks])

  // 每步的最新状态（同一 n 取最后一行）
  const stateByN = useMemo(() => {
    const m = new Map<number, ReplayStep>()
    for (const s of rp?.steps || []) m.set(s.n, s)
    return m
  }, [rp?.steps])

  // 焦点行：优先"正在等回包的那一步"，否则最后有状态的那一步，再否则第 1 步。
  const focusN = useMemo(() => {
    const waiting = (rp?.steps || []).find((s) => s.result === 'waiting')
    if (waiting) return waiting.n
    let last = 0
    for (const s of rp?.steps || []) if (s.n > last) last = s.n
    return last || 1
  }, [rp?.steps])

  // 焦点行滚进视野（列表自己滚，不带动整页）。
  useEffect(() => {
    const box = listRef.current
    if (!box) return
    const row = box.querySelector<HTMLElement>(`[data-step="${focusN}"]`)
    row?.scrollIntoView({ block: 'nearest' })
  }, [focusN, plan.length])

  const ctl = async (cmd: Cmd) => {
    if (!rp?.id) return
    try {
      await api.replayCtl(rp.id, cmd)
    } catch (e: any) {
      toast.error('回放控制失败', { description: e?.message || String(e) })
    }
  }

  const toggleBreak = async (n: number) => {
    if (!rp?.id) return
    try {
      const r = await api.replayBreak(rp.id, n, !breaks.has(n))
      setRp((prev) => (prev ? { ...prev, breaks: r.breaks } : prev))
    } catch (e: any) {
      toast.error('断点没设上', { description: e?.message || String(e) })
    }
  }

  const btn = 'inline-flex items-center gap-1 rounded-ui border border-border px-2 py-1 text-[11px] hover:bg-accent disabled:opacity-40 disabled:hover:bg-transparent'

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      {/* 手柄：与 Debug 页同一套（继续 / 暂停 / 发一步 / 跳过 / 停止） */}
      <div className="flex shrink-0 flex-wrap items-center gap-2 border-b border-border bg-card/40 px-2 py-1.5 text-[11px]">
        <span className="font-mono">
          {running
            ? (paused ? `已停在第 ${focusN} 步` : `正在回放 · 第 ${progress?.done ?? 0}/${progress?.total ?? plan.length} 步`)
            : (progress ? `完成 ${progress.done}/${progress.total}` : '还没有回放')}
        </span>
        <span className="text-muted-foreground">
          无应答 {progress?.noResp ?? 0} · 跳过 {progress?.skipped ?? 0}
          {(progress?.degraded ?? 0) > 0 ? ` · 降级 ${progress?.degraded}` : ''} · 失败 {progress?.failed ?? 0}
        </span>
        <span className="ml-auto flex items-center gap-1">
          {paused ? (
            <>
              <button onClick={() => void ctl('step')} className={btn} title="发这一步（发完继续停住）">
                <Play size={11} /> 发一步
              </button>
              <button onClick={() => void ctl('skip')} className={btn} title="跳过这一步（不发，继续停着）">
                <SkipForward size={11} /> 跳过
              </button>
              <button onClick={() => void ctl('resume')} className={btn} title="继续跑到下一个断点（或跑完）">
                继续
              </button>
            </>
          ) : (
            <button onClick={() => void ctl('pause')} className={btn} disabled={!running} title="暂停：当前这一步回包后生效">
              <Pause size={11} /> 暂停
            </button>
          )}
          <button onClick={() => void ctl('stop')} className={`${btn} text-red-600 dark:text-red-400`}
            disabled={!running} title="停止：不再发后面的步骤">
            <Square size={11} /> 停止
          </button>
        </span>
      </div>

      {/* 两栏：左计划 / 右实时流 */}
      <div className="flex min-h-0 flex-1 overflow-hidden">
        <div className="flex min-h-0 w-[46%] min-w-[360px] flex-col border-r border-border">
          <div className="flex shrink-0 items-center gap-2 border-b border-border px-2 py-1 text-[10px] text-muted-foreground">
            <span>计划 {plan.length ? `${plan.length} 步` : '…'}</span>
            {skips.length > 0 && <span title={skips.map((s) => s.reason).join('；')}>白名单跳过 {skips.length} 条</span>}
            <span className="ml-auto">点行首圆点设断点（发之前停）</span>
          </div>
          {planErr && <div className="px-2 py-1 text-[10px] text-red-600 dark:text-red-400">读计划失败：{planErr}</div>}
          <div ref={listRef} className="min-h-0 flex-1 overflow-auto font-mono text-[11px]">
            {plan.map((st) => {
              const cur = stateByN.get(st.n)
              const stt = cur ? STATUS_LABEL[cur.result] : undefined
              const isFocus = st.n === focusN
              const hasBp = breaks.has(st.n)
              return (
                <div key={st.n} data-step={st.n}
                  className={`flex items-start gap-1.5 border-b border-border/40 px-1.5 py-0.5 ${isFocus ? 'bg-accent/40' : 'hover:bg-accent/20'}`}>
                  <button onClick={() => void toggleBreak(st.n)} disabled={!rp?.id}
                    title={hasBp ? `取消第 ${st.n} 步的断点` : `在第 ${st.n} 步（发之前）停住`}
                    className="mt-[3px] h-2.5 w-2.5 shrink-0 rounded-full border border-muted-foreground/50 disabled:opacity-30"
                    style={{ background: hasBp ? 'var(--dsh-accent)' : 'transparent' }} />
                  <span className="w-8 shrink-0 text-right text-muted-foreground">{st.n}</span>
                  <span className="min-w-0 flex-1 truncate" title={st.event}>{st.event}</span>
                  {st.degraded && (
                    <span className="shrink-0 text-[9px] text-amber-600 dark:text-amber-400" title="没有正式身份：真跑时按录制时的临时编号发">降级</span>
                  )}
                  {stt && <span className={`shrink-0 text-[9px] ${stt.cls}`}>{stt.text}</span>}
                </div>
              )
            })}
            {plan.length === 0 && !planErr && (
              <div className="p-3 font-sans text-[11px] text-muted-foreground">还没有计划：回放跑起来之后这里会铺满整份事件。</div>
            )}
          </div>
        </div>
        <div className="flex min-h-0 min-w-0 flex-1 flex-col">
          <FepFrames sessionId={sessionId} />
        </div>
      </div>
    </div>
  )
}
