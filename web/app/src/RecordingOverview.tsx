// 录制的**概览条带**(计划 C 的 D4):把一份录制在时间轴(或字节轴)上摊平。
//
// 三条泳道,一条轴,一根游标:
//
//	运行期 → 前端   界面状态与回应(`om`/`un`/`rn`)
//	前端 → 运行期   用户/注入的事件(`event`)
//	缓存 / 其它     图片缓存(type=5)与没见过的类型 —— 它们不是消息,但**占时间**
//
// 为什么值得单独一块:用户问的是"播放过账中我可以观察走到哪了"。两个方向的录制件是**两条
// 独立的字节流**(各自一个文件、序号分属两侧),只有时间能把它们排到一起 —— 所以条带是这一页
// 唯一能回答"这条事件引出的是哪条推送"的东西(见 internal/fep/proxy.go 的 timelineLine 注释:
// 当时正是卡在这一步)。
//
// 三条纪律(都写进了界面上,不藏在注释里):
//  1. **没有帧级时间就不画时间轴**:老件的横轴是"字节序",轴标签直接写出来(见 recordingTime.ts);
//  2. **缩放窗口外的东西压暗贴边**,不许画到框外(否则"放大后那块去哪了"没人说得清);
//  3. 悬停 **500 ms** 才出精确时刻与大小(照 DSH 的做法):鼠标扫过时不该满屏 tooltip。
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link2, Maximize2, ZoomIn, ZoomOut } from 'lucide-react'
import type { RecordingBlock, RecordingFrame } from './api'
import {
  axisValue, clampDomain, domainOf, fmtBytes, fmtClock, fmtRel, makeScale, panDomain, ticks, zoomDomain,
  type AxisMode, type Domain,
} from './recordingTime'

/** 泳道:按 kind/方向分。顺序就是显示顺序(上=运行期,下=缓存)。 */
const LANES: { key: string; label: string; hint: string; match: (f: RecordingFrame) => boolean }[] = [
  {
    key: 'rt', label: '运行期 → 前端', hint: '界面状态与回应(om / un / rn):界面变成什么样,都在这条线上',
    match: (f) => f.d === 1 && !isOther(f),
  },
  {
    key: 'fe', label: '前端 → 运行期', hint: '用户/注入的事件(event _om):"谁点了什么"在这条线上',
    match: (f) => f.d === 0 && !isOther(f),
  },
  {
    key: 'other', label: '缓存 / 其它', hint: '图片缓存(type=5)与没认出来的帧:不是消息、不解析、不重放,但占时间',
    match: (f) => isOther(f),
  },
]

/** 不是消息的帧:图片缓存与没见过的类型。它们单独一条泳道(见 index.go 的规矩 3)。 */
function isOther(f: RecordingFrame): boolean {
  return f.k === 'cache' || f.k.startsWith('type')
}

/** 帧的颜色:按消息类别分。`event` 是"人做的",`om/un/rn` 是"运行期回的",一眼要能分开。 */
const KIND_CLASS: Record<string, string> = {
  event: 'bg-sky-500',
  om: 'bg-emerald-500',
  un: 'bg-emerald-600',
  rn: 'bg-rose-500',
  meta: 'bg-violet-500',
  cache: 'bg-muted-foreground/40',
}
const OTHER_CLASS = 'bg-muted-foreground/60'

function kindClass(k: string): string {
  if (k === 'cache') return KIND_CLASS.cache
  if (k.startsWith('type')) return OTHER_CLASS
  return KIND_CLASS[k] || OTHER_CLASS
}

const LANE_H = 26 // 每条泳道的高度(px)
const RULER_H = 18
/** 悬停多久才出精确信息(毫秒)。照 DSH 的做法:鼠标扫过时不弹。 */
const HOVER_DELAY = 500
/** 静默带的判据:两个方向都没有帧、且间隔超过这个比例(相对整段跨度)。 */
const SILENT_FRAC = 0.02
const SILENT_MIN_MS = 300

export interface StripSelection {
  c: number
  d: number
  o: number
}

interface Props {
  items: RecordingFrame[]
  blocks?: RecordingBlock[]
  mode: AxisMode
  /** 数据的全览域(边界)。缩放/平移都夹在它里面。 */
  bounds: Domain
  /** 当前可视域。**由父级持有** —— 表格里那个迷你轴用的是同一个 x(),两处必须共用一份。 */
  domain: Domain
  onDomain: (d: Domain) => void
  cursor: number // 游标(轴上的取值;time 模式是毫秒,bytes 模式是字节)
  onCursor: (v: number) => void
  /** 选中的那一帧(条带要用它的轴取值画选择线,不能只给 key —— 那还得反查一次) */
  sel: RecordingFrame | null
  onSelect: (f: RecordingFrame) => void
  /** 复制"这一刻"的深链接(由父级拼 `#record/<id>?t=` —— 它才知道录制 ID 与光标模型)。 */
  onCopyLink?: () => void
}

export function RecordingOverview(p: Props) {
  const boxRef = useRef<HTMLDivElement | null>(null)
  const [boxW, setBoxW] = useState(0)
  useEffect(() => {
    const el = boxRef.current
    if (!el) return
    const ro = new ResizeObserver(() => setBoxW(el.clientWidth))
    ro.observe(el)
    setBoxW(el.clientWidth)
    return () => ro.disconnect()
  }, [])

  const dom = p.domain
  const bounds = p.bounds
  const setDom = (updater: Domain | ((d: Domain) => Domain)) => {
    p.onDomain(typeof updater === 'function' ? updater(dom) : updater)
  }

  const plotW = Math.max(0, boxW - 68) // 左侧留出泳道标签
  const x = useMemo(() => makeScale(dom, plotW), [dom, plotW])
  const ticksOut = useMemo(() => ticks(dom, plotW, p.items.find((f) => f.t > 0)?.t || 0, 96), [dom, plotW, p.items])
  const zoomed = dom.min !== bounds.min || dom.max !== bounds.max

  // 悬停:等 500ms 才出信息。mouseleave 要清掉定时器,否则"扫过去一下"也会在 500ms 后弹出来。
  const [hover, setHover] = useState<{ f: RecordingFrame; x: number } | null>(null)
  const hoverTimer = useRef<number | null>(null)
  const clearHover = useCallback(() => {
    if (hoverTimer.current !== null) { window.clearTimeout(hoverTimer.current); hoverTimer.current = null }
    setHover(null)
  }, [])
  const armHover = useCallback((f: RecordingFrame, left: number) => {
    if (hoverTimer.current !== null) window.clearTimeout(hoverTimer.current)
    hoverTimer.current = window.setTimeout(() => setHover({ f, x: left }), HOVER_DELAY)
  }, [])
  useEffect(() => clearHover, [clearHover])

  const valueAt = useCallback((clientX: number) => {
    const el = boxRef.current
    if (!el || plotW <= 0) return dom.min
    const r = el.getBoundingClientRect()
    const frac = Math.min(1, Math.max(0, (clientX - r.left - 68) / plotW))
    return dom.min + frac * (dom.max - dom.min)
  }, [dom, plotW])

  // 滚轮缩放:锚在鼠标下的那个取值上(锚在最左端的话,想放大的位置会跑出视野)。
  // 必须用非 passive 的监听器:React 的 onWheel 是 passive 的,preventDefault 无效。
  useEffect(() => {
    const el = boxRef.current
    if (!el) return
    const onWheel = (e: WheelEvent) => {
      if (plotW <= 0) return
      e.preventDefault()
      const anchor = valueAt(e.clientX)
      p.onDomain(zoomDomain(dom, bounds, e.deltaY > 0 ? 1.15 : 0.87, anchor))
    }
    el.addEventListener('wheel', onWheel, { passive: false })
    return () => el.removeEventListener('wheel', onWheel)
  }, [bounds, dom, plotW, valueAt, p])

  // 左键拖动 = 移游标;右键拖动 = 平移(放大之后才用得上);右键单击 = 复位缩放 + 清选区。
  const drag = useRef<{ kind: 'cursor' | 'pan'; fromX: number; dom: Domain } | null>(null)
  const onDown = (e: React.MouseEvent) => {
    if (e.button === 2) { drag.current = { kind: 'pan', fromX: e.clientX, dom }; return }
    if (e.button !== 0) return
    drag.current = { kind: 'cursor', fromX: e.clientX, dom }
    p.onCursor(valueAt(e.clientX))
  }
  useEffect(() => {
    const move = (e: MouseEvent) => {
      const d = drag.current
      if (!d) return
      if (d.kind === 'cursor') { p.onCursor(valueAt(e.clientX)); return }
      if (plotW <= 0) return
      // 平移量按**按下那一刻**的窗口算:跟着当前窗口算会让拖拽自我加速(实测那种"越拖越快")。
      const frac = -(e.clientX - d.fromX) / plotW
      p.onDomain(panDomain(d.dom, bounds, frac))
    }
    const up = () => { drag.current = null }
    window.addEventListener('mousemove', move)
    window.addEventListener('mouseup', up)
    window.addEventListener('blur', up)
    return () => {
      window.removeEventListener('mousemove', move)
      window.removeEventListener('mouseup', up)
      window.removeEventListener('blur', up)
    }
  }, [bounds, p, plotW, valueAt])

  // 静默带:两个方向都没有帧的区间。只在时间轴上有意义(字节轴上"间隔"不存在)。
  const silents = useMemo(() => {
    if (p.mode !== 'time') return []
    const ts = p.items.filter((f) => f.t > 0).map((f) => f.t).sort((a, b) => a - b)
    if (ts.length < 2) return []
    const minGap = Math.max(SILENT_MIN_MS, (ts[ts.length - 1] - ts[0]) * SILENT_FRAC)
    const out: { from: number; to: number }[] = []
    for (let i = 1; i < ts.length; i++) {
      if (ts[i] - ts[i - 1] >= minGap) out.push({ from: ts[i - 1], to: ts[i] })
    }
    return out
  }, [p.items, p.mode])

  const selX = p.sel ? x(axisValue(p.sel, dom.mode)) : null
  const cursorX = p.cursor > 0 || dom.mode === 'bytes' ? x(p.cursor) : null

  return (
    <div className="border-b border-border bg-card/40">
      {/* 工具条:轴口径 + 缩放。轴口径要**写出来** —— "横轴是时间还是字节序"是这一屏的前提。 */}
      <div className="flex h-7 items-center gap-2 px-2 text-[11px] text-muted-foreground">
        <span className="text-foreground/80">概览</span>
        <span className="rounded-ui border border-border px-1.5 py-0.5"
          title={p.mode === 'time'
            ? '横轴 = 帧首块时刻(帧级精度)。可以滚轮缩放、右键拖动平移、右键单击复位'
            : '横轴 = 各方向文件里的**字节序** —— 这一份没有帧级时间(索引缺失或由老件重建),所以位置不是时间'}>
          {p.mode === 'time' ? '横轴:时刻' : '横轴:字节序(无帧级时间)'}
        </span>
        <span>{p.items.length} 帧</span>
        {zoomed && (
          <button onClick={() => p.onDomain(bounds)} className="flex items-center gap-0.5 rounded-ui border border-border px-1.5 py-0.5 hover:bg-accent"
            title="复位到全览(也可以右键单击条带)">
            <Maximize2 size={11} /> 全览
          </button>
        )}
        <span className="ml-auto flex items-center gap-1">
          {/* 把游标所在的那一刻做成一条能给别人的地址(区间深链接,见 routing.ts 的 recordHash)。 */}
          {p.onCopyLink && p.mode === 'time' && (
            <button onClick={p.onCopyLink} className="flex items-center gap-0.5 rounded-ui border border-border px-1.5 py-0.5 hover:bg-accent"
              title="复制这一刻的深链接（#record/<ID>?t=…）：打开它会直接落到游标所在的位置">
              <Link2 size={11} /> 这一刻
            </button>
          )}
          <button onClick={() => p.onDomain(zoomDomain(dom, bounds, 0.7, (dom.min + dom.max) / 2))}
            className="rounded-ui border border-border p-0.5 hover:bg-accent" title="放大(滚轮也可以)"><ZoomIn size={11} /></button>
          <button onClick={() => p.onDomain(zoomDomain(dom, bounds, 1.4, (dom.min + dom.max) / 2))}
            className="rounded-ui border border-border p-0.5 hover:bg-accent" title="缩小(滚轮也可以)"><ZoomOut size={11} /></button>
        </span>
      </div>

      <div ref={boxRef} className="relative select-none pb-1"
        onMouseDown={onDown}
        onContextMenu={(e) => {
          // 右键:单击 = 清选区 + 复位缩放(拖动平移由 mousedown/mousemove 处理)
          e.preventDefault()
          if (Math.abs(e.clientX - (drag.current?.fromX ?? e.clientX)) < 3) p.onDomain(bounds)
        }}>
        {/* 时间尺 */}
        <div className="relative ml-[68px] h-[18px] border-b border-border/60">
          {ticksOut.map((t) => (
            <span key={t.v} className="absolute -translate-x-1/2 whitespace-nowrap text-[10px] text-muted-foreground" style={{ left: t.x }}>
              {t.label}
            </span>
          ))}
        </div>

        {LANES.map((lane) => {
          const frames = p.items.filter(lane.match)
          return (
            <div key={lane.key} className="flex items-stretch" style={{ height: LANE_H }}>
              <div className="w-[68px] shrink-0 overflow-hidden pr-1 text-right text-[10px] leading-[26px] text-muted-foreground"
                title={lane.hint}>
                {lane.label}
              </div>
              <div className="relative min-w-0 flex-1 border-b border-border/40">
                {/* 静默带:通高斜纹 + 一句"多久没有字节"。它比空白更有信息量 —— 
                    空白既可能是"没流量",也可能是我没画出来。 */}
                {silents.map((s) => {
                  const a = x(s.from), b = x(s.to)
                  if (b < 0 || a > plotW) return null
                  const left = Math.max(0, a), w = Math.max(1, Math.min(plotW, b) - left)
                  return (
                    <div key={s.from} className="absolute inset-y-0 border-x border-dashed border-border/70 text-muted-foreground/30"
                      style={{
                        left, width: w,
                        backgroundImage: 'repeating-linear-gradient(45deg, currentColor 0 1px, transparent 1px 5px)',
                      }}
                      title={`${((s.to - s.from) / 1000).toFixed(1)} s 静默（两个方向都没有字节）`} />
                  )
                })}
                {frames.map((f) => {
                  const v = axisValue(f, dom.mode)
                  let cx = x(v)
                  let dim = false
                  // 窗口外:压暗 + 贴边。不许画到框外 —— 否则"放大后它去哪了"说不清。
                  if (cx < 0) { cx = 0; dim = true } else if (cx > plotW - 2) { cx = plotW - 2; dim = true }
                  const isSel = !!p.sel && p.sel.c === f.c && p.sel.d === f.d && p.sel.o === f.o
                  return (
                    <div key={`${f.c}:${f.d}:${f.o}`}
                      className={`absolute top-1 rounded-[1px] ${kindClass(f.k)} ${dim ? 'opacity-30' : ''} ${isSel ? 'ring-2 ring-sky-400 ring-offset-0' : ''}`}
                      style={{ left: cx, width: 3, height: LANE_H - 10 }}
                      onMouseEnter={() => armHover(f, cx)}
                      onMouseLeave={clearHover}
                      onClick={(e) => { e.stopPropagation(); p.onSelect(f) }} />
                  )
                })}
              </div>
            </div>
          )
        })}

        {/* 游标:全条带**一根**(不搞每条泳道一根)。与表格行双向联动,位置用同一个 x()。 */}
        {cursorX !== null && cursorX >= 0 && cursorX <= plotW && (
          <div className="pointer-events-none absolute bottom-0 top-[18px]" style={{ left: 68 + cursorX }}>
            <div className="h-full w-px bg-sky-500" />
          </div>
        )}
        {/* 选中帧:一根更细的竖线,颜色与"游标"分开(它们是两个概念:游标=鼠标,选中=点过) */}
        {selX !== null && selX >= 0 && selX <= plotW && (
          <div className="pointer-events-none absolute bottom-0 top-[18px]" style={{ left: 68 + selX }}>
            <div className="h-full w-px bg-amber-500" />
          </div>
        )}

        {hover && (
          <div className="pointer-events-none absolute z-20 max-w-[420px] rounded-ui border border-border bg-popover px-2 py-1 text-[11px] shadow-md"
            style={{ left: Math.min(68 + hover.x + 6, Math.max(0, boxW - 300)), top: 2 }}>
            <div className="font-mono">
              {hover.f.k} #{hover.f.q || 0} · {hover.f.d === 1 ? '运行期→前端' : '前端→运行期'} · 连接 {hover.f.c}
            </div>
            <div className="font-mono text-muted-foreground">
              {p.mode === 'time' ? `${fmtClock(hover.f.t)}（${fmtRel(hover.f.t, p.items.find((f) => f.t > 0)?.t || 0)}）` : '时刻未知（无帧索引）'}
            </div>
            <div className="font-mono text-muted-foreground">
              {fmtBytes(hover.f.n)} @ 偏移 {hover.f.o}
              {hover.f.r ? ` · 指向 #${hover.f.r}` : ''}
            </div>
          </div>
        )}
      </div>

      {/* 图例:颜色与"一根游标"的含义。不写图例的话,三色块就是装饰。 */}
      <div className="flex flex-wrap items-center gap-3 px-2 pb-1 text-[10px] text-muted-foreground">
        {[['event', '事件'], ['om', 'om'], ['un', 'un'], ['rn', 'rn'], ['meta', '握手'], ['cache', '图片缓存']].map(([k, label]) => (
          <span key={k} className="flex items-center gap-1">
            <span className={`inline-block h-2 w-2 rounded-[1px] ${kindClass(k)}`} />{label}
          </span>
        ))}
        <span className="ml-auto">
          {p.mode === 'time'
            ? '左键拖动=移游标 · 滚轮=缩放 · 右键拖动=平移 · 右键单击=复位'
            : '左键拖动=移游标（横轴是字节序,不能缩放成时间）'}
        </span>
      </div>
    </div>
  )
}
