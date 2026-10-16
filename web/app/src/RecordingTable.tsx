// 三层记录表(计划 C 的 D5):**语义动作 > 原始事件 > 运行期回应**。
//
// 为什么是三层的:H2 实测一份录制有 383 帧 / 事件侧 34 帧 / 63 个原始事件 / 17 个语义动作。
// 只报"帧数"说不清"用户做了什么",只报"动作数"又说不清"线上到底过了什么" —— 三层各是一层事实,
// 界面上就必须是三层(而且每层的数字都要对得上,见 counts)。
//
// 两个页签的差别不是详略,而是**看的层次**:
//
//	分层记录  —— 一个语义动作一行(可展开看它的回应),回答"用户做了什么、有没有回包";
//	全文流    —— 一帧一行(含图片缓存),回答"这条线上到底过了什么字节"。
//
// 虚拟化是**自研的固定行高**(22px)窗口化,不引第三方:我们行高固定、只需要窗口化,
// 为一个界面加依赖不值(计划 C 决策 ⑥)。行 key 一律 `连接:方向:字节偏移` —— **不是步号**:
// 步号会随 `--stop-at/--skip` 变,不是身份(H3)。
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { ChevronDown, ChevronRight, CornerDownLeft, MousePointerClick, Type as TypeIcon } from 'lucide-react'
import type { RecordPrecision, RecordingAction, RecordingFrame, ReplayStep } from './api'
import { fmtBytes, fmtClock, fmtDur, fmtRel } from './recordingTime'

const ROW_H = 22
/** 上下各多渲染几行:滚动时不会露出空白(但也不会把 DOM 撑起来)。 */
const OVERSCAN = 6
/** 表格的最小宽度:窄容器下横向滚动,而不是把每一列压成一个字一行。 */
const MIN_W = 840

/** 行 key:`连接:方向:字节偏移`。 */
export function frameKey(f: { c: number; d: number; o: number }): string {
  return `${f.c}:${f.d}:${f.o}`
}

/** 带预览的帧(回应的预览来自服务端;全文流的预览来自 timeline 端点)。 */
type FrameRow = RecordingFrame & { preview?: string; note?: string }

/** 一行。三层的形状不同,但**行高相同** —— 虚拟化按固定行高算,不许某一行长得不一样。 */
type Row =
  | { kind: 'action'; key: string; a: RecordingAction }
  | { kind: 'frame'; key: string; f: FrameRow }
  | { kind: 'response'; key: string; f: FrameRow; n: number }

interface Props {
  tab: 'steps' | 'stream'
  actions: RecordingAction[]
  /** 全部帧(概览给的、已排好序)。分层页签用它找出"没有动作的事件帧"。 */
  frames: RecordingFrame[]
  precision: RecordPrecision
  /** 迷你轴:与概览条带**同一个** x()(见 recordingTime.ts 的纪律)。 */
  xOf: (f: RecordingFrame) => number
  miniW: number
  cursor: number
  onCursor: (v: number) => void
  sel: RecordingFrame | null
  onSelectFrame: (f: RecordingFrame) => void
  selAction: number | null
  onSelectAction: (a: RecordingAction) => void
  /**
   * 这一步**回放到哪了**（按事件文本查）。
   *
   * 有回放日志时，动作行上会多一个结果徽标 —— 把"用户点的那一下"与"回放有没有把它发出去、
   * 运行期回没回"对到一起看，是这一页最有用的一个交叉。没回过放时它是 undefined，
   * 表里也就不挂徽标（**不显示"未知"** —— 没回过放不是一种状态）。
   */
  stepByEvent?: (ev: string) => ReplayStep | undefined
  /** 全文流那一侧当前已加载的帧(分页)与总数。 */
  streamFrames: RecordingFrame[]
  streamTotal: number
  onLoadMore: () => void
}

export function RecordingTable(p: Props) {
  const { tab, actions, frames, precision, xOf, miniW } = p
  // 折叠状态按行 key 记。切换页签/换一份录制时由父级的 key 复位(见 RecordingDetail)。
  const [collapsed, setCollapsed] = useState<Set<string>>(() => new Set())

  const rows = useMemo<Row[]>(() => {
    if (tab === 'stream') {
      return p.streamFrames.map((f) => ({ kind: 'frame' as const, key: frameKey(f), f }))
    }
    // 分层:按**事件侧帧的流顺序**走一遍,有动作的展开成一组,没有的单独一行 ——
    // "34 个事件帧里只有 17 个有动作"这件事在表里就看得见(H2)。
    const byFrame = new Map<string, RecordingAction[]>()
    for (const a of actions) {
      const k = frameKey(a.frame)
      const list = byFrame.get(k)
      if (list) list.push(a)
      else byFrame.set(k, [a])
    }
    const out: Row[] = []
    for (const f of frames) {
      if (f.k !== 'event') continue
      const k = frameKey(f)
      const acts = byFrame.get(k)
      if (!acts) {
        out.push({
          kind: 'frame', key: k, f,
          // 照实说清"为什么这一帧没有动作":里面全是导航/状态上报(语义层丢掉的那些)
          // 这一句不能省 —— 否则看起来像"漏了一步"。
        })
        continue
      }
      for (const a of acts) {
        const gk = `a${a.n}`
        out.push({ kind: 'action', key: gk, a })
        out.push({ kind: 'frame', key: `${gk}:f`, f: a.frame })
        if (collapsed.has(gk)) continue
        for (const r of a.responses || []) {
          // 回应永远是"运行期→前端"(d=1);补成帧的形状,好让三层的列宽/迷你轴共用一套渲染。
          out.push({
            kind: 'response', key: `${gk}:r:${r.c}:${r.o}`, n: a.n,
            f: { o: r.o, n: r.n, t: r.t, c: r.c, d: 1, k: r.k, q: r.q, r: 0, preview: r.preview },
          })
        }
      }
    }
    return out
  }, [tab, actions, frames, collapsed, p.streamFrames])

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

  const total = rows.length
  const first = Math.max(0, Math.floor(scrollTop / ROW_H) - OVERSCAN)
  const last = Math.min(total, Math.ceil((scrollTop + (viewH || 400)) / ROW_H) + OVERSCAN)
  const slice = rows.slice(first, last)

  // 滚到底自动补下一页(全文流)。留一屏的余量:用户不会看到"滚到底才开始转圈"。
  const onScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    const el = e.currentTarget
    setScrollTop(el.scrollTop)
    if (tab === 'stream' && p.streamFrames.length < p.streamTotal
      && el.scrollTop + el.clientHeight > el.scrollHeight - ROW_H * 8) {
      p.onLoadMore()
    }
  }, [p, tab])

  const t0 = useMemo(() => frames.find((f) => f.t > 0)?.t || 0, [frames])
  const clock = (t: number) => (precision === 'frame' ? fmtClock(t) : '—')

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      {/* 横向滚动**只在这一层**:表头与表体共用同一个滚动容器,列才对得齐。
          每行**不能**自己开 overflow-x-auto —— 那会让每一行都变成一个滚动容器
          (CSS:一轴非 visible 时另一轴的计算值也变成 auto),行高 22px 而内容高 24px,
          结果每行都被裁掉一截、还能各自上下滚(实测截图里文字互相压着)。 */}
      <div ref={scrollRef} onScroll={onScroll} className="min-h-0 flex-1 overflow-auto">
        <div style={{ minWidth: MIN_W }}>
          <div className="sticky top-0 z-10 flex h-6 items-center border-b border-border bg-card px-1 text-[10px] text-muted-foreground">
            <div className="w-[92px] shrink-0">时间</div>
            <div className="shrink-0" style={{ width: miniW }} title="位置与上面条带同一根轴、同一个函数">轴</div>
            <div className="w-6 shrink-0 text-center" title="→ 运行期向前端 / ← 前端向运行期">向</div>
            <div className="w-12 shrink-0 text-right" title="语义步号(它随 --stop-at/--skip 变,不是身份)">步</div>
            <div className="w-[210px] shrink-0 pl-1" title="这一步/这一帧指向的控件。名字解不出来就显示 Type#id,绝不编">目标</div>
            <div className="min-w-0 flex-1 pl-1">文本</div>
            <div className="w-14 shrink-0 text-right">字节</div>
            <div className="w-14 shrink-0 text-right" title="从这一步到它第一条回应的时间。没有回包显示「—」—— 不是 0 毫秒">耗时</div>
          </div>

          <div style={{ height: total * ROW_H, position: 'relative' }}>
            {slice.map((row, i) => (
              <div key={row.key} style={{ position: 'absolute', top: (first + i) * ROW_H, left: 0, right: 0, height: ROW_H }}>
                <RowView row={row} p={p} xOf={xOf} miniW={miniW} clock={clock} t0={t0}
                  collapsed={row.kind === 'action' && collapsed.has(row.key)}
                  onToggle={() => {
                    if (row.kind !== 'action') return
                    setCollapsed((s) => {
                      const n = new Set(s)
                      if (n.has(row.key)) n.delete(row.key)
                      else n.add(row.key)
                      return n
                    })
                  }} />
              </div>
            ))}
          </div>
          {total === 0 && (
            <div className="p-4 text-center text-xs text-muted-foreground">
              {tab === 'steps'
                ? '这一份录制里没有语义动作(只有界面同步,没有可识别的用户操作)。'
                : '没有帧可显示。'}
            </div>
          )}
          {tab === 'stream' && p.streamFrames.length < p.streamTotal && (
            <div className="p-2 text-center text-[11px] text-muted-foreground">
              已加载 {p.streamFrames.length} / {p.streamTotal} 帧,滚动到底继续加载…
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

/** 一行的渲染。三种行共用同一套列宽,否则滚动起来会"跳列"。 */
function RowView({ row, p, xOf, miniW, clock, t0, collapsed, onToggle }: {
  row: Row
  p: Props
  xOf: (f: RecordingFrame) => number
  miniW: number
  clock: (t: number) => string
  t0: number
  collapsed: boolean
  onToggle: () => void
}) {
  const wrap = 'flex h-[22px] items-center overflow-hidden px-1 '

  if (row.kind === 'action') {
    const a = row.a
    const sel = p.selAction === a.n
    const resp0 = a.responses && a.responses[0]
    // 耗时只在"有回应且两个时刻都知道"时给。没有回包 ≠ 0 毫秒(见 D7 的纪律),所以不给 0。
    const dur = resp0 && resp0.t > 0 && a.frame.t > 0 ? resp0.t - a.frame.t : 0
    return (
      <div className={wrap + `cursor-pointer ${sel ? 'bg-sky-500/10' : 'hover:bg-accent/40'}`}
        onClick={() => p.onSelectAction(a)}
        onMouseEnter={() => a.frame.t > 0 && p.onCursor(a.frame.t)}
        title={a.frameText || ''}>
        <div className="w-[92px] shrink-0 font-mono text-[11px] text-muted-foreground">{clock(a.frame.t)}</div>
        <MiniAxis xOf={xOf} miniW={miniW} f={a.frame} cls="bg-sky-500" />
        <div className="w-6 shrink-0 text-center">{a.frame.d === 1 ? '→' : '←'}</div>
        <div className="w-12 shrink-0 text-right font-mono text-[11px]">#{a.n}</div>
        <div className="w-[210px] shrink-0 pl-1"><TargetCell a={a} /></div>
        <div className="flex min-w-0 flex-1 items-center gap-1 pl-1">
          <button onClick={(e) => { e.stopPropagation(); onToggle() }}
            className="shrink-0 rounded-ui p-0.5 text-muted-foreground hover:bg-accent"
            title={collapsed ? '展开这一步引起的回应' : '折叠回应'}>
            {collapsed ? <ChevronRight size={11} /> : <ChevronDown size={11} />}
          </button>
          {a.kind === 'input'
            ? <span className="flex items-center gap-1 truncate"><TypeIcon size={11} className="shrink-0 text-muted-foreground" />输入 <span className="font-mono">{a.value || ''}</span></span>
            : <span className="flex items-center gap-1 truncate"><MousePointerClick size={11} className="shrink-0 text-muted-foreground" />点击</span>}
          {a.dropped > 0 && (
            <span className="shrink-0 rounded-ui border border-border px-1 text-[10px] text-muted-foreground"
              title={`同一帧里还有 ${a.dropped} 个原始事件被语义层丢掉(导航/状态上报:KeyEvent、tabIndex、pageSize…)`}>
              同帧另有 {a.dropped} 个上报
            </span>
          )}
          {a.responsesState === 'attributed' ? (
            <span className="shrink-0 rounded-ui border border-border px-1 text-[10px] text-emerald-600 dark:text-emerald-400"
              title={`运行期在这一步之后回了 ${a.responsesTotal} 帧（按时间窗归属）`}>
              ← {a.responsesTotal}
            </span>
          ) : a.responsesState === 'unknown' ? (
            /* 归不了 ≠ 没回包:前者的下一步是"这一份的时间精度不够",后者是"去看回放日志"。
               两者都画成"无回包"就是在冤枉运行期。 */
            <span className="shrink-0 rounded-ui border border-dashed border-amber-500/40 px-1 text-[10px] text-amber-600 dark:text-amber-400"
              title={a.responsesNote || '这一步的回应归属不了'}>
              归不了
            </span>
          ) : (
            <span className="shrink-0 rounded-ui border border-dashed border-border px-1 text-[10px] text-muted-foreground"
              title={a.responsesNote || '这一步之后没有回包'}>
              无回包
            </span>
          )}
          {/* 回放结果徽标：只有回过放、而且这一步在回放计划里（按事件文本查到）才挂。
              没查到就什么都不显示 —— "没回过放"与"回放跳过了它"是两件事，不许混成一个灰标签。 */}
          {(() => {
            const st = p.stepByEvent && a.events.map(p.stepByEvent).find(Boolean)
            if (!st) return null
            const b = stepBadge(st)
            return (
              <span className={`shrink-0 rounded-ui border px-1 text-[10px] ${b.cls}`}
                title={`回放第 ${st.n} 步：${b.why}`}>
                {b.text}
              </span>
            )
          })()}
        </div>
        <div className="w-14 shrink-0 text-right font-mono text-[11px] text-muted-foreground">{fmtBytes(a.frame.n)}</div>
        <div className="w-14 shrink-0 text-right font-mono text-[11px] text-muted-foreground">{fmtDur(dur)}</div>
      </div>
    )
  }

  const f = row.f
  const sel = !!p.sel && frameKey(p.sel) === frameKey(f)
  return (
    <div className={wrap + `cursor-pointer text-muted-foreground ${sel ? 'bg-amber-500/10' : 'hover:bg-accent/30'}`}
      onClick={() => p.onSelectFrame(f)}
      onMouseEnter={() => f.t > 0 && p.onCursor(f.t)}
      title={row.kind === 'frame' && f.note ? f.note : undefined}>
      <div className="w-[92px] shrink-0 font-mono text-[11px]">{clock(f.t)}</div>
      <MiniAxis xOf={xOf} miniW={miniW} f={f} cls={f.d === 1 ? 'bg-emerald-500' : 'bg-sky-500'} />
      <div className="w-6 shrink-0 text-center">
        {row.kind === 'response' ? <CornerDownLeft size={11} className="inline" /> : f.d === 1 ? '→' : '←'}
      </div>
      <div className="w-12 shrink-0" />
      <div className="w-[210px] shrink-0 truncate pl-1 font-mono text-[11px]">{f.r ? `#${f.r}` : f.k}</div>
      <div className="min-w-0 flex-1 truncate pl-1 font-mono text-[11px]">
        <span className="text-foreground/70">{f.k}</span>
        {(f.q > 0 || f.k === 'event' || f.k === 'om') && <span> #{f.q}</span>}
        {f.preview && <span className="pl-1">{f.preview}</span>}
        {!f.preview && f.note && <span className="pl-1">{f.note}</span>}
      </div>
      <div className="w-14 shrink-0 text-right font-mono text-[11px]">{fmtBytes(f.n)}</div>
      <div className="w-14 shrink-0 text-right font-mono text-[11px]"
        title={f.t > 0 ? fmtRel(f.t, t0) : '这一份没有帧级时间(没有帧索引,或索引是事后重建的)'}>
        {f.t > 0 ? '' : '—'}
      </div>
    </div>
  )
}

/**
 * 表里那个回放徽标。与 ReplayProgress.tsx 的 badge 是**同一套说法**，但这里只有一格宽，
 * 所以用短词（两处的判定规则一致：`suspect` 才是黄灯的"可疑"）。
 */
function stepBadge(s: ReplayStep): { text: string; cls: string; why: string } {
  switch (s.result) {
    case 'sent':
      return { text: '已确认', cls: 'border-emerald-500/30 text-emerald-600 dark:text-emerald-400', why: '运行期回了包' }
    case 'no-response':
      return s.suspect
        ? { text: '无回包?', cls: 'border-amber-500/40 text-amber-600 dark:text-amber-400', why: '点击类事件本该有回包，8 秒内没有任何回包（可疑）' }
        : { text: '无应答', cls: 'border-border text-muted-foreground', why: '这一步本来就没有应答' }
    case 'skipped':
      return { text: '跳过', cls: 'border-border text-muted-foreground', why: s.reason || '按要求跳过' }
    case 'failed':
      return { text: '失败', cls: 'border-red-500/30 text-red-600 dark:text-red-400', why: s.reason || '发送出错' }
    case 'waiting':
      return { text: '进行中', cls: 'border-sky-500/30 text-sky-600 dark:text-sky-400', why: '已发出，正在等回包（进行中不画时长）' }
    default:
      return { text: '已暂停', cls: 'border-border text-muted-foreground', why: '等「继续」或「单步」' }
  }
}

/** 迷你轴:行里那个小方块。位置来自**同一个** x()(与条带一致)。 */function MiniAxis({ xOf, miniW, f, cls }: { xOf: (f: RecordingFrame) => number; miniW: number; f: RecordingFrame; cls: string }) {
  const raw = xOf(f)
  const within = raw >= 0 && raw <= miniW - 2
  const left = Math.max(0, Math.min(miniW - 2, raw))
  return (
    <div className="relative shrink-0" style={{ width: miniW }}>
      <div className={`absolute top-1/2 h-2 w-[3px] -translate-y-1/2 rounded-[1px] ${cls} ${within ? '' : 'opacity-30'}`}
        style={{ left }} />
    </div>
  )
}

/** 目标那一格:名字 + 置信度。**unresolved 绝不显示名字**(显示 `Type#id` 与原因)。 */
function TargetCell({ a }: { a: RecordingAction }) {
  const t = a.target
  if (t.confidence === 'high' && t.name) {
    return (
      <span className="flex items-center gap-1 truncate"
        title={`高置信(${t.source || ''})${t.windowName ? ` · 窗口 ${t.windowName}` : ''}`}>
        <span className="truncate text-foreground">{t.name}</span>
        <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{t.type}#{t.id}</span>
      </span>
    )
  }
  if (t.confidence === 'low' && t.name) {
    return (
      <span className="flex items-center gap-1 truncate text-muted-foreground"
        title={`低置信(仅排障,不能当手册名引用):${t.why || ''}${t.source ? ` · 来源 ${t.source}` : ''}`}>
        <span className="truncate">{t.name}</span>
        <span className="shrink-0 rounded-ui border border-border px-1 text-[10px]">低置信</span>
      </span>
    )
  }
  return (
    <span className="flex items-center gap-1 truncate" title={t.why || '这个名字没解出来'}>
      <span className="truncate font-mono">{t.type || '?'}#{t.id || a.id}</span>
      <span className="shrink-0 rounded-ui border border-dashed border-border px-1 text-[10px]">未解析</span>
    </span>
  )
}
