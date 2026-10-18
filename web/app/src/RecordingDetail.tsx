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
import { RecordingTable, frameKey } from './RecordingTable'
import { RecordingInspector } from './RecordingInspector'
import { ReplayProgressBar } from './ReplayProgress'
import { axisValue, clampDomain, domainOf, fmtBytes, zoomDomain, type AxisMode, type Domain } from './recordingTime'

/** 迷你轴(表格每行里那个小方块)的宽度。 */
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
  /**
   * 这一页**改了列表会关心的事实**（删掉了一份录制、起了一轮回放）。
   *
   * 为什么要有它：列表是 keep-alive 的（只切显隐不卸载），它的重拉条件是
   * `view / sessionEnv / tick`，而这几个动作都不改它们 —— 不报一声的话，
   * 回到列表还看得见刚删掉的那一行。
   */
  onChanged?: () => void
}

export function RecordingDetail({ id, onBack, onChanged }: Props) {
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

  // 选中：只认**帧**。语义动作不再是一个独立的选中态（没有"分层"那一页了）——
  // 它由帧推出来（`actionOfFrame`），这样"点一行 → 右侧显示这一行是什么"只有一个入口。
  const [selFrame, setSelFrame] = useState<RecordingFrame | null>(null)
  const [cursor, setCursor] = useState(0)
  // 被选中那一帧对应的语义动作（没有就是 undefined：缓存帧、批量同步帧都不是用户动作）
  const selActionObj = useMemo(
    () => (selFrame ? actions.find((a) => frameKey(a.frame) === frameKey(selFrame)) || null : null),
    [actions, selFrame])
  // 帧 → 语义动作。event 帧才可能命中（动作只从事件帧里长出来）。
  const actionByKey = useMemo(() => {
    const m = new Map<string, RecordingAction>()
    for (const a of actions) m.set(frameKey(a.frame), a)
    return m
  }, [actions])
  const actionOfFrame = useCallback((f: RecordingFrame) => actionByKey.get(frameKey(f)), [actionByKey])

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
  // 说明：`hasLog` 仍然读进来（"这一份回过放没有"是服务端的事实），只是界面不再用它决定显示 ——
  // 显示与否只看"现在跑不跑"（见下面那条 `rp?.running`）。留着它是为了 toast 只在真有日志时说话。
  const rpSince = useRef(0)
  // 上一拍"还在跑吗"。用来抓"跑完了"那一次跳变 —— 进度条跑完就整行撤掉（用户 2026-10-16），
  // 不撤的话屏幕上会永久留一行 `完成 73/73` 的死信息；结果改由一条 toast 说一次。
  const wasRunning = useRef(false)
  const readProgress = useCallback(async (watch: boolean) => {
    try {
      const r = await api.replaySteps(id, 0, watch) // 每次整份读:几十到几千行,比"补页"少一类错
      rpSince.current = r.count || 0
      const pr = r.progress || emptyProgress()
      // 跑完的那一次：报一次结果（失败数不为 0 时用 error 的样式，让人一眼看出该去看日志）
      if (wasRunning.current && !r.running) {
        // 描述里只说**异常**（与进度条、表里的徽标同一套口径：正常回包不报）。
        // 全 0 就说明每一步都有回应 —— 那也要说出来，不然 toast 上什么都没有像是没跑成。
        const bad: string[] = []
        if (pr.noResp) bad.push(`无应答 ${pr.noResp}${pr.suspect ? `（可疑 ${pr.suspect}）` : ''}`)
        if (pr.skipped) bad.push(`跳过 ${pr.skipped}`)
        if (pr.failed) bad.push(`失败 ${pr.failed}`)
        const said = bad.length ? bad.join(' · ') : '每一步都有回应'
        if (pr.failed > 0) toast.error(`回放结束：${pr.done}/${pr.total}，失败 ${pr.failed}`, { description: said, duration: 0 })
        else toast.success(`回放完成：${pr.done}/${pr.total}`, { description: said })
      }
      wasRunning.current = !!r.running
      setRp({
        steps: r.states || [], progress: pr, running: !!r.running,
        note: r.note, hasLog: !!r.hasLog,
      })
    } catch { /* 读进度失败不动界面:它只是旁证,下一拍再试 */ }
  }, [id, toast])
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
  const onCtl = async (cmd: 'pause' | 'resume' | 'step' | 'skip' | 'stop') => {
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
      // 回放起了一轮会话 —— 列表上那行"正在录/正在回放"得跟着变（它在 view 不变时不会自己重拉）。
      onChanged?.()
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
  const doDelete = async () => {
    setConfirmDel(false)
    try {
      const r = await api.recordingDelete(id)
      toast.success(`已删除录制:${r.id}`, { description: `清除 ${r.files} 个文件,释放 ${fmtBytes(r.bytes)}` })
      // 删完必须让**列表**重拉一次，否则回到列表还看得见刚删掉的那一行 ——
      // 列表是 keep-alive 的（只切显隐不卸载），它的重拉条件是 `view/sessionEnv/tick`，
      // 而删除只把 recordId 清空（view 仍是 record）→ 那条 effect 不会重跑（实测踩到）。
      onChanged?.()
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
  // 只有一页了：进页面就拉第一页（不用等"切到全文流"）。
  useEffect(() => {
    if (streamPage.current === 0) loadMore()
  }, [loadMore])
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

  // Esc：**先关明细面板，再退列表**（面板是"点出来的东西"，用户预期是先收它）。
  // 正在选文本时不抢 ESC（那时它该是"取消选择"用）。用 ref 读最新的选中帧，
  // 免得把 selFrame 写进依赖、每点一行都重挂一次监听。
  const selFrameRef = useRef<RecordingFrame | null>(null)
  selFrameRef.current = selFrame
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return
      const sel = window.getSelection()?.toString()
      if (sel) return
      e.preventDefault()
      if (selFrameRef.current) { setSelFrame(null); return }
      onBack()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onBack])

  // **唯一**的轴域:条带与表格共用它。
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

  const pick = (f: RecordingFrame) => {
    setSelFrame(f)
    setCursor(axisValue(f, domain.mode))
  }

  // 条带上的缩放**只在键盘上**（用户 2026-10-16：把 `＋/－` 按钮拿掉，走快捷键）。
  // 锚点选游标（"我正在看这一刻，放大它"），没有游标就锚中间。
  // 输入框里打字不抢键（`+` 在文本框里是加号，不是缩放）。
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.ctrlKey || e.metaKey || e.altKey) return
      const el = e.target as HTMLElement | null
      const tag = el?.tagName?.toLowerCase()
      if (tag === 'input' || tag === 'textarea' || tag === 'select' || el?.isContentEditable) return
      const mid = (bounds.min + bounds.max) / 2
      const anchor = cursor > 0 ? cursor : mid
      if (e.key === '+' || e.key === '=') setDomain((d) => zoomDomain(d, bounds, 0.7, anchor))
      else if (e.key === '-' || e.key === '_') setDomain((d) => zoomDomain(d, bounds, 1.4, anchor))
      else if (e.key === '0') setDomain(bounds)
      else return
      e.preventDefault()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [bounds, cursor])

  const firstT = ov?.items.find((f) => f.t > 0)?.t || 0
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
        {/* 头里**只留身份**：ID 与作业名。存储位置（长期库/缓存）、三层计数（帧/事件/步）、
            时间精度都不在这儿了（用户 2026-10-16 定为"移除"）—— 它们是「这份录制是什么样」
            的答案，在明细面板的「概览」页签里本来就有一份（帧/事件/语义动作/时间精度四行），
            头里再写一遍就是同一件事有两个家。横轴不是时间时，条带自己会挂一条警告。 */}
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
          {/* 回放进度条**只在回放跑着的时候存在**（用户 2026-10-16 要求移除它）：
              跑完就整行撤掉 —— 否则屏幕上会永久留一行 `完成 73/73` 的死信息，而它答不了
              任何当下问题。这一行同时是**唯一的回放控制台**（暂停/单步/停止、走到第几步、
              正在等的那一步、已等 N ms），所以"跑着的时候"必须留着；
              跑完的结果由 `readProgress` 里那一次跳变发一条 toast 说，每步的徽标仍在表里对应行上。 */}
          {rp?.running && (
            <ReplayProgressBar
              id={id} progress={rp.progress} steps={rp.steps} running={rp.running} note={rp.note}
              onCtl={(c) => void onCtl(c)} />
          )}
          <RecordingOverview
            items={ov.items} blocks={ov.blocks} mode={bounds.mode}
            bounds={bounds} domain={domain} onDomain={setDomain}
            cursor={cursor} onCursor={setCursor}
            sel={selFrame} onSelect={pick} />

          <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden" style={{ flexDirection: narrow ? 'column' : 'row' }}>
            <div className="flex min-h-0 min-w-0 flex-1 flex-col">
              {/* 没有页签了（用户 2026-10-16 定的形态）：**一屏一条全文流**。
                  计数只在"还没加载完"时才说一句 —— 头上那行已经有帧/事件/步的全貌。 */}
              {streamFrames.length < streamTotal && streamTotal > 0 && (
                <div className="flex h-6 shrink-0 items-center border-b border-border px-2 text-[10px] text-muted-foreground">
                  已加载 {streamFrames.length} / {streamTotal} 帧，滚动到底继续加载
                </div>
              )}
              <RecordingTable
                key={id}
                frames={streamFrames.length ? streamFrames : ov.items}
                precision={ov.timePrecision}
                cursor={cursor} onCursor={setCursor}
                sel={selFrame} onSelect={pick}
                actionOf={actionOfFrame}
                stepByEvent={stepByEvent}
                total={streamTotal || ov.items.length}
                onLoadMore={loadMore} />
            </div>

            {/* 明细面板：**默认不显示**（用户定的），点一行才出现；左边缘一个关闭按钮。
                窄容器下它落到下面（整屏只剩这一条流时，下面的面板仍然可读）。 */}
            {selFrame && !narrow && <div className="col-resizer self-stretch" onMouseDown={onResizeDown} title="拖拽分配宽度" />}

            {selFrame && (
              <div className={narrow
                ? 'h-56 shrink-0 overflow-hidden border-t border-border'
                : 'flex shrink-0 flex-col overflow-hidden border-l border-border'}
                style={{ width: narrow ? undefined : inspW }}>
                <RecordingInspector
                  id={id} sel={selFrame} action={selActionObj}
                  precision={ov.timePrecision} counts={ov.counts} notes={ov.notes}
                  directions={ov.directions} firstT={firstT} meta={meta} metaLoaded={metaLoaded}
                  onClose={() => setSelFrame(null)} />
              </div>
            )}
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




