// 会话页:一条会话对应的工作面板。
//
// 两个 Tab(Debug / FEP监控),**顺序与默认落点由用途决定**:
//   - debug  → Debug + FEP监控(先看 Debug);
//   - record → **异常调试 + FEP监控**(先看异常调试,作业自动放行,只给日志与「结束」);
//   - replay → FEP监控 + Debug(先看监控,编辑器只读由模式决定)。
//
// Debug Tab 是原来整块调试工作台(见 DebugWorkbench.tsx);**FEP监控 那一屏只放协议流**
// (见 FepMonitor.tsx):三泳道时间轴 + 事件列表,只读、没有按钮。起停录制/录制 ID/端口/GDC
// 那些属于「环境页 → FEP录像」那套录制工作台 —— 会话页这边只回答"这条通道上过了什么"。
import { useEffect, useRef, useState } from 'react'
import { FolderOpen } from 'lucide-react'
import type { ReactNode } from 'react'
import { useStore, type SessionItem } from './store'
import type { SessionTab } from './routing'
import { Pane, TabStrip, ViewHeader, Meta, type TabItem } from './Shell'
import { DebugWorkbench } from './DebugWorkbench'
import { StatusBar } from './StatusBar'
import { SESSION_STATE, SESSION_TYPE_LABEL } from './Sidebar'
import { FepMonitor, useScreenInfo } from './FepMonitor'
import { ReplayDebugger } from './ReplayDebugger'


export function SessionView({ session, tab, onTab, onOpenEnv, leading }: {
  session: SessionItem
  tab: SessionTab
  onTab: (t: SessionTab) => void
  onOpenEnv: () => void
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

  const isRecord = session.purpose === 'record'
  // 同一屏两个用途两个意思：debug 是编辑器，record 是「异常调试」（只有日志 + 结束）。
  const debugTab: TabItem<SessionTab> = isRecord
    ? { key: 'debug', label: '异常调试', hint: '作业已自动放行；这里只实时显示日志，并提供结束' }
    : { key: 'debug', label: 'Debug', hint: '断点、单步、变量与调用栈' }
  const fepTab: TabItem<SessionTab> = { key: 'fep', label: 'FEP监控', hint: '这条会话的画面通道上过了什么（只读）' }
  // 用途决定先看哪一屏（W7 的 PANE_POLICY）：debug/record 先看第一屏，replay 先看监控。
  const tabs = session.purpose === 'replay' ? [fepTab, debugTab] : [debugTab, fepTab]
  const shownTab: SessionTab = tab

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
        <Pane show={shownTab === 'debug'} label={isRecord ? '异常调试' : 'Debug'}>
          {isRecord
            ? <AbnormalPane session={session} />
            : <DebugWorkbench active={shownTab === 'debug'} />}
        </Pane>
        <Pane show={shownTab === 'fep'} label="FEP监控">
          {session.purpose === 'replay'
            ? <ReplayDebugger sessionId={session.id} name={session.name} />
            : <FepMonitor sessionId={session.id} />}
        </Pane>
      </div>
      {/* 状态栏（停站橙/运行蓝/空闲灰 + 接管按钮）属于**这条会话**，所以钉在会话页底部，
          不再横跨整个壳子 —— 左侧树和环境页没有"调试状态"可报。 */}
      <StatusBar />
    </div>
  )
}

/**
 * 「异常调试」：record 用途会话的那一屏。
 *
 * 纯录制会话没有正常调试可做（不设断点、不单步）：作业在建会话时就自动放行、已经跑起来了。
 * 这一屏只做两件事 —— **实时日志** + **结束**：
 *   - 日志来自 `output` 事件（`store.rawLog`），与程序同源：出异常时你看的就是这几行；
 *   - 结束 = **结束整条会话**：服务端级联收录制进程 / GDC / 代理，录制件落长期库。
 *     刻意不做"只停录制" —— 这条会话存在的意义就是这一轮录制，留一条既不能调也不能录的
 *     空壳没有价值（用户 2026-10-18 定的）。
 */
function AbnormalPane({ session }: { session: SessionItem }) {
  const rawLog = useStore((s) => s.rawLog)
  const state = useStore((s) => s.state)
  const closeSession = useStore((s) => s.closeSession)
  const [confirming, setConfirming] = useState(false)
  const [busy, setBusy] = useState(false)
  const boxRef = useRef<HTMLDivElement | null>(null)

  // 跟着新行滚到底：异常是"刚刚发生"的事，停在上面等于没看见。
  useEffect(() => {
    const el = boxRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [rawLog.length])

  const end = async () => {
    if (busy) return
    setBusy(true)
    try {
      await closeSession(session.id)
    } finally {
      setBusy(false)
      setConfirming(false)
    }
  }
  const st = SESSION_STATE[state] || { dot: 'bg-muted-foreground/25', label: '状态未知' }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      <div className="flex flex-wrap items-center gap-2 border-b border-border px-3 py-2">
        <span className={`h-2 w-2 shrink-0 rounded-full ${st.dot}`} />
        <span className="text-xs text-muted-foreground">{st.label}</span>
        <span className="text-[11px] text-muted-foreground/80">作业已自动放行 · 只实时显示日志</span>
        <div className="ml-auto flex items-center gap-2">
          {confirming ? (
            <>
              <span className="text-[11px] text-muted-foreground">结束会同时停止录制</span>
              <button onClick={() => void end()} disabled={busy}
                className="rounded-[7px] border border-red-500/40 bg-red-500/10 px-2.5 py-1 text-[12px] text-red-600 transition-colors hover:bg-red-500/20 disabled:opacity-50 dark:text-red-400">
                {busy ? '结束中…' : '确认结束'}
              </button>
              <button onClick={() => setConfirming(false)} disabled={busy}
                className="rounded-[7px] border border-border px-2 py-1 text-[12px] text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:opacity-50">
                取消
              </button>
            </>
          ) : (
            <button onClick={() => setConfirming(true)}
              className="rounded-[7px] border border-border px-2.5 py-1 text-[12px] text-foreground transition-colors hover:bg-accent">
              结束
            </button>
          )}
        </div>
      </div>
      <div ref={boxRef} className="min-h-0 flex-1 overflow-auto bg-card font-mono text-[11px] leading-[1.55]">
        {rawLog.length === 0 ? (
          <div className="p-3 text-muted-foreground">
            作业已放行，正在等程序输出……
            <div className="mt-1 text-[10px] opacity-80">
              这一屏只显示**打开页面之后**的新输出（历史行不在缓冲里）；要回看报文请切「FEP监控」。
            </div>
          </div>
        ) : (
          rawLog.map((line, i) => (
            <div key={i} className="whitespace-pre-wrap break-all px-3 py-[1px] text-foreground/90">{line || ' '}</div>
          ))
        )}
      </div>
    </div>
  )
}