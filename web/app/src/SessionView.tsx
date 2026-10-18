// 会话页:一条会话对应的工作面板。
//
// 两个 Tab 恒在(Debug / FEP录制),**顺序与默认落点由会话类型决定**:Debug 会话先看 Debug,
// FEP 会话先看 FEP录制。不要把"另一条 Tab"藏掉 —— 会话类型只决定"先看哪一个",
// 两条链路都要能从这里走到。
//
// Debug Tab 是原来整块调试工作台(见 DebugWorkbench.tsx);FEP录制 这一屏现在如实显示
// **这条会话挂着的画面通道**(建会话时默认就挂 —— 一个脱离进程的 `tt recording start`,
// 录制代理在中间)。没挂上就说没挂上,不留一块什么都不说的白板。
import { useEffect, useRef, useState } from 'react'
import { FolderOpen, Video } from 'lucide-react'
import type { ReactNode } from 'react'
import { api } from './api'
import { useStore, type SessionItem } from './store'
import type { SessionTab } from './routing'
import { Pane, TabStrip, ViewHeader, Meta, type TabItem } from './Shell'
import { DebugWorkbench } from './DebugWorkbench'
import { StatusBar } from './StatusBar'
import { SESSION_TYPE_LABEL } from './Sidebar'

/** 画面通道(录制代理)的只读视图，与后端 `GET /api/sessions/{id}/screen` 的 `screen` 块同形。 */
interface ScreenInfo {
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
 * 两个来源，合起来用：
 *   - **服务端的 screen 事件**（`store.screens`，W2）：挂了/收了立刻反映，不必等轮询；
 *   - `GET /sessions/{id}/screen`：进页面时补一次（事件是"叫一声"，权威在查询）。
 * 取不到就显示成"没有"，并把真实的报错留在 `err` 里 —— 两者不能混成一句话。
 */
function useScreenInfo(id: string) {
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

export function SessionView({ session, tab, onTab, onOpenEnv, onOpenRecording, leading }: {
  session: SessionItem
  tab: SessionTab
  onTab: (t: SessionTab) => void
  onOpenEnv: () => void
  /** 打开这一轮录制件的明细（环境页 → FEP录像 → 那一份）。 */
  onOpenRecording: (recId: string) => void
  leading?: ReactNode
}) {
  const screen = useScreenInfo(session.id)

  // 会话页一进来（以及状态每次变化）就问一句"要不要把母版源码打开"。
  //
  // 为什么放在这一层而不是绑定时一次性做完：会话刚建好是 idle、挂上 proxy 后变成
  // running、命中第一个断点又变 stopped —— 模块名要等**起跑那一刻**服务端才解析出来，
  // 绑定时拉不到。所以由状态变化驱动重试，`ensureSource` 自己带幂等判断（有源码就早退）。
  const state = useStore((s) => s.state)
  useEffect(() => {
    void useStore.getState().ensureSource()
  }, [session.id, state])

  const debugTab: TabItem<SessionTab> = { key: 'debug', label: 'Debug', hint: '断点、单步、变量与调用栈' }
  const fepTab: TabItem<SessionTab> = { key: 'fep', label: 'FEP录制', hint: '这条会话的画面通道与录制' }
  // 用途决定给什么（W7 的 PANE_POLICY）：
  //   - record：**不给编辑器**（已经放行跑着，防误操作），只剩 FEP 监控与收尾；
  //   - 其余：Debug + FEP 两条，顺序按用途（debug 先看 Debug，replay 先看 FEP）。
  const recordOnly = session.purpose === 'record'
  const tabs = recordOnly
    ? [fepTab]
    : (session.purpose === 'replay' ? [fepTab, debugTab] : [debugTab, fepTab])
  // 深链接可能写着 `/debug`，而 record 会话没有那一屏 —— 实际展示的那一屏以**用途**为准。
  const shownTab: SessionTab = recordOnly ? 'fep' : tab

  const meta = (
    <>
      <Meta title="会话所属环境">{session.env}</Meta>
      {session.ent && <><span className="opacity-30">·</span>
        <Meta title="企业编号（建会话时点名的那个）">ENT {session.ent}</Meta></>}
      {session.prog && <><span className="opacity-30">·</span>
        <Meta title="作业编号">作业 {session.prog}</Meta></>}
      {session.ref && <><span className="opacity-30">·</span>
        <Meta title="服务端规范 id（环境#序号）">{session.ref}</Meta></>}
      <span className="opacity-30">·</span>
      <Meta title={screen.attached ? '这条会话正挂着一块画面通道（录制代理）' : '这条会话当前没有画面通道'}>
        {screen.attached ? `录制 ${screen.screen?.tag || '已挂'}` : '无画面'}
      </Meta>
    </>
  )

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col bg-background">
      <ViewHeader title={session.name} meta={meta} leading={leading}
        actions={
          <button onClick={onOpenEnv} title={`打开环境「${session.env}」`}
            className="flex h-7 items-center gap-1 rounded-[7px] border border-border px-2 text-[12px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
            <FolderOpen className="h-3.5 w-3.5" />
            {session.env}
          </button>
        } />
      <TabStrip items={tabs} active={shownTab} onChange={onTab} />

      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <Pane show={shownTab === 'debug'} label="Debug">
          <DebugWorkbench active={shownTab === 'debug'} />
        </Pane>
        <Pane show={shownTab === 'fep'} label="FEP录制">
          <FepPane session={session} screen={screen} onOpenRecording={onOpenRecording} />
        </Pane>
      </div>
      {/* 状态栏（停站橙/运行蓝/空闲灰 + 接管按钮）属于**这条会话**，所以钉在会话页底部，
          不再横跨整个壳子 —— 左侧树和环境页没有"调试状态"可报。 */}
      <StatusBar />
    </div>
  )
}

/**
 * FEP录制:这条会话的画面通道。
 *
 * 写清楚四件事:这条会话是哪条、画面挂着没有、挂着的是哪一份录制、下一步能做什么。
 * 这一屏的功能(起停录制 / 回放 / 对拍)后续逐个接进来;现在能看的"挂着没有"本身
 * 就有用 —— 建会话默认挂代理，挂失败了必须能在这里看见，而不是只能翻日志。
 */
function FepPane({ session, screen, onOpenRecording }: {
  session: SessionItem
  screen: ReturnType<typeof useScreenInfo>
  onOpenRecording: (recId: string) => void
}) {
  const s = screen.screen
  const startRecording = useStore((st) => st.startRecording)
  const stopRecording = useStore((st) => st.stopRecording)
  const [prog, setProg] = useState(session.prog || '')
  const [ent, setEnt] = useState(session.ent || '')
  const [busy, setBusy] = useState<'' | 'start' | 'stop'>('')
  const [msg, setMsg] = useState('')
  const attached = screen.attached
  const tag = attached ? (s?.tag || '') : ''
  const isRec = (s?.kind || 'rec') === 'rec'
  const stale = !!s?.stale

  const onStart = async () => {
    setBusy('start'); setMsg('')
    try {
      await startRecording(prog, ent)
      screen.refresh()
    } catch (e: any) {
      setMsg(String(e?.message || e))
    } finally { setBusy('') }
  }
  const onStop = async () => {
    setBusy('stop'); setMsg('')
    try {
      await stopRecording()
      screen.refresh()
    } catch (e: any) {
      setMsg(String(e?.message || e))
    } finally { setBusy('') }
  }

  return (
    <div className="min-h-0 flex-1 overflow-auto p-6">
      <div className="mx-auto w-full max-w-[860px]">
        <div className="flex items-center gap-2 text-[15px] font-semibold text-foreground">
          <Video className="h-4 w-4" />
          FEP 录制
        </div>
        <p className="mt-1 text-xs leading-relaxed text-muted-foreground">
          画面通道 = 一块 GDC 画面 + 一个录制代理。这条会话{screen.attached ? '现在正挂着一块画面通道' : '现在没有画面通道'}。
          {attached && isRec && '下面的帧流是**只读**的实时监控（像 F12 的 Network）：它不发送任何命令，也不会改这一轮录制。'}
        </p>

        {screen.err && (
          <div className="mt-3 rounded-[8px] border border-red-500/30 bg-red-500/10 px-3 py-2 text-[11px] leading-relaxed text-red-600 dark:text-red-400">
            查画面通道失败：{screen.err}
          </div>
        )}

        <dl className="mt-4 space-y-2 rounded-[12px] border border-border bg-card p-4 text-xs">
          <Row k="环境" v={session.env} />
          <Row k="会话" v={session.ref || session.id} mono />
          <Row k="ENT" v={session.ent || '未记录'} />
          <Row k="作业" v={session.prog || '未记录'} />
          <Row k="画面" v={attached ? (stale ? '挂着，但进程已不在（stale）' : '已挂上') : '没有画面通道'} />
          {attached && s?.tag && <Row k="录制 ID" v={s.tag} mono />}
          {attached && s?.kind && <Row k="种类" v={isRec ? '录制' : '回放'} />}
          {attached && (s?.proxyPort || s?.injectPort) && (
            <Row k="端口" v={`代理 ${s?.proxyPort || '?'} → 注入 ${s?.injectPort || '?'}`} mono />
          )}
          {attached && !!s?.gdcPid && (
            <Row k="GDC" v={`进程 ${s.gdcPid}${s.gdcPort ? `（端口 ${s.gdcPort}）` : ''}`} mono />
          )}
          {attached && s?.started && <Row k="开始" v={String(s.started)} mono />}
          {attached && s?.recordDir && <Row k="录制目录" v={s.recordDir} mono />}
        </dl>

        <div className="mt-3 rounded-[12px] border border-border bg-card p-4">
          {!attached ? (
            <>
              <div className="text-xs font-medium text-foreground">开始录制</div>
              <p className="mt-1 text-[11px] leading-relaxed text-muted-foreground">
                会话要**空闲**、**有名字**、点了作业才挂得上（后端会拒，并把原因说清楚）。
              </p>
              <div className="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-2">
                <label className="text-[11px] text-muted-foreground">
                  作业编号
                  <input value={prog} onChange={(e) => setProg(e.target.value)} placeholder="如 axmt500_wf"
                    className="mt-1 w-full rounded-[6px] border border-border bg-background px-2 py-1 font-mono text-[12px] text-foreground" />
                </label>
                <label className="text-[11px] text-muted-foreground">
                  企业编号 ENT
                  <input value={ent} onChange={(e) => setEnt(e.target.value)} placeholder="如 99"
                    className="mt-1 w-full rounded-[6px] border border-border bg-background px-2 py-1 font-mono text-[12px] text-foreground" />
                </label>
              </div>
              <div className="mt-2 flex items-center gap-2">
                <button onClick={onStart} disabled={busy !== ''}
                  className="rounded-[7px] bg-primary px-2.5 py-1 text-[12px] text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50">
                  {busy === 'start' ? '正在起画面…' : '开始录制'}
                </button>
                <button onClick={screen.refresh}
                  className="rounded-[7px] border border-border px-2 py-1 text-[12px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
                  重新查一次
                </button>
              </div>
            </>
          ) : (
            <>
              <div className="flex flex-wrap items-center gap-2">
                <button onClick={onStop} disabled={busy !== ''}
                  className="rounded-[7px] border border-border px-2.5 py-1 text-[12px] text-foreground transition-colors hover:bg-accent disabled:opacity-50">
                  {busy === 'stop' ? '正在收…' : isRec ? '停止录制' : '停止回放'}
                </button>
                {s?.tag && (
                  <button onClick={() => onOpenRecording(s.tag!)}
                    className="rounded-[7px] border border-border px-2 py-1 text-[12px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
                    打开录制明细
                  </button>
                )}
                <button onClick={screen.refresh}
                  className="rounded-[7px] border border-border px-2 py-1 text-[12px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
                  重新查一次
                </button>
              </div>
              <p className="mt-2 text-[11px] leading-relaxed text-muted-foreground">
                停止会把这轮收干净（GDC / 代理 / 录制进程），录制件落进长期库或缓存；这是**唯一**该做的收尾动作。
              </p>
            </>
          )}
          {msg && (
            <div className="mt-2 rounded-[6px] border border-red-500/30 bg-red-500/10 px-2 py-1 text-[11px] leading-relaxed text-red-600 dark:text-red-400">
              {msg}
            </div>
          )}
        </div>

        {attached && isRec && tag && <LiveFrames tag={tag} stale={stale} />}
      </div>
    </div>
  )
}

/** 实时帧流的一行（`/api/recordings/{id}/timeline` 的 frames 元素）。 */
interface LiveRow {
  o: number; n: number; t: number; c: number; d: number; k: string; q: number; r: number
  preview?: string
}

/**
 * 一块**只读**的实时帧流（F12/Network 那种）：按 `from=<已看条数>` 增量取，不重排、不补历史。
 *
 * 为什么是轮询而不是长连接：录制文件是**追加写**的，`timeline` 每次重新读一遍就能看到新帧
 * （W5 的增量解析是性能优化，不是语义要求）。1 秒一取足以跟上人眼；等有了 `watch=1`
 * 的长连接再换，接口形状不用变。
 */
function LiveFrames({ tag, stale }: { tag: string; stale: boolean }) {
  const [rows, setRows] = useState<LiveRow[]>([])
  const [total, setTotal] = useState(0)
  const [err, setErr] = useState('')
  const [follow, setFollow] = useState(true)
  // 默认只看**消息/事件**（om/un/rn/event）：缓存帧（图片）占八成，混进来就是刷屏，
  // 反而看不出"这一步发了什么"。需要时再打开（就是 Network 的 "All" 那一档）。
  const [showAll, setShowAll] = useState(false)
  const kindsRef = useRef('om,un,rn,event')
  useEffect(() => { kindsRef.current = showAll ? '' : 'om,un,rn,event' }, [showAll])
  // seen / 跟随开关放 ref：轮询闭包里要读最新值，而 effect 不该因为它们重启动
  // （重启动会把 seen 归零，帧流里就会出现重复行）。
  const seenRef = useRef(0)
  const followRef = useRef(true)
  const tagRef = useRef('')
  useEffect(() => { followRef.current = follow }, [follow])

  useEffect(() => {
    let cancelled = false
    if (tagRef.current !== tag) {
      tagRef.current = tag
      seenRef.current = 0
      setRows([]); setTotal(0); setErr('')
    }
    const tick = async () => {
      try {
        const from = followRef.current ? seenRef.current : 0
        const r = await api.recordingTimeline(tag, {
          from,
          limit: followRef.current ? 200 : 1,
          kind: kindsRef.current || undefined,
        })
        if (cancelled) return
        setTotal(r.total || 0)
        setErr('')
        if (!followRef.current) return
        const got = (r.frames || []) as LiveRow[]
        if (!got.length) return
        seenRef.current += got.length
        setRows((prev) => prev.concat(got).slice(-800))
      } catch (e: any) {
        if (!cancelled) setErr(String(e?.message || e))
      }
    }
    void tick()
    const h = window.setInterval(tick, 1000)
    return () => { cancelled = true; clearInterval(h) }
  }, [tag, showAll])

  return (
    <div className="mt-3 flex min-h-[240px] flex-col rounded-[12px] border border-border bg-card">
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <span className="text-xs font-medium text-foreground">实时帧流</span>
        <span className="text-[11px] text-muted-foreground">共 {total} 帧 · 已看 {seenRef.current} 帧</span>
        {stale && <span className="text-[11px] text-amber-600 dark:text-amber-400">录制进程已不在</span>}
        <div className="ml-auto flex items-center gap-2">
          <button onClick={() => setShowAll((v) => !v)}
            className="rounded-[6px] border border-border px-2 py-0.5 text-[11px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
            {showAll ? '只看消息' : '显示缓存帧'}
          </button>
          <button onClick={() => setFollow((f) => !f)}
            className="rounded-[6px] border border-border px-2 py-0.5 text-[11px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
            {follow ? '暂停跟随' : '继续跟随'}
          </button>
          <button onClick={() => { seenRef.current = 0; setRows([]) }}
            className="rounded-[6px] border border-border px-2 py-0.5 text-[11px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
            从头看
          </button>
        </div>
      </div>
      {err ? (
        <div className="p-3 text-[11px] leading-relaxed text-muted-foreground">
          还读不到这份录制：{err}
          <div className="mt-1 opacity-80">录制件要等这一段落盘才读得到；刚挂上时等一两秒再看。</div>
        </div>
      ) : rows.length === 0 ? (
        <div className="p-3 text-[11px] text-muted-foreground">还没有帧（程序还没动静，或者这一轮还没落盘）。</div>
      ) : (
        <ul className="min-h-0 flex-1 overflow-auto font-mono text-[11px] leading-[1.5]">
          {rows.map((r, i) => (
            <li key={`${r.o}-${r.d}-${i}`} className="flex gap-2 border-b border-border/40 px-3 py-1">
              <span className="w-10 shrink-0 text-muted-foreground">#{r.q || i + 1}</span>
              <span className={r.d === 1 ? 'w-4 shrink-0 text-sky-600 dark:text-sky-400' : 'w-4 shrink-0 text-emerald-600 dark:text-emerald-400'}>
                {r.d === 1 ? '→' : '←'}
              </span>
              <span className="w-12 shrink-0 text-muted-foreground">{r.k}</span>
              <span className="w-12 shrink-0 text-right text-muted-foreground">{r.n}B</span>
              <span className="w-24 shrink-0 truncate text-muted-foreground">c{r.c} o{r.o}</span>
              <span className="w-16 shrink-0 text-muted-foreground">
                {r.t > 0 ? new Date(r.t).toLocaleTimeString('zh-CN', { hour12: false }) : ''}
              </span>
              <span className="min-w-0 flex-1 truncate text-foreground/90">{r.preview || ''}</span>
            </li>
          ))}
        </ul>
      )}
      <div className="border-t border-border px-3 py-1.5 text-[10px] text-muted-foreground">
        只读监控：这里不会向录制/会话发任何命令。
      </div>
    </div>
  )
}

function Row({ k, v, mono }: { k: string; v: string; mono?: boolean }) {
  return (
    <div className="flex items-start gap-3">
      <dt className="w-16 shrink-0 text-muted-foreground">{k}</dt>
      <dd className={mono ? 'min-w-0 break-all font-mono text-foreground' : 'min-w-0 break-all text-foreground'}>{v}</dd>
    </div>
  )
}
