// FEP 监控（会话页第二个页签）。
//
// 它回答一件事：**这条会话的画面通道上，此刻过了什么**。所以它只放两样东西：
// 一条三泳道的时间轴（服务 / 前端 / 其它）+ 一个事件列表 —— 没有起停按钮、没有
// 录制 ID/端口/GDC 那些状态，也没有大段说明。那些属于「环境页 → FEP录像」那套录制工作台。
//
// 数据是**追加读**的：按 `from=<已看帧数>` 每秒问一次 `/api/recordings/{id}/timeline`，
// 只把新帧接在列表末尾（不重排、不补历史）。录制文件本来就是追加写的，所以这个读法
// 天然跟得上"正在录"。
import { useEffect, useRef, useState } from 'react'
import { MonitorPlay } from 'lucide-react'
import { api, type RecordPrecision, type RecordingFrame } from './api'
import { useStore } from './store'
import { RecordingTable } from './RecordingTable'

/** 画面通道的只读视图，与后端 `GET /api/sessions/{id}/screen` 的 `screen` 块同形。 */
export interface ScreenInfo {
  tag?: string
  kind?: string
  prog?: string
  gdcPid?: number
  gdcPort?: number
  proxyPort?: number
  injectPort?: number
  recordDir?: string
  stale?: boolean
  started?: string
}

/**
 * 读这条会话挂着的画面通道。
 *
 * 两个来源：服务端的 screen 事件（挂了/收了立刻反映）+ `GET /sessions/{id}/screen`
 * （进页面补一次；事件是"叫一声"，权威在查询）。取不到就显示成"没有"，
 * 真实的报错留在 `err` 里 —— 两者不能混成一句话。
 */
export function useScreenInfo(id: string) {
  const [state, setState] = useState<{ attached: boolean; screen: ScreenInfo | null; err: string }>(
    { attached: false, screen: null, err: '' },
  )
  const [nonce, setNonce] = useState(0)
  const live = useStore((s) => s.screens[id])

  useEffect(() => {
    let cancelled = false
    if (!id) return
    setState((s) => ({ ...s, err: '' }))
    void api.screen(id).then((r) => {
      if (cancelled) return
      setState({ attached: !!r.attached, screen: (r.screen as ScreenInfo) || null, err: '' })
    }).catch((e: any) => {
      if (cancelled) return
      setState({ attached: false, screen: null, err: e?.message || String(e) })
    })
    return () => { cancelled = true }
  }, [id, nonce])

  // 事件比查询新：挂了/收了就当场改（保留查询拿到的端口、目录等细节）。
  useEffect(() => {
    if (!live) return
    setState((s) => ({
      attached: live.attached,
      screen: live.attached
        ? { ...(s.screen || {}), tag: live.tag || s.screen?.tag, kind: live.kind || s.screen?.kind }
        : null,
      err: '',
    }))
  }, [live?.attached, live?.tag, live?.kind])

  return { ...state, refresh: () => setNonce((n) => n + 1) }
}

export function FepMonitor({ sessionId }: { sessionId: string }) {
  const screen = useScreenInfo(sessionId)
  const tag = screen.attached ? (screen.screen?.tag || '') : ''
  const [frames, setFrames] = useState<RecordingFrame[]>([])
  const [total, setTotal] = useState(0)
  const [precision, setPrecision] = useState<RecordPrecision>('frame')
  const [err, setErr] = useState('')
  // 还没有报文（作业停在入口 / 还没开跑）：这时读录制会 404（raw 文件尚未产生），
  // 那不是错误，是"这一轮还没开始过东西"——照实说，别拿红字吓人。
  const [waiting, setWaiting] = useState(false)
  const [sel, setSel] = useState<RecordingFrame | null>(null)
  const [cursor, setCursor] = useState(0)
  const seenRef = useRef(0)
  const tagRef = useRef('')

  useEffect(() => {
    let cancelled = false
    if (!tag) {
      setFrames([]); setTotal(0); setErr(''); seenRef.current = 0; tagRef.current = ''
      return
    }
    if (tagRef.current !== tag) {
      // 换成另一份录制：从头开始读（不保留上一轮的行）。
      tagRef.current = tag
      seenRef.current = 0
      setFrames([]); setTotal(0); setErr('')
    }
    const tick = async () => {
      try {
        const r = await api.recordingTimeline(tag, { from: seenRef.current, limit: 500 })
        if (cancelled) return
        setTotal(r.total || 0)
        setPrecision((r.timePrecision || 'frame') as RecordPrecision)
        setErr('')
        setWaiting(false)
        const got = (r.frames || []) as RecordingFrame[]
        if (!got.length) return
        seenRef.current += got.length
        // 留最近 5000 帧：监控是"看现在"，不是"翻历史"（历史在录制明细页）。
        setFrames((prev) => prev.concat(got).slice(-5000))
      } catch (e: any) {
        const msg = String(e?.message || e)
        if (cancelled) return
        if (/找不到|NOT_FOUND|no such|没有 .* 的录制件/i.test(msg)) {
          setWaiting(true); setErr('')
        } else {
          setErr(msg)
        }
      }
    }
    void tick()
    const h = window.setInterval(tick, 1000)
    return () => { cancelled = true; clearInterval(h) }
  }, [tag])

  if (!screen.attached) {
    return (
      <div className="flex min-h-0 flex-1 items-center justify-center p-8">
        <div className="max-w-[420px] text-center text-xs leading-relaxed text-muted-foreground">
          <MonitorPlay className="mx-auto mb-2 h-5 w-5 opacity-60" />
          这条会话现在没有画面通道。
          <div className="mt-1 opacity-80">建会话时会自动同时开启监控；没有的话看这条会话是不是刚收掉过。</div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      <LaneStrip frames={frames} precision={precision} />
      {waiting && frames.length === 0 ? (
        <div className="px-4 py-3 text-[11px] leading-relaxed text-muted-foreground">
          这一轮还没有报文：作业停在入口（或还没开跑）。点「继续」放行之后，这一屏就开始滚动。
        </div>
      ) : err ? (
        <div className="px-4 py-2 text-[11px] text-muted-foreground">还读不到这一轮：{err}</div>
      ) : (
      <RecordingTable
        frames={frames}
        precision={precision}
        sel={sel}
        onSelect={(f) => { setSel(f); setCursor(f.t > 0 ? f.t : cursor) }}
        cursor={cursor}
        onCursor={setCursor}
        total={Math.max(total, frames.length)}
        onLoadMore={() => { /* 监控是跟随，不翻页 */ }}
      />
      )}
    </div>
  )
}

/** 三泳道时间轴：服务 / 前端 / 其它（图一里那三行）。 */
function LaneStrip({ frames, precision }: { frames: RecordingFrame[]; precision: RecordPrecision }) {
  const timed = precision === 'frame' && frames.some((f) => f.t > 0)
  const t0 = frames.find((f) => f.t > 0)?.t || 0
  const t1 = [...frames].reverse().find((f) => f.t > 0)?.t || t0
  const span = Math.max(1, t1 - t0)
  const xOf = (f: RecordingFrame, i: number) => {
    if (timed && f.t > 0) return Math.min(1, Math.max(0, (f.t - t0) / span))
    return frames.length > 1 ? i / (frames.length - 1) : 0
  }
  const isOther = (f: RecordingFrame) => f.k === 'cache' || f.k === 'meta' || f.k === 'type'
  const lanes: { label: string; pick: (f: RecordingFrame) => boolean; cls: string }[] = [
    { label: '服务', pick: (f) => !isOther(f) && f.d === 1, cls: 'bg-sky-500' },
    { label: '前端', pick: (f) => !isOther(f) && f.d === 0, cls: 'bg-emerald-500' },
    { label: '其它', pick: isOther, cls: 'bg-muted-foreground/50' },
  ]
  // 刻度：起点 + 每秒一个（最多 12 个，免得宽屏幕上挤成一片）。
  const markN = Math.min(12, Math.max(2, Math.round(span / 2000)))
  const marks = timed && span > 0
    ? Array.from({ length: markN }, (_, i) => ({
      left: (i / (markN - 1)) * 100,
      label: `T+${((span * i) / (markN - 1) / 1000).toFixed(1)}s`,
    }))
    : []

  return (
    <div className="shrink-0 border-b border-border bg-card px-3 pb-1 pt-2">
      {marks.length > 0 && (
        <div className="relative ml-[44px] h-4">
          {marks.map((m) => (
            <span key={m.left} className="absolute top-0 -translate-x-1/2 text-[10px] text-muted-foreground/80" style={{ left: `${m.left}%` }}>
              {m.label}
            </span>
          ))}
        </div>
      )}
      {lanes.map((lane) => (
        <div key={lane.label} className="flex items-center gap-2">
          <span className="w-10 shrink-0 text-right text-[10px] text-muted-foreground">{lane.label}</span>
          <div className="relative h-3 flex-1">
            {frames.map((f, i) => (lane.pick(f) ? (
              <span key={`${lane.label}-${f.c}-${f.d}-${f.o}`}
                className={`absolute top-0 h-3 w-[2px] opacity-80 ${lane.cls}`}
                style={{ left: `${xOf(f, i) * 100}%` }} />
            ) : null))}
          </div>
        </div>
      ))}
    </div>
  )
}
