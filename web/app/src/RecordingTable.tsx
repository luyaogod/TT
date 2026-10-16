// 全文流记录表（用户 2026-10-16 定的形态：**只有一页、只有三列**）。
//
// 之前的"分层记录 / 全文流"两个页签被去掉了：一屏就一条流，行是**帧**，列是
// **时间 / 向 / 文本**。简化掉的列（步号、迷你轴、目标、字节、耗时）不是丢了信息，
// 而是换了地方：
//
//	名字与置信度  → 内联在文本列里（`录入` / `Value#5485 未解析`）—— 这才是"点了什么"
//	回应与耗时    → 右侧明细面板（点一下这一行就看到：引起的回应、流量、名字来源）
//	位置          → 上面的条带（那里才有轴）
//
// 行高固定 22px、窗口化渲染（自研，不引第三方）。行 key 用 `连接:方向:字节偏移` ——
// **不是步号**（步号随 --stop-at/--skip 变，不是身份）。
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { RecordPrecision, RecordingAction, RecordingFrame, ReplayStep } from './api'
import { kindLabel, kindTagClass } from './recordingKind'
import { fmtBytes, fmtClock, fmtRel } from './recordingTime'

const ROW_H = 22
/** 上下各多渲染几行:滚动时不会露出空白(但也不会把 DOM 撑起来)。 */
const OVERSCAN = 6

/** 行 key:`连接:方向:字节偏移`。 */
export function frameKey(f: { c: number; d: number; o: number }): string {
  return `${f.c}:${f.d}:${f.o}`
}

interface Props {
  /** 全部帧（已按时间/字节序排好；全文流那一侧可能只加载了前几页）。 */
  frames: RecordingFrame[]
  precision: RecordPrecision
  sel: RecordingFrame | null
  onSelect: (f: RecordingFrame) => void
  cursor: number
  onCursor: (v: number) => void
  /** 帧 → 语义动作（名字与回应都从这里来）。查不到 = 这一帧不是用户动作。 */
  actionOf?: (f: RecordingFrame) => RecordingAction | undefined
  /** 事件文本 → 回放步（有回放日志时才有）。 */
  stepByEvent?: (ev: string) => ReplayStep | undefined
  /** 服务端说一共多少帧 / 是否还有下一页。 */
  total: number
  onLoadMore: () => void
}

export function RecordingTable(p: Props) {
  const { frames, precision, actionOf } = p
  const scrollRef = useRef<HTMLDivElement | null>(null)
  const [scrollTop, setScrollTop] = useState(0)
  const [viewH, setViewH] = useState(0)
  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const ro = new ResizeObserver(() => setViewH(el.clientHeight))
    ro.observe(el)
    setViewH(el.clientHeight)
    return () => ro.disconnect()
  }, [])

  const total = frames.length
  const first = Math.max(0, Math.floor(scrollTop / ROW_H) - OVERSCAN)
  const last = Math.min(total, Math.ceil((scrollTop + (viewH || 400)) / ROW_H) + OVERSCAN)
  const slice = frames.slice(first, last)

  // 滚到底自动补下一页。留一屏余量，用户不会看到"滚到底才开始转圈"。
  const onScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    const el = e.currentTarget
    setScrollTop(el.scrollTop)
    if (frames.length < p.total && el.scrollTop + el.clientHeight > el.scrollHeight - ROW_H * 8) {
      p.onLoadMore()
    }
  }, [p, frames.length])

  const t0 = useMemo(() => frames.find((f) => f.t > 0)?.t || 0, [frames])
  const clock = (t: number) => (precision === 'frame' ? fmtClock(t) : '—')

  return (
    // 底色与右侧明细面板**同一个** `bg-card`（用户 2026-10-16）：一屏里"记录"与"明细"是
    // 同一块面板的两半，底色不同就会看起来像两个不相干的东西。
    <div className="flex min-h-0 min-w-0 flex-1 flex-col bg-card">
      {/* 没有表头了（用户 2026-10-16 定的）：四列的含义从行本身看得出来 ——
          类型是彩色 tag、时间是时钟、向是箭头、剩下的是文本。表头占掉 24px 却不带来新信息。 */}
      <div ref={scrollRef} onScroll={onScroll} className="min-h-0 flex-1 overflow-auto">
        <div style={{ minWidth: 460 }}>
          <div style={{ height: total * ROW_H, position: 'relative' }}>
            {slice.map((f, i) => (
              <div key={frameKey(f)} style={{ position: 'absolute', top: (first + i) * ROW_H, left: 0, right: 0, height: ROW_H }}>
                <Row f={f} p={p} clock={clock} t0={t0} action={actionOf ? actionOf(f) : undefined} />
              </div>
            ))}
          </div>
          {total === 0 && <div className="p-4 text-center text-xs text-muted-foreground">这一份录制里没有帧。</div>}
          {frames.length < p.total && (
            <div className="p-2 text-center text-[11px] text-muted-foreground">
              已加载 {frames.length} / {p.total} 帧，滚动到底继续加载…
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

/** 一行。四列：类型 / 时间 / 向 / 文本（名字、原始载荷、回放结果都内联在文本里）。 */
function Row({ f, p, clock, t0, action }: {
  f: RecordingFrame
  p: Props
  clock: (t: number) => string
  t0: number
  action?: RecordingAction
}) {
  const sel = !!p.sel && frameKey(p.sel) === frameKey(f)
  const tgt = action?.target
  const named = !!tgt && tgt.confidence !== 'unresolved' && !!tgt.name
  const step = p.stepByEvent && action ? action.events.map(p.stepByEvent).find(Boolean) : undefined
  return (
    // 每行一条底边线（用户 2026-10-16）：固定 22px 的行里，一条 1px 的线是**唯一**能让
    // 眼睛横着走的东西 —— 没有它，一列时间读起来会串行。
    // 行高仍是 22px（border-box，线算在行高里），窗口化的算术不受影响。
    <div className={`flex h-[22px] cursor-pointer items-center gap-2 overflow-hidden border-b border-border/50 px-2 font-mono text-[11px]
      ${sel ? 'bg-amber-500/10' : 'hover:bg-accent/30'}`}
      onClick={() => p.onSelect(f)}
      onMouseEnter={() => f.t > 0 && p.onCursor(f.t)}
      title={`${f.k}${f.q ? ` #${f.q}` : ''} · ${f.d === 1 ? '服务→前端' : '前端→服务'} · 连接 ${f.c} · ${fmtBytes(f.n)} @ 偏移 ${f.o}`
        + (f.t > 0 ? `\n${fmtClock(f.t)}（${fmtRel(f.t, t0)}）` : '\n时刻未知（无帧索引）')}>
      {/* 类型：与条带上的块**同一个色**（颜色只在 recordingKind.ts 里定义一次）。
          右对齐、紧贴时间列 —— 一屏扫下来颜色成一条竖线，比看文字快。 */}
      <div className="flex w-[52px] shrink-0 justify-end">
        <span className={`rounded-ui px-1 text-[10px] leading-4 ${kindTagClass(f.k)}`}>{kindLabel(f.k)}</span>
      </div>
      <div className={`w-[92px] shrink-0 ${f.t > 0 ? 'text-muted-foreground' : 'text-muted-foreground/60'}`}>{clock(f.t)}</div>
      <div className="w-4 shrink-0 text-center text-muted-foreground">{f.d === 1 ? '→' : '←'}</div>

      <div className="flex min-w-0 flex-1 items-center gap-2">
        {/* 序号（类别已经在类型列里了，这里不重复） */}
        {f.q ? <span className="shrink-0 text-muted-foreground/70">#{f.q}</span> : null}

        {/* 名字：解析出来就给名字（带类型#id），解析不出来就照实显示 id + 未解析 —— 不编名字 */}
        {action && (
          named ? (
            <span className="shrink-0" title={`${tgt!.name}｜${tgt!.confidence === 'high' ? '高置信' : '低置信（仅排障）'}｜来源 ${tgt!.source || '—'}`}>
              <span className="text-foreground">{tgt!.name}</span>
              <span className="pl-1 text-muted-foreground">{tgt!.type}#{tgt!.id}</span>
            </span>
          ) : (
            <span className="shrink-0 text-muted-foreground" title={tgt?.why || '这个名字没解出来'}>
              {tgt?.type || '?'}#{tgt?.id || action.id} <span className="text-muted-foreground/70">未解析</span>
            </span>
          )
        )}

        {/* 原始载荷（证据本身）。event 帧的载荷在文本里已经带 idRef 了，不再重复。 */}
        {f.preview && <span className="min-w-0 flex-1 truncate text-muted-foreground/90">{f.preview}</span>}
        {!f.preview && <span className="min-w-0 flex-1" />}

        {/* 回放走到这一步了没有（有回放日志时才挂） */}
        {step && <StepChip s={step} />}
      </div>
    </div>
  )
}

/** 回放结果徽标（与回放进度条同一套说法，这里只有一格宽所以用短词）。 */
function StepChip({ s }: { s: ReplayStep }) {
  const m: Record<string, { t: string; c: string; why: string }> = {
    sent: { t: '已确认', c: 'border-emerald-500/30 text-emerald-600 dark:text-emerald-400', why: '回放发出去了，运行期回了包' },
    'no-response': { t: '无应答', c: 'border-border text-muted-foreground', why: '回放发出去了，运行期没回包（本来就没应答）' },
    skipped: { t: '跳过', c: 'border-border text-muted-foreground', why: s.reason || '回放按要求跳过' },
    failed: { t: '失败', c: 'border-red-500/30 text-red-600 dark:text-red-400', why: s.reason || '回放发送出错' },
    waiting: { t: '进行中', c: 'border-sky-500/30 text-sky-600 dark:text-sky-400', why: '已发出，正在等运行期回包' },
    paused: { t: '已暂停', c: 'border-border text-muted-foreground', why: '回放停在这一步' },
  }
  const x = s.suspect && s.result === 'no-response'
    ? { t: '无回包?', c: 'border-amber-500/40 text-amber-600 dark:text-amber-400', why: '点击类事件本该有回包，8 秒内没有（可疑）' }
    : (m[s.result] || { t: s.result, c: 'border-border text-muted-foreground', why: '' })
  return <span className={`shrink-0 rounded-ui border px-1 text-[10px] ${x.c}`} title={`回放第 ${s.n} 步：${x.why}`}>{x.t}</span>
}

