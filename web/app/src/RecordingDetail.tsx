// 录制明细页(计划 C 的 D3/D4/D5/D6 的装配处)。
//
// 一屏三块,从上到下三件事:
//
//	概览条带    这段时间线上发生了什么(两个方向 + 缓存,一根游标)
//	三层记录表  用户做了什么(语义动作 > 原始事件 > 运行期回应)/ 线上过了什么(全文流)
//	检查器      选中的那一行到底是什么(原始字节、引起的回应、名字是怎么来的)
//
// 三条纪律的落地处:
//  · **域只有一份**:条带的 x() 与表格里迷你轴的 x() 用的是同一个 Domain(见 recordingTime.ts);
//  · **没有帧级时间就画字节序**,并把这件事写在轴标签上(不许插值出假的时刻);
//  · 读侧永远"照着服务端说的显示":名字的置信度、索引的状态、归属不了的原因,一律原样展示。
//
// 数据来自四个只读端点(见 internal/web/recordingdetail.go),每次进这一页重新拉 ——
// 录制件是不可变的,所以不存在"拉早了"的问题。
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { AlertTriangle, ArrowLeft, Check, Fingerprint, RefreshCw, RotateCcw, Trash2 } from 'lucide-react'
import {
  api, type RecordingAction, type RecordingFrame, type RecordingMeta, type RecordingOverview as Ov,
  type ReplayProgress as Progress, type ReplayStep,
} from './api'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '../../shared/ui-radix'
import { useStore } from './store'
import { useToast } from './toast'
import { copyText } from './clipboard'
import { RecordingOverview } from './RecordingOverview'
import { RecordingTable } from './RecordingTable'
import { RecordingInspector } from './RecordingInspector'
import { ReplayProgressBar } from './ReplayProgress'
import { axisValue, clampDomain, domainOf, fmtBytes, makeScale, type AxisMode, type Domain } from './recordingTime'

/** 迷你轴(表格每行里那个小方块)的宽度。 */
const MINI_W = 72
const INSP_W_KEY = 'tt.recordInspW'
const INSP_MIN = 260
const INSP_MAX = 640
/** 窄到这个宽度以下就不再左右分栏,把检查器放到底部(与录制列表页的 narrowSplit 同一个道理)。 */
const NARROW = 760
const STREAM_PAGE = 500

/** 空进度(服务端没给 progress 时的兜底：宁可显示 0，也不显示 undefined 炸掉渲染)。 */
function emptyProgress(): Progress {
  return { total: 0, done: 0, sent: 0, noResp: 0, suspect: 0, skipped: 0, failed: 0, pending: false, paused: false, running: false }
}

interface Props {
  id: string
  onBack: () => void
}

export function RecordingDetail({ id, onBack }: Props) {
  const tab = useStore((s) => s.recordTab)
  const setTab = useStore((s) => s.setRecordTab)
  const toast = useToast()

  const [ov, setOv] = useState<Ov | null>(null)
  const [meta, setMeta] = useState<RecordingMeta | null>(null)
  // metaLoaded=false 表示"还没问到" —— 与"问了,但这份没有身份文件"是两件事。
  const [metaLoaded, setMetaLoaded] = useState(false)
  const [actions, setActions] = useState<RecordingAction[]>([])
  const [loading, setLoading] = useState(false)
  const [err, setErr] = useState('')
  const [tick, setTick] = useState(0)
  const [replaying, setReplaying] = useState(false)
  const [confirmDel, setConfirmDel] = useState(false)
  const [copied, setCopied] = useState<'ok' | 'fail' | null>(null)

  // 选中:语义动作(分层)与帧(全文流/条带)各一个。点一个会清掉另一个 ——
  // 否则检查器不知道该显示谁(两个都非空时以动作为准,但那种状态本身就不该存在)。
  const [selAction, setSelAction] = useState<number | null>(null)
  const [selFrame, setSelFrame] = useState<RecordingFrame | null>(null)
  const [cursor, setCursor] = useState(0)

  // 区间深链接(`?t=`):**只读的入参**,进页面时把游标落到那一刻并放大到它周围。
  // 只认一次(它一变就把用户自己拖出来的窗口覆盖掉,那是最烦人的一种"自动纠正")。
  const recordAt = useStore((s) => s.recordAt)
  const atApplied = useRef(0)

  // 全文流的分页(服务端过滤 + 分页:自己在前端滤会让"这一页只剩两条"变成常态)。
  const [streamFrames, setStreamFrames] = useState<RecordingFrame[]>([])
  const [streamTotal, setStreamTotal] = useState(0)
  const streamPage = useRef(0)

  // ---- 回放进度(计划 C 的 D7) ----
  //
  // 两条触发：**轮询**(权威，断线也能补齐) + **WS `rec.progress`**(叫一声，听见了立刻读一次)。
  // 数据永远来自 `/replay-steps` —— 这样即使那个事件类型被漏注册，退化成"晚几百毫秒"。
  const progressTick = useStore((s) => s.recProgressTick)
  const progressID = useStore((s) => s.recProgressID)
  const [rp, setRp] = useState<{ steps: ReplayStep[]; progress: Progress; running: boolean; note?: string; hasLog: boolean } | null>(null)
  const rpSince = useRef(0)
  const readProgress = useCallback(async (watch: boolean) => {
    try {
      const r = await api.replaySteps(id, 0, watch) // 每次整份读:几十到几千行,比"补页"少一类错
      rpSince.current = r.count || 0
      setRp({
        steps: r.states || [], progress: r.progress || emptyProgress(), running: !!r.running,
        note: r.note, hasLog: !!r.hasLog,
      })
    } catch { /* 读进度失败不动界面:它只是旁证,下一拍再试 */ }
  }, [id])
  // 进页面就带上 watch=1（让服务端开始推），并每隔一会儿兜底读一次。
  useEffect(() => {
    rpSince.current = 0
    setRp(null)
    void readProgress(true)
  }, [id, readProgress, tick])
  useEffect(() => {
    if (!rp) return
    // 还在跑:1s 兜底轮询(WS 一般会更快地叫我们);结束了:停掉,别空转。
    if (!rp.running) return
    const t = window.setInterval(() => void readProgress(false), 1000)
    return () => window.clearInterval(t)
  }, [rp, readProgress])
  // WS 叫一声:只认自己这一份的进度(可能同时有别的回放在跑)。
  useEffect(() => {
    if (!progressTick) return
    if (progressID && progressID !== id) return
    void readProgress(false)
  }, [progressTick, progressID, id, readProgress])
  const onCtl = async (cmd: 'pause' | 'resume' | 'step' | 'stop') => {
    try {
      const r = await api.replayCtl(id, cmd)
      toast.success('已下达命令', { description: r.note })
      void readProgress(false)
    } catch (e: any) {
      toast.error('命令没生效', { description: e?.message || String(e) })
    }
  }
  // 事件文本 → 步（表里那些"有动作的帧"据此挂上每一步的结果徽标）。
  // 用文本当键：回放计划与界面的动作都来自同一份 `event _om` 帧，文本是逐字相同的。
  const stepByEvent = useMemo(() => {
    const m = new Map<string, ReplayStep>()
    for (const s of rp?.steps || []) m.set(s.event.trim(), s) // 同一 n 的中间态/终态已由服务端收敛
    return (ev: string) => m.get(ev.trim())
  }, [rp])

  // 拉数据:概览 + 分层动作一起(它们是同一屏的两个事实,分两次进入会看到互相矛盾的画面)。
  useEffect(() => {
    let alive = true
    setLoading(true)
    setErr('')
    setSelAction(null)
    setSelFrame(null)
    setCursor(0)
    streamPage.current = 0
    setStreamFrames([])
    setStreamTotal(0)
    Promise.all([api.recordingOverview(id), api.recordingSteps(id)])
      .then(([o, s]) => {
        if (!alive) return
        setOv(o)
        setActions(s.actions || [])
      })
      .catch((e: any) => { if (alive) { setOv(null); setErr(e?.message || String(e)) } })
      .finally(() => { if (alive) setLoading(false) })
    // 身份**单独**拉,而且失败不算整页失败:没有 `.meta.json` 的老件(或收工被硬杀的那一轮)
    // 是合法状态 —— 界面只能照实说"身份不详",不能让整页打不开(实测踩到过:
    // /api/recordings/{id} 对没有身份的件回 404,而明细页的帧数据明明拿得到)。
    api.recordingShow(id)
      .then((r) => { if (alive) setMeta(r.meta || null) })
      .catch(() => { if (alive) setMeta(null) })
      .finally(() => { if (alive) setMetaLoaded(true) })
    return () => { alive = false }
  }, [id, tick])

  const doReplay = async () => {
    setReplaying(true)
    try {
      const r = await api.recordingReplay(id)
      toast.success(`已开始回放:${r.id}`, {
        description: `后台进程 pid ${r.pid};进度看日志 ${r.log};会话留着给 AI 取证`,
        duration: 8000, // 日志路径要看得清,别 3.5 秒就没了
      })
    } catch (e: any) {
      toast.error('回放失败', { description: e?.message || String(e) })
    } finally {
      setReplaying(false)
    }
  }
  const doCopy = async () => {
    const ok = await copyText(id)
    setCopied(ok ? 'ok' : 'fail')
    if (!ok) {
      toast.error('复制失败', {
        description: `系统剪贴板不可用(多半是窗口没有焦点);这份录制的 ID 是 ${id},可以在上方选中复制。`,
        duration: 0,
      })
    }
    window.setTimeout(() => setCopied(null), ok ? 1500 : 4000)
  }
  // 复制"这一刻"的深链接(`?t=`):把条带上的游标位置做成一条能给别人的地址。
  // 只在时间轴上做 —— 字节序那种坐标是"这一份文件里的第几个字节",给别人没有意义。
  const copyMoment = async () => {
    if (!(cursor > 0) || domain.mode !== 'time') {
      toast.error('先把游标放到某一刻', { description: '在条带上点一下（或拖一下），再复制链接。' })
      return
    }
    const url = `${location.origin}${location.pathname}#record/${id}?t=${Math.round(cursor)}`
    const ok = await copyText(url)
    toast[ok ? 'success' : 'error'](ok ? '已复制这一刻的链接' : '复制失败', {
      description: ok ? `打开它会直接落到 ${new Date(Math.round(cursor)).toLocaleTimeString('zh-CN', { hour12: false })} 附近` : url,
      duration: ok ? 4000 : 0,
    })
  }
  const doDelete = async () => {
    setConfirmDel(false)
    try {
      const r = await api.recordingDelete(id)
      toast.success(`已删除录制:${r.id}`, { description: `清除 ${r.files} 个文件,释放 ${fmtBytes(r.bytes)}` })
      onBack() // 这一份没了:明细页没有可显示的东西
    } catch (e: any) {
      toast.error('删除录制失败', { description: e?.message || String(e) })
    }
  }

  // 全文流按需拉:第一次进 `stream` 页签才拉第一页(分层页签不需要它)。
  const loadMore = useCallback(() => {
    if (streamPage.current * STREAM_PAGE >= streamTotal && streamPage.current > 0) return
    const page = streamPage.current
    api.recordingTimeline(id, { from: page * STREAM_PAGE, limit: STREAM_PAGE })
      .then((r) => {
        setStreamTotal(r.total)
        setStreamFrames((prev) => (page === 0 ? r.frames : [...prev, ...r.frames]))
        streamPage.current = page + 1
      })
      .catch((e: any) => setErr(e?.message || String(e)))
  }, [id, streamTotal])
  useEffect(() => {
    if (tab === 'stream' && streamPage.current === 0) loadMore()
  }, [tab, loadMore])

  // 容器宽度:决定"右侧检查器"还是"底部检查器"。
  const rootRef = useRef<HTMLDivElement | null>(null)
  const [rootW, setRootW] = useState(0)
  useEffect(() => {
    const el = rootRef.current
    if (!el) return
    const ro = new ResizeObserver(() => setRootW(el.clientWidth))
    ro.observe(el)
    setRootW(el.clientWidth)
    return () => ro.disconnect()
  }, [])
  const narrow = rootW > 0 && rootW < NARROW
  const [inspW, setInspW] = useState(() => {
    const v = Number(localStorage.getItem(INSP_W_KEY))
    return v >= INSP_MIN && v <= INSP_MAX ? v : 380
  })
  const onResizeDown = (e: React.MouseEvent) => {
    e.preventDefault()
    const startX = e.clientX
    const start = inspW
    let latest = start
    const move = (ev: MouseEvent) => {
      latest = Math.min(INSP_MAX, Math.max(INSP_MIN, start - (ev.clientX - startX)))
      setInspW(latest)
    }
    const up = () => {
      localStorage.setItem(INSP_W_KEY, String(latest))
      window.removeEventListener('mousemove', move)
      window.removeEventListener('mouseup', up)
      window.removeEventListener('blur', up)
    }
    window.addEventListener('mousemove', move)
    window.addEventListener('mouseup', up)
    window.addEventListener('blur', up)
  }

  // Esc 返回列表。有对话框开着时不抢(与列表页同一条规矩:删除的二次确认自己在别处)。
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return
      const sel = window.getSelection()?.toString()
      if (sel) return // 正在选文本:ESC 先当"取消选择"用
      e.preventDefault()
      onBack()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onBack])

  // **唯一**的轴域:条带与表格迷你轴共用它。
  const mode: AxisMode = ov?.timePrecision === 'frame' ? 'time' : 'bytes'
  const totalBytes = useMemo(
    () => (ov?.directions || []).reduce((s, d) => s + d.bytes, 0) || 1,
    [ov])
  const bounds = useMemo<Domain>(() => {
    if (!ov) return { mode, min: 0, max: 1 }
    if (mode === 'time') {
      const ts = ov.items.filter((f) => f.t > 0).map((f) => f.t)
      if (ts.length > 0) return domainOf(ts, 'time')
      return { mode: 'bytes', min: 0, max: totalBytes }
    }
    return { mode: 'bytes', min: 0, max: totalBytes }
  }, [ov, mode, totalBytes])
  const [domain, setDomain] = useState<Domain>(bounds)
  // 换一份录制 / 轴口径变了 → 复位(带着上一份的缩放窗口去画这一份,看到的会是空白)
  useEffect(() => { setDomain(bounds) }, [bounds, id])

  const xOf = useMemo(() => {
    const x = makeScale(domain, MINI_W)
    return (f: RecordingFrame) => x(axisValue(f, domain.mode))
  }, [domain])

  const pick = (f: RecordingFrame) => {
    setSelAction(null)
    setSelFrame(f)
    setCursor(axisValue(f, domain.mode))
  }
  const pickAction = (a: RecordingAction) => {
    setSelAction(a.n)
    setSelFrame(a.frame)
    setCursor(axisValue(a.frame, domain.mode))
  }

  const firstT = ov?.items.find((f) => f.t > 0)?.t || 0
  const counts = ov?.counts
  const title = ov ? `${ov.id}` : id

  // 深链接带来的那一刻:等数据到手再落位(那时才知道时间域)。
  useEffect(() => {
    if (!recordAt || !ov || atApplied.current === recordAt) return
    atApplied.current = recordAt
    setCursor(recordAt)
    if (bounds.mode === 'time' && bounds.max > bounds.min) {
      // 放大到"那一刻周围约 5% 宽度"的窗口：既能看清它，也不至于只剩一两帧。
      const pad = Math.max(50, (bounds.max - bounds.min) * 0.025)
      const next = clampDomain({ mode: 'time', min: recordAt - pad, max: recordAt + pad }, bounds)
      setDomain(next)
    }
  }, [recordAt, ov, bounds])

  return (
    <div ref={rootRef} className="flex h-full min-h-0 min-w-0 flex-col overflow-hidden bg-background">
      {/* 明细页的头:**返回**在最左(它是这一屏唯一的出口),右边是"这是哪一份"的关键几项。 */}
      <div className="flex h-8 shrink-0 items-center gap-2 overflow-x-auto whitespace-nowrap border-b border-border px-2 text-[11px]">
        <button onClick={onBack} className="flex shrink-0 items-center gap-0.5 rounded-ui border border-border px-1.5 py-0.5 hover:bg-accent"
          title="返回列表（Esc）">
          <ArrowLeft size={12} /> 列表
        </button>
        <span className="shrink-0 font-mono">{title}</span>
        {ov?.prog && <span className="shrink-0 text-sky-600 dark:text-sky-400">{ov.prog}</span>}
        {ov && (
          <>
            <span className="shrink-0 text-muted-foreground">{ov.where}</span>
            {counts && (
              <span className="shrink-0 text-muted-foreground"
                title="三层各是一层事实：帧数（含图片缓存）/ 事件侧帧数 → 原始事件数 / 语义动作数">
                {counts.frames} 帧 · 事件侧 {counts.eventFrames} 帧 → {counts.events} 个事件 · {counts.actions} 步
              </span>
            )}
            <span className="shrink-0 rounded-ui border border-border px-1.5 py-0.5"
              title={ov.timePrecision === 'frame'
                ? '每帧都有首块时刻(来自帧索引)'
                : '这一份没有帧级时间:横轴是各方向的字节序,顺序不等于时间'}>
              {ov.timePrecision === 'frame' ? '时间:帧级' : '时间:块级'}
            </span>
          </>
        )}
        {/* 看着这一份做决定的三个动作。它们原来在列表页的侧栏里 —— 搬到这里是因为
            "一份录制只该有一个我正看着它的地方",而这三个动作都得看着它才做得对。
            回放:r.r 录的件直接禁用(编号表与调试线不是同一张,重放会点错控件)。 */}
        <button onClick={() => void doReplay()} disabled={replaying || meta?.path === 'r.r'}
          title={meta?.path === 'r.r'
            ? '这份是 r.r 录的:编号表与调试线不是同一张,重放会点错控件(要回放请用「开始录制」录的那份)'
            : '把这份录制在调试器下重放出来:服务端另起一个进程,按身份起调试线会话、过指纹闸、把事件原样发出去(不等停站);会话留着给 AI 取证,进度看日志'}
          className="ml-auto inline-flex shrink-0 items-center gap-1 rounded-ui border border-emerald-500/20 px-1.5 py-0.5 text-emerald-600 hover:bg-emerald-500/10 disabled:pointer-events-none disabled:opacity-40 dark:text-emerald-400">
          <RotateCcw size={11} />
          {replaying ? '正在起…' : '回放'}
        </button>
        <button onClick={() => void doCopy()}
          title="复制这份录制的 ID:把它发给 AI,AI 就能用 tt fep rec reproduce <ID> 重放这一份"
          className={`inline-flex shrink-0 items-center gap-1 rounded-ui border px-1.5 py-0.5 ${
            copied === 'fail' ? 'border-red-500/20 text-red-600 dark:text-red-400' : 'border-border text-muted-foreground hover:bg-accent'}`}>
          {copied === 'ok' ? <Check size={11} /> : <Fingerprint size={11} />}
          {copied === 'ok' ? '已复制' : copied === 'fail' ? '复制失败' : '复制 ID'}
        </button>
        {/* 删除只对长期库里的那些开放:缓存里的那份归统一清理框架管(服务端会拒),摆在这里只会得到一个 409。 */}
        {ov?.where === '长期库' && (
          <button onClick={() => setConfirmDel(true)}
            title="从长期库里删掉这一份。录制不可再生,会先弹二次确认"
            className="inline-flex shrink-0 items-center gap-1 rounded-ui border border-red-500/20 px-1.5 py-0.5 text-red-600 hover:bg-red-500/10 dark:text-red-400">
            <Trash2 size={11} /> 删除
          </button>
        )}
        <button onClick={() => setTick((t) => t + 1)} className="flex shrink-0 items-center gap-0.5 rounded-ui border border-border px-1.5 py-0.5 hover:bg-accent">
          <RefreshCw size={11} className={loading ? 'animate-spin' : ''} /> 刷新
        </button>
      </div>

      {err && (
        <div className="flex shrink-0 items-start gap-2 border-b border-border bg-red-500/10 px-2 py-1 text-[11px]">
          <AlertTriangle size={12} className="mt-0.5 shrink-0 text-red-600 dark:text-red-400" />
          <div className="min-w-0 flex-1 whitespace-pre-wrap break-words">{err}</div>
          <button onClick={() => setTick((t) => t + 1)} className="shrink-0 rounded-ui border border-border px-1.5 py-0.5 hover:bg-accent">重试</button>
        </div>
      )}

      {loading && !ov && <div className="p-4 text-center text-xs text-muted-foreground">读取录制中…</div>}

      {ov && (
        <>
          {/* 回放进度：只有回过放才出现（没回过放 = 这一块整块收起来，不是显示一条空进度）。 */}
          {rp?.hasLog && (
            <ReplayProgressBar
              id={id} progress={rp.progress} steps={rp.steps} running={rp.running} note={rp.note}
              onCtl={(c) => void onCtl(c)} onRefresh={() => void readProgress(false)} />
          )}
          <RecordingOverview
            items={ov.items} blocks={ov.blocks} mode={bounds.mode}
            bounds={bounds} domain={domain} onDomain={setDomain}
            cursor={cursor} onCursor={setCursor}
            sel={selFrame} onSelect={pick}
            onCopyLink={() => void copyMoment()} />

          <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden" style={{ flexDirection: narrow ? 'column' : 'row' }}>
            <div className="flex min-h-0 min-w-0 flex-1 flex-col">
              {/* 页签:两个。计数只给**一个数** —— 头里已经有"帧/事件/步"三层的全貌，
                  这里只需要说明"这一页有多少行"。哪一层是什么，页签名与行本身说得清。 */}
              <div className="flex h-7 shrink-0 items-center gap-1 border-b border-border px-1 text-[11px]">
                {([['steps', '分层记录'], ['stream', '全文流']] as const).map(([k, label]) => (
                  <button key={k} onClick={() => setTab(k)}
                    className={`rounded-ui px-2 py-0.5 ${tab === k ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent/50'}`}
                    title={k === 'steps'
                      ? '一个语义动作一行：回答"用户做了什么、运行期回没回"。同一行里可展开看回应'
                      : '一帧一行（含图片缓存）：回答"这条线上到底过了什么字节"'}>
                    {label}
                  </button>
                ))}
                <span className="ml-2 truncate text-[10px] text-muted-foreground">
                  {tab === 'steps'
                    ? `共 ${actions.length} 步`
                    : (streamFrames.length < (streamTotal || ov.items.length)
                      ? `已加载 ${streamFrames.length} / ${streamTotal || ov.items.length} 帧`
                      : `共 ${streamTotal || ov.items.length} 帧`)}
                </span>
              </div>
              <RecordingTable
                key={id}
                tab={tab} actions={actions} frames={ov.items} precision={ov.timePrecision}
                xOf={xOf} miniW={MINI_W} cursor={cursor} onCursor={setCursor}
                sel={selFrame} onSelectFrame={pick}
                selAction={selAction} onSelectAction={pickAction}
                stepByEvent={stepByEvent}
                streamFrames={tab === 'stream' ? (streamFrames.length ? streamFrames : ov.items) : []}
                streamTotal={tab === 'stream' ? (streamTotal || ov.items.length) : 0}
                onLoadMore={loadMore} />
            </div>

            {!narrow && <div className="col-resizer self-stretch" onMouseDown={onResizeDown} title="拖拽分配宽度" />}

            <div className={narrow
              ? 'h-56 shrink-0 overflow-hidden border-t border-border'
              : 'flex shrink-0 flex-col overflow-hidden border-l border-border'}
              style={{ width: narrow ? undefined : inspW }}>
              <RecordingInspector
                id={id} sel={selFrame} action={actions.find((a) => a.n === selAction) || null}
                precision={ov.timePrecision} counts={ov.counts} notes={ov.notes}
                directions={ov.directions} firstT={firstT} meta={meta} metaLoaded={metaLoaded} />
            </div>
          </div>
        </>
      )}

      {/* 删除的二次确认(与「会话」面板同一个基元:shadcn AlertDialog)。
          录制不可再生 —— 要再拿到同一份,得重现一次问题再录一遍。 */}
      <AlertDialog open={confirmDel} onOpenChange={(o) => { if (!o) setConfirmDel(false) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>删除录制「{id}」?</AlertDialogTitle>
            <AlertDialogDescription>
              将从长期库里删掉这一份({meta?.prog || '未知作业'},约 {fmtBytes((ov?.directions || []).reduce((s, d) => s + d.bytes, 0))})。
              录制不可再生 —— 要再拿到同一份,得重现一次问题再录一遍;此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => void doDelete()}>确定删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
